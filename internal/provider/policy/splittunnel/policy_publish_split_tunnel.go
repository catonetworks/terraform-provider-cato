package splittunnel

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/diag"

	sharedpolicy "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policy"
)

// PublishSplitTunnelPolicyRevision publishes the split tunnel draft when one exists.
// If there is no draft, the API returns FAILURE with PolicyRevisionNotFound — that is treated as success.
func PublishSplitTunnelPolicyRevision(ctx context.Context, client PublishClient, accountID string) diag.Diagnostics {
	var diags diag.Diagnostics
	res, err := client.PolicySplitTunnelPublishPolicyRevision(ctx, accountID)
	if err != nil {
		diags.AddError("PolicySplitTunnelPublishPolicyRevision", err.Error())
		return diags
	}
	pub := res.GetPolicy().GetSplitTunnel().GetPublishPolicyRevision()
	if err := sharedpolicy.PublishResult(pub.GetStatus(), pub.GetErrors()); err != nil {
		diags.AddError("PublishSplitTunnelPolicyRevision", err.Error())
	}
	return diags
}

// PublishClient isolates the SDK operation used to publish this policy family.
type PublishClient interface {
	PolicySplitTunnelPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicySplitTunnelPublishPolicyRevision, error)
}
