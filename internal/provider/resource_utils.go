package provider

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	cato_models "github.com/catonetworks/cato-go-sdk/models" // Import the correct package
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/fatih/structs"
	"github.com/gobeam/stringy"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/spf13/cast"
)

func contains(nameToIdMap map[string]struct{}, name string) bool {
	_, exists := nameToIdMap[name]
	return exists
}

// getSiteById retrieves site information by site ID using EntityLookup
func getSiteNetworkInterface(ctx context.Context, client *catoClientData, siteID string, interfaceId string, interfaceIndex string, interfaceName string) (string, string, error) {
	tflog.Debug(ctx, "getSiteNetworkInterface()", map[string]interface{}{
		"siteID":         utils.InterfaceToJSONString(siteID),
		"interfaceId":    utils.InterfaceToJSONString(interfaceId),
		"interfaceName":  utils.InterfaceToJSONString(interfaceName),
		"interfaceIndex": utils.InterfaceToJSONString(interfaceIndex),
	})
	zeroInt64 := int64(0)
	// func (c *Client) EntityLookup(ctx context.Context, accountID string, typeArg cato_models.EntityType, limit *int64, from *int64, parent *cato_models.EntityInput, search *string, entityIDs []string, sort []*cato_models.SortInput, filters []*cato_models.LookupFilterInput, helperFields []string, interceptors ...clientv2.RequestInterceptor) (*EntityLookup, error) {
	site := &cato_models.EntityInput{
		Type: cato_models.EntityTypeSite,
		ID:   siteID,
	}
	// Add filter for interfaceId if set
	var entityIDs []string
	if interfaceId != "" {
		entityIDs = []string{interfaceId}
	} else {
		entityIDs = nil
	}
	networkInterfaceResponse, err := client.catov2.EntityLookup(ctx, client.AccountId, cato_models.EntityTypeNetworkInterface, &zeroInt64, nil, site, nil, entityIDs, nil, nil, nil)
	tflog.Warn(ctx, "getSiteNetworkInterfaceById.EntityLookup.networkInterfaceResponse", map[string]interface{}{
		"response": utils.InterfaceToJSONString(networkInterfaceResponse),
		"err":      utils.InterfaceToJSONString(err),
	})
	if err != nil {
		return "", "", err
	}
	items := networkInterfaceResponse.GetEntityLookup().GetItems()
	// Check for interfaceIndex
	tflog.Info(ctx, "getSiteNetworkInterface.inputs", map[string]interface{}{
		"interfaceIndex": interfaceIndex,
		"interfaceName":  interfaceName,
	})

	if interfaceIndex != "" {
		for _, item := range items {
			helperFields := item.GetHelperFields()
			curInterfaceId := item.GetEntity().GetID()
			curInterfaceIndex := cast.ToString(helperFields["interfaceId"])
			tflog.Info(ctx, "getSiteNetworkInterface.interfaceItem", map[string]interface{}{
				"interfaceItem":  item,
				"interfaceIndex": interfaceIndex,
				"interfaceName":  interfaceName,
				"check":          cast.ToString(helperFields["interfaceName"]) == interfaceName,
			})

			if len(curInterfaceIndex) == 1 {
				intVal, err := cast.ToIntE(curInterfaceIndex)
				if err == nil {
					curInterfaceIndex = fmt.Sprintf("INT_%d", intVal)
				}
			}
			if curInterfaceIndex == interfaceIndex {
				tflog.Info(ctx, "getSiteNetworkInterfaceById found interface by index", map[string]interface{}{
					"interfaceId":    curInterfaceId,
					"interfaceIndex": curInterfaceIndex,
				})
				return curInterfaceId, curInterfaceIndex, nil
			}
		}
		return "", "", fmt.Errorf("network interface with index '%s' not found in site '%s'", interfaceIndex, siteID)
	} else {
		// Confirm single record retiurned by Id
		if len(items) == 0 {
			return "", "", fmt.Errorf("network interface with ID '%s' not found in site '%s'", interfaceId, siteID)
		} else if len(items) == 1 {
			interfaceItem := items[0]
			// Return the first (and should be only) interface item
			helperFields := interfaceItem.GetHelperFields()
			curInterfaceIndex := cast.ToString(helperFields["interfaceId"])
			if len(curInterfaceIndex) == 1 {
				intVal, err := cast.ToIntE(curInterfaceIndex)
				if err == nil {
					curInterfaceIndex = fmt.Sprintf("INT_%d", intVal)
				}
			}
			interfaceId := interfaceItem.GetEntity().GetID()
			return interfaceId, curInterfaceIndex, nil
		} else {
			var curInterfaceIndex string
			var interfaceId string
			for _, interfaceItem := range items {
				helperFields := interfaceItem.GetHelperFields()
				curInterfaceName := cast.ToString(helperFields["interfaceName"])
				tflog.Info(ctx, "getSiteNetworkInterface.interfaceItem", map[string]interface{}{
					"interfaceItem":                         interfaceItem,
					"interfaceName":                         interfaceName,
					"curInterfaceName":                      curInterfaceName,
					"check curInterfaceName==interfaceName": (curInterfaceName == interfaceName),
				})
				if interfaceName == curInterfaceName {
					curInterfaceIndex = cast.ToString(helperFields["interfaceId"])
					if len(curInterfaceIndex) == 1 {
						intVal, err := cast.ToIntE(curInterfaceIndex)
						if err == nil {
							curInterfaceIndex = fmt.Sprintf("INT_%d", intVal)
						}
					}
					interfaceId = interfaceItem.GetEntity().GetID()
				}
			}
			if interfaceId == "" {
				return "", "", fmt.Errorf("network interface with name '%s' not found in site '%s'", interfaceName, siteID)
			}
			tflog.Info(ctx, "getSiteNetworkInterface.return", map[string]interface{}{
				"interfaceId":       interfaceId,
				"curInterfaceIndex": curInterfaceIndex,
			})
			return interfaceId, curInterfaceIndex, nil
		}
	}
}

func mapObjectList(ctx context.Context, srcItemObjList any, resp *resource.ReadResponse) []types.Object {
	vals := reflect.ValueOf(srcItemObjList)
	var objList []types.Object
	for i := range vals.Len() {
		objList = append(objList, mapStructFields(ctx, vals.Index(i), resp))
	}

	return objList
}

func mapStructFields(ctx context.Context, srcItemObj any, resp *resource.ReadResponse) types.Object {
	vals := reflect.ValueOf(srcItemObj)
	names := structs.Names(srcItemObj)
	tflog.Info(ctx, "pointer val: ",
		map[string]interface{}{
			"pointer_val": vals,
		})
	attrTypes := map[string]attr.Type{}
	attrValues := map[string]attr.Value{}

	for _, v := range names {
		attrTypes[v] = types.StringType
		attrValues[v] = basetypes.NewStringValue(vals.FieldByName(v).String())
	}

	newObj, diags := basetypes.NewObjectValueFrom(ctx, attrTypes, attrValues)
	resp.Diagnostics.Append(diags...)

	return newObj
}

func mapAttributeTypes(ctx context.Context, srcItemObj any, resp *resource.ReadResponse) map[string]attr.Type {
	attrTypes := map[string]attr.Type{}

	names := structs.Names(srcItemObj)
	vals := structs.Map(srcItemObj)
	for _, v := range names {
		intV := stringy.New(v).SnakeCase().ToLower()

		attrTypes[strings.ToLower(intV)] = convertGoTypeToTfType(ctx, vals[v])
	}

	return attrTypes
}

func reverseSlice[S ~[]E, E any](s S) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

func convertGoTypeToTfType(ctx context.Context, srcItemObj any) attr.Type {
	srcItemObjType := reflect.TypeOf(srcItemObj).String()
	tflog.Error(ctx, "srcItemObjType", map[string]interface{}{
		"srcItemObjType": srcItemObjType,
	})
	switch srcItemObjType {
	case "string":
		return types.StringType
	case "basetypes.StringValue":
		return types.StringType
	case "bool":
		return types.BoolType
	case "basetypes.BoolValue":
		return types.BoolType
	case "int":
		return types.Int32Type
	case "int64":
		return types.Int64Type
	case "float":
		return types.Float32Type
	case "float64":
		return types.Float64Type
	case "basetypes.ObjectValue":
		return types.ObjectType{}
	}

	return types.StringType
}

func parseList[T any](ctx context.Context, elemType attr.Type, items []T, attrName string) types.List {
	tflog.Debug(ctx, "parseList() "+attrName+" - "+fmt.Sprintf("%v", items))
	diags := make(diag.Diagnostics, 0)

	// Handle empty list - return null (len(nil) == 0 so nil check is redundant)
	if len(items) == 0 {
		tflog.Info(ctx, "parseList() - empty list, returning null")
		return types.ListNull(elemType)
	}

	tflog.Info(ctx, "parseList() - "+fmt.Sprintf("%v", items))

	// Convert to types.List using ListValueFrom
	listValue, listDiags := types.ListValueFrom(ctx, elemType, items)
	diags.Append(listDiags...)
	return listValue
}

func parseNameIDList[T any](ctx context.Context, items []T, attrName string) types.Set {
	tflog.Debug(ctx, "parseNameIDList() "+attrName+" - "+fmt.Sprintf("%v", items))
	diags := make(diag.Diagnostics, 0)

	tflog.Debug(ctx, "parseNameIDList.items", map[string]interface{}{
		"v": utils.InterfaceToJSONString(fmt.Sprintf("%v", items)),
		"T": utils.InterfaceToJSONString(fmt.Sprintf("%T", items)),
	})
	// Handle empty list - return null (len(nil) == 0 so nil check is redundant)
	if len(items) == 0 {
		tflog.Debug(ctx, "parseNameIDList() - empty input list, returning null")
		return types.SetNull(NameIDObjectType)
	}

	// Process each item into an attr.Value
	nameIDValues := make([]attr.Value, 0, len(items))
	for i, item := range items {
		obj := parseNameID(ctx, item, attrName)
		if !obj.IsNull() && !obj.IsUnknown() { // Include only non-null/unknown values, adjust as needed
			nameIDValues = append(nameIDValues, obj)
		} else {
			tflog.Debug(ctx, "parseNameIDList() - skipping null/unknown item at index "+fmt.Sprintf("%d", i))
		}
	}

	// Convert to types.List using SetValueFrom
	setValue, diagstmp := types.SetValueFrom(ctx, NameIDObjectType, nameIDValues)
	diags = append(diags, diagstmp...)
	return setValue
}

// parseNameIDListOrEmptySet is like parseNameIDList but returns an empty set instead of null for empty lists
// This is used when the configuration explicitly has an empty set (e.g. country = []) to preserve that intent
func parseNameIDListOrEmptySet[T any](ctx context.Context, items []T, attrName string) types.Set {
	tflog.Debug(ctx, "parseNameIDListOrEmptySet() "+attrName+" - "+fmt.Sprintf("%v", items))
	diags := make(diag.Diagnostics, 0)

	// Handle empty list - return empty set instead of null
	if len(items) == 0 {
		tflog.Debug(ctx, "parseNameIDListOrEmptySet() - empty input list, returning empty set")
		emptySet, diagstmp := types.SetValue(NameIDObjectType, []attr.Value{})
		diags = append(diags, diagstmp...)
		return emptySet
	}

	// Process each item into an attr.Value
	nameIDValues := make([]attr.Value, 0, len(items))
	for i, item := range items {
		obj := parseNameID(ctx, item, attrName)
		if !obj.IsNull() && !obj.IsUnknown() { // Include only non-null/unknown values, adjust as needed
			nameIDValues = append(nameIDValues, obj)
		} else {
			tflog.Debug(ctx, "parseNameIDList() - skipping null/unknown item at index "+fmt.Sprintf("%d", i))
		}
	}

	// Convert to types.List using SetValueFrom
	setValue, diagstmp := types.SetValueFrom(ctx, NameIDObjectType, nameIDValues)
	diags = append(diags, diagstmp...)
	return setValue
}

// parseNameIDListWithUnknownName creates a Set of NameID objects but keeps all names as unknown
// This is used for computed name fields with UseStateForUnknown() to prevent spurious diffs
func parseNameIDListWithUnknownName[T any](ctx context.Context, items []T, attrName string) types.Set {
	tflog.Debug(ctx, "parseNameIDListWithUnknownName() "+attrName+" - "+fmt.Sprintf("%v", items))
	diags := make(diag.Diagnostics, 0)

	tflog.Debug(ctx, "parseNameIDListWithUnknownName.items", map[string]interface{}{
		"v": utils.InterfaceToJSONString(fmt.Sprintf("%v", items)),
		"T": utils.InterfaceToJSONString(fmt.Sprintf("%T", items)),
	})
	// Handle empty list (len(nil) == 0 so nil check is redundant)
	if len(items) == 0 {
		tflog.Debug(ctx, "parseNameIDListWithUnknownName() - empty input list")
		return types.SetNull(NameIDObjectType)
	}

	// Process each item into an attr.Value using the unknown name variant
	nameIDValues := make([]attr.Value, 0, len(items))
	for i, item := range items {
		obj := parseNameIDWithUnknownName(ctx, item, attrName)
		if !obj.IsNull() { // Include non-null values (unknown is allowed)
			nameIDValues = append(nameIDValues, obj)
		} else {
			tflog.Debug(ctx, "parseNameIDListWithUnknownName() - skipping null item at index "+fmt.Sprintf("%d", i))
		}
	}

	// Convert to types.Set using SetValueFrom
	setValue, diagstmp := types.SetValueFrom(ctx, NameIDObjectType, nameIDValues)
	diags = append(diags, diagstmp...)
	return setValue
}

func parseNameID(ctx context.Context, item interface{}, attrName string) types.Object {
	tflog.Debug(ctx, "parseNameID() "+attrName+" - "+fmt.Sprintf("%v", item))
	diags := make(diag.Diagnostics, 0)

	// Get the reflect.Value of the input
	itemValue := reflect.ValueOf(item)

	// Handle nil or invalid input (must be a struct, not a slice/array)
	if item == nil || itemValue.Kind() != reflect.Struct {
		if itemValue.Kind() == reflect.Ptr && !itemValue.IsNil() {
			itemValue = itemValue.Elem()
			if itemValue.Kind() != reflect.Struct {
				return types.ObjectNull(NameIDAttrTypes)
			}
		} else {
			return types.ObjectNull(NameIDAttrTypes)
		}
	}

	// Handle pointer to struct
	if itemValue.Kind() == reflect.Ptr {
		if itemValue.IsNil() {
			return types.ObjectNull(NameIDAttrTypes)
		}
		itemValue = itemValue.Elem()
	}

	// Get Name and ID fields
	nameField := itemValue.FieldByName("Name")
	idField := itemValue.FieldByName("ID")

	if !nameField.IsValid() || !idField.IsValid() {
		tflog.Debug(ctx, "parseNameID() !nameField.IsValid() - "+fmt.Sprintf("%v", nameField))
		return types.ObjectNull(NameIDAttrTypes)
	}

	// Safely extract string values, handling pointers and empty values
	var nameValue types.String
	var idValue types.String

	// Handle ID field first (this is always populated)
	if idField.Kind() == reflect.Ptr {
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
	if nameField.Kind() == reflect.Ptr {
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
		NameIDAttrTypes,
		map[string]attr.Value{
			"name": nameValue,
			"id":   idValue,
		},
	)
	tflog.Debug(ctx, "parseNameID() obj - "+fmt.Sprintf("%v", obj))
	diags = append(diags, diagstmp...)
	return obj
}

// parseNameIDWithUnknownName creates a NameID object but keeps the name as unknown
// This is used for computed name fields with UseStateForUnknown() to prevent spurious diffs
func parseNameIDWithUnknownName(ctx context.Context, item interface{}, attrName string) types.Object {
	tflog.Debug(ctx, "parseNameIDWithUnknownName() "+attrName+" - "+fmt.Sprintf("%v", item))
	diags := make(diag.Diagnostics, 0)

	// Get the reflect.Value of the input
	itemValue := reflect.ValueOf(item)

	// Handle nil or invalid input (must be a struct, not a slice/array)
	if item == nil || itemValue.Kind() != reflect.Struct {
		if itemValue.Kind() == reflect.Ptr && !itemValue.IsNil() {
			itemValue = itemValue.Elem()
			if itemValue.Kind() != reflect.Struct {
				return types.ObjectNull(NameIDAttrTypes)
			}
		} else {
			return types.ObjectNull(NameIDAttrTypes)
		}
	}

	// Handle pointer to struct
	if itemValue.Kind() == reflect.Ptr {
		if itemValue.IsNil() {
			return types.ObjectNull(NameIDAttrTypes)
		}
		itemValue = itemValue.Elem()
	}

	// Get Name and ID fields
	nameField := itemValue.FieldByName("Name")
	idField := itemValue.FieldByName("ID")

	if !nameField.IsValid() || !idField.IsValid() {
		tflog.Debug(ctx, "parseNameIDWithUnknownName() !nameField.IsValid() - "+fmt.Sprintf("%v", nameField))
		return types.ObjectNull(NameIDAttrTypes)
	}

	// Safely extract string values, handling pointers and empty values
	var nameValue types.String
	var idValue types.String

	// Handle ID field first (this is always populated)
	if idField.Kind() == reflect.Ptr {
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

	// For computed name fields with UseStateForUnknown(), keep name as unknown
	// to let Terraform's plan modifier handle it properly
	nameValue = types.StringUnknown()

	// Create object value with proper null/unknown handling
	obj, diagstmp := types.ObjectValue(
		NameIDAttrTypes,
		map[string]attr.Value{
			"name": nameValue,
			"id":   idValue,
		},
	)
	tflog.Debug(ctx, "parseNameIDWithUnknownName() obj - "+fmt.Sprintf("%v", obj))
	diags = append(diags, diagstmp...)
	return obj
}

func parseFromToList[T any](ctx context.Context, items []T, attrName string) types.List {
	tflog.Debug(ctx, "parseFromToList() "+attrName+" - "+fmt.Sprintf("%v", items))
	diags := make(diag.Diagnostics, 0)

	// Handle empty list - return null (len(nil) == 0 so nil check is redundant)
	if len(items) == 0 {
		tflog.Debug(ctx, "parseFromToList() - empty input list, returning null")
		return types.ListNull(FromToObjectType)
	}
	// Process each item into an attr.Value
	fromToValues := make([]attr.Value, 0, len(items))
	for i, item := range items {
		obj := parseFromTo(ctx, item, attrName)
		if !obj.IsNull() && !obj.IsUnknown() { // Include only non-null/unknown values, adjust as needed
			fromToValues = append(fromToValues, obj)
		} else {
			tflog.Debug(ctx, "parseFromToList() - skipping null/unknown item at index "+fmt.Sprintf("%d", i))
		}
	}

	// Convert to types.List using ListValueFrom
	listValue, diagstmp := types.ListValueFrom(ctx, FromToObjectType, fromToValues)
	diags = append(diags, diagstmp...)
	return listValue
}

func parseFromTo(ctx context.Context, item interface{}, attrName string) types.Object {
	tflog.Debug(ctx, "parseFromTo() "+attrName+" - "+fmt.Sprintf("%v", item))
	diags := make(diag.Diagnostics, 0)

	if item == nil {
		tflog.Debug(ctx, "parseFromTo() - nil")
		return types.ObjectNull(FromToAttrTypes)
	}

	// Get the reflect.Value of the input
	itemValue := reflect.ValueOf(item)

	// Handle nil or invalid input
	tflog.Debug(ctx, "parseFromTo() itemValue.Kind()- "+fmt.Sprintf("%v", itemValue.Kind()))
	if item == nil || itemValue.Kind() != reflect.Struct {
		if itemValue.Kind() == reflect.Ptr && !itemValue.IsNil() {
			itemValue = itemValue.Elem()
			// Keep dereferencing until we get to a struct or can't anymore
			for itemValue.Kind() == reflect.Ptr && !itemValue.IsNil() {
				itemValue = itemValue.Elem()
			}
			if itemValue.Kind() != reflect.Struct {
				return types.ObjectNull(FromToAttrTypes)
			}
		} else {
			return types.ObjectNull(FromToAttrTypes)
		}
	}

	// Handle pointer to struct
	if itemValue.Kind() == reflect.Ptr {
		if itemValue.IsNil() {
			return types.ObjectNull(FromToAttrTypes)
		}
		itemValue = itemValue.Elem()
	}

	// Get Name and ID fields
	fromField := itemValue.FieldByName("From")
	toField := itemValue.FieldByName("To")
	tflog.Info(ctx, "parseFromTo() value", map[string]interface{}{
		"fromField": fromField,
		"toField":   toField,
	})

	if !fromField.IsValid() || !toField.IsValid() {
		tflog.Debug(ctx, "parseFromTo() fromField.IsValid() - "+fmt.Sprintf("%v", fromField))
		tflog.Debug(ctx, "parseFromTo() toField.IsValid() - "+fmt.Sprintf("%v", toField))
		return types.ObjectNull(FromToAttrTypes)
	}

	// Create object value
	obj, diagstmp := types.ObjectValue(
		FromToAttrTypes,
		map[string]attr.Value{
			"from": basetypes.NewStringValue(fromField.String()),
			"to":   basetypes.NewStringValue(toField.String()),
		},
	)
	tflog.Debug(ctx, "parseFromTo() obj - "+fmt.Sprintf("%v", obj))
	diags = append(diags, diagstmp...)
	return obj
}

func parseFromToDays(ctx context.Context, item interface{}, attrName string) types.Object {
	tflog.Debug(ctx, "parseFromToDays() "+attrName+" - "+fmt.Sprintf("%v", item))
	diags := make(diag.Diagnostics, 0)

	if item == nil {
		tflog.Debug(ctx, "parseFromToDays() - nil")
		return types.ObjectNull(FromToAttrTypes)
	}

	// Get the reflect.Value of the input
	itemValue := reflect.ValueOf(item)

	// Handle nil or invalid input
	tflog.Debug(ctx, "parseFromToDays() itemValue.Kind()- "+fmt.Sprintf("%v", itemValue.Kind()))
	if item == nil || itemValue.Kind() != reflect.Struct {
		if itemValue.Kind() == reflect.Ptr && !itemValue.IsNil() {
			itemValue = itemValue.Elem()
			// Keep dereferencing until we get to a struct or can't anymore
			for itemValue.Kind() == reflect.Ptr && !itemValue.IsNil() {
				itemValue = itemValue.Elem()
			}
			if itemValue.Kind() != reflect.Struct {
				return types.ObjectNull(FromToDaysAttrTypes)
			}
		} else {
			return types.ObjectNull(FromToDaysAttrTypes)
		}
	}

	// Handle pointer to struct
	if itemValue.Kind() == reflect.Ptr {
		if itemValue.IsNil() {
			return types.ObjectNull(FromToDaysAttrTypes)
		}
		itemValue = itemValue.Elem()
	}

	// Get Name and ID fields
	fromField := itemValue.FieldByName("From")
	toField := itemValue.FieldByName("To")
	daysField := itemValue.FieldByName("Days")

	if !fromField.IsValid() || !toField.IsValid() {
		tflog.Debug(ctx, "parseFromTo() fromField.IsValid() - "+fmt.Sprintf("%v", fromField))
		tflog.Debug(ctx, "parseFromTo() toField.IsValid() - "+fmt.Sprintf("%v", toField))
		return types.ObjectNull(FromToDaysAttrTypes)
	}

	// Create object value
	obj, diagstmp := types.ObjectValue(
		FromToDaysAttrTypes,
		map[string]attr.Value{
			"from": basetypes.NewStringValue(fromField.String()),
			"to":   basetypes.NewStringValue(toField.String()),
			"days": parseList(ctx, types.StringType, daysField.Interface().([]cato_models.DayOfWeek), "rule.schedule.custom_recurring.days"),
		},
	)
	tflog.Debug(ctx, "parseFromToDays() obj - "+fmt.Sprintf("%v", obj))
	diags = append(diags, diagstmp...)
	return obj
}

func parseCustomService(ctx context.Context, item interface{}, attrName string) types.Object {
	tflog.Debug(ctx, "parseCustomService() "+attrName+" - "+fmt.Sprintf("%v", item))
	diags := make(diag.Diagnostics, 0)

	// Get the reflect.Value of the input
	itemValue := reflect.ValueOf(item)

	// Handle nil or invalid input
	if item == nil || itemValue.Kind() != reflect.Struct {
		if itemValue.Kind() == reflect.Ptr && !itemValue.IsNil() {
			itemValue = itemValue.Elem()
			// Keep dereferencing until we get to a struct or can't anymore
			for itemValue.Kind() == reflect.Ptr && !itemValue.IsNil() {
				itemValue = itemValue.Elem()
			}
			if itemValue.Kind() != reflect.Struct {
				return types.ObjectNull(CustomServiceAttrTypes)
			}
		} else {
			return types.ObjectNull(CustomServiceAttrTypes)
		}
	}

	// Handle pointer to struct
	if itemValue.Kind() == reflect.Ptr {
		tflog.Debug(ctx, "parseCustomService() itemValue.Kind()- "+fmt.Sprintf("%v", itemValue))
		if itemValue.IsNil() {
			return types.ObjectNull(CustomServiceAttrTypes)
		}
		itemValue = itemValue.Elem()
		tflog.Debug(ctx, "parseCustomService() itemValue.Elem()- "+fmt.Sprintf("%v", itemValue))
	}

	// Get fields
	portField := itemValue.FieldByName("Port")
	protocolField := itemValue.FieldByName("Protocol")
	portRangeField := itemValue.FieldByName("PortRange")

	tflog.Debug(ctx, "parseCustomService() portField - "+fmt.Sprintf("%v", portField))
	tflog.Debug(ctx, "parseCustomService() protocolField - "+fmt.Sprintf("%v", protocolField))
	tflog.Debug(ctx, "parseCustomService() portRangeField - "+fmt.Sprintf("%v", portRangeField))
	// Handle port_range first to check if it's set
	var portRangeVal types.Object
	var hasPortRange bool
	if portRangeField.Kind() == reflect.Ptr {
		if portRangeField.IsNil() {
			portRangeVal = types.ObjectNull(FromToAttrTypes)
			hasPortRange = false
		} else {
			portRangeField = portRangeField.Elem()
			hasPortRange = true
		}
	}
	if portRangeField.IsValid() && hasPortRange {
		from := portRangeField.FieldByName("From")
		to := portRangeField.FieldByName("To")
		var diagsTmp diag.Diagnostics
		portRangeVal, diagsTmp = types.ObjectValue(
			FromToAttrTypes,
			map[string]attr.Value{
				"from": types.StringValue(from.String()),
				"to":   types.StringValue(to.String()),
			},
		)
		diags = append(diags, diagsTmp...)
	} else if !hasPortRange {
		portRangeVal = types.ObjectNull(FromToAttrTypes)
	}

	// Handle port field - match config logic behavior
	// When port_range is set, port should always be null even if API returns empty array
	var portList types.List
	if hasPortRange {
		// When port_range is present, port must be null to match schema validation
		portList = types.ListNull(types.StringType)
	} else if portField.IsValid() && portField.Kind() == reflect.Slice {
		// Only process port field if port_range is not set
		if portField.IsNil() {
			// Source data was null - don't include port field in state to match config
			portList = types.ListNull(types.StringType)
		} else if portField.Len() == 0 {
			// Source data was empty array [] - treat as null to avoid inconsistency
			// Note: empty array from API likely means port_range is set
			portList = types.ListNull(types.StringType)
		} else {
			// Source data has values - include them
			ports := make([]attr.Value, portField.Len())
			for i := range portField.Len() {
				portValue := portField.Index(i)
				var portStr string
				switch portValue.Kind() {
				case reflect.String:
					portStr = portValue.String()
				case reflect.Int, reflect.Int64:
					portStr = fmt.Sprintf("%d", portValue.Int())
				default:
					tflog.Info(ctx, "parseCustomService() unsupported port type - "+fmt.Sprintf("%v", portValue.Kind()))
					portStr = fmt.Sprintf("%v", portValue.Interface())
				}
				ports[i] = types.StringValue(portStr)
			}
			var diagsTmp diag.Diagnostics
			portList, diagsTmp = types.ListValue(types.StringType, ports)
			diags = append(diags, diagsTmp...)
		}
	} else {
		// Invalid or non-slice field
		portList = types.ListNull(types.StringType)
	}

	// Handle protocol
	var protocolVal types.String
	if protocolField.IsValid() {
		protocolVal = types.StringValue(protocolField.String())
	} else {
		protocolVal = types.StringNull()
	}

	// Create final custom service object
	obj, diagstmp := types.ObjectValue(
		CustomServiceAttrTypes,
		map[string]attr.Value{
			"port":       portList,
			"port_range": portRangeVal,
			"protocol":   protocolVal,
		},
	)
	tflog.Debug(ctx, "parseCustomService() obj - "+fmt.Sprintf("%v", obj))
	diags = append(diags, diagstmp...)
	return obj
}

// parseCustomServiceIp handles the custom service IP object from API response
func parseCustomServiceIp(ctx context.Context, item interface{}, attrName string) types.Object {
	tflog.Debug(ctx, "parseCustomServiceIp() "+attrName+" - "+fmt.Sprintf("%v", item))
	diags := make(diag.Diagnostics, 0)

	// Get the reflect.Value of the input
	itemValue := reflect.ValueOf(item)

	// Handle nil or invalid input
	if item == nil || itemValue.Kind() != reflect.Struct {
		if itemValue.Kind() == reflect.Ptr && !itemValue.IsNil() {
			itemValue = itemValue.Elem()
			for itemValue.Kind() == reflect.Ptr && !itemValue.IsNil() {
				itemValue = itemValue.Elem()
			}
			if itemValue.Kind() != reflect.Struct {
				return types.ObjectNull(CustomServiceIpAttrTypes)
			}
		} else {
			return types.ObjectNull(CustomServiceIpAttrTypes)
		}
	}

	// Handle pointer to struct
	if itemValue.Kind() == reflect.Ptr {
		if itemValue.IsNil() {
			return types.ObjectNull(CustomServiceIpAttrTypes)
		}
		itemValue = itemValue.Elem()
	}

	// Get fields
	nameField := itemValue.FieldByName("Name")
	ipField := itemValue.FieldByName("IP")
	ipRangeField := itemValue.FieldByName("IPRange")

	// Handle name
	var nameVal types.String
	if nameField.IsValid() && nameField.String() != "" {
		nameVal = types.StringValue(nameField.String())
	} else {
		nameVal = types.StringNull()
	}

	// Handle IP - only include if it has a value
	var ipVal types.String
	if ipField.IsValid() && !ipField.IsZero() && ipField.String() != "" {
		ipVal = types.StringValue(ipField.String())
	} else {
		ipVal = types.StringNull()
	}

	// Handle IP range
	var ipRangeVal types.Object
	if ipRangeField.Kind() == reflect.Ptr {
		if ipRangeField.IsNil() {
			ipRangeVal = types.ObjectNull(FromToAttrTypes)
		} else {
			ipRangeField = ipRangeField.Elem()
			from := ipRangeField.FieldByName("From")
			to := ipRangeField.FieldByName("To")
			var diagsTmp diag.Diagnostics
			ipRangeVal, diagsTmp = types.ObjectValue(
				FromToAttrTypes,
				map[string]attr.Value{
					"from": types.StringValue(from.String()),
					"to":   types.StringValue(to.String()),
				},
			)
			diags = append(diags, diagsTmp...)
		}
	} else if ipRangeField.IsValid() {
		from := ipRangeField.FieldByName("From")
		to := ipRangeField.FieldByName("To")
		var diagsTmp diag.Diagnostics
		ipRangeVal, diagsTmp = types.ObjectValue(
			FromToAttrTypes,
			map[string]attr.Value{
				"from": types.StringValue(from.String()),
				"to":   types.StringValue(to.String()),
			},
		)
		diags = append(diags, diagsTmp...)
	} else {
		ipRangeVal = types.ObjectNull(FromToAttrTypes)
	}

	// Create final custom service IP object
	obj, diagstmp := types.ObjectValue(
		CustomServiceIpAttrTypes,
		map[string]attr.Value{
			"name":     nameVal,
			"ip":       ipVal,
			"ip_range": ipRangeVal,
		},
	)
	tflog.Debug(ctx, "parseCustomServiceIp() obj - "+fmt.Sprintf("%v", obj))
	diags = append(diags, diagstmp...)
	return obj
}

// parseExceptionCustomService handles custom services in exceptions which use "portRangeCustomService" field instead of "portRange"
func parseExceptionCustomService(ctx context.Context, item interface{}, attrName string) types.Object {
	tflog.Debug(ctx, "parseExceptionCustomService() "+attrName+" - "+fmt.Sprintf("%v", item))
	diags := make(diag.Diagnostics, 0)

	// Get the reflect.Value of the input
	itemValue := reflect.ValueOf(item)

	// Handle nil or invalid input
	if item == nil || itemValue.Kind() != reflect.Struct {
		if itemValue.Kind() == reflect.Ptr && !itemValue.IsNil() {
			itemValue = itemValue.Elem()
			// Keep dereferencing until we get to a struct or can't anymore
			for itemValue.Kind() == reflect.Ptr && !itemValue.IsNil() {
				itemValue = itemValue.Elem()
			}
			if itemValue.Kind() != reflect.Struct {
				return types.ObjectNull(CustomServiceAttrTypes)
			}
		} else {
			return types.ObjectNull(CustomServiceAttrTypes)
		}
	}

	// Handle pointer to struct
	if itemValue.Kind() == reflect.Ptr {
		tflog.Debug(ctx, "parseExceptionCustomService() itemValue.Kind()- "+fmt.Sprintf("%v", itemValue))
		if itemValue.IsNil() {
			return types.ObjectNull(CustomServiceAttrTypes)
		}
		itemValue = itemValue.Elem()
		tflog.Debug(ctx, "parseExceptionCustomService() itemValue.Elem()- "+fmt.Sprintf("%v", itemValue))
	}

	// Get fields - note that exceptions use "PortRangeCustomService" instead of "PortRange"
	portField := itemValue.FieldByName("Port")
	protocolField := itemValue.FieldByName("Protocol")
	portRangeField := itemValue.FieldByName("PortRangeCustomService") // Different field name for exceptions

	tflog.Debug(ctx, "parseExceptionCustomService() portField - "+fmt.Sprintf("%v", portField))
	tflog.Debug(ctx, "parseExceptionCustomService() protocolField - "+fmt.Sprintf("%v", protocolField))
	tflog.Debug(ctx, "parseExceptionCustomService() portRangeField - "+fmt.Sprintf("%v", portRangeField))

	// Handle port_range first (from PortRangeCustomService field) to check if it's set
	var portRangeVal types.Object
	var hasPortRange bool
	if portRangeField.Kind() == reflect.Ptr {
		if portRangeField.IsNil() {
			portRangeVal = types.ObjectNull(FromToAttrTypes)
			hasPortRange = false
		} else {
			portRangeField = portRangeField.Elem()
			hasPortRange = true
		}
	}
	if portRangeField.IsValid() && !portRangeField.IsZero() && hasPortRange {
		from := portRangeField.FieldByName("From")
		to := portRangeField.FieldByName("To")
		if from.IsValid() && to.IsValid() {
			var diagsTmp diag.Diagnostics
			portRangeVal, diagsTmp = types.ObjectValue(
				FromToAttrTypes,
				map[string]attr.Value{
					"from": types.StringValue(from.String()),
					"to":   types.StringValue(to.String()),
				},
			)
			diags = append(diags, diagsTmp...)
		} else {
			portRangeVal = types.ObjectNull(FromToAttrTypes)
			hasPortRange = false
		}
	} else if !hasPortRange {
		portRangeVal = types.ObjectNull(FromToAttrTypes)
	}

	// Handle port field - match config logic behavior
	// When port_range is set, port should always be null even if API returns empty array
	var portList types.List
	if hasPortRange {
		// When port_range is present, port must be null to match schema validation
		portList = types.ListNull(types.StringType)
	} else if portField.IsValid() && portField.Kind() == reflect.Slice {
		// Only process port field if port_range is not set
		if portField.IsNil() {
			// Source data was null - don't include port field in state to match config
			portList = types.ListNull(types.StringType)
		} else if portField.Len() == 0 {
			// Source data was empty array [] - treat as null to avoid inconsistency
			// Note: empty array from API likely means port_range is set
			portList = types.ListNull(types.StringType)
		} else {
			// Source data has values - include them
			ports := make([]attr.Value, portField.Len())
			for i := range portField.Len() {
				portValue := portField.Index(i)
				var portStr string
				switch portValue.Kind() {
				case reflect.String:
					portStr = portValue.String()
				case reflect.Int, reflect.Int64:
					portStr = fmt.Sprintf("%d", portValue.Int())
				default:
					tflog.Info(ctx, "parseExceptionCustomService() unsupported port type - "+fmt.Sprintf("%v", portValue.Kind()))
					portStr = fmt.Sprintf("%v", portValue.Interface())
				}
				ports[i] = types.StringValue(portStr)
			}
			var diagsTmp diag.Diagnostics
			portList, diagsTmp = types.ListValue(types.StringType, ports)
			diags = append(diags, diagsTmp...)
		}
	} else {
		// Invalid or non-slice field
		portList = types.ListNull(types.StringType)
	}

	// Handle protocol
	var protocolVal types.String
	if protocolField.IsValid() {
		protocolVal = types.StringValue(protocolField.String())
	} else {
		protocolVal = types.StringNull()
	}

	// Create final custom service object
	obj, diagstmp := types.ObjectValue(
		CustomServiceAttrTypes,
		map[string]attr.Value{
			"port":       portList,
			"port_range": portRangeVal,
			"protocol":   protocolVal,
		},
	)
	tflog.Debug(ctx, "parseExceptionCustomService() obj - "+fmt.Sprintf("%v", obj))
	diags = append(diags, diagstmp...)
	return obj
}

func parseTimeString(timeStr string) (string, error) {
	t, err := time.Parse("2006-01-02T15:04:05", timeStr)
	if err != nil {
		return "", err
	}
	// If all formats fail, return an error
	return t.Format("2006-01-02T15:04:05"), nil
}

func parseTimeStringWithTZ(timeStr string) (string, error) {
	t, err := time.Parse("2006-01-02T15:04:05", timeStr)
	if err != nil {
		return "", err
	}
	// If all formats fail, return an error
	return t.Format("2006-01-02T15:04:05+00:00"), nil
}

// getActivePeriodString safely converts API active period string fields to Terraform string values
// Returns null string if empty, otherwise returns the string value
func getActivePeriodString(value *string) types.String {
	if value == nil || *value == "" {
		return types.StringNull()
	}
	return types.StringValue(*value)
}
