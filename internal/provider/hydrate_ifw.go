package provider

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	"github.com/fatih/structs"
	"github.com/gobeam/stringy"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func hydrateIfwRuleState(ctx context.Context, state InternetFirewallRule, currentRule *cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_Rule, req resource.ReadRequest, resp *resource.ReadResponse, diags diag.Diagnostics) {

	ruleInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule{}
	// sourceInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source{}
	// destInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination{}

	diags = state.Rule.As(ctx, &ruleInput, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Rule -> Name
	ruleInput.Name = types.StringValue(currentRule.Name)
	// Rule -> Description
	ruleInput.Description = types.StringValue(currentRule.Description)
	// Rule -> Action
	ruleInput.Action = types.StringValue(currentRule.Action.String())
	// Rule -> ConnectionOrigin
	ruleInput.ConnectionOrigin = types.StringValue(currentRule.ConnectionOrigin.String())

	// diags = ruleInput.Source.As(ctx, &sourceInput, basetypes.ObjectAsOptions{})
	// resp.Diagnostics.Append(diags...)

	// ruleInput.Source, diags = types.ObjectValueFrom(ctx, ruleInput.Source.AttributeTypes(ctx), sourceInput)
	// resp.Diagnostics.Append(diags...)

	// //////////// rule.source ///////////////

	curRuleSourceObj, diags := types.ObjectValue(
		map[string]attr.Type{
			"ip":                  types.ListType{ElemType: types.StringType},
			"host":                types.ListType{ElemType: NameIDObjectType},
			"site":                types.ListType{ElemType: NameIDObjectType},
			"subnet":              types.ListType{ElemType: types.StringType},
			"ip_range":            types.ListType{ElemType: FromToObjectType},
			"global_ip_range":     types.ListType{ElemType: NameIDObjectType},
			"network_interface":   types.ListType{ElemType: NameIDObjectType},
			"site_network_subnet": types.ListType{ElemType: NameIDObjectType},
			"floating_subnet":     types.ListType{ElemType: NameIDObjectType},
			"user":                types.ListType{ElemType: NameIDObjectType},
			"users_group":         types.ListType{ElemType: NameIDObjectType},
			"group":               types.ListType{ElemType: NameIDObjectType},
			"system_group":        types.ListType{ElemType: NameIDObjectType},
		},
		map[string]attr.Value{
			"ip":                  types.ListNull(types.StringType),
			"host":                types.ListNull(NameIDObjectType),
			"site":                types.ListNull(NameIDObjectType),
			"subnet":              types.ListNull(types.StringType),
			"ip_range":            types.ListNull(FromToObjectType),
			"global_ip_range":     types.ListNull(NameIDObjectType),
			"network_interface":   types.ListNull(NameIDObjectType),
			"site_network_subnet": types.ListNull(NameIDObjectType),
			"floating_subnet":     types.ListNull(NameIDObjectType),
			"user":                types.ListNull(NameIDObjectType),
			"users_group":         types.ListNull(NameIDObjectType),
			"group":               types.ListNull(NameIDObjectType),
			"system_group":        types.ListNull(NameIDObjectType),
		},
	)
	resp.Diagnostics.Append(diags...)
	curRuleSourceObjAttrs := curRuleSourceObj.Attributes()

	// // rule.source.subnet[]
	curSourceSubnets := []string{}
	for _, sourceSubnet := range currentRule.Source.Subnet {
		curSourceSubnets = append(curSourceSubnets, sourceSubnet)
	}
	curSourceSubnetList, diagstmp := types.ListValueFrom(ctx, types.StringType, curSourceSubnets)
	diags = append(diags, diagstmp...)
	curRuleSourceObjAttrs["subnet"] = curSourceSubnetList

	curSourceSourceIps := []string{}
	for _, sourceIp := range currentRule.Source.IP {
		curSourceSourceIps = append(curSourceSourceIps, sourceIp)
	}
	curSourceSourceIpList, diagstmp := types.ListValueFrom(ctx, types.StringType, curSourceSubnets)
	diags = append(diags, diagstmp...)
	curRuleSourceObjAttrs["ip"] = curSourceSourceIpList

	// rule.source.host[]
	var curSourceHosts []types.Object
	tflog.Warn(ctx, "ruleResponse.Source.Host - "+fmt.Sprintf("%v", currentRule.Source.Host))
	for _, host := range currentRule.Source.Host {
		curSourceHostObj := parseNameID(ctx, host)
		curSourceHosts = append(curSourceHosts, curSourceHostObj)
	}
	curSourceHostValues := make([]attr.Value, len(curSourceHosts))
	for i, v := range curSourceHosts {
		curSourceHostValues[i] = v
	}
	curRuleSourceObjAttrs["host"], diagstmp = types.ListValue(NameIDObjectType, curSourceHostValues)
	diags = append(diags, diagstmp...)

	// sourceInput.Subnet = curSourceSubnetList
	curRuleSourceObj, diags = types.ObjectValue(curRuleSourceObj.AttributeTypes(ctx), curRuleSourceObjAttrs)
	resp.Diagnostics.Append(diags...)
	ruleInput.Source = curRuleSourceObj

	////////////// end rule.source ///////////////

	// // Rule -> Source -> IP
	// if currentRule.Source.IP != nil {
	// 	if len(currentRule.Source.IP) > 0 {
	// 		sourceInput.IP, diags = types.ListValueFrom(ctx, sourceInput.IP.ElementType(ctx), currentRule.Source.IP)
	// 		resp.Diagnostics.Append(diags...)
	// 	}
	// }

	// // Rule -> Source -> Subnet
	// if currentRule.Source.Subnet != nil {
	// 	if len(currentRule.Source.Subnet) > 0 {
	// 		sourceInput.Subnet, diags = types.ListValueFrom(ctx, sourceInput.Subnet.ElementType(ctx), currentRule.Source.Subnet)
	// 		resp.Diagnostics.Append(diags...)
	// 	}
	// }

	// // Rule -> Source -> Host
	// if currentRule.Source.Host != nil {
	// 	if len(currentRule.Source.Host) > 0 {
	// 		sourceInput.Host, diags = types.ListValueFrom(ctx, sourceInput.Host.ElementType(ctx), parseNameIDList(ctx, currentRule.Source.Host, resp))
	// 		resp.Diagnostics.Append(diags...)
	// 	}
	// }

	// // Rule -> Source -> Site
	// if currentRule.Source.Site != nil {
	// 	if len(currentRule.Source.Site) > 0 {
	// 		sourceInput.Site, diags = types.ListValueFrom(ctx, sourceInput.Site.ElementType(ctx), parseNameIDList(ctx, currentRule.Source.Site, resp))
	// 		resp.Diagnostics.Append(diags...)
	// 	}
	// }

	// // Rule -> Source -> IPRange
	// if currentRule.Source.IPRange != nil {
	// 	if len(currentRule.Source.IPRange) > 0 {
	// 		sourceInput.IPRange, diags = types.ListValueFrom(ctx, sourceInput.IPRange.ElementType(ctx), parseFromToList(ctx, currentRule.Source.IPRange, resp))
	// 		resp.Diagnostics.Append(diags...)
	// 	}
	// }

	// // Rule -> Source -> GlobalIPRange
	// if currentRule.Source.GlobalIPRange != nil {
	// 	if len(currentRule.Source.GlobalIPRange) > 0 {
	// 		sourceInput.GlobalIPRange, diags = types.ListValueFrom(ctx, sourceInput.GlobalIPRange.ElementType(ctx), parseNameIDList(ctx, currentRule.Source.GlobalIPRange, resp))
	// 		resp.Diagnostics.Append(diags...)
	// 	}
	// }

	// // Rule -> Source -> NetworkInterface
	// if currentRule.Source.NetworkInterface != nil {
	// 	if len(currentRule.Source.NetworkInterface) > 0 {
	// 		sourceInput.NetworkInterface, diags = types.ListValueFrom(ctx, sourceInput.NetworkInterface.ElementType(ctx), parseNameIDList(ctx, currentRule.Source.NetworkInterface, resp))
	// 		resp.Diagnostics.Append(diags...)
	// 	}
	// }

	// // Rule -> Source -> SiteNetworkSubnet
	// if currentRule.Source.SiteNetworkSubnet != nil {
	// 	if len(currentRule.Source.SiteNetworkSubnet) > 0 {
	// 		sourceInput.SiteNetworkSubnet, diags = types.ListValueFrom(ctx, sourceInput.SiteNetworkSubnet.ElementType(ctx), parseNameIDList(ctx, currentRule.Source.SiteNetworkSubnet, resp))
	// 		resp.Diagnostics.Append(diags...)
	// 	}
	// }

	// // Rule -> Source -> FloatingSubnet
	// if currentRule.Source.FloatingSubnet != nil {
	// 	if len(currentRule.Source.FloatingSubnet) > 0 {
	// 		sourceInput.FloatingSubnet, diags = types.ListValueFrom(ctx, sourceInput.FloatingSubnet.ElementType(ctx), parseNameIDList(ctx, currentRule.Source.FloatingSubnet, resp))
	// 		resp.Diagnostics.Append(diags...)
	// 	}
	// }

	// // Rule -> Source -> User
	// if currentRule.Source.User != nil {
	// 	if len(currentRule.Source.User) > 0 {
	// 		sourceInput.User, diags = types.ListValueFrom(ctx, sourceInput.User.ElementType(ctx), parseNameIDList(ctx, currentRule.Source.User, resp))
	// 		resp.Diagnostics.Append(diags...)
	// 	}
	// }

	// // Rule -> Source -> UsersGroup
	// if currentRule.Source.UsersGroup != nil {
	// 	if len(currentRule.Source.UsersGroup) > 0 {
	// 		sourceInput.UsersGroup, diags = types.ListValueFrom(ctx, sourceInput.UsersGroup.ElementType(ctx), parseNameIDList(ctx, currentRule.Source.UsersGroup, resp))
	// 		resp.Diagnostics.Append(diags...)
	// 	}
	// }

	// // Rule -> Source -> Group
	// if currentRule.Source.Group != nil {
	// 	if len(currentRule.Source.Group) > 0 {
	// 		sourceInput.Group, diags = types.ListValueFrom(ctx, sourceInput.Group.ElementType(ctx), parseNameIDList(ctx, currentRule.Source.Group, resp))
	// 		resp.Diagnostics.Append(diags...)
	// 	}
	// }

	// // Rule -> Source -> SystemGroup
	// if currentRule.Source.SystemGroup != nil {
	// 	if len(currentRule.Source.SystemGroup) > 0 {
	// 		sourceInput.SystemGroup, diags = types.ListValueFrom(ctx, sourceInput.SystemGroup.ElementType(ctx), parseNameIDList(ctx, currentRule.Source.SystemGroup, resp))
	// 		resp.Diagnostics.Append(diags...)
	// 	}
	// }

	// curRule["source"] = curRuleSourceObj

	// Rule -> Source
	// ruleInput.Source, diags = types.ObjectValueFrom(ctx, ruleInput.Source.AttributeTypes(ctx), sourceInput)
	// resp.Diagnostics.Append(diags...)

	// // Rule -> Country
	// ruleInput.Country, diags = basetypes.NewListValueFrom(ctx, ruleInput.Country.ElementType(ctx), parseNameIDList(ctx, currentRule.Country, resp))
	// resp.Diagnostics.Append(diags...)

	// // // Rule -> Device
	// ruleInput.Device, diags = basetypes.NewListValueFrom(ctx, ruleInput.Device.ElementType(ctx), parseNameIDList(ctx, currentRule.Device, resp))
	// resp.Diagnostics.Append(diags...)

	// // Rule -> DeviceOS
	// ruleInput.DeviceOs, diags = types.ListValueFrom(ctx, ruleInput.DeviceOs.ElementType(ctx), currentRule.DeviceOs)
	// resp.Diagnostics.Append(diags...)

	// // Rule -> Destination
	// diags = ruleInput.Destination.As(ctx, &destInput, basetypes.ObjectAsOptions{})
	// resp.Diagnostics.Append(diags...)

	// // // Rule -> Destination -> IP
	// if currentRule.Destination.IP != nil {
	// 	if len(currentRule.Destination.IP) > 0 {
	// 		destInput.IP, diags = basetypes.NewListValueFrom(ctx, destInput.IP.ElementType(ctx), currentRule.Destination.IP)
	// 		resp.Diagnostics.Append(diags...)
	// 	}
	// }

	// // Rule -> Destination -> Subnet
	// if currentRule.Destination.Subnet != nil {
	// 	if len(currentRule.Destination.Subnet) > 0 {
	// 		destInput.Subnet, diags = basetypes.NewListValueFrom(ctx, destInput.Subnet.ElementType(ctx), currentRule.Destination.Subnet)
	// 		resp.Diagnostics.Append(diags...)
	// 	}
	// }

	// // Rule -> Destination -> Domain
	// tflog.Info(ctx, "destInput.Domain", map[string]interface{}{
	// 	"count": len(currentRule.Destination.Domain),
	// })

	// // Rule -> Destination -> Domain
	// if currentRule.Destination.Domain != nil {
	// 	if len(currentRule.Destination.Domain) > 0 {
	// 		destInput.Domain, diags = types.ListValueFrom(ctx, types.StringType, currentRule.Destination.Domain)
	// 		resp.Diagnostics.Append(diags...)
	// 	}
	// }

	// // Rule -> Destination -> Fqdn
	// if currentRule.Destination.Fqdn != nil {
	// 	if len(currentRule.Destination.Fqdn) > 0 {
	// 		destInput.Fqdn, diags = types.ListValueFrom(ctx, types.StringType, currentRule.Destination.Fqdn)
	// 		resp.Diagnostics.Append(diags...)
	// 	}
	// }

	// // Rule -> Destination -> RemoteAsn
	// if currentRule.Destination.RemoteAsn != nil {
	// 	if len(currentRule.Destination.RemoteAsn) > 0 {
	// 		destInput.RemoteAsn, diags = types.ListValueFrom(ctx, types.StringType, currentRule.Destination.RemoteAsn)
	// 		resp.Diagnostics.Append(diags...)
	// 	}
	// }

	// // Rule -> Destination -> Application
	// if currentRule.Destination.Application != nil {
	// 	if len(currentRule.Destination.Application) > 0 {
	// 		destInput.Application, diags = types.ListValueFrom(ctx, destInput.Application.ElementType(ctx), parseNameIDList(ctx, currentRule.Destination.Application, resp))
	// 		resp.Diagnostics.Append(diags...)
	// 	}
	// }

	// // Rule -> Destination -> CustomApp
	// if currentRule.Destination.CustomApp != nil {
	// 	if len(currentRule.Destination.CustomApp) > 0 {
	// 		destInput.CustomApp, diags = types.ListValueFrom(ctx, destInput.CustomApp.ElementType(ctx), parseNameIDList(ctx, currentRule.Destination.CustomApp, resp))
	// 		resp.Diagnostics.Append(diags...)
	// 	}
	// }

	// // Rule -> Destination -> IPRange
	// // if currentRule.Destination.IPRange != nil {
	// // 	if len(currentRule.Destination.IPRange) > 0 {
	// // 		destInput.IPRange, diags = types.ListValueFrom(ctx, destInput.IPRange.ElementType(ctx), parseFromToList(ctx, currentRule.Destination.IPRange, resp))
	// // 		resp.Diagnostics.Append(diags...)
	// // 	}
	// // }

	// // Rule -> Destination -> GlobalIPRange
	// if currentRule.Destination.GlobalIPRange != nil {
	// 	if len(currentRule.Destination.GlobalIPRange) > 0 {
	// 		destInput.GlobalIPRange, diags = types.ListValueFrom(ctx, destInput.GlobalIPRange.ElementType(ctx), parseNameIDList(ctx, currentRule.Destination.GlobalIPRange, resp))
	// 		resp.Diagnostics.Append(diags...)
	// 	}
	// }

	// // Rule -> Destination -> AppCategory
	// if currentRule.Destination.AppCategory != nil {
	// 	if len(currentRule.Destination.AppCategory) > 0 {
	// 		destInput.GlobalIPRange, diags = types.ListValueFrom(ctx, destInput.AppCategory.ElementType(ctx), parseNameIDList(ctx, currentRule.Destination.AppCategory, resp))
	// 		resp.Diagnostics.Append(diags...)
	// 	}
	// }

	// // Rule -> Destination -> CustomCategory
	// if currentRule.Destination.CustomCategory != nil {
	// 	if len(currentRule.Destination.CustomCategory) > 0 {
	// 		destInput.CustomCategory, diags = types.ListValueFrom(ctx, destInput.CustomCategory.ElementType(ctx), parseNameIDList(ctx, currentRule.Destination.CustomCategory, resp))
	// 		resp.Diagnostics.Append(diags...)
	// 	}
	// }

	// // Rule -> Destination -> SanctionedAppsCategory
	// if currentRule.Destination.SanctionedAppsCategory != nil {
	// 	if len(currentRule.Destination.SanctionedAppsCategory) > 0 {
	// 		destInput.SanctionedAppsCategory, diags = types.ListValueFrom(ctx, destInput.SanctionedAppsCategory.ElementType(ctx), parseNameIDList(ctx, currentRule.Destination.SanctionedAppsCategory, resp))
	// 		resp.Diagnostics.Append(diags...)
	// 	}
	// }

	// // Rule -> Destination -> Country
	// if currentRule.Destination.Country != nil {
	// 	if len(currentRule.Destination.Country) > 0 {
	// 		destInput.Country, diags = types.ListValueFrom(ctx, destInput.Country.ElementType(ctx), parseNameIDList(ctx, currentRule.Destination.Country, resp))
	// 		resp.Diagnostics.Append(diags...)
	// 	}
	// }

	// // Rule -> Service
	// if len(currentRule.Service.Custom) > 0 || len(currentRule.Service.Standard) > 0 {
	// 	var serviceInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service
	// 	diags = ruleInput.Service.As(ctx, &serviceInput, basetypes.ObjectAsOptions{})
	// 	resp.Diagnostics.Append(diags...)

	// 	// Rule -> Service -> Standard
	// 	if currentRule.Service.Standard != nil {
	// 		if len(currentRule.Service.Standard) > 0 {
	// 			serviceInput.Standard, diags = types.ListValueFrom(ctx, serviceInput.Standard.ElementType(ctx), parseNameIDList(ctx, currentRule.Service.Standard, resp))
	// 			resp.Diagnostics.Append(diags...)
	// 		}
	// 	}

	// 	// Rule -> Service -> Custom
	// 	// elementsServiceCustomInput := []Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service_Custom{}
	// 	// diags = serviceInput.Custom.ElementsAs(ctx, &elementsServiceCustomInput, false)
	// 	// resp.Diagnostics.Append(diags...)
	// 	// tflog.Info(ctx, "currentRule.Service.Custom", map[string]interface{}{
	// 	// 	"currentRule.Service.Custom": currentRule.Service.Custom,
	// 	// 	"elementsServiceCustomInput": elementsServiceCustomInput,
	// 	// })

	// 	// for _, customServiceElementsList := range currentRule.Service.Custom {
	// 	// 	custServiceInternal := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service_Custom{}
	// 	// 	diags = ruleInput.Source.As(ctx, &sourceInput, basetypes.ObjectAsOptions{})
	// 	// 	resp.Diagnostics.Append(diags...)

	// 	// 	custServiceInternal.Protocol = basetypes.NewStringValue(customServiceElementsList.Protocol.String())

	// 	// 	// Rule -> Service -> Custom -> Port

	// 	// 	// if !customServiceElementsList.Port.IsNull() {
	// 	// 	// if len(customServiceElementsList.Port) > 0 {
	// 	// 	// 	var elementsServiceCustomPortInput []attr.Value
	// 	// 	// 	for _, v := range customServiceElementsList.Port {
	// 	// 	// 		elementsServiceCustomPortInput = append(elementsServiceCustomPortInput, basetypes.NewStringValue(string(v)))
	// 	// 	// 	}

	// 	// 	// 	custServiceInternal.Port, diags = basetypes.NewListValue(types.StringType, elementsServiceCustomPortInput)
	// 	// 	// 	resp.Diagnostics.Append(diags...)
	// 	// 	// } else {
	// 	// 	// 	custServiceInternal.Port = basetypes.NewListNull(types.StringType)
	// 	// 	// }

	// 	// 	// Rule -> Service -> Custom -> PortRange
	// 	// 	if string(*customServiceElementsList.PortRange.GetFrom()) != "" && string(*customServiceElementsList.PortRange.GetTo()) != "" {
	// 	// 		custServiceInternalPortrange := &Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service_Custom_PortRange{}

	// 	// 		custServiceInternalPortrange.From = basetypes.NewStringValue(string(*customServiceElementsList.PortRange.GetFrom()))
	// 	// 		custServiceInternalPortrange.To = basetypes.NewStringValue(string(*customServiceElementsList.PortRange.GetTo()))

	// 	// 		custServiceInternal.PortRange, diags = basetypes.NewObjectValueFrom(ctx, mapAttributeTypes(ctx, custServiceInternalPortrange, resp), custServiceInternalPortrange)
	// 	// 		resp.Diagnostics.Append(diags...)

	// 	// 	} else {
	// 	// 		custServiceInternalPortrange := &Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service_Custom_PortRange{}
	// 	// 		custServiceInternal.PortRange = basetypes.NewObjectNull(mapAttributeTypes(ctx, custServiceInternalPortrange, resp))
	// 	// 	}

	// 	// 	//elementsServiceCustomInput = append(elementsServiceCustomInput, custServiceInternal)
	// 	// }

	// 	// serviceInput.Custom, diags = basetypes.NewListValueFrom(ctx, serviceInput.Custom.ElementType(ctx), elementsServiceCustomInput)
	// 	// resp.Diagnostics.Append(diags...)

	// 	ruleInput.Service, diags = types.ObjectValueFrom(ctx, ruleInput.Service.AttributeTypes(ctx), serviceInput)
	// 	resp.Diagnostics.Append(diags...)
	// }

	// // Rule -> Tracking
	// var trackingInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking
	// diags = ruleInput.Tracking.As(ctx, &trackingInput, basetypes.ObjectAsOptions{})
	// resp.Diagnostics.Append(diags...)

	// // Rule -> Tracking -> Event
	// trackingEventInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Event{}
	// diags = trackingInput.Event.As(ctx, &trackingEventInput, basetypes.ObjectAsOptions{})
	// resp.Diagnostics.Append(diags...)

	// trackingEventInput.Enabled = basetypes.NewBoolValue(currentRule.Tracking.Event.Enabled)
	// // trackingEventInput.Enabled = basetypes.NewBoolValue(currentRule.Tracking.Event.Enabled)
	// // trackingInput.Event, diags = types.ObjectValueFrom(ctx, trackingInput.Event.AttributeTypes(ctx), trackingEventInput)
	// // resp.Diagnostics.Append(diags...)

	// // Rule -> Tracking -> Alert
	// trackingAlertInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Alert{}
	// diags = trackingInput.Alert.As(ctx, &trackingAlertInput, basetypes.ObjectAsOptions{})
	// resp.Diagnostics.Append(diags...)

	// trackingAlertInput.Enabled = basetypes.NewBoolValue(currentRule.Tracking.Alert.Enabled)
	// trackingAlertInput.Frequency = basetypes.NewStringValue(currentRule.Tracking.Alert.Frequency.String())

	// // Rule -> Tracking -> Alert -> SubscriptionGroup
	// if len(currentRule.Tracking.Alert.SubscriptionGroup) > 0 {
	// 	trackingAlertInput.SubscriptionGroup, diags = types.ListValueFrom(ctx, trackingAlertInput.SubscriptionGroup.ElementType(ctx), parseNameIDList(ctx, currentRule.Tracking.Alert.SubscriptionGroup, resp))
	// 	resp.Diagnostics.Append(diags...)
	// }

	// // Rule -> Tracking -> Alert -> Webhook
	// if len(currentRule.Tracking.Alert.Webhook) > 0 {
	// 	trackingAlertInput.Webhook, diags = types.ListValueFrom(ctx, trackingAlertInput.Webhook.ElementType(ctx), parseNameIDList(ctx, currentRule.Tracking.Alert.Webhook, resp))
	// 	resp.Diagnostics.Append(diags...)
	// }

	// // Rule -> Tracking -> Alert -> MailingList
	// if len(currentRule.Tracking.Alert.MailingList) > 0 {
	// 	trackingAlertInput.MailingList, diags = types.ListValueFrom(ctx, trackingAlertInput.MailingList.ElementType(ctx), parseNameIDList(ctx, currentRule.Tracking.Alert.MailingList, resp))
	// 	resp.Diagnostics.Append(diags...)
	// }

	// trackingInput.Event, diags = types.ObjectValueFrom(ctx, trackingInput.Event.AttributeTypes(ctx), trackingEventInput)
	// resp.Diagnostics.Append(diags...)
	// trackingInput.Alert, diags = types.ObjectValueFrom(ctx, trackingInput.Alert.AttributeTypes(ctx), trackingAlertInput)
	// resp.Diagnostics.Append(diags...)

	// // Rule -> Tracking
	// ruleInput.Tracking, diags = types.ObjectValueFrom(ctx, ruleInput.Tracking.AttributeTypes(ctx), trackingInput)
	// resp.Diagnostics.Append(diags...)

	// // Rule -> Destination
	// ruleInput.Destination, diags = types.ObjectValueFrom(ctx, ruleInput.Destination.AttributeTypes(ctx), destInput)
	// resp.Diagnostics.Append(diags...)

	// // Rule -> Schedule
	// scheduleInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Schedule{}
	// diags = ruleInput.Schedule.As(ctx, &scheduleInput, basetypes.ObjectAsOptions{})
	// resp.Diagnostics.Append(diags...)

	// scheduleInput.ActiveOn = basetypes.NewStringValue(currentRule.Schedule.ActiveOn.String())

	// // Rule -> Schedule -> CustomTimeframe
	// if currentRule.Schedule.GetCustomTimeframePolicySchedule() != nil {
	// 	if currentRule.Schedule.GetCustomTimeframePolicySchedule().From != "" && currentRule.Schedule.GetCustomTimeframePolicySchedule().To != "" {
	// 		customeTimeFrameInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Schedule_CustomTimeframe{}
	// 		diags = scheduleInput.CustomTimeframe.As(ctx, &customeTimeFrameInput, basetypes.ObjectAsOptions{})
	// 		resp.Diagnostics.Append(diags...)
	// 		customeTimeFrameInput.From = basetypes.NewStringValue(currentRule.Schedule.CustomTimeframePolicySchedule.From)
	// 		customeTimeFrameInput.To = basetypes.NewStringValue(currentRule.Schedule.CustomTimeframePolicySchedule.To)
	// 		scheduleInput.CustomTimeframe, diags = types.ObjectValueFrom(ctx, scheduleInput.CustomTimeframe.AttributeTypes(ctx), customeTimeFrameInput)
	// 		resp.Diagnostics.Append(diags...)
	// 	}
	// }

	// // Rule -> Schedule -> CustomRecurring
	// if currentRule.Schedule.GetCustomRecurringPolicySchedule() != nil {
	// 	if currentRule.Schedule.GetCustomRecurringPolicySchedule().From != "" && currentRule.Schedule.GetCustomRecurringPolicySchedule().To != "" {
	// 		customRecurringInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Schedule_CustomRecurring{}
	// 		diags = scheduleInput.CustomRecurring.As(ctx, &customRecurringInput, basetypes.ObjectAsOptions{})
	// 		resp.Diagnostics.Append(diags...)
	// 		customRecurringInput.From = basetypes.NewStringValue(string(currentRule.Schedule.CustomRecurringPolicySchedule.From))
	// 		customRecurringInput.To = basetypes.NewStringValue(string(currentRule.Schedule.CustomRecurringPolicySchedule.To))
	// 		scheduleInput.CustomRecurring, diags = types.ObjectValueFrom(ctx, scheduleInput.CustomRecurring.AttributeTypes(ctx), customRecurringInput)
	// 		resp.Diagnostics.Append(diags...)
	// 	}
	// }

	// // Rule -> Schedule
	// ruleInput.Schedule, diags = types.ObjectValueFrom(ctx, ruleInput.Schedule.AttributeTypes(ctx), scheduleInput)
	// resp.Diagnostics.Append(diags...)

	// Rule -> Exceptions

	// if currentRule.Exceptions != nil && len(currentRule.Exceptions) > 0 {
	// 	elementsExceptionsInput := make([]types.Object, 0, len(currentRule.Exceptions))
	// 	diags = ruleInput.Exceptions.ElementsAs(ctx, &elementsExceptionsInput, false)
	// 	elementsExceptionsInputType := ruleInput.Exceptions.ElementType(ctx)
	// 	resp.Diagnostics.Append(diags...)

	// 	for _, item := range currentRule.Exceptions {
	// 		var itemExceptionsInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Exceptions
	// 		itemExceptionsInput.ConnectionOrigin = basetypes.NewStringValue(item.ConnectionOrigin.String())
	// 		itemExceptionsInput.
	// 	}
	// }

	// if currentRule.Exceptions != nil {
	// 	if len(currentRule.Exceptions) > 0 {
	// 		elementsExceptionsInput := make([]types.Object, len(currentRule.Exceptions))
	// 		diags = ruleInput.Exceptions.ElementsAs(ctx, &elementsExceptionsInput, false)
	// 		resp.Diagnostics.Append(diags...)

	// 		// elementsExceptionsInputType := ruleInput.Exceptions.ElementType(ctx)

	// 		tflog.Info(ctx, "currentRule.Exceptions", map[string]interface{}{
	// 			"currentRule.Exceptions":           currentRule.Exceptions,
	// 			"len(currentRule.Exceptions)":      len(currentRule.Exceptions),
	// 			"elementsExceptionsInput":          elementsExceptionsInput,
	// 			"cap(elementsExceptionsInput)":     cap(elementsExceptionsInput),
	// 			"len(elementsExceptionsInput)":     len(elementsExceptionsInput),
	// 			"len(currentRule.GetExceptions())": len(currentRule.GetExceptions()),
	// 		})

	// 		// for key, exceptItem := range currentRule.Exceptions {
	// 		// 	var exceptionInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Exceptions
	// 		// 	exceptionInput.Name = basetypes.NewStringValue(exceptItem.Name)

	// 		// 	if exceptItem.ConnectionOrigin.String() != "" {
	// 		// 		exceptionInput.ConnectionOrigin = basetypes.NewStringValue(exceptItem.ConnectionOrigin.String())
	// 		// 	} else {
	// 		// 		exceptionInput.ConnectionOrigin = basetypes.NewStringValue("ANY")
	// 		// 	}

	// 		// 	// PICKUP FROM HERE!!!!

	// 		// 	diags = elementsExceptionsInput[key].As(ctx, &exceptionInput, basetypes.ObjectAsOptions{})
	// 		// 	resp.Diagnostics.Append(diags...)
	// 		// }
	// 		var itemExceptionsInputType Policy_Policy_InternetFirewall_Policy_Rules_Rule_Exceptions

	// 		// diags = elementsExceptionsInput[0].As(ctx, &itemExceptionsInputType, basetypes.ObjectAsOptions{})
	// 		// resp.Diagnostics.Append(diags...)

	// 		for _, item := range currentRule.Exceptions {
	// 			var itemExceptionsInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Exceptions
	// 			itemExceptionsInput.Name = basetypes.NewStringValue((item.Name))
	// 			itemExceptionsInput.ConnectionOrigin = basetypes.NewStringValue(item.ConnectionOrigin.String())

	// 			// Rule -> Exceptions -> Source
	// 			itemExceptionsSourceInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source{}
	// 			diags = itemExceptionsInput.Source.As(ctx, &itemExceptionsSourceInput, basetypes.ObjectAsOptions{})
	// 			resp.Diagnostics.Append(diags...)
	// 			// itemExceptionsSourceInputType := itemExceptionsInput.Source.AttributeTypes(ctx)

	// 			// Rule -> Exceptions -> Source -> IP
	// 			if item.Source.IP != nil {
	// 				itemExceptionsSourceInput.IP, diags = basetypes.NewListValueFrom(ctx, types.StringType, item.Source.IP)
	// 				resp.Diagnostics.Append(diags...)
	// 			}

	// 			// Rule -> Exceptions -> Source -> Subnet
	// 			if item.Source.Subnet != nil {
	// 				itemExceptionsSourceInput.Subnet, diags = basetypes.NewListValueFrom(ctx, types.StringType, item.Source.Subnet)
	// 				resp.Diagnostics.Append(diags...)
	// 			}

	// 			// Rule -> Exceptions -> Source -> Host
	// 			if item.Source.Host != nil {
	// 				itemExceptionsSourceInput.Host, diags = basetypes.NewListValueFrom(ctx, itemExceptionsSourceInput.Host.ElementType(ctx), item.Source.Host)
	// 				resp.Diagnostics.Append(diags...)
	// 			}

	// 			// // Rule -> Exceptions -> Source -> Site
	// 			if item.Source.Site != nil {
	// 				itemExceptionsSourceInput.Site, diags = basetypes.NewListValueFrom(ctx, itemExceptionsSourceInput.Site.ElementType(ctx), parseNameIDList(ctx, item.Source.Site, resp))
	// 				resp.Diagnostics.Append(diags...)
	// 			}

	// 			// Rule -> Exceptions -> Source -> IPRange
	// 			if item.Source.IPRange != nil {
	// 				if item.Source.IPRange != nil {
	// 					itemExceptionsSourceInput.IPRange, diags = basetypes.NewListValueFrom(ctx, itemExceptionsSourceInput.IPRange.ElementType(ctx), parseFromToList(ctx, item.Source.IPRange, resp))
	// 					resp.Diagnostics.Append(diags...)
	// 				}
	// 			}

	// 			// Rule -> Exceptions -> Source -> GlobalIPRange
	// 			if item.Source.GlobalIPRange != nil {
	// 				if item.Source.GlobalIPRange != nil {
	// 					itemExceptionsSourceInput.GlobalIPRange, diags = basetypes.NewListValueFrom(ctx, itemExceptionsSourceInput.GlobalIPRange.ElementType(ctx), parseNameIDList(ctx, item.Source.GlobalIPRange, resp))
	// 					resp.Diagnostics.Append(diags...)
	// 				}
	// 			}

	// 			// Rule -> Exceptions -> Source -> NetworkInterface
	// 			if item.Source.NetworkInterface != nil {
	// 				if item.Source.NetworkInterface != nil {
	// 					itemExceptionsSourceInput.NetworkInterface, diags = basetypes.NewListValueFrom(ctx, itemExceptionsSourceInput.NetworkInterface.ElementType(ctx), parseNameIDList(ctx, item.Source.NetworkInterface, resp))
	// 					resp.Diagnostics.Append(diags...)
	// 				}
	// 			}

	// 			// Rule -> Exceptions -> Source -> SiteNetworkSubnet
	// 			if item.Source.SiteNetworkSubnet != nil {
	// 				if item.Source.SiteNetworkSubnet != nil {
	// 					itemExceptionsSourceInput.SiteNetworkSubnet, diags = basetypes.NewListValueFrom(ctx, itemExceptionsSourceInput.SiteNetworkSubnet.ElementType(ctx), parseNameIDList(ctx, item.Source.SiteNetworkSubnet, resp))
	// 					resp.Diagnostics.Append(diags...)
	// 				}
	// 			}

	// 			// Rule -> Exceptions -> Source -> FloatingSubnet
	// 			if item.Source.FloatingSubnet != nil {
	// 				if item.Source.FloatingSubnet != nil {
	// 					itemExceptionsSourceInput.FloatingSubnet, diags = basetypes.NewListValueFrom(ctx, itemExceptionsSourceInput.FloatingSubnet.ElementType(ctx), parseNameIDList(ctx, item.Source.FloatingSubnet, resp))
	// 					resp.Diagnostics.Append(diags...)
	// 				}
	// 			}

	// 			// Rule -> Exceptions -> Source -> User
	// 			if item.Source.User != nil {
	// 				if item.Source.User != nil {
	// 					itemExceptionsSourceInput.User, diags = basetypes.NewListValueFrom(ctx, itemExceptionsSourceInput.User.ElementType(ctx), parseNameIDList(ctx, item.Source.User, resp))
	// 					resp.Diagnostics.Append(diags...)
	// 				}
	// 			}

	// 			// Rule -> Exceptions -> Source -> UsersGroup
	// 			if item.Source.UsersGroup != nil {
	// 				if item.Source.UsersGroup != nil {
	// 					itemExceptionsSourceInput.UsersGroup, diags = basetypes.NewListValueFrom(ctx, itemExceptionsSourceInput.UsersGroup.ElementType(ctx), parseNameIDList(ctx, item.Source.UsersGroup, resp))
	// 					resp.Diagnostics.Append(diags...)
	// 				}
	// 			}

	// 			// Rule -> Exceptions -> Source -> Group
	// 			if item.Source.Group != nil {
	// 				if item.Source.Group != nil {
	// 					itemExceptionsSourceInput.Group, diags = basetypes.NewListValueFrom(ctx, itemExceptionsSourceInput.Group.ElementType(ctx), parseNameIDList(ctx, item.Source.Group, resp))
	// 					resp.Diagnostics.Append(diags...)
	// 				}
	// 			}

	// 			// Rule -> Exceptions -> Source -> SystemGroup
	// 			if item.Source.SystemGroup != nil {
	// 				if item.Source.SystemGroup != nil {
	// 					itemExceptionsSourceInput.SystemGroup, diags = basetypes.NewListValueFrom(ctx, itemExceptionsSourceInput.SystemGroup.ElementType(ctx), parseNameIDList(ctx, item.Source.SystemGroup, resp))
	// 					resp.Diagnostics.Append(diags...)
	// 				}
	// 			}

	// 			// Rule -> Exceptions -> Country
	// 			if item.Country != nil {
	// 				if len(item.Country) > 0 {
	// 					itemExceptionsInput.Country, diags = basetypes.NewListValueFrom(ctx, itemExceptionsInput.Country.ElementType(ctx), parseNameIDList(ctx, item.Country, resp))
	// 					resp.Diagnostics.Append(diags...)
	// 				}
	// 			}

	// 			// Rule -> Exceptions -> Device
	// 			if item.Device != nil {
	// 				if len(item.Device) > 0 {
	// 					itemExceptionsInput.Device, diags = basetypes.NewListValueFrom(ctx, itemExceptionsInput.Device.ElementType(ctx), parseNameIDList(ctx, item.Device, resp))
	// 					resp.Diagnostics.Append(diags...)
	// 				}
	// 			}

	// 			// Rule -> Exceptions -> DeviceOS
	// 			if item.DeviceOs != nil {
	// 				if len(item.DeviceOs) > 0 {
	// 					var strtmp []string
	// 					for _, strtmpVal := range item.DeviceOs {
	// 						strtmp = append(strtmp, strtmpVal.String())
	// 					}
	// 					itemExceptionsInput.DeviceOs, diags = basetypes.NewListValueFrom(ctx, itemExceptionsInput.DeviceOs.ElementType(ctx), strtmp)
	// 					resp.Diagnostics.Append(diags...)
	// 				}
	// 			}

	// 			// Rule -> Exceptions -> Destination
	// 			itemExceptionsDestinationInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination{}
	// 			diags = itemExceptionsInput.Destination.As(ctx, &itemExceptionsDestinationInput, basetypes.ObjectAsOptions{})
	// 			resp.Diagnostics.Append(diags...)

	// 			// Rule -> Exceptions -> Destination -> IP
	// 			if item.Destination.IP != nil {
	// 				if len(item.Destination.IP) > 0 {
	// 					itemExceptionsDestinationInput.IP, diags = basetypes.NewListValueFrom(ctx, itemExceptionsDestinationInput.IP.ElementType((ctx)), item.Destination.IP)
	// 					resp.Diagnostics.Append(diags...)
	// 				}
	// 			}

	// 			// Rule -> Exceptions -> Destination -> Subnet
	// 			if item.Destination.Subnet != nil {
	// 				if len(item.Destination.Subnet) > 0 {
	// 					itemExceptionsDestinationInput.Subnet, diags = basetypes.NewListValueFrom(ctx, itemExceptionsDestinationInput.Subnet.ElementType((ctx)), item.Destination.Subnet)
	// 					resp.Diagnostics.Append(diags...)
	// 				}
	// 			}

	// 			// Rule -> Exceptions -> Destination -> Domain
	// 			if item.Destination.Domain != nil {
	// 				if len(item.Destination.Domain) > 0 {
	// 					itemExceptionsDestinationInput.Domain, diags = basetypes.NewListValueFrom(ctx, itemExceptionsDestinationInput.Domain.ElementType((ctx)), item.Destination.Domain)
	// 					resp.Diagnostics.Append(diags...)
	// 				}
	// 			}

	// 			// Rule -> Exceptions -> Destination -> Fqdn
	// 			if item.Destination.Fqdn != nil {
	// 				if len(item.Destination.Fqdn) > 0 {
	// 					itemExceptionsDestinationInput.Fqdn, diags = basetypes.NewListValueFrom(ctx, itemExceptionsDestinationInput.Fqdn.ElementType((ctx)), item.Destination.Fqdn)
	// 					resp.Diagnostics.Append(diags...)
	// 				}
	// 			}

	// 			// Rule -> Exceptions -> Destination -> RemoteAsn
	// 			if item.Destination.RemoteAsn != nil {
	// 				if len(item.Destination.RemoteAsn) > 0 {
	// 					var strtmp []string
	// 					for _, strtmpVal := range item.Destination.RemoteAsn {
	// 						strtmp = append(strtmp, string(strtmpVal))
	// 					}
	// 					itemExceptionsDestinationInput.RemoteAsn, diags = basetypes.NewListValueFrom(ctx, itemExceptionsDestinationInput.RemoteAsn.ElementType((ctx)), item.Destination.RemoteAsn)
	// 					resp.Diagnostics.Append(diags...)
	// 				}
	// 			}

	// 			// Rule -> Exceptions -> Destination -> Application
	// 			if item.Destination.Application != nil {
	// 				if len(item.Destination.Application) > 0 {
	// 					itemExceptionsDestinationInput.Application, diags = basetypes.NewListValueFrom(ctx, itemExceptionsDestinationInput.Application.ElementType((ctx)), parseNameIDList(ctx, item.Destination.Application, resp))
	// 					resp.Diagnostics.Append(diags...)
	// 				}
	// 			}

	// 			// Rule -> Exceptions -> Destination -> CustomApp
	// 			if item.Destination.CustomApp != nil {
	// 				if len(item.Destination.CustomApp) > 0 {
	// 					itemExceptionsDestinationInput.CustomApp, diags = basetypes.NewListValueFrom(ctx, itemExceptionsDestinationInput.CustomApp.ElementType((ctx)), parseNameIDList(ctx, item.Destination.CustomApp, resp))
	// 					resp.Diagnostics.Append(diags...)
	// 				}
	// 			}

	// 			// Rule -> Exceptions -> Destination -> IPRange
	// 			if item.Destination.IPRange != nil {
	// 				if len(item.Destination.IPRange) > 0 {
	// 					itemExceptionsDestinationInput.IPRange, diags = basetypes.NewListValueFrom(ctx, itemExceptionsDestinationInput.IPRange.ElementType((ctx)), parseFromToList(ctx, item.Destination.IPRange, resp))
	// 					resp.Diagnostics.Append(diags...)
	// 				}
	// 			}

	// 			// Rule -> Exceptions -> Destination -> GlobalIPRange
	// 			if item.Destination.GlobalIPRange != nil {
	// 				if len(item.Destination.GlobalIPRange) > 0 {
	// 					itemExceptionsDestinationInput.GlobalIPRange, diags = basetypes.NewListValueFrom(ctx, itemExceptionsDestinationInput.GlobalIPRange.ElementType((ctx)), parseNameIDList(ctx, item.Destination.GlobalIPRange, resp))
	// 					resp.Diagnostics.Append(diags...)
	// 				}
	// 			}

	// 			// Rule -> Exceptions -> Destination -> AppCategory
	// 			if item.Destination.AppCategory != nil {
	// 				if len(item.Destination.AppCategory) > 0 {
	// 					itemExceptionsDestinationInput.AppCategory, diags = basetypes.NewListValueFrom(ctx, itemExceptionsDestinationInput.AppCategory.ElementType((ctx)), parseNameIDList(ctx, item.Destination.AppCategory, resp))
	// 					resp.Diagnostics.Append(diags...)
	// 				}
	// 			}

	// 			// Rule -> Exceptions -> Destination -> CustomCategory
	// 			if item.Destination.CustomCategory != nil {
	// 				if len(item.Destination.CustomCategory) > 0 {
	// 					itemExceptionsDestinationInput.CustomCategory, diags = basetypes.NewListValueFrom(ctx, itemExceptionsDestinationInput.CustomCategory.ElementType((ctx)), parseNameIDList(ctx, item.Destination.CustomCategory, resp))
	// 					resp.Diagnostics.Append(diags...)
	// 				}
	// 			}

	// 			// Rule -> Exceptions -> Destination -> SanctionedAppsCategory
	// 			if item.Destination.SanctionedAppsCategory != nil {
	// 				if len(item.Destination.SanctionedAppsCategory) > 0 {
	// 					itemExceptionsDestinationInput.SanctionedAppsCategory, diags = basetypes.NewListValueFrom(ctx, itemExceptionsDestinationInput.SanctionedAppsCategory.ElementType((ctx)), parseNameIDList(ctx, item.Destination.SanctionedAppsCategory, resp))
	// 					resp.Diagnostics.Append(diags...)
	// 				}
	// 			}

	// 			// Rule -> Exceptions -> Destination -> Country
	// 			if item.Destination.Country != nil {
	// 				if len(item.Destination.Country) > 0 {
	// 					itemExceptionsDestinationInput.Country, diags = basetypes.NewListValueFrom(ctx, itemExceptionsDestinationInput.Country.ElementType((ctx)), parseNameIDList(ctx, item.Destination.Country, resp))
	// 					resp.Diagnostics.Append(diags...)
	// 				}
	// 			}

	// 			// Rule -> Exceptions -> Service
	// 			serviceInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service{}
	// 			diags = itemExceptionsInput.Service.As(ctx, &serviceInput, basetypes.ObjectAsOptions{})
	// 			resp.Diagnostics.Append(diags...)

	// 			// Rule -> Exceptions -> Service -> Standard
	// 			if item.Service.Standard != nil {
	// 				if len(item.Service.Standard) > 0 {
	// 					serviceInput.Standard, diags = basetypes.NewListValueFrom(ctx, serviceInput.Standard.ElementType((ctx)), parseNameIDList(ctx, item.Service.Standard, resp))
	// 					resp.Diagnostics.Append(diags...)
	// 				}
	// 			}

	// 			// Rule -> Exceptions -> Service -> Custom
	// 			if item.Service.Custom != nil {
	// 				if len(item.Service.Custom) > 0 {

	// 					elementsExceptionsServiceCustomInput := make([]types.Object, 0, len(item.Service.Custom))
	// 					diags = serviceInput.Custom.ElementsAs(ctx, &elementsExceptionsServiceCustomInput, false)
	// 					resp.Diagnostics.Append(diags...)

	// 					var itemServiceCustomInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service_Custom
	// 					diags = elementsExceptionsServiceCustomInput[0].As(ctx, &itemServiceCustomInput, basetypes.ObjectAsOptions{})
	// 					resp.Diagnostics.Append(diags...)
	// 					for _, elementsServiceCustomInput := range item.Service.Custom {

	// 						// Rule -> Exceptions -> Service -> Custom -> Port
	// 						if len(elementsServiceCustomInput.Port) > 0 {
	// 							var elementsServiceCustomPortInput []attr.Value
	// 							for _, v := range elementsServiceCustomInput.Port {
	// 								elementsServiceCustomPortInput = append(elementsServiceCustomPortInput, basetypes.NewStringValue(string(v)))
	// 							}
	// 							itemServiceCustomInput.Port, diags = basetypes.NewListValue(types.StringType, elementsServiceCustomPortInput)
	// 							resp.Diagnostics.Append(diags...)
	// 						}
	// 						// Rule -> Exceptions -> Service -> Custom -> PortRange
	// 						if elementsServiceCustomInput.PortRangeCustomService != nil {
	// 							itemServiceCustomInput.PortRange, diags = basetypes.NewObjectValueFrom(ctx, itemServiceCustomInput.PortRange.AttributeTypes(ctx), elementsServiceCustomInput.PortRangeCustomService)
	// 							resp.Diagnostics.Append(diags...)
	// 						}

	// 						itemServiceCustomInput.Protocol = basetypes.NewStringValue(string(*elementsServiceCustomInput.GetProtocol()))

	// 						resp.Diagnostics.Append(diags...)
	// 					}

	// 					serviceInput.Custom, diags = basetypes.NewListValueFrom(ctx, serviceInput.Custom.ElementType((ctx)), elementsExceptionsServiceCustomInput)
	// 					resp.Diagnostics.Append(diags...)
	// 				}
	// 			}
	// 		}
	// 	}
	// }

	diags = resp.State.SetAttribute(ctx, path.Root("rule"), ruleInput)
	resp.Diagnostics.Append(diags...)

}

var FromToAttrTypes = map[string]attr.Type{
	"from": types.StringType,
	"to":   types.StringType,
}

var FromToObjectType = types.ObjectType{AttrTypes: FromToAttrTypes}

var NameIDAttrTypes = map[string]attr.Type{
	"name": types.StringType,
	"id":   types.StringType,
}

// ObjectType wrapper for ListValue
var NameIDObjectType = types.ObjectType{AttrTypes: NameIDAttrTypes}

func mapObjectList(ctx context.Context, srcItemObjList any, resp *resource.ReadResponse) []types.Object {
	vals := reflect.ValueOf(srcItemObjList)
	var objList []types.Object
	for i := range vals.Len() {
		objList = append(objList, mapStructFields(ctx, vals.Index(i), resp))
	}

	return objList
}

func mapStructFields(ctx context.Context, srcItemObj any, resp *resource.ReadResponse) types.Object {
	vals := reflect.ValueOf(srcItemObj)
	names := structs.Names(srcItemObj)
	tflog.Info(ctx, "pointer val: ",
		map[string]interface{}{
			"pointer_val": vals,
		})
	attrTypes := map[string]attr.Type{}
	attrValues := map[string]attr.Value{}

	// for i := range vals.NumField() {
	// 	attrTypes[vals.Type().Field(i).Name] = types.StringType
	// 	attrValues[vals.Type().Field(i).Name] = basetypes.NewStringValue(vals.Field(i).String())
	// }

	for _, v := range names {
		attrTypes[v] = types.StringType
		attrValues[v] = basetypes.NewStringValue(vals.FieldByName(v).String())
	}

	newObj, diags := basetypes.NewObjectValueFrom(ctx, attrTypes, attrValues)
	resp.Diagnostics.Append(diags...)

	return newObj
}

func mapAttributeTypes(ctx context.Context, srcItemObj any, resp *resource.ReadResponse) map[string]attr.Type {
	attrTypes := map[string]attr.Type{}

	names := structs.Names(srcItemObj)
	vals := structs.Map(srcItemObj)
	for _, v := range names {
		intV := stringy.New(v).SnakeCase().ToLower()

		attrTypes[strings.ToLower(intV)] = convertGoTypeToTfType(ctx, vals[v])
	}

	return attrTypes
}

func convertGoTypeToTfType(ctx context.Context, srcItemObj any) attr.Type {
	srcItemObjType := reflect.TypeOf(srcItemObj).String()
	tflog.Error(ctx, "srcItemObjType", map[string]interface{}{
		"srcItemObjType": srcItemObjType,
	})
	switch srcItemObjType {
	case "string":
		return types.StringType
	case "basetypes.StringValue":
		return types.StringType
	case "bool":
		return types.BoolType
	case "basetypes.BoolValue":
		return types.BoolType
	case "int":
		return types.Int32Type
	case "int64":
		return types.Int64Type
	case "float":
		return types.Float32Type
	case "float64":
		return types.Float64Type
	case "basetypes.ObjectValue":
		return types.ObjectType{}
	}

	return types.StringType
}

// func mapAttributeValues(ctx context.Context, srcItemObj any, resp *resource.ReadResponse) map[string]attr.Value {
// 	attrTypes := map[string]attr.Value{}

// 	names := structs.Names(srcItemObj)
// 	rt := reflect.TypeOf(srcItemObj)
// 	for _, v := range names {
// 		fv, _ := rt.FieldByName(v)
// 		attrTypes[v] = basetypes.NewStringValue(fv.String())
// 	}

// 	return attrTypes
// }

func parseNameID(ctx context.Context, item interface{}) types.Object {
	tflog.Warn(ctx, "parseNameID() - "+fmt.Sprintf("%v", item))
	diags := make(diag.Diagnostics, 0)

	// Get the reflect.Value of the input
	itemValue := reflect.ValueOf(item)

	// Handle nil or invalid input (must be a struct, not a slice/array)
	if item == nil || itemValue.Kind() != reflect.Struct {
		if itemValue.Kind() == reflect.Ptr && !itemValue.IsNil() {
			itemValue = itemValue.Elem()
			if itemValue.Kind() != reflect.Struct {
				return types.ObjectNull(NameIDAttrTypes)
			}
		} else {
			return types.ObjectNull(NameIDAttrTypes)
		}
	}

	// Handle pointer to struct
	if itemValue.Kind() == reflect.Ptr {
		if itemValue.IsNil() {
			return types.ObjectNull(NameIDAttrTypes)
		}
		itemValue = itemValue.Elem()
	}

	// Get Name and ID fields
	nameField := itemValue.FieldByName("Name")
	idField := itemValue.FieldByName("ID")

	if !nameField.IsValid() || !idField.IsValid() {
		return types.ObjectNull(NameIDAttrTypes)
	}

	// Create object value
	obj, diagstmp := types.ObjectValue(
		NameIDAttrTypes,
		map[string]attr.Value{
			"name": basetypes.NewStringValue(nameField.String()),
			"id":   basetypes.NewStringValue(idField.String()),
		},
	)
	diags = append(diags, diagstmp...)
	return obj
}

// func parseNameIDList2(ctx context.Context, items interface{}, resp *resource.ReadResponse) types.List {
// 	tflog.Warn(ctx, "parseNameIDList() - "+fmt.Sprintf("%v", items))
// 	diags := make(diag.Diagnostics, 0)
// 	// Get the reflect.Value of the input
// 	itemsValue := reflect.ValueOf(items)

// 	// Handle nil or empty input
// 	rt := reflect.TypeOf(items)
// 	if items == nil || (rt.Kind() != reflect.Array && rt.Kind() != reflect.Slice) {
// 		return types.ListNull(NameIDObjectType)
// 	} else {
// 		if itemsValue.Len() == 0 {
// 			return types.ListNull(NameIDObjectType)
// 		}
// 	}

// 	values := make([]attr.Value, itemsValue.Len())
// 	for i := range itemsValue.Len() {
// 		item := itemsValue.Index(i)

// 		// Handle pointer elements
// 		if item.Kind() == reflect.Ptr {
// 			item = item.Elem()
// 		}

// 		// Get Name and ID fields
// 		nameField := item.FieldByName("Name")
// 		idField := item.FieldByName("ID")

// 		if !nameField.IsValid() || !idField.IsValid() {
// 			return types.ListNull(NameIDObjectType)
// 		}

// 		// Create object value
// 		obj, diagstmp := types.ObjectValue(
// 			NameIDAttrTypes,
// 			map[string]attr.Value{
// 				"name": basetypes.NewStringValue(nameField.String()),
// 				"id":   basetypes.NewStringValue(idField.String()),
// 			},
// 		)

// 		diags = append(diags, diagstmp...)
// 		values[i] = obj
// 	}

// 	// Convert to List
// 	list, diagstmp := types.ListValue(NameIDObjectType, values)

// 	diags = append(diags, diagstmp...)
// 	resp.Diagnostics.Append(diags...)

// 	return list
// }

func parseFromToList(ctx context.Context, items interface{}, resp *resource.ReadResponse) types.List {
	tflog.Warn(ctx, "parseFromToList() - "+fmt.Sprintf("%v", items))
	diags := make(diag.Diagnostics, 0)
	// Get the reflect.Value of the input
	itemsValue := reflect.ValueOf(items)

	// Handle nil or empty input
	rt := reflect.TypeOf(items)
	if items == nil || (rt.Kind() != reflect.Array && rt.Kind() != reflect.Slice) {
		return types.ListNull(FromToObjectType)
	} else {
		if itemsValue.Len() == 0 {
			return types.ListNull(FromToObjectType)
		}
	}

	values := make([]attr.Value, itemsValue.Len())
	for i := range itemsValue.Len() {
		item := itemsValue.Index(i)

		// Handle pointer elements
		if item.Kind() == reflect.Ptr {
			item = item.Elem()
		}

		// Get Name and ID fields
		fromField := item.FieldByName("From")
		toField := item.FieldByName("To")

		if !fromField.IsValid() || !toField.IsValid() {
			return types.ListNull(FromToObjectType)
		}

		// Create object value
		obj, diagstmp := types.ObjectValue(
			FromToAttrTypes,
			map[string]attr.Value{
				"from": basetypes.NewStringValue(fromField.String()),
				"to":   basetypes.NewStringValue(toField.String()),
			},
		)

		diags = append(diags, diagstmp...)
		values[i] = obj
	}

	// Convert to List
	list, diagstmp := types.ListValue(FromToObjectType, values)

	diags = append(diags, diagstmp...)
	resp.Diagnostics.Append(diags...)

	return list
}
