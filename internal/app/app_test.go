package app

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/tiendu/gentr/internal/config"
)

type stubResolver struct {
	files     []string
	err       error
	input     string
	recursive bool
}

func (r *stubResolver) Resolve(input string, recursive bool) ([]string, error) {
	r.input, r.recursive = input, recursive
	return r.files, r.err
}

type stubStdinReader struct {
	files []string
	read  bool
}

func (r *stubStdinReader) ReadFiles(io.Reader) []string {
	r.read = true
	return r.files
}

func TestSelectInputUsesPipedStdin(t *testing.T) {
	path := filepath.Join(t.TempDir(), "stdin.txt")
	if err := os.WriteFile(path, []byte("ignored"), 0o644); err != nil {
		t.Fatal(err)
	}
	stdin, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer stdin.Close()

	resolver := &stubResolver{files: []string{"resolver"}}
	reader := &stubStdinReader{files: []string{"stdin"}}
	got, err := selectInput(stdin, resolver, reader, config.New(false, false, ".", 0, false))
	if err != nil || !reflect.DeepEqual(got, []string{"stdin"}) || !reader.read || resolver.input != "" {
		t.Fatalf("got=%#v err=%v reader=%+v resolver=%+v", got, err, reader, resolver)
	}
}

func TestSelectInputUsesResolverForTerminal(t *testing.T) {
	resolver := &stubResolver{files: []string{"resolver"}}
	reader := &stubStdinReader{files: []string{"stdin"}}
	got, err := selectInput(os.Stdin, resolver, reader, config.New(false, true, "./src", 0, false))
	if err != nil || !reflect.DeepEqual(got, []string{"resolver"}) || resolver.input != "./src" || !resolver.recursive || reader.read {
		t.Fatalf("got=%#v err=%v reader=%+v resolver=%+v", got, err, reader, resolver)
	}
}

func TestSelectInputReturnsResolverError(t *testing.T) {
	want := errors.New("boom")
	_, err := selectInput(os.Stdin, &stubResolver{err: want}, &stubStdinReader{}, config.New(false, false, ".", 0, false))
	if !errors.Is(err, want) {
		t.Fatalf("expected %v, got %v", want, err)
	}
}

func TestRunHelpAndInvalidFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"help"}, os.Stdin, &stdout, &stderr); code != 0 {
		t.Fatalf("help returned %d", code)
	}
	if code := Run([]string{"--nope"}, os.Stdin, &stdout, &stderr); code != 1 {
		t.Fatalf("invalid flag returned %d", code)
	}
}
