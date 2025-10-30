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
)

// hydrateTlsApiTypes create sub-types for both create and update calls to populate both entries
type hydrateTlsApiTypes struct {
	create cato_models.TLSInspectAddRuleInput
	update cato_models.TLSInspectUpdateRuleInput
}

// hydrateTlsRuleApi takes in the current state/plan along with context and returns the created
// diagnostic data as well as cato api data used to either create or update TLS entries
func hydrateTlsRuleApi(ctx context.Context, plan TlsInspectionRule) (hydrateTlsApiTypes, diag.Diagnostics) {
	diags := []diag.Diagnostic{}

	hydrateApiReturn := hydrateTlsApiTypes{}
	hydrateApiReturn.create = cato_models.TLSInspectAddRuleInput{}
	hydrateApiReturn.update = cato_models.TLSInspectUpdateRuleInput{}
	hydrateApiReturn.create.At = &cato_models.PolicyRulePositionInput{}

	rootAddRule := &cato_models.TLSInspectAddRuleDataInput{}
	rootUpdateRule := &cato_models.TLSInspectUpdateRuleDataInput{}

	//setting at for creation only
	if !plan.At.IsNull() {

		positionInput := PolicyRulePositionInput{}
		diags = append(diags, plan.At.As(ctx, &positionInput, basetypes.ObjectAsOptions{})...)

		hydrateApiReturn.create.At.Position = (*cato_models.PolicyRulePositionEnum)(positionInput.Position.ValueStringPointer())
		hydrateApiReturn.create.At.Ref = positionInput.Ref.ValueStringPointer()

	}

	// setting rule
	if !plan.Rule.IsNull() {

		ruleInput := Policy_Policy_TlsInspect_Policy_Rules_Rule{}
		diags = append(diags, plan.Rule.As(ctx, &ruleInput, basetypes.ObjectAsOptions{})...)

		// setting name
		if !ruleInput.Name.IsNull() && !ruleInput.Name.IsUnknown() {
			rootAddRule.Name = ruleInput.Name.ValueString()
			nameVal := ruleInput.Name.ValueString()
			rootUpdateRule.Name = &nameVal
		}

		// setting description
		if !ruleInput.Description.IsNull() && !ruleInput.Description.IsUnknown() {
			rootAddRule.Description = ruleInput.Description.ValueString()
			descVal := ruleInput.Description.ValueString()
			rootUpdateRule.Description = &descVal
		} else {
			emptyDesc := ""
			rootUpdateRule.Description = &emptyDesc
		}

		// setting enabled
		if !ruleInput.Enabled.IsNull() && !ruleInput.Enabled.IsUnknown() {
			rootAddRule.Enabled = ruleInput.Enabled.ValueBool()
			enabledVal := ruleInput.Enabled.ValueBool()
			rootUpdateRule.Enabled = &enabledVal
		}

		// setting action
		if !ruleInput.Action.IsNull() && !ruleInput.Action.IsUnknown() {
			rootAddRule.Action = cato_models.TLSInspectAction(ruleInput.Action.ValueString())
			actionVal := cato_models.TLSInspectAction(ruleInput.Action.ValueString())
			rootUpdateRule.Action = &actionVal
		}

		// setting connection origin
		if !ruleInput.ConnectionOrigin.IsNull() && !ruleInput.ConnectionOrigin.IsUnknown() {
			rootAddRule.ConnectionOrigin = cato_models.ConnectionOriginEnum(ruleInput.ConnectionOrigin.ValueString())
			connOriginVal := cato_models.ConnectionOriginEnum(ruleInput.ConnectionOrigin.ValueString())
			rootUpdateRule.ConnectionOrigin = &connOriginVal
		}

		// setting untrusted certificate action
		if !ruleInput.UntrustedCertificateAction.IsNull() && !ruleInput.UntrustedCertificateAction.IsUnknown() {
			rootAddRule.UntrustedCertificateAction = cato_models.TLSInspectUntrustedCertificateAction(ruleInput.UntrustedCertificateAction.ValueString())
			untrustedCertVal := cato_models.TLSInspectUntrustedCertificateAction(ruleInput.UntrustedCertificateAction.ValueString())
			rootUpdateRule.UntrustedCertificateAction = &untrustedCertVal
		}

		// setting source
		if !ruleInput.Source.IsNull() {

			ruleSourceInput := &cato_models.TLSInspectSourceInput{}
			ruleSourceUpdateInput := &cato_models.TLSInspectSourceUpdateInput{}

			sourceInput := Policy_Policy_TlsInspect_Policy_Rules_Rule_Source{}
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

				var itemSourceHostInput NameIDRef
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

				var itemSourceSiteInput NameIDRef
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

				var itemSourceIPRangeInput FromTo
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

				var itemSourceGlobalIPRangeInput NameIDRef
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

			// setting source site network subnet
			if !sourceInput.SiteNetworkSubnet.IsNull() {
				elementsSourceSiteNetworkSubnetInput := make([]types.Object, 0, len(sourceInput.SiteNetworkSubnet.Elements()))
				diags = append(diags, sourceInput.SiteNetworkSubnet.ElementsAs(ctx, &elementsSourceSiteNetworkSubnetInput, false)...)

				var itemSourceSiteNetworkSubnetInput NameIDRef
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

				var itemSourceFloatingSubnetInput NameIDRef
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

				var itemSourceUserInput NameIDRef
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

				var itemSourceUsersGroupInput NameIDRef
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

				var itemSourceGroupInput NameIDRef
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

				var itemSourceSystemGroupInput NameIDRef
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

			// setting source network_interface
			if !sourceInput.NetworkInterface.IsNull() {
				elementsSourceNetworkInterfaceInput := make([]types.Object, 0, len(sourceInput.NetworkInterface.Elements()))
				diags = append(diags, sourceInput.NetworkInterface.ElementsAs(ctx, &elementsSourceNetworkInterfaceInput, false)...)

				var itemSourceNetworkInterfaceInput NameIDRef
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

			rootAddRule.Source = ruleSourceInput
			rootUpdateRule.Source = ruleSourceUpdateInput
		}
		////////////// end rule.source ///////////////

		// setting country
		if !ruleInput.Country.IsNull() && !ruleInput.Country.IsUnknown() {
			elementsCountryInput := make([]types.Object, 0, len(ruleInput.Country.Elements()))
			diags = append(diags, ruleInput.Country.ElementsAs(ctx, &elementsCountryInput, false)...)

			var itemCountryInput NameIDRef
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

		// setting device posture profile
		if !ruleInput.DevicePostureProfile.IsNull() && !ruleInput.DevicePostureProfile.IsUnknown() {
			elementsDevicePostureProfileInput := make([]types.Object, 0, len(ruleInput.DevicePostureProfile.Elements()))
			diags = append(diags, ruleInput.DevicePostureProfile.ElementsAs(ctx, &elementsDevicePostureProfileInput, false)...)

			var itemDevicePostureProfileInput NameIDRef
			for _, item := range elementsDevicePostureProfileInput {
				diags = append(diags, item.As(ctx, &itemDevicePostureProfileInput, basetypes.ObjectAsOptions{})...)

				ObjectRefOutput, err := utils.TransformObjectRefInput(itemDevicePostureProfileInput)
				if err != nil {
					tflog.Error(ctx, err.Error())
				}

				rootAddRule.DevicePostureProfile = append(rootAddRule.DevicePostureProfile, &cato_models.DeviceProfileRefInput{
					By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
					Input: ObjectRefOutput.Input,
				})
			}
			rootUpdateRule.DevicePostureProfile = rootAddRule.DevicePostureProfile
		} else {
			rootUpdateRule.DevicePostureProfile = make([]*cato_models.DeviceProfileRefInput, 0)
		}

		// setting platform
		if !ruleInput.Platform.IsNull() && !ruleInput.Platform.IsUnknown() {
			// Platform is currently a string in the type definition, so we handle it as a single value
			platformVal := ruleInput.Platform.ValueString()
			if platformVal != "" {
				rootAddRule.Platform = append(rootAddRule.Platform, cato_models.OperatingSystem(platformVal))
				rootUpdateRule.Platform = rootAddRule.Platform
			} else {
				rootUpdateRule.Platform = make([]cato_models.OperatingSystem, 0)
			}
		} else {
			rootUpdateRule.Platform = make([]cato_models.OperatingSystem, 0)
		}

		// setting application
		if !ruleInput.Application.IsNull() {

			ruleApplicationInput := &cato_models.TLSInspectApplicationInput{}
			ruleApplicationUpdateInput := &cato_models.TLSInspectApplicationUpdateInput{}

			applicationInput := Policy_Policy_TlsInspect_Policy_Rules_Rule_Application{}
			diags = append(diags, ruleInput.Application.As(ctx, &applicationInput, basetypes.ObjectAsOptions{})...)

			// setting application.application
			if !applicationInput.Application.IsUnknown() && !applicationInput.Application.IsNull() {
				elementsApplicationInput := make([]types.Object, 0, len(applicationInput.Application.Elements()))
				diags = append(diags, applicationInput.Application.ElementsAs(ctx, &elementsApplicationInput, false)...)

				var itemApplicationInput NameIDRef
				for _, item := range elementsApplicationInput {
					diags = append(diags, item.As(ctx, &itemApplicationInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationInput)
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

			// setting application.custom_app
			if !applicationInput.CustomApp.IsUnknown() && !applicationInput.CustomApp.IsNull() {
				elementsCustomAppInput := make([]types.Object, 0, len(applicationInput.CustomApp.Elements()))
				diags = append(diags, applicationInput.CustomApp.ElementsAs(ctx, &elementsCustomAppInput, false)...)

				var itemCustomAppInput NameIDRef
				for _, item := range elementsCustomAppInput {
					diags = append(diags, item.As(ctx, &itemCustomAppInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemCustomAppInput)
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

			// setting application.app_category
			if !applicationInput.AppCategory.IsUnknown() && !applicationInput.AppCategory.IsNull() {
				elementsAppCategoryInput := make([]types.Object, 0, len(applicationInput.AppCategory.Elements()))
				diags = append(diags, applicationInput.AppCategory.ElementsAs(ctx, &elementsAppCategoryInput, false)...)

				var itemAppCategoryInput NameIDRef
				for _, item := range elementsAppCategoryInput {
					diags = append(diags, item.As(ctx, &itemAppCategoryInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemAppCategoryInput)
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

			// setting application.custom_category
			if !applicationInput.CustomCategory.IsUnknown() && !applicationInput.CustomCategory.IsNull() {
				elementsCustomCategoryInput := make([]types.Object, 0, len(applicationInput.CustomCategory.Elements()))
				diags = append(diags, applicationInput.CustomCategory.ElementsAs(ctx, &elementsCustomCategoryInput, false)...)

				var itemCustomCategoryInput NameIDRef
				for _, item := range elementsCustomCategoryInput {
					diags = append(diags, item.As(ctx, &itemCustomCategoryInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemCustomCategoryInput)
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


			// setting application.domain
			if !applicationInput.Domain.IsUnknown() && !applicationInput.Domain.IsNull() {
				diags = append(diags, applicationInput.Domain.ElementsAs(ctx, &ruleApplicationInput.Domain, false)...)
				diags = append(diags, applicationInput.Domain.ElementsAs(ctx, &ruleApplicationUpdateInput.Domain, false)...)
			} else {
				ruleApplicationUpdateInput.Domain = make([]string, 0)
			}

			// setting application.fqdn
			if !applicationInput.Fqdn.IsUnknown() && !applicationInput.Fqdn.IsNull() {
				diags = append(diags, applicationInput.Fqdn.ElementsAs(ctx, &ruleApplicationInput.Fqdn, false)...)
				diags = append(diags, applicationInput.Fqdn.ElementsAs(ctx, &ruleApplicationUpdateInput.Fqdn, false)...)
			} else {
				ruleApplicationUpdateInput.Fqdn = make([]string, 0)
			}

			// setting application.ip
			if !applicationInput.IP.IsUnknown() && !applicationInput.IP.IsNull() {
				diags = append(diags, applicationInput.IP.ElementsAs(ctx, &ruleApplicationInput.IP, false)...)
				diags = append(diags, applicationInput.IP.ElementsAs(ctx, &ruleApplicationUpdateInput.IP, false)...)
			} else {
				ruleApplicationUpdateInput.IP = make([]string, 0)
			}

			// setting application.subnet
			if !applicationInput.Subnet.IsUnknown() && !applicationInput.Subnet.IsNull() {
				diags = append(diags, applicationInput.Subnet.ElementsAs(ctx, &ruleApplicationInput.Subnet, false)...)
				diags = append(diags, applicationInput.Subnet.ElementsAs(ctx, &ruleApplicationUpdateInput.Subnet, false)...)
			} else {
				ruleApplicationUpdateInput.Subnet = make([]string, 0)
			}

			// setting application.ip_range
			if !applicationInput.IPRange.IsUnknown() && !applicationInput.IPRange.IsNull() {
				elementsApplicationIPRangeInput := make([]types.Object, 0, len(applicationInput.IPRange.Elements()))
				diags = append(diags, applicationInput.IPRange.ElementsAs(ctx, &elementsApplicationIPRangeInput, false)...)

				var itemApplicationIPRangeInput FromTo
				for _, item := range elementsApplicationIPRangeInput {
					diags = append(diags, item.As(ctx, &itemApplicationIPRangeInput, basetypes.ObjectAsOptions{})...)

					ruleApplicationInput.IPRange = append(ruleApplicationInput.IPRange, &cato_models.IPAddressRangeInput{
						From: itemApplicationIPRangeInput.From.ValueString(),
						To:   itemApplicationIPRangeInput.To.ValueString(),
					})
				}
				ruleApplicationUpdateInput.IPRange = ruleApplicationInput.IPRange
			} else {
				ruleApplicationUpdateInput.IPRange = make([]*cato_models.IPAddressRangeInput, 0)
			}

			// setting application.global_ip_range
			if !applicationInput.GlobalIPRange.IsNull() {
				elementsApplicationGlobalIPRangeInput := make([]types.Object, 0, len(applicationInput.GlobalIPRange.Elements()))
				diags = append(diags, applicationInput.GlobalIPRange.ElementsAs(ctx, &elementsApplicationGlobalIPRangeInput, false)...)

				var itemApplicationGlobalIPRangeInput NameIDRef
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
			} else {
				ruleApplicationUpdateInput.GlobalIPRange = make([]*cato_models.GlobalIPRangeRefInput, 0)
			}

			// setting application.remote_asn
			// Note: RemoteAsn is []scalars.Asn32 in SDK (string type), converted from Terraform []string
			if !applicationInput.RemoteAsn.IsUnknown() && !applicationInput.RemoteAsn.IsNull() {
				diags = append(diags, applicationInput.RemoteAsn.ElementsAs(ctx, &ruleApplicationInput.RemoteAsn, false)...)
				diags = append(diags, applicationInput.RemoteAsn.ElementsAs(ctx, &ruleApplicationUpdateInput.RemoteAsn, false)...)
			} else {
				ruleApplicationUpdateInput.RemoteAsn = make([]cato_scalars.Asn32, 0)
			}

			// setting application.service
			if !applicationInput.Service.IsUnknown() && !applicationInput.Service.IsNull() {
				elementsApplicationServiceInput := make([]types.Object, 0, len(applicationInput.Service.Elements()))
				diags = append(diags, applicationInput.Service.ElementsAs(ctx, &elementsApplicationServiceInput, false)...)

				var itemApplicationServiceInput NameIDRef
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

			// setting application.country
			if !applicationInput.Country.IsUnknown() && !applicationInput.Country.IsNull() {
				elementsApplicationCountryInput := make([]types.Object, 0, len(applicationInput.Country.Elements()))
				diags = append(diags, applicationInput.Country.ElementsAs(ctx, &elementsApplicationCountryInput, false)...)

				var itemApplicationCountryInput NameIDRef
				for _, item := range elementsApplicationCountryInput {
					diags = append(diags, item.As(ctx, &itemApplicationCountryInput, basetypes.ObjectAsOptions{})...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationCountryInput)
					if err != nil {
						tflog.Error(ctx, err.Error())
					}

					ruleApplicationInput.Country = append(ruleApplicationInput.Country, &cato_models.CountryRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
				ruleApplicationUpdateInput.Country = ruleApplicationInput.Country
			} else {
				ruleApplicationUpdateInput.Country = make([]*cato_models.CountryRefInput, 0)
			}

			// TODO: Add custom_service, custom_service_ip, and tls_inspect_category if needed
			// These fields are in the schema but may require additional implementation

			rootAddRule.Application = ruleApplicationInput
			rootUpdateRule.Application = ruleApplicationUpdateInput
		}
		////////////// end rule.application ///////////////
	}

	hydrateApiReturn.create.Rule = rootAddRule
	hydrateApiReturn.update.Rule = rootUpdateRule

	return hydrateApiReturn, diags
}
