package sim

import (
	"math"
	"slices"
)

// Brains is the island's brains at a glance, for the observer's Neuron view:
// how big they are (inherited and grown), how fast they are growing and being
// pruned, the genes behind it, and the biggest brains alive.
type Brains struct {
	Population   int     `json:"population"`
	AvgTotal     float64 `json:"avgTotal"`
	AvgInherited float64 `json:"avgInherited"`
	AvgGrown     float64 `json:"avgGrown"`
	Max          int     `json:"max"`
	// Histogram counts people by brain size in bins of HistogramBin neurons.
	Histogram    []int `json:"histogram"`
	HistogramBin int   `json:"histogramBin"`
	// Neurons grown and pruned in all lives, and per simulated year lately.
	Grown         int64   `json:"grown"`
	Pruned        int64   `json:"pruned"`
	GrownPerYear  float64 `json:"grownPerYear"`
	PrunedPerYear float64 `json:"prunedPerYear"`
	// Averages of the genes and states that drive it.
	AvgNeurogenesis float64      `json:"avgNeurogenesis"`
	AvgNovelty      float64      `json:"avgNovelty"`
	AvgTau          float64      `json:"avgTau"`
	SlowNeurons     float64      `json:"slowNeurons"` // share of inherited neurons with tau ≥ 3 (a working memory)
	Biggest         []BrainEntry `json:"biggest"`
	History         []BrainPoint `json:"history"`
	SecondsPerYear  float64      `json:"secondsPerYear"`
	// Limit is the most neurons a brain can have; 0: no limit but energy.
	Limit int `json:"limit"`
}

type BrainEntry struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	Sex       string  `json:"sex"`
	Age       float64 `json:"age"` // years
	Inherited int     `json:"inherited"`
	Grown     int     `json:"grown"`
	Novelty   float64 `json:"novelty"`
}

type BrainPoint struct {
	Time     float64 `json:"time"`
	AvgTotal float64 `json:"avgTotal"`
	AvgGrown float64 `json:"avgGrown"`
	Max      int     `json:"max"`
}

const histogramBin = 8

func (s *Sim) Brains() Brains {
	s.mu.Lock()
	defer s.mu.Unlock()
	b := Brains{
		Population:     len(s.creatures),
		Histogram:      []int{},
		HistogramBin:   histogramBin,
		Grown:          s.grown,
		Pruned:         s.pruned,
		SecondsPerYear: SecondsPerYear,
		Limit:          0, // none: energy limits brains
	}
	if b.Population > 0 {
		var inherited, grown, gene, novelty, tau, slow, neurons float64
		for _, c := range s.creatures {
			g := c.Genome
			inherited += float64(g.Hidden)
			grown += float64(c.Mind.Grown)
			gene += g.Traits.Neurogenesis
			novelty += c.Mind.Novelty
			for _, t := range g.Tau {
				tau += float64(t)
				if t >= 3 {
					slow++
				}
			}
			neurons += float64(g.Hidden)
			size := brainSize(c)
			b.Max = max(b.Max, size)
			for len(b.Histogram) <= size/histogramBin {
				b.Histogram = append(b.Histogram, 0)
			}
			b.Histogram[size/histogramBin]++
		}
		n := float64(b.Population)
		b.AvgInherited = r1(inherited / n)
		b.AvgGrown = r1(grown / n)
		b.AvgTotal = r1((inherited + grown) / n)
		b.AvgNeurogenesis = r3(gene / n)
		b.AvgNovelty = r3(novelty / n)
		b.AvgTau = r3(tau / neurons)
		b.SlowNeurons = r3(slow / neurons)
	}
	// The biggest brains alive, most grown first among equals.
	ranked := slices.Clone(s.creatures)
	slices.SortStableFunc(ranked, func(a, c *Creature) int {
		if d := brainSize(c) - brainSize(a); d != 0 {
			return d
		}
		return c.Mind.Grown - a.Mind.Grown
	})
	for _, c := range ranked[:min(8, len(ranked))] {
		b.Biggest = append(b.Biggest, BrainEntry{
			ID: c.ID, Name: c.Name, Sex: c.Sex.String(), Age: r1(s.age(c) / SecondsPerYear),
			Inherited: c.Genome.Hidden, Grown: c.Mind.Grown, Novelty: r3(c.Mind.Novelty),
		})
	}
	for _, h := range s.history {
		b.History = append(b.History, BrainPoint{Time: h.Time, AvgTotal: h.AvgBrainSize, AvgGrown: h.AvgGrown, Max: h.MaxBrain})
	}
	// Growth and pruning over the last ten years or so of history.
	if n := len(s.history); n >= 2 {
		last := s.history[n-1]
		first := s.history[max(0, n-1-10)]
		for i := n - 2; i >= 0 && last.Time-s.history[i].Time <= 10*SecondsPerYear; i-- {
			first = s.history[i]
		}
		if years := (last.Time - first.Time) / SecondsPerYear; years > 0 && first.Grown <= last.Grown {
			b.GrownPerYear = r1(float64(last.Grown-first.Grown) / years)
			b.PrunedPerYear = r1(float64(last.Pruned-first.Pruned) / years)
		}
	}
	return b
}

func r1(v float64) float64 { return math.Round(v*10) / 10 }
