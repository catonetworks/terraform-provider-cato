//go:build acctest

package application_control_rule

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

// TestAccApplicationControlRule_Application tests a full create/import/update lifecycle
// for an APPLICATION-type Application Control rule.
func TestAccApplicationControlRule_Application(t *testing.T) {
	acc.SkipByEnv(t)
	mockSrv := accmock.NewMockServer(t, "TestAccApplicationControlRule_Application")
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newRuleCfg(t)
	res := "cato_application_control_rule.application"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				// Create
				Config: cfg.getApplicationConfig(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "at.position", "LAST_IN_POLICY"),
					resource.TestCheckResourceAttrSet(res, "rule.id"),
					resource.TestCheckResourceAttr(res, "rule.name", cfg.resName),
					resource.TestCheckResourceAttr(res, "rule.enabled", "true"),
					resource.TestCheckResourceAttr(res, "rule.rule_type", "APPLICATION"),
					resource.TestCheckResourceAttr(res, "rule.application_rule.action", "ALLOW"),
					resource.TestCheckResourceAttr(res, "rule.application_rule.severity", "LOW"),
					resource.TestCheckResourceAttr(res, "rule.application_rule.tracking.event.enabled", "false"),
				),
			},
			{
				// Import
				ImportState:  true,
				ResourceName: res,
			},
			{
				// Update: change name, action, severity, and tracking
				Config: cfg.getApplicationConfig(1),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "at.position", "LAST_IN_POLICY"),
					resource.TestCheckResourceAttrSet(res, "rule.id"),
					resource.TestCheckResourceAttr(res, "rule.name", cfg.resName+"-2"),
					resource.TestCheckResourceAttr(res, "rule.enabled", "false"),
					resource.TestCheckResourceAttr(res, "rule.rule_type", "APPLICATION"),
					resource.TestCheckResourceAttr(res, "rule.application_rule.action", "BLOCK"),
					resource.TestCheckResourceAttr(res, "rule.application_rule.severity", "HIGH"),
					resource.TestCheckResourceAttr(res, "rule.application_rule.tracking.event.enabled", "true"),
				),
			},
		},
	})
}

// TestAccApplicationControlRule_ApplicationWithSection tests positioning a rule inside
// a section and then referencing it via AFTER_RULE.
func TestAccApplicationControlRule_ApplicationWithSection(t *testing.T) {
	acc.SkipByEnv(t)
	mockSrv := accmock.NewMockServer(t, "TestAccApplicationControlRule_ApplicationWithSection")
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newRuleCfg(t)
	res := "cato_application_control_rule.in_section"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				// Create section + rule inside it
				Config: cfg.getApplicationWithSectionConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "at.position", "LAST_IN_SECTION"),
					resource.TestCheckResourceAttrSet(res, "rule.id"),
					resource.TestCheckResourceAttr(res, "rule.rule_type", "APPLICATION"),
					resource.TestCheckResourceAttr(res, "rule.application_rule.action", "ALLOW"),
				),
			},
		},
	})
}

// TestAccApplicationControlRule_File tests a FILE-type rule with application_activity
// and CONTENT_SIZE file_attribute. FILE rules require action=MONITOR with CONTENT_SIZE.
func TestAccApplicationControlRule_File(t *testing.T) {
	acc.SkipByEnv(t)
	mockSrv := accmock.NewMockServer(t, "TestAccApplicationControlRule_File")
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newRuleCfg(t)
	res := "cato_application_control_rule.file"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				// Create
				Config: cfg.getFileConfig(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "at.position", "LAST_IN_POLICY"),
					resource.TestCheckResourceAttrSet(res, "rule.id"),
					resource.TestCheckResourceAttr(res, "rule.name", cfg.resName+"-file"),
					resource.TestCheckResourceAttr(res, "rule.enabled", "true"),
					resource.TestCheckResourceAttr(res, "rule.rule_type", "FILE"),
					resource.TestCheckResourceAttr(res, "rule.file_rule.action", "MONITOR"),
					resource.TestCheckResourceAttr(res, "rule.file_rule.severity", "MEDIUM"),
					resource.TestCheckResourceAttr(res, "rule.file_rule.application_activity.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.file_rule.application_activity.0.activity.id", "content_transfer_action_upload"),
					resource.TestCheckResourceAttr(res, "rule.file_rule.file_attribute.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.file_rule.file_attribute.0.file_attribute", "CONTENT_SIZE"),
					resource.TestCheckResourceAttr(res, "rule.file_rule.file_attribute.0.operator", "GREATER_THAN"),
					resource.TestCheckResourceAttr(res, "rule.file_rule.file_attribute.0.value", "10485760"),
					resource.TestCheckResourceAttr(res, "rule.file_rule.tracking.event.enabled", "false"),
				),
			},
			{
				// Update: change threshold
				Config: cfg.getFileConfig(1),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "rule.name", cfg.resName+"-file-2"),
					resource.TestCheckResourceAttr(res, "rule.file_rule.file_attribute.0.value", "104857600"),
					resource.TestCheckResourceAttr(res, "rule.file_rule.tracking.event.enabled", "true"),
				),
			},
		},
	})
}

type ruleCfg struct {
	resName string
	t       *testing.T
}

func newRuleCfg(t *testing.T) ruleCfg {
	return ruleCfg{
		resName: acc.GetRandName("ac_rule"),
		t:       t,
	}
}

func (p ruleCfg) render(index int, tpls []string, data map[string]any) string {
	tmpl, err := template.New("tmpl").Parse(tpls[index])
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

func (p ruleCfg) getApplicationConfig(index int) string {
	return p.render(index, applicationRuleTFs, map[string]any{"Name": p.resName})
}

func (p ruleCfg) getApplicationWithSectionConfig() string {
	return p.render(0, applicationWithSectionTFs, map[string]any{"Name": p.resName})
}

func (p ruleCfg) getFileConfig(index int) string {
	return p.render(index, fileRuleTFs, map[string]any{"Name": p.resName})
}

var applicationRuleTFs = []string{
	`resource "cato_application_control_rule" "application" {
  at = {
    position = "LAST_IN_POLICY"
  }
  rule = {
    name      = "{{.Name}}"
    enabled   = true
    rule_type = "APPLICATION"
    application_rule = {
      action   = "ALLOW"
      severity = "LOW"
      application = { application = [{ id = "slack" }] }
      source      = {}
      tracking = {
        event = { enabled = false }
      }
    }
  }
}
`,
	`resource "cato_application_control_rule" "application" {
  at = {
    position = "LAST_IN_POLICY"
  }
  rule = {
    name      = "{{.Name}}-2"
    enabled   = false
    rule_type = "APPLICATION"
    application_rule = {
      action   = "BLOCK"
      severity = "HIGH"
      application = { application = [{ id = "slack" }] }
      source      = {}
      tracking = {
        event = { enabled = true }
      }
    }
  }
}
`,
}

var applicationWithSectionTFs = []string{
	`resource "cato_application_control_section" "sec" {
  at = {
    position = "LAST_IN_POLICY"
  }
  section = {
    name = "{{.Name}}-sec"
  }
}

resource "cato_application_control_rule" "in_section" {
  at = {
    position = "LAST_IN_SECTION"
    ref      = cato_application_control_section.sec.section.id
  }
  rule = {
    name      = "{{.Name}}-in-sec"
    enabled   = true
    rule_type = "APPLICATION"
    application_rule = {
      action   = "ALLOW"
      severity = "LOW"
      application = { application = [{ id = "slack" }] }
      source      = {}
      tracking = {
        event = { enabled = false }
      }
    }
  }
}
`,
}

var fileRuleTFs = []string{
	`resource "cato_application_control_rule" "file" {
  at = {
    position = "LAST_IN_POLICY"
  }
  rule = {
    name      = "{{.Name}}-file"
    enabled   = true
    rule_type = "FILE"
    file_rule = {
      action   = "MONITOR"
      severity = "MEDIUM"
      application = {}
      source      = {}
      application_activity = [
        { activity = { id = "content_transfer_action_upload" } }
      ]
      file_attribute = [
        {
          file_attribute = "CONTENT_SIZE"
          operator       = "GREATER_THAN"
          value          = "10485760"
        }
      ]
      tracking = {
        event = { enabled = false }
      }
    }
  }
}
`,
	`resource "cato_application_control_rule" "file" {
  at = {
    position = "LAST_IN_POLICY"
  }
  rule = {
    name      = "{{.Name}}-file-2"
    enabled   = true
    rule_type = "FILE"
    file_rule = {
      action   = "MONITOR"
      severity = "MEDIUM"
      application = {}
      source      = {}
      application_activity = [
        { activity = { id = "content_transfer_action_upload" } }
      ]
      file_attribute = [
        {
          file_attribute = "CONTENT_SIZE"
          operator       = "GREATER_THAN"
          value          = "104857600"
        }
      ]
      tracking = {
        event = { enabled = true }
      }
    }
  }
}
`,
}
