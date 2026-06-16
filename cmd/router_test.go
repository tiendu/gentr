package cmd

import "testing"

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

func TestNewCommandRouterRegistersAdminCommands(t *testing.T) {
	router := NewCommandRouter()

	for _, name := range []string{"version", "help", "--help", "-h"} {
		if !router.Has(name) {
			t.Fatalf("expected router to have command %q", name)
		}
	}
}

func TestCommandRouterRunRegisteredCommand(t *testing.T) {
	router := &CommandRouter{commands: make(map[string]Command)}
	cmd := &recordingCommand{code: 7}
	router.Register("demo", cmd)

	code := router.Run("demo", []string{"a", "b"})
	if code != 7 {
		t.Fatalf("expected exit code 7, got %d", code)
	}
	if !cmd.called {
		t.Fatal("expected command to be called")
	}
	if len(cmd.args) != 2 || cmd.args[0] != "a" || cmd.args[1] != "b" {
		t.Fatalf("unexpected args: %#v", cmd.args)
	}
}

func TestCommandRouterRunUnknownCommand(t *testing.T) {
	router := &CommandRouter{commands: make(map[string]Command)}
	code := router.Run("missing", nil)
	if code != 1 {
		t.Fatalf("expected unknown command to return 1, got %d", code)
	}
}
