package terminalserverrule

import (
	"fmt"
	"slices"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type knownValue struct {
	name     string
	value    attr.Value
	required bool
}

func validateValues(diags *diag.Diagnostics, values ...knownValue) {
	for _, value := range values {
		if elements, ok := value.value.(types.Set); ok {
			for _, element := range elements.Elements() {
				if element.IsNull() || element.IsUnknown() {
					diags.AddError("Invalid configuration", value.name+" must not contain null or unknown elements.")
					break
				}
			}
		}
		if value.value.IsUnknown() || (value.required && value.value.IsNull()) {
			diags.AddError("Invalid configuration", value.name+" must contain a known value.")
		}
	}
}

func validID(id types.String, diags *diag.Diagnostics) bool {
	if id.IsNull() || id.IsUnknown() || id.ValueString() == "" {
		diags.AddError("Invalid resource ID", "A known, non-empty resource ID is required in state.")
		return false
	}
	return true
}

func validatePosition(at *cato_models.PolicyRulePositionInput, diags *diag.Diagnostics) {
	if at == nil {
		return
	}
	if at.Position == nil {
		diags.AddError("Missing position", "A position is required when at is configured.")
		return
	}
	position := *at.Position
	if !slices.Contains(cato_models.AllPolicyRulePositionEnum, position) {
		diags.AddError("Invalid position", fmt.Sprintf("Unsupported position %q.", position))
		return
	}
	if position != cato_models.PolicyRulePositionEnumLastInPolicy && position != cato_models.PolicyRulePositionEnumFirstInPolicy {
		if at.Ref == nil || *at.Ref == "" {
			diags.AddError("Missing position reference", "Relative placement requires a known, non-empty ref.")
		}
	}
}
