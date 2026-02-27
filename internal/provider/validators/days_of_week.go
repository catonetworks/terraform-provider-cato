package validators

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/parse"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// DaysValidator validate days of week
type DaysValidator struct{}

func (v DaysValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	var diags diag.Diagnostics
	if req.ConfigValue.IsUnknown() {
		return
	}

	days := parse.PrepareStrings[cato_models.DayOfWeek](ctx, req.ConfigValue, &diags, "days")
	if diags.HasError() {
		resp.Diagnostics = append(resp.Diagnostics, diags...)
		return
	}

	for _, day := range days {
		if !day.IsValid() {
			resp.Diagnostics.AddError("Field validation error", fmt.Sprintf("invalid day (%s: %s)\n - valid options: %+v", req.Path.String(),
				day, cato_models.AllDayOfWeek))
			return
		}
	}
}

func (v DaysValidator) Description(_ context.Context) string {
	return fmt.Sprintf("Days must be one of: %v", cato_models.AllDayOfWeek)
}
func (v DaysValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
