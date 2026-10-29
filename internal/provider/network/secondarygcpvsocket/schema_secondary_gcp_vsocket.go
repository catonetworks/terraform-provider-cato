package secondarygcpvsocket

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/planmodifiers"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

func (r *socketResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Description: "Manages the secondary GCP vSocket configuration on an existing compatible GCP site. " +
		"Does not provision GCP infrastructure. The API requires a registered GCP " +
		"socket for read and update; " +
		"new unregistered sockets may be rejected until registration completes. " +
		"Import using the secondary socket ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Description: "Secondary socket ID",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"site": schema.SingleNestedAttribute{Description: "Parent GCP site. Changes require replacement.",
				Required:      true,
				Attributes:    idname.SchemaNameID("Site"),
				PlanModifiers: []planmodifier.Object{planmodifiers.IDNamePlanModifier(), objectplanmodifier.RequiresReplace()}},
			"gcp_config": schema.SingleNestedAttribute{Description: "Secondary GCP interface and load balancer configuration.", Required: true,
				Attributes: map[string]schema.Attribute{
					"interface_ip":     schema.StringAttribute{Description: "Secondary GCP interface IPv4 address.", Required: true},
					"load_balancer_ip": schema.StringAttribute{Description: "GCP HA load balancer IPv4 address.", Required: true},
				}},
		}}
}

var gcpConfigAttrTypes = map[string]attr.Type{"interface_ip": types.StringType, "load_balancer_ip": types.StringType}
