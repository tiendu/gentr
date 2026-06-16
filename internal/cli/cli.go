package cli

import (
	"flag"
	"fmt"
	"io"

	"github.com/tiendu/gentr/internal/buildinfo"
	"github.com/tiendu/gentr/internal/config"
)

type Command interface {
	Run(args []string) int
}

type CommandFunc func(args []string) int

func (f CommandFunc) Run(args []string) int {
	return f(args)
}

type Router struct {
	commands map[string]Command
	stderr   io.Writer
}

func NewRouter(stdout, stderr io.Writer) *Router {
	if stdout == nil {
		stdout = io.Discard
	}
	if stderr == nil {
		stderr = io.Discard
	}

	router := &Router{
		commands: make(map[string]Command),
		stderr:   stderr,
	}

	router.Register("version", CommandFunc(func([]string) int { return Version(stdout) }))
	router.Register("help", CommandFunc(func([]string) int { return Help(stdout) }))
	router.Register("--help", CommandFunc(func([]string) int { return Help(stdout) }))
	router.Register("-h", CommandFunc(func([]string) int { return Help(stdout) }))

	return router
}

func (r *Router) Register(name string, command Command) {
	r.commands[name] = command
}

func (r *Router) Has(name string) bool {
	_, ok := r.commands[name]
	return ok
}

func (r *Router) Run(name string, args []string) int {
	command, ok := r.commands[name]
	if !ok {
		fmt.Fprintf(r.stderr, "Unknown command: %s\n", name)
		return 1
	}

	return command.Run(args)
}

func Parse(args []string) (config.Options, []string, error) {
	var (
		debug      bool
		recursive  bool
		input      string
		length     int
		logEnabled bool
	)

	flags := flag.NewFlagSet("gentr", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	flags.BoolVar(&debug, "debug", false, "Enable debug mode")
	flags.BoolVar(&debug, "d", false, "Enable debug mode (short)")
	flags.BoolVar(&recursive, "recursive", false, "Watch directories recursively")
	flags.BoolVar(&recursive, "r", false, "Watch directories recursively (short)")
	flags.IntVar(&length, "length", 0, "Limit output lines")
	flags.IntVar(&length, "l", 0, "Limit output lines (short)")
	flags.StringVar(&input, "input", ".", "Input path or glob pattern")
	flags.StringVar(&input, "i", ".", "Input path or glob pattern (short)")
	flags.BoolVar(&logEnabled, "log", false, "Enable logging")

	if err := flags.Parse(args); err != nil {
		return config.Options{}, nil, err
	}

	return config.New(debug, recursive, input, length, logEnabled), flags.Args(), nil
}

func Help(writer io.Writer) int {
	fmt.Fprint(writer, `Usage: gentr [options] <command>
       gentr <command>

Commands:
  version      Print version
  help         Show this message

Watch options:
  --debug, -d        Enable debug mode
  --recursive, -r    Watch directories recursively
  --length, -l       Limit output lines
  --log              Enable logging
  --input, -i        Input path or glob pattern

Examples:
  gentr --input 'logs/*.log' 'echo changed /_'
  find testdir -type f | gentr cat /_
  gentr --input . --recursive go test ./...
`)
	return 0
}

func Version(writer io.Writer) int {
	fmt.Fprintf(writer, "gentr version %s, build revision %s\n", buildinfo.Version, buildinfo.Revision)
	return 0
}
