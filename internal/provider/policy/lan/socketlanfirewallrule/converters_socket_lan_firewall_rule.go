package socketlanfirewallrule

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
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

//nolint:gocyclo,funlen
func hydrateSocketLanFirewallRuleAPI(
	ctx context.Context,
	plan SocketLanFirewallRule,
) (SocketLanFirewallRuleAPIInput, diag.Diagnostics) {
	var result SocketLanFirewallRuleAPIInput
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
		var positionInput policyrule.Position
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
			hydrateSocketLanFirewallSourceAPI(ctx, ruleData.Source, result.create.Rule.Source, result.update.Rule.Source, &diags)
		} else {
			// Initialize with empty slices when source not specified
			initializeEmptySocketLanFirewallSource(result.create.Rule.Source, result.update.Rule.Source)
		}

		// Destination - always initialize to avoid null serialization
		result.create.Rule.Destination = &cato_models.SocketLanFirewallDestinationInput{}
		result.update.Rule.Destination = &cato_models.SocketLanFirewallDestinationUpdateInput{}
		if !ruleData.Destination.IsNull() {
			hydrateSocketLanFirewallDestinationAPI(ctx, ruleData.Destination, result.create.Rule.Destination, result.update.Rule.Destination, &diags)
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
				var apps []idname.Ref
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
				var customApps []idname.Ref
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
				var ipRanges []iprange.Range
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
				var globalIPRanges []idname.Ref
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

			// Standard services
			if !serviceData.Standard.IsNull() && len(serviceData.Standard.Elements()) > 0 {
				var standardServices []idname.Ref
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

		// Tracking
		if !ruleData.Tracking.IsNull() {
			var trackingData policyrule.Tracking
			diagstmp = ruleData.Tracking.As(ctx, &trackingData, basetypes.ObjectAsOptions{})
			diags.Append(diagstmp...)

			result.create.Rule.Tracking = &cato_models.PolicyTrackingInput{}
			result.update.Rule.Tracking = &cato_models.PolicyTrackingUpdateInput{}

			// Event
			if !trackingData.Event.IsNull() {
				var eventData policyrule.TrackingEvent
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
				var alertData policyrule.TrackingAlert
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
					var subGroups []idname.Ref
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
					var webhooks []idname.Ref
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
					var mailingLists []idname.Ref
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

//nolint:gocyclo,funlen
func hydrateSocketLanFirewallSourceAPI(
	ctx context.Context,
	sourceObj basetypes.ObjectValue,
	createInput *cato_models.SocketLanFirewallSourceInput,
	updateInput *cato_models.SocketLanFirewallSourceUpdateInput,
	diags *diag.Diagnostics,
) {
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

	// Site
	if !sourceData.Site.IsNull() && len(sourceData.Site.Elements()) > 0 {
		var sites []idname.Ref
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
func hydrateSocketLanFirewallDestinationAPI(
	ctx context.Context,
	destObj basetypes.ObjectValue,
	createInput *cato_models.SocketLanFirewallDestinationInput,
	updateInput *cato_models.SocketLanFirewallDestinationUpdateInput,
	diags *diag.Diagnostics,
) {
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

	// Site
	if !destData.Site.IsNull() && len(destData.Site.Elements()) > 0 {
		var sites []idname.Ref
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

// initializeEmptySocketLanFirewallSource initializes all list fields to empty slices to avoid null serialization
func initializeEmptySocketLanFirewallSource(
	createInput *cato_models.SocketLanFirewallSourceInput,
	updateInput *cato_models.SocketLanFirewallSourceUpdateInput,
) {
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
func initializeEmptySocketLanFirewallDestination(
	createInput *cato_models.SocketLanFirewallDestinationInput,
	updateInput *cato_models.SocketLanFirewallDestinationUpdateInput,
) {
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

func hydrateSocketLanFirewallRuleState(
	ctx context.Context,
	plan SocketLanFirewallRule,
	apiRule *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Firewall_Rule,
) SocketLanFirewallRuleStateOutput {
	result := SocketLanFirewallRuleStateOutput{
		ID:        types.StringValue(apiRule.ID),
		Name:      types.StringValue(apiRule.Name),
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
	planHasService := !planRuleData.Service.IsNull()
	result.Application = hydrateSocketLanFirewallApplicationState(ctx, apiRule.Application, planHasApplication)

	// Service - check if plan had service block defined (following Application pattern)
	result.Service = hydrateSocketLanFirewallServiceState(ctx, apiRule.Service, planHasService)

	// Tracking
	result.Tracking = hydrateSocketLanFirewallTrackingState(ctx, apiRule.Tracking)

	return result
}

//nolint:gocyclo,funlen
func hydrateSocketLanFirewallSourceState(
	_ context.Context,
	apiSource cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Firewall_Rule_Source,
) types.Object {
	sourceAttrs := map[string]attr.Value{
		"vlan":                types.ListNull(types.Int64Type),
		"mac":                 types.ListNull(types.StringType),
		"ip":                  types.ListNull(types.StringType),
		"subnet":              types.ListNull(types.StringType),
		"ip_range":            types.ListNull(iprange.RangeObjectType),
		"host":                types.SetNull(idname.RefObjectType),
		"site":                types.SetNull(idname.RefObjectType),
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

	// Site
	if len(apiSource.Site) > 0 {
		var sites []attr.Value
		for _, s := range apiSource.Site {
			siteObj, _ := types.ObjectValue(idname.RefAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(s.ID),
				"name": types.StringValue(s.Name),
			})
			sites = append(sites, siteObj)
		}
		siteSet, _ := types.SetValue(idname.RefObjectType, sites)
		sourceAttrs["site"] = siteSet
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

	sourceObj, _ := types.ObjectValue(SocketLanFirewallSourceAttrTypes, sourceAttrs)
	return sourceObj
}

//nolint:gocyclo,funlen
func hydrateSocketLanFirewallDestinationState(
	_ context.Context,
	apiDest cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Firewall_Rule_Destination,
) types.Object {
	destAttrs := map[string]attr.Value{
		"vlan":                types.ListNull(types.Int64Type),
		"ip":                  types.ListNull(types.StringType),
		"subnet":              types.ListNull(types.StringType),
		"ip_range":            types.ListNull(iprange.RangeObjectType),
		"host":                types.SetNull(idname.RefObjectType),
		"site":                types.SetNull(idname.RefObjectType),
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

	// Site
	if len(apiDest.Site) > 0 {
		var sites []attr.Value
		for _, s := range apiDest.Site {
			siteObj, _ := types.ObjectValue(idname.RefAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(s.ID),
				"name": types.StringValue(s.Name),
			})
			sites = append(sites, siteObj)
		}
		siteSet, _ := types.SetValue(idname.RefObjectType, sites)
		destAttrs["site"] = siteSet
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

	destObj, _ := types.ObjectValue(SocketLanFirewallDestinationAttrTypes, destAttrs)
	return destObj
}

//nolint:gocyclo,funlen
func hydrateSocketLanFirewallApplicationState(
	_ context.Context,
	apiApp cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Firewall_Rule_Application,
	planHasApplication bool,
) types.Object {
	appAttrs := map[string]attr.Value{
		"application":     types.SetNull(idname.RefObjectType),
		"custom_app":      types.SetNull(idname.RefObjectType),
		"domain":          types.ListNull(types.StringType),
		"fqdn":            types.ListNull(types.StringType),
		"ip":              types.ListNull(types.StringType),
		"subnet":          types.ListNull(types.StringType),
		"ip_range":        types.ListNull(iprange.RangeObjectType),
		"global_ip_range": types.SetNull(idname.RefObjectType),
	}

	// Application
	if len(apiApp.Application) > 0 {
		var apps []attr.Value
		for _, app := range apiApp.Application {
			appObj, _ := types.ObjectValue(idname.RefAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(app.ID),
				"name": types.StringValue(app.Name),
			})
			apps = append(apps, appObj)
		}
		appSet, _ := types.SetValue(idname.RefObjectType, apps)
		appAttrs["application"] = appSet
	}

	// CustomApp
	if len(apiApp.CustomApp) > 0 {
		var customApps []attr.Value
		for _, ca := range apiApp.CustomApp {
			caObj, _ := types.ObjectValue(idname.RefAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(ca.ID),
				"name": types.StringValue(ca.Name),
			})
			customApps = append(customApps, caObj)
		}
		caSet, _ := types.SetValue(idname.RefObjectType, customApps)
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
			rangeObj, _ := types.ObjectValue(iprange.RangeAttrTypes, map[string]attr.Value{
				"from": types.StringValue(r.From),
				"to":   types.StringValue(r.To),
			})
			ipRanges = append(ipRanges, rangeObj)
		}
		ipRangeList, _ := types.ListValue(iprange.RangeObjectType, ipRanges)
		appAttrs["ip_range"] = ipRangeList
	}

	// Global IP Range
	if len(apiApp.GlobalIPRange) > 0 {
		var globalIPRanges []attr.Value
		for _, gir := range apiApp.GlobalIPRange {
			girObj, _ := types.ObjectValue(idname.RefAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(gir.ID),
				"name": types.StringValue(gir.Name),
			})
			globalIPRanges = append(globalIPRanges, girObj)
		}
		girSet, _ := types.SetValue(idname.RefObjectType, globalIPRanges)
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

func hydrateSocketLanFirewallServiceState(
	_ context.Context,
	apiService cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Firewall_Rule_Service,
	planHasService bool,
) types.Object {
	serviceAttrs := map[string]attr.Value{
		"simple":   types.SetNull(simpleservice.SimpleServiceObjectType),
		"standard": types.SetNull(idname.RefObjectType),
		"custom":   types.ListNull(customservice.CustomServiceObjectType),
	}

	// Simple services
	if len(apiService.Simple) > 0 {
		var simpleServices []attr.Value
		for _, s := range apiService.Simple {
			svcObj, _ := types.ObjectValue(simpleservice.SimpleServiceAttrTypes, map[string]attr.Value{
				"name": types.StringValue(string(s.Name)),
			})
			simpleServices = append(simpleServices, svcObj)
		}
		simpleSet, _ := types.SetValue(simpleservice.SimpleServiceObjectType, simpleServices)
		serviceAttrs["simple"] = simpleSet
	}

	// Standard services
	if len(apiService.Standard) > 0 {
		var standardServices []attr.Value
		for _, s := range apiService.Standard {
			svcObj, _ := types.ObjectValue(idname.RefAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(s.ID),
				"name": types.StringValue(s.Name),
			})
			standardServices = append(standardServices, svcObj)
		}
		standardSet, _ := types.SetValue(idname.RefObjectType, standardServices)
		serviceAttrs["standard"] = standardSet
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

	// Check if there's any actual content from the API
	hasContent := len(apiService.Simple) > 0 || len(apiService.Standard) > 0 || len(apiService.Custom) > 0

	// If plan didn't have service block and API has no content, return null
	// If plan had service block (even empty), always return object structure
	if !planHasService && !hasContent {
		return types.ObjectNull(SocketLanFirewallServiceAttrTypes)
	}

	serviceObj, _ := types.ObjectValue(SocketLanFirewallServiceAttrTypes, serviceAttrs)
	return serviceObj
}

func hydrateSocketLanFirewallTrackingState(
	_ context.Context,
	apiTracking cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Firewall_Rule_Tracking,
) types.Object {
	trackingAttrs := map[string]attr.Value{
		"event": types.ObjectNull(policyrule.TrackingEventAttrTypes),
		"alert": types.ObjectNull(policyrule.TrackingAlertAttrTypes),
	}

	// Event - always create event object (API always returns tracking)
	eventObj, _ := types.ObjectValue(policyrule.TrackingEventAttrTypes, map[string]attr.Value{
		"enabled": types.BoolValue(apiTracking.Event.Enabled),
	})
	trackingAttrs["event"] = eventObj

	// Alert - always create alert object with all fields (even when enabled = false)
	alertAttrs := map[string]attr.Value{
		"enabled":            types.BoolValue(apiTracking.Alert.Enabled),
		"frequency":          types.StringNull(),
		"subscription_group": types.SetNull(idname.RefObjectType),
		"webhook":            types.SetNull(idname.RefObjectType),
		"mailing_list":       types.SetNull(idname.RefObjectType),
	}

	if apiTracking.Alert.Frequency != "" {
		alertAttrs["frequency"] = types.StringValue(string(apiTracking.Alert.Frequency))
	}

	// Subscription groups
	if len(apiTracking.Alert.SubscriptionGroup) > 0 {
		var subGroups []attr.Value
		for _, sg := range apiTracking.Alert.SubscriptionGroup {
			sgObj, _ := types.ObjectValue(idname.RefAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(sg.ID),
				"name": types.StringValue(sg.Name),
			})
			subGroups = append(subGroups, sgObj)
		}
		sgSet, _ := types.SetValue(idname.RefObjectType, subGroups)
		alertAttrs["subscription_group"] = sgSet
	}

	// Webhooks
	if len(apiTracking.Alert.Webhook) > 0 {
		var webhooks []attr.Value
		for _, wh := range apiTracking.Alert.Webhook {
			whObj, _ := types.ObjectValue(idname.RefAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(wh.ID),
				"name": types.StringValue(wh.Name),
			})
			webhooks = append(webhooks, whObj)
		}
		whSet, _ := types.SetValue(idname.RefObjectType, webhooks)
		alertAttrs["webhook"] = whSet
	}

	// Mailing lists
	if len(apiTracking.Alert.MailingList) > 0 {
		var mailingLists []attr.Value
		for _, ml := range apiTracking.Alert.MailingList {
			mlObj, _ := types.ObjectValue(idname.RefAttrTypes, map[string]attr.Value{
				"id":   types.StringValue(ml.ID),
				"name": types.StringValue(ml.Name),
			})
			mailingLists = append(mailingLists, mlObj)
		}
		mlSet, _ := types.SetValue(idname.RefObjectType, mailingLists)
		alertAttrs["mailing_list"] = mlSet
	}

	alertObj, _ := types.ObjectValue(policyrule.TrackingAlertAttrTypes, alertAttrs)
	trackingAttrs["alert"] = alertObj

	trackingObj, _ := types.ObjectValue(policyrule.TrackingAttrTypes, trackingAttrs)
	return trackingObj
}

// convertCreateInput2 converts between the Terraform model and SDK values.
func convertCreateInput2(
	ctx context.Context,
	plan SocketLanFirewallRule,
	ruleID string,
	queryResult *cato_go_sdk.PolicySocketLanPolicy,
	diagnostics *diag.Diagnostics,
) (resultRuleObj basetypes.ObjectValue, resultDiagstmp diag.Diagnostics) {
	var currentRule *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Firewall_Rule
	for _, ruleWrapper := range queryResult.Policy.SocketLan.Policy.Rules {
		for _, fwWrapper := range ruleWrapper.Rule.Firewall {
			if fwWrapper.Rule.ID == ruleID {
				currentRule = &fwWrapper.Rule
				break
			}
		}
		if currentRule != nil {
			break
		}
	}

	if currentRule == nil {
		diagnostics.AddError(
			"Rule not found",
			fmt.Sprintf("Could not find created firewall rule with ID %s", ruleID),
		)
		return resultRuleObj, resultDiagstmp
	}

	// Hydrate state from API response
	ruleState := hydrateSocketLanFirewallRuleState(ctx, plan, currentRule)

	// Build the rule object
	resultRuleObj, resultDiagstmp = types.ObjectValueFrom(ctx, SocketLanFirewallRuleRuleAttrTypes, ruleState)
	return resultRuleObj, resultDiagstmp
}

// convertReadState converts between the Terraform model and SDK values.
func convertReadState(
	ctx context.Context,
	state SocketLanFirewallRule,
	priorDiags *diag.Diagnostics,
) (resultRuleData SocketLanFirewallRuleData) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()

	diags = state.Rule.As(ctx, &resultRuleData, basetypes.ObjectAsOptions{})
	return resultRuleData
}

// convertReadState2 converts between the Terraform model and SDK values.
func convertReadState2(
	ctx context.Context,
	state SocketLanFirewallRule,
	currentRule *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Firewall_Rule,
) (resultRuleObj basetypes.ObjectValue, resultDiagstmp diag.Diagnostics) {
	ruleState := hydrateSocketLanFirewallRuleState(ctx, state, currentRule)

	// Build the rule object
	resultRuleObj, resultDiagstmp = types.ObjectValueFrom(ctx, SocketLanFirewallRuleRuleAttrTypes, ruleState)
	return resultRuleObj, resultDiagstmp
}

// convertReadState3 converts between the Terraform model and SDK values.
func convertReadState3(
	state *SocketLanFirewallRule,
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
				"position": types.StringValue("LAST_IN_RULE"),
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
	plan SocketLanFirewallRule,
	priorDiags *diag.Diagnostics,
) (resultRuleData SocketLanFirewallRuleData) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()

	diags = plan.Rule.As(ctx, &resultRuleData, basetypes.ObjectAsOptions{})
	return resultRuleData
}

// convertUpdateInput2 converts between the Terraform model and SDK values.
func convertUpdateInput2(
	ctx context.Context,
	plan SocketLanFirewallRule,
	priorDiags *diag.Diagnostics,
	ruleID string,
	apiInput *SocketLanFirewallRuleAPIInput,
) (resultPositionInput policyrule.Position) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()

	apiInput.update.ID = ruleID

	// Move rule if position changed

	diags = plan.At.As(ctx, &resultPositionInput, basetypes.ObjectAsOptions{})
	return resultPositionInput
}

// convertUpdateInput3 converts between the Terraform model and SDK values.
func convertUpdateInput3(
	ruleID string,
	positionInput policyrule.Position,
) (resultMoveInput cato_models.PolicyMoveSubRuleInput) {
	resultMoveInput = cato_models.PolicyMoveSubRuleInput{
		ID: ruleID,
		To: &cato_models.PolicySubRulePositionInput{
			Position: cato_models.PolicySubRulePositionEnum(positionInput.Position.ValueString()),
			Ref:      positionInput.Ref.ValueString(),
		},
	}
	return resultMoveInput
}

// convertUpdateInput5 converts between the Terraform model and SDK values.
func convertUpdateInput5(
	ctx context.Context,
	plan SocketLanFirewallRule,
	ruleID string,
	queryResult *cato_go_sdk.PolicySocketLanPolicy,
	diagnostics *diag.Diagnostics,
) (resultRuleObj basetypes.ObjectValue, resultDiagstmp diag.Diagnostics) {
	var currentRule *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Firewall_Rule
	for _, ruleWrapper := range queryResult.Policy.SocketLan.Policy.Rules {
		for _, fwWrapper := range ruleWrapper.Rule.Firewall {
			if fwWrapper.Rule.ID == ruleID {
				currentRule = &fwWrapper.Rule
				break
			}
		}
		if currentRule != nil {
			break
		}
	}

	if currentRule == nil {
		diagnostics.AddError(
			"Rule not found",
			fmt.Sprintf("Could not find firewall rule with ID %s after update", ruleID),
		)
		return resultRuleObj, resultDiagstmp
	}

	// Hydrate state from API response
	ruleState := hydrateSocketLanFirewallRuleState(ctx, plan, currentRule)

	// Build the rule object
	resultRuleObj, resultDiagstmp = types.ObjectValueFrom(ctx, SocketLanFirewallRuleRuleAttrTypes, ruleState)
	return resultRuleObj, resultDiagstmp
}

// convertDeleteInput converts between the Terraform model and SDK values.
func convertDeleteInput(
	ctx context.Context,
	state SocketLanFirewallRule,
	priorDiags *diag.Diagnostics,
) (resultRuleData SocketLanFirewallRuleData) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()

	diags = state.Rule.As(ctx, &resultRuleData, basetypes.ObjectAsOptions{})
	return resultRuleData
}

// convertDeleteInput2 converts between the Terraform model and SDK values.
func convertDeleteInput2(
	ruleData SocketLanFirewallRuleData,
) (resultRemoveInput cato_models.SocketLanFirewallRemoveRuleInput) {
	ruleID := ruleData.ID.ValueString()

	resultRemoveInput = cato_models.SocketLanFirewallRemoveRuleInput{
		ID: ruleID,
	}
	return resultRemoveInput
}
