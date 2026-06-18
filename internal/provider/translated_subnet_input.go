package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func stringPointerForOptionalInput(value types.String) *string {
	if value.IsNull() || value.IsUnknown() || value.ValueString() == "" {
		return nil
	}
	return value.ValueStringPointer()
}

// translatedSubnetForAPIInput omits translated_subnet from API payloads when the attribute is not
// explicitly set in Terraform config, even if plan/state still carry an API-hydrated value
// (e.g. equal to subnet when Static Range Translation is disabled).
func translatedSubnetForAPIInput(configValue, planValue types.String) *string {
	if configValue.IsNull() || configValue.IsUnknown() {
		return nil
	}
	return stringPointerForOptionalInput(planValue)
}
