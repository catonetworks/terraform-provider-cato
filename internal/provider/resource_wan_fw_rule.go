package provider

import (
	"context"
	"fmt"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	cato_scalars "github.com/catonetworks/cato-go-sdk/scalars"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &wanFwRuleResource{}
	_ resource.ResourceWithConfigure   = &wanFwRuleResource{}
	_ resource.ResourceWithImportState = &wanFwRuleResource{}
)

func NewWanFwRuleResource() resource.Resource {
	return &wanFwRuleResource{}
}

type wanFwRuleResource struct {
	client *catoClientData
}

func (r *wanFwRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wf_rule"
}

func (r *wanFwRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_wf_rule` resource contains the configuration parameters necessary to add rule to the WAN Firewall. (check https://support.catonetworks.com/hc/en-us/articles/4413265660305-What-is-the-Cato-WAN-Firewall for more details). Documentation for the underlying API used in this resource can be found at [mutation.policy.wanFirewall.addRule()](https://api.catonetworks.com/documentation/#mutation-policy.wanFirewall.addRule).",
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
						Description: "ID of the  rule",
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
						Description: "",
						Required:    false,
						Optional:    true,
					},
					"enabled": schema.BoolAttribute{
						Description: "Attribute to define rule status (enabled or disabled)",
						Required:    true,
						Optional:    false,
					},
					// "section": schema.SingleNestedAttribute{
					// 	Required: false,
					// 	Optional: true,
					// 	Attributes: map[string]schema.Attribute{
					// 		"name": schema.StringAttribute{
					// 			Description: "",
					// 			Required:    false,
					// 			Optional:    true,
					// 			Validators: []validator.String{
					// 				stringvalidator.ConflictsWith(path.Expressions{
					// 					path.MatchRelative().AtParent().AtName("id"),
					// 				}...),
					// 			},
					// 		},
					// 		"id": schema.StringAttribute{
					// 			Description: "",
					// 			Required:    false,
					// 			Optional:    true,
					// 		},
					// 	},
					// },
					"action": schema.StringAttribute{
						Description: "The action applied by the Wan Firewall if the rule is matched (https://api.catonetworks.com/documentation/#definition-WanFirewallActionEnum)",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("ALLOW", "BLOCK", "PROMPT"),
						},
					},
					"direction": schema.StringAttribute{
						Description: "Define the direction on which the rule is applied (https://api.catonetworks.com/documentation/#definition-WanFirewallDirectionEnum)",
						Required:    true,
					},
					"source": schema.SingleNestedAttribute{
						Description: "Source traffic matching criteria. Logical ‘OR’ is applied within the criteria set. Logical ‘AND’ is applied between criteria sets.",
						Required:    false,
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"ip": schema.ListAttribute{
								Description: "Pv4 address list",
								ElementType: types.StringType,
								Required:    false,
								Optional:    true,
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
						},
					},
					"destination": schema.SingleNestedAttribute{
						Description: "Destination traffic matching criteria. Logical ‘OR’ is applied within the criteria set. Logical ‘AND’ is applied between criteria sets.",
						Required:    false,
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"ip": schema.ListAttribute{
								Description: "Pv4 address list",
								ElementType: types.StringType,
								Required:    false,
								Optional:    true,
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
						},
					},
					"connection_origin": schema.StringAttribute{
						Description: "Connection origin of the traffic (https://api.catonetworks.com/documentation/#definition-ConnectionOriginEnum)",
						Optional:    true,
						Required:    false,
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
								},
								"id": schema.StringAttribute{
									Description: "",
									Required:    false,
									Optional:    true,
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
					"application": schema.SingleNestedAttribute{
						Description: "Application traffic matching criteria. Logical ‘OR’ is applied within the criteria set. Logical ‘AND’ is applied between criteria sets.",
						Optional:    true,
						Required:    false,
						Attributes: map[string]schema.Attribute{
							"application": schema.SetNestedAttribute{
								Description: "Applications for the rule (pre-defined)",
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
										},
										"id": schema.StringAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
										},
									},
								},
							},
							"custom_app": schema.SetNestedAttribute{
								Description: "Custom (user-defined) applications",
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
										},
										"id": schema.StringAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
										},
									},
								},
							},
							"app_category": schema.SetNestedAttribute{
								Description: "Cato category of applications which are dynamically updated by Cato",
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
										},
										"id": schema.StringAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
										},
									},
								},
							},
							"custom_category": schema.SetNestedAttribute{
								Description: "Custom Categories – Groups of objects such as predefined and custom applications, predefined and custom services, domains, FQDNs etc.",
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
										},
										"id": schema.StringAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
										},
									},
								},
							},
							"sanctioned_apps_category": schema.SetNestedAttribute{
								Description: "Sanctioned Cloud Applications - apps that are approved and generally represent an understood and acceptable level of risk in your organization.",
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
										},
										"id": schema.StringAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
										},
									},
								},
							},
							"domain": schema.ListAttribute{
								ElementType: types.StringType,
								Description: "A Second-Level Domain (SLD). It matches all Top-Level Domains (TLD), and subdomains that include the Domain. Example: example.com.",
								Required:    false,
								Optional:    true,
							},
							"fqdn": schema.ListAttribute{
								ElementType: types.StringType,
								Description: "An exact match of the fully qualified domain (FQDN). Example: www.my.example.com.",
								Required:    false,
								Optional:    true,
							},
							"ip": schema.ListAttribute{
								ElementType: types.StringType,
								Description: "IPv4 addresses",
								Required:    false,
								Optional:    true,
							},
							"subnet": schema.ListAttribute{
								ElementType: types.StringType,
								Description: "Network subnets in CIDR notation",
								Required:    false,
								Optional:    true,
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
										},
										"id": schema.StringAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
										},
									},
								},
							},
						},
					},
					"service": schema.SingleNestedAttribute{
						Description: "Destination service traffic matching criteria. Logical ‘OR’ is applied within the criteria set. Logical ‘AND’ is applied between criteria sets.",
						Required:    false,
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"standard": schema.SetNestedAttribute{
								Description: "Standard Service to which this Wan Firewall rule applies",
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
					"tracking": schema.SingleNestedAttribute{
						Description: "Tracking information when the rule is matched, such as events and notifications",
						Required:    false,
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"event": schema.SingleNestedAttribute{
								Description: "When enabled, create an event each time the rule is matched",
								Required:    true,
								Attributes: map[string]schema.Attribute{
									"enabled": schema.BoolAttribute{
										Description: "",
										Required:    true,
										Optional:    false,
									},
								},
							},
							"alert": schema.SingleNestedAttribute{
								Description: "When enabled, send an alert each time the rule is matched",
								Required:    false,
								Optional:    true,
								Attributes: map[string]schema.Attribute{
									"enabled": schema.BoolAttribute{
										Description: "",
										Required:    true,
										Optional:    false,
									},
									"frequency": schema.StringAttribute{
										Description: "Returns data for the alert frequency (https://api.catonetworks.com/documentation/#definition-PolicyRuleTrackingFrequencyEnum)",
										Required:    true,
										Optional:    false,
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
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"active_on": schema.StringAttribute{
								Description: "Define when the rule is active (https://api.catonetworks.com/documentation/#definition-PolicyActiveOnEnum)",
								Required:    true,
								Optional:    false,
							},
							"custom_timeframe": schema.SingleNestedAttribute{
								Description: "Input of data for a custom one-time time range that a rule is active",
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
							"custom_recurring": schema.SingleNestedAttribute{
								Description: "Input of data for a custom recurring time range that a rule is active",
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
									"days": schema.ListAttribute{
										ElementType: types.StringType,
										Description: "(https://api.catonetworks.com/documentation/#definition-DayOfWeek)",
										Required:    true,
										Optional:    false,
									},
								},
							},
						},
					},
					"exceptions": schema.SetNestedAttribute{
						Description: "The set of exceptions for the rule. Exceptions define when the rule will be ignored and the firewall evaluation will continue with the lower priority rules.",
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
									Attributes: map[string]schema.Attribute{
										"ip": schema.ListAttribute{
											Description: "",
											ElementType: types.StringType,
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.List{
												listplanmodifier.UseStateForUnknown(), // Avoid drift
											},
										},
										"host": schema.SetNestedAttribute{
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
											Description: "",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.List{
												listplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Validators: []validator.List{
												listvalidator.SizeAtLeast(1),
											},
											Computed: true,
										},
										"ip_range": schema.ListNestedAttribute{
											Required: false,
											Optional: true,
											PlanModifiers: []planmodifier.List{
												listplanmodifier.UseStateForUnknown(), // Avoid drift
											},
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
										"user": schema.SetNestedAttribute{
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
								"destination": schema.SingleNestedAttribute{
									Description: "Destination traffic matching criteria for the exception.",
									Required:    false,
									Optional:    true,
									Attributes: map[string]schema.Attribute{
										"ip": schema.ListAttribute{
											Description: "",
											ElementType: types.StringType,
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.List{
												listplanmodifier.UseStateForUnknown(), // Avoid drift
											},
										},
										"host": schema.SetNestedAttribute{
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
											Description: "",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.List{
												listplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Computed: true,
										},
										"ip_range": schema.ListNestedAttribute{
											Required: false,
											Optional: true,
											PlanModifiers: []planmodifier.List{
												listplanmodifier.UseStateForUnknown(), // Avoid drift
											},
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
										"user": schema.SetNestedAttribute{
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
								"country": schema.SetNestedAttribute{
									Description: "Source country matching criteria for the exception.",
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
								"device": schema.SetNestedAttribute{
									Description: "Source Device Profile traffic matching criteria. Logical ‘OR’ is applied within the criteria set. Logical ‘AND’ is applied between criteria sets.",
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
									},
									Computed: true,
								},
								"application": schema.SingleNestedAttribute{
									Description: "Application matching criteria for the exception.",
									Optional:    true,
									Required:    false,
									Attributes: map[string]schema.Attribute{
										"application": schema.SetNestedAttribute{
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
										"custom_category": schema.SetNestedAttribute{
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
											Description: "",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.List{
												listplanmodifier.UseStateForUnknown(), // Avoid drift
											},
										},
										"fqdn": schema.ListAttribute{
											ElementType: types.StringType,
											Description: "",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.List{
												listplanmodifier.UseStateForUnknown(), // Avoid drift
											},
										},
										"ip": schema.ListAttribute{
											ElementType: types.StringType,
											Description: "",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.List{
												listplanmodifier.UseStateForUnknown(), // Avoid drift
											},
										},
										"subnet": schema.ListAttribute{
											ElementType: types.StringType,
											Description: "",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.List{
												listplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Computed: true,
										},
										"ip_range": schema.ListNestedAttribute{
											Required: false,
											Optional: true,
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
								"service": schema.SingleNestedAttribute{
									Description: "Destination service matching criteria for the exception.",
									Required:    false,
									Optional:    true,
									Attributes: map[string]schema.Attribute{
										"standard": schema.SetNestedAttribute{
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
											Description: "",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.List{
												listplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"port": schema.ListAttribute{
														ElementType: types.StringType,
														Description: "",
														Optional:    true,
														Required:    false,
													},
													"port_range": schema.SingleNestedAttribute{
														Required: false,
														Optional: true,
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
														Description: "",
														Required:    false,
														Optional:    true,
													},
												},
											},
										},
									},
								},
								"direction": schema.StringAttribute{
									Description: "Direction matching criteria for the exception.",
									Required:    true,
									PlanModifiers: []planmodifier.String{
										stringplanmodifier.UseStateForUnknown(), // Avoid drift
									},
									Validators: []validator.String{
										stringvalidator.OneOf("BOTH", "TO"),
									},
								},
								"connection_origin": schema.StringAttribute{
									Description: "Connection origin matching criteria for the exception.",
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

func (r *wanFwRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoClientData)
}

func (r *wanFwRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("rule").AtName("id"), req, resp)
}

func (r *wanFwRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	var plan WanFirewallRule
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	//initiate input
	input := cato_models.WanFirewallAddRuleInput{}

	//setting at
	if !plan.At.IsNull() {
		input.At = &cato_models.PolicyRulePositionInput{}
		positionInput := PolicyRulePositionInput{}
		diags = plan.At.As(ctx, &positionInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)

		input.At.Position = (*cato_models.PolicyRulePositionEnum)(positionInput.Position.ValueStringPointer())
		input.At.Ref = positionInput.Ref.ValueStringPointer()
	}

	// setting rule
	if !plan.Rule.IsNull() {

		input.Rule = &cato_models.WanFirewallAddRuleDataInput{}
		ruleInput := Policy_Policy_WanFirewall_Policy_Rules_Rule{}
		diags = plan.Rule.As(ctx, &ruleInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)

		// setting source
		if !ruleInput.Source.IsNull() {

			input.Rule.Source = &cato_models.WanFirewallSourceInput{}
			sourceInput := Policy_Policy_WanFirewall_Policy_Rules_Rule_Source{}
			diags = ruleInput.Source.As(ctx, &sourceInput, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)

			// setting source IP
			if !sourceInput.IP.IsNull() {
				diags = sourceInput.IP.ElementsAs(ctx, &input.Rule.Source.IP, false)
				resp.Diagnostics.Append(diags...)
			}

			// setting source subnet
			if !sourceInput.Subnet.IsNull() {
				diags = sourceInput.Subnet.ElementsAs(ctx, &input.Rule.Source.Subnet, false)
				resp.Diagnostics.Append(diags...)
			}

			// setting source host
			if !sourceInput.Host.IsNull() {
				elementsSourceHostInput := make([]types.Object, 0, len(sourceInput.Host.Elements()))
				diags = sourceInput.Host.ElementsAs(ctx, &elementsSourceHostInput, false)
				resp.Diagnostics.Append(diags...)

				var itemSourceHostInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_Host
				for _, item := range elementsSourceHostInput {
					diags = item.As(ctx, &itemSourceHostInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceHostInput)
					if err != nil {
						resp.Diagnostics.AddError(
							"Object Ref transformation failed",
							err.Error(),
						)
						return
					}

					input.Rule.Source.Host = append(input.Rule.Source.Host, &cato_models.HostRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
			}

			// setting source site
			if !sourceInput.Site.IsNull() {
				elementsSourceSiteInput := make([]types.Object, 0, len(sourceInput.Site.Elements()))
				diags = sourceInput.Site.ElementsAs(ctx, &elementsSourceSiteInput, false)
				resp.Diagnostics.Append(diags...)

				var itemSourceSiteInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_Site
				for _, item := range elementsSourceSiteInput {
					diags = item.As(ctx, &itemSourceSiteInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceSiteInput)
					if err != nil {
						resp.Diagnostics.AddError(
							"Object Ref transformation failed",
							err.Error(),
						)
						return
					}

					input.Rule.Source.Site = append(input.Rule.Source.Site, &cato_models.SiteRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
			}

			// setting source ip range
			if !sourceInput.IPRange.IsNull() {
				elementsSourceIPRangeInput := make([]types.Object, 0, len(sourceInput.IPRange.Elements()))
				diags = sourceInput.IPRange.ElementsAs(ctx, &elementsSourceIPRangeInput, false)
				resp.Diagnostics.Append(diags...)

				var itemSourceIPRangeInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_IPRange
				for _, item := range elementsSourceIPRangeInput {
					diags = item.As(ctx, &itemSourceIPRangeInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					input.Rule.Source.IPRange = append(input.Rule.Source.IPRange, &cato_models.IPAddressRangeInput{
						From: itemSourceIPRangeInput.From.ValueString(),
						To:   itemSourceIPRangeInput.To.ValueString(),
					})
				}
			}

			// setting source global ip range
			if !sourceInput.GlobalIPRange.IsNull() {
				elementsSourceGlobalIPRangeInput := make([]types.Object, 0, len(sourceInput.GlobalIPRange.Elements()))
				diags = sourceInput.GlobalIPRange.ElementsAs(ctx, &elementsSourceGlobalIPRangeInput, false)
				resp.Diagnostics.Append(diags...)

				var itemSourceGlobalIPRangeInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_GlobalIPRange
				for _, item := range elementsSourceGlobalIPRangeInput {
					diags = item.As(ctx, &itemSourceGlobalIPRangeInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceGlobalIPRangeInput)
					if err != nil {
						resp.Diagnostics.AddError(
							"Object Ref transformation failed for",
							err.Error(),
						)
						return
					}

					input.Rule.Source.GlobalIPRange = append(input.Rule.Source.GlobalIPRange, &cato_models.GlobalIPRangeRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
			}

			// setting source network interface
			if !sourceInput.NetworkInterface.IsNull() {
				elementsSourceNetworkInterfaceInput := make([]types.Object, 0, len(sourceInput.NetworkInterface.Elements()))
				diags = sourceInput.NetworkInterface.ElementsAs(ctx, &elementsSourceNetworkInterfaceInput, false)
				resp.Diagnostics.Append(diags...)

				var itemSourceNetworkInterfaceInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_NetworkInterface
				for _, item := range elementsSourceNetworkInterfaceInput {
					diags = item.As(ctx, &itemSourceNetworkInterfaceInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceNetworkInterfaceInput)
					if err != nil {
						resp.Diagnostics.AddError(
							"Object Ref transformation failed",
							err.Error(),
						)
						return
					}

					input.Rule.Source.NetworkInterface = append(input.Rule.Source.NetworkInterface, &cato_models.NetworkInterfaceRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
			}

			// setting source site network subnet
			if !sourceInput.SiteNetworkSubnet.IsNull() {
				elementsSourceSiteNetworkSubnetInput := make([]types.Object, 0, len(sourceInput.SiteNetworkSubnet.Elements()))
				diags = sourceInput.SiteNetworkSubnet.ElementsAs(ctx, &elementsSourceSiteNetworkSubnetInput, false)
				resp.Diagnostics.Append(diags...)

				var itemSourceSiteNetworkSubnetInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_SiteNetworkSubnet
				for _, item := range elementsSourceSiteNetworkSubnetInput {
					diags = item.As(ctx, &itemSourceSiteNetworkSubnetInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceSiteNetworkSubnetInput)
					if err != nil {
						resp.Diagnostics.AddError(
							"Object Ref transformation failed",
							err.Error(),
						)
						return
					}

					input.Rule.Source.SiteNetworkSubnet = append(input.Rule.Source.SiteNetworkSubnet, &cato_models.SiteNetworkSubnetRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
			}

			// setting source floating subnet
			if !sourceInput.FloatingSubnet.IsNull() {
				elementsSourceFloatingSubnetInput := make([]types.Object, 0, len(sourceInput.FloatingSubnet.Elements()))
				diags = sourceInput.FloatingSubnet.ElementsAs(ctx, &elementsSourceFloatingSubnetInput, false)
				resp.Diagnostics.Append(diags...)

				var itemSourceFloatingSubnetInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_FloatingSubnet
				for _, item := range elementsSourceFloatingSubnetInput {
					diags = item.As(ctx, &itemSourceFloatingSubnetInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceFloatingSubnetInput)
					if err != nil {
						resp.Diagnostics.AddError(
							"Object Ref transformation failed",
							err.Error(),
						)
						return
					}

					input.Rule.Source.FloatingSubnet = append(input.Rule.Source.FloatingSubnet, &cato_models.FloatingSubnetRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
			}

			// setting source user
			if !sourceInput.User.IsNull() {
				elementsSourceUserInput := make([]types.Object, 0, len(sourceInput.User.Elements()))
				diags = sourceInput.User.ElementsAs(ctx, &elementsSourceUserInput, false)
				resp.Diagnostics.Append(diags...)

				var itemSourceUserInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_User
				for _, item := range elementsSourceUserInput {
					diags = item.As(ctx, &itemSourceUserInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceUserInput)
					if err != nil {
						resp.Diagnostics.AddError(
							"Object Ref transformation failed",
							err.Error(),
						)
						return
					}

					input.Rule.Source.User = append(input.Rule.Source.User, &cato_models.UserRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
			}

			// setting source users group
			if !sourceInput.UsersGroup.IsNull() {
				elementsSourceUsersGroupInput := make([]types.Object, 0, len(sourceInput.UsersGroup.Elements()))
				diags = sourceInput.UsersGroup.ElementsAs(ctx, &elementsSourceUsersGroupInput, false)
				resp.Diagnostics.Append(diags...)

				var itemSourceUsersGroupInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_UsersGroup
				for _, item := range elementsSourceUsersGroupInput {
					diags = item.As(ctx, &itemSourceUsersGroupInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceUsersGroupInput)
					if err != nil {
						resp.Diagnostics.AddError(
							"Object Ref transformation failed",
							err.Error(),
						)
						return
					}

					input.Rule.Source.UsersGroup = append(input.Rule.Source.UsersGroup, &cato_models.UsersGroupRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
			}

			// setting source group
			if !sourceInput.Group.IsNull() {
				elementsSourceGroupInput := make([]types.Object, 0, len(sourceInput.Group.Elements()))
				diags = sourceInput.Group.ElementsAs(ctx, &elementsSourceGroupInput, false)
				resp.Diagnostics.Append(diags...)

				var itemSourceGroupInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_Group
				for _, item := range elementsSourceGroupInput {
					diags = item.As(ctx, &itemSourceGroupInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceGroupInput)
					if err != nil {
						resp.Diagnostics.AddError(
							"Object Ref transformation failed",
							err.Error(),
						)
						return
					}

					input.Rule.Source.Group = append(input.Rule.Source.Group, &cato_models.GroupRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
			}

			// setting source system group
			if !sourceInput.SystemGroup.IsNull() {
				elementsSourceSystemGroupInput := make([]types.Object, 0, len(sourceInput.SystemGroup.Elements()))
				diags = sourceInput.SystemGroup.ElementsAs(ctx, &elementsSourceSystemGroupInput, false)
				resp.Diagnostics.Append(diags...)

				var itemSourceSystemGroupInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_SystemGroup
				for _, item := range elementsSourceSystemGroupInput {
					diags = item.As(ctx, &itemSourceSystemGroupInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceSystemGroupInput)
					if err != nil {
						resp.Diagnostics.AddError(
							"Object Ref transformation failed",
							err.Error(),
						)
						return
					}

					input.Rule.Source.SystemGroup = append(input.Rule.Source.SystemGroup, &cato_models.SystemGroupRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
			}
		}

		// setting country
		if !ruleInput.Country.IsNull() {
			elementsCountryInput := make([]types.Object, 0, len(ruleInput.Country.Elements()))
			diags = ruleInput.Country.ElementsAs(ctx, &elementsCountryInput, false)
			resp.Diagnostics.Append(diags...)

			var itemCountryInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Country
			for _, item := range elementsCountryInput {
				diags = item.As(ctx, &itemCountryInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				ObjectRefOutput, err := utils.TransformObjectRefInput(itemCountryInput)
				if err != nil {
					resp.Diagnostics.AddError(
						"Object Ref transformation failed",
						err.Error(),
					)
					return
				}

				input.Rule.Country = append(input.Rule.Country, &cato_models.CountryRefInput{
					By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
					Input: ObjectRefOutput.Input,
				})
			}
		}

		// setting device
		if !ruleInput.Device.IsNull() {
			elementsDeviceInput := make([]types.Object, 0, len(ruleInput.Device.Elements()))
			diags = ruleInput.Device.ElementsAs(ctx, &elementsDeviceInput, false)
			resp.Diagnostics.Append(diags...)

			var itemDeviceInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Device
			for _, item := range elementsDeviceInput {
				diags = item.As(ctx, &itemDeviceInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				ObjectRefOutput, err := utils.TransformObjectRefInput(itemDeviceInput)
				if err != nil {
					resp.Diagnostics.AddError(
						"Object Ref transformation failed",
						err.Error(),
					)
					return
				}

				input.Rule.Device = append(input.Rule.Device, &cato_models.DeviceProfileRefInput{
					By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
					Input: ObjectRefOutput.Input,
				})
			}
		}

		// setting device OS
		if !ruleInput.DeviceOs.IsNull() {
			diags = ruleInput.DeviceOs.ElementsAs(ctx, &input.Rule.DeviceOs, false)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
		}

		// setting destination
		if !ruleInput.Destination.IsNull() {

			input.Rule.Destination = &cato_models.WanFirewallDestinationInput{}
			destinationInput := Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination{}
			diags = ruleInput.Destination.As(ctx, &destinationInput, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)

			// setting destination IP
			if !destinationInput.IP.IsNull() {
				diags = destinationInput.IP.ElementsAs(ctx, &input.Rule.Destination.IP, false)
				resp.Diagnostics.Append(diags...)
			}

			// setting destination subnet
			if !destinationInput.Subnet.IsNull() {
				diags = destinationInput.Subnet.ElementsAs(ctx, &input.Rule.Destination.Subnet, false)
				resp.Diagnostics.Append(diags...)
			}

			// setting destination host
			if !destinationInput.Host.IsNull() {
				elementsDestinationHostInput := make([]types.Object, 0, len(destinationInput.Host.Elements()))
				diags = destinationInput.Host.ElementsAs(ctx, &elementsDestinationHostInput, false)
				resp.Diagnostics.Append(diags...)

				var itemDestinationHostInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_Host
				for _, item := range elementsDestinationHostInput {
					diags = item.As(ctx, &itemDestinationHostInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationHostInput)
					if err != nil {
						resp.Diagnostics.AddError(
							"Object Ref transformation failed",
							err.Error(),
						)
						return
					}

					input.Rule.Destination.Host = append(input.Rule.Destination.Host, &cato_models.HostRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
			}

			// setting destination site
			if !destinationInput.Site.IsNull() {
				elementsDestinationSiteInput := make([]types.Object, 0, len(destinationInput.Site.Elements()))
				diags = destinationInput.Site.ElementsAs(ctx, &elementsDestinationSiteInput, false)
				resp.Diagnostics.Append(diags...)

				var itemDestinationSiteInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_Site
				for _, item := range elementsDestinationSiteInput {
					diags = item.As(ctx, &itemDestinationSiteInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationSiteInput)
					if err != nil {
						resp.Diagnostics.AddError(
							"Object Ref transformation failed",
							err.Error(),
						)
						return
					}

					input.Rule.Destination.Site = append(input.Rule.Destination.Site, &cato_models.SiteRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
			}

			// setting destination ip range
			if !destinationInput.IPRange.IsNull() {
				elementsDestinationIPRangeInput := make([]types.Object, 0, len(destinationInput.IPRange.Elements()))
				diags = destinationInput.IPRange.ElementsAs(ctx, &elementsDestinationIPRangeInput, false)
				resp.Diagnostics.Append(diags...)

				var itemDestinationIPRangeInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_IPRange
				for _, item := range elementsDestinationIPRangeInput {
					diags = item.As(ctx, &itemDestinationIPRangeInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					input.Rule.Destination.IPRange = append(input.Rule.Destination.IPRange, &cato_models.IPAddressRangeInput{
						From: itemDestinationIPRangeInput.From.ValueString(),
						To:   itemDestinationIPRangeInput.To.ValueString(),
					})
				}
			}

			// setting destination global ip range
			if !destinationInput.GlobalIPRange.IsNull() {
				elementsDestinationGlobalIPRangeInput := make([]types.Object, 0, len(destinationInput.GlobalIPRange.Elements()))
				diags = destinationInput.GlobalIPRange.ElementsAs(ctx, &elementsDestinationGlobalIPRangeInput, false)
				resp.Diagnostics.Append(diags...)

				var itemDestinationGlobalIPRangeInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_GlobalIPRange
				for _, item := range elementsDestinationGlobalIPRangeInput {
					diags = item.As(ctx, &itemDestinationGlobalIPRangeInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationGlobalIPRangeInput)
					if err != nil {
						resp.Diagnostics.AddError(
							"Object Ref transformation failed for",
							err.Error(),
						)
						return
					}

					input.Rule.Destination.GlobalIPRange = append(input.Rule.Destination.GlobalIPRange, &cato_models.GlobalIPRangeRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
			}

			// setting destination network interface
			if !destinationInput.NetworkInterface.IsNull() {
				elementsDestinationNetworkInterfaceInput := make([]types.Object, 0, len(destinationInput.NetworkInterface.Elements()))
				diags = destinationInput.NetworkInterface.ElementsAs(ctx, &elementsDestinationNetworkInterfaceInput, false)
				resp.Diagnostics.Append(diags...)

				var itemDestinationNetworkInterfaceInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_NetworkInterface
				for _, item := range elementsDestinationNetworkInterfaceInput {
					diags = item.As(ctx, &itemDestinationNetworkInterfaceInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationNetworkInterfaceInput)
					if err != nil {
						resp.Diagnostics.AddError(
							"Object Ref transformation failed",
							err.Error(),
						)
						return
					}

					input.Rule.Destination.NetworkInterface = append(input.Rule.Destination.NetworkInterface, &cato_models.NetworkInterfaceRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
			}

			// setting destination site network subnet
			if !destinationInput.SiteNetworkSubnet.IsNull() {
				elementsDestinationSiteNetworkSubnetInput := make([]types.Object, 0, len(destinationInput.SiteNetworkSubnet.Elements()))
				diags = destinationInput.SiteNetworkSubnet.ElementsAs(ctx, &elementsDestinationSiteNetworkSubnetInput, false)
				resp.Diagnostics.Append(diags...)

				var itemDestinationSiteNetworkSubnetInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_SiteNetworkSubnet
				for _, item := range elementsDestinationSiteNetworkSubnetInput {
					diags = item.As(ctx, &itemDestinationSiteNetworkSubnetInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationSiteNetworkSubnetInput)
					if err != nil {
						resp.Diagnostics.AddError(
							"Object Ref transformation failed",
							err.Error(),
						)
						return
					}

					input.Rule.Destination.SiteNetworkSubnet = append(input.Rule.Destination.SiteNetworkSubnet, &cato_models.SiteNetworkSubnetRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
			}

			// setting destination floating subnet
			if !destinationInput.FloatingSubnet.IsNull() {
				elementsDestinationFloatingSubnetInput := make([]types.Object, 0, len(destinationInput.FloatingSubnet.Elements()))
				diags = destinationInput.FloatingSubnet.ElementsAs(ctx, &elementsDestinationFloatingSubnetInput, false)
				resp.Diagnostics.Append(diags...)

				var itemDestinationFloatingSubnetInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_FloatingSubnet
				for _, item := range elementsDestinationFloatingSubnetInput {
					diags = item.As(ctx, &itemDestinationFloatingSubnetInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationFloatingSubnetInput)
					if err != nil {
						resp.Diagnostics.AddError(
							"Object Ref transformation failed",
							err.Error(),
						)
						return
					}

					input.Rule.Destination.FloatingSubnet = append(input.Rule.Destination.FloatingSubnet, &cato_models.FloatingSubnetRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
			}

			// setting destination user
			if !destinationInput.User.IsNull() {
				elementsDestinationUserInput := make([]types.Object, 0, len(destinationInput.User.Elements()))
				diags = destinationInput.User.ElementsAs(ctx, &elementsDestinationUserInput, false)
				resp.Diagnostics.Append(diags...)

				var itemDestinationUserInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_User
				for _, item := range elementsDestinationUserInput {
					diags = item.As(ctx, &itemDestinationUserInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationUserInput)
					if err != nil {
						resp.Diagnostics.AddError(
							"Object Ref transformation failed",
							err.Error(),
						)
						return
					}

					input.Rule.Destination.User = append(input.Rule.Destination.User, &cato_models.UserRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
			}

			// setting destination users group
			if !destinationInput.UsersGroup.IsNull() {
				elementsDestinationUsersGroupInput := make([]types.Object, 0, len(destinationInput.UsersGroup.Elements()))
				diags = destinationInput.UsersGroup.ElementsAs(ctx, &elementsDestinationUsersGroupInput, false)
				resp.Diagnostics.Append(diags...)

				var itemDestinationUsersGroupInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_UsersGroup
				for _, item := range elementsDestinationUsersGroupInput {
					diags = item.As(ctx, &itemDestinationUsersGroupInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationUsersGroupInput)
					if err != nil {
						resp.Diagnostics.AddError(
							"Object Ref transformation failed",
							err.Error(),
						)
						return
					}

					input.Rule.Destination.UsersGroup = append(input.Rule.Destination.UsersGroup, &cato_models.UsersGroupRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
			}

			// setting destination group
			if !destinationInput.Group.IsNull() {
				elementsDestinationGroupInput := make([]types.Object, 0, len(destinationInput.Group.Elements()))
				diags = destinationInput.Group.ElementsAs(ctx, &elementsDestinationGroupInput, false)
				resp.Diagnostics.Append(diags...)

				var itemDestinationGroupInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_Group
				for _, item := range elementsDestinationGroupInput {
					diags = item.As(ctx, &itemDestinationGroupInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationGroupInput)
					if err != nil {
						resp.Diagnostics.AddError(
							"Object Ref transformation failed",
							err.Error(),
						)
						return
					}

					input.Rule.Destination.Group = append(input.Rule.Destination.Group, &cato_models.GroupRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
			}

			// setting destination system group
			if !destinationInput.SystemGroup.IsNull() {
				elementsDestinationSystemGroupInput := make([]types.Object, 0, len(destinationInput.SystemGroup.Elements()))
				diags = destinationInput.SystemGroup.ElementsAs(ctx, &elementsDestinationSystemGroupInput, false)
				resp.Diagnostics.Append(diags...)

				var itemDestinationSystemGroupInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_SystemGroup
				for _, item := range elementsDestinationSystemGroupInput {
					diags = item.As(ctx, &itemDestinationSystemGroupInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationSystemGroupInput)
					if err != nil {
						resp.Diagnostics.AddError(
							"Object Ref transformation failed",
							err.Error(),
						)
						return
					}

					input.Rule.Destination.SystemGroup = append(input.Rule.Destination.SystemGroup, &cato_models.SystemGroupRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
			}
		}

		// setting application
		if !ruleInput.Application.IsNull() {
			input.Rule.Application = &cato_models.WanFirewallApplicationInput{}
			applicationInput := Policy_Policy_WanFirewall_Policy_Rules_Rule_Application{}
			diags = ruleInput.Application.As(ctx, &applicationInput, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)

			// setting application IP
			if !applicationInput.IP.IsNull() {
				diags = applicationInput.IP.ElementsAs(ctx, &input.Rule.Application.IP, false)
				resp.Diagnostics.Append(diags...)
			}

			// setting application subnet
			if !applicationInput.Subnet.IsNull() {
				diags = applicationInput.Subnet.ElementsAs(ctx, &input.Rule.Application.Subnet, false)
				resp.Diagnostics.Append(diags...)
			}

			// setting application domain
			if !applicationInput.Domain.IsNull() {
				diags = applicationInput.Domain.ElementsAs(ctx, &input.Rule.Application.Domain, false)
				resp.Diagnostics.Append(diags...)
			}

			// setting application fqdn
			if !applicationInput.Fqdn.IsNull() {
				diags = applicationInput.Fqdn.ElementsAs(ctx, &input.Rule.Application.Fqdn, false)
				resp.Diagnostics.Append(diags...)
			}

			// setting application application
			if !applicationInput.Application.IsNull() {
				elementsApplicationApplicationInput := make([]types.Object, 0, len(applicationInput.Application.Elements()))
				diags = applicationInput.Application.ElementsAs(ctx, &elementsApplicationApplicationInput, false)
				resp.Diagnostics.Append(diags...)

				var itemApplicationApplicationInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_Application
				for _, item := range elementsApplicationApplicationInput {
					diags = item.As(ctx, &itemApplicationApplicationInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationApplicationInput)
					if err != nil {
						resp.Diagnostics.AddError(
							"Object Ref transformation failed",
							err.Error(),
						)
						return
					}

					input.Rule.Application.Application = append(input.Rule.Application.Application, &cato_models.ApplicationRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
			}

			// setting application custom app
			if !applicationInput.CustomApp.IsNull() {
				elementsApplicationCustomAppInput := make([]types.Object, 0, len(applicationInput.CustomApp.Elements()))
				diags = applicationInput.CustomApp.ElementsAs(ctx, &elementsApplicationCustomAppInput, false)
				resp.Diagnostics.Append(diags...)

				var itemApplicationCustomAppInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_CustomApp
				for _, item := range elementsApplicationCustomAppInput {
					diags = item.As(ctx, &itemApplicationCustomAppInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationCustomAppInput)
					if err != nil {
						resp.Diagnostics.AddError(
							"Object Ref transformation failed",
							err.Error(),
						)
						return
					}

					input.Rule.Application.CustomApp = append(input.Rule.Application.CustomApp, &cato_models.CustomApplicationRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
			}

			// setting application ip range
			if !applicationInput.IPRange.IsNull() {
				elementsApplicationIPRangeInput := make([]types.Object, 0, len(applicationInput.IPRange.Elements()))
				diags = applicationInput.IPRange.ElementsAs(ctx, &elementsApplicationIPRangeInput, false)
				resp.Diagnostics.Append(diags...)

				var itemApplicationIPRangeInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_IPRange
				for _, item := range elementsApplicationIPRangeInput {
					diags = item.As(ctx, &itemApplicationIPRangeInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					input.Rule.Application.IPRange = append(input.Rule.Application.IPRange, &cato_models.IPAddressRangeInput{
						From: itemApplicationIPRangeInput.From.ValueString(),
						To:   itemApplicationIPRangeInput.To.ValueString(),
					})
				}
			}

			// setting application global ip range
			if !applicationInput.GlobalIPRange.IsNull() {
				elementsApplicationGlobalIPRangeInput := make([]types.Object, 0, len(applicationInput.GlobalIPRange.Elements()))
				diags = applicationInput.GlobalIPRange.ElementsAs(ctx, &elementsApplicationGlobalIPRangeInput, false)
				resp.Diagnostics.Append(diags...)

				var itemApplicationGlobalIPRangeInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_GlobalIPRange
				for _, item := range elementsApplicationGlobalIPRangeInput {
					diags = item.As(ctx, &itemApplicationGlobalIPRangeInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationGlobalIPRangeInput)
					if err != nil {
						resp.Diagnostics.AddError(
							"Object Ref transformation failed",
							err.Error(),
						)
						return
					}

					input.Rule.Application.GlobalIPRange = append(input.Rule.Application.GlobalIPRange, &cato_models.GlobalIPRangeRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
			}

			// setting application app category
			if !applicationInput.AppCategory.IsNull() {
				elementsApplicationAppCategoryInput := make([]types.Object, 0, len(applicationInput.AppCategory.Elements()))
				diags = applicationInput.AppCategory.ElementsAs(ctx, &elementsApplicationAppCategoryInput, false)
				resp.Diagnostics.Append(diags...)

				var itemApplicationAppCategoryInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_AppCategory
				for _, item := range elementsApplicationAppCategoryInput {
					diags = item.As(ctx, &itemApplicationAppCategoryInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationAppCategoryInput)
					if err != nil {
						resp.Diagnostics.AddError(
							"Object Ref transformation failed",
							err.Error(),
						)
						return
					}

					input.Rule.Application.AppCategory = append(input.Rule.Application.AppCategory, &cato_models.ApplicationCategoryRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
			}

			// setting application custom app category
			if !applicationInput.CustomCategory.IsNull() {
				elementsApplicationCustomCategoryInput := make([]types.Object, 0, len(applicationInput.CustomCategory.Elements()))
				diags = applicationInput.CustomCategory.ElementsAs(ctx, &elementsApplicationCustomCategoryInput, false)
				resp.Diagnostics.Append(diags...)

				var itemApplicationCustomCategoryInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_CustomCategory
				for _, item := range elementsApplicationCustomCategoryInput {
					diags = item.As(ctx, &itemApplicationCustomCategoryInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationCustomCategoryInput)
					if err != nil {
						resp.Diagnostics.AddError(
							"Object Ref transformation failed",
							err.Error(),
						)
						return
					}

					input.Rule.Application.CustomCategory = append(input.Rule.Application.CustomCategory, &cato_models.CustomCategoryRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
			}

			// setting application sanctionned apps category
			if !applicationInput.SanctionedAppsCategory.IsNull() {
				elementsApplicationSanctionedAppsCategoryInput := make([]types.Object, 0, len(applicationInput.SanctionedAppsCategory.Elements()))
				diags = applicationInput.SanctionedAppsCategory.ElementsAs(ctx, &elementsApplicationSanctionedAppsCategoryInput, false)
				resp.Diagnostics.Append(diags...)

				var itemApplicationSanctionedAppsCategoryInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_SanctionedAppsCategory
				for _, item := range elementsApplicationSanctionedAppsCategoryInput {
					diags = item.As(ctx, &itemApplicationSanctionedAppsCategoryInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationSanctionedAppsCategoryInput)
					if err != nil {
						resp.Diagnostics.AddError(
							"Object Ref transformation failed",
							err.Error(),
						)
						return
					}

					input.Rule.Application.SanctionedAppsCategory = append(input.Rule.Application.SanctionedAppsCategory, &cato_models.SanctionedAppsCategoryRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
			}
		}

		// setting service
		if !ruleInput.Service.IsNull() {
			input.Rule.Service = &cato_models.WanFirewallServiceTypeInput{}
			serviceInput := Policy_Policy_WanFirewall_Policy_Rules_Rule_Service{}
			diags = ruleInput.Service.As(ctx, &serviceInput, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			// setting service standard
			if !serviceInput.Standard.IsNull() {
				elementsServiceStandardInput := make([]types.Object, 0, len(serviceInput.Standard.Elements()))
				diags = serviceInput.Standard.ElementsAs(ctx, &elementsServiceStandardInput, false)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}

				var itemServiceStandardInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Service_Standard
				for _, item := range elementsServiceStandardInput {
					diags = item.As(ctx, &itemServiceStandardInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemServiceStandardInput)
					if err != nil {
						resp.Diagnostics.AddError(
							"Object Ref transformation failed",
							err.Error(),
						)
						return
					}

					input.Rule.Service.Standard = append(input.Rule.Service.Standard, &cato_models.ServiceRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
			}

			// setting service custom
			if !serviceInput.Custom.IsNull() {
				elementsServiceCustomInput := make([]types.Object, 0, len(serviceInput.Custom.Elements()))
				diags = serviceInput.Custom.ElementsAs(ctx, &elementsServiceCustomInput, false)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}

				var itemServiceCustomInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Service_Custom
				for _, item := range elementsServiceCustomInput {
					diags = item.As(ctx, &itemServiceCustomInput, basetypes.ObjectAsOptions{})

					customInput := &cato_models.CustomServiceInput{
						Protocol: cato_models.IPProtocol(itemServiceCustomInput.Protocol.ValueString()),
					}

					// setting service custom port
					if !itemServiceCustomInput.Port.IsNull() {
						elementsPort := make([]types.String, 0, len(itemServiceCustomInput.Port.Elements()))
						diags = itemServiceCustomInput.Port.ElementsAs(ctx, &elementsPort, false)
						resp.Diagnostics.Append(diags...)

						inputPort := []cato_scalars.Port{}
						for _, item := range elementsPort {
							inputPort = append(inputPort, cato_scalars.Port(item.ValueString()))
						}

						customInput.Port = inputPort
					}

					// setting service custom port range
					if !itemServiceCustomInput.PortRange.IsNull() {
						var itemPortRange Policy_Policy_WanFirewall_Policy_Rules_Rule_Service_Custom_PortRange
						diags = itemServiceCustomInput.PortRange.As(ctx, &itemPortRange, basetypes.ObjectAsOptions{})

						inputPortRange := cato_models.PortRangeInput{
							From: cato_scalars.Port(itemPortRange.From.ValueString()),
							To:   cato_scalars.Port(itemPortRange.To.ValueString()),
						}

						customInput.PortRange = &inputPortRange
					}

					// append custom service
					input.Rule.Service.Custom = append(input.Rule.Service.Custom, customInput)
				}
			}
		}

		// setting tracking
		if !ruleInput.Tracking.IsNull() {

			input.Rule.Tracking = &cato_models.PolicyTrackingInput{
				Event: &cato_models.PolicyRuleTrackingEventInput{},
			}

			trackingInput := Policy_Policy_WanFirewall_Policy_Rules_Rule_Tracking{}
			diags = ruleInput.Tracking.As(ctx, &trackingInput, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			// setting tracking event
			trackingEventInput := Policy_Policy_WanFirewall_Policy_Rules_Rule_Tracking_Event{}
			diags = trackingInput.Event.As(ctx, &trackingEventInput, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			input.Rule.Tracking.Event.Enabled = trackingEventInput.Enabled.ValueBool()

			if !trackingInput.Alert.IsNull() {

				input.Rule.Tracking.Alert = &cato_models.PolicyRuleTrackingAlertInput{}

				trackingAlertInput := Policy_Policy_WanFirewall_Policy_Rules_Rule_Tracking_Alert{}
				diags = trackingInput.Alert.As(ctx, &trackingAlertInput, basetypes.ObjectAsOptions{})

				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
				input.Rule.Tracking.Alert.Enabled = trackingAlertInput.Enabled.ValueBool()
				input.Rule.Tracking.Alert.Frequency = (cato_models.PolicyRuleTrackingFrequencyEnum)(trackingAlertInput.Frequency.ValueString())

				// setting tracking alert subscription group
				if !trackingAlertInput.SubscriptionGroup.IsNull() {
					elementsAlertSubscriptionGroupInput := make([]types.Object, 0, len(trackingAlertInput.SubscriptionGroup.Elements()))
					diags = trackingAlertInput.SubscriptionGroup.ElementsAs(ctx, &elementsAlertSubscriptionGroupInput, false)
					resp.Diagnostics.Append(diags...)
					if resp.Diagnostics.HasError() {
						return
					}

					var itemAlertSubscriptionGroupInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Tracking_Alert_SubscriptionGroup
					for _, item := range elementsAlertSubscriptionGroupInput {
						diags = item.As(ctx, &itemAlertSubscriptionGroupInput, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemAlertSubscriptionGroupInput)
						if err != nil {
							resp.Diagnostics.AddError(
								"Object Ref transformation failed",
								err.Error(),
							)
							return
						}

						input.Rule.Tracking.Alert.SubscriptionGroup = append(input.Rule.Tracking.Alert.SubscriptionGroup, &cato_models.SubscriptionGroupRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
				}

				// setting tracking alert webhook
				if !trackingAlertInput.Webhook.IsNull() {
					if !trackingAlertInput.Webhook.IsNull() {
						elementsAlertWebHookInput := make([]types.Object, 0, len(trackingAlertInput.Webhook.Elements()))
						diags = trackingAlertInput.Webhook.ElementsAs(ctx, &elementsAlertWebHookInput, false)
						resp.Diagnostics.Append(diags...)
						if resp.Diagnostics.HasError() {
							return
						}

						var itemAlertWebHookInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Tracking_Alert_SubscriptionGroup
						for _, item := range elementsAlertWebHookInput {
							diags = item.As(ctx, &itemAlertWebHookInput, basetypes.ObjectAsOptions{})
							resp.Diagnostics.Append(diags...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemAlertWebHookInput)
							if err != nil {
								resp.Diagnostics.AddError(
									"Object Ref transformation failed",
									err.Error(),
								)
								return
							}

							input.Rule.Tracking.Alert.Webhook = append(input.Rule.Tracking.Alert.Webhook, &cato_models.SubscriptionWebhookRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}
				}

				// setting tracking alert mailing list
				if !trackingAlertInput.MailingList.IsNull() {
					elementsAlertMailingListInput := make([]types.Object, 0, len(trackingAlertInput.MailingList.Elements()))
					diags = trackingAlertInput.MailingList.ElementsAs(ctx, &elementsAlertMailingListInput, false)
					resp.Diagnostics.Append(diags...)
					if resp.Diagnostics.HasError() {
						return
					}

					var itemAlertMailingListInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Tracking_Alert_SubscriptionGroup
					for _, item := range elementsAlertMailingListInput {
						diags = item.As(ctx, &itemAlertMailingListInput, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemAlertMailingListInput)
						if err != nil {
							resp.Diagnostics.AddError(
								"Object Ref transformation failed",
								err.Error(),
							)
							return
						}

						input.Rule.Tracking.Alert.MailingList = append(input.Rule.Tracking.Alert.MailingList, &cato_models.SubscriptionMailingListRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
				}
			}
		}

		// setting schedule
		input.Rule.Schedule = &cato_models.PolicyScheduleInput{
			ActiveOn: (cato_models.PolicyActiveOnEnum)("ALWAYS"),
		}

		if !ruleInput.Schedule.IsNull() {

			scheduleInput := Policy_Policy_WanFirewall_Policy_Rules_Rule_Schedule{}
			diags = ruleInput.Schedule.As(ctx, &scheduleInput, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			input.Rule.Schedule.ActiveOn = cato_models.PolicyActiveOnEnum(scheduleInput.ActiveOn.ValueString())

			// setting schedule custome time frame
			if !scheduleInput.CustomTimeframe.IsNull() {
				input.Rule.Schedule.CustomTimeframe = &cato_models.PolicyCustomTimeframeInput{}

				customeTimeFrameInput := Policy_Policy_WanFirewall_Policy_Rules_Rule_Schedule_CustomTimeframe{}
				diags = scheduleInput.CustomTimeframe.As(ctx, &customeTimeFrameInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}

				input.Rule.Schedule.CustomTimeframe.From = customeTimeFrameInput.From.ValueString()
				input.Rule.Schedule.CustomTimeframe.To = customeTimeFrameInput.To.ValueString()

			}

			// setting schedule custom recurring
			if !scheduleInput.CustomRecurring.IsNull() {
				input.Rule.Schedule.CustomRecurring = &cato_models.PolicyCustomRecurringInput{}

				customRecurringInput := Policy_Policy_WanFirewall_Policy_Rules_Rule_Schedule_CustomRecurring{}
				diags = scheduleInput.CustomRecurring.As(ctx, &customRecurringInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}

				input.Rule.Schedule.CustomRecurring.From = cato_scalars.Time(customRecurringInput.From.ValueString())
				input.Rule.Schedule.CustomRecurring.To = cato_scalars.Time(customRecurringInput.To.ValueString())

				// setting schedule custom recurring days
				diags = customRecurringInput.Days.ElementsAs(ctx, &input.Rule.Schedule.CustomRecurring.Days, false)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
			}
		}

		// settings exceptions
		if !ruleInput.Exceptions.IsNull() {
			elementsExceptionsInput := make([]types.Object, 0, len(ruleInput.Exceptions.Elements()))
			diags = ruleInput.Exceptions.ElementsAs(ctx, &elementsExceptionsInput, false)
			resp.Diagnostics.Append(diags...)

			// loop over exceptions
			var itemExceptionsInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Exceptions
			for _, item := range elementsExceptionsInput {

				exceptionInput := cato_models.WanFirewallRuleExceptionInput{}

				diags = item.As(ctx, &itemExceptionsInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				// setting exception name
				exceptionInput.Name = itemExceptionsInput.Name.ValueString()

				// setting exception direction
				exceptionInput.Direction = cato_models.WanFirewallDirectionEnum(itemExceptionsInput.Direction.ValueString())

				// setting exception connection origin
				if !itemExceptionsInput.ConnectionOrigin.IsNull() {
					exceptionInput.ConnectionOrigin = cato_models.ConnectionOriginEnum(itemExceptionsInput.ConnectionOrigin.ValueString())
				} else {
					exceptionInput.ConnectionOrigin = cato_models.ConnectionOriginEnum("ANY")
				}

				// setting source
				if !itemExceptionsInput.Source.IsNull() {

					exceptionInput.Source = &cato_models.WanFirewallSourceInput{}
					sourceInput := Policy_Policy_WanFirewall_Policy_Rules_Rule_Source{}
					diags = itemExceptionsInput.Source.As(ctx, &sourceInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					// setting source IP
					if !sourceInput.IP.IsNull() {
						diags = sourceInput.IP.ElementsAs(ctx, &exceptionInput.Source.IP, false)
						resp.Diagnostics.Append(diags...)
					}

					// setting source subnet
					if !sourceInput.Subnet.IsNull() {
						diags = sourceInput.Subnet.ElementsAs(ctx, &exceptionInput.Source.Subnet, false)
						resp.Diagnostics.Append(diags...)
					}

					// setting source host
					if !sourceInput.Host.IsNull() {
						elementsSourceHostInput := make([]types.Object, 0, len(sourceInput.Host.Elements()))
						diags = sourceInput.Host.ElementsAs(ctx, &elementsSourceHostInput, false)
						resp.Diagnostics.Append(diags...)

						var itemSourceHostInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_Host
						for _, item := range elementsSourceHostInput {
							diags = item.As(ctx, &itemSourceHostInput, basetypes.ObjectAsOptions{})
							resp.Diagnostics.Append(diags...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceHostInput)
							if err != nil {
								resp.Diagnostics.AddError(
									"Object Ref transformation failed",
									err.Error(),
								)
								return
							}

							exceptionInput.Source.Host = append(exceptionInput.Source.Host, &cato_models.HostRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting source site
					if !sourceInput.Site.IsNull() {
						elementsSourceSiteInput := make([]types.Object, 0, len(sourceInput.Site.Elements()))
						diags = sourceInput.Site.ElementsAs(ctx, &elementsSourceSiteInput, false)
						resp.Diagnostics.Append(diags...)

						var itemSourceSiteInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_Site
						for _, item := range elementsSourceSiteInput {
							diags = item.As(ctx, &itemSourceSiteInput, basetypes.ObjectAsOptions{})
							resp.Diagnostics.Append(diags...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceSiteInput)
							if err != nil {
								resp.Diagnostics.AddError(
									"Object Ref transformation failed",
									err.Error(),
								)
								return
							}

							exceptionInput.Source.Site = append(exceptionInput.Source.Site, &cato_models.SiteRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting source ip range
					if !sourceInput.IPRange.IsNull() {
						elementsSourceIPRangeInput := make([]types.Object, 0, len(sourceInput.IPRange.Elements()))
						diags = sourceInput.IPRange.ElementsAs(ctx, &elementsSourceIPRangeInput, false)
						resp.Diagnostics.Append(diags...)

						var itemSourceIPRangeInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_IPRange
						for _, item := range elementsSourceIPRangeInput {
							diags = item.As(ctx, &itemSourceIPRangeInput, basetypes.ObjectAsOptions{})
							resp.Diagnostics.Append(diags...)

							exceptionInput.Source.IPRange = append(exceptionInput.Source.IPRange, &cato_models.IPAddressRangeInput{
								From: itemSourceIPRangeInput.From.ValueString(),
								To:   itemSourceIPRangeInput.To.ValueString(),
							})
						}
					}

					// setting source global ip range
					if !sourceInput.GlobalIPRange.IsNull() {
						elementsSourceGlobalIPRangeInput := make([]types.Object, 0, len(sourceInput.GlobalIPRange.Elements()))
						diags = sourceInput.GlobalIPRange.ElementsAs(ctx, &elementsSourceGlobalIPRangeInput, false)
						resp.Diagnostics.Append(diags...)

						var itemSourceGlobalIPRangeInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_GlobalIPRange
						for _, item := range elementsSourceGlobalIPRangeInput {
							diags = item.As(ctx, &itemSourceGlobalIPRangeInput, basetypes.ObjectAsOptions{})
							resp.Diagnostics.Append(diags...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceGlobalIPRangeInput)
							if err != nil {
								resp.Diagnostics.AddError(
									"Object Ref transformation failed for",
									err.Error(),
								)
								return
							}

							exceptionInput.Source.GlobalIPRange = append(exceptionInput.Source.GlobalIPRange, &cato_models.GlobalIPRangeRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting source network interface
					if !sourceInput.NetworkInterface.IsNull() {
						elementsSourceNetworkInterfaceInput := make([]types.Object, 0, len(sourceInput.NetworkInterface.Elements()))
						diags = sourceInput.NetworkInterface.ElementsAs(ctx, &elementsSourceNetworkInterfaceInput, false)
						resp.Diagnostics.Append(diags...)

						var itemSourceNetworkInterfaceInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_NetworkInterface
						for _, item := range elementsSourceNetworkInterfaceInput {
							diags = item.As(ctx, &itemSourceNetworkInterfaceInput, basetypes.ObjectAsOptions{})
							resp.Diagnostics.Append(diags...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceNetworkInterfaceInput)
							if err != nil {
								resp.Diagnostics.AddError(
									"Object Ref transformation failed",
									err.Error(),
								)
								return
							}

							exceptionInput.Source.NetworkInterface = append(exceptionInput.Source.NetworkInterface, &cato_models.NetworkInterfaceRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting source site network subnet
					if !sourceInput.SiteNetworkSubnet.IsNull() {
						elementsSourceSiteNetworkSubnetInput := make([]types.Object, 0, len(sourceInput.SiteNetworkSubnet.Elements()))
						diags = sourceInput.SiteNetworkSubnet.ElementsAs(ctx, &elementsSourceSiteNetworkSubnetInput, false)
						resp.Diagnostics.Append(diags...)

						var itemSourceSiteNetworkSubnetInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_SiteNetworkSubnet
						for _, item := range elementsSourceSiteNetworkSubnetInput {
							diags = item.As(ctx, &itemSourceSiteNetworkSubnetInput, basetypes.ObjectAsOptions{})
							resp.Diagnostics.Append(diags...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceSiteNetworkSubnetInput)
							if err != nil {
								resp.Diagnostics.AddError(
									"Object Ref transformation failed",
									err.Error(),
								)
								return
							}

							exceptionInput.Source.SiteNetworkSubnet = append(exceptionInput.Source.SiteNetworkSubnet, &cato_models.SiteNetworkSubnetRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting source floating subnet
					if !sourceInput.FloatingSubnet.IsNull() {
						elementsSourceFloatingSubnetInput := make([]types.Object, 0, len(sourceInput.FloatingSubnet.Elements()))
						diags = sourceInput.FloatingSubnet.ElementsAs(ctx, &elementsSourceFloatingSubnetInput, false)
						resp.Diagnostics.Append(diags...)

						var itemSourceFloatingSubnetInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_FloatingSubnet
						for _, item := range elementsSourceFloatingSubnetInput {
							diags = item.As(ctx, &itemSourceFloatingSubnetInput, basetypes.ObjectAsOptions{})
							resp.Diagnostics.Append(diags...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceFloatingSubnetInput)
							if err != nil {
								resp.Diagnostics.AddError(
									"Object Ref transformation failed",
									err.Error(),
								)
								return
							}

							exceptionInput.Source.FloatingSubnet = append(exceptionInput.Source.FloatingSubnet, &cato_models.FloatingSubnetRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting source user
					if !sourceInput.User.IsNull() {
						elementsSourceUserInput := make([]types.Object, 0, len(sourceInput.User.Elements()))
						diags = sourceInput.User.ElementsAs(ctx, &elementsSourceUserInput, false)
						resp.Diagnostics.Append(diags...)

						var itemSourceUserInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_User
						for _, item := range elementsSourceUserInput {
							diags = item.As(ctx, &itemSourceUserInput, basetypes.ObjectAsOptions{})
							resp.Diagnostics.Append(diags...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceUserInput)
							if err != nil {
								resp.Diagnostics.AddError(
									"Object Ref transformation failed",
									err.Error(),
								)
								return
							}

							exceptionInput.Source.User = append(exceptionInput.Source.User, &cato_models.UserRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting source users group
					if !sourceInput.UsersGroup.IsNull() {
						elementsSourceUsersGroupInput := make([]types.Object, 0, len(sourceInput.UsersGroup.Elements()))
						diags = sourceInput.UsersGroup.ElementsAs(ctx, &elementsSourceUsersGroupInput, false)
						resp.Diagnostics.Append(diags...)

						var itemSourceUsersGroupInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_UsersGroup
						for _, item := range elementsSourceUsersGroupInput {
							diags = item.As(ctx, &itemSourceUsersGroupInput, basetypes.ObjectAsOptions{})
							resp.Diagnostics.Append(diags...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceUsersGroupInput)
							if err != nil {
								resp.Diagnostics.AddError(
									"Object Ref transformation failed",
									err.Error(),
								)
								return
							}

							exceptionInput.Source.UsersGroup = append(exceptionInput.Source.UsersGroup, &cato_models.UsersGroupRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting source group
					if !sourceInput.Group.IsNull() {
						elementsSourceGroupInput := make([]types.Object, 0, len(sourceInput.Group.Elements()))
						diags = sourceInput.Group.ElementsAs(ctx, &elementsSourceGroupInput, false)
						resp.Diagnostics.Append(diags...)

						var itemSourceGroupInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_Group
						for _, item := range elementsSourceGroupInput {
							diags = item.As(ctx, &itemSourceGroupInput, basetypes.ObjectAsOptions{})
							resp.Diagnostics.Append(diags...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceGroupInput)
							if err != nil {
								resp.Diagnostics.AddError(
									"Object Ref transformation failed",
									err.Error(),
								)
								return
							}

							exceptionInput.Source.Group = append(exceptionInput.Source.Group, &cato_models.GroupRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting source system group
					if !sourceInput.SystemGroup.IsNull() {
						elementsSourceSystemGroupInput := make([]types.Object, 0, len(sourceInput.SystemGroup.Elements()))
						diags = sourceInput.SystemGroup.ElementsAs(ctx, &elementsSourceSystemGroupInput, false)
						resp.Diagnostics.Append(diags...)

						var itemSourceSystemGroupInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_SystemGroup
						for _, item := range elementsSourceSystemGroupInput {
							diags = item.As(ctx, &itemSourceSystemGroupInput, basetypes.ObjectAsOptions{})
							resp.Diagnostics.Append(diags...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceSystemGroupInput)
							if err != nil {
								resp.Diagnostics.AddError(
									"Object Ref transformation failed",
									err.Error(),
								)
								return
							}

							exceptionInput.Source.SystemGroup = append(exceptionInput.Source.SystemGroup, &cato_models.SystemGroupRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}
				}

				// setting country
				if !itemExceptionsInput.Country.IsNull() {

					exceptionInput.Country = []*cato_models.CountryRefInput{}
					elementsCountryInput := make([]types.Object, 0, len(itemExceptionsInput.Country.Elements()))
					diags = itemExceptionsInput.Country.ElementsAs(ctx, &elementsCountryInput, false)
					resp.Diagnostics.Append(diags...)

					var itemCountryInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Country
					for _, item := range elementsCountryInput {
						diags = item.As(ctx, &itemCountryInput, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemCountryInput)
						if err != nil {
							resp.Diagnostics.AddError(
								"Object Ref transformation failed",
								err.Error(),
							)
							return
						}

						exceptionInput.Country = append(exceptionInput.Country, &cato_models.CountryRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
				}

				// setting device
				if !itemExceptionsInput.Device.IsNull() {

					exceptionInput.Device = []*cato_models.DeviceProfileRefInput{}
					elementsDeviceInput := make([]types.Object, 0, len(itemExceptionsInput.Device.Elements()))
					diags = itemExceptionsInput.Device.ElementsAs(ctx, &elementsDeviceInput, false)
					resp.Diagnostics.Append(diags...)

					var itemDeviceInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Device
					for _, item := range elementsDeviceInput {
						diags = item.As(ctx, &itemDeviceInput, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemDeviceInput)
						if err != nil {
							resp.Diagnostics.AddError(
								"Object Ref transformation failed",
								err.Error(),
							)
							return
						}

						exceptionInput.Device = append(exceptionInput.Device, &cato_models.DeviceProfileRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
				}

				// setting device OS
				if !itemExceptionsInput.DeviceOs.IsNull() {
					diags = itemExceptionsInput.DeviceOs.ElementsAs(ctx, &exceptionInput.DeviceOs, false)
					resp.Diagnostics.Append(diags...)
					if resp.Diagnostics.HasError() {
						return
					}
				}

				// setting destination
				if !itemExceptionsInput.Destination.IsNull() {

					exceptionInput.Destination = &cato_models.WanFirewallDestinationInput{}
					destinationInput := Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination{}
					diags = itemExceptionsInput.Destination.As(ctx, &destinationInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					// setting destination IP
					if !destinationInput.IP.IsNull() {
						diags = destinationInput.IP.ElementsAs(ctx, &exceptionInput.Destination.IP, false)
						resp.Diagnostics.Append(diags...)
					}

					// setting destination subnet
					if !destinationInput.Subnet.IsNull() {
						diags = destinationInput.Subnet.ElementsAs(ctx, &exceptionInput.Destination.Subnet, false)
						resp.Diagnostics.Append(diags...)
					}

					// setting destination host
					if !destinationInput.Host.IsNull() {
						elementsDestinationHostInput := make([]types.Object, 0, len(destinationInput.Host.Elements()))
						diags = destinationInput.Host.ElementsAs(ctx, &elementsDestinationHostInput, false)
						resp.Diagnostics.Append(diags...)

						var itemDestinationHostInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_Host
						for _, item := range elementsDestinationHostInput {
							diags = item.As(ctx, &itemDestinationHostInput, basetypes.ObjectAsOptions{})
							resp.Diagnostics.Append(diags...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationHostInput)
							if err != nil {
								resp.Diagnostics.AddError(
									"Object Ref transformation failed",
									err.Error(),
								)
								return
							}

							exceptionInput.Destination.Host = append(exceptionInput.Destination.Host, &cato_models.HostRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting destination site
					if !destinationInput.Site.IsNull() {
						elementsDestinationSiteInput := make([]types.Object, 0, len(destinationInput.Site.Elements()))
						diags = destinationInput.Site.ElementsAs(ctx, &elementsDestinationSiteInput, false)
						resp.Diagnostics.Append(diags...)

						var itemDestinationSiteInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_Site
						for _, item := range elementsDestinationSiteInput {
							diags = item.As(ctx, &itemDestinationSiteInput, basetypes.ObjectAsOptions{})
							resp.Diagnostics.Append(diags...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationSiteInput)
							if err != nil {
								resp.Diagnostics.AddError(
									"Object Ref transformation failed",
									err.Error(),
								)
								return
							}

							exceptionInput.Destination.Site = append(exceptionInput.Destination.Site, &cato_models.SiteRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting destination ip range
					if !destinationInput.IPRange.IsNull() {
						elementsDestinationIPRangeInput := make([]types.Object, 0, len(destinationInput.IPRange.Elements()))
						diags = destinationInput.IPRange.ElementsAs(ctx, &elementsDestinationIPRangeInput, false)
						resp.Diagnostics.Append(diags...)

						var itemDestinationIPRangeInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_IPRange
						for _, item := range elementsDestinationIPRangeInput {
							diags = item.As(ctx, &itemDestinationIPRangeInput, basetypes.ObjectAsOptions{})
							resp.Diagnostics.Append(diags...)

							exceptionInput.Destination.IPRange = append(exceptionInput.Destination.IPRange, &cato_models.IPAddressRangeInput{
								From: itemDestinationIPRangeInput.From.ValueString(),
								To:   itemDestinationIPRangeInput.To.ValueString(),
							})
						}
					}

					// setting destination global ip range
					if !destinationInput.GlobalIPRange.IsNull() {
						elementsDestinationGlobalIPRangeInput := make([]types.Object, 0, len(destinationInput.GlobalIPRange.Elements()))
						diags = destinationInput.GlobalIPRange.ElementsAs(ctx, &elementsDestinationGlobalIPRangeInput, false)
						resp.Diagnostics.Append(diags...)

						var itemDestinationGlobalIPRangeInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_GlobalIPRange
						for _, item := range elementsDestinationGlobalIPRangeInput {
							diags = item.As(ctx, &itemDestinationGlobalIPRangeInput, basetypes.ObjectAsOptions{})
							resp.Diagnostics.Append(diags...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationGlobalIPRangeInput)
							if err != nil {
								resp.Diagnostics.AddError(
									"Object Ref transformation failed for",
									err.Error(),
								)
								return
							}

							exceptionInput.Destination.GlobalIPRange = append(exceptionInput.Destination.GlobalIPRange, &cato_models.GlobalIPRangeRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting destination network interface
					if !destinationInput.NetworkInterface.IsNull() {
						elementsDestinationNetworkInterfaceInput := make([]types.Object, 0, len(destinationInput.NetworkInterface.Elements()))
						diags = destinationInput.NetworkInterface.ElementsAs(ctx, &elementsDestinationNetworkInterfaceInput, false)
						resp.Diagnostics.Append(diags...)

						var itemDestinationNetworkInterfaceInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_NetworkInterface
						for _, item := range elementsDestinationNetworkInterfaceInput {
							diags = item.As(ctx, &itemDestinationNetworkInterfaceInput, basetypes.ObjectAsOptions{})
							resp.Diagnostics.Append(diags...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationNetworkInterfaceInput)
							if err != nil {
								resp.Diagnostics.AddError(
									"Object Ref transformation failed",
									err.Error(),
								)
								return
							}

							exceptionInput.Destination.NetworkInterface = append(exceptionInput.Destination.NetworkInterface, &cato_models.NetworkInterfaceRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting destination site network subnet
					if !destinationInput.SiteNetworkSubnet.IsNull() {
						elementsDestinationSiteNetworkSubnetInput := make([]types.Object, 0, len(destinationInput.SiteNetworkSubnet.Elements()))
						diags = destinationInput.SiteNetworkSubnet.ElementsAs(ctx, &elementsDestinationSiteNetworkSubnetInput, false)
						resp.Diagnostics.Append(diags...)

						var itemDestinationSiteNetworkSubnetInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_SiteNetworkSubnet
						for _, item := range elementsDestinationSiteNetworkSubnetInput {
							diags = item.As(ctx, &itemDestinationSiteNetworkSubnetInput, basetypes.ObjectAsOptions{})
							resp.Diagnostics.Append(diags...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationSiteNetworkSubnetInput)
							if err != nil {
								resp.Diagnostics.AddError(
									"Object Ref transformation failed",
									err.Error(),
								)
								return
							}

							exceptionInput.Destination.SiteNetworkSubnet = append(exceptionInput.Destination.SiteNetworkSubnet, &cato_models.SiteNetworkSubnetRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting destination floating subnet
					if !destinationInput.FloatingSubnet.IsNull() {
						elementsDestinationFloatingSubnetInput := make([]types.Object, 0, len(destinationInput.FloatingSubnet.Elements()))
						diags = destinationInput.FloatingSubnet.ElementsAs(ctx, &elementsDestinationFloatingSubnetInput, false)
						resp.Diagnostics.Append(diags...)

						var itemDestinationFloatingSubnetInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_FloatingSubnet
						for _, item := range elementsDestinationFloatingSubnetInput {
							diags = item.As(ctx, &itemDestinationFloatingSubnetInput, basetypes.ObjectAsOptions{})
							resp.Diagnostics.Append(diags...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationFloatingSubnetInput)
							if err != nil {
								resp.Diagnostics.AddError(
									"Object Ref transformation failed",
									err.Error(),
								)
								return
							}

							exceptionInput.Destination.FloatingSubnet = append(exceptionInput.Destination.FloatingSubnet, &cato_models.FloatingSubnetRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting destination user
					if !destinationInput.User.IsNull() {
						elementsDestinationUserInput := make([]types.Object, 0, len(destinationInput.User.Elements()))
						diags = destinationInput.User.ElementsAs(ctx, &elementsDestinationUserInput, false)
						resp.Diagnostics.Append(diags...)

						var itemDestinationUserInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_User
						for _, item := range elementsDestinationUserInput {
							diags = item.As(ctx, &itemDestinationUserInput, basetypes.ObjectAsOptions{})
							resp.Diagnostics.Append(diags...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationUserInput)
							if err != nil {
								resp.Diagnostics.AddError(
									"Object Ref transformation failed",
									err.Error(),
								)
								return
							}

							exceptionInput.Destination.User = append(exceptionInput.Destination.User, &cato_models.UserRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting destination users group
					if !destinationInput.UsersGroup.IsNull() {
						elementsDestinationUsersGroupInput := make([]types.Object, 0, len(destinationInput.UsersGroup.Elements()))
						diags = destinationInput.UsersGroup.ElementsAs(ctx, &elementsDestinationUsersGroupInput, false)
						resp.Diagnostics.Append(diags...)

						var itemDestinationUsersGroupInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_UsersGroup
						for _, item := range elementsDestinationUsersGroupInput {
							diags = item.As(ctx, &itemDestinationUsersGroupInput, basetypes.ObjectAsOptions{})
							resp.Diagnostics.Append(diags...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationUsersGroupInput)
							if err != nil {
								resp.Diagnostics.AddError(
									"Object Ref transformation failed",
									err.Error(),
								)
								return
							}

							exceptionInput.Destination.UsersGroup = append(exceptionInput.Destination.UsersGroup, &cato_models.UsersGroupRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting destination group
					if !destinationInput.Group.IsNull() {
						elementsDestinationGroupInput := make([]types.Object, 0, len(destinationInput.Group.Elements()))
						diags = destinationInput.Group.ElementsAs(ctx, &elementsDestinationGroupInput, false)
						resp.Diagnostics.Append(diags...)

						var itemDestinationGroupInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_Group
						for _, item := range elementsDestinationGroupInput {
							diags = item.As(ctx, &itemDestinationGroupInput, basetypes.ObjectAsOptions{})
							resp.Diagnostics.Append(diags...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationGroupInput)
							if err != nil {
								resp.Diagnostics.AddError(
									"Object Ref transformation failed",
									err.Error(),
								)
								return
							}

							exceptionInput.Destination.Group = append(exceptionInput.Destination.Group, &cato_models.GroupRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting destination system group
					if !destinationInput.SystemGroup.IsNull() {
						elementsDestinationSystemGroupInput := make([]types.Object, 0, len(destinationInput.SystemGroup.Elements()))
						diags = destinationInput.SystemGroup.ElementsAs(ctx, &elementsDestinationSystemGroupInput, false)
						resp.Diagnostics.Append(diags...)

						var itemDestinationSystemGroupInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_SystemGroup
						for _, item := range elementsDestinationSystemGroupInput {
							diags = item.As(ctx, &itemDestinationSystemGroupInput, basetypes.ObjectAsOptions{})
							resp.Diagnostics.Append(diags...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationSystemGroupInput)
							if err != nil {
								resp.Diagnostics.AddError(
									"Object Ref transformation failed",
									err.Error(),
								)
								return
							}

							exceptionInput.Destination.SystemGroup = append(exceptionInput.Destination.SystemGroup, &cato_models.SystemGroupRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}
				}

				// setting application
				if !itemExceptionsInput.Application.IsNull() {

					exceptionInput.Application = &cato_models.WanFirewallApplicationInput{}
					applicationInput := Policy_Policy_WanFirewall_Policy_Rules_Rule_Application{}
					diags = itemExceptionsInput.Application.As(ctx, &applicationInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					// setting application IP
					if !applicationInput.IP.IsNull() {
						diags = applicationInput.IP.ElementsAs(ctx, &exceptionInput.Application.IP, false)
						resp.Diagnostics.Append(diags...)
					}

					// setting application subnet
					if !applicationInput.Subnet.IsNull() {
						diags = applicationInput.Subnet.ElementsAs(ctx, &exceptionInput.Application.Subnet, false)
						resp.Diagnostics.Append(diags...)
					}

					// setting application domain
					if !applicationInput.Domain.IsNull() {
						diags = applicationInput.Domain.ElementsAs(ctx, &exceptionInput.Application.Domain, false)
						resp.Diagnostics.Append(diags...)
					}

					// setting application fqdn
					if !applicationInput.Fqdn.IsNull() {
						diags = applicationInput.Fqdn.ElementsAs(ctx, &exceptionInput.Application.Fqdn, false)
						resp.Diagnostics.Append(diags...)
					}

					// setting application application
					if !applicationInput.Application.IsNull() {
						elementsApplicationApplicationInput := make([]types.Object, 0, len(applicationInput.Application.Elements()))
						diags = applicationInput.Application.ElementsAs(ctx, &elementsApplicationApplicationInput, false)
						resp.Diagnostics.Append(diags...)

						var itemApplicationApplicationInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_Application
						for _, item := range elementsApplicationApplicationInput {
							diags = item.As(ctx, &itemApplicationApplicationInput, basetypes.ObjectAsOptions{})
							resp.Diagnostics.Append(diags...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationApplicationInput)
							if err != nil {
								resp.Diagnostics.AddError(
									"Object Ref transformation failed",
									err.Error(),
								)
								return
							}

							exceptionInput.Application.Application = append(exceptionInput.Application.Application, &cato_models.ApplicationRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting application custom app
					if !applicationInput.CustomApp.IsNull() {
						elementsApplicationCustomAppInput := make([]types.Object, 0, len(applicationInput.CustomApp.Elements()))
						diags = applicationInput.CustomApp.ElementsAs(ctx, &elementsApplicationCustomAppInput, false)
						resp.Diagnostics.Append(diags...)

						var itemApplicationCustomAppInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_CustomApp
						for _, item := range elementsApplicationCustomAppInput {
							diags = item.As(ctx, &itemApplicationCustomAppInput, basetypes.ObjectAsOptions{})
							resp.Diagnostics.Append(diags...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationCustomAppInput)
							if err != nil {
								resp.Diagnostics.AddError(
									"Object Ref transformation failed",
									err.Error(),
								)
								return
							}

							exceptionInput.Application.CustomApp = append(exceptionInput.Application.CustomApp, &cato_models.CustomApplicationRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting application ip range
					if !applicationInput.IPRange.IsNull() {
						elementsApplicationIPRangeInput := make([]types.Object, 0, len(applicationInput.IPRange.Elements()))
						diags = applicationInput.IPRange.ElementsAs(ctx, &elementsApplicationIPRangeInput, false)
						resp.Diagnostics.Append(diags...)

						var itemApplicationIPRangeInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_IPRange
						for _, item := range elementsApplicationIPRangeInput {
							diags = item.As(ctx, &itemApplicationIPRangeInput, basetypes.ObjectAsOptions{})
							resp.Diagnostics.Append(diags...)

							exceptionInput.Application.IPRange = append(exceptionInput.Application.IPRange, &cato_models.IPAddressRangeInput{
								From: itemApplicationIPRangeInput.From.ValueString(),
								To:   itemApplicationIPRangeInput.To.ValueString(),
							})
						}
					}

					// setting application global ip range
					if !applicationInput.GlobalIPRange.IsNull() {
						elementsApplicationGlobalIPRangeInput := make([]types.Object, 0, len(applicationInput.GlobalIPRange.Elements()))
						diags = applicationInput.GlobalIPRange.ElementsAs(ctx, &elementsApplicationGlobalIPRangeInput, false)
						resp.Diagnostics.Append(diags...)

						var itemApplicationGlobalIPRangeInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_GlobalIPRange
						for _, item := range elementsApplicationGlobalIPRangeInput {
							diags = item.As(ctx, &itemApplicationGlobalIPRangeInput, basetypes.ObjectAsOptions{})
							resp.Diagnostics.Append(diags...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationGlobalIPRangeInput)
							if err != nil {
								resp.Diagnostics.AddError(
									"Object Ref transformation failed",
									err.Error(),
								)
								return
							}

							exceptionInput.Application.GlobalIPRange = append(exceptionInput.Application.GlobalIPRange, &cato_models.GlobalIPRangeRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting application app category
					if !applicationInput.AppCategory.IsNull() {
						elementsApplicationAppCategoryInput := make([]types.Object, 0, len(applicationInput.AppCategory.Elements()))
						diags = applicationInput.AppCategory.ElementsAs(ctx, &elementsApplicationAppCategoryInput, false)
						resp.Diagnostics.Append(diags...)

						var itemApplicationAppCategoryInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_AppCategory
						for _, item := range elementsApplicationAppCategoryInput {
							diags = item.As(ctx, &itemApplicationAppCategoryInput, basetypes.ObjectAsOptions{})
							resp.Diagnostics.Append(diags...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationAppCategoryInput)
							if err != nil {
								resp.Diagnostics.AddError(
									"Object Ref transformation failed",
									err.Error(),
								)
								return
							}

							exceptionInput.Application.AppCategory = append(exceptionInput.Application.AppCategory, &cato_models.ApplicationCategoryRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting application custom app category
					if !applicationInput.CustomCategory.IsNull() {
						elementsApplicationCustomCategoryInput := make([]types.Object, 0, len(applicationInput.CustomCategory.Elements()))
						diags = applicationInput.CustomCategory.ElementsAs(ctx, &elementsApplicationCustomCategoryInput, false)
						resp.Diagnostics.Append(diags...)

						var itemApplicationCustomCategoryInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_CustomCategory
						for _, item := range elementsApplicationCustomCategoryInput {
							diags = item.As(ctx, &itemApplicationCustomCategoryInput, basetypes.ObjectAsOptions{})
							resp.Diagnostics.Append(diags...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationCustomCategoryInput)
							if err != nil {
								resp.Diagnostics.AddError(
									"Object Ref transformation failed",
									err.Error(),
								)
								return
							}

							exceptionInput.Application.CustomCategory = append(exceptionInput.Application.CustomCategory, &cato_models.CustomCategoryRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting application sanctionned apps category
					if !applicationInput.SanctionedAppsCategory.IsNull() {
						elementsApplicationSanctionedAppsCategoryInput := make([]types.Object, 0, len(applicationInput.SanctionedAppsCategory.Elements()))
						diags = applicationInput.SanctionedAppsCategory.ElementsAs(ctx, &elementsApplicationSanctionedAppsCategoryInput, false)
						resp.Diagnostics.Append(diags...)

						var itemApplicationSanctionedAppsCategoryInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_SanctionedAppsCategory
						for _, item := range elementsApplicationSanctionedAppsCategoryInput {
							diags = item.As(ctx, &itemApplicationSanctionedAppsCategoryInput, basetypes.ObjectAsOptions{})
							resp.Diagnostics.Append(diags...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationSanctionedAppsCategoryInput)
							if err != nil {
								resp.Diagnostics.AddError(
									"Object Ref transformation failed",
									err.Error(),
								)
								return
							}

							exceptionInput.Application.SanctionedAppsCategory = append(exceptionInput.Application.SanctionedAppsCategory, &cato_models.SanctionedAppsCategoryRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}
				}

				// setting service
				if !itemExceptionsInput.Service.IsNull() {

					exceptionInput.Service = &cato_models.WanFirewallServiceTypeInput{}
					serviceInput := Policy_Policy_WanFirewall_Policy_Rules_Rule_Service{}
					diags = itemExceptionsInput.Service.As(ctx, &serviceInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)
					if resp.Diagnostics.HasError() {
						return
					}

					// setting service standard
					if !serviceInput.Standard.IsNull() {
						elementsServiceStandardInput := make([]types.Object, 0, len(serviceInput.Standard.Elements()))
						diags = serviceInput.Standard.ElementsAs(ctx, &elementsServiceStandardInput, false)
						resp.Diagnostics.Append(diags...)
						if resp.Diagnostics.HasError() {
							return
						}

						var itemServiceStandardInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Service_Standard
						for _, item := range elementsServiceStandardInput {
							diags = item.As(ctx, &itemServiceStandardInput, basetypes.ObjectAsOptions{})
							resp.Diagnostics.Append(diags...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemServiceStandardInput)
							if err != nil {
								resp.Diagnostics.AddError(
									"Object Ref transformation failed",
									err.Error(),
								)
								return
							}

							exceptionInput.Service.Standard = append(exceptionInput.Service.Standard, &cato_models.ServiceRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting service custom
					if !serviceInput.Custom.IsNull() {
						elementsServiceCustomInput := make([]types.Object, 0, len(serviceInput.Custom.Elements()))
						diags = serviceInput.Custom.ElementsAs(ctx, &elementsServiceCustomInput, false)
						resp.Diagnostics.Append(diags...)
						if resp.Diagnostics.HasError() {
							return
						}

						var itemServiceCustomInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Service_Custom
						for _, item := range elementsServiceCustomInput {
							diags = item.As(ctx, &itemServiceCustomInput, basetypes.ObjectAsOptions{})

							customInput := &cato_models.CustomServiceInput{
								Protocol: cato_models.IPProtocol(itemServiceCustomInput.Protocol.ValueString()),
							}

							// setting service custom port
							if !itemServiceCustomInput.Port.IsNull() {
								elementsPort := make([]types.String, 0, len(itemServiceCustomInput.Port.Elements()))
								diags = itemServiceCustomInput.Port.ElementsAs(ctx, &elementsPort, false)
								resp.Diagnostics.Append(diags...)

								inputPort := []cato_scalars.Port{}
								for _, item := range elementsPort {
									inputPort = append(inputPort, cato_scalars.Port(item.ValueString()))
								}

								customInput.Port = inputPort
							}

							// setting service custom port range
							if !itemServiceCustomInput.PortRange.IsNull() {
								var itemPortRange Policy_Policy_WanFirewall_Policy_Rules_Rule_Service_Custom_PortRange
								diags = itemServiceCustomInput.PortRange.As(ctx, &itemPortRange, basetypes.ObjectAsOptions{})

								inputPortRange := cato_models.PortRangeInput{
									From: cato_scalars.Port(itemPortRange.From.ValueString()),
									To:   cato_scalars.Port(itemPortRange.To.ValueString()),
								}

								customInput.PortRange = &inputPortRange
							}

							// append custom service
							exceptionInput.Service.Custom = append(exceptionInput.Service.Custom, customInput)
						}
					}
				}

				input.Rule.Exceptions = append(input.Rule.Exceptions, &exceptionInput)

			}
		}

		// settings other rule attributes
		input.Rule.Name = ruleInput.Name.ValueString()
		input.Rule.Description = ruleInput.Description.ValueString()
		input.Rule.Enabled = ruleInput.Enabled.ValueBool()
		input.Rule.Action = cato_models.WanFirewallActionEnum(ruleInput.Action.ValueString())
		input.Rule.Direction = cato_models.WanFirewallDirectionEnum(ruleInput.Direction.ValueString())
		if !ruleInput.ConnectionOrigin.IsNull() {
			input.Rule.ConnectionOrigin = cato_models.ConnectionOriginEnum(ruleInput.ConnectionOrigin.ValueString())
		} else {
			input.Rule.ConnectionOrigin = "ANY"
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "wan_fw_policy create", map[string]interface{}{
		"input": utils.InterfaceToJSONString(input),
	})

	//creating new rule
	policyChange, err := r.client.catov2.PolicyWanFirewallAddRule(ctx, input, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanFirewallAddRule error",
			err.Error(),
		)
		return
	}

	// check for errors
	if policyChange.Policy.WanFirewall.AddRule.Status != "SUCCESS" {
		for _, item := range policyChange.Policy.WanFirewall.AddRule.GetErrors() {
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
	_, err = r.client.catov2.PolicyWanFirewallPublishPolicyRevision(ctx, publishDataIfEnabled, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanFirewallPublishPolicyRevision error",
			err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// overiding state with rule id
	resp.State.SetAttribute(
		ctx,
		path.Root("rule").AtName("id"),
		policyChange.GetPolicy().GetWanFirewall().GetAddRule().Rule.GetRule().ID)
}

func (r *wanFwRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Warn(ctx, "Read(1)")
	var state WanFirewallRule
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Warn(ctx, "Read(2)")

	// body, err := r.client.catov2.Policy(ctx, &cato_models.WanFirewallPolicyInput{}, &cato_models.WanFirewallPolicyInput{}, r.client.AccountId)
	queryWanPolicy := &cato_models.WanFirewallPolicyInput{}
	body, err := r.client.catov2.PolicyWanFirewall(ctx, queryWanPolicy, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanFirewall error",
			err.Error(),
		)
		return
	}
	tflog.Warn(ctx, "Read(3)")

	//retrieve rule ID
	rule := Policy_Policy_WanFirewall_Policy_Rules_Rule{}
	diags = state.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(diags...)
	tflog.Warn(ctx, "Read(4)")

	ruleList := body.GetPolicy().WanFirewall.Policy.GetRules()
	ruleExist := false
	currentRule := &cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules_Rule{}
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
		tflog.Warn(ctx, "wan firewall rule not found, resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	ruleInput := hydrateWanRuleState(ctx, state, currentRule)
	tflog.Info(ctx, "ruleInput - "+fmt.Sprintf("%v", ruleInput))
	// tflog.Info(ctx, "ruleInput.ID", map[string]interface{}{
	// 	"value":      ruleInput.ID,
	// 	"is_null":    ruleInput.ID.IsNull(),
	// 	"is_unknown": ruleInput.ID.IsUnknown(),
	// })
	// tflog.Info(ctx, "ruleInput.Name", map[string]interface{}{
	// 	"value":      ruleInput.Name,
	// 	"is_null":    ruleInput.Name.IsNull(),
	// 	"is_unknown": ruleInput.Name.IsUnknown(),
	// })
	// tflog.Info(ctx, "ruleInput.Description", map[string]interface{}{
	// 	"value":      ruleInput.Description,
	// 	"is_null":    ruleInput.Description.IsNull(),
	// 	"is_unknown": ruleInput.Description.IsUnknown(),
	// })
	// tflog.Info(ctx, "ruleInput.Index", map[string]interface{}{
	// 	"value":      ruleInput.Index,
	// 	"is_null":    ruleInput.Index.IsNull(),
	// 	"is_unknown": ruleInput.Index.IsUnknown(),
	// })
	// tflog.Info(ctx, "ruleInput.Enabled", map[string]interface{}{
	// 	"value":      ruleInput.Enabled,
	// 	"is_null":    ruleInput.Enabled.IsNull(),
	// 	"is_unknown": ruleInput.Enabled.IsUnknown(),
	// })
	// // tflog.Info(ctx, "ruleInput.Section", map[string]interface{}{
	// // 	"value":      ruleInput.Section,
	// // 	"is_null":    ruleInput.Section.IsNull(),
	// // 	"is_unknown": ruleInput.Section.IsUnknown(),
	// // })
	// tflog.Info(ctx, "ruleInput.Source", map[string]interface{}{
	// 	"value":      ruleInput.Source,
	// 	"is_null":    ruleInput.Source.IsNull(),
	// 	"is_unknown": ruleInput.Source.IsUnknown(),
	// })
	// tflog.Info(ctx, "ruleInput.ConnectionOrigin", map[string]interface{}{
	// 	"value":      ruleInput.ConnectionOrigin,
	// 	"is_null":    ruleInput.ConnectionOrigin.IsNull(),
	// 	"is_unknown": ruleInput.ConnectionOrigin.IsUnknown(),
	// })
	// tflog.Info(ctx, "ruleInput.Country", map[string]interface{}{
	// 	"value":      ruleInput.Country,
	// 	"is_null":    ruleInput.Country.IsNull(),
	// 	"is_unknown": ruleInput.Country.IsUnknown(),
	// })
	// tflog.Info(ctx, "ruleInput.Device", map[string]interface{}{
	// 	"value":      ruleInput.Device,
	// 	"is_null":    ruleInput.Device.IsNull(),
	// 	"is_unknown": ruleInput.Device.IsUnknown(),
	// })
	// tflog.Info(ctx, "ruleInput.DeviceOs", map[string]interface{}{
	// 	"value":      ruleInput.DeviceOs,
	// 	"is_null":    ruleInput.DeviceOs.IsNull(),
	// 	"is_unknown": ruleInput.DeviceOs.IsUnknown(),
	// })
	// tflog.Info(ctx, "ruleInput.Application", map[string]interface{}{
	// 	"value":      ruleInput.Application,
	// 	"is_null":    ruleInput.Application.IsNull(),
	// 	"is_unknown": ruleInput.Application.IsUnknown(),
	// })
	// tflog.Info(ctx, "ruleInput.Destination", map[string]interface{}{
	// 	"value":      ruleInput.Destination,
	// 	"is_null":    ruleInput.Destination.IsNull(),
	// 	"is_unknown": ruleInput.Destination.IsUnknown(),
	// })
	// tflog.Info(ctx, "ruleInput.Service", map[string]interface{}{
	// 	"value":      ruleInput.Service,
	// 	"is_null":    ruleInput.Service.IsNull(),
	// 	"is_unknown": ruleInput.Service.IsUnknown(),
	// })
	// tflog.Info(ctx, "ruleInput.Action", map[string]interface{}{
	// 	"value":      ruleInput.Action,
	// 	"is_null":    ruleInput.Action.IsNull(),
	// 	"is_unknown": ruleInput.Action.IsUnknown(),
	// })
	// tflog.Info(ctx, "ruleInput.Tracking", map[string]interface{}{
	// 	"value":      ruleInput.Tracking,
	// 	"is_null":    ruleInput.Tracking.IsNull(),
	// 	"is_unknown": ruleInput.Tracking.IsUnknown(),
	// })
	// tflog.Info(ctx, "ruleInput.Schedule", map[string]interface{}{
	// 	"value":      ruleInput.Schedule,
	// 	"is_null":    ruleInput.Schedule.IsNull(),
	// 	"is_unknown": ruleInput.Schedule.IsUnknown(),
	// })
	// tflog.Info(ctx, "ruleInput.Direction", map[string]interface{}{
	// 	"value":      ruleInput.Direction,
	// 	"is_null":    ruleInput.Direction.IsNull(),
	// 	"is_unknown": ruleInput.Direction.IsUnknown(),
	// })
	// tflog.Info(ctx, "ruleInput.Exceptions", map[string]interface{}{
	// 	"value":      ruleInput.Exceptions,
	// 	"is_null":    ruleInput.Exceptions.IsNull(),
	// 	"is_unknown": ruleInput.Exceptions.IsUnknown(),
	// })

	// tflog.Info(ctx, "ruleInput.Source", map[string]interface{}{"attrs": ruleInput.Source.Attributes()})
	// tflog.Info(ctx, "ruleInput.Destination", map[string]interface{}{"attrs": ruleInput.Destination.Attributes()})
	// tflog.Info(ctx, "ruleInput.Application", map[string]interface{}{"attrs": ruleInput.Application.Attributes()})
	diags = resp.State.SetAttribute(ctx, path.Root("rule"), ruleInput)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(diags...)
	tflog.Warn(ctx, "Read(7)")

}

func (r *wanFwRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	var plan WanFirewallRule
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := cato_models.WanFirewallUpdateRuleInput{
		Rule: &cato_models.WanFirewallUpdateRuleDataInput{
			Source: &cato_models.WanFirewallSourceUpdateInput{
				IP:                []string{},
				Host:              []*cato_models.HostRefInput{},
				Site:              []*cato_models.SiteRefInput{},
				Subnet:            []string{},
				IPRange:           []*cato_models.IPAddressRangeInput{},
				GlobalIPRange:     []*cato_models.GlobalIPRangeRefInput{},
				NetworkInterface:  []*cato_models.NetworkInterfaceRefInput{},
				SiteNetworkSubnet: []*cato_models.SiteNetworkSubnetRefInput{},
				FloatingSubnet:    []*cato_models.FloatingSubnetRefInput{},
				User:              []*cato_models.UserRefInput{},
				UsersGroup:        []*cato_models.UsersGroupRefInput{},
				Group:             []*cato_models.GroupRefInput{},
				SystemGroup:       []*cato_models.SystemGroupRefInput{},
			},
			Country: []*cato_models.CountryRefInput{},
			Destination: &cato_models.WanFirewallDestinationUpdateInput{
				IP:                []string{},
				Host:              []*cato_models.HostRefInput{},
				Site:              []*cato_models.SiteRefInput{},
				Subnet:            []string{},
				IPRange:           []*cato_models.IPAddressRangeInput{},
				GlobalIPRange:     []*cato_models.GlobalIPRangeRefInput{},
				NetworkInterface:  []*cato_models.NetworkInterfaceRefInput{},
				SiteNetworkSubnet: []*cato_models.SiteNetworkSubnetRefInput{},
				FloatingSubnet:    []*cato_models.FloatingSubnetRefInput{},
				User:              []*cato_models.UserRefInput{},
				UsersGroup:        []*cato_models.UsersGroupRefInput{},
				Group:             []*cato_models.GroupRefInput{},
				SystemGroup:       []*cato_models.SystemGroupRefInput{},
			},
			Device:   []*cato_models.DeviceProfileRefInput{},
			DeviceOs: []cato_models.OperatingSystem{},
			Application: &cato_models.WanFirewallApplicationUpdateInput{
				Application:            []*cato_models.ApplicationRefInput{},
				CustomApp:              []*cato_models.CustomApplicationRefInput{},
				AppCategory:            []*cato_models.ApplicationCategoryRefInput{},
				CustomCategory:         []*cato_models.CustomCategoryRefInput{},
				SanctionedAppsCategory: []*cato_models.SanctionedAppsCategoryRefInput{},
				Domain:                 []string{},
				Fqdn:                   []string{},
				IP:                     []string{},
				Subnet:                 []string{},
				IPRange:                []*cato_models.IPAddressRangeInput{},
				GlobalIPRange:          []*cato_models.GlobalIPRangeRefInput{},
			},
			Service: &cato_models.WanFirewallServiceTypeUpdateInput{
				Standard: []*cato_models.ServiceRefInput{},
				Custom:   []*cato_models.CustomServiceInput{},
			},
			Tracking: &cato_models.PolicyTrackingUpdateInput{
				Event: &cato_models.PolicyRuleTrackingEventUpdateInput{},
				Alert: &cato_models.PolicyRuleTrackingAlertUpdateInput{
					SubscriptionGroup: []*cato_models.SubscriptionGroupRefInput{},
					Webhook:           []*cato_models.SubscriptionWebhookRefInput{},
					MailingList:       []*cato_models.SubscriptionMailingListRefInput{},
				},
			},
			Schedule: &cato_models.PolicyScheduleUpdateInput{
				CustomTimeframe: &cato_models.PolicyCustomTimeframeUpdateInput{},
				CustomRecurring: &cato_models.PolicyCustomRecurringUpdateInput{},
			},
			Exceptions: []*cato_models.WanFirewallRuleExceptionInput{},
		},
	}

	// setting input for moving rule
	inputMoveRule := cato_models.PolicyMoveRuleInput{}

	// setting rule
	ruleInput := Policy_Policy_WanFirewall_Policy_Rules_Rule{}
	diags = plan.Rule.As(ctx, &ruleInput, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)

	// setting source
	if !ruleInput.Source.IsNull() {
		sourceInput := Policy_Policy_WanFirewall_Policy_Rules_Rule_Source{}
		diags = ruleInput.Source.As(ctx, &sourceInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)

		// setting source IP
		if !sourceInput.IP.IsNull() {
			diags = sourceInput.IP.ElementsAs(ctx, &input.Rule.Source.IP, false)
			resp.Diagnostics.Append(diags...)
		}

		// setting source subnet
		if !sourceInput.Subnet.IsNull() {
			diags = sourceInput.Subnet.ElementsAs(ctx, &input.Rule.Source.Subnet, false)
			resp.Diagnostics.Append(diags...)
		}

		// setting source host
		if !sourceInput.Host.IsNull() {
			elementsSourceHostInput := make([]types.Object, 0, len(sourceInput.Host.Elements()))
			diags = sourceInput.Host.ElementsAs(ctx, &elementsSourceHostInput, false)
			resp.Diagnostics.Append(diags...)

			var itemSourceHostInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_Host
			for _, item := range elementsSourceHostInput {
				diags = item.As(ctx, &itemSourceHostInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceHostInput)
				if err != nil {
					resp.Diagnostics.AddError(
						"Object Ref transformation failed",
						err.Error(),
					)
					return
				}

				input.Rule.Source.Host = append(input.Rule.Source.Host, &cato_models.HostRefInput{
					By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
					Input: ObjectRefOutput.Input,
				})
			}
		}

		// setting source site
		if !sourceInput.Site.IsNull() {
			elementsSourceSiteInput := make([]types.Object, 0, len(sourceInput.Site.Elements()))
			diags = sourceInput.Site.ElementsAs(ctx, &elementsSourceSiteInput, false)
			resp.Diagnostics.Append(diags...)

			var itemSourceSiteInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_Site
			for _, item := range elementsSourceSiteInput {
				diags = item.As(ctx, &itemSourceSiteInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceSiteInput)
				if err != nil {
					resp.Diagnostics.AddError(
						"Object Ref transformation failed",
						err.Error(),
					)
					return
				}

				input.Rule.Source.Site = append(input.Rule.Source.Site, &cato_models.SiteRefInput{
					By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
					Input: ObjectRefOutput.Input,
				})
			}
		}

		// setting source ip range
		if !sourceInput.IPRange.IsNull() {
			elementsSourceIPRangeInput := make([]types.Object, 0, len(sourceInput.IPRange.Elements()))
			diags = sourceInput.IPRange.ElementsAs(ctx, &elementsSourceIPRangeInput, false)
			resp.Diagnostics.Append(diags...)

			var itemSourceIPRangeInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_IPRange
			for _, item := range elementsSourceIPRangeInput {
				diags = item.As(ctx, &itemSourceIPRangeInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				input.Rule.Source.IPRange = append(input.Rule.Source.IPRange, &cato_models.IPAddressRangeInput{
					From: itemSourceIPRangeInput.From.ValueString(),
					To:   itemSourceIPRangeInput.To.ValueString(),
				})
			}
		}

		// setting source global ip range
		if !sourceInput.GlobalIPRange.IsNull() {
			elementsSourceGlobalIPRangeInput := make([]types.Object, 0, len(sourceInput.GlobalIPRange.Elements()))
			diags = sourceInput.GlobalIPRange.ElementsAs(ctx, &elementsSourceGlobalIPRangeInput, false)
			resp.Diagnostics.Append(diags...)

			var itemSourceGlobalIPRangeInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_GlobalIPRange
			for _, item := range elementsSourceGlobalIPRangeInput {
				diags = item.As(ctx, &itemSourceGlobalIPRangeInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceGlobalIPRangeInput)
				if err != nil {
					resp.Diagnostics.AddError(
						"Object Ref transformation failed for",
						err.Error(),
					)
					return
				}

				input.Rule.Source.GlobalIPRange = append(input.Rule.Source.GlobalIPRange, &cato_models.GlobalIPRangeRefInput{
					By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
					Input: ObjectRefOutput.Input,
				})
			}
		}

		// setting source network interface
		if !sourceInput.NetworkInterface.IsNull() {
			elementsSourceNetworkInterfaceInput := make([]types.Object, 0, len(sourceInput.NetworkInterface.Elements()))
			diags = sourceInput.NetworkInterface.ElementsAs(ctx, &elementsSourceNetworkInterfaceInput, false)
			resp.Diagnostics.Append(diags...)

			var itemSourceNetworkInterfaceInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_NetworkInterface
			for _, item := range elementsSourceNetworkInterfaceInput {
				diags = item.As(ctx, &itemSourceNetworkInterfaceInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceNetworkInterfaceInput)
				if err != nil {
					resp.Diagnostics.AddError(
						"Object Ref transformation failed",
						err.Error(),
					)
					return
				}

				input.Rule.Source.NetworkInterface = append(input.Rule.Source.NetworkInterface, &cato_models.NetworkInterfaceRefInput{
					By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
					Input: ObjectRefOutput.Input,
				})
			}
		}

		// setting source site network subnet
		if !sourceInput.SiteNetworkSubnet.IsNull() {
			elementsSourceSiteNetworkSubnetInput := make([]types.Object, 0, len(sourceInput.SiteNetworkSubnet.Elements()))
			diags = sourceInput.SiteNetworkSubnet.ElementsAs(ctx, &elementsSourceSiteNetworkSubnetInput, false)
			resp.Diagnostics.Append(diags...)

			var itemSourceSiteNetworkSubnetInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_SiteNetworkSubnet
			for _, item := range elementsSourceSiteNetworkSubnetInput {
				diags = item.As(ctx, &itemSourceSiteNetworkSubnetInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceSiteNetworkSubnetInput)
				if err != nil {
					resp.Diagnostics.AddError(
						"Object Ref transformation failed",
						err.Error(),
					)
					return
				}

				input.Rule.Source.SiteNetworkSubnet = append(input.Rule.Source.SiteNetworkSubnet, &cato_models.SiteNetworkSubnetRefInput{
					By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
					Input: ObjectRefOutput.Input,
				})
			}
		}

		// setting source floating subnet
		if !sourceInput.FloatingSubnet.IsNull() {
			elementsSourceFloatingSubnetInput := make([]types.Object, 0, len(sourceInput.FloatingSubnet.Elements()))
			diags = sourceInput.FloatingSubnet.ElementsAs(ctx, &elementsSourceFloatingSubnetInput, false)
			resp.Diagnostics.Append(diags...)

			var itemSourceFloatingSubnetInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_FloatingSubnet
			for _, item := range elementsSourceFloatingSubnetInput {
				diags = item.As(ctx, &itemSourceFloatingSubnetInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceFloatingSubnetInput)
				if err != nil {
					resp.Diagnostics.AddError(
						"Object Ref transformation failed",
						err.Error(),
					)
					return
				}

				input.Rule.Source.FloatingSubnet = append(input.Rule.Source.FloatingSubnet, &cato_models.FloatingSubnetRefInput{
					By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
					Input: ObjectRefOutput.Input,
				})
			}
		}

		// setting source user
		if !sourceInput.User.IsNull() {
			elementsSourceUserInput := make([]types.Object, 0, len(sourceInput.User.Elements()))
			diags = sourceInput.User.ElementsAs(ctx, &elementsSourceUserInput, false)
			resp.Diagnostics.Append(diags...)

			var itemSourceUserInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_User
			for _, item := range elementsSourceUserInput {
				diags = item.As(ctx, &itemSourceUserInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceUserInput)
				if err != nil {
					resp.Diagnostics.AddError(
						"Object Ref transformation failed",
						err.Error(),
					)
					return
				}

				input.Rule.Source.User = append(input.Rule.Source.User, &cato_models.UserRefInput{
					By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
					Input: ObjectRefOutput.Input,
				})
			}
		}

		// setting source users group
		if !sourceInput.UsersGroup.IsNull() {
			elementsSourceUsersGroupInput := make([]types.Object, 0, len(sourceInput.UsersGroup.Elements()))
			diags = sourceInput.UsersGroup.ElementsAs(ctx, &elementsSourceUsersGroupInput, false)
			resp.Diagnostics.Append(diags...)

			var itemSourceUsersGroupInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_UsersGroup
			for _, item := range elementsSourceUsersGroupInput {
				diags = item.As(ctx, &itemSourceUsersGroupInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceUsersGroupInput)
				if err != nil {
					resp.Diagnostics.AddError(
						"Object Ref transformation failed",
						err.Error(),
					)
					return
				}

				input.Rule.Source.UsersGroup = append(input.Rule.Source.UsersGroup, &cato_models.UsersGroupRefInput{
					By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
					Input: ObjectRefOutput.Input,
				})
			}
		}

		// setting source group
		if !sourceInput.Group.IsNull() {
			elementsSourceGroupInput := make([]types.Object, 0, len(sourceInput.Group.Elements()))
			diags = sourceInput.Group.ElementsAs(ctx, &elementsSourceGroupInput, false)
			resp.Diagnostics.Append(diags...)

			var itemSourceGroupInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_Group
			for _, item := range elementsSourceGroupInput {
				diags = item.As(ctx, &itemSourceGroupInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceGroupInput)
				if err != nil {
					resp.Diagnostics.AddError(
						"Object Ref transformation failed",
						err.Error(),
					)
					return
				}

				input.Rule.Source.Group = append(input.Rule.Source.Group, &cato_models.GroupRefInput{
					By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
					Input: ObjectRefOutput.Input,
				})
			}
		}

		// setting source system group
		if !sourceInput.SystemGroup.IsNull() {
			elementsSourceSystemGroupInput := make([]types.Object, 0, len(sourceInput.SystemGroup.Elements()))
			diags = sourceInput.SystemGroup.ElementsAs(ctx, &elementsSourceSystemGroupInput, false)
			resp.Diagnostics.Append(diags...)

			var itemSourceSystemGroupInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_SystemGroup
			for _, item := range elementsSourceSystemGroupInput {
				diags = item.As(ctx, &itemSourceSystemGroupInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceSystemGroupInput)
				if err != nil {
					resp.Diagnostics.AddError(
						"Object Ref transformation failed",
						err.Error(),
					)
					return
				}

				input.Rule.Source.SystemGroup = append(input.Rule.Source.SystemGroup, &cato_models.SystemGroupRefInput{
					By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
					Input: ObjectRefOutput.Input,
				})
			}
		}
	}

	// setting country
	if !ruleInput.Country.IsNull() {
		elementsCountryInput := make([]types.Object, 0, len(ruleInput.Country.Elements()))
		diags = ruleInput.Country.ElementsAs(ctx, &elementsCountryInput, false)
		resp.Diagnostics.Append(diags...)

		var itemCountryInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Country
		for _, item := range elementsCountryInput {
			diags = item.As(ctx, &itemCountryInput, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)

			ObjectRefOutput, err := utils.TransformObjectRefInput(itemCountryInput)
			if err != nil {
				resp.Diagnostics.AddError(
					"Object Ref transformation failed",
					err.Error(),
				)
				return
			}

			input.Rule.Country = append(input.Rule.Country, &cato_models.CountryRefInput{
				By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
				Input: ObjectRefOutput.Input,
			})
		}
	}

	// setting device
	if !ruleInput.Device.IsNull() {
		elementsDeviceInput := make([]types.Object, 0, len(ruleInput.Device.Elements()))
		diags = ruleInput.Device.ElementsAs(ctx, &elementsDeviceInput, false)
		resp.Diagnostics.Append(diags...)

		var itemDeviceInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Device
		for _, item := range elementsDeviceInput {
			diags = item.As(ctx, &itemDeviceInput, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)

			ObjectRefOutput, err := utils.TransformObjectRefInput(itemDeviceInput)
			if err != nil {
				resp.Diagnostics.AddError(
					"Object Ref transformation failed",
					err.Error(),
				)
				return
			}

			input.Rule.Device = append(input.Rule.Device, &cato_models.DeviceProfileRefInput{
				By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
				Input: ObjectRefOutput.Input,
			})
		}
	}

	// setting device OS
	if !ruleInput.DeviceOs.IsNull() {
		diags = ruleInput.DeviceOs.ElementsAs(ctx, &input.Rule.DeviceOs, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// setting destination
	if !ruleInput.Destination.IsNull() {
		destinationInput := Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination{}
		diags = ruleInput.Destination.As(ctx, &destinationInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)

		// setting destination IP
		if !destinationInput.IP.IsNull() {
			diags = destinationInput.IP.ElementsAs(ctx, &input.Rule.Destination.IP, false)
			resp.Diagnostics.Append(diags...)
		}

		// setting destination subnet
		if !destinationInput.Subnet.IsNull() {
			diags = destinationInput.Subnet.ElementsAs(ctx, &input.Rule.Destination.Subnet, false)
			resp.Diagnostics.Append(diags...)
		}

		// setting destination host
		if !destinationInput.Host.IsNull() {
			elementsDestinationHostInput := make([]types.Object, 0, len(destinationInput.Host.Elements()))
			diags = destinationInput.Host.ElementsAs(ctx, &elementsDestinationHostInput, false)
			resp.Diagnostics.Append(diags...)

			var itemDestinationHostInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_Host
			for _, item := range elementsDestinationHostInput {
				diags = item.As(ctx, &itemDestinationHostInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationHostInput)
				if err != nil {
					resp.Diagnostics.AddError(
						"Object Ref transformation failed",
						err.Error(),
					)
					return
				}

				input.Rule.Destination.Host = append(input.Rule.Destination.Host, &cato_models.HostRefInput{
					By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
					Input: ObjectRefOutput.Input,
				})
			}
		}

		// setting destination site
		if !destinationInput.Site.IsNull() {
			elementsDestinationSiteInput := make([]types.Object, 0, len(destinationInput.Site.Elements()))
			diags = destinationInput.Site.ElementsAs(ctx, &elementsDestinationSiteInput, false)
			resp.Diagnostics.Append(diags...)

			var itemDestinationSiteInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_Site
			for _, item := range elementsDestinationSiteInput {
				diags = item.As(ctx, &itemDestinationSiteInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationSiteInput)
				if err != nil {
					resp.Diagnostics.AddError(
						"Object Ref transformation failed",
						err.Error(),
					)
					return
				}

				input.Rule.Destination.Site = append(input.Rule.Destination.Site, &cato_models.SiteRefInput{
					By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
					Input: ObjectRefOutput.Input,
				})
			}
		}

		// setting destination ip range
		if !destinationInput.IPRange.IsNull() {
			elementsDestinationIPRangeInput := make([]types.Object, 0, len(destinationInput.IPRange.Elements()))
			diags = destinationInput.IPRange.ElementsAs(ctx, &elementsDestinationIPRangeInput, false)
			resp.Diagnostics.Append(diags...)

			var itemDestinationIPRangeInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_IPRange
			for _, item := range elementsDestinationIPRangeInput {
				diags = item.As(ctx, &itemDestinationIPRangeInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				input.Rule.Destination.IPRange = append(input.Rule.Destination.IPRange, &cato_models.IPAddressRangeInput{
					From: itemDestinationIPRangeInput.From.ValueString(),
					To:   itemDestinationIPRangeInput.To.ValueString(),
				})
			}
		}

		// setting destination global ip range
		if !destinationInput.GlobalIPRange.IsNull() {
			elementsDestinationGlobalIPRangeInput := make([]types.Object, 0, len(destinationInput.GlobalIPRange.Elements()))
			diags = destinationInput.GlobalIPRange.ElementsAs(ctx, &elementsDestinationGlobalIPRangeInput, false)
			resp.Diagnostics.Append(diags...)

			var itemDestinationGlobalIPRangeInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_GlobalIPRange
			for _, item := range elementsDestinationGlobalIPRangeInput {
				diags = item.As(ctx, &itemDestinationGlobalIPRangeInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationGlobalIPRangeInput)
				if err != nil {
					resp.Diagnostics.AddError(
						"Object Ref transformation failed for",
						err.Error(),
					)
					return
				}

				input.Rule.Destination.GlobalIPRange = append(input.Rule.Destination.GlobalIPRange, &cato_models.GlobalIPRangeRefInput{
					By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
					Input: ObjectRefOutput.Input,
				})
			}
		}

		// setting destination network interface
		if !destinationInput.NetworkInterface.IsNull() {
			elementsDestinationNetworkInterfaceInput := make([]types.Object, 0, len(destinationInput.NetworkInterface.Elements()))
			diags = destinationInput.NetworkInterface.ElementsAs(ctx, &elementsDestinationNetworkInterfaceInput, false)
			resp.Diagnostics.Append(diags...)

			var itemDestinationNetworkInterfaceInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_NetworkInterface
			for _, item := range elementsDestinationNetworkInterfaceInput {
				diags = item.As(ctx, &itemDestinationNetworkInterfaceInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationNetworkInterfaceInput)
				if err != nil {
					resp.Diagnostics.AddError(
						"Object Ref transformation failed",
						err.Error(),
					)
					return
				}

				input.Rule.Destination.NetworkInterface = append(input.Rule.Destination.NetworkInterface, &cato_models.NetworkInterfaceRefInput{
					By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
					Input: ObjectRefOutput.Input,
				})
			}
		}

		// setting destination site network subnet
		if !destinationInput.SiteNetworkSubnet.IsNull() {
			elementsDestinationSiteNetworkSubnetInput := make([]types.Object, 0, len(destinationInput.SiteNetworkSubnet.Elements()))
			diags = destinationInput.SiteNetworkSubnet.ElementsAs(ctx, &elementsDestinationSiteNetworkSubnetInput, false)
			resp.Diagnostics.Append(diags...)

			var itemDestinationSiteNetworkSubnetInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_SiteNetworkSubnet
			for _, item := range elementsDestinationSiteNetworkSubnetInput {
				diags = item.As(ctx, &itemDestinationSiteNetworkSubnetInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationSiteNetworkSubnetInput)
				if err != nil {
					resp.Diagnostics.AddError(
						"Object Ref transformation failed",
						err.Error(),
					)
					return
				}

				input.Rule.Destination.SiteNetworkSubnet = append(input.Rule.Destination.SiteNetworkSubnet, &cato_models.SiteNetworkSubnetRefInput{
					By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
					Input: ObjectRefOutput.Input,
				})
			}
		}

		// setting destination floating subnet
		if !destinationInput.FloatingSubnet.IsNull() {
			elementsDestinationFloatingSubnetInput := make([]types.Object, 0, len(destinationInput.FloatingSubnet.Elements()))
			diags = destinationInput.FloatingSubnet.ElementsAs(ctx, &elementsDestinationFloatingSubnetInput, false)
			resp.Diagnostics.Append(diags...)

			var itemDestinationFloatingSubnetInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_FloatingSubnet
			for _, item := range elementsDestinationFloatingSubnetInput {
				diags = item.As(ctx, &itemDestinationFloatingSubnetInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationFloatingSubnetInput)
				if err != nil {
					resp.Diagnostics.AddError(
						"Object Ref transformation failed",
						err.Error(),
					)
					return
				}

				input.Rule.Destination.FloatingSubnet = append(input.Rule.Destination.FloatingSubnet, &cato_models.FloatingSubnetRefInput{
					By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
					Input: ObjectRefOutput.Input,
				})
			}
		}

		// setting destination user
		if !destinationInput.User.IsNull() {
			elementsDestinationUserInput := make([]types.Object, 0, len(destinationInput.User.Elements()))
			diags = destinationInput.User.ElementsAs(ctx, &elementsDestinationUserInput, false)
			resp.Diagnostics.Append(diags...)

			var itemDestinationUserInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_User
			for _, item := range elementsDestinationUserInput {
				diags = item.As(ctx, &itemDestinationUserInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationUserInput)
				if err != nil {
					resp.Diagnostics.AddError(
						"Object Ref transformation failed",
						err.Error(),
					)
					return
				}

				input.Rule.Destination.User = append(input.Rule.Destination.User, &cato_models.UserRefInput{
					By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
					Input: ObjectRefOutput.Input,
				})
			}
		}

		// setting destination users group
		if !destinationInput.UsersGroup.IsNull() {
			elementsDestinationUsersGroupInput := make([]types.Object, 0, len(destinationInput.UsersGroup.Elements()))
			diags = destinationInput.UsersGroup.ElementsAs(ctx, &elementsDestinationUsersGroupInput, false)
			resp.Diagnostics.Append(diags...)

			var itemDestinationUsersGroupInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_UsersGroup
			for _, item := range elementsDestinationUsersGroupInput {
				diags = item.As(ctx, &itemDestinationUsersGroupInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationUsersGroupInput)
				if err != nil {
					resp.Diagnostics.AddError(
						"Object Ref transformation failed",
						err.Error(),
					)
					return
				}

				input.Rule.Destination.UsersGroup = append(input.Rule.Destination.UsersGroup, &cato_models.UsersGroupRefInput{
					By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
					Input: ObjectRefOutput.Input,
				})
			}
		}

		// setting destination group
		if !destinationInput.Group.IsNull() {
			elementsDestinationGroupInput := make([]types.Object, 0, len(destinationInput.Group.Elements()))
			diags = destinationInput.Group.ElementsAs(ctx, &elementsDestinationGroupInput, false)
			resp.Diagnostics.Append(diags...)

			var itemDestinationGroupInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_Group
			for _, item := range elementsDestinationGroupInput {
				diags = item.As(ctx, &itemDestinationGroupInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationGroupInput)
				if err != nil {
					resp.Diagnostics.AddError(
						"Object Ref transformation failed",
						err.Error(),
					)
					return
				}

				input.Rule.Destination.Group = append(input.Rule.Destination.Group, &cato_models.GroupRefInput{
					By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
					Input: ObjectRefOutput.Input,
				})
			}
		}

		// setting destination system group
		if !destinationInput.SystemGroup.IsNull() {
			elementsDestinationSystemGroupInput := make([]types.Object, 0, len(destinationInput.SystemGroup.Elements()))
			diags = destinationInput.SystemGroup.ElementsAs(ctx, &elementsDestinationSystemGroupInput, false)
			resp.Diagnostics.Append(diags...)

			var itemDestinationSystemGroupInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_SystemGroup
			for _, item := range elementsDestinationSystemGroupInput {
				diags = item.As(ctx, &itemDestinationSystemGroupInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationSystemGroupInput)
				if err != nil {
					resp.Diagnostics.AddError(
						"Object Ref transformation failed",
						err.Error(),
					)
					return
				}

				input.Rule.Destination.SystemGroup = append(input.Rule.Destination.SystemGroup, &cato_models.SystemGroupRefInput{
					By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
					Input: ObjectRefOutput.Input,
				})
			}
		}
	}

	// setting application
	if !ruleInput.Application.IsNull() {
		applicationInput := Policy_Policy_WanFirewall_Policy_Rules_Rule_Application{}
		diags = ruleInput.Application.As(ctx, &applicationInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)

		// setting application IP
		if !applicationInput.IP.IsNull() {
			diags = applicationInput.IP.ElementsAs(ctx, &input.Rule.Application.IP, false)
			resp.Diagnostics.Append(diags...)
		}

		// setting application subnet
		if !applicationInput.Subnet.IsNull() {
			diags = applicationInput.Subnet.ElementsAs(ctx, &input.Rule.Application.Subnet, false)
			resp.Diagnostics.Append(diags...)
		}

		// setting application domain
		if !applicationInput.Domain.IsNull() {
			diags = applicationInput.Domain.ElementsAs(ctx, &input.Rule.Application.Domain, false)
			resp.Diagnostics.Append(diags...)
		}

		// setting application fqdn
		if !applicationInput.Fqdn.IsNull() {
			diags = applicationInput.Fqdn.ElementsAs(ctx, &input.Rule.Application.Fqdn, false)
			resp.Diagnostics.Append(diags...)
		}

		// setting application application
		if !applicationInput.Application.IsNull() {
			elementsApplicationApplicationInput := make([]types.Object, 0, len(applicationInput.Application.Elements()))
			diags = applicationInput.Application.ElementsAs(ctx, &elementsApplicationApplicationInput, false)
			resp.Diagnostics.Append(diags...)

			var itemApplicationApplicationInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_Application
			for _, item := range elementsApplicationApplicationInput {
				diags = item.As(ctx, &itemApplicationApplicationInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationApplicationInput)
				if err != nil {
					resp.Diagnostics.AddError(
						"Object Ref transformation failed",
						err.Error(),
					)
					return
				}

				input.Rule.Application.Application = append(input.Rule.Application.Application, &cato_models.ApplicationRefInput{
					By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
					Input: ObjectRefOutput.Input,
				})
			}
		}

		// setting application custom app
		if !applicationInput.CustomApp.IsNull() {
			elementsApplicationCustomAppInput := make([]types.Object, 0, len(applicationInput.CustomApp.Elements()))
			diags = applicationInput.CustomApp.ElementsAs(ctx, &elementsApplicationCustomAppInput, false)
			resp.Diagnostics.Append(diags...)

			var itemApplicationCustomAppInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_CustomApp
			for _, item := range elementsApplicationCustomAppInput {
				diags = item.As(ctx, &itemApplicationCustomAppInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationCustomAppInput)
				if err != nil {
					resp.Diagnostics.AddError(
						"Object Ref transformation failed",
						err.Error(),
					)
					return
				}

				input.Rule.Application.CustomApp = append(input.Rule.Application.CustomApp, &cato_models.CustomApplicationRefInput{
					By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
					Input: ObjectRefOutput.Input,
				})
			}
		}

		// setting application ip range
		if !applicationInput.IPRange.IsNull() {
			elementsApplicationIPRangeInput := make([]types.Object, 0, len(applicationInput.IPRange.Elements()))
			diags = applicationInput.IPRange.ElementsAs(ctx, &elementsApplicationIPRangeInput, false)
			resp.Diagnostics.Append(diags...)

			var itemApplicationIPRangeInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_IPRange
			for _, item := range elementsApplicationIPRangeInput {
				diags = item.As(ctx, &itemApplicationIPRangeInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				input.Rule.Application.IPRange = append(input.Rule.Application.IPRange, &cato_models.IPAddressRangeInput{
					From: itemApplicationIPRangeInput.From.ValueString(),
					To:   itemApplicationIPRangeInput.To.ValueString(),
				})
			}
		}

		// setting application global ip range
		if !applicationInput.GlobalIPRange.IsNull() {
			elementsApplicationGlobalIPRangeInput := make([]types.Object, 0, len(applicationInput.GlobalIPRange.Elements()))
			diags = applicationInput.GlobalIPRange.ElementsAs(ctx, &elementsApplicationGlobalIPRangeInput, false)
			resp.Diagnostics.Append(diags...)

			var itemApplicationGlobalIPRangeInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_GlobalIPRange
			for _, item := range elementsApplicationGlobalIPRangeInput {
				diags = item.As(ctx, &itemApplicationGlobalIPRangeInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationGlobalIPRangeInput)
				if err != nil {
					resp.Diagnostics.AddError(
						"Object Ref transformation failed",
						err.Error(),
					)
					return
				}

				input.Rule.Application.GlobalIPRange = append(input.Rule.Application.GlobalIPRange, &cato_models.GlobalIPRangeRefInput{
					By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
					Input: ObjectRefOutput.Input,
				})
			}
		}

		// setting application app category
		if !applicationInput.AppCategory.IsNull() {
			elementsApplicationAppCategoryInput := make([]types.Object, 0, len(applicationInput.AppCategory.Elements()))
			diags = applicationInput.AppCategory.ElementsAs(ctx, &elementsApplicationAppCategoryInput, false)
			resp.Diagnostics.Append(diags...)

			var itemApplicationAppCategoryInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_AppCategory
			for _, item := range elementsApplicationAppCategoryInput {
				diags = item.As(ctx, &itemApplicationAppCategoryInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationAppCategoryInput)
				if err != nil {
					resp.Diagnostics.AddError(
						"Object Ref transformation failed",
						err.Error(),
					)
					return
				}

				input.Rule.Application.AppCategory = append(input.Rule.Application.AppCategory, &cato_models.ApplicationCategoryRefInput{
					By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
					Input: ObjectRefOutput.Input,
				})
			}
		}

		// setting application custom app category
		if !applicationInput.CustomCategory.IsNull() {
			elementsApplicationCustomCategoryInput := make([]types.Object, 0, len(applicationInput.CustomCategory.Elements()))
			diags = applicationInput.CustomCategory.ElementsAs(ctx, &elementsApplicationCustomCategoryInput, false)
			resp.Diagnostics.Append(diags...)

			var itemApplicationCustomCategoryInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_CustomCategory
			for _, item := range elementsApplicationCustomCategoryInput {
				diags = item.As(ctx, &itemApplicationCustomCategoryInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationCustomCategoryInput)
				if err != nil {
					resp.Diagnostics.AddError(
						"Object Ref transformation failed",
						err.Error(),
					)
					return
				}

				input.Rule.Application.CustomCategory = append(input.Rule.Application.CustomCategory, &cato_models.CustomCategoryRefInput{
					By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
					Input: ObjectRefOutput.Input,
				})
			}
		}

		// setting application sanctionned apps category
		if !applicationInput.SanctionedAppsCategory.IsNull() {
			elementsApplicationSanctionedAppsCategoryInput := make([]types.Object, 0, len(applicationInput.SanctionedAppsCategory.Elements()))
			diags = applicationInput.SanctionedAppsCategory.ElementsAs(ctx, &elementsApplicationSanctionedAppsCategoryInput, false)
			resp.Diagnostics.Append(diags...)

			var itemApplicationSanctionedAppsCategoryInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_SanctionedAppsCategory
			for _, item := range elementsApplicationSanctionedAppsCategoryInput {
				diags = item.As(ctx, &itemApplicationSanctionedAppsCategoryInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationSanctionedAppsCategoryInput)
				if err != nil {
					resp.Diagnostics.AddError(
						"Object Ref transformation failed",
						err.Error(),
					)
					return
				}

				input.Rule.Application.SanctionedAppsCategory = append(input.Rule.Application.SanctionedAppsCategory, &cato_models.SanctionedAppsCategoryRefInput{
					By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
					Input: ObjectRefOutput.Input,
				})
			}
		}
	}

	// setting service
	if !ruleInput.Service.IsNull() {

		serviceInput := Policy_Policy_WanFirewall_Policy_Rules_Rule_Service{}

		diags = ruleInput.Service.As(ctx, &serviceInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// setting service standard
		if !serviceInput.Standard.IsNull() {
			elementsServiceStandardInput := make([]types.Object, 0, len(serviceInput.Standard.Elements()))
			diags = serviceInput.Standard.ElementsAs(ctx, &elementsServiceStandardInput, false)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			var itemServiceStandardInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Service_Standard
			for _, item := range elementsServiceStandardInput {
				diags = item.As(ctx, &itemServiceStandardInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				ObjectRefOutput, err := utils.TransformObjectRefInput(itemServiceStandardInput)
				if err != nil {
					resp.Diagnostics.AddError(
						"Object Ref transformation failed",
						err.Error(),
					)
					return
				}

				input.Rule.Service.Standard = append(input.Rule.Service.Standard, &cato_models.ServiceRefInput{
					By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
					Input: ObjectRefOutput.Input,
				})
			}
		}

		// setting service custom
		if !serviceInput.Custom.IsNull() {
			elementsServiceCustomInput := make([]types.Object, 0, len(serviceInput.Custom.Elements()))
			diags = serviceInput.Custom.ElementsAs(ctx, &elementsServiceCustomInput, false)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			var itemServiceCustomInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Service_Custom
			for _, item := range elementsServiceCustomInput {
				diags = item.As(ctx, &itemServiceCustomInput, basetypes.ObjectAsOptions{})

				customInput := &cato_models.CustomServiceInput{
					Protocol: cato_models.IPProtocol(itemServiceCustomInput.Protocol.ValueString()),
				}

				// setting service custom port
				if !itemServiceCustomInput.Port.IsNull() {
					elementsPort := make([]types.String, 0, len(itemServiceCustomInput.Port.Elements()))
					diags = itemServiceCustomInput.Port.ElementsAs(ctx, &elementsPort, false)
					resp.Diagnostics.Append(diags...)

					inputPort := []cato_scalars.Port{}
					for _, item := range elementsPort {
						inputPort = append(inputPort, cato_scalars.Port(item.ValueString()))
					}

					customInput.Port = inputPort
				}

				// setting service custom port range
				if !itemServiceCustomInput.PortRange.IsNull() {
					var itemPortRange Policy_Policy_WanFirewall_Policy_Rules_Rule_Service_Custom_PortRange
					diags = itemServiceCustomInput.PortRange.As(ctx, &itemPortRange, basetypes.ObjectAsOptions{})

					inputPortRange := cato_models.PortRangeInput{
						From: cato_scalars.Port(itemPortRange.From.ValueString()),
						To:   cato_scalars.Port(itemPortRange.To.ValueString()),
					}

					customInput.PortRange = &inputPortRange
				}

				// append custom service
				input.Rule.Service.Custom = append(input.Rule.Service.Custom, customInput)
			}
		}
	}

	// setting tracking
	if !ruleInput.Tracking.IsNull() {

		trackingInput := Policy_Policy_WanFirewall_Policy_Rules_Rule_Tracking{}
		diags = ruleInput.Tracking.As(ctx, &trackingInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// setting tracking event
		trackingEventInput := Policy_Policy_WanFirewall_Policy_Rules_Rule_Tracking_Event{}
		diags = trackingInput.Event.As(ctx, &trackingEventInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		input.Rule.Tracking.Event.Enabled = trackingEventInput.Enabled.ValueBoolPointer()

		if !trackingInput.Alert.IsNull() {

			trackingAlertInput := Policy_Policy_WanFirewall_Policy_Rules_Rule_Tracking_Alert{}
			diags = trackingInput.Alert.As(ctx, &trackingAlertInput, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			input.Rule.Tracking.Alert.Enabled = trackingAlertInput.Enabled.ValueBoolPointer()
			input.Rule.Tracking.Alert.Frequency = (*cato_models.PolicyRuleTrackingFrequencyEnum)(trackingAlertInput.Frequency.ValueStringPointer())

			// setting tracking alert subscription group
			if !trackingAlertInput.SubscriptionGroup.IsNull() {
				elementsAlertSubscriptionGroupInput := make([]types.Object, 0, len(trackingAlertInput.SubscriptionGroup.Elements()))
				diags = trackingAlertInput.SubscriptionGroup.ElementsAs(ctx, &elementsAlertSubscriptionGroupInput, false)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}

				var itemAlertSubscriptionGroupInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Tracking_Alert_SubscriptionGroup
				for _, item := range elementsAlertSubscriptionGroupInput {
					diags = item.As(ctx, &itemAlertSubscriptionGroupInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemAlertSubscriptionGroupInput)
					if err != nil {
						resp.Diagnostics.AddError(
							"Object Ref transformation failed",
							err.Error(),
						)
						return
					}

					input.Rule.Tracking.Alert.SubscriptionGroup = append(input.Rule.Tracking.Alert.SubscriptionGroup, &cato_models.SubscriptionGroupRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
			}

			// setting tracking alert webhook
			if !trackingAlertInput.Webhook.IsNull() {
				if !trackingAlertInput.Webhook.IsNull() {
					elementsAlertWebHookInput := make([]types.Object, 0, len(trackingAlertInput.Webhook.Elements()))
					diags = trackingAlertInput.Webhook.ElementsAs(ctx, &elementsAlertWebHookInput, false)
					resp.Diagnostics.Append(diags...)
					if resp.Diagnostics.HasError() {
						return
					}

					var itemAlertWebHookInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Tracking_Alert_SubscriptionGroup
					for _, item := range elementsAlertWebHookInput {
						diags = item.As(ctx, &itemAlertWebHookInput, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemAlertWebHookInput)
						if err != nil {
							resp.Diagnostics.AddError(
								"Object Ref transformation failed",
								err.Error(),
							)
							return
						}

						input.Rule.Tracking.Alert.Webhook = append(input.Rule.Tracking.Alert.Webhook, &cato_models.SubscriptionWebhookRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
				}
			}

			// setting tracking alert mailing list
			if !trackingAlertInput.MailingList.IsNull() {
				elementsAlertMailingListInput := make([]types.Object, 0, len(trackingAlertInput.MailingList.Elements()))
				diags = trackingAlertInput.MailingList.ElementsAs(ctx, &elementsAlertMailingListInput, false)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}

				var itemAlertMailingListInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Tracking_Alert_SubscriptionGroup
				for _, item := range elementsAlertMailingListInput {
					diags = item.As(ctx, &itemAlertMailingListInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemAlertMailingListInput)
					if err != nil {
						resp.Diagnostics.AddError(
							"Object Ref transformation failed",
							err.Error(),
						)
						return
					}

					input.Rule.Tracking.Alert.MailingList = append(input.Rule.Tracking.Alert.MailingList, &cato_models.SubscriptionMailingListRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
			}
		}
	} else {
		// set default value if tracking null
		defaultEnabled := false
		input.Rule.Tracking.Event.Enabled = &defaultEnabled
		input.Rule.Tracking.Alert.Enabled = &defaultEnabled
	}

	// setting schedule
	if !ruleInput.Schedule.IsNull() {

		scheduleInput := Policy_Policy_WanFirewall_Policy_Rules_Rule_Schedule{}
		diags = ruleInput.Schedule.As(ctx, &scheduleInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		input.Rule.Schedule.ActiveOn = (*cato_models.PolicyActiveOnEnum)(scheduleInput.ActiveOn.ValueStringPointer())

		// setting schedule custome time frame
		if !scheduleInput.CustomTimeframe.IsNull() {
			input.Rule.Schedule.CustomTimeframe = &cato_models.PolicyCustomTimeframeUpdateInput{}

			customeTimeFrameInput := Policy_Policy_WanFirewall_Policy_Rules_Rule_Schedule_CustomTimeframe{}
			diags = scheduleInput.CustomTimeframe.As(ctx, &customeTimeFrameInput, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			input.Rule.Schedule.CustomTimeframe.From = customeTimeFrameInput.From.ValueStringPointer()
			input.Rule.Schedule.CustomTimeframe.To = customeTimeFrameInput.To.ValueStringPointer()

		}

		if !scheduleInput.CustomRecurring.IsNull() {
			input.Rule.Schedule.CustomRecurring = &cato_models.PolicyCustomRecurringUpdateInput{}

			customRecurringInput := Policy_Policy_WanFirewall_Policy_Rules_Rule_Schedule_CustomRecurring{}
			diags = scheduleInput.CustomRecurring.As(ctx, &customRecurringInput, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			input.Rule.Schedule.CustomRecurring.From = (*cato_scalars.Time)(customRecurringInput.From.ValueStringPointer())
			input.Rule.Schedule.CustomRecurring.To = (*cato_scalars.Time)(customRecurringInput.To.ValueStringPointer())

			// setting schedule custom recurring days
			diags = customRecurringInput.Days.ElementsAs(ctx, &input.Rule.Schedule.CustomRecurring.Days, false)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
		}
	} else {
		// set default value if tracking null
		defaultActiveOn := "ALWAYS"
		input.Rule.Schedule.ActiveOn = (*cato_models.PolicyActiveOnEnum)(&defaultActiveOn)
	}

	// settings exceptions
	if !ruleInput.Exceptions.IsNull() {
		elementsExceptionsInput := make([]types.Object, 0, len(ruleInput.Exceptions.Elements()))
		diags = ruleInput.Exceptions.ElementsAs(ctx, &elementsExceptionsInput, false)
		resp.Diagnostics.Append(diags...)

		// loop over exceptions
		var itemExceptionsInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Exceptions
		for _, item := range elementsExceptionsInput {

			exceptionInput := cato_models.WanFirewallRuleExceptionInput{}

			diags = item.As(ctx, &itemExceptionsInput, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)

			// setting exception name
			exceptionInput.Name = itemExceptionsInput.Name.ValueString()

			// setting exception direction
			exceptionInput.Direction = cato_models.WanFirewallDirectionEnum(itemExceptionsInput.Direction.ValueString())

			// setting exception connection origin
			if !itemExceptionsInput.ConnectionOrigin.IsNull() {
				exceptionInput.ConnectionOrigin = cato_models.ConnectionOriginEnum(itemExceptionsInput.ConnectionOrigin.ValueString())
			} else {
				exceptionInput.ConnectionOrigin = cato_models.ConnectionOriginEnum("ANY")
			}

			// setting source
			if !itemExceptionsInput.Source.IsNull() {

				exceptionInput.Source = &cato_models.WanFirewallSourceInput{}
				sourceInput := Policy_Policy_WanFirewall_Policy_Rules_Rule_Source{}
				diags = itemExceptionsInput.Source.As(ctx, &sourceInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				// setting source IP
				if !sourceInput.IP.IsNull() {
					diags = sourceInput.IP.ElementsAs(ctx, &exceptionInput.Source.IP, false)
					resp.Diagnostics.Append(diags...)
				}

				// setting source subnet
				if !sourceInput.Subnet.IsNull() {
					diags = sourceInput.Subnet.ElementsAs(ctx, &exceptionInput.Source.Subnet, false)
					resp.Diagnostics.Append(diags...)
				}

				// setting source host
				if !sourceInput.Host.IsNull() {
					elementsSourceHostInput := make([]types.Object, 0, len(sourceInput.Host.Elements()))
					diags = sourceInput.Host.ElementsAs(ctx, &elementsSourceHostInput, false)
					resp.Diagnostics.Append(diags...)

					var itemSourceHostInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_Host
					for _, item := range elementsSourceHostInput {
						diags = item.As(ctx, &itemSourceHostInput, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceHostInput)
						if err != nil {
							resp.Diagnostics.AddError(
								"Object Ref transformation failed",
								err.Error(),
							)
							return
						}

						exceptionInput.Source.Host = append(exceptionInput.Source.Host, &cato_models.HostRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
				}

				// setting source site
				if !sourceInput.Site.IsNull() {
					elementsSourceSiteInput := make([]types.Object, 0, len(sourceInput.Site.Elements()))
					diags = sourceInput.Site.ElementsAs(ctx, &elementsSourceSiteInput, false)
					resp.Diagnostics.Append(diags...)

					var itemSourceSiteInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_Site
					for _, item := range elementsSourceSiteInput {
						diags = item.As(ctx, &itemSourceSiteInput, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceSiteInput)
						if err != nil {
							resp.Diagnostics.AddError(
								"Object Ref transformation failed",
								err.Error(),
							)
							return
						}

						exceptionInput.Source.Site = append(exceptionInput.Source.Site, &cato_models.SiteRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
				}

				// setting source ip range
				if !sourceInput.IPRange.IsNull() {
					elementsSourceIPRangeInput := make([]types.Object, 0, len(sourceInput.IPRange.Elements()))
					diags = sourceInput.IPRange.ElementsAs(ctx, &elementsSourceIPRangeInput, false)
					resp.Diagnostics.Append(diags...)

					var itemSourceIPRangeInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_IPRange
					for _, item := range elementsSourceIPRangeInput {
						diags = item.As(ctx, &itemSourceIPRangeInput, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						exceptionInput.Source.IPRange = append(exceptionInput.Source.IPRange, &cato_models.IPAddressRangeInput{
							From: itemSourceIPRangeInput.From.ValueString(),
							To:   itemSourceIPRangeInput.To.ValueString(),
						})
					}
				}

				// setting source global ip range
				if !sourceInput.GlobalIPRange.IsNull() {
					elementsSourceGlobalIPRangeInput := make([]types.Object, 0, len(sourceInput.GlobalIPRange.Elements()))
					diags = sourceInput.GlobalIPRange.ElementsAs(ctx, &elementsSourceGlobalIPRangeInput, false)
					resp.Diagnostics.Append(diags...)

					var itemSourceGlobalIPRangeInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_GlobalIPRange
					for _, item := range elementsSourceGlobalIPRangeInput {
						diags = item.As(ctx, &itemSourceGlobalIPRangeInput, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceGlobalIPRangeInput)
						if err != nil {
							resp.Diagnostics.AddError(
								"Object Ref transformation failed for",
								err.Error(),
							)
							return
						}

						exceptionInput.Source.GlobalIPRange = append(exceptionInput.Source.GlobalIPRange, &cato_models.GlobalIPRangeRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
				}

				// setting source network interface
				if !sourceInput.NetworkInterface.IsNull() {
					elementsSourceNetworkInterfaceInput := make([]types.Object, 0, len(sourceInput.NetworkInterface.Elements()))
					diags = sourceInput.NetworkInterface.ElementsAs(ctx, &elementsSourceNetworkInterfaceInput, false)
					resp.Diagnostics.Append(diags...)

					var itemSourceNetworkInterfaceInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_NetworkInterface
					for _, item := range elementsSourceNetworkInterfaceInput {
						diags = item.As(ctx, &itemSourceNetworkInterfaceInput, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceNetworkInterfaceInput)
						if err != nil {
							resp.Diagnostics.AddError(
								"Object Ref transformation failed",
								err.Error(),
							)
							return
						}

						exceptionInput.Source.NetworkInterface = append(exceptionInput.Source.NetworkInterface, &cato_models.NetworkInterfaceRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
				}

				// setting source site network subnet
				if !sourceInput.SiteNetworkSubnet.IsNull() {
					elementsSourceSiteNetworkSubnetInput := make([]types.Object, 0, len(sourceInput.SiteNetworkSubnet.Elements()))
					diags = sourceInput.SiteNetworkSubnet.ElementsAs(ctx, &elementsSourceSiteNetworkSubnetInput, false)
					resp.Diagnostics.Append(diags...)

					var itemSourceSiteNetworkSubnetInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_SiteNetworkSubnet
					for _, item := range elementsSourceSiteNetworkSubnetInput {
						diags = item.As(ctx, &itemSourceSiteNetworkSubnetInput, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceSiteNetworkSubnetInput)
						if err != nil {
							resp.Diagnostics.AddError(
								"Object Ref transformation failed",
								err.Error(),
							)
							return
						}

						exceptionInput.Source.SiteNetworkSubnet = append(exceptionInput.Source.SiteNetworkSubnet, &cato_models.SiteNetworkSubnetRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
				}

				// setting source floating subnet
				if !sourceInput.FloatingSubnet.IsNull() {
					elementsSourceFloatingSubnetInput := make([]types.Object, 0, len(sourceInput.FloatingSubnet.Elements()))
					diags = sourceInput.FloatingSubnet.ElementsAs(ctx, &elementsSourceFloatingSubnetInput, false)
					resp.Diagnostics.Append(diags...)

					var itemSourceFloatingSubnetInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_FloatingSubnet
					for _, item := range elementsSourceFloatingSubnetInput {
						diags = item.As(ctx, &itemSourceFloatingSubnetInput, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceFloatingSubnetInput)
						if err != nil {
							resp.Diagnostics.AddError(
								"Object Ref transformation failed",
								err.Error(),
							)
							return
						}

						exceptionInput.Source.FloatingSubnet = append(exceptionInput.Source.FloatingSubnet, &cato_models.FloatingSubnetRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
				}

				// setting source user
				if !sourceInput.User.IsNull() {
					elementsSourceUserInput := make([]types.Object, 0, len(sourceInput.User.Elements()))
					diags = sourceInput.User.ElementsAs(ctx, &elementsSourceUserInput, false)
					resp.Diagnostics.Append(diags...)

					var itemSourceUserInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_User
					for _, item := range elementsSourceUserInput {
						diags = item.As(ctx, &itemSourceUserInput, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceUserInput)
						if err != nil {
							resp.Diagnostics.AddError(
								"Object Ref transformation failed",
								err.Error(),
							)
							return
						}

						exceptionInput.Source.User = append(exceptionInput.Source.User, &cato_models.UserRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
				}

				// setting source users group
				if !sourceInput.UsersGroup.IsNull() {
					elementsSourceUsersGroupInput := make([]types.Object, 0, len(sourceInput.UsersGroup.Elements()))
					diags = sourceInput.UsersGroup.ElementsAs(ctx, &elementsSourceUsersGroupInput, false)
					resp.Diagnostics.Append(diags...)

					var itemSourceUsersGroupInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_UsersGroup
					for _, item := range elementsSourceUsersGroupInput {
						diags = item.As(ctx, &itemSourceUsersGroupInput, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceUsersGroupInput)
						if err != nil {
							resp.Diagnostics.AddError(
								"Object Ref transformation failed",
								err.Error(),
							)
							return
						}

						exceptionInput.Source.UsersGroup = append(exceptionInput.Source.UsersGroup, &cato_models.UsersGroupRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
				}

				// setting source group
				if !sourceInput.Group.IsNull() {
					elementsSourceGroupInput := make([]types.Object, 0, len(sourceInput.Group.Elements()))
					diags = sourceInput.Group.ElementsAs(ctx, &elementsSourceGroupInput, false)
					resp.Diagnostics.Append(diags...)

					var itemSourceGroupInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_Group
					for _, item := range elementsSourceGroupInput {
						diags = item.As(ctx, &itemSourceGroupInput, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceGroupInput)
						if err != nil {
							resp.Diagnostics.AddError(
								"Object Ref transformation failed",
								err.Error(),
							)
							return
						}

						exceptionInput.Source.Group = append(exceptionInput.Source.Group, &cato_models.GroupRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
				}

				// setting source system group
				if !sourceInput.SystemGroup.IsNull() {
					elementsSourceSystemGroupInput := make([]types.Object, 0, len(sourceInput.SystemGroup.Elements()))
					diags = sourceInput.SystemGroup.ElementsAs(ctx, &elementsSourceSystemGroupInput, false)
					resp.Diagnostics.Append(diags...)

					var itemSourceSystemGroupInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Source_SystemGroup
					for _, item := range elementsSourceSystemGroupInput {
						diags = item.As(ctx, &itemSourceSystemGroupInput, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemSourceSystemGroupInput)
						if err != nil {
							resp.Diagnostics.AddError(
								"Object Ref transformation failed",
								err.Error(),
							)
							return
						}

						exceptionInput.Source.SystemGroup = append(exceptionInput.Source.SystemGroup, &cato_models.SystemGroupRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
				}
			}

			// setting country
			if !itemExceptionsInput.Country.IsNull() {

				exceptionInput.Country = []*cato_models.CountryRefInput{}
				elementsCountryInput := make([]types.Object, 0, len(itemExceptionsInput.Country.Elements()))
				diags = itemExceptionsInput.Country.ElementsAs(ctx, &elementsCountryInput, false)
				resp.Diagnostics.Append(diags...)

				var itemCountryInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Country
				for _, item := range elementsCountryInput {
					diags = item.As(ctx, &itemCountryInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemCountryInput)
					if err != nil {
						resp.Diagnostics.AddError(
							"Object Ref transformation failed",
							err.Error(),
						)
						return
					}

					exceptionInput.Country = append(exceptionInput.Country, &cato_models.CountryRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
			}

			// setting device
			if !itemExceptionsInput.Device.IsNull() {

				exceptionInput.Device = []*cato_models.DeviceProfileRefInput{}
				elementsDeviceInput := make([]types.Object, 0, len(itemExceptionsInput.Device.Elements()))
				diags = itemExceptionsInput.Device.ElementsAs(ctx, &elementsDeviceInput, false)
				resp.Diagnostics.Append(diags...)

				var itemDeviceInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Device
				for _, item := range elementsDeviceInput {
					diags = item.As(ctx, &itemDeviceInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemDeviceInput)
					if err != nil {
						resp.Diagnostics.AddError(
							"Object Ref transformation failed",
							err.Error(),
						)
						return
					}

					exceptionInput.Device = append(exceptionInput.Device, &cato_models.DeviceProfileRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
			}

			// setting device OS
			if !itemExceptionsInput.DeviceOs.IsNull() {
				diags = itemExceptionsInput.DeviceOs.ElementsAs(ctx, &exceptionInput.DeviceOs, false)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
			}

			// setting destination
			if !itemExceptionsInput.Destination.IsNull() {

				exceptionInput.Destination = &cato_models.WanFirewallDestinationInput{}
				destinationInput := Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination{}
				diags = itemExceptionsInput.Destination.As(ctx, &destinationInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				// setting destination IP
				if !destinationInput.IP.IsNull() {
					diags = destinationInput.IP.ElementsAs(ctx, &exceptionInput.Destination.IP, false)
					resp.Diagnostics.Append(diags...)
				}

				// setting destination subnet
				if !destinationInput.Subnet.IsNull() {
					diags = destinationInput.Subnet.ElementsAs(ctx, &exceptionInput.Destination.Subnet, false)
					resp.Diagnostics.Append(diags...)
				}

				// setting destination host
				if !destinationInput.Host.IsNull() {
					elementsDestinationHostInput := make([]types.Object, 0, len(destinationInput.Host.Elements()))
					diags = destinationInput.Host.ElementsAs(ctx, &elementsDestinationHostInput, false)
					resp.Diagnostics.Append(diags...)

					var itemDestinationHostInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_Host
					for _, item := range elementsDestinationHostInput {
						diags = item.As(ctx, &itemDestinationHostInput, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationHostInput)
						if err != nil {
							resp.Diagnostics.AddError(
								"Object Ref transformation failed",
								err.Error(),
							)
							return
						}

						exceptionInput.Destination.Host = append(exceptionInput.Destination.Host, &cato_models.HostRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
				}

				// setting destination site
				if !destinationInput.Site.IsNull() {
					elementsDestinationSiteInput := make([]types.Object, 0, len(destinationInput.Site.Elements()))
					diags = destinationInput.Site.ElementsAs(ctx, &elementsDestinationSiteInput, false)
					resp.Diagnostics.Append(diags...)

					var itemDestinationSiteInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_Site
					for _, item := range elementsDestinationSiteInput {
						diags = item.As(ctx, &itemDestinationSiteInput, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationSiteInput)
						if err != nil {
							resp.Diagnostics.AddError(
								"Object Ref transformation failed",
								err.Error(),
							)
							return
						}

						exceptionInput.Destination.Site = append(exceptionInput.Destination.Site, &cato_models.SiteRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
				}

				// setting destination ip range
				if !destinationInput.IPRange.IsNull() {
					elementsDestinationIPRangeInput := make([]types.Object, 0, len(destinationInput.IPRange.Elements()))
					diags = destinationInput.IPRange.ElementsAs(ctx, &elementsDestinationIPRangeInput, false)
					resp.Diagnostics.Append(diags...)

					var itemDestinationIPRangeInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_IPRange
					for _, item := range elementsDestinationIPRangeInput {
						diags = item.As(ctx, &itemDestinationIPRangeInput, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						exceptionInput.Destination.IPRange = append(exceptionInput.Destination.IPRange, &cato_models.IPAddressRangeInput{
							From: itemDestinationIPRangeInput.From.ValueString(),
							To:   itemDestinationIPRangeInput.To.ValueString(),
						})
					}
				}

				// setting destination global ip range
				if !destinationInput.GlobalIPRange.IsNull() {
					elementsDestinationGlobalIPRangeInput := make([]types.Object, 0, len(destinationInput.GlobalIPRange.Elements()))
					diags = destinationInput.GlobalIPRange.ElementsAs(ctx, &elementsDestinationGlobalIPRangeInput, false)
					resp.Diagnostics.Append(diags...)

					var itemDestinationGlobalIPRangeInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_GlobalIPRange
					for _, item := range elementsDestinationGlobalIPRangeInput {
						diags = item.As(ctx, &itemDestinationGlobalIPRangeInput, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationGlobalIPRangeInput)
						if err != nil {
							resp.Diagnostics.AddError(
								"Object Ref transformation failed for",
								err.Error(),
							)
							return
						}

						exceptionInput.Destination.GlobalIPRange = append(exceptionInput.Destination.GlobalIPRange, &cato_models.GlobalIPRangeRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
				}

				// setting destination network interface
				if !destinationInput.NetworkInterface.IsNull() {
					elementsDestinationNetworkInterfaceInput := make([]types.Object, 0, len(destinationInput.NetworkInterface.Elements()))
					diags = destinationInput.NetworkInterface.ElementsAs(ctx, &elementsDestinationNetworkInterfaceInput, false)
					resp.Diagnostics.Append(diags...)

					var itemDestinationNetworkInterfaceInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_NetworkInterface
					for _, item := range elementsDestinationNetworkInterfaceInput {
						diags = item.As(ctx, &itemDestinationNetworkInterfaceInput, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationNetworkInterfaceInput)
						if err != nil {
							resp.Diagnostics.AddError(
								"Object Ref transformation failed",
								err.Error(),
							)
							return
						}

						exceptionInput.Destination.NetworkInterface = append(exceptionInput.Destination.NetworkInterface, &cato_models.NetworkInterfaceRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
				}

				// setting destination site network subnet
				if !destinationInput.SiteNetworkSubnet.IsNull() {
					elementsDestinationSiteNetworkSubnetInput := make([]types.Object, 0, len(destinationInput.SiteNetworkSubnet.Elements()))
					diags = destinationInput.SiteNetworkSubnet.ElementsAs(ctx, &elementsDestinationSiteNetworkSubnetInput, false)
					resp.Diagnostics.Append(diags...)

					var itemDestinationSiteNetworkSubnetInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_SiteNetworkSubnet
					for _, item := range elementsDestinationSiteNetworkSubnetInput {
						diags = item.As(ctx, &itemDestinationSiteNetworkSubnetInput, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationSiteNetworkSubnetInput)
						if err != nil {
							resp.Diagnostics.AddError(
								"Object Ref transformation failed",
								err.Error(),
							)
							return
						}

						exceptionInput.Destination.SiteNetworkSubnet = append(exceptionInput.Destination.SiteNetworkSubnet, &cato_models.SiteNetworkSubnetRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
				}

				// setting destination floating subnet
				if !destinationInput.FloatingSubnet.IsNull() {
					elementsDestinationFloatingSubnetInput := make([]types.Object, 0, len(destinationInput.FloatingSubnet.Elements()))
					diags = destinationInput.FloatingSubnet.ElementsAs(ctx, &elementsDestinationFloatingSubnetInput, false)
					resp.Diagnostics.Append(diags...)

					var itemDestinationFloatingSubnetInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_FloatingSubnet
					for _, item := range elementsDestinationFloatingSubnetInput {
						diags = item.As(ctx, &itemDestinationFloatingSubnetInput, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationFloatingSubnetInput)
						if err != nil {
							resp.Diagnostics.AddError(
								"Object Ref transformation failed",
								err.Error(),
							)
							return
						}

						exceptionInput.Destination.FloatingSubnet = append(exceptionInput.Destination.FloatingSubnet, &cato_models.FloatingSubnetRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
				}

				// setting destination user
				if !destinationInput.User.IsNull() {
					elementsDestinationUserInput := make([]types.Object, 0, len(destinationInput.User.Elements()))
					diags = destinationInput.User.ElementsAs(ctx, &elementsDestinationUserInput, false)
					resp.Diagnostics.Append(diags...)

					var itemDestinationUserInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_User
					for _, item := range elementsDestinationUserInput {
						diags = item.As(ctx, &itemDestinationUserInput, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationUserInput)
						if err != nil {
							resp.Diagnostics.AddError(
								"Object Ref transformation failed",
								err.Error(),
							)
							return
						}

						exceptionInput.Destination.User = append(exceptionInput.Destination.User, &cato_models.UserRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
				}

				// setting destination users group
				if !destinationInput.UsersGroup.IsNull() {
					elementsDestinationUsersGroupInput := make([]types.Object, 0, len(destinationInput.UsersGroup.Elements()))
					diags = destinationInput.UsersGroup.ElementsAs(ctx, &elementsDestinationUsersGroupInput, false)
					resp.Diagnostics.Append(diags...)

					var itemDestinationUsersGroupInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_UsersGroup
					for _, item := range elementsDestinationUsersGroupInput {
						diags = item.As(ctx, &itemDestinationUsersGroupInput, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationUsersGroupInput)
						if err != nil {
							resp.Diagnostics.AddError(
								"Object Ref transformation failed",
								err.Error(),
							)
							return
						}

						exceptionInput.Destination.UsersGroup = append(exceptionInput.Destination.UsersGroup, &cato_models.UsersGroupRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
				}

				// setting destination group
				if !destinationInput.Group.IsNull() {
					elementsDestinationGroupInput := make([]types.Object, 0, len(destinationInput.Group.Elements()))
					diags = destinationInput.Group.ElementsAs(ctx, &elementsDestinationGroupInput, false)
					resp.Diagnostics.Append(diags...)

					var itemDestinationGroupInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_Group
					for _, item := range elementsDestinationGroupInput {
						diags = item.As(ctx, &itemDestinationGroupInput, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationGroupInput)
						if err != nil {
							resp.Diagnostics.AddError(
								"Object Ref transformation failed",
								err.Error(),
							)
							return
						}

						exceptionInput.Destination.Group = append(exceptionInput.Destination.Group, &cato_models.GroupRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
				}

				// setting destination system group
				if !destinationInput.SystemGroup.IsNull() {
					elementsDestinationSystemGroupInput := make([]types.Object, 0, len(destinationInput.SystemGroup.Elements()))
					diags = destinationInput.SystemGroup.ElementsAs(ctx, &elementsDestinationSystemGroupInput, false)
					resp.Diagnostics.Append(diags...)

					var itemDestinationSystemGroupInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Destination_SystemGroup
					for _, item := range elementsDestinationSystemGroupInput {
						diags = item.As(ctx, &itemDestinationSystemGroupInput, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationSystemGroupInput)
						if err != nil {
							resp.Diagnostics.AddError(
								"Object Ref transformation failed",
								err.Error(),
							)
							return
						}

						exceptionInput.Destination.SystemGroup = append(exceptionInput.Destination.SystemGroup, &cato_models.SystemGroupRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
				}
			}

			// setting application
			if !itemExceptionsInput.Application.IsNull() {

				exceptionInput.Application = &cato_models.WanFirewallApplicationInput{}
				applicationInput := Policy_Policy_WanFirewall_Policy_Rules_Rule_Application{}
				diags = itemExceptionsInput.Application.As(ctx, &applicationInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				// setting application IP
				if !applicationInput.IP.IsNull() {
					diags = applicationInput.IP.ElementsAs(ctx, &exceptionInput.Application.IP, false)
					resp.Diagnostics.Append(diags...)
				}

				// setting application subnet
				if !applicationInput.Subnet.IsNull() {
					diags = applicationInput.Subnet.ElementsAs(ctx, &exceptionInput.Application.Subnet, false)
					resp.Diagnostics.Append(diags...)
				}

				// setting application domain
				if !applicationInput.Domain.IsNull() {
					diags = applicationInput.Domain.ElementsAs(ctx, &exceptionInput.Application.Domain, false)
					resp.Diagnostics.Append(diags...)
				}

				// setting application fqdn
				if !applicationInput.Fqdn.IsNull() {
					diags = applicationInput.Fqdn.ElementsAs(ctx, &exceptionInput.Application.Fqdn, false)
					resp.Diagnostics.Append(diags...)
				}

				// setting application application
				if !applicationInput.Application.IsNull() {
					elementsApplicationApplicationInput := make([]types.Object, 0, len(applicationInput.Application.Elements()))
					diags = applicationInput.Application.ElementsAs(ctx, &elementsApplicationApplicationInput, false)
					resp.Diagnostics.Append(diags...)

					var itemApplicationApplicationInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_Application
					for _, item := range elementsApplicationApplicationInput {
						diags = item.As(ctx, &itemApplicationApplicationInput, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationApplicationInput)
						if err != nil {
							resp.Diagnostics.AddError(
								"Object Ref transformation failed",
								err.Error(),
							)
							return
						}

						exceptionInput.Application.Application = append(exceptionInput.Application.Application, &cato_models.ApplicationRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
				}

				// setting application custom app
				if !applicationInput.CustomApp.IsNull() {
					elementsApplicationCustomAppInput := make([]types.Object, 0, len(applicationInput.CustomApp.Elements()))
					diags = applicationInput.CustomApp.ElementsAs(ctx, &elementsApplicationCustomAppInput, false)
					resp.Diagnostics.Append(diags...)

					var itemApplicationCustomAppInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_CustomApp
					for _, item := range elementsApplicationCustomAppInput {
						diags = item.As(ctx, &itemApplicationCustomAppInput, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationCustomAppInput)
						if err != nil {
							resp.Diagnostics.AddError(
								"Object Ref transformation failed",
								err.Error(),
							)
							return
						}

						exceptionInput.Application.CustomApp = append(exceptionInput.Application.CustomApp, &cato_models.CustomApplicationRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
				}

				// setting application ip range
				if !applicationInput.IPRange.IsNull() {
					elementsApplicationIPRangeInput := make([]types.Object, 0, len(applicationInput.IPRange.Elements()))
					diags = applicationInput.IPRange.ElementsAs(ctx, &elementsApplicationIPRangeInput, false)
					resp.Diagnostics.Append(diags...)

					var itemApplicationIPRangeInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_IPRange
					for _, item := range elementsApplicationIPRangeInput {
						diags = item.As(ctx, &itemApplicationIPRangeInput, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						exceptionInput.Application.IPRange = append(exceptionInput.Application.IPRange, &cato_models.IPAddressRangeInput{
							From: itemApplicationIPRangeInput.From.ValueString(),
							To:   itemApplicationIPRangeInput.To.ValueString(),
						})
					}
				}

				// setting application global ip range
				if !applicationInput.GlobalIPRange.IsNull() {
					elementsApplicationGlobalIPRangeInput := make([]types.Object, 0, len(applicationInput.GlobalIPRange.Elements()))
					diags = applicationInput.GlobalIPRange.ElementsAs(ctx, &elementsApplicationGlobalIPRangeInput, false)
					resp.Diagnostics.Append(diags...)

					var itemApplicationGlobalIPRangeInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_GlobalIPRange
					for _, item := range elementsApplicationGlobalIPRangeInput {
						diags = item.As(ctx, &itemApplicationGlobalIPRangeInput, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationGlobalIPRangeInput)
						if err != nil {
							resp.Diagnostics.AddError(
								"Object Ref transformation failed",
								err.Error(),
							)
							return
						}

						exceptionInput.Application.GlobalIPRange = append(exceptionInput.Application.GlobalIPRange, &cato_models.GlobalIPRangeRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
				}

				// setting application app category
				if !applicationInput.AppCategory.IsNull() {
					elementsApplicationAppCategoryInput := make([]types.Object, 0, len(applicationInput.AppCategory.Elements()))
					diags = applicationInput.AppCategory.ElementsAs(ctx, &elementsApplicationAppCategoryInput, false)
					resp.Diagnostics.Append(diags...)

					var itemApplicationAppCategoryInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_AppCategory
					for _, item := range elementsApplicationAppCategoryInput {
						diags = item.As(ctx, &itemApplicationAppCategoryInput, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationAppCategoryInput)
						if err != nil {
							resp.Diagnostics.AddError(
								"Object Ref transformation failed",
								err.Error(),
							)
							return
						}

						exceptionInput.Application.AppCategory = append(exceptionInput.Application.AppCategory, &cato_models.ApplicationCategoryRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
				}

				// setting application custom app category
				if !applicationInput.CustomCategory.IsNull() {
					elementsApplicationCustomCategoryInput := make([]types.Object, 0, len(applicationInput.CustomCategory.Elements()))
					diags = applicationInput.CustomCategory.ElementsAs(ctx, &elementsApplicationCustomCategoryInput, false)
					resp.Diagnostics.Append(diags...)

					var itemApplicationCustomCategoryInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_CustomCategory
					for _, item := range elementsApplicationCustomCategoryInput {
						diags = item.As(ctx, &itemApplicationCustomCategoryInput, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationCustomCategoryInput)
						if err != nil {
							resp.Diagnostics.AddError(
								"Object Ref transformation failed",
								err.Error(),
							)
							return
						}

						exceptionInput.Application.CustomCategory = append(exceptionInput.Application.CustomCategory, &cato_models.CustomCategoryRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
				}

				// setting application sanctionned apps category
				if !applicationInput.SanctionedAppsCategory.IsNull() {
					elementsApplicationSanctionedAppsCategoryInput := make([]types.Object, 0, len(applicationInput.SanctionedAppsCategory.Elements()))
					diags = applicationInput.SanctionedAppsCategory.ElementsAs(ctx, &elementsApplicationSanctionedAppsCategoryInput, false)
					resp.Diagnostics.Append(diags...)

					var itemApplicationSanctionedAppsCategoryInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Application_SanctionedAppsCategory
					for _, item := range elementsApplicationSanctionedAppsCategoryInput {
						diags = item.As(ctx, &itemApplicationSanctionedAppsCategoryInput, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemApplicationSanctionedAppsCategoryInput)
						if err != nil {
							resp.Diagnostics.AddError(
								"Object Ref transformation failed",
								err.Error(),
							)
							return
						}

						exceptionInput.Application.SanctionedAppsCategory = append(exceptionInput.Application.SanctionedAppsCategory, &cato_models.SanctionedAppsCategoryRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
				}
			}

			// setting service
			if !itemExceptionsInput.Service.IsNull() {

				exceptionInput.Service = &cato_models.WanFirewallServiceTypeInput{}
				serviceInput := Policy_Policy_WanFirewall_Policy_Rules_Rule_Service{}
				diags = itemExceptionsInput.Service.As(ctx, &serviceInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}

				// setting service standard
				if !serviceInput.Standard.IsNull() {
					elementsServiceStandardInput := make([]types.Object, 0, len(serviceInput.Standard.Elements()))
					diags = serviceInput.Standard.ElementsAs(ctx, &elementsServiceStandardInput, false)
					resp.Diagnostics.Append(diags...)
					if resp.Diagnostics.HasError() {
						return
					}

					var itemServiceStandardInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Service_Standard
					for _, item := range elementsServiceStandardInput {
						diags = item.As(ctx, &itemServiceStandardInput, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemServiceStandardInput)
						if err != nil {
							resp.Diagnostics.AddError(
								"Object Ref transformation failed",
								err.Error(),
							)
							return
						}

						exceptionInput.Service.Standard = append(exceptionInput.Service.Standard, &cato_models.ServiceRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
				}

				// setting service custom
				if !serviceInput.Custom.IsNull() {
					elementsServiceCustomInput := make([]types.Object, 0, len(serviceInput.Custom.Elements()))
					diags = serviceInput.Custom.ElementsAs(ctx, &elementsServiceCustomInput, false)
					resp.Diagnostics.Append(diags...)
					if resp.Diagnostics.HasError() {
						return
					}

					var itemServiceCustomInput Policy_Policy_WanFirewall_Policy_Rules_Rule_Service_Custom
					for _, item := range elementsServiceCustomInput {
						diags = item.As(ctx, &itemServiceCustomInput, basetypes.ObjectAsOptions{})

						customInput := &cato_models.CustomServiceInput{
							Protocol: cato_models.IPProtocol(itemServiceCustomInput.Protocol.ValueString()),
						}

						// setting service custom port
						if !itemServiceCustomInput.Port.IsNull() {
							elementsPort := make([]types.String, 0, len(itemServiceCustomInput.Port.Elements()))
							diags = itemServiceCustomInput.Port.ElementsAs(ctx, &elementsPort, false)
							resp.Diagnostics.Append(diags...)

							inputPort := []cato_scalars.Port{}
							for _, item := range elementsPort {
								inputPort = append(inputPort, cato_scalars.Port(item.ValueString()))
							}

							customInput.Port = inputPort
						}

						// setting service custom port range
						if !itemServiceCustomInput.PortRange.IsNull() {
							var itemPortRange Policy_Policy_WanFirewall_Policy_Rules_Rule_Service_Custom_PortRange
							diags = itemServiceCustomInput.PortRange.As(ctx, &itemPortRange, basetypes.ObjectAsOptions{})

							inputPortRange := cato_models.PortRangeInput{
								From: cato_scalars.Port(itemPortRange.From.ValueString()),
								To:   cato_scalars.Port(itemPortRange.To.ValueString()),
							}

							customInput.PortRange = &inputPortRange
						}

						// append custom service
						exceptionInput.Service.Custom = append(exceptionInput.Service.Custom, customInput)
					}
				}
			}

			input.Rule.Exceptions = append(input.Rule.Exceptions, &exceptionInput)

		}
	}

	//setting at (to move rule)
	if !plan.At.IsNull() {
		inputMoveRule.To = &cato_models.PolicyRulePositionInput{}
		positionInput := PolicyRulePositionInput{}
		diags = plan.At.As(ctx, &positionInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)

		inputMoveRule.To.Position = (*cato_models.PolicyRulePositionEnum)(positionInput.Position.ValueStringPointer())
		inputMoveRule.To.Ref = positionInput.Ref.ValueStringPointer()
	}

	// settings other rule attributes
	inputMoveRule.ID = *ruleInput.ID.ValueStringPointer()
	input.ID = *ruleInput.ID.ValueStringPointer()
	input.Rule.Name = ruleInput.Name.ValueStringPointer()
	input.Rule.Description = ruleInput.Description.ValueStringPointer()
	input.Rule.Enabled = ruleInput.Enabled.ValueBoolPointer()
	input.Rule.Action = (*cato_models.WanFirewallActionEnum)(ruleInput.Action.ValueStringPointer())
	input.Rule.Direction = (*cato_models.WanFirewallDirectionEnum)(ruleInput.Direction.ValueStringPointer())
	if !ruleInput.ConnectionOrigin.IsNull() {
		input.Rule.ConnectionOrigin = (*cato_models.ConnectionOriginEnum)(ruleInput.ConnectionOrigin.ValueStringPointer())
	} else {
		connectionOrigin := "ANY"
		input.Rule.ConnectionOrigin = (*cato_models.ConnectionOriginEnum)(&connectionOrigin)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "wan_fw_rule move", map[string]interface{}{
		"input": utils.InterfaceToJSONString(inputMoveRule),
	})

	//move rule
	moveRule, err := r.client.catov2.PolicyWanFirewallMoveRule(ctx, inputMoveRule, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanFirewallMoveRule error",
			err.Error(),
		)
		return
	}

	// check for errors
	if moveRule.Policy.WanFirewall.MoveRule.Status != "SUCCESS" {
		for _, item := range moveRule.Policy.WanFirewall.MoveRule.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Moving Rule Resource",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	tflog.Debug(ctx, "wan_fw_policy update", map[string]interface{}{
		"input": utils.InterfaceToJSONString(input),
	})

	//creating new rule
	updateRule, err := r.client.catov2.PolicyWanFirewallUpdateRule(ctx, input, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanFirewallUpdateRule error",
			err.Error(),
		)
		return
	}

	// check for errors
	if updateRule.Policy.WanFirewall.UpdateRule.Status != "SUCCESS" {
		for _, item := range updateRule.Policy.WanFirewall.UpdateRule.GetErrors() {
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
	_, err = r.client.catov2.PolicyWanFirewallPublishPolicyRevision(ctx, publishDataIfEnabled, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanFirewallPublishPolicyRevision error",
			err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *wanFwRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	var state WanFirewallRule
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	//retrieve rule ID
	rule := Policy_Policy_WanFirewall_Policy_Rules_Rule{}
	diags = state.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	removeRule := cato_models.WanFirewallRemoveRuleInput{
		ID: rule.ID.ValueString(),
	}
	tflog.Debug(ctx, "wan_fw_policy delete", map[string]interface{}{
		"input": utils.InterfaceToJSONString(removeRule),
	})

	_, err := r.client.catov2.PolicyWanFirewallRemoveRule(ctx, removeRule, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to connect or request the Catov2 API",
			err.Error(),
		)
		return
	}

	publishDataIfEnabled := &cato_models.PolicyPublishRevisionInput{}
	_, err = r.client.catov2.PolicyWanFirewallPublishPolicyRevision(ctx, publishDataIfEnabled, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API Delete/PolicyWanFirewallPublishPolicyRevision error",
			err.Error(),
		)
		return
	}
}
