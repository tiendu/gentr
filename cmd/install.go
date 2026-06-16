package cmd

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gentr/internal/utils"
)

// InstallCommand installs the currently running binary.
func InstallCommand(args []string) int {
	installPath := parsePathFlag(args)

	appFilePath, err := os.Executable()
	if err != nil {
		fmt.Printf("[x] Failed to determine executable path: %v\n", err)
		return 1
	}

	appFilePath, err = filepath.EvalSymlinks(appFilePath)
	if err != nil {
		fmt.Printf("[x] Failed to resolve executable path: %v\n", err)
		return 1
	}

	dest, err := installBinary(appFilePath, installPath)
	if err != nil {
		if errors.Is(err, errAlreadyInstalled) {
			fmt.Printf("[!] %v\n", err)
			return 0
		}

		fmt.Printf("[x] Install failed: %v\n", err)
		return 1
	}

	fmt.Printf("[v] Installed gentr to %s\n", dest)
	notifyPathInstruction(installPath)
	return 0
}

// UninstallCommand removes the binary from the specified install path.
func UninstallCommand(args []string) int {
	installPath := parsePathFlag(args)
	destBinary := filepath.Join(installPath, "gentr")

	if _, err := os.Stat(destBinary); err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("[!] Nothing to uninstall at %s\n", destBinary)
			return 0
		}

		fmt.Printf("[x] Failed to inspect %s: %v\n", destBinary, err)
		return 1
	}

	if err := os.Remove(destBinary); err != nil {
		fmt.Printf("[x] Uninstall failed: %v\n", err)
		return 1
	}

	fmt.Printf("[v] Uninstalled gentr from %s\n", destBinary)
	return 0
}

var errAlreadyInstalled = errors.New(
	"the running executable is already the installed binary; " +
		"run the newly built binary or `go run . install` instead",
)

// installBinary atomically replaces the target binary with the running binary.
func installBinary(appPath, installPath string) (string, error) {
	if err := os.MkdirAll(installPath, 0o755); err != nil {
		return "", fmt.Errorf(
			"failed to create directory %s: %w",
			installPath,
			err,
		)
	}

	sourcePath, err := filepath.Abs(appPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve source path: %w", err)
	}

	destBinary, err := filepath.Abs(filepath.Join(installPath, "gentr"))
	if err != nil {
		return "", fmt.Errorf("failed to resolve destination path: %w", err)
	}

	same, err := sameFile(sourcePath, destBinary)
	if err != nil {
		return "", err
	}

	if same {
		return destBinary, errAlreadyInstalled
	}

	if err := copyFileAtomic(sourcePath, destBinary, 0o755); err != nil {
		return "", err
	}

	return destBinary, nil
}

func sameFile(sourcePath, destPath string) (bool, error) {
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return false, fmt.Errorf(
			"failed to inspect source binary %s: %w",
			sourcePath,
			err,
		)
	}

	destInfo, err := os.Stat(destPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, fmt.Errorf(
			"failed to inspect destination binary %s: %w",
			destPath,
			err,
		)
	}

	return os.SameFile(sourceInfo, destInfo), nil
}

// copyFileAtomic copies src to a temporary file and renames it over dst.
//
// This avoids truncating the existing installation before the new binary has
// been copied successfully.
func copyFileAtomic(srcPath, dstPath string, mode os.FileMode) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source binary: %w", err)
	}
	defer src.Close()

	destDir := filepath.Dir(dstPath)

	tempFile, err := os.CreateTemp(destDir, ".gentr-install-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary install file: %w", err)
	}

	tempPath := tempFile.Name()
	keepTemp := false

	defer func() {
		_ = tempFile.Close()

		if !keepTemp {
			_ = os.Remove(tempPath)
		}
	}()

	if _, err := io.Copy(tempFile, src); err != nil {
		return fmt.Errorf("failed to copy binary: %w", err)
	}

	if err := tempFile.Chmod(mode); err != nil {
		return fmt.Errorf("failed to set executable permissions: %w", err)
	}

	if err := tempFile.Sync(); err != nil {
		return fmt.Errorf("failed to flush installed binary: %w", err)
	}

	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close installed binary: %w", err)
	}

	if err := os.Rename(tempPath, dstPath); err != nil {
		return fmt.Errorf("failed to replace %s: %w", dstPath, err)
	}

	keepTemp = true
	return nil
}

// parsePathFlag parses INSTALL_PATH or fallback CLI flags.
func parsePathFlag(args []string) string {
	if envPath := os.Getenv("INSTALL_PATH"); envPath != "" {
		return envPath
	}

	fs := flag.NewFlagSet("install-path", flag.ExitOnError)

	defaultPath := defaultInstallPath()
	installPath := fs.String(
		"path",
		defaultPath,
		"Specify custom install path",
	)
	shortPath := fs.String(
		"p",
		defaultPath,
		"Specify custom install path (short)",
	)

	_ = fs.Parse(args)

	if *shortPath != defaultPath {
		return *shortPath
	}

	return *installPath
}

// defaultInstallPath returns ~/.local/bin or fallback ./gentr.
func defaultInstallPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "./gentr"
	}

	return filepath.Join(home, ".local", "bin")
}

// notifyPathInstruction gives the user PATH update instructions.
func notifyPathInstruction(installPath string) {
	if isInPath(installPath) {
		return
	}

	shell := detectShell()

	fmt.Println()
	fmt.Println(utils.Bold(utils.Color("[!] PATH notice", "yellow")))
	fmt.Printf(
		"The install path %s is %s in your PATH.\n",
		utils.Color(installPath, "cyan"),
		utils.Color("not", "red"),
	)

	fmt.Println()
	fmt.Println(
		utils.Color(
			"To use gentr from anywhere, add it to your PATH:",
			"magenta",
		),
	)

	var exportCommand string

	switch shell {
	case "zsh":
		exportCommand = fmt.Sprintf(
			`echo 'export PATH="%s:$PATH"' >> ~/.zshrc && source ~/.zshrc`,
			installPath,
		)
	case "bash":
		exportCommand = fmt.Sprintf(
			`echo 'export PATH="%s:$PATH"' >> ~/.bashrc && source ~/.bashrc`,
			installPath,
		)
	default:
		exportCommand = fmt.Sprintf(
			"Manually add %s to your PATH.",
			installPath,
		)
	}

	fmt.Println()
	fmt.Println(utils.Color(exportCommand, "green"))

	fmt.Println()
	fmt.Println(
		utils.Color(
			"Restart the terminal or source the shell configuration afterward.",
			"yellow",
		),
	)
}

func detectShell() string {
	shell := os.Getenv("SHELL")

	switch {
	case shell == "", shell == "unknown":
		return "unknown"
	case strings.Contains(shell, "zsh"):
		return "zsh"
	case strings.Contains(shell, "bash"):
		return "bash"
	default:
		return "unknown"
	}
}

func isInPath(dir string) bool {
	cleanDir := filepath.Clean(dir)

	for _, pathEntry := range filepath.SplitList(os.Getenv("PATH")) {
		if filepath.Clean(pathEntry) == cleanDir {
			return true
		}
	}

	return false
}
