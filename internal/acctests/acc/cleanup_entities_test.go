//go:build acctest

package acc

import (
	"reflect"
	"testing"
)

func TestSelectAcctestRefs(t *testing.T) {
	t.Parallel()

	refs := []Ref{
		{ID: "1", Name: "acctest_site"},
		{ID: "2", Name: "production_site"},
		{ID: "", Name: "acctest_missing_id"},
		{ID: "root", Name: "acctest_root"},
	}
	got := selectAcctestRefs(refs, "root")
	want := []Ref{{ID: "1", Name: "acctest_site"}}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("selectAcctestRefs() = %#v, want %#v", got, want)
	}
}

func TestSameIDSet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		left  []string
		right []string
		want  bool
	}{
		{name: "same unordered IDs", left: []string{"1", "2"}, right: []string{"2", "1"}, want: true},
		{name: "different IDs", left: []string{"1", "2"}, right: []string{"1", "3"}},
		{name: "different lengths", left: []string{"1"}, right: []string{"1", "2"}},
		{name: "blank ID", left: []string{""}, right: []string{""}},
		{name: "duplicate expected ID", left: []string{"1", "1"}, right: []string{"1", "1"}},
		{name: "duplicate returned ID", left: []string{"1", "2"}, right: []string{"1", "1"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := sameIDSet(test.left, test.right); got != test.want {
				t.Fatalf("sameIDSet(%v, %v) = %t, want %t", test.left, test.right, got, test.want)
			}
		})
	}
}
