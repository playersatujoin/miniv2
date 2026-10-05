package world

import (
	"fmt"
	"slices"
	"strings"
	"testing"
)

func generatedMaps() []*Map {
	var maps []*Map
	for _, size := range []int{48, 64, 128, 256} {
		for _, seed := range []uint32{1, 7, 42, 1337} {
			maps = append(maps, Generate(GenOptions{Name: "m", Width: size, Height: size, Seed: seed}))
		}
	}
	return maps
}

func TestGeoModelHasEveryUnitAndFeature(t *testing.T) {
	kinds := []string{"volcano", "pluton", "pegmatite", "ophiolite", "basin", "evaporite", "carbonatite", "karst", "river", "delta"}
	for _, m := range generatedMaps() {
		g := BuildGeoModel(m)
		count := make([]int, numRocks)
		for i, r := range g.Rock {
			if g.Land(i) {
				count[r]++
			}
		}
		for r := RockAlluvium; r < numRocks; r++ {
			if count[r] == 0 {
				t.Errorf("%d×%d seed %d: no %s", m.Width, m.Height, m.Seed, RockTypes[r].Key)
			}
		}
		for _, k := range kinds {
			if !slices.ContainsFunc(g.Features, func(f Feature) bool { return f.Kind == k }) {
				t.Errorf("%d×%d seed %d: no %s feature", m.Width, m.Height, m.Seed, k)
			}
		}
		var tin, porphyry bool
		for _, p := range g.Plutons {
			tin = tin || p.Tin
			porphyry = porphyry || p.Porphyry
		}
		if !tin || !porphyry {
			t.Errorf("%d×%d seed %d: tin granite %v, porphyry %v", m.Width, m.Height, m.Seed, tin, porphyry)
		}
	}
}

func TestGeoModelIgnoresEdits(t *testing.T) {
	m := Generate(GenOptions{Name: "m", Width: 96, Height: 96, Seed: 9})
	a := BuildGeoModel(m)
	edited := *m
	edited.Layers.Ground = slices.Clone(m.Layers.Ground)
	edited.Layers.Objects = make([]int, len(m.Layers.Objects))
	for i := range edited.Layers.Ground {
		edited.Layers.Ground[i] = Grass
	}
	b := BuildGeoModel(&edited)
	if !slices.Equal(a.Rock, b.Rock) || !slices.Equal(a.Features, b.Features) || len(a.Rivers) != len(b.Rivers) {
		t.Fatal("geology changed when only the terrain was edited")
	}
	for r := range a.Rivers {
		if !slices.Equal(a.Rivers[r], b.Rivers[r]) {
			t.Fatalf("river %d changed", r)
		}
	}
}

func TestVolcanoesHaveCratersAndCones(t *testing.T) {
	for _, m := range generatedMaps() {
		g := BuildGeoModel(m)
		if len(g.Volcanoes) == 0 {
			t.Errorf("%d seed %d: no volcano", m.Width, m.Seed)
			continue
		}
		comp, _ := walkableComponents(m)
		home := comp[m.Spawn.Y*m.Width+m.Spawn.X]
		for _, v := range g.Volcanoes {
			craters, rimReachable := 0, false
			g.disk(v.X, v.Y, v.Crater+1.6, func(x, y int, d float64) {
				i := y*m.Width + x
				gr := m.Layers.Ground[i]
				if d <= v.Crater {
					if g.Land(i) {
						craters++
						if gr != Crater || !GroundTiles[gr].Solid {
							t.Errorf("%d seed %d: crater tile (%d,%d) is %s", m.Width, m.Seed, x, y, GroundTiles[gr].Key)
						}
					}
					return
				}
				switch gr {
				case VolcanicRock, Crater, Water, DeepWater, Sand, Bridge:
				default:
					t.Errorf("%d seed %d: crater rim (%d,%d) is %s, want volcanic rock", m.Width, m.Seed, x, y, GroundTiles[gr].Key)
				}
				if comp[i] == home {
					rimReachable = true
				}
			})
			if craters == 0 {
				t.Errorf("%d seed %d: %s has no crater", m.Width, m.Seed, v.Name)
			}
			if !rimReachable {
				t.Errorf("%d seed %d: rim of %s cannot be reached from spawn", m.Width, m.Seed, v.Name)
			}
		}
	}
}

func TestRiversFlowDownhillToTheSea(t *testing.T) {
	for _, m := range generatedMaps() {
		g := BuildGeoModel(m)
		_, filled := g.drainage()
		if len(g.Rivers) < 2 {
			t.Errorf("%d seed %d: %d rivers", m.Width, m.Seed, len(g.Rivers))
		}
		for r, path := range g.Rivers {
			last := path[len(path)-1]
			li := last.Y*m.Width + last.X
			if g.Land(li) && g.RiverOf[li] == int16(r) {
				t.Errorf("%d seed %d: river %d ends on land", m.Width, m.Seed, r)
			}
			first := path[0]
			if g.Elevation[li] >= g.Elevation[first.Y*m.Width+first.X] {
				t.Errorf("%d seed %d: river %d does not descend", m.Width, m.Seed, r)
			}
			// Water may cross a closed hollow, filling it like a lake, but its
			// spill level (filled elevation) never rises downstream.
			for k := 1; k < len(path); k++ {
				a, b := path[k-1], path[k]
				if filled[b.Y*m.Width+b.X] > filled[a.Y*m.Width+a.X]+1e-9 {
					t.Errorf("%d seed %d: river %d flows uphill at (%d,%d)", m.Width, m.Seed, r, b.X, b.Y)
					break
				}
			}
			for _, p := range path {
				i := p.Y*m.Width + p.X
				if g.RiverOf[i] == int16(r) {
					if gr := m.Layers.Ground[i]; gr != Water && gr != Bridge {
						t.Errorf("%d seed %d: river %d tile (%d,%d) is %s", m.Width, m.Seed, r, p.X, p.Y, GroundTiles[gr].Key)
					}
				}
			}
		}
		if !m.Walkable(m.Spawn.X, m.Spawn.Y) || g.RiverOf[m.Spawn.Y*m.Width+m.Spawn.X] >= 0 {
			t.Errorf("%d seed %d: spawn blocked", m.Width, m.Seed)
		}
	}
}

func TestIslandStaysConnected(t *testing.T) {
	for _, m := range generatedMaps() {
		comp, sizes := walkableComponents(m)
		home := comp[m.Spawn.Y*m.Width+m.Spawn.X]
		total := 0
		for _, s := range sizes {
			total += s
		}
		if share := float64(sizes[home]) / float64(total); share < 0.95 {
			t.Errorf("%d seed %d: spawn reaches only %.0f%% of walkable land", m.Width, m.Seed, share*100)
		}
	}
}

func TestGeologyDump(t *testing.T) {
	if !testing.Verbose() {
		t.Skip("run with -v to print the geological map")
	}
	m := Generate(GenOptions{Name: "Starter Island", Width: 128, Height: 128, Seed: 1337})
	g := BuildGeoModel(m)
	letters := []byte("~asgevGpucm")
	var sb strings.Builder
	for y := 0; y < m.Height; y += 2 {
		for x := range m.Width {
			i := y*m.Width + x
			switch {
			case m.Layers.Ground[i] == Crater:
				sb.WriteByte('O')
			case g.RiverOf[i] >= 0:
				sb.WriteByte('=')
			case x == m.Spawn.X && (y == m.Spawn.Y || y+1 == m.Spawn.Y):
				sb.WriteByte('@')
			default:
				sb.WriteByte(letters[g.Rock[i]])
			}
		}
		sb.WriteByte('\n')
	}
	for _, f := range g.Features {
		fmt.Fprintf(&sb, "%-12s %-28s (%d,%d)\n", f.Kind, f.Name, f.X, f.Y)
	}
	t.Log("\n~ water  a alluvium  s sedimentary  g limestone  e evaporite  v volcanic  G granite  p pegmatite  u ultramafic  c carbonatite  m metamorphic  O crater  = river\n" + sb.String())
}
