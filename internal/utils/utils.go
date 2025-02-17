package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ObjectRef struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func ObjectRefSchemaAttr() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Description: "",
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"id": schema.StringAttribute{
			Description: "",
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
	}
}

type ObjectRefOutput struct {
	By    string `json:"By"`
	Input string `json:"Input"`
}

// TransformObjectRefInput is used to transform object {id = "1234"} or {name = "entites"}
// with the following format { by = "ID" input = "1234"} or  { by = "NAME" input = "entities"}.
// this is mandatory to cover difference between Create/Update & Read in the schema
func TransformObjectRefInput(input interface{}) (ObjectRefOutput, error) {
	val := reflect.ValueOf(input)

	if val.Kind() != reflect.Struct {
		return ObjectRefOutput{}, fmt.Errorf("input isn't a type strut")
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := val.Type().Field(i)

		if field.Type() == reflect.TypeOf(types.String{}) {
			terraformString := field.Interface().(types.String)

			if !terraformString.IsNull() && !terraformString.IsUnknown() {
				return ObjectRefOutput{
					By:    strings.ToUpper(fieldType.Name),
					Input: terraformString.ValueString(),
				}, nil
			}
		}
	}

	return ObjectRefOutput{}, fmt.Errorf("No attribute of types.String found")
}

func TransformObjRefToInput(ctx context.Context, fieldValue basetypes.ObjectValue, fieldName string) (*ObjectRefOutput, diag.Diagnostics) {
	var diags diag.Diagnostics

	var objRefInput ObjectRef
	diags.Append(fieldValue.As(ctx, &objRefInput, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil, diags
	}

	objRefOutput, err := TransformObjectRefInput(objRefInput)
	if err != nil {
		diags.AddError(fmt.Sprintf("%s field", fieldName), err.Error())
		return nil, diags
	}

	return &objRefOutput, diags
}

func ToMap(s interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	v := reflect.ValueOf(s)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return result
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		fieldName := field.Name

		result[fieldName] = fieldValue.Interface()
	}

	return result
}

func InterfaceToJSONString(data interface{}) string {
	jsonData, _ := json.Marshal(data)
	return string(jsonData)
}
