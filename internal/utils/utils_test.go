package utils

import "testing"

func TestStripANSI(t *testing.T) {
	input := "\033[31mred\033[0m plain"
	if got := StripANSI(input); got != "red plain" {
		t.Fatalf("expected stripped text, got %q", got)
	}
}

func TestTruncateLine(t *testing.T) {
	if got := TruncateLine("short", 10); got != "short" {
		t.Fatalf("expected unchanged line, got %q", got)
	}
	if got := TruncateLine("abcdefghij", 5); got != "abcde..." {
		t.Fatalf("expected truncated line, got %q", got)
	}
}

func TestColorHelpers(t *testing.T) {
	if got := StripANSI(Bold("x")); got != "x" {
		t.Fatalf("expected Bold to preserve visible text, got %q", got)
	}
	if got := StripANSI(Color("x", "red")); got != "x" {
		t.Fatalf("expected Color to preserve visible text, got %q", got)
	}
	if got := StripANSI(Highlight("x", "white", "green")); got != "x" {
		t.Fatalf("expected Highlight to preserve visible text, got %q", got)
	}
	if ColorReset() != "\033[0m" {
		t.Fatalf("unexpected reset sequence")
	}
}
