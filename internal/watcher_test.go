package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type fakeSpinner struct {
	paused  int
	resumed int
}

func (s *fakeSpinner) Start()  {}
func (s *fakeSpinner) Stop()   {}
func (s *fakeSpinner) Pause()  { s.paused++ }
func (s *fakeSpinner) Resume() { s.resumed++ }

type fakeRunner struct {
	commands []string
	files    []string
	result   CommandResult
}

func (r *fakeRunner) Run(command string, file string) CommandResult {
	r.commands = append(r.commands, command)
	r.files = append(r.files, file)
	if r.result.Command == "" {
		r.result.Command = command
	}
	return r.result
}

type fakeReporter struct {
	results []CommandResult
}

func (r *fakeReporter) Report(result CommandResult, opts WatchOptions) {
	r.results = append(r.results, result)
}

type fakeLogger struct {
	entries []string
	results []CommandResult
}

func (l *fakeLogger) Init(opts WatchOptions, command string) error { return nil }
func (l *fakeLogger) Write(entry string, result CommandResult) error {
	l.entries = append(l.entries, entry)
	l.results = append(l.results, result)
	return nil
}

func TestWatcherTracksAndRemovesFiles(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "a.txt")
	if err := os.WriteFile(file, []byte("hello"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	watcher := NewWatcher(NewWatchOptions(false, false, tmp, 0, false), nil, nil, nil, nil)
	if err := watcher.trackFile(file, false); err != nil {
		t.Fatalf("trackFile returned error: %v", err)
	}

	files := watcher.snapshotFiles()
	if len(files) != 1 || files[0] != file {
		t.Fatalf("unexpected snapshot: %#v", files)
	}

	watcher.removeFile(file)
	if len(watcher.snapshotFiles()) != 0 {
		t.Fatalf("expected file to be removed")
	}
}

func TestWatcherMarkModified(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "a.txt")
	if err := os.WriteFile(file, []byte("hello"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	watcher := NewWatcher(NewWatchOptions(false, false, tmp, 0, false), nil, nil, nil, nil)
	if err := watcher.trackFile(file, false); err != nil {
		t.Fatalf("trackFile returned error: %v", err)
	}

	oldMod := watcher.modTimes[file]
	if watcher.markModified(file, oldMod) {
		t.Fatal("same mod time should not be treated as modified")
	}
	if !watcher.markModified(file, oldMod.Add(time.Second)) {
		t.Fatal("newer mod time should be treated as modified")
	}
}

func TestWatcherHandleChangeUsesInterfaces(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "a.txt")
	if err := os.WriteFile(file, []byte("old\n"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	opts := NewWatchOptions(false, false, tmp, 0, true)
	opts.DebounceDuration = time.Millisecond
	spinner := &fakeSpinner{}
	runner := &fakeRunner{result: CommandResult{RawOutput: "ok", ExitCode: 0, Command: "go test"}}
	reporter := &fakeReporter{}
	logger := &fakeLogger{}
	watcher := NewWatcher(opts, spinner, runner, reporter, logger)

	if err := watcher.trackFile(file, false); err != nil {
		t.Fatalf("trackFile returned error: %v", err)
	}
	if err := os.WriteFile(file, []byte("new\n"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	watcher.handleChange(file, "go test /_")

	if spinner.paused != 1 || spinner.resumed != 1 {
		t.Fatalf("expected spinner pause/resume once, got pause=%d resume=%d", spinner.paused, spinner.resumed)
	}
	if len(runner.commands) != 1 || runner.commands[0] != "go test /_" || runner.files[0] != file {
		t.Fatalf("unexpected runner calls: %#v %#v", runner.commands, runner.files)
	}
	if len(reporter.results) != 1 {
		t.Fatalf("expected reporter to be called once, got %d", len(reporter.results))
	}
	if len(logger.entries) == 0 {
		t.Fatal("expected diff entries to be logged")
	}
}

func TestReadFileLinesAndFormatDiffEntry(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "a.txt")
	if err := os.WriteFile(file, []byte("a\nb"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	lines, err := readFileLines(file)
	if err != nil {
		t.Fatalf("readFileLines returned error: %v", err)
	}
	if len(lines) != 2 || lines[0] != "a" || lines[1] != "b" {
		t.Fatalf("unexpected lines: %#v", lines)
	}

	entry := formatDiffEntry(file, DiffChange{LineNumber: 1, Type: "ADD", Text: "hello"})
	if !strings.Contains(entry, file) || !strings.Contains(entry, "ADD") || !strings.Contains(entry, "hello") {
		t.Fatalf("unexpected diff entry: %q", entry)
	}
	if got := formatDiffEntry(file, DiffChange{Type: "NOPE"}); got != "" {
		t.Fatalf("expected unknown diff type to return empty string, got %q", got)
	}
}
