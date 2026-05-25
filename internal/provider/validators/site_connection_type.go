package validators

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// SiteConnectionTypeValidator validates that the provided string is a valid site connection type
type (
	SiteConnectionTypeValidator struct{}
)

func (v SiteConnectionTypeValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() {
		return
	}
	value := req.ConfigValue.ValueString()
	connType := cato_models.SiteConnectionTypeEnum(value)
	if !connType.IsValid() {
		resp.Diagnostics.AddError("Field validation error",
			fmt.Sprintf("invalid site connection type (%s: %s)\n - valid options: %+v", req.Path.String(),
				value, cato_models.AllSiteConnectionTypeEnum))
		return
	}
}

func (v SiteConnectionTypeValidator) Description(_ context.Context) string {
	return fmt.Sprintf("Site connection type must be one of: %v", cato_models.AllSiteConnectionTypeEnum)
}
func (v SiteConnectionTypeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
