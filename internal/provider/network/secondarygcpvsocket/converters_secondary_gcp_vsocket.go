package secondarygcpvsocket

import (
	"context"

	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

func buildCreate(ctx context.Context, p resourceModel, d *diag.Diagnostics) m.AddSecondaryGCPVSocketInput {
	validateFields(p, d)
	in := m.AddSecondaryGCPVSocketInput{Site: idname.PrepareIDName[m.SiteRefInput](ctx, p.Site, d)}
	if in.Site == nil {
		d.AddError("Missing site", "A known non-null site reference is required.")
	}
	if !d.HasError() {
		attrs := p.GCPConfig.Attributes()
		in.GCPConfig = &m.GCPConfigInput{
			InterfaceIP:    attrs["interface_ip"].(types.String).ValueString(),
			LoadBalancerIP: attrs["load_balancer_ip"].(types.String).ValueString(),
		}
	}
	return in
}
func buildUpdate(in m.AddSecondaryGCPVSocketInput, id string) m.UpdateSecondaryGCPVSocketInput {
	return m.UpdateSecondaryGCPVSocketInput{ID: id, GCPConfig: in.GCPConfig}
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
	config, diags := types.ObjectValue(gcpConfigAttrTypes, map[string]attr.Value{
		"interface_ip":     types.StringValue(item.GetInterfaceIP()),
		"load_balancer_ip": types.StringValue(item.GetLoadBalancerIP()),
	})
	d.Append(diags...)
	p := resourceModel{ID: types.StringValue(item.GetID()), Site: idname.ParseIDName(ctx, *item.GetSite(), d), GCPConfig: config}
	buildCreate(ctx, p, d)
	return p
}

func matchesReference(ref *m.SiteRefInput, id, name string) bool {
	return ref != nil && ((ref.By == m.ObjectRefByID && ref.Input == id) || (ref.By == m.ObjectRefByName && ref.Input == name))
}
func matchesSocket(item *socketRecord, in m.AddSecondaryGCPVSocketInput) bool {
	return matchesReference(in.Site, item.GetSite().GetID(), item.GetSite().GetName()) &&
		in.GCPConfig != nil && item.GetInterfaceIP() == in.GCPConfig.InterfaceIP &&
		item.GetLoadBalancerIP() == in.GCPConfig.LoadBalancerIP
}
