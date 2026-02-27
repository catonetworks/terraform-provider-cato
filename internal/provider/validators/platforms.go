package validators

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/parse"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Platform validator
type PlatformValidator struct{}

func (v PlatformValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	var diags diag.Diagnostics
	if req.ConfigValue.IsUnknown() {
		return
	}

	platforms := parse.PrepareStrings[cato_models.OperatingSystem](ctx, req.ConfigValue, &diags, "platform")
	if diags.HasError() {
		resp.Diagnostics = append(resp.Diagnostics, diags...)
		return
	}

	for _, platform := range platforms {
		if !platform.IsValid() {
			resp.Diagnostics.AddError("Field validation error", fmt.Sprintf("invalid platform (%s: %s)\n - valid options: %+v", req.Path.String(),
				platform, cato_models.AllOperatingSystem))
			return
		}
	}
}

func (v PlatformValidator) Description(_ context.Context) string {
	return fmt.Sprintf("Platforms must be one of: %v", cato_models.AllOperatingSystem)
}
func (v PlatformValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
