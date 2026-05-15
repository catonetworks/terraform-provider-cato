//go:build acctest

package wf_section

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccWanFwSection(t *testing.T) {
	mockSrv := accmock.NewMockServer(t, "TestAccWanFwSection")
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newWanFwSectionCfg(t)
	res := "cato_wf_section.this"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				// Create the resource
				Config: cfg.getTfConfig(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "%", "3"),
					resource.TestCheckResourceAttr(res, "at.%", "2"),
					resource.TestCheckResourceAttr(res, "at.position", "LAST_IN_POLICY"),
					resource.TestCheckResourceAttrSet(res, "id"),
					resource.TestCheckResourceAttr(res, "section.%", "2"),
					resource.TestCheckResourceAttrSet(res, "section.id"),
					resource.TestCheckResourceAttr(res, "section.name", cfg.resName),
				),
			},
		},
	})
}

type wanFwSectionCfg struct {
	resName string
	t       *testing.T
}

func newWanFwSectionCfg(t *testing.T) wanFwSectionCfg {
	return wanFwSectionCfg{
		resName: acc.GetRandName("wan_fw_section"),
		t:       t,
	}
}

func (p wanFwSectionCfg) getTfConfig(index int) string {
	tmpl, err := template.New("tmpl").Parse(wanFwSectionTFs[index])
	if err != nil {
		p.t.Fatal(err)
	}
	var buf bytes.Buffer
	data := map[string]any{
		"Name": p.resName,
	}
	if err := tmpl.Execute(&buf, data); err != nil {
		p.t.Fatal(err)
	}

	cfg := acc.ProviderCfg() + buf.String()
	fmt.Println(cfg)
	return cfg
}

var wanFwSectionTFs = []string{
	`resource "cato_wf_section" "this" {
		at = {
			position = "LAST_IN_POLICY"
		}

		section = {
			name = "{{.Name}}"
		}
	}
	`,
}
