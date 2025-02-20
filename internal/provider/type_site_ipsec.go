package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type SiteIpsecIkeV2 struct {
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	SiteType             types.String `tfsdk:"site_type"`
	Description          types.String `tfsdk:"description"`
	NativeNetworkRange   types.String `tfsdk:"native_network_range"`
	NativeNetworkRangeId types.String `tfsdk:"native_network_range_id"`
	SiteLocation         types.Object `tfsdk:"site_location"`
	IPSec                types.Object `tfsdk:"ipsec"`
}

type AddIpsecSiteLocationInput struct {
	CountryCode types.String `tfsdk:"country_code"`
	StateCode   types.String `tfsdk:"state_code"`
	Timezone    types.String `tfsdk:"timezone"`
	Address     types.String `tfsdk:"address"`
	// City        types.String `tfsdk:"city"`
}

type AddIpsecIkeV2SiteTunnelsInput struct {
	SiteId    types.String `tfsdk:"site_id"`
	Primary   types.Object `tfsdk:"primary"`   //AddIpsecIkeV2TunnelsInput
	Secondary types.Object `tfsdk:"secondary"` //AddIpsecIkeV2TunnelsInput
}

type AddIpsecIkeV2TunnelsInput struct {
	DestinationType types.String `tfsdk:"destination_type"`
	PublicCatoIPID  types.String `tfsdk:"public_cato_ip_id"`
	PopLocationID   types.String `tfsdk:"pop_location_id"`
	Tunnels         types.List   `tfsdk:"tunnels"` //[]*AddIpsecIkeV2TunnelInput
}

type AddIpsecIkeV2TunnelInput struct {
	TunnelID      types.String `tfsdk:"tunnel_id"`
	PublicSiteIP  types.String `tfsdk:"public_site_ip"`
	PrivateCatoIP types.String `tfsdk:"private_cato_ip"`
	PrivateSiteIP types.String `tfsdk:"private_site_ip"`
	LastMileBw    types.Object `tfsdk:"last_mile_bw"` //*LastMileBwInput
	Psk           types.String `tfsdk:"psk"`
}

type LastMileBwInput struct {
	Downstream              types.Int64   `tfsdk:"downstream"`
	Upstream                types.Int64   `tfsdk:"upstream"`
	DownstreamMbpsPrecision types.Float64 `tfsdk:"downstream_mbps_precision"`
	UpstreamMbpsPrecision   types.Float64 `tfsdk:"upstream_mbps_precision"`
}
