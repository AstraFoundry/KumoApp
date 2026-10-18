package process

import (
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/windows"
)

// Alive reports whether pid exists and has not exited.
func Alive(pid int) bool {
	if pid <= 0 {
		return false
	}
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return false
	}
	defer windows.CloseHandle(h)
	var code uint32
	if err := windows.GetExitCodeProcess(h, &code); err != nil {
		return false
	}
	return code == 259 // STILL_ACTIVE
}

// Name is the executable name of pid, or "" when unknown.
func Name(pid int) string {
	out := output("tasklist", "/FI", "PID eq "+strconv.Itoa(pid), "/FO", "CSV", "/NH")
	if fields := strings.Split(out, ","); len(fields) > 0 && strings.HasPrefix(fields[0], `"`) {
		return strings.TrimSuffix(strings.Trim(fields[0], `"`), ".exe")
	}
	return ""
}

type step struct {
	force bool
	wait  time.Duration
}

func (s step) send(pid int) error {
	args := []string{"/PID", strconv.Itoa(pid), "/T"}
	if s.force {
		args = append(args, "/F")
	}
	return runQuiet("taskkill", args...)
}

func steps(grace time.Duration) []step {
	return []step{{false, grace}, {true, time.Second}}
}

func runQuiet(name string, args ...string) error {
	p, err := os.StartProcess(lookPath(name), append([]string{name}, args...), &os.ProcAttr{Sys: &syscall.SysProcAttr{HideWindow: true}})
	if err != nil {
		return err
	}
	_, err = p.Wait()
	return err
}

func lookPath(name string) string {
	return os.Getenv("SystemRoot") + `\System32\` + name + ".exe"
}

// Detach starts the process in a new process group without a console.
func Detach() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{CreationFlags: windows.CREATE_NEW_PROCESS_GROUP | windows.CREATE_NO_WINDOW, HideWindow: true}
}

// ParentAlive reports whether the process that started this one runs.
func ParentAlive(parent int) bool { return Alive(parent) }
