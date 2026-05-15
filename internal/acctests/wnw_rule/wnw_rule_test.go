//go:build acctest

package wnw_rule

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccWanNetwork_Simple(t *testing.T) {
	mockSrv := accmock.NewMockServer(t, "TestAccWanNetwork_Simple")
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newWanNetworkCfg(t)
	res := "cato_wnw_rule.simple"

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
					resource.TestCheckResourceAttr(res, "rule.bandwidth_priority.id", "-1"),
					resource.TestCheckResourceAttr(res, "rule.bandwidth_priority.name", "255"),
					resource.TestCheckResourceAttr(res, "rule.configuration.active_tcp_acceleration", "false"),
					resource.TestCheckResourceAttr(res, "rule.configuration.packet_loss_mitigation", "true"),
					resource.TestCheckResourceAttr(res, "rule.configuration.preserve_source_port", "false"),
					resource.TestCheckResourceAttr(res, "rule.configuration.primary_transport.primary_interface_role", "WAN1"),
					resource.TestCheckResourceAttr(res, "rule.configuration.primary_transport.secondary_interface_role", "WAN1"),
					resource.TestCheckResourceAttr(res, "rule.configuration.primary_transport.transport_type", "ALTERNATIVE_WAN"),
					resource.TestCheckResourceAttr(res, "rule.configuration.secondary_transport.primary_interface_role", "AUTOMATIC"),
					resource.TestCheckResourceAttr(res, "rule.configuration.secondary_transport.secondary_interface_role", "NONE"),
					resource.TestCheckResourceAttr(res, "rule.configuration.secondary_transport.transport_type", "NONE"),
					resource.TestCheckResourceAttr(res, "rule.enabled", "true"),
					resource.TestCheckResourceAttrSet(res, "rule.id"),
					resource.TestCheckResourceAttr(res, "rule.name", cfg.resName),
					resource.TestCheckResourceAttr(res, "rule.route_type", "NONE"),
					resource.TestCheckResourceAttr(res, "rule.rule_type", "WAN"),
				),
			},
		},
	})
}

type wanNetworkCfg struct {
	resName string
	t       *testing.T
}

func newWanNetworkCfg(t *testing.T) wanNetworkCfg {
	return wanNetworkCfg{
		resName: acc.GetRandName("wan_network"),
		t:       t,
	}
}

func (p wanNetworkCfg) prepareTfCfg(data map[string]any, tmplText string) string {
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

func (p wanNetworkCfg) getTfConfigSimple(index int) string {
	data := map[string]any{
		"Name": p.resName,
	}
	return p.prepareTfCfg(data, wanNetworkSimpleTFs[index])
}

var wanNetworkSimpleTFs = []string{
	`resource "cato_wnw_rule" "simple" {
		at = {
			position = "LAST_IN_POLICY"
		}
		rule = {
			name       = "{{ .Name }}"
			enabled    = true
			rule_type  = "WAN"
			route_type = "NONE"

			bandwidth_priority = {
				name = "255"
			}

			configuration = {
				active_tcp_acceleration = false
				packet_loss_mitigation  = true
				preserve_source_port    = false
				primary_transport = {
					primary_interface_role   = "WAN1"
					secondary_interface_role = "WAN1"
					transport_type           = "ALTERNATIVE_WAN"
				}
				secondary_transport = {
					primary_interface_role   = "AUTOMATIC"
					secondary_interface_role = "NONE"
					transport_type           = "NONE"
				}
			}

			source      = {}
			destination = {}
			application = {}
		}
	}
	`,
}
