package sim

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestViewportFiltersTransmissionWithoutChangingWorld(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	a := person(s, Male, "A", x, y)
	b := person(s, Female, "B", x+3, y)
	before, err := s.MarshalState()
	if err != nil {
		t.Fatal(err)
	}
	v := &Viewport{MinX: float64(x), MinY: float64(y), MaxX: float64(x + 1), MaxY: float64(y + 1)}
	decode := func(data []byte) []json.RawMessage {
		var frame struct {
			Creatures []json.RawMessage `json:"c"`
		}
		if err := json.Unmarshal(data, &frame); err != nil {
			t.Fatal(err)
		}
		return frame.Creatures
	}
	if cs := decode(s.ViewFrame(v)); len(cs) != 1 {
		t.Fatalf("viewport contains %d people, want only %d", len(cs), a.ID)
	}
	v.Follow = b.ID
	if cs := decode(s.ViewFrame(v)); len(cs) != 2 {
		t.Fatal("followed person was lost outside camera")
	}
	if cs := decode(s.ViewFrame(nil)); len(cs) != 2 {
		t.Fatal("full stream lost entities")
	}
	after, err := s.MarshalState()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("camera changed simulation state")
	}
}
