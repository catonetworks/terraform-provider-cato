//go:build acctest

package license

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"
	"text/template"

	cato "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccLicense(t *testing.T) {
	acc.SkipByEnv(t)
	mockSrv := accmock.NewMockServer(t, "TestAccLicense")
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newLicenseCfg(t)
	res := "cato_license.this"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		CheckDestroy:             testAccLicenseDestroy,
		Steps: []resource.TestStep{
			{
				// Create the resource
				Config: cfg.getTfConfig(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttrSet(res, "id"),
					resource.TestCheckResourceAttr(res, "license_id", cfg.licenseID),
					resource.TestCheckResourceAttrSet(res, "site_id"),
					resource.TestCheckResourceAttr(res, "license_info.sku", cfg.sku),
					resource.TestCheckResourceAttr(res, "license_info.status", "ACTIVE"),
				),
			},
			{
				ImportState:  true,
				ResourceName: res,
			},
			{
				// Update license assignment (for pooled bandwidth licenses).
				Config: cfg.getTfConfig(1),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttrSet(res, "id"),
					resource.TestCheckResourceAttr(res, "license_id", cfg.licenseID),
					resource.TestCheckResourceAttrSet(res, "site_id"),
					resource.TestCheckResourceAttr(res, "license_info.sku", cfg.sku),
					resource.TestCheckResourceAttr(res, "license_info.status", "ACTIVE"),
				),
			},
		},
	})
}

func testAccLicenseDestroy(st *terraform.State) error {
	client, err := cato.New(os.Getenv("CATO_BASEURL"), os.Getenv("CATO_TOKEN"), acc.CatoAccountID, nil, map[string]string{
		"User-Agent": "cato-terraform-test",
	})
	if err != nil {
		return err
	}
	licensingInfoResponse, err := client.Licensing(context.Background(), acc.CatoAccountID)
	if err != nil {
		return err
	}

	for _, rs := range st.RootModule().Resources {
		if rs.Type != "cato_license" {
			continue
		}

		licenseID := rs.Primary.Attributes["license_id"]
		for _, curLicense := range licensingInfoResponse.GetLicensing().GetLicensingInfo().GetLicenses() {
			if curLicense.ID == nil || *curLicense.ID != licenseID {
				continue
			}

			if curLicense.SiteLicense.Site != nil {
				return fmt.Errorf("license %s is still attached to a site", licenseID)
			}
		}
	}

	return nil
}

type licenseCfg struct {
	resName   string
	licenseID string
	sku       string
	bw        int64
	t         *testing.T
}

func newLicenseCfg(t *testing.T) licenseCfg {
	licenseID, sku, bw := getAssignableLicense(t)
	return licenseCfg{
		resName:   acc.GetRandName("license"),
		licenseID: licenseID,
		sku:       sku,
		bw:        bw,
		t:         t,
	}
}

func getAssignableLicense(t *testing.T) (license string, sku string, bw int64) {
	const (
		statusActive = "ACTIVE"
		skuCatoPB    = "CATO_PB"
	)
	client := acc.GetClient(t)
	licensing, err := client.Licensing(context.Background(), acc.CatoAccountID)
	if err != nil {
		t.Fatalf("ERROR fetching licenses: %v", err)
	}

	for _, curLicense := range licensing.GetLicensing().GetLicensingInfo().GetLicenses() {
		if curLicense.ID == nil || curLicense.Sku != skuCatoPB || curLicense.Status != statusActive {
			continue
		}
		available := curLicense.PooledBandwidthLicense.Total - curLicense.PooledBandwidthLicense.AllocatedBandwidth
		if available >= 10 {
			return *curLicense.ID, skuCatoPB, 10
		}
	}

	t.Skip("skipping license acceptance test: no active CATO_PB license with at least 10 Mbps available")
	return "", "", 0
}

func (p licenseCfg) getTfConfig(index int) string {
	tmpl, err := template.New("tmpl").Parse(licenseTFs[index])
	if err != nil {
		p.t.Fatal(err)
	}
	var buf bytes.Buffer
	data := map[string]any{
		"Name":      p.resName,
		"LicenseID": p.licenseID,
		"SKU":       p.sku,
		"BW":        p.bw,
		"BWUpdate":  p.bw + 1,
	}
	if err := tmpl.Execute(&buf, data); err != nil {
		p.t.Fatal(err)
	}

	cfg := acc.ProviderCfg() + buf.String()
	fmt.Println(cfg)
	return cfg
}

var licenseTFs = []string{
	`resource "cato_socket_site" "this" {
		name            = "{{.Name}}"
		description     = "{{.Name}} description"
		site_type       = "BRANCH"
		connection_type = "SOCKET_X1500"

		native_range = {
			native_network_range = "192.168.250.0/24"
			local_ip             = "192.168.250.1"
			dhcp_settings = {
				dhcp_type = "DHCP_RANGE"
				ip_range  = "192.168.250.10-192.168.250.22"
			}
		}

		site_location = {
			country_code = "FR"
			timezone     = "Europe/Paris"
		}
	}

	resource "cato_license" "this" {
		site_id    = cato_socket_site.this.id
		license_id = "{{.LicenseID}}"
		bw         = {{.BW}}
	}
	`,
	`resource "cato_socket_site" "this" {
		name            = "{{.Name}}-2"
		description     = "{{.Name}} description 2"
		site_type       = "BRANCH"
		connection_type = "SOCKET_X1500"

		native_range = {
			native_network_range = "192.168.251.0/24"
			local_ip             = "192.168.251.1"
			dhcp_settings = {
				dhcp_type = "DHCP_RANGE"
				ip_range  = "192.168.251.10-192.168.251.22"
			}
		}

		site_location = {
			country_code = "FR"
			timezone     = "Europe/Paris"
		}
	}

	resource "cato_license" "this" {
		site_id    = cato_socket_site.this.id
		license_id = "{{.LicenseID}}"
		bw         = {{.BWUpdate}}
	}
	`,
}
