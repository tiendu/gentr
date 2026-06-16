package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gentr/internal/utils"
)

func TestFormatStatusLogSuccess(t *testing.T) {
	got := utils.StripANSI(formatStatusLog(CommandResult{ExitCode: 0, Command: "go test"}))
	if got != "exit|0|go test" {
		t.Fatalf("unexpected status log: %q", got)
	}
}

func TestFormatStatusLogFailure(t *testing.T) {
	got := utils.StripANSI(formatStatusLog(CommandResult{ExitCode: 2, Command: "go test"}))
	if got != "exit|2|go test" {
		t.Fatalf("unexpected status log: %q", got)
	}
}

func TestFormatStatusLogSignal(t *testing.T) {
	got := utils.StripANSI(formatStatusLog(CommandResult{ExitCode: 130, Command: "go test"}))
	if got != "signal|130|go test" {
		t.Fatalf("unexpected status log: %q", got)
	}
}

func TestSessionLoggerWriteBeforeInitReturnsError(t *testing.T) {
	logger := &SessionLogger{}
	if err := logger.Write("entry", CommandResult{}); err == nil {
		t.Fatal("expected write before init to fail")
	}
}

func TestSessionLoggerInitAndWrite(t *testing.T) {
	tmp := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir tmp: %v", err)
	}
	defer os.Chdir(oldWd)

	logger := &SessionLogger{}
	opts := NewWatchOptions(true, false, ".", 3, true)

	if err := logger.Init(opts, "go test"); err != nil {
		t.Fatalf("Init returned error: %v", err)
	}
	if logger.path == "" {
		t.Fatal("expected logger path to be set")
	}
	if err := logger.Write("file.go:1 ADD: hello", CommandResult{ExitCode: 0}); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tmp, logger.path))
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	text := string(data)
	for _, expected := range []string{"# Command: go test", "Output\tExitStatus", "file.go:1 ADD: hello", "ExitStatus: 0"} {
		if !strings.Contains(text, expected) {
			t.Fatalf("expected log to contain %q, got:\n%s", expected, text)
		}
	}
}
