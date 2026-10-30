//go:build acctest

package dns_dhcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	cato "github.com/catonetworks/cato-go-sdk"
	models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
	"github.com/catonetworks/terraform-provider-cato/internal/provider/shared/catoclient"
)

var liveResources = []string{"dns_settings", "dns_site_settings", "dns_forwarding_rule", "dhcp_settings", "dhcp_option", "dhcp_relay_group"}

type liveFixture struct {
	t                      *testing.T
	sdk                    *cato.Client
	client                 *catoclient.ProviderSDKClient
	siteID, siteName, name string
	servers                []string
	tag                    int
	baseline               *liveSnapshot
	last                   *liveSnapshot
	ids                    map[string]string
}

type liveSnapshot struct {
	dns         *cato.NetworkConfigDNSSettings
	dhcp        *cato.NetworkConfigDhcpSettings
	site        *cato.NetworkConfigDNSSiteSettings
	collections map[string]map[string]string
}

func newLiveFixture(t *testing.T) *liveFixture {
	t.Helper()
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("TF_ACC=1 required before live fixture creation")
	}
	if os.Getenv(envEnableDNSDHCPCRUD) != "true" {
		t.Skipf("set %s=true to run DNS/DHCP acceptance tests", envEnableDNSDHCPCRUD)
	}
	siteID, servers, tag := liveInputs(t)
	require.GreaterOrEqual(t, len(servers), 2, "Provide at least two test server IPs to verify ordered updates")
	sdk, err := cato.New(os.Getenv("CATO_BASEURL"), os.Getenv("CATO_TOKEN"), acc.CatoAccountID, &http.Client{Timeout: time.Minute}, nil)
	require.NoError(t, err)
	f := &liveFixture{t: t, sdk: sdk, client: catoclient.NewProviderSDKClient(sdk), siteID: siteID, servers: servers, tag: tag, name: fmt.Sprintf("acctest-dns-dhcp-%d", time.Now().UnixNano()), ids: map[string]string{}}
	f.baseline = f.snapshot()
	details, err := sdk.SiteGeneralDetails(context.Background(), models.SiteRefInput{By: models.ObjectRefByID, Input: siteID}, acc.CatoAccountID)
	require.NoError(t, err)
	require.NotNil(t, details.GetSite().GetSiteGeneralDetails())
	f.siteName = details.Site.SiteGeneralDetails.Site.Name
	for _, encoded := range f.baseline.collections["dhcp_option"] {
		var item struct {
			Tag int `json:"tag"`
		}
		require.NoError(t, json.Unmarshal([]byte(encoded), &item))
		require.NotEqual(t, tag, item.Tag, "Select an unused option code")
	}
	t.Cleanup(func() {
		restoreBaseline(t, f.client, f.baseline.dns, f.baseline.dhcp, f.baseline.site, f.siteID)
		f.cleanupOwnedCollections()
		actual := f.snapshot()
		require.Equal(t, hcl(f.baseline.dns), hcl(actual.dns), "original account DNS settings restored")
		require.Equal(t, hcl(f.baseline.dhcp), hcl(actual.dhcp), "original account DHCP settings restored")
		require.Equal(t, hcl(f.baseline.site), hcl(actual.site), "original site override/inheritance restored")
		require.Equal(t, f.baseline.collections, actual.collections, "all original collection items restored")
	})
	return f
}

func (f *liveFixture) snapshot() *liveSnapshot {
	f.t.Helper()
	ctx := context.Background()
	s := &liveSnapshot{}
	var err error
	s.dns, err = f.sdk.NetworkConfigDNSSettings(ctx, acc.CatoAccountID)
	require.NoError(f.t, err)
	require.NotNil(f.t, s.dns.GetNetworkConfig().GetDNS().GetSettings())
	s.dhcp, err = f.sdk.NetworkConfigDhcpSettings(ctx, acc.CatoAccountID)
	require.NoError(f.t, err)
	require.NotNil(f.t, s.dhcp.GetNetworkConfig().GetDhcp().GetSettings())
	s.site, err = f.sdk.NetworkConfigDNSSiteSettings(ctx, acc.CatoAccountID, models.SiteRefInput{By: models.ObjectRefByID, Input: f.siteID})
	require.NoError(f.t, err)
	require.NotNil(f.t, s.site.NetworkConfig)
	s.collections = map[string]map[string]string{}
	for _, name := range []string{"dns_forwarding_rule", "dhcp_option", "dhcp_relay_group"} {
		s.collections[name] = f.collection(name)
	}
	return s
}

func (f *liveFixture) collection(name string) map[string]string {
	items := map[string]string{}
	ctx := context.Background()
	for offset := int64(0); ; {
		paging := &models.PagingInput{From: offset, Limit: 100}
		sort := &models.SortOrderInput{Direction: models.SortOrderAsc, Priority: 1}
		var total int64
		before := len(items)
		switch name {
		case "dns_forwarding_rule":
			result, err := f.sdk.NetworkConfigDNSForwardingRuleList(ctx, acc.CatoAccountID, models.NetworkConfigDNSForwardingRuleListInput{Paging: paging, Sort: &models.NetworkConfigDNSForwardingRuleSortInput{Domain: sort}})
			require.NoError(f.t, err)
			require.NotNil(f.t, result.NetworkConfig)
			page := result.NetworkConfig.DNS.ForwardingRuleList
			require.NotNil(f.t, page)
			addCollectionItems(items, page.Items)
			total = page.Paging.Total
		case "dhcp_option":
			result, err := f.sdk.NetworkConfigDhcpOptionList(ctx, acc.CatoAccountID, models.NetworkConfigDhcpOptionListInput{Paging: paging, Sort: &models.NetworkConfigDhcpOptionSortInput{Tag: sort}})
			require.NoError(f.t, err)
			require.NotNil(f.t, result.NetworkConfig)
			page := result.NetworkConfig.Dhcp.OptionList
			require.NotNil(f.t, page)
			addCollectionItems(items, page.Items)
			total = page.Paging.Total
		case "dhcp_relay_group":
			result, err := f.sdk.NetworkConfigDhcpRelayGroupList(ctx, acc.CatoAccountID, models.NetworkConfigDhcpRelayGroupListInput{Paging: paging, Sort: &models.NetworkConfigDhcpRelayGroupSortInput{Name: sort}})
			require.NoError(f.t, err)
			require.NotNil(f.t, result.NetworkConfig)
			page := result.NetworkConfig.Dhcp.RelayGroupList
			require.NotNil(f.t, page)
			addCollectionItems(items, page.Items)
			total = page.Paging.Total
		}
		if int64(len(items)) == total {
			return items
		}
		require.Greater(f.t, len(items), before, "incomplete or overlapping collection pagination")
		offset = int64(len(items))
	}
}

func addCollectionItems[T interface{ GetID() string }](out map[string]string, items []T) {
	for _, item := range items {
		out[item.GetID()] = hcl(item)
	}
}

func (f *liveFixture) config(mode string) string {
	dns := f.baseline.dns.NetworkConfig.DNS.Settings
	dhcp := f.baseline.dhcp.NetworkConfig.Dhcp.Settings
	primary, secondary, suffix := hcl(dns.PrimaryServer), hcl(dns.SecondaryServer), hcl(dns.Suffix)
	sitePrimary, siteSecondary, siteSuffix := primary, secondary, suffix
	if original := f.baseline.site.NetworkConfig.DNS.SiteSettings; original != nil {
		sitePrimary, siteSecondary, siteSuffix = hcl(original.PrimaryServer), hcl(original.SecondaryServer), hcl(original.Suffix)
	}
	name, domain, value, description := f.name, f.name+".test", "initial", hcl(f.name)
	servers := append([]string{}, f.servers...)
	lease := dhcp.LeaseTime
	accept := dns.AcceptDNSRequestsOnLanInterfaceIP
	group := "null"
	if dhcp.Relay.Group != nil {
		group = "{ id = " + hcl(dhcp.Relay.Group.ID) + " }"
	}
	relay := fmt.Sprintf("{ enabled = %t, timeout = %s, group = %s }", dhcp.Relay.Enabled, hcl(dhcp.Relay.Timeout), group)
	siteRef := "{ id = " + hcl(f.siteID) + " }"
	if mode != "initial" {
		primary, secondary = hcl(f.servers[1]), hcl(f.servers[0])
		suffix = hcl([]string{f.name + ".test", "second.test"})
		sitePrimary, siteSecondary, siteSuffix = secondary, primary, hcl([]string{"site." + f.name + ".test"})
		name, domain, value, description = f.name+"-updated", "updated."+f.name+".test", "updated", hcl(f.name+"-updated")
		for left, right := 0, len(servers)-1; left < right; left, right = left+1, right-1 {
			servers[left], servers[right] = servers[right], servers[left]
		}
		lease += 60
		accept = !accept
		if mode != "retained" && dhcp.Relay.Group != nil && dhcp.Relay.Timeout != nil {
			field := "id"
			if mode == "by_name" {
				field = "name"
			}
			timeout := *dhcp.Relay.Timeout + 1
			if timeout > 90 {
				timeout = 0
			}
			relay = fmt.Sprintf("{ enabled = true, timeout = %d, group = { %s = cato_dhcp_relay_group.test.%s } }", timeout, field, field)
		}
	}
	if mode == "by_name" {
		siteRef = "{ name = " + hcl(f.siteName) + " }"
	}
	if mode == "cleared" || mode == "empty_description" {
		primary, secondary, suffix = "null", "null", "[]"
		sitePrimary, siteSecondary, siteSuffix = primary, secondary, suffix
		description = "null"
		if mode == "empty_description" {
			description = `""`
		}
	}
	dhcpFields := fmt.Sprintf("lease_time = %d\n relay = %s", lease, relay)
	if mode == "omitted" {
		dhcpFields = ""
	}
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
 depends_on = [cato_dns_settings.test]
 site = %s
 primary_server = %s
 secondary_server = %s
 suffix = %s
}
resource "cato_dns_forwarding_rule" "test" {
 depends_on = [cato_dns_site_settings.test]
 domain = %q
 server = %s
}
resource "cato_dhcp_relay_group" "test" {
 depends_on = [cato_dns_forwarding_rule.test]
 name = %q
 server = %s
}
resource "cato_dhcp_settings" "test" {
 depends_on = [cato_dhcp_relay_group.test]
 %s
}
resource "cato_dhcp_option" "test" {
 depends_on = [cato_dhcp_settings.test]
 tag = %d
 type = "ASCII"
 value = %q
 description = %s
}
`, acc.CatoAccountID, primary, secondary, suffix, accept, siteRef, sitePrimary, siteSecondary, siteSuffix, domain, hcl(servers), name, hcl(servers), dhcpFields, f.tag, value, description)
}

func (f *liveFixture) check(recreated bool) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		actual := f.snapshot()
		for _, name := range liveResources {
			item, ok := state.RootModule().Resources["cato_"+name+".test"]
			if !ok {
				return fmt.Errorf("missing state for %s", name)
			}
			id := item.Primary.ID
			if id == "" {
				return fmt.Errorf("empty ID for %s", name)
			}
			collection := actual.collections[name]
			if old := f.ids[name]; old != "" && (!recreated || collection == nil) && old != id {
				return fmt.Errorf("%s ID changed", name)
			}
			f.ids[name] = id
			attrs := item.Primary.Attributes
			expected, err := apiExpected(name, id, actual)
			if err != nil {
				return err
			}
			for key, value := range expected {
				if attrs[key] != value {
					return fmt.Errorf("%s.%s: state %q differs from API %q", name, key, attrs[key], value)
				}
			}
		}
		for name, original := range f.baseline.collections {
			remaining := map[string]string{}
			for id, value := range actual.collections[name] {
				if id != f.ids[name] {
					remaining[id] = value
				}
			}
			if !reflect.DeepEqual(original, remaining) {
				return fmt.Errorf("unrelated %s items changed", name)
			}
		}
		f.last = actual
		return nil
	}
}

func dnsExpected(out map[string]string, primary, secondary *string, suffix []string) {
	out["primary_server"] = ""
	if primary != nil {
		out["primary_server"] = *primary
	}
	out["secondary_server"] = ""
	if secondary != nil {
		out["secondary_server"] = *secondary
	}
	out["suffix.#"] = strconv.Itoa(len(suffix))
	for i, value := range suffix {
		out[fmt.Sprintf("suffix.%d", i)] = value
	}
}

func (f *liveFixture) deleteOwnedItems() {
	ctx := context.Background()
	// Restore the original selection before deleting a test-owned relay group.
	original := f.baseline.dhcp.NetworkConfig.Dhcp.Settings
	relay := &models.NetworkConfigDhcpRelayInput{Enabled: &original.Relay.Enabled, Timeout: original.Relay.Timeout}
	if original.Relay.Group != nil {
		relay.Group = &models.NetworkConfigDhcpRelayGroupRefInput{By: models.ObjectRefByID, Input: original.Relay.Group.ID}
	}
	_, err := f.sdk.NetworkConfigDhcpUpdateSettings(ctx, acc.CatoAccountID, models.NetworkConfigDhcpUpdateSettingsInput{Relay: relay})
	require.NoError(f.t, err)
	for _, name := range []string{"dns_forwarding_rule", "dhcp_option", "dhcp_relay_group"} {
		require.NoError(f.t, f.deleteCollectionItem(name, f.ids[name]))
	}
	_, err = f.sdk.NetworkConfigDNSRemoveSiteSettings(ctx, acc.CatoAccountID, models.NetworkConfigDNSRemoveSiteSettingsInput{Site: []*models.SiteRefInput{{By: models.ObjectRefByID, Input: f.siteID}}})
	require.NoError(f.t, err)
}

func (f *liveFixture) checkDestroy(_ *terraform.State) error {
	actual := f.snapshot()
	if f.last == nil {
		return fmt.Errorf("no successful apply to verify")
	}
	if hcl(f.last.dns) != hcl(actual.dns) || hcl(f.last.dhcp) != hcl(actual.dhcp) {
		return fmt.Errorf("account settings changed during destroy")
	}
	if actual.site.NetworkConfig.DNS.SiteSettings != nil {
		return fmt.Errorf("site override survived destroy")
	}
	if !reflect.DeepEqual(f.baseline.collections, actual.collections) {
		return fmt.Errorf("test-owned items remain or unrelated items changed")
	}
	return nil
}

func (f *liveFixture) deleteCollectionItem(name, id string) error {
	ctx := context.Background()
	switch name {
	case "dns_forwarding_rule":
		_, err := f.sdk.NetworkConfigDNSDeleteForwardingRule(ctx, acc.CatoAccountID, models.NetworkConfigDNSDeleteForwardingRuleInput{ID: []string{id}})
		return err
	case "dhcp_option":
		_, err := f.sdk.NetworkConfigDhcpDeleteOption(ctx, acc.CatoAccountID, models.NetworkConfigDhcpDeleteOptionInput{ID: []string{id}})
		return err
	case "dhcp_relay_group":
		_, err := f.sdk.NetworkConfigDhcpDeleteRelayGroup(ctx, acc.CatoAccountID, models.NetworkConfigDhcpDeleteRelayGroupInput{ID: []string{id}})
		return err
	default:
		return fmt.Errorf("unknown test collection %s", name)
	}
}

func (f *liveFixture) cleanupOwnedCollections() {
	for _, name := range []string{"dns_forwarding_rule", "dhcp_option", "dhcp_relay_group"} {
		for id, raw := range f.collection(name) {
			var item struct {
				Domain, Name string
				Description  *string
			}
			require.NoError(f.t, json.Unmarshal([]byte(raw), &item))
			owned := id == f.ids[name] || item.Domain == f.name+".test" || item.Domain == "updated."+f.name+".test" || item.Name == f.name || item.Name == f.name+"-updated"
			if item.Description != nil {
				owned = owned || *item.Description == f.name || *item.Description == f.name+"-updated"
			}
			if owned {
				if err := f.deleteCollectionItem(name, id); err != nil {
					f.t.Errorf("clean up owned %s %s: %v", name, id, err)
				}
			}
		}
	}
}

func apiExpected(name, id string, actual *liveSnapshot) (map[string]string, error) {
	expected := map[string]string{}
	switch name {
	case "dns_settings":
		dns := actual.dns.NetworkConfig.DNS.Settings
		dnsExpected(expected, dns.PrimaryServer, dns.SecondaryServer, dns.Suffix)
		expected["accept_dns_requests_on_lan_interface_ip"] = strconv.FormatBool(dns.AcceptDNSRequestsOnLanInterfaceIP)
	case "dns_site_settings":
		site := actual.site.NetworkConfig.DNS.SiteSettings
		if site == nil {
			return nil, fmt.Errorf("site override absent")
		}
		dnsExpected(expected, site.PrimaryServer, site.SecondaryServer, site.Suffix)
	case "dhcp_settings":
		dhcp := actual.dhcp.NetworkConfig.Dhcp.Settings
		expected["lease_time"] = strconv.FormatInt(dhcp.LeaseTime, 10)
		expected["relay.enabled"] = strconv.FormatBool(dhcp.Relay.Enabled)
		if dhcp.Relay.Timeout != nil {
			expected["relay.timeout"] = strconv.FormatInt(*dhcp.Relay.Timeout, 10)
		}
		if dhcp.Relay.Group != nil {
			expected["relay.group.id"] = dhcp.Relay.Group.ID
			expected["relay.group.name"] = dhcp.Relay.Group.Name
		}
	default:
		return collectionExpected(name, id, actual)
	}
	return expected, nil
}

func collectionExpected(name, id string, actual *liveSnapshot) (map[string]string, error) {
	expected := map[string]string{}

	raw, ok := actual.collections[name][id]
	if !ok {
		return nil, fmt.Errorf("%s not found through API", name)
	}
	var fields map[string]any
	if err := json.Unmarshal([]byte(raw), &fields); err != nil {
		return nil, err
	}
	for _, key := range []string{"domain", "name", "value", "type", "tag", "description"} {
		if value, present := fields[key]; present && value != nil {
			expected[key] = fmt.Sprint(value)
		}
	}
	if name == "dhcp_option" {
		if _, ok := expected["description"]; !ok {
			expected["description"] = ""
		}
	}
	if servers, ok := fields["server"].([]any); ok {
		expected["server.#"] = strconv.Itoa(len(servers))
		for i, value := range servers {
			expected[fmt.Sprintf("server.%d", i)] = fmt.Sprint(value)
		}
	}
	return expected, nil
}
