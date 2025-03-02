package internal

import (
    "encoding/csv"
    "fmt"
    "strings"
)

// FilterLogs processes the command output, expecting structured logs that are pipe-separated.
// If the output is not structured as expected, it prints the output as-is.
func FilterLogs(input string) {
    lines := strings.Split(input, "\n")
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if line == "" {
            continue
        }
        // Look for a pipe-separated structured log.
        if strings.Contains(line, "|") {
            reader := csv.NewReader(strings.NewReader(line))
            reader.Comma = '|'
            records, err := reader.ReadAll()
            if err == nil && len(records) > 0 && len(records[0]) >= 3 {
                eventType := records[0][0]
                code := records[0][1]
                cmdName := records[0][2]
                switch eventType {
                case "exit":
                    fmt.Printf("Command '%s' returned exit code %s\n", cmdName, code)
                case "signal":
                    fmt.Printf("Command '%s' terminated by signal %s\n", cmdName, code)
                default:
                    fmt.Println(line)
                }
            } else {
                // Fallback: print the line.
                fmt.Println(line)
            }
        } else {
            // For plain text output.
            fmt.Println(line)
        }
    }
}

