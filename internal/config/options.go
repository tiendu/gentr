package config

import (
	"fmt"
	"time"

	"github.com/tiendu/gentr/internal/terminal"
)

type Options struct {
	Debug            bool
	Recursive        bool
	Input            string
	Length           int
	Log              bool
	PollInterval     time.Duration
	DebounceDuration time.Duration
	RescanInterval   time.Duration
}

func New(debug, recursive bool, input string, length int, logEnabled bool) Options {
	return Options{
		Debug:            debug,
		Recursive:        recursive,
		Input:            input,
		Length:           length,
		Log:              logEnabled,
		PollInterval:     time.Second,
		DebounceDuration: 500 * time.Millisecond,
		RescanInterval:   10 * time.Second,
	}
}

func (o Options) String() string {
	formatBool := func(value bool) string {
		if value {
			return terminal.Highlight("true", "white", "green")
		}
		return terminal.Highlight("false", "white", "red")
	}
	formatInt := func(value int) string {
		if value > 0 {
			return terminal.Highlight(fmt.Sprintf("%d", value), "white", "green")
		}
		return terminal.Highlight("none", "white", "red")
	}

	return fmt.Sprintf(
		"--debug %s; --recursive %s; --length %s; --log %s; --input %s",
		formatBool(o.Debug),
		formatBool(o.Recursive),
		formatInt(o.Length),
		formatBool(o.Log),
		terminal.Bold(terminal.Color(o.Input, "cyan")),
	)
}
