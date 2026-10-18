package storage

import (
	"github.com/AstraFoundry/KumoApp/internal/domain/preferences"
	"github.com/AstraFoundry/KumoApp/internal/infrastructure/paths"
)

// PreferenceStore keeps preferences in preferences.json.
type PreferenceStore struct{ paths paths.Paths }

// NewPreferenceStore stores preferences under p.
func NewPreferenceStore(p paths.Paths) *PreferenceStore { return &PreferenceStore{paths: p} }

// Load returns the stored preferences, or the defaults.
func (s *PreferenceStore) Load() (preferences.Preferences, error) {
	prefs := preferences.Default()
	_, err := ReadJSON(s.paths.Preferences(), &prefs)
	return prefs, err
}

// Save writes the preferences.
func (s *PreferenceStore) Save(prefs preferences.Preferences) error {
	return WriteJSON(s.paths.Preferences(), prefs)
}
