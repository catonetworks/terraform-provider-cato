//go:build acctest

package wf_rules_index

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

// TestAccWfRulesIndex_ReorderTwoRules_VerifiesAPIOrder mirrors the IF reorder API test
// for WAN firewall. WAN rules must be created inside the target section (FIRST_IN_SECTION /
// AFTER_RULE); LAST_IN_POLICY can leave the second rule outside that section in the API
// rules index, which breaks bulk reorder validation. Requires Cato credentials.
func TestAccWfRulesIndex_ReorderTwoRules_VerifiesAPIOrder(t *testing.T) {
	acc.SkipByEnv(t)
	defer acc.CleanupFirewallAndWANPolicyRevisions(t)

	mockSrv := accmock.NewMockServer(t, "TestAccWfRulesIndex_ReorderTwoRules_VerifiesAPIOrder")
	defer mockSrv.Close()
	mockSrv.Run()

	cfg := newWfRulesIndexCfg(t)
	secName := cfg.name + "-sec"
	ruleA := cfg.name + "-a"
	ruleB := cfg.name + "-b"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				Config: cfg.getTwoRuleReorderTF(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.AssertWanRuleNamesOrderInSection(t, secName, ruleA, ruleB),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: cfg.getTwoRuleReorderTF(1),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.AssertWanRuleNamesOrderInSection(t, secName, ruleB, ruleA),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func (p wfRulesIndexCfg) getTwoRuleReorderTF(step int) string {
	tmpl, err := template.New("reorder").Parse(wfTwoRuleReorderTFs[step])
	if err != nil {
		p.t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, map[string]any{"Name": p.name}); err != nil {
		p.t.Fatal(err)
	}
	cfg := acc.ProviderCfg() + buf.String()
	fmt.Println(cfg)
	return cfg
}

var wfTwoRuleReorderTFs = []string{
	`resource "cato_wf_section" "only" {
		at = { position = "LAST_IN_POLICY" }
		section = { name = "{{.Name}}-sec" }
	}

	resource "cato_wf_rule" "a" {
		at = {
			position = "FIRST_IN_SECTION"
			ref      = cato_wf_section.only.section.id
		}
		rule = {
			name        = "{{.Name}}-a"
			enabled     = true
			action      = "ALLOW"
			direction   = "BOTH"
			source      = {}
			destination = {}
			application = {}
			tracking = { event = { enabled = true } }
		}
	}

	resource "cato_wf_rule" "b" {
		at = {
			position = "AFTER_RULE"
			ref      = cato_wf_rule.a.rule.id
		}
		rule = {
			name        = "{{.Name}}-b"
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
			"{{.Name}}-sec" = {
				section_name  = cato_wf_section.only.section.name
				section_index = 1
			}
		}
		rule_data = {
			"{{.Name}}-a" = {
				rule_name        = cato_wf_rule.a.rule.name
				section_name     = cato_wf_section.only.section.name
				index_in_section = 1
				enabled          = true
			}
			"{{.Name}}-b" = {
				rule_name        = cato_wf_rule.b.rule.name
				section_name     = cato_wf_section.only.section.name
				index_in_section = 2
				enabled          = true
			}
		}
	}
	`,
	`resource "cato_wf_section" "only" {
		at = { position = "LAST_IN_POLICY" }
		section = { name = "{{.Name}}-sec" }
	}

	resource "cato_wf_rule" "a" {
		at = {
			position = "FIRST_IN_SECTION"
			ref      = cato_wf_section.only.section.id
		}
		rule = {
			name        = "{{.Name}}-a"
			enabled     = true
			action      = "ALLOW"
			direction   = "BOTH"
			source      = {}
			destination = {}
			application = {}
			tracking = { event = { enabled = true } }
		}
	}

	resource "cato_wf_rule" "b" {
		at = {
			position = "AFTER_RULE"
			ref      = cato_wf_rule.a.rule.id
		}
		rule = {
			name        = "{{.Name}}-b"
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
			"{{.Name}}-sec" = {
				section_name  = cato_wf_section.only.section.name
				section_index = 1
			}
		}
		rule_data = {
			"{{.Name}}-a" = {
				rule_name        = cato_wf_rule.a.rule.name
				section_name     = cato_wf_section.only.section.name
				index_in_section = 2
				enabled          = true
			}
			"{{.Name}}-b" = {
				rule_name        = cato_wf_rule.b.rule.name
				section_name     = cato_wf_section.only.section.name
				index_in_section = 1
				enabled          = true
			}
		}
	}
	`,
}
