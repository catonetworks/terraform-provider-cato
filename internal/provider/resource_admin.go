package provider

import (
	"context"
	"encoding/json"
	"regexp"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &adminResource{}
	_ resource.ResourceWithConfigure   = &adminResource{}
	_ resource.ResourceWithImportState = &adminResource{}
)

func NewAdminResource() resource.Resource {
	return &adminResource{}
}

type adminResource struct {
	client *catoClientData
}

func (r *adminResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_admin"
}

func (r *adminResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_admin` resource contains the configuration parameters necessary to manage a Cato Networks admin user.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique ID for the admin user resource, consisting of ACCOUNT_ID:ADMIN_ID. Example: `123456:98765`",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"admin_id": schema.StringAttribute{
				Description: "Admin user ID",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"account_id": schema.StringAttribute{
				Description: "Account ID to which the admin belongs.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"email": schema.StringAttribute{
				Description: "Email address of the account",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}$`),
						"Must be a valid email address",
					),
				},
			},
			"first_name": schema.StringAttribute{
				Description: "Account first name",
				Required:    true,
			},
			"last_name": schema.StringAttribute{
				Description: "Account last name",
				Required:    true,
			},
			"password_never_expires": schema.BoolAttribute{
				Description: "Whether the password never expires",
				Optional:    true,
				Computed:    true,
			},
			"mfa_enabled": schema.BoolAttribute{
				Description: "Whether MFA is enabled for this admin, always true for admin users",
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"managed_roles": schema.SetNestedAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Role ID",
							Required:    true,
						},
						"name": schema.StringAttribute{
							Description: "Role name",
							Computed:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
			"reseller_roles": schema.SetNestedAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Role ID",
							Required:    true,
						},
						"name": schema.StringAttribute{
							Description: "Role name",
							Computed:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"allowed_accounts": schema.SetAttribute{
							ElementType: types.StringType,
							Optional:    true,
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.UseStateForUnknown(),
							},
							Description: "List of allowed account IDs",
						},
					},
				},
			},
		},
	}
}

func (r *adminResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoClientData)
}

func (r *adminResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *adminResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	var plan Admin
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	curAccountId := r.client.AccountId
	if !plan.AccountId.IsNull() && !plan.AccountId.IsUnknown() {
		curAccountId = plan.AccountId.ValueString()
	}
	input := cato_models.AddAdminInput{}

	// Set required fields
	if !plan.AccountId.IsNull() && !plan.AccountId.IsUnknown() {
		curAccountId = plan.AccountId.ValueString()
	}

	if !plan.FirstName.IsNull() && !plan.FirstName.IsUnknown() {
		input.FirstName = plan.FirstName.ValueString()
	}

	if !plan.LastName.IsNull() && !plan.LastName.IsUnknown() {
		input.LastName = plan.LastName.ValueString()
	}

	if !plan.Email.IsNull() && !plan.Email.IsUnknown() {
		input.Email = plan.Email.ValueString()
	}

	// Set optional fields
	if !plan.PasswordNeverExpires.IsNull() && !plan.PasswordNeverExpires.IsUnknown() {
		input.PasswordNeverExpires = plan.PasswordNeverExpires.ValueBool()
	}

	// Handle managed roles
	if !plan.ManagedRoles.IsNull() && !plan.ManagedRoles.IsUnknown() {
		var managedRoles []ManagedRole
		diags := plan.ManagedRoles.ElementsAs(ctx, &managedRoles, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		managedRolesInput := make([]*cato_models.UpdateAdminRoleInput, 0, len(managedRoles))
		for _, role := range managedRoles {
			if !role.ID.IsNull() && !role.ID.IsUnknown() {
				managedRolesInput = append(managedRolesInput, &cato_models.UpdateAdminRoleInput{
					Role: &cato_models.UpdateAccountRoleInput{
						ID:   role.ID.ValueString(),
						Name: role.Name.ValueStringPointer(),
					},
				})
			}
		}
		if len(managedRolesInput) > 0 {
			input.ManagedRoles = managedRolesInput
		}
	}

	// Handle reseller roles
	if !plan.ResellerRoles.IsNull() && !plan.ResellerRoles.IsUnknown() {
		var resellerRoles []ResellerRole
		diags := plan.ResellerRoles.ElementsAs(ctx, &resellerRoles, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		resellerRolesInput := make([]*cato_models.UpdateAdminRoleInput, 0, len(resellerRoles))
		for _, role := range resellerRoles {
			if !role.ID.IsNull() && !role.ID.IsUnknown() {
				roleInput := &cato_models.UpdateAdminRoleInput{
					Role: &cato_models.UpdateAccountRoleInput{
						ID:   role.ID.ValueString(),
						Name: role.Name.ValueStringPointer(),
					},
				}

				// Handle allowed accounts if specified
				if !role.AllowedAccounts.IsNull() && !role.AllowedAccounts.IsUnknown() {
					var allowedAccounts []string
					diags := role.AllowedAccounts.ElementsAs(ctx, &allowedAccounts, false)
					resp.Diagnostics.Append(diags...)
					if resp.Diagnostics.HasError() {
						return
					}
					roleInput.AllowedAccounts = allowedAccounts
				}

				resellerRolesInput = append(resellerRolesInput, roleInput)
			}
		}
		if len(resellerRolesInput) > 0 {
			input.ResellerRoles = resellerRolesInput
		}
	}
	mfaEnabled := true
	input.MfaEnabled = mfaEnabled // Always true for admin users

	// Log the request
	tflog.Info(ctx, "Create.admin")
	tflog.Debug(ctx, "Create.Admin.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})

	// Call the API
	addAdminResult, err := r.client.catov2.AdminAddAdmin(ctx, input, curAccountId)
	tflog.Debug(ctx, "Create.Admin.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(addAdminResult),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Cato API Admin error",
			err.Error(),
		)
		return
	}

	// Validate admin creation result
	if addAdminResult.GetAdmin() == nil || addAdminResult.GetAdmin().AddAdmin == nil {
		resp.Diagnostics.AddError(
			"Cato API Admin error",
			"Admin creation failed - no admin data returned",
		)
		return
	}

	// Get the admin ID from the response
	adminID := addAdminResult.GetAdmin().AddAdmin.GetAdminID()
	if adminID == "" {
		resp.Diagnostics.AddError(
			"Cato API Admin error",
			"Admin creation failed - no admin ID returned",
		)
		return
	}

	tflog.Debug(ctx, "Admin created successfully", map[string]interface{}{
		"admin_id": adminID,
	})

	// Update the state with the admin ID
	plan.Id = types.StringValue(curAccountId + ":" + adminID)
	plan.AdminId = types.StringValue(adminID)

	// Read the admin data to populate computed fields
	readAdminResult, err := r.client.catov2.Admin(ctx, curAccountId, adminID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cato API Admin error",
			"Failed to read admin after creation: "+err.Error(),
		)
		return
	}

	if readAdminResult.GetAdmin() == nil {
		resp.Diagnostics.AddError(
			"Cato API Admin error",
			"Failed to read admin after creation - no admin data returned",
		)
		return
	}

	adminData := readAdminResult.GetAdmin()

	// Update computed fields from the API response
	plan.MFAEnabled = types.BoolValue(true)
	plan.AccountId = types.StringValue(curAccountId)

	// Handle managed roles with computed names
	if !plan.ManagedRoles.IsNull() && !plan.ManagedRoles.IsUnknown() {
		var planManagedRoles []ManagedRole
		diags := plan.ManagedRoles.ElementsAs(ctx, &planManagedRoles, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Get the managed roles from the API response
		if managedRoles := adminData.GetManagedRoles(); managedRoles != nil {
			updatedManagedRoles := make([]ManagedRole, 0)
			for _, planRole := range planManagedRoles {
				// Find the corresponding role in the API response
				for _, apiRole := range managedRoles {
					if roleData := apiRole.GetRole(); roleData != nil && roleData.GetID() == planRole.ID.ValueString() {
						updatedManagedRoles = append(updatedManagedRoles, ManagedRole{
							ID:   types.StringValue(roleData.GetID()),
							Name: types.StringValue(roleData.GetName()),
						})
						break
					}
				}
			}
			if len(updatedManagedRoles) > 0 {
				managedRolesValue, diags := types.SetValueFrom(ctx, types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":   types.StringType,
						"name": types.StringType,
					},
				}, updatedManagedRoles)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
				plan.ManagedRoles = managedRolesValue
			}
		}
	}

	// Handle reseller roles with computed names
	if !plan.ResellerRoles.IsNull() && !plan.ResellerRoles.IsUnknown() {
		var planResellerRoles []ResellerRole
		diags := plan.ResellerRoles.ElementsAs(ctx, &planResellerRoles, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Get the reseller roles from the API response
		if resellerRoles := adminData.GetResellerRoles(); resellerRoles != nil {
			updatedResellerRoles := make([]ResellerRole, 0)
			for _, planRole := range planResellerRoles {
				// Find the corresponding role in the API response
				for _, apiRole := range resellerRoles {
					if roleData := apiRole.GetRole(); roleData != nil && roleData.GetID() == planRole.ID.ValueString() {
						resellerRole := ResellerRole{
							ID:   types.StringValue(roleData.GetID()),
							Name: types.StringValue(roleData.GetName()),
						}

						// Handle allowed accounts
						if allowedAccounts := apiRole.GetAllowedAccounts(); allowedAccounts != nil {
							allowedAccountsValue, diags := types.SetValueFrom(ctx, types.StringType, allowedAccounts)
							resp.Diagnostics.Append(diags...)
							if resp.Diagnostics.HasError() {
								return
							}
							resellerRole.AllowedAccounts = allowedAccountsValue
						} else {
							resellerRole.AllowedAccounts = types.SetNull(types.StringType)
						}

						updatedResellerRoles = append(updatedResellerRoles, resellerRole)
						break
					}
				}
			}
			if len(updatedResellerRoles) > 0 {
				resellerRolesValue, diags := types.SetValueFrom(ctx, types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":               types.StringType,
						"name":             types.StringType,
						"allowed_accounts": types.SetType{ElemType: types.StringType},
					},
				}, updatedResellerRoles)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
				plan.ResellerRoles = resellerRolesValue
			}
		}
	}

	// Set the state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *adminResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

	var state Admin
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	idParts := regexp.MustCompile(":").Split(state.Id.ValueString(), 2)
	curAccountId := ""
	curAdminId := ""
	if len(idParts) == 2 {
		curAccountId = idParts[0]
		curAdminId = idParts[1]
	}

	// Fail and return if either curAccountId or curAdminId is empty
	if curAccountId == "" || curAdminId == "" {
		resp.Diagnostics.AddError(
			"Invalid Admin ID",
			"Failed to parse admin ID: either account ID or admin ID is empty.",
		)
		return
	}

	// Log the request
	tflog.Info(ctx, "Read.admin")
	tflog.Debug(ctx, "Read.Admin.request", map[string]interface{}{
		"admin_id":   curAdminId,
		"account_id": curAccountId,
	})

	// Call the API to read the admin
	readAdminResult, err := r.client.catov2.Admins(ctx, curAccountId, nil, nil, nil, nil, []string{curAdminId})
	tflog.Debug(ctx, "Read.Admins.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(readAdminResult),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Cato API Admin error",
			err.Error(),
		)
		return
	}

	var adminData *cato_models.Admin
	if readAdminResult != nil && readAdminResult.Admins != nil && readAdminResult.Admins.Items != nil {
		for _, admin := range readAdminResult.Admins.Items {
			if admin != nil && admin.ID == curAdminId {
				// Convert or map admin (*cato_go_sdk.Admins_Admins_Items) to *cato_models.Admin
				adminBytes, err := json.Marshal(admin)
				if err == nil {
					var mappedAdmin cato_models.Admin
					if err := json.Unmarshal(adminBytes, &mappedAdmin); err == nil {
						adminData = &mappedAdmin
					}
				}
				break
			}
		}
	}

	if adminData != nil {
		matchedAdminJson, _ := json.MarshalIndent(adminData, "", "  ")
		tflog.Debug(ctx, "Matched admin record", map[string]interface{}{
			"response": string(matchedAdminJson),
		})

		// Update the state with data from the API response
		state.Id = types.StringValue(curAccountId + ":" + curAdminId)
		state.Email = types.StringValue(*adminData.Email)
		state.FirstName = types.StringValue(*adminData.FirstName)
		state.LastName = types.StringValue(*adminData.LastName)

		state.PasswordNeverExpires = types.BoolValue(*adminData.PasswordNeverExpires)
		state.MFAEnabled = types.BoolValue(true)
		state.AccountId = types.StringValue(curAccountId)
		state.AdminId = types.StringValue(curAdminId)

	} else {
		resp.Diagnostics.AddError("Admin not found", "The admin with the given ID does not exist.")
		return
	}

	// Handle managed roles
	if managedRoles := adminData.ManagedRoles; managedRoles != nil {
		managedRolesList := make([]ManagedRole, 0)
		for _, role := range managedRoles {
			if roleData := role.Role; roleData != nil {
				managedRolesList = append(managedRolesList, ManagedRole{
					ID:   types.StringValue(roleData.ID),
					Name: types.StringValue(roleData.Name),
				})
			}
		}
		if len(managedRolesList) > 0 {
			managedRolesValue, diags := types.SetValueFrom(ctx, types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"id":   types.StringType,
					"name": types.StringType,
				},
			}, managedRolesList)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			state.ManagedRoles = managedRolesValue
		} else {
			state.ManagedRoles = types.SetNull(types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"id":   types.StringType,
					"name": types.StringType,
				},
			})
		}
	} else {
		state.ManagedRoles = types.SetNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"id":   types.StringType,
				"name": types.StringType,
			},
		})
	}

	// Handle reseller roles
	if resellerRoles := adminData.ResellerRoles; resellerRoles != nil {
		resellerRolesList := make([]ResellerRole, 0)
		for _, role := range resellerRoles {
			if roleData := role.Role; roleData != nil {
				resellerRole := ResellerRole{
					ID:   types.StringValue(roleData.ID),
					Name: types.StringValue(roleData.Name),
				}
				// Handle allowed accounts
				if allowedAccounts := role.AllowedAccounts; allowedAccounts != nil {
					allowedAccountsValue, diags := types.SetValueFrom(ctx, types.StringType, allowedAccounts)
					resp.Diagnostics.Append(diags...)
					if resp.Diagnostics.HasError() {
						return
					}
					resellerRole.AllowedAccounts = allowedAccountsValue
				} else {
					resellerRole.AllowedAccounts = types.SetNull(types.StringType)
				}

				resellerRolesList = append(resellerRolesList, resellerRole)
			}
		}
		if len(resellerRolesList) > 0 {
			resellerRolesValue, diags := types.SetValueFrom(ctx, types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"id":               types.StringType,
					"name":             types.StringType,
					"allowed_accounts": types.SetType{ElemType: types.StringType},
				},
			}, resellerRolesList)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			state.ResellerRoles = resellerRolesValue
		} else {
			state.ResellerRoles = types.SetNull(types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"id":               types.StringType,
					"name":             types.StringType,
					"allowed_accounts": types.SetType{ElemType: types.StringType},
				},
			})
		}
	} else {
		state.ResellerRoles = types.SetNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"id":               types.StringType,
				"name":             types.StringType,
				"allowed_accounts": types.SetType{ElemType: types.StringType},
			},
		})
	}

	// Set the updated state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *adminResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	var plan Admin
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state Admin
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	idParts := regexp.MustCompile(":").Split(state.Id.ValueString(), 2)
	curAccountId := ""
	curAdminId := ""
	if len(idParts) == 2 {
		curAccountId = idParts[0]
		curAdminId = idParts[1]
	}

	// Fail and return if either curAccountId or curAdminId is empty
	if curAccountId == "" || curAdminId == "" {
		resp.Diagnostics.AddError(
			"Invalid Admin ID",
			"Failed to parse admin ID: either account ID or admin ID is empty.",
		)
		return
	}

	// Log the request
	tflog.Info(ctx, "Read.admin")
	tflog.Debug(ctx, "Read.Admin.request", map[string]interface{}{
		"admin_id":   curAdminId,
		"account_id": curAccountId,
	})

	// Check if any immutable fields have changed
	if !plan.Email.Equal(state.Email) {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"Admin email cannot be changed after creation. Current: "+state.Email.ValueString()+", Requested: "+plan.Email.ValueString(),
		)
		return
	}

	// Build the update input
	input := cato_models.UpdateAdminInput{}

	// Set optional fields that can be updated
	if !plan.FirstName.IsNull() && !plan.FirstName.IsUnknown() {
		firstName := plan.FirstName.ValueString()
		input.FirstName = &firstName
	}

	if !plan.LastName.IsNull() && !plan.LastName.IsUnknown() {
		lastName := plan.LastName.ValueString()
		input.LastName = &lastName
	}

	if !plan.PasswordNeverExpires.IsNull() && !plan.PasswordNeverExpires.IsUnknown() {
		passwordNeverExpires := plan.PasswordNeverExpires.ValueBool()
		input.PasswordNeverExpires = &passwordNeverExpires
	}

	// Handle managed roles
	if !plan.ManagedRoles.IsNull() && !plan.ManagedRoles.IsUnknown() {
		var managedRoles []ManagedRole
		diags := plan.ManagedRoles.ElementsAs(ctx, &managedRoles, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		managedRolesInput := make([]*cato_models.UpdateAdminRoleInput, 0, len(managedRoles))
		for _, role := range managedRoles {
			if !role.ID.IsNull() && !role.ID.IsUnknown() {
				managedRolesInput = append(managedRolesInput, &cato_models.UpdateAdminRoleInput{
					Role: &cato_models.UpdateAccountRoleInput{
						ID:   role.ID.ValueString(),
						Name: role.Name.ValueStringPointer(),
					},
				})
			}
		}
		if len(managedRolesInput) > 0 {
			input.ManagedRoles = managedRolesInput
		}
	}

	// Handle reseller roles
	if !plan.ResellerRoles.IsNull() && !plan.ResellerRoles.IsUnknown() {
		var resellerRoles []ResellerRole
		diags := plan.ResellerRoles.ElementsAs(ctx, &resellerRoles, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		resellerRolesInput := make([]*cato_models.UpdateAdminRoleInput, 0, len(resellerRoles))
		for _, role := range resellerRoles {
			if !role.ID.IsNull() && !role.ID.IsUnknown() {
				roleInput := &cato_models.UpdateAdminRoleInput{
					Role: &cato_models.UpdateAccountRoleInput{
						ID:   role.ID.ValueString(),
						Name: role.Name.ValueStringPointer(),
					},
				}

				// Handle allowed accounts if specified
				if !role.AllowedAccounts.IsNull() && !role.AllowedAccounts.IsUnknown() {
					var allowedAccounts []string
					diags := role.AllowedAccounts.ElementsAs(ctx, &allowedAccounts, false)
					resp.Diagnostics.Append(diags...)
					if resp.Diagnostics.HasError() {
						return
					}
					roleInput.AllowedAccounts = allowedAccounts
				}

				resellerRolesInput = append(resellerRolesInput, roleInput)
			}
		}
		if len(resellerRolesInput) > 0 {
			input.ResellerRoles = resellerRolesInput
		}
	}
	mfaEnabled := true
	input.MfaEnabled = &mfaEnabled // Always true for admin users

	// Log the request
	tflog.Info(ctx, "Update.admin")
	tflog.Debug(ctx, "Update.Admin.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(input),
	})

	// Call the API
	updateAdminResult, err := r.client.catov2.AdminUpdateAdmin(ctx, curAdminId, input, curAccountId)
	tflog.Debug(ctx, "Update.Admin.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(updateAdminResult),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Cato API Admin error",
			err.Error(),
		)
		return
	}

	// Validate admin update result
	if updateAdminResult.GetAdmin() == nil || updateAdminResult.GetAdmin().UpdateAdmin == nil {
		resp.Diagnostics.AddError(
			"Cato API Admin error",
			"Admin update failed - no admin data returned",
		)
		return
	}

	tflog.Debug(ctx, "Admin updated successfully", map[string]interface{}{
		"admin_id":   curAdminId,
		"account_id": curAccountId,
	})

	// Update the state with the admin ID (should be the same)
	plan.Id = types.StringValue(curAccountId + ":" + curAdminId)
	plan.MFAEnabled = types.BoolValue(true)
	plan.AccountId = types.StringValue(curAccountId)

	// Set the state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *adminResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	var state Admin
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	idParts := regexp.MustCompile(":").Split(state.Id.ValueString(), 2)
	curAccountId := ""
	curAdminId := ""
	if len(idParts) == 2 {
		curAccountId = idParts[0]
		curAdminId = idParts[1]
	}

	// Fail and return if either curAccountId or curAdminId is empty
	if curAccountId == "" || curAdminId == "" {
		resp.Diagnostics.AddError(
			"Invalid Admin ID",
			"Failed to parse admin ID: either account ID or admin ID is empty.",
		)
		return
	}

	// Log the request
	tflog.Info(ctx, "Read.admin")
	tflog.Debug(ctx, "Read.Admin.request", map[string]interface{}{
		"admin_id":   curAdminId,
		"account_id": curAccountId,
	})

	// Log the request
	tflog.Info(ctx, "Delete.admin")
	tflog.Debug(ctx, "Delete.Admin.request", map[string]interface{}{
		"admin_id": state.Id.ValueString(),
	})

	// Call the API
	removeAdminResult, err := r.client.catov2.AdminRemoveAdmin(ctx, curAdminId, curAccountId)
	tflog.Debug(ctx, "Delete.Admin.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(removeAdminResult),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Cato API Admin error",
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Admin deleted successfully", map[string]interface{}{
		"admin_id": state.Id.ValueString(),
	})
}
