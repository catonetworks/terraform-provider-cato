//go:build acctest

package wf_bulk_move_rule

import (
	"bytes"
	"fmt"
	"strconv"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccWanFwBulkReorderPolicy(t *testing.T) {
	acc.SkipByEnv(t)
	mockSrv := accmock.NewMockServer(t, "TestAccWanFwBulkReorderPolicy")
	defer mockSrv.Close()
	mockSrv.Run()

	cfg := newWfBulkReorderCfg(t)
	resBulk := "cato_bulk_wf_move_rule.reorder"
	ds := "data.cato_wfRulesIndex.current"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				Config:             cfg.getTfConfig(0),
				ExpectNonEmptyPlan: true, // TODO: provider reads empty exceptions as unknown during apply.
			},
			{
				Config:             cfg.getTfConfig(1),
				ExpectNonEmptyPlan: true, // TODO: provider reads empty exceptions as unknown during apply.
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(resBulk),
					acc.PrintAttributes(ds),
					checkWfRuleOrderInSection(ds, cfg.rule2Name, cfg.sectionName, 1),
					checkWfRuleOrderInSection(ds, cfg.rule1Name, cfg.sectionName, 2),
				),
			},
		},
	})
}

type wfBulkReorderCfg struct {
	sectionName string
	rule1Name   string
	rule2Name   string
	t           *testing.T
}

func newWfBulkReorderCfg(t *testing.T) wfBulkReorderCfg {
	return wfBulkReorderCfg{
		sectionName: acc.GetRandName("wf_reorder_section"),
		rule1Name:   acc.GetRandName("wf_reorder_rule_1"),
		rule2Name:   acc.GetRandName("wf_reorder_rule_2"),
		t:           t,
	}
}

func (c wfBulkReorderCfg) getTfConfig(index int) string {
	tmpl, err := template.New("wf-bulk-reorder").Parse(wfBulkReorderTfs[index])
	if err != nil {
		c.t.Fatal(err)
	}
	var buf bytes.Buffer
	data := map[string]any{
		"SectionName": c.sectionName,
		"Rule1Name":   c.rule1Name,
		"Rule2Name":   c.rule2Name,
	}
	if err := tmpl.Execute(&buf, data); err != nil {
		c.t.Fatal(err)
	}

	cfg := acc.ProviderCfg() + buf.String()
	fmt.Println(cfg)
	return cfg
}

func checkWfRuleOrderInSection(resourceName, ruleName, sectionName string, expectedIndex int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}

		attrs := res.Primary.Attributes
		count, err := strconv.Atoi(attrs["rules.#"])
		if err != nil {
			return fmt.Errorf("unable to parse rules count: %w", err)
		}

		for i := 0; i < count; i++ {
			prefix := fmt.Sprintf("rules.%d.", i)
			if attrs[prefix+"name"] != ruleName {
				continue
			}
			if attrs[prefix+"section_name"] != sectionName {
				return fmt.Errorf("rule %s found in section %s, expected %s", ruleName, attrs[prefix+"section_name"], sectionName)
			}
			got, convErr := strconv.Atoi(attrs[prefix+"index_in_section"])
			if convErr != nil {
				return fmt.Errorf("unable to parse index_in_section for %s: %w", ruleName, convErr)
			}
			if got != expectedIndex {
				return fmt.Errorf("rule %s index_in_section=%d, expected %d", ruleName, got, expectedIndex)
			}
			return nil
		}

		return fmt.Errorf("rule %s not found in %s", ruleName, resourceName)
	}
}

var wfBulkReorderTfs = []string{
	`
resource "cato_wf_section" "reorder_section" {
  at = {
    position = "LAST_IN_POLICY"
  }
  section = {
    name = "{{ .SectionName }}"
  }
}

resource "cato_wf_rule" "rule_1" {
  at = {
    position = "LAST_IN_SECTION"
    ref      = cato_wf_section.reorder_section.section.id
  }
  rule = {
    name      = "{{ .Rule1Name }}"
    enabled   = true
    action    = "ALLOW"
    direction = "BOTH"
    source      = {}
    destination = {}
    application = {}
    tracking = {
      event = {
        enabled = true
      }
    }
  }
}

resource "cato_wf_rule" "rule_2" {
  at = {
    position = "LAST_IN_SECTION"
    ref      = cato_wf_section.reorder_section.section.id
  }
  rule = {
    name      = "{{ .Rule2Name }}"
    enabled   = true
    action    = "ALLOW"
    direction = "BOTH"
    source      = {}
    destination = {}
    application = {}
    tracking = {
      event = {
        enabled = true
      }
    }
  }
}
`,
	`
resource "cato_wf_section" "reorder_section" {
  at = {
    position = "LAST_IN_POLICY"
  }
  section = {
    name = "{{ .SectionName }}"
  }
}

resource "cato_wf_rule" "rule_1" {
  at = {
    position = "LAST_IN_SECTION"
    ref      = cato_wf_section.reorder_section.section.id
  }
  rule = {
    name      = "{{ .Rule1Name }}"
    enabled   = true
    action    = "ALLOW"
    direction = "BOTH"
    source      = {}
    destination = {}
    application = {}
    tracking = {
      event = {
        enabled = true
      }
    }
  }
}

resource "cato_wf_rule" "rule_2" {
  at = {
    position = "LAST_IN_SECTION"
    ref      = cato_wf_section.reorder_section.section.id
  }
  rule = {
    name      = "{{ .Rule2Name }}"
    enabled   = true
    action    = "ALLOW"
    direction = "BOTH"
    source      = {}
    destination = {}
    application = {}
    tracking = {
      event = {
        enabled = true
      }
    }
  }
}

resource "cato_bulk_wf_move_rule" "reorder" {
  section_data = {
    "tf section" = {
      section_index = 1
      section_name  = "tf section"
    }
    (cato_wf_section.reorder_section.section.name) = {
      section_index = 2
      section_name  = cato_wf_section.reorder_section.section.name
    }
  }

  rule_data = {
    (cato_wf_rule.rule_2.rule.name) = {
      index_in_section = 1
      section_name     = cato_wf_section.reorder_section.section.name
      rule_name        = cato_wf_rule.rule_2.rule.name
    }
    (cato_wf_rule.rule_1.rule.name) = {
      index_in_section = 2
      section_name     = cato_wf_section.reorder_section.section.name
      rule_name        = cato_wf_rule.rule_1.rule.name
    }
  }

  depends_on = [cato_wf_rule.rule_1, cato_wf_rule.rule_2]
}

data "cato_wfRulesIndex" "current" {
  depends_on = [cato_bulk_wf_move_rule.reorder]
}
`,
}
