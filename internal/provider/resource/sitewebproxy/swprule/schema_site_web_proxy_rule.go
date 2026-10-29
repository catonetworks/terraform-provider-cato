package swprule

import (
	"context"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/planmodifiers"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

func (r *policyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a site web proxy rule. Changes are automatically published. Omitting at on update preserves " +
			"placement.",
		Attributes: schemaResource(),
	}
}

func schemaResource() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"at": schema.SingleNestedAttribute{
			Description: "At",
			Optional:    true,
			Attributes:  schemaResourceAt(),
		},
		"rule": schema.SingleNestedAttribute{
			Description: "Rule",
			Required:    true,
			Attributes:  schemaResourceRule(),
		},
	}
}

func schemaResourceAt() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"position": schema.StringAttribute{
			Description: "Position",
			Required:    true,
			Validators:  []validator.String{stringvalidator.OneOf(utils.ToStrings(cato_models.AllPolicyRulePositionEnum)...)},
		},
		"ref": schema.StringAttribute{
			Description:   "Ref",
			Optional:      true,
			Computed:      true,
			PlanModifiers: []planmodifier.String{planmodifiers.NullStringModifier()},
		},
	}
}

func schemaResourceRule() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"allowed_destination_protocol": schema.StringAttribute{
			Description: "Allowed destination protocol",
			Required:    true,
			Validators:  []validator.String{stringvalidator.OneOf(utils.ToStrings(cato_models.AllSiteWebProxyAllowedDestinationProtocol)...)},
		},
		"associated_site": schema.SetNestedAttribute{
			Description: "Associated site",
			Required:    true,
			NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Associated site"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier()},
			},
		},
		"authentication_config": schema.SingleNestedAttribute{
			Description: "Authentication config",
			Required:    true,
			Attributes:  schemaResourceRuleAuthenticationConfig(),
		},
		"description": schema.StringAttribute{
			Description: "Description",
			Required:    true,
		},
		"enabled": schema.BoolAttribute{
			Description: "Enabled",
			Required:    true,
		},
		"fqdn": schema.StringAttribute{
			Description: "Fqdn",
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
		"port": schema.StringAttribute{
			Description: "Canonical decimal listening port from 1 through 65535.",
			Required:    true,
		},
		"section": schema.SingleNestedAttribute{
			Description: "Policy section containing the rule",
			Computed:    true,
			Attributes:  schemaResourceRuleSection(),
		},
		"should_associate_with_all_sites": schema.BoolAttribute{
			Description: "Should associate with all sites",
			Required:    true,
		},
	}
}

func schemaResourceRuleAuthenticationConfig() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"kerberos_config": schema.SingleNestedAttribute{
			Description: "Kerberos configuration. Omit disabled, empty blocks; omission resets the legacy fields.",
			Optional:    true,
			Attributes:  schemaResourceRuleAuthenticationConfigKerberosConfig(),
		},
		"method": schema.StringAttribute{
			Description: "Method",
			Required:    true,
			Validators:  []validator.String{stringvalidator.OneOf(utils.ToStrings(cato_models.AllSiteWebProxyAuthMethod)...)},
		},
	}
}

func schemaResourceRuleAuthenticationConfigKerberosConfig() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"encrypted_keytab": schema.StringAttribute{
			Description:   "Encrypted keytab",
			Sensitive:     true,
			Optional:      true,
			Computed:      true,
			PlanModifiers: []planmodifier.String{planmodifiers.NullStringModifier()},
		},
		"is_enabled": schema.BoolAttribute{
			Description: "Is enabled",
			Required:    true,
		},
	}
}

func schemaResourceRuleSection() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description:   "Section ID",
			Computed:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": schema.StringAttribute{
			Description: "Section name",
			Computed:    true,
		},
	}
}
