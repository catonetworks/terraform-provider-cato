package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PrivateAppModel struct {
	AllowIcmpProtocol  types.Bool   `tfsdk:"allow_icmp_protocol"`
	CreationTime       types.String `tfsdk:"creation_time"`
	Description        types.String `tfsdk:"description"`
	ID                 types.String `tfsdk:"id"`
	InternalAppAddress types.String `tfsdk:"internal_app_address"`
	Name               types.String `tfsdk:"name"`
	PrivateAppProbing  types.Object `tfsdk:"private_app_probing"` // PrivateAppProbing
	ProbingEnabled     types.Bool   `tfsdk:"probing_enabled"`
	ProtocolPorts      types.Set    `tfsdk:"protocol_ports"` // []ProtocolPort
	Published          types.Bool   `tfsdk:"published"`
	PublishedAppDomain types.Object `tfsdk:"published_app_domain"` // PublishedAppDomain
}

type PublishedAppDomain struct {
	ConnectorGroupName types.String `tfsdk:"connector_group_name"`
	CreationTime       types.String `tfsdk:"creation_time"`
	ID                 types.String `tfsdk:"id"`
	PublishedAppDomain types.String `tfsdk:"published_app_domain"`
}

var PublishedAppDomainTypes = map[string]attr.Type{
	"connector_group_name": types.StringType,
	"creation_time":        types.StringType,
	"id":                   types.StringType,
	"published_app_domain": types.StringType,
}

type PrivateAppProbing struct {
	FaultThresholdDown types.Int64  `tfsdk:"fault_threshold_down"`
	ID                 types.String `tfsdk:"id"`
	Interval           types.Int64  `tfsdk:"interval"`
	Type               types.String `tfsdk:"type"`
}

var PrivateAppProbingTypes = map[string]attr.Type{
	"fault_threshold_down": types.Int64Type,
	"id":                   types.StringType,
	"interval":             types.Int64Type,
	"type":                 types.StringType,
}

type ProtocolPort struct {
	PortRange types.Object `tfsdk:"port_range"` // PortRange
	Ports     types.Set    `tfsdk:"ports"`      // []types.Int64
	Protocol  types.String `tfsdk:"protocol"`
}

var ProtocolPortTypes = map[string]attr.Type{
	"port_range": types.ObjectType{AttrTypes: PortRangeTypes},
	"ports":      types.SetType{ElemType: types.Int64Type},
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
