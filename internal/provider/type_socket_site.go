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
	NativeNetworkRange   types.String `tfsdk:"native_network_range"`
	NativeNetworkRangeId types.String `tfsdk:"native_network_range_id"`
	LocalIp              types.String `tfsdk:"local_ip"`
	TranslatedSubnet     types.String `tfsdk:"translated_subnet"`
	DhcpSettings         types.Object `tfsdk:"dhcp_settings"`
	Vlan                 types.Int64  `tfsdk:"vlan"`
	MdnsReflector        types.Bool   `tfsdk:"mdns_reflector"`
	InternetOnly         types.Bool   `tfsdk:"internet_only"`
}

type SiteLocation struct {
	CountryCode types.String `tfsdk:"country_code"`
	StateCode   types.String `tfsdk:"state_code"`
	Timezone    types.String `tfsdk:"timezone"`
	Address     types.String `tfsdk:"address"`
	City        types.String `tfsdk:"city"`
}

var SiteNativeRangeResourceAttrTypes = map[string]attr.Type{
	"native_network_range":    types.StringType,
	"native_network_range_id": types.StringType,
	"local_ip":                types.StringType,
	"translated_subnet":       types.StringType,
	"vlan":                    types.Int64Type,
	"mdns_reflector":          types.BoolType,
	"internet_only":           types.BoolType,
	"dhcp_settings":           types.ObjectType{AttrTypes: SiteNativeRangeDhcpResourceAttrTypes},
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
