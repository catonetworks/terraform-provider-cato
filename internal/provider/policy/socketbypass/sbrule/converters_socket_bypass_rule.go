package sbrule

import (
	"context"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	cato_scalars "github.com/catonetworks/cato-go-sdk/scalars"
	attr "github.com/hashicorp/terraform-plugin-framework/attr"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	convert "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	oldidname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/oldidname"
	policyrule "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
	refinput "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/refinput"
)

var sbAsOpts = basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

func sbPortStr(p *cato_scalars.Port) string {
	if p == nil {
		return ""
	}
	return string(*p)
}

// ── plan → SDK input builders ───────────────────────────────────────────────

func sbFromToRanges(
	ctx context.Context,
	list types.List,
	diags *diag.Diagnostics,
) []*cato_models.IPAddressRangeInput {
	out := []*cato_models.IPAddressRangeInput{}
	if list.IsNull() || list.IsUnknown() {
		return out
	}
	var rows []SocketBypassFromToPlan
	diags.Append(list.ElementsAs(ctx, &rows, false)...)
	for _, row := range rows {
		out = append(out, &cato_models.IPAddressRangeInput{
			From: row.From.ValueString(),
			To:   row.To.ValueString(),
		})
	}
	return out
}

func sbActionInput(
	ctx context.Context,
	obj types.Object,
	diags *diag.Diagnostics,
) *cato_models.SocketBypassActionInput {
	out := &cato_models.SocketBypassActionInput{
		Port:  cato_models.BypassSocketPortTypeAutomatic,
		Event: &cato_models.PolicyRuleTrackingEventInput{Enabled: false},
	}
	if obj.IsNull() || obj.IsUnknown() {
		return out
	}
	a := SocketBypassActionPlan{}
	diags.Append(obj.As(ctx, &a, sbAsOpts)...)
	if !a.Port.IsNull() && !a.Port.IsUnknown() && a.Port.ValueString() != "" {
		out.Port = cato_models.BypassSocketPortType(a.Port.ValueString())
	}
	if !a.EventEnabled.IsNull() && !a.EventEnabled.IsUnknown() {
		out.Event.Enabled = a.EventEnabled.ValueBool()
	}
	return out
}

func sbSiteInput(
	ctx context.Context,
	obj types.Object,
	diags *diag.Diagnostics,
) *cato_models.SocketBypassSiteInput {
	out := &cato_models.SocketBypassSiteInput{
		Site:  []*cato_models.SiteRefInput{},
		Group: []*cato_models.GroupRefInput{},
	}
	if obj.IsNull() || obj.IsUnknown() {
		return out
	}
	s := SocketBypassSitePlan{}
	diags.Append(obj.As(ctx, &s, sbAsOpts)...)
	if diags.HasError() {
		return out
	}
	if v := refinput.FromSet(ctx, s.Site, diags, "site", refinput.Site); v != nil {
		out.Site = v
	}
	if v := refinput.FromSet(ctx, s.Group, diags, "group", refinput.Group); v != nil {
		out.Group = v
	}
	return out
}

func sbSourceInput(
	ctx context.Context,
	obj types.Object,
	diags *diag.Diagnostics,
) *cato_models.SocketBypassSourceInput {
	out := &cato_models.SocketBypassSourceInput{
		FloatingSubnet:    []*cato_models.FloatingSubnetRefInput{},
		GlobalIPRange:     []*cato_models.GlobalIPRangeRefInput{},
		Group:             []*cato_models.GroupRefInput{},
		Host:              []*cato_models.HostRefInput{},
		IP:                []string{},
		IPRange:           []*cato_models.IPAddressRangeInput{},
		NetworkInterface:  []*cato_models.NetworkInterfaceRefInput{},
		SiteNetworkSubnet: []*cato_models.SiteNetworkSubnetRefInput{},
		Subnet:            []string{},
		Vlan:              []cato_scalars.Vlan{},
	}
	if obj.IsNull() || obj.IsUnknown() {
		return out
	}
	s := SocketBypassSourcePlan{}
	diags.Append(obj.As(ctx, &s, sbAsOpts)...)
	if diags.HasError() {
		return out
	}
	if v := refinput.FromSet(ctx, s.FloatingSubnet, diags, "floating subnet", refinput.FloatingSubnet); v != nil {
		out.FloatingSubnet = v
	}
	if v := refinput.FromSet(ctx, s.GlobalIPRange, diags, "global IP range", refinput.GlobalIPRange); v != nil {
		out.GlobalIPRange = v
	}
	if v := refinput.FromSet(ctx, s.Group, diags, "group", refinput.Group); v != nil {
		out.Group = v
	}
	if v := refinput.FromSet(ctx, s.Host, diags, "host", refinput.Host); v != nil {
		out.Host = v
	}
	if v := refinput.StringsFromList(ctx, s.IP, diags); v != nil {
		out.IP = v
	}
	out.IPRange = sbFromToRanges(ctx, s.IPRange, diags)
	if v := refinput.FromSet(ctx, s.NetworkInterface, diags, "network interface", refinput.NetworkInterface); v != nil {
		out.NetworkInterface = v
	}
	if v := refinput.FromSet(ctx, s.SiteNetworkSubnet, diags, "site network subnet", refinput.SiteNetworkSubnet); v != nil {
		out.SiteNetworkSubnet = v
	}
	if v := refinput.StringsFromList(ctx, s.Subnet, diags); v != nil {
		out.Subnet = v
	}
	if !s.Vlan.IsNull() && !s.Vlan.IsUnknown() {
		var vlans []int64
		diags.Append(s.Vlan.ElementsAs(ctx, &vlans, false)...)
		for _, v := range vlans {
			out.Vlan = append(out.Vlan, cato_scalars.Vlan(v))
		}
	}
	return out
}

func sbDestinationInput(
	ctx context.Context,
	obj types.Object,
	diags *diag.Diagnostics,
) *cato_models.SocketBypassDestinationInput {
	out := &cato_models.SocketBypassDestinationInput{
		Application:   []*cato_models.ApplicationRefInput{},
		CustomApp:     []*cato_models.CustomApplicationRefInput{},
		Domain:        []string{},
		Fqdn:          []string{},
		GlobalIPRange: []*cato_models.GlobalIPRangeRefInput{},
		IP:            []string{},
		IPRange:       []*cato_models.IPAddressRangeInput{},
		Subnet:        []string{},
	}
	if obj.IsNull() || obj.IsUnknown() {
		return out
	}
	d := SocketBypassDestinationPlan{}
	diags.Append(obj.As(ctx, &d, sbAsOpts)...)
	if diags.HasError() {
		return out
	}
	if v := refinput.FromSet(ctx, d.Application, diags, "application", refinput.Application); v != nil {
		out.Application = v
	}
	if v := refinput.FromSet(ctx, d.CustomApp, diags, "custom application", refinput.CustomApplication); v != nil {
		out.CustomApp = v
	}
	if v := refinput.StringsFromList(ctx, d.Domain, diags); v != nil {
		out.Domain = v
	}
	if v := refinput.StringsFromList(ctx, d.Fqdn, diags); v != nil {
		out.Fqdn = v
	}
	if v := refinput.FromSet(ctx, d.GlobalIPRange, diags, "global IP range", refinput.GlobalIPRange); v != nil {
		out.GlobalIPRange = v
	}
	if v := refinput.StringsFromList(ctx, d.IP, diags); v != nil {
		out.IP = v
	}
	out.IPRange = sbFromToRanges(ctx, d.IPRange, diags)
	if v := refinput.StringsFromList(ctx, d.Subnet, diags); v != nil {
		out.Subnet = v
	}
	return out
}

func sbServiceInput(
	ctx context.Context,
	obj types.Object,
	diags *diag.Diagnostics,
) *cato_models.SocketBypassServiceInput {
	out := &cato_models.SocketBypassServiceInput{
		Simple: []*cato_models.SimpleServiceInput{},
		Custom: []*cato_models.CustomServiceInput{},
	}
	if obj.IsNull() || obj.IsUnknown() {
		return out
	}
	s := SocketBypassServicePlan{}
	diags.Append(obj.As(ctx, &s, sbAsOpts)...)
	if diags.HasError() {
		return out
	}
	for _, name := range refinput.StringsFromList(ctx, s.Simple, diags) {
		out.Simple = append(out.Simple, &cato_models.SimpleServiceInput{
			Name: cato_models.SimpleServiceType(name),
		})
	}
	if !s.Custom.IsNull() && !s.Custom.IsUnknown() {
		var rows []SocketBypassCustomServicePlan
		diags.Append(s.Custom.ElementsAs(ctx, &rows, false)...)
		for _, row := range rows {
			svc := &cato_models.CustomServiceInput{
				Protocol: cato_models.IPProtocol(row.Protocol.ValueString()),
			}
			for _, p := range refinput.StringsFromList(ctx, row.Port, diags) {
				svc.Port = append(svc.Port, cato_scalars.Port(p))
			}
			if !row.PortRange.IsNull() && !row.PortRange.IsUnknown() {
				pr := SocketBypassFromToPlan{}
				diags.Append(row.PortRange.As(ctx, &pr, sbAsOpts)...)
				svc.PortRange = &cato_models.PortRangeInput{
					From: cato_scalars.Port(pr.From.ValueString()),
					To:   cato_scalars.Port(pr.To.ValueString()),
				}
			}
			out.Custom = append(out.Custom, svc)
		}
	}
	return out
}

func sbExceptions(
	ctx context.Context,
	list types.List,
	diags *diag.Diagnostics,
) []*cato_models.SocketBypassExceptionInput {
	out := []*cato_models.SocketBypassExceptionInput{}
	if list.IsNull() || list.IsUnknown() {
		return out
	}
	var rows []SocketBypassExceptionPlan
	diags.Append(list.ElementsAs(ctx, &rows, false)...)
	for _, row := range rows {
		out = append(out, &cato_models.SocketBypassExceptionInput{
			Name:        row.Name.ValueString(),
			Site:        sbSiteInput(ctx, row.Site, diags),
			Source:      sbSourceInput(ctx, row.Source, diags),
			Destination: sbDestinationInput(ctx, row.Destination, diags),
			Service:     sbServiceInput(ctx, row.Service, diags),
		})
	}
	return out
}

func hydrateSocketBypassAddRuleInput(
	ctx context.Context,
	plan SocketBypassRule,
) (cato_models.SocketBypassAddRuleInput, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := cato_models.SocketBypassAddRuleInput{}

	if !plan.At.IsNull() {
		out.At = &cato_models.PolicyRulePositionInput{}
		pos := policyrule.Position{}
		diags.Append(plan.At.As(ctx, &pos, sbAsOpts)...)
		out.At.Position = (*cato_models.PolicyRulePositionEnum)(pos.Position.ValueStringPointer())
		out.At.Ref = pos.Ref.ValueStringPointer()
	}

	rule := SocketBypassRuleRulePlan{}
	diags.Append(plan.Rule.As(ctx, &rule, sbAsOpts)...)
	if diags.HasError() {
		return out, diags
	}

	out.Rule = &cato_models.SocketBypassAddRuleDataInput{
		Name:        rule.Name.ValueString(),
		Description: rule.Description.ValueString(),
		Enabled:     rule.Enabled.ValueBool(),
		Action:      sbActionInput(ctx, rule.Action, &diags),
		Site:        sbSiteInput(ctx, rule.Site, &diags),
		Source:      sbSourceInput(ctx, rule.Source, &diags),
		Destination: sbDestinationInput(ctx, rule.Destination, &diags),
		Service:     sbServiceInput(ctx, rule.Service, &diags),
		Exception:   sbExceptions(ctx, rule.Exception, &diags),
	}
	return out, diags
}

func hydrateSocketBypassUpdateRuleInput(
	ctx context.Context,
	plan SocketBypassRule,
) (cato_models.SocketBypassUpdateRuleInput, diag.Diagnostics) {
	var diags diag.Diagnostics

	rule := SocketBypassRuleRulePlan{}
	diags.Append(plan.Rule.As(ctx, &rule, sbAsOpts)...)
	if diags.HasError() {
		return cato_models.SocketBypassUpdateRuleInput{}, diags
	}

	action := sbActionInput(ctx, rule.Action, &diags)
	site := sbSiteInput(ctx, rule.Site, &diags)
	src := sbSourceInput(ctx, rule.Source, &diags)
	dst := sbDestinationInput(ctx, rule.Destination, &diags)
	svc := sbServiceInput(ctx, rule.Service, &diags)

	upd := cato_models.SocketBypassUpdateRuleInput{
		ID: rule.ID.ValueString(),
		Rule: &cato_models.SocketBypassUpdateRuleDataInput{
			Name:        rule.Name.ValueStringPointer(),
			Description: rule.Description.ValueStringPointer(),
			Enabled:     rule.Enabled.ValueBoolPointer(),
			Action: &cato_models.SocketBypassActionUpdateInput{
				Port:  &action.Port,
				Event: &cato_models.PolicyRuleTrackingEventUpdateInput{Enabled: &action.Event.Enabled},
			},
			Site: &cato_models.SocketBypassSiteUpdateInput{
				Site:  site.Site,
				Group: site.Group,
			},
			Source: &cato_models.SocketBypassSourceUpdateInput{
				FloatingSubnet:    src.FloatingSubnet,
				GlobalIPRange:     src.GlobalIPRange,
				Group:             src.Group,
				Host:              src.Host,
				IP:                src.IP,
				IPRange:           src.IPRange,
				NetworkInterface:  src.NetworkInterface,
				SiteNetworkSubnet: src.SiteNetworkSubnet,
				Subnet:            src.Subnet,
				Vlan:              src.Vlan,
			},
			Destination: &cato_models.SocketBypassDestinationUpdateInput{
				Application:   dst.Application,
				CustomApp:     dst.CustomApp,
				Domain:        dst.Domain,
				Fqdn:          dst.Fqdn,
				GlobalIPRange: dst.GlobalIPRange,
				IP:            dst.IP,
				IPRange:       dst.IPRange,
				Subnet:        dst.Subnet,
			},
			Service: &cato_models.SocketBypassServiceUpdateInput{
				Simple: svc.Simple,
				Custom: svc.Custom,
			},
			Exception: sbExceptions(ctx, rule.Exception, &diags),
		},
	}
	return upd, diags
}

// ── standard convert helpers (same shapes as the other family rules) ───────

func convertCreateInput(
	ctx context.Context,
	priorRulePlan *SocketBypassRuleRulePlan,
	planRule SocketBypassRuleRulePlan,
) (resultRuleObj basetypes.ObjectValue, resultOdiags diag.Diagnostics) {
	rulePlan := *priorRulePlan
	defer func() { *priorRulePlan = rulePlan }()
	if !planRule.Exception.IsNull() && !planRule.Exception.IsUnknown() {
		rulePlan.Exception = planRule.Exception
	}
	resultRuleObj, resultOdiags = types.ObjectValueFrom(ctx, SocketBypassRuleRuleAttrTypes, rulePlan)
	return resultRuleObj, resultOdiags
}

func convertCreateInput2(plan *SocketBypassRule, newID string, ruleObj basetypes.ObjectValue) {
	plan.ID = types.StringValue(newID)
	plan.Rule = ruleObj
}

func convertReadState(
	ctx context.Context,
	rulePlan SocketBypassRuleRulePlan,
	priorHydrated *SocketBypassRuleRulePlan,
) (resultRuleObj basetypes.ObjectValue, resultOdiags diag.Diagnostics) {
	hydrated := *priorHydrated
	defer func() { *priorHydrated = hydrated }()
	if !rulePlan.Exception.IsNull() && !rulePlan.Exception.IsUnknown() {
		hydrated.Exception = rulePlan.Exception
	}
	resultRuleObj, resultOdiags = types.ObjectValueFrom(ctx, SocketBypassRuleRuleAttrTypes, hydrated)
	return resultRuleObj, resultOdiags
}

func convertReadState2(
	state *SocketBypassRule,
	cur *cato_go_sdk.PolicySocketBypassPolicy_Policy_SocketBypass_Policy_Rules_Rule,
	ruleObj basetypes.ObjectValue,
	diagnostics *diag.Diagnostics,
) {
	if state.At.IsNull() || state.At.IsUnknown() {
		atObj, d := types.ObjectValue(policyrule.PositionAttrTypes, map[string]attr.Value{
			"position": types.StringValue("LAST_IN_POLICY"),
			"ref":      types.StringNull(),
		})
		diagnostics.Append(d...)
		state.At = atObj
	}

	state.ID = types.StringValue(cur.GetID())
	state.Rule = ruleObj
}

func convertUpdateInput(
	ctx context.Context,
	plan SocketBypassRule,
	diagnostics *diag.Diagnostics,
) (
	resultMove cato_models.PolicyMoveRuleInput,
	resultAtAsOpts basetypes.ObjectAsOptions,
) {
	resultMove = cato_models.PolicyMoveRuleInput{}
	resultAtAsOpts = basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}
	if !plan.At.IsNull() {
		pos := policyrule.Position{}
		diagnostics.Append(plan.At.As(ctx, &pos, resultAtAsOpts)...)
		resultMove.To = &cato_models.PolicyRulePositionInput{
			Position: (*cato_models.PolicyRulePositionEnum)(pos.Position.ValueStringPointer()),
			Ref:      pos.Ref.ValueStringPointer(),
		}
	}
	return resultMove, resultAtAsOpts
}

func convertUpdateInput2(
	ctx context.Context,
	rule SocketBypassRuleRulePlan,
	priorRulePlan *SocketBypassRuleRulePlan,
) (resultRuleObj basetypes.ObjectValue, resultOdiags diag.Diagnostics) {
	rulePlan := *priorRulePlan
	defer func() { *priorRulePlan = rulePlan }()
	if !rule.Exception.IsNull() && !rule.Exception.IsUnknown() {
		rulePlan.Exception = rule.Exception
	}
	resultRuleObj, resultOdiags = types.ObjectValueFrom(ctx, SocketBypassRuleRuleAttrTypes, rulePlan)
	return resultRuleObj, resultOdiags
}

func convertUpdateInput3(pos policyrule.Position) (resultAtObj basetypes.ObjectValue, resultAdiags diag.Diagnostics) {
	resultAtObj, resultAdiags = types.ObjectValue(policyrule.PositionAttrTypes, map[string]attr.Value{
		"position": pos.Position,
		"ref":      pos.Ref,
	})
	return resultAtObj, resultAdiags
}

func convertUpdateInput4(
	rule SocketBypassRuleRulePlan,
	ruleObj basetypes.ObjectValue,
	atObj basetypes.ObjectValue,
) (resultSt SocketBypassRule) {
	resultSt = SocketBypassRule{
		ID:   types.StringValue(rule.ID.ValueString()),
		At:   atObj,
		Rule: ruleObj,
	}
	return resultSt
}

func convertDeleteInput(rule SocketBypassRuleRulePlan) (resultRm cato_models.SocketBypassRemoveRuleInput) {
	resultRm = cato_models.SocketBypassRemoveRuleInput{ID: rule.ID.ValueString()}
	return resultRm
}

// ── SDK response → state hydrators ─────────────────────────────────────────

func sbFromToObj(from, to string, diags *diag.Diagnostics) types.Object {
	o, d := types.ObjectValue(SocketBypassFromToAttrTypes, map[string]attr.Value{
		"from": types.StringValue(from),
		"to":   types.StringValue(to),
	})
	diags.Append(d...)
	return o
}

type sbIPRangeGetter interface {
	GetFrom() string
	GetTo() string
}

func sbRangeList[T sbIPRangeGetter](items []T, diags *diag.Diagnostics) types.List {
	if len(items) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: SocketBypassFromToAttrTypes})
	}
	elems := make([]attr.Value, 0, len(items))
	for _, it := range items {
		elems = append(elems, sbFromToObj(it.GetFrom(), it.GetTo(), diags))
	}
	lst, d := types.ListValue(types.ObjectType{AttrTypes: SocketBypassFromToAttrTypes}, elems)
	diags.Append(d...)
	return lst
}

// hydrateSocketBypassRuleStateFromClient maps a policy read rule into Terraform nested objects.
//
//nolint:gocyclo,funlen
func hydrateSocketBypassRuleStateFromClient(
	ctx context.Context,
	r *cato_go_sdk.PolicySocketBypassPolicy_Policy_SocketBypass_Policy_Rules_Rule,
) (SocketBypassRuleRulePlan, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := SocketBypassRuleRulePlan{}
	if r == nil {
		return out, diags
	}

	out.Action = types.ObjectNull(SocketBypassActionAttrTypes)
	out.Site = types.ObjectNull(SocketBypassSiteAttrTypes)
	out.Source = types.ObjectNull(SocketBypassSourceAttrTypes)
	out.Destination = types.ObjectNull(SocketBypassDestinationAttrTypes)
	out.Service = types.ObjectNull(SocketBypassServiceAttrTypes)
	out.Exception = types.ListNull(types.ObjectType{AttrTypes: SocketBypassExceptionAttrTypes})

	out.ID = types.StringValue(r.GetID())
	out.Name = types.StringValue(r.GetName())
	if d := r.GetDescription(); d != "" {
		out.Description = types.StringValue(d)
	} else {
		out.Description = types.StringNull()
	}
	out.Enabled = types.BoolValue(r.GetEnabled())

	if a := r.GetAction(); a != nil {
		port := ""
		if a.GetPort() != nil {
			port = string(*a.GetPort())
		}
		eventEnabled := false
		if a.GetEvent() != nil {
			eventEnabled = a.GetEvent().GetEnabled()
		}
		o, d := types.ObjectValue(SocketBypassActionAttrTypes, map[string]attr.Value{
			"port":          types.StringValue(port),
			"event_enabled": types.BoolValue(eventEnabled),
		})
		diags.Append(d...)
		out.Action = o
	}

	if s := r.GetSite(); s != nil && (len(s.GetSite()) > 0 || len(s.GetGroup()) > 0) {
		o, d := types.ObjectValue(SocketBypassSiteAttrTypes, map[string]attr.Value{
			"site":  oldidname.ParseNameIDList(ctx, s.GetSite(), "rule.site.site"),
			"group": oldidname.ParseNameIDList(ctx, s.GetGroup(), "rule.site.group"),
		})
		diags.Append(d...)
		out.Site = o
	}

	if s := r.GetSource(); s != nil {
		vlans := types.ListNull(types.Int64Type)
		if vs := s.GetVlan(); len(vs) > 0 {
			elems := make([]attr.Value, 0, len(vs))
			for _, v := range vs {
				elems = append(elems, types.Int64Value(int64(v)))
			}
			lst, d := types.ListValue(types.Int64Type, elems)
			diags.Append(d...)
			vlans = lst
		}
		o, d := types.ObjectValue(SocketBypassSourceAttrTypes, map[string]attr.Value{
			"floating_subnet":     oldidname.ParseNameIDList(ctx, s.GetFloatingSubnet(), "rule.source.floating_subnet"),
			"global_ip_range":     oldidname.ParseNameIDList(ctx, s.GetGlobalIPRange(), "rule.source.global_ip_range"),
			"group":               oldidname.ParseNameIDList(ctx, s.GetGroup(), "rule.source.group"),
			"host":                oldidname.ParseNameIDList(ctx, s.GetHost(), "rule.source.host"),
			"ip":                  convert.ParseList(ctx, types.StringType, s.GetIP(), "rule.source.ip"),
			"ip_range":            sbRangeList(s.GetIPRange(), &diags),
			"network_interface":   oldidname.ParseNameIDList(ctx, s.GetNetworkInterface(), "rule.source.network_interface"),
			"site_network_subnet": oldidname.ParseNameIDList(ctx, s.GetSiteNetworkSubnet(), "rule.source.site_network_subnet"),
			"subnet":              convert.ParseList(ctx, types.StringType, s.GetSubnet(), "rule.source.subnet"),
			"vlan":                vlans,
		})
		diags.Append(d...)
		out.Source = o
	}

	if dst := r.GetDestination(); dst != nil {
		o, d := types.ObjectValue(SocketBypassDestinationAttrTypes, map[string]attr.Value{
			"application":     oldidname.ParseNameIDList(ctx, dst.GetApplication(), "rule.destination.application"),
			"custom_app":      oldidname.ParseNameIDList(ctx, dst.GetCustomApp(), "rule.destination.custom_app"),
			"domain":          convert.ParseList(ctx, types.StringType, dst.GetDomain(), "rule.destination.domain"),
			"fqdn":            convert.ParseList(ctx, types.StringType, dst.GetFqdn(), "rule.destination.fqdn"),
			"global_ip_range": oldidname.ParseNameIDList(ctx, dst.GetGlobalIPRange(), "rule.destination.global_ip_range"),
			"ip":              convert.ParseList(ctx, types.StringType, dst.GetIP(), "rule.destination.ip"),
			"ip_range":        sbRangeList(dst.GetIPRange(), &diags),
			"subnet":          convert.ParseList(ctx, types.StringType, dst.GetSubnet(), "rule.destination.subnet"),
		})
		diags.Append(d...)
		out.Destination = o
	}

	if svc := r.GetService(); svc != nil && (len(svc.GetSimple()) > 0 || len(svc.GetCustom()) > 0) {
		simple := []string{}
		for _, s := range svc.GetSimple() {
			if s == nil {
				continue
			}
			if s.GetName() != nil {
				simple = append(simple, string(*s.GetName()))
			}
		}
		customElems := []attr.Value{}
		for _, c := range svc.GetCustom() {
			if c == nil {
				continue
			}
			ports := []string{}
			for _, p := range c.GetPort() {
				ports = append(ports, string(p))
			}
			prObj := types.ObjectNull(SocketBypassFromToAttrTypes)
			if pr := c.GetPortRange(); pr != nil {
				prObj = sbFromToObj(sbPortStr(pr.GetFrom()), sbPortStr(pr.GetTo()), &diags)
			}
			proto := ""
			if c.GetProtocol() != nil {
				proto = string(*c.GetProtocol())
			}
			o, d := types.ObjectValue(SocketBypassCustomServiceAttrTypes, map[string]attr.Value{
				"port":       convert.ParseList(ctx, types.StringType, ports, "rule.service.custom.port"),
				"port_range": prObj,
				"protocol":   types.StringValue(proto),
			})
			diags.Append(d...)
			customElems = append(customElems, o)
		}
		customList := types.ListNull(types.ObjectType{AttrTypes: SocketBypassCustomServiceAttrTypes})
		if len(customElems) > 0 {
			lst, d := types.ListValue(types.ObjectType{AttrTypes: SocketBypassCustomServiceAttrTypes}, customElems)
			diags.Append(d...)
			customList = lst
		}
		o, d := types.ObjectValue(SocketBypassServiceAttrTypes, map[string]attr.Value{
			"simple": convert.ParseList(ctx, types.StringType, simple, "rule.service.simple"),
			"custom": customList,
		})
		diags.Append(d...)
		out.Service = o
	}

	// Exceptions: preserved from plan/state by the resource layer (the read
	// response nests the same selector shapes; plan-authoritative handling
	// keeps the implementation simple and avoids order-normalization churn).

	return out, diags
}
