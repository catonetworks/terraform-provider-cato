package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// groupMembersSetModifier handles matching group members by ID+type or name+type
// to prevent "Provider produced inconsistent result after apply" errors when
// member names change but IDs remain the same
type groupMembersSetModifier struct{}

// Description returns a human-readable description of the plan modifier
func (m groupMembersSetModifier) Description(_ context.Context) string {
	return "Handles group member correlation by matching on ID+type or name+type, allowing names to update from API"
}

// MarkdownDescription returns a markdown description of the plan modifier
func (m groupMembersSetModifier) MarkdownDescription(_ context.Context) string {
	return "Handles group member correlation by matching on ID+type or name+type, allowing names to update from API"
}

// PlanModifySet implements the plan modification logic for Set attributes
func (m groupMembersSetModifier) PlanModifySet(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	tflog.Debug(ctx, "GroupMembersSetModifier: Plan modifier invoked")

	// If config, state, or plan is null/unknown, use default behavior
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() ||
		req.StateValue.IsNull() || req.StateValue.IsUnknown() ||
		req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		tflog.Debug(ctx, "GroupMembersSetModifier: Config, state, or plan is null/unknown, using default behavior")
		return
	}

	tflog.Debug(ctx, "GroupMembersSetModifier: Processing group members correlation")

	// Get elements from config, state, and plan
	configElements := req.ConfigValue.Elements()
	stateElements := req.StateValue.Elements()
	planElements := req.PlanValue.Elements()

	if len(configElements) == 0 || len(planElements) == 0 {
		tflog.Debug(ctx, "GroupMembersSetModifier: Empty config or plan, using default behavior")
		return
	}

	// Create a map of state elements for quick lookup
	// Key format: "type:id" when ID is present, "type:name" when only name is present
	stateMap := make(map[string]types.Object)
	for _, stateElement := range stateElements {
		stateObj, ok := stateElement.(types.Object)
		if !ok {
			continue
		}
		stateAttrs := stateObj.Attributes()
		memberType := getStringValue(stateAttrs, "type")
		memberID := getStringValue(stateAttrs, "id")
		memberName := getStringValue(stateAttrs, "name")

		if memberType != "" {
			if memberID != "" {
				key := memberType + ":id:" + memberID
				stateMap[key] = stateObj
			}
			if memberName != "" {
				key := memberType + ":name:" + memberName
				stateMap[key] = stateObj
			}
		}
	}

	// Process each plan element and try to preserve name from state or allow API update
	modifiedElements := make([]attr.Value, 0, len(planElements))
	for _, planElement := range planElements {
		planObj, ok := planElement.(types.Object)
		if !ok {
			modifiedElements = append(modifiedElements, planElement)
			continue
		}

		planAttrs := planObj.Attributes()
		memberType := getStringValue(planAttrs, "type")
		memberID := getStringValue(planAttrs, "id")
		memberName := getStringValue(planAttrs, "name")

		// Try to find corresponding state element
		var correspondingState *types.Object
		if memberType != "" && memberID != "" {
			// Match by type:id
			key := memberType + ":id:" + memberID
			if stateObj, exists := stateMap[key]; exists {
				correspondingState = &stateObj
			}
		} else if memberType != "" && memberName != "" {
			// Match by type:name
			key := memberType + ":name:" + memberName
			if stateObj, exists := stateMap[key]; exists {
				correspondingState = &stateObj
			}
		}

		// If we found a corresponding state element, check if we need to update
		if correspondingState != nil {
			stateAttrs := correspondingState.Attributes()
			stateName := getStringValue(stateAttrs, "name")
			stateID := getStringValue(stateAttrs, "id")

			// If the ID changed or name changed (from API), keep the planned values
			// The API will populate both fields correctly
			tflog.Debug(ctx, "GroupMembersSetModifier: Found corresponding state", map[string]interface{}{
				"plan_id":    memberID,
				"plan_name":  memberName,
				"state_id":   stateID,
				"state_name": stateName,
			})
		}

		// Always use the plan element as-is - the API hydration will populate both fields
		modifiedElements = append(modifiedElements, planElement)
	}

	// Create new set with modified elements
	if len(modifiedElements) > 0 {
		newSet, diags := types.SetValue(req.PlanValue.ElementType(ctx), modifiedElements)
		if !diags.HasError() {
			resp.PlanValue = newSet
			tflog.Debug(ctx, "GroupMembersSetModifier: Successfully processed group members")
		} else {
			tflog.Error(ctx, "GroupMembersSetModifier: Failed to create new set")
		}
	}
}

// getStringValue safely extracts a string value from an attribute map
func getStringValue(attrs map[string]attr.Value, key string) string {
	if val, exists := attrs[key]; exists {
		if strVal, ok := val.(types.String); ok && !strVal.IsNull() && !strVal.IsUnknown() {
			return strVal.ValueString()
		}
	}
	return ""
}

// GroupMembersSetModifier returns a new group members set plan modifier
func GroupMembersSetModifier() planmodifier.Set {
	return groupMembersSetModifier{}
}
