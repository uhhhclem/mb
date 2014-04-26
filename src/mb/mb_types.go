package mb

import (
	"fmt"
)

var (
	tribeNames     []string
	tradeGoodNames []string
)

type Tribe int

func (t Tribe) String() string {
	return tribeNames[int(t)]
}

const (
	HoChunk Tribe = iota
	Shawnee
	Cherokee
	Natchez
	Caddo
	SpanishTribe   // only for HistoryCard and targeting an enemy
	All            // only for HistoryCard/WarpathStatus, never for Land
	None           // only for HistoryCard, never for Land
	CaddoOrShawnee // only for HistoryCard 37
)

var tribeNameLookup = map[string]Tribe{
	"HoChunk":        HoChunk,
	"Shawnee":        Shawnee,
	"Cherokee":       Cherokee,
	"Natchez":        Natchez,
	"Caddo":          Caddo,
	"Spanish":        SpanishTribe,
	"All":            All,
	"None":           None,
	"CaddoOrShawnee": CaddoOrShawnee,
}

// Land represents a space on the board
type Land struct {
	Warpath      Tribe
	Name         string
	Space        int
    Index 		 int
	IsWilderness bool
	IsControlled bool
}

func (l Land) String() string {
	return fmt.Sprintf("%d %s: %s (%t)", l.Space, l.Warpath, l.Name, l.IsWilderness)
}

type WarpathStatus struct {
	Warpath Tribe
	Modifier     int
}

func (w WarpathStatus) String() string {
	switch w.Modifier {
	case 1:
		return w.Warpath.String() + " + 1"
	case -1:
		return w.Warpath.String() + " - 1"
	}
	return "None"
}

// TradeGood represents the nine trade goods.
type TradeGood int

func (t TradeGood) String() string {
	return tradeGoodNames[int(t)]
}

const (
	Hides TradeGood = iota
	Chert
	Feathers
	Copper
	Mica
	Chalcedony
	Pipestone
	Obsidian
	Seashells
)

var tradeGoodNameLookup = map[string]TradeGood{
	"Hides":      Hides,
	"Chert":      Chert,
	"Mica":       Mica,
	"Copper":     Copper,
	"Chalcedony": Chalcedony,
	"Pipestone":  Pipestone,
	"Feathers":   Feathers,
	"Seashells":  Seashells,
	"Obsidian":   Obsidian,
}

// ChiefdomCounter represents a single chiefdom counter.
type ChiefdomCounter struct {
	Good    TradeGood
	Plain   ChiefdomCounterFace
	Mounded ChiefdomCounterFace
}

func (c *ChiefdomCounter) String() string {
	return fmt.Sprintf("%s (%d %t/%d %t)", c.Good, c.Plain.Value, c.Plain.IsGreenBird, c.Mounded.Value, c.Mounded.IsGreenBird)
}

// Cup is the cup full of chiefdom counters.
type Cup []*ChiefdomCounter

// ChiefdomCounterFace represents one side of a chiefdom counter.
type ChiefdomCounterFace struct {
	Value       int
	IsGreenBird bool
}

// Chiefdom represents a chiefdom on the board.
type Chiefdom struct {
	Counter      *ChiefdomCounter
	IsMounded    bool
	IsControlled bool
	LandIndex    int // index into Board.Lands
}

func (c Chiefdom) String() string {
	f := c.getCounterFace()
	side := "Plain"
	if c.IsMounded {
		side = "Mounded"
	}
	icon := "R"
	if f.IsGreenBird {
		icon = "B"
	}
	return fmt.Sprintf("%s (%d) - %s %s", c.Counter.Good, f.Value, side, icon)
}

func (c Chiefdom) getCounterFace() ChiefdomCounterFace {
	if c.IsMounded {
		return c.Counter.Mounded
	}
	return c.Counter.Plain
}

func (c Chiefdom) getValue() int {
	return c.getCounterFace().Value
}

func (c Chiefdom) IsGreenBirdman() bool {
	return c.getCounterFace().IsGreenBird
}

// HostileMarker represents one of the six hostile markers (including the Spanish)
type HostileMarker struct {
	LandIndex   int // index into Board.Lands
	BattleValue int
	IsSpanish   bool // if not, Space.Warpath tells you the tribe
	Dice        int  // only if Spanish
}

func (h HostileMarker) String() string {
	t, i := fromLandIndex(h.LandIndex)
	return fmt.Sprintf("%s (%d) on %d", t, h.BattleValue, i)
}

type Era int

const (
	Hopewell Era = iota
	Mississippian
	Spanish
	Generic // used only for HistoryCards, never for the Board
)

var eraNameLookup = map[string]Era{
	"HOPEWELL":      Hopewell,
	"MISSISSIPPIAN": Mississippian,
	"SPANISH":       Spanish,
	"GENERIC":       Generic,
}

type Palisade struct {
	Label string
	Value int
}

type Board struct {
	CurrentEra    Era
	Card          *HistoryCard
	ActionPoints  int
	TradeGoods    int
	PalisadeIndex int
	Palisades     []Palisade
	IsBreached    bool
	Lands         []Land
	Chiefdoms     []*Chiefdom
	Hostiles      []*HostileMarker
	PeacePipes    []bool
	WarpathStatus WarpathStatus
}

type HistoryCard struct {
	Number          int
	Title           string
	Era             Era
	ActionPoints    int
	IsWhite         bool
	ResourceBonus   []TradeGood
	Revolt          Tribe
	AdvancingArmies []Tribe
	Modifier        Tribe // may be None or All
	IsAscendant     bool  // indicates if tribe's Modifier is positive or negative
	IsAvaricia      bool
	IsSpanish       bool
}

func (h *HistoryCard) String() string {
	return fmt.Sprintf("%d: %s", h.Number, h.Title)
}

type Pile []*HistoryCard

type Prompt string

type Input string

func init() {
	tradeGoodNames = make([]string, len(tradeGoodNameLookup))
	for k, v := range tradeGoodNameLookup {
		tradeGoodNames[int(v)] = k
	}

	tribeNames = make([]string, len(tribeNameLookup))
	for k, v := range tribeNameLookup {
		tribeNames[int(v)] = k
	}
}
