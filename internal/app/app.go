package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/tiendu/gentr/internal/cli"
	"github.com/tiendu/gentr/internal/config"
	inputpkg "github.com/tiendu/gentr/internal/input"
	"github.com/tiendu/gentr/internal/output"
	"github.com/tiendu/gentr/internal/runner"
	"github.com/tiendu/gentr/internal/spinner"
	"github.com/tiendu/gentr/internal/watch"
)

func Run(args []string, stdin *os.File, stdout, stderr io.Writer) int {
	router := cli.NewRouter(stdout, stderr)
	if len(args) == 0 {
		return cli.Help(stdout)
	}
	if router.Has(args[0]) {
		return router.Run(args[0], args[1:])
	}

	opts, commandArgs, err := cli.Parse(args)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintln(stdout, "Starting with options:", opts)

	resolver := inputpkg.FileResolver{}
	files, err := selectInput(stdin, resolver, inputpkg.LineStdinReader{}, opts)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if len(files) == 0 {
		fmt.Fprintln(stderr, "No files provided via STDIN or --input flag")
		return 1
	}
	if len(commandArgs) == 0 {
		fmt.Fprintln(stderr, "No command provided to execute")
		return 1
	}

	command := strings.Join(commandArgs, " ")
	logger := output.NewSessionLogger(stdout)
	if opts.Log {
		if err := logger.Init(opts, command); err != nil {
			fmt.Fprintf(stderr, "[x] Error initializing log file: %v\n", err)
			return 1
		}
	}

	activity := spinner.NewSnake(30, 5, 81, stdout)
	activity.Start()
	defer activity.Stop()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	watcher := watch.New(
		opts,
		activity,
		runner.Shell{},
		output.ConsoleReporter{Writer: stdout},
		logger,
		resolver,
		stdout,
	)
	watcher.Run(ctx, files, command)
	fmt.Fprintln(stdout, "\nShutting down gentr...")
	return 0
}

func selectInput(
	stdin *os.File,
	resolver inputpkg.Resolver,
	stdinReader inputpkg.StdinReader,
	opts config.Options,
) ([]string, error) {
	info, err := stdin.Stat()
	if err == nil && info.Mode()&os.ModeCharDevice == 0 {
		return stdinReader.ReadFiles(stdin), nil
	}
	return resolver.Resolve(opts.Input, opts.Recursive)
}
