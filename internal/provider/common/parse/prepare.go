package parse

/*
The "prepare" functions convert terraform type system variables (types.*) into
plain go types used by API SDK (cato_models.*)
*/

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/common/apperr"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/common/idname"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/common/utils"
)

// PrepareIDName prepares the id and name input for the Cato API
// on error it sets the diagnostics error
func PrepareIDName(ctx context.Context, idName types.Object, diags *diag.Diagnostics,
) (by cato_models.ObjectRefBy, input string, isSet bool) {
	var tfIDName idname.RefModel
	if !utils.HasValue(idName) {
		return by, input, false
	}
	if apperr.CheckErr(diags, idName.As(ctx, &tfIDName, basetypes.ObjectAsOptions{})) {
		return by, input, false
	}

	// ref by ID
	if !tfIDName.ID.IsUnknown() {
		return cato_models.ObjectRefByID, tfIDName.ID.ValueString(), true
	}

	// ref by Name
	if !tfIDName.Name.IsUnknown() {
		return cato_models.ObjectRefByName, tfIDName.Name.ValueString(), true
	}

	return by, input, false
}

func PrepareIDRef[T idRefInputs](ctx context.Context, tfObj types.Object, diags *diag.Diagnostics) (sdkRef *T) {
	refBy, refInput, isSet := PrepareIDName(ctx, tfObj, diags)
	if !isSet {
		return nil
	}
	return &T{By: refBy, Input: refInput}
}

// IDRef parses the ID reference object for the given type T and returns a Terraform Object value
func IDRef[T idRefTypes](ctx context.Context, ref T, diags *diag.Diagnostics) types.Object {
	type idn struct {
		ID   string `json:"id" tfsdk:"id"`
		Name string `json:"name" tfsdk:"name"`
	}

	// make IDNameRefModel Object
	obj, valueDiag := types.ObjectValueFrom(ctx, idname.RefModelTypes, idn(ref))
	if apperr.CheckErr(diags, valueDiag) {
		return types.ObjectNull(idname.RefModelTypes)
	}
	return obj
}

func PrepareIDRefSet[T idRefInputs](ctx context.Context, tfSet types.Set, diags *diag.Diagnostics) (sdkList []*T) {
	if !utils.HasValue(tfSet) {
		return nil
	}

	for _, idName := range tfSet.Elements() {
		refBy, refInput, isSet := PrepareIDName(ctx, idName.(types.Object), diags)
		if diags.HasError() {
			return nil
		}
		if isSet {
			sdkList = append(sdkList, &T{By: refBy, Input: refInput})
		}
	}
	return sdkList
}

// StringPointerForOptionalInput returns nil for NULL/Unknown/empty-string, pointer otherwise
func StringPointerForOptionalInput(value types.String) *string {
	if value.IsNull() || value.IsUnknown() || value.ValueString() == "" {
		return nil
	}
	return value.ValueStringPointer()
}

func PrepareStrings[T ~string](ctx context.Context, tfSet types.Set, diags *diag.Diagnostics) (sdkList []T) {
	if !utils.HasValue(tfSet) {
		return nil
	}
	var tfStrings []types.String
	if apperr.CheckErr(diags, tfSet.ElementsAs(ctx, &tfStrings, false)) {
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
	if apperr.CheckErr(diags, tfList.ElementsAs(ctx, &tfStrings, false)) {
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
	if apperr.CheckErr(diags, tfList.ElementsAs(ctx, &tfInts, false)) {
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
