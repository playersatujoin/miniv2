package sim

import (
	"fmt"
	"math"
	"slices"

	"miniv2/backend/internal/chem"
)

// Culture: know-how lives in people. Practising a technology needs a skill of
// at least skillPractise in it. Skills grow by practice, by being taught, by
// watching someone work and by reading at a library; they fade slowly when
// unused. When the last person who can practise something dies and nothing
// was written down, that knowledge is lost until someone re-invents it.
// With Options.NoLearning the world falls back to the old model: whatever
// anyone ever discovered, everybody can do.

const (
	skillPractise = 0.3        // needed to practise a technology
	discoverSkill = 0.5        // an inventor's own skill
	buildSkill    = 0.4        // putting up a station or library teaches how it works
	practiceGain  = 0.08       // per finished job, × (1 - level)
	skillForget   = 1.0 / 2000 // procedural know-how fades slowly: half-life ≈ 170 years unused
	teachRange    = 2.0
	teachRate     = 0.1 // per second × the teacher's lead
	minLead       = 0.1 // a teacher must know at least this much more
	watchRange    = 3.0
	watchRate     = 0.02 // per second × the worker's lead
	homeLearnRate = 0.01 // per second × lead: growing up among family who know it
	libraryReach  = 2.0
	writeRate     = 0.04 // per second, towards the writer's level
	readRate      = 0.01 // per second × what the text holds beyond the reader
	literacyTech  = "tulisan"
	teacherMemory = 3 * TicksPerSecond
	learnEventGap = 10.0
	maxStudents   = 64 // distinct students remembered per teacher
)

func (s *Sim) skill(c *Creature, tech string) float64 { return c.Skills[tech] }

// canPractise reports whether c can use a technology ("" needs none).
func (s *Sim) canPractise(c *Creature, tech string) bool {
	if tech == "" {
		return true
	}
	if s.noCulture() {
		return s.techKnown(tech)
	}
	return c.Skills[tech] >= skillPractise
}

func (s *Sim) bestSkill(c *Creature) float64 {
	best := 0.0
	for _, v := range c.Skills {
		best = math.Max(best, v)
	}
	return best
}

func (s *Sim) techName(id string) string {
	for _, t := range s.cat.techs {
		if t.ID == id {
			return t.Name
		}
	}
	return id
}

// stationTech is the know-how needed to work a station.
func stationTech(k chem.StructureKind) string {
	if k.Teaches != "" {
		return k.Teaches
	}
	return k.Tech
}

// raise sets c's skill in tech to level if that is higher, noting when c
// becomes able to practise it. how is "teach", "read", "watch" or "practice";
// from is the teacher or library.
func (s *Sim) raise(c *Creature, tech string, level float64, how string, from *Ref) {
	if s.noCulture() || tech == "" {
		return
	}
	old := c.Skills[tech]
	level = math.Min(1, level)
	if level <= old {
		return
	}
	if c.Skills == nil {
		c.Skills = map[string]float64{}
	}
	c.Skills[tech] = level
	if old >= skillPractise || level < skillPractise {
		return
	}
	if s.learnedVia == nil {
		s.learnedVia = map[string]int{}
	}
	s.learnedVia[how]++
	name := s.techName(tech)
	if s.lost[tech] {
		delete(s.lost, tech)
		s.event("learning", fmt.Sprintf("Pengetahuan ditemukan kembali: %s — %s", name, c.Name), c.ID)
		return
	}
	if from == nil || s.time()-s.lastLearnEvent < learnEventGap {
		return
	}
	s.lastLearnEvent = s.time()
	switch how {
	case "teach":
		s.event("learning", fmt.Sprintf("%s belajar %s dari %s", c.Name, name, from.Name), c.ID)
	case "read":
		s.event("learning", fmt.Sprintf("%s belajar %s dari %s", c.Name, name, from.Name), c.ID)
	case "watch":
		s.event("learning", fmt.Sprintf("%s belajar %s dengan memperhatikan %s", c.Name, name, from.Name), c.ID)
	case "home":
		s.event("learning", fmt.Sprintf("%s belajar %s dari keluarganya, %s", c.Name, name, from.Name), c.ID)
	}
}

// novel reports whether tech is worth inventing: nobody in the world knows it,
// or everybody who did is gone. People learn known crafts from each other
// rather than inventing them again.
func (s *Sim) novel(tech string) bool {
	if tech == "" {
		return false
	}
	return !s.techKnown(tech) || !s.noCulture() && s.lost[tech]
}

// practise strengthens a skill that was just used.
func (s *Sim) practise(c *Creature, tech string) {
	if tech == "" {
		return
	}
	l := c.Skills[tech]
	s.raise(c, tech, l+practiceGain*(1-l), "practice", nil)
}

// learn records a technology the world didn't know yet (its inventor starts
// fairly skilled) or, if it is known, counts as practice towards it. floor is
// the least skill the act leaves behind (building a station teaches more than
// one try at a recipe).
func (s *Sim) learn(c *Creature, tech string) { s.learnAt(c, tech, 0) }

func (s *Sim) learnAt(c *Creature, tech string, floor float64) {
	if tech == "" {
		return
	}
	if !s.techKnown(tech) {
		s.techs[tech] = &Discovery{Time: s.time(), By: Ref{c.ID, c.Name}, Era: s.era}
		c.Deeds.Discoveries++
		s.event("discovery", fmt.Sprintf("Teknologi baru: %s — %s", s.techName(tech), c.Name), c.ID)
		s.raise(c, tech, math.Max(floor, discoverSkill), "invent", nil)
		return
	}
	l := c.Skills[tech]
	s.raise(c, tech, math.Max(floor, l+practiceGain*(1-l)), "practice", nil)
}

// --- teaching and writing -------------------------------------------------------

// teach passes on what c knows: to the nearby person it can teach the most,
// or, with nobody to teach, into a library within reach if c can write.
func (s *Sim) teach(c *Creature) bool {
	if s.noCulture() || len(c.Skills) == 0 {
		return false
	}
	var student *Creature
	tech, lead := "", minLead
	s.grid.near(c.X, c.Y, teachRange, func(o *Creature) {
		if o == c || o.Health <= 0 || math.Hypot(o.X-c.X, o.Y-c.Y) > teachRange {
			return
		}
		for _, t := range s.cat.techs {
			if l := c.Skills[t.ID]; l >= skillPractise {
				if d := l - o.Skills[t.ID]; d > lead || d == lead && student != nil && o.ID < student.ID {
					student, tech, lead = o, t.ID, d
				}
			}
		}
	})
	if student != nil {
		from := &Ref{c.ID, c.Name}
		student.Teacher, student.TeacherUntil = from, s.tick+teacherMemory
		s.raise(student, tech, student.Skills[tech]+teachRate*lead*dt, "teach", from)
		if !slices.Contains(c.Students, student.ID) && len(c.Students) < maxStudents {
			c.Students = append(c.Students, student.ID)
		}
		s.flash(c, fxTeach)
		c.action = ActTeach
		return true
	}
	return s.write(c)
}

func (s *Sim) libraryNear(x, y, r float64) *Structure {
	var best *Structure
	bestD := r
	for _, st := range s.structures {
		if st.kind.Library {
			if d := st.dist(x, y); d <= bestD {
				best, bestD = st, d
			}
		}
	}
	return best
}

// write copies c's know-how into a nearby library, skill by skill.
func (s *Sim) write(c *Creature) bool {
	if c.Skills[literacyTech] < skillPractise {
		return false
	}
	lib := s.libraryNear(c.X, c.Y, libraryReach)
	if lib == nil {
		return false
	}
	for _, t := range s.cat.techs {
		l := c.Skills[t.ID]
		if t.ID == literacyTech || l < skillPractise || lib.Written[t.ID] >= l-0.05 {
			continue
		}
		if lib.Written == nil {
			lib.Written = map[string]float64{}
		}
		old := lib.Written[t.ID]
		lib.Written[t.ID] = old + (l-old)*writeRate*dt
		if old < skillPractise && lib.Written[t.ID] >= skillPractise {
			s.event("learning", fmt.Sprintf("%s menuliskan %s di %s", c.Name, t.Name, lib.kind.Name), c.ID)
		}
		s.flash(c, fxTeach)
		c.action = ActTeach
		return true
	}
	return false
}

// writtenLevel is the best written account of tech in any standing library.
func (s *Sim) writtenLevel(tech string) float64 {
	best := 0.0
	for _, st := range s.structures {
		if st.kind.Library {
			best = math.Max(best, st.Written[tech])
		}
	}
	return best
}

// --- once a second ---------------------------------------------------------------

// cultureTick fades unused skills, lets people learn by watching and reading,
// and notices knowledge that died out.
func (s *Sim) cultureTick() {
	if s.noCulture() {
		return
	}
	for _, c := range s.creatures {
		for t, l := range c.Skills {
			if l *= 1 - skillForget; l < 0.001 {
				delete(c.Skills, t)
			} else {
				c.Skills[t] = l
			}
		}
		if c.Teacher != nil && s.tick > c.TeacherUntil {
			c.Teacher = nil
		}
	}
	for _, w := range s.creatures {
		if w.Job != nil && w.Health > 0 {
			s.beWatched(w)
		}
	}
	for _, c := range s.creatures {
		if c.Health > 0 {
			s.learnAtHome(c)
		}
	}
	for _, c := range s.creatures {
		if c.Job == nil && c.Health > 0 && c.Skills[literacyTech] >= skillPractise {
			s.read(c)
		}
	}
	s.countHolders()
}

// jobTechs are the technologies a piece of work shows to onlookers.
func (s *Sim) jobTechs(j *Job) []string {
	switch j.Kind {
	case "craft":
		if r, ok := s.recipeByID(j.Recipe); ok {
			return []string{r.Tech, r.Teaches}
		}
	case "build", "upgrade":
		if k, ok := s.cat.structure[j.Structure]; ok {
			return []string{k.Tech, k.Teaches}
		}
	}
	return nil
}

// beWatched lets everyone near a worker pick up a little of the craft.
func (s *Sim) beWatched(w *Creature) {
	techs := s.jobTechs(w.Job)
	if len(techs) == 0 {
		return
	}
	from := &Ref{w.ID, w.Name}
	s.grid.near(w.X, w.Y, watchRange, func(o *Creature) {
		if o == w || o.Health <= 0 || math.Hypot(o.X-w.X, o.Y-w.Y) > watchRange {
			return
		}
		for _, t := range techs {
			if t == "" {
				continue
			}
			if lead := w.Skills[t] - o.Skills[t]; lead > 0 {
				s.raise(o, t, o.Skills[t]+watchRate*lead, "watch", from)
			}
		}
	})
}

// learnAtHome is enculturation: living close to family who can do something,
// one slowly picks it up, the way children learn the everyday skills of
// their household. Only kin within watchRange count, and only what they can
// practise.
func (s *Sim) learnAtHome(c *Creature) {
	s.grid.near(c.X, c.Y, watchRange, func(o *Creature) {
		if o == c || o.Health <= 0 || len(o.Skills) == 0 || !s.kin(c, o) || math.Hypot(o.X-c.X, o.Y-c.Y) > watchRange {
			return
		}
		for _, t := range s.cat.techs {
			theirs := o.Skills[t.ID]
			if lead := theirs - c.Skills[t.ID]; theirs >= skillPractise && lead > 0 {
				s.raise(c, t.ID, c.Skills[t.ID]+homeLearnRate*lead, "home", &Ref{o.ID, o.Name})
			}
		}
	})
}

// read lets a literate creature at a library learn what is written there.
func (s *Sim) read(c *Creature) {
	lib := s.libraryNear(c.X, c.Y, libraryReach)
	if lib == nil || len(lib.Written) == 0 {
		return
	}
	from := &Ref{lib.Owner, lib.kind.Name}
	if lib.Owner != 0 {
		from.Name += " " + lib.OwnerName
	}
	for _, t := range s.cat.techs {
		w := lib.Written[t.ID]
		if own := c.Skills[t.ID]; w > own+0.05 {
			s.raise(c, t.ID, own+readRate*(w-own), "read", from)
			c.Teacher, c.TeacherUntil = from, s.tick+teacherMemory
		}
	}
}

// countHolders counts who can practise each technology and marks what was
// known but now has no holder and no written record.
func (s *Sim) countHolders() {
	holders := map[string]int{}
	best := map[string]*Creature{}
	for _, c := range s.creatures {
		for t, l := range c.Skills {
			if l < skillPractise {
				continue
			}
			holders[t]++
			if b := best[t]; b == nil || l > b.Skills[t] || l == b.Skills[t] && c.ID < b.ID {
				best[t] = c
			}
		}
	}
	s.holders = holders
	for _, t := range s.cat.techs {
		if s.techs[t.ID] == nil {
			continue
		}
		if b := best[t.ID]; b != nil {
			s.lastHolder[t.ID] = Ref{b.ID, b.Name}
			continue
		}
		if s.lost[t.ID] || s.writtenLevel(t.ID) >= skillPractise {
			continue
		}
		s.lost[t.ID] = true
		s.knowledgeLost++
		text := "Pengetahuan hilang: " + t.Name
		if lh, ok := s.lastHolder[t.ID]; ok {
			if s.byID[lh.ID] == nil {
				text += fmt.Sprintf(" — pemegang terakhir, %s, meninggal", lh.Name)
			} else {
				text += fmt.Sprintf(" — %s sudah melupakannya", lh.Name)
			}
		}
		s.event("learning", text, 0)
	}
}

// LearningChannels counts how people became able to practise something:
// invent, practice, teach, home, watch or read.
func (s *Sim) LearningChannels() map[string]int {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := map[string]int{}
	for k, v := range s.learnedVia {
		out[k] = v
	}
	return out
}

// holderCount is how many living people can practise tech.
func (s *Sim) holderCount(tech string) int {
	if s.noCulture() {
		if s.techKnown(tech) {
			return len(s.creatures)
		}
		return 0
	}
	return s.holders[tech]
}

// teacherOrStudentNear senses whether someone nearby knows clearly more
// (a possible teacher) or clearly less (a possible student) than c.
func (s *Sim) teacherOrStudentNear(c, o *Creature) (teacher, student bool) {
	if len(c.Skills) == 0 && len(o.Skills) == 0 {
		return false, false
	}
	for _, t := range s.cat.techs {
		mine, theirs := c.Skills[t.ID], o.Skills[t.ID]
		if theirs >= skillPractise && theirs-mine > minLead {
			teacher = true
		}
		if mine >= skillPractise && mine-theirs > minLead {
			student = true
		}
	}
	return teacher, student
}

// SkillView is one learned skill for the inspector.
type SkillView struct {
	Tech  string  `json:"tech"`
	Name  string  `json:"name"`
	Level float64 `json:"level"`
}

func (s *Sim) skillViews(c *Creature) []SkillView {
	out := []SkillView{}
	for _, t := range s.cat.techs {
		if l := c.Skills[t.ID]; l > 0 {
			out = append(out, SkillView{Tech: t.ID, Name: t.Name, Level: r3(l)})
		}
	}
	slices.SortStableFunc(out, func(a, b SkillView) int {
		switch {
		case a.Level > b.Level:
			return -1
		case a.Level < b.Level:
			return 1
		}
		return 0
	})
	return out
}

// recountHolders refreshes who can practise what without reporting losses
// (used after loading a world).
func (s *Sim) recountHolders() {
	s.holders = map[string]int{}
	for _, c := range s.creatures {
		for t, l := range c.Skills {
			if l >= skillPractise {
				s.holders[t]++
			}
		}
	}
}

// AgeSkill is the median best skill of the living in one age band.
type AgeSkill struct {
	Label  string   `json:"label"`
	People int      `json:"people"`
	Median *float64 `json:"median"`
}

var skillBands = []struct {
	label    string
	from, to float64
}{{"0–14", 0, 15}, {"15–29", 15, 30}, {"30–44", 30, 45}, {"45–59", 45, 60}, {"60+", 60, math.Inf(1)}}

// SkillProfile is how skilled people are by age: a culture that learns
// should show skill rising with age.
func (s *Sim) SkillProfile() []AgeSkill {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]AgeSkill, len(skillBands))
	for i, b := range skillBands {
		var vs []float64
		for _, c := range s.creatures {
			if a := s.ageYears(c); a >= b.from && a < b.to {
				vs = append(vs, s.bestSkill(c))
			}
		}
		out[i] = AgeSkill{Label: b.label, People: len(vs)}
		if len(vs) > 0 {
			slices.Sort(vs)
			m := vs[len(vs)/2]
			if len(vs)%2 == 0 {
				m = (vs[len(vs)/2-1] + m) / 2
			}
			m = r3(m)
			out[i].Median = &m
		}
	}
	return out
}
