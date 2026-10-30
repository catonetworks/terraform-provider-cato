package secondaryazurevsocket

import (
	"context"
	"fmt"

	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

var (
	_ resource.Resource                = (*socketResource)(nil)
	_ resource.ResourceWithConfigure   = (*socketResource)(nil)
	_ resource.ResourceWithImportState = (*socketResource)(nil)
)

type socketResource struct {
	client    *catoclient.Service
	apiClient ResourceClient
}

// NewResource returns the secondary Azure vSocket resource.
func NewResource() resource.Resource { return &socketResource{} }
func (r *socketResource) getClient() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	if r.client == nil || r.client.Catov2 == nil {
		return nil
	}
	return &sdkClient{client: r.client}
}
func (r *socketResource) configured(diags *diag.Diagnostics) bool {
	if r.client == nil || r.client.AccountId == "" || r.getClient() == nil {
		diags.AddError("Resource not configured", "A Cato client and explicit account ID are required.")
		return false
	}
	return true
}
func (r *socketResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secondary_azure_vsocket"
}
func (r *socketResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *socketResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
func (r *socketResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	in := buildCreate(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || !r.configured(&resp.Diagnostics) {
		return
	}
	site, err := r.resolveEmptySlot(ctx, in.Site)
	if err != nil {
		resp.Diagnostics.AddError("Failed to check secondary socket slot", err.Error())
		return
	}
	in.Site = &m.SiteRefInput{By: m.ObjectRefByID, Input: site}
	result, err := r.getClient().SiteAddSecondaryAzureVSocket(ctx, r.client.AccountId, in)
	id := result.GetSite().GetAddSecondaryAzureVSocket().GetID()
	if id != "" {
		resp.Diagnostics.Append(resp.State.Set(ctx, identityState(plan, id))...)
	}
	if err != nil {
		resp.Diagnostics.AddError("Failed to add secondary Azure vSocket", err.Error())
		return
	}
	if !validID(types.StringValue(id), &resp.Diagnostics) {
		return
	}
	item, err := r.waitForSocket(ctx, id, in)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read created secondary socket", err.Error())
		return
	}
	current := parseSocket(ctx, item, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
	}
}
func (r *socketResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || !validID(state.ID, &resp.Diagnostics) || !r.configured(&resp.Diagnostics) {
		return
	}
	item, err := r.find(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read secondary socket", err.Error())
		return
	}
	if item == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	current := parseSocket(ctx, item, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
	}
}
func (r *socketResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || !validID(state.ID, &resp.Diagnostics) {
		return
	}
	in := buildCreate(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || !r.configured(&resp.Diagnostics) {
		return
	}
	existing, err := r.find(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read secondary socket before update", err.Error())
		return
	}
	if existing == nil {
		resp.Diagnostics.AddError("Secondary socket no longer exists", "Refresh state before applying again.")
		return
	}
	if !matchesReference(in.Site, existing.GetSite().GetID(), existing.GetSite().GetName()) {
		resp.Diagnostics.AddError("Immutable socket parent", "Site changes require replacement.")
		return
	}
	result, err := r.getClient().SiteUpdateSecondaryAzureVSocket(ctx, r.client.AccountId, buildUpdate(in, state.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to update secondary socket", err.Error())
		return
	}
	updated := result.GetSite().GetUpdateSecondaryAzureVSocket().GetSecondaryAzureVSocket()
	if updated == nil || updated.GetID() != state.ID.ValueString() || updated.GetSite().GetID() != existing.GetSite().GetID() {
		resp.Diagnostics.AddError("Invalid API response", "Update did not confirm the requested secondary socket identity.")
		return
	}
	item, err := r.waitForSocket(ctx, state.ID.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read updated secondary socket", err.Error())
		return
	}
	current := parseSocket(ctx, item, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
	}
}
func (r *socketResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || !validID(state.ID, &resp.Diagnostics) || !r.configured(&resp.Diagnostics) {
		return
	}
	existing, err := r.find(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read secondary socket before deletion", err.Error())
		return
	}
	if existing == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	result, err := r.getClient().SiteRemoveSecondaryAzureVSocket(ctx, r.client.AccountId, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to remove secondary socket", err.Error())
		return
	}
	removed := result.GetSite().GetRemoveSecondaryAzureVSocket().GetSecondaryAzureVSocket()
	if removed == nil || removed.GetID() != state.ID.ValueString() || removed.GetSite().GetID() != existing.GetSite().GetID() {
		resp.Diagnostics.AddError("Invalid API response", "Removal did not confirm the requested secondary socket identity.")
		return
	}
	resp.State.RemoveResource(ctx)
}
