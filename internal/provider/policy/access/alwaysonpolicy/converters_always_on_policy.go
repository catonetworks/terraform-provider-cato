package alwaysonpolicy

import (
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

func buildPolicyInput(plan PolicyModel, diags *diag.Diagnostics) cato_models.ZtnaAlwaysOnPolicyUpdateInput {
	if !utils.HasValue(plan.Enabled) {
		diags.AddError("Invalid Always On policy", "enabled must be a known boolean before applying the resource.")
		return cato_models.ZtnaAlwaysOnPolicyUpdateInput{}
	}
	state := cato_models.PolicyToggleStateDisabled
	if plan.Enabled.ValueBool() {
		state = cato_models.PolicyToggleStateEnabled
	}
	return cato_models.ZtnaAlwaysOnPolicyUpdateInput{State: &state}
}

func parsePolicyUpdate(p *cato_go_sdk.PolicyZtnaAlwaysOnUpdatePolicy_Policy_ZtnaAlwaysOn_UpdatePolicy) *PolicyModel {
	policy := p.GetPolicy()
	return &PolicyModel{
		Enabled: types.BoolValue(policy.GetEnabled()),
	}
}

func parsePolicy(p *cato_go_sdk.PolicyZtnaAlwaysOnPolicy) *PolicyModel {
	policy := p.GetPolicy().GetZtnaAlwaysOn().GetPolicy()
	return &PolicyModel{
		Enabled: types.BoolValue(policy.GetEnabled()),
	}
}
