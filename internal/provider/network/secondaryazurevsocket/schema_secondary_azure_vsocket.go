package secondaryazurevsocket

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/planmodifiers"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

func (r *socketResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Description: "Manages the secondary Azure vSocket configuration on an existing compatible Azure site. " +
		"Does not provision Azure infrastructure. " +
		"Import using the secondary socket ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Description: "Secondary socket ID",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"site": schema.SingleNestedAttribute{Description: "Parent Azure site. Changes require replacement.",
				Required:      true,
				Attributes:    idname.SchemaNameID("Site"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier(), objectplanmodifier.RequiresReplace()}},
			"interface_ip": schema.StringAttribute{Description: "Secondary Azure interface IPv4 address within the parent site native range.",
				Required: true},
			"floating_ip": schema.StringAttribute{Description: "Azure HA floating IPv4 address within the parent site native range.",
				Required: true},
		}}
}
