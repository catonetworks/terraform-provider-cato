package parse

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/catonetworks/terraform-provider-cato/internal/utils"
)

// PrepareIDName prepares the id and name input for the Cato API
// on error it sets the diagnostics error
func PrepareIDName(ctx context.Context, idName types.Object, diags *diag.Diagnostics,
) (by cato_models.ObjectRefBy, input string, isSet bool) {
	var tfIDName IDNameRefModel
	if !utils.HasValue(idName) {
		return by, input, false
	}
	if utils.CheckErr(diags, idName.As(ctx, &tfIDName, basetypes.ObjectAsOptions{})) {
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
	obj, valueDiag := types.ObjectValueFrom(ctx, IDNameRefModelTypes, idn(ref))
	if utils.CheckErr(diags, valueDiag) {
		return types.ObjectNull(IDNameRefModelTypes)
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
