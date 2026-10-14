//go:build acctest

package role

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

func TestAccRole(t *testing.T) {
	acc.SkipByEnv(t)
	name := acc.GetRandName("iam-role")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		CheckDestroy:             checkRoleDestroy,
		Steps: []resource.TestStep{
			{Config: roleConfig(name, true), Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("cato_role.this", "name", name),
				resource.TestCheckResourceAttr("cato_role.this", "description", "Acceptance role"),
				resource.TestCheckResourceAttr("cato_role.this", "permissions.#", "1"),
				resource.TestCheckResourceAttr("cato_role.this", "predefined", "false"),
				resource.TestCheckResourceAttrSet("cato_role.this", "role_id"),
				resource.TestCheckResourceAttrPair("data.cato_role.this", "role_id", "cato_role.this", "role_id"),
				resource.TestCheckResourceAttr("data.cato_roles.this", "total", "1"),
				resource.TestCheckResourceAttr("data.cato_roles.this", "items.#", "1"),
				resource.TestCheckResourceAttrSet("data.cato_permission_catalog.this", "resources.#"),
			)},
			{ResourceName: "cato_role.this", ImportState: true, ImportStateVerify: true},
			{Config: roleConfig(name+"-updated", false), Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("cato_role.this", "name", name+"-updated"),
				resource.TestCheckResourceAttr("cato_role.this", "description", ""),
				resource.TestCheckTypeSetElemNestedAttrs("cato_role.this", "permissions.*", map[string]string{"resource": "Sites", "action": "EDIT"}),
				resource.TestCheckResourceAttrPair("data.cato_role.this", "name", "cato_role.this", "name"),
				resource.TestCheckResourceAttr("data.cato_roles.this", "total", "1"),
			)},
		},
	})
}

func roleConfig(name string, initial bool) string {
	description := ""
	action := "EDIT"
	if initial {
		description = `description = "Acceptance role"`
		action = "VIEW"
	}
	return acc.ProviderCfg() + fmt.Sprintf(`
resource "cato_role" "this" {
 name = %q
 %s
 permissions = [{resource = "Sites", action = %q}]
}
data "cato_role" "this" { role_id = cato_role.this.role_id }
data "cato_roles" "this" {
 filter = {id = {eq = cato_role.this.role_id}, predefined = {eq = false}}
}
data "cato_permission_catalog" "this" {}
`, name, description, action)
}

func checkRoleDestroy(state *terraform.State) error {
	client, err := cato.New(os.Getenv("CATO_BASEURL"), os.Getenv("CATO_TOKEN"), acc.CatoAccountID, nil, nil)
	if err != nil {
		return err
	}
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "cato_role" {
			continue
		}
		result, queryErr := client.RbacRoleManagementRole(context.Background(), rs.Primary.Attributes["account_id"], rs.Primary.Attributes["role_id"])
		if queryErr != nil {
			return queryErr
		}
		if result.GetRbac().GetRoleManagement().GetRole() != nil {
			return fmt.Errorf("IAM role %s still exists", rs.Primary.Attributes["role_id"])
		}
	}
	return nil
}
