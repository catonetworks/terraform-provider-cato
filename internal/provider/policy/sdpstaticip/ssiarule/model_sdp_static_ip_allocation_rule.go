package ssiarule

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
)

// SdpStaticIPAllocationSourceAttrTypes is the single-user source selector.
var SdpStaticIPAllocationSourceAttrTypes = map[string]attr.Type{
	"user": idname.RefObjectType,
}

// SdpStaticIPAllocationRuleRuleAttrTypes defines the nested `rule` object for SDP static IP allocation rules.
var SdpStaticIPAllocationRuleRuleAttrTypes = map[string]attr.Type{
	"id":          types.StringType,
	"name":        types.StringType,
	"description": types.StringType,
	"enabled":     types.BoolType,
	"ip":          types.StringType,
	"source":      types.ObjectType{AttrTypes: SdpStaticIPAllocationSourceAttrTypes},
}

// SdpStaticIPAllocationRuleAttrTypes is the top-level resource attribute map.
var SdpStaticIPAllocationRuleAttrTypes = map[string]attr.Type{
	"id":   types.StringType,
	"at":   policyrule.PositionObjectType,
	"rule": types.ObjectType{AttrTypes: SdpStaticIPAllocationRuleRuleAttrTypes},
}

// SdpStaticIPAllocationRule is the Terraform model for cato_sdp_static_ip_allocation_rule.
type SdpStaticIPAllocationRule struct {
	ID   types.String `tfsdk:"id"`
	At   types.Object `tfsdk:"at"`
	Rule types.Object `tfsdk:"rule"`
}

// SdpStaticIPAllocationRuleRulePlan maps the nested rule block (ObjectAs).
type SdpStaticIPAllocationRuleRulePlan struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	IP          types.String `tfsdk:"ip"`
	Source      types.Object `tfsdk:"source"`
}

// SdpStaticIPAllocationSourcePlan maps the source selector.
type SdpStaticIPAllocationSourcePlan struct {
	User types.Object `tfsdk:"user"`
}
