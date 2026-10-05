// Package api exposes the map store over HTTP/JSON.
package api

import (
	"encoding/json"
	"errors"
	"image"
	"image/color"
	"image/png"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"

	"miniv2/backend/internal/sim"
	"miniv2/backend/internal/store"
	"miniv2/backend/internal/world"
)

const maxBodyBytes = 4 << 20

type Server struct {
	store *store.FileStore
	sims  *sim.Manager
}

// New serves the map API; every stored map has a living world in sims.
func New(st *store.FileStore, sims *sim.Manager) http.Handler {
	s := &Server{store: st, sims: sims}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /api/tiles", s.tiles)
	mux.HandleFunc("GET /api/maps", s.listMaps)
	mux.HandleFunc("POST /api/maps", s.createMap)
	mux.HandleFunc("GET /api/maps/{id}", s.getMap)
	mux.HandleFunc("PUT /api/maps/{id}", s.updateMap)
	mux.HandleFunc("DELETE /api/maps/{id}", s.deleteMap)
	mux.HandleFunc("GET /api/maps/{id}/preview.png", s.preview)
	mux.HandleFunc("GET /api/maps/{id}/geology", s.geology)
	mux.HandleFunc("GET /api/maps/{id}/sim", s.simInfo)
	mux.HandleFunc("PUT /api/maps/{id}/sim/speed", s.setSimSpeed)
	mux.HandleFunc("GET /api/maps/{id}/sim/creatures/{cid}", s.simCreature)
	mux.HandleFunc("GET /api/maps/{id}/sim/knowledge", s.simKnowledge)
	mux.HandleFunc("GET /api/maps/{id}/sim/demography", s.simDemography)
	mux.HandleFunc("GET /api/maps/{id}/sim/mined", s.simMined)
	mux.HandleFunc("GET /api/maps/{id}/sim/stream", s.simStream)
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotFound, "not found")
	})
	return mux
}

func (s *Server) tiles(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string][]world.TileDef{
		"ground":  world.GroundTiles,
		"objects": world.ObjectTiles,
	})
}

func (s *Server) listMaps(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.store.List())
}

type createRequest struct {
	Name   string  `json:"name"`
	Width  int     `json:"width"`
	Height int     `json:"height"`
	Seed   *uint32 `json:"seed"`
}

func (s *Server) createMap(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if !decode(w, r, &req) {
		return
	}
	name, err := world.NormalizeName(req.Name)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := world.ValidateSize(req.Width, req.Height); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	seed := rand.Uint32()
	if req.Seed != nil {
		seed = *req.Seed
	}

	m := world.Generate(world.GenOptions{Name: name, Width: req.Width, Height: req.Height, Seed: seed})
	m.ID = store.NewID()
	m.CreatedAt = time.Now().UTC()
	m.UpdatedAt = m.CreatedAt
	if err := s.store.Save(m); err != nil {
		internalError(w, r, err)
		return
	}
	s.sims.Start(m)
	writeJSON(w, http.StatusCreated, m)
}

func (s *Server) getMap(w http.ResponseWriter, r *http.Request) {
	m, ok := s.lookup(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, m)
}

// updateRequest carries the editable parts of a map. Size, seed and ID are fixed.
type updateRequest struct {
	Name   string       `json:"name"`
	Spawn  world.Point  `json:"spawn"`
	Layers world.Layers `json:"layers"`
}

func (s *Server) updateMap(w http.ResponseWriter, r *http.Request) {
	cur, ok := s.lookup(w, r)
	if !ok {
		return
	}
	var req updateRequest
	if !decode(w, r, &req) {
		return
	}

	next := *cur
	next.Spawn = req.Spawn
	next.Layers = req.Layers
	name, err := world.NormalizeName(req.Name)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	next.Name = name
	if err := next.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	next.UpdatedAt = time.Now().UTC()
	if err := s.store.Save(&next); err != nil {
		internalError(w, r, err)
		return
	}
	s.sims.Update(&next)
	writeJSON(w, http.StatusOK, &next)
}

func (s *Server) deleteMap(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	err := s.store.Delete(id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		internalError(w, r, err)
		return
	}
	s.sims.Remove(id)
	w.WriteHeader(http.StatusNoContent)
}

// preview renders the map as a PNG thumbnail, `scale` pixels per tile.
func (s *Server) preview(w http.ResponseWriter, r *http.Request) {
	m, ok := s.lookup(w, r)
	if !ok {
		return
	}
	scale := 2
	if v := r.URL.Query().Get("scale"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 8 {
			writeError(w, http.StatusBadRequest, "scale must be an integer from 1 to 8")
			return
		}
		scale = n
	}

	etag := `"` + m.ID + "-" + strconv.FormatInt(m.UpdatedAt.UnixNano(), 36) + "-" + strconv.Itoa(scale) + `"`
	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", "no-cache")
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	ground := palette(world.GroundTiles)
	objects := palette(world.ObjectTiles)
	img := image.NewRGBA(image.Rect(0, 0, m.Width*scale, m.Height*scale))
	for ty := range m.Height {
		for tx := range m.Width {
			i := ty*m.Width + tx
			g := ground[m.Layers.Ground[i]]
			o := m.Layers.Objects[i]
			for py := range scale {
				for px := range scale {
					c := g
					// Small scales fill the whole tile; larger ones keep a ground-coloured border.
					border := scale >= 3 && (px == 0 || py == 0 || px == scale-1 || py == scale-1)
					if o != world.None && !border {
						c = objects[o]
					}
					img.SetRGBA(tx*scale+px, ty*scale+py, c)
				}
			}
		}
	}
	sx, sy := m.Spawn.X*scale, m.Spawn.Y*scale
	for d := range scale {
		img.SetRGBA(sx+d, sy+d, color.RGBA{230, 40, 40, 255})
	}

	w.Header().Set("Content-Type", "image/png")
	png.Encode(w, img)
}

func (s *Server) lookup(w http.ResponseWriter, r *http.Request) (*world.Map, bool) {
	m, err := s.store.Get(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return nil, false
	}
	return m, true
}

func palette(defs []world.TileDef) []color.RGBA {
	out := make([]color.RGBA, len(defs))
	for i, d := range defs {
		out[i] = parseHex(d.Color)
	}
	return out
}

// parseHex parses "#rrggbb" or "#rrggbbaa".
func parseHex(s string) color.RGBA {
	v, _ := strconv.ParseUint(s[1:], 16, 32)
	if len(s) == 9 {
		return color.RGBA{uint8(v >> 24), uint8(v >> 16), uint8(v >> 8), uint8(v)}
	}
	return color.RGBA{uint8(v >> 16), uint8(v >> 8), uint8(v), 255}
}

func decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func internalError(w http.ResponseWriter, r *http.Request, err error) {
	slog.Error("request failed", "method", r.Method, "path", r.URL.Path, "err", err)
	writeError(w, http.StatusInternalServerError, "internal server error")
}
