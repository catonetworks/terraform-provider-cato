package customapplication

import (
	"context"
	"encoding/json"

	cato "github.com/catonetworks/cato-go-sdk"
	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/cato-go-sdk/scalars"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

func buildCreate(ctx context.Context, p resourceModel, d *diag.Diagnostics) m.AddCustomApplicationInput {
	validateRoot(p, d)
	validateKnown(p.Criteria, d)
	in := m.AddCustomApplicationInput{Name: p.Name.ValueString(),
		Description: new(string),
		Category:    []*m.ApplicationCategoryRefInput{},
		Criteria:    []*m.CustomApplicationCriteriaInput{}}
	*in.Description = p.Description.ValueString()
	if d.HasError() {
		return in
	}
	in.Category = idname.PrepareIDNameSet[m.ApplicationCategoryRefInput](ctx, p.Category, d)
	for _, value := range p.Criteria.Elements() {
		if value.IsNull() {
			d.AddError("Null criterion", "Criteria cannot contain null members.")
			continue
		}
		attrs := value.(types.Object).Attributes()
		ports := attrs["port"].(types.Set)
		ranges := attrs["port_range"].(types.Set)
		if ports.IsNull() || ranges.IsNull() || attrs["protocol"].IsNull() {
			d.AddError("Invalid criterion", "Protocol, port and port_range must be explicitly configured; use empty "+
				"sets for unused collections.")
			continue
		}
		item := &m.CustomApplicationCriteriaInput{Protocol: m.IPProtocol(attrs["protocol"].(types.String).ValueString()),
			Port: convert.PrepareStrings[scalars.Port](ctx,
				ports,
				d),
			PortRange: []*m.PortRangeInput{}}
		if item.Port == nil {
			item.Port = []scalars.Port{}
		}
		for _, value := range ranges.Elements() {
			pair := value.(types.Object).Attributes()
			item.PortRange = append(item.PortRange,
				&m.PortRangeInput{From: scalars.Port(pair["from"].(types.String).ValueString()),
					To: scalars.Port(pair["to"].(types.String).ValueString())})
		}
		item.Destination = buildDestination(ctx, attrs["destination"].(types.Object), d)
		validateCriterion(item, d)
		in.Criteria = append(in.Criteria, item)
	}
	return in
}
func buildDestination(ctx context.Context, value types.Object, d *diag.Diagnostics) *m.CustomApplicationDestinationInput {
	if value.IsNull() {
		return nil
	}
	attrs := value.Attributes()
	in := &m.CustomApplicationDestinationInput{Domain: convert.PrepareStrings[string](ctx,
		attrs["domain"].(types.Set),
		d),
		Fqdn: convert.PrepareStrings[string](ctx,
			attrs["fqdn"].(types.Set),
			d)}
	ip := attrs["destination_ip"].(types.Object)
	if ip.IsNull() {
		return in
	}
	fields := ip.Attributes()
	in.DestinationIP = &m.CustomApplicationIPInput{IP: convert.PrepareStrings[string](ctx,
		fields["ip"].(types.Set),
		d),
		Subnet: convert.PrepareStrings[string](ctx,
			fields["subnet"].(types.Set),
			d),
		IPRange: []*m.IPAddressRangeInput{}}
	for _, value := range fields["ip_range"].(types.Set).Elements() {
		pair := value.(types.Object).Attributes()
		in.DestinationIP.IPRange = append(in.DestinationIP.IPRange,
			&m.IPAddressRangeInput{From: pair["from"].(types.String).ValueString(),
				To: pair["to"].(types.String).ValueString()})
	}
	return in
}
func buildUpdate(in m.AddCustomApplicationInput, id string) m.UpdateCustomApplicationInput {
	return m.UpdateCustomApplicationInput{ID: id, Name: &in.Name, Description: in.Description, Category: in.Category, Criteria: in.Criteria}
}
func identityState(p resourceModel, id string) resourceModel {
	p.ID = types.StringValue(id)
	members := []attr.Value{}
	for _, value := range p.Category.Elements() {
		attrs := value.(types.Object).Attributes()
		for key, item := range attrs {
			if item.IsUnknown() {
				attrs[key] = types.StringValue("")
			}
		}
		members = append(members, types.ObjectValueMust(idname.RefAttrTypes, attrs))
	}
	p.Category = types.SetValueMust(types.ObjectType{AttrTypes: idname.RefAttrTypes}, members)
	if p.Description.IsNull() {
		p.Description = types.StringValue("")
	}
	return p
}
func mutationRecord[T any](value *T, d *diag.Diagnostics) *appRecord {
	if value == nil {
		d.AddError("Invalid API response", "API returned no custom application.")
		return nil
	}
	raw, err := json.Marshal(value)
	if err != nil {
		d.AddError("Invalid API response", "Could not encode application fields.")
		return nil
	}
	item := new(appRecord)
	if json.Unmarshal(raw, item) != nil {
		d.AddError("Invalid API response", "Could not decode application fields.")
		return nil
	}
	return item
}
func parseApplication(ctx context.Context, item *appRecord, d *diag.Diagnostics) resourceModel {
	p := resourceModel{ID: types.StringValue(item.GetID()),
		Name:        types.StringValue(item.GetName()),
		Description: types.StringValue(""),
		Category: types.SetValueMust(types.ObjectType{AttrTypes: idname.RefAttrTypes},
			[]attr.Value{})}
	if item.GetDescription() != nil {
		p.Description = types.StringValue(*item.GetDescription())
	}
	for _, category := range item.GetCategory() {
		if category == nil || category.GetID() == "" {
			d.AddError("Invalid API response", "API returned an incomplete category identity.")
		}
	}
	if item.GetCategory() != nil {
		p.Category = idname.ParseIDNameSet(ctx, item.GetCategory(), d)
	}
	rows := []attr.Value{}
	for _, item := range item.GetCriteria() {
		if item == nil {
			d.AddError("Invalid API response", "API returned a null criterion.")
			continue
		}
		ports := []string{}
		for _, value := range item.GetPort() {
			ports = append(ports, string(value))
		}
		ranges := []attr.Value{}
		for _, pair := range item.GetPortRange() {
			if pair == nil {
				d.AddError("Invalid API response", "API returned a null port range.")
				continue
			}
			ranges = append(ranges, parseRange(string(*pair.GetFrom()), string(*pair.GetTo())))
		}
		rows = append(rows, types.ObjectValueMust(criteriaAttrTypes, map[string]attr.Value{
			"protocol": types.StringValue(string(*item.GetProtocol())), "port": parseStrings(ctx, ports, d),
			"port_range": types.SetValueMust(types.ObjectType{AttrTypes: rangeAttrTypes},
				ranges),
			"destination": parseDestination(ctx,
				item.GetDestination(),
				d),
		}))
	}
	p.Criteria = types.SetValueMust(types.ObjectType{AttrTypes: criteriaAttrTypes}, rows)
	buildCreate(ctx, p, d)
	return p
}

type destinationRecord = cato.CustomAppDataCustomApplication_CustomAppData_CustomApplication_Criteria_Destination

func parseDestination(ctx context.Context, item *destinationRecord, d *diag.Diagnostics) types.Object {
	if item == nil {
		return types.ObjectNull(destinationAttrTypes)
	}
	ip := types.ObjectNull(ipAttrTypes)
	if item.GetDestinationIP() != nil {
		value := item.GetDestinationIP()
		ranges := []attr.Value{}
		for _, pair := range value.GetIPRange() {
			if pair == nil {
				d.AddError("Invalid API response", "API returned a null IP range.")
				continue
			}
			ranges = append(ranges, parseRange(pair.GetFrom(), pair.GetTo()))
		}
		ip = types.ObjectValueMust(ipAttrTypes,
			map[string]attr.Value{"ip": parseStrings(ctx,
				value.GetIP(),
				d),
				"subnet": parseStrings(ctx,
					value.GetSubnet(),
					d),
				"ip_range": types.SetValueMust(types.ObjectType{AttrTypes: rangeAttrTypes},
					ranges)})
	}
	return types.ObjectValueMust(destinationAttrTypes,
		map[string]attr.Value{"domain": parseStrings(ctx,
			item.GetDomain(),
			d),
			"fqdn": parseStrings(ctx,
				item.GetFqdn(),
				d),
			"destination_ip": ip})
}
func parseStrings[T ~string](ctx context.Context, values []T, d *diag.Diagnostics) types.Set {
	if values == nil {
		values = []T{}
	}
	result, diags := types.SetValueFrom(ctx, types.StringType, values)
	d.Append(diags...)
	return result
}
func parseRange(from, to string) types.Object {
	return types.ObjectValueMust(rangeAttrTypes, map[string]attr.Value{"from": types.StringValue(from), "to": types.StringValue(to)})
}
func categoriesMatch(in []*m.ApplicationCategoryRefInput, item *appRecord) bool {
	if len(in) != len(item.GetCategory()) {
		return false
	}
	for _, ref := range in {
		found := false
		for _, actual := range item.GetCategory() {
			if actual != nil &&
				((ref.By == m.ObjectRefByID &&
					ref.Input == actual.GetID()) ||
					(ref.By == m.ObjectRefByName &&
						ref.Input == actual.GetName())) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
