package cmd

import (
	"flag"
	"fmt"
	"gentr/internal/utils"
)

type Options struct {
	Debug     bool
	Recursive bool
	Input     string
	Length    int
	Log       bool
}

func ParseOptions() Options {
	var opts Options

	flag.BoolVar(&opts.Debug, "debug", false, "Enable debug mode")
	flag.BoolVar(&opts.Debug, "d", false, "Enable debug mode (short)")
	flag.BoolVar(&opts.Recursive, "recursive", false, "Watch directories recursively")
	flag.BoolVar(&opts.Recursive, "r", false, "Watch directories recursively (short)")
	flag.IntVar(&opts.Length, "length", 0, "Limit output lines")
	flag.IntVar(&opts.Length, "l", 0, "Limit output lines (short)")
	flag.StringVar(&opts.Input, "input", ".", "Input path")
	flag.StringVar(&opts.Input, "i", ".", "Input path (short)")
	flag.BoolVar(&opts.Log, "log", false, "Enable logging")

	flag.Parse()
	return opts
}

func GetCommandArgs() []string {
	return flag.Args()
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
