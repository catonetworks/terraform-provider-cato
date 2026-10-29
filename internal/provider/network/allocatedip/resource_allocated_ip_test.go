package allocatedip

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
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
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

type backend struct {
	object    map[string]any
	fail      string
	malformed bool
	calls     []string
	inputs    []map[string]any
}

func remote() map[string]any {
	return map[string]any{"id": "allocation-1", "account": map[string]any{"id": "account-123", "name": "Test"}, "ip": "192.0.2.10", "allocationType": "SYSTEM", "popLocation": map[string]any{"id": "1", "name": "Test Pop"}, "description": "test"}
}
func setup(t *testing.T, b *backend) *allocatedIPResource {
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
		case "popLocationMutationsAddAllocatedIp":
			in := body.Variables["popLocationAddAllocatedIpInput"].(map[string]any)
			b.inputs = append(b.inputs, in)
			require.Equal(t, "SYSTEM", in["type"])
			require.Equal(t, map[string]any{"by": "ID", "input": "1"}, in["popLocation"])
			b.object = remote()
			if value, ok := in["description"]; ok {
				b.object["description"] = value
			}
			payload["addAllocatedIp"] = map[string]any{"allocatedIp": b.object}
		case "popLocationMutationsUpdateAllocatedIpDescription":
			in := body.Variables["popLocationUpdateAllocatedIpDescriptionInput"].(map[string]any)
			b.inputs = append(b.inputs, in)
			require.Equal(t, "192.0.2.10", in["ip"])
			b.object["description"] = in["description"]
			payload["updateAllocatedIpDescription"] = map[string]any{"allocatedIp": b.object}
		case "popLocationMutationsRemoveAllocatedIp":
			in := body.Variables["popLocationRemoveAllocatedIpInput"].(map[string]any)
			b.inputs = append(b.inputs, in)
			require.Equal(t, "192.0.2.10", in["ip"])
			require.Equal(t, "SYSTEM", in["type"])
			payload["removeAllocatedIp"] = map[string]any{"allocatedIp": b.object}
			b.object = nil
		case "popLocationsAllocatedIpList":
			root = "popLocations"
			in := body.Variables["popLocationAllocatedIpInput"].(map[string]any)
			require.Equal(t, map[string]any{"eq": "allocation-1"}, in["filter"].(map[string]any)["id"])
			items := []any{}
			if b.object != nil {
				items = append(items, b.object)
			}
			payload["allocatedIpList"] = map[string]any{"allocatedIp": items, "paging": map[string]any{"total": len(items)}}
		default:
			t.Fatalf("unexpected operation %s", body.OperationName)
		}
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{root: payload}}))
	}))
	t.Cleanup(server.Close)
	sdk, err := cato.New(server.URL, "test-token", "wrong-default", nil, nil)
	require.NoError(t, err)
	return &allocatedIPResource{client: &catoclient.Service{AccountId: "account-123", Catov2: catoclient.NewProviderSDKClient(sdk)}}
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
	return resourceModel{ID: types.StringUnknown(), IP: types.StringUnknown(), Subnet: types.StringUnknown(), State: types.StringUnknown(), Type: types.StringValue("SYSTEM"), Description: types.StringValue("test"), PopLocation: types.ObjectValueMust(idname.RefAttrTypes, map[string]attr.Value{"id": types.StringValue("1"), "name": types.StringUnknown()})}
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
	require.Equal(t, "allocation-1", current.ID.ValueString())
	require.Equal(t, "192.0.2.10", current.IP.ValueString())
	require.Equal(t, "", current.Subnet.ValueString())
	require.Equal(t, "", current.State.ValueString())
	current.IP = types.StringValue("stale-address")
	s := stateFor(t, current)
	p = current
	p.Description = types.StringNull()
	update := resource.UpdateResponse{State: s}
	r.Update(ctx, resource.UpdateRequest{Plan: planFor(t, p), State: s}, &update)
	require.False(t, update.Diagnostics.HasError(), update.Diagnostics)
	require.Equal(t, "", b.inputs[1]["description"])
	require.False(t, update.State.Get(ctx, &current).HasError())
	require.Equal(t, "192.0.2.10", current.IP.ValueString())
	require.Equal(t, "", current.Description.ValueString())
	b.object["description"] = "outside"
	read := resource.ReadResponse{State: update.State}
	r.Read(ctx, resource.ReadRequest{State: update.State}, &read)
	require.False(t, read.Diagnostics.HasError(), read.Diagnostics)
	require.False(t, read.State.Get(ctx, &current).HasError())
	require.Equal(t, "outside", current.Description.ValueString())
	del := resource.DeleteResponse{State: read.State}
	r.Delete(ctx, resource.DeleteRequest{State: read.State}, &del)
	require.False(t, del.Diagnostics.HasError(), del.Diagnostics)
	require.True(t, del.State.Raw.IsNull())
}
func TestReadMissingAndInvalid(t *testing.T) {
	for _, kind := range []string{"missing", "denied", "malformed", "wrong_account", "wrong_id", "byoip"} {
		t.Run(kind, func(t *testing.T) {
			b := &backend{}
			r := setup(t, b)
			p := model()
			p.ID = types.StringValue("allocation-1")
			s := stateFor(t, p)
			switch kind {
			case "denied":
				b.fail = "popLocationsAllocatedIpList"
			case "malformed":
				b.malformed = true
			case "wrong_account":
				b.object = remote()
				b.object["account"] = map[string]any{"id": "another"}
			case "wrong_id":
				b.object = remote()
				b.object["id"] = "another"
			case "byoip":
				b.object = remote()
				b.object["allocationType"] = "BYOIP"
			}
			resp := resource.ReadResponse{State: s}
			r.Read(context.Background(), resource.ReadRequest{State: s}, &resp)
			require.Equal(t, kind != "missing", resp.Diagnostics.HasError(), resp.Diagnostics)
			require.Equal(t, kind == "missing", resp.State.Raw.IsNull())
		})
	}
}
func TestValidationBeforeMutation(t *testing.T) {
	for _, kind := range []string{"unknown_type", "byoip", "unknown_description", "null_pop", "unknown_pop", "empty_reference"} {
		t.Run(kind, func(t *testing.T) {
			b := &backend{}
			r := setup(t, b)
			p := model()
			switch kind {
			case "unknown_type":
				p.Type = types.StringUnknown()
			case "byoip":
				p.Type = types.StringValue("BYOIP")
			case "unknown_description":
				p.Description = types.StringUnknown()
			case "null_pop":
				p.PopLocation = types.ObjectNull(idname.RefAttrTypes)
			case "unknown_pop":
				p.PopLocation = types.ObjectUnknown(idname.RefAttrTypes)
			case "empty_reference":
				p.PopLocation = types.ObjectValueMust(idname.RefAttrTypes, map[string]attr.Value{"id": types.StringNull(), "name": types.StringNull()})
			}
			resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
			r.Create(context.Background(), resource.CreateRequest{Plan: planFor(t, p)}, &resp)
			require.True(t, resp.Diagnostics.HasError())
			require.Empty(t, b.calls)
		})
	}
}
func TestMutationFailures(t *testing.T) {
	for _, op := range []string{"popLocationMutationsAddAllocatedIp", "popLocationMutationsUpdateAllocatedIpDescription", "popLocationMutationsRemoveAllocatedIp"} {
		t.Run(op, func(t *testing.T) {
			b := &backend{fail: op, object: remote()}
			r := setup(t, b)
			p := model()
			p.ID = types.StringValue("allocation-1")
			s := stateFor(t, p)
			ctx := context.Background()
			switch op {
			case "popLocationMutationsAddAllocatedIp":
				resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
				r.Create(ctx, resource.CreateRequest{Plan: planFor(t, p)}, &resp)
				require.True(t, resp.Diagnostics.HasError())
			case "popLocationMutationsUpdateAllocatedIpDescription":
				resp := resource.UpdateResponse{State: s}
				r.Update(ctx, resource.UpdateRequest{Plan: planFor(t, p), State: s}, &resp)
				require.True(t, resp.Diagnostics.HasError())
				require.Equal(t, s.Raw, resp.State.Raw)
			case "popLocationMutationsRemoveAllocatedIp":
				resp := resource.DeleteResponse{State: s}
				r.Delete(ctx, resource.DeleteRequest{State: s}, &resp)
				require.True(t, resp.Diagnostics.HasError())
				require.False(t, resp.State.Raw.IsNull())
			}
		})
	}
}
func TestPartialCreateKeepsID(t *testing.T) {
	ctx := context.Background()
	api := mocks.NewAllocatedIPResourceClient(t)
	result := new(cato.PopLocationMutationsAddAllocatedIP)
	require.NoError(t, json.Unmarshal([]byte(`{"popLocationMutations":{"addAllocatedIp":{"allocatedIp":{"id":"allocation-1","ip":"192.0.2.10"}}}}`), result))
	api.EXPECT().PopLocationMutationsAddAllocatedIP(mock.Anything, "account-123", mock.Anything).Return(result, errors.New("partial failure")).Once()
	r := &allocatedIPResource{client: &catoclient.Service{AccountId: "account-123"}, apiClient: api}
	resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
	r.Create(ctx, resource.CreateRequest{Plan: planFor(t, model())}, &resp)
	require.True(t, resp.Diagnostics.HasError())
	var current resourceModel
	require.False(t, resp.State.Get(ctx, &current).HasError())
	require.Equal(t, "allocation-1", current.ID.ValueString())
	require.False(t, current.IP.IsUnknown())
	require.False(t, current.Subnet.IsUnknown())
	require.False(t, current.State.IsUnknown())
	require.False(t, current.PopLocation.Attributes()["name"].IsUnknown())
}
func TestImportAndMetadata(t *testing.T) {
	ctx := context.Background()
	r := NewResource().(*allocatedIPResource)
	var meta resource.MetadataResponse
	r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "cato"}, &meta)
	require.Equal(t, "cato_allocated_ip", meta.TypeName)
	p := model()
	p.ID = types.StringNull()
	resp := resource.ImportStateResponse{State: stateFor(t, p)}
	r.ImportState(ctx, resource.ImportStateRequest{ID: "allocation-1"}, &resp)
	require.False(t, resp.Diagnostics.HasError())
	var id types.String
	require.False(t, resp.State.GetAttribute(ctx, path.Root("id"), &id).HasError())
	require.Equal(t, "allocation-1", id.ValueString())
}
