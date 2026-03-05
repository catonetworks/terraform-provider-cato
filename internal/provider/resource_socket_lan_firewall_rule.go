package provider

import (
	"context"
	"fmt"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
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
	_ resource.Resource                = &socketLanFirewallRuleResource{}
	_ resource.ResourceWithConfigure   = &socketLanFirewallRuleResource{}
	_ resource.ResourceWithImportState = &socketLanFirewallRuleResource{}
)

func NewSocketLanFirewallRuleResource() resource.Resource {
	return &socketLanFirewallRuleResource{}
}

type socketLanFirewallRuleResource struct {
	client *catoClientData
}

func (r *socketLanFirewallRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_socket_lan_firewall_rule"
}

func (r *socketLanFirewallRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_socket_lan_firewall_rule` resource contains the configuration parameters necessary to add a Socket LAN firewall rule (child rule under a network rule). Documentation for the underlying API used in this resource can be found at [mutation.policy.socketLan.firewall.addRule()](https://api.catonetworks.com/documentation/#mutation-policy.socketLan.firewall.addRule).",
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
								Description: "Hosts",
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
											Description: "Host name",
											Required:    false,
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"id": schema.StringAttribute{
											Description: "Host ID",
											Required:    false,
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
							"site": schema.SetNestedAttribute{
								Description: "Sites",
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
											Description: "Site name",
											Required:    false,
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"id": schema.StringAttribute{
											Description: "Site ID",
											Required:    false,
											Optional:    true,
											Computed:    true,
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
										"name": schema.StringAttribute{
											Description: "Group name",
											Required:    false,
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"id": schema.StringAttribute{
											Description: "Group ID",
											Required:    false,
											Optional:    true,
											Computed:    true,
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
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Description: "System group name",
											Required:    false,
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"id": schema.StringAttribute{
											Description: "System group ID",
											Required:    false,
											Optional:    true,
											Computed:    true,
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
										"name": schema.StringAttribute{
											Description: "Network interface name",
											Required:    false,
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"id": schema.StringAttribute{
											Description: "Network interface ID",
											Required:    false,
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
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
										"name": schema.StringAttribute{
											Description: "Global IP range name",
											Required:    false,
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"id": schema.StringAttribute{
											Description: "Global IP range ID",
											Required:    false,
											Optional:    true,
											Computed:    true,
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
										"name": schema.StringAttribute{
											Description: "Floating subnet name",
											Required:    false,
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"id": schema.StringAttribute{
											Description: "Floating subnet ID",
											Required:    false,
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
											Description: "Site network subnet name",
											Required:    false,
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"id": schema.StringAttribute{
											Description: "Site network subnet ID",
											Required:    false,
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
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
								Description: "Hosts",
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
											Description: "Host name",
											Required:    false,
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"id": schema.StringAttribute{
											Description: "Host ID",
											Required:    false,
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
							"site": schema.SetNestedAttribute{
								Description: "Sites",
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
											Description: "Site name",
											Required:    false,
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"id": schema.StringAttribute{
											Description: "Site ID",
											Required:    false,
											Optional:    true,
											Computed:    true,
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
										"name": schema.StringAttribute{
											Description: "Group name",
											Required:    false,
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"id": schema.StringAttribute{
											Description: "Group ID",
											Required:    false,
											Optional:    true,
											Computed:    true,
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
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Description: "System group name",
											Required:    false,
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"id": schema.StringAttribute{
											Description: "System group ID",
											Required:    false,
											Optional:    true,
											Computed:    true,
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
										"name": schema.StringAttribute{
											Description: "Network interface name",
											Required:    false,
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"id": schema.StringAttribute{
											Description: "Network interface ID",
											Required:    false,
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
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
										"name": schema.StringAttribute{
											Description: "Global IP range name",
											Required:    false,
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"id": schema.StringAttribute{
											Description: "Global IP range ID",
											Required:    false,
											Optional:    true,
											Computed:    true,
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
										"name": schema.StringAttribute{
											Description: "Floating subnet name",
											Required:    false,
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"id": schema.StringAttribute{
											Description: "Floating subnet ID",
											Required:    false,
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
											Description: "Site network subnet name",
											Required:    false,
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
										"id": schema.StringAttribute{
											Description: "Site network subnet ID",
											Required:    false,
											Optional:    true,
											Computed:    true,
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
						Required:    false,
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"application": schema.SetNestedAttribute{
								Description: "Applications",
								Required:    false,
								Optional:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Description: "Application name",
											Required:    false,
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
										"id": schema.StringAttribute{
											Description: "Application ID",
											Required:    false,
											Optional:    true,
											Computed:    true,
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
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Description: "Custom app name",
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
											Description: "Custom app ID",
											Required:    false,
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
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
								Description: "Global IP ranges",
								Required:    false,
								Optional:    true,
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Description: "Global IP range name",
											Required:    false,
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
										"id": schema.StringAttribute{
											Description: "Global IP range ID",
											Required:    false,
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
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
								Description: "Standard services",
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
											Description: "Service name",
											Required:    false,
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
										"id": schema.StringAttribute{
											Description: "Service ID",
											Required:    false,
											Optional:    true,
											Computed:    true,
											PlanModifiers: []planmodifier.String{
												stringplanmodifier.UseStateForUnknown(),
											},
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
				},
			},
		},
	}
}

func (r *socketLanFirewallRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoClientData)
}

func (r *socketLanFirewallRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("rule").AtName("id"), req, resp)
}

func (r *socketLanFirewallRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SocketLanFirewallRule
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build API input
	apiInput, inputDiags := hydrateSocketLanFirewallRuleApi(ctx, plan)
	resp.Diagnostics.Append(inputDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Create.PolicySocketLanFirewallAddRule.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(apiInput.create),
	})

	policyChange, err := r.client.catov2.PolicySocketLanFirewallAddRule(ctx, r.client.AccountId, nil, apiInput.create)
	tflog.Debug(ctx, "Create.PolicySocketLanFirewallAddRule.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(policyChange),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicySocketLanFirewallAddRule error",
			err.Error(),
		)
		return
	}

	if len(policyChange.Policy.SocketLan.Firewall.AddRule.Errors) > 0 {
		for _, e := range policyChange.Policy.SocketLan.Firewall.AddRule.Errors {
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
	ruleId := policyChange.GetPolicy().GetSocketLan().GetFirewall().GetAddRule().Rule.GetRule().ID

	// Read back the rule to populate state
	queryResult, err := r.client.catov2.PolicySocketLanPolicy(ctx, r.client.AccountId, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicySocketLanPolicy error",
			err.Error(),
		)
		return
	}

	// Find the created firewall rule
	var currentRule *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Firewall_Rule
	for _, ruleWrapper := range queryResult.Policy.SocketLan.Policy.Rules {
		for _, fwWrapper := range ruleWrapper.Rule.Firewall {
			if fwWrapper.Rule.ID == ruleId {
				currentRule = &fwWrapper.Rule
				break
			}
		}
		if currentRule != nil {
			break
		}
	}

	if currentRule == nil {
		resp.Diagnostics.AddError(
			"Rule not found",
			fmt.Sprintf("Could not find created firewall rule with ID %s", ruleId),
		)
		return
	}

	// Hydrate state from API response
	ruleState := hydrateSocketLanFirewallRuleState(ctx, plan, currentRule)

	// Build the rule object
	ruleObj, diagstmp := types.ObjectValueFrom(ctx, SocketLanFirewallRuleRuleAttrTypes, ruleState)
	resp.Diagnostics.Append(diagstmp...)

	// Set the state
	plan.Rule = ruleObj
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *socketLanFirewallRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SocketLanFirewallRule
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the rule ID from state
	var ruleData SocketLanFirewallRuleData
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

	// Find the firewall rule
	var currentRule *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Firewall_Rule
	ruleExists := false
	for _, ruleWrapper := range queryResult.Policy.SocketLan.Policy.Rules {
		for _, fwWrapper := range ruleWrapper.Rule.Firewall {
			if fwWrapper.Rule.ID == ruleId {
				currentRule = &fwWrapper.Rule
				ruleExists = true
				break
			}
		}
		if currentRule != nil {
			break
		}
	}

	if !ruleExists {
		tflog.Warn(ctx, "socket lan firewall rule not found, resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	// Hydrate state from API response
	ruleState := hydrateSocketLanFirewallRuleState(ctx, state, currentRule)

	// Build the rule object
	ruleObj, diagstmp := types.ObjectValueFrom(ctx, SocketLanFirewallRuleRuleAttrTypes, ruleState)
	resp.Diagnostics.Append(diagstmp...)

	// Hard code position to avoid drift
	curAtObj, diagstmp := types.ObjectValue(
		PositionAttrTypes,
		map[string]attr.Value{
			"position": types.StringValue("LAST_IN_RULE"),
			"ref":      types.StringNull(),
		},
	)
	resp.Diagnostics.Append(diagstmp...)

	state.Rule = ruleObj
	state.At = curAtObj

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *socketLanFirewallRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SocketLanFirewallRule
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the rule ID from plan
	var ruleData SocketLanFirewallRuleData
	diags = plan.Rule.As(ctx, &ruleData, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ruleId := ruleData.ID.ValueString()

	// Build API input
	apiInput, inputDiags := hydrateSocketLanFirewallRuleApi(ctx, plan)
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

	moveInput := cato_models.PolicyMoveSubRuleInput{
		ID: ruleId,
		To: &cato_models.PolicySubRulePositionInput{
			Position: cato_models.PolicySubRulePositionEnum(positionInput.Position.ValueString()),
			Ref:      positionInput.Ref.ValueString(),
		},
	}

	tflog.Debug(ctx, "Update.PolicySocketLanFirewallMoveRule.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(moveInput),
	})
	moveResult, err := r.client.catov2.PolicySocketLanFirewallMoveRule(ctx, r.client.AccountId, nil, moveInput)
	tflog.Debug(ctx, "Update.PolicySocketLanFirewallMoveRule.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(moveResult),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicySocketLanFirewallMoveRule error",
			err.Error(),
		)
		return
	}

	// Update the rule
	tflog.Debug(ctx, "Update.PolicySocketLanFirewallUpdateRule.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(apiInput.update),
	})
	updateResult, err := r.client.catov2.PolicySocketLanFirewallUpdateRule(ctx, r.client.AccountId, nil, apiInput.update)
	tflog.Debug(ctx, "Update.PolicySocketLanFirewallUpdateRule.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(updateResult),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicySocketLanFirewallUpdateRule error",
			err.Error(),
		)
		return
	}

	if updateResult.Policy.SocketLan.Firewall.UpdateRule.Status != "SUCCESS" {
		for _, item := range updateResult.Policy.SocketLan.Firewall.UpdateRule.GetErrors() {
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

	var currentRule *cato_go_sdk.PolicySocketLanPolicy_Policy_SocketLan_Policy_Rules_Rule_Firewall_Rule
	for _, ruleWrapper := range queryResult.Policy.SocketLan.Policy.Rules {
		for _, fwWrapper := range ruleWrapper.Rule.Firewall {
			if fwWrapper.Rule.ID == ruleId {
				currentRule = &fwWrapper.Rule
				break
			}
		}
		if currentRule != nil {
			break
		}
	}

	if currentRule == nil {
		resp.Diagnostics.AddError(
			"Rule not found",
			fmt.Sprintf("Could not find firewall rule with ID %s after update", ruleId),
		)
		return
	}

	// Hydrate state from API response
	ruleState := hydrateSocketLanFirewallRuleState(ctx, plan, currentRule)

	// Build the rule object
	ruleObj, diagstmp := types.ObjectValueFrom(ctx, SocketLanFirewallRuleRuleAttrTypes, ruleState)
	resp.Diagnostics.Append(diagstmp...)

	plan.Rule = ruleObj
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *socketLanFirewallRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SocketLanFirewallRule
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the rule ID from state
	var ruleData SocketLanFirewallRuleData
	diags = state.Rule.As(ctx, &ruleData, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ruleId := ruleData.ID.ValueString()

	removeInput := cato_models.SocketLanFirewallRemoveRuleInput{
		ID: ruleId,
	}

	tflog.Debug(ctx, "Delete.PolicySocketLanFirewallRemoveRule.request", map[string]interface{}{
		"request": utils.InterfaceToJSONString(removeInput),
	})
	removeResult, err := r.client.catov2.PolicySocketLanFirewallRemoveRule(ctx, r.client.AccountId, nil, removeInput)
	tflog.Debug(ctx, "Delete.PolicySocketLanFirewallRemoveRule.response", map[string]interface{}{
		"response": utils.InterfaceToJSONString(removeResult),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicySocketLanFirewallRemoveRule error",
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
