//go:build acctest

package lan_interface

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccLanInterface(t *testing.T) {
	acc.SkipByEnv(t)
	mockSrv := accmock.NewMockServer(t, "TestAccLanInterface")
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newLanInterfaceCfg(t)
	res := "cato_lan_interface.this"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				// Create the resource
				Config: cfg.getTfConfig(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "%", "10"),
					resource.TestCheckResourceAttr(res, "dest_type", "LAN"),
					resource.TestCheckResourceAttrSet(res, "id"),
					resource.TestCheckResourceAttr(res, "interface_id", "LAN2"),
					resource.TestCheckResourceAttr(res, "local_ip", "192.168.211.1"),
					resource.TestCheckResourceAttr(res, "name", cfg.resName+"_lan"),
					resource.TestCheckResourceAttrSet(res, "site_id"),
					resource.TestCheckResourceAttr(res, "subnet", "192.168.211.0/24"),
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
					resource.TestCheckResourceAttr(res, "%", "10"),
					resource.TestCheckResourceAttr(res, "dest_type", "LAN"),
					resource.TestCheckResourceAttrSet(res, "id"),
					resource.TestCheckResourceAttr(res, "interface_id", "LAN2"),
					resource.TestCheckResourceAttr(res, "local_ip", "192.168.211.3"),
					resource.TestCheckResourceAttr(res, "name", cfg.resName+"_lan-3"),
					resource.TestCheckResourceAttrSet(res, "site_id"),
					resource.TestCheckResourceAttr(res, "subnet", "192.168.211.0/24"),
				),
			},
		},
	})
}

type lanInterfaceCfg struct {
	resName string
	t       *testing.T
}

func newLanInterfaceCfg(t *testing.T) lanInterfaceCfg {
	return lanInterfaceCfg{
		resName: acc.GetRandName("lan_interface"),
		t:       t,
	}
}

func (p lanInterfaceCfg) getTfConfig(index int) string {
	tmpl, err := template.New("tmpl").Parse(lanInterfaceTFs[index])
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

var lanInterfaceTFs = []string{
	siteResource +
		`resource "cato_lan_interface" "this" {
		dest_type    = "LAN"
		interface_id = "LAN2"
		local_ip     = "192.168.211.1"
		name         = "{{.Name}}_lan"
		site_id      = cato_socket_site.this.id
		subnet       = "192.168.211.0/24"
	}
	`,

	siteResource +
		`resource "cato_lan_interface" "this" {
		dest_type    = "LAN"
		interface_id = "LAN2"
		local_ip     = "192.168.211.3"
		name         = "{{.Name}}_lan-3"
		site_id      = cato_socket_site.this.id
		subnet       = "192.168.211.0/24"
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
			native_network_range = "192.168.210.0/24"
			local_ip             = "192.168.210.1"
			dhcp_settings = {
				dhcp_type = "DHCP_RANGE"
				ip_range  = "192.168.210.10-192.168.210.22"
			}
		}

		site_location = {
			country_code = "FR"
			timezone     = "Europe/Paris"
		}
	}
`
