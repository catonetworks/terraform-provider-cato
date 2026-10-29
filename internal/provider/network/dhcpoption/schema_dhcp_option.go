package dhcpoption

import (
	"context"

	models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/planmodifiers"
	nc "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/networkconfig"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/utils"
)

func (r *optionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Description: "Manages one standalone DHCP option via networkConfig.dhcp. " +
		"Unrelated collection items remain unchanged. Requires NCS_DHCP_PUBLIC_API.", Attributes: map[string]schema.Attribute{
		"id":    nc.IDAttribute(),
		"tag":   schema.Int64Attribute{Required: true, Description: "DHCP option code."},
		"value": schema.StringAttribute{Required: true, Description: "Option value formatted according to type."},
		"type": schema.StringAttribute{Required: true,
			Description: "DHCP option encoding.",
			Validators:  []validator.String{stringvalidator.OneOf(utils.ToStrings(models.AllNetworkConfigDhcpOptionType)...)}},
		"description": schema.StringAttribute{
			Optional:      true,
			Computed:      true,
			PlanModifiers: []planmodifier.String{planmodifiers.NullStringModifier()},
			Description:   "Optional description; omit, null, or empty string clears it.",
		},
	}}
}
