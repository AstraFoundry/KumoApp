package mihomo

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/AstraFoundry/KumoApp/internal/domain/core"
	"github.com/AstraFoundry/KumoApp/internal/infrastructure/process"
)

// Launch describes one start of the core.
type Launch struct {
	CorePath   string
	WorkDir    string
	ConfigPath string
	LogPath    string
	Endpoint   core.Endpoint
}

// Process is a launched core. Wait reports its exit; only the process that
// launched it can wait, other surfaces poll its pid.
type Process struct {
	PID  int
	cmd  *exec.Cmd
	done chan struct{}
	err  error
}

// Done is closed when the core exits.
func (p *Process) Done() <-chan struct{} { return p.done }

// ExitCode is the core's exit code once Done is closed.
func (p *Process) ExitCode() int {
	if p.cmd.ProcessState == nil {
		return -1
	}
	return p.cmd.ProcessState.ExitCode()
}

// Err is the error Wait returned once Done is closed.
func (p *Process) Err() error { return p.err }

// Start launches Mihomo detached from the caller's session, with its output
// appended to the core log.
func Start(l Launch) (*Process, error) {
	if err := os.MkdirAll(l.WorkDir, 0o755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(l.LogPath), 0o755); err != nil {
		return nil, err
	}
	log, err := os.OpenFile(l.LogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(log, "\n--- Kumo started %s at %s ---\n", l.CorePath, time.Now().Format(time.RFC3339))
	args := []string{"-d", l.WorkDir, "-f", l.ConfigPath, "-ext-ctl", l.Endpoint.Address()}
	if l.Endpoint.Secret != "" {
		args = append(args, "-secret", l.Endpoint.Secret)
	}
	cmd := exec.Command(l.CorePath, args...)
	cmd.Dir = l.WorkDir
	cmd.Stdout = log
	cmd.Stderr = log
	cmd.SysProcAttr = process.Detach()
	// The core must not route its own downloads through the system proxy
	// it is about to become.
	cmd.Env = append(os.Environ(), "http_proxy=", "https_proxy=", "all_proxy=", "HTTP_PROXY=", "HTTPS_PROXY=", "ALL_PROXY=")
	if err := cmd.Start(); err != nil {
		log.Close()
		return nil, err
	}
	p := &Process{PID: cmd.Process.Pid, cmd: cmd, done: make(chan struct{})}
	go func() {
		p.err = cmd.Wait()
		log.Close()
		close(p.done)
	}()
	return p, nil
}

// Validate runs the core's configuration test (-t) on a file, so a broken
// profile fails before the running core is replaced.
func Validate(corePath, workDir, configPath string) error {
	cmd := exec.Command(corePath, "-t", "-d", workDir, "-f", configPath)
	cmd.Dir = workDir
	out, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return fmt.Errorf("mihomo rejected the configuration: %s", lastLines(string(out), 3))
	}
	return err
}

func lastLines(text string, n int) string {
	lines := []string{}
	for _, line := range splitLines(text) {
		if line != "" {
			lines = append(lines, line)
		}
	}
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	out := ""
	for i, line := range lines {
		if i > 0 {
			out += " · "
		}
		out += line
	}
	return out
}

func splitLines(text string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(text); i++ {
		if text[i] == '\n' {
			lines = append(lines, text[start:i])
			start = i + 1
		}
	}
	return append(lines, text[start:])
}
