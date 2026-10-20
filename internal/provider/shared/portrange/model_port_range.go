package portrange

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// PolicyPolicyInternetFirewallPolicyRulesRuleServiceCustomPortRange
// Graphql type PortRange

type PortRange struct {
	From types.String `tfsdk:"from" json:"from,omitempty"`
	To   types.String `tfsdk:"to" json:"to,omitempty"`
}

var PortRangeAttrTypes = map[string]attr.Type{
	"from": types.StringType,
	"to":   types.StringType,
}
var PortRangeObjectType = types.ObjectType{AttrTypes: PortRangeAttrTypes}
