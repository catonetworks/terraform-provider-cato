package mailinglist

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
	_ resource.Resource                = (*mailingListResource)(nil)
	_ resource.ResourceWithConfigure   = (*mailingListResource)(nil)
	_ resource.ResourceWithImportState = (*mailingListResource)(nil)
)

type mailingListResource struct {
	client    *catoclient.Service
	apiClient ResourceClient
}

// NewResource returns the mailing-list resource.
func NewResource() resource.Resource { return &mailingListResource{} }
func (r *mailingListResource) getClient() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	if r.client == nil || r.client.Catov2 == nil {
		return nil
	}
	return &sdkClient{client: r.client}
}
func (r *mailingListResource) configured(diags *diag.Diagnostics) bool {
	if r.client == nil || r.client.AccountId == "" || r.getClient() == nil {
		diags.AddError("Resource not configured", "A Cato client and explicit account ID are required.")
		return false
	}
	return true
}
func (r *mailingListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mailing_list"
}
func (r *mailingListResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *mailingListResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
func (r *mailingListResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := buildCreate(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || !r.configured(&resp.Diagnostics) {
		return
	}
	result, err := r.getClient().NotificationCreateMailingList(ctx, r.client.AccountId, input)
	item := result.GetNotification().GetCreateMailingList().GetMailingList()
	if item.GetID() != "" {
		resp.Diagnostics.Append(resp.State.Set(ctx, identityState(plan, item.GetID()))...)
	}
	if err != nil {
		resp.Diagnostics.AddError("Failed to create mailing list", err.Error())
		return
	}
	if !validID(types.StringValue(item.GetID()), &resp.Diagnostics) {
		return
	}
	current := parseMailingList(ctx, item, item.GetAdmin(), &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
	}
}
func (r *mailingListResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || !validID(state.ID, &resp.Diagnostics) || !r.configured(&resp.Diagnostics) {
		return
	}
	item, err := r.find(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read mailing list", err.Error())
		return
	}
	if item == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	current := parseMailingList(ctx, item, item.GetAdmin(), &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
	}
}
func (r *mailingListResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
	result, err := r.getClient().NotificationUpdateMailingList(ctx, r.client.AccountId, input)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update mailing list", err.Error())
		return
	}
	item := result.GetNotification().GetUpdateMailingList().GetMailingList()
	if item.GetID() != state.ID.ValueString() {
		resp.Diagnostics.AddError("Invalid API response", "Update did not return the requested mailing-list identity.")
		return
	}
	current := parseMailingList(ctx, item, item.GetAdmin(), &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
	}
}
func (r *mailingListResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || !validID(state.ID, &resp.Diagnostics) || !r.configured(&resp.Diagnostics) {
		return
	}
	item, err := r.find(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read mailing list before deletion", err.Error())
		return
	}
	if item == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	result, err := r.getClient().NotificationDeleteMailingList(ctx, r.client.AccountId, listRef(state.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete mailing list", err.Error())
		return
	}
	if result.GetNotification().GetDeleteMailingList().GetMailingList().GetID() != state.ID.ValueString() {
		resp.Diagnostics.AddError("Invalid API response", "Delete did not confirm the requested mailing-list identity.")
		return
	}
	resp.State.RemoveResource(ctx)
}
