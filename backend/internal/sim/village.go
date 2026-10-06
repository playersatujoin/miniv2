package sim

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
)

// Villages: houses close together make a village with its own land (the
// hull around its houses and fields, after RAGE's map zones, game/MapZones.h
// CConvexHull2D), its people, a name, and a leader: the person its people
// trust most (after Peds/Relationships.h and PedGroup/PedGroup.h). Nobody is
// told to follow the leader; people only sense where they are.
//
// Every few seconds the inhabited houses are clustered (village_land.go draws
// the land, village_leader.go counts the trust); villages keep their identity
// across clusterings by the houses they share. Who lives in which village is
// refreshed every second from the houses people live in.

const (
	// villageEvery: houses are clustered twice a simulated year. A house
	// takes seconds to build and a village years to grow, so a finer clock
	// would only cost time. Absolute ticks keep a restored world on the beat.
	villageEvery = SecondsPerYear / 2 * TicksPerSecond

	// villageLink: two inhabited houses this close belong to one settlement.
	// Houses stand at least houseSpacing (3) apart and a family gardens
	// within gardenRange (8) of its house, so within 8 tiles neighbours work
	// the same ground and share paths: one hamlet. Further apart their
	// gardens no longer meet. The map's space is compressed, so this follows
	// the world's own distances rather than metres.
	villageLink = gardenRange

	// A cluster becomes a village with three houses, or with two that are
	// home to six or more people (two extended families); a lone homestead
	// or a couple of huts is not one. Once founded it lasts while two of its
	// houses are lived in. Estimates: the smallest named settlements
	// (Indonesian dusun, hamlets elsewhere) are a handful of households.
	villageMinHouses  = 3
	villagePairPeople = 6
	villageKeepHouses = 2

	// landMargin: the land reaches this far past the outermost house, field
	// or building: yards, paths and the edges of the gardens (estimate,
	// about homeReach plus half a tile).
	landMargin = 2.5
	// villageReach: a field or building belongs to the village of the
	// nearest house within this distance: the same reach as the gardens.
	villageReach = gardenRange

	// villageNewsPerTick: village reports per check, so the first clustering
	// of an old world with many houses doesn't flood the log.
	villageNewsPerTick = 4
)

// village is one village. Its exported fields are saved.
type village struct {
	ID      int64   `json:"id"`
	Name    string  `json:"name"`
	Hue     int     `json:"hue"`
	Founded float64 `json:"founded"` // simulated seconds
	// Inhabited houses by ID, its other buildings, and its families' planted
	// fields, at the last clustering.
	Houses    []int64 `json:"houses"`
	Buildings []int64 `json:"buildings,omitempty"`
	Fields    int     `json:"fields,omitempty"`
	// Land: counter-clockwise hull in tiles, and the centre of its houses.
	Hull [][2]float64 `json:"hull"`
	X    float64      `json:"x"`
	Y    float64      `json:"y"`
	// People living in its houses at the last count.
	People int `json:"people"`
	// Leader, since when, their trust and supporters at the last count, and
	// how many leaders it has had.
	Leader      int64   `json:"leader,omitempty"`
	LeaderName  string  `json:"leaderName,omitempty"`
	LeaderSince float64 `json:"leaderSince,omitempty"`
	Trust       float64 `json:"trust,omitempty"`
	Supporters  int     `json:"supporters,omitempty"`
	// Standing is the sitting leader's trust remembered over a few years
	// (see standingYears): what they keep office by.
	Standing float64 `json:"standing,omitempty"`
	Leaders  int     `json:"leaders,omitempty"`
	LastNews float64 `json:"lastNews,omitempty"` // when a leadership change was last reported

	members []*Creature // residents, refreshed every second
	area    int         // tiles of land, from the land grid
}

// villageSys is a world's villages; vil() finds it.
// vil is this world's village state.
func (s *Sim) vil() *villageSys { return &s.village }

type villageSys struct {
	list    []*village // by ID: oldest first
	byID    map[int64]*village
	byHouse map[int64]*village
	land    []int32 // per tile, the village whose land it is (0 none)
	lastID  int64
	version int64
	body    []byte // the encoded village list, to tell when it changes
	// All-time counts.
	founded, abandoned, leaderChanges int
	// news is the reports left in this check; held counts those left out.
	news, held int
	tally      map[int64]tally // scratch for elections
}

// villageState is the saved villages. The land grid, the lookups and the
// encoded list are rebuilt from it.
type villageState struct {
	List          []*village `json:"list,omitempty"`
	LastID        int64      `json:"lastId,omitempty"`
	Version       int64      `json:"version,omitempty"`
	Founded       int        `json:"founded,omitempty"`
	Abandoned     int        `json:"abandoned,omitempty"`
	LeaderChanges int        `json:"leaderChanges,omitempty"`
}

// VillageView is one village for the observer: the `villages` stream fields
// and what the village panel shows besides.
type VillageView struct {
	ID      int64        `json:"id"`
	Name    string       `json:"name"`
	Hue     int          `json:"hue"`
	Hull    [][2]float64 `json:"hull"`
	X       float64      `json:"x"`
	Y       float64      `json:"y"`
	People  int          `json:"people"`
	Houses  int          `json:"houses"`
	Leader  *Ref         `json:"leader"`
	Founded float64      `json:"founded"`

	Area        int               `json:"area"` // tiles of land
	HouseIDs    []int64           `json:"houseIds"`
	Fields      int               `json:"fields"` // planted plots of its families
	Buildings   []VillageBuilding `json:"buildings"`
	LeaderSince float64           `json:"leaderSince,omitempty"`
	LeaderTrust float64           `json:"leaderTrust"` // the leader's summed trust among the adults
	Supporters  int               `json:"supporters"`
	Leaders     int               `json:"leaders"` // how many people have led it
	Residents   []Villager        `json:"residents"`
	Attitudes   []VillageAttitude `json:"attitudes"`
	// Food stored in its houses and in its granaries (sharing.go).
	HouseFood   int `json:"houseFood"`
	GranaryFood int `json:"granaryFood"`
}

// VillageBuilding is a station or other building of a village.
type VillageBuilding struct {
	ID   int64  `json:"id"`
	Kind string `json:"kind"`
	Name string `json:"name"`
	X    int    `json:"x"`
	Y    int    `json:"y"`
}

// Villager is someone living in a village.
type Villager struct {
	ID     int64   `json:"id"`
	Name   string  `json:"name"`
	House  int64   `json:"house"`
	Adult  bool    `json:"adult"`
	Trust  float64 `json:"trust"` // their trust in the leader (0 if they have no opinion)
	Leader bool    `json:"leader,omitempty"`
}

// VillageAttitude is how one village's adults regard another village: the
// mean trust they have in the people of it they know, and a word for it.
type VillageAttitude struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	Trust    float64 `json:"trust"`
	Known    int     `json:"known"`    // acquaintances counted
	Attitude string  `json:"attitude"` // hormat, suka, acuh, tidak suka, benci, or asing
}

// VillageInfo summarises villages for the world info.
type VillageInfo struct {
	Villages      int `json:"villages"`
	Villagers     int `json:"villagers"`     // people living in a village
	Led           int `json:"led"`           // villages with a leader
	Largest       int `json:"largest"`       // people in the biggest village
	Founded       int `json:"founded"`       // villages ever recognised
	Abandoned     int `json:"abandoned"`     // villages ever abandoned or merged into another
	LeaderChanges int `json:"leaderChanges"` // times a village got a new leader
}

// villageTick finds villages, their land, people and leaders; called once a
// simulated second (it may do its work less often).
func (s *Sim) villageTick() {
	vs := s.vil()
	if s.opts.NoVillages {
		if len(vs.list) > 0 {
			s.dropVillages()
		}
		return
	}
	vs.news, vs.held = villageNewsPerTick, 0
	if s.tick%villageEvery == 0 {
		s.recluster()
	} else {
		s.refreshVillagers()
	}
	s.succession()
	if vs.held > 0 {
		s.event("village", fmt.Sprintf("Ada %d perubahan desa lainnya", vs.held), 0)
	}
	s.encodeVillages()
}

// spendNews takes one report from this check's budget.
func (s *Sim) spendNews() bool {
	vs := s.vil()
	if vs.news <= 0 {
		return false
	}
	vs.news--
	return true
}

// villageEvent reports a founding, merger or abandonment, or counts it for
// the summary when the budget is spent.
func (s *Sim) villageEvent(text string) {
	if s.spendNews() {
		s.event("village", text, 0)
	} else {
		s.vil().held++
	}
}

// dropVillages forgets every village (the switch was turned off).
func (s *Sim) dropVillages() {
	vs := s.vil()
	vs.list = nil
	s.indexVillages()
	clear(vs.land)
	for _, c := range s.creatures {
		c.VillageID = 0
	}
	s.encodeVillages()
}

// indexVillages sorts the list and rebuilds the lookups.
func (s *Sim) indexVillages() {
	vs := s.vil()
	slices.SortFunc(vs.list, func(a, b *village) int { return cmp.Compare(a.ID, b.ID) })
	vs.byID = make(map[int64]*village, len(vs.list))
	vs.byHouse = map[int64]*village{}
	for _, v := range vs.list {
		vs.byID[v.ID] = v
		for _, h := range v.Houses {
			vs.byHouse[h] = v
		}
	}
}

// refreshVillagers sets everyone's VillageID from the house they live in and
// recounts each village's people.
func (s *Sim) refreshVillagers() {
	vs := s.vil()
	for _, v := range vs.list {
		clear(v.members)
		v.members = v.members[:0]
	}
	for _, c := range s.creatures {
		var v *village
		if c.HouseID != 0 && c.Health > 0 {
			v = vs.byHouse[c.HouseID]
		}
		if v == nil {
			c.VillageID = 0
			continue
		}
		c.VillageID = v.ID
		v.members = append(v.members, c)
	}
	for _, v := range vs.list {
		v.People = len(v.members)
	}
}

// cluster is a group of inhabited houses linked by villageLink.
type cluster struct {
	houses []*Structure // by ID
	people int
	shared map[int64]int // houses each old village has in it
	v      *village
	from   *village // for a split: the old village it broke away from
}

// recluster groups the inhabited houses, matches the groups to the villages
// there were (a merger keeps the older village, a part that splits off is
// a new one), and redraws their land and counts their leaders.
func (s *Sim) recluster() {
	vs := s.vil()
	clusters, cells := s.clusterHouses()

	// Old villages in age order each take the cluster holding most of their
	// houses, if it is still big enough and no older village took it.
	for _, old := range vs.list {
		var best *cluster
		for _, cl := range clusters {
			if cl.v != nil || len(cl.houses) < villageKeepHouses || cl.shared[old.ID] == 0 {
				continue
			}
			if best == nil || cl.shared[old.ID] > best.shared[old.ID] {
				best = cl
			}
		}
		if best != nil {
			best.v = old
		}
	}
	kept := map[*village]bool{}
	for _, cl := range clusters {
		if cl.v != nil {
			kept[cl.v] = true
		}
	}
	// The villages no cluster kept merged into another or were left.
	for _, old := range vs.list {
		if kept[old] {
			continue
		}
		var into *cluster
		for _, cl := range clusters {
			if cl.v != nil && cl.shared[old.ID] > 0 && (into == nil || cl.shared[old.ID] > into.shared[old.ID]) {
				into = cl
			}
		}
		vs.abandoned++
		if into != nil {
			s.villageEvent(fmt.Sprintf("Desa %s bergabung dengan Desa %s", old.Name, into.v.Name))
		} else {
			s.villageEvent(fmt.Sprintf("Desa %s ditinggalkan", old.Name))
		}
	}
	// Clusters left over that are big enough are new villages.
	var list []*village
	for _, cl := range clusters {
		if cl.v == nil && (len(cl.houses) >= villageMinHouses || len(cl.houses) >= 2 && cl.people >= villagePairPeople) {
			for _, old := range vs.list {
				if cl.shared[old.ID] > 0 && (cl.from == nil || cl.shared[old.ID] > cl.shared[cl.from.ID]) {
					cl.from = old
				}
			}
			cl.v = s.newVillage(list)
			if cl.from != nil {
				s.villageEvent(fmt.Sprintf("Desa %s memisahkan diri dari Desa %s dengan %d rumah", cl.v.Name, cl.from.Name, len(cl.houses)))
			} else {
				s.villageEvent(fmt.Sprintf("Desa %s terbentuk dari %d rumah", cl.v.Name, len(cl.houses)))
			}
		}
		if cl.v != nil {
			cl.v.Houses = cl.v.Houses[:0]
			for _, h := range cl.houses {
				cl.v.Houses = append(cl.v.Houses, h.ID)
			}
			list = append(list, cl.v)
		}
	}
	vs.list = list
	s.indexVillages()
	s.surveyLand(cells)
	s.paintLand()
	s.refreshVillagers()
	for _, v := range vs.list {
		s.elect(v)
	}
}

// houseCell is the bucket of villageLink-sized cells a house falls in.
func houseCell(x, y int) [2]int { return [2]int{x / villageLink, y / villageLink} }

// clusterHouses groups the inhabited houses: union-find over every pair
// within villageLink, found through cells of that size, in ID order. It
// returns the clusters by their first house and the cells.
func (s *Sim) clusterHouses() ([]*cluster, map[[2]int][]*Structure) {
	vs := s.vil()
	var houses []*Structure
	for _, st := range s.structures {
		if st.house() && st.Owner != 0 {
			houses = append(houses, st)
		}
	}
	slices.SortFunc(houses, func(a, b *Structure) int { return cmp.Compare(a.ID, b.ID) })
	index := make(map[int64]int, len(houses))
	cells := map[[2]int][]*Structure{}
	for i, h := range houses {
		index[h.ID] = i
		k := houseCell(h.X, h.Y)
		cells[k] = append(cells[k], h)
	}
	parent := make([]int, len(houses))
	for i := range parent {
		parent[i] = i
	}
	find := func(i int) int {
		for parent[i] != i {
			parent[i] = parent[parent[i]]
			i = parent[i]
		}
		return i
	}
	for i, h := range houses {
		k := houseCell(h.X, h.Y)
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				for _, o := range cells[[2]int{k[0] + dx, k[1] + dy}] {
					j := index[o.ID]
					if j <= i || math.Hypot(float64(o.X-h.X), float64(o.Y-h.Y)) > villageLink {
						continue
					}
					if a, b := find(i), find(j); a != b {
						parent[max(a, b)] = min(a, b) // the older house is the root
					}
				}
			}
		}
	}
	people := map[int64]int{}
	for _, c := range s.creatures {
		if c.HouseID != 0 && c.Health > 0 {
			people[c.HouseID]++
		}
	}
	byRoot := map[int]*cluster{}
	var out []*cluster
	for i, h := range houses {
		r := find(i)
		cl := byRoot[r]
		if cl == nil {
			cl = &cluster{shared: map[int64]int{}}
			byRoot[r] = cl
			out = append(out, cl)
		}
		cl.houses = append(cl.houses, h)
		cl.people += people[h.ID]
		if old := vs.byHouse[h.ID]; old != nil {
			cl.shared[old.ID]++
		}
	}
	return out, cells
}

// surveyLand finds each village's buildings and fields and draws its land
// around them and its houses.
func (s *Sim) surveyLand(cells map[[2]int][]*Structure) {
	vs := s.vil()
	pts := map[*village][]point{}
	for _, v := range vs.list {
		v.Buildings, v.Fields = v.Buildings[:0], 0
		sx, sy := 0.0, 0.0
		for _, id := range v.Houses {
			h := s.structByID[id]
			p := point{float64(h.X) + 0.5, float64(h.Y) + 0.5}
			pts[v] = append(pts[v], p)
			sx, sy = sx+p[0], sy+p[1]
		}
		n := float64(len(v.Houses))
		v.X, v.Y = sx/n, sy/n
	}
	// A building belongs to the village of the nearest house within reach.
	for _, st := range s.structures {
		if st.house() {
			continue
		}
		var near *village
		bestD := 0.0
		k := houseCell(st.X, st.Y)
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				for _, h := range cells[[2]int{k[0] + dx, k[1] + dy}] {
					v := vs.byHouse[h.ID]
					if v == nil {
						continue
					}
					d := math.Hypot(float64(h.X-st.X), float64(h.Y-st.Y))
					if d <= villageReach && (near == nil || d < bestD || d == bestD && v.ID < near.ID) {
						near, bestD = v, d
					}
				}
			}
		}
		if near != nil {
			near.Buildings = append(near.Buildings, st.ID)
			pts[near] = append(pts[near], point{float64(st.X) + 0.5, float64(st.Y) + 0.5})
		}
	}
	// A field belongs to its family's village while it is near their house.
	w := s.terrain.w
	for _, p := range s.eco.Plots() {
		v := vs.byHouse[p.House]
		if v == nil {
			continue
		}
		x, y := int(p.Tile)%w, int(p.Tile)/w
		if h := s.structByID[p.House]; h == nil || math.Hypot(float64(h.X-x), float64(h.Y-y)) > villageReach {
			continue
		}
		v.Fields++
		pts[v] = append(pts[v], point{float64(x) + 0.5, float64(y) + 0.5})
	}
	for _, v := range vs.list {
		v.Hull = landHull(pts[v])
	}
}

// newVillage names a newly recognised village; taken lists the villages
// already made in this pass, besides the old ones.
func (s *Sim) newVillage(taken []*village) *village {
	vs := s.vil()
	vs.lastID++
	vs.founded++
	id := vs.lastID
	used := func(name string) bool {
		for _, l := range [][]*village{vs.list, taken} {
			for _, v := range l {
				if v.Name == name {
					return true
				}
			}
		}
		return false
	}
	name := villageName(id, 0)
	for try := int64(1); used(name) && try < 16; try++ {
		name = villageName(id, try)
	}
	return &village{ID: id, Name: name, Hue: villageHue(id), Founded: s.time()}
}

// Village names: an invented two-syllable root, mostly with one of the
// endings Javanese and Sundanese village names often carry (sari "essence",
// jaya "victory", rejo "prosperous" …) or after a landscape word (karang
// "reef, yard", wana "forest", sumber "spring"). Any likeness to a real
// village is chance.
var (
	villageOnsets   = []string{"b", "d", "g", "j", "k", "l", "m", "n", "p", "r", "s", "t", "w", "c"}
	villageMidOns   = []string{"b", "d", "g", "j", "k", "l", "m", "n", "p", "r", "s", "t", "w", "ng", "nd", "mb"}
	villageVowels   = []string{"a", "a", "i", "u", "e", "o"}
	villageCodas    = []string{"", "", "", "n", "ng", "r"}
	villageSuffixes = []string{"sari", "jaya", "mulya", "rejo", "wangi", "harjo", "makmur", "mukti", "tani", "rahayu", "luhur", "sejati"}
	villagePrefixes = []string{"Karang", "Wana", "Sumber", "Tegal", "Banyu", "Giri", "Kali", "Watu"}
)

// villageName is the name of village id; try picks another one when that
// name is taken.
func villageName(id, try int64) string {
	pick := func(list []string, salt uint64) string {
		return list[min(len(list)-1, int(hash01(id, try, salt)*float64(len(list))))]
	}
	root := pick(villageOnsets, 101) + pick(villageVowels, 102) + pick(villageMidOns, 103) + pick(villageVowels, 104)
	switch form := hash01(id, try, 105); {
	case form < 0.6:
		return capital(root + pick(villageSuffixes, 106))
	case form < 0.85:
		return pick(villagePrefixes, 107) + root + pick(villageCodas, 108)
	default:
		return capital(root + pick(villageCodas, 108))
	}
}

func capital(s string) string { return strings.ToUpper(s[:1]) + s[1:] }

// villageHue spreads the villages' colours by the golden angle.
func villageHue(id int64) int { return int(math.Mod(float64(id)*137.508, 360)) }

// villageWire is a village in the `villages` stream message.
type villageWire struct {
	ID      int64        `json:"id"`
	Name    string       `json:"name"`
	Hue     int          `json:"hue"`
	Hull    [][2]float64 `json:"hull"`
	X       float64      `json:"x"`
	Y       float64      `json:"y"`
	People  int          `json:"people"`
	Houses  int          `json:"houses"`
	Leader  *Ref         `json:"leader"`
	Founded float64      `json:"founded"`
}

func (v *village) leaderRef() *Ref {
	if v.Leader == 0 {
		return nil
	}
	return &Ref{v.Leader, v.LeaderName}
}

func roundHull(h [][2]float64) [][2]float64 {
	out := make([][2]float64, len(h))
	for i, p := range h {
		out[i] = [2]float64{r1(p[0]), r1(p[1])}
	}
	return out
}

// encodeVillages re-encodes the village list and moves the version on if it
// changed.
func (s *Sim) encodeVillages() {
	vs := s.vil()
	wire := make([]villageWire, len(vs.list))
	for i, v := range vs.list {
		wire[i] = villageWire{ID: v.ID, Name: v.Name, Hue: v.Hue, Hull: roundHull(v.Hull), X: r1(v.X), Y: r1(v.Y),
			People: v.People, Houses: len(v.Houses), Leader: v.leaderRef(), Founded: r1(v.Founded)}
	}
	body, err := json.Marshal(wire)
	if err != nil {
		panic(err) // plain numbers and strings
	}
	if !bytes.Equal(body, vs.body) {
		if vs.body != nil || len(wire) > 0 {
			vs.version++
		}
		vs.body = body
	}
}

// Villages returns the encoded `villages` stream message and its version.
func (s *Sim) Villages() ([]byte, int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	vs := s.vil()
	b := make([]byte, 0, 24+len(vs.body))
	b = append(b, `{"v":`...)
	b = strconv.AppendInt(b, vs.version, 10)
	b = append(b, `,"villages":`...)
	if vs.body == nil {
		b = append(b, "[]"...)
	} else {
		b = append(b, vs.body...)
	}
	return append(b, '}'), vs.version
}

// VillageVersion changes whenever a village does.
func (s *Sim) VillageVersion() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.vil().version
}

// VillagesView lists the villages for the API.
func (s *Sim) VillagesView() []VillageView {
	s.mu.Lock()
	defer s.mu.Unlock()
	vs := s.vil()
	members := map[int64][]*Creature{}
	for _, c := range s.creatures {
		if c.VillageID != 0 && c.Health > 0 {
			members[c.VillageID] = append(members[c.VillageID], c)
		}
	}
	out := make([]VillageView, 0, len(vs.list))
	for _, v := range vs.list {
		view := VillageView{ID: v.ID, Name: v.Name, Hue: v.Hue, Hull: roundHull(v.Hull), X: r1(v.X), Y: r1(v.Y),
			People: v.People, Houses: len(v.Houses), Leader: v.leaderRef(), Founded: r1(v.Founded),
			Area: v.area, HouseIDs: slices.Clone(v.Houses), Fields: v.Fields, Buildings: []VillageBuilding{},
			LeaderSince: r1(v.LeaderSince), LeaderTrust: v.Trust, Supporters: v.Supporters, Leaders: v.Leaders,
			Residents: []Villager{}, Attitudes: s.attitudes(v, members[v.ID])}
		for _, id := range v.Buildings {
			if st := s.structByID[id]; st != nil {
				view.Buildings = append(view.Buildings, VillageBuilding{st.ID, st.Kind, st.kind.Name, st.X, st.Y})
			}
		}
		view.HouseFood, view.GranaryFood = s.villageFood(v)
		for _, c := range members[v.ID] {
			r := Villager{ID: c.ID, Name: c.Name, House: c.HouseID, Adult: s.adult(c), Leader: c.ID == v.Leader}
			for _, rel := range c.Relations {
				if rel.Person.ID == v.Leader && v.Leader != 0 {
					r.Trust = r3(s.relationTrust(rel))
				}
			}
			view.Residents = append(view.Residents, r)
		}
		out = append(out, view)
	}
	return out
}

// villageRef names c's village, nil if none.
func (s *Sim) villageRef(c *Creature) *Ref {
	if c.VillageID == 0 {
		return nil
	}
	if v := s.vil().byID[c.VillageID]; v != nil {
		return &Ref{v.ID, v.Name}
	}
	return nil
}

func (s *Sim) saveVillages() *villageState {
	vs := s.vil()
	if vs.lastID == 0 && vs.version == 0 {
		return nil // never had a village
	}
	return &villageState{List: vs.list, LastID: vs.lastID, Version: vs.version, Founded: vs.founded,
		Abandoned: vs.abandoned, LeaderChanges: vs.leaderChanges}
}

func (s *Sim) restoreVillages(st *villageState) {
	vs := s.vil()
	*vs = villageSys{}
	if st != nil {
		vs.list = st.List
		vs.lastID, vs.version = st.LastID, st.Version
		vs.founded, vs.abandoned, vs.leaderChanges = st.Founded, st.Abandoned, st.LeaderChanges
	}
	s.indexVillages()
	s.paintLand()
	for _, c := range s.creatures {
		if v := vs.byID[c.VillageID]; v != nil {
			v.members = append(v.members, c)
		}
	}
	if len(vs.list) > 0 || vs.version > 0 {
		s.encodeVillages()
		vs.version = st.Version // the list is what it was when saved
	}
}

func (s *Sim) villageInfo() VillageInfo {
	vs := s.vil()
	info := VillageInfo{Villages: len(vs.list), Founded: vs.founded, Abandoned: vs.abandoned, LeaderChanges: vs.leaderChanges}
	for _, v := range vs.list {
		info.Villagers += v.People
		info.Largest = max(info.Largest, v.People)
		if v.Leader != 0 {
			info.Led++
		}
	}
	return info
}
