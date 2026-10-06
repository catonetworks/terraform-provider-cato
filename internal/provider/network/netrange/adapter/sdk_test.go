package adapter

import (
	"context"
	"errors"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	apperr "github.com/catonetworks/terraform-provider-cato/internal/provider/common/apperr/application"
	dhcp "github.com/catonetworks/terraform-provider-cato/internal/provider/common/dhcp/application"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/network/netrange/application"
)

//nolint:gocyclo // Payload matchers assert independent forwarded fields.
func TestMutationPayloads(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	api := mocks.NewNetworkRangeClient(t)
	a := SDK{Client: api, AccountID: "account"}
	in := application.Input{ID: "range", Name: new("name"), RangeType: new("VLAN"), Subnet: new("10.0.0.0/24"), LocalIP: new("10.0.0.1"), InternetOnly: new(false), MdnsReflector: new(true), Vlan: new(int64(123)), DhcpSettings: &dhcp.Settings{DhcpType: "DHCP_RELAY", RelayGroupID: new("relay")}}
	api.EXPECT().SiteAddNetworkRange(mock.Anything, "interface", mock.MatchedBy(func(v models.AddNetworkRangeInput) bool {
		return v.Name == "name" && v.RangeType == models.SubnetTypeVlan && v.Subnet == "10.0.0.0/24" && v.TranslatedSubnet == nil && v.Vlan != nil && *v.Vlan == 123 && v.DhcpSettings != nil && *v.DhcpSettings.RelayGroupID == "relay" && v.InternetOnly != nil && !*v.InternetOnly && v.MdnsReflector != nil && *v.MdnsReflector
	}), "account").Return(&cato.SiteAddNetworkRange{Site: cato.SiteAddNetworkRange_Site{AddNetworkRange: &cato.SiteAddNetworkRange_Site_AddNetworkRange{NetworkRangeID: "range"}}}, nil).Once()
	id, err := a.Add(ctx, "interface", in)
	require.NoError(t, err)
	require.Equal(t, "range", id)
	api.EXPECT().SiteUpdateNetworkRange(mock.Anything, "range", mock.MatchedBy(func(v models.UpdateNetworkRangeInput) bool {
		return v.Name != nil && *v.Name == "name" && v.RangeType != nil && *v.RangeType == models.SubnetTypeVlan && v.TranslatedSubnet == nil && v.LocalIP != nil && *v.LocalIP == "10.0.0.1" && v.DhcpSettings != nil && *v.DhcpSettings.RelayGroupID == "relay"
	}), "account").Return(nil, nil).Once()
	require.NoError(t, a.Update(ctx, in))
}
func TestInterfaceLookupShapeAndNormalization(t *testing.T) {
	t.Parallel()
	api := mocks.NewNetworkRangeClient(t)
	a := SDK{Client: api, AccountID: "account"}
	parent := &models.EntityInput{Type: models.EntityTypeSite, ID: "site"}
	response := &cato.EntityLookup{EntityLookup: cato.EntityLookup_EntityLookup{Items: []*cato.EntityLookup_EntityLookup_Items{{Entity: cato.EntityLookup_EntityLookup_Items_Entity{ID: "interface"}, HelperFields: map[string]any{"interfaceId": "5"}}}}}
	api.EXPECT().EntityLookup(mock.Anything, "account", models.EntityTypeNetworkInterface, new(int64(0)), (*int64)(nil), parent, (*string)(nil), []string{"interface"}, []*models.SortInput(nil), []*models.LookupFilterInput(nil), []string(nil)).Return(response, nil).Once()
	index, err := a.InterfaceIndex(context.Background(), "site", "interface")
	require.NoError(t, err)
	require.Equal(t, "INT_5", index)
	api.EXPECT().EntityLookup(mock.Anything, "account", models.EntityTypeNetworkInterface, new(int64(0)), (*int64)(nil), parent, (*string)(nil), []string(nil), []*models.SortInput(nil), []*models.LookupFilterInput(nil), []string(nil)).Return(response, nil).Once()
	id, err := a.InterfaceID(context.Background(), "site", "INT_5")
	require.NoError(t, err)
	require.Equal(t, "interface", id)
}
func TestMutationAbsentClassification(t *testing.T) {
	t.Parallel()
	require.ErrorIs(t, mutationError(errors.New(`{"graphqlErrors":[{"message":"Network range with id: range is not found"}]}`)), apperr.ErrAbsent)
	boom := errors.New("transport failed")
	require.ErrorIs(t, mutationError(boom), boom)
	var operation *apperr.Error
	require.ErrorAs(t, mutationError(boom), &operation)
	require.Equal(t, "Catov2 API error", operation.Summary)
}
