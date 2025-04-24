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
	if len(os.Args) < 2 {
		cmd.HelpCommand()
		os.Exit(0)
	}

	args := os.Args[1:]
	subcommand := args[0]
	subArgs := args[1:]

	// Handle routing first (install, uninstall, etc.)
	router := cmd.NewCommandRouter()
	if isBuiltinCommand(subcommand) {
		router.Run(subcommand, subArgs)
		return
	}

	// Continue with watch mode
	opts := cmd.ParseOptions()
	fmt.Println("Starting with options:", opts)

	var files []string
	fi, _ := os.Stdin.Stat()
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		files = internal.ReadFilesFromStdin()
	} else {
		var err error
		files, err = internal.ResolveFiles(opts)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "No files provided via STDIN or --input flag")
		os.Exit(1)
	}

	commandArgs := cmd.GetCommandArgs()
	if len(commandArgs) == 0 {
		fmt.Fprintln(os.Stderr, "No command provided to execute")
		os.Exit(1)
	}
	command := strings.Join(commandArgs, " ")

	if opts.Log {
		if err := internal.InitSessionLog(opts, command); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing log file: %v\n", err)
		}
	}

	sp := cmd.NewSnakeSpinner(30, 5, 81)
	sp.Start()
	go internal.WatchFiles(files, command, opts, sp)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	sp.Stop()
	fmt.Println("\nShutting down gentr...")
}

func isBuiltinCommand(name string) bool {
	switch name {
	case "install", "uninstall", "reinstall", "version", "bump", "help", "--help", "-h":
		return true
	default:
		return false
	}
}
