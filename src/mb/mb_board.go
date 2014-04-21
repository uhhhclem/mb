package mb

var boardData = `Tribe,Name,IsWilderness
HoChunk,Ho-Chunk Homeland,
HoChunk,Red Wing,TRUE
HoChunk,Lizard Mound,
HoChunk,Adtalan,
HoChunk,Lake Koshkonong,
HoChunk,Dickson,
Shawnee,Shawnee Homeland,
Shawnee,Newark,
Shawnee,Serpent Mound,TRUE
Shawnee,Portsmouth,
Shawnee,Fort Ancient,
Shawnee,Angel,
Cherokee,Cherokee Homeland,
Cherokee,Anhaica,
Cherokee,Ocmulgee,
Cherokee,Etowah,
Cherokee,Coosa,TRUE
Cherokee,Pinson,
Natchez,Natchez Homeland,
Natchez,Bottle Creek,
Natchez,Moundville,
Natchez,Bynum,TRUE
Natchez,Chucalissa,
Natchez,Kincaid,
Caddo,Caddo Homeland,
Caddo,Harlan,
Caddo,Spiro,
Caddo,Marksville,
Caddo,Poverty Point,TRUE
Caddo,Toltec,`

const (
	colTribe = iota
	colName
	colIsWilderness
)

const LandCount = 30

func toLandIndex(t Tribe, n int) int {
	return 6*int(t) + (n - 1)
}

func fromLandIndex(i int) (Tribe, int) {
	var t Tribe
	t = Tribe(i / 6)
	n := i%6 + 1
	return t, n
}

func (b Board) findHostile(t Tribe) *HostileMarker {
	for i := 0; i < len(b.Hostiles); i++ {
		h := b.Hostiles[i]
		if h != nil && t == Tribe(i/6) {
			return h
		}
	}
	return nil
}

func (b Board) findLand(t Tribe, n int) Land {
	return b.Lands[toLandIndex(t, n)]
}

func (b Board) findChiefdom(t Tribe, n int) *Chiefdom {
	return b.Chiefdoms[toLandIndex(t, n)]
}

// moveHostile moves a hostile marker one space towards or away from Cahokia.
// It is assumed that the destination space is legal, e.g. this will move
// a hostile into Cahokia whether or not the palisade is intact.
func (b *Board) moveHostile(h *HostileMarker, dir int) {
	switch {
	case dir > 0:
		dir = 1
	case dir < 0:
		dir = -1
	case dir == 0:
		return
	}
	_, i := fromLandIndex(h.LandIndex)
	if dir > 0 && i >= 6 {
		return
	}
	if dir < 0 && i <= 1 {
		return
	}
	b.Hostiles[h.LandIndex] = nil
	h.LandIndex += dir
	b.Hostiles[h.LandIndex] = h
}

// makeBoard makes a new game board.
func makeBoard() Board {
	board := Board{
		Lands:      make([]Land, LandCount),
		Chiefdoms:  make([]*Chiefdom, LandCount),
		Hostiles:   make([]*HostileMarker, LandCount),
		PeacePipes: make([]*PeacePipeMarker, LandCount),
	}
	for i, r := range parse(boardData) {
		if i == 0 {
			// skip the header row
			continue
		}
		l := Land{
			Warpath:      tribeNameLookup[r[colTribe]],
			Name:         r[colName],
			Space:        6 - ((i - 1) % 6),
			IsWilderness: r[colIsWilderness] == "TRUE",
		}
		idx := toLandIndex(l.Warpath, l.Space)
		board.Lands[idx] = l
	}

	// position the hostiles
	vals := []int{4, 3, 2, 2, 3}
	for t := HoChunk; t <= Caddo; t++ {
		i := toLandIndex(t, 6)
		board.Hostiles[i] = &HostileMarker{
			LandIndex:   i,
			BattleValue: vals[int(t)],
		}
	}

	// initialize the palisade
	pv := []struct {
		label string
		value int
	}{
		{"4F", 4},
		{"4E", 4},
		{"4D", 4},
		{"3C", 3},
		{"3B", 3},
		{"2A", 2},
	}
	board.Palisades = make([]Palisade, 6)
	for i, p := range pv {
		board.Palisades[i] = Palisade{Label: p.label, Value: p.value}
	}

	return board
}
