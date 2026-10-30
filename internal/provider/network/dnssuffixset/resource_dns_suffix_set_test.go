package dnssuffixset

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

type backend struct {
	object    map[string]any
	fail      string
	malformed bool
	calls     []string
	inputs    []map[string]any
}

func setup(t *testing.T, b *backend) *dnsSetResource {
	t.Helper()
	suffix := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() { require.NoError(t, req.Body.Close()) }()
		var body struct {
			OperationName string         `json:"operationName"`
			Query         string         `json:"query"`
			Variables     map[string]any `json:"variables"`
		}
		require.NoError(t, json.NewDecoder(req.Body).Decode(&body))
		require.Equal(t, "account-123", body.Variables["accountId"])
		op := strings.TrimPrefix(body.OperationName, "networkConfigDns")
		op = strings.ToLower(op[:1]) + op[1:]
		b.calls = append(b.calls, op)
		require.NotContains(t, body.Query, "audit")
		w.Header().Set("Content-Type", "application/json")
		if op == b.fail {
			require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"errors": []any{map[string]any{"message": "denied"}}}))
			return
		}
		if b.malformed {
			require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{}}))
			return
		}
		var payload any
		switch op {
		case "createSuffixSet", "updateSuffixSet":
			in := body.Variables["networkConfigDns"+strings.ToUpper(op[:1])+op[1:]+"Input"].(map[string]any)
			b.inputs = append(b.inputs, in)
			items := in["dnsSuffixSet"].([]any)
			require.Len(t, items, 1)
			item := items[0].(map[string]any)
			if op == "createSuffixSet" {
				b.object = map[string]any{"id": "dns-set-1"}
			} else {
				require.Equal(t, "dns-set-1", item["id"])
				require.Contains(t, item, "suffix")
			}
			b.object["name"] = item["name"]
			b.object["suffix"] = item["suffix"]
			payload = map[string]any{"dnsSuffixSet": []any{b.object}}
		case "suffixSet":
			require.Equal(t, map[string]any{"by": "ID", "input": "dns-set-1"}, body.Variables["networkConfigDnsSuffixSetRefInput"])
			payload = b.object
		case "deleteSuffixSet":
			in := body.Variables["networkConfigDnsDeleteSuffixSetInput"].(map[string]any)
			require.Equal(t, []any{"dns-set-1"}, in["id"])
			payload = map[string]any{"dnsSuffixSet": []any{b.object}}
			b.object = nil
		default:
			t.Fatalf("unexpected or account-wide operation: %s", op)
		}
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"networkConfig": map[string]any{"dns": map[string]any{op: payload}}}}))
	}))
	t.Cleanup(suffix.Close)
	sdk, err := cato.New(suffix.URL, "test-token", "wrong-default", nil, nil)
	require.NoError(t, err)
	return &dnsSetResource{client: &catoclient.Service{AccountId: "account-123", Catov2: catoclient.NewProviderSDKClient(sdk)}}
}
func schemaFor(t *testing.T) schema.Schema {
	t.Helper()
	var resp resource.SchemaResponse
	NewResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp.Schema
}
func planFor(t *testing.T, m resourceModel) tfsdk.Plan {
	t.Helper()
	p := tfsdk.Plan{Schema: schemaFor(t)}
	require.False(t, p.Set(context.Background(), m).HasError())
	return p
}
func stateFor(t *testing.T, m resourceModel) tfsdk.State {
	t.Helper()
	s := tfsdk.State{Schema: schemaFor(t)}
	require.False(t, s.Set(context.Background(), m).HasError())
	return s
}
func model() resourceModel {
	return resourceModel{ID: types.StringUnknown(), Name: types.StringValue("test"), Suffix: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("first.example.invalid"), types.StringValue("second.example.invalid")})}
}
func TestCRUDAndOrder(t *testing.T) {
	ctx := context.Background()
	b := &backend{}
	r := setup(t, b)
	p := model()
	create := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
	r.Create(ctx, resource.CreateRequest{Plan: planFor(t, p)}, &create)
	require.False(t, create.Diagnostics.HasError(), create.Diagnostics)
	var current resourceModel
	require.False(t, create.State.Get(ctx, &current).HasError())
	require.Equal(t, "dns-set-1", current.ID.ValueString())
	require.Equal(t, p.Suffix, current.Suffix)
	p.ID = current.ID
	p.Name = types.StringValue("updated")
	p.Suffix = types.ListValueMust(types.StringType, []attr.Value{types.StringValue("second.example.invalid"), types.StringValue("first.example.invalid")})
	update := resource.UpdateResponse{State: create.State}
	r.Update(ctx, resource.UpdateRequest{Plan: planFor(t, p), State: create.State}, &update)
	require.False(t, update.Diagnostics.HasError(), update.Diagnostics)
	require.False(t, update.State.Get(ctx, &current).HasError())
	require.Equal(t, p.Suffix, current.Suffix)
	p.Suffix = types.ListValueMust(types.StringType, nil)
	clearResponse := resource.UpdateResponse{State: update.State}
	r.Update(ctx, resource.UpdateRequest{Plan: planFor(t, p), State: update.State}, &clearResponse)
	require.False(t, clearResponse.Diagnostics.HasError(), clearResponse.Diagnostics)
	require.False(t, clearResponse.State.Get(ctx, &current).HasError())
	require.Empty(t, current.Suffix.Elements())
	require.False(t, current.Suffix.IsNull())
	b.object["name"] = "outside"
	read := resource.ReadResponse{State: clearResponse.State}
	r.Read(ctx, resource.ReadRequest{State: clearResponse.State}, &read)
	require.False(t, read.Diagnostics.HasError())
	require.False(t, read.State.Get(ctx, &current).HasError())
	require.Equal(t, "outside", current.Name.ValueString())
	del := resource.DeleteResponse{State: read.State}
	r.Delete(ctx, resource.DeleteRequest{State: read.State}, &del)
	require.False(t, del.Diagnostics.HasError(), del.Diagnostics)
	require.True(t, del.State.Raw.IsNull())
}
func TestReadMissingAndFailures(t *testing.T) {
	for _, kind := range []string{"missing", "denied", "malformed", "different"} {
		t.Run(kind, func(t *testing.T) {
			b := &backend{}
			r := setup(t, b)
			p := model()
			p.ID = types.StringValue("dns-set-1")
			s := stateFor(t, p)
			if kind == "denied" {
				b.fail = "suffixSet"
			}
			if kind == "malformed" {
				b.malformed = true
			}
			if kind == "different" {
				b.object = map[string]any{"id": "other", "name": "other", "suffix": []any{}}
			}
			resp := resource.ReadResponse{State: s}
			r.Read(context.Background(), resource.ReadRequest{State: s}, &resp)
			require.Equal(t, kind != "missing", resp.Diagnostics.HasError(), resp.Diagnostics)
			require.Equal(t, kind == "missing", resp.State.Raw.IsNull())
		})
	}
}
func TestMutationErrors(t *testing.T) {
	for _, op := range []string{"createSuffixSet", "updateSuffixSet", "deleteSuffixSet"} {
		t.Run(op, func(t *testing.T) {
			b := &backend{fail: op, object: map[string]any{"id": "dns-set-1", "name": "test", "suffix": []any{}}}
			r := setup(t, b)
			p := model()
			p.ID = types.StringValue("dns-set-1")
			s := stateFor(t, p)
			ctx := context.Background()
			switch op {
			case "createSuffixSet":
				resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
				r.Create(ctx, resource.CreateRequest{Plan: planFor(t, p)}, &resp)
				require.True(t, resp.Diagnostics.HasError())
			case "updateSuffixSet":
				resp := resource.UpdateResponse{State: s}
				r.Update(ctx, resource.UpdateRequest{Plan: planFor(t, p), State: s}, &resp)
				require.True(t, resp.Diagnostics.HasError())
				require.Equal(t, s.Raw, resp.State.Raw)
			case "deleteSuffixSet":
				resp := resource.DeleteResponse{State: s}
				r.Delete(ctx, resource.DeleteRequest{State: s}, &resp)
				require.True(t, resp.Diagnostics.HasError())
				require.False(t, resp.State.Raw.IsNull())
			}
		})
	}
}
func TestNullUnknownValidation(t *testing.T) {
	for _, kind := range []string{"name", "null", "unknown", "null_element", "unknown_element", "duplicate", "invalid"} {
		t.Run(kind, func(t *testing.T) {
			b := &backend{}
			r := setup(t, b)
			p := model()
			switch kind {
			case "name":
				p.Name = types.StringUnknown()
			case "null":
				p.Suffix = types.ListNull(types.StringType)
			case "unknown":
				p.Suffix = types.ListUnknown(types.StringType)
			case "null_element":
				p.Suffix = types.ListValueMust(types.StringType, []attr.Value{types.StringNull()})
			case "unknown_element":
				p.Suffix = types.ListValueMust(types.StringType, []attr.Value{types.StringUnknown()})
			case "duplicate":
				p.Suffix = types.ListValueMust(types.StringType, []attr.Value{types.StringValue("first.example.invalid"), types.StringValue("first.example.invalid")})
			case "invalid":
				p.Suffix = types.ListValueMust(types.StringType, []attr.Value{types.StringValue("invalid value")})
			}
			resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
			r.Create(context.Background(), resource.CreateRequest{Plan: planFor(t, p)}, &resp)
			require.True(t, resp.Diagnostics.HasError())
			require.Empty(t, b.calls)
		})
	}
}
func TestImportMetadataConfigure(t *testing.T) {
	ctx := context.Background()
	r := NewResource().(*dnsSetResource)
	var meta resource.MetadataResponse
	r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "cato"}, &meta)
	require.Equal(t, "cato_dns_suffix_set", meta.TypeName)
	var config resource.ConfigureResponse
	r.Configure(ctx, resource.ConfigureRequest{}, &config)
	require.False(t, config.Diagnostics.HasError())
	r.Configure(ctx, resource.ConfigureRequest{ProviderData: "invalid"}, &config)
	require.True(t, config.Diagnostics.HasError())
	resp := resource.ImportStateResponse{State: stateFor(t, resourceModel{ID: types.StringNull(), Name: types.StringNull(), Suffix: types.ListNull(types.StringType)})}
	r.ImportState(ctx, resource.ImportStateRequest{ID: "dns-set-1"}, &resp)
	require.False(t, resp.Diagnostics.HasError())
	var id types.String
	require.False(t, resp.State.GetAttribute(ctx, path.Root("id"), &id).HasError())
	require.Equal(t, "dns-set-1", id.ValueString())
}
func TestPartialCreateIdentity(t *testing.T) {
	ctx := context.Background()
	api := mocks.NewDNSSuffixSetClient(t)
	result := new(cato.NetworkConfigDNSCreateSuffixSet)
	require.NoError(t, json.Unmarshal([]byte(`{"networkConfig":{"dns":{"createSuffixSet":{"dnsSuffixSet":[{"id":"dns-set-1"}]}}}}`), result))
	api.EXPECT().NetworkConfigDNSCreateSuffixSet(mock.Anything, "account-123", mock.Anything).Return(result, errors.New("partial failure")).Once()
	r := &dnsSetResource{client: &catoclient.Service{AccountId: "account-123"}, apiClient: api}
	resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
	r.Create(ctx, resource.CreateRequest{Plan: planFor(t, model())}, &resp)
	require.True(t, resp.Diagnostics.HasError())
	var state resourceModel
	require.False(t, resp.State.Get(ctx, &state).HasError())
	require.Equal(t, "dns-set-1", state.ID.ValueString())
}

func TestRejectNameNormalization(t *testing.T) {
	b := &backend{}
	r := setup(t, b)
	p := model()
	p.Name = types.StringValue(" test ")
	resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
	r.Create(context.Background(), resource.CreateRequest{Plan: planFor(t, p)}, &resp)
	require.True(t, resp.Diagnostics.HasError())
	require.Empty(t, b.calls)
}

func TestSuffixLengthAndWhitespace(t *testing.T) {
	require.True(t, validMember(strings.Repeat("a", 253)))
	require.False(t, validMember(strings.Repeat("a", 254)))
	require.True(t, validMember(strings.Repeat("😀", 126)))
	require.False(t, validMember(strings.Repeat("😀", 127)))
	require.False(t, validMember(" example.invalid"))
	require.False(t, validMember("example.invalid\t"))
	require.True(t, validMember("Mixed.Example.Invalid"))
}
