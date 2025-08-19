package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DEPRECATED: This file previously contained semantic equality functions for WAN firewall rules.
// These functions have been replaced by the ExceptionsSetModifier plan modifier
// which provides better handling of exceptions state management.
//
// The ExceptionsSetModifier approach is preferred because:
// 1. It works with the supported Terraform plugin framework version (1.14.1)
// 2. It preserves nested object IDs correctly
// 3. It avoids "Provider produced inconsistent result after apply" errors
// 4. It follows the existing codebase patterns for set handling
//
// The semantic equality functions used unsupported types like tftypes.SemanticEqualityRequest
// that are not available in framework version 1.14.1. The plan modifier approach
// achieves the same goals while being compatible with the current framework version.

// exceptionsElementSemanticEquals compares two exception elements for semantic equality
func exceptionsElementSemanticEquals(ctx context.Context, oldObj, newObj types.Object) bool {
	if oldObj.IsNull() && newObj.IsNull() {
		return true
	}
	if oldObj.IsNull() || newObj.IsNull() {
		return false
	}
	if oldObj.IsUnknown() || newObj.IsUnknown() {
		return false
	}

	oldAttrs := oldObj.Attributes()
	newAttrs := newObj.Attributes()

	// Compare name (primary identifier)
	if !stringAttrSemanticEquals(oldAttrs["name"], newAttrs["name"]) {
		return false
	}

	// Compare direction
	if !stringAttrSemanticEquals(oldAttrs["direction"], newAttrs["direction"]) {
		return false
	}

	// Compare connection_origin
	if !stringAttrSemanticEquals(oldAttrs["connection_origin"], newAttrs["connection_origin"]) {
		return false
	}

	// Compare nested objects
	if !objectAttrSemanticEquals(ctx, oldAttrs["source"], newAttrs["source"]) {
		return false
	}

	if !objectAttrSemanticEquals(ctx, oldAttrs["destination"], newAttrs["destination"]) {
		return false
	}

	if !objectAttrSemanticEquals(ctx, oldAttrs["application"], newAttrs["application"]) {
		return false
	}

	if !objectAttrSemanticEquals(ctx, oldAttrs["service"], newAttrs["service"]) {
		return false
	}

	return true
}

// stringAttrSemanticEquals compares two string attributes for semantic equality
func stringAttrSemanticEquals(oldAttr, newAttr attr.Value) bool {
	if oldAttr == nil && newAttr == nil {
		return true
	}
	if oldAttr == nil || newAttr == nil {
		return false
	}

	oldStr, ok1 := oldAttr.(types.String)
	newStr, ok2 := newAttr.(types.String)
	if !ok1 || !ok2 {
		return false
	}

	// Handle null/unknown states
	if oldStr.IsNull() && newStr.IsNull() {
		return true
	}
	if oldStr.IsNull() || newStr.IsNull() {
		return false
	}
	if oldStr.IsUnknown() || newStr.IsUnknown() {
		return false
	}

	return oldStr.ValueString() == newStr.ValueString()
}

// objectAttrSemanticEquals compares two object attributes for semantic equality
func objectAttrSemanticEquals(ctx context.Context, oldAttr, newAttr attr.Value) bool {
	if oldAttr == nil && newAttr == nil {
		return true
	}
	if oldAttr == nil || newAttr == nil {
		return false
	}

	oldObj, ok1 := oldAttr.(types.Object)
	newObj, ok2 := newAttr.(types.Object)
	if !ok1 || !ok2 {
		return false
	}

	// Handle null/unknown states
	if oldObj.IsNull() && newObj.IsNull() {
		return true
	}
	if oldObj.IsNull() || newObj.IsNull() {
		return false
	}
	if oldObj.IsUnknown() || newObj.IsUnknown() {
		return false
	}

	oldAttrs := oldObj.Attributes()
	newAttrs := newObj.Attributes()

	// Compare all attributes in the object
	for attrName, oldValue := range oldAttrs {
		newValue, exists := newAttrs[attrName]
		if !exists {
			return false
		}

		if !attributeSemanticEquals(ctx, oldValue, newValue) {
			return false
		}
	}

	// Check for any new attributes that weren't in old
	for attrName := range newAttrs {
		if _, exists := oldAttrs[attrName]; !exists {
			return false
		}
	}

	return true
}

// attributeSemanticEquals compares any two attributes for semantic equality
func attributeSemanticEquals(ctx context.Context, oldAttr, newAttr attr.Value) bool {
	if oldAttr == nil && newAttr == nil {
		return true
	}
	if oldAttr == nil || newAttr == nil {
		return false
	}

	// Handle different types
	switch oldTyped := oldAttr.(type) {
	case types.String:
		newTyped, ok := newAttr.(types.String)
		if !ok {
			return false
		}
		return stringSemanticEquals(oldTyped, newTyped)

	case types.List:
		newTyped, ok := newAttr.(types.List)
		if !ok {
			return false
		}
		return listSemanticEquals(ctx, oldTyped, newTyped)

	case types.Set:
		newTyped, ok := newAttr.(types.Set)
		if !ok {
			return false
		}
		return setSemanticEquals(ctx, oldTyped, newTyped)

	case types.Object:
		newTyped, ok := newAttr.(types.Object)
		if !ok {
			return false
		}
		return objectSemanticEquals(ctx, oldTyped, newTyped)

	default:
		// For other types, use direct equality
		return oldAttr.Equal(newAttr)
	}
}

// stringSemanticEquals compares two string values
func stringSemanticEquals(oldStr, newStr types.String) bool {
	if oldStr.IsNull() && newStr.IsNull() {
		return true
	}
	if oldStr.IsNull() || newStr.IsNull() {
		return false
	}
	if oldStr.IsUnknown() || newStr.IsUnknown() {
		return false
	}
	return oldStr.ValueString() == newStr.ValueString()
}

// listSemanticEquals compares two list values
func listSemanticEquals(ctx context.Context, oldList, newList types.List) bool {
	if oldList.IsNull() && newList.IsNull() {
		return true
	}
	if oldList.IsNull() || newList.IsNull() {
		return false
	}
	if oldList.IsUnknown() || newList.IsUnknown() {
		return false
	}

	oldElements := oldList.Elements()
	newElements := newList.Elements()

	if len(oldElements) != len(newElements) {
		return false
	}

	for i, oldElement := range oldElements {
		if !attributeSemanticEquals(ctx, oldElement, newElements[i]) {
			return false
		}
	}

	return true
}

// setSemanticEquals compares two set values
func setSemanticEquals(ctx context.Context, oldSet, newSet types.Set) bool {
	if oldSet.IsNull() && newSet.IsNull() {
		return true
	}
	if oldSet.IsNull() || newSet.IsNull() {
		return false
	}
	if oldSet.IsUnknown() || newSet.IsUnknown() {
		return false
	}

	oldElements := oldSet.Elements()
	newElements := newSet.Elements()

	if len(oldElements) != len(newElements) {
		return false
	}

	// For sets with objects, we need to match by semantic equality rather than order
	if len(oldElements) > 0 {
		// Check if elements are objects (which need special handling)
		if _, isObject := oldElements[0].(types.Object); isObject {
			return setOfObjectsSemanticEquals(ctx, oldElements, newElements)
		}
	}

	// For non-object sets, use simple matching
	for _, oldElement := range oldElements {
		found := false
		for _, newElement := range newElements {
			if attributeSemanticEquals(ctx, oldElement, newElement) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// setOfObjectsSemanticEquals compares sets of objects by finding semantic matches
func setOfObjectsSemanticEquals(ctx context.Context, oldElements, newElements []attr.Value) bool {
	for _, oldElement := range oldElements {
		oldObj, ok := oldElement.(types.Object)
		if !ok {
			return false
		}

		found := false
		for _, newElement := range newElements {
			newObj, ok := newElement.(types.Object)
			if !ok {
				continue
			}

			if objectSemanticEquals(ctx, oldObj, newObj) {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}

	return true
}

// objectSemanticEquals compares two object values
func objectSemanticEquals(ctx context.Context, oldObj, newObj types.Object) bool {
	if oldObj.IsNull() && newObj.IsNull() {
		return true
	}
	if oldObj.IsNull() || newObj.IsNull() {
		return false
	}
	if oldObj.IsUnknown() || newObj.IsUnknown() {
		return false
	}

	oldAttrs := oldObj.Attributes()
	newAttrs := newObj.Attributes()

	// For objects with name fields, use name as primary identifier
	if oldName, hasOldName := oldAttrs["name"]; hasOldName {
		if newName, hasNewName := newAttrs["name"]; hasNewName {
			if !stringAttrSemanticEquals(oldName, newName) {
				return false
			}
		}
	}

	// Compare all other non-ID attributes (IDs can be computed and may differ)
	for attrName, oldValue := range oldAttrs {
		// Skip ID fields as they may be computed and differ between plan/state
		if attrName == "id" {
			continue
		}

		newValue, exists := newAttrs[attrName]
		if !exists {
			return false
		}

		if !attributeSemanticEquals(ctx, oldValue, newValue) {
			return false
		}
	}

	// Check for any new non-ID attributes that weren't in old
	for attrName := range newAttrs {
		if attrName == "id" {
			continue
		}
		if _, exists := oldAttrs[attrName]; !exists {
			return false
		}
	}

	return true
}
