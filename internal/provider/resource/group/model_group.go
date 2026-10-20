package group

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/groupmembers"
)

type Group struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Members     types.Set    `tfsdk:"members"`
}

// Lookup is the type for the groups data source with filters
type Lookup struct {
	IDFilter   types.List `tfsdk:"id_filter"`
	NameFilter types.List `tfsdk:"name_filter"`
	Items      types.List `tfsdk:"items"`
}

// ItemAttrTypes defines the attribute types for a group item in the lookup results
var ItemAttrTypes = map[string]attr.Type{
	"id":          types.StringType,
	"name":        types.StringType,
	"description": types.StringType,
	"members":     types.SetType{ElemType: groupmembers.MemberObjectType},
}
