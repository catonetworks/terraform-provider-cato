package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

type idNameListModifier struct{}

// IDNameListPlanModifier is the types.List counterpart of IDNameSetPlanModifier. It
// preserves the computed counterpart of an ID/name reference (the id when configured by
// name, or the name when configured by id) from the prior state when the configured
// selector matches a state member, so a name-only reference does not re-plan its Computed
// id as Unknown on every apply (which would churn the plan forever).
//
// Unlike the Set variant, this is used on a List attribute: Terraform correlates List
// elements by index, so an element whose Computed id is still Unknown (e.g. the FIRST
// create, when there is no prior state to resolve from) correlates cleanly after apply
// instead of raising "Provider produced inconsistent result after apply". List elements
// must use idname.RefAttrTypes and configure exactly one of name or ID.
func IDNameListPlanModifier() planmodifier.List {
	return idNameListModifier{}
}

func (m idNameListModifier) Description(_ context.Context) string {
	return "Preserves reference IDs and names only for matching configured references."
}

func (m idNameListModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m idNameListModifier) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() || req.PlanValue.IsNull() || req.PlanValue.IsUnknown() ||
		req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}
	// Preserve configured order (List is index-correlated). Each configured element is
	// resolved against state by its selector; if any element can't be resolved we leave
	// the whole plan untouched so an element keeps its Unknown computed value -- that is
	// still valid for a List (index correlation) even though it would break a Set.
	elements := make([]attr.Value, 0, len(req.ConfigValue.Elements()))
	for _, element := range req.ConfigValue.Elements() {
		value, diags, known := idNameListReferenceValue(element.(types.Object), req.StateValue)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() || !known {
			return
		}
		elements = append(elements, value)
	}

	value, diags := types.ListValue(req.PlanValue.ElementType(ctx), elements)
	resp.Diagnostics.Append(diags...)
	if !resp.Diagnostics.HasError() {
		resp.PlanValue = value
	}
}

func idNameListReferenceValue(configured types.Object, state types.List) (types.Object, diag.Diagnostics, bool) {
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
	attrs[computed] = matchingIDNameListValue(state, selector, attrs[selector], computed)
	value, diags := types.ObjectValue(idname.RefAttrTypes, attrs)
	return value, diags, true
}

func matchingIDNameListValue(state types.List, selector string, configured attr.Value, computed string) types.String {
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
