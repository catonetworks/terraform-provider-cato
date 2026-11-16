package customapplications

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Lookup is the type for the cato_custom_applications data source with filters.
type Lookup struct {
	IDFilter   types.List `tfsdk:"id_filter"`
	NameFilter types.List `tfsdk:"name_filter"`
	Items      types.List `tfsdk:"items"`
}

// ItemAttrTypes defines the attribute types for a custom application item in the lookup results.
// Deeply nested criteria/category fields are intentionally left out - this data source surfaces
// scalar identity fields only.
var ItemAttrTypes = map[string]attr.Type{
	"id":          types.StringType,
	"name":        types.StringType,
	"description": types.StringType,
}
