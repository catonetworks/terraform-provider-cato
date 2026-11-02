package secondaryazurevsocket

import (
	"context"

	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

func buildCreate(ctx context.Context, p resourceModel, d *diag.Diagnostics) m.AddSecondaryAzureVSocketInput {
	validateFields(p, d)
	in := m.AddSecondaryAzureVSocketInput{
		Site:        idname.PrepareIDName[m.SiteRefInput](ctx, p.Site, d),
		InterfaceIP: p.InterfaceIP.ValueString(),
		FloatingIP:  p.FloatingIP.ValueString(),
	}
	if in.Site == nil {
		d.AddError("Missing site", "A known non-null site reference is required.")
	}
	return in
}
func buildUpdate(in m.AddSecondaryAzureVSocketInput, id string) m.UpdateSecondaryAzureVSocketInput {
	return m.UpdateSecondaryAzureVSocketInput{ID: id, InterfaceIP: &in.InterfaceIP, FloatingIP: &in.FloatingIP}
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
		ID:          types.StringValue(item.GetID()),
		Site:        idname.ParseIDName(ctx, *item.GetSite(), d),
		InterfaceIP: types.StringValue(item.GetInterfaceIP()),
		FloatingIP:  types.StringValue(item.GetFloatingIP()),
	}
	buildCreate(ctx, p, d)
	return p
}
func matchesReference(ref *m.SiteRefInput, id, name string) bool {
	return ref != nil && ((ref.By == m.ObjectRefByID && ref.Input == id) || (ref.By == m.ObjectRefByName && ref.Input == name))
}
func matchesSocket(item *socketRecord, in m.AddSecondaryAzureVSocketInput) bool {
	return matchesReference(in.Site, item.GetSite().GetID(), item.GetSite().GetName()) &&
		item.GetInterfaceIP() == in.InterfaceIP &&
		item.GetFloatingIP() == in.FloatingIP
}
