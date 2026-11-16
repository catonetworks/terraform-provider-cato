package mailinglist

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Lookup is the type for the mailing lists data source with filters
type Lookup struct {
	IDFilter   types.List `tfsdk:"id_filter"`
	NameFilter types.List `tfsdk:"name_filter"`
	Items      types.List `tfsdk:"items"`
}

// AdminAttrTypes defines the attribute types for a mailing list admin entry
var AdminAttrTypes = map[string]attr.Type{
	"id":   types.StringType,
	"name": types.StringType,
}

// AdminObjectType is the object type for a mailing list admin entry
var AdminObjectType = types.ObjectType{AttrTypes: AdminAttrTypes}

// ItemAttrTypes defines the attribute types for a mailing list item in the lookup results
var ItemAttrTypes = map[string]attr.Type{
	"id":      types.StringType,
	"name":    types.StringType,
	"type":    types.StringType,
	"address": types.ListType{ElemType: types.StringType},
	"admin":   types.ListType{ElemType: AdminObjectType},
}
