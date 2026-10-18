// Package preferences models the app-level choices of the user, apart from
// how the core runs.
package preferences

import "encoding/json"

// Theme is the appearance the app shows.
type Theme string

const (
	ThemeSystem Theme = "system"
	ThemeLight  Theme = "light"
	ThemeDark   Theme = "dark"
)

// Preferences are stored in preferences.json.
type Preferences struct {
	LaunchAtLogin bool `json:"launchAtLogin"`
	// StartCoreOnLaunch starts Mihomo when the app opens.
	StartCoreOnLaunch bool `json:"startCoreOnLaunch"`
	// EnableSystemProxyOnStart turns the system proxy on once the core runs.
	EnableSystemProxyOnStart bool `json:"enableSystemProxyOnStart"`
	// ReclaimSystemProxy puts Kumo's proxy back when another app replaces it.
	ReclaimSystemProxy bool `json:"reclaimSystemProxy"`
	HideMenuBarIcon    bool `json:"hideMenuBarIcon"`
	// QuitOnLastWindowClose quits instead of staying in the menu bar.
	QuitOnLastWindowClose bool `json:"quitOnLastWindowClose"`
	// Language is a BCP 47 tag, or "" to follow the system.
	Language string `json:"appLanguage,omitempty"`
	Theme    Theme  `json:"theme"`
	// CompletedSetup is set once the user finished or skipped setup.
	CompletedSetup bool `json:"hasCompletedOnboarding"`
	// AutoCheckUpdates looks for app updates in the background.
	AutoCheckUpdates bool `json:"autoCheckUpdates"`
}

// Default is what a fresh install uses.
func Default() Preferences {
	return Preferences{Theme: ThemeSystem, AutoCheckUpdates: true, EnableSystemProxyOnStart: true}
}

// UnmarshalJSON keeps defaults for fields older files leave out.
func (p *Preferences) UnmarshalJSON(data []byte) error {
	type plain Preferences
	value := plain(Default())
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	if value.Theme == "" {
		value.Theme = ThemeSystem
	}
	*p = Preferences(value)
	return nil
}

// Store persists preferences.
type Store interface {
	Load() (Preferences, error)
	Save(Preferences) error
}
