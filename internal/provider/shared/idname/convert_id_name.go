package idname

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
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

// ParseIDName parses the ID reference object for the given type T and returns a Terraform Object value
func ParseIDName[T idRefTypes](ctx context.Context, ref T, diags *diag.Diagnostics) types.Object {
	type idn struct {
		ID   string `json:"id" tfsdk:"id"`
		Name string `json:"name" tfsdk:"name"`
	}

	// make IDNameRefModel Object
	obj, valueDiag := types.ObjectValueFrom(ctx, ModelTypes, idn(ref))
	if utils.CheckErr(diags, valueDiag) {
		return types.ObjectNull(ModelTypes)
	}
	return obj
}

// ParseIDNameSet parses a set of ID reference objects for the given type T and returns a types.Set of Terraform Object values
func ParseIDNameSet[T idRefTypes](ctx context.Context, items []*T, diags *diag.Diagnostics) types.Set {
	type idn struct{ ID, Name string }

	// null value
	if items == nil {
		return types.SetNull(types.ObjectType{AttrTypes: ModelTypes})
	}

	refObjects := make([]attr.Value, 0, len(items))
	for _, i := range items {
		if i == nil {
			continue
		}
		// make IDNameRefModel struct
		val := idn(*i)
		ref := Model{ID: types.StringValue(val.ID), Name: types.StringValue(val.Name)}
		// make IDNameRefModel Object
		obj, valueDiag := types.ObjectValueFrom(ctx, ModelTypes, ref)
		if utils.CheckErr(diags, valueDiag) {
			return types.SetNull(types.ObjectType{AttrTypes: ModelTypes})
		}
		// append to Object slice
		refObjects = append(refObjects, obj)
	}
	// make Set value
	setValues, valueDiag := types.SetValue(types.ObjectType{AttrTypes: ModelTypes}, refObjects)
	diags.Append(valueDiag...)

	return setValues
}

// ParseIDNameList parses a slice of ID reference objects for the given type T and returns a Terraform List value
func ParseIDNameList[T idRefTypes](ctx context.Context, items []*T, diags *diag.Diagnostics) types.List {
	type idn struct{ ID, Name string }

	// null value
	if items == nil {
		return types.ListNull(types.ObjectType{AttrTypes: ModelTypes})
	}

	refObjects := make([]attr.Value, 0, len(items))
	for _, i := range items {
		if i == nil {
			continue
		}
		// make IDNameRefModel struct
		val := idn(*i)
		ref := Model{ID: types.StringValue(val.ID), Name: types.StringValue(val.Name)}
		// make IDNameRefModel Object
		obj, valueDiag := types.ObjectValueFrom(ctx, ModelTypes, ref)
		if utils.CheckErr(diags, valueDiag) {
			return types.ListNull(types.ObjectType{AttrTypes: ModelTypes})
		}
		// append to Object slice
		refObjects = append(refObjects, obj)
	}
	// make List value
	list, valueDiag := types.ListValue(types.ObjectType{AttrTypes: ModelTypes}, refObjects)
	diags.Append(valueDiag...)

	return list
}

func PrepareIDName[T idRefInputs](ctx context.Context, tfObj types.Object, diags *diag.Diagnostics) (sdkRef *T) {
	refBy, refInput, isSet := prepareIDName(ctx, tfObj, diags)
	if !isSet {
		return nil
	}
	return &T{By: refBy, Input: refInput}
}

func PrepareIDNameSet[T idRefInputs](ctx context.Context, tfSet types.Set, diags *diag.Diagnostics) (sdkList []*T) {
	if !utils.HasValue(tfSet) {
		return nil
	}

	for _, idName := range tfSet.Elements() {
		refBy, refInput, isSet := prepareIDName(ctx, idName.(types.Object), diags)
		if diags.HasError() {
			return nil
		}
		if isSet {
			sdkList = append(sdkList, &T{By: refBy, Input: refInput})
		}
	}
	return sdkList
}

// prepareIDName prepares the id and name input for the Cato API
// on error it sets the diagnostics error
func prepareIDName(ctx context.Context, idName types.Object, diags *diag.Diagnostics,
) (by cato_models.ObjectRefBy, input string, isSet bool) {
	var tfIDName Model
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
