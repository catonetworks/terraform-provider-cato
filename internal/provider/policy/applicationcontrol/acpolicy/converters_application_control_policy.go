package acpolicy

import (
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	types "github.com/hashicorp/terraform-plugin-framework/types"
)

// convertUpdatePolicyValues converts between the Terraform model and SDK values.
func convertUpdatePolicyValues(
	plan ApplicationControlPolicyModel,
) (resultIn cato_models.ApplicationControlPolicyUpdateInput) {
	dc := cato_models.PolicyToggleState(plan.DataControlEnabled.ValueString())
	st := cato_models.PolicyToggleStateDisabled
	if plan.Enabled.ValueBool() {
		st = cato_models.PolicyToggleStateEnabled
	}
	resultIn = cato_models.ApplicationControlPolicyUpdateInput{
		State: &st,
		AdditionalAttributes: &cato_models.ApplicationControlConfigInput{
			DataControlEnabled: dc,
		},
	}
	return resultIn
}

// convertReadStateValues converts between the Terraform model and SDK values.
func convertReadStateValues() (resultOut ApplicationControlPolicyModel) {
	resultOut = ApplicationControlPolicyModel{ID: types.StringValue("application_control")}
	return resultOut
}

// convertReadStateValues2 converts between the Terraform model and SDK values.
func convertReadStateValues2(
	out *ApplicationControlPolicyModel,
	pol *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy,
) {
	out.Enabled = types.BoolValue(pol.Enabled)
	if pol.GetAdditionalAttributes() != nil && pol.GetAdditionalAttributes().GetDataControlEnabled() != nil {
		out.DataControlEnabled = types.StringValue(string(*pol.GetAdditionalAttributes().GetDataControlEnabled()))
	} else {
		out.DataControlEnabled = types.StringNull()
	}
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
