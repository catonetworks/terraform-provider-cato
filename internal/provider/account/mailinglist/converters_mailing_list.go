package mailinglist

import (
	"context"

	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

func buildCreate(ctx context.Context, plan resourceModel, diags *diag.Diagnostics) m.CreateMailingListInput {
	validateValues(plan, diags)
	address := convert.PrepareStrings[string](ctx, plan.Address, diags)
	if address == nil {
		address = []string{}
	}
	for _, item := range plan.Address.Elements() {
		if item.IsNull() || item.IsUnknown() {
			diags.AddError("Invalid address", "Each email address must be known and non-null.")
		}
	}
	admins := idname.PrepareIDNameSet[m.AdminRefInput](ctx, plan.Admin, diags)
	for _, item := range plan.Admin.Elements() {
		if item.IsNull() {
			diags.AddError("Invalid administrator", "Each administrator reference must be non-null.")
		}
	}
	if len(address) == 0 && len(admins) == 0 {
		diags.AddError("Empty mailing list", "At least one email address or administrator is required.")
	}
	for _, admin := range admins {
		if admin.By != m.ObjectRefByID {
			diags.AddError("Unsupported administrator reference",
				"The mailing-list API requires administrator IDs; names are returned for display only.")
		}
	}
	return m.CreateMailingListInput{Name: plan.Name.ValueString(), Address: address, Admin: admins}
}
func buildUpdate(ctx context.Context, plan resourceModel, id string, diags *diag.Diagnostics) m.UpdateMailingListInput {
	in := buildCreate(ctx, plan, diags)
	ref := listRef(id)
	return m.UpdateMailingListInput{MailingList: &ref, Name: &in.Name, Address: in.Address, Admin: in.Admin}
}
func identityState(plan resourceModel, id string) resourceModel {
	plan.ID = types.StringValue(id)
	if plan.Address.IsNull() || plan.Address.IsUnknown() {
		plan.Address = types.SetValueMust(types.StringType, nil)
	}
	if plan.Admin.IsNull() || plan.Admin.IsUnknown() {
		plan.Admin = types.SetValueMust(idname.RefObjectType, nil)
	}
	members := make([]attr.Value, 0, len(plan.Admin.Elements()))
	for _, member := range plan.Admin.Elements() {
		if member.IsNull() || member.IsUnknown() {
			continue
		}
		attrs := member.(types.Object).Attributes()
		for k, v := range attrs {
			if v.IsUnknown() {
				attrs[k] = types.StringValue("")
			}
		}
		members = append(members, types.ObjectValueMust(idname.RefAttrTypes, attrs))
	}
	plan.Admin = types.SetValueMust(idname.RefObjectType, members)
	return plan
}

type remoteList interface {
	GetID() string
	GetName() string
	GetAddress() []string
}
type remoteRef interface {
	~struct {
		ID   string `json:"id" graphql:"id"`
		Name string `json:"name" graphql:"name"`
	}
}

func parseMailingList[T remoteRef](ctx context.Context, item remoteList, admins []*T, diags *diag.Diagnostics) resourceModel {
	address := item.GetAddress()
	if address == nil {
		address = []string{}
	}
	value, d := types.SetValueFrom(ctx, types.StringType, address)
	diags.Append(d...)
	if admins == nil {
		admins = []*T{}
	}
	for _, admin := range admins {
		if admin == nil {
			diags.AddError("Invalid API response", "The API returned a null administrator reference.")
		}
	}
	members := idname.ParseIDNameSet(ctx, admins, diags)
	return resourceModel{ID: types.StringValue(item.GetID()), Name: types.StringValue(item.GetName()), Address: value, Admin: members}
}
