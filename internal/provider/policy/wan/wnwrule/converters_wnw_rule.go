package wnwrule

import (
	"context"
	"fmt"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	cato_scalars "github.com/catonetworks/cato-go-sdk/scalars"
	attr "github.com/hashicorp/terraform-plugin-framework/attr"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	wan "github.com/catonetworks/terraform-provider-cato/internal/provider/policy/wan"
	convert "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	customservice "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/customservice"
	idname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	iprange "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/iprange"
	oldidname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/oldidname"
	policyrule "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
	portrange "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/portrange"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

// hydrateWanNetworkRuleAPI takes in the current state/plan along with context and returns the created
// diagnostic data as well as cato API data used to either create or update WAN Network entries
//
//nolint:gocyclo,funlen
func hydrateWanNetworkRuleAPI(ctx context.Context, plan Rule) (hydrateWanNetworkAPITypes, diag.Diagnostics) {
	diags := []diag.Diagnostic{}

	hydrateAPIReturn := hydrateWanNetworkAPITypes{}
	hydrateAPIReturn.create = cato_models.WanNetworkAddRuleInput{}
	hydrateAPIReturn.update = cato_models.WanNetworkUpdateRuleInput{}
	hydrateAPIReturn.create.At = &cato_models.PolicyRulePositionInput{}

	rootAddRule := &cato_models.WanNetworkAddRuleDataInput{}
	rootUpdateRule := &cato_models.WanNetworkUpdateRuleDataInput{}

	// setting at for creation only
	if !plan.At.IsNull() {
		positionInput := policyrule.Position{}
		diags = append(diags, plan.At.As(ctx, &positionInput, basetypes.ObjectAsOptions{})...)

		hydrateAPIReturn.create.At.Position = (*cato_models.PolicyRulePositionEnum)(positionInput.Position.ValueStringPointer())
		hydrateAPIReturn.create.At.Ref = positionInput.Ref.ValueStringPointer()
	}

	// setting rule
	if !plan.Rule.IsNull() {
		ruleInput := PolicyPolicyWanNetworkPolicyRulesRule{}
		diags = append(diags, plan.Rule.As(ctx, &ruleInput, basetypes.ObjectAsOptions{})...)

		// setting source
		if !ruleInput.Source.IsNull() {
			ruleSourceInput := &cato_models.WanNetworkRuleSourceInput{}
			ruleSourceUpdateInput := &cato_models.WanNetworkRuleSourceUpdateInput{}

			sourceInput := PolicyPolicyWanNetworkPolicyRulesRuleSource{}
			diags = append(diags, ruleInput.Source.As(ctx, &sourceInput, basetypes.ObjectAsOptions{})...)

			// setting source IP
			if !sourceInput.IP.IsUnknown() && !sourceInput.IP.IsNull() {
				diags = append(diags, sourceInput.IP.ElementsAs(ctx, &ruleSourceInput.IP, false)...)
				diags = append(diags, sourceInput.IP.ElementsAs(ctx, &ruleSourceUpdateInput.IP, false)...)
			} else {
				ruleSourceUpdateInput.IP = make([]string, 0)
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

				var itemSourceHostInput PolicyPolicyWanNetworkPolicyRulesRuleSourceHost
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

				var itemSourceSiteInput PolicyPolicyWanNetworkPolicyRulesRuleSourceSite
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

				var itemSourceIPRangeInput PolicyPolicyWanNetworkPolicyRulesRuleSourceIPRange
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

				var itemSourceGlobalIPRangeInput PolicyPolicyWanNetworkPolicyRulesRuleSourceGlobalIPRange
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

				var itemSourceNetworkInterfaceInput PolicyPolicyWanNetworkPolicyRulesRuleSourceNetworkInterface
				for _, item := range elementsSourceNetworkInterfaceInput {
					diags = append(diags, item.As(ctx, &itemSourceNetworkInterfaceInput, basetypes.ObjectAsOptions{})...)

					// Prefer ID if present; only fall back to name when safe (no slashes)
					switch {
					case !itemSourceNetworkInterfaceInput.ID.IsNull() && !itemSourceNetworkInterfaceInput.ID.IsUnknown():
						ruleSourceInput.NetworkInterface = append(ruleSourceInput.NetworkInterface, &cato_models.NetworkInterfaceRefInput{
							By:    cato_models.ObjectRefBy("ID"),
							Input: itemSourceNetworkInterfaceInput.ID.ValueString(),
						})
					case !itemSourceNetworkInterfaceInput.Name.IsNull() && !itemSourceNetworkInterfaceInput.Name.IsUnknown():
						nameVal := itemSourceNetworkInterfaceInput.Name.ValueString()
						ruleSourceInput.NetworkInterface = append(ruleSourceInput.NetworkInterface, &cato_models.NetworkInterfaceRefInput{
							By:    cato_models.ObjectRefBy("NAME"),
							Input: nameVal,
						})
					default:
						diags = append(diags, diag.NewErrorDiagnostic(
							"Missing network_interface selector",
							"Neither id nor name provided for a source.network_interface element.",
						))
					}
				}
				ruleSourceUpdateInput.NetworkInterface = ruleSourceInput.NetworkInterface
			} else {
				ruleSourceUpdateInput.NetworkInterface = make([]*cato_models.NetworkInterfaceRefInput, 0)
			}

			// setting source site network subnet
			if !sourceInput.SiteNetworkSubnet.IsNull() {
				elementsSourceSiteNetworkSubnetInput := make([]types.Object, 0, len(sourceInput.SiteNetworkSubnet.Elements()))
				diags = append(diags, sourceInput.SiteNetworkSubnet.ElementsAs(ctx, &elementsSourceSiteNetworkSubnetInput, false)...)

				var itemSourceSiteNetworkSubnetInput PolicyPolicyWanNetworkPolicyRulesRuleSourceSiteNetworkSubnet
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

				var itemSourceFloatingSubnetInput PolicyPolicyWanNetworkPolicyRulesRuleSourceFloatingSubnet
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

				var itemSourceUserInput PolicyPolicyWanNetworkPolicyRulesRuleSourceUser
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

				var itemSourceUsersGroupInput PolicyPolicyWanNetworkPolicyRulesRuleSourceUsersGroup
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

				var itemSourceGroupInput PolicyPolicyWanNetworkPolicyRulesRuleSourceGroup
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

				var itemSourceSystemGroupInput PolicyPolicyWanNetworkPolicyRulesRuleSourceSystemGroup
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
		} else {
			tflog.Warn(ctx, "TFLOG_SOURCE_WANNETWORK_IS_NULL")
		}

		// setting destination
		if !ruleInput.Destination.IsUnknown() && !ruleInput.Destination.IsNull() {
			ruleDestinationInput := &cato_models.WanNetworkRuleDestinationInput{}
			ruleDestinationUpdateInput := &cato_models.WanNetworkRuleDestinationUpdateInput{}

			destinationInput := PolicyPolicyWanNetworkPolicyRulesRuleDestination{}
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

			// setting destination host
			if !destinationInput.Host.IsNull() {
				elementsDestinationHostInput := make([]types.Object, 0, len(destinationInput.Host.Elements()))
				diags = append(diags, destinationInput.Host.ElementsAs(ctx, &elementsDestinationHostInput, false)...)

				var itemDestinationHostInput PolicyPolicyWanNetworkPolicyRulesRuleDestinationHost
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
			} else {
				ruleDestinationUpdateInput.Host = make([]*cato_models.HostRefInput, 0)
			}

			// setting destination site
			if !destinationInput.Site.IsNull() {
				elementsDestinationSiteInput := make([]types.Object, 0, len(destinationInput.Site.Elements()))
				diags = append(diags, destinationInput.Site.ElementsAs(ctx, &elementsDestinationSiteInput, false)...)

				var itemDestinationSiteInput PolicyPolicyWanNetworkPolicyRulesRuleDestinationSite
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
			} else {
				ruleDestinationUpdateInput.Site = make([]*cato_models.SiteRefInput, 0)
			}

			// setting destination ip range
			if !destinationInput.IPRange.IsNull() {
				elementsDestinationIPRangeInput := make([]types.Object, 0, len(destinationInput.IPRange.Elements()))
				diags = append(diags, destinationInput.IPRange.ElementsAs(ctx, &elementsDestinationIPRangeInput, false)...)

				var itemDestinationIPRangeInput PolicyPolicyWanNetworkPolicyRulesRuleDestinationIPRange
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

				var itemDestinationGlobalIPRangeInput PolicyPolicyWanNetworkPolicyRulesRuleDestinationGlobalIPRange
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

			// setting destination network interface
			if !destinationInput.NetworkInterface.IsNull() {
				elementsDestinationNetworkInterfaceInput := make([]types.Object, 0, len(destinationInput.NetworkInterface.Elements()))
				diags = append(diags, destinationInput.NetworkInterface.ElementsAs(ctx, &elementsDestinationNetworkInterfaceInput, false)...)

				var itemDestinationNetworkInterfaceInput PolicyPolicyWanNetworkPolicyRulesRuleDestinationNetworkInterface
				for _, item := range elementsDestinationNetworkInterfaceInput {
					diags = append(diags, item.As(ctx, &itemDestinationNetworkInterfaceInput, basetypes.ObjectAsOptions{})...)

					switch {
					case !itemDestinationNetworkInterfaceInput.ID.IsNull() && !itemDestinationNetworkInterfaceInput.ID.IsUnknown():
						ruleDestinationInput.NetworkInterface = append(ruleDestinationInput.NetworkInterface, &cato_models.NetworkInterfaceRefInput{
							By:    cato_models.ObjectRefBy("ID"),
							Input: itemDestinationNetworkInterfaceInput.ID.ValueString(),
						})
					case !itemDestinationNetworkInterfaceInput.Name.IsNull() && !itemDestinationNetworkInterfaceInput.Name.IsUnknown():
						nameVal := itemDestinationNetworkInterfaceInput.Name.ValueString()
						ruleDestinationInput.NetworkInterface = append(ruleDestinationInput.NetworkInterface, &cato_models.NetworkInterfaceRefInput{
							By:    cato_models.ObjectRefBy("NAME"),
							Input: nameVal,
						})
					default:
						diags = append(diags, diag.NewErrorDiagnostic(
							"Missing network_interface selector",
							"Neither id nor name provided for a destination.network_interface element.",
						))
					}
				}
				ruleDestinationUpdateInput.NetworkInterface = ruleDestinationInput.NetworkInterface
			} else {
				ruleDestinationUpdateInput.NetworkInterface = make([]*cato_models.NetworkInterfaceRefInput, 0)
			}

			// setting destination site network subnet
			if !destinationInput.SiteNetworkSubnet.IsNull() {
				elementsDestinationSiteNetworkSubnetInput := make([]types.Object, 0, len(destinationInput.SiteNetworkSubnet.Elements()))
				diags = append(diags, destinationInput.SiteNetworkSubnet.ElementsAs(ctx, &elementsDestinationSiteNetworkSubnetInput, false)...)

				var itemDestinationSiteNetworkSubnetInput PolicyPolicyWanNetworkPolicyRulesRuleDestinationSiteNetworkSubnet
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
			} else {
				ruleDestinationUpdateInput.SiteNetworkSubnet = make([]*cato_models.SiteNetworkSubnetRefInput, 0)
			}

			// setting destination floating subnet
			if !destinationInput.FloatingSubnet.IsNull() {
				elementsDestinationFloatingSubnetInput := make([]types.Object, 0, len(destinationInput.FloatingSubnet.Elements()))
				diags = append(diags, destinationInput.FloatingSubnet.ElementsAs(ctx, &elementsDestinationFloatingSubnetInput, false)...)

				var itemDestinationFloatingSubnetInput PolicyPolicyWanNetworkPolicyRulesRuleDestinationFloatingSubnet
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
			} else {
				ruleDestinationUpdateInput.FloatingSubnet = make([]*cato_models.FloatingSubnetRefInput, 0)
			}

			// setting destination user
			if !destinationInput.User.IsNull() {
				elementsDestinationUserInput := make([]types.Object, 0, len(destinationInput.User.Elements()))
				diags = append(diags, destinationInput.User.ElementsAs(ctx, &elementsDestinationUserInput, false)...)

				var itemDestinationUserInput PolicyPolicyWanNetworkPolicyRulesRuleDestinationUser
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
			} else {
				ruleDestinationUpdateInput.User = make([]*cato_models.UserRefInput, 0)
			}

			// setting destination users group
			if !destinationInput.UsersGroup.IsNull() {
				elementsDestinationUsersGroupInput := make([]types.Object, 0, len(destinationInput.UsersGroup.Elements()))
				diags = append(diags, destinationInput.UsersGroup.ElementsAs(ctx, &elementsDestinationUsersGroupInput, false)...)

				var itemDestinationUsersGroupInput PolicyPolicyWanNetworkPolicyRulesRuleDestinationUsersGroup
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
			} else {
				ruleDestinationUpdateInput.UsersGroup = make([]*cato_models.UsersGroupRefInput, 0)
			}

			// setting destination group
			if !destinationInput.Group.IsNull() {
				elementsDestinationGroupInput := make([]types.Object, 0, len(destinationInput.Group.Elements()))
				diags = append(diags, destinationInput.Group.ElementsAs(ctx, &elementsDestinationGroupInput, false)...)

				var itemDestinationGroupInput PolicyPolicyWanNetworkPolicyRulesRuleDestinationGroup
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
			} else {
				ruleDestinationUpdateInput.Group = make([]*cato_models.GroupRefInput, 0)
			}

			// setting destination system group
			if !destinationInput.SystemGroup.IsNull() {
				elementsDestinationSystemGroupInput := make([]types.Object, 0, len(destinationInput.SystemGroup.Elements()))
				diags = append(diags, destinationInput.SystemGroup.ElementsAs(ctx, &elementsDestinationSystemGroupInput, false)...)

				var itemDestinationSystemGroupInput PolicyPolicyWanNetworkPolicyRulesRuleDestinationSystemGroup
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
			} else {
				ruleDestinationUpdateInput.SystemGroup = make([]*cato_models.SystemGroupRefInput, 0)
			}

			rootAddRule.Destination = ruleDestinationInput
			rootUpdateRule.Destination = ruleDestinationUpdateInput
		}

		// setting application
		if !ruleInput.Application.IsUnknown() && !ruleInput.Application.IsNull() {
			ruleApplicationInput := &cato_models.WanNetworkRuleApplicationInput{}
			ruleApplicationUpdateInput := &cato_models.WanNetworkRuleApplicationUpdateInput{}

			applicationInput := PolicyPolicyWanNetworkPolicyRulesRuleApplication{}
			diags = append(diags, ruleInput.Application.As(ctx, &applicationInput, basetypes.ObjectAsOptions{})...)

			// setting application domain
			if !applicationInput.Domain.IsUnknown() && !applicationInput.Domain.IsNull() {
				diags = append(diags, applicationInput.Domain.ElementsAs(ctx, &ruleApplicationInput.Domain, false)...)
				diags = append(diags, applicationInput.Domain.ElementsAs(ctx, &ruleApplicationUpdateInput.Domain, false)...)
			} else {
				ruleApplicationUpdateInput.Domain = make([]string, 0)
			}

			// setting application fqdn
			if !applicationInput.Fqdn.IsUnknown() && !applicationInput.Fqdn.IsNull() {
				diags = append(diags, applicationInput.Fqdn.ElementsAs(ctx, &ruleApplicationInput.Fqdn, false)...)
				diags = append(diags, applicationInput.Fqdn.ElementsAs(ctx, &ruleApplicationUpdateInput.Fqdn, false)...)
			} else {
				ruleApplicationUpdateInput.Fqdn = make([]string, 0)
			}

			// setting application application
			if !applicationInput.Application.IsUnknown() && !applicationInput.Application.IsNull() {
				elementsApplicationApplicationInput := make([]types.Object, 0, len(applicationInput.Application.Elements()))
				diags = append(diags, applicationInput.Application.ElementsAs(ctx, &elementsApplicationApplicationInput, false)...)

				var itemApplicationApplicationInput PolicyPolicyWanNetworkPolicyRulesRuleApplicationApplication
				for _, item := range elementsApplicationApplicationInput {
					diags = append(diags, item.As(ctx, &itemApplicationApplicationInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationApplicationInput)
					if err != nil {
						tflog.Error(ctx, err.Error())
					}

					ruleApplicationInput.Application = append(ruleApplicationInput.Application, &cato_models.ApplicationRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
				ruleApplicationUpdateInput.Application = ruleApplicationInput.Application
			} else {
				ruleApplicationUpdateInput.Application = make([]*cato_models.ApplicationRefInput, 0)
			}

			// setting application custom app
			if !applicationInput.CustomApp.IsNull() {
				elementsApplicationCustomAppInput := make([]types.Object, 0, len(applicationInput.CustomApp.Elements()))
				diags = append(diags, applicationInput.CustomApp.ElementsAs(ctx, &elementsApplicationCustomAppInput, false)...)

				var itemApplicationCustomAppInput PolicyPolicyWanNetworkPolicyRulesRuleApplicationCustomApp
				for _, item := range elementsApplicationCustomAppInput {
					diags = append(diags, item.As(ctx, &itemApplicationCustomAppInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationCustomAppInput)
					if err != nil {
						tflog.Error(ctx, err.Error())
					}

					ruleApplicationInput.CustomApp = append(ruleApplicationInput.CustomApp, &cato_models.CustomApplicationRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
				ruleApplicationUpdateInput.CustomApp = ruleApplicationInput.CustomApp
			} else {
				ruleApplicationUpdateInput.CustomApp = make([]*cato_models.CustomApplicationRefInput, 0)
			}

			// setting application app category
			if !applicationInput.AppCategory.IsUnknown() && !applicationInput.AppCategory.IsNull() {
				elementsApplicationAppCategoryInput := make([]types.Object, 0, len(applicationInput.AppCategory.Elements()))
				diags = append(diags, applicationInput.AppCategory.ElementsAs(ctx, &elementsApplicationAppCategoryInput, false)...)

				var itemApplicationAppCategoryInput PolicyPolicyWanNetworkPolicyRulesRuleApplicationAppCategory
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
			} else {
				ruleApplicationUpdateInput.AppCategory = make([]*cato_models.ApplicationCategoryRefInput, 0)
			}

			// setting application custom category
			if !applicationInput.CustomCategory.IsNull() {
				elementsApplicationCustomCategoryInput := make([]types.Object, 0, len(applicationInput.CustomCategory.Elements()))
				diags = append(diags, applicationInput.CustomCategory.ElementsAs(ctx, &elementsApplicationCustomCategoryInput, false)...)

				var itemApplicationCustomCategoryInput PolicyPolicyWanNetworkPolicyRulesRuleApplicationCustomCategory
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
			} else {
				ruleApplicationUpdateInput.CustomCategory = make([]*cato_models.CustomCategoryRefInput, 0)
			}

			// setting application service
			if !applicationInput.Service.IsNull() {
				elementsApplicationServiceInput := make([]types.Object, 0, len(applicationInput.Service.Elements()))
				diags = append(diags, applicationInput.Service.ElementsAs(ctx, &elementsApplicationServiceInput, false)...)

				var itemApplicationServiceInput PolicyPolicyWanNetworkPolicyRulesRuleApplicationService
				for _, item := range elementsApplicationServiceInput {
					diags = append(diags, item.As(ctx, &itemApplicationServiceInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationServiceInput)
					if err != nil {
						tflog.Error(ctx, err.Error())
					}

					ruleApplicationInput.Service = append(ruleApplicationInput.Service, &cato_models.ServiceRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
				ruleApplicationUpdateInput.Service = ruleApplicationInput.Service
			} else {
				ruleApplicationUpdateInput.Service = make([]*cato_models.ServiceRefInput, 0)
			}

			// setting application custom service
			if !applicationInput.CustomService.IsNull() {
				elementsApplicationCustomServiceInput := make([]types.Object, 0, len(applicationInput.CustomService.Elements()))
				diags = append(diags, applicationInput.CustomService.ElementsAs(ctx, &elementsApplicationCustomServiceInput, false)...)

				var itemApplicationCustomServiceInput PolicyPolicyWanNetworkPolicyRulesRuleApplicationCustomService
				for _, item := range elementsApplicationCustomServiceInput {
					diags = append(diags, item.As(ctx, &itemApplicationCustomServiceInput, basetypes.ObjectAsOptions{})...)

					customServiceInput := &cato_models.CustomServiceInput{
						Protocol: cato_models.IPProtocol(itemApplicationCustomServiceInput.Protocol.ValueString()),
					}

					// setting custom service port
					if !itemApplicationCustomServiceInput.Port.IsNull() {
						elementsPort := make([]types.String, 0, len(itemApplicationCustomServiceInput.Port.Elements()))
						diags = append(diags, itemApplicationCustomServiceInput.Port.ElementsAs(ctx, &elementsPort, false)...)

						inputPort := []cato_scalars.Port{}
						for _, portItem := range elementsPort {
							inputPort = append(inputPort, cato_scalars.Port(portItem.ValueString()))
						}
						customServiceInput.Port = inputPort
					}

					// setting custom service port range
					if !itemApplicationCustomServiceInput.PortRange.IsNull() {
						var itemPortRange PolicyPolicyWanNetworkPolicyRulesRuleApplicationCustomServicePortRange
						diags = append(diags, itemApplicationCustomServiceInput.PortRange.As(ctx, &itemPortRange, basetypes.ObjectAsOptions{})...)

						inputPortRange := cato_models.PortRangeInput{
							From: cato_scalars.Port(itemPortRange.From.ValueString()),
							To:   cato_scalars.Port(itemPortRange.To.ValueString()),
						}
						customServiceInput.PortRange = &inputPortRange
					}

					ruleApplicationInput.CustomService = append(ruleApplicationInput.CustomService, customServiceInput)
				}
				ruleApplicationUpdateInput.CustomService = ruleApplicationInput.CustomService
			} else {
				ruleApplicationUpdateInput.CustomService = make([]*cato_models.CustomServiceInput, 0)
			}

			// setting application custom service IP
			if !applicationInput.CustomServiceIP.IsNull() {
				elementsApplicationCustomServiceIPInput := make([]types.Object, 0, len(applicationInput.CustomServiceIP.Elements()))
				diags = append(diags, applicationInput.CustomServiceIP.ElementsAs(ctx, &elementsApplicationCustomServiceIPInput, false)...)

				var itemApplicationCustomServiceIPInput PolicyPolicyWanNetworkPolicyRulesRuleApplicationCustomServiceIP
				for _, item := range elementsApplicationCustomServiceIPInput {
					diags = append(diags, item.As(ctx, &itemApplicationCustomServiceIPInput, basetypes.ObjectAsOptions{})...)

					customServiceIPInput := &cato_models.CustomServiceIPInput{
						Name: itemApplicationCustomServiceIPInput.Name.ValueString(),
					}

					// setting custom service IP address
					if !itemApplicationCustomServiceIPInput.IP.IsNull() {
						ipValue := itemApplicationCustomServiceIPInput.IP.ValueString()
						customServiceIPInput.IP = &ipValue
					}

					// setting custom service IP range
					if !itemApplicationCustomServiceIPInput.IPRange.IsNull() {
						var itemIPRange PolicyPolicyWanNetworkPolicyRulesRuleApplicationCustomServiceIPIPRange
						diags = append(diags, itemApplicationCustomServiceIPInput.IPRange.As(ctx, &itemIPRange, basetypes.ObjectAsOptions{})...)

						inputIPRange := cato_models.IPAddressRangeInput{
							From: itemIPRange.From.ValueString(),
							To:   itemIPRange.To.ValueString(),
						}
						customServiceIPInput.IPRange = &inputIPRange
					}

					ruleApplicationInput.CustomServiceIP = append(ruleApplicationInput.CustomServiceIP, customServiceIPInput)
				}
				ruleApplicationUpdateInput.CustomServiceIP = ruleApplicationInput.CustomServiceIP
			} else {
				ruleApplicationUpdateInput.CustomServiceIP = make([]*cato_models.CustomServiceIPInput, 0)
			}

			rootAddRule.Application = ruleApplicationInput
			rootUpdateRule.Application = ruleApplicationUpdateInput
		}

		// settings exceptions
		if !ruleInput.Exceptions.IsNull() && !ruleInput.Exceptions.IsUnknown() {
			elementsExceptionsInput := make([]types.Object, 0, len(ruleInput.Exceptions.Elements()))
			diags = append(diags, ruleInput.Exceptions.ElementsAs(ctx, &elementsExceptionsInput, false)...)

			// loop over exceptions
			var itemExceptionsInput PolicyPolicyWanNetworkPolicyRulesRuleExceptions
			for _, item := range elementsExceptionsInput {
				exceptionAddInput := cato_models.WanNetworkRuleExceptionInput{}
				exceptionUpdateInput := cato_models.WanNetworkRuleExceptionInput{}

				diags = append(diags, item.As(ctx, &itemExceptionsInput, basetypes.ObjectAsOptions{})...)

				// setting exception name
				exceptionAddInput.Name = itemExceptionsInput.Name.ValueString()
				exceptionUpdateInput.Name = itemExceptionsInput.Name.ValueString()

				// setting source
				if !itemExceptionsInput.Source.IsNull() {
					exceptionAddInput.Source = &cato_models.WanNetworkRuleSourceInput{}
					exceptionUpdateInput.Source = &cato_models.WanNetworkRuleSourceInput{}

					exceptionSourceInput := PolicyPolicyWanNetworkPolicyRulesRuleSource{}
					diags = append(diags, itemExceptionsInput.Source.As(ctx, &exceptionSourceInput, basetypes.ObjectAsOptions{})...)

					// setting source IP
					if !exceptionSourceInput.IP.IsNull() {
						diags = append(diags, exceptionSourceInput.IP.ElementsAs(ctx, &exceptionAddInput.Source.IP, false)...)
						exceptionUpdateInput.Source.IP = exceptionAddInput.Source.IP
					} else {
						exceptionUpdateInput.Source.IP = make([]string, 0)
					}

					// setting source subnet
					if !exceptionSourceInput.Subnet.IsNull() {
						diags = append(diags, exceptionSourceInput.Subnet.ElementsAs(ctx, &exceptionAddInput.Source.Subnet, false)...)
						exceptionUpdateInput.Source.Subnet = exceptionAddInput.Source.Subnet
					} else {
						exceptionUpdateInput.Source.Subnet = make([]string, 0)
					}

					// setting source host
					if !exceptionSourceInput.Host.IsNull() {
						elementsSourceHostInput := make([]types.Object, 0, len(exceptionSourceInput.Host.Elements()))
						diags = append(diags, exceptionSourceInput.Host.ElementsAs(ctx, &elementsSourceHostInput, false)...)

						var itemSourceHostInput PolicyPolicyWanNetworkPolicyRulesRuleSourceHost
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
					if !exceptionSourceInput.Site.IsNull() {
						elementsSourceSiteInput := make([]types.Object, 0, len(exceptionSourceInput.Site.Elements()))
						diags = append(diags, exceptionSourceInput.Site.ElementsAs(ctx, &elementsSourceSiteInput, false)...)

						var itemSourceSiteInput PolicyPolicyWanNetworkPolicyRulesRuleSourceSite
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
					if !exceptionSourceInput.IPRange.IsNull() {
						elementsExceptionSourceIPRangeInput := make([]types.Object, 0, len(exceptionSourceInput.IPRange.Elements()))
						diags = append(diags, exceptionSourceInput.IPRange.ElementsAs(ctx, &elementsExceptionSourceIPRangeInput, false)...)

						var itemSourceIPRangeInput PolicyPolicyWanNetworkPolicyRulesRuleSourceIPRange
						for _, item := range elementsExceptionSourceIPRangeInput {
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
					if !exceptionSourceInput.GlobalIPRange.IsNull() {
						elementsSourceGlobalIPRangeInput := make([]types.Object, 0, len(exceptionSourceInput.GlobalIPRange.Elements()))
						diags = append(diags, exceptionSourceInput.GlobalIPRange.ElementsAs(ctx, &elementsSourceGlobalIPRangeInput, false)...)

						var itemSourceGlobalIPRangeInput PolicyPolicyWanNetworkPolicyRulesRuleSourceGlobalIPRange
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
					if !exceptionSourceInput.NetworkInterface.IsNull() {
						elementsSourceNetworkInterfaceInput := make([]types.Object, 0, len(exceptionSourceInput.NetworkInterface.Elements()))
						diags = append(diags, exceptionSourceInput.NetworkInterface.ElementsAs(ctx, &elementsSourceNetworkInterfaceInput, false)...)

						var itemSourceNetworkInterfaceInput PolicyPolicyWanNetworkPolicyRulesRuleSourceNetworkInterface
						for _, item := range elementsSourceNetworkInterfaceInput {
							diags = append(diags, item.As(ctx, &itemSourceNetworkInterfaceInput, basetypes.ObjectAsOptions{})...)

							switch {
							case !itemSourceNetworkInterfaceInput.ID.IsNull() && !itemSourceNetworkInterfaceInput.ID.IsUnknown():
								exceptionAddInput.Source.NetworkInterface = append(exceptionAddInput.Source.NetworkInterface, &cato_models.NetworkInterfaceRefInput{
									By:    cato_models.ObjectRefBy("ID"),
									Input: itemSourceNetworkInterfaceInput.ID.ValueString(),
								})
							case !itemSourceNetworkInterfaceInput.Name.IsNull() && !itemSourceNetworkInterfaceInput.Name.IsUnknown():
								nameVal := itemSourceNetworkInterfaceInput.Name.ValueString()
								exceptionAddInput.Source.NetworkInterface = append(exceptionAddInput.Source.NetworkInterface, &cato_models.NetworkInterfaceRefInput{
									By:    cato_models.ObjectRefBy("NAME"),
									Input: nameVal,
								})
							default:
								diags = append(diags, diag.NewErrorDiagnostic(
									"Missing network_interface selector",
									"Neither id nor name provided for an exception.source.network_interface element.",
								))
							}
						}
						exceptionUpdateInput.Source.NetworkInterface = exceptionAddInput.Source.NetworkInterface
					} else {
						exceptionUpdateInput.Source.NetworkInterface = make([]*cato_models.NetworkInterfaceRefInput, 0)
					}

					// setting source site network subnet
					if !exceptionSourceInput.SiteNetworkSubnet.IsNull() {
						elementsSourceSiteNetworkSubnetInput := make([]types.Object, 0, len(exceptionSourceInput.SiteNetworkSubnet.Elements()))
						diags = append(diags, exceptionSourceInput.SiteNetworkSubnet.ElementsAs(ctx, &elementsSourceSiteNetworkSubnetInput, false)...)

						var itemSourceSiteNetworkSubnetInput PolicyPolicyWanNetworkPolicyRulesRuleSourceSiteNetworkSubnet
						for _, item := range elementsSourceSiteNetworkSubnetInput {
							diags = append(diags, item.As(ctx, &itemSourceSiteNetworkSubnetInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceSiteNetworkSubnetInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
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
					if !exceptionSourceInput.FloatingSubnet.IsNull() {
						elementsSourceFloatingSubnetInput := make([]types.Object, 0, len(exceptionSourceInput.FloatingSubnet.Elements()))
						diags = append(diags, exceptionSourceInput.FloatingSubnet.ElementsAs(ctx, &elementsSourceFloatingSubnetInput, false)...)

						var itemSourceFloatingSubnetInput PolicyPolicyWanNetworkPolicyRulesRuleSourceFloatingSubnet
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
					if !exceptionSourceInput.User.IsNull() {
						elementsSourceUserInput := make([]types.Object, 0, len(exceptionSourceInput.User.Elements()))
						diags = append(diags, exceptionSourceInput.User.ElementsAs(ctx, &elementsSourceUserInput, false)...)

						var itemSourceUserInput PolicyPolicyWanNetworkPolicyRulesRuleSourceUser
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
					if !exceptionSourceInput.UsersGroup.IsNull() {
						elementsSourceUsersGroupInput := make([]types.Object, 0, len(exceptionSourceInput.UsersGroup.Elements()))
						diags = append(diags, exceptionSourceInput.UsersGroup.ElementsAs(ctx, &elementsSourceUsersGroupInput, false)...)

						var itemSourceUsersGroupInput PolicyPolicyWanNetworkPolicyRulesRuleSourceUsersGroup
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
					if !exceptionSourceInput.Group.IsNull() {
						elementsSourceGroupInput := make([]types.Object, 0, len(exceptionSourceInput.Group.Elements()))
						diags = append(diags, exceptionSourceInput.Group.ElementsAs(ctx, &elementsSourceGroupInput, false)...)

						var itemSourceGroupInput PolicyPolicyWanNetworkPolicyRulesRuleSourceGroup
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
					if !exceptionSourceInput.SystemGroup.IsNull() {
						elementsSourceSystemGroupInput := make([]types.Object, 0, len(exceptionSourceInput.SystemGroup.Elements()))
						diags = append(diags, exceptionSourceInput.SystemGroup.ElementsAs(ctx, &elementsSourceSystemGroupInput, false)...)

						var itemSourceSystemGroupInput PolicyPolicyWanNetworkPolicyRulesRuleSourceSystemGroup
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

				// setting destination
				if !itemExceptionsInput.Destination.IsNull() {
					exceptionAddInput.Destination = &cato_models.WanNetworkRuleDestinationInput{}
					exceptionUpdateInput.Destination = &cato_models.WanNetworkRuleDestinationInput{}

					exceptionDestinationInput := PolicyPolicyWanNetworkPolicyRulesRuleDestination{}
					diags = append(diags, itemExceptionsInput.Destination.As(ctx, &exceptionDestinationInput, basetypes.ObjectAsOptions{})...)

					// setting destination IP
					if !exceptionDestinationInput.IP.IsUnknown() && !exceptionDestinationInput.IP.IsNull() {
						diags = append(diags, exceptionDestinationInput.IP.ElementsAs(ctx, &exceptionAddInput.Destination.IP, false)...)
						exceptionUpdateInput.Destination.IP = exceptionAddInput.Destination.IP
					} else {
						exceptionUpdateInput.Destination.IP = make([]string, 0)
					}

					// setting destination subnet
					if !exceptionDestinationInput.Subnet.IsUnknown() && !exceptionDestinationInput.Subnet.IsNull() {
						diags = append(diags, exceptionDestinationInput.Subnet.ElementsAs(ctx, &exceptionAddInput.Destination.Subnet, false)...)
						exceptionUpdateInput.Destination.Subnet = exceptionAddInput.Destination.Subnet
					} else {
						exceptionUpdateInput.Destination.Subnet = make([]string, 0)
					}

					// setting destination host
					if !exceptionDestinationInput.Host.IsNull() {
						elementsDestinationHostInput := make([]types.Object, 0, len(exceptionDestinationInput.Host.Elements()))
						diags = append(diags, exceptionDestinationInput.Host.ElementsAs(ctx, &elementsDestinationHostInput, false)...)

						var itemDestinationHostInput PolicyPolicyWanNetworkPolicyRulesRuleDestinationHost
						for _, item := range elementsDestinationHostInput {
							diags = append(diags, item.As(ctx, &itemDestinationHostInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationHostInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionAddInput.Destination.Host = append(exceptionAddInput.Destination.Host, &cato_models.HostRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
						exceptionUpdateInput.Destination.Host = exceptionAddInput.Destination.Host
					} else {
						exceptionUpdateInput.Destination.Host = make([]*cato_models.HostRefInput, 0)
					}

					// setting destination site
					if !exceptionDestinationInput.Site.IsNull() {
						elementsDestinationSiteInput := make([]types.Object, 0, len(exceptionDestinationInput.Site.Elements()))
						diags = append(diags, exceptionDestinationInput.Site.ElementsAs(ctx, &elementsDestinationSiteInput, false)...)

						var itemDestinationSiteInput PolicyPolicyWanNetworkPolicyRulesRuleDestinationSite
						for _, item := range elementsDestinationSiteInput {
							diags = append(diags, item.As(ctx, &itemDestinationSiteInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationSiteInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionAddInput.Destination.Site = append(exceptionAddInput.Destination.Site, &cato_models.SiteRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
						exceptionUpdateInput.Destination.Site = exceptionAddInput.Destination.Site
					} else {
						exceptionUpdateInput.Destination.Site = make([]*cato_models.SiteRefInput, 0)
					}

					// setting destination ip range
					if !exceptionDestinationInput.IPRange.IsNull() {
						elementsDestinationIPRangeInput := make([]types.Object, 0, len(exceptionDestinationInput.IPRange.Elements()))
						diags = append(diags, exceptionDestinationInput.IPRange.ElementsAs(ctx, &elementsDestinationIPRangeInput, false)...)

						var itemDestinationIPRangeInput PolicyPolicyWanNetworkPolicyRulesRuleDestinationIPRange
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
					if !exceptionDestinationInput.GlobalIPRange.IsNull() {
						elementsDestinationGlobalIPRangeInput := make([]types.Object, 0, len(exceptionDestinationInput.GlobalIPRange.Elements()))
						diags = append(diags, exceptionDestinationInput.GlobalIPRange.ElementsAs(ctx, &elementsDestinationGlobalIPRangeInput, false)...)

						var itemDestinationGlobalIPRangeInput PolicyPolicyWanNetworkPolicyRulesRuleDestinationGlobalIPRange
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

					// setting destination network interface
					if !exceptionDestinationInput.NetworkInterface.IsNull() {
						elementsDestinationNetworkInterfaceInput := make([]types.Object, 0, len(exceptionDestinationInput.NetworkInterface.Elements()))
						diags = append(diags, exceptionDestinationInput.NetworkInterface.ElementsAs(ctx, &elementsDestinationNetworkInterfaceInput, false)...)

						var itemDestinationNetworkInterfaceInput PolicyPolicyWanNetworkPolicyRulesRuleDestinationNetworkInterface
						for _, item := range elementsDestinationNetworkInterfaceInput {
							diags = append(diags, item.As(ctx, &itemDestinationNetworkInterfaceInput, basetypes.ObjectAsOptions{})...)

							switch {
							case !itemDestinationNetworkInterfaceInput.ID.IsNull() && !itemDestinationNetworkInterfaceInput.ID.IsUnknown():
								exceptionAddInput.Destination.NetworkInterface = append(
									exceptionAddInput.Destination.NetworkInterface,
									&cato_models.NetworkInterfaceRefInput{
										By:    cato_models.ObjectRefBy("ID"),
										Input: itemDestinationNetworkInterfaceInput.ID.ValueString(),
									},
								)
							case !itemDestinationNetworkInterfaceInput.Name.IsNull() && !itemDestinationNetworkInterfaceInput.Name.IsUnknown():
								nameVal := itemDestinationNetworkInterfaceInput.Name.ValueString()
								exceptionAddInput.Destination.NetworkInterface = append(
									exceptionAddInput.Destination.NetworkInterface,
									&cato_models.NetworkInterfaceRefInput{
										By:    cato_models.ObjectRefBy("NAME"),
										Input: nameVal,
									},
								)
							default:
								diags = append(diags, diag.NewErrorDiagnostic(
									"Missing network_interface selector",
									"Neither id nor name provided for an exception.destination.network_interface element.",
								))
							}
						}
						exceptionUpdateInput.Destination.NetworkInterface = exceptionAddInput.Destination.NetworkInterface
					} else {
						exceptionUpdateInput.Destination.NetworkInterface = make([]*cato_models.NetworkInterfaceRefInput, 0)
					}

					// setting destination site network subnet
					if !exceptionDestinationInput.SiteNetworkSubnet.IsNull() {
						elementsDestinationSiteNetworkSubnetInput := make([]types.Object, 0, len(exceptionDestinationInput.SiteNetworkSubnet.Elements()))
						diags = append(
							diags,
							exceptionDestinationInput.SiteNetworkSubnet.ElementsAs(
								ctx,
								&elementsDestinationSiteNetworkSubnetInput,
								false,
							)...,
						)

						var itemDestinationSiteNetworkSubnetInput PolicyPolicyWanNetworkPolicyRulesRuleDestinationSiteNetworkSubnet
						for _, item := range elementsDestinationSiteNetworkSubnetInput {
							diags = append(diags, item.As(ctx, &itemDestinationSiteNetworkSubnetInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationSiteNetworkSubnetInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionAddInput.Destination.SiteNetworkSubnet = append(
								exceptionAddInput.Destination.SiteNetworkSubnet,
								&cato_models.SiteNetworkSubnetRefInput{
									By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
									Input: ObjectRefOutput.Input,
								},
							)
						}
						exceptionUpdateInput.Destination.SiteNetworkSubnet = exceptionAddInput.Destination.SiteNetworkSubnet
					} else {
						exceptionUpdateInput.Destination.SiteNetworkSubnet = make([]*cato_models.SiteNetworkSubnetRefInput, 0)
					}

					// setting destination floating subnet
					if !exceptionDestinationInput.FloatingSubnet.IsNull() {
						elementsDestinationFloatingSubnetInput := make([]types.Object, 0, len(exceptionDestinationInput.FloatingSubnet.Elements()))
						diags = append(diags, exceptionDestinationInput.FloatingSubnet.ElementsAs(ctx, &elementsDestinationFloatingSubnetInput, false)...)

						var itemDestinationFloatingSubnetInput PolicyPolicyWanNetworkPolicyRulesRuleDestinationFloatingSubnet
						for _, item := range elementsDestinationFloatingSubnetInput {
							diags = append(diags, item.As(ctx, &itemDestinationFloatingSubnetInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationFloatingSubnetInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionAddInput.Destination.FloatingSubnet = append(
								exceptionAddInput.Destination.FloatingSubnet,
								&cato_models.FloatingSubnetRefInput{
									By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
									Input: ObjectRefOutput.Input,
								},
							)
						}
						exceptionUpdateInput.Destination.FloatingSubnet = exceptionAddInput.Destination.FloatingSubnet
					} else {
						exceptionUpdateInput.Destination.FloatingSubnet = make([]*cato_models.FloatingSubnetRefInput, 0)
					}

					// setting destination user
					if !exceptionDestinationInput.User.IsNull() {
						elementsDestinationUserInput := make([]types.Object, 0, len(exceptionDestinationInput.User.Elements()))
						diags = append(diags, exceptionDestinationInput.User.ElementsAs(ctx, &elementsDestinationUserInput, false)...)

						var itemDestinationUserInput PolicyPolicyWanNetworkPolicyRulesRuleDestinationUser
						for _, item := range elementsDestinationUserInput {
							diags = append(diags, item.As(ctx, &itemDestinationUserInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationUserInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionAddInput.Destination.User = append(exceptionAddInput.Destination.User, &cato_models.UserRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
						exceptionUpdateInput.Destination.User = exceptionAddInput.Destination.User
					} else {
						exceptionUpdateInput.Destination.User = make([]*cato_models.UserRefInput, 0)
					}

					// setting destination users group
					if !exceptionDestinationInput.UsersGroup.IsNull() {
						elementsDestinationUsersGroupInput := make([]types.Object, 0, len(exceptionDestinationInput.UsersGroup.Elements()))
						diags = append(diags, exceptionDestinationInput.UsersGroup.ElementsAs(ctx, &elementsDestinationUsersGroupInput, false)...)

						var itemDestinationUsersGroupInput PolicyPolicyWanNetworkPolicyRulesRuleDestinationUsersGroup
						for _, item := range elementsDestinationUsersGroupInput {
							diags = append(diags, item.As(ctx, &itemDestinationUsersGroupInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationUsersGroupInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionAddInput.Destination.UsersGroup = append(exceptionAddInput.Destination.UsersGroup, &cato_models.UsersGroupRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
						exceptionUpdateInput.Destination.UsersGroup = exceptionAddInput.Destination.UsersGroup
					} else {
						exceptionUpdateInput.Destination.UsersGroup = make([]*cato_models.UsersGroupRefInput, 0)
					}

					// setting destination group
					if !exceptionDestinationInput.Group.IsNull() {
						elementsDestinationGroupInput := make([]types.Object, 0, len(exceptionDestinationInput.Group.Elements()))
						diags = append(diags, exceptionDestinationInput.Group.ElementsAs(ctx, &elementsDestinationGroupInput, false)...)

						var itemDestinationGroupInput PolicyPolicyWanNetworkPolicyRulesRuleDestinationGroup
						for _, item := range elementsDestinationGroupInput {
							diags = append(diags, item.As(ctx, &itemDestinationGroupInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationGroupInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionAddInput.Destination.Group = append(exceptionAddInput.Destination.Group, &cato_models.GroupRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
						exceptionUpdateInput.Destination.Group = exceptionAddInput.Destination.Group
					} else {
						exceptionUpdateInput.Destination.Group = make([]*cato_models.GroupRefInput, 0)
					}

					// setting destination system group
					if !exceptionDestinationInput.SystemGroup.IsNull() {
						elementsDestinationSystemGroupInput := make([]types.Object, 0, len(exceptionDestinationInput.SystemGroup.Elements()))
						diags = append(diags, exceptionDestinationInput.SystemGroup.ElementsAs(ctx, &elementsDestinationSystemGroupInput, false)...)

						var itemDestinationSystemGroupInput PolicyPolicyWanNetworkPolicyRulesRuleDestinationSystemGroup
						for _, item := range elementsDestinationSystemGroupInput {
							diags = append(diags, item.As(ctx, &itemDestinationSystemGroupInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationSystemGroupInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionAddInput.Destination.SystemGroup = append(exceptionAddInput.Destination.SystemGroup, &cato_models.SystemGroupRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
						exceptionUpdateInput.Destination.SystemGroup = exceptionAddInput.Destination.SystemGroup
					} else {
						exceptionUpdateInput.Destination.SystemGroup = make([]*cato_models.SystemGroupRefInput, 0)
					}
				}

				// setting application
				if !itemExceptionsInput.Application.IsNull() {
					exceptionAddInput.Application = &cato_models.WanNetworkRuleApplicationInput{}
					exceptionUpdateInput.Application = &cato_models.WanNetworkRuleApplicationInput{}

					exceptionApplicationInput := PolicyPolicyWanNetworkPolicyRulesRuleApplication{}
					diags = append(diags, itemExceptionsInput.Application.As(ctx, &exceptionApplicationInput, basetypes.ObjectAsOptions{})...)

					// setting application domain
					if !exceptionApplicationInput.Domain.IsNull() {
						diags = append(diags, exceptionApplicationInput.Domain.ElementsAs(ctx, &exceptionAddInput.Application.Domain, false)...)
						exceptionUpdateInput.Application.Domain = exceptionAddInput.Application.Domain
					} else {
						exceptionUpdateInput.Application.Domain = make([]string, 0)
					}

					// setting application fqdn
					if !exceptionApplicationInput.Fqdn.IsNull() {
						diags = append(diags, exceptionApplicationInput.Fqdn.ElementsAs(ctx, &exceptionAddInput.Application.Fqdn, false)...)
						exceptionUpdateInput.Application.Fqdn = exceptionAddInput.Application.Fqdn
					} else {
						exceptionUpdateInput.Application.Fqdn = make([]string, 0)
					}

					// setting application application
					if !exceptionApplicationInput.Application.IsNull() {
						elementsApplicationApplicationInput := make([]types.Object, 0, len(exceptionApplicationInput.Application.Elements()))
						diags = append(diags, exceptionApplicationInput.Application.ElementsAs(ctx, &elementsApplicationApplicationInput, false)...)

						var itemApplicationApplicationInput PolicyPolicyWanNetworkPolicyRulesRuleApplicationApplication
						for _, item := range elementsApplicationApplicationInput {
							diags = append(diags, item.As(ctx, &itemApplicationApplicationInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationApplicationInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionAddInput.Application.Application = append(exceptionAddInput.Application.Application, &cato_models.ApplicationRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
						exceptionUpdateInput.Application.Application = exceptionAddInput.Application.Application
					} else {
						exceptionUpdateInput.Application.Application = make([]*cato_models.ApplicationRefInput, 0)
					}

					// setting application custom app
					if !exceptionApplicationInput.CustomApp.IsNull() {
						elementsApplicationCustomAppInput := make([]types.Object, 0, len(exceptionApplicationInput.CustomApp.Elements()))
						diags = append(diags, exceptionApplicationInput.CustomApp.ElementsAs(ctx, &elementsApplicationCustomAppInput, false)...)

						var itemApplicationCustomAppInput PolicyPolicyWanNetworkPolicyRulesRuleApplicationCustomApp
						for _, item := range elementsApplicationCustomAppInput {
							diags = append(diags, item.As(ctx, &itemApplicationCustomAppInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationCustomAppInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionAddInput.Application.CustomApp = append(exceptionAddInput.Application.CustomApp, &cato_models.CustomApplicationRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
						exceptionUpdateInput.Application.CustomApp = exceptionAddInput.Application.CustomApp
					} else {
						exceptionUpdateInput.Application.CustomApp = make([]*cato_models.CustomApplicationRefInput, 0)
					}

					// setting application app category
					if !exceptionApplicationInput.AppCategory.IsNull() {
						elementsApplicationAppCategoryInput := make([]types.Object, 0, len(exceptionApplicationInput.AppCategory.Elements()))
						diags = append(diags, exceptionApplicationInput.AppCategory.ElementsAs(ctx, &elementsApplicationAppCategoryInput, false)...)

						var itemApplicationAppCategoryInput PolicyPolicyWanNetworkPolicyRulesRuleApplicationAppCategory
						for _, item := range elementsApplicationAppCategoryInput {
							diags = append(diags, item.As(ctx, &itemApplicationAppCategoryInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationAppCategoryInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionAddInput.Application.AppCategory = append(
								exceptionAddInput.Application.AppCategory,
								&cato_models.ApplicationCategoryRefInput{
									By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
									Input: ObjectRefOutput.Input,
								},
							)
						}
						exceptionUpdateInput.Application.AppCategory = exceptionAddInput.Application.AppCategory
					} else {
						exceptionUpdateInput.Application.AppCategory = make([]*cato_models.ApplicationCategoryRefInput, 0)
					}

					// setting application custom category
					if !exceptionApplicationInput.CustomCategory.IsNull() {
						elementsApplicationCustomCategoryInput := make([]types.Object, 0, len(exceptionApplicationInput.CustomCategory.Elements()))
						diags = append(diags, exceptionApplicationInput.CustomCategory.ElementsAs(ctx, &elementsApplicationCustomCategoryInput, false)...)

						var itemApplicationCustomCategoryInput PolicyPolicyWanNetworkPolicyRulesRuleApplicationCustomCategory
						for _, item := range elementsApplicationCustomCategoryInput {
							diags = append(diags, item.As(ctx, &itemApplicationCustomCategoryInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationCustomCategoryInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionAddInput.Application.CustomCategory = append(
								exceptionAddInput.Application.CustomCategory,
								&cato_models.CustomCategoryRefInput{
									By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
									Input: ObjectRefOutput.Input,
								},
							)
						}
						exceptionUpdateInput.Application.CustomCategory = exceptionAddInput.Application.CustomCategory
					} else {
						exceptionUpdateInput.Application.CustomCategory = make([]*cato_models.CustomCategoryRefInput, 0)
					}

					// setting application service
					if !exceptionApplicationInput.Service.IsNull() {
						elementsApplicationServiceInput := make([]types.Object, 0, len(exceptionApplicationInput.Service.Elements()))
						diags = append(diags, exceptionApplicationInput.Service.ElementsAs(ctx, &elementsApplicationServiceInput, false)...)

						var itemApplicationServiceInput PolicyPolicyWanNetworkPolicyRulesRuleApplicationService
						for _, item := range elementsApplicationServiceInput {
							diags = append(diags, item.As(ctx, &itemApplicationServiceInput, basetypes.ObjectAsOptions{})...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationServiceInput)
							if err != nil {
								tflog.Error(ctx, err.Error())
							}

							exceptionAddInput.Application.Service = append(exceptionAddInput.Application.Service, &cato_models.ServiceRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
						exceptionUpdateInput.Application.Service = exceptionAddInput.Application.Service
					} else {
						exceptionUpdateInput.Application.Service = make([]*cato_models.ServiceRefInput, 0)
					}

					// setting application custom service
					if !exceptionApplicationInput.CustomService.IsNull() {
						elementsApplicationCustomServiceInput := make([]types.Object, 0, len(exceptionApplicationInput.CustomService.Elements()))
						diags = append(diags, exceptionApplicationInput.CustomService.ElementsAs(ctx, &elementsApplicationCustomServiceInput, false)...)

						var itemApplicationCustomServiceInput PolicyPolicyWanNetworkPolicyRulesRuleApplicationCustomService
						for _, item := range elementsApplicationCustomServiceInput {
							diags = append(diags, item.As(ctx, &itemApplicationCustomServiceInput, basetypes.ObjectAsOptions{})...)

							customServiceInput := &cato_models.CustomServiceInput{
								Protocol: cato_models.IPProtocol(itemApplicationCustomServiceInput.Protocol.ValueString()),
							}

							// setting custom service port
							if !itemApplicationCustomServiceInput.Port.IsNull() {
								elementsPort := make([]types.String, 0, len(itemApplicationCustomServiceInput.Port.Elements()))
								diags = append(diags, itemApplicationCustomServiceInput.Port.ElementsAs(ctx, &elementsPort, false)...)

								inputPort := []cato_scalars.Port{}
								for _, portItem := range elementsPort {
									inputPort = append(inputPort, cato_scalars.Port(portItem.ValueString()))
								}
								customServiceInput.Port = inputPort
							}

							// setting custom service port range
							if !itemApplicationCustomServiceInput.PortRange.IsNull() {
								var itemPortRange PolicyPolicyWanNetworkPolicyRulesRuleApplicationCustomServicePortRange
								diags = append(diags, itemApplicationCustomServiceInput.PortRange.As(ctx, &itemPortRange, basetypes.ObjectAsOptions{})...)

								inputPortRange := cato_models.PortRangeInput{
									From: cato_scalars.Port(itemPortRange.From.ValueString()),
									To:   cato_scalars.Port(itemPortRange.To.ValueString()),
								}
								customServiceInput.PortRange = &inputPortRange
							}

							exceptionAddInput.Application.CustomService = append(exceptionAddInput.Application.CustomService, customServiceInput)
						}
						exceptionUpdateInput.Application.CustomService = exceptionAddInput.Application.CustomService
					} else {
						exceptionUpdateInput.Application.CustomService = make([]*cato_models.CustomServiceInput, 0)
					}

					// setting application custom service IP
					if !exceptionApplicationInput.CustomServiceIP.IsNull() {
						elementsApplicationCustomServiceIPInput := make([]types.Object, 0, len(exceptionApplicationInput.CustomServiceIP.Elements()))
						diags = append(diags, exceptionApplicationInput.CustomServiceIP.ElementsAs(ctx, &elementsApplicationCustomServiceIPInput, false)...)

						var itemApplicationCustomServiceIPInput PolicyPolicyWanNetworkPolicyRulesRuleApplicationCustomServiceIP
						for _, item := range elementsApplicationCustomServiceIPInput {
							diags = append(diags, item.As(ctx, &itemApplicationCustomServiceIPInput, basetypes.ObjectAsOptions{})...)

							customServiceIPInput := &cato_models.CustomServiceIPInput{
								Name: itemApplicationCustomServiceIPInput.Name.ValueString(),
							}

							// setting custom service IP address
							if !itemApplicationCustomServiceIPInput.IP.IsNull() {
								ipValue := itemApplicationCustomServiceIPInput.IP.ValueString()
								customServiceIPInput.IP = &ipValue
							}

							// setting custom service IP range
							if !itemApplicationCustomServiceIPInput.IPRange.IsNull() {
								var itemIPRange PolicyPolicyWanNetworkPolicyRulesRuleApplicationCustomServiceIPIPRange
								diags = append(diags, itemApplicationCustomServiceIPInput.IPRange.As(ctx, &itemIPRange, basetypes.ObjectAsOptions{})...)

								inputIPRange := cato_models.IPAddressRangeInput{
									From: itemIPRange.From.ValueString(),
									To:   itemIPRange.To.ValueString(),
								}
								customServiceIPInput.IPRange = &inputIPRange
							}

							exceptionAddInput.Application.CustomServiceIP = append(exceptionAddInput.Application.CustomServiceIP, customServiceIPInput)
						}
						exceptionUpdateInput.Application.CustomServiceIP = exceptionAddInput.Application.CustomServiceIP
					} else {
						exceptionUpdateInput.Application.CustomServiceIP = make([]*cato_models.CustomServiceIPInput, 0)
					}
				}

				rootAddRule.Exceptions = append(rootAddRule.Exceptions, &exceptionAddInput)
				rootUpdateRule.Exceptions = append(rootUpdateRule.Exceptions, &exceptionUpdateInput)
			}
		} else {
			rootUpdateRule.Exceptions = make([]*cato_models.WanNetworkRuleExceptionInput, 0)
		}

		// setting configuration
		if !ruleInput.Configuration.IsUnknown() && !ruleInput.Configuration.IsNull() {
			ruleConfigurationInput := &cato_models.WanNetworkRuleConfigurationInput{}
			ruleConfigurationUpdateInput := &cato_models.WanNetworkRuleConfigurationUpdateInput{}

			configurationInput := PolicyPolicyWanNetworkPolicyRulesRuleConfiguration{}
			diags = append(diags, ruleInput.Configuration.As(ctx, &configurationInput, basetypes.ObjectAsOptions{})...)

			ruleConfigurationInput.ActiveTCPAcceleration = configurationInput.ActiveTCPAcceleration
			ruleConfigurationInput.PacketLossMitigation = configurationInput.PacketLossMitigation
			ruleConfigurationInput.PreserveSourcePort = configurationInput.PreserveSourcePort

			ruleConfigurationUpdateInput.ActiveTCPAcceleration = &configurationInput.ActiveTCPAcceleration
			ruleConfigurationUpdateInput.PacketLossMitigation = &configurationInput.PacketLossMitigation
			ruleConfigurationUpdateInput.PreserveSourcePort = &configurationInput.PreserveSourcePort

			// setting primary transport
			if !configurationInput.PrimaryTransport.IsNull() {
				var primaryTransportInput PolicyPolicyWanNetworkPolicyRulesRuleConfigurationTransport
				diags = append(diags, configurationInput.PrimaryTransport.As(ctx, &primaryTransportInput, basetypes.ObjectAsOptions{})...)

				ruleConfigurationInput.PrimaryTransport = &cato_models.WanNetworkRuleTransportInput{
					TransportType:          cato_models.WanNetworkRuleTransportType(primaryTransportInput.TransportType.ValueString()),
					PrimaryInterfaceRole:   cato_models.WanNetworkRuleInterfaceRole(primaryTransportInput.PrimaryInterfaceRole.ValueString()),
					SecondaryInterfaceRole: cato_models.WanNetworkRuleInterfaceRole(primaryTransportInput.SecondaryInterfaceRole.ValueString()),
				}

				ruleConfigurationUpdateInput.PrimaryTransport = &cato_models.WanNetworkRuleTransportUpdateInput{
					TransportType:          (*cato_models.WanNetworkRuleTransportType)(primaryTransportInput.TransportType.ValueStringPointer()),
					PrimaryInterfaceRole:   (*cato_models.WanNetworkRuleInterfaceRole)(primaryTransportInput.PrimaryInterfaceRole.ValueStringPointer()),
					SecondaryInterfaceRole: (*cato_models.WanNetworkRuleInterfaceRole)(primaryTransportInput.SecondaryInterfaceRole.ValueStringPointer()),
				}
			}

			// setting secondary transport
			if !configurationInput.SecondaryTransport.IsNull() {
				var secondaryTransportInput PolicyPolicyWanNetworkPolicyRulesRuleConfigurationTransport
				diags = append(diags, configurationInput.SecondaryTransport.As(ctx, &secondaryTransportInput, basetypes.ObjectAsOptions{})...)

				ruleConfigurationInput.SecondaryTransport = &cato_models.WanNetworkRuleTransportInput{
					TransportType:        cato_models.WanNetworkRuleTransportType(secondaryTransportInput.TransportType.ValueString()),
					PrimaryInterfaceRole: cato_models.WanNetworkRuleInterfaceRole(secondaryTransportInput.PrimaryInterfaceRole.ValueString()),
					SecondaryInterfaceRole: cato_models.WanNetworkRuleInterfaceRole(
						secondaryTransportInput.SecondaryInterfaceRole.ValueString(),
					),
				}

				ruleConfigurationUpdateInput.SecondaryTransport = &cato_models.WanNetworkRuleTransportUpdateInput{
					TransportType:        (*cato_models.WanNetworkRuleTransportType)(secondaryTransportInput.TransportType.ValueStringPointer()),
					PrimaryInterfaceRole: (*cato_models.WanNetworkRuleInterfaceRole)(secondaryTransportInput.PrimaryInterfaceRole.ValueStringPointer()),
					SecondaryInterfaceRole: (*cato_models.WanNetworkRuleInterfaceRole)(
						secondaryTransportInput.SecondaryInterfaceRole.ValueStringPointer(),
					),
				}
			}

			// setting allocation IP
			if !configurationInput.AllocationIP.IsNull() {
				elementsAllocationIPInput := make([]types.Object, 0, len(configurationInput.AllocationIP.Elements()))
				diags = append(diags, configurationInput.AllocationIP.ElementsAs(ctx, &elementsAllocationIPInput, false)...)

				var itemAllocationIPInput PolicyPolicyWanNetworkPolicyRulesRuleConfigurationAllocationIP
				for _, item := range elementsAllocationIPInput {
					diags = append(diags, item.As(ctx, &itemAllocationIPInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemAllocationIPInput)
					if err != nil {
						tflog.Error(ctx, err.Error())
					}

					ruleConfigurationInput.AllocationIP = append(ruleConfigurationInput.AllocationIP, &cato_models.AllocatedIPRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
				ruleConfigurationUpdateInput.AllocationIP = ruleConfigurationInput.AllocationIP
			} else {
				ruleConfigurationUpdateInput.AllocationIP = make([]*cato_models.AllocatedIPRefInput, 0)
			}

			// setting pop location
			if !configurationInput.PopLocation.IsNull() {
				elementsPopLocationInput := make([]types.Object, 0, len(configurationInput.PopLocation.Elements()))
				diags = append(diags, configurationInput.PopLocation.ElementsAs(ctx, &elementsPopLocationInput, false)...)

				var itemPopLocationInput PolicyPolicyWanNetworkPolicyRulesRuleConfigurationPopLocation
				for _, item := range elementsPopLocationInput {
					diags = append(diags, item.As(ctx, &itemPopLocationInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemPopLocationInput)
					if err != nil {
						tflog.Error(ctx, err.Error())
					}

					ruleConfigurationInput.PopLocation = append(ruleConfigurationInput.PopLocation, &cato_models.PopLocationRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
				ruleConfigurationUpdateInput.PopLocation = ruleConfigurationInput.PopLocation
			} else {
				ruleConfigurationUpdateInput.PopLocation = make([]*cato_models.PopLocationRefInput, 0)
			}

			// setting backhauling site
			if !configurationInput.BackhaulingSite.IsNull() {
				elementsBackhaulingSiteInput := make([]types.Object, 0, len(configurationInput.BackhaulingSite.Elements()))
				diags = append(diags, configurationInput.BackhaulingSite.ElementsAs(ctx, &elementsBackhaulingSiteInput, false)...)

				var itemBackhaulingSiteInput PolicyPolicyWanNetworkPolicyRulesRuleConfigurationBackhaulingSite
				for _, item := range elementsBackhaulingSiteInput {
					diags = append(diags, item.As(ctx, &itemBackhaulingSiteInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemBackhaulingSiteInput)
					if err != nil {
						tflog.Error(ctx, err.Error())
					}

					ruleConfigurationInput.BackhaulingSite = append(ruleConfigurationInput.BackhaulingSite, &cato_models.SiteRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
				ruleConfigurationUpdateInput.BackhaulingSite = ruleConfigurationInput.BackhaulingSite
			} else {
				ruleConfigurationUpdateInput.BackhaulingSite = make([]*cato_models.SiteRefInput, 0)
			}

			rootAddRule.Configuration = ruleConfigurationInput
			rootUpdateRule.Configuration = ruleConfigurationUpdateInput
		}

		// setting bandwidth priority
		if !ruleInput.BandwidthPriority.IsUnknown() && !ruleInput.BandwidthPriority.IsNull() {
			var bandwidthPriorityInput PolicyPolicyWanNetworkPolicyRulesRuleBandwidthPriority
			diags = append(diags, ruleInput.BandwidthPriority.As(ctx, &bandwidthPriorityInput, basetypes.ObjectAsOptions{})...)

			ObjectRefOutput, err := utils.TransformObjectRefInput(bandwidthPriorityInput)
			if err != nil {
				tflog.Error(ctx, err.Error())
			}

			// Special case: if name equals "255", use id="-1" instead due to bug in API not accepting name properly for the default priority
			if ObjectRefOutput.By == "NAME" && ObjectRefOutput.Input == "255" {
				rootAddRule.BandwidthPriority = &cato_models.BandwidthManagementRefInput{
					By:    cato_models.ObjectRefBy("ID"),
					Input: "-1",
				}
			} else {
				rootAddRule.BandwidthPriority = &cato_models.BandwidthManagementRefInput{
					By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
					Input: ObjectRefOutput.Input,
				}
			}
			rootUpdateRule.BandwidthPriority = rootAddRule.BandwidthPriority
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

		rootAddRule.RuleType = cato_models.WanNetworkRuleType(ruleInput.RuleType.ValueString())
		rootUpdateRule.RuleType = (*cato_models.WanNetworkRuleType)(ruleInput.RuleType.ValueStringPointer())

		rootAddRule.RouteType = cato_models.WanNetworkRuleRouteType(ruleInput.RouteType.ValueString())
		rootUpdateRule.RouteType = (*cato_models.WanNetworkRuleRouteType)(ruleInput.RouteType.ValueStringPointer())
	}

	hydrateAPIReturn.create.Rule = rootAddRule
	hydrateAPIReturn.update.Rule = rootUpdateRule

	return hydrateAPIReturn, diags
}

//nolint:funlen
func hydrateWanNetworkRuleState(
	ctx context.Context,
	state Rule,
	currentRule *cato_go_sdk.WanNetworkPolicy_Policy_WanNetwork_Policy_Rules_Rule,
) (PolicyPolicyWanNetworkPolicyRulesRule, diag.Diagnostics) {
	ruleInput := PolicyPolicyWanNetworkPolicyRulesRule{}
	diags := make(diag.Diagnostics, 0)

	// Handle case where state might be incomplete (e.g., during import)
	if !state.Rule.IsNull() && !state.Rule.IsUnknown() {
		diagstmp := state.Rule.As(ctx, &ruleInput, basetypes.ObjectAsOptions{})
		if diagstmp.HasError() {
			tflog.Debug(ctx, "Failed to convert state to struct in hydrateWanNetworkRuleState, using empty struct for hydration")
			ruleInput = PolicyPolicyWanNetworkPolicyRulesRule{}
		} else {
			diags = append(diags, diagstmp...)
		}
	}

	// Basic rule fields
	ruleInput.Name = types.StringValue(currentRule.Name)
	if currentRule.Description == "" {
		ruleInput.Description = types.StringNull()
	} else {
		ruleInput.Description = types.StringValue(currentRule.Description)
	}
	ruleInput.ID = types.StringValue(currentRule.ID)
	ruleInput.Enabled = types.BoolValue(currentRule.Enabled)
	ruleInput.RuleType = types.StringValue(currentRule.RuleType.String())
	ruleInput.RouteType = types.StringValue(currentRule.RouteType.String())

	// /// /// /// / Rule -> Source // /// /// /// /// /
	curRuleSourceObj, diagstmp := types.ObjectValue(
		WanNetworkSourceAttrTypes,
		map[string]attr.Value{
			"ip": convert.ParseList(ctx, types.StringType, currentRule.Source.IP,
				"rule.source.ip"),
			"host": oldidname.ParseNameIDList(ctx, currentRule.Source.Host,
				"rule.source.host"),
			"site": oldidname.ParseNameIDList(ctx, currentRule.Source.Site,
				"rule.source.site"),
			"subnet": convert.ParseList(ctx, types.StringType, currentRule.Source.Subnet,
				"rule.source.subnet"),
			"ip_range": convert.ParseFromToList(ctx, currentRule.Source.IPRangeWanNetworkRuleSource,
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
				srcAttrs["network_interface"] = types.SetNull(idname.RefObjectType)
				var diagsTmp diag.Diagnostics
				curRuleSourceObj, diagsTmp = types.ObjectValue(curRuleSourceObj.Type(ctx).(types.ObjectType).AttrTypes, srcAttrs)
				diags = append(diags, diagsTmp...)
			}
		}
	}

	ruleInput.Source = curRuleSourceObj
	// /// /// /// /// end rule.source // /// /// /// /// /

	// /// /// /// / Rule -> Destination // /// /// /// /// /
	curRuleDestinationObj, diagstmp := types.ObjectValue(
		WanNetworkDestAttrTypes,
		map[string]attr.Value{
			"ip": convert.ParseList(ctx, types.StringType, currentRule.Destination.IP,
				"rule.destination.ip"),
			"host": oldidname.ParseNameIDList(ctx, currentRule.Destination.Host,
				"rule.destination.host"),
			"site": oldidname.ParseNameIDList(ctx, currentRule.Destination.Site,
				"rule.destination.site"),
			"subnet": convert.ParseList(ctx, types.StringType, currentRule.Destination.Subnet,
				"rule.destination.subnet"),
			"ip_range": convert.ParseFromToList(ctx, currentRule.Destination.IPRangeWanNetworkRuleDestination,
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

	// /// /// /// / Rule -> Application // /// /// /// /// /
	curRuleApplicationObj, diagstmp := types.ObjectValue(
		WanNetworkApplicationAttrTypes,
		map[string]attr.Value{
			"application": oldidname.ParseNameIDList(ctx, currentRule.Application.Application,
				"rule.application.application"),
			"custom_app": oldidname.ParseNameIDList(ctx, currentRule.Application.CustomApp,
				"rule.application.custom_app"),
			"app_category": oldidname.ParseNameIDList(ctx, currentRule.Application.AppCategory,
				"rule.application.app_category"),
			"custom_category": oldidname.ParseNameIDList(ctx, currentRule.Application.CustomCategory,
				"rule.application.custom_category"),
			"domain": convert.ParseList(ctx, types.StringType, currentRule.Application.Domain,
				"rule.application.domain"),
			"fqdn": convert.ParseList(ctx, types.StringType, currentRule.Application.Fqdn,
				"rule.application.fqdn"),
			"service": oldidname.ParseNameIDList(ctx, currentRule.Application.Service,
				"rule.application.service"),
			"custom_service": parseWanNetworkCustomServiceListSDK(ctx, currentRule.Application.CustomService,
				"rule.application.custom_service"),
			"custom_service_ip": parseWanNetworkCustomServiceIPListSDK(
				ctx,
				currentRule.Application.CustomServiceIP,
				"rule.application.custom_service_ip",
			),
		},
	)
	ruleInput.Application = curRuleApplicationObj
	diags = append(diags, diagstmp...)
	// /// /// /// /// end Rule -> Application // /// /// /// /// /

	// /// /// /// /// start Rule -> Configuration // /// /// /// /// /
	curRuleConfigurationObj, diagstmp := parseWanNetworkConfigurationSDK(ctx, currentRule.Configuration,
		"rule.configuration")
	diags = append(diags, diagstmp...)
	ruleInput.Configuration = curRuleConfigurationObj
	// /// /// /// /// end Rule -> Configuration // /// /// /// /// /

	// /// /// /// /// start Rule -> BandwidthPriority // /// /// /// /// /
	// BandwidthPriority is a struct, not a pointer in SDK type
	if currentRule.BandwidthPriority.ID != "" {
		bandwidthPriorityObj, diagstmp := types.ObjectValue(
			BandwidthPriorityAttrTypes,
			map[string]attr.Value{
				"id":   types.StringValue(currentRule.BandwidthPriority.ID),
				"name": types.StringValue(currentRule.BandwidthPriority.Name),
			},
		)
		diags = append(diags, diagstmp...)
		ruleInput.BandwidthPriority = bandwidthPriorityObj
	} else {
		ruleInput.BandwidthPriority = types.ObjectNull(BandwidthPriorityAttrTypes)
	}
	// /// /// /// /// end Rule -> BandwidthPriority // /// /// /// /// /

	// /// /// /// /// // start Rule -> Exceptions // /// /// /// /// /
	exceptions := []attr.Value{}

	tflog.Warn(ctx, "hydrateWanNetworkRuleState() currentRule.Exceptions", map[string]interface{}{
		"exceptions": currentRule.Exceptions,
		"len":        len(currentRule.Exceptions),
	})
	if len(currentRule.Exceptions) > 0 {
		for _, ruleException := range currentRule.Exceptions {
			// Rule -> Exceptions -> Source
			curExceptionSourceObj, diagstmp := types.ObjectValue(
				WanNetworkSourceAttrTypes,
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
				WanNetworkDestAttrTypes,
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
				WanNetworkExceptionApplicationAttrTypes,
				map[string]attr.Value{
					"application": oldidname.ParseNameIDList(ctx, ruleException.Application.Application,
						"rule.exception.application.application"),
					"custom_app": oldidname.ParseNameIDList(ctx, ruleException.Application.CustomApp,
						"rule.exception.application.custom_app"),
					"app_category": oldidname.ParseNameIDList(ctx, ruleException.Application.AppCategory,
						"rule.exception.application.app_category"),
					"custom_category": oldidname.ParseNameIDList(ctx, ruleException.Application.CustomCategory,
						"rule.exception.application.custom_category"),
					"domain": convert.ParseList(ctx, types.StringType, ruleException.Application.Domain,
						"rule.exception.application.domain"),
					"fqdn": convert.ParseList(ctx, types.StringType, ruleException.Application.Fqdn,
						"rule.exception.application.fqdn"),
					"service": oldidname.ParseNameIDList(ctx, ruleException.Application.Service,
						"rule.exception.application.service"),
					"custom_service": parseWanNetworkExceptionCustomServiceListSDK(
						ctx,
						ruleException.Application.CustomService,
						"rule.exception.application.custom_service",
					),
					"custom_service_ip": parseWanNetworkExceptionCustomServiceIPListSDK(
						ctx,
						ruleException.Application.CustomServiceIP,
						"rule.exception.application.custom_service_ip",
					),
				},
			)
			diags = append(diags, diagstmp...)
			// /// /// /// /// end Rule -> Exceptions -> Application // /// /// /// /// /

			// Initialize Exception object with populated values
			curException, diagstmp := types.ObjectValue(
				WanNetworkExceptionAttrTypes,
				map[string]attr.Value{
					"name":        types.StringValue(ruleException.Name),
					"source":      curExceptionSourceObj,
					"destination": curExceptionDestObj,
					"application": curExceptionApplicationObj,
				},
			)
			diags = append(diags, diagstmp...)
			exceptions = append(exceptions, curException)
		}
		curRuleExceptionsObj, diagstmp := types.SetValue(WanNetworkExceptionObjectType, exceptions)
		diags = append(diags, diagstmp...)
		ruleInput.Exceptions = curRuleExceptionsObj
	} else {
		// Use empty set instead of null to match schema expectations
		curRuleExceptionsObj, diagstmp := types.SetValue(WanNetworkExceptionObjectType, []attr.Value{})
		diags = append(diags, diagstmp...)
		ruleInput.Exceptions = curRuleExceptionsObj
	}
	// /// /// /// /// end Rule -> Exceptions // /// /// /// /// /

	return ruleInput, diags
}

// parseWanNetworkCustomServiceListSDK parses custom service list from SDK types
func parseWanNetworkCustomServiceListSDK(
	ctx context.Context,
	items []*cato_go_sdk.WanNetworkPolicy_Policy_WanNetwork_Policy_Rules_Rule_Application_CustomService,
	_ string,
) types.List {
	if len(items) == 0 {
		return types.ListNull(customservice.CustomServiceObjectType)
	}

	var customServices []types.Object
	for _, item := range items {
		attrs := map[string]attr.Value{
			"port":       types.ListNull(types.StringType),
			"port_range": types.ObjectNull(portrange.PortRangeAttrTypes),
			"protocol":   types.StringNull(),
		}

		if len(item.Port) > 0 {
			var ports []attr.Value
			for _, p := range item.Port {
				ports = append(ports, types.StringValue(fmt.Sprintf("%v", p)))
			}
			attrs["port"], _ = types.ListValue(types.StringType, ports)
		}

		if item.PortRange != nil && (item.PortRange.From != "" || item.PortRange.To != "") {
			attrs["port_range"], _ = types.ObjectValue(
				portrange.PortRangeAttrTypes,
				map[string]attr.Value{
					"from": types.StringValue(fmt.Sprintf("%v", item.PortRange.From)),
					"to":   types.StringValue(fmt.Sprintf("%v", item.PortRange.To)),
				},
			)
		}

		if item.Protocol != "" {
			attrs["protocol"] = types.StringValue(string(item.Protocol))
		}

		obj, _ := types.ObjectValue(WanNetworkCustomServiceAttrTypes, attrs)
		customServices = append(customServices, obj)
	}

	listValue, _ := types.ListValueFrom(ctx, customservice.CustomServiceObjectType, customServices)
	return listValue
}

// parseWanNetworkCustomServiceIPListSDK parses custom service IP list from SDK types
func parseWanNetworkCustomServiceIPListSDK(
	ctx context.Context,
	items []*cato_go_sdk.WanNetworkPolicy_Policy_WanNetwork_Policy_Rules_Rule_Application_CustomServiceIP,
	_ string,
) types.List {
	if len(items) == 0 {
		return types.ListNull(WanNetworkCustomServiceIPObjectType)
	}

	var customServiceIps []types.Object
	for _, item := range items {
		attrs := map[string]attr.Value{
			"name":     types.StringValue(item.Name),
			"ip":       types.StringNull(),
			"ip_range": types.ObjectNull(iprange.RangeAttrTypes),
		}

		if item.IP != nil && *item.IP != "" {
			attrs["ip"] = types.StringValue(*item.IP)
		}

		if item.IPRange != nil {
			attrs["ip_range"], _ = types.ObjectValue(
				iprange.RangeAttrTypes,
				map[string]attr.Value{
					"from": types.StringValue(item.IPRange.From),
					"to":   types.StringValue(item.IPRange.To),
				},
			)
		}

		obj, _ := types.ObjectValue(WanNetworkCustomServiceIPAttrTypes, attrs)
		customServiceIps = append(customServiceIps, obj)
	}

	listValue, _ := types.ListValueFrom(ctx, WanNetworkCustomServiceIPObjectType, customServiceIps)
	return listValue
}

// parseWanNetworkConfigurationSDK parses the configuration object from SDK types
func parseWanNetworkConfigurationSDK(
	ctx context.Context,
	config cato_go_sdk.WanNetworkPolicy_Policy_WanNetwork_Policy_Rules_Rule_Configuration,
	logKey string,
) (types.Object, diag.Diagnostics) {
	diags := make(diag.Diagnostics, 0)

	// Parse primary transport
	primaryTransportObj := types.ObjectNull(WanNetworkTransportAttrTypes)
	if config.PrimaryTransport.TransportType != "" {
		primaryTransportObj, _ = types.ObjectValue(
			WanNetworkTransportAttrTypes,
			map[string]attr.Value{
				"transport_type":           types.StringValue(config.PrimaryTransport.TransportType.String()),
				"primary_interface_role":   types.StringValue(config.PrimaryTransport.PrimaryInterfaceRole.String()),
				"secondary_interface_role": types.StringValue(config.PrimaryTransport.SecondaryInterfaceRole.String()),
			},
		)
	}

	// Parse secondary transport
	secondaryTransportObj := types.ObjectNull(WanNetworkTransportAttrTypes)
	if config.SecondaryTransport.TransportType != "" {
		secondaryTransportObj, _ = types.ObjectValue(
			WanNetworkTransportAttrTypes,
			map[string]attr.Value{
				"transport_type":           types.StringValue(config.SecondaryTransport.TransportType.String()),
				"primary_interface_role":   types.StringValue(config.SecondaryTransport.PrimaryInterfaceRole.String()),
				"secondary_interface_role": types.StringValue(config.SecondaryTransport.SecondaryInterfaceRole.String()),
			},
		)
	}

	configObj, diagstmp := types.ObjectValue(
		WanNetworkConfigurationAttrTypes,
		map[string]attr.Value{
			"active_tcp_acceleration": types.BoolValue(config.ActiveTCPAcceleration),
			"packet_loss_mitigation":  types.BoolValue(config.PacketLossMitigation),
			"preserve_source_port":    types.BoolValue(config.PreserveSourcePort),
			"primary_transport":       primaryTransportObj,
			"secondary_transport":     secondaryTransportObj,
			"allocation_ip":           oldidname.ParseNameIDList(ctx, config.AllocationIP, logKey+".allocation_ip"),
			"pop_location":            oldidname.ParseNameIDList(ctx, config.PopLocation, logKey+".pop_location"),
			"backhauling_site":        oldidname.ParseNameIDList(ctx, config.BackhaulingSite, logKey+".backhauling_site"),
		},
	)
	diags = append(diags, diagstmp...)

	return configObj, diags
}

// parseWanNetworkExceptionCustomServiceListSDK parses exception custom service list from SDK types
func parseWanNetworkExceptionCustomServiceListSDK(
	ctx context.Context,
	items []*cato_go_sdk.WanNetworkPolicy_Policy_WanNetwork_Policy_Rules_Rule_Exceptions_Application_CustomService,
	_ string,
) types.List {
	if len(items) == 0 {
		return types.ListNull(customservice.CustomServiceObjectType)
	}

	var customServices []types.Object
	for _, item := range items {
		attrs := map[string]attr.Value{
			"port":       types.ListNull(types.StringType),
			"port_range": types.ObjectNull(portrange.PortRangeAttrTypes),
			"protocol":   types.StringNull(),
		}

		// Check if port_range is set first
		hasPortRange := item.PortRangeCustomService != nil && (item.PortRangeCustomService.From != "" || item.PortRangeCustomService.To != "")

		if hasPortRange {
			attrs["port_range"], _ = types.ObjectValue(
				portrange.PortRangeAttrTypes,
				map[string]attr.Value{
					"from": types.StringValue(fmt.Sprintf("%v", item.PortRangeCustomService.From)),
					"to":   types.StringValue(fmt.Sprintf("%v", item.PortRangeCustomService.To)),
				},
			)
			// When port_range is set, port must remain null
			attrs["port"] = types.ListNull(types.StringType)
		} else if len(item.Port) > 0 {
			// Only populate port if port_range is not set
			var ports []attr.Value
			for _, p := range item.Port {
				ports = append(ports, types.StringValue(fmt.Sprintf("%v", p)))
			}
			attrs["port"], _ = types.ListValue(types.StringType, ports)
		}

		if item.Protocol != "" {
			attrs["protocol"] = types.StringValue(string(item.Protocol))
		}

		obj, _ := types.ObjectValue(WanNetworkCustomServiceAttrTypes, attrs)
		customServices = append(customServices, obj)
	}

	listValue, _ := types.ListValueFrom(ctx, customservice.CustomServiceObjectType, customServices)
	return listValue
}

// parseWanNetworkExceptionCustomServiceIPListSDK parses exception custom service IP list from SDK types
func parseWanNetworkExceptionCustomServiceIPListSDK(
	ctx context.Context,
	items []*cato_go_sdk.WanNetworkPolicy_Policy_WanNetwork_Policy_Rules_Rule_Exceptions_Application_CustomServiceIP,
	_ string,
) types.List {
	if len(items) == 0 {
		return types.ListNull(WanNetworkCustomServiceIPObjectType)
	}

	var customServiceIps []types.Object
	for _, item := range items {
		attrs := map[string]attr.Value{
			"name":     types.StringValue(item.Name),
			"ip":       types.StringNull(),
			"ip_range": types.ObjectNull(iprange.RangeAttrTypes),
		}

		if item.IP != nil && *item.IP != "" {
			attrs["ip"] = types.StringValue(*item.IP)
		}

		if item.IPRangeCustomServiceIP != nil {
			attrs["ip_range"], _ = types.ObjectValue(
				iprange.RangeAttrTypes,
				map[string]attr.Value{
					"from": types.StringValue(item.IPRangeCustomServiceIP.From),
					"to":   types.StringValue(item.IPRangeCustomServiceIP.To),
				},
			)
		}

		obj, _ := types.ObjectValue(WanNetworkCustomServiceIPAttrTypes, attrs)
		customServiceIps = append(customServiceIps, obj)
	}

	listValue, _ := types.ListValueFrom(ctx, WanNetworkCustomServiceIPObjectType, customServiceIps)
	return listValue
}

// convertCreateInput converts between the Terraform model and SDK values.
func convertCreateInput(
	currentRule *cato_go_sdk.WanNetworkPolicy_Policy_WanNetwork_Policy_Rules_Rule,
	ruleInputRead *PolicyPolicyWanNetworkPolicyRulesRule,
) {
	ruleInputRead.ID = types.StringValue(currentRule.ID)
}

// convertCreateInput2 converts between the Terraform model and SDK values.
func convertCreateInput2(
	ctx context.Context,
	plan Rule,
	priorDiags *diag.Diagnostics,
	priorRuleInputRead *PolicyPolicyWanNetworkPolicyRulesRule,
) (resultRuleObject basetypes.ObjectValue) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	ruleInputRead := *priorRuleInputRead
	defer func() { *priorRuleInputRead = ruleInputRead }()
	if !plan.Rule.IsNull() && !plan.Rule.IsUnknown() {
		planRule := PolicyPolicyWanNetworkPolicyRulesRule{}
		diags = plan.Rule.As(ctx, &planRule, basetypes.ObjectAsOptions{})
		if !diags.HasError() && !planRule.Exceptions.IsNull() && !planRule.Exceptions.IsUnknown() {
			// Correlate exceptions between plan and hydrated response
			correlatedExceptions := wan.CorrelateWanExceptions(ctx, planRule.Exceptions, ruleInputRead.Exceptions)
			if correlatedExceptions != nil {
				ruleInputRead.Exceptions = *correlatedExceptions
			}
		}
	}

	resultRuleObject, diags = types.ObjectValueFrom(ctx, WanNetworkRuleRuleAttrTypes, ruleInputRead)
	return resultRuleObject
}

// convertReadState converts between the Terraform model and SDK values.
func convertReadState(
	ctx context.Context,
	state Rule,
	priorDiags *diag.Diagnostics,
	ruleInput *PolicyPolicyWanNetworkPolicyRulesRule,
) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()

	if !state.Rule.IsNull() && !state.Rule.IsUnknown() {
		stateRule := PolicyPolicyWanNetworkPolicyRulesRule{}
		diags = state.Rule.As(ctx, &stateRule, basetypes.ObjectAsOptions{})
		if !diags.HasError() && !stateRule.Exceptions.IsNull() && !stateRule.Exceptions.IsUnknown() {
			// Correlate exceptions between state and hydrated response
			correlatedExceptions := wan.CorrelateWanExceptions(ctx, stateRule.Exceptions, ruleInput.Exceptions)
			if correlatedExceptions != nil {
				ruleInput.Exceptions = *correlatedExceptions
			}
		}
	}
}

// convertReadState2 converts between the Terraform model and SDK values.
func convertReadState2(
	ctx context.Context,
	state Rule,
	priorDiags *diag.Diagnostics,
	diagnostics *diag.Diagnostics,
) (resultCurAtObj basetypes.ObjectValue, resultDiagstmp diag.Diagnostics) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	positionValue := "LAST_IN_POLICY"
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
	plan Rule,
	priorDiags *diag.Diagnostics,
	input hydrateWanNetworkAPITypes,
	diagnostics *diag.Diagnostics,
) (resultUpdateInput cato_models.WanNetworkUpdateRuleInput) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	var ruleID string
	if !plan.Rule.IsNull() && !plan.Rule.IsUnknown() {
		rule := PolicyPolicyWanNetworkPolicyRulesRule{}
		diags = plan.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{})
		diagnostics.Append(diags...)
		if diagnostics.HasError() {
			return resultUpdateInput
		}
		ruleID = rule.ID.ValueString()
	}

	// Prepare update input with rule ID
	resultUpdateInput = cato_models.WanNetworkUpdateRuleInput{
		ID:   ruleID,
		Rule: input.update.Rule,
	}
	return resultUpdateInput
}

// convertUpdateInput2 converts between the Terraform model and SDK values.
func convertUpdateInput2(
	ctx context.Context,
	plan Rule,
	priorDiags *diag.Diagnostics,
	currentRule *cato_go_sdk.WanNetworkPolicy_Policy_WanNetwork_Policy_Rules_Rule,
	priorRuleInputRead *PolicyPolicyWanNetworkPolicyRulesRule,
) (resultRuleObject basetypes.ObjectValue) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	ruleInputRead := *priorRuleInputRead
	defer func() { *priorRuleInputRead = ruleInputRead }()
	ruleInputRead.ID = types.StringValue(currentRule.ID)

	// Handle exceptions correlation manually to preserve plan structure
	if !plan.Rule.IsNull() && !plan.Rule.IsUnknown() {
		planRule := PolicyPolicyWanNetworkPolicyRulesRule{}
		diags = plan.Rule.As(ctx, &planRule, basetypes.ObjectAsOptions{})
		if !diags.HasError() && !planRule.Exceptions.IsNull() && !planRule.Exceptions.IsUnknown() {
			// Correlate exceptions between plan and hydrated response
			correlatedExceptions := wan.CorrelateWanExceptions(ctx, planRule.Exceptions, ruleInputRead.Exceptions)
			if correlatedExceptions != nil {
				ruleInputRead.Exceptions = *correlatedExceptions
			}
		}
	}

	resultRuleObject, diags = types.ObjectValueFrom(ctx, WanNetworkRuleRuleAttrTypes, ruleInputRead)
	return resultRuleObject
}

// convertDeleteInput converts between the Terraform model and SDK values.
func convertDeleteInput(
	ctx context.Context,
	state Rule,
	priorDiags *diag.Diagnostics,
	diagnostics *diag.Diagnostics,
) (resultRemoveInput cato_models.WanNetworkRemoveRuleInput) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	var ruleID string
	if !state.Rule.IsNull() && !state.Rule.IsUnknown() {
		rule := PolicyPolicyWanNetworkPolicyRulesRule{}
		diags = state.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{})
		diagnostics.Append(diags...)
		if diagnostics.HasError() {
			return resultRemoveInput
		}
		ruleID = rule.ID.ValueString()
	}

	// Prepare remove input
	resultRemoveInput = cato_models.WanNetworkRemoveRuleInput{
		ID: ruleID,
	}
	return resultRemoveInput
}

// convertReadValues converts between the Terraform model and SDK values.
func convertReadValues(
	ctx context.Context,
	state Rule,
	priorDiags *diag.Diagnostics,
) (resultRule PolicyPolicyWanNetworkPolicyRulesRule) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	resultRule = PolicyPolicyWanNetworkPolicyRulesRule{}
	diags = state.Rule.As(ctx, &resultRule, basetypes.ObjectAsOptions{})
	return resultRule
}
