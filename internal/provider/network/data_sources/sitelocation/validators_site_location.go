package sitelocation

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type sldFilterValidator struct{}

func (v sldFilterValidator) Description(_ context.Context) string {
	return "Ensures at least one filter is set with " +
		"field city, state_name, or country_name"
}

func (v sldFilterValidator) MarkdownDescription(_ context.Context) string {
	return "Ensures at least one filter is set with " +
		"field `city`, `state_name`, or `country_name`"
}

func (v sldFilterValidator) ValidateList(
	_ context.Context,
	req validator.ListRequest,
	resp *validator.ListResponse,
) {
	// Allow unknown values - they will be validated once resolved during planning
	if req.ConfigValue.IsUnknown() {
		return
	}

	if req.ConfigValue.IsNull() {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic(
			"Invalid Filters Configuration",
			"At least one filter must be specified with field city, state_name, or country_name",
		))
		return
	}

	elements := req.ConfigValue.Elements()
	if len(elements) == 0 {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic(
			"Invalid Filters Configuration",
			"At least one filter must be specified with field city, state_name, or country_name",
		))
		return
	}
}
