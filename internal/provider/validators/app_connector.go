package validators

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type AppConTypeValidator struct{}

func (v AppConTypeValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() {
		return
	}
	value := req.ConfigValue.ValueString()
	connType := cato_models.ZtnaAppConnectorType(value)
	if !connType.IsValid() {
		resp.Diagnostics.AddError("Field validation error", fmt.Sprintf("invalid connector type (%s: %s)\n - valid options: %+v", req.Path.String(),
			value, cato_models.AllZtnaAppConnectorType))
		return
	}
}
func (v AppConTypeValidator) Description(ctx context.Context) string {
	return fmt.Sprintf("AppConnector type must be one of: %v", cato_models.AllZtnaAppConnectorType)
}
func (v AppConTypeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
