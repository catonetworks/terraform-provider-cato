package provider

import (
	"context"
	"errors"
	"strings"
	"testing"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/mock"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
)

func TestNewWanRulesIndexResource(t *testing.T) {
	t.Parallel()

	res := NewWanRulesIndexResource()
	if _, ok := res.(*wanRulesIndexResource); !ok {
		t.Fatalf("expected *wanRulesIndexResource, got %T", res)
	}
}

func TestWanRulesIndexResourceMetadata(t *testing.T) {
	t.Parallel()

	r := &wanRulesIndexResource{}
	req := resource.MetadataRequest{ProviderTypeName: "cato"}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	if resp.TypeName != "cato_bulk_wf_move_rule" {
		t.Fatalf("unexpected type name %q", resp.TypeName)
	}
}

func TestWanRulesIndexResourceConfigure(t *testing.T) {
	t.Parallel()

	t.Run("nil provider data", func(t *testing.T) {
		t.Parallel()

		r := &wanRulesIndexResource{}
		req := resource.ConfigureRequest{ProviderData: nil}
		resp := &resource.ConfigureResponse{}

		r.Configure(context.Background(), req, resp)

		if r.client != nil {
			t.Fatalf("expected nil client when provider data is nil")
		}
	})

	t.Run("sets provider client", func(t *testing.T) {
		t.Parallel()

		providerData := &catoClientData{AccountId: "account-123"}
		r := &wanRulesIndexResource{}
		req := resource.ConfigureRequest{ProviderData: providerData}
		resp := &resource.ConfigureResponse{}

		r.Configure(context.Background(), req, resp)

		if r.client != providerData {
			t.Fatalf("expected resource client to be set from provider data")
		}
	})
}

func TestWanRulesIndexResourceGetClient(t *testing.T) {
	t.Parallel()

	t.Run("nil without provider client", func(t *testing.T) {
		t.Parallel()

		r := &wanRulesIndexResource{}
		if got := r.getWanRulesIndexClient(); got != nil {
			t.Fatalf("expected nil client, got %T", got)
		}
	})

	t.Run("uses injected client", func(t *testing.T) {
		t.Parallel()

		mockClient := mocks.NewWanRulesIndexClient(t)
		r := &wanRulesIndexResource{wanRulesIndexClient: mockClient}
		if got := r.getWanRulesIndexClient(); got != mockClient {
			t.Fatalf("expected injected client, got %T", got)
		}
	})

	t.Run("falls back to provider client", func(t *testing.T) {
		t.Parallel()

		sdkClient := &cato_go_sdk.Client{}
		r := &wanRulesIndexResource{client: &catoClientData{catov2: sdkClient}}
		if got := r.getWanRulesIndexClient(); got != sdkClient {
			t.Fatalf("expected provider SDK client, got %T", got)
		}
	})
}

func TestWanRulesIndexCreateReturnsDiagnosticsOnSectionsIndexError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockClient := mocks.NewWanRulesIndexClient(t)
	mockClient.EXPECT().
		PolicyWanFirewallSectionsIndex(ctx, "account-123").
		Return(nil, errors.New("sections index failed")).
		Once()

	r := &wanRulesIndexResource{
		client:              &catoClientData{AccountId: "account-123"},
		wanRulesIndexClient: mockClient,
	}
	req := resource.CreateRequest{Plan: newWanRulesIndexPlan(ctx, t)}
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: getWanRulesIndexSchema(ctx, t)}}

	r.Create(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostics for sections index error")
	}
}

func TestWanRulesIndexUpdateReturnsDiagnosticsOnSectionsIndexError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockClient := mocks.NewWanRulesIndexClient(t)
	mockClient.EXPECT().
		PolicyWanFirewallSectionsIndex(ctx, "account-123").
		Return(nil, errors.New("sections index failed")).
		Once()

	r := &wanRulesIndexResource{
		client:              &catoClientData{AccountId: "account-123"},
		wanRulesIndexClient: mockClient,
	}
	req := resource.UpdateRequest{Plan: newWanRulesIndexPlan(ctx, t)}
	resp := &resource.UpdateResponse{State: tfsdk.State{Schema: getWanRulesIndexSchema(ctx, t)}}

	r.Update(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostics for sections index error")
	}
}

func TestMoveWanRulesAndSectionsReturnsErrorWhenClientNotConfigured(t *testing.T) {
	t.Parallel()

	r := &wanRulesIndexResource{}

	_, _, diags, err := r.moveWanRulesAndSections(context.Background(), WanRulesIndex{})
	if err == nil {
		t.Fatal("expected error when WAN client is not configured")
	}
	if !strings.Contains(err.Error(), "wan rules index client is not configured") {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatal("expected diagnostics for missing WAN client")
	}
}

func TestMoveWanRulesAndSectionsReturnsErrorForUnknownSectionToStartAfterID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockClient := mocks.NewWanRulesIndexClient(t)
	mockClient.EXPECT().
		PolicyWanFirewallSectionsIndex(ctx, "account-123").
		Return(wanSectionsIndexResponse([]wanSection{{id: "section-1", name: "first"}}), nil).
		Once()

	r := &wanRulesIndexResource{
		client:              &catoClientData{AccountId: "account-123"},
		wanRulesIndexClient: mockClient,
	}

	_, _, diags, err := r.moveWanRulesAndSections(ctx, WanRulesIndex{
		SectionToStartAfterID: types.StringValue("missing-id"),
		SectionData:           mustMapValue(t, WanSectionIndexResourceObjectTypes, map[string]attr.Value{}),
		RuleData:              mustMapValue(t, WanRuleIndexResourceObjectTypes, map[string]attr.Value{}),
	})
	if err == nil {
		t.Fatal("expected error for missing section_to_start_after_id")
	}
	if !strings.Contains(err.Error(), "sectionToStartAfterId not found") {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatal("expected diagnostics for missing section_to_start_after_id")
	}
}

func TestMoveWanRulesAndSectionsReturnsReorderAPIErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockClient := mocks.NewWanRulesIndexClient(t)
	sectionName := "managed-section"
	ruleName := "managed-rule"

	mockClient.EXPECT().
		PolicyWanFirewallSectionsIndex(ctx, "account-123").
		Return(wanSectionsIndexResponse([]wanSection{{id: "section-1", name: sectionName}}), nil).
		Once()
	mockClient.EXPECT().
		PolicyWanFirewallMoveSection(ctx, mock.Anything, "account-123").
		Return(&cato_go_sdk.PolicyWanFirewallMoveSection{}, nil).
		Once()
	mockClient.EXPECT().
		PolicyWanFirewallRulesIndex(ctx, "account-123").
		Return(wanRulesIndexResponse([]wanRule{{id: "rule-1", name: ruleName, sectionName: sectionName}}), nil).
		Once()
	mockClient.EXPECT().
		PolicyWanFirewallReorderPolicy(ctx, (*cato_models.WanFirewallPolicyMutationInput)(nil), mock.Anything, "account-123").
		Return(wanReorderResponseWithError("reorder failed from api"), nil).
		Once()

	r := &wanRulesIndexResource{
		client:              &catoClientData{AccountId: "account-123"},
		wanRulesIndexClient: mockClient,
	}

	sectionData := mustMapValue(t, WanSectionIndexResourceObjectTypes, map[string]attr.Value{
		sectionName: mustSectionObject(t, "", sectionName, 1),
	})
	ruleData := mustMapValue(t, WanRuleIndexResourceObjectTypes, map[string]attr.Value{
		ruleName: mustRuleObject(t, "", sectionName, ruleName, 1),
	})
	_, _, diags, err := r.moveWanRulesAndSections(ctx, WanRulesIndex{
		SectionToStartAfterID: types.StringNull(),
		SectionData:           sectionData,
		RuleData:              ruleData,
	})
	if err == nil {
		t.Fatal("expected error from reorder API failure")
	}
	if !strings.Contains(err.Error(), "reorder failed from api") {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatal("expected diagnostics for reorder API failure")
	}
}

func TestKeepPlannedMapKeys_UsesActualForPlannedKeysOnly(t *testing.T) {
	t.Parallel()

	planMap := mustMapValue(t, WanSectionIndexResourceObjectTypes, map[string]attr.Value{
		"managed": mustSectionObject(t, "plan-id", "managed", 1),
	})
	actualMap := mustMapValue(t, WanSectionIndexResourceObjectTypes, map[string]attr.Value{
		"managed":  mustSectionObject(t, "actual-id", "managed", 10),
		"unmanaged": mustSectionObject(t, "extra-id", "unmanaged", 20),
	})

	got, diags := keepPlannedMapKeys(planMap, actualMap, WanSectionIndexResourceObjectTypes)
	if diags.HasError() {
		t.Fatalf("expected no diagnostics, got: %+v", diags)
	}

	gotElements := got.Elements()
	if len(gotElements) != 1 {
		t.Fatalf("expected one managed element, got %d", len(gotElements))
	}

	obj, ok := gotElements["managed"].(types.Object)
	if !ok {
		t.Fatalf("expected managed value to be object, got %T", gotElements["managed"])
	}
	gotSection := asSectionItem(t, context.Background(), obj)
	if gotSection.ID.ValueString() != "actual-id" {
		t.Fatalf("expected managed element from actual state, got id %q", gotSection.ID.ValueString())
	}
	if gotSection.SectionIndex.ValueInt64() != 10 {
		t.Fatalf("expected section_index 10 from actual state, got %d", gotSection.SectionIndex.ValueInt64())
	}
}

func TestKeepPlannedMapKeys_UsesPlanValueWhenActualKeyMissing(t *testing.T) {
	t.Parallel()

	planMap := mustMapValue(t, WanSectionIndexResourceObjectTypes, map[string]attr.Value{
		"managed": mustSectionObject(t, "plan-id", "managed", 1),
	})
	actualMap := mustMapValue(t, WanSectionIndexResourceObjectTypes, map[string]attr.Value{})

	got, diags := keepPlannedMapKeys(planMap, actualMap, WanSectionIndexResourceObjectTypes)
	if diags.HasError() {
		t.Fatalf("expected no diagnostics, got: %+v", diags)
	}

	gotElements := got.Elements()
	if len(gotElements) != 1 {
		t.Fatalf("expected one managed element, got %d", len(gotElements))
	}

	obj, ok := gotElements["managed"].(types.Object)
	if !ok {
		t.Fatalf("expected managed value to be object, got %T", gotElements["managed"])
	}
	gotSection := asSectionItem(t, context.Background(), obj)
	if gotSection.ID.ValueString() != "plan-id" {
		t.Fatalf("expected managed element to fall back to plan value, got id %q", gotSection.ID.ValueString())
	}
}

func TestKeepPlannedMapKeys_ReturnsActualWhenPlanMapNullOrUnknown(t *testing.T) {
	t.Parallel()

	actualMap := mustMapValue(t, WanSectionIndexResourceObjectTypes, map[string]attr.Value{
		"managed": mustSectionObject(t, "actual-id", "managed", 10),
	})

	testCases := []struct {
		name    string
		planMap types.Map
	}{
		{
			name:    "null plan",
			planMap: types.MapNull(WanSectionIndexResourceObjectTypes),
		},
		{
			name:    "unknown plan",
			planMap: types.MapUnknown(WanSectionIndexResourceObjectTypes),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, diags := keepPlannedMapKeys(tc.planMap, actualMap, WanSectionIndexResourceObjectTypes)
			if diags.HasError() {
				t.Fatalf("expected no diagnostics, got: %+v", diags)
			}
			if !got.Equal(actualMap) {
				t.Fatalf("expected actual map to be returned when plan is %s", tc.name)
			}
		})
	}
}

func mustSectionObject(t *testing.T, id, sectionName string, sectionIndex int64) basetypes.ObjectValue {
	t.Helper()

	value, diags := types.ObjectValue(
		WanSectionIndexResourceAttrTypes,
		map[string]attr.Value{
			"id":            types.StringValue(id),
			"section_name":  types.StringValue(sectionName),
			"section_index": types.Int64Value(sectionIndex),
		},
	)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics creating section object: %+v", diags)
	}
	return value
}

func mustMapValue(t *testing.T, objectType types.ObjectType, elements map[string]attr.Value) types.Map {
	t.Helper()

	value, diags := types.MapValue(objectType, elements)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics creating map value: %+v", diags)
	}
	return value
}

func mustRuleObject(t *testing.T, id, sectionName, ruleName string, indexInSection int64) basetypes.ObjectValue {
	t.Helper()

	value, diags := types.ObjectValue(
		WanRuleIndexResourceAttrTypes,
		map[string]attr.Value{
			"id":               types.StringValue(id),
			"index_in_section": types.Int64Value(indexInSection),
			"section_name":     types.StringValue(sectionName),
			"rule_name":        types.StringValue(ruleName),
			"description":      types.StringValue(""),
			"enabled":          types.BoolValue(true),
		},
	)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics creating rule object: %+v", diags)
	}
	return value
}

func asSectionItem(t *testing.T, ctx context.Context, object types.Object) WanRulesSectionItemIndex {
	t.Helper()

	var item WanRulesSectionItemIndex
	diags := object.As(ctx, &item, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics decoding object: %+v", diags)
	}
	return item
}

func getWanRulesIndexSchema(ctx context.Context, t *testing.T) schema.Schema {
	t.Helper()

	r := &wanRulesIndexResource{}
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)
	return schemaResp.Schema
}

func newWanRulesIndexPlan(ctx context.Context, t *testing.T) tfsdk.Plan {
	t.Helper()

	emptyRuleMap := mustMapValue(t, WanRuleIndexResourceObjectTypes, map[string]attr.Value{})
	emptySectionMap := mustMapValue(t, WanSectionIndexResourceObjectTypes, map[string]attr.Value{})

	plan := tfsdk.Plan{Schema: getWanRulesIndexSchema(ctx, t)}
	diags := plan.Set(ctx, WanRulesIndex{
		SectionToStartAfterID: types.StringNull(),
		RuleData:              emptyRuleMap,
		SectionData:           emptySectionMap,
	})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics creating plan: %+v", diags)
	}

	return plan
}

type wanSection struct {
	id   string
	name string
}

func wanSectionsIndexResponse(sections []wanSection) *cato_go_sdk.WanSectionsIndexPolicy {
	respSections := make([]*cato_go_sdk.Policy_Policy_WanFirewall_Policy_Sections, 0, len(sections))
	for _, section := range sections {
		respSections = append(respSections, &cato_go_sdk.Policy_Policy_WanFirewall_Policy_Sections{
			Section: cato_go_sdk.Policy_Policy_WanFirewall_Policy_Sections_Section{
				ID:   section.id,
				Name: section.name,
			},
		})
	}

	return &cato_go_sdk.WanSectionsIndexPolicy{
		Policy: &cato_go_sdk.WanSectionsIndexPolicy_Policy{
			WanFirewall: &cato_go_sdk.Policy_Policy_WanFirewall{
				Policy: cato_go_sdk.Policy_Policy_WanFirewall_Policy{
					Sections: respSections,
				},
			},
		},
	}
}

type wanRule struct {
	id          string
	name        string
	sectionName string
}

func wanRulesIndexResponse(rules []wanRule) *cato_go_sdk.WanRulesIndexPolicy {
	respRules := make([]*cato_go_sdk.WanRulesIndexPolicy_Policy_WanFirewall_Policy_Rules, 0, len(rules))
	for _, rule := range rules {
		respRules = append(respRules, &cato_go_sdk.WanRulesIndexPolicy_Policy_WanFirewall_Policy_Rules{
			Rule: cato_go_sdk.WanRulesIndexPolicy_Policy_WanFirewall_Policy_Rules_Rule{
				ID:   rule.id,
				Name: rule.name,
				Section: cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules_Rule_Section{
					Name: rule.sectionName,
				},
			},
		})
	}

	return &cato_go_sdk.WanRulesIndexPolicy{
		Policy: &cato_go_sdk.WanRulesIndexPolicy_Policy{
			WanFirewall: &cato_go_sdk.WanRulesIndexPolicy_Policy_WanFirewall{
				Policy: cato_go_sdk.WanRulesIndexPolicy_Policy_WanFirewall_Policy{
					Rules: respRules,
				},
			},
		},
	}
}

func wanReorderResponseWithError(message string) *cato_go_sdk.PolicyWanFirewallReorderPolicy {
	status := cato_models.PolicyMutationStatus("FAILED")

	return &cato_go_sdk.PolicyWanFirewallReorderPolicy{
		Policy: &cato_go_sdk.PolicyWanFirewallReorderPolicy_Policy{
			WanFirewall: &cato_go_sdk.PolicyWanFirewallReorderPolicy_Policy_WanFirewall{
				ReorderPolicy: cato_go_sdk.PolicyWanFirewallReorderPolicy_Policy_WanFirewall_ReorderPolicy{
					Status: status,
					Errors: []*cato_go_sdk.PolicyWanFirewallReorderPolicy_Policy_WanFirewall_ReorderPolicy_Errors{
						{
							ErrorMessage: &message,
						},
					},
				},
			},
		},
	}
}
