//go:build acctest

package rbac_assignment

import (
	"context"
	"fmt"
	"os"
	"testing"

	cato "github.com/catonetworks/cato-go-sdk"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

// Use a dedicated, unassigned test user or users group. These tests grant access
// to that principal and remove the assignment at the end.
func TestAccRBACAssignment(t *testing.T) {
	acc.SkipByEnv(t)
	for _, principal := range []struct{ field, env string }{{"user_id", "CATO_TEST_RBAC_USER_ID"}, {"group_id", "CATO_TEST_RBAC_GROUP_ID"}} {
		t.Run(principal.field, func(t *testing.T) {
			id := os.Getenv(principal.env)
			if id == "" {
				t.Skip("Set " + principal.env + " to a dedicated unassigned test principal")
			}
			siteID := os.Getenv("CATO_TEST_RBAC_SITE_ID")
			if siteID == "" {
				t.Skip("Set CATO_TEST_RBAC_SITE_ID to an existing test site")
			}
			name := acc.GetRandName("rbac-assignment")
			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories, PreCheck: acc.CheckCMAVars(t), CheckDestroy: checkAssignmentDestroy,
				Steps: []resource.TestStep{
					{Config: assignmentConfig(name, principal.field, id, siteID, false), Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("cato_rbac_assignment.this", principal.field, id),
						resource.TestCheckResourceAttr("cato_rbac_assignment.this", "role_ids.#", "1"),
						resource.TestCheckResourceAttr("cato_rbac_assignment.this", "allowed_entities.#", "1"),
						resource.TestCheckResourceAttrSet("cato_rbac_assignment.this", "assignment_id"),
					)},
					{ResourceName: "cato_rbac_assignment.this", ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"allow_viewing_unobfuscated_data"}},
					{Config: assignmentConfig(name, principal.field, id, siteID, true), Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("cato_rbac_assignment.this", "role_ids.#", "2"),
						resource.TestCheckResourceAttr("cato_rbac_assignment.this", "allowed_entities.#", "0"),
						resource.TestCheckResourceAttr("cato_rbac_assignment.this", "description", ""),
					)},
				},
			})
		})
	}
}
func assignmentConfig(name, field, id, siteID string, updated bool) string {
	roles := "[cato_role.viewer.role_id]"
	description := "Acceptance assignment"
	entities := fmt.Sprintf(`[{id = %q, type = "SITE", action = "VIEW"}]`, siteID)
	if updated {
		roles = "[cato_role.viewer.role_id, cato_role.editor.role_id]"
		description = ""
		entities = "[]"
	}
	return acc.ProviderCfg() + fmt.Sprintf(`
resource "cato_role" "viewer" {
 name = %q
 permissions = [{resource = "Sites", action = "VIEW"}]
}
resource "cato_role" "editor" {
 name = %q
 permissions = [{resource = "Sites", action = "EDIT"}]
}
resource "cato_rbac_assignment" "this" {
 %s = %q
 role_ids = %s
 allowed_entities = %s
 allow_viewing_unobfuscated_data = false
 description = %q
}
`, name+"-viewer", name+"-editor", field, id, roles, entities, description)
}
func checkAssignmentDestroy(state *terraform.State) error {
	client, err := cato.New(os.Getenv("CATO_BASEURL"), os.Getenv("CATO_TOKEN"), acc.CatoAccountID, nil, nil)
	if err != nil {
		return err
	}
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "cato_rbac_assignment" {
			continue
		}
		for page := 0; ; page++ {
			var result struct {
				Rbac *struct {
					List *struct {
						Items []struct {
							ID string `graphql:"id"`
						} `graphql:"items"`
						Paging *struct {
							Total int `graphql:"total"`
						} `graphql:"paging"`
					} `graphql:"rbacAssignmentList"`
				} `graphql:"rbac"`
			}
			err := client.Client.Post(context.Background(), "assignmentDestroyCheck", `query assignmentDestroyCheck($accountId: ID!, $input: RBACAssignmentListInput!) {rbac(accountId: $accountId) {rbacAssignmentList(input: $input) {items {id} paging {total}}}}`, &result, map[string]any{"accountId": rs.Primary.Attributes["account_id"], "input": map[string]any{"paging": map[string]int{"from": page, "limit": 100}}})
			if err != nil {
				return err
			}
			if result.Rbac == nil || result.Rbac.List == nil || result.Rbac.List.Paging == nil {
				return fmt.Errorf("incomplete assignment destroy-check response")
			}
			for _, item := range result.Rbac.List.Items {
				if item.ID == rs.Primary.Attributes["assignment_id"] {
					return fmt.Errorf("RBAC assignment %s still exists", item.ID)
				}
			}
			if (page+1)*100 >= result.Rbac.List.Paging.Total {
				break
			}
			if len(result.Rbac.List.Items) == 0 {
				return fmt.Errorf("assignment destroy-check pagination made no progress")
			}
		}
	}
	return nil
}
