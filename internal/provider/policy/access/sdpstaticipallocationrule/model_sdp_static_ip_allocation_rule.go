package sdpstaticipallocationrule

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

// resourceModel contains Terraform values for sdp_static_ip_allocation_rule.
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
	Description types.String `tfsdk:"description"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	ID          types.String `tfsdk:"id"`
	Index       types.Int64  `tfsdk:"index"`
	IP          types.String `tfsdk:"ip"`
	Name        types.String `tfsdk:"name"`
	Section     types.Object `tfsdk:"section"`
	Source      types.Object `tfsdk:"source"`
}

var ruleAttrTypes = map[string]attr.Type{
	"description": types.StringType,
	"enabled":     types.BoolType,
	"id":          types.StringType,
	"index":       types.Int64Type,
	"ip":          types.StringType,
	"name":        types.StringType,
	"section":     types.ObjectType{AttrTypes: ruleSectionAttrTypes},
	"source":      types.ObjectType{AttrTypes: ruleSourceAttrTypes},
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
	User types.Object `tfsdk:"user"`
}

var ruleSourceAttrTypes = map[string]attr.Type{
	"user": idname.RefObjectType,
}
