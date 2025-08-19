package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// sourceDestObjectModifier handles the specific case where source/destination object elements
// have nested ID fields that transition from unknown to known values during apply.
// This prevents "Provider produced inconsistent result after apply" errors.
type sourceDestObjectModifier struct{}

// Description returns a human-readable description of the plan modifier
func (m sourceDestObjectModifier) Description(_ context.Context) string {
	return "Handles source/destination object element correlation when nested ID fields change from unknown to known"
}

// MarkdownDescription returns a markdown description of the plan modifier
func (m sourceDestObjectModifier) MarkdownDescription(_ context.Context) string {
	return "Handles source/destination object element correlation when nested ID fields change from unknown to known"
}

// PlanModifyObject implements the plan modification logic for Object attributes
func (m sourceDestObjectModifier) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	// Log entry into the plan modifier
	tflog.Warn(ctx, "SourceDestObjectModifier: Plan modifier invoked")
	
	// If config is null or unknown, use default behavior
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		tflog.Debug(ctx, "SourceDestObjectModifier: Config is null or unknown, using default behavior")
		return
	}
	
	// If state doesn't exist (first apply), use plan as-is but log it
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		tflog.Debug(ctx, "SourceDestObjectModifier: State is null or unknown (first apply), using plan as-is")
		return
	}
	
	// If plan is null or unknown, use default behavior
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		tflog.Debug(ctx, "SourceDestObjectModifier: Plan is null or unknown, using default behavior")
		return
	}
	
	tflog.Warn(ctx, "SourceDestObjectModifier: Processing source/destination object correlation")

	// Get the planned, state, and config objects
	plannedObj := req.PlanValue
	stateObj := req.StateValue

	// Create new object with preserved IDs for nested sets
	modifiedObj := m.preserveNestedSetIds(ctx, plannedObj, stateObj)
	
	if !modifiedObj.Equal(plannedObj) {
		resp.PlanValue = modifiedObj
		tflog.Warn(ctx, "SourceDestObjectModifier: Successfully modified source/destination object")
	} else {
		tflog.Debug(ctx, "SourceDestObjectModifier: No changes needed")
	}
}

// preserveNestedSetIds creates a new object that preserves state ID values for nested set elements
func (m sourceDestObjectModifier) preserveNestedSetIds(ctx context.Context, plannedObj types.Object, stateObj types.Object) types.Object {
	plannedAttrs := plannedObj.Attributes()
	stateAttrs := stateObj.Attributes()
	
	// Start with planned attributes
	newAttrs := make(map[string]attr.Value, len(plannedAttrs))
	for k, v := range plannedAttrs {
		newAttrs[k] = v
	}

	// Preserve IDs in nested sets that commonly have correlation issues
	setFieldsToProcess := []string{
		"host", "site", "global_ip_range", "network_interface", 
		"site_network_subnet", "floating_subnet", "group", 
		"system_group", "user", "users_group",
	}

	for _, fieldName := range setFieldsToProcess {
		m.preserveSetElementIds(ctx, newAttrs, stateAttrs, fieldName)
	}
	
	// Create the new object
	objectType := plannedObj.Type(ctx).(types.ObjectType)
	newObj, diags := types.ObjectValue(objectType.AttrTypes, newAttrs)
	if diags.HasError() {
		tflog.Error(ctx, "SourceDestObjectModifier: Failed to create preserved object, using planned")
		return plannedObj
	}

	return newObj
}

// preserveSetElementIds preserves ID values within set elements by matching on name
func (m sourceDestObjectModifier) preserveSetElementIds(ctx context.Context, newAttrs map[string]attr.Value, stateAttrs map[string]attr.Value, setFieldName string) {
	plannedSet, exists := newAttrs[setFieldName]
	if !exists {
		return
	}

	stateSet, stateExists := stateAttrs[setFieldName]
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

	// Handle null/unknown values
	if plannedSetVal.IsNull() || plannedSetVal.IsUnknown() || stateSetVal.IsNull() || stateSetVal.IsUnknown() {
		return
	}

	plannedElements := plannedSetVal.Elements()
	stateElements := stateSetVal.Elements()

	if len(plannedElements) == 0 || len(stateElements) == 0 {
		return
	}

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
			newAttrs[setFieldName] = newSet
			tflog.Debug(ctx, "SourceDestObjectModifier: Preserved IDs for set field", map[string]interface{}{"field": setFieldName})
		}
	}
}

// findElementByName finds an element in the list by matching the name field
func (m sourceDestObjectModifier) findElementByName(ctx context.Context, targetObj types.Object, elements []attr.Value) *types.Object {
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

// preserveElementId creates a new element object with ID preserved from state
func (m sourceDestObjectModifier) preserveElementId(ctx context.Context, plannedObj types.Object, stateObj types.Object) types.Object {
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

	objectType := plannedObj.Type(ctx).(types.ObjectType)
	newObj, diags := types.ObjectValue(objectType.AttrTypes, newAttrs)
	if diags.HasError() {
		return plannedObj
	}

	return newObj
}

// SourceDestObjectModifier returns a new source/destination object plan modifier
func SourceDestObjectModifier() planmodifier.Object {
	return sourceDestObjectModifier{}
}
