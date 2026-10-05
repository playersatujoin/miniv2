package sim

import (
	"context"
	"errors"
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
			return s
		}
		slog.Warn("discarding unreadable world save", "map", mp.ID, "err", err)
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
