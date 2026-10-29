package clientconnectivityrule

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
	policyshared "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/policy"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

func (r *policyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a client connectivity rule. Changes are automatically published. " +
			"Omitting at on update " +
			"preserves placement.",
		Attributes: schemaResource(),
	}
}

func schemaResource() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"at": schema.SingleNestedAttribute{
			Description: "Position of the rule in the policy",
			Optional:    true,
			Attributes: policyshared.RulePositionAttributes("Position relative to a policy, a section or another rule",
				"The identifier of the object (e.g. a rule, a section) relative to which "+
					"the position of the added rule is defined"),
		},
		"rule": schema.SingleNestedAttribute{
			Description: "Parameters for the rule you are adding",
			Required:    true,
			Attributes:  schemaResourceRule(),
		},
	}
}

func schemaResourceRule() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"action": schema.StringAttribute{
			Description: "The action applied by the client connectivity if the rule is matched",
			Required:    true,
			Validators:  []validator.String{stringvalidator.OneOf(utils.ToStrings(cato_models.AllClientConnectivityActionEnum)...)},
		},
		"confidence_level": schema.StringAttribute{
			Description: "User confidence level",
			Required:    true,
			Validators:  []validator.String{stringvalidator.OneOf(utils.ToStrings(cato_models.AllClientConnectivityConfidenceLevelEnum)...)},
		},
		"connection_origin": schema.SetAttribute{
			Description: "Connection origin matching criteria. Logical 'OR' is applied within the " +
				"criteria set. Logical 'AND' " +
				"is applied between criteria sets.",
			Required:    true,
			ElementType: types.StringType,
			Validators: []validator.Set{
				setvalidator.ValueStringsAre(
					stringvalidator.OneOf(utils.ToStrings(cato_models.AllClientConnectivityOriginEnum)...),
				),
			},
		},
		"country": schema.SetNestedAttribute{
			Description: "Country traffic matching criteria. Logical ‘OR’ is applied within the " +
				"criteria set. Logical ‘AND’ " +
				"is applied between criteria sets.",
			Required: true,
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Country"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"description": schema.StringAttribute{
			Description: "Description",
			Required:    true,
		},
		"device": schema.SetNestedAttribute{
			Description: "Device Profile traffic matching criteria. Logical ‘OR’ is applied within " +
				"the criteria set. Logical " +
				"‘AND’ is applied between criteria sets.",
			Required: true,
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Device"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"enabled": schema.BoolAttribute{
			Description: "Enabled",
			Required:    true,
		},
		"id": schema.StringAttribute{
			Description:   "Rule ID",
			Computed:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"index": schema.Int64Attribute{
			Description: "Current rule index",
			Computed:    true,
		},
		"name": schema.StringAttribute{
			Description: "Name",
			Required:    true,
			Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"platform": schema.SetAttribute{
			Description: "Source device Operating System traffic matching criteria. Logical ‘OR’ is " +
				"applied within the " +
				"criteria set. Logical ‘AND’ is applied between criteria sets.",
			Required:    true,
			ElementType: types.StringType,
			Validators: []validator.Set{
				setvalidator.ValueStringsAre(
					stringvalidator.OneOf(utils.ToStrings(cato_models.AllOperatingSystem)...),
				),
			},
		},
		"section": schema.SingleNestedAttribute{
			Description: "Policy section containing the rule",
			Computed:    true,
			Attributes:  policyshared.RuleSectionAttributes(),
		},
		"source": schema.SingleNestedAttribute{
			Description: "Source traffic matching criteria. Logical ‘OR’ is applied within the " +
				"criteria set. Logical ‘AND’ is " +
				"applied between criteria sets.",
			Required:   true,
			Attributes: schemaResourceRuleSource(),
		},
		"source_range": schema.SetNestedAttribute{
			Description:  "Public ISP IP Range matching criteria.",
			Required:     true,
			NestedObject: schema.NestedAttributeObject{Attributes: schemaResourceRuleSourceRange()},
		},
	}
}

func schemaResourceRuleSource() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"user": schema.SetNestedAttribute{
			Description: "Individual users defined for the account",
			Required:    true,
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("User"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"users_group": schema.SetNestedAttribute{
			Description: "Group of users",
			Required:    true,
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Users group"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
	}
}

func schemaResourceRuleSourceRange() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"global_ip_range": schema.SingleNestedAttribute{
			Description:   "Globally defined IP range",
			Required:      true,
			Attributes:    idname.SchemaNameID("Global ip range"),
			PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
		},
	}
}
