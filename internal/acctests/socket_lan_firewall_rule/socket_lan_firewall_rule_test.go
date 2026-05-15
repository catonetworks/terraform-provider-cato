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
				Config: cfg.getTfConfigSimple(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "at.position", "LAST_IN_RULE"),
					resource.TestCheckResourceAttr(res, "rule.action", "ALLOW"),
					resource.TestCheckResourceAttr(res, "rule.direction", "TO"),
					resource.TestCheckResourceAttr(res, "rule.enabled", "true"),
					resource.TestCheckResourceAttrSet(res, "rule.id"),
					resource.TestCheckResourceAttr(res, "rule.name", cfg.resName),
					resource.TestCheckResourceAttr(res, "rule.tracking.alert.enabled", "false"),
					resource.TestCheckResourceAttr(res, "rule.tracking.alert.frequency", "DAILY"),
					resource.TestCheckResourceAttr(res, "rule.tracking.event.enabled", "true"),
				),
				ExpectNonEmptyPlan: true, // TODO: provider read drops at.ref from state.
			},
		},
	})
}

type socketLanFirewallRuleCfg struct {
	resName string
	t       *testing.T
}

func newSocketLanFirewallRuleCfg(t *testing.T) socketLanFirewallRuleCfg {
	return socketLanFirewallRuleCfg{
		resName: acc.GetRandName("socket_lan_firewall_rule"),
		t:       t,
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
		"Name": p.resName,
	}
	return p.prepareTfCfg(data, socketLanFirewallRuleSimpleTFs[index])
}

var socketLanFirewallRuleSimpleTFs = []string{
	`resource "cato_socket_site" "this" {
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

	resource "cato_socket_lan_firewall_rule" "simple" {
		at = {
			position = "LAST_IN_RULE"
			ref      = cato_socket_lan_network_rule.parent.rule.id
		}
		rule = {
			name        = "{{ .Name }}"
			enabled     = true
			direction   = "TO"
			action      = "ALLOW"
			source      = {}
			destination = {}
			tracking = {
				event = {
					enabled = true
				}
				alert = {
					enabled   = false
					frequency = "DAILY"
				}
			}
		}
	}
	`,
}
