package alwaysonsection

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

type sectionPositionValidator struct{}

func (sectionPositionValidator) Description(context.Context) string {
	return "AFTER_SECTION and BEFORE_SECTION require a non-empty reference section ID"
}

func (v sectionPositionValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (sectionPositionValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if !utils.HasValue(req.ConfigValue) {
		return
	}
	var position policyrule.Position
	resp.Diagnostics.Append(req.ConfigValue.As(ctx, &position, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() || !utils.HasValue(position.Position) || position.Ref.IsUnknown() {
		return
	}
	if (position.Position.ValueString() == "AFTER_SECTION" || position.Position.ValueString() == "BEFORE_SECTION") &&
		position.Ref.ValueString() == "" {
		resp.Diagnostics.AddAttributeError(req.Path.AtName("ref"), "Missing section reference",
			"A non-empty ref is required for AFTER_SECTION and BEFORE_SECTION.")
	}
}
