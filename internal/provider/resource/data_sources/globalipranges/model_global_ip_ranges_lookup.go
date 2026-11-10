package globalipranges

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Lookup is the type for the cato_global_ip_ranges data source with filters.
type Lookup struct {
	IDFilter   types.List `tfsdk:"id_filter"`
	NameFilter types.List `tfsdk:"name_filter"`
	Items      types.List `tfsdk:"items"`
}

// ItemAttrTypes defines the attribute types for a global IP range item in the lookup results.
var ItemAttrTypes = map[string]attr.Type{
	"id":          types.StringType,
	"name":        types.StringType,
	"description": types.StringType,
	"ip_range":    types.StringType,
	"subnet":      types.StringType,
}
