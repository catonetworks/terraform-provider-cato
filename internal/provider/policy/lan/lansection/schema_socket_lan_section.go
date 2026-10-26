package lansection

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func (r *socketLanSectionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_socket_lan_section` resource contains the configuration parameters necessary to add a Socket LAN " +
			"firewall section. Documentation for the underlying API used in this resource can be found at " +
			"[mutation.policy.socketLan.addSection()](https://api.catonetworks.com/documentation/#mutation-policy.socketLan.addSection).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Section ID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"at": schema.SingleNestedAttribute{
				Description: "Position of the section in the policy",
				Required:    true,
				Optional:    false,
				Attributes: map[string]schema.Attribute{
					"position": schema.StringAttribute{
						Description: "Position relative to a policy or another section (AFTER_SECTION, BEFORE_SECTION, LAST_IN_POLICY)",
						Required:    true,
						Optional:    false,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.String{
							stringvalidator.OneOf("AFTER_SECTION", "BEFORE_SECTION", "LAST_IN_POLICY"),
						},
					},
					"ref": schema.StringAttribute{
						Description: "The identifier of the section relative to which the position of the added section is defined",
						Required:    false,
						Optional:    true,
					},
				},
			},
			"section": schema.SingleNestedAttribute{
				Description: "Section parameters",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Description: "Section ID",
						Computed:    true,
						Optional:    false,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"name": schema.StringAttribute{
						Description: "Section Name",
						Required:    true,
						Optional:    false,
					},
				},
			},
		},
	}
}
