package license

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const multipleOfTenValidatorDescription = "Ensures the value is a multiple of 10"

type customInt64Validator struct{}

func (v customInt64Validator) Description(_ context.Context) string {
	return multipleOfTenValidatorDescription
}
func (v customInt64Validator) MarkdownDescription(_ context.Context) string {
	return multipleOfTenValidatorDescription
}
func (v customInt64Validator) ValidateInt64(_ context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	val := req.ConfigValue.ValueInt64()
	if val%10 != 0 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid bandwidth, must be in increments of 10",
			fmt.Sprintf("The value %d is not a multiple of 10", val),
		)
	}
}
