package dnssitesettings

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

func TestInjectedClientReadErrorRetainsState(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := mocks.NewDNSSiteSettingsResourceClient(t)
	client.EXPECT().NetworkConfigDNSSiteSettings(mock.Anything, "account-123", mock.Anything).Return(nil, errors.New("unavailable")).Once()
	r := &siteSettingsResource{client: &catoclient.Service{AccountId: "account-123"}, apiClient: client}
	var schema resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schema)
	state := tfsdk.State{Schema: schema.Schema}
	require.False(t, state.Set(ctx, SiteSettingsModel{ID: types.StringValue("site-123"), Site: types.ObjectValueMust(idname.RefAttrTypes, map[string]attr.Value{"id": types.StringValue("site-123"), "name": types.StringValue("branch")}), PrimaryServer: types.StringNull(), SecondaryServer: types.StringNull(), Suffix: types.ListValueMust(types.StringType, nil)}).HasError())
	response := &resource.ReadResponse{State: state}
	r.Read(ctx, resource.ReadRequest{State: state}, response)
	require.True(t, response.Diagnostics.HasError())
	require.Equal(t, state.Raw, response.State.Raw)
}

func TestDeleteAlreadyMissingOverride(t *testing.T) {
	ctx := context.Background()
	var apiError clientv2.GqlErrorList
	require.NoError(t, json.Unmarshal([]byte(`{"errors":[{"message":"missing override","path":["networkConfig","dns","removeSiteSettings"],"extensions":{"code":"NOT_FOUND"}}]}`), &apiError))
	client := mocks.NewDNSSiteSettingsResourceClient(t)
	client.EXPECT().NetworkConfigDNSRemoveSiteSettings(mock.Anything, "account-123", mock.Anything).Return(nil, &apiError).Once()
	r := &siteSettingsResource{client: &catoclient.Service{AccountId: "account-123"}, apiClient: client}
	var schema resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schema)
	state := tfsdk.State{Schema: schema.Schema}
	require.False(t, state.Set(ctx, SiteSettingsModel{ID: types.StringValue("site-123"), Site: types.ObjectNull(idname.RefAttrTypes), PrimaryServer: types.StringNull(), SecondaryServer: types.StringNull(), Suffix: types.ListValueMust(types.StringType, nil)}).HasError())
	response := &resource.DeleteResponse{}
	r.Delete(ctx, resource.DeleteRequest{State: state}, response)
	require.False(t, response.Diagnostics.HasError(), response.Diagnostics)
}
