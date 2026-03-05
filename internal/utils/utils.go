package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ObjectRefOutput struct {
	By    string `json:"By"`
	Input string `json:"Input"`
}

// TransformObjectRefInput is used to transform object {id = "1234"} or {name = "entites"}
// with the following format { by = "ID" input = "1234"} or  { by = "NAME" input = "entities"}.
// this is mandatory to cover difference between Create/Update & Read in the schema
// IMPORTANT: Only id OR name can be submitted to the API, not both. Preference is given to ID for stability.
func TransformObjectRefInput(input interface{}) (ObjectRefOutput, error) {
	val := reflect.ValueOf(input)

	if val.Kind() != reflect.Struct {
		return ObjectRefOutput{}, fmt.Errorf("input isn't a type strut")
	}

	// First pass: look for ID field (preferred for stability)
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := val.Type().Field(i)

		if field.Type() == reflect.TypeOf(types.String{}) && strings.ToUpper(fieldType.Name) == "ID" {
			terraformString := field.Interface().(types.String)

			if !terraformString.IsNull() && !terraformString.IsUnknown() {
				return ObjectRefOutput{
					By:    "ID",
					Input: terraformString.ValueString(),
				}, nil
			}
		}
	}

	// Second pass: look for Name field (fallback only if no valid ID)
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := val.Type().Field(i)

		if field.Type() == reflect.TypeOf(types.String{}) && strings.ToUpper(fieldType.Name) == "NAME" {
			terraformString := field.Interface().(types.String)

			if !terraformString.IsNull() && !terraformString.IsUnknown() {
				return ObjectRefOutput{
					By:    "NAME",
					Input: terraformString.ValueString(),
				}, nil
			}
		}
	}

	return ObjectRefOutput{}, fmt.Errorf("No valid Name or ID attribute found")
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

func ConvertOptionalString(input *string) types.String {
	if input != nil {
		return types.StringValue(*input)
	}
	return types.StringNull()
}

func CheckErr(diags *diag.Diagnostics, in diag.Diagnostics) bool {
	diags.Append(in...)
	return diags.HasError()
}

type hasValuer interface {
	IsUnknown() bool
	IsNull() bool
}

func HasValue(v hasValuer) bool { return (!v.IsUnknown()) && (!v.IsNull()) }
