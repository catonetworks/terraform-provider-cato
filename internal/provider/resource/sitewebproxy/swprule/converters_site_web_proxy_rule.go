package swprule

import (
	"context"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/cato-go-sdk/scalars"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	policyshared "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policy"
)

func buildCreateObject(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SiteWebProxyAddRuleDataInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule must be known before applying.")
		return nil
	}
	var model RuleModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	policyshared.ValidateRuleValues(diags,
		policyshared.KnownValue{Name: "rule.allowed_destination_protocol",
			Value:    model.AllowedDestinationProtocol,
			Required: true},
		policyshared.KnownValue{Name: "rule.associated_site",
			Value:    model.AssociatedSite,
			Required: true},
		policyshared.KnownValue{Name: "rule.authentication_config",
			Value:    model.AuthenticationConfig,
			Required: true},
		policyshared.KnownValue{Name: "rule.description",
			Value:    model.Description,
			Required: true},
		policyshared.KnownValue{Name: "rule.enabled",
			Value:    model.Enabled,
			Required: true},
		policyshared.KnownValue{Name: "rule.fqdn",
			Value:    model.Fqdn,
			Required: true},
		policyshared.KnownValue{Name: "rule.name",
			Value:    model.Name,
			Required: true},
		policyshared.KnownValue{Name: "rule.port",
			Value:    model.Port,
			Required: true},
		policyshared.KnownValue{Name: "rule.should_associate_with_all_sites",
			Value:    model.ShouldAssociateWithAllSites,
			Required: true})
	if diags.HasError() {
		return nil
	}
	validateProxyRule(ctx, model, diags)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SiteWebProxyAddRuleDataInput{}
	vAllowedDestinationProtocol := cato_models.SiteWebProxyAllowedDestinationProtocol(model.AllowedDestinationProtocol.ValueString())
	input.AllowedDestinationProtocol = &vAllowedDestinationProtocol
	input.AssociatedSite = idname.PrepareIDNameSet[cato_models.SiteRefInput](ctx, model.AssociatedSite, diags)
	input.AuthenticationConfig = buildCreateObjectAuthenticationConfig(ctx, model.AuthenticationConfig, diags)
	input.Description = model.Description.ValueString()
	input.Enabled = model.Enabled.ValueBool()
	input.Fqdn = model.Fqdn.ValueString()
	input.Name = model.Name.ValueString()
	input.Port = scalars.Port(model.Port.ValueString())
	input.ShouldAssociateWithAllSites = model.ShouldAssociateWithAllSites.ValueBool()
	return input
}

func buildCreateObjectAuthenticationConfig(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SiteWebProxyAuthenticationConfigInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.authentication_config must be known before applying.")
		return nil
	}
	var model RuleAuthenticationConfigModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	policyshared.ValidateRuleValues(diags,
		policyshared.KnownValue{Name: "rule.authentication_config.kerberos_config",
			Value:    model.KerberosConfig,
			Required: false},
		policyshared.KnownValue{Name: "rule.authentication_config.method",
			Value:    model.Method,
			Required: true})
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SiteWebProxyAuthenticationConfigInput{}
	input.KerberosConfig = buildCreateObjectAuthenticationConfigKerberosConfig(ctx, model.KerberosConfig, diags)
	vMethod := cato_models.SiteWebProxyAuthMethod(model.Method.ValueString())
	input.Method = &vMethod
	return input
}

func buildCreateObjectAuthenticationConfigKerberosConfig(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SiteWebProxyKerberosConfigInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.authentication_config.kerberos_config must be known before applying.")
		return nil
	}
	var model RuleAuthenticationConfigKerberosConfigModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	policyshared.ValidateRuleValues(diags,
		policyshared.KnownValue{Name: "rule.authentication_config.kerberos_config.encrypted_keytab",
			Value:    model.EncryptedKeytab,
			Required: false},
		policyshared.KnownValue{Name: "rule.authentication_config.kerberos_config.is_enabled",
			Value:    model.IsEnabled,
			Required: true})
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SiteWebProxyKerberosConfigInput{}
	if !model.EncryptedKeytab.IsNull() {
		vEncryptedKeytab := model.EncryptedKeytab.ValueString()
		input.EncryptedKeytab = &vEncryptedKeytab
	}
	input.IsEnabled = model.IsEnabled.ValueBool()
	return input
}

func buildUpdateObject(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SiteWebProxyUpdateRuleDataInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule must be known before applying.")
		return nil
	}
	var model RuleModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	policyshared.ValidateRuleValues(diags,
		policyshared.KnownValue{Name: "rule.allowed_destination_protocol",
			Value:    model.AllowedDestinationProtocol,
			Required: true},
		policyshared.KnownValue{Name: "rule.associated_site",
			Value:    model.AssociatedSite,
			Required: true},
		policyshared.KnownValue{Name: "rule.authentication_config",
			Value:    model.AuthenticationConfig,
			Required: true},
		policyshared.KnownValue{Name: "rule.description",
			Value:    model.Description,
			Required: true},
		policyshared.KnownValue{Name: "rule.enabled",
			Value:    model.Enabled,
			Required: true},
		policyshared.KnownValue{Name: "rule.fqdn",
			Value:    model.Fqdn,
			Required: true},
		policyshared.KnownValue{Name: "rule.name",
			Value:    model.Name,
			Required: true},
		policyshared.KnownValue{Name: "rule.port",
			Value:    model.Port,
			Required: true},
		policyshared.KnownValue{Name: "rule.should_associate_with_all_sites",
			Value:    model.ShouldAssociateWithAllSites,
			Required: true})
	if diags.HasError() {
		return nil
	}
	validateProxyRule(ctx, model, diags)
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SiteWebProxyUpdateRuleDataInput{}
	vAllowedDestinationProtocol := cato_models.SiteWebProxyAllowedDestinationProtocol(model.AllowedDestinationProtocol.ValueString())
	input.AllowedDestinationProtocol = &vAllowedDestinationProtocol
	input.AssociatedSite = idname.PrepareIDNameSet[cato_models.SiteRefInput](ctx, model.AssociatedSite, diags)
	input.AuthenticationConfig = buildUpdateObjectAuthenticationConfig(ctx, model.AuthenticationConfig, diags)
	vDescription := model.Description.ValueString()
	input.Description = &vDescription
	vEnabled := model.Enabled.ValueBool()
	input.Enabled = &vEnabled
	vFqdn := model.Fqdn.ValueString()
	input.Fqdn = &vFqdn
	vName := model.Name.ValueString()
	input.Name = &vName
	vPort := scalars.Port(model.Port.ValueString())
	input.Port = &vPort
	vShouldAssociateWithAllSites := model.ShouldAssociateWithAllSites.ValueBool()
	input.ShouldAssociateWithAllSites = &vShouldAssociateWithAllSites
	return input
}

func buildUpdateObjectAuthenticationConfig(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SiteWebProxyAuthenticationConfigUpdateInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.authentication_config must be known before applying.")
		return nil
	}
	var model RuleAuthenticationConfigModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	policyshared.ValidateRuleValues(diags,
		policyshared.KnownValue{Name: "rule.authentication_config.kerberos_config",
			Value:    model.KerberosConfig,
			Required: false},
		policyshared.KnownValue{Name: "rule.authentication_config.method",
			Value:    model.Method,
			Required: true})
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SiteWebProxyAuthenticationConfigUpdateInput{}
	input.KerberosConfig = buildUpdateObjectAuthenticationConfigKerberosConfig(ctx, model.KerberosConfig, diags)
	if input.KerberosConfig == nil {
		disabled,
			empty := false,
			""
		input.KerberosConfig = &cato_models.SiteWebProxyKerberosConfigUpdateInput{IsEnabled: &disabled,
			EncryptedKeytab: &empty}
	}
	vMethod := cato_models.SiteWebProxyAuthMethod(model.Method.ValueString())
	input.Method = &vMethod
	return input
}

func buildUpdateObjectAuthenticationConfigKerberosConfig(
	ctx context.Context, value types.Object, diags *diag.Diagnostics,
) *cato_models.SiteWebProxyKerberosConfigUpdateInput {
	if value.IsNull() {
		return nil
	}
	if value.IsUnknown() {
		diags.AddError("Unknown configuration", "rule.authentication_config.kerberos_config must be known before applying.")
		return nil
	}
	var model RuleAuthenticationConfigKerberosConfigModel
	diags.Append(value.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	policyshared.ValidateRuleValues(diags,
		policyshared.KnownValue{Name: "rule.authentication_config.kerberos_config.encrypted_keytab",
			Value:    model.EncryptedKeytab,
			Required: false},
		policyshared.KnownValue{Name: "rule.authentication_config.kerberos_config.is_enabled",
			Value:    model.IsEnabled,
			Required: true})
	if diags.HasError() {
		return nil
	}
	input := &cato_models.SiteWebProxyKerberosConfigUpdateInput{}
	if !model.EncryptedKeytab.IsNull() {
		vEncryptedKeytab := model.EncryptedKeytab.ValueString()
		input.EncryptedKeytab = &vEncryptedKeytab
	}
	vIsEnabled := model.IsEnabled.ValueBool()
	input.IsEnabled = &vIsEnabled
	return input
}

type read = cato.PolicySiteWebProxyPolicy_Policy_SiteWebProxy_Policy_Rules_Rule

func parseRead(ctx context.Context, value *read, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule object.")
		return types.ObjectNull(ruleAttrTypes)
	}
	model := RuleModel{}
	if value.AllowedDestinationProtocol == nil {
		diags.AddError("Invalid API response", "Missing rule.allowed_destination_protocol")
		return types.ObjectNull(ruleAttrTypes)
	}
	model.AllowedDestinationProtocol = types.StringValue(string(*value.AllowedDestinationProtocol))
	model.AssociatedSite = idname.ParseIDNameSet(ctx, value.AssociatedSite, diags)
	if model.AssociatedSite.IsNull() {
		model.AssociatedSite = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.AuthenticationConfig = parseReadAuthenticationConfig(ctx, &value.AuthenticationConfig, diags)
	model.Description = types.StringValue(value.Description)
	model.Enabled = types.BoolValue(value.Enabled)
	model.Fqdn = types.StringValue(value.Fqdn)
	model.ID = types.StringValue(value.ID)
	model.Index = types.Int64Value(value.Index)
	model.Name = types.StringValue(value.Name)
	model.Port = types.StringValue(string(value.Port))
	model.Section = parseReadSection(ctx, &value.Section, diags)
	model.ShouldAssociateWithAllSites = types.BoolValue(value.ShouldAssociateWithAllSites)
	result, d := types.ObjectValueFrom(ctx, ruleAttrTypes, model)
	diags.Append(d...)
	return result
}

type readAuth = cato.PolicySiteWebProxyPolicy_Policy_SiteWebProxy_Policy_Rules_Rule_AuthenticationConfig

func parseReadAuthenticationConfig(ctx context.Context,
	value *readAuth,
	diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.authentication_config object.")
		return types.ObjectNull(ruleAuthenticationConfigAttrTypes)
	}
	model := RuleAuthenticationConfigModel{}
	if value.KerberosConfig == nil || emptyKerberos(value.KerberosConfig) {
		model.KerberosConfig = types.ObjectNull(ruleAuthenticationConfigKerberosConfigAttrTypes)
	} else {
		model.KerberosConfig = parseReadAuthenticationConfigKerberosConfig(ctx, value.KerberosConfig, diags)
	}
	if value.Method == nil {
		diags.AddError("Invalid API response", "Missing rule.authentication_config.method")
		return types.ObjectNull(ruleAuthenticationConfigAttrTypes)
	}
	model.Method = types.StringValue(string(*value.Method))
	result, d := types.ObjectValueFrom(ctx, ruleAuthenticationConfigAttrTypes, model)
	diags.Append(d...)
	return result
}

type readAuthKerberos = cato.PolicySiteWebProxyPolicy_Policy_SiteWebProxy_Policy_Rules_Rule_AuthenticationConfig_KerberosConfig

func parseReadAuthenticationConfigKerberosConfig(ctx context.Context,
	value *readAuthKerberos,
	diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.authentication_config.kerberos_config object.")
		return types.ObjectNull(ruleAuthenticationConfigKerberosConfigAttrTypes)
	}
	model := RuleAuthenticationConfigKerberosConfigModel{}
	if value.EncryptedKeytab == nil {
		model.EncryptedKeytab = types.StringValue("")
	} else {
		model.EncryptedKeytab = types.StringValue(*value.EncryptedKeytab)
	}
	model.IsEnabled = types.BoolValue(value.IsEnabled)
	result, d := types.ObjectValueFrom(ctx, ruleAuthenticationConfigKerberosConfigAttrTypes, model)
	diags.Append(d...)
	return result
}

type readSection = cato.PolicySiteWebProxyPolicy_Policy_SiteWebProxy_Policy_Rules_Rule_Section

func parseReadSection(ctx context.Context, value *readSection, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.section object.")
		return types.ObjectNull(ruleSectionAttrTypes)
	}
	return idname.ParseIDName(ctx, *value, diags)
}

type add = cato.PolicySiteWebProxyAddRule_Policy_SiteWebProxy_AddRule_Rule_Rule

func parseAdd(ctx context.Context, value *add, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule object.")
		return types.ObjectNull(ruleAttrTypes)
	}
	model := RuleModel{}
	if value.AllowedDestinationProtocol == nil {
		diags.AddError("Invalid API response", "Missing rule.allowed_destination_protocol")
		return types.ObjectNull(ruleAttrTypes)
	}
	model.AllowedDestinationProtocol = types.StringValue(string(*value.AllowedDestinationProtocol))
	model.AssociatedSite = idname.ParseIDNameSet(ctx, value.AssociatedSite, diags)
	if model.AssociatedSite.IsNull() {
		model.AssociatedSite = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.AuthenticationConfig = parseAddAuthenticationConfig(ctx, &value.AuthenticationConfig, diags)
	model.Description = types.StringValue(value.Description)
	model.Enabled = types.BoolValue(value.Enabled)
	model.Fqdn = types.StringValue(value.Fqdn)
	model.ID = types.StringValue(value.ID)
	model.Index = types.Int64Value(value.Index)
	model.Name = types.StringValue(value.Name)
	model.Port = types.StringValue(string(value.Port))
	model.Section = parseAddSection(ctx, &value.Section, diags)
	model.ShouldAssociateWithAllSites = types.BoolValue(value.ShouldAssociateWithAllSites)
	result, d := types.ObjectValueFrom(ctx, ruleAttrTypes, model)
	diags.Append(d...)
	return result
}

type addAuth = cato.PolicySiteWebProxyAddRule_Policy_SiteWebProxy_AddRule_Rule_Rule_AuthenticationConfig

func parseAddAuthenticationConfig(ctx context.Context, value *addAuth, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.authentication_config object.")
		return types.ObjectNull(ruleAuthenticationConfigAttrTypes)
	}
	model := RuleAuthenticationConfigModel{}
	if value.KerberosConfig == nil || emptyKerberos(value.KerberosConfig) {
		model.KerberosConfig = types.ObjectNull(ruleAuthenticationConfigKerberosConfigAttrTypes)
	} else {
		model.KerberosConfig = parseAddAuthenticationConfigKerberosConfig(ctx, value.KerberosConfig, diags)
	}
	if value.Method == nil {
		diags.AddError("Invalid API response", "Missing rule.authentication_config.method")
		return types.ObjectNull(ruleAuthenticationConfigAttrTypes)
	}
	model.Method = types.StringValue(string(*value.Method))
	result, d := types.ObjectValueFrom(ctx, ruleAuthenticationConfigAttrTypes, model)
	diags.Append(d...)
	return result
}

type addAuthKerberos = cato.PolicySiteWebProxyAddRule_Policy_SiteWebProxy_AddRule_Rule_Rule_AuthenticationConfig_KerberosConfig

func parseAddAuthenticationConfigKerberosConfig(ctx context.Context,
	value *addAuthKerberos,
	diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.authentication_config.kerberos_config object.")
		return types.ObjectNull(ruleAuthenticationConfigKerberosConfigAttrTypes)
	}
	model := RuleAuthenticationConfigKerberosConfigModel{}
	if value.EncryptedKeytab == nil {
		model.EncryptedKeytab = types.StringValue("")
	} else {
		model.EncryptedKeytab = types.StringValue(*value.EncryptedKeytab)
	}
	model.IsEnabled = types.BoolValue(value.IsEnabled)
	result, d := types.ObjectValueFrom(ctx, ruleAuthenticationConfigKerberosConfigAttrTypes, model)
	diags.Append(d...)
	return result
}

type addSection = cato.PolicySiteWebProxyAddRule_Policy_SiteWebProxy_AddRule_Rule_Rule_Section

func parseAddSection(ctx context.Context, value *addSection, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.section object.")
		return types.ObjectNull(ruleSectionAttrTypes)
	}
	return idname.ParseIDName(ctx, *value, diags)
}

type update = cato.PolicySiteWebProxyUpdateRule_Policy_SiteWebProxy_UpdateRule_Rule_Rule

func parseUpdate(ctx context.Context, value *update, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule object.")
		return types.ObjectNull(ruleAttrTypes)
	}
	model := RuleModel{}
	if value.AllowedDestinationProtocol == nil {
		diags.AddError("Invalid API response", "Missing rule.allowed_destination_protocol")
		return types.ObjectNull(ruleAttrTypes)
	}
	model.AllowedDestinationProtocol = types.StringValue(string(*value.AllowedDestinationProtocol))
	model.AssociatedSite = idname.ParseIDNameSet(ctx, value.AssociatedSite, diags)
	if model.AssociatedSite.IsNull() {
		model.AssociatedSite = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.AuthenticationConfig = parseUpdateAuthenticationConfig(ctx, &value.AuthenticationConfig, diags)
	model.Description = types.StringValue(value.Description)
	model.Enabled = types.BoolValue(value.Enabled)
	model.Fqdn = types.StringValue(value.Fqdn)
	model.ID = types.StringValue(value.ID)
	model.Index = types.Int64Value(value.Index)
	model.Name = types.StringValue(value.Name)
	model.Port = types.StringValue(string(value.Port))
	model.Section = parseUpdateSection(ctx, &value.Section, diags)
	model.ShouldAssociateWithAllSites = types.BoolValue(value.ShouldAssociateWithAllSites)
	result, d := types.ObjectValueFrom(ctx, ruleAttrTypes, model)
	diags.Append(d...)
	return result
}

type updateAuth = cato.PolicySiteWebProxyUpdateRule_Policy_SiteWebProxy_UpdateRule_Rule_Rule_AuthenticationConfig

func parseUpdateAuthenticationConfig(ctx context.Context,
	value *updateAuth,
	diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.authentication_config object.")
		return types.ObjectNull(ruleAuthenticationConfigAttrTypes)
	}
	model := RuleAuthenticationConfigModel{}
	if value.KerberosConfig == nil || emptyKerberos(value.KerberosConfig) {
		model.KerberosConfig = types.ObjectNull(ruleAuthenticationConfigKerberosConfigAttrTypes)
	} else {
		model.KerberosConfig = parseUpdateAuthenticationConfigKerberosConfig(ctx, value.KerberosConfig, diags)
	}
	if value.Method == nil {
		diags.AddError("Invalid API response", "Missing rule.authentication_config.method")
		return types.ObjectNull(ruleAuthenticationConfigAttrTypes)
	}
	model.Method = types.StringValue(string(*value.Method))
	result, d := types.ObjectValueFrom(ctx, ruleAuthenticationConfigAttrTypes, model)
	diags.Append(d...)
	return result
}

type updateAuthKerberos = cato.PolicySiteWebProxyUpdateRule_Policy_SiteWebProxy_UpdateRule_Rule_Rule_AuthenticationConfig_KerberosConfig

func parseUpdateAuthenticationConfigKerberosConfig(ctx context.Context,
	value *updateAuthKerberos,
	diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.authentication_config.kerberos_config object.")
		return types.ObjectNull(ruleAuthenticationConfigKerberosConfigAttrTypes)
	}
	model := RuleAuthenticationConfigKerberosConfigModel{}
	if value.EncryptedKeytab == nil {
		model.EncryptedKeytab = types.StringValue("")
	} else {
		model.EncryptedKeytab = types.StringValue(*value.EncryptedKeytab)
	}
	model.IsEnabled = types.BoolValue(value.IsEnabled)
	result, d := types.ObjectValueFrom(ctx, ruleAuthenticationConfigKerberosConfigAttrTypes, model)
	diags.Append(d...)
	return result
}

type updateSection = cato.PolicySiteWebProxyUpdateRule_Policy_SiteWebProxy_UpdateRule_Rule_Rule_Section

func parseUpdateSection(ctx context.Context, value *updateSection, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.section object.")
		return types.ObjectNull(ruleSectionAttrTypes)
	}
	return idname.ParseIDName(ctx, *value, diags)
}

type move = cato.PolicySiteWebProxyMoveRule_Policy_SiteWebProxy_MoveRule_Rule_Rule

func parseMove(ctx context.Context, value *move, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule object.")
		return types.ObjectNull(ruleAttrTypes)
	}
	model := RuleModel{}
	if value.AllowedDestinationProtocol == nil {
		diags.AddError("Invalid API response", "Missing rule.allowed_destination_protocol")
		return types.ObjectNull(ruleAttrTypes)
	}
	model.AllowedDestinationProtocol = types.StringValue(string(*value.AllowedDestinationProtocol))
	model.AssociatedSite = idname.ParseIDNameSet(ctx, value.AssociatedSite, diags)
	if model.AssociatedSite.IsNull() {
		model.AssociatedSite = types.SetValueMust(idname.RefObjectType, nil)
	}
	model.AuthenticationConfig = parseMoveAuthenticationConfig(ctx, &value.AuthenticationConfig, diags)
	model.Description = types.StringValue(value.Description)
	model.Enabled = types.BoolValue(value.Enabled)
	model.Fqdn = types.StringValue(value.Fqdn)
	model.ID = types.StringValue(value.ID)
	model.Index = types.Int64Value(value.Index)
	model.Name = types.StringValue(value.Name)
	model.Port = types.StringValue(string(value.Port))
	model.Section = parseMoveSection(ctx, &value.Section, diags)
	model.ShouldAssociateWithAllSites = types.BoolValue(value.ShouldAssociateWithAllSites)
	result, d := types.ObjectValueFrom(ctx, ruleAttrTypes, model)
	diags.Append(d...)
	return result
}

type moveAuth = cato.PolicySiteWebProxyMoveRule_Policy_SiteWebProxy_MoveRule_Rule_Rule_AuthenticationConfig

func parseMoveAuthenticationConfig(ctx context.Context,
	value *moveAuth,
	diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.authentication_config object.")
		return types.ObjectNull(ruleAuthenticationConfigAttrTypes)
	}
	model := RuleAuthenticationConfigModel{}
	if value.KerberosConfig == nil || emptyKerberos(value.KerberosConfig) {
		model.KerberosConfig = types.ObjectNull(ruleAuthenticationConfigKerberosConfigAttrTypes)
	} else {
		model.KerberosConfig = parseMoveAuthenticationConfigKerberosConfig(ctx, value.KerberosConfig, diags)
	}
	if value.Method == nil {
		diags.AddError("Invalid API response", "Missing rule.authentication_config.method")
		return types.ObjectNull(ruleAuthenticationConfigAttrTypes)
	}
	model.Method = types.StringValue(string(*value.Method))
	result, d := types.ObjectValueFrom(ctx, ruleAuthenticationConfigAttrTypes, model)
	diags.Append(d...)
	return result
}

type moveAuthKerberos = cato.PolicySiteWebProxyMoveRule_Policy_SiteWebProxy_MoveRule_Rule_Rule_AuthenticationConfig_KerberosConfig

func parseMoveAuthenticationConfigKerberosConfig(ctx context.Context,
	value *moveAuthKerberos,
	diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.authentication_config.kerberos_config object.")
		return types.ObjectNull(ruleAuthenticationConfigKerberosConfigAttrTypes)
	}
	model := RuleAuthenticationConfigKerberosConfigModel{}
	if value.EncryptedKeytab == nil {
		model.EncryptedKeytab = types.StringValue("")
	} else {
		model.EncryptedKeytab = types.StringValue(*value.EncryptedKeytab)
	}
	model.IsEnabled = types.BoolValue(value.IsEnabled)
	result, d := types.ObjectValueFrom(ctx, ruleAuthenticationConfigKerberosConfigAttrTypes, model)
	diags.Append(d...)
	return result
}

type moveSection = cato.PolicySiteWebProxyMoveRule_Policy_SiteWebProxy_MoveRule_Rule_Rule_Section

func parseMoveSection(ctx context.Context, value *moveSection, diags *diag.Diagnostics) types.Object {
	if value == nil {
		diags.AddError("Invalid API response", "Missing rule.section object.")
		return types.ObjectNull(ruleSectionAttrTypes)
	}
	return idname.ParseIDName(ctx, *value, diags)
}

// API defaults are equivalent to an omitted optional block. Nondefault remote values remain visible as drift.
func emptyKerberos(value interface {
	GetIsEnabled() bool
	GetEncryptedKeytab() *string
}) bool {
	key := value.GetEncryptedKeytab()
	return !value.GetIsEnabled() &&
		(key == nil ||
			*key == "")
}
