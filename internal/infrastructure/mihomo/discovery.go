package mihomo

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
)

// Candidate is a Mihomo executable Kumo found.
type Candidate struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Source string `json:"source"`
}

var names = []string{"mihomo", "mihomo-alpha", "clash-meta", "clash"}

// Discover lists executables in order of preference: the one the user
// chose, the one Kumo manages, $KUMO_MIHOMO_PATH, one shipped beside the
// app, then any on $PATH or in common install locations.
func Discover(selected, managed, bundledDir string) []Candidate {
	var found []Candidate
	seen := map[string]bool{}
	add := func(path, source string) {
		if path == "" {
			return
		}
		if abs, err := filepath.Abs(path); err == nil {
			path = abs
		}
		if seen[path] || !executable(path) {
			return
		}
		seen[path] = true
		found = append(found, Candidate{Name: filepath.Base(path), Path: path, Source: source})
	}
	add(selected, "selected")
	add(managed, "managed")
	add(os.Getenv("KUMO_MIHOMO_PATH"), "environment")
	if bundledDir != "" {
		for _, name := range names {
			add(filepath.Join(bundledDir, exe(name)), "bundled")
		}
	}
	for _, name := range names {
		if path, err := exec.LookPath(exe(name)); err == nil {
			add(path, "path")
		}
	}
	for _, dir := range searchDirs() {
		entries, _ := os.ReadDir(dir)
		for _, entry := range entries {
			name := strings.ToLower(entry.Name())
			if slices.ContainsFunc(names, func(n string) bool { return strings.HasPrefix(name, n) }) {
				add(filepath.Join(dir, entry.Name()), dir)
			}
		}
	}
	return found
}

func exe(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

func executable(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	if runtime.GOOS == "windows" {
		return strings.HasSuffix(strings.ToLower(path), ".exe")
	}
	return info.Mode()&0o111 != 0
}

func searchDirs() []string {
	home, _ := os.UserHomeDir()
	switch runtime.GOOS {
	case "windows":
		return nil
	default:
		return []string{"/opt/homebrew/bin", "/usr/local/bin", "/usr/bin", filepath.Join(home, ".local", "bin"), filepath.Join(home, "bin")}
	}
}
