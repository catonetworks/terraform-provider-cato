package validators

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/catonetworks/terraform-provider-cato/internal/utils"
)

// SocketInterfaceIndexValidator validates that the provided string is a valid socket interface index
type (
	SocketInterfaceIndexValidator struct{}
)

func (v SocketInterfaceIndexValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if !utils.HasValue(req.ConfigValue) {
		return
	}
	value := req.ConfigValue.ValueString()
	ifaceID := cato_models.SocketInterfaceIDEnum(value)
	if !ifaceID.IsValid() {
		resp.Diagnostics.AddError("Field validation error",
			fmt.Sprintf("invalid socket interface index (%s: %s)\n - valid options: %+v", req.Path.String(),
				value, cato_models.AllSocketInterfaceIDEnum))
		return
	}
}

func (v SocketInterfaceIndexValidator) Description(_ context.Context) string {
	return fmt.Sprintf("Socket interface index must be one of: %v", cato_models.AllSocketInterfaceIDEnum)
}
func (v SocketInterfaceIndexValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
