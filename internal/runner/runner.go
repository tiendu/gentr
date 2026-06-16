package runner

import (
	"os/exec"
	"strings"
)

type Result struct {
	RawOutput string
	ExitCode  int
	Command   string
}

type Runner interface {
	Run(command, file string) Result
}

type Shell struct{}

func (Shell) Run(command, file string) Result {
	resolvedCommand := strings.ReplaceAll(command, "/_", file)
	cmd := exec.Command("sh", "-c", resolvedCommand)
	output, err := cmd.CombinedOutput()

	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = 1
			if len(output) == 0 {
				output = []byte(err.Error())
			}
		}
	}

	return Result{
		RawOutput: string(output),
		ExitCode:  exitCode,
		Command:   resolvedCommand,
	}
}
