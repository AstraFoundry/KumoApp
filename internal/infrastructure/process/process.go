// Package process inspects and ends processes by pid, the way Kumo finds a
// core it launched in an earlier run or from another surface.
package process

import (
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Terminate asks pid to stop, then insists: an interrupt, a termination
// request, then a kill, each with time to exit. It reports whether the
// process is gone.
func Terminate(pid int, grace time.Duration) bool {
	if pid <= 0 || !Alive(pid) {
		return true
	}
	for _, step := range steps(grace) {
		if err := step.send(pid); err != nil && !Alive(pid) {
			return true
		}
		if waitExit(pid, step.wait) {
			return true
		}
	}
	return !Alive(pid)
}

func waitExit(pid int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !Alive(pid) {
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}
	return !Alive(pid)
}

// IsCore reports whether pid runs a Mihomo-like executable, so a pid file
// left by a crash never makes Kumo kill an unrelated process that reused
// the pid.
func IsCore(pid int, corePath string) bool {
	name := strings.ToLower(Name(pid))
	if name == "" {
		return false
	}
	if corePath != "" && strings.EqualFold(name, strings.ToLower(filepath.Base(corePath))) {
		return true
	}
	return strings.Contains(name, "mihomo") || strings.Contains(name, "clash")
}

// output runs a command and returns its trimmed standard output.
func output(name string, args ...string) string {
	out, err := exec.Command(name, args...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
