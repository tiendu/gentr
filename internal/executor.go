package internal

import (
    "fmt"
    "os/exec"
    "strings"
    "syscall"
)

// CommandResult holds the output of a command execution.
type CommandResult struct {
    RawOutput string
    ExitCode  int
    Command   string
}

// RunCommand executes the provided shell command using "sh -c", substituting "/_" with the changed file name.
// It returns a CommandResult struct containing the raw output, the exit code, and the executed command.
func RunCommand(command, file string) CommandResult {
    // Substitute placeholder if present.
    if strings.Contains(command, "/_") {
        command = strings.ReplaceAll(command, "/_", file)
    }
    cmd := exec.Command("sh", "-c", command)
    out, err := cmd.CombinedOutput()
    exitCode := 0
    if err != nil {
        if exitError, ok := err.(*exec.ExitError); ok {
            if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
                exitCode = status.ExitStatus()
            }
        } else {
            fmt.Println("Error running command:", err)
        }
    }
    return CommandResult{
        RawOutput: string(out),
        ExitCode:  exitCode,
        Command:   command,
    }
}

