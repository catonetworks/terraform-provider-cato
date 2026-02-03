package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PrivateAppModel struct {
	ID                 types.String        `tfsdk:"id"`
	CreationTime       types.String        `tfsdk:"creation_time"`
	Name               types.String        `tfsdk:"name"`
	Description        types.String        `tfsdk:"description"`
	InternalAppAddress types.String        `tfsdk:"internal_app_address"`
	ProbingEnabled     types.Bool          `tfsdk:"probing_enabled"`
	Published          types.Bool          `tfsdk:"published"`
	AllowIcmpProtocol  types.Bool          `tfsdk:"allow_icmp_protocol"`
	PublishedAppDomain *PublishedAppDomain `tfsdk:"published_app_domain"`
	PrivateAppProbing  *PrivateAppProbing  `tfsdk:"private_app_probing"`
	ProtocolPorts      []ProtocolPort      `tfsdk:"protocol_ports"`
}

type PublishedAppDomain struct {
	ID                 types.String `tfsdk:"id"`
	CreationTime       types.String `tfsdk:"creation_time"`
	PublishedAppDomain types.String `tfsdk:"published_app_domain"`
	CatoIP             types.String `tfsdk:"cato_ip"`
	ConnectorGroupName types.String `tfsdk:"connector_group_name"`
}

type PrivateAppProbing struct {
	ID                 types.String `tfsdk:"id"`
	Type               types.String `tfsdk:"type"`
	Interval           types.Int64  `tfsdk:"interval"`
	FaultThresholdDown types.Int64  `tfsdk:"fault_threshold_down"`
}

type ProtocolPort struct {
	Ports     []types.Int64 `tfsdk:"ports"`
	PortRange *PortRange    `tfsdk:"port_range"`
	Protocol  types.String  `tfsdk:"protocol"`
}
type PortRange struct {
	From types.Int64 `tfsdk:"from"`
	To   types.Int64 `tfsdk:"to"`
}
