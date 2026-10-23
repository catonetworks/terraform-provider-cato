package lfsubpolicy

import (
	"context"
	"fmt"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	scalars "github.com/catonetworks/cato-go-sdk/scalars"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	path "github.com/hashicorp/terraform-plugin-framework/path"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	convert "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	customservice "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/customservice"
	idname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	iprange "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/iprange"
	policyrule "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
	portrange "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/portrange"
	simpleservice "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/simpleservice"
	socketlandestination "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/socketlandestination"
	socketlansite "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/socketlansite"
	socketlansource "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/socketlansource"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

// prepareAt converts the Terraform rule position into an API position input.
func (r *lfSubPolicyResource) prepareAt(ctx context.Context, at types.Object, diags *diag.Diagnostics,
) *cato_models.PolicyRulePositionInput {
	if !utils.HasValue(at) {
		return nil
	}
	var tfPosition policyrule.Position
	if utils.CheckErr(diags, at.As(ctx, &tfPosition, basetypes.ObjectAsOptions{})) {
		return nil
	}

	return &cato_models.PolicyRulePositionInput{
		Position: (*cato_models.PolicyRulePositionEnum)(tfPosition.Position.ValueStringPointer()),
		Ref:      tfPosition.Ref.ValueStringPointer(),
	}
}

func checkPolicyMutationStatus(
	status *cato_models.PolicyMutationStatus,
	summary string,
	diags *diag.Diagnostics,
) bool {
	if status != nil && *status == cato_models.PolicyMutationStatusSuccess {
		return false
	}
	detail := "API mutation returned no status"
	if status != nil {
		detail = fmt.Sprintf("API mutation returned status %q", *status)
	}
	diags.AddError(summary, detail)
	return true
}

// preparePolicy builds the API input containing the sub-policy name and description.
func (r *lfSubPolicyResource) preparePolicy(name, desc types.String) *cato_models.SocketLanAddSubPolicyDataInput {
	return &cato_models.SocketLanAddSubPolicyDataInput{
		Description: desc.ValueString(),
		Name:        name.ValueString(),
	}
}

// prepareScope converts the Terraform scope rule into an API add-rule input.
func (r *lfSubPolicyResource) prepareScope(ctx context.Context, scope types.Object, name, desc types.String, diags *diag.Diagnostics,
) *cato_models.SocketLanAddRuleDataInput {
	if !utils.HasValue(scope) {
		return nil
	}
	var tfScope LanFirewallSubPolicyScope
	if utils.CheckErr(diags, scope.As(ctx, &tfScope, basetypes.ObjectAsOptions{})) {
		return nil
	}

	return &cato_models.SocketLanAddRuleDataInput{
		Description: desc.ValueString(),
		Destination: r.prepareDestination(ctx, tfScope.Destination, diags),
		Direction:   cato_models.SocketLanDirection(tfScope.Direction.ValueString()),
		Enabled:     tfScope.Enabled.ValueBool(),
		Name:        name.ValueString(),
		Nat:         r.prepareNat(ctx, tfScope.NAT, diags),
		Service:     r.prepareService(ctx, tfScope.Service, diags),
		Site:        r.prepareSite(ctx, tfScope.Site, diags),
		Source:      r.prepareSource(ctx, tfScope.Source, diags),
		Transport:   cato_models.SocketLanTransportTypeLan,
	}
}

// prepareSite converts Terraform site and group references into an API site input.
func (r *lfSubPolicyResource) prepareSite(
	ctx context.Context,
	site types.Object,
	diags *diag.Diagnostics,
) *cato_models.SocketLanSiteInput {
	if !utils.HasValue(site) {
		return nil
	}
	var tfSite socketlansite.SocketLanSite
	if utils.CheckErr(diags, site.As(ctx, &tfSite, basetypes.ObjectAsOptions{})) {
		return nil
	}
	return &cato_models.SocketLanSiteInput{
		Group: idname.PrepareIDNameSet[cato_models.GroupRefInput](ctx, tfSite.Group, diags),
		Site:  idname.PrepareIDNameSet[cato_models.SiteRefInput](ctx, tfSite.Site, diags),
	}
}

// prepareSiteUpdate converts Terraform site and group references into an API site update input.
func (r *lfSubPolicyResource) prepareSiteUpdate(ctx context.Context, site types.Object, diags *diag.Diagnostics,
) *cato_models.SocketLanSiteUpdateInput {
	if !utils.HasValue(site) {
		return nil
	}
	upd := r.prepareSite(ctx, site, diags)
	return &cato_models.SocketLanSiteUpdateInput{
		Group: upd.Group,
		Site:  upd.Site,
	}
}

// prepareSource converts Terraform source selectors into an API source input.
func (r *lfSubPolicyResource) prepareSource(ctx context.Context, src types.Object, diags *diag.Diagnostics,
) *cato_models.SocketLanSourceInput {
	if !utils.HasValue(src) {
		return nil
	}
	var tfSource socketlansource.SocketLanSource
	if utils.CheckErr(diags, src.As(ctx, &tfSource, basetypes.ObjectAsOptions{})) {
		return nil
	}
	return &cato_models.SocketLanSourceInput{
		FloatingSubnet:    idname.PrepareIDNameSet[cato_models.FloatingSubnetRefInput](ctx, tfSource.FloatingSubnet, diags),
		GlobalIPRange:     idname.PrepareIDNameSet[cato_models.GlobalIPRangeRefInput](ctx, tfSource.GlobalIPRange, diags),
		Group:             idname.PrepareIDNameSet[cato_models.GroupRefInput](ctx, tfSource.Group, diags),
		Host:              idname.PrepareIDNameSet[cato_models.HostRefInput](ctx, tfSource.Host, diags),
		IP:                convert.PrepareStringList[string](ctx, tfSource.IP, diags),
		IPRange:           r.prepareIPRange(ctx, tfSource.IPRange, diags),
		NetworkInterface:  idname.PrepareIDNameSet[cato_models.NetworkInterfaceRefInput](ctx, tfSource.NetworkInterface, diags),
		SiteNetworkSubnet: idname.PrepareIDNameSet[cato_models.SiteNetworkSubnetRefInput](ctx, tfSource.SiteNetworkSubnet, diags),
		Subnet:            convert.PrepareStringList[string](ctx, tfSource.Subnet, diags),
		SystemGroup:       idname.PrepareIDNameSet[cato_models.SystemGroupRefInput](ctx, tfSource.SystemGroup, diags),
		Vlan:              convert.PrepareInt64List[scalars.Vlan](ctx, tfSource.Vlan, diags),
	}
}

// prepareSourceUpdate converts Terraform source selectors into an API source update input.
func (r *lfSubPolicyResource) prepareSourceUpdate(ctx context.Context, src types.Object, diags *diag.Diagnostics,
) *cato_models.SocketLanSourceUpdateInput {
	if !utils.HasValue(src) {
		return nil
	}
	upd := r.prepareSource(ctx, src, diags)
	return &cato_models.SocketLanSourceUpdateInput{
		FloatingSubnet:    upd.FloatingSubnet,
		GlobalIPRange:     upd.GlobalIPRange,
		Group:             upd.Group,
		Host:              upd.Host,
		IP:                upd.IP,
		IPRange:           upd.IPRange,
		NetworkInterface:  upd.NetworkInterface,
		SiteNetworkSubnet: upd.SiteNetworkSubnet,
		Subnet:            upd.Subnet,
		SystemGroup:       upd.SystemGroup,
		Vlan:              upd.Vlan,
	}
}

// prepareDestination converts Terraform destination selectors into an API destination input.
func (r *lfSubPolicyResource) prepareDestination(ctx context.Context, dest types.Object, diags *diag.Diagnostics,
) *cato_models.SocketLanDestinationInput {
	if !utils.HasValue(dest) {
		return nil
	}
	var tfDestination socketlandestination.SocketLanDestination
	if utils.CheckErr(diags, dest.As(ctx, &tfDestination, basetypes.ObjectAsOptions{})) {
		return nil
	}
	return &cato_models.SocketLanDestinationInput{
		FloatingSubnet:    idname.PrepareIDNameSet[cato_models.FloatingSubnetRefInput](ctx, tfDestination.FloatingSubnet, diags),
		GlobalIPRange:     idname.PrepareIDNameSet[cato_models.GlobalIPRangeRefInput](ctx, tfDestination.GlobalIPRange, diags),
		Group:             idname.PrepareIDNameSet[cato_models.GroupRefInput](ctx, tfDestination.Group, diags),
		Host:              idname.PrepareIDNameSet[cato_models.HostRefInput](ctx, tfDestination.Host, diags),
		IP:                convert.PrepareStringList[string](ctx, tfDestination.IP, diags),
		IPRange:           r.prepareIPRange(ctx, tfDestination.IPRange, diags),
		NetworkInterface:  idname.PrepareIDNameSet[cato_models.NetworkInterfaceRefInput](ctx, tfDestination.NetworkInterface, diags),
		SiteNetworkSubnet: idname.PrepareIDNameSet[cato_models.SiteNetworkSubnetRefInput](ctx, tfDestination.SiteNetworkSubnet, diags),
		Subnet:            convert.PrepareStringList[string](ctx, tfDestination.Subnet, diags),
		SystemGroup:       idname.PrepareIDNameSet[cato_models.SystemGroupRefInput](ctx, tfDestination.SystemGroup, diags),
		Vlan:              convert.PrepareInt64List[scalars.Vlan](ctx, tfDestination.Vlan, diags),
	}
}

// prepareDestinationUpdate converts Terraform destination selectors into an API destination update input.
func (r *lfSubPolicyResource) prepareDestinationUpdate(ctx context.Context, dest types.Object, diags *diag.Diagnostics,
) *cato_models.SocketLanDestinationUpdateInput {
	if !utils.HasValue(dest) {
		return nil
	}
	upd := r.prepareDestination(ctx, dest, diags)
	return &cato_models.SocketLanDestinationUpdateInput{
		FloatingSubnet:    upd.FloatingSubnet,
		GlobalIPRange:     upd.GlobalIPRange,
		Group:             upd.Group,
		Host:              upd.Host,
		IP:                upd.IP,
		IPRange:           upd.IPRange,
		NetworkInterface:  upd.NetworkInterface,
		SiteNetworkSubnet: upd.SiteNetworkSubnet,
		Subnet:            upd.Subnet,
		SystemGroup:       upd.SystemGroup,
		Vlan:              upd.Vlan,
	}
}

// prepareIPRange converts Terraform IP ranges into API address range inputs.
func (r *lfSubPolicyResource) prepareIPRange(ctx context.Context, ipRange types.List, diags *diag.Diagnostics,
) []*cato_models.IPAddressRangeInput {
	if !utils.HasValue(ipRange) {
		return nil
	}
	var tfFromTo []iprange.Range
	if utils.CheckErr(diags, ipRange.ElementsAs(ctx, &tfFromTo, false)) {
		return nil
	}

	out := make([]*cato_models.IPAddressRangeInput, 0, len(tfFromTo))
	for _, o := range tfFromTo {
		out = append(out, &cato_models.IPAddressRangeInput{
			From: o.From.ValueString(),
			To:   o.To.ValueString(),
		})
	}
	return out
}

// prepareNat converts Terraform NAT settings into an API input with disabled dynamic PAT defaults.
func (r *lfSubPolicyResource) prepareNat(ctx context.Context, nat types.Object, diags *diag.Diagnostics,
) *cato_models.SocketLanNatSettingsInput {
	if !utils.HasValue(nat) {
		return &cato_models.SocketLanNatSettingsInput{
			Enabled: false,
			NatType: cato_models.SocketLanNatTypeDynamicPat,
		}
	}
	var tfNat PolicyNatSettings
	if utils.CheckErr(diags, nat.As(ctx, &tfNat, basetypes.ObjectAsOptions{})) {
		return nil
	}
	natType := tfNat.NatType.ValueString()
	if natType == "" {
		natType = string(cato_models.SocketLanNatTypeDynamicPat)
	}
	return &cato_models.SocketLanNatSettingsInput{
		Enabled: tfNat.Enabled.ValueBool(),
		NatType: cato_models.SocketLanNatType(natType),
	}
}

// prepareNatUpdate converts Terraform NAT settings into an API update input.
func (r *lfSubPolicyResource) prepareNatUpdate(ctx context.Context, nat types.Object, diags *diag.Diagnostics,
) *cato_models.SocketLanNatSettingsUpdateInput {
	upd := r.prepareNat(ctx, nat, diags)
	return &cato_models.SocketLanNatSettingsUpdateInput{
		Enabled: &upd.Enabled,
		NatType: &upd.NatType,
	}
}

// prepareService converts Terraform custom and simple services into an API service input.
func (r *lfSubPolicyResource) prepareService(ctx context.Context, svc types.Object, diags *diag.Diagnostics,
) *cato_models.SocketLanServiceInput {
	if !utils.HasValue(svc) {
		return &cato_models.SocketLanServiceInput{}
	}
	var tfService PolicyService
	if utils.CheckErr(diags, svc.As(ctx, &tfService, basetypes.ObjectAsOptions{})) {
		return nil
	}
	return &cato_models.SocketLanServiceInput{
		Custom: r.prepareCustomService(ctx, tfService.Custom, diags),
		Simple: r.prepareSimpleService(ctx, tfService.Simple, diags),
	}
}

// prepareServiceUpdate converts Terraform services into an API service update input.
func (r *lfSubPolicyResource) prepareServiceUpdate(ctx context.Context, svc types.Object, diags *diag.Diagnostics,
) *cato_models.SocketLanServiceUpdateInput {
	upd := r.prepareService(ctx, svc, diags)
	return &cato_models.SocketLanServiceUpdateInput{
		Custom: upd.Custom,
		Simple: upd.Simple,
	}
}

// prepareSimpleService converts Terraform simple service names into API inputs.
func (r *lfSubPolicyResource) prepareSimpleService(ctx context.Context, svc types.Set, diags *diag.Diagnostics,
) []*cato_models.SimpleServiceInput {
	if !utils.HasValue(svc) {
		return nil
	}
	var tfSimpleServices []simpleservice.SimpleService
	if utils.CheckErr(diags, svc.ElementsAs(ctx, &tfSimpleServices, false)) {
		return nil
	}
	out := make([]*cato_models.SimpleServiceInput, 0, len(tfSimpleServices))
	for _, s := range tfSimpleServices {
		svcInput := cato_models.SimpleServiceInput{
			Name: cato_models.SimpleServiceType(s.Name.ValueString()),
		}
		out = append(out, &svcInput)
	}
	return out
}

// prepareCustomService converts Terraform custom service definitions into API inputs.
func (r *lfSubPolicyResource) prepareCustomService(ctx context.Context, svc types.List, diags *diag.Diagnostics,
) []*cato_models.CustomServiceInput {
	if !utils.HasValue(svc) {
		return nil
	}
	var tfCustServices []PolicyCustomService
	if utils.CheckErr(diags, svc.ElementsAs(ctx, &tfCustServices, false)) {
		return nil
	}
	out := make([]*cato_models.CustomServiceInput, 0, len(tfCustServices))
	for _, s := range tfCustServices {
		svcInput := cato_models.CustomServiceInput{
			Port:      convert.PrepareStringList[scalars.Port](ctx, s.Port, diags),
			PortRange: r.preparePortRange(ctx, s.PortRange, diags),
			Protocol:  cato_models.IPProtocol(s.Protocol.ValueString()),
		}
		out = append(out, &svcInput)
	}
	return out
}

// preparePortRange converts a Terraform port range into an API port range input.
func (r *lfSubPolicyResource) preparePortRange(ctx context.Context, portRange types.Object, diags *diag.Diagnostics,
) *cato_models.PortRangeInput {
	if !utils.HasValue(portRange) {
		return nil
	}
	var tfFromTo portrange.PortRange
	if utils.CheckErr(diags, portRange.As(ctx, &tfFromTo, basetypes.ObjectAsOptions{})) {
		return nil
	}

	return &cato_models.PortRangeInput{
		From: scalars.Port(tfFromTo.From.ValueString()),
		To:   scalars.Port(tfFromTo.To.ValueString()),
	}
}

// parseRuleScope pares the rule scope from the API response and returns terraform LanFirewallSubPolicyScope object
func (r *lfSubPolicyResource) parseRuleScope(ctx context.Context,
	rule cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule, diags *diag.Diagnostics,
) types.Object {
	// Prepare LanFirewallSubPolicyScope object
	tfScope := LanFirewallSubPolicyScope{
		Description: types.StringValue(rule.Description),
		Destination: r.parseDestination(ctx, rule.Destination, diags),
		Direction:   types.StringValue(string(rule.Direction)),
		Enabled:     types.BoolValue(rule.Enabled),
		ID:          types.StringValue(rule.ID),
		Name:        types.StringValue(rule.Name),
		NAT:         r.parseNat(ctx, rule.Nat, diags),
		Service:     r.parseService(ctx, rule.Service, diags),
		Site:        r.parseSite(ctx, rule.Site, diags),
		Source:      r.parseSource(ctx, rule.Source, diags),
	}
	subPolicyObj, objDiags := types.ObjectValueFrom(ctx, LanFirewallSubPolicyScopeTypes, tfScope)
	diags.Append(objDiags...)
	if diags.HasError() {
		return types.ObjectNull(LanFirewallSubPolicyScopeTypes)
	}

	return subPolicyObj
}

// parseSource parses the source of a rule into a SocketLanSource object
func (r *lfSubPolicyResource) parseSource(ctx context.Context,
	dst cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Source, diags *diag.Diagnostics,
) types.Object {
	tfSocketLanSource := socketlansource.SocketLanSource{
		Vlan:              convert.ParseInt64List(ctx, dst.GetVlan(), diags),
		IP:                convert.ParseStringList(ctx, dst.GetIP(), diags),
		Subnet:            convert.ParseStringList(ctx, dst.GetSubnet(), diags),
		IPRange:           FromToList(ctx, dst.GetIPRange(), diags),
		Host:              idname.ParseIDNameSet(ctx, dst.GetHost(), diags),
		Group:             idname.ParseIDNameSet(ctx, dst.GetGroup(), diags),
		SystemGroup:       idname.ParseIDNameSet(ctx, dst.GetSystemGroup(), diags),
		NetworkInterface:  idname.ParseIDNameSet(ctx, dst.GetNetworkInterface(), diags),
		GlobalIPRange:     idname.ParseIDNameSet(ctx, dst.GetGlobalIPRange(), diags),
		FloatingSubnet:    idname.ParseIDNameSet(ctx, dst.GetFloatingSubnet(), diags),
		SiteNetworkSubnet: idname.ParseIDNameSet(ctx, dst.GetSiteNetworkSubnet(), diags),
	}
	ruleSourceObj, objDiags := types.ObjectValueFrom(ctx, socketlansource.SocketLanSourceAttrTypes, tfSocketLanSource)
	diags.Append(objDiags...)
	if diags.HasError() {
		return types.ObjectNull(socketlansource.SocketLanSourceAttrTypes)
	}

	return ruleSourceObj
}

// parseDestination parses the destination of a rule into a SocketLanDestination object
func (r *lfSubPolicyResource) parseDestination(ctx context.Context,
	dst cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Destination, diags *diag.Diagnostics,
) types.Object {
	tfSocketLanDestination := socketlandestination.SocketLanDestination{
		Vlan:              convert.ParseInt64List(ctx, dst.GetVlan(), diags),
		IP:                convert.ParseStringList(ctx, dst.GetIP(), diags),
		Subnet:            convert.ParseStringList(ctx, dst.GetSubnet(), diags),
		IPRange:           FromToList(ctx, dst.GetIPRange(), diags),
		Host:              idname.ParseIDNameSet(ctx, dst.GetHost(), diags),
		Group:             idname.ParseIDNameSet(ctx, dst.GetGroup(), diags),
		SystemGroup:       idname.ParseIDNameSet(ctx, dst.GetSystemGroup(), diags),
		NetworkInterface:  idname.ParseIDNameSet(ctx, dst.GetNetworkInterface(), diags),
		GlobalIPRange:     idname.ParseIDNameSet(ctx, dst.GetGlobalIPRange(), diags),
		FloatingSubnet:    idname.ParseIDNameSet(ctx, dst.GetFloatingSubnet(), diags),
		SiteNetworkSubnet: idname.ParseIDNameSet(ctx, dst.GetSiteNetworkSubnet(), diags),
	}
	ruleDestinationObj, objDiags := types.ObjectValueFrom(ctx, socketlandestination.SocketLanDestinationAttrTypes, tfSocketLanDestination)
	diags.Append(objDiags...)
	if diags.HasError() {
		return types.ObjectNull(socketlandestination.SocketLanDestinationAttrTypes)
	}

	return ruleDestinationObj
}

// FromToList parses a slice of ip ranges (from-to) into a types.List of FromToAttrTypes
func FromToList[T fromtoer](ctx context.Context, fts []T, diags *diag.Diagnostics) types.List {
	// null value
	if fts == nil {
		return types.ListNull(iprange.RangeObjectType)
	}

	// existing empty list
	if len(fts) == 0 {
		val, valueDiag := types.ListValue(iprange.RangeObjectType, nil)
		diags.Append(valueDiag...)
		return val
	}

	// make []FromToObject
	fromToSlice := make([]types.Object, 0, len(fts))
	for _, ft := range fts {
		tfFromTo := iprange.Range{
			From: types.StringValue(ft.GetFrom()),
			To:   types.StringValue(ft.GetTo()),
		}
		ftObj, objDiags := types.ObjectValueFrom(ctx, iprange.RangeAttrTypes, tfFromTo)
		diags.Append(objDiags...)
		fromToSlice = append(fromToSlice, ftObj)
	}
	// convert to types.List
	fromToList, valueDiag := types.ListValueFrom(ctx, iprange.RangeObjectType, fromToSlice)
	diags.Append(valueDiag...)

	return fromToList
}

// parseNat converts API NAT settings into a Terraform object.
func (r *lfSubPolicyResource) parseNat(ctx context.Context,
	nat cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Nat, diags *diag.Diagnostics,
) types.Object {
	tfNat := PolicyNatSettings{
		Enabled: types.BoolValue(nat.Enabled),
		NatType: types.StringValue(string(nat.NatType)),
	}
	natObj, objDiags := types.ObjectValueFrom(ctx, PolicyNatSettingsTypes, tfNat)
	diags.Append(objDiags...)
	if diags.HasError() {
		return types.ObjectNull(PolicyNatSettingsTypes)
	}

	return natObj
}

// parseService converts API custom and simple services into a Terraform object.
func (r *lfSubPolicyResource) parseService(ctx context.Context,
	svc cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Service, diags *diag.Diagnostics,
) types.Object {
	if len(svc.Custom) == 0 && len(svc.Simple) == 0 {
		return types.ObjectNull(PolicyServiceTypes)
	}
	tfSvc := PolicyService{
		Custom: r.ParseCustomService(ctx, svc.Custom, diags),
		Simple: r.parseSimpleService(ctx, svc.Simple, diags),
	}
	svcObj, objDiags := types.ObjectValueFrom(ctx, PolicyServiceTypes, tfSvc)
	diags.Append(objDiags...)
	if diags.HasError() {
		return types.ObjectNull(PolicyServiceTypes)
	}

	return svcObj
}

// parseSimpleService converts API simple services into a Terraform set.
func (r *lfSubPolicyResource) parseSimpleService(ctx context.Context,
	svcs []*cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Service_Simple, diags *diag.Diagnostics,
) types.Set {
	if len(svcs) == 0 {
		return types.SetNull(simpleservice.SimpleServiceObjectType)
	}
	serviceSlice := make([]types.Object, 0, len(svcs))
	for _, svc := range svcs {
		tfSimpleService := simpleservice.SimpleService{
			Name: types.StringValue(svc.GetName().String()),
		}
		simpleServiceObj, objDiags := types.ObjectValueFrom(ctx, simpleservice.SimpleServiceAttrTypes, tfSimpleService)
		diags.Append(objDiags...)
		serviceSlice = append(serviceSlice, simpleServiceObj)
	}
	serviceList, valueDiag := types.SetValueFrom(ctx, simpleservice.SimpleServiceObjectType, serviceSlice)
	diags.Append(valueDiag...)

	if diags.HasError() {
		return types.SetNull(types.ObjectType{AttrTypes: simpleservice.SimpleServiceAttrTypes})
	}

	return serviceList
}

// customservice.ParseCustomService converts API custom services into a Terraform list.
func (r *lfSubPolicyResource) ParseCustomService(ctx context.Context,
	svcs []*cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Service_Custom, diags *diag.Diagnostics,
) types.List {
	if len(svcs) == 0 {
		return types.ListNull(customservice.CustomServiceObjectType)
	}
	serviceSlice := make([]types.Object, 0, len(svcs))
	for _, svc := range svcs {
		portRangeObj := types.ObjectNull(portrange.PortRangeAttrTypes)
		if svc.PortRange != nil {
			tfPortRange := portrange.PortRange{
				From: types.StringPointerValue((*string)(svc.PortRange.GetFrom())),
				To:   types.StringPointerValue((*string)(svc.PortRange.GetTo())),
			}
			var objDiags diag.Diagnostics
			portRangeObj, objDiags = types.ObjectValueFrom(ctx, portrange.PortRangeAttrTypes, tfPortRange)
			diags.Append(objDiags...)
		}

		tfCustomService := PolicyCustomService{
			PortRange: portRangeObj,
			Port:      convert.ParseStringList(ctx, svc.Port, diags),
			Protocol:  types.StringValue(string(svc.Protocol)),
		}
		customServiceObj, objDiags := types.ObjectValueFrom(ctx, PolicyCustomServiceTypes, tfCustomService)
		diags.Append(objDiags...)
		serviceSlice = append(serviceSlice, customServiceObj)
	}
	serviceList, valueDiag := types.ListValueFrom(ctx, customservice.CustomServiceObjectType, serviceSlice)
	diags.Append(valueDiag...)

	if diags.HasError() {
		return types.ListNull(types.ObjectType{AttrTypes: PolicyCustomServiceTypes})
	}

	return serviceList
}

// parseSite converts API site and group references into a Terraform object.
func (r *lfSubPolicyResource) parseSite(ctx context.Context,
	site cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Site, diags *diag.Diagnostics,
) types.Object {
	tfSite := PolicySite{
		Group: idname.ParseIDNameSet(ctx, site.GetGroup(), diags),
		Site:  idname.ParseIDNameSet(ctx, site.GetSite(), diags),
	}
	siteObj, objDiags := types.ObjectValueFrom(ctx, PolicySiteTypes, tfSite)
	diags.Append(objDiags...)
	if diags.HasError() {
		return types.ObjectNull(PolicySiteTypes)
	}

	return siteObj
}

// convertCreateInput converts between the Terraform model and SDK values.
func convertCreateInput(
	ctx context.Context,
	r *lfSubPolicyResource,
	plan LanFirewallSubPolicy,
	diagnostics *diag.Diagnostics,
) (resultInput cato_models.SocketLanAddSubPolicyInput) {
	if !utils.HasValue(plan.At) {
		diagnostics.AddAttributeError(
			path.Root("at"),
			"Missing LAN sub-policy position",
			"The at attribute must be configured when creating a LAN sub-policy.",
		)
		return resultInput
	}

	resultInput = cato_models.SocketLanAddSubPolicyInput{
		At:     r.prepareAt(ctx, plan.At, diagnostics),
		Policy: r.preparePolicy(plan.Name, plan.Description),
		Scope:  r.prepareScope(ctx, plan.Scope, plan.Name, plan.Description, diagnostics),
	}
	return resultInput
}

// convertCreateInput2 converts between the Terraform model and SDK values.
func convertCreateInput2(
	plan *LanFirewallSubPolicy,
	subpol *cato_go_sdk.PolicySocketLanAddSubPolicy_Policy_SocketLan_AddSubPolicy,
	diagnostics *diag.Diagnostics,
) (resultPolicyID string) {
	if checkPolicyMutationStatus(subpol.GetStatus(),
		fmt.Sprintf("failed to add LAN sub-policy '%s'", plan.Name), diagnostics) {
		return resultPolicyID
	}

	resultPolicyID = subpol.GetPolicy().GetID()
	if resultPolicyID == "" {
		diagnostics.AddError("Cato API PolicySocketLanAddSubPolicy error", "policy ID is empty")
		return resultPolicyID
	}

	// Set the ID from the response
	plan.ID = types.StringValue(resultPolicyID)
	return resultPolicyID
}

// convertUpdateInput converts between the Terraform model and SDK values.
func convertUpdateInput(
	ctx context.Context,
	plan LanFirewallSubPolicy,
	diagnostics *diag.Diagnostics,
) (resultTfScope LanFirewallSubPolicyScope) {
	if utils.CheckErr(diagnostics, plan.Scope.As(ctx, &resultTfScope, basetypes.ObjectAsOptions{})) {
		return resultTfScope
	}
	return resultTfScope
}

// convertDeleteInput converts between the Terraform model and SDK values.
func convertDeleteInput(state LanFirewallSubPolicy) (resultInput cato_models.SocketLanRemoveSubPolicyInput) {
	resultInput = cato_models.SocketLanRemoveSubPolicyInput{
		Ref: &cato_models.SocketLanPolicyRefInput{
			By:    cato_models.ObjectRefByID,
			Input: state.ID.ValueString(),
		},
	}
	return resultInput
}

// convertHydrateLfSubPolicyValues converts between the Terraform model and SDK values.
func convertHydrateLfSubPolicyValues(
	ctx context.Context,
	r *lfSubPolicyResource,
	subPolicyID string,
	planOrState *LanFirewallSubPolicy,
	diags *diag.Diagnostics,
	priorState **LanFirewallSubPolicy,
	polRule *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules,
) {
	state := *priorState
	defer func() { *priorState = state }()
	apiRule := polRule.Rule
	state = &LanFirewallSubPolicy{
		ID:          types.StringValue(subPolicyID),
		Name:        types.StringValue(apiRule.Name),
		Description: types.StringValue(apiRule.Description),
		At:          planOrState.At,
		ScopeRuleID: types.StringValue(apiRule.ID),
		Scope:       r.parseRuleScope(ctx, apiRule, diags),
	}
}

// convertMoveValues converts between the Terraform model and SDK values.
func convertMoveValues(
	ctx context.Context,
	r *lfSubPolicyResource,
	scopeRuleID string,
	at types.Object,
	diags *diag.Diagnostics,
) (resultInput cato_models.PolicyMoveRuleInput) {
	resultInput = cato_models.PolicyMoveRuleInput{
		ID: scopeRuleID,
		To: r.prepareAt(ctx, at, diags),
	}
	return resultInput
}

func socketLanUpdateRuleInputFromTerraform(
	ctx context.Context,
	r *lfSubPolicyResource,
	diagnostics *diag.Diagnostics,
	plan LanFirewallSubPolicy,
	tfScope LanFirewallSubPolicyScope,
) cato_models.SocketLanUpdateRuleInput {
	return cato_models.SocketLanUpdateRuleInput{
		ID: plan.ScopeRuleID.ValueString(),
		Rule: &cato_models.SocketLanUpdateRuleDataInput{
			Description: plan.Description.ValueStringPointer(),
			Destination: r.prepareDestinationUpdate(ctx, tfScope.Destination, diagnostics),
			Direction:   (*cato_models.SocketLanDirection)(tfScope.Direction.ValueStringPointer()),
			Enabled:     tfScope.Enabled.ValueBoolPointer(),
			Name:        plan.Name.ValueStringPointer(),
			Nat:         r.prepareNatUpdate(ctx, tfScope.NAT, diagnostics),
			Service:     r.prepareServiceUpdate(ctx, tfScope.Service, diagnostics),
			Site:        r.prepareSiteUpdate(ctx, tfScope.Site, diagnostics),
			Source:      r.prepareSourceUpdate(ctx, tfScope.Source, diagnostics),
			Transport:   new(cato_models.SocketLanTransportTypeLan),
		},
	}
}
