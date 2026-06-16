package cli

import (
	"bytes"
	"strings"
	"testing"
)

type recordingCommand struct {
	called bool
	args   []string
	code   int
}

func (c *recordingCommand) Run(args []string) int {
	c.called = true
	c.args = append([]string(nil), args...)
	return c.code
}

func TestParseDefaults(t *testing.T) {
	opts, commandArgs, err := Parse([]string{"echo", "ok"})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if opts.Debug || opts.Recursive || opts.Log || opts.Input != "." || opts.Length != 0 {
		t.Fatalf("unexpected defaults: %+v", opts)
	}
	if strings.Join(commandArgs, " ") != "echo ok" {
		t.Fatalf("unexpected command args: %#v", commandArgs)
	}
}

func TestParseLongAndShortFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"long", []string{"--debug", "--recursive", "--log", "--length", "5", "--input", "./src", "go", "test", "./..."}},
		{"short", []string{"-d", "-r", "--log", "-l", "5", "-i", "./src", "go", "test", "./..."}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			opts, commandArgs, err := Parse(test.args)
			if err != nil {
				t.Fatalf("Parse returned error: %v", err)
			}
			if !opts.Debug || !opts.Recursive || !opts.Log || opts.Length != 5 || opts.Input != "./src" {
				t.Fatalf("unexpected options: %+v", opts)
			}
			if strings.Join(commandArgs, " ") != "go test ./..." {
				t.Fatalf("unexpected command args: %#v", commandArgs)
			}
		})
	}
}

func TestParseRejectsUnknownFlag(t *testing.T) {
	if _, _, err := Parse([]string{"--nope"}); err == nil {
		t.Fatal("expected unknown flag to return an error")
	}
}

func TestRouterRunsRegisteredCommand(t *testing.T) {
	router := NewRouter(ioDiscard{}, ioDiscard{})
	command := &recordingCommand{code: 7}
	router.Register("demo", command)

	if code := router.Run("demo", []string{"a", "b"}); code != 7 {
		t.Fatalf("expected exit code 7, got %d", code)
	}
	if !command.called || strings.Join(command.args, ",") != "a,b" {
		t.Fatalf("unexpected command invocation: %+v", command)
	}
}

func TestRouterRegistersBuiltins(t *testing.T) {
	router := NewRouter(ioDiscard{}, ioDiscard{})
	for _, name := range []string{"version", "help", "--help", "-h"} {
		if !router.Has(name) {
			t.Fatalf("expected router to register %q", name)
		}
	}
}

func TestHelpAndVersionWriteOutput(t *testing.T) {
	var output bytes.Buffer
	if Help(&output) != 0 || !strings.Contains(output.String(), "Usage: gentr") {
		t.Fatalf("unexpected help output: %q", output.String())
	}

	output.Reset()
	if Version(&output) != 0 || !strings.Contains(output.String(), "gentr version") {
		t.Fatalf("unexpected version output: %q", output.String())
	}
}

type ioDiscard struct{}

func (ioDiscard) Write(data []byte) (int, error) { return len(data), nil }
