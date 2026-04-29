//go:build acctest

package acctests

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccSocketSite_Basic tests basic CRUD operations for cato_socket_site resource with connection type SOCKET_X1500 and update to SOCKET_AWS1500
func TestAccSocketSite_Basic(t *testing.T) {
	t.Parallel()
	mockSrv := accmock.NewMockServer(t, "TestAccSocketSite_Basic")
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newsocketSiteCfg(t)
	res := "cato_socket_site.this"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 checkCMAVars(t),
		Steps: []resource.TestStep{
			{
				// Create the resource
				Config: cfg.getTfConfigBasic(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					printAttributes(res),
					resource.TestCheckResourceAttr(res, "%", "7"),
					resource.TestCheckResourceAttr(res, "connection_type", "SOCKET_X1500"),
					resource.TestCheckResourceAttr(res, "description", cfg.resName+" description"),
					resource.TestCheckResourceAttrSet(res, "id"),
					resource.TestCheckResourceAttr(res, "name", cfg.resName),

					resource.TestCheckResourceAttr(res, "native_range.%", "17"),
					resource.TestCheckResourceAttr(res, "native_range.dhcp_settings.%", "5"),
					resource.TestCheckResourceAttr(res, "native_range.dhcp_settings.dhcp_microsegmentation", "false"),
					resource.TestCheckResourceAttr(res, "native_range.dhcp_settings.dhcp_type", "DHCP_RANGE"),
					resource.TestCheckResourceAttr(res, "native_range.dhcp_settings.ip_range", "192.168.20.10-192.168.20.22"),
					resource.TestCheckResourceAttr(res, "native_range.interface_dest_type", "LAN"),
					resource.TestCheckResourceAttrSet(res, "native_range.interface_id"),
					resource.TestCheckResourceAttr(res, "native_range.interface_index", "LAN1"),
					resource.TestCheckResourceAttr(res, "native_range.interface_name", "LAN1"),
					resource.TestCheckResourceAttr(res, "native_range.local_ip", "192.168.20.1"),
					resource.TestCheckResourceAttr(res, "native_range.mdns_reflector", "false"),
					resource.TestCheckResourceAttrSet(res, "native_range.native_network_lan_interface_id"),
					resource.TestCheckResourceAttr(res, "native_range.native_network_range", "192.168.20.0/24"),
					resource.TestCheckResourceAttrSet(res, "native_range.native_network_range_id"),
					resource.TestCheckResourceAttr(res, "native_range.range_name", "Native Range"),
					resource.TestCheckResourceAttr(res, "native_range.range_type", "NATIVE"),
					resource.TestCheckResourceAttr(res, "native_range.translated_subnet", "192.168.20.0/24"),

					resource.TestCheckResourceAttr(res, "site_location.%", "5"),
					resource.TestCheckResourceAttr(res, "site_location.country_code", "FR"),
					resource.TestCheckResourceAttr(res, "site_location.timezone", "Europe/Paris"),
					resource.TestCheckResourceAttr(res, "site_type", "BRANCH"),
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
					printAttributes(res),
					resource.TestCheckResourceAttr(res, "%", "7"),
					resource.TestCheckResourceAttr(res, "connection_type", "SOCKET_AWS1500"),
					resource.TestCheckResourceAttr(res, "description", cfg.resName+" description 2"),
					resource.TestCheckResourceAttrSet(res, "id"),
					resource.TestCheckResourceAttr(res, "name", cfg.resName+" 2"),

					resource.TestCheckResourceAttr(res, "native_range.%", "17"),
					resource.TestCheckResourceAttr(res, "native_range.interface_dest_type", "LAN"),
					resource.TestCheckResourceAttrSet(res, "native_range.interface_id"),
					resource.TestCheckResourceAttr(res, "native_range.interface_index", "LAN1"),
					resource.TestCheckResourceAttr(res, "native_range.interface_name", "LAN1"),
					resource.TestCheckResourceAttr(res, "native_range.local_ip", "192.168.30.4"),
					resource.TestCheckResourceAttr(res, "native_range.mdns_reflector", "false"),
					resource.TestCheckResourceAttrSet(res, "native_range.native_network_lan_interface_id"),
					resource.TestCheckResourceAttr(res, "native_range.native_network_range", "192.168.30.0/24"),
					resource.TestCheckResourceAttrSet(res, "native_range.native_network_range_id"),
					resource.TestCheckResourceAttr(res, "native_range.range_name", "Native Range"),
					resource.TestCheckResourceAttr(res, "native_range.range_type", "NATIVE"),
					resource.TestCheckResourceAttr(res, "native_range.translated_subnet", "192.168.30.0/24"),

					resource.TestCheckResourceAttr(res, "site_location.%", "5"),
					resource.TestCheckResourceAttr(res, "site_location.country_code", "US"),
					resource.TestCheckResourceAttr(res, "site_location.state_code", "US-CA"),
					resource.TestCheckResourceAttr(res, "site_location.timezone", "America/Los_Angeles"),
					resource.TestCheckResourceAttr(res, "site_type", "CLOUD_DC"),
				),
			},
		},
	})
}

// TestAccSocketSite_ConnType tests creating socket sites with different connection types and updating connection type of an existing site
func TestAccSocketSite_ConnType(t *testing.T) {
	t.Parallel()
	mockSrv := accmock.NewMockServer(t, "TestAccSocketSite_ConnType")
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newsocketSiteCfg(t)
	res := "cato_socket_site.this"

	var steps []resource.TestStep
	for i, connType := range cfg.connTypes {
		iface := cfg.defaultIface[i]
		step := resource.TestStep{
			Config: cfg.getTfConfigConnTypes(i),
			Check: resource.ComposeAggregateTestCheckFunc(
				printAttributes(res),
				resource.TestCheckResourceAttr(res, "%", "7"),
				resource.TestCheckResourceAttr(res, "connection_type", connType),
				resource.TestCheckResourceAttr(res, "description", fmt.Sprintf("%s-%d description", cfg.resName, i)),
				resource.TestCheckResourceAttrSet(res, "id"),
				resource.TestCheckResourceAttr(res, "name", fmt.Sprintf("%s-%d", cfg.resName, i)),

				resource.TestCheckResourceAttr(res, "native_range.%", "17"),
				resource.TestCheckResourceAttr(res, "native_range.interface_dest_type", "LAN"),
				resource.TestCheckResourceAttrSet(res, "native_range.interface_id"),
				resource.TestCheckResourceAttr(res, "native_range.interface_index", iface),
				resource.TestCheckResourceAttr(res, "native_range.interface_name", iface),
				resource.TestCheckResourceAttr(res, "native_range.local_ip", "192.168.120.5"),
				resource.TestCheckResourceAttr(res, "native_range.mdns_reflector", "false"),
				resource.TestCheckResourceAttrSet(res, "native_range.native_network_lan_interface_id"),
				resource.TestCheckResourceAttr(res, "native_range.native_network_range", "192.168.120.0/24"),
				resource.TestCheckResourceAttrSet(res, "native_range.native_network_range_id"),
				resource.TestCheckResourceAttr(res, "native_range.range_name", "Native Range"),
				resource.TestCheckResourceAttr(res, "native_range.range_type", "NATIVE"),
				resource.TestCheckResourceAttr(res, "native_range.translated_subnet", "192.168.120.0/24"),

				resource.TestCheckResourceAttr(res, "site_location.%", "5"),
				resource.TestCheckResourceAttr(res, "site_location.country_code", "FR"),
				resource.TestCheckResourceAttr(res, "site_location.timezone", "Europe/Paris"),
				resource.TestCheckResourceAttr(res, "site_type", "BRANCH"),
			),
		}
		steps = append(steps, step)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 checkCMAVars(t),
		Steps:                    steps,
	})
}

// TestAccSocketSite_Location tests creating a socket site with location attributes and updating those attributes
// - test that state_code is properly removed from state when switching from US,US-CA to FR,null
func TestAccSocketSite_Location(t *testing.T) {
	t.Parallel()
	t.Skip("TODO: fix API bug and related TF resource")
	mockSrv := accmock.NewMockServer(t, "TestAccSocketSite_Location")
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newsocketSiteCfg(t)
	res := "cato_socket_site.this"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 checkCMAVars(t),
		Steps: []resource.TestStep{
			{
				// Create the resource
				Config: cfg.getTfConfigLocation(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					printAttributes(res),
					resource.TestCheckResourceAttr(res, "%", "7"),
					resource.TestCheckResourceAttr(res, "connection_type", "SOCKET_X1500"),
					resource.TestCheckResourceAttr(res, "description", cfg.resName+" description"),
					resource.TestCheckResourceAttrSet(res, "id"),
					resource.TestCheckResourceAttr(res, "name", cfg.resName),

					resource.TestCheckResourceAttr(res, "native_range.%", "17"),
					resource.TestCheckResourceAttr(res, "native_range.interface_dest_type", "LAN"),
					resource.TestCheckResourceAttrSet(res, "native_range.interface_id"),
					resource.TestCheckResourceAttr(res, "native_range.interface_index", "LAN1"),
					resource.TestCheckResourceAttr(res, "native_range.interface_name", "LAN1"),
					resource.TestCheckResourceAttr(res, "native_range.local_ip", "192.168.120.5"),
					resource.TestCheckResourceAttr(res, "native_range.mdns_reflector", "false"),
					resource.TestCheckResourceAttrSet(res, "native_range.native_network_lan_interface_id"),
					resource.TestCheckResourceAttr(res, "native_range.native_network_range", "192.168.120.0/24"),
					resource.TestCheckResourceAttrSet(res, "native_range.native_network_range_id"),
					resource.TestCheckResourceAttr(res, "native_range.range_name", "Native Range"),
					resource.TestCheckResourceAttr(res, "native_range.range_type", "NATIVE"),
					resource.TestCheckResourceAttr(res, "native_range.translated_subnet", "192.168.120.0/24"),

					resource.TestCheckResourceAttr(res, "site_location.%", "5"),
					resource.TestCheckResourceAttr(res, "site_location.country_code", "US"),
					resource.TestCheckResourceAttr(res, "site_location.state_code", "US-CA"),
					resource.TestCheckResourceAttr(res, "site_location.timezone", "America/Los_Angeles"),
					resource.TestCheckResourceAttr(res, "site_type", "BRANCH"),
				),
			},
			{
				// Update the resource
				Config: cfg.getTfConfigLocation(1),
				Check: resource.ComposeAggregateTestCheckFunc(
					printAttributes(res),
					resource.TestCheckResourceAttr(res, "name", cfg.resName),
					resource.TestCheckResourceAttr(res, "site_location.%", "5"),
					resource.TestCheckResourceAttr(res, "site_location.country_code", "FR"),
					resource.TestCheckResourceAttr(res, "site_location.state_code", ""),
					resource.TestCheckResourceAttr(res, "site_location.timezone", "Europe/Paris"),
				),
			},
		},
	})
}

// TestAccSocketSite_DHCP tests creating a socket site with DHCP settings and updating those settings
func TestAccSocketSite_DHCP(t *testing.T) {
	t.Parallel()
	mockSrv := accmock.NewMockServer(t, "TestAccSocketSite_DHCP")
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newsocketSiteCfg(t)
	res := "cato_socket_site.this"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 checkCMAVars(t),
		Steps: []resource.TestStep{
			{
				// Create the resource DHCP-RANGE
				Config: cfg.getTfConfigDHCP(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					printAttributes(res),
					resource.TestCheckResourceAttr(res, "%", "7"),
					resource.TestCheckResourceAttr(res, "connection_type", "SOCKET_X1500"),
					resource.TestCheckResourceAttr(res, "description", cfg.resName+" description"),
					resource.TestCheckResourceAttrSet(res, "id"),
					resource.TestCheckResourceAttr(res, "name", cfg.resName),

					resource.TestCheckResourceAttr(res, "native_range.%", "17"),
					resource.TestCheckResourceAttr(res, "native_range.dhcp_settings.%", "5"),
					resource.TestCheckResourceAttr(res, "native_range.dhcp_settings.dhcp_microsegmentation", "false"),
					resource.TestCheckResourceAttr(res, "native_range.dhcp_settings.dhcp_type", "DHCP_RANGE"),
					resource.TestCheckResourceAttr(res, "native_range.dhcp_settings.ip_range", "192.168.140.10-192.168.140.22"),
					resource.TestCheckResourceAttr(res, "native_range.interface_dest_type", "LAN"),
					resource.TestCheckResourceAttrSet(res, "native_range.interface_id"),
					resource.TestCheckResourceAttr(res, "native_range.interface_index", "LAN1"),
					resource.TestCheckResourceAttr(res, "native_range.interface_name", "LAN1"),
					resource.TestCheckResourceAttr(res, "native_range.local_ip", "192.168.140.1"),
					resource.TestCheckResourceAttr(res, "native_range.mdns_reflector", "false"),
					resource.TestCheckResourceAttrSet(res, "native_range.native_network_lan_interface_id"),
					resource.TestCheckResourceAttr(res, "native_range.native_network_range", "192.168.140.0/24"),
					resource.TestCheckResourceAttrSet(res, "native_range.native_network_range_id"),
					resource.TestCheckResourceAttr(res, "native_range.range_name", "Native Range"),
					resource.TestCheckResourceAttr(res, "native_range.range_type", "NATIVE"),
					resource.TestCheckResourceAttr(res, "native_range.translated_subnet", "192.168.140.0/24"),

					resource.TestCheckResourceAttr(res, "site_location.%", "5"),
					resource.TestCheckResourceAttr(res, "site_location.country_code", "FR"),
					resource.TestCheckResourceAttr(res, "site_location.timezone", "Europe/Paris"),
					resource.TestCheckResourceAttr(res, "site_type", "BRANCH"),
				),
			},
			{
				// Create the resource DHCP-RANGE - microsegmentation enabled
				Config: cfg.getTfConfigDHCP(1),
				Check: resource.ComposeAggregateTestCheckFunc(
					printAttributes(res),
					resource.TestCheckResourceAttr(res, "native_range.dhcp_settings.dhcp_microsegmentation", "true"),
					resource.TestCheckResourceAttr(res, "native_range.dhcp_settings.dhcp_type", "DHCP_RANGE"),
				),
			},
			{
				// Update the resource - DHCP-RELAY
				Config: cfg.getTfConfigDHCP(2),
				Check: resource.ComposeAggregateTestCheckFunc(
					printAttributes(res),
					resource.TestCheckResourceAttr(res, "name", cfg.resName),
					resource.TestCheckResourceAttr(res, "native_range.dhcp_settings.dhcp_type", "DHCP_RELAY"),
					resource.TestCheckResourceAttr(res, "native_range.dhcp_settings.relay_group_id", cfg.dhcpRelayGroups[0].ID),
					resource.TestCheckResourceAttr(res, "native_range.dhcp_settings.relay_group_name", cfg.dhcpRelayGroups[0].Name),
					resource.TestCheckResourceAttr(res, "native_range.interface_dest_type", "LAN"),
				),
			},
			{
				// Update the resource - DHCP-DISABLED
				Config: cfg.getTfConfigDHCP(3),
				Check: resource.ComposeAggregateTestCheckFunc(
					printAttributes(res),
					resource.TestCheckResourceAttr(res, "name", cfg.resName),
					resource.TestCheckResourceAttr(res, "native_range.dhcp_settings.dhcp_type", "DHCP_DISABLED"),
				),
			},
		},
	})
}

type socketSiteCfg struct {
	resName         string
	connTypes       []string
	defaultIface    []string
	dhcpRelayGroups []Ref
	t               *testing.T
}

func newsocketSiteCfg(t *testing.T) socketSiteCfg {
	connTypes := []string{
		"SOCKET_AWS1500",
		"SOCKET_AZ1500",
		"SOCKET_ESX1500",
		"SOCKET_GCP1500",
		"SOCKET_X1500",
		"SOCKET_X1600",
		"SOCKET_X1600_LTE",
		"SOCKET_X1700",
	}

	defaultIface := []string{
		"LAN1",
		"LAN1",
		"LAN1",
		"LAN1",
		"LAN1",
		"INT_5",
		"INT_5",
		"INT_3",
	}
	return socketSiteCfg{
		resName:         getRandName("socket_site"),
		connTypes:       connTypes,
		defaultIface:    defaultIface,
		dhcpRelayGroups: getDhcpRelayGroups(t),
		t:               t,
	}
}

func (p socketSiteCfg) prepareTfCfg(data map[string]any, tmplText string) string {
	tmpl, err := template.New("tmpl").Parse(tmplText)
	if err != nil {
		p.t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		p.t.Fatal(err)
	}
	cfg := providerCfg() + buf.String()
	fmt.Println(cfg)
	return cfg
}

// ------------------------------------------------------------------
// Basic cato_socket_site configurations
// ------------------------------------------------------------------
func (p socketSiteCfg) getTfConfigBasic(index int) string {
	data := map[string]any{
		"Name": p.resName,
	}
	return p.prepareTfCfg(data, socketSiteBasicTFs[index])
}

var socketSiteBasicTFs = []string{
	// SOCKET_X1500
	`
	resource "cato_socket_site" "this" {
		name            = "{{.Name}}"
		description     = "{{.Name}} description"
		site_type       = "BRANCH"
		connection_type = "SOCKET_X1500"

		native_range = {
			native_network_range = "192.168.20.0/24"
			local_ip             = "192.168.20.1"
			dhcp_settings = {
				dhcp_type = "DHCP_RANGE"
				ip_range  = "192.168.20.10-192.168.20.22"
			}
		}

		site_location = {
			country_code = "FR"
			timezone     = "Europe/Paris"
		}
	}`,

	// Update site type - SOCKET_AWS1500
	`resource "cato_socket_site" "this" {
		name            = "{{.Name}} 2"
		description     = "{{.Name}} description 2"
		site_type       = "CLOUD_DC"
		connection_type = "SOCKET_AWS1500"

		native_range = {
			native_network_range = "192.168.30.0/24"
			local_ip             = "192.168.30.4"
		}

		site_location = {
			country_code = "US"
			state_code   = "US-CA"
			timezone     = "America/Los_Angeles"
		}
	}`,
}

// ------------------------------------------------------------------
// Connection type cato_socket_site configurations
// ------------------------------------------------------------------
func (p socketSiteCfg) getTfConfigConnTypes(index int) string {
	data := map[string]any{
		"Name":     p.resName,
		"ConnType": p.connTypes[index],
		"Index":    index,
	}
	return p.prepareTfCfg(data, socketSiteConnTypesTFs)
}

// Test SOCKET_AWS1500, SOCKET_AZ1500, SOCKET_ESX1500, SOCKET_GCP1500, SOCKET_X1500, SOCKET_X1600, SOCKET_X1600_LTE, SOCKET_X1700,
var socketSiteConnTypesTFs = `
	resource "cato_socket_site" "this" {
		name            = "{{.Name}}-{{.Index}}"
		description     = "{{.Name}}-{{.Index}} description"
		site_type       = "BRANCH"
		connection_type = "{{.ConnType}}"

		native_range = {
			native_network_range = "192.168.120.0/24"
			local_ip             = "192.168.120.5"
		}

		site_location = {
			country_code = "FR"
			timezone     = "Europe/Paris"
		}
	}`

// ------------------------------------------------------------------
// Location cato_socket_site configurations
// ------------------------------------------------------------------
func (p socketSiteCfg) getTfConfigLocation(index int) string {
	data := map[string]any{
		"Name": p.resName,
	}
	return p.prepareTfCfg(data, socketSiteLocationTFs[index])
}

// Test switch from US,US-CA to FR,null -> should delete state_code
var socketSiteLocationTFs = []string{
	// Set site location to US,US-CA
	`resource "cato_socket_site" "this" {
		name            = "{{.Name}}"
		description     = "{{.Name}} description"
		site_type       = "BRANCH"
		connection_type = "SOCKET_X1500"

		native_range = {
			native_network_range = "192.168.130.0/24"
			local_ip             = "192.168.130.5"
		}

		site_location = {
			country_code = "US"
			state_code   = "US-CA"
			timezone     = "America/Los_Angeles"
		}
	}`,

	// Update site location to FR,null
	`resource "cato_socket_site" "this" {
		name            = "{{.Name}}"
		description     = "{{.Name}} description"
		site_type       = "BRANCH"
		connection_type = "SOCKET_X1500"

		native_range = {
			native_network_range = "192.168.130.0/24"
			local_ip             = "192.168.130.5"
		}

		site_location = {
			country_code = "FR"
			timezone     = "Europe/Paris"
		}
	}`,
}

// ------------------------------------------------------------------
// DHCP cato_socket_site configurations
// ------------------------------------------------------------------
func (p socketSiteCfg) getTfConfigDHCP(index int) string {
	data := map[string]any{
		"Name":        p.resName,
		"RelayGroups": p.dhcpRelayGroups,
	}
	return p.prepareTfCfg(data, socketSiteDhcpTFs[index])
}

// Test switch from US,US-CA to FR,null -> should delete state_code
var socketSiteDhcpTFs = []string{
	// DHCP_RANGE
	`resource "cato_socket_site" "this" {
		name            = "{{.Name}}"
		description     = "{{.Name}} description"
		site_type       = "BRANCH"
		connection_type = "SOCKET_X1500"

		native_range = {
			native_network_range = "192.168.140.0/24"
			local_ip             = "192.168.140.1"
			dhcp_settings = {
				dhcp_type = "DHCP_RANGE"
				ip_range  = "192.168.140.10-192.168.140.22"
			}
		}

		site_location = {
			country_code = "FR"
			timezone     = "Europe/Paris"
		}
	}`,

	// DHCP_RANGE + microsegmentation
	`resource "cato_socket_site" "this" {
		name            = "{{.Name}}"
		description     = "{{.Name}} description"
		site_type       = "BRANCH"
		connection_type = "SOCKET_X1500"

		native_range = {
			native_network_range = "192.168.140.0/24"
			local_ip             = "192.168.140.1"
			dhcp_settings = {
				dhcp_type              = "DHCP_RANGE"
				ip_range               = "192.168.140.10-192.168.140.22"
				dhcp_microsegmentation = true
			}
		}

		site_location = {
			country_code = "FR"
			timezone     = "Europe/Paris"
		}
	}`,

	// DHCP_RELAY
	`resource "cato_socket_site" "this" {
		name            = "{{.Name}}"
		description     = "{{.Name}} description"
		site_type       = "BRANCH"
		connection_type = "SOCKET_X1500"

		native_range = {
			native_network_range = "192.168.140.0/24"
			local_ip             = "192.168.140.1"
			dhcp_settings = {
				dhcp_type = "DHCP_RELAY"
				relay_group_name  = "{{ (index .RelayGroups 0).Name }}"
			}
		}

		site_location = {
			country_code = "FR"
			timezone     = "Europe/Paris"
		}
	}`,

	// DHCP_DISABLED
	`resource "cato_socket_site" "this" {
		name            = "{{.Name}}"
		description     = "{{.Name}} description"
		site_type       = "BRANCH"
		connection_type = "SOCKET_X1500"

		native_range = {
			native_network_range = "192.168.140.0/24"
			local_ip             = "192.168.140.1"
			dhcp_settings = {
				dhcp_type = "DHCP_DISABLED"
			}
		}

		site_location = {
			country_code = "FR"
			timezone     = "Europe/Paris"
		}
	}`,
}
