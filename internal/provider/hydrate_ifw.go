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
	"github.com/hashicorp/terraform-plugin-framework/resource"
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
		SourceAttrTypes,
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
		DestAttrTypes,
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
			ServiceAttrTypes,
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
	curRuleExceptionsObj, diagstmp := types.ListValue(
		types.ObjectType{AttrTypes: ExceptionAttrTypes},
		[]attr.Value{
			types.ObjectValueMust( // Single exception object with all null values
				ExceptionAttrTypes,
				map[string]attr.Value{
					"name":    types.StringNull(),
					"source":  types.ObjectNull(SourceAttrTypes),
					"country": types.SetNull(types.ObjectType{AttrTypes: NameIDAttrTypes}),
					"device":  types.SetNull(types.ObjectType{AttrTypes: NameIDAttrTypes}),
					// "device_attributes": types.ObjectNull(DeviceAttrAttrTypes),
					"device_os":         types.ListNull(types.StringType),
					"destination":       types.ObjectNull(DestAttrTypes),
					"service":           types.ObjectNull(ServiceAttrTypes),
					"connection_origin": types.StringNull(),
				},
			),
		},
	)
	diags = append(diags, diagstmp...)
	exceptions := []attr.Value{}

	// Rule -> Exceptions -> Source
	if currentRule.Exceptions != nil && len(currentRule.Exceptions) > 0 {
		for _, ruleException := range currentRule.Exceptions {
			// Rule -> Exceptions -> Source
			curExceptionSourceObj, diags := types.ObjectValue(
				SourceAttrTypes,
				map[string]attr.Value{
					"ip": parseList(ctx, types.StringType, ruleException.Source.IP, "rule.exception.source.ip"),
					// "subnet":              parseList(ctx, types.StringType, ruleException.Source.Subnet),
					"subnet":              types.ListNull(types.StringType),
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
				DestAttrTypes,
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
				ServiceAttrTypes,
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
				ExceptionAttrTypes,
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
		curRuleExceptionsObj, diagstmp = types.ListValue(types.ObjectType{AttrTypes: ExceptionAttrTypes}, exceptions)
		diags = append(diags, diagstmp...)
		ruleInput.Exceptions = curRuleExceptionsObj
	} else {
		ruleInput.Exceptions = curRuleExceptionsObj
	}
	////////////// end Rule -> Exceptions ///////////////

	return ruleInput

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

func parseList[T any](ctx context.Context, elemType attr.Type, items []T, attrName string) types.List {
	tflog.Warn(ctx, "parseList() "+attrName+" - "+fmt.Sprintf("%v", items))
	diags := make(diag.Diagnostics, 0)

	// if stateItems != nil && len(stateItems) > 0 {
	// 	tflog.Info(ctx, "parseList() - empty input list found in state, returning empty array")
	// 	return types.ListValueMust(elemType, []attr.Value{})
	// } else
	if items == nil || len(items) == 0 {
		tflog.Info(ctx, "parseList() - nil or empty input list, returning null")
		return types.ListNull(elemType)
	}
	// if len(items) == 0 {
	// 	tflog.Info(ctx, "parseList() - empty input list")
	// 	return types.ListValueMust(elemType, []attr.Value{})
	// }

	tflog.Info(ctx, "parseList() - "+fmt.Sprintf("%v", items))

	// Convert to types.List using ListValueFrom
	listValue, listDiags := types.ListValueFrom(ctx, elemType, items)
	diags.Append(listDiags...)
	return listValue
}

func parseNameIDList[T any](ctx context.Context, items []T, attrName string) types.Set {
	tflog.Warn(ctx, "parseNameIDList() "+attrName+" - "+fmt.Sprintf("%v", items))
	diags := make(diag.Diagnostics, 0)

	// Handle nil or empty list
	if items == nil || len(items) == 0 {
		tflog.Warn(ctx, "parseNameIDList() - nil or empty input list")
		return types.SetNull(NameIDObjectType)
	}

	// Process each item into an attr.Value
	nameIDValues := make([]attr.Value, 0, len(items))
	for i, item := range items {
		obj := parseNameID(ctx, item, attrName)
		if !obj.IsNull() && !obj.IsUnknown() { // Include only non-null/unknown values, adjust as needed
			nameIDValues = append(nameIDValues, obj)
		} else {
			tflog.Warn(ctx, "parseNameIDList() - skipping null/unknown item at index "+fmt.Sprintf("%d", i))
		}
	}

	// Convert to types.List using SetValueFrom
	setValue, diagstmp := types.SetValueFrom(ctx, NameIDObjectType, nameIDValues)
	diags = append(diags, diagstmp...)
	return setValue
}

func parseNameID(ctx context.Context, item interface{}, attrName string) types.Object {
	tflog.Warn(ctx, "parseNameID() "+attrName+" - "+fmt.Sprintf("%v", item))
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

func parseFromToList[T any](ctx context.Context, items []T, attrName string) types.List {
	tflog.Warn(ctx, "parseFromToList() "+attrName+" - "+fmt.Sprintf("%v", items))
	diags := make(diag.Diagnostics, 0)

	// Handle nil or empty list
	if items == nil || len(items) == 0 {
		tflog.Warn(ctx, "parseFromToList() - nil or empty input list")
		return types.ListNull(FromToObjectType)
	}
	// Process each item into an attr.Value
	fromToValues := make([]attr.Value, 0, len(items))
	for i, item := range items {
		obj := parseFromTo(ctx, item, attrName)
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

func parseFromTo(ctx context.Context, item interface{}, attrName string) types.Object {
	tflog.Warn(ctx, "parseFromTo() "+attrName+" - "+fmt.Sprintf("%v", item))
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

func parseFromToDays(ctx context.Context, item interface{}, attrName string) types.Object {
	tflog.Warn(ctx, "parseFromToDays() "+attrName+" - "+fmt.Sprintf("%v", item))
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
			"days": parseList(ctx, types.StringType, daysField.Interface().([]cato_models.DayOfWeek), "rule.schedule.custom_recurring.days"),
		},
	)
	tflog.Warn(ctx, "parseFromToDays() obj - "+fmt.Sprintf("%v", obj))
	diags = append(diags, diagstmp...)
	return obj
}

func parseCustomService(ctx context.Context, item interface{}, attrName string) types.Object {
	tflog.Warn(ctx, "parseCustomService() "+attrName+" - "+fmt.Sprintf("%v", item))
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

	tflog.Warn(ctx, "parseCustomService() portField - "+fmt.Sprintf("%v", portField))
	tflog.Warn(ctx, "parseCustomService() protocolField - "+fmt.Sprintf("%v", protocolField))
	tflog.Warn(ctx, "parseCustomService() portRangeField - "+fmt.Sprintf("%v", portRangeField))
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
			// ports[i] = types.StringNull()
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
