package internal

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Resolver interface {
	Resolve(input string, recursive bool) ([]string, error)
}

type StdinReader interface {
	ReadFiles(r io.Reader) []string
}

type FileResolver struct{}

type LineStdinReader struct{}

func (FileResolver) Resolve(input string, recursive bool) ([]string, error) {
	if strings.ContainsAny(input, "*?[]") {
		matches, err := filepath.Glob(input)
		if err != nil {
			return nil, fmt.Errorf("error processing glob pattern: %w", err)
		}
		return matches, nil
	}

	info, err := os.Stat(input)
	if err != nil {
		return nil, fmt.Errorf("error accessing input %s: %w", input, err)
	}

	if !info.IsDir() {
		return []string{input}, nil
	}

	if recursive {
		return walkFiles(input)
	}

	return listTopLevelFiles(input)
}

func (LineStdinReader) ReadFiles(r io.Reader) []string {
	var files []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			files = append(files, line)
		}
	}
	return files
}

func ResolveFiles(opts WatchOptions) ([]string, error) {
	return FileResolver{}.Resolve(opts.Input, opts.Recursive)
}

func ReadFilesFromStdin() []string {
	return LineStdinReader{}.ReadFiles(os.Stdin)
}

func walkFiles(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking directory %s: %w", root, err)
	}
	return files, nil
}

func listTopLevelFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("error reading directory %s: %w", dir, err)
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}
	return files, nil
}
