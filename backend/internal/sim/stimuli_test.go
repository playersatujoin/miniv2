package sim

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"math"
	"math/rand/v2"
	"os"
	"slices"
	"strconv"
	"testing"

	"miniv2/backend/internal/world"
)

func TestStimuliMergeExpireAndStayBounded(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	fx, fy := float64(x)+0.5, float64(y)+0.5
	w := s.stims()

	s.addStimulus(stimFight, fx, fy, 1, nil)
	first := w.list[0]
	s.tick += TicksPerSecond
	s.addStimulus(stimFight, fx+1, fy, 1, nil) // the same fight, a step away
	if len(w.list) != 1 || first.Tick != s.tick || first.Until != s.tick+int64(stimKinds[stimFight].life*TicksPerSecond) {
		t.Fatalf("near duplicate did not refresh the live fight: %d stimuli", len(w.list))
	}
	if first.X != fx {
		t.Fatal("a refreshed stimulus moved (the sensing grid would go stale)")
	}
	s.addStimulus(stimFight, fx+5, fy, 1, nil)
	s.addStimulus(stimTheft, fx, fy, 1, nil) // another kind never merges
	if len(w.list) != 3 {
		t.Fatalf("got %d stimuli, want 3", len(w.list))
	}
	c := person(s, Male, "Gone", x, y)
	s.leaveCorpse(c, "killed")
	s.leaveCorpse(c, "killed")
	if len(w.list) != 5 {
		t.Fatal("bodies merged")
	}

	// Everything fades: the theft first, the bodies last.
	s.tick = w.list[2].Until
	s.stimuliTick()
	if len(w.list) != 4 || slices.ContainsFunc(w.list, func(st *Stimulus) bool { return st.Kind == stimTheft }) {
		t.Fatal("faded theft kept")
	}
	s.tick += 10 * TicksPerSecond
	s.stimuliTick()
	if len(w.list) != 0 || s.stimuliInfo().Active != 0 {
		t.Fatal("stimuli outlived their lifetime")
	}

	// A full list: equals give no way, a graver kind takes the oldest of the
	// least grave.
	for i := range maxStimuli {
		s.addStimulus(stimTheft, float64(2*(i%24))+0.5, float64(2*(i/24))+0.5, 1, nil)
	}
	if len(w.list) != maxStimuli {
		t.Fatalf("filled %d of %d", len(w.list), maxStimuli)
	}
	oldest := w.list[0]
	s.addStimulus(stimTheft, 47.5, 47.5, 1, nil)
	if len(w.list) != maxStimuli || w.list[len(w.list)-1].X == 47.5 {
		t.Fatal("a full list grew or let an equal push in")
	}
	s.addStimulus(stimPredator, 47.5, 47.5, 1, nil)
	if len(w.list) != maxStimuli || w.list[0] == oldest || w.list[len(w.list)-1].Kind != stimPredator {
		t.Fatal("a predator did not displace the oldest theft")
	}
	info := s.stimuliInfo()
	if info.Active != maxStimuli || info.ByKind[stimPredator] != 1 || info.ByKind[stimTheft] != maxStimuli-1 {
		t.Fatalf("info %+v", info)
	}
}

// facing puts an adult at the centre of tile (x, y) looking along heading.
func facing(s *Sim, x, y int, heading float64) *Creature {
	c := person(s, Female, "Eye", x, y)
	c.Heading = heading
	return c
}

func TestShockIsSeenHeardOrBlocked(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	c := facing(s, x, y, 0)
	c.Genome.Traits.Vision = 9 // sees five tiles even at night
	at := func(dx, dy int) (float64, float64) { return float64(x+dx) + 0.5, float64(y+dy) + 0.5 }
	shock := func() float64 {
		var in [NumInputs]float64
		s.senseStimuli(c, &in)
		return in[inShock]
	}

	// A body two tiles ahead is seen; behind, it is neither seen nor heard.
	bx, by := at(2, 0)
	s.putStimulus(stimCorpse, bx, by, 1, nil, &CorpseBody{ID: 99})
	seen := shock()
	want := stimKinds[stimCorpse].shock * (1 - 0.5*2/(9*(0.55+0.45*s.eco.Light())))
	if math.Abs(seen-want) > 1e-9 {
		t.Fatalf("seen body: shock %.4f, want %.4f", seen, want)
	}
	c.Heading = math.Pi
	if shock() != 0 {
		t.Fatal("a silent body behind was perceived")
	}

	// A wall between hides the body even when looking at it.
	c.Heading = 0
	s.terrain.blocked[y*s.terrain.w+x+1] = true
	if shock() != 0 {
		t.Fatal("a body was seen through a wall")
	}
	s.terrain.blocked[y*s.terrain.w+x+1] = false
	s.stims().list = nil
	s.stims().dirty = true

	// A fight behind is heard; through a wall, half as loud.
	fx, fy := at(-3, 0)
	s.addStimulus(stimFight, fx, fy, 1, nil)
	heard := shock()
	k := stimKinds[stimFight]
	if math.Abs(heard-k.shock*(1-3/k.hearing)) > 1e-9 {
		t.Fatalf("heard fight: shock %.4f", heard)
	}
	s.terrain.blocked[y*s.terrain.w+x-1] = true
	if got := shock(); math.Abs(got-heard/2) > 1e-9 {
		t.Fatalf("fight through a wall: %.4f, want %.4f", got, heard/2)
	}
	s.terrain.blocked[y*s.terrain.w+x-1] = false

	// Seen and heard, a fight ahead is as striking as the clearer of the two;
	// its own maker isn't struck by it; it fades with age.
	c.Heading = math.Pi
	if got := shock(); got <= heard {
		t.Fatalf("a fight in view (%.3f) no stronger than heard (%.3f)", got, heard)
	}
	s.stims().list[0].Actor = &Ref{c.ID, c.Name}
	if shock() != 0 {
		t.Fatal("the fighter was shocked by their own fight")
	}
	s.stims().list[0].Actor = nil
	fresh := shock()
	s.tick += TicksPerSecond
	if got := shock(); got >= fresh || got <= 0 {
		t.Fatalf("an older fight did not fade: %.3f → %.3f", fresh, got)
	}
}

func TestShockSideTellsLeftFromRight(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	c := facing(s, x, y, 0) // looking towards +x; +y is on the right
	side := func(dx, dy int) float64 {
		s.stims().list = nil
		s.stims().dirty = true
		s.addStimulus(stimFight, float64(x+dx)+0.5, float64(y+dy)+0.5, 1, nil)
		var in [NumInputs]float64
		s.senseStimuli(c, &in)
		if in[inShock] <= 0 {
			t.Fatalf("fight at (%d,%d) not perceived", dx, dy)
		}
		return in[inShockSide]
	}
	if r := side(2, 2); math.Abs(r-math.Sqrt2/2) > 1e-9 {
		t.Fatalf("ahead right: side %.3f", r)
	}
	if l := side(2, -2); math.Abs(l+math.Sqrt2/2) > 1e-9 {
		t.Fatalf("ahead left: side %.3f", l)
	}
	if r := side(0, 3); math.Abs(r-1) > 1e-9 {
		t.Fatalf("right: side %.3f", r)
	}
	if a := side(3, 0); math.Abs(a) > 1e-9 {
		t.Fatalf("straight ahead: side %.3f", a)
	}
	c.Heading = math.Pi / 2 // now looking towards +y: +x is on the left
	if l := side(3, 0); math.Abs(l+1) > 1e-9 {
		t.Fatalf("after turning: side %.3f", l)
	}
}

func TestCorpseInFrameKeepsTupleAndViewport(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	c := person(s, Male, "Korban", x, y)
	c.Heading = -1.234
	c.Genome.Traits.Hue = 725.9
	c.Genome.Traits.Size = 1.087
	c.Health = 0
	age := s.ageYears(c)
	s.reap()
	if len(s.creatures) != 0 {
		t.Fatal("not reaped")
	}
	s.tick += 30
	corpses := func(v *Viewport) [][]json.RawMessage {
		var f struct {
			D [][]json.RawMessage `json:"d"`
		}
		if err := json.Unmarshal(s.encodeViewFrame(v), &f); err != nil {
			t.Fatal(err)
		}
		return f.D
	}
	d := corpses(nil)
	if len(d) != 1 || len(d[0]) != 10 {
		t.Fatalf("frame bodies %v", d)
	}
	want := []string{
		strconv.FormatInt(c.ID, 10),
		strconv.FormatFloat(c.X, 'f', 2, 64),
		strconv.FormatFloat(c.Y, 'f', 2, 64),
		"-1.23", "1", "5", "1.09",
		strconv.FormatFloat(age, 'f', 1, 64),
		`"killed"`, "1.5",
	}
	for i, w := range want {
		if string(d[0][i]) != w {
			t.Fatalf("field %d is %s, want %s", i, d[0][i], w)
		}
	}
	away := &Viewport{MinX: c.X + 2, MinY: c.Y + 2, MaxX: c.X + 6, MaxY: c.Y + 6}
	if len(corpses(away)) != 0 || bytes.Contains(s.encodeViewFrame(away), []byte(`"d"`)) {
		t.Fatal("a body outside the viewport was sent")
	}
	if len(corpses(&Viewport{MinX: c.X - 1, MinY: c.Y - 1, MaxX: c.X + 1, MaxY: c.Y + 1})) != 1 {
		t.Fatal("a body inside the viewport was not sent")
	}
	s.opts.NoStimuli = true
	if bytes.Contains(s.encodeFrame(), []byte(`"d"`)) {
		t.Fatal("bodies sent with stimuli off")
	}
}

func TestStimuliSurviveSaveAndRestore(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	a := person(s, Male, "A", x, y)
	b := person(s, Female, "B", x+1, y)
	s.tick = 100
	s.perceiveEvent(a, b, "attack", -0.5, 8)
	s.addStimulus(stimFire, float64(x)+3.5, float64(y)+0.5, 0.5, nil)
	dead := person(s, Male, "C", x-1, y)
	dead.Health = 0
	s.reap()
	s.tick += 7
	data, err := s.MarshalState()
	if err != nil {
		t.Fatal(err)
	}
	r, err := restoreWith(testMap(t, 48), data, fakeCatalog())
	if err != nil {
		t.Fatal(err)
	}
	ja, _ := json.Marshal(s.saveStimuli())
	jb, _ := json.Marshal(r.saveStimuli())
	if !bytes.Equal(ja, jb) || len(s.stims().list) != 3 {
		t.Fatalf("stimuli changed in a save:\n%s\n%s", ja, jb)
	}
	if !bytes.Equal(s.encodeFrame(), r.encodeFrame()) {
		t.Fatal("restored frame differs")
	}
	for _, c := range s.creatures {
		rc := r.byID[c.ID]
		s1, d1, f1 := s.perceivedShock(c)
		s2, d2, f2 := r.perceivedShock(rc)
		if s1 != s2 || d1 != d2 || f1 != f2 || s1 == 0 {
			t.Fatalf("%s senses differently after a restore", c.Name)
		}
	}
	s.addStimulus(stimFall, 1.5, 1.5, 1, nil)
	r.addStimulus(stimFall, 1.5, 1.5, 1, nil)
	if s.stims().list[3].ID != r.stims().list[3].ID {
		t.Fatal("stimulus ids restart after a restore")
	}
}

// witnessRecord is one perceived act, for comparing witness searches.
type witnessRecord struct {
	id   int64
	mode string
	conf float64
}

func TestGridWitnessesMatchAskingEveryone(t *testing.T) {
	s := newSimWith(testMap(t, 64), 3, fakeCatalog())
	rng := rand.New(rand.NewPCG(7, 9))
	// A crowd round the middle of the island, among its hills and water.
	cx, cy := 32.0, 32.0
	var spots []int
	for _, i := range s.terrain.walkable {
		x, y := float64(int(i)%s.terrain.w)+0.5, float64(int(i)/s.terrain.w)+0.5
		if math.Abs(x-cx) <= 14 && math.Abs(y-cy) <= 14 {
			spots = append(spots, int(i))
		}
	}
	if len(spots) < 100 {
		t.Fatalf("only %d walkable tiles near the middle", len(spots))
	}
	for n := range 300 {
		i := spots[rng.IntN(len(spots))]
		c := person(s, Sex(n%2), "P", i%s.terrain.w, i/s.terrain.w)
		c.X += rng.Float64()*0.8 - 0.4
		c.Y += rng.Float64()*0.8 - 0.4
		c.Heading = rng.Float64()*2*math.Pi - math.Pi
		c.Genome.Traits.Vision = 3 + 6*rng.Float64()
		if n%17 == 0 {
			c.Health = 0 // fallen this tick: no witness
		}
	}
	s.grid.rebuild(s.terrain.w, s.terrain.h, s.creatures)
	// People move on after the grid is built, as they do during a step.
	for _, c := range s.creatures {
		c.X += rng.Float64()*0.3 - 0.15
		c.Y += rng.Float64()*0.3 - 0.15
	}
	collect := func(actor, target *Creature, loudness float64) []witnessRecord {
		var out []witnessRecord
		s.witnesses(actor, target, loudness, func(o *Creature, mode string, conf float64) {
			out = append(out, witnessRecord{o.ID, mode, conf})
		})
		slices.SortFunc(out, func(a, b witnessRecord) int { return int(a.id - b.id) })
		return out
	}
	everyone := func(actor, target *Creature, loudness float64) []witnessRecord {
		g := s.grid
		s.grid = grid{}
		defer func() { s.grid = g }()
		return collect(actor, target, loudness)
	}
	compared, seen, heard := 0, 0, 0
	for i := 0; i < len(s.creatures); i += 7 {
		actor := s.creatures[i]
		target := s.creatures[(i+1)%len(s.creatures)]
		for _, loud := range []float64{0, 8} {
			got, want := collect(actor, target, loud), everyone(actor, target, loud)
			if !slices.Equal(got, want) {
				t.Fatalf("actor %d, loudness %g: grid found %d witnesses, everyone %d", actor.ID, loud, len(got), len(want))
			}
			compared++
			for _, w := range got {
				switch w.mode {
				case "seen":
					seen++
				case "heard":
					heard++
				}
			}
		}
	}
	if seen == 0 || heard == 0 {
		t.Fatalf("scene too sparse: %d seen, %d heard over %d acts", seen, heard, compared)
	}
	t.Logf("%d acts, %d sightings, %d hearings", compared, seen, heard)

	// And perceiveEvent through the grid leaves the same memories as before.
	for _, c := range s.creatures {
		c.Memories, c.Relations = nil, nil
	}
	actor, target := s.creatures[0], s.creatures[1]
	s.perceiveEvent(actor, target, "attack", -0.5, 8)
	want := everyone(actor, target, 8)
	n := 0
	for _, c := range s.creatures {
		n += len(c.Memories)
	}
	if n != len(want) {
		t.Fatalf("%d memories for %d witnesses", n, len(want))
	}
}

// TestLiveWorldReplaysWithStimuliAndMoods continues a saved crowded world
// (PROFILE_MAP and PROFILE_STATE, as for TestProfileSave) for a few seconds,
// saves it once bodies, fights and moods are about, and checks that the
// restored copy carries on exactly like the original. Skipped otherwise.
func TestLiveWorldReplaysWithStimuliAndMoods(t *testing.T) {
	mapPath, statePath := os.Getenv("PROFILE_MAP"), os.Getenv("PROFILE_STATE")
	if mapPath == "" || statePath == "" {
		t.Skip("set PROFILE_MAP and PROFILE_STATE")
	}
	load := func() *world.Map {
		raw, err := os.ReadFile(mapPath)
		if err != nil {
			t.Fatal(err)
		}
		var m world.Map
		if err := json.Unmarshal(raw, &m); err != nil {
			t.Fatal(err)
		}
		return &m
	}
	data, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}
	a, err := Restore(load(), data)
	if err != nil {
		t.Fatal(err)
	}
	a.opts.NoStimuli = os.Getenv("X_NOSTIM") != ""
	a.opts.NoAffect = os.Getenv("X_NOAFFECT") != ""
	a.Advance(5 * TicksPerSecond)
	moody := 0
	var faces [5]int
	var mean Affect
	for _, c := range a.creatures {
		m, _ := a.moodCode(c)
		faces[m]++
		if m != 0 {
			moody++
		}
		mean.Fear += c.Affect.Fear / float64(len(a.creatures))
		mean.Anger += c.Affect.Anger / float64(len(a.creatures))
		mean.Joy += c.Affect.Joy / float64(len(a.creatures))
		mean.Grief += c.Affect.Grief / float64(len(a.creatures))
	}
	info := a.stimuliInfo()
	t.Logf("%d people, faces calm/fear/anger/joy/grief %v, mean moods %+v, stimuli %+v", len(a.creatures), faces, mean, info)
	if (info.Active == 0 || moody == 0) && os.Getenv("X_NOSTIM") == "" {
		t.Fatal("nothing to replay")
	}
	saved, err := a.MarshalState()
	if err != nil {
		t.Fatal(err)
	}
	b, err := Restore(load(), saved)
	if err != nil {
		t.Fatal(err)
	}
	a.Advance(10 * TicksPerSecond)
	b.Advance(10 * TicksPerSecond)
	sa, err := a.MarshalState()
	if err != nil {
		t.Fatal(err)
	}
	sb, err := b.MarshalState()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(a.encodeFrame(), b.encodeFrame()) {
		t.Fatal("the restored world's frame diverged")
	}
	// A restore bumps the fields' version for the observer (eco.plotVer);
	// everything else must match.
	ja, jb := liveState(t, sa), liveState(t, sb)
	for k, v := range ja {
		if !bytes.Equal(v, jb[k]) {
			t.Fatalf("the restored world diverged in %s", k)
		}
	}
}

// liveState unpacks a save into its parts, leaving out the observer's
// version of the fields.
func liveState(t *testing.T, data []byte) map[string]json.RawMessage {
	t.Helper()
	zr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := io.ReadAll(zr)
	if err != nil {
		t.Fatal(err)
	}
	var parts, eco map[string]json.RawMessage
	if err := json.Unmarshal(raw, &parts); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(parts["eco"], &eco); err != nil {
		t.Fatal(err)
	}
	delete(eco, "plotVer")
	if parts["eco"], err = json.Marshal(eco); err != nil {
		t.Fatal(err)
	}
	return parts
}
