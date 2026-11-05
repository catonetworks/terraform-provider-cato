//go:build acctest

package acc

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// RequireLive must precede fixture API calls, which run before resource.Test's own TF_ACC guard.
func RequireLive(t *testing.T) {
	t.Helper()
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("TF_ACC=1 required before live fixture creation")
	}
	if os.Getenv("DISABLE_POLICY_RULE_CLEANUP") != "true" {
		t.Fatal("DISABLE_POLICY_RULE_CLEANUP=true required to preserve unrelated policy revisions")
	}
	CheckCMAVars(t)()
}

// WithStableResourceID checks real apply steps for accidental replacement, together with
// optional API checks. It leaves plans, imports, refreshes and destroys unchanged.
func WithStableResourceID(steps []resource.TestStep, address, attribute string,
	extra ...resource.TestCheckFunc,
) []resource.TestStep {
	var original string
	stable := func(state *terraform.State) error {
		id, err := stateAttribute(state, address, attribute)
		if err != nil {
			return err
		}
		if original == "" {
			original = id
		} else if id != original {
			return fmt.Errorf("%s unexpectedly replaced %s during an update", address, attribute)
		}
		return nil
	}
	for i := range steps {
		step := &steps[i]
		if step.Config == "" || step.PlanOnly || step.Destroy || step.ImportState || step.RefreshState {
			continue
		}
		checks := []resource.TestCheckFunc{stable}
		if step.Check != nil {
			checks = append(checks, step.Check)
		}
		checks = append(checks, extra...)
		step.Check = resource.ComposeTestCheckFunc(checks...)
	}
	return steps
}

// CheckPolicyPlacement compares Terraform's requested placement with IDs read from
// the published policy, so a successful draft mutation alone cannot satisfy the test.
func CheckPolicyPlacement(state *terraform.State, address, attribute string, ids []string) error {
	id, err := stateAttribute(state, address, attribute)
	if err != nil {
		return err
	}
	index, err := publishedIndex(ids, id)
	if err != nil {
		return err
	}
	position, err := stateAttribute(state, address, "at.position")
	if err != nil {
		return err
	}
	switch position {
	case "FIRST_IN_POLICY":
		if index == 0 {
			return nil
		}
	case "LAST_IN_POLICY":
		if index == len(ids)-1 {
			return nil
		}
	case "BEFORE_RULE", "BEFORE_SECTION", "AFTER_RULE", "AFTER_SECTION":
		ref, refErr := stateAttribute(state, address, "at.ref")
		if refErr != nil {
			return refErr
		}
		other := slices.Index(ids, ref)
		before := strings.HasPrefix(position, "BEFORE_")
		if other >= 0 && ((before && index+1 == other) || (!before && index == other+1)) {
			return nil
		}
	default:
		return fmt.Errorf("unsupported acceptance-test placement %q", position)
	}
	return fmt.Errorf("%s published order does not match %s", address, position)
}

func stateAttribute(state *terraform.State, address, attribute string) (string, error) {
	if state == nil {
		return "", fmt.Errorf("missing resource state for %s", address)
	}
	var root *terraform.ModuleState
	for _, module := range state.Modules {
		if module != nil && slices.Equal(module.Path, []string{"root"}) {
			root = module
			break
		}
	}
	if root == nil {
		return "", fmt.Errorf("missing root module for %s", address)
	}
	item := root.Resources[address]
	if item == nil || item.Primary == nil || item.Primary.Attributes[attribute] == "" {
		return "", fmt.Errorf("missing %s in state for %s", attribute, address)
	}
	return item.Primary.Attributes[attribute], nil
}

func publishedIndex(ids []string, id string) (int, error) {
	seen := make(map[string]bool, len(ids))
	for _, current := range ids {
		if current == "" || seen[current] {
			return -1, fmt.Errorf("published policy returned an empty or duplicate identity")
		}
		seen[current] = true
	}
	index := slices.Index(ids, id)
	if index < 0 {
		return -1, fmt.Errorf("resource is absent from the published policy")
	}
	return index, nil
}
