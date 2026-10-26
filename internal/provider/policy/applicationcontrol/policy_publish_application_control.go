package applicationcontrol

import (
	"context"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

const errPolicyRevisionNotFound = "PolicyRevisionNotFound"

// PublishApplicationControlPolicyRevision publishes the Application Control draft when one exists.
// If there is no draft, the API returns FAILURE with PolicyRevisionNotFound — that is treated as success.
func PublishApplicationControlPolicyRevision(ctx context.Context, client PublishClient, accountID string) diag.Diagnostics {
	var diags diag.Diagnostics
	res, err := client.PolicyApplicationControlPublishPolicyRevision(ctx, accountID)
	if err != nil {
		diags.AddError("PolicyApplicationControlPublishPolicyRevision", err.Error())
		return diags
	}
	pub := res.GetPolicy().GetApplicationControl().GetPublishPolicyRevision()
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
			diags.AddError("PolicyApplicationControlPublishPolicyRevision", msg)
		}
		return diags
	default:
		diags.AddError("PolicyApplicationControlPublishPolicyRevision", string(*pub.GetStatus()))
		return diags
	}
}

// PublishClient isolates the SDK operation used to publish this policy family.
type PublishClient interface {
	PolicyApplicationControlPublishPolicyRevision(
		ctx context.Context,
		accountID string,
		interceptors ...clientv2.RequestInterceptor,
	) (*cato_go_sdk.PolicyApplicationControlPublishPolicyRevision, error)
}
