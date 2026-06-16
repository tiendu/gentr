package watch

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tiendu/gentr/internal/config"
	"github.com/tiendu/gentr/internal/diff"
	"github.com/tiendu/gentr/internal/runner"
	"github.com/tiendu/gentr/internal/terminal"
)

type fakeSpinner struct{ paused, resumed int }

func (*fakeSpinner) Start()    {}
func (*fakeSpinner) Stop()     {}
func (s *fakeSpinner) Pause()  { s.paused++ }
func (s *fakeSpinner) Resume() { s.resumed++ }

type fakeRunner struct {
	commands []string
	files    []string
	result   runner.Result
}

func (r *fakeRunner) Run(command, file string) runner.Result {
	r.commands = append(r.commands, command)
	r.files = append(r.files, file)
	return r.result
}

type fakeReporter struct{ results []runner.Result }

func (r *fakeReporter) Report(result runner.Result, _ config.Options) {
	r.results = append(r.results, result)
}

type fakeLogger struct{ entries []string }

func (l *fakeLogger) Write(entry string, _ runner.Result) error {
	l.entries = append(l.entries, entry)
	return nil
}

type fakeResolver struct{ files []string }

func (r fakeResolver) Resolve(string, bool) ([]string, error) { return r.files, nil }

func TestWatcherTracksMarksAndRemovesFiles(t *testing.T) {
	path := writeTestFile(t, "hello")
	watcher := New(config.New(false, false, ".", 0, false), nil, nil, nil, nil, nil, nil)

	if err := watcher.trackFile(path, false); err != nil {
		t.Fatal(err)
	}
	if len(watcher.snapshotFiles()) != 1 {
		t.Fatalf("expected one tracked file")
	}
	oldMod := watcher.modTimes[path]
	if watcher.markModified(path, oldMod) {
		t.Fatal("same modtime should not be modified")
	}
	if !watcher.markModified(path, oldMod.Add(time.Second)) {
		t.Fatal("newer modtime should be modified")
	}
	watcher.removeFile(path)
	if len(watcher.snapshotFiles()) != 0 {
		t.Fatal("expected file to be removed")
	}
}

func TestHandleChangeUsesInterfaces(t *testing.T) {
	path := writeTestFile(t, "old\n")
	opts := config.New(false, false, ".", 0, true)
	opts.DebounceDuration = time.Millisecond
	spinner := &fakeSpinner{}
	commandRunner := &fakeRunner{result: runner.Result{RawOutput: "ok", ExitCode: 0, Command: "go test"}}
	reporter := &fakeReporter{}
	logger := &fakeLogger{}
	var output bytes.Buffer
	watcher := New(opts, spinner, commandRunner, reporter, logger, fakeResolver{}, &output)

	if err := watcher.trackFile(path, false); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("new\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	watcher.handleChange(context.Background(), path, "go test /_")

	if spinner.paused != 1 || spinner.resumed != 1 {
		t.Fatalf("unexpected spinner calls: %+v", spinner)
	}
	if len(commandRunner.commands) != 1 || commandRunner.commands[0] != "go test /_" || commandRunner.files[0] != path {
		t.Fatalf("unexpected runner calls: %+v", commandRunner)
	}
	if len(reporter.results) != 1 || len(logger.entries) == 0 {
		t.Fatalf("reporter=%+v logger=%+v", reporter, logger)
	}
}

func TestRunStopsWithContext(t *testing.T) {
	path := writeTestFile(t, "hello")
	opts := config.New(false, false, ".", 0, false)
	opts.PollInterval = time.Millisecond
	watcher := New(opts, nil, nil, nil, nil, nil, nil)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		watcher.Run(ctx, []string{path}, "true")
		close(done)
	}()
	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("watcher did not stop after cancellation")
	}
}

func TestReadFileLinesAndFormatDiffEntry(t *testing.T) {
	path := writeTestFile(t, "a\nb")
	lines, err := readFileLines(path)
	if err != nil || len(lines) != 2 {
		t.Fatalf("lines=%#v err=%v", lines, err)
	}

	entry := terminal.StripANSI(formatDiffEntry(path, diff.Change{LineNumber: 1, Kind: diff.Added, Text: "hello"}))
	if !strings.Contains(entry, path) || !strings.Contains(entry, "ADD") || !strings.Contains(entry, "hello") {
		t.Fatalf("unexpected diff entry: %q", entry)
	}
	if got := formatDiffEntry(path, diff.Change{}); got != "" {
		t.Fatalf("expected empty entry, got %q", got)
	}
}

func writeTestFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "a.txt")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
