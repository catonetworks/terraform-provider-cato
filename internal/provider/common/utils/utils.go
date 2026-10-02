package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ObjectRefOutput struct {
	By    string `json:"By"`
	Input string `json:"Input"`
}

var dateTimeRE = regexp.MustCompile(`(^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})(\.\d+)?`)

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

		if field.Type() == reflect.TypeOf(types.String{}) && strings.EqualFold(fieldType.Name, "ID") {
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

		if field.Type() == reflect.TypeOf(types.String{}) && strings.EqualFold(fieldType.Name, "NAME") {
			terraformString := field.Interface().(types.String)

			if !terraformString.IsNull() && !terraformString.IsUnknown() {
				return ObjectRefOutput{
					By:    "NAME",
					Input: terraformString.ValueString(),
				}, nil
			}
		}
	}

	return ObjectRefOutput{}, fmt.Errorf("no valid Name or ID attribute found")
}

func ToMap(s interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	v := reflect.ValueOf(s)

	if v.Kind() == reflect.Pointer {
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

type HasValuer interface {
	IsUnknown() bool
	IsNull() bool
}

func HasValue(v HasValuer) bool { return (!v.IsUnknown()) && (!v.IsNull()) }

// StringPointerValue handles empty string API values, returns StringNull or StringValue(""), based on the plan
func StringPointerValue(s *string, plan types.String) types.String {
	if (s == nil) || (*s == "") {
		if plan.IsNull() {
			return types.StringNull()
		}
		return types.StringValue("")
	}
	return types.StringPointerValue(s)
}

func NormalizeDateTimePtr(s *string) *string {
	if s == nil {
		return nil
	}
	t := NormalizeDateTime(*s)
	return &t
}

func NormalizeDateTime(s string) string {
	if m := dateTimeRE.FindStringSubmatch(s); m != nil {
		return m[1] + "Z"
	}
	return s
}
