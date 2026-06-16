package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestRunDelegatesToApplication(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"version"}, os.Stdin, &stdout, &stderr); code != 0 {
		t.Fatalf("expected version to return 0, got %d", code)
	}
	if !strings.Contains(stdout.String(), "gentr version") {
		t.Fatalf("unexpected output: %q", stdout.String())
	}
}
