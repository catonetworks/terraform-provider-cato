package popbgpprofile

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
	_ resource.Resource                = (*bgpProfileResource)(nil)
	_ resource.ResourceWithConfigure   = (*bgpProfileResource)(nil)
	_ resource.ResourceWithImportState = (*bgpProfileResource)(nil)
)

type bgpProfileResource struct {
	client    *catoclient.Service
	apiClient ResourceClient
}

// NewResource returns the PoP BGP profile resource.
func NewResource() resource.Resource { return &bgpProfileResource{} }
func (r *bgpProfileResource) getClient() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	if r.client == nil || r.client.Catov2 == nil {
		return nil
	}
	return &sdkClient{client: r.client}
}
func (r *bgpProfileResource) configured(diags *diag.Diagnostics) bool {
	if r.client == nil || r.client.AccountId == "" || r.getClient() == nil {
		diags.AddError("Resource not configured", "A Cato client and explicit account ID are required.")
		return false
	}
	return true
}
func (r *bgpProfileResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pop_bgp_profile"
}
func (r *bgpProfileResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *bgpProfileResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
func (r *bgpProfileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := buildCreate(plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || !r.configured(&resp.Diagnostics) {
		return
	}
	result, err := r.getClient().PopLocationMutationsAddBgpProfile(ctx, r.client.AccountId, input)
	item := result.GetPopLocationMutations().GetAddBgpProfile().GetBgpProfile()
	if item.GetID() != "" {
		resp.Diagnostics.Append(resp.State.Set(ctx, identityState(plan, item.GetID()))...)
	}
	if err != nil {
		resp.Diagnostics.AddError("Failed to create PoP BGP profile", err.Error())
		return
	}
	if !validID(types.StringValue(item.GetID()), &resp.Diagnostics) {
		return
	}
	current := parseProfile(item, item.GetCommunities(), &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
	}
}
func (r *bgpProfileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || !validID(state.ID, &resp.Diagnostics) || !r.configured(&resp.Diagnostics) {
		return
	}
	item, err := r.find(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read PoP BGP profile", err.Error())
		return
	}
	if item == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	current := parseProfile(item, item.GetCommunities(), &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
	}
}
func (r *bgpProfileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || !validID(state.ID, &resp.Diagnostics) {
		return
	}
	buildCreate(plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || !r.configured(&resp.Diagnostics) {
		return
	}
	existing, err := r.find(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read PoP BGP profile before update", err.Error())
		return
	}
	if existing == nil {
		resp.Diagnostics.AddError("Profile no longer exists", "Refresh state before applying again.")
		return
	}
	current := parseProfile(existing, existing.GetCommunities(), &resp.Diagnostics)
	input := buildUpdate(plan, current, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.getClient().PopLocationMutationsUpdateBgpProfile(ctx, r.client.AccountId, input)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update PoP BGP profile", err.Error())
		return
	}
	item := result.GetPopLocationMutations().GetUpdateBgpProfile().GetBgpProfile()
	if item.GetID() != state.ID.ValueString() {
		resp.Diagnostics.AddError("Invalid API response", "Update returned a different BGP profile identity.")
		return
	}
	current = parseProfile(item, item.GetCommunities(), &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
	}
}
func (r *bgpProfileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || !validID(state.ID, &resp.Diagnostics) || !r.configured(&resp.Diagnostics) {
		return
	}
	existing, err := r.find(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read PoP BGP profile before deletion", err.Error())
		return
	}
	if existing == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	result, err := r.getClient().PopLocationMutationsRemoveBgpProfile(ctx, r.client.AccountId, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to remove PoP BGP profile", err.Error())
		return
	}
	if result.GetPopLocationMutations().GetRemoveBgpProfile().GetBgpProfile().GetID() != state.ID.ValueString() {
		resp.Diagnostics.AddError("Invalid API response", "Deletion did not confirm the requested BGP profile identity.")
		return
	}
	resp.State.RemoveResource(ctx)
}
