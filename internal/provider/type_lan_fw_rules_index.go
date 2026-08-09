package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type LanFwRulesIndex struct {
	SectionData types.Map `tfsdk:"section_data"` // map[section_name]LanFwSectionData
	RuleData    types.Map `tfsdk:"rule_data"`    // map[rule_name]LanFwRuleData
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

type LanFwRuleData struct {
	ID             types.String `tfsdk:"id"`
	RuleType       types.String `tfsdk:"rule_type"`
	SectionName    types.String `tfsdk:"section_name"`
	IndexInSection types.Int64  `tfsdk:"index_in_section"`
}

var LanFwRuleDataTypes = map[string]attr.Type{
	"id":               types.StringType,
	"rule_type":        types.StringType,
	"section_name":     types.StringType,
	"index_in_section": types.Int64Type,
}
