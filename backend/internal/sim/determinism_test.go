package sim

import (
	"bytes"
	"encoding/json"
	"os"
	"strconv"
	"testing"

	"miniv2/backend/internal/world"
)

// A world on the real starter island, saved once people farm and remember
// their rivers, must carry on exactly as if it had never been saved, and a
// second run from the same seed must match the first.
func TestSameSeedAndSaveRestoreReplayExactly(t *testing.T) {
	if testing.Short() {
		t.Skip("runs several simulated minutes")
	}
	before, after := 6*60, 2*60 // simulated seconds
	if v, err := strconv.Atoi(os.Getenv("REPLAY_MINUTES")); err == nil {
		before, after = v*60, v*30
	}
	island := func() *world.Map {
		m := world.Generate(world.GenOptions{Name: "Starter Island", Width: 128, Height: 128, Seed: 1337})
		m.ID = "0000000000000000"
		return m
	}
	a, twin := New(island(), 5), New(island(), 5)
	a.Advance(before * TicksPerSecond)
	twin.Advance(before * TicksPerSecond)

	farmers, remember := 0, 0
	for _, c := range a.creatures {
		if s := a.seedInHand(c); s != "" {
			farmers++
		}
		if c.WaterX != 0 || c.WaterY != 0 {
			remember++
		}
	}
	if len(a.eco.Plots()) == 0 || remember == 0 {
		t.Fatalf("the test should cover fields and water memory: %d plots, %d remember water", len(a.eco.Plots()), remember)
	}
	t.Logf("saving at %d s: %d people, %d plots, %d carrying seed", before, len(a.creatures), len(a.eco.Plots()), farmers)

	data, err := a.MarshalState()
	if err != nil {
		t.Fatal(err)
	}
	b, err := Restore(island(), data)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range []*Sim{a, b, twin} {
		s.Advance(after * TicksPerSecond)
	}
	snapshot := func(s *Sim) []byte {
		structs, _ := s.Structures()
		views, err := json.Marshal([]any{s.Info(), s.Ecology(), s.Demography(), s.creatures})
		if err != nil {
			t.Fatal(err)
		}
		return bytes.Join([][]byte{s.encodeFrame(), structs, views}, nil)
	}
	sa, sb, st := snapshot(a), snapshot(b), snapshot(twin)
	if !bytes.Equal(sa, sb) {
		t.Fatal("the restored world diverged from the original")
	}
	if !bytes.Equal(sa, st) {
		t.Fatal("two worlds from the same seed diverged")
	}
}
