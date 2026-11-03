package diarule

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
)

// DynamicIPAllocationRangeAttrTypes is the allocation range selector.
var DynamicIPAllocationRangeAttrTypes = map[string]attr.Type{
	"global_ip_range": idname.RefObjectType,
}

// DynamicIPAllocationSourceAttrTypes is the user/users-group source selector.
var DynamicIPAllocationSourceAttrTypes = map[string]attr.Type{
	"user":        types.SetType{ElemType: idname.RefObjectType},
	"users_group": types.SetType{ElemType: idname.RefObjectType},
}

// DynamicIPAllocationRuleRuleAttrTypes defines the nested `rule` object for dynamic IP allocation rules.
var DynamicIPAllocationRuleRuleAttrTypes = map[string]attr.Type{
	"id":                     types.StringType,
	"name":                   types.StringType,
	"description":            types.StringType,
	"enabled":                types.BoolType,
	"country":                types.SetType{ElemType: idname.RefObjectType},
	"device_posture_profile": types.SetType{ElemType: idname.RefObjectType},
	"platform":               types.ListType{ElemType: types.StringType},
	"range":                  types.ObjectType{AttrTypes: DynamicIPAllocationRangeAttrTypes},
	"source":                 types.ObjectType{AttrTypes: DynamicIPAllocationSourceAttrTypes},
}

// DynamicIPAllocationRuleAttrTypes is the top-level resource attribute map.
var DynamicIPAllocationRuleAttrTypes = map[string]attr.Type{
	"id":   types.StringType,
	"at":   policyrule.PositionObjectType,
	"rule": types.ObjectType{AttrTypes: DynamicIPAllocationRuleRuleAttrTypes},
}

// DynamicIPAllocationRule is the Terraform model for cato_dynamic_ip_allocation_rule.
type DynamicIPAllocationRule struct {
	ID   types.String `tfsdk:"id"`
	At   types.Object `tfsdk:"at"`
	Rule types.Object `tfsdk:"rule"`
}

// DynamicIPAllocationRuleRulePlan maps the nested rule block (ObjectAs).
type DynamicIPAllocationRuleRulePlan struct {
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	Description          types.String `tfsdk:"description"`
	Enabled              types.Bool   `tfsdk:"enabled"`
	Country              types.Set    `tfsdk:"country"`
	DevicePostureProfile types.Set    `tfsdk:"device_posture_profile"`
	Platform             types.List   `tfsdk:"platform"`
	Range                types.Object `tfsdk:"range"`
	Source               types.Object `tfsdk:"source"`
}

// DynamicIPAllocationRangePlan maps the range selector.
type DynamicIPAllocationRangePlan struct {
	GlobalIPRange types.Object `tfsdk:"global_ip_range"`
}

// DynamicIPAllocationSourcePlan maps the source selector.
type DynamicIPAllocationSourcePlan struct {
	User       types.Set `tfsdk:"user"`
	UsersGroup types.Set `tfsdk:"users_group"`
}
