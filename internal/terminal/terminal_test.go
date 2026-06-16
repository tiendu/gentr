package terminal

import "testing"

func TestStripANSI(t *testing.T) {
	input := "\x1b[31mred\x1b[0m plain"
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
	if got := TruncateLine("你好世界", 2); got != "你好..." {
		t.Fatalf("expected rune-safe truncation, got %q", got)
	}
}

func TestColorHelpersPreserveVisibleText(t *testing.T) {
	for name, value := range map[string]string{
		"bold":      Bold("x"),
		"color":     Color("x", "red"),
		"highlight": Highlight("x", "white", "green"),
	} {
		if got := StripANSI(value); got != "x" {
			t.Fatalf("%s: expected visible text x, got %q", name, got)
		}
	}

	if Reset() != "\x1b[0m" {
		t.Fatalf("unexpected reset sequence")
	}
}
