package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

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
}

type SiteLocation struct {
	CountryCode types.String `tfsdk:"country_code"`
	StateCode   types.String `tfsdk:"state_code"`
	Timezone    types.String `tfsdk:"timezone"`
	Address     types.String `tfsdk:"address"`
	// City        types.String `tfsdk:"city"`
}
