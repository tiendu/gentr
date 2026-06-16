package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParsePathFlagUsesEnvironmentFirst(t *testing.T) {
	t.Setenv("INSTALL_PATH", "/tmp/from-env")

	got := parsePathFlag([]string{"--path", "/tmp/from-flag"})
	if got != "/tmp/from-env" {
		t.Fatalf("expected env path, got %q", got)
	}
}

func TestParsePathFlagLongAndShort(t *testing.T) {
	t.Setenv("INSTALL_PATH", "")

	if got := parsePathFlag([]string{"--path", "/tmp/long"}); got != "/tmp/long" {
		t.Fatalf("expected long path, got %q", got)
	}
	if got := parsePathFlag([]string{"-p", "/tmp/short"}); got != "/tmp/short" {
		t.Fatalf("expected short path, got %q", got)
	}
}

func TestInstallBinaryCopiesFile(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "source-gentr")
	installDir := filepath.Join(tmp, "bin")

	if err := os.WriteFile(src, []byte("binary-content"), 0644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	dest, err := installBinary(src, installDir)
	if err != nil {
		t.Fatalf("installBinary returned error: %v", err)
	}

	if dest != filepath.Join(installDir, "gentr") {
		t.Fatalf("unexpected dest: %q", dest)
	}

	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("read dest: %v", err)
	}
	if string(data) != "binary-content" {
		t.Fatalf("unexpected copied content: %q", string(data))
	}
}

func TestDetectShell(t *testing.T) {
	t.Setenv("SHELL", "/bin/zsh")
	if got := detectShell(); got != "zsh" {
		t.Fatalf("expected zsh, got %q", got)
	}

	t.Setenv("SHELL", "/usr/bin/bash")
	if got := detectShell(); got != "bash" {
		t.Fatalf("expected bash, got %q", got)
	}

	t.Setenv("SHELL", "/bin/fish")
	if got := detectShell(); got != "unknown" {
		t.Fatalf("expected unknown, got %q", got)
	}
}
