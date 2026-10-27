//go:build acctest

package clientconnectivityrule

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccClientConnectivityRule(t *testing.T) {
	acc.SkipByEnv(t)
	fixtureRange := ""

	config := func(name, description string) string {
		return strings.NewReplacer("${name}", name, "${description}", description, "${range}", fixtureRange).Replace(`resource "cato_client_connectivity_rule" "test" {
rule = {"action": "ALLOW", "confidence_level": "HIGH", "connection_origin": [], "country": [], "description": "${description}", "device": [], "enabled": true, "name": "acctest-client_connectivity_rule-${name}", "platform": [], "source": {"user": [], "users_group": []}, "source_range": []}
}
`)
	}
	name := acc.GetRandName("client_connectivity_rule")
	resourceName := "cato_client_connectivity_rule.test"
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
			{Config: config(name+"-updated", "updated"), Check: resource.TestCheckResourceAttr(resourceName, "rule.name", "acctest-client_connectivity_rule-"+name+"-updated")},
		},
	})
}
