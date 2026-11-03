package sdpstaticip

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	if pub.GetStatus() == nil {
		return diags
	}
	switch *pub.GetStatus() {
	case cato_models.PolicyMutationStatusSuccess:
		return diags
	case cato_models.PolicyMutationStatusFailure:
		for _, e := range pub.GetErrors() {
			if e == nil || e.GetErrorCode() == nil {
				continue
			}
			if *e.GetErrorCode() == errPolicyRevisionNotFound {
				continue
			}
			msg := ""
			if e.GetErrorMessage() != nil {
				msg = *e.GetErrorMessage()
			}
			diags.AddError("PolicySdpStaticIPAllocationPublishPolicyRevision", msg)
		}
		return diags
	default:
		diags.AddError("PolicySdpStaticIPAllocationPublishPolicyRevision", string(*pub.GetStatus()))
		return diags
	}
}

const errPolicyRevisionNotFound = "PolicyRevisionNotFound"

// PublishClient isolates the SDK operation used to publish this policy family.
type PublishClient interface {
	PolicySdpStaticIPAllocationPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySdpStaticIPAllocationPublishPolicyRevision, error)
}
