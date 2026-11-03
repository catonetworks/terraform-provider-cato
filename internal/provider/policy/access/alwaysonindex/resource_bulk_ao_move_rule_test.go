package alwaysonindex

import "testing"

func TestBuildAoMoveSequence(t *testing.T) {
	t.Parallel()
	sections := map[string]int64{"B Section": 2, "A Section": 1}
	rules := map[string]AoRuleRow{
		"A2":    {IndexInSection: 2, SectionName: "A Section"},
		"A1":    {IndexInSection: 1, SectionName: "A Section"},
		"B1":    {IndexInSection: 1, SectionName: "B Section"},
		"Loose": {IndexInSection: 1, SectionName: "Z Unmanaged"},
	}
	ops := BuildAoMoveSequence(sections, rules)
	var got []string
	for _, op := range ops {
		kind := "rule"
		if op.IsSection {
			kind = "section"
		}
		got = append(got, kind+":"+op.Name)
	}
	want := []string{
		"section:A Section", "section:B Section",
		"rule:A1", "rule:A2", "rule:B1", "rule:Loose",
	}
	if len(got) != len(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("op %d: got %v want %v (full: %v)", i, got[i], want[i], got)
		}
	}
}

func TestBuildAoMoveSequenceEmpty(t *testing.T) {
	t.Parallel()
	if ops := BuildAoMoveSequence(nil, nil); len(ops) != 0 {
		t.Fatalf("expected no ops, got %v", ops)
	}
}
