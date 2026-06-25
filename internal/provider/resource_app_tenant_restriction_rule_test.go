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

func TestNewAppTenantRestrictionRuleResource(t *testing.T) {
	t.Parallel()
	r := NewAppTenantRestrictionRuleResource()
	if r == nil {
		t.Fatal("expected resource instance, got nil")
	}
	if _, ok := r.(*appTenantRestrictionRuleResource); !ok {
		t.Fatalf("expected *appTenantRestrictionRuleResource, got %T", r)
	}
}

func TestAppTenantRestrictionRuleMetadata(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := &appTenantRestrictionRuleResource{}
	resp := &resource.MetadataResponse{}
	r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "cato"}, resp)
	if resp.TypeName != "cato_app_tenant_restriction_rule" {
		t.Fatalf("expected type name cato_app_tenant_restriction_rule, got %q", resp.TypeName)
	}
}

func TestAppTenantRestrictionRuleConfigureNilProviderData(t *testing.T) {
	t.Parallel()
	r := &appTenantRestrictionRuleResource{}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{}, resp)
	if r.client != nil {
		t.Fatal("expected client to remain nil when provider data is nil")
	}
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}
}

func TestAppTenantRestrictionRuleConfigureSetsClient(t *testing.T) {
	t.Parallel()
	client := &catoClientData{AccountId: "123"}
	r := &appTenantRestrictionRuleResource{}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{ProviderData: client}, resp)
	if r.client != client {
		t.Fatal("expected resource client to be set from provider data")
	}
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}
}

func TestAppTenantRestrictionRuleImportState(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := &appTenantRestrictionRuleResource{}
	resp := &resource.ImportStateResponse{
		State: tfsdk.State{Schema: getAppTenantRestrictionRuleSchema(ctx, t)},
	}
	diags := resp.State.Set(ctx, AppTenantRestrictionRule{
		Rule: types.ObjectNull(AppTenantRestrictionRuleRuleAttrTypes),
		At:   types.ObjectNull(PositionAttrTypes),
	})
	if diags.HasError() {
		t.Fatalf("unexpected seed state diagnostics: %+v", diags)
	}
	r.ImportState(ctx, resource.ImportStateRequest{ID: "rule-456"}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}
	var imported AppTenantRestrictionRule
	diags = resp.State.Get(ctx, &imported)
	if diags.HasError() {
		t.Fatalf("unexpected state diagnostics: %+v", diags)
	}
	ruleModel := AppTenantRestrictionRuleRulePlan{}
	diags = imported.Rule.As(ctx, &ruleModel, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		t.Fatalf("unexpected rule diagnostics: %+v", diags)
	}
	if ruleModel.ID.ValueString() != "rule-456" {
		t.Fatalf("expected imported rule id rule-456, got %q", ruleModel.ID.ValueString())
	}
}

func getAppTenantRestrictionRuleSchema(ctx context.Context, t *testing.T) schema.Schema {
	t.Helper()
	r := &appTenantRestrictionRuleResource{}
	sr := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, sr)
	return sr.Schema
}
