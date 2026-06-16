package provider

import (
	"context"
	"errors"
	"strings"
	"testing"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/mock"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
)

func TestNewIfwRulesIndexResource(t *testing.T) {
	t.Parallel()

	res := NewIfwRulesIndexResource()
	if _, ok := res.(*ifwRulesIndexResource); !ok {
		t.Fatalf("expected *ifwRulesIndexResource, got %T", res)
	}
}

func TestIfwRulesIndexResourceMetadata(t *testing.T) {
	t.Parallel()

	r := &ifwRulesIndexResource{}
	req := resource.MetadataRequest{ProviderTypeName: "cato"}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	if resp.TypeName != "cato_bulk_if_move_rule" {
		t.Fatalf("unexpected type name %q", resp.TypeName)
	}
}

func TestIfwRulesIndexResourceConfigure(t *testing.T) {
	t.Parallel()

	t.Run("nil provider data", func(t *testing.T) {
		t.Parallel()

		r := &ifwRulesIndexResource{}
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
		r := &ifwRulesIndexResource{}
		req := resource.ConfigureRequest{ProviderData: providerData}
		resp := &resource.ConfigureResponse{}

		r.Configure(context.Background(), req, resp)

		if r.client != providerData {
			t.Fatalf("expected resource client to be set from provider data")
		}
	})
}

func TestIfwRulesIndexResourceGetClient(t *testing.T) {
	t.Parallel()

	t.Run("nil without provider client", func(t *testing.T) {
		t.Parallel()

		r := &ifwRulesIndexResource{}
		if got := r.getIfwRulesIndexClient(); got != nil {
			t.Fatalf("expected nil client, got %T", got)
		}
	})

	t.Run("uses injected client", func(t *testing.T) {
		t.Parallel()

		mockClient := mocks.NewIfwRulesIndexClient(t)
		r := &ifwRulesIndexResource{ifwRulesIndexClient: mockClient}
		if got := r.getIfwRulesIndexClient(); got != mockClient {
			t.Fatalf("expected injected client, got %T", got)
		}
	})

	t.Run("falls back to provider client", func(t *testing.T) {
		t.Parallel()

		sdkClient := &cato_go_sdk.Client{}
		r := &ifwRulesIndexResource{client: &catoClientData{catov2: sdkClient}}
		if got := r.getIfwRulesIndexClient(); got != sdkClient {
			t.Fatalf("expected provider SDK client, got %T", got)
		}
	})
}

func TestEnsureIfwDraftMutationInputReusesExistingDraft(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockClient := mocks.NewIfwRulesIndexClient(t)
	mockClient.EXPECT().
		PolicyInternetFirewall(ctx, mock.Anything, "account-123").
		Return(ifwPolicyWithDraftRevision("draft-rev-1"), nil).
		Once()

	got, err := ensureIfwDraftMutationInput(ctx, mockClient, "account-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.Revision == nil || got.Revision.ID == nil {
		t.Fatal("expected revision id in mutation input")
	}
	if *got.Revision.ID != "draft-rev-1" {
		t.Fatalf("expected draft-rev-1, got %q", *got.Revision.ID)
	}
}

func TestEnsureIfwDraftMutationInputCreatesDraftWhenMissing(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockClient := mocks.NewIfwRulesIndexClient(t)
	mockClient.EXPECT().
		PolicyInternetFirewall(ctx, mock.Anything, "account-123").
		Return(ifwPolicyWithDraftRevision(""), nil).
		Once()
	mockClient.EXPECT().
		PolicyInternetFirewallCreatePolicyRevision(ctx, mock.Anything, mock.Anything, "account-123").
		Return(ifwCreateRevisionResponse("new-draft-rev"), nil).
		Once()

	got, err := ensureIfwDraftMutationInput(ctx, mockClient, "account-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.Revision == nil || got.Revision.ID == nil {
		t.Fatal("expected revision id in mutation input")
	}
	if *got.Revision.ID != "new-draft-rev" {
		t.Fatalf("expected new-draft-rev, got %q", *got.Revision.ID)
	}
}

func TestIfwRulesIndexCreateReturnsDiagnosticsOnDraftRevisionError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockClient := mocks.NewIfwRulesIndexClient(t)
	mockClient.EXPECT().
		PolicyInternetFirewall(ctx, mock.Anything, "account-123").
		Return(nil, errors.New("policy query failed")).
		Once()

	r := &ifwRulesIndexResource{
		client:              &catoClientData{AccountId: "account-123"},
		ifwRulesIndexClient: mockClient,
	}
	req := resource.CreateRequest{Plan: newIfwRulesIndexPlan(ctx, t)}
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: getIfwRulesIndexSchema(ctx, t)}}

	r.Create(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostics for draft revision error")
	}
}

func TestMoveIfwRulesAndSectionsReturnsErrorWhenClientNotConfigured(t *testing.T) {
	t.Parallel()

	r := &ifwRulesIndexResource{}

	_, _, diags, err := r.moveIfwRulesAndSections(context.Background(), IfwRulesIndex{})
	if err == nil {
		t.Fatal("expected error when IFW client is not configured")
	}
	if !strings.Contains(err.Error(), "ifw rules index client is not configured") {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatal("expected diagnostics for missing IFW client")
	}
}

func TestMoveIfwRulesAndSectionsReturnsErrorForUnknownSectionToStartAfterID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockClient := mocks.NewIfwRulesIndexClient(t)
	mockClient.EXPECT().
		PolicyInternetFirewall(ctx, mock.Anything, "account-123").
		Return(ifwPolicyWithDraftRevision("draft-rev-1"), nil).
		Once()
	mockClient.EXPECT().
		PolicyInternetFirewallSectionsIndex(ctx, "account-123").
		Return(ifwSectionsIndexResponse([]ifwSection{{id: "section-1", name: "first"}}), nil).
		Once()

	r := &ifwRulesIndexResource{
		client:              &catoClientData{AccountId: "account-123"},
		ifwRulesIndexClient: mockClient,
	}

	_, _, diags, err := r.moveIfwRulesAndSections(ctx, IfwRulesIndex{
		SectionToStartAfterID: types.StringValue("missing-id"),
		SectionData:           mustIfwMapValue(t, IfwSectionIndexResourceObjectTypes, map[string]attr.Value{}),
		RuleData:              mustIfwMapValue(t, IfwRuleIndexResourceObjectTypes, map[string]attr.Value{}),
	})
	if err == nil {
		t.Fatal("expected error for missing section_to_start_after_id")
	}
	if !strings.Contains(err.Error(), "SectionToStartAfterID not found") {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatal("expected diagnostics for missing section_to_start_after_id")
	}
}

func TestMoveIfwRulesAndSectionsReturnsReorderAPIErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockClient := mocks.NewIfwRulesIndexClient(t)
	sectionName := "managed-section"
	ruleName := "managed-rule"
	revisionID := "draft-rev-1"

	mockClient.EXPECT().
		PolicyInternetFirewall(ctx, mock.Anything, "account-123").
		Return(ifwPolicyWithDraftRevision(revisionID), nil).
		Once()
	mockClient.EXPECT().
		PolicyInternetFirewallSectionsIndex(ctx, "account-123").
		Return(ifwSectionsIndexResponse([]ifwSection{{id: "section-1", name: sectionName}}), nil).
		Once()
	mockClient.EXPECT().
		PolicyInternetFirewallMoveSection(ctx, mock.MatchedBy(func(input *cato_models.InternetFirewallPolicyMutationInput) bool {
			return input != nil && input.Revision != nil && input.Revision.ID != nil && *input.Revision.ID == revisionID
		}), mock.Anything, "account-123").
		Return(&cato_go_sdk.PolicyInternetFirewallMoveSection{}, nil).
		Once()
	mockClient.EXPECT().
		PolicyInternetFirewall(ctx, mock.Anything, "account-123").
		Return(ifwFullPolicyResponse(sectionName, ruleName), nil).
		Once()
	mockClient.EXPECT().
		PolicyInternetFirewallReorderPolicy(ctx, mock.MatchedBy(func(input *cato_models.InternetFirewallPolicyMutationInput) bool {
			return input != nil && input.Revision != nil && input.Revision.ID != nil && *input.Revision.ID == revisionID
		}), mock.Anything, "account-123").
		Return(ifwReorderResponseWithError("reorder failed from api"), nil).
		Once()

	r := &ifwRulesIndexResource{
		client:              &catoClientData{AccountId: "account-123"},
		ifwRulesIndexClient: mockClient,
	}

	sectionData := mustIfwMapValue(t, IfwSectionIndexResourceObjectTypes, map[string]attr.Value{
		sectionName: mustIfwSectionObject(t, "", sectionName, 1),
	})
	ruleData := mustIfwMapValue(t, IfwRuleIndexResourceObjectTypes, map[string]attr.Value{
		ruleName: mustIfwRuleObject(t, "", sectionName, ruleName, 1),
	})
	_, _, diags, err := r.moveIfwRulesAndSections(ctx, IfwRulesIndex{
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

func mustIfwSectionObject(t *testing.T, id, sectionName string, sectionIndex int64) basetypes.ObjectValue {
	t.Helper()

	value, diags := types.ObjectValue(
		IfwSectionIndexResourceAttrTypes,
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

func mustIfwMapValue(t *testing.T, objectType types.ObjectType, elements map[string]attr.Value) types.Map {
	t.Helper()

	value, diags := types.MapValue(objectType, elements)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics creating map value: %+v", diags)
	}
	return value
}

func mustIfwRuleObject(t *testing.T, id, sectionName, ruleName string, indexInSection int64) basetypes.ObjectValue {
	t.Helper()

	value, diags := types.ObjectValue(
		IfwRuleIndexResourceAttrTypes,
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

func getIfwRulesIndexSchema(ctx context.Context, t *testing.T) schema.Schema {
	t.Helper()

	r := &ifwRulesIndexResource{}
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)
	return schemaResp.Schema
}

func newIfwRulesIndexPlan(ctx context.Context, t *testing.T) tfsdk.Plan {
	t.Helper()

	emptyRuleMap := mustIfwMapValue(t, IfwRuleIndexResourceObjectTypes, map[string]attr.Value{})
	emptySectionMap := mustIfwMapValue(t, IfwSectionIndexResourceObjectTypes, map[string]attr.Value{})

	plan := tfsdk.Plan{Schema: getIfwRulesIndexSchema(ctx, t)}
	diags := plan.Set(ctx, IfwRulesIndex{
		SectionToStartAfterID: types.StringNull(),
		RuleData:              emptyRuleMap,
		SectionData:           emptySectionMap,
	})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics creating plan: %+v", diags)
	}

	return plan
}

type ifwSection struct {
	id   string
	name string
}

func ifwPolicyWithDraftRevision(revisionID string) *cato_go_sdk.Policy {
	policy := &cato_go_sdk.Policy{
		Policy: &cato_go_sdk.Policy_Policy{
			InternetFirewall: &cato_go_sdk.Policy_Policy_InternetFirewall{},
		},
	}
	if revisionID != "" {
		policy.Policy.InternetFirewall.RevisionsInternetFirewallPolicyQueries = &cato_go_sdk.Policy_Policy_InternetFirewall_RevisionsInternetFirewallPolicyQueries{
			Revision: []*cato_go_sdk.Policy_Policy_InternetFirewall_RevisionsInternetFirewallPolicyQueries_Revision{
				{ID: revisionID},
			},
		}
	}
	return policy
}

func ifwCreateRevisionResponse(revisionID string) *cato_go_sdk.PolicyInternetFirewallCreatePolicyRevision {
	status := cato_models.PolicyMutationStatusSuccess
	return &cato_go_sdk.PolicyInternetFirewallCreatePolicyRevision{
		Policy: &cato_go_sdk.PolicyInternetFirewallCreatePolicyRevision_Policy{
			InternetFirewall: &cato_go_sdk.PolicyInternetFirewallCreatePolicyRevision_Policy_InternetFirewall{
				CreatePolicyRevision: cato_go_sdk.PolicyInternetFirewallCreatePolicyRevision_Policy_InternetFirewall_CreatePolicyRevision{
					Status: status,
					Policy: &cato_go_sdk.PolicyInternetFirewallCreatePolicyRevision_Policy_InternetFirewall_CreatePolicyRevision_Policy{
						RevisionInternetFirewallPolicy: &cato_go_sdk.PolicyInternetFirewallCreatePolicyRevision_Policy_InternetFirewall_CreatePolicyRevision_Policy_RevisionInternetFirewallPolicy{
							ID: revisionID,
						},
					},
				},
			},
		},
	}
}

func ifwSectionsIndexResponse(sections []ifwSection) *cato_go_sdk.IfwSectionsIndexPolicy {
	respSections := make([]*cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Sections, 0, len(sections))
	for _, section := range sections {
		respSections = append(respSections, &cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Sections{
			Section: cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Sections_Section{
				ID:   section.id,
				Name: section.name,
			},
		})
	}

	return &cato_go_sdk.IfwSectionsIndexPolicy{
		Policy: &cato_go_sdk.IfwSectionsIndexPolicy_Policy{
			InternetFirewall: &cato_go_sdk.IfwSectionsIndexPolicy_Policy_Policy{
				Policy: cato_go_sdk.Policy_Policy_IfwSectionsIndexPolicy_Policy_Policy{
					Sections: respSections,
				},
			},
		},
	}
}

func ifwFullPolicyResponse(sectionName, ruleName string) *cato_go_sdk.Policy {
	return &cato_go_sdk.Policy{
		Policy: &cato_go_sdk.Policy_Policy{
			InternetFirewall: &cato_go_sdk.Policy_Policy_InternetFirewall{
				Policy: cato_go_sdk.Policy_Policy_InternetFirewall_Policy{
					Rules: []*cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules{
						{
							Rule: cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_Rule{
								ID:          "rule-1",
								Name:        ruleName,
								Description: "rule description",
								Enabled:     true,
								Section: cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_Rule_Section{
									ID:   "section-1",
									Name: sectionName,
								},
							},
						},
					},
				},
			},
		},
	}
}

func ifwReorderResponseWithError(message string) *cato_go_sdk.PolicyInternetFirewallReorderPolicy {
	status := cato_models.PolicyMutationStatus("FAILED")

	return &cato_go_sdk.PolicyInternetFirewallReorderPolicy{
		Policy: &cato_go_sdk.PolicyInternetFirewallReorderPolicy_Policy{
			InternetFirewall: &cato_go_sdk.PolicyInternetFirewallReorderPolicy_Policy_InternetFirewall{
				ReorderPolicy: cato_go_sdk.PolicyInternetFirewallReorderPolicy_Policy_InternetFirewall_ReorderPolicy{
					Status: status,
					Errors: []*cato_go_sdk.PolicyInternetFirewallReorderPolicy_Policy_InternetFirewall_ReorderPolicy_Errors{
						{
							ErrorMessage: &message,
						},
					},
				},
			},
		},
	}
}
