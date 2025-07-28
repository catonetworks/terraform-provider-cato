package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type NetworkRange struct {
	Id               types.String `tfsdk:"id"`
	DhcpSettings     types.Object `tfsdk:"dhcp_settings"`
	Gateway          types.String `tfsdk:"gateway"`
	InterfaceId      types.String `tfsdk:"interface_id"`
	InternetOnly     types.Bool   `tfsdk:"internet_only"`
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
	DhcpMicrosegmentation types.Bool   `tfsdk:"dhcp_microsegmentation"`
}
