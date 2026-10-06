package sim

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestDecisionBudgetPersistsBetweenMotorTicks(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	c := person(s, Male, "A", x, y)
	s.tick = 1
	s.decide(c)
	if c.Cognition == nil {
		t.Fatal("first decision was not recorded")
	}
	want := c.output
	c.Genome.BOut[outRest] = 4
	s.tick = 2 // id 1 is next scheduled at tick 3
	s.decide(c)
	if c.output != want {
		t.Fatal("decision changed on a motor-only tick")
	}
	s.tick = 3
	s.decide(c)
	if c.output[outRest] <= want[outRest] {
		t.Fatal("scheduled decision did not run")
	}
}

// Three in four creatures act on their saved decision on the first tick after
// a restore, so anything their route executor reads must survive the save:
// gathering looks at how much they want each item before they sense again.
func TestRestoredWorldActsBeforeItSenses(t *testing.T) {
	a := New(testMap(t, 64), 7)
	a.Advance(20 * TicksPerSecond)
	if len(a.creatures) == 0 {
		t.Fatal("nobody alive to save")
	}
	data, err := a.MarshalState()
	if err != nil {
		t.Fatal(err)
	}
	b, err := Restore(testMap(t, 64), data)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range b.creatures {
		if c.Cognition == nil {
			t.Fatalf("creature %d has no saved decision", c.ID)
		}
		b.travelTarget(c, ActGather) // panicked on an unsaved cache
		o := a.byID[c.ID]
		if !slices.Equal(b.appeal(c), a.appeal(o)) || c.Sample != o.Sample {
			t.Fatalf("creature %d wants different things after the restore", c.ID)
		}
	}
	b.Advance(TicksPerSecond)
}

func TestMigrationArchivesOriginalAndRefusesInvalidBackup(t *testing.T) {
	m := &Manager{dir: t.TempDir()}
	s := &Sim{loadedVersion: 9, migrationOriginal: []byte("original saved world")}
	if err := m.archiveMigration("world", s); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(m.dir, "world.json.gz.v9-before-v13.bak")
	data, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(data, []byte("original saved world")) {
		t.Fatal("original was not archived")
	}
	s.migrationOriginal = []byte("later save")
	if err := m.archiveMigration("world", s); err != nil {
		t.Fatal(err)
	}
	data, _ = os.ReadFile(path)
	if string(data) != "original saved world" {
		t.Fatal("archive overwritten")
	}
	bad := filepath.Join(m.dir, "bad.json.gz.v9-before-v13.bak")
	if err := os.Mkdir(bad, 0700); err != nil {
		t.Fatal(err)
	}
	s.migrationOriginal = []byte("retain me")
	if err := m.archiveMigration("bad", s); err == nil || len(s.migrationOriginal) == 0 {
		t.Fatal("bad backup permitted overwriting original")
	}
}

// A v10 brain (78 senses, 15 actions) widens to today's shape with every
// inherited and learned weight where it was and the new ones unwired.
func TestV10BrainWidensKeepingWeights(t *testing.T) {
	const H = 3
	g := emptyGenome(H)
	g.WIn = make(weights, alarmInputs*H)
	g.WOut = make(weights, legacyOutputs*H)
	g.BOut = make(weights, legacyOutputs)
	for i := range g.WIn {
		g.WIn[i] = float32(i + 1)
	}
	for i := range g.WOut {
		g.WOut[i] = float32(-i - 1)
	}
	for i := range g.BOut {
		g.BOut[i] = float32(i) + 0.5
	}
	m := &Mind{DIn: append(weights{}, g.WIn...), EIn: append(weights{}, g.WIn...),
		DOut: append(weights{}, g.WOut...), EOut: append(weights{}, g.WOut...),
		Grown: 1, GIn: make(weights, alarmInputs), GRec: make(weights, H), GOut: make(weights, legacyOutputs),
		GBias: make(weights, 1), GTau: filled(1, 1), GBorn: make(weights, 1), GUse: make(weights, 1),
		EGIn: make(weights, alarmInputs), EGOut: make(weights, legacyOutputs), GAct: make([]float64, 1)}
	for i := range m.GOut {
		m.GOut[i] = float32(i + 100)
	}
	g.upgrade(10)
	m.upgrade(g)
	if !g.valid() || !m.valid(g) {
		t.Fatalf("widened brain is malformed: wIn %d wOut %d bOut %d", len(g.WIn), len(g.WOut), len(g.BOut))
	}
	for i := range alarmInputs * H {
		if g.WIn[i] != float32(i+1) || m.DIn[i] != float32(i+1) {
			t.Fatalf("sense weight %d moved", i)
		}
	}
	for i := alarmInputs * H; i < NumInputs*H; i++ {
		if g.WIn[i] != 0 || m.DIn[i] != 0 {
			t.Fatalf("new sense weight %d is wired", i)
		}
	}
	for h := range H {
		for o := range NumOutputs {
			want := float32(0)
			if o < legacyOutputs {
				want = float32(-(h*legacyOutputs + o) - 1)
			}
			if g.WOut[h*NumOutputs+o] != want || m.DOut[h*NumOutputs+o] != want {
				t.Fatalf("action weight h%d o%d = %v, want %v", h, o, g.WOut[h*NumOutputs+o], want)
			}
		}
	}
	if g.BOut[legacyOutputs-1] != float32(legacyOutputs-1)+0.5 || g.BOut[outTalk] != 0 || g.BOut[outTrade] != 0 {
		t.Fatalf("action biases %v", g.BOut)
	}
	if m.GOut[legacyOutputs-1] != float32(legacyOutputs-1+100) || m.GOut[outTalk] != 0 || len(m.GIn) != NumInputs {
		t.Fatalf("grown neuron not widened in place: %v", m.GOut)
	}
}

// A v11 brain (96 senses) widens to today's shape with every weight in place.
func TestV11BrainWidensKeepingWeights(t *testing.T) {
	const H = 4
	g := emptyGenome(H)
	g.WIn = make(weights, engineInputs*H)
	for i := range g.WIn {
		g.WIn[i] = float32(i + 1)
	}
	m := newMind(g)
	m.Grown = 1
	m.GIn, m.EGIn = make(weights, engineInputs), make(weights, engineInputs)
	m.GIn[engineInputs-1] = 7
	m.GRec, m.GOut, m.EGOut = make(weights, H), make(weights, NumOutputs), make(weights, NumOutputs)
	m.GBias, m.GTau, m.GBorn, m.GUse, m.GAct = make(weights, 1), filled(1, 1), make(weights, 1), make(weights, 1), make([]float64, 1)
	g.upgrade(11)
	m.upgrade(g)
	if !g.valid() || !m.valid(g) {
		t.Fatalf("widened brain is malformed: wIn %d", len(g.WIn))
	}
	for i := range engineInputs * H {
		if g.WIn[i] != float32(i+1) {
			t.Fatalf("sense weight %d moved", i)
		}
	}
	for i := engineInputs * H; i < NumInputs*H; i++ {
		if g.WIn[i] != 0 || m.DIn[i] != 0 {
			t.Fatalf("new sense weight %d is wired", i)
		}
	}
	if m.GIn[engineInputs-1] != 7 || m.GIn[NumInputs-1] != 0 {
		t.Fatal("grown neuron's senses not widened in place")
	}
}
