package dynamicipallocationrule

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

// resourceModel contains Terraform values for dynamic_ip_allocation_rule.
type resourceModel struct {
	At   types.Object `tfsdk:"at"`
	Rule types.Object `tfsdk:"rule"`
}

// AtModel contains Terraform values for at.
type AtModel struct {
	Position types.String `tfsdk:"position"`
	Ref      types.String `tfsdk:"ref"`
}

var atAttrTypes = map[string]attr.Type{
	"position": types.StringType,
	"ref":      types.StringType,
}

// RuleModel contains Terraform values for rule.
type RuleModel struct {
	Country              types.Set    `tfsdk:"country"`
	Description          types.String `tfsdk:"description"`
	DevicePostureProfile types.Set    `tfsdk:"device_posture_profile"`
	Enabled              types.Bool   `tfsdk:"enabled"`
	ID                   types.String `tfsdk:"id"`
	Index                types.Int64  `tfsdk:"index"`
	Name                 types.String `tfsdk:"name"`
	Platform             types.Set    `tfsdk:"platform"`
	Range                types.Object `tfsdk:"range"`
	Section              types.Object `tfsdk:"section"`
	Source               types.Object `tfsdk:"source"`
}

var ruleAttrTypes = map[string]attr.Type{
	"country":                types.SetType{ElemType: idname.RefObjectType},
	"description":            types.StringType,
	"device_posture_profile": types.SetType{ElemType: idname.RefObjectType},
	"enabled":                types.BoolType,
	"id":                     types.StringType,
	"index":                  types.Int64Type,
	"name":                   types.StringType,
	"platform":               types.SetType{ElemType: types.StringType},
	"range":                  types.ObjectType{AttrTypes: ruleRangeAttrTypes},
	"section":                types.ObjectType{AttrTypes: ruleSectionAttrTypes},
	"source":                 types.ObjectType{AttrTypes: ruleSourceAttrTypes},
}

// RuleRangeModel contains Terraform values for rule.range.
type RuleRangeModel struct {
	GlobalIPRange types.Object `tfsdk:"global_ip_range"`
}

var ruleRangeAttrTypes = map[string]attr.Type{
	"global_ip_range": idname.RefObjectType,
}

// RuleSectionModel contains Terraform values for rule.section.
type RuleSectionModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

var ruleSectionAttrTypes = map[string]attr.Type{
	"id":   types.StringType,
	"name": types.StringType,
}

// RuleSourceModel contains Terraform values for rule.source.
type RuleSourceModel struct {
	User       types.Set `tfsdk:"user"`
	UsersGroup types.Set `tfsdk:"users_group"`
}

var ruleSourceAttrTypes = map[string]attr.Type{
	"user":        types.SetType{ElemType: idname.RefObjectType},
	"users_group": types.SetType{ElemType: idname.RefObjectType},
}
