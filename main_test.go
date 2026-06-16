package main

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/tiendu/gentr/internal"
)

type stubResolver struct {
	files     []string
	err       error
	input     string
	recursive bool
}

func (r *stubResolver) Resolve(input string, recursive bool) ([]string, error) {
	r.input = input
	r.recursive = recursive
	return r.files, r.err
}

type stubStdinReader struct {
	files []string
	read  bool
}

func (r *stubStdinReader) ReadFiles(reader io.Reader) []string {
	r.read = true
	return r.files
}

func TestInputFilesUsesStdinWhenPiped(t *testing.T) {
	tmp := t.TempDir()
	stdinPath := filepath.Join(tmp, "stdin.txt")
	if err := os.WriteFile(stdinPath, []byte("ignored"), 0644); err != nil {
		t.Fatalf("write stdin file: %v", err)
	}

	stdin, err := os.Open(stdinPath)
	if err != nil {
		t.Fatalf("open stdin file: %v", err)
	}
	defer stdin.Close()

	resolver := &stubResolver{files: []string{"from-resolver"}}
	reader := &stubStdinReader{files: []string{"from-stdin"}}

	got, err := inputFiles(
		stdin,
		resolver,
		reader,
		internal.NewWatchOptions(false, false, ".", 0, false),
	)
	if err != nil {
		t.Fatalf("inputFiles returned error: %v", err)
	}
	if !reflect.DeepEqual(got, []string{"from-stdin"}) {
		t.Fatalf("expected stdin files, got %#v", got)
	}
	if !reader.read {
		t.Fatal("expected stdin reader to be used")
	}
	if resolver.input != "" {
		t.Fatalf("resolver should not be used, got input %q", resolver.input)
	}
}

func TestInputFilesUsesResolverForTerminalLikeStdin(t *testing.T) {
	resolver := &stubResolver{files: []string{"from-resolver"}}
	reader := &stubStdinReader{files: []string{"from-stdin"}}
	opts := internal.NewWatchOptions(false, true, "./src", 0, false)

	got, err := inputFiles(os.Stdin, resolver, reader, opts)
	if err != nil {
		t.Fatalf("inputFiles returned error: %v", err)
	}
	if !reflect.DeepEqual(got, []string{"from-resolver"}) {
		t.Fatalf("expected resolver files, got %#v", got)
	}
	if resolver.input != "./src" || !resolver.recursive {
		t.Fatalf("unexpected resolver call: input=%q recursive=%v", resolver.input, resolver.recursive)
	}
	if reader.read {
		t.Fatal("stdin reader should not be used")
	}
}

func TestInputFilesReturnsResolverError(t *testing.T) {
	wantErr := errors.New("boom")
	resolver := &stubResolver{err: wantErr}
	reader := &stubStdinReader{}

	_, err := inputFiles(
		os.Stdin,
		resolver,
		reader,
		internal.NewWatchOptions(false, false, ".", 0, false),
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected resolver error %v, got %v", wantErr, err)
	}
}

func TestRunHelpReturnsZero(t *testing.T) {
	if code := run([]string{"help"}, os.Stdin); code != 0 {
		t.Fatalf("expected help to return 0, got %d", code)
	}
}

func TestRunUnknownFlagReturnsOne(t *testing.T) {
	if code := run([]string{"--nope"}, os.Stdin); code != 1 {
		t.Fatalf("expected invalid flag to return 1, got %d", code)
	}
}
