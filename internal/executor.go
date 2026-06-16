package internal

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"
)

type CommandRunner interface {
	Run(command string, file string) CommandResult
}

type CommandResult struct {
	RawOutput string
	ExitCode  int
	Command   string
}

type ShellRunner struct{}

func (ShellRunner) Run(command string, file string) CommandResult {
	if strings.Contains(command, "/_") {
		command = strings.ReplaceAll(command, "/_", file)
	}

	cmd := exec.Command("sh", "-c", command)
	out, err := cmd.CombinedOutput()
	exitCode := 0

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
			}
		} else {
			fmt.Println("Error running command:", err)
			exitCode = 1
		}
	}

	return CommandResult{
		RawOutput: string(out),
		ExitCode:  exitCode,
		Command:   command,
	}
}

func RunCommand(command, file string) CommandResult {
	return ShellRunner{}.Run(command, file)
}
