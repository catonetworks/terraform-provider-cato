//go:build acctest

package socket_lan_firewall_rule

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccSocketLanFirewallRule_Simple(t *testing.T) {
	mockSrv := accmock.NewMockServer(t, "TestAccSocketLanFirewallRule_Simple")
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newSocketLanFirewallRuleCfg(t)
	res := "cato_socket_lan_firewall_rule.simple"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				// Create the resource
				Config:             cfg.getTfConfigSimple(0),
				Check:              socketLanFirewallRuleChecks(res, cfg, cfg.resName, "ALLOW", "true", "DAILY"),
				ExpectNonEmptyPlan: true, // TODO: provider read drops at.ref from state.
			},
			{
				// Update the resource
				Config:             cfg.getTfConfigSimple(1),
				Check:              socketLanFirewallRuleChecks(res, cfg, cfg.resName+"-2", "BLOCK", "false", "HOURLY"),
				ExpectNonEmptyPlan: true, // TODO: provider read drops at.ref from state.
			},
		},
	})
}

type socketLanFirewallRuleCfg struct {
	resName            string
	hosts              []acc.Ref
	globalIPRanges     []acc.Ref
	siteRanges         []acc.Ref
	floatingRanges     []acc.Ref
	interfaces         []acc.Ref
	groups             []acc.Ref
	systemGroups       []acc.Ref
	customApps         []acc.Ref
	subscriptionGroups []acc.Ref
	webhooks           []acc.Ref
	mailingLists       []acc.Ref
	t                  *testing.T
}

func newSocketLanFirewallRuleCfg(t *testing.T) socketLanFirewallRuleCfg {
	return socketLanFirewallRuleCfg{
		resName:            acc.GetRandName("socket_lan_firewall_rule"),
		hosts:              acc.GetHosts(t),
		globalIPRanges:     acc.GetGlobalIPRanges(t),
		siteRanges:         acc.GetSiteRanges(t),
		floatingRanges:     acc.GetFloatingRanges(t),
		interfaces:         acc.GetInterfaces(t),
		groups:             acc.GetAdvancedGroups(t),
		systemGroups:       acc.GetSystemGroups(t),
		customApps:         acc.GetCustomApps(t),
		subscriptionGroups: acc.GetSubscriptionGroups(t),
		webhooks:           acc.GetWebhooks(t),
		mailingLists:       acc.GetMailingLists(t),
		t:                  t,
	}
}

func (p socketLanFirewallRuleCfg) prepareTfCfg(data map[string]any, tmplText string) string {
	tmpl, err := template.New("tmpl").Parse(tmplText)
	if err != nil {
		p.t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		p.t.Fatal(err)
	}
	cfg := acc.ProviderCfg() + buf.String()
	fmt.Println(cfg)
	return cfg
}

func (p socketLanFirewallRuleCfg) getTfConfigSimple(index int) string {
	data := map[string]any{
		"Name":               p.resName,
		"Hosts":              p.hosts,
		"GlobalIPRanges":     p.globalIPRanges,
		"SiteRanges":         p.siteRanges,
		"FloatingRanges":     p.floatingRanges,
		"Interfaces":         p.interfaces,
		"Groups":             p.groups,
		"SystemGroups":       p.systemGroups,
		"CustomApps":         p.customApps,
		"SubscriptionGroups": p.subscriptionGroups,
		"Webhooks":           p.webhooks,
		"MailingLists":       p.mailingLists,
	}
	return p.prepareTfCfg(data, socketLanFirewallRuleSimpleTFs[index])
}

func socketLanFirewallRuleChecks(res string, cfg socketLanFirewallRuleCfg, name, action, enabled, frequency string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc(
		acc.PrintAttributes(res),
		resource.TestCheckResourceAttr(res, "at.position", "LAST_IN_RULE"),
		resource.TestCheckResourceAttr(res, "rule.action", action),
		resource.TestCheckResourceAttr(res, "rule.application.custom_app.#", "1"),
		resource.TestCheckTypeSetElemNestedAttrs(res, "rule.application.custom_app.*",
			map[string]string{"id": cfg.customApps[0].ID},
		),
		resource.TestCheckResourceAttr(res, "rule.application.domain.#", "1"),
		resource.TestCheckResourceAttr(res, "rule.application.domain.0", "app.example.com"),
		resource.TestCheckResourceAttr(res, "rule.application.fqdn.#", "1"),
		resource.TestCheckResourceAttr(res, "rule.application.fqdn.0", "www.app.example.com"),
		resource.TestCheckResourceAttr(res, "rule.application.global_ip_range.#", "1"),
		resource.TestCheckTypeSetElemNestedAttrs(res, "rule.application.global_ip_range.*",
			map[string]string{"id": cfg.globalIPRanges[0].ID},
		),
		resource.TestCheckResourceAttr(res, "rule.application.ip.#", "1"),
		resource.TestCheckResourceAttr(res, "rule.application.ip.0", "198.51.100.10"),
		resource.TestCheckResourceAttr(res, "rule.application.ip_range.#", "1"),
		resource.TestCheckResourceAttr(res, "rule.application.ip_range.0.from", "198.51.100.20"),
		resource.TestCheckResourceAttr(res, "rule.application.ip_range.0.to", "198.51.100.30"),
		resource.TestCheckResourceAttr(res, "rule.application.subnet.#", "1"),
		resource.TestCheckResourceAttr(res, "rule.application.subnet.0", "198.51.100.0/24"),
		resource.TestCheckResourceAttr(res, "rule.description", name+" description"),
		resource.TestCheckResourceAttr(res, "rule.destination.floating_subnet.#", "1"),
		resource.TestCheckTypeSetElemNestedAttrs(res, "rule.destination.floating_subnet.*",
			map[string]string{"id": cfg.floatingRanges[0].ID},
		),
		resource.TestCheckResourceAttr(res, "rule.destination.global_ip_range.#", "1"),
		resource.TestCheckTypeSetElemNestedAttrs(res, "rule.destination.global_ip_range.*",
			map[string]string{"id": cfg.globalIPRanges[0].ID},
		),
		resource.TestCheckResourceAttr(res, "rule.destination.group.#", "1"),
		resource.TestCheckTypeSetElemNestedAttrs(res, "rule.destination.group.*",
			map[string]string{"id": cfg.groups[0].ID},
		),
		resource.TestCheckResourceAttr(res, "rule.destination.host.#", "1"),
		resource.TestCheckTypeSetElemNestedAttrs(res, "rule.destination.host.*",
			map[string]string{"id": cfg.hosts[0].ID},
		),
		resource.TestCheckResourceAttr(res, "rule.destination.ip.#", "1"),
		resource.TestCheckResourceAttr(res, "rule.destination.ip.0", "192.0.2.10"),
		resource.TestCheckResourceAttr(res, "rule.destination.ip_range.#", "1"),
		resource.TestCheckResourceAttr(res, "rule.destination.ip_range.0.from", "192.0.2.20"),
		resource.TestCheckResourceAttr(res, "rule.destination.ip_range.0.to", "192.0.2.30"),
		resource.TestCheckResourceAttr(res, "rule.destination.site.#", "1"),
		resource.TestCheckTypeSetElemNestedAttrs(res, "rule.destination.site.*",
			map[string]string{"name": cfg.resName},
		),
		resource.TestCheckResourceAttr(res, "rule.destination.site_network_subnet.#", "1"),
		resource.TestCheckTypeSetElemNestedAttrs(res, "rule.destination.site_network_subnet.*",
			map[string]string{"id": cfg.siteRanges[0].ID},
		),
		resource.TestCheckResourceAttr(res, "rule.destination.subnet.#", "1"),
		resource.TestCheckResourceAttr(res, "rule.destination.subnet.0", "192.0.2.0/24"),
		resource.TestCheckResourceAttr(res, "rule.destination.system_group.#", "1"),
		resource.TestCheckTypeSetElemNestedAttrs(res, "rule.destination.system_group.*",
			map[string]string{"id": cfg.systemGroups[0].ID},
		),
		resource.TestCheckResourceAttr(res, "rule.destination.vlan.#", "1"),
		resource.TestCheckResourceAttr(res, "rule.destination.vlan.0", "20"),
		resource.TestCheckResourceAttr(res, "rule.direction", "TO"),
		resource.TestCheckResourceAttr(res, "rule.enabled", enabled),
		resource.TestCheckResourceAttrSet(res, "rule.id"),
		resource.TestCheckResourceAttr(res, "rule.name", name),
		resource.TestCheckResourceAttr(res, "rule.service.custom.#", "2"),
		resource.TestCheckResourceAttr(res, "rule.service.custom.0.port.#", "1"),
		resource.TestCheckResourceAttr(res, "rule.service.custom.0.port.0", "8022"),
		resource.TestCheckResourceAttr(res, "rule.service.custom.0.protocol", "TCP"),
		resource.TestCheckResourceAttr(res, "rule.service.custom.1.port_range.from", "6000"),
		resource.TestCheckResourceAttr(res, "rule.service.custom.1.port_range.to", "6010"),
		resource.TestCheckResourceAttr(res, "rule.service.custom.1.protocol", "TCP"),
		resource.TestCheckResourceAttr(res, "rule.service.simple.#", "1"),
		resource.TestCheckTypeSetElemNestedAttrs(res, "rule.service.simple.*",
			map[string]string{"name": "SSH"},
		),
		resource.TestCheckResourceAttr(res, "rule.source.floating_subnet.#", "1"),
		resource.TestCheckTypeSetElemNestedAttrs(res, "rule.source.floating_subnet.*",
			map[string]string{"id": cfg.floatingRanges[0].ID},
		),
		resource.TestCheckResourceAttr(res, "rule.source.global_ip_range.#", "1"),
		resource.TestCheckTypeSetElemNestedAttrs(res, "rule.source.global_ip_range.*",
			map[string]string{"id": cfg.globalIPRanges[0].ID},
		),
		resource.TestCheckResourceAttr(res, "rule.source.group.#", "1"),
		resource.TestCheckTypeSetElemNestedAttrs(res, "rule.source.group.*",
			map[string]string{"id": cfg.groups[0].ID},
		),
		resource.TestCheckResourceAttr(res, "rule.source.host.#", "1"),
		resource.TestCheckTypeSetElemNestedAttrs(res, "rule.source.host.*",
			map[string]string{"id": cfg.hosts[0].ID},
		),
		resource.TestCheckResourceAttr(res, "rule.source.ip.#", "1"),
		resource.TestCheckResourceAttr(res, "rule.source.ip.0", "10.99.12.31"),
		resource.TestCheckResourceAttr(res, "rule.source.ip_range.#", "1"),
		resource.TestCheckResourceAttr(res, "rule.source.ip_range.0.from", "10.99.12.10"),
		resource.TestCheckResourceAttr(res, "rule.source.ip_range.0.to", "10.99.12.20"),
		resource.TestCheckResourceAttr(res, "rule.source.mac.#", "1"),
		resource.TestCheckResourceAttr(res, "rule.source.mac.0", "00:11:22:33:44:55"),
		resource.TestCheckResourceAttr(res, "rule.source.site.#", "1"),
		resource.TestCheckTypeSetElemNestedAttrs(res, "rule.source.site.*",
			map[string]string{"name": cfg.resName},
		),
		resource.TestCheckResourceAttr(res, "rule.source.site_network_subnet.#", "1"),
		resource.TestCheckTypeSetElemNestedAttrs(res, "rule.source.site_network_subnet.*",
			map[string]string{"id": cfg.siteRanges[0].ID},
		),
		resource.TestCheckResourceAttr(res, "rule.source.subnet.#", "1"),
		resource.TestCheckResourceAttr(res, "rule.source.subnet.0", "10.99.12.0/24"),
		resource.TestCheckResourceAttr(res, "rule.source.system_group.#", "1"),
		resource.TestCheckTypeSetElemNestedAttrs(res, "rule.source.system_group.*",
			map[string]string{"id": cfg.systemGroups[0].ID},
		),
		resource.TestCheckResourceAttr(res, "rule.source.vlan.#", "1"),
		resource.TestCheckResourceAttr(res, "rule.source.vlan.0", "10"),
		resource.TestCheckResourceAttr(res, "rule.tracking.alert.enabled", "true"),
		resource.TestCheckResourceAttr(res, "rule.tracking.alert.frequency", frequency),
		resource.TestCheckResourceAttr(res, "rule.tracking.alert.mailing_list.#", "1"),
		resource.TestCheckTypeSetElemNestedAttrs(res, "rule.tracking.alert.mailing_list.*",
			map[string]string{"id": cfg.mailingLists[0].ID},
		),
		resource.TestCheckResourceAttr(res, "rule.tracking.alert.subscription_group.#", "1"),
		resource.TestCheckTypeSetElemNestedAttrs(res, "rule.tracking.alert.subscription_group.*",
			map[string]string{"id": cfg.subscriptionGroups[0].ID},
		),
		resource.TestCheckResourceAttr(res, "rule.tracking.alert.webhook.#", "1"),
		resource.TestCheckTypeSetElemNestedAttrs(res, "rule.tracking.alert.webhook.*",
			map[string]string{"id": cfg.webhooks[0].ID},
		),
		resource.TestCheckResourceAttr(res, "rule.tracking.event.enabled", "true"),
	)
}

var socketLanFirewallRuleSimpleTFs = []string{
	siteResource + `
	resource "cato_socket_lan_firewall_rule" "simple" {
		at = {
			position = "LAST_IN_RULE"
			ref      = cato_socket_lan_network_rule.parent.rule.id
		}
		rule = {
			name        = "{{ .Name }}"
			description = "{{ .Name }} description"
			enabled     = true
			direction   = "TO"
			action      = "ALLOW"
			source = {
				vlan   = [10]
				mac    = ["00:11:22:33:44:55"]
				ip     = ["10.99.12.31"]
				subnet = ["10.99.12.0/24"]
				ip_range = [
					{ from = "10.99.12.10", to = "10.99.12.20" },
				]
				host = [
					{ id = "{{ (index .Hosts 0).ID }}" },
				]
				site = [
					{ id = cato_socket_site.this.id },
				]
				group = [
					{ id = "{{ (index .Groups 0).ID }}" },
				]
				system_group = [
					{ id = "{{ (index .SystemGroups 0).ID }}" },
				]
				global_ip_range = [
					{ id = "{{ (index .GlobalIPRanges 0).ID }}" },
				]
				floating_subnet = [
					{ id = "{{ (index .FloatingRanges 0).ID }}" },
				]
				site_network_subnet = [
					{ id = "{{ (index .SiteRanges 0).ID }}" },
				]
			}
			destination = {
				vlan   = [20]
				ip     = ["192.0.2.10"]
				subnet = ["192.0.2.0/24"]
				ip_range = [
					{ from = "192.0.2.20", to = "192.0.2.30" },
				]
				host = [
					{ id = "{{ (index .Hosts 0).ID }}" },
				]
				site = [
					{ id = cato_socket_site.this.id },
				]
				group = [
					{ id = "{{ (index .Groups 0).ID }}" },
				]
				system_group = [
					{ id = "{{ (index .SystemGroups 0).ID }}" },
				]
				global_ip_range = [
					{ id = "{{ (index .GlobalIPRanges 0).ID }}" },
				]
				floating_subnet = [
					{ id = "{{ (index .FloatingRanges 0).ID }}" },
				]
				site_network_subnet = [
					{ id = "{{ (index .SiteRanges 0).ID }}" },
				]
			}
			application = {
				custom_app = [
					{ id = "{{ (index .CustomApps 0).ID }}" },
				]
				domain = ["app.example.com"]
				fqdn   = ["www.app.example.com"]
				ip     = ["198.51.100.10"]
				subnet = ["198.51.100.0/24"]
				ip_range = [
					{ from = "198.51.100.20", to = "198.51.100.30" },
				]
				global_ip_range = [
					{ id = "{{ (index .GlobalIPRanges 0).ID }}" },
				]
			}
			service = {
				simple = [
					{ name = "SSH" },
				]
				custom = [
					{ port = ["8022"], protocol = "TCP" },
					{ port_range = { from = "6000", to = "6010" }, protocol = "TCP" },
				]
			}
			tracking = {
				event = {
					enabled = true
				}
				alert = {
					enabled   = true
					frequency = "DAILY"
					subscription_group = [
						{ id = "{{ (index .SubscriptionGroups 0).ID }}" },
					]
					webhook = [
						{ id = "{{ (index .Webhooks 0).ID }}" },
					]
					mailing_list = [
						{ id = "{{ (index .MailingLists 0).ID }}" },
					]
				}
			}
		}
	}
	`,
	siteResource + `
	resource "cato_socket_lan_firewall_rule" "simple" {
		at = {
			position = "LAST_IN_RULE"
			ref      = cato_socket_lan_network_rule.parent.rule.id
		}
		rule = {
			name        = "{{ .Name }}-2"
			description = "{{ .Name }}-2 description"
			enabled     = false
			direction   = "TO"
			action      = "BLOCK"
			source = {
				vlan   = [10]
				mac    = ["00:11:22:33:44:55"]
				ip     = ["10.99.12.31"]
				subnet = ["10.99.12.0/24"]
				ip_range = [
					{ from = "10.99.12.10", to = "10.99.12.20" },
				]
				host = [
					{ id = "{{ (index .Hosts 0).ID }}" },
				]
				site = [
					{ id = cato_socket_site.this.id },
				]
				group = [
					{ id = "{{ (index .Groups 0).ID }}" },
				]
				system_group = [
					{ id = "{{ (index .SystemGroups 0).ID }}" },
				]
				global_ip_range = [
					{ id = "{{ (index .GlobalIPRanges 0).ID }}" },
				]
				floating_subnet = [
					{ id = "{{ (index .FloatingRanges 0).ID }}" },
				]
				site_network_subnet = [
					{ id = "{{ (index .SiteRanges 0).ID }}" },
				]
			}
			destination = {
				vlan   = [20]
				ip     = ["192.0.2.10"]
				subnet = ["192.0.2.0/24"]
				ip_range = [
					{ from = "192.0.2.20", to = "192.0.2.30" },
				]
				host = [
					{ id = "{{ (index .Hosts 0).ID }}" },
				]
				site = [
					{ id = cato_socket_site.this.id },
				]
				group = [
					{ id = "{{ (index .Groups 0).ID }}" },
				]
				system_group = [
					{ id = "{{ (index .SystemGroups 0).ID }}" },
				]
				global_ip_range = [
					{ id = "{{ (index .GlobalIPRanges 0).ID }}" },
				]
				floating_subnet = [
					{ id = "{{ (index .FloatingRanges 0).ID }}" },
				]
				site_network_subnet = [
					{ id = "{{ (index .SiteRanges 0).ID }}" },
				]
			}
			application = {
				custom_app = [
					{ id = "{{ (index .CustomApps 0).ID }}" },
				]
				domain = ["app.example.com"]
				fqdn   = ["www.app.example.com"]
				ip     = ["198.51.100.10"]
				subnet = ["198.51.100.0/24"]
				ip_range = [
					{ from = "198.51.100.20", to = "198.51.100.30" },
				]
				global_ip_range = [
					{ id = "{{ (index .GlobalIPRanges 0).ID }}" },
				]
			}
			service = {
				simple = [
					{ name = "SSH" },
				]
				custom = [
					{ port = ["8022"], protocol = "TCP" },
					{ port_range = { from = "6000", to = "6010" }, protocol = "TCP" },
				]
			}
			tracking = {
				event = {
					enabled = true
				}
				alert = {
					enabled   = true
					frequency = "HOURLY"
					subscription_group = [
						{ id = "{{ (index .SubscriptionGroups 0).ID }}" },
					]
					webhook = [
						{ id = "{{ (index .Webhooks 0).ID }}" },
					]
					mailing_list = [
						{ id = "{{ (index .MailingLists 0).ID }}" },
					]
				}
			}
		}
	}
	`,
}

const siteResource = `
	resource "cato_socket_site" "this" {
		name            = "{{ .Name }}"
		description     = "{{ .Name }} description"
		site_type       = "BRANCH"
		connection_type = "SOCKET_X1500"

		native_range = {
			native_network_range = "192.168.246.0/24"
			local_ip             = "192.168.246.1"
			dhcp_settings = {
				dhcp_type = "DHCP_RANGE"
				ip_range  = "192.168.246.10-192.168.246.22"
			}
		}

		site_location = {
			country_code = "FR"
			timezone     = "Europe/Paris"
		}
	}

	resource "cato_socket_lan_network_rule" "parent" {
		at = {
			position = "LAST_IN_POLICY"
		}
		rule = {
			name      = "{{ .Name }}_parent"
			enabled   = true
			direction = "TO"
			transport = "LAN"
			site = {
				site = [
					{
						id = cato_socket_site.this.id
					}
				]
			}
			source      = {}
			destination = {}
			nat = {
				enabled  = false
				nat_type = "DYNAMIC_PAT"
			}
		}
	}
`
