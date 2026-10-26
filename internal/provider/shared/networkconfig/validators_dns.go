package networkconfig

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ValidateDNS checks dependencies between DNS servers and suffixes.
func ValidateDNS(primary, secondary types.String, suffix types.List, diags *diag.Diagnostics) {
	if primary.IsUnknown() || primary.ValueString() != "" {
		return
	}
	if secondary.ValueString() != "" && !secondary.IsUnknown() {
		diags.AddAttributeError(path.Root("secondary_server"),
			"Invalid DNS settings",
			"A secondary DNS server requires a primary DNS server.")
	}
	if !suffix.IsNull() && !suffix.IsUnknown() && len(suffix.Elements()) > 0 {
		diags.AddAttributeError(path.Root("suffix"), "Invalid DNS settings", "DNS search suffixes require a primary DNS server.")
	}
}
