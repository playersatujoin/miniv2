package sim

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
)

func TestReplayKeepsTheLastFramesAndFindsTheOneBefore(t *testing.T) {
	var r recorder
	for i := range replayFrames + 150 {
		var structs []byte
		if !r.knowsStructures(int64(i / 100)) {
			structs = []byte(fmt.Sprintf(`{"v":%d,"s":[]}`, i/100))
		}
		r.record(int64(i*10), float64(i)/2, []byte(fmt.Sprintf(`{"t":%d}`, i*10)), int64(i/100), structs)
	}
	idx := r.index()
	if len(idx.Frames) != replayFrames || idx.Frames[0][0] != 1500 || idx.Frames[len(idx.Frames)-1][0] != float64((replayFrames+149)*10) {
		t.Fatalf("kept %d frames from tick %v", len(idx.Frames), idx.Frames[0])
	}
	frame, v, ok := r.frameAt(2345)
	if !ok || !bytes.Equal(frame, []byte(`{"t":2340}`)) || v != 2 {
		t.Fatalf("frame at 1234: %s v%d", frame, v)
	}
	if frame, _, _ := r.frameAt(5); !bytes.Equal(frame, []byte(`{"t":1500}`)) {
		t.Fatalf("before the recording: %s", frame)
	}
	if _, ok := r.structuresAt(0); ok {
		t.Fatal("buildings no kept frame shows were not forgotten")
	}
	if data, ok := r.structuresAt(6); !ok || !bytes.Equal(data, []byte(`{"v":6,"s":[]}`)) {
		t.Fatalf("buildings v6: %s", data)
	}
}

func TestReplayMarksTheRecordedWindowOnly(t *testing.T) {
	var r recorder
	r.mark(ReplayMark{Tick: 5, Kind: "birth"})
	r.record(10, 0.5, []byte(`{}`), 0, []byte(`{}`))
	r.mark(ReplayMark{Tick: 12, Kind: "death", Importance: importance("death", true)})
	idx := r.index()
	if len(idx.Marks) != 1 || idx.Marks[0].Kind != "death" || idx.Marks[0].Importance != markHigh {
		t.Fatalf("marks %+v", idx.Marks)
	}
	if _, err := json.Marshal(idx); err != nil {
		t.Fatal(err)
	}
}
