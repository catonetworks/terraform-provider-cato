//go:build acctest

package acc

import "testing"

func TestSkipByEnvUsesTestIdentity(t *testing.T) {
	prior := skipTests
	t.Cleanup(func() { skipTests = prior })
	skipTests = SkipTests{t.Name(): "blocked backend prerequisite"}
	reached := false
	t.Run("through helper", func(t *testing.T) {
		callSkipFromHelper(t)
		reached = true
	})
	if reached {
		t.Fatal("configured test exclusion was ignored through a helper")
	}
	skipTests = SkipTests{}
	t.Run("unlisted test", func(t *testing.T) {
		callSkipFromHelper(t)
		reached = true
	})
	if !reached {
		t.Fatal("unlisted test was skipped")
	}
}
func callSkipFromHelper(t *testing.T) { t.Helper(); SkipByEnv(t) }
