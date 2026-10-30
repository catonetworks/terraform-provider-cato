package secondaryawsvsocket

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
	resp.Schema = schema.Schema{Description: "Manages the secondary AWS vSocket configuration on an existing compatible AWS site. " +
		"Does not provision AWS infrastructure. Route table changes update the " +
		"site's shared route table configuration. " +
		"Import using the secondary socket ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Description: "Secondary socket ID",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"site": schema.SingleNestedAttribute{Description: "Parent AWS site. Changes require replacement.",
				Required:      true,
				Attributes:    idname.SchemaNameID("Site"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier(), objectplanmodifier.RequiresReplace()}},
			"eni_ip_address": schema.StringAttribute{Description: "Secondary ENI IPv4 address within eni_ip_subnet.",
				Required: true},
			"eni_ip_subnet": schema.StringAttribute{Description: "Canonical IPv4 network prefix for the secondary ENI.",
				Required: true},
			"route_table_id": schema.StringAttribute{Description: "AWS route table ID, shared by the site's socket connection configuration.",
				Required: true},
		}}
}
