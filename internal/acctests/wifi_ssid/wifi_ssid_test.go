//go:build acctest

package wifissid

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"testing"

	clientv2 "github.com/Yamashou/gqlgenc/clientv2"
	m "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccWifiSSID(t *testing.T) {
	acc.SkipByEnv(t)
	acc.CheckCMAVars(t)()
	ctx := context.Background()
	client := acc.GetClient(t)

	// Probe the rollout gate before creating a parent; Site 0 cannot exist.
	probe, probeErr := client.SiteWifiSsidList(ctx, acc.CatoAccountID, m.WifiSsidListInput{Site: &m.SiteRefInput{By: m.ObjectRefByID, Input: "0"}})
	if probeErr != nil {
		var api *clientv2.ErrorResponse
		if !errors.As(probeErr, &api) || api.GqlErrors == nil || len(*api.GqlErrors) != 1 || (*api.GqlErrors)[0] == nil || (*api.GqlErrors)[0].Message != "Site 0 not found" {
			t.Fatal("Wi-Fi SSID list API unavailable: enable its rollout and permissions before running this test")
		}
	} else if probe.GetSite().GetWifiSsidList() == nil {
		t.Fatal("Wi-Fi API probe returned an incomplete response")
	}
	parent, err := client.SiteAddSocketSite(ctx, m.AddSocketSiteInput{ConnectionType: m.SiteConnectionTypeEnumSocketX1600, Name: acc.GetRandName("wifi_site"), NativeNetworkRange: acc.GetRandNetworkPrefix() + ".0/24", SiteType: m.SiteTypeBranch, SiteLocation: &m.AddSiteLocationInput{CountryCode: "GB", Timezone: "Europe/London"}}, acc.CatoAccountID)
	if err != nil {
		t.Fatalf("create isolated Wi-Fi parent: %s", err)
	}
	siteID := parent.GetSite().GetAddSocketSite().GetSiteID()
	if siteID == "" {
		t.Fatal("parent returned no identity")
	}
	t.Logf("Owned Wi-Fi fixture parent: %s", siteID)
	t.Cleanup(func() { cleanupOwnedWifiSite(t, siteID) })
	prefix := acc.GetRandNetworkPrefix()
	name := acc.GetRandName("wifi")
	passkey := randomPasskey(t)
	config := func(stage int) string {
		suffix, band, dhcp := "", "BOTH", ""
		key := passkey
		if stage >= 1 {
			suffix = "_up"
			band = "BAND_5G"
			dhcp = fmt.Sprintf("dhcp = { dhcp_subnet = %q }", prefix+".128/25")
		}
		if stage >= 2 {
			dhcp = ""
			key = passkey + "u"
		}
		return acc.ProviderCfg() + fmt.Sprintf(`resource "cato_wifi_ssid" "test" {
 name = %q
 site = { id = %q }
 network_interface_id = "WBR1"
 category = "GUEST"
 band = %q
 enabled = false
 subnet = %q
 local_ip = %q
 mdns_enabled = false
 microsegmentation_enabled = false
 %s
 security = {
  mode = "PSK"
  auth_protocol = "WPA3"
  psk = { passkey = %q }
 }
}`, name+suffix, siteID, band, prefix+".0/24", prefix+".1", dhcp, key)
	}
	object := "cato_wifi_ssid.test"
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories, PreCheck: acc.CheckCMAVars(t), CheckDestroy: func(state *terraform.State) error { return checkWifiDestroyed(t, state) }, Steps: []resource.TestStep{
		{Config: config(0), Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttrSet(object, "id"), resource.TestCheckResourceAttr(object, "enabled", "false"))},
		{Config: config(0), PlanOnly: true, ExpectNonEmptyPlan: false},
		{ResourceName: object, ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"security.psk"}},
		{Config: config(1), Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr(object, "band", "BAND_5G"), resource.TestCheckResourceAttr(object, "dhcp.dhcp_subnet", prefix+".128/25"))},
		{Config: config(1), PlanOnly: true, ExpectNonEmptyPlan: false},
		{Config: config(2), Check: resource.TestCheckResourceAttr(object, "security.psk.passkey", passkey+"u")},
		{Config: config(2), PlanOnly: true, ExpectNonEmptyPlan: false},
	}})
}
func randomPasskey(t *testing.T) string {
	t.Helper()
	var raw [18]byte
	if _, err := rand.Read(raw[:]); err != nil {
		t.Fatal("generate temporary Wi-Fi passkey")
	}
	return base64.RawURLEncoding.EncodeToString(raw[:])
}
func cleanupOwnedWifiSite(t *testing.T, id string) {
	ctx := context.Background()
	client := acc.GetClient(t)
	result, err := client.SiteWifiSsidList(ctx, acc.CatoAccountID, m.WifiSsidListInput{Site: &m.SiteRefInput{By: m.ObjectRefByID, Input: id}})
	if err != nil {
		t.Error("read owned SSIDs before parent cleanup failed")
		return
	}
	payload := result.GetSite().GetWifiSsidList()
	if payload == nil || payload.GetPageInfo() == nil || payload.GetPageInfo().GetTotal() != int64(len(payload.GetWifiSsid())) {
		t.Error("incomplete SSID list; retaining parent")
		return
	}
	for _, item := range payload.GetWifiSsid() {
		if item == nil || item.GetID() == "" || item.GetSite().GetID() != id {
			t.Error("SSID ownership mismatch; retaining parent")
			return
		}
		_, err := client.SiteRemoveWifiSsid(ctx, acc.CatoAccountID, m.RemoveWifiSsidInput{WifiSsid: &m.WifiSsidRefInput{By: m.ObjectRefByID, Input: item.GetID()}})
		if err != nil {
			t.Error("remove owned SSID failed")
			return
		}
	}
	if _, err := client.SiteRemoveSite(ctx, id, acc.CatoAccountID); err != nil {
		t.Errorf("remove owned Wi-Fi parent: %s", err)
		return
	}
	limit, from := int64(1), int64(0)
	lookup, err := client.EntityLookup(ctx, acc.CatoAccountID, m.EntityTypeSite, &limit, &from, nil, nil, []string{id}, nil, nil, nil)
	if err != nil || lookup.GetEntityLookup() == nil || len(lookup.GetEntityLookup().GetItems()) != 0 {
		t.Error("owned Wi-Fi parent deletion unconfirmed")
		return
	}
	t.Logf("Owned Wi-Fi parent %s removed and absence verified", id)
}
func checkWifiDestroyed(t *testing.T, state *terraform.State) error {
	for _, item := range state.RootModule().Resources {
		if item.Type != "cato_wifi_ssid" || item.Primary == nil {
			continue
		}
		_, err := acc.GetClient(t).SiteWifiSsid(context.Background(), acc.CatoAccountID, m.WifiSsidRefInput{By: m.ObjectRefByID, Input: item.Primary.ID})
		var response *clientv2.ErrorResponse
		if !errors.As(err, &response) || response.GqlErrors == nil || len(*response.GqlErrors) == 0 {
			return fmt.Errorf("SSID absence not confirmed")
		}
		for _, e := range *response.GqlErrors {
			if e == nil || e.Message != "Wifi SSID "+item.Primary.ID+" not found" {
				return fmt.Errorf("SSID absence not confirmed")
			}
		}
	}
	return nil
}
