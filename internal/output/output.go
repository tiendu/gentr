package output

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/tiendu/gentr/internal/config"
	"github.com/tiendu/gentr/internal/runner"
	"github.com/tiendu/gentr/internal/terminal"
)

type ConsoleReporter struct {
	Writer io.Writer
}

func (r ConsoleReporter) Report(result runner.Result, opts config.Options) {
	writer := r.Writer
	if writer == nil {
		writer = os.Stdout
	}

	lines := strings.Split(result.RawOutput, "\n")
	if opts.Length > 0 && len(lines) > opts.Length {
		lines = append([]string{"..."}, lines[len(lines)-opts.Length+1:]...)
	}
	for index, line := range lines {
		lines[index] = terminal.TruncateLine(line, 60)
	}

	fmt.Fprintln(writer, terminal.Bold(terminal.Color("Command Output:", "blue")))
	fmt.Fprintln(writer, strings.Join(lines, "\n"))
	fmt.Fprintln(writer, terminal.Bold(terminal.Color("Status Log:", "blue")))
	fmt.Fprintln(writer, formatStatus(result))
}

type SessionLogger struct {
	path   string
	output io.Writer
	now    func() time.Time
}

func NewSessionLogger(output io.Writer) *SessionLogger {
	if output == nil {
		output = io.Discard
	}
	return &SessionLogger{output: output, now: time.Now}
}

func (l *SessionLogger) Init(opts config.Options, command string) error {
	now := l.now
	if now == nil {
		now = time.Now
	}
	l.path = now().Format("2006-01-02T15-04-05") + ".log"
	fmt.Fprintf(l.output, "Created log file: %s\n", l.path)

	file, err := os.Create(l.path)
	if err != nil {
		return fmt.Errorf("create session log file: %w", err)
	}
	defer file.Close()

	header := fmt.Sprintf(
		"# Options: %s\n# Command: %s\n# Timestamp: %s\n",
		terminal.StripANSI(opts.String()),
		command,
		now().Format(time.RFC3339),
	)
	separator := strings.Repeat("-", 80) + "\n"
	_, err = file.WriteString(header + separator + "Output\tExitStatus\n" + separator)
	return err
}

func (l *SessionLogger) Write(entry string, result runner.Result) error {
	if l.path == "" {
		return fmt.Errorf("session log file not initialized")
	}

	file, err := os.OpenFile(l.path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	writer.Comma = '\t'
	if err := writer.Write([]string{
		terminal.StripANSI(entry),
		fmt.Sprintf("ExitStatus: %d", result.ExitCode),
	}); err != nil {
		return fmt.Errorf("write log record: %w", err)
	}
	writer.Flush()
	return writer.Error()
}

func formatStatus(result runner.Result) string {
	switch {
	case result.ExitCode == 0:
		return fmt.Sprintf(
			"%s%s",
			terminal.Bold(terminal.Highlight(fmt.Sprintf("exit|%d", result.ExitCode), "white", "green")),
			terminal.Color(fmt.Sprintf("|%s", result.Command), "green"),
		)
	case result.ExitCode >= 128:
		return fmt.Sprintf(
			"%s%s",
			terminal.Bold(terminal.Highlight(fmt.Sprintf("signal|%d", result.ExitCode), "white", "yellow")),
			terminal.Color(fmt.Sprintf("|%s", result.Command), "yellow"),
		)
	default:
		return fmt.Sprintf(
			"%s%s",
			terminal.Bold(terminal.Highlight(fmt.Sprintf("exit|%d", result.ExitCode), "white", "red")),
			terminal.Color(fmt.Sprintf("|%s", result.Command), "red"),
		)
	}
}
