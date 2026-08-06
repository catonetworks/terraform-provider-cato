package parse

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/utils"
)

type idRefTypes interface {
	~struct {
		ID   string `json:"id" graphql:"id"`
		Name string `json:"name" graphql:"name"`
	}
}

type idRefInputs interface {
	~struct {
		By    cato_models.ObjectRefBy `json:"by"`
		Input string                  `json:"input"`
	}
}

// IDRefSet parses a set of ID reference objects for the given type T and returns a types.Set of Terraform Object values
func IDRefSet[T idRefTypes](ctx context.Context, items []*T, diags *diag.Diagnostics) types.Set {
	type idn struct{ ID, Name string }

	// null value
	if items == nil {
		return types.SetNull(types.ObjectType{AttrTypes: IDNameRefModelTypes})
	}

	refObjects := make([]attr.Value, 0, len(items))
	for _, i := range items {
		if i == nil {
			continue
		}
		// make IDNameRefModel struct
		val := idn(*i)
		ref := IDNameRefModel{ID: types.StringValue(val.ID), Name: types.StringValue(val.Name)}
		// make IDNameRefModel Object
		obj, valueDiag := types.ObjectValueFrom(ctx, IDNameRefModelTypes, ref)
		if utils.CheckErr(diags, valueDiag) {
			return types.SetNull(types.ObjectType{AttrTypes: IDNameRefModelTypes})
		}
		// append to Object slice
		refObjects = append(refObjects, obj)
	}
	// make Set value
	setValues, valueDiag := types.SetValue(types.ObjectType{AttrTypes: IDNameRefModelTypes}, refObjects)
	diags.Append(valueDiag...)

	return setValues
}

// IDRefList parses a slice of ID reference objects for the given type T and returns a Terraform List value
func IDRefList[T idRefTypes](ctx context.Context, items []*T, diags *diag.Diagnostics) types.List {
	type idn struct{ ID, Name string }

	// null value
	if items == nil {
		return types.ListNull(types.ObjectType{AttrTypes: IDNameRefModelTypes})
	}

	refObjects := make([]attr.Value, 0, len(items))
	for _, i := range items {
		if i == nil {
			continue
		}
		// make IDNameRefModel struct
		val := idn(*i)
		ref := IDNameRefModel{ID: types.StringValue(val.ID), Name: types.StringValue(val.Name)}
		// make IDNameRefModel Object
		obj, valueDiag := types.ObjectValueFrom(ctx, IDNameRefModelTypes, ref)
		if utils.CheckErr(diags, valueDiag) {
			return types.ListNull(types.ObjectType{AttrTypes: IDNameRefModelTypes})
		}
		// append to Object slice
		refObjects = append(refObjects, obj)
	}
	// make List value
	list, valueDiag := types.ListValue(types.ObjectType{AttrTypes: IDNameRefModelTypes}, refObjects)
	diags.Append(valueDiag...)

	return list
}

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

// StringSetFunc parses a slice of convertable items into a types.Set of strings
func StringSetFunc[T any](ctx context.Context, stringers []T, convert func(x T) string, diags *diag.Diagnostics,
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
	return StringSetFunc(ctx, stringers, func(s T) string { return s.String() }, diags)
}

// StringListFunc parses a slice of any type that can be converted to string,
// returns TF list of StringType
func StringListFunc[T any](ctx context.Context, stringers []T, convert func(x T) string, diags *diag.Diagnostics) types.List {
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

// StringerList parses a slice of fmt.Stringer into a types.List of strings
func StringerList(ctx context.Context, stringers []fmt.Stringer, diags *diag.Diagnostics) types.List {
	return StringListFunc(ctx, stringers,
		func(s fmt.Stringer) string { return s.String() },
		diags)
}

// StringList parses a slice of fmt.Stringer into a types.List of strings
func StringList[T ~string](ctx context.Context, stringers []T, diags *diag.Diagnostics) types.List {
	return StringListFunc(ctx, stringers,
		func(s T) string { return string(s) },
		diags)
}

type int64er interface{ GetInt64() int64 }

// Int64List parses a slice of int64er into a types.List of int64
func Int64List[T int64er](ctx context.Context, ints []T, diags *diag.Diagnostics) types.List {
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

// Int64SetFunc parses a slice of convertable items into a types.Set of int64
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
