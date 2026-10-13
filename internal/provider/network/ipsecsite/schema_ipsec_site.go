package ipsecsite

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *siteIpsecResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier for Ipsec Site",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Ipsec Site Name",
				Required:    true,
			},
			"site_type": schema.StringAttribute{
				Description: "Valid values are: BRANCH, HEADQUARTERS, CLOUD_DC, and DATACENTER.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "description",
				Required:    true,
			},
			"native_network_range": schema.StringAttribute{
				Description: "NativeNetworkRange",
				Required:    true,
			},
			"native_network_range_id": schema.StringAttribute{
				Description: "Site native IP range ID (for update purpose)",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"interface_id": schema.StringAttribute{
				Description: "IPSec interface ID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"site_location": siteLocationAttribute(),
			"ipsec":         ipsecAttribute(),
		},
	}
}

func siteLocationAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "SiteLocation",
		Required:    true,
		Attributes: map[string]schema.Attribute{
			"country_code": schema.StringAttribute{
				Description: "Country Code",
				Required:    true,
			},
			"state_code": schema.StringAttribute{
				Description: "State Code",
				Optional:    true,
			},
			"timezone": schema.StringAttribute{
				Description: "Timezone",
				Required:    true,
			},
			"address": schema.StringAttribute{
				Description: "Address",
				Optional:    true,
			},
			"city": schema.StringAttribute{
				Description: "City",
				Optional:    true,
			},
		},
	}
}

func ipsecAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "IPSec Configuration",
		Required:    true,
		Attributes: map[string]schema.Attribute{
			"site_id": schema.StringAttribute{
				Description: "Site Identifier for Ipsec Site",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"primary":             ipsecEndpointAttribute("primary", true),
			"secondary":           ipsecEndpointAttribute("secondary", false),
			"connection_mode":     connectionModeAttribute(),
			"identification_type": identificationTypeAttribute(),
			"init_message":        ipsecMessageAttribute("IKE initialization message configuration"),
			"auth_message":        ipsecMessageAttribute("IKE authentication message configuration"),
			"network_ranges": schema.ListAttribute{
				Description: "List of network ranges (e.g., ['servers:192.168.11.0/24', 'desktops:192.169.11.0/24'])",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func ipsecEndpointAttribute(description string, required bool) schema.SingleNestedAttribute {
	attribute := schema.SingleNestedAttribute{
		Description: description,
		Attributes: map[string]schema.Attribute{
			"destination_type": schema.StringAttribute{
				Description: "destinationtype",
				Optional:    true,
			},
			"public_cato_ip_id": schema.StringAttribute{
				Description: "publiccatoipid",
				Optional:    true,
			},
			"pop_location_id": schema.StringAttribute{
				Description: "poplocationid",
				Optional:    true,
			},
			"tunnels": ipsecTunnelsAttribute(),
		},
	}

	if required {
		attribute.Required = true
	} else {
		attribute.Optional = true
	}

	return attribute
}

func ipsecTunnelsAttribute() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		Description: "tunnels",
		Optional:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"tunnel_id": schema.StringAttribute{
					Description: "tunnel ID",
					Computed:    true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"public_site_ip": schema.StringAttribute{
					Description: "publicsiteip",
					Optional:    true,
				},
				"private_cato_ip": schema.StringAttribute{
					Description: "privatecatoip",
					Optional:    true,
				},
				"private_site_ip": schema.StringAttribute{
					Description: "privatesiteip",
					Optional:    true,
				},
				"psk": schema.StringAttribute{
					Description: "psk",
					Required:    true,
					Sensitive:   true,
				},
				"last_mile_bw": lastMileBandwidthAttribute(),
			},
		},
	}
}

func lastMileBandwidthAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "lastmilebw",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"downstream": schema.Int64Attribute{
				Description: "Downstream",
				Required:    true,
			},
			"upstream": schema.Int64Attribute{
				Description: "upstream",
				Required:    true,
			},
			"downstream_mbps_precision": schema.Float64Attribute{
				Description: "downstreamMbpsPrecision",
				Optional:    true,
			},
			"upstream_mbps_precision": schema.Float64Attribute{
				Description: "upstreamMbpsPrecision",
				Optional:    true,
			},
		},
	}
}

func connectionModeAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		Description: "Connection mode for IPSec tunnel. Valid values: RESPONDER_ONLY, BIDIRECTIONAL",
		Optional:    true,
		Validators: []validator.String{
			stringvalidator.OneOf("RESPONDER_ONLY", "BIDIRECTIONAL"),
		},
	}
}

func identificationTypeAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		Description: "Identification type for IPSec. Only applicable when connection_mode is " +
			"RESPONDER_ONLY. Valid values: IPV4, FQDN, EMAIL, KEY_ID",
		Optional: true,
		Validators: []validator.String{
			stringvalidator.OneOf("IPV4", "FQDN", "EMAIL", "KEY_ID"),
		},
		PlanModifiers: []planmodifier.String{IdentificationTypeValidator()},
	}
}

func ipsecMessageAttribute(description string) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: description,
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"cipher": schema.StringAttribute{
				Description: "Cipher algorithm. Valid values: NONE, AUTOMATIC, AES_CBC_128, AES_CBC_256, AES_GCM_128, AES_GCM_256, DES3_CBC",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("NONE", "AUTOMATIC", "AES_CBC_128", "AES_CBC_256", "AES_GCM_128", "AES_GCM_256", "DES3_CBC"),
				},
				Default: stringdefault.StaticString("AUTOMATIC"),
			},
			"dh_group": schema.StringAttribute{
				Description: "Diffie-Hellman group. Valid values: NONE, AUTOMATIC, DH_2_MODP1024, " +
					"DH_5_MODP1536, DH_14_MODP2048, DH_15_MODP3072, DH_16_MODP4096, DH_19_ECP256, " +
					"DH_20_ECP384, DH_21_ECP521",
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"NONE",
						"AUTOMATIC",
						"DH_2_MODP1024",
						"DH_5_MODP1536",
						"DH_14_MODP2048",
						"DH_15_MODP3072",
						"DH_16_MODP4096",
						"DH_19_ECP256",
						"DH_20_ECP384",
						"DH_21_ECP521",
					),
				},
				Default: stringdefault.StaticString("AUTOMATIC"),
			},
			"integrity": schema.StringAttribute{
				Description: "Integrity algorithm. Valid values: NONE, AUTOMATIC, MD5, SHA1, SHA256, SHA384, SHA512",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("NONE", "AUTOMATIC", "MD5", "SHA1", "SHA256", "SHA384", "SHA512"),
				},
				Default: stringdefault.StaticString("AUTOMATIC"),
			},
			"prf": schema.StringAttribute{
				Description: "Pseudo-Random Function. Valid values: NONE, AUTOMATIC, MD5, SHA1, SHA256, SHA384, SHA512",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("NONE", "AUTOMATIC", "MD5", "SHA1", "SHA256", "SHA384", "SHA512"),
				},
				Default: stringdefault.StaticString("AUTOMATIC"),
			},
		},
	}
}
