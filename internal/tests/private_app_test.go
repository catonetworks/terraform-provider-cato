package tests

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"slices"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// example.Widget represents a concrete Go type that represents an API resource
func TestAccPrivateApp_basic(t *testing.T) {
	cfg := newPrivateAppCfg(t)
	myApp := "cato_private_app.this"

	timeRE := regexp.MustCompile(`^\d\d\d\d-\d\d-\d\dT\d\d:\d\d:\d\d`)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 checkCMAVars(t),
		Steps: []resource.TestStep{
			{
				PreConfig: func() { fmt.Println(cfg.getPrivateApp(0)) },
				Config:    cfg.getPrivateApp(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					printAttributes(myApp),
					resource.TestCheckResourceAttr(myApp, "allow_icmp_protocol", "false"),
					resource.TestMatchResourceAttr(myApp, "creation_time", timeRE),
					resource.TestCheckResourceAttr(myApp, "description", cfg.resName+" description"),
					resource.TestCheckResourceAttr(myApp, "internal_app_address", cfg.ipAddr),
					resource.TestCheckResourceAttr(myApp, "name", cfg.resName),
					resource.TestCheckResourceAttr(myApp, "private_app_probing.%", "4"),
					resource.TestCheckResourceAttr(myApp, "private_app_probing.fault_threshold_down", "10"),
					resource.TestCheckResourceAttrSet(myApp, "private_app_probing.id"),
					resource.TestCheckResourceAttr(myApp, "private_app_probing.interval", "5"),
					resource.TestCheckResourceAttr(myApp, "private_app_probing.type", "ICMP_PING"),
					resource.TestCheckResourceAttr(myApp, "probing_enabled", "true"),

					resource.TestCheckResourceAttr(myApp, "protocol_ports.#", "3"),
					resource.TestCheckResourceAttr(myApp, "protocol_ports.0.protocol", "TCP"),
					resource.TestCheckTypeSetElemNestedAttrs(myApp, "protocol_ports.*",
						map[string]string{"protocol": "TCP", "ports.#": "1", "ports.0": "80"},
					),
					resource.TestCheckTypeSetElemNestedAttrs(myApp, "protocol_ports.*",
						map[string]string{"protocol": "TCP", "ports.#": "1", "ports.0": "443"},
					),
					resource.TestCheckTypeSetElemNestedAttrs(myApp, "protocol_ports.*",
						map[string]string{"protocol": "TCP", "port_range.from": "6000", "port_range.to": "6010"},
					),
					resource.TestCheckResourceAttr(myApp, "published", "true"),
					resource.TestCheckResourceAttr(myApp, "published_app_domain.%", "4"),
					resource.TestCheckResourceAttr(myApp, "published_app_domain.connector_group_name", "conn-group-1"),
					resource.TestMatchResourceAttr(myApp, "published_app_domain.creation_time", timeRE),
					resource.TestCheckResourceAttrSet(myApp, "published_app_domain.id"),
					resource.TestCheckResourceAttr(myApp, "published_app_domain.published_app_domain", cfg.resName+".example.com"),
				),
			},
			{
				PreConfig: func() { fmt.Println(cfg.getPrivateApp(1)) },
				Config:    cfg.getPrivateApp(1),
				Check: resource.ComposeAggregateTestCheckFunc(
					printAttributes(myApp),
					resource.TestCheckResourceAttr(myApp, "allow_icmp_protocol", "true"),
				),
			},
		},
	})
}

func printAttributes(resource string) func(st *terraform.State) error {
	return func(st *terraform.State) error {
		attrs := st.Modules[0].Resources[resource].Primary.Attributes
		keys := make([]string, 0, len(attrs))
		for k := range attrs {
			keys = append(keys, k)
		}
		slices.Sort(keys)

		fmt.Printf("Resource attributes (%s):\n", resource)
		for _, k := range keys {
			fmt.Printf("\t%s: %s\n", k, attrs[k])
		}
		return nil
	}
}

type privateAppCfg struct {
	resName string
	ipAddr  string
	t       *testing.T
}

func newPrivateAppCfg(t *testing.T) privateAppCfg {
	return privateAppCfg{
		resName: getRandName("private_app"),
		ipAddr:  getRandIP(),
		t:       t,
	}
}

func (p privateAppCfg) getPrivateApp(index int) string {
	providerCfg := fmt.Sprintf("provider \"cato\" {\n  account_id = \"%s\"\n}\n", os.Getenv("CATO_ACCOUNT_ID"))

	tmpl, err := template.New("tmpl").Parse(private_app_resources[index])
	if err != nil {
		p.t.Fatal(err)
	}
	var buf bytes.Buffer
	data := map[string]string{"Name": p.resName, "IP": p.ipAddr}
	if err := tmpl.Execute(&buf, data); err != nil {
		p.t.Fatal(err)
	}

	return providerCfg + buf.String()
}

var private_app_resources = []string{
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
			connector_group_name = "conn-group-1"
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
