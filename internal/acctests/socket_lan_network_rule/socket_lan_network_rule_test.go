//go:build acctest

package socket_lan_network_rule

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccSocketLanNetworkRule_Simple(t *testing.T) {
	acc.SkipByEnv(t)
	mockSrv := accmock.NewMockServer(t, "TestAccSocketLanNetworkRule_Simple")
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newSocketLanNetworkRuleCfg(t)
	res := "cato_socket_lan_network_rule.simple"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				// Create the resource
				Config: cfg.getTfConfigSimple(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "at.position", "LAST_IN_POLICY"),
					resource.TestCheckResourceAttr(res, "rule.direction", "TO"),
					resource.TestCheckResourceAttr(res, "rule.enabled", "true"),
					resource.TestCheckResourceAttrSet(res, "rule.id"),
					resource.TestCheckResourceAttr(res, "rule.name", cfg.resName),
					resource.TestCheckResourceAttr(res, "rule.nat.enabled", "false"),
					resource.TestCheckResourceAttr(res, "rule.nat.nat_type", "DYNAMIC_PAT"),
					resource.TestCheckResourceAttr(res, "rule.site.site.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.site.site.*",
						map[string]string{"name": cfg.resName},
					),
					resource.TestCheckResourceAttr(res, "rule.transport", "LAN"),
				),
			},
			{
				// Update the resource
				Config: cfg.getTfConfigSimple(1),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "at.position", "LAST_IN_POLICY"),
					resource.TestCheckResourceAttr(res, "rule.direction", "BOTH"),
					resource.TestCheckResourceAttr(res, "rule.enabled", "false"),
					resource.TestCheckResourceAttrSet(res, "rule.id"),
					resource.TestCheckResourceAttr(res, "rule.name", cfg.resName+"-2"),
					resource.TestCheckResourceAttr(res, "rule.nat.enabled", "false"),
					resource.TestCheckResourceAttr(res, "rule.nat.nat_type", "DYNAMIC_PAT"),
					resource.TestCheckResourceAttr(res, "rule.site.site.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.site.site.*",
						map[string]string{"name": cfg.resName},
					),
					resource.TestCheckResourceAttr(res, "rule.transport", "LAN"),
				),
			},
		},
	})
}

func TestAccSocketLanNetworkRule_Full(t *testing.T) {
	acc.SkipByEnv(t)
	mockSrv := accmock.NewMockServer(t, "TestAccSocketLanNetworkRule_Full")
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newSocketLanNetworkRuleFullCfg(t)
	res := "cato_socket_lan_network_rule.full"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				// Create the resource
				Config: cfg.getTfConfigFull(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "at.position", "LAST_IN_POLICY"),
					resource.TestCheckResourceAttr(res, "rule.description", cfg.resName+" description"),
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
					resource.TestCheckResourceAttr(res, "rule.destination.ip.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.destination.ip.0", "192.0.2.10"),
					resource.TestCheckResourceAttr(res, "rule.destination.ip_range.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.destination.ip_range.0.from", "192.0.2.20"),
					resource.TestCheckResourceAttr(res, "rule.destination.ip_range.0.to", "192.0.2.30"),
					resource.TestCheckResourceAttr(res, "rule.destination.subnet.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.destination.subnet.0", "192.0.2.0/24"),
					resource.TestCheckResourceAttr(res, "rule.destination.system_group.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.destination.system_group.*",
						map[string]string{"id": cfg.systemGroups[0].ID},
					),
					resource.TestCheckResourceAttr(res, "rule.destination.vlan.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.destination.vlan.0", "20"),
					resource.TestCheckResourceAttr(res, "rule.direction", "TO"),
					resource.TestCheckResourceAttr(res, "rule.enabled", "true"),
					resource.TestCheckResourceAttrSet(res, "rule.id"),
					resource.TestCheckResourceAttr(res, "rule.name", cfg.resName),
					resource.TestCheckResourceAttr(res, "rule.nat.enabled", "true"),
					resource.TestCheckResourceAttr(res, "rule.nat.nat_type", "DYNAMIC_PAT"),
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
					resource.TestCheckResourceAttr(res, "rule.site.group.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.site.group.*",
						map[string]string{"id": cfg.groups[0].ID},
					),
					resource.TestCheckResourceAttr(res, "rule.site.site.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.site.site.*",
						map[string]string{"name": cfg.resName},
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
					resource.TestCheckResourceAttr(res, "rule.source.ip.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.source.ip.0", "10.99.12.31"),
					resource.TestCheckResourceAttr(res, "rule.source.ip_range.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.source.ip_range.0.from", "10.99.12.10"),
					resource.TestCheckResourceAttr(res, "rule.source.ip_range.0.to", "10.99.12.20"),
					resource.TestCheckResourceAttr(res, "rule.source.subnet.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.source.subnet.0", "10.99.12.0/24"),
					resource.TestCheckResourceAttr(res, "rule.source.system_group.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.source.system_group.*",
						map[string]string{"id": cfg.systemGroups[0].ID},
					),
					resource.TestCheckResourceAttr(res, "rule.source.vlan.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.source.vlan.0", "10"),
					resource.TestCheckResourceAttr(res, "rule.transport", "LAN"),
				),
			},
			{
				// Update the resource
				Config: cfg.getTfConfigFull(1),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "at.position", "LAST_IN_POLICY"),
					resource.TestCheckResourceAttr(res, "rule.description", cfg.resName+" description 2"),
					resource.TestCheckResourceAttr(res, "rule.destination.floating_subnet.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.destination.floating_subnet.*",
						map[string]string{"id": cfg.floatingRanges[1].ID},
					),
					resource.TestCheckResourceAttr(res, "rule.destination.global_ip_range.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.destination.global_ip_range.*",
						map[string]string{"id": cfg.globalIPRanges[1].ID},
					),
					resource.TestCheckResourceAttr(res, "rule.destination.group.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.destination.group.*",
						map[string]string{"id": cfg.groups[1].ID},
					),
					resource.TestCheckResourceAttr(res, "rule.destination.ip.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.destination.ip.0", "192.0.3.10"),
					resource.TestCheckResourceAttr(res, "rule.destination.ip_range.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.destination.ip_range.0.from", "192.0.3.20"),
					resource.TestCheckResourceAttr(res, "rule.destination.ip_range.0.to", "192.0.3.30"),
					resource.TestCheckResourceAttr(res, "rule.destination.subnet.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.destination.subnet.0", "192.0.3.0/24"),
					resource.TestCheckResourceAttr(res, "rule.destination.system_group.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.destination.system_group.*",
						map[string]string{"id": cfg.systemGroups[1].ID},
					),
					resource.TestCheckResourceAttr(res, "rule.destination.vlan.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.destination.vlan.0", "21"),
					resource.TestCheckResourceAttr(res, "rule.direction", "TO"),
					resource.TestCheckResourceAttr(res, "rule.enabled", "true"),
					resource.TestCheckResourceAttrSet(res, "rule.id"),
					resource.TestCheckResourceAttr(res, "rule.name", cfg.resName),
					resource.TestCheckResourceAttr(res, "rule.nat.enabled", "true"),
					resource.TestCheckResourceAttr(res, "rule.nat.nat_type", "DYNAMIC_PAT"),
					resource.TestCheckResourceAttr(res, "rule.service.custom.#", "2"),
					resource.TestCheckResourceAttr(res, "rule.service.custom.0.port.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.service.custom.0.port.0", "8023"),
					resource.TestCheckResourceAttr(res, "rule.service.custom.0.protocol", "TCP"),
					resource.TestCheckResourceAttr(res, "rule.service.custom.1.port_range.from", "6001"),
					resource.TestCheckResourceAttr(res, "rule.service.custom.1.port_range.to", "6011"),
					resource.TestCheckResourceAttr(res, "rule.service.custom.1.protocol", "TCP"),
					resource.TestCheckResourceAttr(res, "rule.service.simple.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.service.simple.*",
						map[string]string{"name": "HTTPS"},
					),
					resource.TestCheckResourceAttr(res, "rule.site.group.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.site.group.*",
						map[string]string{"id": cfg.groups[1].ID},
					),
					resource.TestCheckResourceAttr(res, "rule.site.site.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.site.site.*",
						map[string]string{"name": cfg.resName},
					),
					resource.TestCheckResourceAttr(res, "rule.source.floating_subnet.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.source.floating_subnet.*",
						map[string]string{"id": cfg.floatingRanges[1].ID},
					),
					resource.TestCheckResourceAttr(res, "rule.source.global_ip_range.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.source.global_ip_range.*",
						map[string]string{"id": cfg.globalIPRanges[1].ID},
					),
					resource.TestCheckResourceAttr(res, "rule.source.group.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.source.group.*",
						map[string]string{"id": cfg.groups[1].ID},
					),
					resource.TestCheckResourceAttr(res, "rule.source.ip.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.source.ip.0", "10.99.13.32"),
					resource.TestCheckResourceAttr(res, "rule.source.ip_range.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.source.ip_range.0.from", "10.99.13.10"),
					resource.TestCheckResourceAttr(res, "rule.source.ip_range.0.to", "10.99.13.20"),
					resource.TestCheckResourceAttr(res, "rule.source.subnet.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.source.subnet.0", "10.99.13.0/24"),
					resource.TestCheckResourceAttr(res, "rule.source.system_group.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "rule.source.system_group.*",
						map[string]string{"id": cfg.systemGroups[1].ID},
					),
					resource.TestCheckResourceAttr(res, "rule.source.vlan.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.source.vlan.0", "11"),
					resource.TestCheckResourceAttr(res, "rule.transport", "LAN"),
				),
			},
		},
	})
}

type socketLanNetworkRuleCfg struct {
	resName        string
	hosts          []acc.Ref
	globalIPRanges []acc.Ref
	siteRanges     []acc.Ref
	floatingRanges []acc.Ref
	interfaces     []acc.Ref
	groups         []acc.Ref
	systemGroups   []acc.Ref
	t              *testing.T
}

func newSocketLanNetworkRuleCfg(t *testing.T) socketLanNetworkRuleCfg {
	return socketLanNetworkRuleCfg{
		resName: acc.GetRandName("socket_lan_network_rule"),
		t:       t,
	}
}

func newSocketLanNetworkRuleFullCfg(t *testing.T) socketLanNetworkRuleCfg {
	cfg := newSocketLanNetworkRuleCfg(t)
	cfg.hosts = acc.GetHosts(t)
	cfg.globalIPRanges = acc.GetGlobalIPRanges(t)
	cfg.siteRanges = acc.GetSiteRanges(t)
	cfg.floatingRanges = acc.GetFloatingRanges(t)
	cfg.interfaces = acc.GetInterfaces(t)
	cfg.groups = acc.GetAdvancedGroups(t)
	cfg.systemGroups = acc.GetSystemGroups(t)
	return cfg
}

func (p socketLanNetworkRuleCfg) prepareTfCfg(data map[string]any, tmplText string) string {
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

// ----------------------------------------------------------------------
// SIMPLE rule
// ----------------------------------------------------------------------
func (p socketLanNetworkRuleCfg) getTfConfigSimple(index int) string {
	data := map[string]any{
		"Name": p.resName,
	}
	return p.prepareTfCfg(data, socketLanNetworkRuleSimpleTFs[index])
}

var socketLanNetworkRuleSimpleTFs = []string{
	siteResource + `	
	resource "cato_socket_lan_network_rule" "simple" {
		at = {
			position = "LAST_IN_POLICY"
		}
		rule = {
			name      = "{{ .Name }}"
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
	`,
	siteResource + `	
	resource "cato_socket_lan_network_rule" "simple" {
		at = {
			position = "LAST_IN_POLICY"
		}
		rule = {
			name      = "{{ .Name }}-2"
			enabled   = false
			direction = "BOTH"
			transport = "LAN"
			site = {
				site = [
					{ id = cato_socket_site.this.id }
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
	`,
}

// ----------------------------------------------------------------------
// Full rule
// ----------------------------------------------------------------------
func (p socketLanNetworkRuleCfg) getTfConfigFull(index int) string {
	data := map[string]any{
		"Name":           p.resName,
		"Hosts":          p.hosts,
		"GlobalIPRanges": p.globalIPRanges,
		"SiteRanges":     p.siteRanges,
		"FloatingRanges": p.floatingRanges,
		"Interfaces":     p.interfaces,
		"Groups":         p.groups,
		"SystemGroups":   p.systemGroups,
	}
	return p.prepareTfCfg(data, socketLanNetworkRuleFullTFs[index])
}

var socketLanNetworkRuleFullTFs = []string{
	siteResource + `
	resource "cato_socket_lan_network_rule" "full" {
		at = {
			position = "LAST_IN_POLICY"
		}
		rule = {
			name        = "{{ .Name }}"
			description = "{{ .Name }} description"
			enabled     = true
			direction   = "TO"
			transport   = "LAN"
			site = {
				site = [
					{
						id = cato_socket_site.this.id
					}
				]
				group = [
					{ id = "{{ (index .Groups 0).ID }}" },
				]
			}
			source = {
				vlan   = [10]
				ip     = ["10.99.12.31"]
				subnet = ["10.99.12.0/24"]
				ip_range = [
					{ from = "10.99.12.10", to = "10.99.12.20" },
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
			}
			destination = {
				vlan   = [20]
				ip     = ["192.0.2.10"]
				subnet = ["192.0.2.0/24"]
				ip_range = [
					{ from = "192.0.2.20", to = "192.0.2.30" },
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
			nat = {
				enabled  = true
				nat_type = "DYNAMIC_PAT"
			}
		}
	}
	`,
	// Update
	siteResource + `
	resource "cato_socket_lan_network_rule" "full" {
		at = {
			position = "LAST_IN_POLICY"
		}
		rule = {
			name        = "{{ .Name }}"
			description = "{{ .Name }} description 2"
			enabled     = true
			direction   = "TO"
			transport   = "LAN"
			site = {
				site = [
					{
						id = cato_socket_site.this.id
					}
				]
				group = [
					{ name = "{{ (index .Groups 1).Name }}" },
				]
			}
			source = {
				vlan   = [11]
				ip     = ["10.99.13.32"]
				subnet = ["10.99.13.0/24"]
				ip_range = [
					{ from = "10.99.13.10", to = "10.99.13.20" },
				]
				group = [
					{ id = "{{ (index .Groups 1).ID }}" },
				]
				system_group = [
					{ id = "{{ (index .SystemGroups 1).ID }}" },
				]
				global_ip_range = [
					{ id = "{{ (index .GlobalIPRanges 1).ID }}" },
				]
				floating_subnet = [
					{ id = "{{ (index .FloatingRanges 1).ID }}" },
				]
			}
			destination = {
				vlan   = [21]
				ip     = ["192.0.3.10"]
				subnet = ["192.0.3.0/24"]
				ip_range = [
					{ from = "192.0.3.20", to = "192.0.3.30" },
				]
				group = [
					{ id = "{{ (index .Groups 1).ID }}" },
				]
				system_group = [
					{ id = "{{ (index .SystemGroups 1).ID }}" },
				]
				global_ip_range = [
					{ id = "{{ (index .GlobalIPRanges 1).ID }}" },
				]
				floating_subnet = [
					{ id = "{{ (index .FloatingRanges 1).ID }}" },
				]
			}
			service = {
				simple = [
					{ name = "HTTPS" },
				]
				custom = [
					{ port = ["8023"], protocol = "TCP" },
					{ port_range = { from = "6001", to = "6011" }, protocol = "TCP" },
				]
			}
			nat = {
				enabled  = true
				nat_type = "DYNAMIC_PAT"
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
			native_network_range = "192.168.247.0/24"
			local_ip             = "192.168.247.1"
			dhcp_settings = {
				dhcp_type = "DHCP_RANGE"
				ip_range  = "192.168.247.10-192.168.247.22"
			}
		}

		site_location = {
			country_code = "FR"
			timezone     = "Europe/Paris"
		}
	}
`

// TODO: add all attributes as soon as the API is fixed
