package internal

import (
    "fmt"
    "os"
    "sync"
    "time"
)

// WatchFiles concurrently watches multiple files and executes a command when any file changes.
func WatchFiles(files []string, command string) {
    changeChan := make(chan string)
    var wg sync.WaitGroup

    // Start a goroutine for each file.
    for _, file := range files {
        wg.Add(1)
        go func(f string) {
            defer wg.Done()
            lastModified := time.Now()
            for {
                stat, err := os.Stat(f)
                if err != nil {
                    fmt.Printf("Error accessing file %s: %v\n", f, err)
                    time.Sleep(time.Second)
                    continue
                }
                // If the file has been modified, send a notification.
                if stat.ModTime().After(lastModified) {
                    changeChan <- f
                    lastModified = stat.ModTime()
                }
                time.Sleep(time.Second) // Adjust polling frequency as needed.
            }
        }(file)
    }

    // Goroutine to listen for change notifications and run the command.
    go func() {
        for changedFile := range changeChan {
            fmt.Printf("File changed: %s\n", changedFile)
            output := RunCommand(command)
            FilterLogs(output)
        }
    }()

    // Block indefinitely; note that wg.Wait() will never return since the watcher loops are infinite.
    wg.Wait()
}

