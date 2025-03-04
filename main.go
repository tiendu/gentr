package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"gentr/cmd"
	"gentr/internal"
)

func main() {
	// Parse CLI flags/options.
	opts := cmd.ParseOptions()
	fmt.Println("Starting with options:", opts)

	var files []string

	// Check if STDIN is being piped in.
	fi, _ := os.Stdin.Stat()
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		// Data is being piped in from STDIN.
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				files = append(files, line)
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "Error reading file list from STDIN:", err)
			os.Exit(1)
		}
	} else {
		// No piped input; use the --input flag.
		if strings.ContainsAny(opts.Input, "*?[]") {
			// Input is a glob pattern.
			matches, err := filepath.Glob(opts.Input)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error processing glob pattern: %v\n", err)
				os.Exit(1)
			}
			files = matches
		} else {
			// Input is a file or directory.
			info, err := os.Stat(opts.Input)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error accessing input %s: %v\n", opts.Input, err)
				os.Exit(1)
			}
			if info.IsDir() {
				// For a directory, pass it to the watcher; it will handle recursion.
				files = []string{opts.Input}
			} else {
				// A single file.
				files = []string{opts.Input}
			}
		}
	}

	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "No files provided via STDIN or --input flag")
		os.Exit(1)
	}

	// The remaining command-line arguments form the command to execute.
	args := cmd.GetCommandArgs()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "No command provided to execute")
		os.Exit(1)
	}
	command := strings.Join(args, " ")

	// Start watching files.
	go internal.WatchFiles(files, command, opts)

	// Setup graceful shutdown on SIGINT/SIGTERM.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	fmt.Println("\nShutting down gentr...")
}
