//go:build acctest

package if_rule

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

// ENG-216525: adding a rule must only create that rule and update bulk ordering.
// Existing rules with name-only group references must retain their resolved IDs.
func TestAccInternetFw_UserGroups_AddRuleDoesNotUpdateExisting(t *testing.T) {
	acc.SkipByEnv(t)
	acc.CleanupFirewallAndWANPolicyRevisions(t)
	defer acc.CleanupFirewallAndWANPolicyRevisions(t)
	mockSrv := accmock.NewMockServer(t, "TestAccInternetFw_UserGroups_AddRuleDoesNotUpdateExisting")
	defer mockSrv.Close()
	mockSrv.Run()
	groups := internetFwUserGroupsForTest(t)
	name := acc.GetRandName("ifw_group_plan")
	existing := "cato_if_rule.existing"
	initial := internetFwUserGroupsConfig(name, []acc.Ref{groups[0], groups[1]}, false)
	reversed := internetFwUserGroupsConfig(name, []acc.Ref{groups[1], groups[0]}, false)
	added := internetFwUserGroupsConfig(name, []acc.Ref{groups[1], groups[0]}, true)
	check := resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr(existing, "rule.source.users_group.#", "2"),
		resource.TestCheckTypeSetElemNestedAttrs(existing, "rule.source.users_group.*", map[string]string{"name": groups[0].Name, "id": groups[0].ID}),
		resource.TestCheckTypeSetElemNestedAttrs(existing, "rule.source.users_group.*", map[string]string{"name": groups[1].Name, "id": groups[1].ID}),
	)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{Config: initial, Check: check},
			{Config: reversed, PlanOnly: true, ConfigPlanChecks: resource.ConfigPlanChecks{PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}},
			{
				Config: added,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(existing, plancheck.ResourceActionNoop),
					plancheck.ExpectResourceAction("cato_if_rule.additional[0]", plancheck.ResourceActionCreate),
					plancheck.ExpectResourceAction("cato_if_section.groups", plancheck.ResourceActionNoop),
					plancheck.ExpectResourceAction("cato_bulk_if_move_rule.groups", plancheck.ResourceActionUpdate),
				}},
				Check: check,
			},
			{Config: added, PlanOnly: true, ConfigPlanChecks: resource.ConfigPlanChecks{PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}},
		},
	})
}

// Changing a configured group name must resolve the new group ID rather than
// retaining the old group's ID through UseStateForUnknown.
func TestAccInternetFw_UserGroups_ChangeReference(t *testing.T) {
	acc.SkipByEnv(t)
	acc.CleanupFirewallAndWANPolicyRevisions(t)
	defer acc.CleanupFirewallAndWANPolicyRevisions(t)
	mockSrv := accmock.NewMockServer(t, "TestAccInternetFw_UserGroups_ChangeReference")
	defer mockSrv.Close()
	mockSrv.Run()
	groups := internetFwUserGroupsForTest(t)
	name := acc.GetRandName("ifw_group_change")
	existing := "cato_if_rule.existing"
	check := func(group acc.Ref) resource.TestCheckFunc {
		return resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr(existing, "rule.source.users_group.#", "1"),
			resource.TestCheckTypeSetElemNestedAttrs(existing, "rule.source.users_group.*", map[string]string{"name": group.Name, "id": group.ID}),
		)
	}
	initial := internetFwUserGroupsConfig(name, groups[:1], false)
	changed := internetFwUserGroupsConfig(name, groups[1:2], false)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{Config: initial, Check: check(groups[0])},
			{Config: changed, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(existing, plancheck.ResourceActionUpdate)}}, Check: check(groups[1])},
			{Config: changed, PlanOnly: true, ConfigPlanChecks: resource.ConfigPlanChecks{PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}},
			{Config: initial, Check: check(groups[0])},
		},
	})
}

func internetFwUserGroupsForTest(t *testing.T) []acc.Ref {
	t.Helper()
	// Name-only references are ambiguous when the account has duplicate names.
	refs := acc.GetUserGroups(t)
	counts := make(map[string]int)
	for _, ref := range refs {
		counts[ref.Name]++
	}
	groups := make([]acc.Ref, 0, 2)
	for _, ref := range refs {
		if ref.ID != "" && ref.Name != "" && counts[ref.Name] == 1 {
			groups = append(groups, ref)
			if len(groups) == 2 {
				return groups
			}
		}
	}
	t.Skip("requires two user groups with distinct, nonempty names and IDs in TFACC_TEST_VARS")
	return nil
}

func internetFwUserGroupsConfig(name string, groups []acc.Ref, additional bool) string {
	refs := make([]string, 0, len(groups))
	for _, group := range groups {
		refs = append(refs, fmt.Sprintf("{ name = %q }", group.Name))
	}
	extra := ""
	if additional {
		extra = fmt.Sprintf("additional = { name = %q, source = {} }", name+"-existing-additional")
	}
	return acc.ProviderCfg() + fmt.Sprintf(`
locals {
 rules = {
  existing = { name = %q, source = { users_group = [%s] } }
  %s
 }
}
resource "cato_if_section" "groups" {
 at = { position = "LAST_IN_POLICY" }
 section = { name = %q }
}
resource "cato_if_rule" "existing" {
 depends_on = [cato_if_section.groups]
 at = { position = "LAST_IN_POLICY" }
 rule = {
  name = local.rules.existing.name
  enabled = true
  action = "ALLOW"
  source = local.rules.existing.source
  destination = { domain = ["example.com"] }
  tracking = { event = { enabled = true } }
 }
}
resource "cato_if_rule" "additional" {
 depends_on = [cato_if_section.groups]
 count = length(local.rules) > 1 ? 1 : 0
 at = { position = "LAST_IN_POLICY" }
 rule = {
  name = "${local.rules.existing.name}-additional"
  enabled = true
  action = "ALLOW"
  source = {}
  destination = { domain = ["example.com"] }
  tracking = { event = { enabled = true } }
 }
}
resource "cato_bulk_if_move_rule" "groups" {
 depends_on = [cato_if_section.groups, cato_if_rule.existing, cato_if_rule.additional]
 section_data = { %q = { section_name = %q, section_index = 1 } }
 rule_data = {
  for idx, key in sort(keys(local.rules)) : local.rules[key].name => {
   rule_name = local.rules[key].name
   section_name = %q
   index_in_section = idx + 1
  }
 }
}
`, name+"-existing", strings.Join(refs, ","), extra, name, name, name, name)
}
