package mailinglist

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/planmodifiers"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

func (r *mailingListResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages one mailing list and its complete address and administrator membership. " +
			"Creating a mailing list does not send notifications. Import using its ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Resource ID", Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Description: "Name", Required: true, Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"address": schema.SetAttribute{
				Description: "Email addresses. Removing this attribute clears all addresses.",
				Optional:    true, Computed: true, ElementType: types.StringType,
				PlanModifiers: []planmodifier.Set{planmodifiers.NullSetModifier(types.StringType)},
			},
			"admin": schema.SetNestedAttribute{
				Description: "Administrator IDs with computed display names. Name-only references are unsupported. " +
					"Removing this attribute clears administrators; at least one address or administrator is required.",
				Optional: true, Computed: true,
				NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID("Administrator")},
				PlanModifiers: []planmodifier.Set{
					planmodifiers.NullSetModifier(idname.RefObjectType), planmodifiers.IDNameSetPlanModifier(),
				},
			},
		},
	}
}
