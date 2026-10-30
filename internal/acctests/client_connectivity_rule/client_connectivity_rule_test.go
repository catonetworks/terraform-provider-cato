//go:build acctest

package clientconnectivityrule

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

func TestAccClientConnectivityRule(t *testing.T) {
	acc.SkipByEnv(t)
	acc.RequireLive(t)
	fixtureRange := ""

	config := func(name, description string) string {
		return acc.ProviderCfg() + strings.NewReplacer("${name}", name, "${description}", description, "${position}", func() string {
			if description == "updated" {
				return "FIRST_IN_POLICY"
			}
			return "LAST_IN_POLICY"
		}(), "${range}", fixtureRange).Replace(`resource "cato_client_connectivity_rule" "test" {
at = { position = "${position}" }
rule = {"action": "ALLOW", "confidence_level": "HIGH", "connection_origin": [], "country": [], "description": "${description}", "device": [], "enabled": false, "name": "acctest-client_connectivity_rule-${name}", "platform": [], "source": {"user": [], "users_group": []}, "source_range": []}
}
`)
	}
	name := acc.GetRandName("client_connectivity_rule")
	resourceName := "cato_client_connectivity_rule.test"

	published := func(state *terraform.State) error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		revisionType := cato_models.PolicyRevisionTypePublic
		policy, err := acc.GetClient(t).PolicyClientConnectivityPolicy(ctx, acc.CatoAccountID,
			&cato_models.ClientConnectivityPolicyInput{Revision: &cato_models.PolicyRevisionInput{Type: &revisionType}})
		if err != nil {
			return err
		}

		ids := []string{}
		for _, item := range policy.GetPolicy().GetClientConnectivity().GetPolicy().GetRules() {
			ids = append(ids, item.GetRule().GetID())
		}
		return acc.CheckPolicyPlacement(state, resourceName, "rule.id", ids)
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		CheckDestroy: func(state *terraform.State) error {
			revisionType := cato_models.PolicyRevisionTypePublic
			policy, err := acc.GetClient(t).PolicyClientConnectivityPolicy(context.Background(), acc.CatoAccountID,
				&cato_models.ClientConnectivityPolicyInput{Revision: &cato_models.PolicyRevisionInput{Type: &revisionType}})
			if err != nil {
				return err
			}
			for _, item := range policy.GetPolicy().GetClientConnectivity().GetPolicy().GetRules() {
				for _, object := range state.RootModule().Resources {
					if object.Type == "cato_client_connectivity_rule" && object.Primary != nil && item.GetRule().GetID() == object.Primary.Attributes["rule.id"] {
						return fmt.Errorf("test resource still exists after destroy")
					}
				}
			}
			return nil
		},
		PreCheck: acc.CheckCMAVars(t),
		Steps: acc.WithStableResourceID([]resource.TestStep{
			{Config: config(name, "initial"), Check: resource.TestCheckResourceAttrSet(resourceName, "rule.id")},
			{Config: config(name, "initial"), PlanOnly: true, ExpectNonEmptyPlan: false},
			{ResourceName: resourceName, ImportState: true, ImportStateVerify: true, ImportStateVerifyIdentifierAttribute: "rule.id", ImportStateIdFunc: func(state *terraform.State) (string, error) {
				object := state.RootModule().Resources[resourceName]
				if object == nil || object.Primary == nil {
					return "", fmt.Errorf("missing resource state")
				}
				return object.Primary.Attributes["rule.id"], nil
			}, ImportStateVerifyIgnore: []string{"at"}},
			{Config: config(name+"-updated", "updated"), Check: resource.TestCheckResourceAttr(resourceName, "rule.name", "acctest-client_connectivity_rule-"+name+"-updated")},
			{Config: config(name+"-updated", "updated"), PlanOnly: true, ExpectNonEmptyPlan: false}, {ResourceName: resourceName, ImportState: true, ImportStateVerify: true, ImportStateVerifyIdentifierAttribute: "rule.id", ImportStateIdFunc: func(state *terraform.State) (string, error) {
				object := state.RootModule().Resources[resourceName]
				if object == nil || object.Primary == nil {
					return "", fmt.Errorf("missing resource state")
				}
				return object.Primary.Attributes["rule.id"], nil
			}, ImportStateVerifyIgnore: []string{"at"}},
		}, resourceName, "rule.id", published),
	})
}
