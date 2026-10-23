package alwaysonrule

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
)

// AlwaysOnRule is the Terraform model for cato_always_on_rule.
// The at object reuses policyrule.Position for GraphQL PolicyRulePositionInput.
type AlwaysOnRule struct {
	At   types.Object `tfsdk:"at"`   // policyrule.Position
	Rule types.Object `tfsdk:"rule"` // Data
}

// AlwaysOnRuleAttrTypes defines the top-level resource attribute types.
var AlwaysOnRuleAttrTypes = map[string]attr.Type{
	"at":   policyrule.PositionObjectType,
	"rule": DataObjectType,
}

// AlwaysOnRuleObjectType defines the Terraform object type for AlwaysOnRule.
var AlwaysOnRuleObjectType = types.ObjectType{AttrTypes: AlwaysOnRuleAttrTypes}

// Data represents the nested rule configuration and computed state.
// Device posture profiles reuse the shared idname.Ref model for API object references.
type Data struct {
	Action               types.String `tfsdk:"action"`
	AllowFailOpen        types.Bool   `tfsdk:"allow_fail_open"`
	AllowUserBypass      types.Bool   `tfsdk:"allow_user_bypass"`
	AntiTamperMode       types.String `tfsdk:"anti_tamper_mode"`
	BypassDuration       types.Object `tfsdk:"bypass_duration"` // BypassDuration
	Description          types.String `tfsdk:"description"`
	DevicePostureProfile types.Set    `tfsdk:"device_posture_profile"` // []idname.Ref
	Enabled              types.Bool   `tfsdk:"enabled"`
	ID                   types.String `tfsdk:"id"`
	Index                types.Int64  `tfsdk:"index"`
	Name                 types.String `tfsdk:"name"`
	Platform             types.Set    `tfsdk:"platform"` // []types.String
	Section              types.Object `tfsdk:"section"`  // Section
	Source               types.Object `tfsdk:"source"`   // Source
}

// DataAttrTypes defines the nested rule attribute types.
var DataAttrTypes = map[string]attr.Type{
	"action":                 types.StringType,
	"allow_fail_open":        types.BoolType,
	"allow_user_bypass":      types.BoolType,
	"anti_tamper_mode":       types.StringType,
	"bypass_duration":        BypassDurationObjectType,
	"description":            types.StringType,
	"device_posture_profile": types.SetType{ElemType: idname.RefObjectType},
	"enabled":                types.BoolType,
	"id":                     types.StringType,
	"index":                  types.Int64Type,
	"name":                   types.StringType,
	"platform":               types.SetType{ElemType: types.StringType},
	"section":                SectionObjectType,
	"source":                 SourceObjectType,
}

// DataObjectType defines the Terraform object type for Data.
var DataObjectType = types.ObjectType{AttrTypes: DataAttrTypes}

// BypassDuration represents the disconnection allowed when bypassing Always On.
type BypassDuration struct {
	Time types.Int64  `tfsdk:"time"`
	Unit types.String `tfsdk:"unit"`
}

// BypassDurationAttrTypes defines the bypass duration attribute types.
var BypassDurationAttrTypes = map[string]attr.Type{
	"time": types.Int64Type,
	"unit": types.StringType,
}

// BypassDurationObjectType defines the Terraform object type for BypassDuration.
var BypassDurationObjectType = types.ObjectType{AttrTypes: BypassDurationAttrTypes}

// Section represents the policy section returned for the rule.
// It maps GraphQL PolicySectionInfo, which is distinct from an ID/name object reference.
type Section struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// SectionAttrTypes defines the policy section attribute types.
var SectionAttrTypes = map[string]attr.Type{
	"id":   types.StringType,
	"name": types.StringType,
}

// SectionObjectType defines the Terraform object type for Section.
var SectionObjectType = types.ObjectType{AttrTypes: SectionAttrTypes}

// Source represents the users and user groups matched by the rule.
// Both collections reuse the shared idname.Ref model for API object references.
type Source struct {
	User       types.Set `tfsdk:"user"`        // []idname.Ref
	UsersGroup types.Set `tfsdk:"users_group"` // []idname.Ref
}

// SourceAttrTypes defines the source matching criteria attribute types.
var SourceAttrTypes = map[string]attr.Type{
	"user":        types.SetType{ElemType: idname.RefObjectType},
	"users_group": types.SetType{ElemType: idname.RefObjectType},
}

// SourceObjectType defines the Terraform object type for Source.
var SourceObjectType = types.ObjectType{AttrTypes: SourceAttrTypes}
