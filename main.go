package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"gentr/cmd"
	"gentr/internal"
)

func main() {
	raise := run(os.Args[1:], os.Stdin)
	os.Exit(raise)
}

func run(args []string, stdin *os.File) int {
	router := cmd.NewCommandRouter()

	if len(args) == 0 {
		return cmd.HelpCommand()
	}

	if router.Has(args[0]) {
		return router.Run(args[0], args[1:])
	}

	opts, commandArgs, err := cmd.ParseOptions(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	watchOpts := internal.NewWatchOptions(
		opts.Debug,
		opts.Recursive,
		opts.Input,
		opts.Length,
		opts.Log,
	)

	fmt.Println("Starting with options:", opts)

	files, err := inputFiles(stdin, internal.FileResolver{}, internal.LineStdinReader{}, watchOpts)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "No files provided via STDIN or --input flag")
		return 1
	}

	if len(commandArgs) == 0 {
		fmt.Fprintln(os.Stderr, "No command provided to execute")
		return 1
	}

	command := strings.Join(commandArgs, " ")
	logger := &internal.SessionLogger{}

	if watchOpts.Log {
		if err := logger.Init(watchOpts, command); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing log file: %v\n", err)
		}
	}

	spinner := cmd.NewSnakeSpinner(30, 5, 81)
	spinner.Start()

	watcher := internal.NewWatcher(
		watchOpts,
		spinner,
		internal.ShellRunner{},
		internal.ConsoleReporter{},
		logger,
	)
	go watcher.Watch(files, command)

	waitForShutdown()
	spinner.Stop()
	fmt.Println("\nShutting down gentr...")

	return 0
}

func inputFiles(
	stdin *os.File,
	resolver internal.Resolver,
	stdinReader internal.StdinReader,
	opts internal.WatchOptions,
) ([]string, error) {
	info, err := stdin.Stat()
	if err == nil && (info.Mode()&os.ModeCharDevice) == 0 {
		return stdinReader.ReadFiles(stdin), nil
	}

	return resolver.Resolve(opts.Input, opts.Recursive)
}

func waitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
}
