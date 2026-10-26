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
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policyrule"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

func (r *alwaysOnRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The cato_always_on_rule resource manages an Always On policy rule in the Cato platform. " +
			"Changes are saved to the draft policy without publishing. Changing at moves the rule; " +
			"omitting at preserves its current position.",
		Attributes: map[string]schema.Attribute{
			"at":   policyrule.SchemaPosition(),
			"rule": schemaAlwaysOnRule(),
		},
	}
}

func schemaAlwaysOnRule() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Parameters for the Always On policy rule",
		Required:    true,
		Attributes: map[string]schema.Attribute{
			"action": schema.StringAttribute{
				Description: "Action applied when the rule matches (ENFORCE or IGNORE; API default is ENFORCE)",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(utils.ToStrings(cato_models.AllZtnaAlwaysOnRuleActionEnum)...),
				},
			},
			"allow_fail_open": schema.BoolAttribute{
				Description: "Allow direct internet access until a connection to the PoP is established (API default is false)",
				Required:    true,
			},
			"allow_user_bypass": schema.BoolAttribute{
				Description: "Allow users to bypass Always On (API default is false)",
				Required:    true,
			},
			"anti_tamper_mode": schema.StringAttribute{
				Description: "Level of protection against tampering (OFF, MONITOR, or PROTECT; API default is OFF)",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(utils.ToStrings(cato_models.AllAntiTamperModeEnum)...),
				},
			},
			"bypass_duration": schemaAlwaysOnRuleBypassDuration(),
			"description": schema.StringAttribute{
				Description: "Rule description (API default is an empty string)",
				Required:    true,
			},
			"device_posture_profile": schemaAlwaysOnRuleReferences("Device posture profile",
				"Device posture profiles matched with logical OR; combined with other criteria using logical AND "+
					"(API default is an empty collection)"),
			"enabled": schema.BoolAttribute{
				Description: "Whether the rule is enabled",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description:   "Always On policy rule ID",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"index": schema.Int64Attribute{
				Description: "Position or priority of the rule in the policy",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Rule name",
				Required:    true,
			},
			"platform": schema.SetAttribute{
				Description: "Operating systems matched with logical OR (ANDROID, CHROMEOS, EMBEDDED, IOS, LINUX, MACOS, or WINDOWS); " +
					"combined with other criteria using logical AND (API default is an empty collection)",
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{setvalidator.ValueStringsAre(
					stringvalidator.OneOf(utils.ToStrings(cato_models.AllOperatingSystem)...),
				)},
			},
			"section": schemaAlwaysOnRuleSection(),
			"source":  schemaAlwaysOnRuleSource(),
		},
	}
}

func schemaAlwaysOnRuleBypassDuration() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Duration of the disconnection allowed when bypassing Always On",
		Required:    true,
		Attributes: map[string]schema.Attribute{
			"time": schema.Int64Attribute{
				Description: "Bypass duration value",
				Required:    true,
			},
			"unit": schema.StringAttribute{
				Description: "Bypass duration unit (HOURS or MINUTES; API default is MINUTES)",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(utils.ToStrings(cato_models.AllZtnaAlwaysOnTimeUnit)...),
				},
			},
		},
	}
}

func schemaAlwaysOnRuleSection() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Policy section containing the rule",
		Computed:    true,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "Policy section ID",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Description: "Policy section name",
				Computed:    true,
			},
		},
	}
}

func schemaAlwaysOnRuleSource() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Source users and user groups matched with logical OR within each criterion and logical AND between criteria " +
			"(API default contains empty user and user_groups collections)",
		Required: true,
		Attributes: map[string]schema.Attribute{
			"users": schemaAlwaysOnRuleReferences("User",
				"Individual users defined for the account (API default is an empty collection)"),
			"user_groups": schemaAlwaysOnRuleReferences("User group",
				"User groups defined for the account (API default is an empty collection)"),
		},
	}
}

func schemaAlwaysOnRuleReferences(prefix, description string) schema.SetNestedAttribute {
	return schema.SetNestedAttribute{
		Description: description,
		Required:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes:    idname.SchemaNameID(prefix),
			PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
		},
	}
}
