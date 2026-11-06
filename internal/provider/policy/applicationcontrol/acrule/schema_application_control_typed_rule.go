package acrule

import (
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/planmodifiers"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/applicationcontrol"
	idname "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	utils "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
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
						stringvalidator.OneOf(utils.ToStrings(cato_models.AllPolicyRuleTrackingFrequencyEnum)...),
					},
					Default: stringdefault.StaticString("HOURLY"),
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"subscription_group": schema.SetNestedAttribute{
					Optional:      true,
					Computed:      true,
					NestedObject:  schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Subscription group")},
					PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
				},
				"webhook": schema.SetNestedAttribute{
					Optional:      true,
					Computed:      true,
					NestedObject:  schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Webhook")},
					PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
				},
				"mailing_list": schema.SetNestedAttribute{
					Optional:      true,
					Computed:      true,
					NestedObject:  schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Mailing list")},
					PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
				},
			},
		},
	}
}

// applicationControlTypedRuleSchemaAttributes defines application_rule / data_rule / file_rule nested blocks.
//
//nolint:funlen
func applicationControlTypedRuleSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"action": schema.StringAttribute{
			Description: "Rule action",
			Required:    true,
			Validators: []validator.String{
				stringvalidator.OneOf(utils.ToStrings(cato_models.AllApplicationControlAction)...),
			},
		},
		"severity": schema.StringAttribute{
			Description: "Rule severity",
			Required:    true,
			Validators: []validator.String{
				stringvalidator.OneOf(utils.ToStrings(cato_models.AllApplicationControlSeverity)...),
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
			Attributes: applicationcontrol.ApplicationControlSourceSchemaAttributes(),
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
			Description:   "Device profiles",
			Optional:      true,
			Computed:      true,
			NestedObject:  schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Device")},
			PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
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
						Required:      true,
						Attributes:    idname.SchemaNameID("Activity"),
						PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
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
				stringvalidator.OneOf(utils.ToStrings(cato_models.AllApplicationControlSatisfy)...),
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
					Optional:      true,
					Computed:      true,
					NestedObject:  schema.NestedAttributeObject{Attributes: idname.SchemaNameID("User notification")},
					PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
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
				stringvalidator.OneOf(utils.ToStrings(cato_models.AllApplicationControlSatisfy)...),
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
					Optional:      true,
					Computed:      true,
					NestedObject:  schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Content profile")},
					PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
				},
				"edm_profile": schema.SetNestedAttribute{
					Optional:      true,
					Computed:      true,
					NestedObject:  schema.NestedAttributeObject{Attributes: idname.SchemaNameID("EDM profile")},
					PlanModifiers: []planmodifier.Set{idname.SetPlanModifier()},
				},
			},
		},
	}
}
