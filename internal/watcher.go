package internal

import (
    "fmt"
    "os"
    "sync"
    "time"
    "strings"
    "strconv"

    "gentr/cmd"
    "gentr/internal/utils"
)

// debounceDuration is the delay to wait after the last change before executing the command.
var debounceDuration = 500 * time.Millisecond

// WatchFiles watches the given paths and triggers the specified command when a change is detected.
// If opts.Recursive is true, directories are scanned recursively.
// It sends "pause" and "resume" commands on the spinnerControl channel to control the spinner.
func WatchFiles(files []string, command string, opts cmd.Options, spinnerControl chan string) {
    // Map to track the last modification time of each file.
    modTimes := make(map[string]time.Time)

    // Map to store previous content of each file (as slice of lines).
    fileContents := make(map[string][]string)

    for _, file := range files {
        info, err := os.Stat(file)
        if err != nil {
            fmt.Printf("Error stating file %s: %v\n", file, err)
            continue
        }
        modTimes[file] = info.ModTime()
        data, err := os.ReadFile(file)
        if err == nil {
            fileContents[file] = strings.Split(string(data), "\n")
        }
    }

    // Create a channel to signal change events, carrying the changed file's name.
    changeChan := make(chan string, 1)
    var mu sync.Mutex

    // Launch a goroutine per file that polls for changes.
    var wg sync.WaitGroup
    for _, file := range files {
        wg.Add(1)
        go func(f string) {
            defer wg.Done()
            for {
                info, err := os.Stat(f)
                if err != nil {
                    fmt.Printf("Error stating file %s: %v\n", f, err)
                    time.Sleep(1 * time.Second)
                    continue
                }
                mu.Lock()
                lastMod := modTimes[f]
                if info.ModTime().After(lastMod) {
                    modTimes[f] = info.ModTime()
                    // Send the changed file name non-blockingly.
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

    // Debounce change events.
    for {
        changedFile := <-changeChan
        // Pause the spinner.
        select {
        case spinnerControl <- "pause":
        default:
        }
        timer := time.NewTimer(debounceDuration)
        <-timer.C
        fmt.Printf("\nChange detected in file: %s. Executing command...\n", changedFile)
        // Read the new content.
        data, err := os.ReadFile(changedFile)
        if err != nil {
            fmt.Printf("Error reading file %s: %v\n", changedFile, err)
            spinnerControl <- "resume"
            continue
        }
        newContent := strings.Split(string(data), "\n")
        oldContent, exists := fileContents[changedFile]
        if !exists {
            oldContent = []string{}
        }
        // Update the stored content.
        fileContents[changedFile] = newContent

        // Optionally, execute the command and show its output.
        output := RunCommand(command, changedFile)
        FilterLogs(output, opts)

        // Extract exit status from the combined output.
        parts := strings.Split(output, "\n----\n")
        exitStatus := 0
        if len(parts) == 2 {
            statusParts := strings.Split(parts[1], "|")
            if len(statusParts) >= 2 {
                // Convert exit status from string to int.
                var err error
                exitStatus, err = strconv.Atoi(statusParts[1])
                if err != nil {
                    fmt.Printf("Error parsing exit status: %v\n", err)
                }
            }
        }

        // Compute diff.
        diffChanges := DiffLines(oldContent, newContent)
        combinedDiffs := CombineModifications(diffChanges)
        // Print the diff.
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
           // Write log entry
            if opts.Log {
                trimmedEntry := strings.TrimSpace(diffEntry)
                if err := WriteLogEntry(trimmedEntry, exitStatus); err != nil {
                    fmt.Printf("Error writing log: %v\n", err)
                }
            }
        }

        // Resume the spinner.
        spinnerControl <- "resume"
    }
    // Note: wg.Wait() is unreachable because of the infinite loop.
}

