package provider

import (
	"context"
	"fmt"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/catonetworks/cato-go-sdk/scalars"
	cato_scalars "github.com/catonetworks/cato-go-sdk/scalars"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
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

type internetFwRuleResource struct {
	client *catoClientData
}

func NewInternetFwRuleResource() resource.Resource {
	return &internetFwRuleResource{}
}

func (r *internetFwRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_if_rule"
}

func (r *internetFwRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_if_rule` resource contains the configuration parameters necessary to add rule to the Internet Firewall. (check https://support.catonetworks.com/hc/en-us/articles/4413273486865-What-is-the-Cato-Internet-Firewall for more details). Documentation for the underlying API used in this resource can be found at [mutation.policy.internetFirewall.addRule()](https://api.catonetworks.com/documentation/#mutation-policy.internetFirewall.addRule).",
		Attributes: map[string]schema.Attribute{
			"at": schema.SingleNestedAttribute{
				Description: "Position of the rule in the policy (https://api.catonetworks.com/documentation/#definition-PolicyRulePositionInput)",
				Required:    true,
				Optional:    false,
				Attributes: map[string]schema.Attribute{
					"position": schema.StringAttribute{
						Description: "Position relative to a policy, a section or another rule (https://api.catonetworks.com/documentation/#definition-PolicyRulePositionEnum)",
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
			"rule": schema.SingleNestedAttribute{
				Description: "Parameters for the rule you are adding (https://api.catonetworks.com/documentation/#definition-InternetFirewallAddRuleDataInput)",
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
						Description: "Source traffic matching criteria. Logical ‘OR’ is applied within the criteria set. Logical ‘AND’ is applied between criteria sets. (https://api.catonetworks.com/documentation/#definition-InternetFirewallSourceInput)",
						Required:    true,
						Optional:    false,
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
											Description: "",
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
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.String{
							stringvalidator.OneOf("ANY", "REMOTE", "SITE"),
						},
						Computed: true,
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
						Description: "Source device Operating System traffic matching criteria. Logical ‘OR’ is applied within the criteria set. Logical ‘AND’ is applied between criteria sets.(https://api.catonetworks.com/documentation/#definition-OperatingSystem)",
						Optional:    true,
						Required:    false,
					},
					"destination": schema.SingleNestedAttribute{
						Description: "Destination traffic matching criteria. Logical ‘OR’ is applied within the criteria set. Logical ‘AND’ is applied between criteria sets. (https://api.catonetworks.com/documentation/#definition-InternetFirewallDestinationInput)",
						Optional:    false,
						Required:    true,
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
											Description: "",
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
								Description: "Globally defined IP range, IP and subnet objects.",
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
							"remote_asn": schema.ListAttribute{
								ElementType: types.StringType,
								Description: "Remote Autonomous System Number (ASN)",
								Required:    false,
								Optional:    true,
							},
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
					"action": schema.StringAttribute{
						Description: "The action applied by the Internet Firewall if the rule is matched (https://api.catonetworks.com/documentation/#definition-InternetFirewallActionEnum)",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("ALLOW", "BLOCK", "PROMPT", "RBI"),
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
										Description: "",
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
										Description: "",
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
									"subscription_group": schema.ListNestedAttribute{
										Description: "Returns data for the Subscription Group that receives the alert",
										Required:    false,
										Optional:    true,
										PlanModifiers: []planmodifier.List{
											listplanmodifier.UseStateForUnknown(), // Avoid drift
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
													// Computed: true,
												},
												"id": schema.StringAttribute{
													Description: "",
													Required:    false,
													Optional:    true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.UseStateForUnknown(), // Avoid drift
													},
													// Computed: true,
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
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.UseStateForUnknown(), // Avoid drift
													},
													// Computed: true,
												},
												"id": schema.StringAttribute{
													Description: "",
													Required:    false,
													Optional:    true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.UseStateForUnknown(), // Avoid drift
													},
													// Computed: true,
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
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.UseStateForUnknown(), // Avoid drift
													},
													// Computed: true,
												},
												"id": schema.StringAttribute{
													Description: "",
													Required:    false,
													Optional:    true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.UseStateForUnknown(), // Avoid drift
													},
													// Computed: true,
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
										Description: "",
										Required:    false,
										Optional:    true,
									},
									"to": schema.StringAttribute{
										Description: "",
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
										Description: "",
										Required:    false,
										Optional:    true,
									},
									"to": schema.StringAttribute{
										Description: "",
										Required:    false,
										Optional:    true,
									},
									"days": schema.ListAttribute{
										ElementType: types.StringType,
										Description: "(https://api.catonetworks.com/documentation/#definition-DayOfWeek)",
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
								"device": schema.ListNestedAttribute{
									Description: "Source Device Profile traffic matching criteria. Logical 'OR' is applied within the criteria set. Logical 'AND' is applied between criteria sets.",
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
														Description: "",
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
											PlanModifiers: []planmodifier.List{
												listplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Computed: true,
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
										"remote_asn": schema.ListAttribute{
											ElementType: types.StringType,
											Description: "",
											Required:    false,
											Optional:    true,
											PlanModifiers: []planmodifier.List{
												listplanmodifier.UseStateForUnknown(), // Avoid drift
											},
											Computed: true,
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
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"port": schema.ListAttribute{
														ElementType: types.StringType,
														Description: "",
														Optional:    true,
														Required:    false,
														PlanModifiers: []planmodifier.List{
															listplanmodifier.UseStateForUnknown(), // avoids plan drift
														},
														Computed: true,
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
									Description: "Connection origin matching criteria for the exception. (https://api.catonetworks.com/documentation/#definition-ConnectionOriginEnum)",
									Optional:    true,
									Required:    false,
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
			if !sourceInput.IP.IsUnknown() && !sourceInput.IP.IsNull() {
				diags = sourceInput.IP.ElementsAs(ctx, &input.Rule.Source.IP, false)
				resp.Diagnostics.Append(diags...)
			}

			// setting source subnet
			if !sourceInput.Subnet.IsUnknown() && !sourceInput.Subnet.IsNull() {
				diags = sourceInput.Subnet.ElementsAs(ctx, &input.Rule.Source.Subnet, false)
				resp.Diagnostics.Append(diags...)
			}

			// setting source host
			if !sourceInput.Host.IsUnknown() && !sourceInput.Host.IsNull() {
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
			if !sourceInput.Site.IsUnknown() && !sourceInput.Site.IsNull() {
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
			if !sourceInput.IPRange.IsUnknown() && !sourceInput.IPRange.IsNull() {
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
		if !ruleInput.DeviceOs.IsUnknown() && !ruleInput.DeviceOs.IsNull() {
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
			if !destinationInput.IP.IsUnknown() && !destinationInput.IP.IsNull() {
				diags = destinationInput.IP.ElementsAs(ctx, &input.Rule.Destination.IP, false)
				resp.Diagnostics.Append(diags...)
			}

			// setting destination subnet
			if !destinationInput.Subnet.IsUnknown() && !destinationInput.Subnet.IsNull() {
				diags = destinationInput.Subnet.ElementsAs(ctx, &input.Rule.Destination.Subnet, false)
				resp.Diagnostics.Append(diags...)
			}

			// setting destination domain
			if !destinationInput.Domain.IsUnknown() && !destinationInput.Domain.IsNull() {
				diags = destinationInput.Domain.ElementsAs(ctx, &input.Rule.Destination.Domain, false)
				resp.Diagnostics.Append(diags...)
			}

			// setting destination fqdn
			if !destinationInput.Fqdn.IsUnknown() && !destinationInput.Fqdn.IsNull() {
				diags = destinationInput.Fqdn.ElementsAs(ctx, &input.Rule.Destination.Fqdn, false)
				resp.Diagnostics.Append(diags...)
			}

			// setting destination remote asn
			if !destinationInput.RemoteAsn.IsUnknown() && !destinationInput.RemoteAsn.IsNull() {
				diags = destinationInput.RemoteAsn.ElementsAs(ctx, &input.Rule.Destination.RemoteAsn, false)
				resp.Diagnostics.Append(diags...)
			}

			// setting destination application
			if !destinationInput.Application.IsUnknown() && !destinationInput.Application.IsNull() {
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
					resp.Diagnostics.Append(diags...)

					customInput := &cato_models.CustomServiceInput{
						Protocol: cato_models.IPProtocol(itemServiceCustomInput.Protocol.ValueString()),
					}

					// setting service custom port
					tflog.Info(ctx, "itemServiceCustomInput.Port - "+fmt.Sprintf("%v", itemServiceCustomInput.Port))
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
						resp.Diagnostics.Append(diags...)

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
		if !ruleInput.Tracking.IsUnknown() && !ruleInput.Tracking.IsNull() {

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

			if !trackingInput.Event.IsUnknown() && !trackingInput.Event.IsNull() {
				// setting tracking event
				trackingEventInput := Policy_Policy_InternetFirewall_Policy_Rules_Rule_Tracking_Event{}
				diags = trackingInput.Event.As(ctx, &trackingEventInput, basetypes.ObjectAsOptions{})
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
				input.Rule.Tracking.Event.Enabled = trackingEventInput.Enabled.ValueBool()
			}

			if !trackingInput.Alert.IsUnknown() && !trackingInput.Alert.IsNull() {

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

		if !ruleInput.Schedule.IsUnknown() && !ruleInput.Schedule.IsNull() {

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
				if !itemExceptionsInput.ConnectionOrigin.IsUnknown() && !itemExceptionsInput.ConnectionOrigin.IsNull() {
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
					if !sourceInput.Subnet.IsUnknown() && !sourceInput.Subnet.IsNull() {
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
				if !itemExceptionsInput.DeviceOs.IsUnknown() && !itemExceptionsInput.DeviceOs.IsNull() {
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
					if !destinationInput.IP.IsUnknown() && !destinationInput.IP.IsNull() {
						diags = destinationInput.IP.ElementsAs(ctx, &exceptionInput.Destination.IP, false)
						resp.Diagnostics.Append(diags...)
					}

					// setting destination subnet
					if !destinationInput.Subnet.IsUnknown() && !destinationInput.Subnet.IsNull() {
						diags = destinationInput.Subnet.ElementsAs(ctx, &exceptionInput.Destination.Subnet, false)
						resp.Diagnostics.Append(diags...)
					}

					// setting destination domain
					if !destinationInput.Domain.IsUnknown() && !destinationInput.Domain.IsNull() {
						diags = destinationInput.Domain.ElementsAs(ctx, &exceptionInput.Destination.Domain, false)
						resp.Diagnostics.Append(diags...)
					}

					// setting destination fqdn
					if !destinationInput.Fqdn.IsUnknown() && !destinationInput.Fqdn.IsNull() {
						diags = destinationInput.Fqdn.ElementsAs(ctx, &exceptionInput.Destination.Fqdn, false)
						resp.Diagnostics.Append(diags...)
					}

					// setting destination remote asn
					if !destinationInput.RemoteAsn.IsUnknown() && !destinationInput.RemoteAsn.IsNull() {
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
							tflog.Info(ctx, "exception.itemServiceCustomInput.Port - "+fmt.Sprintf("%v", itemServiceCustomInput.Port))
							if !itemServiceCustomInput.Port.IsUnknown() && !itemServiceCustomInput.Port.IsNull() {
								tflog.Info(ctx, "Port is known and not null")

								var elementsPort []types.String
								diags = itemServiceCustomInput.Port.ElementsAs(ctx, &elementsPort, false)
								resp.Diagnostics.Append(diags...)

								inputPort := make([]cato_scalars.Port, 0, len(elementsPort))
								for _, item := range elementsPort {
									inputPort = append(inputPort, cato_scalars.Port(item.ValueString()))
								}
								customInput.Port = inputPort

							} else {
								tflog.Info(ctx, "Port is either unknown or null; skipping assignment")
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
		if !ruleInput.Description.IsNull() && !ruleInput.Description.IsUnknown() {
			input.Rule.Description = ruleInput.Description.ValueString()
		}
		input.Rule.Enabled = ruleInput.Enabled.ValueBool()
		input.Rule.Action = cato_models.InternetFirewallActionEnum(ruleInput.Action.ValueString())
		if !ruleInput.ConnectionOrigin.IsNull() && !ruleInput.ConnectionOrigin.IsUnknown() {
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

	// Read rule and hydrate response to state
	queryIfwPolicy := &cato_models.InternetFirewallPolicyInput{}
	body, err := r.client.catov2.PolicyInternetFirewall(ctx, queryIfwPolicy, r.client.AccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Catov2 API PolicyInternetFirewall error",
			err.Error(),
		)
		return
	}

	ruleList := body.GetPolicy().InternetFirewall.Policy.GetRules()
	currentRule := &cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_Rule{}
	// Get current rule from response by ID
	for _, ruleListItem := range ruleList {
		if ruleListItem.GetRule().ID == policyChange.GetPolicy().GetInternetFirewall().GetAddRule().Rule.GetRule().ID {
			currentRule = ruleListItem.GetRule()
			break
		}
	}
	// Hydrate ruleInput from api respoonse
	ruleInput := hydrateIfwRuleState(ctx, plan, currentRule)
	ruleInput.ID = types.StringValue(currentRule.ID)
	ruleObject, diags := types.ObjectValueFrom(ctx, InternetFirewallRuleRuleAttrTypes, ruleInput)
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

func (r *internetFwRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

	var state InternetFirewallRule
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

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
	rule := Policy_Policy_InternetFirewall_Policy_Rules_Rule{}
	diags = state.Rule.As(ctx, &rule, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ruleList := body.GetPolicy().InternetFirewall.Policy.GetRules()
	ruleExist := false
	currentRule := &cato_go_sdk.Policy_Policy_InternetFirewall_Policy_Rules_Rule{}
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
		tflog.Warn(ctx, "internet firewall rule not found, resource removed")
		resp.State.RemoveResource(ctx)
		return
	}

	ruleInput := hydrateIfwRuleState(ctx, state, currentRule)
	diags = resp.State.SetAttribute(ctx, path.Root("rule"), ruleInput)
	resp.Diagnostics.Append(diags...)
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
		if !sourceInput.IP.IsUnknown() && !sourceInput.IP.IsNull() {
			diags = sourceInput.IP.ElementsAs(ctx, &input.Rule.Source.IP, false)
			resp.Diagnostics.Append(diags...)
		}

		// setting source subnet
		if !sourceInput.Subnet.IsUnknown() && !sourceInput.Subnet.IsNull() {
			diags = sourceInput.Subnet.ElementsAs(ctx, &input.Rule.Source.Subnet, false)
			resp.Diagnostics.Append(diags...)
		}

		// setting source host
		if !sourceInput.Host.IsUnknown() && !sourceInput.Host.IsNull() {
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
		if !sourceInput.Site.IsUnknown() && !sourceInput.Site.IsNull() {
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
		if !sourceInput.IPRange.IsUnknown() && !sourceInput.IPRange.IsNull() {
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
	if !ruleInput.DeviceOs.IsUnknown() && !ruleInput.DeviceOs.IsNull() {
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
		if !destinationInput.IP.IsUnknown() && !destinationInput.IP.IsNull() {
			diags = destinationInput.IP.ElementsAs(ctx, &input.Rule.Destination.IP, false)
			resp.Diagnostics.Append(diags...)
		}

		// setting destination subnet
		if !destinationInput.Subnet.IsUnknown() && !destinationInput.Subnet.IsNull() {
			diags = destinationInput.Subnet.ElementsAs(ctx, &input.Rule.Destination.Subnet, false)
			resp.Diagnostics.Append(diags...)
		}

		// setting destination domain
		if !destinationInput.Domain.IsUnknown() && !destinationInput.Domain.IsNull() {
			diags = destinationInput.Domain.ElementsAs(ctx, &input.Rule.Destination.Domain, false)
			resp.Diagnostics.Append(diags...)
		}

		// setting destination fqdn
		if !destinationInput.Fqdn.IsUnknown() && !destinationInput.Fqdn.IsNull() {
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
				resp.Diagnostics.Append(diags...)

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
					resp.Diagnostics.Append(diags...)

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
		if !trackingInput.Alert.IsUnknown() && !trackingInput.Alert.IsNull() {

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
				if !sourceInput.IP.IsUnknown() && !sourceInput.IP.IsNull() {
					diags = sourceInput.IP.ElementsAs(ctx, &exceptionInput.Source.IP, false)
					resp.Diagnostics.Append(diags...)
				}

				// setting source subnet
				if !sourceInput.Subnet.IsUnknown() && !sourceInput.Subnet.IsNull() {
					diags = sourceInput.Subnet.ElementsAs(ctx, &exceptionInput.Source.Subnet, false)
					resp.Diagnostics.Append(diags...)
				}

				// setting source host
				if !sourceInput.Host.IsUnknown() && !sourceInput.Host.IsNull() {
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
						resp.Diagnostics.Append(diags...)

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
							resp.Diagnostics.Append(diags...)

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

// IFW Rule Schema validation fuctions

// Rule -> Alert Valildator
type ruleAlertValidator struct{}

func (v ruleAlertValidator) Description(ctx context.Context) string {
	return "If 'alert' is provided, both 'enabled' and 'frequency' must also be set, and must specify values for mailing_list, subscription_group, or web_hook."
}

func (v ruleAlertValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v ruleAlertValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	alertObj := req.ConfigValue
	alertMap := alertObj.Attributes()

	enabled, hasEnabled := alertMap["enabled"]
	frequency, hasFrequency := alertMap["frequency"]

	if !hasEnabled || enabled.IsNull() || enabled.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("alert").AtName("enabled"),
			"'enabled' is required",
			"If 'alert' is provided, the 'enabled' field must also be set.",
		)
	}

	if !hasFrequency || frequency.IsNull() || frequency.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("alert").AtName("frequency"),
			"'frequency' is required",
			"If 'alert' is provided, the 'frequency' field must also be set.",
		)
	}
}
