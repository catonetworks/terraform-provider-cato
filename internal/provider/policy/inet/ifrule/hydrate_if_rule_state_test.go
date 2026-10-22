package ifrule

import (
	"context"
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
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

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
				mockClient.EXPECT().PolicyInternetFirewallMoveRule(mock.Anything, mock.Anything, mock.Anything, "account-123").Return(successfulMoveRuleResponse("rule-123"), nil).Once()
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
