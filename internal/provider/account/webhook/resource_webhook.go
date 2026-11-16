package webhook

import (
	"context"
	"errors"
	"fmt"

	cato "github.com/catonetworks/cato-go-sdk"
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
// unchanged. Not-found returns nil (create/update WITHOUT the related link) and
// WARNS rather than erroring, so the referenced webhook being created later in
// the same apply doesn't fail the run -- a convergence re-apply wires it.
func (r *webhookResource) resolveRelatedWebhookRefToID(
	ctx context.Context, ref *m.WebhookIntegrationRefInput, diags *diag.Diagnostics,
) *m.WebhookIntegrationRefInput {
	if ref == nil || ref.By != m.ObjectRefByName {
		return ref
	}
	id, found, err := r.getClient().WebhookIDByName(ctx, r.client.AccountId, ref.Input)
	if err != nil {
		diags.AddError("Failed to resolve related webhook", webhookErrorDetail(err))
		return ref
	}
	if !found {
		// The referenced webhook may simply not exist YET -- a common case is two
		// webhooks in the same for_each where one links the other by name: Terraform
		// can't order for_each siblings by a name-string reference, so the linking one
		// may be created first. Rather than hard-failing the whole apply (which, under
		// parallelism=1, would also prevent the referenced webhook from ever being
		// created and so make the documented "re-apply to converge" impossible), create
		// this webhook WITHOUT the related ref and WARN. The config still carries the
		// ref, so the next apply plans an update that wires it once the target exists.
		diags.AddWarning(
			"Related webhook not resolved yet",
			fmt.Sprintf("related_webhook_ref references webhook %q by name, but it does not exist on this tenant yet. "+
				"Created without the related link; re-apply once %q exists to wire it (expected when both are created in the same apply).",
				ref.Input, ref.Input),
		)
		return nil
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
	hadRelatedRef := input.RelatedWebhookRef != nil
	input.RelatedWebhookRef = r.resolveRelatedWebhookRefToID(ctx, input.RelatedWebhookRef, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	// The related ref was dropped because its target doesn't exist yet (warned in
	// resolveRelatedWebhookRefToID). Record it as unset in state so the NEXT plan
	// shows a diff against the still-present config ref and converges -- otherwise
	// state would claim a link the API never received.
	droppedRelatedRef := hadRelatedRef && input.RelatedWebhookRef == nil
	result, err := r.getClient().CreateWebhook(ctx, r.client.AccountId, input)
	id := ""
	if result != nil {
		id = result.ID
	}
	if id != "" {
		resp.Diagnostics.Append(resp.State.Set(ctx, identityStateRelated(plan, id, droppedRelatedRef))...)
	}
	if err != nil {
		resp.Diagnostics.AddError("Failed to create webhook", webhookErrorDetail(err))
		return
	}
	if !validID(types.StringValue(id), &resp.Diagnostics) {
		return
	}
	current := identityStateRelated(plan, id, droppedRelatedRef)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
	}
}

// identityStateRelated is identityState plus: when the related ref was dropped
// (target webhook not created yet), force related_webhook_ref to null in state so
// the still-present config ref re-diffs and a convergence apply wires it.
func identityStateRelated(plan resourceModel, id string, droppedRelatedRef bool) resourceModel {
	state := identityState(plan, id)
	if droppedRelatedRef {
		state.RelatedWebhookRef = types.ObjectNull(RelatedWebhookRefAttrTypes)
	}
	return state
}
func (r *webhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || !validID(state.ID, &resp.Diagnostics) || !r.configured(&resp.Diagnostics) {
		return
	}
	item, err := r.find(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read webhook", webhookErrorDetail(err))
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
	hadRelatedRef := input.RelatedWebhookRef != nil
	input.RelatedWebhookRef = r.resolveRelatedWebhookRefToID(ctx, input.RelatedWebhookRef, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	droppedRelatedRef := hadRelatedRef && input.RelatedWebhookRef == nil
	result, err := r.getClient().UpdateWebhook(ctx, r.client.AccountId, input)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update webhook", webhookErrorDetail(err))
		return
	}
	if result == nil || result.ID != state.ID.ValueString() {
		resp.Diagnostics.AddError("Invalid API response", "Update did not return the requested webhook identity.")
		return
	}
	current := identityStateRelated(plan, state.ID.ValueString(), droppedRelatedRef)
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
		resp.Diagnostics.AddError("Failed to read webhook before deletion", webhookErrorDetail(err))
		return
	}
	if item == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	result, err := r.getClient().DeleteWebhook(ctx, r.client.AccountId, webhookRef(state.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete webhook", webhookErrorDetail(err))
		return
	}
	if result == nil || result.ID != state.ID.ValueString() {
		resp.Diagnostics.AddError("Invalid API response", "Delete did not confirm the requested webhook identity.")
		return
	}
	resp.State.RemoveResource(ctx)
}

// The SDK attaches GraphQL variables to APIError.RequestBody. Webhook inputs
// contain write-only credentials, so diagnostics must omit that request body.
func webhookErrorDetail(err error) string {
	var apiErr *cato.APIError
	if errors.As(err, &apiErr) {
		return (&cato.APIError{Err: apiErr.Err, TraceID: apiErr.TraceID}).Error()
	}
	return err.Error()
}
