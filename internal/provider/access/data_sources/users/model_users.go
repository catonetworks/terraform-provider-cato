package users

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Lookup struct {
	NameFilter       types.List   `tfsdk:"name_filter"`
	EmailFilter      types.List   `tfsdk:"email_filter"`
	ImportTypeFilter types.String `tfsdk:"import_type_filter"`
	Items            types.List   `tfsdk:"items"`
}

type User struct {
	ID             types.String `tfsdk:"id"`
	Email          types.String `tfsdk:"email"`
	FirstName      types.String `tfsdk:"first_name"`
	LastName       types.String `tfsdk:"last_name"`
	UserStatus     types.String `tfsdk:"user_status"`
	UserImportType types.String `tfsdk:"user_import_type"`
}

type userFilter struct {
	namesMap  map[string]struct{}
	emailsMap map[string]struct{}
}
