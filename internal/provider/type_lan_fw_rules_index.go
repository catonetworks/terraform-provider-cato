package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type LanFwRulesIndex struct {
	SectionData   types.Map `tfsdk:"section_data"`   // map[opaque_key]LanFwSectionData
	NetworkRules  types.Map `tfsdk:"network_rules"`  // map[opaque_key]LanNetworkRule
	FirewallRules types.Map `tfsdk:"firewall_rules"` // map[opaque_key]LanFirewallRule
}

type LanFwSectionData struct {
	ID            types.String `tfsdk:"id"`
	SectionName   types.String `tfsdk:"section_name"`
	SectionIndex  types.Int64  `tfsdk:"section_index"`
	SubPolicyName types.String `tfsdk:"sub_policy_name"`
}

var LanFwSectionDataTypes = map[string]attr.Type{
	"id":              types.StringType,
	"section_name":    types.StringType,
	"section_index":   types.Int64Type,
	"sub_policy_name": types.StringType,
}

type LanNetworkRule struct {
	ID             types.String `tfsdk:"id"`
	RuleType       types.String `tfsdk:"rule_type"`
	RuleName       types.String `tfsdk:"rule_name"`
	SectionName    types.String `tfsdk:"section_name"`
	SectionKey     types.String `tfsdk:"section_key"`
	IndexInSection types.Int64  `tfsdk:"index_in_section"`
}

var LanNetworkRuleTypes = map[string]attr.Type{
	"id":               types.StringType,
	"rule_type":        types.StringType,
	"rule_name":        types.StringType,
	"section_name":     types.StringType,
	"section_key":      types.StringType,
	"index_in_section": types.Int64Type,
}

type LanFirewallRule struct {
	ID               types.String `tfsdk:"id"`
	FirewallRuleName types.String `tfsdk:"firewall_rule_name"`
	NetRuleName      types.String `tfsdk:"net_rule_name"`
	NetRuleKey       types.String `tfsdk:"net_rule_key"`
	IndexInRule      types.Int64  `tfsdk:"index_in_rule"`
}

var LanFirewallRuleTypes = map[string]attr.Type{
	"id":                 types.StringType,
	"firewall_rule_name": types.StringType,
	"net_rule_name":      types.StringType,
	"net_rule_key":       types.StringType,
	"index_in_rule":      types.Int64Type,
}
