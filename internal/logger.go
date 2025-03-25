package internal

import (
    "encoding/csv"
    "fmt"
    "os"
    "strings"
    "time"

    "gentr/cmd"
    "gentr/internal/utils"
)

// sessionLogFile holds the filename for the current session.
var sessionLogFile string

// FilterLogs processes a CommandResult and prints the raw command output and a formatted status log.
func FilterLogs(cr CommandResult, opts cmd.Options) {
    // Split the raw output into lines.
    rawLines := strings.Split(cr.RawOutput, "\n")
    if opts.Length > 0 && len(rawLines) > opts.Length {
        rawLines = append([]string{"..."}, rawLines[len(rawLines)-opts.Length+1:]...)
    }
    // Truncate each line to 60 characters.
    for i, line := range rawLines {
        rawLines[i] = utils.TruncateLine(line, 60)
    }

    // Print the command output.
    fmt.Println(utils.Bold(utils.Color("Command Output:", "blue")))
    fmt.Println(strings.Join(rawLines, "\n"))

    // Build and highlight the status log string based on the exit code.
    var statusLog string
    if cr.ExitCode == 0 {
        statusLog = fmt.Sprintf("%s%s",
            utils.Bold(utils.Highlight(fmt.Sprintf("exit|%d", cr.ExitCode), "white", "green")),
            utils.Color(fmt.Sprintf("|%s", cr.Command), "green"))
    } else if cr.ExitCode >= 128 {
        statusLog = fmt.Sprintf("%s%s",
            utils.Bold(utils.Highlight(fmt.Sprintf("signal|%d", cr.ExitCode), "white", "yellow")),
            utils.Color(fmt.Sprintf("|%s", cr.Command), "yellow"))
    } else {
        statusLog = fmt.Sprintf("%s%s",
            utils.Bold(utils.Highlight(fmt.Sprintf("exit|%d", cr.ExitCode), "white", "red")),
            utils.Color(fmt.Sprintf("|%s", cr.Command), "red"))
    }

    fmt.Println(utils.Bold(utils.Color("Status Log:", "blue")))
    fmt.Println(statusLog)
}

// InitSessionLog creates (or truncates) a log file for the current session,
// using a timestamp-based filename, and writes the header (metadata) into it.
func InitSessionLog(opts cmd.Options, command string) error {
    sessionLogFile = time.Now().Format("2006-01-02T15-04-05") + ".log"
    fmt.Printf("Created log file: %s\n", sessionLogFile)

    // Open the file in create/truncate mode.
    f, err := os.Create(sessionLogFile)
    if err != nil {
        return fmt.Errorf("failed to create session log file: %v", err)
    }
    defer f.Close()

    // Write metadata header (each line starting with "#").
    timestamp := time.Now().Format(time.RFC3339)
    header := fmt.Sprintf("# Options: %s\n# Command: %s\n# Timestamp: %s\n", utils.StripANSI(opts.String()), command, timestamp)
    separator := strings.Repeat("-", 80) + "\n"
    if _, err := f.WriteString(header + separator); err != nil {
        return err
    }
    // Write the TSV header.
    tsvHeader := "Output\tExitStatus\n"
    if _, err := f.WriteString(tsvHeader + separator); err != nil {
        return err
    }
    return nil
}

// WriteLogEntry appends a log entry to the session log file in TSV format.
// diffEntry is a plain text description of the file change, and cr is the CommandResult.
func WriteLogEntry(diffEntry string, cr CommandResult) error {
    if sessionLogFile == "" {
        return fmt.Errorf("session log file not initialized")
    }
    f, err := os.OpenFile(sessionLogFile, os.O_APPEND|os.O_WRONLY, 0644)
    if err != nil {
        return fmt.Errorf("failed to open log file: %v", err)
    }
    defer f.Close()

    // Prepare the TSV record: strip ANSI codes from the diff entry.
    record := []string{
        utils.StripANSI(diffEntry),
        fmt.Sprintf("ExitStatus: %d", cr.ExitCode),
    }

    writer := csv.NewWriter(f)
    writer.Comma = '\t'
    if err := writer.Write(record); err != nil {
        return fmt.Errorf("failed to write record: %v", err)
    }
    writer.Flush()
    return writer.Error()
}

