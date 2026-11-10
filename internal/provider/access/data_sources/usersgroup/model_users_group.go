package usersgroup

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Lookup struct {
	NameFilter       types.List   `tfsdk:"name_filter"`
	ImportTypeFilter types.String `tfsdk:"import_type_filter"`
	Items            types.List   `tfsdk:"items"`
}

type UsersGroup struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	ImportType   types.String `tfsdk:"import_type"`
	MembersCount types.Int64  `tfsdk:"members_count"`
}

type usersGroupFilter struct {
	namesMap map[string]struct{}
}
