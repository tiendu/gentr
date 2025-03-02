package cmd

import (
    "flag"
    "fmt"
)

// VersionStr and RevisionStr can be set at build time.
var VersionStr = "0.1.0"
var RevisionStr = "unknown"

// Options holds command-line flag settings.
type Options struct {
    Debug     bool
    Recursive bool
}

// String implements the Stringer interface for pretty-printing.
func (o Options) String() string {
    return fmt.Sprintf("--debug %t; --recursive %t", o.Debug, o.Recursive)
}

// ParseOptions parses command-line flags and returns an Options struct.
func ParseOptions() Options {
    var opts Options
    flag.BoolVar(&opts.Debug, "debug", false, "Enable debug mode")
    flag.BoolVar(&opts.Recursive, "recursive", false, "Watch directories recursively")
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

