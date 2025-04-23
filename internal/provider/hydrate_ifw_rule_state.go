package provider

import (
	"context"
	"fmt"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk" // Import the correct package
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func hydrateIfwRuleState(ctx context.Context, state InternetFirewallRule, currentRule *cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_Rule) Policy_Policy_InternetFirewall_Policy_Rules_Rule {
	ruleInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule{}
	diags := make(diag.Diagnostics, 0)
	diagstmp := state.Rule.As(ctx, &ruleInput, basetypes.ObjectAsOptions{})
	diags = append(diags, diagstmp...)
	if diags.HasError() {
		return ruleInput
	}
	ruleInput.Name = types.StringValue(currentRule.Name)
	if currentRule.Description == "" {
		ruleInput.Description = types.StringNull()
	} else {
		ruleInput.Description = types.StringValue(currentRule.Description)
	}
	ruleInput.Action = types.StringValue(currentRule.Action.String())
	ruleInput.ID = types.StringValue(currentRule.ID)
	// ruleInput.Index = types.StringValue(currentRule.Index.String())
	ruleInput.ConnectionOrigin = types.StringValue(currentRule.ConnectionOrigin.String())

	// //////////// Rule -> Source ///////////////
	curRuleSourceObj, diagstmp := types.ObjectValue(
		IfwSourceAttrTypes,
		map[string]attr.Value{
			"ip":                  parseList(ctx, types.StringType, currentRule.Source.IP, "rule.source.ip"),
			"host":                parseNameIDList(ctx, currentRule.Source.Host, "rule.source.host"),
			"site":                parseNameIDList(ctx, currentRule.Source.Site, "rule.source.site"),
			"subnet":              parseList(ctx, types.StringType, currentRule.Source.Subnet, "rule.source.subnet"),
			"ip_range":            parseFromToList(ctx, currentRule.Source.IPRange, "rule.source.ip_range"),
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
	ruleInput.Source = curRuleSourceObj
	////////////// end rule.source ///////////////

	// Rule -> Country
	ruleInput.Country = parseNameIDList(ctx, currentRule.Country, "rule.country")
	// Rule -> Device
	ruleInput.Device = parseNameIDList(ctx, currentRule.Device, "rule.device")

	// Rule -> DeviceOS
	ruleInput.DeviceOs = parseList(ctx, types.StringType, currentRule.DeviceOs, "rule.source.device_os")

	//////////// Rule -> Destination ///////////////
	curRuleDestinationObj, diagstmp := types.ObjectValue(
		IfwDestAttrTypes,
		map[string]attr.Value{
			"application":              parseNameIDList(ctx, currentRule.Destination.Application, "rule.destination.application"),
			"custom_app":               parseNameIDList(ctx, currentRule.Destination.CustomApp, "rule.destination.custom_app"),
			"app_category":             parseNameIDList(ctx, currentRule.Destination.AppCategory, "rule.destination.app_category"),
			"custom_category":          parseNameIDList(ctx, currentRule.Destination.CustomCategory, "rule.destination.custom_category"),
			"sanctioned_apps_category": parseNameIDList(ctx, currentRule.Destination.SanctionedAppsCategory, "rule.destination.sanctioned_apps_category"),
			"country":                  parseNameIDList(ctx, currentRule.Destination.Country, "rule.destination.country"),
			"domain":                   parseList(ctx, types.StringType, currentRule.Destination.Domain, "rule.destination.domain"),
			"fqdn":                     parseList(ctx, types.StringType, currentRule.Destination.Fqdn, "rule.destination.fqdn"),
			"ip":                       parseList(ctx, types.StringType, currentRule.Destination.IP, "rule.destination.ip"),
			"subnet":                   parseList(ctx, types.StringType, currentRule.Destination.Subnet, "rule.destination.subnet"),
			"ip_range":                 parseFromToList(ctx, currentRule.Destination.IPRange, "rule.destination.ip_range"),
			"global_ip_range":          parseNameIDList(ctx, currentRule.Destination.GlobalIPRange, "rule.destination.global_ip_range"),
			"remote_asn":               parseList(ctx, types.StringType, currentRule.Destination.RemoteAsn, "rule.destination.remote_asn"),
		},
	)
	diags = append(diags, diagstmp...)
	ruleInput.Destination = curRuleDestinationObj
	////////////// end Rule -> Destination ///////////////

	////////////// start Rule -> Service ///////////////
	if len(currentRule.Service.Custom) > 0 || len(currentRule.Service.Standard) > 0 {
		// Initialize Service object with null values
		curRuleServiceObj, diagstmp := types.ObjectValue(
			IfwServiceAttrTypes,
			map[string]attr.Value{
				"standard": types.SetNull(NameIDObjectType),
				"custom":   types.ListNull(CustomServiceObjectType),
			},
		)
		diags = append(diags, diagstmp...)
		curRuleServiceObjAttrs := curRuleServiceObj.Attributes()

		// Rule -> Service -> Standard
		curRuleServiceObjAttrs["standard"] = parseNameIDList(ctx, currentRule.Service.Standard, "rule.service.standard")

		// Rule -> Service -> Custom
		if len(currentRule.Service.Custom) > 0 {
			var curRuleCustomServices []types.Object
			tflog.Info(ctx, "ruleResponse.Service.Custom - "+fmt.Sprintf("%v", currentRule.Service.Custom))
			for _, item := range currentRule.Service.Custom {
				curRuleCustomServices = append(curRuleCustomServices, parseCustomService(ctx, item, "rule.service.custom"))
			}
			curRuleServiceObjAttrs["custom"], diagstmp = types.ListValueFrom(ctx, CustomServiceObjectType, curRuleCustomServices)
			diags = append(diags, diagstmp...)
		}

		curRuleServiceObj, diagstmp = types.ObjectValue(curRuleServiceObj.AttributeTypes(ctx), curRuleServiceObjAttrs)
		diags = append(diags, diagstmp...)
		ruleInput.Service = curRuleServiceObj
	}
	////////////// end Rule -> Service ///////////////

	////////////// start Rule -> Tracking ///////////////
	curRuleTrackingObj, diagstmp := types.ObjectValue(
		TrackingAttrTypes,
		map[string]attr.Value{
			"event": types.ObjectNull(TrackingEventAttrTypes),
			"alert": types.ObjectNull(TrackingAlertAttrTypes),
		},
	)
	diags = append(diags, diagstmp...)
	curRuleTrackingObjAttrs := curRuleTrackingObj.Attributes()

	// Rule -> Tracking -> Event
	trackingEventValue, diags := types.ObjectValue(
		TrackingEventAttrTypes,
		map[string]attr.Value{
			"enabled": types.BoolValue(currentRule.Tracking.Event.Enabled),
		},
	)
	curRuleTrackingObjAttrs["event"] = trackingEventValue

	// Rule -> Tracking -> Alert
	trackingAlertValue, diagstmp := types.ObjectValue(
		TrackingAlertAttrTypes,
		map[string]attr.Value{
			"enabled":            types.BoolValue(currentRule.Tracking.Alert.Enabled),
			"frequency":          types.StringValue(currentRule.Tracking.Alert.Frequency.String()),
			"subscription_group": parseNameIDList(ctx, currentRule.Tracking.Alert.SubscriptionGroup, "rule.tracking.alert.subscription_group"),
			"webhook":            parseNameIDList(ctx, currentRule.Tracking.Alert.Webhook, "rule.tracking.alert.webhook"),
			"mailing_list":       parseNameIDList(ctx, currentRule.Tracking.Alert.MailingList, "rule.tracking.alert.mailing_list"),
		},
	)
	diags = append(diags, diagstmp...)
	curRuleTrackingObjAttrs["alert"] = trackingAlertValue
	tflog.Warn(ctx, "Updated tracking object: "+fmt.Sprintf("%v", curRuleTrackingObj))

	curRuleTrackingObj, diagstmp = types.ObjectValue(curRuleTrackingObj.AttributeTypes(ctx), curRuleTrackingObjAttrs)
	diags = append(diags, diagstmp...)
	ruleInput.Tracking = curRuleTrackingObj
	////////////// end Rule -> Tracking ///////////////

	////////////// start Rule -> Schedule ///////////////
	curRuleScheduleObj, diagstmp := types.ObjectValue(
		ScheduleAttrTypes,
		map[string]attr.Value{
			"active_on":        types.StringValue(currentRule.Schedule.ActiveOn.String()),
			"custom_timeframe": parseFromTo(ctx, currentRule.Schedule.CustomTimeframePolicySchedule, "rule.schedule.custom_timeframe"),
			"custom_recurring": parseFromToDays(ctx, currentRule.Schedule.CustomRecurringPolicySchedule, "rule.schedule.custom_recurring"),
		},
	)
	diags = append(diags, diagstmp...)
	ruleInput.Schedule = curRuleScheduleObj
	////////////// end Rule -> Schedule ///////////////

	// ////////////// start Rule -> Exceptions ///////////////
	// TODO set to set not list
	curRuleExceptionsObj, diagstmp := types.SetValue(
		types.ObjectType{AttrTypes: IfwExceptionAttrTypes},
		[]attr.Value{
			types.ObjectValueMust( // Single exception object with all null values
				IfwExceptionAttrTypes,
				map[string]attr.Value{
					"name":    types.StringNull(),
					"source":  types.ObjectNull(IfwSourceAttrTypes),
					"country": types.SetNull(types.ObjectType{AttrTypes: NameIDAttrTypes}),
					"device":  types.SetNull(types.ObjectType{AttrTypes: NameIDAttrTypes}),
					// "device_attributes": types.ObjectNull(DeviceAttrAttrTypes),
					"device_os":         types.ListNull(types.StringType),
					"destination":       types.ObjectNull(IfwDestAttrTypes),
					"service":           types.ObjectNull(IfwServiceAttrTypes),
					"connection_origin": types.StringNull(),
				},
			),
		},
	)
	diags = append(diags, diagstmp...)
	exceptions := []attr.Value{}

	// Rule -> Exceptions -> Source
	tflog.Warn(ctx, "hydrateIFRuleState() currentRule.Exceptions - "+fmt.Sprintf("%v", currentRule.Exceptions))
	if currentRule.Exceptions != nil && len(currentRule.Exceptions) > 0 {
		for _, ruleException := range currentRule.Exceptions {
			// Rule -> Exceptions -> Source
			curExceptionSourceObj, diags := types.ObjectValue(
				IfwSourceAttrTypes,
				map[string]attr.Value{
					"ip":                  parseList(ctx, types.StringType, ruleException.Source.IP, "rule.exception.source.ip"),
					"subnet":              parseList(ctx, types.StringType, ruleException.Source.Subnet, "rule.exception.source.subnet"),
					"host":                parseNameIDList(ctx, ruleException.Source.Host, "rule.exception.source.host"),
					"site":                parseNameIDList(ctx, ruleException.Source.Site, "rule.exception.source.site"),
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

			// Rule -> Exceptions -> Destination
			curExceptionDestObj, diags := types.ObjectValue(
				IfwDestAttrTypes,
				map[string]attr.Value{
					"application":              parseNameIDList(ctx, ruleException.Destination.Application, "rule.exception.destination.application"),
					"custom_app":               parseNameIDList(ctx, ruleException.Destination.CustomApp, "rule.exception.destination.custom_app"),
					"app_category":             parseNameIDList(ctx, ruleException.Destination.AppCategory, "rule.exception.destination.app_category"),
					"custom_category":          parseNameIDList(ctx, ruleException.Destination.CustomCategory, "rule.exception.destination.custom_category"),
					"sanctioned_apps_category": parseNameIDList(ctx, ruleException.Destination.SanctionedAppsCategory, "rule.exception.destination.sanctioned_apps_category"),
					"country":                  parseNameIDList(ctx, ruleException.Destination.Country, "rule.exception.destination.country"),
					"domain":                   parseList(ctx, types.StringType, ruleException.Destination.Domain, "rule.exception.destination.domain"),
					"fqdn":                     parseList(ctx, types.StringType, ruleException.Destination.Fqdn, "rule.exception.destination.fqdn"),
					"ip":                       parseList(ctx, types.StringType, ruleException.Destination.IP, "rule.exception.destination.ip"),
					"subnet":                   parseList(ctx, types.StringType, ruleException.Destination.Subnet, "rule.exception.destination.subnet"),
					"ip_range":                 parseFromToList(ctx, ruleException.Destination.IPRange, "rule.exception.destination.ip_range"),
					"global_ip_range":          parseNameIDList(ctx, ruleException.Destination.GlobalIPRange, "rule.exception.destination.global_ip_range"),
					"remote_asn":               parseList(ctx, types.StringType, ruleException.Destination.RemoteAsn, "rule.exception.destination.remote_asn"),
				},
			)

			////////////// start Rule -> Service ///////////////
			// Initialize Service object with null values
			curExceptionServiceObj, diagstmp := types.ObjectValue(
				IfwServiceAttrTypes,
				map[string]attr.Value{
					"standard": types.SetNull(NameIDObjectType),
					"custom":   types.ListNull(CustomServiceObjectType),
				},
			)
			diags = append(diags, diagstmp...)
			curExceptionServiceObjAttrs := curExceptionServiceObj.Attributes()
			if len(ruleException.Service.Custom) > 0 || len(ruleException.Service.Standard) > 0 {
				// Rule -> Service -> Standard
				curExceptionServiceObjAttrs["standard"] = parseNameIDList(ctx, ruleException.Service.Standard, "rule.exception.service.standard")

				// Rule -> Service -> Custom
				if ruleException.Service.Custom != nil {
					if len(ruleException.Service.Custom) > 0 {
						var curExceptionCustomServices []types.Object
						tflog.Info(ctx, "ruleException.Service.Custom - "+fmt.Sprintf("%v", ruleException.Service.Custom))
						for _, item := range ruleException.Service.Custom {
							curExceptionCustomServices = append(curExceptionCustomServices, parseCustomService(ctx, item, "rule.exception.service.custom"))
						}
						curExceptionServiceObjAttrs["custom"], diagstmp = types.ListValueFrom(ctx, CustomServiceObjectType, curExceptionCustomServices)
						diags = append(diags, diagstmp...)
					}
				}

				curExceptionServiceObj, diagstmp = types.ObjectValue(curExceptionServiceObj.AttributeTypes(ctx), curExceptionServiceObjAttrs)
				diags = append(diags, diagstmp...)
			}
			////////////// end Rule -> Service ///////////////

			// Initialize Exception object with populated values
			curException, diagstmp := types.ObjectValue(
				IfwExceptionAttrTypes,
				map[string]attr.Value{
					"name":              types.StringValue(ruleException.Name),
					"source":            curExceptionSourceObj,
					"country":           parseNameIDList(ctx, ruleException.Country, "rule.exception.country"),
					"device":            parseNameIDList(ctx, ruleException.Device, "rule.exception.device"),
					"device_os":         parseList(ctx, types.StringType, ruleException.DeviceOs, "rule.exception.device_os"),
					"destination":       curExceptionDestObj,
					"service":           curExceptionServiceObj,
					"connection_origin": types.StringValue(ruleException.ConnectionOrigin.String()),
				},
			)
			diags = append(diags, diagstmp...)
			exceptions = append(exceptions, curException)
		}
		curRuleExceptionsObj, diagstmp = types.SetValue(types.ObjectType{AttrTypes: IfwExceptionAttrTypes}, exceptions)
		diags = append(diags, diagstmp...)
		ruleInput.Exceptions = curRuleExceptionsObj
	} else {
		ruleInput.Exceptions = types.SetNull(types.ObjectType{AttrTypes: IfwExceptionAttrTypes})
	}
	////////////// end Rule -> Exceptions ///////////////

	return ruleInput

}
