package cmd

import (
    "flag"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "strings"

    "gentr/internal/utils"
)

// VersionStr and RevisionStr can be set at build time.
var VersionStr = "0.2.0"
var RevisionStr = "unknown"

// Options holds command-line flag settings.
type Options struct {
    Debug     bool
    Recursive bool
    Input     string
    Length    int
    Log       bool
}

// String implements the Stringer interface for pretty-printing the options with enhanced formatting.
func (o Options) String() string {
    var debugVal, recursiveVal, lengthVal, logVal string
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
    if o.Log {
        logVal = utils.Highlight("true", "white", "green")
    } else {
        logVal = utils.Highlight("false", "white", "red")
    }
    // For input, we just color it normally (or you can also choose to highlight).
    inputVal := utils.Bold(utils.Color(o.Input, "cyan"))
    return fmt.Sprintf("--debug %s; --recursive %s; --length %s; --log %s; --input %s", debugVal, recursiveVal, lengthVal, logVal, inputVal)
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

    // Boolean flag for log.
    flag.BoolVar(&opts.Log, "log", false, "Enable logging")

    flag.Parse()
    return opts
}

// GetCommandArgs returns the remaining non-flag arguments.
func GetCommandArgs() []string {
    return flag.Args()
}

// HelpCommand displays help message for gentr CLI.
func HelpCommand() {
    fmt.Printf(`Usage: gentr <command> [options]

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

  --log
       Enable logging (writes output to a .log file).

  --input, -i
       Input directory, file, or glob pattern (e.g., '.', 'logs/*.log'). Specifies the files or directories to monitor for changes.
       Example: gentr --input 'logs/*.log' --recursive

  INSTALL_PATH
       Override the default installation path (e.g., '/usr/local/bin')
       Example: INSTALL_PATH=/usr/local/bin gentr install`)
}

// InstallCommand installs this binary to a specific directory.
func InstallCommand(args []string) {
    installPath := parsePathFlag(args)

    appFilePath, err := os.Executable()
    if err != nil {
        fmt.Printf("Error determining executable path: %v\n", err)
        return
    }

    dest, err := installSubCmd(appFilePath, installPath)
    if err != nil {
        fmt.Printf("❌ Install failed: %v\n", err)
        return
    }

    fmt.Printf("✅ Installed gentr to %s\n", dest)
    checkPathNotice(installPath)
}

// installSubCmd copies the current binary to a desired installation path.
func installSubCmd(appFilePath, installPath string) (string, error) {
    if err := os.MkdirAll(installPath, 0755); err != nil {
        return "", fmt.Errorf("failed to create directory %s: %v", installPath, err)
    }

    destBinary := filepath.Join(installPath, "gentr")

    if err := copyFile(appFilePath, destBinary, 0755); err != nil {
        return "", err
    }

    return destBinary, nil
}

// UninstallCommand removes the installed gentr binary.
func UninstallCommand(args []string) {
    installPath := parsePathFlag(args)
    destBinary := filepath.Join(installPath, "gentr")

    if _, err := os.Stat(destBinary); os.IsNotExist(err) {
        fmt.Printf("Nothing to uninstall at %s\n", destBinary)
        return
    }

    if err := os.Remove(destBinary); err != nil {
        fmt.Printf("❌ Uninstall failed: %v\n", err)
        return
    }

    fmt.Printf("✅ Uninstalled gentr from %s\n", destBinary)
}

// ReinstallCommand uninstalls and then reinstalls the tool.
func ReinstallCommand(args []string) {
    UninstallCommand(args)
    InstallCommand(args)
}

// VersionCommand prints version information.
func VersionCommand() {
    fmt.Printf("gentr version %s, build revision %s\n", VersionStr, RevisionStr)
}

// Helpers
// parsePathFlag parses the INSTALL_PATH environment variable or defaults.
func parsePathFlag(args []string) string {
    if envPath := os.Getenv("INSTALL_PATH"); envPath != "" {
        return envPath
    }

    fs := flag.NewFlagSet("path", flag.ExitOnError)
    defaultPath := defaultInstallPath()
    installPath := fs.String("path", defaultPath, "Specify custom install path")
    installPathShort := fs.String("p", defaultPath, "Specify custom install path (shorthand)")
    _ = fs.Parse(args)

    if *installPathShort != defaultPath {
        return *installPathShort
    }

    return *installPath
}

// defaultInstallPath returns the default installation path ~/.local/bin
func defaultInstallPath() string {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return "./gentr"
    }
    return filepath.Join(homeDir, ".local", "bin")
}

// copyFile copies a file from src to dst with given permissions.
func copyFile(srcPath, destPath string, mode os.FileMode) error {
    srcFile, err := os.Open(srcPath)
    if err != nil {
        return err
    }
    defer srcFile.Close()

    destFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
    if err != nil {
        return err
    }
    defer destFile.Close()

    _, err = io.Copy(destFile, srcFile)
    return err
}

// checkPathNotice advises the user to add install path to PATH if it's not already.
func checkPathNotice(installPath string) {
    if isInPath(installPath) {
        return
    }

    shell := detectShell()

    fmt.Println()
    fmt.Println(utils.Bold(utils.Color("⚠️  PATH Notice:", "yellow")))
    fmt.Printf("Your install path %s is %s in your PATH.\n",
        utils.Color(installPath, "cyan"),
        utils.Color("not", "red"),
    )

    fmt.Println(utils.Color("\nTo use gentr from anywhere, add it to your PATH:", "magenta"))

    switch shell {
    case "zsh":
        fmt.Printf("\n%s\n", utils.Color(
            fmt.Sprintf("echo 'export PATH=\"%s:$PATH\"' >> ~/.zshrc && source ~/.zshrc", installPath),
            "green",
        ))
    case "bash":
        fmt.Printf("\n%s\n", utils.Color(
            fmt.Sprintf("echo 'export PATH=\"%s:$PATH\"' >> ~/.bashrc && source ~/.bashrc", installPath),
            "green",
        ))
    default:
        fmt.Println(utils.Color(
            fmt.Sprintf("\nYour shell could not be detected. Please manually add %s to your PATH.", installPath),
            "yellow",
        ))
    }

    fmt.Println(utils.Color("\nAfter adding, restart your terminal or run the above command to apply changes immediately.", "yellow"))
    fmt.Println()
}

// detectShell tries to detect the user's shell from environment.
func detectShell() string {
    shellEnv := os.Getenv("SHELL")
    if strings.Contains(shellEnv, "zsh") {
        return "zsh"
    }
    if strings.Contains(shellEnv, "bash") {
        return "bash"
    }
    return "unknown"
}

// isInPath checks if the given directory is in the PATH environment variable.
func isInPath(dir string) bool {
    pathEnv := os.Getenv("PATH")
    for _, p := range filepath.SplitList(pathEnv) {
        if p == dir {
            return true
        }
    }
    return false
}
