//go:build acctest

package splittunnelrule

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccSplitTunnelRule(t *testing.T) {
	acc.SkipByEnv(t)
	fixtureRange := ""

	config := func(name, description string) string {
		return strings.NewReplacer("${name}", name, "${description}", description, "${range}", fixtureRange).Replace(`resource "cato_split_tunnel_rule" "test" {
rule = {"action": "ROUTE_ONLY_SELECTED", "country": [], "coverage": "ALL", "description": "${description}", "destination": {"application": [], "domain": [], "fqdn": [], "global_ip_range": []}, "enabled": true, "name": "acctest-split_tunnel_rule-${name}", "platform": [], "routing_priority": "LAN", "source": {"user": [], "users_group": []}, "source_network": {"source_network_type": "ANY"}}
}
`)
	}
	name := acc.GetRandName("split_tunnel_rule")
	resourceName := "cato_split_tunnel_rule.test"
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
			{Config: config(name+"-updated", "updated"), Check: resource.TestCheckResourceAttr(resourceName, "rule.name", "acctest-split_tunnel_rule-"+name+"-updated")},
		},
	})
}
