//go:build acctest

package tls_rules_index

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

const envEnableTLSRulesIndexCRUD = "TFACC_ENABLE_RULES_INDEX_CRUD"

func TestAccTLSRulesIndex(t *testing.T) {
	acc.SkipByEnv(t)
	if os.Getenv(envEnableTLSRulesIndexCRUD) != "true" {
		t.Skipf("set %s=true to run bulk TLS rules index acceptance test", envEnableTLSRulesIndexCRUD)
	}
	mockSrv := accmock.NewMockServer(t, "TestAccTLSRulesIndex")
	defer mockSrv.Close()
	mockSrv.Run()

	cfg := newTLSRulesIndexCfg(t)
	res := "cato_bulk_tls_move_rule.this"

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

func TestAccTLSRulesIndex_InvalidSectionStartAfterID(t *testing.T) {
	acc.SkipByEnv(t)
	if os.Getenv(envEnableTLSRulesIndexCRUD) != "true" {
		t.Skipf("set %s=true to run bulk TLS rules index acceptance test", envEnableTLSRulesIndexCRUD)
	}

	mockSrv := accmock.NewMockServer(t, "TestAccTLSRulesIndex_InvalidSectionStartAfterID")
	defer mockSrv.Close()
	mockSrv.Run()

	cfg := newTLSRulesIndexCfg(t)

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

type tlsRulesIndexCfg struct {
	name string
	t    *testing.T
}

func newTLSRulesIndexCfg(t *testing.T) tlsRulesIndexCfg {
	return tlsRulesIndexCfg{
		name: acc.GetRandName("tls_rules_index"),
		t:    t,
	}
}

func (p tlsRulesIndexCfg) getTfConfig(index int) string {
	tmpl, err := template.New("tmpl").Parse(tlsRulesIndexTFs[index])
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

var tlsRulesIndexTFs = []string{
	`resource "cato_tls_section" "first" {
		at = { position = "LAST_IN_POLICY" }
		section = { name = "{{.Name}}-a" }
	}

	resource "cato_tls_section" "second" {
		at = { position = "LAST_IN_POLICY" }
		section = { name = "{{.Name}}-b" }
	}

	resource "cato_bulk_tls_move_rule" "this" {
		section_data = {
			"{{.Name}}-a" = {
				section_name  = cato_tls_section.first.section.name
				section_index = 1
			}
			"{{.Name}}-b" = {
				section_name  = cato_tls_section.second.section.name
				section_index = 2
			}
		}
		rule_data = {}
	}
	`,
	`resource "cato_tls_section" "first" {
		at = { position = "LAST_IN_POLICY" }
		section = { name = "{{.Name}}-a" }
	}

	resource "cato_tls_section" "second" {
		at = { position = "LAST_IN_POLICY" }
		section = { name = "{{.Name}}-b" }
	}

	resource "cato_bulk_tls_move_rule" "this" {
		section_data = {
			"{{.Name}}-a" = {
				section_name  = cato_tls_section.first.section.name
				section_index = 2
			}
			"{{.Name}}-b" = {
				section_name  = cato_tls_section.second.section.name
				section_index = 1
			}
		}
		rule_data = {}
	}
	`,
	`resource "cato_bulk_tls_move_rule" "this" {
		section_to_start_after_id = "999999999"
		section_data = {}
	}
	`,
}
