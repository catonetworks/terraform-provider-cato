//go:build acctest

package static_host

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
	"text/template"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccStaticHost(t *testing.T) {
	acc.SkipByEnv(t)
	mockSrv := accmock.NewMockServer(t, "TestAccStaticHost")
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newStaticHostCfg(t)
	res := "cato_static_host.this"
	var siteID, hostID string

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
					resource.TestCheckResourceAttrSet(res, "id"),
					resource.TestCheckResourceAttr(res, "ip", "192.168.220.20"),
					resource.TestCheckResourceAttr(res, "mac_address", "00:00:00:00:00:50"),
					resource.TestCheckResourceAttr(res, "name", cfg.resName+"_host"),
					resource.TestCheckResourceAttrSet(res, "site_id"),
					captureStaticHostIDs(res, &siteID, &hostID),
				),
			},
			{
				// Import requires both the parent site ID and static host ID.
				ResourceName:      res,
				ImportState:       true,
				ImportStateIdFunc: staticHostImportID(res),
				ImportStateVerify: true,
			},
			{
				// Simulate a CMA edit, then ensure Read refreshes Terraform state.
				PreConfig: func() {
					if accmock.ACCMockActive {
						return
					}
					_, err := acc.GetClient(t).SiteUpdateStaticHost(
						context.Background(),
						hostID,
						cato_models.UpdateStaticHostInput{
							Name:       ptr(cfg.resName + "_host_drifted"),
							IP:         ptr("192.168.220.22"),
							MacAddress: ptr("00:00:00:00:00:52"),
						},
						acc.CatoAccountID,
					)
					if err != nil {
						t.Fatalf("failed to update static host outside Terraform: %v", err)
					}
				},
				SkipFunc: func() (bool, error) {
					return accmock.ACCMockActive, nil
				},
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(res, "ip", "192.168.220.22"),
					resource.TestCheckResourceAttr(res, "mac_address", "00:00:00:00:00:52"),
					resource.TestCheckResourceAttr(res, "name", cfg.resName+"_host_drifted"),
				),
			},
			{
				// Update the resource through Terraform.
				Config: cfg.getTfConfig(1),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "%", "5"),
					resource.TestCheckResourceAttrSet(res, "id"),
					resource.TestCheckResourceAttr(res, "ip", "192.168.220.21"),
					resource.TestCheckResourceAttr(res, "mac_address", "00:00:00:00:00:51"),
					resource.TestCheckResourceAttr(res, "name", cfg.resName+"_host-2"),
					resource.TestCheckResourceAttrSet(res, "site_id"),
				),
			},
		},
	})
}

func captureStaticHostIDs(resourceName string, siteID, hostID *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[resourceName]
		if !ok || resourceState.Primary == nil {
			return fmt.Errorf("resource %q not found in Terraform state", resourceName)
		}

		*siteID = resourceState.Primary.Attributes["site_id"]
		*hostID = resourceState.Primary.ID
		if *siteID == "" || *hostID == "" {
			return fmt.Errorf("resource %q has empty site_id or id", resourceName)
		}

		return nil
	}
}

func staticHostImportID(resourceName string) resource.ImportStateIdFunc {
	return func(state *terraform.State) (string, error) {
		resourceState, ok := state.RootModule().Resources[resourceName]
		if !ok || resourceState.Primary == nil {
			return "", fmt.Errorf("resource %q not found in Terraform state", resourceName)
		}

		siteID, hostID := resourceState.Primary.Attributes["site_id"], resourceState.Primary.ID
		if siteID == "" || hostID == "" {
			return "", errors.New("static host site_id or id is empty")
		}

		return siteID + "/" + hostID, nil
	}
}

type staticHostCfg struct {
	resName string
	t       *testing.T
}

func newStaticHostCfg(t *testing.T) staticHostCfg {
	return staticHostCfg{
		resName: acc.GetRandName("static_host"),
		t:       t,
	}
}

func (p staticHostCfg) getTfConfig(index int) string {
	tmpl, err := template.New("tmpl").Parse(staticHostTFs[index])
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

var staticHostTFs = []string{
	siteResource + `
	resource "cato_static_host" "this" {
		site_id     = cato_socket_site.this.id
		name        = "{{.Name}}_host"
		ip          = "192.168.220.20"
		mac_address = "00:00:00:00:00:50"
	}
	`,
	siteResource + `
	resource "cato_static_host" "this" {
		site_id     = cato_socket_site.this.id
		name        = "{{.Name}}_host-2"
		ip          = "192.168.220.21"
		mac_address = "00:00:00:00:00:51"
	}
	`,
}

func ptr(value string) *string {
	return &value
}

const siteResource = `
	resource "cato_socket_site" "this" {
		name            = "{{.Name}}"
		description     = "{{.Name}} description"
		site_type       = "BRANCH"
		connection_type = "SOCKET_X1500"

		native_range = {
			native_network_range = "192.168.220.0/24"
			local_ip             = "192.168.220.1"
			dhcp_settings = {
				dhcp_type = "DHCP_RANGE"
				ip_range  = "192.168.220.10-192.168.220.22"
			}
		}

		site_location = {
			country_code = "FR"
			timezone     = "Europe/Paris"
		}
	}
`

// TODO:  Add ImportState step, fix TF bug
