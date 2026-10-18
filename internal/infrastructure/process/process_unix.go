//go:build !windows

package process

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Alive reports whether pid exists. A zombie (an exited child not yet
// reaped) counts as gone.
func Alive(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, 0)
	if err != nil && !errors.Is(err, syscall.EPERM) {
		return false
	}
	if runtime.GOOS == "linux" {
		if data, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/stat"); err == nil {
			if i := strings.LastIndexByte(string(data), ')'); i > 0 && i+2 < len(data) && data[i+2] == 'Z' {
				return false
			}
		}
	} else if state := output("ps", "-o", "stat=", "-p", strconv.Itoa(pid)); strings.HasPrefix(state, "Z") {
		return false
	}
	return true
}

// Name is the executable name of pid, or "" when unknown.
func Name(pid int) string {
	if runtime.GOOS == "linux" {
		if exe, err := os.Readlink("/proc/" + strconv.Itoa(pid) + "/exe"); err == nil {
			return filepath.Base(strings.TrimSuffix(exe, " (deleted)"))
		}
		if data, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/comm"); err == nil {
			return strings.TrimSpace(string(data))
		}
	}
	return filepath.Base(output("ps", "-o", "comm=", "-p", strconv.Itoa(pid)))
}

type step struct {
	signal syscall.Signal
	wait   time.Duration
}

func (s step) send(pid int) error { return syscall.Kill(pid, s.signal) }

func steps(grace time.Duration) []step {
	return []step{{syscall.SIGINT, grace / 2}, {syscall.SIGTERM, grace}, {syscall.SIGKILL, time.Second}}
}

// Detach makes a launched process the leader of its own session, so a
// terminal's interrupt or a quitting parent does not take it down.
func Detach() *syscall.SysProcAttr { return &syscall.SysProcAttr{Setsid: true} }

// ParentAlive reports whether the process that started this one runs.
func ParentAlive(parent int) bool { return os.Getppid() == parent && Alive(parent) }
