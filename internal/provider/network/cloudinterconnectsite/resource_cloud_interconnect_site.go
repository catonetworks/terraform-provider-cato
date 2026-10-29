package cloudinterconnectsite

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

var (
	_ resource.Resource                = (*cloudSiteResource)(nil)
	_ resource.ResourceWithConfigure   = (*cloudSiteResource)(nil)
	_ resource.ResourceWithImportState = (*cloudSiteResource)(nil)
)

type cloudSiteResource struct {
	client    *catoclient.Service
	apiClient ResourceClient
}

// NewResource returns the Cloud Interconnect site resource.
func NewResource() resource.Resource { return &cloudSiteResource{} }
func (r *cloudSiteResource) getClient() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	if r.client == nil || r.client.Catov2 == nil {
		return nil
	}
	return &sdkClient{client: r.client}
}
func (r *cloudSiteResource) configured(diags *diag.Diagnostics) bool {
	if r.client == nil || r.client.AccountId == "" || r.getClient() == nil {
		diags.AddError("Resource not configured", "A Cato client and explicit account ID are required.")
		return false
	}
	return true
}
func (r *cloudSiteResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_interconnect_site"
}
func (r *cloudSiteResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*catoclient.Service)
	if !ok {
		resp.Diagnostics.AddError("Unexpected resource configuration", fmt.Sprintf("Expected *catoclient.Service, got %T", req.ProviderData))
		return
	}
	r.client = client
}
func (r *cloudSiteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
func (r *cloudSiteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := buildCreate(plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || !r.configured(&resp.Diagnostics) {
		return
	}
	result, err := r.getClient().SiteAddCloudInterconnectSite(ctx, input, r.client.AccountId)
	id := result.GetSite().GetAddCloudInterconnectSite().GetSiteID()
	if id != "" {
		resp.Diagnostics.Append(resp.State.Set(ctx, identityState(plan, id))...)
	}
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Cloud Interconnect site", err.Error())
		return
	}
	if !validID(types.StringValue(id), &resp.Diagnostics) {
		return
	}
	item, err := r.waitForSite(ctx, id, buildUpdate(plan, &resp.Diagnostics))
	if err != nil {
		resp.Diagnostics.AddError("Failed to read created site", err.Error())
		return
	}
	current := parseSite(item, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
	}
}
func (r *cloudSiteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || !validID(state.ID, &resp.Diagnostics) || !r.configured(&resp.Diagnostics) {
		return
	}
	item, err := r.find(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read Cloud Interconnect site", err.Error())
		return
	}
	if item == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	current := parseSite(item, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
	}
}
func (r *cloudSiteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || !validID(state.ID, &resp.Diagnostics) {
		return
	}
	input := buildUpdate(plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || !r.configured(&resp.Diagnostics) {
		return
	}
	existing, err := r.find(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read site before update", err.Error())
		return
	}
	if existing == nil {
		resp.Diagnostics.AddError("Site no longer exists", "Refresh state before applying again.")
		return
	}
	result, err := r.getClient().SiteUpdateSiteGeneralDetails(ctx, state.ID.ValueString(), input, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update site metadata", err.Error())
		return
	}
	if result.GetSite().GetUpdateSiteGeneralDetails().GetSiteID() != state.ID.ValueString() {
		resp.Diagnostics.AddError("Invalid API response", "Update returned a different site identity.")
		return
	}
	item, err := r.waitForSite(ctx, state.ID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read updated site", err.Error())
		return
	}
	current := parseSite(item, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
	}
}
func (r *cloudSiteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || !validID(state.ID, &resp.Diagnostics) || !r.configured(&resp.Diagnostics) {
		return
	}
	existing, err := r.find(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read site before deletion", err.Error())
		return
	}
	if existing == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	connected, err := r.hasPhysicalConnections(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to check physical connections", err.Error())
		return
	}
	if connected {
		resp.Diagnostics.AddError("Site has physical connections",
			"Remove the separately managed physical connections before deleting this site.")
		return
	}
	result, err := r.getClient().SiteRemoveSite(ctx, state.ID.ValueString(), r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete Cloud Interconnect site", err.Error())
		return
	}
	if result.GetSite().GetRemoveSite().GetSiteID() != state.ID.ValueString() {
		resp.Diagnostics.AddError("Invalid API response", "Deletion did not confirm the requested site identity.")
		return
	}
	resp.State.RemoveResource(ctx)
}
