// Package store persists maps as one JSON file each, cached in memory.
package store

import (
	"cmp"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"

	"miniv2/backend/internal/world"
)

var ErrNotFound = errors.New("map not found")

var idPattern = regexp.MustCompile(`^[a-f0-9]{16}$`)

// Stored maps are treated as immutable: Save swaps in a new pointer rather
// than mutating, so callers can read a *world.Map without holding the lock.
type FileStore struct {
	dir  string
	mu   sync.RWMutex
	maps map[string]*world.Map
}

func Open(dir string) (*FileStore, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	s := &FileStore{dir: dir, maps: map[string]*world.Map{}}

	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		var m world.Map
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("%s: %w", f, err)
		}
		if !idPattern.MatchString(m.ID) {
			return nil, fmt.Errorf("%s: invalid map id %q", f, m.ID)
		}
		s.maps[m.ID] = &m
	}
	return s, nil
}

func NewID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *FileStore) List() []world.Summary {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]world.Summary, 0, len(s.maps))
	for _, m := range s.maps {
		out = append(out, m.Summary())
	}
	slices.SortFunc(out, func(a, b world.Summary) int {
		return cmp.Or(b.UpdatedAt.Compare(a.UpdatedAt), strings.Compare(a.ID, b.ID))
	})
	return out
}

func (s *FileStore) Get(id string) (*world.Map, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.maps[id]
	if !ok {
		return nil, ErrNotFound
	}
	return m, nil
}

// Save writes the map to disk atomically, then publishes it in memory.
func (s *FileStore) Save(m *world.Map) error {
	if !idPattern.MatchString(m.ID) {
		return fmt.Errorf("invalid map id %q", m.ID)
	}
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	tmp, err := os.CreateTemp(s.dir, m.ID+".*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmp.Name(), s.path(m.ID)); err != nil {
		return err
	}
	s.maps[m.ID] = m
	return nil
}

func (s *FileStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.maps[id]; !ok {
		return ErrNotFound
	}
	if err := os.Remove(s.path(id)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	delete(s.maps, id)
	return nil
}

func (s *FileStore) path(id string) string { return filepath.Join(s.dir, id+".json") }
