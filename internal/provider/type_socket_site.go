package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SocketSite struct {
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	ConnectionType types.String `tfsdk:"connection_type"`
	SiteType       types.String `tfsdk:"site_type"`
	Description    types.String `tfsdk:"description"`
	NativeRange    types.Object `tfsdk:"native_range"`
	SiteLocation   types.Object `tfsdk:"site_location"`
}

type NativeRange struct {
	InterfaceIndex              types.String `tfsdk:"interface_index"`
	InterfaceId                 types.String `tfsdk:"interface_id"`
	InterfaceName               types.String `tfsdk:"interface_name"`
	NativeNetworkLanInterfaceId types.String `tfsdk:"native_network_lan_interface_id"`
	NativeNetworkRange          types.String `tfsdk:"native_network_range"`
	NativeNetworkRangeId        types.String `tfsdk:"native_network_range_id"`
	RangeName                   types.String `tfsdk:"range_name"`
	RangeId                     types.String `tfsdk:"range_id"`
	LocalIp                     types.String `tfsdk:"local_ip"`
	TranslatedSubnet            types.String `tfsdk:"translated_subnet"`
	Gateway                     types.String `tfsdk:"gateway"`
	RangeType                   types.String `tfsdk:"range_type"`
	DhcpSettings                types.Object `tfsdk:"dhcp_settings"`
	Vlan                        types.Int64  `tfsdk:"vlan"`
	MdnsReflector               types.Bool   `tfsdk:"mdns_reflector"`
	LagMinLinks                 types.Int64  `tfsdk:"lag_min_links"`
	InterfaceDestType           types.String `tfsdk:"interface_dest_type"`
	// InternetOnly                types.Bool   `tfsdk:"internet_only"`
}

type SiteLocation struct {
	CountryCode types.String `tfsdk:"country_code"`
	StateCode   types.String `tfsdk:"state_code"`
	Timezone    types.String `tfsdk:"timezone"`
	Address     types.String `tfsdk:"address"`
	City        types.String `tfsdk:"city"`
}

var SiteNativeRangeResourceAttrTypes = map[string]attr.Type{
	"interface_index":                 types.StringType,
	"interface_id":                    types.StringType,
	"interface_name":                  types.StringType,
	"native_network_lan_interface_id": types.StringType,
	"native_network_range":            types.StringType,
	"native_network_range_id":         types.StringType,
	"range_name":                      types.StringType,
	"range_id":                        types.StringType,
	"local_ip":                        types.StringType,
	"translated_subnet":               types.StringType,
	"gateway":                         types.StringType,
	"range_type":                      types.StringType,
	"vlan":                            types.Int64Type,
	"mdns_reflector":                  types.BoolType,
	"lag_min_links":                   types.Int64Type,
	"interface_dest_type":             types.StringType,
	// "internet_only":                   types.BoolType,
	"dhcp_settings": types.ObjectType{AttrTypes: SiteNativeRangeDhcpResourceAttrTypes},
}

var SiteNativeRangeDhcpResourceAttrTypes = map[string]attr.Type{
	"dhcp_type":              types.StringType,
	"ip_range":               types.StringType,
	"relay_group_id":         types.StringType,
	"relay_group_name":       types.StringType,
	"dhcp_microsegmentation": types.BoolType,
}

var SiteLocationResourceAttrTypes = map[string]attr.Type{
	"country_code": types.StringType,
	"state_code":   types.StringType,
	"timezone":     types.StringType,
	"address":      types.StringType,
	"city":         types.StringType,
}
