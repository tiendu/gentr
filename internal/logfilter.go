package internal

import (
    "encoding/csv"
    "fmt"
    "strings"

    "gentr/cmd"
    "gentr/internal/beautify"
)

// truncateLine truncates a string to maxLen characters, appending "..." if truncated.
func truncateLine(line string, maxLen int) string {
    if len(line) > maxLen {
        return line[:maxLen] + "..."
    }
    return line
}

// FilterLogs processes the combined command output and status log.
// It splits the output (separated by a delimiter) and prints the parts with enhanced formatting.
// Highlight is used to emphasize the command name.
func FilterLogs(input string, opts cmd.Options) {
    // Split the input by the delimiter we defined in RunCommand.
    parts := strings.Split(input, "\n----\n")
    if len(parts) != 2 {
        // Fallback: just print the input if it doesn't match the expected format.
        fmt.Println(input)
        return
    }
    rawOutput := parts[0]
    statusLine := strings.TrimSpace(parts[1])

    // Process raw command output:
    rawLines := strings.Split(rawOutput, "\n")
    if opts.Length > 0 {
        if len(rawLines) > opts.Length {
            rawLines = append([]string{"..."}, rawLines[len(rawLines)-opts.Length+1:]...)
        }
    }
    // Truncate each line to 60 characters.
    for i, line := range rawLines {
        rawLines[i] = truncateLine(line, 60)
    }

    // Print raw command output with a header.
    fmt.Println(beautify.Bold(beautify.Color("Command Output:", "blue")))
    fmt.Println(strings.Join(rawLines, "\n"))

    // Process the status log.
    if strings.Contains(statusLine, "|") {
        reader := csv.NewReader(strings.NewReader(statusLine))
        reader.Comma = '|'
        records, err := reader.ReadAll()
        if err == nil && len(records) > 0 && len(records[0]) >= 3 {
            eventType := records[0][0]
            code := records[0][1]
            cmdName := records[0][2]
            var formattedStatus string
            switch eventType {
            case "exit":
                if code == "0" {
                    // Use Highlight to emphasize a successful command.
                    formattedStatus = fmt.Sprintf("%s %s",
                        beautify.Bold(beautify.Highlight(fmt.Sprintf("Command '%s'", cmdName), "white", "green")),
                        beautify.Color(fmt.Sprintf("returned exit code %s", code), "green"))
                } else {
                    // Use Highlight for a failed command.
                    formattedStatus = fmt.Sprintf("%s %s",
                        beautify.Bold(beautify.Highlight(fmt.Sprintf("Command '%s'", cmdName), "white", "red")),
                        beautify.Color(fmt.Sprintf("returned exit code %s", code), "red"))
                }
            case "signal":
                // Use Highlight for a command terminated by signal.
                formattedStatus = fmt.Sprintf("%s %s",
                    beautify.Bold(beautify.Highlight(fmt.Sprintf("Command '%s'", cmdName), "white", "yellow")),
                    beautify.Color(fmt.Sprintf("terminated by signal %s", code), "yellow"))
            default:
                formattedStatus = statusLine
            }
            fmt.Println(beautify.Bold(beautify.Color("Status Log:", "blue")))
            fmt.Println(formattedStatus)
        } else {
            fmt.Println(statusLine)
        }
    } else {
        fmt.Println(statusLine)
    }
}

