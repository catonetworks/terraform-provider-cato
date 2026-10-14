package dhcp

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	nc "github.com/catonetworks/terraform-provider-cato/internal/provider/network/networkconfig"
)

type SettingsModel struct {
	ID           types.String `tfsdk:"id"`
	LeaseTime    types.Int64  `tfsdk:"lease_time"`
	RelayEnabled types.Bool   `tfsdk:"relay_enabled"`
	RelayTimeout types.Int64  `tfsdk:"relay_timeout"`
	RelayGroupID types.String `tfsdk:"relay_group_id"`
}

func (r *settingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Description: "Manages existing account DHCP settings via networkConfig.dhcp. " +
		"Destroy only removes Terraform ownership; remote settings are retained. Requires NCS_DHCP_PUBLIC_API.",
		Attributes: map[string]schema.Attribute{
			"id": nc.IDAttribute(),
			"lease_time": schema.Int64Attribute{Required: true,
				Description: "DHCP lease duration in seconds (minimum 60).",
				Validators:  []validator.Int64{int64validator.AtLeast(60)}},
			"relay_enabled": schema.BoolAttribute{Required: true, Description: "Enable DHCP relay. Enabling requires an effective group and " +
				"timeout."},
			"relay_timeout": schema.Int64Attribute{Optional: true,
				Computed:    true,
				Description: "Relay timeout in seconds (0–90). Omission preserves existing value; cannot be cleared.",
				Validators: []validator.Int64{int64validator.Between(0,
					90)}},
			"relay_group_id": schema.StringAttribute{Optional: true,
				Computed:    true,
				Description: "Opaque standalone relay group ID. Omission preserves selection; disable relay to stop using it."},
		}}
}
