package rpfrule

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
)

// RemotePortFwdPortRangeAttrTypes is a from/to port range.
var RemotePortFwdPortRangeAttrTypes = map[string]attr.Type{
	"from": types.StringType,
	"to":   types.StringType,
}

// RemotePortFwdRemoteIpsAttrTypes is the remote IPs allow/block selector.
var RemotePortFwdRemoteIpsAttrTypes = map[string]attr.Type{
	"global_ip_range": types.SetType{ElemType: idname.RefObjectType},
	"ip":              types.ListType{ElemType: types.StringType},
	"ip_range":        types.ListType{ElemType: types.ObjectType{AttrTypes: RemotePortFwdPortlessRangeAttrTypes}},
	"subnet":          types.ListType{ElemType: types.StringType},
}

// RemotePortFwdPortlessRangeAttrTypes is a from/to IP range.
var RemotePortFwdPortlessRangeAttrTypes = map[string]attr.Type{
	"from": types.StringType,
	"to":   types.StringType,
}

// RemotePortFwdTrackingAttrTypes is the alert tracking selector.
var RemotePortFwdTrackingAttrTypes = map[string]attr.Type{
	"enabled":            types.BoolType,
	"frequency":          types.StringType,
	"mailing_list":       types.SetType{ElemType: idname.RefObjectType},
	"subscription_group": types.SetType{ElemType: idname.RefObjectType},
	"webhook":            types.SetType{ElemType: idname.RefObjectType},
}

// RemotePortFwdRuleRuleAttrTypes defines the nested `rule` object for remote port forwarding rules.
var RemotePortFwdRuleRuleAttrTypes = map[string]attr.Type{
	"id":                  types.StringType,
	"name":                types.StringType,
	"description":         types.StringType,
	"enabled":             types.BoolType,
	"external_ip":         idname.RefObjectType,
	"external_port_range": types.ObjectType{AttrTypes: RemotePortFwdPortRangeAttrTypes},
	"internal_ip":         types.StringType,
	"internal_port_range": types.ObjectType{AttrTypes: RemotePortFwdPortRangeAttrTypes},
	"forward_icmp":        types.BoolType,
	"remote_ips":          types.ObjectType{AttrTypes: RemotePortFwdRemoteIpsAttrTypes},
	"restriction_type":    types.StringType,
	"tracking":            types.ObjectType{AttrTypes: RemotePortFwdTrackingAttrTypes},
}

// RemotePortFwdRuleAttrTypes is the top-level resource attribute map.
var RemotePortFwdRuleAttrTypes = map[string]attr.Type{
	"id":   types.StringType,
	"at":   policyrule.PositionObjectType,
	"rule": types.ObjectType{AttrTypes: RemotePortFwdRuleRuleAttrTypes},
}

// RemotePortFwdRule is the Terraform model for cato_remote_port_fwd_rule.
type RemotePortFwdRule struct {
	ID   types.String `tfsdk:"id"`
	At   types.Object `tfsdk:"at"`
	Rule types.Object `tfsdk:"rule"`
}

// RemotePortFwdRuleRulePlan maps the nested rule block (ObjectAs).
type RemotePortFwdRuleRulePlan struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	Enabled           types.Bool   `tfsdk:"enabled"`
	ExternalIP        types.Object `tfsdk:"external_ip"`
	ExternalPortRange types.Object `tfsdk:"external_port_range"`
	InternalIP        types.String `tfsdk:"internal_ip"`
	InternalPortRange types.Object `tfsdk:"internal_port_range"`
	ForwardICMP       types.Bool   `tfsdk:"forward_icmp"`
	RemoteIPs         types.Object `tfsdk:"remote_ips"`
	RestrictionType   types.String `tfsdk:"restriction_type"`
	Tracking          types.Object `tfsdk:"tracking"`
}

// RemotePortFwdPortRangePlan maps a from/to port range.
type RemotePortFwdPortRangePlan struct {
	From types.String `tfsdk:"from"`
	To   types.String `tfsdk:"to"`
}

// RemotePortFwdRemoteIpsPlan maps the remote IPs selector.
type RemotePortFwdRemoteIpsPlan struct {
	GlobalIPRange types.Set  `tfsdk:"global_ip_range"`
	IP            types.List `tfsdk:"ip"`
	IPRange       types.List `tfsdk:"ip_range"`
	Subnet        types.List `tfsdk:"subnet"`
}

// RemotePortFwdTrackingPlan maps the alert tracking selector.
type RemotePortFwdTrackingPlan struct {
	Enabled           types.Bool   `tfsdk:"enabled"`
	Frequency         types.String `tfsdk:"frequency"`
	MailingList       types.Set    `tfsdk:"mailing_list"`
	SubscriptionGroup types.Set    `tfsdk:"subscription_group"`
	Webhook           types.Set    `tfsdk:"webhook"`
}
