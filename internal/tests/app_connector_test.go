package tests

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAppConnector(t *testing.T) {
	cfg := newAppConnectorCfg(t)
	res := "cato_app_connector.this"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 checkCMAVars(t),
		Steps: []resource.TestStep{
			{
				// Create the resource
				Config: cfg.getTfConfig(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					printAttributes(res),
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
				// Update the resource
				Config: cfg.getTfConfig(1),
				Check: resource.ComposeAggregateTestCheckFunc(
					printAttributes(res),
					resource.TestCheckResourceAttrSet(res, "id"),
					resource.TestCheckResourceAttr(res, "name", cfg.resName+" new"),
					resource.TestCheckResourceAttr(res, "description", cfg.resName+" description new"),
					resource.TestCheckResourceAttr(res, "group_name", "example-group-new"),
					resource.TestCheckResourceAttr(res, "type", "VIRTUAL"),

					resource.TestCheckResourceAttr(res, "location.%", "5"),
					resource.TestCheckResourceAttr(res, "location.address", "123 Main St new"),
					resource.TestCheckResourceAttr(res, "location.city_name", "London"),
					resource.TestCheckResourceAttr(res, "location.country_code", "GB"),
					resource.TestCheckResourceAttr(res, "location.timezone", "Europe/London"),

					resource.TestCheckResourceAttr(res, "preferred_pop_location.%", "4"),
					resource.TestCheckResourceAttr(res, "preferred_pop_location.automatic", "false"),
					resource.TestCheckResourceAttr(res, "preferred_pop_location.preferred_only", "false"),
					resource.TestCheckResourceAttr(res, "preferred_pop_location.primary.%", "2"),
					resource.TestCheckResourceAttr(res, "preferred_pop_location.primary.id", cfg.locations[3].ID),
					resource.TestCheckResourceAttr(res, "preferred_pop_location.primary.name", cfg.locations[3].Name),
					resource.TestCheckResourceAttr(res, "preferred_pop_location.secondary.%", "2"),
					resource.TestCheckResourceAttr(res, "preferred_pop_location.secondary.id", cfg.locations[0].ID),
					resource.TestCheckResourceAttr(res, "preferred_pop_location.secondary.name", cfg.locations[0].Name),

					resource.TestCheckResourceAttr(res, "private_apps.#", "0"),
					resource.TestCheckResourceAttrSet(res, "serial_number"),
					resource.TestCheckResourceAttrSet(res, "socket_id"),
				),
			},
		},
	})
}

type appConnectorCfg struct {
	resName   string
	locations testLocations
	t         *testing.T
}

func newAppConnectorCfg(t *testing.T) appConnectorCfg {
	return appConnectorCfg{
		resName:   getRandName("private_app"),
		locations: getLocations(t),
		t:         t,
	}
}

func (p appConnectorCfg) getTfConfig(index int) string {
	tmpl, err := template.New("tmpl").Parse(appConnectorTFs[index])
	if err != nil {
		p.t.Fatal(err)
	}
	var buf bytes.Buffer
	data := map[string]any{"Name": p.resName, "Location": p.locations}
	if err := tmpl.Execute(&buf, data); err != nil {
		p.t.Fatal(err)
	}

	cfg := providerCfg() + buf.String()
	fmt.Println(cfg)
	return cfg
}

var appConnectorTFs = []string{`
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
		preferred_pop_location = {
			automatic      = false
			preferred_only = true
			primary        = { name = "{{(index .Location 0).Name}}" }
			secondary      = { name = "{{(index .Location 1).Name}}" }
		}
		type = "VIRTUAL"
	}`,

	`resource "cato_app_connector" "this" {
		name        = "{{.Name}} new"
		description = "{{.Name}} description new"
		group_name  = "example-group-new"
		location = {
			address      = "123 Main St new"
			city_name    = "London"
			country_code = "GB"
			timezone     = "Europe/London"
		}
		preferred_pop_location = {
			automatic      = false
			preferred_only = false
			primary        = { id = "{{(index .Location 3).ID}}" }
			secondary      = { id = "{{(index .Location 0).ID}}" }
		}
		type = "VIRTUAL"
	}`,
}
