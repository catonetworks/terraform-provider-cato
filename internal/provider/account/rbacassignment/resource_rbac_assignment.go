package rbacassignment

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

type assignmentResource struct {
	client    *catoclient.Service
	apiClient ResourceClient
}

var _ resource.ResourceWithConfigure = (*assignmentResource)(nil)
var _ resource.ResourceWithImportState = (*assignmentResource)(nil)

// NewResource creates the RBAC assignment resource.
func NewResource() resource.Resource { return &assignmentResource{} }
func (r *assignmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rbac_assignment"
}
func (r *assignmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *assignmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	account, id, err := iam.ParseID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid assignment import", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("account_id"), account)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("assignment_id"), id)...)
}
func (r *assignmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	account, err := r.account(plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid assignment account", err.Error())
		return
	}
	input, d := createInput(ctx, plan)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	response, err := r.resourceAPI().RbacAddRbacAssignments(ctx, account, input)
	if err != nil {
		resp.Diagnostics.AddError("Create assignment failed", err.Error())
		return
	}
	items := response.GetRbac().GetAddRbacAssignments().GetItems()
	if len(items) != 1 {
		resp.Diagnostics.AddError("Invalid assignment response", "API must return exactly one created assignment.")
		return
	}
	state, d := assignmentState(ctx, account, createdRecord(items[0]), &plan)
	resp.Diagnostics.Append(d...)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	}
}
func (r *assignmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	account, id, err := iam.ParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid assignment identifier", err.Error())
		return
	}
	value, err := lookupAssignment(ctx, r.resourceAPI(), account, id)
	if err != nil {
		resp.Diagnostics.AddError("Read assignment failed", err.Error())
		return
	}
	if value == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	updated, d := assignmentState(ctx, account, value, &state)
	resp.Diagnostics.Append(d...)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &updated)...)
	}
}
func (r *assignmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state Model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	account, id, err := iam.ParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid assignment identifier", err.Error())
		return
	}
	target, err := r.account(plan)
	if err != nil || target != account || !plan.UserID.Equal(state.UserID) || !plan.GroupID.Equal(state.GroupID) {
		resp.Diagnostics.AddError("Assignment replacement required", "Changing account or principal requires replacement.")
		return
	}
	current, err := lookupAssignment(ctx, r.resourceAPI(), account, id)
	if err == nil {
		err = validateRecord(account, current)
	}
	if err != nil {
		resp.Diagnostics.AddError("Cannot update assignment", err.Error())
		return
	}
	input, d := createInput(ctx, plan)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	response,
		err := r.resourceAPI().RbacUpdateRbacAssignment(ctx,
		account,
		m.UpdateRbacAssignmentInput{RbacAssignmentID: id,
			Roles:                        input.Roles,
			AllowedEntities:              input.AllowedEntities,
			AllowAllSites:                input.AllowAllSites,
			AllowAllUsers:                input.AllowAllUsers,
			Description:                  input.Description,
			AllowViewingUnobfuscatedData: input.AllowViewingUnobfuscatedData})
	if err != nil {
		resp.Diagnostics.AddError("Update assignment failed", err.Error())
		return
	}
	value := updatedRecord(response.GetRbac().GetUpdateRbacAssignment().GetAssignment())
	if value != nil && value.ID != id {
		resp.Diagnostics.AddError("Invalid assignment response", "Updated assignment identifier differs from state.")
		return
	}
	updated, d := assignmentState(ctx, account, value, &plan)
	resp.Diagnostics.Append(d...)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &updated)...)
	}
}
func (r *assignmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	account, id, err := iam.ParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid assignment identifier", err.Error())
		return
	}
	value, err := lookupAssignment(ctx, r.resourceAPI(), account, id)
	if err != nil {
		resp.Diagnostics.AddError("Read assignment before deletion failed", err.Error())
		return
	}
	if value == nil {
		return
	}
	if err = validateRecord(account, value); err != nil {
		resp.Diagnostics.AddError("Cannot delete assignment", err.Error())
		return
	}
	response, err := r.resourceAPI().RbacRemoveRbacAssignment(ctx, account, m.RemoveRbacAssignmentInput{RbacAssignmentID: id})
	if err != nil {
		resp.Diagnostics.AddError("Delete assignment failed", err.Error())
		return
	}
	if response.GetRbac().GetRemoveRbacAssignment().GetRbacAssignmentID() != id {
		resp.Diagnostics.AddError("Invalid assignment response", "API did not confirm deletion of the requested assignment.")
	}
}
func (r *assignmentResource) account(plan Model) (string, error) {
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
