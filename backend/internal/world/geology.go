package world

import (
	"container/heap"
	"math"
	"math/rand/v2"
	"slices"
	"sync"
)

// The island is a volcanic island arc above a subduction zone, like Sumatra
// or Java: a chain of volcanoes along its spine, eroded granite plutons
// beneath it, a forearc of scraped-off ocean floor (metamorphic rocks with an
// ophiolite belt) on the trench side, and a back-arc sedimentary basin with
// coal, limestone and an evaporite pan on the other side. Rivers carry
// weathered rock from the highlands down to deltas and beaches.

// Rock is a rock unit of the geological map.
type Rock uint8

const (
	RockNone Rock = iota // sea, lakes and river channels
	RockAlluvium
	RockSedimentary
	RockLimestone
	RockEvaporite
	RockVolcanic
	RockGranite
	RockPegmatite
	RockUltramafic
	RockCarbonatite
	RockMetamorphic
	numRocks
)

type RockType struct {
	Rock        Rock   `json:"id"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description"`
}

// RockTypes is indexed by Rock.
var RockTypes = []RockType{
	{RockNone, "none", "Perairan", "#1f4e8c", "Laut, danau dan alur sungai."},
	{RockAlluvium, "aluvium", "Aluvium", "#d8c27a", "Endapan lepas sungai dan pantai: pasir, kerikil dan lumpur. Mineral berat (plaser) terkumpul di sini."},
	{RockSedimentary, "sedimen", "Batuan Sedimen", "#b7a07a", "Batu pasir, batu lempung dan serpih di cekungan belakang busur; tempat lapisan batu bara terbentuk."},
	{RockLimestone, "gamping", "Batu Gamping", "#d9d6c8", "Batuan karbonat dari terumbu dan paparan laut dangkal; membentuk bukit karst."},
	{RockEvaporite, "evaporit", "Evaporit", "#efe3c1", "Garam yang mengendap ketika air danau atau laguna tertutup menguap di bagian cekungan yang kering."},
	{RockVolcanic, "vulkanik", "Batuan Vulkanik", "#5e5552", "Andesit dan basal dari gunung api busur kepulauan."},
	{RockGranite, "granit", "Granit", "#d39b96", "Batuan beku dalam yang membeku perlahan di bawah busur gunung api lalu tersingkap oleh erosi."},
	{RockPegmatite, "pegmatit", "Pegmatit", "#f0b8d8", "Retas berbutir sangat kasar dari sisa cairan magma granit, kaya unsur langka."},
	{RockUltramafic, "ultrabasa", "Ultrabasa (Ofiolit)", "#4f7a54", "Peridotit dan serpentinit: potongan mantel bumi dan kerak samudra yang terangkat di sisi palung."},
	{RockCarbonatite, "karbonatit", "Karbonatit", "#e0874a", "Batuan beku langka yang sebagian besar berupa mineral karbonat; sumber utama unsur tanah jarang dan niobium. Di dunia nyata umumnya di zona rekahan benua; di pulau ini dianggap sisa kerak benua tua."},
	{RockMetamorphic, "metamorf", "Batuan Metamorf", "#8c86a8", "Sekis dan rijang dari kompleks akresi di sisi palung yang terlipat dan termalihkan."},
}

// Feature is a named geological feature, for labels on the geological map.
// Kind is one of volcano, pluton, pegmatite, ophiolite, basin, evaporite,
// carbonatite, karst, river, delta.
type Feature struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
	X    int    `json:"x"`
	Y    int    `json:"y"`
}

// Volcano is a stratovolcano on the arc: an acid crater lake at the summit
// inside a cone of young volcanic rock.
type Volcano struct {
	X, Y   int
	Crater float64 // crater radius in tiles
	Apron  float64 // radius of the young cone
	Name   string
}

// Pluton is an eroded granite intrusion. The porphyry stock (I-type
// granodiorite) sits just beside a volcano; the tin granite (S-type) carries
// Sn-W veins and the pegmatite field.
type Pluton struct {
	X, Y     int
	R        float64
	Porphyry bool
	Tin      bool
	Name     string
}

// GeoModel is the geological story of a map. It depends only on the map's
// seed and size, so hand edits to the terrain never change the bedrock.
type GeoModel struct {
	W, H      int
	Rock      []Rock
	Features  []Feature
	Rivers    [][]Point // each traced from its source to the sea (or to the river it joins)
	Volcanoes []Volcano
	Plutons   []Pluton
	PlutonOf  []int16 // per tile: index into Plutons for granite and pegmatite, else -1
	RiverOf   []int16 // per tile: index into Rivers for river channel tiles, else -1
	// Down is where each tile's water flows next on its way to the sea (-1:
	// the sea itself), across closed hollows as a filling lake would.
	Down      []int32
	Elevation []float64
	Moisture  []float64
	// Elevation thresholds the generator paints with: below SeaLevel is
	// water, below ShoreLevel beach, above HighLevel rocky, above PeakLevel peaks.
	SeaLevel, ShoreLevel, HighLevel, PeakLevel float64

	along, across    []float64
	backArc, foreArc float64 // across thresholds: back-arc | arc | forearc
	arcAngle         float64
	evaporite        Point
	carbonatite      Point
	seed             uint64
	size             float64
}

// Across is the signed distance from the arc axis, normalised to the map
// size: negative on the back-arc side, positive towards the trench.
func (g *GeoModel) Across(x, y int) float64 { return g.across[y*g.W+x] }

// Along is the distance along the arc axis in tiles.
func (g *GeoModel) Along(x, y int) float64 { return g.along[y*g.W+x] }

// Size is the shorter side of the map in tiles.
func (g *GeoModel) Size() float64 { return g.size }

// ArcAngle is the direction of the arc axis in radians; sedimentary layers
// and coal seams in the basin run parallel to it.
func (g *GeoModel) ArcAngle() float64 { return g.arcAngle }

// Land reports whether the generator made tile i land.
func (g *GeoModel) Land(i int) bool { return g.Elevation[i] >= g.SeaLevel }

// VolcanoAt returns the volcano whose young cone covers (x, y), or -1.
func (g *GeoModel) VolcanoAt(x, y int) int {
	for vi, v := range g.Volcanoes {
		if math.Hypot(float64(x-v.X), float64(y-v.Y)) <= v.Apron {
			return vi
		}
	}
	return -1
}

// BuildGeoModel derives the geology of a map from its seed and size alone.
// Models are cached and shared, so callers must treat them as read-only.
func BuildGeoModel(m *Map) *GeoModel {
	key := geoKey{uint64(m.Seed), m.Width, m.Height}
	geoCache.mu.Lock()
	g, ok := geoCache.models[key]
	geoCache.mu.Unlock()
	if ok {
		return g
	}
	elev, moist := terrainFields(m.Width, m.Height, uint64(m.Seed))
	g = buildGeoModel(m.Width, m.Height, uint64(m.Seed), elev, moist)
	geoCache.mu.Lock()
	if len(geoCache.models) >= geoCacheSize {
		clear(geoCache.models) // a handful of maps at most; simplest is to start over
	}
	geoCache.models[key] = g
	geoCache.mu.Unlock()
	return g
}

type geoKey struct {
	seed uint64
	w, h int
}

const geoCacheSize = 16

var geoCache = struct {
	mu     sync.Mutex
	models map[geoKey]*GeoModel
}{models: map[geoKey]*GeoModel{}}

func buildGeoModel(w, h int, seed uint64, elev, moist []float64) *GeoModel {
	t := quantiles(elev, 0.20, 0.32, 0.37, 0.87, 0.94)
	g := &GeoModel{
		W: w, H: h,
		Rock:      make([]Rock, w*h),
		PlutonOf:  filled(w*h, -1),
		RiverOf:   filled(w*h, -1),
		Elevation: elev,
		Moisture:  moist,
		SeaLevel:  t[1], ShoreLevel: t[2], HighLevel: t[3], PeakLevel: t[4],
		along:  make([]float64, w*h),
		across: make([]float64, w*h),
		seed:   seed,
		size:   float64(min(w, h)),
	}
	rng := rand.New(rand.NewPCG(seed, 0x6e01061a5eed))
	names := &namer{rng: rng}
	g.fitArc()
	g.placeVolcanoes(names)
	g.placePlutons(rng, names)
	g.paintRocks(rng)
	g.traceRivers(rng, names)
	g.paintAlluvium()
	g.ensurePegmatite()
	g.collectFeatures(names)
	return g
}

func filled(n int, v int16) []int16 {
	s := make([]int16, n)
	for i := range s {
		s[i] = v
	}
	return s
}

func clampf(v, lo, hi float64) float64 { return math.Max(lo, math.Min(hi, v)) }

func (g *GeoModel) xy(i int) (int, int)  { return i % g.W, i / g.W }
func (g *GeoModel) inside(x, y int) bool { return x >= 0 && y >= 0 && x < g.W && y < g.H }
func (g *GeoModel) noise(x, y, scale float64, salt uint64) float64 {
	return noise2D{g.seed ^ salt}.at(x/scale, y/scale)
}

// disk calls fn for every tile within r of (cx, cy).
func (g *GeoModel) disk(cx, cy int, r float64, fn func(x, y int, d float64)) {
	ir := int(math.Ceil(r))
	for y := cy - ir; y <= cy+ir; y++ {
		for x := cx - ir; x <= cx+ir; x++ {
			if !g.inside(x, y) {
				continue
			}
			if d := math.Hypot(float64(x-cx), float64(y-cy)); d <= r {
				fn(x, y, d)
			}
		}
	}
}

// landShare reports which fraction of the disk is land, and land below the peaks.
func (g *GeoModel) landShare(cx, cy int, r float64) (land, low float64) {
	n := 0
	g.disk(cx, cy, r, func(x, y int, _ float64) {
		n++
		e := g.Elevation[y*g.W+x]
		if e >= g.SeaLevel {
			land++
			if e < g.PeakLevel {
				low++
			}
		}
	})
	if n == 0 {
		return 0, 0
	}
	return land / float64(n), low / float64(n)
}

// fitArc lays the arc axis along the island's mountainous spine (the principal
// axis of its high ground) and picks which side faces the trench.
func (g *GeoModel) fitArc() {
	weigh := func(i int) float64 {
		if e := g.Elevation[i]; e >= g.HighLevel {
			return e - g.HighLevel + 0.01
		}
		return 0
	}
	var sw, sx, sy float64
	sum := func() {
		sw, sx, sy = 0, 0, 0
		for i := range g.Elevation {
			if wt := weigh(i); wt > 0 {
				x, y := g.xy(i)
				sw += wt
				sx += wt * float64(x)
				sy += wt * float64(y)
			}
		}
	}
	sum()
	if sw == 0 {
		weigh = func(i int) float64 {
			if g.Land(i) {
				return 1
			}
			return 0
		}
		sum()
	}
	cx, cy := float64(g.W-1)/2, float64(g.H-1)/2
	if sw > 0 {
		cx, cy = sx/sw, sy/sw
	}
	var sxx, syy, sxy float64
	for i := range g.Elevation {
		if wt := weigh(i); wt > 0 {
			x, y := g.xy(i)
			dx, dy := float64(x)-cx, float64(y)-cy
			sxx += wt * dx * dx
			syy += wt * dy * dy
			sxy += wt * dx * dy
		}
	}
	theta := 0.5 * math.Atan2(2*sxy, sxx-syy)
	if sxx == 0 && syy == 0 {
		theta = hash2(g.seed, 3, 5) * math.Pi
	}
	g.arcAngle = theta
	dx, dy := math.Cos(theta), math.Sin(theta)
	sign := 1.0
	if hash2(g.seed^0xa2c, 1, 2) < 0.5 {
		sign = -1
	}
	scale := 0.5 * g.size
	var land []float64
	for i := range g.Elevation {
		x, y := g.xy(i)
		fx, fy := float64(x)-cx, float64(y)-cy
		g.along[i] = fx*dx + fy*dy
		g.across[i] = (-fx*dy + fy*dx) * sign / scale
		if g.Land(i) {
			land = append(land, g.across[i])
		}
	}
	g.backArc, g.foreArc = -0.2, 0.2
	if len(land) > 0 {
		q := quantiles(land, 0.35, 0.65)
		g.backArc, g.foreArc = q[0], q[1]
	}
}

func (g *GeoModel) inArc(i int) bool {
	return g.across[i] >= g.backArc && g.across[i] < g.foreArc
}

func (g *GeoModel) localMax(i, r int) bool {
	x, y := g.xy(i)
	for yy := y - r; yy <= y+r; yy++ {
		for xx := x - r; xx <= x+r; xx++ {
			if g.inside(xx, yy) && g.Elevation[yy*g.W+xx] > g.Elevation[i] {
				return false
			}
		}
	}
	return true
}

// placeVolcanoes puts stratovolcanoes on the highest points of the arc.
func (g *GeoModel) placeVolcanoes(names *namer) {
	want := max(1, min(3, 1+int(g.size)/96))
	crater := 1.0
	if g.size >= 160 {
		crater = 1.5
	}
	apron := crater + 2 + g.size/64
	spacing := math.Max(8, g.size/4)
	margin := int(math.Ceil(apron)) + 1

	tiers := []func(i int) bool{
		func(i int) bool { return g.Elevation[i] >= g.PeakLevel && g.inArc(i) && g.localMax(i, 3) },
		func(i int) bool { return g.Elevation[i] >= g.PeakLevel && g.localMax(i, 3) },
		func(i int) bool { return g.Elevation[i] >= g.HighLevel && g.localMax(i, 2) },
		func(i int) bool { return g.Land(i) },
	}
	for _, ok := range tiers {
		var cands []int
		for i := range g.Elevation {
			x, y := g.xy(i)
			if x >= margin && y >= margin && x < g.W-margin && y < g.H-margin && ok(i) {
				cands = append(cands, i)
			}
		}
		slices.SortStableFunc(cands, func(a, b int) int {
			return cmpFloat(g.Elevation[b], g.Elevation[a])
		})
		for _, i := range cands {
			if len(g.Volcanoes) >= want {
				break
			}
			x, y := g.xy(i)
			if slices.ContainsFunc(g.Volcanoes, func(v Volcano) bool {
				return math.Hypot(float64(x-v.X), float64(y-v.Y)) < spacing
			}) {
				continue
			}
			g.Volcanoes = append(g.Volcanoes, Volcano{X: x, Y: y, Crater: crater, Apron: apron, Name: "Gunung " + names.next()})
		}
		if len(g.Volcanoes) > 0 {
			return
		}
	}
}

func cmpFloat(a, b float64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

// placePlutons chooses where eroded granite intrusions crop out.
func (g *GeoModel) placePlutons(rng *rand.Rand, names *namer) {
	count := 2 + int(g.size)/96
	base := clampf(g.size*0.055, 2.5, 8)

	var cands []int
	for i := range g.Elevation {
		if g.Elevation[i] >= g.ShoreLevel {
			cands = append(cands, i)
		}
	}
	rng.Shuffle(len(cands), func(a, b int) { cands[a], cands[b] = cands[b], cands[a] })

	clearOfVolcanoes := func(x, y int, r float64) bool {
		for _, v := range g.Volcanoes {
			if math.Hypot(float64(x-v.X), float64(y-v.Y)) < v.Apron+r*0.6 {
				return false
			}
		}
		return true
	}
	clearOfPlutons := func(x, y int, r float64) bool {
		for _, p := range g.Plutons {
			if math.Hypot(float64(x-p.X), float64(y-p.Y)) < p.R+r+3 {
				return false
			}
		}
		return true
	}
	// choose returns the best candidate under progressively relaxed land requirements.
	choose := func(r float64, accept func(i int) bool, score func(x, y int) float64) int {
		for _, need := range []float64{0.75, 0.45, 0} {
			best, bestScore := -1, math.Inf(1)
			for _, i := range cands {
				x, y := g.xy(i)
				if !accept(i) || !clearOfVolcanoes(x, y, r) || !clearOfPlutons(x, y, r) {
					continue
				}
				if land, low := g.landShare(x, y, r); land < need || low < need*0.7 {
					continue
				}
				if score == nil {
					return i // first fit
				}
				if s := score(x, y); s < bestScore {
					best, bestScore = i, s
				}
			}
			if best >= 0 {
				return best
			}
		}
		return -1
	}
	add := func(i int, r float64, porphyry, tin bool) {
		x, y := g.xy(i)
		name := "Granit " + names.next()
		if porphyry {
			name = "Intrusi Porfiri " + names.next()
		}
		g.Plutons = append(g.Plutons, Pluton{X: x, Y: y, R: r, Porphyry: porphyry, Tin: tin, Name: name})
	}

	// The porphyry stock sits just outside the first volcano's cone.
	r0 := base * (0.85 + 0.3*rng.Float64())
	near := func(x, y int) float64 {
		if len(g.Volcanoes) == 0 {
			return -g.Elevation[y*g.W+x]
		}
		v := g.Volcanoes[0]
		return math.Abs(math.Hypot(float64(x-v.X), float64(y-v.Y)) - (v.Apron + r0 + 1))
	}
	arcish := func(i int) bool {
		return g.across[i] >= g.backArc-0.1 && g.across[i] < g.foreArc+0.1
	}
	if i := choose(r0, arcish, near); i >= 0 {
		add(i, r0, true, false)
	}

	// The tin granite lies on the back-arc side, as in the Southeast Asian tin belt.
	r1 := base * (0.9 + 0.3*rng.Float64())
	mid := (g.backArc + g.foreArc) / 2
	if i := choose(r1, func(i int) bool { return g.across[i] < mid }, nil); i >= 0 {
		add(i, r1, false, true)
	} else if i := choose(r1, func(int) bool { return true }, nil); i >= 0 {
		add(i, r1, false, true)
	}

	for len(g.Plutons) < count {
		r := base * (0.75 + 0.4*rng.Float64())
		i := choose(r, func(i int) bool { return g.across[i] < g.foreArc+0.1 }, nil)
		if i < 0 {
			break
		}
		add(i, r, false, false)
	}
}

func (g *GeoModel) inPluton(p Pluton, x, y int) bool {
	d := math.Hypot(float64(x-p.X), float64(y-p.Y))
	return d <= p.R*(0.8+0.4*g.noise(float64(x), float64(y), 3, 0x9147))
}

// paintRocks assigns bedrock units, oldest first; younger intrusions and
// volcanic cones overprint what they cut through.
func (g *GeoModel) paintRocks(rng *rand.Rand) {
	maxAcross := math.Inf(-1)
	for i := range g.Rock {
		if !g.Land(i) {
			continue
		}
		a := g.across[i]
		maxAcross = math.Max(maxAcross, a)
		switch {
		case a < g.backArc:
			g.Rock[i] = RockSedimentary
		case a < g.foreArc:
			g.Rock[i] = RockVolcanic
		default:
			g.Rock[i] = RockMetamorphic
		}
	}

	// paint marks tiles matching ok, relaxing the noise threshold until at least minimum are marked.
	paint := func(rock Rock, minimum int, ok func(i int, relax float64) bool) {
		for _, relax := range []float64{0, 0.15, 0.3, 1} {
			var hit []int
			for i := range g.Rock {
				if g.Land(i) && ok(i, relax) {
					hit = append(hit, i)
				}
			}
			if len(hit) >= minimum || relax == 1 {
				for _, i := range hit {
					g.Rock[i] = rock
				}
				return
			}
		}
	}

	// Ophiolite belt in the middle of the forearc, broken into segments.
	paint(RockUltramafic, 8, func(i int, relax float64) bool {
		f := (g.across[i] - g.foreArc) / math.Max(1e-9, maxAcross-g.foreArc)
		return f >= 0.2-relax/2 && f <= 0.8+relax/2 && g.noise(g.along[i], 0, 7, 0x0f1)+relax > 0.4
	})

	// Carbonate platform along the back-arc coast and karst hills inland.
	coastTop := g.ShoreLevel + 0.3*(g.HighLevel-g.ShoreLevel)
	paint(RockLimestone, 8, func(i int, relax float64) bool {
		if g.across[i] >= g.backArc+relax/4 {
			return false
		}
		x, y := g.xy(i)
		n := g.noise(float64(x), float64(y), 6, 0x11e) + relax
		e := g.Elevation[i]
		return (e < coastTop && n > 0.55) || (e >= g.HighLevel && n > 0.5)
	})

	// An evaporite pan in the driest lowland of the basin.
	lowland := func(i int) bool { return g.Elevation[i] >= g.ShoreLevel && g.Elevation[i] < g.HighLevel }
	// The pan lies in a quiet basin, clear of the intrusions and the cones.
	quiet := func(i int) bool {
		x, y := g.xy(i)
		for _, p := range g.Plutons {
			if math.Hypot(float64(x-p.X), float64(y-p.Y)) < p.R+3 {
				return false
			}
		}
		for _, v := range g.Volcanoes {
			if math.Hypot(float64(x-v.X), float64(y-v.Y)) < v.Apron+3 {
				return false
			}
		}
		return true
	}
	best := -1
	for pass := range 3 {
		for i := range g.Rock {
			if !g.Land(i) || !lowland(i) || g.Rock[i] == RockLimestone || (pass == 0 && g.across[i] >= g.backArc) || (pass < 2 && !quiet(i)) {
				continue
			}
			if best < 0 || g.Moisture[i] < g.Moisture[best] {
				best = i
			}
		}
		if best >= 0 {
			break
		}
	}
	if best >= 0 {
		ex, ey := g.xy(best)
		g.evaporite = Point{ex, ey}
		re := clampf(g.size/30, 2, 5)
		paintPan := func(r float64) int {
			n := 0
			g.disk(ex, ey, r*1.2, func(x, y int, d float64) {
				i := y*g.W + x
				if g.Land(i) && lowland(i) && d <= r*(0.8+0.4*g.noise(float64(x), float64(y), 2, 0xe7a)) {
					g.Rock[i] = RockEvaporite
					n++
				}
			})
			return n
		}
		// A pan squeezed against the coast or hills widens until it has room.
		for r := re; paintPan(r) < 10 && r < re*3; r += 1 {
		}
		g.Rock[best] = RockEvaporite
	}

	// One small carbonatite plug, away from the arc's intrusions.
	rk := clampf(g.size/50, 1.5, 3)
	var cands []int
	for i := range g.Rock {
		if g.Land(i) && g.Elevation[i] >= g.ShoreLevel && g.Elevation[i] < g.PeakLevel && g.across[i] < g.foreArc {
			cands = append(cands, i)
		}
	}
	rng.Shuffle(len(cands), func(a, b int) { cands[a], cands[b] = cands[b], cands[a] })
	away := func(x, y int, slack float64) bool {
		for _, p := range g.Plutons {
			if math.Hypot(float64(x-p.X), float64(y-p.Y)) < p.R+3-slack {
				return false
			}
		}
		for _, v := range g.Volcanoes {
			if math.Hypot(float64(x-v.X), float64(y-v.Y)) < v.Apron+3-slack {
				return false
			}
		}
		if best >= 0 && math.Hypot(float64(x-g.evaporite.X), float64(y-g.evaporite.Y)) < clampf(g.size/30, 2, 5)+3-slack {
			return false
		}
		return true
	}
	placed := false
	for _, need := range []float64{0.85, 0.6, 0} {
		for _, i := range cands {
			x, y := g.xy(i)
			if !away(x, y, (0.85-need)*6) {
				continue
			}
			if land, low := g.landShare(x, y, rk); land < need || low < need*0.8 {
				continue
			}
			g.carbonatite = Point{x, y}
			g.disk(x, y, rk, func(xx, yy int, _ float64) {
				if j := yy*g.W + xx; g.Land(j) {
					g.Rock[j] = RockCarbonatite
				}
			})
			placed = true
			break
		}
		if placed {
			break
		}
	}

	// Granite plutons and the pegmatite dykes of the tin granite.
	for pi, p := range g.Plutons {
		g.disk(p.X, p.Y, p.R*1.25, func(x, y int, _ float64) {
			i := y*g.W + x
			if g.Land(i) && g.inPluton(p, x, y) {
				g.Rock[i] = RockGranite
				g.PlutonOf[i] = int16(pi)
			}
		})
	}
	for pi, p := range g.Plutons {
		if !p.Tin {
			continue
		}
		phi := rng.Float64() * math.Pi
		cs, sn := math.Cos(phi), math.Sin(phi)
		// Dykes cut the granite and run a little way into the country rock.
		var dykes []int
		g.disk(p.X, p.Y, p.R*1.1, func(x, y int, _ float64) {
			i := y*g.W + x
			if !g.Land(i) || (g.Rock[i] != RockGranite && g.VolcanoAt(x, y) >= 0) {
				return
			}
			s := float64(x)*cs + float64(y)*sn + 1.2*g.noise(float64(x), float64(y), 3, 0x9e6)
			if f := s/3 - math.Floor(s/3); f < 0.4 {
				dykes = append(dykes, i)
			}
		})
		// A small pluton gets a pegmatite body at its core as well, until enough
		// of the field crops out below the peaks.
		exposed := func() int {
			n := 0
			for _, i := range dykes {
				if g.Elevation[i] < g.PeakLevel {
					n++
				}
			}
			return n
		}
		for r := 1.0; exposed() < 12 && r <= p.R+3; r += 0.5 {
			g.disk(p.X, p.Y, r, func(x, y int, _ float64) {
				if i := y*g.W + x; g.Land(i) && !slices.Contains(dykes, i) {
					dykes = append(dykes, i)
				}
			})
		}
		for _, i := range dykes {
			g.Rock[i] = RockPegmatite
			g.PlutonOf[i] = int16(pi)
		}
	}

	// A limestone roof pendant against the porphyry stock: where intrusions
	// meet limestone, skarn forms (like the Ertsberg skarn beside Grasberg).
	for _, p := range g.Plutons {
		if !p.Porphyry {
			continue
		}
		bestDir, bestLand := -1, 0
		for k := range 8 {
			a := float64(k) * math.Pi / 4
			cx := p.X + int(math.Round(math.Cos(a)*(p.R+0.5)))
			cy := p.Y + int(math.Round(math.Sin(a)*(p.R+0.5)))
			n := 0
			g.disk(cx, cy, 1.8, func(x, y int, _ float64) {
				if g.Land(y*g.W + x) {
					n++
				}
			})
			// Prefer the back-arc side, where limestone belongs.
			if cx2, cy2 := clampInt(cx, 0, g.W-1), clampInt(cy, 0, g.H-1); g.across[cy2*g.W+cx2] < g.across[p.Y*g.W+p.X] {
				n += 2
			}
			if n > bestLand {
				bestDir, bestLand = k, n
			}
		}
		if bestDir < 0 {
			continue
		}
		a := float64(bestDir) * math.Pi / 4
		cx := p.X + int(math.Round(math.Cos(a)*(p.R+0.5)))
		cy := p.Y + int(math.Round(math.Sin(a)*(p.R+0.5)))
		g.disk(cx, cy, 1.8, func(x, y int, _ float64) {
			if i := y*g.W + x; g.Land(i) {
				g.Rock[i] = RockLimestone
				g.PlutonOf[i] = -1
			}
		})
	}

	// Young volcanic cones cover everything older.
	for _, v := range g.Volcanoes {
		g.disk(v.X, v.Y, v.Apron, func(x, y int, _ float64) {
			if i := y*g.W + x; g.Land(i) {
				g.Rock[i] = RockVolcanic
				g.PlutonOf[i] = -1
			}
		})
	}
}

func clampInt(v, lo, hi int) int { return max(lo, min(hi, v)) }

// traceRivers runs rivers from the highlands down to the sea along the
// drainage network. The first drains the tin granite and the second the first
// volcano's flank, so their placers carry tin and gold the way real rivers do.
func (g *GeoModel) traceRivers(rng *rand.Rand, names *namer) {
	want := max(2, min(6, int(g.size)/40))
	minSpacing := math.Max(6, g.size/6)
	down, _ := g.drainage()
	g.Down = down
	var sources []Point

	high := func(i int) bool {
		e := g.Elevation[i]
		return g.Land(i) && e >= g.ShoreLevel+0.5*(g.HighLevel-g.ShoreLevel) && e < g.PeakLevel
	}
	pickNear := func(cx, cy int, rmin, rmax float64) (Point, bool) {
		best, bestE := Point{}, math.Inf(-1)
		g.disk(cx, cy, rmax, func(x, y int, d float64) {
			i := y*g.W + x
			if d < rmin || !high(i) {
				return
			}
			if e := g.Elevation[i]; e > bestE {
				best, bestE = Point{x, y}, e
			}
		})
		return best, !math.IsInf(bestE, -1)
	}
	for _, p := range g.Plutons {
		if p.Tin {
			if s, ok := pickNear(p.X, p.Y, 0, p.R+3); ok {
				sources = append(sources, s)
			}
		}
	}
	if len(g.Volcanoes) > 0 {
		v := g.Volcanoes[0]
		if s, ok := pickNear(v.X, v.Y, v.Apron+0.5, v.Apron+5); ok {
			sources = append(sources, s)
		}
	}
	var rest []int
	for i := range g.Elevation {
		if high(i) {
			rest = append(rest, i)
		}
	}
	rng.Shuffle(len(rest), func(a, b int) { rest[a], rest[b] = rest[b], rest[a] })
	for _, i := range rest {
		x, y := g.xy(i)
		sources = append(sources, Point{x, y})
	}

	for _, s := range sources {
		if len(g.Rivers) >= want {
			break
		}
		if g.RiverOf[s.Y*g.W+s.X] >= 0 || slices.ContainsFunc(g.Rivers, func(r []Point) bool { return dist(r[0], s) < minSpacing }) {
			continue
		}
		var path []Point
		for i := s.Y*g.W + s.X; ; i = int(down[i]) {
			x, y := g.xy(i)
			path = append(path, Point{x, y})
			if !g.Land(i) || down[i] < 0 || (len(path) > 1 && g.RiverOf[i] >= 0) {
				break
			}
		}
		if len(path) < 6 {
			continue
		}
		ri := int16(len(g.Rivers))
		for _, p := range path {
			if i := p.Y*g.W + p.X; g.Land(i) && g.RiverOf[i] < 0 {
				g.RiverOf[i] = ri
			}
		}
		g.Rivers = append(g.Rivers, path)
	}
}

// drainage returns, for every land tile, the neighbour its water flows to on
// the way to the sea (-1 for the sea itself). It floods the terrain from the
// coast upwards (priority-flood), so water crosses closed hollows the way a
// river fills and overflows a lake instead of getting stuck.
func (g *GeoModel) drainage() (down []int32, filled []float64) {
	n := g.W * g.H
	down = make([]int32, n)
	filled = make([]float64, n)
	done := make([]bool, n)
	q := &floodQueue{}
	for i := range down {
		down[i] = -1
		if !g.Land(i) {
			filled[i] = g.Elevation[i]
			done[i] = true
			heap.Push(q, floodCell{i, filled[i]})
		}
	}
	for q.Len() > 0 {
		c := heap.Pop(q).(floodCell)
		x, y := g.xy(c.i)
		for _, d := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			nx, ny := x+d[0], y+d[1]
			if !g.inside(nx, ny) {
				continue
			}
			j := ny*g.W + nx
			if done[j] {
				continue
			}
			done[j] = true
			down[j] = int32(c.i)
			// Craters drain nowhere: keep rivers off them.
			filled[j] = math.Max(g.Elevation[j], c.e+1e-6)
			for _, v := range g.Volcanoes {
				if math.Hypot(float64(nx-v.X), float64(ny-v.Y)) <= v.Crater+0.5 {
					filled[j] += 1
				}
			}
			heap.Push(q, floodCell{j, filled[j]})
		}
	}
	return down, filled
}

type floodCell struct {
	i int
	e float64
}

type floodQueue []floodCell

func (q floodQueue) Len() int { return len(q) }
func (q floodQueue) Less(a, b int) bool {
	if q[a].e != q[b].e {
		return q[a].e < q[b].e
	}
	return q[a].i < q[b].i
}
func (q floodQueue) Swap(a, b int) { q[a], q[b] = q[b], q[a] }
func (q *floodQueue) Push(x any)   { *q = append(*q, x.(floodCell)) }
func (q *floodQueue) Pop() any {
	old := *q
	c := old[len(old)-1]
	*q = old[:len(old)-1]
	return c
}

// paintAlluvium spreads loose sediment over floodplains, deltas and beaches.
func (g *GeoModel) paintAlluvium() {
	coverable := func(i int) bool {
		x, y := g.xy(i)
		switch g.Rock[i] {
		case RockSedimentary, RockMetamorphic:
			return true
		case RockVolcanic:
			return g.VolcanoAt(x, y) < 0
		}
		return false
	}
	for i, r := range g.RiverOf {
		if r >= 0 {
			g.Rock[i] = RockNone
		}
	}
	for _, path := range g.Rivers {
		n := len(path)
		for k, p := range path {
			if k < int(0.3*float64(n)) {
				continue
			}
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					x, y := p.X+dx, p.Y+dy
					if !g.inside(x, y) {
						continue
					}
					if i := y*g.W + x; g.Land(i) && g.RiverOf[i] < 0 && coverable(i) {
						g.Rock[i] = RockAlluvium
					}
				}
			}
		}
		mouth := g.mouth(path)
		g.disk(mouth.X, mouth.Y, 2+g.size/96, func(x, y int, _ float64) {
			if i := y*g.W + x; g.Land(i) && g.RiverOf[i] < 0 && coverable(i) {
				g.Rock[i] = RockAlluvium
			}
		})
	}
	for i := range g.Rock {
		if g.Land(i) && g.Elevation[i] < g.ShoreLevel && coverable(i) {
			g.Rock[i] = RockAlluvium
		}
	}
}

// ensurePegmatite makes sure the pegmatite field still crops out after cones
// and rivers overprinted it: on a cramped island more of the tin granite's
// cupola is pegmatite, nearest the centre first.
func (g *GeoModel) ensurePegmatite() {
	exposed := func(i int) bool { return g.Elevation[i] < g.PeakLevel && g.RiverOf[i] < 0 }
	n := 0
	for i, r := range g.Rock {
		if r == RockPegmatite && exposed(i) {
			n++
		}
	}
	for pi, p := range g.Plutons {
		if !p.Tin || n >= 12 {
			continue
		}
		var granite []int
		for i, r := range g.Rock {
			if r == RockGranite && g.PlutonOf[i] == int16(pi) && exposed(i) {
				granite = append(granite, i)
			}
		}
		slices.SortStableFunc(granite, func(a, b int) int {
			ax, ay := g.xy(a)
			bx, by := g.xy(b)
			return cmpFloat(math.Hypot(float64(ax-p.X), float64(ay-p.Y)), math.Hypot(float64(bx-p.X), float64(by-p.Y)))
		})
		for _, i := range granite {
			if n >= 12 {
				break
			}
			g.Rock[i] = RockPegmatite
			n++
		}
	}
}

// mouth is the last land tile of a river's channel.
func (g *GeoModel) mouth(path []Point) Point {
	for k := len(path) - 1; k >= 0; k-- {
		if g.Land(path[k].Y*g.W + path[k].X) {
			return path[k]
		}
	}
	return path[0]
}

// representative returns the tile of a unit closest to the unit's centroid,
// so labels sit on the unit itself.
func (g *GeoModel) representative(match func(i int) bool) (Point, bool) {
	var sx, sy float64
	var tiles []int
	for i := range g.Rock {
		if match(i) {
			x, y := g.xy(i)
			sx += float64(x)
			sy += float64(y)
			tiles = append(tiles, i)
		}
	}
	if len(tiles) == 0 {
		return Point{}, false
	}
	cx, cy := sx/float64(len(tiles)), sy/float64(len(tiles))
	best, bestD := tiles[0], math.Inf(1)
	for _, i := range tiles {
		x, y := g.xy(i)
		if d := math.Hypot(float64(x)-cx, float64(y)-cy); d < bestD {
			best, bestD = i, d
		}
	}
	x, y := g.xy(best)
	return Point{x, y}, true
}

func (g *GeoModel) collectFeatures(names *namer) {
	add := func(kind, name string, p Point) {
		g.Features = append(g.Features, Feature{Kind: kind, Name: name, X: p.X, Y: p.Y})
	}
	for _, v := range g.Volcanoes {
		add("volcano", v.Name, Point{v.X, v.Y})
	}
	for _, p := range g.Plutons {
		add("pluton", p.Name, Point{p.X, p.Y})
	}
	units := []struct {
		kind, prefix string
		rock         Rock
	}{
		{"pegmatite", "Pegmatit ", RockPegmatite},
		{"ophiolite", "Ofiolit ", RockUltramafic},
		{"basin", "Cekungan ", RockSedimentary},
		{"evaporite", "Dataran Garam ", RockEvaporite},
		{"carbonatite", "Karbonatit ", RockCarbonatite},
		{"karst", "Karst ", RockLimestone},
	}
	for _, u := range units {
		if p, ok := g.representative(func(i int) bool { return g.Rock[i] == u.rock }); ok {
			add(u.kind, u.prefix+names.next(), p)
		}
	}
	for _, path := range g.Rivers {
		name := names.next()
		add("river", "Sungai "+name, path[len(path)*2/5])
		add("delta", "Muara "+name, g.mouth(path))
	}
}

// namer makes Indonesian-sounding place names.
type namer struct {
	rng  *rand.Rand
	used []string
}

var (
	placeStarts = []string{"Ma", "Si", "Ba", "Ta", "Ka", "Lo", "Ra", "Wa", "Bu", "Pa", "Se", "Ja", "Ga", "Te", "Ni", "Sa", "Me", "Ke", "Pu", "Lu", "Ci", "Su"}
	placeMids   = []string{"ra", "la", "ma", "nda", "ri", "wa", "ti", "lu", "ka", "ngi", "ro", "si", "", ""}
	placeEnds   = []string{"pi", "ngan", "ra", "wa", "ti", "ni", "sa", "ru", "bo", "nga", "li", "jang", "tu", "lang", "ran", "mas"}
)

func (n *namer) next() string {
	pick := func(s []string) string { return s[n.rng.IntN(len(s))] }
	var name string
	for range 20 {
		name = pick(placeStarts) + pick(placeMids) + pick(placeEnds)
		if !slices.Contains(n.used, name) {
			break
		}
	}
	n.used = append(n.used, name)
	return name
}
