package dnssuffixset

import (
	"strings"
	"unicode"
	"unicode/utf16"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const maxSuffixLength = 253

func validID(id types.String, diags *diag.Diagnostics) bool {
	if id.IsNull() || id.IsUnknown() || strings.TrimSpace(id.ValueString()) == "" {
		diags.AddError("Invalid DNS set ID", "A known, non-empty ID is required.")
		return false
	}
	return true
}
func validateInputs(plan resourceModel, diags *diag.Diagnostics) {
	if plan.Name.IsNull() || plan.Name.IsUnknown() || strings.TrimSpace(plan.Name.ValueString()) == "" {
		diags.AddError("Invalid name", "A known, non-empty name is required.")
	}
	if plan.Name.ValueString() != strings.TrimSpace(plan.Name.ValueString()) {
		diags.AddError("Invalid name", "Leading and trailing whitespace is unsupported because the API trims names.")
	}
	if plan.Suffix.IsNull() || plan.Suffix.IsUnknown() {
		diags.AddError("Invalid suffix list", "A known, non-null ordered list is required.")
		return
	}

	seen := map[string]bool{}
	for _, value := range plan.Suffix.Elements() {
		if value.IsNull() || value.IsUnknown() {
			diags.AddError("Invalid suffix", "Each member must be known and non-null.")
			continue
		}
		text := value.(types.String).ValueString()
		if !validMember(text) {
			diags.AddError("Invalid suffix", "Each suffix must be non-empty, contain no whitespace, and be at most 253 characters.")
		}
		if seen[text] {
			diags.AddError("Duplicate suffix", "DNS suffix members must be unique.")
		}
		seen[text] = true
	}
}
func validMember(value string) bool {
	return value != "" && strings.IndexFunc(value, unicode.IsSpace) < 0 && len(utf16.Encode([]rune(value))) <= maxSuffixLength
}
