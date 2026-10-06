package sim

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"os"
	"path/filepath"
	"sync"
	"time"

	"miniv2/backend/internal/world"
)

const autosaveEvery = time.Minute

// Manager runs one Sim per map in the background and persists them.
type Manager struct {
	dir string

	mu   sync.Mutex
	sims map[string]*runner

	saveMu sync.Mutex // serialises saves with removals so deleted worlds stay deleted
	stop   chan struct{}
	done   chan struct{}
	once   sync.Once
}

type runner struct {
	sim    *Sim
	cancel context.CancelFunc
	done   chan struct{}
}

// NewManager stores world saves in dir and autosaves every minute until Close.
func NewManager(dir string) (*Manager, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	m := &Manager{
		dir:  dir,
		sims: map[string]*runner{},
		stop: make(chan struct{}),
		done: make(chan struct{}),
	}
	go m.autosave()
	return m, nil
}

func (m *Manager) path(id string) string { return filepath.Join(m.dir, id+".json.gz") }

// legacyPath is where saves from before the civilisation update lived.
func (m *Manager) legacyPath(id string) string { return filepath.Join(m.dir, id+".json") }

// Start brings a map's world to life, resuming its save if there is one.
func (m *Manager) Start(mp *world.Map) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.sims[mp.ID]; ok {
		return
	}
	s := m.load(mp)
	ctx, cancel := context.WithCancel(context.Background())
	r := &runner{sim: s, cancel: cancel, done: make(chan struct{})}
	go func() {
		defer close(r.done)
		s.Run(ctx)
	}()
	m.sims[mp.ID] = r
}

func (m *Manager) load(mp *world.Map) *Sim {
	if err := os.Remove(m.legacyPath(mp.ID)); err == nil {
		slog.Info("discarded pre-civilisation world save; starting again from Adam and Hawa", "map", mp.ID)
	}
	data, err := os.ReadFile(m.path(mp.ID))
	if err == nil {
		s, err := Restore(mp, data)
		if err == nil {
			if s.loadedVersion < stateVersion {
				s.migrationOriginal = data
				if berr := m.archiveMigration(mp.ID, s); berr != nil {
					slog.Error("cannot archive pre-migration save", "map", mp.ID, "err", berr)
				} else {
					slog.Info("world migrated with original save retained", "map", mp.ID, "from", s.loadedVersion, "to", stateVersion)
				}
			}
			return s
		}
		// Keep the old world on disk instead of overwriting it.
		backup := m.path(mp.ID) + fmt.Sprintf(".bak-%d", time.Now().Unix())
		if rerr := os.Rename(m.path(mp.ID), backup); rerr != nil {
			backup = ""
		}
		slog.Warn("world save can't be loaded; starting again from Adam and Hawa", "map", mp.ID, "err", err, "backup", backup)
		s = New(mp, rand.Uint64())
		s.event("milestone", "Dunia lama tersimpan dalam format lama dan sudah dicadangkan; dunia dimulai ulang dari Adam & Hawa.", 0)
		return s
	} else if !errors.Is(err, os.ErrNotExist) {
		slog.Warn("reading world save", "map", mp.ID, "err", err)
	}
	return New(mp, rand.Uint64())
}

func (m *Manager) Get(id string) (*Sim, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.sims[id]
	if !ok {
		return nil, false
	}
	return r.sim, true
}

// Update applies an edited map to its running world.
func (m *Manager) Update(mp *world.Map) {
	if s, ok := m.Get(mp.ID); ok {
		s.UpdateMap(mp)
	}
}

// Remove stops a map's world and deletes its save.
func (m *Manager) Remove(id string) {
	m.saveMu.Lock()
	defer m.saveMu.Unlock()
	m.mu.Lock()
	r, ok := m.sims[id]
	delete(m.sims, id)
	m.mu.Unlock()
	if ok {
		r.cancel()
		<-r.done
	}
	for _, p := range []string{m.path(id), m.legacyPath(id)} {
		if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
			slog.Warn("removing world save", "map", id, "err", err)
		}
	}
}

// SaveAll writes every running world to disk.
func (m *Manager) SaveAll() error {
	m.saveMu.Lock()
	defer m.saveMu.Unlock()
	m.mu.Lock()
	sims := make(map[string]*Sim, len(m.sims))
	for id, r := range m.sims {
		sims[id] = r.sim
	}
	m.mu.Unlock()

	var errs []error
	for id, s := range sims {
		if err := m.archiveMigration(id, s); err != nil {
			errs = append(errs, err)
			continue // never overwrite the original when its backup failed
		}
		data, err := s.MarshalState()
		if err == nil {
			err = writeFileAtomic(m.path(id), data)
		}
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (m *Manager) archiveMigration(id string, s *Sim) error {
	if len(s.migrationOriginal) == 0 {
		return nil
	}
	backup := m.path(id) + fmt.Sprintf(".v%d-before-v%d.bak", s.loadedVersion, stateVersion)
	f, err := os.OpenFile(backup, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if errors.Is(err, os.ErrExist) {
		info, statErr := os.Stat(backup)
		if statErr != nil {
			return statErr
		}
		if !info.Mode().IsRegular() || info.Size() == 0 {
			return fmt.Errorf("migration backup is not a nonempty file: %s", backup)
		}
		s.migrationOriginal = nil
		return nil
	}
	if err != nil {
		return err
	}
	_, err = f.Write(s.migrationOriginal)
	if err == nil {
		err = f.Sync()
	}
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		os.Remove(backup)
		return err
	}
	s.migrationOriginal = nil
	return nil
}

func (m *Manager) autosave() {
	defer close(m.done)
	t := time.NewTicker(autosaveEvery)
	defer t.Stop()
	for {
		select {
		case <-m.stop:
			return
		case <-t.C:
			if err := m.SaveAll(); err != nil {
				slog.Error("autosave worlds", "err", err)
			}
		}
	}
}

// Close stops every world and saves it.
func (m *Manager) Close() error {
	m.once.Do(func() { close(m.stop) })
	<-m.done
	m.mu.Lock()
	for _, r := range m.sims {
		r.cancel()
		<-r.done
	}
	m.mu.Unlock()
	return m.SaveAll()
}

func writeFileAtomic(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*.tmp")
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
	return os.Rename(tmp.Name(), path)
}
