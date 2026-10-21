package validators

import (
	"context"
	"testing"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"

	tf "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/tfmodel"
)

func TestDHCPTypeValidator(t *testing.T) {
	t.Parallel()
	checkStringValidator(t, DHCPTypeValidator{}, enumStrings(cato_models.AllDhcpType), true)
}
func dhcpConfig(t *testing.T, s tf.DhcpSettings) types.Object {
	t.Helper()
	v, d := types.ObjectValueFrom(context.Background(), tf.DhcpSettingsAttrTypes, s)
	require.False(t, d.HasError(), "%v", d)
	return v
}
func defaultDHCPConfig() tf.DhcpSettings {
	return tf.DhcpSettings{DhcpType: types.StringValue("DHCP_DISABLED"), IPRange: types.StringNull(), RelayGroupID: types.StringNull(), RelayGroupName: types.StringNull(), DhcpMicrosegmentation: types.BoolNull()}
}
func TestDHCPCheckerRules(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, kind, ip, id, group string
		micro                     bool
		wantError                 string
	}{
		{name: "disabled", kind: "DHCP_DISABLED"}, {name: "account defaults", kind: "ACCOUNT_DEFAULT"},
		{name: "range", kind: "DHCP_RANGE", ip: "192.0.2.1-192.0.2.20"},
		{name: "range with microsegmentation", kind: "DHCP_RANGE", ip: "192.0.2.1-192.0.2.20", micro: true},
		{name: "range missing IP range", kind: "DHCP_RANGE", wantError: "ip_range must be configured"},
		{name: "IP range with disabled DHCP", kind: "DHCP_DISABLED", ip: "192.0.2.1-192.0.2.20", wantError: "ip_range can only"},
		{name: "microsegmentation outside range", kind: "DHCP_DISABLED", micro: true, wantError: "dhcp_microsegmentation"},
		{name: "relay by ID", kind: "DHCP_RELAY", id: "relay-1"}, {name: "relay by name", kind: "DHCP_RELAY", group: "relay-group"},
		{name: "relay without group", kind: "DHCP_RELAY", wantError: "either relay_group_id or relay_group_name"},
		{name: "relay with both selectors", kind: "DHCP_RELAY", id: "relay-1", group: "relay-group", wantError: "only one of"},
		{name: "relay selector with disabled DHCP", kind: "DHCP_DISABLED", id: "relay-1", wantError: "can only be configured"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := defaultDHCPConfig()
			s.DhcpType = types.StringValue(tc.kind)
			if tc.ip != "" {
				s.IPRange = types.StringValue(tc.ip)
			}
			if tc.id != "" {
				s.RelayGroupID = types.StringValue(tc.id)
			}
			if tc.group != "" {
				s.RelayGroupName = types.StringValue(tc.group)
			}
			s.DhcpMicrosegmentation = types.BoolValue(tc.micro)
			var d diag.Diagnostics
			err := DHCPChecker.Check(context.Background(), &d, dhcpConfig(t, s))
			require.Equal(t, tc.wantError != "", err != nil)
			require.Equal(t, tc.wantError != "", d.HasError(), "%v", d)
			if tc.wantError != "" {
				require.Contains(t, d[0].Detail(), tc.wantError)
			}
		})
	}
}
func TestDHCPCheckerPriorState(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	prior := defaultDHCPConfig()
	prior.DhcpType = types.StringValue("DHCP_RELAY")
	prior.RelayGroupID = types.StringValue("relay-1")
	prior.RelayGroupName = types.StringValue("old-name")
	for _, tc := range []struct {
		name      string
		change    func(*tf.DhcpSettings)
		wantError bool
	}{
		{"unchanged selectors", func(*tf.DhcpSettings) {}, false},
		{"new name with propagated ID", func(s *tf.DhcpSettings) { s.RelayGroupName = types.StringValue("new-name") }, false},
		{"new ID with propagated name", func(s *tf.DhcpSettings) { s.RelayGroupID = types.StringValue("relay-2") }, false},
		{"both selectors explicitly changed", func(s *tf.DhcpSettings) {
			s.RelayGroupName = types.StringValue("new-name")
			s.RelayGroupID = types.StringValue("relay-2")
		}, true},
		{"disabled with propagated selectors", func(s *tf.DhcpSettings) { s.DhcpType = types.StringValue("DHCP_DISABLED") }, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := prior
			tc.change(&s)
			var d diag.Diagnostics
			err := DHCPChecker.CheckWithPriorState(ctx, &d, dhcpConfig(t, s), dhcpConfig(t, prior))
			require.Equal(t, tc.wantError, err != nil)
			require.Equal(t, tc.wantError, d.HasError())
		})
	}
}
func TestDHCPCheckerUnresolvedAndMalformedObjects(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	unknown := defaultDHCPConfig()
	unknown.RelayGroupID = types.StringUnknown()
	bad := types.ObjectValueMust(map[string]attr.Type{"dhcp_type": types.Int64Type}, map[string]attr.Value{"dhcp_type": types.Int64Value(1)})
	for _, tc := range []struct {
		name          string
		config, prior types.Object
		wantError     bool
	}{
		{"null", types.ObjectNull(tf.DhcpSettingsAttrTypes), types.ObjectNull(tf.DhcpSettingsAttrTypes), false},
		{"unknown", types.ObjectUnknown(tf.DhcpSettingsAttrTypes), types.ObjectNull(tf.DhcpSettingsAttrTypes), false},
		{"unknown field", dhcpConfig(t, unknown), types.ObjectNull(tf.DhcpSettingsAttrTypes), false},
		{"unknown prior field", dhcpConfig(t, defaultDHCPConfig()), dhcpConfig(t, unknown), false},
		{"malformed config", bad, types.ObjectNull(tf.DhcpSettingsAttrTypes), true},
		{"malformed prior state", dhcpConfig(t, defaultDHCPConfig()), bad, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var d diag.Diagnostics
			err := DHCPChecker.CheckWithPriorState(ctx, &d, tc.config, tc.prior)
			require.Equal(t, tc.wantError, err != nil)
			require.Equal(t, tc.wantError, d.HasError())
		})
	}
}
