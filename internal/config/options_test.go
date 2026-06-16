package config

import (
	"strings"
	"testing"
	"time"

	"github.com/tiendu/gentr/internal/terminal"
)

func TestNewUsesOperationalDefaults(t *testing.T) {
	opts := New(true, true, "./src", 10, true)

	if !opts.Debug || !opts.Recursive || !opts.Log {
		t.Fatalf("unexpected boolean options: %+v", opts)
	}
	if opts.Input != "./src" || opts.Length != 10 {
		t.Fatalf("unexpected input options: %+v", opts)
	}
	if opts.PollInterval != time.Second || opts.DebounceDuration != 500*time.Millisecond || opts.RescanInterval != 10*time.Second {
		t.Fatalf("unexpected timing defaults: %+v", opts)
	}
}

func TestOptionsString(t *testing.T) {
	text := terminal.StripANSI(New(false, true, ".", 0, false).String())
	for _, expected := range []string{"--debug false", "--recursive true", "--length none", "--log false", "--input ."} {
		if !strings.Contains(text, expected) {
			t.Fatalf("expected %q in %q", expected, text)
		}
	}
}
