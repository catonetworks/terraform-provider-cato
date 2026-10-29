package cloudinterconnectsite

import (
	"strconv"
	"strings"

	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func validID(id types.String, diags *diag.Diagnostics) bool {
	number, err := strconv.ParseInt(id.ValueString(), 10, 64)
	if id.IsNull() || id.IsUnknown() || err != nil || number <= 0 {
		diags.AddError("Invalid site ID", "A known positive numeric site ID is required.")
		return false
	}
	return true
}
func validateFields(plan resourceModel, diags *diag.Diagnostics) {
	if plan.Name.IsNull() || plan.Name.IsUnknown() || strings.TrimSpace(plan.Name.ValueString()) == "" {
		diags.AddError("Invalid name", "A known non-empty name is required.")
	}
	if plan.SiteType.IsNull() || plan.SiteType.IsUnknown() || !m.SiteType(plan.SiteType.ValueString()).IsValid() {
		diags.AddError("Invalid site type", "A known supported site type is required.")
	}
	if plan.Description.IsUnknown() {
		diags.AddError("Unknown description", "Description must be known before applying.")
	}
}
func locationValues(value types.Object, diags *diag.Diagnostics) map[string]attr.Value {
	if value.IsNull() || value.IsUnknown() {
		diags.AddError("Invalid site location", "A known site location is required.")
		return nil
	}
	attrs := value.Attributes()
	for _, key := range []string{"country_code", "timezone"} {
		field := attrs[key].(types.String)
		if field.IsNull() || field.IsUnknown() || strings.TrimSpace(field.ValueString()) == "" {
			diags.AddError("Invalid site location", "Country code and timezone must be known non-empty values.")
		}
	}
	return attrs
}
