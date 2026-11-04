package ifsubpolicy

import (
	"context"
	"encoding/json"
	"testing"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/policy/inet/ifrule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/apperror"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
)

func TestIfSubPolicyUpdatePosition(t *testing.T) {
	for _, tc := range []struct {
		name         string
		oldPosition  string
		oldRef       string
		position     string
		ref          string
		movePayload  string
		nilResponse  bool
		moveError    error
		publishError error
		wantError    string
	}{
		{name: "unchanged", oldPosition: "LAST_IN_POLICY", position: "LAST_IN_POLICY"},
		{name: "first in policy", oldPosition: "LAST_IN_POLICY", position: "FIRST_IN_POLICY"},
		{name: "before rule", oldPosition: "LAST_IN_POLICY", position: "BEFORE_RULE", ref: "rule-2"},
		{name: "reference only", oldPosition: "BEFORE_RULE", oldRef: "rule-2", position: "BEFORE_RULE", ref: "rule-3"},
		{name: "clear reference", oldPosition: "BEFORE_RULE", oldRef: "rule-2", position: "LAST_IN_POLICY"},
		{name: "transport failure", position: "FIRST_IN_POLICY", moveError: apperror.AssertErr("move failed"), wantError: "move failed"},
		{name: "rejected move", position: "FIRST_IN_POLICY", movePayload: `{"status":"FAILURE","errors":[{"errorCode":"Rejected","errorMessage":"invalid target"}]}`, wantError: "invalid target"},
		{name: "failure without details", position: "FIRST_IN_POLICY", movePayload: `{"status":"FAILURE"}`, wantError: "FAILURE"},
		{name: "missing status", position: "FIRST_IN_POLICY", movePayload: `{}`, wantError: "policy mutation returned status"},
		{name: "missing response", position: "FIRST_IN_POLICY", nilResponse: true, wantError: "policy mutation returned status"},
		{name: "publish failure", position: "FIRST_IN_POLICY", publishError: apperror.AssertErr("publish failed"), wantError: "publish failed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			state := newIfSubPolicyStateWithID(ctx, t)
			var model InternetFirewallSubPolicy
			require.False(t, state.Get(ctx, &model).HasError())
			if tc.oldPosition != "" {
				model.At = subPolicyPosition(tc.oldPosition, tc.oldRef)
			}
			require.False(t, state.Set(ctx, model).HasError())
			oldAt := model.At
			model.At = subPolicyPosition(tc.position, tc.ref)
			plan := tfsdk.Plan{Schema: state.Schema}
			require.False(t, plan.Set(ctx, model).HasError())

			client := mocks.NewInternetFirewallSubPolicyClient(t)
			var calls []string
			client.On("PolicyInternetFirewallUpdateRule", mock.Anything, mock.Anything, mock.Anything, "account-123").
				Return(ifrule.SuccessfulUpdateRuleResponse("scope-1"), nil).
				Run(func(mock.Arguments) { calls = append(calls, "update") }).Once()
			wantCalls := []string{"update"}
			if !oldAt.Equal(model.At) {
				payload := tc.movePayload
				if payload == "" {
					payload = `{"status":"SUCCESS"}`
				}
				var moved cato_go_sdk.PolicyInternetFirewallMoveRule
				require.NoError(t, json.Unmarshal([]byte(`{"policy":{"internetFirewall":{"moveRule":`+payload+`}}}`), &moved))
				moveResponse := &moved
				if tc.nilResponse {
					moveResponse = nil
				}
				position := cato_models.PolicyRulePositionEnum(tc.position)
				var ref *string
				if tc.ref != "" {
					ref = &tc.ref
				}
				client.On("PolicyInternetFirewallMoveRule", mock.Anything,
					mock.MatchedBy(func(input *cato_models.InternetFirewallPolicyMutationInput) bool {
						return input != nil && *input == (cato_models.InternetFirewallPolicyMutationInput{})
					}), cato_models.PolicyMoveRuleInput{
						ID: "scope-1", To: &cato_models.PolicyRulePositionInput{Position: &position, Ref: ref},
					}, "account-123").Return(moveResponse, tc.moveError).
					Run(func(mock.Arguments) { calls = append(calls, "move") }).Once()
				wantCalls = append(wantCalls, "move")
			}
			moveFailed := tc.moveError != nil || tc.movePayload != "" || tc.nilResponse
			if !moveFailed {
				client.On("PolicyInternetFirewallPublishPolicyRevision", mock.Anything, mock.Anything, mock.Anything, "account-123").
					Return(successfulSubPolicyPublish(t), tc.publishError).
					Run(func(mock.Arguments) { calls = append(calls, "publish") }).Once()
				wantCalls = append(wantCalls, "publish")
				if tc.publishError == nil {
					client.On("PolicyInternetFirewall", mock.Anything, mock.Anything, "account-123").
						Return(ifSubPolicyResponse("test-sub", "a sub", "scope-name"), nil).
						Run(func(mock.Arguments) { calls = append(calls, "read") }).Once()
					wantCalls = append(wantCalls, "read")
				}
			}

			r := &ifSubPolicyResource{client: &catoclient.Service{AccountId: "account-123"}, subPolyClient: client}
			resp := &resource.UpdateResponse{State: state}
			r.Update(ctx, resource.UpdateRequest{Plan: plan, State: state}, resp)
			require.Equal(t, wantCalls, calls)
			if tc.wantError != "" {
				require.True(t, resp.Diagnostics.HasError())
				require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), tc.wantError)
				require.True(t, state.Raw.Equal(resp.State.Raw), "a failed move/publish must not record the requested position")
				return
			}
			require.False(t, resp.Diagnostics.HasError(), "%+v", resp.Diagnostics)
			var got InternetFirewallSubPolicy
			require.False(t, resp.State.Get(ctx, &got).HasError())
			require.Equal(t, "sub-1", got.ID.ValueString())
			require.Equal(t, "scope-1", got.ScopeRuleID.ValueString())
			require.Equal(t, "cleanup-1", got.CleanupRuleID.ValueString())
			require.True(t, model.At.Equal(got.At), "state must retain the successfully applied position")
		})
	}
}

func subPolicyPosition(position, ref string) types.Object {
	refValue := types.StringNull()
	if ref != "" {
		refValue = types.StringValue(ref)
	}
	return types.ObjectValueMust(policyrule.PositionAttrTypes, map[string]attr.Value{
		"position": types.StringValue(position),
		"ref":      refValue,
	})
}
