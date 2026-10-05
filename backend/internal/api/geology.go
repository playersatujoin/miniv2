package api

import (
	"net/http"
	"strconv"

	"miniv2/backend/internal/chem"
	"miniv2/backend/internal/world"
)

type depositItemJSON struct {
	ID       chem.ItemID `json:"id"`
	Name     string      `json:"name"`
	Kind     string      `json:"kind"`
	Formula  string      `json:"formula,omitempty"`
	Elements []string    `json:"elements"`
}

// geologyResponse matches MapGeology in frontend/src/api/client.ts.
type geologyResponse struct {
	Width     int                 `json:"width"`
	Height    int                 `json:"height"`
	RockTypes []world.RockType    `json:"rockTypes"`
	Rocks     []int               `json:"rocks"` // rock ids, row-major
	Features  []world.Feature     `json:"features"`
	Models    []chem.DepositModel `json:"models"`
	Items     []depositItemJSON   `json:"items"`
	Deposits  [][5]int            `json:"deposits"` // x, y, item index, 1 if on the surface, model index
}

// Resources found almost everywhere; the terrain itself already shows them.
var everywhere = map[chem.ItemID]bool{"serat": true, "air": true, "udara": true, "air_laut": true, "tanah_liat": true}

// obvious reports deposits the map already makes visible: sand on sand, plain
// stone in rock, limestone in karst, wood in trees, stones in boulders.
func obvious(m *world.Map, d chem.Deposit) bool {
	i := d.Y*m.Width + d.X
	if d.Surface {
		obj := m.Layers.Objects[i]
		return d.Item == "kayu" && (obj == world.Tree || obj == world.Pine) || d.Item == "batu" && obj == world.Boulder
	}
	switch m.Layers.Ground[i] {
	case world.Sand:
		return d.Item == "pasir"
	case world.Mountain, world.StoneFloor, world.VolcanicRock:
		return d.Item == "batu"
	case world.Limestone:
		return d.Item == "batu" || d.Item == "batu_kapur"
	}
	return false
}

// geology serves the map's geological model and its notable deposits.
func (s *Server) geology(w http.ResponseWriter, r *http.Request) {
	m, ok := s.lookup(w, r)
	if !ok {
		return
	}
	etag := `"geo-` + m.ID + "-" + strconv.FormatInt(m.UpdatedAt.UnixNano(), 36) + `"`
	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", "no-cache")
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	g := chem.NewGeology(m)
	geo := g.GeoModel()
	models := chem.DepositModels()
	modelIndex := make(map[string]int, len(models))
	for i, md := range models {
		modelIndex[md.Key] = i
	}
	resp := geologyResponse{
		Width:     m.Width,
		Height:    m.Height,
		RockTypes: world.RockTypes,
		Rocks:     make([]int, len(geo.Rock)),
		Features:  geo.Features,
		Models:    models,
		Items:     []depositItemJSON{},
		Deposits:  [][5]int{},
	}
	for i, r := range geo.Rock {
		resp.Rocks[i] = int(r) // not []uint8, which JSON would encode as base64
	}
	if resp.Features == nil {
		resp.Features = []world.Feature{}
	}
	itemIndex := map[chem.ItemID]int{}
	for _, d := range g.Deposits() {
		if everywhere[d.Item] || obvious(m, d) {
			continue
		}
		idx, ok := itemIndex[d.Item]
		if !ok {
			it, _ := chem.ItemByID(d.Item)
			idx = len(resp.Items)
			itemIndex[d.Item] = idx
			resp.Items = append(resp.Items, depositItemJSON{ID: it.ID, Name: it.Name, Kind: it.Kind, Formula: it.Formula, Elements: it.Elements})
		}
		surface := 0
		if d.Surface {
			surface = 1
		}
		resp.Deposits = append(resp.Deposits, [5]int{d.X, d.Y, idx, surface, modelIndex[d.Model]})
	}
	writeJSON(w, http.StatusOK, resp)
}
