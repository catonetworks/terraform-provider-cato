//go:build acctest

// Package wf_rules_index_with_rule_data holds WAN bulk rules-index tests that create
// multiple rules and reorder via rule_data. Kept separate from wf_rules_index so flaky
// CI (go test -timeout=5m per directory) gives this suite its own timeout budget; see
// internal/acctests/wf_rules_index/wf_rules_index_test.go for section-only cases.
package wf_rules_index_with_rule_data

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccWfRulesIndex_WithRuleData(t *testing.T) {
	acc.SkipByEnv(t)
	acc.CleanupFirewallAndWANPolicyRevisions(t)
	defer acc.CleanupFirewallAndWANPolicyRevisions(t)
	mockSrv := accmock.NewMockServer(t, "TestAccWfRulesIndex_WithRuleData")
	defer mockSrv.Close()
	mockSrv.Run()

	cfg := newWfRulesIndexCfg(t)
	res := "cato_bulk_wf_move_rule.this"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				Config: cfg.getTfConfig(0),
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
				ExpectNonEmptyPlan: true, // cato_wf_rule currently refreshes with exceptions drift.
			},
			{
				Config: cfg.getTfConfig(1),
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
				ExpectNonEmptyPlan: true, // cato_wf_rule currently refreshes with exceptions drift.
			},
		},
	})
}

type wfRulesIndexCfg struct {
	name string
	t    *testing.T
}

func newWfRulesIndexCfg(t *testing.T) wfRulesIndexCfg {
	return wfRulesIndexCfg{
		name: acc.GetRandName("wf_rules_index"),
		t:    t,
	}
}

func (p wfRulesIndexCfg) getTfConfig(step int) string {
	if step < 0 || step >= len(wfRulesIndexWithRuleDataTFs) {
		p.t.Fatalf("invalid tf step %d", step)
	}
	tmpl, err := template.New("tmpl").Parse(wfRulesIndexWithRuleDataTFs[step])
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

// Keep in sync with wf_rules_index_test.go wfRulesIndexTFs[3] and wfRulesIndexTFs[4].
var wfRulesIndexWithRuleDataTFs = []string{
	`resource "cato_wf_section" "first" {
		at = { position = "LAST_IN_POLICY" }
		section = { name = "{{.Name}}-a" }
	}

	resource "cato_wf_section" "second" {
		at = { position = "LAST_IN_POLICY" }
		section = { name = "{{.Name}}-b" }
	}

	resource "cato_wf_rule" "r1" {
		at = {
			position = "FIRST_IN_SECTION"
			ref      = cato_wf_section.first.section.id
		}
		rule = {
			name        = "{{.Name}}-r1"
			enabled     = true
			action      = "ALLOW"
			direction   = "BOTH"
			source      = {}
			destination = {}
			application = {}
			tracking = { event = { enabled = true } }
		}
	}

	resource "cato_wf_rule" "r2" {
		at = {
			position = "AFTER_RULE"
			ref      = cato_wf_rule.r1.rule.id
		}
		rule = {
			name        = "{{.Name}}-r2"
			enabled     = true
			action      = "ALLOW"
			direction   = "BOTH"
			source      = {}
			destination = {}
			application = {}
			tracking = { event = { enabled = true } }
		}
	}

	resource "cato_wf_rule" "r3" {
		at = {
			position = "FIRST_IN_SECTION"
			ref      = cato_wf_section.second.section.id
		}
		rule = {
			name        = "{{.Name}}-r3"
			enabled     = true
			action      = "ALLOW"
			direction   = "BOTH"
			source      = {}
			destination = {}
			application = {}
			tracking = { event = { enabled = true } }
		}
	}

	resource "cato_bulk_wf_move_rule" "this" {
		section_data = {
			"{{.Name}}-a" = {
				section_name  = cato_wf_section.first.section.name
				section_index = 1
			}
			"{{.Name}}-b" = {
				section_name  = cato_wf_section.second.section.name
				section_index = 2
			}
		}
		rule_data = {
			"{{.Name}}-r1" = {
				rule_name        = cato_wf_rule.r1.rule.name
				section_name     = cato_wf_section.first.section.name
				index_in_section = 1
				enabled          = true
			}
			"{{.Name}}-r2" = {
				rule_name        = cato_wf_rule.r2.rule.name
				section_name     = cato_wf_section.first.section.name
				index_in_section = 2
				enabled          = true
			}
			"{{.Name}}-r3" = {
				rule_name        = cato_wf_rule.r3.rule.name
				section_name     = cato_wf_section.second.section.name
				index_in_section = 1
				enabled          = true
			}
		}
	}
	`,
	`resource "cato_wf_section" "first" {
		at = { position = "LAST_IN_POLICY" }
		section = { name = "{{.Name}}-a" }
	}

	resource "cato_wf_section" "second" {
		at = { position = "LAST_IN_POLICY" }
		section = { name = "{{.Name}}-b" }
	}

	resource "cato_wf_rule" "r1" {
		at = {
			position = "FIRST_IN_SECTION"
			ref      = cato_wf_section.first.section.id
		}
		rule = {
			name        = "{{.Name}}-r1"
			enabled     = true
			action      = "ALLOW"
			direction   = "BOTH"
			source      = {}
			destination = {}
			application = {}
			tracking = { event = { enabled = true } }
		}
	}

	resource "cato_wf_rule" "r2" {
		at = {
			position = "AFTER_RULE"
			ref      = cato_wf_rule.r1.rule.id
		}
		rule = {
			name        = "{{.Name}}-r2"
			enabled     = true
			action      = "ALLOW"
			direction   = "BOTH"
			source      = {}
			destination = {}
			application = {}
			tracking = { event = { enabled = true } }
		}
	}

	resource "cato_wf_rule" "r3" {
		at = {
			position = "FIRST_IN_SECTION"
			ref      = cato_wf_section.second.section.id
		}
		rule = {
			name        = "{{.Name}}-r3"
			enabled     = true
			action      = "ALLOW"
			direction   = "BOTH"
			source      = {}
			destination = {}
			application = {}
			tracking = { event = { enabled = true } }
		}
	}

	resource "cato_bulk_wf_move_rule" "this" {
		section_data = {
			"{{.Name}}-a" = {
				section_name  = cato_wf_section.first.section.name
				section_index = 1
			}
			"{{.Name}}-b" = {
				section_name  = cato_wf_section.second.section.name
				section_index = 2
			}
		}
		rule_data = {
			"{{.Name}}-r1" = {
				rule_name        = cato_wf_rule.r1.rule.name
				section_name     = cato_wf_section.second.section.name
				index_in_section = 1
				enabled          = true
			}
			"{{.Name}}-r2" = {
				rule_name        = cato_wf_rule.r2.rule.name
				section_name     = cato_wf_section.first.section.name
				index_in_section = 1
				enabled          = true
			}
			"{{.Name}}-r3" = {
				rule_name        = cato_wf_rule.r3.rule.name
				section_name     = cato_wf_section.first.section.name
				index_in_section = 2
				enabled          = true
			}
		}
	}
	`,
}
