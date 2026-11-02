package dhcpsettings

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
	nc "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/networkconfig"
)

func (r *settingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages existing account DHCP settings via networkConfig.dhcp. " +
			"Destroy removes Terraform ownership and retains remote settings.",
		Attributes: map[string]schema.Attribute{
			"id": nc.IDAttribute(),
			"lease_time": schema.Int64Attribute{
				Optional:      true,
				Computed:      true,
				Description:   "DHCP lease duration in seconds, minimum 60. Omission preserves its current value.",
				Validators:    []validator.Int64{int64validator.AtLeast(60)},
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"relay": schema.SingleNestedAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "DHCP relay configuration. Omission preserves the current configuration.",
				PlanModifiers: []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "Enable DHCP relay. Enabling requires an effective group and timeout.",
					},
					"group": schema.SingleNestedAttribute{
						Optional:      true,
						Computed:      true,
						Description:   "Existing relay group by ID or name. Omission preserves selection; disable relay to stop using it.",
						Attributes:    idname.SchemaNameID("Relay group"),
						PlanModifiers: []planmodifier.Object{relayGroupModifier{}},
					},
					"timeout": schema.Int64Attribute{
						Optional:      true,
						Computed:      true,
						Description:   "Relay timeout in seconds, 0–90. Omission preserves the value; API cannot clear it.",
						Validators:    []validator.Int64{int64validator.Between(0, 90)},
						PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
					},
				},
			},
		},
	}
}
