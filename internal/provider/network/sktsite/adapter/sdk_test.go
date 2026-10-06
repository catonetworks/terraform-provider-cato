package adapter

import (
	"context"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/sktsite/application"
)

//nolint:gocyclo // Payload matchers assert independent forwarded fields.
func TestSocketMutationPayloads(t *testing.T) {
	t.Parallel()
	api := mocks.NewSocketSiteClient(t)
	a := SDK{Client: api, AccountID: "account"}
	ctx := context.Background()
	api.EXPECT().SiteAddSocketSite(mock.Anything, mock.MatchedBy(func(v models.AddSocketSiteInput) bool {
		return v.ConnectionType == models.SiteConnectionTypeEnumSocketAWS1500 && v.Name == "site" && v.Description != nil && *v.Description == "" && v.TranslatedSubnet == nil && v.SiteLocation != nil && v.SiteLocation.CountryCode == "US" && v.Vlan != nil && int64(*v.Vlan) == 20
	}), "account").Return(&cato.SiteAddSocketSite{Site: cato.SiteAddSocketSite_Site{AddSocketSite: &cato.SiteAddSocketSite_Site_AddSocketSite{SiteID: "site"}}}, nil).Once()
	id, err := a.Add(ctx, application.AddInput{Name: "site", ConnectionType: "SOCKET_AWS1500", Description: new(""), SiteLocation: &application.Location{CountryCode: "US", Timezone: "America/New_York"}, Vlan: new(int64(20))})
	require.NoError(t, err)
	require.Equal(t, "site", id)
	api.EXPECT().SiteUpdateNetworkRange(mock.Anything, "range", mock.MatchedBy(func(v models.UpdateNetworkRangeInput) bool {
		return v.Subnet != nil && *v.Subnet == "10.0.0.0/24" && v.LocalIP == nil && v.TranslatedSubnet == nil
	}), "account").Return(nil, nil).Once()
	require.NoError(t, a.UpdateRange(ctx, "range", application.RangeInput{Subnet: new("10.0.0.0/24")}))
	api.EXPECT().SiteUpdateSocketInterface(mock.Anything, "site", models.SocketInterfaceIDEnumInt5, mock.MatchedBy(func(v models.UpdateSocketInterfaceInput) bool {
		return v.DestType == models.SocketInterfaceDestTypeLanLagMaster && v.Lag != nil && v.Lag.MinLinks == 2 && v.Lan == nil
	}), "account").Return(nil, nil).Once()
	require.NoError(t, a.UpdateInterface(ctx, "site", "INT_5", application.InterfaceInput{DestType: "LAN_LAG_MASTER", Lag: &application.LagInput{MinLinks: 2}}))
	api.EXPECT().SiteExchangeSocketPorts(mock.Anything, "account", mock.MatchedBy(func(v models.ExchangeSocketPortsInput) bool {
		return v.Site != nil && v.Site.By == models.ObjectRefByID && v.Site.Input == "site" && v.FirstInterface.InterfaceID == models.SocketInterfaceIDEnumInt3 && v.SecondInterface.InterfaceID == models.SocketInterfaceIDEnumInt5
	})).Return(nil, nil).Once()
	require.NoError(t, a.Exchange(ctx, "site", "INT_3", "INT_5"))
}
func TestDefaultInterfaceStateFallback(t *testing.T) {
	t.Parallel()
	api := mocks.NewSocketSiteClient(t)
	a := SDK{Client: api, AccountID: "account"}
	api.EXPECT().EntityLookup(mock.Anything, "account", models.EntityTypeNetworkInterface, new(int64(0)), (*int64)(nil), &models.EntityInput{Type: models.EntityTypeSite, ID: "site"}, (*string)(nil), []string(nil), []*models.SortInput(nil), []*models.LookupFilterInput(nil), []string(nil)).Return(&cato.EntityLookup{EntityLookup: cato.EntityLookup_EntityLookup{Items: []*cato.EntityLookup_EntityLookup_Items{{Entity: cato.EntityLookup_EntityLookup_Items_Entity{ID: "interface"}, HelperFields: map[string]any{"interfaceId": "5", "interfaceName": "LAN", "destType": "LAN"}}}}}, nil).Once()
	got, err := a.DefaultInterface(context.Background(), "site", "interface")
	require.NoError(t, err)
	require.Equal(t, "interface", *got.ID)
	require.Equal(t, "5", *got.Index)
}

func TestGeneralUpdateOptionalFields(t *testing.T) {
	t.Parallel()
	api := mocks.NewSocketSiteClient(t)
	a := SDK{Client: api, AccountID: "account"}
	api.EXPECT().SiteUpdateSiteGeneralDetails(mock.Anything, "site", mock.MatchedBy(func(in models.UpdateSiteGeneralDetailsInput) bool {
		require.Equal(t, "", *in.Description)
		require.NotNil(t, in.SiteLocation)
		require.Nil(t, in.SiteLocation.CountryCode)
		require.Nil(t, in.SiteLocation.Timezone)
		require.Equal(t, "", *in.SiteLocation.CityName)
		return true
	}), "account").Return(nil, nil).Once()
	require.NoError(t, a.UpdateGeneral(context.Background(), "site", application.GeneralInput{Description: new(""), SiteLocation: &application.GeneralLocation{City: new("")}}))
}
