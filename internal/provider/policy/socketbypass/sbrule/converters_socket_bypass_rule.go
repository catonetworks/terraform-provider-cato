package sbrule

import (
	"context"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/cato-go-sdk/scalars"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

func buildCreateObject(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SocketBypassAddRuleDataInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule must be known before applying.")
		return nil
	}
	var model RuleModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.action", model.Action, true},
		knownValue{"rule.description", model.Description, true},
		knownValue{"rule.destination", model.Destination, true},
		knownValue{"rule.enabled", model.Enabled, true},
		knownValue{"rule.exception", model.Exception, true},
		knownValue{"rule.name", model.Name, true},
		knownValue{"rule.service", model.Service, true},
		knownValue{"rule.site", model.Site, true},
		knownValue{"rule.source", model.Source, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SocketBypassAddRuleDataInput{}
	input.Action = buildCreateObjectAction(ctx, model.Action, diags)
	input.Description = model.Description.ValueString()
	input.Destination = buildCreateObjectDestination(ctx, model.Destination, diags)
	input.Enabled = model.Enabled.ValueBool()
	input.Exception = make([]*cato_models.SocketBypassExceptionInput, 0, len(model.Exception.Elements()))
	for _, item := range model.Exception.Elements() {
		input.Exception = append(input.Exception, buildCreateObjectException(ctx, item.(types.Object), diags))
	}
	input.Name = model.Name.ValueString()
	input.Service = buildCreateObjectService(ctx, model.Service, diags)
	input.Site = buildCreateObjectSite(ctx, model.Site, diags)
	input.Source = buildCreateObjectSource(ctx, model.Source, diags)
	return input
}

func buildCreateObjectAction(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SocketBypassActionInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.action must be known before applying.")
		return nil
	}
	var model RuleActionModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.action.event", model.Event, true},
		knownValue{"rule.action.port", model.Port, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SocketBypassActionInput{}
	input.Event = buildCreateObjectActionEvent(ctx, model.Event, diags)
	input.Port = cato_models.BypassSocketPortType(model.Port.ValueString())
	return input
}

func buildCreateObjectActionEvent(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.PolicyRuleTrackingEventInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.action.event must be known before applying.")
		return nil
	}
	var model RuleActionEventModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.action.event.enabled", model.Enabled, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.PolicyRuleTrackingEventInput{}
	input.Enabled = model.Enabled.ValueBool()
	return input
}

func buildCreateObjectDestination(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SocketBypassDestinationInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.destination must be known before applying.")
		return nil
	}
	var model RuleDestinationModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.destination.application", model.Application, true},
		knownValue{"rule.destination.custom_app", model.CustomApp, true},
		knownValue{"rule.destination.domain", model.Domain, true},
		knownValue{"rule.destination.fqdn", model.Fqdn, true},
		knownValue{"rule.destination.global_ip_range", model.GlobalIPRange, true},
		knownValue{"rule.destination.ip", model.IP, true},
		knownValue{"rule.destination.ip_range", model.IPRange, true},
		knownValue{"rule.destination.subnet", model.Subnet, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SocketBypassDestinationInput{}
	input.Application = idname.PrepareIDNameSet[cato_models.ApplicationRefInput](ctx, model.Application, diags)
	input.CustomApp = idname.PrepareIDNameSet[cato_models.CustomApplicationRefInput](ctx, model.CustomApp, diags)
	input.Domain = convert.PrepareStrings[string](ctx, model.Domain, diags)
	if input.Domain == nil {
		input.Domain = []string{}
	}
	input.Fqdn = convert.PrepareStrings[string](ctx, model.Fqdn, diags)
	if input.Fqdn == nil {
		input.Fqdn = []string{}
	}
	input.GlobalIPRange = idname.PrepareIDNameSet[cato_models.GlobalIPRangeRefInput](ctx, model.GlobalIPRange, diags)
	input.IP = convert.PrepareStrings[string](ctx, model.IP, diags)
	if input.IP == nil {
		input.IP = []string{}
	}
	input.IPRange = make([]*cato_models.IPAddressRangeInput, 0, len(model.IPRange.Elements()))
	for _, item := range model.IPRange.Elements() {
		input.IPRange = append(input.IPRange, buildCreateObjectDestinationIPRange(ctx, item.(types.Object), diags))
	}
	input.Subnet = convert.PrepareStrings[string](ctx, model.Subnet, diags)
	if input.Subnet == nil {
		input.Subnet = []string{}
	}
	return input
}

func buildCreateObjectDestinationIPRange(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.IPAddressRangeInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.destination.ip_range must be known before applying.")
		return nil
	}
	var model RuleDestinationIPRangeModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.destination.ip_range.from", model.From, true},
		knownValue{"rule.destination.ip_range.to", model.To, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.IPAddressRangeInput{}
	input.From = model.From.ValueString()
	input.To = model.To.ValueString()
	return input
}

func buildCreateObjectException(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SocketBypassExceptionInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.exception must be known before applying.")
		return nil
	}
	var model RuleExceptionModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.exception.destination", model.Destination, true},
		knownValue{"rule.exception.name", model.Name, true},
		knownValue{"rule.exception.service", model.Service, true},
		knownValue{"rule.exception.site", model.Site, true},
		knownValue{"rule.exception.source", model.Source, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SocketBypassExceptionInput{}
	input.Destination = buildCreateObjectExceptionDestination(ctx, model.Destination, diags)
	input.Name = model.Name.ValueString()
	input.Service = buildCreateObjectExceptionService(ctx, model.Service, diags)
	input.Site = buildCreateObjectExceptionSite(ctx, model.Site, diags)
	input.Source = buildCreateObjectExceptionSource(ctx, model.Source, diags)
	return input
}

func buildCreateObjectExceptionDestination(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SocketBypassDestinationInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.exception.destination must be known before applying.")
		return nil
	}
	var model RuleExceptionDestinationModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.exception.destination.application", model.Application, true},
		knownValue{"rule.exception.destination.custom_app", model.CustomApp, true},
		knownValue{"rule.exception.destination.domain", model.Domain, true},
		knownValue{"rule.exception.destination.fqdn", model.Fqdn, true},
		knownValue{"rule.exception.destination.global_ip_range", model.GlobalIPRange, true},
		knownValue{"rule.exception.destination.ip", model.IP, true},
		knownValue{"rule.exception.destination.ip_range", model.IPRange, true},
		knownValue{"rule.exception.destination.subnet", model.Subnet, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SocketBypassDestinationInput{}
	input.Application = idname.PrepareIDNameSet[cato_models.ApplicationRefInput](ctx, model.Application, diags)
	input.CustomApp = idname.PrepareIDNameSet[cato_models.CustomApplicationRefInput](ctx, model.CustomApp, diags)
	input.Domain = convert.PrepareStrings[string](ctx, model.Domain, diags)
	if input.Domain == nil {
		input.Domain = []string{}
	}
	input.Fqdn = convert.PrepareStrings[string](ctx, model.Fqdn, diags)
	if input.Fqdn == nil {
		input.Fqdn = []string{}
	}
	input.GlobalIPRange = idname.PrepareIDNameSet[cato_models.GlobalIPRangeRefInput](ctx, model.GlobalIPRange, diags)
	input.IP = convert.PrepareStrings[string](ctx, model.IP, diags)
	if input.IP == nil {
		input.IP = []string{}
	}
	input.IPRange = make([]*cato_models.IPAddressRangeInput, 0, len(model.IPRange.Elements()))
	for _, item := range model.IPRange.Elements() {
		input.IPRange = append(input.IPRange, buildCreateObjectExceptionDestinationIPRange(ctx, item.(types.Object), diags))
	}
	input.Subnet = convert.PrepareStrings[string](ctx, model.Subnet, diags)
	if input.Subnet == nil {
		input.Subnet = []string{}
	}
	return input
}

func buildCreateObjectExceptionDestinationIPRange(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.IPAddressRangeInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.exception.destination.ip_range must be known before applying.")
		return nil
	}
	var model RuleExceptionDestinationIPRangeModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.exception.destination.ip_range.from", model.From, true},
		knownValue{"rule.exception.destination.ip_range.to", model.To, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.IPAddressRangeInput{}
	input.From = model.From.ValueString()
	input.To = model.To.ValueString()
	return input
}

func buildCreateObjectExceptionService(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SocketBypassServiceInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.exception.service must be known before applying.")
		return nil
	}
	var model RuleExceptionServiceModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.exception.service.custom", model.Custom, true},
		knownValue{"rule.exception.service.simple", model.Simple, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SocketBypassServiceInput{}
	input.Custom = make([]*cato_models.CustomServiceInput, 0, len(model.Custom.Elements()))
	for _, item := range model.Custom.Elements() {
		input.Custom = append(input.Custom, buildCreateObjectExceptionServiceCustom(ctx, item.(types.Object), diags))
	}
	input.Simple = make([]*cato_models.SimpleServiceInput, 0, len(model.Simple.Elements()))
	for _, item := range model.Simple.Elements() {
		input.Simple = append(input.Simple, buildCreateObjectExceptionServiceSimple(ctx, item.(types.Object), diags))
	}
	return input
}

func buildCreateObjectExceptionServiceCustom(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.CustomServiceInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.exception.service.custom must be known before applying.")
		return nil
	}
	var model RuleExceptionServiceCustomModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.exception.service.custom.port", model.Port, false},
		knownValue{"rule.exception.service.custom.port_range", model.PortRange, false},
		knownValue{"rule.exception.service.custom.protocol", model.Protocol, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.CustomServiceInput{}
	input.Port = convert.PrepareStrings[scalars.Port](ctx, model.Port, diags)
	if input.Port == nil {
		input.Port = []scalars.Port{}
	}
	input.PortRange = buildCreateObjectExceptionServiceCustomPortRange(ctx, model.PortRange, diags)
	input.Protocol = cato_models.IPProtocol(model.Protocol.ValueString())
	return input
}

func buildCreateObjectExceptionServiceCustomPortRange(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.PortRangeInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.exception.service.custom.port_range must be known before applying.")
		return nil
	}
	var model RuleExceptionServiceCustomPortRangeModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.exception.service.custom.port_range.from", model.From, true},
		knownValue{"rule.exception.service.custom.port_range.to", model.To, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.PortRangeInput{}
	input.From = scalars.Port(model.From.ValueString())
	input.To = scalars.Port(model.To.ValueString())
	return input
}

func buildCreateObjectExceptionServiceSimple(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SimpleServiceInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.exception.service.simple must be known before applying.")
		return nil
	}
	var model RuleExceptionServiceSimpleModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.exception.service.simple.name", model.Name, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SimpleServiceInput{}
	input.Name = cato_models.SimpleServiceType(model.Name.ValueString())
	return input
}

func buildCreateObjectExceptionSite(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SocketBypassSiteInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.exception.site must be known before applying.")
		return nil
	}
	var model RuleExceptionSiteModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.exception.site.group", model.Group, true},
		knownValue{"rule.exception.site.site", model.Site, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SocketBypassSiteInput{}
	input.Group = idname.PrepareIDNameSet[cato_models.GroupRefInput](ctx, model.Group, diags)
	input.Site = idname.PrepareIDNameSet[cato_models.SiteRefInput](ctx, model.Site, diags)
	return input
}

func buildCreateObjectExceptionSource(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SocketBypassSourceInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.exception.source must be known before applying.")
		return nil
	}
	var model RuleExceptionSourceModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.exception.source.floating_subnet", model.FloatingSubnet, true},
		knownValue{"rule.exception.source.global_ip_range", model.GlobalIPRange, true},
		knownValue{"rule.exception.source.group", model.Group, true},
		knownValue{"rule.exception.source.host", model.Host, true},
		knownValue{"rule.exception.source.ip", model.IP, true},
		knownValue{"rule.exception.source.ip_range", model.IPRange, true},
		knownValue{"rule.exception.source.network_interface", model.NetworkInterface, true},
		knownValue{"rule.exception.source.site_network_subnet", model.SiteNetworkSubnet, true},
		knownValue{"rule.exception.source.subnet", model.Subnet, true},
		knownValue{"rule.exception.source.vlan", model.Vlan, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SocketBypassSourceInput{}
	input.FloatingSubnet = idname.PrepareIDNameSet[cato_models.FloatingSubnetRefInput](ctx, model.FloatingSubnet, diags)
	input.GlobalIPRange = idname.PrepareIDNameSet[cato_models.GlobalIPRangeRefInput](ctx, model.GlobalIPRange, diags)
	input.Group = idname.PrepareIDNameSet[cato_models.GroupRefInput](ctx, model.Group, diags)
	input.Host = idname.PrepareIDNameSet[cato_models.HostRefInput](ctx, model.Host, diags)
	input.IP = convert.PrepareStrings[string](ctx, model.IP, diags)
	if input.IP == nil {
		input.IP = []string{}
	}
	input.IPRange = make([]*cato_models.IPAddressRangeInput, 0, len(model.IPRange.Elements()))
	for _, item := range model.IPRange.Elements() {
		input.IPRange = append(input.IPRange, buildCreateObjectExceptionSourceIPRange(ctx, item.(types.Object), diags))
	}
	input.NetworkInterface = idname.PrepareIDNameSet[cato_models.NetworkInterfaceRefInput](ctx, model.NetworkInterface, diags)
	input.SiteNetworkSubnet = idname.PrepareIDNameSet[cato_models.SiteNetworkSubnetRefInput](ctx, model.SiteNetworkSubnet, diags)
	input.Subnet = convert.PrepareStrings[string](ctx, model.Subnet, diags)
	if input.Subnet == nil {
		input.Subnet = []string{}
	}
	input.Vlan = buildVLANs(ctx, model.Vlan, diags)
	if input.Vlan == nil {
		input.Vlan = []scalars.Vlan{}
	}
	return input
}

func buildCreateObjectExceptionSourceIPRange(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.IPAddressRangeInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.exception.source.ip_range must be known before applying.")
		return nil
	}
	var model RuleExceptionSourceIPRangeModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.exception.source.ip_range.from", model.From, true},
		knownValue{"rule.exception.source.ip_range.to", model.To, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.IPAddressRangeInput{}
	input.From = model.From.ValueString()
	input.To = model.To.ValueString()
	return input
}

func buildCreateObjectService(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SocketBypassServiceInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.service must be known before applying.")
		return nil
	}
	var model RuleServiceModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.service.custom", model.Custom, true},
		knownValue{"rule.service.simple", model.Simple, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SocketBypassServiceInput{}
	input.Custom = make([]*cato_models.CustomServiceInput, 0, len(model.Custom.Elements()))
	for _, item := range model.Custom.Elements() {
		input.Custom = append(input.Custom, buildCreateObjectServiceCustom(ctx, item.(types.Object), diags))
	}
	input.Simple = make([]*cato_models.SimpleServiceInput, 0, len(model.Simple.Elements()))
	for _, item := range model.Simple.Elements() {
		input.Simple = append(input.Simple, buildCreateObjectServiceSimple(ctx, item.(types.Object), diags))
	}
	return input
}

func buildCreateObjectServiceCustom(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.CustomServiceInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.service.custom must be known before applying.")
		return nil
	}
	var model RuleServiceCustomModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.service.custom.port", model.Port, false},
		knownValue{"rule.service.custom.port_range", model.PortRange, false},
		knownValue{"rule.service.custom.protocol", model.Protocol, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.CustomServiceInput{}
	input.Port = convert.PrepareStrings[scalars.Port](ctx, model.Port, diags)
	if input.Port == nil {
		input.Port = []scalars.Port{}
	}
	input.PortRange = buildCreateObjectServiceCustomPortRange(ctx, model.PortRange, diags)
	input.Protocol = cato_models.IPProtocol(model.Protocol.ValueString())
	return input
}

func buildCreateObjectServiceCustomPortRange(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.PortRangeInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.service.custom.port_range must be known before applying.")
		return nil
	}
	var model RuleServiceCustomPortRangeModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.service.custom.port_range.from", model.From, true},
		knownValue{"rule.service.custom.port_range.to", model.To, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.PortRangeInput{}
	input.From = scalars.Port(model.From.ValueString())
	input.To = scalars.Port(model.To.ValueString())
	return input
}

func buildCreateObjectServiceSimple(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SimpleServiceInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.service.simple must be known before applying.")
		return nil
	}
	var model RuleServiceSimpleModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.service.simple.name", model.Name, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SimpleServiceInput{}
	input.Name = cato_models.SimpleServiceType(model.Name.ValueString())
	return input
}

func buildCreateObjectSite(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SocketBypassSiteInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.site must be known before applying.")
		return nil
	}
	var model RuleSiteModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.site.group", model.Group, true},
		knownValue{"rule.site.site", model.Site, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SocketBypassSiteInput{}
	input.Group = idname.PrepareIDNameSet[cato_models.GroupRefInput](ctx, model.Group, diags)
	input.Site = idname.PrepareIDNameSet[cato_models.SiteRefInput](ctx, model.Site, diags)
	return input
}

func buildCreateObjectSource(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SocketBypassSourceInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.source must be known before applying.")
		return nil
	}
	var model RuleSourceModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.source.floating_subnet", model.FloatingSubnet, true},
		knownValue{"rule.source.global_ip_range", model.GlobalIPRange, true},
		knownValue{"rule.source.group", model.Group, true},
		knownValue{"rule.source.host", model.Host, true},
		knownValue{"rule.source.ip", model.IP, true},
		knownValue{"rule.source.ip_range", model.IPRange, true},
		knownValue{"rule.source.network_interface", model.NetworkInterface, true},
		knownValue{"rule.source.site_network_subnet", model.SiteNetworkSubnet, true},
		knownValue{"rule.source.subnet", model.Subnet, true},
		knownValue{"rule.source.vlan", model.Vlan, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SocketBypassSourceInput{}
	input.FloatingSubnet = idname.PrepareIDNameSet[cato_models.FloatingSubnetRefInput](ctx, model.FloatingSubnet, diags)
	input.GlobalIPRange = idname.PrepareIDNameSet[cato_models.GlobalIPRangeRefInput](ctx, model.GlobalIPRange, diags)
	input.Group = idname.PrepareIDNameSet[cato_models.GroupRefInput](ctx, model.Group, diags)
	input.Host = idname.PrepareIDNameSet[cato_models.HostRefInput](ctx, model.Host, diags)
	input.IP = convert.PrepareStrings[string](ctx, model.IP, diags)
	if input.IP == nil {
		input.IP = []string{}
	}
	input.IPRange = make([]*cato_models.IPAddressRangeInput, 0, len(model.IPRange.Elements()))
	for _, item := range model.IPRange.Elements() {
		input.IPRange = append(input.IPRange, buildCreateObjectSourceIPRange(ctx, item.(types.Object), diags))
	}
	input.NetworkInterface = idname.PrepareIDNameSet[cato_models.NetworkInterfaceRefInput](ctx, model.NetworkInterface, diags)
	input.SiteNetworkSubnet = idname.PrepareIDNameSet[cato_models.SiteNetworkSubnetRefInput](ctx, model.SiteNetworkSubnet, diags)
	input.Subnet = convert.PrepareStrings[string](ctx, model.Subnet, diags)
	if input.Subnet == nil {
		input.Subnet = []string{}
	}
	input.Vlan = buildVLANs(ctx, model.Vlan, diags)
	if input.Vlan == nil {
		input.Vlan = []scalars.Vlan{}
	}
	return input
}

func buildCreateObjectSourceIPRange(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.IPAddressRangeInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.source.ip_range must be known before applying.")
		return nil
	}
	var model RuleSourceIPRangeModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.source.ip_range.from", model.From, true},
		knownValue{"rule.source.ip_range.to", model.To, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.IPAddressRangeInput{}
	input.From = model.From.ValueString()
	input.To = model.To.ValueString()
	return input
}

func buildUpdateObject(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SocketBypassUpdateRuleDataInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule must be known before applying.")
		return nil
	}
	var model RuleModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.action", model.Action, true},
		knownValue{"rule.description", model.Description, true},
		knownValue{"rule.destination", model.Destination, true},
		knownValue{"rule.enabled", model.Enabled, true},
		knownValue{"rule.exception", model.Exception, true},
		knownValue{"rule.name", model.Name, true},
		knownValue{"rule.service", model.Service, true},
		knownValue{"rule.site", model.Site, true},
		knownValue{"rule.source", model.Source, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SocketBypassUpdateRuleDataInput{}
	input.Action = buildUpdateObjectAction(ctx, model.Action, diags)
	vDescription := model.Description.ValueString()
	input.Description = &vDescription
	input.Destination = buildUpdateObjectDestination(ctx, model.Destination, diags)
	vEnabled := model.Enabled.ValueBool()
	input.Enabled = &vEnabled
	input.Exception = make([]*cato_models.SocketBypassExceptionInput, 0, len(model.Exception.Elements()))
	for _, item := range model.Exception.Elements() {
		input.Exception = append(input.Exception, buildUpdateObjectException(ctx, item.(types.Object), diags))
	}
	vName := model.Name.ValueString()
	input.Name = &vName
	input.Service = buildUpdateObjectService(ctx, model.Service, diags)
	input.Site = buildUpdateObjectSite(ctx, model.Site, diags)
	input.Source = buildUpdateObjectSource(ctx, model.Source, diags)
	return input
}

func buildUpdateObjectAction(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SocketBypassActionUpdateInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.action must be known before applying.")
		return nil
	}
	var model RuleActionModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.action.event", model.Event, true},
		knownValue{"rule.action.port", model.Port, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SocketBypassActionUpdateInput{}
	input.Event = buildUpdateObjectActionEvent(ctx, model.Event, diags)
	vPort := cato_models.BypassSocketPortType(model.Port.ValueString())
	input.Port = &vPort
	return input
}

func buildUpdateObjectActionEvent(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.PolicyRuleTrackingEventUpdateInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.action.event must be known before applying.")
		return nil
	}
	var model RuleActionEventModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.action.event.enabled", model.Enabled, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.PolicyRuleTrackingEventUpdateInput{}
	vEnabled := model.Enabled.ValueBool()
	input.Enabled = &vEnabled
	return input
}

func buildUpdateObjectDestination(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SocketBypassDestinationUpdateInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.destination must be known before applying.")
		return nil
	}
	var model RuleDestinationModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.destination.application", model.Application, true},
		knownValue{"rule.destination.custom_app", model.CustomApp, true},
		knownValue{"rule.destination.domain", model.Domain, true},
		knownValue{"rule.destination.fqdn", model.Fqdn, true},
		knownValue{"rule.destination.global_ip_range", model.GlobalIPRange, true},
		knownValue{"rule.destination.ip", model.IP, true},
		knownValue{"rule.destination.ip_range", model.IPRange, true},
		knownValue{"rule.destination.subnet", model.Subnet, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SocketBypassDestinationUpdateInput{}
	input.Application = idname.PrepareIDNameSet[cato_models.ApplicationRefInput](ctx, model.Application, diags)
	input.CustomApp = idname.PrepareIDNameSet[cato_models.CustomApplicationRefInput](ctx, model.CustomApp, diags)
	input.Domain = convert.PrepareStrings[string](ctx, model.Domain, diags)
	if input.Domain == nil {
		input.Domain = []string{}
	}
	input.Fqdn = convert.PrepareStrings[string](ctx, model.Fqdn, diags)
	if input.Fqdn == nil {
		input.Fqdn = []string{}
	}
	input.GlobalIPRange = idname.PrepareIDNameSet[cato_models.GlobalIPRangeRefInput](ctx, model.GlobalIPRange, diags)
	input.IP = convert.PrepareStrings[string](ctx, model.IP, diags)
	if input.IP == nil {
		input.IP = []string{}
	}
	input.IPRange = make([]*cato_models.IPAddressRangeInput, 0, len(model.IPRange.Elements()))
	for _, item := range model.IPRange.Elements() {
		input.IPRange = append(input.IPRange, buildUpdateObjectDestinationIPRange(ctx, item.(types.Object), diags))
	}
	input.Subnet = convert.PrepareStrings[string](ctx, model.Subnet, diags)
	if input.Subnet == nil {
		input.Subnet = []string{}
	}
	return input
}

func buildUpdateObjectDestinationIPRange(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.IPAddressRangeInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.destination.ip_range must be known before applying.")
		return nil
	}
	var model RuleDestinationIPRangeModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.destination.ip_range.from", model.From, true},
		knownValue{"rule.destination.ip_range.to", model.To, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.IPAddressRangeInput{}
	input.From = model.From.ValueString()
	input.To = model.To.ValueString()
	return input
}

func buildUpdateObjectException(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SocketBypassExceptionInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.exception must be known before applying.")
		return nil
	}
	var model RuleExceptionModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.exception.destination", model.Destination, true},
		knownValue{"rule.exception.name", model.Name, true},
		knownValue{"rule.exception.service", model.Service, true},
		knownValue{"rule.exception.site", model.Site, true},
		knownValue{"rule.exception.source", model.Source, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SocketBypassExceptionInput{}
	input.Destination = buildUpdateObjectExceptionDestination(ctx, model.Destination, diags)
	input.Name = model.Name.ValueString()
	input.Service = buildUpdateObjectExceptionService(ctx, model.Service, diags)
	input.Site = buildUpdateObjectExceptionSite(ctx, model.Site, diags)
	input.Source = buildUpdateObjectExceptionSource(ctx, model.Source, diags)
	return input
}

func buildUpdateObjectExceptionDestination(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SocketBypassDestinationInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.exception.destination must be known before applying.")
		return nil
	}
	var model RuleExceptionDestinationModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.exception.destination.application", model.Application, true},
		knownValue{"rule.exception.destination.custom_app", model.CustomApp, true},
		knownValue{"rule.exception.destination.domain", model.Domain, true},
		knownValue{"rule.exception.destination.fqdn", model.Fqdn, true},
		knownValue{"rule.exception.destination.global_ip_range", model.GlobalIPRange, true},
		knownValue{"rule.exception.destination.ip", model.IP, true},
		knownValue{"rule.exception.destination.ip_range", model.IPRange, true},
		knownValue{"rule.exception.destination.subnet", model.Subnet, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SocketBypassDestinationInput{}
	input.Application = idname.PrepareIDNameSet[cato_models.ApplicationRefInput](ctx, model.Application, diags)
	input.CustomApp = idname.PrepareIDNameSet[cato_models.CustomApplicationRefInput](ctx, model.CustomApp, diags)
	input.Domain = convert.PrepareStrings[string](ctx, model.Domain, diags)
	if input.Domain == nil {
		input.Domain = []string{}
	}
	input.Fqdn = convert.PrepareStrings[string](ctx, model.Fqdn, diags)
	if input.Fqdn == nil {
		input.Fqdn = []string{}
	}
	input.GlobalIPRange = idname.PrepareIDNameSet[cato_models.GlobalIPRangeRefInput](ctx, model.GlobalIPRange, diags)
	input.IP = convert.PrepareStrings[string](ctx, model.IP, diags)
	if input.IP == nil {
		input.IP = []string{}
	}
	input.IPRange = make([]*cato_models.IPAddressRangeInput, 0, len(model.IPRange.Elements()))
	for _, item := range model.IPRange.Elements() {
		input.IPRange = append(input.IPRange, buildUpdateObjectExceptionDestinationIPRange(ctx, item.(types.Object), diags))
	}
	input.Subnet = convert.PrepareStrings[string](ctx, model.Subnet, diags)
	if input.Subnet == nil {
		input.Subnet = []string{}
	}
	return input
}

func buildUpdateObjectExceptionDestinationIPRange(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.IPAddressRangeInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.exception.destination.ip_range must be known before applying.")
		return nil
	}
	var model RuleExceptionDestinationIPRangeModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.exception.destination.ip_range.from", model.From, true},
		knownValue{"rule.exception.destination.ip_range.to", model.To, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.IPAddressRangeInput{}
	input.From = model.From.ValueString()
	input.To = model.To.ValueString()
	return input
}

func buildUpdateObjectExceptionService(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SocketBypassServiceInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.exception.service must be known before applying.")
		return nil
	}
	var model RuleExceptionServiceModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.exception.service.custom", model.Custom, true},
		knownValue{"rule.exception.service.simple", model.Simple, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SocketBypassServiceInput{}
	input.Custom = make([]*cato_models.CustomServiceInput, 0, len(model.Custom.Elements()))
	for _, item := range model.Custom.Elements() {
		input.Custom = append(input.Custom, buildUpdateObjectExceptionServiceCustom(ctx, item.(types.Object), diags))
	}
	input.Simple = make([]*cato_models.SimpleServiceInput, 0, len(model.Simple.Elements()))
	for _, item := range model.Simple.Elements() {
		input.Simple = append(input.Simple, buildUpdateObjectExceptionServiceSimple(ctx, item.(types.Object), diags))
	}
	return input
}

func buildUpdateObjectExceptionServiceCustom(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.CustomServiceInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.exception.service.custom must be known before applying.")
		return nil
	}
	var model RuleExceptionServiceCustomModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.exception.service.custom.port", model.Port, false},
		knownValue{"rule.exception.service.custom.port_range", model.PortRange, false},
		knownValue{"rule.exception.service.custom.protocol", model.Protocol, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.CustomServiceInput{}
	input.Port = convert.PrepareStrings[scalars.Port](ctx, model.Port, diags)
	if input.Port == nil {
		input.Port = []scalars.Port{}
	}
	input.PortRange = buildUpdateObjectExceptionServiceCustomPortRange(ctx, model.PortRange, diags)
	input.Protocol = cato_models.IPProtocol(model.Protocol.ValueString())
	return input
}

func buildUpdateObjectExceptionServiceCustomPortRange(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.PortRangeInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.exception.service.custom.port_range must be known before applying.")
		return nil
	}
	var model RuleExceptionServiceCustomPortRangeModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.exception.service.custom.port_range.from", model.From, true},
		knownValue{"rule.exception.service.custom.port_range.to", model.To, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.PortRangeInput{}
	input.From = scalars.Port(model.From.ValueString())
	input.To = scalars.Port(model.To.ValueString())
	return input
}

func buildUpdateObjectExceptionServiceSimple(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SimpleServiceInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.exception.service.simple must be known before applying.")
		return nil
	}
	var model RuleExceptionServiceSimpleModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.exception.service.simple.name", model.Name, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SimpleServiceInput{}
	input.Name = cato_models.SimpleServiceType(model.Name.ValueString())
	return input
}

func buildUpdateObjectExceptionSite(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SocketBypassSiteInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.exception.site must be known before applying.")
		return nil
	}
	var model RuleExceptionSiteModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.exception.site.group", model.Group, true},
		knownValue{"rule.exception.site.site", model.Site, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SocketBypassSiteInput{}
	input.Group = idname.PrepareIDNameSet[cato_models.GroupRefInput](ctx, model.Group, diags)
	input.Site = idname.PrepareIDNameSet[cato_models.SiteRefInput](ctx, model.Site, diags)
	return input
}

func buildUpdateObjectExceptionSource(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SocketBypassSourceInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.exception.source must be known before applying.")
		return nil
	}
	var model RuleExceptionSourceModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.exception.source.floating_subnet", model.FloatingSubnet, true},
		knownValue{"rule.exception.source.global_ip_range", model.GlobalIPRange, true},
		knownValue{"rule.exception.source.group", model.Group, true},
		knownValue{"rule.exception.source.host", model.Host, true},
		knownValue{"rule.exception.source.ip", model.IP, true},
		knownValue{"rule.exception.source.ip_range", model.IPRange, true},
		knownValue{"rule.exception.source.network_interface", model.NetworkInterface, true},
		knownValue{"rule.exception.source.site_network_subnet", model.SiteNetworkSubnet, true},
		knownValue{"rule.exception.source.subnet", model.Subnet, true},
		knownValue{"rule.exception.source.vlan", model.Vlan, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SocketBypassSourceInput{}
	input.FloatingSubnet = idname.PrepareIDNameSet[cato_models.FloatingSubnetRefInput](ctx, model.FloatingSubnet, diags)
	input.GlobalIPRange = idname.PrepareIDNameSet[cato_models.GlobalIPRangeRefInput](ctx, model.GlobalIPRange, diags)
	input.Group = idname.PrepareIDNameSet[cato_models.GroupRefInput](ctx, model.Group, diags)
	input.Host = idname.PrepareIDNameSet[cato_models.HostRefInput](ctx, model.Host, diags)
	input.IP = convert.PrepareStrings[string](ctx, model.IP, diags)
	if input.IP == nil {
		input.IP = []string{}
	}
	input.IPRange = make([]*cato_models.IPAddressRangeInput, 0, len(model.IPRange.Elements()))
	for _, item := range model.IPRange.Elements() {
		input.IPRange = append(input.IPRange, buildUpdateObjectExceptionSourceIPRange(ctx, item.(types.Object), diags))
	}
	input.NetworkInterface = idname.PrepareIDNameSet[cato_models.NetworkInterfaceRefInput](ctx, model.NetworkInterface, diags)
	input.SiteNetworkSubnet = idname.PrepareIDNameSet[cato_models.SiteNetworkSubnetRefInput](ctx, model.SiteNetworkSubnet, diags)
	input.Subnet = convert.PrepareStrings[string](ctx, model.Subnet, diags)
	if input.Subnet == nil {
		input.Subnet = []string{}
	}
	input.Vlan = buildVLANs(ctx, model.Vlan, diags)
	if input.Vlan == nil {
		input.Vlan = []scalars.Vlan{}
	}
	return input
}

func buildUpdateObjectExceptionSourceIPRange(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.IPAddressRangeInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.exception.source.ip_range must be known before applying.")
		return nil
	}
	var model RuleExceptionSourceIPRangeModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.exception.source.ip_range.from", model.From, true},
		knownValue{"rule.exception.source.ip_range.to", model.To, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.IPAddressRangeInput{}
	input.From = model.From.ValueString()
	input.To = model.To.ValueString()
	return input
}

func buildUpdateObjectService(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SocketBypassServiceUpdateInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.service must be known before applying.")
		return nil
	}
	var model RuleServiceModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.service.custom", model.Custom, true},
		knownValue{"rule.service.simple", model.Simple, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SocketBypassServiceUpdateInput{}
	input.Custom = make([]*cato_models.CustomServiceInput, 0, len(model.Custom.Elements()))
	for _, item := range model.Custom.Elements() {
		input.Custom = append(input.Custom, buildUpdateObjectServiceCustom(ctx, item.(types.Object), diags))
	}
	input.Simple = make([]*cato_models.SimpleServiceInput, 0, len(model.Simple.Elements()))
	for _, item := range model.Simple.Elements() {
		input.Simple = append(input.Simple, buildUpdateObjectServiceSimple(ctx, item.(types.Object), diags))
	}
	return input
}

func buildUpdateObjectServiceCustom(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.CustomServiceInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.service.custom must be known before applying.")
		return nil
	}
	var model RuleServiceCustomModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.service.custom.port", model.Port, false},
		knownValue{"rule.service.custom.port_range", model.PortRange, false},
		knownValue{"rule.service.custom.protocol", model.Protocol, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.CustomServiceInput{}
	input.Port = convert.PrepareStrings[scalars.Port](ctx, model.Port, diags)
	if input.Port == nil {
		input.Port = []scalars.Port{}
	}
	input.PortRange = buildUpdateObjectServiceCustomPortRange(ctx, model.PortRange, diags)
	input.Protocol = cato_models.IPProtocol(model.Protocol.ValueString())
	return input
}

func buildUpdateObjectServiceCustomPortRange(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.PortRangeInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.service.custom.port_range must be known before applying.")
		return nil
	}
	var model RuleServiceCustomPortRangeModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.service.custom.port_range.from", model.From, true},
		knownValue{"rule.service.custom.port_range.to", model.To, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.PortRangeInput{}
	input.From = scalars.Port(model.From.ValueString())
	input.To = scalars.Port(model.To.ValueString())
	return input
}

func buildUpdateObjectServiceSimple(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SimpleServiceInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.service.simple must be known before applying.")
		return nil
	}
	var model RuleServiceSimpleModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.service.simple.name", model.Name, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SimpleServiceInput{}
	input.Name = cato_models.SimpleServiceType(model.Name.ValueString())
	return input
}

func buildUpdateObjectSite(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SocketBypassSiteUpdateInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.site must be known before applying.")
		return nil
	}
	var model RuleSiteModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.site.group", model.Group, true},
		knownValue{"rule.site.site", model.Site, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SocketBypassSiteUpdateInput{}
	input.Group = idname.PrepareIDNameSet[cato_models.GroupRefInput](ctx, model.Group, diags)
	input.Site = idname.PrepareIDNameSet[cato_models.SiteRefInput](ctx, model.Site, diags)
	return input
}

func buildUpdateObjectSource(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SocketBypassSourceUpdateInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.source must be known before applying.")
		return nil
	}
	var model RuleSourceModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.source.floating_subnet", model.FloatingSubnet, true},
		knownValue{"rule.source.global_ip_range", model.GlobalIPRange, true},
		knownValue{"rule.source.group", model.Group, true},
		knownValue{"rule.source.host", model.Host, true},
		knownValue{"rule.source.ip", model.IP, true},
		knownValue{"rule.source.ip_range", model.IPRange, true},
		knownValue{"rule.source.network_interface", model.NetworkInterface, true},
		knownValue{"rule.source.site_network_subnet", model.SiteNetworkSubnet, true},
		knownValue{"rule.source.subnet", model.Subnet, true},
		knownValue{"rule.source.vlan", model.Vlan, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SocketBypassSourceUpdateInput{}
	input.FloatingSubnet = idname.PrepareIDNameSet[cato_models.FloatingSubnetRefInput](ctx, model.FloatingSubnet, diags)
	input.GlobalIPRange = idname.PrepareIDNameSet[cato_models.GlobalIPRangeRefInput](ctx, model.GlobalIPRange, diags)
	input.Group = idname.PrepareIDNameSet[cato_models.GroupRefInput](ctx, model.Group, diags)
	input.Host = idname.PrepareIDNameSet[cato_models.HostRefInput](ctx, model.Host, diags)
	input.IP = convert.PrepareStrings[string](ctx, model.IP, diags)
	if input.IP == nil {
		input.IP = []string{}
	}
	input.IPRange = make([]*cato_models.IPAddressRangeInput, 0, len(model.IPRange.Elements()))
	for _, item := range model.IPRange.Elements() {
		input.IPRange = append(input.IPRange, buildUpdateObjectSourceIPRange(ctx, item.(types.Object), diags))
	}
	input.NetworkInterface = idname.PrepareIDNameSet[cato_models.NetworkInterfaceRefInput](ctx, model.NetworkInterface, diags)
	input.SiteNetworkSubnet = idname.PrepareIDNameSet[cato_models.SiteNetworkSubnetRefInput](ctx, model.SiteNetworkSubnet, diags)
	input.Subnet = convert.PrepareStrings[string](ctx, model.Subnet, diags)
	if input.Subnet == nil {
		input.Subnet = []string{}
	}
	input.Vlan = buildVLANs(ctx, model.Vlan, diags)
	if input.Vlan == nil {
		input.Vlan = []scalars.Vlan{}
	}
	return input
}

func buildUpdateObjectSourceIPRange(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.IPAddressRangeInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.source.ip_range must be known before applying.")
		return nil
	}
	var model RuleSourceIPRangeModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	validateValues(diags,
		knownValue{"rule.source.ip_range.from", model.From, true},
		knownValue{"rule.source.ip_range.to", model.To, true},
	)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.IPAddressRangeInput{}
	input.From = model.From.ValueString()
	input.To = model.To.ValueString()
	return input
}

type read = cato.PolicySocketBypassPolicy_Policy_SocketBypass_Policy_Rules_Rule

func parseRead(
	ctx context.Context, value *read, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule object.")
		return types.ObjectNull(ruleAttrTypes)
	}
	model := RuleModel{}
	model.Action = parseReadAction(ctx, &value.Action, diags)
	model.Description = types.StringValue(value.Description)
	model.Destination = parseReadDestination(ctx, &value.Destination, diags)
	model.Enabled = types.BoolValue(value.Enabled)
	valuesException := make([]attr.Value, 0, len(value.Exception))
	for _, item := range value.Exception {
		valuesException = append(valuesException, parseReadException(ctx, item, diags))
	}
	var dException diag.Diagnostics
	model.Exception, dException = types.SetValue(types.ObjectType{AttrTypes: ruleExceptionAttrTypes}, valuesException)
	diags.Append(dException...)
	model.ID = types.StringValue(value.ID)
	model.Index = types.Int64Value(value.Index)
	model.Name = types.StringValue(value.Name)
	model.Section = parseReadSection(ctx, &value.Section, diags)
	model.Service = parseReadService(ctx, &value.Service, diags)
	model.Site = parseReadSite(ctx, &value.Site, diags)
	model.Source = parseReadSource(ctx, &value.Source, diags)
	result, d := types.ObjectValueFrom(ctx, ruleAttrTypes, model)
	diags.Append(d...)
	return result
}

type readAction = cato.PolicySocketBypassPolicy_Policy_SocketBypass_Policy_Rules_Rule_Action

func parseReadAction(
	ctx context.Context, value *readAction, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.action object.")
		return types.ObjectNull(ruleActionAttrTypes)
	}
	model := RuleActionModel{}
	model.Event = parseReadActionEvent(ctx, &value.Event, diags)
	model.Port = types.StringValue(string(value.Port))
	result, d := types.ObjectValueFrom(ctx, ruleActionAttrTypes, model)
	diags.Append(d...)
	return result
}

type readActionEvent = cato.PolicySocketBypassPolicy_Policy_SocketBypass_Policy_Rules_Rule_Action_Event

func parseReadActionEvent(
	ctx context.Context, value *readActionEvent, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.action.event object.")
		return types.ObjectNull(ruleActionEventAttrTypes)
	}
	model := RuleActionEventModel{}
	model.Enabled = types.BoolValue(value.Enabled)
	result, d := types.ObjectValueFrom(ctx, ruleActionEventAttrTypes, model)
	diags.Append(d...)
	return result
}

type readDst = cato.PolicySocketBypassPolicy_Policy_SocketBypass_Policy_Rules_Rule_Destination

func parseReadDestination(
	ctx context.Context, value *readDst, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.destination object.")
		return types.ObjectNull(ruleDestinationAttrTypes)
	}
	model := RuleDestinationModel{}
	model.Application = idname.ParseIDNameSet(ctx, value.Application, diags)
	if model.Application.IsNull() {
		model.Application = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.CustomApp = idname.ParseIDNameSet(ctx, value.CustomApp, diags)
	if model.CustomApp.IsNull() {
		model.CustomApp = types.SetValueMust(idname.RefObjectType, nil)
	}
	var dDomain diag.Diagnostics
	model.Domain, dDomain = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Domain...))
	diags.Append(dDomain...)
	var dFqdn diag.Diagnostics
	model.Fqdn, dFqdn = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Fqdn...))
	diags.Append(dFqdn...)
	model.GlobalIPRange = idname.ParseIDNameSet(ctx, value.GlobalIPRange, diags)
	if model.GlobalIPRange.IsNull() {
		model.GlobalIPRange = types.SetValueMust(idname.RefObjectType, nil)
	}
	var dIP diag.Diagnostics
	model.IP, dIP = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.IP...))
	diags.Append(dIP...)
	valuesIPRange := make([]attr.Value, 0, len(value.IPRange))
	for _, item := range value.IPRange {
		valuesIPRange = append(valuesIPRange, parseReadDestinationIPRange(ctx, item, diags))
	}
	var dIPRange diag.Diagnostics
	model.IPRange, dIPRange = types.SetValue(types.ObjectType{AttrTypes: ruleDestinationIPRangeAttrTypes}, valuesIPRange)
	diags.Append(dIPRange...)
	var dSubnet diag.Diagnostics
	model.Subnet, dSubnet = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Subnet...))
	diags.Append(dSubnet...)
	result, d := types.ObjectValueFrom(ctx, ruleDestinationAttrTypes, model)
	diags.Append(d...)
	return result
}

type readDstIPRange = cato.PolicySocketBypassPolicy_Policy_SocketBypass_Policy_Rules_Rule_Destination_IPRange

func parseReadDestinationIPRange(
	ctx context.Context, value *readDstIPRange, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.destination.ip_range object.")
		return types.ObjectNull(ruleDestinationIPRangeAttrTypes)
	}
	model := RuleDestinationIPRangeModel{}
	model.From = types.StringValue(value.From)
	model.To = types.StringValue(value.To)
	result, d := types.ObjectValueFrom(ctx, ruleDestinationIPRangeAttrTypes, model)
	diags.Append(d...)
	return result
}

type readExc = cato.PolicySocketBypassPolicy_Policy_SocketBypass_Policy_Rules_Rule_Exception

func parseReadException(
	ctx context.Context, value *readExc, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.exception object.")
		return types.ObjectNull(ruleExceptionAttrTypes)
	}
	model := RuleExceptionModel{}
	model.Destination = parseReadExceptionDestination(ctx, &value.Destination, diags)
	model.Name = types.StringValue(value.Name)
	model.Service = parseReadExceptionService(ctx, &value.Service, diags)
	model.Site = parseReadExceptionSite(ctx, &value.Site, diags)
	model.Source = parseReadExceptionSource(ctx, &value.Source, diags)
	result, d := types.ObjectValueFrom(ctx, ruleExceptionAttrTypes, model)
	diags.Append(d...)
	return result
}

type readExcDst = cato.PolicySocketBypassPolicy_Policy_SocketBypass_Policy_Rules_Rule_Exception_Destination

func parseReadExceptionDestination(
	ctx context.Context, value *readExcDst, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.exception.destination object.")
		return types.ObjectNull(ruleExceptionDestinationAttrTypes)
	}
	model := RuleExceptionDestinationModel{}
	model.Application = idname.ParseIDNameSet(ctx, value.Application, diags)
	if model.Application.IsNull() {
		model.Application = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.CustomApp = idname.ParseIDNameSet(ctx, value.CustomApp, diags)
	if model.CustomApp.IsNull() {
		model.CustomApp = types.SetValueMust(idname.RefObjectType, nil)
	}
	var dDomain diag.Diagnostics
	model.Domain, dDomain = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Domain...))
	diags.Append(dDomain...)
	var dFqdn diag.Diagnostics
	model.Fqdn, dFqdn = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Fqdn...))
	diags.Append(dFqdn...)
	model.GlobalIPRange = idname.ParseIDNameSet(ctx, value.GlobalIPRange, diags)
	if model.GlobalIPRange.IsNull() {
		model.GlobalIPRange = types.SetValueMust(idname.RefObjectType, nil)
	}
	var dIP diag.Diagnostics
	model.IP, dIP = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.IP...))
	diags.Append(dIP...)
	valuesIPRange := make([]attr.Value, 0, len(value.IPRange))
	for _, item := range value.IPRange {
		valuesIPRange = append(valuesIPRange, parseReadExceptionDestinationIPRange(ctx, item, diags))
	}
	var dIPRange diag.Diagnostics
	model.IPRange, dIPRange = types.SetValue(types.ObjectType{AttrTypes: ruleExceptionDestinationIPRangeAttrTypes}, valuesIPRange)
	diags.Append(dIPRange...)
	var dSubnet diag.Diagnostics
	model.Subnet, dSubnet = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Subnet...))
	diags.Append(dSubnet...)
	result, d := types.ObjectValueFrom(ctx, ruleExceptionDestinationAttrTypes, model)
	diags.Append(d...)
	return result
}

type readExcDstIPRange = cato.PolicySocketBypassPolicy_Policy_SocketBypass_Policy_Rules_Rule_Exception_Destination_IPRange

func parseReadExceptionDestinationIPRange(
	ctx context.Context, value *readExcDstIPRange, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.exception.destination.ip_range object.")
		return types.ObjectNull(ruleExceptionDestinationIPRangeAttrTypes)
	}
	model := RuleExceptionDestinationIPRangeModel{}
	model.From = types.StringValue(value.From)
	model.To = types.StringValue(value.To)
	result, d := types.ObjectValueFrom(ctx, ruleExceptionDestinationIPRangeAttrTypes, model)
	diags.Append(d...)
	return result
}

type readExcSvc = cato.PolicySocketBypassPolicy_Policy_SocketBypass_Policy_Rules_Rule_Exception_Service

func parseReadExceptionService(
	ctx context.Context, value *readExcSvc, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.exception.service object.")
		return types.ObjectNull(ruleExceptionServiceAttrTypes)
	}
	model := RuleExceptionServiceModel{}
	valuesCustom := make([]attr.Value, 0, len(value.Custom))
	for _, item := range value.Custom {
		valuesCustom = append(valuesCustom, parseReadExceptionServiceCustom(ctx, item, diags))
	}
	var dCustom diag.Diagnostics
	model.Custom, dCustom = types.SetValue(types.ObjectType{AttrTypes: ruleExceptionServiceCustomAttrTypes}, valuesCustom)
	diags.Append(dCustom...)
	valuesSimple := make([]attr.Value, 0, len(value.Simple))
	for _, item := range value.Simple {
		valuesSimple = append(valuesSimple, parseReadExceptionServiceSimple(ctx, item, diags))
	}
	var dSimple diag.Diagnostics
	model.Simple, dSimple = types.SetValue(types.ObjectType{AttrTypes: ruleExceptionServiceSimpleAttrTypes}, valuesSimple)
	diags.Append(dSimple...)
	result, d := types.ObjectValueFrom(ctx, ruleExceptionServiceAttrTypes, model)
	diags.Append(d...)
	return result
}

type readExcSvcCustom = cato.PolicySocketBypassPolicy_Policy_SocketBypass_Policy_Rules_Rule_Exception_Service_Custom

func parseReadExceptionServiceCustom(
	ctx context.Context, value *readExcSvcCustom, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.exception.service.custom object.")
		return types.ObjectNull(ruleExceptionServiceCustomAttrTypes)
	}
	model := RuleExceptionServiceCustomModel{}
	model.Port = parsePorts(ctx, value.Port, diags)
	if value.PortRange == nil {
		model.PortRange = types.ObjectNull(ruleExceptionServiceCustomPortRangeAttrTypes)
	} else {
		model.PortRange = parseReadExceptionServiceCustomPortRange(ctx, value.PortRange, diags)
	}
	model.Protocol = types.StringValue(string(value.Protocol))
	result, d := types.ObjectValueFrom(ctx, ruleExceptionServiceCustomAttrTypes, model)
	diags.Append(d...)
	return result
}

type readExcSvcCustomPortRange = cato.PolicySocketBypassPolicy_Policy_SocketBypass_Policy_Rules_Rule_Exception_Service_Custom_PortRange

func parseReadExceptionServiceCustomPortRange(
	ctx context.Context, value *readExcSvcCustomPortRange, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.exception.service.custom.port_range object.")
		return types.ObjectNull(ruleExceptionServiceCustomPortRangeAttrTypes)
	}
	model := RuleExceptionServiceCustomPortRangeModel{}
	model.From = types.StringValue(string(value.From))
	model.To = types.StringValue(string(value.To))
	result, d := types.ObjectValueFrom(ctx, ruleExceptionServiceCustomPortRangeAttrTypes, model)
	diags.Append(d...)
	return result
}

type readExcSvcSimple = cato.PolicySocketBypassPolicy_Policy_SocketBypass_Policy_Rules_Rule_Exception_Service_Simple

func parseReadExceptionServiceSimple(
	ctx context.Context, value *readExcSvcSimple, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.exception.service.simple object.")
		return types.ObjectNull(ruleExceptionServiceSimpleAttrTypes)
	}
	model := RuleExceptionServiceSimpleModel{}
	model.Name = types.StringValue(string(value.Name))
	result, d := types.ObjectValueFrom(ctx, ruleExceptionServiceSimpleAttrTypes, model)
	diags.Append(d...)
	return result
}

type readExcSite = cato.PolicySocketBypassPolicy_Policy_SocketBypass_Policy_Rules_Rule_Exception_Site

func parseReadExceptionSite(
	ctx context.Context, value *readExcSite, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.exception.site object.")
		return types.ObjectNull(ruleExceptionSiteAttrTypes)
	}
	model := RuleExceptionSiteModel{}
	model.Group = idname.ParseIDNameSet(ctx, value.Group, diags)
	if model.Group.IsNull() {
		model.Group = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Site = idname.ParseIDNameSet(ctx, value.Site, diags)
	if model.Site.IsNull() {
		model.Site = types.SetValueMust(idname.RefObjectType, nil)
	}
	result, d := types.ObjectValueFrom(ctx, ruleExceptionSiteAttrTypes, model)
	diags.Append(d...)
	return result
}

type readExcSource = cato.PolicySocketBypassPolicy_Policy_SocketBypass_Policy_Rules_Rule_Exception_Source

func parseReadExceptionSource(
	ctx context.Context, value *readExcSource, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.exception.source object.")
		return types.ObjectNull(ruleExceptionSourceAttrTypes)
	}
	model := RuleExceptionSourceModel{}
	model.FloatingSubnet = idname.ParseIDNameSet(ctx, value.FloatingSubnet, diags)
	if model.FloatingSubnet.IsNull() {
		model.FloatingSubnet = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.GlobalIPRange = idname.ParseIDNameSet(ctx, value.GlobalIPRange, diags)
	if model.GlobalIPRange.IsNull() {
		model.GlobalIPRange = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Group = idname.ParseIDNameSet(ctx, value.Group, diags)
	if model.Group.IsNull() {
		model.Group = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Host = idname.ParseIDNameSet(ctx, value.Host, diags)
	if model.Host.IsNull() {
		model.Host = types.SetValueMust(idname.RefObjectType, nil)
	}
	var dIP diag.Diagnostics
	model.IP, dIP = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.IP...))
	diags.Append(dIP...)
	valuesIPRange := make([]attr.Value, 0, len(value.IPRange))
	for _, item := range value.IPRange {
		valuesIPRange = append(valuesIPRange, parseReadExceptionSourceIPRange(ctx, item, diags))
	}
	var dIPRange diag.Diagnostics
	model.IPRange, dIPRange = types.SetValue(types.ObjectType{AttrTypes: ruleExceptionSourceIPRangeAttrTypes}, valuesIPRange)
	diags.Append(dIPRange...)
	model.NetworkInterface = idname.ParseIDNameSet(ctx, value.NetworkInterface, diags)
	if model.NetworkInterface.IsNull() {
		model.NetworkInterface = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.SiteNetworkSubnet = idname.ParseIDNameSet(ctx, value.SiteNetworkSubnet, diags)
	if model.SiteNetworkSubnet.IsNull() {
		model.SiteNetworkSubnet = types.SetValueMust(idname.RefObjectType, nil)
	}
	var dSubnet diag.Diagnostics
	model.Subnet, dSubnet = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Subnet...))
	diags.Append(dSubnet...)
	model.Vlan = parseVLANs(ctx, value.Vlan, diags)
	result, d := types.ObjectValueFrom(ctx, ruleExceptionSourceAttrTypes, model)
	diags.Append(d...)
	return result
}

type readExcSourceIPRange = cato.PolicySocketBypassPolicy_Policy_SocketBypass_Policy_Rules_Rule_Exception_Source_IPRange

func parseReadExceptionSourceIPRange(
	ctx context.Context, value *readExcSourceIPRange, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.exception.source.ip_range object.")
		return types.ObjectNull(ruleExceptionSourceIPRangeAttrTypes)
	}
	model := RuleExceptionSourceIPRangeModel{}
	model.From = types.StringValue(value.From)
	model.To = types.StringValue(value.To)
	result, d := types.ObjectValueFrom(ctx, ruleExceptionSourceIPRangeAttrTypes, model)
	diags.Append(d...)
	return result
}

type readSection = cato.PolicySocketBypassPolicy_Policy_SocketBypass_Policy_Rules_Rule_Section

func parseReadSection(
	ctx context.Context, value *readSection, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.section object.")
		return types.ObjectNull(ruleSectionAttrTypes)
	}
	return idname.ParseIDName(ctx, *value, diags)
}

type readSvc = cato.PolicySocketBypassPolicy_Policy_SocketBypass_Policy_Rules_Rule_Service

func parseReadService(
	ctx context.Context, value *readSvc, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.service object.")
		return types.ObjectNull(ruleServiceAttrTypes)
	}
	model := RuleServiceModel{}
	valuesCustom := make([]attr.Value, 0, len(value.Custom))
	for _, item := range value.Custom {
		valuesCustom = append(valuesCustom, parseReadServiceCustom(ctx, item, diags))
	}
	var dCustom diag.Diagnostics
	model.Custom, dCustom = types.SetValue(types.ObjectType{AttrTypes: ruleServiceCustomAttrTypes}, valuesCustom)
	diags.Append(dCustom...)
	valuesSimple := make([]attr.Value, 0, len(value.Simple))
	for _, item := range value.Simple {
		valuesSimple = append(valuesSimple, parseReadServiceSimple(ctx, item, diags))
	}
	var dSimple diag.Diagnostics
	model.Simple, dSimple = types.SetValue(types.ObjectType{AttrTypes: ruleServiceSimpleAttrTypes}, valuesSimple)
	diags.Append(dSimple...)
	result, d := types.ObjectValueFrom(ctx, ruleServiceAttrTypes, model)
	diags.Append(d...)
	return result
}

type readSvcCustom = cato.PolicySocketBypassPolicy_Policy_SocketBypass_Policy_Rules_Rule_Service_Custom

func parseReadServiceCustom(
	ctx context.Context, value *readSvcCustom, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.service.custom object.")
		return types.ObjectNull(ruleServiceCustomAttrTypes)
	}
	model := RuleServiceCustomModel{}
	model.Port = parsePorts(ctx, value.Port, diags)
	if value.PortRange == nil {
		model.PortRange = types.ObjectNull(ruleServiceCustomPortRangeAttrTypes)
	} else {
		model.PortRange = parseReadServiceCustomPortRange(ctx, value.PortRange, diags)
	}
	model.Protocol = types.StringValue(string(value.Protocol))
	result, d := types.ObjectValueFrom(ctx, ruleServiceCustomAttrTypes, model)
	diags.Append(d...)
	return result
}

type readSvcCustomPortRange = cato.PolicySocketBypassPolicy_Policy_SocketBypass_Policy_Rules_Rule_Service_Custom_PortRange

func parseReadServiceCustomPortRange(
	ctx context.Context, value *readSvcCustomPortRange, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.service.custom.port_range object.")
		return types.ObjectNull(ruleServiceCustomPortRangeAttrTypes)
	}
	model := RuleServiceCustomPortRangeModel{}
	model.From = types.StringValue(string(value.From))
	model.To = types.StringValue(string(value.To))
	result, d := types.ObjectValueFrom(ctx, ruleServiceCustomPortRangeAttrTypes, model)
	diags.Append(d...)
	return result
}

type readSvcSimple = cato.PolicySocketBypassPolicy_Policy_SocketBypass_Policy_Rules_Rule_Service_Simple

func parseReadServiceSimple(
	ctx context.Context, value *readSvcSimple, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.service.simple object.")
		return types.ObjectNull(ruleServiceSimpleAttrTypes)
	}
	model := RuleServiceSimpleModel{}
	model.Name = types.StringValue(string(value.Name))
	result, d := types.ObjectValueFrom(ctx, ruleServiceSimpleAttrTypes, model)
	diags.Append(d...)
	return result
}

type readSite = cato.PolicySocketBypassPolicy_Policy_SocketBypass_Policy_Rules_Rule_Site

func parseReadSite(
	ctx context.Context, value *readSite, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.site object.")
		return types.ObjectNull(ruleSiteAttrTypes)
	}
	model := RuleSiteModel{}
	model.Group = idname.ParseIDNameSet(ctx, value.Group, diags)
	if model.Group.IsNull() {
		model.Group = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Site = idname.ParseIDNameSet(ctx, value.Site, diags)
	if model.Site.IsNull() {
		model.Site = types.SetValueMust(idname.RefObjectType, nil)
	}
	result, d := types.ObjectValueFrom(ctx, ruleSiteAttrTypes, model)
	diags.Append(d...)
	return result
}

type readSource = cato.PolicySocketBypassPolicy_Policy_SocketBypass_Policy_Rules_Rule_Source

func parseReadSource(
	ctx context.Context, value *readSource, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.source object.")
		return types.ObjectNull(ruleSourceAttrTypes)
	}
	model := RuleSourceModel{}
	model.FloatingSubnet = idname.ParseIDNameSet(ctx, value.FloatingSubnet, diags)
	if model.FloatingSubnet.IsNull() {
		model.FloatingSubnet = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.GlobalIPRange = idname.ParseIDNameSet(ctx, value.GlobalIPRange, diags)
	if model.GlobalIPRange.IsNull() {
		model.GlobalIPRange = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Group = idname.ParseIDNameSet(ctx, value.Group, diags)
	if model.Group.IsNull() {
		model.Group = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Host = idname.ParseIDNameSet(ctx, value.Host, diags)
	if model.Host.IsNull() {
		model.Host = types.SetValueMust(idname.RefObjectType, nil)
	}
	var dIP diag.Diagnostics
	model.IP, dIP = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.IP...))
	diags.Append(dIP...)
	valuesIPRange := make([]attr.Value, 0, len(value.IPRange))
	for _, item := range value.IPRange {
		valuesIPRange = append(valuesIPRange, parseReadSourceIPRange(ctx, item, diags))
	}
	var dIPRange diag.Diagnostics
	model.IPRange, dIPRange = types.SetValue(types.ObjectType{AttrTypes: ruleSourceIPRangeAttrTypes}, valuesIPRange)
	diags.Append(dIPRange...)
	model.NetworkInterface = idname.ParseIDNameSet(ctx, value.NetworkInterface, diags)
	if model.NetworkInterface.IsNull() {
		model.NetworkInterface = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.SiteNetworkSubnet = idname.ParseIDNameSet(ctx, value.SiteNetworkSubnet, diags)
	if model.SiteNetworkSubnet.IsNull() {
		model.SiteNetworkSubnet = types.SetValueMust(idname.RefObjectType, nil)
	}
	var dSubnet diag.Diagnostics
	model.Subnet, dSubnet = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Subnet...))
	diags.Append(dSubnet...)
	model.Vlan = parseVLANs(ctx, value.Vlan, diags)
	result, d := types.ObjectValueFrom(ctx, ruleSourceAttrTypes, model)
	diags.Append(d...)
	return result
}

type readSourceIPRange = cato.PolicySocketBypassPolicy_Policy_SocketBypass_Policy_Rules_Rule_Source_IPRange

func parseReadSourceIPRange(
	ctx context.Context, value *readSourceIPRange, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.source.ip_range object.")
		return types.ObjectNull(ruleSourceIPRangeAttrTypes)
	}
	model := RuleSourceIPRangeModel{}
	model.From = types.StringValue(value.From)
	model.To = types.StringValue(value.To)
	result, d := types.ObjectValueFrom(ctx, ruleSourceIPRangeAttrTypes, model)
	diags.Append(d...)
	return result
}

type add = cato.PolicySocketBypassAddRule_Policy_SocketBypass_AddRule_Rule_Rule

func parseAdd(
	ctx context.Context, value *add, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule object.")
		return types.ObjectNull(ruleAttrTypes)
	}
	model := RuleModel{}
	model.Action = parseAddAction(ctx, &value.Action, diags)
	model.Description = types.StringValue(value.Description)
	model.Destination = parseAddDestination(ctx, &value.Destination, diags)
	model.Enabled = types.BoolValue(value.Enabled)
	valuesException := make([]attr.Value, 0, len(value.Exception))
	for _, item := range value.Exception {
		valuesException = append(valuesException, parseAddException(ctx, item, diags))
	}
	var dException diag.Diagnostics
	model.Exception, dException = types.SetValue(types.ObjectType{AttrTypes: ruleExceptionAttrTypes}, valuesException)
	diags.Append(dException...)
	model.ID = types.StringValue(value.ID)
	model.Index = types.Int64Value(value.Index)
	model.Name = types.StringValue(value.Name)
	model.Section = parseAddSection(ctx, &value.Section, diags)
	model.Service = parseAddService(ctx, &value.Service, diags)
	model.Site = parseAddSite(ctx, &value.Site, diags)
	model.Source = parseAddSource(ctx, &value.Source, diags)
	result, d := types.ObjectValueFrom(ctx, ruleAttrTypes, model)
	diags.Append(d...)
	return result
}

type addAction = cato.PolicySocketBypassAddRule_Policy_SocketBypass_AddRule_Rule_Rule_Action

func parseAddAction(
	ctx context.Context, value *addAction, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.action object.")
		return types.ObjectNull(ruleActionAttrTypes)
	}
	model := RuleActionModel{}
	model.Event = parseAddActionEvent(ctx, &value.Event, diags)
	model.Port = types.StringValue(string(value.Port))
	result, d := types.ObjectValueFrom(ctx, ruleActionAttrTypes, model)
	diags.Append(d...)
	return result
}

type addActionEvent = cato.PolicySocketBypassAddRule_Policy_SocketBypass_AddRule_Rule_Rule_Action_Event

func parseAddActionEvent(
	ctx context.Context, value *addActionEvent, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.action.event object.")
		return types.ObjectNull(ruleActionEventAttrTypes)
	}
	model := RuleActionEventModel{}
	model.Enabled = types.BoolValue(value.Enabled)
	result, d := types.ObjectValueFrom(ctx, ruleActionEventAttrTypes, model)
	diags.Append(d...)
	return result
}

type addDst = cato.PolicySocketBypassAddRule_Policy_SocketBypass_AddRule_Rule_Rule_Destination

func parseAddDestination(
	ctx context.Context, value *addDst, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.destination object.")
		return types.ObjectNull(ruleDestinationAttrTypes)
	}
	model := RuleDestinationModel{}
	model.Application = idname.ParseIDNameSet(ctx, value.Application, diags)
	if model.Application.IsNull() {
		model.Application = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.CustomApp = idname.ParseIDNameSet(ctx, value.CustomApp, diags)
	if model.CustomApp.IsNull() {
		model.CustomApp = types.SetValueMust(idname.RefObjectType, nil)
	}
	var dDomain diag.Diagnostics
	model.Domain, dDomain = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Domain...))
	diags.Append(dDomain...)
	var dFqdn diag.Diagnostics
	model.Fqdn, dFqdn = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Fqdn...))
	diags.Append(dFqdn...)
	model.GlobalIPRange = idname.ParseIDNameSet(ctx, value.GlobalIPRange, diags)
	if model.GlobalIPRange.IsNull() {
		model.GlobalIPRange = types.SetValueMust(idname.RefObjectType, nil)
	}
	var dIP diag.Diagnostics
	model.IP, dIP = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.IP...))
	diags.Append(dIP...)
	valuesIPRange := make([]attr.Value, 0, len(value.IPRange))
	for _, item := range value.IPRange {
		valuesIPRange = append(valuesIPRange, parseAddDestinationIPRange(ctx, item, diags))
	}
	var dIPRange diag.Diagnostics
	model.IPRange, dIPRange = types.SetValue(types.ObjectType{AttrTypes: ruleDestinationIPRangeAttrTypes}, valuesIPRange)
	diags.Append(dIPRange...)
	var dSubnet diag.Diagnostics
	model.Subnet, dSubnet = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Subnet...))
	diags.Append(dSubnet...)
	result, d := types.ObjectValueFrom(ctx, ruleDestinationAttrTypes, model)
	diags.Append(d...)
	return result
}

type addDstIPRange = cato.PolicySocketBypassAddRule_Policy_SocketBypass_AddRule_Rule_Rule_Destination_IPRange

func parseAddDestinationIPRange(
	ctx context.Context, value *addDstIPRange, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.destination.ip_range object.")
		return types.ObjectNull(ruleDestinationIPRangeAttrTypes)
	}
	model := RuleDestinationIPRangeModel{}
	model.From = types.StringValue(value.From)
	model.To = types.StringValue(value.To)
	result, d := types.ObjectValueFrom(ctx, ruleDestinationIPRangeAttrTypes, model)
	diags.Append(d...)
	return result
}

type addExc = cato.PolicySocketBypassAddRule_Policy_SocketBypass_AddRule_Rule_Rule_Exception

func parseAddException(
	ctx context.Context, value *addExc, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.exception object.")
		return types.ObjectNull(ruleExceptionAttrTypes)
	}
	model := RuleExceptionModel{}
	model.Destination = parseAddExceptionDestination(ctx, &value.Destination, diags)
	model.Name = types.StringValue(value.Name)
	model.Service = parseAddExceptionService(ctx, &value.Service, diags)
	model.Site = parseAddExceptionSite(ctx, &value.Site, diags)
	model.Source = parseAddExceptionSource(ctx, &value.Source, diags)
	result, d := types.ObjectValueFrom(ctx, ruleExceptionAttrTypes, model)
	diags.Append(d...)
	return result
}

type addExcDst = cato.PolicySocketBypassAddRule_Policy_SocketBypass_AddRule_Rule_Rule_Exception_Destination

func parseAddExceptionDestination(
	ctx context.Context, value *addExcDst, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.exception.destination object.")
		return types.ObjectNull(ruleExceptionDestinationAttrTypes)
	}
	model := RuleExceptionDestinationModel{}
	model.Application = idname.ParseIDNameSet(ctx, value.Application, diags)
	if model.Application.IsNull() {
		model.Application = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.CustomApp = idname.ParseIDNameSet(ctx, value.CustomApp, diags)
	if model.CustomApp.IsNull() {
		model.CustomApp = types.SetValueMust(idname.RefObjectType, nil)
	}
	var dDomain diag.Diagnostics
	model.Domain, dDomain = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Domain...))
	diags.Append(dDomain...)
	var dFqdn diag.Diagnostics
	model.Fqdn, dFqdn = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Fqdn...))
	diags.Append(dFqdn...)
	model.GlobalIPRange = idname.ParseIDNameSet(ctx, value.GlobalIPRange, diags)
	if model.GlobalIPRange.IsNull() {
		model.GlobalIPRange = types.SetValueMust(idname.RefObjectType, nil)
	}
	var dIP diag.Diagnostics
	model.IP, dIP = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.IP...))
	diags.Append(dIP...)
	valuesIPRange := make([]attr.Value, 0, len(value.IPRange))
	for _, item := range value.IPRange {
		valuesIPRange = append(valuesIPRange, parseAddExceptionDestinationIPRange(ctx, item, diags))
	}
	var dIPRange diag.Diagnostics
	model.IPRange, dIPRange = types.SetValue(types.ObjectType{AttrTypes: ruleExceptionDestinationIPRangeAttrTypes}, valuesIPRange)
	diags.Append(dIPRange...)
	var dSubnet diag.Diagnostics
	model.Subnet, dSubnet = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Subnet...))
	diags.Append(dSubnet...)
	result, d := types.ObjectValueFrom(ctx, ruleExceptionDestinationAttrTypes, model)
	diags.Append(d...)
	return result
}

type addExcDstIPRange = cato.PolicySocketBypassAddRule_Policy_SocketBypass_AddRule_Rule_Rule_Exception_Destination_IPRange

func parseAddExceptionDestinationIPRange(
	ctx context.Context, value *addExcDstIPRange, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.exception.destination.ip_range object.")
		return types.ObjectNull(ruleExceptionDestinationIPRangeAttrTypes)
	}
	model := RuleExceptionDestinationIPRangeModel{}
	model.From = types.StringValue(value.From)
	model.To = types.StringValue(value.To)
	result, d := types.ObjectValueFrom(ctx, ruleExceptionDestinationIPRangeAttrTypes, model)
	diags.Append(d...)
	return result
}

type addExcSvc = cato.PolicySocketBypassAddRule_Policy_SocketBypass_AddRule_Rule_Rule_Exception_Service

func parseAddExceptionService(
	ctx context.Context, value *addExcSvc, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.exception.service object.")
		return types.ObjectNull(ruleExceptionServiceAttrTypes)
	}
	model := RuleExceptionServiceModel{}
	valuesCustom := make([]attr.Value, 0, len(value.Custom))
	for _, item := range value.Custom {
		valuesCustom = append(valuesCustom, parseAddExceptionServiceCustom(ctx, item, diags))
	}
	var dCustom diag.Diagnostics
	model.Custom, dCustom = types.SetValue(types.ObjectType{AttrTypes: ruleExceptionServiceCustomAttrTypes}, valuesCustom)
	diags.Append(dCustom...)
	valuesSimple := make([]attr.Value, 0, len(value.Simple))
	for _, item := range value.Simple {
		valuesSimple = append(valuesSimple, parseAddExceptionServiceSimple(ctx, item, diags))
	}
	var dSimple diag.Diagnostics
	model.Simple, dSimple = types.SetValue(types.ObjectType{AttrTypes: ruleExceptionServiceSimpleAttrTypes}, valuesSimple)
	diags.Append(dSimple...)
	result, d := types.ObjectValueFrom(ctx, ruleExceptionServiceAttrTypes, model)
	diags.Append(d...)
	return result
}

type addExcSvcCustom = cato.PolicySocketBypassAddRule_Policy_SocketBypass_AddRule_Rule_Rule_Exception_Service_Custom

func parseAddExceptionServiceCustom(
	ctx context.Context, value *addExcSvcCustom, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.exception.service.custom object.")
		return types.ObjectNull(ruleExceptionServiceCustomAttrTypes)
	}
	model := RuleExceptionServiceCustomModel{}
	model.Port = parsePorts(ctx, value.Port, diags)
	if value.PortRange == nil {
		model.PortRange = types.ObjectNull(ruleExceptionServiceCustomPortRangeAttrTypes)
	} else {
		model.PortRange = parseAddExceptionServiceCustomPortRange(ctx, value.PortRange, diags)
	}
	model.Protocol = types.StringValue(string(value.Protocol))
	result, d := types.ObjectValueFrom(ctx, ruleExceptionServiceCustomAttrTypes, model)
	diags.Append(d...)
	return result
}

type addExcSvcCustomPortRange = cato.PolicySocketBypassAddRule_Policy_SocketBypass_AddRule_Rule_Rule_Exception_Service_Custom_PortRange

func parseAddExceptionServiceCustomPortRange(
	ctx context.Context, value *addExcSvcCustomPortRange, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.exception.service.custom.port_range object.")
		return types.ObjectNull(ruleExceptionServiceCustomPortRangeAttrTypes)
	}
	model := RuleExceptionServiceCustomPortRangeModel{}
	model.From = types.StringValue(string(value.From))
	model.To = types.StringValue(string(value.To))
	result, d := types.ObjectValueFrom(ctx, ruleExceptionServiceCustomPortRangeAttrTypes, model)
	diags.Append(d...)
	return result
}

type addExcSvcSimple = cato.PolicySocketBypassAddRule_Policy_SocketBypass_AddRule_Rule_Rule_Exception_Service_Simple

func parseAddExceptionServiceSimple(
	ctx context.Context, value *addExcSvcSimple, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.exception.service.simple object.")
		return types.ObjectNull(ruleExceptionServiceSimpleAttrTypes)
	}
	model := RuleExceptionServiceSimpleModel{}
	model.Name = types.StringValue(string(value.Name))
	result, d := types.ObjectValueFrom(ctx, ruleExceptionServiceSimpleAttrTypes, model)
	diags.Append(d...)
	return result
}

type addExcSite = cato.PolicySocketBypassAddRule_Policy_SocketBypass_AddRule_Rule_Rule_Exception_Site

func parseAddExceptionSite(
	ctx context.Context, value *addExcSite, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.exception.site object.")
		return types.ObjectNull(ruleExceptionSiteAttrTypes)
	}
	model := RuleExceptionSiteModel{}
	model.Group = idname.ParseIDNameSet(ctx, value.Group, diags)
	if model.Group.IsNull() {
		model.Group = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Site = idname.ParseIDNameSet(ctx, value.Site, diags)
	if model.Site.IsNull() {
		model.Site = types.SetValueMust(idname.RefObjectType, nil)
	}
	result, d := types.ObjectValueFrom(ctx, ruleExceptionSiteAttrTypes, model)
	diags.Append(d...)
	return result
}

type addExcSource = cato.PolicySocketBypassAddRule_Policy_SocketBypass_AddRule_Rule_Rule_Exception_Source

func parseAddExceptionSource(
	ctx context.Context, value *addExcSource, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.exception.source object.")
		return types.ObjectNull(ruleExceptionSourceAttrTypes)
	}
	model := RuleExceptionSourceModel{}
	model.FloatingSubnet = idname.ParseIDNameSet(ctx, value.FloatingSubnet, diags)
	if model.FloatingSubnet.IsNull() {
		model.FloatingSubnet = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.GlobalIPRange = idname.ParseIDNameSet(ctx, value.GlobalIPRange, diags)
	if model.GlobalIPRange.IsNull() {
		model.GlobalIPRange = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Group = idname.ParseIDNameSet(ctx, value.Group, diags)
	if model.Group.IsNull() {
		model.Group = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Host = idname.ParseIDNameSet(ctx, value.Host, diags)
	if model.Host.IsNull() {
		model.Host = types.SetValueMust(idname.RefObjectType, nil)
	}
	var dIP diag.Diagnostics
	model.IP, dIP = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.IP...))
	diags.Append(dIP...)
	valuesIPRange := make([]attr.Value, 0, len(value.IPRange))
	for _, item := range value.IPRange {
		valuesIPRange = append(valuesIPRange, parseAddExceptionSourceIPRange(ctx, item, diags))
	}
	var dIPRange diag.Diagnostics
	model.IPRange, dIPRange = types.SetValue(types.ObjectType{AttrTypes: ruleExceptionSourceIPRangeAttrTypes}, valuesIPRange)
	diags.Append(dIPRange...)
	model.NetworkInterface = idname.ParseIDNameSet(ctx, value.NetworkInterface, diags)
	if model.NetworkInterface.IsNull() {
		model.NetworkInterface = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.SiteNetworkSubnet = idname.ParseIDNameSet(ctx, value.SiteNetworkSubnet, diags)
	if model.SiteNetworkSubnet.IsNull() {
		model.SiteNetworkSubnet = types.SetValueMust(idname.RefObjectType, nil)
	}
	var dSubnet diag.Diagnostics
	model.Subnet, dSubnet = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Subnet...))
	diags.Append(dSubnet...)
	model.Vlan = parseVLANs(ctx, value.Vlan, diags)
	result, d := types.ObjectValueFrom(ctx, ruleExceptionSourceAttrTypes, model)
	diags.Append(d...)
	return result
}

type addExcSourceIPRange = cato.PolicySocketBypassAddRule_Policy_SocketBypass_AddRule_Rule_Rule_Exception_Source_IPRange

func parseAddExceptionSourceIPRange(
	ctx context.Context, value *addExcSourceIPRange, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.exception.source.ip_range object.")
		return types.ObjectNull(ruleExceptionSourceIPRangeAttrTypes)
	}
	model := RuleExceptionSourceIPRangeModel{}
	model.From = types.StringValue(value.From)
	model.To = types.StringValue(value.To)
	result, d := types.ObjectValueFrom(ctx, ruleExceptionSourceIPRangeAttrTypes, model)
	diags.Append(d...)
	return result
}

type addSection = cato.PolicySocketBypassAddRule_Policy_SocketBypass_AddRule_Rule_Rule_Section

func parseAddSection(
	ctx context.Context, value *addSection, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.section object.")
		return types.ObjectNull(ruleSectionAttrTypes)
	}
	return idname.ParseIDName(ctx, *value, diags)
}

type addSvc = cato.PolicySocketBypassAddRule_Policy_SocketBypass_AddRule_Rule_Rule_Service

func parseAddService(
	ctx context.Context, value *addSvc, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.service object.")
		return types.ObjectNull(ruleServiceAttrTypes)
	}
	model := RuleServiceModel{}
	valuesCustom := make([]attr.Value, 0, len(value.Custom))
	for _, item := range value.Custom {
		valuesCustom = append(valuesCustom, parseAddServiceCustom(ctx, item, diags))
	}
	var dCustom diag.Diagnostics
	model.Custom, dCustom = types.SetValue(types.ObjectType{AttrTypes: ruleServiceCustomAttrTypes}, valuesCustom)
	diags.Append(dCustom...)
	valuesSimple := make([]attr.Value, 0, len(value.Simple))
	for _, item := range value.Simple {
		valuesSimple = append(valuesSimple, parseAddServiceSimple(ctx, item, diags))
	}
	var dSimple diag.Diagnostics
	model.Simple, dSimple = types.SetValue(types.ObjectType{AttrTypes: ruleServiceSimpleAttrTypes}, valuesSimple)
	diags.Append(dSimple...)
	result, d := types.ObjectValueFrom(ctx, ruleServiceAttrTypes, model)
	diags.Append(d...)
	return result
}

type addSvcCustom = cato.PolicySocketBypassAddRule_Policy_SocketBypass_AddRule_Rule_Rule_Service_Custom

func parseAddServiceCustom(
	ctx context.Context, value *addSvcCustom, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.service.custom object.")
		return types.ObjectNull(ruleServiceCustomAttrTypes)
	}
	model := RuleServiceCustomModel{}
	model.Port = parsePorts(ctx, value.Port, diags)
	if value.PortRange == nil {
		model.PortRange = types.ObjectNull(ruleServiceCustomPortRangeAttrTypes)
	} else {
		model.PortRange = parseAddServiceCustomPortRange(ctx, value.PortRange, diags)
	}
	model.Protocol = types.StringValue(string(value.Protocol))
	result, d := types.ObjectValueFrom(ctx, ruleServiceCustomAttrTypes, model)
	diags.Append(d...)
	return result
}

type addSvcCustomPortRange = cato.PolicySocketBypassAddRule_Policy_SocketBypass_AddRule_Rule_Rule_Service_Custom_PortRange

func parseAddServiceCustomPortRange(
	ctx context.Context, value *addSvcCustomPortRange, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.service.custom.port_range object.")
		return types.ObjectNull(ruleServiceCustomPortRangeAttrTypes)
	}
	model := RuleServiceCustomPortRangeModel{}
	model.From = types.StringValue(string(value.From))
	model.To = types.StringValue(string(value.To))
	result, d := types.ObjectValueFrom(ctx, ruleServiceCustomPortRangeAttrTypes, model)
	diags.Append(d...)
	return result
}

type addSvcSimple = cato.PolicySocketBypassAddRule_Policy_SocketBypass_AddRule_Rule_Rule_Service_Simple

func parseAddServiceSimple(
	ctx context.Context, value *addSvcSimple, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.service.simple object.")
		return types.ObjectNull(ruleServiceSimpleAttrTypes)
	}
	model := RuleServiceSimpleModel{}
	model.Name = types.StringValue(string(value.Name))
	result, d := types.ObjectValueFrom(ctx, ruleServiceSimpleAttrTypes, model)
	diags.Append(d...)
	return result
}

type addSite = cato.PolicySocketBypassAddRule_Policy_SocketBypass_AddRule_Rule_Rule_Site

func parseAddSite(
	ctx context.Context, value *addSite, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.site object.")
		return types.ObjectNull(ruleSiteAttrTypes)
	}
	model := RuleSiteModel{}
	model.Group = idname.ParseIDNameSet(ctx, value.Group, diags)
	if model.Group.IsNull() {
		model.Group = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Site = idname.ParseIDNameSet(ctx, value.Site, diags)
	if model.Site.IsNull() {
		model.Site = types.SetValueMust(idname.RefObjectType, nil)
	}
	result, d := types.ObjectValueFrom(ctx, ruleSiteAttrTypes, model)
	diags.Append(d...)
	return result
}

type addSource = cato.PolicySocketBypassAddRule_Policy_SocketBypass_AddRule_Rule_Rule_Source

func parseAddSource(
	ctx context.Context, value *addSource, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.source object.")
		return types.ObjectNull(ruleSourceAttrTypes)
	}
	model := RuleSourceModel{}
	model.FloatingSubnet = idname.ParseIDNameSet(ctx, value.FloatingSubnet, diags)
	if model.FloatingSubnet.IsNull() {
		model.FloatingSubnet = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.GlobalIPRange = idname.ParseIDNameSet(ctx, value.GlobalIPRange, diags)
	if model.GlobalIPRange.IsNull() {
		model.GlobalIPRange = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Group = idname.ParseIDNameSet(ctx, value.Group, diags)
	if model.Group.IsNull() {
		model.Group = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Host = idname.ParseIDNameSet(ctx, value.Host, diags)
	if model.Host.IsNull() {
		model.Host = types.SetValueMust(idname.RefObjectType, nil)
	}
	var dIP diag.Diagnostics
	model.IP, dIP = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.IP...))
	diags.Append(dIP...)
	valuesIPRange := make([]attr.Value, 0, len(value.IPRange))
	for _, item := range value.IPRange {
		valuesIPRange = append(valuesIPRange, parseAddSourceIPRange(ctx, item, diags))
	}
	var dIPRange diag.Diagnostics
	model.IPRange, dIPRange = types.SetValue(types.ObjectType{AttrTypes: ruleSourceIPRangeAttrTypes}, valuesIPRange)
	diags.Append(dIPRange...)
	model.NetworkInterface = idname.ParseIDNameSet(ctx, value.NetworkInterface, diags)
	if model.NetworkInterface.IsNull() {
		model.NetworkInterface = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.SiteNetworkSubnet = idname.ParseIDNameSet(ctx, value.SiteNetworkSubnet, diags)
	if model.SiteNetworkSubnet.IsNull() {
		model.SiteNetworkSubnet = types.SetValueMust(idname.RefObjectType, nil)
	}
	var dSubnet diag.Diagnostics
	model.Subnet, dSubnet = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Subnet...))
	diags.Append(dSubnet...)
	model.Vlan = parseVLANs(ctx, value.Vlan, diags)
	result, d := types.ObjectValueFrom(ctx, ruleSourceAttrTypes, model)
	diags.Append(d...)
	return result
}

type addSourceIPRange = cato.PolicySocketBypassAddRule_Policy_SocketBypass_AddRule_Rule_Rule_Source_IPRange

func parseAddSourceIPRange(
	ctx context.Context, value *addSourceIPRange, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.source.ip_range object.")
		return types.ObjectNull(ruleSourceIPRangeAttrTypes)
	}
	model := RuleSourceIPRangeModel{}
	model.From = types.StringValue(value.From)
	model.To = types.StringValue(value.To)
	result, d := types.ObjectValueFrom(ctx, ruleSourceIPRangeAttrTypes, model)
	diags.Append(d...)
	return result
}

type update = cato.PolicySocketBypassUpdateRule_Policy_SocketBypass_UpdateRule_Rule_Rule

func parseUpdate(
	ctx context.Context, value *update, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule object.")
		return types.ObjectNull(ruleAttrTypes)
	}
	model := RuleModel{}
	model.Action = parseUpdateAction(ctx, &value.Action, diags)
	model.Description = types.StringValue(value.Description)
	model.Destination = parseUpdateDestination(ctx, &value.Destination, diags)
	model.Enabled = types.BoolValue(value.Enabled)
	valuesException := make([]attr.Value, 0, len(value.Exception))
	for _, item := range value.Exception {
		valuesException = append(valuesException, parseUpdateException(ctx, item, diags))
	}
	var dException diag.Diagnostics
	model.Exception, dException = types.SetValue(types.ObjectType{AttrTypes: ruleExceptionAttrTypes}, valuesException)
	diags.Append(dException...)
	model.ID = types.StringValue(value.ID)
	model.Index = types.Int64Value(value.Index)
	model.Name = types.StringValue(value.Name)
	model.Section = parseUpdateSection(ctx, &value.Section, diags)
	model.Service = parseUpdateService(ctx, &value.Service, diags)
	model.Site = parseUpdateSite(ctx, &value.Site, diags)
	model.Source = parseUpdateSource(ctx, &value.Source, diags)
	result, d := types.ObjectValueFrom(ctx, ruleAttrTypes, model)
	diags.Append(d...)
	return result
}

type updateAction = cato.PolicySocketBypassUpdateRule_Policy_SocketBypass_UpdateRule_Rule_Rule_Action

func parseUpdateAction(
	ctx context.Context, value *updateAction, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.action object.")
		return types.ObjectNull(ruleActionAttrTypes)
	}
	model := RuleActionModel{}
	model.Event = parseUpdateActionEvent(ctx, &value.Event, diags)
	model.Port = types.StringValue(string(value.Port))
	result, d := types.ObjectValueFrom(ctx, ruleActionAttrTypes, model)
	diags.Append(d...)
	return result
}

type updateActionEvent = cato.PolicySocketBypassUpdateRule_Policy_SocketBypass_UpdateRule_Rule_Rule_Action_Event

func parseUpdateActionEvent(
	ctx context.Context, value *updateActionEvent, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.action.event object.")
		return types.ObjectNull(ruleActionEventAttrTypes)
	}
	model := RuleActionEventModel{}
	model.Enabled = types.BoolValue(value.Enabled)
	result, d := types.ObjectValueFrom(ctx, ruleActionEventAttrTypes, model)
	diags.Append(d...)
	return result
}

type updateDst = cato.PolicySocketBypassUpdateRule_Policy_SocketBypass_UpdateRule_Rule_Rule_Destination

func parseUpdateDestination(
	ctx context.Context, value *updateDst, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.destination object.")
		return types.ObjectNull(ruleDestinationAttrTypes)
	}
	model := RuleDestinationModel{}
	model.Application = idname.ParseIDNameSet(ctx, value.Application, diags)
	if model.Application.IsNull() {
		model.Application = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.CustomApp = idname.ParseIDNameSet(ctx, value.CustomApp, diags)
	if model.CustomApp.IsNull() {
		model.CustomApp = types.SetValueMust(idname.RefObjectType, nil)
	}
	var dDomain diag.Diagnostics
	model.Domain, dDomain = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Domain...))
	diags.Append(dDomain...)
	var dFqdn diag.Diagnostics
	model.Fqdn, dFqdn = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Fqdn...))
	diags.Append(dFqdn...)
	model.GlobalIPRange = idname.ParseIDNameSet(ctx, value.GlobalIPRange, diags)
	if model.GlobalIPRange.IsNull() {
		model.GlobalIPRange = types.SetValueMust(idname.RefObjectType, nil)
	}
	var dIP diag.Diagnostics
	model.IP, dIP = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.IP...))
	diags.Append(dIP...)
	valuesIPRange := make([]attr.Value, 0, len(value.IPRange))
	for _, item := range value.IPRange {
		valuesIPRange = append(valuesIPRange, parseUpdateDestinationIPRange(ctx, item, diags))
	}
	var dIPRange diag.Diagnostics
	model.IPRange, dIPRange = types.SetValue(types.ObjectType{AttrTypes: ruleDestinationIPRangeAttrTypes}, valuesIPRange)
	diags.Append(dIPRange...)
	var dSubnet diag.Diagnostics
	model.Subnet, dSubnet = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Subnet...))
	diags.Append(dSubnet...)
	result, d := types.ObjectValueFrom(ctx, ruleDestinationAttrTypes, model)
	diags.Append(d...)
	return result
}

type updateDstIPRange = cato.PolicySocketBypassUpdateRule_Policy_SocketBypass_UpdateRule_Rule_Rule_Destination_IPRange

func parseUpdateDestinationIPRange(
	ctx context.Context, value *updateDstIPRange, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.destination.ip_range object.")
		return types.ObjectNull(ruleDestinationIPRangeAttrTypes)
	}
	model := RuleDestinationIPRangeModel{}
	model.From = types.StringValue(value.From)
	model.To = types.StringValue(value.To)
	result, d := types.ObjectValueFrom(ctx, ruleDestinationIPRangeAttrTypes, model)
	diags.Append(d...)
	return result
}

type updateExc = cato.PolicySocketBypassUpdateRule_Policy_SocketBypass_UpdateRule_Rule_Rule_Exception

func parseUpdateException(
	ctx context.Context, value *updateExc, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.exception object.")
		return types.ObjectNull(ruleExceptionAttrTypes)
	}
	model := RuleExceptionModel{}
	model.Destination = parseUpdateExceptionDestination(ctx, &value.Destination, diags)
	model.Name = types.StringValue(value.Name)
	model.Service = parseUpdateExceptionService(ctx, &value.Service, diags)
	model.Site = parseUpdateExceptionSite(ctx, &value.Site, diags)
	model.Source = parseUpdateExceptionSource(ctx, &value.Source, diags)
	result, d := types.ObjectValueFrom(ctx, ruleExceptionAttrTypes, model)
	diags.Append(d...)
	return result
}

type updateExcDst = cato.PolicySocketBypassUpdateRule_Policy_SocketBypass_UpdateRule_Rule_Rule_Exception_Destination

func parseUpdateExceptionDestination(
	ctx context.Context, value *updateExcDst, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.exception.destination object.")
		return types.ObjectNull(ruleExceptionDestinationAttrTypes)
	}
	model := RuleExceptionDestinationModel{}
	model.Application = idname.ParseIDNameSet(ctx, value.Application, diags)
	if model.Application.IsNull() {
		model.Application = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.CustomApp = idname.ParseIDNameSet(ctx, value.CustomApp, diags)
	if model.CustomApp.IsNull() {
		model.CustomApp = types.SetValueMust(idname.RefObjectType, nil)
	}
	var dDomain diag.Diagnostics
	model.Domain, dDomain = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Domain...))
	diags.Append(dDomain...)
	var dFqdn diag.Diagnostics
	model.Fqdn, dFqdn = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Fqdn...))
	diags.Append(dFqdn...)
	model.GlobalIPRange = idname.ParseIDNameSet(ctx, value.GlobalIPRange, diags)
	if model.GlobalIPRange.IsNull() {
		model.GlobalIPRange = types.SetValueMust(idname.RefObjectType, nil)
	}
	var dIP diag.Diagnostics
	model.IP, dIP = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.IP...))
	diags.Append(dIP...)
	valuesIPRange := make([]attr.Value, 0, len(value.IPRange))
	for _, item := range value.IPRange {
		valuesIPRange = append(valuesIPRange, parseUpdateExceptionDestinationIPRange(ctx, item, diags))
	}
	var dIPRange diag.Diagnostics
	model.IPRange, dIPRange = types.SetValue(types.ObjectType{AttrTypes: ruleExceptionDestinationIPRangeAttrTypes}, valuesIPRange)
	diags.Append(dIPRange...)
	var dSubnet diag.Diagnostics
	model.Subnet, dSubnet = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Subnet...))
	diags.Append(dSubnet...)
	result, d := types.ObjectValueFrom(ctx, ruleExceptionDestinationAttrTypes, model)
	diags.Append(d...)
	return result
}

type updateExcDstIPRange = cato.PolicySocketBypassUpdateRule_Policy_SocketBypass_UpdateRule_Rule_Rule_Exception_Destination_IPRange

func parseUpdateExceptionDestinationIPRange(
	ctx context.Context, value *updateExcDstIPRange, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.exception.destination.ip_range object.")
		return types.ObjectNull(ruleExceptionDestinationIPRangeAttrTypes)
	}
	model := RuleExceptionDestinationIPRangeModel{}
	model.From = types.StringValue(value.From)
	model.To = types.StringValue(value.To)
	result, d := types.ObjectValueFrom(ctx, ruleExceptionDestinationIPRangeAttrTypes, model)
	diags.Append(d...)
	return result
}

type updateExcSvc = cato.PolicySocketBypassUpdateRule_Policy_SocketBypass_UpdateRule_Rule_Rule_Exception_Service

func parseUpdateExceptionService(
	ctx context.Context, value *updateExcSvc, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.exception.service object.")
		return types.ObjectNull(ruleExceptionServiceAttrTypes)
	}
	model := RuleExceptionServiceModel{}
	valuesCustom := make([]attr.Value, 0, len(value.Custom))
	for _, item := range value.Custom {
		valuesCustom = append(valuesCustom, parseUpdateExceptionServiceCustom(ctx, item, diags))
	}
	var dCustom diag.Diagnostics
	model.Custom, dCustom = types.SetValue(types.ObjectType{AttrTypes: ruleExceptionServiceCustomAttrTypes}, valuesCustom)
	diags.Append(dCustom...)
	valuesSimple := make([]attr.Value, 0, len(value.Simple))
	for _, item := range value.Simple {
		valuesSimple = append(valuesSimple, parseUpdateExceptionServiceSimple(ctx, item, diags))
	}
	var dSimple diag.Diagnostics
	model.Simple, dSimple = types.SetValue(types.ObjectType{AttrTypes: ruleExceptionServiceSimpleAttrTypes}, valuesSimple)
	diags.Append(dSimple...)
	result, d := types.ObjectValueFrom(ctx, ruleExceptionServiceAttrTypes, model)
	diags.Append(d...)
	return result
}

type updateExcSvcCustom = cato.PolicySocketBypassUpdateRule_Policy_SocketBypass_UpdateRule_Rule_Rule_Exception_Service_Custom

func parseUpdateExceptionServiceCustom(
	ctx context.Context, value *updateExcSvcCustom, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.exception.service.custom object.")
		return types.ObjectNull(ruleExceptionServiceCustomAttrTypes)
	}
	model := RuleExceptionServiceCustomModel{}
	model.Port = parsePorts(ctx, value.Port, diags)
	if value.PortRange == nil {
		model.PortRange = types.ObjectNull(ruleExceptionServiceCustomPortRangeAttrTypes)
	} else {
		model.PortRange = parseUpdateExceptionServiceCustomPortRange(ctx, value.PortRange, diags)
	}
	model.Protocol = types.StringValue(string(value.Protocol))
	result, d := types.ObjectValueFrom(ctx, ruleExceptionServiceCustomAttrTypes, model)
	diags.Append(d...)
	return result
}

type updateExcSvcPorts = cato.PolicySocketBypassUpdateRule_Policy_SocketBypass_UpdateRule_Rule_Rule_Exception_Service_Custom_PortRange

func parseUpdateExceptionServiceCustomPortRange(
	ctx context.Context, value *updateExcSvcPorts, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.exception.service.custom.port_range object.")
		return types.ObjectNull(ruleExceptionServiceCustomPortRangeAttrTypes)
	}
	model := RuleExceptionServiceCustomPortRangeModel{}
	model.From = types.StringValue(string(value.From))
	model.To = types.StringValue(string(value.To))
	result, d := types.ObjectValueFrom(ctx, ruleExceptionServiceCustomPortRangeAttrTypes, model)
	diags.Append(d...)
	return result
}

type updateExcSvcSimple = cato.PolicySocketBypassUpdateRule_Policy_SocketBypass_UpdateRule_Rule_Rule_Exception_Service_Simple

func parseUpdateExceptionServiceSimple(
	ctx context.Context, value *updateExcSvcSimple, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.exception.service.simple object.")
		return types.ObjectNull(ruleExceptionServiceSimpleAttrTypes)
	}
	model := RuleExceptionServiceSimpleModel{}
	model.Name = types.StringValue(string(value.Name))
	result, d := types.ObjectValueFrom(ctx, ruleExceptionServiceSimpleAttrTypes, model)
	diags.Append(d...)
	return result
}

type updateExcSite = cato.PolicySocketBypassUpdateRule_Policy_SocketBypass_UpdateRule_Rule_Rule_Exception_Site

func parseUpdateExceptionSite(
	ctx context.Context, value *updateExcSite, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.exception.site object.")
		return types.ObjectNull(ruleExceptionSiteAttrTypes)
	}
	model := RuleExceptionSiteModel{}
	model.Group = idname.ParseIDNameSet(ctx, value.Group, diags)
	if model.Group.IsNull() {
		model.Group = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Site = idname.ParseIDNameSet(ctx, value.Site, diags)
	if model.Site.IsNull() {
		model.Site = types.SetValueMust(idname.RefObjectType, nil)
	}
	result, d := types.ObjectValueFrom(ctx, ruleExceptionSiteAttrTypes, model)
	diags.Append(d...)
	return result
}

type updateExcSource = cato.PolicySocketBypassUpdateRule_Policy_SocketBypass_UpdateRule_Rule_Rule_Exception_Source

func parseUpdateExceptionSource(
	ctx context.Context, value *updateExcSource, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.exception.source object.")
		return types.ObjectNull(ruleExceptionSourceAttrTypes)
	}
	model := RuleExceptionSourceModel{}
	model.FloatingSubnet = idname.ParseIDNameSet(ctx, value.FloatingSubnet, diags)
	if model.FloatingSubnet.IsNull() {
		model.FloatingSubnet = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.GlobalIPRange = idname.ParseIDNameSet(ctx, value.GlobalIPRange, diags)
	if model.GlobalIPRange.IsNull() {
		model.GlobalIPRange = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Group = idname.ParseIDNameSet(ctx, value.Group, diags)
	if model.Group.IsNull() {
		model.Group = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Host = idname.ParseIDNameSet(ctx, value.Host, diags)
	if model.Host.IsNull() {
		model.Host = types.SetValueMust(idname.RefObjectType, nil)
	}
	var dIP diag.Diagnostics
	model.IP, dIP = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.IP...))
	diags.Append(dIP...)
	valuesIPRange := make([]attr.Value, 0, len(value.IPRange))
	for _, item := range value.IPRange {
		valuesIPRange = append(valuesIPRange, parseUpdateExceptionSourceIPRange(ctx, item, diags))
	}
	var dIPRange diag.Diagnostics
	model.IPRange, dIPRange = types.SetValue(types.ObjectType{AttrTypes: ruleExceptionSourceIPRangeAttrTypes}, valuesIPRange)
	diags.Append(dIPRange...)
	model.NetworkInterface = idname.ParseIDNameSet(ctx, value.NetworkInterface, diags)
	if model.NetworkInterface.IsNull() {
		model.NetworkInterface = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.SiteNetworkSubnet = idname.ParseIDNameSet(ctx, value.SiteNetworkSubnet, diags)
	if model.SiteNetworkSubnet.IsNull() {
		model.SiteNetworkSubnet = types.SetValueMust(idname.RefObjectType, nil)
	}
	var dSubnet diag.Diagnostics
	model.Subnet, dSubnet = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Subnet...))
	diags.Append(dSubnet...)
	model.Vlan = parseVLANs(ctx, value.Vlan, diags)
	result, d := types.ObjectValueFrom(ctx, ruleExceptionSourceAttrTypes, model)
	diags.Append(d...)
	return result
}

type updateExcSourceIPRange = cato.PolicySocketBypassUpdateRule_Policy_SocketBypass_UpdateRule_Rule_Rule_Exception_Source_IPRange

func parseUpdateExceptionSourceIPRange(
	ctx context.Context, value *updateExcSourceIPRange, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.exception.source.ip_range object.")
		return types.ObjectNull(ruleExceptionSourceIPRangeAttrTypes)
	}
	model := RuleExceptionSourceIPRangeModel{}
	model.From = types.StringValue(value.From)
	model.To = types.StringValue(value.To)
	result, d := types.ObjectValueFrom(ctx, ruleExceptionSourceIPRangeAttrTypes, model)
	diags.Append(d...)
	return result
}

type updateSection = cato.PolicySocketBypassUpdateRule_Policy_SocketBypass_UpdateRule_Rule_Rule_Section

func parseUpdateSection(
	ctx context.Context, value *updateSection, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.section object.")
		return types.ObjectNull(ruleSectionAttrTypes)
	}
	return idname.ParseIDName(ctx, *value, diags)
}

type updateSvc = cato.PolicySocketBypassUpdateRule_Policy_SocketBypass_UpdateRule_Rule_Rule_Service

func parseUpdateService(
	ctx context.Context, value *updateSvc, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.service object.")
		return types.ObjectNull(ruleServiceAttrTypes)
	}
	model := RuleServiceModel{}
	valuesCustom := make([]attr.Value, 0, len(value.Custom))
	for _, item := range value.Custom {
		valuesCustom = append(valuesCustom, parseUpdateServiceCustom(ctx, item, diags))
	}
	var dCustom diag.Diagnostics
	model.Custom, dCustom = types.SetValue(types.ObjectType{AttrTypes: ruleServiceCustomAttrTypes}, valuesCustom)
	diags.Append(dCustom...)
	valuesSimple := make([]attr.Value, 0, len(value.Simple))
	for _, item := range value.Simple {
		valuesSimple = append(valuesSimple, parseUpdateServiceSimple(ctx, item, diags))
	}
	var dSimple diag.Diagnostics
	model.Simple, dSimple = types.SetValue(types.ObjectType{AttrTypes: ruleServiceSimpleAttrTypes}, valuesSimple)
	diags.Append(dSimple...)
	result, d := types.ObjectValueFrom(ctx, ruleServiceAttrTypes, model)
	diags.Append(d...)
	return result
}

type updateSvcCustom = cato.PolicySocketBypassUpdateRule_Policy_SocketBypass_UpdateRule_Rule_Rule_Service_Custom

func parseUpdateServiceCustom(
	ctx context.Context, value *updateSvcCustom, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.service.custom object.")
		return types.ObjectNull(ruleServiceCustomAttrTypes)
	}
	model := RuleServiceCustomModel{}
	model.Port = parsePorts(ctx, value.Port, diags)
	if value.PortRange == nil {
		model.PortRange = types.ObjectNull(ruleServiceCustomPortRangeAttrTypes)
	} else {
		model.PortRange = parseUpdateServiceCustomPortRange(ctx, value.PortRange, diags)
	}
	model.Protocol = types.StringValue(string(value.Protocol))
	result, d := types.ObjectValueFrom(ctx, ruleServiceCustomAttrTypes, model)
	diags.Append(d...)
	return result
}

type updateSvcCustomPortRange = cato.PolicySocketBypassUpdateRule_Policy_SocketBypass_UpdateRule_Rule_Rule_Service_Custom_PortRange

func parseUpdateServiceCustomPortRange(
	ctx context.Context, value *updateSvcCustomPortRange, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.service.custom.port_range object.")
		return types.ObjectNull(ruleServiceCustomPortRangeAttrTypes)
	}
	model := RuleServiceCustomPortRangeModel{}
	model.From = types.StringValue(string(value.From))
	model.To = types.StringValue(string(value.To))
	result, d := types.ObjectValueFrom(ctx, ruleServiceCustomPortRangeAttrTypes, model)
	diags.Append(d...)
	return result
}

type updateSvcSimple = cato.PolicySocketBypassUpdateRule_Policy_SocketBypass_UpdateRule_Rule_Rule_Service_Simple

func parseUpdateServiceSimple(
	ctx context.Context, value *updateSvcSimple, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.service.simple object.")
		return types.ObjectNull(ruleServiceSimpleAttrTypes)
	}
	model := RuleServiceSimpleModel{}
	model.Name = types.StringValue(string(value.Name))
	result, d := types.ObjectValueFrom(ctx, ruleServiceSimpleAttrTypes, model)
	diags.Append(d...)
	return result
}

type updateSite = cato.PolicySocketBypassUpdateRule_Policy_SocketBypass_UpdateRule_Rule_Rule_Site

func parseUpdateSite(
	ctx context.Context, value *updateSite, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.site object.")
		return types.ObjectNull(ruleSiteAttrTypes)
	}
	model := RuleSiteModel{}
	model.Group = idname.ParseIDNameSet(ctx, value.Group, diags)
	if model.Group.IsNull() {
		model.Group = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Site = idname.ParseIDNameSet(ctx, value.Site, diags)
	if model.Site.IsNull() {
		model.Site = types.SetValueMust(idname.RefObjectType, nil)
	}
	result, d := types.ObjectValueFrom(ctx, ruleSiteAttrTypes, model)
	diags.Append(d...)
	return result
}

type updateSource = cato.PolicySocketBypassUpdateRule_Policy_SocketBypass_UpdateRule_Rule_Rule_Source

func parseUpdateSource(
	ctx context.Context, value *updateSource, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.source object.")
		return types.ObjectNull(ruleSourceAttrTypes)
	}
	model := RuleSourceModel{}
	model.FloatingSubnet = idname.ParseIDNameSet(ctx, value.FloatingSubnet, diags)
	if model.FloatingSubnet.IsNull() {
		model.FloatingSubnet = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.GlobalIPRange = idname.ParseIDNameSet(ctx, value.GlobalIPRange, diags)
	if model.GlobalIPRange.IsNull() {
		model.GlobalIPRange = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Group = idname.ParseIDNameSet(ctx, value.Group, diags)
	if model.Group.IsNull() {
		model.Group = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Host = idname.ParseIDNameSet(ctx, value.Host, diags)
	if model.Host.IsNull() {
		model.Host = types.SetValueMust(idname.RefObjectType, nil)
	}
	var dIP diag.Diagnostics
	model.IP, dIP = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.IP...))
	diags.Append(dIP...)
	valuesIPRange := make([]attr.Value, 0, len(value.IPRange))
	for _, item := range value.IPRange {
		valuesIPRange = append(valuesIPRange, parseUpdateSourceIPRange(ctx, item, diags))
	}
	var dIPRange diag.Diagnostics
	model.IPRange, dIPRange = types.SetValue(types.ObjectType{AttrTypes: ruleSourceIPRangeAttrTypes}, valuesIPRange)
	diags.Append(dIPRange...)
	model.NetworkInterface = idname.ParseIDNameSet(ctx, value.NetworkInterface, diags)
	if model.NetworkInterface.IsNull() {
		model.NetworkInterface = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.SiteNetworkSubnet = idname.ParseIDNameSet(ctx, value.SiteNetworkSubnet, diags)
	if model.SiteNetworkSubnet.IsNull() {
		model.SiteNetworkSubnet = types.SetValueMust(idname.RefObjectType, nil)
	}
	var dSubnet diag.Diagnostics
	model.Subnet, dSubnet = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Subnet...))
	diags.Append(dSubnet...)
	model.Vlan = parseVLANs(ctx, value.Vlan, diags)
	result, d := types.ObjectValueFrom(ctx, ruleSourceAttrTypes, model)
	diags.Append(d...)
	return result
}

type updateSourceIPRange = cato.PolicySocketBypassUpdateRule_Policy_SocketBypass_UpdateRule_Rule_Rule_Source_IPRange

func parseUpdateSourceIPRange(
	ctx context.Context, value *updateSourceIPRange, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.source.ip_range object.")
		return types.ObjectNull(ruleSourceIPRangeAttrTypes)
	}
	model := RuleSourceIPRangeModel{}
	model.From = types.StringValue(value.From)
	model.To = types.StringValue(value.To)
	result, d := types.ObjectValueFrom(ctx, ruleSourceIPRangeAttrTypes, model)
	diags.Append(d...)
	return result
}

type move = cato.PolicySocketBypassMoveRule_Policy_SocketBypass_MoveRule_Rule_Rule

func parseMove(
	ctx context.Context, value *move, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule object.")
		return types.ObjectNull(ruleAttrTypes)
	}
	model := RuleModel{}
	model.Action = parseMoveAction(ctx, &value.Action, diags)
	model.Description = types.StringValue(value.Description)
	model.Destination = parseMoveDestination(ctx, &value.Destination, diags)
	model.Enabled = types.BoolValue(value.Enabled)
	valuesException := make([]attr.Value, 0, len(value.Exception))
	for _, item := range value.Exception {
		valuesException = append(valuesException, parseMoveException(ctx, item, diags))
	}
	var dException diag.Diagnostics
	model.Exception, dException = types.SetValue(types.ObjectType{AttrTypes: ruleExceptionAttrTypes}, valuesException)
	diags.Append(dException...)
	model.ID = types.StringValue(value.ID)
	model.Index = types.Int64Value(value.Index)
	model.Name = types.StringValue(value.Name)
	model.Section = parseMoveSection(ctx, &value.Section, diags)
	model.Service = parseMoveService(ctx, &value.Service, diags)
	model.Site = parseMoveSite(ctx, &value.Site, diags)
	model.Source = parseMoveSource(ctx, &value.Source, diags)
	result, d := types.ObjectValueFrom(ctx, ruleAttrTypes, model)
	diags.Append(d...)
	return result
}

type moveAction = cato.PolicySocketBypassMoveRule_Policy_SocketBypass_MoveRule_Rule_Rule_Action

func parseMoveAction(
	ctx context.Context, value *moveAction, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.action object.")
		return types.ObjectNull(ruleActionAttrTypes)
	}
	model := RuleActionModel{}
	model.Event = parseMoveActionEvent(ctx, &value.Event, diags)
	model.Port = types.StringValue(string(value.Port))
	result, d := types.ObjectValueFrom(ctx, ruleActionAttrTypes, model)
	diags.Append(d...)
	return result
}

type moveActionEvent = cato.PolicySocketBypassMoveRule_Policy_SocketBypass_MoveRule_Rule_Rule_Action_Event

func parseMoveActionEvent(
	ctx context.Context, value *moveActionEvent, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.action.event object.")
		return types.ObjectNull(ruleActionEventAttrTypes)
	}
	model := RuleActionEventModel{}
	model.Enabled = types.BoolValue(value.Enabled)
	result, d := types.ObjectValueFrom(ctx, ruleActionEventAttrTypes, model)
	diags.Append(d...)
	return result
}

type moveDst = cato.PolicySocketBypassMoveRule_Policy_SocketBypass_MoveRule_Rule_Rule_Destination

func parseMoveDestination(
	ctx context.Context, value *moveDst, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.destination object.")
		return types.ObjectNull(ruleDestinationAttrTypes)
	}
	model := RuleDestinationModel{}
	model.Application = idname.ParseIDNameSet(ctx, value.Application, diags)
	if model.Application.IsNull() {
		model.Application = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.CustomApp = idname.ParseIDNameSet(ctx, value.CustomApp, diags)
	if model.CustomApp.IsNull() {
		model.CustomApp = types.SetValueMust(idname.RefObjectType, nil)
	}
	var dDomain diag.Diagnostics
	model.Domain, dDomain = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Domain...))
	diags.Append(dDomain...)
	var dFqdn diag.Diagnostics
	model.Fqdn, dFqdn = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Fqdn...))
	diags.Append(dFqdn...)
	model.GlobalIPRange = idname.ParseIDNameSet(ctx, value.GlobalIPRange, diags)
	if model.GlobalIPRange.IsNull() {
		model.GlobalIPRange = types.SetValueMust(idname.RefObjectType, nil)
	}
	var dIP diag.Diagnostics
	model.IP, dIP = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.IP...))
	diags.Append(dIP...)
	valuesIPRange := make([]attr.Value, 0, len(value.IPRange))
	for _, item := range value.IPRange {
		valuesIPRange = append(valuesIPRange, parseMoveDestinationIPRange(ctx, item, diags))
	}
	var dIPRange diag.Diagnostics
	model.IPRange, dIPRange = types.SetValue(types.ObjectType{AttrTypes: ruleDestinationIPRangeAttrTypes}, valuesIPRange)
	diags.Append(dIPRange...)
	var dSubnet diag.Diagnostics
	model.Subnet, dSubnet = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Subnet...))
	diags.Append(dSubnet...)
	result, d := types.ObjectValueFrom(ctx, ruleDestinationAttrTypes, model)
	diags.Append(d...)
	return result
}

type moveDstIPRange = cato.PolicySocketBypassMoveRule_Policy_SocketBypass_MoveRule_Rule_Rule_Destination_IPRange

func parseMoveDestinationIPRange(
	ctx context.Context, value *moveDstIPRange, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.destination.ip_range object.")
		return types.ObjectNull(ruleDestinationIPRangeAttrTypes)
	}
	model := RuleDestinationIPRangeModel{}
	model.From = types.StringValue(value.From)
	model.To = types.StringValue(value.To)
	result, d := types.ObjectValueFrom(ctx, ruleDestinationIPRangeAttrTypes, model)
	diags.Append(d...)
	return result
}

type moveExc = cato.PolicySocketBypassMoveRule_Policy_SocketBypass_MoveRule_Rule_Rule_Exception

func parseMoveException(
	ctx context.Context, value *moveExc, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.exception object.")
		return types.ObjectNull(ruleExceptionAttrTypes)
	}
	model := RuleExceptionModel{}
	model.Destination = parseMoveExceptionDestination(ctx, &value.Destination, diags)
	model.Name = types.StringValue(value.Name)
	model.Service = parseMoveExceptionService(ctx, &value.Service, diags)
	model.Site = parseMoveExceptionSite(ctx, &value.Site, diags)
	model.Source = parseMoveExceptionSource(ctx, &value.Source, diags)
	result, d := types.ObjectValueFrom(ctx, ruleExceptionAttrTypes, model)
	diags.Append(d...)
	return result
}

type moveExcDst = cato.PolicySocketBypassMoveRule_Policy_SocketBypass_MoveRule_Rule_Rule_Exception_Destination

func parseMoveExceptionDestination(
	ctx context.Context, value *moveExcDst, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.exception.destination object.")
		return types.ObjectNull(ruleExceptionDestinationAttrTypes)
	}
	model := RuleExceptionDestinationModel{}
	model.Application = idname.ParseIDNameSet(ctx, value.Application, diags)
	if model.Application.IsNull() {
		model.Application = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.CustomApp = idname.ParseIDNameSet(ctx, value.CustomApp, diags)
	if model.CustomApp.IsNull() {
		model.CustomApp = types.SetValueMust(idname.RefObjectType, nil)
	}
	var dDomain diag.Diagnostics
	model.Domain, dDomain = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Domain...))
	diags.Append(dDomain...)
	var dFqdn diag.Diagnostics
	model.Fqdn, dFqdn = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Fqdn...))
	diags.Append(dFqdn...)
	model.GlobalIPRange = idname.ParseIDNameSet(ctx, value.GlobalIPRange, diags)
	if model.GlobalIPRange.IsNull() {
		model.GlobalIPRange = types.SetValueMust(idname.RefObjectType, nil)
	}
	var dIP diag.Diagnostics
	model.IP, dIP = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.IP...))
	diags.Append(dIP...)
	valuesIPRange := make([]attr.Value, 0, len(value.IPRange))
	for _, item := range value.IPRange {
		valuesIPRange = append(valuesIPRange, parseMoveExceptionDestinationIPRange(ctx, item, diags))
	}
	var dIPRange diag.Diagnostics
	model.IPRange, dIPRange = types.SetValue(types.ObjectType{AttrTypes: ruleExceptionDestinationIPRangeAttrTypes}, valuesIPRange)
	diags.Append(dIPRange...)
	var dSubnet diag.Diagnostics
	model.Subnet, dSubnet = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Subnet...))
	diags.Append(dSubnet...)
	result, d := types.ObjectValueFrom(ctx, ruleExceptionDestinationAttrTypes, model)
	diags.Append(d...)
	return result
}

type moveExcDstIPRange = cato.PolicySocketBypassMoveRule_Policy_SocketBypass_MoveRule_Rule_Rule_Exception_Destination_IPRange

func parseMoveExceptionDestinationIPRange(
	ctx context.Context, value *moveExcDstIPRange, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.exception.destination.ip_range object.")
		return types.ObjectNull(ruleExceptionDestinationIPRangeAttrTypes)
	}
	model := RuleExceptionDestinationIPRangeModel{}
	model.From = types.StringValue(value.From)
	model.To = types.StringValue(value.To)
	result, d := types.ObjectValueFrom(ctx, ruleExceptionDestinationIPRangeAttrTypes, model)
	diags.Append(d...)
	return result
}

type moveExcSvc = cato.PolicySocketBypassMoveRule_Policy_SocketBypass_MoveRule_Rule_Rule_Exception_Service

func parseMoveExceptionService(
	ctx context.Context, value *moveExcSvc, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.exception.service object.")
		return types.ObjectNull(ruleExceptionServiceAttrTypes)
	}
	model := RuleExceptionServiceModel{}
	valuesCustom := make([]attr.Value, 0, len(value.Custom))
	for _, item := range value.Custom {
		valuesCustom = append(valuesCustom, parseMoveExceptionServiceCustom(ctx, item, diags))
	}
	var dCustom diag.Diagnostics
	model.Custom, dCustom = types.SetValue(types.ObjectType{AttrTypes: ruleExceptionServiceCustomAttrTypes}, valuesCustom)
	diags.Append(dCustom...)
	valuesSimple := make([]attr.Value, 0, len(value.Simple))
	for _, item := range value.Simple {
		valuesSimple = append(valuesSimple, parseMoveExceptionServiceSimple(ctx, item, diags))
	}
	var dSimple diag.Diagnostics
	model.Simple, dSimple = types.SetValue(types.ObjectType{AttrTypes: ruleExceptionServiceSimpleAttrTypes}, valuesSimple)
	diags.Append(dSimple...)
	result, d := types.ObjectValueFrom(ctx, ruleExceptionServiceAttrTypes, model)
	diags.Append(d...)
	return result
}

type moveExcSvcCustom = cato.PolicySocketBypassMoveRule_Policy_SocketBypass_MoveRule_Rule_Rule_Exception_Service_Custom

func parseMoveExceptionServiceCustom(
	ctx context.Context, value *moveExcSvcCustom, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.exception.service.custom object.")
		return types.ObjectNull(ruleExceptionServiceCustomAttrTypes)
	}
	model := RuleExceptionServiceCustomModel{}
	model.Port = parsePorts(ctx, value.Port, diags)
	if value.PortRange == nil {
		model.PortRange = types.ObjectNull(ruleExceptionServiceCustomPortRangeAttrTypes)
	} else {
		model.PortRange = parseMoveExceptionServiceCustomPortRange(ctx, value.PortRange, diags)
	}
	model.Protocol = types.StringValue(string(value.Protocol))
	result, d := types.ObjectValueFrom(ctx, ruleExceptionServiceCustomAttrTypes, model)
	diags.Append(d...)
	return result
}

type moveExcSvcCustomPortRange = cato.PolicySocketBypassMoveRule_Policy_SocketBypass_MoveRule_Rule_Rule_Exception_Service_Custom_PortRange

func parseMoveExceptionServiceCustomPortRange(
	ctx context.Context, value *moveExcSvcCustomPortRange, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.exception.service.custom.port_range object.")
		return types.ObjectNull(ruleExceptionServiceCustomPortRangeAttrTypes)
	}
	model := RuleExceptionServiceCustomPortRangeModel{}
	model.From = types.StringValue(string(value.From))
	model.To = types.StringValue(string(value.To))
	result, d := types.ObjectValueFrom(ctx, ruleExceptionServiceCustomPortRangeAttrTypes, model)
	diags.Append(d...)
	return result
}

type moveExcSvcSimple = cato.PolicySocketBypassMoveRule_Policy_SocketBypass_MoveRule_Rule_Rule_Exception_Service_Simple

func parseMoveExceptionServiceSimple(
	ctx context.Context, value *moveExcSvcSimple, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.exception.service.simple object.")
		return types.ObjectNull(ruleExceptionServiceSimpleAttrTypes)
	}
	model := RuleExceptionServiceSimpleModel{}
	model.Name = types.StringValue(string(value.Name))
	result, d := types.ObjectValueFrom(ctx, ruleExceptionServiceSimpleAttrTypes, model)
	diags.Append(d...)
	return result
}

type moveExcSite = cato.PolicySocketBypassMoveRule_Policy_SocketBypass_MoveRule_Rule_Rule_Exception_Site

func parseMoveExceptionSite(
	ctx context.Context, value *moveExcSite, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.exception.site object.")
		return types.ObjectNull(ruleExceptionSiteAttrTypes)
	}
	model := RuleExceptionSiteModel{}
	model.Group = idname.ParseIDNameSet(ctx, value.Group, diags)
	if model.Group.IsNull() {
		model.Group = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Site = idname.ParseIDNameSet(ctx, value.Site, diags)
	if model.Site.IsNull() {
		model.Site = types.SetValueMust(idname.RefObjectType, nil)
	}
	result, d := types.ObjectValueFrom(ctx, ruleExceptionSiteAttrTypes, model)
	diags.Append(d...)
	return result
}

type moveExcSource = cato.PolicySocketBypassMoveRule_Policy_SocketBypass_MoveRule_Rule_Rule_Exception_Source

func parseMoveExceptionSource(
	ctx context.Context, value *moveExcSource, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.exception.source object.")
		return types.ObjectNull(ruleExceptionSourceAttrTypes)
	}
	model := RuleExceptionSourceModel{}
	model.FloatingSubnet = idname.ParseIDNameSet(ctx, value.FloatingSubnet, diags)
	if model.FloatingSubnet.IsNull() {
		model.FloatingSubnet = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.GlobalIPRange = idname.ParseIDNameSet(ctx, value.GlobalIPRange, diags)
	if model.GlobalIPRange.IsNull() {
		model.GlobalIPRange = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Group = idname.ParseIDNameSet(ctx, value.Group, diags)
	if model.Group.IsNull() {
		model.Group = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Host = idname.ParseIDNameSet(ctx, value.Host, diags)
	if model.Host.IsNull() {
		model.Host = types.SetValueMust(idname.RefObjectType, nil)
	}
	var dIP diag.Diagnostics
	model.IP, dIP = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.IP...))
	diags.Append(dIP...)
	valuesIPRange := make([]attr.Value, 0, len(value.IPRange))
	for _, item := range value.IPRange {
		valuesIPRange = append(valuesIPRange, parseMoveExceptionSourceIPRange(ctx, item, diags))
	}
	var dIPRange diag.Diagnostics
	model.IPRange, dIPRange = types.SetValue(types.ObjectType{AttrTypes: ruleExceptionSourceIPRangeAttrTypes}, valuesIPRange)
	diags.Append(dIPRange...)
	model.NetworkInterface = idname.ParseIDNameSet(ctx, value.NetworkInterface, diags)
	if model.NetworkInterface.IsNull() {
		model.NetworkInterface = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.SiteNetworkSubnet = idname.ParseIDNameSet(ctx, value.SiteNetworkSubnet, diags)
	if model.SiteNetworkSubnet.IsNull() {
		model.SiteNetworkSubnet = types.SetValueMust(idname.RefObjectType, nil)
	}
	var dSubnet diag.Diagnostics
	model.Subnet, dSubnet = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Subnet...))
	diags.Append(dSubnet...)
	model.Vlan = parseVLANs(ctx, value.Vlan, diags)
	result, d := types.ObjectValueFrom(ctx, ruleExceptionSourceAttrTypes, model)
	diags.Append(d...)
	return result
}

type moveExcSourceIPRange = cato.PolicySocketBypassMoveRule_Policy_SocketBypass_MoveRule_Rule_Rule_Exception_Source_IPRange

func parseMoveExceptionSourceIPRange(
	ctx context.Context, value *moveExcSourceIPRange, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.exception.source.ip_range object.")
		return types.ObjectNull(ruleExceptionSourceIPRangeAttrTypes)
	}
	model := RuleExceptionSourceIPRangeModel{}
	model.From = types.StringValue(value.From)
	model.To = types.StringValue(value.To)
	result, d := types.ObjectValueFrom(ctx, ruleExceptionSourceIPRangeAttrTypes, model)
	diags.Append(d...)
	return result
}

type moveSection = cato.PolicySocketBypassMoveRule_Policy_SocketBypass_MoveRule_Rule_Rule_Section

func parseMoveSection(
	ctx context.Context, value *moveSection, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.section object.")
		return types.ObjectNull(ruleSectionAttrTypes)
	}
	return idname.ParseIDName(ctx, *value, diags)
}

type moveSvc = cato.PolicySocketBypassMoveRule_Policy_SocketBypass_MoveRule_Rule_Rule_Service

func parseMoveService(
	ctx context.Context, value *moveSvc, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.service object.")
		return types.ObjectNull(ruleServiceAttrTypes)
	}
	model := RuleServiceModel{}
	valuesCustom := make([]attr.Value, 0, len(value.Custom))
	for _, item := range value.Custom {
		valuesCustom = append(valuesCustom, parseMoveServiceCustom(ctx, item, diags))
	}
	var dCustom diag.Diagnostics
	model.Custom, dCustom = types.SetValue(types.ObjectType{AttrTypes: ruleServiceCustomAttrTypes}, valuesCustom)
	diags.Append(dCustom...)
	valuesSimple := make([]attr.Value, 0, len(value.Simple))
	for _, item := range value.Simple {
		valuesSimple = append(valuesSimple, parseMoveServiceSimple(ctx, item, diags))
	}
	var dSimple diag.Diagnostics
	model.Simple, dSimple = types.SetValue(types.ObjectType{AttrTypes: ruleServiceSimpleAttrTypes}, valuesSimple)
	diags.Append(dSimple...)
	result, d := types.ObjectValueFrom(ctx, ruleServiceAttrTypes, model)
	diags.Append(d...)
	return result
}

type moveSvcCustom = cato.PolicySocketBypassMoveRule_Policy_SocketBypass_MoveRule_Rule_Rule_Service_Custom

func parseMoveServiceCustom(
	ctx context.Context, value *moveSvcCustom, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.service.custom object.")
		return types.ObjectNull(ruleServiceCustomAttrTypes)
	}
	model := RuleServiceCustomModel{}
	model.Port = parsePorts(ctx, value.Port, diags)
	if value.PortRange == nil {
		model.PortRange = types.ObjectNull(ruleServiceCustomPortRangeAttrTypes)
	} else {
		model.PortRange = parseMoveServiceCustomPortRange(ctx, value.PortRange, diags)
	}
	model.Protocol = types.StringValue(string(value.Protocol))
	result, d := types.ObjectValueFrom(ctx, ruleServiceCustomAttrTypes, model)
	diags.Append(d...)
	return result
}

type moveSvcCustomPortRange = cato.PolicySocketBypassMoveRule_Policy_SocketBypass_MoveRule_Rule_Rule_Service_Custom_PortRange

func parseMoveServiceCustomPortRange(
	ctx context.Context, value *moveSvcCustomPortRange, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.service.custom.port_range object.")
		return types.ObjectNull(ruleServiceCustomPortRangeAttrTypes)
	}
	model := RuleServiceCustomPortRangeModel{}
	model.From = types.StringValue(string(value.From))
	model.To = types.StringValue(string(value.To))
	result, d := types.ObjectValueFrom(ctx, ruleServiceCustomPortRangeAttrTypes, model)
	diags.Append(d...)
	return result
}

type moveSvcSimple = cato.PolicySocketBypassMoveRule_Policy_SocketBypass_MoveRule_Rule_Rule_Service_Simple

func parseMoveServiceSimple(
	ctx context.Context, value *moveSvcSimple, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.service.simple object.")
		return types.ObjectNull(ruleServiceSimpleAttrTypes)
	}
	model := RuleServiceSimpleModel{}
	model.Name = types.StringValue(string(value.Name))
	result, d := types.ObjectValueFrom(ctx, ruleServiceSimpleAttrTypes, model)
	diags.Append(d...)
	return result
}

type moveSite = cato.PolicySocketBypassMoveRule_Policy_SocketBypass_MoveRule_Rule_Rule_Site

func parseMoveSite(
	ctx context.Context, value *moveSite, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.site object.")
		return types.ObjectNull(ruleSiteAttrTypes)
	}
	model := RuleSiteModel{}
	model.Group = idname.ParseIDNameSet(ctx, value.Group, diags)
	if model.Group.IsNull() {
		model.Group = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Site = idname.ParseIDNameSet(ctx, value.Site, diags)
	if model.Site.IsNull() {
		model.Site = types.SetValueMust(idname.RefObjectType, nil)
	}
	result, d := types.ObjectValueFrom(ctx, ruleSiteAttrTypes, model)
	diags.Append(d...)
	return result
}

type moveSource = cato.PolicySocketBypassMoveRule_Policy_SocketBypass_MoveRule_Rule_Rule_Source

func parseMoveSource(
	ctx context.Context, value *moveSource, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.source object.")
		return types.ObjectNull(ruleSourceAttrTypes)
	}
	model := RuleSourceModel{}
	model.FloatingSubnet = idname.ParseIDNameSet(ctx, value.FloatingSubnet, diags)
	if model.FloatingSubnet.IsNull() {
		model.FloatingSubnet = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.GlobalIPRange = idname.ParseIDNameSet(ctx, value.GlobalIPRange, diags)
	if model.GlobalIPRange.IsNull() {
		model.GlobalIPRange = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Group = idname.ParseIDNameSet(ctx, value.Group, diags)
	if model.Group.IsNull() {
		model.Group = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.Host = idname.ParseIDNameSet(ctx, value.Host, diags)
	if model.Host.IsNull() {
		model.Host = types.SetValueMust(idname.RefObjectType, nil)
	}
	var dIP diag.Diagnostics
	model.IP, dIP = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.IP...))
	diags.Append(dIP...)
	valuesIPRange := make([]attr.Value, 0, len(value.IPRange))
	for _, item := range value.IPRange {
		valuesIPRange = append(valuesIPRange, parseMoveSourceIPRange(ctx, item, diags))
	}
	var dIPRange diag.Diagnostics
	model.IPRange, dIPRange = types.SetValue(types.ObjectType{AttrTypes: ruleSourceIPRangeAttrTypes}, valuesIPRange)
	diags.Append(dIPRange...)
	model.NetworkInterface = idname.ParseIDNameSet(ctx, value.NetworkInterface, diags)
	if model.NetworkInterface.IsNull() {
		model.NetworkInterface = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.SiteNetworkSubnet = idname.ParseIDNameSet(ctx, value.SiteNetworkSubnet, diags)
	if model.SiteNetworkSubnet.IsNull() {
		model.SiteNetworkSubnet = types.SetValueMust(idname.RefObjectType, nil)
	}
	var dSubnet diag.Diagnostics
	model.Subnet, dSubnet = types.SetValueFrom(ctx, types.StringType, append([]string{}, value.Subnet...))
	diags.Append(dSubnet...)
	model.Vlan = parseVLANs(ctx, value.Vlan, diags)
	result, d := types.ObjectValueFrom(ctx, ruleSourceAttrTypes, model)
	diags.Append(d...)
	return result
}

type moveSourceIPRange = cato.PolicySocketBypassMoveRule_Policy_SocketBypass_MoveRule_Rule_Rule_Source_IPRange

func parseMoveSourceIPRange(
	ctx context.Context, value *moveSourceIPRange, diags *diag.Diagnostics,
) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.source.ip_range object.")
		return types.ObjectNull(ruleSourceIPRangeAttrTypes)
	}
	model := RuleSourceIPRangeModel{}
	model.From = types.StringValue(value.From)
	model.To = types.StringValue(value.To)
	result, d := types.ObjectValueFrom(ctx, ruleSourceIPRangeAttrTypes, model)
	diags.Append(d...)
	return result
}

func buildVLANs(ctx context.Context, value types.Set, diags *diag.Diagnostics) []scalars.Vlan {
	var numbers []int64
	diags.Append(value.ElementsAs(ctx, &numbers, false)...)
	result := make([]scalars.Vlan, 0, len(numbers))
	for _, number := range numbers {
		result = append(result, scalars.Vlan(number))
	}
	return result
}

func parseVLANs(ctx context.Context, values []scalars.Vlan, diags *diag.Diagnostics) types.Set {
	return convert.Int64SetFunc(ctx, append([]scalars.Vlan{}, values...), func(value scalars.Vlan) int64 { return int64(value) }, diags)
}

func parsePorts(ctx context.Context, values []scalars.Port, diags *diag.Diagnostics) types.Set {
	numbers := make([]string, 0, len(values))
	for _, value := range values {
		numbers = append(numbers, string(value))
	}
	result, diagnostics := types.SetValueFrom(ctx, types.StringType, numbers)
	diags.Append(diagnostics...)
	return result
}
