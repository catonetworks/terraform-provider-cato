package terminalserver

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/diag"

	sharedpolicy "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policy"
)

// PublishTerminalServerPolicyRevision publishes the terminal server draft when one exists.
// If there is no draft, the API returns FAILURE with PolicyRevisionNotFound — that is treated as success.
func PublishTerminalServerPolicyRevision(ctx context.Context, client PublishClient, accountID string) diag.Diagnostics {
	var diags diag.Diagnostics
	res, err := client.PolicyTerminalServerPublishPolicyRevision(ctx, accountID)
	if err != nil {
		diags.AddError("PolicyTerminalServerPublishPolicyRevision", err.Error())
		return diags
	}
	pub := res.GetPolicy().GetTerminalServer().GetPublishPolicyRevision()
	if err := sharedpolicy.PublishResult(pub.GetStatus(), pub.GetErrors()); err != nil {
		diags.AddError("PublishTerminalServerPolicyRevision", err.Error())
	}
	return diags
}

// PublishClient isolates the SDK operation used to publish this policy family.
type PublishClient interface {
	PolicyTerminalServerPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyTerminalServerPublishPolicyRevision, error)
}
