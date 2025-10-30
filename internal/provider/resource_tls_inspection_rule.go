package provider

import (
	"context"
	"fmt"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/planmodifiers"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &tlsInspectionRuleResource{}
	_ resource.ResourceWithConfigure   = &tlsInspectionRuleResource{}
	_ resource.ResourceWithImportState = &tlsInspectionRuleResource{}
)

func NewTlsInspectionRuleResource() resource.Resource {
	return &tlsInspectionRuleResource{}
}

type tlsInspectionRuleResource struct {
	client *catoClientData
}

func (r *tlsInspectionRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tls_rule"
}

func (r *tlsInspectionRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `_tls_rule` resource contains the configuration parameters necessary to add rule to the TLS Inspection Policy. Documentation for the underlying API used in this resource can be found at [mutation.policy.tlsInspect.addRule()](https://api.catonetworks.com/documentation/#mutation-policy.tlsInspect.addRule).",
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
						Description: "Position relative to a policy, a section or another rule (FIRST_IN_POLICY, LAST_IN_POLICY, FIRST_IN_SECTION, LAST_IN_SECTION, BEFORE_RULE, AFTER_RULE)",
						Required:    true,
						Optional:    false,
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
					"index": schema.Int64Attribute{
						Description: "Rule Index - computed value that may change due to rule reordering",
						Computed:    true,
						Optional:    false,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
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
						Description: "Source traffic matching criteria. Logical 'OR' is applied within the criteria set. Logical 'AND' is applied between criteria sets.",
						Required:    false,
						Optional:    true,
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

func (r *tlsInspectionRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoClientData)
}

func (r *tlsInspectionRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	var plan TlsInspectionRule
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := hydrateTlsRuleApi(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Warn(ctx, "TFLOG_WARN_TLS_input.create", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(input.create),
	})

	createRuleResponse, err := r.client.catov2.PolicyTLSInspectAddRule(ctx, input.create, r.client.AccountId)

	tflog.Warn(ctx, "TFLOG_WARN_TLS_createRuleResponse", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(createRuleResponse),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyTlsInspectAddRule error",
			err.Error(),
		)
		return
	}

	// Check for errors
	if createRuleResponse.Policy.TLSInspect.AddRule.Status != "SUCCESS" {
		for _, item := range createRuleResponse.Policy.TLSInspect.AddRule.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Creating Resource",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	// Publishing new rule
	tflog.Info(ctx, "publishing new TLS rule")
	_, err = r.client.catov2.PolicyTLSInspectPublishPolicyRevision(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyTlsInspectPublishPolicyRevision error",
			err.Error(),
		)
		return
	}

	// Read rule and hydrate response to state
	body, err := r.client.catov2.Tlsinspectpolicy(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyTlsInspect error",
			err.Error(),
		)
		return
	}

	ruleList := body.GetPolicy().TLSInspect.Policy.GetRules()
	currentRule := &cato_go_sdk.Tlsinspectpolicy_Policy_TLSInspect_Policy_Rules_Rule{}
	// Get current rule from response by ID
	for _, ruleListItem := range ruleList {
		if ruleListItem.GetRule().ID == createRuleResponse.GetPolicy().GetTLSInspect().GetAddRule().Rule.GetRule().ID {
			currentRule = ruleListItem.GetRule()
			resp.State.SetAttribute(
				ctx,
				path.Root("rule").AtName("id"),
				ruleListItem.GetRule().ID)
			break
		}
	}
	tflog.Warn(ctx, "TFLOG_WARN_TLS_createRule.readResponse", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(currentRule),
	})

	// Hydrate ruleInput from api response
	ruleInputRead, hydrateDiags := hydrateTlsRuleState(ctx, plan, currentRule)
	resp.Diagnostics.Append(hydrateDiags...)
	ruleInputRead.ID = types.StringValue(currentRule.ID)
	tflog.Info(ctx, "ruleInputRead - "+fmt.Sprintf("%v", ruleInputRead))

	ruleObject, diags := types.ObjectValueFrom(ctx, TlsInspectionRuleRuleAttrTypes, ruleInputRead)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Rule = ruleObject
	plan.ID = types.StringValue(currentRule.ID)

	// Set the complete plan to state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(diags...)
}

func (r *tlsInspectionRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TlsInspectionRule
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.client.catov2.Tlsinspectpolicy(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyTlsInspect error",
			err.Error(),
		)
		return
	}

	// Retrieve rule ID
	var ruleID string
	if !state.Rule.IsNull() && !state.Rule.IsUnknown() {
		rule := Policy_Policy_TlsInspect_Policy_Rules_Rule{}
		diags = state.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			tflog.Debug(ctx, "Failed to convert full state to struct, attempting to extract ID only")
			resp.Diagnostics = resp.Diagnostics[:0]

			ruleAttrs := state.Rule.Attributes()
			if idValue, ok := ruleAttrs["id"]; ok && !idValue.IsNull() {
				ruleID = idValue.(types.String).ValueString()
			} else {
				resp.Diagnostics.AddError(
					"Unable to extract rule ID from state",
					"Could not find rule ID in state during import",
				)
				return
			}
		} else {
			ruleID = rule.ID.ValueString()
		}
	} else {
		resp.Diagnostics.AddError(
			"Invalid state",
			"Rule state is null or unknown",
		)
		return
	}

	ruleList := body.GetPolicy().TLSInspect.Policy.GetRules()
	ruleExist := false
	currentRule := &cato_go_sdk.Tlsinspectpolicy_Policy_TLSInspect_Policy_Rules_Rule{}
	for _, ruleListItem := range ruleList {
		if ruleListItem.GetRule().ID == ruleID {
			ruleExist = true
			currentRule = ruleListItem.GetRule()

			resp.State.SetAttribute(
				ctx,
				path.Root("rule").AtName("id"),
				ruleListItem.GetRule().ID)
		}
	}

	// Remove resource if it doesn't exist anymore
	if !ruleExist {
		tflog.Warn(ctx, "TLS inspection rule not found, resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	tflog.Warn(ctx, "TFLOG_WARN_TLS_readRule.readResponse", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(currentRule),
	})

	ruleInput, hydrateDiags := hydrateTlsRuleState(ctx, state, currentRule)
	resp.Diagnostics.Append(hydrateDiags...)

	diags = resp.State.SetAttribute(ctx, path.Root("rule"), ruleInput)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(diags...)

	// Check if position is set in state, if not default to LAST_IN_POLICY
	positionValue := "LAST_IN_POLICY"
	refValue := types.StringNull()

	if !state.At.IsNull() && !state.At.IsUnknown() {
		statePosInput := PolicyRulePositionInput{}
		diags = state.At.As(ctx, &statePosInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if !diags.HasError() && !statePosInput.Position.IsNull() && !statePosInput.Position.IsUnknown() {
			positionValue = statePosInput.Position.ValueString()
			if !statePosInput.Ref.IsNull() && !statePosInput.Ref.IsUnknown() {
				refValue = statePosInput.Ref
			}
		}
	}

	curAtObj, diagstmp := types.ObjectValue(
		PositionAttrTypes,
		map[string]attr.Value{
			"position": types.StringValue(positionValue),
			"ref":      refValue,
		},
	)
	diags = resp.State.SetAttribute(ctx, path.Root("at"), curAtObj)
	diags = append(diags, diagstmp...)

	// Set ID at root level
	resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(ruleID))
}

func (r *tlsInspectionRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	var plan TlsInspectionRule
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := hydrateTlsRuleApi(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Setting input for moving rule
	inputMoveRule := cato_models.PolicyMoveRuleInput{}

	// Setting at (to move rule)
	if !plan.At.IsNull() {
		inputMoveRule.To = &cato_models.PolicyRulePositionInput{}
		positionInput := PolicyRulePositionInput{}
		diags = plan.At.As(ctx, &positionInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)

		inputMoveRule.To.Position = (*cato_models.PolicyRulePositionEnum)(positionInput.Position.ValueStringPointer())
		inputMoveRule.To.Ref = positionInput.Ref.ValueStringPointer()
	}

	ruleInput := Policy_Policy_TlsInspect_Policy_Rules_Rule{}
	diags = plan.Rule.As(ctx, &ruleInput, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)

	// Settings other rule attributes
	inputMoveRule.ID = *ruleInput.ID.ValueStringPointer()
	input.update.ID = *ruleInput.ID.ValueStringPointer()

	// Move rule
	moveRule, err := r.client.catov2.PolicyTLSInspectMoveRule(ctx, inputMoveRule, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyTlsInspectMoveRule error",
			err.Error(),
		)
		return
	}

	// Check for errors
	if moveRule.Policy.TLSInspect.MoveRule.Status != "SUCCESS" {
		for _, item := range moveRule.Policy.TLSInspect.MoveRule.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Moving Rule Resource",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	tflog.Warn(ctx, "TFLOG_WARN_TLS_input.update", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(input.update),
	})

	// Updating rule
	updateRuleResponse, err := r.client.catov2.PolicyTLSInspectUpdateRule(ctx, input.update, r.client.AccountId)
	tflog.Warn(ctx, "TFLOG_WARN_TLS_updateRuleResponse", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(updateRuleResponse),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyTlsInspectUpdateRule error",
			err.Error(),
		)
		return
	}

	// Check for errors
	if updateRuleResponse.Policy.TLSInspect.UpdateRule.Status != "SUCCESS" {
		for _, item := range updateRuleResponse.Policy.TLSInspect.UpdateRule.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Updating Resource",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	// Publishing updated rule
	tflog.Info(ctx, "publishing updated TLS rule")
	_, err = r.client.catov2.PolicyTLSInspectPublishPolicyRevision(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyTlsInspectPublishPolicyRevision error",
			err.Error(),
		)
		return
	}

	// Read rule and hydrate response to state
	body, err := r.client.catov2.Tlsinspectpolicy(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyTlsInspect error",
			err.Error(),
		)
		return
	}

	ruleList := body.GetPolicy().TLSInspect.Policy.GetRules()
	currentRule := &cato_go_sdk.Tlsinspectpolicy_Policy_TLSInspect_Policy_Rules_Rule{}
	for _, ruleListItem := range ruleList {
		if ruleListItem.GetRule().ID == updateRuleResponse.GetPolicy().GetTLSInspect().GetUpdateRule().Rule.GetRule().ID {
			currentRule = ruleListItem.GetRule()
			break
		}
	}

	tflog.Warn(ctx, "TFLOG_WARN_TLS_updateRule.readResponse", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(currentRule),
	})

	// Hydrate ruleInput from api response
	ruleInputRead, hydrateDiags := hydrateTlsRuleState(ctx, plan, currentRule)
	resp.Diagnostics.Append(hydrateDiags...)
	ruleInputRead.ID = types.StringValue(currentRule.ID)

	ruleObject, diags := types.ObjectValueFrom(ctx, TlsInspectionRuleRuleAttrTypes, ruleInputRead)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Rule = ruleObject

	// Set the complete plan to state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(diags...)
}

func (r *tlsInspectionRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state TlsInspectionRule
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ruleInput := Policy_Policy_TlsInspect_Policy_Rules_Rule{}
	diags = state.Rule.As(ctx, &ruleInput, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)

	removeRuleInput := cato_models.TLSInspectRemoveRuleInput{
		ID: *ruleInput.ID.ValueStringPointer(),
	}

	removeRuleResponse, err := r.client.catov2.PolicyTLSInspectRemoveRule(ctx, removeRuleInput, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyTlsInspectRemoveRule error",
			err.Error(),
		)
		return
	}

	// Check for errors
	if removeRuleResponse.Policy.TLSInspect.RemoveRule.Status != "SUCCESS" {
		for _, item := range removeRuleResponse.Policy.TLSInspect.RemoveRule.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Deleting Resource",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	// Publishing rule deletion
	tflog.Info(ctx, "publishing TLS rule deletion")
	_, err = r.client.catov2.PolicyTLSInspectPublishPolicyRevision(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyTlsInspectPublishPolicyRevision error",
			err.Error(),
		)
		return
	}
}

func (r *tlsInspectionRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	// Set rule.id to the imported ID
	resp.State.SetAttribute(ctx, path.Root("rule").AtName("id"), req.ID)
}
