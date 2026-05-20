package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type LanInterface struct {
	ID               types.String `tfsdk:"id"`
	SiteID           types.String `tfsdk:"site_id"`
	InterfaceID      types.String `tfsdk:"interface_id"`
	Name             types.String `tfsdk:"name"`
	DestType         types.String `tfsdk:"dest_type"`
	LocalIP          types.String `tfsdk:"local_ip"`
	LagMinLinks      types.Int64  `tfsdk:"lag_min_links"`
	Subnet           types.String `tfsdk:"subnet"`
	TranslatedSubnet types.String `tfsdk:"translated_subnet"`
	VrrpType         types.String `tfsdk:"vrrp_type"`
}

type LanInterfaceLagMember struct {
	ID          types.String `tfsdk:"id"`
	SiteID      types.String `tfsdk:"site_id"`
	InterfaceID types.String `tfsdk:"interface_id"`
	Name        types.String `tfsdk:"name"`
	DestType    types.String `tfsdk:"dest_type"`
}
