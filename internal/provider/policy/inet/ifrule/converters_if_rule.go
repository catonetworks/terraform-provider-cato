package ifrule

import (
	"context"
	"fmt"
	"testing"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	cato_scalars "github.com/catonetworks/cato-go-sdk/scalars"
	attr "github.com/hashicorp/terraform-plugin-framework/attr"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	resource "github.com/hashicorp/terraform-plugin-framework/resource"
	schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	activeperiod "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/activeperiod"
	convert "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	customservice "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/customservice"
	idname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	iprange "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/iprange"
	oldidname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/oldidname"
	policyrule "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
	policyschedule "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyschedule"
	portrange "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/portrange"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

// ifwSubPolicyCleanupRuleID returns the id of the auto-created cleanup rule for
// the given sub-policy. The cleanup rule is a POLICY_RULE owned by the
// sub-policy named "<sub-policy name> - Cleanup Rule"; it always exists and is
// always last, making it a stable anchor for inserting rules into the
// sub-policy. Falls back to any non-scope rule owned by the sub-policy.
func ifwSubPolicyCleanupRuleID(body *cato_go_sdk.Policy, subID string) string {
	info := ifwSubPolicyInfo(body, subID)
	if info == nil {
		return ""
	}
	cleanupName := info.GetName() + " - Cleanup Rule"
	fallback := ""
	for _, rp := range body.GetPolicy().GetInternetFirewall().GetPolicy().GetRules() {
		sp := rp.GetSubPolicy()
		rt := rp.GetRuleType()
		if sp == nil || sp.GetID() != subID || rt == nil || *rt != cato_models.PolicyRuleTypeEnumPolicyRule {
			continue
		}
		if rp.GetRule().GetName() == cleanupName {
			return rp.GetRule().GetID()
		}
		if fallback == "" {
			fallback = rp.GetRule().GetID()
		}
	}
	return fallback
}

// ifwSubPolicyInfo returns the sub-policy info block for the given sub-policy id,
// or nil when it is not present in the policy snapshot.
func ifwSubPolicyInfo(
	body *cato_go_sdk.Policy,
	subID string,
) *cato_go_sdk.Policy_Policy_InternetFirewall_Policy_SubPolicies_Policy {
	if body == nil || subID == "" {
		return nil
	}
	for _, sp := range body.GetPolicy().GetInternetFirewall().GetPolicy().GetSubPolicies() {
		if sp.GetPolicy().GetID() == subID {
			return sp.GetPolicy()
		}
	}
	return nil
}

const (
	policyActiveOnAlways       = "ALWAYS"
	defaultConnectionOriginAny = "ANY"
)

func appendObjectRefTransformError(diags []diag.Diagnostic, fieldPath string, err error) []diag.Diagnostic {
	return append(diags, diag.NewErrorDiagnostic(
		"Invalid object reference",
		fmt.Sprintf("Unable to convert %s into API reference: %v", fieldPath, err),
	))
}

// HydrateIfwRuleAPI takes in the current state/plan along with context and returns the created
// diagnostic data as well as cato api data used to either create or update IFW entries
//
//nolint:gocyclo,funlen
func HydrateIfwRuleAPI(ctx context.Context, plan InternetFirewallRule) (HydrateIfwAPITypes, diag.Diagnostics) {
	diags := []diag.Diagnostic{}

	hydrateAPIReturn := HydrateIfwAPITypes{}
	hydrateAPIReturn.Create = cato_models.InternetFirewallAddRuleInput{}
	hydrateAPIReturn.Update = cato_models.InternetFirewallUpdateRuleInput{}
	hydrateAPIReturn.Create.At = &cato_models.PolicyRulePositionInput{}

	rootAddRule := &cato_models.InternetFirewallAddRuleDataInput{}
	rootUpdateRule := &cato_models.InternetFirewallUpdateRuleDataInput{}

	// setting at for creation only
	if !plan.At.IsNull() {
		positionInput := policyrule.Position{}
		diags = append(diags, plan.At.As(ctx, &positionInput, basetypes.ObjectAsOptions{})...)

		hydrateAPIReturn.Create.At.Position = (*cato_models.PolicyRulePositionEnum)(positionInput.Position.ValueStringPointer())
		hydrateAPIReturn.Create.At.Ref = positionInput.Ref.ValueStringPointer()
	}

	// setting rule
	if !plan.Rule.IsNull() {
		ruleInput := PolicyPolicyInternetFirewallPolicyRulesRule{}
		diags = append(diags, plan.Rule.As(ctx, &ruleInput, basetypes.ObjectAsOptions{})...)

		// setting source
		if !ruleInput.Source.IsNull() {
			ruleSourceInput := &cato_models.InternetFirewallSourceInput{}
			ruleSourceUpdateInput := &cato_models.InternetFirewallSourceUpdateInput{}

			sourceInput := PolicyPolicyInternetFirewallPolicyRulesRuleSource{}
			diags = append(diags, ruleInput.Source.As(ctx, &sourceInput, basetypes.ObjectAsOptions{})...)

			// setting source IP
			if !sourceInput.IP.IsUnknown() && !sourceInput.IP.IsNull() {
				diags = append(diags, sourceInput.IP.ElementsAs(ctx, &ruleSourceInput.IP, false)...)
				diags = append(diags, sourceInput.IP.ElementsAs(ctx, &ruleSourceUpdateInput.IP, false)...)
			} else {
				ruleSourceUpdateInput.IP = (make([]string, 0))
			}

			// setting source subnet
			if !sourceInput.Subnet.IsUnknown() && !sourceInput.Subnet.IsNull() {
				diags = append(diags, sourceInput.Subnet.ElementsAs(ctx, &ruleSourceInput.Subnet, false)...)
				diags = append(diags, sourceInput.Subnet.ElementsAs(ctx, &ruleSourceUpdateInput.Subnet, false)...)
			} else {
				ruleSourceUpdateInput.Subnet = make([]string, 0)
			}

			// setting source host
			if !sourceInput.Host.IsUnknown() && !sourceInput.Host.IsNull() {
				elementsSourceHostInput := make([]types.Object, 0, len(sourceInput.Host.Elements()))
				diags = append(diags, sourceInput.Host.ElementsAs(ctx, &elementsSourceHostInput, false)...)

				var itemSourceHostInput PolicyPolicyInternetFirewallPolicyRulesRuleSourceHost
				for _, item := range elementsSourceHostInput {
					diags = append(diags, item.As(ctx, &itemSourceHostInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceHostInput)
					if err != nil {
						diags = appendObjectRefTransformError(diags, "rule.source.host", err)
						return hydrateAPIReturn, diags
					}

					ruleSourceInput.Host = append(ruleSourceInput.Host, &cato_models.HostRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
				ruleSourceUpdateInput.Host = ruleSourceInput.Host
			} else {
				ruleSourceUpdateInput.Host = make([]*cato_models.HostRefInput, 0)
			}

			// setting source site
			if !sourceInput.Site.IsUnknown() && !sourceInput.Site.IsNull() {
				elementsSourceSiteInput := make([]types.Object, 0, len(sourceInput.Site.Elements()))
				diags = append(diags, sourceInput.Site.ElementsAs(ctx, &elementsSourceSiteInput, false)...)

				var itemSourceSiteInput PolicyPolicyInternetFirewallPolicyRulesRuleSourceSite
				for _, item := range elementsSourceSiteInput {
					diags = append(diags, item.As(ctx, &itemSourceSiteInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceSiteInput)
					if err != nil {
						diags = appendObjectRefTransformError(diags, "rule.source.site", err)
						return hydrateAPIReturn, diags
					}

					ruleSourceInput.Site = append(ruleSourceInput.Site, &cato_models.SiteRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
				ruleSourceUpdateInput.Site = ruleSourceInput.Site
			} else {
				ruleSourceUpdateInput.Site = make([]*cato_models.SiteRefInput, 0)
			}

			// setting source ip range
			if !sourceInput.IPRange.IsUnknown() && !sourceInput.IPRange.IsNull() {
				elementsSourceIPRangeInput := make([]types.Object, 0, len(sourceInput.IPRange.Elements()))
				diags = append(diags, sourceInput.IPRange.ElementsAs(ctx, &elementsSourceIPRangeInput, false)...)

				var itemSourceIPRangeInput PolicyPolicyInternetFirewallPolicyRulesRuleSourceIPRange
				for _, item := range elementsSourceIPRangeInput {
					diags = append(diags, item.As(ctx, &itemSourceIPRangeInput, basetypes.ObjectAsOptions{})...)

					ruleSourceInput.IPRange = append(ruleSourceInput.IPRange, &cato_models.IPAddressRangeInput{
						From: itemSourceIPRangeInput.From.ValueString(),
						To:   itemSourceIPRangeInput.To.ValueString(),
					})
				}
				ruleSourceUpdateInput.IPRange = ruleSourceInput.IPRange
			} else {
				ruleSourceUpdateInput.IPRange = make([]*cato_models.IPAddressRangeInput, 0)
			}

			// setting source global ip range
			if !sourceInput.GlobalIPRange.IsNull() {
				elementsSourceGlobalIPRangeInput := make([]types.Object, 0, len(sourceInput.GlobalIPRange.Elements()))
				diags = append(diags, sourceInput.GlobalIPRange.ElementsAs(ctx, &elementsSourceGlobalIPRangeInput, false)...)

				var itemSourceGlobalIPRangeInput PolicyPolicyInternetFirewallPolicyRulesRuleSourceGlobalIPRange
				for _, item := range elementsSourceGlobalIPRangeInput {
					diags = append(diags, item.As(ctx, &itemSourceGlobalIPRangeInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceGlobalIPRangeInput)
					if err != nil {
						diags = appendObjectRefTransformError(diags, "rule.source.global_ip_range", err)
						return hydrateAPIReturn, diags
					}

					ruleSourceInput.GlobalIPRange = append(ruleSourceInput.GlobalIPRange, &cato_models.GlobalIPRangeRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
				ruleSourceUpdateInput.GlobalIPRange = ruleSourceInput.GlobalIPRange
			} else {
				ruleSourceUpdateInput.GlobalIPRange = make([]*cato_models.GlobalIPRangeRefInput, 0)
			}

			// setting source network interface
			if !sourceInput.NetworkInterface.IsNull() {
				elementsSourceNetworkInterfaceInput := make([]types.Object, 0, len(sourceInput.NetworkInterface.Elements()))
				diags = append(diags, sourceInput.NetworkInterface.ElementsAs(ctx, &elementsSourceNetworkInterfaceInput, false)...)

				var itemSourceNetworkInterfaceInput PolicyPolicyInternetFirewallPolicyRulesRuleSourceNetworkInterface
				for _, item := range elementsSourceNetworkInterfaceInput {
					diags = append(diags, item.As(ctx, &itemSourceNetworkInterfaceInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceNetworkInterfaceInput)
					if err != nil {
						diags = appendObjectRefTransformError(diags, "rule.source.network_interface", err)
						return hydrateAPIReturn, diags
					}

					ruleSourceInput.NetworkInterface = append(ruleSourceInput.NetworkInterface, &cato_models.NetworkInterfaceRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
				ruleSourceUpdateInput.NetworkInterface = ruleSourceInput.NetworkInterface
			} else {
				ruleSourceUpdateInput.NetworkInterface = make([]*cato_models.NetworkInterfaceRefInput, 0)
			}

			// setting source site network subnet
			if !sourceInput.SiteNetworkSubnet.IsNull() {
				elementsSourceSiteNetworkSubnetInput := make([]types.Object, 0, len(sourceInput.SiteNetworkSubnet.Elements()))
				diags = append(diags, sourceInput.SiteNetworkSubnet.ElementsAs(ctx, &elementsSourceSiteNetworkSubnetInput, false)...)

				var itemSourceSiteNetworkSubnetInput PolicyPolicyInternetFirewallPolicyRulesRuleSourceSiteNetworkSubnet
				for _, item := range elementsSourceSiteNetworkSubnetInput {
					diags = append(diags, item.As(ctx, &itemSourceSiteNetworkSubnetInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceSiteNetworkSubnetInput)
					if err != nil {
						diags = appendObjectRefTransformError(diags, "rule.source.site_network_subnet", err)
						return hydrateAPIReturn, diags
					}

					ruleSourceInput.SiteNetworkSubnet = append(ruleSourceInput.SiteNetworkSubnet, &cato_models.SiteNetworkSubnetRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
				ruleSourceUpdateInput.SiteNetworkSubnet = ruleSourceInput.SiteNetworkSubnet
			} else {
				ruleSourceUpdateInput.SiteNetworkSubnet = make([]*cato_models.SiteNetworkSubnetRefInput, 0)
			}

			// setting source floating subnet
			if !sourceInput.FloatingSubnet.IsNull() {
				elementsSourceFloatingSubnetInput := make([]types.Object, 0, len(sourceInput.FloatingSubnet.Elements()))
				diags = append(diags, sourceInput.FloatingSubnet.ElementsAs(ctx, &elementsSourceFloatingSubnetInput, false)...)

				var itemSourceFloatingSubnetInput PolicyPolicyInternetFirewallPolicyRulesRuleSourceFloatingSubnet
				for _, item := range elementsSourceFloatingSubnetInput {
					diags = append(diags, item.As(ctx, &itemSourceFloatingSubnetInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceFloatingSubnetInput)
					if err != nil {
						diags = appendObjectRefTransformError(diags, "rule.source.floating_subnet", err)
						return hydrateAPIReturn, diags
					}

					ruleSourceInput.FloatingSubnet = append(ruleSourceInput.FloatingSubnet, &cato_models.FloatingSubnetRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
				ruleSourceUpdateInput.FloatingSubnet = ruleSourceInput.FloatingSubnet
			} else {
				ruleSourceUpdateInput.FloatingSubnet = make([]*cato_models.FloatingSubnetRefInput, 0)
			}

			// setting source user
			if !sourceInput.User.IsNull() {
				elementsSourceUserInput := make([]types.Object, 0, len(sourceInput.User.Elements()))
				diags = append(diags, sourceInput.User.ElementsAs(ctx, &elementsSourceUserInput, false)...)

				var itemSourceUserInput PolicyPolicyInternetFirewallPolicyRulesRuleSourceUser
				for _, item := range elementsSourceUserInput {
					diags = append(diags, item.As(ctx, &itemSourceUserInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceUserInput)
					if err != nil {
						diags = appendObjectRefTransformError(diags, "rule.source.user", err)
						return hydrateAPIReturn, diags
					}

					ruleSourceInput.User = append(ruleSourceInput.User, &cato_models.UserRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
				ruleSourceUpdateInput.User = ruleSourceInput.User
			} else {
				ruleSourceUpdateInput.User = make([]*cato_models.UserRefInput, 0)
			}

			// setting source users group
			if !sourceInput.UsersGroup.IsNull() {
				elementsSourceUsersGroupInput := make([]types.Object, 0, len(sourceInput.UsersGroup.Elements()))
				diags = append(diags, sourceInput.UsersGroup.ElementsAs(ctx, &elementsSourceUsersGroupInput, false)...)

				var itemSourceUsersGroupInput PolicyPolicyInternetFirewallPolicyRulesRuleSourceUsersGroup
				for _, item := range elementsSourceUsersGroupInput {
					diags = append(diags, item.As(ctx, &itemSourceUsersGroupInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceUsersGroupInput)
					if err != nil {
						diags = appendObjectRefTransformError(diags, "rule.source.users_group", err)
						return hydrateAPIReturn, diags
					}

					ruleSourceInput.UsersGroup = append(ruleSourceInput.UsersGroup, &cato_models.UsersGroupRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
				ruleSourceUpdateInput.UsersGroup = ruleSourceInput.UsersGroup
			} else {
				ruleSourceUpdateInput.UsersGroup = make([]*cato_models.UsersGroupRefInput, 0)
			}

			// setting source group
			if !sourceInput.Group.IsNull() {
				elementsSourceGroupInput := make([]types.Object, 0, len(sourceInput.Group.Elements()))
				diags = append(diags, sourceInput.Group.ElementsAs(ctx, &elementsSourceGroupInput, false)...)

				var itemSourceGroupInput PolicyPolicyInternetFirewallPolicyRulesRuleSourceGroup
				for _, item := range elementsSourceGroupInput {
					diags = append(diags, item.As(ctx, &itemSourceGroupInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceGroupInput)
					if err != nil {
						diags = appendObjectRefTransformError(diags, "rule.source.group", err)
						return hydrateAPIReturn, diags
					}

					ruleSourceInput.Group = append(ruleSourceInput.Group, &cato_models.GroupRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
				ruleSourceUpdateInput.Group = ruleSourceInput.Group
			} else {
				ruleSourceUpdateInput.Group = make([]*cato_models.GroupRefInput, 0)
			}

			// setting source system group
			if !sourceInput.SystemGroup.IsNull() {
				elementsSourceSystemGroupInput := make([]types.Object, 0, len(sourceInput.SystemGroup.Elements()))
				diags = append(diags, sourceInput.SystemGroup.ElementsAs(ctx, &elementsSourceSystemGroupInput, false)...)

				var itemSourceSystemGroupInput PolicyPolicyInternetFirewallPolicyRulesRuleSourceSystemGroup
				for _, item := range elementsSourceSystemGroupInput {
					diags = append(diags, item.As(ctx, &itemSourceSystemGroupInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceSystemGroupInput)
					if err != nil {
						diags = appendObjectRefTransformError(diags, "rule.source.system_group", err)
						return hydrateAPIReturn, diags
					}

					ruleSourceInput.SystemGroup = append(ruleSourceInput.SystemGroup, &cato_models.SystemGroupRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
				ruleSourceUpdateInput.SystemGroup = ruleSourceInput.SystemGroup
			} else {
				ruleSourceUpdateInput.SystemGroup = make([]*cato_models.SystemGroupRefInput, 0)
			}

			rootAddRule.Source = ruleSourceInput
			rootUpdateRule.Source = ruleSourceUpdateInput
		}

		// setting country
		if !ruleInput.Country.IsNull() {
			elementsCountryInput := make([]types.Object, 0, len(ruleInput.Country.Elements()))
			diags = append(diags, ruleInput.Country.ElementsAs(ctx, &elementsCountryInput, false)...)

			var itemCountryInput PolicyPolicyInternetFirewallPolicyRulesRuleCountry
			for _, item := range elementsCountryInput {
				diags = append(diags, item.As(ctx, &itemCountryInput, basetypes.ObjectAsOptions{})...)

				ObjectRefOutput, err := utils.TransformObjectRefInput(itemCountryInput)
				if err != nil {
					diags = appendObjectRefTransformError(diags, "rule.country", err)
					return hydrateAPIReturn, diags
				}

				rootAddRule.Country = append(rootAddRule.Country, &cato_models.CountryRefInput{
					By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
					Input: ObjectRefOutput.Input,
				})
			}
			rootUpdateRule.Country = rootAddRule.Country
		} else {
			rootUpdateRule.Country = make([]*cato_models.CountryRefInput, 0)
		}

		// setting device
		if !ruleInput.Device.IsNull() {
			elementsDeviceInput := make([]types.Object, 0, len(ruleInput.Device.Elements()))
			diags = append(diags, ruleInput.Device.ElementsAs(ctx, &elementsDeviceInput, false)...)

			var itemDeviceInput PolicyPolicyInternetFirewallPolicyRulesRuleDevice
			for _, item := range elementsDeviceInput {
				diags = append(diags, item.As(ctx, &itemDeviceInput, basetypes.ObjectAsOptions{})...)

				ObjectRefOutput, err := utils.TransformObjectRefInput(itemDeviceInput)
				if err != nil {
					diags = appendObjectRefTransformError(diags, "rule.device", err)
					return hydrateAPIReturn, diags
				}

				rootAddRule.Device = append(rootAddRule.Device, &cato_models.DeviceProfileRefInput{
					By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
					Input: ObjectRefOutput.Input,
				})
			}
			rootUpdateRule.Device = rootAddRule.Device
		} else {
			rootUpdateRule.Device = make([]*cato_models.DeviceProfileRefInput, 0)
		}

		// setting device OS
		if !ruleInput.DeviceOs.IsUnknown() && !ruleInput.DeviceOs.IsNull() {
			diags = append(diags, ruleInput.DeviceOs.ElementsAs(ctx, &rootAddRule.DeviceOs, false)...)
			diags = append(diags, ruleInput.DeviceOs.ElementsAs(ctx, &rootUpdateRule.DeviceOs, false)...)
		} else {
			rootUpdateRule.DeviceOs = make([]cato_models.OperatingSystem, 0)
		}

		// setting device attributes
		if !ruleInput.DeviceAttributes.IsNull() {
			var curDeviceAttributes *DeviceAttributesInputIfw
			diags = append(diags, ruleInput.DeviceAttributes.As(ctx, &curDeviceAttributes, basetypes.ObjectAsOptions{})...)
			if curDeviceAttributes != nil {
				// Handle each field with proper null checking
				categoryValues := make([]string, 0)
				if curDeviceAttributes.Category != nil {
					categoryValues = curDeviceAttributes.Category
				}

				manufacturerValues := make([]string, 0)
				if curDeviceAttributes.Manufacturer != nil {
					manufacturerValues = curDeviceAttributes.Manufacturer
				}

				modelValues := make([]string, 0)
				if curDeviceAttributes.Model != nil {
					modelValues = curDeviceAttributes.Model
				}

				osValues := make([]string, 0)
				if curDeviceAttributes.Os != nil {
					osValues = curDeviceAttributes.Os
				}

				osVersionValues := make([]string, 0)
				if curDeviceAttributes.OsVersion != nil {
					osVersionValues = curDeviceAttributes.OsVersion
				}

				typeValues := make([]string, 0)
				if curDeviceAttributes.Type != nil {
					typeValues = curDeviceAttributes.Type
				}

				rootAddRule.DeviceAttributes = &cato_models.DeviceAttributesInput{
					Category:     categoryValues,
					Manufacturer: manufacturerValues,
					Model:        modelValues,
					Os:           osValues,
					OsVersion:    osVersionValues,
					Type:         typeValues,
				}
				rootUpdateRule.DeviceAttributes = &cato_models.DeviceAttributesUpdateInput{
					Category:     categoryValues,
					Manufacturer: manufacturerValues,
					Model:        modelValues,
					Os:           osValues,
					OsVersion:    osVersionValues,
					Type:         typeValues,
				}
			}
		} else {
			// DeviceAttributes should never be null in API - always provide empty arrays
			emptyDeviceAttributes := &cato_models.DeviceAttributesInput{
				Category:     make([]string, 0),
				Manufacturer: make([]string, 0),
				Model:        make([]string, 0),
				Os:           make([]string, 0),
				OsVersion:    make([]string, 0),
				Type:         make([]string, 0),
			}
			rootUpdateRule.DeviceAttributes = (*cato_models.DeviceAttributesUpdateInput)(emptyDeviceAttributes)
			rootAddRule.DeviceAttributes = emptyDeviceAttributes
		}

		// setting destination
		if !ruleInput.Destination.IsUnknown() && !ruleInput.Destination.IsNull() {
			ruleDestinationInput := &cato_models.InternetFirewallDestinationInput{}
			ruleDestinationUpdateInput := &cato_models.InternetFirewallDestinationUpdateInput{}
			ruleDestinationInput.Containers = &cato_models.InternetFirewallContainerInput{
				FqdnContainer:           make([]*cato_models.FqdnContainerRefInput, 0),
				IPAddressRangeContainer: make([]*cato_models.IPAddressRangeContainerRefInput, 0),
			}

			destinationInput := PolicyPolicyInternetFirewallPolicyRulesRuleDestination{}
			diags = append(diags, ruleInput.Destination.As(ctx, &destinationInput, basetypes.ObjectAsOptions{})...)

			// setting destination IP
			if !destinationInput.IP.IsUnknown() && !destinationInput.IP.IsNull() {
				diags = append(diags, destinationInput.IP.ElementsAs(ctx, &ruleDestinationInput.IP, false)...)
				diags = append(diags, destinationInput.IP.ElementsAs(ctx, &ruleDestinationUpdateInput.IP, false)...)
			} else {
				ruleDestinationUpdateInput.IP = make([]string, 0)
			}

			// setting destination subnet
			if !destinationInput.Subnet.IsUnknown() && !destinationInput.Subnet.IsNull() {
				diags = append(diags, destinationInput.Subnet.ElementsAs(ctx, &ruleDestinationInput.Subnet, false)...)
				diags = append(diags, destinationInput.Subnet.ElementsAs(ctx, &ruleDestinationUpdateInput.Subnet, false)...)
			} else {
				ruleDestinationUpdateInput.Subnet = make([]string, 0)
			}

			// setting destination domain
			if !destinationInput.Domain.IsUnknown() && !destinationInput.Domain.IsNull() {
				diags = append(diags, destinationInput.Domain.ElementsAs(ctx, &ruleDestinationInput.Domain, false)...)
				diags = append(diags, destinationInput.Domain.ElementsAs(ctx, &ruleDestinationUpdateInput.Domain, false)...)
			} else {
				ruleDestinationUpdateInput.Domain = make([]string, 0)
			}

			// setting destination fqdn
			if !destinationInput.Fqdn.IsUnknown() && !destinationInput.Fqdn.IsNull() {
				diags = append(diags, destinationInput.Fqdn.ElementsAs(ctx, &ruleDestinationInput.Fqdn, false)...)
				diags = append(diags, destinationInput.Fqdn.ElementsAs(ctx, &ruleDestinationUpdateInput.Fqdn, false)...)
			} else {
				ruleDestinationUpdateInput.Fqdn = make([]string, 0)
			}

			// setting destination remote asn
			if !destinationInput.RemoteAsn.IsUnknown() && !destinationInput.RemoteAsn.IsNull() {
				diags = append(diags, destinationInput.RemoteAsn.ElementsAs(ctx, &ruleDestinationInput.RemoteAsn, false)...)
				diags = append(diags, destinationInput.RemoteAsn.ElementsAs(ctx, &ruleDestinationUpdateInput.RemoteAsn, false)...)
			} else {
				ruleDestinationUpdateInput.RemoteAsn = make([]cato_scalars.Asn32, 0)
			}

			// setting destination application
			if !destinationInput.Application.IsUnknown() && !destinationInput.Application.IsNull() {
				elementsDestinationApplicationInput := make([]types.Object, 0, len(destinationInput.Application.Elements()))
				diags = append(diags, destinationInput.Application.ElementsAs(ctx, &elementsDestinationApplicationInput, false)...)

				var itemDestinationApplicationInput PolicyPolicyInternetFirewallPolicyRulesRuleDestinationApplication
				for _, item := range elementsDestinationApplicationInput {
					diags = append(diags, item.As(ctx, &itemDestinationApplicationInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationApplicationInput)
					if err != nil {
						diags = appendObjectRefTransformError(diags, "rule.destination.application", err)
						return hydrateAPIReturn, diags
					}

					ruleDestinationInput.Application = append(ruleDestinationInput.Application, &cato_models.ApplicationRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
				ruleDestinationUpdateInput.Application = ruleDestinationInput.Application
			} else {
				ruleDestinationUpdateInput.Application = make([]*cato_models.ApplicationRefInput, 0)
			}

			// setting destination custom app
			if !destinationInput.CustomApp.IsNull() {
				elementsDestinationCustomAppInput := make([]types.Object, 0, len(destinationInput.CustomApp.Elements()))
				diags = append(diags, destinationInput.CustomApp.ElementsAs(ctx, &elementsDestinationCustomAppInput, false)...)

				var itemDestinationCustomAppInput PolicyPolicyInternetFirewallPolicyRulesRuleDestinationCustomApp
				for _, item := range elementsDestinationCustomAppInput {
					diags = append(diags, item.As(ctx, &itemDestinationCustomAppInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationCustomAppInput)
					if err != nil {
						diags = appendObjectRefTransformError(diags, "rule.destination.custom_app", err)
						return hydrateAPIReturn, diags
					}

					ruleDestinationInput.CustomApp = append(ruleDestinationInput.CustomApp, &cato_models.CustomApplicationRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
				ruleDestinationUpdateInput.CustomApp = ruleDestinationInput.CustomApp
			} else {
				ruleDestinationUpdateInput.CustomApp = make([]*cato_models.CustomApplicationRefInput, 0)
			}

			// setting destination ip range
			if !destinationInput.IPRange.IsNull() {
				elementsDestinationIPRangeInput := make([]types.Object, 0, len(destinationInput.IPRange.Elements()))
				diags = append(diags, destinationInput.IPRange.ElementsAs(ctx, &elementsDestinationIPRangeInput, false)...)

				var itemDestinationIPRangeInput iprange.Range
				for _, item := range elementsDestinationIPRangeInput {
					diags = append(diags, item.As(ctx, &itemDestinationIPRangeInput, basetypes.ObjectAsOptions{})...)

					ruleDestinationInput.IPRange = append(ruleDestinationInput.IPRange, &cato_models.IPAddressRangeInput{
						From: itemDestinationIPRangeInput.From.ValueString(),
						To:   itemDestinationIPRangeInput.To.ValueString(),
					})
				}
				ruleDestinationUpdateInput.IPRange = ruleDestinationInput.IPRange
			} else {
				ruleDestinationUpdateInput.IPRange = make([]*cato_models.IPAddressRangeInput, 0)
			}

			// setting destination global ip range
			if !destinationInput.GlobalIPRange.IsNull() {
				elementsDestinationGlobalIPRangeInput := make([]types.Object, 0, len(destinationInput.GlobalIPRange.Elements()))
				diags = append(diags, destinationInput.GlobalIPRange.ElementsAs(ctx, &elementsDestinationGlobalIPRangeInput, false)...)

				var itemDestinationGlobalIPRangeInput PolicyPolicyInternetFirewallPolicyRulesRuleDestinationGlobalIPRange
				for _, item := range elementsDestinationGlobalIPRangeInput {
					diags = append(diags, item.As(ctx, &itemDestinationGlobalIPRangeInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationGlobalIPRangeInput)
					if err != nil {
						diags = appendObjectRefTransformError(diags, "rule.destination.global_ip_range", err)
						return hydrateAPIReturn, diags
					}

					ruleDestinationInput.GlobalIPRange = append(ruleDestinationInput.GlobalIPRange, &cato_models.GlobalIPRangeRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
				ruleDestinationUpdateInput.GlobalIPRange = ruleDestinationInput.GlobalIPRange
			} else {
				ruleDestinationUpdateInput.GlobalIPRange = make([]*cato_models.GlobalIPRangeRefInput, 0)
			}

			// setting destination app category
			if !destinationInput.AppCategory.IsUnknown() && !destinationInput.AppCategory.IsNull() {
				elementsDestinationAppCategoryInput := make([]types.Object, 0, len(destinationInput.AppCategory.Elements()))
				diags = append(diags, destinationInput.AppCategory.ElementsAs(ctx, &elementsDestinationAppCategoryInput, false)...)

				var itemDestinationAppCategoryInput PolicyPolicyInternetFirewallPolicyRulesRuleDestinationAppCategory
				for _, item := range elementsDestinationAppCategoryInput {
					diags = append(diags, item.As(ctx, &itemDestinationAppCategoryInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationAppCategoryInput)
					if err != nil {
						diags = appendObjectRefTransformError(diags, "rule.destination.app_category", err)
						return hydrateAPIReturn, diags
					}

					ruleDestinationInput.AppCategory = append(ruleDestinationInput.AppCategory, &cato_models.ApplicationCategoryRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
				ruleDestinationUpdateInput.AppCategory = ruleDestinationInput.AppCategory
			} else {
				ruleDestinationUpdateInput.AppCategory = make([]*cato_models.ApplicationCategoryRefInput, 0)
			}

			// setting destination custom app category
			if !destinationInput.CustomCategory.IsNull() {
				elementsDestinationCustomCategoryInput := make([]types.Object, 0, len(destinationInput.CustomCategory.Elements()))
				diags = append(diags, destinationInput.CustomCategory.ElementsAs(ctx, &elementsDestinationCustomCategoryInput, false)...)

				var itemDestinationCustomCategoryInput PolicyPolicyInternetFirewallPolicyRulesRuleDestinationCustomCategory
				for _, item := range elementsDestinationCustomCategoryInput {
					diags = append(diags, item.As(ctx, &itemDestinationCustomCategoryInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationCustomCategoryInput)
					if err != nil {
						diags = appendObjectRefTransformError(diags, "rule.destination.custom_category", err)
						return hydrateAPIReturn, diags
					}

					ruleDestinationInput.CustomCategory = append(ruleDestinationInput.CustomCategory, &cato_models.CustomCategoryRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
				ruleDestinationUpdateInput.CustomCategory = ruleDestinationInput.CustomCategory
			} else {
				ruleDestinationUpdateInput.CustomCategory = make([]*cato_models.CustomCategoryRefInput, 0)
			}

			// setting destination sanctioned apps category
			if !destinationInput.SanctionedAppsCategory.IsNull() {
				elementsDestinationSanctionedAppsCategoryInput := make([]types.Object, 0, len(destinationInput.SanctionedAppsCategory.Elements()))
				diags = append(
					diags,
					destinationInput.SanctionedAppsCategory.ElementsAs(ctx, &elementsDestinationSanctionedAppsCategoryInput, false)...,
				)

				var itemDestinationSanctionedAppsCategoryInput PolicyPolicyInternetFirewallPolicyRulesRuleDestinationSanctionedAppsCategory
				for _, item := range elementsDestinationSanctionedAppsCategoryInput {
					diags = append(diags, item.As(ctx, &itemDestinationSanctionedAppsCategoryInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationSanctionedAppsCategoryInput)
					if err != nil {
						diags = appendObjectRefTransformError(diags, "rule.destination.sanctioned_apps_category", err)
						return hydrateAPIReturn, diags
					}

					ruleDestinationInput.SanctionedAppsCategory = append(
						ruleDestinationInput.SanctionedAppsCategory,
						&cato_models.SanctionedAppsCategoryRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						},
					)
				}
				ruleDestinationUpdateInput.SanctionedAppsCategory = ruleDestinationInput.SanctionedAppsCategory
			} else {
				ruleDestinationUpdateInput.SanctionedAppsCategory = make([]*cato_models.SanctionedAppsCategoryRefInput, 0)
			}

			// setting destination country
			if !destinationInput.Country.IsNull() {
				elementsDestinationCountryInput := make([]types.Object, 0, len(destinationInput.Country.Elements()))
				diags = append(diags, destinationInput.Country.ElementsAs(ctx, &elementsDestinationCountryInput, false)...)

				var itemDestinationCountryInput PolicyPolicyInternetFirewallPolicyRulesRuleDestinationCountry
				for _, item := range elementsDestinationCountryInput {
					diags = append(diags, item.As(ctx, &itemDestinationCountryInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationCountryInput)
					if err != nil {
						diags = appendObjectRefTransformError(diags, "rule.destination.country", err)
						return hydrateAPIReturn, diags
					}

					ruleDestinationInput.Country = append(ruleDestinationInput.Country, &cato_models.CountryRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
				ruleDestinationUpdateInput.Country = ruleDestinationInput.Country
			} else {
				ruleDestinationUpdateInput.Country = make([]*cato_models.CountryRefInput, 0)
			}

			rootAddRule.Destination = ruleDestinationInput
			rootUpdateRule.Destination = ruleDestinationUpdateInput
		}

		// setting service
		if !ruleInput.Service.IsNull() {
			ruleServiceInput := &cato_models.InternetFirewallServiceTypeInput{}
			ruleServiceUpdateInput := &cato_models.InternetFirewallServiceTypeUpdateInput{}

			serviceInput := PolicyPolicyInternetFirewallPolicyRulesRuleService{}
			diags = append(diags, ruleInput.Service.As(ctx, &serviceInput, basetypes.ObjectAsOptions{})...)

			// setting service standard
			if !serviceInput.Standard.IsNull() {
				elementsServiceStandardInput := make([]types.Object, 0, len(serviceInput.Standard.Elements()))
				diags = append(diags, serviceInput.Standard.ElementsAs(ctx, &elementsServiceStandardInput, false)...)

				var itemServiceStandardInput PolicyPolicyInternetFirewallPolicyRulesRuleServiceStandard
				for _, item := range elementsServiceStandardInput {
					diags = append(diags, item.As(ctx, &itemServiceStandardInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemServiceStandardInput)
					if err != nil {
						diags = appendObjectRefTransformError(diags, "rule.service.standard", err)
						return hydrateAPIReturn, diags
					}

					ruleServiceInput.Standard = append(ruleServiceInput.Standard, &cato_models.ServiceRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
				ruleServiceUpdateInput.Standard = ruleServiceInput.Standard
			} else {
				ruleServiceUpdateInput.Standard = make([]*cato_models.ServiceRefInput, 0)
			}

			// setting service custom
			if !serviceInput.Custom.IsNull() {
				elementsServiceCustomInput := make([]types.Object, 0, len(serviceInput.Custom.Elements()))
				diags = append(diags, serviceInput.Custom.ElementsAs(ctx, &elementsServiceCustomInput, false)...)

				var itemServiceCustomInput customservice.CustomService
				for _, item := range elementsServiceCustomInput {
					diags = append(diags, item.As(ctx, &itemServiceCustomInput, basetypes.ObjectAsOptions{})...)

					tflog.Debug(ctx, "Processing custom service entry", map[string]interface{}{
						"port_is_null":       itemServiceCustomInput.Port.IsNull(),
						"port_range_is_null": itemServiceCustomInput.PortRange.IsNull(),
						"protocol":           itemServiceCustomInput.Protocol.ValueString(),
					})

					customInput := &cato_models.CustomServiceInput{
						Protocol: cato_models.IPProtocol(itemServiceCustomInput.Protocol.ValueString()),
					}

					// setting service custom port
					if !itemServiceCustomInput.Port.IsNull() {
						tflog.Debug(ctx, "Processing port field")
						elementsPort := make([]types.String, 0, len(itemServiceCustomInput.Port.Elements()))
						diags = append(diags, itemServiceCustomInput.Port.ElementsAs(ctx, &elementsPort, false)...)

						inputPort := []cato_scalars.Port{}
						for _, item := range elementsPort {
							inputPort = append(inputPort, cato_scalars.Port(item.ValueString()))
						}

						customInput.Port = inputPort
						tflog.Debug(ctx, "Set port field", map[string]interface{}{
							"port_count": len(inputPort),
						})
					}

					// setting service custom port range
					if !itemServiceCustomInput.PortRange.IsNull() {
						tflog.Debug(ctx, "Processing port_range field")
						var itemPortRange portrange.PortRange
						diags = append(diags, itemServiceCustomInput.PortRange.As(ctx, &itemPortRange, basetypes.ObjectAsOptions{})...)

						inputPortRange := cato_models.PortRangeInput{
							From: cato_scalars.Port(itemPortRange.From.ValueString()),
							To:   cato_scalars.Port(itemPortRange.To.ValueString()),
						}

						customInput.PortRange = &inputPortRange
						tflog.Debug(ctx, "Set port_range field", map[string]interface{}{
							"from": itemPortRange.From.ValueString(),
							"to":   itemPortRange.To.ValueString(),
						})
					}

					// append custom service
					ruleServiceInput.Custom = append(ruleServiceInput.Custom, customInput)
				}
				ruleServiceUpdateInput.Custom = ruleServiceInput.Custom
			} else {
				ruleServiceUpdateInput.Custom = make([]*cato_models.CustomServiceInput, 0)
			}

			rootAddRule.Service = ruleServiceInput
			rootUpdateRule.Service = ruleServiceUpdateInput
		} else {
			rootAddRule.Service = &cato_models.InternetFirewallServiceTypeInput{
				Standard: make([]*cato_models.ServiceRefInput, 0),
				Custom:   make([]*cato_models.CustomServiceInput, 0),
			}
			rootUpdateRule.Service = &cato_models.InternetFirewallServiceTypeUpdateInput{
				Standard: make([]*cato_models.ServiceRefInput, 0),
				Custom:   make([]*cato_models.CustomServiceInput, 0),
			}
		}

		// setting tracking
		if !ruleInput.Tracking.IsUnknown() && !ruleInput.Tracking.IsNull() {
			rootAddRule.Tracking = &cato_models.PolicyTrackingInput{
				Event: &cato_models.PolicyRuleTrackingEventInput{},
				Alert: &cato_models.PolicyRuleTrackingAlertInput{
					Enabled:   false,
					Frequency: "DAILY",
				},
			}
			rootUpdateRule.Tracking = &cato_models.PolicyTrackingUpdateInput{
				Event: &cato_models.PolicyRuleTrackingEventUpdateInput{},
				Alert: &cato_models.PolicyRuleTrackingAlertUpdateInput{},
			}

			trackingInput := policyrule.Tracking{}
			diags = append(diags, ruleInput.Tracking.As(ctx, &trackingInput, basetypes.ObjectAsOptions{})...)

			if !trackingInput.Event.IsUnknown() && !trackingInput.Event.IsNull() {
				// setting tracking event
				trackingEventInput := policyrule.TrackingEvent{}
				diags = append(diags, trackingInput.Event.As(ctx, &trackingEventInput, basetypes.ObjectAsOptions{})...)
				rootAddRule.Tracking.Event.Enabled = trackingEventInput.Enabled.ValueBool()
				rootUpdateRule.Tracking.Event.Enabled = trackingEventInput.Enabled.ValueBoolPointer()
			}

			if !trackingInput.Alert.IsUnknown() && !trackingInput.Alert.IsNull() {
				rootAddRule.Tracking.Alert = &cato_models.PolicyRuleTrackingAlertInput{}

				trackingAlertInput := policyrule.TrackingAlert{}
				diags = append(diags, trackingInput.Alert.As(ctx, &trackingAlertInput, basetypes.ObjectAsOptions{})...)

				rootAddRule.Tracking.Alert.Enabled = trackingAlertInput.Enabled.ValueBool()
				rootAddRule.Tracking.Alert.Frequency = cato_models.PolicyRuleTrackingFrequencyEnum(trackingAlertInput.Frequency.ValueString())

				rootUpdateRule.Tracking.Alert.Enabled = trackingAlertInput.Enabled.ValueBoolPointer()
				rootUpdateRule.Tracking.Alert.Frequency = (*cato_models.PolicyRuleTrackingFrequencyEnum)(
					trackingAlertInput.Frequency.ValueStringPointer(),
				)

				// setting tracking alert subscription group
				if !trackingAlertInput.SubscriptionGroup.IsNull() {
					elementsAlertSubscriptionGroupInput := make([]types.Object, 0, len(trackingAlertInput.SubscriptionGroup.Elements()))
					diags = append(diags, trackingAlertInput.SubscriptionGroup.ElementsAs(ctx, &elementsAlertSubscriptionGroupInput, false)...)

					var itemAlertSubscriptionGroupInput idname.Ref
					for _, item := range elementsAlertSubscriptionGroupInput {
						diags = append(diags, item.As(ctx, &itemAlertSubscriptionGroupInput, basetypes.ObjectAsOptions{})...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemAlertSubscriptionGroupInput)
						if err != nil {
							diags = appendObjectRefTransformError(diags, "rule.tracking.alert.subscription_group", err)
							return hydrateAPIReturn, diags
						}

						rootAddRule.Tracking.Alert.SubscriptionGroup = append(
							rootAddRule.Tracking.Alert.SubscriptionGroup,
							&cato_models.SubscriptionGroupRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							},
						)
					}
					rootUpdateRule.Tracking.Alert.SubscriptionGroup = rootAddRule.Tracking.Alert.SubscriptionGroup
				} else {
					rootUpdateRule.Tracking.Alert.SubscriptionGroup = make([]*cato_models.SubscriptionGroupRefInput, 0)
				}

				// setting tracking alert webhook
				if !trackingAlertInput.Webhook.IsNull() {
					elementsAlertWebHookInput := make([]types.Object, 0, len(trackingAlertInput.Webhook.Elements()))
					diags = append(diags, trackingAlertInput.Webhook.ElementsAs(ctx, &elementsAlertWebHookInput, false)...)

					var itemAlertWebHookInput idname.Ref
					for _, item := range elementsAlertWebHookInput {
						diags = append(diags, item.As(ctx, &itemAlertWebHookInput, basetypes.ObjectAsOptions{})...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemAlertWebHookInput)
						if err != nil {
							diags = appendObjectRefTransformError(diags, "rule.tracking.alert.webhook", err)
							return hydrateAPIReturn, diags
						}

						rootAddRule.Tracking.Alert.Webhook = append(rootAddRule.Tracking.Alert.Webhook, &cato_models.SubscriptionWebhookRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
					rootUpdateRule.Tracking.Alert.Webhook = rootAddRule.Tracking.Alert.Webhook
				} else {
					rootUpdateRule.Tracking.Alert.Webhook = make([]*cato_models.SubscriptionWebhookRefInput, 0)
				}

				// setting tracking alert mailing list
				tflog.Warn(ctx, "hydrateIfwApiRuleState() trackingAlertInput.MailingList "+trackingAlertInput.MailingList.String())
				if !trackingAlertInput.MailingList.IsNull() {
					elementsAlertMailingListInput := make([]types.Object, 0, len(trackingAlertInput.MailingList.Elements()))
					diags = append(diags, trackingAlertInput.MailingList.ElementsAs(ctx, &elementsAlertMailingListInput, false)...)

					var itemAlertMailingListInput idname.Ref
					for _, item := range elementsAlertMailingListInput {
						diags = append(diags, item.As(ctx, &itemAlertMailingListInput, basetypes.ObjectAsOptions{})...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemAlertMailingListInput)
						if err != nil {
							diags = appendObjectRefTransformError(diags, "rule.tracking.alert.mailing_list", err)
							return hydrateAPIReturn, diags
						}

						rootAddRule.Tracking.Alert.MailingList = append(rootAddRule.Tracking.Alert.MailingList, &cato_models.SubscriptionMailingListRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
					rootUpdateRule.Tracking.Alert.MailingList = rootAddRule.Tracking.Alert.MailingList
				} else {
					rootUpdateRule.Tracking.Alert.MailingList = make([]*cato_models.SubscriptionMailingListRefInput, 0)
				}
			}
		}

		// setting schedule
		rootAddRule.Schedule = &cato_models.PolicyScheduleInput{
			ActiveOn: cato_models.PolicyActiveOnEnum(policyActiveOnAlways),
		}

		activeOn := policyActiveOnAlways
		rootUpdateRule.Schedule = &cato_models.PolicyScheduleUpdateInput{
			ActiveOn: (*cato_models.PolicyActiveOnEnum)(&activeOn),
		}

		if !ruleInput.Schedule.IsUnknown() && !ruleInput.Schedule.IsNull() {
			scheduleInput := PolicyPolicyInternetFirewallPolicyRulesRuleSchedule{}
			diags = append(diags, ruleInput.Schedule.As(ctx, &scheduleInput, basetypes.ObjectAsOptions{})...)

			rootAddRule.Schedule.ActiveOn = cato_models.PolicyActiveOnEnum(scheduleInput.ActiveOn.ValueString())
			rootUpdateRule.Schedule.ActiveOn = (*cato_models.PolicyActiveOnEnum)(scheduleInput.ActiveOn.ValueStringPointer())

			// setting schedule custom time frame
			if !scheduleInput.CustomTimeframe.IsNull() {
				rootAddRule.Schedule.CustomTimeframe = &cato_models.PolicyCustomTimeframeInput{}
				rootUpdateRule.Schedule.CustomTimeframe = &cato_models.PolicyCustomTimeframeUpdateInput{}

				customTimeFrameInput := PolicyPolicyInternetFirewallPolicyRulesRuleScheduleCustomTimeframe{}
				diags = append(diags, scheduleInput.CustomTimeframe.As(ctx, &customTimeFrameInput, basetypes.ObjectAsOptions{})...)

				rootAddRule.Schedule.CustomTimeframe.From = customTimeFrameInput.From.ValueString()
				rootAddRule.Schedule.CustomTimeframe.To = customTimeFrameInput.To.ValueString()

				rootUpdateRule.Schedule.CustomTimeframe.From = customTimeFrameInput.From.ValueStringPointer()
				rootUpdateRule.Schedule.CustomTimeframe.To = customTimeFrameInput.To.ValueStringPointer()
			} else {
				rootUpdateRule.Schedule.CustomTimeframe = &cato_models.PolicyCustomTimeframeUpdateInput{}
			}

			// setting schedule custom recurring
			if !scheduleInput.CustomRecurring.IsNull() {
				rootAddRule.Schedule.CustomRecurring = &cato_models.PolicyCustomRecurringInput{}
				rootUpdateRule.Schedule.CustomRecurring = &cato_models.PolicyCustomRecurringUpdateInput{}

				customRecurringInput := PolicyPolicyInternetFirewallPolicyRulesRuleScheduleCustomRecurring{}
				diags = append(diags, scheduleInput.CustomRecurring.As(ctx, &customRecurringInput, basetypes.ObjectAsOptions{})...)

				rootAddRule.Schedule.CustomRecurring.From = cato_scalars.Time(customRecurringInput.From.ValueString())
				rootAddRule.Schedule.CustomRecurring.To = cato_scalars.Time(customRecurringInput.To.ValueString())
				rootUpdateRule.Schedule.CustomRecurring.From = (*cato_scalars.Time)(customRecurringInput.From.ValueStringPointer())
				rootUpdateRule.Schedule.CustomRecurring.To = (*cato_scalars.Time)(customRecurringInput.To.ValueStringPointer())

				// setting schedule custom recurring days
				diags = append(diags, customRecurringInput.Days.ElementsAs(ctx, &rootAddRule.Schedule.CustomRecurring.Days, false)...)
				rootUpdateRule.Schedule.CustomRecurring.Days = rootAddRule.Schedule.CustomRecurring.Days
			} else {
				rootUpdateRule.Schedule.CustomRecurring = &cato_models.PolicyCustomRecurringUpdateInput{}
			}
		} else {
			rootUpdateRule.Schedule = &cato_models.PolicyScheduleUpdateInput{}
		}

		// settings exceptions
		if !ruleInput.Exceptions.IsNull() && !ruleInput.Exceptions.IsUnknown() {
			elementsExceptionsInput := make([]types.Object, 0, len(ruleInput.Exceptions.Elements()))
			diags = append(diags, ruleInput.Exceptions.ElementsAs(ctx, &elementsExceptionsInput, false)...)

			// loop over exceptions
			var itemExceptionsInput PolicyPolicyInternetFirewallPolicyRulesRuleExceptions
			for _, item := range elementsExceptionsInput {
				exceptionAddInput := cato_models.InternetFirewallRuleExceptionInput{}
				exceptionUpdateInput := cato_models.InternetFirewallRuleExceptionInput{}

				diags = append(diags, item.As(ctx, &itemExceptionsInput, basetypes.ObjectAsOptions{})...)

				// setting exception name
				exceptionAddInput.Name = itemExceptionsInput.Name.ValueString()
				exceptionUpdateInput.Name = itemExceptionsInput.Name.ValueString()

				// setting exception connection origin
				if !itemExceptionsInput.ConnectionOrigin.IsUnknown() && !itemExceptionsInput.ConnectionOrigin.IsNull() {
					exceptionAddInput.ConnectionOrigin = cato_models.ConnectionOriginEnum(itemExceptionsInput.ConnectionOrigin.ValueString())
				} else {
					exceptionAddInput.ConnectionOrigin = cato_models.ConnectionOriginEnum(defaultConnectionOriginAny)
				}

				exceptionUpdateInput.ConnectionOrigin = exceptionAddInput.ConnectionOrigin

				// setting source
				if !itemExceptionsInput.Source.IsNull() {
					exceptionAddInput.Source = &cato_models.InternetFirewallSourceInput{}
					exceptionUpdateInput.Source = &cato_models.InternetFirewallSourceInput{}

					sourceInput := PolicyPolicyInternetFirewallPolicyRulesRuleSource{}
					diags = append(diags, itemExceptionsInput.Source.As(ctx, &sourceInput, basetypes.ObjectAsOptions{})...)

					// setting source IP
					if !sourceInput.IP.IsNull() {
						diags = append(diags, sourceInput.IP.ElementsAs(ctx, &exceptionAddInput.Source.IP, false)...)
						exceptionUpdateInput.Source.IP = exceptionAddInput.Source.IP
					} else {
						exceptionUpdateInput.Source.IP = make([]string, 0)
					}

					// setting source subnet
					if !sourceInput.Subnet.IsUnknown() && !sourceInput.Subnet.IsNull() {
						diags = append(diags, sourceInput.Subnet.ElementsAs(ctx, &exceptionAddInput.Source.Subnet, false)...)
						exceptionUpdateInput.Source.Subnet = exceptionAddInput.Source.Subnet
					} else {
						exceptionUpdateInput.Source.Subnet = make([]string, 0)
					}

					// setting source host
					if !sourceInput.Host.IsNull() {
						elementsSourceHostInput := make([]types.Object, 0, len(sourceInput.Host.Elements()))
						diags = append(diags, sourceInput.Host.ElementsAs(ctx, &elementsSourceHostInput, false)...)

						var itemSourceHostInput PolicyPolicyInternetFirewallPolicyRulesRuleSourceHost
						for _, item := range elementsSourceHostInput {
							diags = append(diags, item.As(ctx, &itemSourceHostInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceHostInput)
							if err != nil {
								diags = appendObjectRefTransformError(diags, "rule.exceptions.source.host", err)
								return hydrateAPIReturn, diags
							}

							exceptionAddInput.Source.Host = append(exceptionAddInput.Source.Host, &cato_models.HostRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
						exceptionUpdateInput.Source.Host = exceptionAddInput.Source.Host
					} else {
						exceptionUpdateInput.Source.Host = make([]*cato_models.HostRefInput, 0)
					}

					// setting source site
					if !sourceInput.Site.IsNull() {
						elementsSourceSiteInput := make([]types.Object, 0, len(sourceInput.Site.Elements()))
						diags = append(diags, sourceInput.Site.ElementsAs(ctx, &elementsSourceSiteInput, false)...)

						var itemSourceSiteInput PolicyPolicyInternetFirewallPolicyRulesRuleSourceSite
						for _, item := range elementsSourceSiteInput {
							diags = append(diags, item.As(ctx, &itemSourceSiteInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceSiteInput)
							if err != nil {
								diags = appendObjectRefTransformError(diags, "rule.exceptions.source.site", err)
								return hydrateAPIReturn, diags
							}

							exceptionAddInput.Source.Site = append(exceptionAddInput.Source.Site, &cato_models.SiteRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
						exceptionUpdateInput.Source.Site = exceptionAddInput.Source.Site
					} else {
						exceptionUpdateInput.Source.Site = make([]*cato_models.SiteRefInput, 0)
					}

					// setting source ip range
					if !sourceInput.IPRange.IsNull() {
						elementsSourceIPRangeInput := make([]types.Object, 0, len(sourceInput.IPRange.Elements()))
						diags = append(diags, sourceInput.IPRange.ElementsAs(ctx, &elementsSourceIPRangeInput, false)...)

						var itemSourceIPRangeInput PolicyPolicyInternetFirewallPolicyRulesRuleSourceIPRange
						for _, item := range elementsSourceIPRangeInput {
							diags = append(diags, item.As(ctx, &itemSourceIPRangeInput, basetypes.ObjectAsOptions{})...)

							exceptionAddInput.Source.IPRange = append(exceptionAddInput.Source.IPRange, &cato_models.IPAddressRangeInput{
								From: itemSourceIPRangeInput.From.ValueString(),
								To:   itemSourceIPRangeInput.To.ValueString(),
							})
						}
						exceptionUpdateInput.Source.IPRange = exceptionAddInput.Source.IPRange
					} else {
						exceptionUpdateInput.Source.IPRange = make([]*cato_models.IPAddressRangeInput, 0)
					}

					// setting source global ip range
					if !sourceInput.GlobalIPRange.IsNull() {
						elementsSourceGlobalIPRangeInput := make([]types.Object, 0, len(sourceInput.GlobalIPRange.Elements()))
						diags = append(diags, sourceInput.GlobalIPRange.ElementsAs(ctx, &elementsSourceGlobalIPRangeInput, false)...)

						var itemSourceGlobalIPRangeInput PolicyPolicyInternetFirewallPolicyRulesRuleSourceGlobalIPRange
						for _, item := range elementsSourceGlobalIPRangeInput {
							diags = append(diags, item.As(ctx, &itemSourceGlobalIPRangeInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceGlobalIPRangeInput)
							if err != nil {
								diags = appendObjectRefTransformError(diags, "rule.exceptions.source.global_ip_range", err)
								return hydrateAPIReturn, diags
							}

							exceptionAddInput.Source.GlobalIPRange = append(exceptionAddInput.Source.GlobalIPRange, &cato_models.GlobalIPRangeRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
						exceptionUpdateInput.Source.GlobalIPRange = exceptionAddInput.Source.GlobalIPRange
					} else {
						exceptionUpdateInput.Source.GlobalIPRange = make([]*cato_models.GlobalIPRangeRefInput, 0)
					}

					// setting source network interface
					if !sourceInput.NetworkInterface.IsNull() {
						elementsSourceNetworkInterfaceInput := make([]types.Object, 0, len(sourceInput.NetworkInterface.Elements()))
						diags = append(diags, sourceInput.NetworkInterface.ElementsAs(ctx, &elementsSourceNetworkInterfaceInput, false)...)

						var itemSourceNetworkInterfaceInput PolicyPolicyInternetFirewallPolicyRulesRuleSourceNetworkInterface
						for _, item := range elementsSourceNetworkInterfaceInput {
							diags = append(diags, item.As(ctx, &itemSourceNetworkInterfaceInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceNetworkInterfaceInput)
							if err != nil {
								diags = appendObjectRefTransformError(diags, "rule.exceptions.source.network_interface", err)
								return hydrateAPIReturn, diags
							}

							exceptionAddInput.Source.NetworkInterface = append(exceptionAddInput.Source.NetworkInterface, &cato_models.NetworkInterfaceRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
						exceptionUpdateInput.Source.NetworkInterface = exceptionAddInput.Source.NetworkInterface
					} else {
						exceptionUpdateInput.Source.NetworkInterface = make([]*cato_models.NetworkInterfaceRefInput, 0)
					}

					// setting source site network subnet
					if !sourceInput.SiteNetworkSubnet.IsNull() {
						elementsSourceSiteNetworkSubnetInput := make([]types.Object, 0, len(sourceInput.SiteNetworkSubnet.Elements()))
						diags = append(diags, sourceInput.SiteNetworkSubnet.ElementsAs(ctx, &elementsSourceSiteNetworkSubnetInput, false)...)

						var itemSourceSiteNetworkSubnetInput PolicyPolicyInternetFirewallPolicyRulesRuleSourceSiteNetworkSubnet
						for _, item := range elementsSourceSiteNetworkSubnetInput {
							diags = append(diags, item.As(ctx, &itemSourceSiteNetworkSubnetInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceSiteNetworkSubnetInput)
							if err != nil {
								diags = appendObjectRefTransformError(diags, "rule.exceptions.source.site_network_subnet", err)
								return hydrateAPIReturn, diags
							}

							exceptionAddInput.Source.SiteNetworkSubnet = append(
								exceptionAddInput.Source.SiteNetworkSubnet,
								&cato_models.SiteNetworkSubnetRefInput{
									By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
									Input: ObjectRefOutput.Input,
								},
							)
						}
						exceptionUpdateInput.Source.SiteNetworkSubnet = exceptionAddInput.Source.SiteNetworkSubnet
					} else {
						exceptionUpdateInput.Source.SiteNetworkSubnet = make([]*cato_models.SiteNetworkSubnetRefInput, 0)
					}

					// setting source floating subnet
					if !sourceInput.FloatingSubnet.IsNull() {
						elementsSourceFloatingSubnetInput := make([]types.Object, 0, len(sourceInput.FloatingSubnet.Elements()))
						diags = append(diags, sourceInput.FloatingSubnet.ElementsAs(ctx, &elementsSourceFloatingSubnetInput, false)...)

						var itemSourceFloatingSubnetInput PolicyPolicyInternetFirewallPolicyRulesRuleSourceFloatingSubnet
						for _, item := range elementsSourceFloatingSubnetInput {
							diags = append(diags, item.As(ctx, &itemSourceFloatingSubnetInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceFloatingSubnetInput)
							if err != nil {
								diags = appendObjectRefTransformError(diags, "rule.exceptions.source.floating_subnet", err)
								return hydrateAPIReturn, diags
							}

							exceptionAddInput.Source.FloatingSubnet = append(exceptionAddInput.Source.FloatingSubnet, &cato_models.FloatingSubnetRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
						exceptionUpdateInput.Source.FloatingSubnet = exceptionAddInput.Source.FloatingSubnet
					} else {
						exceptionUpdateInput.Source.FloatingSubnet = make([]*cato_models.FloatingSubnetRefInput, 0)
					}

					// setting source user
					if !sourceInput.User.IsNull() {
						elementsSourceUserInput := make([]types.Object, 0, len(sourceInput.User.Elements()))
						diags = append(diags, sourceInput.User.ElementsAs(ctx, &elementsSourceUserInput, false)...)

						var itemSourceUserInput PolicyPolicyInternetFirewallPolicyRulesRuleSourceUser
						for _, item := range elementsSourceUserInput {
							diags = append(diags, item.As(ctx, &itemSourceUserInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceUserInput)
							if err != nil {
								diags = appendObjectRefTransformError(diags, "rule.exceptions.source.user", err)
								return hydrateAPIReturn, diags
							}

							exceptionAddInput.Source.User = append(exceptionAddInput.Source.User, &cato_models.UserRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
						exceptionUpdateInput.Source.User = exceptionAddInput.Source.User
					} else {
						exceptionUpdateInput.Source.User = make([]*cato_models.UserRefInput, 0)
					}

					// setting source users group
					if !sourceInput.UsersGroup.IsNull() {
						elementsSourceUsersGroupInput := make([]types.Object, 0, len(sourceInput.UsersGroup.Elements()))
						diags = append(diags, sourceInput.UsersGroup.ElementsAs(ctx, &elementsSourceUsersGroupInput, false)...)

						var itemSourceUsersGroupInput PolicyPolicyInternetFirewallPolicyRulesRuleSourceUsersGroup
						for _, item := range elementsSourceUsersGroupInput {
							diags = append(diags, item.As(ctx, &itemSourceUsersGroupInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceUsersGroupInput)
							if err != nil {
								diags = appendObjectRefTransformError(diags, "rule.exceptions.source.users_group", err)
								return hydrateAPIReturn, diags
							}

							exceptionAddInput.Source.UsersGroup = append(exceptionAddInput.Source.UsersGroup, &cato_models.UsersGroupRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
						exceptionUpdateInput.Source.UsersGroup = exceptionAddInput.Source.UsersGroup
					} else {
						exceptionUpdateInput.Source.UsersGroup = make([]*cato_models.UsersGroupRefInput, 0)
					}

					// setting source group
					if !sourceInput.Group.IsNull() {
						elementsSourceGroupInput := make([]types.Object, 0, len(sourceInput.Group.Elements()))
						diags = append(diags, sourceInput.Group.ElementsAs(ctx, &elementsSourceGroupInput, false)...)

						var itemSourceGroupInput PolicyPolicyInternetFirewallPolicyRulesRuleSourceGroup
						for _, item := range elementsSourceGroupInput {
							diags = append(diags, item.As(ctx, &itemSourceGroupInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceGroupInput)
							if err != nil {
								diags = appendObjectRefTransformError(diags, "rule.exceptions.source.group", err)
								return hydrateAPIReturn, diags
							}

							exceptionAddInput.Source.Group = append(exceptionAddInput.Source.Group, &cato_models.GroupRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
						exceptionUpdateInput.Source.Group = exceptionAddInput.Source.Group
					} else {
						exceptionUpdateInput.Source.Group = make([]*cato_models.GroupRefInput, 0)
					}

					// setting source system group
					if !sourceInput.SystemGroup.IsNull() {
						elementsSourceSystemGroupInput := make([]types.Object, 0, len(sourceInput.SystemGroup.Elements()))
						diags = append(diags, sourceInput.SystemGroup.ElementsAs(ctx, &elementsSourceSystemGroupInput, false)...)

						var itemSourceSystemGroupInput PolicyPolicyInternetFirewallPolicyRulesRuleSourceSystemGroup
						for _, item := range elementsSourceSystemGroupInput {
							diags = append(diags, item.As(ctx, &itemSourceSystemGroupInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceSystemGroupInput)
							if err != nil {
								diags = appendObjectRefTransformError(diags, "rule.exceptions.source.system_group", err)
								return hydrateAPIReturn, diags
							}

							exceptionAddInput.Source.SystemGroup = append(exceptionAddInput.Source.SystemGroup, &cato_models.SystemGroupRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
						exceptionUpdateInput.Source.SystemGroup = exceptionAddInput.Source.SystemGroup
					} else {
						exceptionUpdateInput.Source.SystemGroup = make([]*cato_models.SystemGroupRefInput, 0)
					}
				}

				// setting country
				if !itemExceptionsInput.Country.IsNull() {
					exceptionAddInput.Country = []*cato_models.CountryRefInput{}
					elementsCountryInput := make([]types.Object, 0, len(itemExceptionsInput.Country.Elements()))
					diags = append(diags, itemExceptionsInput.Country.ElementsAs(ctx, &elementsCountryInput, false)...)

					var itemCountryInput PolicyPolicyInternetFirewallPolicyRulesRuleCountry
					for _, item := range elementsCountryInput {
						diags = append(diags, item.As(ctx, &itemCountryInput, basetypes.ObjectAsOptions{})...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemCountryInput)
						if err != nil {
							diags = appendObjectRefTransformError(diags, "rule.exceptions.country", err)
							return hydrateAPIReturn, diags
						}

						exceptionAddInput.Country = append(exceptionAddInput.Country, &cato_models.CountryRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
					exceptionUpdateInput.Country = exceptionAddInput.Country
				} else {
					exceptionUpdateInput.Country = []*cato_models.CountryRefInput{}
				}

				// setting device
				if !itemExceptionsInput.Device.IsNull() {
					exceptionAddInput.Device = []*cato_models.DeviceProfileRefInput{}
					elementsDeviceInput := make([]types.Object, 0, len(itemExceptionsInput.Device.Elements()))
					diags = append(diags, itemExceptionsInput.Device.ElementsAs(ctx, &elementsDeviceInput, false)...)

					var itemDeviceInput PolicyPolicyInternetFirewallPolicyRulesRuleDevice
					for _, item := range elementsDeviceInput {
						diags = append(diags, item.As(ctx, &itemDeviceInput, basetypes.ObjectAsOptions{})...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemDeviceInput)
						if err != nil {
							diags = appendObjectRefTransformError(diags, "rule.exceptions.device", err)
							return hydrateAPIReturn, diags
						}

						exceptionAddInput.Device = append(exceptionAddInput.Device, &cato_models.DeviceProfileRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
					exceptionUpdateInput.Device = exceptionAddInput.Device
				} else {
					exceptionUpdateInput.Device = []*cato_models.DeviceProfileRefInput{}
				}

				// setting device OS
				if !itemExceptionsInput.DeviceOs.IsUnknown() && !itemExceptionsInput.DeviceOs.IsNull() {
					diags = append(diags, itemExceptionsInput.DeviceOs.ElementsAs(ctx, &exceptionAddInput.DeviceOs, false)...)
					exceptionUpdateInput.DeviceOs = exceptionAddInput.DeviceOs
				} else {
					exceptionUpdateInput.DeviceOs = make([]cato_models.OperatingSystem, 0)
				}

				// setting device_attributes
				if !itemExceptionsInput.DeviceAttributes.IsNull() {
					exceptionAddInput.DeviceAttributes = &cato_models.DeviceAttributesInput{}
					exceptionUpdateInput.DeviceAttributes = &cato_models.DeviceAttributesInput{}

					deviceAttributesInput := PolicyPolicyInternetFirewallPolicyRulesRuleDeviceAttributes{}
					diags = append(diags, itemExceptionsInput.DeviceAttributes.As(ctx, &deviceAttributesInput, basetypes.ObjectAsOptions{})...)

					// setting device_attributes category
					if !deviceAttributesInput.Category.IsUnknown() && !deviceAttributesInput.Category.IsNull() {
						diags = append(diags, deviceAttributesInput.Category.ElementsAs(ctx, &exceptionAddInput.DeviceAttributes.Category, false)...)
						exceptionUpdateInput.DeviceAttributes.Category = exceptionAddInput.DeviceAttributes.Category
					} else {
						exceptionUpdateInput.DeviceAttributes.Category = make([]string, 0)
					}

					// setting device_attributes manufacturer
					if !deviceAttributesInput.Manufacturer.IsUnknown() && !deviceAttributesInput.Manufacturer.IsNull() {
						diags = append(diags, deviceAttributesInput.Manufacturer.ElementsAs(ctx, &exceptionAddInput.DeviceAttributes.Manufacturer, false)...)
						exceptionUpdateInput.DeviceAttributes.Manufacturer = exceptionAddInput.DeviceAttributes.Manufacturer
					} else {
						exceptionUpdateInput.DeviceAttributes.Manufacturer = make([]string, 0)
					}

					// setting device_attributes model
					if !deviceAttributesInput.Model.IsUnknown() && !deviceAttributesInput.Model.IsNull() {
						diags = append(diags, deviceAttributesInput.Model.ElementsAs(ctx, &exceptionAddInput.DeviceAttributes.Model, false)...)
						exceptionUpdateInput.DeviceAttributes.Model = exceptionAddInput.DeviceAttributes.Model
					} else {
						exceptionUpdateInput.DeviceAttributes.Model = make([]string, 0)
					}

					// setting device_attributes os
					if !deviceAttributesInput.Os.IsUnknown() && !deviceAttributesInput.Os.IsNull() {
						diags = append(diags, deviceAttributesInput.Os.ElementsAs(ctx, &exceptionAddInput.DeviceAttributes.Os, false)...)
						exceptionUpdateInput.DeviceAttributes.Os = exceptionAddInput.DeviceAttributes.Os
					} else {
						exceptionUpdateInput.DeviceAttributes.Os = make([]string, 0)
					}

					// setting device_attributes os_version
					if !deviceAttributesInput.OsVersion.IsUnknown() && !deviceAttributesInput.OsVersion.IsNull() {
						diags = append(diags, deviceAttributesInput.OsVersion.ElementsAs(ctx, &exceptionAddInput.DeviceAttributes.OsVersion, false)...)
						exceptionUpdateInput.DeviceAttributes.OsVersion = exceptionAddInput.DeviceAttributes.OsVersion
					} else {
						exceptionUpdateInput.DeviceAttributes.OsVersion = make([]string, 0)
					}

					// setting device_attributes type
					if !deviceAttributesInput.Type.IsUnknown() && !deviceAttributesInput.Type.IsNull() {
						diags = append(diags, deviceAttributesInput.Type.ElementsAs(ctx, &exceptionAddInput.DeviceAttributes.Type, false)...)
						exceptionUpdateInput.DeviceAttributes.Type = exceptionAddInput.DeviceAttributes.Type
					} else {
						exceptionUpdateInput.DeviceAttributes.Type = make([]string, 0)
					}
				} else {
					emptyExceptionDeviceAttributes := &cato_models.DeviceAttributesInput{
						Category:     make([]string, 0),
						Manufacturer: make([]string, 0),
						Model:        make([]string, 0),
						Os:           make([]string, 0),
						OsVersion:    make([]string, 0),
						Type:         make([]string, 0),
					}
					exceptionUpdateInput.DeviceAttributes = emptyExceptionDeviceAttributes
					exceptionAddInput.DeviceAttributes = emptyExceptionDeviceAttributes
				}

				// setting destination
				if !itemExceptionsInput.Destination.IsNull() {
					exceptionAddInput.Destination = &cato_models.InternetFirewallDestinationInput{}
					exceptionUpdateInput.Destination = &cato_models.InternetFirewallDestinationInput{}
					exceptionAddInput.Destination.Containers = &cato_models.InternetFirewallContainerInput{
						FqdnContainer:           make([]*cato_models.FqdnContainerRefInput, 0),
						IPAddressRangeContainer: make([]*cato_models.IPAddressRangeContainerRefInput, 0),
					}
					exceptionUpdateInput.Destination.Containers = &cato_models.InternetFirewallContainerInput{
						FqdnContainer:           make([]*cato_models.FqdnContainerRefInput, 0),
						IPAddressRangeContainer: make([]*cato_models.IPAddressRangeContainerRefInput, 0),
					}

					destinationInput := PolicyPolicyInternetFirewallPolicyRulesRuleDestination{}
					diags = append(diags, itemExceptionsInput.Destination.As(ctx, &destinationInput, basetypes.ObjectAsOptions{})...)

					// setting destination IP
					if !destinationInput.IP.IsUnknown() && !destinationInput.IP.IsNull() {
						diags = append(diags, destinationInput.IP.ElementsAs(ctx, &exceptionAddInput.Destination.IP, false)...)
						exceptionUpdateInput.Destination.IP = exceptionAddInput.Destination.IP
					} else {
						exceptionUpdateInput.Destination.IP = make([]string, 0)
					}

					// setting destination subnet
					if !destinationInput.Subnet.IsUnknown() && !destinationInput.Subnet.IsNull() {
						diags = append(diags, destinationInput.Subnet.ElementsAs(ctx, &exceptionAddInput.Destination.Subnet, false)...)
						exceptionUpdateInput.Destination.Subnet = exceptionAddInput.Destination.Subnet
					} else {
						exceptionUpdateInput.Destination.Subnet = make([]string, 0)
					}

					// setting destination domain
					if !destinationInput.Domain.IsUnknown() && !destinationInput.Domain.IsNull() {
						diags = append(diags, destinationInput.Domain.ElementsAs(ctx, &exceptionAddInput.Destination.Domain, false)...)
						exceptionUpdateInput.Destination.Domain = exceptionAddInput.Destination.Domain
					} else {
						exceptionUpdateInput.Destination.Domain = make([]string, 0)
					}

					// setting destination fqdn
					if !destinationInput.Fqdn.IsUnknown() && !destinationInput.Fqdn.IsNull() {
						diags = append(diags, destinationInput.Fqdn.ElementsAs(ctx, &exceptionAddInput.Destination.Fqdn, false)...)
						exceptionUpdateInput.Destination.Fqdn = exceptionAddInput.Destination.Fqdn
					} else {
						exceptionUpdateInput.Destination.Fqdn = make([]string, 0)
					}

					// setting destination remote asn
					if !destinationInput.RemoteAsn.IsUnknown() && !destinationInput.RemoteAsn.IsNull() {
						diags = append(diags, destinationInput.RemoteAsn.ElementsAs(ctx, &exceptionAddInput.Destination.RemoteAsn, false)...)
						exceptionUpdateInput.Destination.RemoteAsn = exceptionAddInput.Destination.RemoteAsn
					} else {
						exceptionUpdateInput.Destination.RemoteAsn = make([]cato_scalars.Asn32, 0)
					}

					// setting destination application
					if !destinationInput.Application.IsNull() {
						elementsDestinationApplicationInput := make([]types.Object, 0, len(destinationInput.Application.Elements()))
						diags = append(diags, destinationInput.Application.ElementsAs(ctx, &elementsDestinationApplicationInput, false)...)

						var itemDestinationApplicationInput PolicyPolicyInternetFirewallPolicyRulesRuleDestinationApplication
						for _, item := range elementsDestinationApplicationInput {
							diags = append(diags, item.As(ctx, &itemDestinationApplicationInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationApplicationInput)
							if err != nil {
								diags = appendObjectRefTransformError(diags, "rule.exceptions.destination.application", err)
								return hydrateAPIReturn, diags
							}

							exceptionAddInput.Destination.Application = append(exceptionAddInput.Destination.Application, &cato_models.ApplicationRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
						exceptionUpdateInput.Destination.Application = exceptionAddInput.Destination.Application
					} else {
						exceptionUpdateInput.Destination.Application = make([]*cato_models.ApplicationRefInput, 0)
					}

					// setting destination custom app
					if !destinationInput.CustomApp.IsNull() {
						elementsDestinationCustomAppInput := make([]types.Object, 0, len(destinationInput.CustomApp.Elements()))
						diags = append(diags, destinationInput.CustomApp.ElementsAs(ctx, &elementsDestinationCustomAppInput, false)...)

						var itemDestinationCustomAppInput PolicyPolicyInternetFirewallPolicyRulesRuleDestinationCustomApp
						for _, item := range elementsDestinationCustomAppInput {
							diags = append(diags, item.As(ctx, &itemDestinationCustomAppInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationCustomAppInput)
							if err != nil {
								diags = appendObjectRefTransformError(diags, "rule.exceptions.destination.custom_app", err)
								return hydrateAPIReturn, diags
							}

							exceptionAddInput.Destination.CustomApp = append(exceptionAddInput.Destination.CustomApp, &cato_models.CustomApplicationRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
						exceptionUpdateInput.Destination.CustomApp = exceptionAddInput.Destination.CustomApp
					} else {
						exceptionUpdateInput.Destination.CustomApp = make([]*cato_models.CustomApplicationRefInput, 0)
					}

					// setting destination ip range
					if !destinationInput.IPRange.IsNull() {
						elementsDestinationIPRangeInput := make([]types.Object, 0, len(destinationInput.IPRange.Elements()))
						diags = append(diags, destinationInput.IPRange.ElementsAs(ctx, &elementsDestinationIPRangeInput, false)...)

						var itemDestinationIPRangeInput iprange.Range
						for _, item := range elementsDestinationIPRangeInput {
							diags = append(diags, item.As(ctx, &itemDestinationIPRangeInput, basetypes.ObjectAsOptions{})...)

							exceptionAddInput.Destination.IPRange = append(exceptionAddInput.Destination.IPRange, &cato_models.IPAddressRangeInput{
								From: itemDestinationIPRangeInput.From.ValueString(),
								To:   itemDestinationIPRangeInput.To.ValueString(),
							})
						}
						exceptionUpdateInput.Destination.IPRange = exceptionAddInput.Destination.IPRange
					} else {
						exceptionUpdateInput.Destination.IPRange = make([]*cato_models.IPAddressRangeInput, 0)
					}

					// setting destination global ip range
					if !destinationInput.GlobalIPRange.IsNull() {
						elementsDestinationGlobalIPRangeInput := make([]types.Object, 0, len(destinationInput.GlobalIPRange.Elements()))
						diags = append(diags, destinationInput.GlobalIPRange.ElementsAs(ctx, &elementsDestinationGlobalIPRangeInput, false)...)

						var itemDestinationGlobalIPRangeInput PolicyPolicyInternetFirewallPolicyRulesRuleDestinationGlobalIPRange
						for _, item := range elementsDestinationGlobalIPRangeInput {
							diags = append(diags, item.As(ctx, &itemDestinationGlobalIPRangeInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationGlobalIPRangeInput)
							if err != nil {
								diags = appendObjectRefTransformError(diags, "rule.exceptions.destination.global_ip_range", err)
								return hydrateAPIReturn, diags
							}

							exceptionAddInput.Destination.GlobalIPRange = append(exceptionAddInput.Destination.GlobalIPRange, &cato_models.GlobalIPRangeRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
						exceptionUpdateInput.Destination.GlobalIPRange = exceptionAddInput.Destination.GlobalIPRange
					} else {
						exceptionUpdateInput.Destination.GlobalIPRange = make([]*cato_models.GlobalIPRangeRefInput, 0)
					}

					// setting destination app category
					if !destinationInput.AppCategory.IsNull() {
						elementsDestinationAppCategoryInput := make([]types.Object, 0, len(destinationInput.AppCategory.Elements()))
						diags = append(diags, destinationInput.AppCategory.ElementsAs(ctx, &elementsDestinationAppCategoryInput, false)...)

						var itemDestinationAppCategoryInput PolicyPolicyInternetFirewallPolicyRulesRuleDestinationAppCategory
						for _, item := range elementsDestinationAppCategoryInput {
							diags = append(diags, item.As(ctx, &itemDestinationAppCategoryInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationAppCategoryInput)
							if err != nil {
								diags = appendObjectRefTransformError(diags, "rule.exceptions.destination.app_category", err)
								return hydrateAPIReturn, diags
							}

							exceptionAddInput.Destination.AppCategory = append(
								exceptionAddInput.Destination.AppCategory,
								&cato_models.ApplicationCategoryRefInput{
									By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
									Input: ObjectRefOutput.Input,
								},
							)
						}
						exceptionUpdateInput.Destination.AppCategory = exceptionAddInput.Destination.AppCategory
					} else {
						exceptionUpdateInput.Destination.AppCategory = make([]*cato_models.ApplicationCategoryRefInput, 0)
					}

					// setting destination custom app category
					if !destinationInput.CustomCategory.IsNull() {
						elementsDestinationCustomCategoryInput := make([]types.Object, 0, len(destinationInput.CustomCategory.Elements()))
						diags = append(diags, destinationInput.CustomCategory.ElementsAs(ctx, &elementsDestinationCustomCategoryInput, false)...)

						var itemDestinationCustomCategoryInput PolicyPolicyInternetFirewallPolicyRulesRuleDestinationCustomCategory
						for _, item := range elementsDestinationCustomCategoryInput {
							diags = append(diags, item.As(ctx, &itemDestinationCustomCategoryInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationCustomCategoryInput)
							if err != nil {
								diags = appendObjectRefTransformError(diags, "rule.exceptions.destination.custom_category", err)
								return hydrateAPIReturn, diags
							}

							exceptionAddInput.Destination.CustomCategory = append(
								exceptionAddInput.Destination.CustomCategory,
								&cato_models.CustomCategoryRefInput{
									By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
									Input: ObjectRefOutput.Input,
								},
							)
						}
						exceptionUpdateInput.Destination.CustomCategory = exceptionAddInput.Destination.CustomCategory
					} else {
						exceptionUpdateInput.Destination.CustomCategory = make([]*cato_models.CustomCategoryRefInput, 0)
					}

					// setting destination sanctioned apps category
					if !destinationInput.SanctionedAppsCategory.IsNull() {
						elementsDestinationSanctionedAppsCategoryInput := make([]types.Object, 0, len(destinationInput.SanctionedAppsCategory.Elements()))
						diags = append(
							diags,
							destinationInput.SanctionedAppsCategory.ElementsAs(ctx, &elementsDestinationSanctionedAppsCategoryInput, false)...,
						)

						var itemDestinationSanctionedAppsCategoryInput PolicyPolicyInternetFirewallPolicyRulesRuleDestinationSanctionedAppsCategory
						for _, item := range elementsDestinationSanctionedAppsCategoryInput {
							diags = append(diags, item.As(ctx, &itemDestinationSanctionedAppsCategoryInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationSanctionedAppsCategoryInput)
							if err != nil {
								diags = appendObjectRefTransformError(diags, "rule.exceptions.destination.sanctioned_apps_category", err)
								return hydrateAPIReturn, diags
							}

							exceptionAddInput.Destination.SanctionedAppsCategory = append(
								exceptionAddInput.Destination.SanctionedAppsCategory,
								&cato_models.SanctionedAppsCategoryRefInput{
									By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
									Input: ObjectRefOutput.Input,
								},
							)
						}
						exceptionUpdateInput.Destination.SanctionedAppsCategory = exceptionAddInput.Destination.SanctionedAppsCategory
					} else {
						exceptionUpdateInput.Destination.SanctionedAppsCategory = make([]*cato_models.SanctionedAppsCategoryRefInput, 0)
					}

					// setting destination country
					if !destinationInput.Country.IsNull() {
						elementsDestinationCountryInput := make([]types.Object, 0, len(destinationInput.Country.Elements()))
						diags = append(diags, destinationInput.Country.ElementsAs(ctx, &elementsDestinationCountryInput, false)...)

						var itemDestinationCountryInput PolicyPolicyInternetFirewallPolicyRulesRuleDestinationCountry
						for _, item := range elementsDestinationCountryInput {
							diags = append(diags, item.As(ctx, &itemDestinationCountryInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationCountryInput)
							if err != nil {
								diags = appendObjectRefTransformError(diags, "rule.exceptions.destination.country", err)
								return hydrateAPIReturn, diags
							}

							exceptionAddInput.Destination.Country = append(exceptionAddInput.Destination.Country, &cato_models.CountryRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
						exceptionUpdateInput.Destination.Country = exceptionAddInput.Destination.Country
					} else {
						exceptionUpdateInput.Destination.Country = make([]*cato_models.CountryRefInput, 0)
					}
				} else {
					exceptionUpdateInput.Destination = &cato_models.InternetFirewallDestinationInput{
						Containers: &cato_models.InternetFirewallContainerInput{
							FqdnContainer:           make([]*cato_models.FqdnContainerRefInput, 0),
							IPAddressRangeContainer: make([]*cato_models.IPAddressRangeContainerRefInput, 0),
						},
					}
				}

				// setting service
				if !itemExceptionsInput.Service.IsNull() {
					exceptionAddInput.Service = &cato_models.InternetFirewallServiceTypeInput{}
					exceptionUpdateInput.Service = &cato_models.InternetFirewallServiceTypeInput{}

					serviceInput := PolicyPolicyInternetFirewallPolicyRulesRuleService{}
					diags = append(diags, itemExceptionsInput.Service.As(ctx, &serviceInput, basetypes.ObjectAsOptions{})...)

					// setting service standard
					if !serviceInput.Standard.IsNull() {
						elementsServiceStandardInput := make([]types.Object, 0, len(serviceInput.Standard.Elements()))
						diags = append(diags, serviceInput.Standard.ElementsAs(ctx, &elementsServiceStandardInput, false)...)

						var itemServiceStandardInput PolicyPolicyInternetFirewallPolicyRulesRuleServiceStandard
						for _, item := range elementsServiceStandardInput {
							diags = append(diags, item.As(ctx, &itemServiceStandardInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemServiceStandardInput)
							if err != nil {
								diags = appendObjectRefTransformError(diags, "rule.exceptions.service.standard", err)
								return hydrateAPIReturn, diags
							}

							exceptionAddInput.Service.Standard = append(exceptionAddInput.Service.Standard, &cato_models.ServiceRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
						exceptionUpdateInput.Service.Standard = exceptionAddInput.Service.Standard
					} else {
						exceptionUpdateInput.Service.Standard = make([]*cato_models.ServiceRefInput, 0)
					}

					// setting service custom
					if !serviceInput.Custom.IsNull() {
						elementsServiceCustomInput := make([]types.Object, 0, len(serviceInput.Custom.Elements()))
						diags = append(diags, serviceInput.Custom.ElementsAs(ctx, &elementsServiceCustomInput, false)...)

						var itemServiceCustomInput customservice.CustomService
						for _, item := range elementsServiceCustomInput {
							diags = append(diags, item.As(ctx, &itemServiceCustomInput, basetypes.ObjectAsOptions{})...)

							tflog.Debug(ctx, "Processing exception custom service entry", map[string]interface{}{
								"port_is_null":       itemServiceCustomInput.Port.IsNull(),
								"port_range_is_null": itemServiceCustomInput.PortRange.IsNull(),
								"protocol":           itemServiceCustomInput.Protocol.ValueString(),
							})

							customInput := &cato_models.CustomServiceInput{
								Protocol: cato_models.IPProtocol(itemServiceCustomInput.Protocol.ValueString()),
							}

							// setting service custom port
							if !itemServiceCustomInput.Port.IsUnknown() && !itemServiceCustomInput.Port.IsNull() {
								tflog.Debug(ctx, "Processing exception port field")
								var elementsPort []types.String
								diags = append(diags, itemServiceCustomInput.Port.ElementsAs(ctx, &elementsPort, false)...)

								inputPort := make([]cato_scalars.Port, 0, len(elementsPort))
								for _, item := range elementsPort {
									inputPort = append(inputPort, cato_scalars.Port(item.ValueString()))
								}
								customInput.Port = inputPort
								tflog.Debug(ctx, "Set exception port field", map[string]interface{}{
									"port_count": len(inputPort),
								})
							} else {
								tflog.Info(ctx, "Port is either unknown or null; skipping assignment")
							}

							// setting service custom port range
							if !itemServiceCustomInput.PortRange.IsNull() {
								tflog.Debug(ctx, "Processing exception port_range field")
								var itemPortRange portrange.PortRange
								diags = append(diags, itemServiceCustomInput.PortRange.As(ctx, &itemPortRange, basetypes.ObjectAsOptions{})...)

								inputPortRange := cato_models.PortRangeInput{
									From: cato_scalars.Port(itemPortRange.From.ValueString()),
									To:   cato_scalars.Port(itemPortRange.To.ValueString()),
								}

								customInput.PortRange = &inputPortRange
								tflog.Debug(ctx, "Set exception port_range field", map[string]interface{}{
									"from": itemPortRange.From.ValueString(),
									"to":   itemPortRange.To.ValueString(),
								})
							}

							// append custom service
							exceptionAddInput.Service.Custom = append(exceptionAddInput.Service.Custom, customInput)
						}
						exceptionUpdateInput.Service.Custom = exceptionAddInput.Service.Custom
					} else {
						exceptionUpdateInput.Service.Custom = make([]*cato_models.CustomServiceInput, 0)
					}
				} else {
					exceptionAddInput.Service = &cato_models.InternetFirewallServiceTypeInput{
						Standard: make([]*cato_models.ServiceRefInput, 0),
						Custom:   make([]*cato_models.CustomServiceInput, 0),
					}
					exceptionUpdateInput.Service = &cato_models.InternetFirewallServiceTypeInput{
						Standard: make([]*cato_models.ServiceRefInput, 0),
						Custom:   make([]*cato_models.CustomServiceInput, 0),
					}
				}

				rootAddRule.Exceptions = append(rootAddRule.Exceptions, &exceptionAddInput)
				rootUpdateRule.Exceptions = append(rootUpdateRule.Exceptions, &exceptionUpdateInput)
			}
		}

		// setting activePeriod
		if !ruleInput.ActivePeriod.IsNull() && !ruleInput.ActivePeriod.IsUnknown() {
			activePeriodInput := activeperiod.ActivePeriod{}
			diags = append(diags, ruleInput.ActivePeriod.As(ctx, &activePeriodInput, basetypes.ObjectAsOptions{})...)

			activePeriodAPIInput := &cato_models.PolicyRuleActivePeriodInput{
				EffectiveFrom:    nil,
				ExpiresAt:        nil,
				UseEffectiveFrom: false,
				UseExpiresAt:     false,
			}

			// Handle effective_from
			if !activePeriodInput.EffectiveFrom.IsNull() && !activePeriodInput.EffectiveFrom.IsUnknown() {
				effectiveFromStr := activePeriodInput.EffectiveFrom.ValueString()
				if effectiveFromStr != "" {
					// Parse the time string - support both RFC3339 format and human-readable formats
					parsedTime, err := utils.ParseTimeStringWithTZ(effectiveFromStr)
					if err != nil {
						diags = append(diags, diag.NewErrorDiagnostic(
							"Invalid effective_from time format",
							fmt.Sprintf(
								"Unable to parse effective_from time '%s': %v. Expected RFC3339 format (e.g., '2024-12-31T23:59:59Z')",
								effectiveFromStr,
								err,
							),
						))
					} else {
						activePeriodAPIInput.EffectiveFrom = &parsedTime
						activePeriodAPIInput.UseEffectiveFrom = true
					}
				}
			}

			// Handle expires_at
			if !activePeriodInput.ExpiresAt.IsNull() && !activePeriodInput.ExpiresAt.IsUnknown() {
				expiresAtStr := activePeriodInput.ExpiresAt.ValueString()
				if expiresAtStr != "" {
					// Parse the time string - support both RFC3339 format and human-readable formats
					parsedTime, err := utils.ParseTimeStringWithTZ(expiresAtStr)
					if err != nil {
						diags = append(diags, diag.NewErrorDiagnostic(
							"Invalid expires_at time format",
							fmt.Sprintf("Unable to parse expires_at time '%s': %v. Expected RFC3339 format (e.g., '2024-12-31T23:59:59Z')", expiresAtStr, err),
						))
					} else {
						activePeriodAPIInput.ExpiresAt = &parsedTime
						activePeriodAPIInput.UseExpiresAt = true
					}
				}
			}

			rootAddRule.ActivePeriod = activePeriodAPIInput
			// For update, we need to create a separate struct with pointer fields for bools
			activePeriodUpdateAPIInput := &cato_models.PolicyRuleActivePeriodUpdateInput{
				EffectiveFrom:    activePeriodAPIInput.EffectiveFrom,
				ExpiresAt:        activePeriodAPIInput.ExpiresAt,
				UseEffectiveFrom: &activePeriodAPIInput.UseEffectiveFrom,
				UseExpiresAt:     &activePeriodAPIInput.UseExpiresAt,
			}
			rootUpdateRule.ActivePeriod = activePeriodUpdateAPIInput
		} else {
			// setting activePeriod with default values when not provided
			rootAddRule.ActivePeriod = &cato_models.PolicyRuleActivePeriodInput{
				EffectiveFrom:    nil,
				ExpiresAt:        nil,
				UseEffectiveFrom: false,
				UseExpiresAt:     false,
			}
			useEffectiveFromDefault := false
			useExpiresAtDefault := false
			rootUpdateRule.ActivePeriod = &cato_models.PolicyRuleActivePeriodUpdateInput{
				EffectiveFrom:    nil,
				ExpiresAt:        nil,
				UseEffectiveFrom: &useEffectiveFromDefault,
				UseExpiresAt:     &useExpiresAtDefault,
			}
		}

		// settings other rule attributes
		rootAddRule.Name = ruleInput.Name.ValueString()
		rootUpdateRule.Name = ruleInput.Name.ValueStringPointer()

		if !ruleInput.Description.IsNull() && !ruleInput.Description.IsUnknown() {
			rootAddRule.Description = ruleInput.Description.ValueString()
			rootUpdateRule.Description = ruleInput.Description.ValueStringPointer()
		}

		rootAddRule.Enabled = ruleInput.Enabled.ValueBool()
		rootUpdateRule.Enabled = ruleInput.Enabled.ValueBoolPointer()

		rootAddRule.Action = cato_models.InternetFirewallActionEnum(ruleInput.Action.ValueString())
		rootUpdateRule.Action = (*cato_models.InternetFirewallActionEnum)(ruleInput.Action.ValueStringPointer())
		rootAddRule.ActionConfig = &cato_models.InternetFirewallActionConfigInput{
			RbiProfile:       make([]*cato_models.RbiProfileRefInput, 0),
			UserNotification: make([]*cato_models.UserNotificationTemplateRefInput, 0),
		}

		if !ruleInput.ConnectionOrigin.IsNull() && !ruleInput.ConnectionOrigin.IsUnknown() {
			rootAddRule.ConnectionOrigin = cato_models.ConnectionOriginEnum(ruleInput.ConnectionOrigin.ValueString())
			rootUpdateRule.ConnectionOrigin = (*cato_models.ConnectionOriginEnum)(ruleInput.ConnectionOrigin.ValueStringPointer())
		} else {
			rootAddRule.ConnectionOrigin = cato_models.ConnectionOriginEnum(defaultConnectionOriginAny)
			connectionOrigin := defaultConnectionOriginAny
			rootUpdateRule.ConnectionOrigin = (*cato_models.ConnectionOriginEnum)(&connectionOrigin)
		}
	}

	hydrateAPIReturn.Create.Rule = rootAddRule
	hydrateAPIReturn.Update.Rule = rootUpdateRule

	return hydrateAPIReturn, diags
}

//nolint:gocyclo,funlen
func HydrateIfwRuleState(
	ctx context.Context,
	state InternetFirewallRule,
	currentRule *cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_Rule,
) PolicyPolicyInternetFirewallPolicyRulesRule {
	ruleInput := PolicyPolicyInternetFirewallPolicyRulesRule{}
	diags := make(diag.Diagnostics, 0)
	stateRuleInput := PolicyPolicyInternetFirewallPolicyRulesRule{}
	if !state.Rule.IsNull() && !state.Rule.IsUnknown() {
		diagstmp := state.Rule.As(ctx, &stateRuleInput, basetypes.ObjectAsOptions{})
		diags = append(diags, diagstmp...)
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
	ruleInput.ConnectionOrigin = types.StringValue(currentRule.ConnectionOrigin.String())

	// // /// /// /// / Rule -> Source // /// /// /// /// /
	curRuleSourceObj, diagstmp := types.ObjectValue(
		IfwSourceAttrTypes,
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
	curRuleSourceObjAttrs := curRuleSourceObj.Attributes()
	if !stateRuleInput.Source.IsNull() && !stateRuleInput.Source.IsUnknown() {
		stateSourceAttrs := stateRuleInput.Source.Attributes()
		curRuleSourceObjAttrs["users_group"] = preserveNullIDNameIDSet(
			curRuleSourceObjAttrs["users_group"].(types.Set),
			stateSourceAttrs["users_group"].(types.Set),
		)
	}
	curRuleSourceObj, diagstmp = types.ObjectValue(curRuleSourceObj.AttributeTypes(ctx), curRuleSourceObjAttrs)
	diags = append(diags, diagstmp...)
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

	// /// /// /// /// Start Rule -> deviceAttributes // /// /// /// /// /
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
		// Set device_attributes to null when no attributes are present
		deviceAttributesObj = types.ObjectNull(IfwDeviceAttrAttrTypes)
	}

	tflog.Debug(ctx, "IFW_rule.read.currentRule.DeviceAttributes", map[string]interface{}{
		"deviceAttributesObj": utils.InterfaceToJSONString(deviceAttributesObj.String()),
	})

	ruleInput.DeviceAttributes = deviceAttributesObj
	// /// /// /// /// End Rule -> deviceAttributes // /// /// /// /// /

	// /// /// /// / Rule -> Destination // /// /// /// /// /
	curRuleDestinationObj, diagstmp := types.ObjectValue(
		IfwDestAttrTypes,
		map[string]attr.Value{
			"application": oldidname.ParseNameIDList(ctx, currentRule.Destination.Application,
				"rule.destination.application"),
			"custom_app": oldidname.ParseNameIDList(ctx, currentRule.Destination.CustomApp,
				"rule.destination.custom_app"),
			"app_category": oldidname.ParseNameIDList(ctx, currentRule.Destination.AppCategory,
				"rule.destination.app_category"),
			"custom_category": oldidname.ParseNameIDList(ctx, currentRule.Destination.CustomCategory,
				"rule.destination.custom_category"),
			"sanctioned_apps_category": oldidname.ParseNameIDList(
				ctx,
				currentRule.Destination.SanctionedAppsCategory,
				"rule.destination.sanctioned_apps_category",
			),
			"country": oldidname.ParseNameIDList(ctx, currentRule.Destination.Country,
				"rule.destination.country"),
			"domain": convert.ParseList(ctx, types.StringType, currentRule.Destination.Domain,
				"rule.destination.domain"),
			"fqdn": convert.ParseList(ctx, types.StringType, currentRule.Destination.Fqdn,
				"rule.destination.fqdn"),
			"ip": convert.ParseList(ctx, types.StringType, currentRule.Destination.IP,
				"rule.destination.ip"),
			"subnet": convert.ParseList(ctx, types.StringType, currentRule.Destination.Subnet,
				"rule.destination.subnet"),
			"ip_range": convert.ParseFromToList(ctx, currentRule.Destination.IPRange,
				"rule.destination.ip_range"),
			"global_ip_range": oldidname.ParseNameIDList(ctx, currentRule.Destination.GlobalIPRange,
				"rule.destination.global_ip_range"),
			"remote_asn": convert.ParseList(ctx, types.StringType, currentRule.Destination.RemoteAsn,
				"rule.destination.remote_asn"),
		},
	)
	diags = append(diags, diagstmp...)
	ruleInput.Destination = curRuleDestinationObj
	// /// /// /// /// end Rule -> Destination // /// /// /// /// /

	// /// /// /// /// start Rule -> Service // /// /// /// /// /
	if len(currentRule.Service.Custom) == 0 && len(currentRule.Service.Standard) == 0 {
		// Preserve prior known non-null service shape for backward compatibility.
		// Some existing states carry an object with null nested fields instead of null.
		if !stateRuleInput.Service.IsUnknown() && !stateRuleInput.Service.IsNull() {
			ruleInput.Service = stateRuleInput.Service
		} else {
			ruleInput.Service = types.ObjectNull(IfwServiceAttrTypes)
		}
	} else {
		customServiceValue := types.ListNull(customservice.CustomServiceObjectType)
		if !stateRuleInput.Service.IsNull() && !stateRuleInput.Service.IsUnknown() {
			if stateServiceCustom, ok := stateRuleInput.Service.Attributes()["custom"]; ok && !stateServiceCustom.IsUnknown() {
				customServiceValue = stateServiceCustom.(types.List)
			}
		}

		curRuleServiceObj, diagstmp := types.ObjectValue(
			IfwServiceAttrTypes,
			map[string]attr.Value{
				"standard": oldidname.ParseNameIDList(ctx, currentRule.Service.Standard,
					"rule.service.standard"),
				"custom": customServiceValue,
			},
		)
		diags = append(diags, diagstmp...)
		curRuleServiceObjAttrs := curRuleServiceObj.Attributes()

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
	}
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
	trackingEventValue, diagstmp := types.ObjectValue(
		policyrule.TrackingEventAttrTypes,
		map[string]attr.Value{
			"enabled": types.BoolValue(currentRule.Tracking.Event.Enabled),
		},
	)
	diags = append(diags, diagstmp...)
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

	// /// /// /// /// start Rule -> Exceptions // /// /// /// /// /
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
			tflog.Warn(ctx, "TFLOG_WARN_IFW.curExceptionSourceObj", map[string]interface{}{
				"curExceptionSourceObj": utils.InterfaceToJSONString(curExceptionSourceObj),
			})

			diags = append(diags, diagstmp...)

			// Rule -> Exceptions -> Destination
			curExceptionDestObj, diagstmp := types.ObjectValue(
				IfwDestAttrTypes,
				map[string]attr.Value{
					"application": oldidname.ParseNameIDList(ctx, ruleException.Destination.Application,
						"rule.exception.destination.application"),
					"custom_app": oldidname.ParseNameIDList(ctx, ruleException.Destination.CustomApp,
						"rule.exception.destination.custom_app"),
					"app_category": oldidname.ParseNameIDList(ctx, ruleException.Destination.AppCategory,
						"rule.exception.destination.app_category"),
					"custom_category": oldidname.ParseNameIDList(
						ctx,
						ruleException.Destination.CustomCategory,
						"rule.exception.destination.custom_category",
					),
					"sanctioned_apps_category": oldidname.ParseNameIDList(
						ctx,
						ruleException.Destination.SanctionedAppsCategory,
						"rule.exception.destination.sanctioned_apps_category",
					),
					"country": oldidname.ParseNameIDList(ctx, ruleException.Destination.Country,
						"rule.exception.destination.country"),
					"domain": convert.ParseList(ctx, types.StringType, ruleException.Destination.Domain,
						"rule.exception.destination.domain"),
					"fqdn": convert.ParseList(ctx, types.StringType, ruleException.Destination.Fqdn,
						"rule.exception.destination.fqdn"),
					"ip": convert.ParseList(ctx, types.StringType, ruleException.Destination.IP,
						"rule.exception.destination.ip"),
					"subnet": convert.ParseList(ctx, types.StringType, ruleException.Destination.Subnet,
						"rule.exception.destination.subnet"),
					"ip_range": convert.ParseFromToList(ctx, ruleException.Destination.IPRange,
						"rule.exception.destination.ip_range"),
					"global_ip_range": oldidname.ParseNameIDList(
						ctx,
						ruleException.Destination.GlobalIPRange,
						"rule.exception.destination.global_ip_range",
					),
					"remote_asn": convert.ParseList(
						ctx,
						types.StringType,
						ruleException.Destination.RemoteAsn,
						"rule.exception.destination.remote_asn",
					),
				},
			)
			diags = append(diags, diagstmp...)

			// /// /// /// /// start Rule -> Service // /// /// /// /// /
			// Initialize Service object with null values (matching WAN pattern)
			curExceptionServiceObj, diagstmp := types.ObjectValue(
				IfwServiceAttrTypes,
				map[string]attr.Value{
					"standard": types.SetNull(idname.RefObjectType),
					"custom":   types.ListNull(customservice.CustomServiceObjectType),
				},
			)
			diags = append(diags, diagstmp...)
			curExceptionServiceObjAttrs := curExceptionServiceObj.Attributes()
			if len(ruleException.Service.Custom) > 0 || len(ruleException.Service.Standard) > 0 {
				// Rule -> Service -> Standard (use regular oldidname.ParseNameIDList like WAN firewall)
				curExceptionServiceObjAttrs["standard"] = oldidname.ParseNameIDList(ctx, ruleException.Service.Standard,
					"rule.exception.service.standard")

				// Rule -> Service -> Custom
				if ruleException.Service.Custom != nil {
					if len(ruleException.Service.Custom) > 0 {
						var curExceptionCustomServices []types.Object
						tflog.Info(ctx, "ruleException.Service.Custom - "+fmt.Sprintf("%v", ruleException.Service.Custom))
						for _, item := range ruleException.Service.Custom {
							// Exceptions use PortRangeCustomService in the API; match WAN behavior
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

				curExceptionServiceObj, diagstmp = types.ObjectValue(curExceptionServiceObj.AttributeTypes(ctx), curExceptionServiceObjAttrs)
				diags = append(diags, diagstmp...)
			}
			// /// /// /// /// end Rule -> Service // /// /// /// /// /

			// Internet Firewall exceptions do not have DeviceAttributes in the API response.
			// Preserve the configured value from state when the matching exception still exists.
			exceptionDeviceAttributesObj := useStateExceptionDeviceAttributes(ctx, state, ruleException.Name)

			// Initialize Exception object with populated values (use regular oldidname.ParseNameIDList like WAN firewall)
			exceptionAttrs := map[string]attr.Value{
				"name":   types.StringValue(ruleException.Name),
				"source": curExceptionSourceObj,
				"country": oldidname.ParseNameIDList(ctx, ruleException.Country,
					"rule.exception.country"),
				"device": oldidname.ParseNameIDList(ctx, ruleException.Device,
					"rule.exception.device"),
				"device_attributes": exceptionDeviceAttributesObj,
				"device_os": convert.ParseList(ctx, types.StringType, ruleException.DeviceOs,
					"rule.exception.device_os"),
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
	// /// /// /// /// end Rule -> Exceptions // /// /// /// /// /

	// /// /// /// /// start Rule -> ActivePeriod // /// /// /// /// /
	// Debug logging to see what API returns
	tflog.Warn(ctx, "ActivePeriod from API", map[string]interface{}{
		"EffectiveFrom":    currentRule.ActivePeriod.EffectiveFrom,
		"ExpiresAt":        currentRule.ActivePeriod.ExpiresAt,
		"UseEffectiveFrom": currentRule.ActivePeriod.UseEffectiveFrom,
		"UseExpiresAt":     currentRule.ActivePeriod.UseExpiresAt,
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

	configuredActivePeriod := PolicyPolicyInternetFirewallPolicyRulesRuleActivePeriod{}
	if !stateRuleInput.ActivePeriod.IsNull() && !stateRuleInput.ActivePeriod.IsUnknown() {
		diagstmp = stateRuleInput.ActivePeriod.As(ctx, &configuredActivePeriod, basetypes.ObjectAsOptions{})
		diags = append(diags, diagstmp...)
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
	if !effectiveFromValue.IsNull() && useEffectiveFromValue.ValueBool() {
		parsedEffectiveFromStr, err := utils.ParseTimeString(effectiveFromValue.ValueString(), configuredActivePeriod.EffectiveFrom.ValueString())
		if err == nil {
			tflog.Warn(ctx, "TFLOG_WARN_IFW.ruleInput.parsedEffectiveFromStr", map[string]interface{}{
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
	useEffectiveFromValue = types.BoolValue(
		!effectiveFromValue.IsNull() && !effectiveFromValue.IsUnknown() && effectiveFromValue.ValueString() != "",
	)
	useExpiresAtValue = types.BoolValue(!expiresAtValue.IsNull() && !expiresAtValue.IsUnknown() && expiresAtValue.ValueString() != "")

	tflog.Warn(ctx, "TFLOG_WARN_IFW.ruleInput.ActivePeriod", map[string]interface{}{
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
	if diags.HasError() {
		return ruleInput
	}
	ruleInput.ActivePeriod = curRuleActivePeriodObj
	tflog.Warn(ctx, "Final ActivePeriod IFW object: "+curRuleActivePeriodObj.String())
	// /// /// /// /// end Rule -> ActivePeriod // /// /// /// /// /

	return ruleInput
}

func preserveNullIDNameIDSet(apiSet types.Set, stateSet types.Set) types.Set {
	if apiSet.IsNull() || apiSet.IsUnknown() || stateSet.IsNull() || stateSet.IsUnknown() {
		return apiSet
	}

	stateByName := make(map[string]types.Object)
	for _, stateElem := range stateSet.Elements() {
		stateObj := stateElem.(types.Object)
		name := stateObj.Attributes()["name"].(types.String).ValueString()
		stateByName[name] = stateObj
	}

	apiElems := apiSet.Elements()
	updatedElems := make([]attr.Value, 0, len(apiElems))
	for _, apiElem := range apiElems {
		apiObj := apiElem.(types.Object)
		apiAttrs := apiObj.Attributes()
		apiName := apiAttrs["name"].(types.String).ValueString()

		if stateObj, ok := stateByName[apiName]; ok {
			stateID := stateObj.Attributes()["id"].(types.String)
			// Preserve null for name-only references, but never replace a known
			// API ID with a plan-time unknown value.
			if stateID.IsNull() {
				apiAttrs["id"] = stateID
				apiObj = types.ObjectValueMust(idname.RefAttrTypes, apiAttrs)
			}
		}

		updatedElems = append(updatedElems, apiObj)
	}

	return types.SetValueMust(idname.RefObjectType, updatedElems)
}

func useStateExceptionDeviceAttributes(ctx context.Context, state InternetFirewallRule, exceptionName string) types.Object {
	if state.Rule.IsNull() || state.Rule.IsUnknown() {
		return types.ObjectNull(IfwDeviceAttrAttrTypes)
	}
	var ruleState PolicyPolicyInternetFirewallPolicyRulesRule
	diags := make(diag.Diagnostics, 0)
	diagstmp := state.Rule.As(ctx, &ruleState, basetypes.ObjectAsOptions{})
	diags = append(diags, diagstmp...)
	if diags.HasError() {
		return types.ObjectNull(IfwDeviceAttrAttrTypes)
	}

	if ruleState.Exceptions.IsNull() || ruleState.Exceptions.IsUnknown() {
		return types.ObjectNull(IfwDeviceAttrAttrTypes)
	}

	var exceptions []PolicyPolicyInternetFirewallPolicyRulesRuleExceptions
	diagstmp = ruleState.Exceptions.ElementsAs(ctx, &exceptions, false)
	diags = append(diags, diagstmp...)
	if diags.HasError() {
		return types.ObjectNull(IfwDeviceAttrAttrTypes)
	}

	for _, exception := range exceptions {
		if exception.Name.ValueString() == exceptionName {
			return exception.DeviceAttributes
		}
	}

	return types.ObjectNull(IfwDeviceAttrAttrTypes)
}

func emptyIfwSourceObject() types.Object {
	return types.ObjectValueMust(IfwSourceAttrTypes, map[string]attr.Value{
		"ip":                  types.ListNull(types.StringType),
		"host":                types.SetNull(idname.RefObjectType),
		"site":                types.SetNull(idname.RefObjectType),
		"subnet":              types.ListNull(types.StringType),
		"ip_range":            types.ListNull(iprange.RangeObjectType),
		"global_ip_range":     types.SetNull(idname.RefObjectType),
		"network_interface":   types.SetNull(idname.RefObjectType),
		"site_network_subnet": types.SetNull(idname.RefObjectType),
		"floating_subnet":     types.SetNull(idname.RefObjectType),
		"user":                types.SetNull(idname.RefObjectType),
		"users_group":         types.SetNull(idname.RefObjectType),
		"group":               types.SetNull(idname.RefObjectType),
		"system_group":        types.SetNull(idname.RefObjectType),
	})
}

func emptyIfwDestinationObject() types.Object {
	return types.ObjectValueMust(IfwDestAttrTypes, map[string]attr.Value{
		"application":              types.SetNull(idname.RefObjectType),
		"custom_app":               types.SetNull(idname.RefObjectType),
		"app_category":             types.SetNull(idname.RefObjectType),
		"custom_category":          types.SetNull(idname.RefObjectType),
		"sanctioned_apps_category": types.SetNull(idname.RefObjectType),
		"country":                  types.SetNull(idname.RefObjectType),
		"domain":                   types.ListNull(types.StringType),
		"fqdn":                     types.ListNull(types.StringType),
		"ip":                       types.ListNull(types.StringType),
		"subnet":                   types.ListNull(types.StringType),
		"ip_range":                 types.ListNull(iprange.RangeObjectType),
		"global_ip_range":          types.SetNull(idname.RefObjectType),
		"remote_asn":               types.ListNull(types.StringType),
	})
}

func emptyIfwServiceObject() types.Object {
	return types.ObjectValueMust(IfwServiceAttrTypes, map[string]attr.Value{
		"standard": types.SetNull(idname.RefObjectType),
		"custom":   types.ListNull(customservice.CustomServiceObjectType),
	})
}

func minimalTrackingObject() types.Object {
	return types.ObjectValueMust(policyrule.TrackingAttrTypes, map[string]attr.Value{
		"event": types.ObjectValueMust(policyrule.TrackingEventAttrTypes, map[string]attr.Value{
			"enabled": types.BoolValue(false),
		}),
		"alert": types.ObjectNull(policyrule.TrackingAlertAttrTypes),
	})
}

func EmptyInternetFirewallPolicyResponse() *cato_go_sdk.Policy {
	return &cato_go_sdk.Policy{
		Policy: cato_go_sdk.Policy_Policy{
			InternetFirewall: &cato_go_sdk.Policy_Policy_InternetFirewall{
				Policy: cato_go_sdk.Policy_Policy_InternetFirewall_Policy{
					Rules: []*cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules{},
				},
			},
		},
	}
}

func SuccessfulUpdateRuleResponse(ruleID string) *cato_go_sdk.PolicyInternetFirewallUpdateRule {
	return &cato_go_sdk.PolicyInternetFirewallUpdateRule{
		Policy: cato_go_sdk.PolicyInternetFirewallUpdateRule_Policy{
			InternetFirewall: &cato_go_sdk.PolicyInternetFirewallUpdateRule_Policy_InternetFirewall{
				UpdateRule: cato_go_sdk.PolicyInternetFirewallUpdateRule_Policy_InternetFirewall_UpdateRule{
					Status: "SUCCESS",
					Rule: &cato_go_sdk.PolicyInternetFirewallUpdateRule_Policy_InternetFirewall_UpdateRule_Rule{
						Rule: cato_go_sdk.PolicyInternetFirewallUpdateRule_Policy_InternetFirewall_UpdateRule_Rule_Rule{
							ID: ruleID,
						},
					},
				},
			},
		},
	}
}

func MinimalAPIRule(name string, index int64) cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_Rule {
	return cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_Rule{
		ID:               "rule-123",
		Name:             name,
		Index:            index,
		Enabled:          true,
		Action:           cato_go_sdk.PolicyInternetFirewallAddRule_Policy_InternetFirewall_AddRule_Rule_Rule{}.Action,
		ConnectionOrigin: cato_go_sdk.PolicyInternetFirewallUpdateRule_Policy_InternetFirewall_UpdateRule_Rule_Rule{}.ConnectionOrigin,
	}
}

func GetInternetFwRuleSchema(ctx context.Context, t *testing.T) schema.Schema {
	t.Helper()

	r := &InternetFwRuleResource{}
	resp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, resp)

	return resp.Schema
}

func SuccessfulAddRuleResponse(ruleID string) *cato_go_sdk.PolicyInternetFirewallAddRule {
	return &cato_go_sdk.PolicyInternetFirewallAddRule{
		Policy: cato_go_sdk.PolicyInternetFirewallAddRule_Policy{
			InternetFirewall: &cato_go_sdk.PolicyInternetFirewallAddRule_Policy_InternetFirewall{
				AddRule: cato_go_sdk.PolicyInternetFirewallAddRule_Policy_InternetFirewall_AddRule{
					Status: "SUCCESS",
					Rule: &cato_go_sdk.PolicyInternetFirewallAddRule_Policy_InternetFirewall_AddRule_Rule{
						Rule: cato_go_sdk.PolicyInternetFirewallAddRule_Policy_InternetFirewall_AddRule_Rule_Rule{
							ID: ruleID,
						},
					},
				},
			},
		},
	}
}

func InternetFirewallPolicyResponseWithRule(
	rule cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_Rule,
) *cato_go_sdk.Policy {
	return &cato_go_sdk.Policy{
		Policy: cato_go_sdk.Policy_Policy{
			InternetFirewall: &cato_go_sdk.Policy_Policy_InternetFirewall{
				Policy: cato_go_sdk.Policy_Policy_InternetFirewall_Policy{
					Rules: []*cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules{
						{
							Rule: rule,
						},
					},
				},
			},
		},
	}
}

// convertCreateInput3 converts between the Terraform model and SDK values.
func convertCreateInput3(
	ctx context.Context,
	plan InternetFirewallRule,
	priorDiags *diag.Diagnostics,
	currentRule *cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_Rule,
) (resultRuleObject basetypes.ObjectValue) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	ruleInputRead := HydrateIfwRuleState(ctx, plan, currentRule)
	ruleInputRead.ID = types.StringValue(currentRule.ID)

	resultRuleObject, diags = types.ObjectValueFrom(ctx, InternetFirewallRuleRuleAttrTypes, ruleInputRead)
	return resultRuleObject
}

// convertReadState2 converts between the Terraform model and SDK values.
func convertReadState2(
	ctx context.Context,
	state InternetFirewallRule,
	priorDiags *diag.Diagnostics,
) (resultRule PolicyPolicyInternetFirewallPolicyRulesRule) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	resultRule = PolicyPolicyInternetFirewallPolicyRulesRule{}
	diags = state.Rule.As(ctx, &resultRule, basetypes.ObjectAsOptions{})
	return resultRule
}

// convertReadState3 converts between the Terraform model and SDK values.
func convertReadState3(
	ctx context.Context,
	state InternetFirewallRule,
	priorDiags *diag.Diagnostics,
	diagnostics *diag.Diagnostics,
) (resultCurAtObj basetypes.ObjectValue, resultDiagstmp diag.Diagnostics) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	positionValue := string(cato_models.PolicySectionPositionEnumLastInPolicy)
	refValue := types.StringNull()

	if !state.At.IsNull() && !state.At.IsUnknown() {
		statePosInput := policyrule.Position{}
		diags = state.At.As(ctx, &statePosInput, basetypes.ObjectAsOptions{})
		diagnostics.Append(diags...)
		if !diags.HasError() && !statePosInput.Position.IsNull() && !statePosInput.Position.IsUnknown() {
			positionValue = statePosInput.Position.ValueString()
			if !statePosInput.Ref.IsNull() && !statePosInput.Ref.IsUnknown() {
				refValue = statePosInput.Ref
			}
		}
	}

	resultCurAtObj, resultDiagstmp = types.ObjectValue(
		policyrule.PositionAttrTypes,
		map[string]attr.Value{
			"position": types.StringValue(positionValue),
			"ref":      refValue,
		},
	)
	return resultCurAtObj, resultDiagstmp
}

// convertUpdateInput converts between the Terraform model and SDK values.
func convertUpdateInput(
	ctx context.Context,
	plan InternetFirewallRule,
	priorDiags *diag.Diagnostics,
	diagnostics *diag.Diagnostics,
) (
	resultInputMoveRule cato_models.PolicyMoveRuleInput,
	resultRuleInput PolicyPolicyInternetFirewallPolicyRulesRule,
) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	resultInputMoveRule = cato_models.PolicyMoveRuleInput{}

	// Setting at (to move rule)
	if !plan.At.IsNull() {
		resultInputMoveRule.To = &cato_models.PolicyRulePositionInput{}
		positionInput := policyrule.Position{}
		diags = plan.At.As(ctx, &positionInput, basetypes.ObjectAsOptions{})
		diagnostics.Append(diags...)

		resultInputMoveRule.To.Position = (*cato_models.PolicyRulePositionEnum)(positionInput.Position.ValueStringPointer())
		resultInputMoveRule.To.Ref = positionInput.Ref.ValueStringPointer()
	}

	// // setting rule
	resultRuleInput = PolicyPolicyInternetFirewallPolicyRulesRule{}
	diags = plan.Rule.As(ctx, &resultRuleInput, basetypes.ObjectAsOptions{})
	return resultInputMoveRule, resultRuleInput
}

// convertUpdateInput4 converts between the Terraform model and SDK values.
func convertUpdateInput4(
	ctx context.Context,
	plan InternetFirewallRule,
	priorDiags *diag.Diagnostics,
	currentRule *cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_Rule,
) (resultRuleObject basetypes.ObjectValue) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	ruleInputRead := HydrateIfwRuleState(ctx, plan, currentRule)
	ruleInputRead.ID = types.StringValue(currentRule.ID)

	resultRuleObject, diags = types.ObjectValueFrom(ctx, InternetFirewallRuleRuleAttrTypes, ruleInputRead)
	return resultRuleObject
}

// convertDeleteInput converts between the Terraform model and SDK values.
func convertDeleteInput(
	ctx context.Context,
	state InternetFirewallRule,
	priorDiags *diag.Diagnostics,
) (resultRule PolicyPolicyInternetFirewallPolicyRulesRule) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	resultRule = PolicyPolicyInternetFirewallPolicyRulesRule{}
	diags = state.Rule.As(ctx, &resultRule, basetypes.ObjectAsOptions{})
	return resultRule
}

// convertDeleteInput2 converts between the Terraform model and SDK values.
func convertDeleteInput2(
	rule PolicyPolicyInternetFirewallPolicyRulesRule,
) (
	resultRemoveMutations *cato_models.InternetFirewallPolicyMutationInput,
	resultRemoveRule cato_models.InternetFirewallRemoveRuleInput,
) {
	resultRemoveMutations = &cato_models.InternetFirewallPolicyMutationInput{}
	resultRemoveRule = cato_models.InternetFirewallRemoveRuleInput{
		ID: rule.ID.ValueString(),
	}
	return resultRemoveMutations, resultRemoveRule
}

// convertCreateValues converts between the Terraform model and SDK values.
func convertCreateValues(
	input *HydrateIfwAPITypes,
	subID string,
	anchorBody *cato_go_sdk.Policy,
	diagnostics *diag.Diagnostics,
) {
	cleanupID := ifwSubPolicyCleanupRuleID(anchorBody, subID)
	if cleanupID == "" {
		diagnostics.AddError(
			"Sub-policy not found",
			fmt.Sprintf("could not locate cleanup rule for sub_policy_id %q; verify the sub-policy exists", subID),
		)
		return
	}
	beforeRule := cato_models.PolicyRulePositionEnumBeforeRule
	input.Create.At = &cato_models.PolicyRulePositionInput{
		Position: &beforeRule,
		Ref:      &cleanupID,
	}
}

// convertReadValues converts between the Terraform model and SDK values.
func convertReadValues(
	priorRuleExist *bool,
	priorCurrentRule **cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_Rule,
	priorCurrentSubPolicyID *basetypes.StringValue,
	ruleListItem *cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules,
) {
	ruleExist := *priorRuleExist
	defer func() { *priorRuleExist = ruleExist }()
	currentRule := *priorCurrentRule
	defer func() { *priorCurrentRule = currentRule }()
	currentSubPolicyID := *priorCurrentSubPolicyID
	defer func() { *priorCurrentSubPolicyID = currentSubPolicyID }()
	ruleExist = true
	currentRule = ruleListItem.GetRule()
	if sp := ruleListItem.GetSubPolicy(); sp != nil && sp.GetID() != "" {
		currentSubPolicyID = types.StringValue(sp.GetID())
	}
}
