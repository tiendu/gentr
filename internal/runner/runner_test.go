package runner

import (
	"strings"
	"testing"
)

func TestShellRunsCommand(t *testing.T) {
	result := Shell{}.Run("printf hello", "ignored")
	if result.ExitCode != 0 || result.RawOutput != "hello" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestShellReplacesPlaceholder(t *testing.T) {
	result := Shell{}.Run("printf /_", "file.txt")
	if result.ExitCode != 0 || result.RawOutput != "file.txt" || strings.Contains(result.Command, "/_") {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestShellReturnsExitCode(t *testing.T) {
	if result := (Shell{}).Run("exit 7", "ignored"); result.ExitCode != 7 {
		t.Fatalf("expected exit code 7, got %+v", result)
	}
}
