package provider

import (
	"context"
	"encoding/json"
	"fmt"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
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
						Description: "Ruile index",
						Required:    false,
						Optional:    true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(), // Avoid drift
						},
						Computed: true,
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
						Validators: []validator.String{
							stringvalidator.OneOf("TO", "BOTH"),
						},
					},
					"source": schema.SingleNestedAttribute{
						Description: "Source traffic matching criteria. Logical ‘OR’ is applied within the criteria set. Logical ‘AND’ is applied between criteria sets.",
						Required:    false,
						Optional:    true,
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(), // Avoid drift
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
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"from": schema.StringAttribute{
											Description: "IP Range Name",
											Required:    false,
											Optional:    true,
										},
										"to": schema.StringAttribute{
											Description: "IP Range ID",
											Required:    false,
											Optional:    true,
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
											Description: "Global IP Range",
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
											Description: "User Group Name",
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
											Description: "User Group ID",
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
											Description: "IP Range Name",
											Required:    false,
											Optional:    true,
										},
										"to": schema.StringAttribute{
											Description: "IP Range ID",
											Required:    false,
											Optional:    true,
										},
									},
								},
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
							"global_ip_range": schema.SetNestedAttribute{
								Description: "Globally defined IP range, IP and subnet objects",
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
											Computed: true,
										},
									},
								},
							},
							"site_network_subnet": schema.SetNestedAttribute{
								Description: "GlobalRange + InterfaceSubnet",
								Required:    false,
								Optional:    true,
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
							"floating_subnet": schema.SetNestedAttribute{
								Description: "Floating Subnets (ie. Floating Ranges) are used to identify traffic exactly matched to the route advertised by BGP. They are not associated with a specific site. This is useful in scenarios such as active-standby high availability routed via BGP.",
								Required:    false,
								Optional:    true,
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
												stringplanmodifier.UseStateForUnknown(),
											},
											Computed: true,
										},
										"id": schema.StringAttribute{
											Description: "Floating Subnet ID",
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
								Description: "Group of users",
								Required:    false,
								Optional:    true,
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
								Description: "Groups defined for your account",
								Required:    false,
								Optional:    true,
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
								Description: "Predefined Cato groups",
								Required:    false,
								Optional:    true,
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
						Description: "Source Device Profile traffic matching criteria. Logical ‘OR’ is applied within the criteria set. Logical ‘AND’ is applied between criteria sets.",
						Required:    false,
						Optional:    true,
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
					"device_os": schema.ListAttribute{
						ElementType: types.StringType,
						Description: "Source device Operating System traffic matching criteria. Logical ‘OR’ is applied within the criteria set. Logical ‘AND’ is applied between criteria sets.(https://api.catonetworks.com/documentation/#definition-OperatingSystem)",
						Optional:    true,
						Required:    false,
						Validators: []validator.List{
							listvalidator.SizeAtLeast(1),
						},
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
											Required:    false,
											Optional:    true,
										},
										"to": schema.StringAttribute{
											Description: "IP Range ID",
											Required:    false,
											Optional:    true,
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
										Description: "Enable event creation",
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
										Description: "Alert creation enabled",
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
													Description: "Subscription Group Name",
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
													Description: "Subscription Group ID",
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
													Description: "Webhook Name",
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
													Description: "Webhook ID",
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
													Description: "Mailing List Name",
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
													Description: "Mailing List ID",
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
										Description: "Custom Timeframe Name",
										Required:    false,
										Optional:    true,
									},
									"to": schema.StringAttribute{
										Description: "Custom Timeframe ID",
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
										Description: "Custom Recurring Name",
										Required:    false,
										Optional:    true,
									},
									"to": schema.StringAttribute{
										Description: "Custom Recurring ID",
										Required:    false,
										Optional:    true,
									},
									"days": schema.ListAttribute{
										ElementType: types.StringType,
										Description: "Custom Recurring Days - (https://api.catonetworks.com/documentation/#definition-DayOfWeek)",
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
									Required:    true,
									Optional:    false,
									PlanModifiers: []planmodifier.Object{
										objectplanmodifier.UseStateForUnknown(), // Avoid drift
									},
									Attributes: map[string]schema.Attribute{
										"ip": schema.ListAttribute{
											Description: "Source IP traffic matching criteria. Logical ‘OR’ is applied within the criteria set. Logical ‘AND’ is applied between criteria sets.",
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
											//Computed: true,
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
								"destination": schema.SingleNestedAttribute{
									Description: "Destination traffic matching criteria for the exception.",
									Required:    false,
									Optional:    true,
									PlanModifiers: []planmodifier.Object{
										objectplanmodifier.UseStateForUnknown(), // Avoid drift
									},
									Attributes: map[string]schema.Attribute{
										"ip": schema.ListAttribute{
											Description: "Pv4 address list",
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
														Description: "Hst ID",
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
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											PlanModifiers: []planmodifier.Set{
												setplanmodifier.UseStateForUnknown(), // Avoid drift
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
											Description: "Subnets and network ranges defined for the LAN interfaces of a site",
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
											Description: "Multiple separate IP addresses or an IP range",
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
											Description: "Globally defined IP range, IP and subnet objects",
											Required:    false,
											Optional:    true,
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											// PlanModifiers: []planmodifier.Set{
											// 	setplanmodifier.UseStateForUnknown(), // Avoid drift
											// },
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
														Computed: true,
													},
												},
											},
										},
										"site_network_subnet": schema.SetNestedAttribute{
											Description: "GlobalRange + InterfaceSubnet",
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
											Description: "Floating Subnets (ie. Floating Ranges) are used to identify traffic exactly matched to the route advertised by BGP. They are not associated with a specific site. This is useful in scenarios such as active-standby high availability routed via BGP.",
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
														Description: "Floating Subnet Name",
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
														Description: "Floating Subnet ID",
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
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											PlanModifiers: []planmodifier.Set{
												setplanmodifier.UseStateForUnknown(), // Avoid drift
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
											Description: "Group of users",
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
											Description: "Groups defined for your account",
											Required:    false,
											Optional:    true,
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
											Description: "Predefined Cato groups",
											Required:    false,
											Optional:    true,
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
								"application": schema.SingleNestedAttribute{
									Description: "Application matching criteria for the exception.",
									Optional:    true,
									Required:    false,
									PlanModifiers: []planmodifier.Object{
										objectplanmodifier.UseStateForUnknown(), // Avoid drift
									},
									Attributes: map[string]schema.Attribute{
										"application": schema.SetNestedAttribute{
											Description: "Application defined for your account",
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
														Description: "Application Category Name",
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
														Description: "Application Category ID",
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
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											PlanModifiers: []planmodifier.Set{
												setplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"name": schema.StringAttribute{
														Description: "Custom Application Category Name",
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
														Description: "Custom Application Category ID",
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
														Description: "Sanctioned Application Category Name",
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
														Description: "Sanctioned Application Category ID",
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
											Description: "Domain names matching criteria for the exception.",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.List{
												listplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Validators: []validator.List{
												listvalidator.SizeAtLeast(1),
											},
										},
										"fqdn": schema.ListAttribute{
											ElementType: types.StringType,
											Description: "Fully Qualified Domain Names matching criteria for the exception.",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.List{
												listplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Validators: []validator.List{
												listvalidator.SizeAtLeast(1),
											},
										},
										"ip": schema.ListAttribute{
											ElementType: types.StringType,
											Description: "IPv4 address list matching criteria for the exception.",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.List{
												listplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Validators: []validator.List{
												listvalidator.SizeAtLeast(1),
											},
										},
										"subnet": schema.ListAttribute{
											ElementType: types.StringType,
											Description: "Subnets and network ranges matching criteria for the exception.",
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
											Description: "IP range matching criteria for the exception.",
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
														Description: "IP Range From",
														Required:    true,
														Optional:    false,
														PlanModifiers: []planmodifier.String{
															stringplanmodifier.UseStateForUnknown(), // Avoid drift
														},
													},
													"to": schema.StringAttribute{
														Description: "IP Range To",
														Required:    true,
														Optional:    false,
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
														Description: "Standard Service Name",
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
														Description: "Standard Service ID",
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
														Description: "Custom Service Port",
														Optional:    true,
														Required:    false,
													},
													"port_range": schema.SingleNestedAttribute{
														Required: false,
														Optional: true,
														Attributes: map[string]schema.Attribute{
															"from": schema.StringAttribute{
																Description: "Port Range From",
																Required:    true,
																Optional:    false,
															},
															"to": schema.StringAttribute{
																Description: "Port Range To",
																Required:    true,
																Optional:    false,
															},
														},
													},
													"protocol": schema.StringAttribute{
														Description: "Protocol matching criteria for the exception.",
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

	input, diags := hydrateWanRuleApi(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Warn(ctx, "TFLOG_WARN_WAN_input.create", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(input.create),
	})

	createRuleResponse, err := r.client.catov2.PolicyWanFirewallAddRule(ctx, input.create, r.client.AccountId)

	tflog.Warn(ctx, "TFLOG_WARN_WAN_createRuleResponse", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(createRuleResponse),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanFirewallAddRule error",
			err.Error(),
		)
		return
	}

	// check for errors
	if createRuleResponse.Policy.WanFirewall.AddRule.Status != "SUCCESS" {
		for _, item := range createRuleResponse.Policy.WanFirewall.AddRule.GetErrors() {
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

	// Read rule and hydrate response to state
	queryWanPolicy := &cato_models.WanFirewallPolicyInput{}
	body, err := r.client.catov2.PolicyWanFirewall(ctx, queryWanPolicy, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanFirewall error",
			err.Error(),
		)
		return
	}

	ruleList := body.GetPolicy().WanFirewall.Policy.GetRules()
	currentRule := &cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules_Rule{}
	// Get current rule from response by ID
	for _, ruleListItem := range ruleList {
		if ruleListItem.GetRule().ID == createRuleResponse.GetPolicy().GetWanFirewall().GetAddRule().Rule.GetRule().ID {
			currentRule = ruleListItem.GetRule()
			resp.State.SetAttribute(
				ctx,
				path.Root("rule").AtName("id"),
				ruleListItem.GetRule().ID)
			break
		}
	}
	tflog.Info(ctx, "ruleObject - "+fmt.Sprintf("%v", currentRule))
	// Hydrate ruleInput from api respoonse
	ruleInputRead, hydrateDiags := hydrateWanRuleState(ctx, plan, currentRule)
	resp.Diagnostics.Append(hydrateDiags...)
	ruleInputRead.ID = types.StringValue(currentRule.ID)
	tflog.Info(ctx, "ruleInputRead - "+fmt.Sprintf("%v", ruleInputRead))
	ruleObject, diags := types.ObjectValueFrom(ctx, WanFirewallRuleRuleAttrTypes, ruleInputRead)
	tflog.Info(ctx, "ruleObject - "+fmt.Sprintf("%v", ruleObject))
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

func (r *wanFwRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WanFirewallRule
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	queryWanPolicy := &cato_models.WanFirewallPolicyInput{}
	body, err := r.client.catov2.PolicyWanFirewall(ctx, queryWanPolicy, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanFirewall error",
			err.Error(),
		)
		return
	}

	//retrieve rule ID
	rule := Policy_Policy_WanFirewall_Policy_Rules_Rule{}
	diags = state.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(diags...)

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

	ruleInput, hydrateDiags := hydrateWanRuleState(ctx, state, currentRule)
	resp.Diagnostics.Append(hydrateDiags...)

	diags = resp.State.SetAttribute(ctx, path.Root("rule"), ruleInput)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(diags...)

	// Hard coding LAST_IN_POLICY position as the API does not return any value and
	// hardcoding position supports the use case of bulk rule import/export
	// getting around state changes for the position field
	curAtObj, diagstmp := types.ObjectValue(
		PositionAttrTypes,
		map[string]attr.Value{
			"position": types.StringValue("LAST_IN_POLICY"),
			"ref":      types.StringNull(),
		},
	)
	diags = resp.State.SetAttribute(ctx, path.Root("at"), curAtObj)
	diags = append(diags, diagstmp...)

}

func (r *wanFwRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	var plan WanFirewallRule
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := hydrateWanRuleApi(ctx, plan)
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

	ruleInput := Policy_Policy_WanFirewall_Policy_Rules_Rule{}
	diags = plan.Rule.As(ctx, &ruleInput, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)

	// settings other rule attributes
	inputMoveRule.ID = *ruleInput.ID.ValueStringPointer()
	input.update.ID = *ruleInput.ID.ValueStringPointer()

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

	tflog.Warn(ctx, "TFLOG_WARN_WAN_input.update", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(input.update),
	})

	//creating new rule
	updateRuleResponse, err := r.client.catov2.PolicyWanFirewallUpdateRule(ctx, input.update, r.client.AccountId)
	tflog.Debug(ctx, "updateRuleResponse", map[string]interface{}{
		"input.update": utils.InterfaceToJSONString(input.update),
	})

	tflog.Warn(ctx, "TFLOG_WARN_WAN_updateRuleResponse", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(updateRuleResponse),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanFirewallUpdateRule error",
			err.Error(),
		)
		return
	}

	// check for errors
	if updateRuleResponse.Policy.WanFirewall.UpdateRule.Status != "SUCCESS" {
		for _, item := range updateRuleResponse.Policy.WanFirewall.UpdateRule.GetErrors() {
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

	// Read rule and hydrate response to state
	queryWanPolicy := &cato_models.WanFirewallPolicyInput{}
	wanFWQueryResponse, err := r.client.catov2.PolicyWanFirewall(ctx, queryWanPolicy, r.client.AccountId)
	tflog.Debug(ctx, "wanFWQueryResponse", map[string]interface{}{
		"wanFWQueryResponse": utils.InterfaceToJSONString(wanFWQueryResponse),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanFirewall error",
			err.Error(),
		)
		return
	}

	ruleList := wanFWQueryResponse.GetPolicy().WanFirewall.Policy.GetRules()
	currentRule := &cato_go_sdk.Policy_Policy_WanFirewall_Policy_Rules_Rule{}
	// Get current rule from response by ID
	for _, ruleListItem := range ruleList {
		if ruleListItem.GetRule().ID == updateRuleResponse.GetPolicy().GetWanFirewall().GetUpdateRule().Rule.GetRule().ID {
			currentRule = ruleListItem.GetRule()
			break
		}
	}

	policyChangeJson, _ := json.Marshal(currentRule)

	tflog.Warn(ctx, "TFLOG_WARN_WAN_currentRule", map[string]interface{}{
		"OUTPUT": string(policyChangeJson),
	})

	// Hydrate ruleInput from api respoonse
	ruleInputRead, hydrateDiags := hydrateWanRuleState(ctx, plan, currentRule)
	resp.Diagnostics.Append(hydrateDiags...)

	ruleInputRead.ID = types.StringValue(currentRule.ID)
	ruleObject, diags := types.ObjectValueFrom(ctx, WanFirewallRuleRuleAttrTypes, ruleInputRead)
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
