package dhcp

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	nc "github.com/catonetworks/terraform-provider-cato/internal/provider/network/networkconfig"
)

type OptionModel struct {
	ID          types.String `tfsdk:"id"`
	Tag         types.Int64  `tfsdk:"tag"`
	Value       types.String `tfsdk:"value"`
	Type        types.String `tfsdk:"type"`
	Description types.String `tfsdk:"description"`
}

func (r *optionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Description: "Manages one standalone DHCP option via networkConfig.dhcp. " +
		"Unrelated collection items remain unchanged. Requires NCS_DHCP_PUBLIC_API.", Attributes: map[string]schema.Attribute{
		"id":    nc.IDAttribute(),
		"tag":   schema.Int64Attribute{Required: true, Description: "DHCP option code."},
		"value": schema.StringAttribute{Required: true, Description: "Option value formatted according to type."},
		"type": schema.StringAttribute{Required: true,
			Description: "DHCP option encoding.",
			Validators: []validator.String{stringvalidator.OneOf("ASCII",
				"HEX",
				"IP")}},
		"description": schema.StringAttribute{
			Optional:    true,
			Description: "Optional nonempty description; omit or set null to clear it.",
			Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
		},
	}}
}
