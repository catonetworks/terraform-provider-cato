package validators

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/catonetworks/terraform-provider-cato/internal/utils"
)

// SocketInterfaceDestTypeValidator validates that the provided string is a valid socket interface destination type
type (
	SocketInterfaceDestTypeValidator struct{}
)

func (v SocketInterfaceDestTypeValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if !utils.HasValue(req.ConfigValue) {
		return
	}
	value := req.ConfigValue.ValueString()
	ifaceDestType := cato_models.SocketInterfaceDestType(value)
	if !ifaceDestType.IsValid() {
		resp.Diagnostics.AddError("Field validation error",
			fmt.Sprintf("invalid interface destination type (%s: %s)\n - valid options: %+v", req.Path.String(),
				value, cato_models.AllSocketInterfaceDestType))
		return
	}
}

func (v SocketInterfaceDestTypeValidator) Description(_ context.Context) string {
	return fmt.Sprintf("Socket interface destination type must be one of: %v", cato_models.AllSocketInterfaceDestType)
}
func (v SocketInterfaceDestTypeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
