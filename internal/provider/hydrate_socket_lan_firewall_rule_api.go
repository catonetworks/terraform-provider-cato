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

// SocketLanFirewallRuleApiInput holds both create and update inputs
type SocketLanFirewallRuleApiInput struct {
	create cato_models.SocketLanFirewallAddRuleInput
	update cato_models.SocketLanFirewallUpdateRuleInput
}

func hydrateSocketLanFirewallRuleApi(ctx context.Context, plan SocketLanFirewallRule) (SocketLanFirewallRuleApiInput, diag.Diagnostics) {
	var result SocketLanFirewallRuleApiInput
	var diags diag.Diagnostics

	// Initialize create input
	result.create = cato_models.SocketLanFirewallAddRuleInput{
		Rule: &cato_models.SocketLanFirewallAddRuleDataInput{},
	}
	result.update = cato_models.SocketLanFirewallUpdateRuleInput{
		Rule: &cato_models.SocketLanFirewallUpdateRuleDataInput{},
	}

	// Parse position
	if !plan.At.IsNull() {
		var positionInput PolicyRulePositionInput
		diagstmp := plan.At.As(ctx, &positionInput, basetypes.ObjectAsOptions{})
		diags.Append(diagstmp...)

		result.create.At = &cato_models.PolicySubRulePositionInput{
			Position: cato_models.PolicySubRulePositionEnum(positionInput.Position.ValueString()),
			Ref:      positionInput.Ref.ValueString(),
		}
	}

	// Parse rule data
	if !plan.Rule.IsNull() {
		var ruleData SocketLanFirewallRuleData
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
		direction := cato_models.SocketLanFirewallDirection(ruleData.Direction.ValueString())
		result.create.Rule.Direction = direction
		result.update.Rule.Direction = &direction

		// Action
		action := cato_models.SocketLanFirewallAction(ruleData.Action.ValueString())
		result.create.Rule.Action = action
		result.update.Rule.Action = &action

		// Source - always initialize to avoid null serialization
		result.create.Rule.Source = &cato_models.SocketLanFirewallSourceInput{}
		result.update.Rule.Source = &cato_models.SocketLanFirewallSourceUpdateInput{}
		if !ruleData.Source.IsNull() {
			hydrateSocketLanFirewallSourceApi(ctx, ruleData.Source, result.create.Rule.Source, result.update.Rule.Source, &diags)
		} else {
			// Initialize with empty slices when source not specified
			initializeEmptySocketLanFirewallSource(result.create.Rule.Source, result.update.Rule.Source)
		}

		// Destination - always initialize to avoid null serialization
		result.create.Rule.Destination = &cato_models.SocketLanFirewallDestinationInput{}
		result.update.Rule.Destination = &cato_models.SocketLanFirewallDestinationUpdateInput{}
		if !ruleData.Destination.IsNull() {
			hydrateSocketLanFirewallDestinationApi(ctx, ruleData.Destination, result.create.Rule.Destination, result.update.Rule.Destination, &diags)
		} else {
			// Initialize with empty slices when destination not specified
			initializeEmptySocketLanFirewallDestination(result.create.Rule.Destination, result.update.Rule.Destination)
		}

		// Application - always initialize to avoid null serialization
		result.create.Rule.Application = &cato_models.SocketLanFirewallApplicationInput{
			Application:   make([]*cato_models.ApplicationRefInput, 0),
			CustomApp:     make([]*cato_models.CustomApplicationRefInput, 0),
			Domain:        make([]string, 0),
			Fqdn:          make([]string, 0),
			GlobalIPRange: make([]*cato_models.GlobalIPRangeRefInput, 0),
			IP:            make([]string, 0),
			IPRange:       make([]*cato_models.IPAddressRangeInput, 0),
			Subnet:        make([]string, 0),
		}
		result.update.Rule.Application = &cato_models.SocketLanFirewallApplicationUpdateInput{
			Application:   make([]*cato_models.ApplicationRefInput, 0),
			CustomApp:     make([]*cato_models.CustomApplicationRefInput, 0),
			Domain:        make([]string, 0),
			Fqdn:          make([]string, 0),
			GlobalIPRange: make([]*cato_models.GlobalIPRangeRefInput, 0),
			IP:            make([]string, 0),
			IPRange:       make([]*cato_models.IPAddressRangeInput, 0),
			Subnet:        make([]string, 0),
		}

		if !ruleData.Application.IsNull() {
			var appData SocketLanFirewallApplication
			diagstmp = ruleData.Application.As(ctx, &appData, basetypes.ObjectAsOptions{})
			diags.Append(diagstmp...)

			// Application references
			if !appData.Application.IsNull() && len(appData.Application.Elements()) > 0 {
				var apps []NameIDRef
				diagstmp = appData.Application.ElementsAs(ctx, &apps, false)
				diags.Append(diagstmp...)
				for _, app := range apps {
					ObjectRefOutput, err := utils.TransformObjectRefInput(app)
					if err != nil {
						tflog.Error(ctx, err.Error())
					}
					appRef := cato_models.ApplicationRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					}
					result.create.Rule.Application.Application = append(result.create.Rule.Application.Application, &appRef)
					result.update.Rule.Application.Application = append(result.update.Rule.Application.Application, &appRef)
				}
			}

			// CustomApp references
			if !appData.CustomApp.IsNull() && len(appData.CustomApp.Elements()) > 0 {
				var customApps []NameIDRef
				diagstmp = appData.CustomApp.ElementsAs(ctx, &customApps, false)
				diags.Append(diagstmp...)
				for _, ca := range customApps {
					ObjectRefOutput, err := utils.TransformObjectRefInput(ca)
					if err != nil {
						tflog.Error(ctx, err.Error())
					}
					caRef := cato_models.CustomApplicationRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					}
					result.create.Rule.Application.CustomApp = append(result.create.Rule.Application.CustomApp, &caRef)
					result.update.Rule.Application.CustomApp = append(result.update.Rule.Application.CustomApp, &caRef)
				}
			}

			// Domain
			if !appData.Domain.IsNull() && len(appData.Domain.Elements()) > 0 {
				var domains []string
				diagstmp = appData.Domain.ElementsAs(ctx, &domains, false)
				diags.Append(diagstmp...)
				result.create.Rule.Application.Domain = domains
				result.update.Rule.Application.Domain = domains
			}

			// FQDN
			if !appData.Fqdn.IsNull() && len(appData.Fqdn.Elements()) > 0 {
				var fqdns []string
				diagstmp = appData.Fqdn.ElementsAs(ctx, &fqdns, false)
				diags.Append(diagstmp...)
				result.create.Rule.Application.Fqdn = fqdns
				result.update.Rule.Application.Fqdn = fqdns
			}

			// IP
			if !appData.IP.IsNull() && len(appData.IP.Elements()) > 0 {
				var ips []string
				diagstmp = appData.IP.ElementsAs(ctx, &ips, false)
				diags.Append(diagstmp...)
				result.create.Rule.Application.IP = ips
				result.update.Rule.Application.IP = ips
			}

			// Subnet
			if !appData.Subnet.IsNull() && len(appData.Subnet.Elements()) > 0 {
				var subnets []string
				diagstmp = appData.Subnet.ElementsAs(ctx, &subnets, false)
				diags.Append(diagstmp...)
				result.create.Rule.Application.Subnet = subnets
				result.update.Rule.Application.Subnet = subnets
			}

			// IP Range
			if !appData.IPRange.IsNull() && len(appData.IPRange.Elements()) > 0 {
				var ipRanges []Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_IPRange
				diagstmp = appData.IPRange.ElementsAs(ctx, &ipRanges, false)
				diags.Append(diagstmp...)
				for _, r := range ipRanges {
					result.create.Rule.Application.IPRange = append(result.create.Rule.Application.IPRange, &cato_models.IPAddressRangeInput{
						From: r.From.ValueString(),
						To:   r.To.ValueString(),
					})
					result.update.Rule.Application.IPRange = append(result.update.Rule.Application.IPRange, &cato_models.IPAddressRangeInput{
						From: r.From.ValueString(),
						To:   r.To.ValueString(),
					})
				}
			}

			// Global IP Range
			if !appData.GlobalIPRange.IsNull() && len(appData.GlobalIPRange.Elements()) > 0 {
				var globalIPRanges []NameIDRef
				diagstmp = appData.GlobalIPRange.ElementsAs(ctx, &globalIPRanges, false)
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
					result.create.Rule.Application.GlobalIPRange = append(result.create.Rule.Application.GlobalIPRange, &girRef)
					result.update.Rule.Application.GlobalIPRange = append(result.update.Rule.Application.GlobalIPRange, &girRef)
				}
			}
		}

		// Service - always initialize to avoid null serialization
		result.create.Rule.Service = &cato_models.SocketLanFirewallServiceTypeInput{
			Simple:   make([]*cato_models.SimpleServiceInput, 0),
			Standard: make([]*cato_models.ServiceRefInput, 0),
			Custom:   make([]*cato_models.CustomServiceInput, 0),
		}
		result.update.Rule.Service = &cato_models.SocketLanFirewallServiceTypeUpdateInput{
			Simple:   make([]*cato_models.SimpleServiceInput, 0),
			Standard: make([]*cato_models.ServiceRefInput, 0),
			Custom:   make([]*cato_models.CustomServiceInput, 0),
		}

		if !ruleData.Service.IsNull() {
			var serviceData SocketLanFirewallService
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

			// Standard services
			if !serviceData.Standard.IsNull() && len(serviceData.Standard.Elements()) > 0 {
				var standardServices []NameIDRef
				diagstmp = serviceData.Standard.ElementsAs(ctx, &standardServices, false)
				diags.Append(diagstmp...)
				for _, svc := range standardServices {
					ObjectRefOutput, err := utils.TransformObjectRefInput(svc)
					if err != nil {
						tflog.Error(ctx, err.Error())
					}
					svcRef := cato_models.ServiceRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					}
					result.create.Rule.Service.Standard = append(result.create.Rule.Service.Standard, &svcRef)
					result.update.Rule.Service.Standard = append(result.update.Rule.Service.Standard, &svcRef)
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

		// Tracking
		if !ruleData.Tracking.IsNull() {
			var trackingData Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking
			diagstmp = ruleData.Tracking.As(ctx, &trackingData, basetypes.ObjectAsOptions{})
			diags.Append(diagstmp...)

			result.create.Rule.Tracking = &cato_models.PolicyTrackingInput{}
			result.update.Rule.Tracking = &cato_models.PolicyTrackingUpdateInput{}

			// Event
			if !trackingData.Event.IsNull() {
				var eventData Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Event
				diagstmp = trackingData.Event.As(ctx, &eventData, basetypes.ObjectAsOptions{})
				diags.Append(diagstmp...)

				result.create.Rule.Tracking.Event = &cato_models.PolicyRuleTrackingEventInput{
					Enabled: eventData.Enabled.ValueBool(),
				}
				result.update.Rule.Tracking.Event = &cato_models.PolicyRuleTrackingEventUpdateInput{
					Enabled: eventData.Enabled.ValueBoolPointer(),
				}
			}

			// Alert
			if !trackingData.Alert.IsNull() {
				var alertData Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Alert
				diagstmp = trackingData.Alert.As(ctx, &alertData, basetypes.ObjectAsOptions{})
				diags.Append(diagstmp...)

				// Initialize alert with empty slices to avoid null serialization
				createAlert := &cato_models.PolicyRuleTrackingAlertInput{
					Enabled:           alertData.Enabled.ValueBool(),
					SubscriptionGroup: make([]*cato_models.SubscriptionGroupRefInput, 0),
					Webhook:           make([]*cato_models.SubscriptionWebhookRefInput, 0),
					MailingList:       make([]*cato_models.SubscriptionMailingListRefInput, 0),
				}
				updateAlert := &cato_models.PolicyRuleTrackingAlertUpdateInput{
					Enabled:           alertData.Enabled.ValueBoolPointer(),
					SubscriptionGroup: make([]*cato_models.SubscriptionGroupRefInput, 0),
					Webhook:           make([]*cato_models.SubscriptionWebhookRefInput, 0),
					MailingList:       make([]*cato_models.SubscriptionMailingListRefInput, 0),
				}

				// Frequency
				if !alertData.Frequency.IsNull() {
					freq := cato_models.PolicyRuleTrackingFrequencyEnum(alertData.Frequency.ValueString())
					createAlert.Frequency = freq
					updateAlert.Frequency = &freq
				}

				// Subscription groups
				if !alertData.SubscriptionGroup.IsNull() && len(alertData.SubscriptionGroup.Elements()) > 0 {
					var subGroups []NameIDRef
					diagstmp = alertData.SubscriptionGroup.ElementsAs(ctx, &subGroups, false)
					diags.Append(diagstmp...)
					for _, sg := range subGroups {
						ObjectRefOutput, err := utils.TransformObjectRefInput(sg)
						if err != nil {
							tflog.Error(ctx, err.Error())
						}
						sgRef := cato_models.SubscriptionGroupRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						}
						createAlert.SubscriptionGroup = append(createAlert.SubscriptionGroup, &sgRef)
						updateAlert.SubscriptionGroup = append(updateAlert.SubscriptionGroup, &sgRef)
					}
				}

				// Webhooks
				if !alertData.Webhook.IsNull() && len(alertData.Webhook.Elements()) > 0 {
					var webhooks []NameIDRef
					diagstmp = alertData.Webhook.ElementsAs(ctx, &webhooks, false)
					diags.Append(diagstmp...)
					for _, wh := range webhooks {
						ObjectRefOutput, err := utils.TransformObjectRefInput(wh)
						if err != nil {
							tflog.Error(ctx, err.Error())
						}
						whRef := cato_models.SubscriptionWebhookRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						}
						createAlert.Webhook = append(createAlert.Webhook, &whRef)
						updateAlert.Webhook = append(updateAlert.Webhook, &whRef)
					}
				}

				// Mailing lists
				if !alertData.MailingList.IsNull() && len(alertData.MailingList.Elements()) > 0 {
					var mailingLists []NameIDRef
					diagstmp = alertData.MailingList.ElementsAs(ctx, &mailingLists, false)
					diags.Append(diagstmp...)
					for _, ml := range mailingLists {
						ObjectRefOutput, err := utils.TransformObjectRefInput(ml)
						if err != nil {
							tflog.Error(ctx, err.Error())
						}
						mlRef := cato_models.SubscriptionMailingListRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						}
						createAlert.MailingList = append(createAlert.MailingList, &mlRef)
						updateAlert.MailingList = append(updateAlert.MailingList, &mlRef)
					}
				}

				result.create.Rule.Tracking.Alert = createAlert
				result.update.Rule.Tracking.Alert = updateAlert
			}
		}
	}

	return result, diags
}

func hydrateSocketLanFirewallSourceApi(ctx context.Context, sourceObj basetypes.ObjectValue, createInput *cato_models.SocketLanFirewallSourceInput, updateInput *cato_models.SocketLanFirewallSourceUpdateInput, diags *diag.Diagnostics) {
	var sourceData SocketLanFirewallSource
	diagstmp := sourceObj.As(ctx, &sourceData, basetypes.ObjectAsOptions{})
	diags.Append(diagstmp...)

	// Initialize all list fields to empty slices to avoid null serialization
	createInput.Vlan = make([]scalars.Vlan, 0)
	updateInput.Vlan = make([]scalars.Vlan, 0)
	createInput.Mac = make([]string, 0)
	updateInput.Mac = make([]string, 0)
	createInput.IP = make([]string, 0)
	updateInput.IP = make([]string, 0)
	createInput.Subnet = make([]string, 0)
	updateInput.Subnet = make([]string, 0)
	createInput.IPRange = make([]*cato_models.IPAddressRangeInput, 0)
	updateInput.IPRange = make([]*cato_models.IPAddressRangeInput, 0)
	createInput.Host = make([]*cato_models.HostRefInput, 0)
	updateInput.Host = make([]*cato_models.HostRefInput, 0)
	createInput.Site = make([]*cato_models.SiteRefInput, 0)
	updateInput.Site = make([]*cato_models.SiteRefInput, 0)
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

	// Mac
	if !sourceData.Mac.IsNull() && len(sourceData.Mac.Elements()) > 0 {
		var macs []string
		diagstmp = sourceData.Mac.ElementsAs(ctx, &macs, false)
		diags.Append(diagstmp...)
		createInput.Mac = macs
		updateInput.Mac = macs
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

	// Site
	if !sourceData.Site.IsNull() && len(sourceData.Site.Elements()) > 0 {
		var sites []NameIDRef
		diagstmp = sourceData.Site.ElementsAs(ctx, &sites, false)
		diags.Append(diagstmp...)
		for _, s := range sites {
			ObjectRefOutput, err := utils.TransformObjectRefInput(s)
			if err != nil {
				tflog.Error(ctx, err.Error())
			}
			siteRef := cato_models.SiteRefInput{
				By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
				Input: ObjectRefOutput.Input,
			}
			createInput.Site = append(createInput.Site, &siteRef)
			updateInput.Site = append(updateInput.Site, &siteRef)
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

func hydrateSocketLanFirewallDestinationApi(ctx context.Context, destObj basetypes.ObjectValue, createInput *cato_models.SocketLanFirewallDestinationInput, updateInput *cato_models.SocketLanFirewallDestinationUpdateInput, diags *diag.Diagnostics) {
	var destData SocketLanFirewallDestination
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
	createInput.Site = make([]*cato_models.SiteRefInput, 0)
	updateInput.Site = make([]*cato_models.SiteRefInput, 0)
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

	// Site
	if !destData.Site.IsNull() && len(destData.Site.Elements()) > 0 {
		var sites []NameIDRef
		diagstmp = destData.Site.ElementsAs(ctx, &sites, false)
		diags.Append(diagstmp...)
		for _, s := range sites {
			ObjectRefOutput, err := utils.TransformObjectRefInput(s)
			if err != nil {
				tflog.Error(ctx, err.Error())
			}
			siteRef := cato_models.SiteRefInput{
				By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
				Input: ObjectRefOutput.Input,
			}
			createInput.Site = append(createInput.Site, &siteRef)
			updateInput.Site = append(updateInput.Site, &siteRef)
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

// initializeEmptySocketLanFirewallSource initializes all list fields to empty slices to avoid null serialization
func initializeEmptySocketLanFirewallSource(createInput *cato_models.SocketLanFirewallSourceInput, updateInput *cato_models.SocketLanFirewallSourceUpdateInput) {
	createInput.Vlan = make([]scalars.Vlan, 0)
	updateInput.Vlan = make([]scalars.Vlan, 0)
	createInput.Mac = make([]string, 0)
	updateInput.Mac = make([]string, 0)
	createInput.IP = make([]string, 0)
	updateInput.IP = make([]string, 0)
	createInput.Subnet = make([]string, 0)
	updateInput.Subnet = make([]string, 0)
	createInput.IPRange = make([]*cato_models.IPAddressRangeInput, 0)
	updateInput.IPRange = make([]*cato_models.IPAddressRangeInput, 0)
	createInput.Host = make([]*cato_models.HostRefInput, 0)
	updateInput.Host = make([]*cato_models.HostRefInput, 0)
	createInput.Site = make([]*cato_models.SiteRefInput, 0)
	updateInput.Site = make([]*cato_models.SiteRefInput, 0)
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
}

// initializeEmptySocketLanFirewallDestination initializes all list fields to empty slices to avoid null serialization
func initializeEmptySocketLanFirewallDestination(createInput *cato_models.SocketLanFirewallDestinationInput, updateInput *cato_models.SocketLanFirewallDestinationUpdateInput) {
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
	createInput.Site = make([]*cato_models.SiteRefInput, 0)
	updateInput.Site = make([]*cato_models.SiteRefInput, 0)
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
}
