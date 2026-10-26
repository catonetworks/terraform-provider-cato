//go:build acctest

package if_rule

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

// ENG-221537: references resolving to the same category must not erase the set.
func TestAccInternetFw_AppCategoriesDuplicateReferences(t *testing.T) {
	acc.SkipByEnv(t)
	acc.CleanupFirewallAndWANPolicyRevisions(t)
	defer acc.CleanupFirewallAndWANPolicyRevisions(t)
	mockSrv := accmock.NewMockServer(t, "TestAccInternetFw_AppCategoriesDuplicateReferences")
	defer mockSrv.Close()
	mockSrv.Run()
	name := acc.GetRandName("ifw_categories")
	config := func(categories string) string {
		return acc.ProviderCfg() + fmt.Sprintf(`
resource "cato_if_rule" "categories" {
  at = { position = "LAST_IN_POLICY" }
  rule = {
    name = %q
    enabled = true
    action = "BLOCK"
    source = {}
    destination = { app_category = [%s] }
    tracking = { event = { enabled = true } }
  }
}`, name, categories)
	}
	initial := config(`{ id = "suspected_malware" }, { name = "Malware" }, { name = "Alcohol and Tobacco" }`)
	updated := config(`{ name = "Alcohol and Tobacco" }, { name = "Botnets" }, { name = "Malware" }, { id = "suspected_malware" }`)
	check := func(count string) resource.TestCheckFunc {
		return resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("cato_if_rule.categories", "rule.destination.app_category.#", count),
			resource.TestCheckTypeSetElemNestedAttrs("cato_if_rule.categories", "rule.destination.app_category.*", map[string]string{"id": "suspected_malware", "name": "Malware"}),
			resource.TestCheckTypeSetElemNestedAttrs("cato_if_rule.categories", "rule.destination.app_category.*", map[string]string{"id": "food_drinks_tobacco", "name": "Alcohol and Tobacco"}),
		)
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{Config: initial, Check: check("2")},
			{Config: initial, PlanOnly: true, ConfigPlanChecks: resource.ConfigPlanChecks{PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}},
			{Config: updated, Check: resource.ComposeAggregateTestCheckFunc(check("3"), resource.TestCheckTypeSetElemNestedAttrs("cato_if_rule.categories", "rule.destination.app_category.*", map[string]string{"id": "botnets", "name": "Botnets"}))},
			{Config: updated, PlanOnly: true, ConfigPlanChecks: resource.ConfigPlanChecks{PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}},
		},
	})
}
