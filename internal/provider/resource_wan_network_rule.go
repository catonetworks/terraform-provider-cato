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
	_ resource.Resource                = &wanNetworkRuleResource{}
	_ resource.ResourceWithConfigure   = &wanNetworkRuleResource{}
	_ resource.ResourceWithImportState = &wanNetworkRuleResource{}
)

func NewWanNetworkRuleResource() resource.Resource {
	return &wanNetworkRuleResource{}
}

type wanNetworkRuleResource struct {
	client *catoClientData
}

func (r *wanNetworkRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wan_nw_rule"
}

func (r *wanNetworkRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_wan_nw_rule` resource contains the configuration parameters necessary to add a rule to the WAN Network policy.",
		Attributes: map[string]schema.Attribute{
			"at": schema.SingleNestedAttribute{
				Description: "Position of the rule in the policy",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"position": schema.StringAttribute{
						Description: "Position relative to a policy or another rule",
						Required:    true,
					},
					"ref": schema.StringAttribute{
						Description: "The identifier of the object relative to which the position is defined",
						Optional:    true,
					},
				},
			},
			"rule": schema.SingleNestedAttribute{
				Description: "Parameters for the WAN Network rule",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Description: "ID of the rule",
						Computed:    true,
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
						Optional:    true,
					},
					"index": schema.Int64Attribute{
						Description: "Rule Index - computed value that may change due to rule reordering",
						Computed:    true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"enabled": schema.BoolAttribute{
						Description: "Whether the rule is enabled",
						Required:    true,
					},
					"rule_type": schema.StringAttribute{
						Description: "Type of WAN rule (INTERNET or WAN)",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("INBOUND_INTERNET", "INTERNET", "WAN"),
						},
					},
					"route_type": schema.StringAttribute{
						Description: "Routing method for the rule",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("BACKHAUL", "BACKHAUL_HAIRPINNING", "NAT", "NONE", "OPTIMIZED", "VIA"),
						},
					},
					"source": schema.SingleNestedAttribute{
						Description: "Source traffic matching criteria",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"ip": schema.ListAttribute{
								Description: "IPv4 address list",
								ElementType: types.StringType,
								Optional:    true,
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
							"subnet": schema.ListAttribute{
								Description: "Subnets in CIDR notation",
								ElementType: types.StringType,
								Optional:    true,
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
							"ip_range": schema.ListNestedAttribute{
								Description: "IP address ranges",
								Optional:    true,
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"from": schema.StringAttribute{
											Description: "Start IP",
											Optional:    true,
										},
										"to": schema.StringAttribute{
											Description: "End IP",
											Optional:    true,
										},
									},
								},
							},
							"host": schema.SetNestedAttribute{
								Description: "Hosts defined for your account",
								Optional:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "Host ID",
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"name": schema.StringAttribute{
											Description: "Host Name",
											Optional:    true,
											Computed:    true,
											Validators: []validator.String{
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
							"site": schema.SetNestedAttribute{
								Description: "Sites defined for the account",
								Optional:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "Site ID",
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"name": schema.StringAttribute{
											Description: "Site Name",
											Optional:    true,
											Computed:    true,
											Validators: []validator.String{
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
							"global_ip_range": schema.SetNestedAttribute{
								Description: "Global IP ranges",
								Optional:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "Global IP Range ID",
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"name": schema.StringAttribute{
											Description: "Global IP Range Name",
											Optional:    true,
											Computed:    true,
											Validators: []validator.String{
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
							"network_interface": schema.SetNestedAttribute{
								Description: "Network interfaces",
								Optional:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "Network Interface ID",
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"name": schema.StringAttribute{
											Description: "Network Interface Name",
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
							"site_network_subnet": schema.SetNestedAttribute{
								Description: "Site network subnets",
								Optional:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "Site Network Subnet ID",
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"name": schema.StringAttribute{
											Description: "Site Network Subnet Name",
											Optional:    true,
											Computed:    true,
											Validators: []validator.String{
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
							"floating_subnet": schema.SetNestedAttribute{
								Description: "Floating subnets",
								Optional:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "Floating Subnet ID",
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"name": schema.StringAttribute{
											Description: "Floating Subnet Name",
											Optional:    true,
											Computed:    true,
											Validators: []validator.String{
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
							"user": schema.SetNestedAttribute{
								Description: "Individual users",
								Optional:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "User ID",
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"name": schema.StringAttribute{
											Description: "User Name",
											Optional:    true,
											Computed:    true,
											Validators: []validator.String{
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
							"users_group": schema.SetNestedAttribute{
								Description: "User groups",
								Optional:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "User Group ID",
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"name": schema.StringAttribute{
											Description: "User Group Name",
											Optional:    true,
											Computed:    true,
											Validators: []validator.String{
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
							"group": schema.SetNestedAttribute{
								Description: "Groups",
								Optional:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "Group ID",
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"name": schema.StringAttribute{
											Description: "Group Name",
											Optional:    true,
											Computed:    true,
											Validators: []validator.String{
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
							"system_group": schema.SetNestedAttribute{
								Description: "System groups",
								Optional:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "System Group ID",
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"name": schema.StringAttribute{
											Description: "System Group Name",
											Optional:    true,
											Computed:    true,
											Validators: []validator.String{
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
						},
					},
					"destination": schema.SingleNestedAttribute{
						Description: "Destination traffic matching criteria",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"ip": schema.ListAttribute{
								Description: "IPv4 address list",
								ElementType: types.StringType,
								Optional:    true,
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
							"subnet": schema.ListAttribute{
								Description: "Subnets in CIDR notation",
								ElementType: types.StringType,
								Optional:    true,
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
							"ip_range": schema.ListNestedAttribute{
								Description: "IP address ranges",
								Optional:    true,
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"from": schema.StringAttribute{
											Description: "Start IP",
											Optional:    true,
										},
										"to": schema.StringAttribute{
											Description: "End IP",
											Optional:    true,
										},
									},
								},
							},
							"host": schema.SetNestedAttribute{
								Description: "Hosts defined for your account",
								Optional:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "Host ID",
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"name": schema.StringAttribute{
											Description: "Host Name",
											Optional:    true,
											Computed:    true,
											Validators: []validator.String{
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
							"site": schema.SetNestedAttribute{
								Description: "Sites defined for the account",
								Optional:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "Site ID",
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"name": schema.StringAttribute{
											Description: "Site Name",
											Optional:    true,
											Computed:    true,
											Validators: []validator.String{
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
							"global_ip_range": schema.SetNestedAttribute{
								Description: "Global IP ranges",
								Optional:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "Global IP Range ID",
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"name": schema.StringAttribute{
											Description: "Global IP Range Name",
											Optional:    true,
											Computed:    true,
											Validators: []validator.String{
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
							"network_interface": schema.SetNestedAttribute{
								Description: "Network interfaces",
								Optional:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "Network Interface ID",
											Optional:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"name": schema.StringAttribute{
											Description: "Network Interface Name",
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
							"site_network_subnet": schema.SetNestedAttribute{
								Description: "Site network subnets",
								Optional:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "Site Network Subnet ID",
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"name": schema.StringAttribute{
											Description: "Site Network Subnet Name",
											Optional:    true,
											Computed:    true,
											Validators: []validator.String{
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
							"floating_subnet": schema.SetNestedAttribute{
								Description: "Floating subnets",
								Optional:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "Floating Subnet ID",
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"name": schema.StringAttribute{
											Description: "Floating Subnet Name",
											Optional:    true,
											Computed:    true,
											Validators: []validator.String{
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
							"user": schema.SetNestedAttribute{
								Description: "Individual users",
								Optional:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "User ID",
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"name": schema.StringAttribute{
											Description: "User Name",
											Optional:    true,
											Computed:    true,
											Validators: []validator.String{
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
							"users_group": schema.SetNestedAttribute{
								Description: "User groups",
								Optional:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "User Group ID",
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"name": schema.StringAttribute{
											Description: "User Group Name",
											Optional:    true,
											Computed:    true,
											Validators: []validator.String{
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
							"group": schema.SetNestedAttribute{
								Description: "Groups",
								Optional:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "Group ID",
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"name": schema.StringAttribute{
											Description: "Group Name",
											Optional:    true,
											Computed:    true,
											Validators: []validator.String{
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
							"system_group": schema.SetNestedAttribute{
								Description: "System groups",
								Optional:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "System Group ID",
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"name": schema.StringAttribute{
											Description: "System Group Name",
											Optional:    true,
											Computed:    true,
											Validators: []validator.String{
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
						},
					},
					"application": schema.SingleNestedAttribute{
						Description: "Application matching criteria",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"application": schema.SetNestedAttribute{
								Description: "Applications",
								Optional:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "Application ID",
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"name": schema.StringAttribute{
											Description: "Application Name",
											Optional:    true,
											Computed:    true,
											Validators: []validator.String{
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
							"custom_app": schema.SetNestedAttribute{
								Description: "Custom applications",
								Optional:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "Custom App ID",
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"name": schema.StringAttribute{
											Description: "Custom App Name",
											Optional:    true,
											Computed:    true,
											Validators: []validator.String{
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
							"app_category": schema.SetNestedAttribute{
								Description: "Application categories",
								Optional:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "App Category ID",
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"name": schema.StringAttribute{
											Description: "App Category Name",
											Optional:    true,
											Computed:    true,
											Validators: []validator.String{
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
							"custom_category": schema.SetNestedAttribute{
								Description: "Custom categories",
								Optional:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "Custom Category ID",
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"name": schema.StringAttribute{
											Description: "Custom Category Name",
											Optional:    true,
											Computed:    true,
											Validators: []validator.String{
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
							"domain": schema.ListAttribute{
								Description: "Domains",
								ElementType: types.StringType,
								Optional:    true,
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
							"fqdn": schema.ListAttribute{
								Description: "Fully qualified domain names",
								ElementType: types.StringType,
								Optional:    true,
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
							"service": schema.SetNestedAttribute{
								Description: "Services",
								Optional:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "Service ID",
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"name": schema.StringAttribute{
											Description: "Service Name",
											Optional:    true,
											Computed:    true,
											Validators: []validator.String{
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
							"custom_service": schema.ListNestedAttribute{
								Description: "Custom services",
								Optional:    true,
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"port": schema.ListAttribute{
											Description: "Port numbers",
											ElementType: types.StringType,
											Optional:    true,
										},
										"port_range": schema.SingleNestedAttribute{
											Description: "Port range",
											Optional:    true,
											Attributes: map[string]schema.Attribute{
												"from": schema.StringAttribute{
													Description: "Start port",
													Optional:    true,
												},
												"to": schema.StringAttribute{
													Description: "End port",
													Optional:    true,
												},
											},
										},
										"protocol": schema.StringAttribute{
											Description: "Protocol",
											Optional:    true,
											Validators: []validator.String{
												stringvalidator.OneOf("ANY", "ICMP", "TCP", "TCP_UDP", "UDP"),
											},
										},
									},
								},
							},
							"custom_service_ip": schema.ListNestedAttribute{
								Description: "Custom service IPs",
								Optional:    true,
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Description: "Name",
											Optional:    true,
										},
										"ip": schema.StringAttribute{
											Description: "IP address",
											Optional:    true,
										},
										"ip_range": schema.SingleNestedAttribute{
											Description: "IP range",
											Optional:    true,
											Attributes: map[string]schema.Attribute{
												"from": schema.StringAttribute{
													Description: "Start IP",
													Optional:    true,
												},
												"to": schema.StringAttribute{
													Description: "End IP",
													Optional:    true,
												},
											},
										},
									},
								},
							},
						},
					},
					"exceptions": schema.SetNestedAttribute{
						Description: "The set of exceptions for the rule. Exceptions define when the rule will be ignored and the WAN Network evaluation will continue with the lower priority rules.",
						Required:    false,
						Optional:    true,
						Computed:    true,
						Validators: []validator.Set{
							setvalidator.SizeAtLeast(1),
						},
						PlanModifiers: []planmodifier.Set{
							planmodifiers.WanExceptionsSetModifier(), // Handle ID correlation for exceptions
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
											Description: "Source IP traffic matching criteria. Logical 'OR' is applied within the criteria set. Logical 'AND' is applied between criteria sets.",
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
															stringplanmodifier.UseStateForUnknown(),
														},
														Computed: true,
													},
												},
											},
										},
										"subnet": schema.ListAttribute{
											ElementType: types.StringType,
											Description: "Subnet traffic matching criteria. Logical 'OR' is applied within the criteria set. Logical 'AND' is applied between criteria sets.",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.List{
												listplanmodifier.UseStateForUnknown(),
											},
											Validators: []validator.List{
												listvalidator.SizeAtLeast(1),
											},
										},
										"ip_range": schema.ListNestedAttribute{
											Description: "IP range traffic matching criteria. Logical 'OR' is applied within the criteria set. Logical 'AND' is applied between criteria sets.",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.List{
												listplanmodifier.UseStateForUnknown(),
											},
											Validators: []validator.List{
												listvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"from": schema.StringAttribute{
														Description: "From IP Range",
														Required:    true,
														Optional:    false,
													},
													"to": schema.StringAttribute{
														Description: "To IP Range",
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
												setplanmodifier.UseStateForUnknown(),
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
															stringplanmodifier.UseStateForUnknown(),
														},
														Computed: true,
													},
													"id": schema.StringAttribute{
														Description: "Global IP Range ID",
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
										"network_interface": schema.SetNestedAttribute{
											Description: "Network range defined for a site",
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
													"name": schema.StringAttribute{
														Description: "Network Interface Name",
														Required:    false,
														Optional:    true,
														PlanModifiers: []planmodifier.String{
															stringplanmodifier.UseStateForUnknown(),
														},
														Computed: true,
													},
													"id": schema.StringAttribute{
														Description: "Network Interface ID",
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
										"site_network_subnet": schema.SetNestedAttribute{
											Required: false,
											Optional: true,
											PlanModifiers: []planmodifier.Set{
												setplanmodifier.UseStateForUnknown(),
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
															stringplanmodifier.UseStateForUnknown(),
														},
														Computed: true,
													},
													"id": schema.StringAttribute{
														Description: "Site Network Subnet ID",
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
										"floating_subnet": schema.SetNestedAttribute{
											Description: "Floating Subnets (ie. Floating Ranges) are used to identify traffic exactly matched to the route advertised by BGP. They are not associated with a specific site. This is useful in scenarios such as active-standby high availability routed via BGP.",
											Required:    false,
											Optional:    true,
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											PlanModifiers: []planmodifier.Set{
												setplanmodifier.UseStateForUnknown(),
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
												setplanmodifier.UseStateForUnknown(),
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
															stringplanmodifier.UseStateForUnknown(),
														},
														Computed: true,
													},
													"id": schema.StringAttribute{
														Description: "User ID",
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
										"users_group": schema.SetNestedAttribute{
											Description: "Group of users",
											Required:    false,
											Optional:    true,
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											PlanModifiers: []planmodifier.Set{
												setplanmodifier.UseStateForUnknown(),
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
															stringplanmodifier.UseStateForUnknown(),
														},
														Computed: true,
													},
													"id": schema.StringAttribute{
														Description: "Users Group ID",
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
										"group": schema.SetNestedAttribute{
											Description: "Groups defined for your account",
											Required:    false,
											Optional:    true,
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											PlanModifiers: []planmodifier.Set{
												setplanmodifier.UseStateForUnknown(),
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
															stringplanmodifier.UseStateForUnknown(),
														},
														Computed: true,
													},
													"id": schema.StringAttribute{
														Description: "Group ID",
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
										"system_group": schema.SetNestedAttribute{
											Description: "Predefined Cato groups",
											Required:    false,
											Optional:    true,
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											PlanModifiers: []planmodifier.Set{
												setplanmodifier.UseStateForUnknown(),
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
															stringplanmodifier.UseStateForUnknown(),
														},
														Computed: true,
													},
													"id": schema.StringAttribute{
														Description: "System Group ID",
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
									},
								},
								"destination": schema.SingleNestedAttribute{
									Description: "Destination traffic matching criteria for the exception.",
									Required:    false,
									Optional:    true,
									Attributes: map[string]schema.Attribute{
										"ip": schema.ListAttribute{
											Description: "IP traffic matching criteria. Logical 'OR' is applied within the criteria set. Logical 'AND' is applied between criteria sets.",
											ElementType: types.StringType,
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.List{
												listplanmodifier.UseStateForUnknown(),
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
												setplanmodifier.UseStateForUnknown(),
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
															stringplanmodifier.UseStateForUnknown(),
														},
														Computed: true,
													},
													"id": schema.StringAttribute{
														Description: "Host ID",
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
										"site": schema.SetNestedAttribute{
											Description: "Sites defined in your account",
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
															stringplanmodifier.UseStateForUnknown(),
														},
														Computed: true,
													},
													"id": schema.StringAttribute{
														Description: "Site ID",
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
										"subnet": schema.ListAttribute{
											ElementType: types.StringType,
											Description: "Subnet traffic matching criteria. Logical 'OR' is applied within the criteria set. Logical 'AND' is applied between criteria sets.",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.List{
												listplanmodifier.UseStateForUnknown(),
											},
											Validators: []validator.List{
												listvalidator.SizeAtLeast(1),
											},
										},
										"ip_range": schema.ListNestedAttribute{
											Description: "IP range traffic matching criteria. Logical 'OR' is applied within the criteria set. Logical 'AND' is applied between criteria sets.",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.List{
												listplanmodifier.UseStateForUnknown(),
											},
											Validators: []validator.List{
												listvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"from": schema.StringAttribute{
														Description: "From IP Range",
														Required:    true,
														Optional:    false,
													},
													"to": schema.StringAttribute{
														Description: "To IP Range",
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
												setplanmodifier.UseStateForUnknown(),
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
															stringplanmodifier.UseStateForUnknown(),
														},
														Computed: true,
													},
													"id": schema.StringAttribute{
														Description: "Global IP Range ID",
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
										"network_interface": schema.SetNestedAttribute{
											Description: "Network range defined for a site",
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
													"name": schema.StringAttribute{
														Description: "Network Interface Name",
														Required:    false,
														Optional:    true,
														PlanModifiers: []planmodifier.String{
															stringplanmodifier.UseStateForUnknown(),
														},
														Computed: true,
													},
													"id": schema.StringAttribute{
														Description: "Network Interface ID",
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
										"site_network_subnet": schema.SetNestedAttribute{
											Required: false,
											Optional: true,
											PlanModifiers: []planmodifier.Set{
												setplanmodifier.UseStateForUnknown(),
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
															stringplanmodifier.UseStateForUnknown(),
														},
														Computed: true,
													},
													"id": schema.StringAttribute{
														Description: "Site Network Subnet ID",
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
										"floating_subnet": schema.SetNestedAttribute{
											Description: "Floating Subnets (ie. Floating Ranges) are used to identify traffic exactly matched to the route advertised by BGP. They are not associated with a specific site. This is useful in scenarios such as active-standby high availability routed via BGP.",
											Required:    false,
											Optional:    true,
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											PlanModifiers: []planmodifier.Set{
												setplanmodifier.UseStateForUnknown(),
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
												setplanmodifier.UseStateForUnknown(),
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
															stringplanmodifier.UseStateForUnknown(),
														},
														Computed: true,
													},
													"id": schema.StringAttribute{
														Description: "User ID",
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
										"users_group": schema.SetNestedAttribute{
											Description: "Group of users",
											Required:    false,
											Optional:    true,
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											PlanModifiers: []planmodifier.Set{
												setplanmodifier.UseStateForUnknown(),
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
															stringplanmodifier.UseStateForUnknown(),
														},
														Computed: true,
													},
													"id": schema.StringAttribute{
														Description: "Users Group ID",
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
										"group": schema.SetNestedAttribute{
											Description: "Groups defined for your account",
											Required:    false,
											Optional:    true,
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											PlanModifiers: []planmodifier.Set{
												setplanmodifier.UseStateForUnknown(),
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
															stringplanmodifier.UseStateForUnknown(),
														},
														Computed: true,
													},
													"id": schema.StringAttribute{
														Description: "Group ID",
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
										"system_group": schema.SetNestedAttribute{
											Description: "Predefined Cato groups",
											Required:    false,
											Optional:    true,
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											PlanModifiers: []planmodifier.Set{
												setplanmodifier.UseStateForUnknown(),
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
															stringplanmodifier.UseStateForUnknown(),
														},
														Computed: true,
													},
													"id": schema.StringAttribute{
														Description: "System Group ID",
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
									},
								},
								"application": schema.SingleNestedAttribute{
									Description: "Application matching criteria for the exception.",
									Optional:    true,
									Required:    false,
									PlanModifiers: []planmodifier.Object{
										objectplanmodifier.UseStateForUnknown(),
									},
									Attributes: map[string]schema.Attribute{
										"application": schema.SetNestedAttribute{
											Description: "Application defined for your account",
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
															stringplanmodifier.UseStateForUnknown(),
														},
														Computed: true,
													},
													"id": schema.StringAttribute{
														Description: "Application ID",
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
										"custom_app": schema.SetNestedAttribute{
											Description: "Custom Applications",
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
															stringplanmodifier.UseStateForUnknown(),
														},
														Computed: true,
													},
													"id": schema.StringAttribute{
														Description: "Custom Application ID",
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
										"app_category": schema.SetNestedAttribute{
											Description: "Application Categories",
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
															stringplanmodifier.UseStateForUnknown(),
														},
														Computed: true,
													},
													"id": schema.StringAttribute{
														Description: "Application Category ID",
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
										"custom_category": schema.SetNestedAttribute{
											Description: "Custom Application Categories",
											Required:    false,
											Optional:    true,
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											PlanModifiers: []planmodifier.Set{
												setplanmodifier.UseStateForUnknown(),
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
															stringplanmodifier.UseStateForUnknown(),
														},
														Computed: true,
													},
													"id": schema.StringAttribute{
														Description: "Custom Application Category ID",
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
										"domain": schema.ListAttribute{
											ElementType: types.StringType,
											Description: "Domain names matching criteria for the exception.",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.List{
												listplanmodifier.UseStateForUnknown(),
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
												listplanmodifier.UseStateForUnknown(),
											},
											Validators: []validator.List{
												listvalidator.SizeAtLeast(1),
											},
										},
										"service": schema.SetNestedAttribute{
											Description: "Services",
											Optional:    true,
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											PlanModifiers: []planmodifier.Set{
												setplanmodifier.UseStateForUnknown(),
											},
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"id": schema.StringAttribute{
														Description: "Service ID",
														Optional:    true,
														Computed:    true,
														PlanModifiers: []planmodifier.String{
															stringplanmodifier.UseStateForUnknown(),
														},
													},
													"name": schema.StringAttribute{
														Description: "Service Name",
														Optional:    true,
														Computed:    true,
														Validators: []validator.String{
															stringvalidator.ConflictsWith(path.Expressions{
																path.MatchRelative().AtParent().AtName("id"),
															}...),
														},
														PlanModifiers: []planmodifier.String{
															stringplanmodifier.UseStateForUnknown(),
														},
													},
												},
											},
										},
										"custom_service": schema.ListNestedAttribute{
											Description: "Custom services",
											Optional:    true,
											Validators: []validator.List{
												listvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"port": schema.ListAttribute{
														Description: "Port numbers",
														ElementType: types.StringType,
														Optional:    true,
													},
													"port_range": schema.SingleNestedAttribute{
														Description: "Port range",
														Optional:    true,
														Attributes: map[string]schema.Attribute{
															"from": schema.StringAttribute{
																Description: "Start port",
																Optional:    true,
															},
															"to": schema.StringAttribute{
																Description: "End port",
																Optional:    true,
															},
														},
													},
													"protocol": schema.StringAttribute{
														Description: "Protocol",
														Optional:    true,
														Validators: []validator.String{
															stringvalidator.OneOf("ANY", "ICMP", "TCP", "TCP_UDP", "UDP"),
														},
													},
												},
											},
										},
										"custom_service_ip": schema.ListNestedAttribute{
											Description: "Custom service IPs",
											Optional:    true,
											Validators: []validator.List{
												listvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"name": schema.StringAttribute{
														Description: "Name",
														Optional:    true,
													},
													"ip": schema.StringAttribute{
														Description: "IP address",
														Optional:    true,
													},
													"ip_range": schema.SingleNestedAttribute{
														Description: "IP range",
														Optional:    true,
														Attributes: map[string]schema.Attribute{
															"from": schema.StringAttribute{
																Description: "Start IP",
																Optional:    true,
															},
															"to": schema.StringAttribute{
																Description: "End IP",
																Optional:    true,
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
					},
					"configuration": schema.SingleNestedAttribute{
						Description: "WAN Network configuration",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"active_tcp_acceleration": schema.BoolAttribute{
								Description: "Enable active TCP acceleration",
								Optional:    true,
							},
							"packet_loss_mitigation": schema.BoolAttribute{
								Description: "Enable packet loss mitigation",
								Optional:    true,
							},
							"preserve_source_port": schema.BoolAttribute{
								Description: "Preserve source port",
								Optional:    true,
							},
							"primary_transport": schema.SingleNestedAttribute{
								Description: "Primary transport configuration",
								Optional:    true,
								Attributes: map[string]schema.Attribute{
									"transport_type": schema.StringAttribute{
										Description: "Transport type",
										Optional:    true,
										Validators: []validator.String{
											stringvalidator.OneOf("ALTERNATIVE_WAN", "AUTOMATIC", "NONE", "OFF_CLOUD", "WAN"),
										},
									},
									"primary_interface_role": schema.StringAttribute{
										Description: "Primary interface role",
										Optional:    true,
										Validators: []validator.String{
											stringvalidator.OneOf("AUTOMATIC", "NONE", "WAN1", "WAN2", "WAN3", "WAN4", "WAN5", "WAN6"),
										},
									},
									"secondary_interface_role": schema.StringAttribute{
										Description: "Secondary interface role",
										Optional:    true,
										Validators: []validator.String{
											stringvalidator.OneOf("AUTOMATIC", "NONE", "WAN1", "WAN2", "WAN3", "WAN4", "WAN5", "WAN6"),
										},
									},
								},
							},
							"secondary_transport": schema.SingleNestedAttribute{
								Description: "Secondary transport configuration",
								Optional:    true,
								Attributes: map[string]schema.Attribute{
									"transport_type": schema.StringAttribute{
										Description: "Transport type",
										Optional:    true,
										Validators: []validator.String{
											stringvalidator.OneOf("ALTERNATIVE_WAN", "AUTOMATIC", "NONE", "OFF_CLOUD", "WAN"),
										},
									},
									"primary_interface_role": schema.StringAttribute{
										Description: "Primary interface role",
										Optional:    true,
										Validators: []validator.String{
											stringvalidator.OneOf("AUTOMATIC", "NONE", "WAN1", "WAN2", "WAN3", "WAN4", "WAN5", "WAN6"),
										},
									},
									"secondary_interface_role": schema.StringAttribute{
										Description: "Secondary interface role",
										Optional:    true,
										Validators: []validator.String{
											stringvalidator.OneOf("AUTOMATIC", "NONE", "WAN1", "WAN2", "WAN3", "WAN4", "WAN5", "WAN6"),
										},
									},
								},
							},
							"allocation_ip": schema.SetNestedAttribute{
								Description: "Allocation IPs",
								Optional:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "Allocation IP ID",
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"name": schema.StringAttribute{
											Description: "Allocation IP Name",
											Optional:    true,
											Computed:    true,
											Validators: []validator.String{
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
							"pop_location": schema.SetNestedAttribute{
								Description: "PoP locations",
								Optional:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "PoP Location ID",
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"name": schema.StringAttribute{
											Description: "PoP Location Name",
											Optional:    true,
											Computed:    true,
											Validators: []validator.String{
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
							"backhauling_site": schema.SetNestedAttribute{
								Description: "Backhauling sites",
								Optional:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								PlanModifiers: []planmodifier.Set{
									setplanmodifier.UseStateForUnknown(),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Description: "Backhauling Site ID",
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"name": schema.StringAttribute{
											Description: "Backhauling Site Name",
											Optional:    true,
											Computed:    true,
											Validators: []validator.String{
												stringvalidator.ConflictsWith(path.Expressions{
													path.MatchRelative().AtParent().AtName("id"),
												}...),
											},
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
						},
					},
					"bandwidth_priority": schema.SingleNestedAttribute{
						Description: "Bandwidth priority",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"id": schema.StringAttribute{
								Description: "Bandwidth Priority ID",
								Optional:    true,
								Computed:    true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
							"name": schema.StringAttribute{
								Description: "Bandwidth Priority Name",
								Optional:    true,
								Computed:    true,
								Validators: []validator.String{
									stringvalidator.ConflictsWith(path.Expressions{
										path.MatchRelative().AtParent().AtName("id"),
									}...),
								},
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *wanNetworkRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*catoClientData)
}

func (r *wanNetworkRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("rule").AtName("id"), req, resp)
}

func (r *wanNetworkRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WanNetworkRule
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := hydrateWanNetworkRuleApi(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Warn(ctx, "TFLOG_WARN_WAN_input.create", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(input.create),
	})

	createRuleResponse, err := r.client.catov2.PolicyWanNetworkAddRule(ctx, input.create, r.client.AccountId)

	tflog.Warn(ctx, "TFLOG_WARN_WAN_createRuleResponse", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(createRuleResponse),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanNetworkAddRule error",
			err.Error(),
		)
		return
	}

	// check for errors
	if createRuleResponse.Policy.WanNetwork.AddRule.Status != "SUCCESS" {
		for _, item := range createRuleResponse.Policy.WanNetwork.AddRule.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Creating Resource",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	// Publish policy revision (align with WAN FW behavior)
	tflog.Info(ctx, "publishing new rule")
	_, err = r.client.catov2.PolicyWanNetworkPublishPolicyRevision(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanNetworkPublishPolicyRevision error",
			err.Error(),
		)
		return
	}

	// Read rule and hydrate response to state
	body, err := r.client.catov2.WanNetworkPolicy(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API WanNetworkPolicy error",
			err.Error(),
		)
		return
	}

	ruleList := body.GetPolicy().WanNetwork.Policy.Rules
	currentRule := &cato_go_sdk.WanNetworkPolicy_Policy_WanNetwork_Policy_Rules_Rule{}
	// Get current rule from response by ID
	if createRuleResponse.Policy.WanNetwork.AddRule.Rule != nil {
		createdRuleID := createRuleResponse.Policy.WanNetwork.AddRule.Rule.GetRule().ID
		for _, ruleListItem := range ruleList {
			if ruleListItem.GetRule().ID == createdRuleID {
				currentRule = ruleListItem.GetRule()
				resp.State.SetAttribute(
					ctx,
					path.Root("rule").AtName("id"),
					ruleListItem.GetRule().ID)
				break
			}
		}
	}
	tflog.Warn(ctx, "TFLOG_WARN_WAN_createRule.readResponse", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(currentRule),
	})
	// Hydrate ruleInput from api response
	ruleInputRead, hydrateDiags := hydrateWanNetworkRuleState(ctx, plan, currentRule)
	resp.Diagnostics.Append(hydrateDiags...)
	ruleInputRead.ID = types.StringValue(currentRule.ID)
	tflog.Info(ctx, "ruleInputRead - "+fmt.Sprintf("%v", ruleInputRead))

	// Handle exceptions correlation manually to preserve plan structure
	if !plan.Rule.IsNull() && !plan.Rule.IsUnknown() {
		planRule := Policy_Policy_WanNetwork_Policy_Rules_Rule{}
		diags = plan.Rule.As(ctx, &planRule, basetypes.ObjectAsOptions{})
		if !diags.HasError() && !planRule.Exceptions.IsNull() && !planRule.Exceptions.IsUnknown() {
			// Correlate exceptions between plan and hydrated response
			correlatedExceptions := correlateWanExceptions(ctx, planRule.Exceptions, ruleInputRead.Exceptions)
			if correlatedExceptions != nil {
				ruleInputRead.Exceptions = *correlatedExceptions
			}
		}
	}

	ruleObject, diags := types.ObjectValueFrom(ctx, WanNetworkRuleRuleAttrTypes, ruleInputRead)
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
}

func (r *wanNetworkRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WanNetworkRule
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Query WAN Network policy
	body, err := r.client.catov2.WanNetworkPolicy(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API WanNetworkPolicy error",
			err.Error(),
		)
		return
	}

	// Extract rule ID from state
	var ruleID string
	if !state.Rule.IsNull() && !state.Rule.IsUnknown() {
		rule := Policy_Policy_WanNetwork_Policy_Rules_Rule{}
		diags = state.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			// If conversion fails, try to extract just the ID from the state
			tflog.Debug(ctx, "Failed to convert full state to struct, attempting to extract ID only")
			// Clear the diagnostics as we'll handle this differently
			resp.Diagnostics = resp.Diagnostics[:0]

			// Try to extract ID directly from the state
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

	// Find the rule in the policy
	ruleList := body.GetPolicy().WanNetwork.Policy.Rules
	ruleExist := false
	var currentRule *cato_go_sdk.WanNetworkPolicy_Policy_WanNetwork_Policy_Rules_Rule
	for _, ruleListItem := range ruleList {
		rule := ruleListItem.GetRule()
		if rule.ID == ruleID {
			ruleExist = true
			currentRule = rule
			break
		}
	}

	// Remove resource if it doesn't exist anymore
	if !ruleExist {
		tflog.Warn(ctx, "WAN Network rule not found, resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	tflog.Debug(ctx, "WAN Network rule found, hydrating state", map[string]interface{}{
		"rule_id": ruleID,
	})

	// Hydrate the state from the API response
	ruleInput, hydrateDiags := hydrateWanNetworkRuleState(ctx, state, currentRule)
	resp.Diagnostics.Append(hydrateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Handle exceptions correlation manually to preserve state structure
	if !state.Rule.IsNull() && !state.Rule.IsUnknown() {
		stateRule := Policy_Policy_WanNetwork_Policy_Rules_Rule{}
		diags = state.Rule.As(ctx, &stateRule, basetypes.ObjectAsOptions{})
		if !diags.HasError() && !stateRule.Exceptions.IsNull() && !stateRule.Exceptions.IsUnknown() {
			// Correlate exceptions between state and hydrated response
			correlatedExceptions := correlateWanExceptions(ctx, stateRule.Exceptions, ruleInput.Exceptions)
			if correlatedExceptions != nil {
				ruleInput.Exceptions = *correlatedExceptions
			}
		}
	}

	// Set the rule attribute in state
	diags = resp.State.SetAttribute(ctx, path.Root("rule"), ruleInput)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Hard code LAST_IN_POLICY position as the API does not return any value
	// This supports the use case of bulk rule import/export
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
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(diagstmp...)
}

func (r *wanNetworkRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan WanNetworkRule
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := hydrateWanNetworkRuleApi(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract rule ID from plan
	var ruleID string
	if !plan.Rule.IsNull() && !plan.Rule.IsUnknown() {
		rule := Policy_Policy_WanNetwork_Policy_Rules_Rule{}
		diags = plan.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		ruleID = rule.ID.ValueString()
	}

	// Prepare update input with rule ID
	updateInput := cato_models.WanNetworkUpdateRuleInput{
		ID:   ruleID,
		Rule: input.update.Rule,
	}

	tflog.Warn(ctx, "TFLOG_WARN_WAN_NW_input.update", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(updateInput),
	})

	updateRuleResponse, err := r.client.catov2.PolicyWanNetworkUpdateRule(ctx, updateInput, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanNetworkUpdateRule error",
			err.Error(),
		)
		return
	}

	tflog.Warn(ctx, "TFLOG_WARN_WAN_NW_input.update.response", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(updateRuleResponse),
	})

	if updateRuleResponse.Policy.WanNetwork.UpdateRule.Status != "SUCCESS" {
		for _, item := range updateRuleResponse.Policy.WanNetwork.UpdateRule.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Updating Resource",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	// Publish policy revision after update
	tflog.Info(ctx, "publishing updated rule")
	_, err = r.client.catov2.PolicyWanNetworkPublishPolicyRevision(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanNetworkPublishPolicyRevision error",
			err.Error(),
		)
		return
	}

	// Read rule and hydrate response to state
	body, err := r.client.catov2.WanNetworkPolicy(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API WanNetworkPolicy error",
			err.Error(),
		)
		return
	}

	ruleList := body.GetPolicy().WanNetwork.Policy.Rules
	currentRule := &cato_go_sdk.WanNetworkPolicy_Policy_WanNetwork_Policy_Rules_Rule{}
	// Get current rule from response by ID
	for _, ruleListItem := range ruleList {
		if ruleListItem.GetRule().ID == updateRuleResponse.GetPolicy().GetWanNetwork().GetUpdateRule().Rule.GetRule().ID {
			currentRule = ruleListItem.GetRule()
			break
		}
	}
	tflog.Warn(ctx, "TFLOG_WARN_WAN_updateRule.readResponse", map[string]interface{}{
		"OUTPUT": utils.InterfaceToJSONString(currentRule),
	})

	// Hydrate ruleInput from api response
	ruleInputRead, hydrateDiags := hydrateWanNetworkRuleState(ctx, plan, currentRule)
	resp.Diagnostics.Append(hydrateDiags...)
	ruleInputRead.ID = types.StringValue(currentRule.ID)

	// Handle exceptions correlation manually to preserve plan structure
	if !plan.Rule.IsNull() && !plan.Rule.IsUnknown() {
		planRule := Policy_Policy_WanNetwork_Policy_Rules_Rule{}
		diags = plan.Rule.As(ctx, &planRule, basetypes.ObjectAsOptions{})
		if !diags.HasError() && !planRule.Exceptions.IsNull() && !planRule.Exceptions.IsUnknown() {
			// Correlate exceptions between plan and hydrated response
			correlatedExceptions := correlateWanExceptions(ctx, planRule.Exceptions, ruleInputRead.Exceptions)
			if correlatedExceptions != nil {
				ruleInputRead.Exceptions = *correlatedExceptions
			}
		}
	}

	ruleObject, diags := types.ObjectValueFrom(ctx, WanNetworkRuleRuleAttrTypes, ruleInputRead)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Rule = ruleObject

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *wanNetworkRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WanNetworkRule
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract rule ID from state
	var ruleID string
	if !state.Rule.IsNull() && !state.Rule.IsUnknown() {
		rule := Policy_Policy_WanNetwork_Policy_Rules_Rule{}
		diags = state.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		ruleID = rule.ID.ValueString()
	}

	// Prepare remove input
	removeInput := cato_models.WanNetworkRemoveRuleInput{
		ID: ruleID,
	}

	deleteRuleResponse, err := r.client.catov2.PolicyWanNetworkRemoveRule(ctx, removeInput, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanNetworkRemoveRule error",
			err.Error(),
		)
		return
	}

	if deleteRuleResponse.Policy.WanNetwork.RemoveRule.Status != "SUCCESS" {
		for _, item := range deleteRuleResponse.Policy.WanNetwork.RemoveRule.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Deleting Resource",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	// Publish policy revision after update
	tflog.Info(ctx, "publishing updated rule")
	_, err = r.client.catov2.PolicyWanNetworkPublishPolicyRevision(ctx, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyWanNetworkPublishPolicyRevision error",
			err.Error(),
		)
		return
	}
}
