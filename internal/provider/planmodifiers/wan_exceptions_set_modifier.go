package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// exceptionsSetModifier handles the specific case where exceptions set elements
// have nested ID fields that transition from unknown to known values during apply.
// This prevents "Provider produced inconsistent result after apply" errors.
type wanExceptionsSetModifier struct{}

// Description returns a human-readable description of the plan modifier
func (m wanExceptionsSetModifier) Description(_ context.Context) string {
	return "Handles WAN exceptions set element correlation when nested ID fields change from unknown to known"
}

// MarkdownDescription returns a markdown description of the plan modifier
func (m wanExceptionsSetModifier) MarkdownDescription(_ context.Context) string {
	return "Handles WAN exceptions set element correlation when nested ID fields change from unknown to known"
}

// PlanModifySet implements the plan modification logic for Set attributes
func (m wanExceptionsSetModifier) PlanModifySet(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	// Log entry into the plan modifier
	tflog.Warn(ctx, "ExceptionsSetModifier: Plan modifier invoked")

	// If config is null or unknown, use default behavior
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		tflog.Debug(ctx, "ExceptionsSetModifier: Config is null or unknown, using default behavior")
		return
	}

	// If plan is null or unknown, use default behavior
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		tflog.Debug(ctx, "ExceptionsSetModifier: Plan is null or unknown, using default behavior")
		return
	}

	// Log the types of values we're processing
	tflog.Debug(ctx, "ExceptionsSetModifier: Value types", map[string]interface{}{
		"plan_type":   req.PlanValue.Type(ctx).String(),
		"state_type":  req.StateValue.Type(ctx).String(),
		"config_type": req.ConfigValue.Type(ctx).String(),
	})

	// Get the planned value elements
	plannedElements := req.PlanValue.Elements()
	stateElements := []attr.Value{}
	if !req.StateValue.IsNull() && !req.StateValue.IsUnknown() {
		stateElements = req.StateValue.Elements()
	}
	configElements := req.ConfigValue.Elements()

	tflog.Warn(ctx, "ExceptionsSetModifier: Element counts", map[string]interface{}{
		"planned": len(plannedElements),
		"state":   len(stateElements),
		"config":  len(configElements),
	})

	// Handle the case where state is empty (first apply or new exceptions being added)
	// In this case, we need to resolve Unknown values to Null to enable correlation
	if len(stateElements) == 0 {
		tflog.Warn(ctx, "ExceptionsSetModifier: No state elements (first apply), resolving Unknown values")
		modifiedElements := make([]attr.Value, 0, len(plannedElements))
		for _, plannedElement := range plannedElements {
			plannedObj, ok := plannedElement.(types.Object)
			if !ok {
				modifiedElements = append(modifiedElements, plannedElement)
				continue
			}
			// Resolve Unknown values to Null for first apply
			modifiedElement := m.resolveUnknownToNull(ctx, plannedObj)
			modifiedElements = append(modifiedElements, modifiedElement)
		}
		if len(modifiedElements) > 0 {
			modifiedSet, diags := types.SetValue(req.PlanValue.ElementType(ctx), modifiedElements)
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				resp.PlanValue = modifiedSet
				tflog.Warn(ctx, "ExceptionsSetModifier: Resolved Unknown values for first apply")
			}
		}
		return
	}

	tflog.Warn(ctx, "ExceptionsSetModifier: Processing exceptions set correlation")

	// Create a new set of elements where we preserve state IDs for matching elements
	modifiedElements := make([]attr.Value, 0, len(plannedElements))
	correlationSuccessCount := 0
	newExceptionsCount := 0

	for i, plannedElement := range plannedElements {
		plannedObj, ok := plannedElement.(types.Object)
		if !ok {
			tflog.Debug(ctx, "ExceptionsSetModifier: Planned element is not an object, using as-is", map[string]interface{}{"index": i})
			modifiedElements = append(modifiedElements, plannedElement)
			continue
		}

		// Find the corresponding state element by matching non-ID fields
		correspondingStateElement := m.findCorrespondingStateElement(ctx, plannedObj, stateElements)

		if correspondingStateElement != nil {
			tflog.Debug(ctx, "ExceptionsSetModifier: Found corresponding state element", map[string]interface{}{"index": i})
			// Preserve the object with state-derived ID values for nested objects
			modifiedElement := m.preserveStateIds(ctx, plannedObj, *correspondingStateElement)
			modifiedElements = append(modifiedElements, modifiedElement)
			correlationSuccessCount++
		} else {
			tflog.Debug(ctx, "ExceptionsSetModifier: No corresponding state element found (new exception), resolving Unknown values", map[string]interface{}{"index": i})
			// No corresponding state element found - this is a new exception being added
			// Resolve Unknown values to Null to enable proper correlation after apply
			modifiedElement := m.resolveUnknownToNull(ctx, plannedObj)
			modifiedElements = append(modifiedElements, modifiedElement)
			newExceptionsCount++
		}
	}

	tflog.Warn(ctx, "ExceptionsSetModifier: Correlation summary", map[string]interface{}{
		"total_elements":     len(plannedElements),
		"correlations_found": correlationSuccessCount,
		"new_exceptions":     newExceptionsCount,
	})

	// Create the modified set
	if len(modifiedElements) > 0 {
		modifiedSet, diags := types.SetValue(req.PlanValue.ElementType(ctx), modifiedElements)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			resp.PlanValue = modifiedSet
			tflog.Warn(ctx, "ExceptionsSetModifier: Successfully modified exceptions set")
		} else {
			tflog.Error(ctx, "ExceptionsSetModifier: Failed to create modified set, using original plan")
		}
	}
}

// findCorrespondingStateElement attempts to find the state element that corresponds
// to the given planned element by matching on non-ID identifying fields
func (m wanExceptionsSetModifier) findCorrespondingStateElement(ctx context.Context, plannedObj types.Object, stateElements []attr.Value) *types.Object {
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
func (m wanExceptionsSetModifier) preserveStateIds(ctx context.Context, plannedObj types.Object, stateObj types.Object) types.Object {
	plannedAttrs := plannedObj.Attributes()
	stateAttrs := stateObj.Attributes()

	// Start with planned attributes
	newAttrs := make(map[string]attr.Value, len(plannedAttrs))
	for k, v := range plannedAttrs {
		newAttrs[k] = v
	}

	// Preserve nested object IDs from state
	m.preserveNestedObjectIds(ctx, newAttrs, stateAttrs, "source")
	m.preserveNestedObjectIds(ctx, newAttrs, stateAttrs, "destination")
	m.preserveNestedObjectIds(ctx, newAttrs, stateAttrs, "application")
	m.preserveNestedObjectIds(ctx, newAttrs, stateAttrs, "service")
	
	// Preserve empty sets from config to prevent null/empty set correlation issues
	m.preserveEmptySet(ctx, newAttrs, stateAttrs, "country")
	m.preserveEmptySet(ctx, newAttrs, stateAttrs, "device")
	
	// Preserve device_attributes from state if both are null or both are objects
	// This prevents correlation issues when device_attributes transitions
	if stateDeviceAttrs, exists := stateAttrs["device_attributes"]; exists {
		if plannedDeviceAttrs, plannedExists := newAttrs["device_attributes"]; plannedExists {
			// If planned is an object with all null fields and state is null, use state's null
			plannedObj, plannedIsObj := plannedDeviceAttrs.(types.Object)
			stateObj, stateIsObj := stateDeviceAttrs.(types.Object)
			
			if plannedIsObj && stateIsObj {
				// Both are objects - check if planned is effectively empty (all nulls)
				plannedAttrsMap := plannedObj.Attributes()
				allNull := true
				for _, v := range plannedAttrsMap {
					if listVal, ok := v.(types.List); ok {
						if !listVal.IsNull() {
							allNull = false
							break
						}
					}
				}
				
				// If state is also all null or if state is null, preserve state value
				if stateObj.IsNull() && allNull {
					newAttrs["device_attributes"] = stateDeviceAttrs
				}
			} else if plannedIsObj && !plannedObj.IsNull() && stateObj.IsNull() {
				// Planned is object but state is null - preserve state's null
				newAttrs["device_attributes"] = stateDeviceAttrs
			}
		}
	}

	// Create the new object
	objectType := plannedObj.Type(ctx).(types.ObjectType)
	newObj, diags := types.ObjectValue(objectType.AttrTypes, newAttrs)
	if diags.HasError() {
		tflog.Error(ctx, "ExceptionsSetModifier: Failed to create preserved object, using planned")
		return plannedObj
	}

	return newObj
}

// preserveNestedObjectIds preserves ID values in nested objects (like source.host, source.site, etc.)
func (m wanExceptionsSetModifier) preserveNestedObjectIds(ctx context.Context, newAttrs map[string]attr.Value, stateAttrs map[string]attr.Value, nestedFieldName string) {
	plannedNested, exists := newAttrs[nestedFieldName]
	if !exists {
		return
	}

	stateNested, stateExists := stateAttrs[nestedFieldName]
	if !stateExists {
		return
	}

	// Handle nested object with sets (like source, destination, application)
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
	// For each attribute in the state object, ensure we also process it
	// This ensures we handle null values properly
	processedAttrs := make(map[string]bool)
	for attrName := range plannedNestedAttrs {
		m.preserveSetElementIds(ctx, newNestedAttrs, stateNestedAttrs, attrName)
		processedAttrs[attrName] = true
	}

	// Also check state attributes that might not be in planned
	for attrName := range stateNestedAttrs {
		if !processedAttrs[attrName] {
			m.preserveSetElementIds(ctx, newNestedAttrs, stateNestedAttrs, attrName)
		}
	}

	// Recreate the nested object
	nestedObjectType := plannedNestedObj.Type(ctx).(types.ObjectType)
	newNestedObj, diags := types.ObjectValue(nestedObjectType.AttrTypes, newNestedAttrs)
	if !diags.HasError() {
		newAttrs[nestedFieldName] = newNestedObj
	}
}

// preserveSetElementIds preserves ID values within set elements by matching on name
func (m wanExceptionsSetModifier) preserveSetElementIds(ctx context.Context, newNestedAttrs map[string]attr.Value, stateNestedAttrs map[string]attr.Value, setFieldName string) {
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
		}
	}
}

// findElementByName finds an element in the list by matching the name field, or by ID if name is unknown
func (m wanExceptionsSetModifier) findElementByName(ctx context.Context, targetObj types.Object, elements []attr.Value) *types.Object {
	targetAttrs := targetObj.Attributes()
	
	// First try to match by name
	targetName, nameExists := targetAttrs["name"]
	if nameExists {
		targetNameStr, ok := targetName.(types.String)
		if ok && !targetNameStr.IsNull() && !targetNameStr.IsUnknown() {
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
		}
	}
	
	// If name matching failed (or name is unknown), try to match by ID
	targetId, idExists := targetAttrs["id"]
	if idExists {
		targetIdStr, ok := targetId.(types.String)
		if ok && !targetIdStr.IsNull() && !targetIdStr.IsUnknown() {
			for _, element := range elements {
				elementObj, ok := element.(types.Object)
				if !ok {
					continue
				}

				elementAttrs := elementObj.Attributes()
				elementId, exists := elementAttrs["id"]
				if !exists {
					continue
				}

				elementIdStr, ok := elementId.(types.String)
				if !ok {
					continue
				}

				if !elementIdStr.IsNull() && !elementIdStr.IsUnknown() &&
					elementIdStr.ValueString() == targetIdStr.ValueString() {
					tflog.Debug(ctx, "ExceptionsSetModifier: Found element by ID", map[string]interface{}{"id": targetIdStr.ValueString()})
					return &elementObj
				}
			}
		}
	}
	
	return nil
}

// preserveElementId creates a new element object with ID and name preserved from state
func (m wanExceptionsSetModifier) preserveElementId(ctx context.Context, plannedObj types.Object, stateObj types.Object) types.Object {
	plannedAttrs := plannedObj.Attributes()
	stateAttrs := stateObj.Attributes()

	newAttrs := make(map[string]attr.Value, len(plannedAttrs))
	for k, v := range plannedAttrs {
		newAttrs[k] = v
	}

	// Preserve ID from state if it exists
	if stateId, exists := stateAttrs["id"]; exists {
		if stateIdStr, ok := stateId.(types.String); ok && !stateIdStr.IsNull() && !stateIdStr.IsUnknown() {
			newAttrs["id"] = stateIdStr
		}
	}
	
	// Preserve name from state if planned name is unknown
	if plannedName, exists := newAttrs["name"]; exists {
		if plannedNameStr, ok := plannedName.(types.String); ok && plannedNameStr.IsUnknown() {
			if stateName, stateExists := stateAttrs["name"]; stateExists {
				if stateNameStr, ok := stateName.(types.String); ok && !stateNameStr.IsNull() && !stateNameStr.IsUnknown() {
					newAttrs["name"] = stateNameStr
					tflog.Debug(ctx, "ExceptionsSetModifier: Preserved name from state", map[string]interface{}{"name": stateNameStr.ValueString()})
				}
			}
		}
	}

	objectType := plannedObj.Type(ctx).(types.ObjectType)
	newObj, diags := types.ObjectValue(objectType.AttrTypes, newAttrs)
	if diags.HasError() {
		return plannedObj
	}

	return newObj
}

// preserveEmptySet
// instead of being converted to null by the state hydration
func (m wanExceptionsSetModifier) preserveEmptySet(ctx context.Context, newAttrs map[string]attr.Value, stateAttrs map[string]attr.Value, fieldName string) {
	plannedField, plannedExists := newAttrs[fieldName]
	if !plannedExists {
		return
	}
	
	stateField, stateExists := stateAttrs[fieldName]
	if !stateExists {
		return
	}
	
	// Check if planned is an empty set and state is null
	plannedSet, plannedIsSet := plannedField.(types.Set)
	stateSet, stateIsSet := stateField.(types.Set)
	
	if plannedIsSet && stateIsSet {
		// If planned is empty set and state is null, preserve the empty set from plan
		if !plannedSet.IsNull() && !plannedSet.IsUnknown() && len(plannedSet.Elements()) == 0 &&
			stateSet.IsNull() {
			tflog.Debug(ctx, "ExceptionsSetModifier: Preserving empty set from plan",
				map[string]interface{}{"field": fieldName})
			// Keep the planned empty set
			newAttrs[fieldName] = plannedSet
		}
	}
}

// resolveUnknownToNull recursively resolves Unknown values to Null in an exception object
// This enables proper correlation between planned and actual values during first apply
func (m wanExceptionsSetModifier) resolveUnknownToNull(ctx context.Context, plannedObj types.Object) types.Object {
	plannedAttrs := plannedObj.Attributes()

	// Start with planned attributes
	newAttrs := make(map[string]attr.Value, len(plannedAttrs))
	for k, v := range plannedAttrs {
		newAttrs[k] = v
	}

	// Process nested objects (source, destination, application, service)
	m.resolveNestedObjectUnknowns(ctx, newAttrs, "source")
	m.resolveNestedObjectUnknowns(ctx, newAttrs, "destination")
	m.resolveNestedObjectUnknowns(ctx, newAttrs, "application")
	m.resolveNestedObjectUnknowns(ctx, newAttrs, "service")

	// Handle device_attributes - if it has all Unknown/Null fields, set to Null
	if deviceAttrs, exists := newAttrs["device_attributes"]; exists {
		if deviceAttrsObj, ok := deviceAttrs.(types.Object); ok && !deviceAttrsObj.IsNull() && !deviceAttrsObj.IsUnknown() {
			newAttrs["device_attributes"] = m.resolveDeviceAttributesUnknowns(ctx, deviceAttrsObj)
		}
	}

	// Create the new object
	objectType := plannedObj.Type(ctx).(types.ObjectType)
	newObj, diags := types.ObjectValue(objectType.AttrTypes, newAttrs)
	if diags.HasError() {
		tflog.Error(ctx, "ExceptionsSetModifier: Failed to create resolved object, using planned")
		return plannedObj
	}

	return newObj
}

// resolveNestedObjectUnknowns resolves Unknown values to Null in nested objects
func (m wanExceptionsSetModifier) resolveNestedObjectUnknowns(ctx context.Context, newAttrs map[string]attr.Value, fieldName string) {
	nestedField, exists := newAttrs[fieldName]
	if !exists {
		return
	}

	nestedObj, ok := nestedField.(types.Object)
	if !ok || nestedObj.IsNull() || nestedObj.IsUnknown() {
		return
	}

	nestedAttrs := nestedObj.Attributes()
	newNestedAttrs := make(map[string]attr.Value, len(nestedAttrs))
	for k, v := range nestedAttrs {
		newNestedAttrs[k] = v
	}

	// Process each attribute in the nested object
	for attrName, attrValue := range nestedAttrs {
		// Handle sets (like host, site, floating_subnet, etc.)
		if setVal, ok := attrValue.(types.Set); ok {
			if !setVal.IsNull() && !setVal.IsUnknown() {
				newNestedAttrs[attrName] = m.resolveSetUnknowns(ctx, setVal)
			}
		}
		// Handle lists (like ip, subnet, etc.)
		if listVal, ok := attrValue.(types.List); ok {
			if listVal.IsUnknown() {
				newNestedAttrs[attrName] = types.ListNull(listVal.ElementType(ctx))
			}
		}
	}

	// Recreate the nested object
	nestedObjectType := nestedObj.Type(ctx).(types.ObjectType)
	newNestedObj, diags := types.ObjectValue(nestedObjectType.AttrTypes, newNestedAttrs)
	if !diags.HasError() {
		newAttrs[fieldName] = newNestedObj
	}
}

// resolveSetUnknowns resolves Unknown values in set elements to Null
func (m wanExceptionsSetModifier) resolveSetUnknowns(ctx context.Context, setVal types.Set) types.Set {
	elements := setVal.Elements()
	if len(elements) == 0 {
		return setVal
	}

	modifiedElements := make([]attr.Value, 0, len(elements))
	for _, element := range elements {
		elementObj, ok := element.(types.Object)
		if !ok {
			modifiedElements = append(modifiedElements, element)
			continue
		}

		// Resolve Unknown name/id to Null
		elementAttrs := elementObj.Attributes()
		newElementAttrs := make(map[string]attr.Value, len(elementAttrs))
		for k, v := range elementAttrs {
			newElementAttrs[k] = v
		}

		// If name is Unknown, set it to Null
		if nameVal, exists := newElementAttrs["name"]; exists {
			if nameStr, ok := nameVal.(types.String); ok && nameStr.IsUnknown() {
				newElementAttrs["name"] = types.StringNull()
			}
		}

		// If id is Unknown, set it to Null
		if idVal, exists := newElementAttrs["id"]; exists {
			if idStr, ok := idVal.(types.String); ok && idStr.IsUnknown() {
				newElementAttrs["id"] = types.StringNull()
			}
		}

		elementObjectType := elementObj.Type(ctx).(types.ObjectType)
		newElementObj, diags := types.ObjectValue(elementObjectType.AttrTypes, newElementAttrs)
		if !diags.HasError() {
			modifiedElements = append(modifiedElements, newElementObj)
		} else {
			modifiedElements = append(modifiedElements, element)
		}
	}

	newSet, diags := types.SetValue(setVal.ElementType(ctx), modifiedElements)
	if !diags.HasError() {
		return newSet
	}
	return setVal
}

// resolveDeviceAttributesUnknowns resolves Unknown values in device_attributes
func (m wanExceptionsSetModifier) resolveDeviceAttributesUnknowns(ctx context.Context, deviceAttrsObj types.Object) types.Object {
	attrs := deviceAttrsObj.Attributes()
	newAttrs := make(map[string]attr.Value, len(attrs))
	
	for k, v := range attrs {
		if listVal, ok := v.(types.List); ok {
			if listVal.IsUnknown() {
				newAttrs[k] = types.ListNull(types.StringType)
			} else {
				newAttrs[k] = v
			}
		} else {
			newAttrs[k] = v
		}
	}

	objectType := deviceAttrsObj.Type(ctx).(types.ObjectType)
	newObj, diags := types.ObjectValue(objectType.AttrTypes, newAttrs)
	if !diags.HasError() {
		return newObj
	}
	return deviceAttrsObj
}

// ExceptionsSetModifier returns a new exceptions set plan modifier
func WanExceptionsSetModifier() planmodifier.Set {
	return wanExceptionsSetModifier{}
}
