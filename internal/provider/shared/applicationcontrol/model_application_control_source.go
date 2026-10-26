package applicationcontrol

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ApplicationControlSourceTF mirrors ApplicationControlSourceAttrTypes / WAN-like source fields.
type ApplicationControlSourceTF struct {
	Country           types.Set  `tfsdk:"country"`
	IP                types.List `tfsdk:"ip"`
	Host              types.Set  `tfsdk:"host"`
	Site              types.Set  `tfsdk:"site"`
	Subnet            types.List `tfsdk:"subnet"`
	IPRange           types.List `tfsdk:"ip_range"`
	GlobalIPRange     types.Set  `tfsdk:"global_ip_range"`
	NetworkInterface  types.Set  `tfsdk:"network_interface"`
	SiteNetworkSubnet types.Set  `tfsdk:"site_network_subnet"`
	FloatingSubnet    types.Set  `tfsdk:"floating_subnet"`
	User              types.Set  `tfsdk:"user"`
	UsersGroup        types.Set  `tfsdk:"users_group"`
	Group             types.Set  `tfsdk:"group"`
	SystemGroup       types.Set  `tfsdk:"system_group"`
}
