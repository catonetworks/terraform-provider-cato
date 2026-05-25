package provider

import (
	"context"
	"strings"
	"testing"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/mock"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/mocks"
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
		r := &socketSiteResource{client: &catoClientData{catov2: sdkClient}}
		if got := r.getSocketSiteClient(); got != sdkClient {
			t.Fatalf("expected provider SDK client, got %T", got)
		}
	})
}

func TestSocketSiteCreateTranslatedSubnetPayload(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		translatedSubnet types.String
		want             *string
	}{
		"not_configured": {
			translatedSubnet: types.StringNull(),
			want:             nil,
		},
		"empty": {
			translatedSubnet: types.StringValue(""),
			want:             nil,
		},
		"configured": {
			translatedSubnet: types.StringValue("192.168.20.0/24"),
			want:             stringPtr("192.168.20.0/24"),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			mockClient := mocks.NewSocketSiteClient(t)
			mockClient.EXPECT().
				SiteAddSocketSite(mock.Anything, mock.MatchedBy(func(input cato_models.AddSocketSiteInput) bool {
					assertSocketSiteAddInput(t, input, tt.want)
					return true
				}), "account-123").
				Return(nil, assertErr("add failed")).
				Once()

			r := &socketSiteResource{
				client:           &catoClientData{AccountId: "account-123"},
				socketSiteClient: mockClient,
			}
			req := resource.CreateRequest{Plan: newSocketSitePlan(ctx, t, tt.translatedSubnet)}
			resp := &resource.CreateResponse{State: tfsdk.State{Schema: getSocketSiteSchema(ctx, t)}}

			r.Create(ctx, req, resp)

			if !resp.Diagnostics.HasError() {
				t.Fatal("expected diagnostics for add socket site error")
			}
		})
	}
}

func TestSocketSiteCreateValidatesNativeRangeBeforeAPI(t *testing.T) {
	t.Parallel()

	tests := map[string]socketSitePlanOptions{
		"invalid_local_ip": {
			LocalIP: "not-an-ip",
			Subnet:  "10.51.0.128/25",
		},
		"invalid_native_range": {
			LocalIP: "10.51.0.129",
			Subnet:  "not-a-cidr",
		},
		"local_ip_outside_native_range": {
			LocalIP: "10.52.0.129",
			Subnet:  "10.51.0.128/25",
		},
	}

	for name, options := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			mockClient := mocks.NewSocketSiteClient(t)
			r := &socketSiteResource{
				client:           &catoClientData{AccountId: "account-123"},
				socketSiteClient: mockClient,
			}
			req := resource.CreateRequest{
				Plan: newSocketSitePlanWithOptions(ctx, t, options),
			}
			resp := &resource.CreateResponse{State: tfsdk.State{Schema: getSocketSiteSchema(ctx, t)}}

			r.Create(ctx, req, resp)

			if !resp.Diagnostics.HasError() {
				t.Fatal("expected diagnostics before socket site API call")
			}
		})
	}
}

func TestCalculateLocalIP(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		subnet   string
		connType string
		want     string
	}{
		"empty_subnet": {
			subnet:   "",
			connType: "SOCKET_X1500",
			want:     "",
		},
		"invalid_subnet": {
			subnet:   "not-a-cidr",
			connType: "SOCKET_X1500",
			want:     "",
		},
		"x1500_uses_first_available_ip": {
			subnet:   "192.168.10.0/24",
			connType: "SOCKET_X1500",
			want:     "192.168.10.1",
		},
		"aws1500_uses_fourth_offset": {
			subnet:   "192.168.10.0/24",
			connType: "SOCKET_AWS1500",
			want:     "192.168.10.4",
		},
		"gcp1500_uses_fourth_offset": {
			subnet:   "192.168.10.0/24",
			connType: "SOCKET_GCP1500",
			want:     "192.168.10.4",
		},
		"ipv6_subnet_is_ignored": {
			subnet:   "2001:db8::/64",
			connType: "SOCKET_X1500",
			want:     "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := calculateLocalIP(context.Background(), tt.subnet, tt.connType)
			if got != tt.want {
				t.Fatalf("expected local IP %q, got %q", tt.want, got)
			}
		})
	}
}

func TestHydrateOptionalLocationString(t *testing.T) {
	t.Parallel()

	empty := ""
	singleSpace := " "
	value := "Berlin"

	tests := []struct {
		name       string
		apiValue   *string
		priorValue types.String
		wantNull   bool
		want       string
	}{
		{
			name:       "api value wins",
			apiValue:   &value,
			priorValue: types.StringValue("Old"),
			want:       "Berlin",
		},
		{
			name:       "preserve explicit empty from prior state",
			apiValue:   nil,
			priorValue: types.StringValue(""),
			want:       "",
		},
		{
			name:       "api empty preserves explicit empty from prior state",
			apiValue:   &empty,
			priorValue: types.StringValue(""),
			want:       "",
		},
		{
			name:       "api whitespace preserves explicit empty from prior state",
			apiValue:   &singleSpace,
			priorValue: types.StringValue(""),
			want:       "",
		},
		{
			name:       "null when api and prior are unset",
			apiValue:   nil,
			priorValue: types.StringNull(),
			wantNull:   true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := hydrateOptionalLocationString(tt.apiValue, tt.priorValue)
			if tt.wantNull {
				if !got.IsNull() {
					t.Fatalf("expected null, got %q", got.ValueString())
				}
				return
			}

			if got.IsNull() || got.IsUnknown() {
				t.Fatalf("expected value %q, got null/unknown", tt.want)
			}
			if got.ValueString() != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got.ValueString())
			}
		})
	}
}

func TestSocketSiteNativeRangeValidator(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		options       socketSitePlanOptions
		wantErrors    bool
		errorContains string
	}{
		"valid_x1500": {
			options: socketSitePlanOptions{
				ConnectionType: "SOCKET_X1500",
				Interface:      "LAN1",
				LocalIP:        "10.51.0.129",
				Subnet:         "10.51.0.128/25",
			},
			wantErrors: false,
		},
		"invalid_interface_for_aws": {
			options: socketSitePlanOptions{
				ConnectionType: "SOCKET_AWS1500",
				Interface:      "LAN1",
				LocalIP:        "10.51.0.129",
				Subnet:         "10.51.0.128/25",
			},
			wantErrors:    true,
			errorContains: "interface_index can only be specified",
		},
		"invalid_local_ip": {
			options: socketSitePlanOptions{
				ConnectionType: "SOCKET_X1500",
				Interface:      "LAN1",
				LocalIP:        "not-an-ip",
				Subnet:         "10.51.0.128/25",
			},
			wantErrors:    true,
			errorContains: "not a valid IP address",
		},
		"invalid_native_range": {
			options: socketSitePlanOptions{
				ConnectionType: "SOCKET_X1500",
				Interface:      "LAN1",
				LocalIP:        "10.51.0.129",
				Subnet:         "not-a-cidr",
			},
			wantErrors:    true,
			errorContains: "not a valid CIDR notation",
		},
		"local_ip_outside_native_range": {
			options: socketSitePlanOptions{
				ConnectionType: "SOCKET_X1500",
				Interface:      "LAN1",
				LocalIP:        "10.52.0.129",
				Subnet:         "10.51.0.128/25",
			},
			wantErrors:    true,
			errorContains: "is not within native_network_range",
		},
		"unknown_interface_skips_validation": {
			options: socketSitePlanOptions{
				ConnectionType: "SOCKET_AWS1500",
				Interface:      unknownString,
				LocalIP:        "not-an-ip",
				Subnet:         "not-a-cidr",
			},
			wantErrors: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			resp := &validator.ObjectResponse{}
			nativeRange := newSocketSiteNativeRangeWithOptions(tt.options)
			socketSiteNativeRangeValidator{}.ValidateObject(ctx, validator.ObjectRequest{
				Path:        path.Root("native_range"),
				Config:      newSocketSiteConfig(ctx, t, tt.options),
				ConfigValue: nativeRange,
			}, resp)

			if tt.wantErrors && !resp.Diagnostics.HasError() {
				t.Fatal("expected validation diagnostics")
			}
			if !tt.wantErrors && resp.Diagnostics.HasError() {
				t.Fatalf("unexpected diagnostics: %+v", resp.Diagnostics)
			}
			if tt.errorContains != "" {
				assertDiagnosticsContain(t, resp.Diagnostics.Errors(), tt.errorContains)
			}
		})
	}
}

func TestSocketSiteNativeRangeValidatorDescription(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	v := socketSiteNativeRangeValidator{}

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

func getSocketSiteSchema(ctx context.Context, t *testing.T) schema.Schema {
	t.Helper()

	r := &socketSiteResource{}
	resp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, resp)

	return resp.Schema
}

type stringSentinel string

const (
	nullString    stringSentinel = "<null>"
	unknownString stringSentinel = "<unknown>"
)

type socketSitePlanOptions struct {
	ConnectionType   string
	Interface        stringSentinel
	LocalIP          string
	Subnet           string
	TranslatedSubnet types.String
}

func newSocketSitePlan(ctx context.Context, t *testing.T, translatedSubnet types.String) tfsdk.Plan {
	t.Helper()

	return newSocketSitePlanWithOptions(ctx, t, socketSitePlanOptions{
		TranslatedSubnet: translatedSubnet,
	})
}

func newSocketSitePlanWithOptions(ctx context.Context, t *testing.T, options socketSitePlanOptions) tfsdk.Plan {
	t.Helper()

	plan := tfsdk.Plan{Schema: getSocketSiteSchema(ctx, t)}
	diags := plan.Set(ctx, SocketSite{
		ID:             types.StringNull(),
		Name:           types.StringValue("aws-site-01"),
		ConnectionType: types.StringValue(valueOrDefault(options.ConnectionType, "SOCKET_AWS1500")),
		SiteType:       types.StringValue("DATACENTER"),
		Description:    types.StringValue("AWS Virtual Socket Site"),
		NativeRange:    newSocketSiteNativeRangeWithOptions(options),
		SiteLocation:   newSocketSiteLocation(),
	})
	if diags.HasError() {
		t.Fatalf("unexpected plan diagnostics: %+v", diags)
	}

	return plan
}

func newSocketSiteNativeRangeWithOptions(options socketSitePlanOptions) types.Object {
	translatedSubnet := options.TranslatedSubnet
	if translatedSubnet.IsNull() && !translatedSubnet.IsUnknown() && translatedSubnet.ValueString() == "" {
		translatedSubnet = types.StringNull()
	}

	return types.ObjectValueMust(SiteNativeRangeResourceAttrTypes, map[string]attr.Value{
		"interface_index":                 stringValueFromSentinel(options.Interface, "LAN1"),
		"interface_id":                    types.StringNull(),
		"interface_name":                  types.StringNull(),
		"native_network_lan_interface_id": types.StringNull(),
		"native_network_range":            types.StringValue(valueOrDefault(options.Subnet, "10.51.0.128/25")),
		"native_network_range_id":         types.StringNull(),
		"range_name":                      types.StringNull(),
		"range_id":                        types.StringNull(),
		"local_ip":                        types.StringValue(valueOrDefault(options.LocalIP, "10.51.0.129")),
		"translated_subnet":               translatedSubnet,
		"gateway":                         types.StringNull(),
		"range_type":                      types.StringNull(),
		"dhcp_settings":                   types.ObjectNull(SiteNativeRangeDhcpResourceAttrTypes),
		"vlan":                            types.Int64Null(),
		"mdns_reflector":                  types.BoolValue(false),
		"lag_min_links":                   types.Int64Null(),
		"interface_dest_type":             types.StringValue("LAN"),
	})
}

func newSocketSiteConfig(ctx context.Context, t *testing.T, options socketSitePlanOptions) tfsdk.Config {
	t.Helper()

	plan := newSocketSitePlanWithOptions(ctx, t, options)

	return tfsdk.Config(plan)
}

func newSocketSiteLocation() types.Object {
	return types.ObjectValueMust(SiteLocationResourceAttrTypes, map[string]attr.Value{
		"country_code": types.StringValue("US"),
		"state_code":   types.StringValue("US-NY"),
		"timezone":     types.StringValue("America/New_York"),
		"address":      types.StringNull(),
		"city":         types.StringValue("New York City"),
	})
}

func assertSocketSiteAddInput(t *testing.T, input cato_models.AddSocketSiteInput, translatedSubnet *string) {
	t.Helper()

	if input.Name != "aws-site-01" {
		t.Fatalf("expected site name aws-site-01, got %q", input.Name)
	}
	if input.ConnectionType != cato_models.SiteConnectionTypeEnum("SOCKET_AWS1500") {
		t.Fatalf("expected connection type SOCKET_AWS1500, got %q", input.ConnectionType)
	}
	if input.SiteType != cato_models.SiteType("DATACENTER") {
		t.Fatalf("expected site type DATACENTER, got %q", input.SiteType)
	}
	if input.NativeNetworkRange != "10.51.0.128/25" {
		t.Fatalf("expected native network range 10.51.0.128/25, got %q", input.NativeNetworkRange)
	}
	if input.SiteLocation == nil {
		t.Fatal("expected site location input")
	}
	if input.SiteLocation.City == nil || *input.SiteLocation.City != "New York City" {
		t.Fatalf("expected site city New York City, got %v", input.SiteLocation.City)
	}

	if translatedSubnet == nil {
		if input.TranslatedSubnet != nil {
			t.Fatalf("expected translated subnet to be omitted, got %q", *input.TranslatedSubnet)
		}
		return
	}

	if input.TranslatedSubnet == nil {
		t.Fatalf("expected translated subnet %q, got nil", *translatedSubnet)
	}
	if *input.TranslatedSubnet != *translatedSubnet {
		t.Fatalf("expected translated subnet %q, got %q", *translatedSubnet, *input.TranslatedSubnet)
	}
}

func valueOrDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func stringValueFromSentinel(value stringSentinel, fallback string) types.String {
	switch value {
	case "":
		return types.StringValue(fallback)
	case nullString:
		return types.StringNull()
	case unknownString:
		return types.StringUnknown()
	default:
		return types.StringValue(string(value))
	}
}

func assertDiagnosticsContain(t *testing.T, errors []diag.Diagnostic, want string) {
	t.Helper()

	for _, diagnostic := range errors {
		if strings.Contains(diagnostic.Summary(), want) || strings.Contains(diagnostic.Detail(), want) {
			return
		}
	}

	t.Fatalf("expected diagnostics to contain %q, got %+v", want, errors)
}
