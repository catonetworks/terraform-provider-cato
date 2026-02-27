package parse

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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

// prepareIdName prepares the id and name input for the Cato API
// on error it sets the diagnostics error
func PrepareIdName(ctx context.Context, idName types.Object, diags *diag.Diagnostics, fieldName string, optional ...bool) (by cato_models.ObjectRefBy, input string, isSet bool) {
	var tfIdName IdNameRefModel
	if !utils.HasValue(idName) {
		return by, input, false
	}
	if utils.CheckErr(diags, idName.As(ctx, &tfIdName, basetypes.ObjectAsOptions{})) {
		return by, input, false
	}
	if tfIdName.Name.IsUnknown() {
		return by, input, false
	}

	return cato_models.ObjectRefByName, tfIdName.Name.ValueString(), true
}

func PrepareIDRef[T idRefInputs](ctx context.Context, tfObj types.Object, diags *diag.Diagnostics, fieldName string) (sdkRef *T) {
	refBy, refInput, isSet := PrepareIdName(ctx, tfObj, diags, fieldName)
	if !isSet {
		return nil
	}
	return &T{By: refBy, Input: refInput}
}

func ParseIDRef[T idRefTypes](ctx context.Context, ref T, diags *diag.Diagnostics) types.Object {
	type idn struct {
		ID   string `json:"id" tfsdk:"id"`
		Name string `json:"name" tfsdk:"name"`
	}

	// make IdNameRefModel Object
	obj, diag := types.ObjectValueFrom(ctx, IdNameRefModelTypes, idn(ref))
	if utils.CheckErr(diags, diag) {
		return types.ObjectNull(IdNameRefModelTypes)
	}
	return obj
}

func PrepareIDRefList[T idRefInputs](ctx context.Context, tfList types.List, diags *diag.Diagnostics, fieldName string) (sdkList []*T) {
	if !utils.HasValue(tfList) {
		return nil
	}

	for _, idName := range tfList.Elements() {
		refBy, refInput, isSet := PrepareIdName(ctx, idName.(types.Object), diags, fieldName)
		if diags.HasError() {
			return nil
		}
		if isSet {
			sdkList = append(sdkList, &T{By: refBy, Input: refInput})
		}
	}
	return sdkList
}

func ParseIDRefList[T idRefTypes](ctx context.Context, items []*T, diags *diag.Diagnostics) types.List {
	type idn struct{ ID, Name string }

	// null value
	if items == nil {
		return types.ListNull(types.ObjectType{AttrTypes: IdNameRefModelTypes})
	}

	refObjects := make([]attr.Value, 0, len(items))
	for _, i := range items {
		if i == nil {
			continue
		}
		// make IdNameRefModel struct
		val := idn(*i)
		ref := IdNameRefModel{ID: types.StringValue(val.ID), Name: types.StringValue(val.Name)}
		// make IdNameRefModel Object
		obj, diag := types.ObjectValueFrom(ctx, IdNameRefModelTypes, ref)
		if utils.CheckErr(diags, diag) {
			return types.ListNull(types.ObjectType{AttrTypes: IdNameRefModelTypes})
		}
		// append to Object slice
		refObjects = append(refObjects, obj)
	}
	// make List value
	list, diag := types.ListValue(types.ObjectType{AttrTypes: IdNameRefModelTypes}, refObjects)
	diags.Append(diag...)

	return list
}

func PrepareStrings[T ~string](ctx context.Context, tfList types.List, diags *diag.Diagnostics, fieldName string) (sdkList []T) {
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

func ParseStringList[T fmt.Stringer](ctx context.Context, stringers []T, diags *diag.Diagnostics) types.List {
	// null value
	if stringers == nil {
		return types.ListNull(types.StringType)
	}

	// existing empty list
	if len(stringers) == 0 {
		val, diag := types.ListValue(types.StringType, nil)
		diags.Append(diag...)
		return val
	}

	// make []types.String
	stringSlice := make([]types.String, 0, len(stringers))
	for _, o := range stringers {
		stringSlice = append(stringSlice, types.StringValue(o.String()))
	}
	// convert to types.List
	stringList, diag := types.ListValueFrom(ctx, types.StringType, stringSlice)
	diags.Append(diag...)

	return stringList
}

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

func SchemaNameID(prefix string) map[string]schema.Attribute {
	if prefix != "" {
		prefix += " "
	}
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Description: prefix + "name",
			Required:    true,
		},
		"id": schema.StringAttribute{
			Description: prefix + "ID",
			Optional:    false,
			Computed:    true,
		},
	}
}

func Ptr[T any](x T) *T { return &x }
