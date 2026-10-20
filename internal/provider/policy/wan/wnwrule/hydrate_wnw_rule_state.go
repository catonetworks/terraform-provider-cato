package wnwrule

import (
	"context"
	"fmt"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/customservice"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/iprange"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/oldidname"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/portrange"
)

//nolint:funlen
func hydrateWanNetworkRuleState(
	ctx context.Context,
	state Rule,
	currentRule *cato_go_sdk.WanNetworkPolicy_Policy_WanNetwork_Policy_Rules_Rule,
) (PolicyPolicyWanNetworkPolicyRulesRule, diag.Diagnostics) {
	ruleInput := PolicyPolicyWanNetworkPolicyRulesRule{}
	diags := make(diag.Diagnostics, 0)

	// Handle case where state might be incomplete (e.g., during import)
	if !state.Rule.IsNull() && !state.Rule.IsUnknown() {
		diagstmp := state.Rule.As(ctx, &ruleInput, basetypes.ObjectAsOptions{})
		if diagstmp.HasError() {
			tflog.Debug(ctx, "Failed to convert state to struct in hydrateWanNetworkRuleState, using empty struct for hydration")
			ruleInput = PolicyPolicyWanNetworkPolicyRulesRule{}
		} else {
			diags = append(diags, diagstmp...)
		}
	}

	// Basic rule fields
	ruleInput.Name = types.StringValue(currentRule.Name)
	if currentRule.Description == "" {
		ruleInput.Description = types.StringNull()
	} else {
		ruleInput.Description = types.StringValue(currentRule.Description)
	}
	ruleInput.ID = types.StringValue(currentRule.ID)
	ruleInput.Enabled = types.BoolValue(currentRule.Enabled)
	ruleInput.RuleType = types.StringValue(currentRule.RuleType.String())
	ruleInput.RouteType = types.StringValue(currentRule.RouteType.String())

	// /// /// /// / Rule -> Source // /// /// /// /// /
	curRuleSourceObj, diagstmp := types.ObjectValue(
		WanNetworkSourceAttrTypes,
		map[string]attr.Value{
			"ip": convert.ParseList(ctx, types.StringType, currentRule.Source.IP,
				"rule.source.ip"),
			"host": oldidname.ParseNameIDList(ctx, currentRule.Source.Host,
				"rule.source.host"),
			"site": oldidname.ParseNameIDList(ctx, currentRule.Source.Site,
				"rule.source.site"),
			"subnet": convert.ParseList(ctx, types.StringType, currentRule.Source.Subnet,
				"rule.source.subnet"),
			"ip_range": convert.ParseFromToList(ctx, currentRule.Source.IPRangeWanNetworkRuleSource,
				"rule.source.ip_range"),
			"global_ip_range": oldidname.ParseNameIDList(ctx, currentRule.Source.GlobalIPRange,
				"rule.source.global_ip_range"),
			"network_interface": oldidname.ParseNameIDList(ctx, currentRule.Source.NetworkInterface,
				"rule.source.network_interface"),
			"site_network_subnet": oldidname.ParseNameIDList(ctx, currentRule.Source.SiteNetworkSubnet,
				"rule.source.site_network_subnet"),
			"floating_subnet": oldidname.ParseNameIDList(ctx, currentRule.Source.FloatingSubnet,
				"rule.source.floating_subnet"),
			"user": oldidname.ParseNameIDList(ctx, currentRule.Source.User,
				"rule.source.user"),
			"users_group": oldidname.ParseNameIDList(ctx, currentRule.Source.UsersGroup,
				"rule.source.users_group"),
			"group": oldidname.ParseNameIDList(ctx, currentRule.Source.Group,
				"rule.source.group"),
			"system_group": oldidname.ParseNameIDList(ctx, currentRule.Source.SystemGroup,
				"rule.source.system_group"),
		},
	)
	diags = append(diags, diagstmp...)

	// Ensure source.network_interface is never unknown after apply
	srcAttrs := curRuleSourceObj.Attributes()
	if v, ok := srcAttrs["network_interface"]; ok {
		if setVal, ok2 := v.(types.Set); ok2 {
			if setVal.IsUnknown() {
				srcAttrs["network_interface"] = types.SetNull(idname.RefObjectType)
				var diagsTmp diag.Diagnostics
				curRuleSourceObj, diagsTmp = types.ObjectValue(curRuleSourceObj.Type(ctx).(types.ObjectType).AttrTypes, srcAttrs)
				diags = append(diags, diagsTmp...)
			}
		}
	}

	ruleInput.Source = curRuleSourceObj
	// /// /// /// /// end rule.source // /// /// /// /// /

	// /// /// /// / Rule -> Destination // /// /// /// /// /
	curRuleDestinationObj, diagstmp := types.ObjectValue(
		WanNetworkDestAttrTypes,
		map[string]attr.Value{
			"ip": convert.ParseList(ctx, types.StringType, currentRule.Destination.IP,
				"rule.destination.ip"),
			"host": oldidname.ParseNameIDList(ctx, currentRule.Destination.Host,
				"rule.destination.host"),
			"site": oldidname.ParseNameIDList(ctx, currentRule.Destination.Site,
				"rule.destination.site"),
			"subnet": convert.ParseList(ctx, types.StringType, currentRule.Destination.Subnet,
				"rule.destination.subnet"),
			"ip_range": convert.ParseFromToList(ctx, currentRule.Destination.IPRangeWanNetworkRuleDestination,
				"rule.destination.ip_range"),
			"global_ip_range": oldidname.ParseNameIDList(ctx, currentRule.Destination.GlobalIPRange,
				"rule.destination.global_ip_range"),
			"network_interface": oldidname.ParseNameIDList(ctx, currentRule.Destination.NetworkInterface,
				"rule.destination.network_interface"),
			"site_network_subnet": oldidname.ParseNameIDList(ctx, currentRule.Destination.SiteNetworkSubnet,
				"rule.destination.site_network_subnet"),
			"floating_subnet": oldidname.ParseNameIDList(ctx, currentRule.Destination.FloatingSubnet,
				"rule.destination.floating_subnet"),
			"user": oldidname.ParseNameIDList(ctx, currentRule.Destination.User,
				"rule.destination.user"),
			"users_group": oldidname.ParseNameIDList(ctx, currentRule.Destination.UsersGroup,
				"rule.destination.users_group"),
			"group": oldidname.ParseNameIDList(ctx, currentRule.Destination.Group,
				"rule.destination.group"),
			"system_group": oldidname.ParseNameIDList(ctx, currentRule.Destination.SystemGroup,
				"rule.destination.system_group"),
		},
	)
	diags = append(diags, diagstmp...)
	ruleInput.Destination = curRuleDestinationObj
	// /// /// /// /// end Rule -> Destination // /// /// /// /// /

	// /// /// /// / Rule -> Application // /// /// /// /// /
	curRuleApplicationObj, diagstmp := types.ObjectValue(
		WanNetworkApplicationAttrTypes,
		map[string]attr.Value{
			"application": oldidname.ParseNameIDList(ctx, currentRule.Application.Application,
				"rule.application.application"),
			"custom_app": oldidname.ParseNameIDList(ctx, currentRule.Application.CustomApp,
				"rule.application.custom_app"),
			"app_category": oldidname.ParseNameIDList(ctx, currentRule.Application.AppCategory,
				"rule.application.app_category"),
			"custom_category": oldidname.ParseNameIDList(ctx, currentRule.Application.CustomCategory,
				"rule.application.custom_category"),
			"domain": convert.ParseList(ctx, types.StringType, currentRule.Application.Domain,
				"rule.application.domain"),
			"fqdn": convert.ParseList(ctx, types.StringType, currentRule.Application.Fqdn,
				"rule.application.fqdn"),
			"service": oldidname.ParseNameIDList(ctx, currentRule.Application.Service,
				"rule.application.service"),
			"custom_service": parseWanNetworkCustomServiceListSDK(ctx, currentRule.Application.CustomService,
				"rule.application.custom_service"),
			"custom_service_ip": parseWanNetworkCustomServiceIPListSDK(
				ctx,
				currentRule.Application.CustomServiceIP,
				"rule.application.custom_service_ip",
			),
		},
	)
	ruleInput.Application = curRuleApplicationObj
	diags = append(diags, diagstmp...)
	// /// /// /// /// end Rule -> Application // /// /// /// /// /

	// /// /// /// /// start Rule -> Configuration // /// /// /// /// /
	curRuleConfigurationObj, diagstmp := parseWanNetworkConfigurationSDK(ctx, currentRule.Configuration,
		"rule.configuration")
	diags = append(diags, diagstmp...)
	ruleInput.Configuration = curRuleConfigurationObj
	// /// /// /// /// end Rule -> Configuration // /// /// /// /// /

	// /// /// /// /// start Rule -> BandwidthPriority // /// /// /// /// /
	// BandwidthPriority is a struct, not a pointer in SDK type
	if currentRule.BandwidthPriority.ID != "" {
		bandwidthPriorityObj, diagstmp := types.ObjectValue(
			BandwidthPriorityAttrTypes,
			map[string]attr.Value{
				"id":   types.StringValue(currentRule.BandwidthPriority.ID),
				"name": types.StringValue(currentRule.BandwidthPriority.Name),
			},
		)
		diags = append(diags, diagstmp...)
		ruleInput.BandwidthPriority = bandwidthPriorityObj
	} else {
		ruleInput.BandwidthPriority = types.ObjectNull(BandwidthPriorityAttrTypes)
	}
	// /// /// /// /// end Rule -> BandwidthPriority // /// /// /// /// /

	// /// /// /// /// // start Rule -> Exceptions // /// /// /// /// /
	exceptions := []attr.Value{}

	tflog.Warn(ctx, "hydrateWanNetworkRuleState() currentRule.Exceptions", map[string]interface{}{
		"exceptions": currentRule.Exceptions,
		"len":        len(currentRule.Exceptions),
	})
	if len(currentRule.Exceptions) > 0 {
		for _, ruleException := range currentRule.Exceptions {
			// Rule -> Exceptions -> Source
			curExceptionSourceObj, diagstmp := types.ObjectValue(
				WanNetworkSourceAttrTypes,
				map[string]attr.Value{
					"ip": convert.ParseList(ctx, types.StringType, ruleException.Source.IP,
						"rule.exception.source.ip"),
					"host": oldidname.ParseNameIDList(ctx, ruleException.Source.Host,
						"rule.exception.source.host"),
					"site": oldidname.ParseNameIDList(ctx, ruleException.Source.Site,
						"rule.exception.source.site"),
					"subnet": convert.ParseList(ctx, types.StringType, ruleException.Source.Subnet,
						"rule.exception.source.subnet"),
					"ip_range": convert.ParseFromToList(ctx, ruleException.Source.IPRange,
						"rule.exception.source.ip_range"),
					"global_ip_range": oldidname.ParseNameIDList(ctx, ruleException.Source.GlobalIPRange,
						"rule.exception.source.global_ip_range"),
					"network_interface": oldidname.ParseNameIDList(ctx, ruleException.Source.NetworkInterface,
						"rule.exception.source.network_interface"),
					"site_network_subnet": oldidname.ParseNameIDList(ctx, ruleException.Source.SiteNetworkSubnet,
						"rule.exception.source.site_network_subnet"),
					"floating_subnet": oldidname.ParseNameIDList(ctx, ruleException.Source.FloatingSubnet,
						"rule.exception.source.floating_subnet"),
					"user": oldidname.ParseNameIDList(ctx, ruleException.Source.User,
						"rule.exception.source.user"),
					"users_group": oldidname.ParseNameIDList(ctx, ruleException.Source.UsersGroup,
						"rule.exception.source.users_group"),
					"group": oldidname.ParseNameIDList(ctx, ruleException.Source.Group,
						"rule.exception.source.group"),
					"system_group": oldidname.ParseNameIDList(ctx, ruleException.Source.SystemGroup,
						"rule.exception.source.system_group"),
				},
			)
			diags = append(diags, diagstmp...)

			// Rule -> Exceptions -> Destination
			curExceptionDestObj, diagstmp := types.ObjectValue(
				WanNetworkDestAttrTypes,
				map[string]attr.Value{
					"ip": convert.ParseList(ctx, types.StringType, ruleException.Destination.IP,
						"rule.exception.destination.ip"),
					"host": oldidname.ParseNameIDList(ctx, ruleException.Destination.Host,
						"rule.exception.destination.host"),
					"site": oldidname.ParseNameIDList(ctx, ruleException.Destination.Site,
						"rule.exception.destination.site"),
					"subnet": convert.ParseList(ctx, types.StringType, ruleException.Destination.Subnet,
						"rule.exception.destination.subnet"),
					"ip_range": convert.ParseFromToList(ctx, ruleException.Destination.IPRange,
						"rule.exception.destination.ip_range"),
					"global_ip_range": oldidname.ParseNameIDList(ctx, ruleException.Destination.GlobalIPRange,
						"rule.exception.destination.global_ip_range"),
					"network_interface": oldidname.ParseNameIDList(
						ctx,
						ruleException.Destination.NetworkInterface,
						"rule.exception.destination.network_interface",
					),
					"site_network_subnet": oldidname.ParseNameIDList(
						ctx,
						ruleException.Destination.SiteNetworkSubnet,
						"rule.exception.destination.site_network_subnet",
					),
					"floating_subnet": oldidname.ParseNameIDList(ctx, ruleException.Destination.FloatingSubnet,
						"rule.exception.destination.floating_subnet"),
					"user": oldidname.ParseNameIDList(ctx, ruleException.Destination.User,
						"rule.exception.destination.user"),
					"users_group": oldidname.ParseNameIDList(ctx, ruleException.Destination.UsersGroup,
						"rule.exception.destination.users_group"),
					"group": oldidname.ParseNameIDList(ctx, ruleException.Destination.Group,
						"rule.exception.destination.group"),
					"system_group": oldidname.ParseNameIDList(ctx, ruleException.Destination.SystemGroup,
						"rule.exception.destination.system_group"),
				},
			)
			diags = append(diags, diagstmp...)

			// // /// /// /// / Rule -> Exceptions -> Application // /// /// /// /// /
			curExceptionApplicationObj, diagstmp := types.ObjectValue(
				WanNetworkExceptionApplicationAttrTypes,
				map[string]attr.Value{
					"application": oldidname.ParseNameIDList(ctx, ruleException.Application.Application,
						"rule.exception.application.application"),
					"custom_app": oldidname.ParseNameIDList(ctx, ruleException.Application.CustomApp,
						"rule.exception.application.custom_app"),
					"app_category": oldidname.ParseNameIDList(ctx, ruleException.Application.AppCategory,
						"rule.exception.application.app_category"),
					"custom_category": oldidname.ParseNameIDList(ctx, ruleException.Application.CustomCategory,
						"rule.exception.application.custom_category"),
					"domain": convert.ParseList(ctx, types.StringType, ruleException.Application.Domain,
						"rule.exception.application.domain"),
					"fqdn": convert.ParseList(ctx, types.StringType, ruleException.Application.Fqdn,
						"rule.exception.application.fqdn"),
					"service": oldidname.ParseNameIDList(ctx, ruleException.Application.Service,
						"rule.exception.application.service"),
					"custom_service": parseWanNetworkExceptionCustomServiceListSDK(
						ctx,
						ruleException.Application.CustomService,
						"rule.exception.application.custom_service",
					),
					"custom_service_ip": parseWanNetworkExceptionCustomServiceIPListSDK(
						ctx,
						ruleException.Application.CustomServiceIP,
						"rule.exception.application.custom_service_ip",
					),
				},
			)
			diags = append(diags, diagstmp...)
			// /// /// /// /// end Rule -> Exceptions -> Application // /// /// /// /// /

			// Initialize Exception object with populated values
			curException, diagstmp := types.ObjectValue(
				WanNetworkExceptionAttrTypes,
				map[string]attr.Value{
					"name":        types.StringValue(ruleException.Name),
					"source":      curExceptionSourceObj,
					"destination": curExceptionDestObj,
					"application": curExceptionApplicationObj,
				},
			)
			diags = append(diags, diagstmp...)
			exceptions = append(exceptions, curException)
		}
		curRuleExceptionsObj, diagstmp := types.SetValue(WanNetworkExceptionObjectType, exceptions)
		diags = append(diags, diagstmp...)
		ruleInput.Exceptions = curRuleExceptionsObj
	} else {
		// Use empty set instead of null to match schema expectations
		curRuleExceptionsObj, diagstmp := types.SetValue(WanNetworkExceptionObjectType, []attr.Value{})
		diags = append(diags, diagstmp...)
		ruleInput.Exceptions = curRuleExceptionsObj
	}
	// /// /// /// /// end Rule -> Exceptions // /// /// /// /// /

	return ruleInput, diags
}

// SDK-specific parse functions for WAN Network query responses

// parseWanNetworkCustomServiceListSDK parses custom service list from SDK types
func parseWanNetworkCustomServiceListSDK(
	ctx context.Context,
	items []*cato_go_sdk.WanNetworkPolicy_Policy_WanNetwork_Policy_Rules_Rule_Application_CustomService,
	_ string,
) types.List {
	if len(items) == 0 {
		return types.ListNull(customservice.CustomServiceObjectType)
	}

	var customServices []types.Object
	for _, item := range items {
		attrs := map[string]attr.Value{
			"port":       types.ListNull(types.StringType),
			"port_range": types.ObjectNull(portrange.PortRangeAttrTypes),
			"protocol":   types.StringNull(),
		}

		if len(item.Port) > 0 {
			var ports []attr.Value
			for _, p := range item.Port {
				ports = append(ports, types.StringValue(fmt.Sprintf("%v", p)))
			}
			attrs["port"], _ = types.ListValue(types.StringType, ports)
		}

		if item.PortRange != nil && (item.PortRange.From != "" || item.PortRange.To != "") {
			attrs["port_range"], _ = types.ObjectValue(
				portrange.PortRangeAttrTypes,
				map[string]attr.Value{
					"from": types.StringValue(fmt.Sprintf("%v", item.PortRange.From)),
					"to":   types.StringValue(fmt.Sprintf("%v", item.PortRange.To)),
				},
			)
		}

		if item.Protocol != "" {
			attrs["protocol"] = types.StringValue(string(item.Protocol))
		}

		obj, _ := types.ObjectValue(WanNetworkCustomServiceAttrTypes, attrs)
		customServices = append(customServices, obj)
	}

	listValue, _ := types.ListValueFrom(ctx, customservice.CustomServiceObjectType, customServices)
	return listValue
}

// parseWanNetworkCustomServiceIPListSDK parses custom service IP list from SDK types
func parseWanNetworkCustomServiceIPListSDK(
	ctx context.Context,
	items []*cato_go_sdk.WanNetworkPolicy_Policy_WanNetwork_Policy_Rules_Rule_Application_CustomServiceIP,
	_ string,
) types.List {
	if len(items) == 0 {
		return types.ListNull(WanNetworkCustomServiceIPObjectType)
	}

	var customServiceIps []types.Object
	for _, item := range items {
		attrs := map[string]attr.Value{
			"name":     types.StringValue(item.Name),
			"ip":       types.StringNull(),
			"ip_range": types.ObjectNull(iprange.RangeAttrTypes),
		}

		if item.IP != nil && *item.IP != "" {
			attrs["ip"] = types.StringValue(*item.IP)
		}

		if item.IPRange != nil {
			attrs["ip_range"], _ = types.ObjectValue(
				iprange.RangeAttrTypes,
				map[string]attr.Value{
					"from": types.StringValue(item.IPRange.From),
					"to":   types.StringValue(item.IPRange.To),
				},
			)
		}

		obj, _ := types.ObjectValue(WanNetworkCustomServiceIPAttrTypes, attrs)
		customServiceIps = append(customServiceIps, obj)
	}

	listValue, _ := types.ListValueFrom(ctx, WanNetworkCustomServiceIPObjectType, customServiceIps)
	return listValue
}

// parseWanNetworkConfigurationSDK parses the configuration object from SDK types
func parseWanNetworkConfigurationSDK(
	ctx context.Context,
	config cato_go_sdk.WanNetworkPolicy_Policy_WanNetwork_Policy_Rules_Rule_Configuration,
	logKey string,
) (types.Object, diag.Diagnostics) {
	diags := make(diag.Diagnostics, 0)

	// Parse primary transport
	primaryTransportObj := types.ObjectNull(WanNetworkTransportAttrTypes)
	if config.PrimaryTransport.TransportType != "" {
		primaryTransportObj, _ = types.ObjectValue(
			WanNetworkTransportAttrTypes,
			map[string]attr.Value{
				"transport_type":           types.StringValue(config.PrimaryTransport.TransportType.String()),
				"primary_interface_role":   types.StringValue(config.PrimaryTransport.PrimaryInterfaceRole.String()),
				"secondary_interface_role": types.StringValue(config.PrimaryTransport.SecondaryInterfaceRole.String()),
			},
		)
	}

	// Parse secondary transport
	secondaryTransportObj := types.ObjectNull(WanNetworkTransportAttrTypes)
	if config.SecondaryTransport.TransportType != "" {
		secondaryTransportObj, _ = types.ObjectValue(
			WanNetworkTransportAttrTypes,
			map[string]attr.Value{
				"transport_type":           types.StringValue(config.SecondaryTransport.TransportType.String()),
				"primary_interface_role":   types.StringValue(config.SecondaryTransport.PrimaryInterfaceRole.String()),
				"secondary_interface_role": types.StringValue(config.SecondaryTransport.SecondaryInterfaceRole.String()),
			},
		)
	}

	configObj, diagstmp := types.ObjectValue(
		WanNetworkConfigurationAttrTypes,
		map[string]attr.Value{
			"active_tcp_acceleration": types.BoolValue(config.ActiveTCPAcceleration),
			"packet_loss_mitigation":  types.BoolValue(config.PacketLossMitigation),
			"preserve_source_port":    types.BoolValue(config.PreserveSourcePort),
			"primary_transport":       primaryTransportObj,
			"secondary_transport":     secondaryTransportObj,
			"allocation_ip":           oldidname.ParseNameIDList(ctx, config.AllocationIP, logKey+".allocation_ip"),
			"pop_location":            oldidname.ParseNameIDList(ctx, config.PopLocation, logKey+".pop_location"),
			"backhauling_site":        oldidname.ParseNameIDList(ctx, config.BackhaulingSite, logKey+".backhauling_site"),
		},
	)
	diags = append(diags, diagstmp...)

	return configObj, diags
}

// parseWanNetworkExceptionCustomServiceListSDK parses exception custom service list from SDK types
func parseWanNetworkExceptionCustomServiceListSDK(
	ctx context.Context,
	items []*cato_go_sdk.WanNetworkPolicy_Policy_WanNetwork_Policy_Rules_Rule_Exceptions_Application_CustomService,
	_ string,
) types.List {
	if len(items) == 0 {
		return types.ListNull(customservice.CustomServiceObjectType)
	}

	var customServices []types.Object
	for _, item := range items {
		attrs := map[string]attr.Value{
			"port":       types.ListNull(types.StringType),
			"port_range": types.ObjectNull(portrange.PortRangeAttrTypes),
			"protocol":   types.StringNull(),
		}

		// Check if port_range is set first
		hasPortRange := item.PortRangeCustomService != nil && (item.PortRangeCustomService.From != "" || item.PortRangeCustomService.To != "")

		if hasPortRange {
			attrs["port_range"], _ = types.ObjectValue(
				portrange.PortRangeAttrTypes,
				map[string]attr.Value{
					"from": types.StringValue(fmt.Sprintf("%v", item.PortRangeCustomService.From)),
					"to":   types.StringValue(fmt.Sprintf("%v", item.PortRangeCustomService.To)),
				},
			)
			// When port_range is set, port must remain null
			attrs["port"] = types.ListNull(types.StringType)
		} else if len(item.Port) > 0 {
			// Only populate port if port_range is not set
			var ports []attr.Value
			for _, p := range item.Port {
				ports = append(ports, types.StringValue(fmt.Sprintf("%v", p)))
			}
			attrs["port"], _ = types.ListValue(types.StringType, ports)
		}

		if item.Protocol != "" {
			attrs["protocol"] = types.StringValue(string(item.Protocol))
		}

		obj, _ := types.ObjectValue(WanNetworkCustomServiceAttrTypes, attrs)
		customServices = append(customServices, obj)
	}

	listValue, _ := types.ListValueFrom(ctx, customservice.CustomServiceObjectType, customServices)
	return listValue
}

// parseWanNetworkExceptionCustomServiceIPListSDK parses exception custom service IP list from SDK types
func parseWanNetworkExceptionCustomServiceIPListSDK(
	ctx context.Context,
	items []*cato_go_sdk.WanNetworkPolicy_Policy_WanNetwork_Policy_Rules_Rule_Exceptions_Application_CustomServiceIP,
	_ string,
) types.List {
	if len(items) == 0 {
		return types.ListNull(WanNetworkCustomServiceIPObjectType)
	}

	var customServiceIps []types.Object
	for _, item := range items {
		attrs := map[string]attr.Value{
			"name":     types.StringValue(item.Name),
			"ip":       types.StringNull(),
			"ip_range": types.ObjectNull(iprange.RangeAttrTypes),
		}

		if item.IP != nil && *item.IP != "" {
			attrs["ip"] = types.StringValue(*item.IP)
		}

		if item.IPRangeCustomServiceIP != nil {
			attrs["ip_range"], _ = types.ObjectValue(
				iprange.RangeAttrTypes,
				map[string]attr.Value{
					"from": types.StringValue(item.IPRangeCustomServiceIP.From),
					"to":   types.StringValue(item.IPRangeCustomServiceIP.To),
				},
			)
		}

		obj, _ := types.ObjectValue(WanNetworkCustomServiceIPAttrTypes, attrs)
		customServiceIps = append(customServiceIps, obj)
	}

	listValue, _ := types.ListValueFrom(ctx, WanNetworkCustomServiceIPObjectType, customServiceIps)
	return listValue
}
