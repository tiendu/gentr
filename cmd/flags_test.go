package cmd

import "testing"

func TestParseOptionsDefaults(t *testing.T) {
	opts, commandArgs, err := ParseOptions([]string{"echo", "ok"})
	if err != nil {
		t.Fatalf("ParseOptions returned error: %v", err)
	}

	if opts.Debug || opts.Recursive || opts.Log {
		t.Fatalf("expected boolean defaults to be false, got %+v", opts)
	}
	if opts.Input != "." {
		t.Fatalf("expected default input '.', got %q", opts.Input)
	}
	if opts.Length != 0 {
		t.Fatalf("expected default length 0, got %d", opts.Length)
	}
	if len(commandArgs) != 2 || commandArgs[0] != "echo" || commandArgs[1] != "ok" {
		t.Fatalf("unexpected command args: %#v", commandArgs)
	}
}

func TestParseOptionsLongFlags(t *testing.T) {
	opts, commandArgs, err := ParseOptions([]string{
		"--debug",
		"--recursive",
		"--log",
		"--length", "5",
		"--input", "./src",
		"go", "test", "./...",
	})
	if err != nil {
		t.Fatalf("ParseOptions returned error: %v", err)
	}

	if !opts.Debug || !opts.Recursive || !opts.Log {
		t.Fatalf("expected debug, recursive, and log to be true, got %+v", opts)
	}
	if opts.Length != 5 {
		t.Fatalf("expected length 5, got %d", opts.Length)
	}
	if opts.Input != "./src" {
		t.Fatalf("expected input ./src, got %q", opts.Input)
	}
	if len(commandArgs) != 3 || commandArgs[0] != "go" || commandArgs[1] != "test" || commandArgs[2] != "./..." {
		t.Fatalf("unexpected command args: %#v", commandArgs)
	}
}

func TestParseOptionsShortFlags(t *testing.T) {
	opts, commandArgs, err := ParseOptions([]string{
		"-d",
		"-r",
		"-l", "3",
		"-i", "./pkg",
		"cat", "/_",
	})
	if err != nil {
		t.Fatalf("ParseOptions returned error: %v", err)
	}

	if !opts.Debug || !opts.Recursive {
		t.Fatalf("expected debug and recursive to be true, got %+v", opts)
	}
	if opts.Length != 3 {
		t.Fatalf("expected length 3, got %d", opts.Length)
	}
	if opts.Input != "./pkg" {
		t.Fatalf("expected input ./pkg, got %q", opts.Input)
	}
	if len(commandArgs) != 2 || commandArgs[0] != "cat" || commandArgs[1] != "/_" {
		t.Fatalf("unexpected command args: %#v", commandArgs)
	}
}

func TestParseOptionsRejectsUnknownFlag(t *testing.T) {
	_, _, err := ParseOptions([]string{"--nope"})
	if err == nil {
		t.Fatal("expected unknown flag to return an error")
	}
}
