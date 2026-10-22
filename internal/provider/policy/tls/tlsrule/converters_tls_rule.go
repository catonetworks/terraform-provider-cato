package tlsrule

import (
	"context"
	"fmt"
	"reflect"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	cato_scalars "github.com/catonetworks/cato-go-sdk/scalars"
	attr "github.com/hashicorp/terraform-plugin-framework/attr"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	convert "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	customservice "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/customservice"
	idname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	iprange "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/iprange"
	oldidname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/oldidname"
	policyrule "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

// hydrateTLSRuleAPI takes in the current state/plan along with context and returns the created
// diagnostic data as well as cato API data used to either create or update TLS entries
//
//nolint:gocyclo,funlen
func hydrateTLSRuleAPI(ctx context.Context, plan InspectionRule) (hydrateTLSAPITypes, diag.Diagnostics) {
	diags := []diag.Diagnostic{}

	hydrateAPIReturn := hydrateTLSAPITypes{}
	hydrateAPIReturn.create = cato_models.TLSInspectAddRuleInput{}
	hydrateAPIReturn.update = cato_models.TLSInspectUpdateRuleInput{}
	hydrateAPIReturn.create.At = &cato_models.PolicyRulePositionInput{}

	rootAddRule := &cato_models.TLSInspectAddRuleDataInput{}
	rootUpdateRule := &cato_models.TLSInspectUpdateRuleDataInput{}

	// setting at for creation only
	if !plan.At.IsNull() {
		positionInput := policyrule.Position{}
		diags = append(diags, plan.At.As(ctx, &positionInput, basetypes.ObjectAsOptions{})...)

		hydrateAPIReturn.create.At.Position = (*cato_models.PolicyRulePositionEnum)(positionInput.Position.ValueStringPointer())
		hydrateAPIReturn.create.At.Ref = positionInput.Ref.ValueStringPointer()
	}

	// setting rule
	if !plan.Rule.IsNull() {
		ruleInput := PolicyPolicyTLSInspectPolicyRulesRule{}
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
			rootAddRule.UntrustedCertificateAction = cato_models.TLSInspectUntrustedCertificateAction(
				ruleInput.UntrustedCertificateAction.ValueString(),
			)
			untrustedCertVal := cato_models.TLSInspectUntrustedCertificateAction(ruleInput.UntrustedCertificateAction.ValueString())
			rootUpdateRule.UntrustedCertificateAction = &untrustedCertVal
		}

		// setting source
		if !ruleInput.Source.IsNull() {
			ruleSourceInput := &cato_models.TLSInspectSourceInput{}
			ruleSourceUpdateInput := &cato_models.TLSInspectSourceUpdateInput{}

			sourceInput := PolicyPolicyTLSInspectPolicyRulesRuleSource{}
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

				var itemSourceHostInput idname.Ref
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

				var itemSourceSiteInput idname.Ref
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

				var itemSourceGlobalIPRangeInput idname.Ref
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

				var itemSourceSiteNetworkSubnetInput idname.Ref
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

				var itemSourceFloatingSubnetInput idname.Ref
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

				var itemSourceUserInput idname.Ref
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

				var itemSourceUsersGroupInput idname.Ref
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

				var itemSourceGroupInput idname.Ref
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

				var itemSourceSystemGroupInput idname.Ref
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

				var itemSourceNetworkInterfaceInput idname.Ref
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
		// /// /// /// /// end rule.source // /// /// /// /// /

		// setting country
		if !ruleInput.Country.IsNull() && !ruleInput.Country.IsUnknown() {
			elementsCountryInput := make([]types.Object, 0, len(ruleInput.Country.Elements()))
			diags = append(diags, ruleInput.Country.ElementsAs(ctx, &elementsCountryInput, false)...)

			var itemCountryInput idname.Ref
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

			var itemDevicePostureProfileInput idname.Ref
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

			applicationInput := PolicyPolicyTLSInspectPolicyRulesRuleApplication{}
			diags = append(diags, ruleInput.Application.As(ctx, &applicationInput, basetypes.ObjectAsOptions{})...)

			// setting application.application
			if !applicationInput.Application.IsUnknown() && !applicationInput.Application.IsNull() {
				elementsApplicationInput := make([]types.Object, 0, len(applicationInput.Application.Elements()))
				diags = append(diags, applicationInput.Application.ElementsAs(ctx, &elementsApplicationInput, false)...)

				var itemApplicationInput idname.Ref
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

				var itemCustomAppInput idname.Ref
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

				var itemAppCategoryInput idname.Ref
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

				var itemCustomCategoryInput idname.Ref
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

				var itemApplicationGlobalIPRangeInput idname.Ref
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

				var itemApplicationServiceInput idname.Ref
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

				var itemApplicationCountryInput idname.Ref
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
		// /// /// /// /// end rule.application // /// /// /// /// /
	}

	hydrateAPIReturn.create.Rule = rootAddRule
	hydrateAPIReturn.update.Rule = rootUpdateRule

	return hydrateAPIReturn, diags
}

//nolint:gocyclo,funlen
func hydrateTLSRuleState(
	ctx context.Context,
	state InspectionRule,
	currentRule *cato_go_sdk.Tlsinspectpolicy_Policy_TLSInspect_Policy_Rules_Rule,
) (PolicyPolicyTLSInspectPolicyRulesRule, diag.Diagnostics) {
	ruleInput := PolicyPolicyTLSInspectPolicyRulesRule{}
	diags := make(diag.Diagnostics, 0)

	// Handle case where state might be incomplete (e.g., during import)
	// Try to extract existing state first, but don't fail if it's incomplete
	if !state.Rule.IsNull() && !state.Rule.IsUnknown() {
		diagstmp := state.Rule.As(ctx, &ruleInput, basetypes.ObjectAsOptions{})
		if diagstmp.HasError() {
			// If conversion fails during import, log it but continue with empty ruleInput
			tflog.Debug(ctx, "Failed to convert state to struct in hydrateTlsRuleState, using empty struct for hydration")
			// Reset ruleInput to empty struct and clear diagnostics
			ruleInput = PolicyPolicyTLSInspectPolicyRulesRule{}
			// Don't add the errors to diags, we'll handle this by populating from API response
		} else {
			diags = append(diags, diagstmp...)
		}
	}

	ruleInput.Name = types.StringValue(currentRule.Name)
	if currentRule.Description == "" {
		ruleInput.Description = types.StringNull()
	} else {
		ruleInput.Description = types.StringValue(currentRule.Description)
	}
	ruleInput.Action = types.StringValue(currentRule.Action.String())
	ruleInput.ID = types.StringValue(currentRule.ID)
	ruleInput.Enabled = types.BoolValue(currentRule.Enabled)

	// Handle optional fields
	if currentRule.UntrustedCertificateAction.String() != "" {
		ruleInput.UntrustedCertificateAction = types.StringValue(currentRule.UntrustedCertificateAction.String())
	} else {
		ruleInput.UntrustedCertificateAction = types.StringNull()
	}

	if currentRule.ConnectionOrigin.String() != "" {
		ruleInput.ConnectionOrigin = types.StringValue(currentRule.ConnectionOrigin.String())
	} else {
		ruleInput.ConnectionOrigin = types.StringNull()
	}

	// Platform is a list of OperatingSystem enums, not a single string
	if len(currentRule.Platform) > 0 {
		// For now, just store the first platform value if present
		ruleInput.Platform = types.StringValue(currentRule.Platform[0].String())
	} else {
		ruleInput.Platform = types.StringNull()
	}

	// // /// /// /// / Rule -> Source // /// /// /// /// /
	curRuleSourceObj, diagstmp := types.ObjectValue(
		TLSSourceAttrTypes,
		map[string]attr.Value{
			"ip":                  convert.ParseList(ctx, types.StringType, currentRule.Source.IP, "rule.source.ip"),
			"host":                oldidname.ParseNameIDList(ctx, currentRule.Source.Host, "rule.source.host"),
			"site":                oldidname.ParseNameIDList(ctx, currentRule.Source.Site, "rule.source.site"),
			"subnet":              convert.ParseList(ctx, types.StringType, currentRule.Source.Subnet, "rule.source.subnet"),
			"ip_range":            convert.ParseFromToList(ctx, currentRule.Source.IPRangeTLSInspectSource, "rule.source.ip_range"),
			"global_ip_range":     oldidname.ParseNameIDList(ctx, currentRule.Source.GlobalIPRange, "rule.source.global_ip_range"),
			"network_interface":   oldidname.ParseNameIDList(ctx, currentRule.Source.NetworkInterface, "rule.source.network_interface"),
			"site_network_subnet": oldidname.ParseNameIDList(ctx, currentRule.Source.SiteNetworkSubnet, "rule.source.site_network_subnet"),
			"floating_subnet":     oldidname.ParseNameIDList(ctx, currentRule.Source.FloatingSubnet, "rule.source.floating_subnet"),
			"user":                oldidname.ParseNameIDList(ctx, currentRule.Source.User, "rule.source.user"),
			"users_group":         oldidname.ParseNameIDList(ctx, currentRule.Source.UsersGroup, "rule.source.users_group"),
			"group":               oldidname.ParseNameIDList(ctx, currentRule.Source.Group, "rule.source.group"),
			"system_group":        oldidname.ParseNameIDList(ctx, currentRule.Source.SystemGroup, "rule.source.system_group"),
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
	ruleInput.Country = oldidname.ParseNameIDList(ctx, currentRule.Country, "rule.country")

	// Rule -> Device Posture Profile
	ruleInput.DevicePostureProfile = oldidname.ParseNameIDList(ctx, currentRule.DevicePostureProfile, "rule.device_posture_profile")

	// // /// /// /// / Rule -> Application // /// /// /// /// /
	// Note: RemoteAsn is []scalars.Asn32 (which is type string in SDK), so convert.ParseList handles it directly

	// Handle custom_service - only parse if it exists in API response
	// API returns an array, but schema expects a single object - take first element
	var customServiceValue attr.Value
	if len(currentRule.Application.CustomService) > 0 {
		customServiceValue = customservice.ParseCustomService(ctx, currentRule.Application.CustomService[0], "rule.application.custom_service")
	} else {
		// Fallback to state/plan value if present to avoid drift
		if !ruleInput.Application.IsNull() && !ruleInput.Application.IsUnknown() {
			if appAttrs := ruleInput.Application.Attributes(); appAttrs != nil {
				if v, ok := appAttrs["custom_service"]; ok && !v.IsNull() && !v.IsUnknown() {
					customServiceValue = v
				} else {
					customServiceValue = types.ObjectNull(customservice.CustomServiceAttrTypes)
				}
			} else {
				customServiceValue = types.ObjectNull(customservice.CustomServiceAttrTypes)
			}
		} else {
			customServiceValue = types.ObjectNull(customservice.CustomServiceAttrTypes)
		}
	}

	// Handle custom_service_ip - only parse if it exists in API response
	// API returns an array, but schema expects a single object - take first element
	var customServiceIPValue attr.Value
	if len(currentRule.Application.CustomServiceIP) > 0 {
		customServiceIPValue = parseCustomServiceIP(ctx, currentRule.Application.CustomServiceIP[0], "rule.application.custom_service_ip")
	} else {
		// Fallback to state/plan value if present to avoid drift
		if !ruleInput.Application.IsNull() && !ruleInput.Application.IsUnknown() {
			if appAttrs := ruleInput.Application.Attributes(); appAttrs != nil {
				if v, ok := appAttrs["custom_service_ip"]; ok && !v.IsNull() && !v.IsUnknown() {
					customServiceIPValue = v
				} else {
					customServiceIPValue = types.ObjectNull(customservice.CustomServiceIPAttrTypes)
				}
			} else {
				customServiceIPValue = types.ObjectNull(customservice.CustomServiceIPAttrTypes)
			}
		} else {
			customServiceIPValue = types.ObjectNull(customservice.CustomServiceIPAttrTypes)
		}
	}

	// Handle tls_inspect_category - only parse if it exists in API response
	// API returns an array, but schema expects a single string - take first element
	var tlsInspectCategoryValue attr.Value
	if len(currentRule.Application.TLSInspectCategory) > 0 {
		tlsInspectCategoryValue = parseTLSInspectCategory(ctx, currentRule.Application.TLSInspectCategory[0])
	} else {
		// Fallback to state/plan value if present to avoid drift
		if !ruleInput.Application.IsNull() && !ruleInput.Application.IsUnknown() {
			if appAttrs := ruleInput.Application.Attributes(); appAttrs != nil {
				if v, ok := appAttrs["tls_inspect_category"]; ok && !v.IsNull() && !v.IsUnknown() {
					tlsInspectCategoryValue = v
				} else {
					tlsInspectCategoryValue = types.StringNull()
				}
			} else {
				tlsInspectCategoryValue = types.StringNull()
			}
		} else {
			tlsInspectCategoryValue = types.StringNull()
		}
	}

	curRuleApplicationObj, diagstmp := types.ObjectValue(
		TLSApplicationAttrTypes,
		map[string]attr.Value{
			"application":          oldidname.ParseNameIDList(ctx, currentRule.Application.Application, "rule.application.application"),
			"custom_app":           oldidname.ParseNameIDList(ctx, currentRule.Application.CustomApp, "rule.application.custom_app"),
			"app_category":         oldidname.ParseNameIDList(ctx, currentRule.Application.AppCategory, "rule.application.app_category"),
			"custom_category":      oldidname.ParseNameIDList(ctx, currentRule.Application.CustomCategory, "rule.application.custom_category"),
			"domain":               convert.ParseList(ctx, types.StringType, currentRule.Application.Domain, "rule.application.domain"),
			"fqdn":                 convert.ParseList(ctx, types.StringType, currentRule.Application.Fqdn, "rule.application.fqdn"),
			"ip":                   convert.ParseList(ctx, types.StringType, currentRule.Application.IP, "rule.application.ip"),
			"subnet":               convert.ParseList(ctx, types.StringType, currentRule.Application.Subnet, "rule.application.subnet"),
			"ip_range":             convert.ParseFromToList(ctx, currentRule.Application.IPRangeTLSInspectApplication, "rule.application.ip_range"),
			"global_ip_range":      oldidname.ParseNameIDList(ctx, currentRule.Application.GlobalIPRange, "rule.application.global_ip_range"),
			"remote_asn":           convert.ParseList(ctx, types.StringType, currentRule.Application.RemoteAsn, "rule.application.remote_asn"),
			"service":              oldidname.ParseNameIDList(ctx, currentRule.Application.Service, "rule.application.service"),
			"custom_service":       customServiceValue,
			"custom_service_ip":    customServiceIPValue,
			"tls_inspect_category": tlsInspectCategoryValue,
			"country":              oldidname.ParseNameIDList(ctx, currentRule.Application.Country, "rule.application.country"),
		},
	)
	ruleInput.Application = curRuleApplicationObj
	diags = append(diags, diagstmp...)
	// /// /// /// /// end Rule -> Application // /// /// /// /// /

	return ruleInput, diags
}

// parseTLSInspectCategory handles the TLS inspection category field
func parseTLSInspectCategory(_ context.Context, category interface{}) types.String {
	if category == nil {
		return types.StringNull()
	}
	// Handle string type
	if categoryStr, ok := category.(string); ok {
		if categoryStr == "" {
			return types.StringNull()
		}
		return types.StringValue(categoryStr)
	}
	// Handle enum type with String() method
	if stringer, ok := category.(interface{ String() string }); ok {
		categoryStr := stringer.String()
		if categoryStr == "" {
			return types.StringNull()
		}
		return types.StringValue(categoryStr)
	}
	return types.StringNull()
}

// parseCustomServiceIP handles the custom service IP object from API response
//
//nolint:gocyclo
func parseCustomServiceIP(ctx context.Context, item interface{}, attrName string) types.Object {
	tflog.Debug(ctx, "parseCustomServiceIP() "+attrName+" - "+fmt.Sprintf("%v", item))

	// Get the reflect.Value of the input
	itemValue := reflect.ValueOf(item)

	// Handle nil or invalid input
	if item == nil || itemValue.Kind() != reflect.Struct {
		if itemValue.Kind() == reflect.Pointer && !itemValue.IsNil() {
			itemValue = itemValue.Elem()
			for itemValue.Kind() == reflect.Pointer && !itemValue.IsNil() {
				itemValue = itemValue.Elem()
			}
			if itemValue.Kind() != reflect.Struct {
				return types.ObjectNull(customservice.CustomServiceIPAttrTypes)
			}
		} else {
			return types.ObjectNull(customservice.CustomServiceIPAttrTypes)
		}
	}

	// Handle pointer to struct
	if itemValue.Kind() == reflect.Pointer {
		if itemValue.IsNil() {
			return types.ObjectNull(customservice.CustomServiceIPAttrTypes)
		}
		itemValue = itemValue.Elem()
	}

	// Get fields
	nameField := itemValue.FieldByName("Name")
	ipField := itemValue.FieldByName("IP")
	ipRangeField := itemValue.FieldByName("IPRange")

	// Handle name
	var nameVal types.String
	if nameField.IsValid() && nameField.String() != "" {
		nameVal = types.StringValue(nameField.String())
	} else {
		nameVal = types.StringNull()
	}

	// Handle IP - only include if it has a value
	var ipVal types.String
	if ipField.IsValid() && !ipField.IsZero() && ipField.String() != "" {
		ipVal = types.StringValue(ipField.String())
	} else {
		ipVal = types.StringNull()
	}

	// Handle IP range
	var ipRangeVal types.Object
	switch {
	case ipRangeField.Kind() == reflect.Pointer:
		if ipRangeField.IsNil() {
			ipRangeVal = types.ObjectNull(iprange.RangeAttrTypes)
		} else {
			ipRangeField = ipRangeField.Elem()
			from := ipRangeField.FieldByName("From")
			to := ipRangeField.FieldByName("To")
			var diagsTmp diag.Diagnostics
			ipRangeVal, diagsTmp = types.ObjectValue(
				iprange.RangeAttrTypes,
				map[string]attr.Value{
					"from": types.StringValue(from.String()),
					"to":   types.StringValue(to.String()),
				},
			)
			_ = diagsTmp
		}
	case ipRangeField.IsValid():
		from := ipRangeField.FieldByName("From")
		to := ipRangeField.FieldByName("To")
		var diagsTmp diag.Diagnostics
		ipRangeVal, diagsTmp = types.ObjectValue(
			iprange.RangeAttrTypes,
			map[string]attr.Value{
				"from": types.StringValue(from.String()),
				"to":   types.StringValue(to.String()),
			},
		)
		_ = diagsTmp
	default:
		ipRangeVal = types.ObjectNull(iprange.RangeAttrTypes)
	}

	// Create final custom service IP object
	obj, diagstmp := types.ObjectValue(
		customservice.CustomServiceIPAttrTypes,
		map[string]attr.Value{
			"name":     nameVal,
			"ip":       ipVal,
			"ip_range": ipRangeVal,
		},
	)
	tflog.Debug(ctx, "parseCustomServiceIP() obj - "+obj.String())
	_ = diagstmp
	return obj
}

// convertCreateInput converts between the Terraform model and SDK values.
func convertCreateInput(
	currentRule *cato_go_sdk.Tlsinspectpolicy_Policy_TLSInspect_Policy_Rules_Rule,
	ruleInputRead *PolicyPolicyTLSInspectPolicyRulesRule,
) {
	ruleInputRead.ID = types.StringValue(currentRule.ID)
}

// convertCreateInput2 converts between the Terraform model and SDK values.
func convertCreateInput2(
	ctx context.Context,
	priorDiags *diag.Diagnostics,
	ruleInputRead PolicyPolicyTLSInspectPolicyRulesRule,
) (resultRuleObject basetypes.ObjectValue) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	resultRuleObject, diags = types.ObjectValueFrom(ctx, TLSInspectionRuleRuleAttrTypes, ruleInputRead)
	return resultRuleObject
}

// convertCreateInput3 converts between the Terraform model and SDK values.
func convertCreateInput3(
	plan *InspectionRule,
	currentRule *cato_go_sdk.Tlsinspectpolicy_Policy_TLSInspect_Policy_Rules_Rule,
	ruleObject basetypes.ObjectValue,
) {
	plan.Rule = ruleObject
	plan.ID = types.StringValue(currentRule.ID)
}

// convertReadState converts between the Terraform model and SDK values.
func convertReadState(
	ctx context.Context,
	state InspectionRule,
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
	plan InspectionRule,
	priorDiags *diag.Diagnostics,
	diagnostics *diag.Diagnostics,
) (
	resultInputMoveRule cato_models.PolicyMoveRuleInput,
	resultRuleInput PolicyPolicyTLSInspectPolicyRulesRule,
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

	resultRuleInput = PolicyPolicyTLSInspectPolicyRulesRule{}
	diags = plan.Rule.As(ctx, &resultRuleInput, basetypes.ObjectAsOptions{})
	return resultInputMoveRule, resultRuleInput
}

// convertUpdateInput2 converts between the Terraform model and SDK values.
func convertUpdateInput2(
	ctx context.Context,
	priorDiags *diag.Diagnostics,
	currentRule *cato_go_sdk.Tlsinspectpolicy_Policy_TLSInspect_Policy_Rules_Rule,
	priorRuleInputRead *PolicyPolicyTLSInspectPolicyRulesRule,
) (resultRuleObject basetypes.ObjectValue) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	ruleInputRead := *priorRuleInputRead
	defer func() { *priorRuleInputRead = ruleInputRead }()
	ruleInputRead.ID = types.StringValue(currentRule.ID)

	resultRuleObject, diags = types.ObjectValueFrom(ctx, TLSInspectionRuleRuleAttrTypes, ruleInputRead)
	return resultRuleObject
}

// convertDeleteInput converts between the Terraform model and SDK values.
func convertDeleteInput(
	ctx context.Context,
	state InspectionRule,
	priorDiags *diag.Diagnostics,
) (resultRuleInput PolicyPolicyTLSInspectPolicyRulesRule) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	resultRuleInput = PolicyPolicyTLSInspectPolicyRulesRule{}
	diags = state.Rule.As(ctx, &resultRuleInput, basetypes.ObjectAsOptions{})
	return resultRuleInput
}

// convertDeleteInput2 converts between the Terraform model and SDK values.
func convertDeleteInput2(
	ruleInput PolicyPolicyTLSInspectPolicyRulesRule,
) (resultRemoveRuleInput cato_models.TLSInspectRemoveRuleInput) {
	resultRemoveRuleInput = cato_models.TLSInspectRemoveRuleInput{
		ID: *ruleInput.ID.ValueStringPointer(),
	}
	return resultRemoveRuleInput
}

// convertReadValues converts between the Terraform model and SDK values.
func convertReadValues(
	ctx context.Context,
	state InspectionRule,
	priorDiags *diag.Diagnostics,
) (resultRule PolicyPolicyTLSInspectPolicyRulesRule) {
	diags := *priorDiags
	defer func() { *priorDiags = diags }()
	resultRule = PolicyPolicyTLSInspectPolicyRulesRule{}
	diags = state.Rule.As(ctx, &resultRule, basetypes.ObjectAsOptions{})
	return resultRule
}
