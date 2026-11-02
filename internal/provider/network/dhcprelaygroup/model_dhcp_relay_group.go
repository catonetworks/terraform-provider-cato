package dhcprelaygroup

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// RelayGroupModel contains one DHCP relay group and its ordered servers.
type RelayGroupModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Server types.List   `tfsdk:"server"`
}
