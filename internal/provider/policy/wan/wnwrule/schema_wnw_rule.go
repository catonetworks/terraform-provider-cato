//nolint:lll
package wnwrule

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/planmodifiers"
)

//nolint:funlen
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
