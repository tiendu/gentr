package internal

import "testing"

func TestDiffLinesDetectsAddAndRemove(t *testing.T) {
	changes := DiffLines(
		[]string{"alpha", "beta", "gamma"},
		[]string{"alpha", "gamma", "delta"},
	)

	if len(changes) != 2 {
		t.Fatalf("expected 2 changes, got %#v", changes)
	}
	if changes[0].Type != "REM" || changes[0].LineNumber != 2 || changes[0].Text != "beta" {
		t.Fatalf("unexpected first change: %#v", changes[0])
	}
	if changes[1].Type != "ADD" || changes[1].LineNumber != 3 || changes[1].Text != "delta" {
		t.Fatalf("unexpected second change: %#v", changes[1])
	}
}

func TestCombineModifications(t *testing.T) {
	changes := []DiffChange{
		{LineNumber: 2, Type: "ADD", Text: "new value"},
		{LineNumber: 2, Type: "REM", Text: "old value"},
		{LineNumber: 5, Type: "ADD", Text: "another"},
	}

	combined := CombineModifications(changes)
	if len(combined) != 2 {
		t.Fatalf("expected 2 changes, got %#v", combined)
	}
	if combined[0].Type != "MOD" {
		t.Fatalf("expected first change to be MOD, got %#v", combined[0])
	}
	if combined[0].Text != "old value -> new value" {
		t.Fatalf("unexpected MOD text: %q", combined[0].Text)
	}
	if combined[1] != changes[2] {
		t.Fatalf("expected second change to be preserved, got %#v", combined[1])
	}
}
