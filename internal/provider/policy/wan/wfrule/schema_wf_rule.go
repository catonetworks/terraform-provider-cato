//nolint:lll
package wfrule

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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

	"github.com/catonetworks/terraform-provider-cato/internal/provider/planmodifiers"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

//nolint:funlen
func (r *WanFwRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_wf_rule` resource contains the configuration parameters necessary to add rule to the WAN Firewall. (check https://support.catonetworks.com/hc/en-us/articles/4413265660305-What-is-the-Cato-WAN-Firewall for more details). Documentation for the underlying API used in this resource can be found at [mutation.policy.wanFirewall.addRule()](https://api.catonetworks.com/documentation/#mutation-policy.wanFirewall.addRule).",
		Attributes: map[string]schema.Attribute{
			"sub_policy_id": schema.StringAttribute{
				Description: "Optional ID of a cato_wf_sub_policy that should own this rule. When set, the rule is created inside the sub-policy (positioned before the sub-policy cleanup rule). Immutable: changing it forces replacement.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
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
							// SUB_POLICY is an internal response action for sub-policy scope rules.
							stringvalidator.OneOf(
								string(cato_models.WanFirewallActionEnumAllow),
								string(cato_models.WanFirewallActionEnumBlock),
								string(cato_models.WanFirewallActionEnumPrompt),
							),
						},
					},
					"direction": schema.StringAttribute{
						Description: "Define the direction on which the rule is applied (https://api.catonetworks.com/documentation/#definition-WanFirewallDirectionEnum)",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf(utils.ToStrings(cato_models.AllWanFirewallDirectionEnum)...),
						},
					},
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
						Description: "Source traffic matching criteria. Logical 'OR' is applied within the criteria set. Logical 'AND' is applied between criteria sets. Can be an empty object {} to match any source.",
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
								Description:   "Site defined for the account",
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
								Description:   "Globally defined IP range, IP and subnet objects",
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
								Description:   "GlobalRange + InterfaceSubnet",
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
						Description: "Destination traffic matching criteria. Logical 'OR' is applied within the criteria set. Logical 'AND' is applied between criteria sets. Can be an empty object {} to match any destination.",
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
								Description:   "Site defined for the account",
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
								Description:   "Globally defined IP range, IP and subnet objects",
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
								Description:   "GlobalRange + InterfaceSubnet",
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
					"connection_origin": schema.StringAttribute{
						Description: "Connection origin of the traffic (https://api.catonetworks.com/documentation/#definition-ConnectionOriginEnum)",
						Optional:    true,
						Required:    false,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.String{
							stringvalidator.OneOf(utils.ToStrings(cato_models.AllConnectionOriginEnum)...),
						},
						Computed: true,
					},
					"country": schema.SetNestedAttribute{
						Description:   "Source country traffic matching criteria. Logical ‘OR’ is applied within the criteria set. Logical ‘AND’ is applied between criteria sets.",
						Required:      false,
						Optional:      true,
						Computed:      true,
						PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
						Validators: []validator.Set{
							setvalidator.SizeAtLeast(1),
						},
						NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("country")},
					},
					"device": schema.SetNestedAttribute{
						Description:   "Source Device Profile traffic matching criteria. Logical 'OR' is applied within the criteria set. Logical 'AND' is applied between criteria sets.",
						Required:      false,
						Optional:      true,
						Computed:      true,
						PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
						Validators: []validator.Set{
							setvalidator.SizeAtLeast(1),
						},
						NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("device")},
					},
					"device_attributes": schema.SingleNestedAttribute{
						Description: "Device attributes matching criteria for the rule.",
						Required:    false,
						Optional:    true,
						Computed:    true,
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
						Description: "Application traffic matching criteria. Logical 'OR' is applied within the criteria set. Logical 'AND' is applied between criteria sets. Can be an empty object {} to match any application.",
						Required:    true,
						Optional:    false,
						Attributes: map[string]schema.Attribute{
							"application": schema.SetNestedAttribute{
								Description:   "Applications for the rule (pre-defined)",
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
								Description:   "Custom (user-defined) applications",
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
								Description:   "Cato category of applications which are dynamically updated by Cato",
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
								Description:   "Custom Categories – Groups of objects such as predefined and custom applications, predefined and custom services, domains, FQDNs etc.",
								Required:      false,
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("custom_category")},
							},
							"sanctioned_apps_category": schema.SetNestedAttribute{
								Description:   "Sanctioned Cloud Applications - apps that are approved and generally represent an understood and acceptable level of risk in your organization.",
								Required:      false,
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
								Validators: []validator.Set{
									setvalidator.SizeAtLeast(1),
								},
								NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("sanctioned_apps_category")},
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
								Description:   "Globally defined IP range, IP and subnet objects.",
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
						Description: "Destination service traffic matching criteria. Logical 'OR' is applied within the criteria set. Logical 'AND' is applied between criteria sets.",
						Required:    false,
						Optional:    true,
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(), // Avoid drift
						},
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"standard": schema.SetNestedAttribute{
								Description:   "Standard Service to which this Internet Firewall rule applies",
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
										Description: "Returns data for the alert frequency (https://api.catonetworks.com/documentation/#definition-PolicyRuleTrackingFrequencyEnum)",
										Optional:    true,
										Required:    false,
										Validators: []validator.String{
											stringvalidator.OneOf(utils.ToStrings(cato_models.AllPolicyRuleTrackingFrequencyEnum)...),
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
									stringvalidator.OneOf(utils.ToStrings(cato_models.AllPolicyActiveOnEnum)...),
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
											listvalidator.ValueStringsAre(stringvalidator.OneOf(utils.ToStrings(cato_models.AllDayOfWeek)...)),
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
							setplanmodifier.UseStateForUnknown(),     // Preserve state value (e.g. empty set) when not in config
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
											Description:   "Floating Subnet defined for a site",
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
											Description:   "User defined for your account",
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
											Description:   "",
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
											Description:   "",
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
											Description:   "",
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
											Description:   "Site defined for the account",
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
											Description:   "Globally defined IP range, IP and subnet objects",
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
											Description:   "GlobalRange + InterfaceSubnet",
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
								"country": schema.SetNestedAttribute{
									Description:   "Source country traffic matching criteria. Logical 'OR' is applied within the criteria set. Logical 'AND' is applied between criteria sets.",
									Required:      false,
									Optional:      true,
									Computed:      true,
									PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
									NestedObject:  schema.NestedAttributeObject{Attributes: idname.SchemaNameID("country")},
								},
								"device": schema.SetNestedAttribute{
									Description:   "Source Device Profile traffic matching criteria. Logical 'OR' is applied within the criteria set. Logical 'AND' is applied between criteria sets.",
									Required:      false,
									Optional:      true,
									Computed:      true,
									PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
									Validators: []validator.Set{
										setvalidator.SizeAtLeast(1),
									},
									NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("device")},
								},
								"device_attributes": schema.SingleNestedAttribute{
									Description: "Device attributes matching criteria for the exception.",
									Required:    false,
									Optional:    true,
									Computed:    true,
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
										listvalidator.ValueStringsAre(stringvalidator.OneOf(utils.ToStrings(cato_models.AllOperatingSystem)...)),
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
											Description:   "",
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
											Description:   "",
											Required:      false,
											Optional:      true,
											Computed:      true,
											PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("custom_category")},
										},
										"sanctioned_apps_category": schema.SetNestedAttribute{
											Required:      false,
											Optional:      true,
											Computed:      true,
											PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
											Validators: []validator.Set{
												setvalidator.SizeAtLeast(1),
											},
											NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("sanctioned_apps_category")},
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
											Description:   "Global IP range matching criteria for the exception.",
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
									Description: "Destination service matching criteria for the exception. This field is required when defining exceptions.",
									Required:    true,
									Optional:    false,
									Attributes: map[string]schema.Attribute{
										"standard": schema.SetNestedAttribute{
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
										stringvalidator.OneOf(utils.ToStrings(cato_models.AllWanFirewallDirectionEnum)...),
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
										stringvalidator.OneOf(utils.ToStrings(cato_models.AllConnectionOriginEnum)...),
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
