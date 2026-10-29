package subscriptiongroup

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
	_ resource.Resource                = (*subscriptionGroupResource)(nil)
	_ resource.ResourceWithConfigure   = (*subscriptionGroupResource)(nil)
	_ resource.ResourceWithImportState = (*subscriptionGroupResource)(nil)
)

type subscriptionGroupResource struct {
	client    *catoclient.Service
	apiClient ResourceClient
}

// NewResource returns the subscription-group resource.
func NewResource() resource.Resource { return &subscriptionGroupResource{} }
func (r *subscriptionGroupResource) getClient() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	if r.client == nil || r.client.Catov2 == nil {
		return nil
	}
	return &sdkClient{client: r.client}
}
func (r *subscriptionGroupResource) configured(diags *diag.Diagnostics) bool {
	if r.client == nil || r.client.AccountId == "" || r.getClient() == nil {
		diags.AddError("Resource not configured", "A Cato client and explicit account ID are required.")
		return false
	}
	return true
}
func (r *subscriptionGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subscription_group"
}
func (r *subscriptionGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *subscriptionGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
func (r *subscriptionGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := buildCreate(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || !r.configured(&resp.Diagnostics) {
		return
	}
	result, err := r.getClient().NotificationCreateSubscriptionGroup(ctx, r.client.AccountId, input)
	item := result.GetNotification().GetCreateSubscriptionGroup().GetSubscriptionGroup()
	if item.GetID() != "" {
		resp.Diagnostics.Append(resp.State.Set(ctx, identityState(plan, item.GetID()))...)
	}
	if err != nil {
		resp.Diagnostics.AddError("Failed to create subscription group", err.Error())
		return
	}
	if !validID(types.StringValue(item.GetID()), &resp.Diagnostics) {
		return
	}
	current := parseSubscriptionGroup(ctx, item, item.GetMailingList(), item.GetIntegration().GetWebhookList(), &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
	}
}
func (r *subscriptionGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || !validID(state.ID, &resp.Diagnostics) || !r.configured(&resp.Diagnostics) {
		return
	}
	item, err := r.find(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read subscription group", err.Error())
		return
	}
	if item == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	current := parseSubscriptionGroup(ctx, item, item.GetMailingList(), item.GetIntegration().GetWebhookList(), &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
	}
}
func (r *subscriptionGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || !validID(state.ID, &resp.Diagnostics) {
		return
	}
	buildMembers(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || !r.configured(&resp.Diagnostics) {
		return
	}
	existing, readErr := r.find(ctx, state.ID.ValueString())
	if readErr != nil {
		resp.Diagnostics.AddError("Failed to read subscription group before update", readErr.Error())
		return
	}
	if existing == nil {
		resp.Diagnostics.AddError("Subscription group no longer exists", "Refresh state before applying again.")
		return
	}
	currentState := parseSubscriptionGroup(ctx, existing, existing.GetMailingList(),
		existing.GetIntegration().GetWebhookList(), &resp.Diagnostics)
	input := buildUpdate(ctx, plan, currentState, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.getClient().NotificationUpdateSubscriptionGroup(ctx, r.client.AccountId, input)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update subscription group", err.Error())
		return
	}
	item := result.GetNotification().GetUpdateSubscriptionGroup().GetSubscriptionGroup()
	if item.GetID() != state.ID.ValueString() {
		resp.Diagnostics.AddError("Invalid API response", "Update did not return the requested subscription-group identity.")
		return
	}
	current := parseSubscriptionGroup(ctx, item, item.GetMailingList(), item.GetIntegration().GetWebhookList(), &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
	}
}
func (r *subscriptionGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || !validID(state.ID, &resp.Diagnostics) || !r.configured(&resp.Diagnostics) {
		return
	}
	item, err := r.find(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read subscription group before deletion", err.Error())
		return
	}
	if item == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	result, err := r.getClient().NotificationDeleteSubscriptionGroup(ctx, r.client.AccountId, listRef(state.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete subscription group", err.Error())
		return
	}
	if result.GetNotification().GetDeleteSubscriptionGroup().GetSubscriptionGroup().GetID() != state.ID.ValueString() {
		resp.Diagnostics.AddError("Invalid API response", "Delete did not confirm the requested subscription-group identity.")
		return
	}
	resp.State.RemoveResource(ctx)
}
