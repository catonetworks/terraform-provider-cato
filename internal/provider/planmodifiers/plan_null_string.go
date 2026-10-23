package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NullStringModifier() *nullStringModifier { return &nullStringModifier{} }

type nullStringModifier struct{}

func (m *nullStringModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	// If the entire plan is null, the resource is planned for destruction.
	if req.ConfigValue.IsUnknown() || req.Plan.Raw.IsNull() {
		return
	}

	// If config is null, plan StringValue("")
	if req.ConfigValue.IsNull() {
		resp.PlanValue = types.StringValue("")
		return
	}

	resp.PlanValue = req.ConfigValue
}

// Description returns a human-readable description of the plan modifier.
func (m nullStringModifier) Description(_ context.Context) string {
	return "NullStringModifier plans null as an empty string"
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m nullStringModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}
