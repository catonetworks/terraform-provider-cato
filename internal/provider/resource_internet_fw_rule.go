package provider

import (
	"context"
	"fmt"
	"reflect"
	"runtime"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/cato-go-sdk/scalars"
	cato_scalars "github.com/catonetworks/cato-go-sdk/scalars"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &internetFwRuleResource{}
	_ resource.ResourceWithConfigure   = &internetFwRuleResource{}
	_ resource.ResourceWithImportState = &internetFwRuleResource{}
)

func NewInternetFwRuleResource() resource.Resource {
	return &internetFwRuleResource{}
}

type internetFwRuleResource struct {
	client *catoClientData
}

func (r *internetFwRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_if_rule"
}

func (r *internetFwRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_if_rule` resource contains the configuration parameters necessary to add rule to the Internet Firewall. (check https://support.catonetworks.com/hc/en-us/articles/4413273486865-What-is-the-Cato-Internet-Firewall for more details). Documentation for the underlying API used in this resource can be found at [mutation.policy.internetFirewall.addRule()](https://api.catonetworks.com/documentation/#mutation-policy.internetFirewall.addRule).",
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
					"section": schema.SingleNestedAttribute{
						Required: false,
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{
								Description: "",
								Required:    false,
								Optional:    true,
							},
							"id": schema.StringAttribute{
								Description: "",
								Required:    false,
								Optional:    true,
							},
						},
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
							"host": schema.ListNestedAttribute{
								Description: "Hosts and servers defined for your account",
								Required:    false,
								Optional:    true,
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
							"site": schema.ListNestedAttribute{
								Description: "Site defined for the account",
								Required:    false,
								Optional:    true,
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
							"global_ip_range": schema.ListNestedAttribute{
								Description: "Globally defined IP range, IP and subnet objects",
								Required:    false,
								Optional:    true,
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
							"network_interface": schema.ListNestedAttribute{
								Description: "Network range defined for a site",
								Required:    false,
								Optional:    true,
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
							"site_network_subnet": schema.ListNestedAttribute{
								Description: "GlobalRange + InterfaceSubnet",
								Required:    false,
								Optional:    true,
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
							"floating_subnet": schema.ListNestedAttribute{
								Description: "Floating Subnets (ie. Floating Ranges) are used to identify traffic exactly matched to the route advertised by BGP. They are not associated with a specific site. This is useful in scenarios such as active-standby high availability routed via BGP.",
								Required:    false,
								Optional:    true,
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
											// Default:     defaults.String(nil),
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
											Computed:    true,
										},
									},
								},
							},
							"user": schema.ListNestedAttribute{
								Description: "Individual users defined for the account",
								Required:    false,
								Optional:    true,
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
							"users_group": schema.ListNestedAttribute{
								Description: "Group of users",
								Required:    false,
								Optional:    true,
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
							"group": schema.ListNestedAttribute{
								Description: "Groups defined for your account",
								Required:    false,
								Optional:    true,
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
							"system_group": schema.ListNestedAttribute{
								Description: "Predefined Cato groups",
								Required:    false,
								Optional:    true,
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
					"connection_origin": schema.StringAttribute{
						Description: "Connection origin of the traffic (https://api.catonetworks.com/documentation/#definition-ConnectionOriginEnum)",
						Optional:    true,
						Required:    false,
					},
					"country": schema.ListNestedAttribute{
						Description: "Source country traffic matching criteria. Logical ‘OR’ is applied within the criteria set. Logical ‘AND’ is applied between criteria sets.",
						Required:    false,
						Optional:    true,
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
					"device": schema.ListNestedAttribute{
						Description: "Source Device Profile traffic matching criteria. Logical ‘OR’ is applied within the criteria set. Logical ‘AND’ is applied between criteria sets.",
						Required:    false,
						Optional:    true,
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
					"destination": schema.SingleNestedAttribute{
						Description: "Destination traffic matching criteria. Logical ‘OR’ is applied within the criteria set. Logical ‘AND’ is applied between criteria sets.",
						Optional:    true,
						Required:    false,
						Attributes: map[string]schema.Attribute{
							"application": schema.ListNestedAttribute{
								Description: "Applications for the rule (pre-defined)",
								Required:    false,
								Optional:    true,
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
							"custom_app": schema.ListNestedAttribute{
								Description: "Custom (user-defined) applications",
								Required:    false,
								Optional:    true,
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
							"app_category": schema.ListNestedAttribute{
								Description: "Cato category of applications which are dynamically updated by Cato",
								Required:    false,
								Optional:    true,
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
							"custom_category": schema.ListNestedAttribute{
								Description: "Custom Categories – Groups of objects such as predefined and custom applications, predefined and custom services, domains, FQDNs etc.",
								Required:    false,
								Optional:    true,
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
							"sanctioned_apps_category": schema.ListNestedAttribute{
								Description: "Sanctioned Cloud Applications - apps that are approved and generally represent an understood and acceptable level of risk in your organization.",
								Required:    false,
								Optional:    true,
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
							"country": schema.ListNestedAttribute{
								Description: "Countries",
								Required:    false,
								Optional:    true,
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
							"global_ip_range": schema.ListNestedAttribute{
								Description: "Globally defined IP range, IP and subnet objects",
								Required:    false,
								Optional:    true,
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
							"remote_asn": schema.ListAttribute{
								ElementType: types.StringType,
								Description: "Remote Autonomous System Number (ASN)",
								Required:    false,
								Optional:    true,
							},
							// "containers": schema.ListNestedAttribute{
							// 	Description: "Globaly defined containers, FQDN and IP objects",
							// 	Required:    false,
							// 	Optional:    true,
							// 	NestedObject: schema.NestedAttributeObject{
							// 		Attributes: map[string]schema.Attribute{
							// 			"fqdnContainer": schema.ListNestedAttribute{
							// 				Description: "Globaly defined container for FQDN ",
							// 				Required:    false,
							// 				Optional:    true,
							// 				NestedObject: schema.NestedAttributeObject{
							// 					Attributes: map[string]schema.Attribute{
							// 						"name": schema.StringAttribute{
							// 							Description: "",
							// 							Required:    false,
							// 							Optional:    true,
							// 							Validators: []validator.String{
							// 								stringvalidator.ConflictsWith(path.Expressions{
							// 									path.MatchRelative().AtParent().AtName("id"),
							// 								}...),
							// 							},
							// 						},
							// 						"id": schema.StringAttribute{
							// 							Description: "",
							// 							Required:    false,
							// 							Optional:    true,
							// 						},
							// 					},
							// 				},
							// 			},
							// 			"ipAddressRangeContainer": schema.ListNestedAttribute{
							// 				Description: "Globaly defined container for FQDN ",
							// 				Required:    false,
							// 				Optional:    true,
							// 				NestedObject: schema.NestedAttributeObject{
							// 					Attributes: map[string]schema.Attribute{
							// 						"name": schema.StringAttribute{
							// 							Description: "",
							// 							Required:    false,
							// 							Optional:    true,
							// 							Validators: []validator.String{
							// 								stringvalidator.ConflictsWith(path.Expressions{
							// 									path.MatchRelative().AtParent().AtName("id"),
							// 								}...),
							// 							},
							// 						},
							// 						"id": schema.StringAttribute{
							// 							Description: "",
							// 							Required:    false,
							// 							Optional:    true,
							// 						},
							// 					},
							// 				},
							// 			},
							// 		},
							// 	},
							// },
						},
					},
					"service": schema.SingleNestedAttribute{
						Description: "Destination service traffic matching criteria. Logical ‘OR’ is applied within the criteria set. Logical ‘AND’ is applied between criteria sets.",
						Required:    false,
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"standard": schema.ListNestedAttribute{
								Description: "Standard Service to which this Internet Firewall rule applies",
								Required:    false,
								Optional:    true,
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
					"action": schema.StringAttribute{
						Description: "The action applied by the Internet Firewall if the rule is matched (https://api.catonetworks.com/documentation/#definition-InternetFirewallActionEnum)",
						Required:    true,
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
									},
									"frequency": schema.StringAttribute{
										Description: "Returns data for the alert frequency (https://api.catonetworks.com/documentation/#definition-PolicyRuleTrackingFrequencyEnum)",
										Required:    true,
									},
									"subscription_group": schema.ListNestedAttribute{
										Description: "Returns data for the Subscription Group that receives the alert",
										Required:    false,
										Optional:    true,
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
									"webhook": schema.ListNestedAttribute{
										Description: "Returns data for the Webhook that receives the alert",
										Required:    false,
										Optional:    true,
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
									"mailing_list": schema.ListNestedAttribute{
										Description: "Returns data for the Mailing List that receives the alert",
										Required:    false,
										Optional:    true,
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
					"exceptions": schema.ListNestedAttribute{
						Description: "The set of exceptions for the rule. Exceptions define when the rule will be ignored and the firewall evaluation will continue with the lower priority rules.",
						Required:    false,
						Optional:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Description: "A unique name of the rule exception.",
									Required:    false,
									Optional:    true,
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
										},
										"host": schema.ListNestedAttribute{
											Required: false,
											Optional: true,
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
										"site": schema.ListNestedAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
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
										"subnet": schema.ListAttribute{
											ElementType: types.StringType,
											Description: "",
											Required:    false,
											Optional:    true,
										},
										"ip_range": schema.ListNestedAttribute{
											Description: "",
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
										"global_ip_range": schema.ListNestedAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
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
										"network_interface": schema.ListNestedAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
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
										"site_network_subnet": schema.ListNestedAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
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
										"floating_subnet": schema.ListNestedAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
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
										"user": schema.ListNestedAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
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
										"users_group": schema.ListNestedAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
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
										"group": schema.ListNestedAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
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
										"system_group": schema.ListNestedAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
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
								"country": schema.ListNestedAttribute{
									Description: "Source country matching criteria for the exception.",
									Required:    false,
									Optional:    true,
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
								"device": schema.ListAttribute{
									ElementType: types.StringType,
									Description: "Source Device Profile matching criteria for the exception.",
									Optional:    true,
									Required:    false,
								},
								"device_os": schema.ListAttribute{
									ElementType: types.StringType,
									Description: "Source device OS matching criteria for the exception. (https://api.catonetworks.com/documentation/#definition-OperatingSystem)",
									Optional:    true,
									Required:    false,
								},
								"destination": schema.SingleNestedAttribute{
									Description: "Destination service matching criteria for the exception.",
									Optional:    true,
									Required:    false,
									Attributes: map[string]schema.Attribute{
										"application": schema.ListNestedAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
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
										"custom_app": schema.ListNestedAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
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
										"app_category": schema.ListNestedAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
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
										"custom_category": schema.ListNestedAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
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
										"sanctioned_apps_category": schema.ListNestedAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
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
										"country": schema.ListNestedAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
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
											Description: "",
											Required:    false,
											Optional:    true,
										},
										"fqdn": schema.ListAttribute{
											ElementType: types.StringType,
											Description: "",
											Required:    false,
											Optional:    true,
										},
										"ip": schema.ListAttribute{
											ElementType: types.StringType,
											Description: "",
											Required:    false,
											Optional:    true,
										},
										"subnet": schema.ListAttribute{
											ElementType: types.StringType,
											Description: "",
											Required:    false,
											Optional:    true,
										},
										"ip_range": schema.ListNestedAttribute{
											Description: "",
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
										"global_ip_range": schema.ListNestedAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
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
										"remote_asn": schema.ListAttribute{
											ElementType: types.StringType,
											Description: "",
											Required:    false,
											Optional:    true,
										},
									},
								},
								"service": schema.SingleNestedAttribute{
									Description: "Destination service matching criteria for the exception.",
									Required:    false,
									Optional:    true,
									Attributes: map[string]schema.Attribute{
										"standard": schema.ListNestedAttribute{
											Required: false,
											Optional: true,
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
										"custom": schema.ListNestedAttribute{
											Description: "",
											Required:    false,
											Optional:    true,
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"port": schema.ListAttribute{
														ElementType: types.StringType,
														Description: "",
														Optional:    true,
														Required:    false,
													},
													"port_range": schema.SingleNestedAttribute{
														Description: "",
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
														Description: "",
														Required:    false,
														Optional:    true,
													},
												},
											},
										},
									},
								},
								"connection_origin": schema.StringAttribute{
									Description: "Connection origin matching criteria for the exception.",
									Optional:    true,
									Required:    false,
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *internetFwRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoClientData)
}

func (r *internetFwRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("rule").AtName("id"), req, resp)
}

func (r *internetFwRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	var plan InternetFirewallRule
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	//initiate input
	input := cato_models.InternetFirewallAddRuleInput{}

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

		input.Rule = &cato_models.InternetFirewallAddRuleDataInput{}
		ruleInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule{}
		diags = plan.Rule.As(ctx, &ruleInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)

		// setting source
		if !ruleInput.Source.IsNull() {

			input.Rule.Source = &cato_models.InternetFirewallSourceInput{}
			sourceInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source{}
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

				var itemSourceHostInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_Host
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

				var itemSourceSiteInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_Site
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

				var itemSourceIPRangeInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_IPRange
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

				var itemSourceGlobalIPRangeInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_GlobalIPRange
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

				var itemSourceNetworkInterfaceInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_NetworkInterface
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

				var itemSourceSiteNetworkSubnetInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_SiteNetworkSubnet
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

				var itemSourceFloatingSubnetInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_FloatingSubnet
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

				var itemSourceUserInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_User
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

				var itemSourceUsersGroupInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_UsersGroup
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

				var itemSourceGroupInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_Group
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

				var itemSourceSystemGroupInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_SystemGroup
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

			var itemCountryInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Country
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

			var itemDeviceInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Device
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
			input.Rule.Destination = &cato_models.InternetFirewallDestinationInput{}
			destinationInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination{}
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

			// setting destination domain
			if !destinationInput.Domain.IsNull() {
				diags = destinationInput.Domain.ElementsAs(ctx, &input.Rule.Destination.Domain, false)
				resp.Diagnostics.Append(diags...)
			}

			// setting destination fqdn
			if !destinationInput.Fqdn.IsNull() {
				diags = destinationInput.Fqdn.ElementsAs(ctx, &input.Rule.Destination.Fqdn, false)
				resp.Diagnostics.Append(diags...)
			}

			// setting destination remote asn
			if !destinationInput.RemoteAsn.IsNull() {
				diags = destinationInput.RemoteAsn.ElementsAs(ctx, &input.Rule.Destination.RemoteAsn, false)
				resp.Diagnostics.Append(diags...)
			}

			// setting destination application
			if !destinationInput.Application.IsNull() {
				elementsDestinationApplicationInput := make([]types.Object, 0, len(destinationInput.Application.Elements()))
				diags = destinationInput.Application.ElementsAs(ctx, &elementsDestinationApplicationInput, false)
				resp.Diagnostics.Append(diags...)

				var itemDestinationApplicationInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_Application
				for _, item := range elementsDestinationApplicationInput {
					diags = item.As(ctx, &itemDestinationApplicationInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationApplicationInput)
					if err != nil {
						resp.Diagnostics.AddError(
							"Object Ref transformation failed",
							err.Error(),
						)
						return
					}

					input.Rule.Destination.Application = append(input.Rule.Destination.Application, &cato_models.ApplicationRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
			}

			// setting destination custom app
			if !destinationInput.CustomApp.IsNull() {
				elementsDestinationCustomAppInput := make([]types.Object, 0, len(destinationInput.CustomApp.Elements()))
				diags = destinationInput.CustomApp.ElementsAs(ctx, &elementsDestinationCustomAppInput, false)
				resp.Diagnostics.Append(diags...)

				var itemDestinationCustomAppInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_CustomApp
				for _, item := range elementsDestinationCustomAppInput {
					diags = item.As(ctx, &itemDestinationCustomAppInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationCustomAppInput)
					if err != nil {
						resp.Diagnostics.AddError(
							"Object Ref transformation failed",
							err.Error(),
						)
						return
					}

					input.Rule.Destination.CustomApp = append(input.Rule.Destination.CustomApp, &cato_models.CustomApplicationRefInput{
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

				var itemDestinationIPRangeInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_IPRange
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

				var itemDestinationGlobalIPRangeInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_GlobalIPRange
				for _, item := range elementsDestinationGlobalIPRangeInput {
					diags = item.As(ctx, &itemDestinationGlobalIPRangeInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationGlobalIPRangeInput)
					if err != nil {
						resp.Diagnostics.AddError(
							"Object Ref transformation failed",
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

			// setting destination app category
			if !destinationInput.AppCategory.IsNull() {
				elementsDestinationAppCategoryInput := make([]types.Object, 0, len(destinationInput.AppCategory.Elements()))
				diags = destinationInput.AppCategory.ElementsAs(ctx, &elementsDestinationAppCategoryInput, false)
				resp.Diagnostics.Append(diags...)

				var itemDestinationAppCategoryInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_AppCategory
				for _, item := range elementsDestinationAppCategoryInput {
					diags = item.As(ctx, &itemDestinationAppCategoryInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationAppCategoryInput)
					if err != nil {
						resp.Diagnostics.AddError(
							"Object Ref transformation failed",
							err.Error(),
						)
						return
					}

					input.Rule.Destination.AppCategory = append(input.Rule.Destination.AppCategory, &cato_models.ApplicationCategoryRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
			}

			// setting destination custom app category
			if !destinationInput.CustomCategory.IsNull() {
				elementsDestinationCustomCategoryInput := make([]types.Object, 0, len(destinationInput.CustomCategory.Elements()))
				diags = destinationInput.CustomCategory.ElementsAs(ctx, &elementsDestinationCustomCategoryInput, false)
				resp.Diagnostics.Append(diags...)

				var itemDestinationCustomCategoryInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_CustomCategory
				for _, item := range elementsDestinationCustomCategoryInput {
					diags = item.As(ctx, &itemDestinationCustomCategoryInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationCustomCategoryInput)
					if err != nil {
						resp.Diagnostics.AddError(
							"Object Ref transformation failed",
							err.Error(),
						)
						return
					}

					input.Rule.Destination.CustomCategory = append(input.Rule.Destination.CustomCategory, &cato_models.CustomCategoryRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
			}

			// setting destination sanctionned apps category
			if !destinationInput.SanctionedAppsCategory.IsNull() {
				elementsDestinationSanctionedAppsCategoryInput := make([]types.Object, 0, len(destinationInput.SanctionedAppsCategory.Elements()))
				diags = destinationInput.SanctionedAppsCategory.ElementsAs(ctx, &elementsDestinationSanctionedAppsCategoryInput, false)
				resp.Diagnostics.Append(diags...)

				var itemDestinationSanctionedAppsCategoryInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_SanctionedAppsCategory
				for _, item := range elementsDestinationSanctionedAppsCategoryInput {
					diags = item.As(ctx, &itemDestinationSanctionedAppsCategoryInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationSanctionedAppsCategoryInput)
					if err != nil {
						resp.Diagnostics.AddError(
							"Object Ref transformation failed",
							err.Error(),
						)
						return
					}

					input.Rule.Destination.SanctionedAppsCategory = append(input.Rule.Destination.SanctionedAppsCategory, &cato_models.SanctionedAppsCategoryRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
			}

			// setting destination country
			if !destinationInput.Country.IsNull() {
				elementsDestinationCountryInput := make([]types.Object, 0, len(destinationInput.Country.Elements()))
				diags = destinationInput.Country.ElementsAs(ctx, &elementsDestinationCountryInput, false)
				resp.Diagnostics.Append(diags...)

				var itemDestinationCountryInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_Country
				for _, item := range elementsDestinationCountryInput {
					diags = item.As(ctx, &itemDestinationCountryInput, basetypes.ObjectAsOptions{})
					resp.Diagnostics.Append(diags...)

					ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationCountryInput)
					if err != nil {
						resp.Diagnostics.AddError(
							"Object Ref transformation failed",
							err.Error(),
						)
						return
					}

					input.Rule.Destination.Country = append(input.Rule.Destination.Country, &cato_models.CountryRefInput{
						By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
						Input: ObjectRefOutput.Input,
					})
				}
			}
		}

		// setting service
		if !ruleInput.Service.IsNull() {
			input.Rule.Service = &cato_models.InternetFirewallServiceTypeInput{}
			serviceInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service{}
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

				var itemServiceStandardInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service_Standard
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

				var itemServiceCustomInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service_Custom
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
						var itemPortRange Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service_Custom_PortRange
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
				Alert: &cato_models.PolicyRuleTrackingAlertInput{
					Enabled:   false,
					Frequency: "DAILY",
				},
			}

			trackingInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking{}
			diags = ruleInput.Tracking.As(ctx, &trackingInput, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			if !trackingInput.Event.IsNull() {
				// setting tracking event
				trackingEventInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Event{}
				diags = trackingInput.Event.As(ctx, &trackingEventInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
				input.Rule.Tracking.Event.Enabled = trackingEventInput.Enabled.ValueBool()
			}

			if !trackingInput.Alert.IsNull() {

				input.Rule.Tracking.Alert = &cato_models.PolicyRuleTrackingAlertInput{}

				trackingAlertInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Alert{}
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

					var itemAlertSubscriptionGroupInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Alert_SubscriptionGroup
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

						var itemAlertWebHookInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Alert_SubscriptionGroup
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

					var itemAlertMailingListInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Alert_SubscriptionGroup
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

			scheduleInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Schedule{}
			diags = ruleInput.Schedule.As(ctx, &scheduleInput, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			input.Rule.Schedule.ActiveOn = cato_models.PolicyActiveOnEnum(scheduleInput.ActiveOn.ValueString())

			// setting schedule custome time frame
			if !scheduleInput.CustomTimeframe.IsNull() {
				input.Rule.Schedule.CustomTimeframe = &cato_models.PolicyCustomTimeframeInput{}

				customeTimeFrameInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Schedule_CustomTimeframe{}
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

				customRecurringInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Schedule_CustomRecurring{}
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
			var itemExceptionsInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Exceptions
			for _, item := range elementsExceptionsInput {

				exceptionInput := cato_models.InternetFirewallRuleExceptionInput{}

				diags = item.As(ctx, &itemExceptionsInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				// setting exception name
				exceptionInput.Name = itemExceptionsInput.Name.ValueString()

				// setting exception connection origin
				if !itemExceptionsInput.ConnectionOrigin.IsNull() {
					exceptionInput.ConnectionOrigin = cato_models.ConnectionOriginEnum(itemExceptionsInput.ConnectionOrigin.ValueString())
				} else {
					exceptionInput.ConnectionOrigin = cato_models.ConnectionOriginEnum("ANY")
				}

				// setting source
				if !itemExceptionsInput.Source.IsNull() {

					exceptionInput.Source = &cato_models.InternetFirewallSourceInput{}
					sourceInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source{}
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

						var itemSourceHostInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_Host
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

						var itemSourceSiteInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_Site
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

						var itemSourceIPRangeInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_IPRange
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

						var itemSourceGlobalIPRangeInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_GlobalIPRange
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

						var itemSourceNetworkInterfaceInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_NetworkInterface
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

						var itemSourceSiteNetworkSubnetInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_SiteNetworkSubnet
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

						var itemSourceFloatingSubnetInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_FloatingSubnet
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

						var itemSourceUserInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_User
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

						var itemSourceUsersGroupInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_UsersGroup
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

						var itemSourceGroupInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_Group
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

						var itemSourceSystemGroupInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_SystemGroup
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

					var itemCountryInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Country
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

					var itemDeviceInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Device
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

					exceptionInput.Destination = &cato_models.InternetFirewallDestinationInput{}
					destinationInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination{}
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

					// setting destination domain
					if !destinationInput.Domain.IsNull() {
						diags = destinationInput.Domain.ElementsAs(ctx, &exceptionInput.Destination.Domain, false)
						resp.Diagnostics.Append(diags...)
					}

					// setting destination fqdn
					if !destinationInput.Fqdn.IsNull() {
						diags = destinationInput.Fqdn.ElementsAs(ctx, &exceptionInput.Destination.Fqdn, false)
						resp.Diagnostics.Append(diags...)
					}

					// setting destination remote asn
					if !destinationInput.RemoteAsn.IsNull() {
						diags = destinationInput.RemoteAsn.ElementsAs(ctx, &exceptionInput.Destination.RemoteAsn, false)
						resp.Diagnostics.Append(diags...)
					}

					// setting destination application
					if !destinationInput.Application.IsNull() {
						elementsDestinationApplicationInput := make([]types.Object, 0, len(destinationInput.Application.Elements()))
						diags = destinationInput.Application.ElementsAs(ctx, &elementsDestinationApplicationInput, false)
						resp.Diagnostics.Append(diags...)

						var itemDestinationApplicationInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_Application
						for _, item := range elementsDestinationApplicationInput {
							diags = item.As(ctx, &itemDestinationApplicationInput, basetypes.ObjectAsOptions{})
							resp.Diagnostics.Append(diags...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationApplicationInput)
							if err != nil {
								resp.Diagnostics.AddError(
									"Object Ref transformation failed",
									err.Error(),
								)
								return
							}

							exceptionInput.Destination.Application = append(exceptionInput.Destination.Application, &cato_models.ApplicationRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting destination custom app
					if !destinationInput.CustomApp.IsNull() {
						elementsDestinationCustomAppInput := make([]types.Object, 0, len(destinationInput.CustomApp.Elements()))
						diags = destinationInput.CustomApp.ElementsAs(ctx, &elementsDestinationCustomAppInput, false)
						resp.Diagnostics.Append(diags...)

						var itemDestinationCustomAppInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_CustomApp
						for _, item := range elementsDestinationCustomAppInput {
							diags = item.As(ctx, &itemDestinationCustomAppInput, basetypes.ObjectAsOptions{})
							resp.Diagnostics.Append(diags...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationCustomAppInput)
							if err != nil {
								resp.Diagnostics.AddError(
									"Object Ref transformation failed",
									err.Error(),
								)
								return
							}

							exceptionInput.Destination.CustomApp = append(exceptionInput.Destination.CustomApp, &cato_models.CustomApplicationRefInput{
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

						var itemDestinationIPRangeInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_IPRange
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

						var itemDestinationGlobalIPRangeInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_GlobalIPRange
						for _, item := range elementsDestinationGlobalIPRangeInput {
							diags = item.As(ctx, &itemDestinationGlobalIPRangeInput, basetypes.ObjectAsOptions{})
							resp.Diagnostics.Append(diags...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationGlobalIPRangeInput)
							if err != nil {
								resp.Diagnostics.AddError(
									"Object Ref transformation failed",
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

					// setting destination app category
					if !destinationInput.AppCategory.IsNull() {
						elementsDestinationAppCategoryInput := make([]types.Object, 0, len(destinationInput.AppCategory.Elements()))
						diags = destinationInput.AppCategory.ElementsAs(ctx, &elementsDestinationAppCategoryInput, false)
						resp.Diagnostics.Append(diags...)

						var itemDestinationAppCategoryInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_AppCategory
						for _, item := range elementsDestinationAppCategoryInput {
							diags = item.As(ctx, &itemDestinationAppCategoryInput, basetypes.ObjectAsOptions{})
							resp.Diagnostics.Append(diags...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationAppCategoryInput)
							if err != nil {
								resp.Diagnostics.AddError(
									"Object Ref transformation failed",
									err.Error(),
								)
								return
							}

							exceptionInput.Destination.AppCategory = append(exceptionInput.Destination.AppCategory, &cato_models.ApplicationCategoryRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting destination custom app category
					if !destinationInput.CustomCategory.IsNull() {
						elementsDestinationCustomCategoryInput := make([]types.Object, 0, len(destinationInput.CustomCategory.Elements()))
						diags = destinationInput.CustomCategory.ElementsAs(ctx, &elementsDestinationCustomCategoryInput, false)
						resp.Diagnostics.Append(diags...)

						var itemDestinationCustomCategoryInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_CustomCategory
						for _, item := range elementsDestinationCustomCategoryInput {
							diags = item.As(ctx, &itemDestinationCustomCategoryInput, basetypes.ObjectAsOptions{})
							resp.Diagnostics.Append(diags...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationCustomCategoryInput)
							if err != nil {
								resp.Diagnostics.AddError(
									"Object Ref transformation failed",
									err.Error(),
								)
								return
							}

							exceptionInput.Destination.CustomCategory = append(exceptionInput.Destination.CustomCategory, &cato_models.CustomCategoryRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting destination sanctionned apps category
					if !destinationInput.SanctionedAppsCategory.IsNull() {
						elementsDestinationSanctionedAppsCategoryInput := make([]types.Object, 0, len(destinationInput.SanctionedAppsCategory.Elements()))
						diags = destinationInput.SanctionedAppsCategory.ElementsAs(ctx, &elementsDestinationSanctionedAppsCategoryInput, false)
						resp.Diagnostics.Append(diags...)

						var itemDestinationSanctionedAppsCategoryInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_SanctionedAppsCategory
						for _, item := range elementsDestinationSanctionedAppsCategoryInput {
							diags = item.As(ctx, &itemDestinationSanctionedAppsCategoryInput, basetypes.ObjectAsOptions{})
							resp.Diagnostics.Append(diags...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationSanctionedAppsCategoryInput)
							if err != nil {
								resp.Diagnostics.AddError(
									"Object Ref transformation failed",
									err.Error(),
								)
								return
							}

							exceptionInput.Destination.SanctionedAppsCategory = append(exceptionInput.Destination.SanctionedAppsCategory, &cato_models.SanctionedAppsCategoryRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}

					// setting destination country
					if !destinationInput.Country.IsNull() {
						elementsDestinationCountryInput := make([]types.Object, 0, len(destinationInput.Country.Elements()))
						diags = destinationInput.Country.ElementsAs(ctx, &elementsDestinationCountryInput, false)
						resp.Diagnostics.Append(diags...)

						var itemDestinationCountryInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_Country
						for _, item := range elementsDestinationCountryInput {
							diags = item.As(ctx, &itemDestinationCountryInput, basetypes.ObjectAsOptions{})
							resp.Diagnostics.Append(diags...)

							ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationCountryInput)
							if err != nil {
								resp.Diagnostics.AddError(
									"Object Ref transformation failed",
									err.Error(),
								)
								return
							}

							exceptionInput.Destination.Country = append(exceptionInput.Destination.Country, &cato_models.CountryRefInput{
								By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
								Input: ObjectRefOutput.Input,
							})
						}
					}
				}

				// setting service
				if !itemExceptionsInput.Service.IsNull() {

					exceptionInput.Service = &cato_models.InternetFirewallServiceTypeInput{}
					serviceInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service{}
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

						var itemServiceStandardInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service_Standard
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

						var itemServiceCustomInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service_Custom
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
								var itemPortRange Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service_Custom_PortRange
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
		input.Rule.Action = cato_models.InternetFirewallActionEnum(ruleInput.Action.ValueString())
		if !ruleInput.ConnectionOrigin.IsNull() {
			input.Rule.ConnectionOrigin = cato_models.ConnectionOriginEnum(ruleInput.ConnectionOrigin.ValueString())
		} else {
			input.Rule.ConnectionOrigin = "ANY"
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "internet_fw_policy create", map[string]interface{}{
		"input": utils.InterfaceToJSONString(input),
	})

	//creating new rule
	policyChange, err := r.client.catov2.PolicyInternetFirewallAddRule(ctx, input, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyInternetFirewallAddRule error",
			err.Error(),
		)
		return
	}

	// check for errors
	if policyChange.Policy.InternetFirewall.AddRule.Status != "SUCCESS" {
		for _, item := range policyChange.Policy.InternetFirewall.AddRule.GetErrors() {
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
	_, err = r.client.catov2.PolicyInternetFirewallPublishPolicyRevision(ctx, &cato_models.InternetFirewallPolicyMutationInput{}, publishDataIfEnabled, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyInternetFirewallPublishPolicyRevision error",
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
		policyChange.GetPolicy().GetInternetFirewall().GetAddRule().Rule.GetRule().ID)
}

func (r *internetFwRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

	// var state InternetFirewallRule
	// diags := req.State.Get(ctx, &state)
	// resp.Diagnostics.Append(diags...)
	// if resp.Diagnostics.HasError() {
	// 	return
	// }

	// // body, err := r.client.catov2.Policy(ctx, &cato_models.InternetFirewallPolicyInput{}, &cato_models.WanFirewallPolicyInput{}, r.client.AccountId)
	// queryIfwPolicy := &cato_models.InternetFirewallPolicyInput{}
	// body, err := r.client.catov2.PolicyInternetFirewall(ctx, queryIfwPolicy, r.client.AccountId)
	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		"Catov2 API PolicyInternetFirewall error",
	// 		err.Error(),
	// 	)
	// 	return
	// }

	// //retrieve rule ID
	// curRule := Policy_Policy_InternetFirewall_Policy_Rules_Rule{}
	// diags = state.Rule.As(ctx, &curRule, basetypes.ObjectAsOptions{})
	// resp.Diagnostics.Append(diags...)
	// if resp.Diagnostics.HasError() {
	// 	return
	// }

	// ruleList := body.GetPolicy().InternetFirewall.Policy.GetRules()
	// ruleExist := false

	// diags = resp.State.Set(ctx, &state)
	// resp.Diagnostics.Append(diags...)
	// if resp.Diagnostics.HasError() {
	// 	return
	// }

	// ruleAPIResponse := &cato.Policy_Policy_InternetFirewall_Policy_Rules_Rule{}
	// // ruleAPIResponse := &r.client.catov2.Policy_Policy_InternetFirewall_Policy_Rules_Rule{}

	// for _, ruleListItem := range ruleList {
	// 	if ruleListItem.GetRule().ID == string(ruleAPIResponse.ID) {
	// 		ruleExist = true
	// 		ruleAPIResponse = ruleListItem.GetRule()
	// 	}
	// }

	// if !ruleExist {
	// 	tflog.Warn(ctx, "internet firewall rule not found, resource removed")
	// 	resp.State.RemoveResource(ctx)
	// 	return
	// }

	// // state = hydrateStateFromAPIResponse(ctx, ruleAPIResponse, state)

	// // Set the rule to state
	// if err := resp.State.SetAttribute(ctx, path.Root("rule"), curRule); err != nil {
	// 	resp.Diagnostics.AddError("Error setting rule to state", fmt.Sprintf("%s", err))
	// 	return
	// }
	// diags = resp.State.Set(ctx, &state)
	// resp.Diagnostics.Append(diags...)
	// if resp.Diagnostics.HasError() {
	// 	return
	// }

	var state InternetFirewallRule
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// body, err := r.client.catov2.Policy(ctx, &cato_models.InternetFirewallPolicyInput{}, &cato_models.WanFirewallPolicyInput{}, r.client.AccountId)
	queryIfwPolicy := &cato_models.InternetFirewallPolicyInput{}
	body, err := r.client.catov2.PolicyInternetFirewall(ctx, queryIfwPolicy, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyInternetFirewall error",
			err.Error(),
		)
		return
	}

	//retrieve rule ID
	curStateRule := Policy_Policy_InternetFirewall_Policy_Rules_Rule{}
	diags = state.Rule.As(ctx, &curStateRule, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ruleAPIResponse := &cato.Policy_Policy_InternetFirewall_Policy_Rules_Rule{}
	ruleList := body.GetPolicy().InternetFirewall.Policy.GetRules()
	ruleExist := false
	for _, ruleListItem := range ruleList {
		if ruleListItem.GetRule().ID == curStateRule.ID.ValueString() {
			ruleExist = true
			ruleAPIResponse = ruleListItem.GetRule()
		}
	}

	// remove resource if it doesn't exist anymore
	if !ruleExist {
		tflog.Warn(ctx, "internet firewall rule not found, resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	state, diags1 := hydrateStateFromAPIResponse(ctx, ruleAPIResponse, state)
	// diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags1...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.SetAttribute(ctx, path.Root("rule"), state.Rule)

	// diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	////////////// end rule.exceptions ///////////////

}

func (r *internetFwRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	var plan InternetFirewallRule
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// setting input for update rule
	input := cato_models.InternetFirewallUpdateRuleInput{
		Rule: &cato_models.InternetFirewallUpdateRuleDataInput{
			Source: &cato_models.InternetFirewallSourceUpdateInput{
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
			Country:  []*cato_models.CountryRefInput{},
			Device:   []*cato_models.DeviceProfileRefInput{},
			DeviceOs: []cato_models.OperatingSystem{},
			Destination: &cato_models.InternetFirewallDestinationUpdateInput{
				Application:            []*cato_models.ApplicationRefInput{},
				CustomApp:              []*cato_models.CustomApplicationRefInput{},
				AppCategory:            []*cato_models.ApplicationCategoryRefInput{},
				CustomCategory:         []*cato_models.CustomCategoryRefInput{},
				SanctionedAppsCategory: []*cato_models.SanctionedAppsCategoryRefInput{},
				Country:                []*cato_models.CountryRefInput{},
				Domain:                 []string{},
				Fqdn:                   []string{},
				IP:                     []string{},
				Subnet:                 []string{},
				IPRange:                []*cato_models.IPAddressRangeInput{},
				GlobalIPRange:          []*cato_models.GlobalIPRangeRefInput{},
				RemoteAsn:              []scalars.Asn16{},
			},
			Service: &cato_models.InternetFirewallServiceTypeUpdateInput{
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
			Exceptions: []*cato_models.InternetFirewallRuleExceptionInput{},
		},
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

	// setting rule
	ruleInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule{}
	diags = plan.Rule.As(ctx, &ruleInput, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)

	// setting source
	if !ruleInput.Source.IsNull() {
		sourceInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source{}
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

			var itemSourceHostInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_Host
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

			var itemSourceSiteInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_Site
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

			var itemSourceIPRangeInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_IPRange
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

			var itemSourceGlobalIPRangeInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_GlobalIPRange
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

			var itemSourceNetworkInterfaceInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_NetworkInterface
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

			var itemSourceSiteNetworkSubnetInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_SiteNetworkSubnet
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

			var itemSourceFloatingSubnetInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_FloatingSubnet
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

			var itemSourceUserInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_User
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

			var itemSourceUsersGroupInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_UsersGroup
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

			var itemSourceGroupInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_Group
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

			var itemSourceSystemGroupInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_SystemGroup
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

		var itemCountryInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Country
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

		var itemDeviceInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Device
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
		destinationInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination{}
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

		// setting destination domain
		if !destinationInput.Domain.IsNull() {
			diags = destinationInput.Domain.ElementsAs(ctx, &input.Rule.Destination.Domain, false)
			resp.Diagnostics.Append(diags...)
		}

		// setting destination fqdn
		if !destinationInput.Fqdn.IsNull() {
			diags = destinationInput.Fqdn.ElementsAs(ctx, &input.Rule.Destination.Fqdn, false)
			resp.Diagnostics.Append(diags...)
		}

		// setting destination remote asn
		if !destinationInput.RemoteAsn.IsNull() {
			diags = destinationInput.RemoteAsn.ElementsAs(ctx, &input.Rule.Destination.RemoteAsn, false)
			resp.Diagnostics.Append(diags...)
		}

		// setting destination application
		if !destinationInput.Application.IsNull() {
			elementsDestinationApplicationInput := make([]types.Object, 0, len(destinationInput.Application.Elements()))
			diags = destinationInput.Application.ElementsAs(ctx, &elementsDestinationApplicationInput, false)
			resp.Diagnostics.Append(diags...)

			var itemDestinationApplicationInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_Application
			for _, item := range elementsDestinationApplicationInput {
				diags = item.As(ctx, &itemDestinationApplicationInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationApplicationInput)
				if err != nil {
					resp.Diagnostics.AddError(
						"Object Ref transformation failed",
						err.Error(),
					)
					return
				}

				input.Rule.Destination.Application = append(input.Rule.Destination.Application, &cato_models.ApplicationRefInput{
					By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
					Input: ObjectRefOutput.Input,
				})
			}
		}

		// setting destination custom app
		if !destinationInput.CustomApp.IsNull() {
			elementsDestinationCustomAppInput := make([]types.Object, 0, len(destinationInput.CustomApp.Elements()))
			diags = destinationInput.CustomApp.ElementsAs(ctx, &elementsDestinationCustomAppInput, false)
			resp.Diagnostics.Append(diags...)

			var itemDestinationCustomAppInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_CustomApp
			for _, item := range elementsDestinationCustomAppInput {
				diags = item.As(ctx, &itemDestinationCustomAppInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationCustomAppInput)
				if err != nil {
					resp.Diagnostics.AddError(
						"Object Ref transformation failed",
						err.Error(),
					)
					return
				}

				input.Rule.Destination.CustomApp = append(input.Rule.Destination.CustomApp, &cato_models.CustomApplicationRefInput{
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

			var itemDestinationIPRangeInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_IPRange
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

			var itemDestinationGlobalIPRangeInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_GlobalIPRange
			for _, item := range elementsDestinationGlobalIPRangeInput {
				diags = item.As(ctx, &itemDestinationGlobalIPRangeInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationGlobalIPRangeInput)
				if err != nil {
					resp.Diagnostics.AddError(
						"Object Ref transformation failed",
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

		// setting destination app category
		if !destinationInput.AppCategory.IsNull() {
			elementsDestinationAppCategoryInput := make([]types.Object, 0, len(destinationInput.AppCategory.Elements()))
			diags = destinationInput.AppCategory.ElementsAs(ctx, &elementsDestinationAppCategoryInput, false)
			resp.Diagnostics.Append(diags...)

			var itemDestinationAppCategoryInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_AppCategory
			for _, item := range elementsDestinationAppCategoryInput {
				diags = item.As(ctx, &itemDestinationAppCategoryInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationAppCategoryInput)
				if err != nil {
					resp.Diagnostics.AddError(
						"Object Ref transformation failed",
						err.Error(),
					)
					return
				}

				input.Rule.Destination.AppCategory = append(input.Rule.Destination.AppCategory, &cato_models.ApplicationCategoryRefInput{
					By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
					Input: ObjectRefOutput.Input,
				})
			}
		}

		// setting destination custom app category
		if !destinationInput.CustomCategory.IsNull() {
			elementsDestinationCustomCategoryInput := make([]types.Object, 0, len(destinationInput.CustomCategory.Elements()))
			diags = destinationInput.CustomCategory.ElementsAs(ctx, &elementsDestinationCustomCategoryInput, false)
			resp.Diagnostics.Append(diags...)

			var itemDestinationCustomCategoryInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_CustomCategory
			for _, item := range elementsDestinationCustomCategoryInput {
				diags = item.As(ctx, &itemDestinationCustomCategoryInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationCustomCategoryInput)
				if err != nil {
					resp.Diagnostics.AddError(
						"Object Ref transformation failed",
						err.Error(),
					)
					return
				}

				input.Rule.Destination.CustomCategory = append(input.Rule.Destination.CustomCategory, &cato_models.CustomCategoryRefInput{
					By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
					Input: ObjectRefOutput.Input,
				})
			}
		}

		// setting destination sanctionned apps category
		if !destinationInput.SanctionedAppsCategory.IsNull() {
			elementsDestinationSanctionedAppsCategoryInput := make([]types.Object, 0, len(destinationInput.SanctionedAppsCategory.Elements()))
			diags = destinationInput.SanctionedAppsCategory.ElementsAs(ctx, &elementsDestinationSanctionedAppsCategoryInput, false)
			resp.Diagnostics.Append(diags...)

			var itemDestinationSanctionedAppsCategoryInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_SanctionedAppsCategory
			for _, item := range elementsDestinationSanctionedAppsCategoryInput {
				diags = item.As(ctx, &itemDestinationSanctionedAppsCategoryInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationSanctionedAppsCategoryInput)
				if err != nil {
					resp.Diagnostics.AddError(
						"Object Ref transformation failed",
						err.Error(),
					)
					return
				}

				input.Rule.Destination.SanctionedAppsCategory = append(input.Rule.Destination.SanctionedAppsCategory, &cato_models.SanctionedAppsCategoryRefInput{
					By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
					Input: ObjectRefOutput.Input,
				})
			}
		}

		// setting destination country
		if !destinationInput.Country.IsNull() {
			elementsDestinationCountryInput := make([]types.Object, 0, len(destinationInput.Country.Elements()))
			diags = destinationInput.Country.ElementsAs(ctx, &elementsDestinationCountryInput, false)
			resp.Diagnostics.Append(diags...)

			var itemDestinationCountryInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_Country
			for _, item := range elementsDestinationCountryInput {
				diags = item.As(ctx, &itemDestinationCountryInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)

				ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationCountryInput)
				if err != nil {
					resp.Diagnostics.AddError(
						"Object Ref transformation failed",
						err.Error(),
					)
					return
				}

				input.Rule.Destination.Country = append(input.Rule.Destination.Country, &cato_models.CountryRefInput{
					By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
					Input: ObjectRefOutput.Input,
				})
			}
		}
	}

	// setting service
	if !ruleInput.Service.IsNull() {

		serviceInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service{}

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

			var itemServiceStandardInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service_Standard
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

			var itemServiceCustomInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service_Custom
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
					var itemPortRange Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service_Custom_PortRange
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

		trackingInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking{}
		diags = ruleInput.Tracking.As(ctx, &trackingInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// setting tracking event
		trackingEventInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Event{}
		diags = trackingInput.Event.As(ctx, &trackingEventInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		input.Rule.Tracking.Event.Enabled = trackingEventInput.Enabled.ValueBoolPointer()

		// setting tracking Alert
		defaultAlert := false
		input.Rule.Tracking.Alert = &cato_models.PolicyRuleTrackingAlertUpdateInput{
			Enabled: &defaultAlert,
		}
		if !trackingInput.Alert.IsNull() {

			trackingAlertInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Alert{}
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

				var itemAlertSubscriptionGroupInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Alert_SubscriptionGroup
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

					var itemAlertWebHookInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Alert_SubscriptionGroup
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

				var itemAlertMailingListInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Alert_SubscriptionGroup
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

		scheduleInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Schedule{}
		diags = ruleInput.Schedule.As(ctx, &scheduleInput, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		input.Rule.Schedule.ActiveOn = (*cato_models.PolicyActiveOnEnum)(scheduleInput.ActiveOn.ValueStringPointer())

		// setting schedule custome time frame
		if !scheduleInput.CustomTimeframe.IsNull() {
			input.Rule.Schedule.CustomTimeframe = &cato_models.PolicyCustomTimeframeUpdateInput{}

			customeTimeFrameInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Schedule_CustomTimeframe{}
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

			customRecurringInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Schedule_CustomRecurring{}
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
		var itemExceptionsInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Exceptions
		for _, item := range elementsExceptionsInput {

			exceptionInput := cato_models.InternetFirewallRuleExceptionInput{}

			diags = item.As(ctx, &itemExceptionsInput, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)

			// setting exception name
			exceptionInput.Name = itemExceptionsInput.Name.ValueString()

			// setting exception connection origin
			if !itemExceptionsInput.ConnectionOrigin.IsNull() {
				exceptionInput.ConnectionOrigin = cato_models.ConnectionOriginEnum(itemExceptionsInput.ConnectionOrigin.ValueString())
			} else {
				exceptionInput.ConnectionOrigin = cato_models.ConnectionOriginEnum("ANY")
			}

			// setting source
			if !itemExceptionsInput.Source.IsNull() {

				exceptionInput.Source = &cato_models.InternetFirewallSourceInput{}
				sourceInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source{}
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

					var itemSourceHostInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_Host
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

					var itemSourceSiteInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_Site
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

					var itemSourceIPRangeInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_IPRange
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

					var itemSourceGlobalIPRangeInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_GlobalIPRange
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

					var itemSourceNetworkInterfaceInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_NetworkInterface
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

					var itemSourceSiteNetworkSubnetInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_SiteNetworkSubnet
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

					var itemSourceFloatingSubnetInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_FloatingSubnet
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

					var itemSourceUserInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_User
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

					var itemSourceUsersGroupInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_UsersGroup
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

					var itemSourceGroupInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_Group
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

					var itemSourceSystemGroupInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source_SystemGroup
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

				var itemCountryInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Country
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

				var itemDeviceInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Device
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

				exceptionInput.Destination = &cato_models.InternetFirewallDestinationInput{}
				destinationInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination{}
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

				// setting destination domain
				if !destinationInput.Domain.IsNull() {
					diags = destinationInput.Domain.ElementsAs(ctx, &exceptionInput.Destination.Domain, false)
					resp.Diagnostics.Append(diags...)
				}

				// setting destination fqdn
				if !destinationInput.Fqdn.IsNull() {
					diags = destinationInput.Fqdn.ElementsAs(ctx, &exceptionInput.Destination.Fqdn, false)
					resp.Diagnostics.Append(diags...)
				}

				// setting destination remote asn
				if !destinationInput.RemoteAsn.IsNull() {
					diags = destinationInput.RemoteAsn.ElementsAs(ctx, &exceptionInput.Destination.RemoteAsn, false)
					resp.Diagnostics.Append(diags...)
				}

				// setting destination application
				if !destinationInput.Application.IsNull() {
					elementsDestinationApplicationInput := make([]types.Object, 0, len(destinationInput.Application.Elements()))
					diags = destinationInput.Application.ElementsAs(ctx, &elementsDestinationApplicationInput, false)
					resp.Diagnostics.Append(diags...)

					var itemDestinationApplicationInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_Application
					for _, item := range elementsDestinationApplicationInput {
						diags = item.As(ctx, &itemDestinationApplicationInput, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationApplicationInput)
						if err != nil {
							resp.Diagnostics.AddError(
								"Object Ref transformation failed",
								err.Error(),
							)
							return
						}

						exceptionInput.Destination.Application = append(exceptionInput.Destination.Application, &cato_models.ApplicationRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
				}

				// setting destination custom app
				if !destinationInput.CustomApp.IsNull() {
					elementsDestinationCustomAppInput := make([]types.Object, 0, len(destinationInput.CustomApp.Elements()))
					diags = destinationInput.CustomApp.ElementsAs(ctx, &elementsDestinationCustomAppInput, false)
					resp.Diagnostics.Append(diags...)

					var itemDestinationCustomAppInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_CustomApp
					for _, item := range elementsDestinationCustomAppInput {
						diags = item.As(ctx, &itemDestinationCustomAppInput, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationCustomAppInput)
						if err != nil {
							resp.Diagnostics.AddError(
								"Object Ref transformation failed",
								err.Error(),
							)
							return
						}

						exceptionInput.Destination.CustomApp = append(exceptionInput.Destination.CustomApp, &cato_models.CustomApplicationRefInput{
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

					var itemDestinationIPRangeInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_IPRange
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

					var itemDestinationGlobalIPRangeInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_GlobalIPRange
					for _, item := range elementsDestinationGlobalIPRangeInput {
						diags = item.As(ctx, &itemDestinationGlobalIPRangeInput, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationGlobalIPRangeInput)
						if err != nil {
							resp.Diagnostics.AddError(
								"Object Ref transformation failed",
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

				// setting destination app category
				if !destinationInput.AppCategory.IsNull() {
					elementsDestinationAppCategoryInput := make([]types.Object, 0, len(destinationInput.AppCategory.Elements()))
					diags = destinationInput.AppCategory.ElementsAs(ctx, &elementsDestinationAppCategoryInput, false)
					resp.Diagnostics.Append(diags...)

					var itemDestinationAppCategoryInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_AppCategory
					for _, item := range elementsDestinationAppCategoryInput {
						diags = item.As(ctx, &itemDestinationAppCategoryInput, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationAppCategoryInput)
						if err != nil {
							resp.Diagnostics.AddError(
								"Object Ref transformation failed",
								err.Error(),
							)
							return
						}

						exceptionInput.Destination.AppCategory = append(exceptionInput.Destination.AppCategory, &cato_models.ApplicationCategoryRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
				}

				// setting destination custom app category
				if !destinationInput.CustomCategory.IsNull() {
					elementsDestinationCustomCategoryInput := make([]types.Object, 0, len(destinationInput.CustomCategory.Elements()))
					diags = destinationInput.CustomCategory.ElementsAs(ctx, &elementsDestinationCustomCategoryInput, false)
					resp.Diagnostics.Append(diags...)

					var itemDestinationCustomCategoryInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_CustomCategory
					for _, item := range elementsDestinationCustomCategoryInput {
						diags = item.As(ctx, &itemDestinationCustomCategoryInput, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationCustomCategoryInput)
						if err != nil {
							resp.Diagnostics.AddError(
								"Object Ref transformation failed",
								err.Error(),
							)
							return
						}

						exceptionInput.Destination.CustomCategory = append(exceptionInput.Destination.CustomCategory, &cato_models.CustomCategoryRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
				}

				// setting destination sanctionned apps category
				if !destinationInput.SanctionedAppsCategory.IsNull() {
					elementsDestinationSanctionedAppsCategoryInput := make([]types.Object, 0, len(destinationInput.SanctionedAppsCategory.Elements()))
					diags = destinationInput.SanctionedAppsCategory.ElementsAs(ctx, &elementsDestinationSanctionedAppsCategoryInput, false)
					resp.Diagnostics.Append(diags...)

					var itemDestinationSanctionedAppsCategoryInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_SanctionedAppsCategory
					for _, item := range elementsDestinationSanctionedAppsCategoryInput {
						diags = item.As(ctx, &itemDestinationSanctionedAppsCategoryInput, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationSanctionedAppsCategoryInput)
						if err != nil {
							resp.Diagnostics.AddError(
								"Object Ref transformation failed",
								err.Error(),
							)
							return
						}

						exceptionInput.Destination.SanctionedAppsCategory = append(exceptionInput.Destination.SanctionedAppsCategory, &cato_models.SanctionedAppsCategoryRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
				}

				// setting destination country
				if !destinationInput.Country.IsNull() {
					elementsDestinationCountryInput := make([]types.Object, 0, len(destinationInput.Country.Elements()))
					diags = destinationInput.Country.ElementsAs(ctx, &elementsDestinationCountryInput, false)
					resp.Diagnostics.Append(diags...)

					var itemDestinationCountryInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination_Country
					for _, item := range elementsDestinationCountryInput {
						diags = item.As(ctx, &itemDestinationCountryInput, basetypes.ObjectAsOptions{})
						resp.Diagnostics.Append(diags...)

						ObjectRefOutput, err := utils.TransformObjectRefInput(itemDestinationCountryInput)
						if err != nil {
							resp.Diagnostics.AddError(
								"Object Ref transformation failed",
								err.Error(),
							)
							return
						}

						exceptionInput.Destination.Country = append(exceptionInput.Destination.Country, &cato_models.CountryRefInput{
							By:    cato_models.ObjectRefBy(ObjectRefOutput.By),
							Input: ObjectRefOutput.Input,
						})
					}
				}
			}

			// setting service
			if !itemExceptionsInput.Service.IsNull() {

				exceptionInput.Service = &cato_models.InternetFirewallServiceTypeInput{}
				serviceInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service{}
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

					var itemServiceStandardInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service_Standard
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

					var itemServiceCustomInput Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service_Custom
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
							var itemPortRange Policy_Policy_InternetFirewall_Policy_Rules_Rule_Service_Custom_PortRange
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
	inputMoveRule.ID = *ruleInput.ID.ValueStringPointer()
	input.ID = *ruleInput.ID.ValueStringPointer()
	input.Rule.Name = ruleInput.Name.ValueStringPointer()
	input.Rule.Description = ruleInput.Description.ValueStringPointer()
	input.Rule.Enabled = ruleInput.Enabled.ValueBoolPointer()
	input.Rule.Action = (*cato_models.InternetFirewallActionEnum)(ruleInput.Action.ValueStringPointer())
	if !ruleInput.ConnectionOrigin.IsNull() {
		input.Rule.ConnectionOrigin = (*cato_models.ConnectionOriginEnum)(ruleInput.ConnectionOrigin.ValueStringPointer())
	} else {
		connectionOrigin := "ANY"
		input.Rule.ConnectionOrigin = (*cato_models.ConnectionOriginEnum)(&connectionOrigin)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "internet_fw_rule move", map[string]interface{}{
		"input": utils.InterfaceToJSONString(inputMoveRule),
	})

	//move rule
	moveRule, err := r.client.catov2.PolicyInternetFirewallMoveRule(ctx, &cato_models.InternetFirewallPolicyMutationInput{}, inputMoveRule, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyInternetFirewallMoveRule error",
			err.Error(),
		)
		return
	}

	// check for errors
	if moveRule.Policy.InternetFirewall.MoveRule.Status != "SUCCESS" {
		for _, item := range moveRule.Policy.InternetFirewall.MoveRule.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Moving Rule Resource",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	tflog.Debug(ctx, "internet_fw_rule update", map[string]interface{}{
		"input": utils.InterfaceToJSONString(input),
	})

	//Update new rule
	updateRule, err := r.client.catov2.PolicyInternetFirewallUpdateRule(ctx, &cato_models.InternetFirewallPolicyMutationInput{}, input, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyInternetFirewallUpdateRule error",
			err.Error(),
		)
		return
	}

	// check for errors
	if updateRule.Policy.InternetFirewall.UpdateRule.Status != "SUCCESS" {
		for _, item := range updateRule.Policy.InternetFirewall.UpdateRule.GetErrors() {
			resp.Diagnostics.AddError(
				"API Error Update Resource",
				fmt.Sprintf("%s : %s", *item.ErrorCode, *item.ErrorMessage),
			)
		}
		return
	}

	//publishing new rule
	tflog.Info(ctx, "publishing new rule")
	publishDataIfEnabled := &cato_models.PolicyPublishRevisionInput{}
	_, err = r.client.catov2.PolicyInternetFirewallPublishPolicyRevision(ctx, &cato_models.InternetFirewallPolicyMutationInput{}, publishDataIfEnabled, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyInternetFirewallPublishPolicyRevision error",
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

func (r *internetFwRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	var state InternetFirewallRule
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	//retrieve rule ID
	rule := Policy_Policy_InternetFirewall_Policy_Rules_Rule{}
	diags = state.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	removeMutations := &cato_models.InternetFirewallPolicyMutationInput{}
	removeRule := cato_models.InternetFirewallRemoveRuleInput{
		ID: rule.ID.ValueString(),
	}
	tflog.Debug(ctx, "internet_fw_policy delete", map[string]interface{}{
		"query": utils.InterfaceToJSONString(removeMutations),
		"input": utils.InterfaceToJSONString(removeRule),
	})

	_, err := r.client.catov2.PolicyInternetFirewallRemoveRule(ctx, removeMutations, removeRule, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to connect or request the Catov2 API",
			err.Error(),
		)
		return
	}

	publishDataIfEnabled := &cato_models.PolicyPublishRevisionInput{}
	_, err = r.client.catov2.PolicyInternetFirewallPublishPolicyRevision(ctx, &cato_models.InternetFirewallPolicyMutationInput{}, publishDataIfEnabled, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API Delete/PolicyInternetFirewallPublishPolicyRevision error",
			err.Error(),
		)
		return
	}
}

// General Purpose Functions:
func hydrateStateFromAPIResponse(ctx context.Context, ruleResponse *cato.Policy_Policy_InternetFirewall_Policy_Rules_Rule, state InternetFirewallRule) (InternetFirewallRule, diag.Diagnostics) {
	diags := make(diag.Diagnostics, 0)
	curRuleObj := createNullCurRule(ctx)
	curRule := curRuleObj.Attributes()
	// Set state attributes
	// state.Rule.ID = basetypes.NewStringValue(curRule.ID)
	curRule["name"] = basetypes.NewStringValue(ruleResponse.Name)
	curRule["description"] = basetypes.NewStringValue(ruleResponse.Description)
	curRule["enabled"] = basetypes.NewBoolValue(ruleResponse.Enabled)
	curRule["action"] = basetypes.NewStringValue(ruleResponse.Action.String())

	// resp.State.SetAttribute(ctx, path.Root("rule").AtName("index"), curRule.Index)
	// resp.State.SetAttribute(ctx, path.Root("rule").AtName("section").AtName("id"), curRule.Section.ID)
	// resp.State.SetAttribute(ctx, path.Root("rule").AtName("section").AtName("name"), curRule.Section.Name)

	// //////////// rule.source ///////////////
	curRuleSourceObj, diagstmp := types.ObjectValue(
		map[string]attr.Type{
			"ip":                  types.ListType{ElemType: types.StringType},
			"host":                types.ListType{ElemType: NameIDObjectType},
			"site":                types.ListType{ElemType: NameIDObjectType},
			"subnet":              types.ListType{ElemType: types.StringType},
			"ip_range":            types.ListType{ElemType: FromToObjectType},
			"global_ip_range":     types.ListType{ElemType: NameIDObjectType},
			"network_interface":   types.ListType{ElemType: NameIDObjectType},
			"site_network_subnet": types.ListType{ElemType: NameIDObjectType},
			"floating_subnet":     types.ListType{ElemType: NameIDObjectType},
			"user":                types.ListType{ElemType: NameIDObjectType},
			"users_group":         types.ListType{ElemType: NameIDObjectType},
			"group":               types.ListType{ElemType: NameIDObjectType},
			"system_group":        types.ListType{ElemType: NameIDObjectType},
		},
		map[string]attr.Value{
			"ip":                  types.ListNull(types.StringType),
			"host":                types.ListNull(NameIDObjectType),
			"site":                types.ListNull(NameIDObjectType),
			"subnet":              types.ListNull(types.StringType),
			"ip_range":            types.ListNull(FromToObjectType),
			"global_ip_range":     types.ListNull(NameIDObjectType),
			"network_interface":   types.ListNull(NameIDObjectType),
			"site_network_subnet": types.ListNull(NameIDObjectType),
			"floating_subnet":     types.ListNull(NameIDObjectType),
			"user":                types.ListNull(NameIDObjectType),
			"users_group":         types.ListNull(NameIDObjectType),
			"group":               types.ListNull(NameIDObjectType),
			"system_group":        types.ListNull(NameIDObjectType),
		},
	)
	// if diagstmp.HasError() {
	// 	diagstmp = addDebugLineNumber(diagstmp, "rule.curRuleSourceObj")
	// }
	diags = append(diags, diagstmp...)
	// Initialize source object attributes
	curRuleSourceObjAttrs := curRuleSourceObj.Attributes()

	// rule.source.subnet[]
	curSourceSubnets := []string{}
	for _, sourceSubnet := range ruleResponse.Source.Subnet {
		curSourceSubnets = append(curSourceSubnets, sourceSubnet)
	}
	curSourceSubnetList, diagstmp := types.ListValueFrom(ctx, types.StringType, curSourceSubnets)
	// if diagstmp.HasError() {
	// 	diagstmp = addDebugLineNumber(diagstmp, "rule.curSourceSubnets")
	// }
	diags = append(diags, diagstmp...)
	curRuleSourceObjAttrs["subnet"] = curSourceSubnetList

	// rule.source.ip[]
	curSourceSourceIps := []string{}
	for _, sourceIp := range ruleResponse.Source.IP {
		curSourceSourceIps = append(curSourceSourceIps, sourceIp)
	}
	curSourceSourceIpList, diagstmp := types.ListValueFrom(ctx, types.StringType, curSourceSubnets)
	// if diagstmp.HasError() {
	// 	diagstmp = addDebugLineNumber(diagstmp, "rule.curSourceSourceIps")
	// }
	diags = append(diags, diagstmp...)
	curRuleSourceObjAttrs["ip"] = curSourceSourceIpList

	// rule.source.host[]
	var curSourceHosts []types.List
	tflog.Warn(ctx, "ruleResponse.Source.Host - "+fmt.Sprintf("%v", ruleResponse.Source.Host))
	for _, host := range ruleResponse.Source.Host {
		curSourceHostObj := parseNameIDList(ctx, host)
		curSourceHosts = append(curSourceHosts, curSourceHostObj)
	}
	curSourceHostValues := make([]attr.Value, len(curSourceHosts))
	for i, v := range curSourceHosts {
		curSourceHostValues[i] = v
	}
	curRuleSourceObjAttrs["host"], diagstmp = types.ListValue(NameIDObjectType, curSourceHostValues)
	// if diagstmp.HasError() {
	// 	diagstmp = addDebugLineNumber(diagstmp, "rule.curSourceHosts")
	// }
	diags = append(diags, diagstmp...)

	// // rule.source.site[]
	tflog.Warn(ctx, "ruleResponse.Source.Site - "+fmt.Sprintf("%v", ruleResponse.Source.Site))
	var curSourceSites []types.List
	for _, site := range ruleResponse.Source.Site {
		curSourceSiteObj := parseNameIDList(ctx, site)
		curSourceSites = append(curSourceSites, curSourceSiteObj)
	}
	curSourceSiteValues := make([]attr.Value, len(curSourceSites))
	for i, v := range curSourceSites {
		curSourceSiteValues[i] = v
	}
	curRuleSourceObjAttrs["site"], diagstmp = types.ListValue(NameIDObjectType, curSourceSiteValues)
	// if diagstmp.HasError() {
	// 	diagstmp = addDebugLineNumber(diagstmp, "rule.curSourceSites")
	// }
	diags = append(diags, diagstmp...)

	// rule.source.ip_range[]
	var curSourceIPRanges []types.List
	tflog.Warn(ctx, "ruleResponse.Source.IPRange - "+fmt.Sprintf("%v", ruleResponse.Source.IPRange))
	for _, iprange := range ruleResponse.Source.IPRange {
		curSourceIPRangeObj := parseNameIDList(ctx, iprange)
		curSourceIPRanges = append(curSourceIPRanges, curSourceIPRangeObj)
	}
	curSourceIPRangeValues := make([]attr.Value, len(curSourceIPRanges))
	for i, v := range curSourceIPRanges {
		curSourceIPRangeValues[i] = v
	}
	curRuleSourceObjAttrs["ip_range"], diagstmp = types.ListValue(NameIDObjectType, curSourceIPRangeValues)
	// if diagstmp.HasError() {
	// 	diagstmp = addDebugLineNumber(diagstmp, "rule.curSourceIPRanges")
	// }
	diags = append(diags, diagstmp...)

	// rule.source.global_ip_range[]
	var curSourceGlobalIPRanges []types.List
	tflog.Warn(ctx, "ruleResponse.Source.GlobalIPRange - "+fmt.Sprintf("%v", ruleResponse.Source.GlobalIPRange))
	for _, globaliprange := range ruleResponse.Source.GlobalIPRange {
		curSourceGlobalIPRangeObj := parseFromToList(globaliprange)
		curSourceGlobalIPRanges = append(curSourceGlobalIPRanges, curSourceGlobalIPRangeObj)
	}
	curSourceGlobalIPRangeValues := make([]attr.Value, len(curSourceGlobalIPRanges))
	for i, v := range curSourceGlobalIPRanges {
		curSourceGlobalIPRangeValues[i] = v
	}
	curRuleSourceObjAttrs["global_ip_range"], diagstmp = types.ListValue(NameIDObjectType, curSourceGlobalIPRangeValues)
	if diagstmp.HasError() {
		diagstmp = addDebugLineNumber(diagstmp, "rule.curSourceGlobalIPRanges")
	}
	diags = append(diags, diagstmp...)

	// rule.source.network_interface[]
	var curSourceNetworkInterfaces []types.List
	tflog.Warn(ctx, "ruleResponse.Source.NetworkInterface - "+fmt.Sprintf("%v", ruleResponse.Source.NetworkInterface))
	for _, networkInterface := range ruleResponse.Source.NetworkInterface {
		curSourceNetworkInterfaceObj := parseNameIDList(ctx, networkInterface)
		curSourceNetworkInterfaces = append(curSourceNetworkInterfaces, curSourceNetworkInterfaceObj)
	}
	curSourceNetworkInterfaceValues := make([]attr.Value, len(curSourceNetworkInterfaces))
	for i, v := range curSourceNetworkInterfaces {
		curSourceNetworkInterfaceValues[i] = v
	}
	curRuleSourceObjAttrs["network_interface"], diagstmp = types.ListValue(NameIDObjectType, curSourceNetworkInterfaceValues)
	// if diagstmp.HasError() {
	// 	diagstmp = addDebugLineNumber(diagstmp, "rule.curSourceNetworkInterfaces")
	// }
	diags = append(diags, diagstmp...)

	// rule.source.site_network_subnet[]
	var curSourceNetworkSubnets []types.List
	tflog.Warn(ctx, "ruleResponse.Source.SiteNetworkSubnet - "+fmt.Sprintf("%v", ruleResponse.Source.SiteNetworkSubnet))
	for _, networkSubnet := range ruleResponse.Source.SiteNetworkSubnet {
		curSourceNetworkSubnetObj := parseNameIDList(ctx, networkSubnet)
		curSourceNetworkSubnets = append(curSourceNetworkSubnets, curSourceNetworkSubnetObj)
	}
	curSourceNetworkSubnetValues := make([]attr.Value, len(curSourceNetworkSubnets))
	for i, v := range curSourceNetworkSubnets {
		curSourceNetworkSubnetValues[i] = v
	}
	curRuleSourceObjAttrs["site_network_subnet"], diagstmp = types.ListValue(NameIDObjectType, curSourceNetworkSubnetValues)
	// if diagstmp.HasError() {
	// 	diagstmp = addDebugLineNumber(diagstmp, "rule.curSourceNetworkSubnets")
	// }
	diags = append(diags, diagstmp...)

	// rule.source.floating_subnet[]
	var curSourceFloatingSubnets []types.List
	tflog.Warn(ctx, "ruleResponse.Source.FloatingSubnet - "+fmt.Sprintf("%v", ruleResponse.Source.FloatingSubnet))
	for _, floatingSubnet := range ruleResponse.Source.FloatingSubnet {
		curSourceFloatingSubnetObj := parseNameIDList(ctx, floatingSubnet)
		curSourceFloatingSubnets = append(curSourceFloatingSubnets, curSourceFloatingSubnetObj)
	}
	curSourceFloatingSubnetValues := make([]attr.Value, len(curSourceFloatingSubnets))
	for i, v := range curSourceFloatingSubnets {
		curSourceFloatingSubnetValues[i] = v
	}
	curRuleSourceObjAttrs["floating_subnet"], diagstmp = types.ListValue(NameIDObjectType, curSourceFloatingSubnetValues)
	// if diagstmp.HasError() {
	// 	diagstmp = addDebugLineNumber(diagstmp, "rule.curSourceFloatingSubnets")
	// }
	diags = append(diags, diagstmp...)

	// rule.source.user[]
	var curSourceUsers []types.List
	tflog.Warn(ctx, "ruleResponse.Source.User - "+fmt.Sprintf("%v", ruleResponse.Source.User))
	for _, user := range ruleResponse.Source.User {
		curSourceUserObj := parseNameIDList(ctx, user)
		curSourceUsers = append(curSourceUsers, curSourceUserObj)
	}
	curSourceUserValues := make([]attr.Value, len(curSourceUsers))
	for i, v := range curSourceUsers {
		curSourceUserValues[i] = v
	}
	curRuleSourceObjAttrs["user"], diagstmp = types.ListValue(NameIDObjectType, curSourceUserValues)
	// if diagstmp.HasError() {
	// 	diagstmp = addDebugLineNumber(diagstmp, "rule.curSourceUsers")
	// }
	diags = append(diags, diagstmp...)

	// rule.source.users_group[]
	var curSourceUsersGroups []types.List
	tflog.Warn(ctx, "ruleResponse.Source.UserGroup - "+fmt.Sprintf("%v", ruleResponse.Source.UsersGroup))
	for _, usersGroup := range ruleResponse.Source.UsersGroup {
		curSourceUsersGroupObj := parseNameIDList(ctx, usersGroup)
		curSourceUsersGroups = append(curSourceUsersGroups, curSourceUsersGroupObj)
	}
	curSourceUsersGroupValues := make([]attr.Value, len(curSourceUsersGroups))
	for i, v := range curSourceUsersGroups {
		curSourceUsersGroupValues[i] = v
	}
	curRuleSourceObjAttrs["users_group"], diagstmp = types.ListValue(NameIDObjectType, curSourceUsersGroupValues)
	// if diagstmp.HasError() {
	// 	diagstmp = addDebugLineNumber(diagstmp, "rule.curSourceUsersGroups")
	// }
	diags = append(diags, diagstmp...)

	// rule.source.group[]
	var curSourceGroups []types.List
	tflog.Warn(ctx, "ruleResponse.Source.Group - "+fmt.Sprintf("%v", ruleResponse.Source.Group))
	for _, group := range ruleResponse.Source.Group {
		curSourceGroupObj := parseNameIDList(ctx, group)
		curSourceGroups = append(curSourceGroups, curSourceGroupObj)
	}
	curSourceGroupValues := make([]attr.Value, len(curSourceGroups))
	for i, v := range curSourceGroups {
		curSourceGroupValues[i] = v
	}
	curRuleSourceObjAttrs["group"], diagstmp = types.ListValue(NameIDObjectType, curSourceGroupValues)
	// if diagstmp.HasError() {
	// 	diagstmp = addDebugLineNumber(diagstmp, "rule.curSourceGroups")
	// }
	diags = append(diags, diagstmp...)

	// rule.source.system_group[]
	var curSourceSystemGroups []types.List
	tflog.Warn(ctx, "ruleResponse.Source.SystemGroup - "+fmt.Sprintf("%v", ruleResponse.Source.SystemGroup))
	for _, systemGroup := range ruleResponse.Source.SystemGroup {
		curSourceSystemGroupObj := parseNameIDList(ctx, systemGroup)
		curSourceSystemGroups = append(curSourceSystemGroups, curSourceSystemGroupObj)
	}
	curSourceSystemGroupValues := make([]attr.Value, len(curSourceSystemGroups))
	for i, v := range curSourceSystemGroups {
		curSourceSystemGroupValues[i] = v
	}
	curRuleSourceObjAttrs["system_group"], diagstmp = types.ListValue(NameIDObjectType, curSourceSystemGroupValues)
	// if diagstmp.HasError() {
	// 	diagstmp = addDebugLineNumber(diagstmp, "rule.curSourceSystemGroups")
	// }
	diags = append(diags, diagstmp...)

	curRule["source"] = curRuleSourceObj
	////////////// end rule.source ///////////////

	// rule.country[]
	var curCountries []types.List
	tflog.Warn(ctx, "ruleResponse.Country - "+fmt.Sprintf("%v", ruleResponse.Country))
	for _, country := range ruleResponse.Country {
		curSourceCountryObj := parseNameIDList(ctx, country)
		curCountries = append(curCountries, curSourceCountryObj)
	}
	curCountryValues := make([]attr.Value, len(curCountries))
	for i, v := range curCountries {
		curCountryValues[i] = v
	}
	curRule["country"], diagstmp = types.ListValue(NameIDObjectType, curCountryValues)
	// if diagstmp.HasError() {
	// 	diagstmp = addDebugLineNumber(diagstmp, "rule.curCountries")
	// }
	diags = append(diags, diagstmp...)

	// rule.device[]
	var curDevices []types.List
	tflog.Warn(ctx, "ruleResponse.Device - "+fmt.Sprintf("%v", ruleResponse.Device))
	for _, device := range ruleResponse.Device {
		curDeviceObj := parseNameIDList(ctx, device)
		curDevices = append(curDevices, curDeviceObj)
	}
	curDeviceValues := make([]attr.Value, len(curDevices))
	for i, v := range curDevices {
		curDeviceValues[i] = v
	}
	curRule["device"], diagstmp = types.ListValue(NameIDObjectType, curDeviceValues)
	// if diagstmp.HasError() {
	// 	diagstmp = addDebugLineNumber(diagstmp, "rule.curDevices")
	// }
	diags = append(diags, diagstmp...)

	// rule.device_os
	curRule["deviceOs"], diagstmp = types.ListValueFrom(ctx, types.StringType, ruleResponse.DeviceOs)
	// if diagstmp.HasError() {
	// 	diagstmp = addDebugLineNumber(diagstmp, "rule.deviceOs")
	// }
	diags = append(diags, diagstmp...)

	// //////////// rule.destination ///////////////
	curRuleDestObj, diagstmp := types.ObjectValue(
		map[string]attr.Type{
			"application":              types.ListType{ElemType: NameIDObjectType},
			"custom_app":               types.ListType{ElemType: NameIDObjectType},
			"app_category":             types.ListType{ElemType: NameIDObjectType},
			"custom_category":          types.ListType{ElemType: NameIDObjectType},
			"sanctioned_apps_category": types.ListType{ElemType: NameIDObjectType},
			"country":                  types.ListType{ElemType: NameIDObjectType},
			"domain":                   types.ListType{ElemType: types.StringType},
			"fqdn":                     types.ListType{ElemType: types.StringType},
			"ip":                       types.ListType{ElemType: types.StringType},
			"subnet":                   types.ListType{ElemType: types.StringType},
			"ip_range":                 types.ListType{ElemType: FromToObjectType},
			"global_ip_range":          types.ListType{ElemType: NameIDObjectType},
			"remote_asn":               types.ListType{ElemType: types.StringType},
		},
		map[string]attr.Value{
			"application":              types.ListNull(NameIDObjectType),
			"custom_app":               types.ListNull(NameIDObjectType),
			"app_category":             types.ListNull(NameIDObjectType),
			"custom_category":          types.ListNull(NameIDObjectType),
			"sanctioned_apps_category": types.ListNull(NameIDObjectType),
			"country":                  types.ListNull(NameIDObjectType),
			"domain":                   types.ListNull(types.StringType),
			"fqdn":                     types.ListNull(types.StringType),
			"ip":                       types.ListNull(types.StringType),
			"subnet":                   types.ListNull(types.StringType),
			"ip_range":                 types.ListNull(FromToObjectType),
			"global_ip_range":          types.ListNull(NameIDObjectType),
			"remote_asn":               types.ListNull(types.StringType),
		},
	)
	// if diagstmp.HasError() {
	// 	diagstmp = addDebugLineNumber(diagstmp, "rule.curRuleDestObj")
	// }
	diags = append(diags, diagstmp...)

	// Initialize source object attributes
	curRuleDestObjAttrs := curRuleDestObj.Attributes()

	// rule.destination.application[]
	var curDestApplications []types.List
	tflog.Warn(ctx, "ruleResponse.Destination.Application - "+fmt.Sprintf("%v", ruleResponse.Destination.Application))
	for _, application := range ruleResponse.Destination.Application {
		curDestApplicationObj := parseNameIDList(ctx, application)
		curDestApplications = append(curDestApplications, curDestApplicationObj)
	}
	curDestApplicationValues := make([]attr.Value, len(curDestApplications))
	for i, v := range curDestApplications {
		curDestApplicationValues[i] = v
	}
	curRuleDestObjAttrs["application"], diagstmp = types.ListValue(NameIDObjectType, curDestApplicationValues)
	// if diagstmp.HasError() {
	// 	diagstmp = addDebugLineNumber(diagstmp, "rule.curDestApplications")
	// }
	diags = append(diags, diagstmp...)

	// rule.destination.custom_app[]
	var curDestCustomApps []types.List
	tflog.Warn(ctx, "ruleResponse.Destination.CustomApp - "+fmt.Sprintf("%v", ruleResponse.Destination.CustomApp))
	for _, customApp := range ruleResponse.Destination.CustomApp {
		curDestCustomAppObj := parseNameIDList(ctx, customApp)
		curDestCustomApps = append(curDestCustomApps, curDestCustomAppObj)
	}
	curDestCustomAppValues := make([]attr.Value, len(curDestCustomApps))
	for i, v := range curDestCustomApps {
		curDestCustomAppValues[i] = v
	}
	curRuleDestObjAttrs["custom_app"], diagstmp = types.ListValue(NameIDObjectType, curDestCustomAppValues)
	// if diagstmp.HasError() {
	// 	diagstmp = addDebugLineNumber(diagstmp, "rule.curDestCustomApps")
	// }
	diags = append(diags, diagstmp...)

	// rule.destination.app_category[]
	var curDestAppCategories []types.List
	tflog.Warn(ctx, "ruleResponse.Destination.CustomApp - "+fmt.Sprintf("%v", ruleResponse.Destination.AppCategory))
	for _, appCategory := range ruleResponse.Destination.AppCategory {
		curDestAppCategoryObj := parseNameIDList(ctx, appCategory)
		curDestAppCategories = append(curDestAppCategories, curDestAppCategoryObj)
	}
	curDestAppCategoryValues := make([]attr.Value, len(curDestAppCategories))
	for i, v := range curDestAppCategories {
		curDestAppCategoryValues[i] = v
	}
	curRuleDestObjAttrs["app_category"], diagstmp = types.ListValue(NameIDObjectType, curDestAppCategoryValues)
	// if diagstmp.HasError() {
	// 	diagstmp = addDebugLineNumber(diagstmp, "rule.curDestAppCategories")
	// }
	diags = append(diags, diagstmp...)

	// rule.destination.custom_category[]
	var curDestCustomCategories []types.List
	tflog.Warn(ctx, "ruleResponse.Destination.CustomCategory - "+fmt.Sprintf("%v", ruleResponse.Destination.CustomCategory))
	for _, customCategory := range ruleResponse.Destination.CustomCategory {
		curDestCustomCategoryObj := parseNameIDList(ctx, customCategory)
		curDestCustomCategories = append(curDestCustomCategories, curDestCustomCategoryObj)
	}
	curDestCustomCategoryValues := make([]attr.Value, len(curDestCustomCategories))
	for i, v := range curDestCustomCategories {
		curDestCustomCategoryValues[i] = v
	}
	curRuleDestObjAttrs["custom_category"], diagstmp = types.ListValue(NameIDObjectType, curDestCustomCategoryValues)
	// if diagstmp.HasError() {
	// 	diagstmp = addDebugLineNumber(diagstmp, "rule.curDestCustomCategories")
	// }
	diags = append(diags, diagstmp...)

	// rule.destination.sanctioned_apps_category[]
	var curDestSanctionedAppsCategories []types.List
	tflog.Warn(ctx, "ruleResponse.Destination.SanctionedAppsCategory - "+fmt.Sprintf("%v", ruleResponse.Destination.SanctionedAppsCategory))
	for _, sanctionedAppsCategory := range ruleResponse.Destination.SanctionedAppsCategory {
		curDestSanctionedAppsCategoryObj := parseNameIDList(ctx, sanctionedAppsCategory)
		curDestSanctionedAppsCategories = append(curDestSanctionedAppsCategories, curDestSanctionedAppsCategoryObj)
	}
	curDestSanctionedAppsCategoryValues := make([]attr.Value, len(curDestSanctionedAppsCategories))
	for i, v := range curDestSanctionedAppsCategories {
		curDestSanctionedAppsCategoryValues[i] = v
	}
	curRuleDestObjAttrs["sanctioned_apps_category"], diagstmp = types.ListValue(NameIDObjectType, curDestSanctionedAppsCategoryValues)
	// if diagstmp.HasError() {
	// 	diagstmp = addDebugLineNumber(diagstmp, "rule.curDestSanctionedAppsCategories")
	// }
	diags = append(diags, diagstmp...)

	// rule.destination.country[]
	var curDestCountries []types.List
	tflog.Warn(ctx, "ruleResponse.Destination.Country - "+fmt.Sprintf("%v", ruleResponse.Destination.Country))
	for _, country := range ruleResponse.Destination.Country {
		curDestCountryObj := parseNameIDList(ctx, country)
		curDestCountries = append(curDestCountries, curDestCountryObj)
	}
	curDestCountryValues := make([]attr.Value, len(curDestCountries))
	for i, v := range curDestCountries {
		curDestCountryValues[i] = v
	}
	curRuleDestObjAttrs["country"], diagstmp = types.ListValue(NameIDObjectType, curDestCountryValues)
	// if diagstmp.HasError() {
	// 	diagstmp = addDebugLineNumber(diagstmp, "rule.curDestCountries")
	// }
	diags = append(diags, diagstmp...)

	// rule.destination.domain[]
	curRuleDestObjAttrs["domain"], diagstmp = types.ListValueFrom(ctx, types.StringType, ruleResponse.Destination.Domain)
	// if diagstmp.HasError() {
	// 	diagstmp = addDebugLineNumber(diagstmp, "rule.curRuleDestObjAttrs.domain")
	// }
	diags = append(diags, diagstmp...)

	// rule.destination.fqdn[]
	curRuleDestObjAttrs["fqdn"], diagstmp = types.ListValueFrom(ctx, types.StringType, ruleResponse.Destination.Fqdn)
	// if diagstmp.HasError() {
	// 	diagstmp = addDebugLineNumber(diagstmp, "rule.curRuleDestObjAttrs.fqdn")
	// }
	diags = append(diags, diagstmp...)

	// rule.destination.ip[]
	curRuleDestObjAttrs["ip"], diagstmp = types.ListValueFrom(ctx, types.StringType, ruleResponse.Destination.IP)
	// if diagstmp.HasError() {
	// 	diagstmp = addDebugLineNumber(diagstmp, "rule.curRuleDestObjAttrs.ip")
	// }
	diags = append(diags, diagstmp...)

	// rule.destination.subnet[]
	curRuleDestObjAttrs["subnet"], diagstmp = types.ListValueFrom(ctx, types.StringType, ruleResponse.Destination.Subnet)
	// if diagstmp.HasError() {
	// 	diagstmp = addDebugLineNumber(diagstmp, "rule.curRuleDestObjAttrs.subnet")
	// }
	diags = append(diags, diagstmp...)

	// // rule.destination.ip_range[]
	// var curDestIPRanges []types.List
	// tflog.Warn(ctx, "ruleResponse.Destination.IPRange - "+fmt.Sprintf("%v", ruleResponse.Destination.IPRange))
	// for _, ipRange := range ruleResponse.Destination.IPRange {
	// 	curDestIPRangeObj := parseNameIDList(ctx, ipRange)
	// 	curDestIPRanges = append(curDestIPRanges, curDestIPRangeObj)
	// }
	// curDestIPRangeValues := make([]attr.Value, len(curDestIPRanges))
	// for i, v := range curDestIPRanges {
	// 	curDestIPRangeValues[i] = v
	// }
	// curRuleDestObjAttrs["ip_range"], diagstmp = types.ListValue(NameIDObjectType, curDestIPRangeValues)
	// // if diagstmp.HasError() {
	// // 	diagstmp = addDebugLineNumber(diagstmp, "rule.curDestIPRanges")
	// // }
	// diags = append(diags, diagstmp...)

	// // rule.destination.global_ip_range[]
	// var curDestGlobalIPRanges []types.List
	// tflog.Warn(ctx, "ruleResponse.Destination.GlobalIPRange - "+fmt.Sprintf("%v", ruleResponse.Destination.GlobalIPRange))
	// for _, globalIPRange := range ruleResponse.Destination.GlobalIPRange {
	// 	curDestGlobalIPRangeObj := parseNameIDList(ctx, globalIPRange)
	// 	curDestGlobalIPRanges = append(curDestGlobalIPRanges, curDestGlobalIPRangeObj)
	// }
	// curDestGlobalIPRangeValues := make([]attr.Value, len(curDestGlobalIPRanges))
	// for i, v := range curDestGlobalIPRanges {
	// 	curDestGlobalIPRangeValues[i] = v
	// }
	// curRuleDestObjAttrs["global_ip_range"], diagstmp = types.ListValue(NameIDObjectType, curDestGlobalIPRangeValues)
	// // if diagstmp.HasError() {
	// // 	diagstmp = addDebugLineNumber(diagstmp, "rule.curDestGlobalIPRanges")
	// // }
	// diags = append(diags, diagstmp...)

	// rule.destination.remote_asn[]
	remoteAsnValues := make([]attr.Value, len(ruleResponse.Destination.RemoteAsn))
	for i, asn := range ruleResponse.Destination.RemoteAsn {
		remoteAsnValues[i] = basetypes.NewStringValue(string(asn))
	}
	curRuleDestObjAttrs["remote_asn"], diagstmp = types.ListValue(types.StringType, remoteAsnValues)
	// if diagstmp.HasError() {
	// 	diagstmp = addDebugLineNumber(diagstmp, "rule.remoteAsnValues")
	// }
	diags = append(diags, diagstmp...)

	curRule["destination"] = curRuleDestObj
	////////////// end rule.destination ///////////////

	// ////////////// rule.containers ///////////////
	// // rule.containers.containers.fqdnContainer[]
	// var fqdnContainers []*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Containers_fqdnContainer
	// for _, fqdnContainer := range curRule.Containers.FqdnContainer {
	// 	curFqdnContainer := &Policy_Policy_InternetFirewall_Policy_Rules_Rule_Containers_fqdnContainer{}
	// 	state.Rule.As(ctx, &curFqdnContainer, basetypes.ObjectAsOptions{})
	// 	curFqdnContainer.Name = basetypes.NewStringValue(fqdnContainer.Name)
	// 	curFqdnContainer.ID = basetypes.NewStringValue(fqdnContainer.ID)
	// 	fqdnContainers = append(fqdnContainers, curFqdnContainer)
	// }
	// resp.State.SetAttribute(ctx, path.Root("rule").AtName("containers").AtName("fqdnContainer"), fqdnContainers)

	// // rule.containers.containers.ipAddressRangeContainer[]
	// var ipAddressContainers []*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Containers_ipAddressContainer
	// for _, ipAddressContainer := range curRule.Containers.IpAddressContainer {
	// 	curIpAddressContainer := &Policy_Policy_InternetFirewall_Policy_Rules_Rule_Containers_ipAddressContainer{}
	// 	state.Rule.As(ctx, &curIpAddressContainer, basetypes.ObjectAsOptions{})
	// 	curIpAddressContainer.Name = basetypes.NewStringValue(ipAddressContainer.Name)
	// 	curIpAddressContainer.ID = basetypes.NewStringValue(ipAddressContainer.ID)
	// 	ipAddressContainers = append(ipAddressContainers, curIpAddressContainer)
	// }
	// resp.State.SetAttribute(ctx, path.Root("rule").AtName("containers").AtName("ipAddressRangeContainer"), ipAddressContainers)
	// ////////////// end rule.containers ///////////////////

	// // //////////// end rule.service ///////////////
	// curRuleServiceObj, diagstmp := types.ObjectValue(
	// 	map[string]attr.Type{
	// 		"standard": types.ListType{ElemType: NameIDObjectType},
	// 		"custom":   types.ListType{ElemType: NameIDObjectType},
	// 	},
	// 	map[string]attr.Value{
	// 		"standard": types.ListNull(NameIDObjectType),
	// 		"custom":   types.ListNull(NameIDObjectType),
	// 	},
	// )
	// // if diagstmp.HasError() {
	// // 	diagstmp = addDebugLineNumber(diagstmp, "rule.curRuleServiceObj")
	// // }
	// diags = append(diags, diagstmp...)
	// curRuleServiceObjAttrs := curRuleServiceObj.Attributes()

	// var curRuleStandardServices []types.List
	// tflog.Warn(ctx, "ruleResponse.Service.Standard - "+fmt.Sprintf("%v", ruleResponse.Service.Standard))
	// for _, curRuleStandardService := range ruleResponse.Service.Standard {
	// 	curRuleStandardServiceObj := parseNameIDList(ctx, curRuleStandardService)
	// 	curRuleStandardServices = append(curRuleStandardServices, curRuleStandardServiceObj)
	// }
	// curRuleStandardServiceValues := make([]attr.Value, len(curRuleStandardServices))
	// for i, v := range curRuleStandardServices {
	// 	curRuleStandardServiceValues[i] = v
	// }
	// curRuleServiceObjAttrs["standard"], diagstmp = types.ListValue(NameIDObjectType, curRuleStandardServiceValues)
	// // if diagstmp.HasError() {
	// // 	diagstmp = addDebugLineNumber(diagstmp, "rule.curRuleStandardServices")
	// // }
	// diags = append(diags, diagstmp...)

	// // rule.service.custom[]
	// var curRuleCustomServices []attr.Value
	// for _, curRuleCustomService := range ruleResponse.Service.Custom {
	// 	curRuleCustomServiceObj, diagstmp := types.ObjectValue(
	// 		map[string]attr.Type{
	// 			"port": types.ListType{ElemType: basetypes.StringType{}},
	// 			"port_range": types.ObjectType{AttrTypes: map[string]attr.Type{
	// 				"from": basetypes.StringType{},
	// 				"to":   basetypes.StringType{},
	// 			}},
	// 			"protocol": basetypes.StringType{},
	// 		},
	// 		map[string]attr.Value{
	// 			"port": types.ListNull(basetypes.StringType{}),
	// 			"port_range": types.ObjectNull(map[string]attr.Type{
	// 				"from": basetypes.StringType{},
	// 				"to":   basetypes.StringType{},
	// 			}),
	// 			"protocol": types.StringNull(),
	// 		},
	// 	)
	// 	// if diagstmp.HasError() {
	// 	// 	diagstmp = addDebugLineNumber(diagstmp, "rule.curRuleCustomServices")
	// 	// }
	// 	diags = append(diags, diagstmp...)

	// 	curRuleSourceObjAttrs := curRuleCustomServiceObj.Attributes()
	// 	curRuleSourceObjAttrs["protocol"] = basetypes.NewStringValue(curRuleCustomService.Protocol.String())

	// 	// rule.service.custom.port[]
	// 	if curRuleCustomService.Port != nil {
	// 		customServicePorts := []string{}
	// 		for _, port := range curRuleCustomService.Port {
	// 			customServicePorts = append(customServicePorts, string(port))
	// 		}
	// 		curRuleSourceObjAttrs["port"], diagstmp = types.ListValueFrom(ctx, basetypes.StringType{}, customServicePorts)
	// 		// if diagstmp.HasError() {
	// 		// 	diagstmp = addDebugLineNumber(diagstmp, "rule.curRuleCustomService.port")
	// 		// }
	// 		diags = append(diags, diagstmp...)
	// 	}

	// 	if curRuleCustomService.PortRange != nil {
	// 		curRuleSourceObjAttrs["port_range"], diagstmp = types.ObjectValue(
	// 			map[string]attr.Type{
	// 				"from": basetypes.StringType{},
	// 				"to":   basetypes.StringType{},
	// 			},
	// 			map[string]attr.Value{
	// 				"from": basetypes.NewStringValue(string(curRuleCustomService.PortRange.From)),
	// 				"to":   basetypes.NewStringValue(string(curRuleCustomService.PortRange.To)),
	// 			},
	// 		)
	// 		// if diagstmp.HasError() {
	// 		// 	diagstmp = addDebugLineNumber(diagstmp, "rule.curRuleCustomService.portRange")
	// 		// }
	// 		diags = append(diags, diagstmp...)
	// 	}

	// 	curRuleCustomServices = append(curRuleCustomServices, curRuleCustomServiceObj)
	// }
	// curRuleServiceObjAttrs["custom"], diagstmp = types.ListValue(NameIDObjectType, curRuleCustomServices)
	// // if diagstmp.HasError() {
	// // 	diagstmp = addDebugLineNumber(diagstmp, "rule.custom_service")
	// // }
	// diags = append(diags, diagstmp...)
	// ////////////// end rule.service ///////////////

	// rule.action
	curRule["action"] = basetypes.NewStringValue(ruleResponse.Action.String())

	// rule.tracking
	// curRuleTrackingObj, destDiags := types.ObjectValue(
	// 	map[string]attr.Type{
	// 		"event": types.ListType{ElemType: NameIDObjectType},
	// 		"alert": types.ListType{ElemType: NameIDObjectType},
	// 	},
	// 	map[string]attr.Value{
	// 		"event": types.ListNull(NameIDObjectType),
	// 		"alert": types.ListNull(NameIDObjectType),
	// 	},
	// )
	// resp.Diagnostics.Append(destDiags...)
	// if resp.Diagnostics.HasError() {
	// 	return curRule
	// }
	// curRuleTrackingObjAttrs := curRuleTrackingObj.Attributes()

	// curRuleTrackingObjAttrs["event"], _ = types.ListValueFrom(ctx, NameIDObjectType, ruleResponse.Tracking.Event.Enabled)
	// curRuleAlertObj, destDiags := types.ObjectValue(
	// 	AlertAttrTypes,
	// 	map[string]attr.Value{
	// 		"event": types.ObjectNull(EnabledAttrTypes),
	// 		"alert": types.ListNull(AlertObjectType),
	// 	},
	// )
	// resp.Diagnostics.Append(destDiags...)
	// if resp.Diagnostics.HasError() {
	// 	return curRule
	// }
	// curRuleTrackingObjAttrs := curRuleTrackingObj.Attributes()

	// // curRuleTrackingObjAttrs["alert"], _ = types.ListValueFrom(ctx, NameIDObjectType, ruleResponse.Tracking.Alert)

	// // resp.State.SetAttribute(ctx, path.Root("rule").AtName("action"), curRule.Action)
	// // resp.State.SetAttribute(ctx, path.Root("rule").AtName("tracking").AtName("event").AtName("enabled"), curRule.Tracking.Event.Enabled)
	// // resp.State.SetAttribute(ctx, path.Root("rule").AtName("tracking").AtName("alert").AtName("enabled"), curRule.Tracking.Alert.Enabled)
	// // resp.State.SetAttribute(ctx, path.Root("rule").AtName("tracking").AtName("alert").AtName("frequency"), curRule.Tracking.Alert.Frequency)

	// // rule.tracking.alert.subscription_group{}
	// var alertSubscriptionGroups []*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Alert_SubscriptionGroup
	// for _, alertSubscriptionGroup := range curRule.Tracking.Alert.SubscriptionGroup {
	// 	curAlertSubscriptionGroup := &Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Alert_SubscriptionGroup{}
	// 	state.Rule.As(ctx, &curAlertSubscriptionGroup, basetypes.ObjectAsOptions{})
	// 	curAlertSubscriptionGroup.Name = basetypes.NewStringValue(alertSubscriptionGroup.Name)
	// 	curAlertSubscriptionGroup.ID = basetypes.NewStringValue(alertSubscriptionGroup.ID)
	// 	alertSubscriptionGroups = append(alertSubscriptionGroups, curAlertSubscriptionGroup)
	// }
	// resp.State.SetAttribute(ctx, path.Root("rule").AtName("tracking").AtName("alert").AtName("subscription_group"), alertSubscriptionGroups)

	// // rule.tracking.alert.webhook{}
	// var alertWebHooks []*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Alert_Webhook
	// for _, alertWebHook := range curRule.Tracking.Alert.Webhook {
	// 	curAlertWebHook := &Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Alert_Webhook{}
	// 	state.Rule.As(ctx, &curAlertWebHook, basetypes.ObjectAsOptions{})
	// 	curAlertWebHook.Name = basetypes.NewStringValue(alertWebHook.Name)
	// 	curAlertWebHook.ID = basetypes.NewStringValue(alertWebHook.ID)
	// 	alertWebHooks = append(alertWebHooks, curAlertWebHook)
	// }
	// resp.State.SetAttribute(ctx, path.Root("rule").AtName("tracking").AtName("alert").AtName("webhooks"), alertWebHooks)

	// // rule.tracking.alert.mailing_list{}
	// var alertMailingLists []*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Alert_MailingList
	// for _, alertMailingList := range curRule.Tracking.Alert.MailingList {
	// 	curAlertMailingList := &Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Alert_MailingList{}
	// 	state.Rule.As(ctx, &curAlertMailingList, basetypes.ObjectAsOptions{})
	// 	curAlertMailingList.Name = basetypes.NewStringValue(alertMailingList.Name)
	// 	curAlertMailingList.ID = basetypes.NewStringValue(alertMailingList.ID)
	// 	alertMailingLists = append(alertMailingLists, curAlertMailingList)
	// }
	// resp.State.SetAttribute(ctx, path.Root("rule").AtName("tracking").AtName("alert").AtName("mailing_list"), alertMailingLists)

	// // rule.schedule.active_on{}
	// resp.State.SetAttribute(ctx, path.Root("rule").AtName("schedule").AtName("active_on"), curRule.Schedule.ActiveOn)
	// // rule.schedule.custom_timeframe{}
	// if curRule.Schedule.CustomTimeframePolicySchedule != nil {
	// 	resp.State.SetAttribute(ctx, path.Root("rule").AtName("schedule").AtName("custom_timeframe").AtName("from"), curRule.Schedule.CustomTimeframePolicySchedule.From)
	// 	resp.State.SetAttribute(ctx, path.Root("rule").AtName("schedule").AtName("custom_timeframe").AtName("to"), curRule.Schedule.CustomTimeframePolicySchedule.To)
	// }
	// // rule.schedule.custom_recurring{}
	// if curRule.Schedule.CustomRecurringPolicySchedule != nil {
	// 	resp.State.SetAttribute(ctx, path.Root("rule").AtName("schedule").AtName("custom_recurring").AtName("from"), curRule.Schedule.CustomRecurringPolicySchedule.From)
	// 	resp.State.SetAttribute(ctx, path.Root("rule").AtName("schedule").AtName("custom_recurring").AtName("to"), curRule.Schedule.CustomRecurringPolicySchedule.To)
	// 	resp.State.SetAttribute(ctx, path.Root("rule").AtName("schedule").AtName("custom_recurring").AtName("days"), curRule.Schedule.CustomRecurringPolicySchedule.Days)
	// }

	// // // rule.exceptions[]
	// var ruleExceptions []*Policy_Policy_InternetFirewall_Policy_Rules_Rule_Exceptions
	// for _, ruleException := range curRule.Exceptions {
	// 	curException := &Policy_Policy_InternetFirewall_Policy_Rules_Rule_Exceptions{}
	// 	state.Rule.As(ctx, &curException, basetypes.ObjectAsOptions{})

	// 	curException.Name = basetypes.NewStringValue(string(ruleException.Name))
	// 	curException.ConnectionOrigin = basetypes.NewStringValue(ruleException.ConnectionOrigin.String())

	// 	// rule.exceptions.source{}
	// 	curSource := &Policy_Policy_InternetFirewall_Policy_Rules_Rule_Source{}
	// 	state.Rule.As(ctx, &curSource, basetypes.ObjectAsOptions{})
	// 	resp.Diagnostics.Append(diags...)
	// 	if resp.Diagnostics.HasError() {
	// 		return
	// 	}
	// 	curSource.IP = parseStringList(ctx, ruleException.Source.IP)
	// 	curSource.Host = parseNameIDList(ctx, ruleException.Source.Host)
	// 	curSource.Site = parseNameIDList(ctx, ruleException.Source.Site)
	// 	curSource.Site = parseNameIDList(ctx, ruleException.Source.Site)
	// 	curSource.Subnet = parseStringList(ctx, ruleException.Source.Subnet)
	// 	curSource.IPRange = parseFromToList(ruleException.Source.IPRange)
	// 	curSource.GlobalIPRange = parseNameIDList(ctx, ruleException.Source.GlobalIPRange)
	// 	curSource.NetworkInterface = parseNameIDList(ctx, ruleException.Source.NetworkInterface)
	// 	curSource.SiteNetworkSubnet = parseNameIDList(ctx, ruleException.Source.SiteNetworkSubnet)
	// 	curSource.FloatingSubnet = parseNameIDList(ctx, ruleException.Source.FloatingSubnet)
	// 	curSource.User = parseNameIDList(ctx, ruleException.Source.User)
	// 	curSource.UsersGroup = parseNameIDList(ctx, ruleException.Source.UsersGroup)
	// 	curSource.Group = parseNameIDList(ctx, ruleException.Source.Group)
	// 	curSource.SystemGroup = parseNameIDList(ctx, ruleException.Source.SystemGroup)

	// 	sourceObj, sourceDiags := types.ObjectValue(
	// 		map[string]attr.Type{
	// 			"ip":                  types.ListType{ElemType: types.StringType},
	// 			"host":                types.ListType{ElemType: NameIDObjectType},
	// 			"site":                types.ListType{ElemType: NameIDObjectType},
	// 			"subnet":              types.ListType{ElemType: types.StringType},
	// 			"ip_range":            types.ListType{ElemType: FromToObjectType},
	// 			"global_ip_range":     types.ListType{ElemType: NameIDObjectType},
	// 			"network_interface":   types.ListType{ElemType: NameIDObjectType},
	// 			"site_network_subnet": types.ListType{ElemType: NameIDObjectType},
	// 			"floating_subnet":     types.ListType{ElemType: NameIDObjectType},
	// 			"user":                types.ListType{ElemType: NameIDObjectType},
	// 			"users_group":         types.ListType{ElemType: NameIDObjectType},
	// 			"group":               types.ListType{ElemType: NameIDObjectType},
	// 			"system_group":        types.ListType{ElemType: NameIDObjectType},
	// 		},
	// 		map[string]attr.Value{
	// 			"ip":                  curSource.IP,
	// 			"host":                curSource.Host,
	// 			"site":                curSource.Site,
	// 			"subnet":              curSource.Subnet,
	// 			"ip_range":            curSource.IPRange,
	// 			"global_ip_range":     curSource.GlobalIPRange,
	// 			"network_interface":   curSource.NetworkInterface,
	// 			"site_network_subnet": curSource.SiteNetworkSubnet,
	// 			"floating_subnet":     curSource.FloatingSubnet,
	// 			"user":                curSource.User,
	// 			"users_group":         curSource.UsersGroup,
	// 			"group":               curSource.Group,
	// 			"system_group":        curSource.SystemGroup,
	// 		},
	// 	)

	// 	resp.Diagnostics.Append(sourceDiags...)
	// 	if resp.Diagnostics.HasError() {
	// 		return
	// 	}
	// 	curException.Source = sourceObj

	// 	// rule.exceptions.country{}
	// 	curException.Country = parseNameIDList(ctx, ruleException.Country)

	// 	// rule.exceptions.Device{}
	// 	curException.Device = parseStringList(ctx, ruleException.Device)

	// 	// rule.exceptions.DeviceOs{}
	// 	curException.DeviceOs = parseStringList(ctx, ruleException.DeviceOs)

	// 	// rule.exceptions.destination{}
	// 	curExceptionDestination := &Policy_Policy_InternetFirewall_Policy_Rules_Rule_Destination{}
	// 	state.Rule.As(ctx, &curExceptionDestination, basetypes.ObjectAsOptions{})
	// 	resp.Diagnostics.Append(diags...)
	// 	if resp.Diagnostics.HasError() {
	// 		return
	// 	}
	// 	curExceptionDestination.Application = parseNameIDList(ctx, ruleException.Destination.Application)
	// 	curExceptionDestination.CustomApp = parseNameIDList(ctx, ruleException.Destination.CustomApp)
	// 	curExceptionDestination.AppCategory = parseNameIDList(ctx, ruleException.Destination.AppCategory)
	// 	curExceptionDestination.CustomCategory = parseNameIDList(ctx, ruleException.Destination.CustomCategory)
	// 	curExceptionDestination.SanctionedAppsCategory = parseNameIDList(ctx, ruleException.Destination.SanctionedAppsCategory)
	// 	curExceptionDestination.Country = parseNameIDList(ctx, ruleException.Destination.Country)
	// 	curExceptionDestination.Domain = parseStringList(ctx, ruleException.Destination.Domain)
	// 	curExceptionDestination.Fqdn = parseStringList(ctx, ruleException.Destination.Fqdn)
	// 	curExceptionDestination.IP = parseStringList(ctx, ruleException.Destination.IP)
	// 	curExceptionDestination.Subnet = parseStringList(ctx, ruleException.Destination.Subnet)
	// 	curExceptionDestination.IPRange = parseFromToList(ruleException.Destination.IPRange)
	// 	curExceptionDestination.GlobalIPRange = parseNameIDList(ctx, ruleException.Destination.GlobalIPRange)
	// 	curExceptionDestination.RemoteAsn = parseStringList(ctx, ruleException.Destination.RemoteAsn)

	// 	destObj, destDiags := types.ObjectValue(
	// 		map[string]attr.Type{
	// 			"application":              types.ListType{ElemType: NameIDObjectType},
	// 			"custom_app":               types.ListType{ElemType: NameIDObjectType},
	// 			"app_category":             types.ListType{ElemType: NameIDObjectType},
	// 			"custom_category":          types.ListType{ElemType: NameIDObjectType},
	// 			"sanctioned_apps_category": types.ListType{ElemType: NameIDObjectType},
	// 			"country":                  types.ListType{ElemType: NameIDObjectType},
	// 			"domain":                   types.ListType{ElemType: types.StringType},
	// 			"fqdn":                     types.ListType{ElemType: types.StringType},
	// 			"ip":                       types.ListType{ElemType: types.StringType},
	// 			"subnet":                   types.ListType{ElemType: types.StringType},
	// 			"ip_range":                 types.ListType{ElemType: FromToObjectType},
	// 			"global_ip_range":          types.ListType{ElemType: NameIDObjectType},
	// 			"remote_asn":               types.ListType{ElemType: types.StringType},
	// 		},
	// 		map[string]attr.Value{
	// 			"application":              curExceptionDestination.Application,
	// 			"custom_app":               curExceptionDestination.CustomApp,
	// 			"app_category":             curExceptionDestination.AppCategory,
	// 			"custom_category":          curExceptionDestination.CustomCategory,
	// 			"sanctioned_apps_category": curExceptionDestination.SanctionedAppsCategory,
	// 			"country":                  curExceptionDestination.Country,
	// 			"domain":                   curExceptionDestination.Domain,
	// 			"fqdn":                     curExceptionDestination.Fqdn,
	// 			"ip":                       curExceptionDestination.IP,
	// 			"subnet":                   curExceptionDestination.Subnet,
	// 			"ip_range":                 curExceptionDestination.IPRange,
	// 			"global_ip_range":          curExceptionDestination.GlobalIPRange,
	// 			"remote_asn":               curExceptionDestination.RemoteAsn,
	// 		},
	// 	)
	// 	resp.Diagnostics.Append(destDiags...)
	// 	if resp.Diagnostics.HasError() {
	// 		return
	// 	}
	// 	curException.Destination = destObj

	// 	// rule.exceptions.service{}
	// 	var CustomServiceAttrTypes = map[string]attr.Type{
	// 		"port":       types.ListType{ElemType: types.StringType},
	// 		"port_range": types.ObjectType{AttrTypes: FromToAttrTypes},
	// 		"protocol":   types.StringType,
	// 	}
	// 	var customServices []attr.Value
	// 	if len(ruleException.Service.Custom) > 0 {
	// 		customServices = make([]attr.Value, len(ruleException.Service.Custom))
	// 		for i, ruleExceptionCustomService := range ruleException.Service.Custom {
	// 			ports := parseStringList(ctx, ruleExceptionCustomService.Port)
	// 			var portRange types.Object
	// 			if ruleExceptionCustomService.PortRangeCustomService != nil {
	// 				portRangeObj, portRangeDiags := types.ObjectValue(
	// 					FromToAttrTypes,
	// 					map[string]attr.Value{
	// 						"from": basetypes.NewStringValue(string(ruleExceptionCustomService.PortRangeCustomService.From)),
	// 						"to":   basetypes.NewStringValue(string(ruleExceptionCustomService.PortRangeCustomService.To)),
	// 					},
	// 				)
	// 				resp.Diagnostics.Append(portRangeDiags...)
	// 				if resp.Diagnostics.HasError() {
	// 					return
	// 				}
	// 				portRange = portRangeObj
	// 			} else {
	// 				portRange = types.ObjectNull(FromToAttrTypes)
	// 			}

	// 			// Create custom service object
	// 			customServiceObj, customDiags := types.ObjectValue(
	// 				CustomServiceAttrTypes,
	// 				map[string]attr.Value{
	// 					"port":       ports,
	// 					"port_range": portRange,
	// 					"protocol":   basetypes.NewStringValue(ruleExceptionCustomService.Protocol.String()),
	// 				},
	// 			)
	// 			resp.Diagnostics.Append(customDiags...)
	// 			if resp.Diagnostics.HasError() {
	// 				return
	// 			}
	// 			customServices[i] = customServiceObj
	// 		}
	// 	}

	// 	// Create custom services list
	// 	var CustomServiceObjectType = types.ObjectType{AttrTypes: CustomServiceAttrTypes}
	// 	customList, customListDiags := types.ListValue(CustomServiceObjectType, customServices)
	// 	resp.Diagnostics.Append(customListDiags...)
	// 	if resp.Diagnostics.HasError() {
	// 		return
	// 	}

	// 	standardList := parseNameIDList(ctx, ruleException.Service.Standard)
	// 	// Create service object
	// 	serviceObj, serviceDiags := types.ObjectValue(
	// 		map[string]attr.Type{
	// 			"standard": types.ListType{ElemType: NameIDObjectType},
	// 			"custom":   types.ListType{ElemType: CustomServiceObjectType},
	// 		},
	// 		map[string]attr.Value{
	// 			"standard": standardList,
	// 			"custom":   customList,
	// 		},
	// 	)
	// 	resp.Diagnostics.Append(serviceDiags...)
	// 	if resp.Diagnostics.HasError() {
	// 		return
	// 	}
	// 	curException.Service = serviceObj

	// 	ruleExceptions = append(ruleExceptions, curException)
	// }
	// curRuleObject, _ := types.ObjectValue(CurRuleType.AttrTypes, curRule)

	diagstmpstate := state.Rule.As(ctx, curRule, basetypes.ObjectAsOptions{})
	diags = append(diags, diagstmpstate...)
	// state.Rule = curRuleObj
	return state, diags
}

func convertStringSlice(input []string) []string {
	var result []string
	for _, v := range input {
		result = append(result, v)
	}
	return result
}

// Define a reusable object types

// Define a reusable type map for name/id pairs
var EnabledAttrTypes = map[string]attr.Type{
	"enabled": types.BoolType,
}
var EnabledObjectType = types.ObjectType{AttrTypes: NameIDAttrTypes}

var NameIDAttrTypes = map[string]attr.Type{
	"name": types.StringType,
	"id":   types.StringType,
}

// ObjectType wrapper for ListValue
var NameIDObjectType = types.ObjectType{AttrTypes: NameIDAttrTypes}

// Define a reusable type map for name/id pairs
var AlertAttrTypes = map[string]attr.Type{
	"enabled":           types.BoolType,
	"frequency":         types.StringType,
	"subscriptionGroup": types.ListType{ElemType: NameIDObjectType},
	"webhook":           types.ListType{ElemType: NameIDObjectType},
	"mailingList":       types.ListType{ElemType: NameIDObjectType},
}

// ObjectType wrapper for ListValue
var AlertObjectType = types.ObjectType{AttrTypes: AlertAttrTypes}

func parseNameIDList(ctx context.Context, items interface{}) types.List {
	tflog.Warn(ctx, "parseNameIDList() - "+fmt.Sprintf("%v", items))
	diags := make(diag.Diagnostics, 0)
	// Get the reflect.Value of the input
	itemsValue := reflect.ValueOf(items)

	// Handle nil or empty input
	rt := reflect.TypeOf(items)
	if items == nil || (rt.Kind() != reflect.Array && rt.Kind() != reflect.Slice) {
		return types.ListNull(NameIDObjectType)
	} else {
		if itemsValue.Len() == 0 {
			return types.ListNull(NameIDObjectType)
		}
	}

	values := make([]attr.Value, itemsValue.Len())
	for i := range itemsValue.Len() {
		item := itemsValue.Index(i)

		// Handle pointer elements
		if item.Kind() == reflect.Ptr {
			item = item.Elem()
		}

		// Get Name and ID fields
		nameField := item.FieldByName("Name")
		idField := item.FieldByName("ID")

		if !nameField.IsValid() || !idField.IsValid() {
			return types.ListNull(NameIDObjectType)
		}

		// Create object value
		obj, diagstmp := types.ObjectValue(
			NameIDAttrTypes,
			map[string]attr.Value{
				"name": basetypes.NewStringValue(nameField.String()),
				"id":   basetypes.NewStringValue(idField.String()),
			},
		)
		if diagstmp.HasError() {
			diagstmp = addDebugLineNumber(diagstmp, "parseNameIDListLoop")
		}
		diags = append(diags, diagstmp...)
		values[i] = obj
	}

	// Convert to List
	list, diagstmp := types.ListValue(NameIDObjectType, values)
	if diagstmp.HasError() {
		diagstmp = addDebugLineNumber(diagstmp, "parseNameIDList")
	}
	diags = append(diags, diagstmp...)

	return list
}

var FromToAttrTypes = map[string]attr.Type{
	"from": types.StringType,
	"to":   types.StringType,
}

var FromToObjectType = types.ObjectType{AttrTypes: FromToAttrTypes}

var GenericObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{},
}

// CurRuleType defines the type structure for "rule"
var CurRuleType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"id":               types.StringType,
		"name":             types.StringType,
		"description":      types.StringType,
		"index":            types.Int64Type,
		"section":          GenericObjectType,
		"enabled":          types.BoolType,
		"source":           GenericObjectType,
		"connectionOrigin": types.StringType,
		"country":          types.ListType{ElemType: NameIDObjectType},
		"device":           types.ListType{ElemType: NameIDObjectType},
		"deviceOS":         types.ListType{ElemType: types.StringType},
		"deviceAttributes": GenericObjectType,
		"destination":      GenericObjectType,
		"service":          GenericObjectType,
		"action":           types.StringType,
		"tracking":         GenericObjectType,
		"schedule":         GenericObjectType,
		"exceptions":       types.ListType{ElemType: GenericObjectType},
	},
}

// Function to create a null curRule object
func createNullCurRule(ctx context.Context) types.Object {
	diags := make(diag.Diagnostics, 0)
	nullValues := map[string]attr.Value{
		"id":               types.StringNull(),
		"name":             types.StringNull(),
		"description":      types.StringNull(),
		"index":            types.Int64Null(),
		"section":          types.ObjectNull(GenericObjectType.AttrTypes),
		"enabled":          types.BoolNull(),
		"source":           types.ObjectNull(GenericObjectType.AttrTypes),
		"connectionOrigin": types.StringNull(),
		"country":          types.ListNull(NameIDObjectType),
		"device":           types.ListNull(NameIDObjectType),
		"deviceOS":         types.ListNull(types.StringType),
		"deviceAttributes": types.ObjectNull(GenericObjectType.AttrTypes),
		"destination":      types.ObjectNull(GenericObjectType.AttrTypes),
		"service":          types.ObjectNull(GenericObjectType.AttrTypes),
		"action":           types.StringNull(),
		"tracking":         types.ObjectNull(GenericObjectType.AttrTypes),
		"schedule":         types.ObjectNull(GenericObjectType.AttrTypes),
		"exceptions":       types.ListNull(GenericObjectType),
	}
	curRule, diagstmp := types.ObjectValue(CurRuleType.AttrTypes, nullValues)
	if diagstmp.HasError() {
		diagstmp = addDebugLineNumber(diagstmp, "createNullCurRule")
	}
	diags = append(diags, diagstmp...)
	return curRule
}

func parseFromToList(items interface{}) types.List {
	diags := make(diag.Diagnostics, 0)
	// Get the reflect.Value of the input
	itemsValue := reflect.ValueOf(items)

	// Handle nil or empty input
	rt := reflect.TypeOf(items)
	if items == nil || (rt.Kind() != reflect.Array && rt.Kind() != reflect.Slice) {
		return types.ListNull(NameIDObjectType)
	} else {
		if itemsValue.Len() == 0 {
			return types.ListNull(NameIDObjectType)
		}
	}

	values := make([]attr.Value, itemsValue.Len())
	for i := range itemsValue.Len() {
		item := itemsValue.Index(i)

		// Handle pointer elements
		if item.Kind() == reflect.Ptr {
			item = item.Elem()
		}

		// Get From and To fields
		fromField := item.FieldByName("From")
		toField := item.FieldByName("To")

		if !fromField.IsValid() || !toField.IsValid() {
			return types.ListNull(FromToObjectType)
		}

		// Create object value
		obj, _ := types.ObjectValue(
			FromToAttrTypes,
			map[string]attr.Value{
				"from": basetypes.NewStringValue(fromField.String()),
				"to":   basetypes.NewStringValue(toField.String()),
			},
		)
		values[i] = obj
	}

	// Convert to List
	list, diagstmp := types.ListValue(FromToObjectType, values)
	if diagstmp.HasError() {
		diagstmp = addDebugLineNumber(diagstmp, "parseFromToList")
	}
	diags = append(diags, diagstmp...)
	return list
}

func parseStringList(ctx context.Context, input interface{}) types.List {
	diags := make(diag.Diagnostics, 0)

	// Get the reflect.Value of the input
	inputValue := reflect.ValueOf(input)

	// Handle nil or empty input
	if input == nil || inputValue.Len() == 0 {
		return types.ListNull(types.StringType)
	}

	// Ensure it's a slice or array
	rt := reflect.TypeOf(inputValue)
	if rt.Kind() != reflect.Array && rt.Kind() != reflect.Slice {
		return types.ListNull(types.StringType)
	}

	// Convert to []string
	stringSlice := make([]string, inputValue.Len())
	for i := range inputValue.Len() {
		item := inputValue.Index(i)

		// Handle interface{} or string values
		switch v := item.Interface().(type) {
		case string:
			stringSlice[i] = v
		case *string:
			if v != nil {
				stringSlice[i] = *v
			} else {
				stringSlice[i] = "" // or handle nil differently if needed
			}
		default:
			return types.ListNull(types.StringType)
		}
	}

	// Convert []string to types.List
	list, diagstmp := types.ListValueFrom(ctx, types.StringType, stringSlice)
	if diagstmp.HasError() {
		diagstmp = addDebugLineNumber(diagstmp, "parseStringList")
	}
	diags = append(diags, diagstmp...)
	return list
}

func addDebugLineNumber(diags diag.Diagnostics, message string) diag.Diagnostics {
	_, file, line, _ := runtime.Caller(1)
	return append(diags, diag.NewErrorDiagnostic(
		message,
		fmt.Sprintf("Occurred in %s:%d", file, line),
	))
}
