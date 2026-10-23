package planmodifiers

import (
	"context"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

type idNameSetModifier struct{}

// IDNameSetPlanModifier preserves the computed counterpart of an ID/name
// reference only when its configured name or ID matches a state member.
// Matching the entire set avoids relying on the framework's positional state
// values for nested elements, which can belong to a different reference.
// Set elements must use IDNameRefModelTypes and configure exactly one of name or ID.
func IDNameSetPlanModifier() planmodifier.Set {
	return idNameSetModifier{}
}

func (m idNameSetModifier) Description(_ context.Context) string {
	return "Preserves reference IDs and names only for matching configured references."
}

func (m idNameSetModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m idNameSetModifier) PlanModifySet(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() || req.PlanValue.IsNull() || req.PlanValue.IsUnknown() ||
		req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}
	elements := make([]attr.Value, 0, len(req.ConfigValue.Elements()))
	hasUnknown := false
	for _, element := range req.ConfigValue.Elements() {
		value, diags, known := idNameReferenceValue(element.(types.Object), req.StateValue)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() || !known {
			return
		}
		hasUnknown = hasUnknown || idNameReferenceHasUnknown(value)
		// A configured name and ID can resolve to the same state reference.
		if !slices.ContainsFunc(elements, value.Equal) {
			elements = append(elements, value)
		}
	}
	// Keep the original unknown elements when a new reference is unresolved.
	// Collapsing aliases in a partially unknown set would narrow its possible
	// cardinality below the configured count and produce an invalid plan.
	if hasUnknown && len(elements) < len(req.ConfigValue.Elements()) {
		return
	}

	value, diags := types.SetValue(req.PlanValue.ElementType(ctx), elements)
	resp.Diagnostics.Append(diags...)
	if !resp.Diagnostics.HasError() {
		resp.PlanValue = value
	}
}

func idNameReferenceHasUnknown(value types.Object) bool {
	attrs := value.Attributes()
	return attrs["name"].IsUnknown() || attrs["id"].IsUnknown()
}

func idNameReferenceValue(configured types.Object, state types.Set) (types.Object, diag.Diagnostics, bool) {
	if configured.IsNull() || configured.IsUnknown() {
		return types.Object{}, nil, false
	}
	attrs := configured.Attributes()
	name, id := attrs["name"].(types.String), attrs["id"].(types.String)
	if name.IsUnknown() || id.IsUnknown() {
		return types.Object{}, nil, false
	}
	selector, computed := "name", "id"
	if name.IsNull() {
		selector, computed = "id", "name"
	}
	if attrs[selector].IsNull() {
		return types.Object{}, nil, false
	}
	attrs[computed] = matchingIDNameValue(state, selector, attrs[selector], computed)
	value, diags := types.ObjectValue(idname.RefAttrTypes, attrs)
	return value, diags, true
}

func matchingIDNameValue(state types.Set, selector string, configured attr.Value, computed string) types.String {
	for _, element := range state.Elements() {
		object := element.(types.Object)
		if object.IsNull() || object.IsUnknown() {
			continue
		}
		attrs := object.Attributes()
		value := attrs[computed].(types.String)
		if configured.Equal(attrs[selector]) && !value.IsNull() && !value.IsUnknown() {
			return value
		}
	}
	return types.StringUnknown()
}
