// Command soak runs several worlds headless and in parallel, then writes a
// report (summary.json + summary.md) comparing their demography with
// pre-modern reference ranges.
//
//	go run ./cmd/soak -seeds 1-8 -minutes 120
//	go run ./cmd/soak -seeds 1-8 -off crime -label tanpa-kejahatan
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"miniv2/backend/internal/sim"
	"miniv2/backend/internal/world"
)

type config struct {
	Seeds          []uint64 `json:"seeds"`
	Minutes        int      `json:"minutes"`
	MapSize        int      `json:"mapSize"`
	MapSeed        uint32   `json:"mapSeed"`
	Off            []string `json:"off"`
	SecondsPerYear float64  `json:"secondsPerYear"`
	SimYears       float64  `json:"simYears"`
}

// point is one sim-minute of a run.
type point struct {
	Minute         int      `json:"minute"`
	Year           int      `json:"year"`
	Era            int      `json:"era"`
	Population     int      `json:"population"`
	Females        int      `json:"females"`
	Males          int      `json:"males"`
	Births         int      `json:"births"`
	Deaths         int      `json:"deaths"`
	MaxGeneration  int      `json:"maxGeneration"`
	AvgGeneration  float64  `json:"avgGeneration"`
	AvgBrainSize   float64  `json:"avgBrainSize"`
	AvgSkill       float64  `json:"avgSkill"`
	KnowledgeLost  int      `json:"knowledgeLost"`
	Tier           int      `json:"tier"`
	Elements       int      `json:"elements"`
	Houses         int      `json:"houses"`
	Crimes         int      `json:"crimes"`
	Kindness       int      `json:"kindness"`
	Kills          int      `json:"kills"`
	LifeExpectancy *float64 `json:"lifeExpectancy"`
	SurvivalTo15   *float64 `json:"survivalTo15"`
	TFR            *float64 `json:"tfr"`
	Gini           *float64 `json:"gini"`
}

type final struct {
	Era1LastedMinutes int                   `json:"era1LastedMinutes"` // = minutes if it never ended
	FirstHouseMinute  *int                  `json:"firstHouseMinute"`
	Eras              int                   `json:"eras"`
	Population        int                   `json:"population"`
	MaxGeneration     int                   `json:"maxGeneration"`
	Tier              int                   `json:"tier"`
	TierName          string                `json:"tierName"`
	Elements          int                   `json:"elements"`
	Houses            int                   `json:"houses"`
	Structures        int                   `json:"structures"`
	Crimes            int                   `json:"crimes"`
	Kindness          int                   `json:"kindness"`
	Kills             int                   `json:"kills"`
	Demography        sim.DemographyMetrics `json:"demography"`
	// TierGeneration: the highest generation alive when each tier was first
	// reached (tier → generation), to compare progress across time scales.
	TierGeneration map[int]int    `json:"tierGeneration"`
	KnowledgeLost  int            `json:"knowledgeLost"`
	AvgBrainSize   float64        `json:"avgBrainSize"`
	AvgSkill       float64        `json:"avgSkill"`
	SkillByAge     []sim.AgeSkill `json:"skillByAge"`
	MsPerTick      float64        `json:"msPerTick"`
	WallSeconds    float64        `json:"wallSeconds"`
}

type run struct {
	Seed   uint64  `json:"seed"`
	Final  final   `json:"final"`
	Series []point `json:"series"`
}

type report struct {
	GeneratedAt time.Time                `json:"generatedAt"`
	Label       string                   `json:"label,omitempty"`
	Config      config                   `json:"config"`
	Reference   map[string]sim.MetricRef `json:"reference"`
	Runs        []run                    `json:"runs"`
	Median      map[string]*float64      `json:"median"`
}

func main() {
	seedsFlag := flag.String("seeds", "1-8", "world seeds, e.g. 1-8 or 1,3,5-7")
	minutes := flag.Int("minutes", 120, "simulated minutes per world")
	size := flag.Int("size", 128, "map width and height")
	mapSeed := flag.Uint("mapseed", 1337, "map generator seed")
	parallel := flag.Int("parallel", max(1, runtime.NumCPU()/2), "worlds simulated at once")
	off := flag.String("off", "", "rules to switch off: crime, instincts, learning (= plasticity + culture)")
	out := flag.String("out", "", "report directory (default <repo>/reports/<timestamp>)")
	label := flag.String("label", "", "name of this run, shown in the report")
	flag.Parse()

	seeds, err := parseSeeds(*seedsFlag)
	if err != nil {
		log.Fatal(err)
	}
	opts, offList, err := parseOff(*off)
	if err != nil {
		log.Fatal(err)
	}
	if err := world.ValidateSize(*size, *size); err != nil {
		log.Fatal(err)
	}
	dir := *out
	if dir == "" {
		dir = filepath.Join(repoRoot(), "reports", time.Now().Format("20060102-150405"))
	}
	if dir, err = filepath.Abs(dir); err != nil {
		log.Fatal(err)
	}

	cfg := config{
		Seeds: seeds, Minutes: *minutes, MapSize: *size, MapSeed: uint32(*mapSeed), Off: offList,
		SecondsPerYear: sim.SecondsPerYear, SimYears: float64(*minutes) * 60 / sim.SecondsPerYear,
	}
	log.Printf("%d worlds × %d sim-minutes (≈ %.0f years), %d at a time", len(seeds), *minutes, cfg.SimYears, *parallel)

	runs := make([]run, len(seeds))
	sem := make(chan struct{}, *parallel)
	var wg sync.WaitGroup
	for i, seed := range seeds {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			runs[i] = simulate(cfg, seed, opts)
			f := runs[i].Final
			log.Printf("seed %d: era1 %dm, pop %d, tier %d, e0 %s, TFR %s, %.2f ms/tick",
				seed, f.Era1LastedMinutes, f.Population, f.Tier, fmtPtr(f.Demography.LifeExpectancy, 1),
				fmtPtr(f.Demography.TFR, 2), f.MsPerTick)
		}()
	}
	wg.Wait()

	rep := report{GeneratedAt: time.Now(), Label: *label, Config: cfg, Reference: sim.DemographyReference(), Runs: runs, Median: medians(runs)}
	if err := write(dir, rep); err != nil {
		log.Fatal(err)
	}
	fmt.Println(dir)
}

func starterMap(cfg config) *world.Map {
	m := world.Generate(world.GenOptions{Name: "Starter Island", Width: cfg.MapSize, Height: cfg.MapSize, Seed: cfg.MapSeed})
	m.ID = "0000000000000000"
	return m
}

func simulate(cfg config, seed uint64, opts sim.Options) run {
	s := sim.NewWithOptions(starterMap(cfg), seed, opts)
	r := run{Seed: seed}
	era1 := -1
	var firstHouse *int
	tierGen := map[int]int{}
	start := time.Now()
	for minute := 1; minute <= cfg.Minutes; minute++ {
		s.Advance(60 * sim.TicksPerSecond)
		info := s.Info()
		d := s.Demography().Current
		if era1 < 0 && info.Era > 1 {
			era1 = minute
		}
		if firstHouse == nil && info.Houses > 0 {
			firstHouse = &minute
		}
		for t := 1; t <= info.Tier; t++ {
			if _, ok := tierGen[t]; !ok {
				tierGen[t] = info.MaxGeneration
			}
		}
		var skill float64
		if n := len(info.History); n > 0 {
			skill = info.History[n-1].AvgSkill
		}
		r.Series = append(r.Series, point{
			Minute: minute, Year: info.Year, Era: info.Era, Population: info.Population,
			Females: info.Females, Males: info.Males, Births: info.Births, Deaths: info.Deaths,
			MaxGeneration: info.MaxGeneration, AvgGeneration: info.AvgGeneration, AvgBrainSize: info.AvgBrainSize,
			AvgSkill: skill, KnowledgeLost: info.KnowledgeLost, Tier: info.Tier, Elements: info.ElementsDiscovered,
			Houses: info.Houses, Crimes: info.Crimes, Kindness: info.Kindness, Kills: info.Kills,
			LifeExpectancy: d.LifeExpectancy, SurvivalTo15: d.SurvivalTo15, TFR: d.TFR, Gini: d.Gini,
		})
	}
	wall := time.Since(start)
	if era1 < 0 {
		era1 = cfg.Minutes
	}
	info := s.Info()
	r.Final = final{
		Era1LastedMinutes: era1, FirstHouseMinute: firstHouse, Eras: info.Era, Population: info.Population,
		MaxGeneration: info.MaxGeneration, Tier: info.Tier, TierName: info.TierName, Elements: info.ElementsDiscovered,
		Houses: info.Houses, Structures: info.Structures, Crimes: info.Crimes, Kindness: info.Kindness, Kills: info.Kills,
		Demography:     s.Demography().Current,
		TierGeneration: tierGen,
		KnowledgeLost:  info.KnowledgeLost,
		AvgBrainSize:   info.AvgBrainSize,
		SkillByAge:     s.SkillProfile(),
		AvgSkill:       r.Series[len(r.Series)-1].AvgSkill,
		MsPerTick:      float64(wall.Microseconds()) / 1000 / float64(cfg.Minutes*60*sim.TicksPerSecond),
		WallSeconds:    wall.Seconds(),
	}
	return r
}

func parseSeeds(spec string) ([]uint64, error) {
	var out []uint64
	for _, part := range strings.Split(spec, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		lo, hi, isRange := strings.Cut(part, "-")
		a, err := strconv.ParseUint(lo, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("bad seed %q", part)
		}
		b := a
		if isRange {
			if b, err = strconv.ParseUint(hi, 10, 64); err != nil || b < a {
				return nil, fmt.Errorf("bad seed range %q", part)
			}
		}
		for v := a; v <= b; v++ {
			out = append(out, v)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no seeds in %q", spec)
	}
	return out, nil
}

func parseOff(spec string) (sim.Options, []string, error) {
	var opts sim.Options
	list := []string{}
	for _, name := range strings.Split(spec, ",") {
		switch name = strings.TrimSpace(name); name {
		case "":
			continue
		case "crime":
			opts.NoCrime = true
		case "instincts":
			opts.NoInstincts = true
		case "learning":
			opts.NoLearning = true
		case "plasticity":
			opts.NoPlasticity = true
		case "culture":
			opts.NoCulture = true
		default:
			return opts, nil, fmt.Errorf("unknown rule %q (known: crime, instincts, learning, plasticity, culture)", name)
		}
		list = append(list, name)
	}
	return opts, list, nil
}

// repoRoot finds the repository (the directory holding backend/go.mod),
// searching up from the working directory, then from this source file.
func repoRoot() string {
	isRoot := func(dir string) bool {
		_, err := os.Stat(filepath.Join(dir, "backend", "go.mod"))
		return err == nil
	}
	if wd, err := os.Getwd(); err == nil {
		for dir := wd; ; dir = filepath.Dir(dir) {
			if isRoot(dir) {
				return dir
			}
			if filepath.Dir(dir) == dir {
				break
			}
		}
	}
	if _, file, _, ok := runtime.Caller(0); ok {
		if dir := filepath.Join(filepath.Dir(file), "..", "..", ".."); isRoot(dir) {
			return filepath.Clean(dir)
		}
	}
	wd, _ := os.Getwd()
	return wd
}

func write(dir string, rep report) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "summary.json"), data, 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "summary.md"), []byte(markdown(rep)), 0o644)
}
