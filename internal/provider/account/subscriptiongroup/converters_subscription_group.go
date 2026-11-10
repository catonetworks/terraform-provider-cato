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
	lists := idname.PrepareIDNameList[m.MailingListRefInput](ctx, plan.MailingList, diags)
	integrations := idname.PrepareIDNameList[m.IntegrationRefInput](ctx, plan.Integration, diags)
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
		result.MailingListToRemove = idname.PrepareIDNameList[m.MailingListRefInput](ctx, current.MailingList, diags)
	}
	if len(in.Integration) == 0 {
		result.IntegrationToRemove = idname.PrepareIDNameList[m.IntegrationRefInput](ctx, current.Integration, diags)
	}
	return result
}
func validateNullMembers(set types.List, diags *diag.Diagnostics) {
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
func knownMembers(set types.List) types.List {
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
	return types.ListValueMust(idname.RefObjectType, members)
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
func parseMembers[T remoteMember](ctx context.Context, items []T, diags *diag.Diagnostics) types.List {
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
	list, d := types.ListValue(idname.RefObjectType, members)
	diags.Append(d...)
	return list
}

// orderLike reorders members to match the order of reference by matching each reference
// element on its configured selector (name when set, otherwise id). Membership is
// unordered on the API side, but the schema models it as a List (so a name-only ref whose
// Computed id is Unknown correlates by index); re-reads can return members in a different
// order than configured, so align to the desired order to keep apply/refresh plans empty.
// Any members not present in reference are appended in their existing order.
func orderLike(members, reference types.List) types.List {
	if members.IsNull() || members.IsUnknown() || reference.IsNull() || reference.IsUnknown() {
		return members
	}
	remaining := append([]attr.Value{}, members.Elements()...)
	ordered := make([]attr.Value, 0, len(remaining))
	for _, ref := range reference.Elements() {
		robj, ok := ref.(types.Object)
		if !ok || robj.IsNull() || robj.IsUnknown() {
			continue
		}
		ra := robj.Attributes()
		rName, _ := ra["name"].(types.String)
		rID, _ := ra["id"].(types.String)
		byName := !rName.IsNull() && !rName.IsUnknown() && rName.ValueString() != ""
		for i, cand := range remaining {
			ca := cand.(types.Object).Attributes()
			cName := ca["name"].(types.String)
			cID := ca["id"].(types.String)
			hit := false
			if byName {
				hit = !cName.IsNull() && cName.ValueString() == rName.ValueString()
			} else if !rID.IsNull() && !rID.IsUnknown() {
				hit = !cID.IsNull() && cID.ValueString() == rID.ValueString()
			}
			if hit {
				ordered = append(ordered, cand)
				remaining = append(remaining[:i], remaining[i+1:]...)
				break
			}
		}
	}
	ordered = append(ordered, remaining...)
	return types.ListValueMust(idname.RefObjectType, ordered)
}

// orderMembersLike aligns a freshly-parsed state's member lists to the order of a
// reference model (the plan on create/update, or the prior state on read).
func orderMembersLike(current, reference resourceModel) resourceModel {
	current.MailingList = orderLike(current.MailingList, reference.MailingList)
	current.Integration = orderLike(current.Integration, reference.Integration)
	return current
}
