package webhook

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
	_ resource.Resource                = (*webhookResource)(nil)
	_ resource.ResourceWithConfigure   = (*webhookResource)(nil)
	_ resource.ResourceWithImportState = (*webhookResource)(nil)
)

type webhookResource struct {
	client    *catoclient.Service
	apiClient ResourceClient
}

// NewResource returns the webhook resource.
func NewResource() resource.Resource { return &webhookResource{} }

func (r *webhookResource) getClient() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	if r.client == nil || r.client.Catov2 == nil {
		return nil
	}
	return &sdkClient{client: r.client}
}
func (r *webhookResource) configured(diags *diag.Diagnostics) bool {
	if r.client == nil || r.client.AccountId == "" || r.getClient() == nil {
		diags.AddError("Resource not configured", "A Cato client and explicit account ID are required.")
		return false
	}
	return true
}
func (r *webhookResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook"
}
func (r *webhookResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *webhookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// resolveRelatedWebhookRefToID rewrites a related_webhook_ref given by NAME into
// a by-ID ref (the backend's by-NAME relatedWebhookRef resolver is unreliable --
// it reports an existing webhook "not found" -- while by-ID always resolves,
// live-verified on badev4). A ref already by ID, or absent, is returned
// unchanged. Not-found yields a clear diagnostic (the referenced webhook may be
// created later in the same apply -- a convergence re-apply resolves it).
func (r *webhookResource) resolveRelatedWebhookRefToID(ctx context.Context, ref *m.WebhookIntegrationRefInput, diags *diag.Diagnostics) *m.WebhookIntegrationRefInput {
	if ref == nil || ref.By != m.ObjectRefByName {
		return ref
	}
	id, found, err := r.getClient().WebhookIDByName(ctx, r.client.AccountId, ref.Input)
	if err != nil {
		diags.AddError("Failed to resolve related webhook", err.Error())
		return ref
	}
	if !found {
		diags.AddError(
			"Related webhook not found",
			fmt.Sprintf("related_webhook_ref references webhook %q by name, but no such webhook exists on this tenant. "+
				"If it is created later in the same apply, re-apply to converge.", ref.Input),
		)
		return ref
	}
	return &m.WebhookIntegrationRefInput{By: m.ObjectRefByID, Input: id}
}

func (r *webhookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := buildCreate(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || !r.configured(&resp.Diagnostics) {
		return
	}
	input.RelatedWebhookRef = r.resolveRelatedWebhookRefToID(ctx, input.RelatedWebhookRef, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.getClient().CreateWebhook(ctx, r.client.AccountId, input)
	id := ""
	if result != nil {
		id = result.ID
	}
	if id != "" {
		resp.Diagnostics.Append(resp.State.Set(ctx, identityState(plan, id))...)
	}
	if err != nil {
		resp.Diagnostics.AddError("Failed to create webhook", err.Error())
		return
	}
	if !validID(types.StringValue(id), &resp.Diagnostics) {
		return
	}
	current := identityState(plan, id)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
	}
}
func (r *webhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || !validID(state.ID, &resp.Diagnostics) || !r.configured(&resp.Diagnostics) {
		return
	}
	item, err := r.find(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read webhook", err.Error())
		return
	}
	if item == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	current := parseWebhook(ctx, item, state, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
	}
}
func (r *webhookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || !validID(state.ID, &resp.Diagnostics) {
		return
	}
	input := buildUpdate(ctx, plan, state.ID.ValueString(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() || !r.configured(&resp.Diagnostics) {
		return
	}
	input.RelatedWebhookRef = r.resolveRelatedWebhookRefToID(ctx, input.RelatedWebhookRef, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.getClient().UpdateWebhook(ctx, r.client.AccountId, input)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update webhook", err.Error())
		return
	}
	if result == nil || result.ID != state.ID.ValueString() {
		resp.Diagnostics.AddError("Invalid API response", "Update did not return the requested webhook identity.")
		return
	}
	current := identityState(plan, state.ID.ValueString())
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
	}
}
func (r *webhookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || !validID(state.ID, &resp.Diagnostics) || !r.configured(&resp.Diagnostics) {
		return
	}
	item, err := r.find(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read webhook before deletion", err.Error())
		return
	}
	if item == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	result, err := r.getClient().DeleteWebhook(ctx, r.client.AccountId, webhookRef(state.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete webhook", err.Error())
		return
	}
	if result == nil || result.ID != state.ID.ValueString() {
		resp.Diagnostics.AddError("Invalid API response", "Delete did not confirm the requested webhook identity.")
		return
	}
	resp.State.RemoveResource(ctx)
}
