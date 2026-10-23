//go:build acctest

package role

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccRole(t *testing.T) {
	acc.SkipByEnv(t)
	name := "tfacc-role-" + acctest.RandString(8)
	config := func(description string) string {
		return fmt.Sprintf(`
resource "cato_role" "test" {
 name = %q
 description = %q
 permissions = [{resource = "Sites", action = "VIEW"}]
}
data "cato_role" "test" { role_id = cato_role.test.role_id }
data "cato_roles" "test" { filter = { name = { eq = cato_role.test.name } } }
data "cato_permission_catalog" "test" {}
`, name, description)
	}
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories, PreCheck: acc.CheckCMAVars(t), Steps: []resource.TestStep{
		{Config: config("initial"), Check: resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttr("cato_role.test", "name", name), resource.TestCheckResourceAttr("data.cato_role.test", "name", name), resource.TestCheckResourceAttr("data.cato_roles.test", "total", "1"))},
		{ResourceName: "cato_role.test", ImportState: true, ImportStateVerify: true},
		{Config: config(""), Check: resource.TestCheckResourceAttr("cato_role.test", "description", "")},
	}})
}
