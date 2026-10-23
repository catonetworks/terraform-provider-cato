package role

import (
	"context"
	"fmt"

	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/iam"
)

type roleResource struct {
	client    *catoclient.Service
	apiClient ResourceClient
}

var _ resource.ResourceWithConfigure = (*roleResource)(nil)
var _ resource.ResourceWithImportState = (*roleResource)(nil)

// NewResource creates the custom role resource.
func NewResource() resource.Resource { return &roleResource{} }
func (r *roleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}
func (r *roleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*catoclient.Service)
	if !ok {
		resp.Diagnostics.AddError("Invalid provider configuration", fmt.Sprintf("Expected *catoclient.Service, got %T", req.ProviderData))
		return
	}
	r.client = client
}
func (r *roleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	account, id, err := iam.ParseID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid role import", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("account_id"), account)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("role_id"), id)...)
}
func (r *roleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	account, err := r.account(plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid role account", err.Error())
		return
	}
	input, d := createInput(ctx, plan)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	response, err := r.resourceAPI().RbacRoleManagementCreateRole(ctx, account, input)
	if err != nil {
		resp.Diagnostics.AddError("Create role failed", err.Error())
		return
	}
	value := iam.CreateRoleRecord(response.GetRbac().GetRoleManagement().GetCreateRole().GetRole())
	state, d := iam.RoleState(ctx, account, value)
	resp.Diagnostics.Append(d...)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	}
	if !resp.Diagnostics.HasError() {
		if err := waitForRole(ctx, r.resourceAPI(), account, value); err != nil {
			resp.Diagnostics.AddError("Role created; read replica not synchronized", err.Error())
		}
	}
}

func (r *roleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	account, id, err := iam.ParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid role identifier", err.Error())
		return
	}
	value, err := iam.LookupRole(ctx, r.resourceAPI(), account, id)
	if err != nil {
		resp.Diagnostics.AddError("Read role failed", err.Error())
		return
	}
	if value == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	state, d := iam.RoleState(ctx, account, value)
	resp.Diagnostics.Append(d...)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	}
}
func (r *roleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state Model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	account, id, err := iam.ParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid role identifier", err.Error())
		return
	}
	target, err := r.account(plan)
	if err != nil || target != account {
		resp.Diagnostics.AddError("Role replacement required", "Changing account_id requires replacement.")
		return
	}
	current, err := iam.LookupRole(ctx, r.resourceAPI(), account, id)
	if err == nil {
		err = mutableRole(current)
	}
	if err != nil {
		resp.Diagnostics.AddError("Cannot update role", err.Error())
		return
	}
	input, d := createInput(ctx, plan)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	response,
		err := r.resourceAPI().RbacRoleManagementUpdateRole(ctx,
		account,
		m.RoleManagementUpdateRoleInput{ID: id,
			Name:        input.Name,
			Description: input.Description,
			Permission:  input.Permission})
	if err != nil {
		resp.Diagnostics.AddError("Update role failed", err.Error())
		return
	}
	value := iam.UpdateRoleRecord(response.GetRbac().GetRoleManagement().GetUpdateRole().GetRole())
	if value != nil && value.ID != id {
		resp.Diagnostics.AddError("Invalid role response", "Updated role identifier differs from state.")
		return
	}
	updated, d := iam.RoleState(ctx, account, value)
	resp.Diagnostics.Append(d...)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &updated)...)
	}
	if !resp.Diagnostics.HasError() {
		if err := waitForRole(ctx, r.resourceAPI(), account, value); err != nil {
			resp.Diagnostics.AddError("Role updated; read replica not synchronized", err.Error())
		}
	}
}

func (r *roleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	account, id, err := iam.ParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid role identifier", err.Error())
		return
	}
	value, err := iam.LookupRole(ctx, r.resourceAPI(), account, id)
	if err != nil {
		resp.Diagnostics.AddError("Read role before deletion failed", err.Error())
		return
	}
	if value == nil {
		return
	}
	if err = mutableRole(value); err != nil {
		resp.Diagnostics.AddError("Cannot delete role", err.Error())
		return
	}
	deleted, err := deleteRoleWithRetry(ctx, r.resourceAPI(), account, id)
	if err != nil {
		resp.Diagnostics.AddError("Delete role failed", err.Error())
		return
	}
	if !deleted {
		resp.Diagnostics.AddError("Invalid role response", "API did not confirm deletion of the requested role.")
	}
}
func (r *roleResource) account(plan Model) (string, error) {
	if r.resourceAPI() == nil {
		return "", fmt.Errorf("configure the Cato provider first")
	}
	fallback := ""
	if r.client != nil {
		fallback = r.client.AccountId
	}
	if plan.AccountID.IsUnknown() {
		plan.AccountID = types.StringNull()
	}
	return iam.Account(plan.AccountID, fallback)
}
