package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type IfwRulesIndex struct {
	SectionToStartAfterId types.String `tfsdk:"section_to_start_after_id"`
	RuleData              types.Map    `tfsdk:"rule_data"`
	SectionData           types.Map    `tfsdk:"section_data"`
}

type IfwRulesSectionDataIndex struct {
	Id           string
	SectionIndex int64
	SectionName  string
}

type IfwRulesRuleDataIndex struct {
	Id             string
	IndexInSection int64
	SectionName    string
	RuleName       string
	Description    string
	Enabled        bool
}

type IfwRulesSectionItemIndex struct {
	Id           types.String `tfsdk:"id"`
	SectionIndex types.Int64  `tfsdk:"section_index"`
	SectionName  types.String `tfsdk:"section_name"`
}

type IfwRulesRuleItemIndex struct {
	Id             types.String `tfsdk:"id"`
	IndexInSection types.Int64  `tfsdk:"index_in_section"`
	SectionName    types.String `tfsdk:"section_name"`
	RuleName       types.String `tfsdk:"rule_name"`
	Description    types.String `tfsdk:"description"`
	Enabled        types.Bool   `tfsdk:"enabled"`
}

// type IfwRulesIndexRule struct {
// 	Id             types.String `tfsdk:"id"`
// 	Index          types.Int64  `tfsdk:"index"`
// 	IndexInSection types.Int64  `tfsdk:"index_in_section"`
// 	SectionName    types.String `tfsdk:"section_name"`
// 	SectionId      types.String `tfsdk:"section_id"`
// 	Description    types.String `tfsdk:"description"`
// 	Enabled        types.Bool   `tfsdk:"enabled"`
// 	Name           types.String `tfsdk:"name"`
// 	Properties     types.List   `tfsdk:"properties"`
// }
