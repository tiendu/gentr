package internal

import (
    "fmt"
    "os/exec"
)

// RunCommand executes the provided shell command using "sh -c".
// It prints the combined output or an error if it occurs.
func RunCommand(command string) string {
    // Use "sh -c" so that the entire command string is interpreted by the shell.
    cmd := exec.Command("sh", "-c", command)
    out, err := cmd.CombinedOutput()
    if err != nil {
        fmt.Println("Error running command:", err)
        return ""
    }
    output := string(out)
    fmt.Println("Command Output:", output)
    return output
}

