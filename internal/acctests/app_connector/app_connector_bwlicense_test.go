//go:build acctest

package app_connector

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccAppConnector_BWAllocation(t *testing.T) {
	if os.Getenv("ACC_TEST_BW_LICENSE") == "" {
		t.Skip("to enable TestAccAppConnector_BWAllocation, set ACC_TEST_BW_LICENSE=<license-id>")
	}
	acc.SkipByEnv(t)
	mockSrv := accmock.NewMockServer(t, "TestAccAppConnector")
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newAppConnectorLicenseCfg(t)
	res := "cato_app_connector.this"

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
					resource.TestCheckResourceAttr(res, "name", cfg.resName),
					resource.TestCheckResourceAttr(res, "description", cfg.resName+" description"),
					resource.TestCheckResourceAttr(res, "group_name", "example-group"),
					resource.TestCheckResourceAttr(res, "type", "VIRTUAL"),

					resource.TestCheckResourceAttr(res, "location.%", "5"),
					resource.TestCheckResourceAttr(res, "location.address", "123 Main St"),
					resource.TestCheckResourceAttr(res, "location.city_name", "Prague"),
					resource.TestCheckResourceAttr(res, "location.country_code", "Cz"),
					resource.TestCheckResourceAttr(res, "location.timezone", "America/New_York"),

					resource.TestCheckResourceAttr(res, "pooled_bandwidth_allocation.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "pooled_bandwidth_allocation.*", map[string]string{
						"license_id": cfg.bwLicenses[0],
						"bandwidth":  "10",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(res, "pooled_bandwidth_allocation.*", map[string]string{
						"license_id": cfg.bwLicenses[1],
						"bandwidth":  "20",
					}),

					resource.TestCheckResourceAttr(res, "preferred_pop_location.%", "4"),
					resource.TestCheckResourceAttr(res, "preferred_pop_location.automatic", "false"),
					resource.TestCheckResourceAttr(res, "preferred_pop_location.preferred_only", "true"),
					resource.TestCheckResourceAttr(res, "preferred_pop_location.primary.%", "2"),
					resource.TestCheckResourceAttr(res, "preferred_pop_location.primary.id", cfg.locations[0].ID),
					resource.TestCheckResourceAttr(res, "preferred_pop_location.primary.name", cfg.locations[0].Name),
					resource.TestCheckResourceAttr(res, "preferred_pop_location.secondary.%", "2"),
					resource.TestCheckResourceAttr(res, "preferred_pop_location.secondary.id", cfg.locations[1].ID),
					resource.TestCheckResourceAttr(res, "preferred_pop_location.secondary.name", cfg.locations[1].Name),

					resource.TestCheckResourceAttr(res, "private_apps.#", "0"),
					resource.TestCheckResourceAttrSet(res, "serial_number"),
					resource.TestCheckResourceAttrSet(res, "socket_id"),
				),
			},
			{
				// Test import mode
				ImportState:  true,
				ResourceName: res,
			},
			{
				// Update the resource - USA state
				Config: cfg.getTfConfig(1),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttrSet(res, "id"),
					resource.TestCheckResourceAttr(res, "name", cfg.resName),
					resource.TestCheckResourceAttr(res, "description", cfg.resName+" description"),
					resource.TestCheckResourceAttr(res, "group_name", "example-group"),
					resource.TestCheckResourceAttr(res, "type", "VIRTUAL"),

					resource.TestCheckResourceAttr(res, "location.%", "5"),
					resource.TestCheckResourceAttr(res, "location.address", "123 Main St"),
					resource.TestCheckResourceAttr(res, "location.city_name", "San Francisco"),
					resource.TestCheckResourceAttr(res, "location.country_code", "US"),
					resource.TestCheckResourceAttr(res, "location.state_code", "US-CA"),
					resource.TestCheckResourceAttr(res, "location.timezone", "America/Los_Angeles"),

					resource.TestCheckResourceAttr(res, "pooled_bandwidth_allocation.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "pooled_bandwidth_allocation.*", map[string]string{
						"license_id": cfg.bwLicenses[0],
						"bandwidth":  "30",
					}),

					resource.TestCheckResourceAttr(res, "preferred_pop_location.%", "4"),
					resource.TestCheckResourceAttr(res, "preferred_pop_location.automatic", "true"),
					resource.TestCheckResourceAttr(res, "preferred_pop_location.preferred_only", "false"),

					resource.TestCheckResourceAttr(res, "private_apps.#", "0"),
					resource.TestCheckResourceAttrSet(res, "serial_number"),
					resource.TestCheckResourceAttrSet(res, "socket_id"),
				),
			},
		},
	})
}

type appConnectorLicenseCfg struct {
	resName    string
	bwLicenses []string
	locations  []acc.Ref
	t          *testing.T
}

func newAppConnectorLicenseCfg(t *testing.T) appConnectorLicenseCfg {
	return appConnectorLicenseCfg{
		resName:    acc.GetRandName("app_connector"),
		bwLicenses: acc.GetBWLicenses(t),
		locations:  acc.GetLocations(t),
		t:          t,
	}
}

func (p appConnectorLicenseCfg) getTfConfig(index int) string {
	tmpl, err := template.New("tmpl").Parse(appConnectorLicenseTFs[index])
	if err != nil {
		p.t.Fatal(err)
	}
	var buf bytes.Buffer
	data := map[string]any{"Name": p.resName, "Location": p.locations, "PooledBWLicenseID": p.bwLicenses}
	if err := tmpl.Execute(&buf, data); err != nil {
		p.t.Fatal(err)
	}

	cfg := acc.ProviderCfg() + buf.String()
	fmt.Println(cfg)
	return cfg
}

var appConnectorLicenseTFs = []string{`
	resource "cato_app_connector" "this" {
		name        = "{{.Name}}"
		description = "{{.Name}} description"
		group_name  = "example-group"
		location = {
			address      = "123 Main St"
			city_name    = "Prague"
			country_code = "Cz"
			timezone     = "America/New_York"
		}
		pooled_bandwidth_allocation = [
			{ license_id = "{{(index .PooledBWLicenseID 0)}}", bandwidth = 10 },
			{ license_id = "{{(index .PooledBWLicenseID 1)}}", bandwidth = 20 },
		]
		preferred_pop_location = {
			automatic      = false
			preferred_only = true
			primary        = { name = "{{(index .Location 0).Name}}" }
			secondary      = { name = "{{(index .Location 1).Name}}" }
		}
		type = "VIRTUAL"
	}`,

	`resource "cato_app_connector" "this" {
		name        = "{{.Name}}"
		description = "{{.Name}} description"
		group_name  = "example-group"
		location = {
			address      = "123 Main St"
			city_name    = "San Francisco"
			country_code = "US"
			state_code   = "US-CA"
			timezone     = "America/Los_Angeles"
		}
		pooled_bandwidth_allocation = [
			{ license_id = "{{(index .PooledBWLicenseID 0)}}", bandwidth = 30 },
		]
		preferred_pop_location = {
			automatic      = true
			preferred_only = false
		}
		type = "VIRTUAL"
	}`,
}
