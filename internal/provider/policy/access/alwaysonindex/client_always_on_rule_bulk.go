package alwaysonindex

import (
	"context"

	"github.com/Yamashou/gqlgenc/clientv2"
	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// ResourceClient contains the SDK operations used for bulk Always On ordering.
type ResourceClient interface {
	PolicyZtnaAlwaysOnPolicy(ctx context.Context, accountID string, input *cato_models.ZtnaAlwaysOnPolicyInput,
		interceptors ...clientv2.RequestInterceptor) (*cato.PolicyZtnaAlwaysOnPolicy, error)
	PolicyZtnaAlwaysOnMoveSection(ctx context.Context, accountID string,
		policyInput *cato_models.ZtnaAlwaysOnPolicyMutationInput, input cato_models.PolicyMoveSectionInput,
		interceptors ...clientv2.RequestInterceptor) (*cato.PolicyZtnaAlwaysOnMoveSection, error)
	PolicyZtnaAlwaysOnMoveRule(ctx context.Context, accountID string,
		policyInput *cato_models.ZtnaAlwaysOnPolicyMutationInput, input cato_models.PolicyMoveRuleInput,
		interceptors ...clientv2.RequestInterceptor) (*cato.PolicyZtnaAlwaysOnMoveRule, error)
	PolicyZtnaAlwaysOnPublishPolicyRevision(ctx context.Context, accountID string,
		policyInput *cato_models.ZtnaAlwaysOnPolicyMutationInput, input *cato_models.PolicyPublishRevisionInput,
		interceptors ...clientv2.RequestInterceptor) (*cato.PolicyZtnaAlwaysOnPublishPolicyRevision, error)
}

func (r *alwaysOnRuleBulkResource) readPolicy(ctx context.Context, diags *diag.Diagnostics) *RulesIndex {
	if r.client == nil || r.client.AccountId == "" || r.getClient() == nil {
		diags.AddError("Always On ordering is not configured", "An API client and account ID are required")
		return nil
	}
	result, err := r.getClient().PolicyZtnaAlwaysOnPolicy(ctx, r.client.AccountId, nil)
	if err != nil {
		diags.AddError("Failed to read Always On policy ordering", err.Error())
		return nil
	}
	return parseRulesIndex(ctx, result, diags)
}
