//go:build acctest

package site_web_proxy_rule

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

// TestAccSiteWebProxyRule verifies create, idempotent plan, import, and
// update of cato_site_web_proxy_rule (the companion to the pre-existing
// cato_site_web_proxy_section coverage).
func TestAccSiteWebProxyRule(t *testing.T) {
	acc.SkipByEnv(t)
	acc.RequireLive(t)
	acc.CheckCMAVars(t)()
	config := func(name, description, port string) string {
		return acc.ProviderCfg() + strings.NewReplacer(
			"${name}", name,
			"${description}", description,
			"${port}", port,
		).Replace(`resource "cato_site_web_proxy_rule" "test" {
at = { position = "LAST_IN_POLICY" }
rule = {
 name = "acctest-site_web_proxy_rule-${name}"
 description = "${description}"
 enabled = false
 fqdn = "acctest-proxy.example.com"
 port = "${port}"
 should_associate_with_all_sites = true
 authentication_config = { method = "NONE", kerberos_config = { is_enabled = false } }
}
}
`)
	}
	name := acc.GetRandName("site_web_proxy_rule")
	resourceName := "cato_site_web_proxy_rule.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		CheckDestroy: func(state *terraform.State) error {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()
			revisionType := cato_models.PolicyRevisionTypePublic
			policy, err := acc.GetClient(t).PolicySiteWebProxyPolicy(ctx, acc.CatoAccountID,
				&cato_models.SiteWebProxyPolicyInput{Revision: &cato_models.PolicyRevisionInput{Type: &revisionType}})
			if err != nil {
				return err
			}
			for _, item := range policy.GetPolicy().GetSiteWebProxy().GetPolicy().GetRules() {
				for _, object := range state.RootModule().Resources {
					if object.Type == "cato_site_web_proxy_rule" && object.Primary != nil &&
						item.GetRule().GetID() == object.Primary.Attributes["rule.id"] {
						return fmt.Errorf("test resource still exists after destroy")
					}
				}
			}
			return nil
		},
		PreCheck: acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				Config: config(name, "initial", "8080"),
				Check: resource.ComposeAggregateTestCheckFunc(
					acc.PrintAttributes(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "rule.id"),
					resource.TestCheckResourceAttr(resourceName, "rule.fqdn", "acctest-proxy.example.com"),
					resource.TestCheckResourceAttr(resourceName, "rule.port", "8080"),
				),
			},
			{Config: config(name, "initial", "8080"), PlanOnly: true, ExpectNonEmptyPlan: false},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "rule.id",
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					object := state.RootModule().Resources[resourceName]
					if object == nil || object.Primary == nil {
						return "", fmt.Errorf("missing resource state")
					}
					return object.Primary.Attributes["rule.id"], nil
				},
				// at is input-only; the keytab inside authentication_config is
				// write-only and preserved from state rather than the API.
				ImportStateVerifyIgnore: []string{"at", "rule.authentication_config"},
			},
			{
				Config: config(name, "updated", "3128"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "rule.description", "updated"),
					resource.TestCheckResourceAttr(resourceName, "rule.port", "3128"),
				),
			},
			{Config: config(name, "updated", "3128"), PlanOnly: true, ExpectNonEmptyPlan: false},
		},
	})
}
