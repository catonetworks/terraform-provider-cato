//go:build acctest

package application_control_policy

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

// TestAccApplicationControlPolicy verifies Create, Read, Update, and Import of the
// singleton Application Control policy resource. The policy already exists in every
// account, so this test only exercises toggling its attributes.
func TestAccApplicationControlPolicy(t *testing.T) {
	acc.SkipByEnv(t)
	mockSrv := accmock.NewMockServer(t, "TestAccApplicationControlPolicy")
	defer mockSrv.Close()
	mockSrv.Run()

	res := "cato_application_control_policy.this"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				// Create / adopt the singleton policy with both toggles enabled.
				Config: policyConfig("true", "ENABLED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "enabled", "true"),
					resource.TestCheckResourceAttr(res, "data_control_enabled", "ENABLED"),
					resource.TestCheckResourceAttrSet(res, "id"),
				),
			},
			{
				// Import the singleton by its fixed ID.
				ImportState:       true,
				ResourceName:      res,
				ImportStateVerify: true,
			},
			{
				// Update: disable data control.
				Config: policyConfig("true", "DISABLED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(res),
					resource.TestCheckResourceAttr(res, "enabled", "true"),
					resource.TestCheckResourceAttr(res, "data_control_enabled", "DISABLED"),
				),
			},
			{
				// Restore: re-enable data control so the account is left in a known state.
				Config: policyConfig("true", "ENABLED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(res, "data_control_enabled", "ENABLED"),
				),
			},
		},
	})
}

func policyConfig(enabled, dataControlEnabled string) string {
	return acc.ProviderCfg() + fmt.Sprintf(`
resource "cato_application_control_policy" "this" {
  enabled              = %s
  data_control_enabled = %q
}
`, enabled, dataControlEnabled)
}
