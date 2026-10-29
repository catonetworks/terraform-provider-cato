package cloudinterconnectphysicalconnection

import (
	"context"
	"strings"

	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/cato-go-sdk/scalars"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

// Computed allocated tags are not user requests for the privileged custom-tag API.
func buildConfiguredCreate(ctx context.Context, cfg tfsdk.Config, p resourceModel, d *diag.Diagnostics,
) m.AddCloudInterconnectPhysicalConnectionInput {
	var vlan types.Int64
	qinq := types.ObjectNull(qinqAttrTypes)
	d.Append(cfg.GetAttribute(ctx, path.Root("vlan"), &vlan)...)
	d.Append(cfg.GetAttribute(ctx, path.Root("qinq_vlan_configuration"), &qinq)...)
	if vlan.IsNull() {
		p.Vlan = types.Int64Null()
	}
	if qinq.IsNull() {
		p.QinqVlanConfiguration = types.ObjectNull(qinqAttrTypes)
	}
	return buildCreate(ctx, p, d)
}
func buildCreate(ctx context.Context, plan resourceModel, diags *diag.Diagnostics) m.AddCloudInterconnectPhysicalConnectionInput {
	validateFields(plan, diags)
	input := m.AddCloudInterconnectPhysicalConnectionInput{
		HaRole: m.HaRole(plan.HARole.ValueString()), EncapsulationMethod: m.TaggingMethod(plan.EncapsulationMethod.ValueString()),
		Site:                idname.PrepareIDName[m.SiteRefInput](ctx, plan.Site, diags),
		PopLocation:         idname.PrepareIDName[m.PopLocationRefInput](ctx, plan.PopLocation, diags),
		ServiceProviderName: plan.ServiceProviderName.ValueString(), Subnet: plan.Subnet.ValueString(),
		PrivateCatoIP: plan.PrivateCatoIP.ValueString(), PrivateSiteIP: plan.PrivateSiteIP.ValueString(),
		UpstreamBwLimit: plan.UpstreamBwLimit.ValueString(), DownstreamBwLimit: plan.DownstreamBwLimit.ValueString(),
	}
	if input.Site == nil || input.PopLocation == nil {
		diags.AddError("Missing reference", "Known non-null site and PoP references are required.")
	}
	if !plan.Vlan.IsNull() && !plan.Vlan.IsUnknown() {
		vlan := scalars.Vlan(plan.Vlan.ValueInt64())
		input.Vlan = &vlan
	}
	if !plan.QinqVlanConfiguration.IsNull() && !plan.QinqVlanConfiguration.IsUnknown() {
		attrs := plan.QinqVlanConfiguration.Attributes()
		input.QinQVlanConfiguration = &m.QinQVlanConfiguration{
			CVlan: scalars.Vlan(attrs["c_vlan"].(types.Int64).ValueInt64()), SVlan: scalars.Vlan(attrs["s_vlan"].(types.Int64).ValueInt64()),
		}
	}
	return input
}
func buildUpdate(input m.AddCloudInterconnectPhysicalConnectionInput, id string) m.UpdateCloudInterconnectPhysicalConnectionInput {
	return m.UpdateCloudInterconnectPhysicalConnectionInput{ID: id, EncapsulationMethod: &input.EncapsulationMethod,
		PopLocation: input.PopLocation, ServiceProviderName: &input.ServiceProviderName, Subnet: &input.Subnet,
		PrivateCatoIP: &input.PrivateCatoIP, PrivateSiteIP: &input.PrivateSiteIP,
		UpstreamBwLimit: &input.UpstreamBwLimit, DownstreamBwLimit: &input.DownstreamBwLimit,
		Vlan: input.Vlan, QinQVlanConfiguration: input.QinQVlanConfiguration}
}

func identityState(plan resourceModel, id string) resourceModel {
	plan.ID = types.StringValue(id)
	if plan.Vlan.IsUnknown() {
		plan.Vlan = types.Int64Null()
	}
	if plan.QinqVlanConfiguration.IsUnknown() {
		plan.QinqVlanConfiguration = types.ObjectNull(qinqAttrTypes)
	}
	plan.Site = knownReference(plan.Site)
	plan.PopLocation = knownReference(plan.PopLocation)
	return plan
}
func knownReference(value types.Object) types.Object {
	attrs := value.Attributes()
	for key, item := range attrs {
		if item.IsUnknown() {
			attrs[key] = types.StringValue("")
		}
	}
	return types.ObjectValueMust(idname.RefAttrTypes, attrs)
}
func parseConnection(ctx context.Context, item *connectionRecord, prior types.String, diags *diag.Diagnostics) resourceModel {
	state := resourceModel{ID: types.StringValue(item.GetID()), Site: idname.ParseIDName(ctx, *item.GetSite(), diags),
		PopLocation: idname.ParseIDName(ctx, *item.GetPopLocation(), diags),
		HARole:      types.StringValue(string(*item.GetHaRole())), EncapsulationMethod: types.StringValue(string(*item.GetEncapsulationMethod())),
		ServiceProviderName: types.StringValue(item.GetServiceProviderName()), Subnet: types.StringValue(item.GetSubnet()),
		PrivateCatoIP: types.StringValue(item.GetPrivateCatoIP()), PrivateSiteIP: types.StringValue(item.GetPrivateSiteIP()),
		UpstreamBwLimit: types.StringValue(item.GetUpstreamBwLimit()), DownstreamBwLimit: types.StringValue(item.GetDownstreamBwLimit()),
		Vlan: types.Int64Null(), QinqVlanConfiguration: types.ObjectNull(qinqAttrTypes),
	}
	if !prior.IsNull() && !prior.IsUnknown() && strings.EqualFold(prior.ValueString(), item.GetServiceProviderName()) {
		state.ServiceProviderName = prior
	}
	// The API can retain inactive tagging fields. Only the selected encapsulation owns their Terraform state.
	if state.EncapsulationMethod.ValueString() == "DOT1Q" && item.GetVlan() != nil {
		state.Vlan = types.Int64Value(int64(*item.GetVlan()))
	}
	if state.EncapsulationMethod.ValueString() == "QINQ" && item.GetCVlan() != nil && item.GetSVlan() != nil {
		object, d := types.ObjectValue(qinqAttrTypes, map[string]attr.Value{
			"c_vlan": types.Int64Value(int64(*item.GetCVlan())), "s_vlan": types.Int64Value(int64(*item.GetSVlan())),
		})
		diags.Append(d...)
		state.QinqVlanConfiguration = object
	}
	buildCreate(ctx, state, diags)
	if state.EncapsulationMethod.ValueString() == "DOT1Q" && !validVLAN(state.Vlan) {
		diags.AddError("Invalid allocated VLAN", "API did not return a valid DOT1Q VLAN.")
	}
	if state.EncapsulationMethod.ValueString() == "QINQ" && state.QinqVlanConfiguration.IsNull() {
		diags.AddError("Invalid allocated VLANs", "API did not return both QINQ VLANs.")
	}
	return state
}
func matchesReference(by m.ObjectRefBy, input, id, name string) bool {
	if by == m.ObjectRefByID {
		return input == id
	}
	return by == m.ObjectRefByName && input == name
}
func matchesConnection(item *connectionRecord, input m.AddCloudInterconnectPhysicalConnectionInput) bool {
	if *item.GetHaRole() != input.HaRole || *item.GetEncapsulationMethod() != input.EncapsulationMethod {
		return false
	}
	if !matchesReference(input.Site.By, input.Site.Input, item.GetSite().GetID(), item.GetSite().GetName()) {
		return false
	}
	if !matchesReference(input.PopLocation.By, input.PopLocation.Input, item.GetPopLocation().GetID(), item.GetPopLocation().GetName()) {
		return false
	}
	if !strings.EqualFold(item.GetServiceProviderName(), input.ServiceProviderName) || item.GetSubnet() != input.Subnet ||
		item.GetPrivateCatoIP() != input.PrivateCatoIP || item.GetPrivateSiteIP() != input.PrivateSiteIP ||
		item.GetUpstreamBwLimit() != input.UpstreamBwLimit || item.GetDownstreamBwLimit() != input.DownstreamBwLimit {
		return false
	}
	return matchesTagging(item, input)
}
func matchesTagging(item *connectionRecord, input m.AddCloudInterconnectPhysicalConnectionInput) bool {
	if input.EncapsulationMethod == m.TaggingMethodDot1q {
		if item.GetVlan() == nil {
			return false
		}
		return input.Vlan == nil || *item.GetVlan() == *input.Vlan
	}
	if item.GetCVlan() == nil || item.GetSVlan() == nil {
		return false
	}
	qinq := input.QinQVlanConfiguration
	return qinq == nil || (*item.GetCVlan() == qinq.CVlan && *item.GetSVlan() == qinq.SVlan)
}
