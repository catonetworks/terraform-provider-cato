package validators

import (
	"context"
	"fmt"
	"net"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	tf "github.com/catonetworks/terraform-provider-cato/internal/provider/tfmodel"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
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
		// check CIDR
		cidr := ipRange.IPRange.ValueString()
		if cidr == "" {
			diags.AddError("Invalid Configuration", "ip_range cannot be empty")
			return
		}
		_, _, err := net.ParseCIDR(cidr)
		if err != nil {
			diags.AddError("Invalid Configuration", fmt.Sprintf("network_network_range '%s' is not a valid CIDR notation", cidr))
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
		if _, exists := uniqueRanges[cidr]; exists {
			diags.AddError("Invalid Configuration", fmt.Sprintf("duplicate ip range '%s'", cidr))
			return
		}
		uniqueRanges[cidr] = struct{}{}
	}
}

func (v GlobalIPRangeValidator) Description(_ context.Context) string {
	return "global ip range settings validation"
}

func (v GlobalIPRangeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
