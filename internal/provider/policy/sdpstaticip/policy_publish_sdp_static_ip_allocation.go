package sdpstaticip

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/diag"

	sharedpolicy "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policy"
)

// PublishSdpStaticIPAllocationPolicyRevision publishes the SDP static IP allocation draft when one exists.
// If there is no draft, the API returns FAILURE with PolicyRevisionNotFound — that is treated as success.
func PublishSdpStaticIPAllocationPolicyRevision(ctx context.Context, client PublishClient, accountID string) diag.Diagnostics {
	var diags diag.Diagnostics
	res, err := client.PolicySdpStaticIPAllocationPublishPolicyRevision(ctx, accountID)
	if err != nil {
		diags.AddError("PolicySdpStaticIPAllocationPublishPolicyRevision", err.Error())
		return diags
	}
	pub := res.GetPolicy().GetSdpStaticIPAllocation().GetPublishPolicyRevision()
	if err := sharedpolicy.PublishResult(pub.GetStatus(), pub.GetErrors()); err != nil {
		diags.AddError("PublishSdpStaticIPAllocationPolicyRevision", err.Error())
	}
	return diags
}

// PublishClient isolates the SDK operation used to publish this policy family.
type PublishClient interface {
	PolicySdpStaticIPAllocationPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySdpStaticIPAllocationPublishPolicyRevision, error)
}
