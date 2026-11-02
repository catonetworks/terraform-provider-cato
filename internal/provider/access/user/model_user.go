package user

import "github.com/hashicorp/terraform-plugin-framework/types"

type resourceModel struct {
	Department     types.String `tfsdk:"department"`
	DisplayName    types.String `tfsdk:"display_name"`
	Email          types.String `tfsdk:"email"`
	Enabled        types.Bool   `tfsdk:"enabled"`
	FirstName      types.String `tfsdk:"first_name"`
	ID             types.String `tfsdk:"id"`
	JobTitle       types.String `tfsdk:"job_title"`
	LastName       types.String `tfsdk:"last_name"`
	PhoneNumber    types.String `tfsdk:"phone_number"`
	UserImportType types.String `tfsdk:"user_import_type"`
	UserStatus     types.String `tfsdk:"user_status"`
}
