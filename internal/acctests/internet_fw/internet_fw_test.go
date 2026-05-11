//go:build acctest

package internet_fw

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccInternetFw_Simple(t *testing.T) {
	mockSrv := accmock.NewMockServer(t, "TestAccInternetFw_Simple")
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newInternetFwCfg(t)
	res := "cato_if_rule.simple"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				// Create the resource
				Config: cfg.getTfConfigSimple(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "%", "2"),
					resource.TestCheckResourceAttr(res, "at.%", "2"),
					resource.TestCheckResourceAttr(res, "at.position", "LAST_IN_POLICY"),
					resource.TestCheckResourceAttr(res, "rule.action", "ALLOW"),
					resource.TestCheckResourceAttr(res, "rule.active_period.%", "4"),
					resource.TestCheckResourceAttr(res, "rule.active_period.use_effective_from", "false"),
					resource.TestCheckResourceAttr(res, "rule.active_period.use_expires_at", "false"),
					resource.TestCheckResourceAttr(res, "rule.connection_origin", "ANY"),
					resource.TestCheckResourceAttr(res, "rule.destination.domain.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.destination.domain.0", "test.com"),
					resource.TestCheckResourceAttr(res, "rule.enabled", "true"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.#", "0"),
					resource.TestCheckResourceAttrSet(res, "rule.id"),
					resource.TestCheckResourceAttr(res, "rule.name", cfg.resName),
					resource.TestCheckResourceAttr(res, "rule.schedule.%", "3"),
					resource.TestCheckResourceAttr(res, "rule.schedule.active_on", "ALWAYS"),
					resource.TestCheckResourceAttr(res, "rule.tracking.%", "2"),
					resource.TestCheckResourceAttr(res, "rule.tracking.alert.%", "5"),
					resource.TestCheckResourceAttr(res, "rule.tracking.alert.enabled", "false"),
					resource.TestCheckResourceAttr(res, "rule.tracking.alert.frequency", "DAILY"),
					resource.TestCheckResourceAttr(res, "rule.tracking.event.%", "1"),
					resource.TestCheckResourceAttr(res, "rule.tracking.event.enabled", "true"),
				),
			},
			{
				// Update the resource
				Config: cfg.getTfConfigSimple(1),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "%", "2"),
					resource.TestCheckResourceAttr(res, "at.%", "2"),
					resource.TestCheckResourceAttr(res, "at.position", "LAST_IN_POLICY"),
					resource.TestCheckResourceAttr(res, "rule.action", "BLOCK"),
					resource.TestCheckResourceAttr(res, "rule.active_period.%", "4"),
					resource.TestCheckResourceAttr(res, "rule.active_period.use_effective_from", "false"),
					resource.TestCheckResourceAttr(res, "rule.active_period.use_expires_at", "false"),
					resource.TestCheckResourceAttr(res, "rule.connection_origin", "ANY"),
					resource.TestCheckResourceAttr(res, "rule.destination.domain.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.destination.domain.0", "new.test.com"),
					resource.TestCheckResourceAttr(res, "rule.enabled", "false"),
					resource.TestCheckResourceAttr(res, "rule.exceptions.#", "0"),
					resource.TestCheckResourceAttrSet(res, "rule.id"),
					resource.TestCheckResourceAttr(res, "rule.name", cfg.resName+" 2"),
					resource.TestCheckResourceAttr(res, "rule.schedule.%", "3"),
					resource.TestCheckResourceAttr(res, "rule.schedule.active_on", "ALWAYS"),
					resource.TestCheckResourceAttr(res, "rule.tracking.%", "2"),
					resource.TestCheckResourceAttr(res, "rule.tracking.alert.%", "5"),
					resource.TestCheckResourceAttr(res, "rule.tracking.alert.enabled", "false"),
					resource.TestCheckResourceAttr(res, "rule.tracking.alert.frequency", "DAILY"),
					resource.TestCheckResourceAttr(res, "rule.tracking.event.%", "1"),
					resource.TestCheckResourceAttr(res, "rule.tracking.event.enabled", "true"),
				),
			},
		},
	})
}
func TestAccInternetFw_Full(t *testing.T) {
	mockSrv := accmock.NewMockServer(t, "TestAccInternetFw_Full")
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newInternetFwCfg(t)
	res := "cato_if_rule.full"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				// Create the resource
				Config: cfg.getTfConfigFull(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
				),
			},
		},
	})
}

type internetFwCfg struct {
	resName          string
	hosts            []acc.Ref
	sites            []acc.Ref
	globalIPRanges   []acc.Ref
	siteRanges       []acc.Ref
	floatingRanges   []acc.Ref
	interfaces       []acc.Ref
	users            []acc.Ref
	usersGroups      []acc.Ref
	groups           []acc.Ref
	systemGroups     []acc.Ref
	devicePostures   []acc.Ref
	customApps       []acc.Ref
	customCategories []acc.Ref
	t                *testing.T
}

func newInternetFwCfg(t *testing.T) internetFwCfg {
	return internetFwCfg{
		resName:          acc.GetRandName("internet_fw"),
		hosts:            acc.GetHosts(t),
		sites:            acc.GetSites(t),
		globalIPRanges:   acc.GetGlobalIPRanges(t),
		siteRanges:       acc.GetSiteRanges(t),
		floatingRanges:   acc.GetFloatingRanges(t),
		interfaces:       acc.GetInterfaces(t),
		users:            acc.GetUsers(t),
		usersGroups:      acc.GetUserGroups(t),
		groups:           acc.GetAdvancedGroups(t),
		systemGroups:     acc.GetSystemGroups(t),
		devicePostures:   acc.GetDevicePostures(t),
		customApps:       acc.GetCustomApps(t),
		customCategories: acc.GetCustomCategories(t),
		t:                t,
	}
}

func (p internetFwCfg) prepareTfCfg(data map[string]any, tmplText string) string {
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

// ------------------------------------------------------------------
// Simple cato_if_rule configurations
// ------------------------------------------------------------------
func (p internetFwCfg) getTfConfigSimple(index int) string {
	data := map[string]any{
		"Name": p.resName,
	}
	return p.prepareTfCfg(data, internetFwSimpleTFs[index])
}

var internetFwSimpleTFs = []string{
	`resource "cato_if_rule" "simple" {
		at = {
			position = "LAST_IN_POLICY"
		}
		rule = {
			name    = "{{ .Name }}"
			enabled = true
			action  = "ALLOW"
			tracking = {
				event = {
					enabled = true
				}
			}
			destination = {
				domain = [ "test.com" ]
			}
			source = {}
		}
	}
	`,

	`resource "cato_if_rule" "simple" {
		at = {
			position = "LAST_IN_POLICY"
		}
		rule = {
			name    = "{{ .Name }} 2"
			enabled = false
			action  = "BLOCK"
			tracking = {
				event = {
					enabled = true
				}
			}
			destination = {
				domain = [ "new.test.com" ]
			}
			source = {}
		}
	}
	`,
}

// ------------------------------------------------------------------
// Full cato_if_rule configurations
// ------------------------------------------------------------------
func (p internetFwCfg) getTfConfigFull(index int) string {
	data := map[string]any{
		"Name":             p.resName,
		"Hosts":            p.hosts,
		"Sites":            p.sites,
		"GlobalIPRanges":   p.globalIPRanges,
		"SiteRanges":       p.siteRanges,
		"FloatingRanges":   p.floatingRanges,
		"Interfaces":       p.interfaces,
		"Users":            p.users,
		"UserGroups":       p.usersGroups,
		"Groups":           p.groups,
		"SystemGroups":     p.systemGroups,
		"DevicePostures":   p.devicePostures,
		"CustomApps":       p.customApps,
		"CustomCategories": p.customCategories,
	}
	fmt.Printf(".GlobalIPRanges=%#v", data["GlobalIPRanges"])
	return p.prepareTfCfg(data, internetFwFullTFs[index])
}

var internetFwFullTFs = []string{
	`resource "cato_if_rule" "full" {
		at   = {
			position = "LAST_IN_POLICY"
		}
		rule = {
			name                         = "{{ .Name }}"
			description                  = "{{ .Name }} description"
			enabled                      = true
			active_period = {
				effective_from = "2024-01-01T00:00:00Z"
				expires_at	 = "2124-12-31T23:59:59Z"
			}
			action                       = "BLOCK"
			platform                     = "EMBEDDED"
			tracking = {
				event = { enabled = false }
			}
			destination = {
				domain = [ "new.test.com" ]
			}
			source = {
				ip = ["10.99.12.31"]
				host = [
					{ id   = "{{ (index .Hosts 0).ID }}" },
					{ name = "{{ (index .Hosts 1).Name }}" },
				]
				site = [
					{ id   = "{{ (index .Sites 0).ID }}" },
					{ name = "{{ (index .Sites 1).Name }}" },
				]
				subnet = [
					"10.99.12.0/24"
				]	
				ip_range = [
					{ from   = "10.99.12.10", to = "10.99.12.20" },
				]
				global_ip_range = [
					{ id   = "{{ (index .GlobalIPRanges 0).ID }}" },
					{ name = "{{ (index .GlobalIPRanges 1).Name }}" }
				]
				network_interface = [
					{ id   = "{{ (index .Interfaces 0).ID }}" },
				]
				site_network_subnet = [
					{ id   = "{{ (index .SiteRanges 0).ID }}" },
				]
				floating_subnet = [
					{ id   = "{{ (index .FloatingRanges 0).ID }}" },
					{ name = "{{ (index .FloatingRanges 1).Name }}" },
				]
				user = [
					{ id   = "{{ (index .Users 0).ID }}" },
					# { name = "{{ (index .Users 1).Name }}" },
				]
				users_group = [
					{ id   = "{{ (index .UserGroups 0).ID }}" },
					# { name = "{{ (index .UserGroups 1).Name }}" },
				]
				group = [
					{ id   = "{{ (index .Groups 0).ID }}" },
					{ name = "{{ (index .Groups 1).Name }}" },
				]
				system_group = [
					{ id   = "{{ (index .SystemGroups 0).ID }}" },
					{ name = "{{ (index .SystemGroups 1).Name }}" },
				]
			}
			connection_origin = "SITE"
			country = [
				{ id   = "US" },
				{ name = "France" },
			]
			device = [
				{ id   = "{{ (index .DevicePostures 0).ID }}" },
				{ name = "{{ (index .DevicePostures 1).Name }}" },
			]
			device_os = [
				"WINDOWS",
				"MACOS",
			]
			device_attributes = {
				category     = [
					"IoT",
					"Mobile",
				]
				type         = [
					"Appliance",
					"Analog Telephone Adapter",
				]
				model        = [
					" 9",
					" 7+",
				]
				manufacturer = [
					"ADTRAN",
					"ACTi",
				]
				os = [
					"Aruba OS",
					"Arch Linux",
				]
				os_version = [
					"10.0"
				]
			}
			destination = {
				application = [
					{ name = "Gmail" },
					{ id   = "zoom" },
				]
				custom_app = [
					{ id   = "{{ (index .CustomApps 0).ID }}" },
					{ name = "{{ (index .CustomApps 1).Name }}" },
				]
				app_category = [
					{ id   = "business_systems" },
					{ name = "Advertisements" },
				]
				custom_category = [
					{ id   = "{{ (index .CustomCategories 0).ID }}" },
					{ name = "{{ (index .CustomCategories 1).Name }}" },
				]
			}
		}
	}
	`,
}

// TODO: fix source.network_interface.name  ("aws-site \ LAN 01" does not work)
// TODO: fix source.site_network_subnet.name
// TODO: fix API bug when source.user has both id and name
// TODO: fix API bug when source.users_group has both id and name, there are also duplicate names
