// Package paths names every file Kumo keeps. The app, the CLI and the
// service resolve the same directory, so they share one state.
package paths

import (
	"os"
	"path/filepath"
)

// AppName names the data directory, as the app's name does.
const AppName = "Kumo"

// Paths is rooted at the data directory.
type Paths struct{ Root string }

// Default resolves the data directory: $KUMO_HOME when set, else the user
// configuration directory of the platform (~/Library/Application Support
// on macOS, ~/.config on Linux, %AppData% on Windows) joined with Kumo.
func Default() (Paths, error) {
	if home := os.Getenv("KUMO_HOME"); home != "" {
		return Paths{Root: home}, nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return Paths{}, err
	}
	return Paths{Root: filepath.Join(dir, AppName)}, nil
}

func (p Paths) join(parts ...string) string {
	return filepath.Join(append([]string{p.Root}, parts...)...)
}

func (p Paths) Profiles() string         { return p.join("profiles") }
func (p Paths) ProfileMetadata() string  { return p.join("profiles", "profiles-metadata.json") }
func (p Paths) CurrentProfile() string   { return p.join("profiles", "current.txt") }
func (p Paths) Overrides() string        { return p.join("overrides") }
func (p Paths) OverrideMetadata() string { return p.join("overrides", "overrides.json") }
func (p Paths) OverrideFiles() string    { return p.join("overrides", "files") }
func (p Paths) Work() string             { return p.join("work") }
func (p Paths) RuntimeConfig() string    { return p.join("work", "config.yaml") }
func (p Paths) CorePID() string          { return p.join("work", "core.pid") }
func (p Paths) Logs() string             { return p.join("logs") }
func (p Paths) CoreLog() string          { return p.join("logs", "core.log") }
func (p Paths) State() string            { return p.join("state.json") }
func (p Paths) StateLock() string        { return p.join(".state.lock") }
func (p Paths) Preferences() string      { return p.join("preferences.json") }
func (p Paths) ManagedCores() string     { return p.join("cores") }
func (p Paths) AgentSkillsState() string { return p.join("agent-skills-state.json") }
func (p Paths) GuardianPID() string      { return p.join("work", "guardian.pid") }

// ManagedCore is the Mihomo executable Kumo downloads itself.
func (p Paths) ManagedCore() string {
	name := "mihomo"
	if os.PathSeparator == '\\' {
		name += ".exe"
	}
	return p.join("cores", name)
}

// Prepare creates the directories Kumo writes into.
func (p Paths) Prepare() error {
	for _, dir := range []string{p.Root, p.Profiles(), p.Work(), p.Logs(), p.OverrideFiles(), p.ManagedCores()} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}
