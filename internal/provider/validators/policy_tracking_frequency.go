package validators

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Tracking frequency validator
type (
	PolicyTrackingFrequency struct{}
)

func (v PolicyTrackingFrequency) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() {
		return
	}
	value := req.ConfigValue.ValueString()
	frequency := cato_models.PolicyRuleTrackingFrequencyEnum(value)
	if !frequency.IsValid() {
		resp.Diagnostics.AddError("Field validation error", fmt.Sprintf("invalid frequency (%s: %s)\n - valid options: %+v", req.Path.String(),
			value, cato_models.AllPolicyRuleTrackingFrequencyEnum))
		return
	}
}
func (v PolicyTrackingFrequency) Description(_ context.Context) string {
	return fmt.Sprintf("Frequency must be one of: %v", cato_models.AllPolicyRuleTrackingFrequencyEnum)
}
func (v PolicyTrackingFrequency) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
