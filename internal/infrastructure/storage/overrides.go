package storage

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	"github.com/AstraFoundry/KumoApp/internal/domain/override"
	"github.com/AstraFoundry/KumoApp/internal/infrastructure/paths"
)

// OverrideStore keeps overrides in overrides.json, in order, with their
// content in files/<id>.yaml or .js.
type OverrideStore struct{ paths paths.Paths }

// NewOverrideStore stores overrides under p.
func NewOverrideStore(p paths.Paths) *OverrideStore { return &OverrideStore{paths: p} }

func (s *OverrideStore) file(o override.Override) string {
	ext := ".yaml"
	if o.Format == override.FormatJavaScript {
		ext = ".js"
	}
	return filepath.Join(s.paths.OverrideFiles(), o.ID+ext)
}

// List returns overrides in the order they apply.
func (s *OverrideStore) List() ([]override.Override, error) {
	items := []override.Override{}
	_, err := ReadJSON(s.paths.OverrideMetadata(), &items)
	return items, err
}

func (s *OverrideStore) find(id string) (override.Override, error) {
	items, err := s.List()
	if err != nil {
		return override.Override{}, err
	}
	for _, item := range items {
		if item.ID == id {
			return item, nil
		}
	}
	return override.Override{}, fmt.Errorf("override %q: %w", id, ErrNotFound)
}

// Content reads an override's file.
func (s *OverrideStore) Content(id string) (string, error) {
	item, err := s.find(id)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(s.file(item))
	if errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}
	return string(data), err
}

// Save adds the override at the end, or replaces it in place.
func (s *OverrideStore) Save(o override.Override, content *string) error {
	if o.ID == "" || filepath.Base(o.ID) != o.ID {
		return fmt.Errorf("invalid override id %q", o.ID)
	}
	lock, err := Lock(s.paths.StateLock())
	if err != nil {
		return err
	}
	defer lock.Unlock()
	items, err := s.List()
	if err != nil {
		return err
	}
	index := slices.IndexFunc(items, func(item override.Override) bool { return item.ID == o.ID })
	if index >= 0 && items[index].Format != o.Format {
		_ = os.Rename(s.file(items[index]), s.file(o))
	}
	if content != nil {
		if err := WriteFile(s.file(o), []byte(*content)); err != nil {
			return err
		}
	}
	if index >= 0 {
		items[index] = o
	} else {
		items = append(items, o)
	}
	return WriteJSON(s.paths.OverrideMetadata(), items)
}

// Delete removes an override and its file.
func (s *OverrideStore) Delete(id string) error {
	lock, err := Lock(s.paths.StateLock())
	if err != nil {
		return err
	}
	defer lock.Unlock()
	items, err := s.List()
	if err != nil {
		return err
	}
	index := slices.IndexFunc(items, func(item override.Override) bool { return item.ID == id })
	if index < 0 {
		return nil
	}
	_ = os.Remove(s.file(items[index]))
	items = slices.Delete(items, index, index+1)
	return WriteJSON(s.paths.OverrideMetadata(), items)
}

// Reorder puts the listed overrides first, in that order.
func (s *OverrideStore) Reorder(ids []string) error {
	lock, err := Lock(s.paths.StateLock())
	if err != nil {
		return err
	}
	defer lock.Unlock()
	items, err := s.List()
	if err != nil {
		return err
	}
	ordered := make([]override.Override, 0, len(items))
	for _, id := range ids {
		for _, item := range items {
			if item.ID == id {
				ordered = append(ordered, item)
			}
		}
	}
	for _, item := range items {
		if !slices.Contains(ids, item.ID) {
			ordered = append(ordered, item)
		}
	}
	return WriteJSON(s.paths.OverrideMetadata(), ordered)
}
