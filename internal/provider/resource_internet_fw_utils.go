package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Internet Firewall utility functions for correlation and state management
// correlateIfwExceptions correlates plan exceptions with response exceptions to preserve ID structure
// It supports both ID and name matching for elements, prioritizing ID matching if available
// For creates, it uses names; for updates if ID is present in state it will use that for correlation
func correlateIfwExceptions(ctx context.Context, planExceptions types.Set, responseExceptions types.Set) *types.Set {
	// Handle null/unknown sets gracefully
	if planExceptions.IsNull() || planExceptions.IsUnknown() || responseExceptions.IsNull() || responseExceptions.IsUnknown() {
		return nil
	}

	// Extract plan and response elements
	planElements := planExceptions.Elements()
	responseElements := responseExceptions.Elements()

	tflog.Debug(ctx, "correlateIfwExceptions: Element counts", map[string]interface{}{
		"plan":     len(planElements),
		"response": len(responseElements),
	})

	// If the counts differ significantly, return the response as-is
	if len(planElements) != len(responseElements) {
		tflog.Debug(ctx, "correlateIfwExceptions: Element count mismatch, using response")
		return &responseExceptions
	}

	// Create a new set of elements where we correlate plan and response
	correlatedElements := make([]attr.Value, 0, len(planElements))

	// Iterate over plan elements to preserve order
	for i, planElement := range planElements {
		planObj, ok := planElement.(types.Object)
		if !ok {
			tflog.Debug(ctx, "correlateIfwExceptions: Plan element is not an object", map[string]interface{}{"index": i})
			correlatedElements = append(correlatedElements, planElement)
			continue
		}

		// Find the corresponding response element by matching ID or name
		correspondingResponseElement := findCorrespondingIfwResponseElement(ctx, planObj, responseElements)

		if correspondingResponseElement != nil {
			tflog.Debug(ctx, "correlateIfwExceptions: Found corresponding response element", map[string]interface{}{
				"index": i,
				"plan_name": getPlanName(planObj),
				"response_name": getPlanName(*correspondingResponseElement),
			})
			// Create correlated element that preserves plan structure for unknowns
			correlatedElement := correlateIfwExceptionElement(ctx, planObj, *correspondingResponseElement)
			correlatedElements = append(correlatedElements, correlatedElement)
		} else {
			tflog.Warn(ctx, "correlateIfwExceptions: No corresponding response element found", map[string]interface{}{
				"index": i,
				"plan_name": getPlanName(planObj),
			})
			// No corresponding response element found, try to use the response element at the same index
			// This handles cases where elements match positionally but not by name/ID
			if i < len(responseElements) {
				if responseObj, ok := responseElements[i].(types.Object); ok {
					tflog.Debug(ctx, "correlateIfwExceptions: Using positional match", map[string]interface{}{
						"index": i,
						"plan_name": getPlanName(planObj),
						"response_name": getPlanName(responseObj),
					})
					correlatedElement := correlateIfwExceptionElement(ctx, planObj, responseObj)
					correlatedElements = append(correlatedElements, correlatedElement)
				} else {
					correlatedElements = append(correlatedElements, planObj)
				}
			} else {
				correlatedElements = append(correlatedElements, planObj)
			}
		}
	}

	// Create the correlated set
	if len(correlatedElements) > 0 {
		correlatedSet, diags := types.SetValue(planExceptions.ElementType(ctx), correlatedElements)
		if !diags.HasError() {
			tflog.Debug(ctx, "correlateIfwExceptions: Successfully correlated exceptions set")
			return &correlatedSet
		} else {
			tflog.Error(ctx, "correlateIfwExceptions: Failed to create correlated set, falling back to response", map[string]interface{}{
				"error": diags.Errors(),
			})
			return &responseExceptions
		}
	}

	return &responseExceptions
}

// Helper function to get the name from an object for debugging
func getPlanName(obj types.Object) string {
	if obj.IsNull() || obj.IsUnknown() {
		return "<null/unknown>"
	}
	attrs := obj.Attributes()
	if nameAttr, exists := attrs["name"]; exists {
		if nameStr, ok := nameAttr.(types.String); ok && !nameStr.IsNull() && !nameStr.IsUnknown() {
			return nameStr.ValueString()
		}
	}
	return "<no name>"
}

// findCorrespondingIfwResponseElement finds the response element that corresponds to the given plan element
// It first tries to match by ID if both elements have IDs, then falls back to name matching,
// and finally uses comprehensive field matching for cases where connection_origin or other fields differ
func findCorrespondingIfwResponseElement(ctx context.Context, planObj types.Object, responseElements []attr.Value) *types.Object {
	planAttrs := planObj.Attributes()

	// First, try to match by ID if both plan and response elements have IDs
	planId, planIdExists := planAttrs["id"]
	if planIdExists {
		planIdStr, ok := planId.(types.String)
		if ok && !planIdStr.IsNull() && !planIdStr.IsUnknown() {
			// Plan has a valid ID, try to find response element with the same ID
			for _, responseElement := range responseElements {
				responseObj, ok := responseElement.(types.Object)
				if !ok {
					continue
				}

				responseAttrs := responseObj.Attributes()
				responseId, responseIdExists := responseAttrs["id"]
				if !responseIdExists {
					continue
				}

				responseIdStr, ok := responseId.(types.String)
				if !ok || responseIdStr.IsNull() || responseIdStr.IsUnknown() {
					continue
				}

				// Match by ID (preferred for updates)
				if responseIdStr.ValueString() == planIdStr.ValueString() {
					tflog.Debug(ctx, "findCorrespondingIfwResponseElement: Found match by ID")
					return &responseObj
				}
			}
		}
	}

	// Fallback to name-based matching (for creates or when ID is not available)
	planName, nameExists := planAttrs["name"]
	if nameExists {
		planNameStr, ok := planName.(types.String)
		if ok && !planNameStr.IsNull() && !planNameStr.IsUnknown() {
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

				// Match by name (fallback identifier for exceptions)
				if !responseNameStr.IsNull() && !responseNameStr.IsUnknown() &&
					responseNameStr.ValueString() == planNameStr.ValueString() {
					tflog.Debug(ctx, "findCorrespondingIfwResponseElement: Found match by name")
					return &responseObj
				}
			}
		}
	}

	// Final fallback: comprehensive field matching for cases where elements might match on multiple criteria
	// This handles cases where connection_origin or other fields might differ between plan and response
	for _, responseElement := range responseElements {
		responseObj, ok := responseElement.(types.Object)
		if !ok {
			continue
		}

		// Try to match by comparing nested objects (source, destination, service)
		if isComprehensiveMatch(ctx, planObj, responseObj) {
			tflog.Debug(ctx, "findCorrespondingIfwResponseElement: Found comprehensive match")
			return &responseObj
		}
	}

	tflog.Debug(ctx, "findCorrespondingIfwResponseElement: No matching element found")
	return nil
}

// isComprehensiveMatch performs a comprehensive comparison of two exception objects
// It compares nested objects (source, destination, service) to determine if they represent the same exception
func isComprehensiveMatch(ctx context.Context, planObj types.Object, responseObj types.Object) bool {
	planAttrs := planObj.Attributes()
	responseAttrs := responseObj.Attributes()

	// First, check if both have the same name (if name exists)
	planName, planHasName := planAttrs["name"]
	responseName, responseHasName := responseAttrs["name"]
	
	if planHasName && responseHasName {
		planNameStr, planOk := planName.(types.String)
		responseNameStr, responseOk := responseName.(types.String)
		
		if planOk && responseOk && !planNameStr.IsNull() && !planNameStr.IsUnknown() &&
			!responseNameStr.IsNull() && !responseNameStr.IsUnknown() {
			// If names exist and are different, it's not a match
			if planNameStr.ValueString() != responseNameStr.ValueString() {
				return false
			}
		}
	}

	// Compare nested objects to determine similarity
	nested_fields := []string{"source", "destination", "service"}
	matchingFields := 0
	totalFields := 0

	for _, fieldName := range nested_fields {
		planField, planExists := planAttrs[fieldName]
		responseField, responseExists := responseAttrs[fieldName]
		
		if planExists && responseExists {
			totalFields++
			if compareNestedObjects(ctx, planField, responseField) {
				matchingFields++
			}
		}
	}

	// Consider it a match if most nested objects are similar (at least 2 out of 3)
	if totalFields > 0 {
		matchRatio := float64(matchingFields) / float64(totalFields)
		tflog.Debug(ctx, "isComprehensiveMatch: Calculated match ratio", map[string]interface{}{
			"matchingFields": matchingFields,
			"totalFields": totalFields,
			"matchRatio": matchRatio,
		})
		return matchRatio >= 0.6 // 60% match threshold
	}

	return false
}

// compareNestedObjects compares two nested objects and returns true if they are structurally similar
func compareNestedObjects(ctx context.Context, planField attr.Value, responseField attr.Value) bool {
	planObj, planOk := planField.(types.Object)
	responseObj, responseOk := responseField.(types.Object)

	if !planOk || !responseOk {
		return false
	}

	if planObj.IsNull() && responseObj.IsNull() {
		return true
	}

	if planObj.IsUnknown() || responseObj.IsUnknown() {
		return true // Consider unknowns as potential matches
	}

	if (planObj.IsNull() && !responseObj.IsNull()) || (!planObj.IsNull() && responseObj.IsNull()) {
		return false
	}

	planAttrs := planObj.Attributes()
	responseAttrs := responseObj.Attributes()

	// Count non-null attributes in both objects
	planNonNullCount := 0
	responseNonNullCount := 0

	for _, attr := range planAttrs {
		if !attr.IsNull() {
			planNonNullCount++
		}
	}

	for _, attr := range responseAttrs {
		if !attr.IsNull() {
			responseNonNullCount++
		}
	}

	// If both objects have similar complexity (number of non-null attributes), consider them similar
	// Allow for empty objects to match with objects that have all null values
	if planNonNullCount == 0 && responseNonNullCount == 0 {
		return true
	}

	tflog.Debug(ctx, "compareNestedObjects: Comparing object complexity", map[string]interface{}{
		"planNonNullCount": planNonNullCount,
		"responseNonNullCount": responseNonNullCount,
	})

	// Objects are similar if they have the same level of complexity
	return planNonNullCount == responseNonNullCount
}

// findCorrespondingIfwPlanElement finds the plan element that corresponds to the given response element
func findCorrespondingIfwPlanElement(ctx context.Context, responseObj types.Object, planElements []attr.Value) *types.Object {
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

// correlateIfwExceptionElement creates a correlated exception element that combines plan structure with response data
func correlateIfwExceptionElement(ctx context.Context, planObj types.Object, responseObj types.Object) types.Object {
	planAttrs := planObj.Attributes()
	responseAttrs := responseObj.Attributes()

	// Start with response attributes (they have the real data)
	newAttrs := make(map[string]attr.Value, len(responseAttrs))
	for k, v := range responseAttrs {
		newAttrs[k] = v
	}

	// CRITICAL: Preserve exception name from plan when it's null
	// This handles the case where exceptions are specified by nested element IDs only, not by name
	if planName, exists := planAttrs["name"]; exists {
		if planNameStr, ok := planName.(types.String); ok && planNameStr.IsNull() {
			newAttrs["name"] = types.StringNull()
			tflog.Debug(ctx, "correlateIfwExceptionElement: Preserving null name from plan")
		}
	}

	// Preserve certain plan values to ensure consistency with planned state
	// This handles cases where the plan has null values but the response has defaults
	preservePlanValue(ctx, newAttrs, planAttrs, "connection_origin")

	// Correlate nested object structures from plan
	correlateIfwNestedObjects(ctx, newAttrs, planAttrs, responseAttrs, "source")
	correlateIfwNestedObjects(ctx, newAttrs, planAttrs, responseAttrs, "destination")
	correlateIfwNestedObjects(ctx, newAttrs, planAttrs, responseAttrs, "service")

	// Preserve plan's null/empty structure for sets (country, device)
	preserveNullOrCorrelateIfwSet(ctx, newAttrs, planAttrs, "country")
	preserveNullOrCorrelateIfwSet(ctx, newAttrs, planAttrs, "device")

	// Handle device_attributes - preserve plan's structure but resolve any Unknown values
	// This is critical: if plan has an object with all null fields, we must preserve that exact structure
	if planDeviceAttrs, exists := planAttrs["device_attributes"]; exists {
		// Resolve any Unknown values to Null to prevent "unknown value after apply" errors
		resolvedDeviceAttrs := resolveIfwDeviceAttributesUnknowns(ctx, planDeviceAttrs)
		newAttrs["device_attributes"] = resolvedDeviceAttrs
		tflog.Debug(ctx, "correlateIfwExceptionElement: Preserving plan's device_attributes structure (resolved unknowns)")
	}

	// Create the new object
	objectType := responseObj.Type(ctx).(types.ObjectType)
	newObj, diags := types.ObjectValue(objectType.AttrTypes, newAttrs)
	if diags.HasError() {
		tflog.Error(ctx, "correlateIfwExceptionElement: Failed to create correlated object, using response")
		return responseObj
	}

	return newObj
}

// resolveIfwDeviceAttributesUnknowns resolves Unknown values in device_attributes to Null
func resolveIfwDeviceAttributesUnknowns(ctx context.Context, deviceAttrs attr.Value) attr.Value {
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
				tflog.Debug(ctx, "resolveIfwDeviceAttributesUnknowns: Resolved unknown list to null", map[string]interface{}{"field": k})
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

// preserveNullOrCorrelateIfwSet preserves plan's null value if response has empty set, otherwise correlates
func preserveNullOrCorrelateIfwSet(ctx context.Context, newAttrs map[string]attr.Value, planAttrs map[string]attr.Value, fieldName string) {
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
		tflog.Debug(ctx, "preserveNullOrCorrelateIfwSet: Preserving null from plan", map[string]interface{}{"field": fieldName})
		return
	}

	// If plan has empty set, preserve it
	if !planSetVal.IsUnknown() && len(planSetVal.Elements()) == 0 {
		newAttrs[fieldName] = planSetVal
		tflog.Debug(ctx, "preserveNullOrCorrelateIfwSet: Preserving empty set from plan", map[string]interface{}{"field": fieldName})
		return
	}

	// Otherwise, correlate the set elements
	correlateIfwSetElements(ctx, newAttrs, planAttrs, fieldName)
}

// preservePlanValue preserves a plan value in the new attributes if it's different from response
// This handles cases where the plan explicitly sets a value to null but the API returns a default
func preservePlanValue(ctx context.Context, newAttrs map[string]attr.Value, planAttrs map[string]attr.Value, fieldName string) {
	planValue, planExists := planAttrs[fieldName]
	responseValue, responseExists := newAttrs[fieldName]

	if !planExists || !responseExists {
		return
	}

	// If plan has null and response has a value, preserve the plan null
	// This ensures that when a user doesn't specify a field (null in plan),
	// it remains null in the correlated result even if API returns a default
	if planValue.IsNull() && !responseValue.IsNull() {
		tflog.Debug(ctx, "preservePlanValue: Preserving null plan value", map[string]interface{}{
			"field": fieldName,
		})
		newAttrs[fieldName] = planValue
	}
}

// correlateIfwNestedObjects correlates nested objects like source, destination, service
func correlateIfwNestedObjects(ctx context.Context, newAttrs map[string]attr.Value, planAttrs map[string]attr.Value, responseAttrs map[string]attr.Value, nestedFieldName string) {
	responsNested, responseExists := newAttrs[nestedFieldName]
	if !responseExists {
		tflog.Debug(ctx, "correlateIfwNestedObjects: Response field does not exist", map[string]interface{}{"field": nestedFieldName})
		return
	}

	planNested, planExists := planAttrs[nestedFieldName]
	if !planExists {
		tflog.Debug(ctx, "correlateIfwNestedObjects: Plan field does not exist", map[string]interface{}{"field": nestedFieldName})
		return
	}

	// Handle nested object with sets (like source, destination, service)
	responseNestedObj, ok := responsNested.(types.Object)
	if !ok {
		tflog.Debug(ctx, "correlateIfwNestedObjects: Response field is not an object", map[string]interface{}{"field": nestedFieldName, "type": fmt.Sprintf("%T", responsNested)})
		return
	}

	planNestedObj, ok := planNested.(types.Object)
	if !ok {
		tflog.Debug(ctx, "correlateIfwNestedObjects: Plan field is not an object", map[string]interface{}{"field": nestedFieldName, "type": fmt.Sprintf("%T", planNested)})
		return
	}

	// Handle null or unknown nested objects gracefully
	if responseNestedObj.IsNull() || responseNestedObj.IsUnknown() {
		tflog.Debug(ctx, "correlateIfwNestedObjects: Response nested object is null/unknown", map[string]interface{}{"field": nestedFieldName})
		return
	}

	if planNestedObj.IsNull() || planNestedObj.IsUnknown() {
		tflog.Debug(ctx, "correlateIfwNestedObjects: Plan nested object is null/unknown", map[string]interface{}{"field": nestedFieldName})
		return
	}

	responseNestedAttrs := responseNestedObj.Attributes()
	planNestedAttrs := planNestedObj.Attributes()

	tflog.Debug(ctx, "correlateIfwNestedObjects: Correlating nested object", map[string]interface{}{
		"field": nestedFieldName,
		"response_attr_count": len(responseNestedAttrs),
		"plan_attr_count": len(planNestedAttrs),
	})

	newNestedAttrs := make(map[string]attr.Value, len(responseNestedAttrs))
	for k, v := range responseNestedAttrs {
		newNestedAttrs[k] = v
	}

	// Correlate sets within nested objects (like host, site, application, etc.)
	for attrName := range responseNestedAttrs {
		correlateIfwSetElements(ctx, newNestedAttrs, planNestedAttrs, attrName)
	}

	// Recreate the nested object
	nestedObjectType := responseNestedObj.Type(ctx).(types.ObjectType)
	newNestedObj, diags := types.ObjectValue(nestedObjectType.AttrTypes, newNestedAttrs)
	if !diags.HasError() {
		newAttrs[nestedFieldName] = newNestedObj
		tflog.Debug(ctx, "correlateIfwNestedObjects: Successfully correlated nested object", map[string]interface{}{"field": nestedFieldName})
	} else {
		tflog.Error(ctx, "correlateIfwNestedObjects: Failed to create nested object", map[string]interface{}{
			"field": nestedFieldName,
			"error": diags.Errors(),
		})
	}
}

// correlateIfwSetElements correlates set elements by matching on name and preserving plan structure
func correlateIfwSetElements(ctx context.Context, newNestedAttrs map[string]attr.Value, planNestedAttrs map[string]attr.Value, setFieldName string) {
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
		correspondingPlanElement := findIfwElementByName(ctx, responseElementObj, planElements)
		if correspondingPlanElement != nil {
			// Create correlated element that preserves plan structure
			correlatedElement := correlateIfwSetElement(ctx, *correspondingPlanElement, responseElementObj)
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

// findIfwElementByName finds an element in the list by matching the ID field first, then name field
// This supports both create operations (name-based) and update operations (ID-based if present)
func findIfwElementByName(ctx context.Context, targetObj types.Object, elements []attr.Value) *types.Object {
	targetAttrs := targetObj.Attributes()

	// First, try to match by ID if both target and elements have IDs
	targetId, targetIdExists := targetAttrs["id"]
	if targetIdExists {
		targetIdStr, ok := targetId.(types.String)
		if ok && !targetIdStr.IsNull() && !targetIdStr.IsUnknown() {
			// Target has a valid ID, try to find element with the same ID
			for _, element := range elements {
				elementObj, ok := element.(types.Object)
				if !ok {
					continue
				}

				elementAttrs := elementObj.Attributes()
				elementId, elementIdExists := elementAttrs["id"]
				if !elementIdExists {
					continue
				}

				elementIdStr, ok := elementId.(types.String)
				if !ok || elementIdStr.IsNull() || elementIdStr.IsUnknown() {
					continue
				}

				// Match by ID (preferred)
				if elementIdStr.ValueString() == targetIdStr.ValueString() {
					return &elementObj
				}
			}
		}
	}

	// Fallback to name-based matching
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

// correlateIfwSetElement creates a correlated set element that uses plan structure with response data
// It preserves IDs from the plan when they exist to maintain consistency during updates
func correlateIfwSetElement(ctx context.Context, planObj types.Object, responseObj types.Object) types.Object {
	planAttrs := planObj.Attributes()
	responseAttrs := responseObj.Attributes()

	newAttrs := make(map[string]attr.Value, len(responseAttrs))
	// Start with response attributes (they have the real data)
	for k, v := range responseAttrs {
		newAttrs[k] = v
	}

	// Preserve null values from plan for computed fields
	// If plan had name=null and user specified id, keep name=null in state
	if planName, exists := planAttrs["name"]; exists {
		if planNameStr, ok := planName.(types.String); ok && planNameStr.IsNull() {
			// Plan had null name (user specified only id), preserve null
			newAttrs["name"] = types.StringNull()
			tflog.Debug(ctx, "correlateIfwSetElement: Preserving null name from plan")
		} else if !planNameStr.IsNull() && !planNameStr.IsUnknown() {
			// Plan has a valid name, preserve it
			newAttrs["name"] = planNameStr
		}
	}

	// Similarly, if plan had id=null and user specified name, preserve null id
	if planId, exists := planAttrs["id"]; exists {
		if planIdStr, ok := planId.(types.String); ok && planIdStr.IsNull() {
			// Plan had null id (user specified only name), preserve null
			newAttrs["id"] = types.StringNull()
			tflog.Debug(ctx, "correlateIfwSetElement: Preserving null id from plan")
		} else if !planIdStr.IsNull() && !planIdStr.IsUnknown() {
			// Plan has a valid id, preserve it
			newAttrs["id"] = planIdStr
		}
	}

	objectType := responseObj.Type(ctx).(types.ObjectType)
	newObj, diags := types.ObjectValue(objectType.AttrTypes, newAttrs)
	if diags.HasError() {
		tflog.Error(ctx, "correlateIfwSetElement: Failed to create correlated object, using response")
		return responseObj
	}

	return newObj
}
