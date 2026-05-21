package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type StaticHost struct {
	ID         types.String `tfsdk:"id"`
	SiteID     types.String `tfsdk:"site_id"`
	Name       types.String `tfsdk:"name"`
	IP         types.String `tfsdk:"ip"`
	MacAddress types.String `tfsdk:"mac_address"`
}
