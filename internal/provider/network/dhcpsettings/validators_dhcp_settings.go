package dhcpsettings

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/planmodifiers"
)

// Omitted relay groups cannot be cleared by this API; retain their effective state.
type relayGroupModifier struct{}

func (relayGroupModifier) Description(context.Context) string {
	return "Preserve an omitted relay group; resolve configured IDs and names."
}
func (m relayGroupModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}
func (relayGroupModifier) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	if req.ConfigValue.IsNull() {
		if !req.StateValue.IsNull() {
			resp.PlanValue = req.StateValue
		}
		return
	}
	planmodifiers.IDNamePlanModifier().PlanModifyObject(ctx, req, resp)
}
