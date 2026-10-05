package world

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	MinSize         = 16
	MaxSize         = 256
	DefaultTileSize = 32
	MaxNameLen      = 64
)

type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Layers holds row-major tile IDs: index = y*width + x.
type Layers struct {
	Ground  []int `json:"ground"`
	Objects []int `json:"objects"`
}

type Map struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Width     int       `json:"width"`
	Height    int       `json:"height"`
	TileSize  int       `json:"tileSize"`
	Seed      uint32    `json:"seed"`
	Spawn     Point     `json:"spawn"`
	Layers    Layers    `json:"layers"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Summary is the lightweight form used in map listings.
type Summary struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Width     int       `json:"width"`
	Height    int       `json:"height"`
	Seed      uint32    `json:"seed"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (m *Map) Summary() Summary {
	return Summary{
		ID:        m.ID,
		Name:      m.Name,
		Width:     m.Width,
		Height:    m.Height,
		Seed:      m.Seed,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func (m *Map) inBounds(x, y int) bool { return x >= 0 && y >= 0 && x < m.Width && y < m.Height }

// Walkable reports whether a player can stand on the tile at (x, y).
func (m *Map) Walkable(x, y int) bool {
	if !m.inBounds(x, y) {
		return false
	}
	i := y*m.Width + x
	return !groundSolid(m.Layers.Ground[i]) && !objectSolid(m.Layers.Objects[i])
}

// ValidateSize checks map dimensions against the allowed range.
func ValidateSize(width, height int) error {
	if width < MinSize || width > MaxSize || height < MinSize || height > MaxSize {
		return fmt.Errorf("width and height must be between %d and %d", MinSize, MaxSize)
	}
	return nil
}

// NormalizeName trims a map name and checks its length.
func NormalizeName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.New("name is required")
	}
	if len([]rune(name)) > MaxNameLen {
		return "", fmt.Errorf("name must be at most %d characters", MaxNameLen)
	}
	return name, nil
}

// Validate checks that the map is internally consistent.
func (m *Map) Validate() error {
	if _, err := NormalizeName(m.Name); err != nil {
		return err
	}
	if err := ValidateSize(m.Width, m.Height); err != nil {
		return err
	}
	n := m.Width * m.Height
	if len(m.Layers.Ground) != n || len(m.Layers.Objects) != n {
		return fmt.Errorf("each layer must have exactly %d tiles", n)
	}
	for i, id := range m.Layers.Ground {
		if !validID(GroundTiles, id) {
			return fmt.Errorf("invalid ground tile %d at index %d", id, i)
		}
	}
	for i, id := range m.Layers.Objects {
		if !validID(ObjectTiles, id) {
			return fmt.Errorf("invalid object tile %d at index %d", id, i)
		}
	}
	if !m.Walkable(m.Spawn.X, m.Spawn.Y) {
		return errors.New("spawn must be on a walkable tile")
	}
	return nil
}
