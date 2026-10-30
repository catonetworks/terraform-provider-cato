package terminalserverrule

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

// resourceModel contains Terraform values for terminal_server_rule.
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
	AllowedHostIP  types.Object `tfsdk:"allowed_host_ip"`
	Description    types.String `tfsdk:"description"`
	Enabled        types.Bool   `tfsdk:"enabled"`
	ExcludeTraffic types.Set    `tfsdk:"exclude_traffic"`
	ID             types.String `tfsdk:"id"`
	Index          types.Int64  `tfsdk:"index"`
	Name           types.String `tfsdk:"name"`
	Section        types.Object `tfsdk:"section"`
}

var ruleAttrTypes = map[string]attr.Type{
	"allowed_host_ip": idname.RefObjectType,
	"description":     types.StringType,
	"enabled":         types.BoolType,
	"exclude_traffic": types.SetType{ElemType: idname.RefObjectType},
	"id":              types.StringType,
	"index":           types.Int64Type,
	"name":            types.StringType,
	"section":         types.ObjectType{AttrTypes: ruleSectionAttrTypes},
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
