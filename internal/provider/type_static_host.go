package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type StaticHost struct {
	Id         types.String `tfsdk:"id"`
	SiteId     types.String `tfsdk:"site_id"`
	Name       types.String `tfsdk:"name"`
	Ip         types.String `tfsdk:"ip"`
	MacAddress types.String `tfsdk:"mac_address"`
}
