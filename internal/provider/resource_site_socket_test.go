package provider

import (
	"context"
	"testing"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

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
		r := &socketSiteResource{client: &catoClientData{catov2: sdkClient}}
		if got := r.getSocketSiteClient(); got != sdkClient {
			t.Fatalf("expected provider SDK client, got %T", got)
		}
	})
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

func TestNativeRangeTranslatedSubnetFromAPI(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		translatedSubnet *string
		nativeSubnet     string
		wantNull         bool
		wantValue        string
	}{
		"nil": {
			wantNull: true,
		},
		"empty": {
			translatedSubnet: stringPtr(""),
			nativeSubnet:     "10.0.0.0/24",
			wantNull:         true,
		},
		"equals_native": {
			translatedSubnet: stringPtr("10.0.0.0/24"),
			nativeSubnet:     "10.0.0.0/24",
			wantNull:         true,
		},
		"distinct": {
			translatedSubnet: stringPtr("192.168.20.0/24"),
			nativeSubnet:     "10.0.0.0/24",
			wantValue:        "192.168.20.0/24",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := nativeRangeTranslatedSubnetFromAPI(tt.translatedSubnet, tt.nativeSubnet)
			if tt.wantNull {
				if !got.IsNull() {
					t.Fatalf("expected null, got %q", got.ValueString())
				}
				return
			}
			if got.IsNull() || got.ValueString() != tt.wantValue {
				t.Fatalf("expected %q, got %v", tt.wantValue, got)
			}
		})
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
