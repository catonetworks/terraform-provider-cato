package provider

import (
	"context"
	"fmt"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/planmodifiers"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &internetFwRuleResource{}
	_ resource.ResourceWithConfigure   = &internetFwRuleResource{}
	_ resource.ResourceWithImportState = &internetFwRuleResource{}
)

type internetFwRuleResource struct {
	client *catoClientData
}

func NewInternetFwRuleResource() resource.Resource {
	return &internetFwRuleResource{}
}

func (r *internetFwRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_if_rule"
}

func (r *internetFwRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_if_rule` resource contains the configuration parameters necessary to add rule to the Internet Firewall. (check https://support.catonetworks.com/hc/en-us/articles/4413273486865-What-is-the-Cato-Internet-Firewall for more details). Documentation for the underlying API used in this resource can be found at [mutation.policy.internetFirewall.addRule()](https://api.catonetworks.com/documentation/#mutation-policy.internetFirewall.addRule).",
		Attributes: map[string]schema.Attribute{
			"at": schema.SingleNestedAttribute{
				Description: "Position of the rule in the policy (https://api.catonetworks.com/documentation/#definition-PolicyRulePositionInput)",
				Required:    true,
				Optional:    false,
				Attributes: map[string]schema.Attribute{
					"position": schema.StringAttribute{
						Description: "Position relative to a policy, a section or another rule (https://api.catonetworks.com/documentation/#definition-PolicyRulePositionEnum)",
						Required:    true,
						Optional:    false,
						Validators: []validator.String{
							stringvalidator.OneOf("AFTER_RULE", "BEFORE_RULE", "FIRST_IN_POLICY", "FIRST_IN_SECTION", "LAST_IN_POLICY", "LAST_IN_SECTION"),
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
				Description: "Parameters for the rule you are adding (https://api.catonetworks.com/documentation/#definition-InternetFirewallAddRuleDataInput)",
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
					},
					"enabled": schema.BoolAttribute{
						Description: "Attribute to define rule status (enabled or disabled)",
						Required:    true,
						Optional:    false,
					},
					// "section": schema.SingleNestedAttribute{
					// 	Required: false,
					// 	Optional: true,
					// 	Computed: true,
					// 	PlanModifiers: []planmodifier.Object{
					// 		objectplanmodifier.UseStateForUnknown(), // Preserve section from API when not in config
					// 	},
					// 	Attributes: map[string]schema.Attribute{
					// 		"name": schema.StringAttribute{
					// 			Description: "",
					// 			Required:    false,
					// 			Optional:    true,
					// 		},
					// 		"id": schema.StringAttribute{
					// 			Description: "",
					// 			Required:    false,
					// 			Optional:    true,
					// 		},
					// 	},
					// },
					"active_period": schema.SingleNestedAttribute{
						Description: "Time period during which the rule is active. Outside this period, the rule is inactive. Times should be in RFC3339 format (e.g., '2024-12-31T23:59:59Z').",
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.Object{
							planmodifiers.ActivePeriodModifier(),
						},
						Attributes: map[string]schema.Attribute{
							"effective_from": schema.StringAttribute{
								Description: "The time the rule becomes active (RFC3339 format, e.g., '2024-01-01T00:00:00Z'). If not specified, the rule is active from creation.",
								Optional:    true,
								Required:    false,
								Computed:    true,
							},
							"expires_at": schema.StringAttribute{
								Description: "The time the rule expires and becomes inactive (RFC3339 format, e.g., '2024-12-31T23:59:59Z'). If not specified, the rule never expires.",
								Optional:    true,
								Required:    false,
								Computed:    true,
							},
							"use_effective_from": schema.BoolAttribute{
								Description: "Whether to use the effective_from time. Computed from the presence of effective_from field.",
								Computed:    true,
								PlanModifiers: []planmodifier.Bool{
									boolplanmodifier.UseStateForUnknown(),
								},
							},
							"use_expires_at": schema.BoolAttribute{
								Description: "Whether to use the expires_at time. Computed from the presence of expires_at field.",
								Computed:    true,
								PlanModifiers: []planmodifier.Bool{
									boolplanmodifier.UseStateForUnknown(),
								},
							},
						},
					},
					"source": schema.SingleNestedAttribute{
						Description: "Source traffic matching criteria. Logical ‘OR’ is applied within the criteria set. Logical ‘AND’ is applied between criteria sets. (https://api.catonetworks.com/documentation/#definition-InternetFirewallSourceInput)",
						Required:    true,
						Optional:    false,
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),  // Avoid drift
							planmodifiers.SourceDestObjectModifier(), // Handle ID correlation for nested sets
						},
						Attributes: map[string]schema.Attribute{
							"ip": schema.ListAttribute{
								Description: "Pv4 address list",
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
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(), // Avoid drift
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
											Validators: []validator.String{
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Computed: true,
										},
										"id": schema.StringAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Computed: true,
										},
									},
								},
							},
							"site": schema.SetNestedAttribute{
								Description: "Site defined for the account",
								Required:    false,
								Optional:    true,
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(), // Avoid drift
								},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
											Validators: []validator.String{
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Computed: true,
										},
										"id": schema.StringAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Computed: true,
										},
									},
								},
							},
							"subnet": schema.ListAttribute{
								ElementType: types.StringType,
								Description: "Subnets and network ranges defined for the LAN interfaces of a site",
								Required:    false,
								Optional:    true,
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
							"ip_range": schema.ListNestedAttribute{
								Description: "Multiple separate IP addresses or an IP range",
								Required:    false,
								Optional:    true,
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"from": schema.StringAttribute{
											Description: "",
											Required:    true,
											Optional:    false,
										},
										"to": schema.StringAttribute{
											Description: "",
											Required:    true,
											Optional:    false,
										},
									},
								},
							},
							"global_ip_range": schema.SetNestedAttribute{
								Description: "Globally defined IP range, IP and subnet objects",
								Required:    false,
								Optional:    true,
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(), // Avoid drift
								},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
											Validators: []validator.String{
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Computed: true,
										},
										"id": schema.StringAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Computed: true,
										},
									},
								},
							},
							"network_interface": schema.SetNestedAttribute{
								Description: "Network range defined for a site",
								Required:    false,
								Optional:    true,
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(), // Avoid drift
								},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
											Validators: []validator.String{
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Computed: true,
										},
										"id": schema.StringAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Computed: true,
										},
									},
								},
							},
							"site_network_subnet": schema.SetNestedAttribute{
								Description: "GlobalRange + InterfaceSubnet",
								Required:    false,
								Optional:    true,
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(), // Avoid drift
								},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
											Validators: []validator.String{
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Computed: true,
										},
										"id": schema.StringAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Computed: true,
										},
									},
								},
							},
							"floating_subnet": schema.SetNestedAttribute{
								Description: "Floating Subnets (ie. Floating Ranges) are used to identify traffic exactly matched to the route advertised by BGP. They are not associated with a specific site. This is useful in scenarios such as active-standby high availability routed via BGP.",
								Required:    false,
								Optional:    true,
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(), // Avoid drift
								},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
											Validators: []validator.String{
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
											Computed: true,
										},
										"id": schema.StringAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
											Computed: true,
										},
									},
								},
							},
							"user": schema.SetNestedAttribute{
								Description: "Individual users defined for the account",
								Required:    false,
								Optional:    true,
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(), // Avoid drift
								},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
											Validators: []validator.String{
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Computed: true,
										},
										"id": schema.StringAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Computed: true,
										},
									},
								},
							},
							"users_group": schema.SetNestedAttribute{
								Description: "Group of users",
								Required:    false,
								Optional:    true,
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(), // Avoid drift
								},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
											Validators: []validator.String{
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Computed: true,
										},
										"id": schema.StringAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Computed: true,
										},
									},
								},
							},
							"group": schema.SetNestedAttribute{
								Description: "Groups defined for your account",
								Required:    false,
								Optional:    true,
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(), // Avoid drift
								},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
											Validators: []validator.String{
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Computed: true,
										},
										"id": schema.StringAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Computed: true,
										},
									},
								},
							},
							"system_group": schema.SetNestedAttribute{
								Description: "Predefined Cato groups",
								Required:    false,
								Optional:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
											Validators: []validator.String{
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Computed: true,
										},
										"id": schema.StringAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Computed: true,
										},
									},
								},
							},
						},
					},
					"connection_origin": schema.StringAttribute{
						Description: "Connection origin of the traffic (https://api.catonetworks.com/documentation/#definition-ConnectionOriginEnum)",
						Optional:    true,
						Required:    false,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.String{
							stringvalidator.OneOf("ANY", "REMOTE", "SITE"),
						},
						Computed: true,
					},
					"country": schema.SetNestedAttribute{
						Description: "Source country traffic matching criteria. Logical ‘OR’ is applied within the criteria set. Logical ‘AND’ is applied between criteria sets.",
						Required:    false,
						Optional:    true,
						Validators: []validator.Set{
							setvalidator.SizeAtLeast(1),
						},
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Description: "",
									Required:    false,
									Optional:    true,
									Validators: []validator.String{
										stringvalidator.ConflictsWith(path.Expressions{
											path.MatchRelative().AtParent().AtName("id"),
										}...),
									},
									PlanModifiers: []planmodifier.String{
										stringplanmodifier.UseStateForUnknown(), // Avoid drift
									},
									Computed: true,
								},
								"id": schema.StringAttribute{
									Description: "",
									Required:    false,
									Optional:    true,
									PlanModifiers: []planmodifier.String{
										stringplanmodifier.UseStateForUnknown(), // Avoid drift
									},
									Computed: true,
								},
							},
						},
					},
					"device": schema.SetNestedAttribute{
						Description: "Source Device Profile traffic matching criteria. Logical ‘OR’ is applied within the criteria set. Logical ‘AND’ is applied between criteria sets.",
						Required:    false,
						Optional:    true,
						Validators: []validator.Set{
							setvalidator.SizeAtLeast(1),
						},
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Description: "",
									Required:    false,
									Optional:    true,
									Validators: []validator.String{
										stringvalidator.ConflictsWith(path.Expressions{
											path.MatchRelative().AtParent().AtName("id"),
										}...),
									},
									PlanModifiers: []planmodifier.String{
										stringplanmodifier.UseStateForUnknown(), // Avoid drift
									},
									Computed: true,
								},
								"id": schema.StringAttribute{
									Description: "",
									Required:    false,
									Optional:    true,
									PlanModifiers: []planmodifier.String{
										stringplanmodifier.UseStateForUnknown(), // Avoid drift
									},
									Computed: true,
								},
							},
						},
					},
					"device_os": schema.ListAttribute{
						ElementType: types.StringType,
						Description: "Source device Operating System traffic matching criteria. Logical ‘OR’ is applied within the criteria set. Logical ‘AND’ is applied between criteria sets.(https://api.catonetworks.com/documentation/#definition-OperatingSystem)",
						Optional:    true,
						Required:    false,
					},
					"device_attributes": schema.SingleNestedAttribute{
						Description: "Device attributes matching criteria for the rule.",
						Required:    false,
						Optional:    true,
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(), // Avoid drift
						},
						Attributes: map[string]schema.Attribute{
							"category": schema.ListAttribute{
								ElementType: types.StringType,
								Description: "Device category matching criteria for the rule.",
								Required:    false,
								Optional:    true,
								PlanModifiers: []planmodifier.List{
									listplanmodifier.UseStateForUnknown(), // Avoid drift
								},
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
							"type": schema.ListAttribute{
								ElementType: types.StringType,
								Description: "Device type matching criteria for the rule.",
								Required:    false,
								Optional:    true,
								PlanModifiers: []planmodifier.List{
									listplanmodifier.UseStateForUnknown(), // Avoid drift
								},
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
							"model": schema.ListAttribute{
								ElementType: types.StringType,
								Description: "Device model matching criteria for the rule.",
								Required:    false,
								Optional:    true,
								PlanModifiers: []planmodifier.List{
									listplanmodifier.UseStateForUnknown(), // Avoid drift
								},
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
							"manufacturer": schema.ListAttribute{
								ElementType: types.StringType,
								Description: "Device manufacturer matching criteria for the rule.",
								Required:    false,
								Optional:    true,
								PlanModifiers: []planmodifier.List{
									listplanmodifier.UseStateForUnknown(), // Avoid drift
								},
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
							"os": schema.ListAttribute{
								ElementType: types.StringType,
								Description: "Device OS matching criteria for the rule.",
								Required:    false,
								Optional:    true,
								PlanModifiers: []planmodifier.List{
									listplanmodifier.UseStateForUnknown(), // Avoid drift
								},
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
							"os_version": schema.ListAttribute{
								ElementType: types.StringType,
								Description: "Device OS version matching criteria for the rule.",
								Required:    false,
								Optional:    true,
								PlanModifiers: []planmodifier.List{
									listplanmodifier.UseStateForUnknown(), // Avoid drift
								},
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
						},
					},
					"destination": schema.SingleNestedAttribute{
						Description: "Destination traffic matching criteria. Logical ‘OR’ is applied within the criteria set. Logical ‘AND’ is applied between criteria sets. (https://api.catonetworks.com/documentation/#definition-InternetFirewallDestinationInput)",
						Optional:    false,
						Required:    true,
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),  // Avoid drift
							planmodifiers.SourceDestObjectModifier(), // Handle ID correlation for nested sets
						},
						Attributes: map[string]schema.Attribute{
							"application": schema.SetNestedAttribute{
								Description: "Applications for the rule (pre-defined)",
								Required:    false,
								Optional:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
											Validators: []validator.String{
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Computed: true,
										},
										"id": schema.StringAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Computed: true,
										},
									},
								},
							},
							"custom_app": schema.SetNestedAttribute{
								Description: "Custom (user-defined) applications",
								Required:    false,
								Optional:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
											Validators: []validator.String{
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Computed: true,
										},
										"id": schema.StringAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Computed: true,
										},
									},
								},
							},
							"app_category": schema.SetNestedAttribute{
								Description: "Cato category of applications which are dynamically updated by Cato",
								Required:    false,
								Optional:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
											Validators: []validator.String{
												stringvalidator.OneOf("Advertisements", "AI Media Generators", "Alcohol and Tobacco", "Anonymizers", "Authentication Services", "Beauty", "Botnets", "Business Information", "Business Operations AI", "Business Systems", "CDN", "Chat and IM", "Cheating", "Code Assistants", "Compromised", "Computers and Technology", "Conversational AI", "Criminal Activity", "Cults", "Database", "DNS over HTTPS", "Education", "Email", "Entertainment", "ERP And CRM", "File Sharing", "Finance", "Gambling", "Games", "General", "Generative AI Tools", "Government", "Greeting Cards", "Hacking", "Health and Medicine", "Healthcare AI", "Hiring", "Illegal Drugs", "Industrial Protocols", "Information Security", "Internet Conferencing", "Keyloggers", "Leisure and Recreation", "Malware", "Media Streams", "Military", "Network Protocol", "Network Utilities", "News", "Nudity", "Office Programs And Services", "Online Storage", "P2P", "Parked domains", "PDF Converters", "Personal Sites", "Phishing", "Politics", "Porn", "Productivity", "Questionable", "Real Estate", "Religion", "Remote Access", "Search Engines and Portals", "Sex education", "Shopping", "Social", "Software Downloads", "Software Updates", "SPAM", "Sports", "Spyware", "Tasteless", "Translation", "Travel AI Assistance", "Travel", "Uncategorized", "Undefined", "Vehicles", "Violence and Hate", "Voip Video", "Weapons", "Web Hosting", "Web Posting", "Writing Assistants"),
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Computed: true,
										},
										"id": schema.StringAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
											Validators: []validator.String{
												stringvalidator.OneOf("advertisements", "ai_tools", "anonymizers", "authentication_services", "beauty", "botnets", "business_systems", "business", "cdn", "chat_and_im", "cheating", "computers_and_technology", "criminal_activity", "cults", "database", "dns_over_https", "drugs", "education", "email", "entertainment", "erp_and_crm", "file_sharing", "finance", "food_drinks_tobacco", "gambling", "games", "gen_ai_business_operations", "gen_ai_code_assistants", "gen_ai_conversational_ai", "gen_ai_healthcare", "gen_ai_media_generators", "gen_ai_productivity", "gen_ai_travel_assistance", "gen_ai_writing_assistants", "general", "government", "greeting_cards", "hacking", "health_and_medicine", "hiring", "information_security", "internet_conferencing", "keyloggers", "leisure_and_recreation", "media_streams", "military", "network_protocol", "network_utilities", "news", "nudity", "office_programs_and_services", "online_storage", "ot_protocols", "p2p", "parked_domains", "pdf_converters", "personal_sites", "politics", "porn", "questionable", "real_estate", "religion", "remote_access", "search_engines_and_portals", "sex_education", "shopping", "social", "software_downloads", "software_updates", "spam", "sports", "spyware", "suspected_malware", "suspected_phishing", "suspected_unwanted", "tasteless", "translation", "travel", "uncategorized", "undefined", "vehicles", "violence", "voip_video", "weapons", "web_hosting", "web_posting"),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Computed: true,
										},
									},
								},
							},
							"custom_category": schema.SetNestedAttribute{
								Description: "Custom Categories – Groups of objects such as predefined and custom applications, predefined and custom services, domains, FQDNs etc.",
								Required:    false,
								Optional:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
											Validators: []validator.String{
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Computed: true,
										},
										"id": schema.StringAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Computed: true,
										},
									},
								},
							},
							"sanctioned_apps_category": schema.SetNestedAttribute{
								Description: "Sanctioned Cloud Applications - apps that are approved and generally represent an understood and acceptable level of risk in your organization.",
								Required:    false,
								Optional:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
											Validators: []validator.String{
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Computed: true,
										},
										"id": schema.StringAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Computed: true,
										},
									},
								},
							},
							"country": schema.SetNestedAttribute{
								Description: "Countries",
								Required:    false,
								Optional:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
											Validators: []validator.String{
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Computed: true,
										},
										"id": schema.StringAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Computed: true,
										},
									},
								},
							},
							"domain": schema.ListAttribute{
								ElementType: types.StringType,
								Description: "A Second-Level Domain (SLD). It matches all Top-Level Domains (TLD), and subdomains that include the Domain. Example: example.com.",
								Required:    false,
								Optional:    true,
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
							"fqdn": schema.ListAttribute{
								ElementType: types.StringType,
								Description: "An exact match of the fully qualified domain (FQDN). Example: www.my.example.com.",
								Required:    false,
								Optional:    true,
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
							"ip": schema.ListAttribute{
								ElementType: types.StringType,
								Description: "IPv4 addresses",
								Required:    false,
								Optional:    true,
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
							"subnet": schema.ListAttribute{
								ElementType: types.StringType,
								Description: "Network subnets in CIDR notation",
								Required:    false,
								Optional:    true,
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
							"ip_range": schema.ListNestedAttribute{
								Description: "A range of IPs. Every IP within the range will be matched",
								Required:    false,
								Optional:    true,
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"from": schema.StringAttribute{
											Description: "",
											Required:    true,
											Optional:    false,
										},
										"to": schema.StringAttribute{
											Description: "",
											Required:    true,
											Optional:    false,
										},
									},
								},
							},
							"global_ip_range": schema.SetNestedAttribute{
								Description: "Globally defined IP range, IP and subnet objects.",
								Required:    false,
								Optional:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
											Validators: []validator.String{
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Computed: true,
										},
										"id": schema.StringAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Computed: true,
										},
									},
								},
							},
							"remote_asn": schema.ListAttribute{
								ElementType: types.StringType,
								Description: "Remote Autonomous System Number (ASN)",
								Required:    false,
								Optional:    true,
							},
						},
					},
					"service": schema.SingleNestedAttribute{
						Description: "Destination service traffic matching criteria. Logical ‘OR’ is applied within the criteria set. Logical ‘AND’ is applied between criteria sets.",
						Required:    false,
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"standard": schema.SetNestedAttribute{
								Description: "Standard Service to which this Internet Firewall rule applies",
								Required:    false,
								Optional:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
											Validators: []validator.String{
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Computed: true,
										},
										"id": schema.StringAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Computed: true,
										},
									},
								},
							},
							"custom": schema.ListNestedAttribute{
								Description: "Custom Service defined by a combination of L4 ports and an IP Protocol",
								Required:    false,
								Optional:    true,
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"port": schema.ListAttribute{
											ElementType: types.StringType,
											Description: "List of TCP/UDP port",
											Optional:    true,
											Required:    false,
											Validators: []validator.List{
												listvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("port_range"),
												}...),
											},
										},
										"port_range": schema.SingleNestedAttribute{
											Description: "TCP/UDP port ranges",
											Required:    false,
											Optional:    true,
											Validators: []validator.Object{
												objectvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("port"),
												}...),
											},
											Attributes: map[string]schema.Attribute{
												"from": schema.StringAttribute{
													Description: "",
													Required:    true,
													Optional:    false,
												},
												"to": schema.StringAttribute{
													Description: "",
													Required:    true,
													Optional:    false,
												},
											},
										},
										"protocol": schema.StringAttribute{
											Description: "IP Protocol (https://api.catonetworks.com/documentation/#definition-IpProtocol)",
											Required:    false,
											Optional:    true,
										},
									},
								},
							},
						},
					},
					"action": schema.StringAttribute{
						Description: "The action applied by the Internet Firewall if the rule is matched (https://api.catonetworks.com/documentation/#definition-InternetFirewallActionEnum)",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("ALLOW", "BLOCK", "PROMPT", "RBI"),
						},
					},
					"tracking": schema.SingleNestedAttribute{
						Description: "Tracking information when the rule is matched, such as events and notifications",
						Required:    true,
						Optional:    false,
						Attributes: map[string]schema.Attribute{
							"event": schema.SingleNestedAttribute{
								Description: "When enabled, create an event each time the rule is matched",
								Required:    true,
								Attributes: map[string]schema.Attribute{
									"enabled": schema.BoolAttribute{
										Description: "",
										Required:    false,
										Optional:    true,
										Computed:    true,
										Default:     booldefault.StaticBool(false),
									},
								},
							},
							"alert": schema.SingleNestedAttribute{
								Description: "When enabled, send an alert each time the rule is matched",
								Required:    false,
								Optional:    true,
								PlanModifiers: []planmodifier.Object{
									objectplanmodifier.UseStateForUnknown(), // Avoid drift
								},
								Computed: true,
								Attributes: map[string]schema.Attribute{
									"enabled": schema.BoolAttribute{
										Description: "",
										Optional:    true,
										Required:    false,
										PlanModifiers: []planmodifier.Bool{
											boolplanmodifier.UseStateForUnknown(), // Avoid drift
										},
										Computed: true,
										Default:  booldefault.StaticBool(false),
									},
									"frequency": schema.StringAttribute{
										Description: "Returns data for the alert frequency (https://api.catonetworks.com/documentation/#definition-PolicyRuleTrackingFrequencyEnum)",
										Optional:    true,
										Required:    false,
										Validators: []validator.String{
											stringvalidator.OneOf("DAILY", "HOURLY", "IMMEDIATE", "WEEKLY"),
										},
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(), // Avoid drift
										},
										Default:  stringdefault.StaticString("HOURLY"),
										Computed: true,
									},
									"subscription_group": schema.SetNestedAttribute{
										Description: "Returns data for the Subscription Group that receives the alert",
										Required:    false,
										Optional:    true,
										Validators: []validator.Set{
											setvalidator.SizeAtLeast(1),
										},
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"name": schema.StringAttribute{
													Description: "",
													Required:    false,
													Optional:    true,
													Validators: []validator.String{
														stringvalidator.ConflictsWith(path.Expressions{
															path.MatchRelative().AtParent().AtName("id"),
														}...),
													},
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.UseStateForUnknown(), // Avoid drift
													},
													Computed: true,
												},
												"id": schema.StringAttribute{
													Description: "",
													Required:    false,
													Optional:    true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.UseStateForUnknown(), // Avoid drift
													},
													Computed: true,
												},
											},
										},
									},
									"webhook": schema.SetNestedAttribute{
										Description: "Returns data for the Webhook that receives the alert",
										Required:    false,
										Optional:    true,
										Validators: []validator.Set{
											setvalidator.SizeAtLeast(1),
										},
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"name": schema.StringAttribute{
													Description: "",
													Required:    false,
													Optional:    true,
													Validators: []validator.String{
														stringvalidator.ConflictsWith(path.Expressions{
															path.MatchRelative().AtParent().AtName("id"),
														}...),
													},
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.UseStateForUnknown(), // Avoid drift
													},
													Computed: true,
												},
												"id": schema.StringAttribute{
													Description: "",
													Required:    false,
													Optional:    true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.UseStateForUnknown(), // Avoid drift
													},
													Computed: true,
												},
											},
										},
									},
									"mailing_list": schema.SetNestedAttribute{
										Description: "Returns data for the Mailing List that receives the alert",
										Required:    false,
										Optional:    true,
										Validators: []validator.Set{
											setvalidator.SizeAtLeast(1),
										},
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"name": schema.StringAttribute{
													Description: "",
													Required:    false,
													Optional:    true,
													Validators: []validator.String{
														stringvalidator.ConflictsWith(path.Expressions{
															path.MatchRelative().AtParent().AtName("id"),
														}...),
													},
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.UseStateForUnknown(), // Avoid drift
													},
													Computed: true,
												},
												"id": schema.StringAttribute{
													Description: "",
													Required:    false,
													Optional:    true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.UseStateForUnknown(), // Avoid drift
													},
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
					},
					"schedule": schema.SingleNestedAttribute{
						Description: "The time period specifying when the rule is enabled, otherwise it is disabled.",
						Required:    false,
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(), // Avoid drift
						},
						Attributes: map[string]schema.Attribute{
							"active_on": schema.StringAttribute{
								Description: "Define when the rule is active (https://api.catonetworks.com/documentation/#definition-PolicyActiveOnEnum)",
								Required:    false,
								Optional:    true,
								Computed:    true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(), // Avoid drift
								},
								Validators: []validator.String{
									stringvalidator.OneOf("ALWAYS", "CUSTOM_RECURRING", "CUSTOM_TIMEFRAME", "WORKING_HOURS"),
								},
							},
							"custom_timeframe": schema.SingleNestedAttribute{
								Description: "Input of data for a custom one-time time range that a rule is active",
								Required:    false,
								Optional:    true,
								PlanModifiers: []planmodifier.Object{
									objectplanmodifier.UseStateForUnknown(), // Avoid drift
								},
								Attributes: map[string]schema.Attribute{
									"from": schema.StringAttribute{
										Description: "",
										Required:    false,
										Optional:    true,
									},
									"to": schema.StringAttribute{
										Description: "",
										Required:    false,
										Optional:    true,
									},
								},
							},
							"custom_recurring": schema.SingleNestedAttribute{
								Description: "Input of data for a custom recurring time range that a rule is active",
								Required:    false,
								Optional:    true,
								PlanModifiers: []planmodifier.Object{
									objectplanmodifier.UseStateForUnknown(), // Avoid drift
								},
								Attributes: map[string]schema.Attribute{
									"from": schema.StringAttribute{
										Description: "",
										Required:    false,
										Optional:    true,
									},
									"to": schema.StringAttribute{
										Description: "",
										Required:    false,
										Optional:    true,
									},
									"days": schema.ListAttribute{
										ElementType: types.StringType,
										Description: "(https://api.catonetworks.com/documentation/#definition-DayOfWeek)",
										Required:    false,
										Optional:    true,
										Validators: []validator.List{
											listvalidator.ValueStringsAre(stringvalidator.OneOf("SUNDAY", "MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY", "FRIDAY", "SATURDAY")),
										},
									},
								},
							},
						},
					},
					"exceptions": schema.SetNestedAttribute{
						Description: "The set of exceptions for the rule. Exceptions define when the rule will be ignored and the firewall evaluation will continue with the lower priority rules.",
						Required:    false,
						Optional:    true,
						Computed:    true,
						Validators: []validator.Set{
							setvalidator.SizeAtLeast(1),
						},
						PlanModifiers: []planmodifier.Set{
							planmodifiers.IfwExceptionsSetModifier(), // Handle ID correlation for Internet FW exceptions
						},
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Description: "A unique name of the rule exception.",
									Required:    false,
									Optional:    true,
									PlanModifiers: []planmodifier.String{
										stringplanmodifier.UseStateForUnknown(), // Avoid drift
									},
								},
								"source": schema.SingleNestedAttribute{
									Description: "Source traffic matching criteria for the exception.",
									Required:    false,
									Optional:    true,
									PlanModifiers: []planmodifier.Object{
										objectplanmodifier.UseStateForUnknown(), // Avoid drift
									},
									Attributes: map[string]schema.Attribute{
										"ip": schema.ListAttribute{
											Description: "",
											ElementType: types.StringType,
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.List{
												listplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Validators: []validator.List{
												listvalidator.SizeAtLeast(1),
											},
										},
										"host": schema.SetNestedAttribute{
											Description: "Hosts and servers defined for your account",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.Set{
												setplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"name": schema.StringAttribute{
														Description: "Host Name",
														Required:    false,
														Optional:    true,
														Validators: []validator.String{
															stringvalidator.ConflictsWith(path.Expressions{
																path.MatchRelative().AtParent().AtName("id"),
															}...),
														},
														PlanModifiers: []planmodifier.String{
															stringplanmodifier.UseStateForUnknown(), // Avoid drift
														},
														Computed: true,
													},
													"id": schema.StringAttribute{
														Description: "Host ID",
														Required:    false,
														Optional:    true,
														PlanModifiers: []planmodifier.String{
															stringplanmodifier.UseStateForUnknown(), // Avoid drift
														},
														Computed: true,
													},
												},
											},
										},
										"site": schema.SetNestedAttribute{
											Description: "Sites defined in your account",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.Set{
												setplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"name": schema.StringAttribute{
														Description: "Site Name",
														Required:    false,
														Optional:    true,
														Validators: []validator.String{
															stringvalidator.ConflictsWith(path.Expressions{
																path.MatchRelative().AtParent().AtName("id"),
															}...),
														},
														PlanModifiers: []planmodifier.String{
															stringplanmodifier.UseStateForUnknown(), // Avoid drift
														},
														Computed: true,
													},
													"id": schema.StringAttribute{
														Description: "Site ID",
														Required:    false,
														Optional:    true,
														PlanModifiers: []planmodifier.String{
															stringplanmodifier.UseStateForUnknown(), // Avoid drift
														},
														Computed: true,
													},
												},
											},
										},
										"subnet": schema.ListAttribute{
											ElementType: types.StringType,
											Description: "Subnet traffic matching criteria. Logical ‘OR’ is applied within the criteria set. Logical ‘AND’ is applied between criteria sets.",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.List{
												listplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Validators: []validator.List{
												listvalidator.SizeAtLeast(1),
											},
										},
										"ip_range": schema.ListNestedAttribute{
											Description: "IP range traffic matching criteria. Logical ‘OR’ is applied within the criteria set. Logical ‘AND’ is applied between criteria sets.",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.List{
												listplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Validators: []validator.List{
												listvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"from": schema.StringAttribute{
														Description: "From IP Range Name",
														Required:    true,
														Optional:    false,
													},
													"to": schema.StringAttribute{
														Description: "To IP Range ID",
														Required:    true,
														Optional:    false,
													},
												},
											},
										},
										"global_ip_range": schema.SetNestedAttribute{
											Description: "Global IP Range",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.Set{
												setplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"name": schema.StringAttribute{
														Description: "Global IP Range Name",
														Required:    false,
														Optional:    true,
														Validators: []validator.String{
															stringvalidator.ConflictsWith(path.Expressions{
																path.MatchRelative().AtParent().AtName("id"),
															}...),
														},
														PlanModifiers: []planmodifier.String{
															stringplanmodifier.UseStateForUnknown(), // Avoid drift
														},
														Computed: true,
													},
													"id": schema.StringAttribute{
														Description: "Global IP Range ID",
														Required:    false,
														Optional:    true,
														PlanModifiers: []planmodifier.String{
															stringplanmodifier.UseStateForUnknown(), // Avoid drift
														},
														Computed: true,
													},
												},
											},
										},
										"network_interface": schema.SetNestedAttribute{
											Description: "Network range defined for a site",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.Set{
												setplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"name": schema.StringAttribute{
														Description: "Network Interface Name",
														Required:    false,
														Optional:    true,
														PlanModifiers: []planmodifier.String{
															stringplanmodifier.UseStateForUnknown(), // Avoid drift
														},
														Computed: true,
													},
													"id": schema.StringAttribute{
														Description: "Network Interface ID",
														Required:    false,
														Optional:    true,
														PlanModifiers: []planmodifier.String{
															stringplanmodifier.UseStateForUnknown(), // Avoid drift
														},
														Computed: true,
													},
												},
											},
										},
										"site_network_subnet": schema.SetNestedAttribute{
											Required: false,
											Optional: true,
											PlanModifiers: []planmodifier.Set{
												setplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"name": schema.StringAttribute{
														Description: "Site Network Subnet Name",
														Required:    false,
														Optional:    true,
														Validators: []validator.String{
															stringvalidator.ConflictsWith(path.Expressions{
																path.MatchRelative().AtParent().AtName("id"),
															}...),
														},
														PlanModifiers: []planmodifier.String{
															stringplanmodifier.UseStateForUnknown(), // Avoid drift
														},
														Computed: true,
													},
													"id": schema.StringAttribute{
														Description: "Site Network Subnet ID",
														Required:    false,
														Optional:    true,
														PlanModifiers: []planmodifier.String{
															stringplanmodifier.UseStateForUnknown(), // Avoid drift
														},
														Computed: true,
													},
												},
											},
										},
										"floating_subnet": schema.SetNestedAttribute{
											Description: "Floating Subnet defined for a site",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.Set{
												setplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"name": schema.StringAttribute{
														Description: "Floating Subnet Name",
														Required:    false,
														Optional:    true,
														Validators: []validator.String{
															stringvalidator.ConflictsWith(path.Expressions{
																path.MatchRelative().AtParent().AtName("id"),
															}...),
														},
														PlanModifiers: []planmodifier.String{
															stringplanmodifier.UseStateForUnknown(), // Avoid drift
														},
														Computed: true,
													},
													"id": schema.StringAttribute{
														Description: "Floating Subnet ID",
														Required:    false,
														Optional:    true,
														PlanModifiers: []planmodifier.String{
															stringplanmodifier.UseStateForUnknown(), // Avoid drift
														},
														Computed: true,
													},
												},
											},
										},
										"user": schema.SetNestedAttribute{
											Description: "User defined for your account",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.Set{
												setplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"name": schema.StringAttribute{
														Description: "User Name",
														Required:    false,
														Optional:    true,
														Validators: []validator.String{
															stringvalidator.ConflictsWith(path.Expressions{
																path.MatchRelative().AtParent().AtName("id"),
															}...),
														},
														PlanModifiers: []planmodifier.String{
															stringplanmodifier.UseStateForUnknown(), // Avoid drift
														},
														Computed: true,
													},
													"id": schema.StringAttribute{
														Description: "User ID",
														Required:    false,
														Optional:    true,
														PlanModifiers: []planmodifier.String{
															stringplanmodifier.UseStateForUnknown(), // Avoid drift
														},
														Computed: true,
													},
												},
											},
										},
										"users_group": schema.SetNestedAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.Set{
												setplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"name": schema.StringAttribute{
														Description: "Users Group Name",
														Required:    false,
														Optional:    true,
														Validators: []validator.String{
															stringvalidator.ConflictsWith(path.Expressions{
																path.MatchRelative().AtParent().AtName("id"),
															}...),
														},
														PlanModifiers: []planmodifier.String{
															stringplanmodifier.UseStateForUnknown(), // Avoid drift
														},
														Computed: true,
													},
													"id": schema.StringAttribute{
														Description: "Users Group ID",
														Required:    false,
														Optional:    true,
														PlanModifiers: []planmodifier.String{
															stringplanmodifier.UseStateForUnknown(), // Avoid drift
														},
														Computed: true,
													},
												},
											},
										},
										"group": schema.SetNestedAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.Set{
												setplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"name": schema.StringAttribute{
														Description: "Group Name",
														Required:    false,
														Optional:    true,
														Validators: []validator.String{
															stringvalidator.ConflictsWith(path.Expressions{
																path.MatchRelative().AtParent().AtName("id"),
															}...),
														},
														PlanModifiers: []planmodifier.String{
															stringplanmodifier.UseStateForUnknown(), // Avoid drift
														},
														Computed: true,
													},
													"id": schema.StringAttribute{
														Description: "Group ID",
														Required:    false,
														Optional:    true,
														PlanModifiers: []planmodifier.String{
															stringplanmodifier.UseStateForUnknown(), // Avoid drift
														},
														Computed: true,
													},
												},
											},
										},
										"system_group": schema.SetNestedAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.Set{
												setplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"name": schema.StringAttribute{
														Description: "System Group Name",
														Required:    false,
														Optional:    true,
														Validators: []validator.String{
															stringvalidator.ConflictsWith(path.Expressions{
																path.MatchRelative().AtParent().AtName("id"),
															}...),
														},
														PlanModifiers: []planmodifier.String{
															stringplanmodifier.UseStateForUnknown(), // Avoid drift
														},
														Computed: true,
													},
													"id": schema.StringAttribute{
														Description: "System Group ID",
														Required:    false,
														Optional:    true,
														PlanModifiers: []planmodifier.String{
															stringplanmodifier.UseStateForUnknown(), // Avoid drift
														},
														Computed: true,
													},
												},
											},
										},
									},
								},
								"country": schema.SetNestedAttribute{
									Description: "Source country traffic matching criteria. Logical ‘OR’ is applied within the criteria set. Logical ‘AND’ is applied between criteria sets.",
									Required:    false,
									Optional:    true,
									PlanModifiers: []planmodifier.Set{
										setplanmodifier.UseStateForUnknown(), // Avoid drift
									},
									Validators: []validator.Set{
										setvalidator.SizeAtLeast(1),
									},
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"name": schema.StringAttribute{
												Description: "Country Name",
												Required:    false,
												Optional:    true,
												Validators: []validator.String{
													stringvalidator.ConflictsWith(path.Expressions{
														path.MatchRelative().AtParent().AtName("id"),
													}...),
												},
												PlanModifiers: []planmodifier.String{
													stringplanmodifier.UseStateForUnknown(), // Avoid drift
												},
												Computed: true,
											},
											"id": schema.StringAttribute{
												Description: "Country ID",
												Required:    false,
												Optional:    true,
												PlanModifiers: []planmodifier.String{
													stringplanmodifier.UseStateForUnknown(), // Avoid drift
												},
												Computed: true,
											},
										},
									},
								},
								"device": schema.SetNestedAttribute{
									Description: "Source Device Profile traffic matching criteria. Logical 'OR' is applied within the criteria set. Logical 'AND' is applied between criteria sets.",
									Required:    false,
									Optional:    true,
									PlanModifiers: []planmodifier.Set{
										setplanmodifier.UseStateForUnknown(), // Avoid drift
									},
									Validators: []validator.Set{
										setvalidator.SizeAtLeast(1),
									},
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"name": schema.StringAttribute{
												Description: "Device Name",
												Required:    false,
												Optional:    true,
												Validators: []validator.String{
													stringvalidator.ConflictsWith(path.Expressions{
														path.MatchRelative().AtParent().AtName("id"),
													}...),
												},
												PlanModifiers: []planmodifier.String{
													stringplanmodifier.UseStateForUnknown(), // Avoid drift
												},
												Computed: true,
											},
											"id": schema.StringAttribute{
												Description: "Device ID",
												Required:    false,
												Optional:    true,
												PlanModifiers: []planmodifier.String{
													stringplanmodifier.UseStateForUnknown(), // Avoid drift
												},
												Computed: true,
											},
										},
									},
								},
								"device_attributes": schema.SingleNestedAttribute{
									Description: "Device attributes matching criteria for the exception.",
									Required:    false,
									Optional:    true,
									PlanModifiers: []planmodifier.Object{
										objectplanmodifier.UseStateForUnknown(), // Avoid drift
									},
									Attributes: map[string]schema.Attribute{
										"category": schema.ListAttribute{
											ElementType: types.StringType,
											Description: "Device category matching criteria for the exception.",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.List{
												listplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Validators: []validator.List{
												listvalidator.SizeAtLeast(1),
											},
										},
										"type": schema.ListAttribute{
											ElementType: types.StringType,
											Description: "Device type matching criteria for the exception.",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.List{
												listplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Validators: []validator.List{
												listvalidator.SizeAtLeast(1),
											},
										},
										"model": schema.ListAttribute{
											ElementType: types.StringType,
											Description: "Device model matching criteria for the exception.",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.List{
												listplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Validators: []validator.List{
												listvalidator.SizeAtLeast(1),
											},
										},
										"manufacturer": schema.ListAttribute{
											ElementType: types.StringType,
											Description: "Device manufacturer matching criteria for the exception.",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.List{
												listplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Validators: []validator.List{
												listvalidator.SizeAtLeast(1),
											},
										},
										"os": schema.ListAttribute{
											ElementType: types.StringType,
											Description: "Device OS matching criteria for the exception.",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.List{
												listplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Validators: []validator.List{
												listvalidator.SizeAtLeast(1),
											},
										},
										"os_version": schema.ListAttribute{
											ElementType: types.StringType,
											Description: "Device OS version matching criteria for the exception.",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.List{
												listplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Validators: []validator.List{
												listvalidator.SizeAtLeast(1),
											},
										},
									},
								},
								"device_os": schema.ListAttribute{
									ElementType: types.StringType,
									Description: "Source device OS matching criteria for the exception. (https://api.catonetworks.com/documentation/#definition-OperatingSystem)",
									Optional:    true,
									Required:    false,
									PlanModifiers: []planmodifier.List{
										listplanmodifier.UseStateForUnknown(), // Avoid drift
									},
									Validators: []validator.List{
										listvalidator.ValueStringsAre(stringvalidator.OneOf("ANDROID", "EMBEDDED", "IOS", "LINUX", "MACOS", "WINDOWS")),
										listvalidator.SizeAtLeast(1),
									},
								},
								"destination": schema.SingleNestedAttribute{
									Description: "Destination service matching criteria for the exception.",
									Required:    true,
									Optional:    false,
									Attributes: map[string]schema.Attribute{
										"application": schema.SetNestedAttribute{
											Description: "Applications for the rule (pre-defined)",
											Required:    false,
											Optional:    true,
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"name": schema.StringAttribute{
														Description: "Application Name",
														Required:    false,
														Optional:    true,
														Validators: []validator.String{
															stringvalidator.ConflictsWith(path.Expressions{
																path.MatchRelative().AtParent().AtName("id"),
															}...),
														},
														PlanModifiers: []planmodifier.String{
															stringplanmodifier.UseStateForUnknown(), // Avoid drift
														},
														Computed: true,
													},
													"id": schema.StringAttribute{
														Description: "Application ID",
														Required:    false,
														Optional:    true,
														PlanModifiers: []planmodifier.String{
															stringplanmodifier.UseStateForUnknown(), // Avoid drift
														},
														Computed: true,
													},
												},
											},
										},
										"custom_app": schema.SetNestedAttribute{
											Description: "Custom (user-defined) applications",
											Required:    false,
											Optional:    true,
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"name": schema.StringAttribute{
														Description: "Custom Application Name",
														Required:    false,
														Optional:    true,
														Validators: []validator.String{
															stringvalidator.ConflictsWith(path.Expressions{
																path.MatchRelative().AtParent().AtName("id"),
															}...),
														},
														PlanModifiers: []planmodifier.String{
															stringplanmodifier.UseStateForUnknown(), // Avoid drift
														},
														Computed: true,
													},
													"id": schema.StringAttribute{
														Description: "Custom Application ID",
														Required:    false,
														Optional:    true,
														PlanModifiers: []planmodifier.String{
															stringplanmodifier.UseStateForUnknown(), // Avoid drift
														},
														Computed: true,
													},
												},
											},
										},
										"app_category": schema.SetNestedAttribute{
											Description: "Cato category of applications which are dynamically updated by Cato",
											Required:    false,
											Optional:    true,
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"name": schema.StringAttribute{
														Description: "Application Category Name",
														Required:    false,
														Optional:    true,
														Validators: []validator.String{
															stringvalidator.OneOf("Advertisements", "AI Media Generators", "Alcohol and Tobacco", "Anonymizers", "Authentication Services", "Beauty", "Botnets", "Business Information", "Business Operations AI", "Business Systems", "CDN", "Chat and IM", "Cheating", "Code Assistants", "Compromised", "Computers and Technology", "Conversational AI", "Criminal Activity", "Cults", "Database", "DNS over HTTPS", "Education", "Email", "Entertainment", "ERP And CRM", "File Sharing", "Finance", "Gambling", "Games", "General", "Generative AI Tools", "Government", "Greeting Cards", "Hacking", "Health and Medicine", "Healthcare AI", "Hiring", "Illegal Drugs", "Industrial Protocols", "Information Security", "Internet Conferencing", "Keyloggers", "Leisure and Recreation", "Malware", "Media Streams", "Military", "Network Protocol", "Network Utilities", "News", "Nudity", "Office Programs And Services", "Online Storage", "P2P", "Parked domains", "PDF Converters", "Personal Sites", "Phishing", "Politics", "Porn", "Productivity", "Questionable", "Real Estate", "Religion", "Remote Access", "Search Engines and Portals", "Sex education", "Shopping", "Social", "Software Downloads", "Software Updates", "SPAM", "Sports", "Spyware", "Tasteless", "Translation", "Travel AI Assistance", "Travel", "Uncategorized", "Undefined", "Vehicles", "Violence and Hate", "Voip Video", "Weapons", "Web Hosting", "Web Posting", "Writing Assistants"),
															stringvalidator.ConflictsWith(path.Expressions{
																path.MatchRelative().AtParent().AtName("id"),
															}...),
														},
														PlanModifiers: []planmodifier.String{
															stringplanmodifier.UseStateForUnknown(), // Avoid drift
														},
														Computed: true,
													},
													"id": schema.StringAttribute{
														Description: "Application Category ID",
														Required:    false,
														Optional:    true,
														Validators: []validator.String{
															stringvalidator.OneOf("advertisements", "ai_tools", "anonymizers", "authentication_services", "beauty", "botnets", "business_systems", "business", "cdn", "chat_and_im", "cheating", "computers_and_technology", "criminal_activity", "cults", "database", "dns_over_https", "drugs", "education", "email", "entertainment", "erp_and_crm", "file_sharing", "finance", "food_drinks_tobacco", "gambling", "games", "gen_ai_business_operations", "gen_ai_code_assistants", "gen_ai_conversational_ai", "gen_ai_healthcare", "gen_ai_media_generators", "gen_ai_productivity", "gen_ai_travel_assistance", "gen_ai_writing_assistants", "general", "government", "greeting_cards", "hacking", "health_and_medicine", "hiring", "information_security", "internet_conferencing", "keyloggers", "leisure_and_recreation", "media_streams", "military", "network_protocol", "network_utilities", "news", "nudity", "office_programs_and_services", "online_storage", "ot_protocols", "p2p", "parked_domains", "pdf_converters", "personal_sites", "politics", "porn", "questionable", "real_estate", "religion", "remote_access", "search_engines_and_portals", "sex_education", "shopping", "social", "software_downloads", "software_updates", "spam", "sports", "spyware", "suspected_malware", "suspected_phishing", "suspected_unwanted", "tasteless", "translation", "travel", "uncategorized", "undefined", "vehicles", "violence", "voip_video", "weapons", "web_hosting", "web_posting"),
														},
														PlanModifiers: []planmodifier.String{
															stringplanmodifier.UseStateForUnknown(), // Avoid drift
														},
														Computed: true,
													},
												},
											},
										},
										"custom_category": schema.SetNestedAttribute{
											Description: "Custom Categories – Groups of objects such as predefined and custom applications, predefined and custom services, domains, FQDNs etc.",
											Required:    false,
											Optional:    true,
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											PlanModifiers: []planmodifier.Set{
												setplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"name": schema.StringAttribute{
														Description: "Custom Category Name",
														Required:    false,
														Optional:    true,
														Validators: []validator.String{
															stringvalidator.ConflictsWith(path.Expressions{
																path.MatchRelative().AtParent().AtName("id"),
															}...),
														},
														PlanModifiers: []planmodifier.String{
															stringplanmodifier.UseStateForUnknown(), // Avoid drift
														},
														Computed: true,
													},
													"id": schema.StringAttribute{
														Description: "Custom Category ID",
														Required:    false,
														Optional:    true,
														PlanModifiers: []planmodifier.String{
															stringplanmodifier.UseStateForUnknown(), // Avoid drift
														},
														Computed: true,
													},
												},
											},
										},
										"sanctioned_apps_category": schema.SetNestedAttribute{
											Description: "Sanctioned Cloud Applications - apps that are approved and generally represent an understood and acceptable level of risk in your organization.",
											Required:    false,
											Optional:    true,
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"name": schema.StringAttribute{
														Description: "Sanctioned Apps Category Name",
														Required:    false,
														Optional:    true,
														Validators: []validator.String{
															stringvalidator.ConflictsWith(path.Expressions{
																path.MatchRelative().AtParent().AtName("id"),
															}...),
														},
														PlanModifiers: []planmodifier.String{
															stringplanmodifier.UseStateForUnknown(), // Avoid drift
														},
														Computed: true,
													},
													"id": schema.StringAttribute{
														Description: "Sanctioned Apps Category ID",
														Required:    false,
														Optional:    true,
														PlanModifiers: []planmodifier.String{
															stringplanmodifier.UseStateForUnknown(), // Avoid drift
														},
														Computed: true,
													},
												},
											},
										},
										"country": schema.SetNestedAttribute{
											Description: "Source country traffic matching criteria. Logical ‘OR’ is applied within the criteria set. Logical ‘AND’ is applied between criteria sets.",
											Required:    false,
											Optional:    true,
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"name": schema.StringAttribute{
														Description: "Country Name",
														Required:    false,
														Optional:    true,
														Validators: []validator.String{
															stringvalidator.ConflictsWith(path.Expressions{
																path.MatchRelative().AtParent().AtName("id"),
															}...),
														},
														PlanModifiers: []planmodifier.String{
															stringplanmodifier.UseStateForUnknown(), // Avoid drift
														},
														Computed: true,
													},
													"id": schema.StringAttribute{
														Description: "Country ID",
														Required:    false,
														Optional:    true,
														PlanModifiers: []planmodifier.String{
															stringplanmodifier.UseStateForUnknown(), // Avoid drift
														},
														Computed: true,
													},
												},
											},
										},
										"domain": schema.ListAttribute{
											ElementType: types.StringType,
											Description: "A Second-Level Domain (SLD). It matches all Top-Level Domains (TLD), and subdomains that include the Domain. Example: example.com.",
											Required:    false,
											Optional:    true,
											Validators: []validator.List{
												listvalidator.SizeAtLeast(1),
											},
										},
										"fqdn": schema.ListAttribute{
											ElementType: types.StringType,
											Description: "An exact match of the fully qualified domain (FQDN). Example: www.my.example.com.",
											Required:    false,
											Optional:    true,
											Validators: []validator.List{
												listvalidator.SizeAtLeast(1),
											},
										},
										"ip": schema.ListAttribute{
											ElementType: types.StringType,
											Description: "IPv4 addresses",
											Required:    false,
											Optional:    true,
											Validators: []validator.List{
												listvalidator.SizeAtLeast(1),
											},
										},
										"subnet": schema.ListAttribute{
											ElementType: types.StringType,
											Description: "Network subnets in CIDR notation",
											Required:    false,
											Optional:    true,
											Validators: []validator.List{
												listvalidator.SizeAtLeast(1),
											},
										},
										"ip_range": schema.ListNestedAttribute{
											Description: "A range of IPs. Every IP within the range will be matched",
											Required:    false,
											Optional:    true,
											Validators: []validator.List{
												listvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"from": schema.StringAttribute{
														Description: "IP Range Name",
														Required:    true,
														Optional:    false,
													},
													"to": schema.StringAttribute{
														Description: "IP Range ID",
														Required:    true,
														Optional:    false,
													},
												},
											},
										},
										"global_ip_range": schema.SetNestedAttribute{
											Description: "Globally defined IP range, IP and subnet objects.",
											Required:    false,
											Optional:    true,
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"name": schema.StringAttribute{
														Description: "Global IP Range Name",
														Required:    false,
														Optional:    true,
														Validators: []validator.String{
															stringvalidator.ConflictsWith(path.Expressions{
																path.MatchRelative().AtParent().AtName("id"),
															}...),
														},
														PlanModifiers: []planmodifier.String{
															stringplanmodifier.UseStateForUnknown(), // Avoid drift
														},
														Computed: true,
													},
													"id": schema.StringAttribute{
														Description: "Global IP ID",
														Required:    false,
														Optional:    true,
														PlanModifiers: []planmodifier.String{
															stringplanmodifier.UseStateForUnknown(), // Avoid drift
														},
														Computed: true,
													},
												},
											},
										},
										"remote_asn": schema.ListAttribute{
											ElementType: types.StringType,
											Description: "",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.List{
												listplanmodifier.UseStateForUnknown(), // Avoid drift
											},
										},
									},
								},
								"service": schema.SingleNestedAttribute{
									Description: "Destination service traffic matching criteria. Logical 'OR' is applied within the criteria set. Logical 'AND' is applied between criteria sets.",
									Required:    false,
									Optional:    true,
									PlanModifiers: []planmodifier.Object{
										objectplanmodifier.UseStateForUnknown(), // Avoid drift
									},
									Computed: true,
									Attributes: map[string]schema.Attribute{
										"standard": schema.SetNestedAttribute{
											Description: "Standard Service to which this Internet Firewall rule applies",
											Required:    false,
											Optional:    true,
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"name": schema.StringAttribute{
														Description: "Service Standard Name",
														Required:    false,
														Optional:    true,
														Validators: []validator.String{
															stringvalidator.ConflictsWith(path.Expressions{
																path.MatchRelative().AtParent().AtName("id"),
															}...),
														},
														PlanModifiers: []planmodifier.String{
															stringplanmodifier.UseStateForUnknown(), // Avoid drift
														},
														Computed: true,
													},
													"id": schema.StringAttribute{
														Description: "Service Standard ID",
														Required:    false,
														Optional:    true,
														PlanModifiers: []planmodifier.String{
															stringplanmodifier.UseStateForUnknown(), // Avoid drift
														},
														Computed: true,
													},
												},
											},
										},
										"custom": schema.ListNestedAttribute{
											Description: "Custom Service defined by a combination of L4 ports and an IP Protocol",
											Required:    false,
											Optional:    true,
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"port": schema.ListAttribute{
														ElementType: types.StringType,
														Description: "List of TCP/UDP port",
														Optional:    true,
														Required:    false,
													},
													"port_range": schema.SingleNestedAttribute{
														Description: "TCP/UDP port ranges",
														Required:    false,
														Optional:    true,
														Attributes: map[string]schema.Attribute{
															"from": schema.StringAttribute{
																Description: "",
																Required:    true,
																Optional:    false,
															},
															"to": schema.StringAttribute{
																Description: "",
																Required:    true,
																Optional:    false,
															},
														},
													},
													"protocol": schema.StringAttribute{
														Description: "IP Protocol (https://api.catonetworks.com/documentation/#definition-IpProtocol)",
														Required:    false,
														Optional:    true,
													},
												},
											},
										},
									},
								},
								"connection_origin": schema.StringAttribute{
									Description: "Connection origin matching criteria for the exception. (https://api.catonetworks.com/documentation/#definition-ConnectionOriginEnum)",
									Optional:    true,
									Required:    false,
									PlanModifiers: []planmodifier.String{
										stringplanmodifier.UseStateForUnknown(), // Avoid drift
									},
									Validators: []validator.String{
										stringvalidator.OneOf("ANY", "REMOTE", "SITE"),
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

func (r *internetFwRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoClientData)
}

func (r *internetFwRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("rule").AtName("id"), req, resp)
}

func (r *internetFwRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	var plan InternetFirewallRule
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := hydrateIfwRuleApi(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Warn(ctx, "TFLOG_WARN_IFW_input.create", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(input.create),
	})

	//creating new rule
	createRuleResponse, err := r.client.catov2.PolicyInternetFirewallAddRule(ctx, input.create, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyInternetFirewallAddRule error",
			err.Error(),
		)
		return
	}

	tflog.Warn(ctx, "TFLOG_WARN_IFW_createRuleResponse", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(createRuleResponse),
	})

	// check for errors
	if createRuleResponse.Policy.InternetFirewall.AddRule.Status != "SUCCESS" {
		for _, item := range createRuleResponse.Policy.InternetFirewall.AddRule.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Creating Resource",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	//publishing new rule
	tflog.Info(ctx, "publishing new rule")
	publishDataIfEnabled := &cato_models.PolicyPublishRevisionInput{}
	_, err = r.client.catov2.PolicyInternetFirewallPublishPolicyRevision(ctx, &cato_models.InternetFirewallPolicyMutationInput{}, publishDataIfEnabled, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyInternetFirewallPublishPolicyRevision error",
			err.Error(),
		)
		return
	}

	// Read rule and hydrate response to state
	queryIfwPolicy := &cato_models.InternetFirewallPolicyInput{}
	body, err := r.client.catov2.PolicyInternetFirewall(ctx, queryIfwPolicy, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyInternetFirewall error",
			err.Error(),
		)
		return
	}

	ruleList := body.GetPolicy().InternetFirewall.Policy.GetRules()
	currentRule := &cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_Rule{}
	// Get current rule from response by ID
	for _, ruleListItem := range ruleList {
		if ruleListItem.GetRule().ID == createRuleResponse.GetPolicy().GetInternetFirewall().GetAddRule().Rule.GetRule().ID {
			currentRule = ruleListItem.GetRule()
			break
		}
	}

	tflog.Warn(ctx, "TFLOG_WARN_IFW_createRule.readResponse", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(currentRule),
	})

	// Hydrate ruleInput from api response
	ruleInputRead := hydrateIfwRuleState(ctx, plan, currentRule)
	ruleInputRead.ID = types.StringValue(currentRule.ID)

	// Handle exceptions correlation manually to preserve plan structure
	if !plan.Rule.IsNull() && !plan.Rule.IsUnknown() {
		planRule := Policy_Policy_InternetFirewall_Policy_Rules_Rule{}
		diags = plan.Rule.As(ctx, &planRule, basetypes.ObjectAsOptions{})
		if !diags.HasError() && !planRule.Exceptions.IsNull() && !planRule.Exceptions.IsUnknown() {
			// Correlate exceptions between plan and hydrated response
			correlatedExceptions := correlateIfwExceptions(ctx, planRule.Exceptions, ruleInputRead.Exceptions)
			if correlatedExceptions != nil {
				ruleInputRead.Exceptions = *correlatedExceptions
			}
		}
	}

	ruleObject, diags := types.ObjectValueFrom(ctx, InternetFirewallRuleRuleAttrTypes, ruleInputRead)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Assign ruleObject to state
	plan.Rule = ruleObject

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(diags...)
}

func (r *internetFwRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

	var state InternetFirewallRule
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	queryIfwPolicy := &cato_models.InternetFirewallPolicyInput{}
	body, err := r.client.catov2.PolicyInternetFirewall(ctx, queryIfwPolicy, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyInternetFirewall error",
			err.Error(),
		)
		return
	}

	//retrieve rule ID
	rule := Policy_Policy_InternetFirewall_Policy_Rules_Rule{}
	diags = state.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ruleList := body.GetPolicy().InternetFirewall.Policy.GetRules()
	ruleExist := false
	currentRule := &cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_Rule{}
	for _, ruleListItem := range ruleList {
		if ruleListItem.GetRule().ID == rule.ID.ValueString() {
			ruleExist = true
			currentRule = ruleListItem.GetRule()

			// Need to refresh STATE
			resp.State.SetAttribute(
				ctx,
				path.Root("rule").AtName("id"),
				ruleListItem.GetRule().ID)
		}
	}

	// remove resource if it doesn't exist anymore
	if !ruleExist {
		tflog.Warn(ctx, "internet firewall rule not found, resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	tflog.Warn(ctx, "TFLOG_WARN_IFW.readResponse", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(currentRule),
	})

	ruleInput := hydrateIfwRuleState(ctx, state, currentRule)
	diags = resp.State.SetAttribute(ctx, path.Root("rule"), ruleInput)
	resp.Diagnostics.Append(diags...)

	// Check if position is set in state, if not default to LAST_IN_POLICY
	// Hard coding LAST_IN_POLICY position as the API does not return any value and
	// hardcoding position supports the use case of bulk rule import/export
	// getting around state changes for the position field
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

}

func (r *internetFwRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	var plan InternetFirewallRule
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := hydrateIfwRuleApi(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// setting input for moving rule
	inputMoveRule := cato_models.PolicyMoveRuleInput{}

	//setting at (to move rule)
	if !plan.At.IsNull() {
		inputMoveRule.To = &cato_models.PolicyRulePositionInput{}
		positionInput := PolicyRulePositionInput{}
		diags = plan.At.As(ctx, &positionInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)

		inputMoveRule.To.Position = (*cato_models.PolicyRulePositionEnum)(positionInput.Position.ValueStringPointer())
		inputMoveRule.To.Ref = positionInput.Ref.ValueStringPointer()
	}

	// // setting rule
	ruleInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule{}
	diags = plan.Rule.As(ctx, &ruleInput, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)

	// settings other rule attributes
	inputMoveRule.ID = *ruleInput.ID.ValueStringPointer()
	input.update.ID = *ruleInput.ID.ValueStringPointer()

	//move rule
	moveRule, err := r.client.catov2.PolicyInternetFirewallMoveRule(ctx, &cato_models.InternetFirewallPolicyMutationInput{}, inputMoveRule, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyInternetFirewallMoveRule error",
			err.Error(),
		)
		return
	}

	// check for errors
	if moveRule.Policy.InternetFirewall.MoveRule.Status != "SUCCESS" {
		for _, item := range moveRule.Policy.InternetFirewall.MoveRule.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Moving Rule Resource",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	tflog.Warn(ctx, "TFLOG_WARN_IFW_input.update", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(input.update),
	})

	//Update new rule
	updateRuleResponse, err := r.client.catov2.PolicyInternetFirewallUpdateRule(ctx, &cato_models.InternetFirewallPolicyMutationInput{}, input.update, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyInternetFirewallUpdateRule error",
			err.Error(),
		)
		return
	}

	// check for errors
	if updateRuleResponse.Policy.InternetFirewall.UpdateRule.Status != "SUCCESS" {
		for _, item := range updateRuleResponse.Policy.InternetFirewall.UpdateRule.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Update Resource",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	//publishing new rule
	tflog.Info(ctx, "publishing new rule")
	publishDataIfEnabled := &cato_models.PolicyPublishRevisionInput{}
	_, err = r.client.catov2.PolicyInternetFirewallPublishPolicyRevision(ctx, &cato_models.InternetFirewallPolicyMutationInput{}, publishDataIfEnabled, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyInternetFirewallPublishPolicyRevision error",
			err.Error(),
		)
		return
	}

	// Read rule and hydrate response to state
	queryIfwPolicy := &cato_models.InternetFirewallPolicyInput{}
	body, err := r.client.catov2.PolicyInternetFirewall(ctx, queryIfwPolicy, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyInternetFirewall error",
			err.Error(),
		)
		return
	}

	ruleList := body.GetPolicy().InternetFirewall.Policy.GetRules()
	currentRule := &cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_Rule{}
	// Get current rule from response by ID
	for _, ruleListItem := range ruleList {
		if ruleListItem.GetRule().ID == updateRuleResponse.GetPolicy().GetInternetFirewall().GetUpdateRule().Rule.GetRule().ID {
			currentRule = ruleListItem.GetRule()
			break
		}
	}
	tflog.Warn(ctx, "TFLOG_WARN_IFW_createRule.readResponse", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(currentRule),
	})

	// Hydrate ruleInput from api respoonse
	ruleInputRead := hydrateIfwRuleState(ctx, plan, currentRule)
	ruleInputRead.ID = types.StringValue(currentRule.ID)

	// Handle exceptions correlation manually to preserve plan structure
	if !plan.Rule.IsNull() && !plan.Rule.IsUnknown() {
		planRule := Policy_Policy_InternetFirewall_Policy_Rules_Rule{}
		diags = plan.Rule.As(ctx, &planRule, basetypes.ObjectAsOptions{})
		if !diags.HasError() && !planRule.Exceptions.IsNull() && !planRule.Exceptions.IsUnknown() {
			// Correlate exceptions between plan and hydrated response
			correlatedExceptions := correlateIfwExceptions(ctx, planRule.Exceptions, ruleInputRead.Exceptions)
			if correlatedExceptions != nil {
				ruleInputRead.Exceptions = *correlatedExceptions
			}
		}
	}

	ruleObject, diags := types.ObjectValueFrom(ctx, InternetFirewallRuleRuleAttrTypes, ruleInputRead)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Assign ruleObject to state
	plan.Rule = ruleObject

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *internetFwRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	var state InternetFirewallRule
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	//retrieve rule ID
	rule := Policy_Policy_InternetFirewall_Policy_Rules_Rule{}
	diags = state.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	removeMutations := &cato_models.InternetFirewallPolicyMutationInput{}
	removeRule := cato_models.InternetFirewallRemoveRuleInput{
		ID: rule.ID.ValueString(),
	}
	tflog.Debug(ctx, "internet_fw_policy delete", map[string]interface{}{
		"query": utils.InterfaceToJSONString(removeMutations),
		"input": utils.InterfaceToJSONString(removeRule),
	})

	_, err := r.client.catov2.PolicyInternetFirewallRemoveRule(ctx, removeMutations, removeRule, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to connect or request the Catov2 API",
			err.Error(),
		)
		return
	}

	publishDataIfEnabled := &cato_models.PolicyPublishRevisionInput{}
	_, err = r.client.catov2.PolicyInternetFirewallPublishPolicyRevision(ctx, &cato_models.InternetFirewallPolicyMutationInput{}, publishDataIfEnabled, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API Delete/PolicyInternetFirewallPublishPolicyRevision error",
			err.Error(),
		)
		return
	}
}

// IFW Rule Schema validation fuctions

// Rule -> Alert Valildator
type ruleAlertValidator struct{}

func (v ruleAlertValidator) Description(ctx context.Context) string {
	return "If 'alert' is provided, both 'enabled' and 'frequency' must also be set, and must specify values for mailing_list, subscription_group, or web_hook."
}

func (v ruleAlertValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v ruleAlertValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	alertObj := req.ConfigValue
	alertMap := alertObj.Attributes()

	enabled, hasEnabled := alertMap["enabled"]
	frequency, hasFrequency := alertMap["frequency"]

	if !hasEnabled || enabled.IsNull() || enabled.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("alert").AtName("enabled"),
			"'enabled' is required",
			"If 'alert' is provided, the 'enabled' field must also be set.",
		)
	}

	if !hasFrequency || frequency.IsNull() || frequency.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("alert").AtName("frequency"),
			"'frequency' is required",
			"If 'alert' is provided, the 'frequency' field must also be set.",
		)
	}
}
