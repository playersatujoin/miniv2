package sim

// Fire's effects on people and their works. The fire itself spreads in the
// ecology (ecology/fire.go, after RAGE's vfx/misc/Fire.h: fuel, flammability,
// wind); here it burns fields and houses, hurts people standing in it, and
// starts from hearths and kilns in dry, windy weather.
//
// Nobody is made to light, flee or fight a fire. People feel the heat and the
// fear when they stand in one, see it (inFireNear) and feel the wind
// (inWind); what they do about it is up to their brains.

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"miniv2/backend/internal/chem"
	"miniv2/backend/internal/ecology"
)

const (
	// A body standing in flames: at full intensity a person is dead in under
	// two seconds (a couple of tiles of fire front to walk through), so a
	// fire kills those caught in it but not those who keep moving. An
	// estimate on the simulation's compressed time.
	fireHarm = 0.6 // health per second at full intensity
	fireFear = 2.0 // fear per second at full intensity
	// Buildings catch from a fire burning on their tile, by how much of them
	// is wood, thatch and rope (see flammability).
	buildingCatch     = 3.0 // per second at full intensity, for a building of nothing but wood and thatch
	buildingCatchFrom = 0.3 // weaker flames don't take a building
	buildingFlare     = 0.9 // a burning building burns this hot on its tile (× its flammability) …
	buildingEmbers    = 0.6 // … and throws embers onto the eight tiles around with this strength
	// Hearths. The houses of people who can keep a fire have a cooking
	// hearth; a kiln burns hotter and longer. In dry, windy weather an ember
	// can blow onto dry grass a few metres away: open hearths in thatched
	// villages are a known cause of grass and village fires. An estimate: an
	// ember lands in the vegetation around a hearth about every four seconds
	// (half a year) in a fresh breeze; whether it catches depends on the
	// fuel there, so villages on damp riverbanks seldom burn. On the starter
	// island that comes to roughly one fire a year per 150 hearths.
	hearthSpark = 0.25 // embers a second a hearth throws in a fresh breeze …
	hearthWind  = 0.3  // … this strong; the chance grows with the wind squared
	emberFlight = 3    // tiles downwind an ember may land, at most
	kilnHearth  = 3.0  // a kiln throws three times as many
	sparkHeat   = 0.5  // a spark is a weak flame: dry fuel catches at half its usual readiness
	// Trees in a fire lose wood: a fierce fire takes most of a tree's 8
	// units, and the woodland grows back over decades.
	treeBurn = 6.0 // wood units per second at full intensity
	// Sight. Fire makes its own light, so the night doesn't hide it, and the
	// smoke rises above trees and hills: within a person's sight the flames
	// need a clear line, beyond it the smoke column shows over anything, out
	// to twice their sight. Its heat is felt close by whichever way one faces.
	fireSight      = 2.0
	fireFeel       = 1.5
	fireCandidates = 6 // nearest flames tested for a clear line of sight
	fireEventGap   = 0.5
)

// Salts for the fire's hashed chances.
const (
	saltBuildingCatch uint64 = 0xF1AE + iota
	saltEmber
	saltSpark
	saltSparkDir
	saltSparkCatch
	saltSparkFar
)

// FireInfo summarises fire and wind for the world info.
type FireInfo struct {
	Burning   int     `json:"burning"`   // tiles alight now
	Scorched  int     `json:"scorched"`  // tiles blackened by fire
	Started   int     `json:"started"`   // fires started, all-time …
	Lightning int     `json:"lightning"` // … by lightning
	Hearth    int     `json:"hearth"`    // … by sparks from hearths and kilns
	Tiles     int     `json:"tiles"`     // tiles that caught, all-time
	Buildings int     `json:"buildings"` // buildings burnt down
	Crops     int     `json:"crops"`     // plots burnt
	Burned    int     `json:"burned"`    // people burned to death
	WindDir   float64 `json:"windDir"`   // radians, where the wind blows towards (0 east)
	Wind      float64 `json:"wind"`      // 0–1
	Storm     bool    `json:"storm"`
	Storms    int     `json:"storms"` // storms so far
}

// fireEffects applies the fire to people, fields and buildings; called once
// per step after the ecology ticks.
func (s *Sim) fireEffects() {
	if s.opts.NoFire {
		return
	}
	s.hearthSparks()
	burning := s.eco.Burning()
	if len(burning) == 0 {
		return
	}
	// The ecology appends to its list when buildings set more tiles alight;
	// this step deals with the tiles that burned when it began.
	burning = burning[:len(burning):len(burning)]
	s.burnPeople()
	s.burnBuildings(burning)
	s.burnTrees(burning)
	if s.tick%TicksPerSecond == 0 {
		w := s.terrain.w
		s.eco.FireClusters(func(tile int, heat float32) {
			s.addStimulus(stimFire, float64(tile%w)+0.5, float64(tile/w)+0.5, float64(heat), nil)
		})
	}
}

// burnPeople hurts and frightens everyone standing in a fire.
func (s *Sim) burnPeople() {
	for _, c := range s.creatures {
		if c.Health <= 0 {
			continue
		}
		i, ok := s.terrain.indexAt(c.X, c.Y)
		if !ok {
			continue
		}
		h := float64(s.eco.Heat(i))
		if h <= 0 {
			continue
		}
		c.Health -= fireHarm * h * dt
		s.feel(c, moodFear, fireFear*h*dt)
		if c.Health <= 0 && c.fate == "" {
			c.fate = "burned"
		}
	}
}

// flammability is the share of a building that burns: its wood, thatch and
// rope against its stone, brick, clay, glass and metal. A hut of wood and
// fibre burns entirely, a brick house only its roof timbers, a kiln not at
// all. Fields, ditches and wells have nothing to burn (a field's crop burns
// in the ecology).
func flammability(k chem.StructureKind) float64 {
	if k.Farm || k.Irrigation || k.Well {
		return 0
	}
	organic, total := 0, 0
	for id, n := range k.Cost {
		total += n
		switch id {
		case "kayu", "serat", "tali":
			organic += n
		}
	}
	if total == 0 {
		return 0
	}
	return float64(organic) / float64(total)
}

// burnBuildings lets the buildings on burning tiles catch and burn down.
func (s *Sim) burnBuildings(burning []int32) {
	var gone []*Structure
	seed := s.eco.Seed()
	for _, t := range burning {
		st := s.structAt[t]
		if st == nil {
			continue
		}
		flam := flammability(st.kind)
		h := float64(s.eco.Heat(int(t)))
		if flam <= 0 || h < buildingCatchFrom || hash01(st.ID, s.tick, saltBuildingCatch^seed) >= buildingCatch*flam*h*dt {
			continue
		}
		gone = append(gone, st)
	}
	if len(gone) == 0 {
		return
	}
	for _, st := range gone {
		s.burnDown(st)
	}
	s.loseStructures(gone)
	s.reportBurnt(gone)
}

// burnDown sets a building's tile ablaze and throws its embers around:
// onto the fuel next to it, and onto the roofs of the buildings next door.
func (s *Sim) burnDown(st *Structure) {
	flam := flammability(st.kind)
	seed := s.eco.Seed()
	if i, ok := s.terrain.index(st.X, st.Y); ok {
		s.eco.Flare(i, buildingFlare*flam)
	}
	for k, d := range [8][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}, {1, 1}, {1, -1}, {-1, 1}, {-1, -1}} {
		j, ok := s.terrain.index(st.X+d[0], st.Y+d[1])
		if !ok {
			continue
		}
		roll := hash01(st.ID*8+int64(k), s.tick, saltEmber^seed)
		if nb := s.structAt[j]; nb != nil && nb != st {
			if roll < buildingEmbers*flam*flammability(nb.kind) {
				s.eco.Flare(j, buildingEmbers*flam)
			}
			continue
		}
		s.eco.Ignite(j, buildingEmbers*flam, roll, ecology.FireSpread)
	}
	s.addStimulus(stimCollapse, float64(st.X)+0.5, float64(st.Y)+0.5, 1, nil)
	s.eco.FireStats().Buildings++
}

// loseStructures removes buildings that burnt down, and everything stored in
// them: the household is left without a home, its granary, pen or snare.
func (s *Sim) loseStructures(gone []*Structure) {
	lost := make(map[int64]bool, len(gone))
	for _, st := range gone {
		lost[st.ID] = true
	}
	kept := s.structures[:0]
	for _, st := range s.structures {
		if !lost[st.ID] {
			kept = append(kept, st)
		}
	}
	clear(s.structures[len(kept):])
	s.structures = kept
	for _, st := range gone {
		delete(s.structByID, st.ID)
		if i, ok := s.terrain.index(st.X, st.Y); ok && s.structAt[i] == st {
			s.structAt[i] = nil
		}
		s.structVersion++
	}
	for _, c := range s.creatures {
		if lost[c.HouseID] {
			c.HouseID = 0
		}
	}
	s.recomputeTier()
	s.applyFarms()
}

// reportBurnt tells the observer which buildings burnt, at most every half second.
func (s *Sim) reportBurnt(gone []*Structure) {
	if last, ok := s.lastEcoEvt["fire"]; ok && s.time()-last < fireEventGap {
		return
	}
	s.lastEcoEvt["fire"] = s.time()
	name := func(st *Structure) string {
		if st.Owner != 0 {
			return st.kind.Name + " milik " + st.OwnerName
		}
		return st.kind.Name + " tak bertuan"
	}
	if len(gone) == 1 {
		st := gone[0]
		var who int64
		if s.living(st.Owner) != nil {
			who = st.Owner
		}
		s.event("climate", name(st)+" terbakar", who)
		return
	}
	names := make([]string, 0, 3)
	for _, st := range gone[:min(3, len(gone))] {
		names = append(names, name(st))
	}
	list := strings.Join(names, ", ")
	if more := len(gone) - len(names); more > 0 {
		list += fmt.Sprintf(" dan %d lainnya", more)
	}
	s.event("climate", fmt.Sprintf("Kebakaran menghanguskan %d bangunan: %s", len(gone), list), 0)
}

// burnTrees burns the wood off trees standing in a fire.
func (s *Sim) burnTrees(burning []int32) {
	w := s.terrain.w
	for _, t := range burning {
		i := int(t)
		if h := float64(s.eco.Heat(i)); h >= 0.3 && s.eco.Canopy(i) {
			s.geo.Take(i%w, i/w, "kayu", treeBurn*h*dt)
		}
	}
}

// hearth is how much fire a building keeps burning: a cooking hearth in the
// house of a family that can keep a fire, a kiln's furnace; none elsewhere.
func (s *Sim) hearth(st *Structure) float64 {
	switch {
	case st.house():
		if owner := s.living(st.Owner); owner != nil && s.canPractise(owner, "api") {
			return 1
		}
	case st.Kind == "tungku":
		return kilnHearth
	}
	return 0
}

// hearthSparks lets hearths and kilns throw sparks onto the ground around
// them, each looked at once a second; in dry, windy weather one may catch.
func (s *Sim) hearthSparks() {
	if s.eco.FireDanger() <= 0 {
		return
	}
	dir, wind, _ := s.eco.Wind()
	rate := hearthSpark * (wind / hearthWind) * (wind / hearthWind)
	seed := s.eco.Seed()
	for _, st := range s.structures {
		if (s.tick+st.ID)%TicksPerSecond != 0 {
			continue
		}
		if hash01(st.ID, s.tick, saltSpark^seed) >= rate*kilnHearth {
			continue // not even a kiln would spark now: skip the closer look
		}
		h := s.hearth(st)
		if h <= 0 || hash01(st.ID, s.tick, saltSpark^seed) >= rate*h {
			continue
		}
		// The ember blows downwind, give or take a quarter turn either way,
		// and lands one to emberFlight tiles off.
		a := dir + (hash01(st.ID, s.tick, saltSparkDir^seed)-0.5)*math.Pi
		far := 1 + math.Floor(emberFlight*hash01(st.ID, s.tick, saltSparkFar^seed))
		j, ok := s.terrain.index(st.X+int(math.Round(far*math.Cos(a))), st.Y+int(math.Round(far*math.Sin(a))))
		if !ok || !s.eco.Ignite(j, sparkHeat, hash01(st.ID, s.tick, saltSparkCatch^seed), ecology.FireHearth) {
			continue
		}
		if last, ok := s.lastEcoEvt["fire-start"]; ok && s.time()-last < 2 {
			continue
		}
		s.lastEcoEvt["fire-start"] = s.time()
		what := "dapur " + st.kind.Name
		if !st.house() {
			what = st.kind.Name
		}
		if st.Owner != 0 {
			what += " milik " + st.OwnerName
		}
		var who int64
		if s.living(st.Owner) != nil {
			who = st.Owner
		}
		s.event("climate", fmt.Sprintf("Percikan api dari %s membakar rumput kering di dekatnya", what), who)
	}
}

// senseFire fills inFireNear (how close the nearest fire in sight is) and
// inWind (how hard the wind blows).
func (s *Sim) senseFire(c *Creature, in *[NumInputs]float64) {
	if s.opts.NoFire {
		return
	}
	_, wind, _ := s.eco.Wind()
	in[inWind] = wind
	if len(s.eco.Burning()) == 0 {
		return
	}
	vision := c.Genome.Traits.Vision
	reach := fireSight * vision
	best := reach
	lookX, lookY := math.Cos(c.Heading), math.Sin(c.Heading)
	all := s.opts.NoPerception
	type flame struct{ d, x, y float64 }
	var near [fireCandidates]flame
	n := 0
	w := s.terrain.w
	s.eco.FireNear(c.X, c.Y, reach, func(tile int, _ float32) {
		x, y := float64(tile%w)+0.5, float64(tile/w)+0.5
		dx, dy := x-c.X, y-c.Y
		d := math.Hypot(dx, dy)
		if d >= best {
			return
		}
		if all || d <= fireFeel {
			best = d
			return
		}
		if dx*lookX+dy*lookY < d*0.2588190451 {
			return // outside the field of view (75° either side)
		}
		if d > vision {
			best = d // the smoke over the trees
			return
		}
		// Flames within sight: keep the nearest few to test for a clear line.
		k := min(n, fireCandidates-1)
		if n == fireCandidates && d >= near[k].d {
			return
		}
		for k > 0 && near[k-1].d > d {
			near[k] = near[k-1]
			k--
		}
		near[k] = flame{d, x, y}
		n = min(n+1, fireCandidates)
	})
	for _, f := range near[:n] {
		if f.d >= best {
			break
		}
		if s.lineOfSight(c.X, c.Y, f.x, f.y) {
			best = f.d
			break
		}
	}
	if best < reach {
		in[inFireNear] = 1 - best/reach
	}
}

// appendFires writes the frame's burning tiles, `,"f":[[x,y,intensity],…]`.
func (s *Sim) appendFires(b []byte, view *Viewport) []byte {
	b = append(b, `,"f":[`...)
	w := s.terrain.w
	first := true
	for _, t := range s.eco.Burning() {
		x, y := int(t)%w, int(t)/w
		if !view.contains(float64(x)+0.5, float64(y)+0.5) {
			continue
		}
		if !first {
			b = append(b, ',')
		}
		first = false
		b = append(b, '[')
		b = strconv.AppendInt(b, int64(x), 10)
		b = append(b, ',')
		b = strconv.AppendInt(b, int64(y), 10)
		b = append(b, ',')
		b = strconv.AppendFloat(b, float64(s.eco.Heat(int(t))), 'f', 2, 32)
		b = append(b, ']')
	}
	return append(b, ']')
}

// appendWindTail adds the wind to the weather array: `,direction,speed,storm`
// (radians, 0–1, 0 or 1).
func (s *Sim) appendWindTail(b []byte) []byte {
	dir, speed, storm := s.eco.Wind()
	b = append(b, ',')
	b = strconv.AppendFloat(b, dir, 'f', 2, 64)
	b = append(b, ',')
	b = strconv.AppendFloat(b, speed, 'f', 2, 64)
	if storm {
		return append(b, ",1"...)
	}
	return append(b, ",0"...)
}

// Burnt returns the encoded `burnt` stream message (scorched tiles and how
// burnt, `{"v":version,"t":[[x,y,level],…]}`) and its version.
func (s *Sim) Burnt() ([]byte, int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v := s.eco.BurntVersion()
	w := s.terrain.w
	b := make([]byte, 0, 32+16*len(s.eco.Burning()))
	b = append(b, `{"v":`...)
	b = strconv.AppendInt(b, v, 10)
	b = append(b, `,"t":[`...)
	first := true
	s.eco.Scorched(func(tile int, level float32) {
		if !first {
			b = append(b, ',')
		}
		first = false
		b = append(b, '[')
		b = strconv.AppendInt(b, int64(tile%w), 10)
		b = append(b, ',')
		b = strconv.AppendInt(b, int64(tile/w), 10)
		b = append(b, ',')
		b = strconv.AppendFloat(b, float64(level), 'f', 2, 32)
		b = append(b, ']')
	})
	return append(b, "]}"...), v
}

// BurntVersion changes whenever the scorched land does.
func (s *Sim) BurntVersion() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.eco.BurntVersion()
}

func (s *Sim) fireInfo() FireInfo {
	st := s.eco.FireStats()
	dir, speed, storm := s.eco.Wind()
	return FireInfo{
		Burning:   len(s.eco.Burning()),
		Scorched:  s.eco.ScorchedCount(),
		Started:   st.Started,
		Lightning: st.Lightning,
		Hearth:    st.Hearth,
		Tiles:     st.Tiles,
		Buildings: st.Buildings,
		Crops:     st.Crops,
		Burned:    s.deathsBy.Burned,
		WindDir:   math.Round(dir*100) / 100,
		Wind:      math.Round(speed*100) / 100,
		Storm:     storm,
		Storms:    s.eco.Climate().Storms,
	}
}
