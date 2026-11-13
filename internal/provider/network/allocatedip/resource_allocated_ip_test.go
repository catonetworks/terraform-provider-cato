package allocatedip

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

func TestCreateStopsWhenAllocationLookupFails(t *testing.T) {
	ctx := context.Background()
	client := mocks.NewAllocatedIPResourceClient(t)
	client.EXPECT().AllocatedIPList(mock.Anything, "12345", mock.Anything).
		Return(nil, errors.New("lookup unavailable")).Once()
	r := &allocatedIPResource{client: &catoclient.Service{AccountId: "12345"}, apiClient: client}
	schema := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schema)
	plan := tfsdk.Plan{Schema: schema.Schema}
	require.False(t, plan.Set(ctx, AllocatedIP{
		PopLocationName: types.StringValue("Paris"),
		Description:     types.StringValue("dedicated"),
		Type:            types.StringValue("SYSTEM"),
	}).HasError())
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: schema.Schema}}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, resp)
	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "lookup unavailable")
}

func TestAllocationIdentityRequiresUniqueMatchingType(t *testing.T) {
	for _, tc := range []struct {
		name, items          string
		wantError, wantMatch bool
	}{
		{"unique", `[{"id":"one","ip":"192.0.2.1","allocationType":"SYSTEM","description":"dedicated","popLocation":{"name":"Paris"}}]`, false, true},
		{"wrong type", `[{"id":"one","ip":"192.0.2.1","allocationType":"OTHER","description":"dedicated","popLocation":{"name":"Paris"}}]`, false, false},
		{"ambiguous", `[{"id":"one","allocationType":"SYSTEM","description":"dedicated","popLocation":{"name":"Paris"}},{"id":"two","allocationType":"SYSTEM","description":"dedicated","popLocation":{"name":"Paris"}}]`, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var response cato.PopLocationsAllocatedIPList
			require.NoError(t, json.Unmarshal([]byte(`{"popLocations":{"allocatedIPList":{"allocatedIP":`+tc.items+`}}}`), &response))
			client := mocks.NewAllocatedIPResourceClient(t)
			client.EXPECT().AllocatedIPList(mock.Anything, "12345", mock.Anything).Return(&response, nil).Once()
			r := &allocatedIPResource{client: &catoclient.Service{AccountId: "12345"}, apiClient: client}
			got, err := r.findByIdentity(context.Background(), "Paris", "dedicated", "SYSTEM")
			require.Equal(t, tc.wantError, err != nil)
			require.Equal(t, tc.wantMatch, got != nil)
		})
	}
}
