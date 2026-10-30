package subscriptiongroup

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func validateValues(plan resourceModel, diags *diag.Diagnostics) {
	if plan.Name.IsNull() || plan.Name.IsUnknown() || strings.TrimSpace(plan.Name.ValueString()) == "" {
		diags.AddError("Invalid name", "A known, non-empty name is required.")
	}
	if plan.MailingList.IsUnknown() || plan.Integration.IsUnknown() {
		diags.AddError("Unknown membership", "Mailing-list membership must be known before applying.")
	}
}
func validID(id types.String, diags *diag.Diagnostics) bool {
	if id.IsNull() || id.IsUnknown() || strings.TrimSpace(id.ValueString()) == "" {
		diags.AddError("Invalid subscription-group ID", "A known, non-empty ID is required.")
		return false
	}
	return true
}
