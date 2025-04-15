package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type LanInterface struct {
	ID               types.String `tfsdk:"id"`
	SiteId           types.String `tfsdk:"site_id"`
	InterfaceID      types.String `tfsdk:"interface_id"`
	Name             types.String `tfsdk:"name"`
	DestType         types.String `tfsdk:"dest_type"`
	LocalIp          types.String `tfsdk:"local_ip"`
	Subnet           types.String `tfsdk:"subnet"`
	TranslatedSubnet types.String `tfsdk:"translated_subnet"`
	VrrpType         types.String `tfsdk:"vrrp_type"`
}
