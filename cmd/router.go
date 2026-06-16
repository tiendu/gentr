package cmd

import (
	"fmt"
	"os"
)

type Command interface {
	Run(args []string) int
}

type CommandFunc func(args []string) int

func (f CommandFunc) Run(args []string) int {
	return f(args)
}

type CommandRouter struct {
	commands map[string]Command
}

func NewCommandRouter() *CommandRouter {
	r := &CommandRouter{commands: make(map[string]Command)}

	r.Register("install", CommandFunc(InstallCommand))
	r.Register("uninstall", CommandFunc(UninstallCommand))
	r.Register("version", CommandFunc(func([]string) int { return VersionCommand() }))
	r.Register("help", CommandFunc(func([]string) int { return HelpCommand() }))
	r.Register("--help", CommandFunc(func([]string) int { return HelpCommand() }))
	r.Register("-h", CommandFunc(func([]string) int { return HelpCommand() }))

	return r
}

func (r *CommandRouter) Register(name string, cmd Command) {
	r.commands[name] = cmd
}

func (r *CommandRouter) Has(name string) bool {
	_, ok := r.commands[name]
	return ok
}

func (r *CommandRouter) Run(name string, args []string) int {
	command, ok := r.commands[name]
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", name)
		HelpCommand()
		return 1
	}

	return command.Run(args)
}
