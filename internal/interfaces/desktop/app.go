package desktop

import (
	"context"
	"errors"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/egoist/mygo"

	"github.com/AstraFoundry/KumoApp/internal/application"
	"github.com/AstraFoundry/KumoApp/internal/domain/preferences"
	"github.com/AstraFoundry/KumoApp/internal/infrastructure/clilink"
	"github.com/AstraFoundry/KumoApp/internal/infrastructure/skills"
)

// Project facts the About page links to.
const (
	Team          = "AstraFoundry"
	Repository    = "AstraFoundry/KumoApp"
	RepositoryURL = "https://github.com/AstraFoundry/KumoApp"
	TeamURL       = "https://github.com/AstraFoundry"
	IssuesURL     = RepositoryURL + "/issues/new"
	License       = "GPL-3.0"
	LicenseURL    = RepositoryURL + "/blob/main/LICENSE"
	MihomoURL     = "https://github.com/MetaCubeX/mihomo"
)

// App is what the app knows about itself: setup, About, agents and updates.
type App struct {
	k *application.Kumo

	mu     sync.Mutex
	update *mygo.Update
}

// Info describes the app for About.
type Info struct {
	Name          string `json:"name"`
	Version       string `json:"version"`
	Platform      string `json:"platform"`
	Arch          string `json:"arch"`
	DataDir       string `json:"dataDir"`
	LogsDir       string `json:"logsDir"`
	Team          string `json:"team"`
	TeamURL       string `json:"teamURL"`
	Repository    string `json:"repository"`
	RepositoryURL string `json:"repositoryURL"`
	IssuesURL     string `json:"issuesURL"`
	License       string `json:"license"`
	LicenseURL    string `json:"licenseURL"`
	MihomoURL     string `json:"mihomoURL"`
	CanUpdate     bool   `json:"canUpdate"`
}

// Info returns the app's name, version and links.
func (a *App) Info() Info {
	version := mygo.App.Version()
	if version == "" {
		version = "dev"
	}
	return Info{
		Name: "Kumo", Version: version, Platform: runtime.GOOS, Arch: runtime.GOARCH,
		DataDir: a.k.Paths().Root, LogsDir: a.k.Paths().Logs(),
		Team: Team, TeamURL: TeamURL, Repository: Repository, RepositoryURL: RepositoryURL,
		IssuesURL: IssuesURL, License: License, LicenseURL: LicenseURL, MihomoURL: MihomoURL,
		CanUpdate: mygo.Updater.Enabled(),
	}
}

// SetupState is what the setup flow checks.
type SetupState struct {
	Completed bool               `json:"completed"`
	CoreFound bool               `json:"coreFound"`
	CorePath  string             `json:"corePath"`
	Profiles  int                `json:"profiles"`
	CLI       clilink.Status     `json:"cli"`
	Skills    skills.Status      `json:"skills"`
	Status    application.Status `json:"status"`
}

// Setup reports how far setup has come.
func (a *App) Setup() (SetupState, error) {
	prefs, err := a.k.Preferences.Get()
	if err != nil {
		return SetupState{}, err
	}
	settings, err := a.k.Settings.Get()
	if err != nil {
		return SetupState{}, err
	}
	profiles, err := a.k.Profiles.List()
	if err != nil {
		return SetupState{}, err
	}
	status, err := a.k.Lifecycle.Status()
	if err != nil {
		return SetupState{}, err
	}
	skillStatus, err := a.Skills("global")
	if err != nil {
		return SetupState{}, err
	}
	return SetupState{
		Completed: prefs.CompletedSetup, CoreFound: len(settings.Candidates) > 0, CorePath: settings.CorePath,
		Profiles: len(profiles), CLI: clilink.Check(), Skills: skillStatus, Status: status,
	}, nil
}

// CompleteSetup records that setup is done (or skipped).
func (a *App) CompleteSetup() error {
	_, err := a.k.Preferences.Update(func(p *preferences.Preferences) { p.CompletedSetup = true })
	return err
}

// RestartSetup shows setup again at the next window.
func (a *App) RestartSetup() error {
	_, err := a.k.Preferences.Update(func(p *preferences.Preferences) { p.CompletedSetup = false })
	return err
}

// InstallCLI links the kumo command into PATH.
func (a *App) InstallCLI() (clilink.Status, error) { return clilink.Install() }

// UninstallCLI removes the link.
func (a *App) UninstallCLI() (clilink.Status, error) { return clilink.Uninstall() }

func (a *App) installer() (*skills.Installer, error) {
	return skills.New(a.k.Paths().AgentSkillsState())
}

// Skills reports where the Kumo skill is installed, in "global" scope.
func (a *App) Skills(scope string) (skills.Status, error) {
	installer, err := a.installer()
	if err != nil {
		return skills.Status{}, err
	}
	return installer.Status(nil, skills.Scope(scope))
}

// InstallSkills installs the skill for the given agents.
func (a *App) InstallSkills(agents []string, force bool) (skills.Report, error) {
	installer, err := a.installer()
	if err != nil {
		return skills.Report{}, err
	}
	targets, err := parseTargets(agents)
	if err != nil {
		return skills.Report{}, err
	}
	return installer.Install(targets, skills.ScopeGlobal, false, force)
}

// UninstallSkills removes the skill from the given agents.
func (a *App) UninstallSkills(agents []string) (skills.Report, error) {
	installer, err := a.installer()
	if err != nil {
		return skills.Report{}, err
	}
	targets, err := parseTargets(agents)
	if err != nil {
		return skills.Report{}, err
	}
	return installer.Uninstall(targets, skills.ScopeGlobal, false)
}

func parseTargets(agents []string) ([]skills.Target, error) {
	if len(agents) == 0 {
		return nil, errors.New("choose at least one agent")
	}
	var targets []skills.Target
	for _, name := range agents {
		t, err := skills.ParseTarget(name)
		if err != nil {
			return nil, err
		}
		targets = append(targets, t)
	}
	return targets, nil
}

// SkillMarkdown is the bundled skill, for agents set up by hand.
func (a *App) SkillMarkdown() string { return skills.SkillMarkdown() }

// Guide returns a page's agent intents.
func (a *App) Guide(page string) application.Guide {
	guide, _ := application.GuideFor(page)
	if guide.Intents == nil {
		guide.Intents = []application.Intent{}
	}
	return guide
}

// CopyPrompt puts a page intent's agent prompt on the clipboard and
// returns it.
func (a *App) CopyPrompt(page, intent string) (string, error) {
	prompt, err := application.Prompt(page, intent)
	if err != nil {
		return "", err
	}
	mygo.Clipboard.WriteText(prompt)
	return prompt, nil
}

// CopyText puts text on the clipboard.
func (a *App) CopyText(text string) { mygo.Clipboard.WriteText(text) }

// OpenURL opens a web address in the browser.
func (a *App) OpenURL(url string) error {
	if !strings.HasPrefix(url, "https://") && !strings.HasPrefix(url, "http://") {
		return errors.New("only web addresses open in the browser")
	}
	return mygo.Shell.OpenExternal(url)
}

// Reveal shows one of Kumo's files in the file manager: "data", "logs",
// "config" or "coreLog".
func (a *App) Reveal(what string) error {
	p := a.k.Paths()
	targets := map[string]string{"data": p.Root, "logs": p.Logs(), "config": p.RuntimeConfig(), "coreLog": p.CoreLog()}
	path, ok := targets[what]
	if !ok {
		return errors.New("nothing to reveal for " + what)
	}
	mygo.Shell.ShowItemInFolder(path)
	return nil
}

// OpenAtLogin reports whether Kumo opens at login.
func (a *App) OpenAtLogin() bool { return mygo.App.OpenAtLogin() }

// SetOpenAtLogin turns opening at login on or off.
func (a *App) SetOpenAtLogin(open bool) error {
	if err := mygo.App.SetOpenAtLogin(open); err != nil {
		return err
	}
	_, err := a.k.Preferences.Update(func(p *preferences.Preferences) { p.LaunchAtLogin = open })
	return err
}

// UpdateInfo is a newer version, when there is one.
type UpdateInfo struct {
	Available bool      `json:"available"`
	Version   string    `json:"version,omitempty"`
	Notes     string    `json:"notes,omitempty"`
	Date      time.Time `json:"date"`
	Enabled   bool      `json:"enabled"`
}

// CheckUpdate looks for a newer version.
func (a *App) CheckUpdate(ctx context.Context) (UpdateInfo, error) {
	if !mygo.Updater.Enabled() {
		return UpdateInfo{}, nil
	}
	up, err := mygo.Updater.Check(ctx)
	if err != nil || up == nil {
		return UpdateInfo{Enabled: true}, err
	}
	a.mu.Lock()
	a.update = up
	a.mu.Unlock()
	return UpdateInfo{Available: true, Version: up.Version, Notes: up.Notes, Date: up.Date, Enabled: true}, nil
}

// InstallUpdate downloads the update CheckUpdate found and relaunches.
// Progress arrives as the updateProgress event.
func (a *App) InstallUpdate(ctx context.Context) error {
	a.mu.Lock()
	up := a.update
	a.mu.Unlock()
	if up == nil {
		return errors.New("check for updates first")
	}
	err := up.Install(context.WithoutCancel(ctx), func(done, total int64) {
		if total > 0 {
			UpdateProgress.Broadcast(float64(done) / float64(total))
		}
	})
	if err != nil {
		return err
	}
	mygo.App.Relaunch()
	return nil
}

// Quit quits Kumo, which releases the system proxy and stops the core.
func (a *App) Quit() { mygo.App.Quit() }

// appBundle is the .app a process runs from, whose icon Finder shows, or
// the path itself.
func appBundle(path string) string {
	for dir := path; dir != "/" && dir != "." && dir != ""; dir = filepath.Dir(dir) {
		if strings.HasSuffix(dir, ".app") {
			return dir
		}
	}
	return path
}
