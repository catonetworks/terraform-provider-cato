package networkrange

import (
	"context"
	"strings"
	"testing"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/dhcp"
	tf "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/tfmodel"
)

func TestNetworkRangeInterfaceIDUsesConfiguredOnlyReplaceModifier(t *testing.T) {
	t.Parallel()

	r := &networkRangeResource{}
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attr, ok := resp.Schema.Attributes["interface_id"]
	if !ok {
		t.Fatalf("expected interface_id attribute in schema")
	}

	stringAttr, ok := attr.(schema.StringAttribute)
	if !ok {
		t.Fatalf("expected interface_id attribute to be schema.StringAttribute")
	}

	if len(stringAttr.PlanModifiers) < 2 {
		t.Fatalf("expected at least 2 plan modifiers on interface_id, got %d", len(stringAttr.PlanModifiers))
	}

	if !strings.Contains(strings.ToLower(stringAttr.PlanModifiers[1].Description(context.Background())), "configured") {
		t.Fatalf("expected interface_id replacement modifier to be configured-only")
	}
}

func TestNetworkRangeInterfaceIDReplaceModifierSkipsUnconfiguredChanges(t *testing.T) {
	t.Parallel()

	modifier := getInterfaceIDReplacementModifier(t)

	req := planmodifier.StringRequest{
		ConfigValue: types.StringNull(),
		StateValue:  types.StringValue("234341"),
		PlanValue:   types.StringUnknown(),
		State:       planmodifier.StringRequest{}.State,
		Plan:        planmodifier.StringRequest{}.Plan,
	}
	req.State.Raw = nonNullRawValue(t)
	req.Plan.Raw = nonNullRawValue(t)
	resp := &planmodifier.StringResponse{
		PlanValue: req.PlanValue,
	}

	modifier.PlanModifyString(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
	if resp.RequiresReplace {
		t.Fatalf("expected unconfigured interface_id change to not require replacement")
	}
}

func TestNetworkRangeInterfaceIDReplaceModifierTriggersWhenConfigured(t *testing.T) {
	t.Parallel()

	modifier := getInterfaceIDReplacementModifier(t)

	req := planmodifier.StringRequest{
		ConfigValue: types.StringValue("999999"),
		StateValue:  types.StringValue("234341"),
		PlanValue:   types.StringValue("999999"),
	}
	req.State.Raw = nonNullRawValue(t)
	req.Plan.Raw = nonNullRawValue(t)

	resp := &planmodifier.StringResponse{
		PlanValue: req.PlanValue,
	}

	modifier.PlanModifyString(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
	if !resp.RequiresReplace {
		t.Fatalf("expected configured interface_id change to require replacement")
	}
}

// TestModifyPlanNoFalsePositiveWhenBothFieldsEqualState verifies that a plan cycle where
// the plan carries both interface_id and interface_index from prior state, but config only
// contains interface_id, does not generate a conflict error.
// This is the exact root cause of the false positive seen in logs.txt: the provider stored
// bare "11" as interface_index in state, Terraform Core propagated it into req.Config, and
// the old validator fired because both fields were non-null.
func TestModifyPlanNoFalsePositiveWhenBothFieldsEqualState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	r := &networkRangeResource{client: &catoclient.Service{AccountId: "account-123"}}

	// Simulate prior state: both fields exist with concrete values.
	stateModel := networkRangeModel{
		InterfaceID:    types.StringValue("148383"),
		InterfaceIndex: types.StringValue("11"), // bare number as stored by older provider versions
	}

	planModel := networkRangeModel{
		InterfaceID:    types.StringValue("148383"),
		InterfaceIndex: types.StringValue("11"),
	}
	cfgModel := networkRangeModel{
		InterfaceID: types.StringValue("148383"),
	}

	plan := newNetworkRangePlan(ctx, t, planModel)
	resp := &resource.ModifyPlanResponse{Plan: plan}

	r.ModifyPlan(ctx, resource.ModifyPlanRequest{
		Plan:   plan,
		Config: newNetworkRangeConfig(ctx, t, cfgModel),
		State:  newNetworkRangeState(ctx, t, stateModel),
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no conflict error for state-propagated interface_index, got: %v", resp.Diagnostics)
	}
}

// TestModifyPlanNoFalsePositiveWhenConfigCarriesBothFieldsEqualState reproduces the request shape
// seen in ENG-193800 logs: Terraform passes both interface fields through req.Config even though
// one came from prior state.
func TestModifyPlanNoFalsePositiveWhenConfigCarriesBothFieldsEqualState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	r := &networkRangeResource{client: &catoclient.Service{AccountId: "account-123"}}

	stateModel := networkRangeModel{
		InterfaceID:    types.StringValue("148383"),
		InterfaceIndex: types.StringValue("11"),
	}
	cfgModel := networkRangeModel{
		InterfaceID:    types.StringValue("148383"),
		InterfaceIndex: types.StringValue("11"),
	}

	plan := newNetworkRangePlan(ctx, t, cfgModel)
	resp := &resource.ModifyPlanResponse{Plan: plan}

	r.ModifyPlan(ctx, resource.ModifyPlanRequest{
		Plan:   plan,
		Config: newNetworkRangeConfig(ctx, t, cfgModel),
		State:  newNetworkRangeState(ctx, t, stateModel),
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no conflict error for state-propagated interface fields, got: %v", resp.Diagnostics)
	}
}

// TestModifyPlanAllowsInterfaceIDChangeWithPropagatedIndex verifies that when the user
// changes interface_id to a new value while the plan carries prior-state interface_index,
// no conflict error is raised and interface_index becomes unknown.
func TestModifyPlanAllowsInterfaceIDChangeWithPropagatedIndex(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	r := &networkRangeResource{client: &catoclient.Service{AccountId: "account-123"}}

	stateModel := networkRangeModel{
		InterfaceID:    types.StringValue("148383"),
		InterfaceIndex: types.StringValue("INT_11"),
	}

	planModel := networkRangeModel{
		InterfaceID:    types.StringValue("999999"), // user changed this
		InterfaceIndex: types.StringValue("INT_11"), // plan carries prior state
	}
	cfgModel := networkRangeModel{
		InterfaceID: types.StringValue("999999"),
	}

	plan := newNetworkRangePlan(ctx, t, planModel)
	resp := &resource.ModifyPlanResponse{Plan: plan}

	r.ModifyPlan(ctx, resource.ModifyPlanRequest{
		Plan:   plan,
		Config: newNetworkRangeConfig(ctx, t, cfgModel),
		State:  newNetworkRangeState(ctx, t, stateModel),
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no error when only interface_id explicitly changed, got: %v", resp.Diagnostics)
	}

	var got tf.NetworkRange
	if diags := resp.Plan.Get(ctx, &got); diags.HasError() {
		t.Fatalf("failed to decode plan: %v", diags)
	}
	if got.InterfaceID.ValueString() != "999999" {
		t.Fatalf("expected interface_id to stay changed, got %q", got.InterfaceID.ValueString())
	}
	if !got.InterfaceIndex.IsUnknown() {
		t.Fatalf("expected propagated interface_index to become unknown, got %q", got.InterfaceIndex.ValueString())
	}
}

// TestModifyPlanAllowsInterfaceIDChangeWhenConfigCarriesPropagatedIndex verifies the log-shaped
// variant where req.Config includes the old interface_index along with a changed interface_id.
func TestModifyPlanAllowsInterfaceIDChangeWhenConfigCarriesPropagatedIndex(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	r := &networkRangeResource{client: &catoclient.Service{AccountId: "account-123"}}

	stateModel := networkRangeModel{
		InterfaceID:    types.StringValue("148383"),
		InterfaceIndex: types.StringValue("11"),
	}

	cfgModel := networkRangeModel{
		InterfaceID:    types.StringValue("999999"),
		InterfaceIndex: types.StringValue("11"), // propagated from state
	}

	plan := newNetworkRangePlan(ctx, t, cfgModel)
	resp := &resource.ModifyPlanResponse{Plan: plan}

	r.ModifyPlan(ctx, resource.ModifyPlanRequest{
		Plan:   plan,
		Config: newNetworkRangeConfig(ctx, t, cfgModel),
		State:  newNetworkRangeState(ctx, t, stateModel),
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no error when only interface_id explicitly changed, got: %v", resp.Diagnostics)
	}

	var got tf.NetworkRange
	if diags := resp.Plan.Get(ctx, &got); diags.HasError() {
		t.Fatalf("failed to decode plan: %v", diags)
	}
	if got.InterfaceID.ValueString() != "999999" {
		t.Fatalf("expected interface_id to stay changed, got %q", got.InterfaceID.ValueString())
	}
	if !got.InterfaceIndex.IsUnknown() {
		t.Fatalf("expected propagated interface_index to become unknown, got %q", got.InterfaceIndex.ValueString())
	}
}

// TestModifyPlanAllowsInterfaceIndexChangeWithPropagatedID verifies the symmetric case:
// when the user changes interface_index and Terraform propagates the old interface_id from
// state, the propagated ID does not override the actual interface_index change.
func TestModifyPlanAllowsInterfaceIndexChangeWithPropagatedID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	r := &networkRangeResource{client: &catoclient.Service{AccountId: "account-123"}}

	stateModel := networkRangeModel{
		InterfaceID:    types.StringValue("148383"),
		InterfaceIndex: types.StringValue("INT_11"),
	}

	planModel := networkRangeModel{
		InterfaceID:    types.StringValue("148383"), // plan carries prior state
		InterfaceIndex: types.StringValue("INT_5"),  // user changed this
	}
	cfgModel := networkRangeModel{
		InterfaceIndex: types.StringValue("INT_5"),
	}

	plan := newNetworkRangePlan(ctx, t, planModel)
	resp := &resource.ModifyPlanResponse{Plan: plan}

	r.ModifyPlan(ctx, resource.ModifyPlanRequest{
		Plan:   plan,
		Config: newNetworkRangeConfig(ctx, t, cfgModel),
		State:  newNetworkRangeState(ctx, t, stateModel),
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no error when only interface_index explicitly changed, got: %v", resp.Diagnostics)
	}

	var got tf.NetworkRange
	if diags := resp.Plan.Get(ctx, &got); diags.HasError() {
		t.Fatalf("failed to decode plan: %v", diags)
	}
	if got.InterfaceIndex.ValueString() != "INT_5" {
		t.Fatalf("expected interface_index to stay changed, got %q", got.InterfaceIndex.ValueString())
	}
	if !got.InterfaceID.IsUnknown() {
		t.Fatalf("expected propagated interface_id to become unknown, got %q", got.InterfaceID.ValueString())
	}
}

// TestModifyPlanAllowsRelayNameWithPropagatedRelayID verifies the customer scenario from
// ENG-193800: config only has relay_group_name and dhcp_microsegmentation=false, while old
// state contributes relay_group_id during planning.
func TestModifyPlanAllowsRelayNameWithPropagatedRelayID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	r := &networkRangeResource{client: &catoclient.Service{AccountId: "account-123"}}

	stateDhcp := makeNetworkRangeDhcpSettingsObj(t, dhcp.Settings{
		DhcpType:              types.StringValue(string(cato_models.DhcpTypeDhcpRelay)),
		RelayGroupName:        types.StringValue("CHCVTPJ-DHCP"),
		RelayGroupID:          types.StringValue("4456"),
		IPRange:               types.StringNull(),
		DhcpMicrosegmentation: types.BoolNull(),
	})
	planDhcp := makeNetworkRangeDhcpSettingsObj(t, dhcp.Settings{
		DhcpType:              types.StringValue(string(cato_models.DhcpTypeDhcpRelay)),
		RelayGroupName:        types.StringValue("CHCVTPJ-DHCP"),
		RelayGroupID:          types.StringValue("4456"), // plan carries prior state
		IPRange:               types.StringNull(),
		DhcpMicrosegmentation: types.BoolValue(false),
	})
	cfgDhcp := makeNetworkRangeDhcpSettingsObj(t, dhcp.Settings{
		DhcpType:              types.StringValue(string(cato_models.DhcpTypeDhcpRelay)),
		RelayGroupName:        types.StringValue("CHCVTPJ-DHCP"),
		RelayGroupID:          types.StringNull(),
		IPRange:               types.StringNull(),
		DhcpMicrosegmentation: types.BoolValue(false),
	})

	stateModel := networkRangeModel{
		InterfaceID:    types.StringValue("148383"),
		InterfaceIndex: types.StringValue("INT_11"),
		RangeType:      types.StringValue("VLAN"),
		Vlan:           types.Int64Value(812),
		DhcpSettings:   stateDhcp,
	}
	planModel := networkRangeModel{
		InterfaceID:    types.StringValue("148383"),
		InterfaceIndex: types.StringValue("INT_11"),
		RangeType:      types.StringValue("VLAN"),
		Vlan:           types.Int64Value(812),
		DhcpSettings:   planDhcp,
	}
	cfgModel := networkRangeModel{
		InterfaceID:  types.StringValue("148383"),
		RangeType:    types.StringValue("VLAN"),
		Vlan:         types.Int64Value(812),
		DhcpSettings: cfgDhcp,
	}

	plan := newNetworkRangePlan(ctx, t, planModel)
	resp := &resource.ModifyPlanResponse{Plan: plan}

	r.ModifyPlan(ctx, resource.ModifyPlanRequest{
		Plan:   plan,
		Config: newNetworkRangeConfig(ctx, t, cfgModel),
		State:  newNetworkRangeState(ctx, t, stateModel),
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no error for state-propagated relay_group_id, got: %v", resp.Diagnostics)
	}
}

// TestModifyPlanAllowsRelayNameWhenConfigCarriesPropagatedRelayID verifies the log-shaped
// ENG-193800 failure where req.Config carries both relay_group_name and a state-propagated
// relay_group_id before the DHCP plan modifier runs.
func TestModifyPlanAllowsRelayNameWhenConfigCarriesPropagatedRelayID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	r := &networkRangeResource{client: &catoclient.Service{AccountId: "account-123"}}

	stateDhcp := makeNetworkRangeDhcpSettingsObj(t, dhcp.Settings{
		DhcpType:              types.StringValue(string(cato_models.DhcpTypeDhcpRelay)),
		RelayGroupName:        types.StringValue("CHCVTPJ-DHCP"),
		RelayGroupID:          types.StringValue("4456"),
		IPRange:               types.StringNull(),
		DhcpMicrosegmentation: types.BoolNull(),
	})
	cfgDhcp := makeNetworkRangeDhcpSettingsObj(t, dhcp.Settings{
		DhcpType:              types.StringValue(string(cato_models.DhcpTypeDhcpRelay)),
		RelayGroupName:        types.StringValue("CHCVTPJ-DHCP"),
		RelayGroupID:          types.StringValue("4456"), // propagated from state, not user config
		IPRange:               types.StringNull(),
		DhcpMicrosegmentation: types.BoolValue(false),
	})

	stateModel := networkRangeModel{
		InterfaceID:    types.StringValue("148383"),
		InterfaceIndex: types.StringValue("INT_11"),
		RangeType:      types.StringValue("VLAN"),
		Vlan:           types.Int64Value(816),
		DhcpSettings:   stateDhcp,
	}
	cfgModel := networkRangeModel{
		InterfaceID:    types.StringValue("148383"),
		InterfaceIndex: types.StringValue("INT_11"),
		RangeType:      types.StringValue("VLAN"),
		Vlan:           types.Int64Value(816),
		DhcpSettings:   cfgDhcp,
	}

	plan := newNetworkRangePlan(ctx, t, cfgModel)
	resp := &resource.ModifyPlanResponse{Plan: plan}

	r.ModifyPlan(ctx, resource.ModifyPlanRequest{
		Plan:   plan,
		Config: newNetworkRangeConfig(ctx, t, cfgModel),
		State:  newNetworkRangeState(ctx, t, stateModel),
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no validator error for state-propagated relay_group_id, got: %v", resp.Diagnostics)
	}
}

// TestModifyPlanDetectsExplicitInterfaceConflict verifies that when a user explicitly writes
// both interface_id and interface_index with different values than the prior state, the
// conflict error is still correctly raised.
func TestModifyPlanDetectsExplicitInterfaceConflict(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	r := &networkRangeResource{client: &catoclient.Service{AccountId: "account-123"}}

	stateModel := networkRangeModel{
		InterfaceID:    types.StringValue("148383"),
		InterfaceIndex: types.StringValue("INT_11"),
	}

	// User explicitly changed BOTH fields to new values → genuine conflict.
	cfgModel := networkRangeModel{
		InterfaceID:    types.StringValue("999999"), // new value
		InterfaceIndex: types.StringValue("INT_5"),  // new value
	}

	plan := newNetworkRangePlan(ctx, t, cfgModel)
	resp := &resource.ModifyPlanResponse{Plan: plan}

	r.ModifyPlan(ctx, resource.ModifyPlanRequest{
		Plan:   plan,
		Config: newNetworkRangeConfig(ctx, t, cfgModel),
		State:  newNetworkRangeState(ctx, t, stateModel),
	}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected conflict error when both interface_id and interface_index are explicitly changed")
	}
}

func getInterfaceIDReplacementModifier(t *testing.T) planmodifier.String {
	t.Helper()

	r := &networkRangeResource{}
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	attr, ok := resp.Schema.Attributes["interface_id"]
	if !ok {
		t.Fatalf("expected interface_id attribute in schema")
	}

	stringAttr, ok := attr.(schema.StringAttribute)
	if !ok {
		t.Fatalf("expected interface_id attribute to be schema.StringAttribute")
	}

	if len(stringAttr.PlanModifiers) < 2 {
		t.Fatalf("expected at least 2 plan modifiers on interface_id, got %d", len(stringAttr.PlanModifiers))
	}

	return stringAttr.PlanModifiers[1]
}

func nonNullRawValue(t *testing.T) tftypes.Value {
	t.Helper()

	return tftypes.NewValue(tftypes.String, "non-null")
}

func makeNetworkRangeDhcpSettingsObj(t *testing.T, s dhcp.Settings) types.Object {
	t.Helper()
	obj, diags := types.ObjectValueFrom(context.Background(), tf.DhcpSettingsAttrTypes, s)
	if diags.HasError() {
		t.Fatalf("failed to create DHCP settings object: %v", diags)
	}
	return obj
}

func TestNetworkRangeSubnetTypeValidation(t *testing.T) {
	t.Parallel()
	v := SubnetTypeValidator{}
	for _, tc := range []struct {
		name      string
		value     types.String
		wantError bool
	}{
		{"valid", types.StringValue(string(cato_models.AllSubnetType[0])), false},
		{"invalid", types.StringValue("INVALID"), true},
		{"null", types.StringNull(), false},
		{"unknown", types.StringUnknown(), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp := &validator.StringResponse{}
			v.ValidateString(context.Background(), validator.StringRequest{Path: path.Root("value"), ConfigValue: tc.value}, resp)
			require.Equal(t, tc.wantError, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
		})
	}
}
