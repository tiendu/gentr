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
        // Enhanced message with color and bold formatting.
        message := beautify.Bold(beautify.Color(fmt.Sprintf("\nChange detected in file: %s. Executing command...", changedFile), "cyan"))
        fmt.Println(message)
        output := RunCommand(command, changedFile)
        FilterLogs(output, opts)
        // Resume the spinner.
        select {
        case spinnerControl <- "resume":
        default:
        }
    }
    // Note: wg.Wait() is unreachable because of the infinite loop.
}

