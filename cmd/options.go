package cmd

import (
    "flag"
    "fmt"
)

// VersionStr and RevisionStr can be set at build time.
var VersionStr = "0.1.0"
var RevisionStr = "unknown"

// ParseOptions handles command-line flags for runtime options.
func ParseOptions() map[string]interface{} {
    // Define flags.
    debug := flag.Bool("debug", false, "Enable debug mode")
    recursive := flag.Bool("recursive", false, "Watch directories recursively")
    // Additional flags can be defined here.

    flag.Parse()
    return map[string]interface{}{
        "debug":     *debug,
        "recursive": *recursive,
    }
}

// GetCommandArgs returns the remaining non-flag arguments.
// These are interpreted as the command to execute when a file change is detected.
func GetCommandArgs() []string {
    return flag.Args()
}

// VersionCommand prints version information.
func VersionCommand() {
    fmt.Printf("gentr version %s, build revision %s\n", VersionStr, RevisionStr)
}

