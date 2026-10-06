package dhcp

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/catonetworks/terraform-provider-cato/internal/provider/common/dhcp/application"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/common/parse"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/common/utils"
)

// DecodeSettings projects Terraform values without performing a relay lookup.
func DecodeSettings(ctx context.Context, value types.Object, diags *diag.Diagnostics) *application.Settings {
	if !utils.HasValue(value) {
		return nil
	}
	var settings Settings
	diags.Append(value.As(ctx, &settings, basetypes.ObjectAsOptions{})...)
	if diags.HasError() || !utils.HasValue(settings.DhcpType) {
		return nil
	}
	return &application.Settings{

		DhcpType:              settings.DhcpType.ValueString(),
		IPRange:               parse.KnownStringPointer(settings.IPRange),
		RelayGroupID:          parse.KnownStringPointer(settings.RelayGroupID),
		RelayGroupName:        parse.KnownStringPointer(settings.RelayGroupName),
		DhcpMicrosegmentation: parse.KnownBoolPointer(settings.DhcpMicrosegmentation),
	}
}

// ProjectSettings consumes a snapshot whose relay name has already been resolved.
func ProjectSettings(ctx context.Context, s *application.Settings, diags *diag.Diagnostics) types.Object {
	if s == nil {
		return types.ObjectNull(SettingsAttrTypes)
	}
	tf := Settings{

		DhcpType:              types.StringValue(s.DhcpType),
		IPRange:               types.StringNull(),
		RelayGroupID:          types.StringNull(),
		RelayGroupName:        types.StringNull(),
		DhcpMicrosegmentation: types.BoolNull(),
	}
	switch s.DhcpType {
	case "DHCP_RELAY":
		tf.RelayGroupID = types.StringPointerValue(s.RelayGroupID)
		tf.RelayGroupName = types.StringPointerValue(s.RelayGroupName)
	case "DHCP_RANGE":
		tf.IPRange = types.StringPointerValue(s.IPRange)
		tf.DhcpMicrosegmentation = types.BoolPointerValue(s.DhcpMicrosegmentation)
	case "ACCOUNT_DEFAULT", "DHCP_DISABLED":
	default:
		diags.AddError("Unsupported DHCP type", "Unknown DHCP type from API: "+s.DhcpType)
		return types.ObjectNull(SettingsAttrTypes)
	}
	value, ds := types.ObjectValueFrom(ctx, SettingsAttrTypes, tf)
	diags.Append(ds...)
	return value
}
