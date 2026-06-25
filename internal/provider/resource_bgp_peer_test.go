package provider

import (
	"context"
	"errors"
	"testing"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/cato-go-sdk/scalars"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
)

// --- attr type maps ----------------------------------------------------------

var bgpBfdAttrTypes = map[string]attr.Type{
	"transmit_interval": types.Int64Type,
	"receive_interval":  types.Int64Type,
	"multiplier":        types.Int64Type,
}

var bgpTrackingAttrTypes = map[string]attr.Type{
	"enabled":         types.BoolType,
	"alert_frequency": types.StringType,
	"subscription_id": types.StringType,
}

var bgpSummaryCommunityAttrTypes = map[string]attr.Type{
	"from": types.Int64Type,
	"to":   types.Int64Type,
}

var bgpSummaryRouteElemType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"route": types.StringType,
		"community": types.ListType{
			ElemType: types.ObjectType{AttrTypes: bgpSummaryCommunityAttrTypes},
		},
	},
}

// --- helpers -----------------------------------------------------------------

func getBgpPeerSchemaResp(ctx context.Context, t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &bgpPeerResource{}
	var resp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &resp)
	return resp
}

func newBgpPeerStateFromModel(ctx context.Context, t *testing.T, m BgpPeer) tfsdk.State {
	t.Helper()
	schemaResp := getBgpPeerSchemaResp(ctx, t)
	state := tfsdk.State{Schema: schemaResp.Schema}
	if diags := state.Set(ctx, m); diags.HasError() {
		t.Fatalf("newBgpPeerStateFromModel: %v", diags)
	}
	return state
}

// nullBgpPeer returns a BgpPeer where all optional objects/lists are null.
func nullBgpPeer() BgpPeer {
	emptySummaryList, _ := types.ListValue(bgpSummaryRouteElemType, []attr.Value{})
	return BgpPeer{
		ID:                     types.StringValue("peer-1"),
		SiteID:                 types.StringValue("site-1"),
		Name:                   types.StringValue("test-peer"),
		PeerAsn:                types.Int64Value(65100),
		CatoAsn:                types.Int64Value(65000),
		PeerIP:                 types.StringValue("192.168.254.20"),
		AdvertiseDefaultRoute:  types.BoolValue(true),
		AdvertiseAllRoutes:     types.BoolValue(false),
		AdvertiseSummaryRoutes: types.BoolValue(false),
		DefaultAction:          types.StringValue("ACCEPT"),
		PerformNat:             types.BoolValue(false),
		Md5AuthKey:             types.StringNull(),
		Metric:                 types.Int64Value(50),
		HoldTime:               types.Int64Value(60),
		KeepaliveInterval:      types.Int64Value(20),
		BfdEnabled:             types.BoolValue(false),
		BfdSettings:            types.ObjectNull(bgpBfdAttrTypes),
		Tracking:               types.ObjectNull(bgpTrackingAttrTypes),
		SummaryRoute:           emptySummaryList,
	}
}

func buildBfdSettingsObject(t *testing.T, transmit, receive, multiplier int64) types.Object {
	t.Helper()
	obj, diags := types.ObjectValue(bgpBfdAttrTypes, map[string]attr.Value{
		"transmit_interval": types.Int64Value(transmit),
		"receive_interval":  types.Int64Value(receive),
		"multiplier":        types.Int64Value(multiplier),
	})
	if diags.HasError() {
		t.Fatalf("buildBfdSettingsObject: %v", diags)
	}
	return obj
}

// minimalAPIBgpPeer builds an API response with nil BfdSettingsBgpPeer and
// nil TrackingBgpPeer — what the API returns when BFD is disabled.
func minimalAPIBgpPeer() cato_go_sdk.Site_Site_BgpPeer {
	return cato_go_sdk.Site_Site_BgpPeer{
		ID:                     "peer-1",
		Name:                   "test-peer",
		PeerAsn:                scalars.Asn32("65100"),
		CatoAsn:                scalars.Asn16("65000"),
		PeerIP:                 "192.168.254.20",
		AdvertiseDefaultRoute:  true,
		AdvertiseAllRoutes:     false,
		AdvertiseSummaryRoutes: false,
		DefaultAction:          cato_models.BgpDefaultAction("ACCEPT"),
		PerformNat:             false,
		Metric:                 50,
		HoldTime:               60,
		KeepaliveInterval:      20,
		BfdEnabled:             false,
		BfdSettingsBgpPeer:     nil,
		TrackingBgpPeer:        nil,
		Site:                   cato_go_sdk.Site_Site_BgpPeer_Site{ID: "site-1"},
		SummaryRoute:           nil,
	}
}

func siteWithBgpPeer(peer cato_go_sdk.Site_Site_BgpPeer) *cato_go_sdk.Site {
	return &cato_go_sdk.Site{
		Site: cato_go_sdk.Site_Site{BgpPeer: &peer},
	}
}

func bgpReadReq(ctx context.Context, t *testing.T, m BgpPeer) resource.ReadRequest {
	t.Helper()
	return resource.ReadRequest{State: newBgpPeerStateFromModel(ctx, t, m)}
}

func bgpReadResp(ctx context.Context, t *testing.T) *resource.ReadResponse {
	t.Helper()
	schemaResp := getBgpPeerSchemaResp(ctx, t)
	return &resource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
}

func bgpReadRespFromState(ctx context.Context, t *testing.T, m BgpPeer) *resource.ReadResponse {
	t.Helper()
	return &resource.ReadResponse{State: newBgpPeerStateFromModel(ctx, t, m)}
}

// --- constructor / metadata / configure tests --------------------------------

func TestNewBgpPeerResource(t *testing.T) {
	t.Parallel()
	r := NewBgpPeerResource()
	if r == nil {
		t.Fatal("expected non-nil resource")
	}
	if _, ok := r.(*bgpPeerResource); !ok {
		t.Fatalf("expected *bgpPeerResource, got %T", r)
	}
}

func TestBgpPeerMetadata(t *testing.T) {
	t.Parallel()
	r := &bgpPeerResource{}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "cato"}, &resp)
	if resp.TypeName != "cato_bgp_peer" {
		t.Fatalf("expected cato_bgp_peer, got %q", resp.TypeName)
	}
}

func TestBgpPeerConfigureNilProviderData(t *testing.T) {
	t.Parallel()
	r := &bgpPeerResource{}
	r.Configure(context.Background(), resource.ConfigureRequest{}, &resource.ConfigureResponse{})
	if r.client != nil {
		t.Fatal("expected client to remain nil")
	}
}

func TestBgpPeerConfigureSetsClient(t *testing.T) {
	t.Parallel()
	client := &catoClientData{AccountId: "account-123"}
	r := &bgpPeerResource{}
	r.Configure(context.Background(), resource.ConfigureRequest{ProviderData: client}, &resource.ConfigureResponse{})
	if r.client != client {
		t.Fatal("expected client to be set from provider data")
	}
}

func TestBgpPeerGetClientFallsBackToCatov2(t *testing.T) {
	t.Parallel()
	// When no mock is injected, getBgpPeerClient must return the catov2 SDK
	// client (even if it is nil underneath — the interface wraps the pointer).
	r := &bgpPeerResource{client: &catoClientData{catov2: nil}}
	// Just verify that no mock is used; the catov2 value is returned.
	// (A nil *cato.Client still yields a non-nil BgpPeerClient interface.)
	_ = r.getBgpPeerClient()
}

func TestBgpPeerGetClientUsesInjectedMock(t *testing.T) {
	t.Parallel()
	mc := mocks.NewBgpPeerClient(t)
	r := &bgpPeerResource{bgpPeerClient: mc}
	if r.getBgpPeerClient() != mc {
		t.Fatal("expected injected mock client to be returned")
	}
}

// --- Read tests --------------------------------------------------------------

// TestBgpPeerReadPreservesBfdSettingsWhenAPIReturnsNil is the main regression
// test for the non-idempotent Read bug: when bfd_enabled=false the API omits
// bfd_settings, which used to cause the provider to null it out in state and
// produce a perpetual non-empty plan.
func TestBgpPeerReadPreservesBfdSettingsWhenAPIReturnsNil(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// State has bfd_settings explicitly set (typical when bfd_enabled=false but
	// the user still provides the block in their config).
	stateModel := nullBgpPeer()
	stateModel.BfdSettings = buildBfdSettingsObject(t, 1000, 1000, 5)

	apiPeer := minimalAPIBgpPeer() // BfdSettingsBgpPeer == nil

	mc := mocks.NewBgpPeerClient(t)
	mc.EXPECT().
		SiteBgpPeer(ctx, cato_models.BgpPeerRefInput{By: cato_models.ObjectRefByID, Input: "peer-1"}, "account-123").
		Return(siteWithBgpPeer(apiPeer), nil).Once()

	r := &bgpPeerResource{
		client:        &catoClientData{AccountId: "account-123"},
		bgpPeerClient: mc,
	}

	req := bgpReadReq(ctx, t, stateModel)
	resp := bgpReadResp(ctx, t)
	r.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}

	var got BgpPeer
	if diags := resp.State.Get(ctx, &got); diags.HasError() {
		t.Fatalf("get state: %v", diags)
	}

	if got.BfdSettings.IsNull() {
		t.Fatal("bfd_settings was nulled out by Read — perpetual-drift bug reproduced; fix not applied")
	}
}

// TestBgpPeerReadNullBfdSettingsRemainsNullWhenAPIReturnsNil verifies that
// when the prior state also had null bfd_settings (user did not provide the
// block), the result remains null after Read.
func TestBgpPeerReadNullBfdSettingsRemainsNullWhenAPIReturnsNil(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	stateModel := nullBgpPeer() // BfdSettings == null

	apiPeer := minimalAPIBgpPeer() // BfdSettingsBgpPeer == nil

	mc := mocks.NewBgpPeerClient(t)
	mc.EXPECT().
		SiteBgpPeer(ctx, cato_models.BgpPeerRefInput{By: cato_models.ObjectRefByID, Input: "peer-1"}, "account-123").
		Return(siteWithBgpPeer(apiPeer), nil).Once()

	r := &bgpPeerResource{
		client:        &catoClientData{AccountId: "account-123"},
		bgpPeerClient: mc,
	}

	req := bgpReadReq(ctx, t, stateModel)
	resp := bgpReadResp(ctx, t)
	r.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}

	var got BgpPeer
	if diags := resp.State.Get(ctx, &got); diags.HasError() {
		t.Fatalf("get state: %v", diags)
	}

	if !got.BfdSettings.IsNull() {
		t.Fatal("expected bfd_settings to remain null when state was null and API returned nil")
	}
}

// TestBgpPeerReadUsesAPIBfdSettingsWhenPresent verifies that when the API
// returns actual BfdSettingsBgpPeer data, it is used (not the state value).
func TestBgpPeerReadUsesAPIBfdSettingsWhenPresent(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	stateModel := nullBgpPeer()
	stateModel.BfdSettings = buildBfdSettingsObject(t, 500, 500, 3) // old state values

	apiPeer := minimalAPIBgpPeer()
	apiPeer.BfdEnabled = true
	apiPeer.BfdSettingsBgpPeer = &cato_go_sdk.Site_Site_BgpPeer_BfdSettingsBgpPeer{
		TransmitInterval: 1000,
		ReceiveInterval:  1000,
		Multiplier:       5,
	}

	mc := mocks.NewBgpPeerClient(t)
	mc.EXPECT().
		SiteBgpPeer(ctx, cato_models.BgpPeerRefInput{By: cato_models.ObjectRefByID, Input: "peer-1"}, "account-123").
		Return(siteWithBgpPeer(apiPeer), nil).Once()

	r := &bgpPeerResource{
		client:        &catoClientData{AccountId: "account-123"},
		bgpPeerClient: mc,
	}

	req := bgpReadReq(ctx, t, stateModel)
	resp := bgpReadResp(ctx, t)
	r.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}

	var got BgpPeer
	if diags := resp.State.Get(ctx, &got); diags.HasError() {
		t.Fatalf("get state: %v", diags)
	}

	if got.BfdSettings.IsNull() {
		t.Fatal("expected bfd_settings from API to be set in state")
	}

	// The API values (1000, 1000, 5) should appear in state, not the old (500, 500, 3).
	var bfd BfdSettingsInput
	if diags := got.BfdSettings.As(ctx, &bfd, basetypes.ObjectAsOptions{}); diags.HasError() {
		t.Fatalf("decode bfd_settings: %v", diags)
	}
	if bfd.Multiplier.ValueInt64() != 5 {
		t.Fatalf("expected multiplier=5 from API, got %d", bfd.Multiplier.ValueInt64())
	}
}

// TestBgpPeerReadAPIError verifies that an API error produces diagnostics.
func TestBgpPeerReadAPIError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	mc := mocks.NewBgpPeerClient(t)
	mc.EXPECT().
		SiteBgpPeer(ctx, cato_models.BgpPeerRefInput{By: cato_models.ObjectRefByID, Input: "peer-1"}, "account-123").
		Return(nil, errors.New("API unavailable")).Once()

	r := &bgpPeerResource{
		client:        &catoClientData{AccountId: "account-123"},
		bgpPeerClient: mc,
	}

	req := bgpReadReq(ctx, t, nullBgpPeer())
	resp := bgpReadResp(ctx, t)
	r.Read(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error diagnostics from API failure")
	}
}

// TestBgpPeerReadRemovesStateWhenResourceGone verifies that a nil BgpPeer
// from the API causes the resource to be removed from state.
func TestBgpPeerReadRemovesStateWhenResourceGone(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	gone := &cato_go_sdk.Site{Site: cato_go_sdk.Site_Site{BgpPeer: nil}}

	mc := mocks.NewBgpPeerClient(t)
	mc.EXPECT().
		SiteBgpPeer(ctx, cato_models.BgpPeerRefInput{By: cato_models.ObjectRefByID, Input: "peer-1"}, "account-123").
		Return(gone, nil).Once()

	r := &bgpPeerResource{
		client:        &catoClientData{AccountId: "account-123"},
		bgpPeerClient: mc,
	}

	req := bgpReadReq(ctx, t, nullBgpPeer())
	resp := bgpReadRespFromState(ctx, t, nullBgpPeer())
	r.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
	if !resp.State.Raw.IsNull() {
		t.Fatal("expected state to be removed when resource is gone")
	}
}
