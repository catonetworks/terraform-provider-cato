//go:build acctest

package app_tenant_restriction_rule

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

// TestAccAppTenantRestrictionRule_Bypass tests a BYPASS rule lifecycle:
// create, import, and update.
func TestAccAppTenantRestrictionRule_Bypass(t *testing.T) {
	acc.SkipByEnv(t)
	mockSrv := accmock.NewMockServer(t, "TestAccAppTenantRestrictionRule_Bypass")
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newRuleCfg(t)
	res := "cato_app_tenant_restriction_rule.bypass"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				// Create
				Config: cfg.getBypassConfig(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "at.position", "LAST_IN_POLICY"),
					resource.TestCheckResourceAttrSet(res, "rule.id"),
					resource.TestCheckResourceAttr(res, "rule.name", cfg.resName+"-bypass"),
					resource.TestCheckResourceAttr(res, "rule.enabled", "true"),
					resource.TestCheckResourceAttr(res, "rule.action", "BYPASS"),
					resource.TestCheckResourceAttr(res, "rule.severity", "LOW"),
					resource.TestCheckResourceAttr(res, "rule.application.id", "microsoft_office_login"),
				),
			},
			{
				// Import
				ImportState:  true,
				ResourceName: res,
			},
			{
				// Update: rename, change severity
				Config: cfg.getBypassConfig(1),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "rule.name", cfg.resName+"-bypass-2"),
					resource.TestCheckResourceAttr(res, "rule.enabled", "false"),
					resource.TestCheckResourceAttr(res, "rule.action", "BYPASS"),
					resource.TestCheckResourceAttr(res, "rule.severity", "MEDIUM"),
				),
			},
		},
	})
}

// TestAccAppTenantRestrictionRule_InjectHeaders tests an INJECT_HEADERS rule lifecycle.
// Header values are marked sensitive in the provider schema, so ImportStateVerify is skipped.
func TestAccAppTenantRestrictionRule_InjectHeaders(t *testing.T) {
	acc.SkipByEnv(t)
	mockSrv := accmock.NewMockServer(t, "TestAccAppTenantRestrictionRule_InjectHeaders")
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newRuleCfg(t)
	res := "cato_app_tenant_restriction_rule.inject"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				// Create
				Config: cfg.getInjectHeadersConfig(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "at.position", "LAST_IN_POLICY"),
					resource.TestCheckResourceAttrSet(res, "rule.id"),
					resource.TestCheckResourceAttr(res, "rule.name", cfg.resName+"-inject"),
					resource.TestCheckResourceAttr(res, "rule.enabled", "true"),
					resource.TestCheckResourceAttr(res, "rule.action", "INJECT_HEADERS"),
					resource.TestCheckResourceAttr(res, "rule.severity", "HIGH"),
					resource.TestCheckResourceAttr(res, "rule.application.id", "microsoft_office_login"),
					resource.TestCheckResourceAttr(res, "rule.headers.#", "1"),
					resource.TestCheckResourceAttr(res, "rule.headers.0.name", "Restrict-Access-To-Tenants"),
				),
			},
			{
				// Import — ImportStateVerify skipped because header values are sensitive and not returned by API
				ImportState:       true,
				ResourceName:      res,
				ImportStateVerify: false,
			},
			{
				// Update: add second header, update name
				Config: cfg.getInjectHeadersConfig(1),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "rule.name", cfg.resName+"-inject-2"),
					resource.TestCheckResourceAttr(res, "rule.action", "INJECT_HEADERS"),
					resource.TestCheckResourceAttr(res, "rule.headers.#", "2"),
				),
			},
		},
	})
}

// TestAccAppTenantRestrictionRule_WithSection tests placing a BYPASS rule inside a
// named section and verifying the at.position/ref is preserved across plan/apply.
func TestAccAppTenantRestrictionRule_WithSection(t *testing.T) {
	acc.SkipByEnv(t)
	mockSrv := accmock.NewMockServer(t, "TestAccAppTenantRestrictionRule_WithSection")
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newRuleCfg(t)
	res := "cato_app_tenant_restriction_rule.in_section"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				// Create section + rule inside it; verify positioning is preserved on re-plan
				Config: cfg.getWithSectionConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "at.position", "LAST_IN_SECTION"),
					resource.TestCheckResourceAttrSet(res, "rule.id"),
					resource.TestCheckResourceAttr(res, "rule.action", "BYPASS"),
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
		resName: acc.GetRandName("atr_rule"),
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

func (p ruleCfg) getBypassConfig(index int) string {
	return p.render(index, bypassRuleTFs, map[string]any{"Name": p.resName})
}

func (p ruleCfg) getInjectHeadersConfig(index int) string {
	return p.render(index, injectHeadersRuleTFs, map[string]any{"Name": p.resName})
}

func (p ruleCfg) getWithSectionConfig() string {
	return p.render(0, withSectionTFs, map[string]any{"Name": p.resName})
}

var bypassRuleTFs = []string{
	`resource "cato_app_tenant_restriction_rule" "bypass" {
  at = {
    position = "LAST_IN_POLICY"
  }
  rule = {
    name     = "{{.Name}}-bypass"
    enabled  = true
    action   = "BYPASS"
    severity = "LOW"
    application = { id = "microsoft_office_login" }
  }
}
`,
	`resource "cato_app_tenant_restriction_rule" "bypass" {
  at = {
    position = "LAST_IN_POLICY"
  }
  rule = {
    name     = "{{.Name}}-bypass-2"
    enabled  = false
    action   = "BYPASS"
    severity = "MEDIUM"
    application = { id = "microsoft_office_login" }
  }
}
`,
}

var injectHeadersRuleTFs = []string{
	`resource "cato_app_tenant_restriction_rule" "inject" {
  at = {
    position = "LAST_IN_POLICY"
  }
  rule = {
    name     = "{{.Name}}-inject"
    enabled  = true
    action   = "INJECT_HEADERS"
    severity = "HIGH"
    application = { id = "microsoft_office_login" }
    source   = {}
    headers = [
      {
        name  = "Restrict-Access-To-Tenants"
        value = "test-tenant-id"
      }
    ]
  }
}
`,
	`resource "cato_app_tenant_restriction_rule" "inject" {
  at = {
    position = "LAST_IN_POLICY"
  }
  rule = {
    name     = "{{.Name}}-inject-2"
    enabled  = true
    action   = "INJECT_HEADERS"
    severity = "HIGH"
    application = { id = "microsoft_office_login" }
    source   = {}
    headers = [
      {
        name  = "Restrict-Access-To-Tenants"
        value = "test-tenant-id"
      },
      {
        name  = "Restrict-Access-Context"
        value = "test-directory-id"
      }
    ]
  }
}
`,
}

var withSectionTFs = []string{
	`resource "cato_app_tenant_restriction_section" "sec" {
  at = {
    position = "LAST_IN_POLICY"
  }
  section = {
    name = "{{.Name}}-sec"
  }
}

resource "cato_app_tenant_restriction_rule" "in_section" {
  at = {
    position = "LAST_IN_SECTION"
    ref      = cato_app_tenant_restriction_section.sec.section.id
  }
  rule = {
    name     = "{{.Name}}-in-sec"
    enabled  = true
    action   = "BYPASS"
    severity = "LOW"
    application = { id = "microsoft_office_login" }
  }
}
`,
}
