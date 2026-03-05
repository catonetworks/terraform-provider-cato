package tests

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccPrivAccessPolicyBulk(t *testing.T) {
	// t.Skip("Skipping this test for now")

	cfg := newPrivAccessPolicyBulkCfg(t)
	resRule1 := "cato_private_access_rule.rule_1"
	resRule2 := "cato_private_access_rule.rule_2"
	resRule3 := "cato_private_access_rule.rule_3"
	resRule4 := "cato_private_access_rule.rule_4"
	resRule5 := "cato_private_access_rule.rule_5"
	resBulk := "cato_private_access_rule_bulk.this"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 checkCMAVars(t),
		Steps: []resource.TestStep{
			{
				// Create the resource
				Config:             cfg.getTfConfig(0),
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					printAttributes(resRule1),
					printAttributes(resBulk),
					resource.TestCheckResourceAttr(resRule1, "description", "rule 1"),
					resource.TestCheckResourceAttr(resRule2, "description", "rule 2"),

					resource.TestCheckResourceAttr(resBulk, "rule_data.%", "2"),
					resource.TestCheckResourceAttr(resBulk, "rule_data.rule_1.index", "0"),
					resource.TestCheckResourceAttr(resBulk, "rule_data.rule_2.index", "1"),
				),
			},
			{
				// Test import mode
				ImportState:  true,
				ResourceName: resRule1,
			},
			{
				// Update the resource
				Config:             cfg.getTfConfig(1),
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					printAttributes(resBulk),
					resource.TestCheckResourceAttr(resRule1, "description", "rule 1 new"),
					resource.TestCheckResourceAttr(resRule2, "description", "rule 2"),
					resource.TestCheckResourceAttr(resRule3, "description", "rule 3"),

					resource.TestCheckResourceAttr(resBulk, "rule_data.%", "3"),
					resource.TestCheckResourceAttr(resBulk, "rule_data.rule_3.index", "0"),
					resource.TestCheckResourceAttr(resBulk, "rule_data.rule_2.index", "1"),
					resource.TestCheckResourceAttr(resBulk, "rule_data.rule_1.index", "2"),
				),
			},
			{
				// Update the resource 2
				Config:             cfg.getTfConfig(2),
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					printAttributes(resBulk),
					resource.TestCheckResourceAttr(resRule1, "description", "rule 1"),
					resource.TestCheckResourceAttr(resRule2, "description", "rule 2"),
					resource.TestCheckResourceAttr(resRule3, "description", "rule 3"),
					resource.TestCheckResourceAttr(resRule4, "description", "rule 4"),

					resource.TestCheckResourceAttr(resBulk, "rule_data.%", "4"),
					resource.TestCheckResourceAttr(resBulk, "rule_data.rule_1.index", "0"),
					resource.TestCheckResourceAttr(resBulk, "rule_data.rule_2.index", "1"),
					resource.TestCheckResourceAttr(resBulk, "rule_data.rule_3.index", "2"),
					resource.TestCheckResourceAttr(resBulk, "rule_data.rule_4.index", "3"),
				),
			},
			{
				// Update the resource 3
				Config:             cfg.getTfConfig(3),
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					printAttributes(resBulk),
					resource.TestCheckResourceAttr(resRule2, "description", "rule 2"),
					resource.TestCheckResourceAttr(resRule4, "description", "rule 4"),
					resource.TestCheckResourceAttr(resRule5, "description", "rule 5"),

					resource.TestCheckResourceAttr(resBulk, "rule_data.%", "3"),
					resource.TestCheckResourceAttr(resBulk, "rule_data.rule_2.index", "0"),
					resource.TestCheckResourceAttr(resBulk, "rule_data.rule_4.index", "1"),
					resource.TestCheckResourceAttr(resBulk, "rule_data.rule_5.index", "2"),
				),
			},
		},
		CheckDestroy: func(*terraform.State) error { publishPrivateAccessPolicy(t); return nil },
	})
}

type privAccessPolicyBulkCfg struct {
	users    []ref
	privApps []ref
	t        *testing.T
}

func newPrivAccessPolicyBulkCfg(t *testing.T) privAccessPolicyBulkCfg {
	return privAccessPolicyBulkCfg{
		users:    getUsers(t),
		privApps: getPrivateApps(t),
		t:        t,
	}
}

func (p privAccessPolicyBulkCfg) getTfConfig(index int) string {
	type ruleData struct {
		Name        string
		Description string
	}
	tmpl, err := template.New("tmpl").Parse(privAccessPolicyBulkTFs[0])
	if err != nil {
		p.t.Fatal(err)
	}

	rules := [][]ruleData{
		{ // index 0
			{Name: "rule_1", Description: "rule 1"},
			{Name: "rule_2", Description: "rule 2"},
		},
		{ // index 1
			{Name: "rule_3", Description: "rule 3"},
			{Name: "rule_2", Description: "rule 2"},
			{Name: "rule_1", Description: "rule 1 new"},
		},
		{ // index 2
			{Name: "rule_1", Description: "rule 1"},
			{Name: "rule_2", Description: "rule 2"},
			{Name: "rule_3", Description: "rule 3"},
			{Name: "rule_4", Description: "rule 4"},
		},
		{ // index 3
			{Name: "rule_2", Description: "rule 2"},
			{Name: "rule_4", Description: "rule 4"},
			{Name: "rule_5", Description: "rule 5"},
		},
	}
	var buf bytes.Buffer
	data := map[string]any{
		"AppName":  p.privApps[0].Name,
		"UserName": p.users[0].Name,
		"Rules":    rules[index],
	}
	if err := tmpl.Execute(&buf, data); err != nil {
		p.t.Fatal(err)
	}

	cfg := providerCfg() + buf.String()
	fmt.Println(cfg)
	return cfg
}

var privAccessPolicyBulkTFs = []string{`
    {{ $app := .AppName }}
    {{ $user := .UserName }}
    {{ range .Rules }}
	resource "cato_private_access_rule" "{{ .Name }}" {
	  action = "BLOCK"
	  applications = [ { "name" = "{{ $app }}" } ]
	  description   = "{{ .Description }}"
	  enabled   = true
	  name      = "{{ .Name }}"
	  source    = { users = [ { name = "{{ $user }}" } ] }
	}
	{{ end }}

	resource "cato_private_access_rule_bulk" "this" {
	  depends_on = [{{- range $i, $r := .Rules -}}{{ if $i }}, {{ end }}cato_private_access_rule.{{ $r.Name }}{{- end -}}]
	  rule_data = {
	    {{ range $i, $r := .Rules }}
	    "{{ $r.Name }}" = {
	      index = {{ $i }}
	      name  = "{{ $r.Name }}"
	    }
	    {{ end }}
	  }
	}
	`,
}
