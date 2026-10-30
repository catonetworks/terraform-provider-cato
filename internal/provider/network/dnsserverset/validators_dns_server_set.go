package dnsserverset

import (
	"net/netip"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const maxServers = 2

func validID(id types.String, diags *diag.Diagnostics) bool {
	if id.IsNull() || id.IsUnknown() || strings.TrimSpace(id.ValueString()) == "" {
		diags.AddError("Invalid DNS set ID", "A known, non-empty ID is required.")
		return false
	}
	return true
}
func validateInputs(plan resourceModel, diags *diag.Diagnostics) {
	if plan.Name.IsNull() || plan.Name.IsUnknown() || strings.TrimSpace(plan.Name.ValueString()) == "" {
		diags.AddError("Invalid name", "A known, non-empty name is required.")
	}
	if plan.Name.ValueString() != strings.TrimSpace(plan.Name.ValueString()) {
		diags.AddError("Invalid name", "Leading and trailing whitespace is unsupported because the API trims names.")
	}
	if plan.Server.IsNull() || plan.Server.IsUnknown() {
		diags.AddError("Invalid server list", "A known, non-null ordered list is required.")
		return
	}
	if len(plan.Server.Elements()) > maxServers {
		diags.AddError("Too many servers", "At most two IPv4 DNS servers are supported.")
	}
	seen := map[string]bool{}
	for _, value := range plan.Server.Elements() {
		if value.IsNull() || value.IsUnknown() {
			diags.AddError("Invalid server", "Each member must be known and non-null.")
			continue
		}
		text := value.(types.String).ValueString()
		if !validMember(text) {
			diags.AddError("Invalid server", "Each server must be an IPv4 address.")
		}
		if seen[text] {
			diags.AddError("Duplicate server", "DNS server members must be unique.")
		}
		seen[text] = true
	}
}
func validMember(value string) bool { ip, err := netip.ParseAddr(value); return err == nil && ip.Is4() }
