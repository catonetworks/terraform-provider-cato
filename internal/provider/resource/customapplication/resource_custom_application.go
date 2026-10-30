package customapplication

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
	_ resource.Resource                = (*appResource)(nil)
	_ resource.ResourceWithConfigure   = (*appResource)(nil)
	_ resource.ResourceWithImportState = (*appResource)(nil)
)

type appResource struct {
	client    *catoclient.Service
	apiClient ResourceClient
}

// NewResource returns the custom application resource.
func NewResource() resource.Resource { return &appResource{} }
func (r *appResource) getClient() ResourceClient {
	if r.apiClient != nil {
		return r.apiClient
	}
	if r.client == nil || r.client.Catov2 == nil {
		return nil
	}
	return &sdkClient{client: r.client}
}
func (r *appResource) configured(diags *diag.Diagnostics) bool {
	if r.client == nil || r.client.AccountId == "" || r.getClient() == nil {
		diags.AddError("Resource not configured", "A Cato client and explicit account ID are required.")
		return false
	}
	return true
}
func (r *appResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_application"
}
func (r *appResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *appResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
func (r *appResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var p resourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &p)...)
	if resp.Diagnostics.HasError() {
		return
	}
	in := buildCreate(ctx, p, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || !r.configured(&resp.Diagnostics) {
		return
	}
	result, err := r.getClient().CustomAppDataAddCustomApplication(ctx, r.client.AccountId, in)
	value := result.GetCustomAppData().GetAddCustomApplication().GetCustomApplication()
	id := value.GetID()
	if id != "" {
		resp.Diagnostics.Append(resp.State.Set(ctx, identityState(p, id))...)
	}
	if err != nil {
		resp.Diagnostics.AddError("Failed to create custom application", err.Error())
		return
	}
	if !validID(types.StringValue(id), &resp.Diagnostics) {
		return
	}
	item := mutationRecord(value, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if !categoriesMatch(in.Category, item) {
		resp.Diagnostics.AddError("Category resolution failed",
			"The API did not resolve every category; correct its ID/name reference and apply again.")
		return
	}
	current := parseApplication(ctx, item, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
	}
}
func (r *appResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var p resourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &p)...)
	if resp.Diagnostics.HasError() || !validID(p.ID, &resp.Diagnostics) || !r.configured(&resp.Diagnostics) {
		return
	}
	item, err := r.find(ctx, p.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read custom application", err.Error())
		return
	}
	if item == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	current := parseApplication(ctx, item, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
	}
}
func (r *appResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var p, state resourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &p)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || !validID(state.ID, &resp.Diagnostics) {
		return
	}
	in := buildCreate(ctx, p, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || !r.configured(&resp.Diagnostics) {
		return
	}
	result, err := r.getClient().CustomAppDataUpdateCustomApplication(ctx, r.client.AccountId, buildUpdate(in, state.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to update custom application", err.Error())
		return
	}
	item := mutationRecord(result.GetCustomAppData().GetUpdateCustomApplication().GetCustomApplication(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if item.GetID() != state.ID.ValueString() {
		resp.Diagnostics.AddError("Invalid API response", "Update returned a different application identity.")
		return
	}
	if !categoriesMatch(in.Category, item) {
		resp.Diagnostics.AddError("Category resolution failed", "The API did not resolve every requested category.")
		return
	}
	current := parseApplication(ctx, item, &resp.Diagnostics)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
	}
}
func (r *appResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var p resourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &p)...)
	if resp.Diagnostics.HasError() || !validID(p.ID, &resp.Diagnostics) || !r.configured(&resp.Diagnostics) {
		return
	}
	item, err := r.find(ctx, p.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read application before deletion", err.Error())
		return
	}
	if item == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	result,
		err := r.getClient().CustomAppDataDeleteCustomApplication(ctx,
		r.client.AccountId,
		m.DeleteCustomApplicationInput{CustomApplication: &m.CustomApplicationRefInput{By: m.ObjectRefByID,
			Input: p.ID.ValueString()}})
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete custom application", err.Error())
		return
	}
	if result.GetCustomAppData().GetDeleteCustomApplication().GetCustomApplication().GetID() != p.ID.ValueString() {
		resp.Diagnostics.AddError("Invalid API response", "Delete did not confirm the requested application identity.")
		return
	}
	resp.State.RemoveResource(ctx)
}
