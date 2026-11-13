package allocatedip

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func (r *allocatedIPResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Allocates an IP address from a Cato PoP location. Cato assigns the IP -- " +
			"the configuration chooses the PoP location, description, and type; the resulting " +
			"ip and display name (\"Portland - 192.0.2.1\") are computed and tenant-specific.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Allocated IP identifier",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"pop_location_name": schema.StringAttribute{
				Description: "Name of the Cato PoP location (city) to allocate the IP in, e.g. \"Portland\"",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the allocated IP (updatable in place). With pop_location_name, " +
					"the cross-tenant identity of the allocation -- keep it unique per PoP location.",
				Optional: true,
			},
			"type": schema.StringAttribute{
				Description: "Allocation type (SYSTEM is the only supported value today)",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("SYSTEM"),
				Validators: []validator.String{
					stringvalidator.OneOf("SYSTEM"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ip": schema.StringAttribute{
				Description: "The IP address Cato assigned to this allocation",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Display name of the allocation (\"Portland - 192.0.2.1\") as shown in " +
					"entity lookups and referenced by policy rules",
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}
