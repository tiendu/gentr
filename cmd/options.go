package cmd

import (
    "flag"
    "fmt"
    "io"
    "os"
    "os/exec"
    "path/filepath"

    "gentr/internal/utils"
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
        debugVal = utils.Highlight("true", "white", "green")
    } else {
        debugVal = utils.Highlight("false", "white", "red")
    }
    if o.Recursive {
        recursiveVal = utils.Highlight("true", "white", "green")
    } else {
        recursiveVal = utils.Highlight("false", "white", "red")
    }
    if o.Length > 0 {
        lengthVal = utils.Highlight(fmt.Sprintf("%d", o.Length), "white", "green")
    } else {
        lengthVal = utils.Highlight("none", "white", "red")
    }
    // For input, we just color it normally (or you can also choose to highlight).
    inputVal := utils.Bold(utils.Color(o.Input, "cyan"))
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

// HelpCommand displays the usage and available commands/options.
func HelpCommand() {
    fmt.Println(`Usage: gentr <command> [options]

Commands:
  install      Install this tool as a system command
  reinstall    Uninstall and then reinstall the tool
  uninstall    Remove the installed tool
  version      Print version information
  bump         Bump the version automatically
  help         Display this help message

Options:
  --debug, -d
       Enable debug mode (Displays verbose output during execution)

  --recursive, -r
       Watch directories recursively. When enabled, gentr traverses directories and watches all files found.

  --length, -l
       Limit the number of output lines to this length. If set, only the most recent lines of output will be displayed.

  --input, -i
       Input directory, file, or glob pattern (e.g., '.', 'logs/*.log'). Specifies the files or directories to monitor for changes.
       Example: gentr --input 'logs/*.log' --recursive

  INSTALL_PATH
       Override the default installation path (e.g., '/usr/local/bin')
       Example: INSTALL_PATH=/usr/local/bin gentr install`)
}

// InstallCommand installs this binary to a specific directory
func InstallCommand() {
    appFilePath, err := os.Executable()
    if err != nil {
        fmt.Printf("Error determining executable path: %v\n", err)
        return
    }

    dest, err := installSubCmd(appFilePath, "entr")
    if err != nil {
        fmt.Printf("Install failed, err=%v\n", err)
    } else {
        fmt.Printf("Installed entr to %s\n", dest)
    }
}

// installSubCmd copies the current binary to a desired installation path.
func installSubCmd(appFilePath, subCmd string) (string, error) {
    // Allow override of installation path using an environment variable
    execPath := os.Getenv("INSTALL_PATH")
    if execPath == "" {
        execPath = "/usr/local/bin" // Default path if not set.
    }

    destPath := filepath.Join(execPath, subCmd)

    // Ensure the exec directory exists.
    if _, err := os.Stat(execPath); os.IsNotExist(err) {
        if err := os.MkdirAll(execPath, 0755); err != nil {
            return "", fmt.Errorf("failed to create directory %s: %v", execPath, err)
        }
    }

    srcFile, err := os.Open(appFilePath)
    if err != nil {
        return "", err
    }
    defer srcFile.Close()

    destFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY, 0755)
    if err != nil {
        return "", err
    }
    defer destFile.Close()

    if _, err := io.Copy(destFile, srcFile); err != nil {
        return "", err
    }
    return destPath, nil
}

// UninstallCommand removes the installed Git subcommand (entr).
func UninstallCommand() {
    path, err := exec.LookPath("entr")
    if err != nil {
        fmt.Printf("entr not found in PATH, nothing to uninstall.\n")
        return
    }
    if err := os.Remove(path); err != nil {
        fmt.Printf("Uninstall failed: %v\n", err)
        return
    }
    fmt.Printf("Uninstalled entr from %s\n", path)
}

// ReinstallCommand uninstalls and then reinstalls the tool.
func ReinstallCommand() {
    UninstallCommand()
    InstallCommand()
}

// VersionCommand prints version information.
func VersionCommand() {
    fmt.Printf("entr version %s, build revision %s\n", VersionStr, RevisionStr)
}

