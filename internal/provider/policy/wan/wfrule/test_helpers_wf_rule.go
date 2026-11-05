package wfrule

import (
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	attr "github.com/hashicorp/terraform-plugin-framework/attr"
	types "github.com/hashicorp/terraform-plugin-framework/types"

	activeperiod "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/activeperiod"
	convert "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	idname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	policyrule "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
	policyschedule "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyschedule"
)

func NewMinimalWanFwRuleModel(ruleID string) Rule {
	ruleAttrs := map[string]attr.Value{
		"id":                convert.NullableString(ruleID),
		"name":              types.StringValue("test-wfw-rule"),
		"description":       types.StringNull(),
		"enabled":           types.BoolValue(true),
		"source":            types.ObjectNull(WanSourceAttrTypes),
		"connection_origin": types.StringValue("ANY"),
		"active_period":     types.ObjectNull(activeperiod.ActivePeriodAttrTypes),
		"country":           types.SetNull(idname.RefObjectType),
		"device":            types.SetNull(idname.RefObjectType),
		"device_attributes": types.ObjectNull(WanDeviceAttrAttrTypes),
		"device_os":         types.ListNull(types.StringType),
		"application":       types.ObjectNull(WanApplicationAttrTypes),
		"destination":       types.ObjectNull(WanDestAttrTypes),
		"service":           types.ObjectNull(WanServiceAttrTypes),
		"action":            types.StringValue("ALLOW"),
		"tracking":          minimalWanTrackingObject(),
		"schedule":          types.ObjectNull(policyschedule.ScheduleAttrTypes),
		"direction":         types.StringValue("TO"),
		"exceptions":        types.SetNull(WanExceptionObjectType),
	}

	return Rule{
		Rule: types.ObjectValueMust(WanFirewallRuleRuleAttrTypes, ruleAttrs),
		At: types.ObjectValueMust(policyrule.PositionAttrTypes, map[string]attr.Value{
			"position": types.StringValue("LAST_IN_POLICY"),
			"ref":      types.StringNull(),
		}),
	}
}

func minimalWanTrackingObject() types.Object {
	return types.ObjectValueMust(policyrule.TrackingAttrTypes, map[string]attr.Value{
		"event": types.ObjectValueMust(policyrule.TrackingEventAttrTypes, map[string]attr.Value{
			"enabled": types.BoolValue(false),
		}),
		"alert": types.ObjectNull(policyrule.TrackingAlertAttrTypes),
	})
}

func EmptyWanFirewallPolicyResponse() *cato_go_sdk.Policy {
	return &cato_go_sdk.Policy{
		Policy: cato_go_sdk.Policy_Policy{
			WanFirewall: &cato_go_sdk.Policy_Policy_WanFirewall{
				Policy: cato_go_sdk.Policy_Policy_WanFirewall_Policy{
					Rules: []*cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules{},
				},
			},
		},
	}
}

func SuccessfulUpdateRuleResponse(ruleID string) *cato_go_sdk.PolicyWanFirewallUpdateRule {
	return &cato_go_sdk.PolicyWanFirewallUpdateRule{
		Policy: cato_go_sdk.PolicyWanFirewallUpdateRule_Policy{
			WanFirewall: &cato_go_sdk.PolicyWanFirewallUpdateRule_Policy_WanFirewall{
				UpdateRule: cato_go_sdk.PolicyWanFirewallUpdateRule_Policy_WanFirewall_UpdateRule{
					Status: "SUCCESS",
					Rule: &cato_go_sdk.PolicyWanFirewallUpdateRule_Policy_WanFirewall_UpdateRule_Rule{
						Rule: cato_go_sdk.PolicyWanFirewallUpdateRule_Policy_WanFirewall_UpdateRule_Rule_Rule{
							ID: ruleID,
						},
					},
				},
			},
		},
	}
}

func MinimalAPIRule(name string, index int64) cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules_Rule {
	return cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules_Rule{
		ID:      "rule-123",
		Name:    name,
		Index:   index,
		Enabled: true,
	}
}

func WanFirewallPolicyResponseWithRule(
	rule cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules_Rule,
) *cato_go_sdk.Policy {
	return &cato_go_sdk.Policy{
		Policy: cato_go_sdk.Policy_Policy{
			WanFirewall: &cato_go_sdk.Policy_Policy_WanFirewall{
				Policy: cato_go_sdk.Policy_Policy_WanFirewall_Policy{
					Rules: []*cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules{
						{
							Rule: rule,
						},
					},
				},
			},
		},
	}
}
