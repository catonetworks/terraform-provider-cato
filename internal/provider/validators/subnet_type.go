package validators

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/catonetworks/terraform-provider-cato/internal/utils"
)

// SubnetTypeValidator validates that the provided string is a valid subnet (range) type
type (
	SubnetTypeValidator struct{}
)

func (v SubnetTypeValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if !utils.HasValue(req.ConfigValue) {
		return
	}
	value := req.ConfigValue.ValueString()
	subnetType := cato_models.SubnetType(value)
	if !subnetType.IsValid() {
		resp.Diagnostics.AddError("Field validation error",
			fmt.Sprintf("invalid subnet type (%s: %s)\n - valid options: %+v", req.Path.String(),
				value, cato_models.AllSubnetType))
		return
	}
}

func (v SubnetTypeValidator) Description(_ context.Context) string {
	return fmt.Sprintf("Subnet type must be one of: %v", cato_models.AllSubnetType)
}
func (v SubnetTypeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
