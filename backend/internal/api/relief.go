package api

import (
	"net/http"
	"strconv"

	"miniv2/backend/internal/world"
)

// reliefResponse matches MapRelief in frontend/src/api/client.ts: the
// height of the land for the 3D view. Heights come from the map's geology
// (its seed), so they don't follow later edits; the view reconciles them
// with the painted tiles.
type reliefResponse struct {
	Width  int `json:"width"`
	Height int `json:"height"`
	// Elevation of every tile, 0–1000 over the map's range, row-major.
	Elevation []int `json:"elevation"`
	// The generator's thresholds on the same scale: below SeaLevel is
	// water, below ShoreLevel beach, above HighLevel rock, above PeakLevel peaks.
	SeaLevel   int `json:"seaLevel"`
	ShoreLevel int `json:"shoreLevel"`
	HighLevel  int `json:"highLevel"`
	PeakLevel  int `json:"peakLevel"`
	// Fresh is 1 for drinkable water (rivers to their mouth, lakes), else 0.
	Fresh []int `json:"fresh"`
}

func (s *Server) relief(w http.ResponseWriter, r *http.Request) {
	m, ok := s.lookup(w, r)
	if !ok {
		return
	}
	etag := `"relief-` + m.ID + "-" + strconv.FormatInt(m.UpdatedAt.UnixNano(), 36) + `"`
	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", "no-cache")
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	g := world.BuildGeoModel(m)
	lo, hi := g.Elevation[0], g.Elevation[0]
	for _, e := range g.Elevation {
		lo, hi = min(lo, e), max(hi, e)
	}
	span := hi - lo
	if span <= 0 {
		span = 1
	}
	q := func(e float64) int { return int((e - lo) / span * 1000) }
	resp := reliefResponse{
		Width: m.Width, Height: m.Height,
		Elevation: make([]int, len(g.Elevation)),
		SeaLevel:  q(g.SeaLevel), ShoreLevel: q(g.ShoreLevel), HighLevel: q(g.HighLevel), PeakLevel: q(g.PeakLevel),
		Fresh: make([]int, len(g.Elevation)),
	}
	for i, e := range g.Elevation {
		resp.Elevation[i] = q(e)
	}
	for i, f := range world.FreshWater(m) {
		if f {
			resp.Fresh[i] = 1
		}
	}
	writeJSON(w, http.StatusOK, resp)
}
