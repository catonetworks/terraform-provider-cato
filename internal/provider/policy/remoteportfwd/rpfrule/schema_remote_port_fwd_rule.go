package rpfrule

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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

func portRangeAttribute(description string) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: description,
		Required:    true,
		Attributes: map[string]schema.Attribute{
			"from": schema.StringAttribute{
				Description: "Range start",
				Required:    true,
			},
			"to": schema.StringAttribute{
				Description: "Range end",
				Required:    true,
			},
		},
	}
}

//nolint:funlen
func (r *remotePortFwdRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a rule in the Cato remote port forwarding policy. " +
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
					"external_ip": schema.SingleNestedAttribute{
						Description: "Allocated IP the external port range listens on",
						Required:    true,
						Attributes: map[string]schema.Attribute{
							"id": schema.StringAttribute{
								Optional: true,
								Computed: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
							"name": schema.StringAttribute{
								Optional: true,
								Computed: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
						},
					},
					"external_port_range": portRangeAttribute("External port range"),
					"internal_ip": schema.StringAttribute{
						Description: "Internal IP traffic is forwarded to",
						Required:    true,
					},
					"internal_port_range": portRangeAttribute("Internal port range"),
					"forward_icmp": schema.BoolAttribute{
						Description: "Whether ICMP is forwarded",
						Optional:    true,
						Computed:    true,
					},
					"remote_ips": schema.SingleNestedAttribute{
						Description: "Remote IPs the restriction type applies to",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"global_ip_range": refSetAttribute("Global IP range references"),
							"ip": schema.ListAttribute{
								Description: "IP addresses",
								Optional:    true,
								ElementType: types.StringType,
							},
							"ip_range": schema.ListNestedAttribute{
								Description: "IP ranges",
								Optional:    true,
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"from": schema.StringAttribute{Required: true},
										"to":   schema.StringAttribute{Required: true},
									},
								},
							},
							"subnet": schema.ListAttribute{
								Description: "Subnets",
								Optional:    true,
								ElementType: types.StringType,
							},
						},
					},
					"restriction_type": schema.StringAttribute{
						Description: "Remote IP restriction type",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("ALLOW_LIST", "BLOCK_LIST"),
						},
					},
					"tracking": schema.SingleNestedAttribute{
						Description: "Alert tracking. Optional - the API default is disabled.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"enabled": schema.BoolAttribute{
								Description: "Whether alerting is enabled",
								Required:    true,
							},
							"frequency": schema.StringAttribute{
								Description: "Alert frequency (HOURLY, DAILY, WEEKLY, IMMEDIATE)",
								Optional:    true,
							},
							"mailing_list":       refSetAttribute("Mailing list references"),
							"subscription_group": refSetAttribute("Subscription group references"),
							"webhook":            refSetAttribute("Webhook references"),
						},
					},
				},
			},
		},
	}
}
