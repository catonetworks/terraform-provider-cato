//go:build acctest

package network_range

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccNetworkRange(t *testing.T) {
	acc.SkipByEnv(t)
	mockSrv := accmock.NewMockServer(t, "TestAccNetworkRange")
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newNetworkRangeCfg(t)
	res := "cato_network_range.this"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		ErrorCheck: func(err error) error {
			if err != nil && (strings.Contains(err.Error(), "Please select a country") ||
				strings.Contains(err.Error(), "Please select a state") ||
				strings.Contains(err.Error(), "state [")) {
				t.Skip("skipping network_range CRUD: account locale metadata rejects test site location")
			}
			return err
		},
		Steps: []resource.TestStep{
			{
				// Create the resource
				Config: cfg.getTfConfig(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "%", "14"),
					resource.TestCheckResourceAttr(res, "dhcp_settings.%", "5"),
					resource.TestCheckResourceAttr(res, "dhcp_settings.dhcp_microsegmentation", "false"),
					resource.TestCheckResourceAttr(res, "dhcp_settings.dhcp_type", "DHCP_RANGE"),
					resource.TestCheckResourceAttr(res, "dhcp_settings.ip_range", "192.168.242.10-192.168.242.22"),
					resource.TestCheckResourceAttrSet(res, "id"),
					resource.TestCheckResourceAttrSet(res, "interface_id"),
					resource.TestCheckResourceAttr(res, "interface_index", "LAN1"),
					resource.TestCheckResourceAttr(res, "internet_only", "false"),
					resource.TestCheckResourceAttr(res, "local_ip", "192.168.242.1"),
					resource.TestCheckResourceAttr(res, "mdns_reflector", "false"),
					resource.TestCheckResourceAttr(res, "name", cfg.resName+"_range"),
					resource.TestCheckResourceAttr(res, "range_type", "VLAN"),
					resource.TestCheckResourceAttrSet(res, "site_id"),
					resource.TestCheckResourceAttr(res, "subnet", "192.168.242.0/24"),
					resource.TestCheckResourceAttr(res, "translated_subnet", "192.168.242.0/24"),
					resource.TestCheckResourceAttr(res, "vlan", "201"),
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
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "%", "14"),
					resource.TestCheckResourceAttr(res, "dhcp_settings.%", "5"),
					resource.TestCheckResourceAttr(res, "dhcp_settings.dhcp_microsegmentation", "true"),
					resource.TestCheckResourceAttr(res, "dhcp_settings.dhcp_type", "DHCP_RANGE"),
					resource.TestCheckResourceAttr(res, "dhcp_settings.ip_range", "192.168.242.20-192.168.242.32"),
					resource.TestCheckResourceAttrSet(res, "id"),
					resource.TestCheckResourceAttrSet(res, "interface_id"),
					resource.TestCheckResourceAttr(res, "interface_index", "LAN1"),
					resource.TestCheckResourceAttr(res, "internet_only", "false"),
					resource.TestCheckResourceAttr(res, "local_ip", "192.168.242.2"),
					resource.TestCheckResourceAttr(res, "mdns_reflector", "false"),
					resource.TestCheckResourceAttr(res, "name", cfg.resName+"_range"),
					resource.TestCheckResourceAttr(res, "range_type", "VLAN"),
					resource.TestCheckResourceAttrSet(res, "site_id"),
					resource.TestCheckResourceAttr(res, "subnet", "192.168.242.0/24"),
					resource.TestCheckResourceAttr(res, "translated_subnet", "192.168.242.0/24"),
					resource.TestCheckResourceAttr(res, "vlan", "202"),
				),
			},
		},
	})
}

type networkRangeCfg struct {
	resName string
	t       *testing.T
}

func newNetworkRangeCfg(t *testing.T) networkRangeCfg {
	return networkRangeCfg{
		resName: acc.GetRandName("network_range"),
		t:       t,
	}
}

func (p networkRangeCfg) getTfConfig(index int) string {
	tmpl, err := template.New("tmpl").Parse(networkRangeTFs[index])
	if err != nil {
		p.t.Fatal(err)
	}
	var buf bytes.Buffer
	data := map[string]any{
		"Name": p.resName,
	}
	if err := tmpl.Execute(&buf, data); err != nil {
		p.t.Fatal(err)
	}

	cfg := acc.ProviderCfg() + buf.String()
	fmt.Println(cfg)
	return cfg
}

var networkRangeTFs = []string{
	siteResource + `
	resource "cato_network_range" "this" {
		site_id         = cato_socket_site.this.id
		interface_index = "LAN1"
		name            = "{{.Name}}_range"
		range_type      = "VLAN"
		subnet          = "192.168.242.0/24"
		local_ip        = "192.168.242.1"
		vlan            = 201
		dhcp_settings = {
			dhcp_type              = "DHCP_RANGE"
			ip_range               = "192.168.242.10-192.168.242.22"
			dhcp_microsegmentation = false
		}
	}
	`,
	siteResource + `
	resource "cato_network_range" "this" {
		site_id         = cato_socket_site.this.id
		interface_index = "LAN1"
		name            = "{{.Name}}_range"
		range_type      = "VLAN"
		subnet          = "192.168.242.0/24"
		local_ip        = "192.168.242.2"
		vlan            = 202
		dhcp_settings = {
			dhcp_type              = "DHCP_RANGE"
			ip_range               = "192.168.242.20-192.168.242.32"
			dhcp_microsegmentation = true
		}
	}
	`,
}

const siteResource = `
	resource "cato_socket_site" "this" {
		name            = "{{.Name}}"
		description     = "{{.Name}} description"
		site_type       = "BRANCH"
		connection_type = "SOCKET_X1500"

		native_range = {
			native_network_range = "192.168.252.0/24"
			local_ip             = "192.168.252.1"
			dhcp_settings = {
				dhcp_type = "DHCP_RANGE"
				ip_range  = "192.168.252.10-192.168.252.22"
			}
		}

		site_location = {
			country_code = "FR"
			timezone     = "Europe/Paris"
		}
	}
`
