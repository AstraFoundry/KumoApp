// Package sysproxy implements systemproxy.Platform for each operating
// system, and an in-memory one that tests and simulations drive.
package sysproxy

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Runner runs a system command. Tests replace it.
type Runner interface {
	Run(name string, args ...string) (string, error)
}

// ExecRunner runs commands with os/exec.
type ExecRunner struct{}

// Run returns the trimmed standard output, or an error with standard error.
func (ExecRunner) Run(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}
		return "", fmt.Errorf("%s %s: %s", name, strings.Join(args, " "), message)
	}
	return strings.TrimSpace(string(out)), nil
}
