//go:build acctest

package application_control_section

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccApplicationControlSection(t *testing.T) {
	acc.SkipByEnv(t)
	mockSrv := accmock.NewMockServer(t, "TestAccApplicationControlSection")
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newSectionCfg(t)
	res := "cato_application_control_section.this"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				// Create
				Config: cfg.getTfConfig(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "at.position", "LAST_IN_POLICY"),
					resource.TestCheckResourceAttrSet(res, "id"),
					resource.TestCheckResourceAttrSet(res, "section.id"),
					resource.TestCheckResourceAttr(res, "section.name", cfg.resName),
				),
			},
			{
				// Import
				ImportState:  true,
				ResourceName: res,
			},
			{
				// Update: rename section
				Config: cfg.getTfConfig(1),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "at.position", "LAST_IN_POLICY"),
					resource.TestCheckResourceAttrSet(res, "section.id"),
					resource.TestCheckResourceAttr(res, "section.name", cfg.resName+"-2"),
				),
			},
		},
	})
}

type sectionCfg struct {
	resName string
	t       *testing.T
}

func newSectionCfg(t *testing.T) sectionCfg {
	return sectionCfg{
		resName: acc.GetRandName("ac_section"),
		t:       t,
	}
}

func (p sectionCfg) getTfConfig(index int) string {
	tmpl, err := template.New("tmpl").Parse(sectionTFs[index])
	if err != nil {
		p.t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, map[string]any{"Name": p.resName}); err != nil {
		p.t.Fatal(err)
	}
	cfg := acc.ProviderCfg() + buf.String()
	fmt.Println(cfg)
	return cfg
}

var sectionTFs = []string{
	`resource "cato_application_control_section" "this" {
  at = {
    position = "LAST_IN_POLICY"
  }
  section = {
    name = "{{.Name}}"
  }
}
`,
	`resource "cato_application_control_section" "this" {
  at = {
    position = "LAST_IN_POLICY"
  }
  section = {
    name = "{{.Name}}-2"
  }
}
`,
}
