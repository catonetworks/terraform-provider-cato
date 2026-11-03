//go:build acctest

package bulk_st_move_rule

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

// TestAccBulkStMoveRule verifies that the sequential bulk mover orders two rules
// inside a managed section by index_in_section and republishes, and that a
// swapped ordering is applied on update.
func TestAccBulkStMoveRule(t *testing.T) {
	acc.SkipByEnv(t)
	acc.RequireLive(t)
	acc.CheckCMAVars(t)()

	name := acc.GetRandName("bulk_st_move_rule")
	sectionName := "acctest-bulk_st_move_rule-" + name
	ruleA := "acctest-bulk_st_move_rule-A-" + name
	ruleB := "acctest-bulk_st_move_rule-B-" + name

	ruleBlock := func(resName, ruleName string, first bool) string {
		return strings.NewReplacer(
			"${res}", resName,
			"${rule_name}", ruleName,
			"${domain}", map[bool]string{true: "mover-a.example.com", false: "mover-b.example.com"}[first],
		).Replace(`resource "cato_split_tunnel_rule" "${res}" {
 depends_on = [cato_split_tunnel_section.s]
 at = { position = "LAST_IN_SECTION", ref = cato_split_tunnel_section.s.id }
 rule = {
   action = "EXCLUDE"
   description = "mover acctest"
   destination = { application = [], domain = ["${domain}"], fqdn = [], global_ip_range = [] }
   enabled = false
   name = "${rule_name}"
   routing_priority = "LAN"
   source = { user = [], users_group = [] }
   source_network = { source_network_type = "ANY" }
 }
}
`)
	}

	config := func(aFirst bool) string {
		idxA, idxB := "1", "2"
		if !aFirst {
			idxA, idxB = "2", "1"
		}
		return acc.ProviderCfg() + `resource "cato_split_tunnel_section" "s" {
 at = { position = "LAST_IN_POLICY" }
 section = { name = "` + sectionName + `" }
}
` + ruleBlock("a", ruleA, true) + ruleBlock("b", ruleB, false) + `
resource "cato_bulk_st_move_rule" "all" {
 depends_on = [cato_split_tunnel_rule.a, cato_split_tunnel_rule.b]
 section_data = { "` + sectionName + `" = { section_index = 1 } }
 rule_data = {
   "` + ruleA + `" = { index_in_section = ` + idxA + `, section_name = "` + sectionName + `" }
   "` + ruleB + `" = { index_in_section = ` + idxB + `, section_name = "` + sectionName + `" }
 }
}
`
	}

	orderCheck := func(firstRule, secondRule string) resource.TestCheckFunc {
		return func(_ *terraform.State) error {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()
			revisionType := cato_models.PolicyRevisionTypePublic
			policy, err := acc.GetClient(t).PolicySplitTunnelPolicy(ctx, acc.CatoAccountID, &cato_models.SplitTunnelPolicyInput{Revision: &cato_models.PolicyRevisionInput{Type: &revisionType}})
			if err != nil {
				return err
			}
			idxByName := map[string]int64{}
			for _, item := range policy.GetPolicy().GetSplitTunnel().GetPolicy().GetRules() {
				idxByName[item.GetRule().GetName()] = item.GetRule().GetIndex()
			}
			a, okA := idxByName[firstRule]
			b, okB := idxByName[secondRule]
			if !okA || !okB {
				return fmt.Errorf("published policy missing mover rules: %v", idxByName)
			}
			if a >= b {
				return fmt.Errorf("published order wrong: %s(%d) should precede %s(%d)", firstRule, a, secondRule, b)
			}
			return nil
		}
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 acc.CheckCMAVars(t),
		Steps: []resource.TestStep{
			{
				// A before B.
				Config: config(true),
				Check:  resource.ComposeAggregateTestCheckFunc(orderCheck(ruleA, ruleB)),
			},
			{Config: config(true), PlanOnly: true, ExpectNonEmptyPlan: false},
			{
				// Swap: B before A — the mover must re-apply ordering and publish.
				Config: config(false),
				Check:  resource.ComposeAggregateTestCheckFunc(orderCheck(ruleB, ruleA)),
			},
		},
	})
}
