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
	expectSiteLookupByID(api, "account-123", "site-123", siteEntityLookup([2]string{"site-123", "a-site"}), nil)
	api.EXPECT().SiteAddStaticHost(mock.Anything, "site-123", mock.MatchedBy(func(input cato_models.AddStaticHostInput) bool {
		return input.Name == "host" && input.IP == "192.0.2.10" && input.MacAddress == nil
	}), "account-123").Return(&cato.SiteAddStaticHost{Site: cato.SiteAddStaticHost_Site{
		AddStaticHost: &cato.SiteAddStaticHost_Site_AddStaticHost{HostID: "host-123"},
	}}, nil).Once()
	model := StaticHost{ID: types.StringUnknown(), SiteID: types.StringValue("site-123"), SiteName: types.StringNull(), Name: types.StringValue("host"), IP: types.StringValue("192.0.2.10"), MacAddress: types.StringNull()}
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

// siteEntityLookup builds a *cato.EntityLookup response of site entities from
// (id, name) pairs -- pass none for an empty result.
func siteEntityLookup(pairs ...[2]string) *cato.EntityLookup {
	items := make([]*cato.EntityLookup_EntityLookup_Items, 0, len(pairs))
	for _, p := range pairs {
		name := p[1]
		items = append(items, &cato.EntityLookup_EntityLookup_Items{
			Entity: cato.EntityLookup_EntityLookup_Items_Entity{
				ID:   p[0],
				Name: &name,
				Type: cato_models.EntityTypeSite,
			},
		})
	}
	return &cato.EntityLookup{EntityLookup: cato.EntityLookup_EntityLookup{Items: items}}
}

func expectSiteLookupByID(api *mocks.StaticHostResourceClient, accountID, siteID string, resp *cato.EntityLookup, err error) {
	api.EXPECT().EntityLookup(mock.Anything, accountID, cato_models.EntityTypeSite,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		[]string{siteID}, mock.Anything, mock.Anything, mock.Anything).
		Return(resp, err).Once()
}

func expectSiteLookupByName(api *mocks.StaticHostResourceClient, accountID, name string, resp *cato.EntityLookup, err error) {
	api.EXPECT().EntityLookup(mock.Anything, accountID, cato_models.EntityTypeSite,
		mock.Anything, mock.Anything, mock.Anything,
		mock.MatchedBy(func(search *string) bool { return search != nil && *search == name }),
		mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(resp, err).Once()
}

func TestStaticHostCreateResolvesSiteByName(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	api := mocks.NewStaticHostResourceClient(t)
	// search is fuzzy: a non-exact hit is returned alongside the exact one and
	// must be ignored. Only "HQ Site" (exact) resolves.
	expectSiteLookupByName(api, "account-123", "HQ Site",
		siteEntityLookup([2]string{"site-x", "HQ Site Backup"}, [2]string{"site-999", "HQ Site"}), nil)
	api.EXPECT().SiteAddStaticHost(mock.Anything, "site-999", mock.MatchedBy(func(input cato_models.AddStaticHostInput) bool {
		return input.Name == "host" && input.IP == "192.0.2.10"
	}), "account-123").Return(&cato.SiteAddStaticHost{Site: cato.SiteAddStaticHost_Site{
		AddStaticHost: &cato.SiteAddStaticHost_Site_AddStaticHost{HostID: "host-123"},
	}}, nil).Once()
	model := StaticHost{ID: types.StringUnknown(), SiteID: types.StringNull(), SiteName: types.StringValue("HQ Site"), Name: types.StringValue("host"), IP: types.StringValue("192.0.2.10"), MacAddress: types.StringNull()}
	plan := tfsdk.Plan{Schema: getStaticHostSchema(ctx, t)}
	require.False(t, plan.Set(ctx, model).HasError())
	r := &staticHostResource{client: &catoclient.Service{AccountId: "account-123"}, apiClient: api}
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: plan.Schema}}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, resp)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	var got StaticHost
	require.False(t, resp.State.Get(ctx, &got).HasError())
	require.Equal(t, "site-999", got.SiteID.ValueString(), "resolved site_id written to state")
	require.Equal(t, "HQ Site", got.SiteName.ValueString())
	require.Equal(t, "host-123", got.ID.ValueString())
}

func TestStaticHostCreateSiteResolutionErrors(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name  string
		setup func(api *mocks.StaticHostResourceClient)
		want  string
	}{
		{
			name: "name not found (only a fuzzy non-exact hit)",
			setup: func(api *mocks.StaticHostResourceClient) {
				expectSiteLookupByName(api, "account-123", "HQ Site",
					siteEntityLookup([2]string{"site-x", "HQ Site Backup"}), nil)
			},
			want: "No site named",
		},
		{
			name: "ambiguous name",
			setup: func(api *mocks.StaticHostResourceClient) {
				expectSiteLookupByName(api, "account-123", "Dup Site",
					siteEntityLookup([2]string{"site-1", "Dup Site"}, [2]string{"site-2", "Dup Site"}), nil)
			},
			want: "Ambiguous site name",
		},
		{
			name: "id not found",
			setup: func(api *mocks.StaticHostResourceClient) {
				expectSiteLookupByID(api, "account-123", "missing", siteEntityLookup(), nil)
			},
			want: "No site with ID",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			api := mocks.NewStaticHostResourceClient(t)
			tc.setup(api)
			// SiteAddStaticHost must NOT be called -- resolution fails first.
			model := StaticHost{ID: types.StringUnknown(), Name: types.StringValue("host"), IP: types.StringValue("192.0.2.10"), MacAddress: types.StringNull()}
			switch tc.want {
			case "No site with ID":
				model.SiteID, model.SiteName = types.StringValue("missing"), types.StringNull()
			case "Ambiguous site name":
				model.SiteID, model.SiteName = types.StringNull(), types.StringValue("Dup Site")
			default:
				model.SiteID, model.SiteName = types.StringNull(), types.StringValue("HQ Site")
			}
			plan := tfsdk.Plan{Schema: getStaticHostSchema(ctx, t)}
			require.False(t, plan.Set(ctx, model).HasError())
			r := &staticHostResource{client: &catoclient.Service{AccountId: "account-123"}, apiClient: api}
			resp := &resource.CreateResponse{State: tfsdk.State{Schema: plan.Schema}}
			r.Create(ctx, resource.CreateRequest{Plan: plan}, resp)
			require.True(t, resp.Diagnostics.HasError())
			require.Contains(t, resp.Diagnostics[0].Summary()+" "+resp.Diagnostics[0].Detail(), tc.want)
		})
	}
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
			expectSiteLookupByID(api, "account-123", "site-123", siteEntityLookup([2]string{"site-123", "a-site"}), nil)
			api.EXPECT().SiteUpdateStaticHost(mock.Anything, "host-123", mock.MatchedBy(func(input cato_models.UpdateStaticHostInput) bool {
				return input.Name != nil && *input.Name == "updated" && input.IP != nil && *input.IP == "192.0.2.11" && input.MacAddress == nil
			}), "account-123").Return(&cato.SiteUpdateStaticHost{}, tc.apiError).Once()
			model := StaticHost{ID: types.StringValue("host-123"), SiteID: types.StringValue("site-123"), SiteName: types.StringNull(), Name: types.StringValue("updated"), IP: types.StringValue("192.0.2.11"), MacAddress: types.StringNull()}
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

func TestResolveSiteRejectsMissingOrUnknownID(t *testing.T) {
	for _, id := range []types.String{types.StringNull(), types.StringUnknown(), types.StringValue("")} {
		api := mocks.NewStaticHostResourceClient(t)
		r := &staticHostResource{client: &catoclient.Service{AccountId: "account-123"}, apiClient: api}
		got, diags := r.resolveSiteID(context.Background(), StaticHost{SiteID: id, SiteName: types.StringNull()})
		require.Empty(t, got)
		require.True(t, diags.HasError())
	}
}
