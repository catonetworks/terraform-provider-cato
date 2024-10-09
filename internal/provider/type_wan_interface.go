package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type WanInterface struct {
	SiteId              types.String `tfsdk:"site_id"`
	InterfaceID         types.String `tfsdk:"interface_id"`
	Name                types.String `tfsdk:"name"`
	UpstreamBandwidth   types.Int64  `tfsdk:"upstream_bandwidth"`
	DownstreamBandwidth types.Int64  `tfsdk:"downstream_bandwidth"`
	Role                types.String `tfsdk:"role"`
	Precedence          types.String `tfsdk:"precedence"`
	// OffCloud            types.Object `tfsdk:"off_cloud"`
}

// type OffCloud struct {
// 	Enabled    types.Bool   `tfsdk:"enabled"`
// 	PublicIp   types.String `tfsdk:"public_ip"`
// 	PublicPort types.String `tfsdk:"public_port"`
// }
