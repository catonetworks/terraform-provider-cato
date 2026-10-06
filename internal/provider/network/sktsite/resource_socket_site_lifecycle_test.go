package sktsite

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	cato "github.com/catonetworks/cato-go-sdk"
	models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/common/client"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/sktsite/adapter"
)

func socketLifecycleModel(ctx context.Context, t *testing.T) SocketSite {
	t.Helper()
	model := *newSocketSitePlanWithTranslatedSubnet(ctx, t, types.StringNull())
	model.ID = types.StringValue("site")
	model.Description = types.StringNull()
	model.Sockets = types.SetNull(types.ObjectType{AttrTypes: SocketTypes})
	return model
}
func expectSocketRead(t *testing.T, api *mocks.SocketSiteClient, absent bool) {
	t.Helper()
	var general cato.SiteGeneralDetails
	if !absent {
		require.NoError(t, json.Unmarshal([]byte(`{"site":{"siteGeneralDetails":{"site":{"id":"site","name":"aws-site-01"},"siteType":"DATACENTER","siteLocation":{"countryCode":"US","timezone":"America/New_York"}}}}`), &general))
	}
	api.EXPECT().SiteGeneralDetails(mock.Anything, models.SiteRefInput{By: models.ObjectRefByID, Input: "site"}, "account").Return(&general, nil).Once()
	if absent {
		return
	}
	api.EXPECT().SiteSocketConfiguration(mock.Anything, mock.Anything, "account").Return(&cato.SiteSocketConfiguration{Site: cato.SiteSocketConfiguration_Site{SiteSocketConfiguration: socketConfigurationForTest(models.SocketModelAWS)}}, nil).Once()
	var ranges cato.NetworkRangeList
	require.NoError(t, json.Unmarshal([]byte(`{"site":{"networkRangeList":{"items":[{"networkRangeId":"range","name":"Native","rangeType":"Native","subnet":"10.51.0.128/25","localIP":"10.51.0.1"}]}}}`), &ranges))
	api.EXPECT().NetworkRangeList(mock.Anything, "account", mock.Anything).Return(&ranges, nil).Once()
	api.EXPECT().EntityLookup(mock.Anything, "account", models.EntityTypeNetworkInterface, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&cato.EntityLookup{EntityLookup: cato.EntityLookup_EntityLookup{Items: []*cato.EntityLookup_EntityLookup_Items{{Entity: cato.EntityLookup_EntityLookup_Items_Entity{ID: "interface"}, HelperFields: map[string]any{"isDefault": true, "interfaceId": "LAN1", "interfaceName": "LAN", "destType": "LAN"}}}}}, nil).Once()
}
func TestSocketSiteReadAndImportLifecycle(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	for _, operation := range []string{"read", "import", "read_absent"} {
		t.Run(operation, func(t *testing.T) {
			t.Parallel()
			api := mocks.NewSocketSiteClient(t)
			r := &socketSiteResource{client: &client.CatoClientData{AccountId: "account"}, socketSiteClient: api}
			state := tfsdk.State{Schema: r.resourceSchema()}
			require.False(t, state.Set(ctx, socketLifecycleModel(ctx, t)).HasError())
			expectSocketRead(t, api, operation == "read_absent")
			if operation == "import" {
				resp := &resource.ImportStateResponse{State: state}
				r.ImportState(ctx, resource.ImportStateRequest{ID: "site"}, resp)
				require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
				state = resp.State
			} else {
				resp := &resource.ReadResponse{State: state}
				r.Read(ctx, resource.ReadRequest{State: state}, resp)
				require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
				state = resp.State
			}
			if operation == "read_absent" {
				require.True(t, state.Raw.IsNull())
				return
			}
			var got SocketSite
			require.False(t, state.Get(ctx, &got).HasError())
			require.Equal(t, "site", got.ID.ValueString())
			require.Equal(t, "SOCKET_AWS1500", got.ConnectionType.ValueString())
		})
	}
}
func TestSocketSiteCreatePendingState(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	api := mocks.NewSocketSiteClient(t)
	r := &socketSiteResource{client: &client.CatoClientData{AccountId: "account"}, socketSiteClient: api, retry: adapter.Retry{Wait: func(context.Context, time.Duration) error { return nil }}}
	model := socketLifecycleModel(ctx, t)
	model.ID = types.StringUnknown()
	plan := tfsdk.Plan{Schema: r.resourceSchema()}
	require.False(t, plan.Set(ctx, model).HasError())
	cfg := tfsdk.Config{Schema: r.resourceSchema(), Raw: plan.Raw}
	api.EXPECT().SiteAddSocketSite(mock.Anything, mock.Anything, "account").Return(&cato.SiteAddSocketSite{Site: cato.SiteAddSocketSite_Site{AddSocketSite: &cato.SiteAddSocketSite_Site_AddSocketSite{SiteID: "site"}}}, nil).Once()
	api.EXPECT().NetworkRangeList(mock.Anything, "account", mock.Anything).Return(&cato.NetworkRangeList{Site: cato.NetworkRangeList_Site{NetworkRangeList: &cato.NetworkRangeList_Site_NetworkRangeList{Items: []*cato.NetworkRangeList_Site_NetworkRangeList_Items{{NetworkRangeID: "range", RangeType: models.SubnetTypeNative}}}}}, nil).Once()
	api.EXPECT().SiteUpdateNetworkRange(mock.Anything, "range", mock.Anything, "account").Return(nil, nil).Once()
	api.EXPECT().SiteUpdateSocketInterface(mock.Anything, "site", models.SocketInterfaceIDEnumLan1, mock.Anything, "account").Return(nil, nil).Once()
	api.EXPECT().SiteGeneralDetails(mock.Anything, mock.Anything, "account").Return(nil, nil).Times(6)
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: r.resourceSchema()}}
	r.Create(ctx, resource.CreateRequest{Plan: plan, Config: cfg}, resp)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	var got SocketSite
	require.False(t, resp.State.Get(ctx, &got).HasError())
	require.Equal(t, "site", got.ID.ValueString())
	require.Equal(t, model.Name, got.Name)
}

func TestSocketSiteCreateAndUpdateLifecycle(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	for _, operation := range []string{"create", "update"} {
		t.Run(operation, func(t *testing.T) {
			t.Parallel()
			api := mocks.NewSocketSiteClient(t)
			r := &socketSiteResource{client: &client.CatoClientData{AccountId: "account"}, socketSiteClient: api}
			model := socketLifecycleModel(ctx, t)
			if operation == "create" {
				api.EXPECT().SiteAddSocketSite(mock.Anything, mock.Anything, "account").Return(&cato.SiteAddSocketSite{Site: cato.SiteAddSocketSite_Site{AddSocketSite: &cato.SiteAddSocketSite_Site_AddSocketSite{SiteID: "site"}}}, nil).Once()
				api.EXPECT().NetworkRangeList(mock.Anything, "account", mock.Anything).Return(&cato.NetworkRangeList{Site: cato.NetworkRangeList_Site{NetworkRangeList: &cato.NetworkRangeList_Site_NetworkRangeList{Items: []*cato.NetworkRangeList_Site_NetworkRangeList_Items{{NetworkRangeID: "range", RangeType: models.SubnetTypeNative}}}}}, nil).Once()
			} else {
				var native NativeRange
				require.False(t, model.NativeRange.As(ctx, &native, basetypes.ObjectAsOptions{}).HasError())
				native.NativeNetworkRangeID = types.StringValue("range")
				native.InterfaceIndex = types.StringValue("LAN1")
				obj, ds := types.ObjectValueFrom(ctx, SiteNativeRangeResourceAttrTypes, native)
				require.False(t, ds.HasError())
				model.NativeRange = obj
				api.EXPECT().SiteUpdateSiteGeneralDetails(mock.Anything, "site", mock.Anything, "account").Return(nil, nil).Once()
			}
			api.EXPECT().SiteUpdateNetworkRange(mock.Anything, "range", mock.Anything, "account").Return(nil, nil).Once()
			api.EXPECT().SiteUpdateSocketInterface(mock.Anything, "site", models.SocketInterfaceIDEnumLan1, mock.Anything, "account").Return(nil, nil).Once()
			expectSocketRead(t, api, false)
			plan := tfsdk.Plan{Schema: r.resourceSchema()}
			require.False(t, plan.Set(ctx, model).HasError())
			cfg := tfsdk.Config{Schema: r.resourceSchema(), Raw: plan.Raw}
			state := tfsdk.State{Schema: r.resourceSchema()}
			require.False(t, state.Set(ctx, model).HasError())
			if operation == "create" {
				resp := &resource.CreateResponse{State: state}
				r.Create(ctx, resource.CreateRequest{Plan: plan, Config: cfg}, resp)
				require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
				state = resp.State
			} else {
				resp := &resource.UpdateResponse{State: state}
				r.Update(ctx, resource.UpdateRequest{Plan: plan, Config: cfg, State: state}, resp)
				require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
				state = resp.State
			}
			var got SocketSite
			require.False(t, state.Get(ctx, &got).HasError())
			require.Equal(t, "site", got.ID.ValueString())
			require.Equal(t, "SOCKET_AWS1500", got.ConnectionType.ValueString())
		})
	}
}
