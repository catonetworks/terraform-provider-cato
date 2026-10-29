package cloudinterconnectphysicalconnection

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
	_ resource.Resource                = (*connectionResource)(nil)
	_ resource.ResourceWithConfigure   = (*connectionResource)(nil)
	_ resource.ResourceWithImportState = (*connectionResource)(nil)
)

type connectionResource struct {
	client    *catoclient.Service
	apiClient ResourceClient
}

// NewResource returns the manual Cloud Interconnect connection resource.
func NewResource() resource.Resource { return &connectionResource{} }
func (r *connectionResource) getClient() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	if r.client == nil || r.client.Catov2 == nil {
		return nil
	}
	return &sdkClient{client: r.client}
}
func (r *connectionResource) configured(diags *diag.Diagnostics) bool {
	if r.client == nil || r.client.AccountId == "" || r.getClient() == nil {
		diags.AddError("Resource not configured", "A Cato client and explicit account ID are required.")
		return false
	}
	return true
}
func (r *connectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_interconnect_physical_connection"
}
func (r *connectionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *connectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
func (r *connectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := buildConfiguredCreate(ctx, req.Config, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || !r.configured(&resp.Diagnostics) {
		return
	}
	occupied, err := r.slotOccupied(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Failed to check HA slot", err.Error())
		return
	}
	if occupied {
		resp.Diagnostics.AddError("HA slot already occupied", "Import the existing connection or choose an unused site/HA role.")
		return
	}
	result, err := r.getClient().SiteAddCloudInterconnectPhysicalConnection(ctx, input, r.client.AccountId)
	id := result.GetSite().GetAddCloudInterconnectPhysicalConnection().GetID()
	if id != "" {
		resp.Diagnostics.Append(resp.State.Set(ctx, identityState(plan, id))...)
	}
	if err != nil {
		resp.Diagnostics.AddError("Failed to create physical connection", err.Error())
		return
	}
	if !validID(types.StringValue(id), &resp.Diagnostics) {
		return
	}
	item, err := r.waitForConnection(ctx, id, input)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read created connection", err.Error())
		return
	}
	current := parseConnection(ctx, item, plan.ServiceProviderName, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
	}
}
func (r *connectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || !validID(state.ID, &resp.Diagnostics) || !r.configured(&resp.Diagnostics) {
		return
	}
	item, err := r.find(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read physical connection", err.Error())
		return
	}
	if item == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	current := parseConnection(ctx, item, state.ServiceProviderName, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
	}
}
func (r *connectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || !validID(state.ID, &resp.Diagnostics) {
		return
	}
	desired := buildConfiguredCreate(ctx, req.Config, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || !r.configured(&resp.Diagnostics) {
		return
	}
	existing, err := r.find(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read connection before update", err.Error())
		return
	}
	if existing == nil {
		resp.Diagnostics.AddError("Connection no longer exists", "Refresh state before applying again.")
		return
	}
	if *existing.GetHaRole() != desired.HaRole || !matchesReference(desired.Site.By, desired.Site.Input,
		existing.GetSite().GetID(), existing.GetSite().GetName()) {
		resp.Diagnostics.AddError("Immutable connection identity", "Site and HA role changes require replacement.")
		return
	}
	if *existing.GetEncapsulationMethod() != desired.EncapsulationMethod {
		resp.Diagnostics.AddError("Encapsulation change requires replacement",
			"The API restores incompatible previous tags and cannot clear them through its validated scalar. "+
				"Replace this resource to change encapsulation_method.")
		return
	}
	result, err := r.getClient().SiteUpdateCloudInterconnectPhysicalConnection(ctx,
		buildUpdate(desired, state.ID.ValueString()), r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update physical connection", err.Error())
		return
	}
	if result.GetSite().GetUpdateCloudInterconnectPhysicalConnection().GetID() != state.ID.ValueString() {
		resp.Diagnostics.AddError("Invalid API response", "Update returned a different connection identity.")
		return
	}
	item, err := r.waitForConnection(ctx, state.ID.ValueString(), desired)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read updated connection", err.Error())
		return
	}
	current := parseConnection(ctx, item, plan.ServiceProviderName, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
	}
}
func (r *connectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || !validID(state.ID, &resp.Diagnostics) || !r.configured(&resp.Diagnostics) {
		return
	}
	existing, err := r.find(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read connection before deletion", err.Error())
		return
	}
	if existing == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	result, err := r.getClient().SiteRemoveCloudInterconnectPhysicalConnection(ctx,
		m.RemoveCloudInterconnectPhysicalConnectionInput{ID: state.ID.ValueString()}, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete physical connection", err.Error())
		return
	}
	if result.GetSite().GetRemoveCloudInterconnectPhysicalConnection().GetID() != state.ID.ValueString() {
		resp.Diagnostics.AddError("Invalid API response", "Deletion did not confirm the requested connection identity.")
		return
	}
	resp.State.RemoveResource(ctx)
}
