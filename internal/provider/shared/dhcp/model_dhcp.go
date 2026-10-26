package dhcp

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/tfmodel"
)

type Settings = tfmodel.DhcpSettings

var SettingsAttrTypes = map[string]attr.Type{
	"dhcp_type":              types.StringType,
	"ip_range":               types.StringType,
	"relay_group_id":         types.StringType,
	"relay_group_name":       types.StringType,
	"dhcp_microsegmentation": types.BoolType,
}
