//nolint:lll
package wnwrule

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/planmodifiers"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
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
							stringvalidator.OneOf(utils.ToStrings(cato_models.AllWanNetworkRuleType)...),
						},
					},
					"route_type": schema.StringAttribute{
						Description: "Routing method for the rule",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf(utils.ToStrings(cato_models.AllWanNetworkRuleRouteType)...),
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
								Description:   "Hosts defined for your account",
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("host")},
							},
							"site": schema.SetNestedAttribute{
								Description:   "Sites defined for the account",
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("site")},
							},
							"global_ip_range": schema.SetNestedAttribute{
								Description:   "Global IP ranges",
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("global_ip_range")},
							},
							"network_interface": schema.SetNestedAttribute{
								Description:   "Network interfaces",
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("network_interface")},
							},
							"site_network_subnet": schema.SetNestedAttribute{
								Description:   "Site network subnets",
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("site_network_subnet")},
							},
							"floating_subnet": schema.SetNestedAttribute{
								Description:   "Floating subnets",
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("floating_subnet")},
							},
							"user": schema.SetNestedAttribute{
								Description:   "Individual users",
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("user")},
							},
							"users_group": schema.SetNestedAttribute{
								Description:   "User groups",
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("users_group")},
							},
							"group": schema.SetNestedAttribute{
								Description:   "Groups",
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("group")},
							},
							"system_group": schema.SetNestedAttribute{
								Description:   "System groups",
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("system_group")},
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
								Description:   "Hosts defined for your account",
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("host")},
							},
							"site": schema.SetNestedAttribute{
								Description:   "Sites defined for the account",
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("site")},
							},
							"global_ip_range": schema.SetNestedAttribute{
								Description:   "Global IP ranges",
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("global_ip_range")},
							},
							"network_interface": schema.SetNestedAttribute{
								Description:   "Network interfaces",
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("network_interface")},
							},
							"site_network_subnet": schema.SetNestedAttribute{
								Description:   "Site network subnets",
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("site_network_subnet")},
							},
							"floating_subnet": schema.SetNestedAttribute{
								Description:   "Floating subnets",
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("floating_subnet")},
							},
							"user": schema.SetNestedAttribute{
								Description:   "Individual users",
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("user")},
							},
							"users_group": schema.SetNestedAttribute{
								Description:   "User groups",
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("users_group")},
							},
							"group": schema.SetNestedAttribute{
								Description:   "Groups",
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("group")},
							},
							"system_group": schema.SetNestedAttribute{
								Description:   "System groups",
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("system_group")},
							},
						},
					},
					"application": schema.SingleNestedAttribute{
						Description: "Application matching criteria",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"application": schema.SetNestedAttribute{
								Description:   "Applications",
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("application")},
							},
							"custom_app": schema.SetNestedAttribute{
								Description:   "Custom applications",
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("custom_app")},
							},
							"app_category": schema.SetNestedAttribute{
								Description:   "Application categories",
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("app_category")},
							},
							"custom_category": schema.SetNestedAttribute{
								Description:   "Custom categories",
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("custom_category")},
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
								Description:   "Services",
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("service")},
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
												stringvalidator.OneOf(utils.ToStrings(cato_models.AllIPProtocol)...),
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
											Description:   "Hosts and servers defined for your account",
											Required:      false,
											Optional:      true,
											Computed:      true,
											PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("host")},
										},
										"site": schema.SetNestedAttribute{
											Description:   "Sites defined in your account",
											Required:      false,
											Optional:      true,
											Computed:      true,
											PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("site")},
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
											Description:   "Global IP Range",
											Required:      false,
											Optional:      true,
											Computed:      true,
											PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("global_ip_range")},
										},
										"network_interface": schema.SetNestedAttribute{
											Description:   "Network range defined for a site",
											Required:      false,
											Optional:      true,
											Computed:      true,
											PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("network_interface")},
										},
										"site_network_subnet": schema.SetNestedAttribute{
											Required:      false,
											Optional:      true,
											Computed:      true,
											PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("site_network_subnet")},
										},
										"floating_subnet": schema.SetNestedAttribute{
											Description:   "Floating Subnets (ie. Floating Ranges) are used to identify traffic exactly matched to the route advertised by BGP. They are not associated with a specific site. This is useful in scenarios such as active-standby high availability routed via BGP.",
											Required:      false,
											Optional:      true,
											Computed:      true,
											PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("floating_subnet")},
										},
										"user": schema.SetNestedAttribute{
											Description:   "Individual users defined for the account",
											Required:      false,
											Optional:      true,
											Computed:      true,
											PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("user")},
										},
										"users_group": schema.SetNestedAttribute{
											Description:   "Group of users",
											Required:      false,
											Optional:      true,
											Computed:      true,
											PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("users_group")},
										},
										"group": schema.SetNestedAttribute{
											Description:   "Groups defined for your account",
											Required:      false,
											Optional:      true,
											Computed:      true,
											PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("group")},
										},
										"system_group": schema.SetNestedAttribute{
											Description:   "Predefined Cato groups",
											Required:      false,
											Optional:      true,
											Computed:      true,
											PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("system_group")},
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
											Description:   "Hosts and servers defined for your account",
											Required:      false,
											Optional:      true,
											Computed:      true,
											PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("host")},
										},
										"site": schema.SetNestedAttribute{
											Description:   "Sites defined in your account",
											Required:      false,
											Optional:      true,
											Computed:      true,
											PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("site")},
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
											Description:   "Global IP Range",
											Required:      false,
											Optional:      true,
											Computed:      true,
											PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("global_ip_range")},
										},
										"network_interface": schema.SetNestedAttribute{
											Description:   "Network range defined for a site",
											Required:      false,
											Optional:      true,
											Computed:      true,
											PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("network_interface")},
										},
										"site_network_subnet": schema.SetNestedAttribute{
											Required:      false,
											Optional:      true,
											Computed:      true,
											PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("site_network_subnet")},
										},
										"floating_subnet": schema.SetNestedAttribute{
											Description:   "Floating Subnets (ie. Floating Ranges) are used to identify traffic exactly matched to the route advertised by BGP. They are not associated with a specific site. This is useful in scenarios such as active-standby high availability routed via BGP.",
											Required:      false,
											Optional:      true,
											Computed:      true,
											PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("floating_subnet")},
										},
										"user": schema.SetNestedAttribute{
											Description:   "Individual users defined for the account",
											Required:      false,
											Optional:      true,
											Computed:      true,
											PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("user")},
										},
										"users_group": schema.SetNestedAttribute{
											Description:   "Group of users",
											Required:      false,
											Optional:      true,
											Computed:      true,
											PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("users_group")},
										},
										"group": schema.SetNestedAttribute{
											Description:   "Groups defined for your account",
											Required:      false,
											Optional:      true,
											Computed:      true,
											PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("group")},
										},
										"system_group": schema.SetNestedAttribute{
											Description:   "Predefined Cato groups",
											Required:      false,
											Optional:      true,
											Computed:      true,
											PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("system_group")},
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
											Description:   "Application defined for your account",
											Required:      false,
											Optional:      true,
											Computed:      true,
											PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("application")},
										},
										"custom_app": schema.SetNestedAttribute{
											Description:   "Custom Applications",
											Required:      false,
											Optional:      true,
											Computed:      true,
											PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("custom_app")},
										},
										"app_category": schema.SetNestedAttribute{
											Description:   "Application Categories",
											Required:      false,
											Optional:      true,
											Computed:      true,
											PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("app_category")},
										},
										"custom_category": schema.SetNestedAttribute{
											Description:   "Custom Application Categories",
											Required:      false,
											Optional:      true,
											Computed:      true,
											PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("custom_category")},
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
											Description:   "Services",
											Optional:      true,
											Computed:      true,
											PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("service")},
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
															stringvalidator.OneOf(utils.ToStrings(cato_models.AllIPProtocol)...),
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
											stringvalidator.OneOf(utils.ToStrings(cato_models.AllWanNetworkRuleTransportType)...),
										},
									},
									"primary_interface_role": schema.StringAttribute{
										Description: "Primary interface role",
										Optional:    true,
										Validators: []validator.String{
											stringvalidator.OneOf(utils.ToStrings(cato_models.AllWanNetworkRuleInterfaceRole)...),
										},
									},
									"secondary_interface_role": schema.StringAttribute{
										Description: "Secondary interface role",
										Optional:    true,
										Validators: []validator.String{
											stringvalidator.OneOf(utils.ToStrings(cato_models.AllWanNetworkRuleInterfaceRole)...),
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
											stringvalidator.OneOf(utils.ToStrings(cato_models.AllWanNetworkRuleTransportType)...),
										},
									},
									"primary_interface_role": schema.StringAttribute{
										Description: "Primary interface role",
										Optional:    true,
										Validators: []validator.String{
											stringvalidator.OneOf(utils.ToStrings(cato_models.AllWanNetworkRuleInterfaceRole)...),
										},
									},
									"secondary_interface_role": schema.StringAttribute{
										Description: "Secondary interface role",
										Optional:    true,
										Validators: []validator.String{
											stringvalidator.OneOf(utils.ToStrings(cato_models.AllWanNetworkRuleInterfaceRole)...),
										},
									},
								},
							},
							"allocation_ip": schema.SetNestedAttribute{
								Description:   "Allocation IPs",
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("allocation_ip")},
							},
							"pop_location": schema.SetNestedAttribute{
								Description:   "PoP locations",
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("pop_location")},
							},
							"backhauling_site": schema.SetNestedAttribute{
								Description:   "Backhauling sites",
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("backhauling_site")},
							},
						},
					},
					"bandwidth_priority": schema.SingleNestedAttribute{
						Description:   "Bandwidth priority",
						Optional:      true,
						PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
						Attributes:    idname.SchemaNameID("Bandwidth priority"),
					},
				},
			},
		},
	}
}
