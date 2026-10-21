package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/AstraFoundry/KumoApp/internal/domain/override"
)

// OverrideService manages the overrides applied on top of every profile.
type OverrideService struct {
	deps      Dependencies
	lifecycle *Lifecycle
}

// OverrideDetail is an override with its content.
type OverrideDetail struct {
	override.Override
	Content string `json:"content"`
}

// List returns overrides in the order they apply.
func (s *OverrideService) List() ([]override.Override, error) { return s.deps.Overrides.List() }

// Get returns an override with its content.
func (s *OverrideService) Get(id string) (OverrideDetail, error) {
	items, err := s.deps.Overrides.List()
	if err != nil {
		return OverrideDetail{}, err
	}
	for _, item := range items {
		if item.ID == id {
			content, err := s.deps.Overrides.Content(id)
			return OverrideDetail{Override: item, Content: content}, err
		}
	}
	return OverrideDetail{}, fmt.Errorf("override %q not found", id)
}

// SaveRequest creates or edits an override. With a URL and no content,
// the content is downloaded.
type SaveRequest struct {
	ID       string          `json:"id,omitempty"`
	Name     string          `json:"name"`
	Format   override.Format `json:"format"`
	Content  string          `json:"content"`
	URL      string          `json:"url,omitempty"`
	Disabled bool            `json:"disabled,omitempty"`
}

// Save validates and stores an override, then reloads a running core.
func (s *OverrideService) Save(ctx context.Context, req SaveRequest) (override.Override, error) {
	if req.Format == "" {
		req.Format = override.FormatYAML
	}
	item := override.Override{ID: req.ID, Name: strings.TrimSpace(req.Name), Kind: override.KindLocal, Format: req.Format, Disabled: req.Disabled, UpdatedAt: s.deps.Now()}
	content := req.Content
	if req.URL != "" {
		u, err := url.Parse(req.URL)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
			return override.Override{}, fmt.Errorf("%q is not an http(s) URL", req.URL)
		}
		item.Kind, item.RemoteURL = override.KindRemote, u.String()
		if item.Name == "" {
			item.Name = strings.TrimSuffix(path.Base(u.Path), path.Ext(u.Path))
		}
		if strings.TrimSpace(content) == "" {
			body, _, err := s.deps.Fetcher.Get(u.String(), 0)
			if err != nil {
				return override.Override{}, err
			}
			content = string(body)
		}
	}
	if item.Name == "" {
		item.Name = "Override"
	}
	if err := override.Validate(item.Format, content); err != nil {
		return override.Override{}, err
	}
	if item.ID == "" {
		item.ID = newID()
	} else if existing, err := s.Get(item.ID); err == nil {
		item.Global = existing.Global
		item.Fingerprint = existing.Fingerprint
	}
	if err := s.deps.Overrides.Save(item, &content); err != nil {
		return override.Override{}, err
	}
	return item, s.lifecycle.Reload(ctx)
}

// SetDisabled turns an override off or on without deleting it.
func (s *OverrideService) SetDisabled(ctx context.Context, id string, disabled bool) error {
	detail, err := s.Get(id)
	if err != nil {
		return err
	}
	detail.Override.Disabled = disabled
	if err := s.deps.Overrides.Save(detail.Override, nil); err != nil {
		return err
	}
	return s.lifecycle.Reload(ctx)
}

// Delete removes an override.
func (s *OverrideService) Delete(ctx context.Context, id string) error {
	if err := s.deps.Overrides.Delete(id); err != nil {
		return err
	}
	return s.lifecycle.Reload(ctx)
}

// Reorder changes the order overrides apply in.
func (s *OverrideService) Reorder(ctx context.Context, ids []string) error {
	if err := s.deps.Overrides.Reorder(ids); err != nil {
		return err
	}
	return s.lifecycle.Reload(ctx)
}

func newID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	h := hex.EncodeToString(b)
	return strings.ToUpper(h[:8] + "-" + h[8:12] + "-" + h[12:16] + "-" + h[16:20] + "-" + h[20:])
}
