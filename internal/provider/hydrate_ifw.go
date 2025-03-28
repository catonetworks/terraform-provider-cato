package provider

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models" // Import the correct package
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
	// // Rule -> Index
	// ruleInput.Index = types.StringValue(currentRule.Index.String())
	// Rule -> ConnectionOrigin
	ruleInput.ConnectionOrigin = types.StringValue(currentRule.ConnectionOrigin.String())

	// //////////// Rule -> Source ///////////////
	curRuleSourceObj, diags := types.ObjectValue(
		SourceAttrTypes,
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

	curRuleSourceObjAttrs["ip"] = parseList(ctx, types.StringType, currentRule.Source.IP)
	curRuleSourceObjAttrs["subnet"] = parseList(ctx, types.StringType, currentRule.Source.Subnet)
	curRuleSourceObjAttrs["host"] = parseNameIDList(ctx, currentRule.Source.Host)
	curRuleSourceObjAttrs["site"] = parseNameIDList(ctx, currentRule.Source.Site)
	curRuleSourceObjAttrs["ip_range"] = parseFromToList(ctx, currentRule.Source.IPRange)
	curRuleSourceObjAttrs["global_ip_range"] = parseNameIDList(ctx, currentRule.Source.GlobalIPRange)
	curRuleSourceObjAttrs["network_interface"] = parseNameIDList(ctx, currentRule.Source.NetworkInterface)
	curRuleSourceObjAttrs["site_network_subnet"] = parseNameIDList(ctx, currentRule.Source.SiteNetworkSubnet)
	curRuleSourceObjAttrs["floating_subnet"] = parseNameIDList(ctx, currentRule.Source.FloatingSubnet)
	curRuleSourceObjAttrs["user"] = parseNameIDList(ctx, currentRule.Source.User)
	curRuleSourceObjAttrs["users_group"] = parseNameIDList(ctx, currentRule.Source.UsersGroup)
	curRuleSourceObjAttrs["group"] = parseNameIDList(ctx, currentRule.Source.Group)
	curRuleSourceObjAttrs["system_group"] = parseNameIDList(ctx, currentRule.Source.SystemGroup)

	curRuleSourceObj, diags = types.ObjectValue(curRuleSourceObj.AttributeTypes(ctx), curRuleSourceObjAttrs)
	resp.Diagnostics.Append(diags...)
	ruleInput.Source = curRuleSourceObj
	////////////// end rule.source ///////////////

	// Rule -> Country
	if currentRule.Country != nil {
		if len(currentRule.Country) > 0 {
			var curSourceCountries []types.Object
			tflog.Info(ctx, "ruleResponse.Country - "+fmt.Sprintf("%v", currentRule.Country))
			for _, item := range currentRule.Country {
				curSourceCountries = append(curSourceCountries, parseNameID(ctx, item))
			}
			ruleInput.Country, diags = types.ListValueFrom(ctx, NameIDObjectType, curSourceCountries)
			resp.Diagnostics.Append(diags...)
		}
	}

	// Rule -> Device
	if currentRule.Device != nil {
		if len(currentRule.Device) > 0 {
			var curSourceDevices []types.Object
			tflog.Info(ctx, "ruleResponse.Device - "+fmt.Sprintf("%v", currentRule.Device))
			for _, item := range currentRule.Device {
				curSourceDevices = append(curSourceDevices, parseNameID(ctx, item))
			}
			ruleInput.Device, diags = types.ListValueFrom(ctx, NameIDObjectType, curSourceDevices)
			resp.Diagnostics.Append(diags...)
		}
	}

	// Rule -> DeviceOS
	ruleInput.DeviceOs, diags = types.ListValueFrom(ctx, types.StringType, currentRule.DeviceOs)
	resp.Diagnostics.Append(diags...)

	//////////// Rule -> Destination ///////////////
	curRuleDestinationObj, diags := types.ObjectValue(
		DestAttrTypes,
		map[string]attr.Value{
			"application":              types.ListNull(NameIDObjectType),
			"custom_app":               types.ListNull(NameIDObjectType),
			"app_category":             types.ListNull(NameIDObjectType),
			"custom_category":          types.ListNull(NameIDObjectType),
			"sanctioned_apps_category": types.ListNull(NameIDObjectType),
			"country":                  types.ListNull(NameIDObjectType),
			"domain":                   types.ListNull(types.StringType),
			"fqdn":                     types.ListNull(types.StringType),
			"ip":                       types.ListNull(types.StringType),
			"subnet":                   types.ListNull(types.StringType),
			"ip_range":                 types.ListNull(FromToObjectType),
			"global_ip_range":          types.ListNull(NameIDObjectType),
			"remote_asn":               types.ListNull(types.StringType),
		},
	)
	resp.Diagnostics.Append(diags...)
	curRuleDestinationObjAttrs := curRuleDestinationObj.Attributes()

	curRuleDestinationObjAttrs["ip"] = parseList(ctx, types.StringType, currentRule.Destination.IP)
	curRuleDestinationObjAttrs["subnet"] = parseList(ctx, types.StringType, currentRule.Destination.Subnet)
	curRuleDestinationObjAttrs["domain"] = parseList(ctx, types.StringType, currentRule.Destination.Domain)
	curRuleDestinationObjAttrs["fqdn"] = parseList(ctx, types.StringType, currentRule.Destination.Fqdn)
	curRuleDestinationObjAttrs["remote_asn"] = parseList(ctx, types.StringType, currentRule.Destination.RemoteAsn)
	curRuleDestinationObjAttrs["application"] = parseNameIDList(ctx, currentRule.Destination.Application)
	curRuleDestinationObjAttrs["custom_app"] = parseNameIDList(ctx, currentRule.Destination.CustomApp)
	curRuleDestinationObjAttrs["ip_range"] = parseFromToList(ctx, currentRule.Destination.IPRange)
	curRuleDestinationObjAttrs["global_ip_range"] = parseNameIDList(ctx, currentRule.Destination.GlobalIPRange)
	curRuleDestinationObjAttrs["app_category"] = parseNameIDList(ctx, currentRule.Destination.AppCategory)
	curRuleDestinationObjAttrs["custom_category"] = parseNameIDList(ctx, currentRule.Destination.CustomCategory)
	curRuleDestinationObjAttrs["sanctioned_apps_category"] = parseNameIDList(ctx, currentRule.Destination.SanctionedAppsCategory)
	curRuleDestinationObjAttrs["country"] = parseNameIDList(ctx, currentRule.Destination.Country)

	curRuleDestinationObj, diags = types.ObjectValue(curRuleDestinationObj.AttributeTypes(ctx), curRuleDestinationObjAttrs)
	resp.Diagnostics.Append(diags...)
	ruleInput.Destination = curRuleDestinationObj
	////////////// end Rule -> Destination ///////////////

	////////////// start Rule -> Service ///////////////
	if len(currentRule.Service.Custom) > 0 || len(currentRule.Service.Standard) > 0 {
		var serviceInput *Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service
		diags = ruleInput.Service.As(ctx, &serviceInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)

		// Initialize Service object with null values
		curRuleServiceObj, diags := types.ObjectValue(
			ServiceAttrTypes,
			map[string]attr.Value{
				"standard": types.ListNull(NameIDObjectType),
				"custom":   types.ListNull(CustomServiceObjectType),
			},
		)
		resp.Diagnostics.Append(diags...)
		curRuleServiceObjAttrs := curRuleServiceObj.Attributes()

		// Rule -> Service -> Standard
		if currentRule.Service.Standard != nil {
			if len(currentRule.Service.Standard) > 0 {
				var curRuleStandardServices []types.Object
				tflog.Info(ctx, "ruleResponse.Service.Standard - "+fmt.Sprintf("%v", currentRule.Service.Standard))
				for _, item := range currentRule.Service.Standard {
					curRuleStandardServices = append(curRuleStandardServices, parseNameID(ctx, item))
				}
				curRuleServiceObjAttrs["standard"], diags = types.ListValueFrom(ctx, NameIDObjectType, curRuleStandardServices)
				resp.Diagnostics.Append(diags...)
			}
		}

		// Rule -> Service -> Custom
		if currentRule.Service.Custom != nil {
			if len(currentRule.Service.Custom) > 0 {
				var curRuleCustomServices []types.Object
				tflog.Info(ctx, "ruleResponse.Service.Custom - "+fmt.Sprintf("%v", currentRule.Service.Custom))
				for _, item := range currentRule.Service.Custom {
					curRuleCustomServices = append(curRuleCustomServices, parseCustomService(ctx, item))
				}
				curRuleServiceObjAttrs["custom"], diags = types.ListValueFrom(ctx, CustomServiceObjectType, curRuleCustomServices)
				resp.Diagnostics.Append(diags...)
			}
		}

		curRuleServiceObj, diags = types.ObjectValue(curRuleServiceObj.AttributeTypes(ctx), curRuleServiceObjAttrs)
		resp.Diagnostics.Append(diags...)
		ruleInput.Service = curRuleServiceObj
	}
	////////////// end Rule -> Service ///////////////

	////////////// start Rule -> Tracking ///////////////
	curRuleTrackingObj, diags := types.ObjectValue(
		TrackingAttrTypes,
		map[string]attr.Value{
			"event": types.ObjectNull(TrackingEventAttrTypes),
			"alert": types.ObjectNull(TrackingAlertAttrTypes),
		},
	)
	resp.Diagnostics.Append(diags...)
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
	trackingAlertValue, diags := types.ObjectValue(
		TrackingAlertAttrTypes,
		map[string]attr.Value{
			"enabled":            types.BoolValue(currentRule.Tracking.Alert.Enabled),
			"frequency":          types.StringValue(currentRule.Tracking.Alert.Frequency.String()),
			"subscription_group": parseNameIDList(ctx, currentRule.Tracking.Alert.SubscriptionGroup),
			"webhook":            parseNameIDList(ctx, currentRule.Tracking.Alert.Webhook),
			"mailing_list":       parseNameIDList(ctx, currentRule.Tracking.Alert.MailingList),
		},
	)
	curRuleTrackingObjAttrs["alert"] = trackingAlertValue
	tflog.Warn(ctx, "Updated tracking object: "+fmt.Sprintf("%v", curRuleTrackingObj))

	curRuleTrackingObj, diags = types.ObjectValue(curRuleTrackingObj.AttributeTypes(ctx), curRuleTrackingObjAttrs)
	resp.Diagnostics.Append(diags...)
	ruleInput.Tracking = curRuleTrackingObj
	////////////// end Rule -> Tracking ///////////////

	////////////// start Rule -> Schedule ///////////////
	curRuleScheduleObj, diags := types.ObjectValue(
		ScheduleAttrTypes,
		map[string]attr.Value{
			"active_on":        types.StringValue(currentRule.Schedule.ActiveOn.String()),
			"custom_timeframe": parseFromTo(ctx, currentRule.Schedule.CustomTimeframePolicySchedule),
			"custom_recurring": parseFromToDays(ctx, currentRule.Schedule.CustomRecurringPolicySchedule),
		},
	)
	resp.Diagnostics.Append(diags...)
	ruleInput.Schedule = curRuleScheduleObj
	////////////// end Rule -> Schedule ///////////////

	// ////////////// start Rule -> Exceptions ///////////////
	// curRuleExceptionsObj, diags := types.ListValue(
	// 	types.ObjectType{AttrTypes: ExceptionAttrTypes},
	// 	[]attr.Value{
	// 		types.ObjectValueMust( // Single exception object with all null values
	// 			ExceptionAttrTypes,
	// 			map[string]attr.Value{
	// 				"name":              types.StringNull(),
	// 				"source":            types.ObjectNull(SourceAttrTypes),
	// 				"country":           types.ListNull(types.ObjectType{AttrTypes: NameIDAttrTypes}),
	// 				"device":            types.ListNull(types.ObjectType{AttrTypes: NameIDAttrTypes}),
	// 				"device_attributes": types.ObjectNull(DeviceAttrAttrTypes),
	// 				"device_os":         types.ListNull(types.StringType),
	// 				"destination":       types.ObjectNull(DestAttrTypes),
	// 				"service":           types.ObjectNull(ServiceAttrTypes),
	// 				"connection_origin": types.StringNull(),
	// 			},
	// 		),
	// 	},
	// )
	// resp.Diagnostics.Append(diags...)
	// ruleInput.Exceptions = curRuleExceptionsObj

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

func parseList[T any](ctx context.Context, elemType attr.Type, items []T) types.List {
	tflog.Warn(ctx, "parseList() - "+fmt.Sprintf("%v", items))
	diags := make(diag.Diagnostics, 0)
	// Handle nil or empty list
	if items == nil || len(items) == 0 {
		tflog.Info(ctx, "toListValue() - nil or empty input list")
		return types.ListNull(elemType)
	}

	tflog.Info(ctx, "toListValue() - "+fmt.Sprintf("%v", items))

	// Convert to types.List using ListValueFrom
	listValue, listDiags := types.ListValueFrom(ctx, elemType, items)
	diags.Append(listDiags...)
	return listValue
}

func parseNameIDList[T any](ctx context.Context, items []T) types.List {
	tflog.Warn(ctx, "parseNameIDList() - "+fmt.Sprintf("%v", items))
	diags := make(diag.Diagnostics, 0)

	// Handle nil or empty list
	if items == nil || len(items) == 0 {
		tflog.Warn(ctx, "parseNameIDList() - nil or empty input list")
		return types.ListNull(NameIDObjectType)
	}

	// Process each item into an attr.Value
	nameIDValues := make([]attr.Value, 0, len(items))
	for i, item := range items {
		obj := parseNameID(ctx, item)
		if !obj.IsNull() && !obj.IsUnknown() { // Include only non-null/unknown values, adjust as needed
			nameIDValues = append(nameIDValues, obj)
		} else {
			tflog.Warn(ctx, "parseNameIDList() - skipping null/unknown item at index "+fmt.Sprintf("%d", i))
		}
	}

	// Convert to types.List using ListValueFrom
	listValue, diagstmp := types.ListValueFrom(ctx, NameIDObjectType, nameIDValues)
	diags = append(diags, diagstmp...)
	return listValue
}

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
		tflog.Warn(ctx, "parseNameID() nameField.IsValid() - "+fmt.Sprintf("%v", nameField))
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
	tflog.Warn(ctx, "parseNameID() obj - "+fmt.Sprintf("%v", obj))
	diags = append(diags, diagstmp...)
	return obj
}

func parseFromToList[T any](ctx context.Context, items []T) types.List {
	tflog.Warn(ctx, "parseFromToList() - "+fmt.Sprintf("%v", items))
	diags := make(diag.Diagnostics, 0)

	// Handle nil or empty list
	if items == nil || len(items) == 0 {
		tflog.Warn(ctx, "parseFromToList() - nil or empty input list")
		return types.ListNull(FromToObjectType)
	}
	// Process each item into an attr.Value
	fromToValues := make([]attr.Value, 0, len(items))
	for i, item := range items {
		obj := parseFromTo(ctx, item)
		if !obj.IsNull() && !obj.IsUnknown() { // Include only non-null/unknown values, adjust as needed
			fromToValues = append(fromToValues, obj)
		} else {
			tflog.Warn(ctx, "parseFromToList() - skipping null/unknown item at index "+fmt.Sprintf("%d", i))
		}
	}

	// Convert to types.List using ListValueFrom
	listValue, diagstmp := types.ListValueFrom(ctx, FromToObjectType, fromToValues)
	diags = append(diags, diagstmp...)
	return listValue
}

func parseFromTo(ctx context.Context, item interface{}) types.Object {
	tflog.Warn(ctx, "parseFromTo() - "+fmt.Sprintf("%v", item))
	diags := make(diag.Diagnostics, 0)

	if item == nil {
		tflog.Warn(ctx, "parseFromTo() - nil")
		return types.ObjectNull(FromToAttrTypes)
	}

	// Get the reflect.Value of the input
	itemValue := reflect.ValueOf(item)

	// Handle nil or invalid input
	tflog.Warn(ctx, "parseFromTo() itemValue.Kind()- "+fmt.Sprintf("%v", itemValue.Kind()))
	if item == nil || itemValue.Kind() != reflect.Struct {
		if itemValue.Kind() == reflect.Ptr && !itemValue.IsNil() {
			itemValue = itemValue.Elem()
			// Keep dereferencing until we get to a struct or can't anymore
			for itemValue.Kind() == reflect.Ptr && !itemValue.IsNil() {
				itemValue = itemValue.Elem()
			}
			if itemValue.Kind() != reflect.Struct {
				return types.ObjectNull(FromToAttrTypes)
			}
		} else {
			return types.ObjectNull(FromToAttrTypes)
		}
	}

	// Handle pointer to struct
	if itemValue.Kind() == reflect.Ptr {
		if itemValue.IsNil() {
			return types.ObjectNull(FromToAttrTypes)
		}
		itemValue = itemValue.Elem()
	}

	// Get Name and ID fields
	fromField := itemValue.FieldByName("From")
	toField := itemValue.FieldByName("To")

	if !fromField.IsValid() || !toField.IsValid() {
		tflog.Warn(ctx, "parseFromTo() fromField.IsValid() - "+fmt.Sprintf("%v", fromField))
		tflog.Warn(ctx, "parseFromTo() toField.IsValid() - "+fmt.Sprintf("%v", toField))
		return types.ObjectNull(FromToAttrTypes)
	}

	// Create object value
	obj, diagstmp := types.ObjectValue(
		FromToAttrTypes,
		map[string]attr.Value{
			"from": basetypes.NewStringValue(fromField.String()),
			"to":   basetypes.NewStringValue(toField.String()),
		},
	)
	tflog.Warn(ctx, "parseFromTo() obj - "+fmt.Sprintf("%v", obj))
	diags = append(diags, diagstmp...)
	return obj
}

func parseFromToDays(ctx context.Context, item interface{}) types.Object {
	tflog.Warn(ctx, "parseFromToDays() - "+fmt.Sprintf("%v", item))
	diags := make(diag.Diagnostics, 0)

	if item == nil {
		tflog.Warn(ctx, "parseFromToDays() - nil")
		return types.ObjectNull(FromToAttrTypes)
	}

	// Get the reflect.Value of the input
	itemValue := reflect.ValueOf(item)

	// Handle nil or invalid input
	tflog.Warn(ctx, "parseFromToDays() itemValue.Kind()- "+fmt.Sprintf("%v", itemValue.Kind()))
	if item == nil || itemValue.Kind() != reflect.Struct {
		if itemValue.Kind() == reflect.Ptr && !itemValue.IsNil() {
			itemValue = itemValue.Elem()
			// Keep dereferencing until we get to a struct or can't anymore
			for itemValue.Kind() == reflect.Ptr && !itemValue.IsNil() {
				itemValue = itemValue.Elem()
			}
			if itemValue.Kind() != reflect.Struct {
				return types.ObjectNull(FromToDaysAttrTypes)
			}
		} else {
			return types.ObjectNull(FromToDaysAttrTypes)
		}
	}

	// Handle pointer to struct
	if itemValue.Kind() == reflect.Ptr {
		if itemValue.IsNil() {
			return types.ObjectNull(FromToDaysAttrTypes)
		}
		itemValue = itemValue.Elem()
	}

	// Get Name and ID fields
	fromField := itemValue.FieldByName("From")
	toField := itemValue.FieldByName("To")
	daysField := itemValue.FieldByName("Days")

	if !fromField.IsValid() || !toField.IsValid() {
		tflog.Warn(ctx, "parseFromTo() fromField.IsValid() - "+fmt.Sprintf("%v", fromField))
		tflog.Warn(ctx, "parseFromTo() toField.IsValid() - "+fmt.Sprintf("%v", toField))
		return types.ObjectNull(FromToDaysAttrTypes)
	}

	// Create object value
	obj, diagstmp := types.ObjectValue(
		FromToDaysAttrTypes,
		map[string]attr.Value{
			"from": basetypes.NewStringValue(fromField.String()),
			"to":   basetypes.NewStringValue(toField.String()),
			"days": parseList(ctx, types.StringType, daysField.Interface().([]cato_models.DayOfWeek)),
		},
	)
	tflog.Warn(ctx, "parseFromToDays() obj - "+fmt.Sprintf("%v", obj))
	diags = append(diags, diagstmp...)
	return obj
}

func parseCustomService(ctx context.Context, item interface{}) types.Object {
	tflog.Warn(ctx, "parseCustomService() - "+fmt.Sprintf("%v", item))
	diags := make(diag.Diagnostics, 0)

	// Get the reflect.Value of the input
	itemValue := reflect.ValueOf(item)

	// Handle nil or invalid input
	if item == nil || itemValue.Kind() != reflect.Struct {
		if itemValue.Kind() == reflect.Ptr && !itemValue.IsNil() {
			itemValue = itemValue.Elem()
			// Keep dereferencing until we get to a struct or can't anymore
			for itemValue.Kind() == reflect.Ptr && !itemValue.IsNil() {
				itemValue = itemValue.Elem()
			}
			if itemValue.Kind() != reflect.Struct {
				return types.ObjectNull(CustomServiceAttrTypes)
			}
		} else {
			return types.ObjectNull(CustomServiceAttrTypes)
		}
	}

	// Handle pointer to struct
	if itemValue.Kind() == reflect.Ptr {
		tflog.Warn(ctx, "parseCustomService() itemValue.Kind()- "+fmt.Sprintf("%v", itemValue))
		if itemValue.IsNil() {
			return types.ObjectNull(CustomServiceAttrTypes)
		}
		itemValue = itemValue.Elem()
		tflog.Warn(ctx, "parseCustomService() itemValue.Elem()- "+fmt.Sprintf("%v", itemValue))
	}

	// Get fields
	portField := itemValue.FieldByName("Port")
	protocolField := itemValue.FieldByName("Protocol")
	portRangeField := itemValue.FieldByName("PortRange")

	tflog.Warn(ctx, "parseCustomService() protocolField- "+fmt.Sprintf("%v", protocolField))
	// Handle port field (allowing null)
	var portList types.List
	if portField.IsValid() && portField.Kind() == reflect.Slice {
		ports := make([]attr.Value, portField.Len())
		for i := range portField.Len() {
			portValue := portField.Index(i)
			var portStr string
			switch portValue.Kind() {
			case reflect.String:
				portStr = portValue.String()
			case reflect.Int, reflect.Int64:
				portStr = fmt.Sprintf("%d", portValue.Int())
			default:
				tflog.Warn(ctx, "parseCustomService() unsupported port type - "+fmt.Sprintf("%v", portValue.Kind()))
				portStr = fmt.Sprintf("%v", portValue.Interface())
			}
			ports[i] = types.StringValue(portStr)
		}
		var diagsTmp diag.Diagnostics
		portList, diagsTmp = types.ListValue(types.StringType, ports)
		diags = append(diags, diagsTmp...)
	} else {
		portList = types.ListNull(types.StringType) // Explicit null handling
	}

	// Handle protocol
	var protocolVal types.String
	if protocolField.IsValid() {
		protocolVal = types.StringValue(protocolField.String())
	} else {
		protocolVal = types.StringNull()
	}

	// Handle port_range
	var portRangeVal types.Object
	if portRangeField.Kind() == reflect.Ptr {
		if portRangeField.IsNil() {
			portRangeVal = types.ObjectNull(FromToAttrTypes)
		}
		portRangeField = portRangeField.Elem()
	}
	if portRangeField.IsValid() {
		from := portRangeField.FieldByName("From")
		to := portRangeField.FieldByName("To")
		var diagsTmp diag.Diagnostics
		portRangeVal, diagsTmp = types.ObjectValue(
			FromToAttrTypes,
			map[string]attr.Value{
				"from": types.StringValue(from.String()),
				"to":   types.StringValue(to.String()),
			},
		)
		diags = append(diags, diagsTmp...)
	} else {
		portRangeVal = types.ObjectNull(FromToAttrTypes)
	}

	// Create final custom service object
	obj, diagstmp := types.ObjectValue(
		CustomServiceAttrTypes,
		map[string]attr.Value{
			"port":       portList,
			"port_range": portRangeVal,
			"protocol":   protocolVal,
		},
	)
	tflog.Warn(ctx, "parseCustomService() obj - "+fmt.Sprintf("%v", obj))
	diags = append(diags, diagstmp...)
	return obj
}
