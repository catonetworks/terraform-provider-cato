package sbrule

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
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
		Description: "Manages a socket bypass rule. Changes are automatically published. " +
			"Omitting at on update preserves " +
			"placement.",
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
		"action": schema.SingleNestedAttribute{
			Description: "Action",
			Required:    true,
			Attributes:  schemaResourceRuleAction(),
		},
		"description": schema.StringAttribute{
			Description: "Description",
			Required:    true,
		},
		"destination": schema.SingleNestedAttribute{
			Description: "Destination",
			Required:    true,
			Attributes:  schemaResourceRuleDestination(),
		},
		"enabled": schema.BoolAttribute{
			Description: "Enabled",
			Required:    true,
		},
		"exception": schema.SetNestedAttribute{
			Description:  "Exception",
			Required:     true,
			NestedObject: schema.NestedAttributeObject{Attributes: schemaResourceRuleException()},
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
		"section": schema.SingleNestedAttribute{
			Description: "Policy section containing the rule",
			Computed:    true,
			Attributes:  policyshared.RuleSectionAttributes(),
		},
		"service": schema.SingleNestedAttribute{
			Description: "Service",
			Required:    true,
			Attributes:  schemaResourceRuleService(),
		},
		"site": schema.SingleNestedAttribute{
			Description: "Site",
			Required:    true,
			Attributes:  schemaResourceRuleSite(),
		},
		"source": schema.SingleNestedAttribute{
			Description: "Source",
			Required:    true,
			Attributes:  schemaResourceRuleSource(),
		},
	}
}

func schemaResourceRuleAction() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"event": schema.SingleNestedAttribute{
			Description: "Event",
			Required:    true,
			Attributes:  schemaResourceRuleActionEvent(),
		},
		"port": schema.StringAttribute{
			Description: "Port",
			Required:    true,
			Validators:  []validator.String{stringvalidator.OneOf(utils.ToStrings(cato_models.AllBypassSocketPortType)...)},
		},
	}
}

func schemaResourceRuleActionEvent() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"enabled": schema.BoolAttribute{
			Description: "Enabled",
			Required:    true,
		},
	}
}

func schemaResourceRuleDestination() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"application": schema.SetNestedAttribute{
			Description: "Application",
			Required:    true,
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Application"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"custom_app": schema.SetNestedAttribute{
			Description: "Custom app",
			Required:    true,
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Custom app"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"domain": schema.SetAttribute{
			Description: "Domain",
			Required:    true,
			ElementType: types.StringType,
		},
		"fqdn": schema.SetAttribute{
			Description: "Fqdn",
			Required:    true,
			ElementType: types.StringType,
		},
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
			NestedObject: schema.NestedAttributeObject{Attributes: schemaResourceRuleDestinationIPRange()},
		},
		"subnet": schema.SetAttribute{
			Description: "Subnet",
			Required:    true,
			ElementType: types.StringType,
		},
	}
}

func schemaResourceRuleDestinationIPRange() map[string]schema.Attribute {
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

func schemaResourceRuleException() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"destination": schema.SingleNestedAttribute{
			Description: "Destination",
			Required:    true,
			Attributes:  schemaResourceRuleExceptionDestination(),
		},
		"name": schema.StringAttribute{
			Description: "Name",
			Required:    true,
			Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"service": schema.SingleNestedAttribute{
			Description: "Service",
			Required:    true,
			Attributes:  schemaResourceRuleExceptionService(),
		},
		"site": schema.SingleNestedAttribute{
			Description: "Site",
			Required:    true,
			Attributes:  schemaResourceRuleExceptionSite(),
		},
		"source": schema.SingleNestedAttribute{
			Description: "Source",
			Required:    true,
			Attributes:  schemaResourceRuleExceptionSource(),
		},
	}
}

func schemaResourceRuleExceptionDestination() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"application": schema.SetNestedAttribute{
			Description: "Application",
			Required:    true,
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Application"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"custom_app": schema.SetNestedAttribute{
			Description: "Custom app",
			Required:    true,
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Custom app"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"domain": schema.SetAttribute{
			Description: "Domain",
			Required:    true,
			ElementType: types.StringType,
		},
		"fqdn": schema.SetAttribute{
			Description: "Fqdn",
			Required:    true,
			ElementType: types.StringType,
		},
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
			NestedObject: schema.NestedAttributeObject{Attributes: schemaResourceRuleExceptionDestinationIPRange()},
		},
		"subnet": schema.SetAttribute{
			Description: "Subnet",
			Required:    true,
			ElementType: types.StringType,
		},
	}
}

func schemaResourceRuleExceptionDestinationIPRange() map[string]schema.Attribute {
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

func schemaResourceRuleExceptionService() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"custom": schema.SetNestedAttribute{
			Description:  "Custom",
			Required:     true,
			NestedObject: schema.NestedAttributeObject{Attributes: schemaResourceRuleExceptionServiceCustom()},
		},
		"simple": schema.SetNestedAttribute{
			Description:  "Simple",
			Required:     true,
			NestedObject: schema.NestedAttributeObject{Attributes: schemaResourceRuleExceptionServiceSimple()},
		},
	}
}

func schemaResourceRuleExceptionServiceCustom() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"port": schema.SetAttribute{
			Description:   "Port",
			Optional:      true,
			Computed:      true,
			ElementType:   types.StringType,
			PlanModifiers: []planmodifier.Set{planmodifiers.NullSetModifier(types.StringType)},
		},
		"port_range": schema.SingleNestedAttribute{
			Description: "Port range",
			Optional:    true,
			Attributes:  schemaResourceRuleExceptionServiceCustomPortRange(),
		},
		"protocol": schema.StringAttribute{
			Description: "Protocol",
			Required:    true,
			Validators:  []validator.String{stringvalidator.OneOf(utils.ToStrings(cato_models.AllIPProtocol)...)},
		},
	}
}

func schemaResourceRuleExceptionServiceCustomPortRange() map[string]schema.Attribute {
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

func schemaResourceRuleExceptionServiceSimple() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Description: "Name",
			Required:    true,
			Validators:  []validator.String{stringvalidator.OneOf(utils.ToStrings(cato_models.AllSimpleServiceType)...)},
		},
	}
}

func schemaResourceRuleExceptionSite() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"group": schema.SetNestedAttribute{
			Description: "Group",
			Required:    true,
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Group"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"site": schema.SetNestedAttribute{
			Description: "Site",
			Required:    true,
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Site"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
	}
}

func schemaResourceRuleExceptionSource() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"floating_subnet": schema.SetNestedAttribute{
			Description: "Floating subnet",
			Required:    true,
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Floating subnet"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"global_ip_range": schema.SetNestedAttribute{
			Description: "Global ip range",
			Required:    true,
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Global ip range"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"group": schema.SetNestedAttribute{
			Description: "Group",
			Required:    true,
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Group"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"host": schema.SetNestedAttribute{
			Description: "Host",
			Required:    true,
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Host"),
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
			NestedObject: schema.NestedAttributeObject{Attributes: schemaResourceRuleExceptionSourceIPRange()},
		},
		"network_interface": schema.SetNestedAttribute{
			Description: "Network interface",
			Required:    true,
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Network interface"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"site_network_subnet": schema.SetNestedAttribute{
			Description: "Site network subnet",
			Required:    true,
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Site network subnet"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"subnet": schema.SetAttribute{
			Description: "Subnet",
			Required:    true,
			ElementType: types.StringType,
		},
		"vlan": schema.SetAttribute{
			Description: "Vlan",
			Required:    true,
			ElementType: types.Int64Type,
			Validators:  []validator.Set{setvalidator.ValueInt64sAre(int64validator.Between(1, 4094))},
		},
	}
}

func schemaResourceRuleExceptionSourceIPRange() map[string]schema.Attribute {
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

func schemaResourceRuleService() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"custom": schema.SetNestedAttribute{
			Description:  "Custom",
			Required:     true,
			NestedObject: schema.NestedAttributeObject{Attributes: schemaResourceRuleServiceCustom()},
		},
		"simple": schema.SetNestedAttribute{
			Description:  "Simple",
			Required:     true,
			NestedObject: schema.NestedAttributeObject{Attributes: schemaResourceRuleServiceSimple()},
		},
	}
}

func schemaResourceRuleServiceCustom() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"port": schema.SetAttribute{
			Description:   "Port",
			Optional:      true,
			Computed:      true,
			ElementType:   types.StringType,
			PlanModifiers: []planmodifier.Set{planmodifiers.NullSetModifier(types.StringType)},
		},
		"port_range": schema.SingleNestedAttribute{
			Description: "Port range",
			Optional:    true,
			Attributes:  schemaResourceRuleServiceCustomPortRange(),
		},
		"protocol": schema.StringAttribute{
			Description: "Protocol",
			Required:    true,
			Validators:  []validator.String{stringvalidator.OneOf(utils.ToStrings(cato_models.AllIPProtocol)...)},
		},
	}
}

func schemaResourceRuleServiceCustomPortRange() map[string]schema.Attribute {
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

func schemaResourceRuleServiceSimple() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Description: "Name",
			Required:    true,
			Validators:  []validator.String{stringvalidator.OneOf(utils.ToStrings(cato_models.AllSimpleServiceType)...)},
		},
	}
}

func schemaResourceRuleSite() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"group": schema.SetNestedAttribute{
			Description: "Group",
			Required:    true,
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Group"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"site": schema.SetNestedAttribute{
			Description: "Site",
			Required:    true,
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Site"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
	}
}

func schemaResourceRuleSource() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"floating_subnet": schema.SetNestedAttribute{
			Description: "Floating subnet",
			Required:    true,
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Floating subnet"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"global_ip_range": schema.SetNestedAttribute{
			Description: "Global ip range",
			Required:    true,
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Global ip range"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"group": schema.SetNestedAttribute{
			Description: "Group",
			Required:    true,
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Group"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"host": schema.SetNestedAttribute{
			Description: "Host",
			Required:    true,
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Host"),
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
			NestedObject: schema.NestedAttributeObject{Attributes: schemaResourceRuleSourceIPRange()},
		},
		"network_interface": schema.SetNestedAttribute{
			Description: "Network interface",
			Required:    true,
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Network interface"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"site_network_subnet": schema.SetNestedAttribute{
			Description: "Site network subnet",
			Required:    true,
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Site network subnet"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"subnet": schema.SetAttribute{
			Description: "Subnet",
			Required:    true,
			ElementType: types.StringType,
		},
		"vlan": schema.SetAttribute{
			Description: "Vlan",
			Required:    true,
			ElementType: types.Int64Type,
			Validators:  []validator.Set{setvalidator.ValueInt64sAre(int64validator.Between(1, 4094))},
		},
	}
}

func schemaResourceRuleSourceIPRange() map[string]schema.Attribute {
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
