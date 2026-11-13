package swppolicy

import (
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	types "github.com/hashicorp/terraform-plugin-framework/types"
)

// convertUpdatePolicyValues converts between the Terraform model and SDK values.
func convertUpdatePolicyValues(
	plan SiteWebProxyPolicyModel,
) (resultIn cato_models.SiteWebProxyPolicyUpdateInput) {
	st := cato_models.PolicyToggleStateDisabled
	if plan.Enabled.ValueBool() {
		st = cato_models.PolicyToggleStateEnabled
	}
	resultIn = cato_models.SiteWebProxyPolicyUpdateInput{
		State: &st,
	}
	return resultIn
}

// convertReadStateValues converts between the Terraform model and SDK values.
func convertReadStateValues() (resultOut SiteWebProxyPolicyModel) {
	resultOut = SiteWebProxyPolicyModel{ID: types.StringValue("site_web_proxy")}
	return resultOut
}

// convertReadStateValues2 converts between the Terraform model and SDK values.
func convertReadStateValues2(
	out *SiteWebProxyPolicyModel,
	pol *cato_go_sdk.PolicySiteWebProxyPolicy_Policy_SiteWebProxy_Policy,
) {
	out.Enabled = types.BoolValue(pol.Enabled)
	if pol.GetAudit() != nil {
		out.PublishedTime = types.StringValue(pol.GetAudit().GetPublishedTime())
		out.PublishedBy = types.StringValue(pol.GetAudit().GetPublishedBy())
	}
	if pol.GetRevision() != nil {
		out.RevisionID = types.StringValue(pol.GetRevision().GetID())
		out.RevisionName = types.StringValue(pol.GetRevision().GetName())
	} else {
		out.RevisionID = types.StringNull()
		out.RevisionName = types.StringNull()
	}
}
