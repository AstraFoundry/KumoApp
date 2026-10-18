package application

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/AstraFoundry/KumoApp/internal/domain/core"
	"github.com/AstraFoundry/KumoApp/internal/domain/profile"
	"github.com/AstraFoundry/KumoApp/internal/domain/runtimeconfig"
)

// ProfileService imports, refreshes and switches profiles.
type ProfileService struct {
	deps      Dependencies
	events    *Bus
	lifecycle *Lifecycle
}

// List returns every profile, the current one marked.
func (s *ProfileService) List() ([]profile.Summary, error) {
	list, err := s.deps.Profiles.List()
	if err != nil {
		return nil, err
	}
	current, err := s.deps.Profiles.CurrentID()
	if err != nil {
		return nil, err
	}
	out := make([]profile.Summary, len(list))
	for i, p := range list {
		out[i] = p.Summarize(current)
	}
	return out, nil
}

// Get returns one profile.
func (s *ProfileService) Get(id string) (profile.Summary, error) {
	p, err := s.deps.Profiles.Get(id)
	if err != nil {
		return profile.Summary{}, err
	}
	current, _ := s.deps.Profiles.CurrentID()
	return p.Summarize(current), nil
}

// Content returns a profile's YAML.
func (s *ProfileService) Content(id string) (string, error) { return s.deps.Profiles.Content(id) }

func (s *ProfileService) changed(ctx context.Context, reload bool) error {
	s.events.Publish(ProfilesChanged{})
	if reload {
		return s.lifecycle.Reload(ctx)
	}
	_, _ = s.lifecycle.Refresh()
	return nil
}

// Select makes a profile current and has a running core switch to it.
func (s *ProfileService) Select(ctx context.Context, id string) error {
	if err := s.deps.Profiles.SetCurrent(id); err != nil {
		return err
	}
	return s.changed(ctx, true)
}

// ImportRequest describes a profile to import from a URL, a file or text.
type ImportRequest struct {
	URL  string `json:"url,omitempty"`
	Path string `json:"path,omitempty"`
	YAML string `json:"yaml,omitempty"`
	Name string `json:"name,omitempty"`
	// UseProxy fetches the URL through the running core.
	UseProxy bool `json:"useProxy,omitempty"`
	// Activate makes the imported profile current.
	Activate bool `json:"activate,omitempty"`
}

// Import adds a profile and, when asked, switches to it.
func (s *ProfileService) Import(ctx context.Context, req ImportRequest) (profile.Summary, error) {
	var p profile.Profile
	var content string
	switch {
	case req.URL != "":
		u, err := url.Parse(strings.TrimSpace(req.URL))
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
			return profile.Summary{}, fmt.Errorf("%q is not an http(s) URL", req.URL)
		}
		doc, err := s.fetch(u.String(), req.UseProxy)
		if err != nil {
			return profile.Summary{}, err
		}
		name := firstNonEmpty(req.Name, doc.Name)
		p = profile.Profile{ID: profile.NewID(name), Name: name, Kind: profile.KindRemote, RemoteURL: u.String(),
			HomeURL: doc.HomeURL, AutoUpdate: true, UseProxy: req.UseProxy, UpdateIntervalSeconds: doc.UpdateIntervalSeconds, Usage: doc.Usage}
		content = doc.YAML
	case req.Path != "":
		data, err := os.ReadFile(req.Path)
		if err != nil {
			return profile.Summary{}, err
		}
		name := firstNonEmpty(req.Name, strings.TrimSuffix(filepath.Base(req.Path), filepath.Ext(req.Path)))
		p = profile.Profile{ID: profile.NewID(name), Name: name, Kind: profile.KindLocal, AutoUpdate: true}
		content = string(data)
	case strings.TrimSpace(req.YAML) != "":
		name := firstNonEmpty(req.Name, "Pasted profile")
		p = profile.Profile{ID: profile.NewID(name), Name: name, Kind: profile.KindLocal, AutoUpdate: true}
		content = req.YAML
	default:
		return profile.Summary{}, errors.New("give a URL, a file or YAML to import")
	}
	if err := runtimeconfig.ValidateProfile(content); err != nil {
		return profile.Summary{}, notAProfile(err)
	}
	if err := s.deps.Profiles.Save(p, &content); err != nil {
		return profile.Summary{}, err
	}
	if req.Activate {
		if err := s.deps.Profiles.SetCurrent(p.ID); err != nil {
			return profile.Summary{}, err
		}
	}
	reloadErr := s.changed(ctx, req.Activate)
	summary, err := s.Get(p.ID)
	return summary, errors.Join(reloadErr, err)
}

func notAProfile(err error) error {
	return fmt.Errorf("this is not a Mihomo profile (%v); subscription links that return base64 node lists need conversion, for example through Sub-Store", err)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return "Profile"
}

func (s *ProfileService) fetch(rawURL string, useProxy bool) (profile.Document, error) {
	port := 0
	if useProxy {
		rt, err := s.deps.Runtime.Load()
		if err != nil {
			return profile.Document{}, err
		}
		if rt.State != core.StateRunning {
			return profile.Document{}, errors.New("start Kumo before updating this profile through the proxy")
		}
		port = rt.Settings.MixedPort
	}
	return s.deps.Fetcher.Fetch(rawURL, port)
}

// Refresh downloads a remote profile again, keeping its ID and settings.
func (s *ProfileService) Refresh(ctx context.Context, id string) (profile.Summary, error) {
	p, err := s.deps.Profiles.Get(id)
	if err != nil {
		return profile.Summary{}, err
	}
	if p.Kind != profile.KindRemote || p.RemoteURL == "" {
		return profile.Summary{}, fmt.Errorf("profile %q has no subscription URL", p.Name)
	}
	doc, err := s.fetch(p.RemoteURL, p.UseProxy)
	if err != nil {
		return profile.Summary{}, err
	}
	if err := runtimeconfig.ValidateProfile(doc.YAML); err != nil {
		return profile.Summary{}, notAProfile(err)
	}
	p.HomeURL = firstOr(doc.HomeURL, p.HomeURL)
	if doc.UpdateIntervalSeconds > 0 {
		p.UpdateIntervalSeconds = doc.UpdateIntervalSeconds
	}
	if doc.Usage != nil {
		p.Usage = doc.Usage
	}
	now := s.deps.Now()
	p.UpdatedAt = &now
	if err := s.deps.Profiles.Save(p, &doc.YAML); err != nil {
		return profile.Summary{}, err
	}
	current, _ := s.deps.Profiles.CurrentID()
	if err := s.changed(ctx, current == id); err != nil {
		return profile.Summary{}, err
	}
	return s.Get(id)
}

func firstOr(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

// RefreshDue refreshes every remote profile whose interval has passed.
func (s *ProfileService) RefreshDue(ctx context.Context) ([]profile.Summary, error) {
	list, err := s.deps.Profiles.List()
	if err != nil {
		return nil, err
	}
	var refreshed []profile.Summary
	var errs []error
	for _, p := range list {
		if !p.IsDue(s.deps.Now()) {
			continue
		}
		summary, err := s.Refresh(ctx, p.ID)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", p.Name, err))
			continue
		}
		refreshed = append(refreshed, summary)
	}
	return refreshed, errors.Join(errs...)
}

// UpdateRequest edits a profile.
type UpdateRequest struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	RemoteURL  string  `json:"remoteURL"`
	AutoUpdate bool    `json:"autoUpdate"`
	UseProxy   bool    `json:"useProxy"`
	YAML       *string `json:"yaml,omitempty"`
}

// Update edits a profile's metadata and, optionally, its YAML.
func (s *ProfileService) Update(ctx context.Context, req UpdateRequest) (profile.Summary, error) {
	p, err := s.deps.Profiles.Get(req.ID)
	if err != nil {
		return profile.Summary{}, err
	}
	if req.YAML != nil {
		if err := runtimeconfig.ValidateProfile(*req.YAML); err != nil {
			return profile.Summary{}, err
		}
	}
	p.Name = firstOr(strings.TrimSpace(req.Name), p.Name)
	p.AutoUpdate, p.UseProxy = req.AutoUpdate, req.UseProxy
	p.RemoteURL = strings.TrimSpace(req.RemoteURL)
	switch {
	case p.RemoteURL != "":
		p.Kind = profile.KindRemote
	case p.Kind == profile.KindRemote:
		p.Kind = profile.KindLocal
	}
	if req.ID == profile.DefaultID && p.Kind == profile.KindInline && req.YAML != nil {
		p.Kind = profile.KindLocal
	}
	if req.YAML != nil {
		now := s.deps.Now()
		p.UpdatedAt = &now
	}
	if err := s.deps.Profiles.Save(p, req.YAML); err != nil {
		return profile.Summary{}, err
	}
	current, _ := s.deps.Profiles.CurrentID()
	if err := s.changed(ctx, current == req.ID && req.YAML != nil); err != nil {
		return profile.Summary{}, err
	}
	return s.Get(req.ID)
}

// Delete removes a profile; when it was current, the core switches to the
// next one.
func (s *ProfileService) Delete(ctx context.Context, id string) error {
	current, _ := s.deps.Profiles.CurrentID()
	if err := s.deps.Profiles.Delete(id); err != nil {
		return err
	}
	return s.changed(ctx, current == id)
}
