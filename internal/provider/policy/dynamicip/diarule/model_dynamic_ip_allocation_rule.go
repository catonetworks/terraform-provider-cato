package diarule

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
)

// DynamicIpAllocationRangeAttrTypes is the allocation range selector.
var DynamicIpAllocationRangeAttrTypes = map[string]attr.Type{
	"global_ip_range": idname.RefObjectType,
}

// DynamicIpAllocationSourceAttrTypes is the user/users-group source selector.
var DynamicIpAllocationSourceAttrTypes = map[string]attr.Type{
	"user":        types.SetType{ElemType: idname.RefObjectType},
	"users_group": types.SetType{ElemType: idname.RefObjectType},
}

// DynamicIpAllocationRuleRuleAttrTypes defines the nested `rule` object for dynamic IP allocation rules.
var DynamicIpAllocationRuleRuleAttrTypes = map[string]attr.Type{
	"id":                     types.StringType,
	"name":                   types.StringType,
	"description":            types.StringType,
	"enabled":                types.BoolType,
	"country":                types.SetType{ElemType: idname.RefObjectType},
	"device_posture_profile": types.SetType{ElemType: idname.RefObjectType},
	"platform":               types.ListType{ElemType: types.StringType},
	"range":                  types.ObjectType{AttrTypes: DynamicIpAllocationRangeAttrTypes},
	"source":                 types.ObjectType{AttrTypes: DynamicIpAllocationSourceAttrTypes},
}

// DynamicIpAllocationRuleAttrTypes is the top-level resource attribute map.
var DynamicIpAllocationRuleAttrTypes = map[string]attr.Type{
	"id":   types.StringType,
	"at":   policyrule.PositionObjectType,
	"rule": types.ObjectType{AttrTypes: DynamicIpAllocationRuleRuleAttrTypes},
}

// DynamicIpAllocationRule is the Terraform model for cato_dynamic_ip_allocation_rule.
type DynamicIpAllocationRule struct {
	ID   types.String `tfsdk:"id"`
	At   types.Object `tfsdk:"at"`
	Rule types.Object `tfsdk:"rule"`
}

// DynamicIpAllocationRuleRulePlan maps the nested rule block (ObjectAs).
type DynamicIpAllocationRuleRulePlan struct {
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

// DynamicIpAllocationRangePlan maps the range selector.
type DynamicIpAllocationRangePlan struct {
	GlobalIPRange types.Object `tfsdk:"global_ip_range"`
}

// DynamicIpAllocationSourcePlan maps the source selector.
type DynamicIpAllocationSourcePlan struct {
	User       types.Set `tfsdk:"user"`
	UsersGroup types.Set `tfsdk:"users_group"`
}
