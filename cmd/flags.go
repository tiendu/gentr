package cmd

import (
	"flag"
	"fmt"
	"io"

	"gentr/internal/utils"
)

type Options struct {
	Debug     bool
	Recursive bool
	Input     string
	Length    int
	Log       bool
}

func ParseOptions(args []string) (Options, []string, error) {
	var opts Options

	fs := flag.NewFlagSet("gentr", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	fs.BoolVar(&opts.Debug, "debug", false, "Enable debug mode")
	fs.BoolVar(&opts.Debug, "d", false, "Enable debug mode (short)")
	fs.BoolVar(&opts.Recursive, "recursive", false, "Watch directories recursively")
	fs.BoolVar(&opts.Recursive, "r", false, "Watch directories recursively (short)")
	fs.IntVar(&opts.Length, "length", 0, "Limit output lines")
	fs.IntVar(&opts.Length, "l", 0, "Limit output lines (short)")
	fs.StringVar(&opts.Input, "input", ".", "Input path")
	fs.StringVar(&opts.Input, "i", ".", "Input path (short)")
	fs.BoolVar(&opts.Log, "log", false, "Enable logging")

	if err := fs.Parse(args); err != nil {
		return Options{}, nil, err
	}

	return opts, fs.Args(), nil
}

func (o Options) String() string {
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
