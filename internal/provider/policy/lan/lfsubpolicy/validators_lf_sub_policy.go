package lfsubpolicy

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

type configuredPositionModifier struct{}

func (configuredPositionModifier) Description(_ context.Context) string {
	return "Allow position updates, but reject removing an already-configured position."
}

func (m configuredPositionModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (configuredPositionModifier) PlanModifyObject(
	_ context.Context,
	req planmodifier.ObjectRequest,
	resp *planmodifier.ObjectResponse,
) {
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}
	if req.PlanValue.IsNull() {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"LAN sub-policy position cannot be removed",
			"The at attribute may be omitted after import, but cannot be removed after it has been configured.",
		)
		return
	}
}
