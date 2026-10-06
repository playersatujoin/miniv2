package api

import (
	"io"
	"net/http"
	"strconv"
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
// the buildings and the planted fields on connect and whenever they change
// (fields at most once a second), until the client goes away or the server
// shuts down.
func (s *Server) simStream(w http.ResponseWriter, r *http.Request) {
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

	sent, fieldsSent := int64(-1), int64(-1)
	var fieldsAt time.Time
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
		frame, next := sm.Frame()
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
