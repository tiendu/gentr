package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gentr/internal/utils"
)

type Spinner interface {
	Start()
	Stop()
	Pause()
	Resume()
}

type FileWatcher interface {
	Watch(files []string, command string)
}

type Watcher struct {
	opts     WatchOptions
	spinner  Spinner
	runner   CommandRunner
	reporter OutputReporter
	logger   ChangeLogger

	mu           sync.Mutex
	modTimes     map[string]time.Time
	fileContents map[string][]string
}

type fileEvent struct {
	path    string
	deleted bool
}

func NewWatcher(
	opts WatchOptions,
	spinner Spinner,
	runner CommandRunner,
	reporter OutputReporter,
	logger ChangeLogger,
) *Watcher {
	if runner == nil {
		runner = ShellRunner{}
	}
	if reporter == nil {
		reporter = ConsoleReporter{}
	}
	if logger == nil {
		logger = &SessionLogger{}
	}

	return &Watcher{
		opts:         opts,
		spinner:      spinner,
		runner:       runner,
		reporter:     reporter,
		logger:       logger,
		modTimes:     make(map[string]time.Time),
		fileContents: make(map[string][]string),
	}
}

func (w *Watcher) Watch(files []string, command string) {
	for _, file := range files {
		if err := w.trackFile(file, false); err != nil {
			fmt.Printf("\nError tracking file %s: %v\n", file, err)
		}
	}

	changeChan := make(chan fileEvent, 16)
	go w.pollFiles(changeChan)

	if w.opts.Recursive {
		go w.rescanInputDirectory()
	}

	for event := range changeChan {
		if event.deleted {
			w.handleDeletion(event.path)
			continue
		}

		w.handleChange(event.path, command)
	}
}

func WatchFiles(files []string, command string, opts WatchOptions, spinner Spinner) {
	NewWatcher(opts, spinner, ShellRunner{}, ConsoleReporter{}, defaultSessionLogger).Watch(files, command)
}

func (w *Watcher) pollFiles(changeChan chan<- fileEvent) {
	for {
		for _, file := range w.snapshotFiles() {
			info, err := os.Stat(file)
			if err != nil {
				if os.IsNotExist(err) {
					w.removeFile(file)
					w.sendEvent(changeChan, fileEvent{path: file, deleted: true})
					continue
				}

				fmt.Printf("\nError stating file %s: %v\n", file, err)
				continue
			}

			if info.IsDir() {
				continue
			}

			if w.markModified(file, info.ModTime()) {
				w.sendEvent(changeChan, fileEvent{path: file})
			}
		}

		time.Sleep(w.opts.PollInterval)
	}
}

func (w *Watcher) rescanInputDirectory() {
	for {
		time.Sleep(w.opts.RescanInterval)

		info, err := os.Stat(w.opts.Input)
		if err != nil || !info.IsDir() {
			continue
		}

		_ = filepath.Walk(w.opts.Input, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}

			if err := w.trackFile(path, true); err != nil {
				fmt.Printf("\nError tracking new file %s: %v\n", path, err)
			}

			return nil
		})
	}
}

func (w *Watcher) handleDeletion(path string) {
	fmt.Printf("\nFile deleted: %s\n", path)

	if !w.opts.Log {
		return
	}

	result := CommandResult{
		RawOutput: "",
		ExitCode:  -1,
		Command:   "DELETED",
	}

	if err := w.logger.Write(fmt.Sprintf("%s: DELETED", path), result); err != nil {
		fmt.Printf("\nError writing deletion log: %v\n", err)
	}
}

func (w *Watcher) handleChange(path string, command string) {
	if w.spinner != nil {
		w.spinner.Pause()
		defer w.spinner.Resume()
	}

	timer := time.NewTimer(w.opts.DebounceDuration)
	<-timer.C

	fmt.Printf("\nChange detected in file: %s. Executing command...\n", path)

	newContent, err := readFileLines(path)
	if err != nil {
		fmt.Printf("\nError reading file %s: %v\n", path, err)
		return
	}

	oldContent := w.replaceFileContent(path, newContent)
	result := w.runner.Run(command, path)
	w.reporter.Report(result, w.opts)
	w.printAndLogDiff(path, oldContent, newContent, result)
}

func (w *Watcher) printAndLogDiff(path string, oldContent []string, newContent []string, result CommandResult) {
	changes := CombineModifications(DiffLines(oldContent, newContent))

	for _, change := range changes {
		entry := formatDiffEntry(path, change)
		if entry == "" {
			continue
		}

		fmt.Println(entry)

		if w.opts.Log {
			if err := w.logger.Write(strings.TrimSpace(entry), result); err != nil {
				fmt.Printf("\nError writing log: %v\n", err)
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

	w.mu.Lock()
	_, exists := w.modTimes[path]
	if exists {
		w.mu.Unlock()
		return nil
	}
	w.modTimes[path] = info.ModTime()
	w.mu.Unlock()

	content, err := readFileLines(path)
	if err == nil {
		w.mu.Lock()
		w.fileContents[path] = content
		w.mu.Unlock()
	}

	if announce {
		fmt.Printf("\nNew file detected and added: %s\n", path)
	}

	return nil
}

func (w *Watcher) snapshotFiles() []string {
	w.mu.Lock()
	defer w.mu.Unlock()

	files := make([]string, 0, len(w.modTimes))
	for file := range w.modTimes {
		files = append(files, file)
	}
	return files
}

func (w *Watcher) markModified(path string, modTime time.Time) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	lastMod, exists := w.modTimes[path]
	if !exists || !modTime.After(lastMod) {
		return false
	}

	w.modTimes[path] = modTime
	return true
}

func (w *Watcher) removeFile(path string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	delete(w.modTimes, path)
	delete(w.fileContents, path)
}

func (w *Watcher) replaceFileContent(path string, newContent []string) []string {
	w.mu.Lock()
	defer w.mu.Unlock()

	oldContent, exists := w.fileContents[path]
	if !exists {
		oldContent = []string{}
	}

	w.fileContents[path] = newContent
	return oldContent
}

func (w *Watcher) sendEvent(changeChan chan<- fileEvent, event fileEvent) {
	select {
	case changeChan <- event:
	default:
	}
}

func readFileLines(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(data), "\n"), nil
}

func formatDiffEntry(path string, change DiffChange) string {
	switch change.Type {
	case "MOD":
		return fmt.Sprintf("%s:%d %s: %s\n",
			utils.Bold(utils.Color(path, "cyan")),
			change.LineNumber,
			utils.Bold(utils.Highlight("MOD", "gray", "yellow")),
			utils.Bold(change.Text),
		)
	case "REM":
		return fmt.Sprintf("%s:%d %s: %s\n",
			utils.Bold(utils.Color(path, "cyan")),
			change.LineNumber,
			utils.Bold(utils.Highlight("REM", "white", "red")),
			utils.Bold(utils.Color(change.Text, "red")),
		)
	case "ADD":
		return fmt.Sprintf("%s:%d %s: %s\n",
			utils.Bold(utils.Color(path, "cyan")),
			change.LineNumber,
			utils.Bold(utils.Highlight("ADD", "white", "green")),
			utils.Bold(utils.Color(change.Text, "green")),
		)
	default:
		return ""
	}
}
