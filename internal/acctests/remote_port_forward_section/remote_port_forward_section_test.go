//go:build acctest

package rpfsection

import (
	"context"
	"fmt"
	"strings"
	"testing"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccRemotePortFwdSection(t *testing.T) {
	acc.SkipByEnv(t)
	fixtureRange := ""

	config := func(name, description string) string {
		return acc.ProviderCfg() + strings.NewReplacer("${name}", name, "${description}", description, "${position}", func() string {
			if description == "updated" {
				return "BEFORE_SECTION"
			}
			return "LAST_IN_POLICY"
		}(), "${range}", fixtureRange, "${anchor_name}", strings.TrimSuffix(name, "-updated"), "${reference}", func() string {
			if description == "updated" {
				return "cato_remote_port_forward_section.anchor.id"
			}
			return "null"
		}()).Replace(`resource "cato_remote_port_forward_section" "anchor" {
 at = { position = "LAST_IN_POLICY" }
 section = { name = "acctest-remote_port_forward_section-anchor-${anchor_name}" }
}
resource "cato_remote_port_forward_section" "test" {
 depends_on = [cato_remote_port_forward_section.anchor]
at = { position = "${position}", ref = ${reference} }
section = {name = "acctest-remote_port_forward_section-${name}"}
}
`)
	}
	name := acc.GetRandName("remote_port_forward_section")
	resourceName := "cato_remote_port_forward_section.test"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		CheckDestroy: func(state *terraform.State) error {
			revisionType := cato_models.PolicyRevisionTypePublic
			policy, err := acc.GetClient(t).RemotePortFwdPolicy(context.Background(), acc.CatoAccountID,
				&cato_models.RemotePortFwdPolicyInput{Revision: &cato_models.PolicyRevisionInput{Type: &revisionType}})
			if err != nil {
				return err
			}
			for _, item := range policy.GetPolicy().GetRemotePortFwd().GetPolicy().GetSections() {
				for _, object := range state.RootModule().Resources {
					if object.Type == "cato_remote_port_forward_section" && object.Primary != nil && item.GetSection().GetID() == object.Primary.Attributes["id"] {
						return fmt.Errorf("test resource still exists after destroy")
					}
				}
			}
			return nil
		},
		PreCheck: acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{Config: config(name, "initial"), Check: resource.TestCheckResourceAttrSet(resourceName, "id")},
			{Config: config(name, "initial"), PlanOnly: true, ExpectNonEmptyPlan: false},
			{ResourceName: resourceName, ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"at"}},
			{Config: config(name+"-updated", "updated"), Check: resource.TestCheckResourceAttr(resourceName, "section.name", "acctest-remote_port_forward_section-"+name+"-updated")},
			{Config: config(name+"-updated", "updated"), PlanOnly: true, ExpectNonEmptyPlan: false},
		},
	})
}
