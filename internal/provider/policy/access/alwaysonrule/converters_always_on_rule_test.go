package alwaysonrule

import (
	"context"
	"encoding/json"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
)

func TestBuildAlwaysOnPositionAndEmptyArrays(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	for _, ref := range []string{"", "section-1"} {
		t.Run(ref, func(t *testing.T) {
			t.Parallel()
			model := alwaysOnPlanModel(t)
			var rule AlwaysOnRuleData
			require.False(t, model.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{}).HasError())
			rule.Platform = types.SetValueMust(types.StringType, nil)
			var conversionDiags diag.Diagnostics
			model.Rule, conversionDiags = types.ObjectValueFrom(ctx, AlwaysOnRuleDataAttrTypes, rule)
			require.False(t, conversionDiags.HasError())
			model.At, conversionDiags = types.ObjectValueFrom(ctx, policyrule.PositionAttrTypes, policyrule.Position{
				Position: types.StringValue("FIRST_IN_POLICY"), Ref: types.StringValue(ref),
			})
			require.False(t, conversionDiags.HasError())
			input := buildAlwaysOnRuleInput(ctx, model, &conversionDiags)
			require.False(t, conversionDiags.HasError(), "%v", conversionDiags)
			body, err := json.Marshal(input)
			require.NoError(t, err)
			var payload map[string]any
			require.NoError(t, json.Unmarshal(body, &payload))
			at := payload["at"].(map[string]any)
			require.Equal(t, "FIRST_IN_POLICY", at["position"])
			if ref == "" {
				require.NotContains(t, at, "ref")
			} else {
				require.Equal(t, ref, at["ref"])
			}
			ruleJSON := payload["rule"].(map[string]any)
			require.Equal(t, []any{}, ruleJSON["platform"])
			require.Equal(t, []any{}, ruleJSON["devicePostureProfile"])
			source := ruleJSON["source"].(map[string]any)
			require.Equal(t, []any{}, source["user"])
			require.Equal(t, []any{}, source["usersGroup"])
		})
	}
}

func TestBuildAlwaysOnRuleAllReferenceCollections(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	plan := alwaysOnPlanModel(t)
	var rule AlwaysOnRuleData
	require.False(t, plan.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{}).HasError())
	refs := types.SetValueMust(idname.RefObjectType, []attr.Value{
		types.ObjectValueMust(idname.RefAttrTypes, map[string]attr.Value{
			"id": types.StringUnknown(), "name": types.StringValue("Example"),
		}),
	})
	rule.DevicePostureProfile = refs
	var diags diag.Diagnostics
	rule.Source, diags = types.ObjectValueFrom(ctx, AlwaysOnRuleSourceAttrTypes, AlwaysOnRuleSource{Users: refs, UserGroups: refs})
	require.False(t, diags.HasError())
	plan.Rule, diags = types.ObjectValueFrom(ctx, AlwaysOnRuleDataAttrTypes, rule)
	require.False(t, diags.HasError())
	input := buildAlwaysOnRuleInput(ctx, plan, &diags)
	require.False(t, diags.HasError(), "%v", diags)
	require.Equal(t, []*cato_models.DeviceProfileRefInput{{By: cato_models.ObjectRefByName, Input: "Example"}}, input.Rule.DevicePostureProfile)
	require.Equal(t, []*cato_models.UserRefInput{{By: cato_models.ObjectRefByName, Input: "Example"}}, input.Rule.Source.User)
	require.Equal(t, []*cato_models.UsersGroupRefInput{{By: cato_models.ObjectRefByName, Input: "Example"}}, input.Rule.Source.UsersGroup)
}

func TestParseAlwaysOnRuleReferencesAndDefaults(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	response := alwaysOnResponse(t)
	rule := response.Policy.ZtnaAlwaysOn.AddRule.Rule.GetRule()
	// Exercise nullable API output and normalize nil collection results to empty sets.
	rule.AntiTamperMode = nil
	rule.Platform = nil
	rule.Source.User = nil
	rule.Source.UsersGroup = nil
	require.NoError(t, json.Unmarshal([]byte(`[{"id":"profile-1","name":"Managed device"}]`), &rule.DevicePostureProfile))
	var diags diag.Diagnostics
	parsed := parseAlwaysOnRule(ctx, rule, &diags)
	require.False(t, diags.HasError(), "%v", diags)
	var state AlwaysOnRuleData
	require.False(t, parsed.As(ctx, &state, basetypes.ObjectAsOptions{}).HasError())
	require.Equal(t, "OFF", state.AntiTamperMode.ValueString())
	require.False(t, state.Platform.IsNull())
	require.Empty(t, state.Platform.Elements())
	var profiles []idname.Ref
	require.False(t, state.DevicePostureProfile.ElementsAs(ctx, &profiles, false).HasError())
	require.Equal(t, []idname.Ref{{ID: types.StringValue("profile-1"), Name: types.StringValue("Managed device")}}, profiles)
	var source AlwaysOnRuleSource
	require.False(t, state.Source.As(ctx, &source, basetypes.ObjectAsOptions{}).HasError())
	require.True(t, source.Users.Equal(types.SetValueMust(idname.RefObjectType, nil)))
	require.True(t, source.UserGroups.Equal(types.SetValueMust(idname.RefObjectType, nil)))
}

func TestParseAlwaysOnPolicyRuleDefaults(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	rule := alwaysOnPolicyResponse(t).Policy.ZtnaAlwaysOn.Policy.Rules[1].GetRule()
	rule.AntiTamperMode = nil
	rule.Platform = nil
	rule.DevicePostureProfile = nil
	rule.Source.User = nil
	rule.Source.UsersGroup = nil
	var diags diag.Diagnostics
	parsed := parseAlwaysOnPolicyRule(ctx, rule, &diags)
	require.False(t, diags.HasError(), "%v", diags)
	var state AlwaysOnRuleData
	require.False(t, parsed.As(ctx, &state, basetypes.ObjectAsOptions{}).HasError())
	require.Equal(t, "OFF", state.AntiTamperMode.ValueString())
	require.True(t, state.Platform.Equal(types.SetValueMust(types.StringType, nil)))
	require.True(t, state.DevicePostureProfile.Equal(types.SetValueMust(idname.RefObjectType, nil)))
	var source AlwaysOnRuleSource
	require.False(t, state.Source.As(ctx, &source, basetypes.ObjectAsOptions{}).HasError())
	require.True(t, source.Users.Equal(types.SetValueMust(idname.RefObjectType, nil)))
	require.True(t, source.UserGroups.Equal(types.SetValueMust(idname.RefObjectType, nil)))
}

func TestParseAlwaysOnRuleCollections(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	for _, operation := range []string{"create", "read", "update", "move"} {
		t.Run(operation, func(t *testing.T) {
			t.Parallel()
			for _, collection := range []string{"null", "empty", "populated"} {
				t.Run(collection, func(t *testing.T) {
					t.Parallel()
					var platforms []string
					var references []map[string]string
					wantPlatforms := types.SetValueMust(types.StringType, nil)
					wantReferences := types.SetValueMust(idname.RefObjectType, nil)
					switch collection {
					case "empty":
						platforms = []string{}
						references = []map[string]string{}
					case "populated":
						platforms = []string{"WINDOWS", "MACOS"}
						references = []map[string]string{{"id": "ref-1", "name": "Example"}}
						wantPlatforms = types.SetValueMust(types.StringType, []attr.Value{
							types.StringValue("MACOS"), types.StringValue("WINDOWS"),
						})
						ref, diags := types.ObjectValueFrom(ctx, idname.RefAttrTypes, idname.Ref{
							ID: types.StringValue("ref-1"), Name: types.StringValue("Example"),
						})
						require.False(t, diags.HasError(), "%v", diags)
						wantReferences = types.SetValueMust(idname.RefObjectType, []attr.Value{ref})
					}
					body, err := json.Marshal(map[string]any{
						"action": "ENFORCE", "platform": platforms, "devicePostureProfile": references,
						"source": map[string]any{"user": references, "usersGroup": references},
					})
					require.NoError(t, err)
					var diags diag.Diagnostics
					parsed := parseAlwaysOnTestRule(t, operation, body, &diags)
					require.False(t, diags.HasError(), "%v", diags)
					var rule AlwaysOnRuleData
					require.False(t, parsed.As(ctx, &rule, basetypes.ObjectAsOptions{}).HasError())
					require.True(t, wantPlatforms.Equal(rule.Platform), "expected %s, got %s", wantPlatforms, rule.Platform)
					require.True(t, wantReferences.Equal(rule.DevicePostureProfile), "expected %s, got %s", wantReferences, rule.DevicePostureProfile)
					var source AlwaysOnRuleSource
					require.False(t, rule.Source.As(ctx, &source, basetypes.ObjectAsOptions{}).HasError())
					require.True(t, wantReferences.Equal(source.Users), "expected %s, got %s", wantReferences, source.Users)
					require.True(t, wantReferences.Equal(source.UserGroups), "expected %s, got %s", wantReferences, source.UserGroups)
				})
			}
		})
	}
}

func parseAlwaysOnTestRule(t *testing.T, operation string, body []byte, diags *diag.Diagnostics) types.Object {
	t.Helper()
	ctx := context.Background()
	switch operation {
	case "create":
		var rule cato.PolicyZtnaAlwaysOnAddRule_Policy_ZtnaAlwaysOn_AddRule_Rule_Rule
		require.NoError(t, json.Unmarshal(body, &rule))
		return parseAlwaysOnRule(ctx, &rule, diags)
	case "read":
		var rule cato.PolicyZtnaAlwaysOnPolicy_Policy_ZtnaAlwaysOn_Policy_Rules_Rule
		require.NoError(t, json.Unmarshal(body, &rule))
		return parseAlwaysOnPolicyRule(ctx, &rule, diags)
	case "update":
		var rule cato.PolicyZtnaAlwaysOnUpdateRule_Policy_ZtnaAlwaysOn_UpdateRule_Rule_Rule
		require.NoError(t, json.Unmarshal(body, &rule))
		return parseUpdatedAlwaysOnRule(ctx, &rule, diags)
	case "move":
		var rule cato.PolicyZtnaAlwaysOnMoveRule_Policy_ZtnaAlwaysOn_MoveRule_Rule_Rule
		require.NoError(t, json.Unmarshal(body, &rule))
		return parseMovedAlwaysOnRule(ctx, &rule, diags)
	default:
		t.Fatalf("unsupported operation %q", operation)
		return types.ObjectNull(AlwaysOnRuleDataAttrTypes)
	}
}
