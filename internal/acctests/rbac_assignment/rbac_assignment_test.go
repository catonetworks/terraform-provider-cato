//go:build acctest

package rbacassignment

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

// TestAccRBACAssignment requires an existing test user and never creates principals.
func TestAccRBACAssignment(t *testing.T) {
	acc.SkipByEnv(t)
	user := os.Getenv("TFACC_RBAC_USER_ID")
	if user == "" {
		t.Skip("set TFACC_RBAC_USER_ID to an existing disposable user principal")
	}
	name := "tfacc-assignment-" + acctest.RandString(8)
	config := func(description string) string {
		return acc.ProviderCfg() + fmt.Sprintf(`
resource "cato_role" "test" {
 name = %q
 permissions = [{resource = "Sites", action = "VIEW"}]
}
resource "cato_rbac_assignment" "test" {
 user_id = %q
 role_ids = [cato_role.test.role_id]
 allowed_entities = []
 description = %q
}
`, name, user, description)
	}
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories, PreCheck: acc.CheckCMAVars(t), Steps: []resource.TestStep{
		{Config: config("initial"), Check: resource.TestCheckResourceAttr("cato_rbac_assignment.test", "user_id", user)},
		{ResourceName: "cato_rbac_assignment.test", ImportState: true, ImportStateVerify: true},
		{Config: config("updated"), Check: resource.TestCheckResourceAttr("cato_rbac_assignment.test", "description", "updated")},
	}})
}
