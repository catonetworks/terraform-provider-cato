package alwaysonrule

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
)

// AlwaysOnRule is the Terraform model for cato_always_on_rule.
// The at object reuses policyrule.Position for GraphQL PolicyRulePositionInput.
// Placement is retained from the plan because addRule does not return this input.
type AlwaysOnRule struct {
	At   types.Object `tfsdk:"at"`   // policyrule.Position
	Rule types.Object `tfsdk:"rule"` // AlwaysOnRuleData
}

// AlwaysOnRuleAttrTypes defines the top-level resource attribute types.
var AlwaysOnRuleAttrTypes = map[string]attr.Type{
	"at":   policyrule.PositionObjectType,
	"rule": AlwaysOnRuleDataObjectType,
}

// AlwaysOnRuleObjectType defines the Terraform object type for AlwaysOnRule.
var AlwaysOnRuleObjectType = types.ObjectType{AttrTypes: AlwaysOnRuleAttrTypes}

// AlwaysOnRuleData represents the nested rule configuration and computed state.
// Device posture profiles reuse the shared idname.Ref model for API object references.
type AlwaysOnRuleData struct {
	Action               types.String `tfsdk:"action"`
	AllowFailOpen        types.Bool   `tfsdk:"allow_fail_open"`
	AllowUserBypass      types.Bool   `tfsdk:"allow_user_bypass"`
	AntiTamperMode       types.String `tfsdk:"anti_tamper_mode"`
	BypassDuration       types.Object `tfsdk:"bypass_duration"` // AlwaysOnRuleBypassDuration
	Description          types.String `tfsdk:"description"`
	DevicePostureProfile types.Set    `tfsdk:"device_posture_profile"` // []idname.Ref
	Enabled              types.Bool   `tfsdk:"enabled"`
	ID                   types.String `tfsdk:"id"`
	Index                types.Int64  `tfsdk:"index"`
	Name                 types.String `tfsdk:"name"`
	Platform             types.Set    `tfsdk:"platform"` // []types.String
	Section              types.Object `tfsdk:"section"`  // AlwaysOnRuleSection
	Source               types.Object `tfsdk:"source"`   // AlwaysOnRuleSource
}

// AlwaysOnRuleDataAttrTypes defines the nested rule attribute types.
var AlwaysOnRuleDataAttrTypes = map[string]attr.Type{
	"action":                 types.StringType,
	"allow_fail_open":        types.BoolType,
	"allow_user_bypass":      types.BoolType,
	"anti_tamper_mode":       types.StringType,
	"bypass_duration":        AlwaysOnRuleBypassDurationObjectType,
	"description":            types.StringType,
	"device_posture_profile": types.SetType{ElemType: idname.RefObjectType},
	"enabled":                types.BoolType,
	"id":                     types.StringType,
	"index":                  types.Int64Type,
	"name":                   types.StringType,
	"platform":               types.SetType{ElemType: types.StringType},
	"section":                AlwaysOnRuleSectionObjectType,
	"source":                 AlwaysOnRuleSourceObjectType,
}

// AlwaysOnRuleDataObjectType defines the Terraform object type for AlwaysOnRuleData.
var AlwaysOnRuleDataObjectType = types.ObjectType{AttrTypes: AlwaysOnRuleDataAttrTypes}

// AlwaysOnRuleBypassDuration represents the disconnection allowed when bypassing Always On.
type AlwaysOnRuleBypassDuration struct {
	Time types.Int64  `tfsdk:"time"`
	Unit types.String `tfsdk:"unit"`
}

// AlwaysOnRuleBypassDurationAttrTypes defines the bypass duration attribute types.
var AlwaysOnRuleBypassDurationAttrTypes = map[string]attr.Type{
	"time": types.Int64Type,
	"unit": types.StringType,
}

// AlwaysOnRuleBypassDurationObjectType defines the Terraform object type for AlwaysOnRuleBypassDuration.
var AlwaysOnRuleBypassDurationObjectType = types.ObjectType{AttrTypes: AlwaysOnRuleBypassDurationAttrTypes}

// AlwaysOnRuleSection represents the policy section returned for the rule.
// It maps GraphQL PolicySectionInfo, which is distinct from an ID/name object reference.
type AlwaysOnRuleSection struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// AlwaysOnRuleSectionAttrTypes defines the policy section attribute types.
var AlwaysOnRuleSectionAttrTypes = map[string]attr.Type{
	"id":   types.StringType,
	"name": types.StringType,
}

// AlwaysOnRuleSectionObjectType defines the Terraform object type for AlwaysOnRuleSection.
var AlwaysOnRuleSectionObjectType = types.ObjectType{AttrTypes: AlwaysOnRuleSectionAttrTypes}

// AlwaysOnRuleSource represents the users and user groups matched by the rule.
// Both collections reuse the shared idname.Ref model for API object references.
type AlwaysOnRuleSource struct {
	Users      types.Set `tfsdk:"users"`       // []idname.Ref
	UserGroups types.Set `tfsdk:"user_groups"` // []idname.Ref
}

// AlwaysOnRuleSourceAttrTypes defines the source matching criteria attribute types.
var AlwaysOnRuleSourceAttrTypes = map[string]attr.Type{
	"users":       types.SetType{ElemType: idname.RefObjectType},
	"user_groups": types.SetType{ElemType: idname.RefObjectType},
}

// AlwaysOnRuleSourceObjectType defines the Terraform object type for AlwaysOnRuleSource.
var AlwaysOnRuleSourceObjectType = types.ObjectType{AttrTypes: AlwaysOnRuleSourceAttrTypes}
