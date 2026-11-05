package dynamicip

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/diag"

	sharedpolicy "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policy"
)

// PublishDynamicIPAllocationPolicyRevision publishes the dynamic IP allocation draft when one exists.
// If there is no draft, the API returns FAILURE with PolicyRevisionNotFound — that is treated as success.
func PublishDynamicIPAllocationPolicyRevision(ctx context.Context, client PublishClient, accountID string) diag.Diagnostics {
	var diags diag.Diagnostics
	res, err := client.PolicyDynamicIPAllocationPublishPolicyRevision(ctx, accountID)
	if err != nil {
		diags.AddError("PolicyDynamicIPAllocationPublishPolicyRevision", err.Error())
		return diags
	}
	pub := res.GetPolicy().GetDynamicIPAllocation().GetPublishPolicyRevision()
	if err := sharedpolicy.PublishResult(pub.GetStatus(), pub.GetErrors()); err != nil {
		diags.AddError("PublishDynamicIPAllocationPolicyRevision", err.Error())
	}
	return diags
}

// PublishClient isolates the SDK operation used to publish this policy family.
type PublishClient interface {
	PolicyDynamicIPAllocationPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyDynamicIPAllocationPublishPolicyRevision, error)
}
