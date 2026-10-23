package oldidname

import (
	"context"
	"fmt"
	"reflect"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

func ParseNameIDList[T any](ctx context.Context, items []T, attrName string) types.Set {
	tflog.Debug(ctx, "ParseNameIDList() "+attrName+" - "+fmt.Sprintf("%v", items))

	tflog.Debug(ctx, "ParseNameIDList.items", map[string]interface{}{
		"v": utils.InterfaceToJSONString(fmt.Sprintf("%v", items)),
		"T": utils.InterfaceToJSONString(fmt.Sprintf("%T", items)),
	})
	// Handle empty list - return null (len(nil) == 0 so nil check is redundant)
	if len(items) == 0 {
		tflog.Debug(ctx, "ParseNameIDList() - empty input list, returning null")
		return types.SetNull(idname.RefObjectType)
	}

	// Process each item into an attr.Value
	nameIDValues := make([]attr.Value, 0, len(items))
	for i, item := range items {
		obj := ParseNameID(ctx, item, attrName)
		if !obj.IsNull() && !obj.IsUnknown() { // Include only non-null/unknown values, adjust as needed
			// API lists may repeat references; Terraform sets require unique values.
			if !slices.ContainsFunc(nameIDValues, obj.Equal) {
				nameIDValues = append(nameIDValues, obj)
			}
		} else {
			tflog.Debug(ctx, "ParseNameIDList() - skipping null/unknown item at index "+fmt.Sprintf("%d", i))
		}
	}

	if len(nameIDValues) == 0 {
		tflog.Debug(ctx, "ParseNameIDList() - all items skipped, returning null")
		return types.SetNull(idname.RefObjectType)
	}

	setValue, setDiags := types.SetValueFrom(ctx, idname.RefObjectType, nameIDValues)
	if setDiags.HasError() || setValue.IsUnknown() {
		return types.SetNull(idname.RefObjectType)
	}
	return setValue
}

// ParseNameIDListOrEmptySet ParseNameIDListOrEmptySet is like ParseNameIDList but returns an empty set instead of null for empty lists
// This is used when the configuration explicitly has an empty set (e.g. country = []) to preserve that intent
func ParseNameIDListOrEmptySet[T any](ctx context.Context, items []T, attrName string) types.Set {
	tflog.Debug(ctx, "ParseNameIDListOrEmptySet() "+attrName+" - "+fmt.Sprintf("%v", items))

	// Handle empty list - return empty set instead of null
	if len(items) == 0 {
		tflog.Debug(ctx, "ParseNameIDListOrEmptySet() - empty input list, returning empty set")
		emptySet, _ := types.SetValue(idname.RefObjectType, []attr.Value{})
		return emptySet
	}

	// Process each item into an attr.Value
	nameIDValues := make([]attr.Value, 0, len(items))
	for i, item := range items {
		obj := ParseNameID(ctx, item, attrName)
		if !obj.IsNull() && !obj.IsUnknown() { // Include only non-null/unknown values, adjust as needed
			// API lists may repeat references; Terraform sets require unique values.
			if !slices.ContainsFunc(nameIDValues, obj.Equal) {
				nameIDValues = append(nameIDValues, obj)
			}
		} else {
			tflog.Debug(ctx, "ParseNameIDList() - skipping null/unknown item at index "+fmt.Sprintf("%d", i))
		}
	}

	// Convert to types.List using SetValueFrom
	setValue, _ := types.SetValueFrom(ctx, idname.RefObjectType, nameIDValues)
	return setValue
}

//nolint:gocyclo
func ParseNameID(ctx context.Context, item interface{}, attrName string) types.Object {
	tflog.Debug(ctx, "ParseNameID() "+attrName+" - "+fmt.Sprintf("%v", item))

	// Get the reflect.Value of the input
	itemValue := reflect.ValueOf(item)

	// Handle nil or invalid input (must be a struct, not a slice/array)
	if item == nil || itemValue.Kind() != reflect.Struct {
		if itemValue.Kind() == reflect.Pointer && !itemValue.IsNil() {
			itemValue = itemValue.Elem()
			if itemValue.Kind() != reflect.Struct {
				return types.ObjectNull(idname.RefAttrTypes)
			}
		} else {
			return types.ObjectNull(idname.RefAttrTypes)
		}
	}

	// Handle pointer to struct
	if itemValue.Kind() == reflect.Pointer {
		if itemValue.IsNil() {
			return types.ObjectNull(idname.RefAttrTypes)
		}
		itemValue = itemValue.Elem()
	}

	// Get Name and ID fields
	nameField := itemValue.FieldByName("Name")
	idField := itemValue.FieldByName("ID")

	if !nameField.IsValid() || !idField.IsValid() {
		tflog.Debug(ctx, "ParseNameID() !nameField.IsValid() - "+fmt.Sprintf("%v", nameField))
		return types.ObjectNull(idname.RefAttrTypes)
	}

	// Safely extract string values, handling pointers and empty values
	var nameValue types.String
	var idValue types.String

	// Handle ID field first (this is always populated)
	if idField.Kind() == reflect.Pointer {
		if idField.IsNil() {
			idValue = types.StringNull()
		} else {
			val := idField.Elem().String()
			if val == "" {
				idValue = types.StringNull()
			} else {
				idValue = types.StringValue(val)
			}
		}
	} else {
		val := idField.String()
		if val == "" {
			idValue = types.StringNull()
		} else {
			idValue = types.StringValue(val)
		}
	}

	// Handle name field - use UseStateForUnknown logic to prevent spurious diffs
	// If name would be empty/null, keep it as null to match Terraform's expectations
	// for computed optional fields
	if nameField.Kind() == reflect.Pointer {
		if nameField.IsNil() {
			nameValue = types.StringNull()
		} else {
			val := nameField.Elem().String()
			if val == "" {
				nameValue = types.StringNull()
			} else {
				// Only populate name if we have a valid value
				// The UseStateForUnknown() plan modifier in the schema should handle the rest
				nameValue = types.StringValue(val)
			}
		}
	} else {
		val := nameField.String()
		if val == "" {
			nameValue = types.StringNull()
		} else {
			nameValue = types.StringValue(val)
		}
	}

	// Create object value with proper null/value handling
	obj, diagstmp := types.ObjectValue(
		idname.RefAttrTypes,
		map[string]attr.Value{
			"name": nameValue,
			"id":   idValue,
		},
	)
	tflog.Debug(ctx, "ParseNameID() obj - "+obj.String())
	_ = diagstmp
	return obj
}
