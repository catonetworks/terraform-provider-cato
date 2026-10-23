package alwaysonrule

import (
	"context"
	"encoding/json"
	"testing"

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
	rule.Source, diags = types.ObjectValueFrom(ctx, AlwaysOnRuleSourceAttrTypes, AlwaysOnRuleSource{User: refs, UsersGroup: refs})
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
	require.True(t, source.User.Equal(types.SetValueMust(idname.RefObjectType, nil)))
	require.True(t, source.UsersGroup.Equal(types.SetValueMust(idname.RefObjectType, nil)))
}
