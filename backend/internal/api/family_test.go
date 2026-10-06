package api

import (
	"net/http"
	"testing"

	"miniv2/backend/internal/sim"
	"miniv2/backend/internal/world"
)

func TestFamilyEndpoint(t *testing.T) {
	srv := newTestServer(t)
	seed := uint32(3)
	var m world.Map
	do(t, "POST", srv.URL+"/api/maps", createRequest{Name: "Silsilah", Width: 48, Height: 48, Seed: &seed}, &m)
	base := srv.URL + "/api/maps/" + m.ID + "/sim"
	do(t, "PUT", base+"/speed", map[string]int{"speed": 0}, nil)

	// Adam is id 1 in a new world.
	var tree sim.FamilyView
	if code := do(t, "GET", base+"/creatures/1/family?depth=3", nil, &tree); code != http.StatusOK {
		t.Fatalf("family: status %d", code)
	}
	if tree.Root != 1 || tree.Depth != 3 || len(tree.People) != 1 || !tree.People[0].Founder || tree.People[0].Name != "Adam" ||
		tree.Genetics == nil {
		t.Fatalf("Adam's tree: %+v", tree)
	}
	var raw map[string]any
	do(t, "GET", base+"/creatures/1/family", nil, &raw)
	for _, key := range []string{"root", "depth", "ancestors", "descendants", "people", "genetics", "parentsRelated"} {
		if _, ok := raw[key]; !ok {
			t.Errorf("family JSON missing %q", key)
		}
	}
	for path, want := range map[string]int{
		"/creatures/999999/family":        http.StatusNotFound,
		"/creatures/abc/family":           http.StatusNotFound,
		"/creatures/1/family?depth=0":     http.StatusBadRequest,
		"/creatures/1/family?depth=lots":  http.StatusBadRequest,
		"/creatures/1/family?depth=99999": http.StatusOK,
	} {
		if code := do(t, "GET", base+path, nil, nil); code != want {
			t.Errorf("%s: status %d, want %d", path, code, want)
		}
	}
	// The creature detail carries the genetics too.
	var detail map[string]any
	do(t, "GET", base+"/creatures/1", nil, &detail)
	if g, ok := detail["genetics"].(map[string]any); !ok || g["f"] == nil || g["defects"] == nil {
		t.Fatalf("creature detail genetics: %v", detail["genetics"])
	}
}
