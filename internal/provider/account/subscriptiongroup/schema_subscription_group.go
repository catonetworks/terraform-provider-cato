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

// Modeled as a List (not a Set): elements accept name-only references whose Computed
// `id` is Unknown at plan time. Terraform correlates List elements by index, so an
// Unknown nested value resolves cleanly after apply (including the first create, when
// there is no prior state); a Set element containing an Unknown cannot be correlated and
// raises "Provider produced inconsistent result after apply".
func schemaGroupMembers(description string) schema.ListNestedAttribute {
	return schema.ListNestedAttribute{Description: description, Optional: true, Computed: true,
		NestedObject: schema.NestedAttributeObject{Attributes: idname.SchemaNameID(description)},
		PlanModifiers: []planmodifier.List{
			planmodifiers.NullListModifier(idname.RefObjectType), planmodifiers.IDNameListPlanModifier(),
		},
	}
}
