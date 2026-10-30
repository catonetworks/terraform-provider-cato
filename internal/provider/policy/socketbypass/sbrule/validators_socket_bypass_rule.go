package sbrule

import (
	"strings"

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
				if strings.HasSuffix(value.name, ".vlan") && !element.IsNull() && !element.IsUnknown() {
					if number, ok := element.(types.Int64); ok && (number.ValueInt64() < 1 || number.ValueInt64() > 4094) {
						diags.AddError("Invalid VLAN", value.name+" must contain VLAN IDs between 1 and 4094.")
					}
				}
			}
		}
		if value.value.IsUnknown() || (value.required && value.value.IsNull()) {
			diags.AddError("Invalid configuration", value.name+" must contain a known value.")
		}
	}
}
