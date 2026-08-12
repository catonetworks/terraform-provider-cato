//go:build acctest

package if_section

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccInternetFwSection(t *testing.T) {
	acc.SkipByEnv(t)
	acc.CleanupFirewallAndWANPolicyRevisions(t)
	defer acc.CleanupFirewallAndWANPolicyRevisions(t)
	mockSrv := accmock.NewMockServer(t, "TestAccInternetFwSection")
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newInternetFwSectionCfg(t)
	res := "cato_if_section.this"

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

func TestAccInternetFwSectionUnderSubPolicy(t *testing.T) {
	acc.SkipByEnv(t)
	acc.CleanupFirewallAndWANPolicyRevisions(t)
	defer acc.CleanupFirewallAndWANPolicyRevisions(t)
	mockSrv := accmock.NewMockServer(t, "TestAccInternetFwSectionUnderSubPolicy")
	defer mockSrv.Close()
	mockSrv.Run()

	name := acc.GetRandName("internet_fw_section_sub_policy")
	section := "cato_if_section.this"
	subPolicy := "cato_if_sub_policy.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				Config: internetFwSectionUnderSubPolicyConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(subPolicy, "id"),
					resource.TestCheckResourceAttrSet(section, "id"),
					resource.TestCheckResourceAttrSet(section, "section.id"),
					resource.TestCheckResourceAttr(section, "section.name", name),
					resource.TestCheckResourceAttr(section, "at.position", "LAST_IN_POLICY"),
					resource.TestCheckResourceAttrPair(section, "at.ref", subPolicy, "id"),
				),
			},
		},
	})
}

type internetFwSectionCfg struct {
	resName string
	t       *testing.T
}

func newInternetFwSectionCfg(t *testing.T) internetFwSectionCfg {
	return internetFwSectionCfg{
		resName: acc.GetRandName("internet_fw_section"),
		t:       t,
	}
}

func (p internetFwSectionCfg) getTfConfig(index int) string {
	tmpl, err := template.New("tmpl").Parse(internetFwSectionTFs[index])
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

var internetFwSectionTFs = []string{
	`resource "cato_if_section" "this" {
		at = {
			position = "LAST_IN_POLICY"
		}

		section = {
			name = "{{.Name}}"
		}
	}
	`,
	`resource "cato_if_section" "this" {
		at = {
			position = "LAST_IN_POLICY"
		}

		section = {
			name = "{{.Name}}-2"
		}
	}
	`,
}

func internetFwSectionUnderSubPolicyConfig(name string) string {
	return acc.ProviderCfg() + fmt.Sprintf(`
resource "cato_if_sub_policy" "test" {
  name        = %q
  description = "IF section acceptance test"

  at = {
    position = "LAST_IN_POLICY"
  }

  scope = {
    enabled     = true
    source      = { ip = ["10.203.0.1"] }
    destination = {}
    tracking    = { event = { enabled = false } }
  }
}

resource "cato_if_section" "this" {
  at = {
    position = "LAST_IN_POLICY"
    ref      = cato_if_sub_policy.test.id
  }

  section = {
    name = %q
  }
}
`, name+"-parent", name)
}
