package internal

import (
	"strings"
	"testing"
)

func TestShellRunnerRunsCommand(t *testing.T) {
	result := ShellRunner{}.Run("printf hello", "ignored")
	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", result.ExitCode)
	}
	if result.RawOutput != "hello" {
		t.Fatalf("unexpected output: %q", result.RawOutput)
	}
}

func TestShellRunnerReplacesFilePlaceholder(t *testing.T) {
	result := ShellRunner{}.Run("printf /_", "file.txt")
	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", result.ExitCode)
	}
	if result.RawOutput != "file.txt" {
		t.Fatalf("unexpected output: %q", result.RawOutput)
	}
	if strings.Contains(result.Command, "/_") {
		t.Fatalf("expected placeholder to be replaced in command, got %q", result.Command)
	}
}

func TestShellRunnerReturnsExitCode(t *testing.T) {
	result := ShellRunner{}.Run("exit 7", "ignored")
	if result.ExitCode != 7 {
		t.Fatalf("expected exit code 7, got %d", result.ExitCode)
	}
}
