package dnsdhcptest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"github.com/catonetworks/terraform-provider-cato/internal/provider"
)

// Runs real Terraform plan/apply/import against an HTTP fixture and local provider source.
func TestDNSDHCPTerraformLifecycle(t *testing.T) {
	t.Setenv("DISABLE_POLICY_RULE_CLEANUP", "true")
	t.Setenv("CATO_VERSION_CHECK_DISABLED", "true")
	backend := newBackend()
	server := httptest.NewServer(backend)
	defer server.Close()
	config := func(updated bool) string { return fixtureConfig(server.URL, updated) }
	steps := []resource.TestStep{
		{Config: config(false)},
		{Config: config(false), PlanOnly: true},
		{Config: config(false), PlanOnly: true, ExpectNonEmptyPlan: true, PreConfig: func() {
			backend.mu.Lock()
			defer backend.mu.Unlock()
			backend.objects["dhcp.option"]["value"] = "out-of-band"
		}},
		{Config: config(false)},
	}
	for _, name := range []string{"dns_settings", "dns_site_settings", "dns_forwarding_rule", "dhcp_settings", "dhcp_option", "dhcp_relay_group"} {
		steps = append(steps, resource.TestStep{ResourceName: "cato_" + name + ".test", ImportState: true, ImportStateVerify: true})
	}
	steps = append(steps, resource.TestStep{Config: config(true)}, resource.TestStep{Config: config(true), PlanOnly: true})
	emptyDescriptionConfig := strings.Replace(config(true), "description = null", `description = ""`, 1)
	steps = append(steps, resource.TestStep{Config: emptyDescriptionConfig, PlanOnly: true}, resource.TestStep{Config: config(true)})
	nameConfig := strings.Replace(config(true), `site = { id = "site-123" }`, `site = { name = "branch" }`, 1)
	steps = append(steps, resource.TestStep{Config: nameConfig, PlanOnly: true})
	relayConfig := strings.Replace(nameConfig, `relay = { enabled = false }`, `relay = { enabled = true, timeout = 10, group = { name = cato_dhcp_relay_group.test.name } }`, 1)
	steps = append(steps, resource.TestStep{Config: relayConfig}, resource.TestStep{Config: relayConfig, PlanOnly: true}, resource.TestStep{Config: nameConfig}, resource.TestStep{Config: nameConfig, PlanOnly: true})
	preserveConfig := strings.Replace(nameConfig, " lease_time = 3600\n relay = { enabled = false }", "", 1)
	steps = append(steps, resource.TestStep{Config: preserveConfig, PlanOnly: true}, resource.TestStep{Config: preserveConfig})
	clearDNSConfig := strings.ReplaceAll(preserveConfig, ` primary_server = "192.0.2.53"`, "")
	steps = append(steps, resource.TestStep{Config: clearDNSConfig}, resource.TestStep{Config: clearDNSConfig, PlanOnly: true})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){"cato": providerserver.NewProtocol6WithError(provider.New("test")())},
		Steps:                    steps,
	})
	backend.mu.Lock()
	defer backend.mu.Unlock()
	require.NotNil(t, backend.objects["dns.settings"], "singleton retained")
	require.NotNil(t, backend.objects["dhcp.settings"], "singleton retained")
	for _, key := range []string{"dns.siteSettings", "dns.forwardingRule", "dhcp.option", "dhcp.relayGroup"} {
		require.Nil(t, backend.objects[key], key)
	}
}

func fixtureConfig(endpoint string, updated bool) string {
	domain, value, secondary, suffix, description := "terraform.test", "first", `"192.0.2.54"`, `["terraform.test"]`, `"description"`
	if updated {
		domain, value, secondary, suffix, description = "updated.test", "second", "null", "[]", "null"
	}
	return fmt.Sprintf(`
provider "cato" {
 baseurl = %q
 account_id = "account-123"
 token = "test-key"
 retry_max = 0
}
resource "cato_dns_settings" "test" {
 primary_server = "192.0.2.53"
 secondary_server = %s
 suffix = %s
 accept_dns_requests_on_lan_interface_ip = false
}
resource "cato_dns_site_settings" "test" {
 site = { id = "site-123" }
 primary_server = "192.0.2.53"
 secondary_server = %s
 suffix = %s
}
resource "cato_dns_forwarding_rule" "test" {
 domain = %q
 server = ["192.0.2.53"]
}
resource "cato_dhcp_settings" "test" {
 lease_time = 3600
 relay = { enabled = false }
}
resource "cato_dhcp_option" "test" {
 tag = 224
 type = "ASCII"
 value = %q
 description = %s
}
resource "cato_dhcp_relay_group" "test" {
 name = "terraform-test"
 server = ["192.0.2.53"]
}
`, endpoint, secondary, suffix, secondary, suffix, domain, value, description)
}

type backend struct {
	mu      sync.Mutex
	objects map[string]map[string]any
}

func newBackend() *backend {
	return &backend{objects: map[string]map[string]any{
		"dns.settings":  {"primaryServer": nil, "secondaryServer": nil, "suffix": []any{}, "acceptDnsRequestsOnLanInterfaceIp": false},
		"dhcp.settings": {"leaseTime": float64(3600), "relay": map[string]any{"enabled": false, "timeout": float64(0), "group": nil}},
	}}
}

func (b *backend) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b.mu.Lock()
	defer b.mu.Unlock()
	var request struct {
		OperationName string         `json:"operationName"`
		Variables     map[string]any `json:"variables"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if request.OperationName == "siteGeneralDetails" {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"site": map[string]any{"siteGeneralDetails": map[string]any{"site": map[string]any{"id": "site-123", "name": "branch"}}}}}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	domain := "dns"
	if strings.Contains(strings.ToLower(request.OperationName), "dhcp") {
		domain = "dhcp"
	}
	position := strings.Index(strings.ToLower(request.OperationName), domain) + len(domain)
	operation := request.OperationName[position:]
	operation = strings.ToLower(operation[:1]) + operation[1:]
	var input map[string]any
	for key, value := range request.Variables {
		if strings.HasSuffix(key, "Input") {
			input, _ = value.(map[string]any)
		}
	}
	leaf := operation
	action := ""
	for _, candidate := range []string{"create", "update", "delete", "remove"} {
		if strings.HasPrefix(operation, candidate) {
			action = candidate
			leaf = operation[len(candidate):]
			leaf = strings.ToLower(leaf[:1]) + leaf[1:]
			break
		}
	}
	key := domain + "." + leaf
	var result any = b.objects[key]
	if action != "" {
		result = b.mutate(domain, leaf, action, input)
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"networkConfig": map[string]any{domain: map[string]any{operation: result}}}}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (b *backend) mutate(domain, leaf, action string, input map[string]any) any {
	key := domain + "." + leaf
	if action == "delete" || action == "remove" {
		old := b.objects[key]
		delete(b.objects, key)
		return map[string]any{leaf: []any{old}}
	}
	if leaf == "settings" {
		current := b.objects[key]
		for k, v := range input {
			if k == "relay" {
				for rk, rv := range v.(map[string]any) {
					if rk == "group" {
						group := b.objects["dhcp.relayGroup"]
						rv = map[string]any{"id": group["id"], "name": group["name"]}
					}
					current[k].(map[string]any)[rk] = rv
				}
			} else {
				current[k] = v
			}
		}
		return map[string]any{leaf: current}
	}
	item := input[leaf].([]any)[0].(map[string]any)
	if leaf == "siteSettings" {
		item["siteId"] = item["site"].(map[string]any)["input"]
		delete(item, "site")
	} else {
		item["id"] = "opaque-" + leaf
	}
	b.objects[key] = item
	return map[string]any{leaf: []any{item}}
}
