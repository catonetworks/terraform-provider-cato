package provider

import (
	"context"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func hydrateTlsRuleState(ctx context.Context, state TlsInspectionRule, currentRule *cato_go_sdk.Tlsinspectpolicy_Policy_TLSInspect_Policy_Rules_Rule) (Policy_Policy_TlsInspect_Policy_Rules_Rule, diag.Diagnostics) {

	ruleInput := Policy_Policy_TlsInspect_Policy_Rules_Rule{}
	diags := make(diag.Diagnostics, 0)

	// Handle case where state might be incomplete (e.g., during import)
	// Try to extract existing state first, but don't fail if it's incomplete
	if !state.Rule.IsNull() && !state.Rule.IsUnknown() {
		diagstmp := state.Rule.As(ctx, &ruleInput, basetypes.ObjectAsOptions{})
		if diagstmp.HasError() {
			// If conversion fails during import, log it but continue with empty ruleInput
			tflog.Debug(ctx, "Failed to convert state to struct in hydrateTlsRuleState, using empty struct for hydration")
			// Reset ruleInput to empty struct and clear diagnostics
			ruleInput = Policy_Policy_TlsInspect_Policy_Rules_Rule{}
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

	// Set index from API response when state index is null/unknown, otherwise preserve state index
	if ruleInput.Index.IsNull() || ruleInput.Index.IsUnknown() {
		ruleInput.Index = types.Int64Value(currentRule.Index)
	}
	// If ruleInput.Index has a value, we preserve it (no action needed)

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

	// //////////// Rule -> Source ///////////////
	curRuleSourceObj, diagstmp := types.ObjectValue(
		TlsSourceAttrTypes,
		map[string]attr.Value{
			"ip":                  parseList(ctx, types.StringType, currentRule.Source.IP, "rule.source.ip"),
			"host":                parseNameIDList(ctx, currentRule.Source.Host, "rule.source.host"),
			"site":                parseNameIDList(ctx, currentRule.Source.Site, "rule.source.site"),
			"subnet":              parseList(ctx, types.StringType, currentRule.Source.Subnet, "rule.source.subnet"),
			"ip_range":            parseFromToList(ctx, currentRule.Source.IPRangeTLSInspectSource, "rule.source.ip_range"),
			"global_ip_range":     parseNameIDList(ctx, currentRule.Source.GlobalIPRange, "rule.source.global_ip_range"),
			"network_interface":   parseNameIDList(ctx, currentRule.Source.NetworkInterface, "rule.source.network_interface"),
			"site_network_subnet": parseNameIDList(ctx, currentRule.Source.SiteNetworkSubnet, "rule.source.site_network_subnet"),
			"floating_subnet":     parseNameIDList(ctx, currentRule.Source.FloatingSubnet, "rule.source.floating_subnet"),
			"user":                parseNameIDList(ctx, currentRule.Source.User, "rule.source.user"),
			"users_group":         parseNameIDList(ctx, currentRule.Source.UsersGroup, "rule.source.users_group"),
			"group":               parseNameIDList(ctx, currentRule.Source.Group, "rule.source.group"),
			"system_group":        parseNameIDList(ctx, currentRule.Source.SystemGroup, "rule.source.system_group"),
		},
	)
	diags = append(diags, diagstmp...)

	// Ensure source.network_interface is never unknown after apply
	srcAttrs := curRuleSourceObj.Attributes()
	if v, ok := srcAttrs["network_interface"]; ok {
		if setVal, ok2 := v.(types.Set); ok2 {
			if setVal.IsUnknown() {
				// Replace unknown with known null set to satisfy post-apply requirements
				srcAttrs["network_interface"] = types.SetNull(NameIDObjectType)
				var diagsTmp diag.Diagnostics
				curRuleSourceObj, diagsTmp = types.ObjectValue(curRuleSourceObj.Type(ctx).(types.ObjectType).AttrTypes, srcAttrs)
				diags = append(diags, diagsTmp...)
			}
		}
	}

	ruleInput.Source = curRuleSourceObj
	////////////// end rule.source ///////////////

	// Rule -> Country
	ruleInput.Country = parseNameIDList(ctx, currentRule.Country, "rule.country")

	// Rule -> Device Posture Profile
	ruleInput.DevicePostureProfile = parseNameIDList(ctx, currentRule.DevicePostureProfile, "rule.device_posture_profile")

	// //////////// Rule -> Application ///////////////
	// Note: RemoteAsn is []scalars.Asn32 (which is type string in SDK), so parseList handles it directly
	curRuleApplicationObj, diagstmp := types.ObjectValue(
		TlsApplicationAttrTypes,
		map[string]attr.Value{
			"application":          parseNameIDList(ctx, currentRule.Application.Application, "rule.application.application"),
			"custom_app":           parseNameIDList(ctx, currentRule.Application.CustomApp, "rule.application.custom_app"),
			"app_category":         parseNameIDList(ctx, currentRule.Application.AppCategory, "rule.application.app_category"),
			"custom_category":      parseNameIDList(ctx, currentRule.Application.CustomCategory, "rule.application.custom_category"),
			"domain":               parseList(ctx, types.StringType, currentRule.Application.Domain, "rule.application.domain"),
			"fqdn":                 parseList(ctx, types.StringType, currentRule.Application.Fqdn, "rule.application.fqdn"),
			"ip":                   parseList(ctx, types.StringType, currentRule.Application.IP, "rule.application.ip"),
			"subnet":               parseList(ctx, types.StringType, currentRule.Application.Subnet, "rule.application.subnet"),
			"ip_range":             parseFromToList(ctx, currentRule.Application.IPRangeTLSInspectApplication, "rule.application.ip_range"),
			"global_ip_range":      parseNameIDList(ctx, currentRule.Application.GlobalIPRange, "rule.application.global_ip_range"),
			"remote_asn":           parseList(ctx, types.StringType, currentRule.Application.RemoteAsn, "rule.application.remote_asn"),
			"service":              parseNameIDList(ctx, currentRule.Application.Service, "rule.application.service"),
			"custom_service":       parseCustomService(ctx, currentRule.Application.CustomService, "rule.application.custom_service"),
			"custom_service_ip":    parseCustomServiceIp(ctx, currentRule.Application.CustomServiceIP, "rule.application.custom_service_ip"),
			"tls_inspect_category": parseTlsInspectCategory(ctx, currentRule.Application.TLSInspectCategory),
			"country":              parseNameIDList(ctx, currentRule.Application.Country, "rule.application.country"),
		},
	)
	ruleInput.Application = curRuleApplicationObj
	diags = append(diags, diagstmp...)
	////////////// end Rule -> Application ///////////////

	return ruleInput, diags
}

// parseTlsInspectCategory handles the TLS inspection category field
func parseTlsInspectCategory(ctx context.Context, category interface{}) types.String {
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

