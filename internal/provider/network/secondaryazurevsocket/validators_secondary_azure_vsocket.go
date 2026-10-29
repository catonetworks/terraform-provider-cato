package secondaryazurevsocket

import (
	"net/netip"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func validID(id types.String, diags *diag.Diagnostics) bool {
	if id.IsNull() || id.IsUnknown() || strings.TrimSpace(id.ValueString()) == "" {
		diags.AddError("Invalid socket ID", "A known non-empty secondary socket ID is required.")
		return false
	}
	return true
}
func validateFields(p resourceModel, d *diag.Diagnostics) {
	for key, value := range map[string]types.String{"interface_ip": p.InterfaceIP, "floating_ip": p.FloatingIP} {
		ip, err := netip.ParseAddr(value.ValueString())
		if value.IsNull() || value.IsUnknown() || err != nil || !ip.Is4() {
			d.AddError("Invalid "+key, "A known IPv4 address is required.")
		}
	}
}
