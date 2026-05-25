package validators

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/catonetworks/terraform-provider-cato/internal/utils"
)

// DHCPTypeValidator validates that the provided string is a valid DHCP type
type (
	DHCPTypeValidator struct{}
)

func (v DHCPTypeValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if !utils.HasValue(req.ConfigValue) {
		return
	}
	value := req.ConfigValue.ValueString()
	dhcpType := cato_models.DhcpType(value)
	if !dhcpType.IsValid() {
		resp.Diagnostics.AddError("Field validation error",
			fmt.Sprintf("invalid DHCP type (%s: %s)\n - valid options: %+v", req.Path.String(),
				value, cato_models.AllDhcpType))
		return
	}
}

func (v DHCPTypeValidator) Description(_ context.Context) string {
	return fmt.Sprintf("DHCP type must be one of: %v", cato_models.AllDhcpType)
}
func (v DHCPTypeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
