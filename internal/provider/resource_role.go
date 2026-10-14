package provider

import (
	"context"
	"strings"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var (
	_ resource.Resource                = &roleResource{}
	_ resource.ResourceWithConfigure   = &roleResource{}
	_ resource.ResourceWithImportState = &roleResource{}
)

type roleResource struct{ roleClientConfig }

func NewRoleResource() resource.Resource { return &roleResource{} }
func (r *roleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}
func (r *roleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req.ProviderData, &resp.Diagnostics)
}
func (r *roleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an account-local custom IAM role. Updates replace the complete permission set. " +
			"Predefined roles are immutable and must be read through data sources.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Resource identifier in ACCOUNT_ID:ROLE_ID format.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"role_id": schema.StringAttribute{
				Description: "Stable API role identifier.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"account_id": schema.StringAttribute{
				Description: "Account owning the role. Defaults to the provider account_id; changing it replaces the resource.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"name": schema.StringAttribute{
				Description: "Custom role name.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"description": schema.StringAttribute{
				Description: "Role description. Omit or set to an empty string to clear it.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"permissions": schema.SetNestedAttribute{
				Description: "Complete set of permissions. Each resource may appear once. Use an explicit empty set to grant no permissions.",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"resource": schema.StringAttribute{
							Description: "Resource identifier from cato_permission_catalog, for example Sites.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
						"action": schema.StringAttribute{
							Description: "Granted action: VIEW or EDIT.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("VIEW", "EDIT"),
							},
						},
					},
				},
			},
			"predefined": schema.BoolAttribute{
				Description: "Whether the role is predefined. Managed custom roles are always false.",
				Computed:    true,
			},
			"account_type": schema.StringAttribute{
				Description: "Account type applicable to this role.",
				Computed:    true,
			},
			"is_used_on_external_access": schema.BoolAttribute{
				Description: "Whether external access currently uses the role.",
				Computed:    true,
			},
		},
	}
}

func (r *roleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ":")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid role import identifier", "Expected ACCOUNT_ID:ROLE_ID.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("account_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("role_id"), parts[1])...)
}

func (r *roleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Role
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	accountID, err := effectiveRoleAccount(plan.AccountID, r.client)
	if err != nil {
		resp.Diagnostics.AddError("Invalid role account", err.Error())
		return
	}
	input, d := rolePermissionsInput(ctx, plan.Permissions)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	client := r.getRoleClient()
	if client == nil {
		resp.Diagnostics.AddError("Unconfigured role client", "Configure the Cato provider before creating roles.")
		return
	}
	description := plan.Description.ValueString()
	result, err := client.RbacRoleManagementCreateRole(ctx, accountID, cato_models.RoleManagementCreateRoleInput{
		Name:        plan.Name.ValueString(),
		Description: &description,
		Permission:  input,
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create role", err.Error())
		return
	}
	value := result.GetRbac().GetRoleManagement().GetCreateRole().GetRole()
	if value == nil {
		resp.Diagnostics.AddError("Invalid create-role response", "The API returned no created role.")
		return
	}
	model, d := roleFromAPI(ctx, accountID, value)
	resp.Diagnostics.Append(d...)
	if model.Predefined.ValueBool() {
		resp.Diagnostics.AddError("Invalid create-role response", "The API returned an immutable predefined role.")
		return
	}
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	}
}

func (r *roleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Role
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	accountID, err := effectiveRoleAccount(state.AccountID, r.client)
	if err != nil {
		resp.Diagnostics.AddError("Invalid role account", err.Error())
		return
	}
	model, d := lookupRole(ctx, r.getRoleClient(), accountID, state.RoleID.ValueString())
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	if model == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if model.Predefined.ValueBool() {
		resp.Diagnostics.AddError("Predefined role cannot be managed", "Use the cato_role data source for predefined roles.")
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (r *roleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state Role
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if state.Predefined.ValueBool() {
		resp.Diagnostics.AddError("Predefined role cannot be managed", "Use the cato_role data source for predefined roles.")
		return
	}
	accountID, err := effectiveRoleAccount(state.AccountID, r.client)
	if err != nil {
		resp.Diagnostics.AddError("Invalid role account", err.Error())
		return
	}
	permissions, d := rolePermissionsInput(ctx, plan.Permissions)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	client := r.getRoleClient()
	if client == nil {
		resp.Diagnostics.AddError("Unconfigured role client", "Configure the Cato provider before updating roles.")
		return
	}
	description := plan.Description.ValueString()
	result, err := client.RbacRoleManagementUpdateRole(ctx, accountID, cato_models.RoleManagementUpdateRoleInput{
		ID:          state.RoleID.ValueString(),
		Name:        plan.Name.ValueString(),
		Description: &description,
		Permission:  permissions,
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to update role", err.Error())
		return
	}
	value := result.GetRbac().GetRoleManagement().GetUpdateRole().GetRole()
	if value == nil {
		resp.Diagnostics.AddError("Invalid update-role response", "The API returned no updated role.")
		return
	}
	model, d := roleFromAPI(ctx, accountID, value)
	resp.Diagnostics.Append(d...)
	if model.RoleID.ValueString() != state.RoleID.ValueString() || model.Predefined.ValueBool() {
		resp.Diagnostics.AddError("Invalid update-role response", "The API returned a different or predefined role.")
		return
	}
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	}
}

func (r *roleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Role
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if state.Predefined.ValueBool() {
		resp.Diagnostics.AddError("Predefined role cannot be managed", "Predefined roles cannot be deleted.")
		return
	}
	accountID, err := effectiveRoleAccount(state.AccountID, r.client)
	if err != nil {
		resp.Diagnostics.AddError("Invalid role account", err.Error())
		return
	}
	client := r.getRoleClient()
	model, d := lookupRole(ctx, client, accountID, state.RoleID.ValueString())
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() || model == nil {
		return
	}
	if model.Predefined.ValueBool() {
		resp.Diagnostics.AddError("Predefined role cannot be managed", "Predefined roles cannot be deleted.")
		return
	}
	result, err := client.RbacRoleManagementDeleteRole(ctx, accountID, cato_models.RoleManagementDeleteRoleInput{
		ID: state.RoleID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete role", err.Error())
		return
	}
	if result.GetRbac().GetRoleManagement().GetDeleteRole() == nil ||
		result.GetRbac().GetRoleManagement().GetDeleteRole().GetID() != state.RoleID.ValueString() {
		resp.Diagnostics.AddError("Invalid delete-role response", "The API did not confirm deletion of the requested role.")
	}
}
