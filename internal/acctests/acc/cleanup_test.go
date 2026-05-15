//go:build acctest

package acc

import (
	"fmt"
	"os"
	"regexp"
	"testing"
)

var acctestRE = regexp.MustCompile(`^acctest_`)

func TestCleanupAccTestResources(t *testing.T) {
	if os.Getenv("ACCTEST_CLEANUP") != "true" {
		t.Log("Skipping cleanup of test resources. Set ACCTEST_CLEANUP=true to enable.")
		return
	}
	var errors []error

	GetClient(t)

	if err := deleteSites(t); err != nil {
		errors = append(errors, err)
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
