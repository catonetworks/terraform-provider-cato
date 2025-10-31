package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type TlsRulesIndex struct {
	SectionToStartAfterId types.String `tfsdk:"section_to_start_after_id"`
	RuleData              types.Map    `tfsdk:"rule_data"`
	SectionData           types.Map    `tfsdk:"section_data"`
}

type TlsRulesSectionDataIndex struct {
	Id           string
	SectionIndex int64
	SectionName  string
}

type TlsRulesRuleDataIndex struct {
	Id             string
	IndexInSection int64
	SectionName    string
	RuleName       string
	Description    string
	Enabled        bool
}

type TlsRulesSectionItemIndex struct {
	Id           types.String `tfsdk:"id"`
	SectionIndex types.Int64  `tfsdk:"section_index"`
	SectionName  types.String `tfsdk:"section_name"`
}

type TlsRulesRuleItemIndex struct {
	Id             types.String `tfsdk:"id"`
	IndexInSection types.Int64  `tfsdk:"index_in_section"`
	SectionName    types.String `tfsdk:"section_name"`
	RuleName       types.String `tfsdk:"rule_name"`
	Description    types.String `tfsdk:"description"`
	Enabled        types.Bool   `tfsdk:"enabled"`
}
