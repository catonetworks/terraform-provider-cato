package privateapp

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/validators"
)

func (r *privateAppResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_private_app` resource contains the configuration parameters necessary to manage a private app.",

		Attributes: map[string]schema.Attribute{
			"allow_icmp_protocol": schema.BoolAttribute{
				Description: "Is ICMP enabled?",
				Required:    true,
			},
			"creation_time": schema.StringAttribute{
				Description: "Creation time",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Optional description of the private App",
				Optional:    true,
			},
			"id": schema.StringAttribute{
				Description: "The unique ID of the Private App",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"internal_app_address": schema.StringAttribute{
				Description: "The local address of the application, IPv4 address or FQDN",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The unique name of the private App",
				Required:    true,
			},
			"private_app_probing": r.schemaPrivateAppProbing(),
			"probing_enabled": schema.BoolAttribute{
				Description: "Is probing is enabled?",
				Required:    true,
			},
			"protocol_ports": r.schemaProtocolPorts(),
			"published": schema.BoolAttribute{
				Description: "Is the private app published?",
				Required:    true,
			},
			"published_app_domain": r.schemaPublishedAppDomain(),
		},
	}
}

func (r *privateAppResource) schemaPrivateAppProbing() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Private app probing settings",
		Optional:    true,
		Computed:    true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.UseStateForUnknown(),
		},
		Attributes: map[string]schema.Attribute{
			"fault_threshold_down": schema.Int64Attribute{
				Description: "Fault threshold",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "Probing ID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"interval": schema.Int64Attribute{
				Description: "Probing interval",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "Probing type",
				Required:    true,
			},
		},
	}
}

func (r *privateAppResource) schemaProtocolPorts() schema.SetNestedAttribute {
	return schema.SetNestedAttribute{
		Description: "List of ports and protocols",
		Optional:    true,
		Computed:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"ports": schema.SetAttribute{
					ElementType: types.Int64Type,
					Description: "List of TCP or UDP ports",
					Optional:    true,
				},
				"port_range": schema.SingleNestedAttribute{
					Description: "Port range",
					Optional:    true,
					Attributes: map[string]schema.Attribute{
						"from": schema.Int64Attribute{
							Description: "From",
							Required:    true,
						},
						"to": schema.Int64Attribute{
							Description: "To",
							Required:    true,
						},
					},
				},
				"protocol": schema.StringAttribute{
					Description: "Protocol; e.g.: TCP, UDP, ICMP",
					Required:    true,
					Validators:  []validator.String{validators.IPProtocolValidator{}},
				},
			},
		},
	}
}

func (r *privateAppResource) schemaPublishedAppDomain() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Domain information about the published private app. Optional if the private app is not published.",
		Optional:    true,
		Computed:    true,
		Attributes: map[string]schema.Attribute{
			// Note: "catoIP" is to be removed
			"connector_group_name": schema.StringAttribute{
				Description: "Connector group name",
				Required:    true,
			},
			"creation_time": schema.StringAttribute{
				Description: "Creation time",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"id": schema.StringAttribute{
				Description: "ID of the private app domain",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"published_app_domain": schema.StringAttribute{
				Description: "Published app domain",
				Required:    true,
			},
		},
	}
}
