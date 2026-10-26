package dnsforwardingrule

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	nc "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/networkconfig"
)

func (r *forwardingRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Description: "Manages one standalone DNS forwarding rule via " +
		"networkConfig.dns. Unrelated collection items remain unchanged. Requires " +
		"NCS_DNS_PUBLIC_API.", Attributes: map[string]schema.Attribute{
		"id": nc.IDAttribute(),
		"domain": schema.StringAttribute{Required: true,
			Description: "Unique DNS forwarding domain.",
			Validators:  []validator.String{stringvalidator.LengthAtLeast(1)}},
		"server": schema.ListAttribute{Required: true,
			ElementType: types.StringType,
			Description: "Ordered forwarding server addresses (1–6).",
			Validators: []validator.List{listvalidator.SizeBetween(1,
				6),
				listvalidator.ValueStringsAre(nc.IPValidator{})}},
	}}
}
