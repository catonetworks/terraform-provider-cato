package provider

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	cato_models "github.com/catonetworks/cato-go-sdk/models" // Import the correct package
	"github.com/fatih/structs"
	"github.com/gobeam/stringy"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func contains(nameToIdMap map[string]struct{}, name string) bool {
	_, exists := nameToIdMap[name]
	return exists
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
	tflog.Warn(ctx, "parseList() "+attrName+" - "+fmt.Sprintf("%v", items))
	diags := make(diag.Diagnostics, 0)

	if items == nil || len(items) == 0 {
		tflog.Info(ctx, "parseList() - nil or empty input list, returning null")
		return types.ListNull(elemType)
	}

	tflog.Info(ctx, "parseList() - "+fmt.Sprintf("%v", items))

	// Convert to types.List using ListValueFrom
	listValue, listDiags := types.ListValueFrom(ctx, elemType, items)
	diags.Append(listDiags...)
	return listValue
}

func parseNameIDList[T any](ctx context.Context, items []T, attrName string) types.Set {
	tflog.Warn(ctx, "parseNameIDList() "+attrName+" - "+fmt.Sprintf("%v", items))
	diags := make(diag.Diagnostics, 0)

	// Handle nil or empty list
	if items == nil || len(items) == 0 {
		tflog.Warn(ctx, "parseNameIDList() - nil or empty input list")
		return types.SetNull(NameIDObjectType)
	}

	// Process each item into an attr.Value
	nameIDValues := make([]attr.Value, 0, len(items))
	for i, item := range items {
		obj := parseNameID(ctx, item, attrName)
		if !obj.IsNull() && !obj.IsUnknown() { // Include only non-null/unknown values, adjust as needed
			nameIDValues = append(nameIDValues, obj)
		} else {
			tflog.Warn(ctx, "parseNameIDList() - skipping null/unknown item at index "+fmt.Sprintf("%d", i))
		}
	}

	// Convert to types.List using SetValueFrom
	setValue, diagstmp := types.SetValueFrom(ctx, NameIDObjectType, nameIDValues)
	diags = append(diags, diagstmp...)
	return setValue
}

func parseNameID(ctx context.Context, item interface{}, attrName string) types.Object {
	tflog.Warn(ctx, "parseNameID() "+attrName+" - "+fmt.Sprintf("%v", item))
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
		tflog.Warn(ctx, "parseNameID() !nameField.IsValid() - "+fmt.Sprintf("%v", nameField))
		return types.ObjectNull(NameIDAttrTypes)
	}

	// Create object value
	obj, diagstmp := types.ObjectValue(
		NameIDAttrTypes,
		map[string]attr.Value{
			"name": basetypes.NewStringValue(nameField.String()),
			"id":   basetypes.NewStringValue(idField.String()),
		},
	)
	tflog.Warn(ctx, "parseNameID() obj - "+fmt.Sprintf("%v", obj))
	diags = append(diags, diagstmp...)
	return obj
}

func parseFromToList[T any](ctx context.Context, items []T, attrName string) types.List {
	tflog.Warn(ctx, "parseFromToList() "+attrName+" - "+fmt.Sprintf("%v", items))
	diags := make(diag.Diagnostics, 0)

	// Handle nil or empty list
	if items == nil || len(items) == 0 {
		tflog.Warn(ctx, "parseFromToList() - nil or empty input list")
		return types.ListNull(FromToObjectType)
	}
	// Process each item into an attr.Value
	fromToValues := make([]attr.Value, 0, len(items))
	for i, item := range items {
		obj := parseFromTo(ctx, item, attrName)
		if !obj.IsNull() && !obj.IsUnknown() { // Include only non-null/unknown values, adjust as needed
			fromToValues = append(fromToValues, obj)
		} else {
			tflog.Warn(ctx, "parseFromToList() - skipping null/unknown item at index "+fmt.Sprintf("%d", i))
		}
	}

	// Convert to types.List using ListValueFrom
	listValue, diagstmp := types.ListValueFrom(ctx, FromToObjectType, fromToValues)
	diags = append(diags, diagstmp...)
	return listValue
}

func parseFromTo(ctx context.Context, item interface{}, attrName string) types.Object {
	tflog.Warn(ctx, "parseFromTo() "+attrName+" - "+fmt.Sprintf("%v", item))
	diags := make(diag.Diagnostics, 0)

	if item == nil {
		tflog.Warn(ctx, "parseFromTo() - nil")
		return types.ObjectNull(FromToAttrTypes)
	}

	// Get the reflect.Value of the input
	itemValue := reflect.ValueOf(item)

	// Handle nil or invalid input
	tflog.Warn(ctx, "parseFromTo() itemValue.Kind()- "+fmt.Sprintf("%v", itemValue.Kind()))
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
		tflog.Warn(ctx, "parseFromTo() fromField.IsValid() - "+fmt.Sprintf("%v", fromField))
		tflog.Warn(ctx, "parseFromTo() toField.IsValid() - "+fmt.Sprintf("%v", toField))
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
	tflog.Warn(ctx, "parseFromTo() obj - "+fmt.Sprintf("%v", obj))
	diags = append(diags, diagstmp...)
	return obj
}

func parseFromToDays(ctx context.Context, item interface{}, attrName string) types.Object {
	tflog.Warn(ctx, "parseFromToDays() "+attrName+" - "+fmt.Sprintf("%v", item))
	diags := make(diag.Diagnostics, 0)

	if item == nil {
		tflog.Warn(ctx, "parseFromToDays() - nil")
		return types.ObjectNull(FromToAttrTypes)
	}

	// Get the reflect.Value of the input
	itemValue := reflect.ValueOf(item)

	// Handle nil or invalid input
	tflog.Warn(ctx, "parseFromToDays() itemValue.Kind()- "+fmt.Sprintf("%v", itemValue.Kind()))
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
		tflog.Warn(ctx, "parseFromTo() fromField.IsValid() - "+fmt.Sprintf("%v", fromField))
		tflog.Warn(ctx, "parseFromTo() toField.IsValid() - "+fmt.Sprintf("%v", toField))
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
	tflog.Warn(ctx, "parseFromToDays() obj - "+fmt.Sprintf("%v", obj))
	diags = append(diags, diagstmp...)
	return obj
}

func parseCustomService(ctx context.Context, item interface{}, attrName string) types.Object {
	tflog.Warn(ctx, "parseCustomService() "+attrName+" - "+fmt.Sprintf("%v", item))
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
		tflog.Warn(ctx, "parseCustomService() itemValue.Kind()- "+fmt.Sprintf("%v", itemValue))
		if itemValue.IsNil() {
			return types.ObjectNull(CustomServiceAttrTypes)
		}
		itemValue = itemValue.Elem()
		tflog.Warn(ctx, "parseCustomService() itemValue.Elem()- "+fmt.Sprintf("%v", itemValue))
	}

	// Get fields
	portField := itemValue.FieldByName("Port")
	protocolField := itemValue.FieldByName("Protocol")
	portRangeField := itemValue.FieldByName("PortRange")

	tflog.Warn(ctx, "parseCustomService() portField - "+fmt.Sprintf("%v", portField))
	tflog.Warn(ctx, "parseCustomService() protocolField - "+fmt.Sprintf("%v", protocolField))
	tflog.Warn(ctx, "parseCustomService() portRangeField - "+fmt.Sprintf("%v", portRangeField))
	// Handle port field (allowing null)
	var portList types.List
	if portField.IsValid() && portField.Kind() == reflect.Slice {
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
				tflog.Warn(ctx, "parseCustomService() unsupported port type - "+fmt.Sprintf("%v", portValue.Kind()))
				portStr = fmt.Sprintf("%v", portValue.Interface())
			}
			ports[i] = types.StringValue(portStr)
			// ports[i] = types.StringNull()
		}
		var diagsTmp diag.Diagnostics
		portList, diagsTmp = types.ListValue(types.StringType, ports)
		diags = append(diags, diagsTmp...)
	} else {
		portList = types.ListNull(types.StringType) // Explicit null handling
	}

	// Handle protocol
	var protocolVal types.String
	if protocolField.IsValid() {
		protocolVal = types.StringValue(protocolField.String())
	} else {
		protocolVal = types.StringNull()
	}

	// Handle port_range
	var portRangeVal types.Object
	if portRangeField.Kind() == reflect.Ptr {
		if portRangeField.IsNil() {
			portRangeVal = types.ObjectNull(FromToAttrTypes)
		}
		portRangeField = portRangeField.Elem()
	}
	if portRangeField.IsValid() {
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
	} else {
		portRangeVal = types.ObjectNull(FromToAttrTypes)
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
	tflog.Warn(ctx, "parseCustomService() obj - "+fmt.Sprintf("%v", obj))
	diags = append(diags, diagstmp...)
	return obj
}
