// Package clilink puts the kumo command on the user's PATH by linking it
// to the binary inside the app.
package clilink

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Status is where the command is and what it points at.
type Status struct {
	Path      string `json:"path"`
	Installed bool   `json:"installed"`
	// Managed is true when the link points at this Kumo.
	Managed bool   `json:"managed"`
	Target  string `json:"target,omitempty"`
	// Source is the binary a new link would point at.
	Source    string `json:"source"`
	Supported bool   `json:"supported"`
}

// LinkPath is where the command goes: /usr/local/bin on macOS, the user's
// ~/.local/bin on Linux.
func LinkPath() string {
	if runtime.GOOS == "linux" {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".local", "bin", "kumo")
	}
	return "/usr/local/bin/kumo"
}

// Source is the executable the link points at: this process's own binary,
// which runs the CLI when invoked as kumo.
func Source() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		return resolved
	}
	return exe
}

// Check reports the link's state.
func Check() Status {
	s := Status{Path: LinkPath(), Source: Source(), Supported: runtime.GOOS != "windows"}
	target, err := os.Readlink(s.Path)
	if err != nil {
		if _, statErr := os.Stat(s.Path); statErr == nil {
			s.Installed = true
		}
		return s
	}
	s.Installed, s.Target = true, target
	resolved, _ := filepath.EvalSymlinks(s.Path)
	s.Managed = resolved == s.Source || strings.Contains(target, "Kumo")
	return s
}

// Install links the command, asking for an administrator's password when
// the directory needs it.
func Install() (Status, error) {
	s := Check()
	if !s.Supported {
		return s, errors.New("add Kumo's folder to PATH to use kumo on Windows")
	}
	if s.Installed && !s.Managed {
		return s, fmt.Errorf("%s already exists and is not Kumo's; remove it first", s.Path)
	}
	if err := link(s.Source, s.Path); err != nil {
		if runtime.GOOS != "darwin" || !errors.Is(err, os.ErrPermission) {
			return Check(), err
		}
		script := fmt.Sprintf(`do shell script "mkdir -p %s && ln -sf %s %s" with administrator privileges with prompt "Kumo wants to install the kumo command."`,
			shellQuote(filepath.Dir(s.Path)), shellQuote(s.Source), shellQuote(s.Path))
		if out, err := exec.Command("/usr/bin/osascript", "-e", script).CombinedOutput(); err != nil {
			return Check(), fmt.Errorf("the kumo command was not installed: %s", strings.TrimSpace(string(out)))
		}
	}
	return Check(), nil
}

func link(source, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	_ = os.Remove(path)
	return os.Symlink(source, path)
}

// Uninstall removes a link Kumo made; anything else is left alone.
func Uninstall() (Status, error) {
	s := Check()
	if !s.Installed {
		return s, nil
	}
	if !s.Managed {
		return s, fmt.Errorf("%s is not Kumo's, so it was left in place", s.Path)
	}
	if err := os.Remove(s.Path); err != nil {
		if runtime.GOOS != "darwin" || !errors.Is(err, os.ErrPermission) {
			return Check(), err
		}
		script := fmt.Sprintf(`do shell script "rm -f %s" with administrator privileges`, shellQuote(s.Path))
		if out, err := exec.Command("/usr/bin/osascript", "-e", script).CombinedOutput(); err != nil {
			return Check(), fmt.Errorf("the kumo command was not removed: %s", strings.TrimSpace(string(out)))
		}
	}
	return Check(), nil
}

func shellQuote(s string) string {
	return `'` + strings.ReplaceAll(strings.ReplaceAll(s, `'`, `'\''`), `"`, `\"`) + `'`
}
