package statichost

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type HostLookup struct {
	NameFilter types.List `tfsdk:"name_filter"`
	IPFilter   types.List `tfsdk:"ip_filter"`
	Items      types.List `tfsdk:"items"`
}

type Host struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	IPAddr  types.String `tfsdk:"ip"`
	Network types.String `tfsdk:"net"`
	SiteID  types.String `tfsdk:"site_id"`
}

type hostFilter struct {
	namesMap map[string]struct{}
	ipsMap   map[string]struct{}
}
