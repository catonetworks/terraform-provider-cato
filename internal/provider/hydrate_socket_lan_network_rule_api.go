package provider

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/cato-go-sdk/scalars"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// SocketLanNetworkRuleApiInput holds both create and update inputs
type SocketLanNetworkRuleApiInput struct {
	create cato_models.SocketLanAddRuleInput
	update cato_models.SocketLanUpdateRuleInput
}

func hydrateSocketLanNetworkRuleApi(ctx context.Context, plan SocketLanNetworkRule) (SocketLanNetworkRuleApiInput, diag.Diagnostics) {
	var result SocketLanNetworkRuleApiInput
	var diags diag.Diagnostics

	// Initialize create input
	result.create = cato_models.SocketLanAddRuleInput{
		Rule: &cato_models.SocketLanAddRuleDataInput{},
	}
	result.update = cato_models.SocketLanUpdateRuleInput{
		Rule: &cato_models.SocketLanUpdateRuleDataInput{},
	}

	// Parse position
	if !plan.At.IsNull() {
		var positionInput PolicyRulePositionInput
		diagstmp := plan.At.As(ctx, &positionInput, basetypes.ObjectAsOptions{})
		diags.Append(diagstmp...)

		result.create.At = &cato_models.PolicyRulePositionInput{
			Position: (*cato_models.PolicyRulePositionEnum)(positionInput.Position.ValueStringPointer()),
			Ref:      positionInput.Ref.ValueStringPointer(),
		}
	}

	// Parse rule data
	if !plan.Rule.IsNull() {
		var ruleData SocketLanNetworkRuleData
		diagstmp := plan.Rule.As(ctx, &ruleData, basetypes.ObjectAsOptions{})
		diags.Append(diagstmp...)

		// Name
		result.create.Rule.Name = ruleData.Name.ValueString()
		result.update.Rule.Name = ruleData.Name.ValueStringPointer()

		// Description
		if !ruleData.Description.IsNull() {
			result.create.Rule.Description = ruleData.Description.ValueString()
			result.update.Rule.Description = ruleData.Description.ValueStringPointer()
		}

		// Enabled
		result.create.Rule.Enabled = ruleData.Enabled.ValueBool()
		result.update.Rule.Enabled = ruleData.Enabled.ValueBoolPointer()

		// Direction
		direction := cato_models.SocketLanDirection(ruleData.Direction.ValueString())
		result.create.Rule.Direction = direction
		result.update.Rule.Direction = &direction

		// Transport
		transport := cato_models.SocketLanTransportType(ruleData.Transport.ValueString())
		result.create.Rule.Transport = transport
		result.update.Rule.Transport = &transport

		// Site
		if !ruleData.Site.IsNull() {
			var siteData SocketLanSite
			diagstmp = ruleData.Site.As(ctx, &siteData, basetypes.ObjectAsOptions{})
			diags.Append(diagstmp...)

			result.create.Rule.Site = &cato_models.SocketLanSiteInput{
				Group: make([]*cato_models.GroupRefInput, 0),
			}
			result.update.Rule.Site = &cato_models.SocketLanSiteUpdateInput{
				Group: make([]*cato_models.GroupRefInput, 0),
			}

			// Site references
			if !siteData.Site.IsNull() && len(siteData.Site.Elements()) > 0 {
				var sites []NameIDRef
				diagstmp = siteData.Site.ElementsAs(ctx, &sites, false)
				diags.Append(diagstmp...)

				for _, site := range sites {
					ObjectRefOutput, err := utils.TransformObjectRefInput(site)
					if err != nil {
						tflog.Error(ctx, err.Error())
					}
					siteRef := cato_models.SiteRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					}
					result.create.Rule.Site.Site = append(result.create.Rule.Site.Site, &siteRef)
					result.update.Rule.Site.Site = append(result.update.Rule.Site.Site, &siteRef)
				}
			}

			// Group references
			if !siteData.Group.IsNull() && len(siteData.Group.Elements()) > 0 {
				var groups []NameIDRef
				diagstmp = siteData.Group.ElementsAs(ctx, &groups, false)
				diags.Append(diagstmp...)

				for _, group := range groups {
					ObjectRefOutput, err := utils.TransformObjectRefInput(group)
					if err != nil {
						tflog.Error(ctx, err.Error())
					}
					groupRef := cato_models.GroupRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					}
					result.create.Rule.Site.Group = append(result.create.Rule.Site.Group, &groupRef)
					result.update.Rule.Site.Group = append(result.update.Rule.Site.Group, &groupRef)
				}
			}
		}

		// Source
		if !ruleData.Source.IsNull() {
			result.create.Rule.Source = &cato_models.SocketLanSourceInput{}
			result.update.Rule.Source = &cato_models.SocketLanSourceUpdateInput{}
			hydrateSocketLanSourceApi(ctx, ruleData.Source, result.create.Rule.Source, result.update.Rule.Source, &diags)
		}

		// Destination
		if !ruleData.Destination.IsNull() {
			result.create.Rule.Destination = &cato_models.SocketLanDestinationInput{}
			result.update.Rule.Destination = &cato_models.SocketLanDestinationUpdateInput{}
			hydrateSocketLanDestinationApi(ctx, ruleData.Destination, result.create.Rule.Destination, result.update.Rule.Destination, &diags)
		}

		// Service - always initialize with empty slices
		result.create.Rule.Service = &cato_models.SocketLanServiceInput{
			Simple: make([]*cato_models.SimpleServiceInput, 0),
			Custom: make([]*cato_models.CustomServiceInput, 0),
		}
		result.update.Rule.Service = &cato_models.SocketLanServiceUpdateInput{
			Simple: make([]*cato_models.SimpleServiceInput, 0),
			Custom: make([]*cato_models.CustomServiceInput, 0),
		}

		if !ruleData.Service.IsNull() && !ruleData.Service.IsUnknown() {
			var serviceData SocketLanService
			diagstmp = ruleData.Service.As(ctx, &serviceData, basetypes.ObjectAsOptions{})
			diags.Append(diagstmp...)

			// Simple services
			if !serviceData.Simple.IsNull() && len(serviceData.Simple.Elements()) > 0 {
				var simpleServices []SocketLanSimpleService
				diagstmp = serviceData.Simple.ElementsAs(ctx, &simpleServices, false)
				diags.Append(diagstmp...)

				for _, svc := range simpleServices {
					result.create.Rule.Service.Simple = append(result.create.Rule.Service.Simple, &cato_models.SimpleServiceInput{
						Name: cato_models.SimpleServiceType(svc.Name.ValueString()),
					})
					result.update.Rule.Service.Simple = append(result.update.Rule.Service.Simple, &cato_models.SimpleServiceInput{
						Name: cato_models.SimpleServiceType(svc.Name.ValueString()),
					})
				}
			}

			// Custom services
			if !serviceData.Custom.IsNull() && len(serviceData.Custom.Elements()) > 0 {
				var customServices []Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service_Custom
				diagstmp = serviceData.Custom.ElementsAs(ctx, &customServices, false)
				diags.Append(diagstmp...)

				for _, svc := range customServices {
					customSvc := &cato_models.CustomServiceInput{
						Protocol: (cato_models.IPProtocol)(svc.Protocol.ValueString()),
					}

					// Ports
					if !svc.Port.IsNull() && len(svc.Port.Elements()) > 0 {
						var ports []string
						diagstmp = svc.Port.ElementsAs(ctx, &ports, false)
						diags.Append(diagstmp...)
						for _, p := range ports {
							customSvc.Port = append(customSvc.Port, scalars.Port(p))
						}
					}

					// Port range
					if !svc.PortRange.IsNull() {
						var portRange Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service_Custom_PortRange
						diagstmp = svc.PortRange.As(ctx, &portRange, basetypes.ObjectAsOptions{})
						diags.Append(diagstmp...)
						customSvc.PortRange = &cato_models.PortRangeInput{
							From: scalars.Port(portRange.From.ValueString()),
							To:   scalars.Port(portRange.To.ValueString()),
						}
					}

					result.create.Rule.Service.Custom = append(result.create.Rule.Service.Custom, customSvc)
					result.update.Rule.Service.Custom = append(result.update.Rule.Service.Custom, customSvc)
				}
			}
		}

		// NAT - always initialize with defaults (API requires nat object, natType defaults to DYNAMIC_PAT)
		result.create.Rule.Nat = &cato_models.SocketLanNatSettingsInput{
			Enabled: false,
			NatType: cato_models.SocketLanNatTypeDynamicPat,
		}
		defaultNatType := cato_models.SocketLanNatTypeDynamicPat
		result.update.Rule.Nat = &cato_models.SocketLanNatSettingsUpdateInput{
			Enabled: new(bool), // false by default
			NatType: &defaultNatType,
		}

		if !ruleData.Nat.IsNull() && !ruleData.Nat.IsUnknown() {
			var natData SocketLanNat
			diagstmp = ruleData.Nat.As(ctx, &natData, basetypes.ObjectAsOptions{})
			diags.Append(diagstmp...)

			result.create.Rule.Nat.Enabled = natData.Enabled.ValueBool()
			result.update.Rule.Nat.Enabled = natData.Enabled.ValueBoolPointer()

			if !natData.NatType.IsNull() && natData.NatType.ValueString() != "" {
				natType := cato_models.SocketLanNatType(natData.NatType.ValueString())
				result.create.Rule.Nat.NatType = natType
				result.update.Rule.Nat.NatType = &natType
			}
		}
	}

	return result, diags
}

func hydrateSocketLanSourceApi(ctx context.Context, sourceObj basetypes.ObjectValue, createInput *cato_models.SocketLanSourceInput, updateInput *cato_models.SocketLanSourceUpdateInput, diags *diag.Diagnostics) {
	var sourceData SocketLanSource
	diagstmp := sourceObj.As(ctx, &sourceData, basetypes.ObjectAsOptions{})
	diags.Append(diagstmp...)

	// Initialize all list fields to empty slices to avoid null serialization
	createInput.Vlan = make([]scalars.Vlan, 0)
	updateInput.Vlan = make([]scalars.Vlan, 0)
	createInput.IP = make([]string, 0)
	updateInput.IP = make([]string, 0)
	createInput.Subnet = make([]string, 0)
	updateInput.Subnet = make([]string, 0)
	createInput.IPRange = make([]*cato_models.IPAddressRangeInput, 0)
	updateInput.IPRange = make([]*cato_models.IPAddressRangeInput, 0)
	createInput.Host = make([]*cato_models.HostRefInput, 0)
	updateInput.Host = make([]*cato_models.HostRefInput, 0)
	createInput.Group = make([]*cato_models.GroupRefInput, 0)
	updateInput.Group = make([]*cato_models.GroupRefInput, 0)
	createInput.SystemGroup = make([]*cato_models.SystemGroupRefInput, 0)
	updateInput.SystemGroup = make([]*cato_models.SystemGroupRefInput, 0)
	createInput.NetworkInterface = make([]*cato_models.NetworkInterfaceRefInput, 0)
	updateInput.NetworkInterface = make([]*cato_models.NetworkInterfaceRefInput, 0)
	createInput.GlobalIPRange = make([]*cato_models.GlobalIPRangeRefInput, 0)
	updateInput.GlobalIPRange = make([]*cato_models.GlobalIPRangeRefInput, 0)
	createInput.FloatingSubnet = make([]*cato_models.FloatingSubnetRefInput, 0)
	updateInput.FloatingSubnet = make([]*cato_models.FloatingSubnetRefInput, 0)
	createInput.SiteNetworkSubnet = make([]*cato_models.SiteNetworkSubnetRefInput, 0)
	updateInput.SiteNetworkSubnet = make([]*cato_models.SiteNetworkSubnetRefInput, 0)

	// VLAN
	if !sourceData.Vlan.IsNull() && len(sourceData.Vlan.Elements()) > 0 {
		var vlans []int64
		diagstmp = sourceData.Vlan.ElementsAs(ctx, &vlans, false)
		diags.Append(diagstmp...)
		for _, v := range vlans {
			createInput.Vlan = append(createInput.Vlan, scalars.Vlan(v))
			updateInput.Vlan = append(updateInput.Vlan, scalars.Vlan(v))
		}
	}

	// IP
	if !sourceData.IP.IsNull() && len(sourceData.IP.Elements()) > 0 {
		var ips []string
		diagstmp = sourceData.IP.ElementsAs(ctx, &ips, false)
		diags.Append(diagstmp...)
		createInput.IP = ips
		updateInput.IP = ips
	}

	// Subnet
	if !sourceData.Subnet.IsNull() && len(sourceData.Subnet.Elements()) > 0 {
		var subnets []string
		diagstmp = sourceData.Subnet.ElementsAs(ctx, &subnets, false)
		diags.Append(diagstmp...)
		createInput.Subnet = subnets
		updateInput.Subnet = subnets
	}

	// IP Range
	if !sourceData.IPRange.IsNull() && len(sourceData.IPRange.Elements()) > 0 {
		var ipRanges []Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_IPRange
		diagstmp = sourceData.IPRange.ElementsAs(ctx, &ipRanges, false)
		diags.Append(diagstmp...)
		for _, r := range ipRanges {
			createInput.IPRange = append(createInput.IPRange, &cato_models.IPAddressRangeInput{
				From: r.From.ValueString(),
				To:   r.To.ValueString(),
			})
			updateInput.IPRange = append(updateInput.IPRange, &cato_models.IPAddressRangeInput{
				From: r.From.ValueString(),
				To:   r.To.ValueString(),
			})
		}
	}

	// Host
	if !sourceData.Host.IsNull() && len(sourceData.Host.Elements()) > 0 {
		var hosts []NameIDRef
		diagstmp = sourceData.Host.ElementsAs(ctx, &hosts, false)
		diags.Append(diagstmp...)
		for _, h := range hosts {
			ObjectRefOutput, err := utils.TransformObjectRefInput(h)
			if err != nil {
				tflog.Error(ctx, err.Error())
			}
			hostRef := cato_models.HostRefInput{
				By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
				Input: ObjectRefOutput.Input,
			}
			createInput.Host = append(createInput.Host, &hostRef)
			updateInput.Host = append(updateInput.Host, &hostRef)
		}
	}

	// Group
	if !sourceData.Group.IsNull() && len(sourceData.Group.Elements()) > 0 {
		var groups []NameIDRef
		diagstmp = sourceData.Group.ElementsAs(ctx, &groups, false)
		diags.Append(diagstmp...)
		for _, g := range groups {
			ObjectRefOutput, err := utils.TransformObjectRefInput(g)
			if err != nil {
				tflog.Error(ctx, err.Error())
			}
			groupRef := cato_models.GroupRefInput{
				By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
				Input: ObjectRefOutput.Input,
			}
			createInput.Group = append(createInput.Group, &groupRef)
			updateInput.Group = append(updateInput.Group, &groupRef)
		}
	}

	// System Group
	if !sourceData.SystemGroup.IsNull() && len(sourceData.SystemGroup.Elements()) > 0 {
		var systemGroups []NameIDRef
		diagstmp = sourceData.SystemGroup.ElementsAs(ctx, &systemGroups, false)
		diags.Append(diagstmp...)
		for _, sg := range systemGroups {
			ObjectRefOutput, err := utils.TransformObjectRefInput(sg)
			if err != nil {
				tflog.Error(ctx, err.Error())
			}
			sgRef := cato_models.SystemGroupRefInput{
				By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
				Input: ObjectRefOutput.Input,
			}
			createInput.SystemGroup = append(createInput.SystemGroup, &sgRef)
			updateInput.SystemGroup = append(updateInput.SystemGroup, &sgRef)
		}
	}

	// Network Interface
	if !sourceData.NetworkInterface.IsNull() && len(sourceData.NetworkInterface.Elements()) > 0 {
		var networkInterfaces []NameIDRef
		diagstmp = sourceData.NetworkInterface.ElementsAs(ctx, &networkInterfaces, false)
		diags.Append(diagstmp...)
		for _, ni := range networkInterfaces {
			ObjectRefOutput, err := utils.TransformObjectRefInput(ni)
			if err != nil {
				tflog.Error(ctx, err.Error())
			}
			niRef := cato_models.NetworkInterfaceRefInput{
				By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
				Input: ObjectRefOutput.Input,
			}
			createInput.NetworkInterface = append(createInput.NetworkInterface, &niRef)
			updateInput.NetworkInterface = append(updateInput.NetworkInterface, &niRef)
		}
	}

	// Global IP Range
	if !sourceData.GlobalIPRange.IsNull() && len(sourceData.GlobalIPRange.Elements()) > 0 {
		var globalIPRanges []NameIDRef
		diagstmp = sourceData.GlobalIPRange.ElementsAs(ctx, &globalIPRanges, false)
		diags.Append(diagstmp...)
		for _, gir := range globalIPRanges {
			ObjectRefOutput, err := utils.TransformObjectRefInput(gir)
			if err != nil {
				tflog.Error(ctx, err.Error())
			}
			girRef := cato_models.GlobalIPRangeRefInput{
				By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
				Input: ObjectRefOutput.Input,
			}
			createInput.GlobalIPRange = append(createInput.GlobalIPRange, &girRef)
			updateInput.GlobalIPRange = append(updateInput.GlobalIPRange, &girRef)
		}
	}

	// Floating Subnet
	if !sourceData.FloatingSubnet.IsNull() && len(sourceData.FloatingSubnet.Elements()) > 0 {
		var floatingSubnets []NameIDRef
		diagstmp = sourceData.FloatingSubnet.ElementsAs(ctx, &floatingSubnets, false)
		diags.Append(diagstmp...)
		for _, fs := range floatingSubnets {
			ObjectRefOutput, err := utils.TransformObjectRefInput(fs)
			if err != nil {
				tflog.Error(ctx, err.Error())
			}
			fsRef := cato_models.FloatingSubnetRefInput{
				By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
				Input: ObjectRefOutput.Input,
			}
			createInput.FloatingSubnet = append(createInput.FloatingSubnet, &fsRef)
			updateInput.FloatingSubnet = append(updateInput.FloatingSubnet, &fsRef)
		}
	}

	// Site Network Subnet
	if !sourceData.SiteNetworkSubnet.IsNull() && len(sourceData.SiteNetworkSubnet.Elements()) > 0 {
		var siteNetworkSubnets []NameIDRef
		diagstmp = sourceData.SiteNetworkSubnet.ElementsAs(ctx, &siteNetworkSubnets, false)
		diags.Append(diagstmp...)
		for _, sns := range siteNetworkSubnets {
			ObjectRefOutput, err := utils.TransformObjectRefInput(sns)
			if err != nil {
				tflog.Error(ctx, err.Error())
			}
			snsRef := cato_models.SiteNetworkSubnetRefInput{
				By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
				Input: ObjectRefOutput.Input,
			}
			createInput.SiteNetworkSubnet = append(createInput.SiteNetworkSubnet, &snsRef)
			updateInput.SiteNetworkSubnet = append(updateInput.SiteNetworkSubnet, &snsRef)
		}
	}
}

func hydrateSocketLanDestinationApi(ctx context.Context, destObj basetypes.ObjectValue, createInput *cato_models.SocketLanDestinationInput, updateInput *cato_models.SocketLanDestinationUpdateInput, diags *diag.Diagnostics) {
	var destData SocketLanDestination
	diagstmp := destObj.As(ctx, &destData, basetypes.ObjectAsOptions{})
	diags.Append(diagstmp...)

	// Initialize all list fields to empty slices to avoid null serialization
	createInput.Vlan = make([]scalars.Vlan, 0)
	updateInput.Vlan = make([]scalars.Vlan, 0)
	createInput.IP = make([]string, 0)
	updateInput.IP = make([]string, 0)
	createInput.Subnet = make([]string, 0)
	updateInput.Subnet = make([]string, 0)
	createInput.IPRange = make([]*cato_models.IPAddressRangeInput, 0)
	updateInput.IPRange = make([]*cato_models.IPAddressRangeInput, 0)
	createInput.Host = make([]*cato_models.HostRefInput, 0)
	updateInput.Host = make([]*cato_models.HostRefInput, 0)
	createInput.Group = make([]*cato_models.GroupRefInput, 0)
	updateInput.Group = make([]*cato_models.GroupRefInput, 0)
	createInput.SystemGroup = make([]*cato_models.SystemGroupRefInput, 0)
	updateInput.SystemGroup = make([]*cato_models.SystemGroupRefInput, 0)
	createInput.NetworkInterface = make([]*cato_models.NetworkInterfaceRefInput, 0)
	updateInput.NetworkInterface = make([]*cato_models.NetworkInterfaceRefInput, 0)
	createInput.GlobalIPRange = make([]*cato_models.GlobalIPRangeRefInput, 0)
	updateInput.GlobalIPRange = make([]*cato_models.GlobalIPRangeRefInput, 0)
	createInput.FloatingSubnet = make([]*cato_models.FloatingSubnetRefInput, 0)
	updateInput.FloatingSubnet = make([]*cato_models.FloatingSubnetRefInput, 0)
	createInput.SiteNetworkSubnet = make([]*cato_models.SiteNetworkSubnetRefInput, 0)
	updateInput.SiteNetworkSubnet = make([]*cato_models.SiteNetworkSubnetRefInput, 0)

	// VLAN
	if !destData.Vlan.IsNull() && len(destData.Vlan.Elements()) > 0 {
		var vlans []int64
		diagstmp = destData.Vlan.ElementsAs(ctx, &vlans, false)
		diags.Append(diagstmp...)
		for _, v := range vlans {
			createInput.Vlan = append(createInput.Vlan, scalars.Vlan(v))
			updateInput.Vlan = append(updateInput.Vlan, scalars.Vlan(v))
		}
	}

	// IP
	if !destData.IP.IsNull() && len(destData.IP.Elements()) > 0 {
		var ips []string
		diagstmp = destData.IP.ElementsAs(ctx, &ips, false)
		diags.Append(diagstmp...)
		createInput.IP = ips
		updateInput.IP = ips
	}

	// Subnet
	if !destData.Subnet.IsNull() && len(destData.Subnet.Elements()) > 0 {
		var subnets []string
		diagstmp = destData.Subnet.ElementsAs(ctx, &subnets, false)
		diags.Append(diagstmp...)
		createInput.Subnet = subnets
		updateInput.Subnet = subnets
	}

	// IP Range
	if !destData.IPRange.IsNull() && len(destData.IPRange.Elements()) > 0 {
		var ipRanges []Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_IPRange
		diagstmp = destData.IPRange.ElementsAs(ctx, &ipRanges, false)
		diags.Append(diagstmp...)
		for _, r := range ipRanges {
			createInput.IPRange = append(createInput.IPRange, &cato_models.IPAddressRangeInput{
				From: r.From.ValueString(),
				To:   r.To.ValueString(),
			})
			updateInput.IPRange = append(updateInput.IPRange, &cato_models.IPAddressRangeInput{
				From: r.From.ValueString(),
				To:   r.To.ValueString(),
			})
		}
	}

	// Host
	if !destData.Host.IsNull() && len(destData.Host.Elements()) > 0 {
		var hosts []NameIDRef
		diagstmp = destData.Host.ElementsAs(ctx, &hosts, false)
		diags.Append(diagstmp...)
		for _, h := range hosts {
			ObjectRefOutput, err := utils.TransformObjectRefInput(h)
			if err != nil {
				tflog.Error(ctx, err.Error())
			}
			hostRef := cato_models.HostRefInput{
				By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
				Input: ObjectRefOutput.Input,
			}
			createInput.Host = append(createInput.Host, &hostRef)
			updateInput.Host = append(updateInput.Host, &hostRef)
		}
	}

	// Group
	if !destData.Group.IsNull() && len(destData.Group.Elements()) > 0 {
		var groups []NameIDRef
		diagstmp = destData.Group.ElementsAs(ctx, &groups, false)
		diags.Append(diagstmp...)
		for _, g := range groups {
			ObjectRefOutput, err := utils.TransformObjectRefInput(g)
			if err != nil {
				tflog.Error(ctx, err.Error())
			}
			groupRef := cato_models.GroupRefInput{
				By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
				Input: ObjectRefOutput.Input,
			}
			createInput.Group = append(createInput.Group, &groupRef)
			updateInput.Group = append(updateInput.Group, &groupRef)
		}
	}

	// System Group
	if !destData.SystemGroup.IsNull() && len(destData.SystemGroup.Elements()) > 0 {
		var systemGroups []NameIDRef
		diagstmp = destData.SystemGroup.ElementsAs(ctx, &systemGroups, false)
		diags.Append(diagstmp...)
		for _, sg := range systemGroups {
			ObjectRefOutput, err := utils.TransformObjectRefInput(sg)
			if err != nil {
				tflog.Error(ctx, err.Error())
			}
			sgRef := cato_models.SystemGroupRefInput{
				By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
				Input: ObjectRefOutput.Input,
			}
			createInput.SystemGroup = append(createInput.SystemGroup, &sgRef)
			updateInput.SystemGroup = append(updateInput.SystemGroup, &sgRef)
		}
	}

	// Network Interface
	if !destData.NetworkInterface.IsNull() && len(destData.NetworkInterface.Elements()) > 0 {
		var networkInterfaces []NameIDRef
		diagstmp = destData.NetworkInterface.ElementsAs(ctx, &networkInterfaces, false)
		diags.Append(diagstmp...)
		for _, ni := range networkInterfaces {
			ObjectRefOutput, err := utils.TransformObjectRefInput(ni)
			if err != nil {
				tflog.Error(ctx, err.Error())
			}
			niRef := cato_models.NetworkInterfaceRefInput{
				By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
				Input: ObjectRefOutput.Input,
			}
			createInput.NetworkInterface = append(createInput.NetworkInterface, &niRef)
			updateInput.NetworkInterface = append(updateInput.NetworkInterface, &niRef)
		}
	}

	// Global IP Range
	if !destData.GlobalIPRange.IsNull() && len(destData.GlobalIPRange.Elements()) > 0 {
		var globalIPRanges []NameIDRef
		diagstmp = destData.GlobalIPRange.ElementsAs(ctx, &globalIPRanges, false)
		diags.Append(diagstmp...)
		for _, gir := range globalIPRanges {
			ObjectRefOutput, err := utils.TransformObjectRefInput(gir)
			if err != nil {
				tflog.Error(ctx, err.Error())
			}
			girRef := cato_models.GlobalIPRangeRefInput{
				By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
				Input: ObjectRefOutput.Input,
			}
			createInput.GlobalIPRange = append(createInput.GlobalIPRange, &girRef)
			updateInput.GlobalIPRange = append(updateInput.GlobalIPRange, &girRef)
		}
	}

	// Floating Subnet
	if !destData.FloatingSubnet.IsNull() && len(destData.FloatingSubnet.Elements()) > 0 {
		var floatingSubnets []NameIDRef
		diagstmp = destData.FloatingSubnet.ElementsAs(ctx, &floatingSubnets, false)
		diags.Append(diagstmp...)
		for _, fs := range floatingSubnets {
			ObjectRefOutput, err := utils.TransformObjectRefInput(fs)
			if err != nil {
				tflog.Error(ctx, err.Error())
			}
			fsRef := cato_models.FloatingSubnetRefInput{
				By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
				Input: ObjectRefOutput.Input,
			}
			createInput.FloatingSubnet = append(createInput.FloatingSubnet, &fsRef)
			updateInput.FloatingSubnet = append(updateInput.FloatingSubnet, &fsRef)
		}
	}

	// Site Network Subnet
	if !destData.SiteNetworkSubnet.IsNull() && len(destData.SiteNetworkSubnet.Elements()) > 0 {
		var siteNetworkSubnets []NameIDRef
		diagstmp = destData.SiteNetworkSubnet.ElementsAs(ctx, &siteNetworkSubnets, false)
		diags.Append(diagstmp...)
		for _, sns := range siteNetworkSubnets {
			ObjectRefOutput, err := utils.TransformObjectRefInput(sns)
			if err != nil {
				tflog.Error(ctx, err.Error())
			}
			snsRef := cato_models.SiteNetworkSubnetRefInput{
				By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
				Input: ObjectRefOutput.Input,
			}
			createInput.SiteNetworkSubnet = append(createInput.SiteNetworkSubnet, &snsRef)
			updateInput.SiteNetworkSubnet = append(updateInput.SiteNetworkSubnet, &snsRef)
		}
	}
}
