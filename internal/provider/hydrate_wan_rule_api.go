package provider

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	cato_scalars "github.com/catonetworks/cato-go-sdk/scalars"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	// Import the correct package
)

// hydrateWanApiTypes create sub-types for both create and update calls to populate both entries
type hydrateWanApiTypes struct {
	create cato_models.WanFirewallAddRuleInput
	update cato_models.WanFirewallUpdateRuleInput
}

// hydrateWanApiRule takes in the current state/plan along with context and returns the created
// diagnostic data as well as cato api data used to either create or update WAN entries
func hydrateWanRuleApi(ctx context.Context, plan WanFirewallRule) (hydrateWanApiTypes, diag.Diagnostics) {
	diags := []diag.Diagnostic{}

	hydrateApiReturn := hydrateWanApiTypes{}
	hydrateApiReturn.create = cato_models.WanFirewallAddRuleInput{}
	hydrateApiReturn.update = cato_models.WanFirewallUpdateRuleInput{}
	hydrateApiReturn.create.At = &cato_models.PolicyRulePositionInput{}

	rootAddRule := &cato_models.WanFirewallAddRuleDataInput{}
	rootUpdateRule := &cato_models.WanFirewallUpdateRuleDataInput{}

	//setting at for creation only
	if !plan.At.IsNull() {

		positionInput := PolicyRulePositionInput{}
		diags = append(diags, plan.At.As(ctx, &positionInput, basetypes.ObjectAsOptions{})...)

		hydrateApiReturn.create.At.Position = (*cato_models.PolicyRulePositionEnum)(positionInput.Position.ValueStringPointer())
		hydrateApiReturn.create.At.Ref = positionInput.Ref.ValueStringPointer()

	}

	// setting rule
	if !plan.Rule.IsNull() {

		ruleInput := Policy_Policy_WanFirewall_Policy_Rules_Rule{}
		diags = append(diags, plan.Rule.As(ctx, &ruleInput, basetypes.ObjectAsOptions{})...)

		// setting source
		if !ruleInput.Source.IsNull() {

			ruleSourceInput := &cato_models.WanFirewallSourceInput{}
			ruleSourceUpdateInput := &cato_models.WanFirewallSourceUpdateInput{}

			sourceInput := Policy_Policy_WanFirewall_Policy_Rules_Rule_Source{}
			diags = append(diags, ruleInput.Source.As(ctx, &sourceInput, basetypes.ObjectAsOptions{})...)

			// setting source IP
			if !sourceInput.IP.IsUnknown() && !sourceInput.IP.IsNull() {
				diags = append(diags, sourceInput.IP.ElementsAs(ctx, &ruleSourceInput.IP, false)...)
				diags = append(diags, sourceInput.IP.ElementsAs(ctx, &ruleSourceUpdateInput.IP, false)...)
			}

			// setting source subnet
			if !sourceInput.Subnet.IsUnknown() && !sourceInput.Subnet.IsNull() {
				diags = append(diags, sourceInput.Subnet.ElementsAs(ctx, &ruleSourceInput.Subnet, false)...)
				diags = append(diags, sourceInput.Subnet.ElementsAs(ctx, &ruleSourceUpdateInput.Subnet, false)...)
			}

			// setting source host
			if !sourceInput.Host.IsUnknown() && !sourceInput.Host.IsNull() {
				elementsSourceHostInput := make([]types.Object, 0, len(sourceInput.Host.Elements()))
				diags = append(diags, sourceInput.Host.ElementsAs(ctx, &elementsSourceHostInput, false)...)

				var itemSourceHostInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_Host
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
			}

			// setting source site
			if !sourceInput.Site.IsUnknown() && !sourceInput.Site.IsNull() {
				elementsSourceSiteInput := make([]types.Object, 0, len(sourceInput.Site.Elements()))
				diags = append(diags, sourceInput.Site.ElementsAs(ctx, &elementsSourceSiteInput, false)...)

				var itemSourceSiteInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_Site
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
			}

			// setting source ip range
			if !sourceInput.IPRange.IsUnknown() && !sourceInput.IPRange.IsNull() {
				elementsSourceIPRangeInput := make([]types.Object, 0, len(sourceInput.IPRange.Elements()))
				diags = append(diags, sourceInput.IPRange.ElementsAs(ctx, &elementsSourceIPRangeInput, false)...)

				var itemSourceIPRangeInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_IPRange
				for _, item := range elementsSourceIPRangeInput {
					diags = append(diags, item.As(ctx, &itemSourceIPRangeInput, basetypes.ObjectAsOptions{})...)

					ruleSourceInput.IPRange = append(ruleSourceInput.IPRange, &cato_models.IPAddressRangeInput{
						From: itemSourceIPRangeInput.From.ValueString(),
						To:   itemSourceIPRangeInput.To.ValueString(),
					})
				}
				ruleSourceUpdateInput.IPRange = ruleSourceInput.IPRange
			}

			// setting source global ip range
			if !sourceInput.GlobalIPRange.IsNull() {
				elementsSourceGlobalIPRangeInput := make([]types.Object, 0, len(sourceInput.GlobalIPRange.Elements()))
				diags = append(diags, sourceInput.GlobalIPRange.ElementsAs(ctx, &elementsSourceGlobalIPRangeInput, false)...)

				var itemSourceGlobalIPRangeInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_GlobalIPRange
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
			}

			// setting source network interface
			if !sourceInput.NetworkInterface.IsNull() {
				elementsSourceNetworkInterfaceInput := make([]types.Object, 0, len(sourceInput.NetworkInterface.Elements()))
				diags = append(diags, sourceInput.NetworkInterface.ElementsAs(ctx, &elementsSourceNetworkInterfaceInput, false)...)

				var itemSourceNetworkInterfaceInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_NetworkInterface
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
			}

			// setting source site network subnet
			if !sourceInput.SiteNetworkSubnet.IsNull() {
				elementsSourceSiteNetworkSubnetInput := make([]types.Object, 0, len(sourceInput.SiteNetworkSubnet.Elements()))
				diags = append(diags, sourceInput.SiteNetworkSubnet.ElementsAs(ctx, &elementsSourceSiteNetworkSubnetInput, false)...)

				var itemSourceSiteNetworkSubnetInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_SiteNetworkSubnet
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
			}

			// setting source floating subnet
			if !sourceInput.FloatingSubnet.IsNull() {
				elementsSourceFloatingSubnetInput := make([]types.Object, 0, len(sourceInput.FloatingSubnet.Elements()))
				diags = append(diags, sourceInput.FloatingSubnet.ElementsAs(ctx, &elementsSourceFloatingSubnetInput, false)...)

				var itemSourceFloatingSubnetInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_FloatingSubnet
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
			}

			// setting source user
			if !sourceInput.User.IsNull() {
				elementsSourceUserInput := make([]types.Object, 0, len(sourceInput.User.Elements()))
				diags = append(diags, sourceInput.User.ElementsAs(ctx, &elementsSourceUserInput, false)...)

				var itemSourceUserInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_User
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
			}

			// setting source users group
			if !sourceInput.UsersGroup.IsNull() {
				elementsSourceUsersGroupInput := make([]types.Object, 0, len(sourceInput.UsersGroup.Elements()))
				diags = append(diags, sourceInput.UsersGroup.ElementsAs(ctx, &elementsSourceUsersGroupInput, false)...)

				var itemSourceUsersGroupInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_UsersGroup
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
			}

			// setting source group
			if !sourceInput.Group.IsNull() {
				elementsSourceGroupInput := make([]types.Object, 0, len(sourceInput.Group.Elements()))
				diags = append(diags, sourceInput.Group.ElementsAs(ctx, &elementsSourceGroupInput, false)...)

				var itemSourceGroupInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_Group
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
			}

			// setting source system group
			if !sourceInput.SystemGroup.IsNull() {
				elementsSourceSystemGroupInput := make([]types.Object, 0, len(sourceInput.SystemGroup.Elements()))
				diags = append(diags, sourceInput.SystemGroup.ElementsAs(ctx, &elementsSourceSystemGroupInput, false)...)

				var itemSourceSystemGroupInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_SystemGroup
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
			}
			rootAddRule.Source = ruleSourceInput
			rootUpdateRule.Source = ruleSourceUpdateInput
		}

		// setting country
		if !ruleInput.Country.IsNull() {
			elementsCountryInput := make([]types.Object, 0, len(ruleInput.Country.Elements()))
			diags = append(diags, ruleInput.Country.ElementsAs(ctx, &elementsCountryInput, false)...)

			var itemCountryInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Country
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
		}

		// setting device
		if !ruleInput.Device.IsNull() {
			elementsDeviceInput := make([]types.Object, 0, len(ruleInput.Device.Elements()))
			diags = append(diags, ruleInput.Device.ElementsAs(ctx, &elementsDeviceInput, false)...)

			var itemDeviceInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Device
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
		}

		// setting device OS
		if !ruleInput.DeviceOs.IsUnknown() && !ruleInput.DeviceOs.IsNull() {
			diags = append(diags, ruleInput.DeviceOs.ElementsAs(ctx, &rootAddRule.DeviceOs, false)...)
			diags = append(diags, ruleInput.DeviceOs.ElementsAs(ctx, &rootUpdateRule.DeviceOs, false)...)
		}

		// setting destination
		if !ruleInput.Destination.IsUnknown() && !ruleInput.Destination.IsNull() {

			ruleDestinationInput := &cato_models.WanFirewallDestinationInput{}
			ruleDestinationUpdateInput := &cato_models.WanFirewallDestinationUpdateInput{}

			destinationInput := Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination{}
			diags = append(diags, ruleInput.Destination.As(ctx, &destinationInput, basetypes.ObjectAsOptions{})...)

			// setting destination IP
			if !destinationInput.IP.IsUnknown() && !destinationInput.IP.IsNull() {
				diags = append(diags, destinationInput.IP.ElementsAs(ctx, &ruleDestinationInput.IP, false)...)
				diags = append(diags, destinationInput.IP.ElementsAs(ctx, &ruleDestinationUpdateInput.IP, false)...)
			}

			// setting destination subnet
			if !destinationInput.Subnet.IsUnknown() && !destinationInput.Subnet.IsNull() {
				diags = append(diags, destinationInput.Subnet.ElementsAs(ctx, &ruleDestinationInput.Subnet, false)...)
				diags = append(diags, destinationInput.Subnet.ElementsAs(ctx, &ruleDestinationUpdateInput.Subnet, false)...)
			}

			// setting destination host
			if !destinationInput.Host.IsNull() {
				elementsDestinationHostInput := make([]types.Object, 0, len(destinationInput.Host.Elements()))
				diags = append(diags, destinationInput.Host.ElementsAs(ctx, &elementsDestinationHostInput, false)...)

				var itemDestinationHostInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_Host
				for _, item := range elementsDestinationHostInput {
					diags = append(diags, item.As(ctx, &itemDestinationHostInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationHostInput)
					if err != nil {
						tflog.Error(ctx, err.Error())
					}

					ruleDestinationInput.Host = append(ruleDestinationInput.Host, &cato_models.HostRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
				ruleDestinationUpdateInput.Host = ruleDestinationInput.Host
			}

			// setting destination site
			if !destinationInput.Site.IsNull() {
				elementsDestinationSiteInput := make([]types.Object, 0, len(destinationInput.Site.Elements()))
				diags = append(diags, destinationInput.Site.ElementsAs(ctx, &elementsDestinationSiteInput, false)...)

				var itemDestinationSiteInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_Site
				for _, item := range elementsDestinationSiteInput {
					diags = append(diags, item.As(ctx, &itemDestinationSiteInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationSiteInput)
					if err != nil {
						tflog.Error(ctx, err.Error())
					}

					ruleDestinationInput.Site = append(ruleDestinationInput.Site, &cato_models.SiteRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
				ruleDestinationUpdateInput.Site = ruleDestinationInput.Site
			}

			// setting destination ip range
			if !destinationInput.IPRange.IsNull() {
				elementsDestinationIPRangeInput := make([]types.Object, 0, len(destinationInput.IPRange.Elements()))
				diags = append(diags, destinationInput.IPRange.ElementsAs(ctx, &elementsDestinationIPRangeInput, false)...)

				var itemDestinationIPRangeInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_IPRange
				for _, item := range elementsDestinationIPRangeInput {
					diags = append(diags, item.As(ctx, &itemDestinationIPRangeInput, basetypes.ObjectAsOptions{})...)

					ruleDestinationInput.IPRange = append(ruleDestinationInput.IPRange, &cato_models.IPAddressRangeInput{
						From: itemDestinationIPRangeInput.From.ValueString(),
						To:   itemDestinationIPRangeInput.To.ValueString(),
					})
				}
				ruleDestinationUpdateInput.IPRange = ruleDestinationInput.IPRange
			}

			// setting destination global ip range
			if !destinationInput.GlobalIPRange.IsNull() {
				elementsDestinationGlobalIPRangeInput := make([]types.Object, 0, len(destinationInput.GlobalIPRange.Elements()))
				diags = append(diags, destinationInput.GlobalIPRange.ElementsAs(ctx, &elementsDestinationGlobalIPRangeInput, false)...)

				var itemDestinationGlobalIPRangeInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_GlobalIPRange
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
			}

			// setting destination network interface
			if !destinationInput.NetworkInterface.IsNull() {
				elementsDestinationNetworkInterfaceInput := make([]types.Object, 0, len(destinationInput.NetworkInterface.Elements()))
				diags = append(diags, destinationInput.NetworkInterface.ElementsAs(ctx, &elementsDestinationNetworkInterfaceInput, false)...)

				var itemDestinationNetworkInterfaceInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_NetworkInterface
				for _, item := range elementsDestinationNetworkInterfaceInput {
					diags = append(diags, item.As(ctx, &itemDestinationNetworkInterfaceInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationNetworkInterfaceInput)
					if err != nil {
						tflog.Error(ctx, err.Error())
					}

					ruleDestinationInput.NetworkInterface = append(ruleDestinationInput.NetworkInterface, &cato_models.NetworkInterfaceRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
				ruleDestinationUpdateInput.NetworkInterface = ruleDestinationInput.NetworkInterface
			}

			// setting destination site network subnet
			if !destinationInput.SiteNetworkSubnet.IsNull() {
				elementsDestinationSiteNetworkSubnetInput := make([]types.Object, 0, len(destinationInput.SiteNetworkSubnet.Elements()))
				diags = append(diags, destinationInput.SiteNetworkSubnet.ElementsAs(ctx, &elementsDestinationSiteNetworkSubnetInput, false)...)

				var itemDestinationSiteNetworkSubnetInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_SiteNetworkSubnet
				for _, item := range elementsDestinationSiteNetworkSubnetInput {
					diags = append(diags, item.As(ctx, &itemDestinationSiteNetworkSubnetInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationSiteNetworkSubnetInput)
					if err != nil {
						tflog.Error(ctx, err.Error())
					}

					ruleDestinationInput.SiteNetworkSubnet = append(ruleDestinationInput.SiteNetworkSubnet, &cato_models.SiteNetworkSubnetRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
				ruleDestinationUpdateInput.SiteNetworkSubnet = ruleDestinationInput.SiteNetworkSubnet
			}

			// setting destination floating subnet
			if !destinationInput.FloatingSubnet.IsNull() {
				elementsDestinationFloatingSubnetInput := make([]types.Object, 0, len(destinationInput.FloatingSubnet.Elements()))
				diags = append(diags, destinationInput.FloatingSubnet.ElementsAs(ctx, &elementsDestinationFloatingSubnetInput, false)...)

				var itemDestinationFloatingSubnetInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_FloatingSubnet
				for _, item := range elementsDestinationFloatingSubnetInput {
					diags = append(diags, item.As(ctx, &itemDestinationFloatingSubnetInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationFloatingSubnetInput)
					if err != nil {
						tflog.Error(ctx, err.Error())
					}

					ruleDestinationInput.FloatingSubnet = append(ruleDestinationInput.FloatingSubnet, &cato_models.FloatingSubnetRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
				ruleDestinationUpdateInput.FloatingSubnet = ruleDestinationInput.FloatingSubnet
			}

			// setting destination user
			if !destinationInput.User.IsNull() {
				elementsDestinationUserInput := make([]types.Object, 0, len(destinationInput.User.Elements()))
				diags = append(diags, destinationInput.User.ElementsAs(ctx, &elementsDestinationUserInput, false)...)

				var itemDestinationUserInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_User
				for _, item := range elementsDestinationUserInput {
					diags = append(diags, item.As(ctx, &itemDestinationUserInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationUserInput)
					if err != nil {
						tflog.Error(ctx, err.Error())
					}

					ruleDestinationInput.User = append(ruleDestinationInput.User, &cato_models.UserRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
				ruleDestinationUpdateInput.User = ruleDestinationInput.User
			}

			// setting destination users group
			if !destinationInput.UsersGroup.IsNull() {
				elementsDestinationUsersGroupInput := make([]types.Object, 0, len(destinationInput.UsersGroup.Elements()))
				diags = append(diags, destinationInput.UsersGroup.ElementsAs(ctx, &elementsDestinationUsersGroupInput, false)...)

				var itemDestinationUsersGroupInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_UsersGroup
				for _, item := range elementsDestinationUsersGroupInput {
					diags = append(diags, item.As(ctx, &itemDestinationUsersGroupInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationUsersGroupInput)
					if err != nil {
						tflog.Error(ctx, err.Error())
					}

					ruleDestinationInput.UsersGroup = append(ruleDestinationInput.UsersGroup, &cato_models.UsersGroupRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
				ruleDestinationUpdateInput.UsersGroup = ruleDestinationInput.UsersGroup
			}

			// setting destination group
			if !destinationInput.Group.IsNull() {
				elementsDestinationGroupInput := make([]types.Object, 0, len(destinationInput.Group.Elements()))
				diags = append(diags, destinationInput.Group.ElementsAs(ctx, &elementsDestinationGroupInput, false)...)

				var itemDestinationGroupInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_Group
				for _, item := range elementsDestinationGroupInput {
					diags = append(diags, item.As(ctx, &itemDestinationGroupInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationGroupInput)
					if err != nil {
						tflog.Error(ctx, err.Error())
					}

					ruleDestinationInput.Group = append(ruleDestinationInput.Group, &cato_models.GroupRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
				ruleDestinationUpdateInput.Group = ruleDestinationInput.Group
			}

			// setting destination system group
			if !destinationInput.SystemGroup.IsNull() {
				elementsDestinationSystemGroupInput := make([]types.Object, 0, len(destinationInput.SystemGroup.Elements()))
				diags = append(diags, destinationInput.SystemGroup.ElementsAs(ctx, &elementsDestinationSystemGroupInput, false)...)

				var itemDestinationSystemGroupInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_SystemGroup
				for _, item := range elementsDestinationSystemGroupInput {
					diags = append(diags, item.As(ctx, &itemDestinationSystemGroupInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationSystemGroupInput)
					if err != nil {
						tflog.Error(ctx, err.Error())
					}

					ruleDestinationInput.SystemGroup = append(ruleDestinationInput.SystemGroup, &cato_models.SystemGroupRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
				ruleDestinationUpdateInput.SystemGroup = ruleDestinationInput.SystemGroup
			}
			rootAddRule.Destination = ruleDestinationInput
			rootUpdateRule.Destination = ruleDestinationUpdateInput
		}

		// setting application
		if !ruleInput.Application.IsUnknown() && !ruleInput.Application.IsNull() {
			ruleApplicationInput := &cato_models.WanFirewallApplicationInput{}
			ruleApplicationUpdateInput := &cato_models.WanFirewallApplicationUpdateInput{}

			applicationInput := Policy_Policy_WanFirewall_Policy_Rules_Rule_Application{}
			diags = append(diags, ruleInput.Application.As(ctx, &applicationInput, basetypes.ObjectAsOptions{})...)

			// setting application IP
			if !applicationInput.IP.IsUnknown() && !applicationInput.IP.IsNull() {
				diags = append(diags, applicationInput.IP.ElementsAs(ctx, &ruleApplicationInput.IP, false)...)
				diags = append(diags, applicationInput.IP.ElementsAs(ctx, &ruleApplicationUpdateInput.IP, false)...)
			}

			// setting application subnet
			if !applicationInput.Subnet.IsUnknown() && !applicationInput.Subnet.IsNull() {
				diags = append(diags, applicationInput.IP.ElementsAs(ctx, &ruleApplicationInput.Subnet, false)...)
				diags = append(diags, applicationInput.IP.ElementsAs(ctx, &ruleApplicationUpdateInput.Subnet, false)...)
			}

			// setting application domain
			if !applicationInput.Domain.IsUnknown() && !applicationInput.Domain.IsNull() {
				diags = append(diags, applicationInput.IP.ElementsAs(ctx, &ruleApplicationInput.Domain, false)...)
				diags = append(diags, applicationInput.IP.ElementsAs(ctx, &ruleApplicationUpdateInput.Domain, false)...)
			}

			// setting application fqdn
			if !applicationInput.Fqdn.IsUnknown() && !applicationInput.Fqdn.IsNull() {
				diags = append(diags, applicationInput.IP.ElementsAs(ctx, &ruleApplicationInput.Fqdn, false)...)
				diags = append(diags, applicationInput.IP.ElementsAs(ctx, &ruleApplicationUpdateInput.Fqdn, false)...)
			}

			// setting application application
			if !applicationInput.Application.IsUnknown() && !applicationInput.Application.IsNull() {
				elementsDestinationApplicationInput := make([]types.Object, 0, len(applicationInput.Application.Elements()))
				diags = append(diags, applicationInput.Application.ElementsAs(ctx, &elementsDestinationApplicationInput, false)...)

				var itemDestinationApplicationInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_Application
				for _, item := range elementsDestinationApplicationInput {
					diags = append(diags, item.As(ctx, &itemDestinationApplicationInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationApplicationInput)
					if err != nil {
						tflog.Error(ctx, err.Error())
					}

					ruleApplicationInput.Application = append(ruleApplicationInput.Application, &cato_models.ApplicationRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
				ruleApplicationUpdateInput.Application = ruleApplicationInput.Application
			}

			// setting application custom app
			if !applicationInput.CustomApp.IsNull() {
				elementsDestinationCustomAppInput := make([]types.Object, 0, len(applicationInput.CustomApp.Elements()))
				diags = append(diags, applicationInput.CustomApp.ElementsAs(ctx, &elementsDestinationCustomAppInput, false)...)

				var itemDestinationCustomAppInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_CustomApp
				for _, item := range elementsDestinationCustomAppInput {
					diags = append(diags, item.As(ctx, &itemDestinationCustomAppInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationCustomAppInput)
					if err != nil {
						tflog.Error(ctx, err.Error())
					}

					ruleApplicationInput.CustomApp = append(ruleApplicationInput.CustomApp, &cato_models.CustomApplicationRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
				ruleApplicationUpdateInput.CustomApp = ruleApplicationInput.CustomApp
			}

			// setting application ip range
			if !applicationInput.IPRange.IsNull() {
				elementsApplicationIPRangeInput := make([]types.Object, 0, len(applicationInput.IPRange.Elements()))
				diags = append(diags, applicationInput.IPRange.ElementsAs(ctx, &elementsApplicationIPRangeInput, false)...)

				var itemApplicationIPRangeInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_IPRange
				for _, item := range elementsApplicationIPRangeInput {
					diags = append(diags, item.As(ctx, &itemApplicationIPRangeInput, basetypes.ObjectAsOptions{})...)

					ruleApplicationInput.IPRange = append(ruleApplicationInput.IPRange, &cato_models.IPAddressRangeInput{
						From: itemApplicationIPRangeInput.From.ValueString(),
						To:   itemApplicationIPRangeInput.To.ValueString(),
					})
				}
				ruleApplicationUpdateInput.IPRange = ruleApplicationInput.IPRange
			}

			// setting application global ip range
			if !applicationInput.GlobalIPRange.IsNull() {
				elementsApplicationGlobalIPRangeInput := make([]types.Object, 0, len(applicationInput.GlobalIPRange.Elements()))
				diags = append(diags, applicationInput.GlobalIPRange.ElementsAs(ctx, &elementsApplicationGlobalIPRangeInput, false)...)

				var itemApplicationGlobalIPRangeInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_GlobalIPRange
				for _, item := range elementsApplicationGlobalIPRangeInput {
					diags = append(diags, item.As(ctx, &itemApplicationGlobalIPRangeInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationGlobalIPRangeInput)
					if err != nil {
						tflog.Error(ctx, err.Error())
					}

					ruleApplicationInput.GlobalIPRange = append(ruleApplicationInput.GlobalIPRange, &cato_models.GlobalIPRangeRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
				ruleApplicationUpdateInput.GlobalIPRange = ruleApplicationInput.GlobalIPRange
			}

			// setting application app category
			if !applicationInput.AppCategory.IsUnknown() && !applicationInput.AppCategory.IsNull() {
				elementsApplicationAppCategoryInput := make([]types.Object, 0, len(applicationInput.AppCategory.Elements()))
				diags = append(diags, applicationInput.AppCategory.ElementsAs(ctx, &elementsApplicationAppCategoryInput, false)...)

				var itemApplicationAppCategoryInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_AppCategory
				for _, item := range elementsApplicationAppCategoryInput {
					diags = append(diags, item.As(ctx, &itemApplicationAppCategoryInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationAppCategoryInput)
					if err != nil {
						tflog.Error(ctx, err.Error())
					}

					ruleApplicationInput.AppCategory = append(ruleApplicationInput.AppCategory, &cato_models.ApplicationCategoryRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
				ruleApplicationUpdateInput.AppCategory = ruleApplicationInput.AppCategory
			}

			// setting application custom app category
			if !applicationInput.CustomCategory.IsNull() {
				elementsApplicationCustomCategoryInput := make([]types.Object, 0, len(applicationInput.CustomCategory.Elements()))
				diags = append(diags, applicationInput.CustomCategory.ElementsAs(ctx, &elementsApplicationCustomCategoryInput, false)...)

				var itemApplicationCustomCategoryInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_CustomCategory
				for _, item := range elementsApplicationCustomCategoryInput {
					diags = append(diags, item.As(ctx, &itemApplicationCustomCategoryInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationCustomCategoryInput)
					if err != nil {
						tflog.Error(ctx, err.Error())
					}

					ruleApplicationInput.CustomCategory = append(ruleApplicationInput.CustomCategory, &cato_models.CustomCategoryRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
				ruleApplicationUpdateInput.CustomCategory = ruleApplicationInput.CustomCategory
			}

			// setting application sanctionned apps category
			if !applicationInput.SanctionedAppsCategory.IsNull() {
				elementsApplicationSanctionedAppsCategoryInput := make([]types.Object, 0, len(applicationInput.SanctionedAppsCategory.Elements()))
				diags = append(diags, applicationInput.SanctionedAppsCategory.ElementsAs(ctx, &elementsApplicationSanctionedAppsCategoryInput, false)...)

				var itemApplicationSanctionedAppsCategoryInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_SanctionedAppsCategory
				for _, item := range elementsApplicationSanctionedAppsCategoryInput {
					diags = append(diags, item.As(ctx, &itemApplicationSanctionedAppsCategoryInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationSanctionedAppsCategoryInput)
					if err != nil {
						tflog.Error(ctx, err.Error())
					}

					ruleApplicationInput.SanctionedAppsCategory = append(ruleApplicationInput.SanctionedAppsCategory, &cato_models.SanctionedAppsCategoryRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
				ruleApplicationUpdateInput.SanctionedAppsCategory = ruleApplicationInput.SanctionedAppsCategory
			}
			rootAddRule.Application = ruleApplicationInput
			rootUpdateRule.Application = ruleApplicationUpdateInput
		}

		// setting service
		if !ruleInput.Service.IsNull() {
			ruleServiceInput := &cato_models.WanFirewallServiceTypeInput{}
			ruleServiceUpdateInput := &cato_models.WanFirewallServiceTypeUpdateInput{}

			serviceInput := Policy_Policy_WanFirewall_Policy_Rules_Rule_Service{}
			diags = append(diags, ruleInput.Service.As(ctx, &serviceInput, basetypes.ObjectAsOptions{})...)

			// setting service standard
			if !serviceInput.Standard.IsNull() {
				elementsServiceStandardInput := make([]types.Object, 0, len(serviceInput.Standard.Elements()))
				diags = append(diags, serviceInput.Standard.ElementsAs(ctx, &elementsServiceStandardInput, false)...)

				var itemServiceStandardInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Service_Standard
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
			}

			// setting service custom
			if !serviceInput.Custom.IsNull() {
				elementsServiceCustomInput := make([]types.Object, 0, len(serviceInput.Custom.Elements()))
				diags = append(diags, serviceInput.Custom.ElementsAs(ctx, &elementsServiceCustomInput, false)...)

				var itemServiceCustomInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Service_Custom
				for _, item := range elementsServiceCustomInput {
					diags = append(diags, item.As(ctx, &itemServiceCustomInput, basetypes.ObjectAsOptions{})...)

					customInput := &cato_models.CustomServiceInput{
						Protocol: cato_models.IPProtocol(itemServiceCustomInput.Protocol.ValueString()),
					}

					// setting service custom port
					if !itemServiceCustomInput.Port.IsNull() {
						elementsPort := make([]types.String, 0, len(itemServiceCustomInput.Port.Elements()))
						diags = append(diags, itemServiceCustomInput.Port.ElementsAs(ctx, &elementsPort, false)...)

						inputPort := []cato_scalars.Port{}
						for _, item := range elementsPort {
							inputPort = append(inputPort, cato_scalars.Port(item.ValueString()))
						}

						customInput.Port = inputPort
					}

					// setting service custom port range
					if !itemServiceCustomInput.PortRange.IsNull() {
						var itemPortRange Policy_Policy_WanFirewall_Policy_Rules_Rule_Service_Custom_PortRange
						diags = append(diags, itemServiceCustomInput.PortRange.As(ctx, &itemPortRange, basetypes.ObjectAsOptions{})...)

						inputPortRange := cato_models.PortRangeInput{
							From: cato_scalars.Port(itemPortRange.From.ValueString()),
							To:   cato_scalars.Port(itemPortRange.To.ValueString()),
						}

						customInput.PortRange = &inputPortRange
					}

					// append custom service
					ruleServiceInput.Custom = append(ruleServiceInput.Custom, customInput)
				}
				ruleServiceUpdateInput.Custom = ruleServiceInput.Custom
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

			trackingInput := Policy_Policy_WanFirewall_Policy_Rules_Rule_Tracking{}
			diags = append(diags, ruleInput.Tracking.As(ctx, &trackingInput, basetypes.ObjectAsOptions{})...)

			if !trackingInput.Event.IsUnknown() && !trackingInput.Event.IsNull() {
				// setting tracking event
				trackingEventInput := Policy_Policy_WanFirewall_Policy_Rules_Rule_Tracking_Event{}
				diags = append(diags, trackingInput.Event.As(ctx, &trackingEventInput, basetypes.ObjectAsOptions{})...)
				rootAddRule.Tracking.Event.Enabled = trackingEventInput.Enabled.ValueBool()
				rootUpdateRule.Tracking.Event.Enabled = trackingEventInput.Enabled.ValueBoolPointer()
			}

			if !trackingInput.Alert.IsUnknown() && !trackingInput.Alert.IsNull() {

				rootAddRule.Tracking.Alert = &cato_models.PolicyRuleTrackingAlertInput{}

				trackingAlertInput := Policy_Policy_WanFirewall_Policy_Rules_Rule_Tracking_Alert{}
				diags = append(diags, trackingInput.Alert.As(ctx, &trackingAlertInput, basetypes.ObjectAsOptions{})...)

				rootAddRule.Tracking.Alert.Enabled = trackingAlertInput.Enabled.ValueBool()
				rootAddRule.Tracking.Alert.Frequency = (cato_models.PolicyRuleTrackingFrequencyEnum)(trackingAlertInput.Frequency.ValueString())

				rootUpdateRule.Tracking.Alert.Enabled = trackingAlertInput.Enabled.ValueBoolPointer()
				rootUpdateRule.Tracking.Alert.Frequency = (*cato_models.PolicyRuleTrackingFrequencyEnum)(trackingAlertInput.Frequency.ValueStringPointer())

				// setting tracking alert subscription group
				if !trackingAlertInput.SubscriptionGroup.IsNull() {
					elementsAlertSubscriptionGroupInput := make([]types.Object, 0, len(trackingAlertInput.SubscriptionGroup.Elements()))
					diags = append(diags, trackingAlertInput.SubscriptionGroup.ElementsAs(ctx, &elementsAlertSubscriptionGroupInput, false)...)

					var itemAlertSubscriptionGroupInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Tracking_Alert_SubscriptionGroup
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
				}

				// setting tracking alert webhook
				if !trackingAlertInput.Webhook.IsNull() {
					if !trackingAlertInput.Webhook.IsNull() {
						elementsAlertWebHookInput := make([]types.Object, 0, len(trackingAlertInput.Webhook.Elements()))
						diags = append(diags, trackingAlertInput.Webhook.ElementsAs(ctx, &elementsAlertWebHookInput, false)...)

						var itemAlertWebHookInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Tracking_Alert_SubscriptionGroup
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
					}
				}

				// setting tracking alert mailing list
				if !trackingAlertInput.MailingList.IsNull() {
					elementsAlertMailingListInput := make([]types.Object, 0, len(trackingAlertInput.MailingList.Elements()))
					diags = append(diags, trackingAlertInput.MailingList.ElementsAs(ctx, &elementsAlertMailingListInput, false)...)

					var itemAlertMailingListInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Tracking_Alert_SubscriptionGroup
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

			scheduleInput := Policy_Policy_WanFirewall_Policy_Rules_Rule_Schedule{}
			diags = append(diags, ruleInput.Schedule.As(ctx, &scheduleInput, basetypes.ObjectAsOptions{})...)

			rootAddRule.Schedule.ActiveOn = cato_models.PolicyActiveOnEnum(scheduleInput.ActiveOn.ValueString())
			rootUpdateRule.Schedule.ActiveOn = (*cato_models.PolicyActiveOnEnum)(scheduleInput.ActiveOn.ValueStringPointer())

			// setting schedule custome time frame
			if !scheduleInput.CustomTimeframe.IsNull() {
				rootAddRule.Schedule.CustomTimeframe = &cato_models.PolicyCustomTimeframeInput{}
				rootUpdateRule.Schedule.CustomTimeframe = &cato_models.PolicyCustomTimeframeUpdateInput{}

				customeTimeFrameInput := Policy_Policy_WanFirewall_Policy_Rules_Rule_Schedule_CustomTimeframe{}
				diags = append(diags, scheduleInput.CustomTimeframe.As(ctx, &customeTimeFrameInput, basetypes.ObjectAsOptions{})...)

				rootAddRule.Schedule.CustomTimeframe.From = customeTimeFrameInput.From.ValueString()
				rootAddRule.Schedule.CustomTimeframe.To = customeTimeFrameInput.To.ValueString()

				rootUpdateRule.Schedule.CustomTimeframe.From = customeTimeFrameInput.From.ValueStringPointer()
				rootUpdateRule.Schedule.CustomTimeframe.To = customeTimeFrameInput.To.ValueStringPointer()
			}

			// setting schedule custom recurring
			if !scheduleInput.CustomRecurring.IsNull() {
				rootAddRule.Schedule.CustomRecurring = &cato_models.PolicyCustomRecurringInput{}
				rootUpdateRule.Schedule.CustomRecurring = &cato_models.PolicyCustomRecurringUpdateInput{}

				customRecurringInput := Policy_Policy_WanFirewall_Policy_Rules_Rule_Schedule_CustomRecurring{}
				diags = append(diags, scheduleInput.CustomRecurring.As(ctx, &customRecurringInput, basetypes.ObjectAsOptions{})...)

				rootAddRule.Schedule.CustomRecurring.From = cato_scalars.Time(customRecurringInput.From.ValueString())
				rootAddRule.Schedule.CustomRecurring.To = cato_scalars.Time(customRecurringInput.To.ValueString())
				rootUpdateRule.Schedule.CustomRecurring.From = (*cato_scalars.Time)(customRecurringInput.From.ValueStringPointer())
				rootUpdateRule.Schedule.CustomRecurring.To = (*cato_scalars.Time)(customRecurringInput.To.ValueStringPointer())

				// setting schedule custom recurring days
				diags = append(diags, customRecurringInput.Days.ElementsAs(ctx, &rootAddRule.Schedule.CustomRecurring.Days, false)...)
				rootUpdateRule.Schedule.CustomRecurring.Days = rootAddRule.Schedule.CustomRecurring.Days
			}
		}

		// settings exceptions
		if !ruleInput.Exceptions.IsNull() && !ruleInput.Exceptions.IsUnknown() {
			elementsExceptionsInput := make([]types.Object, 0, len(ruleInput.Exceptions.Elements()))
			diags = append(diags, ruleInput.Exceptions.ElementsAs(ctx, &elementsExceptionsInput, false)...)

			// loop over exceptions
			var itemExceptionsInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Exceptions
			for _, item := range elementsExceptionsInput {

				exceptionInput := cato_models.WanFirewallRuleExceptionInput{}

				diags = append(diags, item.As(ctx, &itemExceptionsInput, basetypes.ObjectAsOptions{})...)

				// setting exception name
				exceptionInput.Name = itemExceptionsInput.Name.ValueString()

				// setting exception direction
				exceptionInput.Direction = cato_models.WanFirewallDirectionEnum(itemExceptionsInput.Direction.ValueString())

				// setting exception connection origin
				if !itemExceptionsInput.ConnectionOrigin.IsNull() {
					exceptionInput.ConnectionOrigin = cato_models.ConnectionOriginEnum(itemExceptionsInput.ConnectionOrigin.ValueString())
				} else {
					exceptionInput.ConnectionOrigin = cato_models.ConnectionOriginEnum("ANY")
				}

				// setting source
				if !itemExceptionsInput.Source.IsNull() {

					exceptionInput.Source = &cato_models.WanFirewallSourceInput{}
					sourceInput := Policy_Policy_WanFirewall_Policy_Rules_Rule_Source{}
					diags = append(diags, itemExceptionsInput.Source.As(ctx, &sourceInput, basetypes.ObjectAsOptions{})...)

					// setting source IP
					if !sourceInput.IP.IsNull() {
						diags = append(diags, sourceInput.IP.ElementsAs(ctx, &exceptionInput.Source.IP, false)...)
					}

					// setting source subnet
					if !sourceInput.Subnet.IsNull() {
						diags = append(diags, sourceInput.Subnet.ElementsAs(ctx, &exceptionInput.Source.Subnet, false)...)
					}

					// setting source host
					if !sourceInput.Host.IsNull() {
						elementsSourceHostInput := make([]types.Object, 0, len(sourceInput.Host.Elements()))
						diags = append(diags, sourceInput.Host.ElementsAs(ctx, &elementsSourceHostInput, false)...)

						var itemSourceHostInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_Host
						for _, item := range elementsSourceHostInput {
							diags = append(diags, item.As(ctx, &itemSourceHostInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceHostInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionInput.Source.Host = append(exceptionInput.Source.Host, &cato_models.HostRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting source site
					if !sourceInput.Site.IsNull() {
						elementsSourceSiteInput := make([]types.Object, 0, len(sourceInput.Site.Elements()))
						diags = append(diags, sourceInput.Site.ElementsAs(ctx, &elementsSourceSiteInput, false)...)

						var itemSourceSiteInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_Site
						for _, item := range elementsSourceSiteInput {
							diags = append(diags, item.As(ctx, &itemSourceSiteInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceSiteInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionInput.Source.Site = append(exceptionInput.Source.Site, &cato_models.SiteRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting source ip range
					if !sourceInput.IPRange.IsNull() {
						elementsSourceIPRangeInput := make([]types.Object, 0, len(sourceInput.IPRange.Elements()))
						diags = append(diags, sourceInput.IPRange.ElementsAs(ctx, &elementsSourceIPRangeInput, false)...)

						var itemSourceIPRangeInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_IPRange
						for _, item := range elementsSourceIPRangeInput {
							diags = append(diags, item.As(ctx, &itemSourceIPRangeInput, basetypes.ObjectAsOptions{})...)

							exceptionInput.Source.IPRange = append(exceptionInput.Source.IPRange, &cato_models.IPAddressRangeInput{
								From: itemSourceIPRangeInput.From.ValueString(),
								To:   itemSourceIPRangeInput.To.ValueString(),
							})
						}
					}

					// setting source global ip range
					if !sourceInput.GlobalIPRange.IsNull() {
						elementsSourceGlobalIPRangeInput := make([]types.Object, 0, len(sourceInput.GlobalIPRange.Elements()))
						diags = append(diags, sourceInput.GlobalIPRange.ElementsAs(ctx, &elementsSourceGlobalIPRangeInput, false)...)

						var itemSourceGlobalIPRangeInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_GlobalIPRange
						for _, item := range elementsSourceGlobalIPRangeInput {
							diags = append(diags, item.As(ctx, &itemSourceGlobalIPRangeInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceGlobalIPRangeInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionInput.Source.GlobalIPRange = append(exceptionInput.Source.GlobalIPRange, &cato_models.GlobalIPRangeRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting source network interface
					if !sourceInput.NetworkInterface.IsNull() {
						elementsSourceNetworkInterfaceInput := make([]types.Object, 0, len(sourceInput.NetworkInterface.Elements()))
						diags = append(diags, sourceInput.NetworkInterface.ElementsAs(ctx, &elementsSourceNetworkInterfaceInput, false)...)

						var itemSourceNetworkInterfaceInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_NetworkInterface
						for _, item := range elementsSourceNetworkInterfaceInput {
							diags = append(diags, item.As(ctx, &itemSourceNetworkInterfaceInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceNetworkInterfaceInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionInput.Source.NetworkInterface = append(exceptionInput.Source.NetworkInterface, &cato_models.NetworkInterfaceRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting source site network subnet
					if !sourceInput.SiteNetworkSubnet.IsNull() {
						elementsSourceSiteNetworkSubnetInput := make([]types.Object, 0, len(sourceInput.SiteNetworkSubnet.Elements()))
						diags = append(diags, sourceInput.SiteNetworkSubnet.ElementsAs(ctx, &elementsSourceSiteNetworkSubnetInput, false)...)

						var itemSourceSiteNetworkSubnetInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_SiteNetworkSubnet
						for _, item := range elementsSourceSiteNetworkSubnetInput {
							diags = append(diags, item.As(ctx, &itemSourceSiteNetworkSubnetInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceSiteNetworkSubnetInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionInput.Source.SiteNetworkSubnet = append(exceptionInput.Source.SiteNetworkSubnet, &cato_models.SiteNetworkSubnetRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting source floating subnet
					if !sourceInput.FloatingSubnet.IsNull() {
						elementsSourceFloatingSubnetInput := make([]types.Object, 0, len(sourceInput.FloatingSubnet.Elements()))
						diags = append(diags, sourceInput.FloatingSubnet.ElementsAs(ctx, &elementsSourceFloatingSubnetInput, false)...)

						var itemSourceFloatingSubnetInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_FloatingSubnet
						for _, item := range elementsSourceFloatingSubnetInput {
							diags = append(diags, item.As(ctx, &itemSourceFloatingSubnetInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceFloatingSubnetInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionInput.Source.FloatingSubnet = append(exceptionInput.Source.FloatingSubnet, &cato_models.FloatingSubnetRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting source user
					if !sourceInput.User.IsNull() {
						elementsSourceUserInput := make([]types.Object, 0, len(sourceInput.User.Elements()))
						diags = append(diags, sourceInput.User.ElementsAs(ctx, &elementsSourceUserInput, false)...)

						var itemSourceUserInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_User
						for _, item := range elementsSourceUserInput {
							diags = append(diags, item.As(ctx, &itemSourceUserInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceUserInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionInput.Source.User = append(exceptionInput.Source.User, &cato_models.UserRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting source users group
					if !sourceInput.UsersGroup.IsNull() {
						elementsSourceUsersGroupInput := make([]types.Object, 0, len(sourceInput.UsersGroup.Elements()))
						diags = append(diags, sourceInput.UsersGroup.ElementsAs(ctx, &elementsSourceUsersGroupInput, false)...)

						var itemSourceUsersGroupInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_UsersGroup
						for _, item := range elementsSourceUsersGroupInput {
							diags = append(diags, item.As(ctx, &itemSourceUsersGroupInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceUsersGroupInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionInput.Source.UsersGroup = append(exceptionInput.Source.UsersGroup, &cato_models.UsersGroupRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting source group
					if !sourceInput.Group.IsNull() {
						elementsSourceGroupInput := make([]types.Object, 0, len(sourceInput.Group.Elements()))
						diags = append(diags, sourceInput.Group.ElementsAs(ctx, &elementsSourceGroupInput, false)...)

						var itemSourceGroupInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_Group
						for _, item := range elementsSourceGroupInput {
							diags = append(diags, item.As(ctx, &itemSourceGroupInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceGroupInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionInput.Source.Group = append(exceptionInput.Source.Group, &cato_models.GroupRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting source system group
					if !sourceInput.SystemGroup.IsNull() {
						elementsSourceSystemGroupInput := make([]types.Object, 0, len(sourceInput.SystemGroup.Elements()))
						diags = append(diags, sourceInput.SystemGroup.ElementsAs(ctx, &elementsSourceSystemGroupInput, false)...)

						var itemSourceSystemGroupInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_SystemGroup
						for _, item := range elementsSourceSystemGroupInput {
							diags = append(diags, item.As(ctx, &itemSourceSystemGroupInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceSystemGroupInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionInput.Source.SystemGroup = append(exceptionInput.Source.SystemGroup, &cato_models.SystemGroupRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}
				}

				// setting country
				if !itemExceptionsInput.Country.IsNull() {

					exceptionInput.Country = []*cato_models.CountryRefInput{}
					elementsCountryInput := make([]types.Object, 0, len(itemExceptionsInput.Country.Elements()))
					diags = append(diags, itemExceptionsInput.Country.ElementsAs(ctx, &elementsCountryInput, false)...)

					var itemCountryInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Country
					for _, item := range elementsCountryInput {
						diags = append(diags, item.As(ctx, &itemCountryInput, basetypes.ObjectAsOptions{})...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemCountryInput)
						if err != nil {
							tflog.Error(ctx, err.Error())
						}

						exceptionInput.Country = append(exceptionInput.Country, &cato_models.CountryRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
				}

				// setting device
				if !itemExceptionsInput.Device.IsNull() {

					exceptionInput.Device = []*cato_models.DeviceProfileRefInput{}
					elementsDeviceInput := make([]types.Object, 0, len(itemExceptionsInput.Device.Elements()))
					diags = append(diags, itemExceptionsInput.Device.ElementsAs(ctx, &elementsDeviceInput, false)...)

					var itemDeviceInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Device
					for _, item := range elementsDeviceInput {
						diags = append(diags, item.As(ctx, &itemDeviceInput, basetypes.ObjectAsOptions{})...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemDeviceInput)
						if err != nil {
							tflog.Error(ctx, err.Error())
						}

						exceptionInput.Device = append(exceptionInput.Device, &cato_models.DeviceProfileRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
				}

				// setting device OS
				if !itemExceptionsInput.DeviceOs.IsUnknown() && !itemExceptionsInput.DeviceOs.IsNull() {
					diags = append(diags, itemExceptionsInput.DeviceOs.ElementsAs(ctx, &exceptionInput.DeviceOs, false)...)
				}

				// setting destination
				if !itemExceptionsInput.Destination.IsNull() {

					exceptionInput.Destination = &cato_models.WanFirewallDestinationInput{}
					destinationInput := Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination{}
					diags = append(diags, itemExceptionsInput.Destination.As(ctx, &destinationInput, basetypes.ObjectAsOptions{})...)

					// setting destination IP
					if !destinationInput.IP.IsUnknown() && !destinationInput.IP.IsNull() {
						diags = append(diags, destinationInput.IP.ElementsAs(ctx, &exceptionInput.Destination.IP, false)...)
					}

					// setting destination subnet
					if !destinationInput.Subnet.IsUnknown() && !destinationInput.Subnet.IsNull() {
						diags = append(diags, destinationInput.Subnet.ElementsAs(ctx, &exceptionInput.Destination.Subnet, false)...)
					}

					// setting destination host
					if !destinationInput.Host.IsNull() {
						elementsDestinationHostInput := make([]types.Object, 0, len(destinationInput.Host.Elements()))
						diags = append(diags, destinationInput.Host.ElementsAs(ctx, &elementsDestinationHostInput, false)...)

						var itemDestinationHostInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_Host
						for _, item := range elementsDestinationHostInput {
							diags = append(diags, item.As(ctx, &itemDestinationHostInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationHostInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionInput.Destination.Host = append(exceptionInput.Destination.Host, &cato_models.HostRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting destination site
					if !destinationInput.Site.IsNull() {
						elementsDestinationSiteInput := make([]types.Object, 0, len(destinationInput.Site.Elements()))
						diags = append(diags, destinationInput.Site.ElementsAs(ctx, &elementsDestinationSiteInput, false)...)

						var itemDestinationSiteInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_Site
						for _, item := range elementsDestinationSiteInput {
							diags = append(diags, item.As(ctx, &itemDestinationSiteInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationSiteInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionInput.Destination.Site = append(exceptionInput.Destination.Site, &cato_models.SiteRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting destination ip range
					if !destinationInput.IPRange.IsNull() {
						elementsDestinationIPRangeInput := make([]types.Object, 0, len(destinationInput.IPRange.Elements()))
						diags = append(diags, destinationInput.IPRange.ElementsAs(ctx, &elementsDestinationIPRangeInput, false)...)

						var itemDestinationIPRangeInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_IPRange
						for _, item := range elementsDestinationIPRangeInput {
							diags = append(diags, item.As(ctx, &itemDestinationIPRangeInput, basetypes.ObjectAsOptions{})...)

							exceptionInput.Destination.IPRange = append(exceptionInput.Destination.IPRange, &cato_models.IPAddressRangeInput{
								From: itemDestinationIPRangeInput.From.ValueString(),
								To:   itemDestinationIPRangeInput.To.ValueString(),
							})
						}
					}

					// setting destination global ip range
					if !destinationInput.GlobalIPRange.IsNull() {
						elementsDestinationGlobalIPRangeInput := make([]types.Object, 0, len(destinationInput.GlobalIPRange.Elements()))
						diags = append(diags, destinationInput.GlobalIPRange.ElementsAs(ctx, &elementsDestinationGlobalIPRangeInput, false)...)

						var itemDestinationGlobalIPRangeInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_GlobalIPRange
						for _, item := range elementsDestinationGlobalIPRangeInput {
							diags = append(diags, item.As(ctx, &itemDestinationGlobalIPRangeInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationGlobalIPRangeInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionInput.Destination.GlobalIPRange = append(exceptionInput.Destination.GlobalIPRange, &cato_models.GlobalIPRangeRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting destination network interface
					if !destinationInput.NetworkInterface.IsNull() {
						elementsDestinationNetworkInterfaceInput := make([]types.Object, 0, len(destinationInput.NetworkInterface.Elements()))
						diags = append(diags, destinationInput.NetworkInterface.ElementsAs(ctx, &elementsDestinationNetworkInterfaceInput, false)...)

						var itemDestinationNetworkInterfaceInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_NetworkInterface
						for _, item := range elementsDestinationNetworkInterfaceInput {
							diags = append(diags, item.As(ctx, &itemDestinationNetworkInterfaceInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationNetworkInterfaceInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionInput.Destination.NetworkInterface = append(exceptionInput.Destination.NetworkInterface, &cato_models.NetworkInterfaceRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting destination site network subnet
					if !destinationInput.SiteNetworkSubnet.IsNull() {
						elementsDestinationSiteNetworkSubnetInput := make([]types.Object, 0, len(destinationInput.SiteNetworkSubnet.Elements()))
						diags = append(diags, destinationInput.SiteNetworkSubnet.ElementsAs(ctx, &elementsDestinationSiteNetworkSubnetInput, false)...)

						var itemDestinationSiteNetworkSubnetInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_SiteNetworkSubnet
						for _, item := range elementsDestinationSiteNetworkSubnetInput {
							diags = append(diags, item.As(ctx, &itemDestinationSiteNetworkSubnetInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationSiteNetworkSubnetInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionInput.Destination.SiteNetworkSubnet = append(exceptionInput.Destination.SiteNetworkSubnet, &cato_models.SiteNetworkSubnetRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting destination floating subnet
					if !destinationInput.FloatingSubnet.IsNull() {
						elementsDestinationFloatingSubnetInput := make([]types.Object, 0, len(destinationInput.FloatingSubnet.Elements()))
						diags = append(diags, destinationInput.FloatingSubnet.ElementsAs(ctx, &elementsDestinationFloatingSubnetInput, false)...)

						var itemDestinationFloatingSubnetInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_FloatingSubnet
						for _, item := range elementsDestinationFloatingSubnetInput {
							diags = append(diags, item.As(ctx, &itemDestinationFloatingSubnetInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationFloatingSubnetInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionInput.Destination.FloatingSubnet = append(exceptionInput.Destination.FloatingSubnet, &cato_models.FloatingSubnetRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting destination user
					if !destinationInput.User.IsNull() {
						elementsDestinationUserInput := make([]types.Object, 0, len(destinationInput.User.Elements()))
						diags = append(diags, destinationInput.User.ElementsAs(ctx, &elementsDestinationUserInput, false)...)

						var itemDestinationUserInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_User
						for _, item := range elementsDestinationUserInput {
							diags = append(diags, item.As(ctx, &itemDestinationUserInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationUserInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionInput.Destination.User = append(exceptionInput.Destination.User, &cato_models.UserRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting destination users group
					if !destinationInput.UsersGroup.IsNull() {
						elementsDestinationUsersGroupInput := make([]types.Object, 0, len(destinationInput.UsersGroup.Elements()))
						diags = append(diags, destinationInput.UsersGroup.ElementsAs(ctx, &elementsDestinationUsersGroupInput, false)...)

						var itemDestinationUsersGroupInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_UsersGroup
						for _, item := range elementsDestinationUsersGroupInput {
							diags = append(diags, item.As(ctx, &itemDestinationUsersGroupInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationUsersGroupInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionInput.Destination.UsersGroup = append(exceptionInput.Destination.UsersGroup, &cato_models.UsersGroupRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting destination group
					if !destinationInput.Group.IsNull() {
						elementsDestinationGroupInput := make([]types.Object, 0, len(destinationInput.Group.Elements()))
						diags = append(diags, destinationInput.Group.ElementsAs(ctx, &elementsDestinationGroupInput, false)...)

						var itemDestinationGroupInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_Group
						for _, item := range elementsDestinationGroupInput {
							diags = append(diags, item.As(ctx, &itemDestinationGroupInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationGroupInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionInput.Destination.Group = append(exceptionInput.Destination.Group, &cato_models.GroupRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting destination system group
					if !destinationInput.SystemGroup.IsNull() {
						elementsDestinationSystemGroupInput := make([]types.Object, 0, len(destinationInput.SystemGroup.Elements()))
						diags = append(diags, destinationInput.SystemGroup.ElementsAs(ctx, &elementsDestinationSystemGroupInput, false)...)

						var itemDestinationSystemGroupInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_SystemGroup
						for _, item := range elementsDestinationSystemGroupInput {
							diags = append(diags, item.As(ctx, &itemDestinationSystemGroupInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationSystemGroupInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionInput.Destination.SystemGroup = append(exceptionInput.Destination.SystemGroup, &cato_models.SystemGroupRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}
				}

				// setting application
				if !itemExceptionsInput.Application.IsNull() {

					exceptionInput.Application = &cato_models.WanFirewallApplicationInput{}
					applicationInput := Policy_Policy_WanFirewall_Policy_Rules_Rule_Application{}
					diags = append(diags, itemExceptionsInput.Application.As(ctx, &applicationInput, basetypes.ObjectAsOptions{})...)

					// setting application IP
					if !applicationInput.IP.IsNull() {
						diags = append(diags, applicationInput.IP.ElementsAs(ctx, &exceptionInput.Application.IP, false)...)
					}

					// setting application subnet
					if !applicationInput.Subnet.IsNull() {
						diags = append(diags, applicationInput.Subnet.ElementsAs(ctx, &exceptionInput.Application.Subnet, false)...)
					}

					// setting application domain
					if !applicationInput.Domain.IsNull() {
						diags = append(diags, applicationInput.Domain.ElementsAs(ctx, &exceptionInput.Application.Domain, false)...)
					}

					// setting application fqdn
					if !applicationInput.Fqdn.IsNull() {
						diags = append(diags, applicationInput.Fqdn.ElementsAs(ctx, &exceptionInput.Application.Fqdn, false)...)
					}

					// setting application application
					if !applicationInput.Application.IsNull() {
						elementsApplicationApplicationInput := make([]types.Object, 0, len(applicationInput.Application.Elements()))
						diags = append(diags, applicationInput.Application.ElementsAs(ctx, &elementsApplicationApplicationInput, false)...)

						var itemApplicationApplicationInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_Application
						for _, item := range elementsApplicationApplicationInput {
							diags = append(diags, item.As(ctx, &itemApplicationApplicationInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationApplicationInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionInput.Application.Application = append(exceptionInput.Application.Application, &cato_models.ApplicationRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting application custom app
					if !applicationInput.CustomApp.IsNull() {
						elementsApplicationCustomAppInput := make([]types.Object, 0, len(applicationInput.CustomApp.Elements()))
						diags = append(diags, applicationInput.CustomApp.ElementsAs(ctx, &elementsApplicationCustomAppInput, false)...)

						var itemApplicationCustomAppInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_CustomApp
						for _, item := range elementsApplicationCustomAppInput {
							diags = append(diags, item.As(ctx, &itemApplicationCustomAppInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationCustomAppInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionInput.Application.CustomApp = append(exceptionInput.Application.CustomApp, &cato_models.CustomApplicationRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting application ip range
					if !applicationInput.IPRange.IsNull() {
						elementsApplicationIPRangeInput := make([]types.Object, 0, len(applicationInput.IPRange.Elements()))
						diags = append(diags, applicationInput.IPRange.ElementsAs(ctx, &elementsApplicationIPRangeInput, false)...)

						var itemApplicationIPRangeInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_IPRange
						for _, item := range elementsApplicationIPRangeInput {
							diags = append(diags, item.As(ctx, &itemApplicationIPRangeInput, basetypes.ObjectAsOptions{})...)

							exceptionInput.Application.IPRange = append(exceptionInput.Application.IPRange, &cato_models.IPAddressRangeInput{
								From: itemApplicationIPRangeInput.From.ValueString(),
								To:   itemApplicationIPRangeInput.To.ValueString(),
							})
						}
					}

					// setting application global ip range
					if !applicationInput.GlobalIPRange.IsNull() {
						elementsApplicationGlobalIPRangeInput := make([]types.Object, 0, len(applicationInput.GlobalIPRange.Elements()))
						diags = append(diags, applicationInput.GlobalIPRange.ElementsAs(ctx, &elementsApplicationGlobalIPRangeInput, false)...)

						var itemApplicationGlobalIPRangeInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_GlobalIPRange
						for _, item := range elementsApplicationGlobalIPRangeInput {
							diags = append(diags, item.As(ctx, &itemApplicationGlobalIPRangeInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationGlobalIPRangeInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionInput.Application.GlobalIPRange = append(exceptionInput.Application.GlobalIPRange, &cato_models.GlobalIPRangeRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting application app category
					if !applicationInput.AppCategory.IsNull() {
						elementsApplicationAppCategoryInput := make([]types.Object, 0, len(applicationInput.AppCategory.Elements()))
						diags = append(diags, applicationInput.AppCategory.ElementsAs(ctx, &elementsApplicationAppCategoryInput, false)...)

						var itemApplicationAppCategoryInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_AppCategory
						for _, item := range elementsApplicationAppCategoryInput {
							diags = append(diags, item.As(ctx, &itemApplicationAppCategoryInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationAppCategoryInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionInput.Application.AppCategory = append(exceptionInput.Application.AppCategory, &cato_models.ApplicationCategoryRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting application custom app category
					if !applicationInput.CustomCategory.IsNull() {
						elementsApplicationCustomCategoryInput := make([]types.Object, 0, len(applicationInput.CustomCategory.Elements()))
						diags = append(diags, applicationInput.CustomCategory.ElementsAs(ctx, &elementsApplicationCustomCategoryInput, false)...)

						var itemApplicationCustomCategoryInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_CustomCategory
						for _, item := range elementsApplicationCustomCategoryInput {
							diags = append(diags, item.As(ctx, &itemApplicationCustomCategoryInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationCustomCategoryInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionInput.Application.CustomCategory = append(exceptionInput.Application.CustomCategory, &cato_models.CustomCategoryRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting application sanctionned apps category
					if !applicationInput.SanctionedAppsCategory.IsNull() {
						elementsApplicationSanctionedAppsCategoryInput := make([]types.Object, 0, len(applicationInput.SanctionedAppsCategory.Elements()))
						diags = append(diags, applicationInput.SanctionedAppsCategory.ElementsAs(ctx, &elementsApplicationSanctionedAppsCategoryInput, false)...)

						var itemApplicationSanctionedAppsCategoryInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_SanctionedAppsCategory
						for _, item := range elementsApplicationSanctionedAppsCategoryInput {
							diags = append(diags, item.As(ctx, &itemApplicationSanctionedAppsCategoryInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationSanctionedAppsCategoryInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionInput.Application.SanctionedAppsCategory = append(exceptionInput.Application.SanctionedAppsCategory, &cato_models.SanctionedAppsCategoryRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}
				}

				// setting service
				if !itemExceptionsInput.Service.IsNull() {

					exceptionInput.Service = &cato_models.WanFirewallServiceTypeInput{}
					serviceInput := Policy_Policy_WanFirewall_Policy_Rules_Rule_Service{}
					diags = append(diags, itemExceptionsInput.Service.As(ctx, &serviceInput, basetypes.ObjectAsOptions{})...)

					// setting service standard
					if !serviceInput.Standard.IsNull() {
						elementsServiceStandardInput := make([]types.Object, 0, len(serviceInput.Standard.Elements()))
						diags = append(diags, serviceInput.Standard.ElementsAs(ctx, &elementsServiceStandardInput, false)...)

						var itemServiceStandardInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Service_Standard
						for _, item := range elementsServiceStandardInput {
							diags = append(diags, item.As(ctx, &itemServiceStandardInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemServiceStandardInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionInput.Service.Standard = append(exceptionInput.Service.Standard, &cato_models.ServiceRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting service custom
					if !serviceInput.Custom.IsNull() {
						elementsServiceCustomInput := make([]types.Object, 0, len(serviceInput.Custom.Elements()))
						diags = append(diags, serviceInput.Custom.ElementsAs(ctx, &elementsServiceCustomInput, false)...)

						var itemServiceCustomInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Service_Custom
						for _, item := range elementsServiceCustomInput {
							diags = append(diags, item.As(ctx, &itemServiceCustomInput, basetypes.ObjectAsOptions{})...)

							customInput := &cato_models.CustomServiceInput{
								Protocol: cato_models.IPProtocol(itemServiceCustomInput.Protocol.ValueString()),
							}

							// setting service custom port
							if !itemServiceCustomInput.Port.IsNull() {
								elementsPort := make([]types.String, 0, len(itemServiceCustomInput.Port.Elements()))
								diags = append(diags, itemServiceCustomInput.Port.ElementsAs(ctx, &elementsPort, false)...)

								inputPort := []cato_scalars.Port{}
								for _, item := range elementsPort {
									inputPort = append(inputPort, cato_scalars.Port(item.ValueString()))
								}

								customInput.Port = inputPort
							}

							// setting service custom port range
							if !itemServiceCustomInput.PortRange.IsNull() {
								var itemPortRange Policy_Policy_WanFirewall_Policy_Rules_Rule_Service_Custom_PortRange
								diags = append(diags, itemServiceCustomInput.PortRange.As(ctx, &itemPortRange, basetypes.ObjectAsOptions{})...)

								inputPortRange := cato_models.PortRangeInput{
									From: cato_scalars.Port(itemPortRange.From.ValueString()),
									To:   cato_scalars.Port(itemPortRange.To.ValueString()),
								}

								customInput.PortRange = &inputPortRange
							}

							// append custom service
							exceptionInput.Service.Custom = append(exceptionInput.Service.Custom, customInput)
						}
					}
				}

				rootAddRule.Exceptions = append(rootAddRule.Exceptions, &exceptionInput)
				rootUpdateRule.Exceptions = append(rootUpdateRule.Exceptions, &exceptionInput)
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

		rootAddRule.Action = cato_models.WanFirewallActionEnum(ruleInput.Action.ValueString())
		rootUpdateRule.Action = (*cato_models.WanFirewallActionEnum)(ruleInput.Action.ValueStringPointer())

		rootAddRule.Direction = cato_models.WanFirewallDirectionEnum(ruleInput.Direction.ValueString())
		rootUpdateRule.Direction = (*cato_models.WanFirewallDirectionEnum)(ruleInput.Direction.ValueStringPointer())

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
