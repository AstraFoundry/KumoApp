package storage

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/AstraFoundry/KumoApp/internal/domain/profile"
	"github.com/AstraFoundry/KumoApp/internal/infrastructure/paths"
)

// ErrNotFound is returned for an unknown ID.
var ErrNotFound = errors.New("not found")

// ProfileStore keeps each profile as <id>.yaml beside one metadata file.
type ProfileStore struct{ paths paths.Paths }

// NewProfileStore stores profiles under p.
func NewProfileStore(p paths.Paths) *ProfileStore { return &ProfileStore{paths: p} }

func (s *ProfileStore) file(id string) string {
	return filepath.Join(s.paths.Profiles(), id+".yaml")
}

func (s *ProfileStore) metadata() (map[string]profile.Profile, error) {
	meta := map[string]profile.Profile{}
	_, err := ReadJSON(s.paths.ProfileMetadata(), &meta)
	return meta, err
}

// List returns the default profile first, then the rest by name.
func (s *ProfileStore) List() ([]profile.Profile, error) {
	meta, err := s.metadata()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(s.paths.Profiles())
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	list := []profile.Profile{}
	hasDefault := false
	for _, entry := range entries {
		ext := filepath.Ext(entry.Name())
		if entry.IsDir() || (ext != ".yaml" && ext != ".yml") || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		id := strings.TrimSuffix(entry.Name(), ext)
		p := s.describe(id, meta)
		hasDefault = hasDefault || id == profile.DefaultID
		list = append(list, p)
	}
	if !hasDefault {
		list = append(list, profile.Default())
	}
	sort.SliceStable(list, func(i, j int) bool {
		if list[i].ID == profile.DefaultID || list[j].ID == profile.DefaultID {
			return list[i].ID == profile.DefaultID
		}
		return strings.ToLower(list[i].Name) < strings.ToLower(list[j].Name)
	})
	return list, nil
}

func (s *ProfileStore) describe(id string, meta map[string]profile.Profile) profile.Profile {
	p, ok := meta[id]
	if !ok {
		p = profile.Profile{ID: id, Kind: profile.KindLocal, AutoUpdate: true, Name: strings.ReplaceAll(id, "-", " ")}
		if id == profile.DefaultID {
			p = profile.Default()
		}
		if info, err := os.Stat(s.file(id)); err == nil {
			at := info.ModTime()
			p.UpdatedAt = &at
		}
	}
	p.ID = id
	return p
}

// Get returns one profile's metadata.
func (s *ProfileStore) Get(id string) (profile.Profile, error) {
	if err := profile.ValidateID(id); err != nil {
		return profile.Profile{}, err
	}
	meta, err := s.metadata()
	if err != nil {
		return profile.Profile{}, err
	}
	if _, err := os.Stat(s.file(id)); err != nil {
		if id == profile.DefaultID {
			return profile.Default(), nil
		}
		return profile.Profile{}, fmt.Errorf("profile %q: %w", id, ErrNotFound)
	}
	return s.describe(id, meta), nil
}

// Content returns the YAML; the default profile without a file is the
// generated direct one.
func (s *ProfileStore) Content(id string) (string, error) {
	if err := profile.ValidateID(id); err != nil {
		return "", err
	}
	data, err := os.ReadFile(s.file(id))
	if errors.Is(err, fs.ErrNotExist) {
		if id == profile.DefaultID {
			return profile.DefaultYAML, nil
		}
		return "", fmt.Errorf("profile %q: %w", id, ErrNotFound)
	}
	return string(data), err
}

// Save writes the profile under the shared lock.
func (s *ProfileStore) Save(p profile.Profile, yaml *string) error {
	if err := profile.ValidateID(p.ID); err != nil {
		return err
	}
	lock, err := Lock(s.paths.StateLock())
	if err != nil {
		return err
	}
	defer lock.Unlock()
	if yaml != nil {
		if err := WriteFile(s.file(p.ID), []byte(*yaml)); err != nil {
			return err
		}
	}
	meta, err := s.metadata()
	if err != nil {
		return err
	}
	if p.UpdatedAt == nil {
		now := time.Now()
		p.UpdatedAt = &now
	}
	meta[p.ID] = p
	return WriteJSON(s.paths.ProfileMetadata(), meta)
}

// Delete removes a profile. The default profile cannot be deleted; when the
// current one goes, the first remaining one becomes current.
func (s *ProfileStore) Delete(id string) error {
	if id == profile.DefaultID {
		return errors.New("the default profile cannot be deleted")
	}
	if err := profile.ValidateID(id); err != nil {
		return err
	}
	current, err := s.CurrentID()
	if err != nil {
		return err
	}
	lock, err := Lock(s.paths.StateLock())
	if err != nil {
		return err
	}
	if err := os.Remove(s.file(id)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		lock.Unlock()
		return err
	}
	meta, err := s.metadata()
	if err == nil {
		delete(meta, id)
		err = WriteJSON(s.paths.ProfileMetadata(), meta)
	}
	lock.Unlock()
	if err != nil || current != id {
		return err
	}
	list, err := s.List()
	if err != nil {
		return err
	}
	next := profile.DefaultID
	for _, p := range list {
		if p.ID != profile.DefaultID {
			next = p.ID
			break
		}
	}
	return s.SetCurrent(next)
}

// CurrentID reads current.txt; nothing there is the default profile.
func (s *ProfileStore) CurrentID() (string, error) {
	data, err := os.ReadFile(s.paths.CurrentProfile())
	if errors.Is(err, fs.ErrNotExist) {
		return profile.DefaultID, nil
	}
	if err != nil {
		return "", err
	}
	id := strings.TrimSpace(string(data))
	if id == "" || profile.ValidateID(id) != nil {
		return profile.DefaultID, nil
	}
	if _, err := os.Stat(s.file(id)); err != nil && id != profile.DefaultID {
		return profile.DefaultID, nil
	}
	return id, nil
}

// SetCurrent makes a profile current.
func (s *ProfileStore) SetCurrent(id string) error {
	if _, err := s.Get(id); err != nil {
		return err
	}
	return WriteFile(s.paths.CurrentProfile(), []byte(id))
}
