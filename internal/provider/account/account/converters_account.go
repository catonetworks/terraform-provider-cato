package account

import (
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	types "github.com/hashicorp/terraform-plugin-framework/types"
)

// convertCreateInput converts between the Terraform model and SDK values.
func convertCreateInput(plan Account) (resultInput cato_models.AddAccountInput) {
	resultInput = cato_models.AddAccountInput{}

	if !plan.Description.IsNull() || !plan.Description.IsUnknown() {
		resultInput.Description = plan.Description.ValueStringPointer()
	}

	if !plan.Name.IsNull() || !plan.Name.IsUnknown() {
		resultInput.Name = plan.Name.ValueString()
	}

	if !plan.Tenancy.IsNull() || !plan.Tenancy.IsUnknown() {
		tenancyStr := plan.Tenancy.ValueString()
		resultInput.Tenancy = cato_models.AccountTenancy(tenancyStr)
	}

	if !plan.Timezone.IsNull() || !plan.Timezone.IsUnknown() {
		resultInput.Timezone = plan.Timezone.ValueString()
	}

	if !plan.Type.IsNull() || !plan.Type.IsUnknown() {
		resultInput.Type = cato_models.AccountProfileType(plan.Type.ValueString())
	}
	return resultInput
}

// convertCreateInput2 converts between the Terraform model and SDK values.
func convertCreateInput2(
	plan *Account,
	createdAccount *cato_go_sdk.AccountManagementAddAccount_AccountManagement_AddAccount,
) {
	plan.ID = types.StringValue(createdAccount.ID)
	plan.Name = types.StringValue(createdAccount.Name)
	plan.Tenancy = types.StringValue(string(createdAccount.Tenancy))
	plan.Timezone = types.StringValue(createdAccount.TimeZone)
	plan.Type = types.StringValue(string(createdAccount.Type))

	if createdAccount.Description != nil {
		plan.Description = types.StringValue(*createdAccount.Description)
	} else {
		plan.Description = types.StringNull()
	}
}

// convertReadState converts between the Terraform model and SDK values.
func convertReadState(state *Account, readAccountResponse *cato_go_sdk.AccountManagement) {
	if readAccountResponse.AccountManagement.Account.Description != nil {
		state.Description = types.StringValue(*readAccountResponse.AccountManagement.Account.Description)
	} else {
		state.Description = types.StringNull()
	}

	state.Name = types.StringValue(readAccountResponse.AccountManagement.Account.Name)
	state.Tenancy = types.StringValue(string(readAccountResponse.AccountManagement.Account.Tenancy))
	state.Timezone = types.StringValue(readAccountResponse.AccountManagement.Account.TimeZone)
	state.ID = types.StringValue(readAccountResponse.AccountManagement.Account.ID)
	state.Type = types.StringValue(string(readAccountResponse.AccountManagement.Account.Type))
}

// convertUpdateInput converts between the Terraform model and SDK values.
func convertUpdateInput(plan Account) (resultInput cato_models.UpdateAccountInput) {
	resultInput = cato_models.UpdateAccountInput{}

	if !plan.Description.IsNull() || !plan.Description.IsUnknown() {
		resultInput.Description = plan.Description.ValueStringPointer()
	}
	return resultInput
}

// convertUpdateInput2 converts between the Terraform model and SDK values.
func convertUpdateInput2(
	plan *Account,
	updatedAccount *cato_go_sdk.AccountManagementUpdateAccount_AccountManagement_UpdateAccount,
) {
	plan.ID = types.StringValue(updatedAccount.ID)
	plan.Name = types.StringValue(updatedAccount.Name)
	plan.Tenancy = types.StringValue(string(updatedAccount.Tenancy))
	plan.Timezone = types.StringValue(updatedAccount.TimeZone)
	plan.Type = types.StringValue(string(updatedAccount.Type))

	if updatedAccount.Description != nil {
		plan.Description = types.StringValue(*updatedAccount.Description)
	} else {
		plan.Description = types.StringNull()
	}
}
