package validators

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Datetime validator - RFC3339 "2006-01-02T15:04:05Z"
type DateTimeValidator struct{}

var dateTimeRE = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)

func (v DateTimeValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() {
		return
	}
	value := req.ConfigValue.ValueString()
	if !dateTimeRE.MatchString(value) {
		resp.Diagnostics.AddError("Field validation error", fmt.Sprintf("invalid datetime (%s: %s)\n - must be in RFC3339 format (2006-01-02T15:04:05Z)",
			req.Path.String(), value))
		return
	}
}
func (v DateTimeValidator) Description(ctx context.Context) string {
	return "DateTime must be in RFC3339 format, UTC (2006-01-02T15:04:05Z)"
}
func (v DateTimeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
