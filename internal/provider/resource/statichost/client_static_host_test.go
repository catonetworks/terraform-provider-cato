package statichost

import (
	"context"
	"errors"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

func TestStaticHostCreateUsesInjectedClient(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	api := mocks.NewStaticHostResourceClient(t)
	api.EXPECT().SiteAddStaticHost(mock.Anything, "site-123", mock.MatchedBy(func(input cato_models.AddStaticHostInput) bool {
		return input.Name == "host" && input.IP == "192.0.2.10" && input.MacAddress == nil
	}), "account-123").Return(&cato.SiteAddStaticHost{Site: cato.SiteAddStaticHost_Site{
		AddStaticHost: &cato.SiteAddStaticHost_Site_AddStaticHost{HostID: "host-123"},
	}}, nil).Once()
	model := StaticHost{ID: types.StringUnknown(), SiteID: types.StringValue("site-123"), Name: types.StringValue("host"), IP: types.StringValue("192.0.2.10"), MacAddress: types.StringNull()}
	plan := tfsdk.Plan{Schema: getStaticHostSchema(ctx, t)}
	require.False(t, plan.Set(ctx, model).HasError())
	// The service deliberately has no SDK client: any bypass of apiClient fails this test.
	r := &staticHostResource{client: &catoclient.Service{AccountId: "account-123"}, apiClient: api}
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: plan.Schema}}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, resp)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	model.ID = types.StringValue("host-123")
	assertStaticHostState(ctx, t, resp.State, model)
}

func TestStaticHostUpdateUsesInjectedClientAndPreservesErrors(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name     string
		apiError error
	}{
		{name: "success"}, {name: "API error", apiError: errors.New("update failed")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			api := mocks.NewStaticHostResourceClient(t)
			api.EXPECT().SiteUpdateStaticHost(mock.Anything, "host-123", mock.MatchedBy(func(input cato_models.UpdateStaticHostInput) bool {
				return input.Name != nil && *input.Name == "updated" && input.IP != nil && *input.IP == "192.0.2.11" && input.MacAddress == nil
			}), "account-123").Return(&cato.SiteUpdateStaticHost{}, tc.apiError).Once()
			model := StaticHost{ID: types.StringValue("host-123"), SiteID: types.StringValue("site-123"), Name: types.StringValue("updated"), IP: types.StringValue("192.0.2.11"), MacAddress: types.StringNull()}
			plan := tfsdk.Plan{Schema: getStaticHostSchema(ctx, t)}
			require.False(t, plan.Set(ctx, model).HasError())
			r := &staticHostResource{client: &catoclient.Service{AccountId: "account-123"}, apiClient: api}
			resp := &resource.UpdateResponse{State: tfsdk.State{Schema: plan.Schema}}
			r.Update(ctx, resource.UpdateRequest{Plan: plan}, resp)
			require.Equal(t, tc.apiError != nil, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			if tc.apiError == nil {
				assertStaticHostState(ctx, t, resp.State, model)
			} else {
				require.Contains(t, resp.Diagnostics[0].Detail(), tc.apiError.Error())
			}
		})
	}
}
