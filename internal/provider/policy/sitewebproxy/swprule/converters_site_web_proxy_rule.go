package swprule

import (
	"context"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	cato_scalars "github.com/catonetworks/cato-go-sdk/scalars"
	attr "github.com/hashicorp/terraform-plugin-framework/attr"
	diag "github.com/hashicorp/terraform-plugin-framework/diag"
	types "github.com/hashicorp/terraform-plugin-framework/types"
	basetypes "github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	idname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	oldidname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/oldidname"
	policyrule "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
	refinput "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/refinput"
)

var swpAsOpts = basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}

func swpAuthConfig(
	ctx context.Context,
	obj types.Object,
	diags *diag.Diagnostics,
) *cato_models.SiteWebProxyAuthenticationConfigInput {
	out := &cato_models.SiteWebProxyAuthenticationConfigInput{
		KerberosConfig: &cato_models.SiteWebProxyKerberosConfigInput{IsEnabled: false},
	}
	if obj.IsNull() || obj.IsUnknown() {
		return out
	}
	ac := SiteWebProxyAuthenticationConfigPlan{}
	diags.Append(obj.As(ctx, &ac, swpAsOpts)...)
	if diags.HasError() {
		return out
	}
	if !ac.Method.IsNull() && !ac.Method.IsUnknown() && ac.Method.ValueString() != "" {
		m := cato_models.SiteWebProxyAuthMethod(ac.Method.ValueString())
		out.Method = &m
	}
	if !ac.KerberosConfig.IsNull() && !ac.KerberosConfig.IsUnknown() {
		kc := SiteWebProxyKerberosConfigPlan{}
		diags.Append(ac.KerberosConfig.As(ctx, &kc, swpAsOpts)...)
		if !diags.HasError() {
			out.KerberosConfig = &cato_models.SiteWebProxyKerberosConfigInput{
				IsEnabled: kc.IsEnabled.ValueBool(),
			}
			if !kc.EncryptedKeytab.IsNull() && !kc.EncryptedKeytab.IsUnknown() && kc.EncryptedKeytab.ValueString() != "" {
				out.KerberosConfig.EncryptedKeytab = kc.EncryptedKeytab.ValueStringPointer()
			}
		}
	}
	return out
}

func hydrateSiteWebProxyAddRuleInput(
	ctx context.Context,
	plan SiteWebProxyRule,
) (cato_models.SiteWebProxyAddRuleInput, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := cato_models.SiteWebProxyAddRuleInput{}

	if !plan.At.IsNull() {
		out.At = &cato_models.PolicyRulePositionInput{}
		pos := policyrule.Position{}
		diags.Append(plan.At.As(ctx, &pos, swpAsOpts)...)
		out.At.Position = (*cato_models.PolicyRulePositionEnum)(pos.Position.ValueStringPointer())
		out.At.Ref = pos.Ref.ValueStringPointer()
	}

	rule := SiteWebProxyRuleRulePlan{}
	diags.Append(plan.Rule.As(ctx, &rule, swpAsOpts)...)
	if diags.HasError() {
		return out, diags
	}

	data := &cato_models.SiteWebProxyAddRuleDataInput{
		Name:                        rule.Name.ValueString(),
		Description:                 rule.Description.ValueString(),
		Enabled:                     rule.Enabled.ValueBool(),
		Fqdn:                        rule.Fqdn.ValueString(),
		Port:                        cato_scalars.Port(rule.Port.ValueString()),
		ShouldAssociateWithAllSites: rule.ShouldAssociateWithAllSites.ValueBool(),
		AssociatedSite:              []*cato_models.SiteRefInput{},
		AuthenticationConfig:        swpAuthConfig(ctx, rule.AuthenticationConfig, &diags),
	}
	if !rule.AllowedDestinationProtocol.IsNull() && !rule.AllowedDestinationProtocol.IsUnknown() &&
		rule.AllowedDestinationProtocol.ValueString() != "" {
		p := cato_models.SiteWebProxyAllowedDestinationProtocol(rule.AllowedDestinationProtocol.ValueString())
		data.AllowedDestinationProtocol = &p
	}
	if s := refinput.FromSet(ctx, rule.AssociatedSite, &diags, "site", refinput.Site); s != nil {
		data.AssociatedSite = s
	}

	out.Rule = data
	return out, diags
}

func hydrateSiteWebProxyUpdateRuleInput(
	ctx context.Context,
	plan SiteWebProxyRule,
) (cato_models.SiteWebProxyUpdateRuleInput, diag.Diagnostics) {
	var diags diag.Diagnostics

	rule := SiteWebProxyRuleRulePlan{}
	diags.Append(plan.Rule.As(ctx, &rule, swpAsOpts)...)
	if diags.HasError() {
		return cato_models.SiteWebProxyUpdateRuleInput{}, diags
	}

	ac := swpAuthConfig(ctx, rule.AuthenticationConfig, &diags)
	port := cato_scalars.Port(rule.Port.ValueString())
	upd := cato_models.SiteWebProxyUpdateRuleInput{
		ID: rule.ID.ValueString(),
		Rule: &cato_models.SiteWebProxyUpdateRuleDataInput{
			Name:                        rule.Name.ValueStringPointer(),
			Description:                 rule.Description.ValueStringPointer(),
			Enabled:                     rule.Enabled.ValueBoolPointer(),
			Fqdn:                        rule.Fqdn.ValueStringPointer(),
			Port:                        &port,
			ShouldAssociateWithAllSites: rule.ShouldAssociateWithAllSites.ValueBoolPointer(),
			AssociatedSite:              []*cato_models.SiteRefInput{},
			AuthenticationConfig: &cato_models.SiteWebProxyAuthenticationConfigUpdateInput{
				Method: ac.Method,
			},
		},
	}
	if ac.KerberosConfig != nil {
		upd.Rule.AuthenticationConfig.KerberosConfig = &cato_models.SiteWebProxyKerberosConfigUpdateInput{
			IsEnabled:       &ac.KerberosConfig.IsEnabled,
			EncryptedKeytab: ac.KerberosConfig.EncryptedKeytab,
		}
	}
	if !rule.AllowedDestinationProtocol.IsNull() && !rule.AllowedDestinationProtocol.IsUnknown() &&
		rule.AllowedDestinationProtocol.ValueString() != "" {
		upd.Rule.AllowedDestinationProtocol = (*cato_models.SiteWebProxyAllowedDestinationProtocol)(
			rule.AllowedDestinationProtocol.ValueStringPointer(),
		)
	}
	if s := refinput.FromSet(ctx, rule.AssociatedSite, &diags, "site", refinput.Site); s != nil {
		upd.Rule.AssociatedSite = s
	}

	return upd, diags
}

// convertCreateInput converts between the Terraform model and SDK values.
func convertCreateInput(
	ctx context.Context,
	priorRulePlan *SiteWebProxyRuleRulePlan,
	planRule SiteWebProxyRuleRulePlan,
) (resultRuleObj basetypes.ObjectValue, resultOdiags diag.Diagnostics) {
	rulePlan := *priorRulePlan
	defer func() { *priorRulePlan = rulePlan }()
	// The API never returns the encrypted keytab - keep the plan's auth config.
	if !planRule.AuthenticationConfig.IsNull() && !planRule.AuthenticationConfig.IsUnknown() {
		rulePlan.AuthenticationConfig = planRule.AuthenticationConfig
	}
	resultRuleObj, resultOdiags = types.ObjectValueFrom(ctx, SiteWebProxyRuleRuleAttrTypes, rulePlan)
	return resultRuleObj, resultOdiags
}

// convertCreateInput2 converts between the Terraform model and SDK values.
func convertCreateInput2(plan *SiteWebProxyRule, newID string, ruleObj basetypes.ObjectValue) {
	plan.ID = types.StringValue(newID)
	plan.Rule = ruleObj
}

// convertReadState converts between the Terraform model and SDK values.
func convertReadState(
	ctx context.Context,
	rulePlan SiteWebProxyRuleRulePlan,
	priorHydrated *SiteWebProxyRuleRulePlan,
) (resultRuleObj basetypes.ObjectValue, resultOdiags diag.Diagnostics) {
	hydrated := *priorHydrated
	defer func() { *priorHydrated = hydrated }()
	// The API never returns the encrypted keytab - keep state's auth config.
	if !rulePlan.AuthenticationConfig.IsNull() && !rulePlan.AuthenticationConfig.IsUnknown() {
		hydrated.AuthenticationConfig = rulePlan.AuthenticationConfig
	}
	resultRuleObj, resultOdiags = types.ObjectValueFrom(ctx, SiteWebProxyRuleRuleAttrTypes, hydrated)
	return resultRuleObj, resultOdiags
}

// convertReadState2 converts between the Terraform model and SDK values.
func convertReadState2(
	state *SiteWebProxyRule,
	cur *cato_go_sdk.PolicySiteWebProxyPolicy_Policy_SiteWebProxy_Policy_Rules_Rule,
	ruleObj basetypes.ObjectValue,
	diagnostics *diag.Diagnostics,
) {
	if state.At.IsNull() || state.At.IsUnknown() {
		atObj, d := types.ObjectValue(policyrule.PositionAttrTypes, map[string]attr.Value{
			"position": types.StringValue("LAST_IN_POLICY"),
			"ref":      types.StringNull(),
		})
		diagnostics.Append(d...)
		state.At = atObj
	}

	state.ID = types.StringValue(cur.GetID())
	state.Rule = ruleObj
}

// convertUpdateInput converts between the Terraform model and SDK values.
func convertUpdateInput(
	ctx context.Context,
	plan SiteWebProxyRule,
	diagnostics *diag.Diagnostics,
) (
	resultMove cato_models.PolicyMoveRuleInput,
	resultAtAsOpts basetypes.ObjectAsOptions,
) {
	resultMove = cato_models.PolicyMoveRuleInput{}
	resultAtAsOpts = basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}
	if !plan.At.IsNull() {
		pos := policyrule.Position{}
		diagnostics.Append(plan.At.As(ctx, &pos, resultAtAsOpts)...)
		resultMove.To = &cato_models.PolicyRulePositionInput{
			Position: (*cato_models.PolicyRulePositionEnum)(pos.Position.ValueStringPointer()),
			Ref:      pos.Ref.ValueStringPointer(),
		}
	}
	return resultMove, resultAtAsOpts
}

// convertUpdateInput2 converts between the Terraform model and SDK values.
func convertUpdateInput2(
	ctx context.Context,
	rule SiteWebProxyRuleRulePlan,
	priorRulePlan *SiteWebProxyRuleRulePlan,
) (resultRuleObj basetypes.ObjectValue, resultOdiags diag.Diagnostics) {
	rulePlan := *priorRulePlan
	defer func() { *priorRulePlan = rulePlan }()
	// The API never returns the encrypted keytab - keep the plan's auth config.
	if !rule.AuthenticationConfig.IsNull() && !rule.AuthenticationConfig.IsUnknown() {
		rulePlan.AuthenticationConfig = rule.AuthenticationConfig
	}
	resultRuleObj, resultOdiags = types.ObjectValueFrom(ctx, SiteWebProxyRuleRuleAttrTypes, rulePlan)
	return resultRuleObj, resultOdiags
}

// convertUpdateInput3 converts between the Terraform model and SDK values.
func convertUpdateInput3(pos policyrule.Position) (resultAtObj basetypes.ObjectValue, resultAdiags diag.Diagnostics) {
	resultAtObj, resultAdiags = types.ObjectValue(policyrule.PositionAttrTypes, map[string]attr.Value{
		"position": pos.Position,
		"ref":      pos.Ref,
	})
	return resultAtObj, resultAdiags
}

// convertUpdateInput4 converts between the Terraform model and SDK values.
func convertUpdateInput4(
	rule SiteWebProxyRuleRulePlan,
	ruleObj basetypes.ObjectValue,
	atObj basetypes.ObjectValue,
) (resultSt SiteWebProxyRule) {
	resultSt = SiteWebProxyRule{
		ID:   types.StringValue(rule.ID.ValueString()),
		At:   atObj,
		Rule: ruleObj,
	}
	return resultSt
}

// convertDeleteInput converts between the Terraform model and SDK values.
func convertDeleteInput(rule SiteWebProxyRuleRulePlan) (resultRm cato_models.SiteWebProxyRemoveRuleInput) {
	resultRm = cato_models.SiteWebProxyRemoveRuleInput{ID: rule.ID.ValueString()}
	return resultRm
}

// hydrateSiteWebProxyRuleStateFromClient maps a policy read rule into Terraform nested objects.
func hydrateSiteWebProxyRuleStateFromClient(
	ctx context.Context,
	r *cato_go_sdk.PolicySiteWebProxyPolicy_Policy_SiteWebProxy_Policy_Rules_Rule,
) (SiteWebProxyRuleRulePlan, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := SiteWebProxyRuleRulePlan{}
	if r == nil {
		return out, diags
	}

	out.AssociatedSite = types.SetNull(idname.RefObjectType)
	out.AuthenticationConfig = types.ObjectNull(SiteWebProxyAuthenticationConfigAttrTypes)

	out.ID = types.StringValue(r.GetID())
	out.Name = types.StringValue(r.GetName())
	if d := r.GetDescription(); d != "" {
		out.Description = types.StringValue(d)
	} else {
		out.Description = types.StringNull()
	}
	out.Enabled = types.BoolValue(r.GetEnabled())
	out.Fqdn = types.StringValue(r.GetFqdn())
	if p := r.GetPort(); p != nil {
		out.Port = types.StringValue(string(*p))
	} else {
		out.Port = types.StringNull()
	}
	out.ShouldAssociateWithAllSites = types.BoolValue(r.GetShouldAssociateWithAllSites())
	if p := r.GetAllowedDestinationProtocol(); p != nil {
		out.AllowedDestinationProtocol = types.StringValue(string(*p))
	} else {
		out.AllowedDestinationProtocol = types.StringNull()
	}

	if sites := r.GetAssociatedSite(); len(sites) > 0 {
		out.AssociatedSite = oldidname.ParseNameIDList(ctx, sites, "rule.associated_site")
	}

	if ac := r.GetAuthenticationConfig(); ac != nil {
		method := types.StringNull()
		if ac.GetMethod() != nil {
			method = types.StringValue(string(*ac.GetMethod()))
		}
		kcObj := types.ObjectNull(SiteWebProxyKerberosConfigAttrTypes)
		if kc := ac.GetKerberosConfig(); kc != nil {
			o, d := types.ObjectValue(SiteWebProxyKerberosConfigAttrTypes, map[string]attr.Value{
				"is_enabled":       types.BoolValue(kc.GetIsEnabled()),
				"encrypted_keytab": types.StringNull(),
			})
			diags.Append(d...)
			kcObj = o
		}
		acObj, d := types.ObjectValue(SiteWebProxyAuthenticationConfigAttrTypes, map[string]attr.Value{
			"method":          method,
			"kerberos_config": kcObj,
		})
		diags.Append(d...)
		out.AuthenticationConfig = acObj
	}

	return out, diags
}
