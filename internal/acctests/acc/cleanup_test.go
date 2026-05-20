//go:build acctest

package acc

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"

	cato_models "github.com/catonetworks/cato-go-sdk/models"
)

var acctestRE = regexp.MustCompile(`^acctest_`)

func TestCleanupAccTestResources(t *testing.T) {
	if os.Getenv("ACCTEST_CLEANUP") != "true" {
		t.Log("Skipping cleanup of test resources. Set ACCTEST_CLEANUP=true to enable.")
		return
	}
	var errors []error
	var helpers = []func(t *testing.T) error{
		deletePrivateAccessRules,
		deleteSites,
	}

	GetClient(t)

	run := func(helper func(t *testing.T) error) {
		if err := helper(t); err != nil {
			errors = append(errors, err)
		}
	}

	for _, helper := range helpers {
		run(helper)
	}

	if len(errors) > 0 {
		t.Fatalf("cleanup errors: %v", errors)
	}
}

func deleteSites(t *testing.T) error {
	sites := getEntities(t, resSite)
	for _, site := range sites {
		if acctestRE.MatchString(site.Name) {
			if _, err := catoClient.SiteRemoveSite(ctx, site.ID, CatoAccountID); err != nil {
				return fmt.Errorf("deleting site %s (%s): %v", site.Name, site.ID, err)
			}
		}
	}
	return nil
}

func deletePrivateAccessRules(t *testing.T) error {
	client := GetClient(t)
	result, err := client.PolicyReadPrivateAccessPolicy(ctx, CatoAccountID)
	if err != nil {
		return err
	}
	rules := result.GetPolicy().GetPrivateAccess().GetPolicy().GetRules()
	if len(rules) == 0 {
		return nil
	}

	for _, rule := range rules {
		if !acctestRE.MatchString(rule.Rule.GetName()) {
			continue
		}

		input := cato_models.PrivateAccessRemoveRuleInput{ID: rule.Rule.ID}
		_, err = client.PolicyPrivateAccessDeleteRule(context.Background(), CatoAccountID, input)
		if err != nil {
			return fmt.Errorf("deleting private access rule %s (%s): %v", rule.Rule.GetName(), rule.Rule.ID, err)
		}
	}
	if _, err = client.PolicyPrivateAccessPublishRevision(ctx, CatoAccountID); err != nil {
		return fmt.Errorf("publishing private access revision: %v", err)
	}

	return nil
}
