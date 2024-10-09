package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type NetworkRange struct {
	Id               types.String `tfsdk:"id"`
	InterfaceId      types.String `tfsdk:"interface_id"`
	SiteId           types.String `tfsdk:"site_id"`
	Name             types.String `tfsdk:"name"`
	RangeType        types.String `tfsdk:"range_type"`
	Subnet           types.String `tfsdk:"subnet"`
	TranslatedSubnet types.String `tfsdk:"translated_subnet"`
	LocalIp          types.String `tfsdk:"local_ip"`
	Gateway          types.String `tfsdk:"gateway"`
	Vlan             types.Int64  `tfsdk:"vlan"`
	DhcpSettings     types.Object `tfsdk:"dhcp_settings"`
}

type DhcpSettings struct {
	DhcpType     types.String `tfsdk:"dhcp_type"`
	IpRange      types.String `tfsdk:"ip_range"`
	RelayGroupId types.String `tfsdk:"relay_group_id"`
}
