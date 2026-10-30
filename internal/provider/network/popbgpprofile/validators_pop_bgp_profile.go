package popbgpprofile

import (
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const communityBits = 16

func validID(value types.String, diags *diag.Diagnostics) bool {
	id, err := strconv.ParseInt(value.ValueString(), 10, 64)
	if value.IsNull() || value.IsUnknown() || err != nil || id <= 0 {
		diags.AddError("Invalid profile ID", "A known positive numeric profile ID is required.")
		return false
	}
	return true
}
func validASN(value string) bool {
	n, err := strconv.ParseUint(value, 10, communityBits)
	return err == nil && strconv.FormatUint(n, 10) == value
}
func validateFields(plan resourceModel, diags *diag.Diagnostics) {
	if plan.Name.IsNull() || plan.Name.IsUnknown() || strings.TrimSpace(plan.Name.ValueString()) == "" {
		diags.AddError("Invalid name", "A known non-empty name is required.")
	}
	if plan.Description.IsUnknown() {
		diags.AddError("Unknown description", "Description must be known before applying.")
	}
	if plan.Communities.IsNull() || plan.Communities.IsUnknown() || len(plan.Communities.Elements()) == 0 {
		diags.AddError("Invalid communities", "At least one known community pair is required.")
	}
}
