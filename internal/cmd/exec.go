package cmd

import (
	"fmt"
	"os"
	"os/exec"
)

// ExecFn runs a child process. Tests replace this with a mock.
var ExecFn = defaultExec

func defaultExec(executable string, args []string) int {
	cmd := exec.Command(executable, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(os.Stderr, "Error: %s: %v\n", executable, err)
		return 1
	}
	return 0
}
