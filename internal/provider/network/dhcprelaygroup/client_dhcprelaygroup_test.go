package dhcprelaygroup

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

func TestInjectedClientReadErrorRetainsState(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := mocks.NewDHCPRelayGroupResourceClient(t)
	client.EXPECT().NetworkConfigDhcpRelayGroup(mock.Anything, "account-123", mock.Anything).Return(nil, errors.New("unavailable")).Once()
	r := &relayGroupResource{client: &catoclient.Service{AccountId: "account-123"}, apiClient: client}
	var schema resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schema)
	state := tfsdk.State{Schema: schema.Schema}
	require.False(t, state.Set(ctx, RelayGroupModel{ID: types.StringValue("opaque-123"), Name: types.StringValue("test"), Server: types.ListValueMust(types.StringType, nil)}).HasError())
	response := &resource.ReadResponse{State: state}
	r.Read(ctx, resource.ReadRequest{State: state}, response)
	require.True(t, response.Diagnostics.HasError())
	require.Equal(t, state.Raw, response.State.Raw)
}
