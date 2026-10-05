package world

import (
	"slices"
	"strings"
	"testing"
)

func TestGenerateDeterministic(t *testing.T) {
	opts := GenOptions{Name: "a", Width: 64, Height: 48, Seed: 42}
	a, b := Generate(opts), Generate(opts)
	if !slices.Equal(a.Layers.Ground, b.Layers.Ground) || !slices.Equal(a.Layers.Objects, b.Layers.Objects) {
		t.Fatal("same seed produced different maps")
	}
	if a.Spawn != b.Spawn {
		t.Fatalf("spawn differs: %v vs %v", a.Spawn, b.Spawn)
	}

	c := Generate(GenOptions{Name: "a", Width: 64, Height: 48, Seed: 43})
	if slices.Equal(a.Layers.Ground, c.Layers.Ground) {
		t.Fatal("different seeds produced identical ground")
	}
}

func TestGenerateValid(t *testing.T) {
	for _, size := range [][2]int{{MinSize, MinSize}, {96, 64}, {MaxSize, MaxSize}} {
		for seed := range uint32(5) {
			m := Generate(GenOptions{Name: "test", Width: size[0], Height: size[1], Seed: seed})
			if err := m.Validate(); err != nil {
				t.Fatalf("%dx%d seed %d: %v", size[0], size[1], seed, err)
			}
		}
	}
}

func TestGenerateHasVariety(t *testing.T) {
	m := Generate(GenOptions{Name: "test", Width: 128, Height: 128, Seed: 1337})
	counts := map[int]int{}
	for _, g := range m.Layers.Ground {
		counts[g]++
	}
	for _, id := range []int{DeepWater, Water, Sand, Grass, ForestFloor, Dirt, StoneFloor, Mountain} {
		if counts[id] == 0 {
			t.Errorf("no %s tiles generated", GroundTiles[id].Key)
		}
	}

	if testing.Verbose() {
		glyph := []byte("~-.,\"=_^#vOg")
		var sb strings.Builder
		for y := 0; y < m.Height; y += 2 {
			for x := range m.Width {
				i := y*m.Width + x
				switch {
				case x == m.Spawn.X && (y == m.Spawn.Y || y+1 == m.Spawn.Y):
					sb.WriteByte('@')
				case m.Layers.Objects[i] == Tree || m.Layers.Objects[i] == Pine:
					sb.WriteByte('T')
				default:
					sb.WriteByte(glyph[m.Layers.Ground[i]])
				}
			}
			sb.WriteByte('\n')
		}
		t.Log("\n" + sb.String())
	}
}
