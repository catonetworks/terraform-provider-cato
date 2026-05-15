//go:build acctest

package tls_rule

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccTlsRule_Simple(t *testing.T) {
	mockSrv := accmock.NewMockServer(t, "TestAccTlsRule_Simple")
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newTlsRuleCfg(t)
	res := "cato_tls_rule.simple"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				// Create the resource
				Config: cfg.getTfConfigSimple(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "at.position", "LAST_IN_POLICY"),
					resource.TestCheckResourceAttrSet(res, "id"),
					resource.TestCheckResourceAttr(res, "rule.action", "INSPECT"),
					resource.TestCheckResourceAttr(res, "rule.connection_origin", "ANY"),
					resource.TestCheckResourceAttr(res, "rule.enabled", "true"),
					resource.TestCheckResourceAttrSet(res, "rule.id"),
					resource.TestCheckResourceAttr(res, "rule.name", cfg.resName),
					resource.TestCheckResourceAttr(res, "rule.untrusted_certificate_action", "ALLOW"),
				),
			},
			{
				// Test import mode
				ImportState:  true,
				ResourceName: res,
			},
			{
				// Update the resource
				Config: cfg.getTfConfigSimple(1),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "at.position", "LAST_IN_POLICY"),
					resource.TestCheckResourceAttrSet(res, "id"),
					resource.TestCheckResourceAttr(res, "rule.action", "INSPECT"),
					resource.TestCheckResourceAttr(res, "rule.connection_origin", "ANY"),
					resource.TestCheckResourceAttr(res, "rule.enabled", "false"),
					resource.TestCheckResourceAttrSet(res, "rule.id"),
					resource.TestCheckResourceAttr(res, "rule.name", cfg.resName+"-2"),
					resource.TestCheckResourceAttr(res, "rule.untrusted_certificate_action", "ALLOW"),
				),
			},
		},
	})
}

type tlsRuleCfg struct {
	resName string
	t       *testing.T
}

func newTlsRuleCfg(t *testing.T) tlsRuleCfg {
	return tlsRuleCfg{
		resName: acc.GetRandName("tls_rule"),
		t:       t,
	}
}

func (p tlsRuleCfg) prepareTfCfg(data map[string]any, tmplText string) string {
	tmpl, err := template.New("tmpl").Parse(tmplText)
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

func (p tlsRuleCfg) getTfConfigSimple(index int) string {
	data := map[string]any{
		"Name": p.resName,
	}
	return p.prepareTfCfg(data, tlsRuleSimpleTFs[index])
}

var tlsRuleSimpleTFs = []string{
	`resource "cato_tls_rule" "simple" {
		at = {
			position = "LAST_IN_POLICY"
		}
		rule = {
			name                         = "{{ .Name }}"
			enabled                      = true
			action                       = "INSPECT"
			connection_origin            = "ANY"
			untrusted_certificate_action = "ALLOW"
			source                       = {}
			application                  = {}
		}
	}
	`,
	`resource "cato_tls_rule" "simple" {
		at = {
			position = "LAST_IN_POLICY"
		}
		rule = {
			name                         = "{{ .Name }}-2"
			enabled                      = false
			action                       = "INSPECT"
			connection_origin            = "ANY"
			untrusted_certificate_action = "ALLOW"
			source                       = {}
			application                  = {}
		}
	}
	`,
}

// TODO: add all attributes as soon as the API is fixed
