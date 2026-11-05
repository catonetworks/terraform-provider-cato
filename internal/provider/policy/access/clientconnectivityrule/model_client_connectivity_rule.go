package clientconnectivityrule

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

// resourceModel contains Terraform values for client_connectivity_rule.
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
	Action           types.String `tfsdk:"action"`
	ConfidenceLevel  types.String `tfsdk:"confidence_level"`
	ConnectionOrigin types.Set    `tfsdk:"connection_origin"`
	Country          types.Set    `tfsdk:"country"`
	Description      types.String `tfsdk:"description"`
	Device           types.Set    `tfsdk:"device"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	ID               types.String `tfsdk:"id"`
	Index            types.Int64  `tfsdk:"index"`
	Name             types.String `tfsdk:"name"`
	Platform         types.Set    `tfsdk:"platform"`
	Section          types.Object `tfsdk:"section"`
	Source           types.Object `tfsdk:"source"`
	SourceRange      types.Set    `tfsdk:"source_range"`
}

var ruleAttrTypes = map[string]attr.Type{
	"action":            types.StringType,
	"confidence_level":  types.StringType,
	"connection_origin": types.SetType{ElemType: types.StringType},
	"country":           types.SetType{ElemType: idname.RefObjectType},
	"description":       types.StringType,
	"device":            types.SetType{ElemType: idname.RefObjectType},
	"enabled":           types.BoolType,
	"id":                types.StringType,
	"index":             types.Int64Type,
	"name":              types.StringType,
	"platform":          types.SetType{ElemType: types.StringType},
	"section":           types.ObjectType{AttrTypes: ruleSectionAttrTypes},
	"source":            types.ObjectType{AttrTypes: ruleSourceAttrTypes},
	"source_range":      types.SetType{ElemType: types.ObjectType{AttrTypes: ruleSourceRangeAttrTypes}},
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

// RuleSourceRangeModel contains Terraform values for rule.source_range.
type RuleSourceRangeModel struct {
	GlobalIPRange types.Object `tfsdk:"global_ip_range"`
}

var ruleSourceRangeAttrTypes = map[string]attr.Type{
	"global_ip_range": idname.RefObjectType,
}
