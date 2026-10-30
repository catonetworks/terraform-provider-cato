//go:build acctest

package always_on_rule_bulk

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccAlwaysOnRuleBulk(t *testing.T) {
	acc.SkipByEnv(t)
	mockSrv := accmock.NewMockServer(t, t.Name())
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newBulkAlwaysOnCfg()
	res := "cato_always_on_rule_bulk.reorder"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				Config: cfg.getTfConfig(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "section_data.%", "3"),
					resource.TestCheckResourceAttr(res, "section_data.acctest_s1.section_index", "1"),
					resource.TestCheckResourceAttr(res, "section_data.acctest_s2.section_index", "2"),
					resource.TestCheckResourceAttr(res, "section_data.acctest_s3.section_index", "3"),
					resource.TestCheckResourceAttrSet(res, "section_data.acctest_s1.id"),
					resource.TestCheckResourceAttrSet(res, "section_data.acctest_s2.id"),
					resource.TestCheckResourceAttrSet(res, "section_data.acctest_s3.id"),
					resource.TestCheckResourceAttr(res, "rule_data.%", "3"),
					resource.TestCheckResourceAttr(res, "rule_data.acctest_r1.index_in_section", "1"),
					resource.TestCheckResourceAttr(res, "rule_data.acctest_r2.index_in_section", "2"),
					resource.TestCheckResourceAttr(res, "rule_data.acctest_r3.index_in_section", "3"),
					resource.TestCheckResourceAttr(res, "rule_data.acctest_r1.section_name", "acctest_s1"),
					resource.TestCheckResourceAttr(res, "rule_data.acctest_r2.section_name", "acctest_s1"),
					resource.TestCheckResourceAttr(res, "rule_data.acctest_r3.section_name", "acctest_s1"),
					resource.TestCheckResourceAttrSet(res, "rule_data.acctest_r1.id"),
					resource.TestCheckResourceAttrSet(res, "rule_data.acctest_r2.id"),
					resource.TestCheckResourceAttrSet(res, "rule_data.acctest_r3.id"),
				),
			},
			{
				Config: cfg.getTfConfig(1),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "section_data.%", "3"),
					resource.TestCheckResourceAttr(res, "section_data.acctest_s1.section_index", "2"),
					resource.TestCheckResourceAttr(res, "section_data.acctest_s2.section_index", "3"),
					resource.TestCheckResourceAttr(res, "section_data.acctest_s4.section_index", "1"),
					resource.TestCheckResourceAttrSet(res, "section_data.acctest_s1.id"),
					resource.TestCheckResourceAttrSet(res, "section_data.acctest_s2.id"),
					resource.TestCheckResourceAttrSet(res, "section_data.acctest_s4.id"),
					resource.TestCheckResourceAttr(res, "rule_data.%", "3"),
					resource.TestCheckResourceAttr(res, "rule_data.acctest_r1.index_in_section", "2"),
					resource.TestCheckResourceAttr(res, "rule_data.acctest_r2.index_in_section", "1"),
					resource.TestCheckResourceAttr(res, "rule_data.acctest_r4.index_in_section", "3"),
					resource.TestCheckResourceAttr(res, "rule_data.acctest_r1.section_name", "acctest_s4"),
					resource.TestCheckResourceAttr(res, "rule_data.acctest_r2.section_name", "acctest_s4"),
					resource.TestCheckResourceAttr(res, "rule_data.acctest_r4.section_name", "acctest_s4"),
					resource.TestCheckResourceAttrSet(res, "rule_data.acctest_r1.id"),
					resource.TestCheckResourceAttrSet(res, "rule_data.acctest_r2.id"),
					resource.TestCheckResourceAttrSet(res, "rule_data.acctest_r4.id"),
				),
			},
		},
	})
}

type bulkAlwaysOnCfg struct{}

func newBulkAlwaysOnCfg() bulkAlwaysOnCfg { return bulkAlwaysOnCfg{} }

func (cfg bulkAlwaysOnCfg) getTfConfig(index int) string {
	return acc.ProviderCfg() + alwaysOnTFs[index]
}

var alwaysOnTFs = []string{
	`
	# Enable or disable always-on policy
	resource "cato_always_on_policy" "this" {
		enabled = "true"
	}

	resource "cato_always_on_section" "acctest_s1" {
		depends_on = [cato_always_on_policy.this]
		at = { position = "LAST_IN_POLICY" }
		section = { name = "acctest_s1" }
	}

	resource "cato_always_on_section" "acctest_s2" {
		depends_on = [cato_always_on_policy.this]
		at = { position = "LAST_IN_POLICY" }
		section = { name = "acctest_s2" }
	}

	resource "cato_always_on_section" "acctest_s3" {
		depends_on = [cato_always_on_policy.this]
		at = { position = "LAST_IN_POLICY" }
		section = { name = "acctest_s3" }
	}

	resource "cato_always_on_rule" "acctest_r1" {
		depends_on = [cato_always_on_section.acctest_s1]
		at = {
			position = "LAST_IN_SECTION"
			ref      = cato_always_on_section.acctest_s1.section.id
		}
		rule = {
          action = "ENFORCE"
          allow_fail_open = false,
          allow_user_bypass = false,
          anti_tamper_mode = "PROTECT"
          bypass_duration = {
            time = 30
            unit = "MINUTES"
          }
          description = "acctest_r1 description"
          enabled = true,
          name = "acctest_r1",
          source = { }
		}
	}

	resource "cato_always_on_rule" "acctest_r2" {
		depends_on = [cato_always_on_section.acctest_s1]
		at = {
			position = "LAST_IN_SECTION"
			ref      = cato_always_on_section.acctest_s1.section.id
		}
		rule = {
          action = "ENFORCE"
          allow_fail_open = false,
          allow_user_bypass = false,
          anti_tamper_mode = "PROTECT"
          bypass_duration = {
            time = 30
            unit = "MINUTES"
          }
          description = "acctest_r2 description"
          enabled = true,
          name = "acctest_r2",
          source = { }
		}
	}

	resource "cato_always_on_rule" "acctest_r3" {
		depends_on = [cato_always_on_section.acctest_s1]
		at = {
			position = "LAST_IN_SECTION"
			ref      = cato_always_on_section.acctest_s1.section.id
		}
		rule = {
          action = "ENFORCE"
          allow_fail_open = false,
          allow_user_bypass = false,
          anti_tamper_mode = "PROTECT"
          bypass_duration = {
            time = 30
            unit = "MINUTES"
          }
          description = "acctest_r3 description"
          enabled = true,
          name = "acctest_r3",
          source = { }
		}
	}

	# Ordering and publishing
	resource "cato_always_on_rule_bulk" "reorder" {
		depends_on   = [cato_always_on_section.acctest_s1, cato_always_on_rule.acctest_r1]
		section_data = {
			"acctest_s1" = { section_index = 1 }
			"acctest_s2" = { section_index = 2 }
			"acctest_s3" = { section_index = 3 }
		}
		rule_data    = {
			"acctest_r1" = {
				section_name = "acctest_s1"
				index_in_section = 1
			}
			"acctest_r2" = {
				section_name = "acctest_s1"
				index_in_section = 2
			}
			"acctest_r3" = {
				section_name = "acctest_s1"
				index_in_section = 3
			}
		}
	}
	`,

	// add/remove/reorder sections
	// add/remove/reorder rules
	`
	# Enable or disable always-on policy
	resource "cato_always_on_policy" "this" {
		enabled = "true"
	}

	resource "cato_always_on_section" "acctest_s1" {
		depends_on = [cato_always_on_policy.this]
		at = { position = "LAST_IN_POLICY" }
		section = { name = "acctest_s1" }
	}

	resource "cato_always_on_section" "acctest_s2" {
		depends_on = [cato_always_on_policy.this]
		at = { position = "LAST_IN_POLICY" }
		section = { name = "acctest_s2" }
	}

	resource "cato_always_on_section" "acctest_s4" {
		depends_on = [cato_always_on_policy.this]
		at = { position = "LAST_IN_POLICY" }
		section = { name = "acctest_s4" }
	}

	resource "cato_always_on_rule" "acctest_r1" {
		depends_on = [cato_always_on_section.acctest_s4]
		at = {
			position = "LAST_IN_SECTION"
			ref      = cato_always_on_section.acctest_s4.section.id
		}
		rule = {
          action = "ENFORCE"
          allow_fail_open = false,
          allow_user_bypass = false,
          anti_tamper_mode = "PROTECT"
          bypass_duration = {
            time = 30
            unit = "MINUTES"
          }
          description = "acctest_r1 description"
          enabled = true,
          name = "acctest_r1",
          source = { }
		}
	}

	resource "cato_always_on_rule" "acctest_r2" {
		depends_on = [cato_always_on_section.acctest_s1]
		at = {
			position = "LAST_IN_SECTION"
			ref      = cato_always_on_section.acctest_s4.section.id
		}
		rule = {
          action = "ENFORCE"
          allow_fail_open = false,
          allow_user_bypass = false,
          anti_tamper_mode = "PROTECT"
          bypass_duration = {
            time = 30
            unit = "MINUTES"
          }
          description = "acctest_r2 description"
          enabled = true,
          name = "acctest_r2",
          source = { }
		}
	}

	resource "cato_always_on_rule" "acctest_r4" {
		depends_on = [cato_always_on_section.acctest_s1]
		at = {
			position = "LAST_IN_SECTION"
			ref      = cato_always_on_section.acctest_s4.section.id
		}
		rule = {
          action = "ENFORCE"
          allow_fail_open = false,
          allow_user_bypass = false,
          anti_tamper_mode = "PROTECT"
          bypass_duration = {
            time = 30
            unit = "MINUTES"
          }
          description = "acctest_r4 description"
          enabled = true,
          name = "acctest_r4",
          source = { }
		}
	}

	# Ordering and publishing
	resource "cato_always_on_rule_bulk" "reorder" {
		depends_on   = [cato_always_on_section.acctest_s1, cato_always_on_rule.acctest_r1]
		section_data = {
			"acctest_s4" = { section_index = 1 }
			"acctest_s1" = { section_index = 2 }
			"acctest_s2" = { section_index = 3 }
		}
		rule_data    = {
			"acctest_r1" = {
				section_name = "acctest_s4"
				index_in_section = 2
			}
			"acctest_r2" = {
				section_name = "acctest_s4"
				index_in_section = 1
			}
			"acctest_r4" = {
				section_name = "acctest_s4"
				index_in_section = 3
			}
		}
	}
	`,
}
