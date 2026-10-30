package alwaysonindex

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// RulesIndex holds the requested positions of existing Always On policy items.
type RulesIndex struct {
	SectionData types.Map   `tfsdk:"section_data"` // [string]Section
	RuleData    types.Map   `tfsdk:"rule_data"`    // [string]Rule
	Changes     types.Int64 `tfsdk:"changes"`
}

// Section contains the identifier and position of a section_data map item.
type Section struct {
	ID    types.String `tfsdk:"id"`
	Index types.Int64  `tfsdk:"section_index"`
}

// SectionAttrTypes defines the section_data map item attribute types.
var SectionAttrTypes = map[string]attr.Type{
	"id":            types.StringType,
	"section_index": types.Int64Type,
}

// SectionObjectType defines the section_data map item object type.
var SectionObjectType = types.ObjectType{AttrTypes: SectionAttrTypes}

// Rule contains the identifier, position, and parent of a rule_data map item.
type Rule struct {
	ID             types.String `tfsdk:"id"`
	IndexInSection types.Int64  `tfsdk:"index_in_section"`
	SectionName    types.String `tfsdk:"section_name"`
}

// RuleAttrTypes defines the rule_data map item attribute types.
var RuleAttrTypes = map[string]attr.Type{
	"id":               types.StringType,
	"index_in_section": types.Int64Type,
	"section_name":     types.StringType,
}

// RuleObjectType defines the rule_data map item object type.
var RuleObjectType = types.ObjectType{AttrTypes: RuleAttrTypes}

type sectionOrder struct {
	current []nameID
	target  []nameID
}

type ruleOrder struct {
	sectionID string
	current   []nameID
	target    []nameID
}

type nameID struct {
	name, id string
}
