package world

// FreshWater marks the water tiles people can drink: river channels from
// the map's geology, and any water that doesn't reach the edge of the map
// (lakes, and ponds painted in the editor). The rest is the sea. A river is
// fresh all the way to its mouth even though it joins the sea there.
func FreshWater(m *Map) []bool {
	w, h := m.Width, m.Height
	n := w * h
	water := func(i int) bool { g := m.Layers.Ground[i]; return g == Water || g == DeepWater }
	fresh := make([]bool, n)
	g := BuildGeoModel(m)
	for i := range n {
		fresh[i] = water(i) && i < len(g.RiverOf) && g.RiverOf[i] >= 0
	}
	// The sea: water reaching the edge, flooding through everything but
	// river channels.
	sea := make([]bool, n)
	var queue []int
	for i := range n {
		if x, y := i%w, i/w; water(i) && !fresh[i] && (x == 0 || y == 0 || x == w-1 || y == h-1) {
			sea[i] = true
			queue = append(queue, i)
		}
	}
	for len(queue) > 0 {
		i := queue[0]
		queue = queue[1:]
		x, y := i%w, i/w
		for _, d := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			nx, ny := x+d[0], y+d[1]
			if nx < 0 || ny < 0 || nx >= w || ny >= h {
				continue
			}
			if j := ny*w + nx; water(j) && !fresh[j] && !sea[j] {
				sea[j] = true
				queue = append(queue, j)
			}
		}
	}
	for i := range n {
		fresh[i] = water(i) && !sea[i]
	}
	return fresh
}
