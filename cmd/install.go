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

// InstallCommand installs the binary to a specified or default path.
func InstallCommand(args []string) {
	installPath := parsePathFlag(args)

	appFilePath, err := os.Executable()
	if err != nil {
		fmt.Printf("Error determining executable path: %v\n", err)
		return
	}

	dest, err := installBinary(appFilePath, installPath)
	if err != nil {
		fmt.Printf("❌ Install failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Installed gentr to %s\n", dest)
	notifyPathInstruction(installPath)
}

// UninstallCommand removes the binary from the specified install path.
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

// ReinstallCommand uninstalls and reinstalls the binary.
func ReinstallCommand(args []string) {
	UninstallCommand(args)
	InstallCommand(args)
}

// installBinary copies the compiled binary to the target path.
func installBinary(appPath, installPath string) (string, error) {
	if err := os.MkdirAll(installPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", installPath, err)
	}

	destBinary := filepath.Join(installPath, "gentr")
	if err := copyFile(appPath, destBinary, 0755); err != nil {
		return "", err
	}
	return destBinary, nil
}

// parsePathFlag parses INSTALL_PATH env or fallback CLI flags.
func parsePathFlag(args []string) string {
	if envPath := os.Getenv("INSTALL_PATH"); envPath != "" {
		return envPath
	}

	fs := flag.NewFlagSet("install-path", flag.ExitOnError)
	defaultPath := defaultInstallPath()
	installPath := fs.String("path", defaultPath, "Specify custom install path")
	short := fs.String("p", defaultPath, "Specify custom install path (short)")
	_ = fs.Parse(args)

	if *short != defaultPath {
		return *short
	}
	return *installPath
}

// defaultInstallPath returns ~/.local/bin or fallback ./gentr
func defaultInstallPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "./gentr"
	}
	return filepath.Join(home, ".local", "bin")
}

// copyFile copies binary from src to dst.
func copyFile(srcPath, dstPath string, mode os.FileMode) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

// notifyPathInstruction gives the user PATH update tips.
func notifyPathInstruction(installPath string) {
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

	var exportCmd string
	switch shell {
	case "zsh":
		exportCmd = fmt.Sprintf(`echo 'export PATH="%s:$PATH"' >> ~/.zshrc && source ~/.zshrc`, installPath)
	case "bash":
		exportCmd = fmt.Sprintf(`echo 'export PATH="%s:$PATH"' >> ~/.bashrc && source ~/.bashrc`, installPath)
	default:
		exportCmd = fmt.Sprintf(`Please manually add %s to your PATH.`, installPath)
	}

	fmt.Println(utils.Color("\n"+exportCmd, "green"))
	fmt.Println(utils.Color("\nAfter adding, restart your terminal or source your shell config.", "yellow"))
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
	for _, p := range filepath.SplitList(os.Getenv("PATH")) {
		if p == dir {
			return true
		}
	}
	return false
}
