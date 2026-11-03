package ssiarule

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
)

// SdpStaticIpAllocationSourceAttrTypes is the single-user source selector.
var SdpStaticIpAllocationSourceAttrTypes = map[string]attr.Type{
	"user": idname.RefObjectType,
}

// SdpStaticIpAllocationRuleRuleAttrTypes defines the nested `rule` object for SDP static IP allocation rules.
var SdpStaticIpAllocationRuleRuleAttrTypes = map[string]attr.Type{
	"id":          types.StringType,
	"name":        types.StringType,
	"description": types.StringType,
	"enabled":     types.BoolType,
	"ip":          types.StringType,
	"source":      types.ObjectType{AttrTypes: SdpStaticIpAllocationSourceAttrTypes},
}

// SdpStaticIpAllocationRuleAttrTypes is the top-level resource attribute map.
var SdpStaticIpAllocationRuleAttrTypes = map[string]attr.Type{
	"id":   types.StringType,
	"at":   policyrule.PositionObjectType,
	"rule": types.ObjectType{AttrTypes: SdpStaticIpAllocationRuleRuleAttrTypes},
}

// SdpStaticIpAllocationRule is the Terraform model for cato_sdp_static_ip_allocation_rule.
type SdpStaticIpAllocationRule struct {
	ID   types.String `tfsdk:"id"`
	At   types.Object `tfsdk:"at"`
	Rule types.Object `tfsdk:"rule"`
}

// SdpStaticIpAllocationRuleRulePlan maps the nested rule block (ObjectAs).
type SdpStaticIpAllocationRuleRulePlan struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	IP          types.String `tfsdk:"ip"`
	Source      types.Object `tfsdk:"source"`
}

// SdpStaticIpAllocationSourcePlan maps the source selector.
type SdpStaticIpAllocationSourcePlan struct {
	User types.Object `tfsdk:"user"`
}
