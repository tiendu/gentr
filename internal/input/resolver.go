package input

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Resolver interface {
	Resolve(input string, recursive bool) ([]string, error)
}

type StdinReader interface {
	ReadFiles(reader io.Reader) []string
}

type FileResolver struct{}

type LineStdinReader struct{}

func (FileResolver) Resolve(value string, recursive bool) ([]string, error) {
	if strings.ContainsAny(value, "*?[]") {
		matches, err := filepath.Glob(value)
		if err != nil {
			return nil, fmt.Errorf("process glob pattern: %w", err)
		}
		sort.Strings(matches)
		return matches, nil
	}

	info, err := os.Stat(value)
	if err != nil {
		return nil, fmt.Errorf("access input %s: %w", value, err)
	}

	if !info.IsDir() {
		return []string{value}, nil
	}

	if recursive {
		return walkFiles(value)
	}

	return listTopLevelFiles(value)
}

func (LineStdinReader) ReadFiles(reader io.Reader) []string {
	files := make([]string, 0)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			files = append(files, line)
		}
	}
	return files
}

func walkFiles(root string) ([]string, error) {
	files := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk directory %s: %w", root, err)
	}
	sort.Strings(files)
	return files, nil
}

func listTopLevelFiles(directory string) ([]string, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("read directory %s: %w", directory, err)
	}

	files := make([]string, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, filepath.Join(directory, entry.Name()))
		}
	}
	sort.Strings(files)
	return files, nil
}
