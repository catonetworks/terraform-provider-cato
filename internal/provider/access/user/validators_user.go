package user

import (
	"context"
	"net/mail"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type userTextValidator struct{ required, email, phone bool }

func (v userTextValidator) Description(context.Context) string {
	return "Use a known, trimmed value and a plain email address for email."
}
func (v userTextValidator) MarkdownDescription(ctx context.Context) string { return v.Description(ctx) }
func (v userTextValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if err := validateText(req.ConfigValue.ValueString(), v.required, v.email); err != "" {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid user configuration", err)
		return
	}
	if v.phone && !validPhone(req.ConfigValue.ValueString()) {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid phone number", "Use digits, spaces, and an optional leading +.")
	}
}
func validateText(value string, required, email bool) string {
	if value != strings.TrimSpace(value) {
		return "Leading and trailing whitespace are not supported by the user API."
	}
	if required && value == "" {
		return "A non-empty value is required."
	}
	if email {
		address, err := mail.ParseAddress(value)
		if err != nil || address.Address != value {
			return "Use a plain, valid email address."
		}
	}
	return ""
}
func validateModel(model resourceModel, diags *diag.Diagnostics) {
	fields := []struct {
		name            string
		value           types.String
		required, email bool
	}{
		{"email", model.Email, true, true},
		{"first_name", model.FirstName, true, false},
		{"last_name", model.LastName, true, false},
		{"department", model.Department, false, false},
		{"job_title", model.JobTitle, false, false},
		{"phone_number", model.PhoneNumber, false, false},
	}
	for _, field := range fields {
		if field.value.IsUnknown() || (field.required && field.value.IsNull()) {
			diags.AddError("Invalid user configuration", field.name+" must have a known value.")
			continue
		}
		if message := validateText(field.value.ValueString(), field.required, field.email); message != "" {
			diags.AddError("Invalid user configuration", field.name+": "+message)
		}
	}
	if !model.PhoneNumber.IsNull() && !model.PhoneNumber.IsUnknown() && !validPhone(model.PhoneNumber.ValueString()) {
		diags.AddError("Invalid phone number", "Use digits, spaces, and an optional leading +.")
	}
	if model.Enabled.IsNull() || model.Enabled.IsUnknown() {
		diags.AddError("Invalid user configuration", "enabled must have a known value.")
	}
}
func validUserID(id types.String, diags *diag.Diagnostics) bool {
	number, err := strconv.ParseInt(id.ValueString(), 10, 64)
	if id.IsNull() || id.IsUnknown() || err != nil || number <= 0 {
		diags.AddError("Invalid user ID", "A known positive numeric user ID is required.")
		return false
	}
	return true
}

var userPhonePattern = regexp.MustCompile(`^[+]?[0-9 ]+$`)

func validPhone(value string) bool { return value == "" || userPhonePattern.MatchString(value) }
