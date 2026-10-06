package sim

import (
	"bytes"
	"compress/gzip"
	"io"
	"sort"
	"sync"
)

// The replay keeps the last few minutes of the world as the observer saw it,
// after RAGE's replay recorder (control/replay/ReplayController.h: a ring of
// compressed blocks; ReplayBufferMarker.h and ReplayEventManager.h: markers
// with an importance). It records the frames that were streamed, the
// buildings whenever they changed, and the event log as markers. It never
// re-simulates and nothing in the world reads it, so it can't change what
// happens; it is not saved.

const (
	// replayEvery: one recorded frame per this many published ones (frames
	// go out ten times a real second, so two a second are kept).
	replayEvery = 5
	// replayFrames recorded frames are kept: five real minutes.
	replayFrames = 600
	// replayMarks events are kept as markers.
	replayMarks = 4000
)

// Marker importance, after MarkerImportance.
const (
	markLow    = 1
	markNormal = 2
	markHigh   = 3
)

// ReplayMark is an event on the replay's timeline.
type ReplayMark struct {
	Tick       int64   `json:"tick"`
	Time       float64 `json:"time"`
	Kind       string  `json:"kind"`
	Text       string  `json:"text"`
	CreatureID int64   `json:"creatureId,omitempty"`
	Importance int     `json:"importance"`
}

// ReplayIndex is what the recording holds: each recorded frame's tick and
// simulated time (oldest first), the building versions it needs, and the
// markers in that window.
type ReplayIndex struct {
	Frames     [][2]float64 `json:"frames"` // [tick, time]
	Structures []int64      `json:"structures"`
	Marks      []ReplayMark `json:"marks"`
}

type replayFrame struct {
	tick    int64
	time    float64
	structs int64  // building version on screen
	data    []byte // gzip of the frame JSON
}

type recorder struct {
	mu        sync.Mutex
	published int
	frames    []replayFrame // ring, oldest at start
	start     int
	structs   map[int64][]byte // building version → gzip of its `structures` message
	marks     []ReplayMark     // ring, oldest at markStart
	markStart int
}

// replayDue reports whether the frame about to be published is one to keep.
func (r *recorder) replayDue() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.published++
	return r.published%replayEvery == 0
}

func (r *recorder) knowsStructures(version int64) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.structs[version]
	return ok
}

// record keeps a frame (and the buildings if they are new). It compresses
// outside the world's lock.
func (r *recorder) record(tick int64, time float64, frame []byte, structVersion int64, structures []byte) {
	f := replayFrame{tick: tick, time: time, structs: structVersion, data: pack(frame)}
	var packed []byte
	if structures != nil {
		packed = pack(structures)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if packed != nil {
		if r.structs == nil {
			r.structs = map[int64][]byte{}
		}
		r.structs[structVersion] = packed
	}
	if len(r.frames) < replayFrames {
		r.frames = append(r.frames, f)
	} else {
		r.frames[r.start] = f
		r.start = (r.start + 1) % replayFrames
	}
	// Forget buildings no kept frame shows any more.
	used := map[int64]bool{}
	for _, fr := range r.frames {
		used[fr.structs] = true
	}
	for v := range r.structs {
		if !used[v] {
			delete(r.structs, v)
		}
	}
}

func (r *recorder) mark(m ReplayMark) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.marks) < replayMarks {
		r.marks = append(r.marks, m)
		return
	}
	r.marks[r.markStart] = m
	r.markStart = (r.markStart + 1) % replayMarks
}

func (r *recorder) ordered() []replayFrame {
	return append(append([]replayFrame{}, r.frames[r.start:]...), r.frames[:r.start]...)
}

func (r *recorder) index() ReplayIndex {
	r.mu.Lock()
	defer r.mu.Unlock()
	idx := ReplayIndex{Frames: [][2]float64{}, Structures: []int64{}, Marks: []ReplayMark{}}
	frames := r.ordered()
	for _, f := range frames {
		idx.Frames = append(idx.Frames, [2]float64{float64(f.tick), f.time})
	}
	for v := range r.structs {
		idx.Structures = append(idx.Structures, v)
	}
	sort.Slice(idx.Structures, func(i, j int) bool { return idx.Structures[i] < idx.Structures[j] })
	if len(frames) > 0 {
		first := frames[0].tick
		marks := append(append([]ReplayMark{}, r.marks[r.markStart:]...), r.marks[:r.markStart]...)
		for _, m := range marks {
			if m.Tick >= first {
				idx.Marks = append(idx.Marks, m)
			}
		}
	}
	return idx
}

// frameAt returns the recorded frame at or just before tick (the oldest if
// tick is earlier still) and the building version it shows.
func (r *recorder) frameAt(tick int64) ([]byte, int64, bool) {
	r.mu.Lock()
	frames := r.ordered()
	r.mu.Unlock()
	if len(frames) == 0 {
		return nil, 0, false
	}
	i := sort.Search(len(frames), func(i int) bool { return frames[i].tick > tick }) - 1
	i = max(i, 0)
	return unpack(frames[i].data), frames[i].structs, true
}

func (r *recorder) structuresAt(version int64) ([]byte, bool) {
	r.mu.Lock()
	data, ok := r.structs[version]
	r.mu.Unlock()
	if !ok {
		return nil, false
	}
	return unpack(data), true
}

func pack(b []byte) []byte {
	var buf bytes.Buffer
	zw, _ := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
	zw.Write(b)
	zw.Close()
	return buf.Bytes()
}

func unpack(b []byte) []byte {
	zr, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return nil
	}
	out, _ := io.ReadAll(zr)
	return out
}

// importance ranks an event for the timeline: deaths by violence, fire,
// water or a fall, and crimes stand out; births and work don't.
func importance(kind string, violent bool) int {
	switch {
	case violent:
		return markHigh
	case kind == "death", kind == "crime", kind == "climate", kind == "fire", kind == "village", kind == "milestone", kind == "genesis":
		return markHigh
	case kind == "birth", kind == "build", kind == "farming", kind == "learning":
		return markLow
	}
	return markNormal
}

// ReplayIndex lists what the replay holds.
func (s *Sim) ReplayIndex() ReplayIndex { return s.replay.index() }

// ReplayFrame returns the recorded stream frame at or before tick, and the
// building version it shows.
func (s *Sim) ReplayFrame(tick int64) ([]byte, int64, bool) { return s.replay.frameAt(tick) }

// ReplayStructures returns a recorded `structures` message.
func (s *Sim) ReplayStructures(version int64) ([]byte, bool) { return s.replay.structuresAt(version) }
