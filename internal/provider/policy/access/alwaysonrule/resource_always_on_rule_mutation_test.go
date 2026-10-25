package alwaysonrule

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
)

func alwaysOnMutationJSON(t *testing.T, operation string, rule any) []byte {
	t.Helper()
	body, err := json.Marshal(map[string]any{"policy": map[string]any{"ztnaAlwaysOn": map[string]any{
		operation: map[string]any{"status": "SUCCESS", "errors": []any{}, "rule": map[string]any{"rule": rule}},
	}}})
	require.NoError(t, err)
	return body
}

func alwaysOnUpdatedResponse(t *testing.T) *cato.PolicyZtnaAlwaysOnUpdateRule {
	t.Helper()
	rule := alwaysOnResponse(t).Policy.ZtnaAlwaysOn.AddRule.Rule.GetRule()
	rule.Name = "Updated rule"
	rule.Description = ""
	rule.Enabled = false
	rule.AllowUserBypass = false
	rule.Platform = []cato_models.OperatingSystem{}
	var result cato.PolicyZtnaAlwaysOnUpdateRule
	require.NoError(t, json.Unmarshal(alwaysOnMutationJSON(t, "updateRule", rule), &result))
	return &result
}

func alwaysOnMovedResponse(t *testing.T) *cato.PolicyZtnaAlwaysOnMoveRule {
	t.Helper()
	rule := alwaysOnUpdatedResponse(t).Policy.ZtnaAlwaysOn.UpdateRule.Rule.GetRule()
	rule.Index = 11
	rule.Section.ID, rule.Section.Name = "section-2", "Moved section"
	var result cato.PolicyZtnaAlwaysOnMoveRule
	require.NoError(t, json.Unmarshal(alwaysOnMutationJSON(t, "moveRule", rule), &result))
	return &result
}

func alwaysOnUpdateModel(t *testing.T) AlwaysOnRule {
	t.Helper()
	ctx := context.Background()
	var model AlwaysOnRule
	require.False(t, alwaysOnState(t).Get(ctx, &model).HasError())
	var rule AlwaysOnRuleData
	require.False(t, model.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{}).HasError())
	rule.Name, rule.Description = types.StringValue("Updated rule"), types.StringValue("")
	rule.Enabled, rule.AllowUserBypass = types.BoolValue(false), types.BoolValue(false)
	rule.Platform = types.SetValueMust(types.StringType, nil)
	// Updates must address the existing state ID, even when a plan ID is unknown.
	rule.ID = types.StringUnknown()
	var diags diag.Diagnostics
	model.Rule, diags = types.ObjectValueFrom(ctx, AlwaysOnRuleDataAttrTypes, rule)
	require.False(t, diags.HasError())
	return model
}

func alwaysOnPosition(t *testing.T, position, ref string) types.Object {
	t.Helper()
	value, diags := types.ObjectValueFrom(context.Background(), policyrule.PositionAttrTypes,
		policyrule.Position{Position: types.StringValue(position), Ref: types.StringValue(ref)})
	require.False(t, diags.HasError())
	return value
}

func TestAlwaysOnRuleUpdate(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"no placement", "unchanged placement", "changed placement", "removed placement"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			state := alwaysOnState(t)
			model := alwaysOnUpdateModel(t)
			switch mode {
			case "unchanged placement", "removed placement":
				at := alwaysOnPosition(t, "FIRST_IN_POLICY", "")
				require.False(t, state.SetAttribute(ctx, path.Root("at"), at).HasError())
				if mode == "unchanged placement" {
					model.At = at
				}
			case "changed placement":
				model.At = alwaysOnPosition(t, "AFTER_RULE", "ref-rule")
			}
			client := mocks.NewAlwaysOnRuleClient(t)
			update := client.On("PolicyZtnaAlwaysOnUpdateRule", mock.Anything, "123", (*cato_models.ZtnaAlwaysOnPolicyMutationInput)(nil),
				mock.MatchedBy(func(input cato_models.ZtnaAlwaysOnUpdateRuleInput) bool {
					require.Equal(t, "rule-123", input.ID)
					require.Equal(t, "Updated rule", *input.Rule.Name)
					require.Equal(t, "", *input.Rule.Description)
					require.False(t, *input.Rule.Enabled)
					require.False(t, *input.Rule.AllowUserBypass)
					require.False(t, *input.Rule.AllowFailOpen)
					require.Equal(t, cato_models.ZtnaAlwaysOnRuleActionEnum("ENFORCE"), *input.Rule.Action)
					require.Equal(t, cato_models.AntiTamperModeEnum("OFF"), *input.Rule.AntiTamperMode)
					require.Equal(t, int64(30), *input.Rule.BypassDuration.Time)
					require.Equal(t, cato_models.ZtnaAlwaysOnTimeUnit("MINUTES"), *input.Rule.BypassDuration.Unit)
					for _, collection := range []any{input.Rule.Platform, input.Rule.DevicePostureProfile, input.Rule.Source.User, input.Rule.Source.UsersGroup} {
						require.NotNil(t, collection)
						require.Empty(t, collection)
					}
					return true
				})).Return(alwaysOnUpdatedResponse(t), nil).Once()
			if mode == "changed placement" {
				client.On("PolicyZtnaAlwaysOnMoveRule", mock.Anything, "123", (*cato_models.ZtnaAlwaysOnPolicyMutationInput)(nil),
					mock.MatchedBy(func(input cato_models.PolicyMoveRuleInput) bool {
						require.Equal(t, "rule-123", input.ID)
						require.Equal(t, cato_models.PolicyRulePositionEnum("AFTER_RULE"), *input.To.Position)
						require.Equal(t, "ref-rule", *input.To.Ref)
						return true
					})).Return(alwaysOnMovedResponse(t), nil).Once().NotBefore(update)
			}
			r := &alwaysOnRuleResource{client: &catoclient.Service{AccountId: "123"}, apiClient: client}
			resp := resource.UpdateResponse{State: state}
			r.Update(ctx, resource.UpdateRequest{Plan: alwaysOnPlan(t, model), State: state}, &resp)
			require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			var actual AlwaysOnRule
			require.False(t, resp.State.Get(ctx, &actual).HasError())
			require.True(t, model.At.Equal(actual.At))
			var rule AlwaysOnRuleData
			require.False(t, actual.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{}).HasError())
			require.Equal(t, "rule-123", rule.ID.ValueString())
			require.Equal(t, "Updated rule", rule.Name.ValueString())
			require.False(t, rule.Enabled.ValueBool())
			require.False(t, rule.AllowUserBypass.ValueBool())
			require.Equal(t, "", rule.Description.ValueString())
			require.Empty(t, rule.Platform.Elements())
			if mode == "changed placement" {
				require.Equal(t, int64(11), rule.Index.ValueInt64())
				require.Equal(t, "section-2", rule.Section.Attributes()["id"].(types.String).ValueString())
			} else {
				require.Equal(t, int64(4), rule.Index.ValueInt64())
			}
		})
	}
}

func TestAlwaysOnRuleUpdateErrors(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, body, message string
		err                 error
	}{
		{name: "transport", err: errors.New("request failed"), message: "request failed"},
		{name: "nil response", message: "mutation status"},
		{name: "missing ID", body: `{"policy":{"ztnaAlwaysOn":{"updateRule":{"status":"SUCCESS"}}}}`, message: "resource ID"},
		{name: "wrong ID", body: `{"policy":{"ztnaAlwaysOn":{"updateRule":{"status":"SUCCESS","rule":{"rule":{"id":"other"}}}}}}`, message: "resource ID"},
		{name: "failure status", body: `{"policy":{"ztnaAlwaysOn":{"updateRule":{"status":"FAILURE"}}}}`, message: "FAILURE"},
		{name: "payload error", body: `{"policy":{"ztnaAlwaysOn":{"updateRule":{"status":"SUCCESS","errors":[{"errorCode":"invalidRule","errorMessage":"Rule rejected"}]}}}}`, message: "Rule rejected"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var result *cato.PolicyZtnaAlwaysOnUpdateRule
			if tc.body != "" {
				require.NoError(t, json.Unmarshal([]byte(tc.body), &result))
			}
			client := mocks.NewAlwaysOnRuleClient(t)
			client.On("PolicyZtnaAlwaysOnUpdateRule", mock.Anything, "123", mock.Anything, mock.Anything).Return(result, tc.err).Once()
			r := &alwaysOnRuleResource{client: &catoclient.Service{AccountId: "123"}, apiClient: client}
			state := alwaysOnState(t)
			model := alwaysOnUpdateModel(t)
			model.At = alwaysOnPosition(t, "LAST_IN_POLICY", "")
			resp := resource.UpdateResponse{State: state}
			r.Update(context.Background(), resource.UpdateRequest{Plan: alwaysOnPlan(t, model), State: state}, &resp)
			require.True(t, resp.Diagnostics.HasError())
			require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), tc.message)
			require.True(t, state.Raw.Equal(resp.State.Raw))
		})
	}
}

func TestAlwaysOnRuleMoveErrorsPreserveUpdatedRule(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, body, message string
		err                 error
	}{
		{name: "transport", err: errors.New("move failed"), message: "move failed"},
		{name: "nil response", message: "mutation status"},
		{name: "missing ID", body: `{"policy":{"ztnaAlwaysOn":{"moveRule":{"status":"SUCCESS"}}}}`, message: "resource ID"},
		{name: "wrong ID", body: `{"policy":{"ztnaAlwaysOn":{"moveRule":{"status":"SUCCESS","rule":{"rule":{"id":"other"}}}}}}`, message: "resource ID"},
		{name: "failure", body: `{"policy":{"ztnaAlwaysOn":{"moveRule":{"status":"FAILURE","errors":[{"errorMessage":"Position rejected"}]}}}}`, message: "Position rejected"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var result *cato.PolicyZtnaAlwaysOnMoveRule
			if tc.body != "" {
				require.NoError(t, json.Unmarshal([]byte(tc.body), &result))
			}
			client := mocks.NewAlwaysOnRuleClient(t)
			update := client.On("PolicyZtnaAlwaysOnUpdateRule", mock.Anything, "123", mock.Anything, mock.Anything).Return(alwaysOnUpdatedResponse(t), nil).Once()
			client.On("PolicyZtnaAlwaysOnMoveRule", mock.Anything, "123", mock.Anything, mock.Anything).Return(result, tc.err).Once().NotBefore(update)
			r := &alwaysOnRuleResource{client: &catoclient.Service{AccountId: "123"}, apiClient: client}
			state := alwaysOnState(t)
			model := alwaysOnUpdateModel(t)
			model.At = alwaysOnPosition(t, "LAST_IN_POLICY", "")
			resp := resource.UpdateResponse{State: state}
			r.Update(context.Background(), resource.UpdateRequest{Plan: alwaysOnPlan(t, model), State: state}, &resp)
			require.True(t, resp.Diagnostics.HasError())
			require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), tc.message)
			var actual AlwaysOnRule
			require.False(t, resp.State.Get(context.Background(), &actual).HasError())
			require.True(t, actual.At.IsNull())
			require.Equal(t, types.StringValue("Updated rule"), actual.Rule.Attributes()["name"])
		})
	}
}

func TestAlwaysOnRuleDelete(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, body, message string
		err                 error
	}{
		{name: "success", body: `{"policy":{"ztnaAlwaysOn":{"removeRule":{"status":"SUCCESS","errors":[]}}}}`},
		{name: "already removed", body: `{"policy":{"ztnaAlwaysOn":{"removeRule":{"status":"FAILURE","errors":[{"errorCode":"ruleNotExist","errorMessage":"Gone"}]}}}}`},
		{name: "transport", err: errors.New("delete failed"), message: "delete failed"},
		{name: "nil response", message: "mutation status"},
		{name: "failure status", body: `{"policy":{"ztnaAlwaysOn":{"removeRule":{"status":"FAILURE"}}}}`, message: "FAILURE"},
		{name: "payload error", body: `{"policy":{"ztnaAlwaysOn":{"removeRule":{"status":"SUCCESS","errors":[{"errorCode":"PolicyLocked","errorMessage":"Policy locked"}]}}}}`, message: "Policy locked"},
		{name: "mixed errors", body: `{"policy":{"ztnaAlwaysOn":{"removeRule":{"status":"FAILURE","errors":[{"errorCode":"ruleNotExist","errorMessage":"Gone"},{"errorCode":"PolicyLocked","errorMessage":"Policy locked"}]}}}}`, message: "Policy locked"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var result *cato.PolicyZtnaAlwaysOnRemoveRule
			if tc.body != "" {
				require.NoError(t, json.Unmarshal([]byte(tc.body), &result))
			}
			client := mocks.NewAlwaysOnRuleClient(t)
			client.On("PolicyZtnaAlwaysOnRemoveRule", mock.Anything, "123", (*cato_models.ZtnaAlwaysOnPolicyMutationInput)(nil),
				cato_models.ZtnaAlwaysOnRemoveRuleInput{ID: "rule-123"}).Return(result, tc.err).Once()
			r := &alwaysOnRuleResource{client: &catoclient.Service{AccountId: "123"}, apiClient: client}
			state := alwaysOnState(t)
			resp := resource.DeleteResponse{State: state}
			r.Delete(context.Background(), resource.DeleteRequest{State: state}, &resp)
			if tc.message == "" {
				require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
				require.True(t, resp.State.Raw.IsNull())
			} else {
				require.True(t, resp.Diagnostics.HasError())
				require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), tc.message)
				require.True(t, resp.State.Raw.Equal(state.Raw))
			}
		})
	}
}

func TestAlwaysOnRuleMutationsInvalidInput(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"null ID", "unknown ID", "empty ID", "unknown rule", "unknown at", "unconfigured"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			state := alwaysOnState(t)
			model := alwaysOnUpdateModel(t)
			client := mocks.NewAlwaysOnRuleClient(t)
			r := &alwaysOnRuleResource{client: &catoclient.Service{AccountId: "123"}, apiClient: client}
			switch mode {
			case "null ID", "unknown ID", "empty ID":
				id := map[string]types.String{"null ID": types.StringNull(), "unknown ID": types.StringUnknown(), "empty ID": types.StringValue("")}[mode]
				require.False(t, state.SetAttribute(ctx, path.Root("rule").AtName("id"), id).HasError())
			case "unknown rule":
				model.Rule = types.ObjectUnknown(AlwaysOnRuleDataAttrTypes)
			case "unknown at":
				model.At = types.ObjectUnknown(policyrule.PositionAttrTypes)
			case "unconfigured":
				r.client, r.apiClient = nil, nil
			}
			update := resource.UpdateResponse{State: state}
			r.Update(ctx, resource.UpdateRequest{Plan: alwaysOnPlan(t, model), State: state}, &update)
			require.True(t, update.Diagnostics.HasError())
			require.True(t, update.State.Raw.Equal(state.Raw))
			if mode != "unknown rule" && mode != "unknown at" {
				remove := resource.DeleteResponse{State: state}
				r.Delete(ctx, resource.DeleteRequest{State: state}, &remove)
				require.True(t, remove.Diagnostics.HasError())
				require.True(t, remove.State.Raw.Equal(state.Raw))
			}
		})
	}
}

func TestAlwaysOnRuleUpdateInvalidReference(t *testing.T) {
	t.Parallel()
	model := alwaysOnUpdateModel(t)
	attrs := model.Rule.Attributes()
	attrs["device_posture_profile"] = types.SetValueMust(idname.RefObjectType, []attr.Value{
		types.ObjectValueMust(idname.RefAttrTypes, map[string]attr.Value{"id": types.StringNull(), "name": types.StringNull()}),
	})
	model.Rule = types.ObjectValueMust(AlwaysOnRuleDataAttrTypes, attrs)
	client := mocks.NewAlwaysOnRuleClient(t)
	r := &alwaysOnRuleResource{client: &catoclient.Service{AccountId: "123"}, apiClient: client}
	state := alwaysOnState(t)
	resp := resource.UpdateResponse{State: state}
	r.Update(context.Background(), resource.UpdateRequest{Plan: alwaysOnPlan(t, model), State: state}, &resp)
	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "non-empty ID or name")
}
