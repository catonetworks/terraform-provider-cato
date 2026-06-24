//go:build acctest

package ipsec_site

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"text/template"
	"time"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

// TestAccIpsecSite_Basic tests basic CRUD operations for cato_ipsec_site resource.
func TestAccIpsecSite_Basic(t *testing.T) {
	acc.SkipByEnv(t)
	mockSrv := accmock.NewMockServer(t, "TestAccIpsecSite_Basic")
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newIpsecSiteCfg(t)
	const res = "cato_ipsec_site.this"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				// Create the resource
				Config: cfg.getTfConfigBasic(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "%", "9"),
					resource.TestCheckResourceAttr(res, "description", cfg.resName+" description"),
					resource.TestCheckResourceAttrSet(res, "id"),
					resource.TestCheckResourceAttrSet(res, "interface_id"),
					resource.TestCheckResourceAttr(res, "ipsec.%", "8"),
					resource.TestCheckResourceAttr(res, "ipsec.primary.%", "4"),
					resource.TestCheckResourceAttr(res, "ipsec.primary.public_cato_ip_id", cfg.allocatedIPIDs[0]),
					resource.TestCheckResourceAttr(res, "ipsec.primary.tunnels.#", "1"),
					resource.TestCheckResourceAttr(res, "ipsec.primary.tunnels.0.%", "6"),
					resource.TestCheckResourceAttr(res, "ipsec.primary.tunnels.0.last_mile_bw.%", "4"),
					resource.TestCheckResourceAttr(res, "ipsec.primary.tunnels.0.last_mile_bw.downstream", "10"),
					resource.TestCheckResourceAttr(res, "ipsec.primary.tunnels.0.last_mile_bw.upstream", "10"),
				resource.TestCheckResourceAttr(res, "ipsec.primary.tunnels.0.private_cato_ip", "192.168.249.10"),
				resource.TestCheckResourceAttr(res, "ipsec.primary.tunnels.0.private_site_ip", "192.168.249.20"),
				resource.TestCheckResourceAttr(res, "ipsec.primary.tunnels.0.psk", "acctest-ipsec-site-psk"),
				resource.TestCheckResourceAttr(res, "ipsec.primary.tunnels.0.public_site_ip", "203.0.113.10"),
				resource.TestCheckResourceAttr(res, "ipsec.primary.tunnels.0.tunnel_id", "PRIMARY1"),
				resource.TestCheckResourceAttrSet(res, "ipsec.site_id"),
				resource.TestCheckResourceAttr(res, "name", cfg.resName),
				resource.TestCheckResourceAttr(res, "native_network_range", "192.168.249.0/24"),
					resource.TestCheckResourceAttrSet(res, "native_network_range_id"),
					resource.TestCheckResourceAttr(res, "site_location.%", "5"),
					resource.TestCheckResourceAttr(res, "site_location.country_code", "FR"),
					resource.TestCheckResourceAttr(res, "site_location.timezone", "Europe/Paris"),
					resource.TestCheckResourceAttr(res, "site_type", "CLOUD_DC"),
				),
			},
			{
				// Test import mode
				ImportState:  true,
				ResourceName: res,
			},
			{
				// Update the resource
				Config: cfg.getTfConfigBasic(1),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "%", "9"),
					resource.TestCheckResourceAttr(res, "description", cfg.resName+" description 2"),
					resource.TestCheckResourceAttrSet(res, "id"),
					resource.TestCheckResourceAttrSet(res, "interface_id"),
					resource.TestCheckResourceAttr(res, "ipsec.%", "8"),
					resource.TestCheckResourceAttr(res, "ipsec.primary.%", "4"),
					resource.TestCheckResourceAttr(res, "ipsec.primary.public_cato_ip_id", cfg.allocatedIPIDs[0]),
					resource.TestCheckResourceAttr(res, "ipsec.primary.tunnels.#", "1"),
					resource.TestCheckResourceAttr(res, "ipsec.primary.tunnels.0.%", "6"),
					resource.TestCheckResourceAttr(res, "ipsec.primary.tunnels.0.last_mile_bw.%", "4"),
					resource.TestCheckResourceAttr(res, "ipsec.primary.tunnels.0.last_mile_bw.downstream", "20"),
					resource.TestCheckResourceAttr(res, "ipsec.primary.tunnels.0.last_mile_bw.upstream", "15"),
					resource.TestCheckResourceAttr(res, "ipsec.primary.tunnels.0.private_cato_ip", "192.168.253.10"),
					resource.TestCheckResourceAttr(res, "ipsec.primary.tunnels.0.private_site_ip", "192.168.253.20"),
					resource.TestCheckResourceAttr(res, "ipsec.primary.tunnels.0.psk", "acctest-ipsec-site-psk-updated"),
					resource.TestCheckResourceAttr(res, "ipsec.primary.tunnels.0.public_site_ip", "203.0.113.11"),
					resource.TestCheckResourceAttr(res, "ipsec.primary.tunnels.0.tunnel_id", "PRIMARY1"),
					resource.TestCheckResourceAttrSet(res, "ipsec.site_id"),
					resource.TestCheckResourceAttr(res, "name", cfg.resName+" 2"),
					resource.TestCheckResourceAttr(res, "native_network_range", "192.168.253.0/24"),
					resource.TestCheckResourceAttrSet(res, "native_network_range_id"),
					resource.TestCheckResourceAttr(res, "site_location.%", "5"),
					resource.TestCheckResourceAttr(res, "site_location.country_code", "FR"),
					resource.TestCheckResourceAttr(res, "site_location.timezone", "Europe/Paris"),
					resource.TestCheckResourceAttr(res, "site_type", "BRANCH"),
				),
			},
		},
	})
}

type ipsecSiteCfg struct {
	resName        string
	allocatedIPIDs []string
	t              *testing.T
}

func newIpsecSiteCfg(t *testing.T) ipsecSiteCfg {
	return ipsecSiteCfg{
		resName:        acc.GetRandName("ipsec_site"),
		allocatedIPIDs: getAllocatedIPIDs(t),
		t:              t,
	}
}

func getAllocatedIPIDs(t *testing.T) []string {
	client := acc.GetClient(t)
	var resultIDs []string
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		result, err := client.EntityLookup(context.Background(), acc.CatoAccountID, cato_models.EntityTypeAllocatedIP, nil, nil, nil, nil, nil, nil, nil, nil)
		if err != nil {
			lastErr = err
			time.Sleep(time.Second)
			continue
		}

		for _, item := range result.GetEntityLookup().GetItems() {
			id := item.GetEntity().GetID()
			if id != "" {
				resultIDs = append(resultIDs, id)
			}
		}
		lastErr = nil
		break
	}
	if lastErr != nil {
		t.Fatalf("ERROR fetching allocated IPs: %v", lastErr)
	}
	if len(resultIDs) == 0 {
		t.Skip("skipping ipsec_site acceptance test: no allocated public IPs found")
	}
	return resultIDs
}

func (p ipsecSiteCfg) prepareTfCfg(data map[string]any, tmplText string) string {
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

func (p ipsecSiteCfg) getTfConfigBasic(index int) string {
	data := map[string]any{
		"Name":          p.resName,
		"AllocatedIPID": p.allocatedIPIDs[0],
	}
	return p.prepareTfCfg(data, ipsecSiteBasicTFs[index])
}

var ipsecSiteBasicTFs = []string{
	`resource "cato_ipsec_site" "this" {
		name                 = "{{.Name}}"
		description          = "{{.Name}} description"
		site_type            = "CLOUD_DC"
		native_network_range = "192.168.249.0/24"

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
						private_cato_ip = "192.168.249.10"
						private_site_ip = "192.168.249.20"
						psk             = "acctest-ipsec-site-psk"
						last_mile_bw = {
							downstream = 10
							upstream   = 10
						}
					}
				]
			}
		}
	}
	`,
	`resource "cato_ipsec_site" "this" {
		name                 = "{{.Name}} 2"
		description          = "{{.Name}} description 2"
		site_type            = "BRANCH"
		native_network_range = "192.168.253.0/24"

		site_location = {
			country_code = "FR"
			timezone     = "Europe/Paris"
		}

		ipsec = {
			primary = {
				public_cato_ip_id = "{{.AllocatedIPID}}"
				tunnels = [
					{
						public_site_ip  = "203.0.113.11"
						private_cato_ip = "192.168.253.10"
						private_site_ip = "192.168.253.20"
						psk             = "acctest-ipsec-site-psk-updated"
						last_mile_bw = {
							downstream = 20
							upstream   = 15
						}
					}
				]
			}
		}
	}
	`,
}
