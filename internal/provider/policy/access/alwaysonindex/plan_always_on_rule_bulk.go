package alwaysonindex

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type modTrigger struct{}

func newModChanges() *modTrigger { return &modTrigger{} }

func (m *modTrigger) PlanModifyInt64(_ context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	if req.ConfigValue.IsUnknown() {
		return
	}
	resp.PlanValue = types.Int64Value(0)
}

func (m *modTrigger) Description(_ context.Context) string {
	return "Modify policy trigger"
}
func (m *modTrigger) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}
