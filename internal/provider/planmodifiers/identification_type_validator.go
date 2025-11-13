package planmodifiers

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// IdentificationTypeValidator returns a plan modifier that validates identification_type
// is only set when connection_mode is RESPONDER_ONLY
func IdentificationTypeValidator() planmodifier.String {
	return &identificationTypeValidator{}
}

type identificationTypeValidator struct{}

func (m *identificationTypeValidator) Description(ctx context.Context) string {
	return "Validates that identification_type is only set when connection_mode is RESPONDER_ONLY"
}

func (m *identificationTypeValidator) MarkdownDescription(ctx context.Context) string {
	return "Validates that `identification_type` is only set when `connection_mode` is `RESPONDER_ONLY`"
}

func (m *identificationTypeValidator) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// If identification_type is null or unknown, no validation needed
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}

	// Get the connection_mode value from the plan
	var connectionMode types.String
	diags := req.Plan.GetAttribute(ctx, path.Root("ipsec").AtName("connection_mode"), &connectionMode)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If identification_type has a value, validate that connection_mode is RESPONDER_ONLY
	if !req.PlanValue.IsNull() && !req.PlanValue.IsUnknown() {
		// Check if connection_mode is set and equals RESPONDER_ONLY
		if connectionMode.IsNull() || connectionMode.ValueString() != "RESPONDER_ONLY" {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid Configuration",
				fmt.Sprintf("identification_type can only be set when connection_mode is RESPONDER_ONLY, but connection_mode is %q", connectionMode.ValueString()),
			)
			return
		}
	}
}
