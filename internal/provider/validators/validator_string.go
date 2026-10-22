package validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type isValider interface {
	~string
	IsValid() bool
}

type stringValidator[T isValider, O ~string] struct {
	name       string
	allOptions []O
}

func StringValidator[T isValider, O ~string](name string, allOptions []O) validator.String {
	return stringValidator[T, O]{
		name:       name,
		allOptions: allOptions,
	}
}

func (v stringValidator[T, O]) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() {
		return
	}
	value := T(req.ConfigValue.ValueString())

	if !value.IsValid() {
		resp.Diagnostics.AddError("Field validation error", fmt.Sprintf("invalid value of %s: '%s'\n - valid options: %+v",
			req.Path.String(), value, v.allOptions))
		return
	}
}

func (v stringValidator[T, O]) Description(_ context.Context) string {
	return fmt.Sprintf("%s must be one of: %v", v.name, v.allOptions)
}
func (v stringValidator[T, O]) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
