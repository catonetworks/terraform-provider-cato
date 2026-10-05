package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type userGroupReferencesModifier struct{}

// UserGroupReferencesModifier preserves the computed counterpart of a user
// group reference only when its configured name or ID matches a state member.
// Matching the entire set avoids relying on the framework's positional state
// values for nested elements, which can belong to a different group.
func UserGroupReferencesModifier() planmodifier.Set {
	return userGroupReferencesModifier{}
}

func (m userGroupReferencesModifier) Description(_ context.Context) string {
	return "Preserves user group IDs and names only for matching configured references."
}

func (m userGroupReferencesModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m userGroupReferencesModifier) PlanModifySet(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() || req.PlanValue.IsNull() || req.PlanValue.IsUnknown() ||
		req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}
	elements := make([]attr.Value, 0, len(req.ConfigValue.Elements()))
	for _, element := range req.ConfigValue.Elements() {
		value, diags, known := userGroupReferenceValue(ctx, element.(types.Object), req.StateValue)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() || !known {
			return
		}
		elements = append(elements, value)
	}

	value, diags := types.SetValue(req.PlanValue.ElementType(ctx), elements)
	resp.Diagnostics.Append(diags...)
	if !resp.Diagnostics.HasError() {
		resp.PlanValue = value
	}
}

func userGroupReferenceValue(ctx context.Context, configured types.Object, state types.Set) (types.Object, diag.Diagnostics, bool) {
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
	attrs[computed] = matchingUserGroupValue(state, selector, attrs[selector], computed)
	value, diags := types.ObjectValue(configured.AttributeTypes(ctx), attrs)
	return value, diags, true
}

func matchingUserGroupValue(state types.Set, selector string, configured attr.Value, computed string) types.String {
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
