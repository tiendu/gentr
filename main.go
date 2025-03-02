package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"
    "gentr/cmd"
    "gentr/internal"
)

func main() {
    // Parse CLI flags/options.
    options := cmd.ParseOptions()
    fmt.Println("Starting gentr with options:", options)

    // Read file list from STDIN.
    scanner := bufio.NewScanner(os.Stdin)
    var files []string
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if line != "" {
            files = append(files, line)
        }
    }
    if err := scanner.Err(); err != nil {
        fmt.Fprintln(os.Stderr, "Error reading files:", err)
        os.Exit(1)
    }
    if len(files) == 0 {
        fmt.Fprintln(os.Stderr, "No files provided via STDIN")
        os.Exit(1)
    }

    // Retrieve the command to execute from the remaining arguments.
    // This assumes cmd.ParseOptions() has consumed the flags.
    commandArgs := cmd.GetCommandArgs()
    if len(commandArgs) == 0 {
        fmt.Fprintln(os.Stderr, "No command provided to execute")
        os.Exit(1)
    }
    command := strings.Join(commandArgs, " ")

    // Start watching files; when a change is detected, run the command.
    // WatchFiles is updated to accept a command parameter.
    go internal.WatchFiles(files, command)

    // Block forever (or until a signal is received).
    select {}
}

