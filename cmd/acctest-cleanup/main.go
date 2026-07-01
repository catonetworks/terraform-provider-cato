// cleanup removes dangling acctest_* resources from the Cato staging environment.
// Run with: go run ./cmd/acctest-cleanup/
// Reads CATO_BASEURL, CATO_TOKEN, CATO_ACCOUNT_ID from the environment.
//
// Cleanup order (to respect API dependencies):
//  1. Socket LAN firewall rules  (block global IP range deletion)
//  2. Socket sites               (cascade-removes LAN configs)
//  3. Global IP ranges           (previously blocked by firewall rule references)
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"

	cato "github.com/catonetworks/cato-go-sdk"
	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

var (
	acctestRE = regexp.MustCompile(`(?i)^acctest_`)
	ctx       = context.Background()
)

func main() {
	endpoint := os.Getenv("CATO_BASEURL")
	token := os.Getenv("CATO_TOKEN")
	accountID := os.Getenv("CATO_ACCOUNT_ID")

	if endpoint == "" || token == "" || accountID == "" {
		log.Fatal("CATO_BASEURL, CATO_TOKEN and CATO_ACCOUNT_ID must be set")
	}

	client, err := cato.New(endpoint, token, accountID, nil, map[string]string{"User-Agent": "acctest-cleanup"})
	if err != nil {
		log.Fatalf("create client: %v", err)
	}

	var errs []error
	for _, step := range []struct {
		name string
		fn   func(*cato.Client, string) error
	}{
		{"socket LAN firewall rules", cleanSocketLanFirewallRules},
		{"acctest sites", cleanSites},
		{"global IP ranges", cleanGlobalIPRanges},
	} {
		fmt.Printf("==> cleaning %s...\n", step.name)
		if err := step.fn(client, accountID); err != nil {
			log.Printf("ERROR cleaning %s: %v", step.name, err)
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		log.Fatalf("%d cleanup step(s) failed", len(errs))
	}
	fmt.Println("done.")
}

// cleanSocketLanFirewallRules reads the account-wide socket LAN policy and removes:
//   - every top-level network rule whose name matches acctest_ (via PolicySocketLanRemoveRule)
//   - every firewall sub-rule whose name matches acctest_ (via PolicySocketLanFirewallRemoveRule)
//
// After all deletions a single publish is issued.
func cleanSocketLanFirewallRules(client *cato.Client, accountID string) error {
	policy, err := client.PolicySocketLanPolicy(ctx, accountID, nil)
	if err != nil {
		return fmt.Errorf("read socket LAN policy: %w", err)
	}

	var deleted int
	for _, ruleWrapper := range policy.GetPolicy().GetSocketLan().GetPolicy().GetRules() {
		rule := ruleWrapper.GetRule()

		// Top-level network rule
		if acctestRE.MatchString(rule.GetName()) {
			fmt.Printf("    removing LAN network rule %q (%s)\n", rule.GetName(), rule.GetID())
			removeInput := cato_models.SocketLanRemoveRuleInput{ID: rule.GetID()}
			if _, err := client.PolicySocketLanRemoveRule(ctx, nil, removeInput, accountID); err != nil {
				log.Printf("    WARN: remove network rule %s: %v", rule.GetName(), err)
			} else {
				deleted++
			}
			// Also delete its firewall sub-rules (they go away with the parent, but be safe)
			continue
		}

		// Firewall sub-rules
		for _, fwWrapper := range rule.GetFirewall() {
			name := fwWrapper.GetRule().GetName()
			ruleID := fwWrapper.GetRule().GetID()
			if !acctestRE.MatchString(name) {
				continue
			}
			fmt.Printf("    removing LAN firewall rule %q (%s)\n", name, ruleID)
			removeInput := cato_models.SocketLanFirewallRemoveRuleInput{ID: ruleID}
			if _, err := client.PolicySocketLanFirewallRemoveRule(ctx, accountID, nil, removeInput); err != nil {
				log.Printf("    WARN: remove firewall rule %s: %v", name, err)
				continue
			}
			deleted++
		}
	}

	if deleted == 0 {
		fmt.Println("    no acctest socket LAN rules found")
		return nil
	}

	publishInput := &cato_models.PolicyPublishRevisionInput{}
	if _, err := client.PolicySocketLanPublishPolicyRevision(ctx, nil, publishInput, accountID); err != nil {
		return fmt.Errorf("publish socket LAN policy: %w", err)
	}
	fmt.Printf("    removed %d rule(s) and published policy\n", deleted)
	return nil
}

// cleanSites removes all sites whose name matches acctest_.
func cleanSites(client *cato.Client, accountID string) error {
	siteType := cato_models.EntityType("site")
	limit := int64(500)
	result, err := client.EntityLookup(ctx, accountID, siteType, &limit, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("entityLookup sites: %w", err)
	}

	var deleted int
	for _, item := range result.GetEntityLookup().GetItems() {
		entity := item.GetEntity()
		siteID := entity.GetID()
		siteNamePtr := entity.GetName()
		siteName := ""
		if siteNamePtr != nil {
			siteName = *siteNamePtr
		}
		if !acctestRE.MatchString(siteName) {
			continue
		}
		fmt.Printf("    deleting site %q (%s)\n", siteName, siteID)
		if _, err := client.SiteRemoveSite(ctx, siteID, accountID); err != nil {
			log.Printf("    WARN: delete site %s: %v", siteName, err)
			continue
		}
		deleted++
	}
	fmt.Printf("    deleted %d site(s)\n", deleted)
	return nil
}

// cleanGlobalIPRanges removes all global IP ranges whose name matches acctest_.
func cleanGlobalIPRanges(client *cato.Client, accountID string) error {
	result, err := client.ObjectGlobalIPRangeList(ctx, accountID, nil)
	if err != nil {
		return fmt.Errorf("list global IP ranges: %w", err)
	}

	var toDelete []*cato_models.GlobalIPRangeRefInput
	for _, r := range result.GetObject().GetGlobalIPRangeList().GetItems() {
		if acctestRE.MatchString(r.GetName()) {
			fmt.Printf("    deleting global IP range %q (%s)\n", r.GetName(), r.GetID())
			toDelete = append(toDelete, &cato_models.GlobalIPRangeRefInput{
				By:    cato_models.ObjectRefByID,
				Input: r.GetID(),
			})
		}
	}

	if len(toDelete) == 0 {
		fmt.Println("    no acctest global IP ranges found")
		return nil
	}

	if _, err := client.ObjectDeleteGlobalIPRangeBulk(ctx, accountID, toDelete); err != nil {
		return fmt.Errorf("delete global IP ranges: %w", err)
	}
	fmt.Printf("    deleted %d global IP range(s)\n", len(toDelete))
	return nil
}
