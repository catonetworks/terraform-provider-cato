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

	// ////////////// start Rule -> Exceptions ///////////////
	exceptions := []attr.Value{}

	// Rule -> Exceptions -> Source
	tflog.Warn(ctx, "hydrateIFRuleState() currentRule.Exceptions - "+fmt.Sprintf("%v", currentRule.Exceptions))
	if currentRule.Exceptions != nil && len(currentRule.Exceptions) > 0 {
		for _, ruleException := range currentRule.Exceptions {
			// Rule -> Exceptions -> Source
			// Build source object using conditional fields to match config logic
			sourceAttrs := map[string]attr.Value{}

			// Only include fields that have data to match config filtering logic (: k => v if v != null)
			if len(ruleException.Source.IP) > 0 {
				sourceAttrs["ip"] = parseList(ctx, types.StringType, ruleException.Source.IP, "rule.exception.source.ip")
			}
			if len(ruleException.Source.Host) > 0 {
				sourceAttrs["host"] = parseNameIDList(ctx, ruleException.Source.Host, "rule.exception.source.host")
			}
			if len(ruleException.Source.Site) > 0 {
				sourceAttrs["site"] = parseNameIDList(ctx, ruleException.Source.Site, "rule.exception.source.site")
			}
			if len(ruleException.Source.Subnet) > 0 {
				sourceAttrs["subnet"] = parseList(ctx, types.StringType, ruleException.Source.Subnet, "rule.exception.source.subnet")
			}
			if len(ruleException.Source.IPRange) > 0 {
				sourceAttrs["ip_range"] = parseFromToList(ctx, ruleException.Source.IPRange, "rule.exception.source.ip_range")
			}
			if len(ruleException.Source.GlobalIPRange) > 0 {
				sourceAttrs["global_ip_range"] = parseNameIDList(ctx, ruleException.Source.GlobalIPRange, "rule.exception.source.global_ip_range")
			}
			if len(ruleException.Source.NetworkInterface) > 0 {
				sourceAttrs["network_interface"] = parseNameIDList(ctx, ruleException.Source.NetworkInterface, "rule.exception.source.network_interface")
			}
			if len(ruleException.Source.SiteNetworkSubnet) > 0 {
				sourceAttrs["site_network_subnet"] = parseNameIDList(ctx, ruleException.Source.SiteNetworkSubnet, "rule.exception.source.site_network_subnet")
			}
			if len(ruleException.Source.FloatingSubnet) > 0 {
				sourceAttrs["floating_subnet"] = parseNameIDList(ctx, ruleException.Source.FloatingSubnet, "rule.exception.source.floating_subnet")
			}
			if len(ruleException.Source.User) > 0 {
				sourceAttrs["user"] = parseNameIDList(ctx, ruleException.Source.User, "rule.exception.source.user")
			}
			if len(ruleException.Source.UsersGroup) > 0 {
				sourceAttrs["users_group"] = parseNameIDList(ctx, ruleException.Source.UsersGroup, "rule.exception.source.users_group")
			}
			if len(ruleException.Source.Group) > 0 {
				sourceAttrs["group"] = parseNameIDList(ctx, ruleException.Source.Group, "rule.exception.source.group")
			}
			if len(ruleException.Source.SystemGroup) > 0 {
				sourceAttrs["system_group"] = parseNameIDList(ctx, ruleException.Source.SystemGroup, "rule.exception.source.system_group")
			}

			// If no source fields have data, create an empty object
			if len(sourceAttrs) == 0 {
				// Create empty source object with all null fields to match schema
				sourceAttrs = map[string]attr.Value{
					"ip":                  types.ListNull(types.StringType),
					"host":                types.SetNull(NameIDObjectType),
					"site":                types.SetNull(NameIDObjectType),
					"subnet":              types.ListNull(types.StringType),
					"ip_range":            types.ListNull(FromToObjectType),
					"global_ip_range":     types.SetNull(NameIDObjectType),
					"network_interface":   types.SetNull(NameIDObjectType),
					"site_network_subnet": types.SetNull(NameIDObjectType),
					"floating_subnet":     types.SetNull(NameIDObjectType),
					"user":                types.SetNull(NameIDObjectType),
					"users_group":         types.SetNull(NameIDObjectType),
					"group":               types.SetNull(NameIDObjectType),
					"system_group":        types.SetNull(NameIDObjectType),
				}
			} else {
				// Fill in missing fields with null values to complete the schema
				schemaFields := []string{"ip", "host", "site", "subnet", "ip_range", "global_ip_range", "network_interface", "site_network_subnet", "floating_subnet", "user", "users_group", "group", "system_group"}
				for _, field := range schemaFields {
					if _, exists := sourceAttrs[field]; !exists {
						switch field {
						case "ip", "subnet":
							sourceAttrs[field] = types.ListNull(types.StringType)
						case "ip_range":
							sourceAttrs[field] = types.ListNull(FromToObjectType)
						default:
							sourceAttrs[field] = types.SetNull(NameIDObjectType)
						}
					}
				}
			}

			curExceptionSourceObj, diagstmp := types.ObjectValue(
				IfwSourceAttrTypes,
				sourceAttrs,
			)
			diags = append(diags, diagstmp...)

			// Rule -> Exceptions -> Destination
			// Build destination object using conditional fields to match config logic
			destAttrs := map[string]attr.Value{}

			// Only include fields that have data to match config filtering logic (: k => v if v != null)
			if len(ruleException.Destination.Application) > 0 {
				destAttrs["application"] = parseNameIDList(ctx, ruleException.Destination.Application, "rule.exception.destination.application")
			}
			if len(ruleException.Destination.CustomApp) > 0 {
				destAttrs["custom_app"] = parseNameIDList(ctx, ruleException.Destination.CustomApp, "rule.exception.destination.custom_app")
			}
			if len(ruleException.Destination.AppCategory) > 0 {
				destAttrs["app_category"] = parseNameIDList(ctx, ruleException.Destination.AppCategory, "rule.exception.destination.app_category")
			}
			if len(ruleException.Destination.CustomCategory) > 0 {
				destAttrs["custom_category"] = parseNameIDList(ctx, ruleException.Destination.CustomCategory, "rule.exception.destination.custom_category")
			}
			if len(ruleException.Destination.SanctionedAppsCategory) > 0 {
				destAttrs["sanctioned_apps_category"] = parseNameIDList(ctx, ruleException.Destination.SanctionedAppsCategory, "rule.exception.destination.sanctioned_apps_category")
			}
			if len(ruleException.Destination.Country) > 0 {
				destAttrs["country"] = parseNameIDList(ctx, ruleException.Destination.Country, "rule.exception.destination.country")
			}
			if len(ruleException.Destination.Domain) > 0 {
				destAttrs["domain"] = parseList(ctx, types.StringType, ruleException.Destination.Domain, "rule.exception.destination.domain")
			}
			if len(ruleException.Destination.Fqdn) > 0 {
				destAttrs["fqdn"] = parseList(ctx, types.StringType, ruleException.Destination.Fqdn, "rule.exception.destination.fqdn")
			}
			// Note: Skip ip, subnet, remote_asn as they are computed/API-only fields per original comment
			if len(ruleException.Destination.IPRange) > 0 {
				destAttrs["ip_range"] = parseFromToList(ctx, ruleException.Destination.IPRange, "rule.exception.destination.ip_range")
			}
			if len(ruleException.Destination.GlobalIPRange) > 0 {
				destAttrs["global_ip_range"] = parseNameIDList(ctx, ruleException.Destination.GlobalIPRange, "rule.exception.destination.global_ip_range")
			}

			// If no destination fields have data, create an empty object
			if len(destAttrs) == 0 {
				// Create empty destination object with all null fields to match schema
				destAttrs = map[string]attr.Value{
					"application":              types.SetNull(NameIDObjectType),
					"custom_app":               types.SetNull(NameIDObjectType),
					"app_category":             types.SetNull(NameIDObjectType),
					"custom_category":          types.SetNull(NameIDObjectType),
					"sanctioned_apps_category": types.SetNull(NameIDObjectType),
					"country":                  types.SetNull(NameIDObjectType),
					"domain":                   types.ListNull(types.StringType),
					"fqdn":                     types.ListNull(types.StringType),
					"ip":                       types.ListNull(types.StringType),
					"subnet":                   types.ListNull(types.StringType),
					"ip_range":                 types.ListNull(FromToObjectType),
					"global_ip_range":          types.SetNull(NameIDObjectType),
					"remote_asn":               types.ListNull(types.StringType),
				}
			} else {
				// Fill in missing fields with null values to complete the schema
				schemaFields := []string{"application", "custom_app", "app_category", "custom_category", "sanctioned_apps_category", "country", "domain", "fqdn", "ip", "subnet", "ip_range", "global_ip_range", "remote_asn"}
				for _, field := range schemaFields {
					if _, exists := destAttrs[field]; !exists {
						switch field {
						case "domain", "fqdn", "ip", "subnet", "remote_asn":
							destAttrs[field] = types.ListNull(types.StringType)
						case "ip_range":
							destAttrs[field] = types.ListNull(FromToObjectType)
						default:
							destAttrs[field] = types.SetNull(NameIDObjectType)
						}
					}
				}
			}

			curExceptionDestObj, diagstmp := types.ObjectValue(
				IfwDestAttrTypes,
				destAttrs,
			)
			diags = append(diags, diagstmp...)

			////////////// start Rule -> Service ///////////////
			// Build service object using conditional fields to match config logic
			serviceAttrs := map[string]attr.Value{}

			// Only include fields that have data to match config filtering logic (: k => v if v != null)
			if len(ruleException.Service.Standard) > 0 {
				serviceAttrs["standard"] = parseNameIDList(ctx, ruleException.Service.Standard, "rule.exception.service.standard")
			}

			if ruleException.Service.Custom != nil && len(ruleException.Service.Custom) > 0 {
				var curExceptionCustomServices []types.Object
				tflog.Info(ctx, "ruleException.Service.Custom - "+fmt.Sprintf("%v", ruleException.Service.Custom))
				for _, item := range ruleException.Service.Custom {
					curExceptionCustomServices = append(curExceptionCustomServices, parseCustomService(ctx, item, "rule.exception.service.custom"))
				}
				serviceAttrs["custom"], diagstmp = types.ListValueFrom(ctx, CustomServiceObjectType, curExceptionCustomServices)
				diags = append(diags, diagstmp...)
			}

			// If no service fields have data, create an empty object
			if len(serviceAttrs) == 0 {
				serviceAttrs = map[string]attr.Value{
					"standard": types.SetNull(NameIDObjectType),
					"custom":   types.ListNull(CustomServiceObjectType),
				}
			} else {
				// Fill in missing fields with null values to complete the schema
				if _, exists := serviceAttrs["standard"]; !exists {
					serviceAttrs["standard"] = types.SetNull(NameIDObjectType)
				}
				if _, exists := serviceAttrs["custom"]; !exists {
					serviceAttrs["custom"] = types.ListNull(CustomServiceObjectType)
				}
			}

			curExceptionServiceObj, diagstmp := types.ObjectValue(IfwServiceAttrTypes, serviceAttrs)
			diags = append(diags, diagstmp...)
			////////////// end Rule -> Service ///////////////

			// Initialize Exception object with populated values
			// Build exception attributes to match config logic
			exceptionAttrs := map[string]attr.Value{
				"name":        types.StringValue(ruleException.Name),
				"source":      curExceptionSourceObj,
				"destination": curExceptionDestObj,
				"service":     curExceptionServiceObj,
				"device_os": func() types.List {
					// Handle device_os to match planned state: empty array [] should be empty list, not null
					if ruleException.DeviceOs == nil {
						return types.ListNull(types.StringType)
					}
					// Create empty list for empty array to match planned state
					emptyList, _ := types.ListValueFrom(ctx, types.StringType, ruleException.DeviceOs)
					return emptyList
				}(),
			}

			// Only include optional fields if they have data (matching config conditional logic)
			if ruleException.ConnectionOrigin.String() != "" {
				exceptionAttrs["connection_origin"] = types.StringValue(ruleException.ConnectionOrigin.String())
			}
			if len(ruleException.Country) > 0 {
				exceptionAttrs["country"] = parseNameIDList(ctx, ruleException.Country, "rule.exception.country")
			}
			if len(ruleException.Device) > 0 {
				exceptionAttrs["device"] = parseNameIDList(ctx, ruleException.Device, "rule.exception.device")
			}

			// Fill in missing required fields with null values to complete the schema
			if _, exists := exceptionAttrs["connection_origin"]; !exists {
				exceptionAttrs["connection_origin"] = types.StringNull()
			}
			if _, exists := exceptionAttrs["country"]; !exists {
				exceptionAttrs["country"] = types.SetNull(types.ObjectType{AttrTypes: NameIDAttrTypes})
			}
			if _, exists := exceptionAttrs["device"]; !exists {
				exceptionAttrs["device"] = types.SetNull(types.ObjectType{AttrTypes: NameIDAttrTypes})
			}

			curException, diagstmp := types.ObjectValue(
				IfwExceptionAttrTypes,
				exceptionAttrs,
			)
			diags = append(diags, diagstmp...)
			exceptions = append(exceptions, curException)
		}
		curRuleExceptionsObj, diagstmp := types.SetValue(types.ObjectType{AttrTypes: IfwExceptionAttrTypes}, exceptions)
		diags = append(diags, diagstmp...)
		ruleInput.Exceptions = curRuleExceptionsObj
	} else {
		ruleInput.Exceptions = types.SetNull(types.ObjectType{AttrTypes: IfwExceptionAttrTypes})
	}
	////////////// end Rule -> Exceptions ///////////////

	return ruleInput

}
