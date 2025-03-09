package internal

import (
    "fmt"
    "os/exec"
    "strings"
    "syscall"
)

// RunCommand executes the provided shell command using "sh -c".
// substituting "/_" with the changed file name.
// It returns the raw output along with a structured status log (pipe-separated).
func RunCommand(command, file string) string {
    // Substitute the placeholder if present.
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
            return ""
        }
    }
    statusLog := fmt.Sprintf("exit|%d|%s", exitCode, command)
    combined := fmt.Sprintf("%s\n----\n%s", string(out), statusLog)
    return combined
}
