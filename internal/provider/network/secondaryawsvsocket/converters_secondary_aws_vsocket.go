package secondaryawsvsocket

import (
	"context"

	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

func buildCreate(ctx context.Context, p resourceModel, d *diag.Diagnostics) m.AddSecondaryAWSVSocketInput {
	validateFields(p, d)
	in := m.AddSecondaryAWSVSocketInput{
		Site:         idname.PrepareIDName[m.SiteRefInput](ctx, p.Site, d),
		EniIPAddress: p.EniIPAddress.ValueString(),
		EniIPSubnet:  p.EniIPSubnet.ValueString(),
		RouteTableID: p.RouteTableID.ValueString(),
	}
	if in.Site == nil {
		d.AddError("Missing site", "A known non-null site reference is required.")
	}
	return in
}
func buildUpdate(in m.AddSecondaryAWSVSocketInput, id string) m.UpdateSecondaryAWSVSocketInput {
	return m.UpdateSecondaryAWSVSocketInput{ID: id, IPAddress: &in.EniIPAddress, Subnet: &in.EniIPSubnet,
		RouteTableID: &in.RouteTableID}
}
func identityState(p resourceModel, id string) resourceModel {
	p.ID = types.StringValue(id)
	attrs := p.Site.Attributes()
	for key, value := range attrs {
		if value.IsUnknown() {
			attrs[key] = types.StringValue("")
		}
	}
	p.Site = types.ObjectValueMust(idname.RefAttrTypes, attrs)
	return p
}
func parseSocket(ctx context.Context, item *socketRecord, d *diag.Diagnostics) resourceModel {
	p := resourceModel{
		ID:           types.StringValue(item.GetID()),
		Site:         idname.ParseIDName(ctx, *item.GetSite(), d),
		EniIPAddress: types.StringValue(item.GetIPAddress()),
		EniIPSubnet:  types.StringValue(item.GetSubnet()),
		RouteTableID: types.StringValue(item.GetRouteTableID()),
	}
	buildCreate(ctx, p, d)
	return p
}
func matchesReference(ref *m.SiteRefInput, id, name string) bool {
	return ref != nil && ((ref.By == m.ObjectRefByID && ref.Input == id) || (ref.By == m.ObjectRefByName && ref.Input == name))
}
func matchesSocket(item *socketRecord, in m.AddSecondaryAWSVSocketInput) bool {
	return matchesReference(in.Site, item.GetSite().GetID(), item.GetSite().GetName()) &&
		item.GetIPAddress() == in.EniIPAddress &&
		item.GetSubnet() == in.EniIPSubnet &&
		item.GetRouteTableID() == in.RouteTableID
}
