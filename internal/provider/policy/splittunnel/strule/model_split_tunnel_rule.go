package strule

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
)

// SplitTunnelDestinationAttrTypes is the destination selector.
var SplitTunnelDestinationAttrTypes = map[string]attr.Type{
	"application":     types.SetType{ElemType: idname.RefObjectType},
	"domain":          types.ListType{ElemType: types.StringType},
	"fqdn":            types.ListType{ElemType: types.StringType},
	"global_ip_range": types.SetType{ElemType: idname.RefObjectType},
}

// SplitTunnelDNSExclusionAttrTypes is the DNS exclusion selector.
var SplitTunnelDNSExclusionAttrTypes = map[string]attr.Type{
	"domain": types.ListType{ElemType: types.StringType},
}

// SplitTunnelSourceAttrTypes is the user/users-group source selector.
var SplitTunnelSourceAttrTypes = map[string]attr.Type{
	"user":        types.SetType{ElemType: idname.RefObjectType},
	"users_group": types.SetType{ElemType: idname.RefObjectType},
}

// SplitTunnelSourceNetworkAttrTypes is the source network selector.
var SplitTunnelSourceNetworkAttrTypes = map[string]attr.Type{
	"source_network_type": types.StringType,
}

// SplitTunnelRuleRuleAttrTypes defines the nested `rule` object for split tunnel rules.
var SplitTunnelRuleRuleAttrTypes = map[string]attr.Type{
	"id":               types.StringType,
	"name":             types.StringType,
	"description":      types.StringType,
	"enabled":          types.BoolType,
	"action":           types.StringType,
	"coverage":         types.StringType,
	"routing_priority": types.StringType,
	"country":          types.SetType{ElemType: idname.RefObjectType},
	"platform":         types.ListType{ElemType: types.StringType},
	"destination":      types.ObjectType{AttrTypes: SplitTunnelDestinationAttrTypes},
	"dns_exclusion":    types.ObjectType{AttrTypes: SplitTunnelDNSExclusionAttrTypes},
	"source":           types.ObjectType{AttrTypes: SplitTunnelSourceAttrTypes},
	"source_network":   types.ObjectType{AttrTypes: SplitTunnelSourceNetworkAttrTypes},
}

// SplitTunnelRuleAttrTypes is the top-level resource attribute map.
var SplitTunnelRuleAttrTypes = map[string]attr.Type{
	"id":   types.StringType,
	"at":   policyrule.PositionObjectType,
	"rule": types.ObjectType{AttrTypes: SplitTunnelRuleRuleAttrTypes},
}

// SplitTunnelRule is the Terraform model for cato_split_tunnel_rule.
type SplitTunnelRule struct {
	ID   types.String `tfsdk:"id"`
	At   types.Object `tfsdk:"at"`
	Rule types.Object `tfsdk:"rule"`
}

// SplitTunnelRuleRulePlan maps the nested rule block (ObjectAs).
type SplitTunnelRuleRulePlan struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	Enabled         types.Bool   `tfsdk:"enabled"`
	Action          types.String `tfsdk:"action"`
	Coverage        types.String `tfsdk:"coverage"`
	RoutingPriority types.String `tfsdk:"routing_priority"`
	Country         types.Set    `tfsdk:"country"`
	Platform        types.List   `tfsdk:"platform"`
	Destination     types.Object `tfsdk:"destination"`
	DNSExclusion    types.Object `tfsdk:"dns_exclusion"`
	Source          types.Object `tfsdk:"source"`
	SourceNetwork   types.Object `tfsdk:"source_network"`
}

// SplitTunnelDestinationPlan maps the destination selector.
type SplitTunnelDestinationPlan struct {
	Application   types.Set  `tfsdk:"application"`
	Domain        types.List `tfsdk:"domain"`
	Fqdn          types.List `tfsdk:"fqdn"`
	GlobalIPRange types.Set  `tfsdk:"global_ip_range"`
}

// SplitTunnelDNSExclusionPlan maps the DNS exclusion selector.
type SplitTunnelDNSExclusionPlan struct {
	Domain types.List `tfsdk:"domain"`
}

// SplitTunnelSourcePlan maps the source selector.
type SplitTunnelSourcePlan struct {
	User       types.Set `tfsdk:"user"`
	UsersGroup types.Set `tfsdk:"users_group"`
}

// SplitTunnelSourceNetworkPlan maps the source network selector.
type SplitTunnelSourceNetworkPlan struct {
	SourceNetworkType types.String `tfsdk:"source_network_type"`
}
