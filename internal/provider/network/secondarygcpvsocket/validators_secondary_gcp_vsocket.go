package secondarygcpvsocket

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
	if p.GCPConfig.IsNull() || p.GCPConfig.IsUnknown() {
		d.AddError("Invalid GCP config", "A known non-null gcp_config object is required.")
		return
	}
	for key, value := range p.GCPConfig.Attributes() {
		ip, err := netip.ParseAddr(value.(types.String).ValueString())
		if value.IsNull() || value.IsUnknown() || err != nil || !ip.Is4() {
			d.AddError("Invalid "+key, "A known IPv4 address is required.")
		}
	}
}
