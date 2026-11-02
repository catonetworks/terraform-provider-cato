//go:build acctest

package firewall_sub_policy

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/catonetworks/terraform-provider-cato/internal/accmock"
	"github.com/catonetworks/terraform-provider-cato/internal/acctests/acc"
)

func TestAccIfSubPolicyPosition(t *testing.T) {
	acc.SkipByEnv(t)
	acc.CleanupFirewallAndWANPolicyRevisions(t)
	defer acc.CleanupFirewallAndWANPolicyRevisions(t)
	mockSrv := accmock.NewMockServer(t, "TestAccIfSubPolicyPosition")
	defer mockSrv.Close()
	mockSrv.Run()

	name := acc.GetRandName("if_sub_policy_position")
	const subPolicy = "cato_if_sub_policy.target"
	identities := map[string]*string{
		"id": new(string), "scope_rule_id": new(string), "cleanup_rule_id": new(string),
	}
	var childID string
	var steps []resource.TestStep
	for i, position := range []string{"AFTER_RULE", "BEFORE_RULE", "AFTER_RULE"} {
		checks := []resource.TestCheckFunc{
			resource.TestCheckResourceAttr(subPolicy, "at.position", position),
			checkSubPolicyScopeOrder(t, position == "BEFORE_RULE"),
		}
		for key, value := range identities {
			if i == 0 {
				checks = append(checks, captureResourceAttr(subPolicy, key, value))
			} else {
				checks = append(checks, checkResourceAttrUnchanged(subPolicy, key, value))
			}
		}
		if i == 0 {
			checks = append(checks, captureResourceAttr("cato_if_rule.child", "rule.id", &childID))
		} else {
			checks = append(checks, checkResourceAttrUnchanged("cato_if_rule.child", "rule.id", &childID))
		}
		steps = append(steps, resource.TestStep{
			Config: ifSubPolicyPositionConfig(name, position),
			Check:  resource.ComposeAggregateTestCheckFunc(checks...),
		})
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps:                    steps,
	})
}

// Verify remote evaluation order, rather than only the position recorded in state.
func checkSubPolicyScopeOrder(t *testing.T, before bool) resource.TestCheckFunc {
	t.Helper()
	return func(state *terraform.State) error {
		target := state.RootModule().Resources["cato_if_sub_policy.target"].Primary.Attributes["scope_rule_id"]
		anchor := state.RootModule().Resources["cato_if_sub_policy.anchor"].Primary.Attributes["scope_rule_id"]
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		response, err := acc.GetClient(t).PolicyInternetFirewallRulesIndex(ctx, acc.CatoAccountID)
		if err != nil {
			return err
		}
		indexes := map[string]int64{}
		for _, row := range response.Policy.InternetFirewall.Policy.Rules {
			indexes[row.Rule.ID] = row.Rule.Index
		}
		targetIndex, targetFound := indexes[target]
		anchorIndex, anchorFound := indexes[anchor]
		if !targetFound || !anchorFound {
			return fmt.Errorf("scope rules missing: target found=%t, anchor found=%t", targetFound, anchorFound)
		}
		if targetIndex == anchorIndex || (targetIndex < anchorIndex) != before {
			return fmt.Errorf("incorrect remote scope order: target index=%d, anchor index=%d, want before=%t", targetIndex, anchorIndex, before)
		}
		return nil
	}
}

func ifSubPolicyPositionConfig(name, position string) string {
	return acc.ProviderCfg() + fmt.Sprintf(`
resource "cato_if_sub_policy" "anchor" {
  name        = %q
  description = "Sub-policy position acceptance anchor"
  at          = { position = "LAST_IN_POLICY" }
  scope = {
    enabled     = false
    source      = { ip = ["10.201.0.10"] }
    destination = {}
  }
}

resource "cato_if_sub_policy" "target" {
  name        = %q
  description = "Sub-policy position acceptance target"
  at = {
    position = %q
    ref      = cato_if_sub_policy.anchor.scope_rule_id
  }
  scope = {
    enabled     = false
    source      = { ip = ["10.201.0.11"] }
    destination = {}
  }
}

resource "cato_if_rule" "child" {
  sub_policy_id = cato_if_sub_policy.target.id
  at            = { position = "LAST_IN_POLICY" }
  rule = {
    name        = %q
    enabled     = false
    action      = "ALLOW"
    source      = { ip = ["10.201.0.12"] }
    destination = {}
    tracking    = { event = { enabled = false } }
  }
}
`, name+"-anchor", name+"-target", position, name+"-child")
}
