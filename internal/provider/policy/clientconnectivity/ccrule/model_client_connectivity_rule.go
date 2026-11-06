package ccrule

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
)

// ClientConnectivitySourceAttrTypes is the user/users-group source selector.
var ClientConnectivitySourceAttrTypes = map[string]attr.Type{
	"user":        types.SetType{ElemType: idname.RefObjectType},
	"users_group": types.SetType{ElemType: idname.RefObjectType},
}

// ClientConnectivitySourceRangeAttrTypes is one source range entry.
var ClientConnectivitySourceRangeAttrTypes = map[string]attr.Type{
	"global_ip_range": idname.RefObjectType,
}

// ClientConnectivityRuleRuleAttrTypes defines the nested `rule` object for client connectivity rules.
var ClientConnectivityRuleRuleAttrTypes = map[string]attr.Type{
	"id":                types.StringType,
	"name":              types.StringType,
	"description":       types.StringType,
	"enabled":           types.BoolType,
	"action":            types.StringType,
	"confidence_level":  types.StringType,
	"connection_origin": types.ListType{ElemType: types.StringType},
	"country":           types.SetType{ElemType: idname.RefObjectType},
	"device":            types.SetType{ElemType: idname.RefObjectType},
	"platform":          types.ListType{ElemType: types.StringType},
	"source":            types.ObjectType{AttrTypes: ClientConnectivitySourceAttrTypes},
	"source_range":      types.ListType{ElemType: types.ObjectType{AttrTypes: ClientConnectivitySourceRangeAttrTypes}},
}

// ClientConnectivityRuleAttrTypes is the top-level resource attribute map.
var ClientConnectivityRuleAttrTypes = map[string]attr.Type{
	"id":   types.StringType,
	"at":   policyrule.PositionObjectType,
	"rule": types.ObjectType{AttrTypes: ClientConnectivityRuleRuleAttrTypes},
}

// ClientConnectivityRule is the Terraform model for cato_client_connectivity_rule.
type ClientConnectivityRule struct {
	ID   types.String `tfsdk:"id"`
	At   types.Object `tfsdk:"at"`
	Rule types.Object `tfsdk:"rule"`
}

// ClientConnectivityRuleRulePlan maps the nested rule block (ObjectAs).
type ClientConnectivityRuleRulePlan struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	Action           types.String `tfsdk:"action"`
	ConfidenceLevel  types.String `tfsdk:"confidence_level"`
	ConnectionOrigin types.List   `tfsdk:"connection_origin"`
	Country          types.Set    `tfsdk:"country"`
	Device           types.Set    `tfsdk:"device"`
	Platform         types.List   `tfsdk:"platform"`
	Source           types.Object `tfsdk:"source"`
	SourceRange      types.List   `tfsdk:"source_range"`
}

// ClientConnectivitySourcePlan maps the source selector.
type ClientConnectivitySourcePlan struct {
	User       types.Set `tfsdk:"user"`
	UsersGroup types.Set `tfsdk:"users_group"`
}

// ClientConnectivitySourceRangePlan maps one source range entry.
type ClientConnectivitySourceRangePlan struct {
	GlobalIPRange types.Object `tfsdk:"global_ip_range"`
}
