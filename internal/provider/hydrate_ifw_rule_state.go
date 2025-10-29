package provider

import (
	"context"
	"fmt"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk" // Import the correct package
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
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
	ruleInput.Enabled = types.BoolValue(currentRule.Enabled)
	if currentRule.Description == "" {
		ruleInput.Description = types.StringNull()
	} else {
		ruleInput.Description = types.StringValue(currentRule.Description)
	}
	ruleInput.Action = types.StringValue(currentRule.Action.String())
	ruleInput.ID = types.StringValue(currentRule.ID)
	ruleInput.Index = types.Int64Value(int64(currentRule.Index))
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

	////////////// Start Rule -> deviceAttributes ///////////////
	// Check if DeviceAttributes has any non-empty fields (not in zero state)
	hasDeviceAttributes := len(currentRule.DeviceAttributes.Category) > 0 ||
		len(currentRule.DeviceAttributes.Type) > 0 ||
		len(currentRule.DeviceAttributes.Model) > 0 ||
		len(currentRule.DeviceAttributes.Manufacturer) > 0 ||
		len(currentRule.DeviceAttributes.Os) > 0 ||
		len(currentRule.DeviceAttributes.OsVersion) > 0

	tflog.Warn(ctx, "IFW_rule.read.currentRule.DeviceAttributes DEBUG", map[string]interface{}{
		"hasDeviceAttributes": hasDeviceAttributes,
		"Category":            len(currentRule.DeviceAttributes.Category),
		"Type":                len(currentRule.DeviceAttributes.Type),
		"Model":               len(currentRule.DeviceAttributes.Model),
		"Manufacturer":        len(currentRule.DeviceAttributes.Manufacturer),
		"Os":                  len(currentRule.DeviceAttributes.Os),
		"OsVersion":           len(currentRule.DeviceAttributes.OsVersion),
		"DeviceAttributes":    utils.InterfaceToJSONString(fmt.Sprintf("%v", currentRule.DeviceAttributes)),
	})

	var deviceAttributesObj types.Object
	if hasDeviceAttributes {
		deviceAttributesObj, diagstmp = types.ObjectValue(
			IfwDeviceAttrAttrTypes,
			map[string]attr.Value{
				"category":     parseList(ctx, types.StringType, currentRule.DeviceAttributes.Category, "rule.deviceattributes.category"),
				"type":         parseList(ctx, types.StringType, currentRule.DeviceAttributes.Type, "rule.deviceattributes.type"),
				"model":        parseList(ctx, types.StringType, currentRule.DeviceAttributes.Model, "rule.deviceattributes.model"),
				"manufacturer": parseList(ctx, types.StringType, currentRule.DeviceAttributes.Manufacturer, "rule.deviceattributes.manufacturer"),
				"os":           parseList(ctx, types.StringType, currentRule.DeviceAttributes.Os, "rule.deviceattributes.os"),
				"os_version":   parseList(ctx, types.StringType, currentRule.DeviceAttributes.OsVersion, "rule.deviceattributes.os_version"),
			},
		)
		diags = append(diags, diagstmp...)
	} else {
		// Set device_attributes to null when no attributes are present
		deviceAttributesObj = types.ObjectNull(IfwDeviceAttrAttrTypes)
	}

	tflog.Debug(ctx, "IFW_rule.read.currentRule.DeviceAttributes", map[string]interface{}{
		"deviceAttributesObj": utils.InterfaceToJSONString(fmt.Sprintf("%v", deviceAttributesObj)),
	})

	ruleInput.DeviceAttributes = deviceAttributesObj
	////////////// End Rule -> deviceAttributes ///////////////

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
	trackingEventValue, diagstmp := types.ObjectValue(
		TrackingEventAttrTypes,
		map[string]attr.Value{
			"enabled": types.BoolValue(currentRule.Tracking.Event.Enabled),
		},
	)
	diags = append(diags, diagstmp...)
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

	////////////// start Rule -> Exceptions ///////////////
	tflog.Warn(ctx, "TFLOG_WARN_IFW.currentRule.Exceptions", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(currentRule.Exceptions),
	})

	exceptions := []attr.Value{}
	if len(currentRule.Exceptions) > 0 {
		for _, ruleException := range currentRule.Exceptions {
			// Rule -> Exceptions -> Source
			curExceptionSourceObj, diagstmp := types.ObjectValue(
				IfwSourceAttrTypes,
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
			tflog.Warn(ctx, "TFLOG_WARN_IFW.curExceptionSourceObj", map[string]interface{}{
				"curExceptionSourceObj": utils.InterfaceToJSONString(curExceptionSourceObj),
			})

			diags = append(diags, diagstmp...)

			// Rule -> Exceptions -> Destination
			curExceptionDestObj, diagstmp := types.ObjectValue(
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
			diags = append(diags, diagstmp...)

			////////////// start Rule -> Service ///////////////
			// Initialize Service object with null values (matching WAN pattern)
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
				// Rule -> Service -> Standard (use regular parseNameIDList like WAN firewall)
				curExceptionServiceObjAttrs["standard"] = parseNameIDList(ctx, ruleException.Service.Standard, "rule.exception.service.standard")

				// Rule -> Service -> Custom
				if ruleException.Service.Custom != nil {
					if len(ruleException.Service.Custom) > 0 {
						var curExceptionCustomServices []types.Object
						tflog.Info(ctx, "ruleException.Service.Custom - "+fmt.Sprintf("%v", ruleException.Service.Custom))
						for _, item := range ruleException.Service.Custom {
							// Exceptions use PortRangeCustomService in the API; match WAN behavior
							curExceptionCustomServices = append(curExceptionCustomServices, parseExceptionCustomService(ctx, item, "rule.exception.service.custom"))
						}
						curExceptionServiceObjAttrs["custom"], diagstmp = types.ListValueFrom(ctx, CustomServiceObjectType, curExceptionCustomServices)
						diags = append(diags, diagstmp...)
					}
				}

				curExceptionServiceObj, diagstmp = types.ObjectValue(curExceptionServiceObj.AttributeTypes(ctx), curExceptionServiceObjAttrs)
				diags = append(diags, diagstmp...)
			}
			////////////// end Rule -> Service ///////////////

			// Note: Internet Firewall exceptions do not have DeviceAttributes in the API response
			// Always set device_attributes to null for IFW exceptions
			exceptionDeviceAttributesObj := types.ObjectNull(IfwDeviceAttrAttrTypes)

			// Initialize Exception object with populated values (use regular parseNameIDList like WAN firewall)
			exceptionAttrs := map[string]attr.Value{
				"name":              types.StringValue(ruleException.Name),
				"source":            curExceptionSourceObj,
				"country":           parseNameIDList(ctx, ruleException.Country, "rule.exception.country"),
				"device":            parseNameIDList(ctx, ruleException.Device, "rule.exception.device"),
				"device_attributes": exceptionDeviceAttributesObj,
				"device_os":         parseList(ctx, types.StringType, ruleException.DeviceOs, "rule.exception.device_os"),
				"destination":       curExceptionDestObj,
				"service":           curExceptionServiceObj,
				"connection_origin": types.StringValue(ruleException.ConnectionOrigin.String()),
			}

			curException, diagstmp := types.ObjectValue(
				IfwExceptionAttrTypes,
				exceptionAttrs,
			)
			diags = append(diags, diagstmp...)
			tflog.Warn(ctx, "hydrateIFRuleState() Created exception: "+fmt.Sprintf("%+v", curException))
			exceptions = append(exceptions, curException)
		}
		curRuleExceptionsObj, diagstmp := types.SetValue(types.ObjectType{AttrTypes: IfwExceptionAttrTypes}, exceptions)
		diags = append(diags, diagstmp...)
		ruleInput.Exceptions = curRuleExceptionsObj
	} else {
		// Use empty set instead of null to match schema expectations
		curRuleExceptionsObj, diagstmp := types.SetValue(types.ObjectType{AttrTypes: IfwExceptionAttrTypes}, []attr.Value{})
		diags = append(diags, diagstmp...)
		ruleInput.Exceptions = curRuleExceptionsObj
	}
	////////////// end Rule -> Exceptions ///////////////

	////////////// start Rule -> ActivePeriod ///////////////
	// Debug logging to see what API returns
	tflog.Warn(ctx, "ActivePeriod from API: EffectiveFrom="+fmt.Sprintf("%v", currentRule.ActivePeriod.EffectiveFrom)+", ExpiresAt="+fmt.Sprintf("%v", currentRule.ActivePeriod.ExpiresAt)+", UseEffectiveFrom="+fmt.Sprintf("%v", currentRule.ActivePeriod.UseEffectiveFrom)+", UseExpiresAt="+fmt.Sprintf("%v", currentRule.ActivePeriod.UseExpiresAt))

	effectiveFromValue := getActivePeriodString(currentRule.ActivePeriod.EffectiveFrom)
	expiresAtValue := getActivePeriodString(currentRule.ActivePeriod.ExpiresAt)
	useEffectiveFromValue := types.BoolValue(currentRule.ActivePeriod.UseEffectiveFrom)
	useExpiresAtValue := types.BoolValue(currentRule.ActivePeriod.UseExpiresAt)

	if effectiveFromValue.IsUnknown() {
		effectiveFromValue = types.StringNull()
	}
	if expiresAtValue.IsUnknown() {
		expiresAtValue = types.StringNull()
	}

	// If API returned nil but we have configured values in state, preserve them
	// Only attempt to preserve values if ActivePeriod is not null and not unknown
	tflog.Warn(ctx, "TFLOG_WARN_IFW.ruleInput.ActivePeriod", map[string]interface{}{
		"OUTPUT":                utils.InterfaceToJSONString(ruleInput.ActivePeriod),
		"effectiveFromValue":    utils.InterfaceToJSONString(effectiveFromValue.ValueString()),
		"expiresAtValue":        utils.InterfaceToJSONString(expiresAtValue.ValueString()),
		"useEffectiveFromValue": utils.InterfaceToJSONString(useEffectiveFromValue.ValueBool()),
		"useExpiresAtValue":     utils.InterfaceToJSONString(useExpiresAtValue.ValueBool()),
	})

	// Preserve effective_from if API returned nil
	if !effectiveFromValue.IsNull() && useEffectiveFromValue.ValueBool() == true {
		parsedEffectiveFromStr, err := parseTimeString(effectiveFromValue.ValueString())
		if err == nil {
			tflog.Warn(ctx, "TFLOG_WARN_IFW.ruleInput.parsedEffectiveFromStr", map[string]interface{}{
				"OUTPUT": utils.InterfaceToJSONString(parsedEffectiveFromStr),
				"error":  utils.InterfaceToJSONString(err),
			})
			effectiveFromValue = types.StringValue(parsedEffectiveFromStr)
		}
		// useEffectiveFromValue = types.BoolValue(true) // If we have a value, set use flag to true
	} else {
		effectiveFromValue = types.StringNull() // If no value, set to null
	}

	// Preserve expires_at if API returned nil
	if !expiresAtValue.IsNull() && useExpiresAtValue.ValueBool() == true {
		parsedExpiresAtStr, err := parseTimeString(expiresAtValue.ValueString())
		if err == nil {
			tflog.Warn(ctx, "TFLOG_WARN_IFW.ruleInput.parsedExpiresAtStr", map[string]interface{}{
				"OUTPUT": utils.InterfaceToJSONString(parsedExpiresAtStr),
				"error":  utils.InterfaceToJSONString(err),
			})
			expiresAtValue = types.StringValue(parsedExpiresAtStr)
		}
	} else {
		expiresAtValue = types.StringNull() // If no value, set to null
	}

	// Recompute use_effective_from and use_expires_at based on final values after preservation logic
	// This ensures consistency: use_effective_from should be true only when effective_from has a value
	useEffectiveFromValue = types.BoolValue(!effectiveFromValue.IsNull() && !effectiveFromValue.IsUnknown() && effectiveFromValue.ValueString() != "")
	useExpiresAtValue = types.BoolValue(!expiresAtValue.IsNull() && !expiresAtValue.IsUnknown() && expiresAtValue.ValueString() != "")

	tflog.Warn(ctx, "TFLOG_WARN_IFW.ruleInput.ActivePeriod", map[string]interface{}{
		"effectiveFromValue": utils.InterfaceToJSONString(effectiveFromValue),
		"expiresAtValue":     utils.InterfaceToJSONString(expiresAtValue),
	})

	curRuleActivePeriodObj, diagstmp := types.ObjectValue(
		ActivePeriodAttrTypes,
		map[string]attr.Value{
			"effective_from":     effectiveFromValue,
			"expires_at":         expiresAtValue,
			"use_effective_from": useEffectiveFromValue,
			"use_expires_at":     useExpiresAtValue,
		},
	)
	diags = append(diags, diagstmp...)
	ruleInput.ActivePeriod = curRuleActivePeriodObj
	tflog.Warn(ctx, "Final ActivePeriod IFW object: "+fmt.Sprintf("%v", curRuleActivePeriodObj))
	////////////// end Rule -> ActivePeriod ///////////////

	return ruleInput

}
