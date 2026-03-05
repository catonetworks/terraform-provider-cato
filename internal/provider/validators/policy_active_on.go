package validators

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Active-on validator
type PolicyActiveOnValidator struct{}

func (v PolicyActiveOnValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() {
		return
	}
	value := req.ConfigValue.ValueString()
	action := cato_models.PolicyActiveOnEnum(value)
	if !action.IsValid() {
		resp.Diagnostics.AddError("Field validation error", fmt.Sprintf("invalid active_on (%s: %s)\n - valid options: %+v", req.Path.String(),
			value, cato_models.AllPolicyActiveOnEnum))
		return
	}
}
func (v PolicyActiveOnValidator) Description(_ context.Context) string {
	return fmt.Sprintf("Policy active_on must be one of: %v", cato_models.AllPolicyActiveOnEnum)
}
func (v PolicyActiveOnValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
