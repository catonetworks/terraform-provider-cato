//go:build acctest

package always_on_section

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"text/template"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccAlwaysOnSection(t *testing.T) {
	acc.SkipByEnv(t)
	mockSrv := accmock.NewMockServer(t, t.Name())
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newAlwaysOnSectionCfg(t)
	res := "cato_always_on_section.this"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				// Create the resource.
				Config: cfg.getTfConfig(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "at.position", "LAST_IN_POLICY"),
					resource.TestCheckResourceAttr(res, "at.ref", ""),
					resource.TestCheckResourceAttrSet(res, "section.id"),
					resource.TestCheckResourceAttr(res, "section.name", cfg.resName),
				),
			},
			{
				// Rename the section without changing its placement.
				Config: cfg.getTfConfig(1),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "at.position", "LAST_IN_POLICY"),
					resource.TestCheckResourceAttr(res, "section.name", cfg.resName+"-renamed"),
					resource.TestCheckResourceAttrSet(res, "section.id"),
				),
			},
			{
				// Import hydrates the section by its nested identifier.
				ResourceName:                         res,
				ImportState:                          true,
				ImportStateIdFunc:                    sectionImportID(res),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "section.id",
				// The read API does not return the original relative-position input.
				ImportStateVerifyIgnore: []string{"at"},
			},
			{
				// Move before another section owned by this test.
				Config: cfg.getTfConfig(2),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "at.position", "BEFORE_SECTION"),
					resource.TestCheckResourceAttrPair(res, "at.ref", "cato_always_on_section.reference", "section.id"),
					sectionOrderCheck(t, res, true),
					resource.TestCheckResourceAttr(res, "section.name", cfg.resName+"-renamed"),
				),
			},
			{
				// Move after the same section; Terraform deletes both sections at the end.
				Config: cfg.getTfConfig(3),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "at.position", "AFTER_SECTION"),
					resource.TestCheckResourceAttrPair(res, "at.ref", "cato_always_on_section.reference", "section.id"),
					sectionOrderCheck(t, res, false),
					resource.TestCheckResourceAttr(res, "section.name", cfg.resName+"-renamed"),
				),
			},
		},
	})
}

func sectionImportID(resourceName string) resource.ImportStateIdFunc {
	return func(state *terraform.State) (string, error) {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok || rs.Primary == nil || rs.Primary.Attributes["section.id"] == "" {
			return "", fmt.Errorf("resource %s has no section.id to import", resourceName)
		}
		return rs.Primary.Attributes["section.id"], nil
	}
}

func sectionOrderCheck(t *testing.T, resourceName string, before bool) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		id, err := sectionImportID(resourceName)(state)
		if err != nil {
			return err
		}
		ref, err := sectionImportID("cato_always_on_section.reference")(state)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		result, err := acc.GetClient(t).PolicyZtnaAlwaysOnPolicy(ctx, acc.CatoAccountID, nil)
		if err != nil {
			return err
		}
		indexes := make(map[string]int)
		for index, entry := range result.GetPolicy().GetZtnaAlwaysOn().GetPolicy().GetSections() {
			indexes[entry.GetSection().GetID()] = index
		}
		index, found := indexes[id]
		refIndex, refFound := indexes[ref]
		if !found || !refFound {
			return fmt.Errorf("test sections are missing from the policy")
		}
		if before && index+1 == refIndex || !before && refIndex+1 == index {
			return nil
		}
		return fmt.Errorf("unexpected section order: section index %d, reference index %d, before=%t", index, refIndex, before)
	}
}

type alwaysOnSectionCfg struct {
	resName string
	t       *testing.T
}

func newAlwaysOnSectionCfg(t *testing.T) alwaysOnSectionCfg {
	return alwaysOnSectionCfg{resName: acc.GetRandName("always_on_section"), t: t}
}

func (cfg alwaysOnSectionCfg) getTfConfig(index int) string {
	cfg.t.Helper()
	tmpl, err := template.New("always_on_section").Parse(alwaysOnSectionTFs[index])
	if err != nil {
		cfg.t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, map[string]string{"Name": cfg.resName}); err != nil {
		cfg.t.Fatal(err)
	}
	return acc.ProviderCfg() + buf.String()
}

var alwaysOnSectionTFs = []string{
	`resource "cato_always_on_section" "this" {
	  at = {
	    position = "LAST_IN_POLICY"
	  }
	  section = {
	    name = "{{.Name}}"
	  }
	}`,
	`resource "cato_always_on_section" "this" {
	  at = { position = "LAST_IN_POLICY" }
	  section = { name = "{{.Name}}-renamed" }
	}`,
	`resource "cato_always_on_section" "reference" {
	  at = { position = "LAST_IN_POLICY" }
	  section = { name = "{{.Name}}-reference" }
	}
	resource "cato_always_on_section" "this" {
	  at = {
	    position = "BEFORE_SECTION"
	    ref = cato_always_on_section.reference.section.id
	  }
	  section = { name = "{{.Name}}-renamed" }
	}`,
	`resource "cato_always_on_section" "reference" {
	  at = { position = "LAST_IN_POLICY" }
	  section = { name = "{{.Name}}-reference" }
	}
	resource "cato_always_on_section" "this" {
	  at = {
	    position = "AFTER_SECTION"
	    ref = cato_always_on_section.reference.section.id
	  }
	  section = { name = "{{.Name}}-renamed" }
	}`,
}
