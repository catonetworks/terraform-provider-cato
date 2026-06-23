package provider

import (
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/planmodifiers"
)

func applicationControlScheduleSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"active_on": schema.StringAttribute{
			Optional: true,
			Computed: true,
		},
		"custom_timeframe": schema.SingleNestedAttribute{
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"from": schema.StringAttribute{Optional: true},
				"to":   schema.StringAttribute{Optional: true},
			},
		},
		"custom_recurring": schema.SingleNestedAttribute{
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"from": schema.StringAttribute{Optional: true},
				"to":   schema.StringAttribute{Optional: true},
				"days": schema.ListAttribute{
					ElementType: types.StringType,
					Optional:    true,
				},
			},
		},
	}
}

func applicationControlTrackingSchemaAttributes() map[string]schema.Attribute {
	nameID := schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRelative().AtParent().AtName("id"),
					}...),
				},
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"id": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
	return map[string]schema.Attribute{
		"event": schema.SingleNestedAttribute{
			Optional: true,
			Computed: true,
			Attributes: map[string]schema.Attribute{
				"enabled": schema.BoolAttribute{
					Optional: true,
					Computed: true,
					Default:  booldefault.StaticBool(false),
				},
			},
		},
		"alert": schema.SingleNestedAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.UseStateForUnknown(),
			},
			Attributes: map[string]schema.Attribute{
				"enabled": schema.BoolAttribute{
					Optional: true,
					Computed: true,
					Default:  booldefault.StaticBool(false),
				},
				"frequency": schema.StringAttribute{
					Optional: true,
					Computed: true,
					Validators: []validator.String{
						stringvalidator.OneOf("DAILY", "HOURLY", "IMMEDIATE", "WEEKLY"),
					},
					Default: stringdefault.StaticString("HOURLY"),
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"subscription_group": schema.SetNestedAttribute{
					Optional:     true,
					Computed:     true,
					NestedObject: nameID,
					PlanModifiers: []planmodifier.Set{
						setplanmodifier.UseStateForUnknown(),
					},
				},
				"webhook": schema.SetNestedAttribute{
					Optional:     true,
					Computed:     true,
					NestedObject: nameID,
					PlanModifiers: []planmodifier.Set{
						setplanmodifier.UseStateForUnknown(),
					},
				},
				"mailing_list": schema.SetNestedAttribute{
					Optional:     true,
					Computed:     true,
					NestedObject: nameID,
					PlanModifiers: []planmodifier.Set{
						setplanmodifier.UseStateForUnknown(),
					},
				},
			},
		},
	}
}

// applicationControlTypedRuleSchemaAttributes defines application_rule / data_rule / file_rule nested blocks.
//
//nolint:funlen
func applicationControlTypedRuleSchemaAttributes() map[string]schema.Attribute {
	deviceNested := schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRelative().AtParent().AtName("id"),
					}...),
				},
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"id": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
	return map[string]schema.Attribute{
		"action": schema.StringAttribute{
			Description: "Rule action",
			Required:    true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(cato_models.ApplicationControlActionAllow),
					string(cato_models.ApplicationControlActionBlock),
					string(cato_models.ApplicationControlActionMonitor),
				),
			},
		},
		"severity": schema.StringAttribute{
			Description: "Rule severity",
			Required:    true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(cato_models.ApplicationControlSeverityHigh),
					string(cato_models.ApplicationControlSeverityMedium),
					string(cato_models.ApplicationControlSeverityLow),
				),
			},
		},
		"schedule": schema.SingleNestedAttribute{
			Description: "Schedule for the typed rule",
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.UseStateForUnknown(),
			},
			Attributes: applicationControlScheduleSchemaAttributes(),
		},
		"source": schema.SingleNestedAttribute{
			Description: "Source traffic matching criteria",
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.UseStateForUnknown(),
				planmodifiers.SourceDestObjectModifier(),
			},
			Attributes: applicationControlSourceSchemaAttributes(),
		},
		"tracking": schema.SingleNestedAttribute{
			Description: "Tracking configuration",
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.UseStateForUnknown(),
			},
			Attributes: applicationControlTrackingSchemaAttributes(),
		},
		"device": schema.SetNestedAttribute{
			Description:  "Device profiles",
			Optional:     true,
			Computed:     true,
			NestedObject: deviceNested,
			PlanModifiers: []planmodifier.Set{
				setplanmodifier.UseStateForUnknown(),
			},
		},
		"access_method": schema.ListNestedAttribute{
			Description: "Access method rows",
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.List{
				listplanmodifier.UseStateForUnknown(),
			},
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"access_method": schema.StringAttribute{Required: true},
					"operator":      schema.StringAttribute{Required: true},
					"value":         schema.StringAttribute{Optional: true, Computed: true},
				},
			},
		},
		"application": schema.SingleNestedAttribute{
			Description: "Application matching (WAN-shaped object)",
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.UseStateForUnknown(),
			},
			Attributes: wanApplicationSchemaAttributes(),
		},
		"application_activity": schema.ListNestedAttribute{
			Description: "Application activities to match (required for file/data rules when an application is specified)",
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.List{
				listplanmodifier.UseStateForUnknown(),
			},
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"activity": schema.SingleNestedAttribute{
						Required: true,
						Attributes: map[string]schema.Attribute{
							"id": schema.StringAttribute{
								Optional: true,
								Computed: true,
								Validators: []validator.String{
									stringvalidator.ConflictsWith(path.Expressions{
										path.MatchRelative().AtParent().AtName("name"),
									}...),
								},
								PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
							},
							"name": schema.StringAttribute{
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
							},
						},
					},
				},
			},
		},
		"application_activity_satisfy": schema.StringAttribute{
			Description: "How application activities are combined (ALL | ANY)",
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString("ALL"),
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(cato_models.ApplicationControlSatisfyAll),
					string(cato_models.ApplicationControlSatisfyAny),
				),
			},
		},
		"action_config": schema.SingleNestedAttribute{
			Description: "Action configuration (e.g. user notifications)",
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.UseStateForUnknown(),
			},
			Attributes: map[string]schema.Attribute{
				"user_notification": schema.SetNestedAttribute{
					Optional: true,
					Computed: true,
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{
								Optional: true,
								Computed: true,
								Validators: []validator.String{
									stringvalidator.ConflictsWith(path.Expressions{
										path.MatchRelative().AtParent().AtName("id"),
									}...),
								},
								PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
							},
							"id": schema.StringAttribute{
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
							},
						},
					},
				},
			},
		},
		"file_attribute": schema.ListNestedAttribute{
			Description: "File attribute rows (data/file rules)",
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.List{
				listplanmodifier.UseStateForUnknown(),
			},
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"file_attribute": schema.StringAttribute{Required: true},
					"operator":       schema.StringAttribute{Required: true},
					"value":          schema.StringAttribute{Optional: true, Computed: true},
				},
			},
		},
		"file_attribute_satisfy": schema.StringAttribute{
			Description: "How file attributes are combined",
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(cato_models.ApplicationControlSatisfyAll),
					string(cato_models.ApplicationControlSatisfyAny),
				),
			},
		},
		"dlp_profile": schema.SingleNestedAttribute{
			Description: "DLP profile (data rules)",
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.UseStateForUnknown(),
			},
			Attributes: map[string]schema.Attribute{
				"content_profile": schema.SetNestedAttribute{
					Optional: true,
					Computed: true,
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{
								Optional: true,
								Computed: true,
								Validators: []validator.String{
									stringvalidator.ConflictsWith(path.Expressions{
										path.MatchRelative().AtParent().AtName("id"),
									}...),
								},
								PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
							},
							"id": schema.StringAttribute{
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
							},
						},
					},
				},
				"edm_profile": schema.SetNestedAttribute{
					Optional: true,
					Computed: true,
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{
								Optional: true,
								Computed: true,
								Validators: []validator.String{
									stringvalidator.ConflictsWith(path.Expressions{
										path.MatchRelative().AtParent().AtName("id"),
									}...),
								},
								PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
							},
							"id": schema.StringAttribute{
								Optional:      true,
								Computed:      true,
								PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
							},
						},
					},
				},
			},
		},
	}
}
