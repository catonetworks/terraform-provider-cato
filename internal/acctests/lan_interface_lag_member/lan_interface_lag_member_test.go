//go:build acctest

package lan_interface_lag_member

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccLanInterfaceLagMember(t *testing.T) {
	mockSrv := accmock.NewMockServer(t, "TestAccLanInterfaceLagMember")
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newLanInterfaceLagMemberCfg(t)
	res := "cato_lan_interface_lag_member.this"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				// Create the resource
				Config: cfg.getTfConfig(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "%", "5"),
					resource.TestCheckResourceAttr(res, "dest_type", "LAN_LAG_MEMBER"),
					resource.TestCheckResourceAttrSet(res, "id"),
					resource.TestCheckResourceAttr(res, "interface_id", "INT_6"),
					resource.TestCheckResourceAttr(res, "name", cfg.resName+"_lag_member"),
					resource.TestCheckResourceAttrSet(res, "site_id"),
				),
			},
		},
	})
}

type lanInterfaceLagMemberCfg struct {
	resName string
	t       *testing.T
}

func newLanInterfaceLagMemberCfg(t *testing.T) lanInterfaceLagMemberCfg {
	return lanInterfaceLagMemberCfg{
		resName: acc.GetRandName("lan_interface_lag_member"),
		t:       t,
	}
}

func (p lanInterfaceLagMemberCfg) getTfConfig(index int) string {
	tmpl, err := template.New("tmpl").Parse(lanInterfaceLagMemberTFs[index])
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

var lanInterfaceLagMemberTFs = []string{
	`resource "cato_socket_site" "this" {
		name            = "{{.Name}}"
		description     = "{{.Name}} description"
		site_type       = "BRANCH"
		connection_type = "SOCKET_X1700"

		native_range = {
			native_network_range = "192.168.240.0/24"
			local_ip             = "192.168.240.1"
			dhcp_settings = {
				dhcp_type = "DHCP_RANGE"
				ip_range  = "192.168.240.10-192.168.240.22"
			}
		}

		site_location = {
			country_code = "FR"
			timezone     = "Europe/Paris"
		}
	}

	resource "cato_lan_interface" "lag_master" {
		site_id       = cato_socket_site.this.id
		interface_id  = "INT_5"
		name          = "{{.Name}}_lag_master"
		dest_type     = "LAN_LAG_MASTER"
		local_ip      = "192.168.241.1"
		lag_min_links = 1
		subnet        = "192.168.241.0/24"
	}

	resource "cato_lan_interface_lag_member" "this" {
		depends_on   = [cato_lan_interface.lag_master]
		site_id      = cato_socket_site.this.id
		interface_id = "INT_6"
		name         = "{{.Name}}_lag_member"
		dest_type    = "LAN_LAG_MEMBER"
	}
	`,
}
