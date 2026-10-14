//go:build acctest

package dns_dhcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	cato "github.com/catonetworks/cato-go-sdk"
	models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

// Opt-in live test: captures/restores account and site settings and creates only its own collection items.
func TestAccDNSDHCP(t *testing.T) {
	siteID, servers, tag := liveInputs(t)
	sdk, err := cato.New(os.Getenv("CATO_BASEURL"), os.Getenv("CATO_TOKEN"), acc.CatoAccountID, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	client := catoclient.NewProviderSDKClient(sdk)
	ctx := context.Background()
	dns, err := sdk.NetworkConfigDNSSettings(ctx, acc.CatoAccountID)
	if err != nil {
		t.Fatal(err)
	}
	dhcp, err := sdk.NetworkConfigDhcpSettings(ctx, acc.CatoAccountID)
	if err != nil {
		t.Fatal(err)
	}
	site, err := sdk.NetworkConfigDNSSiteSettings(ctx, acc.CatoAccountID, models.SiteRefInput{By: models.ObjectRefByID, Input: siteID})
	if err != nil {
		t.Fatal(err)
	}
	if dns.NetworkConfig == nil || dns.NetworkConfig.DNS.Settings == nil || dhcp.NetworkConfig == nil || dhcp.NetworkConfig.Dhcp.Settings == nil || site.NetworkConfig == nil {
		t.Fatal("Missing baseline settings")
	}
	t.Cleanup(func() { restoreBaseline(t, client, dns, dhcp, site, siteID) })
	name := fmt.Sprintf("acctest-dns-dhcp-%d", time.Now().UnixNano())
	dnsConfig := dns.NetworkConfig.DNS.Settings
	dhcpConfig := dhcp.NetworkConfig.Dhcp.Settings
	primary, secondary, suffix := hcl(dnsConfig.PrimaryServer), hcl(dnsConfig.SecondaryServer), hcl(dnsConfig.Suffix)
	sitePrimary, siteSecondary, siteSuffix := primary, secondary, suffix
	if original := site.NetworkConfig.DNS.SiteSettings; original != nil {
		sitePrimary, siteSecondary, siteSuffix = hcl(original.PrimaryServer), hcl(original.SecondaryServer), hcl(original.Suffix)
	}
	group := "null"
	if dhcpConfig.Relay.Group != nil {
		group = hcl(dhcpConfig.Relay.Group.ID)
	}
	config := func(value string) string {
		return fmt.Sprintf(`
provider "cato" {
 account_id = %q
 retry_max = 0
 version_check_disabled = true
}
resource "cato_dns_settings" "test" {
 primary_server = %s
 secondary_server = %s
 suffix = %s
 accept_dns_requests_on_lan_interface_ip = %t
}
resource "cato_dns_site_settings" "test" {
 site_id = %q
 primary_server = %s
 secondary_server = %s
 suffix = %s
}
resource "cato_dns_forwarding_rule" "test" {
 domain = %q
 server = %s
}
resource "cato_dhcp_settings" "test" {
 lease_time = %d
 relay_enabled = %t
 relay_timeout = %s
 relay_group_id = %s
}
resource "cato_dhcp_option" "test" {
 tag = %d
 type = "ASCII"
 value = %q
 description = %q
}
resource "cato_dhcp_relay_group" "test" {
 name = %q
 server = %s
}
`, acc.CatoAccountID, primary, secondary, suffix, dnsConfig.AcceptDNSRequestsOnLanInterfaceIP,
			siteID, sitePrimary, siteSecondary, siteSuffix, name+".test", hcl(servers),
			dhcpConfig.LeaseTime, dhcpConfig.Relay.Enabled, hcl(dhcpConfig.Relay.Timeout), group,
			tag, value, name, name, hcl(servers))
	}
	steps := []resource.TestStep{{Config: config("initial")}, {Config: config("initial"), PlanOnly: true}}
	for _, res := range []string{"dns_settings", "dns_site_settings", "dns_forwarding_rule", "dhcp_settings", "dhcp_option", "dhcp_relay_group"} {
		steps = append(steps, resource.TestStep{ResourceName: "cato_" + res + ".test", ImportState: true, ImportStateVerify: true})
	}
	steps = append(steps, resource.TestStep{Config: config("updated")}, resource.TestStep{Config: config("updated"), PlanOnly: true})
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories, Steps: steps})
}

func hcl(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return string(encoded)
}

func restoreBaseline(t *testing.T, client *catoclient.ProviderSDKClient, dns *cato.NetworkConfigDNSSettings,
	dhcp *cato.NetworkConfigDhcpSettings, site *cato.NetworkConfigDNSSiteSettings, siteID string) {
	t.Helper()
	ctx := context.Background()
	settings := dns.NetworkConfig.DNS.Settings
	_, err := catoclient.NewDNSClient(client).UpdateDNSSettings(ctx, acc.CatoAccountID, catoclient.DNSSettingsInput{
		PrimaryServer: settings.PrimaryServer, SecondaryServer: settings.SecondaryServer, Suffix: settings.Suffix,
		AcceptDNSRequestsOnLanInterfaceIP: settings.AcceptDNSRequestsOnLanInterfaceIP})
	if err != nil {
		t.Errorf("restore DNS settings: %v", err)
	}
	original := dhcp.NetworkConfig.Dhcp.Settings
	relay := &models.NetworkConfigDhcpRelayInput{Enabled: &original.Relay.Enabled, Timeout: original.Relay.Timeout}
	if original.Relay.Group != nil {
		relay.Group = &models.NetworkConfigDhcpRelayGroupRefInput{By: models.ObjectRefByID, Input: original.Relay.Group.ID}
	}
	_, err = client.NetworkConfigDhcpUpdateSettings(ctx, acc.CatoAccountID, models.NetworkConfigDhcpUpdateSettingsInput{LeaseTime: &original.LeaseTime, Relay: relay})
	if err != nil {
		t.Errorf("restore DHCP settings: %v", err)
	}
	if override := site.NetworkConfig.DNS.SiteSettings; override != nil {
		_, err = catoclient.NewDNSClient(client).UpdateDNSSiteSettings(ctx, acc.CatoAccountID, siteID, catoclient.DNSSettingsInput{
			PrimaryServer: override.PrimaryServer, SecondaryServer: override.SecondaryServer, Suffix: override.Suffix})
	} else {
		_, err = client.NetworkConfigDNSRemoveSiteSettings(ctx, acc.CatoAccountID, models.NetworkConfigDNSRemoveSiteSettingsInput{
			Site: []*models.SiteRefInput{{By: models.ObjectRefByID, Input: siteID}}})
	}
	if err != nil {
		t.Errorf("restore site DNS: %v", err)
	}
}

func liveInputs(t *testing.T) (siteID string, servers []string, tag int) {
	t.Helper()
	acc.CheckCMAVars(t)()
	if os.Getenv("CATO_BASEURL") != "https://system.cc.dev-us1-m1.catonet.works/api/v1/graphql" {
		t.Fatal("DNS/DHCP suite requires dev-us1-m1")
	}
	if os.Getenv("DISABLE_POLICY_RULE_CLEANUP") != "true" {
		t.Fatal("Set DISABLE_POLICY_RULE_CLEANUP=true")
	}
	if os.Getenv("TF_ACC_MOCK") != "" {
		t.Fatal("Unset TF_ACC_MOCK for live tests")
	}
	siteID = os.Getenv("CATO_TEST_SITE_ID")
	if siteID == "" {
		t.Fatal("Set CATO_TEST_SITE_ID to an existing site in the test account")
	}
	servers = strings.Split(os.Getenv("TFACC_DNS_DHCP_SERVERS"), ",")
	if servers[0] == "" {
		t.Fatal("Set TFACC_DNS_DHCP_SERVERS to test DNS/relay IP addresses")
	}
	tag, err := strconv.Atoi(os.Getenv("TFACC_DHCP_OPTION_TAG"))
	if err != nil {
		t.Fatal("Set TFACC_DHCP_OPTION_TAG to an unused option code")
	}

	return siteID, servers, tag
}
