package main

import (
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "strings"

    "gentr/cmd"
    "gentr/internal"
)

func main() {
    // Parse CLI flags/options.
    opts := cmd.ParseOptions()
    fmt.Println("Starting with options:", opts)

    // First, try to read file paths from STDIN.
    var files []string
    fi, _ := os.Stdin.Stat()
    if (fi.Mode() & os.ModeCharDevice) == 0 {
        // Data is being piped in from STDIN.
        files = internal.ReadFilesFromStdin()
    } else {
        // No piped input; resolve file list based on --input and --recursive.
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

    // Get the command to execute.
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

