// Package skills installs Kumo's bundled Agent Skills into the skill
// directories of coding agents, so they know how to drive the kumo CLI.
package skills

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/AstraFoundry/KumoApp/internal/infrastructure/storage"
)

//go:embed all:bundle
var bundle embed.FS

// Scope is where skills install: the user's home, or the current project.
type Scope string

const (
	ScopeGlobal  Scope = "global"
	ScopeProject Scope = "project"
)

// Target is a coding agent.
type Target string

const (
	Cursor Target = "cursor"
	Claude Target = "claude"
	Codex  Target = "codex"
	Gemini Target = "gemini"
	Agents Target = "agents"
)

// Targets lists every agent in display order.
var Targets = []Target{Claude, Codex, Cursor, Gemini, Agents}

// SupportsProject reports whether an agent reads project skills.
func (t Target) SupportsProject() bool { return t != Codex && t != Gemini }

// ParseTarget validates an agent name.
func ParseTarget(name string) (Target, error) {
	for _, t := range Targets {
		if string(t) == name {
			return t, nil
		}
	}
	return "", fmt.Errorf("unknown agent %q: use cursor, claude, codex, gemini, agents or all", name)
}

type manifest struct {
	BundleVersion string `json:"bundleVersion"`
	Skills        []struct {
		ID           string `json:"id"`
		RelativePath string `json:"relativePath"`
	} `json:"skills"`
}

// Entry records one install, so Kumo knows which directories it owns.
type Entry struct {
	Target          string    `json:"target"`
	Scope           string    `json:"scope"`
	DestinationRoot string    `json:"destinationRoot"`
	SkillIDs        []string  `json:"installedSkillIds"`
	BundleVersion   string    `json:"bundleVersion"`
	InstalledAt     time.Time `json:"installedAt"`
}

type state struct {
	Entries []Entry `json:"entries"`
}

// TargetStatus is one agent's install state.
type TargetStatus struct {
	Target          Target   `json:"target"`
	Scope           Scope    `json:"scope"`
	DestinationRoot string   `json:"destinationRoot"`
	Supported       bool     `json:"supported"`
	Installed       bool     `json:"installed"`
	UpToDate        bool     `json:"upToDate"`
	SkillIDs        []string `json:"installedSkillIds"`
	// Detected is true when the agent's configuration directory exists.
	Detected bool `json:"detected"`
}

// Status is every selected agent's state.
type Status struct {
	BundleVersion string         `json:"bundleVersion"`
	SkillIDs      []string       `json:"skillIds"`
	Scope         Scope          `json:"scope"`
	Targets       []TargetStatus `json:"targets"`
}

// Report says what an install or uninstall did, or would do.
type Report struct {
	DestinationRoots []string `json:"destinationRoots"`
	BundleVersion    string   `json:"bundleVersion"`
	SkillIDs         []string `json:"skillIds"`
	DryRun           bool     `json:"dryRun"`
}

// Installer installs the bundle.
type Installer struct {
	StateFile string
	Home      string
	// Project is the working directory for project scope.
	Project string
}

// New returns an installer for the current user.
func New(stateFile string) (*Installer, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	wd, _ := os.Getwd()
	return &Installer{StateFile: stateFile, Home: home, Project: wd}, nil
}

func (i *Installer) manifest() (manifest, error) {
	var m manifest
	data, err := bundle.ReadFile("bundle/manifest.json")
	if err != nil {
		return m, err
	}
	return m, json.Unmarshal(data, &m)
}

func (i *Installer) agentHome(t Target) string {
	switch t {
	case Codex:
		if home := strings.TrimSpace(os.Getenv("CODEX_HOME")); home != "" {
			return home
		}
		return filepath.Join(i.Home, ".codex")
	default:
		return filepath.Join(i.Home, "."+string(t))
	}
}

// Root is where an agent keeps skills in a scope, or "" when unsupported.
func (i *Installer) Root(t Target, scope Scope) string {
	if scope == ScopeProject {
		if !t.SupportsProject() {
			return ""
		}
		return filepath.Join(i.Project, "."+string(t), "skills")
	}
	return filepath.Join(i.agentHome(t), "skills")
}

func (i *Installer) entries() []Entry {
	var s state
	_, _ = storage.ReadJSON(i.StateFile, &s)
	return s.Entries
}

func (i *Installer) save(entries []Entry) error {
	return storage.WriteJSON(i.StateFile, state{Entries: entries})
}

func resolve(targets []Target, scope Scope) []Target {
	if len(targets) == 0 {
		targets = Targets
	}
	var out []Target
	for _, t := range targets {
		if scope == ScopeGlobal || t.SupportsProject() {
			out = append(out, t)
		}
	}
	return out
}

// Status reports each agent's install.
func (i *Installer) Status(targets []Target, scope Scope) (Status, error) {
	m, err := i.manifest()
	if err != nil {
		return Status{}, err
	}
	ids := skillIDs(m)
	status := Status{BundleVersion: m.BundleVersion, SkillIDs: ids, Scope: scope, Targets: []TargetStatus{}}
	entries := i.entries()
	for _, t := range resolve(targets, scope) {
		root := i.Root(t, scope)
		ts := TargetStatus{Target: t, Scope: scope, DestinationRoot: root, Supported: root != "", SkillIDs: []string{}}
		if _, err := os.Stat(i.agentHome(t)); err == nil {
			ts.Detected = true
		}
		for _, e := range entries {
			if e.Target == string(t) && e.Scope == string(scope) && e.DestinationRoot == root {
				ts.Installed = true
				ts.SkillIDs = e.SkillIDs
				ts.UpToDate = e.BundleVersion == m.BundleVersion && slices.Equal(sorted(e.SkillIDs), sorted(ids))
			}
		}
		status.Targets = append(status.Targets, ts)
	}
	return status, nil
}

func skillIDs(m manifest) []string {
	ids := []string{}
	for _, s := range m.Skills {
		ids = append(ids, s.ID)
	}
	return ids
}

func sorted(values []string) []string {
	out := slices.Clone(values)
	slices.Sort(out)
	return out
}

var validName = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// Install copies the bundle into each agent's skill directory. A skill
// directory Kumo did not install is left alone unless force is set.
func (i *Installer) Install(targets []Target, scope Scope, dryRun, force bool) (Report, error) {
	m, err := i.manifest()
	if err != nil {
		return Report{}, err
	}
	selected := resolve(targets, scope)
	if len(selected) == 0 {
		return Report{}, errors.New("select at least one agent that supports this scope")
	}
	entries := i.entries()
	report := Report{BundleVersion: m.BundleVersion, SkillIDs: skillIDs(m), DryRun: dryRun, DestinationRoots: []string{}}
	for _, t := range selected {
		root := i.Root(t, scope)
		if root == "" {
			return Report{}, fmt.Errorf("%s does not support %s scope", t, scope)
		}
		owned := slices.ContainsFunc(entries, func(e Entry) bool { return e.Target == string(t) && e.DestinationRoot == root })
		for _, s := range m.Skills {
			if !validName.MatchString(s.ID) {
				return Report{}, fmt.Errorf("invalid skill id %q", s.ID)
			}
			if _, err := os.Stat(filepath.Join(root, s.ID)); err == nil && !owned && !force {
				return Report{}, fmt.Errorf("%s already exists and was not installed by Kumo; pass --force to replace it", filepath.Join(root, s.ID))
			}
		}
		report.DestinationRoots = append(report.DestinationRoots, root)
	}
	if dryRun {
		return report, nil
	}
	now := time.Now()
	for _, t := range selected {
		root := i.Root(t, scope)
		for _, s := range m.Skills {
			dest := filepath.Join(root, s.ID)
			if err := os.RemoveAll(dest); err != nil {
				return Report{}, err
			}
			if err := copyBundle(path.Join("bundle", s.RelativePath), dest); err != nil {
				return Report{}, err
			}
		}
		entries = slices.DeleteFunc(entries, func(e Entry) bool { return e.Target == string(t) && e.DestinationRoot == root })
		entries = append(entries, Entry{Target: string(t), Scope: string(scope), DestinationRoot: root, SkillIDs: report.SkillIDs, BundleVersion: m.BundleVersion, InstalledAt: now})
	}
	return report, i.save(entries)
}

// Uninstall removes skills Kumo installed for the selected agents.
func (i *Installer) Uninstall(targets []Target, scope Scope, dryRun bool) (Report, error) {
	selected := resolve(targets, scope)
	entries := i.entries()
	report := Report{DryRun: dryRun, DestinationRoots: []string{}, SkillIDs: []string{}}
	var kept []Entry
	for _, e := range entries {
		match := e.Scope == string(scope) && slices.ContainsFunc(selected, func(t Target) bool {
			return string(t) == e.Target && i.Root(t, scope) == e.DestinationRoot
		})
		if !match {
			kept = append(kept, e)
			continue
		}
		report.DestinationRoots = append(report.DestinationRoots, e.DestinationRoot)
		report.BundleVersion = e.BundleVersion
		for _, id := range e.SkillIDs {
			if !slices.Contains(report.SkillIDs, id) {
				report.SkillIDs = append(report.SkillIDs, id)
			}
			if !dryRun && validName.MatchString(id) {
				if err := os.RemoveAll(filepath.Join(e.DestinationRoot, id)); err != nil {
					return Report{}, err
				}
			}
		}
	}
	if len(report.DestinationRoots) == 0 {
		return report, errors.New("no matching agent skill install")
	}
	if dryRun {
		return report, nil
	}
	return report, i.save(kept)
}

func copyBundle(src, dest string) error {
	return fs.WalkDir(bundle, src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(strings.TrimPrefix(p, src), "/")
		target := filepath.Join(dest, filepath.FromSlash(rel))
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := bundle.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}

// SkillMarkdown returns the bundled kumo-cli skill, for surfaces that
// show or copy it.
func SkillMarkdown() string {
	data, _ := bundle.ReadFile("bundle/skills/kumo-cli/SKILL.md")
	return string(data)
}
