package rpfrule

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
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
		Description: "Manages a remote port forward rule. Changes are automatically published. " +
			"Omitting at on update " +
			"preserves placement.",
		Attributes: schemaResource(),
	}
}

func schemaResource() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"at": schema.SingleNestedAttribute{
			Description: "At",
			Optional:    true,
			Attributes:  policyshared.RulePositionAttributes("Position", "Ref"),
		},
		"rule": schema.SingleNestedAttribute{
			Description: "Rule",
			Required:    true,
			Attributes:  schemaResourceRule(),
		},
	}
}

func schemaResourceRule() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"description": schema.StringAttribute{
			Description: "Description",
			Required:    true,
		},
		"enabled": schema.BoolAttribute{
			Description: "Enabled",
			Required:    true,
		},
		"external_ip": schema.SingleNestedAttribute{
			Description:   "External ip",
			Required:      true,
			Attributes:    idname.SchemaNameID("External ip"),
			PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
		},
		"external_port_range": schema.SingleNestedAttribute{
			Description: "External port range",
			Required:    true,
			Attributes:  schemaResourceRuleExternalPortRange(),
		},
		"forward_icmp": schema.BoolAttribute{
			Description: "Forward icmp",
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
		"internal_ip": schema.StringAttribute{
			Description: "Internal ip",
			Required:    true,
		},
		"internal_port_range": schema.SingleNestedAttribute{
			Description: "Internal port range",
			Required:    true,
			Attributes:  schemaResourceRuleInternalPortRange(),
		},
		"name": schema.StringAttribute{
			Description: "Name",
			Required:    true,
			Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"remote_ips": schema.SingleNestedAttribute{
			Description: "Remote i ps",
			Required:    true,
			Attributes:  schemaResourceRuleRemoteIps(),
		},
		"restriction_type": schema.StringAttribute{
			Description: "Restriction type",
			Required:    true,
			Validators:  []validator.String{stringvalidator.OneOf(utils.ToStrings(cato_models.AllRemotePortFwdRestrictionType)...)},
		},
		"section": schema.SingleNestedAttribute{
			Description: "Policy section containing the rule",
			Computed:    true,
			Attributes:  policyshared.RuleSectionAttributes(),
		},
		"tracking": schema.SingleNestedAttribute{
			Description: "Tracking",
			Required:    true,
			Attributes:  schemaResourceRuleTracking(),
		},
	}
}

func schemaResourceRuleExternalPortRange() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"from": schema.StringAttribute{
			Description: "From",
			Required:    true,
		},
		"to": schema.StringAttribute{
			Description: "To",
			Required:    true,
		},
	}
}

func schemaResourceRuleInternalPortRange() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"from": schema.StringAttribute{
			Description: "From",
			Required:    true,
		},
		"to": schema.StringAttribute{
			Description: "To",
			Required:    true,
		},
	}
}

func schemaResourceRuleRemoteIps() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"global_ip_range": schema.SetNestedAttribute{
			Description: "Global ip range",
			Required:    true,
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Global ip range"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"ip": schema.SetAttribute{
			Description: "Ip",
			Required:    true,
			ElementType: types.StringType,
		},
		"ip_range": schema.SetNestedAttribute{
			Description:  "Ip range",
			Required:     true,
			NestedObject: schema.NestedAttributeObject{Attributes: schemaResourceRuleRemoteIpsIPRange()},
		},
		"subnet": schema.SetAttribute{
			Description: "Subnet",
			Required:    true,
			ElementType: types.StringType,
		},
	}
}

func schemaResourceRuleRemoteIpsIPRange() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"from": schema.StringAttribute{
			Description: "From",
			Required:    true,
		},
		"to": schema.StringAttribute{
			Description: "To",
			Required:    true,
		},
	}
}

func schemaResourceRuleTracking() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"enabled": schema.BoolAttribute{
			Description: "Enabled",
			Required:    true,
		},
		"frequency": schema.StringAttribute{
			Description: "Frequency",
			Required:    true,
			Validators:  []validator.String{stringvalidator.OneOf(utils.ToStrings(cato_models.AllPolicyRuleTrackingFrequencyEnum)...)},
		},
		"mailing_list": schema.SetNestedAttribute{
			Description: "Mailing list",
			Required:    true,
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Mailing list"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"subscription_group": schema.SetNestedAttribute{
			Description: "Subscription group",
			Required:    true,
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Subscription group"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"webhook": schema.SetNestedAttribute{
			Description: "Webhook",
			Required:    true,
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Webhook"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
	}
}
