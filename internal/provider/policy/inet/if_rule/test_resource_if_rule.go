package inetfirewall

import (
	"context"
	"testing"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	activeperiod "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/active_period"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	customservice "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/custom_service"
	idname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/id_name"
	ipaddr "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/ip_range"
	policyrule "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policy_rule"
	policyschedule "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policy_schedule"
)

func NewMinimalInternetFwRuleModel(ruleID string) InternetFirewallRule {
	ruleAttrs := map[string]attr.Value{
		"id":                convert.NullableString(ruleID),
		"name":              types.StringValue("test-ifw-rule"),
		"description":       types.StringNull(),
		"enabled":           types.BoolValue(true),
		"source":            emptyIfwSourceObject(),
		"connection_origin": types.StringNull(),
		"active_period":     types.ObjectNull(activeperiod.ActivePeriodAttrTypes),
		"country":           types.SetNull(idname.RefObjectType),
		"device":            types.SetNull(idname.RefObjectType),
		"device_attributes": types.ObjectNull(IfwDeviceAttrAttrTypes),
		"device_os":         types.ListNull(types.StringType),
		"destination":       emptyIfwDestinationObject(),
		"service":           emptyIfwServiceObject(),
		"action":            types.StringValue("ALLOW"),
		"tracking":          minimalTrackingObject(),
		"schedule":          types.ObjectNull(policyschedule.ScheduleAttrTypes),
		"exceptions":        types.SetNull(IfwExceptionObjectType),
	}

	return InternetFirewallRule{
		Rule: types.ObjectValueMust(InternetFirewallRuleRuleAttrTypes, ruleAttrs),
		At: types.ObjectValueMust(policyrule.PositionAttrTypes, map[string]attr.Value{
			"position": types.StringValue("LAST_IN_POLICY"),
			"ref":      types.StringNull(),
		}),
	}
}

func emptyIfwSourceObject() types.Object {
	return types.ObjectValueMust(IfwSourceAttrTypes, map[string]attr.Value{
		"ip":                  types.ListNull(types.StringType),
		"host":                types.SetNull(idname.RefObjectType),
		"site":                types.SetNull(idname.RefObjectType),
		"subnet":              types.ListNull(types.StringType),
		"ip_range":            types.ListNull(ipaddr.RangeObjectType),
		"global_ip_range":     types.SetNull(idname.RefObjectType),
		"network_interface":   types.SetNull(idname.RefObjectType),
		"site_network_subnet": types.SetNull(idname.RefObjectType),
		"floating_subnet":     types.SetNull(idname.RefObjectType),
		"user":                types.SetNull(idname.RefObjectType),
		"users_group":         types.SetNull(idname.RefObjectType),
		"group":               types.SetNull(idname.RefObjectType),
		"system_group":        types.SetNull(idname.RefObjectType),
	})
}

func emptyIfwDestinationObject() types.Object {
	return types.ObjectValueMust(IfwDestAttrTypes, map[string]attr.Value{
		"application":              types.SetNull(idname.RefObjectType),
		"custom_app":               types.SetNull(idname.RefObjectType),
		"app_category":             types.SetNull(idname.RefObjectType),
		"custom_category":          types.SetNull(idname.RefObjectType),
		"sanctioned_apps_category": types.SetNull(idname.RefObjectType),
		"country":                  types.SetNull(idname.RefObjectType),
		"domain":                   types.ListNull(types.StringType),
		"fqdn":                     types.ListNull(types.StringType),
		"ip":                       types.ListNull(types.StringType),
		"subnet":                   types.ListNull(types.StringType),
		"ip_range":                 types.ListNull(ipaddr.RangeObjectType),
		"global_ip_range":          types.SetNull(idname.RefObjectType),
		"remote_asn":               types.ListNull(types.StringType),
	})
}

func emptyIfwServiceObject() types.Object {
	return types.ObjectValueMust(IfwServiceAttrTypes, map[string]attr.Value{
		"standard": types.SetNull(idname.RefObjectType),
		"custom":   types.ListNull(customservice.CustomServiceObjectType),
	})
}

func minimalTrackingObject() types.Object {
	return types.ObjectValueMust(policyrule.TrackingAttrTypes, map[string]attr.Value{
		"event": types.ObjectValueMust(policyrule.TrackingEventAttrTypes, map[string]attr.Value{
			"enabled": types.BoolValue(false),
		}),
		"alert": types.ObjectNull(policyrule.TrackingAlertAttrTypes),
	})
}

func EmptyInternetFirewallPolicyResponse() *cato_go_sdk.Policy {
	return &cato_go_sdk.Policy{
		Policy: cato_go_sdk.Policy_Policy{
			InternetFirewall: &cato_go_sdk.Policy_Policy_InternetFirewall{
				Policy: cato_go_sdk.Policy_Policy_InternetFirewall_Policy{
					Rules: []*cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules{},
				},
			},
		},
	}
}

func SuccessfulUpdateRuleResponse(ruleID string) *cato_go_sdk.PolicyInternetFirewallUpdateRule {
	return &cato_go_sdk.PolicyInternetFirewallUpdateRule{
		Policy: cato_go_sdk.PolicyInternetFirewallUpdateRule_Policy{
			InternetFirewall: &cato_go_sdk.PolicyInternetFirewallUpdateRule_Policy_InternetFirewall{
				UpdateRule: cato_go_sdk.PolicyInternetFirewallUpdateRule_Policy_InternetFirewall_UpdateRule{
					Status: "SUCCESS",
					Rule: &cato_go_sdk.PolicyInternetFirewallUpdateRule_Policy_InternetFirewall_UpdateRule_Rule{
						Rule: cato_go_sdk.PolicyInternetFirewallUpdateRule_Policy_InternetFirewall_UpdateRule_Rule_Rule{
							ID: ruleID,
						},
					},
				},
			},
		},
	}
}

func MinimalAPIRule(name string, index int64) cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_Rule {
	return cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_Rule{
		ID:               "rule-123",
		Name:             name,
		Index:            index,
		Enabled:          true,
		Action:           cato_go_sdk.PolicyInternetFirewallAddRule_Policy_InternetFirewall_AddRule_Rule_Rule{}.Action,
		ConnectionOrigin: cato_go_sdk.PolicyInternetFirewallUpdateRule_Policy_InternetFirewall_UpdateRule_Rule_Rule{}.ConnectionOrigin,
	}
}

func GetInternetFwRuleSchema(ctx context.Context, t *testing.T) schema.Schema {
	t.Helper()

	r := &InternetFwRuleResource{}
	resp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, resp)

	return resp.Schema
}

func SuccessfulAddRuleResponse(ruleID string) *cato_go_sdk.PolicyInternetFirewallAddRule {
	return &cato_go_sdk.PolicyInternetFirewallAddRule{
		Policy: cato_go_sdk.PolicyInternetFirewallAddRule_Policy{
			InternetFirewall: &cato_go_sdk.PolicyInternetFirewallAddRule_Policy_InternetFirewall{
				AddRule: cato_go_sdk.PolicyInternetFirewallAddRule_Policy_InternetFirewall_AddRule{
					Status: "SUCCESS",
					Rule: &cato_go_sdk.PolicyInternetFirewallAddRule_Policy_InternetFirewall_AddRule_Rule{
						Rule: cato_go_sdk.PolicyInternetFirewallAddRule_Policy_InternetFirewall_AddRule_Rule_Rule{
							ID: ruleID,
						},
					},
				},
			},
		},
	}
}

func InternetFirewallPolicyResponseWithRule(rule cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_Rule) *cato_go_sdk.Policy {
	return &cato_go_sdk.Policy{
		Policy: cato_go_sdk.Policy_Policy{
			InternetFirewall: &cato_go_sdk.Policy_Policy_InternetFirewall{
				Policy: cato_go_sdk.Policy_Policy_InternetFirewall_Policy{
					Rules: []*cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules{
						{
							Rule: rule,
						},
					},
				},
			},
		},
	}
}
