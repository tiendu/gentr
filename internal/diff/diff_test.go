package diff

import "testing"

func TestLinesDetectsAddAndRemove(t *testing.T) {
	changes := Lines(
		[]string{"alpha", "beta", "gamma"},
		[]string{"alpha", "gamma", "delta"},
	)

	if len(changes) != 2 {
		t.Fatalf("expected 2 changes, got %#v", changes)
	}
	if changes[0] != (Change{LineNumber: 2, Kind: Removed, Text: "beta"}) {
		t.Fatalf("unexpected first change: %#v", changes[0])
	}
	if changes[1] != (Change{LineNumber: 3, Kind: Added, Text: "delta"}) {
		t.Fatalf("unexpected second change: %#v", changes[1])
	}
}

func TestCombineModifications(t *testing.T) {
	changes := []Change{
		{LineNumber: 2, Kind: Added, Text: "new value"},
		{LineNumber: 2, Kind: Removed, Text: "old value"},
		{LineNumber: 5, Kind: Added, Text: "another"},
	}

	combined := CombineModifications(changes)
	if len(combined) != 2 || combined[0].Kind != Modified || combined[0].Text != "old value -> new value" {
		t.Fatalf("unexpected combined changes: %#v", combined)
	}
}
