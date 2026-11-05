package subscriptiongroup

import (
	"context"

	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

func buildMembers(ctx context.Context, plan resourceModel, diags *diag.Diagnostics) m.CreateSubscriptionGroupInput {
	validateValues(plan, diags)
	lists := idname.PrepareIDNameSet[m.MailingListRefInput](ctx, plan.MailingList, diags)
	integrations := idname.PrepareIDNameSet[m.IntegrationRefInput](ctx, plan.Integration, diags)
	validateNullMembers(plan.MailingList, diags)
	validateNullMembers(plan.Integration, diags)

	return m.CreateSubscriptionGroupInput{Name: plan.Name.ValueString(), MailingList: lists, Integration: integrations}
}
func buildCreate(ctx context.Context, plan resourceModel, diags *diag.Diagnostics) m.CreateSubscriptionGroupInput {
	input := buildMembers(ctx, plan, diags)
	if len(input.MailingList) == 0 && len(input.Integration) == 0 {
		diags.AddError("Empty subscription group", "At least one mailing list or integration is required on creation.")
	}
	return input
}
func buildUpdate(ctx context.Context, plan, current resourceModel, diags *diag.Diagnostics) m.UpdateSubscriptionGroupInput {
	in := buildMembers(ctx, plan, diags)
	ref := listRef(current.ID.ValueString())
	result := m.UpdateSubscriptionGroupInput{SubscriptionGroup: &ref, Name: &in.Name, MailingList: in.MailingList, Integration: in.Integration}
	// The backend treats empty set lists as absent. Remove every currently observed
	// member of a cleared collection, including members introduced outside Terraform.
	if len(in.MailingList) == 0 {
		result.MailingListToRemove = idname.PrepareIDNameSet[m.MailingListRefInput](ctx, current.MailingList, diags)
	}
	if len(in.Integration) == 0 {
		result.IntegrationToRemove = idname.PrepareIDNameSet[m.IntegrationRefInput](ctx, current.Integration, diags)
	}
	return result
}
func validateNullMembers(set types.Set, diags *diag.Diagnostics) {
	for _, item := range set.Elements() {
		if item.IsNull() {
			diags.AddError("Invalid member reference", "Members must be non-null references.")
		}
	}
}
func identityState(plan resourceModel, id string) resourceModel {
	plan.ID = types.StringValue(id)
	plan.MailingList = knownMembers(plan.MailingList)
	plan.Integration = knownMembers(plan.Integration)
	return plan
}
func knownMembers(set types.Set) types.Set {
	members := make([]attr.Value, 0, len(set.Elements()))
	for _, member := range set.Elements() {
		if member.IsNull() || member.IsUnknown() {
			continue
		}
		attrs := member.(types.Object).Attributes()
		for key, value := range attrs {
			if value.IsUnknown() {
				attrs[key] = types.StringValue("")
			}
		}
		members = append(members, types.ObjectValueMust(idname.RefAttrTypes, attrs))
	}
	return types.SetValueMust(idname.RefObjectType, members)
}

type remoteGroup interface {
	GetID() string
	GetName() string
}
type remoteMember interface {
	GetID() string
	GetName() string
}

func parseSubscriptionGroup[L remoteMember, I remoteMember](ctx context.Context, item remoteGroup, lists []L, integrations []I,
	diags *diag.Diagnostics,
) resourceModel {
	return resourceModel{ID: types.StringValue(item.GetID()), Name: types.StringValue(item.GetName()),
		MailingList: parseMembers(ctx, lists, diags), Integration: parseMembers(ctx, integrations, diags)}
}
func parseMembers[T remoteMember](ctx context.Context, items []T, diags *diag.Diagnostics) types.Set {
	members := make([]attr.Value, 0, len(items))
	for _, item := range items {
		if item.GetID() == "" {
			diags.AddError("Invalid API response", "A member reference has no ID.")
			continue
		}
		value, d := types.ObjectValueFrom(ctx, idname.RefAttrTypes,
			idname.Ref{ID: types.StringValue(item.GetID()), Name: types.StringValue(item.GetName())})
		diags.Append(d...)
		members = append(members, value)
	}
	set, d := types.SetValue(idname.RefObjectType, members)
	diags.Append(d...)
	return set
}
