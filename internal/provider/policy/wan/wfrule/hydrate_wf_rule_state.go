package wfrule

import (
	"context"
	"fmt"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk" // Import the correct package
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/activeperiod"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/customservice"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/oldidname"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyschedule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

//nolint:gocyclo,funlen
func HydrateWanRuleState(
	ctx context.Context,
	state Rule,
	currentRule *cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules_Rule,
) (PolicyPolicyWanFirewallPolicyRulesRule, diag.Diagnostics) {
	ruleInput := PolicyPolicyWanFirewallPolicyRulesRule{}
	diags := make(diag.Diagnostics, 0)

	// Handle case where state might be incomplete (e.g., during import)
	// Try to extract existing state first, but don't fail if it's incomplete
	if !state.Rule.IsNull() && !state.Rule.IsUnknown() {
		diagstmp := state.Rule.As(ctx, &ruleInput, basetypes.ObjectAsOptions{})
		if diagstmp.HasError() {
			// If conversion fails during import, log it but continue with empty ruleInput
			tflog.Debug(ctx, "Failed to convert state to struct in hydrateWanRuleState, using empty struct for hydration")
			// Reset ruleInput to empty struct and clear diagnostics
			ruleInput = PolicyPolicyWanFirewallPolicyRulesRule{}
			// Don't add the errors to diags, we'll handle this by populating from API response
		} else {
			diags = append(diags, diagstmp...)
		}
	}
	ruleInput.Name = types.StringValue(currentRule.Name)
	ruleInput.Direction = types.StringValue(string(currentRule.Direction))
	if currentRule.Description == "" {
		ruleInput.Description = types.StringNull()
	} else {
		ruleInput.Description = types.StringValue(currentRule.Description)
	}
	ruleInput.Action = types.StringValue(currentRule.Action.String())
	ruleInput.ID = types.StringValue(currentRule.ID)
	ruleInput.Enabled = types.BoolValue(currentRule.Enabled)
	ruleInput.ConnectionOrigin = types.StringValue(currentRule.ConnectionOrigin.String())

	// // /// /// /// / Rule -> Source // /// /// /// /// /
	curRuleSourceObj, diagstmp := types.ObjectValue(
		WanSourceAttrTypes,
		map[string]attr.Value{
			"ip": convert.ParseList(ctx, types.StringType, currentRule.Source.IP,
				"rule.source.ip"),
			"host": oldidname.ParseNameIDList(ctx, currentRule.Source.Host,
				"rule.source.host"),
			"site": oldidname.ParseNameIDList(ctx, currentRule.Source.Site,
				"rule.source.site"),
			"subnet": convert.ParseList(ctx, types.StringType, currentRule.Source.Subnet,
				"rule.source.subnet"),
			"ip_range": convert.ParseFromToList(ctx, currentRule.Source.IPRange,
				"rule.source.ip_range"),
			"global_ip_range": oldidname.ParseNameIDList(ctx, currentRule.Source.GlobalIPRange,
				"rule.source.global_ip_range"),
			"network_interface": oldidname.ParseNameIDList(ctx, currentRule.Source.NetworkInterface,
				"rule.source.network_interface"),
			"site_network_subnet": oldidname.ParseNameIDList(ctx, currentRule.Source.SiteNetworkSubnet,
				"rule.source.site_network_subnet"),
			"floating_subnet": oldidname.ParseNameIDList(ctx, currentRule.Source.FloatingSubnet,
				"rule.source.floating_subnet"),
			"user": oldidname.ParseNameIDList(ctx, currentRule.Source.User,
				"rule.source.user"),
			"users_group": oldidname.ParseNameIDList(ctx, currentRule.Source.UsersGroup,
				"rule.source.users_group"),
			"group": oldidname.ParseNameIDList(ctx, currentRule.Source.Group,
				"rule.source.group"),
			"system_group": oldidname.ParseNameIDList(ctx, currentRule.Source.SystemGroup,
				"rule.source.system_group"),
		},
	)
	diags = append(diags, diagstmp...)
	// Ensure source.network_interface is never unknown after apply
	srcAttrs := curRuleSourceObj.Attributes()
	if v, ok := srcAttrs["network_interface"]; ok {
		if setVal, ok2 := v.(types.Set); ok2 {
			if setVal.IsUnknown() {
				// Replace unknown with known null set to satisfy post-apply requirements
				srcAttrs["network_interface"] = types.SetNull(idname.RefObjectType)
				var diagsTmp diag.Diagnostics
				curRuleSourceObj, diagsTmp = types.ObjectValue(curRuleSourceObj.Type(ctx).(types.ObjectType).AttrTypes, srcAttrs)
				diags = append(diags, diagsTmp...)
			}
		}
	}

	ruleInput.Source = curRuleSourceObj
	// /// /// /// /// end rule.source // /// /// /// /// /

	// Rule -> Country
	ruleInput.Country = oldidname.ParseNameIDList(ctx, currentRule.Country,
		"rule.country")
	// Rule -> Device
	ruleInput.Device = oldidname.ParseNameIDList(ctx, currentRule.Device,
		"rule.device")

	// Rule -> DeviceOS
	ruleInput.DeviceOs = convert.ParseList(ctx, types.StringType, currentRule.DeviceOs,
		"rule.source.device_os")

	// /// /// /// / Rule -> Destination // /// /// /// /// /
	curRuleDestinationObj, diagstmp := types.ObjectValue(
		WanDestAttrTypes,
		map[string]attr.Value{
			"ip": convert.ParseList(ctx, types.StringType, currentRule.Destination.IP,
				"rule.destination.ip"),
			"host": oldidname.ParseNameIDList(ctx, currentRule.Destination.Host,
				"rule.destination.host"),
			"site": oldidname.ParseNameIDList(ctx, currentRule.Destination.Site,
				"rule.destination.site"),
			"subnet": convert.ParseList(ctx, types.StringType, currentRule.Destination.Subnet,
				"rule.destination.subnet"),
			"ip_range": convert.ParseFromToList(ctx, currentRule.Destination.IPRange,
				"rule.destination.ip_range"),
			"global_ip_range": oldidname.ParseNameIDList(ctx, currentRule.Destination.GlobalIPRange,
				"rule.destination.global_ip_range"),
			"network_interface": oldidname.ParseNameIDList(ctx, currentRule.Destination.NetworkInterface,
				"rule.destination.network_interface"),
			"site_network_subnet": oldidname.ParseNameIDList(ctx, currentRule.Destination.SiteNetworkSubnet,
				"rule.destination.site_network_subnet"),
			"floating_subnet": oldidname.ParseNameIDList(ctx, currentRule.Destination.FloatingSubnet,
				"rule.destination.floating_subnet"),
			"user": oldidname.ParseNameIDList(ctx, currentRule.Destination.User,
				"rule.destination.user"),
			"users_group": oldidname.ParseNameIDList(ctx, currentRule.Destination.UsersGroup,
				"rule.destination.users_group"),
			"group": oldidname.ParseNameIDList(ctx, currentRule.Destination.Group,
				"rule.destination.group"),
			"system_group": oldidname.ParseNameIDList(ctx, currentRule.Destination.SystemGroup,
				"rule.destination.system_group"),
		},
	)
	diags = append(diags, diagstmp...)
	ruleInput.Destination = curRuleDestinationObj
	// /// /// /// /// end Rule -> Destination // /// /// /// /// /

	// // /// /// /// / Rule -> Application // /// /// /// /// /
	curRuleApplicationObj, diagstmp := types.ObjectValue(
		WanApplicationAttrTypes,
		map[string]attr.Value{
			"application": oldidname.ParseNameIDList(ctx, currentRule.Application.Application,
				"rule.application.application"),
			"custom_app": oldidname.ParseNameIDList(ctx, currentRule.Application.CustomApp,
				"rule.application.host"),
			"app_category": oldidname.ParseNameIDList(ctx, currentRule.Application.AppCategory,
				"rule.application.app_category"),
			"custom_category": oldidname.ParseNameIDList(ctx, currentRule.Application.CustomCategory,
				"rule.application.custom_category"),
			"sanctioned_apps_category": oldidname.ParseNameIDList(
				ctx,
				currentRule.Application.SanctionedAppsCategory,
				"rule.application.sanctioned_apps_category",
			),
			"domain": convert.ParseList(ctx, types.StringType, currentRule.Application.Domain,
				"rule.application.domain"),
			"fqdn": convert.ParseList(ctx, types.StringType, currentRule.Application.Fqdn,
				"rule.application.fqdn"),
			"ip": convert.ParseList(ctx, types.StringType, currentRule.Application.IP,
				"rule.application.ip"),
			"subnet": convert.ParseList(ctx, types.StringType, currentRule.Application.Subnet,
				"rule.application.subnet"),
			"ip_range": convert.ParseFromToList(ctx, currentRule.Application.IPRange,
				"rule.application.ip_range"),
			"global_ip_range": oldidname.ParseNameIDList(ctx, currentRule.Application.GlobalIPRange,
				"rule.source.global_ip_range"),
		},
	)
	ruleInput.Application = curRuleApplicationObj
	diags = append(diags, diagstmp...)
	// /// /// /// /// end Rul -> Application // /// /// /// /// /

	// /// /// /// /// Start Rule -> deviceAttributes // /// /// /// /// /
	// Check if DeviceAttributes has any non-empty fields (not in zero state)
	hasDeviceAttributes := len(currentRule.DeviceAttributes.Category) > 0 ||
		len(currentRule.DeviceAttributes.Type) > 0 ||
		len(currentRule.DeviceAttributes.Model) > 0 ||
		len(currentRule.DeviceAttributes.Manufacturer) > 0 ||
		len(currentRule.DeviceAttributes.Os) > 0 ||
		len(currentRule.DeviceAttributes.OsVersion) > 0

	tflog.Debug(ctx, "WAN_rule.read.currentRule.DeviceAttributes", map[string]interface{}{
		"v": utils.InterfaceToJSONString(fmt.Sprintf("%v", currentRule.DeviceAttributes)),
	})

	var deviceAttributesObj types.Object
	if hasDeviceAttributes {
		deviceAttributesObj, diagstmp = types.ObjectValue(
			WanDeviceAttrAttrTypes,
			map[string]attr.Value{
				"category": convert.ParseList(ctx, types.StringType, currentRule.DeviceAttributes.Category,
					"rule.deviceattributes.category"),
				"type": convert.ParseList(ctx, types.StringType, currentRule.DeviceAttributes.Type,
					"rule.deviceattributes.type"),
				"model": convert.ParseList(ctx, types.StringType, currentRule.DeviceAttributes.Model,
					"rule.deviceattributes.model"),
				"manufacturer": convert.ParseList(ctx, types.StringType, currentRule.DeviceAttributes.Manufacturer,
					"rule.deviceattributes.manufacturer"),
				"os": convert.ParseList(ctx, types.StringType, currentRule.DeviceAttributes.Os,
					"rule.deviceattributes.os"),
				"os_version": convert.ParseList(ctx, types.StringType, currentRule.DeviceAttributes.OsVersion,
					"rule.deviceattributes.os_version"),
			},
		)
		diags = append(diags, diagstmp...)
	} else {
		// Check if plan/state has device_attributes as an object (even if empty)
		// If so, preserve the object structure with null lists instead of returning null object
		if !ruleInput.DeviceAttributes.IsNull() && !ruleInput.DeviceAttributes.IsUnknown() {
			// Plan has device_attributes object, preserve structure with null lists
			deviceAttributesObj, diagstmp = types.ObjectValue(
				WanDeviceAttrAttrTypes,
				map[string]attr.Value{
					"category":     types.ListNull(types.StringType),
					"type":         types.ListNull(types.StringType),
					"model":        types.ListNull(types.StringType),
					"manufacturer": types.ListNull(types.StringType),
					"os":           types.ListNull(types.StringType),
					"os_version":   types.ListNull(types.StringType),
				},
			)
			diags = append(diags, diagstmp...)
		} else {
			// No device attributes in plan: keep null
			deviceAttributesObj = types.ObjectNull(WanDeviceAttrAttrTypes)
		}
	}

	tflog.Debug(ctx, "WAN_rule.read.currentRule.DeviceAttributes", map[string]interface{}{
		"deviceAttributesObj": utils.InterfaceToJSONString(deviceAttributesObj.String()),
	})

	ruleInput.DeviceAttributes = deviceAttributesObj
	// /// /// /// /// End Rule -> deviceAttributes // /// /// /// /// /

	// /// /// /// /// start Rule -> Service // /// /// /// /// /
	// Always initialize Service object to prevent drift
	curRuleServiceObj, diagstmp := types.ObjectValue(
		WanServiceAttrTypes,
		map[string]attr.Value{
			"standard": types.SetNull(idname.RefObjectType),
			"custom":   types.ListNull(customservice.CustomServiceObjectType),
		},
	)
	diags = append(diags, diagstmp...)
	curRuleServiceObjAttrs := curRuleServiceObj.Attributes()

	// Rule -> Service -> Standard
	curRuleServiceObjAttrs["standard"] = oldidname.ParseNameIDList(ctx, currentRule.Service.Standard,
		"rule.service.standard")

	// Rule -> Service -> Custom
	if len(currentRule.Service.Custom) > 0 {
		var curRuleCustomServices []types.Object
		tflog.Info(ctx, "ruleResponse.Service.Custom - "+fmt.Sprintf("%v", currentRule.Service.Custom))
		for _, item := range currentRule.Service.Custom {
			curRuleCustomServices = append(curRuleCustomServices, customservice.ParseCustomService(ctx, item,
				"rule.service.custom"))
		}
		curRuleServiceObjAttrs["custom"], diagstmp = types.ListValueFrom(ctx, customservice.CustomServiceObjectType, curRuleCustomServices)
		diags = append(diags, diagstmp...)
	}

	curRuleServiceObj, diagstmp = types.ObjectValue(curRuleServiceObj.AttributeTypes(ctx), curRuleServiceObjAttrs)
	diags = append(diags, diagstmp...)
	ruleInput.Service = curRuleServiceObj
	// /// /// /// /// end Rule -> Service // /// /// /// /// /

	// /// /// /// /// start Rule -> Tracking // /// /// /// /// /
	curRuleTrackingObj, diagstmp := types.ObjectValue(
		policyrule.TrackingAttrTypes,
		map[string]attr.Value{
			"event": types.ObjectNull(policyrule.TrackingEventAttrTypes),
			"alert": types.ObjectNull(policyrule.TrackingAlertAttrTypes),
		},
	)
	diags = append(diags, diagstmp...)
	curRuleTrackingObjAttrs := curRuleTrackingObj.Attributes()

	// Rule -> Tracking -> Event
	trackingEventValue, diagsObj := types.ObjectValue(
		policyrule.TrackingEventAttrTypes,
		map[string]attr.Value{
			"enabled": types.BoolValue(currentRule.Tracking.Event.Enabled),
		},
	)
	diags = append(diags, diagsObj...)
	curRuleTrackingObjAttrs["event"] = trackingEventValue

	// Rule -> Tracking -> Alert
	trackingAlertValue, diagstmp := types.ObjectValue(
		policyrule.TrackingAlertAttrTypes,
		map[string]attr.Value{
			"enabled":   types.BoolValue(currentRule.Tracking.Alert.Enabled),
			"frequency": types.StringValue(currentRule.Tracking.Alert.Frequency.String()),
			"subscription_group": oldidname.ParseNameIDList(ctx, currentRule.Tracking.Alert.SubscriptionGroup,
				"rule.tracking.alert.subscription_group"),
			"webhook": oldidname.ParseNameIDList(ctx, currentRule.Tracking.Alert.Webhook,
				"rule.tracking.alert.webhook"),
			"mailing_list": oldidname.ParseNameIDList(ctx, currentRule.Tracking.Alert.MailingList,
				"rule.tracking.alert.mailing_list"),
		},
	)
	diags = append(diags, diagstmp...)
	curRuleTrackingObjAttrs["alert"] = trackingAlertValue
	tflog.Warn(ctx, "Updated tracking object: "+curRuleTrackingObj.String())

	curRuleTrackingObj, diagstmp = types.ObjectValue(curRuleTrackingObj.AttributeTypes(ctx), curRuleTrackingObjAttrs)
	diags = append(diags, diagstmp...)
	ruleInput.Tracking = curRuleTrackingObj
	// /// /// /// /// end Rule -> Tracking // /// /// /// /// /

	// /// /// /// /// start Rule -> Schedule // /// /// /// /// /
	curRuleScheduleObj, diagstmp := types.ObjectValue(
		policyschedule.ScheduleAttrTypes,
		map[string]attr.Value{
			"active_on": types.StringValue(currentRule.Schedule.ActiveOn.String()),
			"custom_timeframe": convert.ParseFromTo(ctx, currentRule.Schedule.CustomTimeframePolicySchedule,
				"rule.schedule.custom_timeframe"),
			"custom_recurring": policyschedule.ParseFromToDays(ctx, currentRule.Schedule.CustomRecurringPolicySchedule,
				"rule.schedule.custom_recurring"),
		},
	)
	diags = append(diags, diagstmp...)
	ruleInput.Schedule = curRuleScheduleObj
	// /// /// /// /// end Rule -> Schedule // /// /// /// /// /

	// /// /// /// /// // start Rule -> Exceptions // /// /// /// /// /
	exceptions := []attr.Value{}

	// Rule -> Exceptions -> Source
	tflog.Warn(ctx, "hydrateWanRuleState() currentRule.Exceptions", map[string]interface{}{
		"exceptions": utils.InterfaceToJSONString(currentRule.Exceptions),
		"len":        utils.InterfaceToJSONString(len(currentRule.Exceptions)),
	})
	if len(currentRule.Exceptions) > 0 {
		for _, ruleException := range currentRule.Exceptions {
			// Rule -> Exceptions -> Source
			curExceptionSourceObj, diagstmp := types.ObjectValue(
				WanSourceAttrTypes,
				map[string]attr.Value{
					"ip": convert.ParseList(ctx, types.StringType, ruleException.Source.IP,
						"rule.exception.source.ip"),
					"host": oldidname.ParseNameIDList(ctx, ruleException.Source.Host,
						"rule.exception.source.host"),
					"site": oldidname.ParseNameIDList(ctx, ruleException.Source.Site,
						"rule.exception.source.site"),
					"subnet": convert.ParseList(ctx, types.StringType, ruleException.Source.Subnet,
						"rule.exception.source.subnet"),
					"ip_range": convert.ParseFromToList(ctx, ruleException.Source.IPRange,
						"rule.exception.source.ip_range"),
					"global_ip_range": oldidname.ParseNameIDList(ctx, ruleException.Source.GlobalIPRange,
						"rule.exception.source.global_ip_range"),
					"network_interface": oldidname.ParseNameIDList(ctx, ruleException.Source.NetworkInterface,
						"rule.exception.source.network_interface"),
					"site_network_subnet": oldidname.ParseNameIDList(ctx, ruleException.Source.SiteNetworkSubnet,
						"rule.exception.source.site_network_subnet"),
					"floating_subnet": oldidname.ParseNameIDList(ctx, ruleException.Source.FloatingSubnet,
						"rule.exception.source.floating_subnet"),
					"user": oldidname.ParseNameIDList(ctx, ruleException.Source.User,
						"rule.exception.source.user"),
					"users_group": oldidname.ParseNameIDList(ctx, ruleException.Source.UsersGroup,
						"rule.exception.source.users_group"),
					"group": oldidname.ParseNameIDList(ctx, ruleException.Source.Group,
						"rule.exception.source.group"),
					"system_group": oldidname.ParseNameIDList(ctx, ruleException.Source.SystemGroup,
						"rule.exception.source.system_group"),
				},
			)
			diags = append(diags, diagstmp...)

			// Rule -> Exceptions -> Destination
			curExceptionDestObj, diagstmp := types.ObjectValue(
				WanDestAttrTypes,
				map[string]attr.Value{
					"ip": convert.ParseList(ctx, types.StringType, ruleException.Destination.IP,
						"rule.exception.destination.ip"),
					"host": oldidname.ParseNameIDList(ctx, ruleException.Destination.Host,
						"rule.exception.destination.host"),
					"site": oldidname.ParseNameIDList(ctx, ruleException.Destination.Site,
						"rule.exception.destination.site"),
					"subnet": convert.ParseList(ctx, types.StringType, ruleException.Destination.Subnet,
						"rule.exception.destination.subnet"),
					"ip_range": convert.ParseFromToList(ctx, ruleException.Destination.IPRange,
						"rule.exception.destination.ip_range"),
					"global_ip_range": oldidname.ParseNameIDList(ctx, ruleException.Destination.GlobalIPRange,
						"rule.exception.destination.global_ip_range"),
					"network_interface": oldidname.ParseNameIDList(
						ctx,
						ruleException.Destination.NetworkInterface,
						"rule.exception.destination.network_interface",
					),
					"site_network_subnet": oldidname.ParseNameIDList(
						ctx,
						ruleException.Destination.SiteNetworkSubnet,
						"rule.exception.destination.site_network_subnet",
					),
					"floating_subnet": oldidname.ParseNameIDList(ctx, ruleException.Destination.FloatingSubnet,
						"rule.exception.destination.floating_subnet"),
					"user": oldidname.ParseNameIDList(ctx, ruleException.Destination.User,
						"rule.exception.destination.user"),
					"users_group": oldidname.ParseNameIDList(ctx, ruleException.Destination.UsersGroup,
						"rule.exception.destination.users_group"),
					"group": oldidname.ParseNameIDList(ctx, ruleException.Destination.Group,
						"rule.exception.destination.group"),
					"system_group": oldidname.ParseNameIDList(ctx, ruleException.Destination.SystemGroup,
						"rule.exception.destination.system_group"),
				},
			)
			diags = append(diags, diagstmp...)

			// // /// /// /// / Rule -> Exceptions -> Application // /// /// /// /// /
			curExceptionApplicationObj, diagstmp := types.ObjectValue(
				WanApplicationAttrTypes,
				map[string]attr.Value{
					"application": oldidname.ParseNameIDList(ctx, ruleException.Application.Application,
						"rule.exception.application.application"),
					"custom_app": oldidname.ParseNameIDList(ctx, ruleException.Application.CustomApp,
						"rule.exception.application.host"),
					"app_category": oldidname.ParseNameIDList(ctx, ruleException.Application.AppCategory,
						"rule.exception.application.app_category"),
					"custom_category": oldidname.ParseNameIDList(
						ctx,
						ruleException.Application.CustomCategory,
						"rule.exception.application.custom_category",
					),
					"sanctioned_apps_category": oldidname.ParseNameIDList(
						ctx,
						ruleException.Application.SanctionedAppsCategory,
						"rule.exception.application.sanctioned_apps_category",
					),
					"domain": convert.ParseList(ctx, types.StringType, ruleException.Application.Domain,
						"rule.exception.application.domain"),
					"fqdn": convert.ParseList(ctx, types.StringType, ruleException.Application.Fqdn,
						"rule.exception.application.fqdn"),
					"ip": convert.ParseList(ctx, types.StringType, ruleException.Application.IP,
						"rule.exception.application.ip"),
					"subnet": convert.ParseList(ctx, types.StringType, ruleException.Application.Subnet,
						"rule.exception.application.subnet"),
					"ip_range": convert.ParseFromToList(ctx, ruleException.Application.IPRange,
						"rule.exception.application.ip_range"),
					"global_ip_range": oldidname.ParseNameIDList(
						ctx,
						ruleException.Application.GlobalIPRange,
						"rule.exception.application.global_ip_range",
					),
				},
			)
			diags = append(diags, diagstmp...)
			// /// /// /// /// end Rul -> Exceptions -> Application // /// /// /// /// /

			// /// /// /// /// start Rule -> Service // /// /// /// /// /
			// Initialize Service object with null values
			curExceptionServiceObj, diagstmp := types.ObjectValue(
				WanServiceAttrTypes,
				map[string]attr.Value{
					"standard": types.SetNull(idname.RefObjectType),
					"custom":   types.ListNull(customservice.CustomServiceObjectType),
				},
			)
			diags = append(diags, diagstmp...)
			curExceptionServiceObjAttrs := curExceptionServiceObj.Attributes()
			if len(ruleException.Service.Custom) > 0 || len(ruleException.Service.Standard) > 0 {
				// Rule -> Service -> Standard
				curExceptionServiceObjAttrs["standard"] = oldidname.ParseNameIDList(ctx, ruleException.Service.Standard,
					"rule.exception.service.standard")

				// Rule -> Service -> Custom (use special parser for exceptions that handles portRangeCustomService)
				if ruleException.Service.Custom != nil {
					if len(ruleException.Service.Custom) > 0 {
						var curExceptionCustomServices []types.Object
						tflog.Info(ctx, "ruleException.Service.Custom - "+fmt.Sprintf("%v", ruleException.Service.Custom))
						for _, item := range ruleException.Service.Custom {
							curExceptionCustomServices = append(
								curExceptionCustomServices,
								customservice.ParseExceptionCustomService(ctx, item,
									"rule.exception.service.custom"),
							)
						}
						curExceptionServiceObjAttrs["custom"], diagstmp = types.ListValueFrom(ctx,
							customservice.CustomServiceObjectType, curExceptionCustomServices)
						diags = append(diags, diagstmp...)
					}
				}

				curExceptionServiceObj, diagstmp = types.ObjectValue(curExceptionServiceObj.AttributeTypes(ctx),
					curExceptionServiceObjAttrs)
				diags = append(diags, diagstmp...)
			}
			// /// /// /// /// end Rule -> Service // /// /// /// /// /

			// Check if DeviceAttributes has any non-empty fields for exceptions (similar to main rule logic)
			exceptionHasDeviceAttributes := len(ruleException.DeviceAttributes.Category) > 0 ||
				len(ruleException.DeviceAttributes.Type) > 0 ||
				len(ruleException.DeviceAttributes.Model) > 0 ||
				len(ruleException.DeviceAttributes.Manufacturer) > 0 ||
				len(ruleException.DeviceAttributes.Os) > 0 ||
				len(ruleException.DeviceAttributes.OsVersion) > 0

			var exceptionDeviceAttributesObj types.Object
			if exceptionHasDeviceAttributes {
				exceptionDeviceAttributesObj, diagstmp = types.ObjectValue(
					WanDeviceAttrAttrTypes,
					map[string]attr.Value{
						"category": convert.ParseList(ctx, types.StringType, ruleException.DeviceAttributes.Category,
							"rule.exception.deviceattributes.category"),
						"type": convert.ParseList(ctx, types.StringType, ruleException.DeviceAttributes.Type,
							"rule.exception.deviceattributes.type"),
						"model": convert.ParseList(ctx, types.StringType, ruleException.DeviceAttributes.Model,
							"rule.exception.deviceattributes.model"),
						"manufacturer": convert.ParseList(
							ctx,
							types.StringType,
							ruleException.DeviceAttributes.Manufacturer,
							"rule.exception.deviceattributes.manufacturer",
						),
						"os": convert.ParseList(ctx, types.StringType, ruleException.DeviceAttributes.Os,
							"rule.exception.deviceattributes.os"),
						"os_version": convert.ParseList(
							ctx,
							types.StringType,
							ruleException.DeviceAttributes.OsVersion,
							"rule.exception.deviceattributes.os_version",
						),
					},
				)
				diags = append(diags, diagstmp...)
			} else {
				// Set to null when no device attributes data from API
				// This prevents drift when device_attributes is not specified in config
				exceptionDeviceAttributesObj = types.ObjectNull(WanDeviceAttrAttrTypes)
			}

			// Initialize Exception object with populated values
			curException, diagstmp := types.ObjectValue(
				WanExceptionAttrTypes,
				map[string]attr.Value{
					"name":   types.StringValue(ruleException.Name),
					"source": curExceptionSourceObj,
					"country": oldidname.ParseNameIDListOrEmptySet(ctx, ruleException.Country,
						"rule.exception.country"),
					"device": oldidname.ParseNameIDListOrEmptySet(ctx, ruleException.Device,
						"rule.exception.device"),
					"device_attributes": exceptionDeviceAttributesObj,
					"device_os": convert.ParseList(ctx, types.StringType, ruleException.DeviceOs,
						"rule.exception.device_os"),
					"destination":       curExceptionDestObj,
					"application":       curExceptionApplicationObj,
					"service":           curExceptionServiceObj,
					"direction":         types.StringValue(ruleException.Direction.String()),
					"connection_origin": types.StringValue(ruleException.ConnectionOrigin.String()),
				},
			)
			diags = append(diags, diagstmp...)
			exceptions = append(exceptions, curException)
		}
		curRuleExceptionsObj, diagstmp := types.SetValue(types.ObjectType{AttrTypes: WanExceptionAttrTypes}, exceptions)
		diags = append(diags, diagstmp...)
		ruleInput.Exceptions = curRuleExceptionsObj
	} else {
		// Use empty set instead of null to match schema expectations
		curRuleExceptionsObj, diagstmp := types.SetValue(types.ObjectType{AttrTypes: WanExceptionAttrTypes}, []attr.Value{})
		diags = append(diags, diagstmp...)
		ruleInput.Exceptions = curRuleExceptionsObj
	}
	// /// /// /// /// end Rule -> Exceptions // /// /// /// /// /

	// /// /// /// /// start Rule -> ActivePeriod // /// /// /// /// /
	// Debug logging to see what API returns
	tflog.Warn(ctx, "ActivePeriod from API", map[string]interface{}{
		"EffectiveFrom":    utils.InterfaceToJSONString(currentRule.ActivePeriod.EffectiveFrom),
		"ExpiresAt":        utils.InterfaceToJSONString(currentRule.ActivePeriod.ExpiresAt),
		"UseEffectiveFrom": utils.InterfaceToJSONString(currentRule.ActivePeriod.UseEffectiveFrom),
		"UseExpiresAt":     utils.InterfaceToJSONString(currentRule.ActivePeriod.UseExpiresAt),
	})

	effectiveFromValue := activeperiod.GetActivePeriodString(currentRule.ActivePeriod.EffectiveFrom)
	expiresAtValue := activeperiod.GetActivePeriodString(currentRule.ActivePeriod.ExpiresAt)
	useEffectiveFromValue := types.BoolValue(currentRule.ActivePeriod.UseEffectiveFrom)
	useExpiresAtValue := types.BoolValue(currentRule.ActivePeriod.UseExpiresAt)

	if effectiveFromValue.IsUnknown() {
		effectiveFromValue = types.StringNull()
	}
	if expiresAtValue.IsUnknown() {
		expiresAtValue = types.StringNull()
	}

	configuredActivePeriod := activeperiod.ActivePeriod{}
	if !ruleInput.ActivePeriod.IsNull() && !ruleInput.ActivePeriod.IsUnknown() {
		diagstmp := ruleInput.ActivePeriod.As(ctx, &configuredActivePeriod, basetypes.ObjectAsOptions{})
		diags = append(diags, diagstmp...)
	}

	// If API returned nil but we have configured values in state, preserve them
	// Only attempt to preserve values if ActivePeriod is not null and not unknown
	tflog.Warn(ctx, "TFLOG_WARN_WAN.ruleInput.ActivePeriod", map[string]interface{}{
		"OUTPUT":                utils.InterfaceToJSONString(ruleInput.ActivePeriod),
		"effectiveFromValue":    utils.InterfaceToJSONString(effectiveFromValue.ValueString()),
		"expiresAtValue":        utils.InterfaceToJSONString(expiresAtValue.ValueString()),
		"useEffectiveFromValue": utils.InterfaceToJSONString(useEffectiveFromValue.ValueBool()),
		"useExpiresAtValue":     utils.InterfaceToJSONString(useExpiresAtValue.ValueBool()),
	})

	// Preserve effective_from if API returned nil
	if !effectiveFromValue.IsNull() && useEffectiveFromValue.ValueBool() {
		parsedEffectiveFromStr, err := utils.ParseTimeString(effectiveFromValue.ValueString(), configuredActivePeriod.EffectiveFrom.ValueString())
		if err == nil {
			tflog.Warn(ctx, "TFLOG_WARN_WAN.ruleInput.parsedEffectiveFromStr", map[string]interface{}{
				"OUTPUT": utils.InterfaceToJSONString(parsedEffectiveFromStr),
				"error":  utils.InterfaceToJSONString(err),
			})
			effectiveFromValue = types.StringValue(parsedEffectiveFromStr)
		}
	} else {
		effectiveFromValue = types.StringNull() // If no value, set to null
	}

	// Preserve expires_at if API returned nil
	if !expiresAtValue.IsNull() && useExpiresAtValue.ValueBool() {
		parsedExpiresAtStr, err := utils.ParseTimeString(expiresAtValue.ValueString(), configuredActivePeriod.ExpiresAt.ValueString())
		if err == nil {
			tflog.Warn(ctx, "TFLOG_WARN_WAN.ruleInput.parsedExpiresAtStr", map[string]interface{}{
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
	hasEffectiveFrom := !effectiveFromValue.IsNull() &&
		!effectiveFromValue.IsUnknown() &&
		effectiveFromValue.ValueString() != ""
	useEffectiveFromValue = types.BoolValue(hasEffectiveFrom)
	useExpiresAtValue = types.BoolValue(!expiresAtValue.IsNull() && !expiresAtValue.IsUnknown() && expiresAtValue.ValueString() != "")

	tflog.Warn(ctx, "TFLOG_WARN_WAN.ruleInput.ActivePeriod", map[string]interface{}{
		"effectiveFromValue": utils.InterfaceToJSONString(effectiveFromValue),
		"expiresAtValue":     utils.InterfaceToJSONString(expiresAtValue),
	})

	curRuleActivePeriodObj, diagstmp := types.ObjectValue(
		activeperiod.ActivePeriodAttrTypes,
		map[string]attr.Value{
			"effective_from":     effectiveFromValue,
			"expires_at":         expiresAtValue,
			"use_effective_from": useEffectiveFromValue,
			"use_expires_at":     useExpiresAtValue,
		},
	)
	diags = append(diags, diagstmp...)
	ruleInput.ActivePeriod = curRuleActivePeriodObj
	tflog.Warn(ctx, "Final ActivePeriod WAN object: "+curRuleActivePeriodObj.String())
	// /// /// /// /// end Rule -> ActivePeriod // /// /// /// /// /

	return ruleInput, diags
}
