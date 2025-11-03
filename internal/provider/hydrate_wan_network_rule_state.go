package provider

import (
	"context"
	"fmt"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func hydrateWanNetworkRuleState(ctx context.Context, state WanNetworkRule, currentRule *cato_go_sdk.WanNetworkPolicy_Policy_WanNetwork_Policy_Rules_Rule) (Policy_Policy_WanNetwork_Policy_Rules_Rule, diag.Diagnostics) {

	ruleInput := Policy_Policy_WanNetwork_Policy_Rules_Rule{}
	diags := make(diag.Diagnostics, 0)

	// Handle case where state might be incomplete (e.g., during import)
	if !state.Rule.IsNull() && !state.Rule.IsUnknown() {
		diagstmp := state.Rule.As(ctx, &ruleInput, basetypes.ObjectAsOptions{})
		if diagstmp.HasError() {
			tflog.Debug(ctx, "Failed to convert state to struct in hydrateWanNetworkRuleState, using empty struct for hydration")
			ruleInput = Policy_Policy_WanNetwork_Policy_Rules_Rule{}
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
	
	// Set index from API response when state index is null/unknown, otherwise preserve state index
	if ruleInput.Index.IsNull() || ruleInput.Index.IsUnknown() {
		ruleInput.Index = types.Int64Value(currentRule.Index)
	}
	
	ruleInput.Enabled = types.BoolValue(currentRule.Enabled)
	ruleInput.RuleType = types.StringValue(currentRule.RuleType.String())
	ruleInput.RouteType = types.StringValue(currentRule.RouteType.String())

	//////////// Rule -> Source ///////////////
	curRuleSourceObj, diagstmp := types.ObjectValue(
		WanNetworkSourceAttrTypes,
		map[string]attr.Value{
			"ip":                  parseList(ctx, types.StringType, currentRule.Source.IP, "rule.source.ip"),
			"host":                parseNameIDList(ctx, currentRule.Source.Host, "rule.source.host"),
			"site":                parseNameIDList(ctx, currentRule.Source.Site, "rule.source.site"),
			"subnet":              parseList(ctx, types.StringType, currentRule.Source.Subnet, "rule.source.subnet"),
			"ip_range":            parseFromToList(ctx, currentRule.Source.IPRangeWanNetworkRuleSource, "rule.source.ip_range"),
			"global_ip_range":     parseNameIDList(ctx, currentRule.Source.GlobalIPRange, "rule.source.global_ip_range"),
			"network_interface":   parseNameIDList(ctx, currentRule.Source.NetworkInterface, "rule.source.network_interface"),
			"site_network_subnet": parseNameIDList(ctx, currentRule.Source.SiteNetworkSubnet, "rule.source.site_network_subnet"),
			"floating_subnet":     parseNameIDList(ctx, currentRule.Source.FloatingSubnet, "rule.source.floating_subnet"),
			"user":                parseNameIDList(ctx, currentRule.Source.User, "rule.source.user"),
			"users_group":         parseNameIDList(ctx, currentRule.Source.UsersGroup, "rule.source.users_group"),
			"group":               parseNameIDList(ctx, currentRule.Source.Group, "rule.source.group"),
			"system_group":        parseNameIDList(ctx, currentRule.Source.SystemGroup, "rule.source.system_group"),
		},
	)
	diags = append(diags, diagstmp...)
	
	// Ensure source.network_interface is never unknown after apply
	srcAttrs := curRuleSourceObj.Attributes()
	if v, ok := srcAttrs["network_interface"]; ok {
		if setVal, ok2 := v.(types.Set); ok2 {
			if setVal.IsUnknown() {
				srcAttrs["network_interface"] = types.SetNull(NameIDObjectType)
				var diagsTmp diag.Diagnostics
				curRuleSourceObj, diagsTmp = types.ObjectValue(curRuleSourceObj.Type(ctx).(types.ObjectType).AttrTypes, srcAttrs)
				diags = append(diags, diagsTmp...)
			}
		}
	}

	ruleInput.Source = curRuleSourceObj
	////////////// end rule.source ///////////////

	//////////// Rule -> Destination ///////////////
	curRuleDestinationObj, diagstmp := types.ObjectValue(
		WanNetworkDestAttrTypes,
		map[string]attr.Value{
			"ip":                  parseList(ctx, types.StringType, currentRule.Destination.IP, "rule.destination.ip"),
			"host":                parseNameIDList(ctx, currentRule.Destination.Host, "rule.destination.host"),
			"site":                parseNameIDList(ctx, currentRule.Destination.Site, "rule.destination.site"),
			"subnet":              parseList(ctx, types.StringType, currentRule.Destination.Subnet, "rule.destination.subnet"),
			"ip_range":            parseFromToList(ctx, currentRule.Destination.IPRangeWanNetworkRuleDestination, "rule.destination.ip_range"),
			"global_ip_range":     parseNameIDList(ctx, currentRule.Destination.GlobalIPRange, "rule.destination.global_ip_range"),
			"network_interface":   parseNameIDList(ctx, currentRule.Destination.NetworkInterface, "rule.destination.network_interface"),
			"site_network_subnet": parseNameIDList(ctx, currentRule.Destination.SiteNetworkSubnet, "rule.destination.site_network_subnet"),
			"floating_subnet":     parseNameIDList(ctx, currentRule.Destination.FloatingSubnet, "rule.destination.floating_subnet"),
			"user":                parseNameIDList(ctx, currentRule.Destination.User, "rule.destination.user"),
			"users_group":         parseNameIDList(ctx, currentRule.Destination.UsersGroup, "rule.destination.users_group"),
			"group":               parseNameIDList(ctx, currentRule.Destination.Group, "rule.destination.group"),
			"system_group":        parseNameIDList(ctx, currentRule.Destination.SystemGroup, "rule.destination.system_group"),
		},
	)
	diags = append(diags, diagstmp...)
	ruleInput.Destination = curRuleDestinationObj
	////////////// end Rule -> Destination ///////////////

	//////////// Rule -> Application ///////////////
	curRuleApplicationObj, diagstmp := types.ObjectValue(
		WanNetworkApplicationAttrTypes,
		map[string]attr.Value{
			"application":       parseNameIDList(ctx, currentRule.Application.Application, "rule.application.application"),
			"custom_app":        parseNameIDList(ctx, currentRule.Application.CustomApp, "rule.application.custom_app"),
			"app_category":      parseNameIDList(ctx, currentRule.Application.AppCategory, "rule.application.app_category"),
			"custom_category":   parseNameIDList(ctx, currentRule.Application.CustomCategory, "rule.application.custom_category"),
			"domain":            parseList(ctx, types.StringType, currentRule.Application.Domain, "rule.application.domain"),
			"fqdn":              parseList(ctx, types.StringType, currentRule.Application.Fqdn, "rule.application.fqdn"),
			"service":           parseNameIDList(ctx, currentRule.Application.Service, "rule.application.service"),
			"custom_service":     parseWanNetworkCustomServiceListSDK(ctx, currentRule.Application.CustomService, "rule.application.custom_service"),
			"custom_service_ip":  parseWanNetworkCustomServiceIpListSDK(ctx, currentRule.Application.CustomServiceIP, "rule.application.custom_service_ip"),
		},
	)
	ruleInput.Application = curRuleApplicationObj
	diags = append(diags, diagstmp...)
	////////////// end Rule -> Application ///////////////

	////////////// start Rule -> Configuration ///////////////
	curRuleConfigurationObj, diagstmp := parseWanNetworkConfigurationSDK(ctx, currentRule.Configuration, "rule.configuration")
	diags = append(diags, diagstmp...)
	ruleInput.Configuration = curRuleConfigurationObj
	////////////// end Rule -> Configuration ///////////////

	////////////// start Rule -> BandwidthPriority ///////////////
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
	////////////// end Rule -> BandwidthPriority ///////////////

	//////////////// start Rule -> Exceptions ///////////////
	exceptions := []attr.Value{}

	tflog.Warn(ctx, "hydrateWanNetworkRuleState() currentRule.Exceptions - "+fmt.Sprintf("%v", currentRule.Exceptions)+" len="+fmt.Sprintf("%v", len(currentRule.Exceptions)))
	if len(currentRule.Exceptions) > 0 {
		for _, ruleException := range currentRule.Exceptions {
			// Rule -> Exceptions -> Source
			curExceptionSourceObj, diagstmp := types.ObjectValue(
				WanNetworkSourceAttrTypes,
				map[string]attr.Value{
					"ip":                  parseList(ctx, types.StringType, ruleException.Source.IP, "rule.exception.source.ip"),
					"host":                parseNameIDList(ctx, ruleException.Source.Host, "rule.exception.source.host"),
					"site":                parseNameIDList(ctx, ruleException.Source.Site, "rule.exception.source.site"),
					"subnet":              parseList(ctx, types.StringType, ruleException.Source.Subnet, "rule.exception.source.subnet"),
					"ip_range":            parseFromToList(ctx, ruleException.Source.IPRange, "rule.exception.source.ip_range"),
					"global_ip_range":     parseNameIDList(ctx, ruleException.Source.GlobalIPRange, "rule.exception.source.global_ip_range"),
					"network_interface":   parseNameIDList(ctx, ruleException.Source.NetworkInterface, "rule.exception.source.network_interface"),
					"site_network_subnet": parseNameIDList(ctx, ruleException.Source.SiteNetworkSubnet, "rule.exception.source.site_network_subnet"),
					"floating_subnet":     parseNameIDList(ctx, ruleException.Source.FloatingSubnet, "rule.exception.source.floating_subnet"),
					"user":                parseNameIDList(ctx, ruleException.Source.User, "rule.exception.source.user"),
					"users_group":         parseNameIDList(ctx, ruleException.Source.UsersGroup, "rule.exception.source.users_group"),
					"group":               parseNameIDList(ctx, ruleException.Source.Group, "rule.exception.source.group"),
					"system_group":        parseNameIDList(ctx, ruleException.Source.SystemGroup, "rule.exception.source.system_group"),
				},
			)
			diags = append(diags, diagstmp...)

			// Rule -> Exceptions -> Destination
			curExceptionDestObj, diagstmp := types.ObjectValue(
				WanNetworkDestAttrTypes,
				map[string]attr.Value{
					"ip":                  parseList(ctx, types.StringType, ruleException.Destination.IP, "rule.exception.destination.ip"),
					"host":                parseNameIDList(ctx, ruleException.Destination.Host, "rule.exception.destination.host"),
					"site":                parseNameIDList(ctx, ruleException.Destination.Site, "rule.exception.destination.site"),
					"subnet":              parseList(ctx, types.StringType, ruleException.Destination.Subnet, "rule.exception.destination.subnet"),
					"ip_range":            parseFromToList(ctx, ruleException.Destination.IPRange, "rule.exception.destination.ip_range"),
					"global_ip_range":     parseNameIDList(ctx, ruleException.Destination.GlobalIPRange, "rule.exception.destination.global_ip_range"),
					"network_interface":   parseNameIDList(ctx, ruleException.Destination.NetworkInterface, "rule.exception.destination.network_interface"),
					"site_network_subnet": parseNameIDList(ctx, ruleException.Destination.SiteNetworkSubnet, "rule.exception.destination.site_network_subnet"),
					"floating_subnet":     parseNameIDList(ctx, ruleException.Destination.FloatingSubnet, "rule.exception.destination.floating_subnet"),
					"user":                parseNameIDList(ctx, ruleException.Destination.User, "rule.exception.destination.user"),
					"users_group":         parseNameIDList(ctx, ruleException.Destination.UsersGroup, "rule.exception.destination.users_group"),
					"group":               parseNameIDList(ctx, ruleException.Destination.Group, "rule.exception.destination.group"),
					"system_group":        parseNameIDList(ctx, ruleException.Destination.SystemGroup, "rule.exception.destination.system_group"),
				},
			)
			diags = append(diags, diagstmp...)

			// //////////// Rule -> Exceptions -> Application ///////////////
			curExceptionApplicationObj, diagstmp := types.ObjectValue(
				WanNetworkExceptionApplicationAttrTypes,
				map[string]attr.Value{
					"application":       parseNameIDList(ctx, ruleException.Application.Application, "rule.exception.application.application"),
					"custom_app":        parseNameIDList(ctx, ruleException.Application.CustomApp, "rule.exception.application.custom_app"),
					"app_category":      parseNameIDList(ctx, ruleException.Application.AppCategory, "rule.exception.application.app_category"),
					"custom_category":   parseNameIDList(ctx, ruleException.Application.CustomCategory, "rule.exception.application.custom_category"),
					"domain":            parseList(ctx, types.StringType, ruleException.Application.Domain, "rule.exception.application.domain"),
					"fqdn":              parseList(ctx, types.StringType, ruleException.Application.Fqdn, "rule.exception.application.fqdn"),
					"service":           parseNameIDList(ctx, ruleException.Application.Service, "rule.exception.application.service"),
					"custom_service":    parseWanNetworkExceptionCustomServiceListSDK(ctx, ruleException.Application.CustomService, "rule.exception.application.custom_service"),
					"custom_service_ip": parseWanNetworkExceptionCustomServiceIpListSDK(ctx, ruleException.Application.CustomServiceIP, "rule.exception.application.custom_service_ip"),
				},
			)
			diags = append(diags, diagstmp...)
			////////////// end Rule -> Exceptions -> Application ///////////////

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
	////////////// end Rule -> Exceptions ///////////////

	return ruleInput, diags
}

// parseWanNetworkCustomServiceList parses custom service list for WAN Network rules
func parseWanNetworkCustomServiceList(ctx context.Context, items []*cato_models.CustomService, logKey string) types.List {
	if len(items) == 0 {
		return types.ListNull(WanNetworkCustomServiceObjectType)
	}

	var customServices []types.Object
	for _, item := range items {
		customServiceObj := parseWanNetworkCustomService(ctx, item, logKey)
		customServices = append(customServices, customServiceObj)
	}

	listValue, _ := types.ListValueFrom(ctx, WanNetworkCustomServiceObjectType, customServices)
	return listValue
}

// parseWanNetworkCustomService parses a single custom service for WAN Network rules
func parseWanNetworkCustomService(ctx context.Context, item *cato_models.CustomService, logKey string) types.Object {
	attrs := map[string]attr.Value{
		"port":       types.ListNull(types.StringType),
		"port_range": types.ObjectNull(FromToAttrTypes),
		"protocol":   types.StringNull(),
	}

	if len(item.Port) > 0 {
		var ports []attr.Value
		for _, p := range item.Port {
			ports = append(ports, types.StringValue(fmt.Sprintf("%v", p)))
		}
		attrs["port"], _ = types.ListValue(types.StringType, ports)
	}

	if item.PortRange != nil {
		attrs["port_range"], _ = types.ObjectValue(
			FromToAttrTypes,
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
	return obj
}

// parseWanNetworkCustomServiceIpList parses custom service IP list for WAN Network rules
func parseWanNetworkCustomServiceIpList(ctx context.Context, items []*cato_models.CustomServiceIP, logKey string) types.List {
	if len(items) == 0 {
		return types.ListNull(WanNetworkCustomServiceIpObjectType)
	}

	var customServiceIps []types.Object
	for _, item := range items {
		attrs := map[string]attr.Value{
			"name":     types.StringValue(item.Name),
			"ip":       types.StringNull(),
			"ip_range": types.ObjectNull(FromToAttrTypes),
		}

		if item.IP != nil && *item.IP != "" {
			attrs["ip"] = types.StringValue(*item.IP)
		}

		if item.IPRange != nil {
			attrs["ip_range"], _ = types.ObjectValue(
				FromToAttrTypes,
				map[string]attr.Value{
					"from": types.StringValue(item.IPRange.From),
					"to":   types.StringValue(item.IPRange.To),
				},
			)
		}

		obj, _ := types.ObjectValue(WanNetworkCustomServiceIpAttrTypes, attrs)
		customServiceIps = append(customServiceIps, obj)
	}

	listValue, _ := types.ListValueFrom(ctx, WanNetworkCustomServiceIpObjectType, customServiceIps)
	return listValue
}

// parseWanNetworkConfiguration parses the configuration object for WAN Network rules
func parseWanNetworkConfiguration(ctx context.Context, config *cato_models.WanNetworkRuleConfiguration, logKey string) (types.Object, diag.Diagnostics) {
	diags := make(diag.Diagnostics, 0)

	if config == nil {
		return types.ObjectNull(WanNetworkConfigurationAttrTypes), diags
	}

	// Parse primary transport
	primaryTransportObj := types.ObjectNull(WanNetworkTransportAttrTypes)
	if config.PrimaryTransport != nil {
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
	if config.SecondaryTransport != nil {
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
			"allocation_ip":           parseNameIDList(ctx, config.AllocationIP, logKey+".allocation_ip"),
			"pop_location":            parseNameIDList(ctx, config.PopLocation, logKey+".pop_location"),
			"backhauling_site":        parseNameIDList(ctx, config.BackhaulingSite, logKey+".backhauling_site"),
		},
	)
	diags = append(diags, diagstmp...)

	return configObj, diags
}

// SDK-specific parse functions for WAN Network query responses

// parseWanNetworkCustomServiceListSDK parses custom service list from SDK types
func parseWanNetworkCustomServiceListSDK(ctx context.Context, items []*cato_go_sdk.WanNetworkPolicy_Policy_WanNetwork_Policy_Rules_Rule_Application_CustomService, logKey string) types.List {
	if len(items) == 0 {
		return types.ListNull(WanNetworkCustomServiceObjectType)
	}

	var customServices []types.Object
	for _, item := range items {
		attrs := map[string]attr.Value{
			"port":       types.ListNull(types.StringType),
			"port_range": types.ObjectNull(FromToAttrTypes),
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
			FromToAttrTypes,
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

	listValue, _ := types.ListValueFrom(ctx, WanNetworkCustomServiceObjectType, customServices)
	return listValue
}

// parseWanNetworkCustomServiceIpListSDK parses custom service IP list from SDK types
func parseWanNetworkCustomServiceIpListSDK(ctx context.Context, items []*cato_go_sdk.WanNetworkPolicy_Policy_WanNetwork_Policy_Rules_Rule_Application_CustomServiceIP, logKey string) types.List {
	if len(items) == 0 {
		return types.ListNull(WanNetworkCustomServiceIpObjectType)
	}

	var customServiceIps []types.Object
	for _, item := range items {
		attrs := map[string]attr.Value{
			"name":     types.StringValue(item.Name),
			"ip":       types.StringNull(),
			"ip_range": types.ObjectNull(FromToAttrTypes),
		}

		if item.IP != nil && *item.IP != "" {
			attrs["ip"] = types.StringValue(*item.IP)
		}

		if item.IPRange != nil {
			attrs["ip_range"], _ = types.ObjectValue(
				FromToAttrTypes,
				map[string]attr.Value{
					"from": types.StringValue(item.IPRange.From),
					"to":   types.StringValue(item.IPRange.To),
				},
			)
		}

		obj, _ := types.ObjectValue(WanNetworkCustomServiceIpAttrTypes, attrs)
		customServiceIps = append(customServiceIps, obj)
	}

	listValue, _ := types.ListValueFrom(ctx, WanNetworkCustomServiceIpObjectType, customServiceIps)
	return listValue
}

// parseWanNetworkConfigurationSDK parses the configuration object from SDK types
func parseWanNetworkConfigurationSDK(ctx context.Context, config cato_go_sdk.WanNetworkPolicy_Policy_WanNetwork_Policy_Rules_Rule_Configuration, logKey string) (types.Object, diag.Diagnostics) {
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
			"allocation_ip":           parseNameIDList(ctx, config.AllocationIP, logKey+".allocation_ip"),
			"pop_location":            parseNameIDList(ctx, config.PopLocation, logKey+".pop_location"),
			"backhauling_site":        parseNameIDList(ctx, config.BackhaulingSite, logKey+".backhauling_site"),
		},
	)
	diags = append(diags, diagstmp...)

	return configObj, diags
}

// parseWanNetworkExceptionCustomServiceListSDK parses exception custom service list from SDK types
func parseWanNetworkExceptionCustomServiceListSDK(ctx context.Context, items []*cato_go_sdk.WanNetworkPolicy_Policy_WanNetwork_Policy_Rules_Rule_Exceptions_Application_CustomService, logKey string) types.List {
	if len(items) == 0 {
		return types.ListNull(WanNetworkCustomServiceObjectType)
	}

	var customServices []types.Object
	for _, item := range items {
		attrs := map[string]attr.Value{
			"port":       types.ListNull(types.StringType),
			"port_range": types.ObjectNull(FromToAttrTypes),
			"protocol":   types.StringNull(),
		}

		if len(item.Port) > 0 {
			var ports []attr.Value
			for _, p := range item.Port {
				ports = append(ports, types.StringValue(fmt.Sprintf("%v", p)))
			}
			attrs["port"], _ = types.ListValue(types.StringType, ports)
		}

		if item.PortRangeCustomService != nil && (item.PortRangeCustomService.From != "" || item.PortRangeCustomService.To != "") {
			attrs["port_range"], _ = types.ObjectValue(
				FromToAttrTypes,
				map[string]attr.Value{
					"from": types.StringValue(fmt.Sprintf("%v", item.PortRangeCustomService.From)),
					"to":   types.StringValue(fmt.Sprintf("%v", item.PortRangeCustomService.To)),
				},
			)
		}

		if item.Protocol != "" {
			attrs["protocol"] = types.StringValue(string(item.Protocol))
		}

		obj, _ := types.ObjectValue(WanNetworkCustomServiceAttrTypes, attrs)
		customServices = append(customServices, obj)
	}

	listValue, _ := types.ListValueFrom(ctx, WanNetworkCustomServiceObjectType, customServices)
	return listValue
}

// parseWanNetworkExceptionCustomServiceIpListSDK parses exception custom service IP list from SDK types
func parseWanNetworkExceptionCustomServiceIpListSDK(ctx context.Context, items []*cato_go_sdk.WanNetworkPolicy_Policy_WanNetwork_Policy_Rules_Rule_Exceptions_Application_CustomServiceIP, logKey string) types.List {
	if len(items) == 0 {
		return types.ListNull(WanNetworkCustomServiceIpObjectType)
	}

	var customServiceIps []types.Object
	for _, item := range items {
		attrs := map[string]attr.Value{
			"name":     types.StringValue(item.Name),
			"ip":       types.StringNull(),
			"ip_range": types.ObjectNull(FromToAttrTypes),
		}

		if item.IP != nil && *item.IP != "" {
			attrs["ip"] = types.StringValue(*item.IP)
		}

		if item.IPRangeCustomServiceIP != nil {
			attrs["ip_range"], _ = types.ObjectValue(
				FromToAttrTypes,
				map[string]attr.Value{
					"from": types.StringValue(item.IPRangeCustomServiceIP.From),
					"to":   types.StringValue(item.IPRangeCustomServiceIP.To),
				},
			)
		}

		obj, _ := types.ObjectValue(WanNetworkCustomServiceIpAttrTypes, attrs)
		customServiceIps = append(customServiceIps, obj)
	}

	listValue, _ := types.ListValueFrom(ctx, WanNetworkCustomServiceIpObjectType, customServiceIps)
	return listValue
}
