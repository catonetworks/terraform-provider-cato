package validators

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Validators
type (
	IPProtocolValidator struct{}
)

func (v IPProtocolValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() {
		return
	}
	value := req.ConfigValue.ValueString()
	protoName := cato_models.IPProtocol(value)
	if !protoName.IsValid() {
		resp.Diagnostics.AddError("Field validation error", fmt.Sprintf("invalid protocol (%s: %s)\n - valid options: %+v", req.Path.String(),
			value, cato_models.AllIPProtocol))
		return
	}
}
func (v IPProtocolValidator) Description(_ context.Context) string {
	return fmt.Sprintf("Protocol must be one of: %v", cato_models.AllIPProtocol)
}
func (v IPProtocolValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
