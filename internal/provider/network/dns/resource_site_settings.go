package dns

import (
	"context"

	models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	nc "github.com/catonetworks/terraform-provider-cato/internal/provider/network/networkconfig"
)

type siteSettingsResource struct{ nc.Resource }

func NewSiteSettingsResource() resource.Resource { return &siteSettingsResource{} }
func (r *siteSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_site_settings"
}
func (r *siteSettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site_id"), req.ID)...)
}

func (r *siteSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SiteSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := dnsInput(ctx, plan.PrimaryServer, plan.SecondaryServer, plan.Suffix, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.Client.DNS().UpdateDNSSiteSettings(ctx, r.Client.AccountId, plan.SiteID.ValueString(), input)
	if err != nil {
		nc.APIError(&resp.Diagnostics, "updateSiteSettings", err)
		return
	}
	if !nc.RequirePayload(result != nil &&
		result.NetworkConfig != nil &&
		result.NetworkConfig.DNS.UpdateSiteSettings != nil, &resp.Diagnostics) {
		return
	}
	items := result.NetworkConfig.DNS.UpdateSiteSettings.SiteSettings
	if !nc.RequirePayload(len(items) == 1 && items[0] != nil && items[0].SiteID == plan.SiteID.ValueString(), &resp.Diagnostics) {
		return
	}
	item := items[0]
	plan.ID = types.StringValue(plan.SiteID.ValueString())
	hydrateDNS(ctx, item, &plan.PrimaryServer, &plan.SecondaryServer, &plan.Suffix, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

func (r *siteSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SiteSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := dnsInput(ctx, plan.PrimaryServer, plan.SecondaryServer, plan.Suffix, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.Client.DNS().UpdateDNSSiteSettings(ctx, r.Client.AccountId, plan.SiteID.ValueString(), input)
	if err != nil {
		nc.APIError(&resp.Diagnostics, "updateSiteSettings", err)
		return
	}
	if !nc.RequirePayload(result != nil &&
		result.NetworkConfig != nil &&
		result.NetworkConfig.DNS.UpdateSiteSettings != nil, &resp.Diagnostics) {
		return
	}
	items := result.NetworkConfig.DNS.UpdateSiteSettings.SiteSettings
	if !nc.RequirePayload(len(items) == 1 && items[0] != nil && items[0].SiteID == plan.SiteID.ValueString(), &resp.Diagnostics) {
		return
	}
	item := items[0]
	plan.ID = types.StringValue(plan.SiteID.ValueString())
	hydrateDNS(ctx, item, &plan.PrimaryServer, &plan.SecondaryServer, &plan.Suffix, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

func (r *siteSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SiteSettingsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.Client.DNS().NetworkConfigDNSSiteSettings(ctx, r.Client.AccountId,
		models.SiteRefInput{By: models.ObjectRefByID, Input: state.SiteID.ValueString()},
	)
	if err != nil {
		nc.APIError(&resp.Diagnostics, "siteSettings", err)
		return
	}
	if !nc.RequirePayload(result != nil && result.NetworkConfig != nil, &resp.Diagnostics) {
		return
	}
	item := result.NetworkConfig.DNS.SiteSettings
	if item == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if !nc.RequirePayload(item.SiteID == state.SiteID.ValueString(), &resp.Diagnostics) {
		return
	}
	state.ID = types.StringValue(item.SiteID)
	hydrateDNS(ctx, item, &state.PrimaryServer, &state.SecondaryServer, &state.Suffix, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	}
}

func (r *siteSettingsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SiteSettingsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := models.NetworkConfigDNSRemoveSiteSettingsInput{Site: []*models.SiteRefInput{{By: models.ObjectRefByID,
		Input: state.SiteID.ValueString()}}}
	result, err := r.Client.DNS().NetworkConfigDNSRemoveSiteSettings(ctx, r.Client.AccountId, input)
	if err != nil {
		if !nc.IsNotFound(err, "removeSiteSettings") {
			nc.APIError(&resp.Diagnostics, "removeSiteSettings", err)
		}
		return
	}
	nc.RequirePayload(result != nil && result.NetworkConfig != nil && result.NetworkConfig.DNS.RemoveSiteSettings != nil, &resp.Diagnostics)
}
