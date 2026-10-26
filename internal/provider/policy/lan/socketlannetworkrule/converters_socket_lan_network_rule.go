package socketlannetworkrule

import (
	"context"
	"fmt"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	scalars "github.com/catonetworks/cato-go-sdk/scalars"
	attr "github.com/hashicorp/terraform-plugin-framework/attr"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	customservice "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/customservice"
	idname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	iprange "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/iprange"
	policyrule "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
	portrange "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/portrange"
	simpleservice "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/simpleservice"
	socketlandestination "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/socketlandestination"
	socketlansite "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/socketlansite"
	socketlansource "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/socketlansource"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

//nolint:gocyclo,funlen
func hydrateSocketLanNetworkRuleAPI(
	ctx context.Context,
	plan SocketLanNetworkRule,
) (SocketLanNetworkRuleAPIInput, diag.Diagnostics) {
	var result SocketLanNetworkRuleAPIInput
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
		var positionInput policyrule.Position
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
			var siteData socketlansite.SocketLanSite
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
				var sites []idname.Ref
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
				var groups []idname.Ref
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
			hydrateSocketLanSourceAPI(ctx, ruleData.Source, result.create.Rule.Source, result.update.Rule.Source, &diags)
		}

		// Destination
		if !ruleData.Destination.IsNull() {
			result.create.Rule.Destination = &cato_models.SocketLanDestinationInput{}
			result.update.Rule.Destination = &cato_models.SocketLanDestinationUpdateInput{}
			hydrateSocketLanDestinationAPI(ctx, ruleData.Destination, result.create.Rule.Destination, result.update.Rule.Destination, &diags)
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
				var simpleServices []simpleservice.SimpleService
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
				var customServices []customservice.CustomService
				diagstmp = serviceData.Custom.ElementsAs(ctx, &customServices, false)
				diags.Append(diagstmp...)

				for _, svc := range customServices {
					customSvc := &cato_models.CustomServiceInput{
						Protocol: cato_models.IPProtocol(svc.Protocol.ValueString()),
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
						var portRange portrange.PortRange
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

//nolint:gocyclo,funlen
func hydrateSocketLanSourceAPI(
	ctx context.Context,
	sourceObj basetypes.ObjectValue,
	createInput *cato_models.SocketLanSourceInput,
	updateInput *cato_models.SocketLanSourceUpdateInput,
	diags *diag.Diagnostics,
) {
	var sourceData socketlansource.SocketLanSource
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
		var ipRanges []iprange.Range
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
		var hosts []idname.Ref
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
		var groups []idname.Ref
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
		var systemGroups []idname.Ref
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
		var networkInterfaces []idname.Ref
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
		var globalIPRanges []idname.Ref
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
		var floatingSubnets []idname.Ref
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
		var siteNetworkSubnets []idname.Ref
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

//nolint:gocyclo,funlen
func hydrateSocketLanDestinationAPI(
	ctx context.Context,
	destObj basetypes.ObjectValue,
	createInput *cato_models.SocketLanDestinationInput,
	updateInput *cato_models.SocketLanDestinationUpdateInput,
	diags *diag.Diagnostics,
) {
	var destData socketlandestination.SocketLanDestination
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
		var ipRanges []iprange.Range
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
		var hosts []idname.Ref
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
		var groups []idname.Ref
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
		var systemGroups []idname.Ref
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
		var networkInterfaces []idname.Ref
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
		var globalIPRanges []idname.Ref
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
		var floatingSubnets []idname.Ref
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
		var siteNetworkSubnets []idname.Ref
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

func hydrateSocketLanNetworkRuleState(
	ctx context.Context,
	_ SocketLanNetworkRule,
	apiRule *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule,
) SocketLanNetworkRuleStateOutput {
	result := SocketLanNetworkRuleStateOutput{
		ID:        types.StringValue(apiRule.ID),
		Name:      types.StringValue(apiRule.Name),
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

func hydrateSocketLanSiteState(
	_ context.Context,
	apiRule *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule,
) types.Object {
	siteAttrs := map[string]attr.Value{
		"site":  types.SetNull(idname.RefObjectType),
		"group": types.SetNull(idname.RefObjectType),
	}

	// Sites
	if len(apiRule.Site.Site) > 0 {
		var sites []attr.Value
		for _, s := range apiRule.Site.Site {
			siteObj, _ := types.ObjectValue(idname.RefAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(s.ID),
				"name": types.StringValue(s.Name),
			})
			sites = append(sites, siteObj)
		}
		siteSet, _ := types.SetValue(idname.RefObjectType, sites)
		siteAttrs["site"] = siteSet
	}

	// Groups
	if len(apiRule.Site.Group) > 0 {
		var groups []attr.Value
		for _, g := range apiRule.Site.Group {
			groupObj, _ := types.ObjectValue(idname.RefAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(g.ID),
				"name": types.StringValue(g.Name),
			})
			groups = append(groups, groupObj)
		}
		groupSet, _ := types.SetValue(idname.RefObjectType, groups)
		siteAttrs["group"] = groupSet
	}

	siteObj, _ := types.ObjectValue(socketlansite.SocketLanSiteAttrTypes, siteAttrs)
	return siteObj
}

//nolint:gocyclo,funlen
func hydrateSocketLanSourceState(
	_ context.Context,
	apiSource cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Source,
) types.Object {
	sourceAttrs := map[string]attr.Value{
		"vlan":                types.ListNull(types.Int64Type),
		"ip":                  types.ListNull(types.StringType),
		"subnet":              types.ListNull(types.StringType),
		"ip_range":            types.ListNull(iprange.RangeObjectType),
		"host":                types.SetNull(idname.RefObjectType),
		"group":               types.SetNull(idname.RefObjectType),
		"system_group":        types.SetNull(idname.RefObjectType),
		"network_interface":   types.SetNull(idname.RefObjectType),
		"global_ip_range":     types.SetNull(idname.RefObjectType),
		"floating_subnet":     types.SetNull(idname.RefObjectType),
		"site_network_subnet": types.SetNull(idname.RefObjectType),
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
			rangeObj, _ := types.ObjectValue(iprange.RangeAttrTypes, map[string]attr.Value{
				"from": types.StringValue(r.From),
				"to":   types.StringValue(r.To),
			})
			ipRanges = append(ipRanges, rangeObj)
		}
		ipRangeList, _ := types.ListValue(iprange.RangeObjectType, ipRanges)
		sourceAttrs["ip_range"] = ipRangeList
	}

	// Host
	if len(apiSource.Host) > 0 {
		var hosts []attr.Value
		for _, h := range apiSource.Host {
			hostObj, _ := types.ObjectValue(idname.RefAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(h.ID),
				"name": types.StringValue(h.Name),
			})
			hosts = append(hosts, hostObj)
		}
		hostSet, _ := types.SetValue(idname.RefObjectType, hosts)
		sourceAttrs["host"] = hostSet
	}

	// Group
	if len(apiSource.Group) > 0 {
		var groups []attr.Value
		for _, g := range apiSource.Group {
			groupObj, _ := types.ObjectValue(idname.RefAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(g.ID),
				"name": types.StringValue(g.Name),
			})
			groups = append(groups, groupObj)
		}
		groupSet, _ := types.SetValue(idname.RefObjectType, groups)
		sourceAttrs["group"] = groupSet
	}

	// System Group
	if len(apiSource.SystemGroup) > 0 {
		var systemGroups []attr.Value
		for _, sg := range apiSource.SystemGroup {
			sgObj, _ := types.ObjectValue(idname.RefAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(sg.ID),
				"name": types.StringValue(sg.Name),
			})
			systemGroups = append(systemGroups, sgObj)
		}
		sgSet, _ := types.SetValue(idname.RefObjectType, systemGroups)
		sourceAttrs["system_group"] = sgSet
	}

	// Network Interface
	if len(apiSource.NetworkInterface) > 0 {
		var networkInterfaces []attr.Value
		for _, ni := range apiSource.NetworkInterface {
			niObj, _ := types.ObjectValue(idname.RefAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(ni.ID),
				"name": types.StringValue(ni.Name),
			})
			networkInterfaces = append(networkInterfaces, niObj)
		}
		niSet, _ := types.SetValue(idname.RefObjectType, networkInterfaces)
		sourceAttrs["network_interface"] = niSet
	}

	// Global IP Range
	if len(apiSource.GlobalIPRange) > 0 {
		var globalIPRanges []attr.Value
		for _, gir := range apiSource.GlobalIPRange {
			girObj, _ := types.ObjectValue(idname.RefAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(gir.ID),
				"name": types.StringValue(gir.Name),
			})
			globalIPRanges = append(globalIPRanges, girObj)
		}
		girSet, _ := types.SetValue(idname.RefObjectType, globalIPRanges)
		sourceAttrs["global_ip_range"] = girSet
	}

	// Floating Subnet
	if len(apiSource.FloatingSubnet) > 0 {
		var floatingSubnets []attr.Value
		for _, fs := range apiSource.FloatingSubnet {
			fsObj, _ := types.ObjectValue(idname.RefAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(fs.ID),
				"name": types.StringValue(fs.Name),
			})
			floatingSubnets = append(floatingSubnets, fsObj)
		}
		fsSet, _ := types.SetValue(idname.RefObjectType, floatingSubnets)
		sourceAttrs["floating_subnet"] = fsSet
	}

	// Site Network Subnet
	if len(apiSource.SiteNetworkSubnet) > 0 {
		var siteNetworkSubnets []attr.Value
		for _, sns := range apiSource.SiteNetworkSubnet {
			snsObj, _ := types.ObjectValue(idname.RefAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(sns.ID),
				"name": types.StringValue(sns.Name),
			})
			siteNetworkSubnets = append(siteNetworkSubnets, snsObj)
		}
		snsSet, _ := types.SetValue(idname.RefObjectType, siteNetworkSubnets)
		sourceAttrs["site_network_subnet"] = snsSet
	}

	sourceObj, _ := types.ObjectValue(socketlansource.SocketLanSourceAttrTypes, sourceAttrs)
	return sourceObj
}

//nolint:gocyclo,funlen
func hydrateSocketLanDestinationState(
	_ context.Context,
	apiDest cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Destination,
) types.Object {
	destAttrs := map[string]attr.Value{
		"vlan":                types.ListNull(types.Int64Type),
		"ip":                  types.ListNull(types.StringType),
		"subnet":              types.ListNull(types.StringType),
		"ip_range":            types.ListNull(iprange.RangeObjectType),
		"host":                types.SetNull(idname.RefObjectType),
		"group":               types.SetNull(idname.RefObjectType),
		"system_group":        types.SetNull(idname.RefObjectType),
		"network_interface":   types.SetNull(idname.RefObjectType),
		"global_ip_range":     types.SetNull(idname.RefObjectType),
		"floating_subnet":     types.SetNull(idname.RefObjectType),
		"site_network_subnet": types.SetNull(idname.RefObjectType),
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
			rangeObj, _ := types.ObjectValue(iprange.RangeAttrTypes, map[string]attr.Value{
				"from": types.StringValue(r.From),
				"to":   types.StringValue(r.To),
			})
			ipRanges = append(ipRanges, rangeObj)
		}
		ipRangeList, _ := types.ListValue(iprange.RangeObjectType, ipRanges)
		destAttrs["ip_range"] = ipRangeList
	}

	// Host
	if len(apiDest.Host) > 0 {
		var hosts []attr.Value
		for _, h := range apiDest.Host {
			hostObj, _ := types.ObjectValue(idname.RefAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(h.ID),
				"name": types.StringValue(h.Name),
			})
			hosts = append(hosts, hostObj)
		}
		hostSet, _ := types.SetValue(idname.RefObjectType, hosts)
		destAttrs["host"] = hostSet
	}

	// Group
	if len(apiDest.Group) > 0 {
		var groups []attr.Value
		for _, g := range apiDest.Group {
			groupObj, _ := types.ObjectValue(idname.RefAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(g.ID),
				"name": types.StringValue(g.Name),
			})
			groups = append(groups, groupObj)
		}
		groupSet, _ := types.SetValue(idname.RefObjectType, groups)
		destAttrs["group"] = groupSet
	}

	// System Group
	if len(apiDest.SystemGroup) > 0 {
		var systemGroups []attr.Value
		for _, sg := range apiDest.SystemGroup {
			sgObj, _ := types.ObjectValue(idname.RefAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(sg.ID),
				"name": types.StringValue(sg.Name),
			})
			systemGroups = append(systemGroups, sgObj)
		}
		sgSet, _ := types.SetValue(idname.RefObjectType, systemGroups)
		destAttrs["system_group"] = sgSet
	}

	// Network Interface
	if len(apiDest.NetworkInterface) > 0 {
		var networkInterfaces []attr.Value
		for _, ni := range apiDest.NetworkInterface {
			niObj, _ := types.ObjectValue(idname.RefAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(ni.ID),
				"name": types.StringValue(ni.Name),
			})
			networkInterfaces = append(networkInterfaces, niObj)
		}
		niSet, _ := types.SetValue(idname.RefObjectType, networkInterfaces)
		destAttrs["network_interface"] = niSet
	}

	// Global IP Range
	if len(apiDest.GlobalIPRange) > 0 {
		var globalIPRanges []attr.Value
		for _, gir := range apiDest.GlobalIPRange {
			girObj, _ := types.ObjectValue(idname.RefAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(gir.ID),
				"name": types.StringValue(gir.Name),
			})
			globalIPRanges = append(globalIPRanges, girObj)
		}
		girSet, _ := types.SetValue(idname.RefObjectType, globalIPRanges)
		destAttrs["global_ip_range"] = girSet
	}

	// Floating Subnet
	if len(apiDest.FloatingSubnet) > 0 {
		var floatingSubnets []attr.Value
		for _, fs := range apiDest.FloatingSubnet {
			fsObj, _ := types.ObjectValue(idname.RefAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(fs.ID),
				"name": types.StringValue(fs.Name),
			})
			floatingSubnets = append(floatingSubnets, fsObj)
		}
		fsSet, _ := types.SetValue(idname.RefObjectType, floatingSubnets)
		destAttrs["floating_subnet"] = fsSet
	}

	// Site Network Subnet
	if len(apiDest.SiteNetworkSubnet) > 0 {
		var siteNetworkSubnets []attr.Value
		for _, sns := range apiDest.SiteNetworkSubnet {
			snsObj, _ := types.ObjectValue(idname.RefAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(sns.ID),
				"name": types.StringValue(sns.Name),
			})
			siteNetworkSubnets = append(siteNetworkSubnets, snsObj)
		}
		snsSet, _ := types.SetValue(idname.RefObjectType, siteNetworkSubnets)
		destAttrs["site_network_subnet"] = snsSet
	}

	destObj, _ := types.ObjectValue(socketlandestination.SocketLanDestinationAttrTypes, destAttrs)
	return destObj
}

func hydrateSocketLanServiceState(
	_ context.Context,
	apiService cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Service,
) types.Object {
	serviceAttrs := map[string]attr.Value{
		"simple": types.SetNull(SimpleServiceObjectType),
		"custom": types.ListNull(customservice.CustomServiceObjectType),
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
				"port_range": types.ObjectNull(iprange.RangeAttrTypes),
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
				portRangeObj, _ := types.ObjectValue(iprange.RangeAttrTypes, map[string]attr.Value{
					"from": types.StringValue(string(c.PortRange.From)),
					"to":   types.StringValue(string(c.PortRange.To)),
				})
				customAttrs["port_range"] = portRangeObj
			}

			customObj, _ := types.ObjectValue(customservice.CustomServiceAttrTypes, customAttrs)
			customServices = append(customServices, customObj)
		}
		customList, _ := types.ListValue(customservice.CustomServiceObjectType, customServices)
		serviceAttrs["custom"] = customList
	}

	serviceObj, _ := types.ObjectValue(SocketLanServiceAttrTypes, serviceAttrs)
	return serviceObj
}

func hydrateSocketLanNatState(
	_ context.Context,
	apiNat cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Nat,
) basetypes.ObjectValue {
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

// convertCreateInput2 converts between the Terraform model and SDK values.
func convertCreateInput2(
	ctx context.Context,
	plan SocketLanNetworkRule,
	ruleID string,
	queryResult *cato_go_sdk.PolicySocketLanPolicy,
	diagnostics *diag.Diagnostics,
) (resultRuleObj basetypes.ObjectValue, resultDiagstmp diag.Diagnostics) {
	var currentRule *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule
	for _, ruleWrapper := range queryResult.Policy.SocketLan.Policy.Rules {
		if ruleWrapper.Rule.ID == ruleID {
			currentRule = &ruleWrapper.Rule
			break
		}
	}

	if currentRule == nil {
		diagnostics.AddError(
			"Rule not found",
			fmt.Sprintf("Could not find created rule with ID %s", ruleID),
		)
		return resultRuleObj, resultDiagstmp
	}

	// Hydrate state from API response
	ruleState := hydrateSocketLanNetworkRuleState(ctx, plan, currentRule)

	// Build the rule object
	resultRuleObj, resultDiagstmp = types.ObjectValueFrom(ctx, SocketLanNetworkRuleRuleAttrTypes, ruleState)
	return resultRuleObj, resultDiagstmp
}

// convertReadState converts between the Terraform model and SDK values.
func convertReadState(
	ctx context.Context,
	state SocketLanNetworkRule,
	priorDiags *diag.Diagnostics,
) (resultRuleData SocketLanNetworkRuleData) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()

	diags = state.Rule.As(ctx, &resultRuleData, basetypes.ObjectAsOptions{})
	return resultRuleData
}

// convertReadState2 converts between the Terraform model and SDK values.
func convertReadState2(
	ctx context.Context,
	state SocketLanNetworkRule,
	currentRule *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule,
) (resultRuleObj basetypes.ObjectValue, resultDiagstmp diag.Diagnostics) {
	ruleState := hydrateSocketLanNetworkRuleState(ctx, state, currentRule)

	// Build the rule object
	resultRuleObj, resultDiagstmp = types.ObjectValueFrom(ctx, SocketLanNetworkRuleRuleAttrTypes, ruleState)
	return resultRuleObj, resultDiagstmp
}

// convertReadState3 converts between the Terraform model and SDK values.
func convertReadState3(
	state *SocketLanNetworkRule,
	ruleObj basetypes.ObjectValue,
	priorDiagstmp *diag.Diagnostics,
	diagnostics *diag.Diagnostics,
) {
	diagstmp := *priorDiagstmp
	defer func() { *priorDiagstmp = diagstmp }()
	curAtObj := state.At
	if !utils.HasValue(state.At) {
		curAtObj, diagstmp = types.ObjectValue(
			policyrule.PositionAttrTypes,
			map[string]attr.Value{
				"position": types.StringValue("LAST_IN_POLICY"),
				"ref":      types.StringNull(),
			},
		)
		diagnostics.Append(diagstmp...)
	}

	state.Rule = ruleObj
	state.At = curAtObj
}

// convertUpdateInput converts between the Terraform model and SDK values.
func convertUpdateInput(
	ctx context.Context,
	plan SocketLanNetworkRule,
	priorDiags *diag.Diagnostics,
) (resultRuleData SocketLanNetworkRuleData) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()

	diags = plan.Rule.As(ctx, &resultRuleData, basetypes.ObjectAsOptions{})
	return resultRuleData
}

// convertUpdateInput2 converts between the Terraform model and SDK values.
func convertUpdateInput2(
	ctx context.Context,
	plan SocketLanNetworkRule,
	priorDiags *diag.Diagnostics,
	ruleID string,
	apiInput *SocketLanNetworkRuleAPIInput,
) (resultPositionInput policyrule.Position) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()

	apiInput.update.ID = ruleID

	// Move rule if position changed

	diags = plan.At.As(ctx, &resultPositionInput, basetypes.ObjectAsOptions{})
	return resultPositionInput
}

// convertUpdateInput3 converts between the Terraform model and SDK values.
func convertUpdateInput3(ruleID string, positionInput policyrule.Position) (resultMoveInput cato_models.PolicyMoveRuleInput) {
	resultMoveInput = cato_models.PolicyMoveRuleInput{
		ID: ruleID,
		To: &cato_models.PolicyRulePositionInput{
			Position: (*cato_models.PolicyRulePositionEnum)(positionInput.Position.ValueStringPointer()),
			Ref:      positionInput.Ref.ValueStringPointer(),
		},
	}
	return resultMoveInput
}

// convertUpdateInput5 converts between the Terraform model and SDK values.
func convertUpdateInput5(
	ctx context.Context,
	plan SocketLanNetworkRule,
	ruleID string,
	queryResult *cato_go_sdk.PolicySocketLanPolicy,
	diagnostics *diag.Diagnostics,
) (resultRuleObj basetypes.ObjectValue, resultDiagstmp diag.Diagnostics) {
	var currentRule *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule
	for _, ruleWrapper := range queryResult.Policy.SocketLan.Policy.Rules {
		if ruleWrapper.Rule.ID == ruleID {
			currentRule = &ruleWrapper.Rule
			break
		}
	}

	if currentRule == nil {
		diagnostics.AddError(
			"Rule not found",
			fmt.Sprintf("Could not find rule with ID %s after update", ruleID),
		)
		return resultRuleObj, resultDiagstmp
	}

	// Hydrate state from API response
	ruleState := hydrateSocketLanNetworkRuleState(ctx, plan, currentRule)

	// Build the rule object
	resultRuleObj, resultDiagstmp = types.ObjectValueFrom(ctx, SocketLanNetworkRuleRuleAttrTypes, ruleState)
	return resultRuleObj, resultDiagstmp
}

// convertDeleteInput converts between the Terraform model and SDK values.
func convertDeleteInput(
	ctx context.Context,
	state SocketLanNetworkRule,
	priorDiags *diag.Diagnostics,
) (resultRuleData SocketLanNetworkRuleData) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()

	diags = state.Rule.As(ctx, &resultRuleData, basetypes.ObjectAsOptions{})
	return resultRuleData
}

// convertDeleteInput2 converts between the Terraform model and SDK values.
func convertDeleteInput2(ruleData SocketLanNetworkRuleData) (resultRemoveInput cato_models.SocketLanRemoveRuleInput) {
	ruleID := ruleData.ID.ValueString()

	resultRemoveInput = cato_models.SocketLanRemoveRuleInput{
		ID: ruleID,
	}
	return resultRemoveInput
}
