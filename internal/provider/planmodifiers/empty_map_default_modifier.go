package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// emptyMapDefaultModifier provides a default empty map when config and state are null.
type emptyMapDefaultModifier struct {
	elementType attr.Type
}

// Description returns a human-readable description of the plan modifier.
func (m emptyMapDefaultModifier) Description(_ context.Context) string {
	return "Provides default empty map when config and state are null"
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m emptyMapDefaultModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

// PlanModifyMap implements the plan modification logic for Map attributes.
func (m emptyMapDefaultModifier) PlanModifyMap(_ context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	// If the entire plan is null, the resource is planned for destruction.
	if req.ConfigValue.IsUnknown() || req.Plan.Raw.IsNull() {
		return
	}

	// If config is null, plan empty set
	if req.ConfigValue.IsNull() {
		resp.PlanValue, _ = types.MapValue(m.elementType, nil)
		return
	}
}

// EmptyMapDefault returns a plan modifier that defaults to an empty map when config and state are null.
func EmptyMapDefault(elementType attr.Type) planmodifier.Map {
	return emptyMapDefaultModifier{
		elementType: elementType,
	}
}
