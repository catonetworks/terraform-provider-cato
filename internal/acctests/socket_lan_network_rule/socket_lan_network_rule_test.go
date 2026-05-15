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
		},
	})
}

type socketLanNetworkRuleCfg struct {
	resName string
	t       *testing.T
}

func newSocketLanNetworkRuleCfg(t *testing.T) socketLanNetworkRuleCfg {
	return socketLanNetworkRuleCfg{
		resName: acc.GetRandName("socket_lan_network_rule"),
		t:       t,
	}
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

func (p socketLanNetworkRuleCfg) getTfConfigSimple(index int) string {
	data := map[string]any{
		"Name": p.resName,
	}
	return p.prepareTfCfg(data, socketLanNetworkRuleSimpleTFs[index])
}

var socketLanNetworkRuleSimpleTFs = []string{
	`resource "cato_socket_site" "this" {
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
}
