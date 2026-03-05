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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
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
	_ resource.Resource                = &socketLanNetworkRuleResource{}
	_ resource.ResourceWithConfigure   = &socketLanNetworkRuleResource{}
	_ resource.ResourceWithImportState = &socketLanNetworkRuleResource{}
)

func NewSocketLanNetworkRuleResource() resource.Resource {
	return &socketLanNetworkRuleResource{}
}

type socketLanNetworkRuleResource struct {
	client *catoClientData
}

func (r *socketLanNetworkRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_socket_lan_network_rule"
}

func (r *socketLanNetworkRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_socket_lan_network_rule` resource contains the configuration parameters necessary to add a Socket LAN network rule. Documentation for the underlying API used in this resource can be found at [mutation.policy.socketLan.addRule()](https://api.catonetworks.com/documentation/#mutation-policy.socketLan.addRule).",
		Attributes: map[string]schema.Attribute{
			"at": schema.SingleNestedAttribute{
				Description: "Position of the rule in the policy",
				Required:    true,
				Optional:    false,
				Attributes: map[string]schema.Attribute{
					"position": schema.StringAttribute{
						Description: "Position relative to a policy, a section or another rule",
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
				Description: "Parameters for the Socket LAN network rule",
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
						Computed:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"index": schema.Int64Attribute{
						Description: "Rule Index",
						Computed:    true,
						Optional:    false,
					},
					"enabled": schema.BoolAttribute{
						Description: "Enable or disable the rule",
						Required:    true,
						Optional:    false,
					},
					"direction": schema.StringAttribute{
						Description: "Direction of the traffic (TO, BOTH)",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("TO", "BOTH"),
						},
					},
					"transport": schema.StringAttribute{
						Description: "Transport type (LAN, WAN)",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("LAN", "WAN"),
						},
					},
					"site": schema.SingleNestedAttribute{
						Description: "Site scope for the rule",
						Required:    true,
						Attributes: map[string]schema.Attribute{
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
						},
					},
					"source": schema.SingleNestedAttribute{
						Description: "Source traffic matching criteria",
						Required:    true,
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
							planmodifiers.SourceDestObjectModifier(),
						},
						Attributes: map[string]schema.Attribute{
							"vlan": schema.ListAttribute{
								ElementType: types.Int64Type,
								Description: "VLAN IDs",
								Required:    false,
								Optional:    true,
								Computed:    true,
								PlanModifiers: []planmodifier.List{
									listplanmodifier.UseStateForUnknown(),
								},
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
							"ip": schema.ListAttribute{
								ElementType: types.StringType,
								Description: "IP addresses",
								Required:    false,
								Optional:    true,
								Computed:    true,
								PlanModifiers: []planmodifier.List{
									listplanmodifier.UseStateForUnknown(),
								},
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
							"subnet": schema.ListAttribute{
								ElementType: types.StringType,
								Description: "Subnets in CIDR notation",
								Required:    false,
								Optional:    true,
								Computed:    true,
								PlanModifiers: []planmodifier.List{
									listplanmodifier.UseStateForUnknown(),
								},
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
							"ip_range": schema.ListNestedAttribute{
								Description: "IP ranges",
								Required:    false,
								Optional:    true,
								Computed:    true,
								PlanModifiers: []planmodifier.List{
									listplanmodifier.UseStateForUnknown(),
								},
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
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
							"host": schema.SetNestedAttribute{
								Description: "Hosts and servers defined for your account",
								Required:    false,
								Optional:    true,
								Computed:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(), // Avoid drift
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
							"group": schema.SetNestedAttribute{
								Description: "",
								Required:    false,
								Optional:    true,
								Computed:    true,
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
								Computed:    true,
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
							"network_interface": schema.SetNestedAttribute{
								Description: "Network range defined for a site",
								Required:    false,
								Optional:    true,
								Computed:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(), // Avoid drift
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Description: "Network Interface Name",
											Required:    false,
											Optional:    true,
											// Validators: []validator.String{
											// 	stringvalidator.ConflictsWith(path.Expressions{
											// 		path.MatchRelative().AtParent().AtName("id"),
											// 	}...),
											// },
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
										},
									},
								},
							},
							"global_ip_range": schema.SetNestedAttribute{
								Description: "Global IP range matching criteria for the exception.",
								Required:    false,
								Optional:    true,
								Computed:    true,
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
							"floating_subnet": schema.SetNestedAttribute{
								Description: "Floating Subnets (ie. Floating Ranges) are used to identify traffic exactly matched to the route advertised by BGP. They are not associated with a specific site. This is useful in scenarios such as active-standby high availability routed via BGP.",
								Required:    false,
								Optional:    true,
								Computed:    true,
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
							"site_network_subnet": schema.SetNestedAttribute{
								Description: "GlobalRange + InterfaceSubnet",
								Required:    false,
								Optional:    true,
								Computed:    true,
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(), // Avoid drift
								},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Description: "Site Natwork Subnet Name",
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
											Description: "Site Natwork Subnet ID",
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
					"destination": schema.SingleNestedAttribute{
						Description: "Destination traffic matching criteria",
						Required:    true,
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
							planmodifiers.SourceDestObjectModifier(),
						},
						Attributes: map[string]schema.Attribute{
							"vlan": schema.ListAttribute{
								ElementType: types.Int64Type,
								Description: "VLAN IDs",
								Required:    false,
								Optional:    true,
								Computed:    true,
								PlanModifiers: []planmodifier.List{
									listplanmodifier.UseStateForUnknown(),
								},
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
							"ip": schema.ListAttribute{
								ElementType: types.StringType,
								Description: "IP addresses",
								Required:    false,
								Optional:    true,
								Computed:    true,
								PlanModifiers: []planmodifier.List{
									listplanmodifier.UseStateForUnknown(),
								},
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
							"subnet": schema.ListAttribute{
								ElementType: types.StringType,
								Description: "Subnets in CIDR notation",
								Required:    false,
								Optional:    true,
								Computed:    true,
								PlanModifiers: []planmodifier.List{
									listplanmodifier.UseStateForUnknown(),
								},
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
							"ip_range": schema.ListNestedAttribute{
								Description: "IP ranges",
								Required:    false,
								Optional:    true,
								Computed:    true,
								PlanModifiers: []planmodifier.List{
									listplanmodifier.UseStateForUnknown(),
								},
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
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
							"host": schema.SetNestedAttribute{
								Description: "Hosts and servers defined for your account",
								Required:    false,
								Optional:    true,
								Computed:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(), // Avoid drift
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
							"group": schema.SetNestedAttribute{
								Description: "",
								Required:    false,
								Optional:    true,
								Computed:    true,
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
								Computed:    true,
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
							"network_interface": schema.SetNestedAttribute{
								Description: "Network range defined for a site",
								Required:    false,
								Optional:    true,
								Computed:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(), // Avoid drift
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Description: "Network Interface Name",
											Required:    false,
											Optional:    true,
											// Validators: []validator.String{
											// 	stringvalidator.ConflictsWith(path.Expressions{
											// 		path.MatchRelative().AtParent().AtName("id"),
											// 	}...),
											// },
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
										},
									},
								},
							},
							"global_ip_range": schema.SetNestedAttribute{
								Description: "Global IP range matching criteria for the exception.",
								Required:    false,
								Optional:    true,
								Computed:    true,
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
							"floating_subnet": schema.SetNestedAttribute{
								Description: "Floating Subnets (ie. Floating Ranges) are used to identify traffic exactly matched to the route advertised by BGP. They are not associated with a specific site. This is useful in scenarios such as active-standby high availability routed via BGP.",
								Required:    false,
								Optional:    true,
								Computed:    true,
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
							"site_network_subnet": schema.SetNestedAttribute{
								Description: "GlobalRange + InterfaceSubnet",
								Required:    false,
								Optional:    true,
								Computed:    true,
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(), // Avoid drift
								},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Description: "Site Natwork Subnet Name",
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
											Description: "Site Natwork Subnet ID",
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
					"service": schema.SingleNestedAttribute{
						Description: "Destination service traffic matching criteria. Logical 'OR' is applied within the criteria set. Logical 'AND' is applied between criteria sets.",
						Required:    false,
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"simple": schema.SetNestedAttribute{
								Description: "Simple Service to which this Socket LAN rule applies",
								Required:    false,
								Optional:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Description: "Simple Service Name (e.g., HTTP, FTP, SSH)",
											Required:    true,
											Optional:    false,
										},
									},
								},
							},
							"custom": schema.ListNestedAttribute{
								Description: "Custom Service defined by a combination of L4 ports and an IP Protocol",
								Required:    false,
								Optional:    true,
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
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
					"nat": schema.SingleNestedAttribute{
						Description: "NAT settings",
						Required:    false,
						Optional:    true,
						Computed:    true,
						Attributes: map[string]schema.Attribute{
							"enabled": schema.BoolAttribute{
								Description: "Enable or disable NAT",
								Optional:    true,
								Computed:    true,
								Default:     booldefault.StaticBool(false),
							},
							"nat_type": schema.StringAttribute{
								Description: "NAT type (DYNAMIC_PAT)",
								Required:    false,
								Optional:    true,
								Validators: []validator.String{
									stringvalidator.OneOf("DYNAMIC_PAT"),
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *socketLanNetworkRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoClientData)
}

func (r *socketLanNetworkRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("rule").AtName("id"), req, resp)
}

func (r *socketLanNetworkRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SocketLanNetworkRule
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build API input
	apiInput, inputDiags := hydrateSocketLanNetworkRuleApi(ctx, plan)
	resp.Diagnostics.Append(inputDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Create.PolicySocketLanAddRule.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(apiInput.create),
	})

	policyChange, err := r.client.catov2.PolicySocketLanAddRule(ctx, apiInput.create, r.client.AccountId)
	tflog.Debug(ctx, "Create.PolicySocketLanAddRule.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(policyChange),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicySocketLanAddRule error",
			err.Error(),
		)
		return
	}

	if len(policyChange.Policy.SocketLan.AddRule.Errors) > 0 {
		for _, e := range policyChange.Policy.SocketLan.AddRule.Errors {
			resp.Diagnostics.AddError("ERROR: "+*e.ErrorCode, *e.ErrorMessage)
		}
		return
	}

	// Publish the changes
	tflog.Info(ctx, "Create.publishing-rule")
	publishInput := &cato_models.PolicyPublishRevisionInput{}
	_, err = r.client.catov2.PolicySocketLanPublishPolicyRevision(ctx, nil, publishInput, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicySocketLanPublishPolicyRevision error",
			err.Error(),
		)
		return
	}

	// Get the rule ID from the response
	ruleId := policyChange.GetPolicy().GetSocketLan().GetAddRule().Rule.GetRule().ID

	// Read back the rule to populate state
	queryResult, err := r.client.catov2.PolicySocketLanPolicy(ctx, r.client.AccountId, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicySocketLanPolicy error",
			err.Error(),
		)
		return
	}

	// Find the created rule
	var currentRule *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule
	for _, ruleWrapper := range queryResult.Policy.SocketLan.Policy.Rules {
		if ruleWrapper.Rule.ID == ruleId {
			currentRule = &ruleWrapper.Rule
			break
		}
	}

	if currentRule == nil {
		resp.Diagnostics.AddError(
			"Rule not found",
			fmt.Sprintf("Could not find created rule with ID %s", ruleId),
		)
		return
	}

	// Hydrate state from API response
	ruleState := hydrateSocketLanNetworkRuleState(ctx, plan, currentRule)

	// Build the rule object
	ruleObj, diagstmp := types.ObjectValueFrom(ctx, SocketLanNetworkRuleRuleAttrTypes, ruleState)
	resp.Diagnostics.Append(diagstmp...)

	// Set the state
	plan.Rule = ruleObj
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *socketLanNetworkRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SocketLanNetworkRule
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the rule ID from state
	var ruleData SocketLanNetworkRuleData
	diags = state.Rule.As(ctx, &ruleData, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ruleId := ruleData.ID.ValueString()

	// Query the API
	queryResult, err := r.client.catov2.PolicySocketLanPolicy(ctx, r.client.AccountId, nil)
	tflog.Debug(ctx, "Read.PolicySocketLanPolicy.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(queryResult),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API error",
			err.Error(),
		)
		return
	}

	// Find the rule
	var currentRule *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule
	ruleExists := false
	for _, ruleWrapper := range queryResult.Policy.SocketLan.Policy.Rules {
		if ruleWrapper.Rule.ID == ruleId {
			currentRule = &ruleWrapper.Rule
			ruleExists = true
			break
		}
	}

	if !ruleExists {
		tflog.Warn(ctx, "socket lan network rule not found, resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	// Hydrate state from API response
	ruleState := hydrateSocketLanNetworkRuleState(ctx, state, currentRule)

	// Build the rule object
	ruleObj, diagstmp := types.ObjectValueFrom(ctx, SocketLanNetworkRuleRuleAttrTypes, ruleState)
	resp.Diagnostics.Append(diagstmp...)

	// Hard code position to avoid drift
	curAtObj, diagstmp := types.ObjectValue(
		PositionAttrTypes,
		map[string]attr.Value{
			"position": types.StringValue("LAST_IN_POLICY"),
			"ref":      types.StringNull(),
		},
	)
	resp.Diagnostics.Append(diagstmp...)

	state.Rule = ruleObj
	state.At = curAtObj

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *socketLanNetworkRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SocketLanNetworkRule
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the rule ID from plan
	var ruleData SocketLanNetworkRuleData
	diags = plan.Rule.As(ctx, &ruleData, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ruleId := ruleData.ID.ValueString()

	// Build API input
	apiInput, inputDiags := hydrateSocketLanNetworkRuleApi(ctx, plan)
	resp.Diagnostics.Append(inputDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set the rule ID for update
	apiInput.update.ID = ruleId

	// Move rule if position changed
	var positionInput PolicyRulePositionInput
	diags = plan.At.As(ctx, &positionInput, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)

	moveInput := cato_models.PolicyMoveRuleInput{
		ID: ruleId,
		To: &cato_models.PolicyRulePositionInput{
			Position: (*cato_models.PolicyRulePositionEnum)(positionInput.Position.ValueStringPointer()),
			Ref:      positionInput.Ref.ValueStringPointer(),
		},
	}

	tflog.Debug(ctx, "Update.PolicySocketLanMoveRule.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(moveInput),
	})
	moveResult, err := r.client.catov2.PolicySocketLanMoveRule(ctx, moveInput, r.client.AccountId)
	tflog.Debug(ctx, "Update.PolicySocketLanMoveRule.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(moveResult),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicySocketLanMoveRule error",
			err.Error(),
		)
		return
	}

	// Update the rule
	tflog.Debug(ctx, "Update.PolicySocketLanUpdateRule.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(apiInput.update),
	})
	updateResult, err := r.client.catov2.PolicySocketLanUpdateRule(ctx, nil, apiInput.update, r.client.AccountId)
	tflog.Debug(ctx, "Update.PolicySocketLanUpdateRule.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(updateResult),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicySocketLanUpdateRule error",
			err.Error(),
		)
		return
	}

	if updateResult.Policy.SocketLan.UpdateRule.Status != "SUCCESS" {
		for _, item := range updateResult.Policy.SocketLan.UpdateRule.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Updating Rule Resource",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	// Publish the changes
	tflog.Info(ctx, "Update.publishing-rule")
	publishInput := &cato_models.PolicyPublishRevisionInput{}
	_, err = r.client.catov2.PolicySocketLanPublishPolicyRevision(ctx, nil, publishInput, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicySocketLanPublishPolicyRevision error",
			err.Error(),
		)
		return
	}

	// Read back and update state
	queryResult, err := r.client.catov2.PolicySocketLanPolicy(ctx, r.client.AccountId, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicySocketLanPolicy error",
			err.Error(),
		)
		return
	}

	var currentRule *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule
	for _, ruleWrapper := range queryResult.Policy.SocketLan.Policy.Rules {
		if ruleWrapper.Rule.ID == ruleId {
			currentRule = &ruleWrapper.Rule
			break
		}
	}

	if currentRule == nil {
		resp.Diagnostics.AddError(
			"Rule not found",
			fmt.Sprintf("Could not find rule with ID %s after update", ruleId),
		)
		return
	}

	// Hydrate state from API response
	ruleState := hydrateSocketLanNetworkRuleState(ctx, plan, currentRule)

	// Build the rule object
	ruleObj, diagstmp := types.ObjectValueFrom(ctx, SocketLanNetworkRuleRuleAttrTypes, ruleState)
	resp.Diagnostics.Append(diagstmp...)

	plan.Rule = ruleObj
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *socketLanNetworkRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SocketLanNetworkRule
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the rule ID from state
	var ruleData SocketLanNetworkRuleData
	diags = state.Rule.As(ctx, &ruleData, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ruleId := ruleData.ID.ValueString()

	removeInput := cato_models.SocketLanRemoveRuleInput{
		ID: ruleId,
	}

	tflog.Debug(ctx, "Delete.PolicySocketLanRemoveRule.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(removeInput),
	})
	removeResult, err := r.client.catov2.PolicySocketLanRemoveRule(ctx, nil, removeInput, r.client.AccountId)
	tflog.Debug(ctx, "Delete.PolicySocketLanRemoveRule.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(removeResult),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicySocketLanRemoveRule error",
			err.Error(),
		)
		return
	}

	// Publish the changes
	tflog.Info(ctx, "Delete.publishing-rule")
	publishInput := &cato_models.PolicyPublishRevisionInput{}
	_, err = r.client.catov2.PolicySocketLanPublishPolicyRevision(ctx, nil, publishInput, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicySocketLanPublishPolicyRevision error",
			err.Error(),
		)
		return
	}
}
