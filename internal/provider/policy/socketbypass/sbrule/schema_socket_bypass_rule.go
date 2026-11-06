package sbrule

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func refSetAttribute(description string) schema.SetNestedAttribute {
	return schema.SetNestedAttribute{
		Description: description,
		Optional:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{
					Optional: true,
					Computed: true,
				},
				"name": schema.StringAttribute{
					Optional: true,
					Computed: true,
				},
			},
		},
	}
}

func fromToListAttribute(description string) schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		Description: description,
		Optional:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"from": schema.StringAttribute{Required: true},
				"to":   schema.StringAttribute{Required: true},
			},
		},
	}
}

func siteScopeAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Sites and site groups the rule applies to",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"site":  refSetAttribute("Site references"),
			"group": refSetAttribute("Site group references"),
		},
	}
}

func sourceAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Source traffic matching criteria",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"floating_subnet":     refSetAttribute("Floating subnet references"),
			"global_ip_range":     refSetAttribute("Global IP range references"),
			"group":               refSetAttribute("Group references"),
			"host":                refSetAttribute("Host references"),
			"ip":                  schema.ListAttribute{Optional: true, ElementType: types.StringType},
			"ip_range":            fromToListAttribute("IP ranges"),
			"network_interface":   refSetAttribute("Network interface references"),
			"site_network_subnet": refSetAttribute("Site network subnet references"),
			"subnet":              schema.ListAttribute{Optional: true, ElementType: types.StringType},
			"vlan":                schema.ListAttribute{Optional: true, ElementType: types.Int64Type},
		},
	}
}

func destinationAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Destination traffic matching criteria",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"application":     refSetAttribute("Application references"),
			"custom_app":      refSetAttribute("Custom application references"),
			"domain":          schema.ListAttribute{Optional: true, ElementType: types.StringType},
			"fqdn":            schema.ListAttribute{Optional: true, ElementType: types.StringType},
			"global_ip_range": refSetAttribute("Global IP range references"),
			"ip":              schema.ListAttribute{Optional: true, ElementType: types.StringType},
			"ip_range":        fromToListAttribute("IP ranges"),
			"subnet":          schema.ListAttribute{Optional: true, ElementType: types.StringType},
		},
	}
}

func serviceAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Services the rule applies to",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"simple": schema.ListAttribute{
				Description: "Simple service names",
				Optional:    true,
				ElementType: types.StringType,
			},
			"custom": schema.ListNestedAttribute{
				Description: "Custom services",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"port": schema.ListAttribute{
							Description: "Ports",
							Optional:    true,
							ElementType: types.StringType,
						},
						"port_range": schema.SingleNestedAttribute{
							Description: "Port range",
							Optional:    true,
							Attributes: map[string]schema.Attribute{
								"from": schema.StringAttribute{Required: true},
								"to":   schema.StringAttribute{Required: true},
							},
						},
						"protocol": schema.StringAttribute{
							Description: "IP protocol",
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func (r *socketBypassRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a rule in the Cato socket bypass policy. " +
			"Underlying GraphQL is marked @beta; behavior and fields may change.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Rule ID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"at": schema.SingleNestedAttribute{
				Description: "Where to insert the rule",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"position": schema.StringAttribute{
						Description: "Position relative to policy, section, or another rule",
						Required:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.String{
							stringvalidator.OneOf(
								"AFTER_RULE", "BEFORE_RULE", "FIRST_IN_POLICY", "FIRST_IN_SECTION",
								"LAST_IN_POLICY", "LAST_IN_SECTION",
							),
						},
					},
					"ref": schema.StringAttribute{
						Description: "Reference rule or section ID when required by position",
						Optional:    true,
					},
				},
			},
			"rule": schema.SingleNestedAttribute{
				Description: "Rule definition",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Description: "Rule ID",
						Computed:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"name": schema.StringAttribute{
						Description: "Rule name",
						Required:    true,
					},
					"description": schema.StringAttribute{
						Description: "Rule description",
						Optional:    true,
					},
					"enabled": schema.BoolAttribute{
						Description: "Whether the rule is enabled",
						Required:    true,
					},
					"action": schema.SingleNestedAttribute{
						Description: "Bypass action",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"port": schema.StringAttribute{
								Description: "Egress port (AUTOMATIC or WAN1-WAN4)",
								Optional:    true,
								Validators: []validator.String{
									stringvalidator.OneOf("AUTOMATIC", "WAN1", "WAN2", "WAN3", "WAN4"),
								},
							},
							"event_enabled": schema.BoolAttribute{
								Description: "Whether event tracking is enabled",
								Optional:    true,
								Computed:    true,
								Default:     booldefault.StaticBool(false),
							},
						},
					},
					"site":        siteScopeAttribute(),
					"source":      sourceAttribute(),
					"destination": destinationAttribute(),
					"service":     serviceAttribute(),
					"exception": schema.ListNestedAttribute{
						Description: "Exceptions to the rule",
						Optional:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Description: "Exception name",
									Required:    true,
								},
								"site":        siteScopeAttribute(),
								"source":      sourceAttribute(),
								"destination": destinationAttribute(),
								"service":     serviceAttribute(),
							},
						},
					},
				},
			},
		},
	}
}
