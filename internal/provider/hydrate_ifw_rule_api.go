package provider

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	cato_scalars "github.com/catonetworks/cato-go-sdk/scalars"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	// Import the correct package
)

// hydrateIfwApiTypes create sub-types for both create and update calls to populate both entries
type hydrateIfwApiTypes struct {
	create cato_models.InternetFirewallAddRuleInput
	update cato_models.InternetFirewallUpdateRuleInput
}

// hydrateIfwApiRuleState takes in the current state/plan along with context and returns the created
// diagnostic data as well as cato api data used to either create or update IFW entries
func hydrateIfwRuleApi(ctx context.Context, plan InternetFirewallRule) (hydrateIfwApiTypes, diag.Diagnostics) {
	diags := []diag.Diagnostic{}

	hydrateApiReturn := hydrateIfwApiTypes{}
	hydrateApiReturn.create = cato_models.InternetFirewallAddRuleInput{}
	hydrateApiReturn.update = cato_models.InternetFirewallUpdateRuleInput{}
	hydrateApiReturn.create.At = &cato_models.PolicyRulePositionInput{}

	rootAddRule := &cato_models.InternetFirewallAddRuleDataInput{}
	rootUpdateRule := &cato_models.InternetFirewallUpdateRuleDataInput{}

	//setting at for creation only
	if !plan.At.IsNull() {

		positionInput := PolicyRulePositionInput{}
		diags = append(diags, plan.At.As(ctx, &positionInput, basetypes.ObjectAsOptions{})...)

		hydrateApiReturn.create.At.Position = (*cato_models.PolicyRulePositionEnum)(positionInput.Position.ValueStringPointer())
		hydrateApiReturn.create.At.Ref = positionInput.Ref.ValueStringPointer()

	}

	// setting rule
	if !plan.Rule.IsNull() {

		ruleInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule{}
		diags = append(diags, plan.Rule.As(ctx, &ruleInput, basetypes.ObjectAsOptions{})...)

		// setting source
		if !ruleInput.Source.IsNull() {

			ruleSourceInput := &cato_models.InternetFirewallSourceInput{
				IP:                make([]string, 0),
				Subnet:            make([]string, 0),
				Host:              make([]*cato_models.HostRefInput, 0),
				Site:              make([]*cato_models.SiteRefInput, 0),
				IPRange:           make([]*cato_models.IPAddressRangeInput, 0),
				GlobalIPRange:     make([]*cato_models.GlobalIPRangeRefInput, 0),
				NetworkInterface:  make([]*cato_models.NetworkInterfaceRefInput, 0),
				SiteNetworkSubnet: make([]*cato_models.SiteNetworkSubnetRefInput, 0),
				FloatingSubnet:    make([]*cato_models.FloatingSubnetRefInput, 0),
				User:              make([]*cato_models.UserRefInput, 0),
				UsersGroup:        make([]*cato_models.UsersGroupRefInput, 0),
				Group:             make([]*cato_models.GroupRefInput, 0),
				SystemGroup:       make([]*cato_models.SystemGroupRefInput, 0),
			}
			ruleSourceUpdateInput := &cato_models.InternetFirewallSourceUpdateInput{}

			sourceInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source{}
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

				var itemSourceHostInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_Host
				for _, item := range elementsSourceHostInput {
					diags = append(diags, item.As(ctx, &itemSourceHostInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceHostInput)
					if err != nil {
						tflog.Error(ctx, err.Error())
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

				var itemSourceSiteInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_Site
				for _, item := range elementsSourceSiteInput {
					diags = append(diags, item.As(ctx, &itemSourceSiteInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceSiteInput)
					if err != nil {
						tflog.Error(ctx, err.Error())
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

				var itemSourceIPRangeInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_IPRange
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

				var itemSourceGlobalIPRangeInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_GlobalIPRange
				for _, item := range elementsSourceGlobalIPRangeInput {
					diags = append(diags, item.As(ctx, &itemSourceGlobalIPRangeInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceGlobalIPRangeInput)
					if err != nil {
						tflog.Error(ctx, err.Error())
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

				var itemSourceNetworkInterfaceInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_NetworkInterface
				for _, item := range elementsSourceNetworkInterfaceInput {
					diags = append(diags, item.As(ctx, &itemSourceNetworkInterfaceInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceNetworkInterfaceInput)
					if err != nil {
						tflog.Error(ctx, err.Error())
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

				var itemSourceSiteNetworkSubnetInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_SiteNetworkSubnet
				for _, item := range elementsSourceSiteNetworkSubnetInput {
					diags = append(diags, item.As(ctx, &itemSourceSiteNetworkSubnetInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceSiteNetworkSubnetInput)
					if err != nil {
						tflog.Error(ctx, err.Error())
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

				var itemSourceFloatingSubnetInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_FloatingSubnet
				for _, item := range elementsSourceFloatingSubnetInput {
					diags = append(diags, item.As(ctx, &itemSourceFloatingSubnetInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceFloatingSubnetInput)
					if err != nil {
						tflog.Error(ctx, err.Error())
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

				var itemSourceUserInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_User
				for _, item := range elementsSourceUserInput {
					diags = append(diags, item.As(ctx, &itemSourceUserInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceUserInput)
					if err != nil {
						tflog.Error(ctx, err.Error())
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

				var itemSourceUsersGroupInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_UsersGroup
				for _, item := range elementsSourceUsersGroupInput {
					diags = append(diags, item.As(ctx, &itemSourceUsersGroupInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceUsersGroupInput)
					if err != nil {
						tflog.Error(ctx, err.Error())
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

				var itemSourceGroupInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_Group
				for _, item := range elementsSourceGroupInput {
					diags = append(diags, item.As(ctx, &itemSourceGroupInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceGroupInput)
					if err != nil {
						tflog.Error(ctx, err.Error())
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

				var itemSourceSystemGroupInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_SystemGroup
				for _, item := range elementsSourceSystemGroupInput {
					diags = append(diags, item.As(ctx, &itemSourceSystemGroupInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceSystemGroupInput)
					if err != nil {
						tflog.Error(ctx, err.Error())
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

			var itemCountryInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Country
			for _, item := range elementsCountryInput {
				diags = append(diags, item.As(ctx, &itemCountryInput, basetypes.ObjectAsOptions{})...)

				ObjectRefOutput, err := utils.TransformObjectRefInput(itemCountryInput)
				if err != nil {
					tflog.Error(ctx, err.Error())
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

			var itemDeviceInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Device
			for _, item := range elementsDeviceInput {
				diags = append(diags, item.As(ctx, &itemDeviceInput, basetypes.ObjectAsOptions{})...)

				ObjectRefOutput, err := utils.TransformObjectRefInput(itemDeviceInput)
				if err != nil {
					tflog.Error(ctx, err.Error())
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

			ruleDestinationInput := &cato_models.InternetFirewallDestinationInput{
				IP:                     make([]string, 0),
				Subnet:                 make([]string, 0),
				Domain:                 make([]string, 0),
				Fqdn:                   make([]string, 0),
				RemoteAsn:              make([]cato_scalars.Asn16, 0),
				Application:            make([]*cato_models.ApplicationRefInput, 0),
				CustomApp:              make([]*cato_models.CustomApplicationRefInput, 0),
				IPRange:                make([]*cato_models.IPAddressRangeInput, 0),
				GlobalIPRange:          make([]*cato_models.GlobalIPRangeRefInput, 0),
				AppCategory:            make([]*cato_models.ApplicationCategoryRefInput, 0),
				CustomCategory:         make([]*cato_models.CustomCategoryRefInput, 0),
				SanctionedAppsCategory: make([]*cato_models.SanctionedAppsCategoryRefInput, 0),
				Country:                make([]*cato_models.CountryRefInput, 0),
			}
			ruleDestinationUpdateInput := &cato_models.InternetFirewallDestinationUpdateInput{}

			destinationInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination{}
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
				ruleDestinationUpdateInput.RemoteAsn = make([]cato_scalars.Asn16, 0)
			}

			// setting destination application
			if !destinationInput.Application.IsUnknown() && !destinationInput.Application.IsNull() {
				elementsDestinationApplicationInput := make([]types.Object, 0, len(destinationInput.Application.Elements()))
				diags = append(diags, destinationInput.Application.ElementsAs(ctx, &elementsDestinationApplicationInput, false)...)

				var itemDestinationApplicationInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_Application
				for _, item := range elementsDestinationApplicationInput {
					diags = append(diags, item.As(ctx, &itemDestinationApplicationInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationApplicationInput)
					if err != nil {
						tflog.Error(ctx, err.Error())
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

				var itemDestinationCustomAppInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_CustomApp
				for _, item := range elementsDestinationCustomAppInput {
					diags = append(diags, item.As(ctx, &itemDestinationCustomAppInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationCustomAppInput)
					if err != nil {
						tflog.Error(ctx, err.Error())
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

				var itemDestinationIPRangeInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_IPRange
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

				var itemDestinationGlobalIPRangeInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_GlobalIPRange
				for _, item := range elementsDestinationGlobalIPRangeInput {
					diags = append(diags, item.As(ctx, &itemDestinationGlobalIPRangeInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationGlobalIPRangeInput)
					if err != nil {
						tflog.Error(ctx, err.Error())
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

				var itemDestinationAppCategoryInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_AppCategory
				for _, item := range elementsDestinationAppCategoryInput {
					diags = append(diags, item.As(ctx, &itemDestinationAppCategoryInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationAppCategoryInput)
					if err != nil {
						tflog.Error(ctx, err.Error())
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

				var itemDestinationCustomCategoryInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_CustomCategory
				for _, item := range elementsDestinationCustomCategoryInput {
					diags = append(diags, item.As(ctx, &itemDestinationCustomCategoryInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationCustomCategoryInput)
					if err != nil {
						tflog.Error(ctx, err.Error())
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

			// setting destination sanctionned apps category
			if !destinationInput.SanctionedAppsCategory.IsNull() {
				elementsDestinationSanctionedAppsCategoryInput := make([]types.Object, 0, len(destinationInput.SanctionedAppsCategory.Elements()))
				diags = append(diags, destinationInput.SanctionedAppsCategory.ElementsAs(ctx, &elementsDestinationSanctionedAppsCategoryInput, false)...)

				var itemDestinationSanctionedAppsCategoryInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_SanctionedAppsCategory
				for _, item := range elementsDestinationSanctionedAppsCategoryInput {
					diags = append(diags, item.As(ctx, &itemDestinationSanctionedAppsCategoryInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationSanctionedAppsCategoryInput)
					if err != nil {
						tflog.Error(ctx, err.Error())
					}

					ruleDestinationInput.SanctionedAppsCategory = append(ruleDestinationInput.SanctionedAppsCategory, &cato_models.SanctionedAppsCategoryRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
				ruleDestinationUpdateInput.SanctionedAppsCategory = ruleDestinationInput.SanctionedAppsCategory
			} else {
				ruleDestinationUpdateInput.SanctionedAppsCategory = make([]*cato_models.SanctionedAppsCategoryRefInput, 0)
			}

			// setting destination country
			if !destinationInput.Country.IsNull() {
				elementsDestinationCountryInput := make([]types.Object, 0, len(destinationInput.Country.Elements()))
				diags = append(diags, destinationInput.Country.ElementsAs(ctx, &elementsDestinationCountryInput, false)...)

				var itemDestinationCountryInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_Country
				for _, item := range elementsDestinationCountryInput {
					diags = append(diags, item.As(ctx, &itemDestinationCountryInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationCountryInput)
					if err != nil {
						tflog.Error(ctx, err.Error())
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
			ruleServiceInput := &cato_models.InternetFirewallServiceTypeInput{
				Custom:   make([]*cato_models.CustomServiceInput, 0),
				Standard: make([]*cato_models.ServiceRefInput, 0),
			}
			ruleServiceUpdateInput := &cato_models.InternetFirewallServiceTypeUpdateInput{}

			serviceInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service{}
			diags = append(diags, ruleInput.Service.As(ctx, &serviceInput, basetypes.ObjectAsOptions{})...)

			// setting service standard
			if !serviceInput.Standard.IsNull() {
				elementsServiceStandardInput := make([]types.Object, 0, len(serviceInput.Standard.Elements()))
				diags = append(diags, serviceInput.Standard.ElementsAs(ctx, &elementsServiceStandardInput, false)...)

				var itemServiceStandardInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service_Standard
				for _, item := range elementsServiceStandardInput {
					diags = append(diags, item.As(ctx, &itemServiceStandardInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemServiceStandardInput)
					if err != nil {
						tflog.Error(ctx, err.Error())
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

				var itemServiceCustomInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service_Custom
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
						var itemPortRange Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service_Custom_PortRange
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

			trackingInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking{}
			diags = append(diags, ruleInput.Tracking.As(ctx, &trackingInput, basetypes.ObjectAsOptions{})...)

			if !trackingInput.Event.IsUnknown() && !trackingInput.Event.IsNull() {
				// setting tracking event
				trackingEventInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Event{}
				diags = append(diags, trackingInput.Event.As(ctx, &trackingEventInput, basetypes.ObjectAsOptions{})...)
				rootAddRule.Tracking.Event.Enabled = trackingEventInput.Enabled.ValueBool()
				rootUpdateRule.Tracking.Event.Enabled = trackingEventInput.Enabled.ValueBoolPointer()
			}

			if !trackingInput.Alert.IsUnknown() && !trackingInput.Alert.IsNull() {

				rootAddRule.Tracking.Alert = &cato_models.PolicyRuleTrackingAlertInput{}

				trackingAlertInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Alert{}
				diags = append(diags, trackingInput.Alert.As(ctx, &trackingAlertInput, basetypes.ObjectAsOptions{})...)

				rootAddRule.Tracking.Alert.Enabled = trackingAlertInput.Enabled.ValueBool()
				rootAddRule.Tracking.Alert.Frequency = (cato_models.PolicyRuleTrackingFrequencyEnum)(trackingAlertInput.Frequency.ValueString())

				rootUpdateRule.Tracking.Alert.Enabled = trackingAlertInput.Enabled.ValueBoolPointer()
				rootUpdateRule.Tracking.Alert.Frequency = (*cato_models.PolicyRuleTrackingFrequencyEnum)(trackingAlertInput.Frequency.ValueStringPointer())

				// setting tracking alert subscription group
				if !trackingAlertInput.SubscriptionGroup.IsNull() {
					elementsAlertSubscriptionGroupInput := make([]types.Object, 0, len(trackingAlertInput.SubscriptionGroup.Elements()))
					diags = append(diags, trackingAlertInput.SubscriptionGroup.ElementsAs(ctx, &elementsAlertSubscriptionGroupInput, false)...)

					var itemAlertSubscriptionGroupInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Alert_SubscriptionGroup
					for _, item := range elementsAlertSubscriptionGroupInput {
						diags = append(diags, item.As(ctx, &itemAlertSubscriptionGroupInput, basetypes.ObjectAsOptions{})...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemAlertSubscriptionGroupInput)
						if err != nil {
							tflog.Error(ctx, err.Error())
						}

						rootAddRule.Tracking.Alert.SubscriptionGroup = append(rootAddRule.Tracking.Alert.SubscriptionGroup, &cato_models.SubscriptionGroupRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
					rootUpdateRule.Tracking.Alert.SubscriptionGroup = rootAddRule.Tracking.Alert.SubscriptionGroup
				} else {
					rootAddRule.Tracking.Alert.SubscriptionGroup = make([]*cato_models.SubscriptionGroupRefInput, 0)
					rootUpdateRule.Tracking.Alert.SubscriptionGroup = make([]*cato_models.SubscriptionGroupRefInput, 0)
				}

				// setting tracking alert webhook
				if !trackingAlertInput.Webhook.IsNull() {
					elementsAlertWebHookInput := make([]types.Object, 0, len(trackingAlertInput.Webhook.Elements()))
					diags = append(diags, trackingAlertInput.Webhook.ElementsAs(ctx, &elementsAlertWebHookInput, false)...)

					var itemAlertWebHookInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Alert_SubscriptionGroup
					for _, item := range elementsAlertWebHookInput {
						diags = append(diags, item.As(ctx, &itemAlertWebHookInput, basetypes.ObjectAsOptions{})...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemAlertWebHookInput)
						if err != nil {
							tflog.Error(ctx, err.Error())
						}

						rootAddRule.Tracking.Alert.Webhook = append(rootAddRule.Tracking.Alert.Webhook, &cato_models.SubscriptionWebhookRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
					rootUpdateRule.Tracking.Alert.Webhook = rootAddRule.Tracking.Alert.Webhook
				} else {
					rootAddRule.Tracking.Alert.Webhook = make([]*cato_models.SubscriptionWebhookRefInput, 0)
					rootUpdateRule.Tracking.Alert.Webhook = make([]*cato_models.SubscriptionWebhookRefInput, 0)
				}

				// setting tracking alert mailing list
				tflog.Warn(ctx, "hydrateIfwApiRuleState() trackingAlertInput.MailingList "+fmt.Sprintf("%v", trackingAlertInput.MailingList))
				if !trackingAlertInput.MailingList.IsNull() {
					elementsAlertMailingListInput := make([]types.Object, 0, len(trackingAlertInput.MailingList.Elements()))
					diags = append(diags, trackingAlertInput.MailingList.ElementsAs(ctx, &elementsAlertMailingListInput, false)...)

					var itemAlertMailingListInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Alert_SubscriptionGroup
					for _, item := range elementsAlertMailingListInput {
						diags = append(diags, item.As(ctx, &itemAlertMailingListInput, basetypes.ObjectAsOptions{})...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemAlertMailingListInput)
						if err != nil {
							tflog.Error(ctx, err.Error())
						}

						rootAddRule.Tracking.Alert.MailingList = append(rootAddRule.Tracking.Alert.MailingList, &cato_models.SubscriptionMailingListRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
					rootUpdateRule.Tracking.Alert.MailingList = rootAddRule.Tracking.Alert.MailingList
				} else {
					rootAddRule.Tracking.Alert.MailingList = make([]*cato_models.SubscriptionMailingListRefInput, 0)
					rootUpdateRule.Tracking.Alert.MailingList = make([]*cato_models.SubscriptionMailingListRefInput, 0)
				}
			}
		}

		// setting schedule
		rootAddRule.Schedule = &cato_models.PolicyScheduleInput{
			ActiveOn: (cato_models.PolicyActiveOnEnum)("ALWAYS"),
		}

		activeOn := "ALWAYS"
		rootUpdateRule.Schedule = &cato_models.PolicyScheduleUpdateInput{
			ActiveOn: (*cato_models.PolicyActiveOnEnum)(&activeOn),
		}

		if !ruleInput.Schedule.IsUnknown() && !ruleInput.Schedule.IsNull() {

			scheduleInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Schedule{}
			diags = append(diags, ruleInput.Schedule.As(ctx, &scheduleInput, basetypes.ObjectAsOptions{})...)

			rootAddRule.Schedule.ActiveOn = cato_models.PolicyActiveOnEnum(scheduleInput.ActiveOn.ValueString())
			rootUpdateRule.Schedule.ActiveOn = (*cato_models.PolicyActiveOnEnum)(scheduleInput.ActiveOn.ValueStringPointer())

			// setting schedule custome time frame
			if !scheduleInput.CustomTimeframe.IsNull() {
				rootAddRule.Schedule.CustomTimeframe = &cato_models.PolicyCustomTimeframeInput{}
				rootUpdateRule.Schedule.CustomTimeframe = &cato_models.PolicyCustomTimeframeUpdateInput{}

				customeTimeFrameInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Schedule_CustomTimeframe{}
				diags = append(diags, scheduleInput.CustomTimeframe.As(ctx, &customeTimeFrameInput, basetypes.ObjectAsOptions{})...)

				rootAddRule.Schedule.CustomTimeframe.From = customeTimeFrameInput.From.ValueString()
				rootAddRule.Schedule.CustomTimeframe.To = customeTimeFrameInput.To.ValueString()

				rootUpdateRule.Schedule.CustomTimeframe.From = customeTimeFrameInput.From.ValueStringPointer()
				rootUpdateRule.Schedule.CustomTimeframe.To = customeTimeFrameInput.To.ValueStringPointer()
			} else {
				rootUpdateRule.Schedule.CustomTimeframe = &cato_models.PolicyCustomTimeframeUpdateInput{}
			}

			// setting schedule custom recurring
			if !scheduleInput.CustomRecurring.IsNull() {
				rootAddRule.Schedule.CustomRecurring = &cato_models.PolicyCustomRecurringInput{}
				rootUpdateRule.Schedule.CustomRecurring = &cato_models.PolicyCustomRecurringUpdateInput{}

				customRecurringInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Schedule_CustomRecurring{}
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
			var itemExceptionsInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Exceptions
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
					exceptionAddInput.ConnectionOrigin = cato_models.ConnectionOriginEnum("ANY")
				}

				exceptionUpdateInput.ConnectionOrigin = exceptionAddInput.ConnectionOrigin

				// setting source
				if !itemExceptionsInput.Source.IsNull() {

					exceptionAddInput.Source = &cato_models.InternetFirewallSourceInput{
						IP:                make([]string, 0),
						Subnet:            make([]string, 0),
						Host:              make([]*cato_models.HostRefInput, 0),
						Site:              make([]*cato_models.SiteRefInput, 0),
						IPRange:           make([]*cato_models.IPAddressRangeInput, 0),
						GlobalIPRange:     make([]*cato_models.GlobalIPRangeRefInput, 0),
						NetworkInterface:  make([]*cato_models.NetworkInterfaceRefInput, 0),
						SiteNetworkSubnet: make([]*cato_models.SiteNetworkSubnetRefInput, 0),
						FloatingSubnet:    make([]*cato_models.FloatingSubnetRefInput, 0),
						User:              make([]*cato_models.UserRefInput, 0),
						UsersGroup:        make([]*cato_models.UsersGroupRefInput, 0),
						Group:             make([]*cato_models.GroupRefInput, 0),
						SystemGroup:       make([]*cato_models.SystemGroupRefInput, 0),
					}
					exceptionUpdateInput.Source = &cato_models.InternetFirewallSourceInput{
						IP:                make([]string, 0),
						Subnet:            make([]string, 0),
						Host:              make([]*cato_models.HostRefInput, 0),
						Site:              make([]*cato_models.SiteRefInput, 0),
						IPRange:           make([]*cato_models.IPAddressRangeInput, 0),
						GlobalIPRange:     make([]*cato_models.GlobalIPRangeRefInput, 0),
						NetworkInterface:  make([]*cato_models.NetworkInterfaceRefInput, 0),
						SiteNetworkSubnet: make([]*cato_models.SiteNetworkSubnetRefInput, 0),
						FloatingSubnet:    make([]*cato_models.FloatingSubnetRefInput, 0),
						User:              make([]*cato_models.UserRefInput, 0),
						UsersGroup:        make([]*cato_models.UsersGroupRefInput, 0),
						Group:             make([]*cato_models.GroupRefInput, 0),
						SystemGroup:       make([]*cato_models.SystemGroupRefInput, 0),
					}

					sourceInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source{}
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

						var itemSourceHostInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_Host
						for _, item := range elementsSourceHostInput {
							diags = append(diags, item.As(ctx, &itemSourceHostInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceHostInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
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

						var itemSourceSiteInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_Site
						for _, item := range elementsSourceSiteInput {
							diags = append(diags, item.As(ctx, &itemSourceSiteInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceSiteInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
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

						var itemSourceIPRangeInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_IPRange
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

						var itemSourceGlobalIPRangeInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_GlobalIPRange
						for _, item := range elementsSourceGlobalIPRangeInput {
							diags = append(diags, item.As(ctx, &itemSourceGlobalIPRangeInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceGlobalIPRangeInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
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

						var itemSourceNetworkInterfaceInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_NetworkInterface
						for _, item := range elementsSourceNetworkInterfaceInput {
							diags = append(diags, item.As(ctx, &itemSourceNetworkInterfaceInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceNetworkInterfaceInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
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

						var itemSourceSiteNetworkSubnetInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_SiteNetworkSubnet
						for _, item := range elementsSourceSiteNetworkSubnetInput {
							diags = append(diags, item.As(ctx, &itemSourceSiteNetworkSubnetInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceSiteNetworkSubnetInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionAddInput.Source.SiteNetworkSubnet = append(exceptionAddInput.Source.SiteNetworkSubnet, &cato_models.SiteNetworkSubnetRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
						exceptionUpdateInput.Source.SiteNetworkSubnet = exceptionAddInput.Source.SiteNetworkSubnet
					} else {
						exceptionUpdateInput.Source.SiteNetworkSubnet = make([]*cato_models.SiteNetworkSubnetRefInput, 0)
					}

					// setting source floating subnet
					if !sourceInput.FloatingSubnet.IsNull() {
						elementsSourceFloatingSubnetInput := make([]types.Object, 0, len(sourceInput.FloatingSubnet.Elements()))
						diags = append(diags, sourceInput.FloatingSubnet.ElementsAs(ctx, &elementsSourceFloatingSubnetInput, false)...)

						var itemSourceFloatingSubnetInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_FloatingSubnet
						for _, item := range elementsSourceFloatingSubnetInput {
							diags = append(diags, item.As(ctx, &itemSourceFloatingSubnetInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceFloatingSubnetInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
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

						var itemSourceUserInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_User
						for _, item := range elementsSourceUserInput {
							diags = append(diags, item.As(ctx, &itemSourceUserInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceUserInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
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

						var itemSourceUsersGroupInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_UsersGroup
						for _, item := range elementsSourceUsersGroupInput {
							diags = append(diags, item.As(ctx, &itemSourceUsersGroupInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceUsersGroupInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
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

						var itemSourceGroupInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_Group
						for _, item := range elementsSourceGroupInput {
							diags = append(diags, item.As(ctx, &itemSourceGroupInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceGroupInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
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

						var itemSourceSystemGroupInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_SystemGroup
						for _, item := range elementsSourceSystemGroupInput {
							diags = append(diags, item.As(ctx, &itemSourceSystemGroupInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceSystemGroupInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
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

					var itemCountryInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Country
					for _, item := range elementsCountryInput {
						diags = append(diags, item.As(ctx, &itemCountryInput, basetypes.ObjectAsOptions{})...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemCountryInput)
						if err != nil {
							tflog.Error(ctx, err.Error())
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

					var itemDeviceInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Device
					for _, item := range elementsDeviceInput {
						diags = append(diags, item.As(ctx, &itemDeviceInput, basetypes.ObjectAsOptions{})...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemDeviceInput)
						if err != nil {
							tflog.Error(ctx, err.Error())
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

					exceptionAddInput.DeviceAttributes = &cato_models.DeviceAttributesInput{
						Category:     make([]string, 0),
						Manufacturer: make([]string, 0),
						Model:        make([]string, 0),
						Os:           make([]string, 0),
						OsVersion:    make([]string, 0),
						Type:         make([]string, 0),
					}
					exceptionUpdateInput.DeviceAttributes = &cato_models.DeviceAttributesInput{
						Category:     make([]string, 0),
						Manufacturer: make([]string, 0),
						Model:        make([]string, 0),
						Os:           make([]string, 0),
						OsVersion:    make([]string, 0),
						Type:         make([]string, 0),
					}

					deviceAttributesInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_DeviceAttributes{}
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

					exceptionAddInput.Destination = &cato_models.InternetFirewallDestinationInput{
						IP:               make([]string, 0),
						Subnet:           make([]string, 0),
						Domain:           make([]string, 0),
						Fqdn:             make([]string, 0),
						RemoteAsn:        make([]cato_scalars.Asn16, 0),
						Application:      make([]*cato_models.ApplicationRefInput, 0),
						CustomApp:        make([]*cato_models.CustomApplicationRefInput, 0),
						IPRange:          make([]*cato_models.IPAddressRangeInput, 0),
						GlobalIPRange:    make([]*cato_models.GlobalIPRangeRefInput, 0),
						AppCategory:      make([]*cato_models.ApplicationCategoryRefInput, 0),
						CustomCategory:   make([]*cato_models.CustomCategoryRefInput, 0),
						SanctionedAppsCategory: make([]*cato_models.SanctionedAppsCategoryRefInput, 0),
						Country:          make([]*cato_models.CountryRefInput, 0),
					}
					exceptionUpdateInput.Destination = &cato_models.InternetFirewallDestinationInput{
						IP:               make([]string, 0),
						Subnet:           make([]string, 0),
						Domain:           make([]string, 0),
						Fqdn:             make([]string, 0),
						RemoteAsn:        make([]cato_scalars.Asn16, 0),
						Application:      make([]*cato_models.ApplicationRefInput, 0),
						CustomApp:        make([]*cato_models.CustomApplicationRefInput, 0),
						IPRange:          make([]*cato_models.IPAddressRangeInput, 0),
						GlobalIPRange:    make([]*cato_models.GlobalIPRangeRefInput, 0),
						AppCategory:      make([]*cato_models.ApplicationCategoryRefInput, 0),
						CustomCategory:   make([]*cato_models.CustomCategoryRefInput, 0),
						SanctionedAppsCategory: make([]*cato_models.SanctionedAppsCategoryRefInput, 0),
						Country:          make([]*cato_models.CountryRefInput, 0),
					}

					destinationInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination{}
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
						exceptionUpdateInput.Destination.RemoteAsn = make([]cato_scalars.Asn16, 0)
					}

					// setting destination application
					if !destinationInput.Application.IsNull() {
						elementsDestinationApplicationInput := make([]types.Object, 0, len(destinationInput.Application.Elements()))
						diags = append(diags, destinationInput.Application.ElementsAs(ctx, &elementsDestinationApplicationInput, false)...)

						var itemDestinationApplicationInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_Application
						for _, item := range elementsDestinationApplicationInput {
							diags = append(diags, item.As(ctx, &itemDestinationApplicationInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationApplicationInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
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

						var itemDestinationCustomAppInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_CustomApp
						for _, item := range elementsDestinationCustomAppInput {
							diags = append(diags, item.As(ctx, &itemDestinationCustomAppInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationCustomAppInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
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

						var itemDestinationIPRangeInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_IPRange
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

						var itemDestinationGlobalIPRangeInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_GlobalIPRange
						for _, item := range elementsDestinationGlobalIPRangeInput {
							diags = append(diags, item.As(ctx, &itemDestinationGlobalIPRangeInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationGlobalIPRangeInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
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

						var itemDestinationAppCategoryInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_AppCategory
						for _, item := range elementsDestinationAppCategoryInput {
							diags = append(diags, item.As(ctx, &itemDestinationAppCategoryInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationAppCategoryInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionAddInput.Destination.AppCategory = append(exceptionAddInput.Destination.AppCategory, &cato_models.ApplicationCategoryRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
						exceptionUpdateInput.Destination.AppCategory = exceptionAddInput.Destination.AppCategory
					} else {
						exceptionUpdateInput.Destination.AppCategory = make([]*cato_models.ApplicationCategoryRefInput, 0)
					}

					// setting destination custom app category
					if !destinationInput.CustomCategory.IsNull() {
						elementsDestinationCustomCategoryInput := make([]types.Object, 0, len(destinationInput.CustomCategory.Elements()))
						diags = append(diags, destinationInput.CustomCategory.ElementsAs(ctx, &elementsDestinationCustomCategoryInput, false)...)

						var itemDestinationCustomCategoryInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_CustomCategory
						for _, item := range elementsDestinationCustomCategoryInput {
							diags = append(diags, item.As(ctx, &itemDestinationCustomCategoryInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationCustomCategoryInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionAddInput.Destination.CustomCategory = append(exceptionAddInput.Destination.CustomCategory, &cato_models.CustomCategoryRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
						exceptionUpdateInput.Destination.CustomCategory = exceptionAddInput.Destination.CustomCategory
					} else {
						exceptionUpdateInput.Destination.CustomCategory = make([]*cato_models.CustomCategoryRefInput, 0)
					}

					// setting destination sanctionned apps category
					if !destinationInput.SanctionedAppsCategory.IsNull() {
						elementsDestinationSanctionedAppsCategoryInput := make([]types.Object, 0, len(destinationInput.SanctionedAppsCategory.Elements()))
						diags = append(diags, destinationInput.SanctionedAppsCategory.ElementsAs(ctx, &elementsDestinationSanctionedAppsCategoryInput, false)...)

						var itemDestinationSanctionedAppsCategoryInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_SanctionedAppsCategory
						for _, item := range elementsDestinationSanctionedAppsCategoryInput {
							diags = append(diags, item.As(ctx, &itemDestinationSanctionedAppsCategoryInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationSanctionedAppsCategoryInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionAddInput.Destination.SanctionedAppsCategory = append(exceptionAddInput.Destination.SanctionedAppsCategory, &cato_models.SanctionedAppsCategoryRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
						exceptionUpdateInput.Destination.SanctionedAppsCategory = exceptionAddInput.Destination.SanctionedAppsCategory
					} else {
						exceptionUpdateInput.Destination.SanctionedAppsCategory = make([]*cato_models.SanctionedAppsCategoryRefInput, 0)
					}

					// setting destination country
					if !destinationInput.Country.IsNull() {
						elementsDestinationCountryInput := make([]types.Object, 0, len(destinationInput.Country.Elements()))
						diags = append(diags, destinationInput.Country.ElementsAs(ctx, &elementsDestinationCountryInput, false)...)

						var itemDestinationCountryInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_Country
						for _, item := range elementsDestinationCountryInput {
							diags = append(diags, item.As(ctx, &itemDestinationCountryInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationCountryInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
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
					exceptionUpdateInput.Destination = &cato_models.InternetFirewallDestinationInput{}
				}

				// setting service
				if !itemExceptionsInput.Service.IsNull() {

					exceptionAddInput.Service = &cato_models.InternetFirewallServiceTypeInput{}
					exceptionUpdateInput.Service = &cato_models.InternetFirewallServiceTypeInput{}

					serviceInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service{}
					diags = append(diags, itemExceptionsInput.Service.As(ctx, &serviceInput, basetypes.ObjectAsOptions{})...)

					// setting service standard
					if !serviceInput.Standard.IsNull() {
						elementsServiceStandardInput := make([]types.Object, 0, len(serviceInput.Standard.Elements()))
						diags = append(diags, serviceInput.Standard.ElementsAs(ctx, &elementsServiceStandardInput, false)...)

						var itemServiceStandardInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service_Standard
						for _, item := range elementsServiceStandardInput {
							diags = append(diags, item.As(ctx, &itemServiceStandardInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemServiceStandardInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
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

						var itemServiceCustomInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service_Custom
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
								var itemPortRange Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service_Custom_PortRange
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
					exceptionUpdateInput.Service = &cato_models.InternetFirewallServiceTypeInput{}
				}

				rootAddRule.Exceptions = append(rootAddRule.Exceptions, &exceptionAddInput)
				rootUpdateRule.Exceptions = append(rootUpdateRule.Exceptions, &exceptionUpdateInput)
			}
		}

		// setting activePeriod
		if !ruleInput.ActivePeriod.IsNull() && !ruleInput.ActivePeriod.IsUnknown() {
			activePeriodInput := Policy_Policy_WanFirewall_Policy_Rules_Rule_ActivePeriod{}
			diags = append(diags, ruleInput.ActivePeriod.As(ctx, &activePeriodInput, basetypes.ObjectAsOptions{})...)

			activePeriodApiInput := &cato_models.PolicyRuleActivePeriodInput{
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
					parsedTime, err := parseTimeStringWithTZ(effectiveFromStr)
					if err != nil {
						diags = append(diags, diag.NewErrorDiagnostic(
							"Invalid effective_from time format",
							fmt.Sprintf("Unable to parse effective_from time '%s': %v. Expected RFC3339 format (e.g., '2024-12-31T23:59:59Z')", effectiveFromStr, err),
						))
					} else {
						activePeriodApiInput.EffectiveFrom = &parsedTime
						activePeriodApiInput.UseEffectiveFrom = true
					}
				}
			}

			// Handle expires_at
			if !activePeriodInput.ExpiresAt.IsNull() && !activePeriodInput.ExpiresAt.IsUnknown() {
				expiresAtStr := activePeriodInput.ExpiresAt.ValueString()
				if expiresAtStr != "" {
					// Parse the time string - support both RFC3339 format and human-readable formats
					parsedTime, err := parseTimeStringWithTZ(expiresAtStr)
					if err != nil {
						diags = append(diags, diag.NewErrorDiagnostic(
							"Invalid expires_at time format",
							fmt.Sprintf("Unable to parse expires_at time '%s': %v. Expected RFC3339 format (e.g., '2024-12-31T23:59:59Z')", expiresAtStr, err),
						))
					} else {
						activePeriodApiInput.ExpiresAt = &parsedTime
						activePeriodApiInput.UseExpiresAt = true
					}
				}
			}

			rootAddRule.ActivePeriod = activePeriodApiInput
			// For update, we need to create a separate struct with pointer fields for bools
			activePeriodUpdateApiInput := &cato_models.PolicyRuleActivePeriodUpdateInput{
				EffectiveFrom:    activePeriodApiInput.EffectiveFrom,
				ExpiresAt:        activePeriodApiInput.ExpiresAt,
				UseEffectiveFrom: &activePeriodApiInput.UseEffectiveFrom,
				UseExpiresAt:     &activePeriodApiInput.UseExpiresAt,
			}
			rootUpdateRule.ActivePeriod = activePeriodUpdateApiInput
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

		if !ruleInput.ConnectionOrigin.IsNull() && !ruleInput.ConnectionOrigin.IsUnknown() {
			rootAddRule.ConnectionOrigin = cato_models.ConnectionOriginEnum(ruleInput.ConnectionOrigin.ValueString())
			rootUpdateRule.ConnectionOrigin = (*cato_models.ConnectionOriginEnum)(ruleInput.ConnectionOrigin.ValueStringPointer())
		} else {
			rootAddRule.ConnectionOrigin = "ANY"
			connectionOrigin := "ANY"
			rootUpdateRule.ConnectionOrigin = (*cato_models.ConnectionOriginEnum)(&connectionOrigin)
		}

	}

	hydrateApiReturn.create.Rule = rootAddRule
	hydrateApiReturn.update.Rule = rootUpdateRule

	return hydrateApiReturn, diags
}
