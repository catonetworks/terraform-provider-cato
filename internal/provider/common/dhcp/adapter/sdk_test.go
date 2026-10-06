package adapter

import (
	"context"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
)

func TestRelayLookupAccountShape(t *testing.T) {
	t.Parallel()
	api := mocks.NewNetworkRangeClient(t)
	lookup := Lookup{Client: api, AccountID: "account"}
	result := &cato.EntityLookup{EntityLookup: cato.EntityLookup_EntityLookup{Items: []*cato.EntityLookup_EntityLookup_Items{{Entity: cato.EntityLookup_EntityLookup_Items_Entity{ID: "relay", Name: new("relay-name")}}}}}
	api.EXPECT().EntityLookupMinimal(mock.Anything, "account", models.EntityTypeDhcpRelayGroup, (*int64)(nil), (*int64)(nil), (*models.EntityInput)(nil), []*models.SortInput(nil), []*models.LookupFilterInput(nil)).Return(result, nil).Twice()
	id, err := lookup.RelayID(context.Background(), "relay-name")
	require.NoError(t, err)
	require.Equal(t, "relay", id)
	name, err := lookup.RelayName(context.Background(), "relay")
	require.NoError(t, err)
	require.Equal(t, "relay-name", *name)
}
