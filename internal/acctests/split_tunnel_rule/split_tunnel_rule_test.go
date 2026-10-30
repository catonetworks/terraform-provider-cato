//go:build acctest

package splittunnelrule

import (
	"context"
	"crypto/rand"
	"fmt"
	"strings"
	"testing"
	"time"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccSplitTunnelRule(t *testing.T) {
	acc.SkipByEnv(t)
	acc.RequireLive(t)
	fixtureRange := ""

	config := func(name, description string) string {
		return acc.ProviderCfg() + strings.NewReplacer("${name}", name, "${description}", description, "${position}", func() string {
			if description == "updated" {
				return "FIRST_IN_POLICY"
			}
			return "LAST_IN_POLICY"
		}(), "${range}", fixtureRange, "${dns}", func() string {
			if description == "dns" {
				return `{domain = ["configured.terraform.invalid"]}`
			}
			return "null"
		}()).Replace(`resource "cato_split_tunnel_rule" "test" {
at = { position = "${position}" }
rule = {"action": "ROUTE_ONLY_SELECTED", "country": [], "coverage": "ALL", "description": "${description}", "destination": {"application": [], "domain": ["selected.terraform.invalid"], "fqdn": [], "global_ip_range": []}, "enabled": false, "name": "acctest-split_tunnel_rule-${name}", "platform": ["WINDOWS"], "routing_priority": "LAN", "source": {"user": [], "users_group": []}, "source_network": {"source_network_type": "ANY"}, "dns_exclusion": ${dns}}
}
`)
	}
	name := acc.GetRandName("split_tunnel_rule")
	resourceName := "cato_split_tunnel_rule.test"
	var createdID string

	published := func(state *terraform.State) error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		revisionType := cato_models.PolicyRevisionTypePublic
		policy, err := acc.GetClient(t).PolicySplitTunnelPolicy(ctx, acc.CatoAccountID,
			&cato_models.SplitTunnelPolicyInput{Revision: &cato_models.PolicyRevisionInput{Type: &revisionType}})
		if err != nil {
			return err
		}

		ids := []string{}
		for _, item := range policy.GetPolicy().GetSplitTunnel().GetPolicy().GetRules() {
			ids = append(ids, item.GetRule().GetID())
		}
		return acc.CheckPolicyPlacement(state, resourceName, "rule.id", ids)
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		CheckDestroy: func(state *terraform.State) error {
			revisionType := cato_models.PolicyRevisionTypePublic
			policy, err := acc.GetClient(t).PolicySplitTunnelPolicy(context.Background(), acc.CatoAccountID,
				&cato_models.SplitTunnelPolicyInput{Revision: &cato_models.PolicyRevisionInput{Type: &revisionType}})
			if err != nil {
				return err
			}
			for _, item := range policy.GetPolicy().GetSplitTunnel().GetPolicy().GetRules() {
				for _, object := range state.RootModule().Resources {
					if object.Type == "cato_split_tunnel_rule" && object.Primary != nil && item.GetRule().GetID() == object.Primary.Attributes["rule.id"] {
						return fmt.Errorf("test resource still exists after destroy")
					}
				}
			}
			return nil
		},
		PreCheck: acc.CheckCMAVars(t),
		Steps: acc.WithStableResourceID([]resource.TestStep{
			{Config: config(name, "initial"), Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttrSet(resourceName, "rule.id"), func(state *terraform.State) error {
				createdID = state.RootModule().Resources[resourceName].Primary.Attributes["rule.id"]
				return nil
			})},
			{Config: config(name, "initial"), PlanOnly: true, ExpectNonEmptyPlan: false},
			{ResourceName: resourceName, ImportState: true, ImportStateVerify: true, ImportStateVerifyIdentifierAttribute: "rule.id", ImportStateIdFunc: func(state *terraform.State) (string, error) {
				object := state.RootModule().Resources[resourceName]
				if object == nil || object.Primary == nil {
					return "", fmt.Errorf("missing resource state")
				}
				return object.Primary.Attributes["rule.id"], nil
			}, ImportStateVerifyIgnore: []string{"at"}},
			{Config: config(name+"-updated", "updated"), Check: resource.TestCheckResourceAttr(resourceName, "rule.name", "acctest-split_tunnel_rule-"+name+"-updated")},
			{Config: config(name+"-updated", "updated"), PlanOnly: true, ExpectNonEmptyPlan: false},
			{Config: config(name+"-updated", "dns"), Check: resource.TestCheckResourceAttr(resourceName, "rule.dns_exclusion.domain.#", "1")},
			{Config: config(name+"-updated", "dns"), PlanOnly: true, ExpectNonEmptyPlan: false},
			{Config: config(name+"-updated", "updated")},
			{Config: config(name+"-updated", "updated"), PlanOnly: true, ExpectNonEmptyPlan: false},
			{PreConfig: func() { addExternalDNSExclusion(t, createdID) }, Config: config(name+"-updated", "updated"), PlanOnly: true, ExpectNonEmptyPlan: true},
			{Config: config(name+"-updated", "updated")},
			{Config: config(name+"-updated", "updated"), PlanOnly: true, ExpectNonEmptyPlan: false}, {ResourceName: resourceName, ImportState: true, ImportStateVerify: true, ImportStateVerifyIdentifierAttribute: "rule.id", ImportStateIdFunc: func(state *terraform.State) (string, error) {
				object := state.RootModule().Resources[resourceName]
				if object == nil || object.Primary == nil {
					return "", fmt.Errorf("missing resource state")
				}
				return object.Primary.Attributes["rule.id"], nil
			}, ImportStateVerifyIgnore: []string{"at"}},
		}, resourceName,

			// Change only the rule created by this test, then let Terraform detect and repair drift.
			"rule.id", published),
	})
}

func addExternalDNSExclusion(t *testing.T, id string) {
	t.Helper()
	if id == "" {
		t.Fatal("missing test rule identity")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	client := acc.GetClient(t)
	created, err := client.PolicySplitTunnelCreatePolicyRevision(ctx, acc.CatoAccountID, nil,
		cato_models.PolicyCreateRevisionInput{Name: "terraform-drift-" + rand.Text()})
	payload := created.GetPolicy().GetSplitTunnel().GetCreatePolicyRevision()
	checkDriftMutation(t, "create drift-test revision", err, payload.GetStatus(), payload.GetErrors())
	revision := payload.GetPolicy().GetRevision().GetID()
	if revision == "" {
		t.Fatal("missing drift-test revision identity")
	}
	input := &cato_models.SplitTunnelPolicyMutationInput{Revision: &cato_models.PolicyMutationRevisionInput{ID: &revision}}
	published := false
	defer func() {
		if !published {
			_, discardErr := client.PolicySplitTunnelDiscardPolicyRevision(ctx, acc.CatoAccountID, input, &cato_models.PolicyDiscardRevisionInput{ID: &revision})
			if discardErr != nil {
				t.Errorf("discard drift-test revision: %v", discardErr)
			}
		}
	}()
	updated, err := client.PolicySplitTunnelUpdateRule(ctx, acc.CatoAccountID, input, cato_models.SplitTunnelUpdateRuleInput{
		ID: id, Rule: &cato_models.SplitTunnelUpdateRuleDataInput{DNSExclusion: &cato_models.SplitTunnelDNSExclusionUpdateInput{Domain: []string{"external.terraform.invalid"}}},
	})
	update := updated.GetPolicy().GetSplitTunnel().GetUpdateRule()
	checkDriftMutation(t, "set external DNS exclusion", err, update.GetStatus(), update.GetErrors())
	result, err := client.PolicySplitTunnelPublishPolicyRevision(ctx, acc.CatoAccountID, input, &cato_models.PolicyPublishRevisionInput{})
	publication := result.GetPolicy().GetSplitTunnel().GetPublishPolicyRevision()
	checkDriftMutation(t, "publish external DNS exclusion", err, publication.GetStatus(), publication.GetErrors())
	published = true
}

func checkDriftMutation[T interface {
	GetErrorMessage() *string
	GetErrorCode() *string
}](t *testing.T, summary string, err error, status *cato_models.PolicyMutationStatus, failures []T) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", summary, err)
	}
	for _, failure := range failures {
		if message := failure.GetErrorMessage(); message != nil {
			t.Errorf("%s: %s", summary, *message)
		}
		if code := failure.GetErrorCode(); code != nil {
			t.Errorf("%s: %s", summary, *code)
		}
	}
	if len(failures) > 0 || status == nil || *status != cato_models.PolicyMutationStatusSuccess {
		t.Fatalf("%s did not succeed", summary)
	}
}
