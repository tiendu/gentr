package internal

import (
    "fmt"
    "os"
    "path/filepath"
    "sync"
    "time"

    "gentr/cmd"
)

// debounceDuration is the delay to wait after the last change before executing the command.
var debounceDuration = 500 * time.Millisecond

// WatchFiles watches the given paths and triggers the specified command when a change is detected.
// If opts.Recursive is true, directories are scanned recursively.
func WatchFiles(paths []string, command string, opts cmd.Options) {
    // Expand provided paths.
    var files []string
    for _, path := range paths {
        info, err := os.Stat(path)
        if err != nil {
            fmt.Printf("Error accessing %s: %v\n", path, err)
            continue
        }
        if info.IsDir() {
            if opts.Recursive {
                err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
                    if err != nil {
                        return err
                    }
                    if !info.IsDir() {
                        files = append(files, p)
                    }
                    return nil
                })
                if err != nil {
                    fmt.Printf("Error walking directory %s: %v\n", path, err)
                }
            } else {
                // Without recursive, add the directory itself.
                files = append(files, path)
            }
        } else {
            files = append(files, path)
        }
    }

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
        timer := time.NewTimer(debounceDuration)
        <-timer.C
        fmt.Printf("Change detected in file: %s. Executing command...\n", changedFile)
        output := RunCommand(command, changedFile)
        FilterLogs(output)
    }
    // Note: wg.Wait() is unreachable because of the infinite loop.
}

