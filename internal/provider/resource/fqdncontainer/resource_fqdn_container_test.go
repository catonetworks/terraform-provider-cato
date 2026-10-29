package fqdncontainer

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

func TestInlineValidation(t *testing.T) {
	p := inlineModel()
	var d diag.Diagnostics
	buildInline(context.Background(), p, &d)
	require.False(t, d.HasError(), "%v", d)
	for _, v := range []types.Set{types.SetNull(types.StringType), types.SetUnknown(types.StringType), types.SetValueMust(types.StringType, []attr.Value{types.StringUnknown()}), types.SetValueMust(types.StringType, []attr.Value{types.StringValue("EXAMPLE.COM")})} {
		p.Values = v
		d = nil
		buildInline(context.Background(), p, &d)
		require.True(t, d.HasError())
	}
}
func TestDownloadValidation(t *testing.T) {
	var d diag.Diagnostics
	v := decodeMembers(base64.StdEncoding.EncodeToString([]byte("example.com\napi.example.com")), &d)
	require.False(t, d.HasError())
	require.Len(t, v.Elements(), 2)
	for _, s := range []string{"invalid!", base64.StdEncoding.EncodeToString([]byte{255}), base64.StdEncoding.EncodeToString([]byte("EXAMPLE.COM"))} {
		d = nil
		decodeMembers(s, &d)
		require.True(t, d.HasError())
	}
	d = nil
	require.Empty(t, decodeMembers("", &d).Elements())
	require.False(t, d.HasError())
}
func inlineModel() resourceModel {
	return resourceModel{ID: types.StringUnknown(), Name: types.StringValue("test"), Description: types.StringValue("test"), Mode: types.StringValue("INLINE"), Size: types.Int64Unknown(), FileType: types.StringNull(), SyncData: types.ObjectNull(syncAttrTypes), Values: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("example.com")})}
}

type backend struct {
	object          map[string]any
	values          []string
	fail, malformed string
	calls           []string
	inputs          []map[string]any
	wrong           bool
}

func remote() map[string]any {
	return map[string]any{"id": "container-1", "name": "test", "description": "test", "size": 1, "syncData": nil}
}
func setup(t *testing.T, b *backend) *containerResource {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() { require.NoError(t, req.Body.Close()) }()
		var body struct {
			OperationName string         `json:"operationName"`
			Query         string         `json:"query"`
			Variables     map[string]any `json:"variables"`
		}
		require.NoError(t, json.NewDecoder(req.Body).Decode(&body))
		op, v := body.OperationName, body.Variables
		verifyQuery(t, op, body.Query)
		b.calls = append(b.calls, op)
		require.Equal(t, "account-123", v["accountId"])
		require.NotContains(t, body.Query, "audit")
		w.Header().Set("Content-Type", "application/json")
		write := func(value any) { require.NoError(t, json.NewEncoder(w).Encode(value)) }
		if b.fail == op {
			write(map[string]any{"errors": []any{map[string]any{"message": "denied"}}})
			return
		}
		payload := map[string]any{}
		family := map[string]any{}
		payload["fqdn"] = family
		switch op {
		case "containerList":
			delete(payload, "fqdn")
			require.Equal(t, map[string]any{"refs": []any{map[string]any{"by": "ID", "input": "container-1"}}, "types": []any{"FQDN"}}, v["containerSearchInput"])
			items := []any{}
			if b.object != nil {
				items = append(items, b.object)
			}
			payload["list"] = map[string]any{"containers": items}
		case "containerDelete":
			delete(payload, "fqdn")
			require.Equal(t, map[string]any{"ref": map[string]any{"by": "ID", "input": "container-1"}}, v["deleteContainerInput"])
			payload["delete"] = map[string]any{"container": b.object}
			b.object = nil
		case "containerFqdnDownloadFile":
			require.Equal(t, map[string]any{"by": "ID", "input": "container-1"}, v["downloadFqdnContainerFileInput"])
			encoded := base64.StdEncoding.EncodeToString([]byte(strings.Join(b.values, "\n")))
			if b.malformed != "" {
				encoded = b.malformed
			}
			id := "container-1"
			if b.wrong {
				id = "foreign"
			}
			family["downloadFile"] = map[string]any{"id": id, "name": "test", "encodedFile": encoded}
		default:
			create := strings.Contains(op, "Create")
			fromURL := strings.HasSuffix(op, "URL")
			action := "updateFromList"
			inputName := "updateFqdnContainerFromListInput"
			if create {
				action = "createFromList"
				inputName = "createFqdnContainerFromListInput"
			}
			if fromURL {
				action = strings.TrimSuffix(action, "List") + "URL"
				inputName = strings.TrimSuffix(inputName, "ListInput") + "UrlInput"
			}
			in, ok := v[inputName].(map[string]any)
			require.True(t, ok, "%v", v)
			b.inputs = append(b.inputs, in)
			if create {
				b.object = remote()
				b.object["name"] = in["name"]
			} else {
				require.Equal(t, map[string]any{"by": "ID", "input": "container-1"}, in["ref"])
			}
			b.object["description"] = in["description"]
			if fromURL {
				s := in["syncData"].(map[string]any)
				s["fileType"] = in["fileType"]
				ns := s["notifications"].(map[string]any)
				for key, items := range ns {
					out := []any{}
					for _, r := range items.([]any) {
						ref := r.(map[string]any)
						out = append(out, map[string]any{"id": ref["input"], "name": "ref"})
					}
					ns[key] = out
				}
				b.object["syncData"] = s
			} else {
				raw, exists := in["values"]
				require.True(t, exists, "empty values must be explicit")
				b.values = []string{}
				for _, value := range raw.([]any) {
					b.values = append(b.values, value.(string))
				}
				b.object["size"] = len(b.values)
			}
			family[action] = map[string]any{"container": b.object}
		}
		write(map[string]any{"data": map[string]any{"container": payload}})
	}))
	t.Cleanup(server.Close)
	sdk, err := cato.New(server.URL, "test-token", "wrong-default", nil, nil)
	require.NoError(t, err)
	return &containerResource{client: &catoclient.Service{AccountId: "account-123", Catov2: catoclient.NewProviderSDKClient(sdk)}}
}
func schemaFor(t *testing.T) schema.Schema {
	t.Helper()
	var r resource.SchemaResponse
	NewResource().Schema(context.Background(), resource.SchemaRequest{}, &r)
	return r.Schema
}
func planFor(t *testing.T, value resourceModel) tfsdk.Plan {
	t.Helper()
	p := tfsdk.Plan{Schema: schemaFor(t)}
	require.False(t, p.Set(context.Background(), value).HasError())
	return p
}
func stateFor(t *testing.T, value resourceModel) tfsdk.State {
	t.Helper()
	s := tfsdk.State{Schema: schemaFor(t)}
	require.False(t, s.Set(context.Background(), value).HasError())
	return s
}
func TestCRUD(t *testing.T) {
	ctx := context.Background()
	for _, urlMode := range []bool{false, true} {
		t.Run(fmt.Sprint(urlMode), func(t *testing.T) {
			b := &backend{}
			r := setup(t, b)
			p := inlineModel()
			if urlMode {
				p = urlModel()
			}
			var created resource.CreateResponse
			created.State = tfsdk.State{Schema: schemaFor(t)}
			r.Create(ctx, resource.CreateRequest{Plan: planFor(t, p)}, &created)
			require.False(t, created.Diagnostics.HasError(), "%v", created.Diagnostics)
			var current resourceModel
			require.False(t, created.State.Get(ctx, &current).HasError())
			require.Equal(t, "container-1", current.ID.ValueString())
			var read resource.ReadResponse
			read.State = created.State
			r.Read(ctx, resource.ReadRequest{State: created.State}, &read)
			require.False(t, read.Diagnostics.HasError(), "%v", read.Diagnostics)
			require.True(t, created.State.Raw.Equal(read.State.Raw))
			p = current
			p.Description = types.StringValue("updated")
			if !urlMode {
				p.Values = types.SetValueMust(types.StringType, []attr.Value{})
			}
			var updated resource.UpdateResponse
			updated.State = created.State
			r.Update(ctx, resource.UpdateRequest{Plan: planFor(t, p), State: created.State}, &updated)
			require.False(t, updated.Diagnostics.HasError(), "%v", updated.Diagnostics)
			require.False(t, updated.State.Get(ctx, &current).HasError())
			require.Equal(t, "updated", current.Description.ValueString())
			if !urlMode {
				require.Empty(t, current.Values.Elements())
			} else {
				require.NotContains(t, b.calls, "containerFqdnDownloadFile")
				require.True(t, current.Values.IsNull())
			}
			var imported resource.ImportStateResponse
			imported.State = stateFor(t, resourceModel{SyncData: types.ObjectNull(syncAttrTypes), Values: types.SetNull(types.StringType)})
			r.ImportState(ctx, resource.ImportStateRequest{ID: "container-1"}, &imported)
			read.State = imported.State
			r.Read(ctx, resource.ReadRequest{State: imported.State}, &read)
			require.False(t, read.Diagnostics.HasError(), "%v", read.Diagnostics)
			require.True(t, read.State.Raw.Equal(updated.State.Raw))
			var deleted resource.DeleteResponse
			deleted.State = updated.State
			r.Delete(ctx, resource.DeleteRequest{State: updated.State}, &deleted)
			require.False(t, deleted.Diagnostics.HasError(), "%v", deleted.Diagnostics)
			require.True(t, deleted.State.Raw.IsNull())
			deleted.State = updated.State
			r.Delete(ctx, resource.DeleteRequest{State: updated.State}, &deleted)
			require.False(t, deleted.Diagnostics.HasError())
			read.State = updated.State
			r.Read(ctx, resource.ReadRequest{State: updated.State}, &read)
			require.False(t, read.Diagnostics.HasError(), "%v", read.Diagnostics)
			require.True(t, read.State.Raw.IsNull())
		})
	}
}
func urlModel() resourceModel {
	p := inlineModel()
	p.Mode = types.StringValue("URL")
	p.FileType = types.StringValue("CSV")
	p.Values = types.SetNull(types.StringType)
	empty := types.SetValueMust(refType, []attr.Value{})
	ref := types.ObjectValueMust(refType.AttrTypes, map[string]attr.Value{"id": types.StringValue("ref-1"), "name": types.StringUnknown()})
	p.SyncData = types.ObjectValueMust(syncAttrTypes, map[string]attr.Value{"time_interval": types.Int64Value(1), "time_unit": types.StringValue("DAY"), "url": types.StringValue("https://example.com/container.csv"), "notifications": types.ObjectValueMust(notificationsAttrTypes, map[string]attr.Value{"mailing_list": empty, "subscription_group": types.SetValueMust(refType, []attr.Value{ref}), "webhook": empty})})
	return p
}
func TestURLValidation(t *testing.T) {
	for _, u := range []string{"http://example.com", "https://u:p@example.com", "https://example.com/#x", "https:///x"} {
		require.False(t, validURL(u))
	}
	require.True(t, validURL("https://example.com/file.csv"))
	for _, n := range []int64{0, -1, 2147483648} {
		require.False(t, validInterval(n))
	}
	p := urlModel()
	var d diag.Diagnostics
	validateModel(context.Background(), p, &d)
	require.False(t, d.HasError(), "%v", d)
	p.Values = inlineModel().Values
	validateModel(context.Background(), p, &d)
	require.True(t, d.HasError())
}
func TestFailureRecovery(t *testing.T) {
	ctx := context.Background()
	b := &backend{malformed: "invalid!"}
	r := setup(t, b)
	var created resource.CreateResponse
	created.State = tfsdk.State{Schema: schemaFor(t)}
	r.Create(ctx, resource.CreateRequest{Plan: planFor(t, inlineModel())}, &created)
	require.True(t, created.Diagnostics.HasError())
	var p resourceModel
	require.False(t, created.State.Get(ctx, &p).HasError())
	require.Equal(t, "container-1", p.ID.ValueString())
	require.False(t, p.Size.IsUnknown())
	b.malformed = ""
	state := created.State
	for _, op := range []string{"containerList", "containerFqdnDownloadFile"} {
		b.fail = op
		read := resource.ReadResponse{State: state}
		r.Read(ctx, resource.ReadRequest{State: state}, &read)
		require.True(t, read.Diagnostics.HasError())
		require.True(t, state.Raw.Equal(read.State.Raw))
	}
	b.fail = "containerFqdnUpdateFromList"
	updated := resource.UpdateResponse{State: state}
	r.Update(ctx, resource.UpdateRequest{State: state, Plan: planFor(t, p)}, &updated)
	require.True(t, updated.Diagnostics.HasError())
	require.True(t, state.Raw.Equal(updated.State.Raw))
	b.fail = "containerDelete"
	deleted := resource.DeleteResponse{State: state}
	r.Delete(ctx, resource.DeleteRequest{State: state}, &deleted)
	require.True(t, deleted.Diagnostics.HasError())
	require.True(t, state.Raw.Equal(deleted.State.Raw))
}
func TestUpdateReplacementGuard(t *testing.T) {
	ctx := context.Background()
	b := &backend{object: remote()}
	r := setup(t, b)
	p := inlineModel()
	p.ID = types.StringValue("container-1")
	state := stateFor(t, p)
	p.Name = types.StringValue("renamed")
	updated := resource.UpdateResponse{State: state}
	r.Update(ctx, resource.UpdateRequest{State: state, Plan: planFor(t, p)}, &updated)
	require.True(t, updated.Diagnostics.HasError())
	require.Equal(t, []string{"containerList"}, b.calls)
}

func TestConfigurationAndMockInjection(t *testing.T) {
	ctx := context.Background()
	r := NewResource().(*containerResource)
	var configured resource.ConfigureResponse
	r.Configure(ctx, resource.ConfigureRequest{}, &configured)
	require.False(t, configured.Diagnostics.HasError())
	r.Configure(ctx, resource.ConfigureRequest{ProviderData: "wrong"}, &configured)
	require.True(t, configured.Diagnostics.HasError())
	var metadata resource.MetadataResponse
	r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "cato"}, &metadata)
	require.Equal(t, "cato_fqdn_container", metadata.TypeName)
	var d diag.Diagnostics
	require.False(t, r.configured(&d))
	client := mocks.NewFQDNContainerClient(t)
	client.EXPECT().Find(mock.Anything, "account-123", "container-1").Return(nil, nil).Once()
	r.client = &catoclient.Service{AccountId: "account-123"}
	r.apiClient = client
	p := inlineModel()
	p.ID = types.StringValue("container-1")
	state := stateFor(t, p)
	read := resource.ReadResponse{State: state}
	r.Read(ctx, resource.ReadRequest{State: state}, &read)
	require.False(t, read.Diagnostics.HasError())
	require.True(t, read.State.Raw.IsNull())
}

func verifyQuery(t *testing.T, op, query string) {
	t.Helper()
	documents := map[string]string{
		"containerFqdnCreateFromList": cato.ContainerFqdnCreateFromListDocument,
		"containerFqdnCreateFromURL":  cato.ContainerFqdnCreateFromURLDocument,
		"containerFqdnUpdateFromList": cato.ContainerFqdnUpdateFromListDocument,
		"containerFqdnUpdateFromURL":  cato.ContainerFqdnUpdateFromURLDocument,
		"containerFqdnDownloadFile":   cato.ContainerFqdnDownloadFileDocument,
		"containerList":               cato.ContainerListDocument, "containerDelete": cato.ContainerDeleteDocument}
	want, ok := documents[op]
	require.True(t, ok)
	actual, err := parser.ParseQuery(&ast.Source{Input: query})
	require.NoError(t, err)
	expected, err := parser.ParseQuery(&ast.Source{Input: want})
	require.NoError(t, err)
	av, ev := actual.Operations[0].VariableDefinitions, expected.Operations[0].VariableDefinitions
	require.Len(t, av, len(ev))
	for i, v := range ev {
		require.Equal(t, v.Variable, av[i].Variable)
		require.Equal(t, v.Type.String(), av[i].Type.String())
	}
}
