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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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

func alwaysOnSchema(t *testing.T) schema.Schema {
	t.Helper()
	resp := resource.SchemaResponse{}
	NewAlwaysOnRuleResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	return resp.Schema
}

func alwaysOnPlanModel(t *testing.T) AlwaysOnRule {
	t.Helper()
	ctx := context.Background()
	duration, diags := types.ObjectValueFrom(ctx, AlwaysOnRuleBypassDurationAttrTypes, AlwaysOnRuleBypassDuration{
		Time: types.Int64Value(30), Unit: types.StringValue("MINUTES"),
	})
	require.False(t, diags.HasError(), "%v", diags)
	source, diags := types.ObjectValueFrom(ctx, AlwaysOnRuleSourceAttrTypes, AlwaysOnRuleSource{
		Users: types.SetValueMust(idname.RefObjectType, nil), UserGroups: types.SetValueMust(idname.RefObjectType, nil),
	})
	require.False(t, diags.HasError(), "%v", diags)
	rule, diags := types.ObjectValueFrom(ctx, AlwaysOnRuleDataAttrTypes, AlwaysOnRuleData{
		Action: types.StringValue("ENFORCE"), AllowFailOpen: types.BoolValue(false),
		AllowUserBypass: types.BoolValue(true), AntiTamperMode: types.StringValue("OFF"),
		BypassDuration: duration, Description: types.StringValue("Test rule"),
		DevicePostureProfile: types.SetValueMust(idname.RefObjectType, nil), Enabled: types.BoolValue(true),
		ID: types.StringUnknown(), Index: types.Int64Unknown(), Name: types.StringValue("Always on test"),
		Platform: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("WINDOWS")}),
		Section:  types.ObjectUnknown(AlwaysOnRuleSectionAttrTypes), Source: source,
	})
	require.False(t, diags.HasError(), "%v", diags)
	return AlwaysOnRule{At: types.ObjectNull(policyrule.PositionAttrTypes), Rule: rule}
}

func alwaysOnPlan(t *testing.T, model AlwaysOnRule) tfsdk.Plan {
	t.Helper()
	plan := tfsdk.Plan{Schema: alwaysOnSchema(t)}
	diags := plan.Set(context.Background(), model)
	require.False(t, diags.HasError(), "%v", diags)
	return plan
}

func alwaysOnResponse(t *testing.T) *cato.PolicyZtnaAlwaysOnAddRule {
	t.Helper()
	var response cato.PolicyZtnaAlwaysOnAddRule
	require.NoError(t, json.Unmarshal([]byte(`{"policy":{"ztnaAlwaysOn":{"addRule":{
		"status":"SUCCESS","errors":[],"rule":{"rule":{
			"id":"rule-123","index":4,"name":"Always on test","description":"Test rule",
			"enabled":true,"action":"ENFORCE","allowFailOpen":false,"allowUserBypass":true,"antiTamperMode":"OFF",
			"bypassDuration":{"time":30,"unit":"MINUTES"},"devicePostureProfile":[],"platform":["WINDOWS"],
			"source":{"user":[],"usersGroup":[]},"section":{"id":"section-1","name":"Default"}
		}}
	}}}}`), &response))
	return &response
}

func TestAlwaysOnRuleCreate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := mocks.NewAlwaysOnRuleClient(t)
	client.On("PolicyZtnaAlwaysOnAddRule", mock.Anything, "123", (*cato_models.ZtnaAlwaysOnPolicyMutationInput)(nil),
		mock.MatchedBy(func(input cato_models.ZtnaAlwaysOnAddRuleInput) bool {
			require.Nil(t, input.At)
			require.Equal(t, "Always on test", input.Rule.Name)
			require.Equal(t, cato_models.ZtnaAlwaysOnRuleActionEnum("ENFORCE"), input.Rule.Action)
			require.Equal(t, "Test rule", input.Rule.Description)
			require.True(t, input.Rule.Enabled)
			require.False(t, input.Rule.AllowFailOpen)
			require.True(t, input.Rule.AllowUserBypass)
			require.Equal(t, cato_models.AntiTamperModeEnum("OFF"), *input.Rule.AntiTamperMode)
			require.Equal(t, int64(30), input.Rule.BypassDuration.Time)
			require.Equal(t, cato_models.ZtnaAlwaysOnTimeUnit("MINUTES"), input.Rule.BypassDuration.Unit)
			require.Equal(t, []cato_models.OperatingSystem{"WINDOWS"}, input.Rule.Platform)
			require.NotNil(t, input.Rule.DevicePostureProfile)
			require.Empty(t, input.Rule.DevicePostureProfile)
			require.NotNil(t, input.Rule.Source.User)
			require.NotNil(t, input.Rule.Source.UsersGroup)
			return true
		})).Return(alwaysOnResponse(t), nil).Once()
	r := &alwaysOnRuleResource{client: &catoclient.Service{AccountId: "123"}, apiClient: client}
	plan := alwaysOnPlan(t, alwaysOnPlanModel(t))
	resp := resource.CreateResponse{State: tfsdk.State{Schema: plan.Schema}}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, &resp)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	var state AlwaysOnRule
	require.False(t, resp.State.Get(ctx, &state).HasError())
	var rule AlwaysOnRuleData
	require.False(t, state.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{}).HasError())
	require.Equal(t, "rule-123", rule.ID.ValueString())
	require.Equal(t, int64(4), rule.Index.ValueInt64())
	require.Equal(t, "section-1", rule.Section.Attributes()["id"].(types.String).ValueString())
	require.Equal(t, "Default", rule.Section.Attributes()["name"].(types.String).ValueString())
	require.True(t, state.At.IsNull())
	// All configured values survive response hydration unchanged.
	plannedAttributes := alwaysOnPlanModel(t).Rule.Attributes()
	for name, value := range plannedAttributes {
		if name != "id" && name != "index" && name != "section" {
			require.True(t, value.Equal(state.Rule.Attributes()[name]), "%s: %s", name, state.Rule.Attributes()[name])
		}
	}
}

func TestAlwaysOnRuleCreateErrors(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name    string
		body    string
		err     error
		message string
	}{
		{name: "transport", err: errors.New("request failed"), message: "request failed"},
		{name: "nil response", message: "API did not return resource ID"},
		{name: "missing policy", body: `{}`, message: "API did not return resource ID"},
		{name: "missing rule", body: `{"policy":{"ztnaAlwaysOn":{"addRule":{"status":"SUCCESS"}}}}`, message: "API did not return resource ID"},
		{name: "empty ID", body: `{"policy":{"ztnaAlwaysOn":{"addRule":{"status":"SUCCESS","rule":{"rule":{}}}}}}`, message: "API did not return resource ID"},
		{name: "mutation failure", body: `{"policy":{"ztnaAlwaysOn":{"addRule":{"status":"FAILURE","errors":[{"errorCode":"invalidRule","errorMessage":"Rule rejected"}]}}}}`, message: "Rule rejected"},
		{name: "failure with ID", body: `{"policy":{"ztnaAlwaysOn":{"addRule":{"status":"FAILURE","rule":{"rule":{"id":"rule-123"}}}}}}`, message: "FAILURE"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var body *cato.PolicyZtnaAlwaysOnAddRule
			if tc.body != "" {
				require.NoError(t, json.Unmarshal([]byte(tc.body), &body))
			}
			client := mocks.NewAlwaysOnRuleClient(t)
			client.On("PolicyZtnaAlwaysOnAddRule", mock.Anything, "123", mock.Anything, mock.Anything).Return(body, tc.err).Once()
			r := &alwaysOnRuleResource{client: &catoclient.Service{AccountId: "123"}, apiClient: client}
			plan := alwaysOnPlan(t, alwaysOnPlanModel(t))
			resp := resource.CreateResponse{State: tfsdk.State{Schema: plan.Schema}}
			r.Create(context.Background(), resource.CreateRequest{Plan: plan}, &resp)
			require.True(t, resp.Diagnostics.HasError())
			require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), tc.message)
			require.True(t, resp.State.Raw.IsNull())
		})
	}
}

func TestAlwaysOnRuleCreateInvalidPlan(t *testing.T) {
	t.Parallel()
	client := mocks.NewAlwaysOnRuleClient(t)
	r := &alwaysOnRuleResource{client: &catoclient.Service{AccountId: "123"}, apiClient: client}
	model := alwaysOnPlanModel(t)
	model.Rule = types.ObjectUnknown(AlwaysOnRuleDataAttrTypes)
	plan := alwaysOnPlan(t, model)
	resp := resource.CreateResponse{State: tfsdk.State{Schema: plan.Schema}}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan}, &resp)
	require.True(t, resp.Diagnostics.HasError())
}

func TestAlwaysOnRuleCreateInvalidReference(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := mocks.NewAlwaysOnRuleClient(t)
	r := &alwaysOnRuleResource{client: &catoclient.Service{AccountId: "123"}, apiClient: client}
	model := alwaysOnPlanModel(t)
	var rule AlwaysOnRuleData
	require.False(t, model.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{}).HasError())
	rule.DevicePostureProfile = types.SetValueMust(idname.RefObjectType, []attr.Value{
		types.ObjectValueMust(idname.RefAttrTypes, map[string]attr.Value{"id": types.StringNull(), "name": types.StringNull()}),
	})
	var diags diag.Diagnostics
	model.Rule, diags = types.ObjectValueFrom(ctx, AlwaysOnRuleDataAttrTypes, rule)
	require.False(t, diags.HasError())
	plan := alwaysOnPlan(t, model)
	resp := resource.CreateResponse{State: tfsdk.State{Schema: plan.Schema}}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, &resp)
	require.True(t, resp.Diagnostics.HasError())
	require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "non-empty ID or name")
	// The generated mock has no expectations: invalid input must never call the API.
}

func TestAlwaysOnRuleCreateUnconfigured(t *testing.T) {
	t.Parallel()
	plan := alwaysOnPlan(t, alwaysOnPlanModel(t))
	for _, service := range []*catoclient.Service{nil, {AccountId: "123"}} {
		r := &alwaysOnRuleResource{client: service}
		resp := resource.CreateResponse{State: tfsdk.State{Schema: plan.Schema}}
		r.Create(context.Background(), resource.CreateRequest{Plan: plan}, &resp)
		require.True(t, resp.Diagnostics.HasError())
		require.Equal(t, "Resource not configured", resp.Diagnostics.Errors()[0].Summary())
	}
}

func TestAlwaysOnRuleMetadataConfigureAndImport(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := NewAlwaysOnRuleResource().(*alwaysOnRuleResource)
	meta := resource.MetadataResponse{}
	r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "cato"}, &meta)
	require.Equal(t, "cato_always_on_rule", meta.TypeName)
	require.Nil(t, r.getClient())
	configure := resource.ConfigureResponse{}
	r.Configure(ctx, resource.ConfigureRequest{}, &configure)
	require.Nil(t, r.client)
	service := &catoclient.Service{AccountId: "123"}
	r.Configure(ctx, resource.ConfigureRequest{ProviderData: service}, &configure)
	require.Same(t, service, r.client)
	require.Nil(t, r.getClient())
	service.Catov2 = catoclient.NewProviderSDKClient(&cato.Client{})
	r.Configure(ctx, resource.ConfigureRequest{ProviderData: "invalid"}, &configure)
	require.True(t, configure.Diagnostics.HasError())
	resp := resource.ImportStateResponse{State: tfsdk.State{Schema: alwaysOnSchema(t)}}
	require.False(t, resp.State.Set(ctx, AlwaysOnRule{
		At: types.ObjectNull(policyrule.PositionAttrTypes), Rule: types.ObjectNull(AlwaysOnRuleDataAttrTypes),
	}).HasError())
	r.ImportState(ctx, resource.ImportStateRequest{ID: "rule-123"}, &resp)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	var id string
	require.False(t, resp.State.GetAttribute(ctx, path.Root("rule").AtName("id"), &id).HasError())
	require.Equal(t, "rule-123", id)
}

func TestAlwaysOnRuleSchema(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	resourceSchema := alwaysOnSchema(t)
	diags := resourceSchema.ValidateImplementation(ctx)
	require.False(t, diags.HasError(), "%v", diags)
}
