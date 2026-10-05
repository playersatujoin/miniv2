// Package chem is the material world: all 118 chemical elements, the items
// creatures gather and craft, where resources lie on a map, the technology
// tree, buildable structures and the recipes that turn resources into new
// resources. It holds data and pure helpers only; the simulation owns state.
package chem

// Tiers of civilisation. A world reaches tier N once a station of tier N
// exists in it.
const MaxTier = 7

var tierNames = [MaxTier + 1]string{
	"Zaman Batu",
	"Zaman Logam",
	"Zaman Kimia",
	"Zaman Listrik",
	"Zaman Spektroskopi",
	"Zaman Radiasi",
	"Zaman Nuklir",
	"Zaman Partikel",
}

// TierName returns the Indonesian name of a tier, e.g. 0 → "Zaman Batu".
func TierName(tier int) string {
	if tier < 0 || tier > MaxTier {
		return ""
	}
	return tierNames[tier]
}

// Element is one of the 118 chemical elements.
type Element struct {
	Z         int     `json:"z"`
	Symbol    string  `json:"symbol"`
	Name      string  `json:"name"`      // Indonesian, e.g. "Besi"
	Category  string  `json:"category"`  // Indonesian, e.g. "logam transisi"
	Period    int     `json:"period"`    // 1–7
	Group     int     `json:"group"`     // 1–18; 0 for lanthanides (57–71) and actinides (89–103)
	Phase     string  `json:"phase"`     // "padat" | "cair" | "gas" at room temperature
	Abundance float64 `json:"abundance"` // crustal abundance in ppm; 0 for synthetic elements
	Natural   bool    `json:"natural"`
	Tier      int     `json:"tier"` // tier needed to isolate it (0 = found by gathering)
}

type ItemID string

// Food is the edible item creatures carry, store and share.
const Food ItemID = "makanan"

// SynthesisFuel is consumed by each synthesis of a synthetic element.
const SynthesisFuel ItemID = "uranium"

type Item struct {
	ID       ItemID   `json:"id"`
	Name     string   `json:"name"`              // Indonesian
	Kind     string   `json:"kind"`              // "makanan" | "bahan" | "mineral" | "olahan" | "logam" | "alat" | "senjata"
	Formula  string   `json:"formula,omitempty"` // e.g. "Fe2O3"
	Elements []string `json:"elements"`          // symbols of the elements it contains
	Food     float64  `json:"food,omitempty"`    // energy per unit eaten; 0 = inedible
	Value    float64  `json:"value"`             // worth; thieves prefer valuable items
	Damage   float64  `json:"damage,omitempty"`  // attack bonus while carried (weapons)
	Gather   float64  `json:"gather,omitempty"`  // gathering speed multiplier while carried (tools), 0 = none
}

// Source is something gatherable from the tile (X, Y), seen from a creature's tile.
type Source struct {
	Item      ItemID
	X, Y      int
	Amount    float64 // what is left right now
	NeedsTool bool    // only gatherable while carrying an item with Gather > 0
	Model     string  // deposit model key, see DepositModels
}

type Tech struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tier        int      `json:"tier"`
	Requires    []string `json:"requires"`
}

type StructureKind struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Cost     map[ItemID]int `json:"cost"`
	Tech     string         `json:"tech,omitempty"`     // tech that must be known to build it
	Teaches  string         `json:"teaches,omitempty"`  // tech learned the first time it is built
	House    bool           `json:"house"`              // homes a family and stores items
	Level    int            `json:"level"`              // house level 1–3, 0 otherwise
	Upgrades string         `json:"upgrades,omitempty"` // the house kind this one replaces
	Storage  int            `json:"storage"`            // item units it can store (houses)
	Tier     int            `json:"tier"`               // station tier (1–7), 0 if not a station
	Farm     bool           `json:"farm,omitempty"`     // produces food over time
	Well     bool           `json:"well,omitempty"`     // creatures nearby can drink
	Library  bool           `json:"library,omitempty"`  // stores written know-how that outlives its writers
}

type Recipe struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Inputs    map[ItemID]int `json:"inputs"`
	Outputs   map[ItemID]int `json:"outputs"`
	Station   string         `json:"station,omitempty"`   // structure kind needed within reach
	Tech      string         `json:"tech,omitempty"`      // tech that must be known
	Teaches   string         `json:"teaches,omitempty"`   // tech learned the first time it is made
	Discovers []string       `json:"discovers,omitempty"` // elements discovered the first time it is made
	Seconds   float64        `json:"seconds"`             // work time
}
