package parse

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

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
