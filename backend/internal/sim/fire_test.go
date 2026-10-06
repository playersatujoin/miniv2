package sim

import (
	"bytes"
	"encoding/json"
	"math"
	"strings"
	"testing"

	"miniv2/backend/internal/chem"
	"miniv2/backend/internal/ecology"
)

// dryWorld is a fake-chemistry world in the middle of a parched dry season
// (July, an El Niño year's start aside), without wild animals, plus a grass
// tile far from water.
func dryWorld(t *testing.T, opts Options) (*Sim, int, int) {
	t.Helper()
	m := testMap(t, 48)
	s := newSimWith(m, 1, fakeCatalog())
	s.opts = opts
	s.tick = int64((2 + 0.55) * SecondsPerYear * TicksPerSecond)
	eo := opts.ecology()
	eo.NoFauna = true
	s.eco = ecology.New(ecology.LandFromMap(m), s.rng, eo, s.time())
	for _, i := range s.terrain.walkable {
		x, y := int(i)%s.terrain.w, int(i)/s.terrain.w
		ok := m.Layers.Ground[i] == 3 && m.Layers.Objects[i] == 0 // open grass
		for dy := -3; dy <= 3 && ok; dy++ {
			for dx := -3; dx <= 3; dx++ {
				if j, in := s.terrain.index(x+dx, y+dy); !in || s.terrain.blocked[j] || s.terrain.nearWater[j] {
					ok = false
					break
				}
			}
		}
		if ok {
			return s, x, y
		}
	}
	t.Fatal("no open grass on the test map")
	return nil, 0, 0
}

// burnFor runs the land, the fire and its effects for seconds; people with
// blank brains stay where they are.
func burnFor(s *Sim, seconds float64, each func()) {
	for range int(seconds * TicksPerSecond) {
		s.tick++
		s.eco.Tick(s.tick, s.time(), dt, s)
		if each != nil {
			each()
		}
		s.fireEffects()
		s.reap()
	}
}

func TestAHouseInAFireBurnsDown(t *testing.T) {
	s, x, y := dryWorld(t, Options{})
	owner := person(s, Male, "Budi", x+5, y)
	k := s.cat.structure["gubuk"]
	h := s.addStructure(k, x, y, owner)
	s.moveIn(owner, h)
	h.Storage.add("kayu", 5)
	tile, _ := s.terrain.index(x, y)
	before := s.structVersion
	if !s.eco.Flare(tile, 1) {
		t.Fatal("the house's tile can't burn")
	}
	burnFor(s, 1, nil)
	if s.structByID[h.ID] != nil || s.structAt[tile] != nil || len(s.structures) != 0 {
		t.Fatal("the hut is still standing in the fire")
	}
	if owner.HouseID != 0 || s.houseOf(owner) != nil || s.structVersion == before {
		t.Fatalf("bookkeeping: house id %d, version %d → %d", owner.HouseID, before, s.structVersion)
	}
	if !strings.Contains(eventsText(s), "Gubuk milik Budi terbakar") {
		t.Fatalf("no event for the burnt hut: %s", eventsText(s))
	}
	if s.eco.FireStats().Buildings != 1 || s.fireInfo().Buildings != 1 {
		t.Fatalf("fire stats %+v", *s.eco.FireStats())
	}
	// A kiln of stone and clay doesn't burn.
	kiln := s.addStructure(s.cat.structure["tungku"], x+2, y+2, owner)
	kt, _ := s.terrain.index(x+2, y+2)
	for range 20 {
		s.eco.Flare(kt, 1)
		burnFor(s, 0.1, nil)
	}
	if s.structByID[kiln.ID] == nil {
		t.Fatal("a stone kiln burnt down")
	}
	if flammability(chem.StructureKind{Cost: map[chem.ItemID]int{"kayu": 6, "serat": 4}, House: true}) != 1 ||
		flammability(chem.StructureKind{Cost: map[chem.ItemID]int{"batu": 4, "kayu": 2}, Irrigation: true}) != 0 {
		t.Fatal("flammability of huts or ditches")
	}
}

func TestPeopleInFireAreHurtAndCanBurnToDeath(t *testing.T) {
	s, x, y := dryWorld(t, Options{})
	c := person(s, Female, "Sari", x, y)
	far := person(s, Male, "Tono", x+3, y+3)
	var cause string
	s.onDeath = func(d *Creature, why string) {
		if d == c {
			cause = why
		}
	}
	tile, _ := s.terrain.index(x, y)
	burnFor(s, 0.2, func() { s.eco.Flare(tile, 1) })
	if c.Health >= 1 || c.Health <= 0 {
		t.Fatalf("health %.2f after a moment in the fire", c.Health)
	}
	burnFor(s, 3, func() { s.eco.Flare(tile, 1) })
	if cause != "burned" || s.deathsBy.Burned != 1 || s.byID[c.ID] != nil {
		t.Fatalf("death cause %q, burned %d", cause, s.deathsBy.Burned)
	}
	if far.Health < 1 && s.eco.Heat(s.tileOf(far)) == 0 {
		t.Fatal("someone out of the fire was hurt")
	}
	if !strings.Contains(eventsText(s), "tewas terbakar") || s.fireInfo().Burned != 1 {
		t.Fatalf("events: %s", eventsText(s))
	}
}

func (s *Sim) tileOf(c *Creature) int {
	i, _ := s.terrain.indexAt(c.X, c.Y)
	return i
}

func TestFireAndWindInTheFrame(t *testing.T) {
	s, x, y := dryWorld(t, Options{})
	burnFor(s, 0.5, nil)
	var raw struct {
		W []float64   `json:"w"`
		F [][]float64 `json:"f"`
	}
	if err := json.Unmarshal(s.encodeFrame(), &raw); err != nil {
		t.Fatal(err)
	}
	if len(raw.W) != 8 || raw.W[6] < 0 || raw.W[6] > 1 || (raw.W[7] != 0 && raw.W[7] != 1) || len(raw.F) != 0 {
		t.Fatalf("weather %v, fires %v", raw.W, raw.F)
	}
	dir, wind, _ := s.eco.Wind()
	if math.Abs(raw.W[5]-dir) > 0.006 || math.Abs(raw.W[6]-wind) > 0.006 {
		t.Fatalf("wind %v in the frame, %.3f %.3f in the world", raw.W[5:], dir, wind)
	}
	a, _ := s.terrain.index(x, y)
	b, _ := s.terrain.index(x+3, y)
	s.eco.Flare(a, 0.8)
	s.eco.Flare(b, 0.6)
	if err := json.Unmarshal(s.encodeFrame(), &raw); err != nil {
		t.Fatal(err)
	}
	if len(raw.F) != 2 || len(raw.F[0]) != 3 || raw.F[0][0] != float64(x) || raw.F[0][1] != float64(y) || raw.F[0][2] != 0.8 {
		t.Fatalf("fires %v", raw.F)
	}
	view := &Viewport{MinX: float64(x + 2), MinY: float64(y - 1), MaxX: float64(x + 5), MaxY: float64(y + 2)}
	if err := json.Unmarshal(s.encodeViewFrame(view), &raw); err != nil {
		t.Fatal(err)
	}
	if len(raw.F) != 1 || raw.F[0][0] != float64(x+3) {
		t.Fatalf("viewport fires %v", raw.F)
	}
	// The scorched land goes out with its version.
	v0 := s.BurntVersion()
	burnFor(s, 1, nil)
	data, v := s.Burnt()
	var burnt struct {
		V int64       `json:"v"`
		T [][]float64 `json:"t"`
	}
	if err := json.Unmarshal(data, &burnt); err != nil {
		t.Fatal(err)
	}
	if v == v0 || burnt.V != v || v != s.BurntVersion() || len(burnt.T) == 0 || len(burnt.T[0]) != 3 || burnt.T[0][2] < 0.05 {
		t.Fatalf("burnt %s (version %d → %d)", data, v0, v)
	}
}

func TestSeeingFireAndFeelingTheWind(t *testing.T) {
	s, x, y := dryWorld(t, Options{})
	c := person(s, Female, "Lihat", x, y)
	c.Heading = 0 // facing east
	in := &[NumInputs]float64{}
	s.senseFire(c, in)
	_, wind, _ := s.eco.Wind()
	if in[inFireNear] != 0 || in[inWind] != wind {
		t.Fatalf("no fire: fire %.2f, wind %.2f (%.2f)", in[inFireNear], in[inWind], wind)
	}
	east, _ := s.terrain.index(x+3, y)
	s.eco.Flare(east, 0.8)
	s.senseFire(c, in)
	if want := 1 - 3/(fireSight*c.Genome.Traits.Vision); math.Abs(in[inFireNear]-want) > 1e-9 {
		t.Fatalf("fire 3 tiles ahead: %.3f, want %.3f", in[inFireNear], want)
	}
	c.Heading = math.Pi // looking the other way
	in[inFireNear] = 0
	s.senseFire(c, in)
	if in[inFireNear] != 0 {
		t.Fatalf("saw a fire behind: %.2f", in[inFireNear])
	}
	// Standing next to it, the heat is felt whichever way one faces.
	c.X = float64(x+2) + 0.5
	s.senseFire(c, in)
	if in[inFireNear] <= 0 {
		t.Fatal("felt nothing beside a fire")
	}
	// Switched off, the senses stay dark.
	off, ox, oy := dryWorld(t, Options{NoFire: true})
	o := person(off, Male, "Gelap", ox, oy)
	in = &[NumInputs]float64{}
	off.senseFire(o, in)
	if in[inFireNear] != 0 || in[inWind] != 0 {
		t.Fatal("senses lit with fire switched off")
	}
}

func TestHearthsSparkFiresInDryWeather(t *testing.T) {
	s, x, y := dryWorld(t, Options{})
	owner := person(s, Male, "Api", x, y)
	owner.Skills = map[string]float64{"api": 1}
	for dy := -10; dy <= 10; dy += 2 {
		for dx := -10; dx <= 10; dx += 2 {
			if j, ok := s.terrain.index(x+dx, y+dy); ok && !s.terrain.blocked[j] && s.structAt[j] == nil {
				s.addStructure(s.cat.structure["gubuk"], x+dx, y+dy, owner)
			}
		}
	}
	if s.hearth(s.structures[0]) != 1 {
		t.Fatal("the house of someone who keeps a fire has no hearth")
	}
	// The weather held as in a dry July, with a breeze of about 5 m/s.
	for range 200 * TicksPerSecond {
		s.tick++
		s.hearthSparks()
		if s.eco.FireStats().Hearth > 0 {
			break
		}
	}
	if s.eco.FireStats().Hearth == 0 {
		t.Fatalf("no spark caught in 200 dry seconds from %d hearths", len(s.structures))
	}
	if !strings.Contains(eventsText(s), "Percikan api") {
		t.Fatalf("events: %s", eventsText(s))
	}
}

func TestNoFireSwitchesFireOff(t *testing.T) {
	s, x, y := dryWorld(t, Options{NoFire: true})
	owner := person(s, Male, "Budi", x, y)
	h := s.addStructure(s.cat.structure["gubuk"], x, y, owner)
	tile, _ := s.terrain.index(x, y)
	if s.eco.Flare(tile, 1) || s.eco.Ignite(tile, 1, 0, ecology.FireLightning) {
		t.Fatal("fire lit with fire switched off")
	}
	burnFor(s, 1, nil)
	if s.structByID[h.ID] == nil || owner.Health < 1 || !bytes.Contains(s.encodeFrame(), []byte(`"f":[]`)) {
		t.Fatal("something burned with fire switched off")
	}
	if _, w, _ := s.eco.Wind(); w <= 0 {
		t.Fatal("the wind stopped with fire")
	}
}

// A world saved while it burns carries on exactly as if it had not been.
func TestFireSaveRestoreReplaysExactly(t *testing.T) {
	m := testMap(t, 48)
	a := New(m, 3)
	// Into the dry season, then set the grass around the first couple alight.
	a.Advance(int(0.6 * SecondsPerYear * TicksPerSecond))
	if len(a.creatures) == 0 {
		t.Fatal("nobody alive")
	}
	c := a.creatures[0]
	h := a.addStructure(a.cat.structure["gubuk"], int(c.X)+1, int(c.Y), c)
	a.moveIn(c, h)
	lit := 0
	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			if j, ok := a.terrain.index(int(c.X)+dx, int(c.Y)+dy); ok && a.eco.Flare(j, 0.9) {
				lit++
			}
		}
	}
	if lit == 0 {
		t.Fatal("nothing would burn")
	}
	a.Advance(10)
	if len(a.eco.Burning()) == 0 {
		t.Fatal("the fire was out before the save")
	}
	data, err := a.MarshalState()
	if err != nil {
		t.Fatal(err)
	}
	b, err := Restore(m, data)
	if err != nil {
		t.Fatal(err)
	}
	a.Advance(3 * TicksPerSecond)
	b.Advance(3 * TicksPerSecond)
	snapshot := func(s *Sim) []byte {
		structs, _ := s.Structures()
		burnt, _ := s.Burnt()
		info, err := json.Marshal([]any{s.Info(), s.creatures})
		if err != nil {
			t.Fatal(err)
		}
		return bytes.Join([][]byte{s.encodeFrame(), structs, burnt, info}, nil)
	}
	if !bytes.Equal(snapshot(a), snapshot(b)) {
		t.Fatal("the restored world burned differently")
	}
	if a.eco.FireStats().Tiles <= lit {
		t.Fatalf("the fire never spread: %+v", *a.eco.FireStats())
	}
}
