package secondaryawsvsocket

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
	object                                                 map[string]any
	fail, missing, kind                                    string
	foreign, occupied, unknownRole, emptySnapshot, wrongID bool
	calls                                                  []string
	inputs                                                 []map[string]any
}

func remote() map[string]any {
	return map[string]any{"id": "2002", "site": map[string]any{"id": "1001", "name": "test"}, "ipAddress": "10.253.1.2", "subnet": "10.253.1.0/24", "routeTableId": "rtb-test"}
}

//nolint:gocyclo // The test HTTP fixture dispatches the six independent SDK operations.
func setup(t *testing.T, b *backend) *socketResource {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() { require.NoError(t, req.Body.Close()) }()
		var body struct {
			OperationName string         `json:"operationName"`
			Variables     map[string]any `json:"variables"`
		}
		require.NoError(t, json.NewDecoder(req.Body).Decode(&body))
		op, v := body.OperationName, body.Variables
		b.calls = append(b.calls, op)
		w.Header().Set("Content-Type", "application/json")
		write := func(data any) { require.NoError(t, json.NewEncoder(w).Encode(data)) }
		fail := func(message string) { write(map[string]any{"errors": []any{map[string]any{"message": message}}}) }
		if b.fail == op {
			fail("denied")
			return
		}
		if op == "accountSnapshot" {
			require.Equal(t, "account-123", v["accountID"])
			require.Equal(t, []any{"1001"}, v["siteIDs"])
			account := "account-123"
			if b.foreign {
				account = "foreign"
			}
			kind := b.kind
			if kind == "" {
				kind = "VSOCKET_VGX_AWS"
			}
			sockets := []any{map[string]any{"id": "2001", "isPrimary": true}}
			if b.occupied {
				sockets = append(sockets, map[string]any{"id": "2002", "isPrimary": false})
			}
			if b.unknownRole {
				sockets = []any{map[string]any{"id": "2001"}}
			}
			sites := []any{map[string]any{"id": "1001", "infoSiteSnapshot": map[string]any{"connType": kind, "sockets": sockets}}}
			if b.emptySnapshot {
				sites = []any{}
			}
			write(map[string]any{"data": map[string]any{"accountSnapshot": map[string]any{"id": account, "sites": sites}}})
			return
		}
		require.Equal(t, "account-123", v["accountId"])
		payload := map[string]any{}
		switch op {
		case "siteGeneralDetails":
			ref := v["siteRefInput"].(map[string]any)
			require.Contains(t, []any{"1001", "test"}, ref["input"])
			payload["siteGeneralDetails"] = map[string]any{"site": map[string]any{"id": "1001", "name": "test"}}
		case "siteAddSecondaryAwsVSocket":
			in := v["addSecondaryAwsVSocketInput"].(map[string]any)
			b.inputs = append(b.inputs, in)
			require.Equal(t, map[string]any{"by": "ID", "input": "1001"}, in["site"])
			b.object = remote()
			b.object["ipAddress"] = in["eniIpAddress"]
			b.object["subnet"] = in["eniIpSubnet"]
			b.object["routeTableId"] = in["routeTableId"]
			payload["addSecondaryAwsVSocket"] = map[string]any{"id": "2002"}
		case "siteSecondaryAwsVSocket":
			require.Equal(t, "2002", v["socketId"])
			if b.object == nil {
				message := b.missing
				if message == "" {
					message = "Socket ID 2002 was not found"
				}
				fail(message)
				return
			}
			payload["secondaryAwsVSocket"] = b.object
		case "siteUpdateSecondaryAwsVSocket":
			in := v["updateSecondaryAwsVSocketInput"].(map[string]any)
			b.inputs = append(b.inputs, in)
			require.Equal(t, "2002", in["id"])
			for _, key := range []string{"ipAddress", "subnet", "routeTableId"} {
				b.object[key] = in[key]
			}
			payload["updateSecondaryAwsVSocket"] = map[string]any{"secondaryAwsVSocket": b.object}
		case "siteRemoveSecondaryAwsVSocket":
			require.Equal(t, "2002", v["socketId"])
			payload["removeSecondaryAwsVSocket"] = map[string]any{"secondaryAwsVSocket": b.object}
			b.object = nil
		default:
			t.Fatalf("unexpected operation %s", op)
		}
		if b.wrongID {
			if obj, ok := payload["secondaryAwsVSocket"].(map[string]any); ok {
				obj["id"] = "foreign"
			}
		}
		write(map[string]any{"data": map[string]any{"site": payload}})
	}))
	t.Cleanup(server.Close)
	sdk, err := cato.New(server.URL, "test-token", "wrong-default", nil, nil)
	require.NoError(t, err)
	return &socketResource{client: &catoclient.Service{AccountId: "account-123", Catov2: catoclient.NewProviderSDKClient(sdk)}}
}
func schemaFor(t *testing.T) schema.Schema {
	t.Helper()
	var r resource.SchemaResponse
	NewResource().Schema(context.Background(), resource.SchemaRequest{}, &r)
	return r.Schema
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
	return resourceModel{ID: types.StringUnknown(), Site: types.ObjectValueMust(idname.RefAttrTypes, map[string]attr.Value{"id": types.StringValue("1001"), "name": types.StringUnknown()}), EniIPAddress: types.StringValue("10.253.1.2"), EniIPSubnet: types.StringValue("10.253.1.0/24"), RouteTableID: types.StringValue("rtb-test")}
}
func TestCRUD(t *testing.T) {
	ctx := context.Background()
	b := &backend{}
	r := setup(t, b)
	p := model()
	create := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
	r.Create(ctx, resource.CreateRequest{Plan: planFor(t, p)}, &create)
	require.False(t, create.Diagnostics.HasError(), create.Diagnostics)
	var state resourceModel
	require.False(t, create.State.Get(ctx, &state).HasError())
	require.Equal(t, "2002", state.ID.ValueString())
	require.Equal(t, "test", state.Site.Attributes()["name"].(types.String).ValueString())
	p = state
	p.EniIPAddress = types.StringValue("10.253.2.3")
	p.EniIPSubnet = types.StringValue("10.253.2.0/24")
	p.RouteTableID = types.StringValue("rtb-updated")
	update := resource.UpdateResponse{State: create.State}
	r.Update(ctx, resource.UpdateRequest{Plan: planFor(t, p), State: create.State}, &update)
	require.False(t, update.Diagnostics.HasError(), update.Diagnostics)
	require.Equal(t, "10.253.2.3", b.inputs[1]["ipAddress"])
	require.NotContains(t, b.inputs[1], "eniIpAddress")
	read := resource.ReadResponse{State: update.State}
	r.Read(ctx, resource.ReadRequest{State: update.State}, &read)
	require.False(t, read.Diagnostics.HasError(), read.Diagnostics)
	require.Equal(t, update.State.Raw, read.State.Raw)
	del := resource.DeleteResponse{State: read.State}
	r.Delete(ctx, resource.DeleteRequest{State: read.State}, &del)
	require.False(t, del.Diagnostics.HasError(), del.Diagnostics)
	require.True(t, del.State.Raw.IsNull())
}
func TestCreateGuard(t *testing.T) {
	for _, kind := range []string{"occupied", "foreign", "unknown_role", "missing_snapshot", "wrong_cloud"} {
		t.Run(kind, func(t *testing.T) {
			b := &backend{}
			switch kind {
			case "occupied":
				b.occupied = true
			case "foreign":
				b.foreign = true
			case "unknown_role":
				b.unknownRole = true
			case "missing_snapshot":
				b.emptySnapshot = true
			default:
				b.kind = "SOCKET_X1500"
			}
			r := setup(t, b)
			resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
			r.Create(context.Background(), resource.CreateRequest{Plan: planFor(t, model())}, &resp)
			require.True(t, resp.Diagnostics.HasError(), resp.Diagnostics)
			require.NotContains(t, b.calls, "siteAddSecondaryAwsVSocket")
		})
	}
}
func TestValidationBeforeAPI(t *testing.T) {
	for _, kind := range []string{"null_ip", "unknown_ip", "bad_ip", "bad_subnet", "outside", "route", "null_site", "unknown_site", "empty_site"} {
		t.Run(kind, func(t *testing.T) {
			p := model()
			switch kind {
			case "null_ip":
				p.EniIPAddress = types.StringNull()
			case "unknown_ip":
				p.EniIPAddress = types.StringUnknown()
			case "bad_ip":
				p.EniIPAddress = types.StringValue("::1")
			case "bad_subnet":
				p.EniIPSubnet = types.StringValue("10.253.1.2/24")
			case "outside":
				p.EniIPAddress = types.StringValue("10.254.1.2")
			case "route":
				p.RouteTableID = types.StringUnknown()
			case "null_site":
				p.Site = types.ObjectNull(idname.RefAttrTypes)
			case "unknown_site":
				p.Site = types.ObjectUnknown(idname.RefAttrTypes)
			default:
				p.Site = types.ObjectValueMust(idname.RefAttrTypes, map[string]attr.Value{"id": types.StringNull(), "name": types.StringNull()})
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
func TestMissingAndFailures(t *testing.T) {
	for _, kind := range []string{"missing", "denied", "primary", "wrong_cloud", "wrong_id"} {
		t.Run(kind, func(t *testing.T) {
			b := &backend{}
			switch kind {
			case "denied":
				b.fail = "siteSecondaryAwsVSocket"
			case "primary":
				b.missing = "Socket ID 2002 is not a secondary socket"
			case "wrong_cloud":
				b.missing = "Socket is not of type VSOCKET_VGX_AWS"
			case "wrong_id":
				b.object = remote()
				b.wrongID = true
			}
			r := setup(t, b)
			p := model()
			p.ID = types.StringValue("2002")
			s := stateFor(t, p)
			resp := resource.ReadResponse{State: s}
			r.Read(context.Background(), resource.ReadRequest{State: s}, &resp)
			require.Equal(t, kind != "missing", resp.Diagnostics.HasError(), resp.Diagnostics)
			require.Equal(t, kind == "missing", resp.State.Raw.IsNull())
			del := resource.DeleteResponse{State: s}
			r.Delete(context.Background(), resource.DeleteRequest{State: s}, &del)
			require.Equal(t, kind != "missing", del.Diagnostics.HasError())
			require.NotContains(t, b.calls, "siteRemoveSecondaryAwsVSocket")
		})
	}
}
func TestMutationFailuresAndIdentity(t *testing.T) {
	for _, op := range []string{"siteGeneralDetails", "accountSnapshot", "siteAddSecondaryAwsVSocket", "siteSecondaryAwsVSocket", "siteUpdateSecondaryAwsVSocket", "siteRemoveSecondaryAwsVSocket"} {
		t.Run(op, func(t *testing.T) {
			b := &backend{object: remote(), fail: op}
			r := setup(t, b)
			p := model()
			p.ID = types.StringValue("2002")
			s := stateFor(t, p)
			ctx := context.Background()
			switch op {
			case "siteUpdateSecondaryAwsVSocket":
				resp := resource.UpdateResponse{State: s}
				r.Update(ctx, resource.UpdateRequest{Plan: planFor(t, p), State: s}, &resp)
				require.True(t, resp.Diagnostics.HasError())
				require.Equal(t, s.Raw, resp.State.Raw)
			case "siteRemoveSecondaryAwsVSocket":
				resp := resource.DeleteResponse{State: s}
				r.Delete(ctx, resource.DeleteRequest{State: s}, &resp)
				require.True(t, resp.Diagnostics.HasError())
				require.Equal(t, s.Raw, resp.State.Raw)
			default:
				resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
				r.Create(ctx, resource.CreateRequest{Plan: planFor(t, p)}, &resp)
				require.True(t, resp.Diagnostics.HasError())
				if op == "siteSecondaryAwsVSocket" {
					var state resourceModel
					require.False(t, resp.State.Get(ctx, &state).HasError())
					require.Equal(t, "2002", state.ID.ValueString())
					require.False(t, state.Site.Attributes()["name"].IsUnknown())
				}
			}
		})
	}
}
func TestImmutableParent(t *testing.T) {
	b := &backend{object: remote()}
	r := setup(t, b)
	p := model()
	p.ID = types.StringValue("2002")
	s := stateFor(t, p)
	attrs := p.Site.Attributes()
	attrs["id"] = types.StringValue("another")
	p.Site = types.ObjectValueMust(idname.RefAttrTypes, attrs)
	resp := resource.UpdateResponse{State: s}
	r.Update(context.Background(), resource.UpdateRequest{Plan: planFor(t, p), State: s}, &resp)
	require.True(t, resp.Diagnostics.HasError())
	require.NotContains(t, b.calls, "siteUpdateSecondaryAwsVSocket")
}
func TestImport(t *testing.T) {
	ctx := context.Background()
	p := model()
	p.ID = types.StringNull()
	resp := resource.ImportStateResponse{State: stateFor(t, p)}
	NewResource().(*socketResource).ImportState(ctx, resource.ImportStateRequest{ID: "2002"}, &resp)
	require.False(t, resp.Diagnostics.HasError())
	var id types.String
	require.False(t, resp.State.GetAttribute(ctx, path.Root("id"), &id).HasError())
	require.Equal(t, "2002", id.ValueString())
}

func TestPartialCreateKeepsID(t *testing.T) {
	ctx := context.Background()
	api := mocks.NewSecondaryAWSVSocketClient(t)
	details := new(cato.SiteGeneralDetails)
	require.NoError(t, json.Unmarshal([]byte(`{"site":{"siteGeneralDetails":{"site":{"id":"1001","name":"test"}}}}`), details))
	snapshot := new(cato.AccountSnapshot)
	require.NoError(t, json.Unmarshal([]byte(`{"accountSnapshot":{"id":"account-123","sites":[{"id":"1001","infoSiteSnapshot":{"connType":"VSOCKET_VGX_AWS","sockets":[{"id":"2001","isPrimary":true}]}}]}}`), snapshot))
	result := new(cato.SiteAddSecondaryAWSVSocket)
	require.NoError(t, json.Unmarshal([]byte(`{"site":{"addSecondaryAwsVSocket":{"id":"2002"}}}`), result))
	api.EXPECT().SiteGeneralDetails(mock.Anything, mock.Anything, "account-123").Return(details, nil).Once()
	api.EXPECT().AccountSnapshot(mock.Anything, []string{"1001"}, []string{}, mock.Anything).Return(snapshot, nil).Once()
	api.EXPECT().SiteAddSecondaryAWSVSocket(mock.Anything, "account-123", mock.Anything).Return(result, errors.New("partial failure")).Once()
	r := &socketResource{client: &catoclient.Service{AccountId: "account-123"}, apiClient: api}
	resp := resource.CreateResponse{State: tfsdk.State{Schema: schemaFor(t)}}
	r.Create(ctx, resource.CreateRequest{Plan: planFor(t, model())}, &resp)
	require.True(t, resp.Diagnostics.HasError())
	var state resourceModel
	require.False(t, resp.State.Get(ctx, &state).HasError())
	require.Equal(t, "2002", state.ID.ValueString())
	require.False(t, state.Site.Attributes()["name"].IsUnknown())
}
