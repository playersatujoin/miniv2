package sim

import (
	"cmp"
	"math"
	"slices"
)

// A village's land is a convex hull, as RAGE's map zones are
// (game/MapZones.h, CConvexHull2D): the hull is all that is kept and saved,
// and "is this point in the zone" is answered from it. Here the answer is
// painted once into a per-tile grid whenever the hulls change, so a sense
// costs one array read.

// point is a position in tiles.
type point = [2]float64

// cross is the z of (a-o)×(b-o): positive when o→a→b turns counter-clockwise.
func cross(o, a, b point) float64 {
	return (a[0]-o[0])*(b[1]-o[1]) - (a[1]-o[1])*(b[0]-o[0])
}

// convexHull is the convex hull of pts, counter-clockwise (positive area in
// map coordinates: x right, y down, so it runs clockwise on screen), without
// collinear points, by Andrew's monotone chain. It sorts pts in place. One or
// two distinct points come back as they are.
func convexHull(pts []point) []point {
	slices.SortFunc(pts, func(a, b point) int {
		if c := cmp.Compare(a[0], b[0]); c != 0 {
			return c
		}
		return cmp.Compare(a[1], b[1])
	})
	pts = slices.Compact(pts)
	if len(pts) < 3 {
		return slices.Clone(pts)
	}
	hull := make([]point, 0, 2*len(pts))
	for _, p := range pts { // lower chain
		for len(hull) >= 2 && cross(hull[len(hull)-2], hull[len(hull)-1], p) <= 0 {
			hull = hull[:len(hull)-1]
		}
		hull = append(hull, p)
	}
	lower := len(hull) + 1
	for i := len(pts) - 2; i >= 0; i-- { // upper chain
		p := pts[i]
		for len(hull) >= lower && cross(hull[len(hull)-2], hull[len(hull)-1], p) <= 0 {
			hull = hull[:len(hull)-1]
		}
		hull = append(hull, p)
	}
	return slices.Clip(hull[:len(hull)-1])
}

// landOctagon surrounds a point with eight corners at a distance chosen so
// the octagon's sides touch the circle of radius landMargin: every point
// within landMargin of the core lies inside.
var landOctagon = func() [8]point {
	var o [8]point
	r := landMargin / math.Cos(math.Pi/8)
	for k := range o {
		a := float64(k)*math.Pi/4 + math.Pi/8
		o[k] = point{r * math.Cos(a), r * math.Sin(a)}
	}
	return o
}()

// landHull is the land around pts: their convex hull grown by landMargin on
// every side (the Minkowski sum with an octagon), so even one or two houses
// have land.
func landHull(pts []point) []point {
	core := convexHull(pts)
	grown := make([]point, 0, 8*len(core))
	for _, p := range core {
		for _, o := range landOctagon {
			grown = append(grown, point{p[0] + o[0], p[1] + o[1]})
		}
	}
	return convexHull(grown)
}

// inHull reports whether (x, y) is inside or on the counter-clockwise convex
// polygon h.
func inHull(h []point, x, y float64) bool {
	if len(h) < 3 {
		return false
	}
	for i, a := range h {
		b := h[(i+1)%len(h)]
		if (b[0]-a[0])*(y-a[1])-(b[1]-a[1])*(x-a[0]) < 0 {
			return false
		}
	}
	return true
}

// paintLand writes each village's ID on the tiles whose centre lies in its
// land. Where two lands overlap the tile goes to the village whose centre is
// nearer (the older one on a tie), so each keeps its own core.
func (s *Sim) paintLand() {
	vs := s.vil()
	t := s.terrain
	if n := t.w * t.h; len(vs.land) != n {
		vs.land = make([]int32, n)
	} else {
		clear(vs.land)
	}
	for _, v := range vs.list { // by ID: older first
		v.area = 0
		if len(v.Hull) < 3 {
			continue
		}
		minX, minY, maxX, maxY := v.Hull[0][0], v.Hull[0][1], v.Hull[0][0], v.Hull[0][1]
		for _, p := range v.Hull[1:] {
			minX, maxX = math.Min(minX, p[0]), math.Max(maxX, p[0])
			minY, maxY = math.Min(minY, p[1]), math.Max(maxY, p[1])
		}
		x0, x1 := max(0, int(math.Floor(minX))), min(t.w-1, int(math.Floor(maxX)))
		y0, y1 := max(0, int(math.Floor(minY))), min(t.h-1, int(math.Floor(maxY)))
		for y := y0; y <= y1; y++ {
			for x := x0; x <= x1; x++ {
				cx, cy := float64(x)+0.5, float64(y)+0.5
				if !inHull(v.Hull, cx, cy) {
					continue
				}
				i := y*t.w + x
				if cur := vs.land[i]; cur != 0 {
					o := vs.byID[int64(cur)]
					if o != nil && math.Hypot(o.X-cx, o.Y-cy) <= math.Hypot(v.X-cx, v.Y-cy) {
						continue
					}
					if o != nil {
						o.area--
					}
				}
				vs.land[i] = int32(v.ID)
				v.area++
			}
		}
	}
}

// landAt is the village whose land (x, y) lies in, 0 for none.
func (s *Sim) landAt(x, y float64) int64 {
	vs := s.vil()
	if i, ok := s.terrain.indexAt(x, y); ok && i < len(vs.land) {
		return int64(vs.land[i])
	}
	return 0
}
