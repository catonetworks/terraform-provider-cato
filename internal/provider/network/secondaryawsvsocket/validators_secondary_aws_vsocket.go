package secondaryawsvsocket

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
	for key, value := range map[string]types.String{
		"eni_ip_address": p.EniIPAddress, "eni_ip_subnet": p.EniIPSubnet, "route_table_id": p.RouteTableID,
	} {
		if value.IsNull() || value.IsUnknown() || strings.TrimSpace(value.ValueString()) == "" ||
			strings.TrimSpace(value.ValueString()) != value.ValueString() {
			d.AddError("Invalid "+key, "A known non-empty value without surrounding whitespace is required.")
		}
	}
	prefix, err := netip.ParsePrefix(p.EniIPSubnet.ValueString())
	if err != nil || !prefix.Addr().Is4() || prefix != prefix.Masked() {
		d.AddError("Invalid ENI subnet", "Use a canonical IPv4 network prefix.")
		return
	}
	ip, err := netip.ParseAddr(p.EniIPAddress.ValueString())
	if err != nil || !ip.Is4() || !prefix.Contains(ip) {
		d.AddError("Invalid ENI IP address", "Use an IPv4 address within eni_ip_subnet.")
	}
}
