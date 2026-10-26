package iprange

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Range - Graphql type IpAddressRange
type Range struct {
	From types.String `tfsdk:"from" json:"from,omitempty"`
	To   types.String `tfsdk:"to" json:"to,omitempty"`
}

var RangeAttrTypes = map[string]attr.Type{
	"from": types.StringType,
	"to":   types.StringType,
}
var RangeObjectType = types.ObjectType{AttrTypes: RangeAttrTypes}
