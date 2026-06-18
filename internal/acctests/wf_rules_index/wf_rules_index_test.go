//go:build acctest

package wf_rules_index

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

func TestAccWfRulesIndex(t *testing.T) {
	acc.SkipByEnv(t)
	mockSrv := accmock.NewMockServer(t, "TestAccWfRulesIndex")
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

func TestAccWfRulesIndex_InvalidSectionStartAfterID(t *testing.T) {
	acc.SkipByEnv(t)
	mockSrv := accmock.NewMockServer(t, "TestAccWfRulesIndex_InvalidSectionStartAfterID")
	defer mockSrv.Close()
	mockSrv.Run()

	cfg := newWfRulesIndexCfg(t)

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

func TestAccWfRulesIndex_WithRuleData(t *testing.T) {
	acc.SkipByEnv(t)
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
				ExpectNonEmptyPlan: true, // cato_wf_rule currently refreshes with exceptions drift.
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

func (p wfRulesIndexCfg) getTfConfig(index int) string {
	tmpl, err := template.New("tmpl").Parse(wfRulesIndexTFs[index])
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

var wfRulesIndexTFs = []string{
	`resource "cato_wf_section" "first" {
		at = { position = "LAST_IN_POLICY" }
		section = { name = "{{.Name}}-a" }
	}

	resource "cato_wf_section" "second" {
		at = { position = "LAST_IN_POLICY" }
		section = { name = "{{.Name}}-b" }
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
		rule_data = {}
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

	resource "cato_bulk_wf_move_rule" "this" {
		section_data = {
			"{{.Name}}-a" = {
				section_name  = cato_wf_section.first.section.name
				section_index = 2
			}
			"{{.Name}}-b" = {
				section_name  = cato_wf_section.second.section.name
				section_index = 1
			}
		}
		rule_data = {}
	}
	`,
	`resource "cato_bulk_wf_move_rule" "this" {
		section_to_start_after_id = "999999999"
		section_data = {}
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
		at = { position = "LAST_IN_POLICY" }
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
		at = { position = "LAST_IN_POLICY" }
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
		at = { position = "LAST_IN_POLICY" }
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
		at = { position = "LAST_IN_POLICY" }
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
		at = { position = "LAST_IN_POLICY" }
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
		at = { position = "LAST_IN_POLICY" }
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
