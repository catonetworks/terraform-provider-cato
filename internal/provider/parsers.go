package provider

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type idRefTypes interface {
	~struct {
		ID   string `json:"id" graphql:"id"`
		Name string `json:"name" graphql:"name"`
	}

	// cato_go_sdk.PolicyReadPrivateAccessPolicy_Policy_PrivateAccess_Policy_Rules_Rule_Country |
	// 	cato_go_sdk.PolicyReadPrivateAccessPolicy_Policy_PrivateAccess_Policy_Rules_Rule_Source_UsersGroup |
	// 	cato_go_sdk.PolicyReadPrivateAccessPolicy_Policy_PrivateAccess_Policy_Rules_Rule_Source_User |
	// 	cato_go_sdk.PolicyReadPrivateAccessPolicy_Policy_PrivateAccess_Policy_Rules_Rule_Applications_Application |
	// 	cato_go_sdk.PolicyReadPrivateAccessPolicy_Policy_PrivateAccess_Policy_Rules_Rule_Tracking_Alert_MailingList |
	// 	cato_go_sdk.PolicyReadPrivateAccessPolicy_Policy_PrivateAccess_Policy_Rules_Rule_Tracking_Alert_Webhook |
	// 	cato_go_sdk.PolicyReadPrivateAccessPolicy_Policy_PrivateAccess_Policy_Rules_Rule_Device |
	// 	cato_go_sdk.PolicyReadPrivateAccessPolicy_Policy_PrivateAccess_Policy_Rules_Rule_Tracking_Alert_SubscriptionGroup
}

type idRefInputs interface {
	~struct {
		By    cato_models.ObjectRefBy `json:"by"`
		Input string                  `json:"input"`
	}
}

func IdNameInput(id, name types.String) (isSet bool, by cato_models.ObjectRefBy, input string, err error) {
	if id.IsUnknown() && name.IsUnknown() {
		return false, "", "", nil // not set
	}
	if !id.IsUnknown() && !name.IsUnknown() {
		return false, "", "", fmt.Errorf("Only one of 'id' or 'name' can be specified")
	}

	if !id.IsUnknown() {
		return true, cato_models.ObjectRefByID, id.ValueString(), nil
	}
	return true, cato_models.ObjectRefByName, name.ValueString(), nil
}

// prepareIdName prepares the id and name input for the Cato API
// on error it sets the diagnostics error
func prepareIdName(ctx context.Context, idName types.Object, diags *diag.Diagnostics, fieldName string, optional ...bool) (by cato_models.ObjectRefBy, input string, isSet bool) {
	var tfIdName IdNameRefModel
	if !hasValue(idName) {
		return by, input, false
	}
	if checkErr(diags, idName.As(ctx, &tfIdName, basetypes.ObjectAsOptions{})) {
		return by, input, false
	}

	idNameSet, by, input, err := IdNameInput(tfIdName.ID, tfIdName.Name)
	if err != nil {
		diags.AddError("invalid configuration of "+fieldName, err.Error())
		return
	}

	if idNameSet {
		return by, input, true
	}

	// not set and it is mandatory
	if len(optional) == 0 || (!optional[0]) {
		diags.AddError("missing configuration of "+fieldName, "id or name must be set on "+fieldName)
	}

	return by, input, false
}

func prepareIDRef[T idRefInputs](ctx context.Context, tfObj types.Object, diags *diag.Diagnostics, fieldName string) (sdkRef *T) {
	refBy, refInput, isSet := prepareIdName(ctx, tfObj, diags, fieldName)
	if !isSet {
		return nil
	}
	return &T{By: refBy, Input: refInput}
}

func parseIDRef[T idRefTypes](ctx context.Context, ref T, diags *diag.Diagnostics) types.Object {
	type idn struct{ ID, Name string }

	// make IdNameRefModel Object
	obj, diag := types.ObjectValueFrom(ctx, IdNameRefModelTypes, ref)
	if checkErr(diags, diag) {
		return types.ObjectNull(IdNameRefModelTypes)
	}
	return obj
}

func prepareIDRefList[T idRefInputs](ctx context.Context, tfList types.List, diags *diag.Diagnostics, fieldName string) (sdkList []*T) {
	if !hasValue(tfList) {
		return nil
	}

	for _, idName := range tfList.Elements() {
		refBy, refInput, isSet := prepareIdName(ctx, idName.(types.Object), diags, fieldName)
		if diags.HasError() {
			return nil
		}
		if isSet {
			sdkList = append(sdkList, &T{By: refBy, Input: refInput})
		}
	}
	return sdkList
}

func parseIDRefList[T idRefTypes](ctx context.Context, items []*T, diags *diag.Diagnostics) types.List {
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
		if checkErr(diags, diag) {
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

func prepareStrings[T ~string](ctx context.Context, tfList types.List, diags *diag.Diagnostics, fieldName string) (sdkList []T) {
	if !hasValue(tfList) {
		return nil
	}
	var tfStrings []types.String
	if checkErr(diags, tfList.ElementsAs(ctx, &tfStrings, false)) {
		return nil
	}

	sdkList = make([]T, 0, len(tfStrings))
	for _, s := range tfStrings {
		if hasValue(s) {
			sdkList = append(sdkList, T(s.ValueString()))
		}
	}
	return sdkList
}

func parseStringList[T fmt.Stringer](ctx context.Context, stringers []T, diags *diag.Diagnostics) types.List {
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

type hasValuer interface {
	IsUnknown() bool
	IsNull() bool
}

func hasValue(v hasValuer) bool { return (!v.IsUnknown()) && (!v.IsNull()) }

func ptr[T any](x T) *T { return &x }
