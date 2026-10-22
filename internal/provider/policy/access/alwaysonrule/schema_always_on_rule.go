package alwaysonrule

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/planmodifiers"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/validators"
)

// alwaysOnRuleResource currently defines the schema; CRUD will be added separately.
type alwaysOnRuleResource struct{}

func (r *alwaysOnRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The cato_always_on_rule resource configures an Always On policy rule in the Cato platform.",
		Attributes: map[string]schema.Attribute{
			"at":   schemaAlwaysOnRulePosition(),
			"rule": schemaAlwaysOnRule(),
		},
	}
}

func schemaAlwaysOnRulePosition() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Position of the new rule in the policy.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"position": schema.StringAttribute{
				Description: "Insertion position relative to a policy, section, or another rule. Empty means unspecified.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{stringvalidator.Any(
					stringvalidator.OneOf(""),
					validators.StringValidator[cato_models.PolicyRulePositionEnum]("position", cato_models.AllPolicyRulePositionEnum),
				)},
				PlanModifiers: []planmodifier.String{planmodifiers.NullStringModifier()},
			},
			"ref": schema.StringAttribute{
				Description:   "ID of the rule or section relative to which the new rule is positioned.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{planmodifiers.NullStringModifier()},
			},
		},
	}
}

func schemaAlwaysOnRule() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Parameters for the Always On policy rule.",
		Required:    true,
		Attributes: map[string]schema.Attribute{
			"action": schema.StringAttribute{
				Description: "Action applied when the rule matches: ENFORCE or IGNORE.",
				Required:    true,
				Validators: []validator.String{
					validators.StringValidator[cato_models.ZtnaAlwaysOnRuleActionEnum]("action", cato_models.AllZtnaAlwaysOnRuleActionEnum),
				},
			},
			"allow_fail_open": schema.BoolAttribute{
				Description: "Allow direct internet access until a connection to the PoP is established.",
				Required:    true,
			},
			"allow_user_bypass": schema.BoolAttribute{
				Description: "Allow users to bypass Always On.",
				Required:    true,
			},
			"anti_tamper_mode": schema.StringAttribute{
				Description: "Level of protection against tampering: OFF, MONITOR, or PROTECT. Empty means unspecified (API default: OFF).",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{stringvalidator.Any(
					stringvalidator.OneOf(""),
					validators.StringValidator[cato_models.AntiTamperModeEnum]("anti_tamper_mode", cato_models.AllAntiTamperModeEnum),
				)},
				PlanModifiers: []planmodifier.String{planmodifiers.NullStringModifier()},
			},
			"bypass_duration": schemaAlwaysOnRuleBypassDuration(),
			"description": schema.StringAttribute{
				Description: "Rule description.",
				Required:    true,
			},
			"device_posture_profile": schemaAlwaysOnRuleReferences("Device posture profile"),
			"enabled": schema.BoolAttribute{
				Description: "Whether the rule is enabled.",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description:   "Always On policy rule ID.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"index": schema.Int64Attribute{
				Description: "Position or priority of the rule in the policy.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Rule name.",
				Required:    true,
			},
			"platform": schema.SetAttribute{
				Description: "Operating systems matched with logical OR; combined with other criteria using logical AND.",
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{setvalidator.ValueStringsAre(
					validators.StringValidator[cato_models.OperatingSystem]("platform", cato_models.AllOperatingSystem),
				)},
			},
			"section": schemaAlwaysOnRuleSection(),
			"source":  schemaAlwaysOnRuleSource(),
		},
	}
}

func schemaAlwaysOnRuleBypassDuration() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Duration of the disconnection allowed when bypassing Always On.",
		Required:    true,
		Attributes: map[string]schema.Attribute{
			"time": schema.Int64Attribute{
				Description: "Bypass duration value.",
				Required:    true,
			},
			"unit": schema.StringAttribute{
				Description: "Bypass duration unit: HOURS or MINUTES.",
				Required:    true,
				Validators: []validator.String{
					validators.StringValidator[cato_models.ZtnaAlwaysOnTimeUnit]("unit", cato_models.AllZtnaAlwaysOnTimeUnit),
				},
			},
		},
	}
}

func schemaAlwaysOnRuleSection() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Policy section containing the rule.",
		Computed:    true,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "Policy section ID.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Description: "Policy section name.",
				Computed:    true,
			},
		},
	}
}

func schemaAlwaysOnRuleSource() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Source users and user groups matched with logical OR within each criterion and logical AND between criteria.",
		Required:    true,
		Attributes: map[string]schema.Attribute{
			"user":        schemaAlwaysOnRuleReferences("User"),
			"users_group": schemaAlwaysOnRuleReferences("User group"),
		},
	}
}

func schemaAlwaysOnRuleReferences(description string) schema.SetNestedAttribute {
	return schema.SetNestedAttribute{
		Description: description + " references matched with logical OR; combined with other criteria using logical AND.",
		Required:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes:    idname.SchemaNameID(description),
			PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
		},
	}
}
