//go:build acctest

package private_access_policy

import (
	"os"
	"strings"
	"testing"
)

// These legacy resources share the account's policy draft. Concurrent rule
// creates/deletes can discard each other's unpublished changes, so each test
// serializes Terraform mutations while separate account groups remain parallel.
func serializePolicyMutations(t *testing.T) {
	t.Helper()
	for _, command := range []string{"apply", "destroy"} {
		key := "TF_CLI_ARGS_" + command
		t.Setenv(key, strings.TrimSpace(os.Getenv(key)+" -parallelism=1"))
	}
}
