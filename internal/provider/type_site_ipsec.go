package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SiteIpsecIkeV2 struct {
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	SiteType             types.String `tfsdk:"site_type"`
	Description          types.String `tfsdk:"description"`
	NativeNetworkRange   types.String `tfsdk:"native_network_range"`
	NativeNetworkRangeId types.String `tfsdk:"native_network_range_id"`
	InterfaceId          types.String `tfsdk:"interface_id"`
	SiteLocation         types.Object `tfsdk:"site_location"`
	IPSec                types.Object `tfsdk:"ipsec"`
}

type AddIpsecSiteLocationInput struct {
	CountryCode types.String `tfsdk:"country_code"`
	StateCode   types.String `tfsdk:"state_code"`
	Timezone    types.String `tfsdk:"timezone"`
	Address     types.String `tfsdk:"address"`
	City        types.String `tfsdk:"city"`
}

type AddIpsecIkeV2SiteTunnelsInput struct {
	SiteId             types.String `tfsdk:"site_id"`
	Primary            types.Object `tfsdk:"primary"`             //AddIpsecIkeV2TunnelsInput
	Secondary          types.Object `tfsdk:"secondary"`           //AddIpsecIkeV2TunnelsInput
	ConnectionMode     types.String `tfsdk:"connection_mode"`     //ConnectionMode enum
	IdentificationType types.String `tfsdk:"identification_type"` //IdentificationType enum
	InitMessage        types.Object `tfsdk:"init_message"`        //IpsecIkeV2MessageInput
	AuthMessage        types.Object `tfsdk:"auth_message"`        //IpsecIkeV2MessageInput
	NetworkRanges      types.List   `tfsdk:"network_ranges"`      //String Array
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

type IpsecIkeV2MessageInput struct {
	Cipher    types.String `tfsdk:"cipher"`    //IpSecCipher enum
	DhGroup   types.String `tfsdk:"dh_group"`  //IpSecDHGroup enum
	Integrity types.String `tfsdk:"integrity"` //IpSecHash enum
	Prf       types.String `tfsdk:"prf"`       //IpSecHash enum
}

// Define attribute types for nested objects
// Note: SiteLocationResourceAttrTypes is already defined in type_socket_site.go

var IpsecMessageResourceAttrTypes = map[string]attr.Type{
	"cipher":    types.StringType,
	"dh_group":  types.StringType,
	"integrity": types.StringType,
	"prf":       types.StringType,
}

var TunnelResourceAttrTypes = map[string]attr.Type{
	"tunnel_id":       types.StringType,
	"public_site_ip":  types.StringType,
	"private_cato_ip": types.StringType,
	"private_site_ip": types.StringType,
	"psk":             types.StringType,
	"last_mile_bw":    types.ObjectType{AttrTypes: LastMileBwResourceAttrTypes},
}

var LastMileBwResourceAttrTypes = map[string]attr.Type{
	"downstream":                types.Int64Type,
	"upstream":                  types.Int64Type,
	"downstream_mbps_precision": types.Float64Type,
	"upstream_mbps_precision":   types.Float64Type,
}

var IpsecTunnelsResourceAttrTypes = map[string]attr.Type{
	"destination_type":  types.StringType,
	"public_cato_ip_id": types.StringType,
	"pop_location_id":   types.StringType,
	"tunnels":           types.ListType{ElemType: types.ObjectType{AttrTypes: TunnelResourceAttrTypes}},
}

var IpsecResourceAttrTypes = map[string]attr.Type{
	"site_id":             types.StringType,
	"primary":             types.ObjectType{AttrTypes: IpsecTunnelsResourceAttrTypes},
	"secondary":           types.ObjectType{AttrTypes: IpsecTunnelsResourceAttrTypes},
	"connection_mode":     types.StringType,
	"identification_type": types.StringType,
	"init_message":        types.ObjectType{AttrTypes: IpsecMessageResourceAttrTypes},
	"auth_message":        types.ObjectType{AttrTypes: IpsecMessageResourceAttrTypes},
	"network_ranges":      types.ListType{ElemType: types.StringType},
}
