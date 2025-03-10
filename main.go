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
    if len(os.Args) < 2 {
        cmd.HelpCommand()
        os.Exit(0)
    }

    // Get the first argument.
    args := os.Args[1:]

    // Handle different subcommands
    switch args[0] {
    case "install":
        cmd.InstallCommand()
        return
    case "reinstall":
        cmd.ReinstallCommand()
        return
    case "uninstall":
        cmd.UninstallCommand()
        return
    case "version":
        cmd.VersionCommand()
        return
    case "bump":
        if newVersion, err := cmd.BumpVersion(); err != nil {
            fmt.Printf("Version bump failed: %v\n", err)
        } else {
            fmt.Printf("New version: %s\n", newVersion)
        }
        return
    case "help", "--help", "-h":
        cmd.HelpCommand()
        os.Exit(0)
    default:
        // If an unknown command is provided, show help.
        cmd.HelpCommand()
        os.Exit(1)
    }

    // Parse the flags/options
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
    commandArgs := cmd.GetCommandArgs()
    if len(commandArgs) == 0 {
        fmt.Fprintln(os.Stderr, "No command provided to execute")
        os.Exit(1)
    }
    command := strings.Join(commandArgs, " ")

    // Start watching files.
    go internal.WatchFiles(files, command, opts)

    // Setup graceful shutdown on SIGINT/SIGTERM.
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    <-sigChan
    fmt.Println("\nShutting down gentr...")
}

