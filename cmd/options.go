package cmd

import (
    "flag"
    "fmt"

    "gentr/internal/beautify"
)

// VersionStr and RevisionStr can be set at build time.
var VersionStr = "0.1.0"
var RevisionStr = "unknown"

// Options holds command-line flag settings.
type Options struct {
    Debug     bool
    Recursive bool
    Input     string
    Length    int
}

// String implements the Stringer interface for pretty-printing the options with enhanced formatting.
func (o Options) String() string {
    var debugVal, recursiveVal, lengthVal string
    if o.Debug {
        debugVal = beautify.Highlight("true", "white", "green")
    } else {
        debugVal = beautify.Highlight("false", "white", "red")
    }
    if o.Recursive {
        recursiveVal = beautify.Highlight("true", "white", "green")
    } else {
        recursiveVal = beautify.Highlight("false", "white", "red")
    }
    if o.Length > 0 {
        lengthVal = beautify.Highlight(fmt.Sprintf("%d", o.Length), "white", "green")
    } else {
        lengthVal = beautify.Highlight("none", "white", "red")
    }
    // For input, we just color it normally (or you can also choose to highlight).
    inputVal := beautify.Bold(beautify.Color(o.Input, "cyan"))
    return fmt.Sprintf("--debug %s; --recursive %s; --length %s; --input %s", debugVal, recursiveVal, lengthVal, inputVal)
}

// ParseOptions parses command-line flags and returns an Options struct.
// It supports both long and short flag names.
func ParseOptions() Options {
    var opts Options

    // Boolean flags for debug.
    flag.BoolVar(&opts.Debug, "debug", false, "Enable debug mode")
    flag.BoolVar(&opts.Debug, "d", false, "Enable debug mode (short)")

    // Boolean flags for recursive.
    flag.BoolVar(&opts.Recursive, "recursive", false, "Watch directories recursively")
    flag.BoolVar(&opts.Recursive, "r", false, "Watch directories recursively (short)")

    // Integer flag for length.
    flag.IntVar(&opts.Length, "length", 0, "Limit the number of output lines to this length")
    flag.IntVar(&opts.Length, "l", 0, "Limit the number of output lines to this length (short)")

    // String flags for input.
    flag.StringVar(&opts.Input, "input", ".", "Input directory, file, or glob pattern (e.g., '.', 'logs/*.log')")
    flag.StringVar(&opts.Input, "i", ".", "Input directory, file, or glob pattern (short)")

    flag.Parse()
    return opts
}

// GetCommandArgs returns the remaining non-flag arguments.
func GetCommandArgs() []string {
    return flag.Args()
}

// VersionCommand prints version information.
func VersionCommand() {
    fmt.Printf("gentr version %s, build revision %s\n", VersionStr, RevisionStr)
}

