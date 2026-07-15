package provider

import (
	"context"
	"testing"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestHydrateUpdateIpsecIkeV2SiteTunnelsRestoresComputedTunnelIDs(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	plan := SiteIpsecIkeV2{
		IPSec: testIPSecObject(
			testIPSecTunnelGroup(types.StringUnknown(), "new-primary-psk"),
			testIPSecTunnelGroup(types.StringValue("SECONDARY2"), "new-secondary-psk"),
		),
	}
	state := SiteIpsecIkeV2{
		IPSec: testIPSecObject(
			testIPSecTunnelGroup(types.StringValue("PRIMARY1"), "old-primary-psk"),
			testIPSecTunnelGroup(types.StringValue("SECONDARY1"), "old-secondary-psk"),
		),
	}

	input, diags := hydrateUpdateIpsecIkeV2SiteTunnels(ctx, plan, state)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if got := input.Primary.Tunnels[0].TunnelID; got != cato_models.IPSecV2InterfaceID("PRIMARY1") {
		t.Fatalf("expected primary tunnel ID PRIMARY1 from state, got %q", got)
	}
	if got := input.Primary.Tunnels[0].Psk; got == nil || *got != "new-primary-psk" {
		t.Fatalf("expected updated primary PSK, got %v", got)
	}
	if got := input.Secondary.Tunnels[0].TunnelID; got != cato_models.IPSecV2InterfaceID("SECONDARY2") {
		t.Fatalf("expected known planned secondary tunnel ID SECONDARY2, got %q", got)
	}
}

func TestHydrateUpdateIpsecIkeV2SiteTunnelsDerivesMissingTunnelID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	plan := SiteIpsecIkeV2{
		IPSec: testIPSecObject(
			testIPSecTunnelGroup(types.StringUnknown(), "new-primary-psk"),
			types.ObjectNull(IpsecTunnelsResourceAttrTypes),
		),
	}
	state := SiteIpsecIkeV2{
		IPSec: testIPSecObject(
			testIPSecTunnelGroup(types.StringNull(), "old-primary-psk"),
			types.ObjectNull(IpsecTunnelsResourceAttrTypes),
		),
	}

	input, diags := hydrateUpdateIpsecIkeV2SiteTunnels(ctx, plan, state)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if got := input.Primary.Tunnels[0].TunnelID; got != cato_models.IPSecV2InterfaceID("PRIMARY1") {
		t.Fatalf("expected derived primary tunnel ID PRIMARY1, got %q", got)
	}
}

func testIPSecObject(primary, secondary types.Object) types.Object {
	return types.ObjectValueMust(IpsecResourceAttrTypes, map[string]attr.Value{
		"site_id":             types.StringValue("site-123"),
		"primary":             primary,
		"secondary":           secondary,
		"connection_mode":     types.StringValue("RESPONDER_ONLY"),
		"identification_type": types.StringValue("FQDN"),
		"init_message":        types.ObjectNull(IpsecMessageResourceAttrTypes),
		"auth_message":        types.ObjectNull(IpsecMessageResourceAttrTypes),
		"network_ranges":      types.ListNull(types.StringType),
	})
}

func testIPSecTunnelGroup(tunnelID types.String, psk string) types.Object {
	tunnel := types.ObjectValueMust(TunnelResourceAttrTypes, map[string]attr.Value{
		"tunnel_id":       tunnelID,
		"public_site_ip":  types.StringNull(),
		"private_cato_ip": types.StringNull(),
		"private_site_ip": types.StringNull(),
		"psk":             types.StringValue(psk),
		"last_mile_bw":    types.ObjectNull(LastMileBwResourceAttrTypes),
	})

	return types.ObjectValueMust(IpsecTunnelsResourceAttrTypes, map[string]attr.Value{
		"destination_type":  types.StringValue("FQDN"),
		"public_cato_ip_id": types.StringNull(),
		"pop_location_id":   types.StringNull(),
		"tunnels": types.ListValueMust(
			types.ObjectType{AttrTypes: TunnelResourceAttrTypes},
			[]attr.Value{tunnel},
		),
	})
}
