//go:build acctest

package socket_lan_section

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccSocketLanSection(t *testing.T) {
	mockSrv := accmock.NewMockServer(t, "TestAccSocketLanSection")
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newSocketLanSectionCfg(t)
	res := "cato_socket_lan_section.this"

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

type socketLanSectionCfg struct {
	resName string
	t       *testing.T
}

func newSocketLanSectionCfg(t *testing.T) socketLanSectionCfg {
	return socketLanSectionCfg{
		resName: acc.GetRandName("socket_lan_section"),
		t:       t,
	}
}

func (p socketLanSectionCfg) getTfConfig(index int) string {
	tmpl, err := template.New("tmpl").Parse(socketLanSectionTFs[index])
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

var socketLanSectionTFs = []string{
	`resource "cato_socket_lan_section" "this" {
		at = {
			position = "LAST_IN_POLICY"
		}

		section = {
			name = "{{.Name}}"
		}
	}
	`,
}
