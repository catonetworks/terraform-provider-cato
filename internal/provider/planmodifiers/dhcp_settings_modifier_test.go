package planmodifiers

import (
	"context"
	"testing"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	tf "github.com/catonetworks/terraform-provider-cato/internal/provider/tfmodel"
)

func makeDhcpSettingsObj(t *testing.T, s tf.DhcpSettings) types.Object {
	t.Helper()
	obj, diags := types.ObjectValueFrom(context.Background(), tf.DhcpSettingsAttrTypes, s)
	if diags.HasError() {
		t.Fatalf("failed to create DhcpSettings object: %v", diags)
	}
	return obj
}

// TestPlanDhcpRelay_StateValuePropagation verifies that when the Terraform framework propagates
// a prior state relay_group_id into req.ConfigValue (because the attribute is Optional+Computed),
// the plan modifier does not raise a false "both fields set" error when the user only configured
// relay_group_name.
func TestPlanDhcpRelay_StateValuePropagation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	m := dhcpSettingsModifier{}

	state := &tf.DhcpSettings{
		DhcpType:              types.StringValue(string(cato_models.DhcpTypeDhcpRelay)),
		RelayGroupName:        types.StringValue("CHCVTPJ-DHCP"),
		RelayGroupID:          types.StringValue("4456"),
		IPRange:               types.StringNull(),
		DhcpMicrosegmentation: types.BoolNull(),
	}

	// Simulate: framework propagated relay_group_id="4456" from state into cfg,
	// even though the user only wrote relay_group_name in their config.
	cfg := &tf.DhcpSettings{
		DhcpType:              types.StringValue(string(cato_models.DhcpTypeDhcpRelay)),
		RelayGroupName:        types.StringValue("CHCVTPJ-DHCP"),
		RelayGroupID:          types.StringValue("4456"), // propagated from state, not set by user
		IPRange:               types.StringNull(),
		DhcpMicrosegmentation: types.BoolNull(),
	}

	var d diag.Diagnostics
	result := m.planDhcpRelay(ctx, state, cfg, &d)

	if d.HasError() {
		t.Fatalf("expected no errors, got: %v", d)
	}
	if result.IsNull() || result.IsUnknown() {
		t.Fatal("expected non-null, non-unknown plan result")
	}

	var planSettings tf.DhcpSettings
	if dd := result.As(ctx, &planSettings, basetypes.ObjectAsOptions{}); dd.HasError() {
		t.Fatalf("failed to decode plan result: %v", dd)
	}

	if planSettings.RelayGroupName.ValueString() != "CHCVTPJ-DHCP" {
		t.Errorf("expected relay_group_name=%q, got %q", "CHCVTPJ-DHCP", planSettings.RelayGroupName.ValueString())
	}
	if planSettings.RelayGroupID.ValueString() != "4456" {
		t.Errorf("expected relay_group_id=%q (preserved from state), got %q", "4456", planSettings.RelayGroupID.ValueString())
	}
}

// TestPlanDhcpRelay_ChangedNameStatePropagatedID verifies that when the user changes
// relay_group_name to a new group but the old relay_group_id is still propagated from state,
// the plan uses the new name and marks relay_group_id as unknown (to be resolved at apply time).
func TestPlanDhcpRelay_ChangedNameStatePropagatedID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	m := dhcpSettingsModifier{}

	state := &tf.DhcpSettings{
		DhcpType:              types.StringValue(string(cato_models.DhcpTypeDhcpRelay)),
		RelayGroupName:        types.StringValue("CHCVTPJ-DHCP"),
		RelayGroupID:          types.StringValue("4456"),
		IPRange:               types.StringNull(),
		DhcpMicrosegmentation: types.BoolNull(),
	}

	// User changed the name to "NEW-DHCP"; framework still carries old relay_group_id from state.
	cfg := &tf.DhcpSettings{
		DhcpType:              types.StringValue(string(cato_models.DhcpTypeDhcpRelay)),
		RelayGroupName:        types.StringValue("NEW-DHCP"),
		RelayGroupID:          types.StringValue("4456"), // old id, propagated from state
		IPRange:               types.StringNull(),
		DhcpMicrosegmentation: types.BoolNull(),
	}

	var d diag.Diagnostics
	result := m.planDhcpRelay(ctx, state, cfg, &d)

	if d.HasError() {
		t.Fatalf("expected no errors, got: %v", d)
	}

	var planSettings tf.DhcpSettings
	if dd := result.As(ctx, &planSettings, basetypes.ObjectAsOptions{}); dd.HasError() {
		t.Fatalf("failed to decode plan result: %v", dd)
	}

	if planSettings.RelayGroupName.ValueString() != "NEW-DHCP" {
		t.Errorf("expected relay_group_name=%q, got %q", "NEW-DHCP", planSettings.RelayGroupName.ValueString())
	}
	if !planSettings.RelayGroupID.IsUnknown() {
		t.Errorf("expected relay_group_id to be unknown (new name → id not yet resolved), got %q",
			planSettings.RelayGroupID.ValueString())
	}
}

// TestPlanDhcpRelay_BothExplicitlyChangedErrors verifies that when both relay_group_name and
// relay_group_id differ from state (user appears to have changed both), an error is produced.
func TestPlanDhcpRelay_BothExplicitlyChangedErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	m := dhcpSettingsModifier{}

	state := &tf.DhcpSettings{
		DhcpType:              types.StringValue(string(cato_models.DhcpTypeDhcpRelay)),
		RelayGroupName:        types.StringValue("OLD-DHCP"),
		RelayGroupID:          types.StringValue("1111"),
		IPRange:               types.StringNull(),
		DhcpMicrosegmentation: types.BoolNull(),
	}

	// Both differ from state → genuine conflict
	cfg := &tf.DhcpSettings{
		DhcpType:              types.StringValue(string(cato_models.DhcpTypeDhcpRelay)),
		RelayGroupName:        types.StringValue("NEW-DHCP"),
		RelayGroupID:          types.StringValue("9999"),
		IPRange:               types.StringNull(),
		DhcpMicrosegmentation: types.BoolNull(),
	}

	var d diag.Diagnostics
	m.planDhcpRelay(ctx, state, cfg, &d)

	if !d.HasError() {
		t.Fatal("expected error when both relay fields are explicitly changed, but got none")
	}
}

// TestPlanDhcpRelay_NeitherSetErrors verifies that omitting both relay fields produces an error.
func TestPlanDhcpRelay_NeitherSetErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	m := dhcpSettingsModifier{}

	cfg := &tf.DhcpSettings{
		DhcpType:              types.StringValue(string(cato_models.DhcpTypeDhcpRelay)),
		RelayGroupName:        types.StringNull(),
		RelayGroupID:          types.StringNull(),
		IPRange:               types.StringNull(),
		DhcpMicrosegmentation: types.BoolNull(),
	}

	var d diag.Diagnostics
	m.planDhcpRelay(ctx, nil, cfg, &d)

	if !d.HasError() {
		t.Fatal("expected error when neither relay field is set, but got none")
	}
}

// TestPlanDhcpRelay_FirstCreateByName verifies correct behavior on first create (no prior state)
// when the user configures only relay_group_name.
func TestPlanDhcpRelay_FirstCreateByName(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	m := dhcpSettingsModifier{}

	cfg := &tf.DhcpSettings{
		DhcpType:              types.StringValue(string(cato_models.DhcpTypeDhcpRelay)),
		RelayGroupName:        types.StringValue("MY-DHCP"),
		RelayGroupID:          types.StringNull(),
		IPRange:               types.StringNull(),
		DhcpMicrosegmentation: types.BoolNull(),
	}

	var d diag.Diagnostics
	result := m.planDhcpRelay(ctx, nil, cfg, &d)

	if d.HasError() {
		t.Fatalf("unexpected error on first create: %v", d)
	}

	var planSettings tf.DhcpSettings
	if dd := result.As(ctx, &planSettings, basetypes.ObjectAsOptions{}); dd.HasError() {
		t.Fatalf("failed to decode plan result: %v", dd)
	}

	if planSettings.RelayGroupName.ValueString() != "MY-DHCP" {
		t.Errorf("expected relay_group_name=%q, got %q", "MY-DHCP", planSettings.RelayGroupName.ValueString())
	}
	if !planSettings.RelayGroupID.IsUnknown() {
		t.Errorf("expected relay_group_id to be unknown on first create, got %q", planSettings.RelayGroupID.ValueString())
	}
}

// TestPlanDhcpRelay_FirstCreateByID verifies correct behavior on first create (no prior state)
// when the user configures only relay_group_id.
func TestPlanDhcpRelay_FirstCreateByID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	m := dhcpSettingsModifier{}

	cfg := &tf.DhcpSettings{
		DhcpType:              types.StringValue(string(cato_models.DhcpTypeDhcpRelay)),
		RelayGroupName:        types.StringNull(),
		RelayGroupID:          types.StringValue("4456"),
		IPRange:               types.StringNull(),
		DhcpMicrosegmentation: types.BoolNull(),
	}

	var d diag.Diagnostics
	result := m.planDhcpRelay(ctx, nil, cfg, &d)

	if d.HasError() {
		t.Fatalf("unexpected error on first create: %v", d)
	}

	var planSettings tf.DhcpSettings
	if dd := result.As(ctx, &planSettings, basetypes.ObjectAsOptions{}); dd.HasError() {
		t.Fatalf("failed to decode plan result: %v", dd)
	}

	if planSettings.RelayGroupID.ValueString() != "4456" {
		t.Errorf("expected relay_group_id=%q, got %q", "4456", planSettings.RelayGroupID.ValueString())
	}
	if !planSettings.RelayGroupName.IsUnknown() {
		t.Errorf("expected relay_group_name to be unknown on first create, got %q", planSettings.RelayGroupName.ValueString())
	}
}

// TestPlanModifyObject_DhcpRelayWithMicrosegmentationFalse verifies that
// dhcp_microsegmentation=false with dhcp_type=DHCP_RELAY does not produce an error.
// false is the zero/disabled value and must always be accepted regardless of dhcp_type.
func TestPlanModifyObject_DhcpRelayWithMicrosegmentationFalse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	m := dhcpSettingsModifier{}

	cfgSettings := tf.DhcpSettings{
		DhcpType:              types.StringValue(string(cato_models.DhcpTypeDhcpRelay)),
		RelayGroupName:        types.StringValue("CHCVTPJ-DHCP"),
		RelayGroupID:          types.StringNull(),
		IPRange:               types.StringNull(),
		DhcpMicrosegmentation: types.BoolValue(false),
	}

	req := planmodifier.ObjectRequest{
		ConfigValue: makeDhcpSettingsObj(t, cfgSettings),
		StateValue:  types.ObjectNull(tf.DhcpSettingsAttrTypes),
	}
	resp := &planmodifier.ObjectResponse{
		PlanValue: req.ConfigValue,
	}

	m.PlanModifyObject(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no error for dhcp_microsegmentation=false with DHCP_RELAY, got: %v",
			resp.Diagnostics)
	}
}

// TestPlanModifyObject_DhcpRelayWithMicrosegmentationTrue verifies that
// dhcp_microsegmentation=true with dhcp_type=DHCP_RELAY produces an error,
// since microsegmentation is only meaningful for DHCP_RANGE.
func TestPlanModifyObject_DhcpRelayWithMicrosegmentationTrue(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	m := dhcpSettingsModifier{}

	cfgSettings := tf.DhcpSettings{
		DhcpType:              types.StringValue(string(cato_models.DhcpTypeDhcpRelay)),
		RelayGroupName:        types.StringValue("CHCVTPJ-DHCP"),
		RelayGroupID:          types.StringNull(),
		IPRange:               types.StringNull(),
		DhcpMicrosegmentation: types.BoolValue(true),
	}

	req := planmodifier.ObjectRequest{
		ConfigValue: makeDhcpSettingsObj(t, cfgSettings),
		StateValue:  types.ObjectNull(tf.DhcpSettingsAttrTypes),
	}
	resp := &planmodifier.ObjectResponse{
		PlanValue: req.ConfigValue,
	}

	m.PlanModifyObject(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error for dhcp_microsegmentation=true with DHCP_RELAY, but got none")
	}
}

// TestPlanModifyObject_DhcpRangeWithMicrosegmentationTrue verifies that
// dhcp_microsegmentation=true with dhcp_type=DHCP_RANGE is accepted.
func TestPlanModifyObject_DhcpRangeWithMicrosegmentationTrue(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	m := dhcpSettingsModifier{}

	cfgSettings := tf.DhcpSettings{
		DhcpType:              types.StringValue(string(cato_models.DhcpTypeDhcpRange)),
		RelayGroupName:        types.StringNull(),
		RelayGroupID:          types.StringNull(),
		IPRange:               types.StringValue("10.0.0.10-10.0.0.100"),
		DhcpMicrosegmentation: types.BoolValue(true),
	}

	req := planmodifier.ObjectRequest{
		ConfigValue: makeDhcpSettingsObj(t, cfgSettings),
		StateValue:  types.ObjectNull(tf.DhcpSettingsAttrTypes),
	}
	resp := &planmodifier.ObjectResponse{
		PlanValue: req.ConfigValue,
	}

	m.PlanModifyObject(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no error for dhcp_microsegmentation=true with DHCP_RANGE, got: %v",
			resp.Diagnostics)
	}
}
