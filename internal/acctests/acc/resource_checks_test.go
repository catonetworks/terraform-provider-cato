//go:build acctest

package acc

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"
)

func fixtureState(id, position, ref string) *terraform.State {
	return &terraform.State{Modules: []*terraform.ModuleState{{Path: []string{"root"}, Resources: map[string]*terraform.ResourceState{
		"cato_example.test": {Primary: &terraform.InstanceState{ID: id, Attributes: map[string]string{
			"id": id, "rule.id": id, "at.position": position, "at.ref": ref,
		}}},
	}}}}
}

func TestWithStableResourceID(t *testing.T) {
	checks := 0
	steps := WithStableResourceID([]resource.TestStep{
		{Config: "create", Check: func(*terraform.State) error { checks++; return nil }},
		{Config: "plan", PlanOnly: true},
		{ImportState: true},
		{Config: "update"},
		{Config: "destroy", Destroy: true},
		{RefreshState: true},
	}, "cato_example.test", "rule.id")
	require.NoError(t, steps[0].Check(fixtureState("one", "", "")))
	require.Equal(t, 1, checks)
	require.NoError(t, steps[3].Check(fixtureState("one", "", "")))
	require.Error(t, steps[3].Check(fixtureState("replacement", "", "")))
	require.Error(t, steps[3].Check(fixtureState("", "", "")))
	require.Error(t, steps[3].Check(&terraform.State{}))
	for _, i := range []int{1, 2, 4, 5} {
		require.Nil(t, steps[i].Check)
	}
}

func TestCheckPolicyPlacement(t *testing.T) {
	for _, tc := range []struct {
		name, position, ref string
		ids                 []string
		wantError           bool
	}{
		{"first", "FIRST_IN_POLICY", "", []string{"owned", "anchor"}, false},
		{"last", "LAST_IN_POLICY", "", []string{"anchor", "owned"}, false},
		{"before section", "BEFORE_SECTION", "anchor", []string{"owned", "anchor"}, false},
		{"after section", "AFTER_SECTION", "anchor", []string{"anchor", "owned"}, false},
		{"before rule", "BEFORE_RULE", "anchor", []string{"owned", "anchor"}, false},
		{"after rule", "AFTER_RULE", "anchor", []string{"anchor", "owned"}, false},
		{"not published", "LAST_IN_POLICY", "", []string{"anchor"}, true},
		{"wrong order", "BEFORE_SECTION", "anchor", []string{"anchor", "owned"}, true},
		{"missing reference", "BEFORE_SECTION", "missing", []string{"owned", "anchor"}, true},
		{"duplicate identity", "LAST_IN_POLICY", "", []string{"owned", "owned"}, true},
		{"incomplete response", "LAST_IN_POLICY", "", []string{"", "owned"}, true},
		{"unsupported position", "INVALID", "", []string{"owned"}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := CheckPolicyPlacement(fixtureState("owned", tc.position, tc.ref), "cato_example.test", "id", tc.ids)
			if tc.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRequireLiveSkipsBeforeFixtures(t *testing.T) {
	t.Setenv("TF_ACC", "")
	fixtureCalled := false
	t.Run("not opted in", func(t *testing.T) {
		RequireLive(t)
		fixtureCalled = true
	})
	require.False(t, fixtureCalled, "fixtures must never run without live-test opt-in")
}
