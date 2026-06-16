package internal

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLineStdinReaderTrimsBlankLines(t *testing.T) {
	input := strings.NewReader(" a.go \n\n\tb.go\n")
	got := LineStdinReader{}.ReadFiles(input)
	want := []string{"a.go", "b.go"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %#v, got %#v", want, got)
	}
}

func TestFileResolverResolvesSingleFile(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "a.txt")
	if err := os.WriteFile(file, []byte("a"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	got, err := FileResolver{}.Resolve(file, false)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if !reflect.DeepEqual(got, []string{file}) {
		t.Fatalf("unexpected files: %#v", got)
	}
}

func TestFileResolverListsTopLevelFiles(t *testing.T) {
	tmp := t.TempDir()
	rootFile := filepath.Join(tmp, "root.txt")
	nestedDir := filepath.Join(tmp, "nested")
	nestedFile := filepath.Join(nestedDir, "nested.txt")

	if err := os.Mkdir(nestedDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	for _, file := range []string{rootFile, nestedFile} {
		if err := os.WriteFile(file, []byte("x"), 0644); err != nil {
			t.Fatalf("write file: %v", err)
		}
	}

	got, err := FileResolver{}.Resolve(tmp, false)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if !reflect.DeepEqual(got, []string{rootFile}) {
		t.Fatalf("expected only top-level file, got %#v", got)
	}
}

func TestFileResolverWalksRecursively(t *testing.T) {
	tmp := t.TempDir()
	rootFile := filepath.Join(tmp, "root.txt")
	nestedDir := filepath.Join(tmp, "nested")
	nestedFile := filepath.Join(nestedDir, "nested.txt")

	if err := os.Mkdir(nestedDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	for _, file := range []string{rootFile, nestedFile} {
		if err := os.WriteFile(file, []byte("x"), 0644); err != nil {
			t.Fatalf("write file: %v", err)
		}
	}

	got, err := FileResolver{}.Resolve(tmp, true)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}

	want := map[string]bool{rootFile: true, nestedFile: true}
	if len(got) != len(want) {
		t.Fatalf("expected %d files, got %#v", len(want), got)
	}
	for _, file := range got {
		if !want[file] {
			t.Fatalf("unexpected file: %s", file)
		}
	}
}

func TestFileResolverGlob(t *testing.T) {
	tmp := t.TempDir()
	match := filepath.Join(tmp, "a.go")
	other := filepath.Join(tmp, "b.txt")
	for _, file := range []string{match, other} {
		if err := os.WriteFile(file, []byte("x"), 0644); err != nil {
			t.Fatalf("write file: %v", err)
		}
	}

	got, err := FileResolver{}.Resolve(filepath.Join(tmp, "*.go"), false)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if !reflect.DeepEqual(got, []string{match}) {
		t.Fatalf("expected glob match %#v, got %#v", []string{match}, got)
	}
}
