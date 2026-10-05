package main

import (
	"math"
	"testing"
)

func TestParseSeedsAndOff(t *testing.T) {
	seeds, err := parseSeeds("1-3, 7,9-10")
	if err != nil || len(seeds) != 6 || seeds[0] != 1 || seeds[3] != 7 || seeds[5] != 10 {
		t.Fatalf("seeds %v %v", seeds, err)
	}
	for _, bad := range []string{"", "x", "5-2"} {
		if _, err := parseSeeds(bad); err == nil {
			t.Errorf("%q should fail", bad)
		}
	}
	opts, list, err := parseOff("crime, instincts, learning")
	if err != nil || !opts.NoCrime || !opts.NoInstincts || !opts.NoLearning || len(list) != 3 {
		t.Fatalf("off: %+v %v %v", opts, list, err)
	}
	if _, _, err := parseOff("gravity"); err == nil {
		t.Error("unknown rule should fail")
	}
}

func TestWilsonAndMedian(t *testing.T) {
	lo, hi := wilson(5, 8)
	if math.Abs(lo-0.306) > 0.01 || math.Abs(hi-0.863) > 0.01 {
		t.Fatalf("wilson(5,8) = %.3f–%.3f", lo, hi)
	}
	if m := median([]float64{3, 1, 2}); *m != 2 {
		t.Fatalf("median %v", *m)
	}
	if m := median([]float64{4, 1, 2, 3}); *m != 2.5 {
		t.Fatalf("median %v", *m)
	}
	if median(nil) != nil {
		t.Fatal("median of nothing should be nil")
	}
}

func TestRepoRootFound(t *testing.T) {
	if root := repoRoot(); root == "" {
		t.Fatal("no repo root")
	}
}
