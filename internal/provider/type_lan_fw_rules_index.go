package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type LanFwRulesIndex struct {
	SectionData   types.Map `tfsdk:"section_data"`   // map[section_name]LanFwSectionData
	NetworkRules  types.Map `tfsdk:"network_rules"`  // map[rule_name]LanNetworkRule
	FirewallRules types.Map `tfsdk:"firewall_rules"` // map[rule_name]LanFirewallRule
}

type LanFwSectionData struct {
	ID            types.String `tfsdk:"id"`
	SectionIndex  types.Int64  `tfsdk:"section_index"`
	SubPolicyName types.String `tfsdk:"sub_policy_name"`
}

var LanFwSectionDataTypes = map[string]attr.Type{
	"id":              types.StringType,
	"section_index":   types.Int64Type,
	"sub_policy_name": types.StringType,
}

type LanNetworkRule struct {
	ID             types.String `tfsdk:"id"`
	RuleType       types.String `tfsdk:"rule_type"`
	SectionName    types.String `tfsdk:"section_name"`
	IndexInSection types.Int64  `tfsdk:"index_in_section"`
}

var LanNetworkRuleTypes = map[string]attr.Type{
	"id":               types.StringType,
	"rule_type":        types.StringType,
	"section_name":     types.StringType,
	"index_in_section": types.Int64Type,
}

type LanFirewallRule struct {
	ID          types.String `tfsdk:"id"`
	NetRuleName types.String `tfsdk:"net_rule_name"`
	IndexInRule types.Int64  `tfsdk:"index_in_rule"`
}

var LanFirewallRuleTypes = map[string]attr.Type{
	"id":            types.StringType,
	"net_rule_name": types.StringType,
	"index_in_rule": types.Int64Type,
}
