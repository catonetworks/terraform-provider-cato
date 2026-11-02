package user

import (
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
)

type remoteUser interface {
	GetID() string
	GetEmail() *string
	GetFirstName() *string
	GetLastName() *string
	GetDepartment() *string
	GetJobTitle() *string
	GetPhone() *string
	GetDisplayName() string
	GetUserStatus() *cato_models.UserStatus
	GetUserImportType() *cato_models.UserImportType
}

func buildCreate(model resourceModel, diags *diag.Diagnostics) cato_models.CreateUserInput {
	validateModel(model, diags)
	return cato_models.CreateUserInput{
		Email: model.Email.ValueString(), FirstName: model.FirstName.ValueString(), LastName: model.LastName.ValueString(),
		Department:  convert.StringPointerForOptionalInput(model.Department),
		JobTitle:    convert.StringPointerForOptionalInput(model.JobTitle),
		PhoneNumber: convert.StringPointerForOptionalInput(model.PhoneNumber),
	}
}
func buildUpdate(model resourceModel, id string, diags *diag.Diagnostics) cato_models.UpdateUserInput {
	validateModel(model, diags)
	return cato_models.UpdateUserInput{
		ID:          id,
		Email:       model.Email.ValueStringPointer(),
		FirstName:   model.FirstName.ValueStringPointer(),
		LastName:    model.LastName.ValueStringPointer(),
		Department:  convert.StringPointerForOptionalInput(model.Department),
		JobTitle:    convert.StringPointerForOptionalInput(model.JobTitle),
		PhoneNumber: convert.StringPointerForOptionalInput(model.PhoneNumber),
	}
}
func parseUser(value remoteUser, diags *diag.Diagnostics) resourceModel {
	status := statusOf(value)
	if !status.IsValid() {
		diags.AddError("Invalid API response", "The API did not return a known user status.")
	}
	for _, required := range []*string{value.GetEmail(), value.GetFirstName(), value.GetLastName()} {
		if required == nil {
			diags.AddError("Invalid API response", "The API omitted required local-user attributes.")
			break
		}
	}
	return resourceModel{
		ID:             types.StringValue(value.GetID()),
		Email:          parseText(value.GetEmail()),
		FirstName:      parseText(value.GetFirstName()),
		LastName:       parseText(value.GetLastName()),
		Department:     parseText(value.GetDepartment()),
		JobTitle:       parseText(value.GetJobTitle()),
		PhoneNumber:    parseText(value.GetPhone()),
		DisplayName:    types.StringValue(value.GetDisplayName()),
		Enabled:        types.BoolValue(status == cato_models.UserStatusActive || status == cato_models.UserStatusLocked),
		UserStatus:     types.StringValue(string(status)),
		UserImportType: types.StringValue(string(sourceOf(value))),
	}
}
func parseText(value *string) types.String {
	if value == nil {
		return types.StringValue("")
	}
	return types.StringValue(*value)
}

func statusOf(value remoteUser) cato_models.UserStatus {
	if status := value.GetUserStatus(); status != nil {
		return *status
	}
	return ""
}
func sourceOf(value remoteUser) cato_models.UserImportType {
	if source := value.GetUserImportType(); source != nil {
		return *source
	}
	return ""
}

// identityState makes a partial create recoverable without unknown state values.
func identityState(plan resourceModel, id string) resourceModel {
	plan.ID = types.StringValue(id)
	plan.DisplayName = types.StringNull()
	plan.UserStatus = types.StringNull()
	plan.UserImportType = types.StringNull()
	return plan
}
