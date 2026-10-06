package sim

// Viewport limits transmission, never simulation. Follow retains a selected
// person when it walks past the viewport before the camera catches up.
type Viewport struct {
	MinX, MinY, MaxX, MaxY float64
	Follow                 int64
}

func (v *Viewport) contains(x, y float64) bool {
	return v == nil || x >= v.MinX && x <= v.MaxX && y >= v.MinY && y <= v.MaxY
}

func (s *Sim) ViewFrame(v *Viewport) []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.encodeViewFrame(v)
}
