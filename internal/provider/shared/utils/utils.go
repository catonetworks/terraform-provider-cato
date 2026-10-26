package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/apperror"
)

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

func CheckErr(diags *diag.Diagnostics, in diag.Diagnostics) bool {
	diags.Append(in...)
	return diags.HasError()
}

type HasValuer interface {
	IsUnknown() bool
	IsNull() bool
}

func HasValue(v HasValuer) bool { return (!v.IsUnknown()) && (!v.IsNull()) }

func CheckAPIErrors[T apperror.APIErrors](err error, errors []T, summary string, diags *diag.Diagnostics) bool {
	if err != nil {
		diags.AddError(summary, err.Error())
		return true
	}
	if len(errors) > 0 {
		for _, e := range errors {
			if msg := e.GetErrorMessage(); msg != nil {
				diags.AddError(summary, *msg)
			} else if code := e.GetErrorCode(); code != nil {
				diags.AddError(summary, *code)
			} else {
				diags.AddError(summary, "API mutation failed without error details")
			}
		}
		return true
	}
	return false
}

var dateTimeRE = regexp.MustCompile(`(^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})(\.\d+)?`)

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

func ParseTimeStringWithTZ(timeStr string) (string, error) {
	t, err := ParseFlexibleTimeString(timeStr)
	if err != nil {
		return "", err
	}
	return t.UTC().Format("2006-01-02T15:04:05+00:00"), nil
}

func ParseFlexibleTimeString(timeStr string) (time.Time, error) {
	t, err := time.Parse("2006-01-02T15:04:05", timeStr)
	if err != nil {
		t, err = time.Parse(time.RFC3339, timeStr)
		if err != nil {
			return time.Time{}, err
		}
	}
	return t, nil
}

func ParseTimeString(timeStr string, preserveWith ...string) (string, error) {
	t, err := ParseFlexibleTimeString(timeStr)
	if err != nil {
		return "", err
	}

	if len(preserveWith) > 0 && preserveWith[0] != "" {
		preservedTime, err := ParseFlexibleTimeString(preserveWith[0])
		if err == nil && preservedTime.UTC().Equal(t.UTC()) {
			return preserveWith[0], nil
		}
	}

	return t.UTC().Format(time.RFC3339), nil
}

func GqlOptionalStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func Contains(nameToIDMap map[string]struct{}, name string) bool {
	_, exists := nameToIDMap[name]
	return exists
}

// NormalizePolicyTimestampString normalizes API datetime strings (e.g. ...00.000) to RFC3339-style
// values that match typical Terraform configuration (…00Z).
func NormalizePolicyTimestampString(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasSuffix(s, ".000") {
		s = strings.TrimSuffix(s, ".000")
		if !strings.HasSuffix(s, "Z") {
			s += "Z"
		}
	}
	return s
}

// NormalizePolicyRecurringClock maps API values like "08:00:00" to "08:00" so state
// matches common Terraform configuration and avoids inconsistent-result errors.
func NormalizePolicyRecurringClock(s string) string {
	s = strings.TrimSpace(s)
	if len(s) == 8 && strings.Count(s, ":") == 2 && strings.HasSuffix(s, ":00") {
		return s[:5]
	}
	return s
}

// ToStrings returns string slice
func ToStrings[T ~string](enums []T) []string {
	result := make([]string, len(enums))
	for i, enum := range enums {
		result[i] = string(enum)
	}
	return result
}
