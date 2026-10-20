package dhcprelaygroup

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type groupLookup struct {
	NameFilter types.List `tfsdk:"name_filter"`
	Items      types.List `tfsdk:"items"`
}

type Group struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	DhcpRelayServers types.List   `tfsdk:"dhcp_relay_servers"`
}
