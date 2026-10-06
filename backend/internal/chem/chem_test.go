package chem

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"testing"

	"miniv2/backend/internal/world"
)

func starterMap() *world.Map {
	return world.Generate(world.GenOptions{Name: "Starter Island", Width: 128, Height: 128, Seed: 1337})
}

func none(string) bool { return false }

func TestElements(t *testing.T) {
	els := Elements()
	if len(els) != 118 {
		t.Fatalf("got %d elements, want 118", len(els))
	}
	tierCounts := make([]int, MaxTier+1)
	cells := map[[2]int]string{}
	categories := map[string]bool{
		catAlkali: true, catAlkaline: true, catLanthanide: true, catActinide: true, catTransition: true,
		catPostTransition: true, catMetalloid: true, catNonmetal: true, catHalogen: true, catNoble: true,
	}
	for i, e := range els {
		if e.Z != i+1 {
			t.Fatalf("element %d has Z %d", i, e.Z)
		}
		if got, ok := ElementBySymbol(e.Symbol); !ok || got.Z != e.Z {
			t.Errorf("lookup of %s failed", e.Symbol)
		}
		if e.Name == "" || !categories[e.Category] || !slices.Contains([]string{solid, liquid, gas}, e.Phase) {
			t.Errorf("%s: bad name/category/phase %q %q %q", e.Symbol, e.Name, e.Category, e.Phase)
		}
		wantNatural := e.Z <= 92 && e.Z != 43 && e.Z != 61 && e.Z != 85
		if e.Natural != wantNatural || (e.Abundance > 0) != e.Natural {
			t.Errorf("%s: natural=%v abundance=%v", e.Symbol, e.Natural, e.Abundance)
		}
		if !e.Natural && e.Tier < 6 || e.Natural && e.Tier > 5 {
			t.Errorf("%s: tier %d does not match natural=%v", e.Symbol, e.Tier, e.Natural)
		}
		lanthanoid := e.Z >= 57 && e.Z <= 71 || e.Z >= 89 && e.Z <= 103
		if (e.Group == 0) != lanthanoid {
			t.Errorf("%s: group %d", e.Symbol, e.Group)
		}
		if e.Group > 0 {
			cell := [2]int{e.Period, e.Group}
			if other, dup := cells[cell]; dup {
				t.Errorf("%s and %s share period %d group %d", e.Symbol, other, e.Period, e.Group)
			}
			cells[cell] = e.Symbol
		}
		tierCounts[e.Tier]++
	}
	if want := []int{5, 8, 25, 21, 22, 8, 11, 18}; !slices.Equal(tierCounts, want) {
		t.Errorf("tier counts %v, want %v", tierCounts, want)
	}

	spots := map[string][3]int{ // period, group, Z
		"H": {1, 1, 1}, "He": {1, 18, 2}, "B": {2, 13, 5}, "Ne": {2, 18, 10}, "Na": {3, 1, 11},
		"Ar": {3, 18, 18}, "Fe": {4, 8, 26}, "Kr": {4, 18, 36}, "Ag": {5, 11, 47}, "Cs": {6, 1, 55},
		"La": {6, 0, 57}, "Hf": {6, 4, 72}, "Rn": {6, 18, 86}, "Fr": {7, 1, 87}, "Ac": {7, 0, 89},
		"Rf": {7, 4, 104}, "Og": {7, 18, 118},
	}
	for sym, want := range spots {
		e, _ := ElementBySymbol(sym)
		if got := [3]int{e.Period, e.Group, e.Z}; got != want {
			t.Errorf("%s at %v, want %v", sym, got, want)
		}
	}
	for sym, name := range map[string]string{"Fe": "Besi", "Au": "Emas", "Hg": "Raksa", "Og": "Oganeson"} {
		if e, _ := ElementBySymbol(sym); e.Name != name {
			t.Errorf("%s named %q, want %q", sym, e.Name, name)
		}
	}
	for tier := range MaxTier + 1 {
		if TierName(tier) == "" {
			t.Errorf("tier %d has no name", tier)
		}
	}
	if TierName(-1) != "" || TierName(MaxTier+1) != "" {
		t.Error("out-of-range tiers should have no name")
	}
}

func TestDataIntegrity(t *testing.T) {
	uniq := func(kind string, ids []string) {
		seen := map[string]bool{}
		for _, id := range ids {
			if id == "" || seen[id] {
				t.Errorf("%s id %q is empty or duplicated", kind, id)
			}
			seen[id] = true
		}
	}
	var itemIDs, techIDs, structIDs, recipeIDs []string
	for _, it := range Items() {
		itemIDs = append(itemIDs, string(it.ID))
		if it.Name == "" || it.Value < 0 || len(it.Elements) == 0 {
			t.Errorf("item %s: incomplete", it.ID)
		}
		for _, sym := range it.Elements {
			e, ok := ElementBySymbol(sym)
			if !ok || !e.Natural {
				t.Errorf("item %s contains %q, which is unknown or synthetic", it.ID, sym)
			}
		}
		if (it.Kind == kindTool) != (it.Gather > 0) || (it.Kind == kindWeapon) != (it.Damage > 0) {
			t.Errorf("item %s: kind %q with gather %v damage %v", it.ID, it.Kind, it.Gather, it.Damage)
		}
	}
	for _, tc := range Techs() {
		techIDs = append(techIDs, tc.ID)
	}
	for _, s := range Structures() {
		structIDs = append(structIDs, s.ID)
	}
	for _, r := range Recipes() {
		recipeIDs = append(recipeIDs, r.ID)
	}
	uniq("item", itemIDs)
	uniq("tech", techIDs)
	uniq("structure", structIDs)
	uniq("recipe", recipeIDs)

	if f, ok := ItemByID(Food); !ok || f.Food <= 0 {
		t.Error("food item missing or inedible")
	}
	if _, ok := ItemByID(SynthesisFuel); !ok {
		t.Error("synthesis fuel item missing")
	}

	tech := func(where, id string) {
		if id != "" && !slices.Contains(techIDs, id) {
			t.Errorf("%s references unknown tech %q", where, id)
		}
	}
	item := func(where string, id ItemID) {
		if _, ok := ItemByID(id); !ok {
			t.Errorf("%s references unknown item %q", where, id)
		}
	}

	stationTier := map[string]int{}
	stationTeaches := map[int]string{}
	for _, s := range Structures() {
		tech(s.ID, s.Tech)
		tech(s.ID, s.Teaches)
		for id, n := range s.Cost {
			item(s.ID, id)
			if n <= 0 {
				t.Errorf("%s costs %d %s", s.ID, n, id)
			}
		}
		if s.House != (s.Level > 0) || (s.House || s.Granary) != (s.Storage > 0) {
			t.Errorf("%s: house=%v level=%d storage=%d", s.ID, s.House, s.Level, s.Storage)
		}
		if s.Upgrades != "" {
			up, ok := StructureByID(s.Upgrades)
			if !ok || !up.House || up.Level != s.Level-1 {
				t.Errorf("%s upgrades %q which is not the house one level below", s.ID, s.Upgrades)
			}
		}
		if s.Tier > 0 {
			if _, dup := stationTeaches[s.Tier]; dup {
				t.Errorf("two stations for tier %d", s.Tier)
			}
			stationTier[s.ID] = s.Tier
			stationTeaches[s.Tier] = s.Teaches
		}
	}
	wantTeaches := map[int]string{1: "peleburan", 2: "kimia", 3: "listrik", 4: "spektroskopi", 5: "radiokimia", 6: "fisika_nuklir", 7: "fisika_partikel"}
	for tier, want := range wantTeaches {
		if stationTeaches[tier] != want {
			t.Errorf("tier %d station teaches %q, want %q", tier, stationTeaches[tier], want)
		}
	}

	for _, r := range Recipes() {
		tech(r.ID, r.Tech)
		tech(r.ID, r.Teaches)
		if len(r.Inputs) == 0 || len(r.Outputs) == 0 || r.Seconds <= 0 {
			t.Errorf("recipe %s is incomplete", r.ID)
		}
		for id := range r.Inputs {
			item(r.ID, id)
		}
		for id := range r.Outputs {
			item(r.ID, id)
		}
		tier := 0
		if r.Station != "" {
			var ok bool
			if tier, ok = stationTier[r.Station]; !ok {
				t.Errorf("recipe %s needs %q, which is not a station", r.ID, r.Station)
			}
		}
		for _, sym := range r.Discovers {
			if e, ok := ElementBySymbol(sym); !ok || e.Tier > tier {
				t.Errorf("recipe %s at tier %d discovers %s (tier %d)", r.ID, tier, sym, e.Tier)
			}
		}
	}

	// The tech graph must be acyclic.
	reqs := map[string][]string{}
	for _, tc := range Techs() {
		reqs[tc.ID] = tc.Requires
		for _, r := range tc.Requires {
			tech(tc.ID, r)
		}
	}
	state := map[string]int{} // 1 visiting, 2 done
	var visit func(id string)
	visit = func(id string) {
		switch state[id] {
		case 1:
			t.Fatalf("tech cycle through %s", id)
		case 2:
			return
		}
		state[id] = 1
		for _, r := range reqs[id] {
			visit(r)
		}
		state[id] = 2
	}
	for id := range reqs {
		visit(id)
	}
}

// progress plays the tech tree forward from the stone age with whatever the
// geology offers, the way a civilisation could.
type progress struct {
	have    map[ItemID]bool
	known   map[string]bool
	built   map[string]bool
	crafted map[string]bool
	order   []string // techs in the order learned
}

func playForward(g *Geology) progress {
	p := progress{have: map[ItemID]bool{"udara": true}, known: map[string]bool{}, built: map[string]bool{}, crafted: map[string]bool{}}
	needTool := map[ItemID]bool{}
	for i, s := range g.slots {
		if s.item == 0 || g.amounts[i] <= 0 {
			continue
		}
		id := items[s.item-1].ID
		if s.tool {
			needTool[id] = true
		} else {
			p.have[id] = true
		}
	}
	learn := func(id string) {
		if id != "" && !p.known[id] {
			p.known[id] = true
			p.order = append(p.order, id)
		}
	}
	techOK := func(id string) bool { return id == "" || p.known[id] }
	// Farming and herding come from living off the land, not from recipes:
	// wild crops and animals are there from the start.
	learn(HerdingTech)
	for _, c := range crops {
		if c.WildChance > 0 {
			learn(FarmingTech)
			p.have[c.Item] = true
		}
	}
	p.have["daging"], p.have["ikan"] = true, true
	all := func(m map[ItemID]int) bool {
		for id := range m {
			if !p.have[id] {
				return false
			}
		}
		return true
	}
	for changed := true; changed; {
		changed = false
		for id := range p.have {
			if it, _ := ItemByID(id); it.Gather > 0 {
				for nt := range needTool {
					if !p.have[nt] {
						p.have[nt] = true
						changed = true
					}
				}
				break
			}
		}
		for _, r := range recipes {
			if p.crafted[r.ID] || !techOK(r.Tech) || (r.Station != "" && !p.built[r.Station]) || !all(r.Inputs) {
				continue
			}
			p.crafted[r.ID] = true
			for id := range r.Outputs {
				p.have[id] = true
			}
			learn(r.Teaches)
			changed = true
		}
		for _, s := range structures {
			if p.built[s.ID] || !techOK(s.Tech) || (s.Upgrades != "" && !p.built[s.Upgrades]) || !all(s.Cost) {
				continue
			}
			p.built[s.ID] = true
			learn(s.Teaches)
			changed = true
		}
	}
	return p
}

func TestEverythingIsReachableFromTheStoneAge(t *testing.T) {
	p := playForward(NewGeology(starterMap()))
	for _, s := range structures {
		if !p.built[s.ID] {
			t.Errorf("structure %s can never be built", s.ID)
		}
	}
	for _, r := range recipes {
		if !p.crafted[r.ID] {
			t.Errorf("recipe %s can never be made", r.ID)
		}
	}
	learnedAt := map[string]int{}
	for i, id := range p.order {
		learnedAt[id] = i
	}
	for _, tc := range techs {
		at, ok := learnedAt[tc.ID]
		if !ok {
			t.Errorf("tech %s is never learned", tc.ID)
			continue
		}
		for _, r := range tc.Requires {
			if ra, ok := learnedAt[r]; !ok || ra > at {
				t.Errorf("tech %s is learned before its requirement %s", tc.ID, r)
			}
		}
	}
}

func TestEveryElementIsDiscoverable(t *testing.T) {
	p := playForward(NewGeology(starterMap()))
	for _, e := range Elements() {
		if !e.Natural {
			if !slices.Contains(Synthesizable(e.Tier, none), e.Symbol) {
				t.Errorf("synthetic %s cannot be synthesized at tier %d", e.Symbol, e.Tier)
			}
			continue
		}
		found := false
		for id := range p.have {
			if slices.Contains(Discoverable(id, e.Tier, none), e.Symbol) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("%s (tier %d) is in no obtainable item", e.Symbol, e.Tier)
		}
	}
}

func TestDiscoverableAndSynthesizable(t *testing.T) {
	if got := Discoverable("galena", 0, none); !slices.Equal(got, []string{"S", "Ag"}) {
		t.Errorf("galena at tier 0: %v", got)
	}
	if got := Discoverable("galena", 4, none); !slices.Equal(got, []string{"S", "Ag", "Tl", "Pb"}) {
		t.Errorf("galena at tier 4: %v", got)
	}
	knownS := func(s string) bool { return s == "S" }
	if got := Discoverable("galena", 0, knownS); !slices.Equal(got, []string{"Ag"}) {
		t.Errorf("galena with S known: %v", got)
	}
	if got := Discoverable("nope", 7, none); got != nil {
		t.Errorf("unknown item: %v", got)
	}
	if got := Synthesizable(5, none); len(got) != 0 {
		t.Errorf("synthesis before the reactor: %v", got)
	}
	if got := Synthesizable(6, none); len(got) != 11 || got[0] != "Tc" || got[10] != "Fm" {
		t.Errorf("reactor synthesis: %v", got)
	}
	if got := Synthesizable(7, func(s string) bool { return s != "Og" }); !slices.Equal(got, []string{"Og"}) {
		t.Errorf("last synthesis: %v", got)
	}
}

func mineralCounts(g *Geology, m *world.Map) map[ItemID]int {
	l := lay{m: m, seed: uint64(m.Seed)}
	out := map[ItemID]int{}
	for i, s := range g.slots {
		x, y := (i/slotsPerTile)%g.w, (i/slotsPerTile)/g.w
		if s.item != 0 && l.reachable(x, y) {
			out[items[s.item-1].ID]++
		}
	}
	return out
}

// generatedMaps returns generated islands of the given sizes for several seeds.
func generatedMaps(sizes ...int) []*world.Map {
	var out []*world.Map
	for _, size := range sizes {
		for _, seed := range []uint32{1, 7, 42, 1337} {
			out = append(out, world.Generate(world.GenOptions{Name: "m", Width: size, Height: size, Seed: seed}))
		}
	}
	return out
}

func TestEveryMineralOnGeneratedMaps(t *testing.T) {
	for _, m := range generatedMaps(48, 64, 128, 256) {
		counts := mineralCounts(NewGeology(m), m)
		for _, it := range items {
			if it.Kind == kindMineral && counts[it.ID] == 0 {
				t.Errorf("%d seed %d: no reachable %s", m.Width, m.Seed, it.ID)
			}
		}
	}
}

func TestSmallMapsAreBestEffort(t *testing.T) {
	for _, size := range []int{world.MinSize, 24, 32} {
		m := world.Generate(world.GenOptions{Name: "small", Width: size, Height: size, Seed: 3})
		counts := mineralCounts(NewGeology(m), m)
		var missing []ItemID
		for _, it := range items {
			if it.Kind == kindMineral && counts[it.ID] == 0 {
				missing = append(missing, it.ID)
			}
		}
		t.Logf("%d×%d: %d minerals missing %v", size, size, len(missing), missing)
	}
}

func TestEveryMineralMustBeDug(t *testing.T) {
	g := NewGeology(starterMap())
	for _, d := range g.Deposits() {
		it, _ := ItemByID(d.Item)
		if it.Kind == kindMineral && !d.NeedsTool {
			t.Fatalf("%s at (%d,%d) can be taken without digging", d.Item, d.X, d.Y)
		}
	}
}

// Where each deposit model is allowed to sit, by host rock unit.
var modelRocks = map[string][]world.Rock{
	"porfiri":         {world.RockGranite, world.RockVolcanic},
	"epitermal":       {world.RockVolcanic},
	"skarn":           {world.RockLimestone, world.RockGranite},
	"granit_timah":    {world.RockGranite},
	"pegmatit":        {world.RockPegmatite},
	"ofiolit":         {world.RockUltramafic},
	"laterit_nikel":   {world.RockUltramafic},
	"laterit_besi":    {world.RockVolcanic, world.RockMetamorphic, world.RockSedimentary},
	"bauksit":         {world.RockGranite, world.RockVolcanic},
	"tembaga_alam":    {world.RockVolcanic},
	"batubara":        {world.RockSedimentary},
	"mvt":             {world.RockLimestone},
	"karbonat":        {world.RockLimestone},
	"evaporit":        {world.RockEvaporite},
	"uranium_sedimen": {world.RockSedimentary},
	"karbonatit":      {world.RockCarbonatite},
	"mangan":          {world.RockMetamorphic},
}

// Which models each telltale mineral may come from.
var mineralModels = map[ItemID][]string{
	"kromit": {"ofiolit"}, "pentlandit": {"ofiolit"}, "kobaltit": {"ofiolit"}, "laterit_nikel": {"laterit_nikel"},
	"kasiterit": {"granit_timah", "pegmatit", "plaser"}, "wolframit": {"granit_timah"}, "scheelit": {"skarn"},
	"bijih_emas": {"porfiri", "epitermal", "plaser"}, "malakit": {"porfiri"}, "kalkopirit": {"porfiri", "skarn"},
	"tembaga_alam": {"tembaga_alam"}, "sinabar": {"epitermal"}, "stibnit": {"epitermal"},
	"belerang": {"solfatara"}, "batu_bara": {"batubara"},
	"halit": {"evaporit"}, "boraks": {"evaporit"}, "silvit": {"evaporit"}, "gipsum": {"evaporit"}, "selestin": {"evaporit"},
	"bastnasit": {"karbonatit"}, "pirokhlor": {"karbonatit"},
	"spodumen": {"pegmatit"}, "polusit": {"pegmatit"}, "beril": {"pegmatit"}, "torvetit": {"pegmatit"},
	"karnotit": {"uranium_sedimen"}, "uraninit": {"pegmatit", "uranium_sedimen"},
	"pirolusit": {"mangan"}, "bauksit": {"bauksit"}, "pasir_besi": {"plaser"},
	"hematit": {"skarn"}, "limonit": {"laterit_besi", "laterit_nikel"},
}

func TestDepositsSitInTheirHostRock(t *testing.T) {
	for _, m := range generatedMaps(64, 128) {
		g := NewGeology(m)
		geo := g.GeoModel()
		nearCrater := func(x, y int) bool {
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					if nx, ny := x+dx, y+dy; nx >= 0 && ny >= 0 && nx < m.Width && ny < m.Height && m.Layers.Ground[ny*m.Width+nx] == world.Crater {
						return true
					}
				}
			}
			return false
		}
		for _, d := range g.Deposits() {
			if allowed, ok := mineralModels[d.Item]; ok && !slices.Contains(allowed, d.Model) {
				t.Errorf("%d seed %d: %s from model %s at (%d,%d)", m.Width, m.Seed, d.Item, d.Model, d.X, d.Y)
			}
			if rocks, ok := modelRocks[d.Model]; ok && !d.Surface {
				if r := geo.Rock[d.Y*m.Width+d.X]; !slices.Contains(rocks, r) {
					t.Errorf("%d seed %d: %s (%s) sits on %s", m.Width, m.Seed, d.Item, d.Model, world.RockTypes[r].Key)
				}
			}
			if d.Model == "solfatara" && !nearCrater(d.X, d.Y) {
				t.Errorf("%d seed %d: sulphur at (%d,%d) is not on a crater rim", m.Width, m.Seed, d.X, d.Y)
			}
		}
	}
}

// Placers must lie downstream of the rocks that shed them.
func TestPlacersComeFromUpstream(t *testing.T) {
	for _, m := range generatedMaps(64, 128, 256) {
		g := NewGeology(m)
		geo := g.GeoModel()
		w := m.Width
		type source struct{ gold, granite, volcanic int } // first course index draining each
		sources := make([]source, len(geo.Rivers))
		for r, path := range geo.Rivers {
			s := source{-1, -1, -1}
			for k, p := range path {
				for dy := -3; dy <= 3; dy++ {
					for dx := -3; dx <= 3; dx++ {
						x, y := p.X+dx, p.Y+dy
						if x < 0 || y < 0 || x >= w || y >= m.Height {
							continue
						}
						j := y*w + x
						if md := g.modelAt(j * slotsPerTile); (md == "porfiri" || md == "epitermal") && s.gold < 0 {
							s.gold = k
						}
						if rk := geo.Rock[j]; (rk == world.RockGranite || rk == world.RockPegmatite) && s.granite < 0 {
							s.granite = k
						}
						if geo.Rock[j] == world.RockVolcanic && s.volcanic < 0 {
							s.volcanic = k
						}
					}
				}
			}
			sources[r] = s
		}
		downstreamOf := func(x, y int, from func(source) int) bool {
			for r, path := range geo.Rivers {
				k0 := from(sources[r])
				if k0 < 0 {
					continue
				}
				for _, p := range path[k0:] {
					if abs(p.X-x)+abs(p.Y-y) <= 1 {
						return true
					}
				}
				// Delta and nearby beaches.
				mouth := path[0]
				for _, q := range path {
					if geo.RiverOf[q.Y*w+q.X] == int16(r) {
						mouth = q
					}
				}
				if math.Hypot(float64(x-mouth.X), float64(y-mouth.Y)) <= 2.5*(3+geo.Size()/64)+0.5 {
					return true
				}
			}
			return false
		}
		for _, d := range g.Deposits() {
			if d.Model != "plaser" {
				continue
			}
			var from func(source) int
			switch d.Item {
			case "bijih_emas":
				from = func(s source) int { return s.gold }
			case "kasiterit", "zirkon", "monasit", "xenotim":
				from = func(s source) int { return s.granite }
			case "pasir_besi":
				from = func(s source) int { return s.volcanic }
			default:
				continue // ilmenite comes from either granite or volcanic catchments
			}
			if !downstreamOf(d.X, d.Y, from) {
				t.Errorf("%d seed %d: %s placer at (%d,%d) has no source upstream", m.Width, m.Seed, d.Item, d.X, d.Y)
			}
		}
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// TestAbundance reports per-ore deposit counts on the starter map and checks
// they follow real-world commonness: iron ≫ copper > tin > gold.
func TestAbundance(t *testing.T) {
	m := starterMap()
	g := NewGeology(m)
	counts := mineralCounts(g, m)
	sum := func(ids ...ItemID) int {
		n := 0
		for _, id := range ids {
			n += counts[id]
		}
		return n
	}
	iron := sum("hematit", "limonit", "pasir_besi")
	copper := sum("kalkopirit", "malakit", "tembaga_alam")
	tin := sum("kasiterit")
	gold := sum("bijih_emas")
	coal := sum("batu_bara")
	t.Logf("reachable deposits on the 128×128 seed-1337 map: iron %d (hematit %d, limonit %d, pasir_besi %d), copper %d (kalkopirit %d, malakit %d, tembaga_alam %d), tin %d, gold %d, coal %d, lead %d, zinc %d",
		iron, counts["hematit"], counts["limonit"], counts["pasir_besi"],
		copper, counts["kalkopirit"], counts["malakit"], counts["tembaga_alam"], tin, gold, coal, counts["galena"], counts["sfalerit"])
	if !(iron > 2*copper && copper > tin && tin > gold && gold > 0) {
		t.Errorf("abundance out of order: iron %d, copper %d, tin %d, gold %d", iron, copper, tin, gold)
	}
	if iron < 60 || copper < 15 || tin < 8 || coal < 20 {
		t.Errorf("too little of a key ore: iron %d, copper %d, tin %d, coal %d", iron, copper, tin, coal)
	}
	if testing.Verbose() {
		var all []string
		for id, n := range counts {
			all = append(all, fmt.Sprintf("%s=%d", id, n))
		}
		slices.Sort(all)
		t.Logf("all reachable deposits: %v", all)
	}
}

func TestGeologyIsDeterministic(t *testing.T) {
	m := starterMap()
	a, b := NewGeology(m), NewGeology(m)
	if !slices.Equal(a.Amounts(), b.Amounts()) || !slices.Equal(a.slots, b.slots) {
		t.Fatal("same map produced different geology")
	}
	other := world.Generate(world.GenOptions{Name: "x", Width: 128, Height: 128, Seed: 7})
	if slices.Equal(NewGeology(other).slots, a.slots) {
		t.Fatal("different maps produced identical geology")
	}
}

// findSource returns a walkable tile from which item can be gathered.
func findSource(t *testing.T, m *world.Map, g *Geology, item ItemID) (int, int, Source) {
	t.Helper()
	l := lay{m: m, seed: uint64(m.Seed)}
	for y := range m.Height {
		for x := range m.Width {
			if !l.walkable(x, y) {
				continue
			}
			for _, s := range g.Sources(x, y) {
				if s.Item == item {
					return x, y, s
				}
			}
		}
	}
	t.Fatalf("no %s anywhere", item)
	return 0, 0, Source{}
}

func TestTakeRegrowAndPersist(t *testing.T) {
	m := starterMap()
	g := NewGeology(m)

	x, y, wood := findSource(t, m, g, "kayu")
	if wood.NeedsTool {
		t.Error("wood should not need a tool")
	}
	if got := g.Take(wood.X, wood.Y, "kayu", 3); got != 3 {
		t.Fatalf("took %v wood, want 3", got)
	}
	if got := g.Take(wood.X, wood.Y, "kayu", 100); got != wood.Amount-3 {
		t.Fatalf("took %v of the rest, want %v", got, wood.Amount-3)
	}
	for _, s := range g.Sources(x, y) {
		if s.Item == "kayu" && s.X == wood.X && s.Y == wood.Y {
			t.Fatal("exhausted wood still listed")
		}
	}
	g.Regrow(10)
	if got := g.Take(wood.X, wood.Y, "kayu", 100); got <= 0 || got > 0.21 {
		t.Fatalf("after 10 s of regrowth took %v wood", got)
	}

	_, _, water := findSource(t, m, g, "air")
	if got := g.Take(water.X, water.Y, "air", 5); got != 5 {
		t.Fatalf("water should be endless, took %v", got)
	}
	if got := g.Take(0, 0, "udara", 2); got != 2 {
		t.Fatalf("air should be endless, took %v", got)
	}

	_, _, ore := findSource(t, m, g, "hematit")
	if !ore.NeedsTool {
		t.Error("hematite should need a tool")
	}
	g.Take(ore.X, ore.Y, "hematit", 5)

	saved := g.Amounts()
	restored := NewGeology(m)
	if err := restored.SetAmounts(saved); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(restored.Amounts(), saved) {
		t.Fatal("amounts differ after restore")
	}
	if err := restored.SetAmounts(saved[1:]); err == nil {
		t.Fatal("wrong-length amounts accepted")
	}

	// Editing the map keeps what was already mined.
	g.Update(m)
	if !slices.Equal(g.Amounts(), saved) {
		t.Fatal("update lost remaining amounts")
	}
}

func TestGeologyOnTinyMap(t *testing.T) {
	m := world.Generate(world.GenOptions{Name: "tiny", Width: world.MinSize, Height: world.MinSize, Seed: 3})
	g := NewGeology(m)
	if len(g.Amounts()) != world.MinSize*world.MinSize*slotsPerTile {
		t.Fatal("wrong deposit count")
	}
	g.Regrow(1)
	g.Sources(0, 0)
	g.Sources(world.MinSize-1, world.MinSize-1)
}

func BenchmarkNewGeology(b *testing.B) {
	m := world.Generate(world.GenOptions{Name: "big", Width: world.MaxSize, Height: world.MaxSize, Seed: 9})
	for range b.N {
		NewGeology(m)
	}
}

func BenchmarkRegrow(b *testing.B) {
	g := NewGeology(starterMap())
	b.ResetTimer()
	for range b.N {
		g.Regrow(0.5)
	}
}

func BenchmarkRegrowAfterHarvest(b *testing.B) {
	m := starterMap()
	g := NewGeology(m)
	n := 0
	for i, s := range g.slots {
		if s.item != 0 && g.regrow[i] > 0 && n < 500 {
			x, y := (i/slotsPerTile)%g.w, (i/slotsPerTile)/g.w
			g.Take(x, y, items[s.item-1].ID, 1)
			n++
		}
	}
	b.ResetTimer()
	for range b.N {
		g.Regrow(0.001)
	}
}

// everythingMissing lists what a civilisation on map m could never build,
// make or discover.
func everythingMissing(m *world.Map) []string {
	p := playForward(NewGeology(m))
	var missing []string
	for _, s := range structures {
		if !p.built[s.ID] {
			missing = append(missing, "structure "+s.ID)
		}
	}
	for _, r := range recipes {
		if !p.crafted[r.ID] {
			missing = append(missing, "recipe "+r.ID)
		}
	}
	for _, e := range Elements() {
		if !e.Natural {
			continue
		}
		found := false
		for id := range p.have {
			if slices.Contains(Discoverable(id, e.Tier, none), e.Symbol) {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, "element "+e.Symbol)
		}
	}
	return missing
}

func TestEverythingBuildableOnGeneratedMaps(t *testing.T) {
	for _, m := range generatedMaps(48, 64, 96, 128, 192, 256) {
		if missing := everythingMissing(m); len(missing) > 0 {
			t.Errorf("%d seed %d: %d missing, e.g. %v", m.Width, m.Seed, len(missing), missing[:min(6, len(missing))])
		}
	}
}

// Hand edits change the surface, not the bedrock: flattening mountains keeps
// every deposit (relocated into the same host rock where needed).
func TestEditsKeepTheBedrock(t *testing.T) {
	noMountains := starterMap()
	for i, g := range noMountains.Layers.Ground {
		if g == world.Mountain || g == world.StoneFloor {
			noMountains.Layers.Ground[i] = world.Grass
		}
	}
	if missing := everythingMissing(noMountains); len(missing) > 0 {
		t.Errorf("starter without mountains: %v", missing)
	}

	// Removing every tree really removes the wood; that is the editor's choice, not faked.
	noTrees := starterMap()
	clear(noTrees.Layers.Objects)
	missing := everythingMissing(noTrees)
	t.Logf("starter without trees or boulders: %d missing %v", len(missing), missing[:min(8, len(missing))])

	meadow := world.Generate(world.GenOptions{Name: "meadow", Width: 48, Height: 48, Seed: 3})
	for i := range meadow.Layers.Ground {
		meadow.Layers.Ground[i] = world.Grass
	}
	t.Logf("all-grass 48×48: missing %v", everythingMissing(meadow))
}

func TestDepositsListsWhatSourcesOffer(t *testing.T) {
	m := starterMap()
	g := NewGeology(m)
	ds := g.Deposits()
	if len(ds) == 0 {
		t.Fatal("no deposits")
	}
	seen := map[ItemID]bool{}
	for _, d := range ds {
		seen[d.Item] = true
		found := false
		for _, s := range g.Sources(d.X, d.Y) {
			if s.Item == d.Item && s.X == d.X && s.Y == d.Y {
				found = true
			}
		}
		if !found {
			t.Fatalf("deposit %+v not offered by Sources on its own tile", d)
		}
	}
	for _, it := range items {
		if it.Kind == kindMineral && !seen[it.ID] {
			t.Errorf("mineral %s missing from Deposits", it.ID)
		}
	}
}

// TestDepositDump prints the starter map's deposits by model, for eyeballing.
func TestDepositDump(t *testing.T) {
	if !testing.Verbose() {
		t.Skip("run with -v to print the deposit map")
	}
	m := starterMap()
	g := NewGeology(m)
	letter := map[string]byte{
		"porfiri": 'P', "epitermal": 'E', "solfatara": 'S', "skarn": 'K', "granit_timah": 'T', "pegmatit": 'G',
		"ofiolit": 'O', "laterit_nikel": 'N', "laterit_besi": 'F', "bauksit": 'B', "tembaga_alam": 'C',
		"batubara": 'c', "karbonat": 'l', "mvt": 'M', "evaporit": 'V', "uranium_sedimen": 'U',
		"karbonatit": 'R', "mangan": 'n', "plaser": '*', "lempung": ',', "pasir": '.', "batu": ':',
		"air": '=', "laut": '~',
	}
	var sb strings.Builder
	for y := 0; y < m.Height; y += 2 {
		for x := range m.Width {
			c := byte(' ')
			if l, ok := letter[g.modelAt((y*m.Width+x)*slotsPerTile)]; ok {
				c = l
			}
			sb.WriteByte(c)
		}
		sb.WriteByte('\n')
	}
	t.Log("\nP porfiri  E epitermal  S belerang kawah  K skarn  T granit timah  G pegmatit  O ofiolit  N laterit nikel  F laterit besi  B bauksit  C tembaga alam  c batu bara  l gamping  M Pb-Zn  V evaporit  U uranium  R karbonatit  n mangan  * plaser  , lempung  . pasir  : batu  = air tawar  ~ laut\n" + sb.String())
}

func TestMinedOutListsDugOutMinerals(t *testing.T) {
	m := starterMap()
	g := NewGeology(m)
	if len(g.MinedOut()) != 0 {
		t.Fatal("a fresh map has mined-out tiles")
	}
	var ore Deposit
	for _, d := range g.Deposits() {
		if !d.Surface && d.NeedsTool {
			if it, _ := ItemByID(d.Item); it.Kind == kindMineral {
				ore = d
				break
			}
		}
	}
	if ore.Item == "" {
		t.Fatal("no ore on the starter map")
	}
	g.Take(ore.X, ore.Y, ore.Item, ore.Amount)
	out := g.MinedOut()
	if len(out) != 1 || out[0] != [2]int{ore.X, ore.Y} {
		t.Fatalf("mined out %v, want [%d %d]", out, ore.X, ore.Y)
	}
	g.Regrow(1000)
	if len(g.MinedOut()) != 1 {
		t.Fatal("a mine should not grow back")
	}
}

func TestCropsAreComplete(t *testing.T) {
	grounds := map[string]bool{}
	for _, g := range world.GroundTiles {
		if !g.Solid {
			grounds[g.Key] = true
		}
	}
	seen := map[ItemID]bool{}
	for _, c := range Crops() {
		if seen[c.Item] {
			t.Errorf("crop %s listed twice", c.Item)
		}
		seen[c.Item] = true
		it, ok := ItemByID(c.Item)
		if !ok || it.Food <= 0 || it.Keeps <= 0 {
			t.Errorf("crop %s: item missing, inedible or never spoiling", c.Item)
		}
		if c.Name == "" || c.GrowYears <= 0 || c.Yield <= 0 || c.Water <= 0 || c.Water > 1 || c.Drain < 0 || c.WildChance <= 0 {
			t.Errorf("crop %s is incomplete: %+v", c.Item, c)
		}
		if (c.RepeatYears > 0) != (c.LifeYears > 0) {
			t.Errorf("crop %s: a perennial needs both RepeatYears and LifeYears", c.Item)
		}
		for _, list := range [][]string{c.Ground, c.Wild} {
			if len(list) == 0 {
				t.Errorf("crop %s has no ground", c.Item)
			}
			for _, g := range list {
				if !grounds[g] {
					t.Errorf("crop %s: %q is not a walkable ground tile", c.Item, g)
				}
			}
		}
	}
	for _, id := range []ItemID{"daging", "ikan", "daging_asap", "ikan_asin"} {
		if it, ok := ItemByID(id); !ok || it.Food <= 0 || it.Keeps <= 0 {
			t.Errorf("food %s missing, inedible or never spoiling", id)
		}
	}
}
