package ifrule

import (
	"context"
	"testing"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

func TestInternetFwRuleAppCategoryPlan(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ref := func(id, name types.String) attr.Value {
		return types.ObjectValueMust(idname.RefAttrTypes, map[string]attr.Value{"id": id, "name": name})
	}
	configRefs := []attr.Value{
		ref(types.StringValue("suspected_malware"), types.StringNull()),
		ref(types.StringNull(), types.StringValue("Alcohol and Tobacco")),
	}
	planRefs := []attr.Value{
		ref(types.StringValue("suspected_malware"), types.StringUnknown()),
		ref(types.StringUnknown(), types.StringValue("Alcohol and Tobacco")),
	}
	state := types.SetValueMust(idname.RefObjectType, []attr.Value{
		ref(types.StringValue("food_drinks_tobacco"), types.StringValue("Alcohol and Tobacco")),
		ref(types.StringValue("suspected_malware"), types.StringValue("Malware")),
	})
	resourceSchema := GetInternetFwRuleSchema(ctx, t)
	rule := resourceSchema.Attributes["rule"].(schema.SingleNestedAttribute)
	destination := rule.Attributes["destination"].(schema.SingleNestedAttribute)
	categories := destination.Attributes["app_category"].(schema.SetNestedAttribute)
	for _, tc := range []struct {
		name    string
		overlap bool
	}{
		{name: "matching references resolve from state"},
		{name: "overlapping ID and name are rejected", overlap: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			config := types.SetValueMust(idname.RefObjectType, configRefs)
			plan := types.SetValueMust(idname.RefObjectType, planRefs)
			if tc.overlap {
				config = types.SetValueMust(idname.RefObjectType, append(config.Elements(), ref(types.StringNull(), types.StringValue("Malware"))))
				plan = types.SetValueMust(idname.RefObjectType, append(plan.Elements(), ref(types.StringUnknown(), types.StringValue("Malware"))))
			}
			resourcePlan := tfsdk.Plan{Schema: resourceSchema}
			require.False(t, resourcePlan.Set(ctx, NewMinimalInternetFwRuleModel("rule-123")).HasError())
			resp := planmodifier.SetResponse{PlanValue: plan}
			for _, modifier := range categories.PlanModifiers {
				modifier.PlanModifySet(ctx, planmodifier.SetRequest{
					ConfigValue: config, PlanValue: resp.PlanValue, StateValue: state,
					Plan: resourcePlan, State: tfsdk.State(resourcePlan),
				}, &resp)
				if tc.overlap {
					require.True(t, resp.Diagnostics.HasError())
					require.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "Duplicate entry in the config")
					require.True(t, plan.Equal(resp.PlanValue), "invalid configuration must preserve the incoming plan")
					return
				}
				require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			}
			require.True(t, resp.PlanValue.Equal(state), "got %s, want %s", resp.PlanValue, state)
		})
	}
}

// TestHydrateIfwRuleStateExceptionEmptyServiceStandard reproduces the
// "Provider produced inconsistent result after apply" error on
// .rule.exceptions. An exception with no service (e.g. the "Bypassed
// Applications" exception on "Block QUIC services") is represented in the
// plan with service.standard as an EMPTY set (the module/plan convention for
// absent Set fields, matching the top-level rule.service hydration). Because
// `exceptions` is a SetNestedAttribute, if hydration returns service.standard
// as null instead of an empty set, the planned set element fails to correlate
// with the hydrated element and Terraform core rejects the apply.
func TestHydrateIfwRuleStateExceptionEmptyServiceStandard(t *testing.T) {
	ctx := context.Background()
	apiRule := MinimalAPIRule("test-ifw-rule", 1)
	apiRule.Exceptions = []*cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_Rule_Exceptions{
		{
			Name:             "Bypassed Applications",
			ConnectionOrigin: cato_models.ConnectionOriginEnum("ANY"),
			Destination: cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_Rule_Exceptions_Destination{
				Domain: []string{"lucidchart.com", "lucid.app"},
			},
			// Service left zero-value: Standard and Custom are nil (no service).
		},
	}

	got := HydrateIfwRuleState(ctx, NewMinimalInternetFwRuleModel("rule-123"), &apiRule)

	require.False(t, got.Exceptions.IsNull() || got.Exceptions.IsUnknown(),
		"exceptions must not be null/unknown: %s", got.Exceptions)
	elems := got.Exceptions.Elements()
	require.Len(t, elems, 1)

	service := elems[0].(types.Object).Attributes()["service"].(types.Object)
	standard := service.Attributes()["standard"].(types.Set)
	require.False(t, standard.IsNull(),
		"exception service.standard must hydrate to an empty set, not null, to match the plan")
	require.Equal(t, 0, len(standard.Elements()),
		"exception service.standard should be empty when the API returns no standard services")
}

func TestHydrateIfwRuleStateDuplicateAppCategories(t *testing.T) {
	apiRule := MinimalAPIRule("test-ifw-rule", 1)
	apiRule.Destination.AppCategory = []*cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_AppCategory{
		{ID: "suspected_malware", Name: "Malware"},
		{ID: "food_drinks_tobacco", Name: "Alcohol and Tobacco"},
		{ID: "food_drinks_tobacco", Name: "Alcohol and Tobacco"},
	}
	got := HydrateIfwRuleState(context.Background(), NewMinimalInternetFwRuleModel("rule-123"), &apiRule)
	categories := got.Destination.Attributes()["app_category"]
	if categories.IsNull() || categories.IsUnknown() {
		t.Fatalf("duplicate API categories must not erase destination categories: %s", categories)
	}
	require.Equal(t, 2, len(categories.(types.Set).Elements()))
}

func TestInternetFwRuleDuplicateAppCategoriesLifecycle(t *testing.T) {
	for _, operation := range []string{"create", "read", "update"} {
		t.Run(operation, func(t *testing.T) {
			ctx := context.Background()
			ref := func(id, name types.String) attr.Value {
				return types.ObjectValueMust(idname.RefAttrTypes, map[string]attr.Value{"id": id, "name": name})
			}
			model := NewMinimalInternetFwRuleModel("rule-123")
			attrs := model.Rule.Attributes()
			destination := emptyIfwDestinationObject().Attributes()
			destination["app_category"] = types.SetValueMust(idname.RefObjectType, []attr.Value{
				ref(types.StringValue("suspected_malware"), types.StringUnknown()),
				ref(types.StringUnknown(), types.StringValue("Alcohol and Tobacco")),
			})
			attrs["destination"] = types.ObjectValueMust(IfwDestAttrTypes, destination)
			model.Rule = types.ObjectValueMust(InternetFirewallRuleRuleAttrTypes, attrs)
			plan := tfsdk.Plan{Schema: GetInternetFwRuleSchema(ctx, t)}
			require.False(t, plan.Set(ctx, model).HasError())
			apiRule := MinimalAPIRule("test-ifw-rule", 1)
			apiRule.Destination.AppCategory = []*cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_AppCategory{
				{ID: "suspected_malware", Name: "Malware"},
				{ID: "food_drinks_tobacco", Name: "Alcohol and Tobacco"},
				{ID: "food_drinks_tobacco", Name: "Alcohol and Tobacco"},
			}
			mockClient := mocks.NewInternetFirewallPolicyClient(t)
			checkInput := func(categories []*cato_models.ApplicationCategoryRefInput) bool {
				require.Len(t, categories, 2)
				require.Equal(t, cato_models.ObjectRefBy("ID"), categories[0].By)
				require.Equal(t, "suspected_malware", categories[0].Input)
				require.Equal(t, cato_models.ObjectRefBy("NAME"), categories[1].By)
				require.Equal(t, "Alcohol and Tobacco", categories[1].Input)
				return true
			}
			switch operation {
			case "create":
				mockClient.EXPECT().PolicyInternetFirewallAddRule(mock.Anything, mock.MatchedBy(func(input cato_models.InternetFirewallAddRuleInput) bool {
					return checkInput(input.Rule.Destination.AppCategory)
				}), "account-123").Return(SuccessfulAddRuleResponse("rule-123"), nil).Once()
			case "update":
				// plan and state share the same `at`, so no MoveRule call is expected
				mockClient.EXPECT().PolicyInternetFirewallUpdateRule(mock.Anything, mock.Anything, mock.MatchedBy(func(input cato_models.InternetFirewallUpdateRuleInput) bool {
					return checkInput(input.Rule.Destination.AppCategory)
				}), "account-123").Return(SuccessfulUpdateRuleResponse("rule-123"), nil).Once()
			}
			if operation != "read" {
				mockClient.EXPECT().PolicyInternetFirewallPublishPolicyRevision(mock.Anything, mock.Anything, mock.Anything, "account-123").Return(nil, nil).Once()
			}
			mockClient.EXPECT().PolicyInternetFirewall(mock.Anything, mock.Anything, "account-123").Return(InternetFirewallPolicyResponseWithRule(apiRule), nil).Once()
			r := &InternetFwRuleResource{Client: &catoclient.Service{AccountId: "account-123"}, IfwClient: mockClient}
			state := newInternetFwRuleStateWithID(ctx, t)
			result := tfsdk.State{Schema: plan.Schema}
			switch operation {
			case "create":
				resp := &resource.CreateResponse{State: result}
				r.Create(ctx, resource.CreateRequest{Plan: plan}, resp)
				require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
				result = resp.State
			case "read":
				resp := &resource.ReadResponse{State: state}
				r.Read(ctx, resource.ReadRequest{State: state}, resp)
				require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
				result = resp.State
			case "update":
				resp := &resource.UpdateResponse{State: result}
				r.Update(ctx, resource.UpdateRequest{Plan: plan, State: state}, resp)
				require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
				result = resp.State
			}
			require.False(t, result.Get(ctx, &model).HasError())
			got := model.Rule.Attributes()["destination"].(types.Object).Attributes()["app_category"].(types.Set)
			want := types.SetValueMust(idname.RefObjectType, []attr.Value{
				ref(types.StringValue("suspected_malware"), types.StringValue("Malware")),
				ref(types.StringValue("food_drinks_tobacco"), types.StringValue("Alcohol and Tobacco")),
			})
			require.True(t, got.Equal(want), "got %s, want %s", got, want)
		})
	}
}
