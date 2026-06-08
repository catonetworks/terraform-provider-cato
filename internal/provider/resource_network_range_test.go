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
	req := resource.CreateRequest{Plan: newNetworkRangePlan(ctx, t, networkRangeModel{
		InterfaceID:    types.StringValue("if-1"),
		InterfaceIndex: types.StringValue("LAN1"),
	})}
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: getNetworkRangeSchema(ctx, t)}}

	r.Create(ctx, req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostics for conflicting interface_id and interface_index")
	}
}

func TestNetworkRangeCreateReturnsDiagnosticsOnAPIError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockClient := mocks.NewNetworkRangeClient(t)
	mockClient.EXPECT().
		SiteAddNetworkRange(mock.Anything, "", mock.Anything, "account-123").
		Return(nil, errors.New("add failed")).
		Once()

	r := &networkRangeResource{
		client:             &catoClientData{AccountId: "account-123"},
		networkRangeClient: mockClient,
	}
	req := resource.CreateRequest{Plan: newNetworkRangePlan(ctx, t, networkRangeModel{})}
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
	req := resource.UpdateRequest{Plan: newNetworkRangePlan(ctx, t, networkRangeModel{ID: types.StringValue("nr-1")})}
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

func TestBuildAddNetworkRangeInput(t *testing.T) {
	t.Parallel()
	const (
		name    = "NR"
		xSubnet = "172.16.0.0/24"
	)

	model := networkRangeModel{
		Name:             types.StringValue("NR"),
		RangeType:        types.StringValue("Direct"),
		Subnet:           types.StringValue("10.0.0.0/24"),
		LocalIP:          types.StringValue("10.0.0.1"),
		TranslatedSubnet: types.StringValue(xSubnet),
	}

	input := buildAddNetworkRangeInput(model.toResourceModel(), nrBoolPtr(true))
	if input.Name != name {
		t.Fatalf("expected name %s, got %q", name, input.Name)
	}
	if input.TranslatedSubnet == nil || *input.TranslatedSubnet != xSubnet {
		t.Fatalf("expected translated subnet to be set, got %v", input.TranslatedSubnet)
	}
}

func TestBuildUpdateNetworkRangeInput(t *testing.T) {
	t.Parallel()

	model := networkRangeModel{
		Name:             types.StringValue("NR"),
		RangeType:        types.StringValue("Direct"),
		Subnet:           types.StringValue("10.0.0.0/24"),
		LocalIP:          types.StringValue("10.0.0.1"),
		TranslatedSubnet: types.StringValue("172.16.0.0/24"),
	}

	input := buildUpdateNetworkRangeInput(model.toResourceModel(), nrBoolPtr(false), nil)
	if input.Name == nil || *input.Name != "NR" {
		t.Fatalf("expected name NR, got %v", input.Name)
	}
	if input.TranslatedSubnet == nil || *input.TranslatedSubnet != "172.16.0.0/24" {
		t.Fatalf("expected translated subnet to be set, got %v", input.TranslatedSubnet)
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

func nrBoolPtr(v bool) *bool {
	return &v
}
