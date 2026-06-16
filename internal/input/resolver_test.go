package input

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLineStdinReaderTrimsBlankLines(t *testing.T) {
	got := LineStdinReader{}.ReadFiles(strings.NewReader(" a.go \n\n\tb.go\n"))
	want := []string{"a.go", "b.go"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %#v, got %#v", want, got)
	}
}

func TestFileResolverResolvesFileDirectoryAndGlob(t *testing.T) {
	tmp := t.TempDir()
	nested := filepath.Join(tmp, "nested")
	if err := os.Mkdir(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	rootGo := filepath.Join(tmp, "a.go")
	rootText := filepath.Join(tmp, "b.txt")
	nestedGo := filepath.Join(nested, "c.go")
	for _, path := range []string{rootGo, rootText, nestedGo} {
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	resolver := FileResolver{}

	got, err := resolver.Resolve(rootGo, false)
	if err != nil || !reflect.DeepEqual(got, []string{rootGo}) {
		t.Fatalf("single file: got=%#v err=%v", got, err)
	}

	got, err = resolver.Resolve(tmp, false)
	if err != nil || !reflect.DeepEqual(got, []string{rootGo, rootText}) {
		t.Fatalf("top-level directory: got=%#v err=%v", got, err)
	}

	got, err = resolver.Resolve(tmp, true)
	if err != nil || !reflect.DeepEqual(got, []string{rootGo, rootText, nestedGo}) {
		t.Fatalf("recursive directory: got=%#v err=%v", got, err)
	}

	got, err = resolver.Resolve(filepath.Join(tmp, "*.go"), false)
	if err != nil || !reflect.DeepEqual(got, []string{rootGo}) {
		t.Fatalf("glob: got=%#v err=%v", got, err)
	}
}
