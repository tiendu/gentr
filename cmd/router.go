package cmd

import (
	"fmt"
	"os"
)

// Command defines the interface all CLI commands must implement.
type Command interface {
	Run(args []string)
}

// CommandFunc is an adapter to allow using ordinary functions as Commands.
type CommandFunc func(args []string)

func (f CommandFunc) Run(args []string) {
	f(args)
}

// CommandRouter routes CLI subcommands to their appropriate handlers.
type CommandRouter struct {
	commands map[string]Command
}

// NewCommandRouter creates a new router with built-in defaults.
func NewCommandRouter() *CommandRouter {
	r := &CommandRouter{commands: make(map[string]Command)}

	// Register built-in commands
	r.Register("install", CommandFunc(InstallCommand))
	r.Register("uninstall", CommandFunc(UninstallCommand))
	r.Register("reinstall", CommandFunc(ReinstallCommand))
	r.Register("version", CommandFunc(func([]string) { VersionCommand() }))
	r.Register("bump", CommandFunc(func([]string) {
		if newVersion, err := BumpVersion(); err != nil {
			fmt.Printf("Version bump failed: %v\n", err)
		} else {
			fmt.Printf("New version: %s\n", newVersion)
		}
	}))
	r.Register("help", CommandFunc(func([]string) { HelpCommand() }))
	r.Register("--help", CommandFunc(func([]string) { HelpCommand() }))
	r.Register("-h", CommandFunc(func([]string) { HelpCommand() }))

	return r
}

// Register adds a new subcommand handler.
func (r *CommandRouter) Register(name string, cmd Command) {
	r.commands[name] = cmd
}

// Run dispatches a command by name, or falls back to help if unknown.
func (r *CommandRouter) Run(name string, args []string) {
	if cmd, ok := r.commands[name]; ok {
		cmd.Run(args)
	} else {
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", name)
		HelpCommand()
		os.Exit(1)
	}
}
