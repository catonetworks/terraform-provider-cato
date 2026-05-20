//go:build acctest

package group_members

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccGroup(t *testing.T) {
	acc.SkipByEnv(t)
	mockSrv := accmock.NewMockServer(t, "TestAccGroup")
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newGroupCfg(t)
	res := "cato_group.this"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				// Create the resource
				Config: cfg.getTfConfig(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "%", "4"),
					resource.TestCheckResourceAttrSet(res, "id"),
					resource.TestCheckResourceAttr(res, "name", cfg.resName),
					resource.TestCheckResourceAttr(res, "description", cfg.resName+" description"),
					resource.TestCheckResourceAttr(res, "members.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "members.*", map[string]string{
						"id":   cfg.hosts[0].ID,
						"type": "HOST",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(res, "members.*", map[string]string{
						"id":   cfg.hosts[1].ID,
						"type": "HOST",
					}),
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
					resource.TestCheckResourceAttr(res, "%", "4"),
					resource.TestCheckResourceAttrSet(res, "id"),
					resource.TestCheckResourceAttr(res, "name", cfg.resName+" 2"),
					resource.TestCheckResourceAttr(res, "description", cfg.resName+" description 2"),
					resource.TestCheckResourceAttr(res, "members.#", "1"),
					resource.TestCheckResourceAttr(res, "members.0.id", cfg.hosts[1].ID),
					resource.TestCheckResourceAttrSet(res, "members.0.name"),
				),
			},
		},
	})
}

type groupCfg struct {
	resName string
	hosts   []acc.Ref
	t       *testing.T
}

func newGroupCfg(t *testing.T) groupCfg {
	return groupCfg{
		resName: acc.GetRandName("group"),
		hosts:   acc.GetHosts(t),
		t:       t,
	}
}

func (p groupCfg) getTfConfig(index int) string {
	tmpl, err := template.New("tmpl").Parse(groupTFs[index])
	if err != nil {
		p.t.Fatal(err)
	}
	var buf bytes.Buffer
	data := map[string]any{
		"Name":  p.resName,
		"Hosts": p.hosts,
	}
	if err := tmpl.Execute(&buf, data); err != nil {
		p.t.Fatal(err)
	}

	cfg := acc.ProviderCfg() + buf.String()
	fmt.Println(cfg)
	return cfg
}

var groupTFs = []string{
	`resource "cato_group" "this" {
		name         = "{{.Name}}"
		description  = "{{.Name}} description"
		members = [
            { type = "HOST", id   = "{{ (index .Hosts 0).ID }}" },
            { type = "HOST", id = "{{ (index .Hosts 1).ID }}" },
		]
	}
	`,
	// change name, description
	`resource "cato_group" "this" {
		name         = "{{.Name}} 2"
		description  = "{{.Name}} description 2"
		members = [
            { type = "HOST", id = "{{ (index .Hosts 1).ID }}" },
		]
	}
	`,
}

// TODO: test selection by name, fix TF, fix API { type = "HOST", name = "{{ (index .Hosts 1).Name }}" },

// TODO: test other member types
