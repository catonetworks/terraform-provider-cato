package tests

import (
	"bytes"
	"fmt"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccPrivateApp(t *testing.T) {
	cfg := newPrivateAppCfg(t)
	res := "cato_private_app.this"

	timeRE := regexp.MustCompile(`^\d\d\d\d-\d\d-\d\dT\d\d:\d\d:\d\d`)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 checkCMAVars(t),
		Steps: []resource.TestStep{
			{
				// Create the resource
				Config: cfg.getTfConfig(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					printAttributes(res),
					resource.TestCheckResourceAttr(res, "allow_icmp_protocol", "false"),
					resource.TestMatchResourceAttr(res, "creation_time", timeRE),
					resource.TestCheckResourceAttr(res, "description", cfg.resName+" description"),
					resource.TestCheckResourceAttr(res, "internal_app_address", cfg.ipAddr),
					resource.TestCheckResourceAttr(res, "name", cfg.resName),
					resource.TestCheckResourceAttr(res, "private_app_probing.%", "4"),
					resource.TestCheckResourceAttr(res, "private_app_probing.fault_threshold_down", "10"),
					resource.TestCheckResourceAttrSet(res, "private_app_probing.id"),
					resource.TestCheckResourceAttr(res, "private_app_probing.interval", "5"),
					resource.TestCheckResourceAttr(res, "private_app_probing.type", "ICMP_PING"),
					resource.TestCheckResourceAttr(res, "probing_enabled", "true"),

					resource.TestCheckResourceAttr(res, "protocol_ports.#", "3"),
					resource.TestCheckResourceAttr(res, "protocol_ports.0.protocol", "TCP"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "protocol_ports.*",
						map[string]string{"protocol": "TCP", "ports.#": "1", "ports.0": "80"},
					),
					resource.TestCheckTypeSetElemNestedAttrs(res, "protocol_ports.*",
						map[string]string{"protocol": "TCP", "ports.#": "1", "ports.0": "443"},
					),
					resource.TestCheckTypeSetElemNestedAttrs(res, "protocol_ports.*",
						map[string]string{"protocol": "TCP", "port_range.from": "6000", "port_range.to": "6010"},
					),
					resource.TestCheckResourceAttr(res, "published", "true"),
					resource.TestCheckResourceAttr(res, "published_app_domain.%", "4"),
					resource.TestCheckResourceAttr(res, "published_app_domain.connector_group_name", cfg.connGroups[0]),
					resource.TestMatchResourceAttr(res, "published_app_domain.creation_time", timeRE),
					resource.TestCheckResourceAttrSet(res, "published_app_domain.id"),
					resource.TestCheckResourceAttr(res, "published_app_domain.published_app_domain", cfg.resName+".example.com"),
				),
			},
			{
				// Test import mode
				ImportState:  true,
				ResourceName: res,
			},
			{
				// Update the resource
				Config: cfg.getTfConfig(1),
				Check: resource.ComposeAggregateTestCheckFunc(
					printAttributes(res),
					resource.TestCheckResourceAttr(res, "allow_icmp_protocol", "true"),
					resource.TestMatchResourceAttr(res, "creation_time", timeRE),
					resource.TestCheckResourceAttr(res, "description", cfg.resName+" description 2"),
					resource.TestCheckResourceAttr(res, "internal_app_address", cfg.ipAddr),
					resource.TestCheckResourceAttr(res, "name", cfg.resName),
					resource.TestCheckResourceAttr(res, "private_app_probing.%", "4"),
					resource.TestCheckResourceAttr(res, "private_app_probing.fault_threshold_down", "11"),
					resource.TestCheckResourceAttrSet(res, "private_app_probing.id"),
					resource.TestCheckResourceAttr(res, "private_app_probing.interval", "6"),
					resource.TestCheckResourceAttr(res, "private_app_probing.type", "ICMP_PING"),
					resource.TestCheckResourceAttr(res, "probing_enabled", "true"),

					resource.TestCheckResourceAttr(res, "protocol_ports.#", "3"),
					resource.TestCheckResourceAttr(res, "protocol_ports.0.protocol", "UDP"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "protocol_ports.*",
						map[string]string{"protocol": "UDP", "ports.#": "1", "ports.0": "81"},
					),
					resource.TestCheckTypeSetElemNestedAttrs(res, "protocol_ports.*",
						map[string]string{"protocol": "TCP", "ports.#": "1", "ports.0": "8443"},
					),
					resource.TestCheckTypeSetElemNestedAttrs(res, "protocol_ports.*",
						map[string]string{"protocol": "UDP", "port_range.from": "5000", "port_range.to": "5010"},
					),
					resource.TestCheckResourceAttr(res, "published", "false"),
				),
			},
		},
	})
}

type privateAppCfg struct {
	resName    string
	ipAddr     string
	connGroups []string
	t          *testing.T
}

func newPrivateAppCfg(t *testing.T) privateAppCfg {
	return privateAppCfg{
		resName:    getRandName("private_app"),
		ipAddr:     getRandIP(),
		connGroups: getConnectorGroups(t),
		t:          t,
	}
}

func (p privateAppCfg) getTfConfig(index int) string {
	tmpl, err := template.New("tmpl").Parse(privateAppTFs[index])
	if err != nil {
		p.t.Fatal(err)
	}
	var buf bytes.Buffer
	data := map[string]string{
		"Name":           p.resName,
		"IP":             p.ipAddr,
		"ConnectorGroup": p.connGroups[0],
	}
	if err := tmpl.Execute(&buf, data); err != nil {
		p.t.Fatal(err)
	}

	cfg := providerCfg() + buf.String()
	fmt.Println(cfg)
	return cfg
}

var privateAppTFs = []string{
	`resource "cato_private_app" "this" {
		allow_icmp_protocol  = false
		description          = "{{.Name}} description"
		internal_app_address = "{{.IP}}"
		name                 = "{{.Name}}"
		private_app_probing = {
			fault_threshold_down = 10
			interval             = 5
			type                 = "ICMP_PING"
		}
		probing_enabled = true
		protocol_ports = [
			{ ports    = [80],  protocol = "TCP" },
			{ ports    = [443], protocol = "TCP" },
			{ port_range = { from = 6000, to = 6010 }, protocol = "TCP" }
		]
		published = true
		published_app_domain = {
			connector_group_name = "{{.ConnectorGroup}}"
			published_app_domain = "{{.Name}}.example.com"
		}
	}
	`,

	`resource "cato_private_app" "this" {
		allow_icmp_protocol  = true
		description          = "{{.Name}} description 2"
		internal_app_address = "{{.IP}}"
		name                 = "{{.Name}}"
		private_app_probing = {
			fault_threshold_down = 11
			interval             = 6
			type                 = "ICMP_PING"
		}
		probing_enabled = true
		protocol_ports = [
			{ ports    = [81],  protocol = "UDP" },
			{ ports    = [8443], protocol = "TCP" },
			{ port_range = { from = 5000, to = 5010 }, protocol = "UDP" }
		]
		published = false
	}
	`,
}
