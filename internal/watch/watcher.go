package watch

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/tiendu/gentr/internal/config"
	"github.com/tiendu/gentr/internal/diff"
	"github.com/tiendu/gentr/internal/runner"
	"github.com/tiendu/gentr/internal/terminal"
)

type Spinner interface {
	Start()
	Stop()
	Pause()
	Resume()
}

type CommandRunner interface {
	Run(command, file string) runner.Result
}

type OutputReporter interface {
	Report(result runner.Result, opts config.Options)
}

type ChangeLogger interface {
	Write(entry string, result runner.Result) error
}

type Resolver interface {
	Resolve(input string, recursive bool) ([]string, error)
}

type Watcher struct {
	opts         config.Options
	spinner      Spinner
	runner       CommandRunner
	reporter     OutputReporter
	logger       ChangeLogger
	resolver     Resolver
	output       io.Writer
	modTimes     map[string]time.Time
	fileContents map[string][]string
}

func New(
	opts config.Options,
	spinner Spinner,
	commandRunner CommandRunner,
	reporter OutputReporter,
	logger ChangeLogger,
	resolver Resolver,
	output io.Writer,
) *Watcher {
	if output == nil {
		output = io.Discard
	}
	if commandRunner == nil {
		commandRunner = runner.Shell{}
	}
	if reporter == nil {
		reporter = discardReporter{}
	}
	if logger == nil {
		logger = discardLogger{}
	}

	return &Watcher{
		opts:         opts,
		spinner:      spinner,
		runner:       commandRunner,
		reporter:     reporter,
		logger:       logger,
		resolver:     resolver,
		output:       output,
		modTimes:     make(map[string]time.Time),
		fileContents: make(map[string][]string),
	}
}

func (w *Watcher) Run(ctx context.Context, files []string, command string) {
	for _, file := range files {
		if err := w.trackFile(file, false); err != nil {
			fmt.Fprintf(w.output, "\n[x] Error tracking file %s: %v\n", file, err)
		}
	}

	pollTicker := time.NewTicker(w.opts.PollInterval)
	defer pollTicker.Stop()

	var rescanTicker *time.Ticker
	var rescanChannel <-chan time.Time
	if w.opts.Recursive && w.resolver != nil {
		rescanTicker = time.NewTicker(w.opts.RescanInterval)
		rescanChannel = rescanTicker.C
		defer rescanTicker.Stop()
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-pollTicker.C:
			w.poll(ctx, command)
		case <-rescanChannel:
			w.rescan()
		}
	}
}

func (w *Watcher) poll(ctx context.Context, command string) {
	for _, file := range w.snapshotFiles() {
		info, err := os.Stat(file)
		if err != nil {
			if os.IsNotExist(err) {
				w.removeFile(file)
				w.handleDeletion(file)
				continue
			}
			fmt.Fprintf(w.output, "\n[x] Error stating file %s: %v\n", file, err)
			continue
		}
		if !info.IsDir() && w.markModified(file, info.ModTime()) {
			w.handleChange(ctx, file, command)
		}
	}
}

func (w *Watcher) rescan() {
	files, err := w.resolver.Resolve(w.opts.Input, true)
	if err != nil {
		fmt.Fprintf(w.output, "\n[x] Error rescanning %s: %v\n", w.opts.Input, err)
		return
	}
	for _, file := range files {
		if err := w.trackFile(file, true); err != nil {
			fmt.Fprintf(w.output, "\n[x] Error tracking new file %s: %v\n", file, err)
		}
	}
}

func (w *Watcher) handleDeletion(path string) {
	fmt.Fprintf(w.output, "\n[!] File deleted: %s\n", path)
	if !w.opts.Log {
		return
	}

	result := runner.Result{ExitCode: -1, Command: "DELETED"}
	if err := w.logger.Write(fmt.Sprintf("%s: DELETED", path), result); err != nil {
		fmt.Fprintf(w.output, "\n[x] Error writing deletion log: %v\n", err)
	}
}

func (w *Watcher) handleChange(ctx context.Context, path, command string) {
	if w.spinner != nil {
		w.spinner.Pause()
		defer w.spinner.Resume()
	}

	timer := time.NewTimer(w.opts.DebounceDuration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return
	case <-timer.C:
	}

	fmt.Fprintf(w.output, "\nChange detected in file: %s. Executing command...\n", path)
	newContent, err := readFileLines(path)
	if err != nil {
		fmt.Fprintf(w.output, "\n[x] Error reading file %s: %v\n", path, err)
		return
	}

	oldContent := w.replaceFileContent(path, newContent)
	result := w.runner.Run(command, path)
	w.reporter.Report(result, w.opts)
	w.printAndLogDiff(path, oldContent, newContent, result)
}

func (w *Watcher) printAndLogDiff(path string, oldContent, newContent []string, result runner.Result) {
	changes := diff.CombineModifications(diff.Lines(oldContent, newContent))
	for _, change := range changes {
		entry := formatDiffEntry(path, change)
		if entry == "" {
			continue
		}
		fmt.Fprintln(w.output, entry)
		if w.opts.Log {
			if err := w.logger.Write(strings.TrimSpace(entry), result); err != nil {
				fmt.Fprintf(w.output, "\n[x] Error writing log: %v\n", err)
			}
		}
	}
}

func (w *Watcher) trackFile(path string, announce bool) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return nil
	}
	if _, exists := w.modTimes[path]; exists {
		return nil
	}

	w.modTimes[path] = info.ModTime()
	if content, err := readFileLines(path); err == nil {
		w.fileContents[path] = content
	}
	if announce {
		fmt.Fprintf(w.output, "\n[v] New file detected and added: %s\n", path)
	}
	return nil
}

func (w *Watcher) snapshotFiles() []string {
	files := make([]string, 0, len(w.modTimes))
	for file := range w.modTimes {
		files = append(files, file)
	}
	return files
}

func (w *Watcher) markModified(path string, modTime time.Time) bool {
	lastMod, exists := w.modTimes[path]
	if !exists || !modTime.After(lastMod) {
		return false
	}
	w.modTimes[path] = modTime
	return true
}

func (w *Watcher) removeFile(path string) {
	delete(w.modTimes, path)
	delete(w.fileContents, path)
}

func (w *Watcher) replaceFileContent(path string, newContent []string) []string {
	oldContent := w.fileContents[path]
	w.fileContents[path] = newContent
	return oldContent
}

func readFileLines(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(data), "\n"), nil
}

func formatDiffEntry(path string, change diff.Change) string {
	text := terminal.TruncateLine(change.Text, 125)
	switch change.Kind {
	case diff.Modified:
		return fmt.Sprintf(
			"%s:%d %s: %s",
			terminal.Bold(terminal.Color(path, "cyan")),
			change.LineNumber,
			terminal.Bold(terminal.Highlight("MOD", "gray", "yellow")),
			terminal.Bold(text),
		)
	case diff.Removed:
		return fmt.Sprintf(
			"%s:%d %s: %s",
			terminal.Bold(terminal.Color(path, "cyan")),
			change.LineNumber,
			terminal.Bold(terminal.Highlight("REM", "white", "red")),
			terminal.Bold(terminal.Color(text, "red")),
		)
	case diff.Added:
		return fmt.Sprintf(
			"%s:%d %s: %s",
			terminal.Bold(terminal.Color(path, "cyan")),
			change.LineNumber,
			terminal.Bold(terminal.Highlight("ADD", "white", "green")),
			terminal.Bold(terminal.Color(text, "green")),
		)
	default:
		return ""
	}
}

type discardReporter struct{}

func (discardReporter) Report(runner.Result, config.Options) {}

type discardLogger struct{}

func (discardLogger) Write(string, runner.Result) error { return nil }
