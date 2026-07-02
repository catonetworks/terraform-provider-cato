package provider

import (
	"context"
	"errors"
	"strings"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/mock"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
	tf "github.com/catonetworks/terraform-provider-cato/internal/provider/tfmodel"
)

const testSiteID = "site-1"

func TestNewNetworkRangeResource(t *testing.T) {
	t.Parallel()

	r := NewNetworkRangeResource()
	if r == nil {
		t.Fatal("expected resource instance, got nil")
	}
	if _, ok := r.(*networkRangeResource); !ok {
		t.Fatalf("expected *networkRangeResource, got %T", r)
	}
}

func TestNetworkRangeMetadata(t *testing.T) {
	t.Parallel()

	r := &networkRangeResource{}
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "cato"}, resp)

	if resp.TypeName != "cato_network_range" {
		t.Fatalf("expected type name cato_network_range, got %q", resp.TypeName)
	}
}

func TestNetworkRangeConfigureNilProviderData(t *testing.T) {
	t.Parallel()

	r := &networkRangeResource{}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), resource.ConfigureRequest{}, resp)

	if r.client != nil {
		t.Fatal("expected client to stay nil")
	}
}

func TestNetworkRangeConfigureSetsClient(t *testing.T) {
	t.Parallel()

	client := &catoClientData{AccountId: "account-123"}
	r := &networkRangeResource{}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), resource.ConfigureRequest{ProviderData: client}, resp)

	if r.client != client {
		t.Fatal("expected client to be set from provider data")
	}
}

func TestNetworkRangeGetNetworkRangeClient(t *testing.T) {
	t.Parallel()

	t.Run("uses injected mock client", func(t *testing.T) {
		t.Parallel()
		mockClient := mocks.NewNetworkRangeClient(t)
		r := &networkRangeResource{networkRangeClient: mockClient}
		if got := r.getNetworkRangeClient(); got != mockClient {
			t.Fatalf("expected injected client, got %T", got)
		}
	})

	t.Run("returns nil when no provider client", func(t *testing.T) {
		t.Parallel()
		r := &networkRangeResource{}
		if got := r.getNetworkRangeClient(); got != nil {
			t.Fatalf("expected nil client, got %T", got)
		}
	})
}

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

func TestNetworkRangeCreateRejectsConflictingInterfaceConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	r := &networkRangeResource{client: &catoClientData{AccountId: "account-123"}}
	model := networkRangeModel{
		InterfaceID:    types.StringValue("if-1"),
		InterfaceIndex: types.StringValue("LAN1"),
	}
	req := resource.CreateRequest{
		Plan:   newNetworkRangePlan(ctx, t, model),
		Config: newNetworkRangeConfig(ctx, t, model),
	}
	resp := &resource.ModifyPlanResponse{Plan: req.Plan}

	r.ModifyPlan(ctx, resource.ModifyPlanRequest{
		Plan:   req.Plan,
		Config: req.Config,
	}, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostics for conflicting interface_id and interface_index")
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
	r := &networkRangeResource{client: &catoClientData{AccountId: "account-123"}}

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
	r := &networkRangeResource{client: &catoClientData{AccountId: "account-123"}}

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
	r := &networkRangeResource{client: &catoClientData{AccountId: "account-123"}}

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
	r := &networkRangeResource{client: &catoClientData{AccountId: "account-123"}}

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
	r := &networkRangeResource{client: &catoClientData{AccountId: "account-123"}}

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
	r := &networkRangeResource{client: &catoClientData{AccountId: "account-123"}}

	stateDhcp := makeNetworkRangeDhcpSettingsObj(t, tf.DhcpSettings{
		DhcpType:              types.StringValue(string(cato_models.DhcpTypeDhcpRelay)),
		RelayGroupName:        types.StringValue("CHCVTPJ-DHCP"),
		RelayGroupID:          types.StringValue("4456"),
		IPRange:               types.StringNull(),
		DhcpMicrosegmentation: types.BoolNull(),
	})
	planDhcp := makeNetworkRangeDhcpSettingsObj(t, tf.DhcpSettings{
		DhcpType:              types.StringValue(string(cato_models.DhcpTypeDhcpRelay)),
		RelayGroupName:        types.StringValue("CHCVTPJ-DHCP"),
		RelayGroupID:          types.StringValue("4456"), // plan carries prior state
		IPRange:               types.StringNull(),
		DhcpMicrosegmentation: types.BoolValue(false),
	})
	cfgDhcp := makeNetworkRangeDhcpSettingsObj(t, tf.DhcpSettings{
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

// TestModifyPlanDetectsExplicitInterfaceConflict verifies that when a user explicitly writes
// both interface_id and interface_index with different values than the prior state, the
// conflict error is still correctly raised.
func TestModifyPlanDetectsExplicitInterfaceConflict(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	r := &networkRangeResource{client: &catoClientData{AccountId: "account-123"}}

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

func TestNetworkRangeCreateReturnsDiagnosticsOnAPIError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockClient := mocks.NewNetworkRangeClient(t)
	mockClient.EXPECT().
		SiteAddNetworkRange(mock.Anything, "if-1", mock.Anything, "account-123").
		Return(nil, errors.New("add failed")).
		Once()

	r := &networkRangeResource{
		client:             &catoClientData{AccountId: "account-123"},
		networkRangeClient: mockClient,
	}
	model := networkRangeModel{InterfaceID: types.StringValue("if-1")}
	req := resource.CreateRequest{
		Plan:   newNetworkRangePlan(ctx, t, model),
		Config: newNetworkRangeConfig(ctx, t, model),
	}
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: getNetworkRangeSchema(ctx, t)}}

	r.Create(ctx, req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostics for add API error")
	}
}

func TestNetworkRangeReadSuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockClient := mocks.NewNetworkRangeClient(t)
	mockClient.EXPECT().
		NetworkRange(mock.Anything, "account-123", "nr-1").
		Return(newNetworkRangeAPIResponse("nr-1"), nil).
		Once()

	r := &networkRangeResource{
		client:             &catoClientData{AccountId: "account-123"},
		networkRangeClient: mockClient,
	}
	req := resource.ReadRequest{State: newNetworkRangeState(ctx, t, networkRangeModel{
		ID:          types.StringValue("nr-1"),
		InterfaceID: types.StringValue("if-1"),
	})}
	resp := &resource.ReadResponse{State: tfsdk.State{Schema: getNetworkRangeSchema(ctx, t)}}

	r.Read(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}
}

func TestNetworkRangeUpdateReturnsDiagnosticsOnAPIError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockClient := mocks.NewNetworkRangeClient(t)
	mockClient.EXPECT().
		SiteUpdateNetworkRange(mock.Anything, "nr-1", mock.Anything, "account-123").
		Return(nil, errors.New("update failed")).
		Once()

	r := &networkRangeResource{
		client:             &catoClientData{AccountId: "account-123"},
		networkRangeClient: mockClient,
	}
	model := networkRangeModel{ID: types.StringValue("nr-1")}
	req := resource.UpdateRequest{
		Plan:   newNetworkRangePlan(ctx, t, model),
		Config: newNetworkRangeConfig(ctx, t, model),
	}
	resp := &resource.UpdateResponse{State: tfsdk.State{Schema: getNetworkRangeSchema(ctx, t)}}

	r.Update(ctx, req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostics for update API error")
	}
}

func TestNetworkRangeDeleteHandlesNotFoundGracefully(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockClient := mocks.NewNetworkRangeClient(t)
	mockClient.EXPECT().
		SiteRemoveNetworkRange(mock.Anything, "nr-1", "account-123").
		Return(nil, errors.New(`{"graphqlErrors":[{"message":"Network range with id: nr-1 is not found"}]}`)).
		Once()

	r := &networkRangeResource{
		client:             &catoClientData{AccountId: "account-123"},
		networkRangeClient: mockClient,
	}
	req := resource.DeleteRequest{State: newNetworkRangeState(ctx, t, networkRangeModel{ID: types.StringValue("nr-1")})}
	resp := &resource.DeleteResponse{}

	r.Delete(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no diagnostics for not found delete, got %+v", resp.Diagnostics)
	}
}

func TestNetworkRangeImportStateNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockClient := mocks.NewNetworkRangeClient(t)
	mockClient.EXPECT().
		NetworkRange(mock.Anything, "account-123", "nr-missing").
		Return(&cato.NetworkRange{Site: cato.NetworkRange_Site{NetworkRange: nil}}, nil).
		Once()

	r := &networkRangeResource{
		client:             &catoClientData{AccountId: "account-123"},
		networkRangeClient: mockClient,
	}
	req := resource.ImportStateRequest{ID: "nr-missing"}
	resp := &resource.ImportStateResponse{State: tfsdk.State{Schema: getNetworkRangeSchema(ctx, t)}}

	r.ImportState(ctx, req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostics for missing network range during import")
	}
}

func TestHydrateNetworkRangeStateNetworkRangeError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockClient := mocks.NewNetworkRangeClient(t)
	mockClient.EXPECT().
		NetworkRange(mock.Anything, "account-123", "nr-1").
		Return(nil, errors.New("boom")).
		Once()

	r := &networkRangeResource{
		client:             &catoClientData{AccountId: "account-123"},
		networkRangeClient: mockClient,
	}

	var diags diag.Diagnostics
	state := networkRangeModel{}.toResourceModel()
	r.hydrateNetworkRangeState(ctx, nil, &state, "nr-1", &diags)
	if !diags.HasError() {
		t.Fatal("expected hydrate error")
	}
}

func TestGetSiteIDFromNetworkRange(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockClient := mocks.NewNetworkRangeClient(t)
	mockClient.EXPECT().
		EntityLookup(
			mock.Anything,
			"account-123",
			cato_models.EntityType("siteRange"),
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			[]string{"nr-1"},
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).
		Return(&cato.EntityLookup{
			EntityLookup: cato.EntityLookup_EntityLookup{
				Items: []*cato.EntityLookup_EntityLookup_Items{
					{
						Entity: cato.EntityLookup_EntityLookup_Items_Entity{ID: "nr-1", Type: cato_models.EntityType("siteRange")},
						HelperFields: map[string]any{
							"siteId":        testSiteID,
							"interfaceName": "LAN 1",
						},
					},
				},
			},
		}, nil).
		Once()

	r := &networkRangeResource{
		client:             &catoClientData{AccountId: "account-123"},
		networkRangeClient: mockClient,
	}

	siteID, interfaceName, err := r.getSiteIDFromNetworkRange(ctx, "nr-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if siteID != testSiteID {
		t.Fatalf("expected site-1, got %q", siteID)
	}
	if interfaceName != "LAN 1" {
		t.Fatalf("expected LAN 1, got %q", interfaceName)
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

func makeNetworkRangeDhcpSettingsObj(t *testing.T, s tf.DhcpSettings) types.Object {
	t.Helper()
	obj, diags := types.ObjectValueFrom(context.Background(), tf.DhcpSettingsAttrTypes, s)
	if diags.HasError() {
		t.Fatalf("failed to create DHCP settings object: %v", diags)
	}
	return obj
}

type networkRangeModel struct {
	ID               types.String
	SiteID           types.String
	InterfaceID      types.String
	InterfaceIndex   types.String
	Name             types.String
	RangeType        types.String
	Subnet           types.String
	LocalIP          types.String
	Gateway          types.String
	TranslatedSubnet types.String
	DhcpSettings     types.Object
	InternetOnly     types.Bool
	MdnsReflector    types.Bool
	Vlan             types.Int64
}

func (m networkRangeModel) toResourceModel() tf.NetworkRange { //nolint:gocyclo
	id := m.ID
	if id.IsNull() && !id.IsUnknown() && id.ValueString() == "" {
		id = types.StringNull()
	}
	siteID := m.SiteID
	if siteID.IsNull() && !siteID.IsUnknown() && siteID.ValueString() == "" {
		siteID = types.StringValue(testSiteID)
	}
	name := m.Name
	if name.IsNull() && !name.IsUnknown() && name.ValueString() == "" {
		name = types.StringValue("nr-1")
	}
	rangeType := m.RangeType
	if rangeType.IsNull() && !rangeType.IsUnknown() && rangeType.ValueString() == "" {
		rangeType = types.StringValue("Direct")
	}
	subnet := m.Subnet
	if subnet.IsNull() && !subnet.IsUnknown() && subnet.ValueString() == "" {
		subnet = types.StringValue("10.0.0.0/24")
	}
	localIP := m.LocalIP
	if localIP.IsNull() && !localIP.IsUnknown() && localIP.ValueString() == "" {
		localIP = types.StringValue("10.0.0.1")
	}
	gateway := m.Gateway
	if gateway.IsNull() && !gateway.IsUnknown() && gateway.ValueString() == "" {
		gateway = types.StringNull()
	}
	interfaceID := m.InterfaceID
	if interfaceID.IsNull() && !interfaceID.IsUnknown() && interfaceID.ValueString() == "" {
		interfaceID = types.StringNull()
	}
	interfaceIndex := m.InterfaceIndex
	if interfaceIndex.IsNull() && !interfaceIndex.IsUnknown() && interfaceIndex.ValueString() == "" {
		interfaceIndex = types.StringNull()
	}
	dhcp := m.DhcpSettings
	if dhcp.IsNull() && !dhcp.IsUnknown() && len(dhcp.Attributes()) == 0 {
		dhcp = types.ObjectNull(tf.DhcpSettingsAttrTypes)
	}
	internetOnly := m.InternetOnly
	if internetOnly.IsNull() && !internetOnly.IsUnknown() {
		internetOnly = types.BoolValue(false)
	}
	mdnsReflector := m.MdnsReflector
	if mdnsReflector.IsNull() && !mdnsReflector.IsUnknown() {
		mdnsReflector = types.BoolValue(false)
	}
	vlan := m.Vlan
	if vlan.IsNull() && !vlan.IsUnknown() && vlan.ValueInt64() == 0 {
		vlan = types.Int64Null()
	}

	return tf.NetworkRange{
		ID:               id,
		SiteID:           siteID,
		InterfaceID:      interfaceID,
		InterfaceIndex:   interfaceIndex,
		Name:             name,
		RangeType:        rangeType,
		Subnet:           subnet,
		LocalIP:          localIP,
		Gateway:          gateway,
		TranslatedSubnet: m.TranslatedSubnet,
		DhcpSettings:     dhcp,
		InternetOnly:     internetOnly,
		MdnsReflector:    mdnsReflector,
		Vlan:             vlan,
	}
}

func newNetworkRangeAPIResponse(id string) *cato.NetworkRange {
	return &cato.NetworkRange{
		Site: cato.NetworkRange_Site{
			NetworkRange: &cato.NetworkRange_Site_NetworkRange{
				NetworkRangeID: id,
				Name:           "nr-1",
				RangeType:      cato_models.SubnetTypeDirect,
				Subnet:         "10.0.0.0/24",
				InternetOnly:   false,
				MdnsReflector:  false,
				LocalIP:        nrStringPtr("10.0.0.1"),
			},
		},
	}
}

func getNetworkRangeSchema(ctx context.Context, t *testing.T) schema.Schema {
	t.Helper()
	r := &networkRangeResource{}
	resp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, resp)
	return resp.Schema
}

func newNetworkRangePlan(ctx context.Context, t *testing.T, model networkRangeModel) tfsdk.Plan {
	t.Helper()
	plan := tfsdk.Plan{Schema: getNetworkRangeSchema(ctx, t)}
	diags := plan.Set(ctx, model.toResourceModel())
	if diags.HasError() {
		t.Fatalf("unexpected plan diagnostics: %+v", diags)
	}
	return plan
}

func newNetworkRangeConfig(ctx context.Context, t *testing.T, model networkRangeModel) tfsdk.Config {
	t.Helper()
	plan := newNetworkRangePlan(ctx, t, model)
	return tfsdk.Config(plan)
}

func newNetworkRangeState(ctx context.Context, t *testing.T, model networkRangeModel) tfsdk.State {
	t.Helper()
	state := tfsdk.State{Schema: getNetworkRangeSchema(ctx, t)}
	diags := state.Set(ctx, model.toResourceModel())
	if diags.HasError() {
		t.Fatalf("unexpected state diagnostics: %+v", diags)
	}
	return state
}

func nrStringPtr(v string) *string {
	return &v
}
