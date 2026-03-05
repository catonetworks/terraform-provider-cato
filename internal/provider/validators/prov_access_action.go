package validators

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Action validator
type PrivAccPolicyActionValidator struct{}

func (v PrivAccPolicyActionValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() {
		return
	}
	value := req.ConfigValue.ValueString()
	action := cato_models.PrivateAccessPolicyActionEnum(value)
	if !action.IsValid() {
		resp.Diagnostics.AddError("Field validation error", fmt.Sprintf("invalid action (%s: %s)\n - valid options: %+v", req.Path.String(),
			value, cato_models.AllPrivateAccessPolicyActionEnum))
		return
	}
}
func (v PrivAccPolicyActionValidator) Description(ctx context.Context) string {
	return fmt.Sprintf("PrivatAccessPolicy action must be one of: %v", cato_models.AllPrivateAccessPolicyActionEnum)
}
func (v PrivAccPolicyActionValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
