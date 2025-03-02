package main

import (
    "bufio"
    "fmt"
    "os"
    "os/signal"
    "strings"
    "syscall"

    "gentr/cmd"
    "gentr/internal"
)

func main() {
    // Parse CLI flags/options.
    options := cmd.ParseOptions()
    fmt.Println("Starting with options:", options)

    // Check for administrative subcommands.
    args := cmd.GetCommandArgs()
    if len(args) > 0 {
        switch strings.ToLower(args[0]) {
        case "version":
            cmd.VersionCommand()
            return
        case "bump":
            if newVer, err := cmd.BumpVersion(); err != nil {
                fmt.Fprintln(os.Stderr, "Error bumping version:", err)
            } else {
                fmt.Println("New version:", newVer)
            }
            return
        }
    }

    // Read file list from STDIN.
    scanner := bufio.NewScanner(os.Stdin)
    var paths []string
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if line != "" {
            paths = append(paths, line)
        }
    }
    if err := scanner.Err(); err != nil {
        fmt.Fprintln(os.Stderr, "Error reading file list:", err)
        os.Exit(1)
    }
    if len(paths) == 0 {
        fmt.Fprintln(os.Stderr, "No files provided via STDIN")
        os.Exit(1)
    }

    // The remaining command-line arguments form the command to execute.
    if len(args) == 0 {
        fmt.Fprintln(os.Stderr, "No command provided to execute")
        os.Exit(1)
    }
    command := strings.Join(args, " ")

    // Start watching files concurrently with debounce and recursive scanning.
    go internal.WatchFiles(paths, command, options)

    // Setup graceful shutdown on SIGINT/SIGTERM.
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    <-sigChan
    fmt.Println("\nShutting down gentr...")
}

