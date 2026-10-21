package socketsite

import (
	"context"
	"strings"
	"testing"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

func TestCloudRouterValidatorValidate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		connectionType types.String
		isCloudRouter  types.Bool
		wantError      bool
	}{
		"enabled for GCP vSocket": {
			connectionType: types.StringValue("SOCKET_GCP1500"),
			isCloudRouter:  types.BoolValue(true),
		},
		"disabled for GCP vSocket": {
			connectionType: types.StringValue("SOCKET_GCP1500"),
			isCloudRouter:  types.BoolValue(false),
		},
		"enabled for non-GCP vSocket": {
			connectionType: types.StringValue("SOCKET_AWS1500"),
			isCloudRouter:  types.BoolValue(true),
			wantError:      true,
		},
		"disabled for non-GCP vSocket": {
			connectionType: types.StringValue("SOCKET_AWS1500"),
			isCloudRouter:  types.BoolValue(false),
			wantError:      true,
		},
		"omitted": {
			connectionType: types.StringValue("SOCKET_AWS1500"),
			isCloudRouter:  types.BoolNull(),
		},
		"unknown connection type": {
			connectionType: types.StringUnknown(),
			isCloudRouter:  types.BoolValue(true),
		},
		"null connection type": {
			connectionType: types.StringNull(),
			isCloudRouter:  types.BoolValue(true),
		},
		"unknown Cloud Router value": {
			connectionType: types.StringValue("SOCKET_AWS1500"),
			isCloudRouter:  types.BoolUnknown(),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var diags diag.Diagnostics
			GetCloudRouterValidator().validate(tt.connectionType, tt.isCloudRouter, &diags)

			if gotError := diags.HasError(); gotError != tt.wantError {
				t.Fatalf("expected error=%t, got %t: %+v", tt.wantError, gotError, diags)
			}
			if tt.wantError && !strings.Contains(diags[0].Detail(), "GCP vSocket") {
				t.Fatalf("expected GCP vSocket validation error, got %q", diags[0].Detail())
			}
		})
	}
}

func TestCloudRouterValidatorValidateBoolReadsConnectionTypeFromConfig(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		connectionType string
		isCloudRouter  bool
		wantError      bool
	}{
		"rejects AWS vSocket": {
			connectionType: "SOCKET_AWS1500",
			isCloudRouter:  true,
			wantError:      true,
		},
		"accepts GCP vSocket": {
			connectionType: "SOCKET_GCP1500",
			isCloudRouter:  true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			config := tfsdk.Config{
				Schema: schema.Schema{
					Attributes: map[string]schema.Attribute{
						"connection_type": schema.StringAttribute{Required: true},
						"is_cloud_router": schema.BoolAttribute{Optional: true},
					},
				},
				Raw: tftypes.NewValue(
					tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"connection_type": tftypes.String,
						"is_cloud_router": tftypes.Bool,
					}},
					map[string]tftypes.Value{
						"connection_type": tftypes.NewValue(tftypes.String, tt.connectionType),
						"is_cloud_router": tftypes.NewValue(tftypes.Bool, tt.isCloudRouter),
					},
				),
			}
			resp := &validator.BoolResponse{}

			GetCloudRouterValidator().ValidateBool(context.Background(), validator.BoolRequest{
				Path:        path.Root("is_cloud_router"),
				Config:      config,
				ConfigValue: types.BoolValue(tt.isCloudRouter),
			}, resp)

			if gotError := resp.Diagnostics.HasError(); gotError != tt.wantError {
				t.Fatalf("expected error=%t, got %t: %+v", tt.wantError, gotError, resp.Diagnostics)
			}
		})
	}
}

func TestSocketSiteValidateCloudRouterInput(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		plan      *SocketSite
		wantError bool
	}{
		"omitted": {
			plan: &SocketSite{
				ConnectionType: types.StringValue("SOCKET_AWS1500"),
				IsCloudRouter:  types.BoolNull(),
			},
		},
		"gcp_socket": {
			plan: &SocketSite{
				ConnectionType: types.StringValue("SOCKET_GCP1500"),
				IsCloudRouter:  types.BoolValue(true),
			},
		},
		"non_gcp_socket": {
			plan: &SocketSite{
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

func TestSocketSiteNativeRangeValidatorDescription(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	v := GetNativeRangeValidator()

	if v.Description(ctx) == "" {
		t.Fatal("expected non-empty description")
	}
	if got, want := v.MarkdownDescription(ctx), v.Description(ctx); got != want {
		t.Fatalf("expected markdown description to match description\nwant: %q\ngot:  %q", want, got)
	}
}

func TestSocketSiteConnectionTypeValidation(t *testing.T) {
	t.Parallel()
	v := SiteConnectionTypeValidator{}
	for _, tc := range []struct {
		name      string
		value     types.String
		wantError bool
	}{
		{"valid", types.StringValue(string(cato_models.AllSiteConnectionTypeEnum[0])), false},
		{"invalid", types.StringValue("INVALID"), true},
		{"null", types.StringNull(), true},
		{"unknown", types.StringUnknown(), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp := &validator.StringResponse{}
			v.ValidateString(context.Background(), validator.StringRequest{Path: path.Root("value"), ConfigValue: tc.value}, resp)
			require.Equal(t, tc.wantError, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
		})
	}
}

func TestSocketSiteInterfaceDestTypeValidation(t *testing.T) {
	t.Parallel()
	v := SocketInterfaceDestTypeValidator{}
	for _, tc := range []struct {
		name      string
		value     types.String
		wantError bool
	}{
		{"valid", types.StringValue(string(cato_models.AllSocketInterfaceDestType[0])), false},
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
