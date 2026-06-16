package internal

import (
	"fmt"
	"time"

	"github.com/tiendu/gentr/internal/utils"
)

type WatchOptions struct {
	Debug            bool
	Recursive        bool
	Input            string
	Length           int
	Log              bool
	PollInterval     time.Duration
	DebounceDuration time.Duration
	RescanInterval   time.Duration
}

func NewWatchOptions(debug, recursive bool, input string, length int, log bool) WatchOptions {
	return WatchOptions{
		Debug:            debug,
		Recursive:        recursive,
		Input:            input,
		Length:           length,
		Log:              log,
		PollInterval:     1 * time.Second,
		DebounceDuration: 500 * time.Millisecond,
		RescanInterval:   10 * time.Second,
	}
}

func (o WatchOptions) String() string {
	formatBool := func(val bool) string {
		if val {
			return utils.Highlight("true", "white", "green")
		}
		return utils.Highlight("false", "white", "red")
	}
	formatInt := func(val int) string {
		if val > 0 {
			return utils.Highlight(fmt.Sprintf("%d", val), "white", "green")
		}
		return utils.Highlight("none", "white", "red")
	}
	return fmt.Sprintf("--debug %s; --recursive %s; --length %s; --log %s; --input %s",
		formatBool(o.Debug),
		formatBool(o.Recursive),
		formatInt(o.Length),
		formatBool(o.Log),
		utils.Bold(utils.Color(o.Input, "cyan")),
	)
}
