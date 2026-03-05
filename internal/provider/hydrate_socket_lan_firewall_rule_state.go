package provider

import (
	"context"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// SocketLanFirewallRuleStateOutput represents the state of a Socket LAN firewall rule
type SocketLanFirewallRuleStateOutput struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Index       types.Int64  `tfsdk:"index"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Direction   types.String `tfsdk:"direction"`
	Action      types.String `tfsdk:"action"`
	Source      types.Object `tfsdk:"source"`
	Destination types.Object `tfsdk:"destination"`
	Application types.Object `tfsdk:"application"`
	Service     types.Object `tfsdk:"service"`
	Tracking    types.Object `tfsdk:"tracking"`
}

func hydrateSocketLanFirewallRuleState(ctx context.Context, plan SocketLanFirewallRule, apiRule *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Firewall_Rule) SocketLanFirewallRuleStateOutput {
	result := SocketLanFirewallRuleStateOutput{
		ID:        types.StringValue(apiRule.ID),
		Name:      types.StringValue(apiRule.Name),
		Index:     types.Int64Value(int64(apiRule.Index)),
		Enabled:   types.BoolValue(apiRule.Enabled),
		Direction: types.StringValue(string(apiRule.Direction)),
		Action:    types.StringValue(string(apiRule.Action)),
	}

	// Description
	if apiRule.Description != "" {
		result.Description = types.StringValue(apiRule.Description)
	} else {
		result.Description = types.StringNull()
	}

	// Source
	result.Source = hydrateSocketLanFirewallSourceState(ctx, apiRule.Source)

	// Destination
	result.Destination = hydrateSocketLanFirewallDestinationState(ctx, apiRule.Destination)

	// Application - check if plan had application block defined
	var planRuleData SocketLanFirewallRuleData
	plan.Rule.As(ctx, &planRuleData, basetypes.ObjectAsOptions{})
	planHasApplication := !planRuleData.Application.IsNull()
	result.Application = hydrateSocketLanFirewallApplicationState(ctx, apiRule.Application, planHasApplication)

	// Service - only create service object if there's actual content (following IFW pattern)
	if len(apiRule.Service.Simple) > 0 || len(apiRule.Service.Standard) > 0 || len(apiRule.Service.Custom) > 0 {
		result.Service = hydrateSocketLanFirewallServiceState(ctx, apiRule.Service)
	} else {
		result.Service = types.ObjectNull(SocketLanFirewallServiceAttrTypes)
	}

	// Tracking
	result.Tracking = hydrateSocketLanFirewallTrackingState(ctx, apiRule.Tracking)

	return result
}

func hydrateSocketLanFirewallSourceState(ctx context.Context, apiSource cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Firewall_Rule_Source) types.Object {
	sourceAttrs := map[string]attr.Value{
		"vlan":                types.ListNull(types.Int64Type),
		"mac":                 types.ListNull(types.StringType),
		"ip":                  types.ListNull(types.StringType),
		"subnet":              types.ListNull(types.StringType),
		"ip_range":            types.ListNull(FromToObjectType),
		"host":                types.SetNull(NameIDObjectType),
		"site":                types.SetNull(NameIDObjectType),
		"group":               types.SetNull(NameIDObjectType),
		"system_group":        types.SetNull(NameIDObjectType),
		"network_interface":   types.SetNull(NameIDObjectType),
		"global_ip_range":     types.SetNull(NameIDObjectType),
		"floating_subnet":     types.SetNull(NameIDObjectType),
		"site_network_subnet": types.SetNull(NameIDObjectType),
	}

	// VLAN
	if len(apiSource.Vlan) > 0 {
		var vlans []attr.Value
		for _, v := range apiSource.Vlan {
			vlans = append(vlans, types.Int64Value(int64(v)))
		}
		vlanList, _ := types.ListValue(types.Int64Type, vlans)
		sourceAttrs["vlan"] = vlanList
	}

	// Mac
	if len(apiSource.Mac) > 0 {
		var macs []attr.Value
		for _, m := range apiSource.Mac {
			macs = append(macs, types.StringValue(m))
		}
		macList, _ := types.ListValue(types.StringType, macs)
		sourceAttrs["mac"] = macList
	}

	// IP
	if len(apiSource.IP) > 0 {
		var ips []attr.Value
		for _, ip := range apiSource.IP {
			ips = append(ips, types.StringValue(ip))
		}
		ipList, _ := types.ListValue(types.StringType, ips)
		sourceAttrs["ip"] = ipList
	}

	// Subnet
	if len(apiSource.Subnet) > 0 {
		var subnets []attr.Value
		for _, s := range apiSource.Subnet {
			subnets = append(subnets, types.StringValue(s))
		}
		subnetList, _ := types.ListValue(types.StringType, subnets)
		sourceAttrs["subnet"] = subnetList
	}

	// IP Range
	if len(apiSource.IPRange) > 0 {
		var ipRanges []attr.Value
		for _, r := range apiSource.IPRange {
			rangeObj, _ := types.ObjectValue(FromToAttrTypes, map[string]attr.Value{
				"from": types.StringValue(r.From),
				"to":   types.StringValue(r.To),
			})
			ipRanges = append(ipRanges, rangeObj)
		}
		ipRangeList, _ := types.ListValue(FromToObjectType, ipRanges)
		sourceAttrs["ip_range"] = ipRangeList
	}

	// Host
	if len(apiSource.Host) > 0 {
		var hosts []attr.Value
		for _, h := range apiSource.Host {
			hostObj, _ := types.ObjectValue(NameIDAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(h.ID),
				"name": types.StringValue(h.Name),
			})
			hosts = append(hosts, hostObj)
		}
		hostSet, _ := types.SetValue(NameIDObjectType, hosts)
		sourceAttrs["host"] = hostSet
	}

	// Site
	if len(apiSource.Site) > 0 {
		var sites []attr.Value
		for _, s := range apiSource.Site {
			siteObj, _ := types.ObjectValue(NameIDAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(s.ID),
				"name": types.StringValue(s.Name),
			})
			sites = append(sites, siteObj)
		}
		siteSet, _ := types.SetValue(NameIDObjectType, sites)
		sourceAttrs["site"] = siteSet
	}

	// Group
	if len(apiSource.Group) > 0 {
		var groups []attr.Value
		for _, g := range apiSource.Group {
			groupObj, _ := types.ObjectValue(NameIDAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(g.ID),
				"name": types.StringValue(g.Name),
			})
			groups = append(groups, groupObj)
		}
		groupSet, _ := types.SetValue(NameIDObjectType, groups)
		sourceAttrs["group"] = groupSet
	}

	// System Group
	if len(apiSource.SystemGroup) > 0 {
		var systemGroups []attr.Value
		for _, sg := range apiSource.SystemGroup {
			sgObj, _ := types.ObjectValue(NameIDAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(sg.ID),
				"name": types.StringValue(sg.Name),
			})
			systemGroups = append(systemGroups, sgObj)
		}
		sgSet, _ := types.SetValue(NameIDObjectType, systemGroups)
		sourceAttrs["system_group"] = sgSet
	}

	// Network Interface
	if len(apiSource.NetworkInterface) > 0 {
		var networkInterfaces []attr.Value
		for _, ni := range apiSource.NetworkInterface {
			niObj, _ := types.ObjectValue(NameIDAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(ni.ID),
				"name": types.StringValue(ni.Name),
			})
			networkInterfaces = append(networkInterfaces, niObj)
		}
		niSet, _ := types.SetValue(NameIDObjectType, networkInterfaces)
		sourceAttrs["network_interface"] = niSet
	}

	// Global IP Range
	if len(apiSource.GlobalIPRange) > 0 {
		var globalIPRanges []attr.Value
		for _, gir := range apiSource.GlobalIPRange {
			girObj, _ := types.ObjectValue(NameIDAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(gir.ID),
				"name": types.StringValue(gir.Name),
			})
			globalIPRanges = append(globalIPRanges, girObj)
		}
		girSet, _ := types.SetValue(NameIDObjectType, globalIPRanges)
		sourceAttrs["global_ip_range"] = girSet
	}

	// Floating Subnet
	if len(apiSource.FloatingSubnet) > 0 {
		var floatingSubnets []attr.Value
		for _, fs := range apiSource.FloatingSubnet {
			fsObj, _ := types.ObjectValue(NameIDAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(fs.ID),
				"name": types.StringValue(fs.Name),
			})
			floatingSubnets = append(floatingSubnets, fsObj)
		}
		fsSet, _ := types.SetValue(NameIDObjectType, floatingSubnets)
		sourceAttrs["floating_subnet"] = fsSet
	}

	// Site Network Subnet
	if len(apiSource.SiteNetworkSubnet) > 0 {
		var siteNetworkSubnets []attr.Value
		for _, sns := range apiSource.SiteNetworkSubnet {
			snsObj, _ := types.ObjectValue(NameIDAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(sns.ID),
				"name": types.StringValue(sns.Name),
			})
			siteNetworkSubnets = append(siteNetworkSubnets, snsObj)
		}
		snsSet, _ := types.SetValue(NameIDObjectType, siteNetworkSubnets)
		sourceAttrs["site_network_subnet"] = snsSet
	}

	sourceObj, _ := types.ObjectValue(SocketLanFirewallSourceAttrTypes, sourceAttrs)
	return sourceObj
}

func hydrateSocketLanFirewallDestinationState(ctx context.Context, apiDest cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Firewall_Rule_Destination) types.Object {
	destAttrs := map[string]attr.Value{
		"vlan":                types.ListNull(types.Int64Type),
		"ip":                  types.ListNull(types.StringType),
		"subnet":              types.ListNull(types.StringType),
		"ip_range":            types.ListNull(FromToObjectType),
		"host":                types.SetNull(NameIDObjectType),
		"site":                types.SetNull(NameIDObjectType),
		"group":               types.SetNull(NameIDObjectType),
		"system_group":        types.SetNull(NameIDObjectType),
		"network_interface":   types.SetNull(NameIDObjectType),
		"global_ip_range":     types.SetNull(NameIDObjectType),
		"floating_subnet":     types.SetNull(NameIDObjectType),
		"site_network_subnet": types.SetNull(NameIDObjectType),
	}

	// VLAN
	if len(apiDest.Vlan) > 0 {
		var vlans []attr.Value
		for _, v := range apiDest.Vlan {
			vlans = append(vlans, types.Int64Value(int64(v)))
		}
		vlanList, _ := types.ListValue(types.Int64Type, vlans)
		destAttrs["vlan"] = vlanList
	}

	// IP
	if len(apiDest.IP) > 0 {
		var ips []attr.Value
		for _, ip := range apiDest.IP {
			ips = append(ips, types.StringValue(ip))
		}
		ipList, _ := types.ListValue(types.StringType, ips)
		destAttrs["ip"] = ipList
	}

	// Subnet
	if len(apiDest.Subnet) > 0 {
		var subnets []attr.Value
		for _, s := range apiDest.Subnet {
			subnets = append(subnets, types.StringValue(s))
		}
		subnetList, _ := types.ListValue(types.StringType, subnets)
		destAttrs["subnet"] = subnetList
	}

	// IP Range
	if len(apiDest.IPRange) > 0 {
		var ipRanges []attr.Value
		for _, r := range apiDest.IPRange {
			rangeObj, _ := types.ObjectValue(FromToAttrTypes, map[string]attr.Value{
				"from": types.StringValue(r.From),
				"to":   types.StringValue(r.To),
			})
			ipRanges = append(ipRanges, rangeObj)
		}
		ipRangeList, _ := types.ListValue(FromToObjectType, ipRanges)
		destAttrs["ip_range"] = ipRangeList
	}

	// Host
	if len(apiDest.Host) > 0 {
		var hosts []attr.Value
		for _, h := range apiDest.Host {
			hostObj, _ := types.ObjectValue(NameIDAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(h.ID),
				"name": types.StringValue(h.Name),
			})
			hosts = append(hosts, hostObj)
		}
		hostSet, _ := types.SetValue(NameIDObjectType, hosts)
		destAttrs["host"] = hostSet
	}

	// Site
	if len(apiDest.Site) > 0 {
		var sites []attr.Value
		for _, s := range apiDest.Site {
			siteObj, _ := types.ObjectValue(NameIDAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(s.ID),
				"name": types.StringValue(s.Name),
			})
			sites = append(sites, siteObj)
		}
		siteSet, _ := types.SetValue(NameIDObjectType, sites)
		destAttrs["site"] = siteSet
	}

	// Group
	if len(apiDest.Group) > 0 {
		var groups []attr.Value
		for _, g := range apiDest.Group {
			groupObj, _ := types.ObjectValue(NameIDAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(g.ID),
				"name": types.StringValue(g.Name),
			})
			groups = append(groups, groupObj)
		}
		groupSet, _ := types.SetValue(NameIDObjectType, groups)
		destAttrs["group"] = groupSet
	}

	// System Group
	if len(apiDest.SystemGroup) > 0 {
		var systemGroups []attr.Value
		for _, sg := range apiDest.SystemGroup {
			sgObj, _ := types.ObjectValue(NameIDAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(sg.ID),
				"name": types.StringValue(sg.Name),
			})
			systemGroups = append(systemGroups, sgObj)
		}
		sgSet, _ := types.SetValue(NameIDObjectType, systemGroups)
		destAttrs["system_group"] = sgSet
	}

	// Network Interface
	if len(apiDest.NetworkInterface) > 0 {
		var networkInterfaces []attr.Value
		for _, ni := range apiDest.NetworkInterface {
			niObj, _ := types.ObjectValue(NameIDAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(ni.ID),
				"name": types.StringValue(ni.Name),
			})
			networkInterfaces = append(networkInterfaces, niObj)
		}
		niSet, _ := types.SetValue(NameIDObjectType, networkInterfaces)
		destAttrs["network_interface"] = niSet
	}

	// Global IP Range
	if len(apiDest.GlobalIPRange) > 0 {
		var globalIPRanges []attr.Value
		for _, gir := range apiDest.GlobalIPRange {
			girObj, _ := types.ObjectValue(NameIDAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(gir.ID),
				"name": types.StringValue(gir.Name),
			})
			globalIPRanges = append(globalIPRanges, girObj)
		}
		girSet, _ := types.SetValue(NameIDObjectType, globalIPRanges)
		destAttrs["global_ip_range"] = girSet
	}

	// Floating Subnet
	if len(apiDest.FloatingSubnet) > 0 {
		var floatingSubnets []attr.Value
		for _, fs := range apiDest.FloatingSubnet {
			fsObj, _ := types.ObjectValue(NameIDAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(fs.ID),
				"name": types.StringValue(fs.Name),
			})
			floatingSubnets = append(floatingSubnets, fsObj)
		}
		fsSet, _ := types.SetValue(NameIDObjectType, floatingSubnets)
		destAttrs["floating_subnet"] = fsSet
	}

	// Site Network Subnet
	if len(apiDest.SiteNetworkSubnet) > 0 {
		var siteNetworkSubnets []attr.Value
		for _, sns := range apiDest.SiteNetworkSubnet {
			snsObj, _ := types.ObjectValue(NameIDAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(sns.ID),
				"name": types.StringValue(sns.Name),
			})
			siteNetworkSubnets = append(siteNetworkSubnets, snsObj)
		}
		snsSet, _ := types.SetValue(NameIDObjectType, siteNetworkSubnets)
		destAttrs["site_network_subnet"] = snsSet
	}

	destObj, _ := types.ObjectValue(SocketLanFirewallDestinationAttrTypes, destAttrs)
	return destObj
}

func hydrateSocketLanFirewallApplicationState(ctx context.Context, apiApp cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Firewall_Rule_Application, planHasApplication bool) types.Object {

	appAttrs := map[string]attr.Value{
		"application":     types.SetNull(NameIDObjectType),
		"custom_app":      types.SetNull(NameIDObjectType),
		"domain":          types.ListNull(types.StringType),
		"fqdn":            types.ListNull(types.StringType),
		"ip":              types.ListNull(types.StringType),
		"subnet":          types.ListNull(types.StringType),
		"ip_range":        types.ListNull(FromToObjectType),
		"global_ip_range": types.SetNull(NameIDObjectType),
	}

	// Application
	if len(apiApp.Application) > 0 {
		var apps []attr.Value
		for _, app := range apiApp.Application {
			appObj, _ := types.ObjectValue(NameIDAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(app.ID),
				"name": types.StringValue(app.Name),
			})
			apps = append(apps, appObj)
		}
		appSet, _ := types.SetValue(NameIDObjectType, apps)
		appAttrs["application"] = appSet
	}

	// CustomApp
	if len(apiApp.CustomApp) > 0 {
		var customApps []attr.Value
		for _, ca := range apiApp.CustomApp {
			caObj, _ := types.ObjectValue(NameIDAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(ca.ID),
				"name": types.StringValue(ca.Name),
			})
			customApps = append(customApps, caObj)
		}
		caSet, _ := types.SetValue(NameIDObjectType, customApps)
		appAttrs["custom_app"] = caSet
	}

	// Domain
	if len(apiApp.Domain) > 0 {
		var domains []attr.Value
		for _, d := range apiApp.Domain {
			domains = append(domains, types.StringValue(d))
		}
		domainList, _ := types.ListValue(types.StringType, domains)
		appAttrs["domain"] = domainList
	}

	// FQDN
	if len(apiApp.Fqdn) > 0 {
		var fqdns []attr.Value
		for _, f := range apiApp.Fqdn {
			fqdns = append(fqdns, types.StringValue(f))
		}
		fqdnList, _ := types.ListValue(types.StringType, fqdns)
		appAttrs["fqdn"] = fqdnList
	}

	// IP
	if len(apiApp.IP) > 0 {
		var ips []attr.Value
		for _, ip := range apiApp.IP {
			ips = append(ips, types.StringValue(ip))
		}
		ipList, _ := types.ListValue(types.StringType, ips)
		appAttrs["ip"] = ipList
	}

	// Subnet
	if len(apiApp.Subnet) > 0 {
		var subnets []attr.Value
		for _, s := range apiApp.Subnet {
			subnets = append(subnets, types.StringValue(s))
		}
		subnetList, _ := types.ListValue(types.StringType, subnets)
		appAttrs["subnet"] = subnetList
	}

	// IP Range
	if len(apiApp.IPRange) > 0 {
		var ipRanges []attr.Value
		for _, r := range apiApp.IPRange {
			rangeObj, _ := types.ObjectValue(FromToAttrTypes, map[string]attr.Value{
				"from": types.StringValue(r.From),
				"to":   types.StringValue(r.To),
			})
			ipRanges = append(ipRanges, rangeObj)
		}
		ipRangeList, _ := types.ListValue(FromToObjectType, ipRanges)
		appAttrs["ip_range"] = ipRangeList
	}

	// Global IP Range
	if len(apiApp.GlobalIPRange) > 0 {
		var globalIPRanges []attr.Value
		for _, gir := range apiApp.GlobalIPRange {
			girObj, _ := types.ObjectValue(NameIDAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(gir.ID),
				"name": types.StringValue(gir.Name),
			})
			globalIPRanges = append(globalIPRanges, girObj)
		}
		girSet, _ := types.SetValue(NameIDObjectType, globalIPRanges)
		appAttrs["global_ip_range"] = girSet
	}

	// Check if there's any actual content from the API
	hasContent := len(apiApp.Application) > 0 || len(apiApp.CustomApp) > 0 || len(apiApp.Domain) > 0 ||
		len(apiApp.Fqdn) > 0 || len(apiApp.IP) > 0 || len(apiApp.Subnet) > 0 ||
		len(apiApp.IPRange) > 0 || len(apiApp.GlobalIPRange) > 0

	// If plan didn't have application block and API has no content, return null
	// If plan had application block (even empty), always return object structure
	if !planHasApplication && !hasContent {
		return types.ObjectNull(SocketLanFirewallApplicationAttrTypes)
	}

	appObj, _ := types.ObjectValue(SocketLanFirewallApplicationAttrTypes, appAttrs)
	return appObj
}

func hydrateSocketLanFirewallServiceState(ctx context.Context, apiService cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Firewall_Rule_Service) types.Object {

	serviceAttrs := map[string]attr.Value{
		"simple":   types.SetNull(SimpleServiceObjectType),
		"standard": types.SetNull(NameIDObjectType),
		"custom":   types.ListNull(CustomServiceObjectType),
	}

	// Simple services
	if len(apiService.Simple) > 0 {
		var simpleServices []attr.Value
		for _, s := range apiService.Simple {
			svcObj, _ := types.ObjectValue(SimpleServiceAttrTypes, map[string]attr.Value{
				"name": types.StringValue(string(s.Name)),
			})
			simpleServices = append(simpleServices, svcObj)
		}
		simpleSet, _ := types.SetValue(SimpleServiceObjectType, simpleServices)
		serviceAttrs["simple"] = simpleSet
	}

	// Standard services
	if len(apiService.Standard) > 0 {
		var standardServices []attr.Value
		for _, s := range apiService.Standard {
			svcObj, _ := types.ObjectValue(NameIDAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(s.ID),
				"name": types.StringValue(s.Name),
			})
			standardServices = append(standardServices, svcObj)
		}
		standardSet, _ := types.SetValue(NameIDObjectType, standardServices)
		serviceAttrs["standard"] = standardSet
	}

	// Custom services
	if len(apiService.Custom) > 0 {
		var customServices []attr.Value
		for _, c := range apiService.Custom {
			customAttrs := map[string]attr.Value{
				"protocol":   types.StringValue(string(c.Protocol)),
				"port":       types.ListNull(types.StringType),
				"port_range": types.ObjectNull(FromToAttrTypes),
			}

			// Ports
			if len(c.Port) > 0 {
				var ports []attr.Value
				for _, p := range c.Port {
					ports = append(ports, types.StringValue(string(p)))
				}
				portList, _ := types.ListValue(types.StringType, ports)
				customAttrs["port"] = portList
			}

			// Port range
			if c.PortRange != nil {
				portRangeObj, _ := types.ObjectValue(FromToAttrTypes, map[string]attr.Value{
					"from": types.StringValue(string(c.PortRange.From)),
					"to":   types.StringValue(string(c.PortRange.To)),
				})
				customAttrs["port_range"] = portRangeObj
			}

			customObj, _ := types.ObjectValue(CustomServiceAttrTypes, customAttrs)
			customServices = append(customServices, customObj)
		}
		customList, _ := types.ListValue(CustomServiceObjectType, customServices)
		serviceAttrs["custom"] = customList
	}

	serviceObj, _ := types.ObjectValue(SocketLanFirewallServiceAttrTypes, serviceAttrs)
	return serviceObj
}

func hydrateSocketLanFirewallTrackingState(ctx context.Context, apiTracking cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Firewall_Rule_Tracking) types.Object {

	trackingAttrs := map[string]attr.Value{
		"event": types.ObjectNull(TrackingEventAttrTypes),
		"alert": types.ObjectNull(TrackingAlertAttrTypes),
	}

	// Event - always create event object (API always returns tracking)
	eventObj, _ := types.ObjectValue(TrackingEventAttrTypes, map[string]attr.Value{
		"enabled": types.BoolValue(apiTracking.Event.Enabled),
	})
	trackingAttrs["event"] = eventObj

	// Alert - always create alert object with all fields (even when enabled = false)
	alertAttrs := map[string]attr.Value{
		"enabled":            types.BoolValue(apiTracking.Alert.Enabled),
		"frequency":          types.StringNull(),
		"subscription_group": types.SetNull(NameIDObjectType),
		"webhook":            types.SetNull(NameIDObjectType),
		"mailing_list":       types.SetNull(NameIDObjectType),
	}

	if apiTracking.Alert.Frequency != "" {
		alertAttrs["frequency"] = types.StringValue(string(apiTracking.Alert.Frequency))
	}

	// Subscription groups
	if len(apiTracking.Alert.SubscriptionGroup) > 0 {
		var subGroups []attr.Value
		for _, sg := range apiTracking.Alert.SubscriptionGroup {
			sgObj, _ := types.ObjectValue(NameIDAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(sg.ID),
				"name": types.StringValue(sg.Name),
			})
			subGroups = append(subGroups, sgObj)
		}
		sgSet, _ := types.SetValue(NameIDObjectType, subGroups)
		alertAttrs["subscription_group"] = sgSet
	}

	// Webhooks
	if len(apiTracking.Alert.Webhook) > 0 {
		var webhooks []attr.Value
		for _, wh := range apiTracking.Alert.Webhook {
			whObj, _ := types.ObjectValue(NameIDAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(wh.ID),
				"name": types.StringValue(wh.Name),
			})
			webhooks = append(webhooks, whObj)
		}
		whSet, _ := types.SetValue(NameIDObjectType, webhooks)
		alertAttrs["webhook"] = whSet
	}

	// Mailing lists
	if len(apiTracking.Alert.MailingList) > 0 {
		var mailingLists []attr.Value
		for _, ml := range apiTracking.Alert.MailingList {
			mlObj, _ := types.ObjectValue(NameIDAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(ml.ID),
				"name": types.StringValue(ml.Name),
			})
			mailingLists = append(mailingLists, mlObj)
		}
		mlSet, _ := types.SetValue(NameIDObjectType, mailingLists)
		alertAttrs["mailing_list"] = mlSet
	}

	alertObj, _ := types.ObjectValue(TrackingAlertAttrTypes, alertAttrs)
	trackingAttrs["alert"] = alertObj

	trackingObj, _ := types.ObjectValue(TrackingAttrTypes, trackingAttrs)
	return trackingObj
}
