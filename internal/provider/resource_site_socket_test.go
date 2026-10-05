package provider

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
	tf "github.com/catonetworks/terraform-provider-cato/internal/provider/tfmodel"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/validators"
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

	client := &catoClientData{AccountId: "account-123"}
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
		providerClient := newProviderSDKClient(sdkClient)
		r := &socketSiteResource{client: &catoClientData{catov2: providerClient}}
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
		client:           &catoClientData{AccountId: "account-123"},
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
		client:           &catoClientData{AccountId: "account-123"},
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

func TestSocketSiteValidateCloudRouterInput(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		plan      *tf.SocketSite
		wantError bool
	}{
		"omitted": {
			plan: &tf.SocketSite{
				ConnectionType: types.StringValue("SOCKET_AWS1500"),
				IsCloudRouter:  types.BoolNull(),
			},
		},
		"gcp_socket": {
			plan: &tf.SocketSite{
				ConnectionType: types.StringValue("SOCKET_GCP1500"),
				IsCloudRouter:  types.BoolValue(true),
			},
		},
		"non_gcp_socket": {
			plan: &tf.SocketSite{
				ConnectionType: types.StringValue("SOCKET_AWS1500"),
				IsCloudRouter:  types.BoolValue(true),
			},
			wantError: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var diags diag.Diagnostics
			(&socketSiteResource{}).validateCloudRouterInput(tt.plan, &diags)

			if gotError := diags.HasError(); gotError != tt.wantError {
				t.Fatalf("expected error=%t, got %t: %+v", tt.wantError, gotError, diags)
			}
		})
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
				client:           &catoClientData{AccountId: "account-123"},
				socketSiteClient: mockClient,
			}
			plan := &tf.SocketSite{IsCloudRouter: types.BoolValue(tt.value)}
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
		client:           &catoClientData{AccountId: "account-123"},
		socketSiteClient: mockClient,
	}
	plan := &tf.SocketSite{IsCloudRouter: types.BoolNull()}
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
		client:           &catoClientData{AccountId: "account-123"},
		socketSiteClient: mockClient,
	}
	plan := &tf.SocketSite{IsCloudRouter: types.BoolValue(true)}
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
		client:           &catoClientData{AccountId: "account-123"},
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
		client:           &catoClientData{AccountId: "account-123"},
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
		client:           &catoClientData{AccountId: "account-123"},
		socketSiteClient: mockClient,
	}
	var diags diag.Diagnostics

	r.configureCloudRouterAfterCreate(context.Background(), &tf.SocketSite{
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
	var sockets []tf.Socket
	diags = got.ElementsAs(ctx, &sockets, false)
	if diags.HasError() {
		t.Fatalf("decode sockets: %+v", diags)
	}
	if len(sockets) != 2 {
		t.Fatalf("expected two sockets, got %d", len(sockets))
	}
	socketBySerial := map[string]tf.Socket{}
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

func TestSocketSiteNativeRangeValidatorDescription(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	v := validators.GetNativeRangeValidator()

	if v.Description(ctx) == "" {
		t.Fatal("expected non-empty description")
	}
	if got, want := v.MarkdownDescription(ctx), v.Description(ctx); got != want {
		t.Fatalf("expected markdown description to match description\nwant: %q\ngot:  %q", want, got)
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

			got := stringPointerForOptionalInput(tt.value)
			if tt.want == nil {
				if got != nil {
					t.Fatalf("expected nil, got %q", *got)
				}
				return
			}

			if got == nil {
				t.Fatalf("expected %q, got nil", *tt.want)
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
			r := &socketSiteResource{client: &catoClientData{}}
			var diags diag.Diagnostics

			addInput := r.prepareSocketSiteInput(ctx, plan, &diags)
			if diags.HasError() {
				t.Fatalf("prepareSocketSiteInput: %v", diags)
			}
			assertTranslatedSubnetPointer(t, addInput.TranslatedSubnet, tt.wantNil, tt.wantValue)

			networkRangeInput := r.prepareNetworkRangeInput(ctx, cfg, plan, false, &diags)
			if diags.HasError() {
				t.Fatalf("prepareNetworkRangeInput: %v", diags)
			}
			assertTranslatedSubnetPointer(t, networkRangeInput.TranslatedSubnet, tt.wantNil, tt.wantValue)

			socketIfaceInput, _ := r.prepareSocketInterfaceInput(ctx, cfg, plan, false, &diags)
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
	r := &socketSiteResource{client: &catoClientData{}}
	var diags diag.Diagnostics

	networkRangeInput := r.prepareNetworkRangeInput(ctx, cfg, plan, false, &diags)
	if diags.HasError() {
		t.Fatalf("prepareNetworkRangeInput: %v", diags)
	}
	if networkRangeInput.TranslatedSubnet != nil {
		t.Fatalf("expected translated subnet omitted when not in config, got %q", *networkRangeInput.TranslatedSubnet)
	}

	socketIfaceInput, _ := r.prepareSocketInterfaceInput(ctx, cfg, plan, false, &diags)
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

func newSocketSitePlanWithTranslatedSubnet(ctx context.Context, t *testing.T, translatedSubnet types.String) *tf.SocketSite {
	t.Helper()

	nativeRange, diags := types.ObjectValueFrom(ctx, tf.SiteNativeRangeResourceAttrTypes, tf.NativeRange{
		NativeNetworkRange: types.StringValue("10.51.0.128/25"),
		LocalIP:            types.StringValue("10.51.0.1"),
		TranslatedSubnet:   translatedSubnet,
		DhcpSettings:       types.ObjectNull(tf.SiteNativeRangeDhcpResourceAttrTypes),
	})
	if diags.HasError() {
		t.Fatalf("build native range: %v", diags)
	}

	return &tf.SocketSite{
		Name:           types.StringValue("aws-site-01"),
		ConnectionType: types.StringValue("SOCKET_AWS1500"),
		SiteType:       types.StringValue("DATACENTER"),
		NativeRange:    nativeRange,
		SiteLocation:   types.ObjectNull(tf.SiteLocationResourceAttrTypes),
	}
}
