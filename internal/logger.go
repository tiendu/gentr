package internal

import (
    "fmt"
    "os"
    "strings"
    "regexp"
    "time"
    "encoding/csv"

    "gentr/cmd"
    "gentr/internal/utils"
)

// ansiRegexp is used to match ANSI escape sequences.
var ansiRegexp = regexp.MustCompile(`\033\[[0-9;]*m`)

// StripANSI removes ANSI escape sequences from a string.
func StripANSI(s string) string {
    return ansiRegexp.ReplaceAllString(s, "")
}

// sessionLogFile holds the filename for the current gentr session.
var sessionLogFile string

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
        rawLines[i] = utils.TruncateLine(line, 60)
    }

    // Print raw command output with a header.
    fmt.Println(utils.Bold(utils.Color("Command Output:", "blue")))
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
                        utils.Bold(utils.Highlight(fmt.Sprintf("Command '%s'", cmdName), "white", "green")),
                        utils.Color(fmt.Sprintf("returned exit code %s", code), "green"))
                } else {
                    // Use Highlight for a failed command.
                    formattedStatus = fmt.Sprintf("%s %s",
                        utils.Bold(utils.Highlight(fmt.Sprintf("Command '%s'", cmdName), "white", "red")),
                        utils.Color(fmt.Sprintf("returned exit code %s", code), "red"))
                }
            case "signal":
                // Use Highlight for a command terminated by signal.
                formattedStatus = fmt.Sprintf("%s %s",
                    utils.Bold(utils.Highlight(fmt.Sprintf("Command '%s'", cmdName), "white", "yellow")),
                    utils.Color(fmt.Sprintf("terminated by signal %s", code), "yellow"))
            default:
                formattedStatus = statusLine
            }
            fmt.Println(utils.Bold(utils.Color("Status Log:", "blue")))
            fmt.Println(formattedStatus)
        } else {
            fmt.Println(statusLine)
        }
    } else {
        fmt.Println(statusLine)
    }
}

// InitSessionLog creates (or truncates) a log file for the current session,
// using a timestamp-based filename, and writes the header (metadata) into it.
func InitSessionLog(opts cmd.Options, command string) error {
    sessionLogFile = time.Now().Format("2006-01-02T15-04-05") + ".log"

    // Inform the user about the log file.
    fmt.Printf("Created log file: %s\n", sessionLogFile) 

    // Open the file in create/truncate mode.
    f, err := os.Create(sessionLogFile)
    if err != nil {
        return fmt.Errorf("failed to create session log file: %v", err)
    }
    defer f.Close()

    // Write a header with metadata.
    timestamp := time.Now().Format(time.RFC3339)
    header := fmt.Sprintf("# Options: %s\n# Command: %s\n# Timestamp: %s\n", StripANSI(opts.String()), command, timestamp)
    separator := strings.Repeat("-", 80) + "\n"
    // Write header and a TSV header.
    if _, err := f.WriteString(header + separator); err != nil {
        return err
    }
    tsvHeader := "Output\tExitStatus\n"
    if _, err := f.WriteString(tsvHeader + separator); err != nil {
        return err
    }
    return nil
}

// WriteLogEntry appends a log entry to a log file whose name is based on the current datetime.
// The log entry includes metadata (options used and command run) and a table with two columns:
// one for the command output (raw output) and one for the exit status.
func WriteLogEntry(diffEntry string, exitStatus int) error {
    if sessionLogFile == "" {
        return fmt.Errorf("session log file not initialized")
    }
    f, err := os.OpenFile(sessionLogFile, os.O_APPEND|os.O_WRONLY, 0644)
    if err != nil {
        return fmt.Errorf("failed to open log file: %v", err)
    }
    defer f.Close()

    // Prepare the record (strip ANSI colors).
    record := []string{
        StripANSI(diffEntry),
        fmt.Sprintf("ExitStatus: %d", exitStatus),
    }

    // Use csv.Writer with tab as delimiter.
    writer := csv.NewWriter(f)
    writer.Comma = '\t'
    if err := writer.Write(record); err != nil {
        return fmt.Errorf("failed to write record: %v", err)
    }
    writer.Flush()
    return writer.Error()
}

