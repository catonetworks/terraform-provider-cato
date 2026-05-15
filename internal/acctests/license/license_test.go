//go:build acctest

package license

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccLicense(t *testing.T) {
	t.Skip("No license IDs available")
	mockSrv := accmock.NewMockServer(t, "TestAccLicense")
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newLicenseCfg(t)
	res := "cato_license.this"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
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
		},
	})
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

func getAssignableLicense(t *testing.T) (string, string, int64) {
	client := acc.GetClient(t)
	licensing, err := client.Licensing(context.Background(), acc.CatoAccountID)
	if err != nil {
		t.Fatalf("ERROR fetching licenses: %v", err)
	}

	for _, curLicense := range licensing.GetLicensing().GetLicensingInfo().GetLicenses() {
		if curLicense.ID == nil || curLicense.Sku != "CATO_SITE" || curLicense.Status != "ACTIVE" {
			continue
		}
		if curLicense.SiteLicense.Site == nil {
			return *curLicense.ID, "CATO_SITE", 0
		}
	}

	for _, curLicense := range licensing.GetLicensing().GetLicensingInfo().GetLicenses() {
		if curLicense.ID == nil || curLicense.Sku != "CATO_PB" || curLicense.Status != "ACTIVE" {
			continue
		}
		available := curLicense.PooledBandwidthLicense.Total - curLicense.PooledBandwidthLicense.AllocatedBandwidth
		if available >= 10 {
			return *curLicense.ID, "CATO_PB", 10
		}
	}

	t.Skip("skipping license acceptance test: no active unassigned CATO_SITE license or active CATO_PB license with at least 10 Mbps available")
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
		{{- if .BW }}
		bw         = {{.BW}}
		{{- end }}
	}
	`,
}
