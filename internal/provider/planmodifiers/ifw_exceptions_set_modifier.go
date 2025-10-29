package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// ifwExceptionsSetModifier handles the specific case where Internet Firewall exceptions set elements
// have nested ID fields that transition from unknown to known values during apply.
// This prevents "Provider produced inconsistent result after apply" errors.
type ifwExceptionsSetModifier struct{}

// Description returns a human-readable description of the plan modifier
func (m ifwExceptionsSetModifier) Description(_ context.Context) string {
	return "Handles Internet Firewall exceptions set element correlation when nested ID fields change from unknown to known"
}

// MarkdownDescription returns a markdown description of the plan modifier
func (m ifwExceptionsSetModifier) MarkdownDescription(_ context.Context) string {
	return "Handles Internet Firewall exceptions set element correlation when nested ID fields change from unknown to known"
}

// PlanModifySet implements the plan modification logic for Set attributes
func (m ifwExceptionsSetModifier) PlanModifySet(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	// Log entry into the plan modifier
	tflog.Warn(ctx, "IfwExceptionsSetModifier: Plan modifier invoked")

	// If config is null or unknown, use default behavior
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		tflog.Debug(ctx, "IfwExceptionsSetModifier: Config is null or unknown, using default behavior")
		return
	}

	// If state doesn't exist (first apply), use plan as-is but log it
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		tflog.Debug(ctx, "IfwExceptionsSetModifier: State is null or unknown (first apply), using plan as-is")
		return
	}

	// If plan is null or unknown, use default behavior
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		tflog.Debug(ctx, "IfwExceptionsSetModifier: Plan is null or unknown, using default behavior")
		return
	}

	// Log the types of values we're processing
	tflog.Debug(ctx, "IfwExceptionsSetModifier: Value types", map[string]interface{}{
		"plan_type":   req.PlanValue.Type(ctx).String(),
		"state_type":  req.StateValue.Type(ctx).String(),
		"config_type": req.ConfigValue.Type(ctx).String(),
	})

	tflog.Warn(ctx, "IfwExceptionsSetModifier: Processing exceptions set correlation")

	// Get the planned value elements
	plannedElements := req.PlanValue.Elements()
	stateElements := req.StateValue.Elements()
	configElements := req.ConfigValue.Elements()

	tflog.Warn(ctx, "IfwExceptionsSetModifier: Element counts", map[string]interface{}{
		"planned": len(plannedElements),
		"state":   len(stateElements),
		"config":  len(configElements),
	})

	// If the number of elements differs significantly, let the default behavior handle it
	// Allow for small variations due to ordering or correlation issues
	if len(plannedElements) != len(stateElements) {
		tflog.Warn(ctx, "IfwExceptionsSetModifier: Element count mismatch between plan and state, using default behavior", map[string]interface{}{
			"planned_count": len(plannedElements),
			"state_count":   len(stateElements),
		})
		return
	}

	// If there are no state elements to correlate with, use plan as-is
	if len(stateElements) == 0 {
		tflog.Debug(ctx, "IfwExceptionsSetModifier: No state elements to correlate with")
		return
	}

	// Create a new set of elements where we preserve state IDs for matching elements
	modifiedElements := make([]attr.Value, 0, len(plannedElements))
	correlationSuccessCount := 0

	for i, plannedElement := range plannedElements {
		plannedObj, ok := plannedElement.(types.Object)
		if !ok {
			tflog.Debug(ctx, "IfwExceptionsSetModifier: Planned element is not an object, using as-is", map[string]interface{}{"index": i})
			modifiedElements = append(modifiedElements, plannedElement)
			continue
		}

		// Find the corresponding state element by matching non-ID fields
		correspondingStateElement := m.findCorrespondingStateElement(ctx, plannedObj, stateElements)

		if correspondingStateElement != nil {
			tflog.Debug(ctx, "IfwExceptionsSetModifier: Found corresponding state element", map[string]interface{}{"index": i})
			// Preserve the object with state-derived ID values for nested objects
			modifiedElement := m.preserveStateIds(ctx, plannedObj, *correspondingStateElement)
			modifiedElements = append(modifiedElements, modifiedElement)
			correlationSuccessCount++
		} else {
			tflog.Debug(ctx, "IfwExceptionsSetModifier: No corresponding state element found, using planned as-is", map[string]interface{}{"index": i})
			// No corresponding state element found, use planned as-is
			modifiedElements = append(modifiedElements, plannedObj)
		}
	}

	tflog.Warn(ctx, "IfwExceptionsSetModifier: Correlation summary", map[string]interface{}{
		"total_elements":     len(plannedElements),
		"correlations_found": correlationSuccessCount,
	})

	// Create the modified set
	if len(modifiedElements) > 0 {
		modifiedSet, diags := types.SetValue(req.PlanValue.ElementType(ctx), modifiedElements)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			resp.PlanValue = modifiedSet
			tflog.Warn(ctx, "IfwExceptionsSetModifier: Successfully modified exceptions set")
		} else {
			tflog.Error(ctx, "IfwExceptionsSetModifier: Failed to create modified set, using original plan")
		}
	}
}

// findCorrespondingStateElement attempts to find the state element that corresponds
// to the given planned element by matching on non-ID identifying fields
func (m ifwExceptionsSetModifier) findCorrespondingStateElement(ctx context.Context, plannedObj types.Object, stateElements []attr.Value) *types.Object {
	plannedAttrs := plannedObj.Attributes()

	// Use exception name as the primary identifier
	plannedName, nameExists := plannedAttrs["name"]
	if !nameExists {
		return nil
	}

	plannedNameStr, ok := plannedName.(types.String)
	if !ok || plannedNameStr.IsNull() || plannedNameStr.IsUnknown() {
		return nil
	}

	for _, stateElement := range stateElements {
		stateObj, ok := stateElement.(types.Object)
		if !ok {
			continue
		}

		stateAttrs := stateObj.Attributes()
		stateName, exists := stateAttrs["name"]
		if !exists {
			continue
		}

		stateNameStr, ok := stateName.(types.String)
		if !ok {
			continue
		}

		// Match by name (primary identifier for exceptions)
		if !stateNameStr.IsNull() && !stateNameStr.IsUnknown() &&
			stateNameStr.ValueString() == plannedNameStr.ValueString() {
			return &stateObj
		}
	}

	return nil
}

// preserveStateIds creates a new object that preserves state ID values for nested objects
// while keeping the planned values for other fields
func (m ifwExceptionsSetModifier) preserveStateIds(ctx context.Context, plannedObj types.Object, stateObj types.Object) types.Object {
	plannedAttrs := plannedObj.Attributes()
	stateAttrs := stateObj.Attributes()

	// Start with planned attributes
	newAttrs := make(map[string]attr.Value, len(plannedAttrs))
	for k, v := range plannedAttrs {
		newAttrs[k] = v
	}

	// Preserve nested object IDs from state for Internet Firewall specific fields
	m.preserveNestedObjectIds(ctx, newAttrs, stateAttrs, "source")
	m.preserveNestedObjectIds(ctx, newAttrs, stateAttrs, "destination")
	m.preserveNestedObjectIds(ctx, newAttrs, stateAttrs, "service")

	// // Preserve device_attributes object from state if it exists
	// if stateDeviceAttrs, exists := stateAttrs["device_attributes"]; exists {
	// 	if stateDeviceAttrsObj, ok := stateDeviceAttrs.(types.Object); ok && !stateDeviceAttrsObj.IsNull() && !stateDeviceAttrsObj.IsUnknown() {
	// 		newAttrs["device_attributes"] = stateDeviceAttrsObj
	// 		tflog.Debug(ctx, "IfwExceptionsSetModifier: Preserved device_attributes from state")
	// 	}
	// }

	// Create the new object
	objectType := plannedObj.Type(ctx).(types.ObjectType)
	newObj, diags := types.ObjectValue(objectType.AttrTypes, newAttrs)
	if diags.HasError() {
		tflog.Error(ctx, "IfwExceptionsSetModifier: Failed to create preserved object, using planned")
		return plannedObj
	}

	return newObj
}

// preserveNestedObjectIds preserves ID values in nested objects (like source.host, destination.application, etc.)
func (m ifwExceptionsSetModifier) preserveNestedObjectIds(ctx context.Context, newAttrs map[string]attr.Value, stateAttrs map[string]attr.Value, nestedFieldName string) {
	plannedNested, exists := newAttrs[nestedFieldName]
	if !exists {
		return
	}

	stateNested, stateExists := stateAttrs[nestedFieldName]
	if !stateExists {
		return
	}

	// Handle nested object with sets (like source, destination, service)
	plannedNestedObj, ok := plannedNested.(types.Object)
	if !ok {
		return
	}

	stateNestedObj, ok := stateNested.(types.Object)
	if !ok {
		return
	}

	plannedNestedAttrs := plannedNestedObj.Attributes()
	stateNestedAttrs := stateNestedObj.Attributes()

	newNestedAttrs := make(map[string]attr.Value, len(plannedNestedAttrs))
	for k, v := range plannedNestedAttrs {
		newNestedAttrs[k] = v
	}

	// Preserve IDs in nested sets (like host, site, etc.)
	// Only process attributes that are in the planned config and are not null
	for attrName, plannedAttrValue := range plannedNestedAttrs {
		// Skip if the planned attribute is null - we don't want to add state values for null plan values
		if plannedSet, ok := plannedAttrValue.(types.Set); ok && plannedSet.IsNull() {
			continue
		}
		m.preserveSetElementIds(ctx, newNestedAttrs, stateNestedAttrs, attrName)
	}

	// Recreate the nested object
	nestedObjectType := plannedNestedObj.Type(ctx).(types.ObjectType)
	newNestedObj, diags := types.ObjectValue(nestedObjectType.AttrTypes, newNestedAttrs)
	if !diags.HasError() {
		newAttrs[nestedFieldName] = newNestedObj
	}

	// // Preserve IDs in nested sets - Internet Firewall specific field names
	// ifwSetFields := []string{
	// 	"host", "site", "global_ip_range", "network_interface",
	// 	"site_network_subnet", "floating_subnet", "group",
	// 	"system_group", "user", "users_group", "application",
	// 	"custom_app", "app_category", "custom_category",
	// 	"sanctioned_apps_category", "country", "standard",
	// }

	// for _, fieldName := range ifwSetFields {
	// 	m.preserveSetElementIds(ctx, newNestedAttrs, stateNestedAttrs, fieldName)
	// }

	// // Recreate the nested object
	// nestedObjectType := plannedNestedObj.Type(ctx).(types.ObjectType)
	// newNestedObj, diags := types.ObjectValue(nestedObjectType.AttrTypes, newNestedAttrs)
	// if !diags.HasError() {
	// 	newAttrs[nestedFieldName] = newNestedObj
	// }
}

// preserveSetElementIds preserves ID values within set elements by matching on name
func (m ifwExceptionsSetModifier) preserveSetElementIds(ctx context.Context, newNestedAttrs map[string]attr.Value, stateNestedAttrs map[string]attr.Value, setFieldName string) {
	plannedSet, exists := newNestedAttrs[setFieldName]
	if !exists {
		// Explicitly handle the case where a set is present in state but null in plan
		// This prevents drift when a set like global_ip_range goes from set to null
		_, stateExists := stateNestedAttrs[setFieldName]
		if stateExists {
			// If state has it but plan doesn't, we might need to preserve the null value
			// Log this situation for debugging
			tflog.Debug(ctx, "ExceptionsSetModifier: Set exists in state but not in plan",
				map[string]interface{}{"field": setFieldName})
		}
		return
	}

	stateSet, stateExists := stateNestedAttrs[setFieldName]
	if !stateExists {
		return
	}

	plannedSetVal, ok := plannedSet.(types.Set)
	if !ok {
		return
	}

	stateSetVal, ok := stateSet.(types.Set)
	if !ok {
		return
	}

	// Explicitly handle null values
	if plannedSetVal.IsNull() {
		// If planned is null but state has a value, log this
		if !stateSetVal.IsNull() {
			tflog.Debug(ctx, "ExceptionsSetModifier: Planned set is null but state has value",
				map[string]interface{}{"field": setFieldName})
		}
		return
	}

	if plannedSetVal.IsUnknown() || stateSetVal.IsNull() || stateSetVal.IsUnknown() {
		return
	}

	plannedElements := plannedSetVal.Elements()
	stateElements := stateSetVal.Elements()

	modifiedElements := make([]attr.Value, 0, len(plannedElements))

	// For each planned element, try to find corresponding state element and preserve its ID
	for _, plannedElement := range plannedElements {
		plannedElementObj, ok := plannedElement.(types.Object)
		if !ok {
			modifiedElements = append(modifiedElements, plannedElement)
			continue
		}

		// Find corresponding state element by name
		correspondingStateElement := m.findElementByName(ctx, plannedElementObj, stateElements)
		if correspondingStateElement != nil {
			// Create new element with preserved ID
			modifiedElement := m.preserveElementId(ctx, plannedElementObj, *correspondingStateElement)
			modifiedElements = append(modifiedElements, modifiedElement)
		} else {
			modifiedElements = append(modifiedElements, plannedElement)
		}
	}

	// Create new set with preserved IDs
	if len(modifiedElements) > 0 {
		newSet, diags := types.SetValue(plannedSetVal.ElementType(ctx), modifiedElements)
		if !diags.HasError() {
			newNestedAttrs[setFieldName] = newSet
			tflog.Debug(ctx, "IfwExceptionsSetModifier: Preserved IDs for set field", map[string]interface{}{"field": setFieldName})
		}
	}
}

// findElementByName finds an element in the list by matching the name field
func (m ifwExceptionsSetModifier) findElementByName(ctx context.Context, targetObj types.Object, elements []attr.Value) *types.Object {
	targetAttrs := targetObj.Attributes()
	targetName, exists := targetAttrs["name"]
	if !exists {
		return nil
	}

	targetNameStr, ok := targetName.(types.String)
	if !ok || targetNameStr.IsNull() || targetNameStr.IsUnknown() {
		return nil
	}

	for _, element := range elements {
		elementObj, ok := element.(types.Object)
		if !ok {
			continue
		}

		elementAttrs := elementObj.Attributes()
		elementName, exists := elementAttrs["name"]
		if !exists {
			continue
		}

		elementNameStr, ok := elementName.(types.String)
		if !ok {
			continue
		}

		if !elementNameStr.IsNull() && !elementNameStr.IsUnknown() &&
			elementNameStr.ValueString() == targetNameStr.ValueString() {
			return &elementObj
		}
	}

	return nil
}

// preserveElementId creates a new element object with ID and name preserved from state
func (m ifwExceptionsSetModifier) preserveElementId(ctx context.Context, plannedObj types.Object, stateObj types.Object) types.Object {
	plannedAttrs := plannedObj.Attributes()
	stateAttrs := stateObj.Attributes()

	newAttrs := make(map[string]attr.Value, len(plannedAttrs))
	for k, v := range plannedAttrs {
		newAttrs[k] = v
	}

	// Preserve ID from state if it exists and is known
	if stateId, exists := stateAttrs["id"]; exists {
		if stateIdStr, ok := stateId.(types.String); ok && !stateIdStr.IsNull() && !stateIdStr.IsUnknown() {
			newAttrs["id"] = stateIdStr
		}
	}

	// Preserve name from state if it exists and is known
	if stateName, exists := stateAttrs["name"]; exists {
		if stateNameStr, ok := stateName.(types.String); ok && !stateNameStr.IsNull() && !stateNameStr.IsUnknown() {
			newAttrs["name"] = stateNameStr
		}
	}

	objectType := plannedObj.Type(ctx).(types.ObjectType)
	newObj, diags := types.ObjectValue(objectType.AttrTypes, newAttrs)
	if diags.HasError() {
		return plannedObj
	}

	return newObj
}

// IfwExceptionsSetModifier returns a new Internet Firewall exceptions set plan modifier
func IfwExceptionsSetModifier() planmodifier.Set {
	return ifwExceptionsSetModifier{}
}
