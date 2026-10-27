//go:build acctest

package clientconnectivitysection

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccClientConnectivitySection(t *testing.T) {
	acc.SkipByEnv(t)
	fixtureRange := ""

	config := func(name, description string) string {
		return strings.NewReplacer("${name}", name, "${description}", description, "${range}", fixtureRange).Replace(`resource "cato_client_connectivity_section" "test" {
at = {position = "LAST_IN_POLICY"}
section = {name = "acctest-client_connectivity_section-${name}"}
}
`)
	}
	name := acc.GetRandName("client_connectivity_section")
	resourceName := "cato_client_connectivity_section.test"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{Config: config(name, "initial"), Check: resource.TestCheckResourceAttrSet(resourceName, "id")},
			{ResourceName: resourceName, ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"at"}},
			{Config: config(name+"-updated", "updated"), Check: resource.TestCheckResourceAttr(resourceName, "section.name", "acctest-client_connectivity_section-"+name+"-updated")},
		},
	})
}
