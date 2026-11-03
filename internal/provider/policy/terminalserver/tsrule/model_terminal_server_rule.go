package tsrule

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
)

// TerminalServerRuleRuleAttrTypes defines the nested `rule` object for terminal server rules.
var TerminalServerRuleRuleAttrTypes = map[string]attr.Type{
	"id":              types.StringType,
	"name":            types.StringType,
	"description":     types.StringType,
	"enabled":         types.BoolType,
	"allowed_host_ip": idname.RefObjectType,
	"exclude_traffic": types.SetType{ElemType: idname.RefObjectType},
}

// TerminalServerRuleAttrTypes is the top-level resource attribute map.
var TerminalServerRuleAttrTypes = map[string]attr.Type{
	"id":   types.StringType,
	"at":   policyrule.PositionObjectType,
	"rule": types.ObjectType{AttrTypes: TerminalServerRuleRuleAttrTypes},
}

// TerminalServerRule is the Terraform model for cato_terminal_server_rule.
type TerminalServerRule struct {
	ID   types.String `tfsdk:"id"`
	At   types.Object `tfsdk:"at"`
	Rule types.Object `tfsdk:"rule"`
}

// TerminalServerRuleRulePlan maps the nested rule block (ObjectAs).
type TerminalServerRuleRulePlan struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Enabled        types.Bool   `tfsdk:"enabled"`
	AllowedHostIP  types.Object `tfsdk:"allowed_host_ip"`
	ExcludeTraffic types.Set    `tfsdk:"exclude_traffic"`
}
