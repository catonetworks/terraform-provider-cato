package ipsecsite

import (
	"context"
	"strings"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/tfmodel"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

var (
	_ resource.Resource                = &siteIpsecResource{}
	_ resource.ResourceWithConfigure   = &siteIpsecResource{}
	_ resource.ResourceWithImportState = &siteIpsecResource{}
)

func NewSiteIpsecResource() resource.Resource {
	return &siteIpsecResource{}
}

type siteIpsecResource struct {
	client *catoclient.Service
}

func (r *siteIpsecResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ipsec_site"
}

func (r *siteIpsecResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoclient.Service)
}

func (r *siteIpsecResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

//nolint:gocyclo,funlen // Existing create flow follows several API calls that must remain ordered.
func (r *siteIpsecResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SiteIpsecIkeV2
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Hydrate API input for site creation
	input, diags := hydrateAddIpsecIkeV2Site(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Create.SiteAddIpsecIkeV2Site.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})
	ipsecSite, err := r.client.Catov2.SiteAddIpsecIkeV2Site(ctx, input, r.client.AccountId)
	tflog.Debug(ctx, "Create.SiteAddIpsecIkeV2Site.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(ipsecSite),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Cato API error",
			err.Error(),
		)
		return
	}
	// overiding state with socket site id
	resp.State.SetAttribute(ctx, path.Empty().AtName("id"), types.StringValue(ipsecSite.Site.AddIpsecIkeV2Site.GetSiteID()))

	// retrieving native-network range ID to update native range
	entityParent := cato_models.EntityInput{
		ID:   ipsecSite.Site.AddIpsecIkeV2Site.GetSiteID(),
		Type: cato_models.EntityType("site"),
	}

	siteRangeEntities, err := r.client.Catov2.EntityLookup(
		ctx,
		r.client.AccountId,
		cato_models.EntityType("siteRange"),
		nil,
		nil,
		&entityParent,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	tflog.Debug(ctx, "Create.EntityLookup.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(siteRangeEntities),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API EntityLookup error",
			err.Error(),
		)
		return
	}

	var networkRangeEntity cato_go_sdk.EntityLookup_EntityLookup_Items_Entity
	for _, item := range siteRangeEntities.EntityLookup.Items {
		splitName := strings.Split(*item.Entity.Name, " \\ ")
		if splitName[2] == "Native Range" {
			networkRangeEntity = item.Entity
		}
	}

	siteID := ipsecSite.Site.AddIpsecIkeV2Site.GetSiteID()

	// Hydrate API input for IPSec general details
	inputIpsecGeneralDetails, diags := hydrateUpdateIpsecIkeV2SiteGeneralDetails(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Create.SiteUpdateIpsecIkeV2SiteGeneralDetails.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(inputIpsecGeneralDetails),
	})
	ipsecGeneralDetailsResponse, err := r.client.Catov2.SiteUpdateIpsecIkeV2SiteGeneralDetails(
		ctx,
		siteID,
		inputIpsecGeneralDetails,
		r.client.AccountId,
	)
	tflog.Debug(ctx, "Create.SiteUpdateIpsecIkeV2SiteGeneralDetails.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(ipsecGeneralDetailsResponse),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API SiteUpdateIpsecIkeV2SiteGeneralDetails error",
			err.Error(),
		)
		return
	}

	// Hydrate API input for tunnels
	tunnelInputs, diags := hydrateAddIpsecIkeV2SiteTunnels(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Create.SiteAddIpsecIkeV2SiteTunnels.request")
	tunnelData, errIPSec := r.client.Catov2.SiteAddIpsecIkeV2SiteTunnels(ctx, siteID, tunnelInputs.add, r.client.AccountId)
	tflog.Debug(ctx, "Create.SiteAddIpsecIkeV2SiteTunnels.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(tunnelData),
	})
	if errIPSec != nil {
		resp.Diagnostics.AddError(
			"Cato API error in SiteAddIpsecIkeV2SiteTunnels",
			errIPSec.Error(),
		)
		return
	}

	// create types to support multiple primary and secondary tunnels
	addTunnels := tunnelData.Site.GetAddIpsecIkeV2SiteTunnels()
	tunnelsPrimaryData := addTunnels.PrimaryAddIpsecIkeV2SiteTunnelsPayload.GetTunnels()
	tunnelsSecondaryData := addTunnels.SecondaryAddIpsecIkeV2SiteTunnelsPayload.GetTunnels()

	// Hydrate the state with API data to ensure consistency
	hydratedState, _, hydrateErr := r.hydrateIpsecSiteState(ctx, plan, siteID)
	if hydrateErr != nil {
		resp.Diagnostics.AddError(
			"Error hydrating IPSec site state",
			hydrateErr.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// supports multiple primary ipsec tunnels
	if tunnelInputs.add.Primary != nil {
		for x := 0; x < len(tunnelsPrimaryData); x++ {
			tunnelIDPath := path.Root("ipsec").AtName("primary").AtName("tunnels").AtListIndex(x).AtName("tunnel_id")
			resp.State.SetAttribute(ctx, tunnelIDPath, tunnelsPrimaryData[x].GetTunnelIDAddIpsecIkeV2SiteTunnelPayload().String())
		}
	}

	// supports multiple secondary ipsec tunnels
	if tunnelInputs.add.Secondary != nil {
		for x := 0; x < len(tunnelsSecondaryData); x++ {
			tunnelIDPath := path.Root("ipsec").AtName("secondary").AtName("tunnels").AtListIndex(x).AtName("tunnel_id")
			resp.State.SetAttribute(ctx, tunnelIDPath, tunnelsSecondaryData[x].GetTunnelIDAddIpsecIkeV2SiteTunnelPayload().String())
		}
	}

	// Override computed fields that hydrate might not get from AccountSnapshot
	resp.State.SetAttribute(ctx, path.Empty().AtName("id"), types.StringValue(siteID))
	resp.State.SetAttribute(ctx, path.Empty().AtName("native_network_range_id"), networkRangeEntity.ID)
	resp.State.SetAttribute(ctx, path.Root("ipsec").AtName("site_id"), siteID)
}

func (r *siteIpsecResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SiteIpsecIkeV2
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Hydrate the state with API data
	hydratedState, siteExists, err := r.hydrateIpsecSiteState(ctx, state, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error hydrating IPSec site state",
			err.Error(),
		)
		return
	}

	// Check if site was found, else remove resource
	if !siteExists {
		tflog.Warn(ctx, "site not found, site resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

//nolint:gocyclo,funlen // Existing update flow follows several API calls that must remain ordered.
func (r *siteIpsecResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SiteIpsecIkeV2
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state SiteIpsecIkeV2
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tunnelInput, diags := hydrateUpdateIpsecIkeV2SiteTunnels(ctx, plan, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// setting input & input to update network range
	inputSiteGeneral := cato_models.UpdateSiteGeneralDetailsInput{
		SiteLocation: &cato_models.UpdateSiteLocationInput{},
	}

	inputUpdateNetworkRange, updateNetworkRange := prepareIpsecNetworkRangeUpdate(plan, state)

	// setting input site location
	if !plan.SiteLocation.IsNull() {
		inputSiteGeneral.SiteLocation = &cato_models.UpdateSiteLocationInput{}
		siteLocationInput := tfmodel.SiteLocation{}
		diags = plan.SiteLocation.As(ctx, &siteLocationInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)

		inputSiteGeneral.SiteLocation.Address = siteLocationInput.Address.ValueStringPointer()
		inputSiteGeneral.SiteLocation.CountryCode = siteLocationInput.CountryCode.ValueStringPointer()
		inputSiteGeneral.SiteLocation.StateCode = siteLocationInput.StateCode.ValueStringPointer()
		inputSiteGeneral.SiteLocation.Timezone = siteLocationInput.Timezone.ValueStringPointer()
	}

	inputSiteGeneral.Name = plan.Name.ValueStringPointer()
	inputSiteGeneral.SiteType = (*cato_models.SiteType)(plan.SiteType.ValueStringPointer())
	inputSiteGeneral.Description = plan.Description.ValueStringPointer()

	tflog.Debug(ctx, "Update.SiteUpdateSiteGeneralDetails.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(inputSiteGeneral),
	})
	inputSiteGeneralResponse, err := r.client.Catov2.SiteUpdateSiteGeneralDetails(
		ctx,
		plan.ID.ValueString(),
		inputSiteGeneral,
		r.client.AccountId,
	)
	tflog.Debug(ctx, "Update.SiteUpdateSiteGeneralDetails", map[string]interface{}{
		"response": utils.InterfaceToJSONString(inputSiteGeneralResponse),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API SiteUpdateSiteGeneralDetails error",
			err.Error(),
		)
		return
	}

	if updateNetworkRange {
		tflog.Debug(ctx, "Update.SiteUpdateNetworkRange.request", map[string]interface{}{
			"request": utils.InterfaceToJSONString(inputUpdateNetworkRange),
		})
		// TODO, look at why response object does not resolve
		_, err = r.client.Catov2.SiteUpdateNetworkRange(
			ctx,
			plan.NativeNetworkRangeID.ValueString(),
			*inputUpdateNetworkRange,
			r.client.AccountId,
		)
		if err != nil {
			resp.Diagnostics.AddError(
				"Catov2 API SiteUpdateNetworkRange error",
				err.Error(),
			)
			return
		}
	}

	// Hydrate API input for IPSec general details
	inputIpsecGeneralDetails, diags := hydrateUpdateIpsecIkeV2SiteGeneralDetails(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get site ID from plan
	planIPSec := AddIpsecIkeV2SiteTunnelsInput{}
	diags = plan.IPSec.As(ctx, &planIPSec, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	siteID := planIPSec.SiteID.ValueString()

	tflog.Debug(ctx, "Update.SiteUpdateIpsecIkeV2SiteGeneralDetails.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(inputIpsecGeneralDetails),
	})
	ipsecGeneralDetailsResponse, err := r.client.Catov2.SiteUpdateIpsecIkeV2SiteGeneralDetails(
		ctx,
		siteID,
		inputIpsecGeneralDetails,
		r.client.AccountId,
	)
	tflog.Debug(ctx, "Update.SiteUpdateIpsecIkeV2SiteGeneralDetails.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(ipsecGeneralDetailsResponse),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API SiteUpdateIpsecIkeV2SiteGeneralDetails error",
			err.Error(),
		)
		return
	}

	// Hydrate the state with API data to ensure consistency
	hydratedState, siteExists, hydrateErr := r.hydrateIpsecSiteState(ctx, plan, plan.ID.ValueString())
	if hydrateErr != nil {
		resp.Diagnostics.AddError(
			"Error hydrating IPSec site state",
			hydrateErr.Error(),
		)
		return
	}

	// Check if site was found, else remove resource
	if !siteExists {
		tflog.Warn(ctx, "site not found after update, site resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	if !plan.IPSec.IsNull() {
		tflog.Debug(ctx, "Update.SiteUpdateIpsecIkeV2SiteTunnels.request")
		tunnelData, errIPSec := r.client.Catov2.SiteUpdateIpsecIkeV2SiteTunnels(ctx, siteID, tunnelInput, r.client.AccountId)
		tflog.Debug(ctx, "Update.SiteUpdateIpsecIkeV2SiteTunnels.response", map[string]interface{}{
			"response": utils.InterfaceToJSONString(tunnelData),
		})
		if errIPSec != nil {
			resp.Diagnostics.AddError(
				"Cato API error in SiteAddIpsecIkeV2SiteTunnels",
				errIPSec.Error(),
			)
			return
		}

		// create types to support multiple primary and secondary tunnels
		updateTunnels := tunnelData.Site.GetUpdateIpsecIkeV2SiteTunnels()
		if len(updateTunnels.GetPrimaryUpdateIpsecIkeV2SiteTunnelsPayload().GetTunnels()) > 0 {
			tunnelsPrimaryData := updateTunnels.GetPrimaryUpdateIpsecIkeV2SiteTunnelsPayload().GetTunnels()
			for x := 0; x < len(tunnelsPrimaryData); x++ {
				tunnelIDPath := path.Root("ipsec").AtName("primary").AtName("tunnels").AtListIndex(x).AtName("tunnel_id")
				resp.State.SetAttribute(ctx, tunnelIDPath, tunnelsPrimaryData[x].GetTunnelIDUpdateIpsecIkeV2SiteTunnelPayload().String())
			}
		}

		if len(updateTunnels.GetSecondaryUpdateIpsecIkeV2SiteTunnelsPayload().GetTunnels()) > 0 {
			tunnelsSecondaryData := updateTunnels.GetSecondaryUpdateIpsecIkeV2SiteTunnelsPayload().GetTunnels()
			for x := 0; x < len(tunnelsSecondaryData); x++ {
				tunnelIDPath := path.Root("ipsec").AtName("primary").AtName("tunnels").AtListIndex(x).AtName("tunnel_id")
				resp.State.SetAttribute(ctx, tunnelIDPath, tunnelsSecondaryData[x].GetTunnelIDUpdateIpsecIkeV2SiteTunnelPayload().String())
			}
		}
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func prepareIpsecNetworkRangeUpdate(plan, state SiteIpsecIkeV2) (*cato_models.UpdateNetworkRangeInput, bool) {
	if plan.NativeNetworkRange.Equal(state.NativeNetworkRange) {
		return nil, false
	}

	return &cato_models.UpdateNetworkRangeInput{
		Subnet: plan.NativeNetworkRange.ValueStringPointer(),
	}, true
}

func (r *siteIpsecResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SiteIpsecIkeV2
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	querySiteResult, err := r.client.Catov2.EntityLookup(
		ctx,
		r.client.AccountId,
		cato_models.EntityType("site"),
		nil,
		nil,
		nil,
		nil,
		[]string{state.ID.ValueString()},
		nil,
		nil,
		nil,
	)
	tflog.Debug(ctx, "Delete.EntityLookup.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(querySiteResult),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API error",
			err.Error(),
		)
		return
	}

	// check if site exist before removing
	if len(querySiteResult.EntityLookup.GetItems()) == 1 {
		tflog.Debug(ctx, "Delete site")
		_, err := r.client.Catov2.SiteRemoveSite(ctx, state.ID.ValueString(), r.client.AccountId)
		if err != nil {
			resp.Diagnostics.AddError(
				"Catov2 API SiteRemoveSite error",
				err.Error(),
			)
			return
		}
	}
}
