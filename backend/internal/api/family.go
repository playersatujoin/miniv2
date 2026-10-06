package api

import (
	"net/http"
	"strconv"
)

// familyDepth is how many generations up and down a family tree shows
// unless the client asks (?depth=N, at most the pedigree's memory).
const familyDepth = 4

// simFamily returns a person's family tree (ancestors and descendants up to
// depth generations), living or dead, as long as the pedigree remembers them.
func (s *Server) simFamily(w http.ResponseWriter, r *http.Request) {
	sm, ok := s.simFor(w, r)
	if !ok {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("cid"), 10, 64)
	if err != nil {
		writeError(w, http.StatusNotFound, "creature not found")
		return
	}
	depth := familyDepth
	if v := r.URL.Query().Get("depth"); v != "" {
		if depth, err = strconv.Atoi(v); err != nil || depth < 1 {
			writeError(w, http.StatusBadRequest, "depth must be a positive number")
			return
		}
	}
	tree, ok := sm.Family(id, depth)
	if !ok {
		writeError(w, http.StatusNotFound, "creature not found")
		return
	}
	writeJSON(w, http.StatusOK, tree)
}
