package splittunnelrule

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

// resourceModel contains Terraform values for split_tunnel_rule.
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
	Action          types.String `tfsdk:"action"`
	Country         types.Set    `tfsdk:"country"`
	Coverage        types.String `tfsdk:"coverage"`
	Description     types.String `tfsdk:"description"`
	Destination     types.Object `tfsdk:"destination"`
	DNSExclusion    types.Object `tfsdk:"dns_exclusion"`
	Enabled         types.Bool   `tfsdk:"enabled"`
	ID              types.String `tfsdk:"id"`
	Index           types.Int64  `tfsdk:"index"`
	Name            types.String `tfsdk:"name"`
	Platform        types.Set    `tfsdk:"platform"`
	RoutingPriority types.String `tfsdk:"routing_priority"`
	Section         types.Object `tfsdk:"section"`
	Source          types.Object `tfsdk:"source"`
	SourceNetwork   types.Object `tfsdk:"source_network"`
}

var ruleAttrTypes = map[string]attr.Type{
	"action":           types.StringType,
	"country":          types.SetType{ElemType: idname.RefObjectType},
	"coverage":         types.StringType,
	"description":      types.StringType,
	"destination":      types.ObjectType{AttrTypes: ruleDestinationAttrTypes},
	"dns_exclusion":    types.ObjectType{AttrTypes: ruleDNSExclusionAttrTypes},
	"enabled":          types.BoolType,
	"id":               types.StringType,
	"index":            types.Int64Type,
	"name":             types.StringType,
	"platform":         types.SetType{ElemType: types.StringType},
	"routing_priority": types.StringType,
	"section":          types.ObjectType{AttrTypes: ruleSectionAttrTypes},
	"source":           types.ObjectType{AttrTypes: ruleSourceAttrTypes},
	"source_network":   types.ObjectType{AttrTypes: ruleSourceNetworkAttrTypes},
}

// RuleDestinationModel contains Terraform values for rule.destination.
type RuleDestinationModel struct {
	Application   types.Set `tfsdk:"application"`
	Domain        types.Set `tfsdk:"domain"`
	Fqdn          types.Set `tfsdk:"fqdn"`
	GlobalIPRange types.Set `tfsdk:"global_ip_range"`
}

var ruleDestinationAttrTypes = map[string]attr.Type{
	"application":     types.SetType{ElemType: idname.RefObjectType},
	"domain":          types.SetType{ElemType: types.StringType},
	"fqdn":            types.SetType{ElemType: types.StringType},
	"global_ip_range": types.SetType{ElemType: idname.RefObjectType},
}

// RuleDNSExclusionModel contains Terraform values for rule.dns_exclusion.
type RuleDNSExclusionModel struct {
	Domain types.Set `tfsdk:"domain"`
}

var ruleDNSExclusionAttrTypes = map[string]attr.Type{
	"domain": types.SetType{ElemType: types.StringType},
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

// RuleSourceModel contains Terraform values for rule.source.
type RuleSourceModel struct {
	User       types.Set `tfsdk:"user"`
	UsersGroup types.Set `tfsdk:"users_group"`
}

var ruleSourceAttrTypes = map[string]attr.Type{
	"user":        types.SetType{ElemType: idname.RefObjectType},
	"users_group": types.SetType{ElemType: idname.RefObjectType},
}

// RuleSourceNetworkModel contains Terraform values for rule.source_network.
type RuleSourceNetworkModel struct {
	SourceNetworkType types.String `tfsdk:"source_network_type"`
}

var ruleSourceNetworkAttrTypes = map[string]attr.Type{
	"source_network_type": types.StringType,
}
