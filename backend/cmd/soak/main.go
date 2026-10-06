// Command soak runs several worlds headless and in parallel, then writes a
// report (summary.json.gz + summary.md) comparing their demography with
// pre-modern reference ranges.
//
//	go run ./cmd/soak -seeds 1-8 -minutes 120
//	go run ./cmd/soak -seeds 1-8 -off crime -label tanpa-kejahatan
//	go run ./cmd/soak -seeds 1-8 -minutes 240 -off humans -label satwa  # the land without people
//	go run ./cmd/soak -render ../reports/fase2/on                        # rewrite summary.md from the report's data
package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"sync"
	"time"

	"miniv2/backend/internal/ecology"
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
	// Fase 2.
	Animals   [ecology.SpeciesCount]int `json:"animals"`
	Livestock int                       `json:"livestock"`
	Plots     int                       `json:"plots"`
	Forest    float64                   `json:"forest"`
	FoodStock int                       `json:"foodStock"`
	ENSO      string                    `json:"enso"`
}

// ecoFinal sums up the land over a run.
type ecoFinal struct {
	PeakPopulation int                           `json:"peakPopulation"`
	CapacityHits   int                           `json:"capacityHits"`
	Animals        [ecology.SpeciesCount]int     `json:"animals"`
	AnimalsMin     [ecology.SpeciesCount]int     `json:"animalsMin"`
	AnimalsMax     [ecology.SpeciesCount]int     `json:"animalsMax"`
	Presence       [ecology.SpeciesCount]float64 `json:"presence"` // share of sampled minutes with the species alive
	Extinctions    int                           `json:"extinctions"`
	Arrivals       int                           `json:"arrivals"`
	Forest         float64                       `json:"forest"`
	MaxPlots       int                           `json:"maxPlots"`
	Livestock      int                           `json:"livestock"`
	FarmingMinute  *int                          `json:"farmingMinute"` // when farming was first invented
	HerdingMinute  *int                          `json:"herdingMinute"`
	Years          []ecology.YearRecord          `json:"years"`
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
	// Fase 3b: sanitation built by the end, and respiratory strains seen.
	Latrines int `json:"latrines"`
	Wells    int `json:"wells"`
	Strains  int `json:"strains"`
	// TierGeneration: the highest generation alive when each tier was first
	// reached (tier → generation), to compare progress across time scales.
	TierGeneration map[int]int    `json:"tierGeneration"`
	KnowledgeLost  int            `json:"knowledgeLost"`
	AvgBrainSize   float64        `json:"avgBrainSize"`
	AvgSkill       float64        `json:"avgSkill"`
	SkillByAge     []sim.AgeSkill `json:"skillByAge"`
	MsPerTick      float64        `json:"msPerTick"`
	WallSeconds    float64        `json:"wallSeconds"`
	Ecology        ecoFinal       `json:"ecology"`
	// Engine adaptation II: exchanges, villages, fire, water and stimuli at the end.
	Engine sim.EngineInfo `json:"engine"`
	// Fase 3d: nutrition over the run.
	Nutrition nutritionFinal `json:"nutrition"`
	// Fase 3c: inbreeding by generation and over time, and its outcomes.
	Genetics sim.GeneticsInfo `json:"genetics"`
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
	off := flag.String("off", "", "rules to switch off: crime, instincts, learning (= plasticity + culture), humans, farming, climate, fauna, space (personal space), neurogenesis, water (rivers always full), disease, sanitation, attachment, perception (local memory), navigation (action routes), ai_budget (think every motor tick), stimuli (lasting shocks and bodies), affect (moods), exchange (talk, gossip, trade), fire, mobility (water and steep ground), villages, nutrition (food is only calories), sharing (carcasses, eating from others' food, bawon, village granaries, knowing the land), genetics (recessive disorders harmless, kin senses 0)")
	out := flag.String("out", "", "report directory (default <repo>/reports/<timestamp>)")
	label := flag.String("label", "", "name of this run, shown in the report")
	cpuprofile := flag.String("cpuprofile", "", "write a CPU profile of the whole run to this file")
	render := flag.String("render", "", "only rewrite summary.md of this report directory from its data")
	flag.Parse()
	if *render != "" {
		if err := rerender(*render); err != nil {
			log.Fatal(err)
		}
		return
	}
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal(err)
		}
		defer pprof.StopCPUProfile()
	}

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
	var eco ecoFinal
	for i := range eco.AnimalsMin {
		eco.AnimalsMin[i] = -1
	}
	lastYear := 0
	var nut nutritionTally
	start := time.Now()
	for minute := 1; minute <= cfg.Minutes; minute++ {
		s.Advance(60 * sim.TicksPerSecond)
		info := s.Info()
		d := s.Demography().Current
		ev := s.Ecology()
		nut.add(s.Nutrition())
		var animals [ecology.SpeciesCount]int
		for i, sp := range ev.Species {
			n := sp.Wild + sp.Tame
			animals[i] = n
			if eco.AnimalsMin[i] < 0 || n < eco.AnimalsMin[i] {
				eco.AnimalsMin[i] = n
			}
			eco.AnimalsMax[i] = max(eco.AnimalsMax[i], n)
			if n > 0 {
				eco.Presence[i]++
			}
		}
		for _, y := range ev.Years {
			if y.Year > lastYear {
				eco.Years = append(eco.Years, y)
				lastYear = y.Year
				eco.Extinctions += len(y.Extinct)
				eco.Arrivals += len(y.Arrived)
			}
		}
		eco.PeakPopulation = max(eco.PeakPopulation, info.Population)
		eco.MaxPlots = max(eco.MaxPlots, ev.Plots)
		known := s.TechsKnown()
		if eco.FarmingMinute == nil && known["pertanian"] {
			eco.FarmingMinute = &minute
		}
		if eco.HerdingMinute == nil && known["peternakan"] {
			eco.HerdingMinute = &minute
		}
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
			Animals: animals, Livestock: info.Livestock, Plots: ev.Plots, Forest: ev.Forest,
			FoodStock: ev.Food.Carried + ev.Food.Stored + ev.Food.Granary, ENSO: ev.ENSO,
		})
	}
	ev := s.Ecology()
	for i, sp := range ev.Species {
		eco.Animals[i] = sp.Wild + sp.Tame
		eco.Presence[i] /= float64(cfg.Minutes)
		eco.Livestock += sp.Tame
	}
	eco.Forest = ev.Forest
	eco.CapacityHits = s.Info().CapacityHits
	wall := time.Since(start)
	if era1 < 0 {
		era1 = cfg.Minutes
	}
	info := s.Info()
	health := s.Health()
	r.Final = final{
		Era1LastedMinutes: era1, FirstHouseMinute: firstHouse, Eras: info.Era, Population: info.Population,
		MaxGeneration: info.MaxGeneration, Tier: info.Tier, TierName: info.TierName, Elements: info.ElementsDiscovered,
		Houses: info.Houses, Structures: info.Structures, Crimes: info.Crimes, Kindness: info.Kindness, Kills: info.Kills,
		Demography:     s.Demography().Current,
		Engine:         info.Engine,
		Genetics:       s.Genetics(),
		Latrines:       health.Latrines,
		Wells:          health.Wells,
		Strains:        health.Strains,
		TierGeneration: tierGen,
		KnowledgeLost:  info.KnowledgeLost,
		AvgBrainSize:   info.AvgBrainSize,
		SkillByAge:     s.SkillProfile(),
		AvgSkill:       r.Series[len(r.Series)-1].AvgSkill,
		MsPerTick:      float64(wall.Microseconds()) / 1000 / float64(cfg.Minutes*60*sim.TicksPerSecond),
		WallSeconds:    wall.Seconds(),
		Ecology:        eco,
		Nutrition:      nut.final(s.Nutrition()),
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
		case "perception":
			opts.NoPerception = true
		case "navigation":
			opts.NoNavigation = true
		case "ai_budget":
			opts.NoAIBudget = true
		case "instincts":
			opts.NoInstincts = true
		case "learning":
			opts.NoLearning = true
		case "plasticity":
			opts.NoPlasticity = true
		case "culture":
			opts.NoCulture = true
		case "humans":
			opts.NoHumans = true
		case "farming":
			opts.NoFarming = true
		case "climate":
			opts.NoClimate = true
		case "fauna":
			opts.NoFauna = true
		case "space":
			opts.NoPersonalSpace = true
		case "neurogenesis":
			opts.NoNeurogenesis = true
		case "water":
			opts.NoWaterCycle = true
		case "disease":
			opts.NoDisease = true
		case "sanitation":
			opts.NoSanitation = true
		case "attachment":
			opts.NoAttachment = true
		case "stimuli":
			opts.NoStimuli = true
		case "affect":
			opts.NoAffect = true
		case "exchange":
			opts.NoExchange = true
		case "fire":
			opts.NoFire = true
		case "mobility":
			opts.NoMobility = true
		case "villages":
			opts.NoVillages = true
		case "nutrition":
			opts.NoNutrition = true
		case "sharing":
			opts.NoSharing = true
		case "genetics":
			opts.NoGenetics = true
		default:
			return opts, nil, fmt.Errorf("unknown rule %q (known: crime, instincts, learning, plasticity, culture, humans, farming, climate, fauna, space, neurogenesis, water, disease, sanitation, attachment, perception, navigation, ai_budget, stimuli, affect, exchange, fire, mobility, villages, nutrition, sharing, genetics)", name)
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

// write saves the report: the full data as gzipped JSON (a world's yearly
// ecology records make it large) and the readable summary.md.
func write(dir string, rep report) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(rep)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write(data); err != nil {
		return err
	}
	if err := zw.Close(); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "summary.json.gz"), buf.Bytes(), 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "summary.md"), []byte(markdown(rep)), 0o644)
}

// readReport loads a report's data: summary.json.gz, or the plain
// summary.json older reports have.
func readReport(dir string) (report, error) {
	var rep report
	data, err := os.ReadFile(filepath.Join(dir, "summary.json.gz"))
	if err == nil {
		zr, zerr := gzip.NewReader(bytes.NewReader(data))
		if zerr != nil {
			return rep, zerr
		}
		if data, err = io.ReadAll(zr); err != nil {
			return rep, err
		}
	} else if data, err = os.ReadFile(filepath.Join(dir, "summary.json")); err != nil {
		return rep, err
	}
	return rep, json.Unmarshal(data, &rep)
}

// rerender rebuilds a report's summary.md from its data, after the report
// layout changed.
func rerender(dir string) error {
	rep, err := readReport(dir)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "summary.md"), []byte(markdown(rep)), 0o644)
}
