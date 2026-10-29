package cloudinterconnectphysicalconnection

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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

func remote() map[string]any {
	return map[string]any{"id": "connection-1", "site": map[string]any{"id": "1001", "name": "Owned site"}, "haRole": "PRIMARY", "popLocation": map[string]any{"id": "10012", "name": "Mock PoP"}, "serviceProviderName": "Mock", "encapsulationMethod": "DOT1Q", "subnet": "192.0.2.0/30", "privateCatoIp": "192.0.2.2", "privateSiteIp": "192.0.2.1", "upstreamBwLimit": float64(100), "downstreamBwLimit": float64(100), "vlan": float64(123), "cVlan": float64(0), "sVlan": float64(0)}
}
func writeError(t *testing.T, w http.ResponseWriter, message string) {
	t.Helper()
	require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"errors": []any{map[string]any{"message": message}}}))
}
func setup(t *testing.T, b *backend) *connectionResource {
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
		payload := map[string]any{}
		switch body.OperationName {
		case "siteCloudInterconnectPhysicalConnectionId":
			in := body.Variables["cloudInterconnectPhysicalConnectionIdInput"].(map[string]any)
			require.Equal(t, map[string]any{"by": "ID", "input": "1001"}, in["site"])
			require.Equal(t, "PRIMARY", in["haRole"])
			if b.object == nil {
				writeError(t, w, "connection not found")
				return
			}
			payload["cloudInterconnectPhysicalConnectionId"] = map[string]any{"id": "connection-1"}
		case "siteAddCloudInterconnectPhysicalConnection":
			in := body.Variables["addCloudInterconnectPhysicalConnectionInput"].(map[string]any)
			b.inputs = append(b.inputs, in)
			require.Equal(t, map[string]any{"by": "ID", "input": "1001"}, in["site"])
			require.Equal(t, "PRIMARY", in["haRole"])
			b.object = remote()
			applyInput(t, b, in)
			payload["addCloudInterconnectPhysicalConnection"] = map[string]any{"id": "connection-1"}
		case "siteUpdateCloudInterconnectPhysicalConnection":
			in := body.Variables["updateCloudInterconnectPhysicalConnectionInput"].(map[string]any)
			b.inputs = append(b.inputs, in)
			require.Equal(t, "connection-1", in["id"])
			require.NotContains(t, in, "site")
			require.NotContains(t, in, "haRole")
			applyInput(t, b, in)
			payload["updateCloudInterconnectPhysicalConnection"] = map[string]any{"id": "connection-1"}
		case "siteCloudInterconnectPhysicalConnection":
			require.Equal(t, cato.SiteCloudInterconnectPhysicalConnectionDocument, body.Query)
			require.Equal(t, map[string]any{"id": "connection-1"}, body.Variables["cloudInterconnectPhysicalConnectionInput"])
			if b.object == nil {
				writeError(t, w, "connection not found")
				return
			}
			payload["cloudInterconnectPhysicalConnection"] = b.object
		case "siteRemoveCloudInterconnectPhysicalConnection":
			require.Equal(t, map[string]any{"id": "connection-1"}, body.Variables["removeCloudInterconnectPhysicalConnectionInput"])
			b.object = nil
			payload["removeCloudInterconnectPhysicalConnection"] = map[string]any{"id": "connection-1"}
		default:
			t.Fatalf("unexpected operation %s", body.OperationName)
		}
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"site": payload}}))
	}))
	t.Cleanup(server.Close)
	sdk, err := cato.New(server.URL, "test-token", "wrong-default", nil, nil)
	require.NoError(t, err)
	return &connectionResource{client: &catoclient.Service{AccountId: "account-123", Catov2: catoclient.NewProviderSDKClient(sdk)}}
}
func applyInput(t *testing.T, b *backend, in map[string]any) {
	t.Helper()
	for _, key := range []string{"encapsulationMethod", "privateCatoIp", "privateSiteIp", "serviceProviderName", "subnet", "vlan"} {
		if value, ok := in[key]; ok {
			b.object[key] = value
		}
	}
	for _, key := range []string{"upstreamBwLimit", "downstreamBwLimit"} {
		if value, ok := in[key]; ok {
			number, err := strconv.Atoi(value.(string))
			require.NoError(t, err)
			b.object[key] = number
		}
	}
	if in["encapsulationMethod"] == "QINQ" && in["QinQVlanConfiguration"] == nil {
		b.object["cVlan"] = float64(234)
		b.object["sVlan"] = float64(345)
	}
	if value, ok := in["QinQVlanConfiguration"]; ok {
		attrs := value.(map[string]any)
		b.object["cVlan"] = attrs["cVlan"]
		b.object["sVlan"] = attrs["sVlan"]
	}
	if name, ok := b.object["serviceProviderName"].(string); ok && strings.EqualFold(name, "Mock") {
		b.object["serviceProviderName"] = "Mock"
	}
	if b.object["encapsulationMethod"] == "DOT1Q" && b.object["vlan"] == float64(0) {
		b.object["vlan"] = float64(123)
	}
	if b.object["encapsulationMethod"] == "QINQ" && b.object["cVlan"] == float64(0) && b.object["sVlan"] == float64(0) {
		b.object["cVlan"] = float64(234)
		b.object["sVlan"] = float64(345)
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
func reference(id string) types.Object {
	return types.ObjectValueMust(idname.RefAttrTypes, map[string]attr.Value{"id": types.StringValue(id), "name": types.StringUnknown()})
}
func model() resourceModel {
	return resourceModel{ID: types.StringUnknown(), Site: reference("1001"), HARole: types.StringValue("PRIMARY"), PopLocation: reference("10012"), EncapsulationMethod: types.StringValue("DOT1Q"), ServiceProviderName: types.StringValue("Mock"), Subnet: types.StringValue("192.0.2.0/30"), PrivateCatoIP: types.StringValue("192.0.2.2"), PrivateSiteIP: types.StringValue("192.0.2.1"), UpstreamBwLimit: types.StringValue("100"), DownstreamBwLimit: types.StringValue("100"), Vlan: types.Int64Value(123), QinqVlanConfiguration: types.ObjectNull(qinqAttrTypes)}
}
func TestCRUDAndQinqHydration(t *testing.T) {
	ctx := context.Background()
	b := &backend{}
	r := setup(t, b)
	p := model()
	create := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
	r.Create(ctx, resource.CreateRequest{Config: configFor(t, p), Plan: planFor(t, p)}, &create)
	require.False(t, create.Diagnostics.HasError(), create.Diagnostics)
	var current resourceModel
	require.False(t, create.State.Get(ctx, &current).HasError())
	require.Equal(t, "connection-1", current.ID.ValueString())
	require.Equal(t, "100", current.UpstreamBwLimit.ValueString())
	require.Equal(t, int64(123), current.Vlan.ValueInt64())
	p = current
	p.Vlan = types.Int64Value(456)
	update := resource.UpdateResponse{State: create.State}
	r.Update(ctx, resource.UpdateRequest{Config: configFor(t, p), Plan: planFor(t, p), State: create.State}, &update)
	require.False(t, update.Diagnostics.HasError(), update.Diagnostics)
	require.False(t, update.State.Get(ctx, &current).HasError())
	require.Equal(t, int64(456), current.Vlan.ValueInt64())
	restore := update
	b.object["upstreamBwLimit"] = float64(250)
	read := resource.ReadResponse{State: restore.State}
	r.Read(ctx, resource.ReadRequest{State: restore.State}, &read)
	require.False(t, read.Diagnostics.HasError(), read.Diagnostics)
	require.False(t, read.State.Get(ctx, &current).HasError())
	require.Equal(t, "250", current.UpstreamBwLimit.ValueString())
	del := resource.DeleteResponse{State: read.State}
	r.Delete(ctx, resource.DeleteRequest{State: read.State}, &del)
	require.False(t, del.Diagnostics.HasError(), del.Diagnostics)
	require.True(t, del.State.Raw.IsNull())
}
func TestOccupiedSlotNotAdopted(t *testing.T) {
	b := &backend{object: remote()}
	r := setup(t, b)
	resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
	r.Create(context.Background(), resource.CreateRequest{Config: configFor(t, model()), Plan: planFor(t, model())}, &resp)
	require.True(t, resp.Diagnostics.HasError())
	require.NotContains(t, b.calls, "siteAddCloudInterconnectPhysicalConnection")
}
func TestValidationBeforeMutation(t *testing.T) {
	for _, kind := range []string{"role", "encapsulation", "subnet", "outside_ip", "equal_ips", "bandwidth", "vlan_range", "vlan_unknown", "qinq_unknown", "qinq_conflict", "ref_unknown", "ref_null_site", "ref_null_pop"} {
		t.Run(kind, func(t *testing.T) {
			b := &backend{}
			r := setup(t, b)
			p := model()
			switch kind {
			case "role":
				p.HARole = types.StringValue("OTHER")
			case "encapsulation":
				p.EncapsulationMethod = types.StringUnknown()
			case "subnet":
				p.Subnet = types.StringValue("192.0.2.1/30")
			case "outside_ip":
				p.PrivateSiteIP = types.StringValue("198.51.100.1")
			case "equal_ips":
				p.PrivateSiteIP = p.PrivateCatoIP
			case "bandwidth":
				p.UpstreamBwLimit = types.StringValue("00100")
			case "vlan_range":
				p.Vlan = types.Int64Value(4095)
			case "vlan_unknown":
				p.Vlan = types.Int64Unknown()
			case "qinq_unknown":
				p.EncapsulationMethod = types.StringValue("QINQ")
				p.Vlan = types.Int64Null()
				p.QinqVlanConfiguration = types.ObjectUnknown(qinqAttrTypes)
			case "qinq_conflict":
				p.QinqVlanConfiguration = types.ObjectValueMust(qinqAttrTypes, map[string]attr.Value{"c_vlan": types.Int64Value(1), "s_vlan": types.Int64Value(2)})
			case "ref_null_site":
				p.Site = types.ObjectNull(idname.RefAttrTypes)
			case "ref_null_pop":
				p.PopLocation = types.ObjectNull(idname.RefAttrTypes)
			case "ref_unknown":
				p.Site = types.ObjectUnknown(idname.RefAttrTypes)
			}
			resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
			r.Create(context.Background(), resource.CreateRequest{Config: configFor(t, p), Plan: planFor(t, p)}, &resp)
			require.True(t, resp.Diagnostics.HasError())
			require.Empty(t, b.calls)
		})
	}
}
func TestReadMissingAndErrors(t *testing.T) {
	for _, kind := range []string{"missing", "denied", "malformed", "wrong_id", "invalid_role"} {
		t.Run(kind, func(t *testing.T) {
			b := &backend{}
			if kind != "missing" {
				b.object = remote()
			}
			switch kind {
			case "denied":
				b.fail = "siteCloudInterconnectPhysicalConnection"
			case "malformed":
				b.malformed = true
			case "wrong_id":
				b.object["id"] = "another"
			case "invalid_role":
				b.object["haRole"] = "OTHER"
			}
			r := setup(t, b)
			p := model()
			p.ID = types.StringValue("connection-1")
			s := stateFor(t, p)
			resp := resource.ReadResponse{State: s}
			r.Read(context.Background(), resource.ReadRequest{State: s}, &resp)
			require.Equal(t, kind != "missing", resp.Diagnostics.HasError(), resp.Diagnostics)
			require.Equal(t, kind == "missing", resp.State.Raw.IsNull())
		})
	}
}
func TestImmutableRole(t *testing.T) {
	b := &backend{object: remote()}
	r := setup(t, b)
	p := model()
	p.ID = types.StringValue("connection-1")
	s := stateFor(t, p)
	p.HARole = types.StringValue("SECONDARY")
	resp := resource.UpdateResponse{State: s}
	r.Update(context.Background(), resource.UpdateRequest{Config: configFor(t, p), State: s, Plan: planFor(t, p)}, &resp)
	require.True(t, resp.Diagnostics.HasError())
	require.NotContains(t, b.calls, "siteUpdateCloudInterconnectPhysicalConnection")
}
func TestCreateReadFailureRetainsIdentity(t *testing.T) {
	b := &backend{fail: "siteCloudInterconnectPhysicalConnection"}
	r := setup(t, b)
	resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
	r.Create(context.Background(), resource.CreateRequest{Config: configFor(t, model()), Plan: planFor(t, model())}, &resp)
	require.True(t, resp.Diagnostics.HasError())
	var state resourceModel
	require.False(t, resp.State.Get(context.Background(), &state).HasError())
	require.Equal(t, "connection-1", state.ID.ValueString())
	require.False(t, state.Site.Attributes()["name"].IsUnknown())
	require.False(t, state.PopLocation.Attributes()["name"].IsUnknown())
}
func TestPartialCreateRetainsIdentity(t *testing.T) {
	ctx := context.Background()
	api := mocks.NewCloudInterconnectPhysicalConnectionClient(t)
	missing := gqlerror.List{&gqlerror.Error{Message: "connection not found"}}
	api.EXPECT().SiteCloudInterconnectPhysicalConnectionID(mock.Anything, "account-123", mock.Anything).Return(nil, &clientv2.ErrorResponse{GqlErrors: &missing}).Once()
	result := new(cato.SiteAddCloudInterconnectPhysicalConnection)
	require.NoError(t, json.Unmarshal([]byte(`{"site":{"addCloudInterconnectPhysicalConnection":{"id":"connection-1"}}}`), result))
	api.EXPECT().SiteAddCloudInterconnectPhysicalConnection(mock.Anything, mock.Anything, "account-123").Return(result, errors.New("partial failure")).Once()
	r := &connectionResource{client: &catoclient.Service{AccountId: "account-123"}, apiClient: api}
	resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
	r.Create(ctx, resource.CreateRequest{Config: configFor(t, model()), Plan: planFor(t, model())}, &resp)
	require.True(t, resp.Diagnostics.HasError())
	var state resourceModel
	require.False(t, resp.State.Get(ctx, &state).HasError())
	require.Equal(t, "connection-1", state.ID.ValueString())
}
func TestImport(t *testing.T) {
	ctx := context.Background()
	p := model()
	p.ID = types.StringNull()
	resp := resource.ImportStateResponse{State: stateFor(t, p)}
	NewResource().(*connectionResource).ImportState(ctx, resource.ImportStateRequest{ID: "connection-1"}, &resp)
	require.False(t, resp.Diagnostics.HasError())
	var id types.String
	require.False(t, resp.State.GetAttribute(ctx, path.Root("id"), &id).HasError())
	require.Equal(t, "connection-1", id.ValueString())
}

func TestBandwidthWireNormalization(t *testing.T) {
	for _, raw := range []string{`100`, `"100"`, `9223372036854775807`} {
		value, err := bandwidthJSON(json.RawMessage(raw))
		require.NoError(t, err)
		var text string
		require.NoError(t, json.Unmarshal(value, &text))
		require.NotEmpty(t, text)
	}
	for _, raw := range []string{`null`, `true`, `1.5`, `-1`, `"invalid"`, ``} {
		_, err := bandwidthJSON(json.RawMessage(raw))
		require.Error(t, err)
	}
}
func TestMutationErrorsKeepState(t *testing.T) {
	for _, op := range []string{"siteCloudInterconnectPhysicalConnectionId", "siteAddCloudInterconnectPhysicalConnection", "siteUpdateCloudInterconnectPhysicalConnection", "siteRemoveCloudInterconnectPhysicalConnection"} {
		t.Run(op, func(t *testing.T) {
			b := &backend{fail: op}
			p := model()
			p.ID = types.StringValue("connection-1")
			if op == "siteUpdateCloudInterconnectPhysicalConnection" || op == "siteRemoveCloudInterconnectPhysicalConnection" {
				b.object = remote()
			}
			r := setup(t, b)
			s := stateFor(t, p)
			ctx := context.Background()
			switch op {
			case "siteCloudInterconnectPhysicalConnectionId", "siteAddCloudInterconnectPhysicalConnection":
				resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
				r.Create(ctx, resource.CreateRequest{Config: configFor(t, p), Plan: planFor(t, p)}, &resp)
				require.True(t, resp.Diagnostics.HasError())
				if op == "siteCloudInterconnectPhysicalConnectionId" {
					require.NotContains(t, b.calls, "siteAddCloudInterconnectPhysicalConnection")
				}
			case "siteUpdateCloudInterconnectPhysicalConnection":
				resp := resource.UpdateResponse{State: s}
				r.Update(ctx, resource.UpdateRequest{Config: configFor(t, p), Plan: planFor(t, p), State: s}, &resp)
				require.True(t, resp.Diagnostics.HasError())
				require.Equal(t, s.Raw, resp.State.Raw)
			default:
				resp := resource.DeleteResponse{State: s}
				r.Delete(ctx, resource.DeleteRequest{State: s}, &resp)
				require.True(t, resp.Diagnostics.HasError())
				require.False(t, resp.State.Raw.IsNull())
			}
		})
	}
}

func TestAutomaticVLANsUseConfigurationOwnership(t *testing.T) {
	for _, mode := range []string{"DOT1Q", "QINQ"} {
		t.Run(mode, func(t *testing.T) {
			ctx := context.Background()
			b := &backend{}
			r := setup(t, b)
			cfg := model()
			cfg.EncapsulationMethod = types.StringValue(mode)
			cfg.Vlan = types.Int64Null()
			cfg.QinqVlanConfiguration = types.ObjectNull(qinqAttrTypes)
			plan := cfg
			plan.Vlan = types.Int64Unknown()
			plan.QinqVlanConfiguration = types.ObjectUnknown(qinqAttrTypes)
			create := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
			r.Create(ctx, resource.CreateRequest{Plan: planFor(t, plan), Config: configFor(t, cfg)}, &create)
			require.False(t, create.Diagnostics.HasError(), create.Diagnostics)
			require.NotContains(t, b.inputs[0], "vlan")
			require.NotContains(t, b.inputs[0], "QinQVlanConfiguration")
			var state resourceModel
			require.False(t, create.State.Get(ctx, &state).HasError())
			if mode == "DOT1Q" {
				require.Equal(t, int64(123), state.Vlan.ValueInt64())
				require.True(t, state.QinqVlanConfiguration.IsNull())
			} else {
				require.False(t, state.QinqVlanConfiguration.IsNull())
				require.True(t, state.Vlan.IsNull())
			}
			updatePlan := state
			updatePlan.UpstreamBwLimit = types.StringValue("250")
			cfg.UpstreamBwLimit = updatePlan.UpstreamBwLimit
			update := resource.UpdateResponse{State: create.State}
			r.Update(ctx, resource.UpdateRequest{Plan: planFor(t, updatePlan), State: create.State, Config: configFor(t, cfg)}, &update)
			require.False(t, update.Diagnostics.HasError(), update.Diagnostics)
			require.NotContains(t, b.inputs[1], "vlan")
			require.NotContains(t, b.inputs[1], "QinQVlanConfiguration")
			read := resource.ReadResponse{State: update.State}
			r.Read(ctx, resource.ReadRequest{State: update.State}, &read)
			require.False(t, read.Diagnostics.HasError(), read.Diagnostics)
			require.Equal(t, update.State.Raw, read.State.Raw)
		})
	}
}
func configFor(t *testing.T, p resourceModel) tfsdk.Config {
	t.Helper()
	plan := planFor(t, p)
	return tfsdk.Config(plan)
}

func TestAutomaticTagsReadFailureKeepsKnownIdentity(t *testing.T) {
	b := &backend{fail: "siteCloudInterconnectPhysicalConnection"}
	r := setup(t, b)
	cfg := model()
	cfg.Vlan = types.Int64Null()
	p := cfg
	p.Vlan = types.Int64Unknown()
	p.QinqVlanConfiguration = types.ObjectUnknown(qinqAttrTypes)
	resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
	r.Create(context.Background(), resource.CreateRequest{Config: configFor(t, cfg), Plan: planFor(t, p)}, &resp)
	require.True(t, resp.Diagnostics.HasError())
	var current resourceModel
	require.False(t, resp.State.Get(context.Background(), &current).HasError())
	require.Equal(t, "connection-1", current.ID.ValueString())
	require.True(t, current.Vlan.IsNull())
	require.True(t, current.QinqVlanConfiguration.IsNull())
}

func TestProviderCasePreserved(t *testing.T) {
	ctx := context.Background()
	b := &backend{}
	r := setup(t, b)
	p := model()
	p.ServiceProviderName = types.StringValue("mOcK")
	create := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
	r.Create(ctx, resource.CreateRequest{Config: configFor(t, p), Plan: planFor(t, p)}, &create)
	require.False(t, create.Diagnostics.HasError(), create.Diagnostics)
	var state resourceModel
	require.False(t, create.State.Get(ctx, &state).HasError())
	require.Equal(t, "mOcK", state.ServiceProviderName.ValueString())
	read := resource.ReadResponse{State: create.State}
	r.Read(ctx, resource.ReadRequest{State: create.State}, &read)
	require.False(t, read.Diagnostics.HasError())
	require.Equal(t, create.State.Raw, read.State.Raw)
	b.object["serviceProviderName"] = "Different"
	r.Read(ctx, resource.ReadRequest{State: create.State}, &read)
	require.False(t, read.Diagnostics.HasError())
	require.False(t, read.State.Get(ctx, &state).HasError())
	require.Equal(t, "Different", state.ServiceProviderName.ValueString())
}

func TestEncapsulationChangeFailsBeforeMutation(t *testing.T) {
	for _, mode := range []string{"DOT1Q", "QINQ"} {
		t.Run(mode, func(t *testing.T) {
			b := &backend{object: remote()}
			b.object["encapsulationMethod"] = mode
			r := setup(t, b)
			prior := model()
			prior.ID = types.StringValue("connection-1")
			state := stateFor(t, prior)
			p := model()
			if mode == "DOT1Q" {
				p.EncapsulationMethod = types.StringValue("QINQ")
				p.Vlan = types.Int64Null()
				p.QinqVlanConfiguration = types.ObjectValueMust(qinqAttrTypes, map[string]attr.Value{"c_vlan": types.Int64Value(234), "s_vlan": types.Int64Value(345)})
			}
			resp := resource.UpdateResponse{State: state}
			r.Update(context.Background(), resource.UpdateRequest{Config: configFor(t, p), Plan: planFor(t, p), State: state}, &resp)
			require.True(t, resp.Diagnostics.HasError())
			require.Equal(t, state.Raw, resp.State.Raw)
			require.NotContains(t, b.calls, "siteUpdateCloudInterconnectPhysicalConnection")
		})
	}
}
