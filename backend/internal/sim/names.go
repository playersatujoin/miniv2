package sim

import (
	"math/rand/v2"
	"strings"
)

var (
	femaleStarts = []string{"Sa", "Ri", "Ma", "De", "A", "Ni", "La", "Sri", "Ra", "Ti", "Wu", "Lan", "Ka", "Pu", "Mi", "In", "Yu", "Na", "Fi", "Ce", "Me", "Ros"}
	femaleMids   = []string{"", "", "ri", "ni", "la", "ra", "wa", "ti", "ya", "sa", "li", "ma", "lu"}
	femaleEnds   = []string{"a", "i", "wati", "ni", "yu", "ti", "na", "ri", "sih", "ra", "ka"}
	maleStarts   = []string{"Bu", "A", "Jo", "Ra", "Wa", "Su", "Har", "Bam", "Gun", "Dar", "Yo", "Ba", "Ko", "Te", "Gi", "Ha", "Pra", "Wi", "Dim", "Fa"}
	maleMids     = []string{"", "", "di", "ma", "ri", "no", "to", "ga", "yo", "ra", "ju"}
	maleEnds     = []string{"di", "to", "man", "wan", "no", "dan", "jo", "ko", "ran", "tok", "yo", "gus", "as", "ang", "et"}
)

// newName makes an Indonesian-sounding ASCII name; female names mostly end in
// a vowel, male names more often in a consonant.
func newName(rng *rand.Rand, sex Sex) string {
	starts, mids, ends := femaleStarts, femaleMids, femaleEnds
	if sex == Male {
		starts, mids, ends = maleStarts, maleMids, maleEnds
	}
	pick := func(s []string) string { return s[rng.IntN(len(s))] }
	var name string
	for range 10 {
		name = pick(starts) + pick(mids) + pick(ends)
		if len(name) >= 3 && len(name) <= 10 && !strings.Contains(name, "aa") && !strings.Contains(name, "ii") {
			break
		}
	}
	return name
}
