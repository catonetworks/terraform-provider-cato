package parule

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/validators"
)

func (r *privAccessRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_private_access_rule` resource contains the configuration parameters for private " +
			"access policy rule in the Cato platform.",
		Attributes: map[string]schema.Attribute{
			"action": schema.StringAttribute{
				Description: "ALLOW or BLOCK",
				Required:    true,
				Validators:  []validator.String{validators.PrivAccPolicyActionValidator{}},
			},
			"active_period": r.schemaActivePeriod(),
			"applications": schema.SetNestedAttribute{
				Description: "Application name or id",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes:    idname.SchemaNameID("Application"),
					PlanModifiers: []planmodifier.Object{idname.PlanModifier()},
				},
			},
			"connection_origins": schema.SetAttribute{
				Description:   "Origin of the connection",
				Optional:      true,
				Computed:      true,
				ElementType:   types.StringType,
				Validators:    []validator.Set{validators.PrivAccPolicyConnOriginValidator{}},
				PlanModifiers: []planmodifier.Set{setplanmodifier.UseStateForUnknown()},
			},
			"countries": schema.SetNestedAttribute{
				Description: "List of countries",
				Optional:    true,
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes:    idname.SchemaNameID("Country"),
					PlanModifiers: []planmodifier.Object{idname.PlanModifier()},
				},
				PlanModifiers: []planmodifier.Set{setplanmodifier.UseStateForUnknown()},
			},
			"description": schema.StringAttribute{
				Description:   "Rule description",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"devices": schema.SetNestedAttribute{
				Description: "List of devices",
				Optional:    true,
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes:    idname.SchemaNameID("Device"),
					PlanModifiers: []planmodifier.Object{idname.PlanModifier()},
				},
				PlanModifiers: []planmodifier.Set{setplanmodifier.UseStateForUnknown()},
			},
			"enabled": schema.BoolAttribute{
				Description: "TRUE = Rule is enabled FALSE = Rule is disabled",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description:   "Rule ID",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Description: "Rule name",
				Required:    true,
			},
			"platforms": schema.SetAttribute{
				Description:   "Platforms, operating systems",
				Optional:      true,
				Computed:      true,
				ElementType:   types.StringType,
				PlanModifiers: []planmodifier.Set{setplanmodifier.UseStateForUnknown()},
				Validators:    []validator.Set{validators.PlatformValidator{}},
			},
			"schedule":        r.schemaSchedule(),
			"source":          r.schemaSource(),
			"tracking":        r.schemaTracking(),
			"user_attributes": r.schemaUserAttributes(),
			// index is only used in bulk operation
			// "section": r.schemaSection(), -- not available in the 1st phase
		},
	}
}

func (r *privAccessRuleResource) schemaTracking() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description:   "Rule tracking",
		Optional:      true,
		Computed:      true,
		PlanModifiers: []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
		Attributes: map[string]schema.Attribute{
			"alert": schema.SingleNestedAttribute{
				Description:   "Alert settings",
				Required:      true,
				PlanModifiers: []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "TRUE – send alerts when the rule is matched, FALSE – don’t send alerts when the rule is matched",
						Required:    true,
					},
					"frequency": schema.StringAttribute{
						Description: "Frequency of an alert event for a rule",
						Required:    true,
						Validators:  []validator.String{validators.PolicyTrackingFrequency{}},
					},
					"mailing_list": schema.SetNestedAttribute{
						Description: "Mailing list name or id",
						Optional:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes:    idname.SchemaNameID("Mailing list"),
							PlanModifiers: []planmodifier.Object{idname.PlanModifier()},
						},
					},
					"subscription_group": schema.SetNestedAttribute{
						Description: "Subscription group name or id",
						Optional:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes:    idname.SchemaNameID("Subscription group"),
							PlanModifiers: []planmodifier.Object{idname.PlanModifier()},
						},
					},
					"webhook": schema.SetNestedAttribute{
						Description: "Webhook name or id",
						Optional:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes:    idname.SchemaNameID("Webhook"),
							PlanModifiers: []planmodifier.Object{idname.PlanModifier()},
						},
					},
				},
			},
			"event": schema.SingleNestedAttribute{
				Description: "Event settings",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description:   "Event tracking enabled",
						Required:      true,
						PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
					},
				},
			},
		},
	}
}

func (r *privAccessRuleResource) schemaUserAttributes() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description:   "User attributes",
		Optional:      true,
		Computed:      true,
		PlanModifiers: []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
		Attributes: map[string]schema.Attribute{
			"risk_score": schema.SingleNestedAttribute{
				Description:   "User's risk score settings",
				Optional:      true,
				PlanModifiers: []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
				Attributes: map[string]schema.Attribute{
					"category": schema.StringAttribute{
						Description:   "Risk score category",
						Required:      true,
						PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
						Validators:    []validator.String{validators.RiscScoreCategory{}},
					},
					"operator": schema.StringAttribute{
						Description:   "Risk score operator",
						Required:      true,
						PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
						Validators:    []validator.String{validators.RiskScoreOperator{}},
					},
				},
			},
		},
	}
}

func (r *privAccessRuleResource) schemaSchedule() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Schedule",
		Optional:    true,
		Computed:    true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.UseStateForUnknown(),
		},
		Attributes: map[string]schema.Attribute{
			"active_on": schema.StringAttribute{
				Description: "Type of a time range when a rule is active",
				Required:    true,
				Validators:  []validator.String{validators.PolicyActiveOnValidator{}},
			},
			"custom_recurring": schema.SingleNestedAttribute{
				Description: "Custom recurring time range that a rule is active",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"days": schema.SetAttribute{
						Description: "Days of the week",
						Required:    true,
						ElementType: types.StringType,
						Validators:  []validator.Set{validators.DaysValidator{}},
					},
					"from": schema.StringAttribute{
						Description: "From time (12:34)",
						Required:    true,
					},
					"to": schema.StringAttribute{
						Description: "To time (12:34)",
						Required:    true,
					},
				},
			},
			"custom_timeframe": schema.SingleNestedAttribute{
				Description: "Custom one-time time range that a rule is active",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"from": schema.StringAttribute{
						Description: "From datetime (2006-01-02T15:04:05Z)",
						Required:    true,
						Validators:  []validator.String{validators.DateTimeValidator{}},
					},
					"to": schema.StringAttribute{
						Description: "To datetime (2006-01-02T15:04:05Z)",
						Required:    true,
						Validators:  []validator.String{validators.DateTimeValidator{}},
					},
				},
			},
		},
	}
}

func (r *privAccessRuleResource) schemaActivePeriod() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Time period during which the rule is active",
		Optional:    true,
		Computed:    true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.UseStateForUnknown(),
		},
		Attributes: map[string]schema.Attribute{
			"effective_from": schema.StringAttribute{
				Description: "Effective from (2006-01-02T15:04:05Z)",
				Optional:    true,
				Validators:  []validator.String{validators.DateTimeValidator{}},
			},
			"expires_at": schema.StringAttribute{
				Description: "Expires at (2006-01-02T15:04:05Z)",
				Optional:    true,
				Validators:  []validator.String{validators.DateTimeValidator{}},
			},
			"use_effective_from": schema.BoolAttribute{
				Description: "Use effective from",
				Optional:    true,
			},
			"use_expires_at": schema.BoolAttribute{
				Description: "Use expires at",
				Optional:    true,
			},
		},
	}
}

func (r *privAccessRuleResource) schemaSource() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Source",
		Required:    true,
		Attributes: map[string]schema.Attribute{
			"users": schema.SetNestedAttribute{
				Description: "Users",
				Optional:    true,
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes:    idname.SchemaNameID("User"),
					PlanModifiers: []planmodifier.Object{idname.PlanModifier()},
				},
				PlanModifiers: []planmodifier.Set{setplanmodifier.UseStateForUnknown()},
			},
			"user_groups": schema.SetNestedAttribute{
				Description: "User groups",
				Optional:    true,
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes:    idname.SchemaNameID("Group"),
					PlanModifiers: []planmodifier.Object{idname.PlanModifier()},
				},
				PlanModifiers: []planmodifier.Set{setplanmodifier.UseStateForUnknown()},
			},
		},
		Validators: []validator.Object{validators.PrivAccPolicySourceValidator{}},
	}
}
