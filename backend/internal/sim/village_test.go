package sim

import (
	"bytes"
	"encoding/json"
	"math"
	"slices"
	"strings"
	"testing"

	"miniv2/backend/internal/world"
)

// villageMap is the 64×64 test map cleared to open grass inside a two-tile
// rim, so houses and people stand anywhere and see each other.
func villageMap(t *testing.T) *world.Map {
	m := testMap(t, 64)
	for y := 2; y < 62; y++ {
		for x := 2; x < 62; x++ {
			m.Layers.Ground[y*64+x], m.Layers.Objects[y*64+x] = world.Grass, world.None
		}
	}
	return m
}

// villageWorld is an empty fake-chemistry world big enough for a few villages.
func villageWorld(t *testing.T) *Sim {
	t.Helper()
	return newSimWith(villageMap(t), 1, fakeCatalog())
}

// home builds a hut at (x, y) for a new adult man and moves him in.
func home(s *Sim, name string, x, y int) (*Creature, *Structure) {
	c := person(s, Male, name, x, y)
	h := s.addStructure(s.cat.structure["gubuk"], x, y, c)
	s.moveIn(c, h)
	return c, h
}

// lodger puts another adult into house h.
func lodger(s *Sim, name string, h *Structure) *Creature {
	c := person(s, Female, name, h.X, h.Y)
	c.HouseID = h.ID
	return c
}

// recount runs the next clustering.
func recount(s *Sim) {
	s.tick += villageEvery - s.tick%villageEvery
	s.villageTick()
}

// nextSecond runs the village check one second on (not a clustering).
func nextSecond(s *Sim) {
	s.tick += TicksPerSecond
	if s.tick%villageEvery == 0 {
		s.tick += TicksPerSecond
	}
	s.villageTick()
}

func villagesOf(s *Sim) []*village { return s.vil().list }

func eventsWith(s *Sim, sub string) int {
	n := 0
	for _, e := range s.events {
		if strings.Contains(e.Text, sub) {
			n++
		}
	}
	return n
}

func TestVillageNeedsThreeHousesWithinLink(t *testing.T) {
	s := villageWorld(t)
	home(s, "Adi", 10, 10)
	home(s, "Budi", 10+villageLink, 10) // exactly at the link distance
	recount(s)
	if len(villagesOf(s)) != 0 {
		t.Fatal("two huts with two people are not a village")
	}
	home(s, "Cahyo", 10+2*villageLink, 10) // links through Budi only
	far, _ := home(s, "Dodi", 10+3*villageLink+1, 10)
	recount(s)
	vs := villagesOf(s)
	if len(vs) != 1 || len(vs[0].Houses) != 3 || vs[0].People != 3 {
		t.Fatalf("want one village of 3 houses chained by %v tiles, got %+v", villageLink, vs)
	}
	if far.VillageID != 0 || s.byID[1].VillageID != vs[0].ID {
		t.Fatal("the house one tile past the link must stay outside")
	}
	if eventsWith(s, "terbentuk dari 3 rumah") != 1 {
		t.Fatalf("no founding report: %q", lastEvent(s))
	}
}

func TestTwoFullHousesMakeAVillage(t *testing.T) {
	s := villageWorld(t)
	_, a := home(s, "Adi", 20, 20)
	_, b := home(s, "Budi", 24, 20)
	for i := range 4 {
		lodger(s, "Sri", []*Structure{a, b}[i%2])
	}
	recount(s)
	if vs := villagesOf(s); len(vs) != 1 || vs[0].People != 6 {
		t.Fatalf("two houses with six people should be a village: %+v", vs)
	}
}

func TestVillageKeepsItsIdentity(t *testing.T) {
	s := villageWorld(t)
	for i := range 3 {
		home(s, "Adi", 10+4*i, 30)
	}
	recount(s)
	v := villagesOf(s)[0]
	id, name, founded := v.ID, v.Name, v.Founded
	home(s, "Eko", 22, 32)
	recount(s)
	recount(s)
	vs := villagesOf(s)
	if len(vs) != 1 || vs[0].ID != id || vs[0].Name != name || vs[0].Founded != founded || len(vs[0].Houses) != 4 {
		t.Fatalf("a new house must join the same village: %+v", vs)
	}
	if villageName(id, 0) != villageName(id, 0) || name == "" {
		t.Fatal("names are a function of the ID")
	}
}

func TestVillageMergeKeepsOlderAndSplitMakesNew(t *testing.T) {
	s := villageWorld(t)
	for i := range 3 {
		home(s, "Adi", 4+4*i, 40) // 4, 8, 12
	}
	recount(s)
	for i := range 3 {
		home(s, "Budi", 28+4*i, 40) // 28, 32, 36: 16 tiles from the first
	}
	recount(s)
	if len(villagesOf(s)) != 2 {
		t.Fatalf("want two villages, got %d", len(villagesOf(s)))
	}
	older, newer := villagesOf(s)[0], villagesOf(s)[1]
	_, bridge := home(s, "Cahyo", 20, 40) // within 8 of both
	recount(s)
	vs := villagesOf(s)
	if len(vs) != 1 || vs[0].ID != older.ID || len(vs[0].Houses) != 7 {
		t.Fatalf("a merger keeps the older village: %+v", vs)
	}
	if eventsWith(s, "Desa "+newer.Name+" bergabung dengan Desa "+older.Name) != 1 {
		t.Fatalf("no merger report: %q", lastEvent(s))
	}
	// The bridge is abandoned: two halves of three houses each. The first
	// half keeps the village, the other is a new one with a new ID.
	bridge.Owner, bridge.OwnerName = 0, ""
	recount(s)
	vs = villagesOf(s)
	if len(vs) != 2 || vs[0].ID != older.ID || vs[1].ID <= newer.ID {
		t.Fatalf("a split must make a new village: %+v", vs)
	}
	if !strings.Contains(lastEvent(s), "memisahkan diri dari Desa "+older.Name) {
		t.Fatalf("no split report: %q", lastEvent(s))
	}
	// Two houses keep a village going; one doesn't.
	for _, id := range vs[1].Houses[:2] {
		s.structByID[id].Owner = 0
	}
	recount(s)
	if len(villagesOf(s)) != 1 || eventsWith(s, "Desa "+vs[1].Name+" ditinggalkan") != 1 {
		t.Fatalf("a village down to one house is left: %d villages, %q", len(villagesOf(s)), lastEvent(s))
	}
}

func TestConvexHull(t *testing.T) {
	pts := []point{{0, 0}, {4, 0}, {4, 4}, {0, 4}, {2, 2}, {2, 0}, {4, 4}}
	h := convexHull(pts)
	if len(h) != 4 {
		t.Fatalf("hull of a square with inner, edge and repeated points: %v", h)
	}
	area := 0.0
	for i, a := range h {
		b := h[(i+1)%len(h)]
		area += a[0]*b[1] - b[0]*a[1]
	}
	if area <= 0 {
		t.Fatalf("hull must run counter-clockwise (positive area), got %v", area/2)
	}
	if !inHull(h, 2, 2) || !inHull(h, 0, 0) || inHull(h, 4.01, 2) || inHull(h, -0.01, 2) {
		t.Fatal("point in hull is wrong")
	}
	if one := landHull([]point{{5, 5}}); len(one) != 8 || !inHull(one, 5+landMargin*0.99, 5) {
		t.Fatalf("a single house still has land around it: %v", one)
	}
}

func TestVillageLandCoversHousesFieldsAndMargin(t *testing.T) {
	s := villageWorld(t)
	var houses []*Structure
	for i := range 3 {
		_, h := home(s, "Adi", 20+3*i, 20+2*i)
		houses = append(houses, h)
	}
	tungku := s.addStructure(s.cat.structure["tungku"], 20, 27, s.byID[1])
	// The first family's field, 7 tiles from their house, and a stranger's.
	field := 15*64 + 25
	if !s.eco.Plant(field, "ubi", 1, houses[0].ID, 0.5, s.time()) || !s.eco.Plant(15*64+26, "ubi", 99, 0, 0.5, s.time()) {
		t.Fatal("cannot plant on the test map")
	}
	recount(s)
	v := villagesOf(s)[0]
	if !slices.Contains(v.Buildings, tungku.ID) || v.Fields != 1 {
		t.Fatalf("the furnace and the family's field belong to the village: %v, %d fields", v.Buildings, v.Fields)
	}
	spots := []point{{float64(field%64) + 0.5, float64(field/64) + 0.5}}
	for _, st := range append(houses, tungku) {
		spots = append(spots, point{float64(st.X) + 0.5, float64(st.Y) + 0.5})
	}
	for _, p := range spots {
		cx, cy := p[0], p[1]
		for k := range 16 {
			a := float64(k) * math.Pi / 8
			if x, y := cx+landMargin*0.999*math.Cos(a), cy+landMargin*0.999*math.Sin(a); !inHull(v.Hull, x, y) {
				t.Fatalf("land misses (%.2f, %.2f), %v from (%v, %v)", x, y, landMargin, cx, cy)
			}
		}
	}
	if inHull(v.Hull, 20.5-landMargin*1.2-1, 20.5) {
		t.Fatal("land reaches too far")
	}
	// The per-tile grid says exactly what the polygons say.
	for _, h := range houses[:1] {
		home(s, "Jauh", h.X+40, h.Y+30)
		home(s, "Jauh", h.X+43, h.Y+31)
		home(s, "Jauh", h.X+40, h.Y+34)
	}
	recount(s)
	if len(villagesOf(s)) != 2 {
		t.Fatalf("want two villages, got %d", len(villagesOf(s)))
	}
	vs := s.vil()
	for i, id := range vs.land {
		x, y := float64(i%s.terrain.w)+0.5, float64(i/s.terrain.w)+0.5
		any := false
		for _, v := range vs.list {
			if inHull(v.Hull, x, y) {
				any = true
			}
		}
		if (id != 0) != any || id != 0 && !inHull(vs.byID[int64(id)].Hull, x, y) {
			t.Fatalf("tile (%v, %v): grid says %d, polygons say %v", x, y, id, any)
		}
		if got := s.landAt(x, y); got != int64(id) {
			t.Fatalf("landAt %d, grid %d", got, id)
		}
	}
	area := 0
	for _, v := range vs.list {
		area += v.area
	}
	if area == 0 {
		t.Fatal("land has no tiles")
	}
}

// trust makes a believe in b by w now.
func trust(s *Sim, a, b *Creature, w float64) {
	a.Relations = append(a.Relations, Relation{Person: Ref{b.ID, b.Name}, Trust: w, Tick: s.tick, Met: 1})
}

func TestLeaderIsTheMostTrustedAndChangesOnDeath(t *testing.T) {
	s := villageWorld(t)
	a, ha := home(s, "Adi", 10, 50)
	b, hb := home(s, "Budi", 14, 50)
	c, hc := home(s, "Cahyo", 18, 50)
	x := lodger(s, "Sari", ha) // lives with Adi
	y := lodger(s, "Ratna", hc)
	// Sari is trusted by Budi and Cahyo; Ratna by Adi and Budi, less.
	trust(s, b, x, 0.5)
	trust(s, c, x, 0.4)
	trust(s, a, y, 0.3)
	trust(s, b, y, 0.3)
	recount(s)
	v := villagesOf(s)[0]
	if v.Leader != x.ID || !s.isLeader(x) || s.isLeader(y) || v.Supporters != 2 {
		t.Fatalf("leader %d (supporters %d), want Sari %d", v.Leader, v.Supporters, x.ID)
	}
	if lastEvent(s) != "Sari menjadi pemimpin Desa "+v.Name {
		t.Fatalf("no leader report: %q", lastEvent(s))
	}
	if r := s.villageRef(b); r == nil || r.ID != v.ID || r.Name != v.Name {
		t.Fatalf("village ref %+v", r)
	}
	// A slightly more trusted rival does not unseat her …
	trust(s, c, y, 0.35)
	recount(s)
	if v.Leader != x.ID {
		t.Fatal("a leader keeps her place unless clearly outranked")
	}
	// … but her death passes leadership on within a second.
	x.Health = 0
	s.reap()
	nextSecond(s)
	if v.Leader != y.ID || !s.isLeader(y) {
		t.Fatalf("leader %d after Sari died, want Ratna %d", v.Leader, y.ID)
	}
	if !strings.Contains(lastEvent(s), "Ratna menjadi pemimpin Desa "+v.Name+", menggantikan mendiang Sari") {
		t.Fatalf("no succession report: %q", lastEvent(s))
	}
	_ = hb
}

func TestLeaderNeedsSupporters(t *testing.T) {
	s := villageWorld(t)
	a, ha := home(s, "Adi", 10, 50)
	b, _ := home(s, "Budi", 14, 50)
	home(s, "Cahyo", 18, 50)
	trust(s, b, a, 0.9)
	recount(s)
	if v := villagesOf(s)[0]; v.Leader != 0 {
		t.Fatal("one admirer makes no leader")
	}
	// A child's trust doesn't count …
	kid := lodger(s, "Ani", ha)
	kid.BornTick = s.tick
	trust(s, kid, a, 0.9)
	recount(s)
	if v := villagesOf(s)[0]; v.Leader != 0 {
		t.Fatal("a child is no supporter")
	}
	// … nor faint goodwill, but a second adult who trusts him does.
	w := lodger(s, "Sri", ha)
	trust(s, w, a, supportTrust/2)
	recount(s)
	if v := villagesOf(s)[0]; v.Leader != 0 {
		t.Fatal("faint goodwill is no support")
	}
	w.Relations[0].Trust = 0.3
	recount(s)
	if v := villagesOf(s)[0]; v.Leader != a.ID || v.Supporters != 2 {
		t.Fatalf("two supporters make a leader: %d (%d)", v.Leader, v.Supporters)
	}
	// Losing them (a quarrel, years later) ends it, though not at once: the
	// standing a leader has earned fades over a few years.
	s.tick += leaderEventGap * TicksPerSecond
	w.Relations[0].Trust, b.Relations[0].Trust = -0.5, -0.5
	w.Relations[0].Tick, b.Relations[0].Tick = s.tick, s.tick
	recount(s)
	if v := villagesOf(s)[0]; v.Leader != a.ID {
		t.Fatal("one bad season already ended a leader's standing")
	}
	elections := 1
	for ; elections < 4*standingYears*2 && villagesOf(s)[0].Leader != 0; elections++ {
		recount(s)
	}
	if v := villagesOf(s)[0]; v.Leader != 0 || !strings.Contains(lastEvent(s), "tanpa pemimpin") {
		t.Fatalf("a leader without support steps down: %d, %q", v.Leader, lastEvent(s))
	}
	t.Logf("stepped down after %d elections", elections)
}

func TestVillageSenses(t *testing.T) {
	s := villageWorld(t)
	a, ha := home(s, "Adi", 20, 20)
	b, _ := home(s, "Budi", 24, 20)
	c, _ := home(s, "Cahyo", 28, 20)
	l := lodger(s, "Sari", ha)
	trust(s, b, l, 0.5)
	trust(s, c, l, 0.5)
	recount(s)
	sense := func(p *Creature) [NumInputs]float64 {
		var in [NumInputs]float64
		s.senseVillage(p, &in)
		return in
	}
	if in := sense(l); in[inIsLeader] != 1 || in[inOwnVillage] != 1 || in[inLeaderNear] != 0 {
		t.Fatalf("the leader at home: %v %v %v", in[inIsLeader], in[inOwnVillage], in[inLeaderNear])
	}
	// Budi faces east (heading 0); Sari two tiles away, south of him (+y).
	b.X, b.Y, b.Heading = 24.5, 20.5, 0
	l.X, l.Y = 24.5, 22.5
	vision := b.Genome.Traits.Vision * (0.55 + 0.45*s.eco.Light())
	in := sense(b)
	if math.Abs(in[inLeaderNear]-(1-2/vision)) > 1e-9 || math.Abs(in[inLeaderSide]-1) > 1e-9 || in[inIsLeader] != 0 || in[inOwnVillage] != 1 {
		t.Fatalf("near %v side %v (vision %v)", in[inLeaderNear], in[inLeaderSide], vision)
	}
	l.Y = 18.5 // north: the other side
	if in := sense(b); math.Abs(in[inLeaderSide]+1) > 1e-9 {
		t.Fatalf("side %v, want -1", in[inLeaderSide])
	}
	l.X, l.Y = 24.5+vision+0.5, 20.5 // out of sight
	if in := sense(b); in[inLeaderNear] != 0 || in[inLeaderSide] != 0 {
		t.Fatal("a leader out of sight is not felt")
	}
	b.X, b.Y = 55.5, 55.5 // far from home
	if in := sense(b); in[inOwnVillage] != 0 {
		t.Fatal("outside the land")
	}
	stranger := person(s, Male, "Tamu", 24, 20)
	if in := sense(stranger); in != ([NumInputs]float64{}) || a.VillageID == 0 {
		t.Fatal("someone without a village senses nothing of it")
	}
}

func TestVillagesPayload(t *testing.T) {
	s := villageWorld(t)
	if data, v := s.Villages(); string(data) != `{"v":0,"villages":[]}` || v != 0 {
		t.Fatalf("empty payload %s", data)
	}
	_, ha := home(s, "Adi", 20, 20)
	b, _ := home(s, "Budi", 23, 21)
	c, _ := home(s, "Cahyo", 26, 23)
	l := lodger(s, "Sari", ha)
	trust(s, b, l, 0.5)
	trust(s, c, l, 0.5)
	recount(s)
	data, version := s.Villages()
	if version != s.VillageVersion() || version == 0 {
		t.Fatalf("version %d / %d", version, s.VillageVersion())
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	keys := func(m map[string]json.RawMessage) []string {
		var out []string
		for k := range m {
			out = append(out, k)
		}
		slices.Sort(out)
		return out
	}
	if got := keys(raw); !slices.Equal(got, []string{"v", "villages"}) {
		t.Fatalf("top-level keys %v", got)
	}
	var list []map[string]json.RawMessage
	if err := json.Unmarshal(raw["villages"], &list); err != nil || len(list) != 1 {
		t.Fatalf("villages %s: %v", raw["villages"], err)
	}
	want := []string{"founded", "houses", "hue", "hull", "id", "leader", "name", "people", "x", "y"}
	if got := keys(list[0]); !slices.Equal(got, want) {
		t.Fatalf("village keys %v, want %v", got, want)
	}
	var v struct {
		ID      int64        `json:"id"`
		Name    string       `json:"name"`
		Hue     int          `json:"hue"`
		Hull    [][2]float64 `json:"hull"`
		X, Y    float64
		People  int  `json:"people"`
		Houses  int  `json:"houses"`
		Leader  *Ref `json:"leader"`
		Founded float64
	}
	if err := json.Unmarshal(data[strings.Index(string(data), "[")+1:len(data)-2], &v); err != nil {
		t.Fatal(err)
	}
	if v.People != 4 || v.Houses != 3 || v.Leader == nil || v.Leader.Name != "Sari" || v.Hue < 0 || v.Hue >= 360 || len(v.Hull) < 3 {
		t.Fatalf("payload %s", data)
	}
	for _, p := range append(v.Hull, [2]float64{v.X, v.Y}) {
		for _, q := range p {
			if math.Abs(q*10-math.Round(q*10)) > 1e-9 {
				t.Fatalf("coordinate %v has more than one decimal", q)
			}
		}
	}
	views := s.VillagesView()
	if len(views) != 1 || len(views[0].Residents) != 4 || views[0].Leader == nil || views[0].Area == 0 || len(views[0].HouseIDs) != 3 {
		t.Fatalf("view %+v", views)
	}
	if _, err := json.Marshal(views); err != nil {
		t.Fatal(err)
	}
	if info := s.villageInfo(); info.Villages != 1 || info.Villagers != 4 || info.Led != 1 || info.Founded != 1 {
		t.Fatalf("info %+v", info)
	}
}

func TestVillageVersionMovesOnlyOnChange(t *testing.T) {
	s := villageWorld(t)
	for i := range 3 {
		home(s, "Adi", 10+4*i, 10)
	}
	recount(s)
	v0 := s.VillageVersion()
	nextSecond(s)
	recount(s)
	nextSecond(s)
	if s.VillageVersion() != v0 {
		t.Fatal("nothing changed, yet the version moved")
	}
	lodger(s, "Sri", s.structByID[1])
	nextSecond(s)
	if s.VillageVersion() == v0 {
		t.Fatal("a new resident must move the version")
	}
}

func TestVillagesSaveAndRestore(t *testing.T) {
	m := villageMap(t)
	a := newSimWith(m, 1, fakeCatalog())
	for i := range 3 {
		home(a, "Adi", 10+4*i, 10)
	}
	_, h := home(a, "Budi", 40, 40)
	home(a, "Cahyo", 44, 40)
	x := lodger(a, "Sari", h)
	for range 4 {
		lodger(a, "Ani", h)
	}
	trust(a, a.byID[4], x, 0.5) // Budi, her household
	trust(a, a.byID[5], x, 0.5) // Cahyo, next door
	recount(a)
	nextSecond(a)
	data, err := a.MarshalState()
	if err != nil {
		t.Fatal(err)
	}
	b, err := restoreWith(m, data, fakeCatalog())
	if err != nil {
		t.Fatal(err)
	}
	same := func(when string) {
		t.Helper()
		pa, va := a.Villages()
		pb, vb := b.Villages()
		if !bytes.Equal(pa, pb) || va != vb {
			t.Fatalf("%s: payloads differ:\n%s\n%s", when, pa, pb)
		}
		if !slices.Equal(a.vil().land, b.vil().land) {
			t.Fatalf("%s: land grids differ", when)
		}
		for i, ca := range a.creatures {
			cb := b.creatures[i]
			var ia, ib [NumInputs]float64
			a.senseVillage(ca, &ia)
			b.senseVillage(cb, &ib)
			if ca.VillageID != cb.VillageID || ia != ib || a.isLeader(ca) != b.isLeader(cb) {
				t.Fatalf("%s: %s differs", when, ca.Name)
			}
		}
		if a.villageInfo() != b.villageInfo() {
			t.Fatalf("%s: info differs", when)
		}
	}
	same("restored")
	for _, s := range []*Sim{a, b} {
		home(s, "Dodi", 22, 12)
		recount(s)
		s.byID[x.ID].Health = 0
		s.reap()
		nextSecond(s)
		recount(s)
	}
	same("carried on")
	if len(a.vil().list) != 2 {
		t.Fatalf("want two villages, got %d", len(a.vil().list))
	}
}

func TestNoVillages(t *testing.T) {
	s := villageWorld(t)
	s.opts.NoVillages = true
	var people []*Creature
	for i := range 4 {
		c, _ := home(s, "Adi", 10+3*i, 10)
		people = append(people, c)
	}
	trust(s, people[1], people[0], 0.5)
	trust(s, people[2], people[0], 0.5)
	recount(s)
	nextSecond(s)
	if len(villagesOf(s)) != 0 || s.villageInfo() != (VillageInfo{}) {
		t.Fatal("no villages with the switch off")
	}
	for _, c := range people {
		var in [NumInputs]float64
		s.senseVillage(c, &in)
		if c.VillageID != 0 || in != ([NumInputs]float64{}) || s.isLeader(c) || s.villageRef(c) != nil {
			t.Fatal("no village membership, leaders or senses with the switch off")
		}
	}
	if data, _ := s.Villages(); string(data) != `{"v":0,"villages":[]}` {
		t.Fatalf("payload %s", data)
	}
	// Switching off a world that had villages forgets them.
	s.opts.NoVillages = false
	recount(s)
	if len(villagesOf(s)) != 1 || people[0].VillageID == 0 {
		t.Fatal("villages with the switch on")
	}
	s.opts.NoVillages = true
	nextSecond(s)
	if len(villagesOf(s)) != 0 || people[0].VillageID != 0 {
		t.Fatal("villages remain after switching off")
	}
}

func TestVillageNames(t *testing.T) {
	seen := map[string]bool{}
	for id := int64(1); id <= 200; id++ {
		n := villageName(id, 0)
		if len(n) < 4 || len(n) > 16 || n[0] < 'A' || n[0] > 'Z' || strings.ContainsAny(n, " -'") {
			t.Fatalf("odd village name %q", n)
		}
		seen[n] = true
	}
	if len(seen) < 190 {
		t.Fatalf("only %d distinct names in 200", len(seen))
	}
	if villageHue(1) == villageHue(2) || villageHue(7) < 0 || villageHue(7) >= 360 {
		t.Fatal("hues")
	}
}

func TestVillageNewsIsRationed(t *testing.T) {
	s := villageWorld(t)
	for i := range 6 {
		x, y := 6+20*(i%3), 6+20*(i/3)
		for k := range 3 {
			home(s, "Adi", x+3*k, y)
		}
	}
	before := len(s.events)
	recount(s)
	if len(villagesOf(s)) != 6 {
		t.Fatalf("want six villages, got %d", len(villagesOf(s)))
	}
	if got := len(s.events) - before; got != villageNewsPerTick+1 || lastEvent(s) != "Ada 2 perubahan desa lainnya" {
		t.Fatalf("%d reports, last %q", got, lastEvent(s))
	}
}

// A real world that grows villages carries on exactly after a save.
func TestVillagesReplayAfterRestore(t *testing.T) {
	if testing.Short() {
		t.Skip("runs several simulated minutes")
	}
	island := func() *world.Map {
		m := world.Generate(world.GenOptions{Name: "Starter Island", Width: 128, Height: 128, Seed: 1337})
		m.ID = "0000000000000000"
		return m
	}
	a := New(island(), 4)
	a.Advance(6*60*TicksPerSecond + 7) // between village checks
	if len(a.vil().list) == 0 {
		t.Skip("no village on this seed with the current rules")
	}
	data, err := a.MarshalState()
	if err != nil {
		t.Fatal(err)
	}
	b, err := Restore(island(), data)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range []*Sim{a, b} {
		s.Advance(2 * 60 * TicksPerSecond)
	}
	pa, _ := a.Villages()
	pb, _ := b.Villages()
	va, _ := json.Marshal(a.VillagesView())
	vb, _ := json.Marshal(b.VillagesView())
	if !bytes.Equal(pa, pb) || !bytes.Equal(va, vb) || !bytes.Equal(a.encodeFrame(), b.encodeFrame()) {
		t.Fatalf("restored villages diverged:\n%s\n%s", pa, pb)
	}
	t.Logf("%s", pa)
}

func TestLeaderByReputationWithoutPerception(t *testing.T) {
	s := villageWorld(t)
	s.opts.NoPerception = true
	a, _ := home(s, "Adi", 10, 50)
	b, _ := home(s, "Budi", 14, 50)
	home(s, "Cahyo", 18, 50)
	a.Reputation, b.Reputation = 0.3, 0.5
	recount(s)
	if v := villagesOf(s)[0]; v.Leader != b.ID || v.Supporters != 2 {
		t.Fatalf("without personal memories the best reputation leads: %d", v.Leader)
	}
}
