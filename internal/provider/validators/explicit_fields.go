package validators

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

func StringFieldIsExplicit(cfgVal, stateVal types.String) bool {
	if !utils.HasValue(cfgVal) {
		return false
	}
	return !utils.HasValue(stateVal) || cfgVal.ValueString() != stateVal.ValueString()
}
