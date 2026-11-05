package mailinglist

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

type backend struct {
	object    map[string]any
	fail      string
	malformed bool
	calls     []string
	inputs    []map[string]any
}

func setup(t *testing.T, b *backend) *mailingListResource {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() { require.NoError(t, req.Body.Close()) }()
		var body struct {
			OperationName string         `json:"operationName"`
			Query         string         `json:"query"`
			Variables     map[string]any `json:"variables"`
		}
		require.NoError(t, json.NewDecoder(req.Body).Decode(&body))
		require.Equal(t, "account-123", body.Variables["accountId"])
		require.NotContains(t, body.Query, "audit")
		op := strings.TrimPrefix(body.OperationName, "notification")
		op = strings.ToLower(op[:1]) + op[1:]
		b.calls = append(b.calls, op)
		w.Header().Set("Content-Type", "application/json")
		if b.fail == op {
			require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"errors": []any{map[string]any{"message": "denied"}}}))
			return
		}
		if b.malformed {
			require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{}}))
			return
		}
		var payload any
		switch op {
		case "createMailingList", "updateMailingList":
			in := body.Variables[op+"Input"].(map[string]any)
			b.inputs = append(b.inputs, in)
			if op == "createMailingList" {
				b.object = map[string]any{"id": "1001", "admin": []any{}, "address": []any{}}
			} else {
				require.Equal(t, updateDocument, body.Query)
				require.Equal(t, map[string]any{"by": "ID", "input": "1001"}, in["mailingList"])
				require.Contains(t, in, "address")
				require.Contains(t, in, "admin")
			}
			for _, k := range []string{"name", "address", "admin"} {
				if v, ok := in[k]; ok {
					if k == "admin" {
						members := []any{}
						for _, member := range v.([]any) {
							ref := member.(map[string]any)
							members = append(members, map[string]any{"id": ref["input"], "name": "Admin"})
						}
						b.object[k] = members
					} else {
						b.object[k] = v
					}
				}
			}
			payload = map[string]any{"mailingList": b.object}
		case "mailingList", "deleteMailingList":
			require.Equal(t, map[string]any{"by": "ID", "input": "1001"}, body.Variables["mailingListRefInput"])
			payload = b.object
			if op == "deleteMailingList" {
				payload = map[string]any{"mailingList": b.object}
				b.object = nil
			}
		default:
			t.Fatalf("unexpected operation: %s", op)
		}
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"notification": map[string]any{op: payload}}}))
	}))
	t.Cleanup(server.Close)
	sdk, err := cato.New(server.URL, "test-token", "wrong-default", nil, nil)
	require.NoError(t, err)
	return &mailingListResource{client: &catoclient.Service{AccountId: "account-123", Catov2: catoclient.NewProviderSDKClient(sdk)}}
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
	return resourceModel{ID: types.StringUnknown(), Name: types.StringValue("test"), Address: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("test@example.invalid")}), Admin: types.SetValueMust(idname.RefObjectType, nil)}
}
func TestCRUD(t *testing.T) {
	ctx := context.Background()
	b := &backend{}
	r := setup(t, b)
	p := model()
	create := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
	r.Create(ctx, resource.CreateRequest{Plan: planFor(t, p)}, &create)
	require.False(t, create.Diagnostics.HasError(), create.Diagnostics)
	var current resourceModel
	require.False(t, create.State.Get(ctx, &current).HasError())
	require.Equal(t, "1001", current.ID.ValueString())
	read := resource.ReadResponse{State: create.State}
	r.Read(ctx, resource.ReadRequest{State: create.State}, &read)
	require.False(t, read.Diagnostics.HasError(), read.Diagnostics)
	p.ID = current.ID
	p.Name = types.StringValue("updated")
	p.Address = types.SetNull(types.StringType)
	p.Admin = types.SetValueMust(idname.RefObjectType, []attr.Value{types.ObjectValueMust(idname.RefAttrTypes, map[string]attr.Value{"id": types.StringValue("42"), "name": types.StringUnknown()})})
	update := resource.UpdateResponse{State: read.State}
	r.Update(ctx, resource.UpdateRequest{Plan: planFor(t, p), State: read.State}, &update)
	require.False(t, update.Diagnostics.HasError(), update.Diagnostics)
	require.Equal(t, []any{}, b.inputs[1]["address"])
	require.Equal(t, []any{map[string]any{"by": "ID", "input": "42"}}, b.inputs[1]["admin"])
	require.False(t, update.State.Get(ctx, &current).HasError())
	require.Empty(t, current.Address.Elements())
	require.False(t, current.Address.IsNull())
	require.False(t, current.Admin.IsNull())
	del := resource.DeleteResponse{State: update.State}
	r.Delete(ctx, resource.DeleteRequest{State: update.State}, &del)
	require.False(t, del.Diagnostics.HasError(), del.Diagnostics)
	require.True(t, del.State.Raw.IsNull())
}
func TestReadMissingAndFailures(t *testing.T) {
	for _, kind := range []string{"missing", "denied", "malformed", "different"} {
		t.Run(kind, func(t *testing.T) {
			b := &backend{}
			r := setup(t, b)
			p := model()
			p.ID = types.StringValue("1001")
			s := stateFor(t, p)
			if kind == "denied" {
				b.fail = "mailingList"
			}
			if kind == "malformed" {
				b.malformed = true
			}
			if kind == "different" {
				b.object = map[string]any{"id": "different", "name": "test"}
			}
			resp := resource.ReadResponse{State: s}
			r.Read(context.Background(), resource.ReadRequest{State: s}, &resp)
			require.Equal(t, kind != "missing", resp.Diagnostics.HasError(), resp.Diagnostics)
			require.Equal(t, kind == "missing", resp.State.Raw.IsNull())
		})
	}
}
func TestMutationFailures(t *testing.T) {
	for _, op := range []string{"createMailingList", "updateMailingList", "deleteMailingList"} {
		t.Run(op, func(t *testing.T) {
			b := &backend{fail: op, object: map[string]any{"id": "1001", "name": "test"}}
			r := setup(t, b)
			m := model()
			m.ID = types.StringValue("1001")
			s := stateFor(t, m)
			ctx := context.Background()
			switch op {
			case "createMailingList":
				resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
				r.Create(ctx, resource.CreateRequest{Plan: planFor(t, m)}, &resp)
				require.True(t, resp.Diagnostics.HasError())
			case "updateMailingList":
				resp := resource.UpdateResponse{State: s}
				r.Update(ctx, resource.UpdateRequest{Plan: planFor(t, m), State: s}, &resp)
				require.True(t, resp.Diagnostics.HasError())
				require.Equal(t, s.Raw, resp.State.Raw)
			case "deleteMailingList":
				resp := resource.DeleteResponse{State: s}
				r.Delete(ctx, resource.DeleteRequest{State: s}, &resp)
				require.True(t, resp.Diagnostics.HasError())
				require.False(t, resp.State.Raw.IsNull())
			}
		})
	}
}
func TestValidationAndImport(t *testing.T) {
	ctx := context.Background()
	for _, field := range []string{"name", "address", "admin"} {
		t.Run(field, func(t *testing.T) {
			b := &backend{}
			r := setup(t, b)
			m := model()
			switch field {
			case "name":
				m.Name = types.StringUnknown()
			case "address":
				m.Address = types.SetUnknown(types.StringType)
			case "admin":
				m.Admin = types.SetUnknown(idname.RefObjectType)
			}
			resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
			r.Create(ctx, resource.CreateRequest{Plan: planFor(t, m)}, &resp)
			require.True(t, resp.Diagnostics.HasError())
			require.Empty(t, b.calls)
		})
	}
	resp := resource.ImportStateResponse{State: stateFor(t, resourceModel{ID: types.StringNull(), Name: types.StringNull(), Address: types.SetNull(types.StringType), Admin: types.SetNull(idname.RefObjectType)})}
	NewResource().(*mailingListResource).ImportState(ctx, resource.ImportStateRequest{ID: "1001"}, &resp)
	require.False(t, resp.Diagnostics.HasError())
	var id types.String
	require.False(t, resp.State.GetAttribute(ctx, path.Root("id"), &id).HasError())
	require.Equal(t, "1001", id.ValueString())
}

func TestPartialCreateRetainsIdentity(t *testing.T) {
	ctx := context.Background()
	api := mocks.NewMailingListClient(t)
	result := new(cato.NotificationCreateMailingList)
	require.NoError(t, json.Unmarshal([]byte(`{"notification":{"createMailingList":{"mailingList":{"id":"1001"}}}}`), result))
	api.EXPECT().NotificationCreateMailingList(mock.Anything, "account-123", mock.Anything).Return(result, errors.New("partial failure")).Once()
	r := &mailingListResource{client: &catoclient.Service{AccountId: "account-123"}, apiClient: api}
	m := model()
	m.Admin = types.SetValueMust(idname.RefObjectType, []attr.Value{types.ObjectValueMust(idname.RefAttrTypes, map[string]attr.Value{"id": types.StringValue("42"), "name": types.StringUnknown()})})
	resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
	r.Create(ctx, resource.CreateRequest{Plan: planFor(t, m)}, &resp)
	require.True(t, resp.Diagnostics.HasError())
	var state resourceModel
	require.False(t, resp.State.Get(ctx, &state).HasError())
	require.Equal(t, "1001", state.ID.ValueString())
	for _, member := range state.Admin.Elements() {
		for _, value := range member.(types.Object).Attributes() {
			require.False(t, value.IsUnknown())
		}
	}
}
func TestMissingListClassification(t *testing.T) {
	missing := func(message string) error {
		list := gqlerror.List{&gqlerror.Error{Message: message}}
		return &clientv2.ErrorResponse{GqlErrors: &list}
	}
	exact := "GQL-User-Error: Account ID [account-123] subscription [1001] not found"
	require.True(t, missingList(missing(exact), "account-123", "1001"))
	require.False(t, missingList(missing(exact), "another", "1001"))
	require.False(t, missingList(missing("permission denied"), "account-123", "1001"))
	require.False(t, missingList(errors.New(exact), "account-123", "1001"))
}
func TestMetadataConfigure(t *testing.T) {
	ctx := context.Background()
	r := NewResource().(*mailingListResource)
	var meta resource.MetadataResponse
	r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "cato"}, &meta)
	require.Equal(t, "cato_mailing_list", meta.TypeName)
	var resp resource.ConfigureResponse
	r.Configure(ctx, resource.ConfigureRequest{}, &resp)
	require.False(t, resp.Diagnostics.HasError())
	r.Configure(ctx, resource.ConfigureRequest{ProviderData: "invalid"}, &resp)
	require.True(t, resp.Diagnostics.HasError())
	resp = resource.ConfigureResponse{}
	r.Configure(ctx, resource.ConfigureRequest{ProviderData: &catoclient.Service{AccountId: "account-123"}}, &resp)
	require.False(t, resp.Diagnostics.HasError())
}

func TestRejectUnsupportedMembership(t *testing.T) {
	for _, kind := range []string{"empty", "admin_name", "null_admin"} {
		t.Run(kind, func(t *testing.T) {
			m := model()
			switch kind {
			case "empty":
				m.Address = types.SetNull(types.StringType)
			case "admin_name":
				m.Admin = types.SetValueMust(idname.RefObjectType, []attr.Value{types.ObjectValueMust(idname.RefAttrTypes, map[string]attr.Value{"id": types.StringNull(), "name": types.StringValue("Admin")})})
			case "null_admin":
				m.Admin = types.SetValueMust(idname.RefObjectType, []attr.Value{types.ObjectNull(idname.RefAttrTypes)})
			}
			var diags diag.Diagnostics
			buildCreate(context.Background(), m, &diags)
			require.True(t, diags.HasError())
		})
	}
}
