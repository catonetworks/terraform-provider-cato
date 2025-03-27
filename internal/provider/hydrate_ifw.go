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
	// // Rule -> Index
	// ruleInput.Index = types.StringValue(currentRule.Index.String())
	// Rule -> ConnectionOrigin
	ruleInput.ConnectionOrigin = types.StringValue(currentRule.ConnectionOrigin.String())

	// //////////// Rule -> Source ///////////////
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

	// // Rule -> Source -> IP
	if currentRule.Source.IP != nil {
		if len(currentRule.Source.IP) > 0 {
			tflog.Info(ctx, "ruleResponse.Source.IP - "+fmt.Sprintf("%v", currentRule.Source.IP))
			curSourceSourceIpList, diagstmp := types.ListValueFrom(ctx, types.StringType, currentRule.Source.IP)
			diags = append(diags, diagstmp...)
			curRuleSourceObjAttrs["ip"] = curSourceSourceIpList
		}
	}

	// Rule -> Source -> Subnet
	if currentRule.Source.Subnet != nil {
		if len(currentRule.Source.Subnet) > 0 {
			tflog.Info(ctx, "ruleResponse.Source.Subnet - "+fmt.Sprintf("%v", currentRule.Source.Subnet))
			curRuleSourceSubnetList, diagstmp := types.ListValueFrom(ctx, types.StringType, currentRule.Source.Subnet)
			resp.Diagnostics.Append(diagstmp...)
			curRuleSourceObjAttrs["subnet"] = curRuleSourceSubnetList
		}
	}

	// Rule -> Source -> Host
	if currentRule.Source.Host != nil {
		if len(currentRule.Source.Host) > 0 {
			var curSourceHosts []types.Object
			tflog.Info(ctx, "ruleResponse.Source.Host - "+fmt.Sprintf("%v", currentRule.Source.Host))
			for _, item := range currentRule.Source.Host {
				curSourceHosts = append(curSourceHosts, parseNameID(ctx, item))
			}
			curRuleSourceObjAttrs["host"], diags = types.ListValueFrom(ctx, NameIDObjectType, curSourceHosts)
			resp.Diagnostics.Append(diags...)
		}
	}

	// Rule -> Source -> Site
	if currentRule.Source.Site != nil {
		if len(currentRule.Source.Site) > 0 {
			var curSourceSites []types.Object
			tflog.Info(ctx, "ruleResponse.Source.Site - "+fmt.Sprintf("%v", currentRule.Source.Site))
			for _, item := range currentRule.Source.Site {
				curSourceSites = append(curSourceSites, parseNameID(ctx, item))
			}
			curRuleSourceObjAttrs["site"], diags = types.ListValueFrom(ctx, NameIDObjectType, curSourceSites)
			resp.Diagnostics.Append(diags...)
		}
	}

	// Rule -> Source -> IPRange
	if currentRule.Source.IPRange != nil {
		if len(currentRule.Source.IPRange) > 0 {
			var curSourceIPRanges []types.Object
			tflog.Info(ctx, "ruleResponse.Source.IPRange - "+fmt.Sprintf("%v", currentRule.Source.IPRange))
			for _, item := range currentRule.Source.IPRange {
				curSourceIPRanges = append(curSourceIPRanges, parseFromTo(ctx, item))
			}
			curRuleSourceObjAttrs["ip_range"], diags = types.ListValueFrom(ctx, FromToObjectType, curSourceIPRanges)
			resp.Diagnostics.Append(diags...)
		}
	}

	// Rule -> Source -> GlobalIPRange
	if currentRule.Source.GlobalIPRange != nil {
		if len(currentRule.Source.GlobalIPRange) > 0 {
			var curSourceGlobalIPRanges []types.Object
			tflog.Info(ctx, "ruleResponse.Source.GlobalIPRange - "+fmt.Sprintf("%v", currentRule.Source.GlobalIPRange))
			for _, item := range currentRule.Source.GlobalIPRange {
				curSourceGlobalIPRanges = append(curSourceGlobalIPRanges, parseNameID(ctx, item))
			}
			curRuleSourceObjAttrs["global_ip_range"], diags = types.ListValueFrom(ctx, NameIDObjectType, curSourceGlobalIPRanges)
			resp.Diagnostics.Append(diags...)
		}
	}

	// Rule -> Source -> NetworkInterface
	if currentRule.Source.NetworkInterface != nil {
		if len(currentRule.Source.NetworkInterface) > 0 {
			var curSourceNetworkInterfaces []types.Object
			tflog.Info(ctx, "ruleResponse.Source.NetworkInterface - "+fmt.Sprintf("%v", currentRule.Source.NetworkInterface))
			for _, item := range currentRule.Source.NetworkInterface {
				curSourceNetworkInterfaces = append(curSourceNetworkInterfaces, parseNameID(ctx, item))
			}
			curRuleSourceObjAttrs["network_interface"], diags = types.ListValueFrom(ctx, NameIDObjectType, curSourceNetworkInterfaces)
			resp.Diagnostics.Append(diags...)
		}
	}

	// Rule -> Source -> SiteNetworkSubnet
	if currentRule.Source.SiteNetworkSubnet != nil {
		if len(currentRule.Source.SiteNetworkSubnet) > 0 {
			var curSourceSiteNetworkSubnets []types.Object
			tflog.Info(ctx, "ruleResponse.Source.SiteNetworkSubnet - "+fmt.Sprintf("%v", currentRule.Source.SiteNetworkSubnet))
			for _, item := range currentRule.Source.SiteNetworkSubnet {
				curSourceSiteNetworkSubnets = append(curSourceSiteNetworkSubnets, parseNameID(ctx, item))
			}
			curRuleSourceObjAttrs["site_network_subnet"], diags = types.ListValueFrom(ctx, NameIDObjectType, curSourceSiteNetworkSubnets)
			resp.Diagnostics.Append(diags...)
		}
	}

	// Rule -> Source -> FloatingSubnet
	if currentRule.Source.FloatingSubnet != nil {
		if len(currentRule.Source.FloatingSubnet) > 0 {
			var curSourceFloatingSubnets []types.Object
			tflog.Info(ctx, "ruleResponse.Source.FloatingSubnet - "+fmt.Sprintf("%v", currentRule.Source.FloatingSubnet))
			for _, item := range currentRule.Source.FloatingSubnet {
				curSourceFloatingSubnets = append(curSourceFloatingSubnets, parseNameID(ctx, item))
			}
			curRuleSourceObjAttrs["floating_subnet"], diags = types.ListValueFrom(ctx, NameIDObjectType, curSourceFloatingSubnets)
			resp.Diagnostics.Append(diags...)
		}
	}

	// Rule -> Source -> User
	if currentRule.Source.User != nil {
		if len(currentRule.Source.User) > 0 {
			var curSourceUsers []types.Object
			tflog.Info(ctx, "ruleResponse.Source.User - "+fmt.Sprintf("%v", currentRule.Source.User))
			for _, item := range currentRule.Source.User {
				curSourceUsers = append(curSourceUsers, parseNameID(ctx, item))
			}
			curRuleSourceObjAttrs["user"], diags = types.ListValueFrom(ctx, NameIDObjectType, curSourceUsers)
			resp.Diagnostics.Append(diags...)
		}
	}

	// Rule -> Source -> UsersGroup
	if currentRule.Source.UsersGroup != nil {
		if len(currentRule.Source.UsersGroup) > 0 {
			var curSourceUsersGroups []types.Object
			tflog.Info(ctx, "ruleResponse.Source.UsersGroup - "+fmt.Sprintf("%v", currentRule.Source.UsersGroup))
			for _, item := range currentRule.Source.UsersGroup {
				curSourceUsersGroups = append(curSourceUsersGroups, parseNameID(ctx, item))
			}
			curRuleSourceObjAttrs["users_group"], diags = types.ListValueFrom(ctx, NameIDObjectType, curSourceUsersGroups)
			resp.Diagnostics.Append(diags...)
		}
	}

	// Rule -> Source -> Group
	if currentRule.Source.Group != nil {
		if len(currentRule.Source.Group) > 0 {
			var curSourceGroups []types.Object
			tflog.Info(ctx, "ruleResponse.Source.Group - "+fmt.Sprintf("%v", currentRule.Source.Group))
			for _, item := range currentRule.Source.Group {
				curSourceGroups = append(curSourceGroups, parseNameID(ctx, item))
			}
			curRuleSourceObjAttrs["group"], diags = types.ListValueFrom(ctx, NameIDObjectType, curSourceGroups)
			resp.Diagnostics.Append(diags...)
		}
	}

	// Rule -> Source -> SystemGroup
	if currentRule.Source.SystemGroup != nil {
		if len(currentRule.Source.SystemGroup) > 0 {
			var curSourceSystemGroups []types.Object
			tflog.Info(ctx, "ruleResponse.Source.SystemGroup - "+fmt.Sprintf("%v", currentRule.Source.SystemGroup))
			for _, item := range currentRule.Source.SystemGroup {
				curSourceSystemGroups = append(curSourceSystemGroups, parseNameID(ctx, item))
			}
			curRuleSourceObjAttrs["system_group"], diags = types.ListValueFrom(ctx, NameIDObjectType, curSourceSystemGroups)
			resp.Diagnostics.Append(diags...)
		}
	}

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
		map[string]attr.Type{
			"application":              types.ListType{ElemType: NameIDObjectType},
			"custom_app":               types.ListType{ElemType: NameIDObjectType},
			"app_category":             types.ListType{ElemType: NameIDObjectType},
			"custom_category":          types.ListType{ElemType: NameIDObjectType},
			"sanctioned_apps_category": types.ListType{ElemType: NameIDObjectType},
			"country":                  types.ListType{ElemType: NameIDObjectType},
			"domain":                   types.ListType{ElemType: types.StringType},
			"fqdn":                     types.ListType{ElemType: types.StringType},
			"ip":                       types.ListType{ElemType: types.StringType},
			"subnet":                   types.ListType{ElemType: types.StringType},
			"ip_range":                 types.ListType{ElemType: FromToObjectType},
			"global_ip_range":          types.ListType{ElemType: NameIDObjectType},
			"remote_asn":               types.ListType{ElemType: types.StringType},
		},
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

	// // Rule -> Destination -> IP
	tflog.Info(ctx, "ruleResponse.Destination.IP - "+fmt.Sprintf("%v", currentRule.Destination.IP))
	if currentRule.Destination.IP != nil {
		if len(currentRule.Destination.IP) > 0 {
			tflog.Info(ctx, "ruleResponse.Destination.IP - "+fmt.Sprintf("%v", currentRule.Destination.IP))
			curDestIpList, diagstmp := types.ListValueFrom(ctx, types.StringType, currentRule.Destination.IP)
			diags = append(diags, diagstmp...)
			curRuleDestinationObjAttrs["ip"] = curDestIpList
		}
	}

	// Rule -> Destination -> Subnet
	if currentRule.Destination.Subnet != nil {
		if len(currentRule.Destination.Subnet) > 0 {
			tflog.Info(ctx, "ruleResponse.Destination.Subnet - "+fmt.Sprintf("%v", currentRule.Destination.Subnet))
			curDestSubnetList, diagstmp := types.ListValueFrom(ctx, types.StringType, currentRule.Destination.Subnet)
			diags = append(diags, diagstmp...)
			curRuleDestinationObjAttrs["subnet"] = curDestSubnetList
		}
	}

	// Rule -> Destination -> Domain
	if currentRule.Destination.Domain != nil {
		if len(currentRule.Destination.Domain) > 0 {
			tflog.Info(ctx, "ruleResponse.Destination.Domain - "+fmt.Sprintf("%v", currentRule.Destination.Domain))
			curDestDomainList, diagstmp := types.ListValueFrom(ctx, types.StringType, currentRule.Destination.Domain)
			diags = append(diags, diagstmp...)
			curRuleDestinationObjAttrs["domain"] = curDestDomainList
		}
	}

	// Rule -> Destination -> Fqdn
	if currentRule.Destination.Fqdn != nil {
		if len(currentRule.Destination.Fqdn) > 0 {
			tflog.Info(ctx, "ruleResponse.Destination.Fqdn - "+fmt.Sprintf("%v", currentRule.Destination.Fqdn))
			curDestFqdnList, diagstmp := types.ListValueFrom(ctx, types.StringType, currentRule.Destination.Fqdn)
			diags = append(diags, diagstmp...)
			curRuleDestinationObjAttrs["fqdn"] = curDestFqdnList
		}
	}

	// Rule -> Destination -> RemoteAsn
	if currentRule.Destination.RemoteAsn != nil {
		if len(currentRule.Destination.RemoteAsn) > 0 {
			tflog.Info(ctx, "ruleResponse.Destination.RemoteAsn - "+fmt.Sprintf("%v", currentRule.Destination.RemoteAsn))
			curDestRemoteAsnList, diagstmp := types.ListValueFrom(ctx, types.StringType, currentRule.Destination.RemoteAsn)
			diags = append(diags, diagstmp...)
			curRuleDestinationObjAttrs["remote_asn"] = curDestRemoteAsnList
		}
	}

	// Rule -> Destination -> Application
	if currentRule.Destination.Application != nil {
		if len(currentRule.Destination.Application) > 0 {
			var curDestApplications []types.Object
			tflog.Info(ctx, "ruleResponse.Destination.Application - "+fmt.Sprintf("%v", currentRule.Destination.Application))
			for _, item := range currentRule.Destination.Application {
				curDestApplications = append(curDestApplications, parseNameID(ctx, item))
			}
			curRuleDestinationObjAttrs["application"], diags = types.ListValueFrom(ctx, NameIDObjectType, curDestApplications)
			resp.Diagnostics.Append(diags...)
		}
	}

	// Rule -> Destination -> CustomApp
	if currentRule.Destination.CustomApp != nil {
		if len(currentRule.Destination.CustomApp) > 0 {
			var curDestCustomApps []types.Object
			tflog.Info(ctx, "ruleResponse.Destination.CustomApp - "+fmt.Sprintf("%v", currentRule.Destination.CustomApp))
			for _, item := range currentRule.Destination.CustomApp {
				curDestCustomApps = append(curDestCustomApps, parseNameID(ctx, item))
			}
			curRuleDestinationObjAttrs["custom_app"], diags = types.ListValueFrom(ctx, NameIDObjectType, curDestCustomApps)
			resp.Diagnostics.Append(diags...)
		}
	}

	// Rule -> Destination -> IPRange
	if currentRule.Destination.IPRange != nil {
		if len(currentRule.Destination.IPRange) > 0 {
			var curDestinationIPRanges []types.Object
			tflog.Info(ctx, "ruleResponse.Destination.IPRange - "+fmt.Sprintf("%v", currentRule.Destination.IPRange))
			for _, item := range currentRule.Destination.IPRange {
				curDestinationIPRanges = append(curDestinationIPRanges, parseFromTo(ctx, item))
			}
			curRuleDestinationObjAttrs["ip_range"], diags = types.ListValueFrom(ctx, FromToObjectType, curDestinationIPRanges)
			resp.Diagnostics.Append(diags...)
		}
	}

	// Rule -> Destination -> GlobalIPRange
	if currentRule.Destination.GlobalIPRange != nil {
		if len(currentRule.Destination.GlobalIPRange) > 0 {
			var curDestinationGlobalIPRanges []types.Object
			tflog.Info(ctx, "ruleResponse.Destination.GlobalIPRange - "+fmt.Sprintf("%v", currentRule.Destination.GlobalIPRange))
			for _, item := range currentRule.Destination.GlobalIPRange {
				curDestinationGlobalIPRanges = append(curDestinationGlobalIPRanges, parseNameID(ctx, item))
			}
			curRuleDestinationObjAttrs["global_ip_range"], diags = types.ListValueFrom(ctx, NameIDObjectType, curDestinationGlobalIPRanges)
			resp.Diagnostics.Append(diags...)
		}
	}

	// Rule -> Destination -> AppCategory
	if currentRule.Destination.AppCategory != nil {
		if len(currentRule.Destination.AppCategory) > 0 {
			var curDestinationAppCategories []types.Object
			tflog.Info(ctx, "ruleResponse.Destination.AppCategory - "+fmt.Sprintf("%v", currentRule.Destination.AppCategory))
			for _, item := range currentRule.Destination.AppCategory {
				curDestinationAppCategories = append(curDestinationAppCategories, parseNameID(ctx, item))
			}
			curRuleDestinationObjAttrs["app_category"], diags = types.ListValueFrom(ctx, NameIDObjectType, curDestinationAppCategories)
			resp.Diagnostics.Append(diags...)
		}
	}

	// Rule -> Destination -> CustomCategory
	if currentRule.Destination.CustomCategory != nil {
		if len(currentRule.Destination.CustomCategory) > 0 {
			var curDestinationCustomCategories []types.Object
			tflog.Info(ctx, "ruleResponse.Destination.CustomCategory - "+fmt.Sprintf("%v", currentRule.Destination.CustomCategory))
			for _, item := range currentRule.Destination.CustomCategory {
				curDestinationCustomCategories = append(curDestinationCustomCategories, parseNameID(ctx, item))
			}
			curRuleDestinationObjAttrs["custom_category"], diags = types.ListValueFrom(ctx, NameIDObjectType, curDestinationCustomCategories)
			resp.Diagnostics.Append(diags...)
		}
	}

	// Rule -> Destination -> SanctionedAppsCategory
	if currentRule.Destination.SanctionedAppsCategory != nil {
		if len(currentRule.Destination.SanctionedAppsCategory) > 0 {
			var curDestinationSanctionedAppsCategories []types.Object
			tflog.Info(ctx, "ruleResponse.Destination.SanctionedAppsCategory - "+fmt.Sprintf("%v", currentRule.Destination.SanctionedAppsCategory))
			for _, item := range currentRule.Destination.SanctionedAppsCategory {
				curDestinationSanctionedAppsCategories = append(curDestinationSanctionedAppsCategories, parseNameID(ctx, item))
			}
			curRuleDestinationObjAttrs["sanctioned_apps_category"], diags = types.ListValueFrom(ctx, NameIDObjectType, curDestinationSanctionedAppsCategories)
			resp.Diagnostics.Append(diags...)
		}
	}

	// Rule -> Destination -> Country
	if currentRule.Destination.Country != nil {
		if len(currentRule.Destination.Country) > 0 {
			var curDestinationCountries []types.Object
			tflog.Info(ctx, "ruleResponse.Destination.Country - "+fmt.Sprintf("%v", currentRule.Destination.Country))
			for _, item := range currentRule.Destination.Country {
				curDestinationCountries = append(curDestinationCountries, parseNameID(ctx, item))
			}
			curRuleDestinationObjAttrs["country"], diags = types.ListValueFrom(ctx, NameIDObjectType, curDestinationCountries)
			resp.Diagnostics.Append(diags...)
		}
	}

	curRuleDestinationObj, diags = types.ObjectValue(curRuleDestinationObj.AttributeTypes(ctx), curRuleDestinationObjAttrs)
	resp.Diagnostics.Append(diags...)
	ruleInput.Destination = curRuleDestinationObj
	////////////// end Rule -> Source ///////////////

	// Rule -> Service
	if len(currentRule.Service.Custom) > 0 || len(currentRule.Service.Standard) > 0 {
		var serviceInput *Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service
		diags = ruleInput.Service.As(ctx, &serviceInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)

		// Initialize Service object with null values
		curRuleServiceObj, diags := types.ObjectValue(
			map[string]attr.Type{
				"standard": types.ListType{ElemType: NameIDObjectType},
				"custom":   types.ListType{ElemType: CustomServiceObjectType},
			},
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

		// for _, customServiceElementsList := range currentRule.Service.Custom {
		// 	custServiceInternal := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service_Custom{}
		// 	diags = ruleInput.Source.As(ctx, &sourceInput, basetypes.ObjectAsOptions{})
		// 	resp.Diagnostics.Append(diags...)

		// 	custServiceInternal.Protocol = basetypes.NewStringValue(customServiceElementsList.Protocol.String())

		// 	// Rule -> Service -> Custom -> Port

		// 	// if !customServiceElementsList.Port.IsNull() {
		// 	// if len(customServiceElementsList.Port) > 0 {
		// 	// 	var elementsServiceCustomPortInput []attr.Value
		// 	// 	for _, v := range customServiceElementsList.Port {
		// 	// 		elementsServiceCustomPortInput = append(elementsServiceCustomPortInput, basetypes.NewStringValue(string(v)))
		// 	// 	}

		// 	// 	custServiceInternal.Port, diags = basetypes.NewListValue(types.StringType, elementsServiceCustomPortInput)
		// 	// 	resp.Diagnostics.Append(diags...)
		// 	// } else {
		// 	// 	custServiceInternal.Port = basetypes.NewListNull(types.StringType)
		// 	// }

		// 	// Rule -> Service -> Custom -> PortRange
		// 	if string(*customServiceElementsList.PortRange.GetFrom()) != "" && string(*customServiceElementsList.PortRange.GetTo()) != "" {
		// 		custServiceInternalPortrange := &Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service_Custom_PortRange{}

		// 		custServiceInternalPortrange.From = basetypes.NewStringValue(string(*customServiceElementsList.PortRange.GetFrom()))
		// 		custServiceInternalPortrange.To = basetypes.NewStringValue(string(*customServiceElementsList.PortRange.GetTo()))

		// 		custServiceInternal.PortRange, diags = basetypes.NewObjectValueFrom(ctx, mapAttributeTypes(ctx, custServiceInternalPortrange, resp), custServiceInternalPortrange)
		// 		resp.Diagnostics.Append(diags...)

		// 	} else {
		// 		custServiceInternalPortrange := &Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service_Custom_PortRange{}
		// 		custServiceInternal.PortRange = basetypes.NewObjectNull(mapAttributeTypes(ctx, custServiceInternalPortrange, resp))
		// 	}

		// 	//elementsServiceCustomInput = append(elementsServiceCustomInput, custServiceInternal)
		// }

		// serviceInput.Custom, diags = basetypes.NewListValueFrom(ctx, serviceInput.Custom.ElementType(ctx), elementsServiceCustomInput)
		// resp.Diagnostics.Append(diags...)

		curRuleServiceObj, diags = types.ObjectValue(curRuleServiceObj.AttributeTypes(ctx), curRuleServiceObjAttrs)
		resp.Diagnostics.Append(diags...)
		ruleInput.Service = curRuleServiceObj
	}

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

// ObjectType wrapper for ListValue
var ServiceObjectType = types.ObjectType{AttrTypes: ServiceAttrTypes}
var ServiceAttrTypes = map[string]attr.Type{
	"standard": types.ListType{ElemType: types.ObjectType{AttrTypes: NameIDAttrTypes}},
	"custom":   types.ListType{ElemType: types.ObjectType{AttrTypes: CustomServiceAttrTypes}},
}
var CustomServiceObjectType = types.ObjectType{AttrTypes: CustomServiceAttrTypes}
var CustomServiceAttrTypes = map[string]attr.Type{
	"port":       types.ListType{ElemType: types.StringType},
	"port_range": FromToObjectType,
	"protocol":   types.StringType,
}
var NameIDObjectType = types.ObjectType{AttrTypes: NameIDAttrTypes}
var NameIDAttrTypes = map[string]attr.Type{
	"name": types.StringType,
	"id":   types.StringType,
}
var FromToObjectType = types.ObjectType{AttrTypes: FromToAttrTypes}
var FromToAttrTypes = map[string]attr.Type{
	"from": types.StringType,
	"to":   types.StringType,
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

func parseFromTo(ctx context.Context, item interface{}) types.Object {
	tflog.Warn(ctx, "parseFromTo() - "+fmt.Sprintf("%v", item))
	diags := make(diag.Diagnostics, 0)

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
				return types.ObjectNull(CustomServiceAttrTypes)
			}
		} else {
			return types.ObjectNull(CustomServiceAttrTypes)
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
