//go:build acctest

package if_rules_index

import (
	"bytes"
	"fmt"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccIfRulesIndex(t *testing.T) {
	acc.SkipByEnv(t)
	defer acc.CleanupFirewallAndWANPolicyRevisions(t)
	mockSrv := accmock.NewMockServer(t, "TestAccIfRulesIndex")
	defer mockSrv.Close()
	mockSrv.Run()

	cfg := newIfRulesIndexCfg(t)
	res := "cato_bulk_if_move_rule.this"

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

func TestAccIfRulesIndex_InvalidSectionStartAfterID(t *testing.T) {
	acc.SkipByEnv(t)
	defer acc.CleanupFirewallAndWANPolicyRevisions(t)
	mockSrv := accmock.NewMockServer(t, "TestAccIfRulesIndex_InvalidSectionStartAfterID")
	defer mockSrv.Close()
	mockSrv.Run()

	cfg := newIfRulesIndexCfg(t)

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

// TestAccIfRulesIndex_WithRuleData exercises bulk IF move with rule_data (including cross-section
// ordering). Steps do not use ExpectNonEmptyPlan: cato_if_rule refresh usually matches state
// (WAN wf_rule tests may set ExpectNonEmptyPlan due to known drift).
func TestAccIfRulesIndex_WithRuleData(t *testing.T) {
	acc.SkipByEnv(t)
	defer acc.CleanupFirewallAndWANPolicyRevisions(t)
	mockSrv := accmock.NewMockServer(t, "TestAccIfRulesIndex_WithRuleData")
	defer mockSrv.Close()
	mockSrv.Run()

	cfg := newIfRulesIndexCfg(t)
	res := "cato_bulk_if_move_rule.this"

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

type ifRulesIndexCfg struct {
	name string
	t    *testing.T
}

func newIfRulesIndexCfg(t *testing.T) ifRulesIndexCfg {
	return ifRulesIndexCfg{
		name: acc.GetRandName("if_rules_index"),
		t:    t,
	}
}

func (p ifRulesIndexCfg) getTfConfig(index int) string {
	tmpl, err := template.New("tmpl").Parse(ifRulesIndexTFs[index])
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

var ifRulesIndexTFs = []string{
	`resource "cato_if_section" "first" {
		at = { position = "LAST_IN_POLICY" }
		section = { name = "{{.Name}}-a" }
	}

	resource "cato_if_section" "second" {
		at = { position = "LAST_IN_POLICY" }
		section = { name = "{{.Name}}-b" }
	}

	resource "cato_bulk_if_move_rule" "this" {
		section_data = {
			"{{.Name}}-a" = {
				section_name  = cato_if_section.first.section.name
				section_index = 1
			}
			"{{.Name}}-b" = {
				section_name  = cato_if_section.second.section.name
				section_index = 2
			}
		}
		rule_data = {}
	}
	`,
	`resource "cato_if_section" "first" {
		at = { position = "LAST_IN_POLICY" }
		section = { name = "{{.Name}}-a" }
	}

	resource "cato_if_section" "second" {
		at = { position = "LAST_IN_POLICY" }
		section = { name = "{{.Name}}-b" }
	}

	resource "cato_bulk_if_move_rule" "this" {
		section_data = {
			"{{.Name}}-a" = {
				section_name  = cato_if_section.first.section.name
				section_index = 2
			}
			"{{.Name}}-b" = {
				section_name  = cato_if_section.second.section.name
				section_index = 1
			}
		}
		rule_data = {}
	}
	`,
	`resource "cato_bulk_if_move_rule" "this" {
		section_to_start_after_id = "999999999"
		section_data = {}
	}
	`,
	`resource "cato_if_section" "first" {
		at = { position = "LAST_IN_POLICY" }
		section = { name = "{{.Name}}-a" }
	}

	resource "cato_if_section" "second" {
		at = { position = "LAST_IN_POLICY" }
		section = { name = "{{.Name}}-b" }
	}

	resource "cato_if_rule" "r1" {
		at = {
			position = "FIRST_IN_SECTION"
			ref      = cato_if_section.first.section.id
		}
		rule = {
			name    = "{{.Name}}-r1"
			enabled = true
			action  = "ALLOW"
			tracking = { event = { enabled = true } }
			destination = { domain = ["r1-example.com"] }
			source      = {}
		}
	}

	resource "cato_if_rule" "r2" {
		at = {
			position = "AFTER_RULE"
			ref      = cato_if_rule.r1.rule.id
		}
		rule = {
			name    = "{{.Name}}-r2"
			enabled = true
			action  = "ALLOW"
			tracking = { event = { enabled = true } }
			destination = { domain = ["r2-example.com"] }
			source      = {}
		}
	}

	resource "cato_if_rule" "r3" {
		at = {
			position = "FIRST_IN_SECTION"
			ref      = cato_if_section.second.section.id
		}
		rule = {
			name    = "{{.Name}}-r3"
			enabled = true
			action  = "ALLOW"
			tracking = { event = { enabled = true } }
			destination = { domain = ["r3-example.com"] }
			source      = {}
		}
	}

	resource "cato_bulk_if_move_rule" "this" {
		section_data = {
			"{{.Name}}-a" = {
				section_name  = cato_if_section.first.section.name
				section_index = 1
			}
			"{{.Name}}-b" = {
				section_name  = cato_if_section.second.section.name
				section_index = 2
			}
		}
		rule_data = {
			"{{.Name}}-r1" = {
				rule_name        = cato_if_rule.r1.rule.name
				section_name     = cato_if_section.first.section.name
				index_in_section = 1
				enabled          = true
			}
			"{{.Name}}-r2" = {
				rule_name        = cato_if_rule.r2.rule.name
				section_name     = cato_if_section.first.section.name
				index_in_section = 2
				enabled          = true
			}
			"{{.Name}}-r3" = {
				rule_name        = cato_if_rule.r3.rule.name
				section_name     = cato_if_section.second.section.name
				index_in_section = 1
				enabled          = true
			}
		}
	}
	`,
	`resource "cato_if_section" "first" {
		at = { position = "LAST_IN_POLICY" }
		section = { name = "{{.Name}}-a" }
	}

	resource "cato_if_section" "second" {
		at = { position = "LAST_IN_POLICY" }
		section = { name = "{{.Name}}-b" }
	}

	resource "cato_if_rule" "r1" {
		at = {
			position = "FIRST_IN_SECTION"
			ref      = cato_if_section.first.section.id
		}
		rule = {
			name    = "{{.Name}}-r1"
			enabled = true
			action  = "ALLOW"
			tracking = { event = { enabled = true } }
			destination = { domain = ["r1-example.com"] }
			source      = {}
		}
	}

	resource "cato_if_rule" "r2" {
		at = {
			position = "AFTER_RULE"
			ref      = cato_if_rule.r1.rule.id
		}
		rule = {
			name    = "{{.Name}}-r2"
			enabled = true
			action  = "ALLOW"
			tracking = { event = { enabled = true } }
			destination = { domain = ["r2-example.com"] }
			source      = {}
		}
	}

	resource "cato_if_rule" "r3" {
		at = {
			position = "FIRST_IN_SECTION"
			ref      = cato_if_section.second.section.id
		}
		rule = {
			name    = "{{.Name}}-r3"
			enabled = true
			action  = "ALLOW"
			tracking = { event = { enabled = true } }
			destination = { domain = ["r3-example.com"] }
			source      = {}
		}
	}

	resource "cato_bulk_if_move_rule" "this" {
		section_data = {
			"{{.Name}}-a" = {
				section_name  = cato_if_section.first.section.name
				section_index = 1
			}
			"{{.Name}}-b" = {
				section_name  = cato_if_section.second.section.name
				section_index = 2
			}
		}
		rule_data = {
			"{{.Name}}-r1" = {
				rule_name        = cato_if_rule.r1.rule.name
				section_name     = cato_if_section.second.section.name
				index_in_section = 1
				enabled          = true
			}
			"{{.Name}}-r2" = {
				rule_name        = cato_if_rule.r2.rule.name
				section_name     = cato_if_section.first.section.name
				index_in_section = 1
				enabled          = true
			}
			"{{.Name}}-r3" = {
				rule_name        = cato_if_rule.r3.rule.name
				section_name     = cato_if_section.first.section.name
				index_in_section = 2
				enabled          = true
			}
		}
	}
	`,
}
