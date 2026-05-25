//go:build acctest

package wan_interface

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccWanInterface(t *testing.T) {
	acc.SkipByEnv(t)
	mockSrv := accmock.NewMockServer(t, "TestAccWanInterface")
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newWanInterfaceCfg(t)
	res := "cato_wan_interface.this"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				// Create the resource
				Config: cfg.getTfConfig(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "%", "8"),
					resource.TestCheckResourceAttr(res, "downstream_bandwidth", "50"),
					resource.TestCheckResourceAttrSet(res, "id"),
					resource.TestCheckResourceAttr(res, "interface_id", "WAN2"),
					resource.TestCheckResourceAttr(res, "name", cfg.resName+"_wan"),
					resource.TestCheckResourceAttr(res, "precedence", "PASSIVE"),
					resource.TestCheckResourceAttr(res, "role", "wan_2"),
					resource.TestCheckResourceAttrSet(res, "site_id"),
					resource.TestCheckResourceAttr(res, "upstream_bandwidth", "25"),
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
					resource.TestCheckResourceAttr(res, "%", "8"),
					resource.TestCheckResourceAttr(res, "downstream_bandwidth", "100"),
					resource.TestCheckResourceAttrSet(res, "id"),
					resource.TestCheckResourceAttr(res, "interface_id", "WAN2"),
					resource.TestCheckResourceAttr(res, "name", cfg.resName+"_wan-2"),
					resource.TestCheckResourceAttr(res, "precedence", "PASSIVE"),
					resource.TestCheckResourceAttr(res, "role", "wan_2"),
					resource.TestCheckResourceAttrSet(res, "site_id"),
					resource.TestCheckResourceAttr(res, "upstream_bandwidth", "50"),
				),
			},
			{
				// Update precedence to ACTIVE (regression coverage for repeated precedence drift)
				Config: cfg.getTfConfig(2),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "%", "8"),
					resource.TestCheckResourceAttr(res, "downstream_bandwidth", "100"),
					resource.TestCheckResourceAttrSet(res, "id"),
					resource.TestCheckResourceAttr(res, "interface_id", "WAN2"),
					resource.TestCheckResourceAttr(res, "name", cfg.resName+"_wan-2"),
					resource.TestCheckResourceAttr(res, "precedence", "ACTIVE"),
					resource.TestCheckResourceAttr(res, "role", "wan_2"),
					resource.TestCheckResourceAttrSet(res, "site_id"),
					resource.TestCheckResourceAttr(res, "upstream_bandwidth", "50"),
				),
			},
		},
	})
}

type wanInterfaceCfg struct {
	resName string
	t       *testing.T
}

func newWanInterfaceCfg(t *testing.T) wanInterfaceCfg {
	return wanInterfaceCfg{
		resName: acc.GetRandName("wan_interface"),
		t:       t,
	}
}

func (p wanInterfaceCfg) getTfConfig(index int) string {
	tmpl, err := template.New("tmpl").Parse(wanInterfaceTFs[index])
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

var wanInterfaceTFs = []string{
	siteResource + `
	resource "cato_wan_interface" "this" {
		site_id              = cato_socket_site.this.id
		interface_id         = "WAN2"
		name                 = "{{.Name}}_wan"
		upstream_bandwidth   = 25
		downstream_bandwidth = 50
		role                 = "wan_2"
		precedence           = "PASSIVE"
	}
	`,
	siteResource + `
	resource "cato_wan_interface" "this" {
		site_id              = cato_socket_site.this.id
		interface_id         = "WAN2"
		name                 = "{{.Name}}_wan-2"
		upstream_bandwidth   = 50
		downstream_bandwidth = 100
		role                 = "wan_2"
		precedence           = "PASSIVE"
	}
	`,
	siteResource + `
	resource "cato_wan_interface" "this" {
		site_id              = cato_socket_site.this.id
		interface_id         = "WAN2"
		name                 = "{{.Name}}_wan-2"
		upstream_bandwidth   = 50
		downstream_bandwidth = 100
		role                 = "wan_2"
		precedence           = "ACTIVE"
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
			native_network_range = "192.168.230.0/24"
			local_ip             = "192.168.230.1"
			dhcp_settings = {
				dhcp_type = "DHCP_RANGE"
				ip_range  = "192.168.230.10-192.168.230.22"
			}
		}

		site_location = {
			country_code = "FR"
			timezone     = "Europe/Paris"
		}
	}
	`
