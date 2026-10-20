package networkinterfaces

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type NetworkInterfaceLookup struct {
	SiteID                types.String `tfsdk:"site_id"`
	NetworkInterfaceIndex types.String `tfsdk:"network_interface_index"`
	Items                 types.List   `tfsdk:"items"`
}

type NetworkInterface struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	SiteID            types.String `tfsdk:"site_id"`
	SiteName          types.String `tfsdk:"site_name"`
	Subnet            types.String `tfsdk:"subnet"`
	SocketInterfaceID types.String `tfsdk:"socket_interface_id"`
}
