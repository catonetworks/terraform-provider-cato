package globalipranges

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	tf "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/tfmodel"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/validators"
)

func GetGlobalIPRangeValidator() GlobalIPRangeValidator {
	return GlobalIPRangeValidator{}
}

// GlobalIPRangeValidator validates the GlobalIPRange settings
type GlobalIPRangeValidator struct{}

func (v GlobalIPRangeValidator) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	var ipRanges []tf.GlobalIPRange

	if !utils.HasValue(req.ConfigValue) || req.ConfigValue.IsUnknown() {
		return
	}

	// get ip ranges as a slice
	if utils.CheckErr(&resp.Diagnostics, req.ConfigValue.ElementsAs(ctx, &ipRanges, false)) {
		return
	}

	v.ValidateGlobalIPRange(ipRanges, &resp.Diagnostics)
}

func (v GlobalIPRangeValidator) ValidateGlobalIPRange(tfRanges []tf.GlobalIPRange, diags *diag.Diagnostics) {
	uniqueNames := make(map[string]struct{})
	uniqueRanges := make(map[string]struct{})

	for _, ipRange := range tfRanges {
		rangeValue := ipRange.IPRange.ValueString()
		if rangeValue == "" {
			diags.AddError("Invalid Configuration", "ip_range cannot be empty")
			return
		}
		if !validators.IsValidIPRange(rangeValue) {
			diags.AddError("Invalid Configuration", fmt.Sprintf(
				"ip_range '%s' must be a valid IP address, CIDR block, or IP range in start-end format",
				rangeValue,
			))
			return
		}

		// check name
		rangeName := ipRange.Name.ValueString()
		if rangeName == "" {
			diags.AddError("Invalid Configuration", "ip range name cannot be empty")
			return
		}
		if _, exists := uniqueNames[rangeName]; exists {
			diags.AddError("Invalid Configuration", fmt.Sprintf("duplicate ip range name '%s'", rangeName))
			return
		}
		uniqueNames[rangeName] = struct{}{}

		// check for duplicate ranges
		if _, exists := uniqueRanges[rangeValue]; exists {
			diags.AddError("Invalid Configuration", fmt.Sprintf("duplicate ip range '%s'", rangeValue))
			return
		}
		uniqueRanges[rangeValue] = struct{}{}
	}
}

func (v GlobalIPRangeValidator) Description(_ context.Context) string {
	return "global ip range settings validation"
}

func (v GlobalIPRangeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
