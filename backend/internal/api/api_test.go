package api

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"miniv2/backend/internal/sim"
	"miniv2/backend/internal/store"
	"miniv2/backend/internal/world"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	dir := t.TempDir()
	st, err := store.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	sims, err := sim.NewManager(filepath.Join(dir, "sims"))
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(New(st, sims))
	t.Cleanup(func() {
		srv.CloseClientConnections()
		srv.Close()
		sims.Close()
	})
	return srv
}

func do(t *testing.T, method, url string, body any, out any) int {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req, _ := http.NewRequest(method, url, &buf)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if out != nil {
		json.NewDecoder(res.Body).Decode(out)
	}
	return res.StatusCode
}

func TestMapLifecycle(t *testing.T) {
	srv := newTestServer(t)
	seed := uint32(7)

	var created world.Map
	if code := do(t, "POST", srv.URL+"/api/maps", createRequest{Name: " Test ", Width: 32, Height: 24, Seed: &seed}, &created); code != http.StatusCreated {
		t.Fatalf("create: status %d", code)
	}
	if created.Name != "Test" || created.Seed != 7 || len(created.Layers.Ground) != 32*24 {
		t.Fatalf("unexpected created map: %+v", created.Summary())
	}

	var list []world.Summary
	do(t, "GET", srv.URL+"/api/maps", nil, &list)
	if len(list) != 1 || list[0].ID != created.ID {
		t.Fatalf("list: %+v", list)
	}

	// Paint a wall next to spawn and rename.
	edit := updateRequest{Name: "Edited", Spawn: created.Spawn, Layers: created.Layers}
	edit.Layers.Objects[created.Spawn.Y*created.Width+created.Spawn.X+1] = world.Wall
	var updated world.Map
	if code := do(t, "PUT", srv.URL+"/api/maps/"+created.ID, edit, &updated); code != http.StatusOK {
		t.Fatalf("update: status %d", code)
	}
	if updated.Name != "Edited" || !updated.UpdatedAt.After(created.UpdatedAt) {
		t.Fatalf("update not applied: %+v", updated.Summary())
	}

	res, err := http.Get(srv.URL + "/api/maps/" + created.ID + "/preview.png?scale=3")
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusOK || res.Header.Get("Content-Type") != "image/png" {
		t.Fatalf("preview: %d %s", res.StatusCode, res.Header.Get("Content-Type"))
	}

	if code := do(t, "DELETE", srv.URL+"/api/maps/"+created.ID, nil, nil); code != http.StatusNoContent {
		t.Fatalf("delete: status %d", code)
	}
	if code := do(t, "GET", srv.URL+"/api/maps/"+created.ID, nil, nil); code != http.StatusNotFound {
		t.Fatalf("get after delete: status %d", code)
	}
}

func TestUpdateRejectsInvalid(t *testing.T) {
	srv := newTestServer(t)
	var m world.Map
	do(t, "POST", srv.URL+"/api/maps", createRequest{Name: "x", Width: 16, Height: 16}, &m)

	cases := map[string]updateRequest{
		"wrong layer size": {Name: "x", Spawn: m.Spawn, Layers: world.Layers{Ground: []int{1}, Objects: []int{0}}},
		"unknown tile":     {Name: "x", Spawn: m.Spawn, Layers: withGround(m.Layers, 0, 99)},
		"spawn in water":   {Name: "x", Spawn: world.Point{X: 0, Y: 0}, Layers: withGround(m.Layers, 0, world.DeepWater)},
		"empty name":       {Name: "  ", Spawn: m.Spawn, Layers: m.Layers},
	}
	for name, req := range cases {
		if code := do(t, "PUT", srv.URL+"/api/maps/"+m.ID, req, nil); code != http.StatusBadRequest {
			t.Errorf("%s: status %d, want 400", name, code)
		}
	}

	if code := do(t, "POST", srv.URL+"/api/maps", createRequest{Name: "big", Width: 1000, Height: 16}, nil); code != http.StatusBadRequest {
		t.Errorf("oversized create: status %d, want 400", code)
	}
}

func withGround(l world.Layers, i, id int) world.Layers {
	g := append([]int(nil), l.Ground...)
	g[i] = id
	return world.Layers{Ground: g, Objects: l.Objects}
}

// firstEvents reads the stream until it has seen one event of each kind and
// returns their decoded payloads.
func firstEvents(t *testing.T, url string, kinds ...string) map[string]map[string]any {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if ct := res.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("content type %q", ct)
	}
	got := map[string]map[string]any{}
	sc := bufio.NewScanner(res.Body)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	event := ""
	for sc.Scan() {
		line := sc.Text()
		if name, ok := strings.CutPrefix(line, "event: "); ok {
			event = name
			continue
		}
		if data, ok := strings.CutPrefix(line, "data: "); ok && event != "" {
			var payload map[string]any
			if err := json.Unmarshal([]byte(data), &payload); err != nil {
				t.Fatalf("bad %s JSON: %v", event, err)
			}
			if _, seen := got[event]; !seen {
				got[event] = payload
			}
			if len(got) == len(kinds) {
				return got
			}
		}
	}
	t.Fatalf("stream ended after %v: %v", got, sc.Err())
	return nil
}

func TestSimEndpoints(t *testing.T) {
	srv := newTestServer(t)
	seed := uint32(3)
	var m world.Map
	do(t, "POST", srv.URL+"/api/maps", createRequest{Name: "Life", Width: 48, Height: 48, Seed: &seed}, &m)
	base := srv.URL + "/api/maps/" + m.ID + "/sim"

	var info sim.Info
	if code := do(t, "GET", base, nil, &info); code != http.StatusOK {
		t.Fatalf("info: status %d", code)
	}
	if info.MapID != m.ID || info.Population == 0 || info.Females+info.Males != info.Population || info.Speed != 1 {
		t.Fatalf("unexpected info: %+v", info)
	}

	if code := do(t, "PUT", base+"/speed", map[string]int{"speed": 5}, &info); code != http.StatusOK || info.Speed != 5 {
		t.Fatalf("set speed: status %d speed %d", code, info.Speed)
	}
	for _, body := range []any{map[string]int{"speed": 3}, map[string]int{"speed": -1}, map[string]string{}} {
		if code := do(t, "PUT", base+"/speed", body, nil); code != http.StatusBadRequest {
			t.Errorf("speed %v: status %d, want 400", body, code)
		}
	}

	// Pause so the creature we inspect can't die mid-test.
	do(t, "PUT", base+"/speed", map[string]int{"speed": 0}, nil)
	events := firstEvents(t, base+"/stream", "frame", "structures")
	creatures, _ := events["frame"]["c"].([]any)
	if len(creatures) != 2 {
		t.Fatalf("a new world should hold Adam and Hawa, frame: %v", events["frame"])
	}
	first := creatures[0].([]any)
	if len(first) != 12 || first[1] != "Adam" {
		t.Fatalf("creature tuple has %d fields: %v", len(first), first)
	}
	if _, ok := events["structures"]["v"].(float64); !ok {
		t.Fatalf("structures event without version: %v", events["structures"])
	}
	id := int64(first[0].(float64))

	var detail sim.CreatureDetail
	if code := do(t, "GET", base+"/creatures/"+strconv.FormatInt(id, 10), nil, &detail); code != http.StatusOK {
		t.Fatalf("creature: status %d", code)
	}
	b := detail.Brain
	if detail.ID != id || len(b.Input) != 55 || len(b.WIn) != 55 || len(b.WIn[0]) != 24 || len(b.WRec) != 24 ||
		len(b.WOut) != 24 || len(b.WOut[0]) != 12 || len(b.Output) != 12 || len(b.InputLabels) != 55 || len(b.OutputLabels) != 12 {
		t.Fatalf("unexpected creature detail shape: id=%d in=%d", detail.ID, len(b.Input))
	}
	if detail.Sex != "male" || detail.Role != "none" || detail.Health != 1 || detail.Inventory == nil || detail.Mother != nil {
		t.Fatalf("unexpected Adam: %+v", detail)
	}

	// Raw JSON keys the frontend relies on.
	var raw map[string]any
	do(t, "GET", base+"/creatures/"+strconv.FormatInt(id, 10), nil, &raw)
	for _, key := range []string{"spouse", "house", "inventory", "deeds", "reputation", "role", "health"} {
		if _, ok := raw[key]; !ok {
			t.Errorf("creature JSON missing %q", key)
		}
	}

	var know sim.Knowledge
	if code := do(t, "GET", base+"/knowledge", nil, &know); code != http.StatusOK {
		t.Fatalf("knowledge: status %d", code)
	}
	if len(know.Elements) != 118 || len(know.TierNames) != 8 || len(know.Techs) == 0 || len(know.StructureKinds) == 0 || know.Era != 1 {
		t.Fatalf("knowledge: %d elements, %d tiers, %d techs, %d kinds, era %d",
			len(know.Elements), len(know.TierNames), len(know.Techs), len(know.StructureKinds), know.Era)
	}
	if e := know.Elements[25]; e.Symbol != "Fe" || e.Z != 26 {
		t.Fatalf("element 26 is %+v", e)
	}

	var info2 map[string]any
	do(t, "GET", base, nil, &info2)
	for _, key := range []string{"era", "tier", "tierName", "elementsDiscovered", "elementsTotal", "houses", "structures", "crimes", "kindness", "kills"} {
		if _, ok := info2[key]; !ok {
			t.Errorf("info JSON missing %q", key)
		}
	}
	if causes, _ := info2["deathsByCause"].(map[string]any); causes["killed"] == nil {
		t.Errorf("deathsByCause missing killed: %v", info2["deathsByCause"])
	}

	for _, path := range []string{"/creatures/999999", "/creatures/abc"} {
		if code := do(t, "GET", base+path, nil, nil); code != http.StatusNotFound {
			t.Errorf("%s: status %d, want 404", path, code)
		}
	}
	if code := do(t, "GET", srv.URL+"/api/maps/ffffffffffffffff/sim", nil, nil); code != http.StatusNotFound {
		t.Errorf("unknown map: status %d, want 404", code)
	}

	// Deleting the map stops its world.
	do(t, "DELETE", srv.URL+"/api/maps/"+m.ID, nil, nil)
	if code := do(t, "GET", base, nil, nil); code != http.StatusNotFound {
		t.Errorf("sim after delete: status %d, want 404", code)
	}
}

func TestGeologyEndpoint(t *testing.T) {
	srv := newTestServer(t)
	seed := uint32(1337)
	var m world.Map
	do(t, "POST", srv.URL+"/api/maps", createRequest{Name: "Geo", Width: 64, Height: 64, Seed: &seed}, &m)

	var geo struct {
		Width, Height int
		RockTypes     []world.RockType
		Rocks         []int
		Features      []world.Feature
		Models        []struct{ Key, Name, Example string }
		Items         []struct {
			ID       string
			Elements []string
		}
		Deposits [][5]int
	}
	if code := do(t, "GET", srv.URL+"/api/maps/"+m.ID+"/geology", nil, &geo); code != http.StatusOK {
		t.Fatalf("geology: status %d", code)
	}
	if geo.Width != 64 || len(geo.Rocks) != 64*64 || len(geo.RockTypes) == 0 || len(geo.Features) == 0 || len(geo.Models) == 0 {
		t.Fatalf("unexpected geology shape: w=%d rocks=%d types=%d features=%d models=%d",
			geo.Width, len(geo.Rocks), len(geo.RockTypes), len(geo.Features), len(geo.Models))
	}
	for _, d := range geo.Deposits {
		if d[0] < 0 || d[0] >= 64 || d[1] < 0 || d[1] >= 64 || d[2] >= len(geo.Items) || d[4] >= len(geo.Models) {
			t.Fatalf("deposit out of range: %v", d)
		}
	}
	hasGold := false
	for _, it := range geo.Items {
		if it.ID == "bijih_emas" {
			hasGold = true
		}
	}
	if !hasGold {
		t.Error("no gold deposits listed on a 64×64 arc island")
	}

	res, err := http.Get(srv.URL + "/api/maps/" + m.ID + "/geology")
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	req, _ := http.NewRequest("GET", srv.URL+"/api/maps/"+m.ID+"/geology", nil)
	req.Header.Set("If-None-Match", res.Header.Get("ETag"))
	if res2, err := http.DefaultClient.Do(req); err != nil || res2.StatusCode != http.StatusNotModified {
		t.Fatalf("conditional geology request: %v %v", res2, err)
	}
}

func TestSimDemographyEndpoint(t *testing.T) {
	srv := newTestServer(t)
	seed := uint32(3)
	var m world.Map
	do(t, "POST", srv.URL+"/api/maps", createRequest{Name: "Demo", Width: 48, Height: 48, Seed: &seed}, &m)
	base := srv.URL + "/api/maps/" + m.ID + "/sim"

	var info struct {
		SecondsPerYear float64 `json:"secondsPerYear"`
		Year           int     `json:"year"`
	}
	if code := do(t, "GET", base, nil, &info); code != http.StatusOK || info.SecondsPerYear != 8 || info.Year < 1 {
		t.Fatalf("info: status %d, %+v", code, info)
	}

	var raw map[string]json.RawMessage
	if code := do(t, "GET", base+"/demography", nil, &raw); code != http.StatusOK {
		t.Fatalf("demography: status %d", code)
	}
	for _, key := range []string{"secondsPerYear", "year", "pyramid", "current", "history", "reference"} {
		if _, ok := raw[key]; !ok {
			t.Errorf("demography is missing %q", key)
		}
	}
	if string(raw["history"]) == "null" {
		t.Error("history should be an array, not null")
	}
	var current map[string]json.RawMessage
	json.Unmarshal(raw["current"], &current)
	for _, key := range []string{"windowYears", "personYears", "lifeExpectancy", "lifeExpectancy15", "survivalTo15",
		"infantMortality", "modalAgeAdultDeath", "tfr", "meanAgeFirstBirth", "meanBirthInterval", "sexRatio",
		"householdSize", "homicideRate", "gini", "deathsByCause"} {
		if _, ok := current[key]; !ok {
			t.Errorf("current is missing %q", key)
		}
	}
	var pyramid []struct {
		Label        string
		From         int
		To           *int
		Female, Male int
	}
	json.Unmarshal(raw["pyramid"], &pyramid)
	if len(pyramid) != 17 || pyramid[0].Label != "0–4" || pyramid[16].To != nil {
		t.Fatalf("pyramid: %+v", pyramid)
	}
	var ref map[string]struct {
		Low, High float64
		Source    string
	}
	json.Unmarshal(raw["reference"], &ref)
	if r := ref["lifeExpectancy"]; r.Low != 21 || r.High != 37 || r.Source == "" {
		t.Fatalf("life expectancy reference: %+v", r)
	}

	if code := do(t, "GET", srv.URL+"/api/maps/ffffffffffffffff/sim/demography", nil, nil); code != http.StatusNotFound {
		t.Errorf("unknown map: status %d", code)
	}
}
