package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type NetworkRange struct {
	Id               types.String `tfsdk:"id"`
	DhcpSettings     types.Object `tfsdk:"dhcp_settings"`
	Gateway          types.String `tfsdk:"gateway"`
	InterfaceId      types.String `tfsdk:"interface_id"`
	InterfaceIndex   types.String `tfsdk:"interface_index"`
	InternetOnly     types.Bool   `tfsdk:"internet_only"`
	MdnsReflector    types.Bool   `tfsdk:"mdns_reflector"`
	LocalIp          types.String `tfsdk:"local_ip"`
	Name             types.String `tfsdk:"name"`
	RangeType        types.String `tfsdk:"range_type"`
	SiteId           types.String `tfsdk:"site_id"`
	Subnet           types.String `tfsdk:"subnet"`
	TranslatedSubnet types.String `tfsdk:"translated_subnet"`
	Vlan             types.Int64  `tfsdk:"vlan"`
}

type DhcpSettings struct {
	DhcpType              types.String `tfsdk:"dhcp_type"`
	IpRange               types.String `tfsdk:"ip_range"`
	RelayGroupId          types.String `tfsdk:"relay_group_id"`
	RelayGroupName        types.String `tfsdk:"relay_group_name"`
	DhcpMicrosegmentation types.Bool   `tfsdk:"dhcp_microsegmentation"`
}

type NetworkRangeLookup struct {
	SiteIdFilter types.List `tfsdk:"site_id_filter"`
	NameFilter   types.List `tfsdk:"name_filter"`
	Items        types.List `tfsdk:"items"`
}

var DhcpSettingsAttrTypes = map[string]attr.Type{
	"dhcp_type":              types.StringType,
	"ip_range":               types.StringType,
	"relay_group_id":         types.StringType,
	"relay_group_name":       types.StringType,
	"dhcp_microsegmentation": types.BoolType,
}
