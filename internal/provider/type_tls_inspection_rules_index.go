package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type TLSRulesIndex struct {
	SectionToStartAfterID types.String `tfsdk:"section_to_start_after_id"`
	RuleData              types.Map    `tfsdk:"rule_data"`
	SectionData           types.Map    `tfsdk:"section_data"`
}

type TLSRulesSectionDataIndex struct {
	ID           string
	SectionIndex int64
	SectionName  string
}

type TLSRulesRuleDataIndex struct {
	ID             string
	IndexInSection int64
	SectionName    string
	RuleName       string
	Description    string
	Enabled        bool
}

type TLSRulesSectionItemIndex struct {
	ID           types.String `tfsdk:"id"`
	SectionIndex types.Int64  `tfsdk:"section_index"`
	SectionName  types.String `tfsdk:"section_name"`
}

type TLSRulesRuleItemIndex struct {
	ID             types.String `tfsdk:"id"`
	IndexInSection types.Int64  `tfsdk:"index_in_section"`
	SectionName    types.String `tfsdk:"section_name"`
	RuleName       types.String `tfsdk:"rule_name"`
	Description    types.String `tfsdk:"description"`
	Enabled        types.Bool   `tfsdk:"enabled"`
}
