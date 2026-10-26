package socketlannetworkrule

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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/planmodifiers"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

func (r *socketLanNetworkRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_socket_lan_network_rule` resource contains the configuration parameters necessary " +
			"to add a Socket LAN network rule. Documentation for the underlying API used in this resource can be " +
			"found at [mutation.policy.socketLan.addRule()]" +
			"(https://api.catonetworks.com/documentation/#mutation-policy.socketLan.addRule).",
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
			"rule": r.lanRuleSchema(),
		},
	}
}

//nolint:funlen
func (r *socketLanNetworkRuleResource) lanRuleSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Parameters for the Socket LAN network rule",
		Required:    true,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "ID of the rule",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Description: "Name of the rule",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description:   "Description of the rule",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"enabled": schema.BoolAttribute{
				Description: "Enable or disable the rule",
				Required:    true,
			},
			"direction": schema.StringAttribute{
				Description: "Direction of the traffic (TO, BOTH)",
				Required:    true,
				Validators:  []validator.String{stringvalidator.OneOf("TO", "BOTH")},
			},
			"transport": schema.StringAttribute{
				Description: "Transport type (LAN, WAN)",
				Required:    true,
				Validators:  []validator.String{stringvalidator.OneOf("LAN", "WAN")},
			},

			"site": schema.SingleNestedAttribute{
				Description: "Site scope for the rule",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"site": schema.SetNestedAttribute{
						Description: "List of sites",
						Optional:    true,
						Computed:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes:    idname.SchemaNameID("Site"),
							PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
						},
						PlanModifiers: []planmodifier.Set{setplanmodifier.UseStateForUnknown()},
					},
					"group": schema.SetNestedAttribute{
						Description: "List of groups",
						Optional:    true,
						Computed:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes:    idname.SchemaNameID("Group"),
							PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
						},
						PlanModifiers: []planmodifier.Set{setplanmodifier.UseStateForUnknown()},
					},
				},
			},
			"source":      r.lanRuleSourceDestinationSchema("Source traffic matching criteria"),
			"destination": r.lanRuleSourceDestinationSchema("Destination traffic matching criteria"),
			"service": schema.SingleNestedAttribute{
				Description: "Destination service traffic matching criteria. Logical 'OR' is applied within " +
					"the criteria set. Logical 'AND' is applied between criteria sets.",
				Required: false,
				Optional: true,
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
									Optional:    true,
									Validators: []validator.Object{
										objectvalidator.ConflictsWith(path.Expressions{
											path.MatchRelative().AtParent().AtName("port"),
										}...),
									},
									Attributes: map[string]schema.Attribute{
										"from": schema.StringAttribute{
											Description: "From port number",
											Required:    true,
										},
										"to": schema.StringAttribute{
											Description: "To port number",
											Required:    true,
										},
									},
								},
								"protocol": schema.StringAttribute{
									Description: "IP Protocol (https://api.catonetworks.com/documentation/#definition-IpProtocol)",
									Optional:    true,
								},
							},
						},
					},
				},
			},
			"nat": schema.SingleNestedAttribute{
				Description:   "NAT settings",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "Enable or disable NAT",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"nat_type": schema.StringAttribute{
						Description: "NAT type (DYNAMIC_PAT)",
						Optional:    true,
						Validators:  []validator.String{stringvalidator.OneOf("DYNAMIC_PAT")},
					},
				},
			},
		},
	}
}

//nolint:funlen
func (r *socketLanNetworkRuleResource) lanRuleSourceDestinationSchema(description string) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: description,
		Required:    true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.UseStateForUnknown(),
			planmodifiers.SourceDestObjectModifier(),
		},
		Attributes: map[string]schema.Attribute{
			"vlan": schema.ListAttribute{
				ElementType:   types.Int64Type,
				Description:   "VLAN IDs",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()},
				Validators:    []validator.List{listvalidator.SizeAtLeast(1)},
			},
			"ip": schema.ListAttribute{
				ElementType:   types.StringType,
				Description:   "IP addresses",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()},
				Validators:    []validator.List{listvalidator.SizeAtLeast(1)},
			},
			"subnet": schema.ListAttribute{
				ElementType:   types.StringType,
				Description:   "Subnets in CIDR notation",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()},
				Validators:    []validator.List{listvalidator.SizeAtLeast(1)},
			},
			"ip_range": schema.ListNestedAttribute{
				Description:   "IP ranges",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()},
				Validators:    []validator.List{listvalidator.SizeAtLeast(1)},
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
				Description: "List of hosts",
				Optional:    true,
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes:    idname.SchemaNameID("Host"),
					PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
				},
				PlanModifiers: []planmodifier.Set{setplanmodifier.UseStateForUnknown()},
			},
			"group": schema.SetNestedAttribute{
				Description: "List of groups",
				Optional:    true,
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes:    idname.SchemaNameID("Group"),
					PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
				},
				PlanModifiers: []planmodifier.Set{setplanmodifier.UseStateForUnknown()},
			},
			"system_group": schema.SetNestedAttribute{
				Description: "List of system groups",
				Optional:    true,
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes:    idname.SchemaNameID("System group"),
					PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
				},
				PlanModifiers: []planmodifier.Set{setplanmodifier.UseStateForUnknown()},
			},
			"network_interface": schema.SetNestedAttribute{
				Description: "List of network interfaces",
				Optional:    true,
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes:    idname.SchemaNameID("Network interface"),
					PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
				},
				PlanModifiers: []planmodifier.Set{setplanmodifier.UseStateForUnknown()},
			},
			"global_ip_range": schema.SetNestedAttribute{
				Description: "List of global IP ranges",
				Optional:    true,
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes:    idname.SchemaNameID("Global IP range"),
					PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
				},
				PlanModifiers: []planmodifier.Set{setplanmodifier.UseStateForUnknown()},
			},
			"floating_subnet": schema.SetNestedAttribute{
				Description: "Floating Subnets (ie. Floating Ranges) are used to identify traffic " +
					"exactly matched to the route advertised by BGP. They are not associated with a " +
					"specific site. This is useful in scenarios such as active-standby high " +
					"availability routed via BGP.",
				Optional: true,
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes:    idname.SchemaNameID("Floating subnet"),
					PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
				},
				PlanModifiers: []planmodifier.Set{setplanmodifier.UseStateForUnknown()},
			},
			"site_network_subnet": schema.SetNestedAttribute{
				Description: "GlobalRange + InterfaceSubnet",
				Optional:    true,
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes:    idname.SchemaNameID("Site Network Subnet"),
					PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
				},
				PlanModifiers: []planmodifier.Set{setplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func LanRuleSchema() schema.SingleNestedAttribute {
	return (&socketLanNetworkRuleResource{}).lanRuleSchema()
}
