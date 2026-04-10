package validators

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// RiskScoreOperator validator
type RiskScoreOperator struct{}

func (v RiskScoreOperator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() {
		return
	}
	value := req.ConfigValue.ValueString()
	action := cato_models.RiskScoreOperator(value)
	if !action.IsValid() {
		resp.Diagnostics.AddError("Field validation error", fmt.Sprintf("invalid risk score operator (%s: %s)\n - valid options: %+v", req.Path.String(),
			value, cato_models.AllRiskScoreOperator))
		return
	}
}
func (v RiskScoreOperator) Description(_ context.Context) string {
	return fmt.Sprintf("Policy active_on must be one of: %v", cato_models.AllRiskScoreOperator)
}
func (v RiskScoreOperator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
