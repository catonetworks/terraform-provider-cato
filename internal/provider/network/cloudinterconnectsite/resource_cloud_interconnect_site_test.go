package cloudinterconnectsite

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
)

type backend struct {
	object          map[string]any
	fail            string
	connection      string
	foreignAccount  bool
	snapshotMissing bool
	malformed       bool
	roles           map[string]bool
	calls           []string
	inputs          []map[string]any
}

func remote() map[string]any {
	return map[string]any{"site": map[string]any{"id": "1001", "name": "test"}, "siteType": "BRANCH", "description": "test", "siteLocation": map[string]any{"countryCode": "GB", "timezone": "Europe/London", "cityName": "London", "stateCode": "", "address": ""}}
}
func setup(t *testing.T, b *backend) *cloudSiteResource {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() { require.NoError(t, req.Body.Close()) }()
		var body struct {
			OperationName string         `json:"operationName"`
			Variables     map[string]any `json:"variables"`
		}
		require.NoError(t, json.NewDecoder(req.Body).Decode(&body))
		b.calls = append(b.calls, body.OperationName)
		w.Header().Set("Content-Type", "application/json")
		if b.fail == body.OperationName {
			writeError(t, w, "denied")
			return
		}
		if b.malformed {
			require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{}}))
			return
		}
		if body.OperationName == "entityLookup" {
			require.Equal(t, "account-123", body.Variables["accountID"])
			require.Equal(t, []any{"1001"}, body.Variables["entityIDs"])
			items := []any{}
			if b.object != nil {
				items = append(items, map[string]any{"entity": map[string]any{"id": "1001", "name": "test", "type": "site"}})
			}
			require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"entityLookup": map[string]any{"items": items}}}))
			return
		}
		if body.OperationName == "accountSnapshot" {
			writeSnapshot(t, w, b, body.Variables)
			return
		}
		require.Equal(t, "account-123", body.Variables["accountId"])
		payload := map[string]any{}
		switch body.OperationName {
		case "siteAddCloudInterconnectSite":
			in := body.Variables["addCloudInterconnectSiteInput"].(map[string]any)
			b.inputs = append(b.inputs, in)
			b.object = remote()
			applyInput(b, in, true)
			payload["addCloudInterconnectSite"] = map[string]any{"siteId": "1001"}
		case "siteUpdateSiteGeneralDetails":
			require.Equal(t, "1001", body.Variables["siteId"])
			in := body.Variables["updateSiteGeneralDetailsInput"].(map[string]any)
			b.inputs = append(b.inputs, in)
			applyInput(b, in, false)
			payload["updateSiteGeneralDetails"] = map[string]any{"siteId": "1001"}
		case "siteGeneralDetails":
			require.Equal(t, map[string]any{"by": "ID", "input": "1001"}, body.Variables["siteRefInput"])
			payload["siteGeneralDetails"] = b.object
		case "siteCloudInterconnectPhysicalConnectionId":
			in := body.Variables["cloudInterconnectPhysicalConnectionIdInput"].(map[string]any)
			require.Equal(t, map[string]any{"by": "ID", "input": "1001"}, in["site"])
			if !b.roles[in["haRole"].(string)] {
				writeError(t, w, "connection not found")
				return
			}
			payload["cloudInterconnectPhysicalConnectionId"] = map[string]any{"id": "external-connection"}
		case "siteRemoveSite":
			require.Equal(t, "1001", body.Variables["siteId"])
			b.object = nil
			payload["removeSite"] = map[string]any{"siteId": "1001"}
		default:
			t.Fatalf("unexpected operation %s", body.OperationName)
		}
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"site": payload}}))
	}))
	t.Cleanup(server.Close)
	sdk, err := cato.New(server.URL, "test-token", "wrong-default", nil, nil)
	require.NoError(t, err)
	return &cloudSiteResource{client: &catoclient.Service{AccountId: "account-123", Catov2: catoclient.NewProviderSDKClient(sdk)}}
}
func writeError(t *testing.T, w http.ResponseWriter, message string) {
	t.Helper()
	require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"errors": []any{map[string]any{"message": message}}}))
}
func writeSnapshot(t *testing.T, w http.ResponseWriter, b *backend, vars map[string]any) {
	t.Helper()
	require.Equal(t, "account-123", vars["accountID"])
	require.Equal(t, []any{"1001"}, vars["siteIDs"])
	account, kind := "account-123", b.connection
	if account != "" && b.foreignAccount {
		account = "another"
	}
	if kind == "" {
		kind = "CROSS_CONNECT"
	}
	sites := []any{}
	if b.object != nil && !b.snapshotMissing {
		sites = append(sites, map[string]any{"id": "1001", "infoSiteSnapshot": map[string]any{"connType": kind}})
	}
	require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"accountSnapshot": map[string]any{"id": account, "sites": sites}}}))
}
func applyInput(b *backend, in map[string]any, create bool) {
	if name, ok := in["name"]; ok {
		b.object["site"].(map[string]any)["name"] = name
	}
	for _, key := range []string{"siteType", "description"} {
		if value, ok := in[key]; ok {
			b.object[key] = value
		}
	}
	if location, ok := in["siteLocation"]; ok {
		for key, value := range location.(map[string]any) {
			if key == "city" && create {
				key = "cityName"
			}
			b.object["siteLocation"].(map[string]any)[key] = value
		}
	}
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
	return resourceModel{ID: types.StringUnknown(), Name: types.StringValue("test"), Description: types.StringValue("test"), SiteType: types.StringValue("BRANCH"), SiteLocation: types.ObjectValueMust(locationAttrTypes, map[string]attr.Value{"country_code": types.StringValue("GB"), "timezone": types.StringValue("Europe/London"), "city": types.StringValue("London"), "state_code": types.StringValue(""), "address": types.StringValue("")})}
}
func TestCRUDAndLocationMapping(t *testing.T) {
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
	require.Equal(t, "London", b.inputs[0]["siteLocation"].(map[string]any)["city"])
	p = current
	p.Name = types.StringValue("updated")
	p.Description = types.StringNull()
	attrs := p.SiteLocation.Attributes()
	attrs["city"] = types.StringValue("Manchester")
	p.SiteLocation = types.ObjectValueMust(locationAttrTypes, attrs)
	update := resource.UpdateResponse{State: create.State}
	r.Update(ctx, resource.UpdateRequest{Plan: planFor(t, p), State: create.State}, &update)
	require.False(t, update.Diagnostics.HasError(), update.Diagnostics)
	require.Equal(t, "Manchester", b.inputs[1]["siteLocation"].(map[string]any)["cityName"])
	require.NotContains(t, b.inputs[1], "workingHours")
	require.NotContains(t, b.inputs[1], "preferredPopLocation")
	require.False(t, update.State.Get(ctx, &current).HasError())
	require.Equal(t, "", current.Description.ValueString())
	require.Equal(t, "updated", current.Name.ValueString())
	del := resource.DeleteResponse{State: update.State}
	r.Delete(ctx, resource.DeleteRequest{State: update.State}, &del)
	require.False(t, del.Diagnostics.HasError(), del.Diagnostics)
	require.True(t, del.State.Raw.IsNull())
}
func TestDeleteRefusesPhysicalConnections(t *testing.T) {
	for _, role := range []string{"PRIMARY", "SECONDARY"} {
		t.Run(role, func(t *testing.T) {
			b := &backend{object: remote(), roles: map[string]bool{role: true}}
			r := setup(t, b)
			p := model()
			p.ID = types.StringValue("1001")
			s := stateFor(t, p)
			resp := resource.DeleteResponse{State: s}
			r.Delete(context.Background(), resource.DeleteRequest{State: s}, &resp)
			require.True(t, resp.Diagnostics.HasError())
			require.NotContains(t, b.calls, "siteRemoveSite")
			require.False(t, resp.State.Raw.IsNull())
		})
	}
}
func TestReadMissingAndOwnership(t *testing.T) {
	for _, kind := range []string{"missing", "wrong_type", "wrong_account", "malformed", "denied"} {
		t.Run(kind, func(t *testing.T) {
			b := &backend{object: remote()}
			if kind == "missing" {
				b.object = nil
			}
			if kind == "wrong_type" {
				b.connection = "SOCKET_X1500"
			}
			if kind == "wrong_account" {
				b.foreignAccount = true
			}
			if kind == "malformed" {
				b.malformed = true
			}
			if kind == "denied" {
				b.fail = "accountSnapshot"
			}
			r := setup(t, b)
			p := model()
			p.ID = types.StringValue("1001")
			s := stateFor(t, p)
			resp := resource.ReadResponse{State: s}
			r.Read(context.Background(), resource.ReadRequest{State: s}, &resp)
			require.Equal(t, kind != "missing", resp.Diagnostics.HasError(), resp.Diagnostics)
			require.Equal(t, kind == "missing", resp.State.Raw.IsNull())
		})
	}
}
func TestCreateReadFailureKeepsIdentity(t *testing.T) {
	b := &backend{fail: "siteGeneralDetails"}
	r := setup(t, b)
	p := model()
	attrs := p.SiteLocation.Attributes()
	attrs["city"] = types.StringUnknown()
	p.SiteLocation = types.ObjectValueMust(locationAttrTypes, attrs)
	resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
	r.Create(context.Background(), resource.CreateRequest{Plan: planFor(t, p)}, &resp)
	require.True(t, resp.Diagnostics.HasError())
	var state resourceModel
	require.False(t, resp.State.Get(context.Background(), &state).HasError())
	require.Equal(t, "1001", state.ID.ValueString())
	require.False(t, state.SiteLocation.Attributes()["city"].IsUnknown())
}
func TestValidationBeforeAPI(t *testing.T) {
	for _, kind := range []string{"name", "type", "description", "null_location", "unknown_location", "country", "timezone"} {
		t.Run(kind, func(t *testing.T) {
			b := &backend{}
			r := setup(t, b)
			p := model()
			switch kind {
			case "name":
				p.Name = types.StringUnknown()
			case "type":
				p.SiteType = types.StringValue("INVALID")
			case "description":
				p.Description = types.StringUnknown()
			case "null_location":
				p.SiteLocation = types.ObjectNull(locationAttrTypes)
			case "unknown_location":
				p.SiteLocation = types.ObjectUnknown(locationAttrTypes)
			default:
				attrs := p.SiteLocation.Attributes()
				key := "country_code"
				if kind == "timezone" {
					key = "timezone"
				}
				attrs[key] = types.StringUnknown()
				p.SiteLocation = types.ObjectValueMust(locationAttrTypes, attrs)
			}
			resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
			r.Create(context.Background(), resource.CreateRequest{Plan: planFor(t, p)}, &resp)
			require.True(t, resp.Diagnostics.HasError())
			require.Empty(t, b.calls)
		})
	}
}
func TestPartialCreateKeepsID(t *testing.T) {
	ctx := context.Background()
	api := mocks.NewCloudInterconnectSiteClient(t)
	result := new(cato.SiteAddCloudInterconnectSite)
	require.NoError(t, json.Unmarshal([]byte(`{"site":{"addCloudInterconnectSite":{"siteId":"1001"}}}`), result))
	api.EXPECT().SiteAddCloudInterconnectSite(mock.Anything, mock.Anything, "account-123").Return(result, errors.New("partial failure")).Once()
	r := &cloudSiteResource{client: &catoclient.Service{AccountId: "account-123"}, apiClient: api}
	resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
	r.Create(ctx, resource.CreateRequest{Plan: planFor(t, model())}, &resp)
	require.True(t, resp.Diagnostics.HasError())
	var state resourceModel
	require.False(t, resp.State.Get(ctx, &state).HasError())
	require.Equal(t, "1001", state.ID.ValueString())
}
func TestImport(t *testing.T) {
	ctx := context.Background()
	p := model()
	p.ID = types.StringNull()
	resp := resource.ImportStateResponse{State: stateFor(t, p)}
	NewResource().(*cloudSiteResource).ImportState(ctx, resource.ImportStateRequest{ID: "1001"}, &resp)
	require.False(t, resp.Diagnostics.HasError())
	var id types.String
	require.False(t, resp.State.GetAttribute(ctx, path.Root("id"), &id).HasError())
	require.Equal(t, "1001", id.ValueString())
}

func TestReadSnapshotMissingPreservesExistingSite(t *testing.T) {
	b := &backend{object: remote(), snapshotMissing: true}
	r := setup(t, b)
	p := model()
	p.ID = types.StringValue("1001")
	s := stateFor(t, p)
	resp := resource.ReadResponse{State: s}
	r.Read(context.Background(), resource.ReadRequest{State: s}, &resp)
	require.True(t, resp.Diagnostics.HasError())
	require.False(t, resp.State.Raw.IsNull())
	require.NotContains(t, b.calls, "siteRemoveSite")
}

func TestMutationFailuresRetainState(t *testing.T) {
	for _, op := range []string{"siteAddCloudInterconnectSite", "siteUpdateSiteGeneralDetails", "siteCloudInterconnectPhysicalConnectionId", "siteRemoveSite"} {
		t.Run(op, func(t *testing.T) {
			b := &backend{object: remote(), fail: op}
			r := setup(t, b)
			p := model()
			p.ID = types.StringValue("1001")
			s := stateFor(t, p)
			ctx := context.Background()
			switch op {
			case "siteAddCloudInterconnectSite":
				resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
				r.Create(ctx, resource.CreateRequest{Plan: planFor(t, p)}, &resp)
				require.True(t, resp.Diagnostics.HasError())
			case "siteUpdateSiteGeneralDetails":
				resp := resource.UpdateResponse{State: s}
				r.Update(ctx, resource.UpdateRequest{Plan: planFor(t, p), State: s}, &resp)
				require.True(t, resp.Diagnostics.HasError())
				require.Equal(t, s.Raw, resp.State.Raw)
			default:
				resp := resource.DeleteResponse{State: s}
				r.Delete(ctx, resource.DeleteRequest{State: s}, &resp)
				require.True(t, resp.Diagnostics.HasError())
				require.False(t, resp.State.Raw.IsNull())
				if op == "siteCloudInterconnectPhysicalConnectionId" {
					require.NotContains(t, b.calls, "siteRemoveSite")
				}
			}
		})
	}
}
