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
	// The first world that farms by the time it is saved, so the save covers
	// fields, seed and water memory whatever the rules make of each seed.
	var a *Sim
	var seed uint64
	farmers, remember := 0, 0
	for seed = 1; seed <= 12; seed++ {
		a = New(island(), seed)
		a.Advance(before * TicksPerSecond)
		farmers, remember = 0, 0
		for _, c := range a.creatures {
			if a.seedInHand(c) != "" {
				farmers++
			}
			if c.WaterX != 0 || c.WaterY != 0 {
				remember++
			}
		}
		if len(a.eco.Plots()) > 0 && remember > 0 {
			break
		}
	}
	if len(a.eco.Plots()) == 0 || remember == 0 {
		t.Fatal("no world among seeds 1–12 farms by the time it is saved")
	}
	twin := New(island(), seed)
	twin.Advance(before * TicksPerSecond)
	t.Logf("seed %d saved at %d s: %d people, %d plots, %d carrying seed", seed, before, len(a.creatures), len(a.eco.Plots()), farmers)

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
