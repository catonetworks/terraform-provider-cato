package convert

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type int64er interface{ GetInt64() int64 }

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
