package world

import "testing"

func TestRiversAreFreshToTheirMouthAndTheSeaIsNot(t *testing.T) {
	m := Generate(GenOptions{Name: "t", Width: 96, Height: 96, Seed: 7})
	fresh := FreshWater(m)
	g := BuildGeoModel(m)
	rivers, mouths := 0, 0
	for i, r := range g.RiverOf {
		gr := m.Layers.Ground[i]
		if r < 0 || (gr != Water && gr != DeepWater) {
			continue
		}
		rivers++
		if !fresh[i] {
			t.Fatalf("river tile %d is not fresh", i)
		}
		// A mouth: river water next to sea water.
		x, y := i%m.Width, i/m.Width
		for _, d := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			if nx, ny := x+d[0], y+d[1]; nx >= 0 && ny >= 0 && nx < m.Width && ny < m.Height {
				j := ny*m.Width + nx
				if (m.Layers.Ground[j] == Water || m.Layers.Ground[j] == DeepWater) && !fresh[j] {
					mouths++
				}
			}
		}
	}
	if rivers == 0 || mouths == 0 {
		t.Fatalf("want rivers reaching the sea on this map: %d river tiles, %d mouths", rivers, mouths)
	}
	for _, i := range []int{0, m.Width - 1, len(fresh) - 1} {
		if fresh[i] {
			t.Fatalf("the sea at the map edge (tile %d) counts as fresh", i)
		}
	}
}
