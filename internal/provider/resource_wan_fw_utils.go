package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// correlateWanExceptions applies similar logic to ExceptionsSetModifier during state setting
// It correlates exception elements between the plan and API response by matching on names
// and preserves IDs from the plan for consistent state
func correlateWanExceptions(ctx context.Context, planExceptions types.Set, responseExceptions types.Set) *types.Set {
	if planExceptions.IsNull() || planExceptions.IsUnknown() || responseExceptions.IsNull() || responseExceptions.IsUnknown() {
		return nil
	}

	planElements := planExceptions.Elements()
	responseElements := responseExceptions.Elements()

	tflog.Debug(ctx, "correlateWanExceptions: Element counts", map[string]interface{}{
		"plan":     len(planElements),
		"response": len(responseElements),
	})

	// If the number of elements differs significantly, return the response as-is
	if len(planElements) != len(responseElements) {
		tflog.Debug(ctx, "correlateWanExceptions: Element count mismatch, using response")
		return &responseExceptions
	}

	// Create a new set of elements where we correlate plan and response
	correlatedElements := make([]attr.Value, 0, len(planElements))

	// Instead of iterating over response elements, iterate over plan elements to preserve order
	for _, planElement := range planElements {
		planObj, ok := planElement.(types.Object)
		if !ok {
			correlatedElements = append(correlatedElements, planElement)
			continue
		}

		// Find the corresponding response element by matching non-ID fields
		correspondingResponseElement := findCorrespondingWanResponseElement(ctx, planObj, responseElements)

		if correspondingResponseElement != nil {
			// Create correlated element that uses response data but preserves plan structure for unknowns
			correlatedElement := correlateWanExceptionElement(ctx, planObj, *correspondingResponseElement)
			correlatedElements = append(correlatedElements, correlatedElement)
		} else {
			// No corresponding response element found, use plan as-is
			correlatedElements = append(correlatedElements, planObj)
		}
	}

	// Create the correlated set
	if len(correlatedElements) > 0 {
		correlatedSet, diags := types.SetValue(planExceptions.ElementType(ctx), correlatedElements)
		if !diags.HasError() {
			tflog.Debug(ctx, "correlateWanExceptions: Successfully correlated exceptions set")
			return &correlatedSet
		}
	}

	return &responseExceptions
}

// findCorrespondingWanResponseElement finds the response element that corresponds to the given plan element
func findCorrespondingWanResponseElement(ctx context.Context, planObj types.Object, responseElements []attr.Value) *types.Object {
	planAttrs := planObj.Attributes()

	// Use exception name as the primary identifier
	planName, nameExists := planAttrs["name"]
	if !nameExists {
		return nil
	}

	planNameStr, ok := planName.(types.String)
	if !ok || planNameStr.IsNull() || planNameStr.IsUnknown() {
		return nil
	}

	for _, responseElement := range responseElements {
		responseObj, ok := responseElement.(types.Object)
		if !ok {
			continue
		}

		responseAttrs := responseObj.Attributes()
		responseName, exists := responseAttrs["name"]
		if !exists {
			continue
		}

		responseNameStr, ok := responseName.(types.String)
		if !ok {
			continue
		}

		// Match by name (primary identifier for exceptions)
		if !responseNameStr.IsNull() && !responseNameStr.IsUnknown() &&
			responseNameStr.ValueString() == planNameStr.ValueString() {
			return &responseObj
		}
	}

	return nil
}

// findCorrespondingWanPlanElement finds the plan element that corresponds to the given response element
func findCorrespondingWanPlanElement(ctx context.Context, responseObj types.Object, planElements []attr.Value) *types.Object {
	responseAttrs := responseObj.Attributes()

	// Use exception name as the primary identifier
	responseName, nameExists := responseAttrs["name"]
	if !nameExists {
		return nil
	}

	responseNameStr, ok := responseName.(types.String)
	if !ok || responseNameStr.IsNull() || responseNameStr.IsUnknown() {
		return nil
	}

	for _, planElement := range planElements {
		planObj, ok := planElement.(types.Object)
		if !ok {
			continue
		}

		planAttrs := planObj.Attributes()
		planName, exists := planAttrs["name"]
		if !exists {
			continue
		}

		planNameStr, ok := planName.(types.String)
		if !ok {
			continue
		}

		// Match by name (primary identifier for exceptions)
		if !planNameStr.IsNull() && !planNameStr.IsUnknown() &&
			planNameStr.ValueString() == responseNameStr.ValueString() {
			return &planObj
		}
	}

	return nil
}

// correlateWanExceptionElement creates a correlated exception element that combines plan structure with response data
func correlateWanExceptionElement(ctx context.Context, planObj types.Object, responseObj types.Object) types.Object {
	planAttrs := planObj.Attributes()
	responseAttrs := responseObj.Attributes()

	// Start with response attributes (they have the real data)
	newAttrs := make(map[string]attr.Value, len(responseAttrs))
	for k, v := range responseAttrs {
		newAttrs[k] = v
	}

	// Correlate nested object structures from plan
	correlateWanNestedObjects(ctx, newAttrs, planAttrs, responseAttrs, "source")
	correlateWanNestedObjects(ctx, newAttrs, planAttrs, responseAttrs, "destination")
	correlateWanNestedObjects(ctx, newAttrs, planAttrs, responseAttrs, "application")
	correlateWanNestedObjects(ctx, newAttrs, planAttrs, responseAttrs, "service")

	// Correlate top-level sets like country and device - preserve plan's null if response has empty set
	preserveNullOrCorrelateSet(ctx, newAttrs, planAttrs, "country")
	preserveNullOrCorrelateSet(ctx, newAttrs, planAttrs, "device")

	// Handle device_attributes - preserve plan's structure but resolve any Unknown values
	// This is critical: if plan has an object with all null fields, we must preserve that exact structure
	if planDeviceAttrs, exists := planAttrs["device_attributes"]; exists {
		// Resolve any Unknown values to Null to prevent "unknown value after apply" errors
		resolvedDeviceAttrs := resolveDeviceAttributesUnknowns(ctx, planDeviceAttrs)
		newAttrs["device_attributes"] = resolvedDeviceAttrs
		tflog.Debug(ctx, "correlateWanExceptionElement: Preserving plan's device_attributes structure (resolved unknowns)")
	}

	// Create the new object
	objectType := responseObj.Type(ctx).(types.ObjectType)
	newObj, diags := types.ObjectValue(objectType.AttrTypes, newAttrs)
	if diags.HasError() {
		tflog.Error(ctx, "correlateWanExceptionElement: Failed to create correlated object, using response")
		return responseObj
	}

	return newObj
}

// resolveDeviceAttributesUnknowns resolves Unknown values in device_attributes to Null
func resolveDeviceAttributesUnknowns(ctx context.Context, deviceAttrs attr.Value) attr.Value {
	// If it's null or unknown at the top level, return null
	if deviceAttrs == nil {
		return nil
	}

	deviceAttrsObj, ok := deviceAttrs.(types.Object)
	if !ok {
		return deviceAttrs
	}

	// If device_attributes is null, return null
	if deviceAttrsObj.IsNull() {
		return types.ObjectNull(deviceAttrsObj.Type(ctx).(types.ObjectType).AttrTypes)
	}

	// If device_attributes is unknown, convert to null
	if deviceAttrsObj.IsUnknown() {
		return types.ObjectNull(deviceAttrsObj.Type(ctx).(types.ObjectType).AttrTypes)
	}

	// Resolve Unknown values in nested lists
	attrs := deviceAttrsObj.Attributes()
	newAttrs := make(map[string]attr.Value, len(attrs))

	for k, v := range attrs {
		if listVal, ok := v.(types.List); ok {
			if listVal.IsUnknown() {
				// Convert Unknown list to Null list
				newAttrs[k] = types.ListNull(types.StringType)
				tflog.Debug(ctx, "resolveDeviceAttributesUnknowns: Resolved unknown list to null", map[string]interface{}{"field": k})
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
	return deviceAttrs
}

// preserveNullOrCorrelateSet preserves plan's null value if response has empty set, otherwise correlates
func preserveNullOrCorrelateSet(ctx context.Context, newAttrs map[string]attr.Value, planAttrs map[string]attr.Value, fieldName string) {
	planField, planExists := planAttrs[fieldName]
	if !planExists {
		return
	}

	planSetVal, ok := planField.(types.Set)
	if !ok {
		return
	}

	// If plan has null, preserve it regardless of response
	if planSetVal.IsNull() {
		newAttrs[fieldName] = planSetVal
		tflog.Debug(ctx, "preserveNullOrCorrelateSet: Preserving null from plan", map[string]interface{}{"field": fieldName})
		return
	}

	// If plan has empty set, preserve it
	if !planSetVal.IsUnknown() && len(planSetVal.Elements()) == 0 {
		newAttrs[fieldName] = planSetVal
		tflog.Debug(ctx, "preserveNullOrCorrelateSet: Preserving empty set from plan", map[string]interface{}{"field": fieldName})
		return
	}

	// Otherwise, correlate the set elements
	correlateWanSetElements(ctx, newAttrs, planAttrs, fieldName)
}

// correlateWanNestedObjects correlates nested objects like source, destination, application
func correlateWanNestedObjects(ctx context.Context, newAttrs map[string]attr.Value, planAttrs map[string]attr.Value, responseAttrs map[string]attr.Value, nestedFieldName string) {
	responsNested, responseExists := newAttrs[nestedFieldName]
	if !responseExists {
		return
	}

	planNested, planExists := planAttrs[nestedFieldName]
	if !planExists {
		return
	}

	// Handle nested object with sets (like source, destination, application)
	responseNestedObj, ok := responsNested.(types.Object)
	if !ok {
		return
	}

	planNestedObj, ok := planNested.(types.Object)
	if !ok {
		return
	}

	responseNestedAttrs := responseNestedObj.Attributes()
	planNestedAttrs := planNestedObj.Attributes()

	newNestedAttrs := make(map[string]attr.Value, len(responseNestedAttrs))
	for k, v := range responseNestedAttrs {
		newNestedAttrs[k] = v
	}

	// Correlate sets within nested objects (like host, site, etc.)
	for attrName := range responseNestedAttrs {
		correlateWanSetElements(ctx, newNestedAttrs, planNestedAttrs, attrName)
	}

	// Recreate the nested object
	nestedObjectType := responseNestedObj.Type(ctx).(types.ObjectType)
	newNestedObj, diags := types.ObjectValue(nestedObjectType.AttrTypes, newNestedAttrs)
	if !diags.HasError() {
		newAttrs[nestedFieldName] = newNestedObj
	}
}

// correlateWanSetElements correlates set elements by matching on name and preserving plan structure
func correlateWanSetElements(ctx context.Context, newNestedAttrs map[string]attr.Value, planNestedAttrs map[string]attr.Value, setFieldName string) {
	responseSet, responseExists := newNestedAttrs[setFieldName]
	if !responseExists {
		return
	}

	planSet, planExists := planNestedAttrs[setFieldName]
	if !planExists {
		return
	}

	responseSetVal, ok := responseSet.(types.Set)
	if !ok {
		return
	}

	planSetVal, ok := planSet.(types.Set)
	if !ok {
		return
	}

	if responseSetVal.IsNull() || responseSetVal.IsUnknown() || planSetVal.IsNull() || planSetVal.IsUnknown() {
		return
	}

	responseElements := responseSetVal.Elements()
	planElements := planSetVal.Elements()

	correlatedElements := make([]attr.Value, 0, len(responseElements))

	// For each response element, try to find corresponding plan element
	for _, responseElement := range responseElements {
		responseElementObj, ok := responseElement.(types.Object)
		if !ok {
			correlatedElements = append(correlatedElements, responseElement)
			continue
		}

		// Find corresponding plan element by name
		correspondingPlanElement := findWanElementByName(ctx, responseElementObj, planElements)
		if correspondingPlanElement != nil {
			// Create correlated element that preserves plan structure
			correlatedElement := correlateWanSetElement(ctx, *correspondingPlanElement, responseElementObj)
			correlatedElements = append(correlatedElements, correlatedElement)
		} else {
			correlatedElements = append(correlatedElements, responseElement)
		}
	}

	// Create new set with correlated elements
	if len(correlatedElements) > 0 {
		newSet, diags := types.SetValue(responseSetVal.ElementType(ctx), correlatedElements)
		if !diags.HasError() {
			newNestedAttrs[setFieldName] = newSet
		}
	}
}

// findWanElementByName finds an element in the list by matching the name or ID field
func findWanElementByName(ctx context.Context, targetObj types.Object, elements []attr.Value) *types.Object {
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

	// If name matching failed, try to match by ID
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
					tflog.Debug(ctx, "findWanElementByName: Found element by ID", map[string]interface{}{"id": targetIdStr.ValueString()})
					return &elementObj
				}
			}
		}
	}

	return nil
}

// correlateWanSetElement creates a correlated set element that uses plan structure with response data
// It preserves null values from the plan for computed fields (like name when user only specified id)
func correlateWanSetElement(ctx context.Context, planObj types.Object, responseObj types.Object) types.Object {
	planAttrs := planObj.Attributes()
	responseAttrs := responseObj.Attributes()

	newAttrs := make(map[string]attr.Value, len(responseAttrs))
	// Start with response attributes
	for k, v := range responseAttrs {
		newAttrs[k] = v
	}

	// Preserve null values from plan for computed fields
	// If plan had name=null and user specified id, keep name=null in state
	// to match the planned value and avoid correlation issues
	if planName, exists := planAttrs["name"]; exists {
		if planNameStr, ok := planName.(types.String); ok && planNameStr.IsNull() {
			// Plan had null name (user specified only id), preserve null
			newAttrs["name"] = types.StringNull()
			tflog.Debug(ctx, "correlateWanSetElement: Preserving null name from plan")
		}
	}

	// Similarly, if plan had id=null and user specified name, preserve null id
	if planId, exists := planAttrs["id"]; exists {
		if planIdStr, ok := planId.(types.String); ok && planIdStr.IsNull() {
			// Plan had null id (user specified only name), preserve null
			newAttrs["id"] = types.StringNull()
			tflog.Debug(ctx, "correlateWanSetElement: Preserving null id from plan")
		}
	}

	objectType := responseObj.Type(ctx).(types.ObjectType)
	newObj, diags := types.ObjectValue(objectType.AttrTypes, newAttrs)
	if diags.HasError() {
		return responseObj
	}

	return newObj
}
