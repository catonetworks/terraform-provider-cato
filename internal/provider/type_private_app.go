package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PrivateAppModel struct {
	ID                 types.String `tfsdk:"id"`
	CreationTime       types.String `tfsdk:"creation_time"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	ConnectorGroupName types.String `tfsdk:"connector_group_name"`
	InternalAppAddress types.String `tfsdk:"internal_app_address"`
	ProbingEnabled     types.Bool   `tfsdk:"probing_enabled"`
	Published          types.Bool   `tfsdk:"published"`
	AllowIcmpProtocol  types.Bool   `tfsdk:"allow_icmp_protocol"`
	ProtocolPorts      types.List   `tfsdk:"protocol_ports"`       // []ProtocolPort
	PublishedAppDomain types.Object `tfsdk:"published_app_domain"` // PublishedAppDomain
	PrivateAppProbing  types.Object `tfsdk:"private_app_probing"`  // PrivateAppProbing
}

type PublishedAppDomain struct {
	ID                 types.String `tfsdk:"id"`
	CreationTime       types.String `tfsdk:"creation_time"`
	PublishedAppDomain types.String `tfsdk:"published_app_domain"`
	CatoIP             types.String `tfsdk:"cato_ip"`
	ConnectorGroupName types.String `tfsdk:"connector_group_name"`
}

var PublishedAppDomainTypes = map[string]attr.Type{
	"id":                   types.StringType,
	"creation_time":        types.StringType,
	"published_app_domain": types.StringType,
	"cato_ip":              types.StringType,
	"connector_group_name": types.StringType,
}

type PrivateAppProbing struct {
	ID                 types.String `tfsdk:"id"`
	Type               types.String `tfsdk:"type"`
	Interval           types.Int64  `tfsdk:"interval"`
	FaultThresholdDown types.Int64  `tfsdk:"fault_threshold_down"`
}

var PrivateAppProbingTypes = map[string]attr.Type{
	"id":                   types.StringType,
	"type":                 types.StringType,
	"interval":             types.Int64Type,
	"fault_threshold_down": types.Int64Type,
}

type ProtocolPort struct {
	Ports     types.List   `tfsdk:"ports"`      // []types.Int64
	PortRange types.Object `tfsdk:"port_range"` // PortRange
	Protocol  types.String `tfsdk:"protocol"`
}

var ProtocolPortTypes = map[string]attr.Type{
	"ports":      types.ListType{ElemType: types.Int64Type},
	"port_range": types.ObjectType{AttrTypes: PortRangeTypes},
	"protocol":   types.StringType,
}

type PortRange struct {
	From types.Int64 `tfsdk:"from"`
	To   types.Int64 `tfsdk:"to"`
}

var PortRangeTypes = map[string]attr.Type{
	"from": types.Int64Type,
	"to":   types.Int64Type,
}
