package application

import "github.com/AstraFoundry/KumoApp/internal/domain/preferences"

// PreferenceService keeps app-level preferences.
type PreferenceService struct{ deps Dependencies }

// Get returns the preferences.
func (s *PreferenceService) Get() (preferences.Preferences, error) { return s.deps.Preferences.Load() }

// Save stores the preferences.
func (s *PreferenceService) Save(p preferences.Preferences) error { return s.deps.Preferences.Save(p) }

// Update changes the preferences with fn and stores them.
func (s *PreferenceService) Update(fn func(*preferences.Preferences)) (preferences.Preferences, error) {
	p, err := s.deps.Preferences.Load()
	if err != nil {
		return p, err
	}
	fn(&p)
	return p, s.deps.Preferences.Save(p)
}
