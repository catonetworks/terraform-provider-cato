package alwaysonrule

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
)

func alwaysOnState(t *testing.T) tfsdk.State {
	t.Helper()
	ctx := context.Background()
	var diags diag.Diagnostics
	model := alwaysOnPlanModel(t)
	model.Rule = parseAlwaysOnRule(ctx, alwaysOnResponse(t).Policy.ZtnaAlwaysOn.AddRule.Rule.GetRule(), &diags)
	require.False(t, diags.HasError(), "%v", diags)
	state := tfsdk.State{Schema: alwaysOnSchema(t)}
	require.False(t, state.Set(ctx, model).HasError())
	return state
}

func alwaysOnPolicyResponse(t *testing.T) *cato.PolicyZtnaAlwaysOnPolicy {
	t.Helper()
	var result cato.PolicyZtnaAlwaysOnPolicy
	require.NoError(t, json.Unmarshal([]byte(`{"policy":{"ztnaAlwaysOn":{"policy":{"rules":[
		{"rule":{"id":"other-rule","name":"Refreshed rule"}},
		{"rule":{
			"id":"rule-123","index":9,"name":"Refreshed rule","description":"Changed remotely",
			"enabled":false,"action":"IGNORE","allowFailOpen":true,"allowUserBypass":false,"antiTamperMode":"PROTECT",
			"bypassDuration":{"time":2,"unit":"HOURS"},
			"devicePostureProfile":[{"id":"profile-1","name":"Managed device"}],"platform":["MAC","WINDOWS"],
			"source":{"user":[{"id":"user-1","name":"Example user"}],"usersGroup":[{"id":"group-1","name":"Example group"}]},
			"section":{"id":"section-2","name":"Moved section"}
		}}
	]}}}}`), &result))
	return &result
}

func TestAlwaysOnRuleRead(t *testing.T) {
	t.Parallel()
	for _, imported := range []bool{false, true} {
		name := "existing state"
		if imported {
			name = "imported state"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			state := alwaysOnState(t)
			at := types.ObjectNull(policyrule.PositionAttrTypes)
			r := &alwaysOnRuleResource{client: &catoclient.Service{AccountId: "123"}}
			if imported {
				require.False(t, state.Set(ctx, AlwaysOnRule{At: at, Rule: types.ObjectNull(AlwaysOnRuleDataAttrTypes)}).HasError())
				resp := resource.ImportStateResponse{State: state}
				r.ImportState(ctx, resource.ImportStateRequest{ID: "rule-123"}, &resp)
				require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
				state = resp.State
			} else {
				var diags diag.Diagnostics
				at, diags = types.ObjectValueFrom(ctx, policyrule.PositionAttrTypes,
					policyrule.Position{Position: types.StringValue("LAST_IN_POLICY"), Ref: types.StringValue("")})
				require.False(t, diags.HasError())
				require.False(t, state.SetAttribute(ctx, path.Root("at"), at).HasError())
			}
			client := mocks.NewAlwaysOnRuleClient(t)
			client.On("PolicyZtnaAlwaysOnPolicy", mock.Anything, "123", (*cato_models.ZtnaAlwaysOnPolicyInput)(nil)).
				Return(alwaysOnPolicyResponse(t), nil).Once()
			r.apiClient = client
			resp := resource.ReadResponse{State: state}
			r.Read(ctx, resource.ReadRequest{State: state}, &resp)
			require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			var model AlwaysOnRule
			require.False(t, resp.State.Get(ctx, &model).HasError())
			require.True(t, at.Equal(model.At))
			var rule AlwaysOnRuleData
			require.False(t, model.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{}).HasError())
			require.Equal(t, "rule-123", rule.ID.ValueString())
			require.Equal(t, int64(9), rule.Index.ValueInt64())
			require.Equal(t, "Refreshed rule", rule.Name.ValueString())
			require.Equal(t, "Changed remotely", rule.Description.ValueString())
			require.Equal(t, "IGNORE", rule.Action.ValueString())
			require.Equal(t, "PROTECT", rule.AntiTamperMode.ValueString())
			require.False(t, rule.Enabled.ValueBool())
			require.True(t, rule.AllowFailOpen.ValueBool())
			require.False(t, rule.AllowUserBypass.ValueBool())
			var duration AlwaysOnRuleBypassDuration
			require.False(t, rule.BypassDuration.As(ctx, &duration, basetypes.ObjectAsOptions{}).HasError())
			require.Equal(t, int64(2), duration.Time.ValueInt64())
			require.Equal(t, "HOURS", duration.Unit.ValueString())
			var platforms []string
			require.False(t, rule.Platform.ElementsAs(ctx, &platforms, false).HasError())
			require.ElementsMatch(t, []string{"MAC", "WINDOWS"}, platforms)
			var section AlwaysOnRuleSection
			require.False(t, rule.Section.As(ctx, &section, basetypes.ObjectAsOptions{}).HasError())
			require.Equal(t, "section-2", section.ID.ValueString())
			require.Equal(t, "Moved section", section.Name.ValueString())
			var source AlwaysOnRuleSource
			require.False(t, rule.Source.As(ctx, &source, basetypes.ObjectAsOptions{}).HasError())
			for _, refs := range []struct {
				set      types.Set
				id, name string
			}{
				{rule.DevicePostureProfile, "profile-1", "Managed device"},
				{source.Users, "user-1", "Example user"},
				{source.UserGroups, "group-1", "Example group"},
			} {
				var values []idname.Ref
				require.False(t, refs.set.ElementsAs(ctx, &values, false).HasError())
				require.Equal(t, []idname.Ref{{ID: types.StringValue(refs.id), Name: types.StringValue(refs.name)}}, values)
			}
		})
	}
}

func TestAlwaysOnRuleReadMissing(t *testing.T) {
	t.Parallel()
	for _, empty := range []bool{false, true} {
		t.Run(map[bool]string{false: "other rules", true: "empty policy"}[empty], func(t *testing.T) {
			t.Parallel()
			result := alwaysOnPolicyResponse(t)
			if empty {
				result.Policy.ZtnaAlwaysOn.Policy.Rules = result.Policy.ZtnaAlwaysOn.Policy.Rules[:0]
			} else {
				result.Policy.ZtnaAlwaysOn.Policy.Rules = result.Policy.ZtnaAlwaysOn.Policy.Rules[:1]
			}
			client := mocks.NewAlwaysOnRuleClient(t)
			client.On("PolicyZtnaAlwaysOnPolicy", mock.Anything, "123", mock.Anything).Return(result, nil).Once()
			r := &alwaysOnRuleResource{client: &catoclient.Service{AccountId: "123"}, apiClient: client}
			state := alwaysOnState(t)
			resp := resource.ReadResponse{State: state}
			r.Read(context.Background(), resource.ReadRequest{State: state}, &resp)
			require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			require.True(t, resp.State.Raw.IsNull())
		})
	}
}

func TestAlwaysOnRuleReadErrorsPreserveState(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, body, message string
		err                 error
	}{
		{name: "transport", err: errors.New("request failed"), message: "request failed"},
		{name: "nil response", message: "API did not return an Always On policy"},
		{name: "missing policy", body: `{}`, message: "API did not return an Always On policy"},
		{name: "missing rules", body: `{"policy":{"ztnaAlwaysOn":{"policy":{}}}}`, message: "API did not return an Always On policy"},
		{name: "null rules", body: `{"policy":{"ztnaAlwaysOn":{"policy":{"rules":null}}}}`, message: "API did not return an Always On policy"},
		{name: "null entry", body: `{"policy":{"ztnaAlwaysOn":{"policy":{"rules":[null]}}}}`, message: "API returned an Always On rule without an ID"},
		{name: "missing rule", body: `{"policy":{"ztnaAlwaysOn":{"policy":{"rules":[{}]}}}}`, message: "API returned an Always On rule without an ID"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var result *cato.PolicyZtnaAlwaysOnPolicy
			if tc.body != "" {
				require.NoError(t, json.Unmarshal([]byte(tc.body), &result))
			}
			client := mocks.NewAlwaysOnRuleClient(t)
			client.On("PolicyZtnaAlwaysOnPolicy", mock.Anything, "123", mock.Anything).Return(result, tc.err).Once()
			r := &alwaysOnRuleResource{client: &catoclient.Service{AccountId: "123"}, apiClient: client}
			state := alwaysOnState(t)
			resp := resource.ReadResponse{State: state}
			r.Read(context.Background(), resource.ReadRequest{State: state}, &resp)
			require.True(t, resp.Diagnostics.HasError())
			require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), tc.message)
			require.True(t, resp.State.Raw.Equal(state.Raw))
		})
	}
}

func TestAlwaysOnRuleReadInvalidID(t *testing.T) {
	t.Parallel()
	for name, id := range map[string]types.String{"null": types.StringNull(), "unknown": types.StringUnknown(), "empty": types.StringValue("")} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			state := alwaysOnState(t)
			require.False(t, state.SetAttribute(ctx, path.Root("rule").AtName("id"), id).HasError())
			client := mocks.NewAlwaysOnRuleClient(t)
			r := &alwaysOnRuleResource{client: &catoclient.Service{AccountId: "123"}, apiClient: client}
			resp := resource.ReadResponse{State: state}
			r.Read(ctx, resource.ReadRequest{State: state}, &resp)
			require.True(t, resp.Diagnostics.HasError())
			require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "known, non-empty rule ID")
			require.True(t, resp.State.Raw.Equal(state.Raw))
		})
	}
}

func TestAlwaysOnRuleReadUnconfigured(t *testing.T) {
	t.Parallel()
	state := alwaysOnState(t)
	for _, service := range []*catoclient.Service{nil, {AccountId: "123"}} {
		r := &alwaysOnRuleResource{client: service}
		resp := resource.ReadResponse{State: state}
		r.Read(context.Background(), resource.ReadRequest{State: state}, &resp)
		require.True(t, resp.Diagnostics.HasError())
		require.Equal(t, "Resource not configured", resp.Diagnostics.Errors()[0].Summary())
		require.True(t, resp.State.Raw.Equal(state.Raw))
	}
}
