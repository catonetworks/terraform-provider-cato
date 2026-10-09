package dhcp

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Settings struct {
	DhcpType              types.String `tfsdk:"dhcp_type"`
	IPRange               types.String `tfsdk:"ip_range"`
	RelayGroupID          types.String `tfsdk:"relay_group_id"`
	RelayGroupName        types.String `tfsdk:"relay_group_name"`
	DhcpMicrosegmentation types.Bool   `tfsdk:"dhcp_microsegmentation"`
}

var SettingsAttrTypes = map[string]attr.Type{
	"dhcp_type":              types.StringType,
	"ip_range":               types.StringType,
	"relay_group_id":         types.StringType,
	"relay_group_name":       types.StringType,
	"dhcp_microsegmentation": types.BoolType,
}
