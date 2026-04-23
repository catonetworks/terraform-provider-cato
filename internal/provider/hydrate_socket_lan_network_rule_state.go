package provider

import (
	"context"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// SocketLanNetworkRuleStateOutput represents the state of a Socket LAN network rule
type SocketLanNetworkRuleStateOutput struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Index       types.Int64  `tfsdk:"index"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Direction   types.String `tfsdk:"direction"`
	Transport   types.String `tfsdk:"transport"`
	Site        types.Object `tfsdk:"site"`
	Source      types.Object `tfsdk:"source"`
	Destination types.Object `tfsdk:"destination"`
	Service     types.Object `tfsdk:"service"`
	Nat         types.Object `tfsdk:"nat"`
}

func hydrateSocketLanNetworkRuleState(ctx context.Context, plan SocketLanNetworkRule, apiRule *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule) SocketLanNetworkRuleStateOutput {
	result := SocketLanNetworkRuleStateOutput{
		ID:        types.StringValue(apiRule.ID),
		Name:      types.StringValue(apiRule.Name),
		Index:     types.Int64Value(int64(apiRule.Index)),
		Enabled:   types.BoolValue(apiRule.Enabled),
		Direction: types.StringValue(string(apiRule.Direction)),
		Transport: types.StringValue(string(apiRule.Transport)),
	}

	// Description
	if apiRule.Description != "" {
		result.Description = types.StringValue(apiRule.Description)
	} else {
		result.Description = types.StringNull()
	}

	// Site
	result.Site = hydrateSocketLanSiteState(ctx, apiRule)

	// Source
	result.Source = hydrateSocketLanSourceState(ctx, apiRule.Source)

	// Destination
	result.Destination = hydrateSocketLanDestinationState(ctx, apiRule.Destination)

	// Service - only create service object if there's actual content (following IFW pattern)
	if len(apiRule.Service.Simple) > 0 || len(apiRule.Service.Custom) > 0 {
		result.Service = hydrateSocketLanServiceState(ctx, apiRule.Service)
	} else {
		result.Service = types.ObjectNull(SocketLanServiceAttrTypes)
	}

	// NAT
	result.Nat = hydrateSocketLanNatState(ctx, apiRule.Nat)

	return result
}

func hydrateSocketLanSiteState(ctx context.Context, apiRule *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule) types.Object {
	siteAttrs := map[string]attr.Value{
		"site":  types.SetNull(NameIDObjectType),
		"group": types.SetNull(NameIDObjectType),
	}

	// Sites
	if len(apiRule.Site.Site) > 0 {
		var sites []attr.Value
		for _, s := range apiRule.Site.Site {
			siteObj, _ := types.ObjectValue(NameIDAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(s.ID),
				"name": types.StringValue(s.Name),
			})
			sites = append(sites, siteObj)
		}
		siteSet, _ := types.SetValue(NameIDObjectType, sites)
		siteAttrs["site"] = siteSet
	}

	// Groups
	if len(apiRule.Site.Group) > 0 {
		var groups []attr.Value
		for _, g := range apiRule.Site.Group {
			groupObj, _ := types.ObjectValue(NameIDAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(g.ID),
				"name": types.StringValue(g.Name),
			})
			groups = append(groups, groupObj)
		}
		groupSet, _ := types.SetValue(NameIDObjectType, groups)
		siteAttrs["group"] = groupSet
	}

	siteObj, _ := types.ObjectValue(SocketLanSiteAttrTypes, siteAttrs)
	return siteObj
}

func hydrateSocketLanSourceState(ctx context.Context, apiSource cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Source) types.Object {
	sourceAttrs := map[string]attr.Value{
		"vlan":                types.ListNull(types.Int64Type),
		"ip":                  types.ListNull(types.StringType),
		"subnet":              types.ListNull(types.StringType),
		"ip_range":            types.ListNull(FromToObjectType),
		"host":                types.SetNull(NameIDObjectType),
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

	sourceObj, _ := types.ObjectValue(SocketLanSourceAttrTypes, sourceAttrs)
	return sourceObj
}

func hydrateSocketLanDestinationState(ctx context.Context, apiDest cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Destination) types.Object {
	destAttrs := map[string]attr.Value{
		"vlan":                types.ListNull(types.Int64Type),
		"ip":                  types.ListNull(types.StringType),
		"subnet":              types.ListNull(types.StringType),
		"ip_range":            types.ListNull(FromToObjectType),
		"host":                types.SetNull(NameIDObjectType),
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

	destObj, _ := types.ObjectValue(SocketLanDestinationAttrTypes, destAttrs)
	return destObj
}

func hydrateSocketLanServiceState(ctx context.Context, apiService cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Service) types.Object {

	serviceAttrs := map[string]attr.Value{
		"simple": types.SetNull(SimpleServiceObjectType),
		"custom": types.ListNull(CustomServiceObjectType),
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

	serviceObj, _ := types.ObjectValue(SocketLanServiceAttrTypes, serviceAttrs)
	return serviceObj
}

func hydrateSocketLanNatState(ctx context.Context, apiNat cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Nat) basetypes.ObjectValue {

	natAttrs := map[string]attr.Value{
		"enabled":  types.BoolValue(apiNat.Enabled),
		"nat_type": types.StringNull(),
	}

	if apiNat.NatType != "" {
		natAttrs["nat_type"] = types.StringValue(string(apiNat.NatType))
	}

	natObj, _ := types.ObjectValue(SocketLanNatAttrTypes, natAttrs)
	return natObj
}
