//go:build acctest

package swpsection

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccSiteWebProxySection(t *testing.T) {
	acc.SkipByEnv(t)
	acc.RequireLive(t)
	fixtureRange := ""

	config := func(name, description string) string {
		return acc.ProviderCfg() + strings.NewReplacer("${name}", name, "${description}", description, "${position}", func() string {
			if description == "updated" {
				return "BEFORE_SECTION"
			}
			return "LAST_IN_POLICY"
		}(), "${range}", fixtureRange, "${anchor_name}", strings.TrimSuffix(name, "-updated"), "${reference}", func() string {
			if description == "updated" {
				return "cato_site_web_proxy_section.anchor.id"
			}
			return "null"
		}()).Replace(`resource "cato_site_web_proxy_section" "anchor" {
 at = { position = "LAST_IN_POLICY" }
 section = { name = "acctest-site_web_proxy_section-anchor-${anchor_name}" }
}
resource "cato_site_web_proxy_section" "test" {
 depends_on = [cato_site_web_proxy_section.anchor]
at = { position = "${position}", ref = ${reference} }
section = {name = "acctest-site_web_proxy_section-${name}"}
}
`)
	}
	name := acc.GetRandName("site_web_proxy_section")
	resourceName := "cato_site_web_proxy_section.test"

	published := func(state *terraform.State) error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		revisionType := cato_models.PolicyRevisionTypePublic
		policy, err := acc.GetClient(t).PolicySiteWebProxyPolicy(ctx, acc.CatoAccountID,
			&cato_models.SiteWebProxyPolicyInput{Revision: &cato_models.PolicyRevisionInput{Type: &revisionType}})
		if err != nil {
			return err
		}

		ids := []string{}
		for _, item := range policy.GetPolicy().GetSiteWebProxy().GetPolicy().GetSections() {
			ids = append(ids, item.GetSection().GetID())
		}
		return acc.CheckPolicyPlacement(state, resourceName, "id", ids)
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		CheckDestroy: func(state *terraform.State) error {
			revisionType := cato_models.PolicyRevisionTypePublic
			policy, err := acc.GetClient(t).PolicySiteWebProxyPolicy(context.Background(), acc.CatoAccountID,
				&cato_models.SiteWebProxyPolicyInput{Revision: &cato_models.PolicyRevisionInput{Type: &revisionType}})
			if err != nil {
				return err
			}
			for _, item := range policy.GetPolicy().GetSiteWebProxy().GetPolicy().GetSections() {
				for _, object := range state.RootModule().Resources {
					if object.Type == "cato_site_web_proxy_section" && object.Primary != nil && item.GetSection().GetID() == object.Primary.Attributes["id"] {
						return fmt.Errorf("test resource still exists after destroy")
					}
				}
			}
			return nil
		},
		PreCheck: acc.CheckCMAVars(t),
		Steps: acc.WithStableResourceID([]resource.TestStep{
			{Config: config(name, "initial"), Check: resource.TestCheckResourceAttrSet(resourceName, "id")},
			{Config: config(name, "initial"), PlanOnly: true, ExpectNonEmptyPlan: false},
			{ResourceName: resourceName, ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"at"}},
			{Config: config(name+"-updated", "updated"), Check: resource.TestCheckResourceAttr(resourceName, "section.name", "acctest-site_web_proxy_section-"+name+"-updated")},
			{Config: config(name+"-updated", "updated"), PlanOnly: true, ExpectNonEmptyPlan: false}, {ResourceName: resourceName, ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"at"}},
		}, resourceName, "id", published),
	})
}
