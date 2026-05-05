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
	resName string
	hosts   []acc.Ref
	t       *testing.T
}

func newInternetFwCfg(t *testing.T) internetFwCfg {
	return internetFwCfg{
		resName: acc.GetRandName("internet_fw"),
		hosts:   acc.GetHosts(t),
		t:       t,
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
		"Name": p.resName,
	}
	return p.prepareTfCfg(data, internetFwFullTFs[index])
}

var internetFwFullTFs = []string{
	`resource "cato_tls_rule" "kitchen_sink" {
		at   = {
			position = "LAST_IN_POLICY"
		}
		rule = {
			enabled                      = true
			name                         = "{{ .Name }}"
			platform                     = "EMBEDDED"
			action                       = "INSPECT"
			application                  = {
				app_category    = [
					{
						# id   = "advertisements"
						name = "Advertisements"
					},
				]
				application     = [
					{
						# id   = "buildmyteam"
						name = "buildmyteam"
					},
				]
				country         = [
					{
						# id   = "AF"
						name = "Afghanistan"
					},
				]
				custom_app      = [
					{
						# id   = "CustomApp_11362_34188"
						name = "Test Custom App"
					},
				]
				custom_category = [
					{
						# id   = "24255"
						name = "Test Custom Category"
					},
				]
				domain          = [
					"something.com",
					"www.something.com",
				]
				fqdn            = [
					"www.something.com",
				]
				global_ip_range = [
					{
						# id   = "1757826"
						name = "global_ip_range"
					},
				]
				ip              = [
					"1.2.3.4",
				]
				ip_range        = [
					{
						from = "1.2.3.4"
						to   = "1.2.3.5"
					},
				]
				remote_asn      = [
					"1234",
				]
				service         = [
					{
						# id   = "THREEPC"
						name = "3PC"
					},
				]
				subnet          = [
					"1.2.3.0/24",
				]
			}
			connection_origin            = "REMOTE"
			description                  = "test"
			device_posture_profile       = [
				{
					id   = "4202"
					name = "Test Device Posture Profile"
				},
			]
			source                       = {
				global_ip_range     = [
					{
						# id   = "1757826"
						name = "global_ip_range"
					},
					{
						# id   = "1910542"
						name = "global_ip_range2"
					},
				]
				group               = [
					{
						# id   = "623603"
						name = "test group"
					},
				]
				host                = [
					{
						# id   = "1778359"
						name = "host31"
					},
				]
				ip                  = [
					"1.2.3.4",
				]
				ip_range            = [
					{
						from = "1.2.3.4"
						to   = "1.2.3.5"
					},
				]
				network_interface   = [
					{
						id   = "124986"
						# name = "ipsec-dev-site \\ Default" 
						# API does not like \\ charaacters in name values
					},
				]
				site                = [
					{
						# id   = "144905"
						# name = "1600"
					},
				]
				site_network_subnet = [
					{
						id   = "UzU4OTI1Mw=="
						# name = "1600LTE \\ INT_5 \\ Direct Network Range" 
						# API does not like \\ charaacters in name values
					},
				]
				subnet              = [
					"1.2.3.0/24",
				]
				system_group        = [
					{
						# id   = "7S"
						name = "All Floating Ranges"
					},
				]
				user                = [
					{
						# id   = "0"
						name = "test user"
					},
				]
				users_group         = [
					{
						# id   = "500000000"
						name = "Test User Group"
					},
				]
			}
			untrusted_certificate_action = "ALLOW"
		}
	}
	`,
}
