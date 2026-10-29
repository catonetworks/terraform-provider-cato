package allocatedip

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/planmodifiers"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

func (r *allocatedIPResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Description: "Manages one account-scoped SYSTEM public IP allocation at an explicit PoP. " +
		"BYOIP creation/release require private inputs absent from the public SDK and are unsupported. Import using the allocation ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Description: "Resource ID", Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"ip":     schema.StringAttribute{Description: "Allocated public IP address", Computed: true},
			"state":  schema.StringAttribute{Description: "Allocation state returned by the API", Computed: true},
			"subnet": schema.StringAttribute{Description: "Allocated subnet returned by the API", Computed: true},
			"description": schema.StringAttribute{Description: "Description. Removing this value clears it.", Optional: true, Computed: true,
				PlanModifiers: []planmodifier.String{planmodifiers.NullStringModifier()}},
			"type": schema.StringAttribute{Description: "Allocation type. Only SYSTEM supports the public lifecycle. Changes require replacement.",
				Required: true, Validators: []validator.String{stringvalidator.OneOf("SYSTEM")},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"pop_location": schema.SingleNestedAttribute{Description: "PoP reference, required for SYSTEM allocations. Changes require replacement.",
				Optional: true, Attributes: idname.SchemaNameID("PoP"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier(), objectplanmodifier.RequiresReplace()}},
		}}
}
