package internal

import (
    "fmt"
    "bufio"
    "os"
    "path/filepath"
    "strings"

    "gentr/cmd"
)

// ResolveFiles returns a slice of file paths based on the --input flag.
// If the input contains glob characters, it resolves them using filepath.Glob.
// If the input is a directory and opts.Recursive is true, it walks the directory recursively.
// Otherwise, if it's a directory and recursive is false, it returns only the files in the top-level directory.
// If the input is a single file, it returns that file.
func ResolveFiles(opts cmd.Options) ([]string, error) {
    var files []string

    // Check if input looks like a glob pattern.
    if strings.ContainsAny(opts.Input, "*?[]") {
        matches, err := filepath.Glob(opts.Input)
        if err != nil {
            return nil, fmt.Errorf("error processing glob pattern: %w", err)
        }
        files = matches
    } else {
        // Input is a file or directory.
        info, err := os.Stat(opts.Input)
        if err != nil {
            return nil, fmt.Errorf("error accessing input %s: %w", opts.Input, err)
        }
        if info.IsDir() {
            if opts.Recursive {
                // Walk the directory recursively.
                err := filepath.Walk(opts.Input, func(p string, info os.FileInfo, err error) error {
                    if err != nil {
                        return err
                    }
                    if !info.IsDir() {
                        files = append(files, p)
                    }
                    return nil
                })
                if err != nil {
                    return nil, fmt.Errorf("error walking directory %s: %w", opts.Input, err)
                }
            } else {
                // Not recursive: list only top-level files.
                entries, err := os.ReadDir(opts.Input)
                if err != nil {
                    return nil, fmt.Errorf("error reading directory %s: %w", opts.Input, err)
                }
                for _, entry := range entries {
                    if !entry.IsDir() {
                        files = append(files, filepath.Join(opts.Input, entry.Name()))
                    }
                }
            }
        } else {
            // A single file.
            files = []string{opts.Input}
        }
    }
    return files, nil
}

// ReadFilesFromStdin reads lines from STDIN and returns them as a slice of strings.
func ReadFilesFromStdin() []string {
    var files []string
    scanner := bufio.NewScanner(os.Stdin)
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if line != "" {
            files = append(files, line)
        }
    }
    return files
}

