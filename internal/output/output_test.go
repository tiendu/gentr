package output

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tiendu/gentr/internal/config"
	"github.com/tiendu/gentr/internal/runner"
	"github.com/tiendu/gentr/internal/terminal"
)

func TestFormatStatus(t *testing.T) {
	tests := []struct {
		code int
		want string
	}{
		{0, "exit|0|go test"},
		{2, "exit|2|go test"},
		{130, "signal|130|go test"},
	}

	for _, test := range tests {
		got := terminal.StripANSI(formatStatus(runner.Result{ExitCode: test.code, Command: "go test"}))
		if got != test.want {
			t.Fatalf("code %d: expected %q, got %q", test.code, test.want, got)
		}
	}
}

func TestConsoleReporterLimitsOutput(t *testing.T) {
	var output bytes.Buffer
	ConsoleReporter{Writer: &output}.Report(
		runner.Result{RawOutput: "a\nb\nc", ExitCode: 0, Command: "demo"},
		config.New(false, false, ".", 2, false),
	)

	text := terminal.StripANSI(output.String())
	if !strings.Contains(text, "...\nc") || !strings.Contains(text, "exit|0|demo") {
		t.Fatalf("unexpected console output:\n%s", text)
	}
}

func TestSessionLoggerInitAndWrite(t *testing.T) {
	tmp := t.TempDir()
	oldDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldDirectory)

	logger := NewSessionLogger(ioDiscard{})
	logger.now = func() time.Time { return time.Date(2026, 6, 16, 10, 0, 0, 0, time.UTC) }

	if err := logger.Init(config.New(true, false, ".", 3, true), "go test"); err != nil {
		t.Fatal(err)
	}
	if err := logger.Write("file.go:1 ADD: hello", runner.Result{ExitCode: 0}); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(tmp, logger.path))
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, expected := range []string{"# Command: go test", "Output\tExitStatus", "file.go:1 ADD: hello", "ExitStatus: 0"} {
		if !strings.Contains(text, expected) {
			t.Fatalf("expected %q in:\n%s", expected, text)
		}
	}
}

func TestSessionLoggerRejectsWriteBeforeInit(t *testing.T) {
	if err := NewSessionLogger(nil).Write("entry", runner.Result{}); err == nil {
		t.Fatal("expected write before initialization to fail")
	}
}

type ioDiscard struct{}

func (ioDiscard) Write(data []byte) (int, error) { return len(data), nil }
