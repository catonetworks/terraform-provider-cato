package admin

import (
	"context"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	attr "github.com/hashicorp/terraform-plugin-framework/attr"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	types "github.com/hashicorp/terraform-plugin-framework/types"
)

// convertCreateInput converts between the Terraform model and SDK values.
//
//nolint:gocyclo
func convertCreateInput(
	ctx context.Context,
	r *adminResource,
	plan Admin,
	diagnostics *diag.Diagnostics,
) (resultCurAccountID string, resultInput cato_models.AddAdminInput) {
	resultCurAccountID = r.client.AccountId
	if !plan.AccountID.IsNull() && !plan.AccountID.IsUnknown() {
		resultCurAccountID = plan.AccountID.ValueString()
	}
	resultInput = cato_models.AddAdminInput{}

	// Set required fields
	if !plan.AccountID.IsNull() && !plan.AccountID.IsUnknown() {
		resultCurAccountID = plan.AccountID.ValueString()
	}

	if !plan.FirstName.IsNull() && !plan.FirstName.IsUnknown() {
		resultInput.FirstName = plan.FirstName.ValueString()
	}

	if !plan.LastName.IsNull() && !plan.LastName.IsUnknown() {
		resultInput.LastName = plan.LastName.ValueString()
	}

	if !plan.Email.IsNull() && !plan.Email.IsUnknown() {
		resultInput.Email = plan.Email.ValueStringPointer()
	}

	// Set optional fields
	if !plan.PasswordNeverExpires.IsNull() && !plan.PasswordNeverExpires.IsUnknown() {
		resultInput.PasswordNeverExpires = plan.PasswordNeverExpires.ValueBool()
	}

	// Handle managed roles
	if !plan.ManagedRoles.IsNull() && !plan.ManagedRoles.IsUnknown() {
		var managedRoles []ManagedRole
		diags := plan.ManagedRoles.ElementsAs(ctx, &managedRoles, false)
		diagnostics.Append(diags...)
		if diagnostics.HasError() {
			return resultCurAccountID, resultInput
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
			resultInput.ManagedRoles = managedRolesInput
		}
	}

	// Handle reseller roles
	if !plan.ResellerRoles.IsNull() && !plan.ResellerRoles.IsUnknown() {
		var resellerRoles []ResellerRole
		diags := plan.ResellerRoles.ElementsAs(ctx, &resellerRoles, false)
		diagnostics.Append(diags...)
		if diagnostics.HasError() {
			return resultCurAccountID, resultInput
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
					diagnostics.Append(diags...)
					if diagnostics.HasError() {
						return resultCurAccountID, resultInput
					}
					roleInput.AllowedAccounts = allowedAccounts
				}

				resellerRolesInput = append(resellerRolesInput, roleInput)
			}
		}
		if len(resellerRolesInput) > 0 {
			resultInput.ResellerRoles = resellerRolesInput
		}
	}
	mfaEnabled := true
	resultInput.MfaEnabled = &mfaEnabled
	return resultCurAccountID, resultInput
}

// convertCreateInput2 converts between the Terraform model and SDK values.
func convertCreateInput2(plan *Admin, curAccountID string, adminID string) {
	plan.ID = types.StringValue(curAccountID + ":" + adminID)
	plan.AdminID = types.StringValue(adminID)
}

// convertCreateInput3 converts between the Terraform model and SDK values.
//
//nolint:gocyclo,funlen
func convertCreateInput3(
	ctx context.Context,
	plan *Admin,
	curAccountID string,
	readAdminResult *cato_go_sdk.Admin,
	diagnostics *diag.Diagnostics,
) {
	if readAdminResult.GetAdmin() == nil {
		diagnostics.AddError(
			"Cato API Admin error",
			"Failed to read admin after creation - no admin data returned",
		)
		return
	}

	adminData := readAdminResult.GetAdmin()

	// Update computed fields from the API response
	plan.MFAEnabled = types.BoolValue(true)
	plan.AccountID = types.StringValue(curAccountID)

	// Handle managed roles with computed names
	if !plan.ManagedRoles.IsNull() && !plan.ManagedRoles.IsUnknown() {
		var planManagedRoles []ManagedRole
		diags := plan.ManagedRoles.ElementsAs(ctx, &planManagedRoles, false)
		diagnostics.Append(diags...)
		if diagnostics.HasError() {
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
				diagnostics.Append(diags...)
				if diagnostics.HasError() {
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
		diagnostics.Append(diags...)
		if diagnostics.HasError() {
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
							diagnostics.Append(diags...)
							if diagnostics.HasError() {
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
				diagnostics.Append(diags...)
				if diagnostics.HasError() {
					return
				}
				plan.ResellerRoles = resellerRolesValue
			}
		}
	}
}

// convertReadState converts between the Terraform model and SDK values.
func convertReadState(ctx context.Context, state *Admin, adminData *cato_models.Admin, diagnostics *diag.Diagnostics) {
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
			diagnostics.Append(diags...)
			if diagnostics.HasError() {
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
					diagnostics.Append(diags...)
					if diagnostics.HasError() {
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
			diagnostics.Append(diags...)
			if diagnostics.HasError() {
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
}

// convertUpdateInput converts between the Terraform model and SDK values.
//
//nolint:gocyclo
func convertUpdateInput(
	ctx context.Context,
	plan Admin,
	state Admin,
	diagnostics *diag.Diagnostics,
) (resultInput cato_models.UpdateAdminInput) {
	if !plan.Email.Equal(state.Email) {
		diagnostics.AddError(
			"Invalid Configuration",
			"Admin email cannot be changed after creation. Current: "+state.Email.ValueString()+", Requested: "+plan.Email.ValueString(),
		)
		return resultInput
	}

	// Build the update input
	resultInput = cato_models.UpdateAdminInput{}

	// Set optional fields that can be updated
	if !plan.FirstName.IsNull() && !plan.FirstName.IsUnknown() {
		firstName := plan.FirstName.ValueString()
		resultInput.FirstName = &firstName
	}

	if !plan.LastName.IsNull() && !plan.LastName.IsUnknown() {
		lastName := plan.LastName.ValueString()
		resultInput.LastName = &lastName
	}

	if !plan.PasswordNeverExpires.IsNull() && !plan.PasswordNeverExpires.IsUnknown() {
		passwordNeverExpires := plan.PasswordNeverExpires.ValueBool()
		resultInput.PasswordNeverExpires = &passwordNeverExpires
	}

	// Handle managed roles
	if !plan.ManagedRoles.IsNull() && !plan.ManagedRoles.IsUnknown() {
		var managedRoles []ManagedRole
		diags := plan.ManagedRoles.ElementsAs(ctx, &managedRoles, false)
		diagnostics.Append(diags...)
		if diagnostics.HasError() {
			return resultInput
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
			resultInput.ManagedRoles = managedRolesInput
		}
	}

	// Handle reseller roles
	if !plan.ResellerRoles.IsNull() && !plan.ResellerRoles.IsUnknown() {
		var resellerRoles []ResellerRole
		diags := plan.ResellerRoles.ElementsAs(ctx, &resellerRoles, false)
		diagnostics.Append(diags...)
		if diagnostics.HasError() {
			return resultInput
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
					diagnostics.Append(diags...)
					if diagnostics.HasError() {
						return resultInput
					}
					roleInput.AllowedAccounts = allowedAccounts
				}

				resellerRolesInput = append(resellerRolesInput, roleInput)
			}
		}
		if len(resellerRolesInput) > 0 {
			resultInput.ResellerRoles = resellerRolesInput
		}
	}
	mfaEnabled := true
	resultInput.MfaEnabled = &mfaEnabled
	return resultInput
}

// convertUpdateInput2 converts between the Terraform model and SDK values.
func convertUpdateInput2(plan *Admin, curAccountID string, curAdminID string) {
	plan.ID = types.StringValue(curAccountID + ":" + curAdminID)
	plan.MFAEnabled = types.BoolValue(true)
	plan.AccountID = types.StringValue(curAccountID)
}

// convertReadValues converts between the Terraform model and SDK values.
func convertReadValues(state *Admin, curAccountID string, curAdminID string, adminData *cato_models.Admin) {
	state.ID = types.StringValue(curAccountID + ":" + curAdminID)
	state.Email = types.StringValue(*adminData.Email)
	state.FirstName = types.StringValue(*adminData.FirstName)
	state.LastName = types.StringValue(*adminData.LastName)

	state.PasswordNeverExpires = types.BoolValue(*adminData.PasswordNeverExpires)
	state.MFAEnabled = types.BoolValue(true)
	state.AccountID = types.StringValue(curAccountID)
	state.AdminID = types.StringValue(curAdminID)
}
