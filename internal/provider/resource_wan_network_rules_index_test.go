package provider

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestWanNetworkRulesIndexResourceCreateReturnsDiagnosticWhenPolicyReadFails(t *testing.T) {
	ctx := context.Background()
	r := &wanNetworkRulesIndexResource{
		client: &catoClientData{AccountId: "account-123"},
		wanNetworkClient: &wanNetworkBulkPolicyClientStub{
			wanNetworkPolicyErr: errors.New("wan network read failed"),
		},
	}

	plan := tfsdk.Plan{Schema: getWanNetworkRulesIndexSchema(ctx, t)}
	diags := plan.Set(ctx, WanNetworkRulesIndex{
		SectionToStartAfterId: types.StringValue("section-123"),
		RuleData:              types.MapNull(WanNetworkRuleIndexResourceObjectTypes),
		SectionData:           types.MapNull(WanNetworkSectionIndexResourceObjectTypes),
	})
	if diags.HasError() {
		t.Fatalf("unexpected plan diagnostics: %+v", diags)
	}

	req := resource.CreateRequest{Plan: plan}
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: getWanNetworkRulesIndexSchema(ctx, t)}}

	r.Create(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostics, got none")
	}
	assertDiagnosticsContainSubstring(t, resp.Diagnostics.Errors(), "wan network read failed")
}

func getWanNetworkRulesIndexSchema(ctx context.Context, t *testing.T) schema.Schema {
	t.Helper()

	r := &wanNetworkRulesIndexResource{}
	resp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, resp)

	return resp.Schema
}

type wanNetworkBulkPolicyClientStub struct {
	wanNetworkPolicyErr error
}

func (s *wanNetworkBulkPolicyClientStub) WanNetworkPolicy(context.Context, string, ...clientv2.RequestInterceptor) (*cato_go_sdk.WanNetworkPolicy, error) {
	if s.wanNetworkPolicyErr != nil {
		return nil, s.wanNetworkPolicyErr
	}
	return &cato_go_sdk.WanNetworkPolicy{}, nil
}

func (*wanNetworkBulkPolicyClientStub) PolicyWanNetworkMoveSection(context.Context, cato_models.PolicyMoveSectionInput, string, ...clientv2.RequestInterceptor) (*cato_go_sdk.PolicyWanNetworkMoveSection, error) {
	panic("unexpected call to PolicyWanNetworkMoveSection")
}

func (*wanNetworkBulkPolicyClientStub) PolicyWanNetworkMoveRule(context.Context, cato_models.PolicyMoveRuleInput, string, ...clientv2.RequestInterceptor) (*cato_go_sdk.PolicyWanNetworkMoveRule, error) {
	panic("unexpected call to PolicyWanNetworkMoveRule")
}

func (*wanNetworkBulkPolicyClientStub) PolicyWanNetworkPublishPolicyRevision(context.Context, string, ...clientv2.RequestInterceptor) (*cato_go_sdk.PolicyWanNetworkPublishPolicyRevision, error) {
	panic("unexpected call to PolicyWanNetworkPublishPolicyRevision")
}

func assertDiagnosticsContainSubstring(t *testing.T, errs []diag.Diagnostic, want string) {
	t.Helper()

	for _, err := range errs {
		if strings.Contains(err.Detail(), want) || strings.Contains(err.Summary(), want) {
			return
		}
	}

	t.Fatalf("expected diagnostics to contain %q, got %+v", want, errs)
}
