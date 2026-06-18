//go:build acctest

package acc

import (
	"fmt"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// AssertIfwRuleNamesOrderInSection queries the Cato API and checks that rules in the
// given section appear in evaluation order (by ascending API index) as wantOrder.
func AssertIfwRuleNamesOrderInSection(t *testing.T, sectionName string, wantOrder ...string) resource.TestCheckFunc {
	t.Helper()
	return func(*terraform.State) error {
		client := GetClient(t)
		res, err := client.PolicyInternetFirewallRulesIndex(ctx, CatoAccountID)
		if err != nil {
			return fmt.Errorf("PolicyInternetFirewallRulesIndex: %w", err)
		}
		type pair struct {
			idx  int64
			name string
		}
		var pairs []pair
		for _, r := range res.Policy.InternetFirewall.Policy.Rules {
			if r.Rule.Section.Name != sectionName {
				continue
			}
			pairs = append(pairs, pair{idx: r.Rule.Index, name: r.Rule.Name})
		}
		sort.Slice(pairs, func(i, j int) bool {
			if pairs[i].idx != pairs[j].idx {
				return pairs[i].idx < pairs[j].idx
			}
			return pairs[i].name < pairs[j].name
		})
		if len(pairs) != len(wantOrder) {
			return fmt.Errorf("IF section %q: want %d rules, got %d (%+v)", sectionName, len(wantOrder), len(pairs), pairs)
		}
		for i, w := range wantOrder {
			if pairs[i].name != w {
				return fmt.Errorf("IF section %q: at sorted position %d want %q got %q (full %+v)", sectionName, i, w, pairs[i].name, pairs)
			}
		}
		return nil
	}
}

// AssertWanRuleNamesOrderInSection queries the Cato API and checks WAN firewall rule
// order within a section (by ascending API index).
func AssertWanRuleNamesOrderInSection(t *testing.T, sectionName string, wantOrder ...string) resource.TestCheckFunc {
	t.Helper()
	return func(*terraform.State) error {
		client := GetClient(t)
		res, err := client.PolicyWanFirewallRulesIndex(ctx, CatoAccountID)
		if err != nil {
			return fmt.Errorf("PolicyWanFirewallRulesIndex: %w", err)
		}
		type pair struct {
			idx  int64
			name string
		}
		var pairs []pair
		for _, r := range res.Policy.WanFirewall.Policy.Rules {
			if r.Rule.Section.Name != sectionName {
				continue
			}
			pairs = append(pairs, pair{idx: r.Rule.Index, name: r.Rule.Name})
		}
		sort.Slice(pairs, func(i, j int) bool {
			if pairs[i].idx != pairs[j].idx {
				return pairs[i].idx < pairs[j].idx
			}
			return pairs[i].name < pairs[j].name
		})
		if len(pairs) != len(wantOrder) {
			return fmt.Errorf("WAN section %q: want %d rules, got %d (%+v)", sectionName, len(wantOrder), len(pairs), pairs)
		}
		for i, w := range wantOrder {
			if pairs[i].name != w {
				return fmt.Errorf("WAN section %q: at sorted position %d want %q got %q (full %+v)", sectionName, i, w, pairs[i].name, pairs)
			}
		}
		return nil
	}
}
