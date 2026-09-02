package global_ip_ranges

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccGlobalIPRanges(t *testing.T) {
	acc.SkipByEnv(t)
	acc.DeleteGlobalIPRanges(t)
	mockSrv := accmock.NewMockServer(t, "TestAccGlobalIPRanges")
	defer mockSrv.Close()
	mockSrv.Run()
	cfg := newGlobalIPRangesCfg(t)
	res := "cato_global_ip_ranges.this"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				// Create the resource
				Config: cfg.getTfConfig(0),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "ranges.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "ranges.*", map[string]string{
						"name":        "acctest_ip_range_1",
						"description": "acctest_ip_range_1 description",
						"ip_range":    "255.255.1.0/24",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(res, "ranges.*", map[string]string{
						"name":        "acctest_ip_range_2",
						"description": "acctest_ip_range_2 description",
						"ip_range":    "255.255.2.10",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(res, "ranges.*", map[string]string{
						"name":        "acctest_ip_range_3",
						"description": "acctest_ip_range_3 description",
						"ip_range":    "255.255.3.10-255.255.3.20",
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
					resource.TestCheckResourceAttr(res, "ranges.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "ranges.*", map[string]string{
						"name":        "acctest_ip_range_1",
						"description": "acctest_ip_range_1 description",
						"ip_range":    "255.255.1.0/24",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(res, "ranges.*", map[string]string{
						"name":        "acctest_ip_range_2",
						"description": "acctest_ip_range_2 new description",
						"ip_range":    "255.255.2.10",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(res, "ranges.*", map[string]string{
						"name":        "acctest_ip_range_4",
						"description": "acctest_ip_range_4 description",
						"ip_range":    "255.255.4.10-255.255.4.20",
					}),
					checkGlobalIPRangeAbsent(t, "acctest_ip_range_3"),
				),
			},
		},
	})
}

func checkGlobalIPRangeAbsent(t *testing.T, name string) resource.TestCheckFunc {
	t.Helper()

	return func(*terraform.State) error {
		result, err := acc.GetClient(t).ObjectGlobalIPRangeList(context.Background(), acc.CatoAccountID, nil)
		if err != nil {
			return fmt.Errorf("failed to list global IP ranges: %w", err)
		}

		for _, item := range result.GetObject().GetGlobalIPRangeList().GetItems() {
			if item.GetName() == name {
				return fmt.Errorf("global IP range %q still exists", name)
			}
		}

		return nil
	}
}

type globalIPRangesCfg struct {
	t *testing.T
}

func newGlobalIPRangesCfg(t *testing.T) globalIPRangesCfg {
	return globalIPRangesCfg{
		t: t,
	}
}

func (p globalIPRangesCfg) getTfConfig(index int) string {
	tmpl, err := template.New("tmpl").Parse(globalIPRangesTFs[index])
	if err != nil {
		p.t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err != nil {
		p.t.Fatal(err)
	}

	cfg := acc.ProviderCfg() + buf.String()
	fmt.Println(cfg)
	return cfg
}

var globalIPRangesTFs = []string{`
	resource "cato_global_ip_ranges" "this" {
	    ranges = [
			{
				name        = "acctest_ip_range_1"
				description = "acctest_ip_range_1 description"
				ip_range    =  "255.255.1.0/24"
			},
			{
				name        = "acctest_ip_range_2"
				description = "acctest_ip_range_2 description"
				ip_range    =  "255.255.2.10"
			},
			{
				name        = "acctest_ip_range_3"
				description = "acctest_ip_range_3 description"
				ip_range    =  "255.255.3.10-255.255.3.20"
			}
		]
	}`,

	`resource "cato_global_ip_ranges" "this" {
	    ranges = [
			{
				name        = "acctest_ip_range_1"
				description = "acctest_ip_range_1 description"
				ip_range    =  "255.255.1.0/24"
			},
			{
				name        = "acctest_ip_range_2"
				description = "acctest_ip_range_2 new description"
				ip_range    =  "255.255.2.10"
			},
			{
				name        = "acctest_ip_range_4"
				description = "acctest_ip_range_4 description"
				ip_range    =  "255.255.4.10-255.255.4.20"
			}
		]
	}`,
}
