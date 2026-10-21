package tfmodel

import "github.com/hashicorp/terraform-plugin-framework/types"

type DhcpSettings struct {
	DhcpType              types.String `tfsdk:"dhcp_type"`
	IPRange               types.String `tfsdk:"ip_range"`
	RelayGroupID          types.String `tfsdk:"relay_group_id"`
	RelayGroupName        types.String `tfsdk:"relay_group_name"`
	DhcpMicrosegmentation types.Bool   `tfsdk:"dhcp_microsegmentation"`
}
