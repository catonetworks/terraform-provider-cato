package socketlanfirewallrule

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

//nolint:funlen
func (r *socketLanFirewallRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_socket_lan_firewall_rule` resource contains the configuration parameters necessary " +
			"to add a Socket LAN firewall rule (child rule under a network rule). Documentation for the underlying " +
			"API used in this resource can be found at [mutation.policy.socketLan.firewall.addRule()]" +
			"(https://api.catonetworks.com/documentation/#mutation-policy.socketLan.firewall.addRule).",
		Attributes: map[string]schema.Attribute{
			"at": schema.SingleNestedAttribute{
				Description: "Position of the rule relative to the parent network rule",
				Required:    true,
				Optional:    false,
				Attributes: map[string]schema.Attribute{
					"position": schema.StringAttribute{
						Description: "Position relative to the parent rule (FIRST_IN_RULE, LAST_IN_RULE, AFTER_RULE, BEFORE_RULE)",
						Required:    true,
						Optional:    false,
						Validators: []validator.String{
							stringvalidator.OneOf("FIRST_IN_RULE", "LAST_IN_RULE", "AFTER_RULE", "BEFORE_RULE"),
						},
					},
					"ref": schema.StringAttribute{
						Description: "The parent network rule ID or sibling firewall rule ID",
						Required:    true,
						Optional:    false,
					},
				},
			},
			"rule": schema.SingleNestedAttribute{
				Description: "Parameters for the Socket LAN firewall rule",
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
					"action": schema.StringAttribute{
						Description: "Action (ALLOW, BLOCK)",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("ALLOW", "BLOCK"),
						},
					},
					"source": schema.SingleNestedAttribute{
						Description: "Source traffic matching criteria",
						Required:    true,
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
						},
						Attributes: map[string]schema.Attribute{
							"vlan": schema.ListAttribute{
								ElementType: types.Int64Type,
								Description: "VLAN IDs",
								Required:    false,
								Optional:    true,
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
							"ip": schema.ListAttribute{
								ElementType: types.StringType,
								Description: "IP addresses",
								Required:    false,
								Optional:    true,
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
							"subnet": schema.ListAttribute{
								ElementType: types.StringType,
								Description: "Subnets in CIDR notation",
								Required:    false,
								Optional:    true,
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
							"ip_range": schema.ListNestedAttribute{
								Description: "IP ranges",
								Required:    false,
								Optional:    true,
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
								Description:   "Hosts",
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
								Description:   "Sites",
								Required:      false,
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("site")},
							},
							"group": schema.SetNestedAttribute{
								Description:   "Groups",
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
								Description:   "System groups",
								Required:      false,
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("system_group")},
							},
							"network_interface": schema.SetNestedAttribute{
								Description:   "Network interfaces",
								Required:      false,
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("network_interface")},
							},
							"global_ip_range": schema.SetNestedAttribute{
								Description:   "Global IP ranges",
								Required:      false,
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("global_ip_range")},
							},
							"floating_subnet": schema.SetNestedAttribute{
								Description:   "Floating subnets",
								Required:      false,
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("floating_subnet")},
							},
							"site_network_subnet": schema.SetNestedAttribute{
								Description:   "Site network subnets",
								Required:      false,
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("site_network_subnet")},
							},
							"mac": schema.ListAttribute{
								ElementType: types.StringType,
								Description: "MAC addresses",
								Required:    false,
								Optional:    true,
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
						},
					},
					"destination": schema.SingleNestedAttribute{
						Description: "Destination traffic matching criteria",
						Required:    true,
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
						},
						Attributes: map[string]schema.Attribute{
							"vlan": schema.ListAttribute{
								ElementType: types.Int64Type,
								Description: "VLAN IDs",
								Required:    false,
								Optional:    true,
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
							"ip": schema.ListAttribute{
								ElementType: types.StringType,
								Description: "IP addresses",
								Required:    false,
								Optional:    true,
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
							"subnet": schema.ListAttribute{
								ElementType: types.StringType,
								Description: "Subnets in CIDR notation",
								Required:    false,
								Optional:    true,
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
							"ip_range": schema.ListNestedAttribute{
								Description: "IP ranges",
								Required:    false,
								Optional:    true,
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
								Description:   "Hosts",
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
								Description:   "Sites",
								Required:      false,
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("site")},
							},
							"group": schema.SetNestedAttribute{
								Description:   "Groups",
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
								Description:   "System groups",
								Required:      false,
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("system_group")},
							},
							"network_interface": schema.SetNestedAttribute{
								Description:   "Network interfaces",
								Required:      false,
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("network_interface")},
							},
							"global_ip_range": schema.SetNestedAttribute{
								Description:   "Global IP ranges",
								Required:      false,
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("global_ip_range")},
							},
							"floating_subnet": schema.SetNestedAttribute{
								Description:   "Floating subnets",
								Required:      false,
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("floating_subnet")},
							},
							"site_network_subnet": schema.SetNestedAttribute{
								Description:   "Site network subnets",
								Required:      false,
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("site_network_subnet")},
							},
						},
					},
					"application": schema.SingleNestedAttribute{
						Description: "Application matching criteria",
						Required:    false,
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"application": schema.SetNestedAttribute{
								Description:   "Applications",
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
								Description:   "Custom applications",
								Required:      false,
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("custom_app")},
							},
							"domain": schema.ListAttribute{
								ElementType: types.StringType,
								Description: "Domains",
								Required:    false,
								Optional:    true,
							},
							"fqdn": schema.ListAttribute{
								ElementType: types.StringType,
								Description: "FQDNs",
								Required:    false,
								Optional:    true,
							},
							"ip": schema.ListAttribute{
								ElementType: types.StringType,
								Description: "IP addresses",
								Required:    false,
								Optional:    true,
							},
							"subnet": schema.ListAttribute{
								ElementType: types.StringType,
								Description: "Subnets",
								Required:    false,
								Optional:    true,
							},
							"ip_range": schema.ListNestedAttribute{
								Description: "IP ranges",
								Required:    false,
								Optional:    true,
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
							"global_ip_range": schema.SetNestedAttribute{
								Description:   "Global IP ranges",
								Required:      false,
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("global_ip_range")},
							},
						},
					},
					"service": schema.SingleNestedAttribute{
						Description: "Service matching criteria",
						Required:    false,
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"simple": schema.SetNestedAttribute{
								Description: "Simple services",
								Required:    false,
								Optional:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Description: "Service name",
											Required:    true,
										},
									},
								},
							},
							"standard": schema.SetNestedAttribute{
								Description:   "Standard services",
								Required:      false,
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("standard")},
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
										Description: "Returns data for the alert frequency " +
											"(https://api.catonetworks.com/documentation/#definition-PolicyRuleTrackingFrequencyEnum)",
										Optional: true,
										Required: false,
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
										Description:   "Returns data for the Subscription Group that receives the alert",
										Required:      false,
										Optional:      true,
										Computed:      true,
										PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
										Validators: []validator.Set{
											setvalidator.SizeAtLeast(1),
										},
										NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("subscription_group")},
									},
									"webhook": schema.SetNestedAttribute{
										Description:   "Returns data for the Webhook that receives the alert",
										Required:      false,
										Optional:      true,
										Computed:      true,
										PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
										Validators: []validator.Set{
											setvalidator.SizeAtLeast(1),
										},
										NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("webhook")},
									},
									"mailing_list": schema.SetNestedAttribute{
										Description:   "Returns data for the Mailing List that receives the alert",
										Required:      false,
										Optional:      true,
										Computed:      true,
										PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
										Validators: []validator.Set{
											setvalidator.SizeAtLeast(1),
										},
										NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("mailing_list")},
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
