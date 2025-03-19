package internal

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "sync"
    "time"

    "gentr/cmd"
    "gentr/internal/utils"
)

// debounceDuration is the delay to wait after the last change before executing the command.
var debounceDuration = 500 * time.Millisecond

// updateFileList rescans directories (from the initial file list) and adds any new files
// to the modTimes and fileContents maps.
func updateFileList(baseDirs []string, modTimes map[string]time.Time, fileContents map[string][]string) {
    for _, dir := range baseDirs {
        filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
            if err != nil || info.IsDir() {
                return nil
            }
            // If the file is not already tracked, add it.
            if _, exists := modTimes[path]; !exists {
                modTimes[path] = info.ModTime()
                data, err := os.ReadFile(path)
                if err == nil {
                    fileContents[path] = strings.Split(string(data), "\n")
                    fmt.Printf("\nNew file detected and added: %s\n", path)
                }
            }
            return nil
        })
    }
}

// WatchFiles monitors the given files and triggers the specified command when a change is detected.
// It uses a map to store previous file contents so that diffs can be computed, prints changes,
// logs them if logging is enabled, and controls a spinner via spinnerControl.
func WatchFiles(files []string, command string, opts cmd.Options, spinnerControl chan string) {
    // Map for last modification times.
    modTimes := make(map[string]time.Time)
    // Map for storing previous content of each file (split into lines).
    fileContents := make(map[string][]string)
    
    // Initialize modTimes and fileContents from the initial file list.
    for _, file := range files {
        info, err := os.Stat(file)
        if err != nil {
            fmt.Printf("\nError stating file %s: %v\n", file, err)
            continue
        }
        modTimes[file] = info.ModTime()
        data, err := os.ReadFile(file)
        if err == nil {
            fileContents[file] = strings.Split(string(data), "\n")
        }
    }

    // Channel to signal change events (file path).
    changeChan := make(chan string, 1)
    var mu sync.Mutex

    // Launch a goroutine per file to poll for changes.
    var wg sync.WaitGroup
    for _, file := range files {
        wg.Add(1)
        go func(f string) {
            defer wg.Done()
            for {
                info, err := os.Stat(f)
                if err != nil {
                    // If the file is deleted.
                    if os.IsNotExist(err) {
                        mu.Lock()
                        delete(modTimes, f)
                        delete(fileContents, f)
                        mu.Unlock()
                        // Send a deletion event with "delete:" prefix.
                        select {
                        case changeChan <- "delete:" + f:
                        default:
                        }
                        // End polling for this file.
                        return
                    }
                    fmt.Printf("\nError stating file %s: %v\n", f, err)
                    time.Sleep(1 * time.Second)
                    continue
                }
                mu.Lock()
                lastMod := modTimes[f]
                if info.ModTime().After(lastMod) {
                    modTimes[f] = info.ModTime()
                    select {
                    case changeChan <- f:
                    default:
                    }
                }
                mu.Unlock()
                time.Sleep(1 * time.Second)
            }
        }(file)
    }

    // If recursive mode is enabled, update file list periodically.
    if opts.Recursive {
        // If opts.Input is a directory, re-scan that directory.
        baseDir := opts.Input
        info, err := os.Stat(baseDir)
        if err == nil && info.IsDir() {
            ticker := time.NewTicker(10 * time.Second)
            go func() {
                for range ticker.C {
                    updateFileList([]string{baseDir}, modTimes, fileContents)
                }
            }()
        }
    }

    // Process change events.
    for {
        changedEvent := <-changeChan
        // Check if this is a deletion event.
        if strings.HasPrefix(changedEvent, "delete:") {
            deletedFile := strings.TrimPrefix(changedEvent, "delete:")
            fmt.Printf("\nFile deleted: %s\n", deletedFile)
            if opts.Log {
                deletionEntry := fmt.Sprintf("%s: DELETED", deletedFile)
                // Create a CommandResult for deletion.
                delResult := CommandResult{
                    RawOutput: "",
                    ExitCode:  -1,
                    Command:   "DELETED",
                }
                if err := WriteLogEntry(deletionEntry, delResult); err != nil {
                    fmt.Printf("\nError writing deletion log: %v\n", err)
                }
            }
            continue
        }
        changedFile := changedEvent
        // Pause the spinner.
        spinnerControl <- "pause"
        timer := time.NewTimer(debounceDuration)
        <-timer.C

        fmt.Printf("\nChange detected in file: %s. Executing command...\n", changedFile)
        // Read new file content.
        data, err := os.ReadFile(changedFile)
        if err != nil {
            fmt.Printf("\nError reading file %s: %v\n", changedFile, err)
            spinnerControl <- "resume"
            continue
        }
        newContent := strings.Split(string(data), "\n")
        oldContent, exists := fileContents[changedFile]
        if !exists {
            oldContent = []string{}
        }
        // Update stored content.
        fileContents[changedFile] = newContent

        // Execute the command; RunCommand returns a CommandResult.
        cr := RunCommand(command, changedFile)
        // Display output (using FilterLogs). You can modify FilterLogs to accept CommandResult directly.
        FilterLogs(cr, opts)

        // Compute diff between old and new content.
        diffChanges := DiffLines(oldContent, newContent)
        combinedDiffs := CombineModifications(diffChanges)
        for _, change := range combinedDiffs {
            var diffEntry string
            switch change.Type {
            case "MOD":
                diffEntry = fmt.Sprintf("%s:%d %s: %s\n",
                    utils.Bold(utils.Color(changedFile, "cyan")),
                    change.LineNumber,
                    utils.Bold(utils.Highlight("MOD", "gray", "yellow")),
                    utils.Bold(change.Text),
                )
            case "REM": 
                diffEntry = fmt.Sprintf("%s:%d %s: %s\n",
                    utils.Bold(utils.Color(changedFile, "cyan")),
                    change.LineNumber,
                    utils.Bold(utils.Highlight("REM", "white", "red")),
                    utils.Bold(utils.Color(change.Text, "red")),
                )
            case "ADD": 
                diffEntry = fmt.Sprintf("%s:%d %s: %s\n",
                    utils.Bold(utils.Color(changedFile, "cyan")),
                    change.LineNumber,
                    utils.Bold(utils.Highlight("ADD", "white", "green")),
                    utils.Bold(utils.Color(change.Text, "green")),
                )
            }
            fmt.Println(diffEntry)
            // Write log entry if logging is enabled.
            if opts.Log {
                trimmedEntry := strings.TrimSpace(diffEntry)
                if err := WriteLogEntry(trimmedEntry, cr); err != nil {
                    fmt.Printf("\nError writing log: %v\n", err)
                }
            }
        }

        // Resume the spinner.
        spinnerControl <- "resume"
    }
    // Note: wg.Wait() is unreachable due to the infinite loop.
}

