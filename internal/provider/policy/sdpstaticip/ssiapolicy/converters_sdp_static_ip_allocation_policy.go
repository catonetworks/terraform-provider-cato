package ssiapolicy

import (
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	types "github.com/hashicorp/terraform-plugin-framework/types"
)

// convertUpdatePolicyValues converts between the Terraform model and SDK values.
func convertUpdatePolicyValues(
	plan SdpStaticIPAllocationPolicyModel,
) (resultIn cato_models.SdpStaticIPAllocationPolicyUpdateInput) {
	st := cato_models.PolicyToggleStateDisabled
	if plan.Enabled.ValueBool() {
		st = cato_models.PolicyToggleStateEnabled
	}
	resultIn = cato_models.SdpStaticIPAllocationPolicyUpdateInput{
		State: &st,
	}
	return resultIn
}

// convertReadStateValues converts between the Terraform model and SDK values.
func convertReadStateValues() (resultOut SdpStaticIPAllocationPolicyModel) {
	resultOut = SdpStaticIPAllocationPolicyModel{ID: types.StringValue("sdp_static_ip_allocation")}
	return resultOut
}

// convertReadStateValues2 converts between the Terraform model and SDK values.
func convertReadStateValues2(
	out *SdpStaticIPAllocationPolicyModel,
	pol *cato_go_sdk.PolicySdpStaticIpAllocationPolicy_Policy_SdpStaticIPAllocation_Policy,
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
