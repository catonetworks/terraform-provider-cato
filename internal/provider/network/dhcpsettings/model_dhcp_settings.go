package dhcpsettings

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/idname"
)

// SettingsModel is the account's singleton DHCP configuration.
type SettingsModel struct {
	ID        types.String `tfsdk:"id"`
	LeaseTime types.Int64  `tfsdk:"lease_time"`
	Relay     types.Object `tfsdk:"relay"`
}

// RelayModel contains relay enablement and a selected group reference.
type RelayModel struct {
	Enabled types.Bool   `tfsdk:"enabled"`
	Timeout types.Int64  `tfsdk:"timeout"`
	Group   types.Object `tfsdk:"group"`
}

// RelayAttrTypes maps the nested DHCP relay schema.
var RelayAttrTypes = map[string]attr.Type{"enabled": types.BoolType, "timeout": types.Int64Type, "group": idname.RefObjectType}

// RelayObjectType is the Terraform type of a relay configuration.
var RelayObjectType = types.ObjectType{AttrTypes: RelayAttrTypes}
