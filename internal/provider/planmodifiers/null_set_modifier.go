package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type NullSetMod struct{ elementType attr.Type }

func NullSetModifier(elementType attr.Type) *NullSetMod {
	return &NullSetMod{elementType: elementType}
}

func (m *NullSetMod) PlanModifySet(_ context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	// If the entire plan is null, the resource is planned for destruction.
	if req.ConfigValue.IsUnknown() || req.Plan.Raw.IsNull() {
		return
	}

	// If config is null, plan empty set
	if req.ConfigValue.IsNull() {
		resp.PlanValue, _ = types.SetValue(m.elementType, nil)
		return
	}
}

// Description returns a human-readable description of the plan modifier.
func (m NullSetMod) Description(_ context.Context) string {
	return "NullSetModifier plans null as an empty set"
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m NullSetMod) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}
