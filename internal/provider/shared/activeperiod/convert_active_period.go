package activeperiod

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// GetActivePeriodString safely converts API active period string fields to Terraform string values
// Returns null string if empty, otherwise returns the string value
func GetActivePeriodString(value *string) types.String {
	if value == nil || *value == "" {
		return types.StringNull()
	}
	return types.StringValue(*value)
}
