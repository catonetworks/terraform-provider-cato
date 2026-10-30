package allocatedip

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func validID(id types.String, diags *diag.Diagnostics) bool {
	if id.IsNull() || id.IsUnknown() || strings.TrimSpace(id.ValueString()) == "" {
		diags.AddError("Invalid allocation ID", "A known, non-empty allocation ID is required.")
		return false
	}
	return true
}
func validateType(value types.String, diags *diag.Diagnostics) {
	if value.IsNull() || value.IsUnknown() || value.ValueString() != "SYSTEM" {
		diags.AddError("Unsupported allocation type", "Only SYSTEM supports public allocation and release; BYOIP requires private API inputs.")
		return
	}
}
