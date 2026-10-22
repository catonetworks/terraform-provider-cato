package papolicy

import (
	"context"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	types "github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *privAccessPolicyResource) parseAudit(
	ctx context.Context,
	aud *cato_go_sdk.PolicyReadPrivateAccessPolicy_Policy_PrivateAccess_Policy_Audit,
	diags *diag.Diagnostics,
) types.Object {
	var objectDiags diag.Diagnostics

	// Prepare PolicyAudit object
	auditObj := types.ObjectNull(PolicyAuditTypes)
	if aud != nil {
		tfAudit := PolicyAudit{
			PublishedBy:   types.StringValue(aud.PublishedBy),
			PublishedTime: types.StringValue(aud.PublishedTime),
		}
		auditObj, objectDiags = types.ObjectValueFrom(ctx, PolicyAuditTypes, tfAudit)
		diags.Append(objectDiags...)
		if diags.HasError() {
			return auditObj
		}
	}

	return auditObj
}

// convertHydratePrivAccessPolicyStateValues converts between the Terraform model and SDK values.
func convertHydratePrivAccessPolicyStateValues(
	ctx context.Context,
	r *privAccessPolicyResource,
	originalDiags *diag.Diagnostics,
	result *cato_go_sdk.PolicyReadPrivateAccessPolicy,
) (resultState *PolicyModel) {
	diags := *originalDiags
	defer func() { *originalDiags = diags }()

	policy := result.GetPolicy().GetPrivateAccess().GetPolicy()
	resultState = &PolicyModel{
		Audit:   r.parseAudit(ctx, policy.Audit, &diags),
		Enabled: types.BoolValue(policy.Enabled),
		ID:      types.StringValue("0"), // there is just 1 default policy, the ID is used in automated tests
	}
	return resultState
}

// privateAccessPolicyInput converts the desired enabled flag into the SDK state.
func privateAccessPolicyInput(isEnabled bool) cato_models.PrivateAccessPolicyUpdateInput {
	polEnabled := cato_models.PolicyToggleStateDisabled
	if isEnabled {
		polEnabled = cato_models.PolicyToggleStateEnabled
	}
	return cato_models.PrivateAccessPolicyUpdateInput{State: &polEnabled}
}
