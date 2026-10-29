package wifissid

import (
	"context"

	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/planmodifiers"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

func (r *ssidResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Description: "Manages a Wi-Fi SSID on an unused Wi-Fi interface. " +
		"Site and interface changes require replacement. " +
		"PSKs remain sensitive configuration in Terraform state; the API cannot return them or detect external PSK-only drift. " +
		"Import by SSID ID and supply a PSK before applying changes to a PSK SSID.", Attributes: schemaAttributes()}
}
func schemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"band": schema.StringAttribute{
			Description: "Band",
			Required:    true,
			Validators:  []validator.String{stringvalidator.OneOf(utils.ToStrings(m.AllWifiRadioBand)...)},
		},
		"category": schema.StringAttribute{
			Description: "SSID category. The API currently supports GUEST only.",
			Required:    true,
			Validators:  []validator.String{stringvalidator.OneOf(utils.ToStrings(m.AllWifiSsidCategory)...), stringvalidator.OneOf("GUEST")},
		},
		"dhcp": schema.SingleNestedAttribute{
			Description: "Dhcp",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"dhcp_subnet": schema.StringAttribute{
					Description: "Dhcp subnet",
					Optional:    true, Computed: true,
					PlanModifiers: []planmodifier.String{planmodifiers.NullStringModifier()},
				},
			},
		},
		"dhcp_enabled": schema.BoolAttribute{
			Description: "Deprecated by the API; only false is supported. Configure DHCP through dhcp.",
			Optional:    true, Computed: true,
			Default: booldefault.StaticBool(false),
		},
		"enabled": schema.BoolAttribute{
			Description: "Enabled",
			Optional:    true, Computed: true,
			Default: booldefault.StaticBool(true),
		},
		"id": schema.StringAttribute{
			Description:   "Resource ID",
			Computed:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"internet_only": schema.BoolAttribute{
			Description: "Internet only",
			Optional:    true, Computed: true,
			Default: booldefault.StaticBool(false),
		},
		"local_ip": schema.StringAttribute{
			Description: "Local ip",
			Required:    true,
		},
		"mdns_enabled": schema.BoolAttribute{
			Description: "Mdns enabled",
			Required:    true,
		},
		"microsegmentation_enabled": schema.BoolAttribute{
			Description: "Microsegmentation enabled",
			Required:    true,
		},
		"name": schema.StringAttribute{
			Description: "Name",
			Required:    true,
		},
		"network_interface_id": schema.StringAttribute{
			Description:   "Network interface id",
			Required:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
		},
		"security": schemaSecurity(),
		"site": schema.SingleNestedAttribute{
			Description:   "Site",
			Required:      true,
			Attributes:    idname.SchemaNameID("Site"),
			PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier(), objectplanmodifier.RequiresReplace()},
		},
		"subnet": schema.StringAttribute{
			Description: "Subnet",
			Required:    true,
		},
		"visible": schema.BoolAttribute{
			Description: "Visible",
			Optional:    true, Computed: true,
			Default: booldefault.StaticBool(true),
		},
	}
}

var pskAttrTypes = map[string]attr.Type{"passkey": types.StringType}
var dhcpAttrTypes = map[string]attr.Type{"dhcp_subnet": types.StringType}
var securityAttrTypes = map[string]attr.Type{
	"mode": types.StringType, "auth_protocol": types.StringType, "track_authentication": types.BoolType,
	"psk": types.ObjectType{AttrTypes: pskAttrTypes},
}

func schemaSecurity() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Security",
		Required:    true,
		Attributes: map[string]schema.Attribute{
			"auth_protocol": schema.StringAttribute{
				Description: "Auth protocol",
				Required:    true,
				Validators:  []validator.String{stringvalidator.OneOf(utils.ToStrings(m.AllWifiAuthProtocol)...)},
			},
			"mode": schema.StringAttribute{
				Description: "Mode",
				Required:    true,
				Validators:  []validator.String{stringvalidator.OneOf(utils.ToStrings(m.AllWifiSecurityMode)...)},
			},
			"psk": schema.SingleNestedAttribute{
				Description: "Configured PSK retained because the API cannot return it. " +
					"Import cannot recover it; external PSK-only drift is undetectable.",
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"passkey": schema.StringAttribute{
						Description: "Passkey",
						Required:    true,
						Sensitive:   true,
					},
				},
			},
			"track_authentication": schema.BoolAttribute{
				Description: "Track authentication",
				Optional:    true, Computed: true,
				Default: booldefault.StaticBool(false),
			},
		},
	}
}
