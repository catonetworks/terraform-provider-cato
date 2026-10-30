package popbgpprofile

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

func remote() map[string]any {
	return map[string]any{"id": "1001", "name": "test", "description": "test", "communities": []any{map[string]any{"from": float64(65001), "to": float64(100)}}}
}
func setup(t *testing.T, b *backend) *bgpProfileResource {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() { require.NoError(t, req.Body.Close()) }()
		var body struct {
			OperationName string         `json:"operationName"`
			Variables     map[string]any `json:"variables"`
		}
		require.NoError(t, json.NewDecoder(req.Body).Decode(&body))
		require.Equal(t, "account-123", body.Variables["accountId"])
		b.calls = append(b.calls, body.OperationName)
		w.Header().Set("Content-Type", "application/json")
		if b.fail == body.OperationName {
			require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"errors": []any{map[string]any{"message": "denied"}}}))
			return
		}
		if b.malformed {
			require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{}}))
			return
		}
		root, payload := "popLocationMutations", map[string]any{}
		switch body.OperationName {
		case "popLocationMutationsAddBgpProfile":
			in := body.Variables["popLocationAddBgpProfileInput"].(map[string]any)
			b.inputs = append(b.inputs, in)
			b.object = remote()
			for _, key := range []string{"name", "description", "communities"} {
				if v, ok := in[key]; ok {
					b.object[key] = v
				}
			}
			payload["addBgpProfile"] = map[string]any{"bgpProfile": b.object}
		case "popLocationMutationsUpdateBgpProfile":
			in := body.Variables["popLocationUpdateBgpProfileInput"].(map[string]any)
			b.inputs = append(b.inputs, in)
			require.Equal(t, "1001", in["id"])
			for _, key := range []string{"name", "description", "communities"} {
				if v, ok := in[key]; ok {
					b.object[key] = v
				}
			}
			payload["updateBgpProfile"] = map[string]any{"bgpProfile": b.object}
		case "popLocationMutationsRemoveBgpProfile":
			require.Equal(t, "1001", body.Variables["bgpProfileId"])
			payload["removeBgpProfile"] = map[string]any{"bgpProfile": b.object}
			b.object = nil
		case "popLocationsBgpProfile":
			root = "popLocations"
			require.Equal(t, "1001", body.Variables["id"])
			if b.object == nil {
				require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"errors": []any{map[string]any{"message": "BGP profile not found"}}}))
				return
			}
			payload["bgpProfile"] = b.object
		default:
			t.Fatalf("unexpected operation %s", body.OperationName)
		}
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{root: payload}}))
	}))
	t.Cleanup(server.Close)
	sdk, err := cato.New(server.URL, "test-token", "wrong-default", nil, nil)
	require.NoError(t, err)
	return &bgpProfileResource{client: &catoclient.Service{AccountId: "account-123", Catov2: catoclient.NewProviderSDKClient(sdk)}}
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
func community(from, to string) types.Object {
	return types.ObjectValueMust(communityAttrTypes, map[string]attr.Value{"from": types.StringValue(from), "to": types.StringValue(to)})
}
func model() resourceModel {
	return resourceModel{ID: types.StringUnknown(), Name: types.StringValue("test"), Description: types.StringValue("test"), Communities: types.SetValueMust(communityObjectType, []attr.Value{community("65001", "100")})}
}
func TestCRUDAndDescriptionOnly(t *testing.T) {
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
	require.True(t, p.Communities.Equal(current.Communities))
	p = current
	p.Description = types.StringNull()
	update := resource.UpdateResponse{State: create.State}
	r.Update(ctx, resource.UpdateRequest{Plan: planFor(t, p), State: create.State}, &update)
	require.False(t, update.Diagnostics.HasError(), update.Diagnostics)
	require.Equal(t, "", b.inputs[1]["description"])
	require.NotContains(t, b.inputs[1], "communities")
	p.Name = types.StringValue("updated")
	p.Communities = types.SetValueMust(communityObjectType, []attr.Value{community("65002", "200")})
	change := resource.UpdateResponse{State: update.State}
	r.Update(ctx, resource.UpdateRequest{Plan: planFor(t, p), State: update.State}, &change)
	require.False(t, change.Diagnostics.HasError(), change.Diagnostics)
	require.Contains(t, b.inputs[2], "communities")
	b.object["name"] = "outside"
	read := resource.ReadResponse{State: change.State}
	r.Read(ctx, resource.ReadRequest{State: change.State}, &read)
	require.False(t, read.Diagnostics.HasError(), read.Diagnostics)
	require.False(t, read.State.Get(ctx, &current).HasError())
	require.Equal(t, "outside", current.Name.ValueString())
	del := resource.DeleteResponse{State: read.State}
	r.Delete(ctx, resource.DeleteRequest{State: read.State}, &del)
	require.False(t, del.Diagnostics.HasError(), del.Diagnostics)
	require.True(t, del.State.Raw.IsNull())
}
func TestReadFailuresAndMissing(t *testing.T) {
	for _, kind := range []string{"missing", "denied", "malformed", "wrong_id"} {
		t.Run(kind, func(t *testing.T) {
			b := &backend{}
			r := setup(t, b)
			p := model()
			p.ID = types.StringValue("1001")
			s := stateFor(t, p)
			switch kind {
			case "denied":
				b.fail = "popLocationsBgpProfile"
			case "malformed":
				b.malformed = true
			case "wrong_id":
				b.object = remote()
				b.object["id"] = "other"
			}
			resp := resource.ReadResponse{State: s}
			r.Read(context.Background(), resource.ReadRequest{State: s}, &resp)
			require.Equal(t, kind != "missing", resp.Diagnostics.HasError(), resp.Diagnostics)
			require.Equal(t, kind == "missing", resp.State.Raw.IsNull())
		})
	}
}
func TestValidationBeforeMutation(t *testing.T) {
	for _, kind := range []string{"name", "description", "null", "unknown", "empty", "unknown_member", "null_member", "range", "canonical"} {
		t.Run(kind, func(t *testing.T) {
			b := &backend{}
			r := setup(t, b)
			p := model()
			switch kind {
			case "name":
				p.Name = types.StringUnknown()
			case "description":
				p.Description = types.StringUnknown()
			case "null":
				p.Communities = types.SetNull(communityObjectType)
			case "unknown":
				p.Communities = types.SetUnknown(communityObjectType)
			case "empty":
				p.Communities = types.SetValueMust(communityObjectType, nil)
			case "unknown_member":
				p.Communities = types.SetValueMust(communityObjectType, []attr.Value{types.ObjectUnknown(communityAttrTypes)})
			case "null_member":
				p.Communities = types.SetValueMust(communityObjectType, []attr.Value{types.ObjectNull(communityAttrTypes)})
			case "range":
				p.Communities = types.SetValueMust(communityObjectType, []attr.Value{community("65536", "1")})
			case "canonical":
				p.Communities = types.SetValueMust(communityObjectType, []attr.Value{community("01", "1")})
			}
			resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
			r.Create(context.Background(), resource.CreateRequest{Plan: planFor(t, p)}, &resp)
			require.True(t, resp.Diagnostics.HasError())
			require.Empty(t, b.calls)
		})
	}
}
func TestMutationErrors(t *testing.T) {
	for _, op := range []string{"popLocationMutationsAddBgpProfile", "popLocationMutationsUpdateBgpProfile", "popLocationMutationsRemoveBgpProfile"} {
		t.Run(op, func(t *testing.T) {
			b := &backend{fail: op, object: remote()}
			r := setup(t, b)
			p := model()
			p.ID = types.StringValue("1001")
			s := stateFor(t, p)
			ctx := context.Background()
			switch op {
			case "popLocationMutationsAddBgpProfile":
				resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
				r.Create(ctx, resource.CreateRequest{Plan: planFor(t, p)}, &resp)
				require.True(t, resp.Diagnostics.HasError())
			case "popLocationMutationsUpdateBgpProfile":
				resp := resource.UpdateResponse{State: s}
				r.Update(ctx, resource.UpdateRequest{Plan: planFor(t, p), State: s}, &resp)
				require.True(t, resp.Diagnostics.HasError())
				require.Equal(t, s.Raw, resp.State.Raw)
			case "popLocationMutationsRemoveBgpProfile":
				resp := resource.DeleteResponse{State: s}
				r.Delete(ctx, resource.DeleteRequest{State: s}, &resp)
				require.True(t, resp.Diagnostics.HasError())
				require.False(t, resp.State.Raw.IsNull())
			}
		})
	}
}
func TestPartialCreateIdentity(t *testing.T) {
	ctx := context.Background()
	api := mocks.NewPopBGPProfileClient(t)
	result := new(cato.PopLocationMutationsAddBgpProfile)
	require.NoError(t, json.Unmarshal([]byte(`{"popLocationMutations":{"addBgpProfile":{"bgpProfile":{"id":"1001"}}}}`), result))
	api.EXPECT().PopLocationMutationsAddBgpProfile(mock.Anything, "account-123", mock.Anything).Return(result, errors.New("partial failure")).Once()
	r := &bgpProfileResource{client: &catoclient.Service{AccountId: "account-123"}, apiClient: api}
	resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
	r.Create(ctx, resource.CreateRequest{Plan: planFor(t, model())}, &resp)
	require.True(t, resp.Diagnostics.HasError())
	var current resourceModel
	require.False(t, resp.State.Get(ctx, &current).HasError())
	require.Equal(t, "1001", current.ID.ValueString())
}
func TestImportAndMetadata(t *testing.T) {
	ctx := context.Background()
	r := NewResource().(*bgpProfileResource)
	var meta resource.MetadataResponse
	r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "cato"}, &meta)
	require.Equal(t, "cato_pop_bgp_profile", meta.TypeName)
	p := model()
	p.ID = types.StringNull()
	resp := resource.ImportStateResponse{State: stateFor(t, p)}
	r.ImportState(ctx, resource.ImportStateRequest{ID: "1001"}, &resp)
	require.False(t, resp.Diagnostics.HasError())
	var id types.String
	require.False(t, resp.State.GetAttribute(ctx, path.Root("id"), &id).HasError())
	require.Equal(t, "1001", id.ValueString())
	require.True(t, validASN("65535"))
	require.True(t, validASN("0"))
	require.False(t, validASN(strings.Repeat("9", 10)))
}
