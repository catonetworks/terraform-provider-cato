package acrule

import (
	"context"
	"encoding/json"
	"testing"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
)

func TestApplicationControlRuleStateIDNameSets(t *testing.T) {
	t.Parallel()
	ref := map[string]any{"id": "ref-1", "name": "Reference"}
	emptySet := types.SetValueMust(idname.RefObjectType, nil)
	populatedSet := types.SetValueMust(idname.RefObjectType, []attr.Value{
		types.ObjectValueMust(idname.RefAttrTypes, map[string]attr.Value{
			"id": types.StringValue("ref-1"), "name": types.StringValue("Reference"),
		}),
	})

	for _, ruleType := range []string{"APPLICATION", "DATA", "FILE"} {
		for _, tc := range []struct {
			name           string
			items          any
			applicationRef any
			want           types.Set
		}{
			{name: "null", want: emptySet},
			{name: "empty array", items: []any{}, want: emptySet},
			{name: "null elements", items: []any{nil}, want: emptySet},
			{name: "populated", items: []any{ref}, applicationRef: ref, want: populatedSet},
		} {
			t.Run(ruleType+"/"+tc.name, func(t *testing.T) {
				t.Parallel()
				ctx := context.Background()
				apiRule, ruleAttribute := applicationControlRuleWithIDNameSets(t, ruleType, tc.items, tc.applicationRef)
				hydrated, diags := hydrateApplicationControlRuleStateFromClient(ctx, apiRule)
				require.False(t, diags.HasError(), "%v", diags)
				ruleObject, diags := applicationControlRuleRuleObjectFromPlan(hydrated)
				require.False(t, diags.HasError(), "%v", diags)
				state := tfsdk.State{Schema: getApplicationControlRuleSchema(ctx, t)}
				diags = state.Set(ctx, ApplicationControlRule{
					ID: types.StringValue("rule-1"), Rule: ruleObject, At: types.ObjectNull(policyrule.PositionAttrTypes),
				})
				require.False(t, diags.HasError(), "%v", diags)

				for _, names := range applicationControlIDNameSetPaths(ruleType) {
					attributePath := path.Root("rule").AtName(ruleAttribute)
					for _, name := range names {
						attributePath = attributePath.AtName(name)
					}
					var got types.Set
					diags = state.GetAttribute(ctx, attributePath, &got)
					require.False(t, diags.HasError(), "%s: %v", attributePath, diags)
					want := tc.want
					if len(names) == 2 && names[0] == "application" && names[1] == "global_ip_range" {
						want = emptySet // The Application Control API has no global IP range matcher.
					}
					require.True(t, got.Equal(want), "%s: got %s, want %s", attributePath, got, want)
				}
			})
		}
	}
}

func applicationControlIDNameSetPaths(ruleType string) [][]string {
	paths := [][]string{{"device"}, {"action_config", "user_notification"}}
	for _, name := range []string{
		"country", "host", "site", "global_ip_range", "network_interface", "site_network_subnet",
		"floating_subnet", "user", "users_group", "group", "system_group",
	} {
		paths = append(paths, []string{"source", name})
	}
	for _, name := range []string{"subscription_group", "webhook", "mailing_list"} {
		paths = append(paths, []string{"tracking", "alert", name})
	}
	for _, name := range []string{
		"application", "custom_app", "app_category", "custom_category", "sanctioned_apps_category", "global_ip_range",
	} {
		paths = append(paths, []string{"application", name})
	}
	if ruleType == "DATA" {
		paths = append(paths, []string{"dlp_profile", "content_profile"}, []string{"dlp_profile", "edm_profile"})
	}
	return paths
}

func applicationControlRuleWithIDNameSets(
	t *testing.T, ruleType string, items, applicationRef any,
) (rule *cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule, stateAttribute string) {
	t.Helper()
	source := make(map[string]any)
	for _, name := range []string{
		"country", "host", "site", "globalIpRange", "networkInterface", "siteNetworkSubnet",
		"floatingSubnet", "user", "usersGroup", "group", "systemGroup",
	} {
		source[name] = items
	}
	application := make(map[string]any)
	for _, name := range []string{"application", "customApp", "appCategory", "customCategory", "sanctionedAppsCategory"} {
		application[name] = applicationRef
	}
	typedRule := map[string]any{
		"action": "ALLOW", "severity": "MEDIUM", "device": items, "source": source, "application": application,
		"tracking": map[string]any{
			"event": map[string]any{"enabled": false},
			"alert": map[string]any{
				"enabled": false, "frequency": "HOURLY", "subscriptionGroup": items, "webhook": items, "mailingList": items,
			},
		},
		"actionConfig": map[string]any{"userNotification": items},
	}
	if ruleType == "DATA" {
		typedRule["dlpProfile"] = map[string]any{"contentProfile": items, "edmProfile": items}
	}
	apiAttribute := map[string]string{"APPLICATION": "applicationRule", "DATA": "dataRule", "FILE": "fileRule"}[ruleType]
	stateAttribute = map[string]string{"APPLICATION": "application_rule", "DATA": "data_rule", "FILE": "file_rule"}[ruleType]
	body, err := json.Marshal(map[string]any{
		"id": "rule-1", "name": "Rule", "enabled": true, "ruleType": ruleType, apiAttribute: typedRule,
	})
	require.NoError(t, err)
	rule = &cato_go_sdk.ApplicationControlPolicy_Policy_ApplicationControl_Policy_Rules_Rule{}
	require.NoError(t, json.Unmarshal(body, rule))
	return rule, stateAttribute
}
