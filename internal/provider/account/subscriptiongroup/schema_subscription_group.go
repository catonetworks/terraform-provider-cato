package subscriptiongroup

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/planmodifiers"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

func (r *subscriptionGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages one subscription group and its complete mailing-list and integration membership. " +
			"Import using its ID. This resource does not configure event producers or " +
			"send test notifications.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Description: "Resource ID", Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name": schema.StringAttribute{Description: "Name", Required: true,
				Validators: []validator.String{stringvalidator.LengthAtLeast(1)}},
			"integration":  schemaGroupMembers("Integration references"),
			"mailing_list": schemaGroupMembers("Mailing list references"),
		},
	}
}
func schemaGroupMembers(description string) schema.SetNestedAttribute {
	return schema.SetNestedAttribute{Description: description, Optional: true, Computed: true,
		NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID(description)},
		PlanModifiers: []planmodifier.Set{
			planmodifiers.NullSetModifier(idname.RefObjectType), planmodifiers.IDNameSetPlanModifier(),
		},
	}
}
