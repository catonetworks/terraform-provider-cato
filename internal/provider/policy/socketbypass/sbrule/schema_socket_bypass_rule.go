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

// Optionality mirrors the canonical Internet Firewall rule schema
// (schema_if_rule.go): every traffic match / scope CRITERION (source,
// destination, service, exception, site, and their nested set fields) is
// Optional, because a socket bypass rule matches on one of them, not all --
// the export emits only the fields a given rule actually uses, so marking them
// Required produced "Inappropriate value for attribute rule: attributes ... are
// required" type-conformance errors in the bulk module. Only structural/control
// fields stay Required: action (+ event/port), enabled, name, and the leaf
// fields that are mandatory once their parent object is instantiated (ip_range
// from/to, custom-service protocol, port_range from/to, simple-service name).
// The converter already tolerates nulls on every criterion (buildCreateObject*
// each short-circuit on value.IsNull()), so this is a schema-only change.

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
			Optional:    true,
		},
		"destination": schema.SingleNestedAttribute{
			Description: "Destination",
			Optional:    true,
			Attributes:  schemaResourceRuleDestination(),
		},
		"enabled": schema.BoolAttribute{
			Description: "Enabled",
			Required:    true,
		},
		"exception": schema.SetNestedAttribute{
			Description: "Exception",
			Optional:    true,
			// Optional-only (NOT Computed): the converter's validateValues rejects
			// unknown values (validators_socket_bypass_rule.go), so a Computed
			// exception left unset in config would surface as "must contain a known
			// value". IFW marks exceptions Computed for drift-suppression, but that
			// relies on its own hydration guarantees; keep SB at Optional to avoid a
			// known/unknown apply error.
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
			Optional:    true,
			Attributes:  schemaResourceRuleService(),
		},
		"site": schema.SingleNestedAttribute{
			Description: "Site",
			Optional:    true,
			Attributes:  schemaResourceRuleSite(),
		},
		"source": schema.SingleNestedAttribute{
			Description: "Source",
			Optional:    true,
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
			Optional:    true,
			// match name/id refs against state as a whole set, so a name-only
			// config does not show the computed id as a change on every plan
			PlanModifiers: []planmodifier.Set{planmodifiers.IDNameSetPlanModifier()},
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Application"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"custom_app": schema.SetNestedAttribute{
			Description: "Custom app",
			Optional:    true,
			// match name/id refs against state as a whole set, so a name-only
			// config does not show the computed id as a change on every plan
			PlanModifiers: []planmodifier.Set{planmodifiers.IDNameSetPlanModifier()},
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Custom app"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"domain": schema.SetAttribute{
			Description: "Domain",
			Optional:    true,
			ElementType: types.StringType,
		},
		"fqdn": schema.SetAttribute{
			Description: "Fqdn",
			Optional:    true,
			ElementType: types.StringType,
		},
		"global_ip_range": schema.SetNestedAttribute{
			Description: "Global ip range",
			Optional:    true,
			// match name/id refs against state as a whole set, so a name-only
			// config does not show the computed id as a change on every plan
			PlanModifiers: []planmodifier.Set{planmodifiers.IDNameSetPlanModifier()},
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Global ip range"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"ip": schema.SetAttribute{
			Description: "Ip",
			Optional:    true,
			ElementType: types.StringType,
		},
		"ip_range": schema.SetNestedAttribute{
			Description:  "Ip range",
			Optional:     true,
			NestedObject: schema.NestedAttributeObject{Attributes: schemaResourceRuleDestinationIPRange()},
		},
		"subnet": schema.SetAttribute{
			Description: "Subnet",
			Optional:    true,
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
			Optional:    true,
			Attributes:  schemaResourceRuleExceptionDestination(),
		},
		"name": schema.StringAttribute{
			Description: "Name",
			Required:    true,
			Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"service": schema.SingleNestedAttribute{
			Description: "Service",
			Optional:    true,
			Attributes:  schemaResourceRuleExceptionService(),
		},
		"site": schema.SingleNestedAttribute{
			Description: "Site",
			Optional:    true,
			Attributes:  schemaResourceRuleExceptionSite(),
		},
		"source": schema.SingleNestedAttribute{
			Description: "Source",
			Optional:    true,
			Attributes:  schemaResourceRuleExceptionSource(),
		},
	}
}

func schemaResourceRuleExceptionDestination() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"application": schema.SetNestedAttribute{
			Description: "Application",
			Optional:    true,
			// match name/id refs against state as a whole set, so a name-only
			// config does not show the computed id as a change on every plan
			PlanModifiers: []planmodifier.Set{planmodifiers.IDNameSetPlanModifier()},
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Application"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"custom_app": schema.SetNestedAttribute{
			Description: "Custom app",
			Optional:    true,
			// match name/id refs against state as a whole set, so a name-only
			// config does not show the computed id as a change on every plan
			PlanModifiers: []planmodifier.Set{planmodifiers.IDNameSetPlanModifier()},
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Custom app"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"domain": schema.SetAttribute{
			Description: "Domain",
			Optional:    true,
			ElementType: types.StringType,
		},
		"fqdn": schema.SetAttribute{
			Description: "Fqdn",
			Optional:    true,
			ElementType: types.StringType,
		},
		"global_ip_range": schema.SetNestedAttribute{
			Description: "Global ip range",
			Optional:    true,
			// match name/id refs against state as a whole set, so a name-only
			// config does not show the computed id as a change on every plan
			PlanModifiers: []planmodifier.Set{planmodifiers.IDNameSetPlanModifier()},
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Global ip range"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"ip": schema.SetAttribute{
			Description: "Ip",
			Optional:    true,
			ElementType: types.StringType,
		},
		"ip_range": schema.SetNestedAttribute{
			Description:  "Ip range",
			Optional:     true,
			NestedObject: schema.NestedAttributeObject{Attributes: schemaResourceRuleExceptionDestinationIPRange()},
		},
		"subnet": schema.SetAttribute{
			Description: "Subnet",
			Optional:    true,
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
			Optional:     true,
			NestedObject: schema.NestedAttributeObject{Attributes: schemaResourceRuleExceptionServiceCustom()},
		},
		"simple": schema.SetNestedAttribute{
			Description:  "Simple",
			Optional:     true,
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
			Optional:    true,
			// match name/id refs against state as a whole set, so a name-only
			// config does not show the computed id as a change on every plan
			PlanModifiers: []planmodifier.Set{planmodifiers.IDNameSetPlanModifier()},
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Group"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"site": schema.SetNestedAttribute{
			Description: "Site",
			Optional:    true,
			// match name/id refs against state as a whole set, so a name-only
			// config does not show the computed id as a change on every plan
			PlanModifiers: []planmodifier.Set{planmodifiers.IDNameSetPlanModifier()},
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
			Optional:    true,
			// match name/id refs against state as a whole set, so a name-only
			// config does not show the computed id as a change on every plan
			PlanModifiers: []planmodifier.Set{planmodifiers.IDNameSetPlanModifier()},
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Floating subnet"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"global_ip_range": schema.SetNestedAttribute{
			Description: "Global ip range",
			Optional:    true,
			// match name/id refs against state as a whole set, so a name-only
			// config does not show the computed id as a change on every plan
			PlanModifiers: []planmodifier.Set{planmodifiers.IDNameSetPlanModifier()},
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Global ip range"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"group": schema.SetNestedAttribute{
			Description: "Group",
			Optional:    true,
			// match name/id refs against state as a whole set, so a name-only
			// config does not show the computed id as a change on every plan
			PlanModifiers: []planmodifier.Set{planmodifiers.IDNameSetPlanModifier()},
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Group"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"host": schema.SetNestedAttribute{
			Description: "Host",
			Optional:    true,
			// match name/id refs against state as a whole set, so a name-only
			// config does not show the computed id as a change on every plan
			PlanModifiers: []planmodifier.Set{planmodifiers.IDNameSetPlanModifier()},
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Host"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"ip": schema.SetAttribute{
			Description: "Ip",
			Optional:    true,
			ElementType: types.StringType,
		},
		"ip_range": schema.SetNestedAttribute{
			Description:  "Ip range",
			Optional:     true,
			NestedObject: schema.NestedAttributeObject{Attributes: schemaResourceRuleExceptionSourceIPRange()},
		},
		"network_interface": schema.SetNestedAttribute{
			Description: "Network interface",
			Optional:    true,
			// match name/id refs against state as a whole set, so a name-only
			// config does not show the computed id as a change on every plan
			PlanModifiers: []planmodifier.Set{planmodifiers.IDNameSetPlanModifier()},
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Network interface"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"site_network_subnet": schema.SetNestedAttribute{
			Description: "Site network subnet",
			Optional:    true,
			// match name/id refs against state as a whole set, so a name-only
			// config does not show the computed id as a change on every plan
			PlanModifiers: []planmodifier.Set{planmodifiers.IDNameSetPlanModifier()},
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Site network subnet"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"subnet": schema.SetAttribute{
			Description: "Subnet",
			Optional:    true,
			ElementType: types.StringType,
		},
		"vlan": schema.SetAttribute{
			Description: "Vlan",
			Optional:    true,
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
			Optional:     true,
			NestedObject: schema.NestedAttributeObject{Attributes: schemaResourceRuleServiceCustom()},
		},
		"simple": schema.SetNestedAttribute{
			Description:  "Simple",
			Optional:     true,
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
			Optional:    true,
			// match name/id refs against state as a whole set, so a name-only
			// config does not show the computed id as a change on every plan
			PlanModifiers: []planmodifier.Set{planmodifiers.IDNameSetPlanModifier()},
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Group"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"site": schema.SetNestedAttribute{
			Description: "Site",
			Optional:    true,
			// match name/id refs against state as a whole set, so a name-only
			// config does not show the computed id as a change on every plan
			PlanModifiers: []planmodifier.Set{planmodifiers.IDNameSetPlanModifier()},
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
			Optional:    true,
			// match name/id refs against state as a whole set, so a name-only
			// config does not show the computed id as a change on every plan
			PlanModifiers: []planmodifier.Set{planmodifiers.IDNameSetPlanModifier()},
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Floating subnet"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"global_ip_range": schema.SetNestedAttribute{
			Description: "Global ip range",
			Optional:    true,
			// match name/id refs against state as a whole set, so a name-only
			// config does not show the computed id as a change on every plan
			PlanModifiers: []planmodifier.Set{planmodifiers.IDNameSetPlanModifier()},
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Global ip range"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"group": schema.SetNestedAttribute{
			Description: "Group",
			Optional:    true,
			// match name/id refs against state as a whole set, so a name-only
			// config does not show the computed id as a change on every plan
			PlanModifiers: []planmodifier.Set{planmodifiers.IDNameSetPlanModifier()},
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Group"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"host": schema.SetNestedAttribute{
			Description: "Host",
			Optional:    true,
			// match name/id refs against state as a whole set, so a name-only
			// config does not show the computed id as a change on every plan
			PlanModifiers: []planmodifier.Set{planmodifiers.IDNameSetPlanModifier()},
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Host"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"ip": schema.SetAttribute{
			Description: "Ip",
			Optional:    true,
			ElementType: types.StringType,
		},
		"ip_range": schema.SetNestedAttribute{
			Description:  "Ip range",
			Optional:     true,
			NestedObject: schema.NestedAttributeObject{Attributes: schemaResourceRuleSourceIPRange()},
		},
		"network_interface": schema.SetNestedAttribute{
			Description: "Network interface",
			Optional:    true,
			// match name/id refs against state as a whole set, so a name-only
			// config does not show the computed id as a change on every plan
			PlanModifiers: []planmodifier.Set{planmodifiers.IDNameSetPlanModifier()},
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Network interface"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"site_network_subnet": schema.SetNestedAttribute{
			Description: "Site network subnet",
			Optional:    true,
			// match name/id refs against state as a whole set, so a name-only
			// config does not show the computed id as a change on every plan
			PlanModifiers: []planmodifier.Set{planmodifiers.IDNameSetPlanModifier()},
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Site network subnet"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"subnet": schema.SetAttribute{
			Description: "Subnet",
			Optional:    true,
			ElementType: types.StringType,
		},
		"vlan": schema.SetAttribute{
			Description: "Vlan",
			Optional:    true,
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
