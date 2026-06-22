//go:build acctest

package wf_rule

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccWanFw_Simple(t *testing.T) {
	acc.SkipByEnv(t)
	acc.CleanupFirewallAndWANPolicyRevisions(t)
	defer acc.CleanupFirewallAndWANPolicyRevisions(t)
	mockSrv := accmock.NewMockServer(t, "TestAccWanFw_Simple")
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newWanFwCfg(t)
	res := "cato_wf_rule.simple"

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
					resource.TestCheckResourceAttr(res, "rule.action", "ALLOW"),
					resource.TestCheckResourceAttr(res, "rule.connection_origin", "ANY"),
					resource.TestCheckResourceAttr(res, "rule.direction", "BOTH"),
					resource.TestCheckResourceAttr(res, "rule.enabled", "true"),
					resource.TestCheckResourceAttrSet(res, "rule.id"),
					resource.TestCheckResourceAttr(res, "rule.name", cfg.resName),
					resource.TestCheckResourceAttr(res, "rule.tracking.alert.enabled", "false"),
					resource.TestCheckResourceAttr(res, "rule.tracking.alert.frequency", "DAILY"),
					resource.TestCheckResourceAttr(res, "rule.tracking.event.enabled", "true"),
				),
				ExpectNonEmptyPlan: true, // TODO: provider reads empty exceptions, but schema disallows configuring an empty set.
			},
			{
				// Update the resource
				Config: cfg.getTfConfigSimple(1),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "at.position", "LAST_IN_POLICY"),
					resource.TestCheckResourceAttr(res, "rule.action", "BLOCK"),
					resource.TestCheckResourceAttr(res, "rule.connection_origin", "ANY"),
					resource.TestCheckResourceAttr(res, "rule.direction", "BOTH"),
					resource.TestCheckResourceAttr(res, "rule.enabled", "true"),
					resource.TestCheckResourceAttrSet(res, "rule.id"),
					resource.TestCheckResourceAttr(res, "rule.name", cfg.resName+"-2"),
					resource.TestCheckResourceAttr(res, "rule.tracking.alert.enabled", "false"),
					resource.TestCheckResourceAttr(res, "rule.tracking.alert.frequency", "DAILY"),
					resource.TestCheckResourceAttr(res, "rule.tracking.event.enabled", "true"),
				),
				ExpectNonEmptyPlan: true, // TODO: provider reads empty exceptions, but schema disallows configuring an empty set.
			},
		},
	})
}

type wanFwCfg struct {
	resName string
	t       *testing.T
}

func newWanFwCfg(t *testing.T) wanFwCfg {
	return wanFwCfg{
		resName: acc.GetRandName("wan_fw"),
		t:       t,
	}
}

func (p wanFwCfg) prepareTfCfg(data map[string]any, tmplText string) string {
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

func (p wanFwCfg) getTfConfigSimple(index int) string {
	data := map[string]any{
		"Name": p.resName,
	}
	return p.prepareTfCfg(data, wanFwSimpleTFs[index])
}

var wanFwSimpleTFs = []string{
	`resource "cato_wf_rule" "simple" {
		at = {
			position = "LAST_IN_POLICY"
		}
		rule = {
			name      = "{{ .Name }}"
			enabled   = true
			action    = "ALLOW"
			direction = "BOTH"
			source      = {}
			destination = {}
			application = {}
			tracking = {
				event = {
					enabled = true
				}
			}
		}
	}
	`,
	`resource "cato_wf_rule" "simple" {
		at = {
			position = "LAST_IN_POLICY"
		}
		rule = {
			name      = "{{ .Name }}-2"
			enabled   = true
			action    = "BLOCK"
			direction = "BOTH"
			source      = {}
			destination = {}
			application = {}
			tracking = {
				event = {
					enabled = true
				}
			}
		}
	}
	`,
}

// TODO: add all attributes as soon as the API is fixed
