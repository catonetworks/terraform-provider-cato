package validators

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// RiscScoreCategory validator
type RiscScoreCategory struct{}

func (v RiscScoreCategory) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() {
		return
	}
	value := req.ConfigValue.ValueString()
	action := cato_models.RiskScoreCategory(value)
	if !action.IsValid() {
		resp.Diagnostics.AddError("Field validation error", fmt.Sprintf("invalid risk score category (%s: %s)\n - valid options: %+v", req.Path.String(),
			value, cato_models.AllRiskScoreCategory))
		return
	}
}
func (v RiscScoreCategory) Description(_ context.Context) string {
	return fmt.Sprintf("Policy active_on must be one of: %v", cato_models.AllRiskScoreCategory)
}
func (v RiscScoreCategory) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
