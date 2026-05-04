package provider

import (
	"context"
	"testing"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/mock"
)

func TestIfwRulesIndexResourceCreateUsesReorderPolicy(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewBulkInternetFirewallPolicyClient(t)

	mockClient.EXPECT().
		PolicyInternetFirewallSectionsIndex(mock.Anything, "account-123").
		Return(ifwBulkSectionsResponse([]bulkPolicySection{
			{ID: "section-existing", Name: "Existing"},
			{ID: "section-alpha", Name: "Alpha"},
			{ID: "section-beta", Name: "Beta"},
		}), nil).
		Once()
	mockClient.EXPECT().
		PolicyInternetFirewallRulesIndex(mock.Anything, "account-123").
		Return(ifwBulkRulesResponse([]bulkPolicyRule{
			{ID: "rule-one", Name: "Rule One", SectionID: "section-beta", SectionName: "Beta", Description: "from-api", Enabled: true, Index: 1},
		}), nil).
		Once()
	mockClient.EXPECT().
		PolicyInternetFirewallReorderPolicy(
			mock.Anything,
			mock.Anything,
			mock.MatchedBy(func(input cato_models.PolicyReorderInput) bool {
				return stringSlicesEqual(reorderSectionIDs(input), []string{"section-existing", "section-alpha", "section-beta"}) &&
					stringSlicesEqual(reorderRuleIDs(input, "section-alpha"), []string{"rule-one"}) &&
					stringSlicesEqual(reorderRuleIDs(input, "section-beta"), []string{})
			}),
			"account-123",
		).
		Return(successfulIfwReorderResponse(), nil).
		Once()
	mockClient.EXPECT().
		PolicyInternetFirewallPublishPolicyRevision(mock.Anything, mock.Anything, mock.Anything, "account-123").
		Return(nil, nil).
		Once()

	r := &ifwRulesIndexResource{
		client:    &catoClientData{AccountId: "account-123"},
		ifwClient: mockClient,
	}
	req := resource.CreateRequest{Plan: newIfwRulesIndexPlan(ctx, t)}
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: getIfwRulesIndexSchema(ctx, t)}}

	r.Create(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}

	var state IfwRulesIndex
	diags := resp.State.Get(ctx, &state)
	if diags.HasError() {
		t.Fatalf("unexpected state diagnostics: %+v", diags)
	}
	if got := state.SectionData.Elements()["Alpha"].(types.Object).Attributes()["id"].(types.String).ValueString(); got != "section-alpha" {
		t.Fatalf("expected Alpha section id section-alpha, got %q", got)
	}
	if got := state.RuleData.Elements()["Rule One"].(types.Object).Attributes()["id"].(types.String).ValueString(); got != "rule-one" {
		t.Fatalf("expected Rule One id rule-one, got %q", got)
	}
}

func TestWanRulesIndexResourceCreateUsesReorderPolicy(t *testing.T) {
	ctx := context.Background()
	mockClient := mocks.NewBulkWanFirewallPolicyClient(t)

	mockClient.EXPECT().
		PolicyWanFirewallSectionsIndex(mock.Anything, "account-123").
		Return(wanBulkSectionsResponse([]bulkPolicySection{
			{ID: "section-existing", Name: "Existing"},
			{ID: "section-alpha", Name: "Alpha"},
			{ID: "section-beta", Name: "Beta"},
		}), nil).
		Once()
	mockClient.EXPECT().
		PolicyWanFirewallRulesIndex(mock.Anything, "account-123").
		Return(wanBulkRulesResponse([]bulkPolicyRule{
			{ID: "rule-one", Name: "Rule One", SectionID: "section-beta", SectionName: "Beta", Description: "from-api", Enabled: true, Index: 1},
		}), nil).
		Once()
	mockClient.EXPECT().
		PolicyWanFirewallReorderPolicy(
			mock.Anything,
			mock.Anything,
			mock.MatchedBy(func(input cato_models.PolicyReorderInput) bool {
				return stringSlicesEqual(reorderSectionIDs(input), []string{"section-existing", "section-alpha", "section-beta"}) &&
					stringSlicesEqual(reorderRuleIDs(input, "section-alpha"), []string{"rule-one"}) &&
					stringSlicesEqual(reorderRuleIDs(input, "section-beta"), []string{})
			}),
			"account-123",
		).
		Return(successfulWanReorderResponse(), nil).
		Once()
	mockClient.EXPECT().
		PolicyWanFirewallPublishPolicyRevision(mock.Anything, mock.Anything, "account-123").
		Return(nil, nil).
		Once()

	r := &wanRulesIndexResource{
		client:    &catoClientData{AccountId: "account-123"},
		wanClient: mockClient,
	}
	req := resource.CreateRequest{Plan: newWanRulesIndexPlan(ctx, t)}
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: getWanRulesIndexSchema(ctx, t)}}

	r.Create(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}

	var state WanRulesIndex
	diags := resp.State.Get(ctx, &state)
	if diags.HasError() {
		t.Fatalf("unexpected state diagnostics: %+v", diags)
	}
	if got := state.SectionData.Elements()["Alpha"].(types.Object).Attributes()["id"].(types.String).ValueString(); got != "section-alpha" {
		t.Fatalf("expected Alpha section id section-alpha, got %q", got)
	}
	if got := state.RuleData.Elements()["Rule One"].(types.Object).Attributes()["id"].(types.String).ValueString(); got != "rule-one" {
		t.Fatalf("expected Rule One id rule-one, got %q", got)
	}
}

func getIfwRulesIndexSchema(ctx context.Context, t *testing.T) schema.Schema {
	t.Helper()

	r := &ifwRulesIndexResource{}
	resp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, resp)

	return resp.Schema
}

func getWanRulesIndexSchema(ctx context.Context, t *testing.T) schema.Schema {
	t.Helper()

	r := &wanRulesIndexResource{}
	resp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, resp)

	return resp.Schema
}

func newIfwRulesIndexPlan(ctx context.Context, t *testing.T) tfsdk.Plan {
	t.Helper()

	plan := tfsdk.Plan{Schema: getIfwRulesIndexSchema(ctx, t)}
	diags := plan.Set(ctx, IfwRulesIndex{
		SectionToStartAfterId: types.StringValue("section-existing"),
		SectionData:           bulkSectionPlanMap(IfwSectionIndexResourceAttrTypes),
		RuleData:              bulkRulePlanMap(IfwRuleIndexResourceAttrTypes),
	})
	if diags.HasError() {
		t.Fatalf("unexpected plan diagnostics: %+v", diags)
	}

	return plan
}

func newWanRulesIndexPlan(ctx context.Context, t *testing.T) tfsdk.Plan {
	t.Helper()

	plan := tfsdk.Plan{Schema: getWanRulesIndexSchema(ctx, t)}
	diags := plan.Set(ctx, WanRulesIndex{
		SectionToStartAfterId: types.StringValue("section-existing"),
		SectionData:           bulkSectionPlanMap(WanSectionIndexResourceAttrTypes),
		RuleData:              bulkRulePlanMap(WanRuleIndexResourceAttrTypes),
	})
	if diags.HasError() {
		t.Fatalf("unexpected plan diagnostics: %+v", diags)
	}

	return plan
}

func bulkSectionPlanMap(attrTypes map[string]attr.Type) types.Map {
	return types.MapValueMust(types.ObjectType{AttrTypes: attrTypes}, map[string]attr.Value{
		"Alpha": types.ObjectValueMust(attrTypes, map[string]attr.Value{
			"id":            types.StringNull(),
			"section_name":  types.StringValue("Alpha"),
			"section_index": types.Int64Value(1),
		}),
		"Beta": types.ObjectValueMust(attrTypes, map[string]attr.Value{
			"id":            types.StringNull(),
			"section_name":  types.StringValue("Beta"),
			"section_index": types.Int64Value(2),
		}),
	})
}

func bulkRulePlanMap(attrTypes map[string]attr.Type) types.Map {
	return types.MapValueMust(types.ObjectType{AttrTypes: attrTypes}, map[string]attr.Value{
		"Rule One": types.ObjectValueMust(attrTypes, map[string]attr.Value{
			"id":               types.StringNull(),
			"index_in_section": types.Int64Value(1),
			"section_name":     types.StringValue("Alpha"),
			"rule_name":        types.StringValue("Rule One"),
			"description":      types.StringValue("from-plan"),
			"enabled":          types.BoolValue(true),
		}),
	})
}

func ifwBulkSectionsResponse(sections []bulkPolicySection) *cato_go_sdk.IfwSectionsIndexPolicy {
	items := make([]*cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Sections, 0, len(sections))
	for _, section := range sections {
		items = append(items, &cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Sections{
			Properties: section.Properties,
			Section: cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Sections_Section{
				ID:   section.ID,
				Name: section.Name,
			},
		})
	}

	return &cato_go_sdk.IfwSectionsIndexPolicy{
		Policy: &cato_go_sdk.IfwSectionsIndexPolicy_Policy{
			InternetFirewall: &cato_go_sdk.IfwSectionsIndexPolicy_Policy_Policy{
				Policy: cato_go_sdk.Policy_Policy_IfwSectionsIndexPolicy_Policy_Policy{
					Sections: items,
				},
			},
		},
	}
}

func ifwBulkRulesResponse(rules []bulkPolicyRule) *cato_go_sdk.IfwRulesIndexPolicy {
	items := make([]*cato_go_sdk.Policy_PIfwRulesIndexPolicy_Policy_Policy_Rules, 0, len(rules))
	for _, rule := range rules {
		items = append(items, &cato_go_sdk.Policy_PIfwRulesIndexPolicy_Policy_Policy_Rules{
			Properties: rule.Properties,
			Rule: cato_go_sdk.Policy_PIfwRulesIndexPolicy_Policy_Policy_Rules_Rule{
				ID:          rule.ID,
				Name:        rule.Name,
				Description: rule.Description,
				Enabled:     rule.Enabled,
				Index:       rule.Index,
				Section: cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_Rule_Section{
					ID:   rule.SectionID,
					Name: rule.SectionName,
				},
			},
		})
	}

	return &cato_go_sdk.IfwRulesIndexPolicy{
		Policy: &cato_go_sdk.IfwRulesIndexPolicy_Policy{
			InternetFirewall: &cato_go_sdk.IfwRulesIndexPolicy_Policy_Policy{
				Policy: cato_go_sdk.Policy_PIfwRulesIndexPolicy_Policy_Policy{
					Rules: items,
				},
			},
		},
	}
}

func wanBulkSectionsResponse(sections []bulkPolicySection) *cato_go_sdk.WanSectionsIndexPolicy {
	items := make([]*cato_go_sdk.Policy_Policy_WanFirewall_Policy_Sections, 0, len(sections))
	for _, section := range sections {
		items = append(items, &cato_go_sdk.Policy_Policy_WanFirewall_Policy_Sections{
			Properties: section.Properties,
			Section: cato_go_sdk.Policy_Policy_WanFirewall_Policy_Sections_Section{
				ID:   section.ID,
				Name: section.Name,
			},
		})
	}

	return &cato_go_sdk.WanSectionsIndexPolicy{
		Policy: &cato_go_sdk.WanSectionsIndexPolicy_Policy{
			WanFirewall: &cato_go_sdk.Policy_Policy_WanFirewall{
				Policy: cato_go_sdk.Policy_Policy_WanFirewall_Policy{
					Sections: items,
				},
			},
		},
	}
}

func wanBulkRulesResponse(rules []bulkPolicyRule) *cato_go_sdk.WanRulesIndexPolicy {
	items := make([]*cato_go_sdk.WanRulesIndexPolicy_Policy_WanFirewall_Policy_Rules, 0, len(rules))
	for _, rule := range rules {
		items = append(items, &cato_go_sdk.WanRulesIndexPolicy_Policy_WanFirewall_Policy_Rules{
			Properties: rule.Properties,
			Rule: cato_go_sdk.WanRulesIndexPolicy_Policy_WanFirewall_Policy_Rules_Rule{
				ID:          rule.ID,
				Name:        rule.Name,
				Description: rule.Description,
				Enabled:     rule.Enabled,
				Index:       rule.Index,
				Section: cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules_Rule_Section{
					ID:   rule.SectionID,
					Name: rule.SectionName,
				},
			},
		})
	}

	return &cato_go_sdk.WanRulesIndexPolicy{
		Policy: &cato_go_sdk.WanRulesIndexPolicy_Policy{
			WanFirewall: &cato_go_sdk.WanRulesIndexPolicy_Policy_WanFirewall{
				Policy: cato_go_sdk.WanRulesIndexPolicy_Policy_WanFirewall_Policy{
					Rules: items,
				},
			},
		},
	}
}

func successfulIfwReorderResponse() *cato_go_sdk.PolicyInternetFirewallReorderPolicy {
	return &cato_go_sdk.PolicyInternetFirewallReorderPolicy{
		Policy: &cato_go_sdk.PolicyInternetFirewallReorderPolicy_Policy{
			InternetFirewall: &cato_go_sdk.PolicyInternetFirewallReorderPolicy_Policy_InternetFirewall{
				ReorderPolicy: cato_go_sdk.PolicyInternetFirewallReorderPolicy_Policy_InternetFirewall_ReorderPolicy{
					Status: cato_models.PolicyMutationStatusSuccess,
				},
			},
		},
	}
}

func successfulWanReorderResponse() *cato_go_sdk.PolicyWanFirewallReorderPolicy {
	return &cato_go_sdk.PolicyWanFirewallReorderPolicy{
		Policy: &cato_go_sdk.PolicyWanFirewallReorderPolicy_Policy{
			WanFirewall: &cato_go_sdk.PolicyWanFirewallReorderPolicy_Policy_WanFirewall{
				ReorderPolicy: cato_go_sdk.PolicyWanFirewallReorderPolicy_Policy_WanFirewall_ReorderPolicy{
					Status: cato_models.PolicyMutationStatusSuccess,
				},
			},
		},
	}
}

func stringSlicesEqual(got []string, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range want {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
