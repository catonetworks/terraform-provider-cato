package dnssettings

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"

	nc "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/networkconfig"
)

func (r *settingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attrs := nc.SchemaDNSAttributes()
	attrs["accept_dns_requests_on_lan_interface_ip"] = schema.BoolAttribute{
		Optional:    true,
		Computed:    true,
		Default:     booldefault.StaticBool(false),
		Description: "Accept DNS requests sent to Socket LAN interface IP.",
	}
	resp.Schema = schema.Schema{
		Description: "Manages existing account DNS settings via networkConfig.dns. " +
			"Destroy removes Terraform ownership and retains remote settings.",
		Attributes: attrs,
	}
}
