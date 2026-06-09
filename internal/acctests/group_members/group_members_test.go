//go:build acctest

package group_members

import (
	"bytes"
	"fmt"
	"regexp"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccGroupMembers(t *testing.T) {
	acc.SkipByEnv(t)
	mockSrv := accmock.NewMockServer(t, "TestAccGroupMembers")
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newGroupMembersCfg(t)
	res := "cato_group_members.this"

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
					resource.TestCheckResourceAttr(res, "group_name", cfg.groups[0].Name),
					resource.TestCheckResourceAttrSet(res, "id"),
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
				// Update path is currently inconsistent in provider/state correlation.
				// Keep explicit update attempt to track/fail-fast on the known issue.
				Config:      cfg.getTfConfig(1),
				ExpectError: regexp.MustCompile("Provider produced inconsistent result after apply"),
			},
		},
	})
}

type groupMembersCfg struct {
	groups []acc.Ref
	hosts  []acc.Ref
	t      *testing.T
}

func newGroupMembersCfg(t *testing.T) groupMembersCfg {
	return groupMembersCfg{
		groups: acc.GetAdvancedGroups(t),
		hosts:  acc.GetHosts(t),
		t:      t,
	}
}

func (p groupMembersCfg) getTfConfig(index int) string {
	tmpl, err := template.New("tmpl").Parse(groupMembersTFs[index])
	if err != nil {
		p.t.Fatal(err)
	}
	var buf bytes.Buffer
	data := map[string]any{
		"Groups": p.groups,
		"Hosts":  p.hosts,
	}
	if err := tmpl.Execute(&buf, data); err != nil {
		p.t.Fatal(err)
	}

	cfg := acc.ProviderCfg() + buf.String()
	fmt.Println(cfg)
	return cfg
}

var groupMembersTFs = []string{
	`resource "cato_group_members" "this" {
		group_name  = "{{ (index .Groups 0).Name }}"
		members = [
            { type = "HOST", id   = "{{ (index .Hosts 0).ID }}" },
            { type = "HOST", id = "{{ (index .Hosts 1).ID }}" },
		]
	}
	`,
	`resource "cato_group_members" "this" {
		group_name  = "{{ (index .Groups 0).Name }}"
		members = [
            { type = "HOST", id = "{{ (index .Hosts 2).ID }}" },
		]
	}
	`,
}

// TODO: test selection by name, fix TF, fix API { type = "HOST", name = "{{ (index .Hosts 1).Name }}" },

// TODO: test other member types
