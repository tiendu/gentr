package internal

import (
    "fmt"
    "os"
    "sync"
    "time"

    "gentr/cmd"
    "gentr/internal/beautify"
)

// debounceDuration is the delay to wait after the last change before executing the command.
var debounceDuration = 500 * time.Millisecond

// WatchFiles watches the given paths and triggers the specified command when a change is detected.
// If opts.Recursive is true, directories are scanned recursively.
// It sends "pause" and "resume" commands on the spinnerControl channel to control the spinner.
func WatchFiles(files []string, command string, opts cmd.Options, spinnerControl chan string) {
    // Map to track the last modification time of each file.
    modTimes := make(map[string]time.Time)
    for _, file := range files {
        info, err := os.Stat(file)
        if err != nil {
            fmt.Printf("Error stating file %s: %v\n", file, err)
            continue
        }
        modTimes[file] = info.ModTime()
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
        // Compute diff.
        added, removed := DiffLines(oldContent, newContent)
        // Print added lines.
        for _, line := range added {
            fmt.Printf("%s %s %s\n",
                beautify.Bold(beautify.Color(changedFile, "cyan")),
                beautify.Bold(beautify.Highlight("ADD", "white", "green")),
                beautify.Bold(beautify.Color(line, "green")),
            )
        }
        // Print removed lines.
        for _, line := range removed {
            fmt.Printf("%s %s %s\n",
                beautify.Bold(beautify.Color(changedFile, "cyan")),
                beautify.Bold(beautify.Highlight("REM", "white", "red")),
                beautify.Bold(beautify.Color(line, "red")),
            )
        }
        // Update the stored content.
        fileContents[changedFile] = newContent

        // Optionally, execute the command and show its output.
        output := RunCommand(command, changedFile)
        FilterLogs(output, opts)

        // Resume the spinner.
        spinnerControl <- "resume"
    }
    // Note: wg.Wait() is unreachable because of the infinite loop.
}

