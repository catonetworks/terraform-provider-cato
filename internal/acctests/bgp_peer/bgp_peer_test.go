//go:build acctest

package bgp_peer

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"
	"text/template"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

const envEnableBgpPeerCRUD = "TFACC_ENABLE_BGP_PEER_CRUD"

func TestAccBgpPeer(t *testing.T) {
	acc.SkipByEnv(t)
	if os.Getenv(envEnableBgpPeerCRUD) != "true" {
		t.Skipf("set %s=true to run BGP peer CRUD acceptance test", envEnableBgpPeerCRUD)
	}

	mockSrv := accmock.NewMockServer(t, "TestAccBgpPeer")
	defer mockSrv.Close()
	mockSrv.Run()

	cfg := newBgpPeerCfg(t)
	res := "cato_bgp_peer.this"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		CheckDestroy:             testAccBgpPeerDestroy,
		Steps: []resource.TestStep{
			{
				Config: cfg.getTfConfig(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttrSet(res, "id"),
					resource.TestCheckResourceAttrSet(res, "site_id"),
					resource.TestCheckResourceAttr(res, "name", cfg.name),
					resource.TestCheckResourceAttr(res, "peer_ip", cfg.peerIP),
					resource.TestCheckResourceAttr(res, "peer_asn", "65100"),
					resource.TestCheckResourceAttr(res, "cato_asn", "65000"),
					resource.TestCheckResourceAttr(res, "default_action", "ACCEPT"),
				),
			},
			{
				ImportState:  true,
				ResourceName: res,
			},
			{
				Config: cfg.getTfConfig(1),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttrSet(res, "id"),
					resource.TestCheckResourceAttr(res, "name", cfg.name+"-2"),
					resource.TestCheckResourceAttr(res, "default_action", "DROP"),
					resource.TestCheckResourceAttr(res, "metric", "51"),
				),
			},
		},
	})
}

func TestAccBgpPeer_InvalidDefaultAction(t *testing.T) {
	acc.SkipByEnv(t)
	if os.Getenv(envEnableBgpPeerCRUD) != "true" {
		t.Skipf("set %s=true to run BGP peer CRUD acceptance test", envEnableBgpPeerCRUD)
	}

	mockSrv := accmock.NewMockServer(t, "TestAccBgpPeer_InvalidDefaultAction")
	defer mockSrv.Close()
	mockSrv.Run()

	cfg := newBgpPeerCfg(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				Config:      cfg.getTfConfig(2),
				ExpectError: regexp.MustCompile("value must be one of"),
			},
		},
	})
}

func testAccBgpPeerDestroy(st *terraform.State) error {
	client, err := cato.New(os.Getenv("CATO_BASEURL"), os.Getenv("CATO_TOKEN"), acc.CatoAccountID, nil, map[string]string{
		"User-Agent": "cato-terraform-test",
	})
	if err != nil {
		return err
	}

	for _, rs := range st.RootModule().Resources {
		if rs.Type != "cato_bgp_peer" {
			continue
		}

		ref := cato_models.BgpPeerRefInput{
			By:    cato_models.ObjectRefByID,
			Input: rs.Primary.ID,
		}

		_, readErr := client.SiteBgpPeer(context.Background(), ref, acc.CatoAccountID)
		if readErr == nil {
			return fmt.Errorf("bgp peer %s still exists", rs.Primary.ID)
		}
	}
	return nil
}

type bgpPeerCfg struct {
	name           string
	peerIP         string
	allocatedIPIDs []string
	t              *testing.T
}

func newBgpPeerCfg(t *testing.T) bgpPeerCfg {
	return bgpPeerCfg{
		name:           acc.GetRandName("bgp_peer"),
		peerIP:         "192.168.254.20",
		allocatedIPIDs: getAllocatedIPIDs(t),
		t:              t,
	}
}

func (p bgpPeerCfg) getTfConfig(index int) string {
	tmpl, err := template.New("tmpl").Parse(bgpPeerTFs[index])
	if err != nil {
		p.t.Fatal(err)
	}
	var buf bytes.Buffer
	data := map[string]any{
		"Name":          p.name,
		"PeerIP":        p.peerIP,
		"AllocatedIPID": p.allocatedIPIDs[0],
	}
	if err := tmpl.Execute(&buf, data); err != nil {
		p.t.Fatal(err)
	}

	cfg := acc.ProviderCfg() + buf.String()
	fmt.Println(cfg)
	return cfg
}

var bgpPeerTFs = []string{
	`resource "cato_ipsec_site" "this" {
		name                 = "{{.Name}}-site"
		description          = "{{.Name}} site for bgp"
		site_type            = "CLOUD_DC"
		native_network_range = "192.168.254.0/24"

		site_location = {
			country_code = "FR"
			timezone     = "Europe/Paris"
		}

		ipsec = {
			primary = {
				public_cato_ip_id = "{{.AllocatedIPID}}"
				tunnels = [
					{
						public_site_ip  = "203.0.113.10"
						private_cato_ip = "192.168.254.10"
						private_site_ip = "192.168.254.20"
						psk             = "acctest-bgp-peer-psk"
						last_mile_bw = {
							downstream = 10
							upstream   = 10
						}
					}
				]
			}
		}
	}

	resource "cato_bgp_peer" "this" {
		site_id                = cato_ipsec_site.this.id
		name                   = "{{.Name}}"
		peer_asn               = 65100
		cato_asn               = 65000
		peer_ip                = "{{.PeerIP}}"
		default_action         = "ACCEPT"
		advertise_default_route = true
		advertise_all_routes   = false
		advertise_summary_routes = false
		perform_nat            = false
		metric                 = 50
		hold_time              = 60
		keepalive_interval     = 20
		bfd_enabled            = false
	}`,
	`resource "cato_ipsec_site" "this" {
		name                 = "{{.Name}}-site"
		description          = "{{.Name}} site for bgp"
		site_type            = "CLOUD_DC"
		native_network_range = "192.168.254.0/24"

		site_location = {
			country_code = "FR"
			timezone     = "Europe/Paris"
		}

		ipsec = {
			primary = {
				public_cato_ip_id = "{{.AllocatedIPID}}"
				tunnels = [
					{
						public_site_ip  = "203.0.113.10"
						private_cato_ip = "192.168.254.10"
						private_site_ip = "192.168.254.20"
						psk             = "acctest-bgp-peer-psk"
						last_mile_bw = {
							downstream = 10
							upstream   = 10
						}
					}
				]
			}
		}
	}

	resource "cato_bgp_peer" "this" {
		site_id                = cato_ipsec_site.this.id
		name                   = "{{.Name}}-2"
		peer_asn               = 65100
		cato_asn               = 65000
		peer_ip                = "{{.PeerIP}}"
		default_action         = "DROP"
		advertise_default_route = true
		advertise_all_routes   = false
		advertise_summary_routes = false
		perform_nat            = false
		metric                 = 51
		hold_time              = 60
		keepalive_interval     = 20
		bfd_enabled            = false
	}`,
	`resource "cato_ipsec_site" "this" {
		name                 = "{{.Name}}-site"
		description          = "{{.Name}} site for bgp"
		site_type            = "CLOUD_DC"
		native_network_range = "192.168.254.0/24"

		site_location = {
			country_code = "FR"
			timezone     = "Europe/Paris"
		}

		ipsec = {
			primary = {
				public_cato_ip_id = "{{.AllocatedIPID}}"
				tunnels = [
					{
						public_site_ip  = "203.0.113.10"
						private_cato_ip = "192.168.254.10"
						private_site_ip = "192.168.254.20"
						psk             = "acctest-bgp-peer-psk"
						last_mile_bw = {
							downstream = 10
							upstream   = 10
						}
					}
				]
			}
		}
	}

	resource "cato_bgp_peer" "this" {
		site_id         = cato_ipsec_site.this.id
		name            = "{{.Name}}-invalid"
		peer_asn        = 65100
		cato_asn        = 65000
		peer_ip         = "{{.PeerIP}}"
		default_action  = "INVALID"
	}
	`,
}

func getAllocatedIPIDs(t *testing.T) []string {
	client := acc.GetClient(t)
	result, err := client.EntityLookup(context.Background(), acc.CatoAccountID, cato_models.EntityTypeAllocatedIP, nil, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("ERROR fetching allocated IPs: %v", err)
	}

	var ids []string
	for _, item := range result.GetEntityLookup().GetItems() {
		if id := item.GetEntity().GetID(); id != "" {
			ids = append(ids, id)
		}
	}

	if len(ids) == 0 {
		t.Skip("skipping bgp_peer acceptance test: no allocated public IPs found")
	}

	return ids
}
