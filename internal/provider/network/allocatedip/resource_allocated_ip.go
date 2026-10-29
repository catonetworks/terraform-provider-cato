package allocatedip

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
	_ resource.Resource                = (*allocatedIPResource)(nil)
	_ resource.ResourceWithConfigure   = (*allocatedIPResource)(nil)
	_ resource.ResourceWithImportState = (*allocatedIPResource)(nil)
)

type allocatedIPResource struct {
	client    *catoclient.Service
	apiClient ResourceClient
}

// NewResource returns the allocated-IP resource.
func NewResource() resource.Resource { return &allocatedIPResource{} }
func (r *allocatedIPResource) getClient() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	if r.client == nil || r.client.Catov2 == nil {
		return nil
	}
	return &sdkClient{client: r.client}
}
func (r *allocatedIPResource) configured(diags *diag.Diagnostics) bool {
	if r.client == nil || r.client.AccountId == "" || r.getClient() == nil {
		diags.AddError("Resource not configured", "A Cato client and explicit account ID are required.")
		return false
	}
	return true
}
func (r *allocatedIPResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_allocated_ip"
}
func (r *allocatedIPResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *allocatedIPResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
func (r *allocatedIPResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := buildCreate(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || !r.configured(&resp.Diagnostics) {
		return
	}
	result, err := r.getClient().PopLocationMutationsAddAllocatedIP(ctx, r.client.AccountId, input)
	item := result.GetPopLocationMutations().GetAddAllocatedIP().GetAllocatedIP()
	if item.GetID() != "" {
		resp.Diagnostics.Append(resp.State.Set(ctx, identityState(plan, item.GetID(), item.GetIP()))...)
	}
	if err != nil {
		resp.Diagnostics.AddError("Failed to allocate IP", err.Error())
		return
	}
	if !validID(types.StringValue(item.GetID()), &resp.Diagnostics) {
		return
	}
	current := parseAllocation(ctx, item, item.GetPopLocation(), item.GetAccount().GetID(), r.client.AccountId, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
	}
}
func (r *allocatedIPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || !validID(state.ID, &resp.Diagnostics) || !r.configured(&resp.Diagnostics) {
		return
	}
	item, err := r.find(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read allocated IP", err.Error())
		return
	}
	if item == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	current := parseAllocation(ctx, item, item.GetPopLocation(), item.GetAccount().GetID(), r.client.AccountId, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
	}
}
func (r *allocatedIPResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || !validID(state.ID, &resp.Diagnostics) {
		return
	}
	buildCreate(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || !r.configured(&resp.Diagnostics) {
		return
	}
	existing, err := r.find(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read allocated IP before update", err.Error())
		return
	}
	if existing == nil {
		resp.Diagnostics.AddError("Allocation no longer exists", "Refresh state before applying again.")
		return
	}
	result, err := r.getClient().PopLocationMutationsUpdateAllocatedIPDescription(ctx, r.client.AccountId,
		m.PopLocationUpdateAllocatedIPDescriptionInput{IP: existing.GetIP(), Description: plan.Description.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Failed to update allocated IP description", err.Error())
		return
	}
	item := result.GetPopLocationMutations().GetUpdateAllocatedIPDescription().GetAllocatedIP()
	if item.GetID() != state.ID.ValueString() {
		resp.Diagnostics.AddError("Invalid API response", "Update returned a different allocation identity.")
		return
	}
	current := parseAllocation(ctx, item, item.GetPopLocation(), item.GetAccount().GetID(), r.client.AccountId, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
	}
}
func (r *allocatedIPResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || !validID(state.ID, &resp.Diagnostics) || !r.configured(&resp.Diagnostics) {
		return
	}
	existing, err := r.find(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read allocated IP before release", err.Error())
		return
	}
	if existing == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	result, err := r.getClient().PopLocationMutationsRemoveAllocatedIP(ctx, r.client.AccountId,
		m.PopLocationRemoveAllocatedIPInput{IP: existing.GetIP(), Type: m.PopLocationAllocatedIPTypeSystem})
	if err != nil {
		resp.Diagnostics.AddError("Failed to release allocated IP", err.Error())
		return
	}
	item := result.GetPopLocationMutations().GetRemoveAllocatedIP().GetAllocatedIP()
	if item.GetID() != state.ID.ValueString() || item.GetAccount().GetID() != r.client.AccountId {
		resp.Diagnostics.AddError("Invalid API response", "Release did not confirm the requested account/allocation identity.")
		return
	}
	resp.State.RemoveResource(ctx)
}
