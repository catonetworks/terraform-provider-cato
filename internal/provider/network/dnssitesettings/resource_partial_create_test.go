package dnssitesettings

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

func partialPlan(t *testing.T, r resource.Resource, suffix types.List) tfsdk.Plan {
	t.Helper()
	var schema resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schema)
	plan := tfsdk.Plan{Schema: schema.Schema}
	require.False(t, plan.Set(context.Background(), SiteSettingsModel{ID: types.StringUnknown(), Site: types.ObjectValueMust(idname.RefAttrTypes, map[string]attr.Value{"id": types.StringValue("site-123"), "name": types.StringValue("branch")}), PrimaryServer: types.StringNull(), SecondaryServer: types.StringNull(), Suffix: suffix}).HasError())
	return plan
}
func TestSiteCreatePreservesPartialAcknowledgement(t *testing.T) {
	ctx := context.Background()
	client := mocks.NewDNSSiteSettingsResourceClient(t)
	r := NewDNSSiteSettingsResource().(*siteSettingsResource)
	r.client = &catoclient.Service{AccountId: "account-123"}
	r.apiClient = client
	var lookup cato.SiteGeneralDetails
	require.NoError(t, json.Unmarshal([]byte(`{"site":{"siteGeneralDetails":{"site":{"id":"site-123","name":"branch"}}}}`), &lookup))
	client.EXPECT().SiteGeneralDetails(mock.Anything, mock.Anything, "account-123").Return(&lookup, nil).Once()
	var result cato.NetworkConfigDNSUpdateSiteSettings
	require.NoError(t, json.Unmarshal([]byte(`{"networkConfig":{"dns":{"updateSiteSettings":{"siteSettings":[{"siteId":"site-123","suffix":[]}]}}}}`), &result))
	client.EXPECT().UpdateDNSSiteSettings(mock.Anything, "account-123", mock.Anything, mock.Anything).Return(&result, errors.New("partial response")).Once()
	plan := partialPlan(t, r, types.ListValueMust(types.StringType, nil))
	response := resource.CreateResponse{State: tfsdk.State{Schema: plan.Schema}}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, &response)
	require.True(t, response.Diagnostics.HasError())
	var id types.String
	require.False(t, response.State.GetAttribute(ctx, path.Root("id"), &id).HasError())
	require.Equal(t, "site-123", id.ValueString())
	require.True(t, response.State.Raw.IsFullyKnown())
}
func TestInvalidSuffixNeverWrites(t *testing.T) {
	for _, suffix := range []types.List{types.ListUnknown(types.StringType), types.ListValueMust(types.StringType, []attr.Value{types.StringNull()}), types.ListValueMust(types.StringType, []attr.Value{types.StringValue("corp.example"), types.StringUnknown()})} {
		client := mocks.NewDNSSiteSettingsResourceClient(t)
		r := NewDNSSiteSettingsResource().(*siteSettingsResource)
		r.client = &catoclient.Service{AccountId: "account-123"}
		r.apiClient = client
		plan := partialPlan(t, r, suffix)
		response := resource.CreateResponse{State: tfsdk.State{Schema: plan.Schema}}
		r.Create(context.Background(), resource.CreateRequest{Plan: plan}, &response)
		require.True(t, response.Diagnostics.HasError())
		require.True(t, response.State.Raw.IsNull())
	}
}
