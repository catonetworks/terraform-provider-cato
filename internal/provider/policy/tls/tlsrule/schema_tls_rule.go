package tlsrule

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/planmodifiers"
)

//nolint:funlen
func (r *tlsInspectionRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `_tls_rule` resource contains the configuration parameters necessary to add rule to the TLS " +
			"Inspection Policy. Documentation for the underlying API used in this resource can be found at " +
			"[mutation.policy.tlsInspect.addRule()](https://api.catonetworks.com/documentation/#mutation-policy.tlsInspect.addRule).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier of the TLS Inspection Rule",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"at": schema.SingleNestedAttribute{
				Description: "Position of the rule in the policy",
				Required:    true,
				Optional:    false,
				Attributes: map[string]schema.Attribute{
					"position": schema.StringAttribute{
						Description: "Position relative to a policy, a section or another rule (FIRST_IN_POLICY, " +
							"LAST_IN_POLICY, FIRST_IN_SECTION, LAST_IN_SECTION, BEFORE_RULE, AFTER_RULE)",
						Required: true,
						Optional: false,
						Validators: []validator.String{
							stringvalidator.OneOf("FIRST_IN_POLICY", "LAST_IN_POLICY", "FIRST_IN_SECTION", "LAST_IN_SECTION", "BEFORE_RULE", "AFTER_RULE"),
						},
					},
					"ref": schema.StringAttribute{
						Description: "The identifier of the object (e.g. a rule, a section) relative to which the position of the added rule is defined",
						Required:    false,
						Optional:    true,
					},
				},
			},
			"rule": schema.SingleNestedAttribute{
				Description: "Parameters for the rule you are adding",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Description: "ID of the rule",
						Computed:    true,
						Optional:    false,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"name": schema.StringAttribute{
						Description: "Name of the rule",
						Required:    true,
					},
					"description": schema.StringAttribute{
						Description: "Description of the rule",
						Required:    false,
						Optional:    true,
					},
					"enabled": schema.BoolAttribute{
						Description: "Attribute to define rule status (enabled or disabled)",
						Required:    true,
						Optional:    false,
					},
					"action": schema.StringAttribute{
						Description: "The action applied by TLS Inspection if the rule is matched (INSPECT, BYPASS)",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("INSPECT", "BYPASS"),
						},
					},
					"untrusted_certificate_action": schema.StringAttribute{
						Description: "Action for untrusted certificates (ALLOW, BLOCK, PROMPT)",
						Required:    false,
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("ALLOW", "BLOCK", "PROMPT"),
						},
					},
					"connection_origin": schema.StringAttribute{
						Description: "Connection origin filter (ANY, REMOTE, SITE)",
						Required:    false,
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("ANY", "REMOTE", "SITE"),
						},
					},
					"source": schema.SingleNestedAttribute{
						Description: "Source traffic matching criteria. Logical 'OR' is applied within the criteria set. " +
							"Logical 'AND' is applied between criteria sets.",
						Required: false,
						Optional: true,
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
							planmodifiers.SourceDestObjectModifier(),
						},
						Attributes: map[string]schema.Attribute{
							"ip": schema.ListAttribute{
								Description: "IPv4 address list",
								ElementType: types.StringType,
								Required:    false,
								Optional:    true,
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
							"host": schema.SetNestedAttribute{
								Description: "Hosts and servers defined for your account",
								Required:    false,
								Optional:    true,
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "Host ID",
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
											Computed: true,
										},
										"name": schema.StringAttribute{
											Description: "Host name",
											Computed:    true,
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
							"site": schema.SetNestedAttribute{
								Description: "Sites defined for your account",
								Required:    false,
								Optional:    true,
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "Site ID",
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
											Computed: true,
										},
										"name": schema.StringAttribute{
											Description: "Site name",
											Computed:    true,
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
							"subnet": schema.ListAttribute{
								Description: "Subnet list",
								ElementType: types.StringType,
								Required:    false,
								Optional:    true,
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
							"ip_range": schema.ListNestedAttribute{
								Description: "IP address ranges",
								Required:    false,
								Optional:    true,
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"from": schema.StringAttribute{
											Description: "Range start IP address",
											Required:    true,
										},
										"to": schema.StringAttribute{
											Description: "Range end IP address",
											Required:    true,
										},
									},
								},
							},
							"global_ip_range": schema.SetNestedAttribute{
								Description: "Global IP ranges",
								Required:    false,
								Optional:    true,
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "Global IP range ID",
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
											Computed: true,
										},
										"name": schema.StringAttribute{
											Description: "Global IP range name",
											Computed:    true,
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
							"network_interface": schema.SetNestedAttribute{
								Description: "Network interfaces",
								Required:    false,
								Optional:    true,
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "Network interface ID",
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
											Computed: true,
										},
										"name": schema.StringAttribute{
											Description: "Network interface name",
											Computed:    true,
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
							"site_network_subnet": schema.SetNestedAttribute{
								Description: "Site network subnets",
								Required:    false,
								Optional:    true,
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "Site network subnet ID",
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
											Computed: true,
										},
										"name": schema.StringAttribute{
											Description: "Site network subnet name",
											Computed:    true,
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
							"floating_subnet": schema.SetNestedAttribute{
								Description: "Floating subnets",
								Required:    false,
								Optional:    true,
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "Floating subnet ID",
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
											Computed: true,
										},
										"name": schema.StringAttribute{
											Description: "Floating subnet name",
											Computed:    true,
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
							"user": schema.SetNestedAttribute{
								Description: "Users",
								Required:    false,
								Optional:    true,
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "User ID",
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
											Computed: true,
										},
										"name": schema.StringAttribute{
											Description: "User name",
											Computed:    true,
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
							"users_group": schema.SetNestedAttribute{
								Description: "User groups",
								Required:    false,
								Optional:    true,
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "User group ID",
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
											Computed: true,
										},
										"name": schema.StringAttribute{
											Description: "User group name",
											Computed:    true,
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
							"group": schema.SetNestedAttribute{
								Description: "Groups",
								Required:    false,
								Optional:    true,
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "Group ID",
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
											Computed: true,
										},
										"name": schema.StringAttribute{
											Description: "Group name",
											Computed:    true,
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
							"system_group": schema.SetNestedAttribute{
								Description: "System groups",
								Required:    false,
								Optional:    true,
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "System group ID",
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
											Computed: true,
										},
										"name": schema.StringAttribute{
											Description: "System group name",
											Computed:    true,
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
						},
					},
					"country": schema.SetNestedAttribute{
						Description: "Countries",
						Required:    false,
						Optional:    true,
						PlanModifiers: []planmodifier.Set{
							setplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.Set{
							setvalidator.SizeAtLeast(1),
						},
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									Description: "Country ID",
									Optional:    true,
									PlanModifiers: []planmodifier.String{
										stringplanmodifier.UseStateForUnknown(),
									},
									Computed: true,
								},
								"name": schema.StringAttribute{
									Description: "Country name",
									Computed:    true,
									Optional:    true,
									PlanModifiers: []planmodifier.String{
										stringplanmodifier.UseStateForUnknown(),
									},
								},
							},
						},
					},
					"device_posture_profile": schema.SetNestedAttribute{
						Description: "Device posture profiles",
						Required:    false,
						Optional:    true,
						PlanModifiers: []planmodifier.Set{
							setplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.Set{
							setvalidator.SizeAtLeast(1),
						},
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									Description: "Device posture profile ID",
									Optional:    true,
									PlanModifiers: []planmodifier.String{
										stringplanmodifier.UseStateForUnknown(),
									},
									Computed: true,
								},
								"name": schema.StringAttribute{
									Description: "Device posture profile name",
									Computed:    true,
									Optional:    true,
									PlanModifiers: []planmodifier.String{
										stringplanmodifier.UseStateForUnknown(),
									},
								},
							},
						},
					},
					"platform": schema.StringAttribute{
						Description: "Platform filter (ANDROID, EMBEDDED, IOS, LINUX, MACOS, WINDOWS)",
						Required:    false,
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("ANDROID", "EMBEDDED", "IOS", "LINUX", "MACOS", "WINDOWS"),
						},
					},
					"application": schema.SingleNestedAttribute{
						Description: "Application matching criteria",
						Required:    false,
						Optional:    true,
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
							planmodifiers.SourceDestObjectModifier(),
						},
						Attributes: map[string]schema.Attribute{
							"application": schema.SetNestedAttribute{
								Description: "Applications",
								Required:    false,
								Optional:    true,
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "Application ID",
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
											Computed: true,
										},
										"name": schema.StringAttribute{
											Description: "Application name",
											Computed:    true,
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
							"custom_app": schema.SetNestedAttribute{
								Description: "Custom applications",
								Required:    false,
								Optional:    true,
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "Custom app ID",
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
											Computed: true,
										},
										"name": schema.StringAttribute{
											Description: "Custom app name",
											Computed:    true,
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
							"app_category": schema.SetNestedAttribute{
								Description: "Application categories",
								Required:    false,
								Optional:    true,
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "App category ID",
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
											Computed: true,
										},
										"name": schema.StringAttribute{
											Description: "App category name",
											Computed:    true,
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
							"custom_category": schema.SetNestedAttribute{
								Description: "Custom categories",
								Required:    false,
								Optional:    true,
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "Custom category ID",
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
											Computed: true,
										},
										"name": schema.StringAttribute{
											Description: "Custom category name",
											Computed:    true,
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
							"domain": schema.ListAttribute{
								Description: "Domain list",
								ElementType: types.StringType,
								Required:    false,
								Optional:    true,
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
							"fqdn": schema.ListAttribute{
								Description: "FQDN list",
								ElementType: types.StringType,
								Required:    false,
								Optional:    true,
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
							"ip": schema.ListAttribute{
								Description: "IPv4 address list",
								ElementType: types.StringType,
								Required:    false,
								Optional:    true,
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
							"subnet": schema.ListAttribute{
								Description: "Subnet list",
								ElementType: types.StringType,
								Required:    false,
								Optional:    true,
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
							"ip_range": schema.ListNestedAttribute{
								Description: "IP address ranges",
								Required:    false,
								Optional:    true,
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"from": schema.StringAttribute{
											Description: "Range start IP address",
											Required:    true,
										},
										"to": schema.StringAttribute{
											Description: "Range end IP address",
											Required:    true,
										},
									},
								},
							},
							"global_ip_range": schema.SetNestedAttribute{
								Description: "Global IP ranges",
								Required:    false,
								Optional:    true,
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "Global IP range ID",
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
											Computed: true,
										},
										"name": schema.StringAttribute{
											Description: "Global IP range name",
											Computed:    true,
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
							"remote_asn": schema.ListAttribute{
								Description: "Remote ASN list",
								ElementType: types.StringType,
								Required:    false,
								Optional:    true,
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
							"service": schema.SetNestedAttribute{
								Description: "Services",
								Required:    false,
								Optional:    true,
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "Service ID",
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
											Computed: true,
										},
										"name": schema.StringAttribute{
											Description: "Service name",
											Computed:    true,
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
							"custom_service": schema.SingleNestedAttribute{
								Description: "Custom service definition",
								Required:    false,
								Optional:    true,
								Attributes: map[string]schema.Attribute{
									"port": schema.ListAttribute{
										Description: "Port list",
										ElementType: types.StringType,
										Required:    false,
										Optional:    true,
										Validators: []validator.List{
											listvalidator.SizeAtLeast(1),
										},
									},
									"port_range": schema.SingleNestedAttribute{
										Description: "Port range",
										Required:    false,
										Optional:    true,
										Attributes: map[string]schema.Attribute{
											"from": schema.StringAttribute{
												Description: "Start port",
												Required:    true,
											},
											"to": schema.StringAttribute{
												Description: "End port",
												Required:    true,
											},
										},
									},
									"protocol": schema.StringAttribute{
										Description: "Protocol (TCP, UDP, ICMP, ANY)",
										Required:    true,
										Validators: []validator.String{
											stringvalidator.OneOf("TCP", "UDP", "ICMP", "ANY"),
										},
									},
								},
							},
							"custom_service_ip": schema.SingleNestedAttribute{
								Description: "Custom service IP definition",
								Required:    false,
								Optional:    true,
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Description: "Service name",
										Required:    true,
									},
									"ip": schema.StringAttribute{
										Description: "IP address",
										Required:    false,
										Optional:    true,
									},
									"ip_range": schema.SingleNestedAttribute{
										Description: "IP range",
										Required:    false,
										Optional:    true,
										Attributes: map[string]schema.Attribute{
											"from": schema.StringAttribute{
												Description: "Start IP",
												Required:    true,
											},
											"to": schema.StringAttribute{
												Description: "End IP",
												Required:    true,
											},
										},
									},
								},
							},
							"tls_inspect_category": schema.StringAttribute{
								Description: "TLS Inspection category (POPULAR_CLOUD_APPS, STREAMING_MEDIA)",
								Required:    false,
								Optional:    true,
								Validators: []validator.String{
									stringvalidator.OneOf("POPULAR_CLOUD_APPS", "STREAMING_MEDIA"),
								},
							},
							"country": schema.SetNestedAttribute{
								Description: "Countries",
								Required:    false,
								Optional:    true,
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "Country ID",
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
											Computed: true,
										},
										"name": schema.StringAttribute{
											Description: "Country name",
											Computed:    true,
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
