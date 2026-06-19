package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestNewApplicationControlRuleResource(t *testing.T) {
	t.Parallel()
	r := NewApplicationControlRuleResource()
	if r == nil {
		t.Fatal("expected resource instance, got nil")
	}
	if _, ok := r.(*applicationControlRuleResource); !ok {
		t.Fatalf("expected *applicationControlRuleResource, got %T", r)
	}
}

func TestApplicationControlRuleMetadata(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := &applicationControlRuleResource{}
	resp := &resource.MetadataResponse{}
	r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "cato"}, resp)
	if resp.TypeName != "cato_application_control_rule" {
		t.Fatalf("expected type name cato_application_control_rule, got %q", resp.TypeName)
	}
}

func TestApplicationControlRuleConfigureNilProviderData(t *testing.T) {
	t.Parallel()
	r := &applicationControlRuleResource{}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{}, resp)
	if r.client != nil {
		t.Fatal("expected client to remain nil when provider data is nil")
	}
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}
}

func TestApplicationControlRuleConfigureSetsClient(t *testing.T) {
	t.Parallel()
	client := &catoClientData{AccountId: "123"}
	r := &applicationControlRuleResource{}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{ProviderData: client}, resp)
	if r.client != client {
		t.Fatal("expected resource client to be set from provider data")
	}
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}
}

func TestApplicationControlRuleImportState(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := &applicationControlRuleResource{}
	resp := &resource.ImportStateResponse{
		State: tfsdk.State{Schema: getApplicationControlRuleSchema(ctx, t)},
	}
	diags := resp.State.Set(ctx, ApplicationControlRule{
		Rule: types.ObjectNull(ApplicationControlRuleRuleAttrTypes),
		At:   types.ObjectNull(PositionAttrTypes),
	})
	if diags.HasError() {
		t.Fatalf("unexpected seed state diagnostics: %+v", diags)
	}
	r.ImportState(ctx, resource.ImportStateRequest{ID: "rule-123"}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}
	var imported ApplicationControlRule
	diags = resp.State.Get(ctx, &imported)
	if diags.HasError() {
		t.Fatalf("unexpected state diagnostics: %+v", diags)
	}
	ruleModel := ApplicationControlRuleRulePlan{}
	diags = imported.Rule.As(ctx, &ruleModel, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		t.Fatalf("unexpected rule diagnostics: %+v", diags)
	}
	if ruleModel.ID.ValueString() != "rule-123" {
		t.Fatalf("expected imported rule id rule-123, got %q", ruleModel.ID.ValueString())
	}
}

func getApplicationControlRuleSchema(ctx context.Context, t *testing.T) schema.Schema {
	t.Helper()
	r := &applicationControlRuleResource{}
	sr := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, sr)
	return sr.Schema
}
