package dhcp

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	nc "github.com/catonetworks/terraform-provider-cato/internal/provider/network/networkconfig"
)

type RelayGroupModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Server types.List   `tfsdk:"server"`
}

func (r *relayGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Description: "Manages one standalone DHCP relay group via networkConfig.dhcp. " +
		"Unrelated collection items remain unchanged. Requires NCS_DHCP_PUBLIC_API.", Attributes: map[string]schema.Attribute{
		"id": nc.IDAttribute(),
		"name": schema.StringAttribute{Required: true,
			Description: "Relay group name, unique within account.",
			Validators:  []validator.String{stringvalidator.LengthAtLeast(1)}},
		"server": schema.ListAttribute{Required: true,
			ElementType: types.StringType,
			Description: "DHCP relay server addresses.",
			Validators: []validator.List{listvalidator.SizeAtLeast(1),
				listvalidator.ValueStringsAre(nc.IPValidator{})}},
	}}
}
