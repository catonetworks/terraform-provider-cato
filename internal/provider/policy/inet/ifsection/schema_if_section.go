package ifsection

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func (r *internetFwSectionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `cato_if_section` resource contains the configuration parameters necessary to Internet firewall section " +
			"(https://support.catonetworks.com/hc/en-us/articles/5590037900701-Adding-Sections-to-the-WAN-and-Internet-Firewalls). " +
			"Documentation for the underlying API used in this resource can be found at mutation.policy.internetFirewall.addSection().",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Section ID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"at": schema.SingleNestedAttribute{
				Description: "",
				Required:    true,
				Optional:    false,
				Attributes: map[string]schema.Attribute{
					"position": schema.StringAttribute{
						Description: "Position relative to a policy, a section or another rule",
						Required:    true,
						Optional:    false,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(), // Avoid drift
						},
						Validators: []validator.String{
							stringvalidator.OneOf("AFTER_SECTION", "BEFORE_SECTION", "LAST_IN_POLICY"),
						},
					},
					"ref": schema.StringAttribute{
						Description: "The identifier of the object (e.g. a rule, a section) relative to which the position of the added rule is defined",
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
