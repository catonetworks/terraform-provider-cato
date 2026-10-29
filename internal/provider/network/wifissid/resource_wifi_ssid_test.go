package wifissid

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	m "github.com/catonetworks/cato-go-sdk/models"
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
	rawOccupied, rawMalformed, paginate bool
	failMessage                         string
	object                              map[string]any
	fail, missing                       string
	occupied, wrongID                   bool
	calls                               []string
	inputs                              []map[string]any
}

func remote() map[string]any {
	return map[string]any{"id": "ssid-1", "site": map[string]any{"id": "1001", "name": "test"}, "networkInterfaceId": "WBR1", "name": "test", "band": "BOTH", "category": "GUEST", "enabled": false, "visible": true, "subnet": "10.252.10.0/24", "localIp": "10.252.10.1", "dhcpEnabled": false, "internetOnly": false, "microsegmentationEnabled": false, "mdnsEnabled": false, "dhcp": map[string]any{"dhcpSubnet": nil}, "security": map[string]any{"mode": "PSK", "authProtocol": "WPA3", "trackAuthentication": false}}
}

//nolint:gocyclo // HTTP fixture dispatches independent SDK operations and their failure cases.
func setup(t *testing.T, b *backend) *ssidResource {
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
		b.calls = append(b.calls, op)
		if op == "entityLookup" {
			require.Equal(t, "account-123", v["accountID"])
		} else {
			require.Equal(t, "account-123", v["accountId"])
		}
		w.Header().Set("Content-Type", "application/json")
		write := func(value any) { require.NoError(t, json.NewEncoder(w).Encode(value)) }
		fail := func(message string) { write(map[string]any{"errors": []any{map[string]any{"message": message}}}) }
		if op == b.fail {
			message := b.failMessage
			if message == "" {
				message = "denied"
			}
			fail(message)
			return
		}

		if op == "entityLookup" {
			require.Equal(t, map[string]any{"id": "1001", "type": "site"}, v["parent"])
			require.Equal(t, []any{"siteId", "interfaceId"}, v["helperFields"])
			require.Equal(t, float64(200), v["limit"])
			items := []any{}
			total := 0
			if b.rawOccupied || b.rawMalformed || b.paginate {
				port, owner := "WBR1", any("1001")
				total = 1
				if b.rawMalformed {
					owner = nil
				}
				if b.paginate {
					total = 2
					if v["from"] == float64(0) {
						port = "INT_4"
					}
				}
				items = append(items, map[string]any{"entity": map[string]any{"id": "n-1", "type": "networkInterface"}, "helperFields": map[string]any{"siteId": owner, "interfaceId": port}})
			}
			write(map[string]any{"data": map[string]any{"entityLookup": map[string]any{"items": items, "total": total}}})
			return
		}
		payload := map[string]any{}
		switch op {
		case "siteGeneralDetails":
			require.Equal(t, map[string]any{"by": "ID", "input": "1001"}, v["siteRefInput"])
			payload["siteGeneralDetails"] = map[string]any{"site": map[string]any{"id": "1001", "name": "test"}}
		case "siteWifiSsidList":
			require.Equal(t, map[string]any{"site": map[string]any{"by": "ID", "input": "1001"}}, v["wifiSsidListInput"])
			items := []any{}
			if b.occupied {
				items = append(items, remote())
			}
			payload["wifiSsidList"] = map[string]any{"wifiSsid": items, "pageInfo": map[string]any{"total": len(items)}}
		case "siteCreateWifiSsid":
			in := v["createWifiSsidInput"].(map[string]any)
			b.inputs = append(b.inputs, in)
			b.object = remote()
			applyInput(b, in)
			require.NotContains(t, in, "vlanId")
			require.NotContains(t, in, "dhcpEnabled")
			payload["createWifiSsid"] = map[string]any{"wifiSsid": b.object}
		case "siteWifiSsid":
			require.Equal(t, map[string]any{"by": "ID", "input": "ssid-1"}, v["wifiSsidRefInput"])
			if b.object == nil || b.missing != "" {
				message := b.missing
				if message == "" {
					message = "Wifi SSID ssid-1 not found"
				}
				fail(message)
				return
			}
			payload["wifiSsid"] = b.object
		case "siteUpdateWifiSsid":
			in := v["updateWifiSsidInput"].(map[string]any)
			require.Equal(t, "ssid-1", in["id"])
			b.inputs = append(b.inputs, in)
			applyInput(b, in)
			require.NotContains(t, in, "site")
			require.NotContains(t, in, "networkInterfaceId")
			require.NotContains(t, in, "dhcpEnabled")
			payload["updateWifiSsid"] = map[string]any{"wifiSsid": b.object}
		case "siteRemoveWifiSsid":
			require.NotContains(t, body.Query, "localIp")
			require.NotContains(t, body.Query, "subnet")
			require.Contains(t, body.Query, "networkInterfaceId")
			require.Equal(t, map[string]any{"wifiSsid": map[string]any{"by": "ID", "input": "ssid-1"}}, v["removeWifiSsidInput"])
			payload["removeWifiSsid"] = map[string]any{"wifiSsid": b.object}
			b.object = nil
		default:
			t.Fatalf("unexpected operation %s", op)
		}
		if b.wrongID && b.object != nil {
			b.object["id"] = "foreign"
		}
		write(map[string]any{"data": map[string]any{"site": payload}})
	}))
	t.Cleanup(server.Close)
	sdk, err := cato.New(server.URL, "test-token", "wrong-default", nil, nil)
	require.NoError(t, err)
	return &ssidResource{client: &catoclient.Service{AccountId: "account-123", Catov2: catoclient.NewProviderSDKClient(sdk)}}
}
func applyInput(b *backend, in map[string]any) {
	for _, key := range []string{"name", "band", "category", "enabled", "visible", "subnet", "localIp", "internetOnly", "mdnsEnabled", "microsegmentationEnabled", "dhcp"} {
		if value, ok := in[key]; ok {
			b.object[key] = value
		}
	}
	if value, ok := in["security"].(map[string]any); ok {
		security := map[string]any{}
		for _, key := range []string{"mode", "authProtocol", "trackAuthentication"} {
			security[key] = value[key]
		}
		b.object["security"] = security
	}
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
func model() resourceModel {
	return resourceModel{
		ID: types.StringUnknown(), Site: types.ObjectValueMust(idname.RefAttrTypes, map[string]attr.Value{"id": types.StringValue("1001"), "name": types.StringUnknown()}), NetworkInterfaceID: types.StringValue("WBR1"), Name: types.StringValue("test"), Band: types.StringValue("BOTH"), Category: types.StringValue("GUEST"), Enabled: types.BoolValue(false), Visible: types.BoolValue(true), Subnet: types.StringValue("10.252.10.0/24"), LocalIP: types.StringValue("10.252.10.1"), DhcpEnabled: types.BoolValue(false), Dhcp: types.ObjectNull(dhcpAttrTypes), MdnsEnabled: types.BoolValue(false), MicrosegmentationEnabled: types.BoolValue(false), InternetOnly: types.BoolValue(false), Security: types.ObjectValueMust(securityAttrTypes, map[string]attr.Value{"mode": types.StringValue("PSK"), "auth_protocol": types.StringValue("WPA3"), "track_authentication": types.BoolValue(false), "psk": types.ObjectValueMust(pskAttrTypes, map[string]attr.Value{"passkey": types.StringValue(strings.Repeat("x", 12))})}),
	}
}
func TestCRUDSecretRetentionAndDHCPReset(t *testing.T) {
	ctx := context.Background()
	b := &backend{}
	r := setup(t, b)
	p := model()
	create := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
	r.Create(ctx, resource.CreateRequest{Plan: planFor(t, p)}, &create)
	require.False(t, create.Diagnostics.HasError(), create.Diagnostics)
	var state resourceModel
	require.False(t, create.State.Get(ctx, &state).HasError())
	require.Equal(t, "ssid-1", state.ID.ValueString())
	require.Equal(t, secretValue(p), secretValue(state))
	state.Name = types.StringValue("updated")
	state.Dhcp = types.ObjectValueMust(dhcpAttrTypes, map[string]attr.Value{"dhcp_subnet": types.StringValue("10.252.10.128/25")})
	update := resource.UpdateResponse{State: create.State}
	r.Update(ctx, resource.UpdateRequest{State: create.State, Plan: planFor(t, state)}, &update)
	require.False(t, update.Diagnostics.HasError(), update.Diagnostics)
	read := resource.ReadResponse{State: update.State}
	r.Read(ctx, resource.ReadRequest{State: update.State}, &read)
	require.False(t, read.Diagnostics.HasError(), read.Diagnostics)
	require.Equal(t, update.State.Raw, read.State.Raw)
	require.False(t, read.State.Get(ctx, &state).HasError())
	state.Dhcp = types.ObjectNull(dhcpAttrTypes)
	reset := resource.UpdateResponse{State: read.State}
	r.Update(ctx, resource.UpdateRequest{State: read.State, Plan: planFor(t, state)}, &reset)
	require.False(t, reset.Diagnostics.HasError(), reset.Diagnostics)
	require.Equal(t, map[string]any{}, b.inputs[2]["dhcp"])
	require.False(t, reset.State.Get(ctx, &state).HasError())
	require.True(t, state.Dhcp.IsNull())
	require.Equal(t, secretValue(p), secretValue(state))
	del := resource.DeleteResponse{State: reset.State}
	r.Delete(ctx, resource.DeleteRequest{State: reset.State}, &del)
	require.False(t, del.Diagnostics.HasError(), del.Diagnostics)
	require.True(t, del.State.Raw.IsNull())
}

//nolint:gocyclo // Independent validation cases share a table and a no-API assertion.
func TestValidationBeforeAPI(t *testing.T) {
	for _, kind := range []string{"unknown_name", "bad_name", "unknown_band", "corporate", "unknown_bool", "unknown_site", "null_security", "unknown_security", "missing_psk", "bad_psk", "open_psk", "dhcp_enabled", "unknown_dhcp", "bad_subnet", "outside_ip", "bad_interface"} {
		t.Run(kind, func(t *testing.T) {
			p := model()
			attrs := p.Security.Attributes()
			switch kind {
			case "unknown_name":
				p.Name = types.StringUnknown()
			case "bad_name":
				p.Name = types.StringValue("invalid/")
			case "unknown_band":
				p.Band = types.StringUnknown()
			case "corporate":
				p.Category = types.StringValue("CORPORATE")
			case "unknown_bool":
				p.Enabled = types.BoolUnknown()
			case "unknown_site":
				p.Site = types.ObjectUnknown(idname.RefAttrTypes)
			case "null_security":
				p.Security = types.ObjectNull(securityAttrTypes)
			case "unknown_security":
				p.Security = types.ObjectUnknown(securityAttrTypes)
			case "dhcp_enabled":
				p.DhcpEnabled = types.BoolValue(true)
			case "unknown_dhcp":
				p.Dhcp = types.ObjectUnknown(dhcpAttrTypes)
			case "bad_subnet":
				p.Subnet = types.StringValue("10.252.10.1/24")
			case "outside_ip":
				p.LocalIP = types.StringValue("10.253.1.1")
			case "bad_interface":
				p.NetworkInterfaceID = types.StringValue("INT_1")
			default:
				switch kind {
				case "missing_psk":
					attrs["psk"] = types.ObjectNull(pskAttrTypes)
				case "bad_psk":
					attrs["psk"] = types.ObjectValueMust(pskAttrTypes, map[string]attr.Value{"passkey": types.StringUnknown()})
				default:
					attrs["mode"] = types.StringValue("OPEN")
				}
				p.Security = types.ObjectValueMust(securityAttrTypes, attrs)
			}
			b := &backend{}
			r := setup(t, b)
			resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
			r.Create(context.Background(), resource.CreateRequest{Plan: planFor(t, p)}, &resp)
			require.True(t, resp.Diagnostics.HasError())
			require.Empty(t, b.calls)
		})
	}
}
func TestOccupiedAndFailures(t *testing.T) {
	for _, op := range []string{"occupied", "siteWifiSsidList", "siteCreateWifiSsid", "siteUpdateWifiSsid", "siteRemoveWifiSsid"} {
		t.Run(op, func(t *testing.T) {
			b := &backend{object: remote(), fail: op, occupied: op == "occupied"}
			r := setup(t, b)
			p := model()
			p.ID = types.StringValue("ssid-1")
			s := stateFor(t, p)
			switch op {
			case "siteUpdateWifiSsid":
				resp := resource.UpdateResponse{State: s}
				r.Update(context.Background(), resource.UpdateRequest{State: s, Plan: planFor(t, p)}, &resp)
				require.True(t, resp.Diagnostics.HasError())
				require.Equal(t, s.Raw, resp.State.Raw)
			case "siteRemoveWifiSsid":
				resp := resource.DeleteResponse{State: s}
				r.Delete(context.Background(), resource.DeleteRequest{State: s}, &resp)
				require.True(t, resp.Diagnostics.HasError())
				require.Equal(t, s.Raw, resp.State.Raw)
			default:
				resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
				r.Create(context.Background(), resource.CreateRequest{Plan: planFor(t, p)}, &resp)
				require.True(t, resp.Diagnostics.HasError())
				if op != "siteCreateWifiSsid" {
					require.NotContains(t, b.calls, "siteCreateWifiSsid")
				}
			}
		})
	}
}
func TestMissingAndImportWithoutSecret(t *testing.T) {
	for _, kind := range []string{"missing", "denied", "wrong_id", "import"} {
		t.Run(kind, func(t *testing.T) {
			b := &backend{object: remote()}
			switch kind {
			case "missing":
				b.object = nil
			case "denied":
				b.fail = "siteWifiSsid"
			case "wrong_id":
				b.wrongID = true
			}
			r := setup(t, b)
			p := model()
			p.ID = types.StringValue("ssid-1")
			if kind == "import" {
				p.Security = types.ObjectNull(securityAttrTypes)
			}
			s := stateFor(t, p)
			resp := resource.ReadResponse{State: s}
			r.Read(context.Background(), resource.ReadRequest{State: s}, &resp)
			require.Equal(t, kind == "denied" || kind == "wrong_id", resp.Diagnostics.HasError(), resp.Diagnostics)
			require.Equal(t, kind == "missing", resp.State.Raw.IsNull())
			if kind == "import" {
				var state resourceModel
				require.False(t, resp.State.Get(context.Background(), &state).HasError())
				require.Empty(t, secretValue(state))
			}
		})
	}
	ctx := context.Background()
	p := model()
	p.ID = types.StringNull()
	resp := resource.ImportStateResponse{State: stateFor(t, p)}
	NewResource().(*ssidResource).ImportState(ctx, resource.ImportStateRequest{ID: "ssid-1"}, &resp)
	require.False(t, resp.Diagnostics.HasError())
	var id types.String
	require.False(t, resp.State.GetAttribute(ctx, path.Root("id"), &id).HasError())
	require.Equal(t, "ssid-1", id.ValueString())
}
func TestImmutableParentAndInterface(t *testing.T) {
	for _, key := range []string{"site", "interface"} {
		t.Run(key, func(t *testing.T) {
			b := &backend{object: remote()}
			r := setup(t, b)
			p := model()
			p.ID = types.StringValue("ssid-1")
			s := stateFor(t, p)
			if key == "site" {
				p.Site = types.ObjectValueMust(idname.RefAttrTypes, map[string]attr.Value{"id": types.StringValue("other"), "name": types.StringNull()})
			} else {
				p.NetworkInterfaceID = types.StringValue("WBR2")
			}
			resp := resource.UpdateResponse{State: s}
			r.Update(context.Background(), resource.UpdateRequest{Plan: planFor(t, p), State: s}, &resp)
			require.True(t, resp.Diagnostics.HasError())
			require.NotContains(t, b.calls, "siteUpdateWifiSsid")
		})
	}
}

func TestPartialCreateAndErrorRedaction(t *testing.T) {
	ctx := context.Background()
	api := mocks.NewWifiSSIDClient(t)
	listing := new(cato.SiteWifiSsidList)
	require.NoError(t, json.Unmarshal([]byte(`{"site":{"wifiSsidList":{"wifiSsid":[],"pageInfo":{"total":0}}}}`), listing))
	partial := new(cato.SiteCreateWifiSsid)
	require.NoError(t, json.Unmarshal([]byte(`{"site":{"createWifiSsid":{"wifiSsid":{"id":"ssid-1"}}}}`), partial))
	api.EXPECT().SiteWifiSsidList(mock.Anything, "account-123", mock.Anything).Return(listing, nil).Once()

	details := new(cato.SiteGeneralDetails)
	require.NoError(t, json.Unmarshal([]byte(`{"site":{"siteGeneralDetails":{"site":{"id":"1001","name":"test"}}}}`), details))
	lookup := new(cato.EntityLookup)
	require.NoError(t, json.Unmarshal([]byte(`{"entityLookup":{"items":[],"total":0}}`), lookup))
	api.EXPECT().SiteGeneralDetails(mock.Anything, mock.Anything, "account-123").Return(details, nil).Once()
	api.EXPECT().EntityLookup(mock.Anything, "account-123", m.EntityTypeNetworkInterface, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, []string{"siteId", "interfaceId"}).Return(lookup, nil).Once()
	api.EXPECT().SiteCreateWifiSsid(mock.Anything, "account-123", mock.Anything).Return(partial, errors.New("requestBody contains "+secretValue(model()))).Once()
	r := &ssidResource{client: &catoclient.Service{AccountId: "account-123"}, apiClient: api}
	resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
	r.Create(ctx, resource.CreateRequest{Plan: planFor(t, model())}, &resp)
	require.True(t, resp.Diagnostics.HasError())
	var state resourceModel
	require.False(t, resp.State.Get(ctx, &state).HasError())
	require.Equal(t, "ssid-1", state.ID.ValueString())
	require.False(t, state.Site.Attributes()["name"].IsUnknown())
	for _, d := range resp.Diagnostics {
		require.NotContains(t, d.Detail(), secretValue(model()))
		require.NotContains(t, d.Detail(), "requestBody")
	}
	b := &backend{fail: "siteCreateWifiSsid", failMessage: "invalid passkey " + secretValue(model())}
	r = setup(t, b)
	redacted := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
	r.Create(ctx, resource.CreateRequest{Plan: planFor(t, model())}, &redacted)
	require.True(t, redacted.Diagnostics.HasError())
	require.Contains(t, redacted.Diagnostics[0].Detail(), "omitted")
	require.NotContains(t, redacted.Diagnostics[0].Detail(), secretValue(model()))
	require.NotContains(t, redacted.Diagnostics[0].Detail(), "requestBody")
}
func TestMissingDeleteAndOpenTransition(t *testing.T) {
	ctx := context.Background()
	b := &backend{}
	r := setup(t, b)
	p := model()
	p.ID = types.StringValue("ssid-1")
	s := stateFor(t, p)
	del := resource.DeleteResponse{State: s}
	r.Delete(ctx, resource.DeleteRequest{State: s}, &del)
	require.False(t, del.Diagnostics.HasError())
	require.True(t, del.State.Raw.IsNull())
	require.NotContains(t, b.calls, "siteRemoveWifiSsid")
	b.object = remote()
	attrs := p.Security.Attributes()
	attrs["mode"] = types.StringValue("OPEN")
	attrs["psk"] = types.ObjectNull(pskAttrTypes)
	p.Security = types.ObjectValueMust(securityAttrTypes, attrs)
	update := resource.UpdateResponse{State: s}
	r.Update(ctx, resource.UpdateRequest{State: s, Plan: planFor(t, p)}, &update)
	require.False(t, update.Diagnostics.HasError(), update.Diagnostics)
	var state resourceModel
	require.False(t, update.State.Get(ctx, &state).HasError())
	require.Empty(t, secretValue(state))
	require.NotContains(t, b.inputs[0]["security"], "psk")
}

func TestAuthoritativeInterfaceGuard(t *testing.T) {
	for _, kind := range []string{"legacy_configured", "incomplete_ownership", "pagination", "lookup_denied", "site_denied"} {
		t.Run(kind, func(t *testing.T) {
			b := &backend{}
			switch kind {
			case "legacy_configured":
				b.rawOccupied = true
			case "incomplete_ownership":
				b.rawMalformed = true
			case "pagination":
				b.paginate = true
			case "lookup_denied":
				b.fail = "entityLookup"
			default:
				b.fail = "siteGeneralDetails"
			}
			r := setup(t, b)
			resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
			r.Create(context.Background(), resource.CreateRequest{Plan: planFor(t, model())}, &resp)
			require.True(t, resp.Diagnostics.HasError())
			require.NotContains(t, b.calls, "siteCreateWifiSsid")
		})
	}
}
func TestStaleCC2Recovery(t *testing.T) {
	for _, kind := range []string{"stale", "wrong_id"} {
		t.Run(kind, func(t *testing.T) {
			message := "Wifi SSID 'test' (ID ssid-1) not found"
			if kind == "wrong_id" {
				message = "Wifi SSID 'test' (ID other) not found"
			}
			b := &backend{object: remote(), missing: message}
			r := setup(t, b)
			p := model()
			p.ID = types.StringValue("ssid-1")
			s := stateFor(t, p)
			read := resource.ReadResponse{State: s}
			r.Read(context.Background(), resource.ReadRequest{State: s}, &read)
			require.Equal(t, kind != "stale", read.Diagnostics.HasError())
			require.Equal(t, kind == "stale", read.State.Raw.IsNull())
			del := resource.DeleteResponse{State: s}
			r.Delete(context.Background(), resource.DeleteRequest{State: s}, &del)
			require.Equal(t, kind != "stale", del.Diagnostics.HasError(), del.Diagnostics)
			require.Equal(t, kind == "stale", del.State.Raw.IsNull())
			if kind == "stale" {
				require.Contains(t, b.calls, "siteRemoveWifiSsid")
			} else {
				require.NotContains(t, b.calls, "siteRemoveWifiSsid")
			}
		})
	}
}
func TestUnsupportedEnterprise(t *testing.T) {
	p := model()
	attrs := p.Security.Attributes()
	attrs["mode"] = types.StringValue("ENTERPRISE")
	p.Security = types.ObjectValueMust(securityAttrTypes, attrs)
	b := &backend{}
	r := setup(t, b)
	resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
	r.Create(context.Background(), resource.CreateRequest{Plan: planFor(t, p)}, &resp)
	require.True(t, resp.Diagnostics.HasError())
	require.Empty(t, b.calls)
}
