package api

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"miniv2/backend/internal/sim"
)

func (s *Server) simFor(w http.ResponseWriter, r *http.Request) (*sim.Sim, bool) {
	sm, ok := s.sims.Get(r.PathValue("id"))
	if !ok {
		writeError(w, http.StatusNotFound, "map not found")
	}
	return sm, ok
}

func (s *Server) simInfo(w http.ResponseWriter, r *http.Request) {
	sm, ok := s.simFor(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, sm.Info())
}

func (s *Server) setSimSpeed(w http.ResponseWriter, r *http.Request) {
	sm, ok := s.simFor(w, r)
	if !ok {
		return
	}
	var req struct {
		Speed *int `json:"speed"`
	}
	if !decode(w, r, &req) {
		return
	}
	if req.Speed == nil {
		writeError(w, http.StatusBadRequest, "speed is required")
		return
	}
	if err := sm.SetSpeed(*req.Speed); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sm.Info())
}

func (s *Server) simKnowledge(w http.ResponseWriter, r *http.Request) {
	sm, ok := s.simFor(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, sm.Knowledge())
}

// simBrains summarises the island's brains (sizes, growth, pruning, the biggest) for the Neuron view.
func (s *Server) simBrains(w http.ResponseWriter, r *http.Request) {
	sm, ok := s.simFor(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, sm.Brains())
}

func (s *Server) simDemography(w http.ResponseWriter, r *http.Request) {
	sm, ok := s.simFor(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, sm.Demography())
}

// simMined lists dug-out mineral deposits (old pits) and felled trees
// (stumps) of a living world.
func (s *Server) simMined(w http.ResponseWriter, r *http.Request) {
	sm, ok := s.simFor(w, r)
	if !ok {
		return
	}
	version, tiles, logged := sm.MinedOut()
	writeJSON(w, http.StatusOK, map[string]any{"version": version, "tiles": tiles, "logged": logged})
}

// simEcology reports the land: seasons and weather, animals, fields and food.
func (s *Server) simEcology(w http.ResponseWriter, r *http.Request) {
	sm, ok := s.simFor(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, sm.Ecology())
}

// simHealth reports disease: who is ill, the epidemic curve, and the
// mosquito, worm and fouled-water maps.
func (s *Server) simHealth(w http.ResponseWriter, r *http.Request) {
	sm, ok := s.simFor(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, sm.Health())
}

// simVillages lists the villages: their land, people and leaders.
func (s *Server) simVillages(w http.ResponseWriter, r *http.Request) {
	sm, ok := s.simFor(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, sm.VillagesView())
}

// simReplay lists the recorded minutes: frame ticks and times, building
// versions and timeline markers.
func (s *Server) simReplay(w http.ResponseWriter, r *http.Request) {
	sm, ok := s.simFor(w, r)
	if !ok {
		return
	}
	writeCompressed(w, r, func(out io.Writer) error { return json.NewEncoder(out).Encode(sm.ReplayIndex()) })
}

// writeCompressed sends a JSON body gzipped when the client takes it: the
// replay's index and frames run to hundreds of kilobytes of very repetitive
// JSON and shrink about tenfold.
func writeCompressed(w http.ResponseWriter, r *http.Request, write func(io.Writer) error) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("Vary", "Accept-Encoding")
	if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		write(w)
		return
	}
	w.Header().Set("Content-Encoding", "gzip")
	zw, _ := gzip.NewWriterLevel(w, gzip.BestSpeed)
	write(zw)
	zw.Close()
}

// simReplayFrame returns the recorded stream frame at or just before ?tick=,
// in the stream's own format, with the building version it shows in the
// X-Structures-Version header.
func (s *Server) simReplayFrame(w http.ResponseWriter, r *http.Request) {
	sm, ok := s.simFor(w, r)
	if !ok {
		return
	}
	tick, err := strconv.ParseInt(r.URL.Query().Get("tick"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "tick must be an integer")
		return
	}
	frame, version, ok := sm.ReplayFrame(tick)
	if !ok {
		writeError(w, http.StatusNotFound, "nothing recorded yet")
		return
	}
	w.Header().Set("X-Structures-Version", strconv.FormatInt(version, 10))
	writeCompressed(w, r, func(out io.Writer) error { _, err := out.Write(frame); return err })
}

// simReplayStructures returns a recorded `structures` message.
func (s *Server) simReplayStructures(w http.ResponseWriter, r *http.Request) {
	sm, ok := s.simFor(w, r)
	if !ok {
		return
	}
	v, err := strconv.ParseInt(r.PathValue("v"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "version must be an integer")
		return
	}
	data, ok := sm.ReplayStructures(v)
	if !ok {
		writeError(w, http.StatusNotFound, "buildings not recorded")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (s *Server) simCreature(w http.ResponseWriter, r *http.Request) {
	sm, ok := s.simFor(w, r)
	if !ok {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("cid"), 10, 64)
	if err != nil {
		writeError(w, http.StatusNotFound, "creature not found")
		return
	}
	c, ok := sm.Creature(id)
	if !ok {
		writeError(w, http.StatusNotFound, "creature not found")
		return
	}
	writeJSON(w, http.StatusOK, c)
}

// simStream pushes every published world frame as a Server-Sent Event, plus
// the buildings, the planted fields, the rivers' water, the villages and the
// scorched land on connect and whenever they change (fields and scorch at
// most once a second, water at most every 0.4 s), and the mosquitoes every
// two seconds, until the client goes away or the server shuts down.
func (s *Server) simStream(w http.ResponseWriter, r *http.Request) {
	view, err := parseViewport(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	sm, ok := s.simFor(w, r)
	if !ok {
		return
	}
	rc := http.NewResponseController(w)
	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "retry: 2000\n\n")
	if rc.Flush() != nil {
		return
	}

	sent, fieldsSent, waterSent := int64(-1), int64(-1), int64(-1)
	villagesSent, burntSent := int64(-1), int64(-1)
	var fieldsAt, waterAt, bugsAt, burntAt time.Time
	for {
		if v := sm.StructureVersion(); v != sent {
			var data []byte
			data, sent = sm.Structures()
			io.WriteString(w, "event: structures\ndata: ")
			w.Write(data)
			io.WriteString(w, "\n\n")
		}
		if v := sm.FieldsVersion(); v != fieldsSent && time.Since(fieldsAt) >= time.Second {
			var data []byte
			data, fieldsSent = sm.Fields()
			fieldsAt = time.Now()
			io.WriteString(w, "event: fields\ndata: ")
			w.Write(data)
			io.WriteString(w, "\n\n")
		}
		if v := sm.WaterVersion(); v != waterSent && time.Since(waterAt) >= 400*time.Millisecond {
			var data []byte
			data, waterSent = sm.Water()
			waterAt = time.Now()
			io.WriteString(w, "event: water\ndata: ")
			w.Write(data)
			io.WriteString(w, "\n\n")
		}
		if v := sm.VillageVersion(); v != villagesSent {
			var data []byte
			data, villagesSent = sm.Villages()
			io.WriteString(w, "event: villages\ndata: ")
			w.Write(data)
			io.WriteString(w, "\n\n")
		}
		if v := sm.BurntVersion(); v != burntSent && time.Since(burntAt) >= time.Second {
			var data []byte
			data, burntSent = sm.Burnt()
			burntAt = time.Now()
			io.WriteString(w, "event: burnt\ndata: ")
			w.Write(data)
			io.WriteString(w, "\n\n")
		}
		if time.Since(bugsAt) >= 2*time.Second {
			bugsAt = time.Now()
			io.WriteString(w, "event: mosquitoes\ndata: ")
			w.Write(sm.Mosquitoes())
			io.WriteString(w, "\n\n")
		}
		frame, next := sm.Frame()
		if view != nil {
			frame = sm.ViewFrame(view)
		}
		if frame != nil {
			io.WriteString(w, "event: frame\ndata: ")
			w.Write(frame)
			io.WriteString(w, "\n\n")
		}
		if rc.Flush() != nil {
			return
		}
		select {
		case <-next:
		case <-r.Context().Done():
			return
		}
	}
}

func parseViewport(r *http.Request) (*sim.Viewport, error) {
	q := r.URL.Query()
	keys := []string{"x0", "y0", "x1", "y1"}
	present := false
	for _, key := range keys {
		present = present || q.Has(key)
	}
	if !present {
		return nil, nil
	}
	var n [4]float64
	for i, key := range keys {
		v, err := strconv.ParseFloat(q.Get(key), 64)
		if err != nil || math.IsNaN(v) || math.IsInf(v, 0) || math.Abs(v) > 1e6 {
			return nil, fmt.Errorf("invalid viewport %s", key)
		}
		n[i] = v
	}
	if n[2] <= n[0] || n[3] <= n[1] {
		return nil, fmt.Errorf("viewport must have positive width and height")
	}
	follow := int64(0)
	if q.Has("follow") {
		v, err := strconv.ParseInt(q.Get("follow"), 10, 64)
		if err != nil || v < 0 {
			return nil, fmt.Errorf("invalid follow id")
		}
		follow = v
	}
	return &sim.Viewport{MinX: n[0], MinY: n[1], MaxX: n[2], MaxY: n[3], Follow: follow}, nil
}
