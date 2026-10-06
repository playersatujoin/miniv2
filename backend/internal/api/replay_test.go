package api

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"miniv2/backend/internal/sim"
	"miniv2/backend/internal/world"
)

func TestReplayEndpoints(t *testing.T) {
	srv := newTestServer(t)
	seed := uint32(5)
	var m world.Map
	do(t, "POST", srv.URL+"/api/maps", createRequest{Name: "Replay", Width: 48, Height: 48, Seed: &seed}, &m)
	base := srv.URL + "/api/maps/" + m.ID + "/sim"
	if code := do(t, "GET", base+"/villages", nil, nil); code != http.StatusOK {
		t.Fatalf("villages: status %d", code)
	}
	var idx sim.ReplayIndex
	deadline := time.Now().Add(5 * time.Second)
	for len(idx.Frames) == 0 && time.Now().Before(deadline) {
		do(t, "GET", base+"/replay", nil, &idx)
		time.Sleep(100 * time.Millisecond)
	}
	if len(idx.Frames) == 0 || len(idx.Structures) == 0 {
		t.Fatalf("nothing recorded: %+v", idx)
	}
	tick := int64(idx.Frames[0][0])
	res, err := http.Get(base + "/replay/frame?tick=" + jsonNumber(tick))
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(res.Body)
	res.Body.Close()
	var frame struct {
		T int64 `json:"t"`
	}
	if res.StatusCode != http.StatusOK || json.Unmarshal(body, &frame) != nil || frame.T != tick {
		t.Fatalf("frame: status %d, %s", res.StatusCode, body)
	}
	v := res.Header.Get("X-Structures-Version")
	if code := do(t, "GET", base+"/replay/structures/"+v, nil, nil); code != http.StatusOK {
		t.Fatalf("structures v%s: status %d", v, code)
	}
	if code := do(t, "GET", base+"/replay/frame?tick=x", nil, nil); code != http.StatusBadRequest {
		t.Fatalf("bad tick: status %d", code)
	}
}

func jsonNumber(n int64) string {
	b, _ := json.Marshal(n)
	return string(b)
}

func TestReplayIndexIsGzippedWhenAccepted(t *testing.T) {
	srv := newTestServer(t)
	seed := uint32(6)
	var m world.Map
	do(t, "POST", srv.URL+"/api/maps", createRequest{Name: "Zip", Width: 32, Height: 32, Seed: &seed}, &m)
	req, _ := http.NewRequest("GET", srv.URL+"/api/maps/"+m.ID+"/sim/replay", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	res, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.Header.Get("Content-Encoding") != "gzip" {
		t.Fatalf("not compressed: %v", res.Header)
	}
	zr, err := gzip.NewReader(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	var idx sim.ReplayIndex
	if err := json.NewDecoder(zr).Decode(&idx); err != nil {
		t.Fatal(err)
	}
}
