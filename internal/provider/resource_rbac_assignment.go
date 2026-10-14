package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &rbacAssignmentResource{}
	_ resource.ResourceWithConfigure   = &rbacAssignmentResource{}
	_ resource.ResourceWithImportState = &rbacAssignmentResource{}
)

type rbacAssignmentResource struct {
	client           *catoClientData
	assignmentClient RBACAssignmentClient
}

func NewRBACAssignmentResource() resource.Resource { return &rbacAssignmentResource{} }
func (r *rbacAssignmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rbac_assignment"
}
func (r *rbacAssignmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*catoClientData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider configuration",
			fmt.Sprintf("Expected Cato provider configuration, received %T.", req.ProviderData))
		return
	}
	r.client = client
}
func (r *rbacAssignmentResource) getAssignmentClient() RBACAssignmentClient {
	if r.assignmentClient != nil {
		return r.assignmentClient
	}
	if r.client == nil || r.client.catov2 == nil {
		return nil
	}
	return r.client.catov2
}
func (r *rbacAssignmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	stable := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		Description: "Manages one RBAC assignment for a user or users group through the beta RBAC APIs. " +
			"Updates replace all roles and scope. Legacy ADMIN assignments cannot be managed. " +
			"Import with ACCOUNT_ID:ASSIGNMENT_ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Terraform identifier in ACCOUNT_ID:ASSIGNMENT_ID format.", Computed: true, PlanModifiers: stable,
			},
			"assignment_id": schema.StringAttribute{
				Description: "Stable API assignment identifier.", Computed: true, PlanModifiers: stable,
			},
			"account_id": schema.StringAttribute{
				Description: "Owning account. Defaults to the provider account_id. Changing it replaces the assignment.",
				Optional:    true, Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace(), stringplanmodifier.UseStateForUnknown()},
				Validators:    []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"user_id": schema.StringAttribute{
				Description: "User ID to assign. Specify exactly one of user_id or group_id. " +
					"Changing the principal replaces the assignment.",
				Optional: true, PlanModifiers: replace,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1), stringvalidator.ExactlyOneOf(path.MatchRoot("group_id")),
				},
			},
			"group_id": schema.StringAttribute{
				Description: "Users group ID to assign. Changing the principal replaces the assignment.",
				Optional:    true, PlanModifiers: replace,
				Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"role_ids": schema.SetAttribute{
				Description: "Complete nonempty set of role IDs to assign, including custom or predefined roles.",
				Required:    true, ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1), setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
				},
			},
			"allowed_entities": schema.SetNestedAttribute{
				Description: "Complete individual entity scope. Use [] for no individual entities; " +
					"use SYSTEMGROUP entries for account-wide access. Each type/ID pair may appear once.",
				Required: true,
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Description: "Stable entity ID.", Required: true,
						Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
					},
					"type": schema.StringAttribute{
						Description: "Entity type.", Required: true,
						Validators: []validator.String{stringvalidator.OneOf(
							"SITE", "USERSGROUP", "GROUP", "SYSTEMGROUP", "WAN_FIREWALL_POLICY", "INTERNET_FIREWALL_POLICY",
						)},
					},
					"action": schema.StringAttribute{
						Description: "Access to the entity: VIEW or EDIT.", Required: true,
						Validators: []validator.String{stringvalidator.OneOf("VIEW", "EDIT")},
					},
				}},
			},
			"allow_all_sites": schema.StringAttribute{
				Description: "Effective access to all sites, derived from SYSTEMGROUP scope: NONE, VIEW or EDIT.", Computed: true,
			},
			"allow_all_users": schema.StringAttribute{
				Description: "Effective access to all users, derived from SYSTEMGROUP scope: NONE, VIEW or EDIT.", Computed: true,
			},
			"allow_viewing_unobfuscated_data": schema.BoolAttribute{
				Description: "Whether this assignment may view unobfuscated data. " +
					"Omit to leave the API preference unchanged on update. When account obfuscation is DISABLED, " +
					"the API hides the stored preference and refresh preserves the configured value. " +
					"Import cannot recover that hidden preference.",
				Optional: true,
			},
			"description": schema.StringAttribute{
				Description: "Assignment description. Omit or set an empty string to clear it.",
				Optional:    true, Computed: true, Default: stringdefault.StaticString(""),
			},
			"assignment_type": schema.StringAttribute{
				Description: "API assignment type: ADMIN_USER or GROUP.", Computed: true,
			},
			"data_obfuscation_status": schema.StringAttribute{
				Description: "Effective API status: OBFUSCATED, UNOBFUSCATED or DISABLED.", Computed: true,
			},
		},
	}
}
func (r *rbacAssignmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ":")
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		resp.Diagnostics.AddError("Invalid assignment import identifier", "Expected ACCOUNT_ID:ASSIGNMENT_ID.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("account_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("assignment_id"), parts[1])...)
}
func (r *rbacAssignmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RBACAssignment
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	account, err := effectiveRoleAccount(plan.AccountID, r.client)
	if err != nil {
		resp.Diagnostics.AddError("Invalid assignment account", err.Error())
		return
	}
	input, d := assignmentInput(ctx, plan, true)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	client := r.getAssignmentClient()
	if client == nil {
		resp.Diagnostics.AddError("Unconfigured assignment client", "Configure the Cato provider before creating assignments.")
		return
	}
	value, err := client.AddAssignment(ctx, account, input)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create assignment", err.Error())
		return
	}
	model, d := assignmentFromAPI(account, value, &plan)
	resp.Diagnostics.Append(d...)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	}
}
func (r *rbacAssignmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RBACAssignment
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	account, err := effectiveRoleAccount(state.AccountID, r.client)
	if err != nil {
		resp.Diagnostics.AddError("Invalid assignment account", err.Error())
		return
	}
	value, err := lookupAssignment(ctx, r.getAssignmentClient(), account, state.AssignmentID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read assignment", err.Error())
		return
	}
	if value == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	model, d := assignmentFromAPI(account, value, &state)
	resp.Diagnostics.Append(d...)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	}
}
func (r *rbacAssignmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state RBACAssignment
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if state.AssignmentType.ValueString() == "ADMIN" {
		resp.Diagnostics.AddError("Unsupported assignment type", "Legacy ADMIN assignments cannot be managed.")
		return
	}
	account, err := effectiveRoleAccount(state.AccountID, r.client)
	if err != nil {
		resp.Diagnostics.AddError("Invalid assignment account", err.Error())
		return
	}
	if plan.UserID.ValueString() != state.UserID.ValueString() || plan.GroupID.ValueString() != state.GroupID.ValueString() ||
		(!plan.AccountID.IsUnknown() && plan.AccountID.ValueString() != account) {
		resp.Diagnostics.AddError("Assignment replacement required", "Changing the account or principal requires replacement.")
		return
	}
	plan.AssignmentID = state.AssignmentID
	input, d := assignmentInput(ctx, plan, false)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	client := r.getAssignmentClient()
	if client == nil {
		resp.Diagnostics.AddError("Unconfigured assignment client", "Configure the Cato provider before updating assignments.")
		return
	}
	value, err := client.UpdateAssignment(ctx, account, input)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update assignment", err.Error())
		return
	}
	model, d := assignmentFromAPI(account, value, &plan)
	resp.Diagnostics.Append(d...)
	if model.AssignmentID.ValueString() != state.AssignmentID.ValueString() {
		resp.Diagnostics.AddError("Invalid update-assignment response", "The API returned a different assignment identifier.")
	}
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	}
}
func (r *rbacAssignmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RBACAssignment
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	account, err := effectiveRoleAccount(state.AccountID, r.client)
	if err != nil {
		resp.Diagnostics.AddError("Invalid assignment account", err.Error())
		return
	}
	client := r.getAssignmentClient()
	value, err := lookupAssignment(ctx, client, account, state.AssignmentID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read assignment before deletion", err.Error())
		return
	}
	if value == nil {
		return
	}
	_, d := assignmentFromAPI(account, value, &state)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := client.RemoveAssignment(ctx, account, state.AssignmentID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Failed to delete assignment", err.Error())
	}
}
