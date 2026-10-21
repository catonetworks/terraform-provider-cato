package ifrule

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// IFW Rule Schema validation functions

// Rule -> Alert Validator
type ruleAlertValidator struct{}

func (v ruleAlertValidator) Description(_ context.Context) string {
	return "If 'alert' is provided, both 'enabled' and 'frequency' must also be set, " +
		"and must specify values for mailing_list, subscription_group, or web_hook."
}

func (v ruleAlertValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v ruleAlertValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	alertObj := req.ConfigValue
	alertMap := alertObj.Attributes()

	enabled, hasEnabled := alertMap["enabled"]
	frequency, hasFrequency := alertMap["frequency"]

	if !hasEnabled || enabled.IsNull() || enabled.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("alert").AtName("enabled"),
			"'enabled' is required",
			"If 'alert' is provided, the 'enabled' field must also be set.",
		)
	}

	if !hasFrequency || frequency.IsNull() || frequency.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("alert").AtName("frequency"),
			"'frequency' is required",
			"If 'alert' is provided, the 'frequency' field must also be set.",
		)
	}
}
