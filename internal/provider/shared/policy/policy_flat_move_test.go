package policy

import "testing"

func TestBuildFlatMoveSequence(t *testing.T) {
	t.Parallel()
	sections := map[string]int64{"B": 2, "A": 1}
	rules := map[string]FlatRuleRow{
		"a2":    {IndexInSection: 2, SectionName: "A"},
		"a1":    {IndexInSection: 1, SectionName: "A"},
		"b1":    {IndexInSection: 1, SectionName: "B"},
		"loose": {IndexInSection: 1, SectionName: "Z"},
	}
	ops := BuildFlatMoveSequence(sections, rules)
	want := []struct {
		section bool
		name    string
	}{
		{true, "A"}, {true, "B"},
		{false, "a1"}, {false, "a2"}, {false, "b1"}, {false, "loose"},
	}
	if len(ops) != len(want) {
		t.Fatalf("got %d ops, want %d: %+v", len(ops), len(want), ops)
	}
	for i, w := range want {
		if ops[i].IsSection != w.section || ops[i].Name != w.name {
			t.Fatalf("op %d: got %+v want %+v", i, ops[i], w)
		}
	}
	if got := BuildFlatMoveSequence(nil, nil); len(got) != 0 {
		t.Fatalf("expected empty, got %+v", got)
	}
}
