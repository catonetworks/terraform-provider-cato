package tenant

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

// PublishAppTenantRestrictionPolicyRevision publishes the app tenant restriction draft when one exists.
// If there is no draft, the API returns FAILURE with PolicyRevisionNotFound — that is treated as success.
func PublishAppTenantRestrictionPolicyRevision(ctx context.Context, c *catoclient.Service) diag.Diagnostics {
	var diags diag.Diagnostics
	res, err := c.Catov2.PolicyAppTenantRestrictionPublishPolicyRevision(ctx, c.AccountId)
	if err != nil {
		diags.AddError("PolicyAppTenantRestrictionPublishPolicyRevision", err.Error())
		return diags
	}
	pub := res.GetPolicy().GetAppTenantRestriction().GetPublishPolicyRevision()
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
			diags.AddError("PolicyAppTenantRestrictionPublishPolicyRevision", msg)
		}
		return diags
	default:
		diags.AddError("PolicyAppTenantRestrictionPublishPolicyRevision", string(*pub.GetStatus()))
		return diags
	}
}

const errPolicyRevisionNotFound = "PolicyRevisionNotFound"
