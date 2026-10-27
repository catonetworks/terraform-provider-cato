//go:build acctest

package terminalserverrule

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccTerminalServerRule(t *testing.T) {
	acc.SkipByEnv(t)
	fixtureRange := ""
	fixtureRange = os.Getenv("TFACC_TERMINAL_SERVER_GLOBAL_IP_RANGE_ID")
	if fixtureRange == "" {
		t.Skip("TFACC_TERMINAL_SERVER_GLOBAL_IP_RANGE_ID is required")
	}
	config := func(name, description string) string {
		return strings.NewReplacer("${name}", name, "${description}", description, "${range}", fixtureRange).Replace(`resource "cato_terminal_server_rule" "test" {
rule = {"allowed_host_ip": {"id": "${range}"}, "description": "${description}", "enabled": true, "exclude_traffic": [], "name": "acctest-terminal_server_rule-${name}"}
}
`)
	}
	name := acc.GetRandName("terminal_server_rule")
	resourceName := "cato_terminal_server_rule.test"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{Config: config(name, "initial"), Check: resource.TestCheckResourceAttrSet(resourceName, "rule.id")},
			{ResourceName: resourceName, ImportState: true, ImportStateVerify: true, ImportStateIdFunc: func(state *terraform.State) (string, error) {
				object := state.RootModule().Resources[resourceName]
				if object == nil || object.Primary == nil {
					return "", fmt.Errorf("missing resource state")
				}
				return object.Primary.Attributes["rule.id"], nil
			}, ImportStateVerifyIgnore: []string{"at"}},
			{Config: config(name+"-updated", "updated"), Check: resource.TestCheckResourceAttr(resourceName, "rule.name", "acctest-terminal_server_rule-"+name+"-updated")},
		},
	})
}
