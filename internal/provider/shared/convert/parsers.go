package convert

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type int64er interface{ GetInt64() int64 }

var FromToObjectType = types.ObjectType{AttrTypes: FromToAttrTypes}
var FromToAttrTypes = map[string]attr.Type{
	"from": types.StringType,
	"to":   types.StringType,
}

func StringPointerForOptionalInput(value types.String) *string {
	if value.IsNull() || value.IsUnknown() || value.ValueString() == "" {
		return nil
	}
	return value.ValueStringPointer()
}

// TranslatedSubnetForAPIInput omits translated_subnet from API payloads when the attribute is not
// explicitly set in Terraform config, even if plan/state still carry an API-hydrated value
// (e.g. equal to subnet when Static Range Translation is disabled).
func TranslatedSubnetForAPIInput(configValue, planValue types.String) *string {
	if configValue.IsNull() || configValue.IsUnknown() {
		return nil
	}
	return StringPointerForOptionalInput(planValue)
}

// StringPointerValue handles empty string API values, returns StringNull or StringValue(""), based on the plan
func StringPointerValue(s *string, plan types.String) types.String {
	if (s == nil) || (*s == "") {
		if plan.IsNull() {
			return types.StringNull()
		}
		return types.StringValue("")
	}
	return types.StringPointerValue(s)
}

// stringSetFunc parses a slice of convertible items into a types.Set of strings
func stringSetFunc[T any](ctx context.Context, stringers []T, convert func(x T) string, diags *diag.Diagnostics,
) types.Set {
	// null value
	if stringers == nil {
		return types.SetNull(types.StringType)
	}

	// existing empty list
	if len(stringers) == 0 {
		val, valueDiag := types.SetValue(types.StringType, nil)
		diags.Append(valueDiag...)
		return val
	}

	// make []types.String
	stringSlice := make([]types.String, 0, len(stringers))
	for _, o := range stringers {
		stringSlice = append(stringSlice, types.StringValue(convert(o)))
	}
	// convert to types.Set
	stringSet, valueDiag := types.SetValueFrom(ctx, types.StringType, stringSlice)
	diags.Append(valueDiag...)

	return stringSet
}

// StringSet parses a slice of fmt.Stringer into a types.Set of strings
func StringSet[T fmt.Stringer](ctx context.Context, stringers []T, diags *diag.Diagnostics) types.Set {
	return stringSetFunc(ctx, stringers, func(s T) string { return s.String() }, diags)
}

// stringListFunc parses a slice of any type that can be converted to string,
// returns TF list of StringType
func stringListFunc[T any](ctx context.Context, stringers []T, convert func(x T) string, diags *diag.Diagnostics) types.List {
	// null value
	if stringers == nil {
		return types.ListNull(types.StringType)
	}

	// existing empty list
	if len(stringers) == 0 {
		val, valueDiag := types.ListValue(types.StringType, nil)
		diags.Append(valueDiag...)
		return val
	}

	// make []types.String
	stringSlice := make([]types.String, 0, len(stringers))
	for _, o := range stringers {
		s := convert(o)
		stringSlice = append(stringSlice, types.StringValue(s))
	}
	// convert to types.List
	stringList, valueDiag := types.ListValueFrom(ctx, types.StringType, stringSlice)
	diags.Append(valueDiag...)

	return stringList
}

// ParseStringerList parses a slice of fmt.Stringer into a types.List of strings
func ParseStringerList(ctx context.Context, stringers []fmt.Stringer, diags *diag.Diagnostics) types.List {
	return stringListFunc(ctx, stringers,
		func(s fmt.Stringer) string { return s.String() },
		diags)
}

// ParseStringList parses a slice of fmt.Stringer into a types.List of strings
func ParseStringList[T ~string](ctx context.Context, stringers []T, diags *diag.Diagnostics) types.List {
	return stringListFunc(ctx, stringers,
		func(s T) string { return string(s) },
		diags)
}

// ParseInt64List parses a slice of int64er into a types.List of int64
func ParseInt64List[T int64er](ctx context.Context, ints []T, diags *diag.Diagnostics) types.List {
	// null value
	if ints == nil {
		return types.ListNull(types.Int64Type)
	}

	// existing empty list
	if len(ints) == 0 {
		val, valueDiag := types.ListValue(types.Int64Type, nil)
		diags.Append(valueDiag...)
		return val
	}

	// make []types.Int64
	int64Slice := make([]types.Int64, 0, len(ints))
	for _, o := range ints {
		int64Slice = append(int64Slice, types.Int64Value(o.GetInt64()))
	}
	// convert to types.List
	int64List, valueDiag := types.ListValueFrom(ctx, types.Int64Type, int64Slice)
	diags.Append(valueDiag...)

	return int64List
}

func ParseList[T any](ctx context.Context, elemType attr.Type, items []T, attrName string) types.List {
	tflog.Debug(ctx, "parseList() "+attrName+" - "+fmt.Sprintf("%v", items))

	// Handle empty list - return null (len(nil) == 0 so nil check is redundant)
	if len(items) == 0 {
		tflog.Info(ctx, "parseList() - empty list, returning null")
		return types.ListNull(elemType)
	}

	tflog.Info(ctx, "parseList() - "+fmt.Sprintf("%v", items))

	// Convert to types.List using ListValueFrom
	listValue, listDiags := types.ListValueFrom(ctx, elemType, items)
	if listDiags.HasError() || listValue.IsUnknown() {
		return types.ListNull(elemType)
	}
	return listValue
}

func ParseFromToList[T any](ctx context.Context, items []T, attrName string) types.List {
	tflog.Debug(ctx, "convert.ParseFromToList() "+attrName+" - "+fmt.Sprintf("%v", items))

	// Handle empty list - return null (len(nil) == 0 so nil check is redundant)
	if len(items) == 0 {
		tflog.Debug(ctx, "convert.ParseFromToList() - empty input list, returning null")
		return types.ListNull(FromToObjectType)
	}
	// Process each item into an attr.Value
	fromToValues := make([]attr.Value, 0, len(items))
	for i, item := range items {
		obj := ParseFromTo(ctx, item, attrName)
		if !obj.IsNull() && !obj.IsUnknown() { // Include only non-null/unknown values, adjust as needed
			fromToValues = append(fromToValues, obj)
		} else {
			tflog.Debug(ctx, "convert.ParseFromToList() - skipping null/unknown item at index "+fmt.Sprintf("%d", i))
		}
	}

	// Convert to types.List using ListValueFrom
	listValue, _ := types.ListValueFrom(ctx, FromToObjectType, fromToValues)
	return listValue
}

func ParseFromTo(ctx context.Context, item interface{}, attrName string) types.Object {
	tflog.Debug(ctx, "parseFromTo() "+attrName+" - "+fmt.Sprintf("%v", item))

	if item == nil {
		tflog.Debug(ctx, "parseFromTo() - nil")
		return types.ObjectNull(FromToAttrTypes)
	}

	// Get the reflect.Value of the input
	itemValue := reflect.ValueOf(item)

	// Handle nil or invalid input
	tflog.Debug(ctx, "parseFromTo() itemValue.Kind()- "+itemValue.Kind().String())
	if itemValue.Kind() != reflect.Struct {
		if itemValue.Kind() == reflect.Pointer && !itemValue.IsNil() {
			itemValue = itemValue.Elem()
			// Keep dereferencing until we get to a struct or can't anymore
			for itemValue.Kind() == reflect.Pointer && !itemValue.IsNil() {
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
	if itemValue.Kind() == reflect.Pointer {
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
	tflog.Debug(ctx, "parseFromTo() obj - "+obj.String())
	_ = diagstmp
	return obj
}

func NullableString(value string) types.String {
	if value == "" {
		return types.StringNull()
	}

	return types.StringValue(value)
}

// ParseOptionalString returns StringValue("") if the string is nil, its real value otherwise.
// It is intended to be used with NullStringModifier plan modifier that defaults to empty string.
func ParseOptionalString(s *string) types.String {
	if s == nil {
		return types.StringValue("")
	}
	return types.StringValue(*s)
}
