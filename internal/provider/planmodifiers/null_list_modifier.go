package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// NullListMod is the types.List counterpart of NullSetMod: it plans a null config
// value as an empty list so cleared collections are sent to the API as empty rather
// than absent.
type NullListMod struct{ elementType attr.Type }

func NullListModifier(elementType attr.Type) *NullListMod {
	return &NullListMod{elementType: elementType}
}

func (m *NullListMod) PlanModifyList(_ context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	// If the entire plan is null, the resource is planned for destruction.
	if req.ConfigValue.IsUnknown() || req.Plan.Raw.IsNull() {
		return
	}

	// If config is null, plan empty list
	if req.ConfigValue.IsNull() {
		resp.PlanValue, _ = types.ListValue(m.elementType, nil)
		return
	}
}

// Description returns a human-readable description of the plan modifier.
func (m NullListMod) Description(_ context.Context) string {
	return "NullListModifier plans null as an empty list"
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m NullListMod) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}
