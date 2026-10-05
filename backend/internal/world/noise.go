package world

import "math"

// noise2D is seeded 2D value noise with smoothstep interpolation.
type noise2D struct{ seed uint64 }

func hash2(seed uint64, x, y int) float64 {
	h := seed ^ uint64(int64(x))*0x9E3779B97F4A7C15 ^ uint64(int64(y))*0xC2B2AE3D27D4EB4F
	h ^= h >> 33
	h *= 0xff51afd7ed558ccd
	h ^= h >> 33
	h *= 0xc4ceb9fe1a85ec53
	h ^= h >> 33
	return float64(h>>11) / float64(1<<53)
}

func smooth(t float64) float64 { return t * t * (3 - 2*t) }

func lerp(a, b, t float64) float64 { return a + (b-a)*t }

func (n noise2D) at(x, y float64) float64 {
	x0, y0 := math.Floor(x), math.Floor(y)
	ix, iy := int(x0), int(y0)
	fx, fy := smooth(x-x0), smooth(y-y0)

	a := hash2(n.seed, ix, iy)
	b := hash2(n.seed, ix+1, iy)
	c := hash2(n.seed, ix, iy+1)
	d := hash2(n.seed, ix+1, iy+1)
	return lerp(lerp(a, b, fx), lerp(c, d, fx), fy)
}

// fbm sums octaves of noise, each at double the frequency and half the amplitude.
// The result is roughly in [0, 1].
func (n noise2D) fbm(x, y float64, octaves int) float64 {
	sum, amp, freq, norm := 0.0, 1.0, 1.0, 0.0
	for i := range octaves {
		// Offset each octave so lattice points don't line up.
		off := float64(i) * 17.31
		sum += amp * n.at(x*freq+off, y*freq-off)
		norm += amp
		amp *= 0.5
		freq *= 2
	}
	return sum / norm
}
