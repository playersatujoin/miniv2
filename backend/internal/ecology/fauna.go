package ecology

import (
	"fmt"
	"math"
	"slices"
)

// Diet says what an animal eats.
type Diet uint8

const (
	Grazer   Diet = iota // grass and leaves
	Forager              // fruit, roots, seeds and small things: the same wild food people gather
	Predator             // other animals
)

// Species is a kind of wild animal. Life history follows the real animals in
// years; feeding and numbers are compressed with the rest of the world.
type Species struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Diet Diet   `json:"diet"`
	// Size is body size relative to a person (drawing, how hard it is to kill).
	Size  float64 `json:"size"`
	Run   float64 `json:"run"`   // tiles per second when fleeing or charging
	Walk  float64 `json:"walk"`  // tiles per second when browsing
	Sight float64 `json:"sight"` // tiles
	Flee  float64 `json:"flee"`  // keeps this far from people; 0 = unafraid
	HP    float64 `json:"hp"`    // blows from a bare-handed person to bring it down
	// Bite is the health a person loses when it fights back (Fierce) or attacks.
	Bite   float64 `json:"bite"`
	Fierce bool    `json:"fierce"`
	Meat   int     `json:"meat"` // units of meat from a kill
	Adult  float64 `json:"adult"`
	Life   float64 `json:"life"`
	// Gestation (years), Litter (most young per birth), Interval (fewest
	// years between births).
	Gestation float64 `json:"gestation"`
	Litter    int     `json:"litter"`
	Interval  float64 `json:"interval"`
	Burn      float64 `json:"burn"` // energy per second
	Eat       float64 `json:"eat"`  // food units per second while eating
	Gain      float64 `json:"gain"` // energy per food unit
	// Breeding stops when more than Crowd of its kind live within Territory
	// tiles: territoriality and density dependence keep numbers in check.
	Territory float64 `json:"territory"`
	Crowd     int     `json:"crowd"`
	// Start is how many live per 1,000 walkable tiles when the world begins.
	Start   float64            `json:"start"`
	Habitat [numCovers]float64 `json:"-"`
	// Tame is the domestic animal it becomes ("" = can't be tamed).
	Tame   string `json:"tame,omitempty"`
	Raider bool   `json:"raider,omitempty"` // eats and tramples fields
	Prey   bool   `json:"prey,omitempty"`   // tigers hunt it
	// Kill words a person's death by it, e.g. "diterkam harimau".
	Kill string `json:"kill,omitempty"`
	// Catch is a tiger's chance of bringing it down in one pounce.
	Catch float64 `json:"catch,omitempty"`
}

// The animals of a western Indonesian island before people changed it:
// Javan rusa deer, wild boar, red junglefowl (the wild ancestor of the
// chicken, domesticated in Southeast Asia), wild water buffalo and the
// Javan/Sundaic tiger. Life histories are rounded textbook values (Corbet &
// Hill 1992 for the mammals; Keuling et al. 2013 for wild boar; Collias &
// Collias 1967 for junglefowl; Sunquist 1981 and Seidensticker 1986 for
// tigers); see docs/reference-ecology.md.
var species = []Species{
	{ID: "rusa", Name: "Rusa", Diet: Grazer, Size: 0.9, Run: 3.2, Walk: 0.6, Sight: 7, Flee: 3, HP: 1.5, Bite: 0.03,
		Meat: 6, Adult: 1.5, Life: 14, Gestation: 0.65, Litter: 1, Interval: 1, Burn: 0.03, Eat: 0.5, Gain: 0.25,
		Territory: 12, Crowd: 8, Start: 8, Habitat: hab(1, 0.7, 0.8, 1, 0.2, 0.3), Raider: true, Prey: true,
		Kill: "terluka ditanduk rusa", Catch: 0.5},
	{ID: "babi_hutan", Name: "Babi Hutan", Diet: Forager, Size: 0.8, Run: 2.6, Walk: 0.5, Sight: 6, Flee: 2.5, HP: 2, Bite: 0.12,
		Fierce: true, Meat: 5, Adult: 1, Life: 10, Gestation: 0.33, Litter: 4, Interval: 1, Burn: 0.03, Eat: 0.4, Gain: 0.3,
		Territory: 10, Crowd: 5, Start: 6, Habitat: hab(0.4, 1, 0.8, 0.5, 0.2, 0.2), Tame: "Babi", Raider: true, Prey: true,
		Kill: "diseruduk babi hutan", Catch: 0.45},
	// Junglefowl peck seeds and insects among grass and scrub rather than
	// the fruit and tubers pigs and people forage, so they don't starve the pigs.
	{ID: "ayam_hutan", Name: "Ayam Hutan", Diet: Grazer, Size: 0.35, Run: 2.2, Walk: 0.4, Sight: 5, Flee: 2, HP: 0.3,
		Meat: 1, Adult: 0.4, Life: 5, Gestation: 0.06, Litter: 5, Interval: 1, Burn: 0.012, Eat: 0.08, Gain: 0.5,
		Territory: 8, Crowd: 5, Start: 8, Habitat: hab(0.5, 0.8, 1, 0.8, 0.2, 0.2), Tame: "Ayam", Prey: true, Catch: 0.6},
	{ID: "kerbau", Name: "Kerbau Liar", Diet: Grazer, Size: 1.4, Run: 2.0, Walk: 0.4, Sight: 6, Flee: 1.5, HP: 4, Bite: 0.2,
		Fierce: true, Meat: 15, Adult: 3, Life: 20, Gestation: 0.85, Litter: 1, Interval: 2, Burn: 0.03, Eat: 0.8, Gain: 0.25,
		Territory: 12, Crowd: 10, Start: 2.5, Habitat: hab(1, 0.3, 0.4, 0.9, 0.3, 0.3), Tame: "Kerbau", Raider: true, Prey: true,
		Kill: "ditanduk kerbau", Catch: 0.25},
	{ID: "harimau", Name: "Harimau", Diet: Predator, Size: 1.1, Run: 3.6, Walk: 1.0, Sight: 10, HP: 4, Bite: 0.35,
		Fierce: true, Meat: 4, Adult: 3, Life: 15, Gestation: 0.3, Litter: 3, Interval: 2.5, Burn: 0.03,
		Territory: 22, Crowd: 1, Start: 0.7, Habitat: hab(0.5, 1, 0.8, 0.5, 0.1, 0.2), Kill: "diterkam harimau"},
}

// hab lists habitat preference for grass, forest, bush, meadow, sand, bare.
func hab(grass, forest, bush, meadow, sand, bare float64) [numCovers]float64 {
	var h [numCovers]float64
	h[coverGrass], h[coverForest], h[coverBush], h[coverMeadow], h[coverSand], h[coverBare] = grass, forest, bush, meadow, sand, bare
	return h
}

// SpeciesList returns the wild animals in a stable order (their index is the
// species number in stream frames).
func SpeciesList() []Species { return slices.Clone(species) }

// NumSpecies is how many kinds of animals there are.
var NumSpecies = len(species)

const (
	tiger          = 4    // index of the predator in species
	faunaCell      = 4.0  // tiles per spatial-hash cell
	recolonize     = 0.05 // yearly chance an extinct species arrives again
	mateSearch     = 25.0 // animals range widely to find a mate
	kidEnergy      = 0.6
	birthCost      = 0.15 // a mother's energy spent on a birth …
	birthCostKid   = 0.02 // … plus this much per young
	breedEnergy    = 0.5  // in good enough condition to breed
	chaseSeconds   = 4.0  // a tiger gives up a stalk that goes nowhere
	cubBurn        = 0.25 // a cub's own energy use while its mother feeds it
	bigMeal        = 5.0  // meat at which a kill fills a tiger
	fleeSeconds    = 1.2
	seekEvery      = 1.0
	leashHouse     = 3.5
	leashPen       = 1.2
	tameScavenge   = 0.01 // tameness gained per second scavenging by houses
	tameFeed       = 0.2  // tameness per meal from a person's hand
	tameFade       = 0.01 // tameness lost per second when away from people
	tameNoFear     = 0.5  // tame enough to stop fleeing
	stalkSpeed     = 1.6  // × walk while a tiger stalks
	pounceRange    = 1.8
	noticeRange    = 2.0  // prey may notice a stalking tiger this close …
	noticeChance   = 0.06 // … with this chance per look
	spotChance     = 0.2  // chance per look that prey notices a person within its flight distance
	pounceRest     = 0.5
	satedSeconds   = 0.25 * SecondsPerYear
	killEnergy     = 0.6
	huntWhenBelow  = 0.6
	manEaterBelow  = 0.35
	manEaterChance = 0.15
	manCatch       = 0.3
	lonelyRange    = 3.0
	pennedSafe     = 1.5
	raidDamage     = 0.15 // health per second a raider takes off a crop
)

// Animal is one wild or domestic animal.
type Animal struct {
	ID      int64   `json:"id"`
	Species uint8   `json:"species"`
	Female  bool    `json:"female"`
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
	Heading float64 `json:"heading"`
	Born    float64 `json:"born"` // simulated time
	Life    float64 `json:"life"` // seconds it will live, barring accidents
	Energy  float64 `json:"energy"`
	Health  float64 `json:"health"`
	Gest    float64 `json:"gest,omitempty"` // seconds of pregnancy left
	Rest    float64 `json:"rest,omitempty"` // seconds before it can breed again
	Tame    float64 `json:"tame,omitempty"` // 0–1 towards domestication
	Owner   int64   `json:"owner,omitempty"`
	Fear    float64 `json:"fear,omitempty"`  // seconds of fleeing left
	Sated   float64 `json:"sated,omitempty"` // a predator's seconds of rest after a kill
	Prey    int64   `json:"prey,omitempty"`  // a predator's target: an animal id, or minus a person's id
	TX      float64 `json:"tx,omitempty"`    // where it is heading to feed
	TY      float64 `json:"ty,omitempty"`
	Seek    float64 `json:"seek,omitempty"`  // seconds before looking for better grazing again
	Chase   float64 `json:"chase,omitempty"` // seconds a tiger has stalked its current target

	moving, eating, hunting bool // what it did this step, for the observer
}

// stalking reports whether a predator is out hunting; it follows from saved
// state so a restored world replays exactly.
func (a *Animal) stalking() bool {
	return a.Kind().Diet == Predator && a.Sated <= 0 && a.Energy < huntWhenBelow
}

// Dead reports whether the animal has died (it is removed at the next step).
func (a *Animal) Dead() bool { return a.Health <= 0 }

// Wild reports whether nobody owns it.
func (a *Animal) Wild() bool { return a.Owner == 0 }

// Kind is its species.
func (a *Animal) Kind() *Species { return &species[a.Species] }

// Flags for stream frames; keep in sync with ANIMAL_FLAG in frontend/src/sim/protocol.ts.
const (
	AnimalRunning = 1
	AnimalEating  = 2
	AnimalYoung   = 4
	AnimalTame    = 8
	AnimalHurt    = 16
	AnimalHunting = 32
)

// Flags describes what the animal looks like it is doing.
func (a *Animal) Flags(t float64) int {
	f := 0
	if a.moving && (a.Fear > 0 || a.hunting) {
		f |= AnimalRunning
	}
	if a.eating {
		f |= AnimalEating
	}
	if t-a.Born < a.Kind().Adult*SecondsPerYear {
		f |= AnimalYoung
	}
	if a.Owner != 0 {
		f |= AnimalTame
	}
	if a.Health < 0.5 {
		f |= AnimalHurt
	}
	if a.hunting {
		f |= AnimalHunting
	}
	return f
}

type fauna struct {
	animals []*Animal
	byID    map[int64]*Animal
	nextID  int64
	start   []int // count of each species when the world began (for caps and recolonising)
	extinct []bool
	gw, gh  int
	cells   [][]*Animal
}

func (f *fauna) init(e *Ecology) {
	f.byID = map[int64]*Animal{}
	f.nextID = 1
	f.start = make([]int, len(species))
	f.extinct = make([]bool, len(species))
	f.gw = int(math.Ceil(float64(e.land.W) / faunaCell))
	f.gh = int(math.Ceil(float64(e.land.H) / faunaCell))
	f.cells = make([][]*Animal, f.gw*f.gh)
}

// populate scatters the starting animals over their habitats, deer and
// buffalo in small herds.
func (e *Ecology) populate(t float64) {
	f := &e.fauna
	n := float64(len(e.walkable)) / 1000
	for si := range species {
		sp := &species[si]
		want := int(math.Round(sp.Start * n))
		if sp.Start > 0 {
			want = max(want, 2)
		}
		f.start[si] = want
		for placed := 0; placed < want; {
			x, y, ok := e.habitatSpot(si, 400)
			if !ok {
				break
			}
			herd := 1
			if sp.Diet == Grazer {
				herd = min(want-placed, 2+e.rng.IntN(3))
			}
			for range herd {
				a := e.newAnimal(si, x+e.rng.Float64()-0.5, y+e.rng.Float64()-0.5, t)
				// Start at a range of ages.
				a.Born = t - e.rng.Float64()*math.Min(a.Life, sp.Life*SecondsPerYear*0.6)
				if e.land.Blocked[e.tileOf(a.X, a.Y)] {
					a.X, a.Y = x, y
				}
				placed++
			}
		}
	}
	f.rebuild(e)
}

func (e *Ecology) tileOf(x, y float64) int {
	i, ok := e.land.indexAt(x, y)
	if !ok {
		return 0
	}
	return i
}

// habitatSpot picks a walkable tile the species likes, trying a few times.
func (e *Ecology) habitatSpot(si int, tries int) (float64, float64, bool) {
	sp := &species[si]
	best, bestScore := -1, 0.0
	for range tries {
		if len(e.walkable) == 0 {
			break
		}
		i := int(e.walkable[e.rng.IntN(len(e.walkable))])
		s := sp.Habitat[e.cover[i]] * e.rng.Float64()
		if si == 3 && e.nearFresh[i] {
			s *= 2 // buffalo wallow by rivers
		}
		if s > bestScore {
			best, bestScore = i, s
		}
		if bestScore > 0.8 {
			break
		}
	}
	if best < 0 {
		return 0, 0, false
	}
	return float64(best%e.land.W) + 0.5, float64(best/e.land.W) + 0.5, true
}

func (e *Ecology) newAnimal(si int, x, y, t float64) *Animal {
	f := &e.fauna
	sp := &species[si]
	a := &Animal{
		ID:      f.nextID,
		Species: uint8(si),
		Female:  e.rng.IntN(2) == 0,
		X:       x,
		Y:       y,
		Heading: e.rng.Float64()*2*math.Pi - math.Pi,
		Born:    t,
		Life:    sp.Life * SecondsPerYear * (0.7 + 0.6*e.rng.Float64()),
		Energy:  0.8,
		Health:  1,
	}
	f.nextID++
	f.animals = append(f.animals, a)
	f.byID[a.ID] = a
	return a
}

func (f *fauna) rebuild(e *Ecology) {
	for i := range f.cells {
		clear(f.cells[i])
		f.cells[i] = f.cells[i][:0]
	}
	for _, a := range f.animals {
		c := f.cellOf(a.X, a.Y)
		f.cells[c] = append(f.cells[c], a)
	}
}

func (f *fauna) cellOf(x, y float64) int {
	cx := min(f.gw-1, max(0, int(x/faunaCell)))
	cy := min(f.gh-1, max(0, int(y/faunaCell)))
	return cy*f.gw + cx
}

// Near calls fn for every living animal within r of (x, y) (and some a bit
// further; callers check the distance).
func (e *Ecology) Near(x, y, r float64, fn func(*Animal)) {
	f := &e.fauna
	x0, x1 := max(0, int((x-r)/faunaCell)), min(f.gw-1, int((x+r)/faunaCell))
	y0, y1 := max(0, int((y-r)/faunaCell)), min(f.gh-1, int((y+r)/faunaCell))
	for cy := y0; cy <= y1; cy++ {
		for cx := x0; cx <= x1; cx++ {
			for _, a := range f.cells[cy*f.gw+cx] {
				if !a.Dead() {
					fn(a)
				}
			}
		}
	}
}

// Animals lists every animal; the slice must not be kept.
func (e *Ecology) Animals() []*Animal { return e.fauna.animals }

// Animal returns the animal with id if it is alive.
func (e *Ecology) Animal(id int64) *Animal {
	if a := e.fauna.byID[id]; a != nil && !a.Dead() {
		return a
	}
	return nil
}

// Counts are the living animals of each species, wild and tame.
func (e *Ecology) Counts() (wild, tame []int) {
	wild = make([]int, len(species))
	tame = make([]int, len(species))
	for _, a := range e.fauna.animals {
		if a.Dead() {
			continue
		}
		if a.Owner != 0 {
			tame[a.Species]++
		} else {
			wild[a.Species]++
		}
	}
	return
}

func (e *Ecology) habitatAt(sp *Species, x, y float64) float64 {
	i, ok := e.land.indexAt(x, y)
	if !ok || e.land.Blocked[i] {
		return 0
	}
	return sp.Habitat[e.cover[i]]
}

func (e *Ecology) passable(x, y float64) bool {
	i, ok := e.land.indexAt(x, y)
	return ok && !e.land.Blocked[i]
}

// step moves, feeds and breeds every animal by dt seconds.
func (f *fauna) step(e *Ecology, t, dt float64, h Humans) {
	f.rebuild(e)
	for _, a := range f.animals {
		if a.Dead() {
			continue
		}
		e.live(a, t, dt, h)
	}
	// Remove the dead.
	alive := f.animals[:0]
	for _, a := range f.animals {
		if a.Dead() {
			delete(f.byID, a.ID)
			continue
		}
		alive = append(alive, a)
	}
	clear(f.animals[len(alive):])
	f.animals = alive
	e.checkExtinctions(t)
}

func (e *Ecology) live(a *Animal, t, dt float64, h Humans) {
	sp := a.Kind()
	a.moving, a.eating, a.hunting = false, false, false
	burn := sp.Burn
	young := t-a.Born < sp.Adult*SecondsPerYear
	if young && sp.Diet == Predator {
		burn *= cubBurn // the mother brings her cubs food
	}
	a.Energy -= burn * dt
	a.Fear = math.Max(0, a.Fear-dt)
	a.Sated = math.Max(0, a.Sated-dt)
	a.Rest = math.Max(0, a.Rest-dt)
	a.Seek -= dt
	if a.Energy <= 0 {
		a.Health = 0
		e.stats.cur.AnimalStarved[a.Species]++
		return
	}
	if t-a.Born >= a.Life {
		a.Health = 0
		return
	}
	if a.Health < 1 {
		a.Health = math.Min(1, a.Health+0.02*dt)
	}
	if a.Gest > 0 {
		if a.Gest -= dt; a.Gest <= 0 {
			e.giveBirth(a, t)
		}
	}
	e.tameStep(a, dt, h)

	speed, heading := 0.0, a.Heading
	switch {
	case sp.Diet == Predator && young:
		speed, heading = e.followMother(a)
	case sp.Diet == Predator:
		speed, heading = e.prowl(a, t, dt, h)
	default:
		if tx, ty, ok := e.threat(a, h); ok {
			a.Fear = fleeSeconds
			heading = math.Atan2(a.Y-ty, a.X-tx) + (e.rng.Float64()-0.5)*0.6
			speed = sp.Run
		} else if a.Fear > 0 {
			speed = sp.Run
		} else {
			speed, heading = e.browse(a, dt)
		}
	}
	if a.Owner != 0 {
		speed, heading = e.leash(a, speed, heading, h)
	}
	e.move(a, heading, speed*dt)
	if a.Female && a.Gest <= 0 && a.Rest <= 0 && a.Energy > breedEnergy && t-a.Born >= sp.Adult*SecondsPerYear {
		e.tryBreed(a, t)
	}
}

// followMother keeps a cub near the nearest adult female of its kind.
func (e *Ecology) followMother(a *Animal) (float64, float64) {
	sp := a.Kind()
	var mx, my float64
	bestD := sp.Sight * 2
	found := false
	e.Near(a.X, a.Y, bestD, func(o *Animal) {
		if o.Species == a.Species && o.Female && o != a {
			if d := math.Hypot(o.X-a.X, o.Y-a.Y); d < bestD {
				mx, my, bestD, found = o.X, o.Y, d, true
			}
		}
	})
	if !found || bestD < 1.5 {
		if e.rng.Float64() < 0.1 {
			a.Heading += (e.rng.Float64() - 0.5) * 1.5
		}
		return sp.Walk * 0.3, a.Heading
	}
	return sp.Walk * 1.5, math.Atan2(my-a.Y, mx-a.X)
}

// threat returns the nearest danger a prey animal flees: a person (unless it
// is tame) or a tiger close enough to notice.
func (e *Ecology) threat(a *Animal, h Humans) (float64, float64, bool) {
	sp := a.Kind()
	flee := sp.Flee * (1 - a.Tame)
	if a.Owner != 0 || a.Tame >= tameNoFear {
		flee = 0
	}
	// A person creeping up is noticed only now and then: stalking works.
	if flee > 0 && h != nil && e.rng.Float64() < spotChance {
		if _, hx, hy, ok := h.Nearest(a.X, a.Y, flee); ok {
			return hx, hy, true
		}
	}
	if sp.Prey {
		var tx, ty float64
		found := false
		e.Near(a.X, a.Y, noticeRange, func(o *Animal) {
			if !found && o.Species == tiger && o.stalking() && math.Hypot(o.X-a.X, o.Y-a.Y) <= noticeRange {
				tx, ty, found = o.X, o.Y, e.rng.Float64() < noticeChance
			}
		})
		if found {
			return tx, ty, true
		}
	}
	return 0, 0, false
}

// browse eats within reach (the tile it stands on and the eight around it),
// or walks towards better food.
func (e *Ecology) browse(a *Animal, dt float64) (float64, float64) {
	sp := a.Kind()
	if a.Energy < 0.9 {
		want := float32(sp.Eat * dt)
		took := e.eatAround(a, want)
		a.Energy = math.Min(1, a.Energy+float64(took)*sp.Gain)
		if took >= want*0.5 {
			a.eating = true
			return 0, a.Heading
		}
		if a.Seek <= 0 {
			a.Seek = seekEvery
			e.seekFood(a)
		}
		if a.TX != 0 || a.TY != 0 {
			if math.Hypot(a.TX-a.X, a.TY-a.Y) > 0.4 {
				return sp.Walk * 1.5, math.Atan2(a.TY-a.Y, a.TX-a.X)
			}
			a.TX, a.TY = 0, 0
		}
	}
	// Wander, turning now and then.
	if e.rng.Float64() < 0.1 {
		a.Heading += (e.rng.Float64() - 0.5) * 1.6
	}
	if e.rng.Float64() < 0.4 {
		return 0, a.Heading
	}
	return sp.Walk * 0.6, a.Heading
}

// eatAround takes up to want food units from the tiles within reach, and
// lets raiders eat and trample a field there.
func (e *Ecology) eatAround(a *Animal, want float32) float32 {
	sp := a.Kind()
	x, y := int(math.Floor(a.X)), int(math.Floor(a.Y))
	var took float32
	for k := 0; k < 9 && took < want; k++ {
		i, ok := e.land.index(x+k%3-1, y+k/3-1)
		if !ok || e.land.Blocked[i] {
			continue
		}
		if sp.Raider && a.Owner == 0 && e.plotAt[i] != 0 {
			// Fields are the best food around: raiders eat and trample them.
			if e.trample(i, float32(raidDamage)*want/float32(sp.Eat)) {
				took = want
				break
			}
		}
		pool := &e.graze[i]
		if sp.Diet == Forager {
			pool = &e.forage[i]
		}
		bite := min(*pool, want-took)
		*pool -= bite
		took += bite
	}
	return took
}

// seekFood looks around for a tile with more food in a liked habitat.
func (e *Ecology) seekFood(a *Animal) {
	sp := a.Kind()
	best := -1.0
	r := 4.0
	if a.Energy < 0.4 {
		r = 9 // hunger drives it further afield
	}
	for range 6 {
		x := a.X + (e.rng.Float64()*2-1)*r
		y := a.Y + (e.rng.Float64()*2-1)*r
		i, ok := e.land.indexAt(x, y)
		if !ok || e.land.Blocked[i] {
			continue
		}
		food := float64(e.graze[i])
		if sp.Diet == Forager {
			food = float64(e.forage[i])
		}
		if sp.Raider && e.plotAt[i] != 0 {
			food += 0.5
		}
		s := food * sp.Habitat[e.cover[i]]
		if s > best {
			best, a.TX, a.TY = s, x, y
		}
	}
}

// leash keeps livestock by its home: within a few tiles of the house, or
// close to the pen.
func (e *Ecology) leash(a *Animal, speed, heading float64, h Humans) (float64, float64) {
	if h == nil {
		return speed, heading
	}
	hx, hy, pen, ok := h.Home(a.Owner)
	if !ok {
		a.Owner = 0 // the family is gone: it goes feral
		a.Tame = 0.6
		return speed, heading
	}
	r := leashHouse
	if pen {
		r = leashPen
	}
	if d := math.Hypot(hx-a.X, hy-a.Y); d > r {
		return math.Max(speed, a.Kind().Walk*1.5), math.Atan2(hy-a.Y, hx-a.X)
	}
	return speed, heading
}

// move steps forward, turning away from walls, water and land it avoids.
func (e *Ecology) move(a *Animal, heading, dist float64) {
	a.Heading = heading
	if dist <= 0 {
		return
	}
	sp := a.Kind()
	for try := range 4 {
		nx, ny := a.X+math.Cos(a.Heading)*dist, a.Y+math.Sin(a.Heading)*dist
		ok := e.passable(nx, ny)
		// Calm animals keep to their habitat; frightened ones go anywhere.
		if ok && a.Fear <= 0 && !a.hunting && a.Owner == 0 && try < 3 && e.habitatAt(sp, nx, ny) < 0.15 {
			ok = false
		}
		if ok {
			a.X, a.Y = nx, ny
			a.moving = true
			return
		}
		a.Heading += math.Pi/2 + e.rng.Float64()*math.Pi
		a.TX, a.TY = 0, 0
	}
}

// tryBreed may make a healthy adult female pregnant if a male is near. The
// more adults of her kind share her neighbourhood, the less likely she
// breeds (density dependence: territoriality, stress, competition), so
// numbers level off below what the land could feed in a good year.
func (e *Ecology) tryBreed(a *Animal, t float64) {
	sp := a.Kind()
	f := &e.fauna
	if f.start[a.Species] > 0 && e.count(int(a.Species)) >= max(20, 4*f.start[a.Species]) {
		return // technical cap, far above what the land supports
	}
	male, crowd := false, 0
	r := math.Max(sp.Territory, mateSearch)
	adult := sp.Adult * SecondsPerYear
	e.Near(a.X, a.Y, r, func(o *Animal) {
		if o == a || o.Species != a.Species || t-o.Born < adult {
			return
		}
		d := math.Hypot(o.X-a.X, o.Y-a.Y)
		if d <= sp.Territory {
			crowd++
		}
		if !o.Female && d <= mateSearch && o.Energy > 0.4 {
			male = true
		}
	})
	if male && e.rng.Float64() < 1-float64(crowd)/float64(sp.Crowd+1) {
		a.Gest = sp.Gestation * SecondsPerYear
	} else {
		a.Rest = 0.1 * SecondsPerYear // try again in a while
	}
}

func (e *Ecology) count(si int) int {
	n := 0
	for _, a := range e.fauna.animals {
		if int(a.Species) == si && !a.Dead() {
			n++
		}
	}
	return n
}

func (e *Ecology) giveBirth(m *Animal, t float64) {
	sp := m.Kind()
	n := 1 + e.rng.IntN(sp.Litter)
	for range n {
		kid := e.newAnimal(int(m.Species), m.X+(e.rng.Float64()-0.5)*0.4, m.Y+(e.rng.Float64()-0.5)*0.4, t)
		kid.Energy = kidEnergy
		kid.Owner = m.Owner
		kid.Tame = m.Tame
		if !e.passable(kid.X, kid.Y) {
			kid.X, kid.Y = m.X, m.Y
		}
	}
	m.Gest = 0
	m.Energy -= birthCost + birthCostKid*float64(n)
	m.Rest = sp.Interval * SecondsPerYear
	e.stats.cur.AnimalBirths[m.Species] += n
}

// tameStep: chickens, pigs and buffalo that hang around houses slowly lose
// their fear of people (the commensal path to domestication); once tame
// enough they belong to the household. Away from people they turn wild again.
func (e *Ecology) tameStep(a *Animal, dt float64, h Humans) {
	sp := a.Kind()
	if sp.Tame == "" || a.Owner != 0 || h == nil {
		return
	}
	if house, ok := h.Settled(a.X, a.Y); ok {
		a.Tame = math.Min(1, a.Tame+tameScavenge*dt)
		if a.Tame >= 1 {
			e.adopt(a, house, h)
		}
		return
	}
	a.Tame = math.Max(0, a.Tame-tameFade*dt)
}

func (e *Ecology) adopt(a *Animal, house int64, h Humans) {
	a.Owner = house
	a.Tame = 1
	a.Fear = 0
	e.stats.cur.Tamed++
	if h != nil {
		h.Tamed(house, a.Kind().ID)
	}
}

// Feed gives an animal a meal from a person's hand. Wild animals grow tamer;
// when tame enough they join household house (if it has one). It reports
// whether the animal became livestock just now.
func (e *Ecology) Feed(a *Animal, energy float64, house int64, h Humans) bool {
	a.Energy = math.Min(1, a.Energy+energy)
	sp := a.Kind()
	if a.Owner != 0 || sp.Tame == "" {
		return false
	}
	a.Tame = math.Min(1, a.Tame+tameFeed)
	if a.Tame >= 1 && house != 0 {
		e.adopt(a, house, h)
		return true
	}
	return false
}

// Tameable reports whether the animal could become livestock.
func (a *Animal) Tameable() bool { return a.Kind().Tame != "" }

// Strike is a blow from a person: damage in units of a bare-handed blow. It
// reports whether the animal died, and the animal flees.
func (e *Ecology) Strike(a *Animal, damage float64) (killed bool) {
	sp := a.Kind()
	a.Health -= damage / sp.HP
	a.Fear = fleeSeconds * 2
	if a.Health <= 0 {
		e.stats.cur.Hunted[a.Species]++
		e.stats.Hunted[a.Species]++
		return true
	}
	return false
}

// Slaughter kills a household's animal for its meat and returns the meat.
func (e *Ecology) Slaughter(a *Animal) int {
	a.Health = 0
	e.stats.cur.Slaughtered++
	return max(1, a.Kind().Meat/2+1)
}

// Livestock lists the living animals household house owns.
func (e *Ecology) Livestock(house int64) []*Animal {
	var out []*Animal
	for _, a := range e.fauna.animals {
		if a.Owner == house && house != 0 && !a.Dead() {
			out = append(out, a)
		}
	}
	return out
}

// prowl is a tiger's turn: rest after a kill, otherwise stalk the nearest
// prey (rarely a lone person, when starving) and pounce from close by.
func (e *Ecology) prowl(a *Animal, t, dt float64, h Humans) (float64, float64) {
	sp := a.Kind()
	if a.Sated > 0 || a.Energy >= huntWhenBelow {
		a.Prey = 0
		if e.rng.Float64() < 0.08 {
			a.Heading += (e.rng.Float64() - 0.5) * 1.5
		}
		if e.rng.Float64() < 0.5 {
			return 0, a.Heading
		}
		return sp.Walk * 0.5, a.Heading
	}
	a.hunting = true
	tx, ty, ok := e.preyPos(a, h)
	if a.Chase += dt; ok && a.Chase > chaseSeconds {
		a.Prey, ok = 0, false // it is getting nowhere: look for another
	}
	if !ok {
		a.Chase = 0
		a.Prey = e.choosePrey(a, h)
		if tx, ty, ok = e.preyPos(a, h); !ok {
			// Nothing in sight: roam the territory.
			if e.rng.Float64() < 0.1 {
				a.Heading += (e.rng.Float64() - 0.5) * 1.5
			}
			return sp.Walk, a.Heading
		}
	}
	d := math.Hypot(tx-a.X, ty-a.Y)
	heading := math.Atan2(ty-a.Y, tx-a.X)
	if d > pounceRange {
		return math.Min(sp.Walk*stalkSpeed, d/dt), heading
	}
	// Pounce.
	if a.Prey > 0 {
		prey := e.fauna.byID[a.Prey]
		if prey != nil && !prey.Dead() && e.rng.Float64() < prey.Kind().Catch {
			prey.Health = 0
			e.stats.cur.Predated[prey.Species]++
			a.Energy = math.Min(1, a.Energy+killEnergy*math.Min(1, float64(prey.Kind().Meat)/bigMeal))
			a.Sated = satedSeconds
		} else if prey != nil {
			prey.Fear = fleeSeconds * 2
		}
	} else if h != nil && e.rng.Float64() < manCatch {
		h.Maul(-a.Prey, sp.Bite*2, sp.ID)
		a.Energy = math.Min(1, a.Energy+killEnergy*0.5)
		a.Sated = satedSeconds
	}
	a.Prey = 0
	a.Sated = math.Max(a.Sated, pounceRest) // catch its breath after a miss
	return 0, heading
}

func (e *Ecology) preyPos(a *Animal, h Humans) (float64, float64, bool) {
	switch {
	case a.Prey > 0:
		if p := e.fauna.byID[a.Prey]; p != nil && !p.Dead() && !e.penned(p, h) &&
			math.Hypot(p.X-a.X, p.Y-a.Y) <= a.Kind().Sight*1.5 {
			return p.X, p.Y, true
		}
	case a.Prey < 0 && h != nil:
		if id, hx, hy, ok := h.Nearest(a.X, a.Y, a.Kind().Sight); ok && id == -a.Prey {
			return hx, hy, true
		}
	}
	a.Prey = 0
	return 0, 0, false
}

// choosePrey picks the nearest animal worth hunting, or, for a starving
// tiger, a person walking alone.
func (e *Ecology) choosePrey(a *Animal, h Humans) int64 {
	sp := a.Kind()
	// The best meal for the walk: meat over distance.
	var best *Animal
	bestScore := 0.0
	e.Near(a.X, a.Y, sp.Sight, func(o *Animal) {
		if !o.Kind().Prey || e.penned(o, h) {
			return
		}
		d := math.Hypot(o.X-a.X, o.Y-a.Y)
		if d > sp.Sight {
			return
		}
		if s := math.Min(float64(o.Kind().Meat), bigMeal) / (1 + d); s > bestScore {
			best, bestScore = o, s
		}
	})
	if best != nil {
		return best.ID
	}
	if a.Energy < manEaterBelow && h != nil && e.rng.Float64() < manEaterChance {
		if id, hx, hy, ok := h.Nearest(a.X, a.Y, sp.Sight); ok && h.Crowd(hx, hy, lonelyRange) <= 1 {
			return -id
		}
	}
	return 0
}

// penned reports whether livestock is safe in its pen.
func (e *Ecology) penned(a *Animal, h Humans) bool {
	if a.Owner == 0 || h == nil {
		return false
	}
	hx, hy, pen, ok := h.Home(a.Owner)
	return ok && pen && math.Hypot(hx-a.X, hy-a.Y) <= pennedSafe
}

// checkExtinctions notes species that died out on the island; now and then a
// pair swims or rafts over from elsewhere (island biogeography: MacArthur &
// Wilson 1967), checked once a year.
func (e *Ecology) checkExtinctions(t float64) {
	f := &e.fauna
	wild, tame := e.Counts()
	for si := range species {
		alive := wild[si]+tame[si] > 0
		switch {
		case alive:
			f.extinct[si] = false
		case !f.extinct[si] && f.start[si] > 0:
			f.extinct[si] = true
			e.stats.cur.Extinct = append(e.stats.cur.Extinct, species[si].ID)
			e.event("ecology", fmt.Sprintf("%s punah di pulau ini (tahun %d)", species[si].Name, e.clim.Year+1))
		}
	}
}

// recolonise runs once a year.
func (e *Ecology) recolonise(t float64) {
	f := &e.fauna
	if e.opts.NoFauna {
		return
	}
	for si := range species {
		if !f.extinct[si] || e.rng.Float64() >= recolonize {
			continue
		}
		x, y, ok := e.coastSpot(si)
		if !ok {
			continue
		}
		m := e.newAnimal(si, x, y, t-species[si].Adult*SecondsPerYear)
		m.Female = false
		w := e.newAnimal(si, x+0.3, y, t-species[si].Adult*SecondsPerYear)
		w.Female = true
		f.extinct[si] = false
		e.stats.cur.Arrived = append(e.stats.cur.Arrived, species[si].ID)
		e.event("ecology", fmt.Sprintf("Sepasang %s tiba dari seberang laut (tahun %d)", lower(species[si].Name), e.clim.Year+1))
	}
}

func (e *Ecology) coastSpot(si int) (float64, float64, bool) {
	for range 200 {
		i := int(e.walkable[e.rng.IntN(len(e.walkable))])
		if e.nearSea[i] && species[si].Habitat[e.cover[i]] > 0.2 {
			return float64(i%e.land.W) + 0.5, float64(i/e.land.W) + 0.5, true
		}
	}
	return e.habitatSpot(si, 100)
}

func lower(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 'a' - 'A'
		}
	}
	return string(b)
}

// AddAnimal puts an adult of species id at (x, y) at time t, for tests and
// tools; it returns nil for an unknown species.
func (e *Ecology) AddAnimal(id string, x, y, t float64) *Animal {
	for si, sp := range species {
		if sp.ID == id {
			a := e.newAnimal(si, x, y, t-sp.Adult*SecondsPerYear)
			e.fauna.rebuild(e)
			return a
		}
	}
	return nil
}
