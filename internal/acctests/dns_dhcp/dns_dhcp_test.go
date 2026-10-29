//go:build acctest

package dns_dhcp

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
	nc "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/networkconfig"
)

const envEnableDNSDHCPCRUD = "TFACC_ENABLE_DNS_DHCP_CRUD"

// Opt-in live test: captures/restores account and site settings and creates only its own collection items.
func TestAccDNSDHCP(t *testing.T) {
	f := newLiveFixture(t)
	steps := []resource.TestStep{
		{Config: f.config("initial"), Check: f.check(false)},
		{Config: f.config("initial"), PlanOnly: true},
	}
	for _, name := range liveResources {
		steps = append(steps, resource.TestStep{ResourceName: "cato_" + name + ".test", ImportState: true, ImportStateVerify: true})
	}
	steps = append(steps,
		resource.TestStep{Config: f.config("updated"), Check: f.check(false)},
		resource.TestStep{Config: f.config("updated"), PlanOnly: true},
		resource.TestStep{Config: f.config("by_name"), PlanOnly: true},
		resource.TestStep{Config: f.config("omitted"), Check: f.check(false)},
		resource.TestStep{Config: f.config("omitted"), PlanOnly: true},
		resource.TestStep{Config: f.config("updated"), PreConfig: f.deleteOwnedItems, Check: f.check(true)},
		resource.TestStep{Config: f.config("updated"), PlanOnly: true},
		resource.TestStep{Config: f.config("cleared"), Check: f.check(false)},
		resource.TestStep{Config: f.config("cleared"), PlanOnly: true},
		resource.TestStep{Config: f.config("empty_description"), PlanOnly: true},
		resource.TestStep{Config: f.config("retained"), Check: f.check(false)},
		resource.TestStep{Config: f.config("retained"), PlanOnly: true},
	)
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories, Steps: steps, CheckDestroy: f.checkDestroy})
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
	_, err := nc.NewDNSClient(client).UpdateDNSSettings(ctx, acc.CatoAccountID, nc.DNSSettingsInput{
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
		_, err = nc.NewDNSClient(client).UpdateDNSSiteSettings(ctx, acc.CatoAccountID, models.SiteRefInput{By: models.ObjectRefByID, Input: siteID}, nc.DNSSettingsInput{
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
