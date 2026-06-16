package internal

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"time"

	"gentr/internal/utils"
)

type OutputReporter interface {
	Report(result CommandResult, opts WatchOptions)
}

type ChangeLogger interface {
	Init(opts WatchOptions, command string) error
	Write(entry string, result CommandResult) error
}

type ConsoleReporter struct{}

type SessionLogger struct {
	path string
}

func (ConsoleReporter) Report(cr CommandResult, opts WatchOptions) {
	rawLines := strings.Split(cr.RawOutput, "\n")
	if opts.Length > 0 && len(rawLines) > opts.Length {
		rawLines = append([]string{"..."}, rawLines[len(rawLines)-opts.Length+1:]...)
	}

	for i, line := range rawLines {
		rawLines[i] = utils.TruncateLine(line, 60)
	}

	fmt.Println(utils.Bold(utils.Color("Command Output:", "blue")))
	fmt.Println(strings.Join(rawLines, "\n"))
	fmt.Println(utils.Bold(utils.Color("Status Log:", "blue")))
	fmt.Println(formatStatusLog(cr))
}

func (l *SessionLogger) Init(opts WatchOptions, command string) error {
	l.path = time.Now().Format("2006-01-02T15-04-05") + ".log"
	fmt.Printf("Created log file: %s\n", l.path)

	f, err := os.Create(l.path)
	if err != nil {
		return fmt.Errorf("failed to create session log file: %w", err)
	}
	defer f.Close()

	timestamp := time.Now().Format(time.RFC3339)
	header := fmt.Sprintf("# Options: %s\n# Command: %s\n# Timestamp: %s\n", utils.StripANSI(opts.String()), command, timestamp)
	separator := strings.Repeat("-", 80) + "\n"

	if _, err := f.WriteString(header + separator); err != nil {
		return err
	}
	if _, err := f.WriteString("Output\tExitStatus\n" + separator); err != nil {
		return err
	}

	return nil
}

func (l *SessionLogger) Write(entry string, cr CommandResult) error {
	if l.path == "" {
		return fmt.Errorf("session log file not initialized")
	}

	f, err := os.OpenFile(l.path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	writer.Comma = '\t'

	record := []string{
		utils.StripANSI(entry),
		fmt.Sprintf("ExitStatus: %d", cr.ExitCode),
	}

	if err := writer.Write(record); err != nil {
		return fmt.Errorf("failed to write record: %w", err)
	}
	writer.Flush()
	return writer.Error()
}

var defaultSessionLogger = &SessionLogger{}

func FilterLogs(cr CommandResult, opts WatchOptions) {
	ConsoleReporter{}.Report(cr, opts)
}

func InitSessionLog(opts WatchOptions, command string) error {
	return defaultSessionLogger.Init(opts, command)
}

func WriteLogEntry(diffEntry string, cr CommandResult) error {
	return defaultSessionLogger.Write(diffEntry, cr)
}

func formatStatusLog(cr CommandResult) string {
	if cr.ExitCode == 0 {
		return fmt.Sprintf("%s%s",
			utils.Bold(utils.Highlight(fmt.Sprintf("exit|%d", cr.ExitCode), "white", "green")),
			utils.Color(fmt.Sprintf("|%s", cr.Command), "green"),
		)
	}

	if cr.ExitCode >= 128 {
		return fmt.Sprintf("%s%s",
			utils.Bold(utils.Highlight(fmt.Sprintf("signal|%d", cr.ExitCode), "white", "yellow")),
			utils.Color(fmt.Sprintf("|%s", cr.Command), "yellow"),
		)
	}

	return fmt.Sprintf("%s%s",
		utils.Bold(utils.Highlight(fmt.Sprintf("exit|%d", cr.ExitCode), "white", "red")),
		utils.Color(fmt.Sprintf("|%s", cr.Command), "red"),
	)
}
