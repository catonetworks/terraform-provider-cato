package socketsite

import (
	"context"
	"errors"
	"testing"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/mock"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/convert"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/dhcp"
	tf "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/tfmodel"
)

func TestNewSocketSiteResource(t *testing.T) {
	t.Parallel()

	r := NewSocketSiteResource()

	if r == nil {
		t.Fatal("expected resource instance, got nil")
	}

	if _, ok := r.(*socketSiteResource); !ok {
		t.Fatalf("expected *socketSiteResource, got %T", r)
	}
}

func TestSocketSiteMetadata(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	r := &socketSiteResource{}
	resp := &resource.MetadataResponse{}

	r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "cato"}, resp)

	if resp.TypeName != "cato_socket_site" {
		t.Fatalf("expected type name cato_socket_site, got %q", resp.TypeName)
	}
}

func TestSocketSiteConfigureNilProviderData(t *testing.T) {
	t.Parallel()

	r := &socketSiteResource{}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), resource.ConfigureRequest{}, resp)

	if r.client != nil {
		t.Fatal("expected client to remain nil when provider data is nil")
	}
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}
}

func TestSocketSiteConfigureSetsClient(t *testing.T) {
	t.Parallel()

	client := &catoclient.Service{AccountId: "account-123"}
	r := &socketSiteResource{}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), resource.ConfigureRequest{ProviderData: client}, resp)

	if r.client != client {
		t.Fatal("expected resource client to be set from provider data")
	}
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
	}
}

func TestSocketSiteGetSocketSiteClient(t *testing.T) {
	t.Parallel()

	t.Run("nil_without_provider_client", func(t *testing.T) {
		t.Parallel()

		r := &socketSiteResource{}
		if got := r.getSocketSiteClient(); got != nil {
			t.Fatalf("expected nil client, got %T", got)
		}
	})

	t.Run("uses_injected_client", func(t *testing.T) {
		t.Parallel()

		mockClient := mocks.NewSocketSiteClient(t)
		r := &socketSiteResource{socketSiteClient: mockClient}
		if got := r.getSocketSiteClient(); got != mockClient {
			t.Fatalf("expected injected client, got %T", got)
		}
	})

	t.Run("falls_back_to_provider_client", func(t *testing.T) {
		t.Parallel()

		sdkClient := &cato_go_sdk.Client{}
		providerClient := catoclient.NewProviderSDKClient(sdkClient)
		r := &socketSiteResource{client: &catoclient.Service{Catov2: providerClient}}
		if got := r.getSocketSiteClient(); got != providerClient {
			t.Fatalf("expected provider SDK client, got %T", got)
		}
	})
}

func TestSocketSiteSchemaIncludesCloudRouter(t *testing.T) {
	t.Parallel()

	resp := &resource.SchemaResponse{}
	(&socketSiteResource{}).Schema(context.Background(), resource.SchemaRequest{}, resp)

	attr, ok := resp.Schema.Attributes["is_cloud_router"].(schema.BoolAttribute)
	if !ok {
		t.Fatalf("expected is_cloud_router bool attribute, got %T", resp.Schema.Attributes["is_cloud_router"])
	}
	if !attr.Optional {
		t.Fatal("expected is_cloud_router to be optional")
	}
	if len(attr.PlanModifiers) != 1 {
		t.Fatalf("expected one plan modifier, got %d", len(attr.PlanModifiers))
	}
	if len(attr.Validators) != 1 {
		t.Fatalf("expected one validator, got %d", len(attr.Validators))
	}
	if attr.Description == "" {
		t.Fatal("expected is_cloud_router description")
	}
}

func TestSocketSiteFetchSocketConfiguration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockClient := mocks.NewSocketSiteClient(t)
	configuration := socketConfigurationForTest(cato_models.SocketModelX1700, true)
	apiResponse := &cato_go_sdk.SiteSocketConfiguration{
		Site: cato_go_sdk.SiteSocketConfiguration_Site{SiteSocketConfiguration: configuration},
	}
	mockClient.EXPECT().SiteSocketConfiguration(
		mock.Anything,
		mock.MatchedBy(func(input cato_models.SiteSocketConfigurationInput) bool {
			return input.Site != nil &&
				input.Site.By == cato_models.ObjectRefByID &&
				input.Site.Input == "site-123"
		}),
		"account-123",
	).Return(apiResponse, nil).Once()
	r := &socketSiteResource{
		client:           &catoclient.Service{AccountId: "account-123"},
		socketSiteClient: mockClient,
	}
	var diags diag.Diagnostics

	got := r.fetchSocketConfiguration(ctx, "site-123", &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}
	if got != configuration {
		t.Fatal("expected returned socket configuration")
	}
}

func TestSocketSiteFetchSocketConfigurationError(t *testing.T) {
	t.Parallel()

	mockClient := mocks.NewSocketSiteClient(t)
	mockClient.EXPECT().SiteSocketConfiguration(mock.Anything, mock.Anything, "account-123").
		Return(nil, errors.New("query failed")).Once()
	r := &socketSiteResource{
		client:           &catoclient.Service{AccountId: "account-123"},
		socketSiteClient: mockClient,
	}
	var diags diag.Diagnostics

	got := r.fetchSocketConfiguration(context.Background(), "site-123", &diags)

	if got != nil {
		t.Fatal("expected nil socket configuration")
	}
	if !diags.HasError() {
		t.Fatal("expected query error diagnostic")
	}
}

func TestSocketSiteUpdateCloudRouter(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		value bool
	}{
		"enabled":  {value: true},
		"disabled": {value: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			mockClient := mocks.NewSocketSiteClient(t)
			mockClient.EXPECT().SiteUpdateHa(
				mock.Anything,
				"account-123",
				"site-123",
				mock.MatchedBy(func(input cato_models.UpdateHaInput) bool {
					return input.IsCloudRouter != nil && *input.IsCloudRouter == tt.value
				}),
			).Return(&cato_go_sdk.SiteUpdateHa{}, nil).Once()
			r := &socketSiteResource{
				client:           &catoclient.Service{AccountId: "account-123"},
				socketSiteClient: mockClient,
			}
			plan := &SocketSite{IsCloudRouter: types.BoolValue(tt.value)}
			var diags diag.Diagnostics

			r.updateCloudRouter(context.Background(), plan, "site-123", &diags)

			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %+v", diags)
			}
		})
	}
}

func TestSocketSiteUpdateCloudRouterOmitted(t *testing.T) {
	t.Parallel()

	mockClient := mocks.NewSocketSiteClient(t)
	r := &socketSiteResource{
		client:           &catoclient.Service{AccountId: "account-123"},
		socketSiteClient: mockClient,
	}
	plan := &SocketSite{IsCloudRouter: types.BoolNull()}
	var diags diag.Diagnostics

	r.updateCloudRouter(context.Background(), plan, "site-123", &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}
}

func TestSocketSiteUpdateCloudRouterError(t *testing.T) {
	t.Parallel()

	mockClient := mocks.NewSocketSiteClient(t)
	mockClient.EXPECT().SiteUpdateHa(
		mock.Anything,
		"account-123",
		"site-123",
		mock.MatchedBy(func(input cato_models.UpdateHaInput) bool {
			return input.IsCloudRouter != nil && *input.IsCloudRouter
		}),
	).Return(nil, errors.New("update ha failed")).Once()
	r := &socketSiteResource{
		client:           &catoclient.Service{AccountId: "account-123"},
		socketSiteClient: mockClient,
	}
	plan := &SocketSite{IsCloudRouter: types.BoolValue(true)}
	var diags diag.Diagnostics

	r.updateCloudRouter(context.Background(), plan, "site-123", &diags)

	if !diags.HasError() {
		t.Fatal("expected diagnostics")
	}
}

func TestSocketSiteCleanupCreatedSocketSite(t *testing.T) {
	t.Parallel()

	mockClient := mocks.NewSocketSiteClient(t)
	mockClient.EXPECT().SiteRemoveSite(mock.Anything, "site-123", "account-123").
		Return(&cato_go_sdk.SiteRemoveSite{}, nil).Once()
	r := &socketSiteResource{
		client:           &catoclient.Service{AccountId: "account-123"},
		socketSiteClient: mockClient,
	}
	var diags diag.Diagnostics

	r.cleanupCreatedSocketSite(context.Background(), "site-123", &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}
}

func TestSocketSiteCleanupCreatedSocketSiteError(t *testing.T) {
	t.Parallel()

	mockClient := mocks.NewSocketSiteClient(t)
	mockClient.EXPECT().SiteRemoveSite(mock.Anything, "site-123", "account-123").
		Return(nil, errors.New("remove site failed")).Once()
	r := &socketSiteResource{
		client:           &catoclient.Service{AccountId: "account-123"},
		socketSiteClient: mockClient,
	}
	var diags diag.Diagnostics

	r.cleanupCreatedSocketSite(context.Background(), "site-123", &diags)

	if !diags.HasError() {
		t.Fatal("expected cleanup failure diagnostic")
	}
}

func TestSocketSiteConfigureCloudRouterAfterCreateCleansUpOnError(t *testing.T) {
	t.Parallel()

	mockClient := mocks.NewSocketSiteClient(t)
	mockClient.EXPECT().SiteUpdateHa(
		mock.Anything,
		"account-123",
		"site-123",
		mock.Anything,
	).Return(nil, errors.New("update ha failed")).Once()
	mockClient.EXPECT().SiteRemoveSite(mock.Anything, "site-123", "account-123").
		Return(&cato_go_sdk.SiteRemoveSite{}, nil).Once()
	r := &socketSiteResource{
		client:           &catoclient.Service{AccountId: "account-123"},
		socketSiteClient: mockClient,
	}
	var diags diag.Diagnostics

	r.configureCloudRouterAfterCreate(context.Background(), &SocketSite{
		IsCloudRouter: types.BoolValue(true),
	}, "site-123", &diags)

	if !diags.HasError() {
		t.Fatal("expected Cloud Router update failure diagnostic")
	}
}

func TestConnectionTypeFromSocketConfiguration(t *testing.T) {
	t.Parallel()

	tests := map[cato_models.SocketModel]string{
		cato_models.SocketModelAWS:      "SOCKET_AWS1500",
		cato_models.SocketModelAzure:    "SOCKET_AZ1500",
		cato_models.SocketModelEsx:      "SOCKET_ESX1500",
		cato_models.SocketModelGCP:      "SOCKET_GCP1500",
		cato_models.SocketModelX1500:    "SOCKET_X1500",
		cato_models.SocketModelX1600:    "SOCKET_X1600",
		cato_models.SocketModelX1600Lte: "SOCKET_X1600_LTE",
		cato_models.SocketModelX1700:    "SOCKET_X1700",
	}
	for model, want := range tests {
		t.Run(string(model), func(t *testing.T) {
			t.Parallel()

			got := connectionTypeFromSocketConfiguration(socketConfigurationForTest(model, true))
			if got.ValueString() != want {
				t.Fatalf("expected %q, got %q", want, got.ValueString())
			}
		})
	}

	if got := connectionTypeFromSocketConfiguration(nil); !got.IsNull() {
		t.Fatalf("expected null for missing configuration, got %s", got)
	}
}

func TestSocketSiteParseSockets(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	primarySerial, primaryPlatform := "primary-serial", "X1700"
	secondarySerial, secondaryPlatform := "secondary-serial", "X1700"
	configuration := socketConfigurationForTest(cato_models.SocketModelX1700, true)
	configuration.PrimarySocketConfiguration.Serial = &primarySerial
	configuration.PrimarySocketConfiguration.SocketInfo.Platform = &primaryPlatform
	configuration.SecondarySocketConfiguration =
		&cato_go_sdk.SiteSocketConfiguration_Site_SiteSocketConfiguration_SecondarySocketConfiguration{
			Serial: &secondarySerial,
			SocketInfo: cato_go_sdk.SiteSocketConfiguration_Site_SiteSocketConfiguration_SecondarySocketConfiguration_SocketInfo{
				IsPrimary: false,
				Platform:  &secondaryPlatform,
			},
		}
	var diags diag.Diagnostics

	got := (&socketSiteResource{}).parseSockets(ctx, configuration, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}
	var sockets []Socket
	diags = got.ElementsAs(ctx, &sockets, false)
	if diags.HasError() {
		t.Fatalf("decode sockets: %+v", diags)
	}
	if len(sockets) != 2 {
		t.Fatalf("expected two sockets, got %d", len(sockets))
	}
	socketBySerial := map[string]Socket{}
	for _, socket := range sockets {
		socketBySerial[socket.SerialNumber.ValueString()] = socket
		if !socket.ID.IsNull() {
			t.Fatalf("expected socket ID to be null, got %q", socket.ID.ValueString())
		}
	}
	if !socketBySerial[primarySerial].IsPrimary.ValueBool() {
		t.Fatal("expected primary socket to be marked primary")
	}
	if socketBySerial[secondarySerial].IsPrimary.ValueBool() {
		t.Fatal("expected secondary socket not to be marked primary")
	}
}

func socketConfigurationForTest(model cato_models.SocketModel, isPrimary bool,
) *cato_go_sdk.SiteSocketConfiguration_Site_SiteSocketConfiguration {
	return &cato_go_sdk.SiteSocketConfiguration_Site_SiteSocketConfiguration{
		PrimarySocketConfiguration: cato_go_sdk.SiteSocketConfiguration_Site_SiteSocketConfiguration_PrimarySocketConfiguration{
			SocketInfo: cato_go_sdk.SiteSocketConfiguration_Site_SiteSocketConfiguration_PrimarySocketConfiguration_SocketInfo{
				IsPrimary: isPrimary,
				Model:     &model,
			},
		},
	}
}

func TestStringPointerForOptionalInput(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		value types.String
		want  *string
	}{
		"null": {
			value: types.StringNull(),
			want:  nil,
		},
		"unknown": {
			value: types.StringUnknown(),
			want:  nil,
		},
		"empty": {
			value: types.StringValue(""),
			want:  nil,
		},
		"value": {
			value: types.StringValue("192.168.20.0/24"),
			want:  stringPtr("192.168.20.0/24"),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := convert.StringPointerForOptionalInput(tt.value)
			if tt.want == nil {
				if got != nil {
					t.Fatalf("expected nil, got %q", *got)
				}
				return
			}

			if got == nil {
				t.Fatalf("expected %q, got nil", *tt.want)
				return
			}
			if *got != *tt.want {
				t.Fatalf("expected %q, got %q", *tt.want, *got)
			}
		})
	}
}

func stringPtr(value string) *string {
	return &value
}

func TestSocketSitePrepareInputsTranslatedSubnet(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		translatedSubnet types.String
		wantNil          bool
		wantValue        string
	}{
		"null_omitted": {
			translatedSubnet: types.StringNull(),
			wantNil:          true,
		},
		"empty_omitted": {
			translatedSubnet: types.StringValue(""),
			wantNil:          true,
		},
		"unknown_omitted": {
			translatedSubnet: types.StringUnknown(),
			wantNil:          true,
		},
		"value_set": {
			translatedSubnet: types.StringValue("192.168.20.0/24"),
			wantValue:        "192.168.20.0/24",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			plan := newSocketSitePlanWithTranslatedSubnet(ctx, t, tt.translatedSubnet)
			cfg := plan
			client := &catoclient.Service{}
			var diags diag.Diagnostics

			addInput := prepareSocketSiteInput(ctx, plan, &diags)
			if diags.HasError() {
				t.Fatalf("prepareSocketSiteInput: %v", diags)
			}
			assertTranslatedSubnetPointer(t, addInput.TranslatedSubnet, tt.wantNil, tt.wantValue)

			networkRangeInput := prepareNetworkRangeInput(ctx, client, cfg, plan, false, &diags)
			if diags.HasError() {
				t.Fatalf("prepareNetworkRangeInput: %v", diags)
			}
			assertTranslatedSubnetPointer(t, networkRangeInput.TranslatedSubnet, tt.wantNil, tt.wantValue)

			socketIfaceInput, _ := prepareSocketInterfaceInput(ctx, cfg, plan, false, &diags)
			if diags.HasError() {
				t.Fatalf("prepareSocketInterfaceInput: %v", diags)
			}
			if socketIfaceInput.Lan == nil {
				t.Fatal("expected LAN input")
			}
			assertTranslatedSubnetPointer(t, socketIfaceInput.Lan.TranslatedSubnet, tt.wantNil, tt.wantValue)
		})
	}
}

func TestSocketSitePrepareUpdateInputsTranslatedSubnetFromConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	plan := newSocketSitePlanWithTranslatedSubnet(ctx, t, types.StringValue("10.24.150.0/24"))
	cfg := newSocketSitePlanWithTranslatedSubnet(ctx, t, types.StringNull())
	client := &catoclient.Service{}
	var diags diag.Diagnostics

	networkRangeInput := prepareNetworkRangeInput(ctx, client, cfg, plan, false, &diags)
	if diags.HasError() {
		t.Fatalf("prepareNetworkRangeInput: %v", diags)
	}
	if networkRangeInput.TranslatedSubnet != nil {
		t.Fatalf("expected translated subnet omitted when not in config, got %q", *networkRangeInput.TranslatedSubnet)
	}

	socketIfaceInput, _ := prepareSocketInterfaceInput(ctx, cfg, plan, false, &diags)
	if diags.HasError() {
		t.Fatalf("prepareSocketInterfaceInput: %v", diags)
	}
	if socketIfaceInput.Lan == nil {
		t.Fatal("expected LAN input")
	}
	if socketIfaceInput.Lan.TranslatedSubnet != nil {
		t.Fatalf("expected translated subnet omitted when not in config, got %q", *socketIfaceInput.Lan.TranslatedSubnet)
	}
}

func TestStripVrrpForAPI(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		destType cato_models.SocketInterfaceDestType
		want     cato_models.SocketInterfaceDestType
	}{
		"lag_master_and_vrrp_reduces_to_lag_master": {
			destType: cato_models.SocketInterfaceDestTypeLanLagMasterAndVrrp,
			want:     cato_models.SocketInterfaceDestTypeLanLagMaster,
		},
		"lag_master_unchanged": {
			destType: cato_models.SocketInterfaceDestTypeLanLagMaster,
			want:     cato_models.SocketInterfaceDestTypeLanLagMaster,
		},
		"lan_unchanged": {
			destType: cato_models.SocketInterfaceDestTypeLan,
			want:     cato_models.SocketInterfaceDestTypeLan,
		},
		"vrrp_and_lan_unchanged": { // prefix form, not an *_AND_VRRP suffix
			destType: cato_models.SocketInterfaceDestTypeVrrpAndLan,
			want:     cato_models.SocketInterfaceDestTypeVrrpAndLan,
		},
		"vrrp_unchanged": {
			destType: cato_models.SocketInterfaceDestTypeVrrp,
			want:     cato_models.SocketInterfaceDestTypeVrrp,
		},
		"future_non_lag_and_vrrp_variant_reduces_to_lan": {
			destType: cato_models.SocketInterfaceDestType("SOMETHING_AND_VRRP"),
			want:     cato_models.SocketInterfaceDestTypeLan,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := stripVrrpForAPI(tt.destType); got != tt.want {
				t.Fatalf("stripVrrpForAPI(%s): expected %s, got %s", tt.destType, tt.want, got)
			}
		})
	}
}

func TestPrepareSocketInterfaceInputVrrpDestType(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		destType     types.String
		lagMinLinks  types.Int64
		wantDestType cato_models.SocketInterfaceDestType
		wantLag      bool
	}{
		"strips_vrrp_and_keeps_lag": {
			destType:     types.StringValue(string(cato_models.SocketInterfaceDestTypeLanLagMasterAndVrrp)),
			lagMinLinks:  types.Int64Value(1),
			wantDestType: cato_models.SocketInterfaceDestTypeLanLagMaster,
			wantLag:      true,
		},
		"leaves_plain_lag_master": {
			destType:     types.StringValue(string(cato_models.SocketInterfaceDestTypeLanLagMaster)),
			lagMinLinks:  types.Int64Value(2),
			wantDestType: cato_models.SocketInterfaceDestTypeLanLagMaster,
			wantLag:      true,
		},
		"leaves_plain_lan": {
			destType:     types.StringValue(string(cato_models.SocketInterfaceDestTypeLan)),
			lagMinLinks:  types.Int64Null(),
			wantDestType: cato_models.SocketInterfaceDestTypeLan,
			wantLag:      false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			plan := newSocketSitePlanWithDestType(ctx, t, tt.destType, tt.lagMinLinks)
			cfg := plan
			var diags diag.Diagnostics

			// The *_AND_VRRP strip is unconditional for the API input (create AND update).
			socketIfaceInput, _ := prepareSocketInterfaceInput(ctx, cfg, plan, false, &diags)
			if diags.HasError() {
				t.Fatalf("prepareSocketInterfaceInput: %v", diags)
			}

			if socketIfaceInput.DestType != tt.wantDestType {
				t.Fatalf("expected dest type %s, got %s", tt.wantDestType, socketIfaceInput.DestType)
			}
			if tt.wantLag {
				if socketIfaceInput.Lag == nil {
					t.Fatal("expected LAG input to be preserved")
				}
				if socketIfaceInput.Lag.MinLinks != tt.lagMinLinks.ValueInt64() {
					t.Fatalf("expected lag min links %d, got %d", tt.lagMinLinks.ValueInt64(), socketIfaceInput.Lag.MinLinks)
				}
			} else if socketIfaceInput.Lag != nil {
				t.Fatalf("expected no LAG input, got %+v", socketIfaceInput.Lag)
			}
			if socketIfaceInput.Lan == nil {
				t.Fatal("expected LAN input")
			}
		})
	}
}

func TestParsedInterfaceDestType(t *testing.T) {
	t.Parallel()

	lagMasterAndVrrp := string(cato_models.SocketInterfaceDestTypeLanLagMasterAndVrrp)
	lagMaster := string(cato_models.SocketInterfaceDestTypeLanLagMaster)
	lan := string(cato_models.SocketInterfaceDestTypeLan)

	tests := map[string]struct {
		prior       types.String
		apiDestType *string
		isApplyPath bool
		want        types.String
	}{
		// Apply paths (Create/Update): the planned value wins whenever the user set one,
		// since the API was only ever sent the stripped base type.
		"apply_keeps_planned_vrrp_over_api_base": {
			prior:       types.StringValue(lagMasterAndVrrp),
			apiDestType: stringPtr(lagMaster),
			isApplyPath: true,
			want:        types.StringValue(lagMasterAndVrrp),
		},
		"apply_keeps_planned_value_identity": {
			prior:       types.StringValue(lan),
			apiDestType: stringPtr(lan),
			isApplyPath: true,
			want:        types.StringValue(lan),
		},
		"apply_unknown_plan_takes_api_value": {
			prior:       types.StringUnknown(),
			apiDestType: stringPtr(lagMaster),
			isApplyPath: true,
			want:        types.StringValue(lagMaster),
		},
		"apply_null_plan_takes_api_value": {
			prior:       types.StringNull(),
			apiDestType: stringPtr(lan),
			isApplyPath: true,
			want:        types.StringValue(lan),
		},
		// Read path: never downgrade *_AND_VRRP while the API still reports the base type
		// (single-socket drift); everything else stores the API value.
		"read_keeps_prior_vrrp_when_api_reports_base": {
			prior:       types.StringValue(lagMasterAndVrrp),
			apiDestType: stringPtr(lagMaster),
			isApplyPath: false,
			want:        types.StringValue(lagMasterAndVrrp),
		},
		"read_stores_api_promotion": {
			prior:       types.StringValue(lagMaster),
			apiDestType: stringPtr(lagMasterAndVrrp),
			isApplyPath: false,
			want:        types.StringValue(lagMasterAndVrrp),
		},
		"read_identity_keeps_value": {
			prior:       types.StringValue(lagMasterAndVrrp),
			apiDestType: stringPtr(lagMasterAndVrrp),
			isApplyPath: false,
			want:        types.StringValue(lagMasterAndVrrp),
		},
		"read_unrelated_api_value_wins": {
			prior:       types.StringValue(lagMasterAndVrrp),
			apiDestType: stringPtr(lan),
			isApplyPath: false,
			want:        types.StringValue(lan),
		},
		"read_no_prior_takes_api_value": {
			prior:       types.StringNull(),
			apiDestType: stringPtr(lagMaster),
			isApplyPath: false,
			want:        types.StringValue(lagMaster),
		},
		"read_api_nil_with_prior_stores_null": {
			prior:       types.StringValue(lagMasterAndVrrp),
			apiDestType: nil,
			isApplyPath: false,
			want:        types.StringNull(),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := parsedInterfaceDestType(tt.prior, tt.apiDestType, tt.isApplyPath)
			if !got.Equal(tt.want) {
				t.Fatalf("parsedInterfaceDestType(%v, %v, %v): expected %v, got %v",
					tt.prior, tt.apiDestType, tt.isApplyPath, tt.want, got)
			}
		})
	}
}

func newSocketSitePlanWithDestType(ctx context.Context, t *testing.T, destType types.String, lagMinLinks types.Int64) *SocketSite {
	t.Helper()

	nativeRange, diags := types.ObjectValueFrom(ctx, SiteNativeRangeResourceAttrTypes, NativeRange{
		NativeNetworkRange: types.StringValue("10.51.0.128/25"),
		LocalIP:            types.StringValue("10.51.0.1"),
		InterfaceDestType:  destType,
		LagMinLinks:        lagMinLinks,
		DhcpSettings:       types.ObjectNull(dhcp.SettingsAttrTypes),
	})
	if diags.HasError() {
		t.Fatalf("build native range: %v", diags)
	}

	return &SocketSite{
		Name:           types.StringValue("socket-site-01"),
		ConnectionType: types.StringValue("SOCKET_X1500"),
		SiteType:       types.StringValue("BRANCH"),
		NativeRange:    nativeRange,
		SiteLocation:   types.ObjectNull(tf.SiteLocationResourceAttrTypes),
	}
}

func newSocketSitePlanWithTranslatedSubnet(ctx context.Context, t *testing.T, translatedSubnet types.String) *SocketSite {
	t.Helper()

	nativeRange, diags := types.ObjectValueFrom(ctx, SiteNativeRangeResourceAttrTypes, NativeRange{
		NativeNetworkRange: types.StringValue("10.51.0.128/25"),
		LocalIP:            types.StringValue("10.51.0.1"),
		TranslatedSubnet:   translatedSubnet,
		DhcpSettings:       types.ObjectNull(dhcp.SettingsAttrTypes),
	})
	if diags.HasError() {
		t.Fatalf("build native range: %v", diags)
	}

	return &SocketSite{
		Name:           types.StringValue("aws-site-01"),
		ConnectionType: types.StringValue("SOCKET_AWS1500"),
		SiteType:       types.StringValue("DATACENTER"),
		NativeRange:    nativeRange,
		SiteLocation:   types.ObjectNull(tf.SiteLocationResourceAttrTypes),
	}
}

func assertTranslatedSubnetPointer(t *testing.T, got *string, wantNil bool, wantValue string) {
	t.Helper()

	if wantNil {
		if got != nil {
			t.Fatalf("expected translated subnet to be nil, got %q", *got)
		}
		return
	}

	if got == nil {
		t.Fatalf("expected translated subnet %q, got nil", wantValue)
		return
	}
	if *got != wantValue {
		t.Fatalf("expected translated subnet %q, got %q", wantValue, *got)
	}
}
