//go:build acctest

package if_sub_policy

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccInternetFwSubPolicy(t *testing.T) {
	acc.SkipByEnv(t)
	mockSrv := accmock.NewMockServer(t, "TestAccInternetFwSubPolicy")
	defer mockSrv.Close()
	mockSrv.Run()

	cfg := newInternetFwSubPolicyCfg(t)
	res := "cato_if_sub_policy.this"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				Config: cfg.getTfConfig(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "%", "2"),
					resource.TestCheckResourceAttr(res, "at.%", "2"),
					resource.TestCheckResourceAttr(res, "at.position", "LAST_IN_POLICY"),
					resource.TestCheckResourceAttr(res, "policy.%", "3"),
					resource.TestCheckResourceAttrSet(res, "policy.id"),
					resource.TestCheckResourceAttr(res, "policy.name", cfg.resName),
					resource.TestCheckResourceAttr(res, "policy.description", cfg.resName+" description"),
				),
			},
			{
				ImportState: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					attrs := state.RootModule().Resources[res].Primary.Attributes
					return attrs["policy.id"], nil
				},
				ResourceName: res,
			},
			{
				Config: cfg.getTfConfig(1),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "%", "2"),
					resource.TestCheckResourceAttr(res, "at.%", "2"),
					resource.TestCheckResourceAttr(res, "at.position", "LAST_IN_POLICY"),
					resource.TestCheckResourceAttr(res, "policy.%", "3"),
					resource.TestCheckResourceAttrSet(res, "policy.id"),
					resource.TestCheckResourceAttr(res, "policy.name", cfg.resName+"-2"),
					resource.TestCheckResourceAttr(res, "policy.description", cfg.resName+" description 2"),
				),
			},
		},
	})
}

type internetFwSubPolicyCfg struct {
	resName string
	t       *testing.T
}

func newInternetFwSubPolicyCfg(t *testing.T) internetFwSubPolicyCfg {
	return internetFwSubPolicyCfg{
		resName: acc.GetRandName("if_sub_policy"),
		t:       t,
	}
}

func (p internetFwSubPolicyCfg) getTfConfig(index int) string {
	tmpl, err := template.New("tmpl").Parse(internetFwSubPolicyTFs[index])
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

var internetFwSubPolicyTFs = []string{
	`resource "cato_if_sub_policy" "this" {
		at = {
			position = "LAST_IN_POLICY"
		}

		policy = {
			name        = "{{ .Name }}"
			description = "{{ .Name }} description"
		}
	}
	`,
	`resource "cato_if_sub_policy" "this" {
		at = {
			position = "LAST_IN_POLICY"
		}

		policy = {
			name        = "{{ .Name }}-2"
			description = "{{ .Name }} description 2"
		}
	}
	`,
}
