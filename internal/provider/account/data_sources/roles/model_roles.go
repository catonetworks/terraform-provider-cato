package roles

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Model contains query options and sorted role results.
type Model struct {
	ID            types.String `tfsdk:"id"`
	AccountID     types.String `tfsdk:"account_id"`
	Filter        types.Object `tfsdk:"filter"`
	SortDirection types.String `tfsdk:"sort_direction"`
	Items         types.List   `tfsdk:"items"`
	Total         types.Int64  `tfsdk:"total"`
}

var stringFilterAttrTypes = map[string]attr.Type{
	"eq":  types.StringType,
	"neq": types.StringType,
	"in":  types.SetType{ElemType: types.StringType},
	"nin": types.SetType{ElemType: types.StringType},
}
var booleanFilterAttrTypes = map[string]attr.Type{"eq": types.BoolType, "neq": types.BoolType}
var filterAttrTypes = map[string]attr.Type{
	"id": types.ObjectType{AttrTypes: stringFilterAttrTypes}, "name": types.ObjectType{AttrTypes: stringFilterAttrTypes},
	"predefined": types.ObjectType{AttrTypes: booleanFilterAttrTypes},
}
