package provider

import (
	"context"
	"errors"
	"testing"

	"github.com/Yamashou/gqlgenc/clientv2"
	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestTlsRulesIndexResourceCreateReturnsDiagnosticWhenPolicyReadFails(t *testing.T) {
	ctx := context.Background()
	r := &tlsRulesIndexResource{
		client: &catoClientData{AccountId: "account-123"},
		tlsClient: &tlsBulkPolicyClientStub{
			tlsPolicyErr: errors.New("tls policy read failed"),
		},
	}

	plan := tfsdk.Plan{Schema: getTlsRulesIndexSchema(ctx, t)}
	diags := plan.Set(ctx, TlsRulesIndex{
		SectionToStartAfterId: types.StringValue("section-123"),
		RuleData:              types.MapNull(TlsRuleIndexResourceObjectTypes),
		SectionData:           types.MapNull(TlsSectionIndexResourceObjectTypes),
	})
	if diags.HasError() {
		t.Fatalf("unexpected plan diagnostics: %+v", diags)
	}

	req := resource.CreateRequest{Plan: plan}
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: getTlsRulesIndexSchema(ctx, t)}}

	r.Create(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostics, got none")
	}
	assertDiagnosticsContainSubstring(t, resp.Diagnostics.Errors(), "tls policy read failed")
}

func getTlsRulesIndexSchema(ctx context.Context, t *testing.T) schema.Schema {
	t.Helper()

	r := &tlsRulesIndexResource{}
	resp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, resp)

	return resp.Schema
}

type tlsBulkPolicyClientStub struct {
	tlsPolicyErr error
}

func (s *tlsBulkPolicyClientStub) Tlsinspectpolicy(context.Context, string, ...clientv2.RequestInterceptor) (*cato_go_sdk.Tlsinspectpolicy, error) {
	if s.tlsPolicyErr != nil {
		return nil, s.tlsPolicyErr
	}
	return &cato_go_sdk.Tlsinspectpolicy{}, nil
}

func (*tlsBulkPolicyClientStub) PolicyTLSInspectMoveSection(context.Context, cato_models.PolicyMoveSectionInput, string, ...clientv2.RequestInterceptor) (*cato_go_sdk.PolicyTLSInspectMoveSection, error) {
	panic("unexpected call to PolicyTLSInspectMoveSection")
}

func (*tlsBulkPolicyClientStub) PolicyTLSInspectMoveRule(context.Context, cato_models.PolicyMoveRuleInput, string, ...clientv2.RequestInterceptor) (*cato_go_sdk.PolicyTLSInspectMoveRule, error) {
	panic("unexpected call to PolicyTLSInspectMoveRule")
}

func (*tlsBulkPolicyClientStub) PolicyTLSInspectPublishPolicyRevision(context.Context, string, ...clientv2.RequestInterceptor) (*cato_go_sdk.PolicyTLSInspectPublishPolicyRevision, error) {
	panic("unexpected call to PolicyTLSInspectPublishPolicyRevision")
}
