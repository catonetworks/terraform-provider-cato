//go:build acctest

package wan_network_section

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccWanNetworkSection(t *testing.T) {
	mockSrv := accmock.NewMockServer(t, "TestAccWanNetworkSection")
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newWanNetworkSectionCfg(t)
	res := "cato_wnw_section.this"

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
			{
				// Test import mode
				ImportState:  true,
				ResourceName: res,
			},
			{
				// Update the resource
				Config: cfg.getTfConfig(1),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "%", "3"),
					resource.TestCheckResourceAttr(res, "at.%", "2"),
					resource.TestCheckResourceAttr(res, "at.position", "LAST_IN_POLICY"),
					resource.TestCheckResourceAttrSet(res, "id"),
					resource.TestCheckResourceAttr(res, "section.%", "2"),
					resource.TestCheckResourceAttrSet(res, "section.id"),
					resource.TestCheckResourceAttr(res, "section.name", cfg.resName+"-2"),
				),
			},
		},
	})
}

type wanNetworkSectionCfg struct {
	resName string
	t       *testing.T
}

func newWanNetworkSectionCfg(t *testing.T) wanNetworkSectionCfg {
	return wanNetworkSectionCfg{
		resName: acc.GetRandName("wan_network_section"),
		t:       t,
	}
}

func (p wanNetworkSectionCfg) getTfConfig(index int) string {
	tmpl, err := template.New("tmpl").Parse(wanNetworkSectionTFs[index])
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

var wanNetworkSectionTFs = []string{
	`resource "cato_wnw_section" "this" {
		at = {
			position = "LAST_IN_POLICY"
		}

		section = {
			name = "{{.Name}}"
		}
	}
	`,
	`resource "cato_wnw_section" "this" {
		at = {
			position = "LAST_IN_POLICY"
		}

		section = {
			name = "{{.Name}}-2"
		}
	}
	`,
}
