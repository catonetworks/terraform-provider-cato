//go:build acctest

package if_rules_index

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

// TestAccIfRulesIndex_ReorderTwoRules_VerifiesAPIOrder applies a single section with two
// rules, reorders them via cato_bulk_if_move_rule, and asserts order via the Cato API
// (not Terraform state). Rules are created inside the section (FIRST_IN_SECTION /
// AFTER_RULE); LAST_IN_POLICY can leave rules outside that section in the rules index,
// which breaks bulk reorder validation. Unlike the WAN equivalent, steps do not use
// ExpectNonEmptyPlan: cato_if_rule usually refreshes without perpetual drift. Requires
// Cato env vars; for local runs source your env file before go test -tags=acctest.
func TestAccIfRulesIndex_ReorderTwoRules_VerifiesAPIOrder(t *testing.T) {
	acc.SkipByEnv(t)
	acc.CleanupFirewallAndWANPolicyRevisions(t)
	defer acc.CleanupFirewallAndWANPolicyRevisions(t)

	mockSrv := accmock.NewMockServer(t, "TestAccIfRulesIndex_ReorderTwoRules_VerifiesAPIOrder")
	defer mockSrv.Close()
	mockSrv.Run()

	cfg := newIfRulesIndexCfg(t)
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
					acc.AssertIfwRuleNamesOrderInSection(t, secName, ruleA, ruleB),
				),
			},
			{
				Config: cfg.getTwoRuleReorderTF(1),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.AssertIfwRuleNamesOrderInSection(t, secName, ruleB, ruleA),
				),
			},
		},
	})
}

func (p ifRulesIndexCfg) getTwoRuleReorderTF(step int) string {
	tmpl, err := template.New("reorder").Parse(ifTwoRuleReorderTFs[step])
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

var ifTwoRuleReorderTFs = []string{
	`resource "cato_if_section" "only" {
		at = { position = "LAST_IN_POLICY" }
		section = { name = "{{.Name}}-sec" }
	}

	resource "cato_if_rule" "a" {
		at = {
			position = "FIRST_IN_SECTION"
			ref      = cato_if_section.only.section.id
		}
		rule = {
			name    = "{{.Name}}-a"
			enabled = true
			action  = "ALLOW"
			tracking = { event = { enabled = true } }
			destination = { domain = ["if-reorder-a.example"] }
			source      = {}
		}
	}

	resource "cato_if_rule" "b" {
		at = {
			position = "AFTER_RULE"
			ref      = cato_if_rule.a.rule.id
		}
		rule = {
			name    = "{{.Name}}-b"
			enabled = true
			action  = "ALLOW"
			tracking = { event = { enabled = true } }
			destination = { domain = ["if-reorder-b.example"] }
			source      = {}
		}
	}

	resource "cato_bulk_if_move_rule" "this" {
		section_data = {
			"{{.Name}}-sec" = {
				section_name  = cato_if_section.only.section.name
				section_index = 1
			}
		}
		rule_data = {
			"{{.Name}}-a" = {
				rule_name        = cato_if_rule.a.rule.name
				section_name     = cato_if_section.only.section.name
				index_in_section = 1
				enabled          = true
			}
			"{{.Name}}-b" = {
				rule_name        = cato_if_rule.b.rule.name
				section_name     = cato_if_section.only.section.name
				index_in_section = 2
				enabled          = true
			}
		}
	}
	`,
	`resource "cato_if_section" "only" {
		at = { position = "LAST_IN_POLICY" }
		section = { name = "{{.Name}}-sec" }
	}

	resource "cato_if_rule" "a" {
		at = {
			position = "FIRST_IN_SECTION"
			ref      = cato_if_section.only.section.id
		}
		rule = {
			name    = "{{.Name}}-a"
			enabled = true
			action  = "ALLOW"
			tracking = { event = { enabled = true } }
			destination = { domain = ["if-reorder-a.example"] }
			source      = {}
		}
	}

	resource "cato_if_rule" "b" {
		at = {
			position = "AFTER_RULE"
			ref      = cato_if_rule.a.rule.id
		}
		rule = {
			name    = "{{.Name}}-b"
			enabled = true
			action  = "ALLOW"
			tracking = { event = { enabled = true } }
			destination = { domain = ["if-reorder-b.example"] }
			source      = {}
		}
	}

	resource "cato_bulk_if_move_rule" "this" {
		section_data = {
			"{{.Name}}-sec" = {
				section_name  = cato_if_section.only.section.name
				section_index = 1
			}
		}
		rule_data = {
			"{{.Name}}-a" = {
				rule_name        = cato_if_rule.a.rule.name
				section_name     = cato_if_section.only.section.name
				index_in_section = 2
				enabled          = true
			}
			"{{.Name}}-b" = {
				rule_name        = cato_if_rule.b.rule.name
				section_name     = cato_if_section.only.section.name
				index_in_section = 1
				enabled          = true
			}
		}
	}
	`,
}
