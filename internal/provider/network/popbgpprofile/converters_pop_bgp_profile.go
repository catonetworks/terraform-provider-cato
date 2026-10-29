package popbgpprofile

import (
	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/cato-go-sdk/scalars"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
)

func buildCreate(plan resourceModel, diags *diag.Diagnostics) m.PopLocationAddBgpProfileInput {
	validateFields(plan, diags)
	return m.PopLocationAddBgpProfileInput{Name: plan.Name.ValueString(), Description: convert.StringPointerForOptionalInput(plan.Description),
		Communities: buildCommunities(plan.Communities, diags)}
}
func buildCommunities(value types.Set, diags *diag.Diagnostics) []*m.BgpCommunityInput {
	result := make([]*m.BgpCommunityInput, 0, len(value.Elements()))
	for _, item := range value.Elements() {
		if item.IsNull() || item.IsUnknown() {
			diags.AddError("Invalid community", "Each community must be known and non-null.")
			continue
		}
		attrs := item.(types.Object).Attributes()
		from, to := attrs["from"].(types.String), attrs["to"].(types.String)
		if !validASN(from.ValueString()) || !validASN(to.ValueString()) {
			diags.AddError("Invalid community", "Both values must be canonical decimal integers from 0 through 65535.")
			continue
		}
		result = append(result, &m.BgpCommunityInput{From: scalars.Asn16(from.ValueString()), To: scalars.Asn16(to.ValueString())})
	}
	return result
}
func buildUpdate(plan, current resourceModel, diags *diag.Diagnostics) m.PopLocationUpdateBgpProfileInput {
	input := buildCreate(plan, diags)
	description := plan.Description.ValueString()
	result := m.PopLocationUpdateBgpProfileInput{ID: current.ID.ValueString(), Name: &input.Name, Description: &description}
	// Attached profiles cannot change communities; omit unchanged communities for metadata-only updates.
	if !plan.Communities.Equal(current.Communities) {
		result.Communities = input.Communities
	}
	return result
}
func identityState(plan resourceModel, id string) resourceModel {
	plan.ID = types.StringValue(id)
	if plan.Description.IsNull() {
		plan.Description = types.StringValue("")
	}
	return plan
}

type remoteProfile interface {
	GetID() string
	GetName() string
	GetDescription() *string
}
type remoteCommunity interface {
	GetFrom() *scalars.Asn16
	GetTo() *scalars.Asn16
}

func parseProfile[T remoteCommunity](item remoteProfile, communities []T, diags *diag.Diagnostics) resourceModel {
	members := make([]attr.Value, 0, len(communities))
	for _, c := range communities {
		from, to := c.GetFrom(), c.GetTo()
		if from == nil || to == nil || !validASN(string(*from)) || !validASN(string(*to)) {
			diags.AddError("Invalid API response", "Profile contains an invalid community pair.")
			continue
		}
		members = append(members, types.ObjectValueMust(communityAttrTypes, map[string]attr.Value{
			"from": types.StringValue(string(*from)), "to": types.StringValue(string(*to)),
		}))
	}
	if len(members) == 0 {
		diags.AddError("Invalid API response", "Profile did not return any valid community pairs.")
	}
	value, d := types.SetValue(communityObjectType, members)
	diags.Append(d...)
	return resourceModel{ID: types.StringValue(item.GetID()), Name: types.StringValue(item.GetName()),
		Description: convert.ParseOptionalString(item.GetDescription()), Communities: value}
}
