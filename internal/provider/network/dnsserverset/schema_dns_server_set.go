package dnsserverset

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *dnsSetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Description: "Manages one DNS server set. Resolver order is preserved. " +
		"Create, update and delete target only this set. Import using its ID.", Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{Description: "Resource ID", Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"name": schema.StringAttribute{Description: "Name", Required: true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)}},
		"server": schema.ListAttribute{Description: "Ordered IPv4 resolver addresses, up to two. Use [] to clear.",
			Required: true, ElementType: types.StringType},
	}}
}
