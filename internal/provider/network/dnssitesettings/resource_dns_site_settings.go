package dnssitesettings

import (
	"context"

	models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	nc "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/networkconfig"
)

type siteSettingsResource struct {
	client    *catoclient.Service
	apiClient ResourceClient
}

// NewDNSSiteSettingsResource returns the standalone site DNS resource.
func NewDNSSiteSettingsResource() resource.Resource { return &siteSettingsResource{} }
func (r *siteSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_site_settings"
}
func (r *siteSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData != nil {
		r.client = req.ProviderData.(*catoclient.Service)
	}
}
func (r *siteSettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
func (r *siteSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SiteSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.write(ctx, &plan, &resp.Diagnostics)
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
	r.write(ctx, &plan, &resp.Diagnostics)
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
	ref := models.SiteRefInput{By: models.ObjectRefByID, Input: state.ID.ValueString()}
	result, err := r.resourceAPI().NetworkConfigDNSSiteSettings(ctx, r.client.AccountId, ref)
	if err != nil {
		absent, lookupErr := confirmParentAbsent(ctx, r.resourceAPI(), r.client.AccountId, state.ID.ValueString(), "siteSettings", err)
		if absent {
			resp.State.RemoveResource(ctx)
		} else {
			nc.APIError(&resp.Diagnostics, "siteSettings", lookupErr)
		}
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
	if !nc.RequirePayload(item.SiteID == state.ID.ValueString(), &resp.Diagnostics) {
		return
	}
	site, err := lookupSite(ctx, r.resourceAPI(), r.client.AccountId, ref)
	if err != nil {
		nc.APIError(&resp.Diagnostics, "site reference", err)
		return
	}
	state.Site = idname.ParseIDName(ctx, *site, &resp.Diagnostics)
	nc.ParseDNS(ctx, item, &state.PrimaryServer, &state.SecondaryServer, &state.Suffix, &resp.Diagnostics)
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
		Input: state.ID.ValueString()}}}
	result, err := r.resourceAPI().NetworkConfigDNSRemoveSiteSettings(ctx, r.client.AccountId, input)
	if err != nil {
		if !nc.IsNotFound(err, "removeSiteSettings") {
			absent, lookupErr := confirmParentAbsent(ctx, r.resourceAPI(), r.client.AccountId, state.ID.ValueString(), "removeSiteSettings", err)
			if !absent {
				nc.APIError(&resp.Diagnostics, "removeSiteSettings", lookupErr)
			}
		}
		return
	}
	nc.RequirePayload(result != nil && result.NetworkConfig != nil && result.NetworkConfig.DNS.RemoveSiteSettings != nil, &resp.Diagnostics)
}

func (r *siteSettingsResource) write(ctx context.Context, plan *SiteSettingsModel, diags *diag.Diagnostics) {
	ref, input := buildSiteSettings(ctx, *plan, diags)
	if ref == nil {
		if !diags.HasError() {
			diags.AddError("Invalid site reference", "A site ID or name is required.")
		}
		return
	}
	if diags.HasError() {
		return
	}
	site, err := lookupSite(ctx, r.resourceAPI(), r.client.AccountId, *ref)
	if err != nil {
		nc.APIError(diags, "site reference", err)
		return
	}
	if !plan.ID.IsNull() && !plan.ID.IsUnknown() && plan.ID.ValueString() != site.ID {
		diags.AddError("Site mismatch", "Changing site identity requires replacement.")
		return
	}
	result,
		err := r.resourceAPI().UpdateDNSSiteSettings(ctx,
		r.client.AccountId,
		models.SiteRefInput{By: models.ObjectRefByID,
			Input: site.ID},
		input)
	if err != nil {
		nc.APIError(diags, "updateSiteSettings", err)
		return
	}
	if !nc.RequirePayload(result != nil && result.NetworkConfig != nil && result.NetworkConfig.DNS.UpdateSiteSettings != nil, diags) {
		return
	}
	items := result.NetworkConfig.DNS.UpdateSiteSettings.SiteSettings
	if !nc.RequirePayload(len(items) == 1 && items[0] != nil && items[0].SiteID == site.ID, diags) {
		return
	}
	plan.ID = types.StringValue(site.ID)
	plan.Site = idname.ParseIDName(ctx, *site, diags)
	nc.ParseDNS(ctx, items[0], &plan.PrimaryServer, &plan.SecondaryServer, &plan.Suffix, diags)
}
