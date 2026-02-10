package provider

import (
	"context"

	cato_go_sdk "github.com/catonetworks/cato-go-sdk"
	"github.com/catonetworks/terraform-provider-cato/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &privAccessPolicyResource{}
	_ resource.ResourceWithConfigure   = &privAccessPolicyResource{}
	_ resource.ResourceWithImportState = &privAccessPolicyResource{}
)

func NewPrivAccessPolicyResource() resource.Resource {
	return &privAccessPolicyResource{}
}

type privAccessPolicyResource struct {
	client *catoClientData
}

func (r *privAccessPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_private_access_policy"
}

func (r *privAccessPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_private_access_policy` resource contains the configuration parameters for private access policies in the Cato platform.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "fake ID",
				Optional:    true,
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Is the private access policy enabled?",
				Required:    true,
			},
			"rules": r.schemaRules(),
			"sections": schema.ListNestedAttribute{
				Description: "List of sections for the private access policy",
				Optional:    true,
			},
			"audit": schema.SingleNestedAttribute{
				Description: "Audit log record",
				Optional:    true,
			},
			"revision": schema.SingleNestedAttribute{
				Description: "Private access policy revision",
				Computed:    true,
			},
		},
	}
}

func (r *privAccessPolicyResource) schemaRules() schema.ListNestedAttribute {
	rules := schema.ListNestedAttribute{
		Description: "List of rules for the private access policy",
		Optional:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"audit": schema.SingleNestedAttribute{
					Description: "Rule audit log record",
					Computed:    true,
					Attributes: map[string]schema.Attribute{
						"updated_by": schema.StringAttribute{
							Description: "Updated by",
							Computed:    true,
						},
						"updated_time": schema.StringAttribute{
							Description: "Update time",
							Computed:    true,
						},
					},
				},
				"rules": r.schemaRule(),
				"properties": schema.ListAttribute{
					Description: "Rule properties",
					Required:    true,
					ElementType: types.StringType,
				},
			},
		},
	}
	return rules
}

func (r *privAccessPolicyResource) schemaRule() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Private access rule details",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Rule ID",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Rule name",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Rule description",
				Required:    true,
			},
			"index": schema.Int64Attribute{
				Description: "Rule index",
				Required:    true,
			},
			"section": schema.SingleNestedAttribute{
				Description: "Settings for a policy section",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Description: "Section ID",
						Required:    true,
					},
					"name": schema.StringAttribute{
						Description: "Section name",
						Required:    true,
					},
					"subpolicy_id": schema.StringAttribute{
						Description: "Subpolicy ID",
						Optional:    true,
					},
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "TRUE = Rule is enabled FALSE = Rule is disabled",
				Required:    true,
			},
			"source": r.schemaSource(),
			"platforms": schema.ListAttribute{
				Description: "Platforms, operating systems",
				Required:    true,
				ElementType: types.StringType,
			},
			"countries": schema.ListNestedAttribute{
				Description: "Country name or id",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Country name",
							Optional:    true,
							Computed:    true,
						},
						"id": schema.StringAttribute{
							Description: "Country code",
							Optional:    true,
							Computed:    true,
						},
					},
				},
			},
			"applications": schema.ListNestedAttribute{
				Description: "Application name or id",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Application name",
							Optional:    true,
							Computed:    true,
						},
						"id": schema.StringAttribute{
							Description: "Application id",
							Optional:    true,
							Computed:    true,
						},
					},
				},
			},
			"connection_origin": schema.ListAttribute{
				Description: "Origin of the connection",
				Optional:    true,
				ElementType: types.StringType,
			},
			"action": schema.StringAttribute{
				Description: "ALLOW or BLOCK",
				Required:    true,
			},
			"tracking": schema.SingleNestedAttribute{
				Description: "Policy tracking",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"event": schema.SingleNestedAttribute{
						Description: "Event settings",
						Required:    true,
						Attributes: map[string]schema.Attribute{
							"enabled": schema.BoolAttribute{
								Description: "Event tracking enabled",
								Required:    true,
							},
						},
					},
					"alert": schema.SingleNestedAttribute{
						Description: "Alert settings",
						Required:    true,
						Attributes: map[string]schema.Attribute{
							"enabled": schema.BoolAttribute{
								Description: "TRUE – send alerts when the rule is matched, FALSE – don’t send alerts when the rule is matched",
								Required:    true,
							},
							"frequency": schema.StringAttribute{
								Description: "Frequency of an alert event for a rule",
								Required:    true,
							},

							"subscription_group": schema.ListNestedAttribute{
								Description: "Subscription group name or id",
								Optional:    true,
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Description: "name",
											Optional:    true,
											Computed:    true,
										},
										"id": schema.StringAttribute{
											Description: "ID",
											Optional:    true,
											Computed:    true,
										},
									},
								},
							},
							"webhook": schema.ListNestedAttribute{
								Description: "Webhook name or id",
								Optional:    true,
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Description: "name",
											Optional:    true,
											Computed:    true,
										},
										"id": schema.StringAttribute{
											Description: "ID",
											Optional:    true,
											Computed:    true,
										},
									},
								},
							},
							"mailing_list": schema.ListNestedAttribute{
								Description: "Mailing list name or id",
								Optional:    true,
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Description: "name",
											Optional:    true,
											Computed:    true,
										},
										"id": schema.StringAttribute{
											Description: "ID",
											Optional:    true,
											Computed:    true,
										},
									},
								},
							},
						},
					},
				},
			},
			"device": schema.ListNestedAttribute{
				Description: "Device group name or id",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "name",
							Optional:    true,
							Computed:    true,
						},
						"id": schema.StringAttribute{
							Description: "ID",
							Optional:    true,
							Computed:    true,
						},
					},
				},
			},
			"user_attributes": schema.SingleNestedAttribute{
				Description: "User attributes",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"risk_score": schema.SingleNestedAttribute{
						Description: "User's risk score settings",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"category": schema.StringAttribute{
								Description: "Risk score category",
								Required:    true,
							},
							"operator": schema.StringAttribute{
								Description: "Risk score operator",
								Required:    true,
							},
						},
					},
				},
			},
			"schedule": schema.SingleNestedAttribute{
				Description: "User attributes",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"active_on": schema.StringAttribute{
						Description: "Type of a time range when a rule is active",
						Required:    true,
					},
					"custom_recurring": schema.SingleNestedAttribute{
						Description: "Custom recurring time range that a rule is active",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"days": schema.ListAttribute{
								Description: "Days of the week",
								Required:    true,
								ElementType: types.StringType,
							},
							"from": schema.StringAttribute{
								Description: "From",
								Optional:    true,
							},
							"to": schema.StringAttribute{
								Description: "To",
								Optional:    true,
							},
						},
					},
					"custom_timeframe": schema.SingleNestedAttribute{
						Description: "Custom one-time time range that a rule is active",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"from": schema.StringAttribute{
								Description: "From",
								Required:    true,
							},
							"to": schema.StringAttribute{
								Description: "To",
								Required:    true,
							},
						},
					},
				},
			},
			"active_period": schema.SingleNestedAttribute{
				Description: "Time period during which the rule is active",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"effective_from": schema.StringAttribute{
						Description: "Effective from",
						Optional:    true,
					},
					"expires_at": schema.StringAttribute{
						Description: "Expires at",
						Optional:    true,
					},
					"use_effective_from": schema.BoolAttribute{
						Description: "Use effective from",
						Required:    true,
					},
					"use_expires_at": schema.BoolAttribute{
						Description: "Use expires at",
						Required:    true,
					},
				},
			},
		},
	}
}

func (r *privAccessPolicyResource) schemaSource() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Source",
		Required:    true,
		Attributes: map[string]schema.Attribute{
			"user": schema.ListNestedAttribute{
				Description: "User",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "User ID",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "User name",
							Required:    true,
						},
					},
				},
			},
			"users_group": schema.ListNestedAttribute{
				Description: "Group",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Group ID",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Group name",
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func (r *privAccessPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*catoClientData)
}

func (r *privAccessPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *privAccessPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// var plan PrivAccessPolicyModel
}

func (r *privAccessPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *privAccessPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PrivAccessPolicyModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	hydratedState, hydrateErr := r.hydratePrivAccessPolicyState(ctx, state)
	if hydrateErr != nil {
		// Check if app-connector not found
		if hydrateErr.Error() == "app_connector not found" { // TODO: check the actual error
			tflog.Warn(ctx, "app_connector not found, resource removed")
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error hydrating group state",
			hydrateErr.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, &hydratedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *privAccessPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

// hydratePrivAccessPolicyState fetches the current state of a privAccessPolicy from the API
// It takes a plan parameter to match config members with API members correctly
func (r *privAccessPolicyResource) hydratePrivAccessPolicyState(ctx context.Context, plan PrivAccessPolicyModel) (*PrivAccessPolicyModel, error) {
	// Call Cato API to get a connector
	result, err := r.client.catov2.PolicyReadPrivateAccessPolicy(ctx, r.client.AccountId)
	tflog.Debug(ctx, "PrivateAppReadPrivateApp", map[string]interface{}{
		"response": utils.InterfaceToJSONString(result),
	})
	if err != nil {
		return nil, err
	}

	// Map API response to PrivAccessPolicyModel
	policy := result.GetPolicy().GetPrivateAccess().GetPolicy()
	state := &PrivAccessPolicyModel{
		Enabled: types.BoolValue(policy.Enabled),
		Rules:   r.parseRules(policy),
	}

	return state, nil
}

func (r *privAccessPolicyResource) parseTracking(t cato_go_sdk.PolicyReadPrivateAccessPolicy_Policy_PrivateAccess_Policy_Rules_Rule_Tracking) *Tracking {
	alert := &t.Alert
	out := Tracking{
		Event: PolicyRuleTrackingEvent{Enabled: types.BoolValue(t.Event.Enabled)},
		Alert: PoliciRuleTrackingAlert{
			Enabled:           types.BoolValue(alert.Enabled),
			Frequency:         types.StringValue(string(alert.Frequency)),
			MailingList:       parseIDRef(alert.MailingList),
			SubscriptionGroup: parseIDRef(alert.MailingList),
			Webhook:           parseIDRef(alert.Webhook),
		},
	}
	return &out
}
func (r *privAccessPolicyResource) parsePolicySchedule(s cato_go_sdk.PolicyReadPrivateAccessPolicy_Policy_PrivateAccess_Policy_Rules_Rule_Schedule) *PolicySchedule {
	p := PolicySchedule{
		ActiveOn: types.StringValue(string(s.ActiveOn)),
	}
	if s.CustomRecurring != nil {
		cr := s.CustomRecurring
		p.CustomRecurring = PolicyCustomRecurring{
			Days: parseStringList(cr.Days),
			From: types.StringValue(string(cr.From)),
			To:   types.StringValue(string(cr.To)),
		}
	}
	if s.CustomTimeframe != nil {
		tf := s.CustomTimeframe
		p.CustomTimeframe = PolicyCustomTimeframe{
			From: types.StringValue(string(tf.From)),
			To:   types.StringValue(string(tf.To)),
		}
	}
	return &p
}

func (r *privAccessPolicyResource) parsePolicyActivePeriod(ap cato_go_sdk.PolicyReadPrivateAccessPolicy_Policy_PrivateAccess_Policy_Rules_Rule_ActivePeriod) *PolicyRuleActivePeriod {
	p := PolicyRuleActivePeriod{
		EffectiveFrom:    types.StringPointerValue(ap.EffectiveFrom),
		ExpiresAt:        types.StringPointerValue(ap.ExpiresAt),
		UseEffectiveFrom: types.BoolValue(ap.UseEffectiveFrom),
		UseExpiresAt:     types.BoolValue(ap.UseExpiresAt),
	}
	return &p
}

func (r *privAccessPolicyResource) parseRules(policy *cato_go_sdk.PolicyReadPrivateAccessPolicy_Policy_PrivateAccess_Policy) []PrivAccessPolicyRule {
	var out []PrivAccessPolicyRule
	for _, policyRule := range policy.Rules {
		rule := &policyRule.Rule
		sRule := PrivAccessPolicyRule{
			Audit: nil,
			Rule: PrivateAccessRule{
				ID:          types.StringValue(rule.ID),
				Name:        types.StringValue(rule.Name),
				Description: types.StringValue(rule.Description),
				Index:       types.Int64Value(rule.Index),
				Section: &PrivAccessPolicySection{
					ID:          types.StringValue(rule.Section.ID),
					Name:        types.StringValue(rule.Section.Name),
					SubpolicyID: types.StringPointerValue(rule.Section.SubPolicyID),
				},
				Enabled:          types.BoolValue(rule.Enabled),
				Source:           &Source{User: parseIDRef(rule.Source.User), UsersGroup: parseIDRef(rule.Source.UsersGroup)},
				Platforms:        parseStringList(rule.Platform),
				Country:          parseIDRef(rule.Country),
				Applications:     parseIDRef(rule.Applications.Application),
				ConnectionOrigin: parseStringList(rule.ConnectionOrigin),
				Action:           types.StringValue(string(rule.Action.Action)),
				Tracking:         r.parseTracking(rule.Tracking),
				Device:           parseIDRef(rule.Device),
				UserAttributes: &UserAttributes{RiskScore: RiskScore{
					Category: types.StringValue(string(rule.UserAttributes.RiskScore.Category)),
					Operator: types.StringValue(string(rule.UserAttributes.RiskScore.Operator)),
				}},
				Schedule:     r.parsePolicySchedule(rule.Schedule),
				ActivePeriod: r.parsePolicyActivePeriod(rule.ActivePeriod),
			},
			Properties: nil,
		}
		out = append(out, sRule)
	}
	return out
}
