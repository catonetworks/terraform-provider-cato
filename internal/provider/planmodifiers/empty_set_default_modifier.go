package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// emptySetDefaultModifier provides a default empty set when the config value is null
type emptySetDefaultModifier struct {
	elementType attr.Type
}

// Description returns a human-readable description of the plan modifier
func (m emptySetDefaultModifier) Description(_ context.Context) string {
	return "Provides default empty set when config is null"
}

// MarkdownDescription returns a markdown description of the plan modifier
func (m emptySetDefaultModifier) MarkdownDescription(_ context.Context) string {
	return "Provides default empty set when config is null"
}

// PlanModifySet implements the plan modification logic for Set attributes
func (m emptySetDefaultModifier) PlanModifySet(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	// Only apply default if config is null and this is a new resource (state is null)
	if req.ConfigValue.IsNull() && req.StateValue.IsNull() {
		emptySet := types.SetValueMust(m.elementType, []attr.Value{})
		resp.PlanValue = emptySet
		tflog.Debug(ctx, "EmptySetDefaultModifier: Applied default empty set")
	}
}

// EmptySetDefault returns a plan modifier that sets an empty set as default when config is null
func EmptySetDefault(elementType attr.Type) planmodifier.Set {
	return emptySetDefaultModifier{
		elementType: elementType,
	}
}
