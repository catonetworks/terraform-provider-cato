package tlsrule

import (
	"context"
	"fmt"
	"reflect"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/customservice"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/iprange"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/oldidname"
)

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
