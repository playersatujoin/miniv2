package sim

import (
	"encoding/json"
	"os"
	"strconv"
	"testing"
	"time"

	"miniv2/backend/internal/world"
)

// TestProfileSave advances a saved world (PROFILE_MAP and PROFILE_STATE) by
// PROFILE_SECONDS (default 20) for profiling with -cpuprofile, and writes the
// result to PROFILE_OUT if set, to compare builds; skipped otherwise.
func TestProfileSave(t *testing.T) {
	mapPath, statePath := os.Getenv("PROFILE_MAP"), os.Getenv("PROFILE_STATE")
	if mapPath == "" || statePath == "" {
		t.Skip("set PROFILE_MAP and PROFILE_STATE")
	}
	raw, err := os.ReadFile(mapPath)
	if err != nil {
		t.Fatal(err)
	}
	var m world.Map
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}
	s, err := Restore(&m, data)
	if err != nil {
		t.Fatal(err)
	}
	seconds := 20
	if v, err := strconv.Atoi(os.Getenv("PROFILE_SECONDS")); err == nil && v > 0 {
		seconds = v
	}
	t.Logf("%d people", len(s.creatures))
	start := time.Now()
	s.Advance(seconds * TicksPerSecond)
	t.Logf("advance %.3f ms/tick", float64(time.Since(start).Microseconds())/1000/float64(seconds*TicksPerSecond))
	t.Logf("%d people after %d s", len(s.creatures), seconds)
	if out := os.Getenv("PROFILE_OUT"); out != "" {
		data, err := s.MarshalState()
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(out, data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
}
