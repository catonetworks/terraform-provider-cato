package planmodifiers

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

// DefaultPlanValue returns the appropriate plan value: cfg -> state -> unknown
func DefaultPlanValue(cfgValue, stateValue types.String, stateDefined bool) (planValue types.String) {
	newValue := types.StringUnknown()
	if utils.HasValue(cfgValue) {
		newValue = cfgValue
	} else if stateDefined && utils.HasValue(stateValue) {
		newValue = stateValue
	}
	return newValue
}
