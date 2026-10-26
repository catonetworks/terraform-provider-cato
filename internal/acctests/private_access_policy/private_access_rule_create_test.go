//go:build acctest

package private_access_policy

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

// The minimal fixture covers Create/read-after-write without depending on
// device posture profiles, notification targets, or whole-policy ordering.
func TestAccPrivAccessRuleCreateReadConsistency(t *testing.T) {
	acc.SkipByEnv(t)
	if os.Getenv("TF_ACC") == "" {
		t.Skip("set TF_ACC to run live acceptance tests")
	}
	acc.CheckCMAVars(t)()
	apps, users := acc.GetPrivateApps(t), acc.GetUsers(t)
	config := acc.ProviderCfg() + fmt.Sprintf(`
resource "cato_private_access_rule" "test" {
 name = %q
 action = "BLOCK"
 enabled = false
 applications = [{name = %q}]
 source = {users = [{name = %q}]}
}
`, acc.GetRandName("private_access_create"), apps[0].Name, users[0].Name)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{Config: config, Check: resource.TestCheckResourceAttrSet("cato_private_access_rule.test", "id")},
			{Config: config, PlanOnly: true},
		},
	})
}
