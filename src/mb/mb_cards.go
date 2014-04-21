package mb

import (
	"fmt"
	"strconv"
	"strings"
)

var historyCardData = `Number,Title,Era,ActionPoints,IsWhite,ResourceBonus,Revolt,Modifier,IsAscendant,AdvancingArmies,IsAvaricia,IsSpanish
1,Poverty Point,Hopewell,1,TRUE,,Caddo,Caddo,,,,
2,Bynum,Hopewell,3,TRUE,,Natchez,Natchez,,,,
3,Marksville,Hopewell,4,TRUE,,Caddo,Caddo,,,,
4,Portsmouth,Hopewell,3,TRUE,,Shawnee,Shawnee,,,,
5,Pinson,Hopewell,2,TRUE,,Cherokee,Cherokee,,,,
6,Newark,Hopewell,4,TRUE,,Natchez,Shawnee,,,,
7,Lizard Mound,Hopewell,3,TRUE,,HoChunk,HoChunk,TRUE,,,
8,Toltec,Hopewell,2,TRUE,,Caddo,Caddo,,,,
9,Lake Koshkonong,Hopewell,4,TRUE,,HoChunk,HoChunk,TRUE,,,
10,Harlan,Hopewell,4,TRUE,,Shawnee,Caddo,,,,
11,Dickson,Hopewell,2,TRUE,,HoChunk,HoChunk,TRUE,,,
12,Spiro,Hopewell,3,TRUE,,Cherokee,Caddo,,,,
13,Aztalan,Mississippian,4,TRUE,,,HoChunk,,Cherokee,,
14,Ocmulgee,Mississippian,3,TRUE,,,Cherokee,TRUE,"Caddo,Cherokee",,
15,Fort Ancient,Mississippian,2,TRUE,,,Shawnee,TRUE,Shawnee,,
16,Red Wing,Mississippian,3,TRUE,,,HoChunk,,"Cherokee,Natchez",,
17,Anhaica,Mississippian,4,TRUE,,HoChunk,Cherokee,,Caddo,,
18,Etowah,Mississippian,2,TRUE,,Natchez,Cherokee,TRUE,Cherokee,,
19,Moundville,Mississippian,4,TRUE,,Cherokee,Natchez,TRUE,"HoChunk,Natchez",,
20,Chucalissa,Mississippian,2,TRUE,,,Natchez,,Caddo,,
21,Angel,Mississippian,4,TRUE,,,Shawnee,,"HoChunk,Cherokee",,
22,Kincaid,Mississippian,1,TRUE,,,Natchez,TRUE,"Natchez,Shawnee",,
23,Serpent Mound,Mississippian,3,TRUE,,Caddo,Shawnee,TRUE,"Natchez,Shawnee",,
24,Bottle Creek,Mississippian,3,TRUE,,,Natchez,,"Caddo,Cherokee",,
25,Coosa,Spanish,4,TRUE,,,Cherokee,TRUE,"Cherokee,Shawnee,Natchez",TRUE,
26,The Spanish,Spanish,0,TRUE,,,,,"HoChunk,Shawnee,Spanish",,TRUE
27,Chalcedony & Obsidian,Generic,4,,"Chalcedony,Obsidian",,,,"Caddo,Caddo,Natchez",,
28,Pipestone,Generic,4,,Pipestone,,,,"HoChunk,Shawnee",,
29,Mica & Seashells,Generic,5,,"Mica,Seashells",,,,"Natchez,Cherokee,Cherokee",,
30,Hides & Feathers,Generic,6,,"Hides,Feathers",,,,"Cherokee,Caddo",,
31,Chert,Generic,5,,Chert,,,,"Natchez,Shawnee,Cherokee",,
32,Copper,Generic,4,,Copper,,,,"HoChunk,HoChunk,Shawnee,Cherokee",,
33,Tobacco,Generic,3,,,,Cherokee,,"Cherokee,Natchez",,
34,Sunflowers,Generic,4,,,,Caddo,,"Caddo,Natchez",,
35,The Three Sisters,Generic,2,,,,Natchez,,"Natchez,Caddo",,
36,Mobilian Jargon,Generic,4,,,Cherokee,,,"Natchez,Shawnee",,
37,The Chunkey Game,Generic,5,,,,,,CaddoOrShawnee,,
38,Adena Culture,Generic,5,,,,,,"Caddo,Shawnee,Shawnee,Natchez",,
39,Hopewell Culture,Generic,3,,,,,,"Shawnee,HoChunk",,
40,Mississippian Culture,Generic,4,,,,,,"Natchez,Cherokee,Cherokee",,
41,Burial Mounds,Generic,3,,,Natchez,,,"HoChunk,HoChunk,Caddo,Caddo",,
42,Platform Mounds,Generic,2,,,Caddo,,,"Shawnee,Natchez,Cherokee",,
43,Effigy Mounds,Generic,2,,,,,,"HoChunk,HoChunk,Shawnee,Shawnee,Caddo",,
44,Pottery,Generic,3,,,Shawnee,All,TRUE,"Natchez,Cherokee",,
45,The Buzzard Cult,Generic,7,,,,,,Cherokee,,
46,Wattle & Daub,Generic,3,,,,All,,"Shawnee,Caddo",,
47,Oneota Culture,Generic,6,,,,,,"HoChunk,HoChunk,Natchez,Shawnee",,
48,Human Sacrifice,Generic,6,,,,All,TRUE,"Cherokee,Cherokee,Natchez,Natchez",,
49,Black Drink,Generic,5,,,,All,,"Cherokee,Caddo",,
50,Cahokia,Generic,1,,,,All,,Cherokee,,`

const (
	colNumber = iota
	colTitle
	colEra
	colActionPoints
	colIsWhite
	colResourceBonus
	colRevolt
	colModifier
	colIsAscendant
	colAdvancingArmies
	colIsAvaricia
	colIsSpanish
)

// drawFromPile draws the next card from a pile.
func drawFromPile(p Pile) (*HistoryCard, Pile) {
	if len(p) == 0 {
		return nil, p
	}
	c := p[0]
	p = p[1:]
	return c, p
}

// makeHistoryCards makes a new Pile of history cards from the card data.
func makeHistoryCards() Pile {
	var err error
	data := parse(historyCardData)
	if err != nil {
		panic(err)
	}
	cards := make([]*HistoryCard, 50)
	for i, r := range data {
		if i == 0 {
			// skip the header row
			continue
		}
		c := &HistoryCard{
			Title:           r[colTitle],
			Era:             eraNameLookup[strings.ToUpper(r[colEra])],
			AdvancingArmies: getTribes(r[colAdvancingArmies]),
			ResourceBonus:   getResourceBonus(r[colResourceBonus]),
			IsWhite:         r[colIsWhite] == "TRUE",
			IsAscendant:     r[colIsAscendant] == "TRUE",
			IsAvaricia:      r[colIsAvaricia] == "TRUE",
			IsSpanish:       r[colIsSpanish] == "TRUE",
		}
		revolt := getTribes(r[colRevolt])
		c.Revolt = None
		if len(revolt) > 0 {
			c.Revolt = revolt[0]
		}
		if c.Number, err = strconv.Atoi(r[colNumber]); err != nil {
			panic(err)
		}
		if c.ActionPoints, err = strconv.Atoi(r[colActionPoints]); err != nil {
			panic(err)
		}
		if modifier := getTribes(r[colModifier]); len(modifier) > 0 {
			c.Modifier = modifier[0]
		}
		cards[i-1] = c
	}
	return cards
}

func getTribes(s string) []Tribe {
	if s == "" {
		return nil
	}
	var ok bool
	names := strings.Split(s, ",")
	t := make([]Tribe, len(names))
	for i, n := range names {
		if t[i], ok = tribeNameLookup[n]; !ok {
			msg := fmt.Sprintf("Bad tribe name %q in card data", n)
			panic(msg)
		}
	}
	return t
}

func getResourceBonus(s string) []TradeGood {
	if s == "" {
		return nil
	}
	var ok bool
	names := strings.Split(s, ",")
	result := make([]TradeGood, len(names))
	for i, n := range names {
		if result[i], ok = tradeGoodNameLookup[n]; !ok {
			msg := fmt.Sprintf("Bad trade good name %q in card data", n)
			panic(msg)
		}
	}
	return result
}

func shufflePile(cards Pile) {
	for i, _ := range cards {
		j := i + rng.Intn(len(cards)-i)
		cards[i], cards[j] = cards[j], cards[i]
	}
}

func splitByEra(cards Pile) map[Era]Pile {
	result := make(map[Era]Pile)
	for _, c := range cards {
		result[c.Era] = append(result[c.Era], c)
	}
	return result
}

// makeHistoryDeck makes the game's History Deck.
func makeHistoryDeck() Pile {
	var deck Pile
	cards := makeHistoryCards()
	var c *HistoryCard
	var early Pile
	var mid Pile
	var late Pile
	var temp Pile

	eras := splitByEra(cards)
	hopewell := eras[Hopewell]
	spanish := eras[Spanish]
	mississippian := eras[Mississippian]
	generic := eras[Mississippian]

	// make the early deck from 10 random Hopewell cards
	shufflePile(hopewell)
	for i := 0; i < 10; i++ {
		c, hopewell = drawFromPile(hopewell)
		early = append(early, c)
	}

	// This logic depends on Coosa being the first Spanish card
	// and The Spanish being the second.
	for i := 0; i < 2; i++ {
		c, spanish = drawFromPile(spanish)
		temp = append(temp, c)
		for j := 0; j < 4; j++ {
			c, generic = drawFromPile(generic)
			temp = append(temp, c)
		}
		shufflePile(temp)
		late = append(late, temp...)
		temp = nil
	}

	mid = append(hopewell, mississippian...)
	mid = append(mid, generic...)
	shufflePile(mid)

	deck = append(early, mid...)
	deck = append(deck, late...)
	return deck
}
