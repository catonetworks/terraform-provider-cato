package convert

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

func PrepareStrings[T ~string](ctx context.Context, tfSet types.Set, diags *diag.Diagnostics) (sdkList []T) {
	if !utils.HasValue(tfSet) {
		return nil
	}
	var tfStrings []types.String
	if utils.CheckErr(diags, tfSet.ElementsAs(ctx, &tfStrings, false)) {
		return nil
	}

	sdkList = make([]T, 0, len(tfStrings))
	for _, s := range tfStrings {
		if utils.HasValue(s) {
			sdkList = append(sdkList, T(s.ValueString()))
		}
	}
	return sdkList
}

func PrepareStringList[T ~string](ctx context.Context, tfList types.List, diags *diag.Diagnostics) (sdkList []T) {
	if !utils.HasValue(tfList) {
		return nil
	}
	var tfStrings []types.String
	if utils.CheckErr(diags, tfList.ElementsAs(ctx, &tfStrings, false)) {
		return nil
	}

	sdkList = make([]T, 0, len(tfStrings))
	for _, s := range tfStrings {
		if utils.HasValue(s) {
			sdkList = append(sdkList, T(s.ValueString()))
		}
	}
	return sdkList
}

func PrepareInt64List[T ~int64](ctx context.Context, tfList types.List, diags *diag.Diagnostics) (sdkList []T) {
	if !utils.HasValue(tfList) {
		return nil
	}
	var tfInts []types.Int64
	if utils.CheckErr(diags, tfList.ElementsAs(ctx, &tfInts, false)) {
		return nil
	}

	sdkList = make([]T, 0, len(tfInts))
	for _, s := range tfInts {
		if utils.HasValue(s) {
			sdkList = append(sdkList, T(s.ValueInt64()))
		}
	}
	return sdkList
}

// Int64SetFunc parses a slice of convertible items into a types.Set of int64
func Int64SetFunc[T any](ctx context.Context, ints []T, convert func(x T) int64, diags *diag.Diagnostics,
) types.Set {
	// null value
	if ints == nil {
		return types.SetNull(types.Int64Type)
	}

	// existing empty list
	if len(ints) == 0 {
		val, valueDiag := types.SetValue(types.Int64Type, nil)
		diags.Append(valueDiag...)
		return val
	}

	// make []types.Int64
	intSlice := make([]types.Int64, 0, len(ints))
	for _, o := range ints {
		intSlice = append(intSlice, types.Int64Value(convert(o)))
	}
	// convert to types.Set
	intSet, valueDiag := types.SetValueFrom(ctx, types.Int64Type, intSlice)
	diags.Append(valueDiag...)

	return intSet
}

// KnownStringPointer returns a pointer to the known string value, nil for a null or unknown value.
func KnownStringPointer(s types.String) *string {
	if s.IsUnknown() {
		return nil
	}
	return s.ValueStringPointer()
}

func KnownInt64Pointer(s types.Int64) *int64 {
	if s.IsUnknown() {
		return nil
	}
	return s.ValueInt64Pointer()
}
func KnownBoolPointer(s types.Bool) *bool {
	if s.IsUnknown() {
		return nil
	}
	return s.ValueBoolPointer()
}
