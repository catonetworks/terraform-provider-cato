//go:build acctest

package wnw_rules_index

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

const envEnableWnwRulesIndexCRUD = "TFACC_ENABLE_RULES_INDEX_CRUD"

func TestAccWnwRulesIndex(t *testing.T) {
	acc.SkipByEnv(t)
	if os.Getenv(envEnableWnwRulesIndexCRUD) != "true" {
		t.Skipf("set %s=true to run bulk WNW rules index acceptance test", envEnableWnwRulesIndexCRUD)
	}
	acc.CleanupFirewallAndWANPolicyRevisions(t)
	defer acc.CleanupFirewallAndWANPolicyRevisions(t)
	mockSrv := accmock.NewMockServer(t, "TestAccWnwRulesIndex")
	defer mockSrv.Close()
	mockSrv.Run()

	cfg := newWnwRulesIndexCfg(t)
	res := "cato_bulk_wnw_move_rule.this"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				Config: cfg.getTfConfig(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "section_data.%", "2"),
					resource.TestCheckResourceAttr(res, "section_data."+cfg.name+"-a.section_name", cfg.name+"-a"),
					resource.TestCheckResourceAttr(res, "section_data."+cfg.name+"-a.section_index", "1"),
					resource.TestCheckResourceAttr(res, "section_data."+cfg.name+"-b.section_name", cfg.name+"-b"),
					resource.TestCheckResourceAttr(res, "section_data."+cfg.name+"-b.section_index", "2"),
				),
			},
			{
				Config: cfg.getTfConfig(1),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "section_data.%", "2"),
					resource.TestCheckResourceAttr(res, "section_data."+cfg.name+"-a.section_index", "2"),
					resource.TestCheckResourceAttr(res, "section_data."+cfg.name+"-b.section_index", "1"),
				),
			},
		},
	})
}

func TestAccWnwRulesIndex_InvalidSectionStartAfterID(t *testing.T) {
	acc.SkipByEnv(t)
	if os.Getenv(envEnableWnwRulesIndexCRUD) != "true" {
		t.Skipf("set %s=true to run bulk WNW rules index acceptance test", envEnableWnwRulesIndexCRUD)
	}
	acc.CleanupFirewallAndWANPolicyRevisions(t)
	defer acc.CleanupFirewallAndWANPolicyRevisions(t)

	mockSrv := accmock.NewMockServer(t, "TestAccWnwRulesIndex_InvalidSectionStartAfterID")
	defer mockSrv.Close()
	mockSrv.Run()

	cfg := newWnwRulesIndexCfg(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				Config:      cfg.getTfConfig(2),
				ExpectError: regexp.MustCompile("(?i)sectiontostartafterid.*not found"),
			},
		},
	})
}

func TestAccWnwRulesIndex_WithRuleData(t *testing.T) {
	acc.SkipByEnv(t)
	if os.Getenv(envEnableWnwRulesIndexCRUD) != "true" {
		t.Skipf("set %s=true to run bulk WNW rules index acceptance test", envEnableWnwRulesIndexCRUD)
	}
	acc.CleanupFirewallAndWANPolicyRevisions(t)
	defer acc.CleanupFirewallAndWANPolicyRevisions(t)

	mockSrv := accmock.NewMockServer(t, "TestAccWnwRulesIndex_WithRuleData")
	defer mockSrv.Close()
	mockSrv.Run()

	cfg := newWnwRulesIndexCfg(t)
	res := "cato_bulk_wnw_move_rule.this"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				Config: cfg.getTfConfig(3),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "section_data.%", "2"),
					resource.TestCheckResourceAttr(res, "rule_data.%", "3"),
					resource.TestCheckResourceAttr(res, "rule_data."+cfg.name+"-r1.section_name", cfg.name+"-a"),
					resource.TestCheckResourceAttr(res, "rule_data."+cfg.name+"-r1.index_in_section", "1"),
					resource.TestCheckResourceAttr(res, "rule_data."+cfg.name+"-r2.section_name", cfg.name+"-a"),
					resource.TestCheckResourceAttr(res, "rule_data."+cfg.name+"-r2.index_in_section", "2"),
					resource.TestCheckResourceAttr(res, "rule_data."+cfg.name+"-r3.section_name", cfg.name+"-b"),
					resource.TestCheckResourceAttr(res, "rule_data."+cfg.name+"-r3.index_in_section", "1"),
				),
			},
			{
				Config: cfg.getTfConfig(4),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "section_data.%", "2"),
					resource.TestCheckResourceAttr(res, "rule_data.%", "3"),
					resource.TestCheckResourceAttr(res, "rule_data."+cfg.name+"-r1.section_name", cfg.name+"-b"),
					resource.TestCheckResourceAttr(res, "rule_data."+cfg.name+"-r1.index_in_section", "1"),
					resource.TestCheckResourceAttr(res, "rule_data."+cfg.name+"-r2.section_name", cfg.name+"-a"),
					resource.TestCheckResourceAttr(res, "rule_data."+cfg.name+"-r2.index_in_section", "1"),
					resource.TestCheckResourceAttr(res, "rule_data."+cfg.name+"-r3.section_name", cfg.name+"-a"),
					resource.TestCheckResourceAttr(res, "rule_data."+cfg.name+"-r3.index_in_section", "2"),
				),
			},
		},
	})
}

type wnwRulesIndexCfg struct {
	name string
	t    *testing.T
}

func newWnwRulesIndexCfg(t *testing.T) wnwRulesIndexCfg {
	return wnwRulesIndexCfg{
		name: acc.GetRandName("wnw_rules_index"),
		t:    t,
	}
}

func (p wnwRulesIndexCfg) getTfConfig(index int) string {
	tmpl, err := template.New("tmpl").Parse(wnwRulesIndexTFs[index])
	if err != nil {
		p.t.Fatal(err)
	}

	var buf bytes.Buffer
	data := map[string]any{"Name": p.name}
	if err := tmpl.Execute(&buf, data); err != nil {
		p.t.Fatal(err)
	}

	cfg := acc.ProviderCfg() + buf.String()
	fmt.Println(cfg)
	return cfg
}

var wnwRulesIndexTFs = []string{
	`resource "cato_wnw_section" "first" {
		at = { position = "LAST_IN_POLICY" }
		section = { name = "{{.Name}}-a" }
	}

	resource "cato_wnw_section" "second" {
		at = { position = "LAST_IN_POLICY" }
		section = { name = "{{.Name}}-b" }
	}

	resource "cato_bulk_wnw_move_rule" "this" {
		section_data = {
			"{{.Name}}-a" = {
				section_name  = cato_wnw_section.first.section.name
				section_index = 1
			}
			"{{.Name}}-b" = {
				section_name  = cato_wnw_section.second.section.name
				section_index = 2
			}
		}
		rule_data = {}
	}
	`,
	`resource "cato_wnw_section" "first" {
		at = { position = "LAST_IN_POLICY" }
		section = { name = "{{.Name}}-a" }
	}

	resource "cato_wnw_section" "second" {
		at = { position = "LAST_IN_POLICY" }
		section = { name = "{{.Name}}-b" }
	}

	resource "cato_bulk_wnw_move_rule" "this" {
		section_data = {
			"{{.Name}}-a" = {
				section_name  = cato_wnw_section.first.section.name
				section_index = 2
			}
			"{{.Name}}-b" = {
				section_name  = cato_wnw_section.second.section.name
				section_index = 1
			}
		}
		rule_data = {}
	}
	`,
	`resource "cato_bulk_wnw_move_rule" "this" {
		section_to_start_after_id = "999999999"
		section_data = {}
	}
	`,
	`resource "cato_wnw_section" "first" {
		at = { position = "LAST_IN_POLICY" }
		section = { name = "{{.Name}}-a" }
	}

	resource "cato_wnw_section" "second" {
		at = { position = "LAST_IN_POLICY" }
		section = { name = "{{.Name}}-b" }
	}

	resource "cato_wnw_rule" "r1" {
		at = { position = "LAST_IN_POLICY" }
		rule = {
			name       = "{{.Name}}-r1"
			enabled    = true
			rule_type  = "WAN"
			route_type = "NONE"
			bandwidth_priority = { name = "255" }
			configuration = {
				active_tcp_acceleration = false
				packet_loss_mitigation  = true
				preserve_source_port    = false
				primary_transport = {
					primary_interface_role   = "WAN1"
					secondary_interface_role = "WAN1"
					transport_type           = "ALTERNATIVE_WAN"
				}
				secondary_transport = {
					primary_interface_role   = "AUTOMATIC"
					secondary_interface_role = "NONE"
					transport_type           = "NONE"
				}
			}
			source      = {}
			destination = {}
			application = {}
		}
	}

	resource "cato_wnw_rule" "r2" {
		at = { position = "LAST_IN_POLICY" }
		rule = {
			name       = "{{.Name}}-r2"
			enabled    = true
			rule_type  = "WAN"
			route_type = "NONE"
			bandwidth_priority = { name = "255" }
			configuration = {
				active_tcp_acceleration = false
				packet_loss_mitigation  = true
				preserve_source_port    = false
				primary_transport = {
					primary_interface_role   = "WAN1"
					secondary_interface_role = "WAN1"
					transport_type           = "ALTERNATIVE_WAN"
				}
				secondary_transport = {
					primary_interface_role   = "AUTOMATIC"
					secondary_interface_role = "NONE"
					transport_type           = "NONE"
				}
			}
			source      = {}
			destination = {}
			application = {}
		}
	}

	resource "cato_wnw_rule" "r3" {
		at = { position = "LAST_IN_POLICY" }
		rule = {
			name       = "{{.Name}}-r3"
			enabled    = true
			rule_type  = "WAN"
			route_type = "NONE"
			bandwidth_priority = { name = "255" }
			configuration = {
				active_tcp_acceleration = false
				packet_loss_mitigation  = true
				preserve_source_port    = false
				primary_transport = {
					primary_interface_role   = "WAN1"
					secondary_interface_role = "WAN1"
					transport_type           = "ALTERNATIVE_WAN"
				}
				secondary_transport = {
					primary_interface_role   = "AUTOMATIC"
					secondary_interface_role = "NONE"
					transport_type           = "NONE"
				}
			}
			source      = {}
			destination = {}
			application = {}
		}
	}

	resource "cato_bulk_wnw_move_rule" "this" {
		section_data = {
			"{{.Name}}-a" = {
				section_name  = cato_wnw_section.first.section.name
				section_index = 1
			}
			"{{.Name}}-b" = {
				section_name  = cato_wnw_section.second.section.name
				section_index = 2
			}
		}
		rule_data = {
			"{{.Name}}-r1" = {
				rule_name        = cato_wnw_rule.r1.rule.name
				section_name     = cato_wnw_section.first.section.name
				index_in_section = 1
				enabled          = true
			}
			"{{.Name}}-r2" = {
				rule_name        = cato_wnw_rule.r2.rule.name
				section_name     = cato_wnw_section.first.section.name
				index_in_section = 2
				enabled          = true
			}
			"{{.Name}}-r3" = {
				rule_name        = cato_wnw_rule.r3.rule.name
				section_name     = cato_wnw_section.second.section.name
				index_in_section = 1
				enabled          = true
			}
		}
	}
	`,
	`resource "cato_wnw_section" "first" {
		at = { position = "LAST_IN_POLICY" }
		section = { name = "{{.Name}}-a" }
	}

	resource "cato_wnw_section" "second" {
		at = { position = "LAST_IN_POLICY" }
		section = { name = "{{.Name}}-b" }
	}

	resource "cato_wnw_rule" "r1" {
		at = { position = "LAST_IN_POLICY" }
		rule = {
			name       = "{{.Name}}-r1"
			enabled    = true
			rule_type  = "WAN"
			route_type = "NONE"
			bandwidth_priority = { name = "255" }
			configuration = {
				active_tcp_acceleration = false
				packet_loss_mitigation  = true
				preserve_source_port    = false
				primary_transport = {
					primary_interface_role   = "WAN1"
					secondary_interface_role = "WAN1"
					transport_type           = "ALTERNATIVE_WAN"
				}
				secondary_transport = {
					primary_interface_role   = "AUTOMATIC"
					secondary_interface_role = "NONE"
					transport_type           = "NONE"
				}
			}
			source      = {}
			destination = {}
			application = {}
		}
	}

	resource "cato_wnw_rule" "r2" {
		at = { position = "LAST_IN_POLICY" }
		rule = {
			name       = "{{.Name}}-r2"
			enabled    = true
			rule_type  = "WAN"
			route_type = "NONE"
			bandwidth_priority = { name = "255" }
			configuration = {
				active_tcp_acceleration = false
				packet_loss_mitigation  = true
				preserve_source_port    = false
				primary_transport = {
					primary_interface_role   = "WAN1"
					secondary_interface_role = "WAN1"
					transport_type           = "ALTERNATIVE_WAN"
				}
				secondary_transport = {
					primary_interface_role   = "AUTOMATIC"
					secondary_interface_role = "NONE"
					transport_type           = "NONE"
				}
			}
			source      = {}
			destination = {}
			application = {}
		}
	}

	resource "cato_wnw_rule" "r3" {
		at = { position = "LAST_IN_POLICY" }
		rule = {
			name       = "{{.Name}}-r3"
			enabled    = true
			rule_type  = "WAN"
			route_type = "NONE"
			bandwidth_priority = { name = "255" }
			configuration = {
				active_tcp_acceleration = false
				packet_loss_mitigation  = true
				preserve_source_port    = false
				primary_transport = {
					primary_interface_role   = "WAN1"
					secondary_interface_role = "WAN1"
					transport_type           = "ALTERNATIVE_WAN"
				}
				secondary_transport = {
					primary_interface_role   = "AUTOMATIC"
					secondary_interface_role = "NONE"
					transport_type           = "NONE"
				}
			}
			source      = {}
			destination = {}
			application = {}
		}
	}

	resource "cato_bulk_wnw_move_rule" "this" {
		section_data = {
			"{{.Name}}-a" = {
				section_name  = cato_wnw_section.first.section.name
				section_index = 1
			}
			"{{.Name}}-b" = {
				section_name  = cato_wnw_section.second.section.name
				section_index = 2
			}
		}
		rule_data = {
			"{{.Name}}-r1" = {
				rule_name        = cato_wnw_rule.r1.rule.name
				section_name     = cato_wnw_section.second.section.name
				index_in_section = 1
				enabled          = true
			}
			"{{.Name}}-r2" = {
				rule_name        = cato_wnw_rule.r2.rule.name
				section_name     = cato_wnw_section.first.section.name
				index_in_section = 1
				enabled          = true
			}
			"{{.Name}}-r3" = {
				rule_name        = cato_wnw_rule.r3.rule.name
				section_name     = cato_wnw_section.first.section.name
				index_in_section = 2
				enabled          = true
			}
		}
	}
	`,
}
