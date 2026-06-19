package provider

import (
	"context"
	"fmt"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/catonetworks/terraform-provider-cato/internal/utils"
)

// applicationControlSourceTF mirrors ApplicationControlSourceAttrTypes / WAN-like source fields.
type applicationControlSourceTF struct {
	Country           types.Set  `tfsdk:"country"`
	IP                types.List `tfsdk:"ip"`
	Host              types.Set  `tfsdk:"host"`
	Site              types.Set  `tfsdk:"site"`
	Subnet            types.List `tfsdk:"subnet"`
	IPRange           types.List `tfsdk:"ip_range"`
	GlobalIPRange     types.Set  `tfsdk:"global_ip_range"`
	NetworkInterface  types.Set  `tfsdk:"network_interface"`
	SiteNetworkSubnet types.Set  `tfsdk:"site_network_subnet"`
	FloatingSubnet    types.Set  `tfsdk:"floating_subnet"`
	User              types.Set  `tfsdk:"user"`
	UsersGroup        types.Set  `tfsdk:"users_group"`
	Group             types.Set  `tfsdk:"group"`
	SystemGroup       types.Set  `tfsdk:"system_group"`
}

//nolint:gocyclo,funlen
func applicationControlSourcePairFromTerraformObject(
	ctx context.Context, source types.Object,
) (*cato_models.ApplicationControlSourceInput, *cato_models.ApplicationControlSourceUpdateInput, diag.Diagnostics) {
	var diags diag.Diagnostics
	if source.IsNull() || source.IsUnknown() {
		return nil, nil, diags
	}

	var sourceInput applicationControlSourceTF
	diags.Append(source.As(ctx, &sourceInput, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil, nil, diags
	}

	add := &cato_models.ApplicationControlSourceInput{}
	upd := &cato_models.ApplicationControlSourceUpdateInput{}

	if !sourceInput.IP.IsUnknown() && !sourceInput.IP.IsNull() {
		diags.Append(sourceInput.IP.ElementsAs(ctx, &add.IP, false)...)
		diags.Append(sourceInput.IP.ElementsAs(ctx, &upd.IP, false)...)
	} else {
		upd.IP = []string{}
	}

	if !sourceInput.Subnet.IsUnknown() && !sourceInput.Subnet.IsNull() {
		diags.Append(sourceInput.Subnet.ElementsAs(ctx, &add.Subnet, false)...)
		diags.Append(sourceInput.Subnet.ElementsAs(ctx, &upd.Subnet, false)...)
	} else {
		upd.Subnet = []string{}
	}

	if !sourceInput.Country.IsUnknown() && !sourceInput.Country.IsNull() {
		objs := make([]types.Object, 0, len(sourceInput.Country.Elements()))
		diags.Append(sourceInput.Country.ElementsAs(ctx, &objs, false)...)
		var item PolicyPolicyInternetFirewallPolicyRulesRuleCountry
		for _, o := range objs {
			diags.Append(o.As(ctx, &item, basetypes.ObjectAsOptions{})...)
			out, err := utils.TransformObjectRefInput(item)
			if err != nil {
				diags.Append(diag.NewErrorDiagnostic("Invalid object reference", fmt.Sprintf("source.country: %v", err)))
				return nil, nil, diags
			}
			add.Country = append(add.Country, &cato_models.CountryRefInput{By: cato_models.ObjectRefBy(out.By), Input: out.Input})
		}
		upd.Country = add.Country
	} else {
		upd.Country = []*cato_models.CountryRefInput{}
	}

	appendHostLike := func(set types.Set, field string,
		appendAdd func(*cato_models.HostRefInput),
	) {
		if set.IsUnknown() || set.IsNull() {
			return
		}
		objs := make([]types.Object, 0, len(set.Elements()))
		diags.Append(set.ElementsAs(ctx, &objs, false)...)
		var item PolicyPolicyInternetFirewallPolicyRulesRuleSourceHost
		for _, o := range objs {
			diags.Append(o.As(ctx, &item, basetypes.ObjectAsOptions{})...)
			out, err := utils.TransformObjectRefInput(item)
			if err != nil {
				diags.Append(diag.NewErrorDiagnostic("Invalid object reference", fmt.Sprintf("%s: %v", field, err)))
				return
			}
			appendAdd(&cato_models.HostRefInput{By: cato_models.ObjectRefBy(out.By), Input: out.Input})
		}
	}

	if !sourceInput.Host.IsUnknown() && !sourceInput.Host.IsNull() {
		appendHostLike(sourceInput.Host, "source.host", func(h *cato_models.HostRefInput) { add.Host = append(add.Host, h) })
		upd.Host = add.Host
	} else {
		upd.Host = []*cato_models.HostRefInput{}
	}

	appendSite := func(set types.Set, field string, appendAdd func(*cato_models.SiteRefInput)) {
		if set.IsUnknown() || set.IsNull() {
			return
		}
		objs := make([]types.Object, 0, len(set.Elements()))
		diags.Append(set.ElementsAs(ctx, &objs, false)...)
		var item PolicyPolicyInternetFirewallPolicyRulesRuleSourceSite
		for _, o := range objs {
			diags.Append(o.As(ctx, &item, basetypes.ObjectAsOptions{})...)
			out, err := utils.TransformObjectRefInput(item)
			if err != nil {
				diags.Append(diag.NewErrorDiagnostic("Invalid object reference", fmt.Sprintf("%s: %v", field, err)))
				return
			}
			appendAdd(&cato_models.SiteRefInput{By: cato_models.ObjectRefBy(out.By), Input: out.Input})
		}
	}
	if !sourceInput.Site.IsUnknown() && !sourceInput.Site.IsNull() {
		appendSite(sourceInput.Site, "source.site", func(s *cato_models.SiteRefInput) { add.Site = append(add.Site, s) })
		upd.Site = add.Site
	} else {
		upd.Site = []*cato_models.SiteRefInput{}
	}

	if !sourceInput.IPRange.IsUnknown() && !sourceInput.IPRange.IsNull() {
		objs := make([]types.Object, 0, len(sourceInput.IPRange.Elements()))
		diags.Append(sourceInput.IPRange.ElementsAs(ctx, &objs, false)...)
		var item PolicyPolicyInternetFirewallPolicyRulesRuleSourceIPRange
		for _, o := range objs {
			diags.Append(o.As(ctx, &item, basetypes.ObjectAsOptions{})...)
			add.IPRange = append(add.IPRange, &cato_models.IPAddressRangeInput{
				From: item.From.ValueString(),
				To:   item.To.ValueString(),
			})
		}
		upd.IPRange = add.IPRange
	} else {
		upd.IPRange = []*cato_models.IPAddressRangeInput{}
	}

	appendGlobalIP := func(set types.Set, field string, appendAdd func(*cato_models.GlobalIPRangeRefInput)) {
		if set.IsUnknown() || set.IsNull() {
			return
		}
		objs := make([]types.Object, 0, len(set.Elements()))
		diags.Append(set.ElementsAs(ctx, &objs, false)...)
		var item PolicyPolicyInternetFirewallPolicyRulesRuleSourceGlobalIPRange
		for _, o := range objs {
			diags.Append(o.As(ctx, &item, basetypes.ObjectAsOptions{})...)
			out, err := utils.TransformObjectRefInput(item)
			if err != nil {
				diags.Append(diag.NewErrorDiagnostic("Invalid object reference", fmt.Sprintf("%s: %v", field, err)))
				return
			}
			appendAdd(&cato_models.GlobalIPRangeRefInput{By: cato_models.ObjectRefBy(out.By), Input: out.Input})
		}
	}
	if !sourceInput.GlobalIPRange.IsUnknown() && !sourceInput.GlobalIPRange.IsNull() {
		appendGlobalIP(sourceInput.GlobalIPRange, "source.global_ip_range", func(g *cato_models.GlobalIPRangeRefInput) {
			add.GlobalIPRange = append(add.GlobalIPRange, g)
		})
		upd.GlobalIPRange = add.GlobalIPRange
	} else {
		upd.GlobalIPRange = []*cato_models.GlobalIPRangeRefInput{}
	}

	appendNI := func(set types.Set, field string, appendAdd func(*cato_models.NetworkInterfaceRefInput)) {
		if set.IsUnknown() || set.IsNull() {
			return
		}
		objs := make([]types.Object, 0, len(set.Elements()))
		diags.Append(set.ElementsAs(ctx, &objs, false)...)
		var item PolicyPolicyInternetFirewallPolicyRulesRuleSourceNetworkInterface
		for _, o := range objs {
			diags.Append(o.As(ctx, &item, basetypes.ObjectAsOptions{})...)
			out, err := utils.TransformObjectRefInput(item)
			if err != nil {
				diags.Append(diag.NewErrorDiagnostic("Invalid object reference", fmt.Sprintf("%s: %v", field, err)))
				return
			}
			appendAdd(&cato_models.NetworkInterfaceRefInput{By: cato_models.ObjectRefBy(out.By), Input: out.Input})
		}
	}
	if !sourceInput.NetworkInterface.IsUnknown() && !sourceInput.NetworkInterface.IsNull() {
		appendNI(sourceInput.NetworkInterface, "source.network_interface", func(v *cato_models.NetworkInterfaceRefInput) {
			add.NetworkInterface = append(add.NetworkInterface, v)
		})
		upd.NetworkInterface = add.NetworkInterface
	} else {
		upd.NetworkInterface = []*cato_models.NetworkInterfaceRefInput{}
	}

	appendSNS := func(set types.Set, field string, appendAdd func(*cato_models.SiteNetworkSubnetRefInput)) {
		if set.IsUnknown() || set.IsNull() {
			return
		}
		objs := make([]types.Object, 0, len(set.Elements()))
		diags.Append(set.ElementsAs(ctx, &objs, false)...)
		var item PolicyPolicyInternetFirewallPolicyRulesRuleSourceSiteNetworkSubnet
		for _, o := range objs {
			diags.Append(o.As(ctx, &item, basetypes.ObjectAsOptions{})...)
			out, err := utils.TransformObjectRefInput(item)
			if err != nil {
				diags.Append(diag.NewErrorDiagnostic("Invalid object reference", fmt.Sprintf("%s: %v", field, err)))
				return
			}
			appendAdd(&cato_models.SiteNetworkSubnetRefInput{By: cato_models.ObjectRefBy(out.By), Input: out.Input})
		}
	}
	if !sourceInput.SiteNetworkSubnet.IsUnknown() && !sourceInput.SiteNetworkSubnet.IsNull() {
		appendSNS(sourceInput.SiteNetworkSubnet, "source.site_network_subnet", func(v *cato_models.SiteNetworkSubnetRefInput) {
			add.SiteNetworkSubnet = append(add.SiteNetworkSubnet, v)
		})
		upd.SiteNetworkSubnet = add.SiteNetworkSubnet
	} else {
		upd.SiteNetworkSubnet = []*cato_models.SiteNetworkSubnetRefInput{}
	}

	appendFloat := func(set types.Set, field string, appendAdd func(*cato_models.FloatingSubnetRefInput)) {
		if set.IsUnknown() || set.IsNull() {
			return
		}
		objs := make([]types.Object, 0, len(set.Elements()))
		diags.Append(set.ElementsAs(ctx, &objs, false)...)
		var item PolicyPolicyInternetFirewallPolicyRulesRuleSourceFloatingSubnet
		for _, o := range objs {
			diags.Append(o.As(ctx, &item, basetypes.ObjectAsOptions{})...)
			out, err := utils.TransformObjectRefInput(item)
			if err != nil {
				diags.Append(diag.NewErrorDiagnostic("Invalid object reference", fmt.Sprintf("%s: %v", field, err)))
				return
			}
			appendAdd(&cato_models.FloatingSubnetRefInput{By: cato_models.ObjectRefBy(out.By), Input: out.Input})
		}
	}
	if !sourceInput.FloatingSubnet.IsUnknown() && !sourceInput.FloatingSubnet.IsNull() {
		appendFloat(sourceInput.FloatingSubnet, "source.floating_subnet", func(v *cato_models.FloatingSubnetRefInput) {
			add.FloatingSubnet = append(add.FloatingSubnet, v)
		})
		upd.FloatingSubnet = add.FloatingSubnet
	} else {
		upd.FloatingSubnet = []*cato_models.FloatingSubnetRefInput{}
	}

	appendUser := func(set types.Set, field string, appendAdd func(*cato_models.UserRefInput)) {
		if set.IsUnknown() || set.IsNull() {
			return
		}
		objs := make([]types.Object, 0, len(set.Elements()))
		diags.Append(set.ElementsAs(ctx, &objs, false)...)
		var item PolicyPolicyInternetFirewallPolicyRulesRuleSourceUser
		for _, o := range objs {
			diags.Append(o.As(ctx, &item, basetypes.ObjectAsOptions{})...)
			out, err := utils.TransformObjectRefInput(item)
			if err != nil {
				diags.Append(diag.NewErrorDiagnostic("Invalid object reference", fmt.Sprintf("%s: %v", field, err)))
				return
			}
			appendAdd(&cato_models.UserRefInput{By: cato_models.ObjectRefBy(out.By), Input: out.Input})
		}
	}
	if !sourceInput.User.IsUnknown() && !sourceInput.User.IsNull() {
		appendUser(sourceInput.User, "source.user", func(v *cato_models.UserRefInput) { add.User = append(add.User, v) })
		upd.User = add.User
	} else {
		upd.User = []*cato_models.UserRefInput{}
	}

	appendUG := func(set types.Set, field string, appendAdd func(*cato_models.UsersGroupRefInput)) {
		if set.IsUnknown() || set.IsNull() {
			return
		}
		objs := make([]types.Object, 0, len(set.Elements()))
		diags.Append(set.ElementsAs(ctx, &objs, false)...)
		var item PolicyPolicyInternetFirewallPolicyRulesRuleSourceUsersGroup
		for _, o := range objs {
			diags.Append(o.As(ctx, &item, basetypes.ObjectAsOptions{})...)
			out, err := utils.TransformObjectRefInput(item)
			if err != nil {
				diags.Append(diag.NewErrorDiagnostic("Invalid object reference", fmt.Sprintf("%s: %v", field, err)))
				return
			}
			appendAdd(&cato_models.UsersGroupRefInput{By: cato_models.ObjectRefBy(out.By), Input: out.Input})
		}
	}
	if !sourceInput.UsersGroup.IsUnknown() && !sourceInput.UsersGroup.IsNull() {
		appendUG(sourceInput.UsersGroup, "source.users_group", func(v *cato_models.UsersGroupRefInput) {
			add.UsersGroup = append(add.UsersGroup, v)
		})
		upd.UsersGroup = add.UsersGroup
	} else {
		upd.UsersGroup = []*cato_models.UsersGroupRefInput{}
	}

	appendGroup := func(set types.Set, field string, appendAdd func(*cato_models.GroupRefInput)) {
		if set.IsUnknown() || set.IsNull() {
			return
		}
		objs := make([]types.Object, 0, len(set.Elements()))
		diags.Append(set.ElementsAs(ctx, &objs, false)...)
		var item PolicyPolicyInternetFirewallPolicyRulesRuleSourceGroup
		for _, o := range objs {
			diags.Append(o.As(ctx, &item, basetypes.ObjectAsOptions{})...)
			out, err := utils.TransformObjectRefInput(item)
			if err != nil {
				diags.Append(diag.NewErrorDiagnostic("Invalid object reference", fmt.Sprintf("%s: %v", field, err)))
				return
			}
			appendAdd(&cato_models.GroupRefInput{By: cato_models.ObjectRefBy(out.By), Input: out.Input})
		}
	}
	if !sourceInput.Group.IsUnknown() && !sourceInput.Group.IsNull() {
		appendGroup(sourceInput.Group, "source.group", func(v *cato_models.GroupRefInput) { add.Group = append(add.Group, v) })
		upd.Group = add.Group
	} else {
		upd.Group = []*cato_models.GroupRefInput{}
	}

	appendSys := func(set types.Set, field string, appendAdd func(*cato_models.SystemGroupRefInput)) {
		if set.IsUnknown() || set.IsNull() {
			return
		}
		objs := make([]types.Object, 0, len(set.Elements()))
		diags.Append(set.ElementsAs(ctx, &objs, false)...)
		var item PolicyPolicyInternetFirewallPolicyRulesRuleSourceSystemGroup
		for _, o := range objs {
			diags.Append(o.As(ctx, &item, basetypes.ObjectAsOptions{})...)
			out, err := utils.TransformObjectRefInput(item)
			if err != nil {
				diags.Append(diag.NewErrorDiagnostic("Invalid object reference", fmt.Sprintf("%s: %v", field, err)))
				return
			}
			appendAdd(&cato_models.SystemGroupRefInput{By: cato_models.ObjectRefBy(out.By), Input: out.Input})
		}
	}
	if !sourceInput.SystemGroup.IsUnknown() && !sourceInput.SystemGroup.IsNull() {
		appendSys(sourceInput.SystemGroup, "source.system_group", func(v *cato_models.SystemGroupRefInput) {
			add.SystemGroup = append(add.SystemGroup, v)
		})
		upd.SystemGroup = add.SystemGroup
	} else {
		upd.SystemGroup = []*cato_models.SystemGroupRefInput{}
	}

	if diags.HasError() {
		return nil, nil, diags
	}
	return add, upd, diags
}

func cloneAppTenantRestrictionSource(in *cato_models.ApplicationControlSourceInput) *cato_models.AppTenantRestrictionSourceInput {
	if in == nil {
		return nil
	}
	return &cato_models.AppTenantRestrictionSourceInput{
		Country:           in.Country,
		FloatingSubnet:    in.FloatingSubnet,
		GlobalIPRange:     in.GlobalIPRange,
		Group:             in.Group,
		Host:              in.Host,
		IP:                in.IP,
		IPRange:           in.IPRange,
		NetworkInterface:  in.NetworkInterface,
		Site:              in.Site,
		SiteNetworkSubnet: in.SiteNetworkSubnet,
		Subnet:            in.Subnet,
		SystemGroup:       in.SystemGroup,
		User:              in.User,
		UsersGroup:        in.UsersGroup,
	}
}

func cloneAppTenantRestrictionSourceUpdate(
	in *cato_models.ApplicationControlSourceUpdateInput,
) *cato_models.AppTenantRestrictionSourceUpdateInput {
	if in == nil {
		return nil
	}
	return &cato_models.AppTenantRestrictionSourceUpdateInput{
		Country:           in.Country,
		FloatingSubnet:    in.FloatingSubnet,
		GlobalIPRange:     in.GlobalIPRange,
		Group:             in.Group,
		Host:              in.Host,
		IP:                in.IP,
		IPRange:           in.IPRange,
		NetworkInterface:  in.NetworkInterface,
		Site:              in.Site,
		SiteNetworkSubnet: in.SiteNetworkSubnet,
		Subnet:            in.Subnet,
		SystemGroup:       in.SystemGroup,
		User:              in.User,
		UsersGroup:        in.UsersGroup,
	}
}
