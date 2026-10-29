package wifissid

import "github.com/hashicorp/terraform-plugin-framework/types"

type resourceModel struct {
	Band                     types.String `tfsdk:"band"`
	Category                 types.String `tfsdk:"category"`
	Dhcp                     types.Object `tfsdk:"dhcp"`
	DhcpEnabled              types.Bool   `tfsdk:"dhcp_enabled"`
	Enabled                  types.Bool   `tfsdk:"enabled"`
	ID                       types.String `tfsdk:"id"`
	InternetOnly             types.Bool   `tfsdk:"internet_only"`
	LocalIP                  types.String `tfsdk:"local_ip"`
	MdnsEnabled              types.Bool   `tfsdk:"mdns_enabled"`
	MicrosegmentationEnabled types.Bool   `tfsdk:"microsegmentation_enabled"`
	Name                     types.String `tfsdk:"name"`
	NetworkInterfaceID       types.String `tfsdk:"network_interface_id"`
	Security                 types.Object `tfsdk:"security"`
	Site                     types.Object `tfsdk:"site"`
	Subnet                   types.String `tfsdk:"subnet"`
	Visible                  types.Bool   `tfsdk:"visible"`
}
